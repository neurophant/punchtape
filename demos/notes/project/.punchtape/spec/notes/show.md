# notes/show — show prints one note and refuses a missing one

run: 2

## Story

show 1 prints the note; show 9 exits non-zero

## Demands

- R3 (lines D3): show prints one note and refuses a missing one

## Claims

- notes.js:22-29 — 1fe0e7e3d3a9
token: 128de6075a9a

## Rows

## row 0 — PROOF

command: rm -f notes.json && node notes.js add first && node notes.js show 1
exit: 0
stdout: 1
first

proof: green (3d5cf57dea82)

## row 1

command: node notes.js show 9
exit: 3
proof: green (57d66e2d9fa3)

## Verify

- rm -f notes.json && node notes.js add first && node notes.js show 1
- node notes.js show 9
