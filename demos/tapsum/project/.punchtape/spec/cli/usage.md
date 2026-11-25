# cli/usage — the usage page

run: 3

## Story

help prints the whole usage page to stdout and exits zero

## Demands

- R2 (lines D2): the usage page prints

## Claims

- cmd/tapsum/main.go:38-55 — c031e07dc82c

## Rows

## row 0 — PROOF

command: ./tapsum-demo help
exit: 0
stdout: tapsum - tape-style checksums for files

Usage:
  tapsum <command> [arguments]

Commands:
  version        Print the version line and exit
  help           Print this usage page and exit
  watch <file>   Re-hash a file forever, one line per pass

Examples:
  tapsum version
  tapsum watch server.log

proof: green (1e891786c4bf)

## Verify

- ./tapsum-demo help
