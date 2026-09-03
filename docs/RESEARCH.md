# Closing the verification loop in generative development: an external transactional state machine and a field evaluation of its economics and convergence

**Anton Smolin**

Independent researcher

hey@neurophant.tech

## Abstract

Generative development ("vibe coding") is an open loop: the code is
produced by a language model, and its readiness is judged by an opinion
— the executor's or the author's. This work studies the hypothesis that
closing the loop with a transactional state machine external to the
agent — one that owns the canon of requirements, an executable checks
table, a double run of every check, and a signed verdict — turns
readiness from an opinion into a reproducible fact at a measurable
price. Field experiments were run on a battery of 25 tasks (5 task
senses × 5 technology stacks): a baseline run, eight control waves, a
repeat full run after systemic fixes, and a three-hand comparative
experiment (a top model in the system, a budget model in the system,
solo execution without the system) with independent judges. The system
hands delivered 50/50 machine verdicts of "verified twice" with zero
minutes of intermediate human attention, with quality invariant to the
model class (the budget loop — the same verdicts for twice the time
and an eighth of the money); the solo executions, with better
economics (−31% tokens, −22% time against the top hand), converged to
the byte-exact reference in only 22% of the judges' table rows, and
the bulk of the divergence was freedom of interpreting the brief, not
functional defects. Systemic fixes between the full runs cut the
battery's cost by 49% of tokens without changing the methodology. A
September replication campaign on the evolved machine (v0.1.x → 0.2.0)
— a 75-cell three-hand wave pair, a 20-cell release-validation wave, a
three-language smoke — reproduced the verdict and quality invariance,
the judges' protocol-red pattern, and the attention economics. The
conclusion: the benefit of the closed loop is measured not in tokens
but in reconcilability; an anti-entropy frame is proposed for
interpreting the result.

**Keywords:** generative development, verification, state machine,
specification, acceptance checks, compute economics, executor
convergence, field experiment.

## Glossary of the paper

The system's full dictionary is `docs/GLOSSARY.md`; below are the
paper's working terms.

- **Battery** — the task set of a field campaign: 25 tasks = 5 senses ×
  5 stacks.
- **Brief** — the task's wish in living human language; the executor's
  only substantive input.
- **Verdict** — the machine's final output on readiness (`READY` /
  `NOT-READY` / `QUARANTINE`), signed with a digest.
- **Control wave** — a run of one sense on five stacks on an unchanged
  binary after a wave of fixes.
- **Gate** — the machine's checkpoint of a submission (statics,
  coverage, tracing, double run, probe, latency).
- **Digest** — a short cryptographic fold; the signature of a verdict
  and of row identity.
- **Double run** — executing every check twice in clean directories;
  the defense against "green on a cache and garbage".
- **Recommended default** — the fork option the machine applies when
  the human stays silent; written to the journal with tracing.
- **Byte capture** — repairing expectations by intent: the machine
  measures the new bytes by a run; the human asserts the meaning.
- **Probe** — a check of the checks: catches checks that cannot fail.
- **Sign test** — a nonparametric test for paired comparisons; here —
  exact binomial probabilities at p = 0.5.
- **Canon** — the task's permanent state (requirements, scenarios,
  checks) owned by the machine; outlives sessions.
- **Quarantine** — an honest stop of a task after the repair budget is
  exhausted, with a named reason.
- **Ledger** — the instance's journal of facts (submissions, gates,
  time, latency); the source of the machine metrics.
- **S band** — the calibrated difficulty band: compact console
  utilities with file state.
- **Seed** — the content of state files, laid out by the machine
  before a row's run.
- **Acceptance** — the machine's repeat reconciliation of the final
  checks composition on the frozen state (acceptance, PASS with a
  digest).
- **Brief fork** — a point where the wish is silent; closed by a
  recorded default or an escalation.
- **Hand** — one parallel executor of the three-hand experiment:
  **top** — GLM-5.3 Max (maximum reasoning mode);
  **flash** — GLM-5.3 Flash (high reasoning mode);
  **solo** — GLM-5.3 Max without the machine; **judge** — GLM-5.3 Max,
  reconciles the solo product against the top hand's checks table.
- **Reconcilability / convergence** — the share of the product's rows
  that match the reference table byte-for-byte; the strict count —
  against the declared table, the conditional count — with the judge's
  neutral driver.
- **Wall** — the executor's time per task. August series: the ledger /
  the environment's accounting of the run. September replication: the
  full wall — the sum of the durations of all the executor's platform
  segments, rotations and restarts included (the campaign registry's
  basis, `docs/METRICS.md` §1.2). The two bases are never mixed inside
  one comparison.
- **submit p95** — the 95th percentile of the machine's submission
  processing time; the budget is 300 ms.
- **# TOKENS** — the submission trailer with the executor environment's
  honest token counter.
- **Harness** — the agent's execution environment (dialog, tools,
  session); it runs the conversation, but not the task's state.

## 1. Introduction

### 1.1. The problem

Language models have learned to produce working code from a human's
wish, but not to answer for its readiness. The typical cycle is "write
and believe": the assessment of the result stays subjective, the checks
are irregular, and the divergence between the customer's intent and the
product's behavior is discovered late and expensively. Agent
orchestration libraries solve an adjacent problem — managing the course
of the dialog — but the task's state in them lives in the session's
memory and dies with it; a guaranteeing party is absent.

The gap has a measurable price. By the data of this work's three-hand
experiment, a solo execution that is "green in its own way" diverges
from the byte-exact reference in 78% of the check rows, while the
executor's self-assessment does not reveal these divergences: finishing
the solo product to verifiable quality cost an additional 4–7.5 hours
of a live human per batch — against zero for the system hands.

### 1.2. The approach

The object of study is an external transactional state machine
(punchtape): a console deterministic arbiter that owns the task's state
for good (the canon: requirements, scenarios, checks), accepts typed
submissions from an executor agent, compiles the formalism itself, runs
every check twice in clean directories itself, and issues a verdict
with a cryptographic digest: `READY`, or an honest `NOT-READY` with
named reasons. The agent never touches the state directly — it can only
propose transactions; every submission is either accepted whole or
rejected with a one-line reason.

The inversion of ownership is the key difference from embedded
libraries: the machine stands outside, the task's state outlives
sessions, and models change without altering what has already been
punched into the canon.

### 1.3. Goal, tasks, novelty

The goal is to test whether an external transactional state machine
makes "done" reconcilable without paying for the guarantee
proportionally, and to measure the price of closing the loop. Tasks:
(1) build a field evaluation methodology; (2) measure the economics and
quality of the closed loop against the open one; (3) separate
functional defects from freedom of interpretation; (4) test the
reachability of zero intermediate human attention. The novelty is in
the experiment's design: identical tasks are executed three ways with
an independent judicial reconciliation of the products against the
byte-exact tables of the machine runs themselves, and all numbers come
from transaction ledgers, not self-reports.

### 1.4. Related work

**Agents in software engineering.** The field is systematized by a
survey of 124 studies of LLM-based agents across the software
lifecycle [1] — the designs surveyed there orchestrate the dialog and
the tools, the task's state living in the session. The same literature
contains its own counterpoint: Agentless [2] shows a simple three-phase
localize–repair–validate pipeline outperforming complex autonomous
agents on SWE-bench Lite at a lower cost. This work takes the
counterpoint further: the guaranteeing party is not a smarter or
simpler agent but a transactional state machine outside the executor —
the state outlives sessions, and the verdict is signed, not asserted.

**Closing the loop with verification.** A growing line of work equips
LLM code generation with verification cycles: Detect–Repair–Verify
pipelines for securing LLM-generated code [3]; formal-method guidance
that routes LLM-generated safety-critical code through deductive
verification and turns each failure into a correction prompt (Forge)
[4]; self-generated formal specifications with verifier feedback,
where the self-generated specification is itself the measured
bottleneck (Self-Spec) [5]; and security tests used as executable
specifications shown to the model up front [6]. The machine of this
paper belongs to this line with two differences: the checks table is
human-readable and the canon is compiled from it by the machine, not a
formalism the model authors; and every check is run twice in clean
directories by the machine itself — the loop's owner and the loop's
executor are different parties.

**Vibe coding, measured.** The open loop this paper studies is the
object of a survey of over a thousand works that defines vibe coding
formally [7]. Empirically, an assessment of 450 LLM-generated scripts
reports ≈85% execution success but ≈45% silent failures of flawed
logic behind syntactic correctness, and calls for deterministic
wrappers and strict governance [8]. Our field data independently
reproduces the same class — "green in its own way", claim ≠ fact — and
measures what a deterministic wrapper costs and buys when it owns the
task's state.

**Judges.** Since the judges here are LLMs, the known fragility of
LLM-as-a-judge evaluation — verdicts sensitive to prompt-induced bias,
enough to change task-level conclusions and model rankings [9] — is
directly relevant; the design answers it by construction: the judge
executes a fixed table row by row, byte-strict, with no right to edit
or adapt, a doubt counts against the product, and a calibration must
pass before any count is recorded.

## 2. Methodology

### 2.1. The object

The machine is a single binary (Go); the sources and the full
documentation are in this repository (v0.2.0). The composition: canon
data (every human-readable surface and threshold as data, not code), a
transaction journal, a metrics ledger, a compiler from the checks table
to scenarios, gates (statics, coverage, tracing, an empty-check probe,
the double run), a spec registry, and a render of human-readable
per-feature documents.

### 2.2. Field campaigns as the measurement method

Measurements are performed by field campaigns: batteries of live tasks
executed by agents in controlled isolation, with numbers taken only
from machine evidence. The rules of an honest run:

- **isolation**: one task — one agent — one throwaway container; the
  network is off, 1 CPU, 1 GB of memory; the repository and the
  documentation are never mounted into the task world
  (anti-cheating); the materials are read-only;
- **the binary is verified**: double build, byte-for-byte comparison,
  the version stamped in, the SHA-256 in the wave journal; inside the
  compared runs the binary does not change;
- **offline builds**: every task dependency is pre-cached in the world
  image; a task that needs the network is refused loudly;
- **metrics**: from only two sources — the instance ledger
  (submissions, verdicts, wall, latency) and the executor environment's
  accounting (time, tokens, tool calls); taken immediately at
  completion; every submission carries its environment's honest token
  counter;
- **a cause registry**: every notable expense — a line of "cause →
  price → who caught it".

### 2.3. The battery

25 tasks: 5 task senses × 5 stacks (python, go, rust, c/c++, js). The
senses: a margin detector for scanned images; a two-stage TCP relay; a
background job queue over Redis; a file-based static site build; a
no-config static site generator from a directory. Each sense exists as
five briefs — the wish in living human language with the stack and the
libraries pinned; within a sense the briefs are identical, the
materials byte-for-byte the same. The battery covers the S band of
compactness (utilities with file state); the tasks of one sense are
comparable across the five run factors.

### 2.4. The comparative experiment design

The three-hand mode: every task is executed three ways — (A) the top
model **GLM-5.3 Max** (maximum reasoning mode) through the machine;
(B) the budget model **GLM-5.3 Flash** (high reasoning mode) through
the machine; (C) the top model GLM-5.3 Max without the machine (solo).
The worlds, the briefs, and the limits are identical; for solo the
binary is not mounted. A fourth agent — the judge — receives the checks
table of hand (A) for the same task (the byte-exact expectations the
machine has already verified) and the solo product, executes every row
by hand, and compares the observed byte-for-byte; it has no right to
edit the product; a doubt counts against the product. The result — the
share of green rows as the measure of strict convergence.

### 2.5. Hypotheses

- **Hypothesis 1 (reconcilability without proportional payment).**
  Moving verification into an executable table with a double run makes
  "done" a reproducible fact, without requiring a multiple increase
  in cost relative to the open loop.
- **Hypothesis 2 (systemic economy).** The cycle's cost reduction is
  achieved by properties of the system itself (byte capture by the
  machine, elimination of manual retranscription, suppression of noisy
  questions), not by tuning the executor's prompts.
- **Hypothesis 3 (freedom of interpretation).** Without an external
  canon, executors of equal class systematically diverge in reading
  the same brief; the divergences are not revealed by self-reports and
  do not correlate with functional quality.
- **Hypothesis 4 (the model's price in the loop).** In the closed loop
  the model class does not determine whether the verdict is reached,
  but determines the time; the budget model in the system loses on the
  total cost of the path.
- **Hypothesis 5 (zero intermediate attention).** For the tasks of the
  S band, a machine `READY` is reachable without a single intermediate
  human answer, with all the forks recorded by the journal and
  defaults.
- **Hypothesis 6 (self-gating overhead).** The machine's own overhead
  can be subjected to the same gates as the product (budget limits on
  latency and the ceremony share).

## 3. Experimental record

All the experiment dates lie in the interval from August 30 to
September 3, 2026; the August series ran on a single machine release
(v0.1.0); the binary inside each comparison is unchanged. The cost is
given at the August 2026 API list prices; for comparability the
ratios matter, not the absolute values (the top model's token is
≈ 9 times more expensive than the flash model's token).

### 3.1. The baseline run (2026-08-30)

25/25 tasks finished with a machine `READY` and a successful
acceptance; 0 quarantines; the token sum 80.31 M; 53 rejected
submissions; 25 questions to the operator.

### 3.2. The repeat full run (2026-08-30, after a wave of systemic fixes)

25/25 `READY`; every task run exactly once; tokens 80.31 → 41.00 M
(**−49%**); the task wall median 17.3 → 13.5 min; submissions 219/53
(accepted/rejected) → 142/12; questions to the operator 25 → 3 (all
closed autonomously by recommended defaults); submit p95 46–175 ms at
the 300 ms budget; the machine's fix cycles — 4 per battery; 0
quarantines. The methodology and the battery did not change; the
changes are only systemic properties of the machine (see hypothesis 2).

### 3.3. The three-hand experiment (2026-08-31)

75 executor runs (25 tasks × 3 hands) and 25 independent judge runs.
The system hands: **50/50 `READY` + acceptance** (25 top + 25 flash),
0 quarantines, a human was never needed. Solo: 25/25 products with
green self-checks (from 12 to 283 own checks per task).

The hands' economics (the summary measures over the 25 tasks of each
hand):

| Hand | Tokens, M (sum / median / min–max) | Wall, min (sum / median / min–max) | Battery cost guide |
|---|---|---|---|
| system + top | 65.31 / 2.23 / 0.60–7.66 | 435.6 / 14.4 / 8.2–33.8 | ≈ $111–131 |
| system + flash | 72.78 / 2.32 / 0.66–10.05 | 756.7 / 29.0 / 11.7–61.9 | ≈ $13.5–16 |
| solo (no machine) | 45.22 / 1.41 / 0.39–5.46 | 339.9 / 11.4 / 7.1–22.5 | ≈ $77–90 |

Solo is the cheapest and the fastest hand; flash is the most expensive
by the total path (the cheap step is paid for by the number of steps).

Strict convergence (the judges against the byte-exact tables): **108/491
green rows (22%)**; per sense (green/rows): conduit 34/97 (35%),
pressmark 32/118 (27%), errand 0/77 (0%), brim 24/112 (21%),
quire 18/87 (21%). The pattern's decoding is stable across all 25
judges:

1. the solo functional core is almost always byte-for-byte green
   (the measured quantities, the built files, determinism, isolation);
2. the red mass is **protocol, not function**: the diagnostics texts,
   the exit codes, the usage lines, the enumeration formats;
3. the same brief forks are decided oppositely by different hands
   (including the top hands): "three top hands — three different
   answers";
4. one wave's zero is an artifact of the comparison surface (the
   reference table drives the library's own test driver, which solo
   does not have); the conditional count with the judge's minimal
   driver: 11/11, 24/24, 8/11, 8/10 — 51/56, the library's behavior is
nearly complete;
5. the solo real functional failures — 5 per battery, all pinpoint:
   an RST cut instead of a clean close and the exit code on a taken
   port (conduit-rust), the missing name field (quire-rust), the
   silently accepted repeated option (pressmark-js), the "None"
   string instead of an empty value (quire-python) — inside "green in
   its own way" they are invisible to the self-report.

Finishing the solo products to verifiable quality (the same battery)
is estimated at ≈ $204–240 and 4–7.5 hours of a live human — against
zero intermediate participation for the system hands.

### The per-task structure (paired comparisons)

The run's full per-task table is Appendix D; its paired structure is
more informative than the summaries:

- **Flash is slower than top in 24 of the 25 tasks** (the only
  exception — errand-go: 20.8 against 21.8 min). In the absence of an
  effect, the probability of a preponderance at least this large is
  7.7·10⁻⁷ (sign test, n = 25 pairs): the flash model's cheaper step is
  paid for by the number of steps practically always, not on average.
- **The hands' token medians are almost equal (2.23 / 2.32 / 1.41 M),
  the tails are not**: the top hand's maximum 7.66 M (conduit-js), the
  flash hand's 10.05 M (pressmark-cpp), solo 5.46 M. Flash's total
  expensiveness (+11.4% against top) is accumulated by the right tail,
  not by the typical task.
- **Solo is cheaper than top in tokens in 19 of the 25 tasks**
  (p ≈ 0.007, sign test) — solo's economic advantage is stable at the
  per-task level, not only in the sum; the losses to top are
  concentrated in the heavy tasks (errand-cpp 3.31 against 1.56 M,
  brim-rust 3.04 against 1.77 M).
- **The spread inside a hand class is larger than the difference
  between the classes**: the top hand's tokens 0.60–7.66 M, flash's
  0.66–10.05 M, solo's 0.39–5.46 M. The dominant factor of a single
  task's cost is the executor's variability (its strategy of probes
  and rehearsals), not the model class; the system gates quality, the
  cost variability remains the executor's property.
- **The errand wave's conditional judge count**: the strict 0/77 — the
  solo product does not have the CLI surface declared by the table
  (a library plus its own tests); with the judge's neutral minimal
  driver the library's behavior is complete: 51/56 conditionally green
  (11/11, 24/24, 8/11, 8/10 across the go/rust/cpp/js tasks). The
  strict red remains an honest fact of the missing surface.

### 3.4. Human attention and forks

In the three-hand run, 74 brief forks (the points where the wish was
silent) were closed by the machine's defaults with a journal and
tracing; not one required a live answer. In the repeat full run the
operator questions: 25 → 3, all closed autonomously. The escalation
rule: when a human decision is requested (code 3/4), the executor
applies the recommended option, marks it "awaiting ratification", and
continues — in the isolated worlds nothing is irreversible.

### 3.5. Environment observations (the cause registry, a selection)

- An OOM kill (code 137) is a fact of the environment's limits, not of
  the task: recreate the world with a larger limit, log the case.
- Files written from a container belong to root: the ownership must be
  returned before the world's removal; after every wave — a zero
  root-owned files check.
- Losing the platform's metrics taught the rule of taking them from
  the instance immediately: a lost cell stays lost.
- The baseline's largest class of losses — the executor's manual
  retranscription of expectation bytes (tens of percent of the tasks'
  tokens); the machine's systemic byte capture (the new expectations
  are measured by a run; the human asserts the meaning) removed the
  class entirely.
- Library resolution from the throwaway run directories is a property
  of the world environment, not of the machine; after its
  systematization the battery's most expensive task fell from 10.45 to
  1.10 M tokens.

## 4. Replication on the evolved machine (2026-09-03)

The study above ran on v0.1.0. Before publication the machine went
through a series of systemic fixes to the 0.2.0 release (v0.1.x →
0.2.0), and the measurement design was re-run on the evolved machine
as a September campaign: two full three-hand waves (wave 1: one
sense × five stacks × three hands, 15 cells; wave S: the other four
senses × five stacks × three hands, 60 cells — together covering the
same 5 × 5 × 3 battery shape as the August series), a
release-validation wave on the frozen 0.2.0 candidate (five senses ×
two stacks × two machine hands, 20 cells, the arbiter's sha recorded
in the wave journal), and a live three-language smoke (ru/en/zh). The
prices are the 2026-09-03 market tariffs; as in §3, the ratios, not
the absolutes, carry the meaning.
All numbers below are the campaign registry's (`docs/METRICS.md`),
every one mechanically revalidated against the run evidence — the cell
captures and the platform ledger; nothing is retyped from a report.

**Measurement bases.** The September campaign reports the executor's
time as the full wall — the sum of the durations of all the executor's
platform segments, rotations and restarts included (the campaign
registry defines the full wall and every other metric's provenance —
`docs/METRICS.md` §1.2); the August series' wall is the ledger /
environment accounting of its era. The bases differ (the full wall is
the stricter one: it adds what a single-segment accounting prunes — the
registry's own basis correction moved 21 multi-segment cells to the
full wall), and no comparison in this paper mixes them: the August
numbers are compared within Appendices A–D, the September numbers
within this section.

| September wave | Machine | Cells | What it replicates | Result |
|---|---|---|---|---|
| wave 1 + wave S | the v0.1.x line (a pinned binary per wave) | 75 = 25 tasks × 3 hands | the three-hand experiment of §3.3 | machine hands 50/50 `VERDICT READY`; blind measurer: suite ×2 green, 5/5 deliverables in 75/75 products; solo 20/20 accepted via operator cycles |
| validation wave | 0.2.0, frozen | 20 machine cells | the fixes hold on the release candidate | 20/20 `VERDICT READY`, zero panics / deadlocks / quarantines / lost cells, zero new machine defect classes |
| smoke (ru/en/zh) | 0.2.0, frozen | 3 live wishes | the wish → verdict cycle with no human in the loop | `READY` ×3, zero question batches, $0.66 total, 256 s parallel wall |

### 4.1. What replicated

- **Verdicts and quality (hypotheses 1, 5).** Every machine run of the
  campaign closed at `VERDICT READY` — 50/50 in the three-hand waves,
  20/20 on the frozen candidate; the blind double-run measurer found
  suite ×2 green and 5/5 deliverables in all 75 wave-1+S products (the
  validation wave: 16/20, the four exceptions are a measured-procedure
  nondeterminism, not product defects — single runs of the affected rows
  are green). The smoke reached READY in all three languages with zero
  question batches.
- **The judges' pattern (hypothesis 3).** The September judges reproduce
  the August decoding: the solo functional core is byte-green, the red
  mass is protocol and surface. Strict convergence against the top
  benchmark: solo 87/336 rows (26%), the budget hand 105/336 (31%);
  13 of 20 solo acceptances went to a second verification cycle with
  real findings — the same class of work as the August 4–7.5-hour
  finishing estimate.
- **The model class (hypothesis 4).** Identical mechanical quality at
  both model classes; the flash hand is 3.75× cheaper by the hand mean
  ($0.31 against $1.16) and ≈2× slower on the full wall (2258 against
  1131 s mean) — "time or money", not a quality tradeoff, confirmed.
- **Attention (hypothesis 5).** The machine hands triggered operator
  involvement in 7 of 40 wave-S cells (waits 16–59 s; campaign-wide,
  at the top model, a batch was triggered in 4 of 25 cells) and in 7
  of 20 validation-wave cells (waits 23–216 s, $0.60 in total at the
  input-rate upper estimate); the solo path needed its acceptance
  operator at every task — the operator's AI wall averaged 530 s per
  task (range 258–937 s), a lower bound for a live human (the anchored
  estimate: ≈10–18 min per task).
- **Economics — sharpened, not just confirmed.** The August series
  priced the hands: solo won the hand's own tokens (−31%) but its
  "done" carried an unpriced human finishing (≈$204–240 and 4–7.5 h per
  batch, §3.3). The September campaign prices the full path — the hand
  plus its verifying operator, cache-aware tariffs: solo $6.07 per task
  (hand $4.88 + operator $1.19) against the top hand's $1.18 (5.1×) and
  the flash hand's $0.32 (19×). Under full-path accounting the closed
  loop is the cheaper path, not the more expensive one — the August
  conclusion (the benefit is reconcilability, not economy) stands, and
  its hidden term is now measured. One further measured mechanism: the
  September solo hand is itself the token-heaviest of the three
  (5.95M life-tokens per task against the top hand's 2.78M, cache
  coverage 58% against 99%) — without a machine that has already
  verified, the agent re-reads and re-probes its own product (the
  per-task full-path table is Appendix E).
- **Self-gating (hypothesis 6).** The validation wave's criteria closed
  green: the gate stack green on every tree of the wave, zero lint
  rejections, zero non-UTF-8 surfaces in the 20 exports, zero zombie
  incidents, zero journal breaks.

### 4.2. What the replication adds

Two findings the August series could not see. First, the artifact-path
class: in the validation wave's cross-judgment (the flash product
against the top benchmark) 7 of 10 tasks scored zero strictly because
the benchmark rows call the top hand's declared surface while the flash
product declared its own — a brief/conventions fix candidate, not a
machine or model defect. Second, the blind-suite nondeterminism of the
four validation cells (a container per run, not per case) — an honest
boundary of the measuring procedure itself, caught by the double run.

## 5. Discussion

### 5.1. Reconcilability, not economy

Hypotheses 1–3 are confirmed jointly. The closed loop does not make
the task cheaper than solo (solo wins both tokens and time — in the
August accounting of the hand's own work; the September full-path
accounting, §4.1, counts the verifying operator in the price and
flips the sign: the machine path is the cheaper one), but it
changes the very subject of the purchase: the executor without the
machine delivers "its own opinion of readiness"; the executor with the
machine — a product to which an executable exam is attached,
re-verifiable at any moment at zero marginal price. Solo's 22% strict
convergence at an almost fully green functional core shows that the
gap lies in interface discipline (protocols, texts, codes) — exactly
where the contract should live. Freedom of interpretation is a
measurable phenomenon: it accumulates across all the brief's forks and
fires at any change of executor; the external fixation of decisions
(the canon + the journal of defaults) moves it from invisible to
signed.

### 5.2. The anti-entropy frame

The uncertainty of meaning behaves like entropy: it grows on its own
unless it is pumped out. The machine acts as a heat pump: the local
reduction of uncertainty is paid for by work — the tokens of double
runs, the containers, the computation. By Landauer, erasing a bit
costs ≥ kT·ln2 ≈ 3·10⁻²¹ J; the task's information entropy (~10⁻¹⁵ J,
even computed generously) is negligible against the ~10⁵ J of real
computation — the heat pays not for the annihilation of uncertainty
but for search and measurement. The unavoidable forks are not erased
but move into an explicit, signed form (question → default →
journal); the order is stored: the canon + a green verdict form a
"battery of order" that makes regression re-runs nearly free. The
result's formula: cheap measurable heat buys expensive unmeasurable
uncertainty.

### 5.3. The model's price and the path

Hypothesis 4 is refined by the experiment: the model class does not
separate from the verdict (50/50 for both system hands), but separates
by time — flash is slower than top in 24 of the 25 tasks (the median
29.0 against 14.4 min; p = 7.7·10⁻⁷) at an equal median token price
(2.32 against 2.23 M); the "top or flash" choice is "time or money",
not a quality tradeoff. Solo's economic advantage (cheaper than top in
19 of the 25 tasks, p ≈ 0.007) is real, but it is bought at the price
of incomparability: 22% strict convergence means that solo's "done" is
the executor's private opinion, not a portable fact.

### 5.4. The invariance of quality to the model class

The three-hand experiment's main practical result: **the delivery
quality does not depend on the model class inside the closed loop** —
both system hands delivered 25/25 `READY` + acceptance, 0 quarantines,
0 live questions. The difference is only in time and money:

| Measure | system + top | system + flash | ratio |
|---|---|---|---|
| Verdicts READY + acceptance | 25/25 | 25/25 | 1 |
| Wall median per task | 14.4 min | 29.0 min | ×2.0 |
| Battery tokens | 65.31 M | 72.78 M | ×1.11 |
| Battery cost | ≈ $111–131 | ≈ $13.5–16 | ≈ ÷8 |
| Task cost | ≈ $4.4–5.2 | ≈ $0.54–0.64 | ≈ ÷8 |

The cheap model in the system delivers the same signed result for
twice the time and roughly an eighth of the cost: the cheap step is
paid for by the number of steps (the token medians are equal — 2.23
against 2.32 M; flash is slower in 24 of the 25 tasks, p = 7.7·10⁻⁷),
not by quality. The "top or flash" choice is "time or money" at the
same result; outside the loop the same choice is "money or risk"
(solo: −31% tokens against top at 22% strict convergence and finishing
by a live human).

### 5.5. The machine and the harness: the separation of ownership

The harness (the agent's execution environment) and the machine do not
compete: the agent lives in the harness, the machine stands outside and
owns the task's state. The separation of responsibility, as shown by
the experiments' data:

| The harness alone | The machine | The evidence |
|---|---|---|
| readiness is the agent's opinion | the verdict is an executable fact with a digest | 50/50 verdicts against solo's 22% strict convergence — in the same harness |
| the task's state dies with the session | the canon lives indefinitely; a change of executor loses nothing | "three top hands — three different answers" on one fork; 74 forks closed by journaled defaults |
| a check on someone's word: one run, possibly with a cache and garbage | every check twice in clean directories, an empty-check probe | determinism is a property of the procedure, not of the agent's discipline |
| the platform's metrics are a side effect and get lost | the instance ledger, the `# TOKENS` trailers: the price does not depend on the platform | the actual losses of the environment's metrics in the runs; the rule of the immediate take |
| the overhead is not measured | the overhead is itself a gate | submit p95 46–175 ms at the 300 ms budget; the ceremony ≤ 10% |
| the economic fixes — by a prompt, anew for every agent | the systemic properties — once, for everyone | −49% of tokens between the full runs without a change of methodology; the manual byte-retranscription class eliminated entirely |
| the endless "almost done" loops | a quarantine with an attempt budget | 0 quarantines across the 50 system runs — the loops end, they are not masked |

The separation formula: the harness runs the conversation; the machine
owns the state and the verdict. The economics of closing the loop pays
cheap measurable heat (the tokens of the double runs) to pump out
expensive unmeasurable uncertainty (the freedom of interpretation); a
harness without an external canon gives the solo mode — fast, cheap,
and incomparable.

### 5.6. Validity

- **Internal**: an unchanged binary inside the comparisons (the double
  build, the digests); identical briefs, materials, and limits of the
  hands; anti-cheating isolation; the judges without the right to
  edit; the control waves between the full runs.
- **The reference's circularity**: the judges' checks table is produced
  by the very system the experiment evaluates. The mitigations: the
  table's rows are the measured bytes of the top hand's product's
  observable behavior, not an opinion; the judge executes the rows
  independently and by hand; a doubt counts against the product. The
  residual risk — expectations systematically simplified by the
  machine would understate the solo assessment — which is why the
  conditional count with the judge's neutral minimal driver is given
  in parallel (item 4 of the pattern's decoding).
- **The statistics of the inference**: no parametric intervals were
  built at n = 25; for the hands' paired comparisons the sign test was
  applied (flash against top by time: 24/25, p = 7.7·10⁻⁷; solo against
  top by tokens: 19/25, p ≈ 0.007; the exact binomial probabilities at
  p = 0.5); the other comparisons rest on the agreement of the sum,
  the median, and the extremes simultaneously (the hands' economics
  table, Appendices A–D). The multiplicity of tests was not corrected:
  the two tests were chosen a priori, by the main hypotheses.
- **Constructive**: the numbers — only from the ledgers and the
  environment's accounting; one parser; honest notes on the losses and
  the comparison artifacts (one wave's zero is parsed and explained;
  the conditional count is given).
- **External** (the limits): one difficulty band (S: compact console
  utilities with file state); one execution environment and one epoch
  of models; the measurable requirements — the unmeasurable ones
  ("make it beautiful") are honestly shelved by the machine into a
  taste queue for the human; the API prices are volatile. The transfer
  to the M/L bands, the long tasks, and the other product classes is
  the subject of the next experiments.

## 6. Conclusions

1. An external transactional state machine moves readiness from an
   opinion to a signed, re-verifiable fact: 50/50 machine verdicts on
   a 25-task battery at two model classes (hypothesis 1).
2. The cycle's economics is controllable by systemic properties: −49%
   of tokens between the two full runs of the same battery without a
   change of methodology (hypothesis 2); the largest classes of losses
   (the byte retranscription, the noisy questions, the environment's
   resolution) are eliminated as classes.
3. The brief's freedom of interpretation is a measurable and the
   principal cause of the executors' divergence (78% of the judge tables' red
   rows at a green functional core); the self-reports do not reveal it
   (hypothesis 3).
4. The delivery quality is invariant to the model class in the closed
   loop: both system hands — 25/25 verdicts; flash delivers the same
   result for twice the time (slower in 24 of the 25 tasks,
   p = 7.7·10⁻⁷, at an equal median token price) and roughly
   one-eighth the money (≈ $13.5–16 against ≈ $111–131 per battery);
   the cheapest hand with a guarantee is the cheap model in the system;
   the most expensive guarantee is the manual finishing of solo
   (hypothesis 4).
5. The dominant factor of a single task's cost is the executor's
   variability, not the model class: the token spread inside every
   hand (0.60–7.66 / 0.66–10.05 / 0.39–5.46 M) is larger than the
   difference of the medians between the hands; the system gates the
   quality, the cost remains the executor's property.
6. The intermediate human attention between the wish and the verdict
   for the S band is reduced to zero (74 forks closed by journaled
   defaults) without losing the machine quality (hypothesis 5).
7. The machine's own overhead is itself gated: the submission latency
   and the ceremony share stay under the budget in every run
   (hypothesis 6).
8. The results replicate on the evolved machine (v0.1.x → 0.2.0,
   the September campaign): 75/75 products green by the blind
   double-run measurer, 50/50 + 20/20 machine verdicts, the judges'
   protocol-red pattern reproduced (solo 87/336 strict rows), and the
   full-path economics made explicit — with the verifying operator
   priced in, the machine path is the cheaper one (solo $6.07 against
   top $1.18 per task). Conclusions 1–7 hold beyond the original
   binary and price epoch.

## 7. Prospects

- **The M/L bands**: extending the battery upward (the large tasks,
  the multi-file products) after calibrating the bands by the
  accumulated measurements; the hypothesis to check — the freedom-of-
  interpretation share grows with the task's size, and the canon's
  benefit grows faster.
- **Bootstrap**: running the machine's own development through its own
  loop — the wishes for the machine as briefs, the loop to the
  verdict, the pains recorded and cured through its own journal; a
  direct test of the approach's scalability beyond the S band.
- **Multi-hand and multi-model**: the three-hand design generalizes to
  N hands with a judicial reconciliation against the best table — as
  the standard acceptance procedure of generative delivery.
- **The unmeasurable remainder**: the taste and aesthetic requirements
  need a human; the prospective direction — the formalization of the
  "taste queue" and safe defaults for it.
- **Reproducibility**: all the experiment materials — the briefs, the
  task materials, the world images, the methodology, and the full
  tables — are in this repository (`campaigns/`, `docs/SAMPLES.md`,
  `docs/SCENARIOS.md`, `docs/METRICS.md` — the replication waves'
  verified registry), which allows the runs to be
  repeated independently.

## Data sources

All the paper's numbers are taken from the runs' machine evidence and
summarized in the repository's documents: `docs/SAMPLES.md` (the
battery and the band calibration), `docs/SCENARIOS.md` (the field
coverage of the features), `campaigns/` (the briefs, the materials,
the generator, the world images, the run methodology). The per-task
data of every run is duplicated in Appendices A–D below; the
replication's full-path breakdown is Appendix E. The September
replication (§4) is documented in `docs/METRICS.md` — a
provenance-first registry (three evidence tiers: the instance-ledger
telemetry, the cell captures at each hand's finish, the platform
session-database ledger) whose every number is mechanically
revalidated against the run evidence. The artifact — the machine's
sources, the campaign methodology and the verified metrics registry —
is at https://github.com/neurophant/punchtape. Manuscript preparation
was assisted by AI tools under the author's full responsibility; the
execution of the runs by AI agents is the subject of the study itself
and is documented in §2–§4.

## Conflicts of interest

The author declares no conflicts of interest.

## References

1. Junwei Liu, Kaixin Wang, Yixuan Chen, Xin Peng, Zhenpeng Chen,
   Lingming Zhang, Yiling Lou. Large Language Model-Based Agents for
   Software Engineering: A Survey. arXiv:2409.02977, 2024 (revised
   2025).
2. Chunqiu Steven Xia, Yinlin Deng, Soren Dunn, Lingming Zhang.
   Agentless: Demystifying LLM-based Software Engineering Agents.
   arXiv:2407.01489, 2024.
3. Cheng Cheng. Detect–Repair–Verify for Securing LLM-Generated Code:
   A Multi-Language Empirical Study. arXiv:2603.00897, 2026.
4. Ran Wei, Le Zhu, Haochi Wang, Jim Woodcock, Fang Yan, Simon Foster,
   Xiangyang Ji. Formal-Method-Guided Vibe Coding: Closing the
   Verification Loop on AI-Generated Safety-Critical Software Through
   Model-Driven Engineering. arXiv:2606.22413, 2026.
5. Jiaru Qian, Yihong Dong, Yongmin Li, Hao Zhu, Bin Gu, Ge Li.
   Self-Spec Verifiable Code Generation. arXiv:2609.39568, 2026.
6. Yunhao Liang, Chengguang Gan, Ruixuan Ying, Hanjun Wei, Zhe Cui,
   Shiwen Ni. Security Tests as Executable Specifications for LLM Code
   Generation: Benefits, Trade-offs, and Coverage Limits.
   arXiv:2608.09740, 2026.
7. Yuyao Ge, Lingrui Mei, Zenghao Duan, Tianhao Li, Yujia Zheng,
   Yiwei Wang, Lexin Wang, Jiayu Yao, Tianyu Liu, Yujun Cai, Baolong
   Bi, Fangda Guo, Jiafeng Guo, Shenghua Liu, Xueqi Cheng. A Survey of
   Vibe Coding with Large Language Models. arXiv:2510.12399, 2025.
8. S M Jamil Uddin. Is Vibe Coding the Future? An Empirical Assessment
   of LLM Generated Codes for Construction Safety. arXiv:2604.12311,
   2026.
9. Zixiao Zhao, Amirreza Esmaeili, Fatemeh Fard. Bias in the Loop:
   Auditing LLM-as-a-Judge for Software Engineering. arXiv:2604.16790,
   2026.

## Appendix A. The baseline run, per task (2026-08-30)

The wall is the executor's time by the ledger; the submissions are
accepted/rejected.

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
| **Summary** | **median 17.3** | **219/53** | **25** | **80.31** |

## Appendix B. The repeat full run, per task (2026-08-30)

The format: the baseline → the run after the wave of systemic fixes.

| Task | Wall, min | Tokens, M |
|---|---|---|
| conduit-python | 29.0 → 14.8 | 5.07 → 0.80 |
| conduit-go | 22.5 → 11.7 | 4.20 → 0.78 |
| conduit-rust | 20.0 → 25.3 | 2.66 → 3.51 |
| conduit-cpp | 19.3 → 19.5 | 6.15 → 1.88 |
| conduit-js | 21.5 → 17.2 | 3.11 → 0.85 |
| pressmark-python | 13.6 → 11.0 | 2.14 → 0.98 |
| pressmark-go | 17.0 → 16.0 | 2.81 → 3.42 |
| pressmark-rust | 17.6 → 18.9 | 2.46 → 2.96 |
| pressmark-cpp | 26.0 → 18.6 | 7.02 → 2.83 |
| pressmark-js | 20.8 → 11.0 | 4.33 → 1.27 |
| errand-python | 18.6 → 13.5 | 1.92 → 1.20 |
| errand-go | 17.3 → 12.8 | 2.91 → 1.40 |
| errand-rust | 14.4 → 19.4 | 1.50 → 2.44 |
| errand-cpp | 15.6 → 11.5 | 3.13 → 1.19 |
| errand-js | 34.8 → 11.9 | 10.45 → 1.10 |
| brim-python | 17.2 → 9.8 | 2.05 → 0.69 |
| brim-go | 10.4 → 8.8 | 0.79 → 1.02 |
| brim-rust | 22.7 → 16.6 | 1.65 → 0.81 |
| brim-cpp | 15.3 → 11.9 | 2.24 → 1.75 |
| brim-js | 14.3 → 14.4 | 2.25 → 2.19 |
| quire-python | 11.7 → 11.3 | 1.62 → 1.02 |
| quire-go | 11.8 → 10.9 | 1.63 → 1.49 |
| quire-rust | 20.7 → 16.4 | 4.28 → 1.82 |
| quire-cpp | 14.8 → 16.7 | 3.01 → 2.59 |
| quire-js | 10.7 → 11.2 | 0.93 → 1.01 |
| **Summary** | **median 17.3 → 13.5** | **80.31 → 41.00** |

The wall grew for seven tasks: conduit-rust +5.3 and errand-rust +5.0
(the price of cargo builds), quire-cpp +1.9, pressmark-rust +1.3,
quire-js +0.5, conduit-cpp +0.2, brim-js +0.1. The token growth for
conduit-rust, pressmark-go/rust, errand-rust, brim-go, and quire-js —
the executors ran more probe sorties; all the values are inside the
tasks' budgets.

## Appendix C. The submissions and questions of the repeat full run

The format: the baseline → the run (the submissions accepted/rejected;
the questions).

| Task | Submissions +/− | Questions |
|---|---|---|
| conduit-python | 10/5 → 5/0 | 1 → 1 |
| conduit-go | 14/4 → 5/0 | 1 → 0 |
| conduit-rust | 14/3 → 11/1 | 1 → 0 |
| conduit-cpp | 12/5 → 6/0 | 1 → 0 |
| conduit-js | 6/5 → 4/0 | 1 → 0 |
| pressmark-python | 7/3 → 5/1 | 1 → 0 |
| pressmark-go | 7/3 → 4/0 | 1 → 0 |
| pressmark-rust | 7/2 → 5/0 | 1 → 0 |
| pressmark-cpp | 8/4 → 5/0 | 1 → 0 |
| pressmark-js | 11/1 → 4/1 | 1 → 0 |
| errand-python | 9/1 → 7/1 | 1 → 0 |
| errand-go | 11/1 → 6/0 | 2 → 0 |
| errand-rust | 6/1 → 6/1 | 0 → 0 |
| errand-cpp | 10/2 → 6/0 | 2 → 0 |
| errand-js | 14/0 → 4/1 | 1 → 0 |
| brim-python | 8/2 → 4/0 | 1 → 0 |
| brim-go | 5/0 → 7/1 | 0 → 0 |
| brim-rust | 5/1 → 5/0 | 1 → 0 |
| brim-cpp | 4/0 → 6/0 | 0 → 0 |
| brim-js | 11/1 → 9/1 | 2 → 0 |
| quire-python | 10/1 → 6/2 | 1 → 1 |
| quire-go | 6/3 → 6/0 | 1 → 0 |
| quire-rust | 10/1 → 5/0 | 1 → 0 |
| quire-cpp | 7/2 → 5/0 | 1 → 0 |
| quire-js | 7/2 → 6/2 | 1 → 1 |
| **Summary** | **219/53 → 142/12** | **25 → 3** |

## Appendix D. The three-hand run, per task (2026-08-31)

The wall is the executor's minutes; the tokens are M; the submissions
are accepted+rejected (the ledger of the system hands; solo has no
submissions); the judge is the strict count of green/all rows against
the top hand's checks table.

| Task | Top: min / M / subm | Flash: min / M / subm | Solo: min / M | Judge |
|---|---|---|---|---|
| conduit-python | 17.1 / 0.88 / 8+0 | 36.6 / 1.89 / 8+0 | 20.1 / 2.05 | 12/18 |
| conduit-go | 28.5 / 5.24 / 9+2 | 29.3 / 1.18 / 7+0 | 11.6 / 1.25 | 9/20 |
| conduit-rust | 15.8 / 1.15 / 5+0 | 37.7 / 2.70 / 4+2 | 17.6 / 2.77 | 4/18 |
| conduit-cpp | 33.8 / 4.99 / 14+2 | 49.1 / 5.91 / 9+1 | 15.8 / 1.59 | 9/25 |
| conduit-js | 33.4 / 7.66 / 10+3 | 41.6 / 3.15 / 4+0 | 18.3 / 1.21 | 0/16 |
| pressmark-python | 13.0 / 1.49 / 5+2 | 17.5 / 2.15 / 7+0 | 10.8 / 1.41 | 3/20 |
| pressmark-go | 13.9 / 2.79 / 7+0 | 42.4 / 4.55 / 6+4 | 10.8 / 0.84 | 8/23 |
| pressmark-rust | 14.8 / 2.39 / 6+0 | 27.0 / 2.72 / 5+0 | 16.2 / 2.02 | 6/20 |
| pressmark-cpp | 23.3 / 5.48 / 8+2 | 61.9 / 10.05 / 14+4 | 18.3 / 3.38 | 7/25 |
| pressmark-js | 13.6 / 1.73 / 6+1 | 30.8 / 2.32 / 9+0 | 11.3 / 1.26 | 8/30 |
| errand-python | 13.4 / 1.17 / 8+3 | 39.9 / 3.22 / 12+1 | 7.3 / 0.91 | 0/21 |
| errand-go | 21.8 / 2.41 / 5+2 | 20.8 / 1.58 / 5+1 | 13.8 / 1.25 | 0/11 |
| errand-rust | 24.1 / 2.39 / 8+1 | 32.0 / 2.35 / 7+1 | 11.2 / 1.79 | 0/24 |
| errand-cpp | 16.3 / 1.56 / 7+1 | 18.9 / 1.04 / 5+0 | 19.8 / 3.31 | 0/11 |
| errand-js | 14.4 / 2.23 / 7+0 | 31.6 / 1.55 / 9+2 | 9.1 / 0.99 | 0/10 |
| brim-python | 8.2 / 0.60 / 4+0 | 11.7 / 0.66 / 4+0 | 7.1 / 0.39 | 6/18 |
| brim-go | 10.3 / 1.10 / 7+0 | 23.2 / 1.71 / 4+0 | 9.8 / 0.88 | 2/19 |
| brim-rust | 12.2 / 1.77 / 5+1 | 22.6 / 1.41 / 5+2 | 19.7 / 3.04 | 0/24 |
| brim-js | 11.2 / 0.91 / 4+0 | 28.8 / 1.35 / 4+1 | 11.4 / 1.46 | 10/28 |
| brim-cpp | 14.1 / 3.51 / 6+2 | 19.1 / 2.43 / 8+0 | 9.1 / 1.20 | 6/23 |
| quire-python | 10.2 / 1.22 / 6+1 | 14.8 / 1.12 / 5+2 | 10.4 / 1.12 | 4/14 |
| quire-go | 14.1 / 1.94 / 5+3 | 20.3 / 1.84 / 6+0 | 11.2 / 1.16 | 3/20 |
| quire-rust | 25.1 / 3.91 / 9+0 | 28.9 / 3.58 / 7+2 | 17.2 / 2.47 | 6/14 |
| quire-cpp | 18.6 / 4.50 / 11+1 | 29.0 / 5.72 / 11+4 | 22.5 / 5.46 | 1/22 |
| quire-js | 14.4 / 2.29 / 7+3 | 41.2 / 6.60 / 13+5 | 9.5 / 2.01 | 4/17 |
| **Summary** | **435.6 / 65.31 M / median 14.4** | **756.7 / 72.78 M / median 29.0** | **339.9 / 45.22 M / median 11.4** | **108/491** |

The row order is the execution order (conduit → pressmark → errand →
brim → quire; the first row is the pilot task that took the hands'
budget measurements). The errand wave's conditional judge counts (the
neutral driver instead of the solo's missing CLI surface): errand-go
11/11, errand-rust 24/24, errand-cpp 8/11, errand-js 8/10 — at the
strict 0 for the whole wave; conduit-cpp by behavior 12/25 at the
strict 9/25. Pinpoint behavioral failures of solo against the top
tables, invisible to the hand's own self-checks: a connection cut by
RST instead of a clean close and the exit code on a taken port
(conduit-rust), a missing name field in the records (quire-rust), the
silent acceptance of a repeated option (pressmark-js), a "None" string
instead of an empty value (quire-python).

## Appendix E. The full path per task, the September replication: hand + its operator

What this table compares is the full price of reaching "done" on one
task — the executor's work and the verification work that necessarily
accompanies it. The operator column is the human-attention proxy: the
scarce resource the closed loop buys back is not the hand's tokens but
this column. Roles are never mixed: the hand's and the operator's
time/tokens are separate columns, the sum is analytical.

Bases: the whole solo side (hand and operator) is the wave-S uniform
tier — full-wall times and life tokens from the platform ledger, the
solo operators being the acceptance operators (errand-rust: its two
agents summed); the brim rows are the wave-1 cell captures; the
machine hands are the validation wave's cell captures (the full wall,
restarts included).

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

Reading: for solo the acceptance operator is a full role (273–937 s
over the ten tasks, up to 2.3M tokens on errand-python and 1.8M on
conduit-go), and the solo hand is the token-heaviest of the three
paths (its mean life is 5.95M tokens against the top hand's 2.78M —
without a machine that has already verified, the agent re-reads and
re-probes its own product); on the wave's machine paths the operator
barely works (0–88 s, fired in 7 cells of 20) — the machine took the
verification over. On the full path the machine top beats solo on
9 of the 10 tasks (the exception is errand-rust, the redis-integration
outlier); zero op for a machine hand means no question batch fired.
The hands' decision waits (23–216 s in the 7 cells) are not part of
elapsed — they are counted separately in each cell's capture.

### The operator column under a live human (the correction)

In every wave of the replication the operator role was executed by an
AI agent (GLM-5.3 Max / GLM-5.3 by the role tables of the wave
journals); the operator seconds recorded above are an AI's wall, and
for the human cost of the solo path they are a lower bound, for three
reasons: the speed of analysis (the AI re-reads the whole cell context
in seconds and fires its own probe runs between readings; a person
reads at reading speed, re-runs by hand, and re-derives the state of
the environment before each judgment); no setup, no fatigue, no
context switch (the AI's wall contains only the verification itself);
and the asynchrony of a live answer is unmeasured (the recorded
decision waits are the machine's side of the pause, not the person's
calendar time).

The measured anchor: the solo acceptance operator's full wall is
530 s (8.8 min) per task on average over the 20 non-brim wave-S tasks
(range 258–937 s). The live-human anchor for the same class of work —
the August finishing estimate, 4–7.5 hours of a live human per
25-task batch — is ≈10–18 min per task, i.e. 1.2–2× the AI wall,
before the asynchronous latency. Per 25-task batch of this class the
solo path therefore needs a verifying human cycle at every task
(≈4–8 hours by the human anchor, or ≈3.7 hours at the AI wall —
whichever operator stands in), while the machine paths fired an
operator in 7 of 40 wave-S cells and 7 of 20 validation-wave cells, at
cents per cell — on the remaining cells the human leaves the loop
entirely.

Honesty marks: the 10–18 min and the derived ≈4–8 hours are an
extrapolation from the documented estimate of the earlier series — a
labeled estimate, not a measurement of these waves; the 530 s mean is
the ledger recomputation; nothing else is extrapolated.
