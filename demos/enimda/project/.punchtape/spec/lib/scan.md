# lib/scan — the borders on the fixtures

run: 2

## Story

three fixtures scanned fast, offsets pinned

## Demands

- R2 (lines D2): the borders on the fixtures print

## Claims

- enimda/__init__.py:13-134 — a8e8b947c8d1
token: 67821d943546

## Rows

## row 0 — PROOF

command: python3 -c "from enimda import ENIMDA; print(ENIMDA(fp='tests/fixed/bordered.jpg').scan(fast=True))"
exit: 0
stdout: (4, 3, 3, 4)

proof: green (6e95a81b3165)

## row 1 — PROOF

command: python3 -c "from enimda import ENIMDA; print(ENIMDA(fp='tests/fixed/clear.jpg').scan(fast=True))"
exit: 0
stdout: (0, 0, 0, 0)

proof: green (6e94f1435749)

## row 2 — PROOF

command: python3 -c "from enimda import ENIMDA; print(ENIMDA(fp='tests/fixed/outlined.jpg').scan(fast=True))"
exit: 0
stdout: (4, 2, 4, 4)

proof: green (88dc08218a2e)

## Verify

- python3 -c "from enimda import ENIMDA; print(ENIMDA(fp='tests/fixed/bordered.jpg').scan(fast=True))"
- python3 -c "from enimda import ENIMDA; print(ENIMDA(fp='tests/fixed/clear.jpg').scan(fast=True))"
- python3 -c "from enimda import ENIMDA; print(ENIMDA(fp='tests/fixed/outlined.jpg').scan(fast=True))"
