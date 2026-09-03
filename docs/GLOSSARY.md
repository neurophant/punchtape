# punchtape GLOSSARY

A dictionary of the system's internal jargon — full coverage: every
word or expression that appears in the machine's operation, output,
and documents and might be unclear. Principle: if the user MIGHT not
understand it, they will not, and it must be here. Every term — a
definition in plain language plus an example or an analogy; → marks
related terms; a Roman numeral marks a section of docs/ARCHITECTURE.md.
Freshness rule: a new system concept appears only together
with its glossary line.

**.punchtape** — the hidden instance directory: state, canon,
passport, renders, journal, ledger, cache. VII

**.yamll** — a file that is a stream of YAML documents (journal,
ledger): records are append-only, one per entry. VII

A

**Acceptance (acceptance point)** — a ledger record with the digest of
the surface and of the definitions of all checks, taken when every
card is green/quarantined; further drift downgrades the verdict. VI

**AGENTS.md (self-doc)** — the only machine file in the root of the
working directory: the executor's full input contract; written by the
machine.

**Agnosticism (total and clean, toward everything)** — the machine
knows and assumes nothing external: neither product, nor stack, nor
environment — no daemons, no protocols, no ports, no vocabularies of
environment operations. Observables are argv/stdout/stderr/exit
code/files; operations are running commands and moving bytes.
Everything concrete is instance data and human decisions. Analogy: a
deaf notary — records only what is seen, without guessing what the
client came FOR. → surface, seed, materials. II, VIII

**Amend (deferred amendment)** — a change to proven expectations held
for human approval: shown verbatim with a plan; approve applies it,
reject lifts it, silence is the conservative default. A blind approval
is not an approval. VI

**amend** — see Amend (deferred amendment).

**Anti-entropy effect** — the machine's measurable action against
uncertainty of meaning: locally, entropy can only be lowered by work
(a heat pump: the machine pumps uncertainty out of the delivery loop
and dissipates what was paid into watts — tokens, double runs); the
bill is not for erasing bits (Landauer) but for search and
measurement; part of the entropy is not erased but moves into an
explicit auditable form (question → default → decision log); order is
stored up (a battery of order). NOT the system's name, but its effect.
Analogy: a pump bailing water out of a hold — the ship does not become
"anti-water", it just stays afloat. → battery of order, guarantee paid
for in watts, external transactional state machine. I

**Artifact** — what results from building by the recipe: a file or a
whole tree, which the machine moves into the clean run directory as
is. → surface, conventions, build recipe. VI, VII

**Assertion** — a typed expectation of an observation: {observation,
condition, value}: stdout equals "…", exit-code fails, file
json-equals…. VI

**Attempt budget** — how many times IN A ROW a card may survive a
barren repair cycle (three) before going to quarantine. A cycle is
barren when the submission changed files, the suite really ran, and
the red line that was red before the submission stayed red; progress
(it went green, or another line took its place) resets the counter.
It counts behavior, not environment failures and not typos: a run over
an unreliable surface is marked "not run" and eats no budget. Analogy:
three defective parts in a row — the crew is pulled off; a defect
every other part — each one gets investigated. → card, quarantine,
env-fail. VI

**Attention (attention wall)** — the count of human time: questions,
choices, silent defaults, and idle minutes; measured by the ledger
(question → choice events). → question batch, ledger. VI

**Autopilot (driver)** — a way of working where the machine itself
turns the loop "slot → external executor → submission", without a
human behind every iteration. The human still answers questions. →
slot, executor, fanout. V

B

**backfill** — closing the gaps of the past: observations without a
scenario are proposed as checks-table rows from what was actually
observed. VI

**Battery of order** — canon plus a green verdict as order in store:
regression re-runs are paid for in advance, fixes are free in the
sense of checking. The first delivery is expensive (pumping out
uncertainty), the next ones nearly free (battery discharge). Analogy:
a charged battery — it pulls not because it is burning fuel right now,
but because the charge is already paid for. → anti-entropy effect,
verdict. I

**bg-suite** — a background rehearsal: a separate process runs the
whole suite over the surface digest while the hand prepares the next
submission; the live run matters more. VI

**Bill** — a verdict line: wall, verb invocations, tokens (or an
honest "no data"). VI

**Black box** — the principle: the machine sees the product only
through its surface (argv, streams, exit code, files); the internals
are not its level. II

**Boundary** — a requirement as a unit of trust in the passport: it
carries a trust-ladder status and run verdicts. Analogy: a boundary
stone: while the land is untouched it stands; touch it — it gets
measured out again. → passport, trust ladder. VI

**Brief ambiguity** — an unresolved choice in the requirements (a
file name, a value's shape, a default…): by the campaign rule it is
closed in the brief BEFORE launch; a question from a run = a defect
of the brief.

**Build recipe** — the conventions' commands that turn the project
into an artifact; one for the whole surface; timeout 300 s. VI

**Byte capture by the machine** — fixing an expectation by intent:
the executor says what changes and why, while the new bytes are
measured by the machine running the delivered product; the change
waits for human approval with "was → measured". Nobody types bytes by
hand. Analogy: calibrating scales with a weight, not by eye. → amend.
VI

**Bytes-equals** — comparing files by exact bytes, without
normalization: trailing newlines are part of the contract. → semantic
comparisons. VI

C

**Cache** — a rebuildable service zone `.punchtape/cache`: slots,
verdicts, digests, rehearsal reports; not a source of truth, safe to
delete. VII

**Canon** — the single source of truth about an instance:
requirements, scenarios, checks, cards, contracts; everything else is
a computable view of the canon. Analogy: a score against a
performance: it can be played many times, written once. VII

**Canon data** — the texts of all surfaces and the numeric thresholds
embedded as data into the binary (not into the code); editing a text
means releasing a new version. → delivery binary. V

**Canon schema (schema-version)** — the data-format version of an
instance: a binary older than the data means an honest refusal;
instances are not versioned by machine releases. VIII

**Card (CRD)** — an implementation assignment: a facet of the product,
the scenarios it closes, and exclusive ownership of files; attempts,
status, blockers. Analogy: a work order to a crew: what to do and
which walls to own. → split, contract. VI

**change request** — a request to change what is proven: a plan of the
effect (cards, contracts, stage) waits for human approval. VI

**Check (TST)** — an executable scenario expanded by the compiler:
seed → backgrounds → setup → action → assertions; written only by the
compiler, never by hand. VI

**Checks table** — the executor's interface to the spec: rows of
"seed + backgrounds + commands + expectations of the last one"; the
machine compiles it into the canon itself. The main authoring path.
VI

**Choice point (code 3)** — the state when the machine needs a human
answer: a question batch or a held amendment; the loop is not
blocked, silence means defaults. → escalation. IV

**CLI/IDE/API (envelopes)** — the forms of invoking the executor: a
terminal session, a pair of files, strict JSON; the envelope is
translated, the slot never is.

**Closed loop** — the machine's cybernetic meaning: wish → check →
double run → verdict — the chain does not break, the result returns
into the work. Vibe coding is an open loop: write and pray. Analogy:
heating with a thermostat versus a fireplace — the thermostat senses
and corrects, the fireplace just burns. → anti-entropy effect,
external transactional state machine. I

**Compatibility (compatibility matrix)** — what has been measured
about environments (GOOS/GOARCH); the unmeasured is honestly "not
measured", silence is forbidden.

**Compositeness (machine question)** — the machine's question when a
checks table drives two or more different products (different
command[0]): split into tasks or keep as one; silence means the
recommendation "one product". VI

**Composition freeze (COMPOSITION FREEZE)** — the first code
submission fixes the composition of the spec: from then on, new rows
(growth) and replacement of existing rows (repair) are lawful, while
changing proven rows takes human approval. Analogy: an order goes to
production — the blueprint is sealed; blueprint edits only by change
request. → change request, proven. VI

**Contract (IFC)** — a frozen boundary between facets of parallel
work: a shared observable asset (a state file, a convention) in
surface terms; the version only grows. VI

**Conventions** — the instance's knowledge about the product's
environment: the build recipe (commands + artifact), optional static
checks (types, lint) with {out}/{name} placeholders ({name} is ONE
name: the alphabetically first of the rows' command[0]; a shell
variable ${...} is not a placeholder — recipe commands run without a
shell). Decided by the operator. → surface.

**Conventions era** — the window since the last recipe change: the
rent of the static gates (types/lint) is counted anew in a new era. →
gate rent.

**CONVENTIONS NEEDED** — a block of the first implement slot when the
surfaces cannot be resolved: names under which nothing executes, and
the choice of the surface form (teach a build recipe — the
recommendation — or submit executable files at the named paths). VI

**Converge (convergence)** — a verdict line about the convergence of
the repair cycle: how many fix cycles, the worst card by attempts,
quarantines. VI

**Cost gate** — the economics gate: ceremony overhead tokens against
task tokens, a 10% ceiling; without a meter it has no right to go
red.

**Coverage** — completeness gaps: families of the wish's commands
without an executable scenario; the gate names them one by one. VI

**Coverage probe** — a check of a check: a green row must go red on a
broken surface (a stub over a really existing surface entry; PATH
names are not substituted); if it did not go red, it checks nothing
(pin-red). Analogy: a test of the test: a sensor that does not notice
an unplugged instrument is not a sensor. VI

D

**Decision log (DEC-)** — the operator's accumulated decisions with
rationale: "we decided so, because". → knowledge. VI

**Default (recommended)** — the answer option the machine applies on
human silence; always marked in the question; a late explicit answer
upgrades the silent choice to an explicit one. → question batch,
silence. VI

**Delivery binary** — a single punchtape executable; inside it — both
the code and all the canon-data texts; the version is stamped in at
build. → canon data, version. VII

**Delta** — one submission from the agent: YAML of one kind (wish,
checks table, code, fix…), possibly with a token tail. Analogy: one
step-letter in the correspondence with the machine. VI

**Delta skeleton** — the canonical sample of the correct shape of a
delta kind; attached to a refusal on a form error: the machine teaches
the format by construction. VI

**Derivatives (derive)** — the auto-expanded suite: from each scenario
the machine deterministically derives variants (empty state, restart,
zero/minus/neighbors of numbers, removing the seed key; budget 8) and
checks robustness.

**Digest** — a short fingerprint of content (SHA-256; the
human-readable short form is 12 characters, the verdict signature is
16): same bytes — same digest, different bytes — almost certainly a
different digest. The machine reconciles the world by digests, not by
bytes. Analogy: the fingerprint of a file. VII

**Double-compile** — the release gate: a fresh twin instance receives
the same spec from the canon and a second executor; the sets of
scenarios must match, and all checks must go green in both. Analogy:
rebuilding a house from the same blueprints with another crew — the
result must be the same. VI

**Double run (transience)** — every check runs twice in the same run
directory path, wiped clean between attempts; observations must match
byte for byte; randomness is caught, not accepted. → volatility.

**Draft (spec draft)** — a cheap spec from an auxiliary executor: it
becomes the spec only after explicit acceptance by the hand.

**Drift** — an instance's divergence from the acceptance point (the
surface files changed after the verdict). A drift claim needs two
signatures — the one recorded at acceptance and a measurable current
one, both computed through the same normalization; an acceptance
that could not measure a signature is named as such, never compared
against. Drift downgrades the verdict to NOT-READY and is named
honestly. → acceptance, verdict, freshness contour. VI

E

**env-fail** — the red of an environment/launch failure (tool missing,
launch timeout): not behavior, eats no attempt budget. → environment
diagnostics. VI

**Environment diagnostics (env doctor, envdiag)** — a deterministic
classification of environment failures: no tool, no permission, a
timeout, no artifact — each with a hint. A red environment reddens
acceptance. → env-fail, acceptance. VI

**Escalation (code 4)** — an open conflict of requirements or a
quarantine: only an explicit, live human choice; silence never
resolves it. IV

**Executor** — the hand's LLM agent: reads the slot, writes deltas,
drives the task to the verdict; one executor per task. II

**Executor environment (task world)** — the environment the executor
works in: the stack, the tools, the stack's linters in PATH (the
campaign rule). VI

**External state machine** — see external transactional state machine
(the second level of the name).

**External transactional state machine** — punchtape's category: the
machine stands OUTSIDE the agent's application and owns the task's
state for good (canon, journal, digests, replay); the agent is a
peripheral and can only propose transactions (submissions), which are
accepted whole or refused without a change of state. Four levels of
the name: external transactional state machine → external state
machine → state machine → the machine. The inversion against
state-machine libraries (LangGraph and kin): there the machine is a
part INSIDE the agent's application and dies with its session. The
chiasmus slogan: "Not a state machine inside the agent's app — the
agent's app inside the state machine." Analogy: a
bank versus a wallet — the bank owns the account state and processes
transactions, the wallet is only a carrier. → inversion of ownership,
transaction, canon. I, II

F

**Fanout** — parallel waves of the autopilot: active and red cards go
to their executors at the same time; applying deltas stays
sequential — there is one writer. → autopilot. V

**fanout** — see Fanout.

**Feature** — a requirement in the human sense: what the product can
do; the name is the phrasing, the behavior is the scenarios, the
description is the human spec's prose. VI

**Feature prose (kind: feature)** — the author's description of a
feature in living words: what it does, how it is invoked, what is
visible, boundaries and errors; a new text waits for ratification
anew. An empty description holds the verdict at NOT-READY: the human
spec is part of "done", not a decoration. → waive. VI

**Feature statuses** — the living words of requirements in the human
spec: draft, accepted, delivered, quarantined, cancelled. VI

**FILES WRITTEN** — the response block of an accepted code/fix
submission: the paths and digests of the files written; the source of
digest references for the following submissions (digests are copied
verbatim, not recomputed by hand). VI

**Format auto-repair** — deterministic repair of purely typographical
submission errors (a tab in an indent, an unclosed quote, a doubled
token tail, a typo in a known field, up to two edits). It never
repairs semantics. Analogy: autocorrect in an editor that does not
change the meaning of words. → submission, delta skeleton. VI

**fragment (why fragment)** — a scenario's observations as a ready
YAML block for surgical match edits: the canon hands out its own bytes
as data.

**Freshness contour** — reconciling the passport with reality before
work: if files or conventions have moved away from the verdicts — the
slot carries a CONTOUR line. Analogy: an "expired" stamp on a
certificate. VI

G

**Gap class** — the classification of a quarantine gap: canon — a
contradiction between spec rows; environment — all the reds are the
environment; code — everything else. VI

**Gate** — the machine's checkpoint for a submission: statics
(types/lint), a suite run, a coverage probe, trace, review-diff,
cost, latency (informational), double-compile. A red gate is not an
opinion, but a named reason. → gate rent. VI

**Gate rent** — retiring the types/lint/coverage-probe gates after
three runs without a single red in the current conventions era:
trust without rechecking until the world changes; a recipe change
returns types/lint to duty (a new conventions era). The stop-gates
(suite, trace, review-diff) cannot be rented. → gate, conventions
era. VI

**Guarantee paid for in watts** — the formula of the machine's
economics: cheap measurable heat (the tokens of double runs, metered
runs) buys expensive unmeasurable uncertainty (freedom to interpret
the brief). The field confirmed it: 50/50 machine verdicts against
22% strict byte convergence for a solo run. → anti-entropy effect.
I

H

**handoff** — a computable handover of a session: the instance's
facts and a ready continuation prompt; nothing is stored. VI

**Healer phrases** — the machine's ready answers to predictable agent
failures (for example, a repeated submission key). VI

**Hold** — the state of an open deferred amendment: a submission that
changes proven expectations waits for human approval (kind: amend);
silence preserves the proven expectations. → amend. VI

**Hypotheses** — the registry of hypotheses about changes to the
delivery canon: change, hypothesis, metric, window, baseline, outcome
(open/confirmed/refuted). Visible via `why hypotheses`. VI

I

**Idempotency (deja-vu defense)** — the same submission (by key or by
content digest) is not applied twice: the answer IDEMPOTENT REPLAY,
the counters do not grow. → submission. VI, VIII

**impact** — the reverse index "file → boundaries": what an edit will
touch, before the edit; a slot line when the spec changes. VI

**Instance** — one task of the machine: a directory with the product,
the AGENTS.md self-doc, and the hidden state `.punchtape`. It lives
with its own binary. II, VII

**Instance changelog** — `.punchtape/CHANGELOG.md`: the diff of the
task's life from the journal — what changed and why; written by the
machine at every verdict. VII

**Intent router** — a translator of everyday phrases into verbs: "what
next" → next, "price" → why price; not a single match — the phrase is
treated as a product wish; two different matches — a menu for the
human, the machine does not resolve an ambiguity. VI

**Inversion of ownership** — the essence of the difference from
state-machine libraries: there, the agent's application owns the
machine (inside, it dies with the session); here the machine stands
outside, it owns the state, and the agent is a peripheral. "External"
= "the owner". → external transactional state machine. I, II

J

**Journal (transactions)** — an append-only stream of the applied
deltas with their effects (what was written/deleted, and where); all
state is always derivable from it; files damaged by hand are restored
from it. Analogy: a flight recorder — the fact is written first, the
world second. → ledger, manifest. VII

K

**kb (knowledge base)** — "when X, do Y" records with a source and
the replacement of stale ones; deterministic matching of triggers
against the context.

**Knowledge** — the accumulated wisdom of an instance: the decision
log, the out-of-scope list, the kb knowledge base. Visible via `why
knowledge`. VI

L

**Latency** — the response time of the machine's verbs (p95); an
informational gate: the duration of honest check work is a property
of the task, not a machine defect; a failure means sustained
non-adjacent spikes.

**Ledger** — the journal of the machine's meter facts: events, wall,
invocations, tokens ("null" = honestly not measured). Analogy: a
counter at a machine tool — it does not do the work itself, but
without it you will not see where the downtime is. → journal. VII

**Ledger events** — the kinds of facts: verb, apply, gate, executor,
question, choice, answer-upgrade, waive, quarantine, gate-retire,
probe, contour. VII

M

**Machine changelog** — see Instance changelog (the same render,
`.punchtape/CHANGELOG.md`).

**Machine-write marker** — the first line of render and data files:
"written by the machine, hand edits are rejected"; a file without
the marker is human, untouchable. VII

**Manifest (lastgood)** — the list of files with digests after the
last successful write; the reconciliation at opening catches manual
edits.

**Manual edits** — changing the machine's files by hand: detected by
reconciliation against the manifest, refused with an honest refusal;
everything goes through submissions. VIII

**Materials** — the project's files moved into the run directory byte
for byte under the same paths: binary inputs are not relayed by hand.
→ seed. VI

N

**next** — the verb that reads the current slot; the default after the
first read is the contextual diff (the ACTION line always in full),
`next --full` prints the whole slot; in an empty directory it creates
an instance and prints the self-doc. IV

**NEXT SLOT:** — the prefix of an accepted submission's response: the
next slot is already in the response, a separate next is not needed.
→ slot, submission. VI

O

**Out-of-scope** — the list of what the product deliberately does not
do, with reasons; instance knowledge. → knowledge. VI

P

**Pain journal** — the mechanical journal of troubles: refused
submissions, repeated keys, red cycles, quarantines, approval needs;
each pain gets a diagnosis and a workaround from the catalog. Analogy:
a machine-tool failure log — the process is fixed by it, instead of
being surprised anew every time. → retrospective. VI

**Passport** — the accumulated knowledge about the product's
boundaries: trust-ladder statuses, run verdicts, conventions,
inventory. A computable view of the canon. → boundary. VI

**PATH launch** — a command[0] name not found among the surface
entries: the machine hands it to the environment as is, and the
environment searches the path. OS level, zero product knowledge. →
surface entry. VI

**probe** — formalizing observed invocations into spec rows at
release: the spec catches up with code that has stabilized. VI

**Prose** — the spec's unmeasurable remainder: what matters in words,
but has no scenario; lives in the canon, visible in the spec and the
verdict, not compiled into checks. VI

**Proven** — "has gone green at least once": protection of
expectations from post-code weakening; a row that has proven nothing
is repaired freely. VI

**PT_SUBMITTER** — a voluntary self-declaration of the calling side
via the environment; logged "as reported", without trust. VI

Q

**Quarantine** — an honest stop of a card after three behaviorally
red attempts, with the gap classified; it reaches the verdict, and
the verdict asks the human. Analogy: a defect on the warehouse shelf —
not thrown away, but set out for review. → attempt budget.

**Question batch** — a package of clarifications for the human (up to
three); one answer closes all of them; silence applies the
recommended defaults; answered questions go into the echo of choices
(CLARIFIED) and are not asked again; the structured input shows "what
I understood" (the wish verbatim) and "what I can do". Analogy: a
doctor asks the list of questions before the appointment, not one at
a time in the doorway. → choice point, default, attention. VI

R

**Ratification** — explicit approval by a human (the operator):
feature prose, executor decisions; for feature prose, a new text
cancels the previous ratification. → feature, spec. VI

**READY / QUARANTINE / NOT-READY** — the statuses of the verdict. VI

**Red list (RED LIST)** — the slot's red checks, folded into runs by
identical cause; the static blockers as a separate GATE REDS list. VI

**Refreeze** — the journal head has changed (a change request, a
double-compile): the verdict is rendered anew, and the knowledge
renders are regenerated in the same step. → composition freeze,
verdict. VI

**Regeneration** — the action proposed by the slot on a deep red (more
than half the card's scenarios, or the second attempt): rewrite the
facet from scratch instead of patching it. VI

**Rerun (RERUN)** — a code/fix submission whose content matched the
project entirely (not a single new byte): nothing is applied, the
counters do not move, the gates run synchronously. The honest way to
recheck after the environment has been restored or the recipe has
changed — before, hands disguised a rerun as kind: fix. Analogy:
rechecking a finished part by re-measuring it, not by putting it into
the work again. → submission, digest, gate. VI

**Retrospective (retro)** — a computable breakdown of pains by
frequency, with workarounds and a call to write the cure into
knowledge. VI

**review-diff** — the honesty gate of review: claimed-but-not-done and
done-but-not-claimed. VI

**Row identity (row key)** — seed + materials + setup + the sequence
of commands: matched byte for byte — the same table row, the
expectations are replaced; otherwise a new row. VI

**Row twin** — an addition to the checks table whose form (seed +
commands) matched an existing row while the set of materials is
different: NOT a replacement. The submission's answer names such twin
rows explicitly — there are no silent duplicates in the canon. → row
identity, checks table. VI

**run-lock** — serialization of the executors of one instance's runs:
a single scenario port. VI

**Run snapshot** — the map "path → bytes" of all the regular files of
the run directory after the action: a full reconciliation of state,
with a reading budget.

S

**Scenario skeleton** — the degree of a scenario's immaturity: not
everything needed for a run is there (no invocation, no assertions).
Degrees: executable / skeleton / prose. → prose. VI

**Seed** — files with content that the machine brings into the clean
run directory before the row's commands: a transfer of bytes, the
meaning unknown to the machine. → materials, checks table row. VI

**Semantic comparisons (json-equals / json-contains)** — comparing
JSON by structure: key order and whitespace do not matter, array
order is part of the state. VI

**Silence** — the absence of a human answer to a question: the
recommended default is applied; a conflict is never resolved by
silence. → default, escalation. VI

**Slot** — one step of work: stage, state, exactly one action
(ACTION), a ready example (EXAMPLE), a context slice (CONTEXT), gates
(GATES ON SUBMIT). Cached; every read after the first serves the
contextual diff — unchanged blocks collapse under digest anchors,
the ACTION line never collapses; the whole slot is the first read or
`next --full`. III, IV

**Spec** — the product's specification as data: requirements,
scenarios, checks; the source is the canon, the views are the checks
table and the human specs. → spec registry, canon. VI

**Spec registry** — the table of contents of the human specs: features,
files, description statuses, scenarios; search via `why spec
<words>`. VI

**SPEC.md / SUMMARY.md** — the machine's renders of the instance's
knowledge in `.punchtape`: the whole canon spec, and a one-page
verdict summary (Made/Checked/Remaining/Price) in living words.
VII

**Split** — dividing the product into facets of parallel work: cards
with exclusive ownership of files and contracts on the shared assets.
Only before the first code; a split proposal in the slot is PROPOSED
CUT (a big scope — BIG SCOPE — splits itself). VI

**Stages** — the steps of a task's life: intake → spec → ac-compile →
implement → review → deliver; the stages and the transitions
themselves are delivery data, the runner executes them. IV

**State machine** — see external transactional state machine (the
third level of the name).

**Submission (submit)** — delivering a delta to the machine: as a
file or from stdin (`-`); `--check` is a dry validation without
applying; `--try` is a dry trial of ONE checks-table row (the same
runner in a clean directory, without entering the canon or changing
the stage); accepted — the response carries NEXT SLOT; refused — one
line of reason. VI

**submit** — see Submission.

**suite** — the full set of an instance's checks. VI

**Surface** — the way of invoking the product that the operator has
declared: an artifact by the build recipe, or files by the names of
the commands. The machine runs behavior only through it; the slot
describes it with a SURFACE CONTRACT block. Analogy: a tourist sees
the city from a sightseeing bus, not from the basements. II, VI

**Surface entry** — a file the machine materializes in the run
directory for a row's command[0]: the project file named by command[0]
(without a recipe), the declared artifact file, or a file inside its
tree. Resolution order: a ready file of the run directory (a seed, a
material, the output of an early command) → a surface entry →
otherwise PATH, by the environment. Analogy: your own brought tool, a
tool lying on the table, a tool from the issued box — and only then
the workshop's shared cabinet. → surface, artifact, PATH launch. VI

T

**Tail** — see token tail.

**Taste queue** — the verdict line with prose (the spec's unmeasurable
remainder) and the active rule waivers: "does not block and does not
stay silent". → prose, waive. VI

**Token tail (# TOKENS)** — the last line of an executor's
submission/response: an honest counter of its own environment ("n/a"
— no environment); reserved by the ledger as the price of the
invocation. The envelope cuts only the tail line — the delta's bytes
before it (including the final newlines of the last block scalar) are
untouchable. VI

**Trace** — the scenario ↔ check link via an identifier inside the
check itself; the gate catches lost links (MISSING LINKS). VI

**Transaction** — an atomic proposal of a state change: an agent's
submission is either accepted whole (the gates run, the effects are
written to the journal, the canon is updated) or refused with one line
of reason without a change of state; the state is always derivable
from the journal. → submission, journal, external transactional state
machine. II, VIII

**Trust ladder** — the statuses of a boundary: claimed → verified (a
green run) → valid → reliable; above "verified" there are no machine
proofs, the rest is human. VI

V

**Validation** — checking a submission at the entrance: form, fields,
semantics. A refusal is one line of reason; on a form error the
machine attaches the canonical skeleton of the correct shape. → delta
skeleton. VI

**Verbs** — the agent's working surface of the machine: `next` (the
current slot), `submit` (a submission), `why` (an explanation). The
human writes no commands — they speak wishes, and their agent carries
them to the machine. III

**Verdict (VERDICT)** — the machine's final output on a task's
readiness: READY / QUARANTINE / NOT-READY, with a full account
(scenarios, checks, bill, acceptance, gates) and a digest; the prose
count in the scenarios line is the number of requirements carrying
an authored description (the human spec), not the count of prose
scenarios — those live in the taste queue. Analogy:
accepting a car by checklist, not by "seems fine". → acceptance,
gate, taste queue. IV, VI

**verifications (verification journal)** — what and when verified each
boundary: verdicts, digests, anchor transactions. VI

**Version (semver)** — the number of the machine's delivery, vX.Y.Z:
major — breaking format changes, minor — new properties, patch —
fixes. Visible via the `punchtape version` command. → machine
changelog. V

**Volatility (volatile)** — an honest declaration of
unpredictability: which observations of the brief are declared random
(identifiers, times) — thread names or file paths; the machine checks
behavior but does not reconcile these bytes between runs. The file's
presence is always checked. Analogy: the order number on a receipt
changes, but the receipt always prints. → double run. VI

W

**Waive (rule waiver)** — a conscious refusal of a completeness rule
with a mandatory reason; the rule stops blocking but stays visible in
the verdict. Only completeness can be waived — not feasibility, not
stop-conditions. Examples of rules: `coverage.family` (a family
without a scenario), `feature.doc` (a feature without an author's
description — it holds the verdict itself; the waiver is submitted at
delivery too). VI

**Whole-product card** — the starting card for all the non-prose
scenarios, where implementation begins; the split replaces it with
facets. VI

**why** — the explanation verb: a topic, or the card of an entity;
every number of the machine has an origin; why price breaks the price
of the current slot down by bytes and tokens. VI

**Wish (intent)** — the human's want, verbatim, in living language;
the source a task originates from; one line in the spec registry. IV,
VI

**Wish language** — the language of the instance's intent, detected
by the deterministic script detector (≥30% Han → zh, else ≥30%
Cyrillic → ru, else en), recorded on the instance; it steers the
human blocks (questions, summaries, specs, the handoff) — the machine
protocol itself is English. II

**Write zone** — the rule of who writes files, and how: canon and
passport — only by transactions under the manifest; journal and
ledger — append-only; cache — rebuilt. The hand writes nowhere. VIII

**Writer monopoly** — one write lock is taken for a whole submission:
read, apply, commit — a single transaction; a second writer gets an
honest refusal after short retries. VIII
