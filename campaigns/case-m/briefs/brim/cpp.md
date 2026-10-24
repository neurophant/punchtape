# Brief: brim — a margin-measurement service
I scan documents by the hundreds now, and cropping by hand does
not scale. So the measurer becomes a service: I upload the scan
once, the service measures the margins the way brim always did —
evenness against content, entropy as the measure of chaos — in the
background, keeps every measurement in its ledger, and answers
repeats from a hot cache without re-measuring. The old law stands:
the margin is an even strip from the edge, four numbers, the same
answer to the letter.

## Stack and tools — my requirement, not the executor's choice

C++17, built with g++. The sense's own: GraphicsMagick++ (reading
and decoding pictures) — the platform prefix of ENV with its
-config wrapper. The service: Crow 1.3.4 (the `crow_all.h` of the
prefix) over standalone Asio 1.38.2 — the HTTP face; the store
through the system libpq 18.6 (runtime libpq.so.5, linked by that
name; the client headers ride the prefix); the cache and queue
through the vendored hiredis 1.5.0 sources of env/deps/hiredis,
compiled in the cell — no other redis library.

## The service and its two processes

1. The product is two cooperating processes: `serve` — the HTTP
   face — and `worker` — the background measurer. The delivery's
   README names exactly one build command (if the stack needs
   one), one `migrate` command, and the two run commands; nothing
   else is needed to run the product.
2. `serve` takes `--listen host:port` and answers HTTP on it;
   `worker` takes no arguments. A stop signal (interrupt or
   terminate) is a calm shutdown for both: the last line reports
   the stop, unfinished work stays honestly queued.
3. The API process never measures a picture: every measurement is
   queued through Redis and done by the worker; the API answers
   from the store of record and the hot cache only.
4. The store (PostgreSQL 18.6) and the cache (Redis 8.0.5) run
   inside the cell: the service files named by ENV (`PG_SERVICE`,
   `REDIS_SERVICE`) bring them up before the processes start; the
   store rides the unix socket directory `PGHOST` (with
   `PGPORT`, `PGUSER`, `PGDATABASE`), the cache the socket
   `REDIS_SOCK`. No other addresses exist for the product.

## The API — thirteen endpoints, honest codes

5. `POST /pictures` uploads one still JPEG (the one format, as
   ever) → 201 with the picture's record: id, sha256 digest,
   width, height. Not a JPEG or unreadable bytes → 415 with the
   reason; an empty body → 422.
6. `GET /pictures` lists the pictures, oldest first.
7. `GET /pictures/{id}` shows the picture's record.
8. `DELETE /pictures/{id}` removes the picture with its
   measurements → 204; unknown id → 404.
9. `POST /pictures/{id}/measurements` with
   `{"band": 0.25, "threshold": 0.5}` queues a measurement → 202
   with its id. The knobs carry the old law: band — a fraction of
   the side, no more than half allowed; threshold — a fraction
   from 0 to 1. A crooked value → 422 naming the knob and its
   requirement.
10. `GET /measurements/{id}` shows the measurement: state
    (waiting, running, done, failed) and, when done, the four
    numbers — top, bottom, left, right — as integers.
11. `GET /pictures/{id}/measurements` lists the picture's
    measurements with their knobs and states.
12. `POST /presets` with `{"name": "...", "band": ..., 
    "threshold": ...}` saves a named pair → 201; a taken name →
    409; crooked knob values → 422.
13. `GET /presets` lists the presets.
14. `DELETE /presets/{name}` removes one → 204; unknown → 404.
15. `POST /pictures/{id}/measurements-by-preset/{name}` queues a
    measurement with the preset's knobs → 202.
16. `GET /cache` answers the cache's honest counters: hits, misses,
    recomputes.
17. `GET /health` answers 200 with the API's state and the
    worker's last heartbeat; a worker silent for over a minute is
    said so in plain words. A wrong method, an unknown route or a
    broken JSON body is an honest 4xx with a one-line reason.

## The measurement — the brim law, unchanged

18. The margin is an even strip running from the edge: white,
    black, anything — the main thing is that it is even. Each side
    is measured on its own; evenness can stretch to the end of the
    search band — then the whole band is the answer; no margin
    means zero.
19. Evenness is measured with entropy: an even strip's entropy is
    near zero, content is alive. The boundary is where the
    difference in chaos between neighboring chunks of the band is
    at its largest — the walk goes from the edge inward.
20. The knobs mean what they always meant: band — how deep from
    the edge to search (a fraction of the side); threshold — the
    sensitivity (the higher, the more aggressively the boundary is
    hunted).
21. A repeat run on the same picture with the same knobs gives the
    same four numbers, to the letter.
22. A picture whose side is shorter than two pixels fails honestly
    — there is nothing to search in — as a failed measurement with
    that reason.

## The store of record

23. PostgreSQL is the system of record, with the exact tables
    `pictures` (id, sha256, bytes, width, height), `measurements`
    (id, picture, band, threshold, state, the four numbers, the
    digest measured, error, timestamps), `presets` (name, band,
    threshold — name unique); the columns beyond these are yours,
    the three names are not.
24. The delivery carries versioned migrations (an ordered
    `migrations/` artifact) applied by its `migrate` command: a
    fresh store gets the full schema, a repeat apply is a no-op.
25. The store is reached only through the ENV-named socket; no
    other database is touched, and no schema change happens
    outside the migrations.

## The hot path

26. Redis carries the measurement queue — the API pushes ids, the
    worker takes them; it is the ONLY channel between the two
    processes.
27. Before measuring, the worker looks in the Redis result cache
    keyed by the picture's DIGEST with the knobs: a hit answers
    without re-measuring (and counts in `cache`); a miss measures,
    records in the store, and fills the cache. The same bytes
    under a new id hit the same cache.
28. Redis is never the record: flushed and restarted, every fact
    (pictures, measurements, presets) stands in the store and the
    service keeps working.

## The invariant

29. One transactional law protects the ledger: a measurement turns
    done together with its four numbers and the digest of the
    exact bytes it was measured from — one atomic step; a recorded
    result always reproduces on re-measurement of the same digest
    and knobs, and a picture's bytes are immutable (no update path
    exists).

## Architecture

30. The delivery's layout names three boundaries — domain (the
    measurement law: entropy, the walk, the knobs), transport (the
    HTTP face), storage (the store and cache gateways) — as
    separate modules or directories; the domain code knows
    neither HTTP, nor SQL, nor Redis.

## How I use it

31. First the platform: `sh "$PG_SERVICE" start` and
    `sh "$REDIS_SERVICE" start` (both idempotent; their state
    lives in the cell's scratch). Then `migrate`, then `serve
    --listen 127.0.0.1:"$SERVE_PORT"` (the port this cell's ENV
    names), then `worker` — in that order, and the product
    answers.

## Test material

32. The material directory holds the two old pictures —
    `bordered.jpg` (250×250, a thin black even frame: four
    identical numbers, four pixels on each side) and `clear.jpg`
    (510×350, no margin: zeros on all sides) — plus `corrupt.jpg`
    (bytes that are not a picture: an honest 415). The old knob
    law is proven on bordered.jpg: band 0.01 → "2 2 2 2", band
    0.05 → "4 4 4 4".
33. The bench drives: bordered and clear measure to their known
    answers; the band matrix holds (0.01, 0.05); a repeat
    measurement is served from the cache (the counters move, the
    answer is identical); the same bytes uploaded twice hit one
    cache key; corrupt.jpg is refused 415; band 0.6 is refused 422
    with the knob named; a measurement queued while the worker is
    down runs when the worker returns; two pictures measure
    concurrently without interference; after a full restart of
    everything the facts stand.

## Limits and errors

34. Every refusal is brief, in English, with the path and the
    reason; the codes are honest: 404 an unknown picture,
    measurement or preset; 409 a taken preset name; 415 a picture
    that is not my one format; 422 a crooked knob or payload with
    the field named; 503 only when the store or the cache is truly
    unreachable, with the reason.

## What must not happen

35. No network beyond the cell's loopback and the ENV-named
    sockets; no reading or writing outside the cell; no state
    outside the store and the cache; the API never measures; a
    cached answer is never stale for its key; a migration change
    after the first apply.
