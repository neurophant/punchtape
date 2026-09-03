#!/bin/bash
# agent-usage.sh — one agent's OWN economy, read mechanically from
# the harness's per-agent model I/O records (th3 METRICS source 1).
# Every output field belongs to exactly one agent; the campaign's
# hard rule: the hand's metrics and the operator's metrics are never
# summed or mixed — call this script per agent and keep the results
# separate.
#
# Usage: agent-usage.sh <model-io-sess_*.jsonl> [<more>...]
# Output: one JSON object per file (no cross-file totals by design —
# one file IS one agent session).

set -euo pipefail

for f in "$@"; do
  [ -f "$f" ] || { echo "no such file: $f" >&2; exit 2; }
  python3 - "$f" <<'PY'
import json, sys, os
f = sys.argv[1]
calls = 0; tin = tout = tcache = 0; model_ms = 0; tools = 0; model = ""
first = None; last = None
for line in open(f, encoding="utf-8"):
    try:
        r = json.loads(line)
    except json.JSONDecodeError:
        continue
    calls += 1
    u = (r.get("response") or {}).get("usage") or {}
    tin += u.get("inputTokens", 0) or 0
    tout += u.get("outputTokens", 0) or 0
    tcache += u.get("cacheReadTokens", 0) or 0
    model_ms += r.get("durationMs", 0) or 0
    tools += len((r.get("response") or {}).get("toolCalls") or [])
    m = (r.get("model") or {}).get("modelId")
    if m: model = m
    ts = r.get("startedAt") or r.get("completedAt")
    if ts:
        if first is None or ts < first: first = ts
        if last is None or ts > last: last = ts
print(json.dumps({
    "file": os.path.basename(f),
    "model": model,
    "model_calls": calls,
    "tokens_in": tin,
    "tokens_out": tout,
    "tokens_cache_read": tcache,
    "tool_calls": tools,
    "model_ms": model_ms,
    "first_call": first,
    "last_call": last,
}, ensure_ascii=False))
PY
done
