# Demo: notes — a new project from zero

A human asks for a tiny notes utility. A coding agent builds it
under punchtape: the words become the law of the task, the checks
are pinned to lines of code, the machine runs them, records the
reds, watches the fix go green, and closes the task with a signed
verdict. Every terminal line quoted below is real output of
[`replay.sh`](replay.sh) — one command replays the whole session
from scratch on your machine.

## The task, in the human's words

```text
Make a notes utility: notes add <text> saves a note and prints its number;
notes list prints the notes one per line, oldest first;
notes show <number> prints one note;
state lives in notes.json and survives a restart.
```

Four lines, verbatim, numbered D1..D4 by the machine. The agent
never rewrites them; it claims them, line ranges and all.

## The session, in real lines

**The birth.** The agent puts the words into
`.punchtape/demands.md` and runs the machine:

```text
punchtape: run 0 — corpus empty — submit the first document as .punchtape/inbox.yaml (copy .punchtape/template.yaml)
```

**The checks — pinned on a defective build.** The agent writes
`notes.js` (32 lines, zero dependencies, state in `notes.json`) —
with two defects: `list` prints newest first, and `show` on a
missing number exits quietly. In the same breath it submits the
checks it vouches for: four behaviors, each claiming exact line
ranges of `notes.js`, each with a row pinning the exact bytes of
its output. The machine runs the rows itself and records the
facts — in [`project/.punchtape/report.md`](project/.punchtape/report.md):

```text
# punchtape report — run 1

green: 3

stdout: want "1\n2\nalpha\nbeta\n", got "1\n2\nbeta\nalpha\n"
- notes/list/0

exit: got 0, want 3
- notes/show/1
```

Two red rows, measured: the wanted bytes and the measured bytes,
side by side. No opinion anywhere.

**The fix.** The agent repairs the two defects. Only the rows
whose claimed lines moved are stale now — the machine re-runs them
on its own, finds the tree no longer fully covered, and names the
holes:

```text
punchtape: run 2 — 10 uncovered lines — submit their claims or loud exclusions to .punchtape/inbox.yaml
```

**Loud exclusions.** Two files in the tree are not the work's
source: the machine's own AGENTS block and the runtime state file.
The agent excludes them loudly, each with a reason and a
reference. The machine answers:

```text
punchtape: run 2 — all green — write done: true to .punchtape/inbox.yaml and run to record the verdict
```

**The completion act.** The agent submits `done: true` — nothing
else. The machine decides:

```text
punchtape: run 2 — verdict #1 standing — show .punchtape/state.md to the human; new work goes to .punchtape/inbox.yaml as a document
```

## The film

The whole story stands in
[`project/.punchtape/state.md`](project/.punchtape/state.md),
readable without running anything:

```text
# punchtape state — run 2

demands: 4/4 lines; requirements 4 standing, 4 covered; markers: 0
queue: empty
last no-new-demands: run 3
corpus: v2, 4 behaviors
submissions: 2 accepted, 0 refused, 1 completed
machine at birth: 0.5.1
verdict: #1 on v2, run 2, signature 2cb5bbe8a1f6 — standing

- notes/add: 1 rows — 1 green (claims 1 files / 19 lines; proof e89f52e44429)
- notes/list: 1 rows — 1 green (claims 1 files / 2 lines; proof 46832763864d)
- notes/show: 2 rows — 2 green (claims 1 files / 8 lines; proof 128de6075a9a)
- notes/persist: 1 rows — 1 green (claims 1 files / 3 lines; proof db3428fe8d95)
```

Two lessons also landed in `knowledge.md` — the machine recorded
both red→green transitions itself, as pure data: the broken
reason, the runs, the row bytes before and after.

## Replay it yourself

Needs `punchtape` on PATH (0.5.1) and `node`:

```bash
./replay.sh
```

The script rewrites `project/` from scratch and replays the whole
session live. On the same machine the film regenerates
byte-identical (the walk cache is gitignored — it carries file
mtimes). Digests and the signature ride the environment hash: on
another machine the same session records the same greens under
different digests.

## Layout

- [`replay.sh`](replay.sh) — the session engine; every step is one
  machine call.
- [`scenario/`](scenario/) — the words, both builds, both
  documents: everything the "agent" submits, ready to read.
- [`project/`](project/) — the code and the film (`.punchtape/`);
  start with `state.md`, then `trace.md` for every demand line's
  fate.

New to the machine? [`../../README.md`](../../README.md) is the
front page; [`../../docs/BOOTSTRAP.md`](../../docs/BOOTSTRAP.md)
is the one-line start for your own agent.
