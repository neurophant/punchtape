# punchtape
**English** · [Русский](README.ru.md) · [中文](README.zh.md)

**Your AI writes the code. punchtape checks it — and does not stop
until "done" is really done.**

punchtape is a small program for your computer. Your AI assistant
runs it while it works. The program runs real checks of the AI's
work, tells the AI what to fix, and writes down the result. When
every check passes, it records the finish. You read the result —
you do not check anything yourself.

This is not another harness, and it has nothing to do with
harnesses. The harness runs your assistant's conversation;
punchtape is an external transactional state machine: one program
beside your project that owns the task's state — the checks, the
facts, the recorded finishes — and drives the agent through the
work from the first line to the sealed end. Not a plugin and not a
library inside the agent: an arbiter on the outside. Whatever
environment your assistant lives in stays exactly what it is;
punchtape adds the one thing it lacks — a party outside the
conversation that owns the task's truth.

Free. MIT license. Linux and macOS. Works offline. No account.

## The problem it solves

This is what normally happens:

1. You give a task to your AI.
2. The AI answers: "Done! All tested!"
3. You check the work yourself. This takes 30–60 minutes.
4. You find two bugs and one missing feature.
5. You send it back. Repeat tomorrow.

Step 3 is the expensive part. Not the writing — the checking. And it
is always your job. punchtape takes step 3 away: the program runs
the checks, not you, and it will not mark the work "done" while
anything is broken or missing.

## How it works, concretely

- You write your task in normal words. Example: *"Make a notes
  utility: notes add saves a note…"*
- Your AI, next to the code, writes a list of checks: real commands
  and the results they must print. Example: run `./notes add hello`,
  it must print `1`.
- The program runs those commands itself, in your project folder.
  A check passes or fails — the program sees the real result, it
  never asks the AI.
- If a check fails, the program tells the AI exactly what failed
  and what was expected. The AI fixes it. Repeat.
- When all checks pass, your AI sends the list back to the program
  unchanged — that is the sign-off. The program records the finish:
  date-free, with a checksum, in files inside your project.
- Every task, every check, every result is written into files in
  your project folder. Nothing is stored in any cloud. Copy the
  folder — the whole history moves with it.

You never learn commands. The AI runs the program; you write prompts
as always.

## What you get

- **You stop checking the AI's work.** The program checks it, every
  time, on every change.
- **Nothing from your task gets silently dropped.** Every line of
  your task text must be covered by a check — or crossed out
  openly, with a quote and a reason. Otherwise: no finish.
- **Edits are cheap.** Change one file → only the checks for that
  file run again. Everything that still passes is not re-run.
- **If something breaks after the finish, you see it.** The program
  writes which change broke it and when it was fixed.
- **Nothing is lost between sessions.** New chat, new AI model, next
  month — open the folder and continue. All history is in files.
- **It works with anything you build, any way you build it.** Any
  language, any framework, any libraries, any project shape, any
  workflow, any harness. If your project can be built and run by a
  command, the program can check it — and nothing about how you
  work is imposed on you.

Full list with every detail: **[docs/FEATURES.md](docs/FEATURES.md)**

## Proof: measured, published

The savings were measured in three rounds of field tests (real
tasks; one group worked with the program, one without; independent
blind judges) and published as a paper:

> **"Closing the Verification Loop in Generative Development: An
> External Transactional State Machine and a Field Evaluation of Its
> Economics and Convergence"**
> — [SSRN](https://doi.org/10.2139/ssrn.7566638) ·
> [Preprints.org](https://doi.org/10.20944/preprints202610.0325.v1) ·
> [Research Square](https://doi.org/10.21203/rs.3.rs-11271483/v1)

The numbers that matter to you:

- **Less work for you.** In the last test round, the group without
  the program needed human-style acceptance **32 times in 25
  tasks**; the group with it — **2 times**.
- **In hours** (an estimate for a live person doing the accepting):
  25 tasks without the program = **4.4–7.3 hours** of your time;
  with it = **0.8–1.3 hours**, and about **0.2–0.5 hours** if you
  count only the real exceptions. That is **3.6–6.0 hours given back
  per 25 tasks**.
- **A task cost 5×–19× more in total** without the program
  (measured at the test prices).
- **Fast finishes:** 25 of 25 tasks with the program were accepted
  **on the first submission**; in the last round all 50 test tasks
  closed with recorded finishes.
- **Quality did not drop:** blind judges checked 1720 points across
  both groups — **0 missing** in each.

All numbers, year by year: **[docs/METRICS.md](docs/METRICS.md)**.
How the tests were run: **[docs/RESEARCH.md](docs/RESEARCH.md)**.

## How to use it

### 1. Install — once per computer

Paste this to your AI assistant — or run it yourself in a terminal:

```text
Install punchtape and run all work through it:

  curl -fsSL https://github.com/neurophant/punchtape/releases/latest/download/get-punchtape.sh | bash
```

What this does: downloads one small program, checks its checksum,
puts one command — `punchtape` — into `~/.local/bin`. No admin
rights, no account. You do this once per computer, never per
project. (Build from source instead: `git clone
https://github.com/neurophant/punchtape && cd punchtape &&
./install.sh` — needs Go 1.26+.)

One more line, once: if your assistant has permanent instructions
(a place where you write standing rules for it), add there "run all
work through punchtape" — and you will never have to write it in a
prompt again.

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

- The folder already runs punchtape (it has a `.punchtape` folder)?
  → **Not needed.** The file `AGENTS.md` in that folder already
  tells every AI tool about it.
- You once put that line into your assistant's permanent
  instructions (most assistants have them)? → **Not needed,
  anywhere, ever.**
- New empty folder, and your assistant has no such permanent
  instruction? → **Add the line.**
- Not sure? → **Add the line.** It never hurts.

What happens, step by step:

1. Your assistant runs `punchtape` in the folder. The folder is
   empty, so the program prints its full instruction page — one
   page, everything on it: what to do, which files to write. Your
   assistant reads it and knows the whole process. You explain
   nothing.
2. Your assistant puts your task text — your words, unchanged —
   into the file `.punchtape/wish.md`. The next run locks it
   forever: the task cannot quietly change later.
3. Your assistant writes the checks (commands + expected results)
   into the file `.punchtape/inbox.yaml`.
4. The program runs every check. Failed checks go into
   `.punchtape/report.md` — what was expected, what actually
   happened. Your assistant reads that file and fixes the code or
   the checks. One line in the terminal always says what to do now,
   for example (a real line):

   ```text
   punchtape: run 1 — 4 row(s) red — read .punchtape/report.md, fix, run again
   ```

5. All checks pass → the program answers: copy
   `.punchtape/corpus.yaml` to `.punchtape/inbox.yaml` and run
   again. Your assistant does exactly that — sending the finished
   list back is the sign-off.
6. The program records the finish. From now on
   `.punchtape/state.md` shows: task, checks, all green, finish #1
   with a checksum. Show that file to anyone who asks if it works.

The next new project after this one is the same thing again: empty
folder, task in words, done. Each folder is separate; nothing from
one project leaks into another.

### 3. Continue a project that already runs on punchtape

Open that project folder — or any folder inside it (the program
looks upward and finds its files). Write what you want, in words:

```text
Add an open command — show only the unclosed notes.
```

It does not matter who picks this up: your usual assistant, a new
session, a different AI model, a teammate. The file `AGENTS.md` in
the project root contains a short note from punchtape — every AI
tool reads that file on arrival — and the note points to
`.punchtape/state.md`, one page: where the work stands, what is
next. So anyone continues in seconds.

What happens with your change: the checks for the changed files run
again; the checks that still pass are not re-run and cost nothing.
When the addition works, the program records finish #2 — then #3,
#4… Each finish is its own recorded fact with its own checksum.

## Real prompts for real situations

**A bug.**
`notes show 3 prints nothing when there is no note 3 — fix it and prove it cannot happen again.`
Your assistant adds a check for the wanted behavior (it fails — the
bug exists), fixes the code, the check passes, and that check stays
in the project forever: it runs again whenever related code changes.

**An existing project that never used punchtape.**
`Bring this repo under punchtape: read the code, write the checks for the behavior that already exists, and get them all passing.`
The program records what your code really does today. From then on
every change is measured against that.

**A new AI model took over.**
`Continue this project through punchtape — read .punchtape/state.md first, then finish what is missing.`
The new model needs no old chat: the project folder holds
everything.

**Hand the project to a teammate.**
Copy the folder. Done. Task, checks, results, history, finishes —
everything is inside, byte-identical.

## Is it for you?

- Works with any language and framework — the program knows none of
  them and does not need to. Stack, libraries, project shape,
  workflow, harness, approach — use whatever you want, however you
  are used to, as the task requires.
- Linux and macOS today; Windows later.
- You need an AI assistant — the program never writes code itself.
- For a one-time throwaway script — too much. For anything you will
  edit again — this is where it pays.

## Free, private, offline

MIT license. One program, one command. No background processes — it
runs only when called. The code contains no network calls at all:
no login, no account, no telemetry. Your code stays on your
machine.

## Documentation

- **[docs/FEATURES.md](docs/FEATURES.md)** — every feature and what
  it gives you.
- **[docs/METRICS.md](docs/METRICS.md)** — the measured numbers.
- **[docs/RESEARCH.md](docs/RESEARCH.md)** — the study behind them.
- **[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)** — how the machine
  works inside, every process in detail.
- **[docs/GLOSSARY.md](docs/GLOSSARY.md)** — every term explained.

## Why the name

Punched tape — a nod at once to the punched card, the paper tape,
and Turing's machine. A hole, once punched, cannot be unpunched.
Same here: every recorded fact stays as it was. Everything
changes — **the tape remains.**
