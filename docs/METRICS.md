# The registry of run metrics and conclusions (cumulative)

One file of all measurements of all runs: per task and cumulative, compared
across hands. Numbers come only from the captured cell files and the platform
ledger (the full provenance of every metric is §1). The quality/completeness
benchmark is the brief + the expectations list, frozen before the first cell;
no hand is the reference point. Hand comparison: (a) the suite ×2 on each
hand's own product under one procedure, (b) deliverables by one counter with a
method line, (c) coverage by recomputation; economics: the base is solo, the
delta of the machine hands is "what the machine buys"; top vs flash is the
second axis.

## 1. Metric provenance and methodology

Every column of every table below is defined here once: where the number is
born (the exact file), how it is computed, and why the method is what it is.
Nothing in this registry is retyped by hand; each value is either read from a
captured file or derived from captured files by a stated formula.

### 1.1 The three evidence tiers

- **T1 — the machine's own telemetry.** The instance's `.punchtape/ledger.yamll`
  inside each cell: one record per machine verb with `wall-ms` / `runs-ms`, the
  accepted/rejected outcome of every `submit`, and the final `VERDICT:` line.
  Parsed by `campaigns/tools/parse-ledger.py` / `campaigns/tools/report.py`.
  The machine measures itself; these numbers are the submission counts,
  verb-wall, submit-p95 and the verdicts.
- **T2 — the cell captures, taken at the hand's finish.** `platform.json`
  (elapsed, tokens in/out/cache, tool calls, decision wait), `operator.json`
  (the operator agent's time/tokens/decisions), `summary.json` (the exported
  case table and prose counts, LOC), `measure.json` (the blind measurer's
  result), `judge.json` (the judge's result), `evidence.json` (identical
  conditions, scans). Written by `campaigns/campaign-03/tools/collect-metrics.sh`
  from the executor platform's rollout records and completion notifications —
  the same platform source for every role.
- **T3 — the platform ledger, the proof tier for token accounting.** The
  extractor `campaigns/campaign-03/tools/usage-ledger.py` reads the platform's
  session database (~/.zcode/cli/db/db.sqlite: every completed turn carries a
  usage block) and emits, per agent, the lifetime token total, the segment
  count, and the exact in/out/cache split of the last segment's metadata. This
  tier is immune to rollout pruning — where T2 and T3 diverge, T3 wins and the
  divergence is recorded (the reconciliation notes below).

### 1.2 The metric dictionary

| metric | origin | method | justification / caveats |
|---|---|---|---|
| time, s — cell tables (wave 1, validation wave, comparison tables) | T2 `platform.json` `elapsed_s` | the driver's capture of the hand's wall at finish | for re-staged hands it includes all model turns incl. restarts; where the capture was lost or later reconciled, the T3 value is used and the row is marked † |
| time, s — the wave S uniform tables | T3 ledger: the sum of the durations of **all** the agent's segments | the full wall of the hand, rollout rotations and restarts included; for single-segment cells it equals that segment | earlier revisions printed only the last segment's duration (the extractor's per-part `totalDurationMs`) — that understated 21 multi-segment cells by up to 6164 s; the registry now publishes the full wall, and the archived `role-summary.json` aggregate still carries the old last-segment basis for its time row |
| life, tokens | T3 `lifetime_tokens_db` | the sum of usage blocks over all the agent's completed turns in the platform DB | the proof of token accounting; a cell closed before the uniformity rule carries the T2 notification total instead, with a mark |
| segments | T3 | the count of platform session parts | every part is one rollout rotation; life sums them all |
| split in/out | T3, the last segment's `metadata.json` usage block | `inputTokens` / `outputTokens` | validated by matching the segment total (`split_matches_last_segment`); where the platform trimmed the metadata the split is restored from the committed ledger |
| split coverage | T3 | `split_total / life`, in % | how much of life the exact split covers; the remainder is priced as an upper estimate (see $) |
| cache | T3 `cacheReadTokens` of the same split | as recorded by the platform | the cache-aware price needs it; cells captured before the rule carry no split and their $ is marked as an upper estimate |
| tools | T3 `totalToolUseCount` / T2 `tool_calls` | as recorded | role counter alongside the token counters |
| $ — wave 1 and validation-wave tables | T2 tokens + `prices.json` | cache-aware over the cell capture: `(in − cache)·rate_in + cache·rate_cached + out·rate_out` | the captured cache makes it a measured value; where the capture lost the cache the value degenerates to the input-rate upper estimate and is marked * |
| $ — wave S uniform tables (hands, operators, layers) | T3 + `prices.json` | the same cache-aware formula over the split, **plus** `(life − split_total)·rate_in` for the uncovered remainder | exact at 100 % coverage; at partial coverage the remainder is priced at the full input rate — an upper estimate, marked by the coverage column itself |
| submissions ±/− | T1 instance ledger | the count of `submit` transactions accepted / rejected | the machine's own gate decisions, not self-reports |
| verb-wall, s | T1 | the sum of `wall-ms` across all verb calls | the machine's own answer-processing time |
| submit-p95, ms | T1 | p95 of (`wall-ms` − `runs-ms`) over submit transactions | the ceremony-share gate metric; the budget is 300 ms |
| Reported verdict | T1 | the `VERDICT:` line parsed from the instance telemetry | the machine's signed stop line |
| cases | T2 `summary.json` | machine hands: the exported case table size; solo: the product's own suite size (stated per row) | the two are never equated — the blind measurer runs each product's own exported table |
| suite ×2 | T2 `measure.json` | the blind measurer ran the product's exported case table twice, each run in a fresh throwaway container (`--network none --cpus 1 --memory 1g`), exact byte compare of exit/stdout/stderr/disk; tolerant only for trailing \n/spaces (`case-runner.py`) | double-run determinism is the anti-"green on a cache" check |
| deliverables | T2 `measure.json` | one counter, five items (spec, code, scenarios, double green, coverage), each with a recorded method string (`count-deliverables.py`) | hand units differ constructively; nobody's unit is equated to another's |
| prose/capab | T2 `summary.json` | the machine's feature-prose gates: describes / capabilities | the machine's own count of the human-readable spec docs |
| coverage | T2 `measure.json` | recomputed by the measurer: line coverage where the stack's tool exists offline; specdocs-linkage where it does not; `n/a` declared honestly | one recomputation method per stack, never a self-report |
| LOC | T2 `summary.json` | code lines by extension, machine dirs excluded | volume reference only |
| operator time / tokens / decisions | T2 `operator.json`; ledger (T3) where the capture was lost (marked †) | the operator agent's own capture | the acceptance cycles (solo) and the code-3/4 micro-decisions (machine hands) |
| decision wait | T2 `platform.json` `decision_wait_s` | how long the hand waited for the human answer | recorded per cell where a question batch fired |
| judge pass / rows | T2 `judge.json` | the judge executes the top hand's exported table row by row on the other hand's product, byte-strict, no right to edit, a doubt counts against the product; calibration: 3 reference rows must pass on the reference artifact before the count is recorded | cross-product reconciliation; the strict count is the published one, conditional counts (a neutral driver) are named separately |
| red lines | T2 `evidence.json` | transcript and product scans for neighbor/home patterns + canary tokens | the anti-cheating evidence |
| question batches / escalations | T2 firings + the driver log | cells where an operator agent fired or a code-3/4 turn is logged | counted per cell, never per batch, unless stated |

### 1.3 Aggregation rules

- Per-hand summary means/medians are computed over the 20 non-brim wave S
  cells (the `role-summary.json` mechanical aggregate; brim belongs to wave 1
  and its registry rows are wave-1 rows). Campaign-wide counts (75 hands,
  25 tasks) include brim and say so.
- The full path = hand $ + its operator $; hands and layers are never mixed
  into one sum (layers are methodology overhead, not the product path).
- Comparison tables pair the same task across runs; totals are sums over all
  paired rows, and the pair count is printed when a pairing is partial.

### 1.4 Prices

`campaigns/campaign-03/prices.json` — the market tariffs of 2026-09-03
(GLM-5.3 Max: in 1.4 / out 4.4 / cached-in 0.26 $ per M; GLM-5.3 Flash High:
0.15 / 0.5 / 0.03), cross-checked against several independent tariff
summaries; recorded in the campaign and unchanged inside every comparison
below. Two bases appear in this registry and are always named: the
**cache-aware** $ (measured, needs the split/cache), and the **input-rate
upper estimate** (all tokens at the full input rate — used where the split is
absent, e.g. the operators' $0.60 of the validation wave).

### 1.5 Honesty marks

`†` — the T2 capture was lost or later reconciled; the T3 ledger value is
used (the affected cells are named in the notes). `*` — a cache-less capture:
the value is the input-rate upper estimate. `n/a` — the metric does not exist
for this cell (no tool, no capture) and is not extrapolated. `queued` — the
layer had not closed when that historical table was cut; the final tables
below carry the closed values.

## 2. Wave 1 — brim × 5 stacks × 3 hands = 15 cells (2026-09-03)

The first three-hand wave: one sense (brim) on all five stacks. Basis: T2
cell captures (the T3 ledger did not exist yet); the operator column is the
operator agent's captured cost. `—` in tokens out = the platform trimmed the
rollout before collection; the notification total fills `in`, and those rows'
$ loses the cache split and degenerates to the input-rate upper estimate
(mark *, §1.5).

### brim-rust

| hand | time, s | tokens in/out | hand $ | submissions ±/− | verb-wall, s | Reported verdict | cases | suite×2 | deliv. | prose | coverage | LOC | operator $ / decisions / wait, s |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| solo | 2473.5 | 11695011/134870 | 3.81 | — | — | — (operator acceptance) | 56/56 | green×2 | 5/5 | 1/— | n/a | 2499 | 5.90 / 2 / — |
| top | 1385.0 | 3032490/— * | 4.25 * | 7/1 | 41.8 | VERDICT: READY | 21/21 | green×2 | 5/5 | 3/3 | n/a | 246 | 0 / 0 / 0.0 |
| weak | 1480.1 | 3376366/59988 | 0.14 | 8/5 | 40.8 | VERDICT: READY | 21/21 | green×2 | 5/5 | 3/3 | n/a | 162 | 0 / 0 / 0.0 |

### brim-cpp

| hand | time, s | tokens in/out | hand $ | submissions ±/− | verb-wall, s | Reported verdict | cases | suite×2 | deliv. | prose | coverage | LOC | operator $ / decisions / wait, s |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| solo | 1174.3 | 2645371/— * | 3.70 * | — | — | — (operator acceptance) | 65/65 | green×2 | 5/5 | 1/— | n/a | 931 | 1.75 / 1 / — |
| top | 638.0 | 1942641/37724 | 0.73 | 9/2 | 4.6 | VERDICT: READY | 19/19 | green×2 | 5/5 | 3/3 | n/a | 250 | 0 / 0 / 79.0 |
| weak | 135.7 | 463821/5399 | 0.02 | 8/2 | 3.5 | VERDICT: READY | 18/18 | green×2 | 5/5 | 3/3 | n/a | 200 | 0 / 0 / 0.0 |

### brim-go

| hand | time, s | tokens in/out | hand $ | submissions ±/− | verb-wall, s | Reported verdict | cases | suite×2 | deliv. | prose | coverage | LOC | operator $ / decisions / wait, s |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| solo | 1132.1 | 7048344/71086 | 2.25 | — | — | — (operator acceptance) | 169/169 | green×2 | 5/5 | 1/— | 99.3% | 1473 | 5.07 / 2 / — |
| top | 894.6 | 2843387/9587 | 2.80 | 7/0 | 4.2 | VERDICT: READY | 18/18 | green×2 | 5/5 | 3/3 | n/a | 198 | 0 / 0 / 102.0 |
| weak | 1579.5 | 3714500/4654 | 0.48 | 11/2 | 2.4 | VERDICT: READY | 13/13 | green×2 | 5/5 | 3/4 | n/a | 209 | 0 / 0 / 0.0 |

### brim-js

| hand | time, s | tokens in/out | hand $ | submissions ±/− | verb-wall, s | Reported verdict | cases | suite×2 | deliv. | prose | coverage | LOC | operator $ / decisions / wait, s |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| solo | 1293.5 | 2623141/— * | 3.67 * | — | — | — (operator acceptance) | 48/48 | green×2 | 5/5 | 1/— | n/a | 919 | 1.49 / 3 / — |
| top | 962.7 | 1212079/15787 | 1.12 | 7/0 | 29.6 | VERDICT: READY | 20/20 | green×2 | 5/5 | 3/3 | n/a | 0 | 0 / 0 / 0.0 |
| weak | 61.4 | 529022/2172 | 0.02 | 11/2 | 372.4 | VERDICT: READY | 28/28 | green×2 | 5/5 | 4/4 | n/a | 203 | 0 / 0 / 0.0 |

### brim-python

| hand | time, s | tokens in/out | hand $ | submissions ±/− | verb-wall, s | Reported verdict | cases | suite×2 | deliv. | prose | coverage | LOC | operator $ / decisions / wait, s |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| solo | 729.3 | 3354880/31806 | 1.06 | — | — | — (operator acceptance) | 55/55 | green×2 | 5/5 | 1/— | 100% | 438 | 0.71 / 1 / — |
| top | 224.5 | 722351/15442 | 0.28 | 5/0 | 32.3 | VERDICT: READY | 19/19 | green×2 | 5/5 | 2/2 | n/a | 0 | 0 / 0 / 0.0 |
| weak | 592.8 | 1015368/29736 | 0.05 | 9/1 | 43.8 | VERDICT: READY | 23/23 | green×2 | 5/5 | 3/3 | n/a | 0 | 0 / 0 / 0.0 |

**Cumulative — wave 1** (means per task over the five stacks, hand + operator):

- solo: full path $5.88, active time 1361 s; tasks 5/5.
- top: full path $1.84, active time 821 s; tasks 5/5.
- weak: full path $0.14, active time 770 s; tasks 5/5.

### Wave 1 conclusions

1. QUALITY: against the benchmark (brief + expectations list) there is NO
   difference between hands — 15/15 cells: 5/5 deliverables, suite ×2 green
   for all; the machine hands 10/10 VERDICT READY.
2. ECONOMICS: solo = hand + a heavy operator (verifying reading, 1–3
   acceptance cycles, $0.71–5.90 per task); the machine hands removed the
   operator from the loop: 0–1 decisions per task (decision waits 79/102 s
   in the two cells where the operator answered), operator overhead ≈ $0.
   Full path per cell: solo $1.77–9.71, top+machine $0.28–4.25, flash+machine
   $0.02–0.48 — the cheapest at the same result.
3. TOP VS FLASH: the result is indistinguishable (READY everywhere, 5/5);
   flash is several times cheaper; the price of flash is more rejected
   submissions, premature turn breaks (driver pokes) and one episode of
   host-boundary violation (caught by the anti-cheat).
4. HUMAN ACCEPTANCE catches classes absent from the machine checks of its own
   hand: a silent zero of a knob, a stale spec counter, imprecise imports,
   spec/code desync, a false PASS number — all found by the customer's review,
   not by the executor's self-check. The verification asymmetry is a subject
   of measurement, not a bias.
5. MACHINE AND CANON: one question batch and one escalation per 10 machine
   cells (resolved by the operator in 79/102 s), zero deadlocks, zero phantom
   families, verb-wall 2.4–41.8 s — the fixes hold.

## 3. The uniformity rule (the owner's word, wave S)

Every role (hand / operator / judge / measurer) reports the same set of
fields: elapsed, tokens in/out/cache separately, tool calls, model, $; the
split source is rollout records (`rollout-usage.py`), on trimming — the
notification total with a loss mark. Role counters (submissions, decisions
and waits, judge rows, measurer runs) sit alongside. Cells closed before the
rule carry T2 totals without a split, marked (§1.5); the T3 ledger (§1.1)
reconciles them afterwards.

## 4. Wave S — the uniform tables (the proven form)

Every number below is the T3 ledger tier over the wave's 149 agents
(determinism reconciled with the closed cells). The proof chain: (1) the
LIFETIME TOTAL and the segment count of every agent — from the platform
session database (the extractor saves the ledger JSON; this is the proof, not
the driver's memory); (2) the EXACT SPLIT (in/out/cache) — the agent's
metadata usage block, validated by matching the segment total; (3) the
life−split REMAINDER — priced at the input rate, an upper estimate; the split
coverage in % in every row. Rollout snapshots remain the proof of per-call
traffic and are never summed into money (rotation overlap).
A reconciliation correction: conduit-go-weak lived 3 segments, life
17,904,719 (previously one segment was recorded — a driver bookkeeping error,
caught by reconciling with the database). A basis correction of this revision:
the time column is the sum of ALL segment durations (the full wall); the
previous print carried only the last segment's duration, understating the 21
multi-segment cells (worst: conduit-go-weak 1702 → 7866 s, pressmark-rust-weak
295 → 2942 s).

### Hands

| cell | hand | time, s (wall = all segments) | life, tokens (segments) | split in/out (coverage) | cache | tools | $ | suite / deliverables | submissions |
|---|---|---|---|---|---|---|---|---|---|
| conduit-rust-solo | solo | 2295.1 | 10570871 (2 segm.) | 4473887 / 20379 (43%) | 4446976 | 32 | 9.79 | green×2, 15 cases, deliv. 5/5 | — / 2 acceptance cycles |
| conduit-rust-top | top | 1465.5 | 2957851 (1 segm.) | 2885181 / 72670 (100%) | 2752256 | 39 | 1.22 | green×2, 18 cases, deliv. 5/5 | 11/0 |
| conduit-rust-weak | weak | 1922.3 | 3723051 (1 segm.) | 3647999 / 75052 (100%) | 3538880 | 54 | 0.16 | green×2, 15 cases, deliv. 5/5 | 7/0 |
| conduit-cpp-solo | solo | 1469.1 | 3426337 (1 segm.) | 3351071 / 75266 (100%) | 3261824 | 78 | 1.30 | green×2, 18 cases, deliv. 5/5 | — / 1 cycle |
| conduit-cpp-top | top | 1583.9 | 2948931 (1 segm.) | 2871301 / 77630 (100%) | 2750784 | 35 | 1.23 | green×2, 14 cases, deliv. 5/5 | 7/0 |
| conduit-cpp-weak | weak | 2701.3 | 5160405 (1 segm.) | 5067695 / 92710 (100%) | 4816000 | 53 | 0.23 | green×2, 18 cases, deliv. 5/5 | 9/2 (replacement after a hang) |
| conduit-go-solo | solo | 2894.0 | 10835440 (2 segm.) | 2305953 / 12549 (21%) | 2289280 | 15 | 12.60 | green×2, 21 cases, deliv. 5/5 | — / 2 cycles |
| conduit-go-top | top | 1758.3 | 4290386 (1 segm.) | 4208768 / 81618 (100%) | 4084160 | 46 | 1.60 | green×2, 16 cases, deliv. 5/5 | 9/0 |
| conduit-go-weak | weak | 7866.0 | 17904719 (3 segm.) | 4623621 / 18935 (26%) | 4597952 | 17 | 2.14 | green×2, 20 cases, deliv. 5/5 | 16/0 + a decision + an environment restart |
| conduit-js-solo | solo | 1841.7 | 5988886 (1 segm.) | 5897090 / 91796 (100%) | 5788288 | 71 | 2.06 | green×2, 24 cases, deliv. 5/5 | — / 1 cycle |
| conduit-js-top | top | 1737.7 | 1926944 (1 segm.) | 1842790 / 84154 (100%) | 1726848 | 28 | 0.98 | green×2, 19 cases, deliv. 5/5 | 8/0 |
| conduit-js-weak | weak | 1986.6 | 3279696 (1 segm.) | 3196688 / 83008 (100%) | 3059072 | 37 | 0.15 | green×2, 17 cases, deliv. 5/5 | 6/0 |
| conduit-python-solo | solo | 1959.2 | 6377394 (2 segm.) | 4024924 / 33461 (64%) | 3980992 | 35 | 4.49 | green×2, 292 cases, deliv. 5/5 | — / 2 cycles |
| conduit-python-top | top | 904.7 | 1190945 (1 segm.) | 1134232 / 56713 (100%) | 1043968 | 18 | 0.65 | green×2, 17 cases, deliv. 5/5 | 6/0 |
| conduit-python-weak | weak | 1586.4 | 2126738 (1 segm.) | 2062092 / 64646 (100%) | 1945472 | 29 | 0.11 | green×2, 14 cases, deliv. 5/5 | 6/0 |
| errand-rust-solo | solo | 1056.5 | 4130645 (2 segm.) | 3911698 / 56291 (96%) | 3841664 | 62 | 1.57 | green×2, 13 cases, deliv. 5/5 | — / decision + acceptance |
| errand-rust-top | top | 1386.1 | 6057541 (2 segm.) | 5406481 / 52197 (90%) | 5320384 | 55 | 2.57 | green×2, 8 cases, deliv. 5/5 | 12/0 + escalation |
| errand-rust-weak | weak | 1815.6 | 3643692 (2 segm.) | 3021562 / 33768 (84%) | 2971776 | 35 | 0.20 | green×2, 8 cases, deliv. 5/5 | 10/0 + escalation |
| errand-cpp-solo | solo | 1480.5 | 5675460 (2 segm.) | 1126631 / 9288 (20%) | 1114240 | 11 | 6.70 | green×2, 21 cases, deliv. 5/5 | — / 2 cycles |
| errand-cpp-top | top | 1092.6 | 2119436 (1 segm.) | 2075378 / 44058 (100%) | 1993920 | 35 | 0.83 | green×2, 9 cases, deliv. 5/5 | 7/0 |
| errand-cpp-weak | weak | 1614.1 | 2236906 (1 segm.) | 2173713 / 63193 (100%) | 2076096 | 32 | 0.11 | green×2, 13 cases, deliv. 5/5 | 9/0 |
| errand-go-solo | solo | 1491.7 | 4032375 (2 segm.) | 1371274 / 9232 (34%) | 1356544 | 16 | 4.13 | green×2, 19 cases, deliv. 5/5 | — / 2 cycles |
| errand-go-top | top | 1157.9 | 2911001 (1 segm.) | 2854957 / 56044 (100%) | 2767552 | 46 | 1.09 | green×2, 9 cases, deliv. 5/5 | 10/0 |
| errand-go-weak | weak | 3496.6 | 7133884 (1 segm.) | 7016174 / 117710 (100%) | 6853440 | 70 | 0.29 | green×2, 18 cases, deliv. 5/5 | 9/0 |
| errand-js-solo | solo | 790.3 | 2010315 (1 segm.) | 1960351 / 49964 (100%) | 1899712 | 50 | 0.80 | green×2 (solo suite 27), deliv. 5/5 | — / 1 cycle |
| errand-js-top | top | 1662.6 | 5675115 (2 segm.) | 4917358 / 46695 (87%) | 4851008 | 49 | 2.56 | green×2, 15 cases, deliv. 5/5 | 8/0 + QC |
| errand-js-weak | weak | 2012.5 | 5970571 (1 segm.) | 5903612 / 66959 (100%) | 5793408 | 77 | 0.22 | green×2, 8 cases, deliv. 5/5 | 9/0 |
| errand-python-solo | solo | 532.7 | 1260075 (2 segm.) | 622998 / 8852 (50%) | 581504 | 12 | 1.13 | green×2 (23 tests), deliv. 5/5 | — / 2 cycles |
| errand-python-top | top | 707.8 | 1278392 (1 segm.) | 1234689 / 43703 (100%) | 1165888 | 25 | 0.59 | green×2, 9 cases, deliv. 5/5 | 7/0 |
| errand-python-weak | weak | 2224.0 | 5170032 (1 segm.) | 5087071 / 82961 (100%) | 4964608 | 62 | 0.21 | green×2, 16 cases, deliv. 5/5 | 9/0 |
| pressmark-rust-solo | solo | 1483.9 | 6834952 (1 segm.) | 6743358 / 91594 (100%) | 6615808 | 97 | 2.30 | green×2 (21 scenarios), deliv. 5/5 | — / 1 cycle |
| pressmark-rust-top | top | 905.2 | 2711678 (1 segm.) | 2659450 / 52228 (100%) | 2571968 | 44 | 1.02 | green×2, 22 cases, deliv. 5/5 | 6/0 |
| pressmark-rust-weak | weak | 2942.0 | 7802473 (3 segm.) | 1434520 / 10344 (19%) | 1421568 | 8 | 1.00 | green×2, 24 cases, deliv. 5/5 | 14/0 + Q1 + amend |
| pressmark-cpp-solo | solo | 1729.8 | 13382104 (2 segm.) | 5816283 / 32668 (44%) | 5772672 | 44 | 12.25 | green×2 (30 scenarios), deliv. 5/5 | — / 2 cycles |
| pressmark-cpp-top | top | 1187.7 | 4638380 (1 segm.) | 4551083 / 87297 (100%) | 4426880 | 55 | 1.71 | green×2, 22 cases, deliv. 5/5 | 7/0 |
| pressmark-cpp-weak | weak | 2137.5 | 4458393 (2 segm.) | 3750817 / 49464 (85%) | 3689088 | 41 | 0.24 | green×2, 27 cases, deliv. 5/5 | 7/0 + Q1 |
| pressmark-go-solo | solo | 1756.5 | 9618938 (1 segm.) | 9521597 / 97341 (100%) | 9399360 | 127 | 3.04 | green×2 (119 checks), deliv. 5/5 | — / 1 cycle |
| pressmark-go-top | top | 995.5 | 3345792 (1 segm.) | 3268492 / 77300 (100%) | 3162304 | 46 | 1.31 | green×2, 23 cases, deliv. 5/5 | 9/0 |
| pressmark-go-weak | weak | 2103.8 | 2426197 (1 segm.) | 2332812 / 93385 (100%) | 2181376 | 31 | 0.13 | green×2, 22 cases, deliv. 5/5 | 22/0 |
| pressmark-js-solo | solo | 1892.7 | 7733831 (2 segm.) | 1433532 / 6390 (19%) | 1423296 | 13 | 9.22 | green×2 (76 scenarios), deliv. 5/5 | — / 2 cycles |
| pressmark-js-top | top | 900.2 | 2091541 (1 segm.) | 2031871 / 59670 (100%) | 1943360 | 36 | 0.89 | green×2, 31 cases, deliv. 5/5 | 11/0 |
| pressmark-js-weak | weak | 2135.7 | 4103534 (1 segm.) | 4007542 / 95992 (100%) | 3874496 | 53 | 0.18 | green×2, 24 cases, deliv. 5/5 | 10/0 |
| pressmark-python-solo | solo | 1477.3 | 5058192 (2 segm.) | 1427885 / 8535 (28%) | 1415680 | 15 | 5.49 | green×2 (38 scenarios), deliv. 5/5 | — / 2 cycles |
| pressmark-python-top | top | 700.6 | 1083056 (1 segm.) | 1036251 / 46805 (100%) | 943104 | 22 | 0.58 | green×2, 24 cases, deliv. 5/5 | 6/0 |
| pressmark-python-weak | weak | 1333.6 | 2736888 (2 segm.) | 1734723 / 19992 (64%) | 1697280 | 18 | 0.21 | green×2, 24 cases, deliv. 5/5 | 10/0 + Q1 |
| quire-rust-solo | solo | 1502.0 | 4351174 (2 segm.) | 1064798 / 12955 (25%) | 1048960 | 9 | 4.93 | green×2 (65 checks), deliv. 5/5 | — / 2 cycles |
| quire-rust-top | top | 1094.1 | 2320712 (1 segm.) | 2257699 / 63013 (100%) | 2158784 | 47 | 0.98 | green×2, 14 cases, deliv. 5/5 | 6/0 |
| quire-rust-weak | weak | 1584.9 | 2438615 (1 segm.) | 2374091 / 64524 (100%) | 2260416 | 37 | 0.12 | green×2, 14 cases, deliv. 5/5 | 14/0 |
| quire-cpp-solo | solo | 1881.9 | 8708178 (2 segm.) | 3146342 / 12920 (36%) | 3127744 | 24 | 8.66 | green×2 (101 checks), deliv. 5/5 | — / 2 cycles |
| quire-cpp-top | top | 1125.8 | 3220137 (1 segm.) | 3147794 / 72343 (100%) | 3035648 | 46 | 1.26 | green×2, 14 cases, deliv. 5/5 | 8/0 |
| quire-cpp-weak | weak | 2252.2 | 5529286 (1 segm.) | 5434720 / 94566 (100%) | 5304896 | 62 | 0.23 | green×2, 15 cases, deliv. 5/5 | 7/0 + 1 fix |
| quire-go-solo | solo | 1364.8 | 2785502 (1 segm.) | 2709306 / 76196 (100%) | 2625152 | 47 | 1.14 | green×2 (27 scenarios), deliv. 5/5 | — / 1 cycle |
| quire-go-top | top | 677.2 | 840271 (1 segm.) | 794419 / 45852 (100%) | 722368 | 19 | 0.49 | green×2, 21 cases, deliv. 5/5 | 7/0 |
| quire-go-weak | weak | 1131.8 | 2060858 (1 segm.) | 2004141 / 56717 (100%) | 1914752 | 36 | 0.10 | green×2, 16 cases, deliv. 5/5 | 9/0 |
| quire-js-solo | solo | 1028.0 | 3028079 (2 segm.) | 1361228 / 14620 (45%) | 1341888 | 26 | 2.75 | green×2 (24 scenarios), deliv. 5/5 | — / 2 cycles |
| quire-js-top | top | 816.3 | 1587284 (1 segm.) | 1535075 / 52209 (100%) | 1454592 | 37 | 0.72 | green×2, 15 cases, deliv. 5/5 | 7/0 |
| quire-js-weak | weak | 1437.9 | 1479547 (1 segm.) | 1423806 / 55741 (100%) | 1333056 | 29 | 0.08 | green×2, 13 cases, deliv. 5/5 | 7/0 |
| quire-python-solo | solo | 1188.3 | 3155047 (2 segm.) | 1088949 / 13561 (35%) | 1072384 | 13 | 3.24 | green×2 (38 scenarios), deliv. 5/5 | — / 2 cycles |
| quire-python-top | top | 753.8 | 2333094 (1 segm.) | 2282381 / 50713 (100%) | 2198016 | 42 | 0.91 | green×2, 16 cases, deliv. 5/5 | 10/0 + 1 fix |
| quire-python-weak | weak | 872.2 | 883001 (1 segm.) | 843130 / 39871 (100%) | 775744 | 21 | 0.05 | green×2, 14 cases, deliv. 5/5 | 5/0 |

The hung first hand of conduit-cpp-weak (the segment before the replacement)
issued no usage — the platform never completed the turn; its rollout snapshot
is in the cell's evidence; the row above is the replacement hand, and the
reconciliation note of §4 covers the same wave's conduit-go-weak correction.

### Operators

| operator | time, s (wall = all segments) | life, tokens | split in/out (coverage) | cache | tools | $ | decisions |
|---|---|---|---|---|---|---|---|
| conduit-rust-solo | 768.4 | 2987853 | 1448919 / 9747 (49%) | 1420032 | 17 | 2.59 | 2 cycles |
| conduit-cpp-solo | 423.6 | 1021877 | 1003190 / 18687 (100%) | 950784 | 43 | 0.40 | 1 cycle |
| conduit-go-solo | 850.7 | 1828141 | 589150 / 7314 (33%) | 576704 | 8 | 1.92 | 2 cycles |
| conduit-js-solo | 587.9 | 1026589 | 1004059 / 22530 (100%) | 950272 | 28 | 0.42 | 1 cycle |
| conduit-python-solo | 746.9 | 2162480 | 1205671 / 11896 (56%) | 1086080 | 12 | 1.83 | 2 cycles |
| errand-rust-solo(v) | 31.7 | 54081 | 52375 / 1706 (100%) | 49216 | 3 | 0.02 | a stack decision |
| errand-rust-solo(acc) | 286.4 | 910399 | 896851 / 13548 (100%) | 855872 | 31 | 0.34 | acceptance, 1 cycle |
| errand-cpp-solo | 569.6 | 2110119 | 1247335 / 13782 (60%) | 1223616 | 20 | 1.60 | 2 cycles |
| errand-go-solo | 631.6 | 1318945 | 562259 / 6306 (43%) | 547776 | 12 | 1.24 | 2 cycles |
| errand-js-solo | 258.3 | 512389 (1 segm.) | 498277 / 14112 (100%) | 461248 | 19 | 0.23 | acceptance, 1 cycle |
| conduit-go-weak | 31.7 | 40799 | 39445 / 1354 (100%) | 35776 | 2 | 0.02 | 1 decision, wait 31.7 s |
| errand-rust-top | 239.1 | 977100 | 968377 / 8723 (100%) | 935040 | 34 | 0.33 | 1 decision (filed itself — findings) |
| errand-rust-weak | 29.9 | 39049 | 37641 / 1408 (100%) | 35840 | 2 | 0.02 | 1 decision, wait 30.0 s |
| pressmark-rust-weak | 59.7 | 109824 (2 agents) | 106888 / 2936 (100%) | 100992 | 6 | 0.05 | 2 decisions (Q1 wait 59 s + amend approve) |
| pressmark-cpp-weak | 16.0 | 39495 (1 agent) | 38623 / 872 (100%) | 36224 | 2 | 0.02 | 1 decision (Q1, wait 44 s) |
| pressmark-python-weak | 24.1 | 53502 (1 agent) | 52333 / 1169 (100%) | 49728 | 3 | 0.02 | 1 decision (Q1) |
| quire-rust-solo | 391.9 | 1283340 (2 segm.) | 546513 / 6852 (43%) | 531968 | 13 | 0.19 | acceptance, 2 cycles |
| quire-cpp-solo | 651.1 | 1932052 (2 segm.) | 1022767 / 9687 (53%) | 990144 | 13 | 0.35 | acceptance, 2 cycles |
| quire-go-solo | 421.6 | 1005211 (1 segm.) | 983841 / 21370 (100%) | 936704 | 27 | 0.40 | acceptance, 1 cycle |
| quire-js-solo | 449.2 | 1124157 (2 segm.) | 540371 / 6721 (49%) | 523840 | 11 | 0.19 | acceptance, 2 cycles |
| quire-python-solo | 305.7 | 834942 (2 segm.) | 372606 / 6813 (45%) | 358528 | 10 | 0.14 | acceptance, 2 cycles |
| pressmark-rust-solo | 382.7 | 2406762 (1 agent) | 2388151 / 18611 (100%) | 2291264 | 44 | 0.81 | acceptance, 1 cycle |
| pressmark-cpp-solo | 446.4 | 1747647 (1 agent) | 1043775 / 7561 (60%) | 1022592 | 13 | 0.33 | acceptance, 2 cycles |
| pressmark-go-solo | 429.7 | 1411906 (1 agent) | 1385418 / 26488 (100%) | 1315072 | 33 | 0.56 | acceptance, 1 cycle |
| pressmark-js-solo | 516.3 | 1985425 (1 agent) | 667235 / 4332 (34%) | 658496 | 8 | 0.20 | acceptance, 2 cycles |
| pressmark-python-solo | 502.2 | 2098064 (1 agent) | 845515 / 6185 (41%) | 831040 | 12 | 0.26 | acceptance, 2 cycles |
| errand-js-top | 15.9 | 37812 (1 agent) | 37191 / 621 (100%) | 35520 | 2 | 0.01 | 1 decision (QC) |
| errand-python-solo | 936.9 | 2286054 (2 agents) | 1496305 / 23859 (66%) | 1438144 | 39 | 0.56 | acceptance, 2 cycles |

((v) — the stack-decision agent, (acc) — the acceptance agent)

### Layers (judges, measurers, smokes)

| layer | time, s (wall = all segments) | life, tokens (DB) | split in/out (coverage) | cache | tools | $ |
|---|---|---|---|---|---|---|
| judge/conduit-cpp | 655.2 | 928283 | 900199 / 28084 (100%) | 844608 | 32 | 0.42 |
| judge/conduit-go | 516.9 | 1194513 | 1176237 / 18276 (100%) | 1135552 | 40 | 0.43 |
| judge/conduit-python | 358.3 | 739734 | 721433 / 18301 (100%) | 680128 | 31 | 0.32 |
| judge/conduit-rust | 309.7 | 771618 | 755636 / 15982 (100%) | 708672 | 27 | 0.32 |
| judge/errand-rust | 381.6 | 967963 | 946382 / 21581 (100%) | 901504 | 44 | 0.39 |
| measurer/conduit-cpp | 760.4 | 1468100 | 1447824 / 20276 (100%) | 1362432 | 53 | 0.56 |
| measurer/conduit-go | 567.3 | 1745280 | 1723573 / 21707 (100%) | 1644416 | 39 | 0.63 |
| measurer/conduit-js | 431.8 | 1732993 | 1714470 / 18523 (100%) | 1651392 | 43 | 0.60 |
| measurer/conduit-python | 353.6 | 1731155 | 1711625 / 19530 (100%) | 1648320 | 39 | 0.60 |
| measurer/conduit-rust | 495.8 | 2045901 | 2020982 / 24919 (100%) | 1953344 | 54 | 0.71 |
| measurer/errand-rust | 443.0 | 1564411 | 1538983 / 25428 (100%) | 1472512 | 51 | 0.59 |
| smoke/en | 370.0 | 644961 | 625651 / 19310 (100%) | 584704 | 19 | 0.29 |
| smoke/ru | 160.3 | 439043 | 431188 / 7855 (100%) | 402496 | 13 | 0.18 |
| measurer/errand-go | 551.8 | 1351791 (1 segm.) | 1330188 / 21603 (100%) | 1282880 | 48 | 0.49 |
| measurer/errand-js | 433.2 | 1242863 (1 segm.) | 1222317 / 20546 (100%) | 1161792 | 42 | 0.48 |
| measurer/errand-python | 360.3 | 848734 (1 segm.) | 832784 / 15950 (100%) | 786240 | 39 | 0.34 |
| judge/errand-js | 546.4 | 1605399 (1 segm.) | 1568218 / 37181 (100%) | 1500160 | 44 | 0.65 |
| judge/errand-go (1st run) | 581.3 | 1117669 (1 segm.) | 1093085 / 24584 (100%) | 1050560 | 40 | 0.44 |
| judge/errand-python | 408.7 | 791339 (1 segm.) | 770538 / 20801 (100%) | 732352 | 29 | 0.34 |
| measurer/pressmark-cpp | 357.9 | 999074 (1 segm.) | 984873 / 14201 (100%) | 948352 | 40 | 0.36 |
| measurer/pressmark-rust | 316.3 | 1497704 (1 segm.) | 1478078 / 19626 (100%) | 1420032 | 45 | 0.54 |
| measurer/pressmark-go | 333.0 | 1150113 (1 segm.) | 1133523 / 16590 (100%) | 1087488 | 38 | 0.42 |
| measurer/pressmark-js | 523.0 | 1921241 (1 segm.) | 1899701 / 21540 (100%) | 1843584 | 53 | 0.65 |
| judge/pressmark-cpp | 456.9 | 1400010 (1 segm.) | 1376323 / 23687 (100%) | 1325504 | 37 | 0.52 |
| judge/pressmark-rust | 377.7 | 1029465 (1 segm.) | 1009186 / 20279 (100%) | 964160 | 37 | 0.40 |
| judge/pressmark-go | 458.0 | 2207952 (1 segm.) | 2182179 / 25773 (100%) | 2108608 | 41 | 0.76 |
| judge/pressmark-js | 438.5 | 1370197 (1 segm.) | 1342365 / 27832 (100%) | 1266112 | 30 | 0.56 |
| judge/pressmark-python | 443.4 | 1671543 (1 segm.) | 1644568 / 26975 (100%) | 1573440 | 38 | 0.63 |
| measurer/quire-rust | 169.8 | 623491 (1 segm.) | 613014 / 10477 (100%) | 578432 | 25 | 0.24 |
| measurer/quire-cpp | 738.0 | 3185244 (1 segm.) | 3152615 / 32629 (100%) | 3090304 | 74 | 1.03 |
| measurer/quire-go | 377.4 | 1080649 (1 segm.) | 1062819 / 17830 (100%) | 1020352 | 36 | 0.40 |
| measurer/quire-js | 292.0 | 907740 (1 segm.) | 891946 / 15794 (100%) | 851776 | 38 | 0.35 |
| measurer/quire-python | 319.3 | 1268480 (1 segm.) | 1251622 / 16858 (100%) | 1203712 | 37 | 0.45 |
| judge/quire-rust | 306.9 | 925985 (1 segm.) | 909677 / 16308 (100%) | 875648 | 34 | 0.35 |
| judge/quire-cpp | 429.9 | 1024880 (1 segm.) | 1003214 / 21666 (100%) | 957056 | 33 | 0.41 |
| judge/quire-go | 465.8 | 1183162 (1 segm.) | 1156861 / 26301 (100%) | 1111616 | 35 | 0.47 |
| judge/quire-js | 570.0 | 1255641 (1 segm.) | 1219098 / 36543 (100%) | 1154880 | 31 | 0.55 |
| measurer/errand-cpp | 352.5 | 1706888 (1 segm.) | 1685728 / 21160 (100%) | 1632576 | 45 | 0.59 |
| judge/errand-go (debt re-run) | 528.0 | 1145796 (1 segm.) | 1120241 / 25555 (100%) | 1071040 | 37 | 0.46 |
| judge/errand-cpp | 355.0 | 1153443 (1 segm.) | 1129724 / 23719 (100%) | 1080960 | 37 | 0.45 |
| judge/conduit-js | 1145.6 | 4239917 (1 segm.) | 4187723 / 52194 (100%) | 4088448 | 68 | 1.43 |

Two judge runs are recorded for errand-go: the first (20:08–20:21Z) and the
debt re-run (23:38–23:46Z, both verdicts identical 0/9 + 0/9); both are kept
as separate ledger rows and never summed into one $.

## 5. Wave S — completeness update (2026-09-03 23:1xZ, historical)

The tables were cut at this moment with all hands of conduit, errand and
pressmark entered (quire was still running); the operators were re-bound per
cell (the sum of all operator agents of a hand: code 3/4 decisions for
machine hands, acceptance cycles for solo) — the "full path = hand + its
operator" economics reads per cell directly. The ledger was re-extracted over
the 107 wave agents known at that moment (the closed wave's ledger holds
149); determinism was reconciled with the closed cells: one divergence —
pressmark-python-weak gained a second segment after the first extraction;
the splits of old agents were restored from the committed ledger where the
platform had trimmed the metadata. The layer debts of that moment (measure
errand-cpp ×3, judge errand-cpp, judge errand-go, judge conduit-js) were
re-run by the same pattern before the wave closed — the re-run rows are in
§4. An unidentified operator agent at 15:34Z (life 1,021,877) was attributed
via the journal to conduit-cpp-solo in phase A.

## 6. WAVE S CLOSED (75/75 hands of the campaign)

All five senses × five stacks × three hands are closed: 60 cells in wave S
(conduit, errand, pressmark, quire) + 15 brim cells of wave 1 = 75 hands of
the campaign. Blind measurer: suite ×2 green and deliverables 5/5 for all 75
products (the `queued` placeholders of the historical cut above are replaced
by the closed values in §4). Judges: closed for 21 of the 25 tasks — all 20
wave-S tasks (40 products: solo and weak against the top benchmark) plus the
brim-rust pilot (2 products + a top self-check 21/21); the other brim tasks
carried no judges in wave 1. The 13 re-run judges carry the calibration
3/3 + the Reported frame. Ledger: 149 wave agents, determinism reconciled.
Canaries: clean in all 75 cells. Transcript scan: three red lines
(pressmark-rust-weak host garbage; quire-rust-weak a single ls of the repo
root; quire-python-weak reading its own platform exec log) — all documented
in the cells' findings.

## 7. Wave S — summary by hands (20 non-brim tasks × 3 hands, means per cell)

| metric | solo | top+machine | flash+machine |
|---|---|---|---|
| cells (n) | 20 | 20 | 20 |
| hand $ (cache-aware, mean) | 4.88 | 1.16 | 0.31 |
| hand $ (median / min..max) | 3.68 / 0.80..12.60 | 1.00 / 0.49..2.57 | 0.19 / 0.05..2.14 |
| operator $ per cell | 1.19 | 0.02 | 0.01 |
| full path $ (hand+operator) | 6.07 | 1.18 | 0.32 |
| hand time, s (wall = all segments, mean / median) | 1556 / 1488 | 1131 / 1093 | 2258 / 2000 |
| hand life-tokens | 5.9M | 2.8M | 4.5M |
| input / output (mean) | 3.2M / 0.04M | 2.6M / 0.06M | 3.4M / 0.06M |
| split coverage | 58% | 99% | 89% |
| tool calls | 40 | 38 | 40 |
| submissions (machine, accepted/rejected) | 1–2 acceptance cycles (7 of 20 on the first) | 5–22 / 0–2 | 5–22 / 0–2 |
| cells with a question batch / escalation | — | 2 of 20 (0.10) | 5 of 20 (0.25, waits 16–59 s) |
| suite ×2 green (blind measurer) | 20/20 | 20/20 | 20/20 |
| deliverables 5/5 | 20/20 | 20/20 | 20/20 |
| VERDICT READY / ACCEPTED | 20/20 accepted | 20/20 READY | 20/20 READY |
| judge: pass rows against the top benchmark | 87/336 (26%) | (benchmark) | 105/336 (31%) |
| host-boundary red lines | 0 | 0 | 3 |

Judge rows: the sum over all 20 wave-S tasks' judge files (solo and weak
products against the top hand's exported table, byte-strict; last run per
file). Campaign-wide incl. the brim-rust pilot: solo 94/357, weak 112/357.

Hand time is the full wall of the hand — the sum of all its platform
segments (§1.2); for the machine hands this includes the machine's own suite
runs (every check runs twice in clean directories). On the full wall the top
hand is the fastest of the three and solo the token-heaviest: the earlier
last-segment view understated solo (mean 684 s) — the wall shows 1556 s. The
campaign-wide question-batch count adds two wave-1 top cells (brim-cpp-top
79 s, brim-go-top 102 s): 4 of the 25 top cells carried a batch.

## 8. Conclusion: what the machine buys (economics)

1. FULL-PATH SAVINGS. The mean solo cell costs 6.07 dollars, of which 1.19 is
   the customer's work (verifying reading, setups, acceptance cycles). The
   same task through the machine: top 1.18 dollars (5.1× cheaper), flash 0.32
   (19×). The machine hands' operator fires rarely (7 of 40 machine cells in
   wave S — 2 top, 5 weak; 0 of 10 in wave 1) and costs cents (micro-decisions
   with decision waits of 16–59 s in wave S; the two wave-1 top answers waited
   79/102 s), because the machine took
   over the verification: every check runs twice in clean directories with
   byte-exact comparison, and invariant violations are caught by the gates at
   submission.
2. QUALITY DID NOT SUFFER. By the mechanical standard (brief + expectations
   list, blind measurer) all 60 wave S products are identical: the suite twice
   green and five deliverables for each of the three hands of every task.
   Cheaper does not mean worse: the measurer sees no difference in result
   between the machine hands and solo.
3. THE PRODUCT OF THE EFFECT IS NOT TOKEN PRICE BUT LABOR DISPLACEMENT.
   The top+machine hand spends 2.8M tokens against solo's 5.9M: the agent does
   not re-read the product hunting its own errors (the machine already
   checked), does not write the suite by hand (the case table is executed by
   the machine), does not argue facts with the customer (the gate shows
   measured bytes). Cache coverage 99% against 58% — the machine's iterations
   inside one context, not new passes.
4. THE PRICE OF FLASH. The flash hand is 3.75× cheaper than top at the same
   verdict and the same measurer result; its hidden price is three of the
   three host-boundary violations of the wave (a weaker model breaks the
   environment mechanics more often) and larger self-repairs. For wave S-class
   tasks flash+machine is the cheapest path to the same quality, with an
   environment supervision caveat.
5. WHAT THE MACHINE DID NOT BUY. The "claim ≠ fact" classes and brief-reading
   forks (refusal wordings, the order of equals, the path base, library-vs-CLI)
   are still caught only by a human (the acceptance operator) or cross-comparison
   (the judge): 13 of 20 wave-S solo acceptances (16 of 25 campaign-wide) went
   to a second cycle with real findings; every cross-judged product of the
   campaign (40 wave-S + the brim-rust pilot + 10 validation-wave products)
   fails judge rows exactly on these classes. Verifying reading is an expense
   solo cannot compress; the machine makes it unnecessary for its own path but
   does not replace it for someone else's.

## 9. The validation wave — a run on the frozen candidate (2026-09-03)

Composition (the owner's word): all five senses × two stacks × 2 machine
hands (top GLM-5.3 Max + flash GLM-5.3 Flash High) = 20 cells, 10 tasks; the
arbiter is the frozen candidate (sha 07c77698…); 10 judges, 10 measurers. The
full wave journal is in the run archive; the 48-agent ledger is reproducible
by the extractor `campaigns/campaign-03/tools/usage-ledger.py` from the
platform database.

### Solo rows: carried over from wave S (the transfer rule observed)

All 10 solo rows are carried over with the "transfer, wave S" mark. Identity
proved: each task's brief sha-matched wave S (verified at the staging of every
batch), the image digest is the same (2602e57c3140…), the prompts and prices
did not change since wave S (the sha256 sums are recorded in the run archive).
The solo metrics live in the wave S cells' evidence.

### Hands (mean per cell; $ cache-aware over the cell captures; time = the
captured wall, restarts included)

| hand | $ per cell | time, s per cell | machine verdict |
|---|---|---|---|
| top | 1.30 | 1244 | 10/10 VERDICT READY |
| flash | 0.17 | 1947 | 10/10 VERDICT READY |

Wave S for comparison (the summary of §7): top hand mean $1.16 (full path
$1.18), flash hand mean $0.31 (full path $0.32) — the economics is within the
spread (flash cheaper; the heavy outlier errand-rust-top $2.78 — redis
integrations, expected). Full path = hand + operator: the operators fired in
7 cells of 20 (code 3/4 decisions; one first operator run cancelled without
the customer-ruling context and superseded), 431,614 tokens in total —
$0.60 at the Max input rate (an upper estimate; the cache-aware ledger cost
is lower), decision waits 23–216 s.

### Layers (methodology overheads, not part of the product path)

| layer | agents | $ (cache-aware, ledger) |
|---|---|---|
| measurers | 10 | part of 7.20 |
| judges | 10 | part of 7.20 |
| total layers | 20 | 7.20 |

Wave total: hands 14.74 + operators 0.60 (input-rate upper estimate) +
layers 7.20 = $22.54.

### Blind measurer (the mechanical standard)

16/20 cells — suite ×2 green, deliverables 5/5, coverage 100%. 4 cells
(errand-python top/flash, conduit-js top/flash) — 4/5 (double_green=0): the
runs caught STRING nondeterminism under the blind procedure — cross-case
interference of port 6379 (errand-python) and timing teardown flakes on
network rows (conduit-js); single runs of those rows and second runs of the
tables are green. The class is an artifact of the measuring procedure (a
container per run, not per case) with machine-independent behavior: the
machine acceptance on those cells is green ×2. A new field finding.

### Judges (the flash product against the top benchmark)

| task | score | failure class (all one family) |
|---|---|---|
| quire-rust | 4/13 | refusal wordings, a reading fork |
| brim-cpp | 8/18 | wordings ×9, --threshold semantics (both hands honestly measured their own) |
| pressmark-cpp | 6/20 | wordings/prefixes, an rc fork, the index-moment fork (top pinned empty without asking; flash — by the operator's word "the newest moment") |
| errand-rust | 0/13 | CLI grammar (redis-url-first) |
| conduit-go | 0/20 | artifact path bin/* vs ./* |
| pressmark-go | 0/20 | artifact path bin/pressmark vs the bare name in the benchmark rows |
| errand-python | 0/10 | enqueue/put subcommands, --at/positional |
| quire-js | 0/15 | the wrapper looks for quire.mjs at the root, the artifact in dist/ |
| conduit-js | 0/18 | the conduit.mjs row material, log format |
| brim-python | 0/21 | wordings ×12, the success trailing \n |

The functional rows (builds, copies, determinism, happy paths) are
byte-for-byte for most products; the zero scores are the mass case of
"surface/artifact path" (7 of 10 tasks): the top benchmark rows call ITS
declared surface, the flash product declared its own — the judge makes no
adaptations (honestly). This is a class refinement for brief/conventions
fixes (the decision is the owner's word).

### Wave criteria

- Zero NEW machine defect classes: confirmed — 20/20 VERDICT READY, zero
  panics, zero deadlocks, zero quarantines, zero lost cells; all findings are
  methodology/environment classes (a host-garbage relapse, blind-suite
  nondeterminism, a material×stack conflict, surface forks) — none is a
  machine defect.
- The fixed does not return: zero lint rejections in this wave's field (the
  channel did not fire naturally), the class is held by the pre-freeze live
  contrast (the candidate answers with a line+skeleton, the former one
  panicked) + canon-mirror in the gate stack mechanically; the gate green on
  every tree of the wave; zero non-UTF-8 surfaces in 20 exports (a side
  effect: brim's byte pins of pictures rendered cleanly); --init gave zero
  zombie incidents, the containers lived within the limits; zero journal
  breaks.
- The economics is within the wave S spread: top 1.30 against 1.16, flash 0.17
  against 0.31 (hand means) — yes.

### Comparison with the previous run of the same tasks (full record)

What is compared: the 10 tasks of the validation wave (binary 07c77698…)
against their previous run — 8 tasks of wave S (binary f40013f8…), brim-cpp
and brim-python of wave 1 (binary e19bb7a5…). The conditions are identical
and proved at the solo transfer (brief shas, image digest, prompts/prices —
recorded as sums); the difference is the machine candidate and the re-run of
the machine hands (solo was not re-run).

Bases (§1): the "was" columns are the wave-S **uniform tier** (T3 ledger:
life tokens; full-wall time — the sum of all segment durations; cache-aware
$) — for the two brim rows the wave-1 capture (in+out); the "wave" columns
are the validation wave's **cell captures** (T2: in+out tokens, captured
wall incl. restarts, cache-aware $).
`†` — the wave-S capture was lost or reconciled, the ledger value is used
(quire-rust-top, pressmark-go top/flash, quire-js-top: capture lost;
conduit-go-weak: the single-segment bookkeeping error, §4; quire-rust-weak:
a partial capture superseded by the ledger). The former no-split upper
estimates (*) of errand-rust-top and errand-js-top are replaced by their
exact ledger values.

#### Table 1 — money / time / tokens per cell: previous run → wave

| task | hand | was $ | was s | was tok (life) | wave $ | wave s | wave tok (in+out) | Δ$ |
|---|---|---|---|---|---|---|---|---|
| errand-rust | top | 2.57 | 1386.1 | 6,057,541 | 2.78 | 2584.6 | 8,568,640 | +0.21 |
| errand-rust | weak | 0.20 | 1815.6 | 3,643,692 | 0.07 | 1239.1 | 1,714,546 | −0.13 |
| quire-rust | top | 0.98 † | 1094.1 | 2,320,712 | 1.42 | 1198.9 | 4,124,665 | +0.44 |
| quire-rust | weak | 0.12 | 1584.9 | 2,438,615 | 0.09 | 1879.2 | 1,991,581 | −0.03 |
| pressmark-cpp | top | 1.71 | 1187.7 | 4,638,380 | 1.66 | 1326.7 | 4,162,966 | −0.05 |
| pressmark-cpp | weak | 0.24 | 2137.5 | 4,458,393 | 0.08 | 2289.3 | 1,764,456 | −0.16 |
| brim-cpp | top | 0.73 | 638.0 | 1,980,365 | 1.29 | 1047.2 | 3,534,844 | +0.56 |
| brim-cpp | weak | 0.02 | 135.7 | 469,220 | 0.12 | 1613.9 | 2,693,469 | +0.10 |
| conduit-go | top | 1.60 | 1758.3 | 4,290,386 | 1.10 | 1231.1 | 3,252,464 | −0.50 |
| conduit-go | weak | 2.14 | 7866.0 | 17,904,719 | 0.14 | 1853.8 | 2,777,242 | −2.00 |
| pressmark-go | top | 1.31 † | 995.5 | 3,345,792 | 1.23 | 1242.0 | 2,667,712 | −0.08 |
| pressmark-go | weak | 0.13 † | 2103.8 | 2,426,197 | 0.14 | 1698.5 | 3,406,402 | +0.01 |
| conduit-js | top | 0.98 | 1737.7 | 1,926,944 | 1.15 | 1480.6 | 2,619,710 | +0.17 |
| conduit-js | weak | 0.15 | 1986.6 | 3,279,696 | 0.79 | 4287.1 | 8,224,739 | +0.64 |
| quire-js | top | 0.72 † | 816.3 | 1,587,284 | 0.94 | 955.7 | 2,439,113 | +0.22 |
| quire-js | weak | 0.08 | 1437.9 | 1,479,547 | 0.13 | 1871.8 | 2,765,424 | +0.05 |
| errand-python | top | 0.59 | 707.8 | 1,278,392 | 0.86 | 701.5 | 2,310,564 | +0.27 |
| errand-python | weak | 0.21 | 2224.0 | 5,170,032 | 0.10 | 1438.6 | 2,192,105 | −0.11 |
| brim-python | top | 0.28 | 224.5 | 737,793 | 0.59 | 669.3 | 1,261,701 | +0.31 |
| brim-python | weak | 0.05 | 592.8 | 1,045,104 | 0.06 | 1298.3 | 995,519 | +0.01 |
| **TOTAL top** | | **11.47** | | | **13.02** | | | **+1.55** |
| **TOTAL weak** | | **3.34** | | | **1.72** | | | **−1.62** |

#### Table 2 — time and token deltas (all 10 pairs per hand)

| task | hand | was s | wave s | Δs | was tok | wave tok | Δtok | Δtok % |
|---|---|---|---|---|---|---|---|---|
| errand-rust | top | 1386.1 | 2584.6 | +1198 | 6,057,541 | 8,568,640 | +2,511,099 | +41.5% |
| errand-rust | weak | 1815.6 | 1239.1 | −576 | 3,643,692 | 1,714,546 | −1,929,146 | −52.9% |
| quire-rust | top | 1094.1 | 1198.9 | +105 | 2,320,712 | 4,124,665 | +1,803,953 | +77.7% |
| quire-rust | weak | 1584.9 | 1879.2 | +294 | 2,438,615 | 1,991,581 | −447,034 | −18.3% |
| pressmark-cpp | top | 1187.7 | 1326.7 | +139 | 4,638,380 | 4,162,966 | −475,414 | −10.2% |
| pressmark-cpp | weak | 2137.5 | 2289.3 | +152 | 4,458,393 | 1,764,456 | −2,693,937 | −60.4% |
| brim-cpp | top | 638.0 | 1047.2 | +409 | 1,980,365 | 3,534,844 | +1,554,479 | +78.5% |
| brim-cpp | weak | 135.7 | 1613.9 | +1478 | 469,220 | 2,693,469 | +2,224,249 | +474.0% |
| conduit-go | top | 1758.3 | 1231.1 | −527 | 4,290,386 | 3,252,464 | −1,037,922 | −24.2% |
| conduit-go | weak | 7866.0 | 1853.8 | −6012 | 17,904,719 | 2,777,242 | −15,127,477 | −84.5% |
| pressmark-go | top | 995.5 | 1242.0 | +246 | 3,345,792 | 2,667,712 | −678,080 | −20.3% |
| pressmark-go | weak | 2103.8 | 1698.5 | −405 | 2,426,197 | 3,406,402 | +980,205 | +40.4% |
| conduit-js | top | 1737.7 | 1480.6 | −257 | 1,926,944 | 2,619,710 | +692,766 | +36.0% |
| conduit-js | weak | 1986.6 | 4287.1 | +2301 | 3,279,696 | 8,224,739 | +4,945,043 | +150.8% |
| quire-js | top | 816.3 | 955.7 | +139 | 1,587,284 | 2,439,113 | +851,829 | +53.7% |
| quire-js | weak | 1437.9 | 1871.8 | +434 | 1,479,547 | 2,765,424 | +1,285,877 | +86.9% |
| errand-python | top | 707.8 | 701.5 | −6 | 1,278,392 | 2,310,564 | +1,032,172 | +80.7% |
| errand-python | weak | 2224.0 | 1438.6 | −785 | 5,170,032 | 2,192,105 | −2,977,927 | −57.6% |
| brim-python | top | 224.5 | 669.3 | +445 | 737,793 | 1,261,701 | +523,908 | +71.0% |
| brim-python | weak | 592.8 | 1298.3 | +706 | 1,045,104 | 995,519 | −49,585 | −4.7% |
| **TOTAL top** | | **10,546** | **12,438** | **+1892** | **28,163,589** | **34,942,379** | **+6,778,790** | **+24.1%** |
| **TOTAL weak** | | **21,885** | **19,470** | **−2415** | **42,315,215** | **28,525,483** | **−13,789,732** | **−32.6%** |

Reading: Δ = wave − was ("+" the wave is more). The top hand's token growth
(+24% over the 10 pairs; wall time +1892 s, +18%) is the payment for fuller
case tables; it did NOT carry into price — ~97% of the wave's input tokens
are cache reads ($0.26/M against $1.4/M), and the totals moved by +1.55
dollars over 10 cells. The weak hand's −33% tokens (wall −2415 s) is
dominated by the reconciled conduit-go-weak cell (its previous-run record
was inflated by that wave's platform breaks and the single-segment
bookkeeping error, §4). The bases of the two sides are stated above the
tables; where the previous run had only upper estimates the exact ledger
values are now used.

#### Table 3 — the full path: every path = hand + its operator (solo's operator is the acceptance operator)

What this table compares is the full price of reaching "done" on one task —
the executor's work and the verification work that necessarily accompanies
it. The operator column is the human-attention proxy: the scarce resource
the machine buys back is not the hand's tokens but this column. Roles are
not mixed: the hand's and the operator's time/tokens are separate columns,
the sum is analytical. Machine hands — the validation wave's cell
captures (T2, the full wall incl. restarts). The whole solo side (hand and
operator) — the wave-S uniform tier (T3: full-wall times, life tokens; the
operators are the §4 operators-table rows, errand-rust = its two agents
summed); brim rows — the wave-1 captures (T2).

| task | path | hand s | op s | Σ s | hand tok | op tok | Σ tok |
|---|---|---|---|---|---|---|---|
| errand-rust | SOLO+OP | 1056 | 318 | 1374 | 4,130,645 | 964,480 | 5,095,125 |
| errand-rust | TOP+OP | 2585 | 20 | 2605 | 8,568,640 | 26,536 | 8,595,176 |
| errand-rust | FLASH+OP | 1239 | 88 | 1327 | 1,714,546 | 101,728 | 1,816,274 |
| quire-rust | SOLO+OP | 1502 | 392 | 1894 | 4,351,174 | 1,283,340 | 5,634,514 |
| quire-rust | TOP+OP | 1199 | 0 | 1199 | 4,124,665 | 0 | 4,124,665 |
| quire-rust | FLASH+OP | 1879 | 26 | 1905 | 1,991,581 | 53,304 | 2,044,885 |
| pressmark-cpp | SOLO+OP | 1730 | 446 | 2176 | 13,382,104 | 1,747,647 | 15,129,751 |
| pressmark-cpp | TOP+OP | 1327 | 0 | 1327 | 4,162,966 | 0 | 4,162,966 |
| pressmark-cpp | FLASH+OP | 2289 | 20 | 2309 | 1,764,456 | 52,896 | 1,817,352 |
| brim-cpp | SOLO+OP | 1174 | 352 | 1526 | 2,645,371 | 1,192,897 | 3,838,268 |
| brim-cpp | TOP+OP | 1047 | 0 | 1047 | 3,534,844 | 0 | 3,534,844 |
| brim-cpp | FLASH+OP | 1614 | 0 | 1614 | 2,693,469 | 0 | 2,693,469 |
| conduit-go | SOLO+OP | 2894 | 851 | 3745 | 10,835,440 | 1,828,141 | 12,663,581 |
| conduit-go | TOP+OP | 1231 | 23 | 1254 | 3,252,464 | 53,672 | 3,306,136 |
| conduit-go | FLASH+OP | 1854 | 0 | 1854 | 2,777,242 | 0 | 2,777,242 |
| pressmark-go | SOLO+OP | 1756 | 430 | 2186 | 9,618,938 | 1,411,906 | 11,030,844 |
| pressmark-go | TOP+OP | 1242 | 0 | 1242 | 2,667,712 | 0 | 2,667,712 |
| pressmark-go | FLASH+OP | 1698 | 29 | 1727 | 3,406,402 | 56,275 | 3,462,677 |
| conduit-js | SOLO+OP | 1842 | 588 | 2430 | 5,988,886 | 1,026,589 | 7,015,475 |
| conduit-js | TOP+OP | 1481 | 0 | 1481 | 2,619,710 | 0 | 2,619,710 |
| conduit-js | FLASH+OP | 4287 | 48 | 4335 | 8,224,739 | 87,203 | 8,311,942 |
| quire-js | SOLO+OP | 1028 | 449 | 1477 | 3,028,079 | 1,124,157 | 4,152,236 |
| quire-js | TOP+OP | 956 | 0 | 956 | 2,439,113 | 0 | 2,439,113 |
| quire-js | FLASH+OP | 1872 | 0 | 1872 | 2,765,424 | 0 | 2,765,424 |
| errand-python | SOLO+OP | 533 | 937 | 1470 | 1,260,075 | 2,286,054 | 3,546,129 |
| errand-python | TOP+OP | 702 | 0 | 702 | 2,310,564 | 0 | 2,310,564 |
| errand-python | FLASH+OP | 1439 | 0 | 1439 | 2,192,105 | 0 | 2,192,105 |
| brim-python | SOLO+OP | 729 | 273 | 1002 | 3,386,686 | 460,723 | 3,847,409 |
| brim-python | TOP+OP | 669 | 0 | 669 | 1,261,701 | 0 | 1,261,701 |
| brim-python | FLASH+OP | 1298 | 0 | 1298 | 995,519 | 0 | 995,519 |

Reading: for solo the acceptance operator is a full role (273–937 s over
the ten tasks, up to 2.3M tokens on errand-python and 1.8M on conduit-go),
and the solo HAND is the token-heaviest of the three paths (its mean life
is 5.95M tokens against the top hand's 2.78M — without a machine that has
already verified, the agent re-reads and re-probes); on the wave's machine
paths the operator barely works (0–88 s, fired in 7 cells of 20) — the
machine took the verification.
On the full path the machine top beats solo on 9 of the 10 tasks (the
exception is errand-rust, the redis-integration outlier); zero op for a
machine hand = no questions/escalations. Caveats: the wave S journal shows
the solo acceptance ran 1–2 cycles per task everywhere; zero op for a
machine hand is real — no question batch fired; the hands' decision waits
(23–216 s in 7 cells) are not in elapsed — counted separately in each
cell's platform capture. The earlier
revision of this table printed the solo hand's last-segment time and partial
captures, which made solo look lean; the full wall and the life tokens
correct that.

#### The operator column under a live human (the correction)

In every wave of this registry the operator role was executed by an AI agent
(GLM-5.3 Max / GLM-5.3 — the role tables of the wave journals; one platform
source for every role, §1.1). The operator seconds recorded above are
therefore an AI's wall, not a person's, and for the human cost of the solo
path they are a **lower bound**, for three reasons:

1. **Speed of analysis.** The AI operator re-reads the entire cell context
   (brief, spec, product, suite output) in seconds and fires its own probe
   runs between readings; a person reads at reading speed, re-runs by hand,
   and re-derives the state of the environment before each judgment.
2. **No setup, no fatigue, no context switch.** The AI's wall contains only
   the verification itself; a human's calendar time adds environment setup
   and the return of attention after an interruption — none of it exists in
   any table of this registry.
3. **Asynchrony is unmeasured.** The only wait this registry records on the
   human side is the hand's decision wait (23–216 s, the machine's side of
   the pause); a live-human cycle costs minutes to hours of wall-clock time
   before the answer arrives.

The measured anchor: the solo acceptance operator's full wall is **530 s
(8.8 min) per task on average** over the 20 non-brim wave-S tasks (range
258–937 s, §4). The live-human anchor for the same class of work — the
earlier campaign's finishing estimate, 4–7.5 hours of a live human per
25-task batch — is **≈10–18 min per task**, i.e. 1.2–2× the AI wall, before
the asynchronous latency of reason 3.

What the machine therefore buys, per 25-task batch of this class: the solo
path needs a verifying human cycle at **every** task (≈4–8 hours by the
human anchor, or ≈3.7 hours at the AI wall — whichever operator stands in);
the machine paths triggered operator involvement in 7 of 40 wave-S cells
and 7 of 20 validation-wave cells, at cents per cell (§7: $0.02/$0.01) and
micro-decisions of seconds — on the remaining cells the human leaves the
loop entirely. At any nonzero human hourly rate the gap only widens: the
machine path's verifiable "done" costs ≈$1.30 + cents with no human cycle,
the solo path costs its hand plus a human acceptance cycle per task.

Honesty marks: the 10–18 min and the derived 4–8 hours are an extrapolation
from the earlier campaign's documented human estimate — a labeled estimate,
not a measurement of these waves; the 530 s mean is the §4 ledger
recomputation; nothing else is extrapolated.

## 10. The live three-language wish smoke on the frozen candidate (2026-09-03)

Composition: three layer agents (ru/en/zh, parallel launch, the session model
GLM-5.3 Max), fresh throwaway directories, one task shape (the deterministic
CLI utility linetally, host shell tools only), the full wish → VERDICT cycle
on the frozen candidate (07c77698…); zh for the first time as a live
instance (the earlier smoke waves carried en/ru only). The full journal and
observations are in the run archive; the ledger is one segment per agent, the
split validated against the platform database (coverage 100%).

| lang | VERDICT | wall, s | question batches | submissions ±/− | calls | life, tokens | split in/out | cache | tools | $ |
|---|---|---|---|---|---|---|---|---|---|---|
| ru | READY | 193 | 0 | 5/0 | 6 | 300,650 | 284,826 / 15,824 | 253,184 | 9 | 0.18 |
| en | READY | 256 | 0 | 5/0 | 6 | 530,416 | 512,418 / 17,998 | 478,464 | 15 | 0.25 |
| zh | READY | 237 | 0 | 5/1 | 6 | 481,268 | 464,735 / 16,533 | 432,512 | 15 | 0.23 |

Smoke total: $0.66 (0.18–0.25 per language), the parallel wall to the last
agent's finish 256 s. All three languages — one continuous run, zero restarts
and pokes (the watcher: only normal rollout rotation); the verdicts were
verified by the driver against each instance's run log, not only by the
agents' reports. Two observation findings (mixed structural-frame
localization under ru/zh; an intermediate VERDICT: NOT-READY against the
self-doc stop rule) — not machine defects, canon candidates by the owner's
word: the smoke journal and the wave's observations.
