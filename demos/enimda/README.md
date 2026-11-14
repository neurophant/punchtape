# Demo: enimda — new words, new lines, a second verdict

enimda is a real, finished Python library — entropy-based image
border detection
([neurophant/enimda@18cd569](https://github.com/neurophant/enimda),
2017 vintage, running today on Python 3.14). This case shows the
half of punchtape's life that matters most in the long run: **a
feature on top of a standing verdict**. The zone already holds the
library; the human adds one line of new words; the agent writes a
CLI the library never had; only the new lines get claimed, only
the new rows run — and the machine records a second verdict beside
the first. Every terminal line below is real output of
[`replay.sh`](replay.sh).

## Round one — the standing library

The human's words:

```text
The library's version line holds;
the borders it finds on the fixed fixtures hold.
```

The universe counts the package directory; the first document
claims the whole standing module in two behaviors — the version
line, and the border detection pinned on three real fixtures
(byte-exact offsets from real scans). Nothing is red: a backfill
checks what already stands.

```text
punchtape: run 1 — all green — write done: true to .punchtape/inbox.yaml and run to record the verdict
```

```text
punchtape: run 1 — verdict #1 standing — show .punchtape/state.md to the human; new work goes to .punchtape/inbox.yaml as a document
```

## Round two — the human's new words

One line lands in `demands.md`:

```text
enimda answers from the command line: python3 -m enimda.cli <image> prints the four borders.
```

The next run is the whole point of this demo — the machine sees
the words and does nothing with them, because words are not
checks:

```text
punchtape: run 1 — unresolved lines: D3 (no outcome) — submit to .punchtape/inbox.yaml the carrier each declared outcome asks (a requirement on the line, a not-a-requirement marker, or an answer — removing a carrier re-opens its line)
```

The agent writes `enimda/cli.py` — 17 new lines, the library
untouched — and submits the second document: one new behavior
claiming exactly those 17 lines, two rows pinning the CLI's output
on two fixtures. The library's rows do not re-run: their claimed
lines never moved. The report of the run shows only the new work:

```text
# punchtape report — run 2

green: 2
```

The completion act, and the film's last word:

```text
punchtape: run 2 — verdict #2 standing — show .punchtape/state.md to the human; new work goes to .punchtape/inbox.yaml as a document
```

## The film

From [`project/.punchtape/state.md`](project/.punchtape/state.md) —
two signatures, one history:

```text
demands: 3/3 lines; requirements 3 standing, 3 covered; markers: 0
queue: empty
corpus: v2, 3 behaviors
submissions: 2 accepted, 0 refused, 2 completed
machine at birth: 0.5.1
verdict: #2 on v2, run 2, signature c987b51fd4ce — standing
verdict history:
- run 1: verdict #1 on v1 recorded — 9ec0cda00d35
- run 2: verdict #2 on v2 recorded — c987b51fd4ce
- lib/version: 1 rows — 1 green (claims 1 files / 12 lines; proof 963fa7269ac2)
- lib/scan: 3 rows — 3 green (claims 1 files / 122 lines; proof 67821d943546)
- cli/borders: 2 rows — 2 green (claims 1 files / 17 lines; proof f4657a24ebe9)
```

Two more honest lines live below the behaviors: the machine marked
the `__pycache__` files outside coverage by itself («a NUL byte»),
and the rest of the repo waits outside the universe — 11 files
there, 4 counted here.

## Replay it yourself

Needs `punchtape` on PATH (0.5.1) and `python3`. The replay builds
its own pinned environment (`.venv`: `numpy` 2.3.4, `Pillow`
11.3.0, `imgpy` 1.1.0 — network once) and puts it first on PATH,
so the machine's clean rows see exactly these packages.

```bash
./replay.sh
```

On the same machine the film regenerates byte-identical. The
border offsets are integers from entropy comparisons — stable
across runs; digests and signatures ride the environment hash and
differ on another machine.

## Layout

- [`replay.sh`](replay.sh) — the session engine: two rounds, two
  verdicts, nine steps.
- [`scenario/`](scenario/) — both rounds of words, both documents,
  and the feature itself (`feature/enimda/cli.py`) — everything
  the "agent" writes, ready to read.
- [`project/`](project/) — the source slice (pinned to `18cd569`;
  the 6.2 MB animated-GIF fixtures stay in the source repo), the
  five static fixtures, and the film (`.punchtape/`).

New to the machine? [`../../README.md`](../../README.md) is the
front page; [`../../docs/BOOTSTRAP.md`](../../docs/BOOTSTRAP.md)
is the one-line start for your own agent. The fresh-project story
is [`../notes/`](../notes/README.md); the standing-repo backfill
is [`../gouija/`](../gouija/README.md).
