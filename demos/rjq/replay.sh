#!/usr/bin/env bash
# replay.sh — replays the rjq bug-fix session from scratch, live:
# contract → birth → the defective CLI (prints the error, exits zero)
# and the checks that catch it red → the fix → green forever → loud
# exclusions for the live-Redis paths → done: true → verdict. Every
# terminal line quoted in README.md comes from this script's real output.
#
# Usage: demos/rjq/replay.sh [path-to-punchtape]
#   binary defaults to punchtape from PATH; cargo (rustc 1.x) on PATH.
# The crates download from crates.io on the first build (network
# needed once). The refusal bytes are Linux's (os error 111).
set -euo pipefail

HERE="$(cd "$(dirname "$0")" && pwd)"
SC="$HERE/scenario"
PROJ="$HERE/project"
PT="${1:-punchtape}"

cd "$PROJ"
rm -rf .punchtape AGENTS.md src/bin target

step() { printf '\n== %s ==\n' "$1"; }

step "0. the standing library, no zone yet — the machine prints its contract"
"$PT" | head -4 || true

step "1. the human's words, verbatim — .punchtape/demands.md (the birth)"
mkdir -p .punchtape && cp "$SC/demands.md" .punchtape/demands.md
"$PT"

step "2. the agent declares the counted root — .punchtape/universe"
printf 'src\n' > .punchtape/universe
"$PT"

step "3. the defective CLI and the checks that will catch it"
mkdir -p src/bin && cp "$SC/stage1/src/bin/rjq-cli.rs" src/bin/
cp "$SC/inbox-first.yaml" .punchtape/inbox.yaml
"$PT"
echo "---- .punchtape/report.md (the facts):"
sed -n '1,14p' .punchtape/report.md

step "4. the fix — one line of exit code; the proofs went stale, the machine re-runs them itself"
cp "$SC/stage2/src/bin/rjq-cli.rs" src/bin/rjq-cli.rs
"$PT"

step "5. loud exclusions for the live-Redis paths"
cp "$SC/inbox-exclusions.yaml" .punchtape/inbox.yaml
"$PT"

step "6. the completion act — done: true"
printf 'done: true\n' > .punchtape/inbox.yaml
"$PT"

step "7. the film — .punchtape/state.md"
sed -n '1,18p' .punchtape/state.md

echo
echo "Replayed live in: $PROJ"
echo "Digests and signatures ride the environment hash: on another machine"
echo "the same session records the same greens under different digests."
