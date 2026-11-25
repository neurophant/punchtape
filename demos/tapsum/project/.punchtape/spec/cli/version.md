# cli/version — the version line

run: 3

## Story

version prints the exact line and exits zero

## Demands

- R1 (lines D1): the version line prints

## Claims

- cmd/tapsum/main.go:1-37 — 8689e88a17b9

## Rows

## row 0 — PROOF

command: ./tapsum-demo version
exit: 0
stdout: tapsum version 1.1.0

proof: green (c272c4dfab3b)

## Verify

- ./tapsum-demo version
