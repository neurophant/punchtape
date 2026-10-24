# Glossary
**English** · [Русский](GLOSSARY.ru.md) · [中文](GLOSSARY.zh.md)

The dictionary of the machine's vocabulary — every word a reader of
its zone files, its terminal lines and its contract can meet. The
principle: if a reader might not understand a word, it is explained
here, in plain language, with an example or an analogy; `→` points to
related terms inside this dictionary. The machine's own closed
vocabulary is small — the law ids, the row standings (green, red,
stale), the wish-line numbers (W1, W2, …) — and everything else
below explains that vocabulary and the world immediately around it.

A

**Accept** — what the machine does with a document that passed every
rule: its canonical bytes and its merged behaviors land in the object
store, the submission record lands in a fresh journal epoch, the
standing moves, and the submission slot is consumed. The opposite of
→ refusal. Example: `inbox.yaml` carrying one new behavior is
accepted; the next terminal line names the next step, not the
acceptance — the acceptance lives in `state.md`'s counters.

**AGENTS block** — the machine's discovery hook: a marked block
inside the working root's `AGENTS.md`, bounded by `<!-- punchtape -->`
and `<!-- /punchtape -->`, ensured whole on every run. Identical
content touches nothing; changed content replaces itself; a foreign
file without markers gains the block at its end. Bytes outside the
markers are never touched. Analogy: a library book's plate — glued
inside the cover, the pages are not the librarian's. → working tree.

**Argument refusal** — the one line the binary answers to any launch
that carries arguments other than the single `--version`: the
refusal names the law (the call takes no arguments) and is the whole
action of that launch — no state is touched, exit 1. → the call.

B

**Artifacts** — the document's declared outputs: root-relative globs
naming the files the work is expected to leave behind. A tree file
covered by no scope and matched by no artifact glob is an → uncovered
source; one matched by an artifact stands in `state.md` under
"artifacts (declared)". Example: `artifacts: [out/**]` declares the
build's output directory. → uncovered sources.

**Behavior** — one named unit of visible behavior: an id (`[a-z0-9-]`
segments joined by `/`, say `tool/build`), a one-line name, the
author's → prose, an optional → scope, optional → covers, and at
least one → row. The id is the merge key: a submission repeating an
id replaces the behavior's head and merges its rows. A behavior with
rows and no scope re-runs every run — honestly always → stale.

**Birth** — the act that creates the instance: one atomic `mkdir` of
the journal directory (exactly one winner among parallel runs — the
losers continue as ordinary passes), the wish's bytes into the object
store as the journal's first record, the wish slot consumed, and the
zone's standing files written from the canon. Nothing else ever
creates an instance. Analogy: cutting the first punch in a fresh
tape. → wish slot, → instance.

**Brief map** — the wish's accounting on one line of `state.md`:
`brief: <claimed>/<lines> lines; requirements <n> standing, <k>
covered; not-a-requirement: <m>`, followed by every marker spelled
out and the uncovered wish lines when any. The same arithmetic the
verdict's wish gate reads — never two truths. → wish coverage,
→ registry.

**Build** — the corpus's optional declared command (a shell line).
When it stands, it runs FIRST in every run, before any row, under the
same execution law as a row. A build that exits non-zero, hangs past
the deadline, or never starts is red for the whole run: every row is
frozen — nothing executes, nothing moves except the build's own
journaled fact. Declaring a build in a document replaces the standing
one; an absent field keeps it. → machine run.

C

**The call** — the machine's entire interface: `punchtape`, no
arguments, anywhere at or below the working root. Every answer beyond
the terminal's one line is a file of the → zone. → terminal line,
→ exit code.

**Canon** — the machine's one embedded source of everything it
prints as law: the English contract (the base and the ruler), the
Russian and Chinese mirrors (data, structurally complete), the
document template, the example, the law map and the environment rule.
The zone's contract files and template are generated from it at the
birth; an example that would lie about the machine fails the build.
→ contract, → law-map.

**Canonical bytes** — a document's one official spelling: strict
decode into the model, then the canonical marshal (two-space indent,
sorted map keys, fields in declaration order, the command's author's
spelling preserved). Two spellings of one content are one content —
the machine compares and journals canonical bytes only. Example: a
re-indented `inbox.yaml` is the same submission as the original.

**Clean environment** — the base every command of the machine sees:
exactly the names the canon declares — `PATH` and `HOME` — taken
from the shell that runs the machine, plus the zone's → pins, nothing
else. The set actually passed, together with the machine version, the
architecture and the kernel release, forms the → environment hash.

**Completion act** — the sign-off: an `inbox.yaml` carrying exactly
the standing corpus's canonical bytes (in practice: `corpus.yaml`
copied byte-for-byte into `inbox.yaml`). Not a document — the slot
sending the corpus back. The run continues its ordinary steps, and
the → verdict step decides: every gate green — a verdict is
recorded; anything short — the report names the holes. Counted as its
own journaled fact, separate from submissions. → machine run.

**Contract** — the machine's whole law as one text (`contract.md` in
the zone), printed in full to the terminal at → first contact. Five
laws with ids — `law:wish` (the wish slot and its numbered lines),
`law:document` (the one format and its merge laws), `law:paths`
(root-relative paths only), `law:done` (the completion act and the
verdict's gates), `law:your-tree` (the execution environment) — plus
the scope law, the refusal law and the document's shape. The Russian
and Chinese mirrors carry the same law set; the English text rules.

**Corpus** — the standing document: the behaviors, the registry and
the build command as they stand after every accepted submission. Its
canonical bytes live as `corpus.yaml` (a generated image, healed
every run) and inside the verdict's signature. Green when every row
is green — the verdict's ground. → corpus version.

**Corpus version** — the counter of accepted submissions: every
acceptance raises it by one. A verdict records the version it was
pronounced on; `state.md`'s verdict line names both, and the next
instruction distinguishes "a verdict stands on the current version"
from "all green, the act is next". → verdict.

**Covers** — the behavior's claim of proof: the list of requirement
ids the behavior's rows stand for. A cover must name a requirement
that stands after the submission's merge; retiring a requirement
demands the same submission clear every cover of it. A standing
requirement no behavior covers is an → orphan requirement and holds
the verdict.

D

**Deadline** — the 300-second ceiling every command of the machine
runs under — each row and the build alike: at the deadline the whole
process group is killed and the fact spells `timed out after 5m0s`.
A hang-killer, never a judge of slow work: the machine limits how
long one command may hang, never how long the agent may work.
→ process group, → red build freeze.

**Digest** — the machine's unit of identity: the sha256 of exact
bytes, in hex. Same bytes — same digest, different bytes — a
different digest with overwhelming certainty. Digests reconcile the
machine's world: file digests, scope fingerprints, proofs, object
addresses, the verdict's signature. Human-facing spellings cut it to
the first 12 characters. Analogy: a fingerprint — not the finger,
but enough to name it. → object store, → proof.

**Document** — the one input format, a strictly decoded YAML file:
`notes`, `build`, `remove`, `requirements`, `not-a-requirement`,
`artifacts`, `behaviors` (with `rows`). A document is a whole table
or a subset of one: what it does not list, it does not change. →
canonical bytes, → merge, → validation.

**Drift** — the derived fact that the tree moved under a standing
verdict: a run whose rows stand red while a verdict stands (the
history names the run and the red row keys), or a living behavior
whose current scope fingerprint differs from the verdict's recorded
pointer (the `scopes moved` mark on the state page). Nothing watches
anything — both marks are arithmetic over journaled facts and fresh
fingerprints. A later run with no red row heals the drift, also
named in the history. → verdict history.

E

**Environment hash** — the fingerprint of what the commands see: the
machine version, `GOOS/GOARCH`, the kernel release, and the sorted
`NAME=value` set actually passed (the clean base plus the pins). It
enters every → proof: a session that changes the environment moves
the proofs — rows go stale, never falsely green.

**Epoch** — one numbered file of the journal stream
(`epoch-NNNNN.yaml` under `journal/`): records append to the live
epoch, one YAML document behind a line-start `---` separator. A
submission rotates the stream before its record lands — every
submission's records live in an epoch of their own; an already-fresh
epoch stays (rotation never stacks empty files). Loading reads every
epoch in number order. → journal, → torn tail.

**Exit code** — the call's only binary answer: 0 — the machine
worked (red rows are data, not an error); 1 — the call did not
happen (zone damage, a journal that could not load, a machine-owned
write that could not land).

**Expect** — the row's demand on what the command leaves behind: any
mix of `rc` (the exact exit code; absent = 0), `out` (stdout: exact
bytes, or a substring behind the `contains: ` prefix) and `files`
(root-relative paths read after the run, the same two match
spellings). At least one field must stand — an empty expect is
refused. The comparison is pure: no interpretation anywhere.

F

**First contact** — a launch with no `.punchtape` at or above the
working directory: the machine prints the whole contract and exits 0,
creating nothing. The terminal's full-page exception — no file channel
exists yet, so the law itself is the answer.

**Fold** — the rebuild of the present from the past: all journal
records, oldest to newest, replayed into the → standing — births fix
the goal, submissions move the corpus, runs set the row facts,
verdicts stand, refusals and completions count. The same records
always fold to the same standing. Analogy: replaying a punched tape
from the start — the tape is the truth, the machine's memory is
derived.

G

**Goal** — the instance's fixed wish: the wish's bytes as a
content-addressed object, recorded at the birth with their digest —
immutable forever. The state page shows the digest and the object's
address; every wish line and every marker quote is derived from these
bytes, so the numbering cannot drift.

**Green** — a row's standing: it ran and its every expectation held
(broken-facts list empty). A green row whose → proof still stands is
→ kept green — not re-executed. The corpus is green when every row
is green; only a green, fresh corpus can close a verdict.

H

**Hash cache** — `cache/hashes.yaml`: the `(path, size, mtime)` →
digest accelerator of the tree walk, every entry tagged with the
digest law it speaks. A touched file recomputes honestly; a corrupt
or missing cache reads as empty; entries of a foreign tag drop out
and recompute rather than lie. Flushed to disk once per run. Delete
it and the next run recomputes the same numbers — an accelerator,
never a source of truth. → scope fingerprint, → tree walk.

**Holes** — the wish lines no requirement span and no marker claim:
`state.md` spells them (`uncovered brief lines: W…`) and the verdict
refuses on them (`not yet — the wish lines W… stand uncovered by any
requirement or marker`). A range spelling that does not parse, or a
number outside the wish's universe, claims nothing — the lines it
names stay honest holes. → wish coverage.

I

**Inbox (the submission slot)** — `inbox.yaml`, the zone's one input
file. Its presence means a submission waits: a document to decode,
validate and accept — or the → completion act — or a → refusal that
changes nothing (the bytes stay, the author fixes the named place and
runs again). Its absence is an ordinary state. Everything that enters
the machine enters through this one file.

J

**Instance** — one task of the machine: a working root, its zone
`.punchtape` (journal, objects, views, slots) and the AGENTS block in
the root. Born once, never dies; a genuinely new task is a new wish
in a fresh directory. → birth, → zone.

**Journal** — the one source of truth: the append-only record stream
under `journal/`; each submission opens a fresh epoch. Six record
kinds — birth, submission, run, verdict, refusal, completion — each
carrying exactly its own payload, no wall clock anywhere: order is a
record's place, identity is its bytes. The journal of one history is
byte-identical on every machine. Every other zone file is a generated
image of it. → fold, → standing.

K

**Kept green** — a green row whose recorded proof equals the current
proof: the row is not executed, not read, not touched — its recorded
green stands. The whole economy of the machine: unchanged inputs cost
nothing. → proof, → stale.

**Knowledge** — the zone's memory file `knowledge.md`: every note of
every accepted submission exactly once (numbered, its → tags on its
line), the tag index, and the → lesson entries grouped by reason.
Generated from the journal; the notes themselves are never
interpreted.

L

**Law ids** — the contract's five named laws, the vocabulary of every
refusal: `law:wish`, `law:document`, `law:paths`, `law:done`,
`law:your-tree`. A refusal always cites one — the named rule and the
contract always speak of the same thing. → refusal.

**Law-map** — the canon's parity table: every validation rule bound
to the law the contract prints, and every cited law one the contract
prints. What is enforced is printed; what is printed is enforced.

**Lesson** — one fixed red, fixed as pure data: the reason class, the
row key, the broken facts, the row's bytes before and after (the
before-bytes recovered from the journal's behavior objects by
digest), and the two run numbers. Written by the run record itself —
zero tokens of any model. A current red whose reason has a lesson is
pointed at from `state.md`. Example: row red on
`exit: got 1, want 0` at run 4, green at run 5 — the lesson carries
both rows and both numbers.

M

**Machine at birth** — the machine version the birth record carries:
the instance's own origin fact, printed by the state page
(`machine at birth:`); an instance whose record lacks one says
`(not recorded)`. The version also enters the environment hash, so
a new machine honestly invalidates old proofs. → birth,
→ environment hash.

**Machine run** — one execution of the call past the birth: the
submission step, the build, the fingerprinting, the execution of
stale rows, the journaling, the verdict, the render. The → run
number counts only runs that changed facts (a build ran or a row
executed); a run that measured and rendered changes nothing.

**Marker (not-a-requirement)** — the registry's dismissal: wish lines
declared to demand nothing, with the lines' own words quoted verbatim
beside a reason. The quote must match the wish's own text exactly —
dismissing a demand demands copying it into the signed corpus; the
mark cannot be quiet. Example: a title line marked
`quote: "a counter", reason: the wish's title names the work, it
demands nothing`.

**Merge** — the submission laws: behaviors merge by id (the listed
head replaces whole — name, prose, scope, covers — the rows merge by
→ run key); requirements merge by id the same way; the marker and
artifact lists replace whole when the submission carries them;
`remove` retires standing ids; an absent field changes nothing. The
economy: the number of submissions never affects the price —
unchanged rows are not re-sent.

**Next instruction** — the machine's one imperative, the same text in
the terminal line and `state.md`'s `next:` line: build red says fix
the build; corpus empty says submit the first document; rows red say
read the report and fix; rows stale say run again; a verdict standing
on the current corpus version says show `state.md` to the human; all
green says copy `corpus.yaml` to `inbox.yaml` and run to record the
verdict.
Every state the machine can stand in has exactly one instruction —
never two truths.

**Note** — free text a document carries (`notes: [{text: …}]`). The
machine never interprets a note: it stores the document's canonical
bytes, and the render lists the notes with mechanical → tags. A
channel for humans, not a channel into the machine.

O

**Object store** — the content-addressed storage under
`journal/objects/`: an object's name is the sha256 of its bytes; its
file is `<first two hex>/<remaining sixty-two>`. Identical bytes are
stored once — dedup by construction; every write is re-read and
verified, every read verifies its digest. The wish's bytes, every
accepted document's canonical bytes and every merged behavior's
canonical bytes live here, immutable. → digest.

**Orphan requirement** — a standing requirement no standing behavior
names in its covers. Holds the verdict (`not yet — the requirements …
stand uncovered by any behavior's covers`). → covers.

P

**Pins** — the zone's optional environment declarations: the `env`
file, one `KEY=VALUE` a line, the key `[A-Za-z_][A-Za-z0-9_]*`, no
`export`, no quotes, no absolute paths (a pin travels with the tree).
Comments and blank lines pass, a trailing carriage return is
tolerated. A line of any other shape is a loud
refusal naming the file, the line and the law. A pin wins on a name
over the clean base. Example: `TOOLS=tools/bin` — the value may be
a relative path, never an absolute one.

**Process group** — the launch discipline: every command of the
machine runs in its own group, and whatever it left alive (servers,
workers, what a shell raised) is killed with the group — a kill
signal to the group and the process, then a bounded wait for actual
death. A row's leftovers never outlive the row.

**Proof** — a green row's fingerprint: the sha256 over the row's
canonical bytes, its behavior's → scope fingerprint and the →
environment hash. While the recorded proof equals the current proof,
the row is → kept green; any of the three moving makes it → stale.
Analogy: a wax seal over the row, its inputs and its surroundings —
break any one, the seal breaks.

**Prose** — the behavior's own description in the author's words:
what the user gets and why. Mandatory and non-empty; carried verbatim
onto the behavior's spec page. The machine checks presence, never
quality — the reading is the human's.

R

**Red** — a row's standing: it ran and at least one expectation broke
(the broken facts, journaled as measured, are the reason). Red rows
group by reason in the report. The skip law is color-blind: a
standing red re-runs exactly when its → proof moves — its inputs
changed — the same law that keeps a green row green; but a standing
red never passes for done, because the → verdict reads the fact's
status, never the proofs alone. Red under a standing verdict is
→ drift.

**Red build freeze** — the law that a failed build stops everything:
no row executes that run, no row fact moves; the run's record carries
the build's fact alone and the next instruction says `build red`.

**Refusal** — the machine's answer to a broken submission: one line
per broken rule — the exact field, the → law ids, the measured fact.
The submission's bytes write nothing; the refusal itself is journaled
as the machine's own fact and counted on `state.md`; the document
stays in `inbox.yaml`, the author fixes the named place and runs
again. Example: `behaviors[0].rows[1].expect: the expect is empty —
name rc, out or files (law:document)`. → validation.

**Registry** — the corpus's account of the wish: the requirements
(each with its wish-line spans), the markers and the declared
artifacts. Merged by the submission laws; digest-whole into the
verdict's → registry digest. → brief map.

**Registry digest** — the sha256 over the canonical bytes of the
requirements and markers together — the wish's whole account sealed
in one digest and carried by the verdict record.

**Report** — `report.md`: the last fact-changing event's whole
picture — for a run, the build's line, the green count and every red
row grouped by its broken reason, plus the verdict's notes when the
completion act spoke; for a refusal, one line per refusal and the
fix pointer. Written whole when it moves; a run that changed nothing
and closed nothing leaves the previous report standing.
→ refusal, → verdict, → views.

**Requirement** — one demand the wish makes: an id (`R` plus
letters, digits and dots — the tail may be empty, say `R1`) and the
wish lines it occupies (W ranges). Merges by id; retires only
through an explicit `remove` that also clears every cover of it. The
wish is accounted whole when every line
sits inside some requirement's span or behind a marker.

**Row** — one executable check inside a behavior: a single command
(`run` — a shell line or an argv list) and its → expect. The
command's canonical bytes are the row's → run key and part of its →
proof; the row's place in the behavior is its journal key
(`<behavior id>/<index>`). Rows are all a behavior proves.

**Run key** — the command's canonical bytes, the row's merge key: a
submitted run that stands replaces that whole row in place; a new
run appends; unlisted rows stand. A re-spelled command is a new run —
the standing one stands. Example: `./tool src out` and
`["./tool", "src", "out"]` are two spellings of one run.

**Run number** — the counter of fact-changing runs. The terminal
line's `run <N>`, the report's, state's, knowledge's and spec's
headers all name the same number. A refusal costs no run number — it
is counted by its own counter.

S

**Scope** — the behavior's declared dependence: root-relative globs
naming the files the behavior stands on. The covered files' digests
form the → scope fingerprint; a change inside the scope makes the
behavior's rows stale. Files covered by no scope are → uncovered
sources or declared artifacts. A behavior with rows and no scope
re-runs every run — always give rows a scope.

**Scope fingerprint** — the digest over the behavior's sorted
`(path, file-digest)` pairs — the content of its scope map. A scope
file that does not exist spells `absent`: the map names what must
stand, the fingerprint sees what stands.

**Scope glob** — the scope's spelling: a root-relative slash pattern
whose segments carry `[A-Za-z0-9-._*?[]`; a `**` segment swallows
zero or more directories, `*`/`?`/`[...]` match inside one segment.
Example: `src/**` covers everything under `src/` (and `src/` itself);
`*.md` covers the root's markdown.

**Signature (the verdict's)** — the sha256 over the corpus's
canonical bytes plus, per behavior in sorted id order, `\0 id \0
scope fingerprint`: the spec and the exact tree state it was
pronounced on, in one digest. Shown short (12 characters) on the
state page and in the verdict history.

**Spec tree** — the view directory `spec/`: `index.md` plus one small
page per living behavior (the id and name, the run number, the prose
verbatim, the scope, and per row the command, the expectation and
the proof standing with its short digest). Retired behaviors' pages
are removed — the tree is an image of the standing corpus, nothing
else.

**Stale** — a row's standing: its → proof moved (its bytes, its
behavior's scope files, or the environment) — or it never ran — or
its behavior declares no scope. Stale rows execute on the next run;
the instruction counts them (`N row(s) stale — run again to execute
them`).

**Standing** — the present rebuilt from the journal by the → fold:
the goal, the corpus version and build, the registry, the behaviors
and their order, the row facts, the run number, the verdict, the
submissions, the refused and completed counts, the lessons.
Everything the machine decides, it decides on the standing — and the
standing is always derivable from the records alone.

**State page** — `state.md`: the standing on one page — the goal and
the wish's address, the brief map with every marker and the
uncovered lines, the corpus version, the three counters, the machine
at birth, the verdict with its scopes-moved mark and its history,
one line per behavior, the declared artifacts and the uncovered
sources, the knowledge counts, and `next:` — the same text as the
terminal line. → standing, → views.

**Submissions** — the accepted documents, counted on `state.md` beside
the refused and the completed — three separate economies: documents
the slot took, submissions it refused, completion acts that finished
the work.

T

**Tags** — the notes' mechanical grouping: the intersection of a
note's own words with the instance's → vocabulary (the wish's and
the standing corpus's words — lowercase runs of letters and digits,
two characters or longer). Each note carries its tags on its line;
the index maps every tag to its note numbers. Grouping only — never
a check; no semantics anywhere.

**Template** — `template.yaml`, generated from the canon at the
birth: the document's shape with teaching comments. Copy, edit,
submit as `inbox.yaml`. A document that decodes but breaks a rule is
still refused — the template teaches the shape, the validation holds
the law.

**Terminal line** — the call's one line of output on a successful
run: `punchtape: run <N> — <instruction>`. The instruction names the
file to touch and is the same text as `state.md`'s `next:` line. The
exceptions: first contact prints the whole contract; the
no-instance state prints its one instruction line; a call that
cannot happen answers its single error line instead — exit 1.
Everything else the machine has to say lives in the zone's files.

**Torn tail** — a journal record cut mid-write (a crash, a lost
power). It does not blind the machine: the intact records load, the
torn bytes quarantine to a sidecar (`<epoch>.torn`), and the epoch
file is rebuilt byte-for-byte from its intact records. A chunk that
does not parse outside the final position is damage by hand — an
honest error. → epoch.

**Tree walk** — the machine's one way of seeing the working tree: a
single ordered walk that skips the zone (`.punchtape`) at the
working root and the version-control bookkeeping directories
(`.git`, `.hg`, `.svn`) by name wherever they stand, collecting
root-relative file paths. The walk feeds the scope
fingerprints and the uncovered-files split. The machine never reads
the tree's content as code — only as bytes to digest and commands to
run.

U

**Uncovered sources** — the tree files no behavior's scope covers and
no artifact glob declares (the AGENTS block's own file aside). Named
one per line on `state.md`; a single uncovered source holds the
verdict (`not yet — the sources … stand uncovered`). The machine's
gaps are visible facts: cover them with a scope or declare them.

V

**Validation** — the whole rule set a document passes before
acceptance: id alphabets and uniqueness, the W-range grammar against
the wish's universe, marker reasons and verbatim quotes, legal globs
and root-relative paths, covers naming standing requirements, rows
carrying expectations, removes naming standing ids — plus the
dangling-cover check on the behaviors that survive unlisted. Every
refusal names the field, the fact and the law id. → refusal.

**Verdict** — the machine's pronouncement of done, recorded only on
the → completion act and only when every gate stands: the corpus
non-empty; every row green on the current fingerprints; the wish
accounted whole (every line inside a requirement span or marked); no
orphan requirement; no uncovered source. The record carries the
number, the corpus version, the → signature, the → registry digest,
every behavior's scope pointer and the closing run number; a
completion record lands beside it. Append-only — a new verdict never
rewrites an old one. Anything short is `not yet` plus one line per
class of hole in the report.

**Verdict history** — the state page's derived chronicle: every
recorded verdict (run, number, version, short signature); every
→ drift (a run with red rows under a standing verdict, with the red
keys); every heal (the first later run with no red row). What the
journal does not carry, the history does not say.

**Version flag** — `punchtape --version`, exactly this spelling and
alone on the line: prints `punchtape <version>` and exits 0 before
any state is touched — the one human affordance; the contract never
mentions it. → the call, → exit code.

**Views** — the zone's generated files: `state.md`, `report.md`,
`knowledge.md`, the spec tree, `corpus.yaml`. Images of the journal
— rewritten whole when their content changes, never touched when it
does not; a view that cannot land is a loud failure (a stale face is
a lie).

**Vocabulary** — the instance's own word set: every word of the
wish's bytes and of the standing corpus's canonical bytes — a
lowercase run of ASCII letters and digits, two characters or longer.
A note's → tags are the mechanical intersection of its words with
this set; nothing semantic anywhere — a word the instance never
carried answers nothing.

W

**W range** — the one spelling of wish lines: `W<n>` or
`W<a>..W<b>` ascending, comma-joined (`W1..W4, W9`). The numbers
name the wish's non-empty lines in file order, fixed at the birth
forever; requirements' brief spans and markers' lines ride this
grammar, and the machine spells number sets back in it with
consecutive runs collapsed (`W19, W23..W25`). → wish, → requirement,
→ marker.

**Wish** — the human's want, verbatim, placed into `wish.md` before
the birth. The birth fixes it forever as the → goal and numbers its
non-empty lines W1…Wm in file order — the numbers never change, every
requirement span and marker quotes them. The wish's size decides the
corpus's size — not taste.

**Wish coverage** — the arithmetic over the wish's universe: which
lines some requirement's span or some marker's span claims, and which
stand as → holes. One function feeds the state page's brief map and
the verdict's wish gate — never two truths.

**Wish slot** — `wish.md`, the pre-birth input file. Present with no
journal — the next run births the instance and consumes the slot;
absent — the machine answers its one instruction line and creates
nothing. The slot is dead after the birth: the goal is fixed, and
nothing else ever creates an instance.

**Working root / working tree** — the directory carrying the zone
(the call finds it at or above its starting point) and everything
beside it. The tree is the agent's territory: the machine walks it,
digests it, runs commands in it — and never reads its content as
code.

Z

**Zone** — `.punchtape/`, the only directory the machine owns: the
slots (`wish.md`, `inbox.yaml`, `env`), the generated standing files
(`corpus.yaml`, `state.md`, `report.md`, `knowledge.md`, `spec/`), the
journal with its objects, and the machine's own `cache/` and `tmp/`.
Never edited by hand; one record stands outside it — the → AGENTS
block in the working root.
