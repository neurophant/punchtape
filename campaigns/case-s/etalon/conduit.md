# Etalon: conduit — a two-story tunnel

The meaning's yardstick, written into the campaign before any run:
every hand of conduit, in every stack, is judged against these same
lines — the instrument is one; no product, no hand and no judge
forms it. Each line is one requirement of the brief with the run
that proves it.

1. Two roles, long-lived processes: `conduit ingress --listen A
   --forward B` and `conduit egress --listen A --target C`; every
   address is "host:port"; flags are written as "--name value" —
   proof: bring both nodes up and watch them work until stopped.
2. Events in lines: "up" with its listening address, "connection
   came" and "connection went" with the address of whoever
   connected; diagnostics go to stderr, stdout stays empty; a
   leading timestamp is fine — proof: a client's life seen in the
   streams.
3. The ingress accepts any connection as is: no handshake; the
   client's first byte is already payload — proof: a client that
   sends immediately upon connecting is served.
4. For every client the ingress opens its own separate connection to
   the egress; one client = one chain, never multiplexed — proof:
   two clients at once and the egress's "connection came" lines.
5. The egress, on receiving a connection, connects to the target and
   ties it to that connection — proof: a byte sent by a client
   reaches the target and the reply returns.
6. Bytes travel both ways unchanged and without buffering: chunks
   are passed on as soon as they are read — proof: the code's
   passing plus an echo round-trip of a short and a longer message.
7. A break on any of the three legs (client—ingress,
   ingress—egress, egress—target) closes that client's whole chain;
   half-close is not supported — proof: kill a leg and watch that
   client's chain close.
8. There can be many clients at once; their streams never mix —
   proof: two clients exchanging different bytes at the same time,
   each receives exactly its own.
9. If the target is unreachable at the moment a client connects,
   that client's chain closes at once and the nodes live on; every
   new client is a fresh attempt — proof: stop the target, connect,
   bring it back, connect again.
10. The ingress need not wait for the egress: it comes up without
    one and opens the connection when a client arrives — proof:
    start the ingress first, the egress after, a client in between
    survives or retries to success.
11. A hung connection to the next node or to the target must not
    disturb other clients — proof: a stuck leg and a second healthy
    client.
12. Stopping on a signal ("interrupt" or "terminate") is a calm
    shutdown, not an error; the last line reports the stop — proof:
    SIGINT and SIGTERM on both nodes.
13. The ingress does not know the target's address, and the egress
    does not know the clients' — the design's point, never broken —
    proof: the two command lines and the code's knowledge.
14. A start with an unknown role or option, without a required
    address, or on an already-taken port is a refusal: the process
    states the reason and does not start, without crashing the node
    already working on that port — proof: each wrong start, with a
    healthy node on the port.
15. The bench is an echo target: what was sent is what came back,
    both ways — proof: the echo round-trips of the whole chain.
