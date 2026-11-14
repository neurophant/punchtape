#!/usr/bin/env bash
# replay.sh — replays the gouija backfill session from scratch, live:
# contract → birth → the universe declared → the standing CLI contract
# claimed whole (build + four pinned rows, all green on the old code,
# nothing red — a backfill checks what already stands) → done: true →
# verdict. Every terminal line quoted in README.md comes from this
# script's real output.
#
# Usage: demos/gouija/replay.sh [path-to-punchtape]
#   binary defaults to punchtape from PATH; go (1.24+) must be on PATH.
# The dependencies download from their modules on the first build
# (network needed once).
set -euo pipefail

HERE="$(cd "$(dirname "$0")" && pwd)"
SC="$HERE/scenario"
PROJ="$HERE/project"
PT="${1:-punchtape}"

cd "$PROJ"
rm -rf .punchtape AGENTS.md gouija-demo

step() { printf '\n== %s ==\n' "$1"; }

step "0. the standing repo, no zone yet — the machine prints its contract"
"$PT" | head -4 || true

step "1. the human's words, verbatim — .punchtape/demands.md (the birth)"
mkdir -p .punchtape && cp "$SC/demands.md" .punchtape/demands.md
"$PT"

step "2. the agent declares the counted root — .punchtape/universe"
printf 'cmd/gouija\n' > .punchtape/universe
"$PT"

step "3. the first document: the CLI contract claimed, the build declared"
cp "$SC/inbox-first.yaml" .punchtape/inbox.yaml
"$PT"
echo "---- .punchtape/report.md (the facts):"
sed -n '1,10p' .punchtape/report.md

step "4. loud exclusions for the start-up paths no byte-exact pin can hold"
cp "$SC/inbox-exclusions.yaml" .punchtape/inbox.yaml
"$PT"
echo "---- .punchtape/report.md:"
sed -n '1,8p' .punchtape/report.md

step "5. the completion act — done: true"
printf 'done: true\n' > .punchtape/inbox.yaml
"$PT"

step "6. the film — .punchtape/state.md"
sed -n '1,16p' .punchtape/state.md

echo
echo "Replayed live in: $PROJ"
echo "Digests and signatures ride the environment hash: on another machine"
echo "the same session records the same greens under different digests."
