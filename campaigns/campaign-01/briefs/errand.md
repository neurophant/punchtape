# Brief: errand — a background task-queue library

My program sometimes has work that must not be done at the moment the
user asks for it: too slow, or not urgent. I want a library: I put a
task into a queue, and a separate worker runs it when its turn comes; I
can always ask what has become of it and fetch the result. The store is
Redis (an address like redis://localhost/): tasks survive restarts of
my program. The queue is not created on the server as a separate act —
it comes into being with the first enqueue; the queue's name sets the
prefix of its keys on the server.

## Stack and tools — my requirement, not the executor's choice

Rust (cargo). Crates: redis 0.8, uuid 0.5 (the v4 feature), serde 1.0,
serde_json 1.0, serde_derive 1.0. The product is a library crate; its
working order is proven by integration tests against a running Redis.

## What the library can do

- A queue is defined by the server's address and a name.
- Put a task: a list of string arguments and the record's time-to-live
  before the task starts, in seconds; returns a unique identifier — the
  standard textual form of a UUID. A new task is "waiting". The queue's
  order is fair: whoever lined up first is taken first.
- Ask for the status by identifier. There are five statuses: "waiting",
  "running", "done", "error", "lost". "Not found" is not a status: the
  record is missing or its TTL has run out — the operation returns an
  error.
- Fetch the result by identifier: the result string, if the task is
  "done"; "waiting"/"running"/"error"/"lost" have no result. The result
  is not taken away: a repeat request returns the same string as long
  as its TTL lives.
- The worker loop is a blocking call: it spins until it decides to
  finish; I need to be able to control it — a finite mode (take one
  task, bring it to a status, return control) and an endless one (spins
  until stopped; waiting for new tasks carries its own wait timeout).
  When needed I run the loop in my own thread — enqueuing tasks while
  the loop is running works freely.
- Execution is by my handler: it receives the identifier and the
  arguments and returns the result string. Each task's handler runs in
  its own thread. The handler returned an error — the task is "error",
  and the loop moves on to the others. The handler missed its allotted
  execution deadline — the task is "lost", and the loop moves on.
- The record's TTL on the server changes with the stage: before start,
  the TTL given at enqueue; while executing, the execution deadline
  plus a margin; after the end (done/error/lost), its own result TTL.
  An expired TTL means the record is gone.
- Delete the queue: the queue's list on the server is deleted; records
  of tasks already enqueued live out their TTLs.

## Test material and bench

There is no external material; a running Redis at redis://localhost/ is
needed — the test environment brings it up before the run. The material
is simple: argument strings ("a", "b", "fail"), handlers that return
constants, a handler that fails on "fail", and a handler that
deliberately hangs past the execution deadline — enough for every
status, the ordering, and the limits.

## Limits and errors

If the server is unreachable or the address is crooked, operations
honestly return an error with the reason, without panicking. An empty
queue name cannot be worked with: the very first server operation
returns an error. A task whose pre-start TTL ran out before the loop
took it vanishes from the queue without a trace — that is honest
behavior, not a loss.

## What must not happen

Keeping task state in my program's memory (tasks live on the server),
blocking enqueues while the loop runs, consuming the result on read,
mixing up "lost" with "error".
