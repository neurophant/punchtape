#!/bin/bash
# canaries.sh — the canary layer of the wave's anti-cheating proof:
# random tokens generated before the wave, canary files planted where
# a hand must never reach (the run root's control zones, the cells);
# the registry lives in the wave journal (never reachable by an
# executor); after the wave the scan mode looks for the tokens in
# everything a hand produced or said — any appearance is proven
# reading of the forbidden.
#
# Usage:
#   canaries.sh init <wave-control-dir> <run-root-zone> [<run-root-zone>...]
#       — generate the tokens, plant the files, print the registry
#         lines for the wave journal (token → planted path list).
#         THE ZONE LAW (the owner's word, AGENTS §14):
#         ~/.experiments is the ONLY directory a campaign run may
#         write, fully and without exception — the control dir AND
#         every planting target must resolve inside it; the wave's
#         baits ride the run root's own zones (measurement, tasks,
#         operators, manifests), the cells' baits are planted by
#         stage-cell.sh; the task sources live in the repo READ-ONLY
#         and a bait planted there REFUSES (it would also ride the
#         staging walks of stage-cell.sh)
#   canaries.sh scan <wave-control-dir> [--not <substr>] <dir> [<dir>...]
#       — scan produced material for the registered tokens; tokens
#         planted under a path matching --not belong to the scanned
#         hand's own workspace and do not count; prints one JSON line
#         {"leaks":[{token,path,line}]} ({} when clean)
set -euo pipefail

case "$1" in
init)
  CTRL="$2"; shift 2
  ZONE="$(readlink -f "$HOME/.experiments")"
  [ -d "$ZONE" ] || { echo "CANARIES REFUSED: no ~/.experiments zone on this host" >&2; exit 2; }
  # nothing is created anywhere before membership is proven on the
  # deepest EXISTING ancestor (a refused path must leave no trace),
  # and the canonical form is re-checked after the mkdir
  # (traversal-proof: ~/.experiments/../x resolves outside)
  zone_member() {
    local a="$1"
    while [ ! -e "$a" ]; do a="$(dirname "$a")"; done
    a="$(readlink -f "$a")"
    case "$a" in "$ZONE"|"$ZONE"/*) return 0 ;; *) return 1 ;; esac
  }
  for p in "$CTRL" "$@"; do
    zone_member "$p" || { echo "CANARIES REFUSED: $p is outside ~/.experiments (the zone law: the ONLY writable directory of a run, no exceptions — the repo and every other host path are never written)" >&2; exit 2; }
  done
  mkdir -p "$CTRL"
  CTRL="$(readlink -f "$CTRL")"
  case "$CTRL" in "$ZONE"|"$ZONE"/*) ;; *)
    echo "CANARIES REFUSED: $CTRL resolves outside ~/.experiments" >&2
    exit 2 ;;
  esac
  mkdir -p "$CTRL/canaries"
  REG="$CTRL/canaries/registry.txt"; : > "$REG"
  for target in "$@"; do
    mkdir -p "$target"
    target="$(readlink -f "$target")"
    case "$target" in "$ZONE"|"$ZONE"/*) ;; *)
      echo "CANARIES REFUSED: target $target resolves outside ~/.experiments" >&2
      exit 2 ;;
    esac
    tok="$(head -c 24 /dev/urandom | od -An -tx1 | tr -d ' \n')"
    name="canary-${tok:0:8}.txt"
    printf 'canary %s: this file proves reading where no hand may look\n' "$tok" \
      > "$target/.$name"
    echo "$tok $target/.$name" >> "$REG"
  done
  # one canary in the sealed tree's parent — the run root's
  # measurement zone, inside ~/.experiments (never the repo)
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
    SCR="${C0_SCRATCH_ROOT:-$HOME/.experiments}"
    mkdir -p "$SCR/tmp"
    REG="$SCR/tmp/canaries-filtered-$$.txt"
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
