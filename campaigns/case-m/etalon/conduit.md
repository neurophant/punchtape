# Etalon: conduit — a managed two-story tunnel service

The meaning's yardstick, written into the campaign before any run:
every hand of conduit, in every stack, is judged against these same
lines — the instrument is one; no product, no hand and no judge
forms it. Each line is one requirement of the brief with the run
that proves it.

1. The product is two cooperating processes — `serve` the control
   plane, `worker` the forwarder that runs the nodes — with the
   README naming exactly one build command, one `migrate` command
   and the two run commands — proof: read the README, find the four
   names, nothing else needed.
2. `serve --listen host:port` answers HTTP there; `worker` takes no
   arguments; a stop signal is a calm shutdown of both, the last
   line reports the stop, the running nodes go down with it —
   proof: start a tunnel, SIGTERM the worker, read the last line
   and the dead listeners.
3. Inside the forwarder the node kinds are separate: the ingress is
   handed only the egress address, the egress only the target —
   proof: the code's wiring read; the ingress path mentions no
   target, the egress path no client address.
4. The store and the cache run inside the cell by the ENV-named
   service files, reached only through the ENV-named sockets; a
   tunnel's own addresses are loopback host:port — proof: run the
   service scripts, register a route, read its three addresses.
5. `POST /tunnels` registers a stopped route → 201; a taken name →
   409; an address not host:port → 422 naming the field — proof:
   three posts, read the codes and the bodies.
6. `GET /tunnels` lists the routes with states — proof: read after
   registering several.
7. `GET /tunnels/{name}` shows one route whole — proof: read it
   against the registration's answer.
8. `PATCH /tunnels/{name}` with a new target changes a stopped
   route → 200; a running one → 409 — proof: patch stopped, start,
   patch again.
9. `DELETE /tunnels/{name}` removes the route with events and
   chains (a running one stopped first) → 204; unknown → 404 —
   proof: delete a running route; the listener dies, the store rows
   are gone.
10. `POST /tunnels/{name}/start` orders the nodes up → 202; already
    running → 409 — proof: start twice, read the second code.
11. `POST /tunnels/{name}/stop` orders a calm stop → 202; not
    running → 409 — proof: stop a stopped route, read the code.
12. `GET /tunnels/{name}/events` returns the recorded lines — "up"
    with the listening address, "connection came" and "connection
    went" with the peer — oldest first — proof: drive one client,
    read the events in order.
13. `GET /tunnels/{name}/chains` returns the closed chains: peer,
    opened at, closed at — oldest first — proof: the same drive,
    the row read.
14. `GET /tunnels/{name}/stats` answers the live counters — chains
    opened total, clients connected now — proof: read during two
    open clients and after their close.
15. `GET /health` answers 200 with the API state, the forwarder's
    last heartbeat and the count of running tunnels; a forwarder
    silent for over a minute is said so; a wrong method, unknown
    route or broken body is an honest 4xx with a one-line reason —
    proof: health read twice around a forwarder stop; one wrong
    method provoked.
16. The ingress accepts any connection as is: no handshake, the
    first byte is payload — proof: a raw client sends immediately,
    the echo answers.
17. One client = one chain: a separate ingress→egress connection
    per client, never multiplexed; the egress ties it to one target
    connection — proof: two clients at once, both answered; the
    code's per-client dial read.
18. Bytes cross unchanged, no whole-message buffering; a break on
    any leg closes that client's whole chain; half-close is not
    supported; many clients never mix streams — proof: an echoed
    payload compared byte for byte; a target killed closes only its
    client's chain; two interleaved clients receive their own bytes
    only.
19. An unreachable target closes that client's chain at once, the
    nodes live on; every new client is a fresh attempt; a hung
    connection does not disturb other clients — proof: stop the
    target, connect (the chain closes, "up" still stands), restart
    the target, connect again.
20. The ingress need not wait for the egress: it comes up without
    one and dials when a client arrives — proof: start a route with
    the egress port closed, the ingress still answers "up"; open
    the egress, a client passes.
21. Every diagnostics line of the nodes — "up", "connection came",
    "connection went" — is recorded in the ledger under the route's
    name, exactly as printed — proof: the events route's lines
    compared with the nodes' printed output.
22. The store owns the exact tables `tunnels` (name unique),
    `events`, `chains` — proof: a store client lists the tables and
    their rows match the API's answers.
23. Versioned migrations applied by `migrate`: fresh store → full
    schema; repeat apply → no-op — proof: run it twice, read the
    second run's silence.
24. The store is reached only through the ENV socket; no schema
    change outside migrations — proof: the code read.
25. Redis carries the command queue and the live counters: every
    opened chain increments, every closed client decrements the
    now-count; `stats` reads them hot — proof: the counters move
    live during a drive; the code's counter transport read.
26. Redis flushed and restarted loses no fact; the counters rebuild
    from the ledger — proof: stop the cache service, start it, read
    the routes, events, chains and stats.
27. The invariant: a closed chain is written exactly once — its
    chains row, its "connection went" event and the tunnel's
    total-chains count grow in one atomic step; the count never
    names a chain the ledger does not hold, no chain twice —
    proof: after a multi-client drive the count equals the chains
    rows; a store client checks the transaction's footprint.
28. The layout names three boundaries — domain (the two node kinds
    and their separation of knowledge), transport, storage — as
    separate modules; the domain knows neither HTTP, nor SQL, nor
    Redis — proof: the tree read, the domain's imports read.
29. The platform comes up by the two service files, then `migrate`,
    `serve`, `worker` — in that order the product answers — proof:
    a cold start from scratch, cell scratch empty.
30. There is no external material: the bench brings up its own echo
    target on the cell's loopback and drives clients through the
    chain — proof: the bench's own echo server seen running in the
    bench output.
31. The bench drives the whole law: "up" lines; bytes both ways
    unchanged; two clients never mixed; the target stopped and
    restarted leaves the nodes alive and serving new clients; every
    closed client has exactly one chains row and its "connection
    went" line; the stats match the ledger; a start order given
    while the forwarder is down runs on its return; a full restart
    keeps every fact — proof: the bench run's output read.
32. Every refusal is brief, in English, with the path and the
    reason; the codes honest (404/409/422/503 by their meanings) —
    proof: each refusal provoked once, its body read.
33. No network beyond the cell's loopback and the ENV sockets; no
    reading or writing outside the cell; no state outside the store
    and the cache; the API never forwards a byte; the ingress path
    never sees the target's address, the egress path never the
    clients'; no migration change after the first apply — proof:
    the code read plus the runs' behavior.
