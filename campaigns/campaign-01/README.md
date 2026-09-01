# Campaign 01 — manual mode (2026-08-27)

The first field run of the battery: five tasks (one per sense), each
executed once, by hand. The operator stages the environment, launches
one agent per task, and takes the evidence; the machine does the rest.

## The machine

Built from the working tree at the moment of the run (the commit is
recorded with the run evidence), built twice and compared
byte-for-byte, installed into `~/.local/bin`.

## Tasks and environments

One task — one agent. The stack is pinned by the brief; the
environment stands outside the machine:

| Task | Stack (pinned by the brief) | Executor environment |
|---|---|---|
| brim | C++17, g++, GraphicsMagick++ | container (ubuntu:24.04, g++ 13.3, GraphicsMagick 1.3.42), the world mounted at /work |
| conduit | Go, stdlib | host: go |
| errand | Rust: redis, uuid, serde | host: cargo; a redis container for the test run, alive only for the task |
| pressmark | Python 3: PyYAML, Markdown, Jinja2, pydantic, python-slugify (versions pinned) | container (python:3.12-slim, versions pinned) |
| quire | Node.js ESM: marked, yeahjs, @mourner/yeahml | host: node, dependencies pre-installed into the task world |

Container images live locally; after the task the container is stopped
and removed (idle containers are garbage).

## Materials

brim works on ten images from the author's enimda image repository
(five fixed JPEGs, five animated GIFs), downloaded by the operator
into the task world before the launch — the same ten files for every
run. The other tasks need no external materials (their stands live in
the world).

## Briefs

`briefs/<sense>.md` — the five wishes exactly as the executors got
them.

## Lessons that shaped campaign-02

- Expectations must be derived from the material by measurement, not
  copied from the original tool's own tests: numbers that exist in one
  implementation's entropy profiles may not exist in another's.
- Interactive containers need stdin wired (`docker exec -i`), and the
  interpreter must exist in the image before the first run.
- Boundary conditions travel from the original to the brief or they
  will be re-invented per hand — the machine catches the difference,
  the run pays the extra submission.
