# Etalon: brim — a margin-measurement service

The meaning's yardstick, written into the campaign before any run:
every hand of brim, in every stack, is judged against these same
lines — the instrument is one; no product, no hand and no judge
forms it. Each line is one requirement of the brief with the run
that proves it.

1. The product is two cooperating processes — `serve` the HTTP
   face, `worker` the background measurer — with the README naming
   exactly one build command, one `migrate` command and the two run
   commands — proof: read the README, find the four names, nothing
   else needed.
2. `serve --listen host:port` answers HTTP there; `worker` takes no
   arguments; a stop signal is a calm shutdown whose last line
   reports the stop, unfinished work honestly queued — proof: start
   both, curl the port, SIGTERM each and read the last line.
3. The API never measures: every measurement is queued through
   Redis and done by the worker — proof: with the worker stopped a
   queued job stays waiting, with it up the job completes; the
   code's process split read.
4. The store and the cache run inside the cell by the ENV-named
   service files, reached only through the ENV-named socket
   addresses — proof: run the two service scripts, connect a store
   client through `PGHOST`, a cache client through `REDIS_SOCK`.
5. `POST /pictures` uploads one still JPEG → 201 with id, sha256,
   width, height; not a JPEG → 415 with the reason; an empty body
   → 422 — proof: upload the material's pictures, then
   `corrupt.jpg`, then an empty body; read the codes.
6. `GET /pictures` lists oldest first — proof: upload two, read the
   order.
7. `GET /pictures/{id}` shows the record — proof: read it against
   the upload's answer.
8. `DELETE /pictures/{id}` removes the picture with its
   measurements → 204; unknown id → 404 — proof: delete one, re-read
   the list and the store's rows.
9. `POST /pictures/{id}/measurements` with band and threshold
   queues a measurement → 202; band over half or not a fraction,
   threshold outside 0..1 → 422 naming the knob and its
   requirement — proof: queue a good one, then band 0.6, then
   threshold 1.5; read the bodies.
10. `GET /measurements/{id}` shows the state and, when done, the
    four integers top, bottom, left, right — proof: read a waiting
    one, then the done one.
11. `GET /pictures/{id}/measurements` lists the picture's jobs with
    knobs and states — proof: read after several measurements.
12. `POST /presets` saves a named pair → 201; a taken name → 409;
    crooked knobs → 422 — proof: three posts, read the codes.
13. `GET /presets` lists them — proof: read after the saves.
14. `DELETE /presets/{name}` removes one → 204; unknown → 404 —
    proof: delete and re-ask.
15. `POST /pictures/{id}/measurements-by-preset/{name}` queues with
    the preset's knobs → 202 — proof: queue via a preset, read the
    job's knobs in the listing.
16. `GET /cache` answers the honest counters hits, misses,
    recomputes — proof: read it before and after a repeat
    measurement; the numbers move as the law says.
17. `GET /health` answers 200 with the API state and the worker's
    last heartbeat, a silent worker named in plain words; a wrong
    method, unknown route or broken body is an honest 4xx with a
    one-line reason — proof: health with the worker up and stopped;
    one wrong method provoked.
18. The margin law stands: an even strip from the edge, each side
    measured on its own, evenness to the band's end is the whole
    band, no margin is zero — proof: `clear.jpg` answers zeros on
    all sides.
19. Evenness is entropy: the boundary is where the chaos difference
    between neighboring chunks is largest, walked from the edge
    inward — proof: `bordered.jpg` finds the frame, not a guess;
    the code's walk read.
20. The knobs mean the old: band — the search depth as a fraction
    of the side; threshold — the sensitivity — proof: the band
    matrix below behaves by it.
21. A repeat run on the same picture with the same knobs gives the
    same four numbers to the letter — proof: two identical
    measurements, compare.
22. A picture with a side shorter than two pixels fails honestly —
    nothing to search in — as a failed measurement with the reason
    — proof: the code's guard read (a 1-pixel picture cannot be a
    JPEG worth uploading; the guard is the proof).
23. The store owns the exact tables `pictures`, `measurements`,
    `presets` (name unique) — proof: a store client lists the
    tables and their rows match the API's answers.
24. Versioned migrations applied by `migrate`: fresh store → full
    schema; repeat apply → no-op — proof: run it twice, read the
    second run's silence.
25. The store is reached only through the ENV socket; no schema
    change outside migrations — proof: the code read.
26. Redis carries the measurement queue, the only channel between
    the processes — proof: the code's transport read; a job queued
    with the API alone reaches the worker.
27. The result cache is keyed by DIGEST with the knobs: a hit
    answers without re-measuring and counts; a miss measures,
    records, fills; the same bytes under a new id hit the same key
    — proof: measure, re-upload the same bytes as a new picture,
    measure again — the counters say hit, the answers identical.
28. Redis flushed and restarted loses no fact — proof: stop the
    cache service, start it, read the pictures, measurements and
    presets.
29. The invariant: a measurement turns done together with its four
    numbers and the digest of the exact bytes measured — one atomic
    step; a recorded result reproduces on re-measurement of the
    same digest and knobs; a picture's bytes are immutable (no
    update path) — proof: the store rows read; a re-measure after
    a cache flush returns the same numbers; no update route exists
    in the API.
30. The layout names three boundaries — domain (entropy, the walk,
    the knobs), transport, storage — as separate modules; the
    domain knows neither HTTP, nor SQL, nor Redis — proof: the tree
    read, the domain's imports read.
31. The platform comes up by the two service files, then `migrate`,
    `serve`, `worker` — in that order the product answers — proof:
    a cold start from scratch, cell scratch empty.
32. The material holds `bordered.jpg` (four pixels a side),
    `clear.jpg` (zeros) and `corrupt.jpg`; the band law on
    bordered: band 0.01 → "2 2 2 2", band 0.05 → "4 4 4 4" —
    proof: the material read; the three measurements run and their
    answers compared.
33. The bench drives the whole law: known answers, the band matrix,
    the repeat served from cache, one key for the same bytes,
    corrupt refused 415, band 0.6 refused 422 with the knob named,
    a job queued while the worker is down runs on its return, two
    pictures measure concurrently, a full restart keeps every fact
    — proof: the bench run's output read.
34. Every refusal is brief, in English, with the path and the
    reason; the codes honest (404/409/415/422/503 by their
    meanings) — proof: each refusal provoked once, its body read.
35. No network beyond the cell's loopback and the ENV sockets; no
    reading or writing outside the cell; no state outside the store
    and the cache; the API never measures; a cached answer never
    stale for its key; no migration change after the first apply —
    proof: the code read plus the runs' behavior.
