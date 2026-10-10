#!/bin/bash
# broker.sh — the RESPONSIVENESS BROKER of the case-0 wave (the
# 0.2.0 lead-session protocol): nobody polls, nobody sleeps —
# a hand's turn that ends on a question batch / escalation hands the
# question to the OPERATOR MICRO-AGENT; the broker composes its
# prompt (the frozen OPERATOR-MACHINE-HOST / OPERATOR-SOLO bytes +
# the question), the driver pastes it into a fresh micro-agent
# session, the answer returns to the hand as its next message. Every
# decision's wait (question returned -> answer sent) is journaled
# into the cell's decision_waits.json (the platform.json
# decision_wait_s source). The operator is a separate agent: its
# tokens are its own column, never the hand's.
#
# Failure protocol: one retry; a repeated failure is a red line and
# the delivered answer is the ask's recommendation — the hand never
# waits in silence.
#
# Usage:
#   broker.sh <run-root> <cell> ask "<question>" [--recommend "<text>"]
#        — prints the composed operator prompt; journals the ask
#   broker.sh <run-root> <cell> answer "<text>"
#        — journals the answer + the wait; prints the delivery line
#   broker.sh <run-root> <cell> failed [--attempt N] [--reason "<text>"]
#        — retry once; the second failure: red line + the default
set -euo pipefail
RR="${1:?run root}"; CELL="${2:?cell}"; CMD="${3:?ask|answer|failed}"; shift 3
REC=""; WHY=""; ATTEMPT=1
while [ $# -gt 0 ]; do
  case "$1" in
    --recommend) REC="$2"; shift 2 ;;
    --reason) WHY="$2"; shift 2 ;;
    --attempt) ATTEMPT="$2"; shift 2 ;;
    -*) echo "unknown: $1" >&2; exit 2 ;;
    *) Q="${Q:-}$1"; shift ;;
  esac
done
HERE="$(cd "$(dirname "$0")" && pwd)"
CAMP="$(cd "$HERE/.." && pwd)"
RR="$(readlink -f "$RR")"
[ -f "$RR/manifests/$CELL.json" ] || { echo "BROKER REFUSED: no manifest for $CELL" >&2; exit 2; }
HAND="$(python3 -c "import json; print(json.load(open('$RR/manifests/$CELL.json'))['hand'])")"
case "$HAND" in
  solo) ROLEFILE="$CAMP/prompts/OPERATOR-SOLO.md" ;;
  top) ROLEFILE="$CAMP/prompts-host/OPERATOR-MACHINE-HOST.md" ;;
esac
[ -f "$ROLEFILE" ] || { echo "BROKER RED: no operator prompt at $ROLEFILE" >&2; exit 2; }

EVENTS="$RR/operators/broker-events.jsonl"
WAITS="$RR/operators/decision-waits-$CELL.json"

log_event() {  # kind, then key=value pairs
  python3 - "$EVENTS" "$CELL" "$@" <<'PY'
import datetime, json, os, sys
path, cell, kind = sys.argv[1:4]
kv = {}
for a in sys.argv[4:]:
    if "=" in a:
        k, v = a.split("=", 1)
        kv[k] = v
kv.update({"cell": cell, "kind": kind,
           "ts": datetime.datetime.now(datetime.timezone.utc).isoformat(timespec="seconds")})
with open(path, "a", encoding="utf-8") as f:
    f.write(json.dumps(kv, ensure_ascii=False) + "\n")
PY
}

case "$CMD" in
ask)
  [ -n "${Q:-}" ] || { echo "broker ask needs the question text" >&2; exit 2; }
  log_event ask question="$Q" recommendation="${REC:-}"
  echo "===== OPERATOR PROMPT (paste into a fresh operator micro-agent of cell $CELL) ====="
  cat "$ROLEFILE"
  printf '\n---\n\nThe question from the executor (verbatim):\n\n%s\n\n' "$Q"
  if [ -n "$REC" ]; then
    printf 'The recommendation of the machine (for code 3 — a question batch; accept it unless the brief dictates otherwise):\n\n%s\n' "$REC"
  fi
  printf '\nAnswer with one decision in plain language — the driver will pass it to the executor verbatim.\n'
  echo "===== END OPERATOR PROMPT ====="
  echo "(after the operator answers: broker.sh $RR $CELL answer \"<text>\")"
  ;;
answer)
  [ -n "${Q:-}" ] || { echo "broker answer needs the answer text" >&2; exit 2; }
  python3 - "$EVENTS" "$WAITS" "$CELL" <<'PY'
import datetime, json, sys
events, waits, cell = sys.argv[1:4]
asks = [json.loads(l) for l in open(events, encoding="utf-8") if l.strip()]
asks = [e for e in asks if e.get("cell") == cell and e.get("kind") == "ask"]
if not asks:
    print("BROKER RED: answer without an open ask", file=sys.stderr)
    sys.exit(2)
last = asks[-1]
end = datetime.datetime.now(datetime.timezone.utc)
wait_s = (end - datetime.datetime.fromisoformat(last["ts"])).total_seconds()
w = json.load(open(waits)) if __import__("os").path.isfile(waits) else []
w.append({"ask_ts": last["ts"], "answer_ts": end.isoformat(timespec="seconds"),
          "wait_s": round(wait_s, 1)})
json.dump(w, open(waits, "w"), indent=1)
print(f"decision wait: {wait_s:.1f}s (journaled: {waits})")
PY
  log_event answer text="$Q"
  printf '===== DELIVER TO THE HAND (its next message, verbatim) =====\n%s\n===== END DELIVER =====\n' "$Q"
  ;;
failed)
  AT="${ATTEMPT:-1}"
  if [ "$AT" -ge 2 ]; then
    python3 - "$RR" "$CELL" <<'PY'
import datetime, json, sys
root, cell = sys.argv[1:3]
rec = {"id": f"broker-failed-{cell}",
       "cell": cell, "class": "operator",
       "text": "the operator micro-agent failed twice: the recommendation delivered as the default",
       "ts": datetime.datetime.now(datetime.timezone.utc).isoformat(timespec="seconds"),
       "disposition": None}
with open(f"{root}/redlines.jsonl", "a", encoding="utf-8") as f:
    f.write(json.dumps(rec, ensure_ascii=False) + "\n")
PY
    log_event failed-final reason="${WHY:-unknown}"
    echo "BROKER RED LINE: second failure — deliver the recommendation as the default answer" >&2
    exit 1
  fi
  log_event failed attempt="$AT" reason="${WHY:-unknown}"
  echo "broker: attempt $AT failed — ONE retry allowed (broker.sh ... failed --attempt 2)" >&2
  ;;
*) echo "cmd must be ask|answer|failed" >&2; exit 2 ;;
esac
