# punchtape

*an external, transactional state machine — the agent works for it.*

**Not a state machine inside the agent's app — the agent's app inside
the state machine.**

You speak in plain human words — the agent works — the machine
guarantees: the wish becomes a verifiable spec, the spec becomes
acceptance checks, code is accepted only through gates, and at the end
the machine prints a verdict with a digest: `READY`, an honest
`NOT-READY` with reasons, or a `QUARANTINE` with a named cause. The model remains a generator; the machine is
what guarantees.

## What this is

punchtape is an **external transactional state machine**: one binary
that stands next to your project and owns the task's state — the
requirements, the checks, the decision journal. Not a plugin and not a
library inside the agent: an arbiter on the outside. The agent does not
"try to please" — it submits its work to the machine, and the machine
itself runs every check twice in clean directories and signs the
result.

The state outlives sessions. Models come and go — what has been punched
into the machine stays.

## Why you need it

Everyone who has worked with AI agents knows four pains:

- **"The AI wrote code" ≠ "done".** The agent always says "done".
  Verifying it for real is your job — for hours.
- **Checks on someone's word.** "I ran it, it works" — one run, on a
  cache, with garbage left from previous attempts.
- **Edits are scary.** Something was changed — and nobody knows what
  broke, because nobody re-verifies everything.
- **Knowledge dies in the chat.** Decisions, agreements, "why it is
  this way" — all of it dissolves together with the session.

punchtape closes all four — not with promises, with mechanics.

## What you get

- **A verdict, not an opinion.** Readiness is computed by the machine:
  every scenario is tied to a check, every check has run twice, a probe
catches empty checks. In the field — **75 of 75 cells** across two
waves (five task families × five stacks × three hand classes)
passed the blind double-run acceptance, and **20 of 20** machine
  cells in the release validation wave closed at `VERDICT: READY` —
  against **22%** strict convergence of solo agents without the
  machine (same model, same tasks).
- **Edits without fear.** Any change automatically re-runs the whole
  checks table again — the price of re-verification is ≈ zero. The
  first delivery is expensive; every next one is nearly free.
- **Ambiguity surfaces before code.** Every machine question is a place
  where the wish was silent. In the field they are rare — a question
  batch fired in 4 of 25 cells at the top model (about one per six
  tasks) and in 5 of 20 at the cheap one (closed in 16–59 s); in the
  release validation an operator decision was needed in 7 of 20
  cells and cost $0.60 in total (the input-rate upper estimate).
- **Code is expendable.** The product can be regenerated from the
  spec by another executor in another language: the double-compilation
  gate opens a twin instance from the same canon spec and requires
  both implementations' checks green — a designed capability of the
  machine, not yet covered by the field runs (see
  `docs/SCENARIOS.md`).
- **No infinite loops.** Repair has an attempt budget; exhausted — an
  honest quarantine with a reason, not a silent "almost done".
- **The price of everything.** A slot shows the token estimate before
  execution, `why` decomposes any number; the machine's own overhead is
  itself a gate: with a token meter, above 10% — red; without a
  meter the gate has no right to go red.
- **Knowledge as files, not chat.** After the verdict the machine
  writes human-readable specs: one file per feature, in living words,
  plus a registry with search. Documentation stops being debt.
- **Quality does not depend on the model's price.** Measured: the
  cheap model inside the machine delivers the same verdict — 20/20
  in the field wave, 10/10 in the release validation — for 1.6–2×
  the time and 3.7–7.6× less money. Expensive does not mean
  reliable; the machine does.

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
notes done <number> closes a note; state lives in notes.json and
survives a restart.
```

From there the machine and the agent do everything themselves: the
first call opens an instance and drops the self-documentation next to
it (agents read it without you), every step carries one action with a
ready example, and the `VERDICT` line is the stop. Between your wish
and the verdict there are zero mandatory points of your attention
(measured in the field).

**Step 3 — edit the same way, in words:**

```text
Add an open command — show only the unclosed notes.
```

The machine reopens the task, rebuilds, and re-verifies everything —
the old and the new alike.

**Questions are rare and convenient.** When the machine needs your
decision, the question arrives as a list with each option's price and a
recommendation. Answer in a couple of words — or stay silent: the
recommended defaults apply, and every decision is recorded in the
journal with tracing.

**What it looks like inside** (the wish → the agent working the machine
→ the verdict; this is the output the agent sees — you see the finished
product and the verdict):

```text
punchtape next "Make a notes utility: add, list, done…"
ACCEPTED: transaction 943c19aa8f85
NEXT SLOT: STAGE: spec …  ACTION: submit a checks delta: one row per
behavior — the machine compiles canon, checks and tracing from the
table itself

punchtape submit checks.yaml          # the agent submits the checks table
ACCEPTED: gates: suite=red coverage-probe=green trace=green
NEXT SLOT: STAGE: implement …  ACTION: submit a code delta for card CRD-001

punchtape submit fix.yaml             # repair of a red row
ACCEPTED: gates: suite=green … review-diff=green
VERDICT: NOT-READY                    # prose gate: no feature description
# the loop stops at a verdict line and shows it; the standing
# instruction "finish to READY" starts it again at the named action:

punchtape submit feature.yaml         # the feature description, in living words
ACCEPTED: transaction ed5a550c6013
VERDICT: READY
digest: b682767ce515b044
scenarios: total 4, green 4, red 0
features: 2 total, 2 carry the author's prose
traceability: 4/4 executable scenarios linked to green checks
checks: green 4, red 0, total 4
derived: 8 variants (max 8 per scenario), green 2, red 6 — marked, not spec scenarios
attention: 0 min (no questions: goal lint clean)
converge: fix cycles 1, worst card attempts 0 of 3, quarantines 0
acceptance: PASS on 779f45c45988 (4/4 checks green)
```

The machine works right in your project directory and creates its
service directory `.punchtape/` there — the instance it owns:
canon, journal, ledger; the executor agents stay out of it.

## Economics — measured, not promised

Five task families × five stacks, each task done three ways: a top
model through the machine, a cheap model through the machine, and a
solo agent without the machine; a blind measurer ran every product's
suite twice in clean directories, and an AI judge compared the
products cross-hand. The original study — `docs/RESEARCH.md`.

- A typical task, full path (the hand plus its verifying operator):
  **solo ≈ $6.07, top model + machine ≈ $1.18, cheap model +
  machine ≈ $0.32** — the same blind-measured quality; the top
  path runs roughly 15–20 minutes, the cheap one about one and a
  half to two times longer.
- A batch of 25 such tasks: ≈ $29 at the top model, ≈ $8 at the
  cheap one — against ≈ $152 for the solo path, and that solo price
  already includes the checking operator.
- The machine path also holds the line where solo does not: in the
  release validation wave every machine cell closed at
  `VERDICT: READY` (top $1.30, cheap $0.17 per cell), and the
  live three-language smoke (ru/en/zh) reached READY with zero
  question batches at $0.66 the whole smoke.
- There is no free option: the price is either in the machine (the
  tokens of double runs) or in invisible defects and your own time.

A side effect is **anti-entropy**: the uncertainty of meaning grows on
its own unless it is pumped out. The machine is a heat pump: cheap
measurable heat (tokens) buys expensive unmeasurable uncertainty. The
unavoidable forks are not erased — they move into a signed form:
question → default → journal.

## Honest boundaries

- **The measured envelope.** All numbers were obtained on the S band —
  compact utilities with file state. The machine is agnostic about the
  product's shape: it sees only commands and their observable output,
  and through commands more is verifiable too (a web server — by
  running requests at it, long-running processes — by live stands);
  web, GUI, and mobile have not been field-tested yet — that is a
  question of run methodology, not a limitation of the machine.
- **Measurable acceptance.** The machine cannot check "make it
  beautiful" — it honestly shelves that into a "taste queue"; your
  call.
- **An agent is required.** The machine does not write code itself — it
  works with your agent in any environment.
- **Stacks.** Go, Python, JavaScript/TypeScript, Rust, C/C++;
  third-party libraries are allowed, but they resolve only from the
  world image's local cache — builds run without network.
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
cannot be unpunched. So here: every transaction is punched
into the journal once and for all, the canon it builds survives every
change of model, and replay reads only the journal.

## Where to dig

- `docs/RESEARCH.md` — the research paper: hypotheses, field
  experiments, metrics, conclusions.
- `docs/ARCHITECTURE.md` — the full anatomy: the meaning, the concept,
  every feature, the formats, the integrity rules.
- `docs/GLOSSARY.md` — the dictionary of every term.
- `docs/SCENARIOS.md` — what the field has verified, and what not yet.
- `docs/SAMPLES.md` — the field task bank and the difficulty
  calibration.
- `campaigns/` — how the field measurements were staged.
