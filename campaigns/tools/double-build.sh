#!/bin/bash
# double-build.sh — the freeze ritual: build
# twice from the same tree, cmp, sha256. A clean double build with
# one sha is the reproducibility witness before any binary freeze
# (per batch, and the th3 run's binary freeze). -buildvcs=false
# keeps the binary identical regardless of git state.

set -euo pipefail

REPO="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$REPO"

make build >/dev/null
cp punchtape /tmp/punchtape.build1
make build >/dev/null
cp punchtape /tmp/punchtape.build2

if ! cmp -s /tmp/punchtape.build1 /tmp/punchtape.build2; then
  echo "DOUBLE BUILD: FAIL — the two builds differ"
  exit 1
fi

sha="$(sha256sum punchtape | awk '{print $1}')"
echo "DOUBLE BUILD: PASS — one sha256 across two clean builds"
echo "sha256: $sha"
echo "binary: $(pwd)/punchtape ($(wc -c < punchtape) bytes)"
rm -f /tmp/punchtape.build1 /tmp/punchtape.build2
