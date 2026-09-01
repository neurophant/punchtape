# SAMPLES MATRIX — tasks × stacks × sizes

The single document of task calibration (self-supporting). The owner of the task
originals is always right; the calibration is recomputed from the metrics
of each campaign.

## 1. Size calibration — bands derived from campaign-01, refined by every run

Rule: calibration is objective and fact-only — it is derived AFTER the
full run of all campaign-01 tasks is complete (five tasks, all the
campaign's stacks). No size bounds are assigned before that point:
the S/M/L below are NAMES of feature-set axes, not measured classes.
When the full run is over, the bounds are derived from the
measurements over five factors together: time, tokens, the number of
scenarios, the volume of code, and the objective complexity of the
logic (adjusted for brief detail); the criterion — tasks of one size,
at comparable brief detail, give a comparable result in comparable
time and cost.

## 2. Campaign-01 measurements (the truth of the calibration)

| Task | Stack | Wall | Tokens | Scenarios | Code (lines) | Bill | Label from measurements |
|---|---|---|---|---|---|---|---|
| conduit | go | 11.3 min | 0.43 M | 9 | 200 | 0.24 | **S** |
| pressmark | python | 10.3 min | n/a | 20 | 184 | 0.34 | **S** (boundary) |
| brim | c/c++ | 27.5 min | 4.64 M | 26 | 275 | 0.53 | **M** |
| errand | rust | 31.4 min | 2.09 M | 30 | 668 | 0.64 | **M** (top) |
| quire | js/ts | 48.9 min | 9.28 M + hand n/a | 22 | 228 | 0.81 | **L** |

The campaign-01 full run is complete (5/5 READY + acceptances). The
bands are derived from the campaign-01 measurements: S ≤ 0.35 ·
M 0.36–0.70 · L > 0.70 — preliminary (5 points), refined by every run
(the databases of all runs are in docs/METRICS.md). By these bands,
THE TASK AXES ARE RE-CUT (§4): the measured set became the axis of its
own band — conduit and pressmark in S, brim and errand in M, quire in
L. The stack multiplier (c/c++ is more expensive) is taken into
account in the comparison: all five points are native stacks.

## 3. Target stacks

python · go · rust · c/c++ · js/ts — the stacks of the current field
campaigns.

## 4. Tasks by sense (5) and their feature-set axes

The upscale features are taken from the source originals (every sense
has a rich original); the downscale is the core. The axes are re-cut
from the campaign-01 measurements: the measured set became the axis of
its own band (the anchor "= meas., c01: bill"); the unmeasured axes
are derived from the sense of the originals and refined by runs.

### brim — an image-field meter (original: a border detector)

- **S (core)**: one command, JPEG, entropy, 4 sides, band,
  threshold; material of 2–3 pictures.
- **M (+)** [= meas., c01: 0.53]: + GIF, all frames with
  aggregation, + strict mode, + GIF material, + the full set of
  errors.
- **L (++)**: + resize with coordinate recomputation, + a frame
  limit, + random column sampling (convergence), + a batch mode over
  a directory, + a CSV report.
- Tools: c/c++ — GraphicsMagick++; python — Pillow+numpy; go — stdlib
  image; rust — the image crate; js/ts — jimp/sharp (offline, local).

### conduit — a two-stage relay (original: an encrypted proxy)

- **S** [= meas., c01: 0.24]: ingress+egress TCP, byte copying,
  break → closing the chain, launch failures, multi-client isolation,
  signals, a busy port, an echo testbed.
- **M (+)**: + configurable keepalive/idle timeouts, + connection
  IDs in events.
- **L (++)**: + a connection pool with tunnel multiplexing, + ws
  transport, + encryption (chacha20-poly1305) and a shared key, +
  authentication (HMAC challenge-response, basic at the entrance), +
  a bandwidth limit (token bucket), + reconnect backoff.
- Tools: go — stdlib (+x/crypto for L); python — asyncio; rust —
  tokio; c/c++ — POSIX sockets/epoll; js/ts — node:net (+ws for L).

### errand — a background queue (original: a redis job queue)

- **S**: push→pending, work-once→done, result. 3 operations.
- **M (+)** [= meas., c01: 0.64]: + 5 statuses (+lost), + TTL by
  stages, + handler errors, + drop-queue, + FIFO.
- **L (++)**: + parallel workers with locks, + priority queues, +
  retrying failed jobs with a delay, + queue metrics (length, age), +
  a monitoring CLI tool.
- Tools: rust — redis 0.8/uuid/serde (original); python — redis-py;
  go — go-redis; c/c++ — hiredis; js/ts — ioredis.

### pressmark — a file-based CMS build (original: a file-based CMS)

- **S** [= meas., c01: 0.34, boundary — the tokens were lost by the
  platform, the bill is by 4 metrics, recounted when the data
  appears]: an entry (folder + description + full and short texts), a
  slug by date, a page per entry, a table of contents from newest to
  oldest, media, an empty store, erasing and rebuilding the output,
  all the errors.
- **M (+)**: + pagination of the table of contents, + tag pages and
  label filters.
- **L (++)**: + drafts/publishing, + incremental rebuild by mtime, +
  an RSS feed, + a server mode for serving.
- Tools: python — PyYAML/Markdown/Jinja2/pydantic/slugify
  (original); go — stdlib+goldmark; rust — serde_yaml+comrak+tera;
  c/c++ — libyaml+md4c+inja; js/ts — js-yaml+marked+nunjucks.

### quire — a static-site generator from a directory (original: a no-config SSG)

- **S**: md with a header → object+body, one template, copying.
- **M (+)** (a downscale from L by measurement): + the directory
  tree, + an item template, + _includes, + rootPath, + --silent, +
  copying the rest; the typographic flags and arbitrary template
  extensions are cut.
- **L** [= meas., c01: 0.81]: everything in M + arbitrary template
  extensions (a style template → a style file), + typography
  (--breaks, --smartypants). The original's beyond-L material
  (template inheritance, taxonomies, incremental builds, a preview
  server, themes) is outside the matrix until a band above L appears.
- Tools: js/ts — marked+yeahjs+yeahml (original); python —
  markdown+Jinja2+PyYAML; go — goldmark+html/template+gopkg.in/yaml;
  rust — comrak+tera+serde_yaml; c/c++ — md4c+inja+libyaml.

## 5. The 5×5×3 = 75-cell matrix

Cell status: ⚪ no brief · 📝 brief written · ▨ run by an auto-run
(machine verdict READY+PASS, no operator acceptance) · 🟢 run by a
campaign with operator acceptance — THIS cell, with its own brief;
one run paints exactly one cell, no "counted for another axis". A
cell's features/tools come from its axis (§4). Campaign-01 painted
five cells with operator acceptances: conduit·go·S,
pressmark·python·S, brim·c/c++·M, errand·rust·M, quire·js/ts·L —
these five 🟢. The S layer is fully written as briefs (the 25 briefs
live at campaigns/campaign-02/briefs/<sense>/<stack>.md). Campaign-02
ran the whole S layer by auto-run on 2026-08-30:
25/25 machine verdicts READY+PASS (the databases — docs/METRICS.md
§2), no operator acceptances — the cells are ▨, not 🟢. The control
fix waves (cw1–cw8) ran brim·S in all stacks eight more times
(docs/METRICS.md §3). The remaining cells are ⚪.

| task\stack | python | go | rust | c/c++ | js/ts |
|---|---|---|---|---|---|
| **brim S** | ▨ | ▨ | ▨ | ▨ | ▨ |
| **brim M** | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ |
| **brim L** | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **conduit S** | ▨ | 🟢 | ▨ | ▨ | ▨ |
| **conduit M** | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **conduit L** | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **errand S** | ▨ | ▨ | ▨ | ▨ | ▨ |
| **errand M** | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| **errand L** | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **pressmark S** | 🟢 | ▨ | ▨ | ▨ | ▨ |
| **pressmark M** | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **pressmark L** | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **quire S** | ▨ | ▨ | ▨ | ▨ | ▨ |
| **quire M** | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **quire L** | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 |

Next: a total proofread of the machine for agnosticism + a check of
the tasks for cheating — done; the full run of the S layer — done
2026-08-30: 25/25 VERDICT READY + acceptance PASS, each task once
(docs/METRICS.md §4) — a repeat auto-run of the same cells with their
own briefs, the ▨ circles stay ▨ (the "one run paints one cell" rule,
no operator acceptances); operator acceptances paint ▨→🟢; the
further layers — as decided.
