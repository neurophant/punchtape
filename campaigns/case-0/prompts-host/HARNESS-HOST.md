The environment (identical for every hand; environment mechanics only):

- Your working directory __WORK__ is already prepared on this
  machine; run EVERY command exactly so:
  cd __WORK__ && . ./ENV && sh -lc '<command>'
  (the ENV file with this cell's environment exports already lies
  in __WORK__; do not run a single command without this prefix).
- The network is off by directive; do not try to reach anything
  outside the cell.
- The task's inputs lie in __INPUT__ and are read only (the brief
  brief.md, the materials materials/, the expectations list
  EXPECTATIONS.md).
- WRITE BOUNDS (absolute): write ONLY inside your own cell — the
  working directory __WORK__ and the temporary directory TMPDIR
  from ENV; every file of any program of yours (builds, tests,
  scratches, caches) lives there. Writing to /tmp, to the home
  directory, to the repository directory, to any other host
  directory — forbidden and tracked by the audit; the foot must
  not leave the cell by a single byte.
- Your round is your turn: having finished the turn (a question batch /
  an escalation at code 3/4, or completed work), END THE TURN with a
  final message; the operator's answer arrives as your next message —
  continue the work from it, one continuous run.
- You are alone: spawn no one — no subagents and no agentic tools.
  External agentic CLIs of any kind are forbidden.
