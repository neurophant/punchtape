# Changelog

All notable changes to punchtape. One law applies to this history: one
zone, one machine — a zone born on one version is refused by another,
so each entry below also answers "what happens to my zone?".

## [0.5.1] — 2026-11-08

The protocol hardened — documents are bounded, the build is part of
the done law — and the zone stands safer under accidents and
concurrency.

### Added

- **Zone lock.** Concurrent calls on one zone serialize on a
  kernel-released `run.lock`; a killed run cannot leave the zone
  frozen — the lock dies with the process that took it.
- **Document ceilings.** A submission document is bounded — its bytes,
  its lists, its behaviors' claims and rows, its expectations' files;
  the constants are printed in the contract. An over-limit submission
  is refused with the ceiling and the place named.
- **Line-range bounds.** A claim that points past the end of the file
  it names is refused at validation, before anything runs.
- **The build in the done law.** `done: true` now requires the declared
  build to stand green; a red build blocks the verdict and the machine
  names it in its one-line instruction.
- **Run budgets for rows.** Command rows carry an explicit run-time
  budget; a hung command cannot hold the zone.
- **The call from below.** The machine finds its zone from any folder
  inside the project — at the root or below it.
- **Continuous integration.** The full battery — vet, staticcheck,
  golint, unparam, gopls, dupl, the 100%-coverage gate, the
  120-second ceiling, ruff and mypy --strict over the Python tools —
  runs on every push and pull request, on ubuntu and macOS.
- **CITATION.cff and CONTRIBUTING.md** — how to cite the work, how to
  help.
- **Repository-local Python tooling** — a uv venv pinned by `uv.lock`
  for the campaign tools: `make venv`, `make python-lint`.
- **Filesystem fault-injection tests** — atomic writes proven under
  injected write failures.

### Changed

- The canon pages state the new laws (bounded documents, the build in
  the done law) in every contract language.
- README carries the demo GIF and the wordmark; the user docs are
  refreshed in all three languages.
- The coverage gate gained an `open` mode — the same diagnosis, no
  failure — for platforms whose kernel leaves branches unreachable;
  the strict reference platform is linux.

### Zone compatibility

A zone born on 0.5.0 is refused by 0.5.1 under the version law:
archive it and start fresh.

## [0.5.0] — 2026-11-01

The first public release. The machine as it stands:

### The protocol

- The human's words enter the work verbatim, one line each, into
  `.punchtape/demands.md`. The machine takes them from the file itself,
  numbers them D1..Dn and seals the past: a touched past line is zone
  damage. (Renamed from the 0.4 wish stream — a paraphrased demand is
  a defect, and the file, not the agent, carries the words.)
- The submission ends with one bare line — `done: true` — and the
  machine decides whether those words may be recorded: it re-runs
  every stale row itself, compares bytes, resolves every word,
  checks the tree against its claims plus loud exclusions. The
  sign-off cannot be said, only earned. (Replaces the 0.4 corpus-copy
  act.)
- The verdict is a journaled, signed fact — sha256 over the canonical
  bytes — that stands as long as the journal does.
- Every line of the tree is either claimed by a check or crossed out
  openly with a reason that cites the words or the law; exclusions
  are data, not shame.
- The machine is one call with no arguments. At birth it prints its
  contract — the whole law on one page; every run answers with one
  line: the next action and the file to touch. The zone carries its
  own manual (contract, template, reference pages), and an `AGENTS.md`
  note points every arriving agent at the state.

### Distribution

- One-line install, no admin rights, no account:
  `curl -fsSL https://github.com/neurophant/punchtape/releases/latest/download/get-punchtape.sh | bash`
- Binaries for linux and darwin, amd64 and arm64, each with a sha256
  sidecar; `make dist` reproduces the set.
- MIT license. The code makes no network calls at all; the machine
  works offline.

## Prehistory: 0.1.0 – 0.4.0

Working iterations of the machine and its protocol — the wish stream,
the corpus-copy sign-off — superseded by 0.5.0. The tags remain for
the record; nothing was publicly released on them.
