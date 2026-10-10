#!/bin/bash
# freeze.sh — the input FREEZE of the case-0 campaign (the P6
# law: a new series freezes BEFORE its first cell): hashes every
# input artifact (the 0.2.0 byte transfers —
# briefs, materials, prompts; the named host adaptations —
# prompts-host with their provenance; prices,
# models, the world lock; this tool suite) plus the series
# binary and METHODOLOGY (ONE binary sha serves the freeze).
# The 0.2.0 protocol has no etalons/prereg/uniform-defaults of its
# own (its measurement is the measurer+judge; its uniformity is the
# frozen wave texts in the inputs) — the freeze pins BYTES ONLY.
#
# Usage: freeze.sh s <binary-path> [--out <FREEZE-file>]
set -euo pipefail
BAND="${1:?band}"; BIN="${2:?binary path}"
OUT="${4:-}"
HERE="$(cd "$(dirname "$0")" && pwd)"
CAMP="$(cd "$HERE/.." && pwd)"
REPO="$(cd "$CAMP/../.." && pwd)"
[ "$BAND" = s ] || { echo "freeze.sh: the case-0 series is the s band" >&2; exit 2; }
[ -n "$OUT" ] || OUT="$CAMP/FREEZE-${BAND^^}.sha256"
BIN="$(readlink -f "$BIN")"

{
  echo "# case-0 freeze — band $BAND"
  echo "series-binary  $(sha256sum "$BIN" | cut -d' ' -f1)  punchtape"
  echo "methodology    $(sha256sum "$REPO/campaigns/METHODOLOGY.md" | cut -d' ' -f1)"
  # the 0.2.0 byte transfers (the transfer manifest holds the source digests);
  # dotfiles never ride the freeze (the staged-input grammar of stage-cell.sh:
  # a stray dotfile — bait or editor junk — is no input of the series)
  while IFS= read -r -d '' f; do
    echo "brief          $(sha256sum "$f" | cut -d' ' -f1)  ${f#"$CAMP"/}"
  done < <(find "$CAMP/briefs" -type f ! -name '.*' -print0 | sort -z)
  # the form-hints sidcar digests:
  # the same deterministic generator over the same brief bytes the
  # stage will run — a staged sidcar that diverges is a freeze drift
  # and an unfrozen regeneration is a freeze miss (check-staged-freeze)
  while IFS= read -r -d '' f; do
    echo "form-hints     $(python3 "$HERE/form-hints.py" "$f" | sha256sum | cut -d' ' -f1)  ${f#"$CAMP"/}"
  done < <(find "$CAMP/briefs" -type f ! -name '.*' -print0 | sort -z)
  while IFS= read -r -d '' f; do
    # the staged probe materials carry their own kind (the witness
    # bytes the machine meets only through a hand's rows)
    case "$f" in "$CAMP"/materials/probes/*)
      echo "probe          $(sha256sum "$f" | cut -d' ' -f1)  ${f#"$CAMP"/}"
      continue ;;
  esac
    echo "material       $(sha256sum "$f" | cut -d' ' -f1)  ${f#"$CAMP"/}"
  done < <(find "$CAMP/materials" -type f ! -name '.*' -print0 | sort -z)
  for f in "$CAMP"/prompts/*.md; do
    echo "prompt         $(sha256sum "$f" | cut -d' ' -f1)  ${f#"$CAMP"/}"
  done
  # the named host adaptations and their provenance
  for f in "$CAMP"/prompts-host/*.md "$CAMP"/prompts-host/*.json; do
    [ -f "$f" ] || continue
    echo "prompt-host    $(sha256sum "$f" | cut -d' ' -f1)  ${f#"$CAMP"/}"
  done
  # (probes-baseline is gone from the campaign: the 0.2.0 probes
  #  targeted the 0.2.0 machine CLI, not this frozen machine)
  echo "prices         $(sha256sum "$CAMP/prices.json" | cut -d' ' -f1)  prices.json"
  echo "models         $(sha256sum "$CAMP/models.json" | cut -d' ' -f1)  models.json"
  echo "world-lock     $(sha256sum "$CAMP/world-lock.txt" | cut -d' ' -f1)  world-lock.txt"
  # the campaign's documentation bytes (the two metrics registries)
  for f in "$CAMP"/docs/*.md "$CAMP"/docs/*.json; do
    [ -f "$f" ] || continue
    echo "docs           $(sha256sum "$f" | cut -d' ' -f1)  ${f#"$CAMP"/}"
  done
  # this campaign's tool suite
  for f in "$HERE"/*.sh "$HERE"/*.py; do
    [ -f "$f" ] || continue
    echo "tool           $(sha256sum "$f" | cut -d' ' -f1)  ${f#"$CAMP"/}"
  done
} > "$OUT"
echo "frozen: $OUT ($(grep -c . "$OUT") lines)"
sha256sum "$OUT"
