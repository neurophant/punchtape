# Glossary
**English** · [Русский](GLOSSARY.ru.md) · [中文](GLOSSARY.zh.md)

The dictionary of the machine's vocabulary — every word a reader
of its zone files, its terminal lines and its contract can meet.
The principle: if a reader might not understand a word, it is
explained here, in plain language, with an example or an analogy;
`→` points to related terms inside this dictionary. The machine's
own closed vocabulary is small — the eight law ids, the six
outcomes, the row standings (green, red, stale), the stream's line
numbers (D1, D2, …) — and everything else below explains that
vocabulary and the world immediately around it.

A

**Accept** — what the machine does with a document that passed
every rule: its canonical bytes and its merged behaviors land in
the → object store, the submission record lands in a fresh journal
epoch, the machine takes the stream's unjournaled words itself,
the standing moves, and the submission slot is consumed. The
opposite of → refusal. The acceptance lives in `state.md`'s
counters; the terminal line names the next step, never the
ceremony.

**AGENTS block** — the machine's discovery hook: a marked block
inside the working root's `AGENTS.md`, bounded by `<!-- punchtape -->`
and `<!-- /punchtape -->`, ensured whole on every run. Identical
content touches nothing; changed content replaces itself; a
foreign file without markers gains the block at its end. Bytes
outside the markers are never touched; a lone marker or an
inverted pair is damage the machine refuses loudly. Analogy: a
library book's plate — glued inside the cover, the pages are not
the librarian's. → working tree.

**Answer** — one of the six → outcomes: the line of the stream
asks a question, and the answering words ride the intent itself
(`text`). An answer intent locks the tree for its run: answering
is not allowed to move the project. → tree lock.

**Argument refusal** — the one line the binary answers to any
launch that carries arguments other than the single `--version`:
the refusal names the law (the call takes no arguments) and is the
whole action of that launch — no state is touched, exit 1. → the
call.

**Ask** — one of the six → outcomes: the line of the stream is an
open question for the human, carried by an awaiting → order. An
awaiting order holds the verdict until a later human line, given
its own outcome, closes it — cited by a decision note. Nothing
closes itself by default; the human's silence is an honest stop.

**Awaiting order** — an → order in its `awaiting` state: the
question stands before the human, the verdict waits. It closes
only over a later human line — occupied by a carrier, cited by a
decision note; until then it holds every verdict. The alternative
state is → deferred.

B

**Behavior** — one named unit of visible behavior: an id
(`[a-z0-9-]` segments joined by `/`, say `tool/build`), a one-line
name, the author's → prose, its → claims, optional → covers, and
at least one → row — one of which must be → pinned. The id is the
merge key: a submission repeating an id replaces the behavior's
head and merges its rows. A behavior with rows and no claims
re-runs every run — honestly always → stale.

**Birth** — the act that creates the instance: one atomic `mkdir`
of the journal directory (exactly one winner among parallel runs —
the losers continue as ordinary passes), the birth record (the
creating machine's version), the genesis chain link over the
stream's initial lines, and the zone's standing files written from
the canon. Nothing else ever creates an instance, and only the
first non-empty line of the stream triggers it. Analogy: cutting
the first punch in a fresh tape. → instance, → stream.

**Build** — the corpus's optional declared command. When it
stands, it runs FIRST on every launch, before any row, under the
same execution law as a row. A build that exits non-zero, hangs
past the → deadline, or never starts is red for the whole run:
every row is frozen — nothing executes, nothing moves except the
build's own journaled fact. Declaring a build in a document
replaces the standing one; an absent field keeps it. A declared
build's artifacts are work — they may legally move the tree.
→ machine run, → red build freeze.

C

**The call** — the machine's entire interface: `punchtape`, no
arguments, anywhere at or below the working root. Every answer
beyond the terminal's one line is a file of the → zone.
→ terminal line, → exit code.

**Candor** — `law:candor`, the law of the loud record: nothing the
work skips, permits or sets aside may pass in silence. The loud
exclusion with its reason and its references; the mandate naming
the line and the replaced thing; orders with real states,
deferrals with their reasons; notes whose references resolve;
prose everywhere non-empty — every mark is one the human can
weigh. A refusal citing candor names the quiet hole.
→ mandate, → exclusion (loud), → note, → law ids.

**Canon** — the machine's one embedded source of everything it
prints as law: the English contract (the base and the ruler), the
Russian and Chinese mirrors (data for humans), the document
template, the reference page with its mirrors, the example, the
law map and the environment rule. The zone's contract, template
and reference are generated from it at the birth and refreshed
when it changes; an example that would lie about the machine fails
the build. → contract, → law-map, → reference.

**Canonical bytes** — a document's one official spelling: strict
decode into the model, then the canonical marshal (two-space
indent, sorted map keys, fields in declaration order, the
command's author's spelling preserved). Two spellings of one
content are one content — the machine compares and journals
canonical bytes only. Example: a re-indented `inbox.yaml` is the
same submission as the original.

**Carrier** — the thing that carries a declared outcome: answer
carries its text, execute its run; work and describe ride a
requirement whose lines carry the line; mark rides a marker on it;
ask rides an awaiting order of the same document. The carrier must
stand in the same submission that declares the outcome; removing
a carrier re-opens its line. → intent, → outcome,
→ unresolved lines.

**Chain** — the sha256 spine of the → stream's past: each
demand-append record carries the hash over the previous link's
bytes joined with the new lines. Every run recomputes the whole
chain from the genesis over the current bytes; a link that does
not recompose is → zone damage, and the run never starts. The
past is immutable; the tail is open.

**Claim** — a line range of the tree a behavior stands on:
`path/file.go:12-48`, root-relative, first ≤ last. One range names
one behavior — a claim already held is refused at its second
holder. The claimed lines' bytes feed the behavior's claims digest
inside every → proof; a change inside a claim makes the behavior's
rows stale. Coverage lives in claimed lines, never in globs.

**Clean environment** — the base every command of the machine
sees: exactly the names the canon declares — `PATH` and `HOME` —
taken from the shell that runs the machine, plus the zone's
→ pins, nothing else. The set actually passed, together with the
machine version, the architecture and the kernel release, forms
the → environment hash.

**Completion act** — the sign-off: an `inbox.yaml` carrying
exactly `done: true` and nothing else. Not a document — the bare
deliberate door to the → verdict. The run continues its ordinary
steps, and the verdict step decides: every gate green — a verdict
is recorded; anything short — the report names the holes. Counted
as its own journaled fact, separate from submissions. → machine
run, → done.

**Contract** — the machine's whole law as one text (`contract.md`
in the zone, from the canon), printed in full to the terminal at
→ first contact and whenever a zone stands unborn. Eight laws with
ids — `law:demand` (the stream, its numbering, the outcomes and
the tree that does not move under quiet intents), `law:document`
(the one format and its merge laws), `law:paths` (root-relative
paths only), `law:claims` (line-range coverage and the universe),
`law:proof` (pinned bytes, one run one behavior, staleness),
`law:candor` (the loud mark, the mandate, the orders), `law:done`
(the completion act and the verdict's gates), `law:your-tree` (the
execution environment) — plus the cycle, the zone files and the
call. The English text rules; the mirrors carry the same law set.

**Corpus** — the standing document: the behaviors, the registry
and the build command as they stand after every accepted
submission. Its canonical bytes live as `corpus.yaml` (a generated
image, healed every run) and inside the verdict's signature.
→ corpus version.

**Corpus version** — the counter of accepted submissions: every
acceptance raises it by one. A verdict records the version it was
pronounced on; `state.md`'s verdict line names both, and the next
instruction distinguishes "a verdict stands on the current
version" from "all green, the act is next". → verdict.

**Covers** — the behavior's claim of proof: the list of
requirement ids the behavior's rows stand for. A cover must name a
requirement that stands after the submission's merge; retiring a
requirement demands the same submission clear every cover of it. A
standing requirement no claiming behavior covers is an → orphan
requirement and holds the verdict.

D

**D range** — the one spelling of stream lines: `D<n>` or
`D<a>..D<b>` ascending, comma-joined (`D1..D4, D9`). The numbers
name the stream's non-empty lines in file order, re-derived from
the bytes on every run; requirements' spans, markers' lines,
intents' lines and orders' demands ride this grammar, and the
machine spells number sets back in it with consecutive runs
collapsed (`D19, D23..D25`). → stream, → requirement, → marker.

**Deadline** — the ceiling every command of the machine runs
under — each row, the build and each execution alike: at the
deadline the whole process group is killed and the fact spells
`timed out after Ns`. The default is 300 seconds; a zone raises it
by one pin — `PUNCHTAPE_ROW_TIMEOUT` names whole seconds. A
hang-killer, never a judge of slow work: the machine limits how
long one command may hang, never how long the agent may work.
→ process group, → red build freeze, → pins.

**Deferred order** — an → order set aside with its reason: the
question stays visible on the state page forever (`deferred
orders:` with the demand, the question and the reason), but it
does not hold the verdict. The visible mark of "the human said
later". → awaiting order.

**Describe** — one of the six → outcomes: the line asks for a
description of what stands, carried by a requirement whose lines
carry the line. Like answer and execute, describe locks the tree:
describing is not allowed to move the project. → tree lock.

**Digest** — the machine's unit of identity: the sha256 of exact
bytes, in hex. Same bytes — same digest, different bytes — a
different digest with overwhelming certainty. Digests reconcile
the machine's world: line-range digests, claims digests, proofs,
object addresses, the tree digest, the verdict's signature.
Human-facing spellings cut it to the first twelve characters —
the → token. Analogy: a fingerprint — not the finger, but enough
to name it. → object store, → proof.

**Done** — `done: true`: the two-word completion act, exactly this
and nothing else in the document. Any other field beside the flag
is refused — the bare act is the point. → completion act.

**Drift** — the derived fact that the tree moved under a standing
verdict: a certified behavior's claims digest no longer equals the
verdict's recorded pointer. Nothing watches anything — the
comparison is arithmetic over journaled facts and fresh digests.
The state page's verdict line says `open (drift since run N)` —
the first run past the verdict that saw the move — and the drift
is resolved only by a submission landing after that witness run.
The verdict history names the run that caught each drift and the
run that healed it (the first later run with no red row).
→ verdict history.

E

**Environment hash** — the fingerprint of what the commands see:
the machine version, `GOOS/GOARCH`, the kernel release, and the
sorted `NAME=value` set actually passed (the clean base plus the
pins, the row-deadline pin included). It enters every → proof: a
session that changes the environment moves the proofs — rows go
stale, never falsely green.

**Epoch** — one numbered file of the journal stream
(`epoch-NNNNN.yaml` under `journal/`): records append to the live
epoch, one YAML document behind a line-start `---` separator. An
accepted submission rotates the stream before its records land —
every submission's records live in an epoch of their own; rotation
never stacks empty files. Loading reads every epoch in number
order. → journal, → torn tail.

**Exit code** — the call's only binary answer: 0 — the machine
worked (red rows are data, not an error); 1 — the call did not
happen (zone damage, a journal that could not load, a machine-
owned write that could not land, a bad pin or universe line, a
foreign machine version, arguments on the line).

**Exclusion (loud)** — the registry's open crossing-out: a real
uncovered line range, marked with a reason that cites a demand
line or a law, and references that resolve. An exclusion marking
no uncovered line is refused — a mark that marks nothing is noise;
an empty reason is refused the same way. Merged by range;
removable by the range key. The loud exclusions complete the
tree-equals-claims arithmetic of the verdict. → holes,
→ exclusions proposal.

**Exclusions proposal** — `exclusions-proposal.yaml`, the
machine's own draft of the exclusion list: every standing hole as
a ready range with empty reason and refs — the author fills the
reasons and drops what is claimed. Generated from the machine's
hole arithmetic; with no hole standing, the draft leaves. The
machine writes the paths; the judgment stays yours.

**Execute** — one of the six → outcomes: the line asks the machine
itself to run a command, carried by the intent (`run`). The
execution lands between the submission and the rows — exit and
stdout journaled (whole to 4096 bytes, the object store's address
beyond), the tree locked for the run. → execution fact.

**Execution fact** — the journaled record of one execution the
machine ran on the agent's behalf: the run, the lines answered,
the command, the exit, the stdout's digest and the stdout itself.
The last execution shows on the state page and the report.
→ execute.

**Expect** — the row's demand on what the command leaves behind:
any mix of `rc` (the exact exit code; absent = 0), `out` (stdout:
exact bytes, or a substring behind the `contains: ` prefix) and
`files` (root-relative paths read after the run, the same two
match spellings). At least one field must stand — an empty expect
is refused. The comparison is pure: no interpretation anywhere.
→ pinned row.

F

**First contact** — a launch with no `.punchtape` at or above the
working directory: the machine prints the whole contract and exits
0, creating nothing. The terminal's full-page exception — no file
channel exists yet, so the law itself is the answer. The same page
prints for a zone directory that stands without a journal,
whatever else it carries.

**Fold** — the rebuild of the present from the past: all journal
records, oldest to newest, replayed into the → standing — births
fix the origin, submissions move the corpus and apply the
carriers, runs set the row facts, verdicts stand, refusals and
completions count. The same records always fold to the same
standing. Analogy: replaying a punched tape from the start — the
tape is the truth, the machine's memory is derived.

G

**Green** — a row's standing: it ran and its every expectation
held (broken-facts list empty). A green row whose → proof still
stands is → kept green — not re-executed. The corpus is green when
every row is green; only a green corpus with pinned proofs can
close a verdict.

H

**Holes** — the tree's uncovered lines: lines no → claim and no
loud → exclusion names. `state.md` spells the first ten files with
counts and maximal ranges; `holes.md` carries the whole uncapped
ledger; an uncovered line holds the next verdict. A range spelling
that does not parse, or a line past a file's own end, claims
nothing — the lines it names stay honest holes. → tree walk,
→ exclusions proposal.

I

**Inbox (the submission slot)** — `inbox.yaml`, the zone's one
input file. Its presence means a submission waits: a document to
decode, validate and accept — or the → completion act — or a
→ refusal that changes nothing (the bytes stay, the author fixes
the named place and runs again). Its absence is an ordinary
state. Everything that enters the machine enters through this one
file.

**Instance** — one task of the machine: a working root, its zone
`.punchtape` (journal, objects, views, slots) and the AGENTS block
in the root. Born once, never dies; a genuinely new task is a new
stream in a fresh directory. One zone serves one machine version:
a zone born by another version is refused — archive and start
fresh. → birth, → zone.

**Intent** — the declared → outcome of new stream lines: which
lines, which of the six outcomes, the agent's retelling in prose —
and the outcome's own carrier inside the same document. Every new
line is declared by exactly one intent of the submission that
meets it; a line that already carries its outcome is never
redeclared. → answer, → execute, → describe, → work, → mark,
→ ask.

J

**Journal** — the one source of truth: the append-only record
stream under `journal/`; each accepted submission opens a fresh
epoch. Ten record kinds — birth, submission, demand-append,
no-new-demands, run, execution-fact, tree-lock, verdict, refusal,
completion — each carrying exactly its own payload, no wall clock
anywhere: order is a record's place, identity is its bytes. The
journal of one history is byte-identical on every machine. Every
other zone file is a generated image of it. Never edited by hand.
→ fold, → standing, → torn tail.

K

**Kept green** — a green row whose recorded proof equals the
current proof: the row is not executed, not read, not touched —
its recorded green stands. The whole economy of the machine:
unchanged inputs cost nothing. → proof, → stale.

**Knowledge** — the zone's memory file `knowledge.md`: every note
of every accepted submission exactly once (numbered, with its kind
and references), then the → lesson entries grouped by reason.
Generated from the journal; the notes themselves are never
interpreted.

L

**Law ids** — the contract's eight named laws, the vocabulary of
every refusal: `law:demand`, `law:document`, `law:paths`,
`law:claims`, `law:proof`, `law:candor`, `law:done`, `law:your-tree`.
A refusal always cites one — the named rule and the contract
always speak of the same thing. → refusal.

**Law-map** — the canon's parity table: every validation rule
bound to the law the contract prints, and every cited law one the
contract prints. What is enforced is printed; what is printed is
enforced.

**Lesson** — one fixed red, fixed as pure data: the reason class,
the row key, the broken facts, the row's bytes before and after
(answered by the journal's row-bytes index), and the two run
numbers. Written by the run record itself — zero tokens of any
model. The lessons land on the knowledge page, grouped by reason.
Example: a row red on `exit: got 1, want 0` at run 4, green at
run 5 — the lesson carries both rows and both numbers.

M

**Machine at birth** — the machine version the birth record
carries: the instance's own origin fact, printed by the state page
(`machine at birth:`). A zone born by any other version is
refused before a single record folds. The version also enters the
environment hash, so a new machine honestly invalidates old
proofs. → birth, → environment hash.

**Machine run** — one execution of the call past the birth: the
chain check, the submission step, the executions, the build, the
fingerprinting, the execution of stale rows, the journaling, the
verdict, the render. The → run number counts only runs that
changed facts (a build ran, a row executed, an execution landed);
a run that measured and rendered changes nothing.

**Mandate** — the permission to replace standing work: the demand
line that ordered the change and the standing thing it permits to
replace — a behavior id or a requirement id, replaced by the very
same submission. A submission that replaces or retires standing
work without a mandate is refused; a mandate naming something this
submission does not replace is refused the same way. The line
proves existence, not meaning — the citation is for the audit
eye. → candor.

**Marker (not-a-requirement)** — the registry's dismissal: stream
lines declared to demand nothing, with a reason a reader can
weigh. The line is either a demand or not, never both — a marker
on a requirement's line is refused. Merged by the lines' fold;
removable by the D-grammar key. Example: a title line marked with
the reason "a heading, not a demand".

**Merge** — the submission laws: behaviors merge by id (the listed
head replaces whole — name, prose, claims, covers — the rows merge
by → run key); requirements merge by id; markers merge by their
lines, exclusions by range, orders by demand and question; removals
retire first; `remove` retires standing keys; an absent field
changes nothing; a remove-and-re-add in one submission starts
fresh. The economy: the number of submissions never affects the
price — unchanged rows are not re-sent.

**Mark** — one of the six → outcomes: the line demands nothing,
carried by a → marker on it. The mark is loud by construction —
a reason, in writing, forever in the corpus.

N

**Next instruction** — the machine's one imperative, the same
text in the terminal line and `state.md`'s `next:` line: a refusal
says fix the named place in the inbox; build red says fix the
build; corpus empty says submit the first document; a registry
without behaviors says submit a behavior with a pinned row; each
verdict blocker speaks its own instruction; a verdict standing on
the current corpus version says show `state.md` to the human; all
green says write `done: true` and run to record the verdict.
Every state the machine can stand in has exactly one instruction —
never two truths.

**No-new-demands** — the signed statement that no new human words
came: the stream's line count at the signing and the run that
signed it, journaled beside the accepted submission. A signature
over words that did come stands in the journal forever — the one
point where the guard is visibility, not a gate: the state page
carries the signature, and the words stay unjournaled and
declarable. The next demand-append clears the standing signature.

**Note** — structured free text a document carries: one of five
kinds (directive, decision, exclusion, lesson, work), a text, and
references that must resolve — a demand line, a standing
requirement or behavior, a parseable line range, or a law id the
contract prints. The machine never interprets a note; it carries
and indexes it. A channel for humans, not a channel into the
machine.

O

**Object store** — the content-addressed storage under
`journal/objects/`: an object's name is the sha256 of its bytes;
its file is `<first two hex>/<remaining sixty-two>`. Identical
bytes are stored once — dedup by construction; every write is
re-read and verified, every read verifies its digest. Every
accepted document's canonical bytes and every merged behavior's
canonical bytes live here, immutable. → digest.

**Order (open)** — one question standing before the human: the
demand line it rides, the question itself and its state —
→ awaiting or → deferred. Awaiting holds the verdict; deferred
stays visible forever with its reason. Merged by the
demand-and-question pair; an order's birth facts are the machine's
own. → ask.

**Orphan requirement** — a standing requirement no standing
behavior with claims names in its covers. Holds the verdict
(`the requirements … stand uncovered by any behavior's covers`).
→ covers.

**Outcome** — the six spellings of what a stream line asks for:
`answer`, `execute`, `describe`, `work`, `mark`, `ask` — the
protocol's own vocabulary, declared by an → intent the moment the
line arrives. Answer and execute carry themselves (text, run);
the other four ride the registry of the same document. A line
without an outcome is an → unresolved line.

P

**Pins** — the zone's environment declarations: the `env` file,
one `KEY=VALUE` a line, the key `[A-Za-z_][A-Za-z0-9_]*`, a name
pinned once, no `export`, no quotes, no absolute paths (a pin
travels with the tree). Comments and blank lines pass, a trailing
carriage return is tolerated. A line of any other shape is a loud
refusal naming the file, the line and the law. A pin wins on a
name over the clean base; every pin rides the environment hash.
Example: `TOOLS=tools/bin` — the value may be a relative path,
never an absolute one.

**Pinned row** — a row whose expect pins exact bytes: an `out`
that is neither empty nor a `contains: ` prefix, or `files` where
every value pins exact bytes. A pinned row is the behavior's
proof; everything else is commentary. `rc` alone pins nothing — a
green exit code proves only that the command ended. Every
behavior carries at least one pinned row or the submission is
refused.

**Process group** — the launch discipline: every command of the
machine runs in its own group, and whatever it left alive
(servers, workers, what a shell raised) is killed with the group —
a kill signal to the group and the process, then a bounded wait
for actual death. A row's leftovers never outlive the row.

**Proof** — a row's fingerprint: the sha256 over the row's
canonical bytes, its behavior's claims digest and the →
environment hash. While the recorded proof equals the current
proof, the row is → kept green; any of the three moving makes it
→ stale. Analogy: a wax seal over the row, its inputs and its
surroundings — break any one, the seal breaks.

**Prose** — the words the author owns: the requirement's unwinding
of the demand in human words, the behavior's story of what the
user gets, the intent's retelling of the line. Mandatory and
non-empty; carried verbatim onto the spec pages and the trace. The
machine checks presence, never quality — the reading is the
human's.

R

**Red** — a row's standing: it ran and at least one expectation
broke (the broken facts, journaled as measured, are the reason).
Red rows group by reason in the report. The skip law is
color-blind: a standing red re-runs exactly when its → proof
moves — its inputs changed — the same law that keeps a green row
green; but a standing red never passes for done, because the
→ verdict reads the fact's status, never the proofs alone. Red
rows under a standing verdict are → drift.

**Red build freeze** — the law that a failed build stops
everything: no row executes that run, no row fact moves; the run's
record carries the build's fact alone and the next instruction
says `build red`. The submission that rode the same run still
lands — the corpus moves, the rows wait.

**Reference (the page)** — `reference.md`, the zone's own manual:
every door of the machine on one page — the call, the stream, the
document, the intents, the registry, the merge, the execution, the
universe, the verdict, the faces, the pins, the damage, the
constants — written in the canon, generated into the zone at the
birth and refreshed whenever the canon changes. The full knowledge
always under hand, so the agent never guesses.
→ canon, → contract, → template.

**Refusal** — the machine's answer to a broken submission: one
line per broken rule — the exact field, the → law ids, the
measured fact. The submission's bytes write nothing; the refusal
itself is journaled as the machine's own fact and counted on
`state.md`; the document stays in `inbox.yaml`, the author fixes
the named place and runs again. Example:
`behaviors[0].rows[1].expect: the expect is empty — name rc, out
or files (law:document)`. → validation.

**Registry** — the corpus's account of the words: the requirements
(each with its stream-line spans), the markers, the loud
exclusions and the open orders. Merged by the submission laws;
digested whole into the verdict's → registry digest.

**Registry digest** — the sha256 over the canonical bytes of the
requirements, markers, exclusions and orders together — the
words' whole account sealed in one digest and carried by the
verdict record.

**Remove** — the document's retirement list, four spellings: a
requirement id, a behavior id, a `D<n>` marker-or-order key, a
`path:first-last` exclusion key. Every key must stand; removals
retire first in the merge; retiring a requirement re-opens its
lines and demands the same submission re-carry every cover of it;
a behavior retires whole with its rows, under a → mandate.

**Report** — `report.md`: the last fact-changing event's whole
picture — for a run, the build's line, the executions, the green
count and every red row grouped by its broken reason, plus the
verdict's notes when the completion act spoke; for a refusal, one
line per refusal and the fix pointer. Written whole when it moves;
a run that changed nothing and closed nothing leaves the previous
report standing. → refusal, → verdict, → views.

**Requirement** — one demand made specific: an id (`R` plus
letters, digits and dots, say `R1`), the stream lines it occupies
(D ranges), the unwinding in prose, an optional `refines` parent.
Merges by id; retires only through an explicit `remove` that also
clears every cover of it. The stream is resolved whole when every
line carries its outcome and its carrier.

**Row** — one executable check inside a behavior: a single command
(`run` — a shell line or an argv list) and its → expect. The
command's canonical bytes are the row's → run key and part of its
→ proof; the row's place in the behavior is its journal key
(`<behavior id>/<index>`). Rows are all a behavior proves.

**Run key** — the command's canonical bytes, the row's merge key:
a submitted run that stands replaces that whole row in place; a
new run appends; unlisted rows stand. A re-spelled command is a
new run — the standing one stands; one command proves one
behavior. Example: `./tool src out` (a shell line) and
`["./tool", "src", "out"]` (an argv) are two different runs — one
through the shell, one direct.

**Run number** — the counter of fact-changing runs. The terminal
line's `run <N>`, the report's, state's, trace's, knowledge's and
spec's headers all name the same number. A refusal costs no run
number — it is counted by its own counter.

S

**Signature (the verdict's)** — the sha256 over the corpus's
canonical bytes plus, per behavior in the corpus's own order,
`\0 id \0 claims digest`: the spec and the exact tree state it
was pronounced on, in one digest. Shown short (twelve characters)
on the state page and in the verdict history.

**Spec tree** — the view directory `spec/`: `index.md` plus one
small page per living behavior — the id and name, the run number,
the Story, the Demands it answers, the Claims with their per-range
digests, the Rows with their proof standings, the Verify commands.
Retired behaviors' pages are removed — the tree is an image of the
standing corpus, nothing else.

**Stale** — a row's standing: its → proof moved (its bytes, its
behavior's claimed lines, or the environment) — or it never ran —
or its behavior declares no claims. Stale rows execute on the next
run; the instruction counts them (`N row(s) stale — run again to
execute them`).

**Standing** — the present rebuilt from the journal by the →
fold: the birth facts, the chain, the corpus version and build,
the registry, the behaviors and their order, the row facts and
the row-bytes index, the line outcomes and carriers, the run
number, the verdict, the counts, the lessons, the tree-lock hold.
Everything the machine decides, it decides on the standing — and
the standing is always derivable from the records alone.

**State page** — `state.md`: the standing on one page — the
demands accounting and unresolved lines, the queue, the signature,
the corpus version, the counters, the machine at birth, the
verdict with its drift mark, history and deferred orders, one line
per behavior, the uncovered and excluded lines, the knowledge
counts, and `next:` — the same text as the terminal line.
→ standing, → views.

**Stream** — the human's words themselves: `demands.md`, open
forever, every non-empty trimmed line numbered D1..Dn in file
order. The words enter the file by the agent's hand; the machine
takes them itself at the next submission (never retyped into
documents), seals the past with the → chain, and journals each
submission's words or its → no-new-demands signature. A touched
past line is → zone damage.

**Submissions** — the accepted documents, counted on `state.md`
beside the refused and the completed — three separate economies:
documents the slot took, submissions it refused, completion acts
that finished the work.

T

**Template** — `template.yaml`, generated from the canon at the
birth and refreshed with it: the document's shape with teaching
comments — every field, the four rules of YAML cooking, the
remove spellings. Copy, edit, submit as `inbox.yaml`. A document
that decodes but breaks a rule is still refused — the template
teaches the shape, the validation holds the law.

**Terminal line** — the call's one line of output on a successful
run: `punchtape: run <N> — <instruction>`. The instruction names
the file to touch and is the same text as `state.md`'s `next:`
line. The exceptions: first contact prints the whole contract;
the unborn zone prints its one instruction line; a call that
cannot happen answers its single error line instead — exit 1.
Everything else the machine has to say lives in the zone's files.

**Token** — the human-facing short spelling of a → digest: its
first twelve hex characters. The state page's verdict signatures
and row proofs, the spec pages' behavior tokens, the trace's
claims tokens — all tokens; the eye compares them, the machine
compares the whole digests. The token is for reading; the full
digest is the identity. → digest.

**Torn tail** — a journal record cut mid-write (a crash, a lost
power). It does not blind the machine: the intact records load,
the torn bytes quarantine to a sidecar (`<epoch>.torn`), and the
epoch file is rebuilt byte-for-byte from its intact records. A
chunk that does not parse outside the final position is damage by
hand — an honest error. → epoch.

**Trace** — `trace.md`: the stream's showcase — every line with
its own text, its declared outcome, the retelling, and the
carriers below (the requirements, the behaviors with claims and
proof rows, the markers, the answers, the last executions, the
orders). Nothing is judged; everything is placed for the eye. The
one page built for the human's audit glance.

**Tree digest** — the working tree's own fingerprint: the sha256
over the sorted `path\0digest` lines of every counted file —
binaries included by their bytes, the zone and the
version-control bookkeeping excluded by name. Measured before and
after a locked submission: a moved digest is the journaled
→ tree-lock fact.

**Tree lock** — the hold against quiet moving: a submission whose
intents stand and none is `work` (and no build stands or is
declared — a build's artifacts are legal work) locks the tree for
its run. The digest is measured before and after; a move is
journaled and holds the verdict until the bytes return or the
change is submitted as declared work. A work intent lifts a
standing hold and arms none. → answer, → execute, → describe.

**Tree walk** — the machine's one way of seeing the working tree:
a single ordered walk that skips the zone (`.punchtape`) and the
version-control bookkeeping directories (`.git`, `.hg`, `.svn`) by
name wherever they stand, filters by the declared → universe,
counts every inside file's lines and digests its bytes, and counts
the outside files by name alone — loudly, never read. A file with
a NUL byte, an empty file or an unreadable file stands outside the
coverage universe wholesale, with the named reason. The machine
never reads the tree's content as code — only as bytes to digest
and commands to run. → walk cache.

U

**Uncovered lines** — the holes of the tree: lines no claim and no
loud exclusion names. Named per file with counts and maximal
ranges — the first ten files on the state page, the whole ledger
in `holes.md`, the draft in the exclusions proposal. An uncovered
line holds the next verdict; a standing verdict may live over
them, visibly (`standing over N uncovered lines`). → holes.

**Universe** — the zone's optional declaration of what counts:
`.punchtape/universe`, one root-relative path a line, `#` a
comment, no declaration meaning the whole tree. A root that names
no file of the tree is refused loudly — never a silent loss of
coverage. Files outside the universe stand counted on the state
page, never judged. → tree walk.

**Unresolved lines** — the stream's lines without a declared
→ outcome, or without the standing carrier of their declared
outcome: "no outcome", "work pending", "describe pending",
"mark pending". Named with reasons on the state page and the
trace; an unresolved line holds the verdict. Removing a carrier
re-opens its line.

V

**Validation** — the whole rule set a document passes before
acceptance: the strict decode and the armor walk; id alphabets and
uniqueness; the D-range grammar against the stream's universe;
outcome carriers in the same breath; claims that are line ranges
the tree carries, one range one behavior; behaviors with names,
standing covers, rows with expectations, at least one pinned row,
runs unique in their rows and across behaviors; markers off the
requirements' lines; exclusions over real holes with citing
reasons; orders on standing lines with real states; a mandate for
replaced standing work; notes whose references resolve; removes
naming standing things; the done document exactly bare — plus the
machine's own gates: the dangling covers, the duplicate claim, the
ask-close law. Every refusal names the field, the fact and the law
id. → refusal.

**Verdict** — the machine's pronouncement of done, recorded only
on the → completion act and only when every gate stands: the
declared build green; the corpus non-empty; every row green on the
current fingerprints;
every behavior proven by a green → pinned row; the stream resolved
whole; every requirement covered; the tree equal to its claims
plus loud exclusions; no order awaiting; no drift unnamed; no
tree-lock hold. The record carries the number, the corpus
version, the → signature, the → registry digest, every behavior's
claims pointer and the closing run; a completion record lands
beside it. Append-only — a new verdict never rewrites an old one.
Anything short is `not yet` plus one line per class of hole in the
report, in the one order every face consumes.

**Verdict history** — the state page's derived chronicle: every
recorded verdict (run, number, version, short signature); every
→ drift (a run with red rows under a standing verdict, with the
row keys); every heal (the first later run with no red row). The
recent ten events, the elision counting the rest. What the journal
does not carry, the history does not say.

**Version flag** — `punchtape --version`, exactly this spelling
and alone on the line: prints `punchtape <version>` and exits 0
before any state is touched — the one human affordance; the
contract never mentions it. → the call, → exit code.

**Views** — the zone's generated files: `state.md`, `report.md`,
`trace.md`, `knowledge.md`, `holes.md`, the exclusions proposal,
the spec tree, `corpus.yaml`. Images of the journal — rewritten
whole when their content changes, never touched when it does not;
a view that cannot land is a loud failure (a stale face is a
lie).

W

**Walk cache** — `cache/tree`: the verified accelerator of the
tree walk, one record a file — path, size, mtime, line count,
digest. A file whose size and mtime stand as the cache last saw
them reuses the cached facts instead of its bytes; the proofs
never ask the cache — claimed lines are read directly, always.
Written once at the run's end, only when the walk moved it.
Delete it and the next run recomputes the same numbers — an
accelerator, never a source of truth. → tree walk.

**Work** — one of the six → outcomes: the line orders a change to
the project, carried by a requirement whose lines carry the line.
The work intent is the legal way to move the tree — it lifts a
standing tree-lock hold and arms none. → tree lock.

**Working root / working tree** — the directory carrying the zone
(the call finds it at or above its starting point) and everything
beside it. The tree is the agent's territory: the machine walks
it, digests it, runs commands in it — and never reads its content
as code.

Z

**Zone** — `.punchtape/`, the only directory the machine owns: the
teaching pages (`contract.md`, `template.yaml`, `reference.md`),
the slots (`demands.md`, `inbox.yaml`, `env`, `universe`), the
generated standing files (`corpus.yaml`, `state.md`, `report.md`,
`trace.md`, `knowledge.md`, `holes.md`,
`exclusions-proposal.yaml`, `spec/`), the journal with its
objects, and the machine's own `cache/`, `tmp/` and `run.lock` —
the zone's own lock: one pass at a time, freed by the kernel even
on the holder's death. Never edited by hand; one record stands outside it — the → AGENTS block in the
working root. → zone damage.

**Zone damage** — the machine's loudest refusal: the past broke,
or the zone is not this machine's to read. A touched past line of
the stream, a shrunken or deleted `demands.md` on a born instance,
a journal record that does not re-compose, a zone born by another
machine version, damaged AGENTS markers — one line, exit 1, the
run never starts. The journal and the chain hold the past; damage
by hand is named, never guessed around.
