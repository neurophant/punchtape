# case-s — a full run of the implementation

## The question

Does the machine lower the cost of reaching a given level of
quality at comparable or better quality, on identical
detailed-brief tasks (METHODOLOGY §I.1). Five meanings, each in all
five stacks, both hand kinds, one model pin — the first full
measured run of the implementation.

## Meanings and waves

- Meanings: brim, conduit, errand, pressmark, quire —
  `briefs/<meaning>/<stack>.md`, `materials/<meaning>/`,
  `etalon/<meaning>.md`.
- Waves: `waves/A.yaml` — the stage A smoke (quire/go, both kinds,
  two hands a kind); `waves/B-<meaning>.yaml` × 5 — all five
  stacks, both kinds, one hand a kind (ten cells a wave).

## World

`world-lock.txt` — the locked platform of the series (the warmed
masters per meaning × stack, the toolchain pins).

## Role tools

`spawn.yaml` (the hands' spawn bytes), `harness.md` (the cells'
preamble), `judge.md` (the judge's frozen prompt),
`operator-machine.md` / `operator-solo.md` (the operators' prompts).

## Launch

    cells.py --wave waves/<W>.yaml --root ~/.experiments/runs --machine <pinned binary>

then the hands, operators and judges as the session law orders,
then `collect.py` and `metrics.py` for the wave's METRICS.

## Status

Closed.
