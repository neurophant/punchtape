# THE PUNCHTAPE FIELD CAMPAIGNS METHODOLOGY

Status: the acting law of punchtape field runs. Scope: any field
runs. It does not depend on task meanings or stacks. A change to the
methodology is a new edition, frozen before the first cell of the
affected series (P6).

---

# 0. THE EXPERIMENT AND THE ISOLATION LAW

## 0.1. The kind of experiment

A self-sufficient folder in `campaigns/` — a full run of the
implementation: detailed briefs; two hands (machine/solo) build the
product whole; the economy and quality of camps is measured. The
current campaign is `campaigns/case-0/` (the 0.2.0 protocol under
current conditions, a comparative pair on one binary). The entry
point is the campaign README.

## 0.2. The campaign isolation law (the owner's word)

- Every campaign carries ITS OWN tools and ITS OWN materials; no
  imports of another campaign's paths.
- What campaigns share is only what lies outside their folders: the
  platform metrics base and this methodology; hidden measurement
  materials lie INSIDE the campaign's folder.
- Any campaign is run, deleted or rewritten at choice, without
  touching another. The isolation check is a run of the campaign's
  entry gate from its own paths.

## 0.2a. THE SESSION LAW (the owner's word)

Task runs are executed only so:

1. **One session — one task (meaning) in all its stacks.** The
   session starts with a clean driver context; no second task enters
   it.
2. **The death of any agent = a clean session restart** (tear down
   the run root, start the task anew in a new session). The
   session's agent census is fixed (§I.6a).
3. **The end of a session = a full collection of metrics over the
   task and its hands/stacks without loss of information** — the
   closing report is final and is never recalculated later.
4. **HANDOFF**: after the session closes,
   `./handoffs/<datetime>-<kind>-<next-task>.md` is written (outside
   git): the results, the state, the order of the next session.
5. **COPY-PASTE PROMPT**: the handoff AND the session's final
   message end with a READY prompt for the next session — a code
   block the owner copies into the new session whole; the prompt is
   self-supporting (the law + the methodology + the exact run
   commands of the next task). The prompt templates are fixed in the
   campaign README. A run without this ritual is not counted as
   closed.

## 0.2b. THE VICTORY CRITERIA LAW (the owner's word)

Unified victory criteria of the machine over solo — for any series
and battery, at any level of aggregation (task, wave/meaning,
series, battery):

1. **VICTORY OVER THE SOLO OPERATOR** (the primary criterion; G1):
   the token AND time metrics of the machine camp (the hand with the
   machine + its operator) are smaller than the metrics of the solo
   camp (the solo hand + its operator): tokens_M < tokens_S AND
   seconds_M < seconds_S.
2. **FULL VICTORY OF THE MACHINE** (G2): the machine hand's metrics
   WITHOUT counting the operator's metrics are smaller than the solo
   hand's metrics without the operator's: tokens_M < tokens_S AND
   seconds_M < seconds_S.
3. **THE VICTORY RATIO** is printed next to every comparison, for
   each metric separately: time (seconds_S / seconds_M) and tokens
   (tokens_S / tokens_M) — "by how many times"; values < 1 (a
   defeat) are printed just as honestly, the ratio is not rounded
   toward the favorable side.

Quality is a precondition of applicability, not part of the
criterion: an economy comparison is valid only with proven quality
(etalon/expectation rows green + judges); a victory with unproven
quality is not read as a victory. Numbers come only from mechanical
sources (the platform base, ledgers, the run journal); not one
number by hand.

## 0.2c. THE AGENT PARALLELISM CEILING = 8 AND THE FILL NORM (the owner's word, FOREVER)

No more than EIGHT concurrent subagents at any moment of any
session — hands, judges, calibrations: all agents count. Exceeding
the ceiling at any moment = an invalid session (stop, tear down the
root, start the task anew).

THE FILL NORM (the owner's word: "parallel subagents must exist and
free subagent slots must not idle", "up to 8… now too"): while the
session/wave has work (the hands of the next cells, judges,
measurers, operators), the driver keeps subagents working to the
norm — a free slot with work available is forbidden just like
exceeding the ceiling. The order of works (stacks heavy→light;
within a task solo before top; judges after the etalon) is preserved
as the ENTRY ORDER into slots; the execution is a pipeline: the next
work in order enters the first freed slot. A wave (a meaning = all
hands + judges) goes as a pipeline under the ceiling: the hands of
the next tasks and the judges of finished tasks enter freed slots,
never exceeding the ceiling and never leaving a slot empty while the
work queue is non-empty. The ceiling is fixed in the series freeze
(Appendix B, item 6).

**THE DRIVER TURN DISCIPLINE (the owner's word, "every time", in
force in all subsequent tasks, waves and runs):** a driver turn does
NOT end until BOTH conditions hold: (1) EVERY subagent completion
that arrived by that moment has been processed — including
completions that arrived in a BATCH together with service/system
reminders (processing completions is the FIRST action of the turn,
before any other business); (2) every slot for which ready work
exists (not blocked by protocol dependencies) is occupied. A slot
without READY work is not filled forcibly — the pipeline waits for
the unblock (for example, a task's judge and measurer are not
spawned until both hands finish), and that is not idling; idling is
only a slot with a non-empty queue of ready work. Every fact of
idling with a queue is a red line of the run registry.

**THE DRIVER'S NON-BLOCKING LOOP H1–H5 — MANDATORY TO EXECUTE (the
owner's word: "filling the slots in time and the non-blocking loop
are the driver's behavior; document it in the methodology in detail
as mandatory to execute"; also fixed as law in AGENTS.md §13 and the
case-0 README).** The operational form of the driver turn
discipline (see above): five guarantees in force in EVERY wave,
every task, every run of any campaign of this repo — always, without
exceptions and without owner prodding. A violation of any H is a red
line of the run registry (H5 — a self-reporting entry BEFORE the
owner's prod).

- **H1 — THE NON-BLOCKING LOOP (the core of the guarantee).** In
  mid-wave there is NEVER a blocking wait on an agent longer than
  120 s. The driver's loop: a non-blocking poll of ALL active agents
  in one batch every 60–90 s + processing of all notifications that
  arrived during the loop; completions that arrived in a BATCH
  together with service/system reminders are processed WHOLE and
  FIRST (before any other business of the turn). A slot with ready
  (not protocol-blocked) work is filled IMMEDIATELY after its
  unblock. The proof in a run is the wave journal's timestamps:
  "completion → processing+spawn" ≤ 90 s at every event; in the
  report's analytics — the line "driver idles: 0" or a named red
  line. Protocol blocks are NOT idling: a layer (measurer/judge) is
  not spawned until both hands of its task finish; an operator —
  until its hand's turn; the judge-etalon — until the top hand's
  verdict; a free slot with an empty queue of ready work is not
  filled forcibly.
- **H2 — USAGE BEFORE THE CALL, ACCEPTANCES AS A WHOLE FILE.**
  Before the FIRST call of every run tool (prep-cell, verify-cell,
  register-agent, broker, finish-cell, close-wave, canaries,
  rollout-watch, case-runner-host, count-deliverables) the driver
  has read its usage block from the tool itself. Acceptance texts
  and operator decisions are passed on (to the broker, finish-cell,
  the hand) ONLY as a whole file: a head -c byte cut breaks
  multi-line UTF-8 notes.
- **H3 — REGISTRATION STRICTLY RIGHT AFTER THE SPAWN.** Every
  subagent of the session is registered IMMEDIATELY after the spawn:
  the platform's real agentId, one register-agent call, the
  --session pin <the agent's platform session> (for a subagent —
  sess_subagent_<agentId>); after registration — a follow-up
  reconciliation of the uuid/session against the platform registry.
  The close-wave census gate catches both an unregistered agent and
  an unpinned session — but H3 means there is nothing to catch: the
  registry is complete at every moment of the wave.
- **H4 — THE LONG RUNS IN THE BACKGROUND.** Any command longer than
  ~2 minutes (close-wave, long builds/runs) is launched only in the
  background with a completion notification; the driver's main
  thread is always free for processing completions, broker answers
  and filling slots. A blocking foreground run of a long command in
  mid-wave = an H1 violation.
- **H5 — THE SELF-RED LINE.** Any fact of a slot idling with a
  non-empty queue of ready work longer than 90 s the driver itself
  enters into redlines.jsonl (class driver-discipline, disposition —
  the owner) BEFORE the owner's prod; the wave analytics carries a
  "Pipeline discipline" section with the idling line and a breakdown
  of every incident.
- **Carrying across sessions.** H1–H5 live not only here: every
  wave's handoff carries them verbatim (§0.2a.4–5), the next
  session's copy-paste prompt includes them as an order item, the
  live session handoff carries them as a GUARANTEES section. The law
  is eternal in the methodology itself, in AGENTS.md §13 and in the
  campaign README.

## 0.2d. THE PERMISSION-GAMES BAN (the owner's word, FOREVER)

Ordinary permissions on the whole run tree and all artifacts: no
chmod a-w, chmod 700, chmod 660 or other forbidding permissions; no
file or directory of the run may require sudo or a permission change
to read, clean or delete. Integrity protection is by digest
manifests and the freeze (P9), NEVER by filesystem permissions. The
campaign tools are cleansed of forbidding chmod; a return of
permission games is a red line.

## 0.3. General principles

P1. **Realism.** The run reproduces real working conditions: the
brief/seed is living human language; the environment is an ordinary
computer with toolchains; no materials a real executor would not
have. The only difference between the camps is the measured variable
(the machine).
P2. **Scientific validity.** Frozen inputs (sha256 freeze
manifests); one binary per series; blind and procedurally identical
judges; mechanical etalons with double validation (gold — all green;
degradates — red exactly in their own rows); post-factum data
selection is forbidden; incidents are red lines of the registry.
P3. **Total isolation of hands.** There are no intersections between
hands: the filesystem (a cell per hand), materials, products.
Verified mechanically: the hands' input digests of a cell matched
byte-for-byte; canary tokens; transcript scans of a closed class of
patterns (other roots' paths, etalon/probe file names, freeze paths,
other cells' canaries).
P4. **Etalon-centricity.** One comparison point per task — the
hidden mechanical etalon; all of a task's hands are measured against
it by one procedure; the assessment is semantic at the judge's,
mechanical at the etalon's; surpassing the etalon is added value, not
a penalty.
P5. **The driver is one agent of the session.** Run control is
executed by one session agent (the driver), with frozen tools; the
same agent plays the operators under the frozen protocol; the human
is removed from the operator loop; LLM judgment by the operator is
forbidden wherever the decision has a mechanical form.
P6. **Immutability within a series.** A change to the methodology is
a new edition, frozen before the first cell of the affected series.
P7. **Zero tolerance for steering influences** (the purity law). Any
pin or nudge that steers the decision or its logic, in any input or
channel — automatic invalidity of the cell and the run. There are no
exceptions; concealment is an aggravating fact.
P8. **The ban on per-hand limits.** No limits of time, tokens,
turns or money per hand; the only numbers are mechanism parameters,
identical for all hands and declared in the environment card (probe
timeouts, the passport build ceiling). Ceilings over the measured
variable are forbidden (a numeric ceiling over the measured variable
distorts the experiment).
P9. **Sealing without forbidden permissions.** Archives are ordinary
permissions plus a digest manifest; "permission denied" has no right
to exist in the repository tree.

## 0.4. The common substrate

- **A session platform** with background subagents; the platform's
  accounting base (db.sqlite: model_usage by sessions —
  input/output/cache_creation/cache_read tokens and duration_ms; the
  truth of metrics, never self-report).
- **The single write zone of runs is `~/.experiments`** (the owner's
  word, uniform with the zone's structural law): `runs/ tmp/
  cache/` and nothing else in its root; no tool of a wave, campaign
  or hand writes to the system `/tmp` or anywhere else. During a run
  the subagents write ONLY into `~/.experiments/`: the driver's
  first action creates the run root and works from it; the spawn
  walls (the ENV prefix + WRITE BOUNDS) hold the hands' writes
  inside the cell whatever the session's starting cwd (the
  subagents' cwd = the session's launch directory — launching from
  the run root adds a wall over faulty commands too); the hands
  write only inside cells (the command prefix
  `cd <work> && . ./ENV && sh -lc '…'`, WRITE BOUNDS in the spawn,
  the cell's TMPDIR and caches — in ENV). The causes of littering
  are eliminated at the root (a sterile launch, ENV delivery, the
  /tmp ban for all tools), not papered over with an audit.
- **The hidden measurement materials live IN the campaign's FOLDER**
  (the owner's word: `~/.experiments/` is clean of anything
  permanent). They are never staged to hands; a leak of their
  contents into a hand's transcript is a red line.
- **The system punchtape binary**: installed into the system before
  the cells are staged, its sha256 pinned into the series freeze
  (`FREEZE-S.sha256`); a rebuild after the series starts is
  forbidden; the solo hand never calls the binary (parity — the
  absence of machine files in the cell for both hands).
- **The single run root**:
  `~/.experiments/runs/<UTC-timestamp>-<kind>-<name>/`, everything
  of the run lives inside; outside the root the run creates nothing.
  The structure of `~/.experiments/` (the owner's word): `runs/` —
  run roots (the only kind of thing written), `tmp/` — measurement
  scratch, `cache/` — restorable environment caches (env-prep).
  Nothing else is created in `~/.experiments/`; handoffs live in the
  repository root's `./handoffs/` (outside git).
- **The host world (without containers)**: the cell = an
  `input/ env/ work/` directory on the host under a single
  scratch-root; dependencies come from offline caches (env-prep:
  cargo/gomod/node_modules registries), delivered to the hand by the
  `work/ENV` file (the cell's exports; the prefix of every hand
  command: `cd work && . ./ENV && sh -lc '…'`); the outward network
  is closed by rule and audited post factum.

---

# PART I. CASE 0 — A FULL RUN OF THE IMPLEMENTATION (DETAILED BRIEFS)

## I.1. The experiment question

Working through the machine (punchtape) versus working without the
machine on IDENTICAL tasks with detailed briefs: does the machine
lower the cost of reaching a given level of quality at comparable or
better quality. H1: the machine is cheaper in operator time and
iterations at the same quality; H0: there is no difference.

## I.2. Roles

- **Hand** — the task's executor. TWO hands per cell:
  - **machine** — an agent working THROUGH the machine (spec
    submissions, runs, verdicts through the machine's door) + its
    own micro-operator (owner-micro);
  - **solo** — an agent without the machine working directly in
    `work/` + its own micro-operator (customer-micro).
  A camp = a hand + its operator (the cell economy).
- **Operator** — the session's driver playing both camps'
  micro-operators under frozen prompts: for the machine —
  owner-micro (short decisions on escalations, starting from
  `report --since` of the last verdict, a spot check of the delta),
  for solo — customer-micro (a living engineer: rereads
  plans/specs/code, checks against the brief, corrects in human
  language — the symptom and the requirement, WITHOUT naming the
  remedy). The stopping rule is unified: the operator's acceptance
  against the brief; for machine hands the machine's final verdict
  is necessary, not sufficient.
- **Driver** — the session's chief agent: cell setup, hand
  spawning, accounting capture, export, judges, tables. One process;
  the agent registry is closed (one cell = one long-lived
  agent-hand; the other agents are judges only, spawned AFTER the
  hands finish; an appearance beyond the list is a red line).
- **Judge** — the blind role of semantic assessment (I.7).
- **Steward** — the human role (the owner): ratifies editions,
  validity decisions, exclusions.

## I.3. Task inputs

All inputs are in the campaign folder:
`briefs/<meaning>/<stack>.md`, `materials/<meaning>/`, `prompts/`
(the result frame, spawn templates, the environment card, the
expectations list, JUDGE.md), `prices.json`, `models.json`,
`FREEZE-S.sha256`.

1. **The brief (detailed)** — one or two pages of living human
   text: what is wanted, why, how it must behave, the bounds of
   decent behavior (understandable errors, robustness, garbage on
   the input), qualitative performance wishes, the surface type.
   Forbidden: CLI grammars, exact diagnostic texts, test tables,
   examples of exact outputs, references to measurement tools, the
   order of work, the stack and versions. The customer's habit (the
   command name, flags, endpoints) and a method requirement
   ("measure by entropy") are part of the wish, equal for all hands.
2. **The cell configuration** (task × stack × size): facts only —
   the toolchain and versions, the complexity band, environment
   timeouts and quotas. A line nudging toward a solution is a P7
   violation.
3. **Materials** — neutral task data without spoilers; all fixtures
   (including the edge cases the brief names) are ready before the
   wave and laid out identically for every hand; second copies for
   probes are generated by the runner with the seed from the
   freeze.
4. **The result frame** — ONE text for all tasks/hands/waves:
   "done" = everything the brief requires is implemented; the
   implementedness of every requirement is shown by its own
   reproducible checks; error situations are handled decently; the
   frame does not say HOW to solve. Per-task frames are forbidden.
5. **The environment card** — one page next to the brief: where the
   checks run (in the cell's live tree), the self-cleanup of seeds,
   the tree sums before/after rows, the passport build ceiling
   300 s, acceptance by canonical probes, the question rules, the
   absence of limits.
6. **Prereg answers** — resolutions of brief ambiguities ratified
   by the steward, a frozen line, staged EQUALLY to ALL of a task's
   hands before the start.
7. **The input control** — two checks before the freeze on the
   brief/frame/materials: a mechanical one (dictionaries/templates
   of CLI pins, exact-output patterns, measurement tool names) and a
   blind semantic review (an independent reviewer without knowledge
   of the etalon looks for nudges and breaches of the measurement
   apparatus). Both records go into the freeze.

## I.4. The world and the cell

- The cell = `input/ env/ work/` under a single scratch-root;
  staged by the placer (`tools/prep-cell.sh` → `stage-cell.sh`): a
  copy of the brief/materials into `input/`, offline caches and
  VERSIONS.txt into `env/`, an empty `work/`; the digest of every
  copied thing is checked against the freeze; a staged manifest
  with the driver's uuid.
- The readiness env-check: the cell's toolchains raise their
  dependencies from the caches (a compiler/runtime launch with the
  task's dependencies).
- The binary is the system one (§0.4); the machine hand's stage
  refuses on a sha drift.
- Canaries: unique tokens per cell and reachability zones, laid out
  before the start; the closing scan looks for OTHER cells' token
  bytes in transcripts/files/exports.

## I.5. The run (the cell cycle)

1. **The spawn** of a hand with a spawn message (a unified
   structure: the role, the stopping rule, the input paths, the
   question channel — by the turn's final message; NOTHING more: no
   summary, no advice, no order).
2. **The hand's work**: only the brief, the materials, the
   environment. Product edits go through the regular channel
   (machine hands — submissions into the machine); a manual file
   edit is a red line. Questions and holds escalate to the
   operator; the hand never resolves them itself; a question does
   NOT block the cell.
3. **The operator** answers within the hand's window; every answer
   goes through the answer journal (situation–choice–grounds; the
   top of similar entries before accepting; contradicting earlier
   ones is not allowed) with the wave's standard reference (frozen
   defaults: diagnostic timestamps, exit codes, transliteration and
   the like). The answer's latency is journaled by frames as
   operator time.
4. **The hand's finish**: machine — the instance's final verdict +
   the agent's exit from processes + the operator's acceptance;
   solo — the operator's acceptance against the brief (the
   processes are stopped by the driver afterwards).
5. **The declared products**: machine — the trace manifest of the
   final accepted submission + the acceptance tree digest from the
   machine; solo — a list of paths against the brief, the digest
   computed by the driver. The declaration is the whitelist
   boundary and the only source of export.
6. **Observation frames**: at every turn — UTC, the turn
   identifier, a portion of platform accounting, the fact of an
   escalation. The truth of the economy is the instance ledger
   (machine) / transcripts (solo).
7. **Export**: the driver builds neutralized copies of the declared
   products (neutral names, without hand metadata, a token scan
   gate) into the measurement zone; the judge works only with the
   copies.
8. **The cell's closing**: the §I.8 checks (input digests,
   whitelist, canaries, transcript scan), the measure, the
   judgment, the snapshot.

### I.5.3a. Journal answers (the gate mechanics)

Every operator answer goes through the answer journal; a record's
question key is identical for all decisions of one cell, so a cell's
second and later records are added ONLY with `--note` naming the
difference from similar records (the duplicate-and-contradiction
gate cuts without a note). This is mechanics, not a driver choice.

### I.5.8a. The "etalon before judge" order (law)

The cell is closed → staging of the measurable product
(`stage-measurable.sh`) → the mechanical etalon (`run-etalon.sh`,
measure.json) → and only then the judge into the slot. Reconcile's
authorship invariant requires measure.json to be older than the
cell's judge registration; a violation is a data red line, not a
byte defect.

## I.6. Waves and the series

### I.6a. THE SESSION DISCIPLINE of case 0 (the owner's word)

Case 0 is run ONE TASK (meaning) PER SESSION (§0.2a): one session =
one meaning in all 5 stacks = 10 cells (5 stacks × 2 camps), the
session census = 10 hands + judges, no second task enters a session.
An agent's death = tear down the root and start the task anew. The
session's closing: a full collection of metrics over the task (all
cells, hands, camps — G1/G2 + judges) without loss of information,
the report into docs/; then the handoff
`./handoffs/<datetime>-case0-<next-task>.md` and the next
session's COPY-PASTE PROMPT as a code block (the template is the
campaign README). A battery = a chain of sessions over meanings; the
battery total is assembled from the session summaries.

### I.6b. Waves inside a session

A wave = a set of cells by one protocol; a series = waves with one
binary and inputs. Inside a session a meaning's cells go in waves by
the accepted schedule, UNDER THE §0.2c PARALLELISM CEILING; the cell
time ceiling is 1800 s (a wave parameter; an overrun is a red line,
not a kill). A cell's hands run in PARALLEL; the schedule is part of
the wave freeze. The wave-closing reconcile: starts = finishes =
snapshots = economy rows; the manifests recomputed; the red lines
dispositioned; the pilot/smoke markings in place (three layers: a
tag in the manifest, the raw path, a TSV column; they do not enter
the main pool).

### I.6c. The wave-closing pipeline (one command)

The wave is closed by `close-wave.sh <run-root>`: collect-cell ×N →
stage-measurable ×N → run-etalon → the measure/judge layout into the
blind export zone (anonymous copies by the sealed manifest) →
operator-ledger → economy.tsv → reconcile → STOPPING the wave's
watchers (the watchers' lifecycle is owned by the closing command —
no danglers). The operator's tokens are attributed EXACTLY
(`operator-ledger.py`): the driver session's db rows, sliced by the
frames' open→closed windows; a row in one window → to the cell
whole; in k≥2 windows → to the shared basket whole with the cell
list (no splitting); outside the windows → unattributed; the parts
== session total invariant is checked by refusal.

## I.7. Measurement

- **Etalons**: a checklist of functions and behaviors — semantic
  rows 1:1 to the brief; a row without a source in the brief does
  not exist; the bounds of decent refusal are semantic; probes are
  canonical influences through the product's surface. Double
  validation before the wave: a golden product (all probes green)
  and degradates (exactly one broken feature — exactly one red
  row); the probe→row map is signed and verified mechanically.
- **Etalon score**: the share of implemented rows and refusals
  (partial = 0.5 at the judge's; at the mechanical measurer — a
  mechanical criterion).
- **Judges** (the frozen prompt `prompts/JUDGE.md`): one micro-judge
  per cell, the FULL working tree (not the export), separately
  against the task's inputs; four axes 0–10 with a file:line
  anchor: SPEQ (spec quality), CODQ (code quality), SPEC (contract
  coverage by the spec), CODC (implemented and proven).
- **The cell economy** (TSV, by scripts; not one number by hand):
  submissions/turns, fix cycles, question batches and waits, wall
  time, operator time (answer latency from frames), tokens and $ by
  the platform base (cache-aware), calls, the etalon score, LOC,
  the productive share of tokens, the time to the first submission.
- **The series metrics** (the campaign metrics registry): G1 (camps:
  hand+operator, tokens and active seconds) — the "victory over the
  solo operator" criterion §0.2b.1; G2 (hands without operators) —
  the "full victory of the machine" criterion §0.2b.2; G3 (cell
  time against the ceiling), G4 (etalons by camps) + judges by
  axes. The victory ratio (§0.2b.3: time and tokens, by how many
  times) is mandatory to print next to G1 and G2 at every level of
  aggregation. The trust order of sources: agent metadata → turn
  notifications (it records EVERY one) → watcher snapshots (an
  upper bound only). Invariants: metadata == the last notification;
  Σ segments == lifetime; a divergence is a stop.
- **Statistics**: a two-stage preregistration (stage A —
  hypotheses/metrics/plan before the pilot; stage B —
  bars/thresholds before the first main cell); shares —
  Clopper–Pearson intervals; differences — bootstrap; multiple
  comparisons — the Holm correction. Carrying a solo row between
  waves — only with mechanically proven identity of all inputs and
  an edition of the law (the whole wave or not a single cell).

## I.8. Validity

A valid cell (a frozen checklist, all green): zero driver/platform
interventions into the hand's work; the hands' input digests matched
byte-for-byte; the file whitelist; canaries and transcript scans
clean; the judge and measurer worked with canonical probes; zero
manual product edits; zero hold decisions by the hand; not one
steering artifact; the answer journal consistent in meaning. A valid
run: all cells valid or excluded by name by the steward; the driver
delivers a mechanical validity report, the decision is the
steward's.

## I.9. The wave readiness checklist (before the first main cell)

1. The briefs/frame/materials passed §I.3.7; the verdicts are in the
   freeze. 2. The run root is created. 3. The etalons are written;
   every row has a probe; gold/degradates are told apart; everything
   is frozen by hashes. 4. The answer journal is open (append-only,
   outside the task roots). 5. The materials are validated against
   the toolchains; the environment caches are collected. 6. The
   series binary is installed and pinned; the prices are fixed.
   7. The agent registry is tried on a smoke (a second uuid →
   stop). 8. The prereg answers are ratified and frozen; the
   environment card is frozen. 9. The driver checklists are frozen
   tools. 10. Preregistration: stage A is signed before the pilot;
   stage B — before the first main cell; a smoke of 1–2 cells
   entirely green, the data marked, the bars immovable after the
   smoke. 11. The canonical runner and the economy TSV scripts —
   digests in the freeze.

## I.10. The task report: the structure contract and assembly

The numeric core of every task's report is generated by `report.py`
(the only source of tables): the sources of every number → cell
blocks → a table per task (machine vs solo, ratios) → the big table
(hand tokens/time; operator tokens/time; HAND+OPERATOR tokens/time;
quality in PERCENTS: etalon and judge) → camp Σ-rows → the S/M TOTAL
row with ratios and G1/G2 verdicts → validity. The structure is a
contract wired into the tool (a hard refusal on section drift); a
structure change = a visible commit. The final document = the
deterministic core + an analytical layer (human explanations of
every value and judgments), GLUED by concatenation; the core MUST
print the UNIFIED HAND TABLE — hand | time (active s) | tokens
in/out (separately) | hand $ (cache-aware by prices.json) |
submissions ±/− | verb-wall | verdict | cases | suite×2 |
deliverables | prose | coverage (with the method line) | LOC |
operator $/decisions/wait — as a SEPARATE TABLE PER TASK and a
SUMMARY TABLE OVER ALL HANDS OF ALL TASKS of the run (the owner's
word); RIGHT AFTER THE DELTAS — THE CAMP TABLE (§6): ALL hands, each
with ITS OWN operator, hand+operator sums by TIME and TOKENS on both
bases (in+out and total), the solo/top ratio for every metric and
the G1 verdict (§0.2b.1: smaller in BOTH tokens AND time on both
bases) for every task + camp Σ-rows over the whole run; not one
number by hand — only cell evidence and instance telemetry; the core
is generated into the run root and is never regenerated over the
glued document. The battery total is assembled from task reports of
identical structure without reformatting.

## I.11. The machine's economy

Requirements on the series binary (the case-0 campaign; the sha
is in the series freeze):

1. **The contract — once at birth**: the full contract is rendered
   to the hand only in the first reply to the instance's birth;
   afterwards a single reference line to the sha256 at birth;
   substituting/deleting the tree copy = a full re-render
   (protection, not an option).
2. **The door renders are lean**: the cell's canon is complete
   byte-for-byte (the judge/ledger/audit read the canon); to the
   hand — what changed since the last reply + the red rows in full
   + counters; `report --full` — the full canon report to the
   driver/judge.
3. **The lean door echo**: the reply to a submission = status +
   transaction + the accepted bytes' sha + the gates' names + the
   measured one-liners (+ the door-reset line on an accepted
   submission); the slot comes by a separate `next` (the cache
   makes it free); the echo stays lean — without a numeric pin (the
   owner's word).
4. **The form block before the run**: a spec row with a form pin
   without a one-line `form:` source = FORM DECISIONS UNRESOLVED,
   run submissions are refused until the source (brief/prereg
   verbatim) or a question to the operator; the machine does not
   judge the quotation.
5. **The ceremony metric**: every machine reply writes the render
   family into the cell's ledger (contract / slot / ack); the dump
   of the four categories (contract / spec / report / door-echo) is
   a deterministic reading of the canon. There are NO byte metrics
   and no bars (the owner's word): the economy is only tokens and
   seconds from the platform db; the ceremony volume, when needed,
   is read from the db messages.
6. **Total agnosticism**: there are no domain witnesses in the
   machine; probes arrive ONLY as staged bytes through the door
   (witness rows); the procedural modes of any row —
   `launch: nonet`, `tree: unchanged`; the eternal mechanical gate
   is the agnostic scan in `make test`.

The series accounting order: a dual token base (in+out AND total =
in+out+cache_read separately), the G1/G2 ratios are printed on both;
the economy gate: economy.tsv has no None in the main cells'
mandatory columns after the wave closes; basis corrections go as
`basis-correction` lines of the wave journal.

---

# APPENDIX A. The "known defect → mechanism" map

| # | A defect of past runs | The mechanism |
|---|---|---|
| 1 | Platform notification drops | The truth of liveness is processes/registry; a loss = a red line |
| 2 | Accounting truncation | Capture at every turn; the truth is db/ledger/transcripts |
| 3 | An agent's false death (rotation = death) | Rotation is routine; diagnosis only by processes |
| 4 | Double control of a cell | The uuid registry; the driver's stamp in the manifest |
| 5 | Sibling completion loops | Stop only by the full uuid |
| 6 | A homemade judge harness | The judge's environment is only the canonical runner and probes |
| 7 | Manual product edits | Only the regular channel; a red line |
| 8 | Hold decisions by the hand | Escalation to the operator |
| 9 | Defective materials | The two input checks before the freeze |
| 10 | Pins/hints in the inputs | The §I.3 bans; P7 |
| 11 | Delivery typos (corrections by resending) | A byte-for-byte check BEFORE sending; drift = stop |
| 12 | Agent deaths → respawns/carriers | A fixed census; death = a session restart |
| 13 | Compaction orphaned db attribution | Isolated sessions; a tiny driver context |
| 14 | Bytes in the main metrics; rounding drift | Only tokens/seconds; integers; sums of what is printed; an audit |
| 15 | Slot idling recorded as the norm | Zero tolerance: idling with a queue = a restart |
| 16 | Blind judges on an impoverished export | Judges: the full working tree |
| 17 | Form cracks of etalons against solo | The law of forms: forms are named by the wave texts |

# APPENDIX B. Parameters frozen per series

1. The series binary's sha (the pin). 2. Prices and models. 3. The
preregistration bars/thresholds (stage B). 4. The mechanism
timeouts (probe 60 s, etalon runner 900 s, case 0's wave ceiling
1800 s). 5. The passport build ceiling 300 s. 6. The series' agent
parallelism ceiling = 8 (the §0.2c law, forever; it enters every
series' freeze).
