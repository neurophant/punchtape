# Etalon: pressmark — a publishing service from record folders

The meaning's yardstick, written into the campaign before any run:
every hand of pressmark, in every stack, is judged against these
same lines — the instrument is one; no product, no hand and no
judge forms it. Each line is one requirement of the brief with the
run that proves it.

1. The product is two cooperating processes — `serve` the HTTP
   face, `worker` the background publisher — with the README
   naming exactly one build command, one `migrate` command and the
   two run commands — proof: read the README, find the four names,
   nothing else needed.
2. `serve --listen host:port` answers HTTP there; `worker` takes no
   arguments; a stop signal is a calm shutdown whose last line
   reports the stop, unfinished work honestly queued — proof: start
   both, curl the port, SIGTERM each and read the last line.
3. The API never renders: every publish is queued through Redis and
   done by the worker — proof: with the worker stopped a queued
   publish stays pending, with it up it completes; the code's
   process split read.
4. The store and the cache run inside the cell by the ENV-named
   service files, reached only through the ENV-named socket
   addresses — proof: run the two service scripts, connect a store
   client through `PGHOST`, a cache client through `REDIS_SOCK`.
5. `POST /records` uploads one record folder → 201 with the
   address; a crooked description (field missing or wrong type) →
   422 naming the field; a named file missing → 422 naming it; a
   taken address (same date and title) → 409; no description →
   422 — proof: upload the three material records, the invalid
   folder, and one record twice; read the codes and bodies.
6. `GET /records` lists newest first (equal moments by address)
   with moment, title, tags — proof: read the order against the
   moments.
7. `GET /records/{address}` shows the record whole: moment, title,
   texts, pictures' names, tags — proof: read one against its
   uploaded bundle.
8. `DELETE /records/{address}` removes the record with its media →
   204; unknown → 404 — proof: delete one, re-read the list and
   the store's rows.
9. `GET /records/{address}/media/{name}` serves the picture's bytes
   — proof: `cmp` against the uploaded original.
10. `GET /tags` answers the tags with their records' addresses —
    proof: read against the records' tags.
11. `POST /publishes` queues a run → 202 with its id — proof:
    queue one, read the id in the runs list.
12. `GET /publishes` lists the runs with states — proof: read after
    several publishes.
13. `GET /publishes/{id}` shows the run: state, and when done the
    "ok" listing — one line per record in build order (the index's
    order, oldest to newest), the last line "ok index" — proof:
    read a done run; the order matches the moments, the last line
    is "ok index".
14. `GET /site/index`, `/site/{address}`, `/site/{address}/media/
    {name}` serve the current publication; no publication yet → 404
    on all three — proof: read all three before any publish and
    after one; `cmp` the media bytes.
15. `GET /health` answers 200 with the API state and the worker's
    last heartbeat, a silent worker named in plain words; a wrong
    method, unknown route or broken payload is an honest 4xx with
    a one-line reason — proof: health read twice around a worker
    stop; one wrong method provoked.
16. The description law stands: the `meta.yaml` moment as
    "year-month-day hour:minute:second", no time zone, same shape
    in all records; a title; the full and short file names; a
    pictures list and a tags list, both required, both allowed
    empty; texts markdown; pictures just files — proof: the three
    records upload and stand; an empty list record among them.
17. The address rule stands: date, hyphen, the title in link form
    (lowercase letters and digits only, others become hyphens,
    consecutive hyphens collapse, edges dropped, other alphabets
    transliterated) — proof: the uploaded records' addresses read
    against their moments and titles, "Article 1" style cases
    included.
18. The publish builds the site whole: a page per record, pictures
    to `media/<address>/<name>`, one index; into the record page go
    moment, title, address, the ready HTML of full and short texts
    (no templating inside the text), pictures (name and address),
    tags; into the index the records without full texts, newest
    first, equal moments by address — proof: the served pages show
    every substituted field; the index's order read.
19. The delivery carries its own two templates — `entity.html` and
    `index.html` — applied to every publish — proof: the delivery's
    template files read; the pages carry their marks.
20. An empty record set publishes honestly: an index with an empty
    list, the "ok index" line, nothing else — proof: publish with
    all records deleted; the listing and the index read.
21. A publish of unchanged records is byte-identical to the
    previous run — proof: publish twice, compare the served pages.
22. The store owns the exact tables `records` (address primary
    key), `record_media`, `publishes`, `pages` and the one-row
    `site_state` — proof: a store client lists the tables and
    their rows match the API's answers.
23. Versioned migrations applied by `migrate`: fresh store → full
    schema; repeat apply → no-op — proof: run it twice, read the
    second run's silence.
24. The store is reached only through the ENV socket; no schema
    change outside migrations — proof: the code read.
25. Redis carries the publish queue, the only channel between the
    processes — proof: the code's transport read; a publish queued
    with the API alone reaches the worker.
26. `/site/*` is served through the Redis cache; a fresh good
    publication swaps the set whole — no stale byte after — proof:
    serve a page, change a record, publish, serve again and
    compare.
27. Redis flushed and restarted loses no fact — proof: stop the
    cache service, start it, read the records, publishes and site.
28. The invariant: the site serves exactly one complete publication
    — a publish turns ok together with all its page rows and the
    `site_state` flip in one atomic step; a failed publish leaves
    the previous publication whole; an address is unique forever
    (the store's own constraint) — proof: the store rows read as
    one set; a duplicate upload refused 409 by the constraint's own
    honesty.
29. The layout names three boundaries — domain (descriptions, the
    address rule, rendering), transport, storage — as separate
    modules; the domain knows neither HTTP, nor SQL, nor Redis —
    proof: the tree read, the domain's imports read.
30. The platform comes up by the two service files, then `migrate`,
    `serve`, `worker` — in that order the product answers — proof:
    a cold start from scratch, cell scratch empty.
31. The material holds the three records, the template pair and
    `storage-invalid/` (a description missing a field) — proof: the
    material directory read.
32. The bench drives the whole law: three records upload with their
    computed addresses; a publish builds with the expected listing
    and byte-honest pages; the invalid folder refused 422 naming
    the field; a duplicate refused 409; a deleted record vanishes
    from the next publish whole; a republish byte-identical; no
    stale byte after a new publication; a publish queued while the
    worker is down runs on its return; a full restart keeps every
    fact — proof: the bench run's output read.
33. All files read and written in utf-8; every refusal brief, in
    English, with the path and the reason; the codes honest
    (404/409/422/503 by their meanings) — proof: each refusal
    provoked once, its body read.
34. No network beyond the cell's loopback and the ENV sockets; no
    reading or writing outside the cell; no state outside the store
    and the cache; the API never renders; a failed publish's page
    never served; the uploaded record folders never changed; no
    migration change after the first apply — proof: the code read
    plus the runs' behavior.
