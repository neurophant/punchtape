# CASE 0 — the 0.2.0 protocol under current conditions

The source 0.2.0 campaign's bytes (briefs, materials, prompts —
byte-for-byte) in the current host world: **one binary** (the sha is
in the series freeze), **one set of briefs**, **one accounting
platform** (db) and **one dual token base** (in+out AND
total=in+out+cache_read).

**Two hands (the owner's word):** in this case, as in the other
cases of this repo, there are TWO hands and only two — `solo`
(a top model without the machine) and `machine-top` (a top model
through the machine). There is NO third hand (`machine-weak` /
flash models) and there must be none: the 0.2.0 protocol is adapted
to this repo's reality, no flash model in any role. A wave =
meaning × 5 stacks × **2 hands** = **10 cells** + 5
judgments/measurements (one measurer and one judge per task).

Law: `campaigns/METHODOLOGY.md` — Part I + §0.2a–§0.2d + §I.11;
isolation §0.2: this campaign carries ITS OWN tools and ITS OWN
materials, no imports of other campaigns' paths.
Runs live in `~/.experiments/runs/<UTC>-case0-<meaning>`.

## The 0.2.0 protocol (the bytes are the source of the law; the hands follow the repo's reality)

The essence:

- **Wave** = one meaning × 5 stacks × 2 hands = 10 cells; stacks
  from heavy to light (rust → cpp → go → js → python); tasks
  SERIALLY, one after another (0.2.0: "tasks go one stack at a
  time").
- **Two hands**: `solo` (a top model without the machine),
  `machine-top` (a top model through the machine). The
  hands'/operators'/judge's/measurer's prompts are the 0.2.0 bytes
  (`prompts/`); the host renders are named adaptations
  (`prompts-host/`).
- **Operator** — a separate micro-agent (not the driver): for the
  machine hand a quick choice by the recommendations
  (OPERATOR-MACHINE), for solo a demanding customer with full
  reading (OPERATOR-SOLO). The response protocol: no one polls and
  no one sleeps — a hand's question ends its turn, the broker wakes
  the operator for exactly one answer.
- **Measurement**: a blind measurer (probes×2 of every product by
  one procedure; deliverables 0–5 by one counter; the method line)
  + a judge (a row-by-row table of the top hand's etalon table
  against the other product, byte-strict, doubt counts against the
  product).
- **Accounting**: METRICS UNIFORMITY — every role carries the same
  set of fields; hand/operator/layers are NEVER mixed into one
  column; full invariance of metrics (nothing is lost). The table
  structure is unified: the columns solo | top and the solo/top
  ratio are printed by `tools/report.py`.

## ADAPTATIONS (each named; the 0.2.0 bytes are the source, the deviations are only these)

| name | where | essence |
|---|---|---|
| The host-cell world | tools/stage-cell.sh, verify-cell.sh, prep-cell.sh | the 0.2.0 docker containers → host cells: the cell = an `input/ env/ work/` directory under the RUN ROOT's `cells/` (two-level random names — the measurement's blindness is preserved; the zone structure law: everything of a run lives in its root, the zone root carries only runs/tmp/cache); NO docker (the docker files were not carried over); the network — rule + post-factum audit (a copy of audit-net.sh); world-lock.txt — the source of env/VERSIONS |
| The harness render | prompts-host/HARNESS-HOST.md | the docker-exec mechanics of HARNESS.md → host mechanics (`__WORK__`/`__INPUT__`); the render is tools/make-host-prompts.sh with exact anchors (a source byte drift = refusal) |
| The contract paths | prompts-host/CONTRACT-MACHINE-HOST.md, OPERATOR-MACHINE-HOST.md | `/run/task` → `__WORK__`/`__INPUT__` (the world changed, the contract changed by not one word more) |
| The measurer and the canon | prompts-host/MEASURER-HOST.md, tools/case-runner-host.py | containers → fresh directories; the corpus = `.punchtape/canon/atoms/*.yaml`; the runner is the campaign's own copy; rows are replayed by the machine's grammar (seeds/run/out/err/rc/state, launch: nonet, tree: unchanged, witness, repeat-run) |
| Probes as staged bytes | materials/probes/, stage-cell.sh | the nonet/tree probes ride into every hand's `input/probes/` AS BYTES (the witness mechanics); the machine hand's spec calls them by witness rows; the rows' form pins carry `form:` citations from the 0.2.0 briefs VERBATIM (the machine blocks unpinned pins) |
| Accounting from the platform db | tools/usage-ledger.py, collect-metrics.sh, report.py | accounting is the platform db (~/.zcode/cli/db/db.sqlite) instead of model-io records; the hands/operators/layers tables of unified structure; printing — the DUAL token base (in+out AND total=in+out+cache_read); the registry roles: hand \| operator \| judge \| measurer; the micro-agents are pinned by session-pins.txt (spawned from the driver's sterile cwd) |
| Shared environment caches | tools/env-prep-check.sh, prep-cell.sh | the master environment caches are the PLATFORM's zone (`~/.experiments/cache/env-prep`, METHODOLOGY §0.4); the campaign tool only CHECKS (warming = platform maintenance, not a campaign action) |
| The model map | models.json, prices.json | the platform map model_id → tariff ONLY for this campaign's hands: GLM-5.3 → Max (both hands). The campaign files are named adaptations, NOT byte copies: the flash tier is removed from both the map and the price list; the Max tariffs are the 0.2.0 bytes (checked — no basis-correction needed) |
| Two hands | tools/{prep-cell,stage-cell,broker,verify-cell,count-deliverables,report,selftest}.sh/.py, prompts-host/MEASURER-HOST.md, models.json, prices.json, README, METHODOLOGY §I.11 | the owner's word: "in this case as in the others there is and must be no third hand", "no flash". There are TWO hands (solo, top); a wave = 10 cells; the weak hand is removed from the tools (staging a weak hand REFUSES — a selftest RED gate; count-deliverables.py and prices/models are named adaptations); the measurer judges TWO products (P1/P2); the delta table — solo \| top \| solo/top |
| File hygiene | tools/canaries.sh, prep-cell.sh (work/ENV), stage-cell.sh (the dotfile shield), freeze.sh (the dotfile shield), prompts-host/HARNESS-HOST.md (the spawn walls) | the owner's word (AGENTS §14): "~/.experiments/ — the only directory allowed for campaign runs, THE ONLY ONE, everything else — FORBIDDEN COMPLETELY AND WITHOUT EXCEPTIONS" — NOT ONE campaign tool writes to the system /tmp, the repo, or any path outside the zone: all scratch is `~/.experiments/tmp` (`C0_SCRATCH_ROOT`), the canary baits live ONLY inside the run root (init refuses any target outside the zone; the task sources in the repo are READ-ONLY — baits in the repo's `briefs/`/`materials/probes` = litter in the repo AND a bypass into the cells' staging); staging and the freeze do not see dotfiles — a hidden stray file rides neither into the freeze nor into cells |
| The submission economy | tools/make-host-prompts.sh (the contract substitution) | the staged-bytes package with one freeze: the mandatory dry check before EVERY submission (`punchtape spec --check delta.yaml`, free, without an LLM — a rejected submission = a whole turn spent); FILES INSTEAD OF DIALOG (drafts/tables/deltas in the `drafts/` directory, the FIRST spec declares `build-ignore: ["drafts/"]` — a draft outside the accounted tree moves no digest and does not drop ratify on "tree moved"; the dialog — commands only, read your own large files by address: the hand's history rides every turn and is paid for every turn); the dead spelling kind: feature → a prose: patch (the machine speaks exactly two kinds spec\|run) |
| Stable rows | tools/make-host-prompts.sh (the contract substitution) | after the first accepted spec ALL submissions are SPARSE PATCHES by row address (replace/add/remove/run/prose; the machine rejects a full repeat of the table, a deliberate regrouping — only full-replace: true — the owner's law "the number of submissions does not affect the economy"); the order of accepted rows does not change (a repair replaces its row in place, new rows — at the END of the table; the platform cache rereads the same bytes, a repeat is many times cheaper) |
| Form hints | tools/form-hints.py, tools/{stage-cell,verify-cell,freeze}.sh, tools/check-staged-freeze.py, prompts-host/CONTRACT-MACHINE-HOST.md (the contract line) | the mechanical sidecar `input/form-hints.md` to the MACHINE cell ONLY (a deterministic grep over the cell's own brief: verbatim lines with numbers — flags/example blocks/exact citations — as sources of form: pins; the machine itself never sees the brief — agnosticism); an empty brief prints an explicit mark "no form-bearing terms", never silence; the sidecar's digest is frozen (freeze: form-hints = the same generator over the same brief bytes) and checked at staging (check-staged-freeze) and in verify-cell (the manifest: the hand-conditional field form_hints; a solo cell does NOT carry the file — parity: the hands' inputs differ ONLY by this file) |

The roles/agents are exactly the 0.2.0 set minus the dead weak hand
(hand, micro-operator, micro-judge, micro-measurer) + the driver
session; nothing beyond the set. There are NO per-hand limits (P8).
Permissions — ordinary only (§0.2d). The ceiling of 8 concurrent
subagents (§0.2c) under the serial order is honored by
construction.

## THE WAVE SESSION ORDER (an 8-slot pipeline)

> The parallelism ceiling/norm = 8 (AGENTS.md §13, METHODOLOGY
> §0.2c): the order is the old one — stacks heavy→light, within a
> task solo before top, the measurer+judge after the task's TWO
> hands; the execution is a pipeline under the ceiling: the next
> work in order enters the first freed slot, a free slot with a
> non-empty queue is forbidden. **THE DRIVER TURN DISCIPLINE (the
> owner's word, "every time"): a driver turn does not end until
> (1) EVERY subagent completion has been processed — including one
> that arrived in a batch together with service reminders
> (processing completions is the FIRST action of the turn), and
> (2) every slot with READY work is occupied; a slot without ready
> work (a task's measurer/judge before both hands finish, an
> operator before the hand's turn) is not filled forcibly — the
> pipeline waits for the unblock; idling with a queue of ready work
> = a red line.** The 0.2.0 protocol's strict seriality ("one cell
> at a time") is replaced in this repo by this law; a change of the
> wave regime is fixed in the wave journal and in the report's
> analytics.
>
> **THE DRIVER'S NON-BLOCKING LOOP H1–H5 — MANDATORY TO EXECUTE**
> (the owner's word; the detailed form is METHODOLOGY §0.2c, the
> law is AGENTS.md §13).** H1: in mid-wave NOT ONE blocking wait on
> an agent >120 s — a non-blocking poll of all active agents in one
> batch every 60–90 s, processing EVERY completion as the turn's
> first action, a slot with ready work — immediately; in the
> report's analytics — the line "driver idles: 0" or a named red
> line; protocol blocks (a layer before the two hands, an operator
> before the hand's turn, the judge-etalon before the top's
> verdict) are not idling. H2: a tool's usage is read BEFORE the
> first call; acceptances are passed on ONLY as a whole file (no
> head -c). H3: registration strictly right after the spawn, with
> the real agentId, with the --session pin; a follow-up
> reconciliation. H4: commands longer than ~2 minutes — in the
> background with a notification only; the main thread stays free
> for processing completions and slots. H5: idling with a queue of
> ready work >90 s — the driver itself enters it into
> redlines.jsonl BEFORE the owner's prod. A violation of any H is a
> red line of the registry.

0. **STERILE ROOT** (AGENTS.md §14, the owner's word): the
   driver's FIRST action creates the root
   `~/.experiments/runs/<UTC>-case0-<meaning>`, with the
   driver-cwd inside it; ALL commands by absolute paths. The
   subagents write ONLY into `~/.experiments/` (the spawn walls:
   the ENV prefix on every command + WRITE BOUNDS = the cell). The
   wave prompt is self-executing: the session does everything
   itself, the owner only pastes the prompt.
1. The root `~/.experiments/runs/<UTC>-case0-<meaning>`;
   `uuidgen` → `$RR/driver-uuid.txt`; `tools/env-prep-check.sh`
   (COMPLETE).
2. The wave's canaries: `tools/canaries.sh init
   "$RR/measurement/sealed" "$RR/tasks" "$RR/operators"
   "$RR/manifests"` — targets ONLY inside the run root (the zone
   law, the owner's word: "~/.experiments/ — the only directory
   allowed for campaign runs, THE ONLY ONE, everything else —
   FORBIDDEN COMPLETELY AND WITHOUT EXCEPTIONS"; the task sources
   live READ-ONLY in the repo, no bait is planted there — init
   REFUSES by the zone law; staging and the freeze do not see
   dotfiles — a stray hidden file is not an input of the series);
   the registry is in the sealed zone; the intra-cell baits are
   laid by stage-cell itself (inside the cell, under `$RR/cells/`).
3. Over the stacks FROM HEAVY TO LIGHT (rust → cpp → go → js →
   python), tasks SERIALLY; within a task the hands solo → top (one
   at a time — there is no third hand):
   `C0_DRIVER_UUID=$(cat $RR/driver-uuid.txt) bash tools/prep-cell.sh
   "$RR" <sense>-<stack> <hand> --class main` (hand = solo | top;
   prep = stage + env + VERSIONS + env-check; SMOKE cells —
   `--class smoke`, outside the series data); `bash tools/verify-cell.sh
   "$RR" <cell>`; the watcher: `bash tools/rollout-watch.sh --out
   "$RR/rollouts" --map <KEY=ID file> &` (the key = the cell;
   cycle-free).
4. Spawning a hand: the message = the BYTES of
   `tasks/<cell>/SPAWN.txt` (composed by the staging from the
   frozen prompts); registration — STRICTLY RIGHT AFTER the spawn
   (H3: the real agentId, one call, the `--session` pin, a
   follow-up reconciliation):
   `bash tools/register-agent.sh "$RR" <cell> hand <full-uuid> --session <sess>`.
   NO second hand in a cell; an agent's death = a clean session
   restart (§0.2a).
5. A question batch/escalation (machine) or a question/checkpoint
   (solo): the hand's turn ends with a message; `bash
   tools/broker.sh "$RR" <cell> ask "<question>" [--recommend
   "<recommendation>"]` → the operator's prompt (a micro-agent; the
   solo operator = OPERATOR-SOLO by bytes) → `broker.sh ... answer
   "<decision>"` (the wait is journaled; the decision — the hand's
   next message verbatim). The operator registers via
   `register-agent.sh "$RR" <cell> operator <uuid> --session <sess>`.
6. The hand's finish: the watcher's snapshot — the last one; `bash
   tools/finish-cell.sh "$RR" <cell> --accept "<acceptance note>"`
   (solo) or `--verdict-from-machine` (machine; the operator's
   acceptance against the brief is still mandatory — the verdict is
   necessary, not sufficient).
7. After the task's two hands: **the judge and the measurer**
   (layers, separate micro-agents): the measurer — the prompt
   `prompts-host/MEASURER-HOST.md` with the paths of the TWO
   products (P1 `solo-product/check.sh`; P2 —
   `tools/case-runner-host.py --product <work> --input <cell>/input`);
   the deliverables — `tools/count-deliverables.py`; the result —
   `$RR/measurement/measure-<cell>.json` for each hand. The judge —
   the prompt `prompts/JUDGE.md` (bytes) with the top hand's etalon
   table (the `punchtape report --full` render from its instance)
   and the other product (the solo hand's);
   `$RR/measurement/judge-<cell>.json`. Both registered with
   `--session`. Spawned AFTER the task's hands finish.
8. Closing — with ONE command: `bash tools/close-wave.sh "$RR"`
   (usage-ledger → collect ×10 → canaries → sweeping its own canary
   bait (registry + name pattern) → audit-net → the report →
   stopping the watchers → the red lines summary). The only write
   zone is `~/.experiments` (AGENTS.md §14): neither a tool nor a
   hand writes to `/tmp` or anywhere else; the hand's environment
   is `work/ENV` (the prefix of every command:
   `cd work && . ./ENV && sh -lc '…'`, the "File hygiene"
   adaptation). The report:
   `$RR/wave-report.md` (the dual token base,
   the §0.2b.3 ratios, the UNIFIED HAND TABLE
   — in/out separately, $ cache-aware,
   submissions, verb-wall, verdict, cases, suite×2, deliverables,
   prose, coverage, LOC, operator $/decisions/wait — PER TASK as a
   separate table + a SUMMARY over all hands of all tasks;
   the §I.10 METHODOLOGY contract, printed by report.py) + RIGHT
   AFTER THE DELTAS (§6) THE CAMP
   TABLE: all hands, the hand+its-operator sums by time and tokens
   on both bases, the solo/top ratios and the G1 verdict per task +
   Σ-rows (the §I.10
   METHODOLOGY contract, printed by report.py); the analytical
   layer — as a separate file, glued by concatenation into
   `docs/case0-<datetime>-<meaning>-session-report.md`.
9. Commit+push; the handoff `./handoffs/<datetime>-case0-<next-
   task>.md` + the COPY-PASTE PROMPT (§0.2a; the template below).

## The series freeze (P6 — BEFORE the first cell)

```
bash tools/freeze.sh s ~/.local/bin/punchtape --out FREEZE-S.sha256
```

ONE sha of the series binary; the METHODOLOGY digest inside.
stage-cell refuses without a fresh freeze and checks every staged
input against it (`check-staged-freeze.py`). An edit of the series'
tools/texts — regenerate the freeze BEFORE the first cell (the
series has not started yet); after the first cell a NEW freeze = a
new series (P6).

## Selftest and smoke

`bash tools/selftest.sh` — the tooling gates (RED-DEMO included):
the prompt renders converge, the brim-cpp-top cell staging
(class=smoke, OUTSIDE the series data) stages and verifies against
the freeze, the weak hand is REFUSED by staging (the two-hands
gate), preflight bites, the corpus replayer is green×2 on synthetics
and reddens a broken product, the accounting pipeline (registry →
db ledger → collect → report) on fixtures prints both bases and the
ratios. The smoke cell does NOT spawn agents and does NOT solve a
task — it is a dry run of the mechanics.

## HANDOFF + COPY-PASTE PROMPT (§0.2a; template)

After the wave session closes — the handoff
`./handoffs/<datetime>-case0-<next-meaning>.md` (the results, the
state, the next session's order) + a READY prompt as a code block
(the owner copies it into the new session whole):

```
Continue punchtape — WAVE case-0: meaning <meaning> (protocol
0.2.0, serial, 10 cells + 5 judgments; TWO hands solo/machine-top —
no third/flash hand, the owner's word). Input:
campaigns/case-0/README.md ("THE WAVE SESSION ORDER"),
campaigns/METHODOLOGY.md (§0.2a–d, §0.2c the 8-slot pipeline
+ H1–H5, ordinary permissions), the fresh handoff handoffs/<latest>.
Order: README steps 0–9 (a sterile cwd; prep over the stacks
rust→cpp→go→js→python serially; the hands within a task solo → top,
one at a time; spawn by the SPAWN.txt bytes + registration RIGHT
AWAY with --session; the broker for questions; finish; the
measurer+judge after the two hands; close-wave with one command;
report+registry+handoff+copy-paste prompt). The driver discipline —
H1–H5 are MANDATORY (METHODOLOGY §0.2c, AGENTS §13): a
non-blocking poll of all active agents every 60–90 s, no blocking
waits >120 s, processing every completion as the turn's first
action, a slot with ready work — immediately, acceptances as a
whole file, the long (>2 min) — in the background, idling with a
queue >90 s — itself into redlines.jsonl. The laws: isolation §0.2
(everything its own), no docker/limits/permission games, no flash,
an agent's death = a session restart, red lines — through
redlines.jsonl with a disposition. The series binary — from the
FREEZE-S freeze (see the fresh handoff).
```
