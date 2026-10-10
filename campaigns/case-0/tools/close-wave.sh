#!/bin/bash
# close-wave.sh — the ONE-COMMAND closing pipeline of a case-0
# wave (the 0.2.0 closure discipline): from
# the finished cells to the wave report with every number placed by
# a frozen tool — zero driver arithmetic, zero manual placement.
#
#   1. usage-ledger.py --root    — the platform db ledger of every
#      registered agent (the session pins of the micro-agents ride
#      session-pins.txt)
#   1b. CENSUS GATE — every subagent session the platform created
#      inside the wave's window must be a pinned census member: an
#      unpinned session is an UNREGISTERED AGENT (the census hole:
#      the cpp-top acceptance operator fell out of every table) —
#      red line and STOP; the driver registers the agent and reruns
#      the close
#   2. collect-metrics.sh xN     — the per-cell evidence (the 0.2.0
#      file set; idempotent)
#   2b. the measurement evidence — measurement/measure-*.json and
#      judge-*.json move into tasks/<cell>/evidence/ BEFORE the
#      report (the lesson: "no measure.json captured yet"
#      and a manual cp)
#   3. canaries.sh scan          — the canary leaks of every product
#      tree and transcript export
#   4. audit-net.sh              — the post-factum network audit
#   5. report.py                 — the wave report (the tables of
#      the DOUBLE token base, honest multiples);
#      with --docs and a written <root>/analytics.md the tool ALSO
#      glues the session report into docs/ itself (no manual cat)
#   5b. verify-metrics.py        — the per-cell printed verification
#      (census<->pins<->db<->ledger<->every table cell with its full
#      source path; Σ re-added; camps independently recounted) — to
#      the screen and into <root>/verification/
#   6. the watchers of this root stop (the wave owns their life)
#
# Usage: close-wave.sh <run-root> [--prices <prices.json>]
#                      [--docs <repo-docs-dir>]
# Env: C0_DB (default ~/.zcode/cli/db/db.sqlite)
set -euo pipefail
RR="${1:?run root}"; shift || true
PRICES=""; DOCS=""
while [ $# -gt 0 ]; do case "$1" in --prices) PRICES="$2"; shift 2;; --docs) DOCS="$2"; shift 2;; *) echo "unknown: $1" >&2; exit 2;; esac; done
HERE="$(cd "$(dirname "$0")" && pwd)"
CAMP="$(cd "$HERE/.." && pwd)"
RR="$(readlink -f "$RR")"
[ -f "$RR/driver-uuid.txt" ] || { echo "CLOSE REFUSED: not a driver run root" >&2; exit 2; }
PRICES="${PRICES:-$CAMP/prices.json}"

echo "== close-wave: $RR"

echo "-- usage ledger"
# the pins file lines are id8=sess; feed each as its own --session
PINS_ARGS=()
if [ -s "$RR/session-pins.txt" ]; then
  while IFS= read -r line; do
    [ -n "$line" ] && PINS_ARGS+=(--session "$line")
  done < "$RR/session-pins.txt"
fi

# LEDGER SETTLE GATE (the owner's
# hard-audit word): the platform BACK-FILLS model_usage rows of a
# finished session minutes after its last request — a ledger read at
# close time can silently understate agents. The close therefore never trusts a
# single fresh read: a baseline extraction at close, then re-extracts
# after a settle pause until two consecutive reads are byte-identical
# (extraction is deterministic, no clock, no RNG); instability is a
# red line and the loop repeats (evidence collection below is
# idempotent, the report regenerates — the pipeline converges on the
# settled truth).
python3 "$HERE/usage-ledger.py" --root "$RR" --db "${C0_DB:-$HOME/.zcode/cli/db/db.sqlite}" "${PINS_ARGS[@]+"${PINS_ARGS[@]}"}"
SETTLE_TRIES=0
while : ; do
  SETTLE_TRIES=$((SETTLE_TRIES + 1))
  sleep 60
  cp "$RR/usage/ledger.json" "$RR/usage/ledger.prev.json"
  python3 "$HERE/usage-ledger.py" --root "$RR" --db "${C0_DB:-$HOME/.zcode/cli/db/db.sqlite}" "${PINS_ARGS[@]+"${PINS_ARGS[@]}"}"
  if cmp -s "$RR/usage/ledger.json" "$RR/usage/ledger.prev.json"; then
    echo "  ledger settled ($((SETTLE_TRIES + 1)) extraction(s), last two byte-identical)"
    break
  fi
  echo "  ledger UNSETTLED (try $SETTLE_TRIES) — the db still back-fills usage rows; re-extracting"
  python3 - "$RR" <<'PY'
import datetime, json, sys
root = sys.argv[1]
rec = {"id": f"ledger-backfill-{datetime.datetime.now(datetime.timezone.utc).strftime('%H%M%S')}",
       "cell": "(driver)", "class": "accounting",
       "text": "close-wave settle gate: the ledger extraction changed between two consecutive reads — the platform back-filled model_usage rows after session end; evidence/report will be rebuilt from the settled ledger",
       "ts": datetime.datetime.now(datetime.timezone.utc).isoformat(timespec="seconds"),
       "disposition": None}
with open(f"{root}/redlines.jsonl", "a") as f:
    f.write(json.dumps(rec, ensure_ascii=False) + "\n")
PY
  if [ "$SETTLE_TRIES" -ge 6 ]; then
    echo "CLOSE RED: the ledger never settled in 6 tries — close anyway on the last extraction, the red lines above carry the instability for the owner" >&2
    break
  fi
done

echo "-- census gate: every subagent session of the wave's window is a pinned census member"
# the census hole (the owner's hard-audit word): an acceptance
# operator spawned but never registered fell out of every table.
# session.time_created is epoch-MILLISECONDS — parsed as ms. An
# unpinned session in the window is an unregistered agent: red line
# and STOP (the driver registers it, the ledger re-extracts, the
# close reruns).
CENSUS_RED=0
python3 - "$RR" "${C0_DB:-$HOME/.zcode/cli/db/db.sqlite}" <<'PY' || CENSUS_RED=1
import datetime, json, os, sqlite3, sys
root, db = sys.argv[1:3]
start_ms = int(os.stat(os.path.join(root, "driver-uuid.txt")).st_mtime * 1000)
pinned = set()
pinf = os.path.join(root, "session-pins.txt")
if os.path.isfile(pinf):
    for l in open(pinf, encoding="utf-8"):
        if "=" in l:
            pinned.add(l.split("=", 1)[1].strip())
con = sqlite3.connect(f"file:{db}?mode=ro", uri=True)
rows = con.execute("SELECT id, time_created FROM session"
                   " WHERE id LIKE 'sess_subagent_agent_%'").fetchall()
con.close()
unpinned = sorted(sid for sid, tc in rows if (tc or 0) >= start_ms
                  and sid not in pinned)
if not unpinned:
    print(f"  census clean: {len([1 for sid, tc in rows if (tc or 0) >= start_ms])} "
          f"subagent session(s) in the window, all pinned")
    sys.exit(0)
for sid in unpinned:
    rec = {"id": f"census-unpinned-{sid[-12:]}",
           "cell": "(driver)", "class": "census",
           "text": f"unpinned subagent session in the wave window: {sid} — "
                   "an unregistered agent (register it, re-extract the ledger, "
                   "rerun the close)",
           "ts": datetime.datetime.now(datetime.timezone.utc).isoformat(timespec="seconds"),
           "disposition": None}
    with open(os.path.join(root, "redlines.jsonl"), "a") as f:
        f.write(json.dumps(rec, ensure_ascii=False) + "\n")
    print(f"  RED: unpinned session in the wave window: {sid}", file=sys.stderr)
sys.exit(1)
PY
if [ "$CENSUS_RED" = 1 ]; then
  echo "CLOSE STOP: the census gate found an unregistered agent — register it, re-extract the ledger, rerun the close" >&2
  exit 1
fi

echo "-- per-cell evidence"
for M in "$RR"/manifests/*.json; do
  [ -f "$M" ] || continue
  CELL="$(basename "$M" .json)"
  if [ -f "$RR/tasks/$CELL/evidence/platform.json" ]; then
    echo "  $CELL: already collected (idempotent)"
    continue
  fi
  echo "  $CELL"
  bash "$HERE/collect-metrics.sh" "$RR" "$CELL" \
    $( [ -f "$RR/operators/decision-waits-$CELL.json" ] || echo --accept "operator acceptance (journaled in the wave journal)" ) \
    || echo "  RED: $CELL collect failed — the wave report carries the hole"
done

echo "-- measurement evidence into the cells (BEFORE the report)"
# measure.json/judge.json live in measurement/
# until the close moves them into the cells' evidence/ — the report
# printed "no measure.json captured yet" until a manual cp fixed it
N_MEAS=0
for SRC in "$RR"/measurement/measure-*.json; do
  [ -f "$SRC" ] || continue
  B="$(basename "$SRC" .json)"; CELL="${B#measure-}"
  EV="$RR/tasks/$CELL/evidence"; mkdir -p "$EV"
  if ! cmp -s "$SRC" "$EV/measure.json"; then
    cp -f "$SRC" "$EV/measure.json"; N_MEAS=$((N_MEAS + 1))
  fi
done
for SRC in "$RR"/measurement/judge-*.json; do
  [ -f "$SRC" ] || continue
  B="$(basename "$SRC" .json)"; CELL="${B#judge-}"
  EV="$RR/tasks/$CELL/evidence"; mkdir -p "$EV"
  if ! cmp -s "$SRC" "$EV/judge.json"; then
    cp -f "$SRC" "$EV/judge.json"; N_MEAS=$((N_MEAS + 1))
  fi
done
echo "  measurement evidence placed/refreshed: $N_MEAS file(s)"

echo "-- canary scan"
CTRL="$RR/measurement/sealed"
if [ -f "$CTRL/canaries/registry.txt" ]; then
  for M in "$RR"/manifests/*.json; do
    [ -f "$M" ] || continue
    CELL="$(basename "$M" .json)"
    CELLPATH="$(python3 -c "import json; print(json.load(open('$M'))['cell_path'])")"
    OUT="$(bash "$HERE/canaries.sh" scan "$CTRL" --not "$CELLPATH" "$CELLPATH/work" "$RR/rollouts" 2>/dev/null || echo '{}')"
    LEAKS="$(printf '%s' "$OUT" | python3 -c 'import json,sys; d=json.load(sys.stdin); print(len(d.get("leaks",[])))' 2>/dev/null || echo "?")"
    if [ "$LEAKS" != "0" ]; then
      python3 - "$RR" "$CELL" "$LEAKS" <<'PY'
import datetime, json, sys
root, cell, n = sys.argv[1:4]
rec = {"id": f"canary-leak-{cell}", "cell": cell, "class": "canary",
       "text": f"{n} canary leak(s) in the cell's product/transcripts",
       "ts": datetime.datetime.now(datetime.timezone.utc).isoformat(timespec="seconds"),
       "disposition": None}
with open(f"{root}/redlines.jsonl", "a") as f:
    f.write(json.dumps(rec, ensure_ascii=False) + "\n")
PY
      echo "  RED: $CELL canary leaks: $LEAKS"
    else
      echo "  $CELL: clean"
    fi
  done
else
  echo "  (no canary registry — a wave without canaries is a red line)"
fi

echo "-- canary bait sweep (ONLY the wave's OWN bait files, each an exact"
echo "   registry line AND a canary-name pattern; anything else is skipped)"
if [ -f "$CTRL/canaries/registry.txt" ]; then
  N=0
  while read -r tok path; do
    case "${path:-}" in
      */.canary-[0-9a-f]*.txt) rm -f -- "$path" && N=$((N + 1)) ;;
      *) echo "  SKIP (not a canary-pattern path): $path" ;;
    esac
  done < "$CTRL/canaries/registry.txt"
  echo "  swept $N bait file(s); the registry record is kept: $CTRL/canaries/registry.txt"
fi

echo "-- network audit"
bash "$HERE/audit-net.sh" "$RR" || true

echo "-- wave report"
GLUE_ARGS=()
if [ -n "$DOCS" ] && [ -f "$RR/analytics.md" ]; then
  # the glue is the tool's own act (the owner saw a stale table in
  # docs/ after a manual cat): the core +
  # the analytics layer land in docs/case0-<ts>-<sense>-session-report.md
  GLUE_ARGS=(--glue "$RR/analytics.md" --glue-out "$DOCS")
elif [ -n "$DOCS" ]; then
  echo "  (no $RR/analytics.md yet — the glue step will run when the driver writes it and re-closes)"
fi
python3 "$HERE/report.py" "$RR" --prices "$PRICES" --models "$CAMP/models.json" \
  --out "$RR/wave-report.md" ${GLUE_ARGS[@]+"${GLUE_ARGS[@]}"} > /dev/null
echo "report: $RR/wave-report.md"

echo "-- metrics verification (per-cell, full source paths)"
if python3 "$HERE/verify-metrics.py" "$RR" --db "${C0_DB:-$HOME/.zcode/cli/db/db.sqlite}" \
     --report "$RR/wave-report.md"; then
  echo "  verification GREEN (the printed copy is also in $RR/verification/)"
else
  echo "  RED: the metrics verification found holes — see the lines above and $RR/verification/" >&2
  python3 - "$RR" <<'PY'
import datetime, json, sys
root = sys.argv[1]
rec = {"id": f"verify-metrics-red-{datetime.datetime.now(datetime.timezone.utc).strftime('%H%M%S')}",
       "cell": "(driver)", "class": "accounting",
       "text": "verify-metrics.py reported RED lines — a table cell diverged from its db->ledger->evidence path",
       "ts": datetime.datetime.now(datetime.timezone.utc).isoformat(timespec="seconds"),
       "disposition": None}
with open(f"{root}/redlines.jsonl", "a") as f:
    f.write(json.dumps(rec, ensure_ascii=False) + "\n")
PY
fi

echo "-- watchers"
PIDS="$(pgrep -f "rollout-watch\.sh.*$RR" || true)"
if [ -n "$PIDS" ]; then
  kill $PIDS 2>/dev/null || true
  echo "watchers stopped: $(echo $PIDS | tr '\n' ' ')"
else
  echo "no watchers bound to this root"
fi

echo "-- red lines summary"
if [ -f "$RR/redlines.jsonl" ]; then
  python3 - "$RR" <<'PY'
import json, sys
root = sys.argv[1]
n0 = n1 = 0
for line in open(f"{root}/redlines.jsonl", encoding="utf-8"):
    if not line.strip():
        continue
    d = json.loads(line)
    if d.get("disposition") is None:
        n0 += 1
        print(f"  OPEN {d.get('id')}: {d.get('text', '')[:120]}")
    else:
        n1 += 1
print(f"red lines: {n0} open, {n1} dispositioned — every open line needs an owner adjudication")
PY
else
  echo "  (no red lines)"
fi

echo "== close-wave done"
