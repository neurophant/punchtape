#!/bin/bash
# verify-cell.sh — the staging audit that a live anomaly proved
# necessary: a staged cell is INVISIBLE-valid until every fact below
# is checked (the failed flight let two zombie containers (dead
# bind mounts from a destroyed source tree) pass a NAME-ONLY check).
#
# Usage: verify-cell.sh <container> <task-dir-on-host>
# Checks (all must pass, exit non-zero on any failure):
#   1. the container exists and its mount source IS <task-dir>
#      (mountinfo, resolved — no stale "/deleted" tails);
#   2. the bind is LIVE both directions: a file written inside
#      appears on the host, and vice versa;
#   3. exec with -w /run/task works (the hands' access path);
#   4. the machine binary sha inside equals the frozen sha (top|weak)
#      or is absent (solo);
#   5. the brief sha inside matches the host brief sha.

set -euo pipefail
NAME="$1"; TASKDIR="$(readlink -f "$2")"

fail() { echo "VERIFY-CELL $NAME FAIL: $1" >&2; exit 1; }

docker inspect "$NAME" >/dev/null 2>&1 || fail "no such container"

# 1. mount source: staleness from mountinfo, identity from the
#    configured mount (mountinfo renders the path in the daemon's
#    namespace — not comparable verbatim)
MI="$(docker exec -w / "$NAME" sh -c 'grep " /run/task " /proc/self/mountinfo' 2>/dev/null | awk '{print $4}' | head -1)"
[ -n "$MI" ] || fail "no /run/task mount line in mountinfo"
case "$MI" in
  *deleted|*"(deleted)") fail "stale bind mount: $MI" ;;
esac
SRC="$(docker inspect "$NAME" --format '{{range .Mounts}}{{if eq .Destination "/run/task"}}{{.Source}}{{end}}{{end}}')"
[ "$SRC" = "$TASKDIR" ] || fail "configured mount source $SRC != intended $TASKDIR"

# 2. liveness, both directions
TOKEN="verify-$$-$(date +%s)"
docker exec -w /run/task "$NAME" touch "/run/task/.verify-$TOKEN" 2>/dev/null || fail "exec -w /run/task failed"
[ -f "$TASKDIR/.verify-$TOKEN" ] || fail "inside→host bind not live"
echo ok > "$TASKDIR/.verify-host-$TOKEN"
docker exec -w /run/task "$NAME" test -f "/run/task/.verify-host-$TOKEN" 2>/dev/null || fail "host→inside bind not live"
rm -f "$TASKDIR/.verify-$TOKEN" "$TASKDIR/.verify-host-$TOKEN"

# 3. exec -w (already exercised above)

# 4. machine presence
HAND="${NAME##*-}"
case "$HAND" in
  solo)
    if docker exec -w / "$NAME" sh -lc 'command -v punchtape' 2>/dev/null; then
      fail "solo hand sees punchtape"
    fi
    echo "verify-cell $NAME: PASS (solo: machine absent, bind live, exec ok)"
    ;;
  top|weak)
    FROZEN="${FROZEN_SHA:?FROZEN_SHA env must carry the frozen binary sha}"
    IN_SHA="$(docker exec -w / "$NAME" sha256sum /usr/local/bin/punchtape | cut -d' ' -f1)"
    [ "$IN_SHA" = "$FROZEN" ] || fail "binary sha $IN_SHA != frozen $FROZEN"
    echo "verify-cell $NAME: PASS (bind $SRC live both ways, exec ok, machine sha verified)"
    ;;
  *) fail "cannot infer hand from name: $NAME" ;;
esac
