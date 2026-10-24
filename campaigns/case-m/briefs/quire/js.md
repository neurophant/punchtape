# Brief: quire — a page service from uploaded sources
I have outgrown one folder: my sources are many, they change, and I
want them served, not rebuilt by hand. So the page builder becomes
a service. I hand it my source folder through an API; a worker
builds the page exactly the way quire always did — texts become
lines, data become substitutions, everything else is copied byte
for byte — every build lands in a ledger, and the last good page
is served hot. The old law stands: my folder's contents are all
the configuration there is.

## Stack and tools — my requirement, not the executor's choice

Node.js (an ESM module, the system node). The sense's own: marked
17.0.5 (markdown → HTML), yeahjs 0.3.1 (templates; fields are
visible through the locals object), @mourner/yeahml 1.0.0 (data).
The service: Fastify 5.12.7 (the HTTP face), pg 8.23.1 (the
store), redis 6.3.1 (the cache and queue). The template is
`index.ejs`. The node_modules of the work directory carries
exactly these and no package is fetched.

## The service and its two processes

1. The product is two cooperating processes: `serve` — the HTTP
   face — and `worker` — the background builder. The delivery's
   README names exactly one build command (if the stack needs
   one), one `migrate` command, and the two run commands; nothing
   else is needed to run the product.
2. `serve` takes `--listen host:port` and answers HTTP on it;
   `worker` takes no arguments. A stop signal (interrupt or
   terminate) is a calm shutdown for both: the last line reports
   the stop, unfinished work stays honestly queued.
3. The API process never renders a page: every build is queued
   through Redis and done by the worker; the API answers from the
   store of record and the hot cache only.
4. The store (PostgreSQL 18.6) and the cache (Redis 8.0.5) run
   inside the cell: the service files named by ENV (`PG_SERVICE`,
   `REDIS_SERVICE`) bring them up before the processes start; the
   store rides the unix socket directory `PGHOST` (with
   `PGPORT`, `PGUSER`, `PGDATABASE`), the cache the socket
   `REDIS_SOCK`. No other addresses exist for the product.

## The API — thirteen endpoints, honest codes

5. `POST /sites` with a JSON body `{"name": "..."}` creates a
   site → 201 with its record. A taken name → 409; a crooked name
   (empty, longer than 64) → 422 naming the field.
6. `GET /sites` lists the sites, oldest first.
7. `GET /sites/{id}` shows the site's record and the state of its
   newest build.
8. `DELETE /sites/{id}` removes the site with its source, builds
   and artifacts → 204; unknown id → 404.
9. `PUT /sites/{id}/source` uploads one source bundle — the quire
   folder as files (multipart or the stack's honest equivalent):
   the markdown texts, `data.yml`, the one index template, the
   extraneous files, nested directories plain. → 201 with the new
   revision number; a second upload replaces the bundle whole, it
   never merges.
10. `GET /sites/{id}/source` lists the current bundle's files with
    sizes; `GET /sites/{id}/source/{path}` returns one bundle
    file's bytes.
11. `POST /sites/{id}/builds` queues a build → 202 with the
    build's number (per site, starting at 1).
12. `GET /sites/{id}/builds` lists the site's builds, newest
    first, with their states.
13. `GET /sites/{id}/builds/{n}` shows the build: its state, and
    when done the "ok" listing — one "ok" line per created file,
    the stable order, the root's files alphabetically.
14. `GET /sites/{id}/page` serves the index.html of the site's
    current good build; no good build yet → 404.
15. `GET /sites/{id}/files/{path}` serves a copied file of the
    current good build's output byte for byte; anything outside
    the output → 404.
16. `GET /health` answers 200 with the API's state and the
    worker's last heartbeat; a worker silent for over a minute is
    said so in plain words.
17. A wrong method, an unknown route or a broken JSON body is an
    honest 4xx with a one-line reason; a client's mistake is never
    a 500.

## The build — the quire law, unchanged

18. The build is the old law over the uploaded bundle: markdown
    files in the root are the entries; front matter between `---`
    lines is the entry's data, the rest is the body and turns
    into HTML; `data.yml` in the root is the data under the name
    `data`; a missing file means no data at all.
19. The one template — a root file named `index` with the
    template engine's extension named in the stack requirements —
    builds into `index.html`, with `data` and the entries
    available to it: each entry carries its front-matter data and
    its body's ready HTML.
20. The entries reach the template as a list in a stable order,
    by file name alphabetically.
21. Everything that is not a text, not `data.yml` and not the
    template is copied into the output byte for byte; nested
    directories are copied plain, their contents not processed;
    the texts, `data.yml` and the template themselves do not land
    in the output.
22. A bundle without a single text and without data is legal: the
    template builds with an empty list of entries.
23. A bundle without a template fails honestly — "there is
    nothing to build" — as a recorded failed build with that
    reason; YAML that does not parse, a template that does not
    compile or breaks on substitution fail the same way: failed
    state, brief English reason, and the site's last good page
    stays untouched.
24. Two builds of one unchanged revision give byte-identical
    output files, to the letter.
25. For every created file the build records an "ok" line with
    the path — the listing shown by endpoint 13.

## The store of record

26. PostgreSQL is the system of record, with the exact tables
    `sites` (the sites), `sources` (each site's current bundle,
    file for file), `builds` (one row per build: number, state,
    detail), `artifacts` (the output files of each good build);
    the columns beyond these are yours, the four names are not.
27. The delivery carries versioned migrations (an ordered
    `migrations/` artifact) applied by its `migrate` command: a
    fresh store gets the full schema, a repeat apply is a no-op.
28. The store is reached only through the ENV-named socket; no
    other database is touched, and no schema change happens
    outside the migrations.

## The hot path

29. Redis carries the build queue — the API pushes build ids, the
    worker takes them; it is the ONLY channel between the two
    processes.
30. `page` and `files` are served through a Redis cache of the
    current good build's artifacts; a fresh good build swaps the
    cached set whole — no stale byte is served after the swap.
31. Redis is never the record: flushed and restarted, every fact
    (sites, sources, builds, artifacts) stands in the store and
    the service keeps working.

## The invariant

32. One transactional law protects the ledger: a build turns ok
    together with ALL its artifact rows and the site's
    current-build pointer — one atomic step; a reader never sees
    a half-built set, and after a failure the pointer still names
    the previous good build.

## Architecture

33. The delivery's layout names three boundaries — domain (the
    quire law: parsing, templating, the build), transport (the
    HTTP face), storage (the store and cache gateways) — as
    separate modules or directories; the domain code knows
    neither HTTP, nor SQL, nor Redis.

## How I use it

34. First the platform: `sh "$PG_SERVICE" start` and
    `sh "$REDIS_SERVICE" start` (both idempotent; their state
    lives in the cell's scratch). Then `migrate`, then `serve
    --listen 127.0.0.1:"$SERVE_PORT"` (the port this cell's ENV
    names), then `worker` — in that order, and the product
    answers.

## Test material

35. The material directory holds: the stack's own bundle
    (`quire/<stack>/`: `data.yml` with a couple of keys,
    `about.md` without front matter, an entry with front matter,
    the stack's index template, one extraneous file),
    `bundle-invalid/` (texts and data but no template — a build
    that must fail honestly) and `bundle-empty/` (only the index
    template — an empty entry list).
36. The bench drives: the stack bundle builds ok with the
    expected listing and a byte-honest `page`; the invalid bundle
    fails with the reason and the previous page stands; a rebuild
    is byte-identical; after a new revision the old artifacts are
    gone whole; a build queued while the worker is down runs when
    the worker returns; two sites build concurrently without
    interference; after a full restart of everything the facts
    stand.

## Limits and errors

37. Every refusal is brief, in English, with the path and the
    reason; the codes are honest: 404 an unknown site, build or
    file; 409 a taken name; 422 a crooked payload with the field
    named; 503 only when the store or the cache is truly
    unreachable, with the reason.

## What must not happen

38. No network beyond the cell's loopback and the ENV-named
    sockets; no reading or writing outside the cell; no state
    outside the store and the cache; the API never renders; an
    artifact of a failed build is never served; a migration
    change after the first apply.
