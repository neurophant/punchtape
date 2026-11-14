# cli/version — the version line

run: 3

## Story

version prints the exact line and exits zero

## Demands

- R1 (lines D1): the version line prints

## Claims

- cmd/gouija/main.go:1-61 — c110b2df82a6
token: 3bd56d1c8b4f

## Rows

## row 0 — PROOF

command: ./gouija-demo version
exit: 0
stdout: gouija version 1.3.0

proof: green (146de32a3659)

## Verify

- ./gouija-demo version
