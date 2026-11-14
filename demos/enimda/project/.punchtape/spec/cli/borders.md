# cli/borders — the command-line borders

run: 2

## Story

two fixtures through the CLI, offsets pinned

## Demands

- R3 (lines D3): the CLI prints the four borders of an image

## Claims

- enimda/cli.py:1-17 — e9e621853062
token: f4657a24ebe9

## Rows

## row 0 — PROOF

command: python3 -m enimda.cli tests/fixed/bordered.jpg
exit: 0
stdout: 4 3 3 4

proof: green (279bc085d81a)

## row 1 — PROOF

command: python3 -m enimda.cli tests/fixed/clear.jpg
exit: 0
stdout: 0 0 0 0

proof: green (7a03bac04266)

## Verify

- python3 -m enimda.cli tests/fixed/bordered.jpg
- python3 -m enimda.cli tests/fixed/clear.jpg
