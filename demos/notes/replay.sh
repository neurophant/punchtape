#!/usr/bin/env bash
# replay.sh — replays the whole notes session from scratch, live:
# contract → birth → red rows on the defective notes.js → fix → green
# → loud exclusions → done: true → verdict.
# Every terminal line quoted in README.md comes from this script's
# real output — nothing is staged or faked.
#
# Usage: demos/notes/replay.sh [path-to-punchtape]
#   binary defaults to punchtape from PATH; node must be on PATH.
# Needs: node. The run rewrites project/ in place (the film regenerates
# byte-identical on the same machine — compare with git afterwards).
set -euo pipefail

HERE="$(cd "$(dirname "$0")" && pwd)"
SC="$HERE/scenario"
PROJ="$HERE/project"
PT="${1:-punchtape}"

cd "$PROJ"
rm -rf .punchtape AGENTS.md notes.json notes.js

step() { printf '\n== %s ==\n' "$1"; }

step "0. empty folder — the machine prints its contract"
"$PT" | head -4 || true

step "1. the human's words, verbatim — .punchtape/demands.md (the birth)"
mkdir -p .punchtape && cp "$SC/demands.md" .punchtape/demands.md
"$PT"

step "2. the agent pins its checks — the defective notes.js stands"
cp "$SC/stage1/notes.js" notes.js
cp "$SC/inbox-first.yaml" .punchtape/inbox.yaml
"$PT"
echo "---- .punchtape/report.md (the facts):"
sed -n '1,12p' .punchtape/report.md

step "3. the agent fixes notes.js — the proofs went stale, the machine re-runs them itself"
cp "$SC/stage2/notes.js" notes.js
"$PT"

step "4. loud exclusions for the files the work does not cover"
cp "$SC/inbox-exclusions.yaml" .punchtape/inbox.yaml
"$PT"

step "5. the completion act — done: true"
printf 'done: true\n' > .punchtape/inbox.yaml
"$PT"

step "6. the film — .punchtape/state.md"
sed -n '1,16p' .punchtape/state.md

echo
echo "Replayed live in: $PROJ"
echo "Digests and signatures ride the environment hash: on another machine"
echo "the same session records the same greens under different digests."
