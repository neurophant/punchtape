# punchtape trace — run 4

D1 «rjq speaks from a command line: rjq-cli <redis-url> <queue> <uuid> asks the queue for a job's status;»
├─ intent: work — «the CLI reaches the queue through status»
│  └─ R1 (span D1)
│     └─ cli/status-err (covers R1)
│        claims: src/bin/rjq-cli.rs:1-24 — token 2af98c7a21cd
│        row 0 PROOF: ./target/debug/rjq-cli redis://127.0.0.1:1/ q x → green
│        row 1 PROOF: ./target/debug/rjq-cli garbage q x → green
D2 «a failed ask is loud: the exact error line on stdout, and the exit code is one.»
├─ intent: work — «a failed ask prints the exact error line and exits one»
│  └─ R2 (span D2)
│     └─ cli/status-err (covers R2)
│        claims: src/bin/rjq-cli.rs:1-24 — token 2af98c7a21cd
│        row 0 PROOF: ./target/debug/rjq-cli redis://127.0.0.1:1/ q x → green
│        row 1 PROOF: ./target/debug/rjq-cli garbage q x → green
