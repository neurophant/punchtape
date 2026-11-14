# punchtape trace — run 3

D1 «The version line prints and holds;»
├─ intent: work — «the version line prints and holds»
│  └─ R1 (span D1)
│     └─ cli/version (covers R1)
│        claims: cmd/gouija/main.go:1-61 — token 3bd56d1c8b4f
│        row 0 PROOF: ./gouija-demo version → green
D2 «the usage page prints and holds.»
├─ intent: work — «the usage page prints and holds»
│  └─ R2 (span D2)
│     └─ cli/usage (covers R2)
│        claims: cmd/gouija/main.go:62-217 — token 467d33a0da70
│        row 0 PROOF: ./gouija-demo help → green
