# lib/version — the version line

run: 2

## Story

the module reports its version

## Demands

- R1 (lines D1): the version line prints

## Claims

- enimda/__init__.py:1-12 — 497e92a435e6
token: 963fa7269ac2

## Rows

## row 0 — PROOF

command: python3 -c "import enimda; print(enimda.__version__)"
exit: 0
stdout: 2.1.0

proof: green (728eec9f7b0e)

## Verify

- python3 -c "import enimda; print(enimda.__version__)"
