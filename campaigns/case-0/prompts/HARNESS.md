The environment (identical for every hand; environment mechanics only):

- Your working container __CONTAINER__ is already running; every
  command goes through docker exec (example:
  docker exec -w /run/task __CONTAINER__ sh -lc '<command>').
- The network in the container is off; do not try to reach anything
  outside it.
- On the host the directory __HOSTDIR__ is the mounted /run/task:
  read and write only it and nothing else on the host.
- Your round is your turn: having finished the turn (a question batch /
  an escalation at code 3/4, or completed work), END THE TURN with a
  final message; the operator's answer arrives as your next message —
  continue the work from it, one continuous run.
- You are alone: spawn no one — no subagents and no agentic tools.
  External agentic CLIs of any kind are forbidden.
