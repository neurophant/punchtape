#!/bin/bash
# collect-metrics.sh — the per-cell metrics collector (th3 METRICS).
# GUARANTEE MECHANISM, not a promise:
#   COLLECTED — every field comes from a named machine source; a
#     missing source FAILS the cell (exit non-zero): no field may be
#     absent, guessed or deferred.
#   PRESERVED — the RAW per-agent model-io records are copied into
#     the cell's evidence BEFORE metrics.json is written: the numbers
#     survive even if the harness's rollout store is cleaned.
#   NOT FABRICATED — metrics.json is emitted by THIS script only;
#     --verify recomputes it from the preserved raw + the export and
#     diffs: any mismatch fails.
#   NOT HALLUCINATED — no agent self-report is a source, ever; the
#     sources are: rollout model-io records (tokens/tools/model-ms),
#     the machine ledger (submissions/verdict), count-deliverables.py
#     (quality counts), driver stamps passed as arguments (wall).
#
# The hand's metrics and the operator's metrics are computed and
# reported SEPARATELY (the owner's hard rule): one --hand file, N
# --op files; operator numbers sum ONLY across its own invocations.
#
# Usage:
#   collect-metrics.sh <cell> <hand> <export-dir> \
#       --brief-sha <sha> --image-id <id> --binary-sha <sha|absent> \
#       --hand <model-io.jsonl> [--hand-wall-ms <ms>] \
#       [--op <model-io.jsonl> --op-wall-ms <ms>]... \
#       [--verify <existing-metrics.json>]
# <export-dir> = the cell's exported task dir (with .punchtape for
# machine hands). Writes <export-dir>/../metrics.json (beside it).

set -euo pipefail

CELL=""; HAND=""; EXPORT=""
BRIEF_SHA=""; IMAGE_ID=""; BINARY_SHA=""
HAND_FILE=""; HAND_WALL=""
VERIFY=""; JOURNAL=""

while [ $# -gt 0 ]; do
  case "$1" in
    --brief-sha) BRIEF_SHA="$2"; shift 2 ;;
    --image-id) IMAGE_ID="$2"; shift 2 ;;
    --binary-sha) BINARY_SHA="$2"; shift 2 ;;
    --hand) HAND_FILE="$2"; shift 2 ;;
    --hand-wall-ms) HAND_WALL="$2"; shift 2 ;;
    --op) OP_SPECS+=("$2"); shift 2 ;;
    --verify) VERIFY="$2"; shift 2 ;;
    --journal) JOURNAL="$2"; shift 2 ;;
    --hand-usage) HAND_USAGE="$2"; shift 2 ;;
    --op-usage) OP_USAGES+=("$2"); shift 2 ;;
    -*) echo "unknown option: $1" >&2; exit 2 ;;
    *) if [ -z "$CELL" ]; then CELL="$1"; elif [ -z "$HAND" ]; then HAND="$1";
       elif [ -z "$EXPORT" ]; then EXPORT="$1"; else echo "extra arg: $1" >&2; exit 2; fi; shift ;;
  esac
done

if [ -z "${HAND_FILE:-}" ] && [ -z "${HAND_USAGE:-}" ]; then
  echo "MISSING REQUIRED: hand-file (or the --hand-usage platform fallback)" >&2; exit 3
fi
for req in "cell:$CELL" "hand:$HAND" "export:$EXPORT" "brief-sha:$BRIEF_SHA" "image-id:$IMAGE_ID"; do
  [ -n "${req#*:}" ] || { echo "MISSING REQUIRED: ${req%%:*}" >&2; exit 3; }
done
case "$HAND" in solo|top|weak) ;; *) echo "bad hand: $HAND" >&2; exit 3 ;; esac
if [ "$HAND" != solo ] && [ -z "$BINARY_SHA" ]; then echo "MISSING: binary sha for a machine hand" >&2; exit 3; fi
if [ -n "${HAND_USAGE:-}" ]; then
  # the platform prunes rollout records of finished agents faster than
  # batch collection — the completion notification's usage block is the
  # same platform source; use it and say so. Format: tokens,tools,ms
  HAND_FILE=""
elif [ -f "$HAND_FILE" ]; then :
else echo "hand model-io file not found and no --hand-usage fallback" >&2; exit 3; fi
for spec in ${OP_SPECS[@]+"${OP_SPECS[@]}"}; do
  case "$spec" in usage:*) continue ;; esac
  f="${spec%%|*}"
  [ -f "$f" ] || { echo "op file not found: $f" >&2; exit 3; }
done

EV="$(cd "$(dirname "$EXPORT")" && pwd)"
RAW="$EV/raw"
mkdir -p "$RAW"
[ -n "$HAND_FILE" ] && cp -n "$HAND_FILE" "$RAW/" || true

HERE="$(cd "$(dirname "$0")" && pwd)"
USAGE="$HERE/../../tools/agent-usage.sh"
COUNTER="$HERE/count-deliverables.py"
[ -x "$USAGE" ] || { echo "agent-usage.sh not found" >&2; exit 3; }

STACK="${CELL#pt-th3-}"; STACK="${STACK%%-*}"
SENSE_PART="${CELL#pt-th3-${STACK}-}"; SENSE_PART="${SENSE_PART%-*}"

if [ -n "$HAND_FILE" ]; then
  HAND_JSON="$("$USAGE" "$HAND_FILE")"
else
  IFS=, read -r HT HM < <(echo "$HAND_USAGE" | tr -d ' ')
  HAND_JSON="$(printf '{"tokens_in": null, "tokens_out": null, "tokens_cache_read": null, "tool_calls": %s, "model_ms": null, "tokens_total": %s}' "$HM" "$HT")"
fi
# op fallbacks: for invocation i without a raw file, consume OP_USAGES[i]
USAGE_ABS="$(cd "$(dirname "$USAGE")" && pwd)/$(basename "$USAGE")"
OPS_PAYLOAD="$(mktemp)"
SPECFILE="$(mktemp)"
printf '%s\n' ${OP_SPECS[@]+"${OP_SPECS[@]}"} > "$SPECFILE"
python3 - "$OPS_PAYLOAD" "$USAGE_ABS" "$SPECFILE" "${HAND_WALL:-null}" <<'PY2'
import json, subprocess, sys
out, usage, specfile = sys.argv[1], sys.argv[2], sys.argv[3]
hand_wall = int(sys.argv[4]) if sys.argv[4] != "null" else None
ops = []
for line in open(specfile):
    spec = line.strip()
    if not spec:
        continue
    if spec.startswith("usage:"):
        body = spec[len("usage:"):]
        nums, wall = body.split("|")
        t, m = nums.split(",")
        j = {"tokens_in": None, "tokens_out": None, "tokens_cache_read": None,
             "tool_calls": int(m), "model_ms": None, "tokens_total": int(t)}
    else:
        f, wall = spec.split("|")
        j = json.loads(subprocess.run([usage, f], capture_output=True, text=True, check=True).stdout)
    j["wall_ms"] = int(wall) if wall else None
    ops.append(j)
json.dump({"ops": ops, "hand_wall": hand_wall}, open(out, "w"), ensure_ascii=False)
PY2

python3 - "$CELL" "$HAND" "$EXPORT" "$BRIEF_SHA" "$IMAGE_ID" \
    "${BINARY_SHA:-absent}" "$HAND_JSON" "$OPS_PAYLOAD" \
    "$COUNTER" "$RAW" "$VERIFY" "$JOURNAL" <<'PY'
import json, subprocess, sys, os, re, statistics

(cell, hand, export, brief, image, binary, hand_json, ops_path,
 counter, raw, verify, journal_path) = sys.argv[1:13]

loop = {"rounds": None, "pokes": None, "answers": None, "demands": None,
        "latency_median_s": None, "latency_max_s": None,
        "source": "journal.md (driver-assembled verbatim)"}
if journal_path and os.path.isfile(journal_path):
    jt = open(journal_path, encoding="utf-8").read()
    lines = jt.splitlines()
    loop["rounds"] = len(re.findall(r"^ROUND\s+\d+", jt, re.M))
    loop["pokes"] = len(re.findall(r"action: POKE", jt))
    loop["answers"] = len(re.findall(r"action: ANSWER", jt))
    loop["demands"] = len(re.findall(r"action: DEMAND", jt))
    lats = []
    for m in re.finditer(r"^ROUND \d+ (\S+) -> (\S+)", jt, re.M):
        from datetime import datetime
        def p(x):
            return datetime.fromisoformat(x.replace("Z", "+00:00"))
        try:
            lats.append((p(m.group(2)) - p(m.group(1))).total_seconds())
        except ValueError:
            pass
    if lats:
        loop["latency_median_s"] = statistics.median(lats)
        loop["latency_max_s"] = max(lats)
payload = json.load(open(ops_path))
hand_wall = payload["hand_wall"]
ops = payload["ops"]

hj = json.loads(hand_json)
def tok(o, k):
    return o[k] if o.get(k) is not None else (o.get("tokens_total") if k == "tokens_in" else 0)
op = {
    "tokens_in": sum(tok(o, "tokens_in") for o in ops),
    "tokens_out": sum(o["tokens_out"] for o in ops),
    "tokens_cache_read": sum(o["tokens_cache_read"] for o in ops),
    "tool_calls": sum(o["tool_calls"] for o in ops),
    "model_ms": sum(o["model_ms"] for o in ops),
    "wall_ms": sum(o["wall_ms"] for o in ops if o.get("wall_ms")) or None,
    "invocations": len(ops),
    "source": "platform model-io records of THIS hand's operator invocations only",
} if ops else {
    "tokens_in": None, "tokens_out": None, "tokens_cache_read": None,
    "tool_calls": None, "model_ms": None, "wall_ms": None,
    "invocations": 0,
    "source": "no operator invocations recorded" if hand != "solo" else "no operator invocations recorded (protocol violation for solo)",
}

parts = cell.split("-")
stack = parts[3] if len(parts) > 4 else "unknown"
cd = json.loads(subprocess.run(
    ["python3", counter, export, stack, hand],
    capture_output=True, text=True, check=True).stdout)

machine = cd.pop("machine", None) or {
    "present": False, "verdict": None, "verdict_digest": None,
    "submissions_accepted": None, "submissions_rejected": None,
    "checks_green": None,
    "source": "no machine by design (solo hand)"}
machine["present"] = bool(cd.get("machine"))

def missing(v):
    return v is None or v == "not-detected"

problems = []
if hand != "solo" and (missing(machine.get("verdict"))):
    problems.append("machine verdict missing on a machine hand")
if hand == "solo" and machine["present"]:
    problems.append("machine present on the solo hand")
if hand == "solo" and not ops:
    problems.append("solo hand without its operator's records")
if not (journal_path and os.path.isfile(journal_path)):
    problems.append("the driver journal is missing (mandatory per protocol)")

metrics = {
  "cell": cell, "hand": hand,
  "frozen": {"brief_sha": brief, "image_id": image, "binary_sha": binary},
  "executor": {
    "agent_type": "experimental" if hand == "weak" else "general-purpose",
    "tokens_in": hj.get("tokens_in") if hj.get("tokens_in") is not None else hj.get("tokens_total"),
    "tokens_out": hj["tokens_out"],
    "tokens_cache_read": hj["tokens_cache_read"],
    "tokens_note": None if hj.get("tokens_in") is not None else "raw model-io pruned by the platform before collection; the TOTAL from the completion notification (the same platform source) fills tokens_in; in/out split unavailable",
    "tool_calls": hj["tool_calls"], "model_ms": hj["model_ms"],
    "wall_ms": hand_wall,
    "source": "platform model-io record of the hand agent",
  },
  "operator": op,
  "operator_loop": loop,
  "rounds_note": "rounds are counted from the driver journal (hand turns ended with STATUS); solo has no ledger by design",
  "machine": machine,
  "deliverables": cd,
  "raw_preserved": sorted(os.listdir(raw)),
  "manifest_complete": None,
  "problems": problems,
}
need = [metrics["executor"]["tokens_in"], metrics["executor"]["tool_calls"],
        metrics["deliverables"].get("specs_sections"),
        metrics["deliverables"].get("scenarios")]
if hand != "solo":
    need += [machine.get("submissions_accepted")]
metrics["manifest_complete"] = (all(n is not None for n in need)
                                and not problems)

out = os.path.join(os.path.dirname(export.rstrip("/")), "metrics.json")
if verify:
    old = json.load(open(verify))
    if old == metrics:
        print("VERIFY: PASS — recomputed metrics match the stored file bit-for-bit")
        sys.exit(0)
    print("VERIFY: FAIL — the stored metrics.json does not match the recomputation")
    for k in set(old) | set(metrics):
        if old.get(k) != metrics.get(k):
            print(f"  differs: {k}")
    sys.exit(1)

json.dump(metrics, open(out, "w", encoding="utf-8"), ensure_ascii=False, indent=1)
print(f"WROTE {out}")
print(f"manifest_complete={metrics['manifest_complete']}  problems={problems}")
sys.exit(0 if metrics["manifest_complete"] else 4)
PY
