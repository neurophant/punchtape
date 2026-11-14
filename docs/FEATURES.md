# Features
**English** · [Русский](FEATURES.ru.md) · [中文](FEATURES.zh.md)

You talk to your AI assistant — and **punchtape drives the agent.**
From the moment your task is born, a second party keeps the task's
record: it hands the agent its next step at every turn, measures
every claim by running it, and refuses to let the work end until
it is genuinely finished. You stop being the quality department
of your own AI.

The features are ordered by size: the largest first, the supporting
next, the small conveniences last. The numbers are
measured facts from the project's field studies, labeled honestly;
everything else is what the shipped machine does.

Five words you will meet below, in plain terms. The **zone** is
the `.punchtape/` folder the machine owns inside your project. A
**row** is one check: a command plus the exact output it must
print. A **proof** is the fingerprint showing a row still holds.
**Pins** are the exact bytes a row waits for. The **corpus** is
the standing set of behaviors and their checks.

## The largest features

### The agent operates the machine; you write prompts

There is nothing to learn, because none of it is yours to operate.
Paste one sentence into your assistant once — from then on your
prompts are the whole interface: you write what you want in plain
words, and the agent does the rest, machine included. The machine
also leaves a short note in your project's `AGENTS.md`, so every
fresh agent session discovers it without a word from you. You
gain: zero new habits, zero new commands.

### The machine drives the work loop

Left alone, an AI grades its own homework: it writes, declares
success, moves on. punchtape splits the roles. The machine hands
the agent exactly one next step at a time — one line saying what
to do now and which file to touch. Everything the agent hands in
goes through one door, and every claim is measured by running it,
not by asking; sloppy paperwork goes back with the exact spot
named. The loop ends on the verdict, not on the worker's say-so —
you get a worker and an inspector, each doing its own job.

### Checks run without you

The assistant always says "done"; finding out whether that's true
has been your evening, task after task, edit after edit. Next to
the code, the AI writes down how each part can be checked (real
commands, the exact bytes they must print), and the machine
executes those commands itself, in your actual project, every time
they go stale. It never asks the AI whether the work works — it
measures. Once the task is written, nothing in between needs your
attention: walk away, come back to the result.

### Every word resolved, every line claimed

You ask for ten things, the delivery shows seven, and the missing
three are found by your customers. Not here: your words are kept
in a file the machine itself owns and numbers — every non-empty
line of them, D1 to Dn. Each line must be given its declared
outcome when it arrives: work, answer, execution, description, a
mark of "not a requirement", or an open question to you. An open
question holds the finish until you answer it in your own words; a
dismissal must be loud, in writing, with a reason. On the other
side, every line of your project must be either claimed by a check
or crossed out openly with a reason that cites your words or the
machine's law. Anything unresolved, unclaimed or unanswered → no
finish; the machine names exactly what is missing. When everything
passes, the finish is sealed with a signature over which checks
passed on which version of your corpus — a "done" you can hand to
your boss, your customer, or yourself six months later.

### One spec, any implementation — and a readable manual for free

Specifications usually rot three ways: they die with the chat
session, they die with the stack, and the docs drift from what was
built. Here everything the AI hands in lands as one machine-held
corpus — each behavior with its description in the author's words,
the exact lines of code it stands on, and the rows that prove it —
and that one document compiles two ways. Into checks: the rows pin
observable behavior (a command and its exact bytes), not
implementation — the machine is mechanically forbidden from
knowing any language or framework — so any AI in any stack can
re-earn the same proofs; rewrite the product in another language
next year, and the same rows verify it, with the claims re-drawn
onto the new files. Into prose: the machine itself renders a page
per behavior — description verbatim, claimed lines with their
digests, commands and expectations, current standing — plus a
one-page status of the whole job, regenerated on every run, never
hand-written, never stale. The spec is the source; code is one
translation of it; the readable manual is another — and the manual
is free.

### Measured economics

Every tool promises savings; this one shows the receipt. Three
generations of field studies — real tasks, two camps (with and
without the machine), blind judges — measured the same article
every time: the verifying work, the part that used to be yours.
Without the machine, every single task needed a human-style
acceptance; through the machine it fired only on exceptions — 7
of 40 cells in the first generation, and **2 times in 25 tasks
against 32 times without it** in the last. The checking article
shrank **×60**, then **×17.6 in tokens / ×9.2 in time**, then
**×24.0 / ×5.5** — the one cost that never flipped sign in any
wave, task or stack. Counted in the currency that matters (a
labeled estimate for a live human accepting the work): **25 tasks
eat 4.4–7.3 hours of your attention without the machine — 0.8–1.3
with it, effectively 0.2–0.5 counting only real exceptions: 3.6–6.0
hours handed back per 25 tasks.** The all-in cost of a task ran
**5×–19× higher** without it (measured, at the runs' recorded
prices). The finishes came fast — **every machine task accepted
finished on the first try (25/25)** in the five-wave battery,
**all 50 cells sealed** in the last series — and the loop stayed
cheap to feed (**99% prompt-cache hit rate through the machine
against 58% without it**; the machine's own work: 0.5–37.8 seconds
a task). Quality didn't pay for any of it: blind judges found
**0 missing items out of 1720 checks** in both camps. Token
prices fall every year; your hour doesn't — the machine spends
the cheap resource to save the expensive one.

### Deterministic runs

No moods, no platforms, no clock time: the same task text, the
same code, the same conditions — the same answer, on your laptop,
on the server, next year. Records carry no timestamps; every write
is atomic; a history file copied between machines is
byte-identical. If the machine dies mid-write, the next run
quarantines the torn tail and keeps every intact record — and
nobody, not the AI and not you by accident, can edit the past:
history only grows, and a touched past line of your words stops
the machine cold. A result that was green in March means the same
thing in October.

### History in files — across sessions, models and months

The decisions, the fixes, the history — all of it used to live in
a chat that gets closed and forgotten, so every new session
re-broke the same things. Now every step is written into files
inside your project, in a form that cannot be edited afterwards:
what was asked, what was delivered, what failed, how it got fixed.
The machine also saves the lessons by itself — each red row it
later saw green, with the row's bytes before and after and the two
run numbers — measured, not typed: zero tokens of any model. Swap
the AI model, close the session, come back in a year: the record
remembers everything, and the work continues exactly where it
stopped.

### Incremental re-runs

After delivery, every edit used to be a gamble, because nobody
re-checks everything by hand. The machine remembers what each
check was run against — its own bytes, the exact claimed lines of
its behavior, and the whole environment — so when you change one
file, it re-runs only the checks that claim lines of that file and
touches nothing else. The first full check is the slow one; every
check after it costs seconds. The machine measures the blast
radius of every change for you.

### Post-finish breaks are caught

Last month's "done" quietly becomes a lie as the code keeps
moving. Not quietly here: after a finish is sealed, any later
change that breaks a claimed line shows up the next time you look
— the run that caught the drift and the run that went green again
are both written into the verdict history, and the standing
verdict itself says `open (drift since run N)` until a submission
re-earns it. No watchers, no background spies: the machine
compares what was sealed with what stands now, every time it runs.
Either it still works, or the machine names the moment it stopped.

### Offline, no account, no telemetry

Development tools that phone home, demand accounts or quietly
upload your code make your project somebody else's business. The
machine is one binary working only with local files and
local commands — its code contains no network calls at all. No
login, no key, no usage reports; it doesn't even read your code as
code, it runs your commands and compares bytes. It runs on a
laptop in a locked basement, offline, on secret code — there is
nowhere for anything to be sent.

## The supporting features

### Stack-agnostic by build law

The machine knows nothing about programming languages, frameworks
or stacks — by design, enforced mechanically at build time: not
one word of any technology may appear in its code, and a scan over
the machine's own sources fails the build the moment such a word
lands. If your project can be built and run by a command, the
machine can verify it — no prescribed structure, no framework tax;
your project's shape stays your business.

### Failures report want and got, in bytes

When a check fails, you see both sides — what was expected and
what actually happened (`exit: got 1, want 0`; expected the exact
bytes `ok\n`, got `okk\n`) — with identical failures grouped into
one list under their reason, and a missing tool explained plainly:
the row's environment is the PATH and HOME of the shell that runs
the machine plus the zone's pins — make the command reachable,
pin it, or call project files by relative path. The agent reads
this and usually fixes everything in one pass; the report is the
debugger, not you.

### The build runs first

If your project declares a build command, it runs first — on
every launch, before any row. If it fails or hangs, nothing else
runs that round and the machine says: fix the build first. One
clear problem at a time, and no fake "everything passed" on top of
a broken foundation.

### Command deadlines and process-group death

Every command gets a hard deadline: 300 seconds by default, and a
zone can raise it by one pinned line (`PUNCHTAPE_ROW_TIMEOUT`
names whole seconds). A hang is killed and recorded as a red fact
with its reason, never wedging the run. And whatever a command
left running — a server, a watcher, a stray worker — dies with its
process group; you never come back to a project bricked by a
forgotten background process.

### Uncovered lines are listed

Every project line no check claims and no loud exclusion covers is
listed by name — per file, with line ranges — on the status page
(the first ten files) and in a full uncapped ledger of its own;
you cover it with a claim or accept the gap openly, and an
uncovered line blocks the next finish. The machine even drafts the
exclusion list for you from its own hole arithmetic — you add the
reasons, not the paths. What is checked is written down, and what
is not checked is written down too.

### The environment is pinned inside every proof

Every check runs under the exact same written-down conditions: the
PATH and HOME of the invoking shell plus the project's own pinned
values (`.punchtape/env`, one KEY=VALUE a line — relative paths
only, so the pins travel with the tree). The machine's own
version, host and the environment actually passed enter every
proof's fingerprint: if any of it moves, the machine notices and
re-checks instead of showing you a stale "passed" sticker. A green
result is a fact about known conditions — you hear about setup
drift from the machine, not from a customer.

### Refusals name the exact place

If the AI's submission file has an error, the machine names the
exact field, the fact and the law it broke, one line per mistake;
the file stays put — fix the named place and run again, nothing
lost. Anything sneaky hidden in the file format — anchors,
aliases, tags, merge keys, duplicate keys, unknown fields, tabs in
indentation — is refused at the door, never interpreted: before
anything runs, the machine walks the file's structure itself and
throws out every trick the format allows. One pass to fix, and an
input channel too strict for tricks.

### Canonical-byte merges

You send only what changed; anything unmentioned stays exactly as
it was; re-formatting counts for nothing, because the machine
compares canonical bytes after decoding — two spellings of one
content are one content. Rows merge by their command: the same
command replaces its row, a new one is appended, unlisted rows
stand. Ten little fixes cost ten little messages, never ten full
rewrites of the spec.

### Quiet intents lock the tree

Your words can ask different things: some order work, some just
ask a question or want a fact reported back. The machine takes the
difference seriously — a submission that only answers, executes or
describes locks the tree for the run: the digest is measured
before and after, and a moved tree is a journaled fact that holds
the verdict until the bytes return or the change is submitted as
declared work. The agent cannot "helpfully" rewrite your project
while answering a question and call it a day: quiet intents leave
the world exactly as they found it — or say so loudly.

### The build tests the machine itself

Who guards the guards? Every time the machine is built, its own
built-in example task must pass through the real machine — the
full cycle, from first words to recorded verdict — or the build
fails; the same build enforces that every line of the machine's
code is actually executed by its own tests, and that the contract
it prints and the rules it enforces are the same set. The thing
that checks your work is itself checked, mechanically, on every
build.

## The small conveniences

### The manual ships inside the zone

Run it in an empty folder and it prints its whole law — one page:
how to start, what the files are, the eight laws of the machine.
The birth then writes the same contract into the zone, beside a
template with the document's exact shape and a reference page that
names every door of the machine — so the full manual lives in your
project, always at hand, and a zone can never teach a law its
machine has left. The user documentation ships in English, Russian
and Chinese; the instructions find the agent, not the other way
around.

### Views rebuild on every run

The status page and the reports are rebuilt from the records on
every run — never edited by hand, never stale — and if a page
cannot be written, the machine raises an alarm instead of quietly
showing an old one. The "what to do next" line is always identical
in the terminal and on the page: if it is on disk, it is current.

### Honest exit codes

When the machine says it worked (exit 0), it worked — problems
found inside the work are data, not crashes. When it says the call
didn't happen (exit 1), it didn't. Wrap it in any script and trust
the answer.

### One command, no flags

No flags, no menus, no modes: the agent runs one word and gets one
line back — the next action, pointing at the file to touch. The
single exception is `--version`, a human affordance the contract
never mentions. That is also why the machine cannot be misused by
a model that hallucinates a flag: there are no flags to
hallucinate.

### `done: true` is the whole sign-off

How does the AI formally say "it's done"? It writes one line to
the inbox — `done: true` and nothing else. That bare line is the
sign-off; the machine then re-measures everything itself. Nothing
to invent, nothing to format, nothing to get wrong on the last
mile.

### Zero footprint between runs

The machine exists only in the moment it runs: no background
process, no service eating memory, nothing watching your files —
change detection is arithmetic at run time, not a resident spy,
and the scratch directory is wiped at the end of every run.
Between runs, your machine doesn't even know it is there.

### The zone is plain files

Everything — the words, the submissions, each run's results, the
finishes — lives in plain files in one hidden folder inside your
project. Copy it to another machine and the state is
byte-identical; back it up, diff it, put it in version control.
No database, no exports, no migration.

### One-line install, MIT license

Paste one sentence to your agent: it fetches the ready-made binary
for your platform (Linux and macOS, x64 and ARM, checksums
published) or builds from source, and puts it on your path — free,
MIT-licensed, no account, no payment wall. From zero to first
sealed finish in one session.

## The honest limits

The machine compares real output byte by byte — it never reads
meaning. Every item below is a lived fact of the shipped machine,
stated plainly:

- **Words enter by the agent's hand.** The machine never sees your
  chat; it takes your words from `demands.md` itself and seals the
  past, but it cannot know a word that was never written down —
  and a submission that brought no new words signs "no new words",
  a signature that stands in the journal forever, visible on the
  state page.
- **A green verdict is formal integrity, not correctness.** Every
  line resolved, every row green with pinned bytes, the tree equal
  to claims plus loud exclusions — the machine cannot read whether
  the code does what you meant; your one glance at the spec pages
  is the audit, and they are built for exactly that glance.
- **The document is YAML, and YAML bites prose.** A value with
  ": " or a comma takes quotes; a multi-line pin stays one line;
  tabs in indentation are refused. In the latest field rehearsal
  this was the most common refusal (24 of 56) — each names its
  place, none loses data.
- **One zone, one machine.** A zone born on another version of
  punchtape is refused — archive it and start fresh; there is no
  migration.
- **Linux and macOS only** (amd64/arm64); rows run through the
  unix shell, and a command sees only PATH, HOME and the pins.
- **Runs serialize on a zone lock.** One pass holds the zone at a
  time (`run.lock`, kernel-released, death-safe); a parallel
  caller waits its turn — up to the deadline of the running
  pass's own commands. Local filesystems only: the network ones
  carry no such promise.
- **Several agents are not coordinated.** The lock serializes the
  runs, not the work: who submits which part is the task owner's
  call, the work wants splitting into non-overlapping parts, and
  the machine is an arbiter of facts — not a coordinator of
  agents.
- **Never edit the journal by hand.** Every record is verified on
  load; what does not recompose is refused by name.
