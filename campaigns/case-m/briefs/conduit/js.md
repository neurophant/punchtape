# Brief: conduit — a managed two-story tunnel service
The tunnel now carries many routes at once, and I manage it from
my chair instead of a terminal. So the two-story tunnel becomes a
service: I register a route — an ingress address, an egress
address, a target — through an API; a forwarder worker brings the
nodes up and keeps them; the event lines and every chain land in a
ledger; the live counters stay hot. The old law stands: the
ingress knows nothing about the target, the egress is the only one
that knows it, and bytes cross unchanged.

## Stack and tools — my requirement, not the executor's choice

Node.js (an ESM module, the system node). The sense's own: the
built-in node:net (the nodes are as they always were). The
service: Fastify 5.12.7 (the HTTP face), pg 8.23.1 (the store),
redis 6.3.1 (the cache and queue). The node_modules of the work
directory carries exactly these and no package is fetched.

## The service and its two processes

1. The product is two cooperating processes: `serve` — the HTTP
   face (the control plane) — and `worker` — the forwarder (the
   data plane that runs the nodes). The delivery's README names
   exactly one build command (if the stack needs one), one
   `migrate` command, and the two run commands; nothing else is
   needed to run the product.
2. `serve` takes `--listen host:port` and answers HTTP on it;
   `worker` takes no arguments. A stop signal (interrupt or
   terminate) is a calm shutdown for both: the last line reports
   the stop, the running tunnels' nodes go down with it.
3. Inside the forwarder the two node kinds are built and configured
   separately, and the old separation is the point: the ingress is
   handed ONLY the egress's address; the egress is handed ONLY the
   target's address. A node has no listening address of its own
   unless one is given.
4. The store (PostgreSQL 18.6) and the cache (Redis 8.0.5) run
   inside the cell: the service files named by ENV (`PG_SERVICE`,
   `REDIS_SERVICE`) bring them up before the processes start; the
   store rides the unix socket directory `PGHOST` (with
   `PGPORT`, `PGUSER`, `PGDATABASE`), the cache the socket
   `REDIS_SOCK`. No other addresses exist for the product; a
   tunnel's own addresses are host:port on the cell's loopback.

## The API — eleven endpoints, honest codes

5. `POST /tunnels` with `{"name": "...", "ingress_listen":
   "127.0.0.1:$INGRESS_PORT", "egress_listen":
   "127.0.0.1:$EGRESS_PORT", "target": "127.0.0.1:$TARGET_PORT"}`
   (the three ports this cell's ENV names) registers a stopped
   route → 201. A taken name → 409; an address that is not
   host:port → 422 naming the field.
6. `GET /tunnels` lists the routes with their states (stopped,
   starting, running, stopping).
7. `GET /tunnels/{name}` shows one route whole.
8. `PATCH /tunnels/{name}` with a new `target` changes a stopped
   route → 200; a running route → 409.
9. `DELETE /tunnels/{name}` removes the route with its events and
   chains (a running one is stopped first) → 204; unknown → 404.
10. `POST /tunnels/{name}/start` orders the forwarder to bring the
    nodes up → 202; already running → 409.
11. `POST /tunnels/{name}/stop` orders a calm stop of the nodes →
    202; not running → 409.
12. `GET /tunnels/{name}/events` returns the recorded diagnostics
    lines — "up" with the listening address, "connection came" and
    "connection went" with the peer's address — oldest first.
13. `GET /tunnels/{name}/chains` returns the closed chains of the
    route: peer, opened at, closed at — oldest first.
14. `GET /tunnels/{name}/stats` answers the live counters from the
    hot path: chains opened total, clients connected now.
15. `GET /health` answers 200 with the API's state, the
    forwarder's last heartbeat and the number of running tunnels;
    a forwarder silent for over a minute is said so in plain
    words. A wrong method, an unknown route or a broken JSON body
    is an honest 4xx with a one-line reason.

## The forwarding — the conduit law, unchanged

16. The ingress accepts any connection as is: no handshake; the
    client's first byte is already payload.
17. For every client the ingress opens its own separate
    connection to the egress — one client = one chain, never
    multiplexed; the egress, on receiving a connection, connects
    to the target and ties it to that connection.
18. Bytes travel both ways unchanged and without buffering:
    chunks pass on as soon as they are read. A break on any of the
    three legs closes that client's whole chain; half-close is not
    supported. Many clients at once never mix their streams.
19. If the target is unreachable at the moment a client connects,
    that client's chain closes at once and the nodes live on;
    every new client is a fresh attempt. A hung connection to the
    next node or to the target must not disturb other clients.
20. The ingress need not wait for the egress: it comes up without
    one and opens the connection when a client arrives.
21. Every diagnostics line of the nodes — "up", "connection
    came", "connection went" — is recorded in the ledger with the
    route's name, exactly as printed.

## The store of record

22. PostgreSQL is the system of record, with the exact tables
    `tunnels` (name unique, ingress_listen, egress_listen, target,
    state), `events` (id, tunnel, line, at), `chains` (id, tunnel,
    peer, opened_at, closed_at); the columns beyond these are
    yours, the three names are not.
23. The delivery carries versioned migrations (an ordered
    `migrations/` artifact) applied by its `migrate` command: a
    fresh store gets the full schema, a repeat apply is a no-op.
24. The store is reached only through the ENV-named socket; no
    other database is touched, and no schema change happens
    outside the migrations.

## The hot path

25. Redis carries the command queue (start and stop orders from
    the API to the forwarder) and the live counters: every opened
    chain increments, every closed client decrements the now-count;
    `stats` reads them hot — never a store round-trip.
26. Redis is never the record: flushed and restarted, every fact
    (routes, events, chains) stands in the store, the counters
    rebuild from it, and the service keeps working.

## The invariant

27. One transactional law protects the ledger: a closed chain is
    written exactly once — its chains row, its "connection went"
    event row and the tunnel's total-chains count grow in one
    atomic step; the count never names a chain the ledger does not
    hold, and no chain is recorded twice.

## Architecture

28. The delivery's layout names three boundaries — domain (the
    forwarding law: the two node kinds and their separation of
    knowledge), transport (the HTTP face), storage (the store and
    cache gateways) — as separate modules or directories; the
    domain code knows neither HTTP, nor SQL, nor Redis.

## How I use it

29. First the platform: `sh "$PG_SERVICE" start` and
    `sh "$REDIS_SERVICE" start` (both idempotent; their state
    lives in the cell's scratch). Then `migrate`, then `serve
    --listen 127.0.0.1:"$SERVE_PORT"` (the port this cell's ENV
    names), then `worker` — in that order, and the product
    answers.

## Test material and bench

30. There is no external material, as ever: the bench brings up an
    echo target (a server that answers with whatever it received)
    on the cell's ENV-named target port and drives clients
    through the chain ingress → egress → target.
31. The bench drives: a registered route starts and answers "up"
    lines; bytes cross both ways unchanged; two clients at once
    never receive each other's bytes; the target stopped and
    brought back up leaves the nodes alive; a stopped-then-started
    target serves new clients; every closed client has exactly one
    chains row and its "connection went" line; the stats counters
    match the ledger; a start order given while the forwarder is
    down runs when it returns; after a full restart of everything
    the facts stand.

## Limits and errors

32. Every refusal is brief, in English, with the path and the
    reason; the codes are honest: 404 an unknown route; 409 a
    taken name, a start of the running, a stop of the stopped, a
    patch of the running; 422 an address that is not host:port
    with the field named; 503 only when the store or the cache is
    truly unreachable, with the reason.

## What must not happen

33. No network beyond the cell's loopback and the ENV-named
    sockets; no reading or writing outside the cell; no state
    outside the store and the cache; the API never forwards a
    byte; the ingress code path never sees the target's address, 
    the egress code path never sees the clients'; a migration
    change after the first apply.
