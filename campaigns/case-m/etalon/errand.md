# Etalon: errand — a task-queue service with a durable ledger

The meaning's yardstick, written into the campaign before any run:
every hand of errand, in every stack, is judged against these same
lines — the instrument is one; no product, no hand and no judge
forms it. Each line is one requirement of the brief with the run
that proves it.

1. The product is two cooperating processes — `serve` the HTTP
   face, `worker` the runner of tasks — with the README naming
   exactly one build command, one `migrate` command and the two run
   commands — proof: read the README, find the four names, nothing
   else needed.
2. `serve --listen host:port` answers HTTP there; `worker` takes no
   arguments; a stop signal is a calm shutdown of both, the last
   line reports the stop, a task being run is finished or honestly
   returned to waiting — proof: stop the worker under a `slow`
   task, read the last line and the task's after-state.
3. The API never runs a task: enqueue writes the ledger row and
   pushes the queue; execution belongs to the worker — proof: a
   task enqueued with the worker stopped stays waiting; the code's
   process split read.
4. The store and the cache run inside the cell by the ENV-named
   service files, reached only through the ENV-named sockets —
   proof: run the two service scripts, connect a store client
   through `PGHOST`, a queue client through `REDIS_SOCK`.
5. Three handlers are registered, exactly these: `echo` joins the
   arguments by single spaces; `fail` always fails with a fixed
   English reason; `slow` sleeps a second then echoes — proof:
   one task of each run and their outcomes read.
6. `GET /handlers` lists the names with a one-line description
   each; an unknown handler is refused at enqueue — proof: read
   the registry; enqueue `nope` and read the 422.
7. `POST /queues/{name}/tasks` enqueues → 202 with the identifier
   in the standard textual UUID form; a new task is waiting; args
   are a list of strings (anything else → 422 naming the field);
   an unknown handler → 422; an empty queue name → 422 — proof:
   enqueue with good args, crooked args, an unknown handler and an
   empty name; read the codes and bodies.
8. A queue is never created as a separate act: it comes into being
   with the first enqueue, its name is the prefix of its keys on
   the server; `GET /queues` lists the queues that came into being
   — proof: enqueue into a fresh name, read the queue list, read
   the server's keys through the socket.
9. `GET /tasks/{id}` shows state (waiting, running, done, failed,
   cancelled), handler, args, attempts; unknown → 404 — proof:
   read a task across its life; ask for a made-up id.
10. `GET /tasks/{id}/result` returns the result string; a repeat
    returns the same string (never taken away); no record → 404;
    not done → 409 naming the state — proof: fetch twice and
    compare; fetch unknown; fetch while waiting.
11. `GET /queues/{name}/tasks` lists the queue's tasks newest first
    with states — proof: read after several enqueues.
12. `GET /queues/{name}/stats` answers waiting now, done, failed —
    proof: read against the queue's task list.
13. `POST /tasks/{id}/cancel` cancels a waiting task → 202; a
    running, done or failed task → 409 naming the state — proof:
    cancel a waiting task (worker stopped), then a done one.
14. `POST /tasks/{id}/retry` re-queues a failed or cancelled task
    with attempts reset → 202; waiting, running or done → 409 —
    proof: retry a failed task, read its attempts; retry a done
    one and read the 409.
15. `GET /dead-letter` lists the tasks that failed their last
    attempt — proof: run a `fail` task to exhaustion, read the
    list.
16. `DELETE /queues/{name}` purges: waiting tasks become cancelled
    → 204; unknown name → 404; `GET /health` answers 200 with the
    API state and the worker's last heartbeat, a silent worker
    named in plain words; a wrong method, unknown route or broken
    body is an honest 4xx with a one-line reason — proof: purge a
    queue holding waiting tasks and read their states; health read
    twice around a worker stop; one wrong method provoked.
17. A task gets at most three attempts: failures retry to the
    third, then it stands failed — the dead letter; attempts are
    counted in the ledger — proof: a `fail` task's history read:
    three runs, attempts 3, state failed, present in the dead
    letter.
18. The worker takes waiting tasks one at a time and brings each
    to its end; which task leaves first of several waiting is not
    promised — proof: enqueue three, start the worker, all three
    stand done, the order is the product's own.
19. An unreachable or crooked Redis address returns an honest error
    with the reason, without crashing — proof: the code's error
    path read; the queue stopped and the API still answers its
    honest 503.
20. No task state lives in the program's memory: waiting lives on
    the server, every state and result in the ledger — proof: the
    code's store layout read.
21. The ledger owns the exact table `tasks` (UUID primary key,
    queue, handler, args, state, attempts, result, error,
    timestamps) — proof: a store client lists the table and its
    rows match the API's answers.
22. Versioned migrations applied by `migrate`: fresh store → full
    schema; repeat apply → no-op — proof: run it twice, read the
    second run's silence.
23. The store is reached only through the ENV socket; no schema
    change outside migrations — proof: the code read.
24. Redis carries the queues and the hot status reads: fresh from
    the cache, truth from the ledger — proof: the code's read path
    read; a status answered while the ledger is asked only on a
    miss.
25. Redis is transport, never the truth: stopped and started again,
    the ledger stands whole and the worker's startup sweep re-queues
    every task still waiting — no task lost, no done task re-run —
    proof: enqueue with the worker stopped, stop and start the
    cache service, start the worker, every waiting task runs and no
    done task runs twice (their attempts unchanged).
26. The invariant: a task's terminal state and its result are
    written in one atomic step — done with its result, failed with
    its error and attempt count; never done without a result, never
    a result without done; the identifier unique forever — proof:
    the ledger rows read across many tasks; the transaction's
    footprint read in the code.
27. The layout names three boundaries — domain (enqueue, the run,
    the retry policy), transport, storage — as separate modules;
    the domain knows neither HTTP, nor SQL, nor Redis — proof: the
    tree read, the domain's imports read.
28. The platform comes up by the two service files, then `migrate`,
    `serve`, `worker` — in that order the product answers — proof:
    a cold start from scratch, cell scratch empty.
29. There is no external material, as ever: the bench speaks HTTP
    and the handlers' names — proof: the bench run read.
30. The bench drives the whole law: `echo` ["a", "b"] → "a b"
    standing, a repeat fetch identical; `slow` takes its second;
    `fail` exhausts three attempts into the dead letter, `retry`
    revives it to the same end; a cancel of a waiting task holds;
    the stats match the ledger; Redis stopped and started loses no
    waiting task (the sweep re-drives them); the empty name and the
    unknown handler refused 422; a full restart keeps every fact —
    proof: the bench run's output read.
31. Every refusal is brief, in English, with the path and the
    reason; the codes honest (404/409/422/503 by their meanings) —
    proof: each refusal provoked once, its body read.
32. No network beyond the cell's loopback and the ENV sockets; no
    reading or writing outside the cell; no state outside the
    ledger and the queue server; the API never runs a task; a
    result never consumed on read; a done task never re-run; no
    migration change after the first apply — proof: the code read
    plus the runs' behavior.
