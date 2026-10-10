#!/bin/bash
# env-prep-check.sh — the PLATFORM CACHE check of the case-0
# campaign (the shared-cache law, README): the master env
# caches of the five stacks live in the PLATFORM cache area
# ($HOME/.experiments/cache/env-prep — METHODOLOGY §0.4 substrate,
# outside every campaign tree; warmed once at series prep
# by the platform maintenance procedure, the same 25 tasks). This
# tool only VERIFIES presence; it never warms (warming with the
# network is platform maintenance, not a campaign action — a wave
# never builds the substrate it stands on).
#
# Usage: env-prep-check.sh    (prints the table; COMPLETE/INCOMPLETE)
# Env: C0_ENV_PREP (default: the platform cache area)
set -euo pipefail
PREP="${C0_ENV_PREP:-$HOME/.experiments/cache/env-prep}"

miss=0
for d in cargo-registry cargo-image cargo-pressmark cargo-errand cargo-quire \
         _gomod-pressmark _gomod-errand _gomod-quire; do
  if [ -d "$PREP/$d" ]; then echo "  OK    $d"; else echo "  MISS  $d"; miss=1; fi
done
for d in node_modules nm-brim/node_modules nm-errand/node_modules nm-quire/node_modules; do
  if [ -d "$PREP/$d" ]; then echo "  OK    $d"; else echo "  MISS  $d"; miss=1; fi
done
for f in include/nlohmann/json.hpp include/inja/inja.hpp; do
  if [ -s "$PREP/$f" ]; then echo "  OK    $f"; else echo "  MISS  $f"; miss=1; fi
done
echo "PREP=$PREP"
if [ "$miss" = 0 ]; then
  echo "env-prep: COMPLETE"
else
  echo "env-prep: INCOMPLETE (platform maintenance: warm the platform cache area per the series-prep procedure)" >&2
  exit 1
fi
