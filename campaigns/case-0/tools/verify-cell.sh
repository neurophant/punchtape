#!/bin/bash
# verify-cell.sh — the staging audit of the case-0 host world
# (the 0.2.0 verify-cell discipline carried to host cells): a staged
# cell is INVISIBLE-valid until every fact below is checked. No
# container exists here — the live-bind check of 0.2.0 becomes the
# work-dir writability check and the input-digest re-verification.
#
# Usage: verify-cell.sh <run-root> <cell>
# Checks (all must pass, exit non-zero on any failure):
#   1. the manifest exists and its cell_path IS the cell directory;
#   2. every staged input still matches its manifest digest (the
#      brief, the expectations list, the materials, the probes);
#   3. the work dir is writable (the hand's access path);
#   4. the machine side: top — the system punchtape sha equals
#      the run's binary_sha256; solo — the spawn message carries no
#      mention of the machine and the cell carries no machine files;
#   5. the spawn message exists and matches the manifest digest;
#   6. the form-hints sidcar:
#      top — the sidcar rides the input and matches its manifest
#      digest; solo — the sidcar is ABSENT (the hands' inputs differ
#      by this one file and nothing else; the common inputs above are
#      byte-identical for both hands of a task).
set -euo pipefail
RR="${1:?run root}"; CELL="${2:?cell}"
HERE="$(cd "$(dirname "$0")" && pwd)"
RR="$(readlink -f "$RR")"

fail() { echo "VERIFY-CELL $CELL FAIL: $1" >&2; exit 1; }
[ -f "$RR/manifests/$CELL.json" ] || fail "no manifest"

python3 - "$RR" "$CELL" <<'PY'
import hashlib, json, os, subprocess, sys

root, cell = sys.argv[1:3]
m = json.load(open(f"{root}/manifests/{cell}.json", encoding="utf-8"))
cellpath = m["cell_path"]
hand = m["hand"]

def sha(p):
    return hashlib.sha256(open(p, "rb").read()).hexdigest()

def fail(msg):
    print(f"VERIFY-CELL {cell} FAIL: {msg}", file=sys.stderr)
    sys.exit(1)

# 1. the cell directory
if not os.path.isdir(cellpath):
    fail(f"cell dir gone: {cellpath}")

# 2. the inputs against the manifest
inp = os.path.join(cellpath, "input")
pairs = [(os.path.join(inp, "brief.md"), m["inputs"]["brief"]["sha256"]),
         (os.path.join(inp, "EXPECTATIONS.md"), m["inputs"]["expectations"]["sha256"])]
for rel, want in sorted(m["inputs"]["materials"].items()):
    pairs.append((os.path.join(inp, rel), want))
for name, want in sorted(m["inputs"]["probes"].items()):
    pairs.append((os.path.join(inp, "probes", name), want))
for f, want in pairs:
    if not os.path.isfile(f):
        fail(f"staged input missing: {f}")
    if sha(f) != want:
        fail(f"staged input drifted: {f}")

# 3. the work dir answers
work = os.path.join(cellpath, "work")
probe = os.path.join(work, ".verify-write")
try:
    with open(probe, "w") as fh:
        fh.write("ok")
    os.remove(probe)
except OSError as e:
    fail(f"work dir not writable: {e}")

# 4. the machine side
if hand == "top":
    sysbin = subprocess.run(["sh", "-lc", "command -v punchtape"],
                            capture_output=True, text=True).stdout.strip()
    if not sysbin:
        fail("punchtape is not installed on the system")
    got = sha(sysbin)
    want = (m.get("binary_sha256") or (m["inputs"]["binary"] or {}).get("sha256"))
    if not want or got != want:
        fail(f"system punchtape {got[:12]} != the run's binary {want}")
    if not os.access(sysbin, os.X_OK):
        fail("the system punchtape is not executable")
else:
    spawn = os.path.join(root, "tasks", cell, "SPAWN.txt")
    if os.path.isfile(spawn) and "punchtape" in open(spawn, encoding="utf-8").read():
        fail("the solo spawn message mentions the machine")
    for name in ("pt", "punchtape"):
        if os.path.exists(os.path.join(cellpath, name)):
            fail(f"a machine binary rides the solo cell: {name}")

# 5. the spawn message
spawn = os.path.join(root, "tasks", cell, "SPAWN.txt")
if not os.path.isfile(spawn):
    fail("no SPAWN.txt in the task dir")
if sha(spawn) != m["inputs"]["spawn"]["sha256"]:
    fail("SPAWN.txt drifted from the manifest")

# 6. the form-hints sidcar: the machine hand's own
# quotable brief lines; the solo hand's input never carries the file
# (the parity law: common inputs byte-identical, this one file is the
# only hand-conditional difference)
fh = os.path.join(inp, "form-hints.md")
if hand == "top":
    want = (m["inputs"].get("form_hints") or {}).get("sha256")
    if not want:
        fail("the machine cell's manifest pins no form-hints sidcar")
    if not os.path.isfile(fh):
        fail("the machine cell carries no form-hints.md")
    if sha(fh) != want:
        fail("form-hints.md drifted from the manifest")
else:
    if os.path.exists(fh):
        fail("the solo cell carries the machine-only form-hints sidcar")
    if m["inputs"].get("form_hints") is not None:
        fail("the solo manifest pins a form-hints sidcar")

print(f"verify-cell {cell}: PASS ({'solo: machine absent' if hand == 'solo' else hand + ': system punchtape sha verified'}; inputs match; work writable; form-hints {'absent' if hand == 'solo' else 'verified'})")
PY
