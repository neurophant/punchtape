#!/bin/bash
# rollout-watch.sh — the wave driver's rollout discipline, generalized
# from the wave-1 watcher scripts (the battle-tested mechanics, one
# reusable tool instead of per-wave copies in the scratch root).
#
# TWO jobs in one loop (one pass per minute):
#   snapshot — every active rollout file is copied (atomic tmp+mv) into
#     the out directory under its key. The platform trims rollouts
#     faster than a finish-time snapshot can catch (7 of 15 cells
#     lost the in/out split in wave 1); a rolling copy loses at most
#     one minute. The finish-time snapshot then just lands on the last
#     copy. Rotation-tolerant: the file may VANISH and REAPPEAR on a
#     turn boundary while the hand still works — each cycle re-globs
#     the key's id, a fresh file resumes the watch.
#   watch — a rollout that stops growing for --stall minutes, or
#     disappears, means the agent's turn ended — often WITHOUT a
#     completion notification (~13 dropped turns per wave). The
#     event is ONE log line (TURN-END/TURN-DROP); the watcher KEEPS
#     WATCHING the other keys and re-admits a reappearing file, so one
#     event never blinds the driver's view of the other five hands.
#     The driver pokes the hand immediately ("continue from where you
#     stopped") — after checking the instance: a stop on exit code 3/4
#     is the hand's question channel (its final message awaits
#     the driver-operator), not a dropped turn (stage state,
#     PendingAmend, the ledger tail); the poke is protocol-only, never
#     task content.
#
# Usage:
#   rollout-watch.sh --out <dir> --map <file> [--cycles N] [--stall M]
#     --map  — a file of KEY=ID lines (one per line; # comments and
#              blanks skipped). The ONLY way to name keys: cell
#              names never ride argv (the 2x2 ps-leak finding —
#              the watcher's command line was readable by hands).
#     KEY    — the key the snapshot is saved under (e.g. a cell id)
#     ID     — the subagent id (the rollout file matches *<ID>*.jsonl)
#   The rollout store root can be overridden: ROLLOUT_DIR=...
#   --driver <run-root>: TURN-END/TURN-DROP events are fed
#     to the driver (driver.sh poke) — the auto red line
#     platform-turn-drop + frame, no manual relaying.
#   --cycles 0 = NO LIMIT (the default; the watcher never
#     goes blind while the wave runs).
#
# Exit codes: 0 — window elapsed; 2 — bad usage / nothing found at start.

set -u

OUT=""; CYCLES=0; STALL=5; DRIVER=""; MAP=""
declare -a KEYS=() IDS=()
while [ $# -gt 0 ]; do
  case "$1" in
    --out)    OUT="$2"; shift 2 ;;
    --cycles) CYCLES="$2"; shift 2 ;;
    --driver) DRIVER="$2"; shift 2 ;;
    --stall)  STALL="$2"; shift 2 ;;
    --map)    MAP="$2"; shift 2 ;;
    *)        echo "usage: $0 --out <dir> --map <file> [--cycles N] [--stall M]"; exit 2 ;;
  esac
done
if [ -z "$OUT" ] || [ -z "$MAP" ]; then
  echo "usage: $0 --out <dir> --map <file>"; exit 2
fi
while IFS='=' read -r k id; do
  case "$k" in ''|'#'*) continue ;; esac
  KEYS+=("$k"); IDS+=("$id")
done < "$MAP"
if [ ${#KEYS[@]} -eq 0 ]; then
  echo "usage: $0 --map file carries no KEY=ID lines"; exit 2
fi
mkdir -p "$OUT"
ROLLOUT_DIR="${ROLLOUT_DIR:-$HOME/.zcode/cli/rollout}"

declare -A R GONE STALLC LINES ROT
for i in "${!KEYS[@]}"; do
  k="${KEYS[$i]}"
  f="$(ls "$ROLLOUT_DIR"/model-io-*"${IDS[$i]}"*.jsonl 2>/dev/null | head -1)"
  if [ -z "$f" ]; then echo "WARN no rollout at start: $k (will re-glob each cycle)"; fi
  R[$k]="$f"; GONE[$k]=1; STALLC[$k]=0; LINES[$k]=0; ROT[$k]=0
done
echo "watching ${#KEYS[@]} rollout(s): ${KEYS[*]} -> $OUT"

maybe_poke() {  # $1 = key, $2 = reason — the serial 0.2.0 protocol has
  # no driver.sh: the event stays a log line the driver session reads
  # (the serial wave watches one hand at a time)
  :
}
i=0
while :; do
  i=$((i+1))
  if [ "$CYCLES" -gt 0 ] && [ "$i" -gt "$CYCLES" ]; then break; fi
  sleep 60
  for j in "${!KEYS[@]}"; do
    k="${KEYS[$j]}"
    f="$(ls "$ROLLOUT_DIR"/model-io-*"${IDS[$j]}"*.jsonl 2>/dev/null | head -1)"
    if [ -n "$f" ]; then
      if [ "${GONE[$k]}" = 1 ]; then
        ROT[$k]=$(( ${ROT[$k]} + 1 ))
        echo "RESUME: $k (rollout reappeared, rotation ${ROT[$k]}: $f)"
        GONE[$k]=0; STALLC[$k]=0
      fi
      R[$k]="$f"
      dst="$OUT/$k.jsonl"
      if [ "${ROT[$k]}" -gt 0 ]; then dst="$OUT/$k.r${ROT[$k]}.jsonl"; fi
      cp "$f" "$OUT/.$k.tmp" && mv "$OUT/.$k.tmp" "$dst"
      n=$(wc -l < "$f" 2>/dev/null || echo 0)
      if [ "$n" = "${LINES[$k]}" ]; then STALLC[$k]=$(( ${STALLC[$k]} + 1 )); else STALLC[$k]=0; fi
      LINES[$k]=$n
      if [ "${STALLC[$k]}" -ge "$STALL" ]; then
        echo "TURN-DROP: $k (stalled ${STALL}min at ${LINES[$k]} lines) — poke the hand; snapshot current in $OUT/$k.jsonl"
        maybe_poke "$k" "rollout stalled ${STALL}min at ${LINES[$k]} lines"
        STALLC[$k]=0
      fi
    else
      if [ "${GONE[$k]}" = 0 ]; then
        echo "TURN-END: $k (rollout gone; last snapshot $OUT/$k.jsonl) — check the instance, poke only if dropped"
        GONE[$k]=1
      fi
    fi
  done
done
echo "window elapsed"
exit 0
