# Etalon: errand — a small background task-queue library

The meaning's yardstick, written into the campaign before any run:
every hand of errand, in every stack, is judged against these same
lines — the instrument is one; no product, no hand and no judge
forms it. Each line is one requirement of the brief with the run
that proves it.

1. A queue is defined by the server's address and a name; the store
   is Redis, so tasks survive restarts of the using program; the
   queue is not created on the server as a separate act — it comes
   into being with the first enqueue; the queue's name sets the
   prefix of its keys — proof: enqueue into a fresh name and read
   the server's keys.
2. Put a task: a list of string arguments; returns a unique
   identifier in the standard textual form of a UUID; a new task is
   "waiting" — proof: an enqueue returns a UUID-looking identifier.
3. The worker step is finite: it takes one waiting task, runs it,
   brings it to the end and returns control; execution is by the
   user's handler — it receives the identifier and the arguments and
   returns the result string; after the step the task is "done" —
   proof: one step with a constant-returning handler returns and a
   second step takes nothing.
4. A "waiting" task has no result — proof: a fetch between enqueue
   and step errors.
5. Fetch the result by identifier: the result string; the result is
   not taken away — a repeat request returns the same string; no
   record — the operation returns an error — proof: fetch twice,
   compare; fetch an unknown identifier.
6. The working order is proven by integration tests against a
   running Redis — proof: the suite runs green in the cell.
7. If the server is unreachable or the address is crooked,
   operations honestly return an error with the reason, without
   panicking — proof: point at a dead address and a malformed one.
8. An empty queue name cannot be worked with: the very first server
   operation returns an error — proof: "" as the name.
9. Which exact task the step takes out of several waiting ones is
   not promised; any single one is enough — proof: enqueue two,
   step, step — both come out, the order is the product's own.
10. No task state in the using program's memory (tasks live on the
    server), the result is not consumed on read, the result is not
    lost between reads — proof: the code's store layout plus the
    repeat fetches.
