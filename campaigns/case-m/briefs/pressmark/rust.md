# Brief: pressmark — a publishing service from record folders
My records are still files and always will be — but the site is no
longer one folder and one command: the records arrive from
everywhere now, the site must always serve the last good
publication, and publishing happens in the background. So the
little site builder becomes a service: I hand it record folders
through an API; a worker publishes the site exactly the way
pressmark always did — one page per record, a shared index, the
address rule untouched — and the pages are served hot. The old law
stands: a record is its own folder, and the store keeps every one.

## Stack and tools — my requirement, not the executor's choice

Rust (cargo). The sense's own: serde_yaml 0.9.34 (descriptions),
comrak 0.42.0 (markdown → HTML, no extensions), tera 1.20.1
(templates), serde 1.0.229 with serde_json 1.0.154. The service:
axum 0.8.9 over tokio 1.54.1 (the HTTP face), tokio-postgres
0.7.18 (the store), redis 0.32.7 with the tokio comp (the cache
and queue). The vendored crate tree of ENV carries exactly these
and no crate is fetched.

## The service and its two processes

1. The product is two cooperating processes: `serve` — the HTTP
   face — and `worker` — the background publisher. The delivery's
   README names exactly one build command (if the stack needs
   one), one `migrate` command, and the two run commands; nothing
   else is needed to run the product.
2. `serve` takes `--listen host:port` and answers HTTP on it;
   `worker` takes no arguments. A stop signal (interrupt or
   terminate) is a calm shutdown for both: the last line reports
   the stop, unfinished work stays honestly queued.
3. The API process never renders a page: every publish is queued
   through Redis and done by the worker; the API answers from the
   store of record and the hot cache only.
4. The store (PostgreSQL 18.6) and the cache (Redis 8.0.5) run
   inside the cell: the service files named by ENV (`PG_SERVICE`,
   `REDIS_SERVICE`) bring them up before the processes start; the
   store rides the unix socket directory `PGHOST` (with
   `PGPORT`, `PGUSER`, `PGDATABASE`), the cache the socket
   `REDIS_SOCK`. No other addresses exist for the product.

## The API — thirteen endpoints, honest codes

5. `POST /records` uploads one record folder (multipart or the
   stack's honest equivalent): the `meta.yaml` description, the
   full and short texts, the pictures. → 201 with the record's
   address. A crooked description — a field missing or of the
   wrong type — → 422 naming the field; a text or picture file
   named by the description and missing → 422 naming it; an
   address already taken (same date and title) → 409; a missing
   description → 422.
6. `GET /records` lists the records newest first (on equal
   moments, by address), with moment, title and tags.
7. `GET /records/{address}` shows the record whole: the moment,
   the title, the texts, the pictures' names, the tags.
8. `DELETE /records/{address}` removes the record with its media
   → 204; unknown address → 404.
9. `GET /records/{address}/media/{name}` serves the picture's
   bytes.
10. `GET /tags` answers the tags with their records' addresses.
11. `POST /publishes` queues a publish run → 202 with its id.
12. `GET /publishes` lists the runs with their states.
13. `GET /publishes/{id}` shows the run: state, and when done the
    "ok" listing — one "ok" line per record in build order (the
    index's order, only oldest to newest), the last line "ok
    index", exactly as pressmark prints it.
14. `GET /site/index` serves the built index.html of the current
    publication; `GET /site/{address}` serves a record's page;
    `GET /site/{address}/media/{name}` serves its published
    picture. No publication yet → 404 on all three.
15. `GET /health` answers 200 with the API's state and the
    worker's last heartbeat; a worker silent for over a minute is
    said so in plain words. A wrong method, an unknown route or a
    broken payload is an honest 4xx with a one-line reason.

## The record and the build — the pressmark law, unchanged

16. A record's description is the `meta.yaml`: the creation moment
    as "year-month-day hour:minute:second", no time zone, the same
    shape in all records; a title; the full-text file name; the
    short-text file name; a list of pictures and a list of tags —
    both required, both allowed to be empty. Texts are markdown;
    pictures are just files.
17. A record's address: "year-month-day", a hyphen, the title
    brought to link form — lowercase letters and digits only,
    spaces and other characters become hyphens, consecutive
    hyphens collapse into one, leading and trailing hyphens are
    dropped, letters of other alphabets are transliterated into
    Latin.
18. The publish builds the site whole: each record gets a page,
    its pictures move to `media/<address>/<name>`, and one index
    is built. Into the record page go the moment, the title, the
    address, the ready HTML of the full and short texts (just
    markdown, no templating inside the text), the pictures (name
    and address) and the tags. Into the index go the same records
    without the full texts, from newest to oldest; on equal
    moments, by address.
19. The templates are the delivery's own two — `entity.html` for
    the record page and `index.html` for the index — carried by
    the delivery and applied to every publish.
20. An empty record set publishes honestly: an index with an empty
    list, the "ok index" line, nothing else.
21. A publish of unchanged records gives byte-identical pages to
    the previous run, to the letter.

## The store of record

22. PostgreSQL is the system of record, with the exact tables
    `records` (address — primary key, moment, title, full_text,
    short_text, tags), `record_media` (address, name, bytes),
    `publishes` (id, state, detail, timestamps), `pages` (publish
    id, path, bytes) and one row `site_state` naming the current
    publication; the columns beyond these are yours, the names are
    not.
23. The delivery carries versioned migrations (an ordered
    `migrations/` artifact) applied by its `migrate` command: a
    fresh store gets the full schema, a repeat apply is a no-op.
24. The store is reached only through the ENV-named socket; no
    other database is touched, and no schema change happens
    outside the migrations.

## The hot path

25. Redis carries the publish queue — the API pushes run ids, the
    worker takes them; it is the ONLY channel between the two
    processes.
26. `/site/*` is served through a Redis cache of the current
    publication's pages; a fresh good publication swaps the cached
    set whole — no stale byte is served after the swap.
27. Redis is never the record: flushed and restarted, every fact
    (records, media, publishes, pages) stands in the store and the
    service keeps working.

## The invariant

28. One transactional law protects the site: the site serves
    exactly one complete publication — a publish turns ok together
    with ALL its page rows and the `site_state` pointer flip, one
    atomic step; a reader never sees a half-published site, a
    failed publish leaves the previous publication standing whole,
    and a record's address is unique forever (the old honest
    refusal is now the store's own constraint).

## Architecture

29. The delivery's layout names three boundaries — domain (the
    record law: descriptions, the address rule, rendering),
    transport (the HTTP face), storage (the store and cache
    gateways) — as separate modules or directories; the domain
    code knows neither HTTP, nor SQL, nor Redis.

## How I use it

30. First the platform: `sh "$PG_SERVICE" start` and
    `sh "$REDIS_SERVICE" start` (both idempotent; their state
    lives in the cell's scratch). Then `migrate`, then `serve
    --listen 127.0.0.1:"$SERVE_PORT"` (the port this cell's ENV
    names), then `worker` — in that order, and the product
    answers.

## Test material

31. The material directory holds the old fixtures: `storage/` —
    three records (one with pictures and tags, one with empty
    lists, one simpler: the titles "first", "second", "third",
    different moments, texts a couple of lines long) and
    `templates/` — the pair substituting all the fields; plus
    `storage-invalid/` — a record folder whose description misses
    a field (the honest 422 of the upload).
32. The bench drives: the three records upload with their computed
    addresses; a publish builds with the expected "ok" listing and
    byte-honest pages; the invalid folder is refused 422 naming
    the field; a duplicate record (same date and title) is refused
    409; a deleted record disappears from the next publish whole;
    a republish is byte-identical; after a new publication no
    stale byte is served; a publish queued while the worker is
    down runs when the worker returns; after a full restart of
    everything the facts stand.

## Limits and errors

33. All files are read and written in utf-8. Every refusal is
    brief, in English, with the path and the reason; the codes are
    honest: 404 an unknown record, page or file; 409 a taken
    address; 422 a crooked description or payload with the field
    named; 503 only when the store or the cache is truly
    unreachable, with the reason.

## What must not happen

34. No network beyond the cell's loopback and the ENV-named
    sockets; no reading or writing outside the cell; no state
    outside the store and the cache; the API never renders; a page
    of a failed publish is never served; my uploaded record
    folders are never changed; a migration change after the first
    apply.
