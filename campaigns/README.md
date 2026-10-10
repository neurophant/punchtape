# Campaigns — punchtape field runs

The campaigns' self-sufficient folders live here. Each is run at
choice, independently of the others: none imports another's paths.
The project's common methodology is `METHODOLOGY.md`.

## case-0 — a full run of the implementation (protocol 0.2.0)

Detailed briefs, materials, prompts, cell preparation and
measurement tools, preregistration and prices. The world is host
cells: without containers or images, cell-directories on the host,
one scratch-root. TWO hands (solo and machine-top), the parallelism
ceiling 8 (METHODOLOGY §0.2c).

Run: by the README inside the folder (cell preparation
`tools/prep-cell.sh`, running the waves, the metrics tables).

## Isolation

- Campaigns do not reference each other's paths.
- What they share is only what lies outside their folders: the
  platform metrics base and the common `METHODOLOGY.md`; hidden
  measurement materials lie INSIDE each folder.
- Any campaign can be run, deleted or rewritten without touching
  the others.

Runs live only in `~/.experiments/runs/`; handoffs — in
`./handoffs/` (outside git).
