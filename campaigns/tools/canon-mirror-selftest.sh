#!/bin/bash
# canon-mirror-selftest.sh — a check that cannot fail proves nothing
# (the gate self-test's rule, applied to the mirror): plant each
# violation class in a scratch copy of the real tree and assert the
# mirror FAILS on it, and that the untouched copy PASSES. Scratch
# lives in /tmp (nothing touches the repo); exit non-zero on any
# self-test failure.
#
# Classes planted:
#   1. a literal text key absent from the registries (the panic class
#      the release candidates carried in the lint rejection path)
#   2. a composed-key prefix matching no registry key
#   3. an absent limit key
#
# Usage: canon-mirror-selftest.sh

set -euo pipefail

HERE="$(cd "$(dirname "$0")" && pwd)"
MIRROR="$HERE/canon-mirror.sh"
WORK="$(mktemp -d /tmp/canon-mirror-selftest.XXXXXX)"
trap 'rm -rf "$WORK"' EXIT

# A minimal machine-shape tree (the mirror scans ROOT/internal and
# ROOT/cmd plus the canon registries).
make_tree() {  # $1 = dir
  mkdir -p "$1/internal/engine" "$1/internal/canondata/data" "$1/cmd/punchtape"
  printf 'module example.test/m\n' > "$1/go.mod"
  printf 'package engine\nvar X = 1\n' > "$1/internal/engine/e.go"
  printf 'package main\nfunc main() {}\n' > "$1/cmd/punchtape/main.go"
  printf 'known.key: "text"\nknown.key.ru: "текст"\nknown.key.zh: "文本"\n' \
    > "$1/internal/canondata/data/texts.yaml"
  printf 'known.limit: 3\n' > "$1/internal/canondata/data/limits.yaml"
  printf '# empty registry\n' > "$1/internal/canondata/data/lint.yaml"
  printf '# empty registry\n' > "$1/internal/canondata/data/envdiag.yaml"
}

check() {  # $1 label, $2 expected-exit, $3 tree dir
  local label="$1" want="$2" tree="$3"
  local out rc=0
  out="$("$MIRROR" "$tree" 2>&1)" || rc=$?
  if [ "$rc" != "$want" ]; then
    echo "SELFTEST [$label] FAIL: expected exit $want, got $rc"
    echo "$out" | head -5
    exit 1
  fi
  echo "SELFTEST [$label] ok (exit $rc)"
}

# 0. clean tree passes
T="$WORK/clean"; make_tree "$T"
check "clean-tree-passes" 0 "$T"

# 1. a literal text key absent from the registries — the carried class
T="$WORK/missing-key"; make_tree "$T"
printf 'package engine\nimport "example.test/m/internal/canondata"\nvar S = canondata.T("known.nope")\n' \
  > "$T/internal/engine/e.go"
check "missing-text-key" 1 "$T"

# 2. a composed-key prefix matching no registry key
T="$WORK/dead-prefix"; make_tree "$T"
printf 'package engine\nimport "example.test/m/internal/canondata"\nvar S = canondata.TFor("en", "known.dead."+x)\n' \
  > "$T/internal/engine/e.go"
check "dead-composed-prefix" 1 "$T"

# 3. an absent limit key
T="$WORK/missing-limit"; make_tree "$T"
printf 'package engine\nimport "example.test/m/internal/canondata"\nvar N = canondata.Limit("known.gone")\n' \
  > "$T/internal/engine/e.go"
check "missing-limit-key" 1 "$T"

echo "SELFTEST: PASS — the mirror fails on every planted class and passes a clean tree"
