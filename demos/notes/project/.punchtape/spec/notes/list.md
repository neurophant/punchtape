# notes/list — list prints notes oldest first

run: 2

## Story

two adds then list prints alpha then beta, oldest first

## Demands

- R2 (lines D2): list prints the notes one per line, oldest first

## Claims

- notes.js:20-21 — 373f31b26286
token: 46832763864d

## Rows

## row 0 — PROOF

command: rm -f notes.json && node notes.js add alpha && node notes.js add beta && node notes.js list
exit: 0
stdout: 1
2
alpha
beta

proof: green (b8b064cb3c7b)

## Verify

- rm -f notes.json && node notes.js add alpha && node notes.js add beta && node notes.js list
