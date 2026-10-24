# Features
**English** · [Русский](FEATURES.ru.md) · [中文](FEATURES.zh.md)

## Contents

- [The life-changers](#the-life-changers)
  - [You never operate it — you just keep talking to your agent](#you-never-operate-it--you-just-keep-talking-to-your-agent)
  - [The machine runs the agent — from first line to sealed finish](#the-machine-runs-the-agent--from-first-line-to-sealed-finish)
  - [The checking comes off you — mechanically, every time](#the-checking-comes-off-you--mechanically-every-time)
  - ["Done" provably means your whole task — line by line](#done-provably-means-your-whole-task--line-by-line)
  - [One spec — any implementation, and a readable manual for free](#one-spec--any-implementation-and-a-readable-manual-for-free)
  - [The economics are measured — and the savings are your hours](#the-economics-are-measured--and-the-savings-are-your-hours)
  - [The same input always gives the same answer](#the-same-input-always-gives-the-same-answer)
  - [Total memory — across sessions, models and months](#total-memory--across-sessions-models-and-months)
  - [Changes are cheap — and re-checked automatically](#changes-are-cheap--and-re-checked-automatically)
  - [If it breaks after the finish — you know](#if-it-breaks-after-the-finish--you-know)
  - [It runs entirely on your machine — no cloud, no account, no telemetry](#it-runs-entirely-on-your-machine--no-cloud-no-account-no-telemetry)
- [The strong helpers](#the-strong-helpers)
  - [Works with whatever you build](#works-with-whatever-you-build)
  - [Failures arrive as repair manuals, not riddles](#failures-arrive-as-repair-manuals-not-riddles)
  - [It refuses to test a broken build](#it-refuses-to-test-a-broken-build)
  - [A hung command can't take your project hostage](#a-hung-command-cant-take-your-project-hostage)
  - [You always see what is NOT being checked](#you-always-see-what-is-not-being-checked)
  - [The conditions are written down — "it worked on my machine" is dead](#the-conditions-are-written-down--it-worked-on-my-machine-is-dead)
  - [Mistakes in the paperwork point at the exact spot](#mistakes-in-the-paperwork-point-at-the-exact-spot)
  - [Small changes cost small messages](#small-changes-cost-small-messages)
  - [The machine polices itself](#the-machine-polices-itself)
- [The nice little things](#the-nice-little-things)
  - [It explains itself — in three languages](#it-explains-itself--in-three-languages)
  - [The pages you read are always true](#the-pages-you-read-are-always-true)
  - [An honest yes/no for your scripts](#an-honest-yesno-for-your-scripts)
  - [One word is the whole interface — for the agent](#one-word-is-the-whole-interface--for-the-agent)
  - [The finish gesture is beautifully dumb](#the-finish-gesture-is-beautifully-dumb)
  - [No daemon, no watchers, zero footprint between runs](#no-daemon-no-watchers-zero-footprint-between-runs)
  - [The truth is a folder](#the-truth-is-a-folder)
  - [One line installs it — and it's free](#one-line-installs-it--and-its-free)

You talk to your AI assistant — and **punchtape runs the agent.**
From the moment your task is born, a second party takes the wheel: a
foreman that never sleeps, never forgets and never takes the
worker's word for anything. It hands the agent its next step at
every turn, measures every claim by actually running it, and refuses
to let the work end until it is genuinely finished — start to
sealed finish, with no way out through charm, fatigue or "trust me".
You stop being the quality department of your own AI.

The features are ordered by size: the life-changers first, the
strong helpers next, the nice little things last. The numbers are
measured facts from the project's field studies, labeled honestly;
everything else is what the shipped machine does.

## The life-changers

### You never operate it — you just keep talking to your agent

There is nothing to learn, because none of it is yours to operate.
Paste one sentence into your assistant once — from then on your
prompts are the whole interface: you write what you want in plain
words, and the agent does the rest, machine included. The machine
even leaves a short note in your project's instructions file, so
every fresh agent session discovers it without a word from you.
**Zero new habits, zero new commands: the machine arrives like
electricity — you never operate it, you just use what comes out of
the socket.**

### The machine runs the agent — from first line to sealed finish

Left alone, an AI grades its own homework: it writes, declares
success, moves on. punchtape inverts the roles — **the AI becomes
the worker, the machine becomes the foreman, and the foreman runs
the site.** At every point the machine hands out exactly one next
step (one line: what to do now, which file to touch); all work
comes in through one door; every claim gets measured by running it,
not by asking; sloppy paperwork goes back with the exact spot
named. And the loop — next step, work, check, next step — simply
does not end on the worker's say-so: the machine keeps driving until
the job is truly done, then seals the finish itself. You get two
parties with clean jobs: a brilliant, occasionally sloppy worker,
and a merciless, tireless inspector. **Your prompt is the only
lever you'll ever need to pull.**

### The checking comes off you — mechanically, every time

The assistant always says "done"; finding out whether that's true
has been your evening — every task, every edit, forever. Now, next
to your task text, the AI writes down how each part can be checked
(real commands, expected results), and the machine executes those
commands itself, in your actual project, every time. It never asks
the AI whether the work works — it measures. **Between your task
and the recorded finish there are zero mandatory points of your
attention: write it, walk away, come back to the result.**

### "Done" provably means your whole task — line by line

You ask for ten things, the delivery shows seven, and the missing
three are found by your customers. Not here: your original task
text is kept untouched — forever — and every single line of it must
be accounted for. Either a check proves it done, or it is crossed
out openly, with its own words quoted and a reason beside them.
Anything unaccounted for? No finish — the machine names exactly
what's missing. Skipping a demand is still possible, but only out
loud, in writing, in the AI's own report. When everything passes,
the finish is sealed with a digital stamp: which checks, on which
version of your code. **A "done" you can hand to your boss, your
customer, or yourself six months later.**

### One spec — any implementation, and a readable manual for free

Specifications usually rot three ways: they die with the chat
session, they die with the stack, and the docs drift from what was
built. Here everything the AI hands in lands as one machine-readable
spec — each feature with its description in the author's words, the
files it stands on, and the checks that prove it — and that one
document compiles two ways. Into code: the spec pins observable
behavior, not implementation (the machine is mechanically forbidden
from knowing any language or framework), so any AI in any stack can
build it — rewrite the product in another language next year, and
the same checks verify it the same way, the green results traveling
with the spec as its proof. Into prose: the machine itself renders a
page per feature — description verbatim, checks, current status —
plus a one-page status of the whole job, regenerated on every run,
never hand-written, never stale. **The spec is the source; code is
one translation of it; the readable manual is another — and the
manual is free.**

### The economics are measured — and the savings are your hours

Every tool promises savings; this one shows the receipt. Three
generations of field studies — real tasks, two camps (with and
without the machine), blind judges — measured the same article every
time: the verifying work, the part that used to be yours. Without
the machine, every single task needed a human-style acceptance;
through the machine it fired only on exceptions — 7 of 40 cells in
the first generation, and **2 times in 25 tasks against 32 times
without it** in the last. The checking article shrank **×60**, then
**×17.6 in tokens / ×9.2 in time**, then **×24.0 / ×5.5** — the one
cost that never flipped sign in any wave, task or stack. Counted in
the currency that matters (a labeled estimate for a live human
accepting the work): **25 tasks eat 4.4–7.3 hours of your attention
without the machine — 0.8–1.3 with it, effectively 0.2–0.5 counting
only real exceptions: 3.6–6.0 hours handed back per 25 tasks.** The
all-in cost of a task ran **5×–19× higher** without it (measured,
at the runs' recorded prices). The finishes came fast — **every
machine task accepted finished on the first try (25/25)** in the
five-wave battery, **all 50 cells sealed** in the last series — and
the loop stayed cheap to feed (**99% prompt-cache hit rate through
the machine against 58% without it**; the machine's own work:
0.5–37.8 seconds a task). Quality didn't pay for any of it: blind
judges found **0 missing items out of 1720 checks** in both camps.
**Token prices fall every year; your hour doesn't. The machine
spends the cheap resource freely to save the expensive one — your
time.**

### The same input always gives the same answer

No moods, no platforms, no clock time: the same task text, the same
code, the same conditions — the same answer, on your laptop, on
the server, next year. Records carry no timestamps; every write is
atomic; a history file copied between machines is byte-identical.
If the machine dies mid-write, the next run repairs the file and
keeps every intact record — and nobody, not the AI and not you by
accident, can edit the past: history only grows. **Facts that don't
decay: a result that was green in March means the same thing in
October.**

### Total memory — across sessions, models and months

The decisions, the fixes, the history — all of it used to live in a
chat that gets closed and forgotten, so every new session re-broke
the same things. Now every step is written into files inside your
project, in a form that cannot be edited afterwards: what was asked,
what was delivered, what failed, how it got fixed. The machine also
saves the lessons by itself — what broke, and the exact change that
fixed it — measured, not typed. **Swap the AI model, close the
session, come back in a year: the foreman remembers everything, and
the work continues exactly where it stopped.**

### Changes are cheap — and re-checked automatically

After delivery, every edit used to be a gamble, because nobody
re-checks everything by hand. The machine remembers what each check
was run against — which files, which conditions — so when you change
one file, it re-runs only the checks that depend on that file and
touches nothing else. The first full check is the slow one; every
check after it costs seconds. **Editing stops being scary — so
editing finally happens — and the machine measures the blast radius
for you every time.**

### If it breaks after the finish — you know

Last month's "done" quietly becomes a lie as the code keeps moving.
Not quietly here: after a finish is sealed, any later change that
breaks something shows up the next time you look — which change
broke it, and when it got fixed again. No watchers, no background
spies: the machine compares what was sealed with what stands now,
every time it runs. **Old promises don't rot in silence — either it
still works, or the machine names the moment it stopped.**

### It runs entirely on your machine — no cloud, no account, no telemetry

Development tools that phone home, demand accounts or quietly
upload your code make your project somebody else's business. This
one is a single small program working only with local files and
local commands — its code contains no network calls at all. No
login, no key, no usage reports; it doesn't even read your code as
code, it just runs your commands and compares bytes. **An inspector
you can run on a laptop in a locked basement — offline, off-grid,
on secret code, because there is nowhere for anything to be sent.**

## The strong helpers

### Works with whatever you build

The machine knows nothing about programming languages, frameworks
or stacks — by design, enforced mechanically at build time: not one
word of any technology may appear in its code. If your project can
be built and run by a command, the machine can verify it — no
prescribed structure, no framework tax. **Your project's shape stays
your business.**

### Failures arrive as repair manuals, not riddles

When a check fails, you see both sides — what was expected and what
actually happened ("expected the program to end cleanly, got error
1"; expected the text "ok", got "okk") — with identical failures
grouped into one list, and a missing tool explained plainly: where
the machine looked and the three ways to fix it. The agent reads
this and usually fixes everything in one pass. **You are not the
debugger of last resort — the failure report is.**

### It refuses to test a broken build

If your project has a build step, it runs first. If it fails or
hangs, nothing else runs that round and the machine says: fix the
build first. **One clear problem at a time — and no fake "everything
passed" on top of a broken foundation.**

### A hung command can't take your project hostage

Every command gets a hard 300-second limit: a hang is killed and
recorded as a failure with its reason, never wedging the run. And
whatever a command left running — a server, a watcher, a stray
worker — is cleaned up together with it. **You never come back to a
project bricked by a forgotten background process.**

### You always see what is NOT being checked

Every project file no check is watching is listed by name on the
status page; you choose to cover it or accept the gap openly, and an
unwatched file blocks the final finish. **No invisible blind spots:
what's checked is written down, and what isn't checked is written
down too.**

### The conditions are written down — "it worked on my machine" is dead

Every check runs under the exact same written-down conditions — same
paths, same settings, plus whatever you fix once inside the project
so it travels with the code. If the conditions change, the machine
notices and re-checks instead of showing you a stale "passed"
sticker. **A green result is a fact about known conditions — and
setup drift reaches you from the machine, not from a customer.**

### Mistakes in the paperwork point at the exact spot

If the AI's report file has an error, the machine names the exact
spot and what's wrong, one line per mistake; the file stays put —
fix the named place and run again, nothing lost. Anything sneaky
hidden in the file format is refused flat-out, unread, before any
state is touched. **One pass to fix, not three — and an input
channel too boring to attack.**

### Small changes cost small messages

You send only what changed; anything unmentioned stays exactly as it
was; re-formatting counts for nothing, because the machine reads
what the text says, not how it's laid out. **Ten little fixes cost
ten little messages — never ten full rewrites of the spec.**

### The machine polices itself

Who guards the guards? Every time the machine is built, its own
built-in example task must pass through the real machine — the full
cycle, from task text to sealed finish — or the build fails, and the
same build enforces that every line of the machine's code is
actually executed by its own tests. **The thing that checks your
work is itself checked, mechanically, on every build.**

## The nice little things

### It explains itself — in three languages

Run it in an empty folder and it prints its full manual, including
the exact shape of the file the AI writes, and creates a fill-in
template. English, Russian, Chinese. **The instructions find the
agent, not the other way around.**

### The pages you read are always true

The status page and the reports are rebuilt from the records on
every run — never edited by hand, never stale — and if a page can't
be written, the machine raises an alarm instead of quietly showing
an old one. The "what to do next" line is always identical in the
terminal and on the page. **If it's on disk, it's current.**

### An honest yes/no for your scripts

When the machine says it worked (exit 0), it worked — problems found
inside the work are data, not crashes. When it says the call didn't
happen (exit 1), it didn't. **Wrap it in any script and trust the
answer.**

### One word is the whole interface — for the agent

No flags, no menus, no modes: the agent runs one word and gets one
line back — the next action, pointing at the file to touch. That's
also why the machine can't be misused by a model that hallucinates
a flag: **there are no flags to hallucinate.**

### The finish gesture is beautifully dumb

How does the AI formally say "it's done"? It copies the machine's
own list of what should be true and hands it back unchanged —
sending the list back as-is IS the signature. **Nothing to invent,
nothing to format, nothing to get wrong on the last mile.**

### No daemon, no watchers, zero footprint between runs

The machine exists only in the moment it runs: no background
process, no service eating memory, nothing watching your files —
change detection is arithmetic at run time, not a resident spy.
**Between runs, your machine doesn't even know it's there.**

### The truth is a folder

Everything — the task, every report the AI handed in, every run's
results, every finish — lives in plain files in one hidden folder
inside your project. Copy it to another machine and the state is
byte-identical; back it up, diff it, put it in version control. **No
database, no exports, no migration.**

### One line installs it — and it's free

Paste one sentence to your agent: it fetches the ready-made binary
for your platform (Linux and macOS, x64 and ARM, checksums
published) or builds from source, and puts it on your path — free,
MIT-licensed, no account, no payment wall. **From zero to first
sealed finish in one session.**
