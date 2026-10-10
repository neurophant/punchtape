#!/bin/bash
# finish-cell.sh — the FINISH CHECKLIST of a case-0 cell (the
# 0.2.0 discipline: the executor's metrics are captured AT THE
# MOMENT THE HAND FINISHES — a lost cell stays lost): first the
# cell's leftover processes are extinguished by cwd (an old
# series' port holder once survived its own finish), then the cell-scoped ledger SETTLE runs (the platform
# back-fills model_usage rows minutes after a session's last
# request — a single fresh read understates the evidence; the same
# law as the close-wave settle gate, applied at EVERY finish, scoped
# to THIS cell's agents so the conveyor's other cells may keep
# moving), then the evidence capture (collect-metrics.sh). The
# machine hand's verdict is a necessary, not a sufficient condition;
# the solo hand's stop is the operator's acceptance.
#
# Usage: finish-cell.sh <run-root> <cell> [--accept "<the operator's
#        acceptance note>"] [--verdict-from-machine]
# Env:  C0_DB (default ~/.zcode/cli/db/db.sqlite)
set -euo pipefail
RR="${1:?run root}"; CELL="${2:?cell}"; shift 2
ACCEPT=""; VFM=""
while [ $# -gt 0 ]; do
  case "$1" in
    --accept) ACCEPT="$2"; shift 2 ;;
    --verdict-from-machine) VFM=1; shift ;;
    *) echo "unknown: $1" >&2; exit 2 ;;
  esac
done
HERE="$(cd "$(dirname "$0")" && pwd)"
RR="$(readlink -f "$RR")"
[ -f "$RR/manifests/$CELL.json" ] || { echo "FINISH REFUSED: no manifest for $CELL" >&2; exit 2; }

# --- 0. extinguish the cell's leftover processes (a port holder
# --- of a CLOSED series can outlive its finish if the finisher does
# --- not sweep by cwd): every process
# --- whose cwd lies inside THIS cell dies now — the hand's session
# --- is over, anything still sitting there is a leak
CELLPATH="$(python3 -c "import json,sys; print(json.load(open('$RR/manifests/$CELL.json'))['cell_path'])")"
python3 - "$CELLPATH" <<'PY'
import os, signal, sys
cell = os.path.realpath(sys.argv[1])
me = os.getpid()
killed = []
for pid in os.listdir("/proc"):
    if not pid.isdigit() or int(pid) == me:
        continue
    try:
        cwd = os.path.realpath(os.readlink(f"/proc/{pid}/cwd"))
        if cwd == cell or cwd.startswith(cell + os.sep):
            os.kill(int(pid), signal.SIGKILL)
            killed.append(pid)
    except (OSError, PermissionError):
        continue
msg = f"finish-cell: extinguished {len(killed)} leftover process(es) of the cell"
if killed:
    msg += " (" + ", ".join(killed) + ")"
print(msg)
PY

# the declared product must be resolvable before the receipt
python3 - "$RR" "$CELL" <<'PY'
import json, os, sys
root, cell = sys.argv[1:3]
m = json.load(open(f"{root}/manifests/{cell}.json"))
work = os.path.join(m["cell_path"], "work")
hand = m["hand"]
if hand == "solo":
    ok = os.path.isdir(os.path.join(work, "solo-product"))
    why = "work/solo-product"
else:
    led = os.path.join(work, ".punchtape", "ledger.yamll")
    ok = os.path.isfile(led)
    why = "work/.punchtape/ledger.yamll (the machine instance)"
print(f"declared product: {'RESOLVES (' + why + ')' if ok else 'UNRESOLVED (' + why + ')'}")
sys.exit(0 if ok else 3)
PY

# --- 1. the CELL-SCOPED LEDGER SETTLE (the platform law,
# --- the same gate close-wave carries, applied at every finish):
# --- re-extract until two consecutive reads of THIS cell's agent
# --- slices are byte-identical — the extraction is deterministic
# --- (no clock, no RNG), so a change between reads means the db is
# --- still back-filling usage rows of the just-finished sessions;
# --- collecting on an unsettled slice understates the evidence
python3 - "$RR" "$CELL" "$HERE" <<'PY'
import datetime, json, os, re, sqlite3, subprocess, sys, time

root, cell, here = sys.argv[1:4]
db = os.environ.get("C0_DB",
                    os.path.expanduser("~/.zcode/cli/db/db.sqlite"))
reg = json.load(open(f"{root}/driver-agents.json")) \
    if os.path.isfile(f"{root}/driver-agents.json") else {}
ids = sorted({m.group(1).lower()
              for e in (reg.get(cell) or [])
              if (m := re.search(r"([0-9a-fA-F]{8})", e.get("uuid") or ""))})
if not ids:
    print(f"FINISH REFUSED: no registered agents for {cell}", file=sys.stderr)
    sys.exit(2)

def redline(text):
    rec = {"id": f"cell-settle-{cell}-{datetime.datetime.now(datetime.timezone.utc).strftime('%H%M%S')}",
           "cell": cell, "class": "accounting", "text": text,
           "ts": datetime.datetime.now(datetime.timezone.utc).isoformat(timespec="seconds"),
           "disposition": None}
    with open(f"{root}/redlines.jsonl", "a") as f:
        f.write(json.dumps(rec, ensure_ascii=False) + "\n")

def extract():
    pins = []
    pinf = f"{root}/session-pins.txt"
    if os.path.isfile(pinf):
        pins = [l.strip() for l in open(pinf, encoding="utf-8") if l.strip()]
    cmd = [sys.executable, f"{here}/usage-ledger.py", "--root", root,
           "--db", db]
    for p in pins:
        cmd += ["--session", p]
    r = subprocess.run(cmd, capture_output=True, text=True)
    if r.returncode != 0:
        print(f"FINISH RED: usage-ledger failed: {r.stderr[:200]}", file=sys.stderr)
        sys.exit(3)

def slice_now():
    led = json.load(open(f"{root}/usage/ledger.json"))
    agents = led.get("agents") or {}
    return json.dumps({i: agents.get(i) for i in ids}, sort_keys=True)

extract()
prev = slice_now()
tries = 0
while True:
    tries += 1
    time.sleep(60)
    extract()
    cur = slice_now()
    if cur == prev:
        print(f"finish-cell: cell ledger settled ({tries + 1} extraction(s), "
              f"last two byte-identical for {cell})")
        break
    prev = cur
    redline(f"finish-cell settle gate ({cell}): the cell's ledger slice changed "
            "between two consecutive reads — the platform back-filled "
            "model_usage rows after session end; the evidence is collected "
            "from the settled slice")
    print(f"  cell ledger UNSETTLED (try {tries}) — re-extracting", file=sys.stderr)
    if tries >= 6:
        print(f"FINISH RED: the cell ledger of {cell} never settled in 6 tries — "
              "collect on the last extraction; the close settle gate re-checks "
              "the whole wave", file=sys.stderr)
        break
PY

# --- 2. the evidence capture (the acceptance verdict freezes at the
# --- first finish: collect-metrics keeps it in evidence/acceptance.json
# --- and a later --accept service note cannot overwrite it)
bash "$HERE/collect-metrics.sh" "$RR" "$CELL" --accept "${ACCEPT:-}" \
  $([ -n "$VFM" ] && echo --verdict-from-machine)

echo "finish-cell: $CELL done — the acceptance rides the wave's operator events"
