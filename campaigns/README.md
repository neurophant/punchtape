# Campaigns — punchtape field runs

Each campaign is a self-sufficient folder, run at choice and
independently of the others: none imports another's paths. The
common law of every run is `METHODOLOGY.md`; each campaign's README
is its entry point.

| Campaign | The question | Meanings | Waves | Status |
|---|---|---|---|---|
| `case-s` | a full run of the implementation | brim, conduit, errand, pressmark, quire | A (smoke) + B × 5 | closed |
| `case-m` | the machine camp against solo, one hand a kind, one binary per series | brim, conduit, errand, pressmark, quire | M × 5 | closed |
| `case-f` | the field cases on the S basis: waves from seeded working products under the blindness and seeding laws | brimfield, brimfeat, brimdrift, brimstack | M × 4 | closed |

## The common run order

The world is host cells: without containers or images,
cell-directories on the host, one scratch-root. Two hand kinds
(machine and solo), one fully automatic operator per hand, the
parallelism ceiling 2 (METHODOLOGY §0.2c). A wave goes:

1. `cells.py --wave waves/<W>.yaml --root <run root> --machine
   <pinned binary>` — stages the wave's cells.
2. The hands, operators and judges — as the methodology's session
   law orders.
3. `collect.py` and `metrics.py` — the wave's usage and METRICS in
   the wave root.

## Isolation

- Campaigns do not reference each other's paths.
- What they share is only what lies outside their folders: the
  platform metrics base and the common `METHODOLOGY.md`; hidden
  measurement materials lie INSIDE each folder.
- Any campaign can be run, deleted or rewritten without touching
  the others.

Runs live only in `~/.experiments/runs/`; handoffs — in
`./handoffs/` (outside git).
