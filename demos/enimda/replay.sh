#!/usr/bin/env bash
# replay.sh — replays the enimda feature-increment session from scratch,
# live: contract → birth → the standing library claimed whole → verdict
# #1 → the human's new words arrive → the CLI feature lands (only the
# new lines claim; the library's rows never re-run) → verdict #2. Every
# terminal line quoted in README.md comes from this script's real output.
#
# Usage: demos/enimda/replay.sh [path-to-punchtape]
#   binary defaults to punchtape from PATH; python3 (3.x) on PATH.
# The replay builds its own pinned environment (.venv: numpy 2.3.4,
# Pillow 11.3.0, imgpy 1.1.0 — network once) and puts it first on
# PATH, so the machine's clean rows see exactly these packages.
set -euo pipefail

HERE="$(cd "$(dirname "$0")" && pwd)"
SC="$HERE/scenario"
PROJ="$HERE/project"
PT="${1:-punchtape}"

VENV="$HERE/.venv"
if [ ! -x "$VENV/bin/python3" ]; then
    python3 -m venv "$VENV"
    "$VENV/bin/pip" --quiet install numpy==2.3.4 Pillow==11.3.0 imgpy==1.1.0
fi
export PATH="$VENV/bin:$PATH"

cd "$PROJ"
rm -rf .punchtape AGENTS.md enimda/cli.py enimda/__pycache__ tests/__pycache__ __pycache__

step() { printf '\n== %s ==\n' "$1"; }

step "0. the standing library, no zone yet — the machine prints its contract"
"$PT" | head -4 || true

step "1. the human's words, verbatim — .punchtape/demands.md (the birth)"
mkdir -p .punchtape && cp "$SC/demands-round1.md" .punchtape/demands.md
"$PT"

step "2. the agent declares the counted root — .punchtape/universe"
printf 'enimda\n' > .punchtape/universe
"$PT"

step "3. the first document: the standing library claimed whole"
cp "$SC/inbox-first.yaml" .punchtape/inbox.yaml
"$PT"
echo "---- .punchtape/report.md (the facts):"
sed -n '1,8p' .punchtape/report.md

step "4. the completion act — done: true (verdict #1)"
printf 'done: true\n' > .punchtape/inbox.yaml
"$PT"

step "5. the human's new words arrive — appended to demands.md, they wait"
cat "$SC/demands-round2-line.txt" >> .punchtape/demands.md
"$PT"

step "6. the agent writes the feature — enimda/cli.py, new lines only"
cp "$SC/feature/enimda/cli.py" enimda/cli.py

step "7. the second document: the CLI claimed, the library untouched"
cp "$SC/inbox-feature.yaml" .punchtape/inbox.yaml
"$PT"
echo "---- .punchtape/report.md (the facts):"
sed -n '1,8p' .punchtape/report.md

step "8. the completion act — done: true (verdict #2)"
printf 'done: true\n' > .punchtape/inbox.yaml
"$PT"

step "9. the film — .punchtape/state.md"
sed -n '1,20p' .punchtape/state.md

echo
echo "Replayed live in: $PROJ"
echo "Digests and signatures ride the environment hash: on another machine"
echo "the same session records the same greens under different digests."
