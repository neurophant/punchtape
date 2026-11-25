# notes/show — show prints one note and refuses a missing one

run: 3

## Story

show 1 prints the note; show 9 exits non-zero

## Demands

- R3 (lines D3): show prints one note and refuses a missing one

## Claims

- notes.js:22-29 — 1fe0e7e3d3a9

## Rows

## row 0 — PROOF

command: rm -f notes.json && node notes.js add first && node notes.js show 1
exit: 0
stdout: 1
first

proof: green (b79b509b8342)

## row 1

command: node notes.js show 9
exit: 3
proof: green (55b763d5ad36)

## Verify

- rm -f notes.json && node notes.js add first && node notes.js show 1
- node notes.js show 9
