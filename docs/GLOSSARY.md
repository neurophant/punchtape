# punchtape GLOSSARY

A dictionary of the system's internal jargon — full coverage: every
word or expression that appears in the machine's operation, output,
and documents and might be unclear. Principle: if the user MIGHT not
understand it, they will not, and it must be here. Every term — a
definition in plain language plus an example or an analogy; → marks
related terms; `reference/<topic>.md` names the machine's own
shipped doc on the subject (the same texts the instance regenerates
into `.punchtape/reference/`). Freshness rule: a new system concept
appears only together with its glossary line.

The machine's own vocabulary is small and closed: the document kinds
(spec, run), the behavior ids and row labels (`B1`, `B1/2`), the
reply lines (STAGE, INBOX, CONTRACT, CHANGED, ACTION, VERDICT,
RED LIST). Everything below explains that vocabulary and the world
around it. What died with the model change died everywhere: the
passport, the trace channel, the
amendments-and-questions machinery, the numbered inbox sets and the
submission-key ritual are gone from the machine and from this
dictionary.

**.punchtape** — the hidden instance directory: the machine's zone
(canon atoms, checkpoints, verdict snapshots, the door, the service
zone — state, journal, ledger, locks, cache, renders). Never edited
by hand. → instance, write zone.

**.yamll** — a file that is a stream of YAML document frames
(journal, ledger): records are append-only, one `---`-separated
document per entry.

A

**Acceptance (acceptance moment)** — the transaction in which every
behavior of the corpus stands green on this tree: it lands in the
ledger as a gate record (suite, scope acceptance) with the
green/total counts and the pre-run tree digest, and the verdict's
acceptance line names it —
`acceptance: PASS on tree <digest> (N/N rows green)`. One moment
per tree digest: a full-fresh ratification signs the tree the last
green run already accepted. A tree that moves after it is drift.
→ checkpoint, drift, pure re-run, verdict.

**Agent** — who does the work: the LLM that reads the slot,
authors the spec document, writes the code directly in the work
tree, and drives the task to the verdict. The machine never writes
code — the agent writes it, byte for byte, outside the machine. →
executor, work tree.

**AGENTS.md (the working contract)** — the only machine file in the
root of the working directory: the agent's full input contract,
written by the machine at birth and injected in full into the first
next's reply; a root without it has no instance yet. The contract
law: the birth records the contract's own sha256, every later slot
carries the one-line reference on it, and the full contract renders
again the moment the tree's copy is gone or changed. If the
repository carried its own AGENTS.md, the file is COMBINED — the
machine contract first and always prevailing, the repository's
rules preserved verbatim below the tail boundary.
→ birth, contract line, renders.

**Agnosticism (total, toward everything)** — the machine knows and
assumes nothing external: neither product, nor stack, nor
environment — and nothing about tasks, domains or roles. The
observables are argv/stdout/stderr/exit code/files; the operations
are running commands and moving bytes. The project's shape is the
agent's business. Analogy: a deaf notary — records only what is seen,
without guessing what the client came FOR. → black box, live tree.

**Answers** — the `answers:` field of the submission document:
parsed and honestly refused — no question batch can be open in this
machine generation, and an answer without a question is noise, not
silence. Kept in the grammar so an old habit fails loudly, not
silently. → document.

**Anti-entropy effect** — the machine's measurable action against
uncertainty of meaning: locally, entropy can only be lowered by work
(a heat pump: the machine pumps uncertainty out of the delivery loop
and dissipates what was paid into watts — tokens, measured runs); the
bill is not for erasing bits (Landauer) but for search and
measurement; part of the entropy is not erased but moves into an
explicit auditable form (row expectations, measured reds, the
journal); order is stored up (a battery of order). NOT the system's
name, but its effect. → battery of order.

**Atom (behavior)** — one unit of behavior of the corpus: the agent's
short id, a living name, the mandatory prose, and the check rows
with the machine's record of their last run (outcome, reason,
env-fail, the freshness memo) — one canon file per behavior,
carrying its own record digest. The atom is the unit of the
verdict: green when all its rows are green.
`reference/grammar.md`. → corpus, freshness, prose, row.

B

**Backfill (the round-trip of the corpus)** — bringing an EXISTING
tree under the machine: birth by next in the existing directory, the
spec authored against the code that already exists, then a run
executes the rows against the ORIGINAL tree — green means the spec
matches the code; a mismatch is a backfill finding (a red row with
the measured values: the row is repaired or the code is). No
separate verb exists; the round-trip is the same single run the
machine always does. Analogy: surveying an old house by measuring
it, not by asking the tenants. → acceptance, checkpoint.

**Battery of order** — the corpus plus a green verdict as order in
store: regression re-runs are paid for in advance, edits are free in
the sense of checking (a pure re-run costs one submission). The
first delivery is expensive (pumping out uncertainty), the next ones
nearly free (battery discharge). → anti-entropy effect, verdict.

**Bill** — a verdict line: machine wall, verb calls. → verdict.

**Birth** — creating the instance: `punchtape next "<brief>"` in a
fresh directory — writes AGENTS.md beside the instance, injects the
full contract into the reply's head, and fixes the brief and its
language in the state in the same transaction (the brief is NOT a
submission). A repeated birth phrase is an honest refusal: the brief
is not edited through rebirth. → brief, instance, next.

**Black box** — the principle: the machine sees the product only
through its commands (argv, streams, exit code, files); the
internals are not its level. It never sees the code at all — not one
byte of it passes through the machine. → agnosticism.

**Brief** — the human's want, verbatim, in living language; fixed
once at the instance's birth by `next "<brief>"` and stored in the
state (the state field is still named `wish`); the origin line of
the verdict and the summary quote it. Not a submission: the birth
phrase cannot be re-edited, a repeated birth is an honest refusal.
→ birth, language.

**Build** — the document's optional build declaration (argv): runs
once in the live tree before the rows of every run submission; rc
must be 0 within 300 s. A present declaration replaces the previous
one; `[]` clears it. `reference/environment.md`. → build-ignore,
run (the kind).

**Build-ignore** — the document's ignore globs: paths of the work
tree outside the accounted surface (build outputs, caches). A
pattern ending in `/` excludes a directory prefix; a plain pattern
matches the relative path or any file name (the gitignore habit).
`.punchtape` is excluded unconditionally. The remainder is the
product the digest signs. → digested surface, tree moved.

**Bytes sample** — the `bytes: file://NAME` reference of a row: a
sample file lying next to the door in `.punchtape/inbox/`; stdout
must equal its bytes exactly — the only binary expectation form.
→ door, row.

C

**Canon** — the single source of truth about an instance: the
behavior atoms. Everything else (spec.md, the summary, the slot) is
a computable view of the canon. Analogy: a score against a
performance: it can be played many times, written once.

**Canon data** — the texts of all surfaces and the numeric thresholds
embedded as data into the binary (not into the code); editing a text
means releasing a new version. → delivery binary.

**Canon schema (schema-version)** — the data-format version of an
instance: a binary older than the data means an honest refusal;
instances are not versioned by machine releases. The current schema
(9) is the wholly procedural row grammar — the witness probe staged
through the door, the launch/tree modes composing with every head;
a schema-8 instance honestly refuses to open.

**Ceremony meter** — the reply bytes the hand paid, summed by
render family over the ledger's verb entries: contract (the
full-contract replies — the birth injection and the anti-tamper
re-render), spec (the intake slots), report (the work-stage slots),
door-echo (the submit acks). Deterministic arithmetic over recorded
facts; printed in SUMMARY.md's price block. → ledger, thin ack.

**Changed list (CHANGED)** — the implement slot's list of the rows
whose standing moved since the LAST slot, retirements included — the
door law: only what moved, never the whole corpus (the full table is
the report's `--full`). The delta mode is the same list alone.
→ slot, surface modes.

**Checkpoint** — a digest snapshot of the digested surface: the file
set plus a per-file sha256, one digest over sorted `path\0hash`
lines, written as `checkpoints/ckp-<seq>.yaml` in the SAME
transaction as the run's outcomes — the tree the outcomes stand on
is fixed with them. → acceptance, digested surface, tree moved.

**Contract line** — the slot's CONTRACT line: the one-line reference
on the birth sha while the tree's AGENTS.md digests to the birth
record; the FULL contract bytes again the moment the copy is gone or
changed (anti-tamper, not an option). → AGENTS.md.

**Corpus** — the behavior atoms of an instance, the product's
specification as data: green when every atom is green, the verdict's
ground. → atom, canon.

D

**Delivery binary** — the compiled machine with its canon data
embedded (`go:embed`): one point of truth for the texts — the file
in the source tree, in the binary, in the instance. → canon data.

**Diagnosis line** — the one-line refusal of a rejected submission:
it names the entity (`behaviors[1]`, `B1/2`, `delta`, `patch`) and
the reason; the draft stays in the door for the rework. The form is
taught by the door's template and `reference/grammar.md`. → door,
form errors are loud.

**Digest** — a short fingerprint of content (SHA-256; the short form
in replies is 12 characters, the verdict signature is 16): same
bytes — same digest, different bytes — almost certainly a different
digest. The machine reconciles the world by digests, not by bytes;
digest values exist only in machine replies — never compute one
yourself. Analogy: the fingerprint of a file.

**Digested surface** — the accounted product: the work tree minus
`.punchtape/` minus the build-ignore globs. The sibling `input/`
directory sits outside the work dir and is never digested. The tree
digest of the verdict covers exactly this. → build-ignore,
checkpoint.

**Door** — THE submission file: `.punchtape/inbox/current.txt`, the
only path submit accepts (any other path is refused with the wanted
one named — the door is impermeable). The machine prefills it from
the embedded template when missing; after an accepted, non-replayed
submission it is rewritten back to the blank template; a rejection
leaves the draft. Byte samples and witness probes lie next to it.
→ document, submit, template catch.

**Document (the submission document)** — the one YAML-shaped text
the agent writes into the door: two kinds (spec, run), one grammar —
the FIRST spec is the full table (behaviors with prose and rows, an
optional build declaration, optional ignore globs); every later
submission is a patch by row address (the full table never rides
twice; the deliberate full rewrite carries `full-replace: true`).
Two spellings of one content are one content (the machine keys by
the content digest). `reference/grammar.md`.
→ door, kind, patch.

**Drift** — the tree moved after the acceptance moment: the verdict
re-snapshots the CURRENT tree (an observation, not a journal record)
and D' ≠ D is the machine-visible red line — the acceptance line
names it, a READY is downgraded to NOT-READY, and re-verification
needs a run on the current tree (the pure re-run driver exists
exactly for this). A drift claim needs TWO signatures: an acceptance
that could not see the surface or an unmeasurable one says nothing
— a false drift does not reach the verdict. → acceptance,
pure re-run, verdict.

E

**env-fail** — the red of an environment or launch failure (a tool
missing, a start failure, a launch timeout, a broken build):
classified by the diagnostics registry and marked not-behavior. A
red row is a red row nonetheless — the release is not proven until
the substrate holds. → environment diagnostics.

**Environment diagnostics (env doctor)** — a deterministic
classification of environment failures: each class renders a
diagnosis plus a FIX hint from the registry (canon data). →
env-fail.

**Executor** — the agent: an older spelling of the same role, kept
by the machine's code comments; both spellings name the one who does
the work. → agent.

F

**Form (the form law)** — the one-line source of a row's formal
choices (`form:`): a row that pins formal surface — a nonzero rc, or
an exact (non-contains) out/err/state text — without a citation is
a FORM DECISION UNRESOLVED: the spec lands, the ack names the rows,
and the run door stays closed until each pin carries its source line
(the brief/prereg phrase it quotes, or the operator's one-line
decision as `form: "op(<date>): <decision>"`). The machine never
judges the citation's content — presence is the whole mechanism.
→ row.

**Form errors are loud** — the machine never auto-repairs a
submission's form: a decode failure or a validation refusal is one
diagnosis line; the draft stays in the door. The only surviving
hint is the kind head (an unknown kind one or two edits from exactly
one known kind gets it named). → diagnosis line.

**Freshness (the content-addressed memo)** — a green row's standing
at unchanged inputs: the row's recorded memo (`ran-at`, `ran-tree`,
`inputs`, `ran-rc`, `ran-out`) keys its result to the digests of the
files its command names in the work tree — plus its from-input
seeds' sources, its byte sample, its witness probe and the declared
build. Nothing moved — the row does not re-run (a kept green); a
named input moved — the row is stale and re-runs; a red row always
re-runs. The standings vocabulary: green fresh / green stale / red /
never. → row, suite.

**from-input** — the seed source form: a copy of a file of the
instance's input directory (the sibling `input/` of the work root)
into the work tree, file-to-file, for the row's duration; the
file's existence is checked at intake. → seed.

**from-run** — the state-expectation form that compares a file after
the row byte-for-byte with a file the run itself produced — for
expectations that ARE run artifacts. → row, state check.

**full-replace (the deliberate rewrite hatch)** — the explicit mark
that a FULL behaviors document is a conscious act, not an accident:
the door refuses an implicit full re-submission after the first
accepted spec (the table never rides twice — the economy law); with
`full-replace: true` it is taken, the ack counts "old N rows -> new
M rows" and the event rides the ledger. Never composes with a
patch. → patch.

G

**Gate** — a named machine check of a stage transition (spec-
accepted, ratify-green, spec-resubmitted): the runner walks the
stage exits while the gates pass; a gate is a mechanic, never an
opinion. The run's own checks show as the suite line and the tree
control. → stages, suite.

I

**Idempotency (deja-vu defense)** — the same document content is not
applied twice: the answer IDEMPOTENT REPLAY with the original
transaction id, the corpus does not change, the counters do not
grow. The machine derives its own submission key from the content
digest — a spec's key also carries the corpus the document implies
(the same document against the corpus it describes is a repeat;
against a different corpus it is a new version), a patch is its own
content — the agent invents no keys. One exception by design: the
pure re-run driver. → pure re-run.

**Instance** — one task of the machine: a directory with the product,
the AGENTS.md working contract, and the hidden state `.punchtape`.
→ .punchtape.

J

**Journal (transactions)** — an append-only stream of the applied
transactions with their effects (the birth's brief transaction
included); all state is always derivable from it; files damaged by
hand are restored from it. A frame cut mid-write does not blind the
machine: the intact records load, the torn tail is quarantined to a
sidecar. The entries carry the retired-row frames and the version
numbers of the spec they created and the verdict they froze. No code
ever enters it — the agent's code lives in the working root.
Analogy: a flight recorder — the fact is written first, the world
second. → ledger, writer monopoly.

K

**Kind** — the submission document's `kind:` head, exactly two
values: spec (the state of the case; nothing runs) and run
(execution). The kind decides what the machine does with the
document, not which fields it has. → document, run (the kind), spec
(the kind).

**Knowledge render** — the machine-written human-readable files
inside the instance: `SUMMARY.md` (the one-page verdict summary, in
the brief's language) and `spec.md` (the current spec render). The
reference library — `.punchtape/reference/` — is maintained at every
open (regenerated when the canon head changed). Read them, do not
edit them. → renders.

L

**Language** — the brief's language, detected once at birth by the
deterministic script detector (≥30% Han → zh, else ≥30% Cyrillic →
ru, else en) and recorded on the instance; it steers the HUMAN
render (the summary; the prose is the agent's own language by
construction). The machine protocol itself is English.

**Ledger** — the journal of the machine's meter facts: events (verb,
apply, gate, heal), wall, calls, reply bytes with the reply's
render family (the ceremony meter's ground), the rejection reason,
PT_SUBMITTER as reported. Analogy: a counter at a machine tool — it
does not do the work itself, but without it you will not see where
the downtime is. → journal.

**Live tree** — the run world: rows run in the work tree itself,
zero copies — cwd is the work dir, PATH is the system PATH, no
wrappers and no containers. A digest is read before the run and one
after: outside the ignores they must name the same surface. A
launch's execute right is a grant for the row's duration, folded
back at its end. `reference/environment.md`. → digested surface,
seed, tree moved.

M

**Measured value** — what a red row's reason carries: the exit code
got versus wanted, the expectation and the measured stream with an
excerpt, a window of bytes around a divergence; a terminator-shaped
divergence (a trailing newline, a missing line break) additionally
carries a copy-ready expectation PROPOSAL — the measured bytes to
paste into the row when the code is the truth. → red list.

N

**next** — the verb that reads the current slot; in a fresh
directory `next "<brief>"` births the instance (writes AGENTS.md,
injects the contract, fixes the brief) and answers with the first
slot; a bare next answers with the slot alone. → birth, slot.

**NEXT SLOT** — the slot render an accepted submission's follow-up
`punchtape next` returns (the same render as a bare next; the
submission's own reply is the thin ack — transaction, bytes,
gates, the measured one-liners — and carries no slot). → slot.

P

**PATH launch** — a command[0] name with no slash that no tree file
answers: the machine hands it to the environment as is, and the
environment searches the path. OS level, zero product knowledge. A
name with a slash is a tree path; a bare tree file is launched
through its tree path. `reference/environment.md`.

**Patch (the sparse submission)** — the later document's shape: the
change only, addressed by the row labels the machine prints
(BEHAVIOR-ID/N): `replace` (new row bytes at the address, in place),
`add` (rows to the END of their table; a new id opens the behavior),
`remove` (rows retired), `run` (rows to execute now, fresh or not),
`prose` (the verdict-gate text replacement). One address — one
operation; addresses read the corpus at the patch's arrival; the
same patch bytes again are an honest replay. The machine EXPANDS
the patch into the full document a re-submission would have been —
the canon bytes of both paths are identical by construction.
→ document, full-replace, idempotency.

**Procedural modes** — the row's runner disciplines, composing with
every head: `launch: nonet` (the launch happens inside a fresh
network-less namespace — nothing is wrapped around the command, the
network simply is not there; a host that cannot raise the namespace
answers an honest environment red) and `tree: unchanged` (the
row-scoped tree identity control). Mechanics, not meaning: whatever
a mode is FOR lives in the probe or the row's own expectations.
→ tree moved, typed heads.

**Prose (behavior prose)** — the author's description of a behavior
in living words: what it is, how it is invoked, what is visible,
boundaries and errors. MANDATORY at the form level — a behavior
without prose does not validate; the human spec is part of "done",
not a decoration. → atom.

**PT_SUBMITTER** — a voluntary self-declaration of the calling side
via the environment; journaled "as reported", without trust.

**Pure re-run** — the re-verification driver: a `kind: run` document
with no behaviors, no patch and no build declaration. Its whole
content is "run again", so it never dedupes: the same bytes after a
fix of the tree are a NEW application, not a replay. Re-verification
after a repair or a drift is its purpose; an all-fresh corpus
executes nothing — every row keeps its green, the acceptance stands.
→ idempotency, drift.

R

**Ratify (ratification)** — `ratify: true`, only legal in kind: run:
the SIGNATURE, not a run — it executes nothing (no row, no build)
and signs exactly when the recorded green still stands: every row
fresh at unchanged inputs and the surface digest equal to the last
green run's. Stale — a red row, a never-run row, a moved input, a
drifted tree — is a refusal naming the rows or the paths (nothing
applied); the honest route back is a plain run. Success freezes the
verdict snapshot inside the same transaction; the case lives on —
at deliver a run is refused, the next spec version reopens the work.
→ freshness, stages, verdict snapshot.

**READY / NOT-READY** — the statuses of the verdict. NOT-READY is a
named red list with the measured facts, never a silent "almost
done". → verdict.

**Red list (RED LIST)** — the standing reds with their measured
reasons: in the focus atom's card of the implement slot and in the
NOT-READY verdict, capped by the list budget (10 lines, canon
data). → measured value.

**Renders** — the machine-written regenerable files of the instance
root and `.punchtape`: AGENTS.md, the reference library, spec.md,
SUMMARY.md — derived from the canon and the journal, never a source
of truth. → knowledge render.

**Report (the verb)** — `punchtape report`: the operator's read-only
table of the stored runs (nothing re-runs): every row's exact
command, its demand, the measured answer (the exit code and output
digest of the last run), the run's time and the freshness mark. The
default surface is the door view (per-atom counters, the red rows in
full); `--full` prints the whole table; `--since vN` aims the same
facts at the delta after a frozen verdict — the changed, new,
re-measured and retired rows by name. → freshness,
verdict snapshot.

**Retired row (retirement)** — a row removed from the corpus by a
spec version (an id the document no longer names, a shrunk row list,
a patch `remove`): recorded as a visible journal frame with its last
standing; the next verdict explains every one by name — greening by
deletion without a trace is impossible. A row dropped before the
case's first run never had a chance to measure (amnestied); a row
retired after runs without ever running is a refusal at ratify.
→ patch, verdict.

**Row (check row)** — one executable end-to-end check inside a
behavior: ONE head — the general `run` command (argv or a quoted
string) or a typed head — plus optional seeds, the expectations
(out/err/rc/state/bytes), an optional timeout, optional
procedural modes and an optional `form:` source line. Expectations
are DIRECT: an echo of a harness self-report is refused, a contains fragment
shorter than 4 runes needs its honest `(contains-loose)` mark. The
row's address in replies and renders is `<id>/<n>` (`B1/2`) — the
address the patches ride. The machine's record of the last run
(outcome, reason, env-fail) and the freshness memo live in the row,
and are never submission content. `reference/grammar.md`.
→ atom, form, seed, suite, typed heads.

**Run (the kind)** — the executing half of the two kinds: after the
apply it runs every STALE row of the corpus (the declared build
first; the kept greens named in the suite line; a run never
retires); behaviors absent = a pure re-run; `ratify: true` signs
the recorded green. → document, freshness, pure re-run, ratify.

S

**Seed** — a file the machine writes into the work tree before the
row and removes after it: `text` (inline content) or `from-input`
(a copy of a file of the instance's input directory). A seed over a
tree file honestly replaces it for the row's duration; when the row
ends the original bytes return, a fresh seed path is removed —
between rows seeds do not live. `reference/grammar.md`.

**Slot** — one step of work: the stage (STAGE), the INBOX line (the
door), the contract line, the stage body (the brief at intake; at
implement the STATE counts, the CHANGED list and — before the first
run — the await-code line; at deliver the verdict), the
SESSION HANDOFF recommendation, exactly one action (ACTION), a
context slice
(CONTEXT — the focus atom's card), the gates (GATES ON SUBMIT).
Deterministic: the same canon and journal — bit-for-bit the same
text; the cache key includes the journal head, the canon digests and
(on deliver) the live tree digest. → changed list, NEXT SLOT,
surface modes.

**Surface modes** — the slot's representation-only flags on next and
submit: `--quiet` (drops the CONTEXT body; the deliver verdict still
rides), `--delta` (only the rows whose standing moved since the LAST
SLOT — the anchor is the last `next` reply, any mode; a submit's ack
is not a slot and never re-anchors it), `--run-after` (submit: the
just-accepted spec's suite runs in the same door call — two
transactions, one reply). The journal and the canon are byte-identical
across the modes. → Slot.

**Spec (the kind)** — the storing half of the two kinds: the machine
validates the form and applies the document's behaviors AS THE WHOLE
CORPUS — an id absent from it retires its atom, nothing runs; the
FIRST spec is the one full document, every later spec rides patches.
→ document, atom, patch.

**Spec version** — the living case's version counter: every accepted
`kind: spec` document is version N; the ack answers with the counts
line `SPEC vN: atoms … (+new, ~changed, -retired); rows …` — the
machine-computed diff of the document against the corpus it arrived
at. → patch, retired row, verdict snapshot.

**spec.md** — the machine's render of the current spec, refreshed
after every accepted submission: one section per behavior — the name,
the prose verbatim, a rows table with the last outcomes, the first
standing red's reason. A regenerated render, not a journal effect. →
knowledge render.

**spec --check (dry validation)** — the free pre-submission check
of a draft: `punchtape spec --check <file>` runs the door's own form
rules over the draft against the read-only corpus and prints one
reason line per finding; a pass is the door's promise — the same
predicates, the first try. Understands patches (the expansion over
the corpus as it stands). Nothing opens, nothing writes — the check
costs nothing.

**Stages** — the steps of a task's life: intake → implement
→ deliver, and deliver → implement again (the reopen on the next
spec version); the stages and their exit gates are delivery data
(stages.yaml), the runner executes them. At deliver a run is
refused — the case lives on for its next spec version; a new task
means a new instance. The stop contract (exit codes 0/1/2) is not
touched by stage data. → gate, ratify.

**State check** — the `state:` expectations of a row: files checked
after the run — exact or contains text (EOL-tolerant), or `from-run`
byte equality. → row.

**Submission (submit)** — delivering the door file to the machine:
`punchtape submit .punchtape/inbox/current.txt`. Accepted — the reply
is the thin ack (a run's adds the suite line with its red rows, a
ratification's the signature and verdict lines); refused — the one
diagnosis line, the draft stays. Everything that enters the machine
goes through the door. → door, document, thin ack.

**Suite** — the full row set of the corpus; the run's reply opens
with the suite line — `suite=green (N rows)`,
`suite=green (N rows: R ran, K kept green at unchanged inputs)`,
`suite=red (N rows, K red)` — one line per red row after.
→ freshness.

T

**Template catch (the door's own blank)** — the door's accident
insurance: after an accepted submission the door resets to the blank
template, and submitting those template bytes (or an empty file) is
refused unread — one line, the instance untouched, idempotent on
repeats; the accepted ack names the reset (DOOR: reset to the blank
template). A template that once decoded as a valid spec and drove
instances off their rails is now mechanically impossible.

**Thin ack** — the accepted submission's reply: ACCEPTED (the
transaction id), BYTES (the content sha), GATES (the gate names),
the DOOR reset note, the measured one-liners (a spec's version
counts; a run's suite line and red rows; a ratification's signature
and verdict line) — and no slot body: the slot rides the follow-up
`next`, and the slot cache makes it free. → NEXT SLOT, submission.

**Timeout** — the mechanism's ceilings: a row runs under 60 s by
default (its own `timeout` may raise it, up to 300 s); the build
gets 300 s. Parameters of the mechanism, not budgets for the agent —
the machine sets no limit on how long the agent may work. →
env-fail.

**Transaction** — an atomic proposal of a state change: a submission
is either accepted whole (validation, the corpus change, — for a
run — the execution of every stale row, the checkpoint, the stage
advance, all under one writer monopoly) or refused with one line of
reason without a change of state; the state is always derivable
from the journal. → submission, journal.

**Tree moved (the red line)** — the live-tree identity control: a
digest after the run must equal the one before it outside the
ignores; a tree that moved under its own measurement makes the run's
greens non-proof — they flip env-red and the line names the paths.
A row with `tree: unchanged` runs its own row-scoped control
(outside `.punchtape/`, the ignores and the row's own seeds, nothing
it leaves survives it). Rows clean up after themselves: seeds are
removed, outputs belong in build-ignore or die inside the row.
→ build-ignore, checkpoint.

**Typed heads** — the mechanical row families as heads of the
grammar, strictly checked — no shorthand, no bundle:
`refuse: {args, says, rc}` (one refusal; decode-only: IS the row
`{run: args, out: "", err: says, rc: rc}`), `det: {run, same}`
(the determinism pin; decode-only: IS `{run: run, out: same}`),
`witness: {probe}` (the staged-bytes probe — its bytes ride the
door and run staged under the row's modes),
`repeat-run: <command>` (the idempotency witness — runs twice,
agreement is green). Seeds and timeout compose with every head.
`reference/grammar.md`.
→ procedural modes, row.

V

**Verbs** — the command surface of the machine, exactly four:
`next` (the slot; with a brief — birth), `submit` (the door file),
`spec --check` (the dry validator) and `report` (the read-only
table of the stored runs). The working loop lives on the first two.
The human writes no commands — they speak the brief, and their
agent carries it to the machine.

**Verdict (VERDICT)** — the machine's final output on a task's
readiness: READY / NOT-READY, with the full account (the version
line of a snapshot, atoms, stages, origin, bill, acceptance, the
retired rows) and the digest of the exact tree it was pronounced
on. Analogy: accepting a car by checklist, not by "seems fine".
→ acceptance, drift, verdict snapshot.

**Verdict snapshot (the versioned freeze)** — every successful
ratification freezes an IMMUTABLE snapshot — the journaled file
`.punchtape/verdicts/vN.txt`, written once inside the ratifying
transaction: v1 is pinned to its digest forever, v2 to its own —
the model of git commits; the case lives on, the snapshots never
move. At deliver the slot renders the latest snapshot; a drifted
tree appends the honest drift line without touching the snapshot.
The same step regenerates SUMMARY.md and spec.md. → drift, verdict.

W

**Wish** — the older word for the brief; the state field that
holds the brief is still named `wish`, the surfaces all say brief.
The wish as a submission kind died with the one-document grammar.
→ brief.

**Work tree (work dir)** — the working directory: the agent's
territory (the product's code lives here, written directly, never
read by the machine) plus the machine's `.punchtape` and AGENTS.md.
The rows run here, in place. → live tree, .punchtape.

**Write zone** — the rule of who writes files, and how: the canon
atoms, the state, the checkpoints and the verdict snapshots — only
by transactions under the lastgood manifest (the recorded digests
of the last machine-written state); the journal and ledger —
append-only; the renders — regenerated; the cache — rebuilt. The
working root outside `.punchtape` is the agent's territory — the
machine writes only AGENTS.md beside the instance.
→ writer monopoly.

**Writer monopoly** — one write lock is taken for a whole submission:
read, validate, apply, run the rows, commit — a single transaction;
a second writer gets an honest refusal after short retries. →
transaction.

## Experiment and run vocabulary (a separate world — the machine never uses these words)

The terms below belong to the EXPERIMENT layer (campaigns, control
waves, runs of hands) that measures the machine from the outside.
They are defined here for the reader of the experiment materials
only; the machine's own code, canon and instance texts never carry
them — the machine stays agnostic forever (its own counterparty is
the agent, its own unit is the instance).

- **Hand** — an experiment participant: an LLM agent session of one
  arm of a control wave (a machine hand works through punchtape; a
  solo hand works without it). The machine never speaks of hands.
- **Cell** — the isolated working directory the experiment allocates
  for one hand: an `input/ env/ work/` layout under the run root's
  `cells/` (two-level random names — the measurement's blindness).
  The machine's own word for its working directory is the instance.
- **Wave** — one control run of the experiment: cells of both arms on
  the same tasks, measured against frozen etalons and the economy
  gates.
- **Campaign** — the whole experimental series (tasks, briefs,
  assignments, materials, tools, freezes) that owns waves.
- **Etalon** — the pre-ratified reference spec of a task used by the
  experiment's mechanical runner to measure a product.
- **Operator** — the human role of the experiment layer (the
  demanding customer of a solo hand, the relaxed owner of a machine
  hand); outside the experiment there is only the machine's user.
