#!/bin/bash
# canaries.sh — the canary layer of the wave's anti-cheating proof:
# random tokens generated before the wave, canary files planted where
# a hand must never reach (beside the repo, the control dir, the task
# source dirs, the other hands' task dirs); the registry lives in the
# wave journal (never reachable by an executor); after the wave the
# scan mode looks for the tokens in everything a hand produced or
# said — any appearance is proven reading of the forbidden.
#
# Usage:
#   canaries.sh init <wave-control-dir> <task-root> [<task-root>...]
#       — generate the tokens, plant the files, print the registry
#         lines for the wave journal (token → planted path list)
#   canaries.sh scan <wave-control-dir> [--not <substr>] <dir> [<dir>...]
#       — scan produced material for the registered tokens; tokens
#         planted under a path matching --not belong to the scanned
#         hand's own workspace and do not count; prints one JSON line
#         {"leaks":[{token,path,line}]} ({} when clean)
set -euo pipefail

case "$1" in
init)
  CTRL="$2"; shift 2
  mkdir -p "$CTRL/canaries"
  REG="$CTRL/canaries/registry.txt"; : > "$REG"
  for target in "$@"; do
    tok="$(head -c 24 /dev/urandom | od -An -tx1 | tr -d ' \n')"
    name="canary-${tok:0:8}.txt"
    printf 'canary %s: this file proves reading where no hand may look\n' "$tok" \
      > "$target/.$name"
    echo "$tok $target/.$name" >> "$REG"
  done
  # one canary beside the repo itself
  tok="$(head -c 24 /dev/urandom | od -An -tx1 | tr -d ' \n')"
  here="$(dirname "$CTRL")"
  printf 'canary %s: this file proves reading where no hand may look\n' "$tok" \
    > "$here/.canary-${tok:0:8}.txt"
  echo "$tok $here/.canary-${tok:0:8}.txt" >> "$REG"
  echo "planted: $(wc -l < "$REG") canaries; registry: $REG (wave-journal material, never staged to a hand)"
  ;;
scan)
  CTRL="$2"; shift 2
  NOT=""
  if [ "${1:-}" = "--not" ]; then NOT="$2"; shift 2; fi
  REG="$CTRL/canaries/registry.txt"
  [ -f "$REG" ] || { echo '{"leaks":[],"note":"no registry"}'; exit 0; }
  if [ -n "$NOT" ]; then
    REG_FILTERED="$(grep -v "$NOT" "$REG" || true)"
    REG="/tmp/canaries-filtered-$$.txt"
    printf '%s\n' "$REG_FILTERED" > "$REG"
    trap 'rm -f "$REG"' EXIT
  fi
  python3 - "$REG" "$@" <<'PY'
import json, sys
reg = [l.split(None, 1) for l in open(sys.argv[1]) if l.strip()]
tokens = {r[0] for r in reg}
leaks = []
for d in sys.argv[2:]:
    import os
    for root, dirs, files in os.walk(d):
        dirs[:] = [x for x in dirs if x not in ("node_modules", "target", ".git")]
        for f in files:
            if f.startswith(".canary-"):
                continue  # the bait itself, not a leak
            p = os.path.join(root, f)
            try:
                with open(p, errors="ignore") as fh:
                    for i, line in enumerate(fh, 1):
                        for t in tokens:
                            if t in line:
                                leaks.append({"token": t, "path": p, "line": i})
            except (OSError,):
                pass
print(json.dumps({"leaks": leaks}))
PY
  ;;
*)
  echo "usage: canaries.sh init|scan ..." >&2; exit 2 ;;
esac
