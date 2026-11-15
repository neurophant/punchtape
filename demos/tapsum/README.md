# Demo: tapsum — a standing repo comes under the machine

tapsum is a small, finished Go utility — tape-style checksums for
files, zero dependencies, fully offline. Nobody wrote a line for
this demo: the code stands as it shipped. The story here is the
other half of punchtape's life — not a new project, but
**a backfill**: checks written against behavior that already
exists, pinned to lines that are already standing, all green on
the first run, the one hole that cannot be pinned yet marked
loudly. Every terminal line below is real output of
[`replay.sh`](replay.sh).

## The task, in the human's words

```text
The version line prints and holds;
the usage page prints and holds.
```

## The session, in real lines

**The birth** — same one line as anywhere:

```text
punchtape: run 0 — corpus empty — submit the first document as .punchtape/inbox.yaml (copy .punchtape/template.yaml)
```

**The counted root.** The agent declares the universe — today's
document counts one directory, `cmd/tapsum`; the other six files
of the repo stay outside until work touches them:

```text
punchtape: run 0 — corpus empty — submit the first document as .punchtape/inbox.yaml (copy .punchtape/template.yaml)
```

**The first document.** The build is declared
(`go build -o tapsum-demo ./cmd/tapsum` — zero dependencies, no
network ever). Two behaviors claim the CLI's printed contract: the
exact version line, and the whole usage page pinned byte for byte
— 300 bytes of exact output. The machine compiles, runs the rows
against the old code, and nothing is red — a backfill checks what
already stands:

```text
# punchtape report — run 1

build: exit 0

green: 2
```

What the machine asks for next is the honest part:

```text
punchtape: run 1 — 27 uncovered lines — submit their claims or loud exclusions to .punchtape/inbox.yaml
```

**The loud exclusion.** The watch loop re-hashes a file forever,
one logged pass a second — and every line carries a wall clock,
so no byte-exact pin can exist. The pin law refuses such a row
outright, so the agent does not fake one: it excludes the range
loudly, with a reason and the law that forced the mark:

```text
punchtape: run 2 — all green — write done: true to .punchtape/inbox.yaml and run to record the verdict
```

**The completion act.**

```text
punchtape: run 3 — verdict #1 standing — show .punchtape/state.md to the human; new work goes to .punchtape/inbox.yaml as a document
```

## The film

From [`project/.punchtape/state.md`](project/.punchtape/state.md):

```text
# punchtape state — run 3

demands: 2/2 lines; requirements 2 standing, 2 covered; markers: 0
queue: empty
last no-new-demands: run 2
corpus: v2, 2 behaviors
submissions: 2 accepted, 0 refused, 1 completed
machine at birth: 0.5.1
verdict: #1 on v2, run 3, signature 8f087e3b7fbb — standing

- cli/version: 1 rows — 1 green (claims 1 files / 37 lines; proof 281bab05363d)
- cli/usage: 1 rows — 1 green (claims 1 files / 18 lines; proof 2fb63c70f69f)
excluded lines: 1 ranges (32 % of tree)
outside the universe: 6 files
universe: 1 of 7 files
```

Read the last two lines twice: a third of the counted file is
excluded — loudly, with a reason — and six files of the standing
repo wait outside the universe for the documents that will claim
them. Nothing is silently green.

## Replay it yourself

Needs `punchtape` on PATH (0.5.1) and `go` (1.24+). Zero
dependencies: the build runs fully offline, no network ever.

```bash
./replay.sh
```

The script rebuilds `project/` from scratch and replays the whole
session live. On the same machine the film regenerates
byte-identical (the walk cache and the built binary are
gitignored). Digests and the signature ride the environment hash:
on another machine the same session records the same greens under
different digests.

## Layout

- [`replay.sh`](replay.sh) — the session engine.
- [`scenario/`](scenario/) — the words and both documents, ready
  to read; the 300-byte usage pin is generated from the binary's
  real output.
- [`project/`](project/) — the source (five files, one command),
  and the film (`.punchtape/`); start with `state.md`, then
  `trace.md`.

New to the machine? [`../../README.md`](../../README.md) is the
front page; [`../../docs/BOOTSTRAP.md`](../../docs/BOOTSTRAP.md)
is the one-line start for your own agent. The fresh-project story
is [`../notes/`](../notes/README.md).
