# Usage — every scenario, step by step

**English** · [Русский](USAGE.ru.md) · [中文](USAGE.zh.md)

This page is for the human who owns the tasks. Your assistant has
its own entry points: [BOOTSTRAP.md](BOOTSTRAP.md) for the first
contact, and — once a zone exists — the printed contract inside
`.punchtape/`. That contract is the law; this page never replaces
it. Everything below ends in one skill: you write tasks in
words; the machine and the agent do the rest.

## Install, once per computer

Paste this to your AI assistant — or run it yourself in a
terminal:

```text
Install punchtape and run all work through it:

  curl -fsSL https://github.com/neurophant/punchtape/releases/latest/download/get-punchtape.sh | bash
```

What this does: downloads one small binary, verifies its
checksum, puts one command — `punchtape` — into `~/.local/bin`.
No admin rights, no account. You do this once per computer, never
per project. (Build from source instead: `git clone
https://github.com/neurophant/punchtape && cd punchtape &&
./install.sh` — needs Go 1.26+.)

One more line, once: if your assistant has permanent instructions
(a place where you write standing rules for it), add the line "run
all work through punchtape" — and you will never have to write it
in a prompt again.

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

## Start any new project — first one or the tenth

Open a new, empty folder. Write your task to your assistant in
normal words:

```text
Make a notes utility: notes add <text> saves a note and prints its
number; notes list prints the notes one per line, oldest first;
notes show <number> prints one note; state lives in notes.json and
survives a restart.
```

What happens, step by step (terminal lines below are real output
of this version):

1. Your assistant runs `punchtape` in the folder. The folder is
   empty, so the machine prints its contract — one page with the
   whole law of the machine on it: what to do, which files to
   write, in what shape. Your assistant reads it and knows the
   entire process. You explain nothing.
2. Your assistant puts your task text — your words, unchanged,
   one line each — into `.punchtape/demands.md`. The next run
   births the zone: the machine's full manual (the contract, the
   template, a reference page) lands beside your words. From that
   moment the past of your words is locked — a touched past line
   is zone damage, and the machine stops. New words keep entering
   the same file at any moment.

   ```
   punchtape: run 0 — corpus empty — submit the first document as .punchtape/inbox.yaml (copy .punchtape/template.yaml)
   ```

3. Your assistant copies `.punchtape/template.yaml` to
   `.punchtape/inbox.yaml` and fills it in: the story of the
   submission, the outcome declared for every new line of your
   words, the requirements, and the checks — each behavior claims
   exact lines of code, each check is a command with its pinned
   expected output. The machine runs the rows itself and says
   what stands between you and the finish, in one line:

   ```
   punchtape: run 1 — behaviors without a green pinned row: notes/list — submit to .punchtape/inbox.yaml a row whose expect pins its exact output
   ```

   Failed checks land in `.punchtape/report.md` — what was
   expected, what actually happened, side by side, in bytes.
   Your assistant reads that file and fixes the code or the rows.
4. The fix lands; the machine re-runs only what went stale and
   names the remaining holes:

   ```
   punchtape: run 2 — 10 uncovered lines — submit their claims or loud exclusions to .punchtape/inbox.yaml
   ```

   Files the work does not cover get loud exclusions — openly,
   with a reason, never silently.
5. All checks green → the terminal says so, and your assistant
   writes exactly one line to the inbox: `done: true`. That is
   the sign-off — nothing to retype, nothing to format.

   ```
   punchtape: run 2 — all green — write done: true to .punchtape/inbox.yaml and run to record the verdict
   ```

6. The machine re-checks everything — every row green with a
   pinned proof, your words resolved whole, every requirement
   covered, the tree equal to its claims plus loud exclusions, no
   open question — and records the verdict:

   ```
   punchtape: run 2 — verdict #1 standing — show .punchtape/state.md to the human; new work goes to .punchtape/inbox.yaml as a document
   ```

   From now on `.punchtape/state.md` shows: `verdict: #1 … —
   standing`, with its checksum. Show that file to anyone who
   asks if it works.

The next new project after this one is the same thing again:
empty folder, task in words, done. Each folder is separate;
nothing from one project leaks into another.

## Continue a project that already runs on punchtape

Open that project folder — or any folder inside it (the machine
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
`.punchtape/demands.md`; the checks claiming the changed files
run again; the checks that still hold are not re-run and cost
nothing. When the addition works, the machine records finish #2 —
then #3, #4… Each verdict is its own recorded fact with its own
checksum, and if a later change breaks a sealed result, the
machine names the run that caught it and the run that healed it.

## Real prompts for real situations

**A bug.**
`notes show 3 prints nothing when there is no note 3 — fix it and prove it cannot happen again.`
Your assistant adds a check for the wanted behavior (it fails —
the bug exists), fixes the code, the check passes, and that check
stays in the project forever: it re-runs whenever the lines it
claims change. The rjq demo in [demos/](../demos/) is exactly
this story, on a standing Rust library.

**An existing project that never used punchtape.**
`Bring this repo under punchtape: read the code, write checks for the behavior that already exists, make them all pass.`
The machine measures what your code really does today and drafts
the exclusion list for the rest. From then on every change is
measured against that. The gouija demo is this story, on a
standing Go repo; the enimda demo shows the feature that comes
after it.

**A new AI model took over.**
`Continue this project through punchtape — read .punchtape/state.md first, then finish what is missing.`
The new model needs no old chat: the project folder holds
everything.

**Hand the project to a teammate.**
Copy the folder. Done. Words, checks, results, history, finishes —
everything is inside, byte-identical.

## The team cookbook

### The zone in the repository

Commit the film whole: `demands.md`, `journal/` (epoch files and
objects), `corpus.yaml`, `state.md`, `report.md`, `spec/`,
`holes.md`, `trace.md`, `knowledge.md`, `contract.md`,
`template.yaml`, `reference.md`, and the machine's marked block in
the root `AGENTS.md`. This is the recorded past of the task — in
version control it survives everything, and a clone instantly
carries the whole record.

Keep the generated service state out of git — `cache/`,
`derived/`, `run.lock` inside `.punchtape/`: the machine rebuilds
them on every run. (`.punchtape/tmp/` needs no rule of its own:
the machine clears it at each run's end.) A new teammate needs no
onboarding: they clone, their agent reads `AGENTS.md`, opens
`.punchtape/state.md`, and continues where the `next:` line
points.

### The zone outside the repository

Nothing forces the zone into git. With `.punchtape/` in your
`.gitignore`, the repository carries the code, the zone carries
the truth, and each stays whole. What each developer (and your
CI) sees is the plain tree. What transfers on a task handover is
the folder, not a branch. When zones diverge, the machine's own
record is where the truth lives: the journal, which cannot be
edited without the machine naming it. The cost is plain too: a
clone without the zone starts empty, so this shape fits a solo
owner or an explicit hand-over more than a crowd.

### Conflicts and their recipes

The machine meets every conflict by refusing loudly and naming
the file — it never guesses, never absorbs damage silently. Every
line below is real output of this version. The recipes are lines
you paste to your assistant; the doing is the agent's.

**A parallel call while a run is standing.** Runs serialize on
the zone's own lock (`run.lock`, kernel-released, death-safe): a
second call waits its turn and then proceeds — measured live, an
external lock holder delayed one run for exactly as long as it
held the lock, and the run then completed normally. You do nothing;
the lock is not yours to manage.

**The journal conflicted in a merge (two clones, both
submitted).** The journal is one linear record: it never merges.
Non-tail damage is refused with the file and record named:

```
punchtape: journal: record 0 of /tmp/epoch-conflict/.punchtape/journal/epoch-00002.yaml does not parse: [1:17] could not find flow map content
```

A torn final record is different — the machine quarantines it to
`<epoch>.torn` beside the intact records and continues. Recipe:

```text
Resolve the .punchtape merge conflict by keeping exactly one side whole — demands.md, journal/, corpus.yaml together, never mixed. Then run punchtape. If it reports damage, restore .punchtape from the last commit where the zone was green (git checkout <sha> -- .punchtape), and resubmit the other side's words and document as new work.
```

**The past of the words was touched.** The chain recomputes over
the bytes on every run, and a changed past line stops everything:

```
punchtape: zone damaged: the demand stream's past changed
```

Restoring the exact bytes heals it — verified live: the same
zone, the same run, standing again after the original bytes came
back. With the zone in git, that restore is one checkout:

```text
.punchtape/demands.md was edited by hand and the machine reports zone damaged. Restore the file's exact bytes from git (git checkout -- .punchtape/demands.md), then run punchtape again.
```

**A zone born by another machine version.** Versions do not mix,
and the machine says the whole recipe itself:

```
punchtape: zone born by machine 9.9.9 — this machine is 0.5.1; archive it (cp -r .punchtape .punchtape-archive) and start fresh
```

**A junk or half-dead zone sits in the tree.** A `.punchtape`
folder with no journal and no words yet is nothing — the machine
says so and waits:

```
punchtape: no instance yet — put the human's words verbatim into .punchtape/demands.md and run again
```

If the journal is gone but words stand, the next run re-births
the zone empty — the recorded past is the journal; without it
there is no past to restore. Recipe:

```text
This .punchtape is junk from an experiment. Archive it (mv .punchtape ../punchtape-junk), and if this project should run on punchtape, start it fresh from my words.
```

With the zone in the repository, "start fresh" has a safety net:
the film is in git history, and any green state of it can come
back with one checkout of `.punchtape` from that commit.

**Two zones over one tree.** The zone is the nearest
`.punchtape` at or above the working directory — a nested folder
shadows the project's real zone silently. Verified live: called
from a deep subdirectory, the machine pointed at
`../../../.punchtape/state.md`, the zone above. Recipe:

```text
Find the nested zone (find . -name .punchtape -not -path './.punchtape'), tell me which one is the project's, archive the other, and run punchtape at the working root.
```

**Several agents, one zone.** The lock serializes the runs, not
the work. Who submits which part is your call as the task owner:
split the work into non-overlapping parts, one submission at a
time. punchtape is an arbiter of facts, not a coordinator of
agents.
