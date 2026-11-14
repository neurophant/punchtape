# punchtape knowledge — run 4


## lesson: exit: got 0, want 1
- cli/status-err/0: runs 1→2
  before: run: ./target/debug/rjq-cli redis://127.0.0.1:1/ q x expect: rc: 1 out: | err Connection refused (os…
  after:  run: ./target/debug/rjq-cli redis://127.0.0.1:1/ q x expect: rc: 1 out: | err Connection refused (os…
- cli/status-err/1: runs 1→2
  before: run: ./target/debug/rjq-cli garbage q x expect: rc: 1 out: | err Redis URL did not parse
  after:  run: ./target/debug/rjq-cli garbage q x expect: rc: 1 out: | err Redis URL did not parse
