#!/bin/bash
# selftest.sh — the self-test of the case-0 campaign suite. Like
# the harness self-tests it replaces, every check is a REAL run; the
# RED-DEMO checks (marked) prove the gates actually bite — a gate
# that cannot fail is decoration. Scratch: a disk-backed dir under
# the platform scratch area, removed on exit; NOTHING here spawns an
# agent or solves a task — the smoke cell stages MECHANICS ONLY
# (class=smoke, outside the serial wave data).
set -uo pipefail

HERE="$(cd "$(dirname "$0")" && pwd)"
CAMP="$(cd "$HERE/.." && pwd)"
SCR="${C0_SCRATCH_ROOT:-$HOME/.experiments}/tmp/c0-selftest-$$"
mkdir -p "$SCR"
FAILS=0
ok()   { echo "  ok  $*"; }
bad()  { echo "  RED $*"; FAILS=$((FAILS + 1)); }
demo() { echo "-- DEMO $*"; }
step() { echo "== $*"; }

cleanup() {
  rm -rf "$SCR"
}
trap cleanup EXIT

# ------------------------------------------------------------------
step "1. the host prompt renders converge (named anchors)"
if bash "$HERE/make-host-prompts.sh" --verify >/dev/null 2>&1; then
  ok "prompts-host renders match"
else
  bad "prompts-host renders drifted"
fi
grep -q "__WORK__" "$CAMP/prompts-host/HARNESS-HOST.md" \
  && ok "HARNESS-HOST carries the __WORK__ placeholder" \
  || bad "HARNESS-HOST lost the placeholder"
grep -qi "docker" "$CAMP/prompts-host/HARNESS-HOST.md" \
  && bad "docker vocabulary survived in HARNESS-HOST" \
  || ok "HARNESS-HOST is docker-free"

# ------------------------------------------------------------------
step "2. the series freeze: one binary, methodology digest, staged inputs"
FREEZE="$(ls -t "$CAMP"/FREEZE-S.sha256 2>/dev/null | head -1 || true)"
[ -f "$FREEZE" ] || { bad "no series freeze"; exit 1; }
SYS_SHA="$(sha256sum "$(command -v punchtape)" | cut -d' ' -f1)"
FROZEN_SHA="$(grep '^series-binary' "$FREEZE" | awk '{print $2}')"
[ "$SYS_SHA" = "$FROZEN_SHA" ] \
  && ok "the system punchtape IS the frozen binary ($SYS_SHA)" \
  || bad "system binary $SYS_SHA != frozen $FROZEN_SHA"
grep -q '^methodology' "$FREEZE" \
  && ok "the freeze carries METHODOLOGY" \
  || bad "no methodology line in the freeze"
# baseline is ABSENT by law: a freeze that ever carries a baseline
# line is a red gate, not a missing kind
for kind in brief material prompt prompt-host probe prices models world-lock docs tool form-hints; do
  grep -q "^$kind " "$FREEZE" || bad "freeze carries no $kind lines"
done
if grep -q '^baseline ' "$FREEZE"; then
  bad "freeze carries baseline lines (a baseline line must never return)"
else
  ok "freeze carries no baseline lines (the law holds)"
fi
ok "freeze kinds complete ($(grep -c . "$FREEZE") lines)"

# ------------------------------------------------------------------
step "3. the smoke cell stages and verifies (brim-cpp-top, class=smoke)"
RR="$SCR/run-root"
mkdir -p "$RR"
echo "selftest-driver-uuid" > "$RR/driver-uuid.txt"
export C0_SCRATCH_ROOT="$SCR/cells"
export C0_DRIVER_UUID="selftest-driver-uuid"
if bash "$HERE/prep-cell.sh" "$RR" brim-cpp top --class smoke >"$SCR/prep.log" 2>&1; then
  ok "prep-cell brim-cpp-top staged + env-checked (log: $SCR/prep.log)"
else
  bad "prep-cell failed:"; tail -5 "$SCR/prep.log" | sed 's/^/    /'
fi
CELL="brim-cpp-top"
MAN="$RR/manifests/$CELL.json"
[ -f "$MAN" ] && ok "manifest written" || bad "no manifest"
python3 - "$MAN" <<'PY' && ok "manifest: hand=top, class=smoke, binary pinned, expectations staged" \
  || bad "manifest content wrong"
import json, sys
m = json.load(open(sys.argv[1]))
assert m["hand"] == "top" and m["class"] == "smoke"
assert m["inputs"]["binary"] and m["inputs"]["binary"]["sha256"]
assert m["inputs"]["expectations"]["sha256"]
assert m["inputs"]["probes"], "no staged probes"
assert m["world"]["exec"] == "host-direct"
PY
[ -f "$RR/tasks/$CELL/SPAWN.txt" ] && ok "SPAWN.txt archived" || bad "no SPAWN.txt"
grep -q "punchtape" "$RR/tasks/$CELL/SPAWN.txt" \
  && ok "the machine hand's spawn names the machine" \
  || bad "the machine hand's spawn lost the machine"
# the harness prefix (cd work && . ./ENV) and the
# walls are IN the spawn bytes; prep wrote the hand's own work/ENV
grep -q "\. \./ENV" "$RR/tasks/$CELL/SPAWN.txt" \
  && ok "the spawn carries the ENV prefix command form" \
  || bad "the spawn lost the ENV prefix form"
grep -q "WRITE BOUNDS" "$RR/tasks/$CELL/SPAWN.txt" \
  && ok "the spawn carries the filesystem walls" \
  || bad "the spawn lost the filesystem walls"
grep -q "TMPDIR" "$RR/tasks/$CELL/SPAWN.txt" >/dev/null 2>&1 || true
CELLPATH="$(python3 -c "import json; print(json.load(open('$RR/manifests/$CELL.json'))['cell_path'])")"
[ -f "$CELLPATH/work/ENV" ] && grep -q "TMPDIR=" "$CELLPATH/work/ENV" \
  && ok "prep wrote work/ENV with the TMPDIR export (the hand's mechanical env)" \
  || bad "work/ENV missing or without TMPDIR"
# the port-range allocation: the cell's
# own disjoint window below the ephemeral range, claimed in the root
ls "$RR"/port-ranges/ >/dev/null 2>&1 \
  && ok "prep claimed a port window in $RR/port-ranges ($(ls "$RR"/port-ranges | head -1)-...)" \
  || bad "no port-range claim under $RR/port-ranges"
grep -q 'C0_PORT_RANGE="' "$CELLPATH/work/ENV" \
  && ok "work/ENV exports the cell's own C0_PORT_RANGE" \
  || bad "work/ENV lost the C0_PORT_RANGE export"
if bash "$HERE/verify-cell.sh" "$RR" "$CELL" >"$SCR/verify.log" 2>&1; then
  ok "verify-cell PASS"
else
  bad "verify-cell failed:"; tail -3 "$SCR/verify.log" | sed 's/^/    /'
fi
# the machine cell's form-hints sidcar —
# verbatim numbered brief lines, manifest digest, frozen digest
FH="$CELLPATH/input/form-hints.md"
[ -f "$FH" ] && ok "the machine cell carries the form-hints sidcar" \
  || bad "no form-hints sidcar on the machine cell"
python3 - "$MAN" "$FH" <<'PY' && ok "sidcar: manifest digest pins it, every L-line is a verbatim numbered brief line" \
  || bad "sidcar content or digest wrong"
import hashlib, json, sys
man, fh = sys.argv[1:3]
m = json.load(open(man))
want = (m["inputs"].get("form_hints") or {}).get("sha256")
assert want, "the manifest pins no form_hints"
assert hashlib.sha256(open(fh, "rb").read()).hexdigest() == want
brief = open(m["cell_path"] + "/input/brief.md", encoding="utf-8").read().splitlines()
hints = [l for l in open(fh, encoding="utf-8").read().splitlines() if l.startswith("L")]
assert hints, "no L-lines in the sidcar"
for h in hints:
    n, _, text = h.partition(": ")
    assert brief[int(n[1:]) - 1].rstrip() == text, h
PY
if python3 "$HERE/check-staged-freeze.py" "$FREEZE" brim cpp "$CELLPATH/input" >/dev/null 2>&1; then
  ok "the sidcar digest matches the series freeze (form-hints line)"
else
  bad "the sidcar failed the staged-freeze check (no form-hints line in the freeze?)"
fi
demo "a doctored sidcar must fail the freeze check"
printf 'L999: forged line\n' >> "$FH"
if python3 "$HERE/check-staged-freeze.py" "$FREEZE" brim cpp "$CELLPATH/input" >/dev/null 2>&1; then
  bad "doctored sidcar passed the staged-freeze check"
else
  ok "doctored sidcar caught (RED DEMO)"
fi
demo "the solo cell never carries the sidcar (parity: inputs differ by this one file only)"
if bash "$HERE/stage-cell.sh" brim cpp solo "$RR" --class smoke >"$SCR/stage-solo.log" 2>&1; then
  SOLOMAN="$(ls -t "$RR"/manifests/brim-cpp-solo.json 2>/dev/null | head -1)"
  SOLOPATH="$(python3 -c "import json;print(json.load(open('$SOLOMAN'))['cell_path'])")"
  if [ -e "$SOLOPATH/input/form-hints.md" ]; then
    bad "the solo cell carries the machine-only sidcar"
  else
    ok "the solo input carries no form-hints.md"
  fi
  if bash "$HERE/verify-cell.sh" "$RR" brim-cpp-solo >"$SCR/verify-solo.log" 2>&1; then
    ok "verify-cell PASS for the solo hand too (the parity gate is green for both hands)"
  else
    bad "verify-cell failed on the solo cell:"; tail -3 "$SCR/verify-solo.log" | sed 's/^/    /'
  fi
else
  bad "solo stage failed:"; tail -3 "$SCR/stage-solo.log" | sed 's/^/    /'
fi
# the generator itself: a formless brief prints the EXPLICIT mark
# (never silence), a formed brief quotes its
# lines verbatim with numbers
GENBRIEF="$SCR/gen-brief.md"
printf 'A plain wish.\nNothing exact anywhere.\n' > "$GENBRIEF"
if python3 "$HERE/form-hints.py" "$GENBRIEF" | grep -q "no form-bearing terms"; then
  ok "a formless brief prints the explicit no-terms mark (not silence)"
else
  bad "the formless brief's sidcar lost the explicit mark"
fi
printf 'Run it as:\n    tool --mode fast host.example\nIt prints "done" at exit.\n' > "$GENBRIEF"
python3 "$HERE/form-hints.py" "$GENBRIEF" > "$SCR/gen-hints.md"
grep -q '^L2:     tool --mode fast host.example$' "$SCR/gen-hints.md" \
  && grep -q '^L3: It prints "done" at exit\.$' "$SCR/gen-hints.md" \
  && ok "a formed brief quotes its lines verbatim with numbers" \
  || bad "the generated quotes are not verbatim numbered lines"
demo "a doctored staged brief must fail the freeze check"
python3 - "$MAN" <<'PY'
import json, os, sys
m = json.load(open(sys.argv[1]))
p = os.path.join(m["cell_path"], "input", "brief.md")
open(p, "a").write("\ndoctor\n")
PY
if python3 "$HERE/check-staged-freeze.py" "$FREEZE" brim cpp "$(python3 -c "import json;print(json.load(open('$MAN'))['cell_path'])")/input" >/dev/null 2>&1; then
  bad "doctored brief passed the staged-freeze check"
else
  ok "doctored brief caught (RED DEMO)"
fi
CELLPATH="$(python3 -c "import json;print(json.load(open('$MAN'))['cell_path'])")"
# remove the doctored smoke cell (scratch cleanup)
python3 - "$RR" <<'PY'
import glob, json, os, shutil, sys
root = sys.argv[1]
for m in glob.glob(f"{root}/manifests/brim-cpp-*.json"):
    d = json.load(open(m))
    shutil.rmtree(d["cell_path"], ignore_errors=True)
    parent = os.path.dirname(d["cell_path"])
    try:
        os.rmdir(parent)
    except OSError:
        pass
    os.remove(m)
PY
demo "a third hand must refuse to stage (the two-hand gate)"
if bash "$HERE/stage-cell.sh" brim cpp weak "$RR" --class smoke >/dev/null 2>&1; then
  bad "stage-cell accepted the dead weak hand"
else
  ok "stage-cell refused the dead weak hand (RED DEMO)"
fi

# ------------------------------------------------------------------
step "4. preflight bites (RED DEMO)"
BADSPAWN="$SCR/bad-spawn.txt"
printf 'You are a hand. spawn a subagent to help you.\n' > "$BADSPAWN"
if bash "$HERE/preflight-audit.sh" prompts "$BADSPAWN" >/dev/null 2>&1; then
  bad "preflight accepted a spawn-instruction prompt"
else
  ok "preflight refused the spawn-instruction prompt (RED DEMO)"
fi

# ------------------------------------------------------------------
step "5. the corpus replay: green x2, and it bites (the blind runner)"
PROD="$SCR/prod"; mkdir -p "$PROD/.punchtape/canon/atoms" "$SCR/input/probes"
cp "$CAMP/materials/probes/"*.sh "$SCR/input/probes/" 2>/dev/null
printf '#!/bin/sh\necho "hello $1"\n' > "$PROD/greet"; chmod +x "$PROD/greet"
cat > "$PROD/.punchtape/canon/atoms/greet.yaml" <<'YAML'
id: greet
name: greeting
prose: the product greets
rows:
  - run: ["./greet", "world"]
    out: "hello world"
    rc: 0
  - run: "./greet"
    out: "(contains) hello"
  - seeds:
      - path: in.txt
        text: "abc"
    run: "sh -c 'cat in.txt'"
    out: "abc"
YAML
R="$(python3 "$HERE/case-runner-host.py" --product "$PROD" --input "$SCR/input" --runs 2 --scratch "$SCR")"
echo "$R" | head -2 | python3 -c '
import json, sys
runs = [json.loads(l) for l in sys.stdin]
assert len(runs) == 2, "two runs expected"
for r in runs:
    assert r["fail"] == 0 and r["env_fail"] == 0, ("run not green", r)
    assert r["tolerant"] == 0, ("EOL is the method default now", r)
' && ok "synthetic corpus green x2, tolerant=0 (the trailing-\\n is the method's own EOL normalization — a strict pass)" || bad "corpus replay not green (or tolerant noise survived)"
demo "a broken product must red"
printf '#!/bin/sh\necho "bye $1"\n' > "$PROD/greet"
R="$(python3 "$HERE/case-runner-host.py" --product "$PROD" --input "$SCR/input" --runs 1 --scratch "$SCR")"
echo "$R" | python3 -c '
import json, sys
r = json.loads(sys.stdin.readline())
assert r["fail"] >= 1, "a broken product passed"
' && ok "broken product reds (RED DEMO)" || bad "broken product passed"
printf '#!/bin/sh\necho "hello $1"\n' > "$PROD/greet"; chmod +x "$PROD/greet"

# ------------------------------------------------------------------
step "6. the accounting pipeline: registry -> db ledger -> collect -> report"
SYN="$SCR/synwave"
mkdir -p "$SYN"
echo "selftest-driver-uuid" > "$SYN/driver-uuid.txt"
# the fake platform db: two sessions (a hand and an operator), the
# tables usage-ledger.py reads
python3 - "$SYN" <<'PY'
import json, os, sqlite3, sys, yaml
root = sys.argv[1]
db = sqlite3.connect(f"{root}/fake.sqlite")
db.execute("CREATE TABLE session (id TEXT PRIMARY KEY, directory TEXT, time_created INT)")
db.execute("""CREATE TABLE model_usage (session_id TEXT, input_tokens INT,
    output_tokens INT, cache_read_input_tokens INT, cache_creation_input_tokens INT,
    duration_ms INT, tool_call_count INT, model_id TEXT, status TEXT)""")
db.execute("CREATE TABLE part (session_id TEXT, data TEXT)")
cells = {}
for hand, sess, inp, outp, cache, dur in (
    ("solo", "sess-solo", 1000, 100, 500, 30000),
    ("top",  "sess-top",  300,  30,  150, 10000),
):
    cell = f"conduit-rust-{hand}"
    cells[cell] = {"sess": sess}
    # time_created predates the run root: outside the census window
    db.execute("INSERT INTO session VALUES (?, ?, 0)", (sess, f"{root}/cells/{cell}/work"))
    db.execute("INSERT INTO model_usage VALUES (?,?,?,?,?,?,?,?,?)",
               (sess, inp, outp, cache, 0, dur, 5, "GLM-5.3" if hand == "solo"
                else "GLM-5.3", "completed"))
    db.execute("INSERT INTO part VALUES (?, ?)", (sess, '{"type":"step-finish"}'))
db.commit()
db.close()
# the synthetic cells: manifest + registry + the machine tree (top) +
# the solo tree (solo)
reg = {}
for hand in ("solo", "top"):
    cell = f"conduit-rust-{hand}"
    cellpath = f"{root}/cells/{cell}"
    work = f"{cellpath}/work"
    os.makedirs(f"{work}/.punchtape/canon/atoms", exist_ok=True)
    os.makedirs(f"{work}/.punchtape/verdicts", exist_ok=True)
    os.makedirs(f"{root}/manifests", exist_ok=True)
    man = {
        "cell": cell, "sense": "conduit", "stack": "rust", "hand": hand,
        "class": "smoke", "protocol": "0.2.0",
        "cell_path": cellpath, "driver_uuid": "selftest-driver-uuid",
        "binary_sha256": "0" * 64,
        "inputs": {
            "brief": {"path": f"briefs/conduit/rust.md", "sha256": "b" * 64},
            "expectations": {"path": "prompts/EXPECTATIONS.md", "sha256": "e" * 64},
            "materials": {}, "probes": {}, "spawn": {"sha256": "s" * 64},
            "prices": {"sha256": "p" * 64},
            "binary": None if hand == "solo" else {"sha256": "0" * 64},
        },
        "world": {"exec": "host-direct", "network": "textual-directive"},
    }
    json.dump(man, open(f"{root}/manifests/{cell}.json", "w"), indent=1, sort_keys=True)
    # the registry uuids carry the 8-hex id8 the ledger keys on
    reg[cell] = [{"uuid": f"agent_{'a' * 7}{1 if hand == 'solo' else 2}", "role": "hand",
                  "registered": "2000-01-01T00:00:00+00:00"}]
    if hand == "top":
        # the machine telemetry: a ledger + a verdict + a green atom
        open(f"{work}/.punchtape/ledger.yamll", "w").write(
            "---\nname: next\noutcome: ok\nwall-ms: 100\nruns-ms: 0\nreply-bytes: 50\n"
            "---\nname: submit\noutcome: accepted\nwall-ms: 120\nruns-ms: 50\nreply-bytes: 80\n"
            "---\nname: submit\noutcome: rejected\nwall-ms: 60\nruns-ms: 10\nreply-bytes: 40\n")
        open(f"{work}/.punchtape/verdicts/v1.txt", "w").write(
            "VERDICT: READY\ndigest: 1a2b3c4d\n")
        yaml.safe_dump({"id": "tunnel", "name": "the tunnel", "prose": "wires bytes",
                        "rows": [{"run": ["./bin/conduit"], "rc": 0,
                                  "outcome": "green"}]},
                       open(f"{work}/.punchtape/canon/atoms/tunnel.yaml", "w"))
    else:
        os.makedirs(f"{work}/solo-product", exist_ok=True)
        open(f"{work}/solo-product/SELF-REPORT.md", "w").write("honest\n")
json.dump(reg, open(f"{root}/driver-agents.json", "w"), indent=1)
PY
if python3 "$HERE/usage-ledger.py" --root "$SYN" --db "$SYN/fake.sqlite" >"$SCR/ul.log" 2>&1; then
  ok "usage-ledger built the ledger from the fake db"
else
  bad "usage-ledger failed:"; tail -3 "$SCR/ul.log" | sed 's/^/    /'
fi
python3 - "$SYN" <<'PY' && ok "the ledger carries both solo and top hands" || bad "ledger incomplete"
import json, sys
d = json.load(open(f"{sys.argv[1]}/usage/ledger.json"))["agents"]
assert any("hand" in (a.get("roles") or []) for a in d.values())
assert len(d) >= 2
PY
COLLECT_OK=1
bash "$HERE/collect-metrics.sh" "$SYN" conduit-rust-top --verdict-from-machine >"$SCR/cm.log" 2>&1 || COLLECT_OK=0
bash "$HERE/collect-metrics.sh" "$SYN" conduit-rust-solo --accept "selftest acceptance" >>"$SCR/cm.log" 2>&1 || COLLECT_OK=0
[ "$COLLECT_OK" = 1 ] && ok "collect-metrics built both cells' evidence" \
  || { bad "collect-metrics failed:"; tail -5 "$SCR/cm.log" | sed 's/^/    /'; }
python3 - "$SYN" <<'PY' && ok "the machine verdict and the solo acceptance landed" || bad "rulings missing"
import json, sys
root = sys.argv[1]
t = json.load(open(f"{root}/tasks/conduit-rust-top/evidence/summary.json"))
assert t["ruling"] and t["ruling"].startswith("VERDICT: READY"), t
s = json.load(open(f"{root}/tasks/conduit-rust-solo/evidence/summary.json"))
assert s["ruling"] == "selftest acceptance"
p = json.load(open(f"{root}/tasks/conduit-rust-top/evidence/platform.json"))
assert p["tokens_inout"] == 330 and p["tokens_total"] == 480, p
PY
# (the byte-metric plumbing is gone by the owner's word:
#  metrics are tokens/seconds from the platform db only; a ceremony
#  size, if ever wanted, is read from the db's own messages)
if python3 "$HERE/report.py" "$SYN" --prices "$CAMP/prices.json" \
     --models "$CAMP/models.json" --out "$SYN/wave-report.md" >"$SCR/rep.log" 2>&1; then
  ok "report.py printed the wave report"
else
  bad "report.py failed:"; tail -5 "$SCR/rep.log" | sed 's/^/    /'
fi
grep -q "in+out" "$SYN/wave-report.md" && grep -q "total" "$SYN/wave-report.md" \
  && ok "the report prints BOTH token bases" \
  || bad "the report lost a base"
grep -q "solo/top" "$SYN/wave-report.md" \
  && ok "the report prints the honest multiples" \
  || bad "no multiples in the report"
grep -q "cache%" "$SYN/wave-report.md" \
  && ok "the report prints the cache% column (informational)" \
  || bad "no cache% column in the report"

# the acceptance freeze: a re-collect's service
# --accept note never overwrites the operator's acceptance verdict
bash "$HERE/collect-metrics.sh" "$SYN" conduit-rust-solo \
     --accept "re-collect on settled ledger" >>"$SCR/cm.log" 2>&1 \
  || bad "the re-collect failed"
python3 - "$SYN" <<'PY' && ok "the acceptance verdict froze at the first finish (the re-collect note only appended)" || bad "the re-collect OVERWROTE the acceptance"
import json, sys
root = sys.argv[1]
s = json.load(open(f"{root}/tasks/conduit-rust-solo/evidence/summary.json"))
assert s["ruling"] == "selftest acceptance", s
a = json.load(open(f"{root}/tasks/conduit-rust-solo/evidence/acceptance.json"))
assert a["ruling"] == "selftest acceptance" and a["recollects"], a
PY

# count-deliverables on the PRODUCT PATH (the input is the product
# tree, not the task dir)
python3 "$HERE/count-deliverables.py" "$SYN/cells/conduit-rust-top/work" rust top >"$SCR/cd-top.json" 2>&1 \
  && python3 - "$SCR/cd-top.json" <<'PY' && ok "count-deliverables: the machine branch reads the machine layout (atoms/rows/verdict coverage linkage)" || bad "count-deliverables machine branch wrong"
import json, sys
d = json.load(open(sys.argv[1]))
assert d["specs_sections"] == 1 and d["scenarios"] == 1, d
assert d["coverage"] == 1 and "machine-verdict-coverage" in d["coverage_method"], d
assert d["machine"]["verdict"] == "READY", d
PY
python3 "$HERE/count-deliverables.py" "$SYN/cells/conduit-rust-solo/work/solo-product" rust solo >"$SCR/cd-solo.json" 2>&1 \
  && python3 - "$SCR/cd-solo.json" <<'PY' && ok "count-deliverables: the solo branch answers on the product path" || bad "count-deliverables solo branch wrong"
import json, sys
d = json.load(open(sys.argv[1]))
assert d["machine"] is None and "product" in d, d
PY

# the glued session report (report.py glues it itself: no manual cat)
mkdir -p "$SCR/docs"
printf '# Analytics layer (synthetic)\n\nAll synthetic.\n' > "$SYN/analytics.md"
python3 "$HERE/report.py" "$SYN" --prices "$CAMP/prices.json" \
  --models "$CAMP/models.json" --out "$SYN/wave-report.md" \
  --glue "$SYN/analytics.md" --glue-out "$SCR/docs" >"$SCR/rep2.log" 2>&1 \
  && ls "$SCR/docs"/case0-*-session-report.md >/dev/null 2>&1 \
  && GLUED="$(ls "$SCR/docs"/case0-*-session-report.md | head -1)" \
  && grep -q "Analytics layer (synthetic)" "$GLUED" \
  && grep -q "in+out" "$GLUED" \
  && ok "report.py --glue wrote the session report itself (core+analytics: $GLUED)" \
  || bad "the glue step failed"

# verify-metrics (the systemic verification: census<->pins<->db<->
# ledger<->every table cell, full source paths)
if python3 "$HERE/verify-metrics.py" "$SYN" --db "$SYN/fake.sqlite" \
     --report "$SYN/wave-report.md" >"$SCR/vm.log" 2>&1; then
  grep -q "RESULT: " "$SCR/vm.log" && grep -qv "RED " "$SCR/vm.log" \
    && ok "verify-metrics GREEN on the synthetic wave ($(grep -o '[0-9]* check(s) AGREE' "$SCR/vm.log" | head -1))" \
    || { bad "verify-metrics printed RED lines:"; grep "RED " "$SCR/vm.log" | head -3 | sed 's/^/    /'; }
else
  bad "verify-metrics failed:"; tail -5 "$SCR/vm.log" | sed 's/^/    /'
fi
demo "an unpinned window session must RED the census"
python3 - "$SYN" <<'PY'
import sqlite3, time, sys
db = sqlite3.connect(f"{sys.argv[1]}/fake.sqlite")
db.execute("INSERT INTO session VALUES ('sess_subagent_agent_rogue', ?, ?)",
           (f"{sys.argv[1]}/nowhere", int(time.time() * 1000)))
db.commit()
PY
if python3 "$HERE/verify-metrics.py" "$SYN" --db "$SYN/fake.sqlite" \
     --report "$SYN/wave-report.md" >"$SCR/vm2.log" 2>&1; then
  bad "the census gate passed an unpinned window session"
else
  grep -q "unpinned session in the wave window" "$SCR/vm2.log" \
    && ok "unpinned window session caught by the census (RED DEMO)" \
    || { bad "verify-metrics failed for another reason:"; tail -5 "$SCR/vm2.log" | sed 's/^/    /'; }
fi
python3 - "$SYN" <<'PY'
import sqlite3, sys
db = sqlite3.connect(f"{sys.argv[1]}/fake.sqlite")
db.execute("DELETE FROM session WHERE id = 'sess_subagent_agent_rogue'")
db.commit()
PY

# the audit's environment-probe exemption: a which/command -v line
# is a probe, an actual nc reds
AUDRR="$SCR/audit-root"; mkdir -p "$AUDRR/rollouts"
echo "selftest-driver-uuid" > "$AUDRR/driver-uuid.txt"
python3 - "$AUDRR" <<'PY'
import json, sys
root = sys.argv[1]
for key, cmd in (("probe-cell", "which python3 nc socat strace"),
                 ("call-cell", "nc host.example 1234")):
    with open(f"{root}/rollouts/{key}.jsonl", "w") as f:
        f.write(json.dumps({"request": {"toolCalls": [
            {"name": "Bash", "input": {"command": cmd}}]}}) + "\n")
PY
bash "$HERE/audit-net.sh" "$AUDRR" >"$SCR/aud.log" 2>&1 || true
if grep -q "probe-cell" "$AUDRR/redlines.jsonl" 2>/dev/null; then
  bad "the which-probe line still reds (the false canary survived)"
else
  ok "the environment probe (which python3 nc socat strace) is exempt"
fi
grep -q "call-cell" "$AUDRR/redlines.jsonl" 2>/dev/null \
  && ok "an actual nc invocation still reds (RED DEMO)" \
  || bad "the real nc call escaped the audit"

# ------------------------------------------------------------------
step "7. the broker: ask -> answer journals the decision wait"
mkdir -p "$SYN/operators"
if bash "$HERE/broker.sh" "$SYN" conduit-rust-top ask "which exit code?" --recommend "exit 1" >"$SCR/broker.log" 2>&1; then
  grep -q "OPERATOR PROMPT" "$SCR/broker.log" && ok "broker composed the operator prompt" \
    || bad "broker prompt missing"
else
  bad "broker ask failed"
fi
sleep 1
if bash "$HERE/broker.sh" "$SYN" conduit-rust-top answer "do exit 1" >>"$SCR/broker.log" 2>&1; then
  python3 - "$SYN" <<'PY' && ok "the decision wait was journaled (>=1s)" || bad "no wait journaled"
import json, sys
w = json.load(open(f"{sys.argv[1]}/operators/decision-waits-conduit-rust-top.json"))
assert w and w[-1]["wait_s"] >= 1.0, w
PY
else
  bad "broker answer failed"
fi

# ------------------------------------------------------------------
step "8. the agent registry: full uuid forms, id8 pins, closed census"
REGRR="$SCR/reg-root"
mkdir -p "$REGRR" "$REGRR/manifests" "$REGRR/tasks/conduit-rust-solo"
printf '{"sense":"conduit","stack":"rust","hand":"solo","cell_path":"%s/x","class":"main"}\n' "$REGRR" \
  > "$REGRR/manifests/conduit-rust-solo.json"
if bash "$HERE/register-agent.sh" "$REGRR" conduit-rust-solo hand adf6b7bd-aa93-4268-b786-f2d1fc561e94 \
     --session sess_subagent_agent_adf6b7bd-aa93-4268-b786-f2d1fc561e94 >"$SCR/reg1.log" 2>&1; then
  ok "the bare 36-char uuid registers"
else
  bad "bare uuid refused:"; tail -2 "$SCR/reg1.log" | sed 's/^/    /'
fi
grep -qx 'adf6b7bd=sess_subagent_agent_adf6b7bd-aa93-4268-b786-f2d1fc561e94' "$REGRR/session-pins.txt" \
  && ok "the pin line keys the 8-hex id8 (the ledger's own key)" \
  || bad "the pin line is not the id8"
bash "$HERE/register-agent.sh" "$REGRR" conduit-rust-top hand agent_0123abcd4321dcba9876fedcba012345 \
  --session sess_subagent_agent_0123abcd4321dcba9876fedcba012345 >/dev/null 2>&1 \
  && grep -qx '0123abcd=sess_subagent_agent_0123abcd4321dcba9876fedcba012345' "$REGRR/session-pins.txt" \
  && ok "the agent_<uuid> form registers and pins by id8 too" \
  || bad "agent_<uuid> form or its pin broke"
demo "a duplicate uuid must be refused"
if bash "$HERE/register-agent.sh" "$REGRR" conduit-rust-top operator adf6b7bd-aa93-4268-b786-f2d1fc561e94 >/dev/null 2>&1; then
  bad "duplicate uuid registered"
else
  ok "duplicate uuid refused (RED DEMO)"
fi
demo "a second hand on a cell must be refused"
if bash "$HERE/register-agent.sh" "$REGRR" conduit-rust-solo hand agent_9999aaaa8888bbbb7777ccccdddd0000 >/dev/null 2>&1; then
  bad "a second hand registered onto the cell"
else
  ok "second hand refused (RED DEMO)"
fi
demo "a mangled uuid must be refused"
if bash "$HERE/register-agent.sh" "$REGRR" conduit-rust-top judge not-a-uuid >/dev/null 2>&1; then
  bad "mangled uuid registered"
else
  ok "mangled uuid refused (RED DEMO)"
fi

echo
if [ "$FAILS" = 0 ]; then
  echo "SELFTEST GREEN — every gate green, every RED DEMO bit"
  exit 0
fi
echo "SELFTEST RED: $FAILS failure(s)"
exit 1
