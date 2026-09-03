#!/bin/bash
# agnostic-gate-selftest.sh — a gate that cannot fail proves
# nothing: plant each violation class in a
# scratch tree and assert the extended gate FAILS on it with the
# right layer, and that a clean tree PASSES. Scratch lives in
# /tmp (nothing touches the repo); exit non-zero on any self-test
# failure.

set -euo pipefail

HERE="$(cd "$(dirname "$0")" && pwd)"
GATE="$HERE/agnostic-gate.sh"
WORK="$(mktemp -d /tmp/agate-selftest.XXXXXX)"
trap 'rm -rf "$WORK"' EXIT

make_tree() {  # $1 = dir: minimal machine-shape tree
  mkdir -p "$1/cmd/punchtape" "$1/internal/engine" "$1/internal/canondata/data"
  printf 'module example.test/m\n' > "$1/go.mod"
  printf 'package main\nfunc main() {}\n' > "$1/cmd/punchtape/main.go"
  printf 'package engine\n// clean\nvar X = 1\n' > "$1/internal/engine/e.go"
  printf 'gate.ok: "clean"\ngate.ok.ru: "чисто"\ngate.ok.zh: "干净"\n' \
    > "$1/internal/canondata/data/test.yaml"
}

check() {  # $1 label, $2 expected-exit, $3 expected-layer ("" = PASS msg), $4 tree dir
  local label="$1" want="$2" layer="$3" tree="$4"
  local out rc=0
  out="$("$GATE" "$tree" 2>&1)" || rc=$?
  if [ "$rc" != "$want" ]; then
    echo "SELFTEST [$label] FAIL: expected exit $want, got $rc"
    echo "$out" | head -5
    exit 1
  fi
  if [ -n "$layer" ] && ! grep -q "\[$layer\] FAIL" <<<"$out"; then
    echo "SELFTEST [$label] FAIL: expected [$layer] FAIL in output"
    echo "$out" | head -8
    exit 1
  fi
  if [ -z "$layer" ] && ! grep -q "GATE: PASS" <<<"$out"; then
    echo "SELFTEST [$label] FAIL: expected GATE: PASS"
    echo "$out" | head -8
    exit 1
  fi
  echo "SELFTEST [$label] ok"
}

# 0. clean tree passes
T="$WORK/clean"; make_tree "$T"
check "clean-tree-passes" 0 "" "$T"

# 1. sense word in a COMMENT (the old gate scanned comments too, but
#    this pins the property for the extended gate)
T="$WORK/sense"; make_tree "$T"
printf 'package engine\n// the conduit relay quirk\nvar Y = 2\n' > "$T/internal/engine/e.go"
check "sense-comment" 1 "sense" "$T"

# 2. near-miss: node invoked on a script in a comment (the exact
#    bgsuite fossil class the old STACK_RE missed)
T="$WORK/node"; make_tree "$T"
printf 'package engine\n// may drive it via node app.mjs today\nvar Z = 3\n' > "$T/internal/engine/e.go"
check "node-invocation" 1 "stack-nearmiss" "$T"

# 3. battery reference: milestone/finding ids + grading count
T="$WORK/batref"; make_tree "$T"
printf 'package engine\n// as seen in the run (15/15 rows)\nvar W = 4\n' > "$T/internal/engine/e.go"
check "battery-ref" 1 "batref" "$T"

# 3b. battery reference: a hyphenated wave name (the phase-C gap —
#     the named list had gone stale between campaign generations)
T="$WORK/waveref"; make_tree "$T"
printf 'package engine\n// observed in wave-S of the battery\nvar V = 5\n' > "$T/internal/engine/e.go"
check "wave-name-ref" 1 "batref" "$T"

# 4. non-synthetic example noun in a canondata block scalar
T="$WORK/noun"; make_tree "$T"
printf 'slot.hint: |\n  spell the entry `bin/serve`, not `serve`\n' > "$T/internal/canondata/data/test.yaml"
printf 'gate.ok: "clean"\ngate.ok.ru: "чисто"\ngate.ok.zh: "干净"\n' >> "$T/internal/canondata/data/test.yaml"
check "example-noun" 1 "nouns" "$T"

# 5. locale gap: .ru without .zh (the summary.checked.unpinned class)
T="$WORK/locale"; make_tree "$T"
printf 'gate.warn: "warn"\ngate.warn.ru: "предупреждение"\n' >> "$T/internal/canondata/data/test.yaml"
check "locale-gap" 1 "locale" "$T"

# 6. locale drift: a blacklisted wrong form
T="$WORK/drift"; make_tree "$T"
printf 'gate.kind: "waived"\ngate.kind.ru: "зазор принят"\ngate.kind.zh: "已豁免"\n' >> "$T/internal/canondata/data/test.yaml"
check "locale-drift" 1 "locale-drift" "$T"

echo "SELFTEST: PASS — the gate fails on every planted violation class and passes a clean tree"
