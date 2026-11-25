# lib/version — the version line

run: 2

## Story

the module reports its version

## Demands

- R1 (lines D1): the version line prints

## Claims

- enimda/__init__.py:1-12 — 497e92a435e6

## Rows

## row 0 — PROOF

command: python3 -c "import enimda; print(enimda.__version__)"
exit: 0
stdout: 2.1.0

proof: green (d9b0d82d1132)

## Verify

- python3 -c "import enimda; print(enimda.__version__)"
