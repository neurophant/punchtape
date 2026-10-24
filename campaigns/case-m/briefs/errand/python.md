# Brief: errand — a task-queue service with a durable ledger
My program's odd jobs have outgrown a library call: the work now
comes from everywhere, and I want a counter's window, not a
process. So the little queue becomes a service: tasks go in
through an HTTP face, a worker runs them by registered handlers,
the results wait for me — and every task lives in a ledger that
survives anything. The old law stands: the queue is Redis, a queue
comes into being with the first enqueue, and a result once written
is never taken away.

## Stack and tools — my requirement, not the executor's choice

Python 3. The sense's own: redis 8.1.0 (the queue client);
identifiers come from the standard uuid module. The service:
FastAPI 0.162.0 with uvicorn 0.54.0 (the HTTP face), pg8000 1.31.6
(the ledger). The cell's venv (ENV) carries exactly these and no
package is fetched.

## The service and its two processes

1. The product is two cooperating processes: `serve` — the HTTP
   face — and `worker` — the runner of tasks. The delivery's
   README names exactly one build command (if the stack needs
   one), one `migrate` command, and the two run commands; nothing
   else is needed to run the product.
2. `serve` takes `--listen host:port` and answers HTTP on it;
   `worker` takes no arguments. A stop signal (interrupt or
   terminate) is a calm shutdown for both: the last line reports
   the stop, a task being run is finished or honestly returned to
   waiting — never lost.
3. The API never runs a task: enqueueing writes the ledger row and
   pushes the queue; execution belongs to the worker only.
4. The store (PostgreSQL 18.6) and the cache (Redis 8.0.5) run
   inside the cell: the service files named by ENV (`PG_SERVICE`,
   `REDIS_SERVICE`) bring them up before the processes start; the
   store rides the unix socket directory `PGHOST` (with
   `PGPORT`, `PGUSER`, `PGDATABASE`), the cache the socket
   `REDIS_SOCK`. No other addresses exist for the product.

## The handlers — the customer's registered work

5. The worker runs tasks by handler NAME, from a registry the
   delivery carries; three handlers are registered, exactly
   these: `echo` — returns the arguments joined by single spaces;
   `fail` — always fails with a fixed English reason; `slow` —
   sleeps a second and then echoes like `echo`.
6. The registry is visible: `GET /handlers` lists the names with a
   one-line description each. A task naming an unknown handler is
   refused at enqueue, not at run.

## The API — twelve endpoints, honest codes

7. `POST /queues/{name}/tasks` with `{"handler": "echo",
   "args": ["a", "b"]}` enqueues → 202 with the task's identifier
   — the standard textual form of a UUID, as ever. A new task is
   waiting. Args are a list of strings; anything else → 422 naming
   the field. An unknown handler → 422. An empty queue name → 422
   (the old law: it cannot be worked with at all).
8. A queue is NOT created as a separate act: it comes into being
   with the first enqueue; the queue's name sets the prefix of its
   keys on the server, as ever. `GET /queues` lists the queues
   that came into being.
9. `GET /tasks/{id}` shows the task: state (waiting, running,
   done, failed, cancelled), handler, args, attempts; unknown id →
   404.
10. `GET /tasks/{id}/result` returns the result string; the result
    is not taken away — a repeat request returns the same string;
    no record → 404; not done yet → 409 naming the state.
11. `GET /queues/{name}/tasks` lists the queue's tasks, newest
    first, with states.
12. `GET /queues/{name}/stats` answers the queue's counters:
    waiting now, done, failed.
13. `POST /tasks/{id}/cancel` cancels a waiting task → 202; a
    running, done or failed task → 409 naming the state.
14. `POST /tasks/{id}/retry` re-queues a failed or cancelled task
    with its attempts reset → 202; waiting, running or done → 409.
15. `GET /dead-letter` lists the tasks that failed their last
    attempt (state failed, attempts exhausted).
16. `DELETE /queues/{name}` purges the queue: its waiting tasks
    become cancelled → 204; unknown name → 404. `GET /health`
    answers 200 with the API's state and the worker's last
    heartbeat; a worker silent for over a minute is said so in
    plain words. A wrong method, an unknown route or a broken JSON
    body is an honest 4xx with a one-line reason.

## The run — the errand law, unchanged and deeper

17. A task gets at most three attempts: a run that fails is
    retried until the third failure, then the task stands failed —
    the dead letter. Attempts are counted in the ledger.
18. The worker takes waiting tasks one at a time, runs the
    handler, and brings the task to its end; which exact task
    leaves the queue first out of several waiting is not promised.
19. If the Redis server is unreachable or an address is crooked,
    operations honestly return an error with the reason, without
    crashing the process — as ever.
20. No task state lives in my program's memory: waiting lives on
    the server, every state and result lives in the ledger.

## The store of record

21. PostgreSQL is the system of record — the ledger — with the
    exact table `tasks` (id — the UUID primary key, queue,
    handler, args, state, attempts, result, error, created_at,
    finished_at); the columns beyond these are yours, the name is
    not.
22. The delivery carries versioned migrations (an ordered
    `migrations/` artifact) applied by its `migrate` command: a
    fresh store gets the full schema, a repeat apply is a no-op.
23. The store is reached only through the ENV-named socket; no
    other database is touched, and no schema change happens
    outside the migrations.

## The hot path

24. Redis carries the queues themselves — waiting task ids under
    the queue's key prefix, as ever — and the hot status reads: a
    task's state is answered from the cache when fresh, from the
    ledger when not.
25. Redis is transport, never the truth: stopped and started again
    (the service file does it), the ledger stands whole — at
    startup the worker sweeps the ledger and re-queues every task
    still waiting, so no task is lost and no done task is re-run.

## The invariant

26. One transactional law protects the ledger: a task's terminal
    state and its result are written in one atomic step — done
    with its result, failed with its error and attempt count; a
    task never stands done without its result nor holds a result
    without standing done, and its identifier is unique forever.

## Architecture

27. The delivery's layout names three boundaries — domain (the
    queue law: enqueue, the run, the retry policy), transport (the
    HTTP face), storage (the ledger and queue gateways) — as
    separate modules or directories; the domain code knows
    neither HTTP, nor SQL, nor Redis.

## How I use it

28. First the platform: `sh "$PG_SERVICE" start` and
    `sh "$REDIS_SERVICE" start` (both idempotent; their state
    lives in the cell's scratch). Then `migrate`, then `serve
    --listen 127.0.0.1:"$SERVE_PORT"` (the port this cell's ENV
    names), then `worker` — in that order, and the product
    answers.

## Test material and bench

29. There is no external material, as ever: the bench speaks HTTP
    and the handlers' names; the arguments are plain strings.
30. The bench drives: `echo` with args ["a", "b"] runs and the
    result "a b" stands, a repeat fetch returns it again; `slow`
    takes its second and stands done; `fail` exhausts three
    attempts and lands in the dead letter, `retry` revives it to
    the same end; a cancel of a waiting task (the worker stopped)
    holds; the queue's stats match the ledger; Redis stopped and
    started loses no waiting task — the sweep re-drives them; an
    empty queue name and an unknown handler are refused 422; after
    a full restart of everything the facts stand.

## Limits and errors

31. Every refusal is brief, in English, with the path and the
    reason; the codes are honest: 404 an unknown task or queue;
    409 a fetch of an unfinished result, a cancel or retry of the
    wrong state — the state named; 422 a crooked payload, args or
    handler, the field named; 503 only when the store or the queue
    server is truly unreachable, with the reason.

## What must not happen

32. No network beyond the cell's loopback and the ENV-named
    sockets; no reading or writing outside the cell; no state
    outside the ledger and the queue server; the API never runs a
    task; a result consumed on read; a done task re-run; a
    migration change after the first apply.
