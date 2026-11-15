# punchtape trace — run 3

D1 «The version line prints and holds;»
├─ intent: work — «the version line prints and holds»
│  └─ R1 (span D1)
│     └─ cli/version (covers R1)
│        claims: cmd/tapsum/main.go:1-37 — token 281bab05363d
│        row 0 PROOF: ./tapsum-demo version → green
D2 «the usage page prints and holds.»
├─ intent: work — «the usage page prints and holds»
│  └─ R2 (span D2)
│     └─ cli/usage (covers R2)
│        claims: cmd/tapsum/main.go:38-55 — token 2fb63c70f69f
│        row 0 PROOF: ./tapsum-demo help → green
