#!/bin/bash
# preflight-audit.sh — the anti-recurrence gate for every agent
# launch of this campaign (the lesson of a live anomaly: a subagent
# told to spawn its engineer shelled out to an external agent CLI).
#
# Two checks:
#   prompts — grep the rendered prompt files for SPAWNING
#       INSTRUCTIONS (an agent must never be told to create/launch
#       another agent: only the run driver launches agents) and for
#       external agent CLI usage words. Zero tolerance: any hit is a
#       refusal to launch.
#   transcript — grep a finished agent's transcript (JSONL) for
#       INVOCATIONS of external agent CLIs in command position. Any
#       hit VOIDS the cell, independently of intent.
#
# Usage:
#   preflight-audit.sh prompts <file> [<file>...]
#   preflight-audit.sh transcript <transcript.jsonl-or-output-file>

set -uo pipefail

SPAWN_PHRASES=(
  'invoke your agent tool'
  'use your agent tool'
  'with your agent tool'
  'agent tool (subagent'
  'spawn a subagent'
  'spawn an agent'
  'launch a subagent'
  'launch an agent'
  'start a subagent'
  'create a subagent'
  'drive the engineer'
  'drive its executor'
  'driving its executor'
  'your subagent'
)
EXTERNAL_CLIS=(
  'opencode'
  'claude'
  'aider'
  'codex'
  'cursor-agent'
  'gh copilot'
  'gemini'
  'copilot'
)

case "${1:-}" in
prompts)
  shift
  status=0
  for f in "$@"; do
    while IFS= read -r phrase; do
      [ -z "$phrase" ] && continue
      if grep -qinF "$phrase" "$f"; then
        echo "PREFLIGHT prompts FAIL: '$phrase' found in $f"
        grep -inF "$phrase" "$f" | head -2 | sed 's/^/    /'
        status=1
      fi
    done < <(printf '%s\n' "${SPAWN_PHRASES[@]}")
    for cli in "${EXTERNAL_CLIS[@]}"; do
      # a prohibition block NAMES the CLI to forbid it: allow exactly
      # the forms "no <cli>", "<cli> are forbidden", "never invoke
      # ... <cli>"; flag any other mention — command position words
      if grep -qiE "(^|[^a-z])(never invoke any external agent CLI|no external CLIs)" "$f"; then
        continue
      fi
      if grep -qiE "\b${cli}\b" "$f"; then
        hits=$(grep -icE "\b${cli}\b" "$f")
        # count mentions NOT inside a prohibition sentence
        bad=$(grep -iE "\b${cli}\b" "$f" | grep -vicE "forbid|never|no external|prohibit|запрет" || true)
        if [ "${bad:-0}" -gt 0 ]; then
          echo "PREFLIGHT prompts FAIL: '$cli' used outside a prohibition in $f ($bad line(s))"
          status=1
        fi
      fi
    done
  done
  [ $status = 0 ] && echo "PREFLIGHT prompts: PASS — no spawning instructions, no external CLI usage"
  exit $status
  ;;
transcript)
  f="$2"
  status=0
  for cli in opencode aider codex "cursor-agent" "gh copilot" gemini-cli; do
    # command position: the CLI invoked as a command in a Bash tool call
    if grep -qE "\"command\": \"[^\"]*(^|[;|&] | exec | env )*${cli} " "$f" 2>/dev/null \
       || grep -qE "(^|[|;&]|\$\()${cli} (run|exec|--)" "$f" 2>/dev/null; then
      echo "PREFLIGHT transcript FAIL: external agent CLI '$cli' INVOKED — the cell is VOID"
      grep -nE "${cli} (run|exec|--)" "$f" | head -2 | sed 's/^/    /'
      status=1
    fi
  done
  if grep -qE 'opencode (run|serve|exec)' "$f" 2>/dev/null; then
    echo "PREFLIGHT transcript FAIL: opencode invocation found — the cell is VOID"
    status=1
  fi
  [ $status = 0 ] && echo "PREFLIGHT transcript: PASS — no external agent CLI invocations"
  exit $status
  ;;
*)
  echo "usage: preflight-audit.sh prompts <file>... | transcript <file>" >&2
  exit 2
  ;;
esac
