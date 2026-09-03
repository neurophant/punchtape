# Campaign 03 — the punchtape three-hand experiment (th3)

Campaign methodology: contracts, operator prompts, the judge and the blind
measurer, metrics discipline — fixed texts, inserted verbatim, never improvised
around.

Hands: 1 solo = GLM-5.3 Max without the machine; 2 machine+top = GLM-5.3 Max;
3 machine+flash = GLM-5.3 Flash High (the `experimental` agent type).
The operators of every hand and the judge — GLM-5.3 Max. Prices — prices.json.

## The wave

A wave is one bank sense across all five stacks at once, one executor per task
(each task carries three hands). Stack order — from the heaviest build to the
lightest (rust → cpp → go → js → python): a staging gap surfaces on the wave's
first run, not its last. Wave 1 — the brim sense (the bank's richest material:
two JPEG fixtures — it exercises the materials path). 15 runs + 5 judgings;
tasks go one stack at a time.

## The task world

    docker run -d --name pt-th3-<sense>-<stack>-<hand> --network none \
      --cpus 1 --memory 1g --pids-limit 512 \
      -v <task-dir>:/run/task \
      -v <punchtape-binary>:/usr/local/bin/punchtape:ro \
      -w /run/task punchtape-world:th3 sleep infinity

- The network is off; the brief sits in /run/task (read-write), the materials
  in /run/task/materials (read-only). The binary is mounted ONLY for the
  machine hands; in the solo world there is no machine.
- The world carries the toolchains (campaigns/world.Dockerfile); the campaign's
  library layer (campaign-03/world.Dockerfile) installs the brief-pinned
  library versions into offline caches; the builds are hermetic.
- The executor platform's metrics are captured at the moment the hand finishes:
  lateness — the cell is lost.
- Cell staging: tools/stage-cell.sh <sense> <stack> <hand> <dir>
  (hand ∈ solo|top|weak = hand 1|2|3); cell check: tools/verify-cell.sh.

## The three-hand experiment

The same bank tasks, each done three ways, plus a judge:

| Hand | Executor | Framework |
|---|---|---|
| machine+top | top model, the executor contract | punchtape in the world |
| machine+flash | cheap flash model, the same contract | punchtape in the world |
| solo | top model, a demanding customer | none — the binary is not mounted |

- Identical inputs: all three hands get one brief, the same materials, one
  world, the same limits and isolation. Only the machine hand knows about the
  framework: it gets the binary mounted and AGENTS.md pulled in; in the solo
  world there is no framework.
- The quality standard is shared, outside the hands: every hand's common input
  is the brief + the expectations list (prompts/EXPECTATIONS.md, frozen before
  the first cell). The machine's verdict is its own hand's internal view
  (Reported data next to mechanical numbers), not a ruler for other cells;
  the verification asymmetry is a subject of measurement.
- The measurements are identical and mechanical, the comparison is pairwise:
  each hand's suite runs twice on its own product under one procedure; the
  five deliverables are counted by one script with a recorded method line;
  coverage is recomputed the same way wherever the stack has tooling. One
  blind agent/script runs the measurements, not knowing which hand made the
  product; a doubt counts against the product, the product is not repaired.
- The operator is a separate supervisor agent, the same top model in every
  hand. In the machine hands it is a human at the forks: a question batch —
  a decision by recommendation, a conflict — an explicit choice. In the solo
  hand it is a demanding customer: it demands the spec, scenarios, tests and
  coverage, drives the "make — check — fix" cycle, runs the checks itself and
  shows the failures, but writes no code and hints no decisions.
- Comparison axes: the base is the solo hand (the vibe-coding status quo);
  "what the machine buys" = the machine hands' delta against solo; a separate
  axis — top vs flash inside machine processing. No hand is the quality
  reference point.

Prompts: prompts/CONTRACT-MACHINE.md (the executor of machine hands 2/3),
prompts/CONTRACT-SOLO.md (the solo executor), prompts/OPERATOR-MACHINE.md
(the operator of hands 2/3), prompts/OPERATOR-SOLO.md (the operator of hand 1),
prompts/JUDGE.md (the judge), prompts/MEASURER.md (the blind measurer),
prompts/EXPECTATIONS.md (the expectations list — the common input). The means
of this IDE's environment (docker exec, the turn mechanics, the ban on
spawning agents and external CLIs) — prompts/HARNESS.md, the same for every
hand, carrying no strategic content.

## The responsiveness protocol (the lead session — the broker)

Nobody polls and nobody sleeps: communication happens only by messages through
the lead session. An executor that hits code 3/4 ENDS its turn and hands over
the question — waiting = the absence of a call, not sleep. The operator is an
event handler: the broker wakes it with a message exactly when there is a
question, its job is one answer; between questions it consumes neither time
nor tokens. The answer goes to the executor immediately — in the same message.
decision_wait_s (the total time from returning the question to sending the
decision) is written into the hand's platform.json; for solo the field is
zero/absent. Operator accounting is separate (the operator.json of its own
hand), never mixed with the hand; platform.json carries only the executor's
active turns.

## Anti-cheating: the ban and the proofs of non-penetration

The hands are categorically forbidden to read the punchtape sources (and
everything outside their own task). Three layers:

1. Inside the container there are no sources: the machine's runtime is one
   compiled binary (Go, sourceless by construction), the smoke check: 0
   punchtape source files in the runtime, version/next work in a
   --network none container. AGENTS.md is the legitimate self-doc.
2. Canaries (tools/canaries.sh): before a wave, random tokens are generated;
   canary files are placed next to the repo, in the control directory, in the
   task and hand directories. The token registry is in the wave journal
   (inaccessible to executors). A leak: canary_leaks in evidence.json →
   a red line of the report, the hand's data under question.
3. Transcript scan: every hand's full command journal is scanned for host
   paths and foreign names (host/repo paths, the names of neighboring tasks
   and containers). Hits — transcript_scan in evidence.json → a red line of
   the report. An empty scan + intact canaries = a "clean" line in the proof
   block.

## The metrics discipline

- Numbers come only from two sources: the instance telemetry
  (.punchtape/ledger.yamll + journal.yamll) and the executor platform's
  metadata (time, tokens, tool calls). Self-reports do not count as numbers.
- The executor's metrics are captured at the moment the hand finishes; a lost
  cell stays lost.
- Rollout discipline: the tools/rollout-watch.sh watcher runs through the
  whole wave (a snapshot every minute into the cell's raw/, turn-break
  detection — a silent rollout for N minutes or a vanished file = the turn
  ended, poke the hand immediately, the snapshot is already saved). The
  finish-time snapshot lands on top of the last minute copy — the platform
  will not have time to trim it.
- Everything the executor writes from the container belongs to root: restore
  ownership before removing the container; after the wave check for zero
  root-owned files.
- Every run updates the metrics journal in the same step as the run itself.
- Role accounting separation: the time/tokens/$ columns — ONLY the hand (its
  platform telemetry and its model's rates); the operator — a separate column
  ($ / tokens / time,s / submissions, wait,s on the machine path), never
  summed or mixed with the hand; the operator's submissions for machine hands
  are code 3/4 decisions, for solo — acceptance cycles. The "full path =
  hand + operator" analytics is the sum of two separately captured numbers,
  not a blended column.
- Methodology overheads (the judge, the blind measurer) — a separate registry
  section per task, not in the hand columns; that is a layer of the
  experiment, not the product path.
- METRICS UNIFORMITY ACROSS ROLES: every role — hand, operator, judge, blind
  measurer — reports ONE AND THE SAME set of fields: time (elapsed_s), tokens
  INPUT / OUTPUT / cache-read separately, tool calls, model, cost. The split
  source is rollout records (tools/rollout-usage.py; the primary copies are
  the watcher's snapshots, which since this wave cover the agents of ALL
  roles, not only the hands); if the platform trimmed a rollout — the
  notification total with an explicit split-loss mark, one fallback form.
  Role counters (submissions for hands, decisions/wait for operators,
  rows/calibration for judges, runs for measurers) ride NEXT TO the uniform
  set, not instead of it.

Files captured per hand (in its evidence/th3/<cell>/ directory):

| File | What is inside | Who captures |
|---|---|---|
| meta.json | sense, stack, hand, model, container | the run |
| platform.json | elapsed_s (ONLY the executor's active turns), tokens_in/out (executor only), tool_calls, decision_wait_s | the run |
| operator.json | SEPARATE operator accounting: model, elapsed_s, tokens_in, tokens_out, decisions | the lead session |
| summary.json | ruling, cases pass/total, tests pass/total, describes/capabilities, families_covered/families_total, cases_linked/cases_runnable, code_coverage_pct, loc; for solo — scenarios/tests by the operator's acceptance | the run at export |
| .punchtape/ (export) | machine telemetry (machine hands only) | the machine |
| judge.json | benchmark_rows, pass, fail, benchmark_sha256 | the judge |
| evidence.json | brief_sha256, materials {path: sha}, materials_after, world_image_digest, limits, network_mode, mounts, punchtape_binary_sha256 (null for solo), product_scan [], canary_leaks [], transcript_scan [] | the run |
| findings.json | [{kind, text}] — the hand's pains and findings | operators, judge, run |
| measure.json | suite_double_green, suite_method, suite_runs, deliverables, deliverables_count, coverage_tooling | the blind measurer |

## The wave report (one command)

    python3 campaigns/tools/report.py <wave-dir> \
        --prices prices.json --out wave-report.md

The report carries: the proofs of identical conditions and no cheating
(brief/materials/image/limits identical, network off, mounts only its own,
materials unchanged, solo without the binary, canaries/transcripts/product
clean — any violation a red line); the metrics table by hands (time, tokens,
cost, submissions accepted/rejected, verb-wall, submit-p95, Reported verdict,
cases, tests, suite×2, deliverables 0–5, prose/capability, coverage, LOC);
the summary and deltas (what the machine buys: solo→top and solo→flash; top
vs flash); comparability (identical measurements, method lines); operator
overheads as a separate section; the registry of problems, pains and findings.

## The wave journal and the launch checklist

Wave journal: evidence/th3/WAVE-JOURNAL.md (judging results + telemetry, the
summary after the wave; the canary registry — there too). Wave launch
checklist: the world built (punchtape-world:th3, cache smoke), the binary of
the double build with one sha (frozen, the sha in the wave journal), the brief
and materials laid out into the task directories with recorded sums, the
prices in prices.json, the canaries generated and laid out, the subagents
defined (hand 3's executor — `experimental`; the operators and the judge —
ordinary subagents with the prompts above).

## The metrics directives (formalized)

1. COMPLETENESS: metrics are collected for EVERY role of the run — hands,
   operators, judges, blind measurers — without exceptions; tokens are
   collected both on input and output for every role.
2. UNIFORMITY: every role reports the same set of fields — time, tokens
   in/out/cache-read separately, tool calls, model, cost; role counters
   (submissions, decisions/wait, judge rows, measurer runs) accompany the set
   but do not replace it. The registry tables' form is uniform for all roles.
3. SEPARATION: the metrics of a hand, an operator and the methodology layers
   are never summed or mixed in one column; methodology overheads (judge,
   measurer) are accounted as a separate registry section.
4. INVIOLABILITY: metrics are never lost under any circumstances (the source
   map and the reconciliation order — in the section below).
5. PROVABILITY: every registry number is reproducibly extracted from the
   platform sources (the session database and the agents' metadata) by the
   committed extractor (tools/usage-ledger.py); the wave ledger is committed
   (evidence/…/ledger/); a re-run of the extractor MUST reproduce the ledger
   byte-for-byte; an agent's lifetime total = the sum of the notification
   segment totals, the exact split = the metadata usage block (validated by
   matching the segment total), the uncovered remainder is billed at the
   input rate with an "upper estimate" mark; the split coverage is given in
   percent. A number that cannot be reproduced this way does not enter the
   registry.
6. CAPTURE CHANNELS: the rollout watcher covers the agents of all roles from
   launch; at every turn completion the driver copies the agent's
   metadata.json into the cell's evidence and writes the notification segment
   total into driver.log with a UTC stamp.

## Metric inviolability — the rule

Losing metrics is unacceptable under any circumstances. Every agent of every
role (hand, operator, judge, measurer) is accounted for in full. WHERE and
WHAT to look at, in order of trust:

1. The executor platform's telemetry — the agent's metadata.json (the usage
   field: inputTokens, outputTokens, cacheReadTokens), totalDurationMs,
   totalToolUseCount. Valid for the LAST segment of the agent's life (for
   multi-resume ones it covers the last turn; for single-segment ones — the
   whole life: checked by in+out matching the notification total). COPIED by
   the driver into the cell's evidence at every turn completion — before
   everything else.
2. The turn-completion notification (the usage block: subagent_tokens,
   tool_uses, duration_ms) — the segment TOTAL; the driver must record EVERY
   one in driver.log (UTC stamp + id + the numbers). The lifetime total =
   Σ segments.
3. The platform's live rollout store (model-io files keyed by agent-id) —
   HOLDS ~2-3 files and trims them every minute; the only source of the
   per-call split. Never rely on it as storage.
4. The watcher's snapshots (tools/rollout-watch.sh, key=agent-id, including
   operators/judges/measurers) — the primary copies of the per-call records;
   the rotation files of one turn OVERLAP (a cumulative rollout) — the sums
   are only an upper bound, do not add them into money.
5. The check before closing a cell: metadata (in+out) == the last
   notification segment total; Σ segments == the lifetime one; the split
   coverage = min(split, life)/life in percent; the life−split remainder — at
   the input rate with an "upper estimate" mark. A discrepancy — stop and
   dissect, not "close enough".

## Carrying a solo row between runs — the rule

Solo does not use the machine: its metrics depend only on the brief, the
world, the contract, the operator's prompt and the prices. If a run repeats
the same task (the same brief) and everything listed matched byte-for-byte —
the solo row may NOT be re-run but carried over from the previous run with a
"transfer, wave N" mark and a proof of identity (the brief's sha256, the
image digest, the prompt hashes, the prices). Gate conditions: (a) the
identity is mechanically proven by the recorded sums; (b) the mark stands in
the metrics registry and in the wave report; (c) inside one wave the
comparability does not break — the transfer is either for the whole wave or
for no cell at all. Transfers between different tasks (different briefs) are
forbidden. The rule does not touch the machine: the machine hands are always
re-run — their result is the subject of the measurement.

## Materials and fixtures — the rule (wave 1)

All materials and fixtures are prepared IN ADVANCE: creating test materials
from scratch is not the hands' duty. Test materials are part of the
acceptance and test-passing criteria: the set (including the edge cases named
by the briefs in the error paths — tiny/truncated/garbage/wrong-format files)
is prepared before the wave from the bank's source material pool and laid out
into /run/task/materials identically for every hand. A hand does not
synthesize test binary data byte-by-byte — that is acceptance plumbing, not a
wish. Wave 1 ran with two primary materials (the edge fixtures the hands
built themselves, identically for all hands — parity kept, the friction is
visible in the economics); from the next wave the set is complete in advance.
Wave 1's materials — a byte-for-byte net of the source material pool (the
sha256 sums are recorded in the wave journal and the cells' evidence).
