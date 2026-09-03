#!/bin/bash
# stage-cell.sh — stage ONE th3 cell: identical inputs and conditions
# for every hand; the ONLY hand-conditional line in this script is the
# machine binary mount (machine hands) and its verification. Proof of
# parity = reading this file: everything else is unconditional.
#
# Usage: stage-cell.sh <sense> <stack> <hand> <task-dir>
#   hand ∈ {solo, top, weak}
# Env:   TH3_IMAGE (default punchtape-world:th3)
#        TH3_BINARY (required for hand top|weak: the frozen binary path)
#
# Cell name: pt-th3-<sense>-<stack>-<hand>. Evidence: the task dir is
# the record (export before removing the container).

set -euo pipefail

SENSE="$1"; STACK="$2"; HAND="$3"; TASKDIR="$4"
HERE="$(cd "$(dirname "$0")" && pwd)"
CAMP="$(cd "$HERE/.." && pwd)"
IMAGE="${TH3_IMAGE:-punchtape-world:th3}"
NAME="pt-th3-${SENSE}-${STACK}-${HAND}"

# --- inputs: the same brief, the same fixtures, byte-identical for
# --- every hand (the campaign's own copies; shas recorded below).
mkdir -p "$TASKDIR"
cp "$CAMP/briefs/$SENSE/$STACK.md" "$TASKDIR/brief.md"
if [ -d "$CAMP/materials" ] && ls "$CAMP/materials"/*.jpg >/dev/null 2>&1 && [ "$SENSE" = brim ]; then
  mkdir -p "$TASKDIR/materials"
  cp "$CAMP/materials"/*.jpg "$TASKDIR/materials/"
  chmod -w "$TASKDIR/materials"/*.jpg
fi

# --- the container: network off, one cpu, 1g, 512 pids, task dir rw.
# --- THE ONLY HAND BRANCH IN THIS SCRIPT: machine hands mount the
# --- binary read-only; solo mounts nothing extra.
BIN_MOUNT=()
case "$HAND" in
  solo) ;;
  top|weak)
    [ -n "${TH3_BINARY:-}" ] || { echo "TH3_BINARY is required for hand $HAND" >&2; exit 2; }
    BIN_MOUNT=(-v "$(readlink -f "$TH3_BINARY"):/usr/local/bin/punchtape:ro")
    ;;
  *) echo "hand must be solo|top|weak" >&2; exit 2 ;;
esac

# --- the container: network off, one cpu, 1g, 512 pids, task dir rw.
# --init: an init process reaps orphans — killed runner processes
# must not accumulate as defunct and exhaust pids.max (the
# conduit-go-weak incident: 503 zombies).
docker run -d --name "$NAME" --network none --init \
  --cpus 1 --memory 1g --pids-limit 512 \
  -v "$(readlink -f "$TASKDIR"):/run/task" \
  "${BIN_MOUNT[@]}" \
  -w /run/task "$IMAGE" sleep infinity

# --- parity verification, printed for the cell journal:
echo "== th3 cell $NAME =="
echo "brief:  $(sha256sum "$TASKDIR/brief.md" | cut -d' ' -f1)  ($SENSE/$STACK.md)"
for f in "$TASKDIR"/materials/*; do
  [ -f "$f" ] && echo "material: $(basename "$f") $(sha256sum "$f" | cut -d' ' -f1)"
done
echo "image:  $IMAGE $(docker image inspect "$IMAGE" --format '{{.Id}}' | cut -c8-19)"
case "$HAND" in
  solo)
    # solo must NOT have the machine: command -v fails inside
    if docker exec "$NAME" sh -lc 'command -v punchtape' 2>/dev/null; then
      echo "PARITY FAIL: the solo hand sees punchtape" >&2; exit 1
    fi
    echo "machine: ABSENT (verified: command -v punchtape fails)"
    ;;
  top|weak)
    # machine hands must see the exact frozen binary
    IN_SHA="$(docker exec "$NAME" sha256sum /usr/local/bin/punchtape | cut -d' ' -f1)"
    LOC_SHA="$(sha256sum "$TH3_BINARY" | cut -d' ' -f1)"
    [ "$IN_SHA" = "$LOC_SHA" ] || { echo "PARITY FAIL: binary sha mismatch" >&2; exit 1; }
    echo "machine: $IN_SHA (mounted ro, sha verified)"
    ;;
esac
echo "limits: none-net / 1 cpu / 1g / 512 pids — identical for every hand"
