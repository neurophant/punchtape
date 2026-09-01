# PROBLEMS — the open problem registry

This file is the machine's problem registry: everything a field run
has surfaced and that is NOT yet fixed, each with its field evidence,
its cost, and a recommended fix. One rule governs every
recommendation: a fix must be a universal property of the machine's
own formats and surfaces — knowledge of any particular task, stack or
donor codebase is a defect of the fix, not an option. Dates and run
identifiers stay in `METRICS.md`; here problems are named by
mechanism, not by occasion.

Structure: [A](#a-machine-defects-open) machine defects (behavior
wrong or misleading), [B](#b-design-tensions) design tensions
(behavior right, friction real — decisions, not bugs),
[C](#c-world-image) world image (environment facts),
[D](#d-executor-guidance) executor guidance (pains that are not
machine defects), [E](#e-closed-since-the-last-registry) recently
closed (so the reader knows what NOT to re-report).

## A. Machine defects (open)

### OP-1 — red rows hide the measured bytes

**Symptom.** A red check's reason truncates the actual output to its
first line; `why fragment <SCN-N>` prints the pinned expectations of
a scenario but never the measured actuals of the last run. To see
what the product actually printed, an executor runs a probe
submission (or runs the command by hand inside the world).

**Evidence.** Every conduit cell of the full battery run and the
conduit control wave — the tasks with the richest red/repair cycles.
Cost: one extra probe submission per non-trivial red; on network
tasks with servers, the probe itself is expensive (a full row rerun
with the double determinism pass).

**Fix (systemic, agnostic).** The failing observation already passes
through the machine with its measured value in hand — the red reason
carries `want X, got Y` for streams but only a byte count for files,
and `why fragment` re-renders only the canon side. Render-only
changes, no product knowledge anywhere:
1. every red reason carries a bounded excerpt of the measured bytes
   for the failed observation (the same excerpt budget the submit
   replies already use for stdout/stderr — streams today, extend to
   file observations);
2. `why fragment` gains an `actual:` section for each assertion that
   is red right now, from the last captured run of that check (the
   capture exists — the transience comparison holds both runs' bytes
   in memory at verdict time; keep the last capture in the run
   evidence zone).

**Verification.** Force a red in a scratch instance; the reason line
and `why fragment` both show the measured bytes; then a wave — the
probe-submission count on red-bearing cells must drop.

### OP-2 — the "submissions accepted" counter can disagree with the hand

**Symptom.** The machine's accepted-submissions number was observed
one higher than the executor's own count (10 vs 9, one cell of the
full battery run).

**Root-cause hypothesis.** The counter counts accepted `submit` verbs;
the hand counts authored delta files. The difference is most plausibly
an idempotent replay (the same key applied once, counted as a verb)
or an internally generated submission — the two bookkeeping surfaces
disagree on WHAT a submission is.

**Fix (systemic, agnostic).** Make the counter's definition explicit
and deduplicated: count distinct submission keys in the journal
(applied-once is the machine's own idempotency invariant), render the
definition next to the number ("accepted submissions: N — distinct
keys, replays not counted"). A number whose definition is printed
cannot silently disagree.

**Verification.** A scratch instance: submit, replay the same bytes
(keyed), submit again — the counter moves exactly twice while three
verbs were accepted.

### OP-3 — the gates headline can disagree with the slot's red list

**Symptom.** A submit reply ends `gates: suite=green …` while the
next slot carries a red card list (observed in one conduit cell).

**Root-cause hypothesis.** The headline summarizes THIS submission's
gate run; the slot's red list shows the standing instance state —
reds of cards this submission did not touch. Both are true; the
reader is not told the scopes differ.

**Fix (systemic, agnostic).** Scope-label both surfaces: the gates
line says "(this submission)" and adds the standing count when
non-zero (`suite=green, 2 standing reds untouched by this
submission`); the slot's red-list header already names its scope —
align the two wordings so a reader cannot conflate them.

**Verification.** A two-card instance with one card red: submit a
clean card — the reply says green AND names the standing red; the
slot list matches.

### OP-4 — YAML diagnostics: one mislabel, one ordering defect

**Symptom.** (a) A block scalar whose first line is blank is
diagnosed as "a long quoted scalar broken by a real newline" — the
wrong class, the advice does not apply. (b) Auto-repair notes
(`fixed: closed quote…`) precede the actual decode error by ~150
lines when the repaired text is long — the cause is buried under its
own repairs.

**Fix (systemic, agnostic).** (a) The block-scalar detector gets the
leading-blank-line signature (a scalar that opens with `|`/`>` and an
empty first line) or stays silent when unsure — a wrong hint costs
more than none. (b) Repair notes render AFTER the error line (the
rejection already carries first-line discipline; notes append, never
prepend).

**Verification.** Scratch instances: submit both malformations; the
first names its class correctly or nothing; the second shows the
error line first.

### OP-5 — submit latency spikes on large single-submission repairs

**Symptom.** submit wall of 0.7–3.8 s when one submission carries a
25-row table reformat. Class: INFO — the latency budget was never
exceeded; p95 across the latest control wave is 110 ms.

**Fix direction.** The decode-driven repair loop re-decodes the whole
delta per attempt (bounded attempts). If waves ever push this class
into budget, decode only the failing fragment. Today: monitor, do not
touch.

### OP-6 — render cosmetics: tombstones and key numbering

**Symptom.** (a) A removed scenario leaves an `id: ""` tombstone line
in the context listing. (b) Suggested submission-keys skip numbers
(the suggestion derives from journal length; replays and gate entries
make it non-monotonic). Both ×3 cells of the full run, zero behavior
impact.

**Fix (systemic, agnostic).** (a) Filter empty identities in the
context render. (b) Derive the suggestion from a dedicated monotonic
counter of suggested keys (persisted in the instance state), not from
journal length. Render/state-only.

## B. Design tensions

Decisions, not defects: the behavior is as designed, the friction is
real. Each entry names the tension and a recommended resolution that
stays inside the agnosticism rule.

### OP-7 — post-delivery formulation edits are all held

The change-request hold treats every requirement edit after code
delivery as a human assertion. The teaching now says to re-cut BEFORE
the first code submission — but a wording-only edit that changes no
expectations, no scenarios and no boundaries is still held. **Open
decision:** release such edits freely. The check is structural (diff
of the requirement's scenario set and expectations = empty), fully
agnostic; the cost is one more trust-ladder exception, and that is
the owner's call, not the machine's.

### OP-8 — `kind: split` at spec is decorative

The machine auto-drafts a whole-product card; a split at stage spec
is not applied, yet the surface stays taught. Either the split
becomes submittable at spec (the composition gate learns to accept a
pre-code regrouping), or the taught surface at spec stops offering
it. Consistency between the taught surface and the stage gates is the
systemic criterion.

### OP-9 — the exit-3 slot does not say how to advance

A question batch returns exit code 3; the slot shows the batch but
not the way forward (answers ride the next checks delta as
`answers:`; silence applies the defaults). Executors burned probe
submissions discovering this. Fix: one clause in the batch header —
canon text, no mechanism change.

### OP-10 — row repair identity is opt-in

A row carrying an authorial `key:` replaces its never-green prior
versions even with different seeds and run — orphan-free repair; the
latest control wave shows the field adopting it (keyed scenarios in
every cell, error-path rows named `file-missing`, `not-jpeg`,
`threshold-range` and the like). Executors that do not give keys
still add new scenarios and orphan the broken originals; removal
stays a separate free channel. Keep the identity opt-in (keys are
authorial intent, the machine must not guess heirs); the remaining
gap is teaching: the red-repair action text is the right place for
"keep the key while you iterate" — one canon clause.

### OP-11 — determinism is strict over the whole run directory

The transience comparison covers every file of the run dir, asserted
or not. This is the documented rule, and it caught real divergence
classes the asserted observations would have missed (ephemeral ports
in logs, logger noise, random ids in error texts). The friction:
network tasks redden until volatility is declared. **Recommended
resolution:** keep the strict rule; make the difference class visible
in the reason — a file that is NOT an asserted observation gets named
as such ("file X — not an asserted observation — differs between the
runs; declare it volatile or move it out of the run dir"), so the
declaration is one read away. Structural (asserted vs not), agnostic.

### OP-12 — derived variants assume rc 0

Empty-state and restart reruns of a scenario that intentionally fails
(the `fails` assertion shape) go red on the exit code. Informational,
non-blocking. Fix direction: derived variant generation respects the
assertion shape of its source scenario (no rc-0 assumption where the
scenario itself asserts refusal) — a property of the machine's own
derivation formats.

### OP-13 — probe-red guidance points at the wrong layer

A coverage-probe red (the check stayed green on a sabotaged surface)
can render with slot guidance aimed at expectations, sending the
executor to repair a table that is not the problem. Fix: route
probe-red guidance to the pin semantics (a pin-red check proves
nothing; repairing it is free) — canon/render text, the outcome
classes already exist.

### OP-14 — `{name}` inference meets orchestration rows

The recipe placeholder `{name}` resolves to the alphabetically first
`command[0]` of the rows; when orchestration rows lead with a shell,
`{name}` becomes the shell's name. Documented, still surprising. The
agnostic resolution is authorial override, not smarter inference: an
explicit `name:` in the conventions recipe wins over the alphabetical
default (a declaration beats a heuristic — and the machine learns
nothing about which words are shells). Decision for the owner.

## C. World image

Environment facts of the campaign layer, not machine behavior; they
shape cell economies and are fixed in staging:

- **OP-15** md4c ships the SAX API only (`md4c-html.h` declares
  `md_html`, nothing exports it) — the C++ cells hand-wrote an HTML
  renderer. Fix: build and link md4c-html into the layer.
- **OP-16** no `gm` CLI despite the GraphicsMagick++ dev package —
  the C++ image cell generated fixtures with a scratch program. Fix:
  add the CLI to the layer.
- **OP-17** no rustfmt in the toolchain (clippy only). Fix: add the
  component.
- **OP-18** go-redis v9 logs timestamped dial-retry lines to stderr —
  byte-nondeterminism until silenced by the product. Not a machine
  defect (the volatile channel is the designed answer); a guidance
  line in the stack notes saves a red cycle.
- **OP-19** the Go module cache advertises go-redis only under the
  `@v` layout — a staging note, no cell was lost to it.

The `/tmp` node-module resolution problem of the earlier runs is
closed by the world JS layer (`/run/node_modules` linked into `/tmp`)
and needs no action.

## D. Executor guidance

Pains observed in the field that are NOT machine defects — they
belong in the contract and the self-doc, and mostly already are:

- heredocs through `docker exec bash -lc` mangle nested quotes —
  write files host-side into the bind mount;
- background helpers must close their inherited stdout/stderr
  (`>/dev/null 2>&1`) or rows time out on the held pipe;
- YAML escape semantics: `\n` inside single-quoted seeds stays
  literal, double quotes decode it; `\xNN` byte escapes are
  generated, never hand-pasted;
- flow-map values containing `,` or `": "` go in double quotes (a
  torn value is auto-repaired once — quoting is the reliable form);
- stack traps seen once each: exception-order subtleties in image
  libraries, exclusive rectangle bounds, RRGGBBAA hex order, lenient
  parsers, and quire briefs not naming their template variables (a
  brief-level improvement).

## E. Closed since the last registry

So the reader does not re-report what is already fixed:

- the verdict's prose count now reflects described features (it used
  to stay at zero after feature texts were accepted);
- an acceptance that could not measure a surface signature says so —
  no more empty digests in the acceptance line; a drift claim needs
  two signatures, computed through one normalization — the false
  `SURFACE DRIFTED` after prose-only submissions is gone (zero
  occurrences in the latest control wave);
- the ACTION line of a slot never collapses under an anchor; a fresh
  render serves the contextual diff (the full print is the first read
  or an explicit `next --full`);
- a flow value torn by an unquoted comma is auto-repaired once, and
  the quoting and seeds-placement rules are taught with anti-examples;
- a seed value wrapped in single quotes is named as content bytes in
  the reply instead of silently reddening;
- a never-green row replaces its prior versions under an authorial
  key, across seeds and run changes; a key held by a proven row keeps
  the change-request channel (no bypass);
- the re-cut is taught for the window before the first code
  submission, with the post-delivery hold named;
- a killed process group is waited for its actual death before the
  second determinism probe of a row;
- the entity reference names the verb (`why`; there is no `card`
  verb).

## Maintenance

Every fix of an OP item follows the same discipline: the fix is a
universal property of the machine's own surfaces (agnostics first,
or it does not ship); `make build` + vet + staticcheck clean; a live
smoke in three wish languages; a field check that the evidence class
of the problem disappeared. Items close by moving their text to
section E with the fix semantics — one registry, no history files.
