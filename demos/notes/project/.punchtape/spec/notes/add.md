# notes/add — add saves a note and prints its number

run: 2

## Story

two adds number 1 then 2

## Demands

- R1 (lines D1): add saves a note and prints its number

## Claims

- notes.js:1-19 — a8c7c173d884
token: e89f52e44429

## Rows

## row 0 — PROOF

command: rm -f notes.json && node notes.js add hello && node notes.js add world
exit: 0
stdout: 1
2

proof: green (d9d59e12b2f7)

## Verify

- rm -f notes.json && node notes.js add hello && node notes.js add world
