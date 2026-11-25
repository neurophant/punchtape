# cli/status-err — the loud ask

run: 4

## Story

the ask goes through status to the queue; a refused connection and an unparsable url both print the exact error line and exit one

## Demands

- R1 (lines D1): the CLI reaches the queue through status
- R2 (lines D2): a failed ask prints the exact error line and exits one

## Claims

- src/bin/rjq-cli.rs:1-24 — 9d9f7d31bd10

## Rows

## row 0

command: ./target/debug/rjq-cli redis://127.0.0.1:1/ q x
exit: 1
stdout: contains: err Connection refused
proof: green (0defcaca6454)

## row 1 — PROOF

command: ./target/debug/rjq-cli garbage q x
exit: 1
stdout: err Redis URL did not parse

proof: green (e517292b4df8)

## Verify

- ./target/debug/rjq-cli redis://127.0.0.1:1/ q x
- ./target/debug/rjq-cli garbage q x
