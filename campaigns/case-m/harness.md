The environment (identical for every hand; environment mechanics only):

- Your working directory __WORK__ is already prepared on this
  machine; run EVERY command exactly so:
  cd __WORK__ && . ./ENV && sh -lc '<command>'
  (the ENV file with this cell's environment exports already lies
  in __WORK__; do not run a single command without this prefix).
- The external network is closed by rule, not by a sandbox: do not
  try to reach anything outside the cell. Everything the work needs
  rides the cell as real bytes — the brief's pinned libraries
  included (the module cache named in ENV); a package fetch or a
  remote host is never part of this work.
- The cell carries two service files named by ENV — PG_SERVICE and
  REDIS_SERVICE: run them as commands (`sh "$PG_SERVICE" start`,
  `sh "$REDIS_SERVICE" start`; both idempotent, both with start,
  stop and status) to bring up the cell's own PostgreSQL and Redis.
  They live on unix sockets inside the cell's scratch (the ENV
  exports PGHOST, PGPORT, PGUSER, PGDATABASE, PGDATA, REDIS_SOCK),
  their state stays inside the cell, and no address outside the
  cell exists for them; they are platform commands, not the
  product's code.
- The cell's loopback listeners are ENV-named too: the product's
  HTTP face takes the port SERVE_PORT, and a conduit cell's route
  listeners and bench target take INGRESS_PORT, EGRESS_PORT and
  TARGET_PORT — no listener of yours ever takes a port of its own
  choosing, so concurrent cells never meet on the loopback.
- The task's inputs lie in __INPUT__ and are read only: the brief
  brief.md and the materials material/.
- WRITE BOUNDS (absolute): write ONLY inside your own cell — the
  working directory __WORK__ and the temporary directory TMPDIR
  from ENV; every file of any program of yours (builds, tests,
  scratches, caches) lives there. Writing to /tmp, to the home
  directory, to any other host directory — forbidden; the foot must
  not leave the cell by a single byte.
- READ BOUNDS (absolute): your reading scope is your own cell — its
  working directory and its input directory — nothing else on the
  host. A program or a tool you call is invoked as a command and
  never inspected: its binary is not documentation.
- You are alone: spawn no one — no subagents and no agentic tools.
  External agentic CLIs of any kind are forbidden.
