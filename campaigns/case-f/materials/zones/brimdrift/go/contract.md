punchtape — the contract

You just ran punchtape and it printed this page: there is no
instance in this directory yet. Nothing exists and nothing was
created. TO START: create the file .punchtape/wish.md with the
human's wish copied verbatim into it, and run punchtape again —
that run births the instance, fixes the goal forever and writes
the zone's standing files.

THE CALL, always. One call, no arguments, ever: `punchtape`. The
terminal gets exactly one line — the next action to take, named
with the file to touch. Every further answer is a file of this
zone: state.md holds the standing and the next: instruction,
report.md holds the last run's facts.

THE CYCLE, after the birth. Submit a document → its rows run →
fix code or rows until green → copy corpus.yaml byte-for-byte
into inbox.yaml and run: that submission is the completion act,
the machine records the verdict. state.md then shows the verdict
and any drift since. At every point the terminal line and
state.md's next: line say exactly what to do now.

law:wish      The human's wish, verbatim, into .punchtape/wish.md.
              The next run births the instance and fixes the goal
              forever; the slot is dead after the birth. No wish —
              no birth, nothing else ever creates the instance.
              At the birth the machine numbers every non-empty
              wish line W1..Wm in file order, and the numbers are
              fixed forever: the wish's bytes never change again.

law:document  All work is one document (the shape is below). Copy
              template.yaml, edit, submit as .punchtape/inbox.yaml.
              A submission MERGES by id: each listed behavior is
              added, or its head replaces the standing one's and
              its rows merge by run — the same run replaces the
              whole row, a new run is appended, unlisted rows
              stand. Requirements merge by id the same way; a
              carried not-a-requirement list or artifacts list
              replaces the standing one whole. What is not listed
              stands untouched; a subset of behaviors is a normal
              submission. The number of submissions does not
              affect the price: the listed bytes travel once,
              unchanged rows are not re-sent. Removal is only an
              explicit remove list of standing ids — behaviors or
              requirements; retiring a requirement demands the
              same submission clear every cover of it. notes are
              free text, never interpreted.

              The registry accounts for the wish: requirements
              name the brief lines they occupy as W ranges;
              behaviors name the requirements they prove in
              covers; lines that demand nothing are marked
              not-a-requirement — the lines' own words quoted
              verbatim beside a reason; artifacts declare the
              work's outputs. Every row carries an explicit
              expect — rc, out or files — and every behavior
              carries at least one row: rows are all a behavior
              proves.

law:paths     Every path — expect files, scope globs, artifacts —
              is relative to the working root. Absolute, "..", or
              escaping spellings are refused.

law:done      Submitting the corpus's own bytes (corpus.yaml
              copied into inbox.yaml unchanged) is the completion
              act that records the verdict. The verdict needs
              every row green on the current fingerprints, the
              wish whole — every W line inside some requirement's
              span or marked not-a-requirement —, every standing
              requirement named in the covers of a standing
              behavior that carries rows, and no uncovered source
              file; declared artifacts are the document's signed
              outputs. The wish's size decides the corpus's size
              — not taste. Anything short means "not yet" —
              report.md names the holes.

law:your-tree Rows and the build run in your working tree, from
              its root, with PATH and HOME, plus the tree's own
              ENV file when it carries one (one KEY=VALUE a line,
              exported as written) — nothing else is passed. Call
              what your project needs by explicit relative path
              (./tool, bin/helper) and put such files into
              the behavior's scope. The build, when declared, runs
              FIRST; its failure stops the rows this run. Clean up
              inside the commands themselves: files left behind
              change the tree.

SCOPE — what a row depends on. Each behavior's scope lists its
files as root-relative globs ("**" crosses directories). A row
re-runs when: its bytes changed, its behavior's scope files
changed, the environment changed, or it never ran. Unchanged scope
— the row never runs again. A behavior with rows and NO scope
re-runs every run: always give rows a scope. Files covered by no
scope are split in state.md: those the document's artifacts
declare stand under "artifacts (declared)", the rest are
uncovered sources — cover them with a scope or declare them, or
they hold the verdict.

REFUSALS. A refused submission stays in inbox.yaml — the document's
bytes write nothing; the machine journals the refusal as its own
fact and state.md counts it. report.md names the field, the fact
and the law id. Fix the named place in the same file and run
again — a refusal costs one run, nothing is lost. Red rows stand in report.md
grouped by reason.

THE DOCUMENT'S SHAPE (copy template.yaml and edit):

  notes: [{text: a free-text note}]
  build: "./build all"
  remove: [old/thing, R2]
  requirements:
    - id: R1                 # R then letters, digits, dots
      brief: [W1..W4, W9]    # the wish lines this requirement occupies
  not-a-requirement:
    - lines: [W5]
      quote: "a heading"     # the marked lines' own text, verbatim
      reason: a heading, not a demand
  artifacts: [out/**]        # the work's declared outputs
  behaviors:
    - id: tool/build         # [a-z0-9-] segments joined by /
      name: one line
      prose: what the user gets and why
      scope: [src/**]        # the behavior's files, globs from the root
      covers: [R1]           # the requirements this behavior proves
      rows:
        - run: "./tool src out"
          expect: {rc: 0, out: "contains: ok",
                   files: {out/built.txt: "contains: done"}}

Every expect field is optional, but at least one of rc, out,
files must stand — an empty expect is refused; rc absent wants 0;
a value with the "contains: " prefix matches a substring, without
it — the exact bytes. run is one command: a shell line or an argv
list. W ranges are W<n> or W<a>..W<b>, ascending, comma-joined —
they name the wish's numbered lines and nothing else.

ZONE FILES. contract.md (this page; contract-ru.md, contract-zh.md
are its mirrors) template.yaml | wish.md inbox.yaml — you write |
corpus.yaml the standing corpus | state.md standing, drift, the
brief map, next step | report.md the last run or the refusals |
knowledge.md notes and lessons | spec/ one page per behavior |
journal/ the canon — never edit | cache/ tmp/ the machine's own.
One record stands outside the zone: the working root's AGENTS.md
carries the machine's marked block (its discovery hook), kept whole
by every run; bytes outside the block are never touched.
