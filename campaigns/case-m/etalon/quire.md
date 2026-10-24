# Etalon: quire — a page service from uploaded sources

The meaning's yardstick, written into the campaign before any run:
every hand of quire, in every stack, is judged against these same
lines — the instrument is one; no product, no hand and no judge
forms it. Each line is one requirement of the brief with the run
that proves it.

1. The product is two cooperating processes — `serve` the HTTP
   face, `worker` the background builder — with the README naming
   exactly one build command, one `migrate` command and the two run
   commands — proof: read the README, find the four names, nothing
   else needed.
2. `serve --listen host:port` answers HTTP there; `worker` takes no
   arguments; a stop signal is a calm shutdown whose last line
   reports the stop, unfinished work honestly queued — proof: start
   both, curl the port, SIGTERM each and read the last line.
3. The API never renders: every build is queued through Redis and
   done by the worker — proof: with the worker stopped a queued
   build stays pending, with it up the build completes; the code's
   process split read.
4. The store and the cache run inside the cell by the ENV-named
   service files, reached only through the ENV-named socket
   addresses — proof: run the two service scripts, connect a store
   client through `PGHOST`, a cache client through `REDIS_SOCK`.
5. `POST /sites` creates a site → 201; a taken name → 409; a
   crooked name → 422 with the field named — proof: three posts,
   read the codes and the bodies.
6. `GET /sites` lists the sites oldest first — proof: create two,
   read the order.
7. `GET /sites/{id}` shows the record and the newest build's state
   — proof: read it before and after a build.
8. `DELETE /sites/{id}` removes the site with source, builds and
   artifacts → 204; unknown id → 404 — proof: delete one, re-read
   the list and the store's tables.
9. `PUT /sites/{id}/source` uploads one bundle → 201 with a new
   revision; a second upload replaces whole, never merges — proof:
   upload, list files; upload a different bundle, list again.
10. `GET /sites/{id}/source` lists the bundle's files with sizes;
    the one-file route returns its bytes — proof: read both against
    the uploaded material.
11. `POST /sites/{id}/builds` queues a build → 202 with the
    per-site number starting at 1 — proof: two builds on one site,
    a first on another.
12. `GET /sites/{id}/builds` lists newest first with states —
    proof: read after several builds.
13. `GET /sites/{id}/builds/{n}` shows the build: state, and when
    done the "ok" listing — one line per created file, stable
    order, the root's files alphabetically — proof: read the done
    build and compare the listing with the output's files.
14. `GET /sites/{id}/page` serves the current good build's
    index.html; none yet → 404 — proof: read before any build and
    after one.
15. `GET /sites/{id}/files/{path}` serves a copied output file byte
    for byte; anything outside → 404 — proof: `cmp` a served file
    against the bundle's original; ask for a path outside.
16. `GET /health` answers 200 with the API state and the worker's
    last heartbeat, a silent worker named in plain words — proof:
    read it with the worker up, then stopped, wait, read again.
17. A wrong method, unknown route or broken JSON body is an honest
    4xx with a one-line reason — proof: provoke each once.
18. The build keeps the quire law: markdown in the root are the
    entries; front matter between `---` is the entry's data, the
    rest the body → HTML; `data.yml` is `data`; no file — no data
    — proof: the material's bundle builds and the page shows data
    substitution, rendered bodies, and the plain `about.md` entry.
19. The one `index` template (the stack's extension) builds into
    `index.html` with `data` and the entries available — proof: the
    built page shows both.
20. The entries reach the template alphabetically by file name —
    proof: the page lists them in that order.
21. Everything not a text, not `data.yml`, not the template is
    copied byte for byte, nested directories plain; the texts,
    `data.yml` and the template do not land in the output — proof:
    `cmp` the extraneous file, check the nested copy, check the
    absents.
22. A bundle without texts and without data builds ok with an
    empty entry list — proof: build the empty bundle.
23. A bundle without a template fails honestly ("nothing to
    build"); crooked YAML, a broken template — the same honest
    failed build with the reason, the last good page stands —
    proof: build the invalid bundle after a good one; read the
    build state, the reason, and the page.
24. Two builds of one unchanged revision give byte-identical
    output — proof: build twice, compare artifacts through the
    file routes.
25. Every created file has its "ok" line in the listing — proof:
    the done build's listing counts the output's files.
26. The store owns the exact tables `sites`, `sources`, `builds`,
    `artifacts` — proof: a store client lists the tables and their
    rows match the API's answers.
27. Versioned migrations applied by `migrate`: fresh store → full
    schema; repeat apply → no-op — proof: run `migrate` on the
    fresh store, run it again, read the second run's silence.
28. The store is reached only through the ENV socket; no schema
    change outside migrations — proof: the code read.
29. Redis carries the build queue, the only channel between the
    processes — proof: the code's transport read; a build queued
    with the API alone reaches the worker.
30. `page` and `files` are served through the Redis cache; a fresh
    good build swaps the set whole — no stale byte after — proof:
    serve a page, publish a new revision, serve again and compare.
31. Redis flushed and restarted loses no fact — proof: stop the
    cache service, start it, read the sites, builds and page.
32. The invariant: a build turns ok together with all its artifact
    rows and the site's current-build pointer, one atomic step; a
    failure leaves the pointer at the previous good build — proof:
    the good-build store rows read as one set; the failed build's
    page route still serves the previous page.
33. The layout names three boundaries — domain, transport, storage
    — as separate modules; the domain knows neither HTTP, nor SQL,
    nor Redis — proof: the tree read, the domain's imports read.
34. The platform comes up by the two service files, then `migrate`,
    `serve`, `worker` — in that order the product answers — proof:
    a cold start from scratch, cell scratch empty.
35. The material holds the stack bundle, `bundle-invalid/` and
    `bundle-empty/` — proof: the material directory read.
36. The bench drives the whole law: the stack bundle builds ok; the
    invalid bundle fails with the reason and the page stands; a
    rebuild is byte-identical; a new revision retires the old
    artifacts whole; a build queued while the worker is down runs
    on its return; two sites build concurrently; a full restart
    keeps every fact — proof: the bench run's output read.
37. Every refusal is brief, in English, with the path and the
    reason; the codes honest (404/409/422/503 by their meanings) —
    proof: each refusal provoked once, its body read.
38. No network beyond the cell's loopback and the ENV sockets; no
    reading or writing outside the cell; no state outside the store
    and the cache; the API never renders; a failed build's artifact
    never served; no migration change after the first apply —
    proof: the code read plus the runs' behavior.
