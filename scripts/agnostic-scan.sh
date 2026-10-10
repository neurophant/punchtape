#!/bin/bash
# THE AGNOSTIC SCAN — the eternal mechanical gate of the
# machine's total agnosticism. The law (the owner's word): there is
# NOTHING unagnostic in THIS repo's machine — not in the code, not
# in the canon schema, not in the reference docs written to an
# instance. The machine knows only workflow mechanics; the
# subject-matter virtues live in the campaigns' probes, delivered
# as staged bytes through the door (witness: file://NAME).
#
# The scan is deterministic: the grep list below is the EXHAUSTIVE
# inventory of forbidden identifiers/fields/yaml-keys/subject-matter
# witness words. Changing the list = a visible commit. The scanner
# fails (rc 1) on any appearance — make test goes red.
#
# The list: pure, no-network, writes-nothing-outside,
# NoNetwork, WritesNothingOutside, headPure — and their direct
# derivative spellings (PureRow, typed-pure-form), which fall under
# the same law. Machine-side checker implementations (executable
# checks of subject properties beyond a row's rc/out comparison)
# are forbidden by this scan together with the vocabulary: probes
# arrive only as staged bytes through the door. The procedural
# words (nonet, unchanged, launch, tree) are allowed only as
# launch-grammar mode names.
set -euo pipefail
cd "$(dirname "$0")/.."

fail=0

# (1) The subject-matter witness vocabulary: word-boundary,
#     case-insensitive, over all the machine's sources and canon
#     data (code, texts, templates, reference docs, tests).
hits="$(grep -rniE --include='*.go' --include='*.yaml' --include='*.md' --include='*.txt' --include='*.sh' \
  -e '[^a-z]pure[^a-z]' -e '^pure[^a-z]' -e '[^a-z]pure$' \
  -e 'no-network' -e 'writes-nothing-outside' \
  cmd internal 2>/dev/null || true)"
if [ -n "$hits" ]; then
  echo "FAIL: agnostic scan — the subject-matter witness vocabulary is in the machine:"
  echo "$hits"
  fail=1
fi

# (2) The identifier spellings of the same list (CamelCase
#     fields/types, validator yaml keys).
hits="$(grep -rnE --include='*.go' --include='*.yaml' --include='*.md' \
  -e 'NoNetwork' -e 'WritesNothingOutside' -e 'headPure' -e 'PureRow' -e 'typed-pure-form' \
  cmd internal 2>/dev/null || true)"
if [ -n "$hits" ]; then
  echo "FAIL: agnostic scan — witness identifier spellings are in the machine:"
  echo "$hits"
  fail=1
fi

if [ "$fail" -eq 0 ]; then
  echo "agnostic scan: CLEAN — the machine knows only mechanics; probes arrive as staged bytes through the door"
fi
exit "$fail"
