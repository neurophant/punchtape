# notes/persist — state survives a restart

run: 3

## Story

a note added in one process is listed by the next one

## Demands

- R4 (lines D4): the state file survives a restart

## Claims

- notes.js:30-32 — 6fb2e8700493

## Rows

## row 0 — PROOF

command: rm -f notes.json && node notes.js add pinned && node notes.js list
exit: 0
stdout: 1
pinned

proof: green (631dfe3e315e)

## Verify

- rm -f notes.json && node notes.js add pinned && node notes.js list
