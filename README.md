# punchtape
**English** · [Русский](README.ru.md) · [中文](README.zh.md)

**Your AI writes the code. punchtape checks it — and does not stop
until "done" is really done.**

punchtape is a small program for your computer. Your AI assistant
runs it while it works. The program runs real checks of the AI's
work, tells the AI what to fix, and records the finish when every
check passes. You read the result — you do not check anything
yourself.

This is not another harness, and it has nothing to do with
harnesses. The harness runs your assistant's conversation;
punchtape is an external transactional state machine: one program
beside your project that owns the task's state — the words, the
checks, the facts, the recorded finishes — and drives the agent
through the work from the first line to the sealed end. Not a
plugin and not a library inside the agent: an arbiter on the
outside. Whatever environment your assistant lives in stays
exactly what it is; punchtape adds the one thing it lacks — a
party outside the conversation that owns the task's truth.

Free. MIT license. Linux and macOS. Works offline. No account.

## The numbers, measured

Three rounds of field tests — real tasks, one group working with
the program, one without, independent blind judges — published as
a paper (a preprint, not peer-reviewed; the methodology and the
raw campaign data are open):

> **"Closing the Verification Loop in Generative Development: An
> External Transactional State Machine and a Field Evaluation of Its
> Economics and Convergence"**
> — [SSRN](https://doi.org/10.2139/ssrn.7566638) ·
> [Preprints.org](https://doi.org/10.20944/preprints202610.0325.v1) ·
> [Research Square](https://doi.org/10.21203/rs.3.rs-11271483/v1)

| per 25 tasks | without punchtape | with punchtape |
|---|---|---|
| human-style acceptances | 32 | 2 |
| your verifying time (est.) | 4.4–7.3 h | 0.8–1.3 h |
| total task cost | 5–19× higher | baseline |
| tasks accepted on first submission | — | 25 of 25 |

Blind judges checked 1720 points across both groups and found 0
missing in each. All numbers, year by year:
**[docs/METRICS.md](docs/METRICS.md)**; how the tests were run:
**[docs/RESEARCH.md](docs/RESEARCH.md)**.

## The problem it solves

This is what normally happens:

1. You give a task to your AI.
2. The AI answers: "Done! All tested!"
3. You check the work yourself. This takes 30–60 minutes.
4. You find two bugs and one missing feature.
5. You send it back. Repeat tomorrow.

Step 3 is the expensive part. Not the writing — the checking. And
it is always your job. punchtape takes step 3 away: the program
runs the checks, not you, and it will not record the finish while
anything is broken, missing or silently skipped.

## How it works, concretely

- You write your task in normal words. Example: *"Make a notes
  utility: notes add saves a note…"* Your assistant puts your
  words — unchanged, one line each — into `.punchtape/demands.md`.
- Next to the code, your assistant writes a document describing
  the work: what each line of your task means (its outcome:
  work, answer, question…), which requirements the words become,
  and for each piece of behavior a check — a real command plus
  the exact output it must print.
- The program takes your words from the file itself (the agent
  never retypes them into documents), runs the checks itself, in
  your project folder, and compares bytes. A check passes or
  fails — the program sees the real result, it never asks the AI.
- If a check fails, the program tells the AI exactly what was
  expected and what actually happened. The AI fixes it. Repeat.
- When everything passes, your AI sends one line — `done: true`
  — and the program re-checks everything and records the finish:
  date-free, with a checksum, in files inside your project.
- Every task, every word, every check, every result is written
  into files in `.punchtape/` inside your project. Nothing is
  stored in any cloud. Copy the folder — the whole history moves
  with it, byte-identical.

You never learn commands. The AI runs the program; you write
prompts as always.

## What you get

- **You stop checking the AI's work.** The program checks it,
  every time, on every change.
- **Nothing from your task gets silently dropped.** Every line of
  your words must be given an outcome, and every line of your
  project must be either covered by a check's claim or crossed
  out openly, with a reason that cites your words or the law.
  Otherwise: no finish.
- **Edits are cheap.** Change one file → only the checks claiming
  lines of that file run again. Everything that still holds is
  not re-run.
- **If something breaks after the finish, you see it.** The
  program records which run broke it and which run healed it.
- **Nothing is lost between sessions.** New chat, new AI model,
  next month — open the folder and continue. All history is in
  files, and the past cannot be edited: a touched past line stops
  the machine.
- **It works with anything you build, any way you build it.** Any
  language, any framework, any project shape, any workflow, any
  harness — the machine knows none of them and needs none of
  them.

Full list with every detail: **[docs/FEATURES.md](docs/FEATURES.md)**

## How to use it

### 1. Install — once per computer

Paste this to your AI assistant — or run it yourself in a
terminal:

```text
Install punchtape and run all work through it:

  curl -fsSL https://github.com/neurophant/punchtape/releases/latest/download/get-punchtape.sh | bash
```

What this does: downloads one small program, verifies its
checksum, puts one command — `punchtape` — into `~/.local/bin`.
No admin rights, no account. You do this once per computer, never
per project. (Build from source instead: `git clone
https://github.com/neurophant/punchtape && cd punchtape &&
./install.sh` — needs Go 1.26+.)

One more line, once: if your assistant has permanent instructions
(a place where you write standing rules for it), add there "run
all work through punchtape" — and you will never have to write it
in a prompt again.

### 2. Start any new project — first one or the tenth

Open a new, empty folder. Write your task to your assistant in
normal words:

```text
Make a notes utility: notes add <text> saves a note and prints its
number; notes list prints the notes one per line, oldest first;
notes show <number> prints one note; state lives in notes.yaml and
survives a restart.
```

When to add the line "run all work through punchtape" to your
prompt — a simple rule:

- The folder already runs punchtape (it has a `.punchtape`
  folder)? → **Not needed.** The file `AGENTS.md` in that folder
  already tells every AI tool about it.
- You once put that line into your assistant's permanent
  instructions (most assistants have them)? → **Not needed,
  anywhere, ever.**
- New empty folder, and your assistant has no such permanent
  instruction? → **Add the line.**
- Not sure? → **Add the line.** It never hurts.

What happens, step by step:

1. Your assistant runs `punchtape` in the folder. The folder is
   empty, so the program prints its contract — one page with the
   whole law of the machine on it: what to do, which files to
   write, in what shape. Your assistant reads it and knows the
   entire process. You explain nothing.
2. Your assistant puts your task text — your words, unchanged,
   one line each — into `.punchtape/demands.md`. The next run
   births the zone: the contract, the template and a reference
   page — the machine's full manual, always under hand — land
   beside the words, and from that moment the past of your words
   is sealed (a touched past line is zone damage — the machine
   stops), while new words keep entering the same file at any
   moment.
3. Your assistant copies `.punchtape/template.yaml` to
   `.punchtape/inbox.yaml` and fills it in: the narrative of the
   submission, the declared outcome of every new line of your
   words, the requirements, and the checks — behaviors with their
   exact claimed lines of code and rows of commands with pinned
   expected output.
4. The program takes your words itself, merges the document,
   runs the build (when declared) and every check that went
   stale. Failed checks land in `.punchtape/report.md` — what
   was expected, what actually happened. Your assistant reads
   that file and fixes the code or the rows. One line in the
   terminal always says what to do now, for example (a real
   line):

   ```text
   punchtape: run 1 — 2 row(s) red — read .punchtape/report.md, fix, run again
   ```

5. All checks green → the terminal says so, and your assistant
   writes exactly one line to the inbox: `done: true`. That is
   the sign-off — nothing to retype, nothing to format.
6. The program re-checks everything — every row green with a
   pinned proof, your words resolved whole, every requirement
   covered, the tree equal to its claims plus loud exclusions, no
   open question — and records the verdict. From now on
   `.punchtape/state.md` shows: `verdict: #1 … — standing`, with
   its checksum. Show that file to anyone who asks if it works.

The next new project after this one is the same thing again:
empty folder, task in words, done. Each folder is separate;
nothing from one project leaks into another.

### 3. Continue a project that already runs on punchtape

Open that project folder — or any folder inside it (the program
looks upward and finds its zone). Write what you want, in words:

```text
Add an open command — show only the unclosed notes.
```

It does not matter who picks this up: your usual assistant, a new
session, a different AI model, a teammate. The file `AGENTS.md`
in the project root contains a short note from punchtape — every
AI tool reads that file on arrival — and the note points to
`.punchtape/state.md`, one page: where the work stands, what is
next. So anyone continues in seconds.

What happens with your change: your words join
`.punchtape/demands.md`; the checks claiming the changed files run again; the
checks that still hold are not re-run and cost nothing. When the
addition works, the machine records finish #2 — then #3, #4…
Each verdict is its own recorded fact with its own checksum, and
if a later change breaks a sealed result, the machine names the
run that caught it and the run that healed it.

## Real prompts for real situations

**A bug.**
`notes show 3 prints nothing when there is no note 3 — fix it and prove it cannot happen again.`
Your assistant adds a check for the wanted behavior (it fails —
the bug exists), fixes the code, the check passes, and that check
stays in the project forever: it re-runs whenever the lines it
claims change.

**An existing project that never used punchtape.**
`Bring this repo under punchtape: read the code, write the checks for the behavior that already exists, and get them all passing.`
The program measures what your code really does today and drafts
the exclusion list for the rest. From then on every change is
measured against that.

**A new AI model took over.**
`Continue this project through punchtape — read .punchtape/state.md first, then finish what is missing.`
The new model needs no old chat: the project folder holds
everything.

**Hand the project to a teammate.**
Copy the folder. Done. Words, checks, results, history, finishes —
everything is inside, byte-identical.

## Honest limits

The machine measures bytes and sets — never meaning. A recorded
verdict proves formal integrity: every line of your words
resolved, every row green with its pinned bytes, the tree equal
to claims plus loud exclusions. It does not prove the code does
what you meant — that last look stays yours (the spec pages put
the claim and the check side by side for exactly this).

- The machine never sees your chat. Words enter
  `.punchtape/demands.md` by the agent's hand; the machine takes them from
  the file itself and seals the past — but it cannot know a word
  that was never written down. A submission that brought no new
  words signs "no new words"; a false signature stands in the
  journal forever, visible on the state page.
- The document is YAML, written by the agent: a prose value
  holding ": " or a comma takes quotes; a multi-line pin stays
  one line; tabs in indentation are refused. In the latest field
  rehearsal this was the most common refusal — each names its
  place, none loses data.
- One zone, one machine: a zone born by another version of
  punchtape is refused — archive it and start fresh.
- Linux and macOS only; rows run through the unix shell, and a
  command sees only `PATH`, `HOME` and the zone's pinned values.
- Two machines on one living zone at the same moment are not
  coordinated (the birth itself elects one winner among parallel
  runs).
- Never edit `.punchtape/journal/` by hand — the machine verifies
  every record and refuses what does not recompose.

## Is it for you?

- Works with any language and framework — the program knows none
  of them and does not need to. Stack, libraries, project shape,
  workflow, harness, approach — use whatever you want, however
  you are used to, as the task requires.
- Linux and macOS today. Windows is not supported.
- You need an AI assistant — the program never writes code
  itself.
- For a one-time throwaway script — too much. For anything you
  will edit again — this is where it pays.

## Free, private, offline

MIT license. One program, one command. No background processes —
it runs only when called, and whatever a command left running
dies with it. The code contains no network calls at all: no
login, no account, no telemetry. Your code stays on your machine.

## Documentation

- **[docs/FEATURES.md](docs/FEATURES.md)** — every feature and
  what it gives you.
- **[docs/METRICS.md](docs/METRICS.md)** — the measured numbers.
- **[docs/RESEARCH.md](docs/RESEARCH.md)** — the study behind
  them.
- **[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)** — how the
  machine works inside, every process in detail.
- **[docs/GLOSSARY.md](docs/GLOSSARY.md)** — every term
  explained.

## FAQ

**Why not just CI or pytest?**
CI checks code after you push it, and the verdict is read by a
human. punchtape works inside the task loop: it owns the task's
coverage (every line of your words declared, every line of the
tree claimed or openly crossed out), runs the checks itself,
feeds failures back to the agent, and records the finish with a
checksum. pytest is one kind of check a row can run — fuel, not a
competitor. Your CI stays; punchtape closes the loop before it.

**How is this different from a harness?**
A harness runs your assistant's conversation and lives inside the
agent's environment — it is one of the parties. punchtape is the
external party that owns the task's truth: the state, the proofs,
the verdict. Your harness stays exactly what it is; the arbiter
is added beside it.

**Can the agent cheat it?**
It can refuse to play — that is your call as the task owner (one
line in your assistant's standing instructions). While it plays,
it cannot declare "done": the act is one bare line the machine
re-checks, the verdict comes from commands the machine runs
itself, and any hole — an unclaimed line, an unresolved word, an
unpinned check — holds it. What the machine cannot do is read
meaning; see Honest limits above.

**Where do the numbers come from?**
Three rounds of field tests with blind judges, published as a
preprint (not peer-reviewed) — the methodology and the raw
campaign data are open in the repository. Poke holes: a found
flaw gets fixed and credited.

## Why the name

Punched tape — a nod at once to the punched card, the paper tape,
and Turing's machine. A hole, once punched, cannot be unpunched.
Same here: every recorded fact stays as it was. Everything
changes — **the tape remains.**
