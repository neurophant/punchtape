#!/bin/bash
# wave-metrics.sh — the per-hand captured-files writer (the wave's
# metrics discipline): meta.json, platform.json (executor ACTIVE
# moves only; operator waits live in decision_wait_s), operator.json
# (the SEPARATE operator accounting — never summed with the hand),
# summary.json (from the exported machine telemetry / the solo
# acceptance), evidence.json (identical-conditions and anti-cheat
# fields), findings.json. Sources: the harness rollout records (one
# JSON per agent), the exported task dir, docker inspect, the staged
# sums; nothing self-reported is a number.
#
# Usage:
#   wave-metrics.sh <cell-evidence-dir> <sense> <stack> <hand> \
#       --container <name> --model <model> \
#       --brief-sha <sha> --image-digest <id> \
#       [--binary-sha <sha|absent>] \
#       --hand-io <rollout.jsonl> [--hand-io <more>...] \
#       [--op-io <rollout.jsonl> --op-io <more>...] \
#       [--decision-wait-s <seconds>] [--findings <findings.json>]
set -euo pipefail

CELL="$1"; SENSE="$2"; STACK="$3"; HAND="$4"; shift 4
CONTAINER=""; MODEL=""; BRIEF=""; IMAGE=""; BIN="absent"; DW=""
declare -a HAND_IO=() OP_IO=(); FINDINGS="[]"
NOTIF_USAGE=""
while [ $# -gt 0 ]; do
  case "$1" in
    --container) CONTAINER="$2"; shift 2 ;;
    --model) MODEL="$2"; shift 2 ;;
    --brief-sha) BRIEF="$2"; shift 2 ;;
    --image-digest) IMAGE="$2"; shift 2 ;;
    --binary-sha) BIN="$2"; shift 2 ;;
    --hand-io) HAND_IO+=("$2"); shift 2 ;;
    --op-io) OP_IO+=("$2"); shift 2 ;;
    --decision-wait-s) DW="$2"; shift 2 ;;
    --findings) FINDINGS="$(cat "$2")"; shift 2 ;;
    --notification-usage) NOTIF_USAGE="$2"; shift 2 ;;
    *) echo "unknown: $1" >&2; exit 2 ;;
  esac
done

usage_json() { # rollouts... -> one usage JSON (sums across THIS role's own files)
  python3 - "$@" <<'PY'
import json, sys
calls = tin = tout = tcache = ms = tools = 0; model = ""
for path in sys.argv[1:]:
    for line in open(path, encoding="utf-8"):
        try: r = json.loads(line)
        except json.JSONDecodeError: continue
        calls += 1
        u = (r.get("response") or {}).get("usage") or {}
        tin += u.get("inputTokens", 0) or 0
        tout += u.get("outputTokens", 0) or 0
        tcache += u.get("cacheReadTokens", 0) or 0
        ms += r.get("durationMs", 0) or 0
        tools += len((r.get("response") or {}).get("toolCalls") or [])
        m = (r.get("model") or {}).get("modelId")
        if m: model = m
print(json.dumps({"calls": calls, "tokens_in": tin, "tokens_out": tout,
                  "tokens_cache_read": tcache, "model_ms": ms,
                  "tool_calls": tools, "model": model}))
PY
}

if [ ${#HAND_IO[@]} -gt 0 ]; then
  HAND_USAGE=$(usage_json "${HAND_IO[@]}")
else
  # the platform prunes rollouts of finished agents faster than any
  # driver can snapshot; the completion notification's usage block is
  # the same platform source as totals (no in/out split — noted)
  if [ -n "$NOTIF_USAGE" ]; then
    HAND_USAGE=$(python3 -c '
import json,sys
u=json.loads(sys.argv[1])
print(json.dumps({"calls":0,"tokens_in":u.get("subagent_tokens",0),
"tokens_out":0,"tokens_cache_read":0,
"model_ms":u.get("duration_ms",0),
"tool_calls":u.get("tool_uses",0),"model":""}))' "$NOTIF_USAGE")
    echo "NOTE: rollout pruned; platform notification usage recorded as totals (no in/out split)" >&2
  else
    HAND_USAGE='{"calls":0,"tokens_in":0,"tokens_out":0,"tokens_cache_read":0,"model_ms":0,"tool_calls":0,"model":""}'
  fi
fi
if [ ${#OP_IO[@]} -gt 0 ]; then OP_USAGE=$(usage_json "${OP_IO[@]}"); else OP_USAGE='{"calls":0,"tokens_in":0,"tokens_out":0,"tokens_cache_read":0,"model_ms":0,"tool_calls":0,"model":""}'; fi

docker inspect "$CONTAINER" > "$CELL/docker-inspect.json" 2>/dev/null || true

python3 - "$CELL" "$SENSE" "$STACK" "$HAND" "$CONTAINER" "$MODEL" "$BRIEF" "$IMAGE" "$BIN" "$DW" \
  "$HAND_USAGE" "$OP_USAGE" "$FINDINGS" <<'PY'
import json, os, re, sys
(cell, sense, stack, hand, container, model, brief, image, binsha, dw,
 hand_u, op_u, findings) = sys.argv[1:14]
hand_u, op_u = json.loads(hand_u), json.loads(op_u)
os.makedirs(cell, exist_ok=True)

meta = {"sense": sense, "stack": stack, "hand": hand, "model": model,
        "container": container}
json.dump(meta, open(os.path.join(cell, "meta.json"), "w"), indent=1)

platform = {
    "elapsed_s": round(hand_u.get("model_ms", 0) / 1000.0, 1),
    "tokens_in": hand_u.get("tokens_in", 0),
    "tokens_out": hand_u.get("tokens_out", 0),
    "tokens_cache_read": hand_u.get("tokens_cache_read", 0),
    "tool_calls": hand_u.get("tool_calls", 0),
    "decision_wait_s": float(dw) if dw not in ("", None) else None,
}
json.dump(platform, open(os.path.join(cell, "platform.json"), "w"), indent=1)

operator = {"model": op_u.get("model") or None,
            "elapsed_s": round(op_u.get("model_ms", 0) / 1000.0, 1),
            "tokens_in": op_u.get("tokens_in", 0),
            "tokens_out": op_u.get("tokens_out", 0),
            "decisions": op_u.get("calls", 0)}
json.dump(operator, open(os.path.join(cell, "operator.json"), "w"), indent=1)

# summary from the exported machine telemetry (machine hands) or the
# exported task dir (solo)
summary = {"ruling": None, "cases_pass": None, "cases_total": None,
           "tests_pass": None, "tests_total": None,
           "describes": None, "capabilities": None,
           "families_covered": None, "families_total": None,
           "cases_linked": None, "cases_runnable": None,
           "code_coverage_pct": None, "loc": None}
taskdir = os.path.join(cell, "task")
led = os.path.join(taskdir, ".punchtape", "ledger.yamll")
if os.path.isfile(led):
    raw = open(led).read()
    verdicts = re.findall(r"^(VERDICT: .*)$", raw, re.M)
    summary["ruling"] = verdicts[-1] if verdicts else None
scen_dir = os.path.join(taskdir, ".punchtape", "canon", "scenarios")
req_dir = os.path.join(taskdir, ".punchtape", "canon", "requirements")
chk_dir = os.path.join(taskdir, ".punchtape", "canon", "checks")
if os.path.isdir(scen_dir):
    summary["capabilities"] = len(os.listdir(req_dir))
    summary["cases_total"] = len(os.listdir(scen_dir))
if os.path.isdir(chk_dir):
    summary["tests_total"] = len(os.listdir(chk_dir))
specs_dir = os.path.join(taskdir, ".punchtape", "specs")
if os.path.isdir(specs_dir):
    summary["describes"] = len([f for f in os.listdir(specs_dir)
                                if f.endswith(".md") and f != "index.md"])
# prose coverage of capabilities
if summary["describes"] is not None and summary["capabilities"]:
    summary["describes"] = min(summary["describes"], summary["capabilities"])
# loc over the product (no machine dirs). Two admission rules, both
# universal: a known source extension, OR a shebang first line (an
# extensionless script — `rpn`, `note`, an executable of any stack —
# declares itself; the counter learns nothing about stacks).
exts = (".py", ".go", ".rs", ".c", ".cc", ".cpp", ".h", ".hpp", ".js",
        ".mjs", ".ts", ".sh")
n = 0
for root, dirs, files in os.walk(taskdir):
    dirs[:] = [d for d in dirs if d not in (".punchtape", "node_modules",
                                            "target", "materials", "scratch")]
    for f in files:
        path = os.path.join(root, f)
        counted = f.endswith(exts)
        if not counted:
            try:
                with open(path, "rb") as fh:
                    counted = fh.read(2) == b"#!"
            except OSError:
                counted = False
        if counted:
            try:
                n += sum(1 for _ in open(path, errors="ignore"))
            except OSError:
                pass
summary["loc"] = n
summary["loc_method"] = "extensions+shebang, machine dirs excluded"
json.dump(summary, open(os.path.join(cell, "summary.json"), "w"), indent=1)

# evidence: identical conditions + anti-cheat fields (scans fill later)
insp = {}
try:
    insp = json.load(open(os.path.join(cell, "docker-inspect.json")))[0]
except Exception:
    pass
hostcfg = insp.get("HostConfig", {}) if insp else {}
mounts = [{"source": m.get("Source"), "dest": m.get("Destination"),
           "mode": (m.get("Mode") or "")} for m in insp.get("Mounts", [])]
evidence = {
    "brief_sha256": brief,
    "materials": {},
    "materials_after": {},
    "world_image_digest": image,
    "limits": {"cpus": hostcfg.get("NanoCpus"), "memory": hostcfg.get("Memory"),
               "pids": (hostcfg.get("PidsLimit"))},
    "network_mode": hostcfg.get("NetworkMode"),
    "mounts": mounts,
    "punchtape_binary_sha256": None if binsha == "absent" else binsha,
    "product_scan": [],
    "canary_leaks": [],
    "transcript_scan": [],
}
json.dump(evidence, open(os.path.join(cell, "evidence.json"), "w"), indent=1)

json.dump(findings if isinstance(findings, list) else json.loads(findings),
          open(os.path.join(cell, "findings.json"), "w"), indent=1)
print("captured files written to", cell)
PY
