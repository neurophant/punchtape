# punchtape trace — run 2

D1 «The library's version line holds;»
├─ intent: work — «the version line holds»
│  └─ R1 (span D1)
│     └─ lib/version (covers R1)
│        claims: enimda/__init__.py:1-12 — token 963fa7269ac2
│        row 0 PROOF: python3 -c "import enimda; print(enimda.__version__)" → green
D2 «the borders it finds on the fixed fixtures hold.»
├─ intent: work — «the fixture borders hold»
│  └─ R2 (span D2)
│     └─ lib/scan (covers R2)
│        claims: enimda/__init__.py:13-134 — token 67821d943546
│        row 0 PROOF: python3 -c "from enimda import ENIMDA; print(ENIMDA(fp='tests/fixed/bordered.jpg').scan(fast=True))" → green
│        row 1 PROOF: python3 -c "from enimda import ENIMDA; print(ENIMDA(fp='tests/fixed/clear.jpg').scan(fast=True))" → green
│        row 2 PROOF: python3 -c "from enimda import ENIMDA; print(ENIMDA(fp='tests/fixed/outlined.jpg').scan(fast=True))" → green
D3 «enimda answers from the command line: python3 -m enimda.cli <image> prints the four borders.»
├─ intent: work — «the CLI prints the four borders»
│  └─ R3 (span D3)
│     └─ cli/borders (covers R3)
│        claims: enimda/cli.py:1-17 — token f4657a24ebe9
│        row 0 PROOF: python3 -m enimda.cli tests/fixed/bordered.jpg → green
│        row 1 PROOF: python3 -m enimda.cli tests/fixed/clear.jpg → green
