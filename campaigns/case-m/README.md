# case-m — the M series

## The question

The machine camp's economy and quality against solo, one hand a
kind and one binary for the whole series: each meaning in all five
stacks, ten cells a wave. The store-and-cache meanings run on
per-cell PostgreSQL and Redis (service scripts staged into each
cell's `env/service/`).

## Meanings and waves

- Meanings: brim, conduit, errand, pressmark, quire —
  `briefs/<meaning>/<stack>.md`, `materials/<meaning>/`,
  `etalon/<meaning>.md`.
- Waves: `waves/M-<meaning>.yaml` × 5 — all five stacks, both
  kinds, one hand a kind (ten cells a wave).

## World

`world-lock.txt` — the locked platform of the series (one warmed
master per meaning × stack, the service and toolchain pins).

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
