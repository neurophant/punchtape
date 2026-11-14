# punchtape knowledge — run 2


## lesson: exit: got 0, want 3
- notes/show/1: runs 1→2
  before: run: node notes.js show 9 expect: rc: 3
  after:  run: node notes.js show 9 expect: rc: 3

## lesson: stdout: want "1\n2\nalpha\nbeta\n", got "1\n2\nbeta\nalpha\n"
- notes/list/0: runs 1→2
  before: run: rm -f notes.json && node notes.js add alpha && node notes.js add beta && node notes.js list exp…
  after:  run: rm -f notes.json && node notes.js add alpha && node notes.js add beta && node notes.js list exp…
