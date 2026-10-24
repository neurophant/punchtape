# Brief: conduit — a two-story tunnel

I need to accept other people's connections in plain sight without
showing where they really go. So, two nodes: the ingress stands on the
perimeter and knows nothing about the target; the egress is hidden away
and is the only one that knows the target. The ingress throws
connections to the egress, and the egress carries them to the target.

## Stack and tools — my requirement, not the executor's choice

C++17, built with g++. POSIX system headers only (sockets, epoll); no
third-party libraries.

## How I use it

    conduit ingress --listen address --forward address
    conduit egress  --listen address --target address

Every address is a "host:port" (for example 127.0.0.1:9001). The
ingress must be given the egress's address, and the egress the target's
address; a node has no listening address of its own unless one is
given. Flags are written as "--name value". Both nodes are long-lived
processes: they come up, listen, and keep working until I stop them.

A node reports events in lines: "up" — with its listening address;
"connection came" and "connection went" — with the address of whoever
connected. All such lines go to the diagnostics stream (stderr); the
ordinary output (stdout) stays empty. A timestamp at the start of a
line is fine.

## How it must work

- The ingress accepts any connection as is: no handshake; the client's
  first byte is already payload.
- For every client the ingress opens its own separate connection to the
  egress — ingress→egress connections are not multiplexed; one client =
  one chain.
- The egress, on receiving a connection, connects to the target and
  ties it to that connection.
- Bytes travel both ways unchanged and without buffering: chunks are
  passed on as soon as they are read.
- A break on any of the three legs (client—ingress, ingress—egress,
  egress—target) closes that client's whole chain; half-close (sent
  everything and waiting for a reply) is not supported.
- There can be many clients at once; their streams never mix.
- If the target is unreachable at the moment a client connects, that
  client's chain closes at once and the nodes live on; every new client
  is a fresh attempt.
- The ingress need not wait for the egress: it comes up without one and
  opens the connection when a client arrives.
- A hung connection to the next node or to the target must not disturb
  other clients.
- Stopping on a signal (either "interrupt" or "terminate") is a calm
  shutdown, not an error; the last line reports the stop.

## Test material and bench

There is no external material: the bench brings up an echo target — a
server that answers with whatever it received — and drives clients
through the chain ingress → egress → target. The echo is enough for
everything: accuracy both ways (what was sent is what came back),
isolation (two clients at once never receive each other's bytes), and
the target being stopped and brought back up (the nodes survive it).

## Limits and errors

The ingress does not know the target's address, and the egress does not
know the clients' addresses — that is the point of the design, and it
must not be broken. A start with an unknown role or option, without a
required address, or on an already-taken port is a refusal: the process
states the reason and does not start, without crashing the node already
working on that port.

## What must not happen

Altering bytes, buffering a whole message before passing it on, mixing
clients' chains, everything falling over because of one dropped client
or an unreachable target.
