# Brief: errand — a small background task-queue library

My program sometimes has work that must not be done at the moment the
user asks for it. I want a simpler library: put a task into a queue,
let a separate step run it, and I fetch the result afterwards. The
store is Redis (an address like redis://localhost/): tasks survive
restarts of my program. The queue is not created on the server as a
separate act — it comes into being with the first enqueue; the queue's
name sets the prefix of its keys on the server.

## Stack and tools — my requirement, not the executor's choice

Python 3. Libraries: redis (a Redis client); identifiers come from the
standard uuid module. The product is a library; its working order is
proven by integration tests against a running Redis.

## What the library can do

A queue is defined by the server's address and a name. Three operations
— that is enough:

- Put a task: a list of string arguments; returns a unique identifier —
  the standard textual form of a UUID. A new task is "waiting".
- The worker step is finite: it takes one waiting task, runs it, brings
  it to the end, and returns control. Execution is by my handler: it
  receives the identifier and the arguments and returns the result
  string. After the step the task is "done"; a "waiting" task has no
  result.
- Fetch the result by identifier: the result string. The result is not
  taken away: a repeat request returns the same string. No record — the
  operation returns an error.

## Test material and bench

There is no external material; a running Redis at redis://localhost/ is
needed — the test environment brings it up before the run. The material
is simple: argument strings ("a", "b") and a handler that returns a
constant — enough for the enqueue, the step, the result, and a repeat
request.

## Limits and errors

If the server is unreachable or the address is crooked, operations
honestly return an error with the reason, without panicking. An empty
queue name cannot be worked with: the very first server operation
returns an error. Which exact task the step will get out of several
waiting ones I do not promise; any single one is enough for me.

## What must not happen

Keeping task state in my program's memory (tasks live on the server),
consuming the result on read, losing the result between reads.
