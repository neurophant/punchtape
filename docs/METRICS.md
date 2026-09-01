# METRICS — the single living file of runs and comparison baselines

The one metrics document: baselines, every new run, summaries.
Self-contained. It is updated in the same step as the run it describes;
numbers come only from instance ledgers and executor platform metadata
(no self-reports without machine evidence); one parser validates them
all. Primary run evidence (ledgers, journals, verdicts, state) is kept
by the operator outside this repository; the campaign methodology —
including how to reproduce a run — is `campaigns/README.md`.

## 1. How metrics are taken

- **Machine side** (instance ledger, the records between `---`): verbs
  by details.name; submissions accepted/rejected; verb-wall = Σ
  wall-ms; submit p95 = p95(wall − runs) over submissions;
  reply-bytes; runs-ms.
- **Executor side** (the agent platform's metadata for the subagent:
  totalDurationMs, totalTokens, totalToolUseCount). Taken IMMEDIATELY
  when the executor finishes — even without a notification (a lost
  cell is a lost cell).
- Tokens are comparable across runs of the same battery when the model
  and task form match; different harnesses are noted honestly in the
  tables. Executor time includes taking the metrics (+1–2 min over the
  clean wall; the baseline's wall comes from ledger timestamps).
- Containers are removed immediately; the ledger and state are
  exported BEFORE removal; heavy artifacts never enter commits.
- File ownership: everything an executor writes into the mounted task
  directory from a container belongs to root (the docker daemon) — the
  host user cannot read it. RULE: before removing each task container,
  `docker exec <c> chown -R <uid>:<gid> /run/task`; after each wave,
  verify no root-owned files remain.

## 2. BASELINE: the full S-layer run, campaign-02 (2026-08-30)

25 tasks (5 senses × 5 stacks), 25/25 READY+PASS, 0 quarantines;
80.31 M tokens; 53 rejected submissions. Walls — ledger; tokens —
platform.

### Campaign waves (sense × 5 stacks)

| Wave | Walls, min | Tokens, M | Rejected submissions |
|---|---|---|---|
| conduit | 19.3–29.0 | 21.19 | 22 |
| pressmark | 13.6–26.0 | 18.76 | 13 |
| errand | 14.4–34.8 | 19.91 | 5 |
| brim | 10.4–22.7 | 8.98 | 4 |
| quire | 10.7–20.7 | 11.47 | 9 |

### Per task (the baselines for the next full run)

| Task | Wall, min | Submissions +/− | Questions | Tokens, M |
|---|---|---|---|---|
| conduit-python | 29.0 | 10/5 | 1 | 5.07 |
| conduit-go | 22.5 | 14/4 | 1 | 4.20 |
| conduit-rust | 20.0 | 14/3 | 1 | 2.66 |
| conduit-cpp | 19.3 | 12/5 | 1 | 6.15 |
| conduit-js | 21.5 | 6/5 | 1 | 3.11 |
| pressmark-python | 13.6 | 7/3 | 1 | 2.14 |
| pressmark-go | 17.0 | 7/3 | 1 | 2.81 |
| pressmark-rust | 17.6 | 7/2 | 1 | 2.46 |
| pressmark-js | 20.8 | 11/1 | 1 | 4.33 |
| pressmark-cpp | 26.0 | 8/4 | 1 | 7.02 |
| errand-python | 18.6 | 9/1 | 1 | 1.92 |
| errand-go | 17.3 | 11/1 | 2 | 2.91 |
| errand-rust | 14.4 | 6/1 | 0 | 1.50 |
| errand-cpp | 15.6 | 10/2 | 2 | 3.13 |
| errand-js | 34.8 | 14/0 | 1 | 10.45 |
| brim-python | 17.2 | 8/2 | 1 | 2.05 |
| brim-go | 10.4 | 5/0 | 0 | 0.79 |
| brim-rust | 22.7 | 5/1 | 1 | 1.65 |
| brim-cpp | 15.3 | 4/0 | 0 | 2.24 |
| brim-js | 14.3 | 11/1 | 2 | 2.25 |
| quire-python | 11.7 | 10/1 | 1 | 1.62 |
| quire-go | 11.8 | 6/3 | 1 | 1.63 |
| quire-rust | 20.7 | 10/1 | 1 | 4.28 |
| quire-cpp | 14.8 | 7/2 | 1 | 3.01 |
| quire-js | 10.7 | 7/2 | 1 | 0.93 |

Honest notes on the baseline: the brim briefs predate a later brief
fix (they contained a false `--band` example); the pressmark-python
brief predated a defaults cleanup (it hinted the slugify defaults).
The comparison run below used the corrected briefs; the fix waves
cross-check both.

## 3. Control fix waves (brim × 5)

Eight runs of five tasks, all 5/5 READY+PASS (the wave after the spec
rework: the prose gate fired in all five, at the price of +1
submission per task; the wave after the trailing-newline fix: all
five hands' trailing `\n` expectations accepted on the first try,
zero false reds of the class, the best token sum of the waves at
6.17 M).

### cw8, the latest wave: time/tokens/submissions

| Task | Wall, min | Tokens, M | Submissions +/− | p95, ms |
|---|---|---|---|---|
| brim-python | 8.3 | 0.43 | 4/0 | 132 |
| brim-go | 10.2 | 1.02 | 8/1 | 129 |
| brim-rust | 12.5 | 1.90 | 5/0 | 140 |
| brim-cpp | 10.5 | 1.22 | 6/0 | 114 |
| brim-js | 12.1 | 1.60 | 10/1 | 178 |
| **Median/sum** | **10.5** | **6.17** | 33/2 | 114–178 |

### Executor wall time, min (baseline → waves)

| Task | Base | cw1 | cw2 | cw3 | cw4 | cw5 | cw6 |
|---|---|---|---|---|---|---|---|
| brim-python | 17.2 | 15.3 | 11.2 | 12.0 | 15.5 | 7.1 | 7.5 |
| brim-go | 10.4 | 8.7 | 11.5 | 10.1 | 11.5 | 9.2 | 8.1 |
| brim-rust | 22.7 | 26.8 | 10.7 | 18.1 | 17.3 | 20.3 | 13.5 |
| brim-cpp | 15.3 | 18.9 | 10.1 | 118.0† | 11.5 | 12.7 | 7.4 |
| brim-js | 14.3 | 20.9 | 17.7 | 18.9 | 13.2 | 16.9 | 10.2 |
| **Median** | **15.3** | 18.9 | **11.2** | 18.1‡ | 13.2 | 12.7 | **8.1** |

### Executor tokens, M

| Task | Base | cw1 | cw2 | cw3 | cw4 | cw5 | cw6 |
|---|---|---|---|---|---|---|---|
| brim-python | 2.05 | 1.61 | 0.84 | 1.38 | 2.19 | 0.76 | 0.81 |
| brim-go | 0.79 | 0.81 | 1.51 | 1.62 | 1.07 | 1.30 | 0.93 |
| brim-rust | 1.65 | 2.45 | 1.37 | 1.76 | 1.79 | 2.68 | 2.27 |
| brim-cpp | 2.24 | 2.32 | 1.03 | 1.89 | 1.49 | 2.11 | 1.19 |
| brim-js | 2.25 | 3.21 | 2.53 | 3.21 | 1.93 | 1.74 | 1.12 |
| **Sum** | **8.98** | 10.40 | 7.28 | 9.86 | 8.47 | 8.59 | **6.32** |

### Machine metrics (summary)

| Metric | Base | cw1 | cw2 | cw3 | cw4 | cw5 | cw6 |
|---|---|---|---|---|---|---|---|
| Rejected submissions (sum) | 4 | 6 | 3 | 0 | 1 | 11* | 1 |
| Question batches | 5+ | 5 | 2 | 0 | 1 | 3 | 3 |
| submit p95, ms (min–max) | 282–481 | 79–116 | 90–235 | 101–149 | 92–151 | 94–212 | 106–167 |
| Ceremony (verb−runs), s/task | 0.6–1.9 | 0.1–0.3 | ≈ | 0.1–0.3 | 0.1–0.3 | 0.1–0.3 | 0.02–2.0 |
| reply-bytes, KB (sum) | 200 | 203 | 164 | 213 | 183 | 167 | 141.5 |

† cw3-cpp — a materials incident (the executor overwrote its rw copy
of the materials and spent ~2 h restoring them; from the next wave on,
materials are mounted read-only per stack).
‡ without the cpp incident the cw3 median is 12.0.
\* cw5: all 11 — the executors' own YAML typos, each named by a
one-line reason (the diagnostics worked); no machine rejections.

## 4. The full S-layer run on the campaign build — DONE 2026-08-30

25 tasks (5 senses × 5 stacks), each run exactly ONCE: **25/25 VERDICT
READY + acceptance PASS**, 0 quarantines, code and canon untouched
during and after the run. Waves: conduit → pressmark → errand → brim →
quire; five executors in parallel; the staging of the campaign image,
read-only per-stack materials, the binary double-built and verified
byte-for-byte with the version stamped in, network off, 1 CPU / 1 GB /
512 pids per world, fresh directories per task, the `# TOKENS n/a`
trailer convention; metrics from the ledger plus platform metadata
taken immediately; containers removed at once, zero root-owned files
after every wave. Evidence: 25 × ledger/verdict/state, kept by the
operator.

### Executor wall time, min (baseline §2 → this run)

| Task | Base | Run | | Task | Base | Run |
|---|---|---|---|---|---|---|
| conduit-python | 29.0 | 14.8 | | errand-python | 18.6 | 13.5 |
| conduit-go | 22.5 | 11.7 | | errand-go | 17.3 | 12.8 |
| conduit-rust | 20.0 | 25.3 | | errand-rust | 14.4 | 19.4 |
| conduit-cpp | 19.3 | 19.5 | | errand-cpp | 15.6 | 11.5 |
| conduit-js | 21.5 | 17.2 | | errand-js | 34.8 | 11.9 |
| pressmark-python | 13.6 | 11.0 | | brim-python | 17.2 | 9.8 |
| pressmark-go | 17.0 | 16.0 | | brim-go | 10.4 | 8.8 |
| pressmark-rust | 17.6 | 18.9 | | brim-rust | 22.7 | 16.6 |
| pressmark-cpp | 26.0 | 18.6 | | brim-cpp | 15.3 | 11.9 |
| pressmark-js | 20.8 | 11.0 | | brim-js | 14.3 | 14.4 |
| quire-python | 11.7 | 11.3 | | quire-cpp | 14.8 | 16.7 |
| quire-go | 11.8 | 10.9 | | quire-rust | 20.7 | 16.4 |
| quire-js | 10.7 | 11.2 | | **Median** | **17.3** | **13.5** |

Down 18/25, up 7 (conduit-rust +5.3 and errand-rust +5.0 — the
price of cargo builds; quire-cpp +1.9; pressmark-rust +1.3;
quire-js +0.5; conduit-cpp +0.2; brim-js +0.1), none flat.

### Executor tokens, M (base → run)

| Task | Base | Run | | Task | Base | Run |
|---|---|---|---|---|---|---|
| conduit-python | 5.07 | 0.80 | | errand-python | 1.92 | 1.20 |
| conduit-go | 4.20 | 0.78 | | errand-go | 2.91 | 1.40 |
| conduit-rust | 2.66 | 3.51 | | errand-rust | 1.50 | 2.44 |
| conduit-cpp | 6.15 | 1.88 | | errand-cpp | 3.13 | 1.19 |
| conduit-js | 3.11 | 0.85 | | errand-js | 10.45 | 1.10 |
| pressmark-python | 2.14 | 0.98 | | brim-python | 2.05 | 0.69 |
| pressmark-go | 2.81 | 3.42 | | brim-go | 0.79 | 1.02 |
| pressmark-rust | 2.46 | 2.96 | | brim-rust | 1.65 | 0.81 |
| pressmark-cpp | 7.02 | 2.83 | | brim-cpp | 2.24 | 1.75 |
| pressmark-js | 4.33 | 1.27 | | brim-js | 2.25 | 2.19 |
| quire-python | 1.62 | 1.02 | | quire-cpp | 3.01 | 2.59 |
| quire-go | 1.63 | 1.49 | | quire-rust | 4.28 | 1.82 |
| quire-js | 0.93 | 1.01 | | **Sum** | **80.31** | **41.00** |

−49% overall; the sharpest single drops: errand-js 10.45→1.10 (bare
imports resolved from the shared node_modules, zero directory walks),
conduit-cpp 6.15→1.88. Up: conduit-rust, pressmark-go/rust, errand-rust, brim-go and
quire-js (the executors ran more probe sorties; within budget).

### Submissions and questions (base → run)

| Task | Subm. +/− | Questions | | Task | Subm. +/− | Questions |
|---|---|---|---|---|---|---|
| conduit-python | 10/5 → 5/0 | 1 → 1 | | errand-python | 9/1 → 7/1 | 1 → 0 |
| conduit-go | 14/4 → 5/0 | 1 → 0 | | errand-go | 11/1 → 6/0 | 2 → 0 |
| conduit-rust | 14/3 → 11/1 | 1 → 0 | | errand-rust | 6/1 → 6/1 | 0 → 0 |
| conduit-cpp | 12/5 → 6/0 | 1 → 0 | | errand-cpp | 10/2 → 6/0 | 2 → 0 |
| conduit-js | 6/5 → 4/0 | 1 → 0 | | errand-js | 14/0 → 4/1 | 1 → 0 |
| pressmark-python | 7/3 → 5/1 | 1 → 0 | | brim-python | 8/2 → 4/0 | 1 → 0 |
| pressmark-go | 7/3 → 4/0 | 1 → 0 | | brim-go | 5/0 → 7/1 | 0 → 0 |
| pressmark-rust | 7/2 → 5/0 | 1 → 0 | | brim-rust | 5/1 → 5/0 | 1 → 0 |
| pressmark-cpp | 8/4 → 5/0 | 1 → 0 | | brim-cpp | 4/0 → 6/0 | 0 → 0 |
| pressmark-js | 11/1 → 4/1 | 1 → 0 | | brim-js | 11/1 → 9/1 | 2 → 0 |
| quire-python | 10/1 → 6/2 | 1 → 1 | | quire-cpp | 7/2 → 5/0 | 1 → 0 |
| quire-go | 6/3 → 6/0 | 1 → 0 | | quire-rust | 10/1 → 5/0 | 1 → 0 |
| quire-js | 7/2 → 6/2 | 1 → 1 | | **Sum** | **219/53 → 142/12** | **25 → 3** |

All 3 questions (conduit-python: the language of event strings;
quire-python and quire-js: the "one product" question) were closed
autonomously — the recommended default, journaled. Machine fix cycles:
conduit-rust 2, conduit-cpp 1, errand-rust 1, the rest 0 (the
baseline counted cycles in waves 2–6: 3).

### Machine side (ledgers; one parser)

| Metric | Run (min–max) | Budget |
|---|---|---|
| submit p95 overhead, ms | 46–175 | 300 — all within budget |
| verb-wall (machine), s | 1.9 (quire-go) – 206.5 (errand-rust) | honest runs: runs-ms ≈ verb-wall |
| reply-bytes, KB/task | 24.2 – 152.6 | — |

### Wave summary (tokens M / median wall min / rejections / questions)

| Wave | Baseline | Run |
|---|---|---|
| conduit | 21.19 / 21.5 / 22 / 5 | 7.82 / 17.2 / 1 / 1 |
| pressmark | 18.76 / 17.6 / 13 / 5 | 11.46 / 16.0 / 2 / 0 |
| errand | 19.91 / 17.3 / 5 / 6 | 7.33 / 12.8 / 3 / 0 |
| brim | 8.98 / 15.3 / 4 / 4 | 6.46 / 11.9 / 2 / 0 |
| quire | 11.47 / 11.8 / 9 / 5 | 7.93 / 11.3 / 4 / 2 |

Honest notes: the baseline ran on the earlier briefs (see §2); this
run used the corrected briefs and the campaign build whose slot texts
are a little shorter (reply-bytes slightly below the baseline class).
The targeted expectations held: conduit/errand — zero PATH detours
(the PATH orchestration raised no obstacles, the stands lived in the
strings); pressmark — zero questions; js — no vendored copies
(errand-js 10.45→1.10 M); quire — the QC noise down to 2, both from an
executor's flow-list artifact.

## 5. The three-hand run — DONE 2026-08-31

25 tasks × 3 hands + 25 judges (100 subagents; the system hands ran
the machine build validated by the latest control wave; staging and
contracts — `campaigns/README.md`; per-task tables kept by the
operator). The three hands: the same task done (a) by a top model
through the machine, (b) by a cheap flash model through the machine,
(c) solo without the machine; a fourth AI judge checked each solo
product against the byte-exact checks table of the top hand of the
same task.

### Verdicts and hands

- **System hands: 50/50 VERDICT READY + acceptance PASS** (25 top +
  25 flash), 0 quarantines; machine questions closed by defaults with
  explicit recorded answers (a human was never needed).
- **Solo: 25/25 products with green self-checks** (their own stands
  and batteries: 12 to 283 checks per task).

### Hand economics (25 tasks; sum / median / min–max)

| Hand | Tokens, M | Wall, min |
|---|---|---|
| system+top | 65.31 / 2.23 / 0.60–7.66 | 435.6 / 14.4 / 8.2–33.8 |
| system+flash | 72.78 / 2.32 / 0.66–10.05 | 756.7 / 29.0 / 11.7–61.9 |
| solo (no machine) | 45.22 / 1.41 / 0.39–5.46 | 339.9 / 11.4 / 7.1–22.5 |

The headline economic fact: solo is the cheapest AND the fastest hand
(−31% tokens and −21% median wall against top); flash is the most
expensive AND the slowest (+11.4% of tokens and twice the top median): a cheap model pays with more steps. The machine does not save tokens by
itself — it buys something else with them: see quality.

### Quality: the judges (solo against the top hand's byte-exact table)

Strict count: **108/491 rows green (22%)**; by wave: conduit 35%,
pressmark 27%, errand 0%, brim 21%, quire 21%. The pattern (consistent
across all 25 judges):

1. **Solo's functional core is almost always byte-for-byte green**:
   measured numbers, built files, copies, determinism, isolation —
   they match the top hand's table.
2. **The red mass is protocol, not function**: diagnostic texts
   (prefixes, quotes, wording), exit codes (2 against 1), usage
   strings, the shape of ok-lists.
3. **Unclosed brief forks spread across hands**: the same question
   (the order of ok lines, the shape of paths, the semantics of a
   threshold, rendering missing data) different hands answer
   OPPOSITELY — top hands too (quire: three top hands, three different
   answers).
4. **Errand 0% — a surface artifact**: the brief asks for a library;
   the top hand's table drives it through the top hand's own test
   driver, which solo does not have; the CONDITIONAL count (the
   judge's minimal driver by the row's meaning): errand-go 11/11,
   errand-rust 24/24, errand-cpp 8/11, errand-js 8/10 — 51/56:
   solo's library behavior is nearly complete.
5. Real functional failures of solo are rare and pinpoint — five per
   battery, invisible to the self-checks: an RST cut instead of a
   clean FIN (conduit-rust), the exit code on a taken port
   (conduit-rust), the name field missing in the record API
   (quire-rust), repeated options silently accepted (pressmark-js),
   "None" instead of an empty value (quire-python).

### The experiment's conclusion (the system's value)

Without the machine the executor is BETTER on economics and WORSE on
convergence with someone else's contract: solo's "done" is its own
opinion of done; a byte-exact table of executable expectations costs
20–45% of solo's savings and buys comparability. On this battery the
flash loop loses to the top loop economically: a cheaper model — a
more expensive path. The system's verifiable benefit is not tokens but
RECONCILABILITY: 50/50 machine verdicts against 22% strict byte
convergence of solo.

### Run findings

Machine fixes landed from this run (trailing `\n` of expectations,
recipe placeholder naming, identical commands with different
expectations, the delta writer losing a newline) plus staging and
environment factors (GOMEMLIMIT against the 1-GB limit on go builds,
a zombie redis, lost harness notifications, replay wedging on
read-only materials) — details kept with the run evidence. Feature
prose was authored voluntarily by 3 of 50 hands (the prose gate is
field-justified).

## 6. The freshness rule

Every run updates this file in the same step (tables + summary); a
discrepancy between the file and reality is a defect of that step.
This file holds only numbers, the rules of taking them, and honest
methodology notes.
