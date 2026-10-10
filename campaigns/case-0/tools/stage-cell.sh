#!/bin/bash
# stage-cell.sh — stage ONE cell of the case-0 campaign (the
# 0.2.0 protocol on the host world): a two-level randomly named
# directory under the run
# root's cells/ tree, identical inputs for every hand; the machine command
# is the system-installed punchtape, called by name (machine hands
# only; solo never calls it — parity is the spawn text and the cell
# carrying nothing machine-side).
#
# Cell layout (the host world):
#   <cell>/input/   brief.md, EXPECTATIONS.md, materials/, probes/
#   <cell>/env/     the task environment (prep-cell.sh fills it)
#   <cell>/work/    the hand's working tree (the machine instance of
#                   a machine hand lives at work/.punchtape)
#
# The hand-conditional bytes in this script, each named (proof of
# parity is reading this file): the spawn template choice
# (CONTRACT-MACHINE-HOST for top, CONTRACT-SOLO for solo), the binary
# sha column of the manifest, and the mechanical form-hints sidcar
# staged to machine cells only (the form-hints law —
# the machine hand's own brief lines as quotable form sources; solo
# never carries it and the sidcar is the ONLY input difference).
#
# Usage: stage-cell.sh <sense> <stack> <hand:solo|top> <run-root> \
#            [--class pilot|smoke|main]
#   (the owner's word: TWO hands —
#    solo and the machine hand on the top model; no third/flash
#    hand in this repo's reality)
# Env:  C0_DRIVER_UUID (required)
#       C0_BINARY (optional pin: the system punchtape must BE it)
set -euo pipefail

SENSE="${1:?sense}"; STACK="${2:?stack}"; HAND="${3:?hand: solo|top}"; RR="${4:?run root}"; shift 4
CLASS=main
while [ $# -gt 0 ]; do case "$1" in --class) CLASS="$2"; shift 2;; *) echo "unknown: $1" >&2; exit 2;; esac; done
case "$SENSE" in brim|conduit|errand|pressmark|quire) ;; *) echo "sense must be brim|conduit|errand|pressmark|quire" >&2; exit 2 ;; esac
case "$STACK" in rust|cpp|go|js|python) ;; *) echo "stack must be rust|cpp|go|js|python" >&2; exit 2 ;; esac
case "$HAND" in solo|top) ;; *) echo "hand must be solo|top (two hands — no third/flash hand exists)" >&2; exit 2 ;; esac
case "$CLASS" in pilot|smoke|main) ;; *) echo "class must be pilot|smoke|main" >&2; exit 2 ;; esac

HERE="$(cd "$(dirname "$0")" && pwd)"
CAMP="$(cd "$HERE/.." && pwd)"
RR="$(readlink -f "$RR")"
REPO="$(cd "$CAMP/../.." && pwd)"
if [ "$RR" = "$REPO" ] || [[ "$RR" == "$REPO"/* ]] || [[ "$REPO" == "$RR"/* ]]; then
  echo "STAGE STOP: the run root must live OUTSIDE the repo tree (got $RR)" >&2
  exit 2
fi
[ -n "${C0_DRIVER_UUID:-}" ] || { echo "C0_DRIVER_UUID is required" >&2; exit 2; }
CELL="${SENSE}-${STACK}-${HAND}"

# --- the cell: two levels of random [a-z0-9]{8} names (nothing in
# --- the name says what it is; collision checked by mkdir) under
# --- the RUN ROOT's cells/ tree — the structure law: the
# --- zone root carries only runs/ tmp/ cache/, everything of a run
# --- (cells included) lives inside its run root; the random leaf
# --- names keep the measurement blindness of the 0.2.0 protocol
rand8() {
  local s=""
  while [ ${#s} -lt 8 ]; do s="$s$(head -c 64 /dev/urandom | tr -dc 'a-z0-9')"; done
  printf '%s' "${s:0:8}"
}
mkdir -p "$RR/cells"
L1=""
for _ in 1 2 3 4 5; do CAND="$RR/cells/$(rand8)"; mkdir "$CAND" 2>/dev/null && { L1="$CAND"; break; }; done
[ -n "$L1" ] || { echo "STAGE STOP: no free random level-1 dir in $RR/cells" >&2; exit 4; }
CELLPATH=""
for _ in 1 2 3 4 5; do CAND="$L1/$(rand8)"; mkdir "$CAND" 2>/dev/null && { CELLPATH="$CAND"; break; }; done
[ -n "$CELLPATH" ] || { rmdir "$L1" 2>/dev/null || true; echo "STAGE STOP: no free cell dir" >&2; exit 4; }
mkdir -p "$CELLPATH/input" "$CELLPATH/env" "$CELLPATH/work"

# --- inputs: byte-identical for every hand of the task (the 0.2.0
# --- bytes; the brief is PER-STACK in 0.2.0 — the stack sentence
# --- and the library pins ride the brief itself)
cp "$CAMP/briefs/$SENSE/$STACK.md" "$CELLPATH/input/brief.md"
cp "$CAMP/prompts/EXPECTATIONS.md" "$CELLPATH/input/EXPECTATIONS.md"

# the form-hints sidcar: machine
# cells ONLY — a mechanical grep over the cell's own brief (verbatim
# lines a form pin can quote; the machine itself never sees the
# brief); the digest is frozen with the series and pinned in the
# manifest; the solo hand's input never carries it (parity)
FH_SHA=null
if [ "$HAND" = top ]; then
  python3 "$HERE/form-hints.py" "$CELLPATH/input/brief.md" > "$CELLPATH/input/form-hints.md"
  FH_SHA="$(sha256sum "$CELLPATH/input/form-hints.md" | cut -d' ' -f1)"
fi

# materials per sense (the 0.2.0 rule generalized to the tree:
# brim -> the two jpegs; pressmark -> the storage+templates subtree;
# quire -> the stack's own subtree; conduit/errand carry none).
# Dotfiles never ride: the staged input set is the declared files
# only (a stray dotfile — bait or editor junk — must neither enter
# a cell nor break the freeze parity; the same grammar rides
# freeze.sh)
MAT_JSON="{}"
case "$SENSE" in
  brim)
    mkdir -p "$CELLPATH/input/materials"
    cp "$CAMP"/materials/*.jpg "$CELLPATH/input/materials/" ;;
  pressmark)
    mkdir -p "$CELLPATH/input/materials"
    cp -r "$CAMP/materials/pressmark/." "$CELLPATH/input/materials/" ;;
  quire)
    mkdir -p "$CELLPATH/input/materials"
    cp -r "$CAMP/materials/quire/$STACK/." "$CELLPATH/input/materials/" ;;
esac
find "$CELLPATH/input/materials" -name '.*' -type f -delete 2>/dev/null || true
while IFS= read -r -d '' f; do
  rel="materials/${f#"$CELLPATH/input/materials/"}"
  sha="$(sha256sum "$f" | cut -d' ' -f1)"
  MAT_JSON="$(printf '%s' "$MAT_JSON" | python3 -c "
import json,sys
d=json.load(sys.stdin); d['$rel']='$sha'; print(json.dumps(d,sort_keys=True))")"
done < <(find "$CELLPATH/input/materials" -type f ! -name '.*' -print0)

# the campaign's staged probes (the witness mechanics):
# bytes only — a machine hand that wants their witness stages them
# beside its submission and names them by a witness row; the digests
# are frozen with the series. Dotfiles never ride (the staged set is
# the declared probes only — same grammar as freeze.sh)
mkdir -p "$CELLPATH/input/probes"
PROBE_JSON="{}"
while IFS= read -r -d '' p; do
  cp "$p" "$CELLPATH/input/probes/"
  psha="$(sha256sum "$p" | cut -d' ' -f1)"
  pname="$(basename "$p")"
  PROBE_JSON="$(printf '%s' "$PROBE_JSON" | python3 -c "
import json,sys
d=json.load(sys.stdin); d['$pname']='$psha'; print(json.dumps(d,sort_keys=True))")"
done < <(find "$CAMP/materials/probes" -maxdepth 1 -type f ! -name '.*' -print0 | sort -z)

# --- the machine binary: the SYSTEM punchtape, called by name; no
# --- copy rides the cell (parity = the solo cell and spawn carrying
# --- nothing machine-side); the pin, when set, must BE the system one
SYS_BIN="$(command -v punchtape || true)"
RUN_BIN=""
if [ -n "${C0_BINARY:-}" ]; then
  RUN_BIN="$(readlink -f "$C0_BINARY")"
  [ -f "$RUN_BIN" ] || { echo "C0_BINARY=$C0_BINARY does not resolve to a file" >&2; exit 2; }
fi
if [ -n "$SYS_BIN" ] && [ -n "$RUN_BIN" ]; then
  PIN_SHA="$(sha256sum "$RUN_BIN" | cut -d' ' -f1)"
  SYS_SHA="$(sha256sum "$SYS_BIN" | cut -d' ' -f1)"
  [ "$SYS_SHA" = "$PIN_SHA" ] || { echo "system punchtape ($SYS_SHA) != pinned ($PIN_SHA): install the pinned binary" >&2; exit 2; }
fi
RUN_SHA=null
if [ -n "$RUN_BIN" ]; then RUN_SHA="$(sha256sum "$RUN_BIN" | cut -d' ' -f1)"
elif [ -n "$SYS_BIN" ]; then RUN_SHA="$(sha256sum "$SYS_BIN" | cut -d' ' -f1)"; fi
BIN_SHA=null
case "$HAND" in
  top)
    [ -n "$SYS_BIN" ] || { echo "punchtape is not installed on the system (command -v punchtape)" >&2; exit 2; }
    BIN_SHA="$(sha256sum "$SYS_BIN" | cut -d' ' -f1)"
    ;;
esac

# --- the spawn message: HARNESS-HOST + the hand's contract +
# --- EXPECTATIONS (all frozen bytes; __WORK__/__INPUT__ rendered to
# --- the cell's absolute paths); the driver-side copy is archived in
# --- the run root and audited BEFORE any spawn
TASKDIR="$RR/tasks/$CELL"
mkdir -p "$TASKDIR"
SPAWN="$TASKDIR/SPAWN.txt"
case "$HAND" in
  solo) CONTRACT="$CAMP/prompts/CONTRACT-SOLO.md" ;;
  top) CONTRACT="$CAMP/prompts-host/CONTRACT-MACHINE-HOST.md" ;;
esac
{
  sed -e "s|__WORK__|$CELLPATH/work|g" -e "s|__INPUT__|$CELLPATH/input|g" \
      "$CAMP/prompts-host/HARNESS-HOST.md"
  printf '\n'
  sed -e "s|__WORK__|$CELLPATH/work|g" -e "s|__INPUT__|$CELLPATH/input|g" "$CONTRACT"
  printf '\n'
  cat "$CAMP/prompts/EXPECTATIONS.md"
} > "$SPAWN"
bash "$HERE/preflight-audit.sh" prompts "$SPAWN" \
  || { echo "STAGE STOP: preflight refused the spawn message" >&2; exit 4; }

# --- the staging-vs-freeze check (the P6 law: a new series freezes
# --- before its first cell; the staging stands on the LATEST freeze)
FREEZE="$(ls -t "$CAMP"/FREEZE-S.sha256 2>/dev/null | head -1 || true)"
if [ -f "$FREEZE" ]; then
  if ! python3 "$HERE/check-staged-freeze.py" "$FREEZE" "$SENSE" "$STACK" "$CELLPATH/input"; then
    echo "STAGE STOP: staged inputs do not match the freeze" >&2
    exit 4
  fi
else
  echo "STAGE STOP: no series freeze in $CAMP (freeze.sh FIRST — the P6 law)" >&2
  exit 4
fi

# --- the per-cell canary: one token planted in the cell root before
# --- the hand starts; the registry lives in the sealed driver zone
MZ="$RR/measurement"
mkdir -p "$MZ/sealed" "$RR/operators"
TOK="$(head -c 24 /dev/urandom | od -An -tx1 | tr -d ' \n')"
printf 'canary %s: this file proves reading where no hand may look\n' "$TOK" \
  > "$CELLPATH/.canary-${TOK:0:8}.txt"
mkdir -p "$MZ/sealed/canaries"
echo "$TOK $CELL $CELLPATH/.canary-${TOK:0:8}.txt" >> "$MZ/sealed/canaries/registry.txt"

# --- the staged manifest (never staged to a hand)
mkdir -p "$RR/manifests"
BRIEF_SHA="$(sha256sum "$CELLPATH/input/brief.md" | cut -d' ' -f1)"
EXP_SHA="$(sha256sum "$CELLPATH/input/EXPECTATIONS.md" | cut -d' ' -f1)"
SPAWN_SHA="$(sha256sum "$SPAWN" | cut -d' ' -f1)"
PRICES_SHA="$(sha256sum "$CAMP/prices.json" | cut -d' ' -f1)"
python3 - "$RR/manifests" "$CELL" "$SENSE" "$STACK" "$HAND" "$CLASS" "$CELLPATH" \
        "$BRIEF_SHA" "$EXP_SHA" "$SPAWN_SHA" "$MAT_JSON" "$BIN_SHA" "$RUN_SHA" \
        "$C0_DRIVER_UUID" "$PRICES_SHA" "$PROBE_JSON" "$FH_SHA" <<'PYEOF'
import json, sys
(mandir, cell, sense, stack, hand, cls, cellpath, brief, exp, spawn,
 mat_json, bin_sha, run_sha, drv, prices, probe_json, fh_sha) = sys.argv[1:18]
m = {
  "cell": cell, "sense": sense, "stack": stack, "hand": hand,
  "class": cls, "protocol": "0.2.0", "wave": None,
  "cell_path": cellpath, "driver_uuid": drv,
  "binary_sha256": None if run_sha == "null" else run_sha,
  "expected": {"hands": 1, "judges_this_task": None, "economy_rows": 1},
  "inputs": {
    "brief": {"path": f"briefs/{sense}/{stack}.md", "sha256": brief},
    "expectations": {"path": "prompts/EXPECTATIONS.md", "sha256": exp},
    "materials": json.loads(mat_json),
    "probes": json.loads(probe_json),
    # the form-hints sidcar: a HAND-CONDITIONAL
    # field by design — the machine hand's quotable brief lines; the
    # solo hand's input never carries the file (verify-cell pins the
    # absence)
    "form_hints": None if fh_sha == "null" else {
        "path": "input/form-hints.md", "sha256": fh_sha},
    "spawn": {"template_per_hand": True, "sha256": spawn},
    "prices": {"sha256": prices},
    "binary": None if bin_sha == "null" else {"sha256": bin_sha},
  },
  "world": {"exec": "host-direct", "network": "textual-directive",
            "versions_source": "world-lock.txt",
            "timeouts": {"row_s": 60, "build_s": 300}},
}
out = f"{mandir}/{cell}.json"
json.dump(m, open(out, "w"), indent=2, sort_keys=True)
print(f"manifest: {out}")
PYEOF

# --- parity echo for the cell journal
echo "== case-0 cell $CELL (class $CLASS) =="
echo "brief:      $BRIEF_SHA (briefs/$SENSE/$STACK.md)"
echo "expectations $EXP_SHA"
echo "materials: $(printf '%s' "$MAT_JSON" | python3 -c 'import json,sys; print(len(json.load(sys.stdin)))' 2>/dev/null || echo 0) files"
echo "probes:     $(printf '%s' "$PROBE_JSON" | python3 -c 'import json,sys; print(len(json.load(sys.stdin)))' 2>/dev/null || echo 0) files"
if [ "$HAND" = top ]; then
  echo "form-hints: $FH_SHA (the machine hand's quotable brief lines)"
else
  echo "form-hints: ABSENT for the solo hand (verified: the sidcar never rides a solo cell)"
fi
case "$HAND" in
  solo)
    if grep -q "punchtape" "$SPAWN"; then
      echo "PARITY FAIL: the solo spawn message mentions the machine" >&2; exit 1
    fi
    if [ -e "$CELLPATH/pt" ] || [ -e "$CELLPATH/punchtape" ]; then
      echo "PARITY FAIL: a machine binary rides the cell" >&2; exit 1
    fi
    echo "machine:    ABSENT for the solo hand (verified: no mention in the spawn, nothing in the cell)"
    ;;
  top)
    [ -n "$SYS_BIN" ] || { echo "PARITY FAIL: punchtape not installed on the system" >&2; exit 1; }
    IN_SHA="$(sha256sum "$SYS_BIN" | cut -d' ' -f1)"
    [ "$IN_SHA" = "$BIN_SHA" ] || { echo "PARITY FAIL: binary sha mismatch" >&2; exit 1; }
    [ -x "$SYS_BIN" ] || { echo "PARITY FAIL: punchtape not executable" >&2; exit 1; }
    echo "machine:    $IN_SHA (system punchtape, executable, verified)"
    ;;
esac
echo "cell:       $CELLPATH"
echo "spawn msg:  $SPAWN"
python3 - "$RR" "$CELL" "$CLASS" <<'PY'
import datetime, json, os, sys
root, cell, cls = sys.argv[1:4]
rec = {"op": "stage", "cell": cell, "class": cls,
       "ts": datetime.datetime.now(datetime.timezone.utc).isoformat(timespec="seconds")}
with open(f"{root}/wave-journal.jsonl", "a") as f:
    f.write(json.dumps(rec, ensure_ascii=False) + "\n")
PY
