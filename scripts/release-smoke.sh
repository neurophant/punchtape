#!/usr/bin/env bash
# release-smoke.sh - a full punchtape session with real teeth, to a
# standing verdict. Covers the three forks of the machine's life:
#   1. a refused document (broken document law) - fixed in place;
#   2. a red row caught byte-exact - the code fixed, green;
#   3. the completion act - verdict #1 standing.
# POSIX-clean on purpose: macOS bash 3.2 included (no tac, no sed -i,
# no bash-4 constructs). Runs on any unix with punchtape installed.
set -euo pipefail

export PATH="$HOME/.local/bin:$PATH"
command -v punchtape >/dev/null || { echo "punchtape is not on PATH" >&2; exit 1; }

fail() { echo "SMOKE FAIL: $*" >&2; exit 1; }
expect_in() { # file, fixed string, meaning
  grep -qF "$2" "$1" || fail "$3 (pattern not in $1)"
}

cd "$HOME"
rm -rf task-smoke
mkdir -p task-smoke && cd task-smoke

# --- the birth: the human's words, verbatim -----------------------------
mkdir -p .punchtape
printf 'A notes utility: notes add <text> appends the note and prints its number;\nnotes list prints the notes one per line, oldest first.\n' > .punchtape/demands.md
punchtape > /dev/null 2>&1 || true

# --- the counted root ----------------------------------------------------
printf 'notes.sh\n' > .punchtape/universe

# --- the utility, buggy on purpose: prints n+1 instead of n --------------
cat > notes.sh <<'SCRIPT'
#!/usr/bin/env bash
# notes - a one-file notes utility.
set -euo pipefail
file="notes.txt"
cmd="${1:-}"

add_note() {
  local n
  n=$(( $( [ -f "$file" ] && wc -l < "$file" || echo 0 ) + 1 ))
  printf '%s %s\n' "$n" "$*" >> "$file"
  printf '%s\n' "$(( n + 1 ))"
}

list_notes() {
  [ -f "$file" ] || exit 0
  cat "$file"
}

case "$cmd" in
  add) shift; add_note "$@" ;;
  list) list_notes ;;
  *) printf 'usage: notes add <text> | notes list\n'; exit 2 ;;
esac
SCRIPT
chmod +x notes.sh

# --- fork 1: a broken document is refused, then fixed in place -----------
printf 'work: {{ the notes utility, broken flow\nintents: []\n' > .punchtape/inbox.yaml
punchtape > /dev/null 2>&1 || true
expect_in .punchtape/report.md 'could not find flow map content (law:document)' \
  'fork 1: the machine refused the broken document with its exact place and law'

# the correct document: the CLI contract claimed, rows pinned byte-exact
cat > .punchtape/inbox.yaml <<'DOC'
work: the notes utility from zero - the CLI contract claimed
intents:
  - {lines: [D1], outcome: work, prose: notes add appends and prints the number}
  - {lines: [D2], outcome: work, prose: notes list prints oldest first}
remove: []
requirements:
  - id: R1
    lines: [D1]
    prose: notes add appends the note and prints its number
  - id: R2
    lines: [D2]
    prose: notes list prints the notes one per line, oldest first
behaviors:
  - id: cli/add
    name: notes add
    prose: add appends the note to notes.txt and prints its number
    claims: ["notes.sh:1-13"]
    covers: [R1]
    rows:
      - run: ./notes.sh add first
        expect: {rc: 0, out: "1\n"}
      - run: ./notes.sh add second
        expect: {rc: 0, out: "2\n"}
  - id: cli/list
    name: notes list
    prose: list prints the notes one per line, oldest first
    claims: ["notes.sh:14-18"]
    covers: [R2]
    rows:
      - run: ./notes.sh list
        expect: {rc: 0, out: "1 first\n2 second\n"}
  - id: cli/usage
    name: the usage line
    prose: no command prints the usage line to stdout and exits two
    claims: ["notes.sh:19-23"]
    rows:
      - run: ./notes.sh
        expect: {rc: 2, out: "usage: notes add <text> | notes list\n"}
DOC

# --- fork 2: the buggy add is caught red, byte-exact ----------------------
punchtape > /dev/null 2>&1 || true
expect_in .punchtape/report.md 'stdout: want "1\n", got "2\n"' \
  'fork 2: the red row shows the exact wanted and got bytes'

# the fix: the whole file rewritten, one line different (same line count,
# so the claims keep their ranges) - the way an agent fixes code
cat > notes.sh <<'SCRIPT'
#!/usr/bin/env bash
# notes - a one-file notes utility.
set -euo pipefail
file="notes.txt"
cmd="${1:-}"

add_note() {
  local n
  n=$(( $( [ -f "$file" ] && wc -l < "$file" || echo 0 ) + 1 ))
  printf '%s %s\n' "$n" "$*" >> "$file"
  printf '%s\n' "$n"
}

list_notes() {
  [ -f "$file" ] || exit 0
  cat "$file"
}

case "$cmd" in
  add) shift; add_note "$@" ;;
  list) list_notes ;;
  *) printf 'usage: notes add <text> | notes list\n'; exit 2 ;;
esac
SCRIPT
chmod +x notes.sh
# the debris of the failed runs: add appends to notes.txt, and red rows
# re-run when their claimed lines change - a clean tree makes the pins
# hold again (idempotent rows, the way the machine expects them)
rm -f notes.txt

# --- fork 3: all green, the completion act, the verdict -------------------
punchtape > /dev/null 2>&1 || true
expect_in .punchtape/state.md 'write done: true' \
  'fork 3 pre: all rows green, the machine asks for the completion act'
printf 'done: true\n' > .punchtape/inbox.yaml
punchtape > /dev/null 2>&1 || true
expect_in .punchtape/state.md 'next: verdict #1 standing' \
  'fork 3: the verdict stands and the machine shows it to the human'
grep -E '^verdict: #1 .* standing' .punchtape/state.md || fail 'the verdict line is missing'
echo "SMOKE OK: refused-fixed, red-fixed, verdict standing"
