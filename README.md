# punchtape

*an external, transactional state machine — the agent works for it.*

**Not a state machine inside the agent's app — the agent's app inside
the state machine.**

You speak the brief — the agent writes the code **and one spec
document** — the machine stores the spec (behaviors: id, name, prose,
check rows) and verifies: it re-runs the rows that went stale in the
live working tree — a digest read before and after must name the same
surface — and signs the result with a digest: a `READY` verdict, or
an honest `NOT-READY` with the named reason and the measured values.
The model remains a generator; the machine is what guarantees.

## What this is

punchtape is an **external transactional state machine**: one binary
that stands next to your project and owns the task's state — the
behavior corpus, the journal, the verdicts. Not a plugin and not a
library inside the agent: an arbiter on the outside. The agent does
not "try to please" — it submits its work to the machine, and the
machine runs every check itself, in the live working tree under
digest control, and signs the result.

The interface is deliberately tiny. One door —
`.punchtape/inbox/current.txt` — is the only way in: the agent
overwrites that one
file and submits it. One document grammar, two kinds: `spec`
(validated and stored, nothing runs) and `run` (executes the stale
rows; `ratify: true` signs the recorded green). The boundary is
strict: `.punchtape/` is the machine's zone — the instance;
everything else in the working root is yours. The machine never
writes project code and never prescribes structure, layout or files:
at the implement stage the agent codes by hand, directly in the
working directory — not one byte of that code passes through the
machine — and the machine answers by running the rows.

The state outlives sessions. Models come and go — what has been punched
into the machine stays.

## Why you need it

Everyone who has worked with AI agents knows four pains:

- **"The AI wrote code" ≠ "done".** The agent always says "done".
  Verifying it for real is your job — for hours.
- **Checks on someone's word.** "I ran it, it works" — one run, on
  a cache, with garbage left from previous attempts.
- **Edits are scary.** Something was changed — and nobody knows what
  broke, because nobody re-verifies everything.
- **Knowledge dies in the chat.** Decisions, agreements, "why is it
  this way" — all of it dissolves together with the session.

punchtape closes all four — not with promises, with mechanics.

## What you get

- **A verdict, not an opinion.** Readiness is computed by the machine:
  every behavior owns executable check rows, a run submission executes
  the stale ones in the live tree under digest control, and the
  verdict carries the
  digest of the exact tree it was pronounced on. `NOT-READY` is a
  named red line with the measured facts — never a silent "almost
  done".
- **The code is yours — all of it.** Zero bytes of code pass through
  the machine (the journal's effects name instance files only — a
  product file appears in no record and no effect). The
  corpus plus the green rows can regenerate the product in another
  language by another executor: the spec is the source of truth, the
  green rows are the proof.
- **Edits without fear.** The machine watches the tree by digest: a
  change after the acceptance moment is caught and named, and a bare
  re-run submission re-verifies the changed tree. The first delivery is
  expensive; every next one is nearly free.
- **Expectations are data, not vibes.** Every check row is a
  mechanical expectation — exact text (one trailing newline
  tolerated), a fragment, an exit code, a file, a byte sample — and
  every behavior carries the author's prose. Where the brief was
  silent, the spec shows the silence.
- **The machine's checks ARE the test suite.** Write no private suite:
  the corpus is the product's memory of its own behavior.
- **Knowledge as files, not chat.** The corpus and the journal are
  the project's memory; after the verdict the machine renders a
  human-readable summary inside the instance — `SUMMARY.md` (what
  was made and what remains) — and a fresh `spec.md` render after
  every accepted submission.

The field numbers behind these claims — 75 of 75 field cells passing
blind double-run acceptance against 22% strict convergence of solo
agents, and the price measurements below — were measured on the
study's earlier delivery loop (the machine applied the code deltas
itself). The interface has since moved to the one door and the
one-document grammar of behaviors; its own field battery — five
waves, two hands, the study's protocol — is recorded in
`docs/RESEARCH.md`, and its live acceptance on every build is the
round-trip battery: full cycles of the real binary to
`VERDICT: READY`, the door the only input.

## How to use it

Everything is driven by prompts in your own agent. You do not need a
console.

**Step 1 — once per machine.** Paste this to your agent:

```text
Install punchtape and run all work through it:

  curl -fsSL https://github.com/neurophant/punchtape/releases/latest/download/get-punchtape.sh | bash
```

The agent will install the machine and drive it from then on without
you.

From source (needs Go 1.26+), the same single prompt:

```text
Install punchtape from source and run all work through it:

  git clone https://github.com/neurophant/punchtape && cd punchtape && ./install.sh
```

**Step 2 — write tasks in plain text.** For example:

```text
Make a notes utility: notes add <text> saves a note and prints its
number; notes list prints the notes one per line, oldest first;
notes show <number> prints one note; state lives in notes.yaml and
survives a restart.
```

From there the machine and the agent do everything themselves: the
first call births the instance and injects the working contract into
the reply (agents receive it without you), every slot names the one
door file and exactly one action, and the `VERDICT` line is the
stop. Between your
brief and the verdict there are zero mandatory points of your
attention.

**Step 3 — edit the same way, in words:**

```text
Add an open command — show only the unclosed notes.
```

An edit is the next spec version of the same living case: the change
lands through the same door as a sparse patch, the work reopens, and
only what moved re-verifies — a row whose definition and inputs did
not change keeps its recorded green. A genuinely new task is a new
brief in a fresh directory.

**Repairs are ordinary submissions.** A red row's reason carries the
measured values — the exit code, the expectation and the measured
stream, a window of bytes around the divergence, and, for a trailing
newline kind of mismatch, a copy-ready expectation proposal. Fix the
code, or replace the row's expectation by its address (`kind: run`,
`patch: replace:`); a bare `kind: run` re-runs whatever went stale.
Every decision lands in the journal.

**What it looks like inside** (the shape of a live session of this
tree's binary — the brief → the agent working the machine through
the door → the verdict; this is the output the agent sees — you
see the finished product and the verdict):

```text
punchtape next "a notes tool: add a note, list notes, show one note"
(the first call births the instance, injects the AGENTS.md working
contract into the reply, pins the brief, answers with the slot)

STAGE: intake
BRIEF: a notes tool: add a note, list notes, show one note
INBOX: .punchtape/inbox/current.txt — the one submission file;
overwrite it, then submit the same path
CONTRACT: AGENTS.md in your tree (unchanged; sha256 5f2a09c41d7eb3c2)
ACTION: write the spec — one document, kind: spec, into the inbox file
below (the form is prefilled; the full grammar:
.punchtape/reference/grammar.md)
CONTEXT:
the spec is your own document: behaviors with prose and check rows;
the machine validates the form without running anything and answers
with the next slot
GATES ON SUBMIT: suite

# the agent overwrites the door with the whole spec — behaviors
# add/list/show, prose on each, one row each — and submits that path

punchtape submit .punchtape/inbox/current.txt
ACCEPTED: transaction 57d8168c0943
BYTES: sha256 00b13e7a2c5d…
GATES: form
DOOR: reset to the blank template — overwrite current.txt with your
next submission before submitting again (a template or empty file is
refused, not read)
SPEC v1: atoms 3 (+3 new, ~0 changed, -0 retired); rows 3 (+3 new,
~0 changed, -0 retired)

# the slot rides the next call — the cache makes it free

punchtape next
STAGE: implement
INBOX: .punchtape/inbox/current.txt — the one submission file;
overwrite it, then submit the same path
CONTRACT: AGENTS.md in your tree (unchanged; sha256 5f2a09c41d7eb3c2)
STATE: atoms 3, rows 3, open 3
CHANGED (3) — rows whose standing moved since the last slot (the
full table: punchtape report --full):
  note-add/1: never (was new)
  note-list/1: never (was new)
  note-show/1: never (was new)
the rows have not run yet — the machine waits for the code first
(write it in the work tree, then kind: run)
ACTION: write the code for note-add in the work tree, then submit
kind: run (row edits ride the same document as a patch, by row
address)
CONTEXT:
ATOM note-add — add a note
  …the behavior's prose…
rows (1): green 0, red 0, never 1
GATES ON SUBMIT: suite, tree

# the agent writes notes (a python script) directly in the working
# directory — structure, layout and files are its own choice, no
# submission carries code — then submits the bare run

punchtape submit .punchtape/inbox/current.txt               # kind: run
ACCEPTED: transaction 7a67adfceaa1
BYTES: sha256 …
GATES: suite, tree
DOOR: …
suite=red (3 rows, 2 red)
  note-add/1: stdout: not equal to "1: first note"; measured: …
  note-show/1: …

# repairs: fix the code, or replace the row's expectation by its
# address (kind: run, patch: replace: note-add/1); after each
# acceptance the door is back to the blank template

punchtape submit .punchtape/inbox/current.txt               # kind: run
ACCEPTED: transaction 998ba52b31f8
BYTES: sha256 …
GATES: suite, tree
DOOR: …
suite=green (3 rows)

punchtape submit .punchtape/inbox/current.txt   # kind: run, ratify: true
ACCEPTED: transaction b92f516bfdc2
BYTES: sha256 …
GATES: freshness, digest, verdict
DOOR: …
ratified: 3 rows kept green at unchanged inputs, tree digest
b95e0dc6cb5a — the signature stands on the recorded green, nothing
re-ran
VERDICT: READY (v1)

punchtape next
STAGE: deliver
INBOX: .punchtape/inbox/current.txt — …
CONTRACT: AGENTS.md in your tree (unchanged; sha256 5f2a09c41d7eb3c2)
ACTION: the verdict above is the accepted state's snapshot (v1); the
case lives on — the next spec version enters through the same door
CONTEXT:
VERDICT: READY
digest: 90dc70da33f37874
version: v1
atoms: 3 total, 3 green; rows: 3
stages: 1 spec submissions, 3 runs
origin: a notes tool: add a note, list notes, show one note
bill: machine wall 704ms, 7 verb calls
acceptance: PASS on tree b95e0dc6cb5a (3/3 rows green)
GATES ON SUBMIT: —
```

The machine works right in your project directory and creates its
service directory `.punchtape/` there — the instance it owns: the
corpus, the checkpoints, the frozen verdicts, the journal and the
ledger; your agent works beside it, never inside it.

## Economics — measured, not promised

The measurements below come from the earlier field study (five task
families × five stacks, each task three ways: a top model through the
machine, a cheap model through the machine, a solo agent without it;
a blind measurer, an AI judge — the full method in `docs/RESEARCH.md`).
The machine's economics argument stands — it replaces your
verification time with metered machine runs — and the current
one-document interface carries its own five-wave field battery in the
same paper: with the verifying operator counted in, the machine path
was the cheaper one in 23 of 25 tasks.

- A typical task, full path (the agent plus its operator):
  **solo ≈ $6.07, top model + machine ≈ $1.18, cheap model +
  machine ≈ $0.32** — the same blind-measured quality; the top path
  runs roughly 15–20 minutes, the cheap one about one and a half to
  two times longer.
- A batch of 25 such tasks: ≈ $29 at the top model, ≈ $8 at the cheap
  one — against ≈ $152 for the solo path, and that solo price already
  includes the checking operator.
- The machine path also holds the line where solo does not: in the
  release validation wave every machine cell closed at
  `VERDICT: READY` (top $1.30, cheap $0.17 per cell), and the live
  three-language smoke (ru/en/zh) reached READY with zero question
  batches at $0.66 the whole smoke.
- There is no free option: the price is either in the machine (its
  measured checks) or in invisible defects and your own time.

A side effect is **anti-entropy**: the uncertainty of meaning grows on
its own unless it is pumped out. The machine is a heat pump: cheap
measurable heat (tokens) buys expensive unmeasurable uncertainty. The
unavoidable forks are not erased — they move into a signed form:
row → measured red → journal.

## Honest boundaries

- **The measured envelope.** The field numbers were obtained on the S
  band — compact utilities with file state. The machine is agnostic
  about the product's shape: it sees only commands and their observable
  output, and through commands more is verifiable too (a web server —
  by running requests at it, long-running processes — by live stands
  inside a row); web, GUI, and mobile have not been field-tested yet.
- **Measurable acceptance.** The machine cannot check "make it
  beautiful" — it requires the author's prose on every behavior and
  checks presence, not quality; the reading is yours.
- **An agent is required.** The machine never writes code itself — the
  code is the executor's territory; it works with your agent in any
  environment.
- **Stacks.** Anything a command can build and run: the rows invoke
  your product through `command[0]` (a project file or the
  environment's PATH), and the spec's `build` names the build.
- **Platforms.** Linux and macOS; Windows — later.
- **For one-off scratches the machine is overkill** — take solo and do
  not pay for a guarantee you do not need. Anything that lives longer
  than a day will be edited — and the economics flip in the machine's
  favor from the first such event.

## Why "punchtape"

The name is from punched tape: the paper strip with punched holes, the
first program that outlived both the operator and the machine. The tape
outlives the machine: models come and go — the state remains. Punched
tape is humanity's earliest state machine: a hole is punched — it
cannot be unpunched. So here: every transaction is punched into the
journal once and for all, the corpus it builds survives every change of
model, and replay reads only the journal.

## Where to dig

- `docs/RESEARCH.md` — the research paper: hypotheses, field
  experiments, metrics, conclusions.
- `docs/ARCHITECTURE.md` — the full anatomy: the meaning, the concept,
  every feature, the formats, the integrity rules.
- `docs/GLOSSARY.md` — the dictionary of every term.
- `campaigns/` — the verification harness (the gates, the selftests,
  the staged-inputs freeze).
