# Field campaigns

A field campaign is how this project measures itself: real tasks, run
for real through the machine, with every number taken from machine
evidence. No simulated users, no cherry-picked demos — a battery of
small utility tasks executed by agents under controlled isolation,
scored by the machine's own gates and verdicts.

This directory holds everything needed to understand and reproduce the
campaigns:

- `campaign-01/` — the first campaign, run by hand (manual mode).
- `campaign-02/` — the second campaign, run as automated waves, plus
  the three-hand experiment (top model / flash model / solo, with
  judges).
- `../docs/SAMPLES.md` — the task battery itself: what each task sense
  is and why it was chosen.
- `../docs/METRICS.md` — the numbers: every run, wave by wave.

## The battery

Five task senses × five stacks = 25 tasks:

| Sense | What it is | Materials |
|---|---|---|
| brim | margin detector for scanned images (entropy-based) | two JPEG fixtures, deterministically generated |
| conduit | two-stage relay over TCP (encrypted ingress, plain egress) | none (an echo stand in the world) |
| errand | background job queue over Redis | none (redis in the world) |
| pressmark | file-based static site builder | a three-record storage + per-stack templates |
| quire | no-config static site generator from a directory | a content set + per-stack templates |

Stacks: python, go, rust, cpp, js. Each sense exists as five briefs —
identical wishes, different pinned stacks and libraries
(`campaign-02/briefs/<sense>/<stack>.md`).

## The world: one container per task

Every task runs in its own throwaway container:

```sh
docker run -d --name pt-<sense>-<stack> --network none \
  --cpus 1 --memory 1g --pids-limit 512 \
  -v <task-dir>:/run/task \
  -v <punchtape-binary>:/usr/local/bin/punchtape:ro \
  -w /run/task punchtape-world:<tag> sleep infinity
```

The rules that make a run honest:

- **Isolation.** Network off; one CPU; the repository, docs, and
  campaign files are never mounted into a task world (anti-cheating).
  Only the brief directory (read-write), the materials (read-only),
  and — for system hands — the machine binary (read-only).
- **One task — one agent — one container.** The container is removed
  the moment the task finishes; evidence (ledger, verdict, state) is
  exported before removal.
- **The binary is verified.** Built twice from the release tag,
  byte-for-byte compared, version stamped in; the SHA-256 goes into
  the wave journal.
- **Offline builds.** Each stack's libraries are pre-cached inside the
  world image (module caches, crate caches, node_modules, headers),
  with offline modes on — a task that needs the network fails loudly
  instead of quietly downloading.
- **OOM is an environment fact.** Exit code 137 means the limits, not
  the task: the world is recreated with a larger memory cap and the
  case goes into the run's cause registry.

## Manual mode (campaign-01)

The operator runs tasks one at a time, one agent per task, with the
per-task environment standing outside the machine: a container where
the stack needs pinning (brim: a g++/GraphicsMagick image;
pressmark: a pinned python image), the host toolchain where it does
not (conduit on go, errand on cargo, quire on node). Materials are
placed by the operator before the launch. This mode is slower and
operator-dependent — and it is how the battery was calibrated before
automation. Details: `campaign-01/README.md`.

## Automated waves (campaign-02)

A wave is one sense on all five stacks at once, five agents in
parallel (one task — one agent). Wave order: conduit → pressmark →
errand → brim → quire (the hardest build first, so a staging gap fails
early). The world image is layered: the base image with all toolchains
(`../world.Dockerfile`) plus the campaign layer
(`campaign-02/world.Dockerfile`) with the pinned libraries and caches.
Smoke checks (go build from cache, cargo offline, imports of every
pinned library) run before the first wave of the day. Details:
`campaign-02/README.md`.

## The three-hand experiment

The same 25 tasks, each done three ways, plus judges:

| Hand | Executor | Machine |
|---|---|---|
| system+top | a top model, as the task agent | punchtape in the world |
| system+flash | a cheap flash model, same contract | punchtape in the world |
| solo | a top model, professional freedom | none — the binary is not mounted |

All three hands get the same brief and the same world; isolation and
limits are identical; the executors' platform metrics (time, tokens,
tool calls) are captured the moment each hand finishes. Solo hands
write a SELF-REPORT and self-check freely (their own stands and
batteries — 12 to 283 checks per task).

The judge (a fourth agent, run after the three hands finish a task)
receives the byte-exact checks table of that task's top hand — the
table the machine itself verified — and the solo product, read-only.
Every row is executed by hand: seeds, materials, commands, and a
byte-for-byte comparison of exit code, stdout, stderr, and files. The
judge may not fix the product; a doubt counts against the product (an
honest judge is not a well-wisher). The output is a row-by-row
green/red table with actual-vs-expected — the measure that makes
"green in its own way" visible. The headline result: 50/50 machine
verdicts for the system hands against 22% strict byte convergence for
solo (`../docs/METRICS.md` §5).

## Metrics discipline

- Numbers come from two sources only: the instance ledger (machine
  side) and the executor platform metadata (time/tokens/tool calls).
  No self-reports.
- Executor metrics are captured IMMEDIATELY at hand completion — a
  lost cell stays lost.
- Everything an executor writes from a container is root-owned;
  `chown` it back before removing the container, and verify zero
  root-owned files after each wave.
- Every run updates `../docs/METRICS.md` in the same step.

## Running your own campaign

1. Build the machine twice from the same tree, compare the binaries
   byte-for-byte, stamp the version in.
2. Build the base world image (`world.Dockerfile`), then the campaign
   layer (`campaign-02/world.Dockerfile`), then the JS resolution layer
   (`campaign-02/world-js.Dockerfile` — it links `/run/node_modules`
   into `/tmp`, so bare imports resolve inside the throwaway run
   directories); run the smoke checks offline.
3. Pick your battery (or take this one), stage a clean directory per
   task: `brief.md` (read-write) and mount the task materials
   separately, read-only, at `/run/task/materials`.
4. Launch one agent per task in its container with the canonical
   executor contract (below).
5. Capture the metrics (`campaigns/tools/parse-ledger.py`), export the
   evidence, remove the containers, update the metrics file.

## The canonical executor contract (system hands)

Paste this to the executor agent verbatim; it is the calibrated
wording of the campaign methodology — do not improvise around it:

```text
Work in /run/task through the machine. The brief (brief.md) is the
only input; read it whole, and read AGENTS.md (the machine's own
reference) before the first call.

- Start with: punchtape next "$(cat brief.md)" — the whole brief is
  the wish; the first call creates the instance and pins it.
- Loop next/submit until the reply prints a VERDICT line; that line is
  the stop. Submit deltas as delta.yaml files, each ending with the
  trailer line: # TOKENS n/a
- ONE run: never restart, never delete the instance. A rejection is
  one line — fix the delta and submit it again.
- Exit codes 3/4 (a question batch / an escalation): decide yourself
  by the recommended option and record it as a default decision,
  awaiting ratification.
- The materials at /run/task/materials are read-only inputs: copy
  from them, never modify them, and never submit material files
  inside code deltas (the replay wedges on read-only mounts).
- Keep the spec lean: re-cut into a few requirements (2–3, the
  wish's user-facing features — not a tree of fragments); the table
  pins the brief's behaviors and nothing more — no scenarios the
  brief never asked for, no gold-plating; feature prose is one text
  per requirement.
- The verdict requires feature prose: after the checks are green,
  submit one kind: feature delta per requirement, in living words —
  the human spec is part of done. Plan for these submissions.
- Submit whole tables: the checks table is one delta (split only when
  a table is genuinely too large, never into fragments); repairs go as
  row replacements, not as new partial tables.
- Every run element is ONE command: a "cmd args" string or an argv
  list [cmd, args...] — a flat multi-element script is not a row;
  long-running helpers and waits live inside the row's own commands
  (a shell one-liner), and anything a command leaves running is
  killed when the row ends; background helpers must close their
  inherited stdout/stderr pipes (>/dev/null 2>&1).
- Nothing but submissions and the product: no side projects inside
  /run/task, no edits to files outside your product and delta.yaml.
```

For the three-hand mode, the solo hand gets NO machine-channel
knowledge: its contract is the original free-professional one — the
brief is the only input; there is no punchtape in its world; implement
the wish as a professional; the product lands in solo-product/ with a
SELF-REPORT.md; time and tokens are not hidden. Everything shared
(brief, world image, limits, isolation, materials) stays
byte-identical across all three hands.
