# Metrics

*The measured economics of three generations — 0.2.0 → 0.3.0 → 0.4.0;
every number from its era's frozen registry.*
**English** · [Русский](METRICS.ru.md) · [中文](METRICS.zh.md)

## Contents

- [0. Provenance, bases, prices, honesty](#0-provenance-bases-prices-honesty)
- [1. The 0.2.0 metrics (September 2026)](#1-the-020-metrics-september-2026)
- [2. The 0.3.0 metrics — the five-wave field battery (October 2026)](#2-the-030-metrics--the-five-wave-field-battery-october-2026)
- [3. The 0.4.0 metrics — the case-m battery, the M-series (October 2026)](#3-the-040-metrics--the-case-m-battery-the-m-series-october-2026)
- [4. The operator is an LLM: the live-human renormalization](#4-the-operator-is-an-llm-the-live-human-renormalization)
  - [4.1. The record and why it matters](#41-the-record-and-why-it-matters)
  - [4.2. The operator article, era by era, as measured](#42-the-operator-article-era-by-era-as-measured)
  - [4.3. Renormalized to a live human operator](#43-renormalized-to-a-live-human-operator)
  - [4.4. Honesty marks of this section](#44-honesty-marks-of-this-section)

One file of the measured economics of the machine's three generations.
Every number is read from the frozen registries of its era — not one new
measurement here — and every fact carries its interpretation. The full
per-cell tables stay in the era registries; this file is the structured
cross-era view: what was measured, what it means, and
what it means in particular when the "operator" column is renormalized
from an AI to a live human.

## 0. Provenance, bases, prices, honesty

- **0.2.0 (September 2026)** — the September campaign's registry:
  `campaigns/case-0/docs/METRICS-0.2.0.md` (preserved at the v0.3.0
  tag), wave S of
  20 non-brim tasks × 3 hands + the validation wave; prices recorded
  2026-09-03 (`prices.json`): **GLM-5.3 Max** in 1.4 / out 4.4 /
  cached-in 0.26 \$ per M; **GLM-5.3 Flash High** 0.15 / 0.5 / 0.03. Two
  dollar bases everywhere: cache-aware \$ (measured) and the input-rate
  upper estimate (where the cache split is absent).
- **0.3.0 (October 2026)** — the five-wave field battery's registry:
  `campaigns/case-0/docs/METRICS-0.3.0.md` (preserved at the v0.3.0
  tag); all numbers are the five
  frozen wave reports' Σ-rows, each wave revalidated green
  (verify-metrics 306–308 GREEN / 0 red).
- **0.4.0 (October 2026)** — the M-series registry (preserved in the
  repository's history): every number assembled mechanically from the
  five waves' `usage.tsv` and the judge pages, then re-computed by an
  independent validator: **0 discrepancies**; §3 below carries the
  era's numbers in full.
- **Token bases.** `in+out` is the working base everywhere; `total` =
  in+out + cache reads is printed beside it in the 0.3.0 and 0.4.0
  registries. No verdict of any era changes between the bases.
- **Time bases differ by era and are never mixed.** 0.2.0 prints the
  hand's full wall (all platform segments); 0.3.0 prints wall seconds of
  hands and operators; 0.4.0 prints pure request-time sums (`work_ms`)
  — the 0.4.0 wave walls are not comparable across waves (live-subagent
  ceilings varied across waves) and are given only as calendar
  context.
- **The operator is an LLM in every experiment of every era.** The
  model record: **0.2.0 — GLM-5.3 Max in every role** (hands,
  operators, judges, measurers); **0.3.0 and 0.4.0 — GLM-5.3 High in
  every role.** The 0.3.0/0.4.0 ledgers pin the model as "GLM-5.3"; the
  tier is that record. Prices for the High tier were not recorded
  in those campaigns — the 0.3.0/0.4.0 economics therefore stand in
  tokens and time, and every dollar figure for them in §4 is parametric,
  labeled as such.
- **Month reckoning.** The 0.4.0 material is dated by a development
  reckoning — months counted from the start of the 0.4.0 development;
  in that reckoning the M-series is October 2026 (the era registry's
  own wave roots keep their platform timestamps).
- **Honesty marks** (the registry's standing law): a measured number is
  a measured number; a labeled estimate says so; nothing extrapolated
  passes silently. The one extrapolation used below — the live-human
  anchor ×1.2–2.0 of the AI operator's wall — is the Appendix E
  correction of the research paper, carried here with its label.
- **The generations measured here are past releases.** The machine
  has since been rebuilt beyond 0.4.0: the shipped 0.5.0 release
  is a further rebuild of the same inversion (the demand stream,
  the six outcomes, line-range claims, the bare `done: true`
  act), and no economics of that rebuild are claimed here — every
  number below belongs to the generation it names. The 0.5.1
  release is a defect-fix patch of that same 0.5.0 rebuild; no
  economics of it are claimed here either.

## 1. The 0.2.0 metrics (September 2026)

Wave S — 20 non-brim tasks × 3 hands (solo; top+machine; flash+machine),
means per cell (the era registry's §7):

| metric | solo | top+machine | flash+machine |
|---|---|---|---|
| hand \$ (cache-aware, mean) | 4.88 | 1.16 | 0.31 |
| operator \$ per cell | 1.19 | 0.02 | 0.01 |
| **full path \$ (hand+operator)** | **6.07** | **1.18** | **0.32** |
| hand time, s (full wall, mean) | 1556 | 1131 | 2258 |
| hand life-tokens | 5.9M | 2.8M | 4.5M |
| cache split coverage | 58% | 99% | 89% |
| suite ×2 green / deliverables 5/5 | 20/20, 20/20 | 20/20, 20/20 | 20/20, 20/20 |
| verdicts | 20/20 ACCEPTED | 20/20 READY | 20/20 READY |
| judge rows passed vs the top benchmark | 87/336 (26%) | (benchmark) | 105/336 (31%) |
| host-boundary red lines | 0 | 0 | 3 |

**Interpretation, fact by fact.**

- **The full path: 5.1× and 19× (dollars, measured).** The mean solo
  cell costs \$6.07, of which \$1.19 is the customer's verifying work
  (reading, setups, acceptance cycles). Through the machine the same
  task costs \$1.18 (top) or \$0.32 (flash). The machine's economy is
  not a token trick: it took over the verification itself (every check
  twice in clean directories, byte-exact), so the buyer stops paying
  for re-reading.
- **The operator column: 1.19 vs 0.02 dollars.** The machine hands'
  operator fired in 7 of 40 wave-S cells (and 0 of 10 in wave 1), at
  cents per firing; solo's operator fired in every cell. This is the
  first measured appearance of the pattern that dominates §3: the
  machine path removes the verifier from the loop, the solo path
  cannot.
- **Tokens: 5.9M vs 2.8M; cache 58% vs 99%.** Without a machine that
  has already verified, the agent re-reads and re-probes its own
  product; with one, iteration stays inside one context. Quality did
  not pay for it: the blind measurer sees no difference between the
  hands' products (suite ×2 green, deliverables 5/5 everywhere).
- **The price of flash.** The cheapest hand (\$0.32) carries all three
  host-boundary violations of the wave: a weaker model breaks the
  environment mechanics more often — cheapness with a supervision
  caveat.
- **What the machine did not buy.** The "claim ≠ fact" classes were
  still caught only by the human-equivalent operator or the judges:
  13 of 20 solo acceptances went to a second cycle with real findings;
  every cross-judged product fails judge rows exactly on these classes.
  The machine makes verifying reading unnecessary for its own path; it
  does not replace it for someone else's.

## 2. The 0.3.0 metrics — the five-wave field battery (October 2026)

Five waves (conduit → brim → errand → pressmark → quire), 5 senses × 5
stacks × 2 hands = 50 cells, one series binary (0.3.0, sha f0a08dd3);
two hands: solo (a top model without the machine) and machine-top (the
same model through the machine). The era registry's battery summary:

| metric | conduit | brim | errand | pressmark | quire |
|---|---|---|---|---|---|
| cells delivered and accepted | 10/10 | 10/10 | 10/10 | 10/10 | 10/10 |
| machine hands' door refusals | 0 (of 24) | 0 (of 26) | 0 (of 24) | 0 (of 30) | 1 (of 33) |
| `VERDICT: READY` on the first acceptance | 5/5 | 5/5 | 5/5 | 5/5 | 5/5 |
| solo-acceptance rejection cycles | 1 | 0 | 2 | 2 | 1 |
| G1 (the camp) won by the machine | 5/5 | 5/5 | 5/5 | 4/5 | 4/5 |
| Σ-camp total tokens (solo/top ×) | 1.52 | 1.95 | 1.83 | 1.99 | 1.16 |
| Σ-hands total tokens (solo/top ×) | 1.18 | 1.56 | 1.35 | 1.45 | 0.87 |
| Σ-hands time (solo/top ×) | 1.31 | 1.41 | 1.33 | 1.30 | 0.99 |
| operators (machine-camp / solo-camp, total tokens) | 11.32× | 11.82× | 26.31× | 46.01× | 10.57× |
| judge pass (solo product vs the machine table) | 21/96 | 46/89 | 46/57 | 43/129 | 42/103 |

Battery sums over all 50 cells:

| article | solo camp | machine camp | ratio |
|---|---|---|---|
| hands: Σ in+out / wall | 116.98M / 30127.3 s | 93.85M / 24054.8 s | solo ×1.25 tokens; machine ×1.25 time |
| operators: Σ in+out / wall | 42.71M / 11107.7 s | 2.43M / 1205 s | **machine ×17.6 tokens, ×9.2 time** |
| camps: Σ total (with cache reads) | 312.09M | 188.41M | solo ×1.66 |

**Interpretation, fact by fact.**

- **25/25 `VERDICT: READY` on the first acceptance; one door refusal in
  137 submissions.** On the moved boundary (not one byte of code through
  the machine) the verdict stayed a first-shot fact, and the machine's
  own verb wall stayed seconds-scale (0.5–37.8 s per machine cell): the
  gate is not the bottleneck.
- **G1 — the full path — went to the machine in 23 of 25 tasks.** Even
  in quire, the one wave where the solo *hands* were cheaper (§ below),
  the solo operator's 10.57× took back what the hand won: the verifying
  reader is the load-bearing wall of the solo path's economy.
- **The operator multiple: 17.6× tokens, 9.2× wall.** Per cell the solo
  operator averaged 444 s and 1.71M in+out; the machine camp's operator
  48 s and 0.10M. In this campaign's design both camps ran an operator
  per cell (the machine's one micro-decision per cell) — so the
  multiple here is the *discounted* form of the 0.2.0 pattern (where the
  machine operator fired in 7 of 40 cells only).
- **G2 — the hands alone — is task-dependent: 16/4/5.** The September
  re-probing mechanism (solo hands heavier in four waves, 1.18–1.56×)
  operates again — with the quire exception: Σ-hands 0.87, the battery's
  only wave below parity. The registry records the inversion; its cause
  is open work.
- **The judges hold the separation.** The solo product against the
  machine hand's own table, byte-strict: 21/96 … 42/103 — every wave's
  table holds a red mass for solo. The counts are recorded; the class
  decoding of the red rows is not re-derived on this interface.
- **The not-×2 cells.** The blind double replay refused non-repeating
  bytes in three machine cells and one solo cell — the measuring
  procedure's honest boundary; the verdicts were unaffected.

## 3. The 0.4.0 metrics — the case-m battery, the M-series (October 2026)

The ceremony-free machine, measured the same way: 5 senses × 5 stacks ×
2 camps (machine / solo) = 50 cells, 150 agents (50 hands, 50 operators,
50 judges), one binary for the whole series (punchtape 0.4.0, sha
5099556e5621), every agent GLM-5.3 High, the injection wall from M-brim
on. Time is pure request-work (`work_ms`); tokens are in+out unless
named.

Camps by sense (hand + its operator):

| sense | machine: in+out / work | solo: in+out / work | token ratio | time ratio |
|---|---|---|---|---|
| quire | 86.04M / 13010 s | 105.79M / 17586 s | machine ×1.23 | machine ×1.34 |
| brim | 92.05M / 13023 s | 76.78M / 12917 s | solo ×1.20 | solo ×1.01 |
| conduit | 90.33M / 13112 s | 131.10M / 19170 s | machine ×1.45 | machine ×1.46 |
| errand | 83.78M / 10526 s | 70.49M / 10612 s | solo ×1.19 | machine ×1.01 |
| pressmark | 94.49M / 12521 s | 90.92M / 16959 s | solo ×1.04 | machine ×1.35 |
| **series** | **446.68M / 62271 s** | **475.07M / 77244 s** | **machine ×1.06** | **machine ×1.24** |

Cost articles of the series (25 tasks per camp):

| article | machine | solo | ratio |
|---|---|---|---|
| hands: Σ in+out | 443.71M | 403.56M | solo ×1.10 |
| hands: Σ work | 59900 s | 64112 s | machine ×1.07 |
| operators: Σ in+out | 2.98M | 71.51M | **machine ×24.0** |
| operators: Σ work | 2372 s | 13132 s | **machine ×5.5** |
| camps: Σ total (with cache reads) | 884.09M | 936.87M | machine ×1.06 |
| judges (outside the camps) | 69.39M / 14988 s | 72.30M / 15676 s | — |

Criteria score: **G1 13/8/4** (machine/solo/split); **G2 11/12/2** —
parity with mirrored signs. Operator DEMANDs (a demand for proof, from
the wave journals): **machine 2, solo 32** (84 operator answers in
total). Hand medians: machine 16.42M vs solo 15.19M. Quality — blind
judges against the per-sense etalons, 1720 row-scores: machine
**847 done / 13 partly / 0 missing**, solo **852 / 8 / 0** — not one
cell in either camp lost a brief property. Stacks, series cumulative:
js the machine's best (×1.59 tokens, ×1.40 time), rust the only stack
where solo is cheaper in tokens (×1.11), cpp a token parity to 12K of
115M (×1.00) with machine time ×1.15, go and python stable ×1.05 /
×1.26–1.32 for the machine. The whole series: 1063.44M tokens in+out,
≈47 h 17 m of pure work, ≈29 h 35 m of calendar time.

**Interpretation, fact by fact.**

- **The camp verdict: ×1.06 on both token bases, ×1.24 in time.** After
  the ceremony-free rework the machine camp is again the cheaper path —
  and the only era where the time advantage is this large: 0.2.0 bought
  dollars (5.1×), 0.3.0 bought tokens (camps ×1.66 total), 0.4.0 buys
  *ticks* first (×1.24) and tokens second (×1.06).
- **The advantage is two articles, not one.** The machine hands are
  *heavier* in tokens (443.71M vs 403.56M, solo ×1.10) and *lighter* in
  time (×1.07); the machine operators are lighter in both (×24.0,
  ×5.5). The camp win is assembled from the operator article plus the
  hand-time article — and the operator article is the only one that is
  unconditional across every sense and stack of the series.
- **G2 parity is the rework's signature.** In 0.2.0 the loop's hand
  spent 2.8M against solo's 5.9M (the machine had absorbed the
  verification); in 0.3.0 the hands still split task-dependently with a
  ritual premium; in 0.4.0 the hands alone are at parity (11/12/2,
  medians 16.42/15.19M). The work itself now costs the same in either
  camp — only in a different currency: the machine trades tokens for
  ticks. All four G1 splits of the series are of exactly one kind —
  tokens to solo, time to the machine — the trade, seen cell by cell.
- **2 vs 32 DEMANDs.** The solo operator had to *construct* proof 32
  times in 25 tasks (spec → scenarios → green run → coverage demands);
  the machine camp's 2 were substantive exceptions (an unproven stop,
  an unproven interrupt). In the machine camp the verification already
  lives inside the verdict — 50/50 cells closed with standing signed
  verdicts of full accounting.
- **Quality does not differentiate; economics does.** 0 missing of 1720
  row-scores in both camps; the 13-vs-8 "partly" difference is edge
  polishing of refusals and stops, within noise. Judging machine cells
  costs slightly less (69.39M vs 72.30M, ×1.04) — the rendered zone
  prose reads heavier than a raw solo tree.
- **Stacks and senses.** The machine's best case is a buildless
  interpreted stack (js: an own short cycle per machine pass); its
  worst is long compiles (rust: each machine pass pays the build, and
  the time edge shrinks to ×1.12). By sense, conduit fits the machine
  best (a counters-and-routes law lands wholly in the corpus, 4/5,
  ×1.45/×1.46), brim worst (a compact image conveyor assembles cheaper
  solo, 2/3 to solo). The win depends on the task's shape more than on
  the stack.
- **Boundaries of the reading.** One binary and one model pin for the
  whole series: internal validity full, external (other models) not
  claimed. Live-subagent ceilings varied by waves — wave walls are not
  compared, only work sums. Medians sit near the means: no single cell
  carries the totals.

## 4. The operator is an LLM: the live-human renormalization

### 4.1. The record and why it matters

In every experiment above the "operator" — the accepting, verifying,
demand-answering party — was itself a language model: **GLM-5.3 Max in
the 0.2.0 era, GLM-5.3 High in 0.3.0 and 0.4.0** (the model record;
the ledgers pin "GLM-5.3"). Two consequences:

1. Every operator dollar in §1 is an *AI-operator* dollar at 2026-09-03
   Max prices; every operator second in §2–§3 is an *AI's* wall.
2. For a live human the same work is strictly more expensive in time:
   a person reads at reading speed and re-runs by hand (the AI re-reads
   the whole context in seconds and fires its own probes); a person
   carries setup, fatigue and context switches the AI's wall does not
   contain; and a person's answer is asynchronous — the recorded waits
   are the machine's side of the pause, never the human's calendar.

The research paper's Appendix E anchored the correction on the 0.2.0
data: the solo acceptance operator's AI wall averaged 530 s (8.8 min)
per task; the live-human anchor for the same class of work is 10–18 min
per task — **×1.2–2.0 of the AI wall** (a labeled estimate from the
August finishing estimate, not a measurement of these waves). That
anchor is carried below unchanged, applied per era to that era's own
measured operator wall.

### 4.2. The operator article, era by era, as measured

| era | solo operator (measured) | machine operator (measured) | ratio |
|---|---|---|---|
| 0.2.0, wave S | \$1.19 per cell; 530 s mean wall; fires in **every** cell (13/20 second cycles) | \$0.02 per cell; fires in 7 of 40 cells | dollars ×60; firings 40:7 |
| 0.3.0 battery | 42.71M in+out / 11107.7 s per 25 cells (444 s, 1.71M per cell) | 2.43M / 1205 s per 25 cells (48 s, 0.10M per cell) | ×17.6 tokens, ×9.2 time |
| 0.4.0 M-series | 71.51M / 13132 s per 25 cells (525 s, 2.86M per cell); **32 DEMANDs** | 2.98M / 2372 s per 25 cells (95 s, 0.12M per cell); **2 DEMANDs** | ×24.0 tokens, ×5.5 time |

Read plainly: across all three generations the machine camp's verifier
costs cents-to-negligible and fires on exceptions; the solo camp's
verifier is a standing article — and the only article of the comparison
that never flips sign in any wave, sense or stack of any era.

### 4.3. Renormalized to a live human operator

Substitute a human into the operator seat; the executor stays a model
in both camps (that is the realistic deployment: the model writes, a
human accepts). Hands are untouched by the substitution; only the
operator article moves. Applying the ×1.2–2.0 anchor to the 0.4.0
M-series walls:

| operator cost per 25-task series | solo camp | machine camp (like-for-like) | machine camp (substantive only) |
|---|---|---|---|
| AI wall (measured) | 13132 s = 3.65 h | 2372 s = 0.66 h | — |
| human wall (×1.2–2.0, labeled) | **4.4–7.3 h** | 0.8–1.3 h | 2 DEMAND answers ≈ 0.2–0.5 h |
| at \$30/h | \$132–219 | \$24–40 | \$6–15 |
| at \$50/h | \$219–365 | \$40–66 | \$10–25 |
| at \$100/h | \$438–730 | \$79–132 | \$19–50 |
| at \$200/h | \$877–1460 | \$158–263 | \$38–99 |

("Like-for-like" prices the same verification work at a human's pace;
"substantive only" counts what the machine camp actually needs a human
for — the two exception answers; the machine's own verdicts carried the
other 23 tasks' acceptance with zero human involvement. Token prices of
the High tier are not recorded — deliberately no dollar figure is
invented for the AI-operator rows of 0.3.0/0.4.0.)

**The savings and the inversion they cause.** Like-for-like, the machine
camp saves **3.6–6.0 hours of verifying labor per 25 tasks** (\$110–1200
across the rate table); counted substantively, it saves **~4–7 hours —
essentially the whole verifying article**. Under an AI operator the
machine camp won the series by ×1.06 in tokens — a margin a token-price
shift could erase. Under a live human the margin does not depend on
token prices at all: it is hours of skilled attention per series, plus
the calendar the AI wall never models — a solo camp human-verified task
by task serializes the whole wave on the operator's availability,
while the measured series ran 25+25 tasks in ≈29.6 h of continuous
calendar. The 0.2.0 paper's conclusion ("the product of the effect is
not token price but labor displacement") is thus the *stronger* claim
of the later eras: every generation moved the same article — the
verifier — further out of the human's day, and the ceremony-free 0.4.0
machine finally left the *hands* at parity too, so the entire remaining
difference between the camps is exactly the article a human would have
to live in.

### 4.4. Honesty marks of this section

The ×1.2–2.0 human anchor is a labeled estimate (Appendix E's August
anchor), not a measurement of these series; the per-rate dollar table
is parametric arithmetic on that anchor; the model-tier record (Max /
High) stands over ledgers that pin "GLM-5.3" without a
tier; the DEMAND counts come from the wave journals; the 3.65 h and
0.66 h AI walls are the measured `work_ms` sums. Nothing else here is
extrapolated.
