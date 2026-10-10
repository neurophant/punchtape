#!/bin/bash
# gate-roundtrip.sh — the live-instance roundtrip of the make test
# gate: the ONLY sanctioned non-unit coverage remainder. The unit
# battery proves the machine over the in-memory ports; this script
# drives the OS implementations of the ports (fsport.OS, execport.OS)
# and cmd/punchtape's main through one real cycle of the real binary
# over the behavior corpus. The console is not a transport: every
# submission goes through the machine's single door — the inbox file
# (.punchtape/inbox/current.txt) that next names in its INBOX line;
# the hand overwrites the prefilled file in place and submits that
# exact path. The cycle: birth by next with the brief and the
# injected agent contract, the spec that stores the behaviors by id
# (nothing runs), the code written DIRECTLY into the repository (the
# executor's free territory — not a byte through the machine), the
# run that executes every row in the live tree (red on a wrong
# expectation, green after the repair resubmitted under the same
# behavior id), ratification, the delivery verdict. The door's
# impermeability is asserted on the way: a submission through any
# other path (a stray copy, a stale inbox name) is refused. Every
# step asserts; the first lie fails the gate.
#
# Usage: gate-roundtrip.sh <path-to-the-instrumented-binary>
set -euo pipefail

PT="${1:?usage: gate-roundtrip.sh <binary>}"
case "$PT" in /*) ;; *) PT="$PWD/$PT" ;; esac # the walk leaves the launch directory
# The instance tree must hold only the real cycle's files: the
# verdict's tree digest signs the working tree, so the run's own
# captures live OUTSIDE the instance.
mkdir -p "$HOME/.experiments/tmp"
ROOT="$(mktemp -d "$HOME/.experiments/tmp/gr.XXXXXX")"
WORK="$ROOT/instance"
OUT="$ROOT/captures"
mkdir -p "$WORK" "$OUT"
trap 'rm -rf "$ROOT"' EXIT
cd "$WORK"

DOOR=".punchtape/inbox/current.txt"

step() { printf '%s\n' "--- $*"; }

# refuse1 — assert that a submission dies with exactly rc 1 (a
# refusal, not a machine failure) and print the reply on any other
# outcome.
refuse1() {
    local path="$1" what="$2" rc=0
    "$PT" submit "$path" > "$OUT/r-refuse.txt" 2>&1 || rc=$?
    [ "$rc" -eq 1 ] || { echo "FAIL: $what was not refused with rc 1 (rc=$rc)"; cat "$OUT/r-refuse.txt"; exit 1; }
}

step "the first next with the brief creates the instance and injects the contract"
"$PT" next "a tool that greps lines in files" > "$OUT/next1.txt"
grep -q "STAGE:" "$OUT/next1.txt" || { echo "FAIL: no slot in the first next"; cat "$OUT/next1.txt"; exit 1; }
grep -q "INBOX: $DOOR" "$OUT/next1.txt" || { echo "FAIL: the slot does not name the door"; cat "$OUT/next1.txt"; exit 1; }
head -1 "$OUT/next1.txt" | grep -q '^THE WORKING CONTRACT — AGENTS.md in full' \
  || { echo "FAIL: the contract injection is not the first next's first line"; cat "$OUT/next1.txt"; exit 1; }

step "the spec (one behavior stored by id; nothing runs)"
cat > "$DOOR" <<'EOF'
kind: spec
build: [true]
behaviors:
  - id: B1
    name: line-grep
    prose: |
      Matching lines print one per line, numbered.
    rows:
      - run: [grep-tool, apple, notes.txt]
        out: "2: apple"
        rc: 0
        form: "the brief's numbered listing line"
EOF
"$PT" submit "$DOOR" > "$OUT/r-spec.txt"
grep -q "ACCEPTED: transaction" "$OUT/r-spec.txt" || { echo "FAIL: the spec did not land"; cat "$OUT/r-spec.txt"; exit 1; }
grep -q "BYTES: sha256 " "$OUT/r-spec.txt" || { echo "FAIL: the ack carries no bytes sha"; cat "$OUT/r-spec.txt"; exit 1; }
grep -q "SPEC v1" "$OUT/r-spec.txt" || { echo "FAIL: the ack carries no version counts"; cat "$OUT/r-spec.txt"; exit 1; }
# the thin-ack law: the reply carries no slot; the slot rides next
grep -q "STAGE:" "$OUT/r-spec.txt" && { echo "FAIL: the spec ack carries a slot body"; cat "$OUT/r-spec.txt"; exit 1; }
"$PT" next > "$OUT/r-spec-next.txt"
grep -q "STAGE: implement" "$OUT/r-spec-next.txt" || { echo "FAIL: the post-spec slot does not advance to implement"; cat "$OUT/r-spec-next.txt"; exit 1; }

step "the door refuses any path but the inbox file (impermeability)"
cp "$DOOR" d-stray.yaml
cp "$DOOR" .punchtape/inbox/0001-spec.yaml
for stray in d-stray.yaml .punchtape/inbox/0001-spec.yaml; do
    refuse1 "$stray" "$stray"
    grep -q "ERROR:" "$OUT/r-refuse.txt" || { echo "FAIL: the refusal of $stray is not an ERROR line"; cat "$OUT/r-refuse.txt"; exit 1; }
    grep -q "current.txt" "$OUT/r-refuse.txt" || { echo "FAIL: the refusal of $stray does not name the wanted path"; cat "$OUT/r-refuse.txt"; exit 1; }
done
rm -f d-stray.yaml .punchtape/inbox/0001-spec.yaml

step "the real sh code written directly into the repository"
cat > grep-tool <<'EOF'
#!/bin/sh
p="$1"; shift
n=0
while IFS= read -r l || [ -n "$l" ]; do n=$((n+1)); case "$l" in *"$p"*) printf "%s: %s\n" "$n" "$l";; esac; done < "$1"
EOF
chmod +x grep-tool
printf 'banana\napple\n' > notes.txt
test -x grep-tool && test -f notes.txt

step "the run with a wrong expectation is red (row B1/1; the build ran live)"
# the economy law: after the first spec the full table never rides again
# — the wrong expectation is a run PATCH by row address
cat > "$DOOR" <<'EOF'
kind: run
build: [true]
patch:
  replace:
    B1/1:
      run: [grep-tool, apple, notes.txt]
      out: "XXX"
      rc: 0
      form: "the brief's numbered listing line"
EOF
"$PT" submit "$DOOR" > "$OUT/r-run-red.txt"
grep -q "suite=red" "$OUT/r-run-red.txt" || { echo "FAIL: the wrong expectation is not red"; cat "$OUT/r-run-red.txt"; exit 1; }
grep -q "B1/1:" "$OUT/r-run-red.txt" || { echo "FAIL: the red row is not named B1/1"; cat "$OUT/r-run-red.txt"; exit 1; }

step "the repair under the same row address is green"
cat > "$DOOR" <<'EOF'
kind: run
patch:
  replace:
    B1/1:
      run: [grep-tool, apple, notes.txt]
      out: "2: apple"
      rc: 0
      form: "the brief's numbered listing line"
EOF
"$PT" submit "$DOOR" > "$OUT/r-run-green.txt"
grep -q "suite=green" "$OUT/r-run-green.txt" || { echo "FAIL: the repaired row is not green"; cat "$OUT/r-run-green.txt"; exit 1; }
# the thin-ack law stays QUALITATIVE (the owner's word: no numeric
# ack pin): the ack carries status+transaction+bytes+gates+the
# measured one-liners (and the DOOR reset line) —
# never a slot body

step "ratification (a bare run, behaviors omitted — the signature: nothing re-ran)"
cat > "$DOOR" <<'EOF'
kind: run
ratify: true
EOF
"$PT" submit "$DOOR" > "$OUT/r-ratify.txt"
grep -q "VERDICT: READY" "$OUT/r-ratify.txt" || { echo "FAIL: ratification did not verdict READY"; cat "$OUT/r-ratify.txt"; exit 1; }
grep -q "nothing re-ran" "$OUT/r-ratify.txt" || { echo "FAIL: the closing does not name its signature"; cat "$OUT/r-ratify.txt"; exit 1; }

step "the delivered slot renders the verdict"
"$PT" next > "$OUT/next2.txt"
grep -q "STAGE: deliver" "$OUT/next2.txt" || { echo "FAIL: the walk did not reach deliver"; cat "$OUT/next2.txt"; exit 1; }
grep -q "VERDICT: READY" "$OUT/next2.txt" || { echo "FAIL: the delivered slot does not render the verdict"; cat "$OUT/next2.txt"; exit 1; }
# The second next carries the one-line contract reference on the
# birth sha — never the full contract again
grep -q "CONTRACT: AGENTS.md in your tree (unchanged; sha256 " "$OUT/next2.txt" \
  || { echo "FAIL: the repeat next carries no contract reference"; cat "$OUT/next2.txt"; exit 1; }
grep -q "THE WORKING CONTRACT" "$OUT/next2.txt" \
  && { echo "FAIL: the repeat next pays for the full contract"; cat "$OUT/next2.txt"; exit 1; }

step "the crate smoke: a real toolchain build through the door"
# a fresh instance; the spec declares a real cargo build with the
# cargo cache as build-ignore; the row runs the built binary
CRATE="$ROOT/crate-instance"
mkdir -p "$CRATE/src"
cd "$CRATE"
cat > Cargo.toml <<'EOF2'
[package]
name = "tick-tool"
version = "0.1.0"
edition = "2021"
EOF2
cat > src/main.rs <<'EOF2'
fn main() { println!("tick"); }
EOF2
"$PT" next "a tool that prints one word" > "$OUT/c-next.txt" 2>&1
grep -q "STAGE:" "$OUT/c-next.txt" || { echo "FAIL: no crate-instance slot"; cat "$OUT/c-next.txt"; exit 1; }
cat > .punchtape/inbox/current.txt <<'EOF2'
kind: spec
build: [cargo, build, --release]
build-ignore: [Cargo.lock, target/]
behaviors:
  - id: B1
    name: tick
    prose: |
      Prints its one word.
    rows:
      - run: ["./target/release/tick-tool"]
        out: "tick"
        rc: 0
        form: "the brief's one word"
EOF2
"$PT" submit .punchtape/inbox/current.txt > "$OUT/c-spec.txt" 2>&1 \
  || { echo "FAIL: the crate spec did not land"; cat "$OUT/c-spec.txt"; exit 1; }
# the suite measures the built binary; the ratify signs that green
printf 'kind: run\n' > .punchtape/inbox/current.txt
"$PT" submit .punchtape/inbox/current.txt > "$OUT/c-run.txt" 2>&1 \
  || { echo "FAIL: the crate run did not land"; cat "$OUT/c-run.txt"; exit 1; }
grep -q "suite=green" "$OUT/c-run.txt" || { echo "FAIL: the crate run is not green"; cat "$OUT/c-run.txt"; exit 1; }
printf 'kind: run\nratify: true\n' > .punchtape/inbox/current.txt
"$PT" submit .punchtape/inbox/current.txt > "$OUT/c-ratify.txt" 2>&1
grep -q "VERDICT: READY" "$OUT/c-ratify.txt" \
  || { echo "FAIL: the crate cycle did not reach READY"; cat "$OUT/c-ratify.txt"; exit 1; }
cd "$WORK"

step "the living case: the next spec version at deliver is ACCEPTED"
cat > "$DOOR" <<'INNER'
kind: spec
patch:
  prose:
    B1: |
      Matching lines print one per line, numbered. V2 wording of the same behavior.
INNER
"$PT" submit "$DOOR" > "$OUT/r-spec2.txt"
grep -q "ACCEPTED" "$OUT/r-spec2.txt" || { echo "FAIL: the next spec version did not land"; cat "$OUT/r-spec2.txt"; exit 1; }
grep -q "SPEC v2" "$OUT/r-spec2.txt" || { echo "FAIL: the door reply carries no version counts"; cat "$OUT/r-spec2.txt"; exit 1; }
"$PT" next > "$OUT/r-spec2-next.txt"
grep -q "STAGE: implement" "$OUT/r-spec2-next.txt" || { echo "FAIL: the case did not reopen"; cat "$OUT/r-spec2-next.txt"; exit 1; }

step "the quiet surfaces: the batch call and the delta/quiet replies"
# The batch: spec + run in ONE door call (--run-after), the delta
# reply names only the changed rows; the quiet closing carries the
# signature and the verdict without the slot body.
cat > "$DOOR" <<'INNER'
kind: spec
patch:
  add:
    - atom: B1
      row:
        run: [grep-tool, banana, notes.txt]
        out: "1: banana"
        rc: 0
        form: "the brief's numbered listing line"
INNER
"$PT" submit "$DOOR" --run-after --delta > "$OUT/r-batch3.txt"
grep -q "SPEC v3" "$OUT/r-batch3.txt" || { echo "FAIL: the batch reply loses the version counts"; cat "$OUT/r-batch3.txt"; exit 1; }
grep -q "suite=green" "$OUT/r-batch3.txt" || { echo "FAIL: the batch reply loses the run table"; cat "$OUT/r-batch3.txt"; exit 1; }
grep -q "+1 new" "$OUT/r-batch3.txt" || { echo "FAIL: the batch reply loses the new-row counts"; cat "$OUT/r-batch3.txt"; exit 1; }
"$PT" next --delta > "$OUT/r-batch3-next.txt"
grep -q "B1/2" "$OUT/r-batch3-next.txt" || { echo "FAIL: the post-batch delta does not name the new row"; cat "$OUT/r-batch3-next.txt"; exit 1; }
printf 'kind: run\nratify: true\n' > "$DOOR"
"$PT" submit "$DOOR" --quiet > "$OUT/r-ratify2.txt"
grep -q "VERDICT: READY" "$OUT/r-ratify2.txt" || { echo "FAIL: v3 did not verdict READY"; cat "$OUT/r-ratify2.txt"; exit 1; }
# the verdict version counts RATIFICATIONS (v2 = the second frozen
# snapshot), the spec's v3 naming rides the diff block above
grep -q "(v2)" "$OUT/r-ratify2.txt" || { echo "FAIL: the round's verdict carries no version line"; cat "$OUT/r-ratify2.txt"; exit 1; }
grep -q "kept" "$OUT/r-ratify2.txt" || { echo "FAIL: the run table does not name the kept rows"; cat "$OUT/r-ratify2.txt"; exit 1; }

step "both verdict snapshots are frozen; a run at deliver is refused"
for v in v1 v2; do
    test -f ".punchtape/verdicts/$v.txt" || { echo "FAIL: verdicts/$v.txt missing"; exit 1; }
    grep -q "VERDICT: READY" ".punchtape/verdicts/$v.txt" || { echo "FAIL: $v is not READY"; cat ".punchtape/verdicts/$v.txt"; exit 1; }
done
before=$(sha256sum .punchtape/verdicts/v1.txt | cut -d' ' -f1)
printf 'kind: run\n' > "$DOOR"
refuse1 "$DOOR" "a run at deliver (only the next spec version reopens)"
grep -q "ERROR:" "$OUT/r-refuse.txt" || { echo "FAIL: the run-at-deliver refusal is not an ERROR line"; cat "$OUT/r-refuse.txt"; exit 1; }
test "$(sha256sum .punchtape/verdicts/v1.txt | cut -d' ' -f1)" = "$before" \
  || { echo "FAIL: snapshot v1 mutated"; exit 1; }

step "the report: the door counters and the full table; --since aims at the delta"
"$PT" report > "$OUT/r-report.txt"
grep -q "door view" "$OUT/r-report.txt" || { echo "FAIL: the default report is not the door view"; cat "$OUT/r-report.txt"; exit 1; }
grep -q "green rows:" "$OUT/r-report.txt" || { echo "FAIL: the door report carries no green counter"; cat "$OUT/r-report.txt"; exit 1; }
"$PT" report --full > "$OUT/r-report-full.txt"
grep -q "nothing re-ran" "$OUT/r-report-full.txt" || { echo "FAIL: the full report does not name its law"; cat "$OUT/r-report-full.txt"; exit 1; }
cp "$OUT/r-report-full.txt" "$OUT/r-report.txt"
grep -q "B1/1" "$OUT/r-report.txt" || { echo "FAIL: the report does not list the row"; cat "$OUT/r-report.txt"; exit 1; }
grep -q "ran-at" "$OUT/r-report.txt" || { echo "FAIL: the report carries no run time"; cat "$OUT/r-report.txt"; exit 1; }
grep -q "fresh" "$OUT/r-report.txt" || { echo "FAIL: the report carries no freshness mark"; cat "$OUT/r-report.txt"; exit 1; }
"$PT" report --since v1 > "$OUT/r-report-since.txt"
grep -q "B1/2" "$OUT/r-report-since.txt" || { echo "FAIL: the since-report does not name the new row"; cat "$OUT/r-report-since.txt"; exit 1; }
grep -q "new since v1" "$OUT/r-report-since.txt" || { echo "FAIL: the since-report carries no section"; cat "$OUT/r-report-since.txt"; exit 1; }
rc=0; "$PT" report --since v9 > "$OUT/r-report-v9.txt" 2>&1 || rc=$?
[ "$rc" -eq 1 ] || { echo "FAIL: an unknown verdict must refuse with rc 1 (rc=$rc)"; cat "$OUT/r-report-v9.txt"; exit 1; }

step "the form gate: a formal pin without its source closes the run door"
FGATE="$ROOT/formgate-instance"
mkdir -p "$FGATE"
cd "$FGATE"
"$PT" next "a tool that greps lines" > "$OUT/m4-next.txt" 2>&1
grep -q "STAGE:" "$OUT/m4-next.txt" || { echo "FAIL: no formgate-instance slot"; cat "$OUT/m4-next.txt"; exit 1; }
cp "$WORK/grep-tool" "$WORK/notes.txt" .
# the spec pins exact surface with no source line: it LANDS, the
# marker names the row in the ratification render
cat > .punchtape/inbox/current.txt <<'EOF2'
kind: spec
behaviors:
  - id: B1
    name: line-grep
    prose: |
      Matching lines print one per line, numbered.
    rows:
      - run: [grep-tool, apple, notes.txt]
        out: "2: apple"
        rc: 0
EOF2
"$PT" submit .punchtape/inbox/current.txt > "$OUT/m4-spec.txt" 2>&1 \
  || { echo "FAIL: the unpinned spec did not land"; cat "$OUT/m4-spec.txt"; exit 1; }
grep -q "FORM DECISIONS UNRESOLVED: B1/1" "$OUT/m4-spec.txt" \
  || { echo "FAIL: the spec ack does not mark the unresolved pin"; cat "$OUT/m4-spec.txt"; exit 1; }
# the run door stays closed while the pin is unresolved
printf 'kind: run\n' > .punchtape/inbox/current.txt
refuse1 .punchtape/inbox/current.txt "the run of an unresolved form pin"
grep -q "FORM DECISIONS UNRESOLVED: B1/1" "$OUT/r-refuse.txt" \
  || { echo "FAIL: the run refusal does not name the unresolved pin"; cat "$OUT/r-refuse.txt"; exit 1; }
# the source line arrives: the same row address carries it, the run opens
cat > .punchtape/inbox/current.txt <<'EOF2'
kind: spec
patch:
  replace:
    B1/1:
      run: [grep-tool, apple, notes.txt]
      out: "2: apple"
      rc: 0
      form: "the brief's numbered listing line"
EOF2
"$PT" submit .punchtape/inbox/current.txt > "$OUT/m4-spec2.txt" 2>&1 \
  || { echo "FAIL: the sourced spec did not land"; cat "$OUT/m4-spec2.txt"; exit 1; }
grep -q "FORM DECISIONS UNRESOLVED" "$OUT/m4-spec2.txt" \
  && { echo "FAIL: the sourced spec still carries the marker"; cat "$OUT/m4-spec2.txt"; exit 1; }
printf 'kind: run\n' > .punchtape/inbox/current.txt
"$PT" submit .punchtape/inbox/current.txt > "$OUT/m4-run.txt" 2>&1 \
  || { echo "FAIL: the sourced run did not land"; cat "$OUT/m4-run.txt"; exit 1; }
grep -q "suite=green" "$OUT/m4-run.txt" || { echo "FAIL: the sourced run is not green"; cat "$OUT/m4-run.txt"; exit 1; }
cd "$WORK"

step "the typed rows and the dry validator: refuse/det/witness/repeat-run through the same door"
# the product gains an honest usage refusal (the code is free territory)
cat > grep-tool <<'EOF2'
#!/bin/sh
if [ $# -lt 2 ]; then echo "grep-tool: usage: grep-tool pattern file" >&2; exit 2; fi
p="$1"; shift
n=0
while IFS= read -r l || [ -n "$l" ]; do n=$((n+1)); case "$l" in *"$p"*) printf "%s: %s\n" "$n" "$l";; esac; done < "$1"
EOF2
chmod +x grep-tool
# the dry validator: a broken draft names its findings BEFORE the
# door (rc 1, one ERROR line per finding) - first the typed
# strictness itself (a decode refusal names the field), then the
# document-level laws over decodable typed rows
cat > draft.yaml <<'EOF2'
kind: spec
behaviors:
  - id: B1
    name: typed families
    prose: |
      The families measured by the machine itself.
    rows:
      - refuse:
          args: [grep-tool]
          says: 'grep-tool: usage: grep-tool pattern file'
EOF2
rc=0; "$PT" spec --check draft.yaml > "$OUT/check-strict.txt" 2>&1 || rc=$?
[ "$rc" -eq 1 ] || { echo "FAIL: the field-less refuse must refuse with rc 1 (rc=$rc)"; cat "$OUT/check-strict.txt"; exit 1; }
grep -q "the fields args, says and rc are all required" "$OUT/check-strict.txt" \
  || { echo "FAIL: the typed strictness does not name the field"; cat "$OUT/check-strict.txt"; exit 1; }
cat > draft.yaml <<'EOF2'
kind: spec
behaviors:
  - id: B1
    name: typed families
    prose: |
      The families measured by the machine itself.
    rows:
      - witness:
          probe: ""
EOF2
rc=0; "$PT" spec --check draft.yaml > "$OUT/check-strict-w.txt" 2>&1 || rc=$?
[ "$rc" -eq 1 ] || { echo "FAIL: the probe-less witness must refuse with rc 1 (rc=$rc)"; cat "$OUT/check-strict-w.txt"; exit 1; }
grep -q "probe is required" "$OUT/check-strict-w.txt" \
  || { echo "FAIL: the witness strictness does not name the probe"; cat "$OUT/check-strict-w.txt"; exit 1; }
printf 'junk bytes\n' > junk.dat
cat > draft.yaml <<'EOF2'
kind: spec
build-ignore: [junk.dat]
behaviors:
  - id: B1
    name: typed families
    prose: |
      The families measured by the machine itself.
    rows:
      - refuse:
          args: [grep-tool]
          says: 'grep-tool: usage: grep-tool pattern file'
          rc: 2
        form: "the brief's usage refusal wording"
      - det:
          run: [grep-tool, apple, notes.txt]
          same: PASS
      - witness:
          probe: file://absent-probe.sh
        launch: nonet
        seeds:
          - path: in/given.txt
            from-input: missing.dat
      - run: [grep-tool, x9, notes.txt]
        out: (contains) x9
      - run: [grep-tool, junk.dat, notes.txt]
EOF2
rc=0; "$PT" spec --check draft.yaml > "$OUT/check-bad.txt" 2>&1 || rc=$?
[ "$rc" -eq 1 ] || { echo "FAIL: the broken draft must refuse with rc 1 (rc=$rc)"; cat "$OUT/check-bad.txt"; exit 1; }
for want in "echoes a self-report line" "shorter than" "has no such file" "outside the accounted surface" "probes arrive as staged bytes through the door"; do
    grep -q "$want" "$OUT/check-bad.txt" || { echo "FAIL: the broken draft loses the finding: $want"; cat "$OUT/check-bad.txt"; exit 1; }
done
rm -f junk.dat
# the witness probe rides the door: its bytes are staged next to the
# submission in the inbox (what the probe checks is its own business;
# the machine reads rc/stdout only)
cat > .punchtape/inbox/net-probe.sh <<'EOF2'
#!/bin/sh
# the probe's bytes decide what it checks; the machine reads rc/stdout only
./grep-tool apple notes.txt >/dev/null || exit 9
echo "probe: apple found"
EOF2
chmod +x .punchtape/inbox/net-probe.sh
cat > draft.yaml <<'EOF2'
kind: spec
behaviors:
  - id: B1
    name: typed families
    prose: |
      The families measured by the machine itself.
    rows:
      - refuse:
          args: [grep-tool]
          says: 'grep-tool: usage: grep-tool pattern file'
          rc: 2
        form: "the brief's usage refusal wording"
      - det:
          run: [grep-tool, apple, notes.txt]
          same: "2: apple"
        form: "the brief's numbered listing line"
      - witness:
          probe: file://net-probe.sh
        launch: nonet
        tree: unchanged
        out: "probe: apple found"
        form: "the campaign's probe answer format"
      - repeat-run: [grep-tool, apple, notes.txt]
        out: "2: apple"
        form: "the brief's numbered listing line"
EOF2
"$PT" spec --check draft.yaml > "$OUT/check-good.txt" \
  || { echo "FAIL: the clean draft did not pass the check"; cat "$OUT/check-good.txt"; exit 1; }
grep -q "PASS" "$OUT/check-good.txt" || { echo "FAIL: the pass line is missing"; cat "$OUT/check-good.txt"; exit 1; }
# the passed bytes go through the door as the next version and run
# green: the refusal refuses, the pin pins, the staged probe runs under
# the row's procedural modes, the repeat agrees with itself
# the deliberate full rewrite hatch: the typed families REPLACE the
# table (an implicit full repeat is refused by the economy law)
sed '1a full-replace: true' draft.yaml > "$DOOR"
"$PT" submit "$DOOR" > "$OUT/r-spec4.txt" \
  || { echo "FAIL: the typed spec version did not land"; cat "$OUT/r-spec4.txt"; exit 1; }
grep -q "SPEC v4" "$OUT/r-spec4.txt" || { echo "FAIL: the typed version carries no diff block"; cat "$OUT/r-spec4.txt"; exit 1; }
printf 'kind: run\n' > "$DOOR"
"$PT" submit "$DOOR" > "$OUT/r-run4.txt"
grep -q "suite=green" "$OUT/r-run4.txt" || { echo "FAIL: the typed rows did not run green"; cat "$OUT/r-run4.txt"; exit 1; }
# without the staged bytes the probe row is an honest red: probes
# arrive through the door, never from the machine
rm .punchtape/inbox/net-probe.sh
printf 'kind: run\n' > "$DOOR"
"$PT" submit "$DOOR" > "$OUT/r-run4-red.txt"
grep -q "suite=red" "$OUT/r-run4-red.txt" || { echo "FAIL: the vanished probe did not redden the row"; cat "$OUT/r-run4-red.txt"; exit 1; }
grep -q "B1/3" "$OUT/r-run4-red.txt" || { echo "FAIL: the probe row is not named B1/3"; cat "$OUT/r-run4-red.txt"; exit 1; }
grep -q "missing from the inbox" "$OUT/r-run4-red.txt" \
  || { echo "FAIL: the vanished probe red does not name the inbox"; cat "$OUT/r-run4-red.txt"; exit 1; }
# the probe's bytes return through the door: the red row re-measures green
cat > .punchtape/inbox/net-probe.sh <<'EOF2'
#!/bin/sh
# the probe's bytes decide what it checks; the machine reads rc/stdout only
./grep-tool apple notes.txt >/dev/null || exit 9
echo "probe: apple found"
EOF2
chmod +x .punchtape/inbox/net-probe.sh
printf 'kind: run\n' > "$DOOR"
"$PT" submit "$DOOR" > "$OUT/r-run4-green.txt"
grep -q "suite=green" "$OUT/r-run4-green.txt" || { echo "FAIL: the restored probe did not re-measure green"; cat "$OUT/r-run4-green.txt"; exit 1; }
printf 'kind: run\nratify: true\n' > "$DOOR"
"$PT" submit "$DOOR" > "$OUT/r-ratify4.txt"
grep -q "VERDICT: READY" "$OUT/r-ratify4.txt" || { echo "FAIL: the typed round did not close"; cat "$OUT/r-ratify4.txt"; exit 1; }
rm -f draft.yaml

step "the ceremony meter: every reply-bearing ledger entry carries its render family"
n_rb=$(grep -cE 'reply-bytes: [0-9]' .punchtape/ledger.yamll || true)
n_rn=$(grep -c 'render:' .punchtape/ledger.yamll || true)
[ "$n_rb" -eq "$n_rn" ] || { echo "FAIL: render families ($n_rn) != reply-bearing entries ($n_rb)"; exit 1; }
for fam in contract slot ack; do
    grep -q "render: $fam" .punchtape/ledger.yamll \
      || { echo "FAIL: the ledger carries no $fam-family reply"; exit 1; }
done

printf '%s\n' "ROUNDTRIP OK: a real cycle of the real binary — the OS paths of the ports are covered"
