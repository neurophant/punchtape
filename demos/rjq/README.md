# Demo: rjq — a red check, a fix, green forever

rjq is a real, finished Rust library — a Redis job queue
([neurophant/rjq@801aec2](https://github.com/neurophant/rjq),
2017 vintage, building today on rustc 1.98). This case is the
one every maintainer knows by heart: **a bug, a check that
catches it, a fix, and the check standing guard ever after** —
no live Redis anywhere. The pins ride the library's own
connection errors: a refused port and an unparsable URL, both
byte-exact. Every terminal line below is real output of
[`replay.sh`](replay.sh).

## The task, in the human's words

```text
rjq speaks from a command line: rjq-cli <redis-url> <queue> <uuid> asks the queue for a job's status;
a failed ask is loud: the exact error line on stdout, and the exit code is one.
```

## The session, in real lines

**The defective CLI.** The agent writes `rjq-cli` — 24 lines on
top of the standing library — and it looks done: on a failure it
prints the error line. But it exits zero. A script would call it
and march on; the silent failure is the bug. The first document
pins the contract byte for byte — the exact error line **and**
exit code one — and the machine runs the rows against the
defective build:

```text
# punchtape report — run 1

build: exit 0

exit: got 0, want 1
- cli/status-err/0
- cli/status-err/1
```

Both rows red, both measured: the wanted exit code, the measured
one, the row keys. The error text was already right, the bytes
came from the library; only the exit code lied.

**The fix.** One line: `process::exit(1)` in the error arm. The
claimed lines moved, the proofs went stale, the machine re-runs
the rows itself, and asks the next honest question:

```text
punchtape: run 2 — 287 uncovered lines — submit their claims or loud exclusions to .punchtape/inbox.yaml
```

**The loud exclusion.** The whole library needs a live Redis; no
offline pin can hold its runtime paths. The agent excludes
`src/lib.rs` whole — loudly, with the reason and the pointer to
the document that will claim it against a service one day:

```text
punchtape: run 3 — all green — write done: true to .punchtape/inbox.yaml and run to record the verdict
```

**The completion act.**

```text
punchtape: run 4 — verdict #1 standing — show .punchtape/state.md to the human; new work goes to .punchtape/inbox.yaml as a document
```

## The film

From [`project/.punchtape/state.md`](project/.punchtape/state.md):

```text
demands: 2/2 lines; requirements 2 standing, 2 covered; markers: 0
queue: empty
corpus: v2, 1 behaviors
submissions: 2 accepted, 0 refused, 1 completed
machine at birth: 0.5.1
verdict: #1 on v2, run 4, signature d9c1f6eb10c7 — standing
- cli/status-err: 2 rows — 2 green (claims 1 files / 24 lines; proof 2af98c7a21cd)
excluded lines: 1 ranges (92 % of tree)
```

Two lessons sit in `knowledge.md` — the machine recorded both
red→green transitions as data. From now on, if anyone rebuilds
this CLI and the exit code goes quiet again, the row goes red the
same hour: that is the "forever" part.

## Replay it yourself

Needs `punchtape` on PATH (0.5.1) and `cargo`. The crates download
from crates.io on the first build (network needed once). One
platform note: the refusal line pins
Linux's `os error 111`; on macOS the os error number differs, so
the pinned row expects the local text; verified on Linux.

```bash
./replay.sh
```

## Layout

- [`replay.sh`](replay.sh) — the session engine.
- [`scenario/`](scenario/) — the words, both documents, and both
  builds of the CLI (the defective one and the one-line fix), all
  ready to read.
- [`project/`](project/) — the source slice (pinned to `801aec2`),
  the pinned lockfile, and the film
  (`.punchtape/`).

New to the machine? [`../../README.md`](../../README.md) is the
front page; [`../../docs/BOOTSTRAP.md`](../../docs/BOOTSTRAP.md)
is the one-line start for your own agent. The other three cases:
[notes](../notes/README.md): a fresh project;
[gouija](../gouija/README.md): a standing repo;
[enimda](../enimda/README.md): a feature and a second verdict.
