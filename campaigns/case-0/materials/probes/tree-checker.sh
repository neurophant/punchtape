#!/bin/sh
# the campaign's tree probe: the bytes
# ride the cell's input/ as staged material; the HAND stages them
# beside its submission in the machine's inbox (with its manifest
# seed) and the row's procedural mode (tree: unchanged) guards the
# run — the machine reads rc/stdout only.
#
# usage: tree-checker.sh <manifest>
# the manifest: one "sha256  path" line per file (seeded beside the
# probe). rc 0 + the fixed line when every listed file digests to
# its manifest sha; rc 1 with the first divergence named on stderr;
# rc 2 — the manifest itself is missing/unreadable (an env fact).
[ $# -ge 1 ] || { echo "usage: tree-checker.sh <manifest>" >&2; exit 2; }
m="$1"
[ -f "$m" ] || { echo "manifest missing: $m" >&2; exit 2; }
bad=0
while read -r sha path; do
  [ -n "$sha" ] || continue
  got="$(sha256sum "$path" 2>/dev/null | cut -d' ' -f1)"
  if [ "$got" != "$sha" ]; then
    echo "tree drift: $path" >&2
    bad=1
  fi
done < "$m"
[ "$bad" -eq 0 ] || exit 1
echo "tree: unchanged against the manifest"
