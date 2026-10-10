The grammar of the one submission document (the door file):

## The form

One text file at .punchtape/inbox/current.txt, YAML-shaped: two-space
indents, dash lists, literal block scalars (prose: |). A kind: spec
submission is validated without running anything; a kind: run
submission executes. The FIRST spec is the one FULL document; every
later submission rides the patch body (see the patches section).

    kind: spec                    # spec | run - exactly two kinds
    build: [make, build]   # optional build argv
    build-ignore: [__pycache__, "*.pyc"]       # globs outside the tree
    behaviors:                    # spec: 1+ required; run: optional
      - id: B1                    # run: absent behaviors = run the stale rows
        name: chain-passthrough   # living name (non-empty)
        prose: |                  # MANDATORY human-readable prose
          ...
        rows:                     # 1+ rows per behavior
          - run: ["app", "relay", "a", "b"]   # argv list OR string
            seeds:                # optional files written before
              - path: in.txt      # the row, removed after it
                text: |           # inline content (text)
                  ...
              - path: big.bin
                from-input: given.bin  # a copy from the input/ directory
        out: "pervaya"            # exact | "(contains) frag" | ""
        err: "(contains) usage:"  # optional, same two forms
        rc: 0                     # optional, default 0
        state:                    # optional file checks after the row
          - path: out.txt
            text: pervaya         # exact (EOL-tolerant) | contains
          - path: result.bin
            from-run: cmd-artifact.bin   # byte-equal with a run file
        bytes: file://big-ref.bin # sample next to the submission
        timeout: 60               # optional row ceiling, 1..300 s
    answers: []                   # question batches: none open (refused)
    ratify: true                  # ONLY in kind: run: sign the recorded green
    full-replace: true            # ONLY on a deliberate FULL rewrite (see patches)

## Patches (the sparse later submissions)

The full table rides the FIRST spec exactly once (the economy law):
every later submission is a PATCH — the change only, addressed by the
row labels the machine itself prints (BEHAVIOR-ID/N, one-based, the
corpus as it stands). A replacement edits its row IN PLACE, an added
row lands at the END of its table, a removal retires its row (a
visible frame; the next verdict explains it) — the byte order of the
table never rewrites, so your platform cache keeps reading the same
prefix. The kinds stay exactly two: kind: spec validates the change,
kind: run applies it and executes (the stale rows, plus every row the
run section names — fresh or not).

    kind: run                    # or spec: the change without the run
    patch:
      replace:                   # full new row bytes at the address
        B1/2:
          run: ["./tool", "arg"]
          out: "new expected stdout"
          form: "the brief phrase this pin quotes"
      add:                       # rows to the END of their table
        - atom: B1               # a NEW id opens the behavior
          row:
            run: "./tool extra"
            out: "(contains) extra done"
        - atom: B9               # an opened behavior carries name+prose
          name: new-capability
          prose: |
            One paragraph of human-readable prose.
          row:
            run: "./tool new"
      remove: ["B2/1"]           # rows retired by address
      run:                       # rows to execute NOW, fresh or not
        B1/1:
          launch: nonet          # optional mode replacement (or tree:)
      prose:                     # the behavior prose replacement
        B1: |
          The verdict-gate prose rides the patch body, once per behavior.

- one address, one operation: a label may ride at most one of
  replace/remove/run in the same patch;
- addresses read the corpus at the patch's ARRIVAL (after a removal
  the labels shift — always re-read the table the last slot printed);
- a patch never empties a behavior through kind: run (a run never
  retires) — emptying a behavior is a spec patch or a full-replace;
- idempotency: the same patch bytes again are an honest replay
  (already applied, nothing re-applied);
- the deliberate FULL rewrite hatch: a full behaviors document with
  full-replace: true is accepted as an explicit act — the ack counts
  "old N rows -> new M rows" and the event rides the cell ledger;
  never compose full-replace with a patch;
- a repeated FULL document without the mark is refused: "repeated
  full document forbidden — the full table rode the first spec
  exactly once; submit a patch (patch: replace/add/remove/run by row
  address) or, for a deliberate full rewrite, spell full-replace:
  true".

## Field notes

- kind - exactly two kinds: spec (the primary full document, and every
  later patch of it) and run (re-submission with execution).
- build - optional argv list; rc must be 0 within 300 s; runs once,
  before the rows.
- build-ignore - optional globs; matched paths sit outside the
  accounted tree (build and run garbage).
- behaviors[].id - your short id: [A-Za-z0-9][A-Za-z0-9._-]*,
  at most 32 chars, unique in the document and in lower case against
  the corpus (the canon file name is the id's lower-case spelling).
  THE repair identity.
- behaviors[].name - a living, non-empty name.
- behaviors[].prose - MANDATORY human-readable prose per behavior; a
  behavior without prose does not validate.
- behaviors[].rows - one or more check rows per behavior; a behavior
  without rows does not validate.

## Row fields

A row carries exactly ONE head - the general `run` command or one of
the typed kinds below. Seeds and timeout compose with every head; the
typed heads own every other field of their row.

- run - the command: an argv list, or one string (split by spaces,
  quotes group). ONE command per row: orchestrate inside the command,
  e.g. run: "sh -c 'this ... & that ...; caller'".
- seeds - optional files written before the row and removed after; a
  seed over an existing file restores the original bytes. Two forms:
  text (inline content) and from-input (file-to-file copy from the
  the instance's input/ directory).
- out / err - optional. Exact text (ONE trailing newline tolerated),
  "(contains) frag" (the fragment must appear), or "" (an explicitly
  empty stream). Absent = not checked.
- rc - optional expected exit code, default 0.
- state - optional file checks after the row: text (exact,
  EOL-tolerant, or the contains form) or from-run (byte-equal with a
  file the run produced).
- bytes - optional: file://NAME names a sample file next to the
  submission in .punchtape/inbox/; stdout must equal its bytes
  exactly (the only binary form).
- timeout - optional row ceiling, 1..300 s, default 60.
- form - the one-line source of the row's formal choices (the
  form law): the brief/prereg phrase the exact pins quote, or the
  operator's one-line decision quoted as form: "op(<date>):
  <decision>" (the escalation's resolution rides the same line; the
  machine has no question channel - the marker names the rows, the
  source is the hand's to carry). The machine never judges the
  citation's content - presence is the whole mechanism. A row that
  pins formal surface — a nonzero rc, or an exact (non-contains)
  out/err/state text — WITHOUT a form line is a FORM DECISION
  UNRESOLVED: the spec lands, the marker names the rows, and runs
  stay closed until each pin carries its source line.
- launch - optional procedural launch mode: "nonet" (see the modes
  section below).
- tree - optional procedural tree mode: "unchanged" (see the modes
  section below).
- answers - answers to machine question batches; none are open, so
  the field is refused.
- ratify - ONLY in kind: run: true signs the recorded green - a
  signature, not a run: nothing executes; a stale row or a tree that
  moved since the last green run refuses with the names (run first,
  then ratify).

## Typed rows (the mechanical families)

The families every executor wrote by hand as four-field rows are
heads of the grammar themselves - strictly checked, no shorthand, no
bundle ("refusals: standard" does not exist: every refusal is its own
row with its own words). What a row MEANS stays in the prose; the
type covers only the mechanics.

- refuse: {args, says, rc} - one refusal: the invocation `args`, the
  one trouble line it must print (`says`, exact, on the error
  stream), the exit code `rc`. The answer stream stays empty by the
  family's own law. refuse IS the row {run: args, out: "", err:
  says, rc: rc} - one row, one canon, no expansion.
- det: {run, same} - the determinism pin: the recorded expectation
  IS the witness (a repeat that diverges reds against it). det IS the
  row {run: run, out: same} (rc 0).
- witness: {probe} - the staged-bytes probe head: the probe's bytes
  (the checker program) ride THE DOOR - `probe: file://NAME` names a
  file next to the submission in .punchtape/inbox/. The machine
  stages those bytes into the tree for the row's duration, runs them
  under the row's procedural modes and compares rc/stdout with the
  row's want (out composes; rc composes). What the probe checks is
  not the machine's knowledge: the machine reads only the universal
  observables. Probes arrive as staged bytes through the door, never
  from the machine itself.
- repeat-run: <command> - the idempotency witness: the command runs
  TWICE in immediate succession over what it leaves (the seeds stand
  for both runs). Green is agreement: both runs exit 0 and print
  byte-identical stdout (out, when spelled, demands its text both
  times); divergence is a red quoting both answers.

## Procedural modes (launch / tree)

Two modes compose with EVERY head (the general run and the typed
kinds alike). A mode is a runner discipline - mechanics, not meaning:
whatever a mode is FOR lives in the probe or the row's own
expectations.

- launch: nonet - the row's command launches inside a fresh
  network-less namespace: nothing is wrapped around the command, the
  network simply is not there; a host that cannot raise the namespace
  answers an honest environment red.
- tree: unchanged - the row-scoped tree identity control: outside
  .punchtape/, build-ignore and the row's own seeds, nothing the row
  leaves behind survives it (the red line "tree moved during the row"
  names the paths).

## Negative obligations (witness rows)

Every "must not happen" of the brief is a row too: prose without a
witness row proves nothing, and a harness self-report is not a
witness (see the expectation laws of the form). The grammar carries
the three canonical witness shapes:

- WRITES - "writes nothing outside <zone>", "the tree stays put":
  the machine's own gate is the witness. The tree may not move
  during a run outside .punchtape/ and build-ignore (the red line
  "tree moved" names the paths), so a row that exercises the
  scenario proves the prohibition; a byte-level pin of what MAY
  change rides state (exact text or from-run) on the named files.
- NETWORK - "touches no network": run the scenario with the network
  syscalls traced and pin the EMPTY trace ("" is the explicit empty
  stream; both the rc and the state carry the emptiness):

      - run: "sh -c 'strace -f -e trace=network -o net.log ./app fetch'; test ! -s net.log"
        out: ""
        rc: 0
        state:
          - path: net.log
            text: ""

  (or run the scenario inside a network-less namespace - unshare -n -
  and expect the product's behavior under it).
- READ SCOPE - "reads nothing outside <folder>": trace the open
  calls, filter the allowed prefix, pin the empty tail:

      - run: "sh -c 'strace -f -e trace=openat -o reads.log ./app build src out && ! grep -v \"src/\" reads.log'"
        out: ""
        rc: 0

The orchestration lives inside the run command (sh -c is the row's
own business); the expectations stay exact - out, err, rc, state. The
witness head and the procedural launch/tree modes carry the
staged-probe and the machine-discipline shapes natively (see the
typed rows and the modes above).

## The dry validator

`punchtape spec --check <file>` runs the door's form rules over a
draft BEFORE the door: grammar, typed heads, the expectation laws
(echo, loose fragments), the surface law (build-ignore vs the files a
row names), the existence of from-input seeds and bytes samples - every finding on its own line. A document that passes is taken by the
door on the first try (write those bytes into the door file); a
rejection cycle never costs a submission again. The stage is the
door's own live fact - the check is the document's.

## Rules

- THE FORM GATE: every formal pin (a nonzero rc, an exact
  out/err/state text) presents its source (form:) BEFORE any run
  measures it; the run door closes on silence, never on a wrong
  value — an operator decision rides the same line as
  form: "op(<date>): <decision>".

- THE SPEC IS THE STATE (kind: spec): the FIRST spec document carries
  EVERY behavior of the case - an id absent from it is RETIRED (a
  visible journal frame; the next verdict explains every retired row),
  a shrunk row list retires the tail rows. Every LATER spec is a patch
  (see the patches section) or a deliberate full-replace. A
  ratification never closes the case: the next spec version enters
  through the same door and the work reopens (a run at deliver is
  refused - only a spec version reopens).
- Re-submission of kind run replaces behaviors BY ID (upsert):
  same id = new rows and prose for that behavior; new ids add
  behaviors; a run never retires anything. A cheap partial upsert
  stays legal; a run whose behaviors cover the WHOLE corpus is a
  repeated full document and is refused without full-replace: true.
- Results are MEMOIZED BY CONTENT: a row keeps its green while its
  definition and the digests of the files its command names in the
  work tree (plus its from-input seeds, its byte sample and the
  declared build) are unchanged. A run executes only stale, new and
  changed rows - kept greens are named in the run table. A red row
  always re-runs; touch a named input and its row goes stale.
- The machine's row records (outcome/reason/env-fail and the memo
  fields ran-at/tree/inputs) are never submission content: a
  document carrying them is refused.
- Idempotency: the same document content again is an honest replay -
  patches included (the same patch bytes = already applied) - EXCEPT
  the bare re-run driver (kind: run with no behaviors and no build
  declaration): it always applies (a fresh transaction) and executes
  what its memos stale.
- The tree may not move during a run outside .punchtape/ and
  build-ignore: the red line "tree moved" names the paths.
- The door never reads its own blank template: a door file that still
  carries the template (or nothing at all) is refused with one line -
  after every accepted submission the door RESETS to the blank
  template, so OVERWRITE it with your next document.
