# SCENARIO MATRIX — 75 tasks × scenarios × system features

A self-supporting map. Purpose: SEE WHICH OF THE 75 TASKS COVER
WHICH SCENARIOS AND SYSTEM FEATURES — so the tasks themselves improve:
uncovered scenarios are added to briefs, uncovered features to
the tasks' checks. Circles: ⚪ not covered · 🟡 partial · 🟢 covered
(the task was run by a campaign and the fact recorded).

Sections: 1 — all 75 tasks; 2 — scenarios (40); 3 — system
features (all, by group); 4 — the "task × scenarios" matrix; 5 — the
"task × features" matrix (tables per task meaning); 6 — governance rules;
7 — honest caveats.
The machine stays totally agnostic about the task world: rows are
arbitrary argv, seeds and materials are byte transfer — no
environment operations, no background processes; the feature numbers
below follow this composition.

## 1. All 75 tasks (code: meaning-stack-axis)

| Code | Meaning | Stack | Axis | Run | Status |
| --- | --- | --- | --- | --- | --- |
| br-py-S | brim | python | S | auto-run | 🟢 |
| br-py-M | brim | python | M | — | ⚪ |
| br-py-L | brim | python | L | — | ⚪ |
| br-go-S | brim | go | S | auto-run | 🟢 |
| br-go-M | brim | go | M | — | ⚪ |
| br-go-L | brim | go | L | — | ⚪ |
| br-rs-S | brim | rust | S | auto-run | 🟢 |
| br-rs-M | brim | rust | M | — | ⚪ |
| br-rs-L | brim | rust | L | — | ⚪ |
| br-cc-S | brim | c/c++ | S | auto-run | 🟢 |
| br-cc-M | brim | c/c++ | M | baseline | 🟢 |
| br-cc-L | brim | c/c++ | L | — | ⚪ |
| br-js-S | brim | js/ts | S | auto-run | 🟢 |
| br-js-M | brim | js/ts | M | — | ⚪ |
| br-js-L | brim | js/ts | L | — | ⚪ |
| cd-py-S | conduit | python | S | auto-run | 🟢 |
| cd-py-M | conduit | python | M | — | ⚪ |
| cd-py-L | conduit | python | L | — | ⚪ |
| cd-go-S | conduit | go | S | baseline, auto-run | 🟢 |
| cd-go-M | conduit | go | M | — | ⚪ |
| cd-go-L | conduit | go | L | — | ⚪ |
| cd-rs-S | conduit | rust | S | auto-run | 🟢 |
| cd-rs-M | conduit | rust | M | — | ⚪ |
| cd-rs-L | conduit | rust | L | — | ⚪ |
| cd-cc-S | conduit | c/c++ | S | auto-run | 🟢 |
| cd-cc-M | conduit | c/c++ | M | — | ⚪ |
| cd-cc-L | conduit | c/c++ | L | — | ⚪ |
| cd-js-S | conduit | js/ts | S | auto-run | 🟢 |
| cd-js-M | conduit | js/ts | M | — | ⚪ |
| cd-js-L | conduit | js/ts | L | — | ⚪ |
| er-py-S | errand | python | S | auto-run | 🟢 |
| er-py-M | errand | python | M | — | ⚪ |
| er-py-L | errand | python | L | — | ⚪ |
| er-go-S | errand | go | S | auto-run | 🟢 |
| er-go-M | errand | go | M | — | ⚪ |
| er-go-L | errand | go | L | — | ⚪ |
| er-rs-S | errand | rust | S | auto-run | 🟢 |
| er-rs-M | errand | rust | M | baseline | 🟢 |
| er-rs-L | errand | rust | L | — | ⚪ |
| er-cc-S | errand | c/c++ | S | auto-run | 🟢 |
| er-cc-M | errand | c/c++ | M | — | ⚪ |
| er-cc-L | errand | c/c++ | L | — | ⚪ |
| er-js-S | errand | js/ts | S | auto-run | 🟢 |
| er-js-M | errand | js/ts | M | — | ⚪ |
| er-js-L | errand | js/ts | L | — | ⚪ |
| pm-py-S | pressmark | python | S | baseline, auto-run | 🟢 |
| pm-py-M | pressmark | python | M | — | ⚪ |
| pm-py-L | pressmark | python | L | — | ⚪ |
| pm-go-S | pressmark | go | S | auto-run | 🟢 |
| pm-go-M | pressmark | go | M | — | ⚪ |
| pm-go-L | pressmark | go | L | — | ⚪ |
| pm-rs-S | pressmark | rust | S | auto-run | 🟢 |
| pm-rs-M | pressmark | rust | M | — | ⚪ |
| pm-rs-L | pressmark | rust | L | — | ⚪ |
| pm-cc-S | pressmark | c/c++ | S | auto-run | 🟢 |
| pm-cc-M | pressmark | c/c++ | M | — | ⚪ |
| pm-cc-L | pressmark | c/c++ | L | — | ⚪ |
| pm-js-S | pressmark | js/ts | S | auto-run | 🟢 |
| pm-js-M | pressmark | js/ts | M | — | ⚪ |
| pm-js-L | pressmark | js/ts | L | — | ⚪ |
| qr-py-S | quire | python | S | auto-run | 🟢 |
| qr-py-M | quire | python | M | — | ⚪ |
| qr-py-L | quire | python | L | — | ⚪ |
| qr-go-S | quire | go | S | auto-run | 🟢 |
| qr-go-M | quire | go | M | — | ⚪ |
| qr-go-L | quire | go | L | — | ⚪ |
| qr-rs-S | quire | rust | S | auto-run | 🟢 |
| qr-rs-M | quire | rust | M | — | ⚪ |
| qr-rs-L | quire | rust | L | — | ⚪ |
| qr-cc-S | quire | c/c++ | S | auto-run | 🟢 |
| qr-cc-M | quire | c/c++ | M | — | ⚪ |
| qr-cc-L | quire | c/c++ | L | — | ⚪ |
| qr-js-S | quire | js/ts | S | auto-run | 🟢 |
| qr-js-M | quire | js/ts | M | — | ⚪ |
| qr-js-L | quire | js/ts | L | baseline | 🟢 |

## 2. User scenarios — 40

Developer S1–S13 · Analyst S14–S19 · Tester S20–S26 · Operations S27–S34 · General S35–S40.

- S1 project from scratch
- S2 new feature
- S3 behavior change
- S4 feature removal
- S5 code backfill
- S6 find/reproduce a bug
- S7 fix loop
- S8 regression
- S9 another stack
- S10 changing the executor
- S11 pause/resume
- S12 parallel edges
- S13 refactoring
- S14 requirement clarification
- S15 ambiguity
- S16 decision recording
- S17 out of scope
- S18 price before work
- S19 readiness audit
- S20 acceptance by probes
- S21 probe breaks things
- S22 derivatives
- S23 data boundaries
- S24 environment failure
- S25 quarantine
- S26 determinism
- S27 operator silence
- S28 escalation
- S29 amend assertion
- S30 waive
- S31 drift
- S32 knowledge/pains
- S33 price/latency
- S34 machine update
- S35 export/import
- S36 format migration
- S37 effort
- S38 prioritization
- S39 documentation
- S40 onboarding

## 3. System features — all (118), by group


**Verbs**

| Code | Feature |
| --- | --- |
| v-next | next: slot, --full, self-doc |
| v-submit | submit: file/stdin/--check/idempotent |
| v-why | why |
| v-bg | bg-suite |
| v-driver | driver fanout/recompile |

**Why topics**

| Code | Feature |
| --- | --- |
| w-price | price |
| w-stage | stage |
| w-derived | derived |
| w-red | red |
| w-latency | latency |
| w-passport | passport |
| w-diary | diary: pain diary |
| w-impact | impact |
| w-knowledge | knowledge |
| w-verif | verifications |
| w-amend | amend |
| w-drift | drift |
| w-accept | acceptance |
| w-surface | surface |
| w-coverage | coverage |
| w-spec | spec: registry+search |
| w-attention | attention |
| w-retro | retro |
| w-handoff | handoff |
| w-hyp | hypotheses |
| w-compat | compatibility |
| w-freeze | freeze: freeze text |
| w-repair | repair: slot notion |
| w-regen | regenerate: slot notion |
| w-rerun | rerun: slot notion |
| w-frag | fragment |
| w-card | entity cards |

**Stages and verdict**

| Code | Feature |
| --- | --- |
| st-intake | intake |
| st-spec | spec |
| st-ac | ac-compile |
| st-impl | implement |
| st-rev | review |
| st-del | deliver |
| st-freeze | composition freeze |
| st-verdict | verdict+digest |
| st-accept | acceptance moment |
| st-refreeze | refreeze |

**Deltas**

| Code | Feature |
| --- | --- |
| d-intent | intent |
| d-lint | lint |
| d-clar | clarify |
| d-checks | checks |
| d-spec | spec |
| d-assert | assert |
| d-split | split |
| d-code | code |
| d-fix | fix |
| d-feature | feature: feature prose + ratification |
| d-probe | probe |
| d-conv | conventions |
| d-amend | amend |
| d-waive | waive |
| d-dec | decision |
| d-kb | kb |

**Spec and edits**

| Code | Feature |
| --- | --- |
| sp-req | requirements CRUD |
| sp-scn | scenarios CRUD/rebind |
| sp-ua | update-assertions |
| sp-detail | detail block |
| sp-dep | dependencies |
| sp-prose | prose scenarios |
| sp-invariant | requirement without a scenario |
| sp-rowid | row identity |

**Checks**

| Code | Feature |
| --- | --- |
| c-suite | suite double run |
| c-probe | coverage-probe |
| c-trace | trace |
| c-derive | derivatives |
| c-proven | proven/free-repair |
| c-envfail | env-fail |
| c-cr | change request line by line |
| c-volatile | volatile |

**Gates**

| Code | Feature |
| --- | --- |
| gt-types | types |
| gt-lint | lint |
| gt-review | review-diff |
| gt-cost | cost |
| gt-latency | latency |
| gt-dc | double-compile |
| gt-rent | gate rental |
| gt-budget | attempt budget |
| gt-quar | quarantine+gap |

**Seeds and materials**

| Code | Feature |
| --- | --- |
| fx-seeds | seeds and materials (seeds/materials) + escapes |

**System fixes**

| Code | Feature |
| --- | --- |
| s-c1 | C1 byte capture |
| s-c1h | C1 amend assertion |
| s-c2 | C2 fragment |
| s-c3 | C3 zero files |
| s-c4 | C4 wish language detection |
| s-c5 | C5 defaults |
| s-c6 | C6 registry+search |
| s-c6r | C6 human spec render |
| s-c7 | C7 reconciliation |

**Renders**

| Code | Feature |
| --- | --- |
| r-agents | AGENTS.md |
| r-spec | SPEC.md |
| r-feat | .punchtape/specs |
| r-sum | SUMMARY.md |
| r-chg | CHANGELOG.md |

**Human attention**

| Code | Feature |
| --- | --- |
| a-q | question batches |
| a-def | silence→defaults |
| a-esc | escalation |
| a-batch | batch in one reply |

**Accounting and knowledge**

| Code | Feature |
| --- | --- |
| k-ledger | ledger |
| k-journal | journal |
| k-att | attention |
| k-diary | pain diary |
| k-retro | retro |
| k-kbadv | kb advice |
| k-oos | out-of-scope |
| k-hyp | hypotheses |
| k-heal | healer |
| k-envdiag | environment diagnoses |
| k-tokens | TOKENS/PT_SUBMITTER |

**Execution and environment**

| Code | Feature |
| --- | --- |
| e-run | run isolation |
| e-lock | run-lock/snapshot |
| e-build | surface materialization |
| e-conv | {out}/{name} conventions |
| e-router | intent router |

## 4. "Task × scenarios" matrix (75 × 40)

| Task | S1 | S2 | S3 | S4 | S5 | S6 | S7 | S8 | S9 | S10 | S11 | S12 | S13 | S14 | S15 | S16 | S17 | S18 | S19 | S20 | S21 | S22 | S23 | S24 | S25 | S26 | S27 | S28 | S29 | S30 | S31 | S32 | S33 | S34 | S35 | S36 | S37 | S38 | S39 | S40 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| br-py-S | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | 🟡 | 🟢 | ⚪ | ⚪ | 🟡 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| br-py-M | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| br-py-L | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| br-go-S | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟡 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | 🟡 | 🟢 | ⚪ | ⚪ | 🟡 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| br-go-M | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| br-go-L | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| br-rs-S | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟡 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | 🟡 | 🟢 | ⚪ | ⚪ | 🟡 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| br-rs-M | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| br-rs-L | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| br-cc-S | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟡 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | 🟡 | 🟢 | ⚪ | ⚪ | 🟡 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| br-cc-M | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | 🟢 | 🟢 | 🟢 | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| br-cc-L | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| br-js-S | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟡 | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | 🟡 | 🟢 | ⚪ | ⚪ | 🟡 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| br-js-M | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| br-js-L | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| cd-py-S | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟡 | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| cd-py-M | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| cd-py-L | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| cd-go-S | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| cd-go-M | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| cd-go-L | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| cd-rs-S | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| cd-rs-M | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| cd-rs-L | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| cd-cc-S | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟡 | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| cd-cc-M | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| cd-cc-L | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| cd-js-S | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | 🟡 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| cd-js-M | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| cd-js-L | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| er-py-S | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟡 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| er-py-M | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| er-py-L | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| er-go-S | 🟢 | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟡 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| er-go-M | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| er-go-L | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| er-rs-S | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟡 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| er-rs-M | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| er-rs-L | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| er-cc-S | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟡 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| er-cc-M | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| er-cc-L | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| er-js-S | 🟢 | 🟡 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟡 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| er-js-M | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| er-js-L | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| pm-py-S | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| pm-py-M | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| pm-py-L | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| pm-go-S | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| pm-go-M | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| pm-go-L | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| pm-rs-S | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| pm-rs-M | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| pm-rs-L | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| pm-cc-S | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟡 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| pm-cc-M | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| pm-cc-L | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| pm-js-S | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟡 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| pm-js-M | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| pm-js-L | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| qr-py-S | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| qr-py-M | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| qr-py-L | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| qr-go-S | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| qr-go-M | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| qr-go-L | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| qr-rs-S | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| qr-rs-M | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| qr-rs-L | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| qr-cc-S | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| qr-cc-M | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| qr-cc-L | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| qr-js-S | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| qr-js-M | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| qr-js-L | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |

## 5. "Task × features" matrix — as tables per meaning

Rows are features (codes from section 3, grouped); columns are a
task's 15 cells (stack × axis). A cell: a coverage circle. The
point of the table: wherever a task under-covers the system in a
feature, that feature goes into rework of its brief/checks.

### brim (br)

| Feature | py-S | py-M | py-L | go-S | go-M | go-L | rs-S | rs-M | rs-L | cc-S | cc-M | cc-L | js-S | js-M | js-L |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| **Verbs** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| v-next next: slot, --full, self-doc | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ |
| v-submit submit: file/stdin/--check/idempotent | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ |
| v-why why | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ |
| v-bg bg-suite | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| v-driver driver fanout/recompile | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **Why topics** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| w-price price | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-stage stage | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-derived derived | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-red red | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-latency latency | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-passport passport | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-diary diary | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-impact impact | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-knowledge knowledge | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-verif verifications | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-amend amend | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-drift drift | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-accept acceptance | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-surface surface | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-coverage coverage | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-spec spec: registry+search | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-attention attention | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-retro retro | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-handoff handoff | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-hyp hypotheses | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-compat compatibility | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-rerun rerun | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-regen regenerate | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-repair repair | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-freeze freeze | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-frag fragment | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-card entity cards | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **Stages and verdict** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| st-intake intake | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ |
| st-spec spec | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ |
| st-ac ac-compile | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ |
| st-impl implement | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ |
| st-rev review | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ |
| st-del deliver | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ |
| st-freeze composition freeze | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ |
| st-verdict verdict+digest | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ |
| st-accept acceptance moment | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| st-refreeze refreeze | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| **Deltas** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| d-intent intent | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ |
| d-lint lint | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| d-clar clarify | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| d-checks checks | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ |
| d-spec spec | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| d-assert assert | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-split split | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-code code | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ |
| d-fix fix | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-feature feature: feature prose + ratification | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-probe probe | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-conv conventions | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ |
| d-amend amend | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-waive waive | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-dec decision | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-kb kb | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **Spec and edits** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| sp-req requirements CRUD | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| sp-scn scenarios CRUD/rebind | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ |
| sp-ua update-assertions | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ |
| sp-detail detail block | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| sp-dep dependencies | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| sp-prose prose scenarios | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| sp-invariant requirement without a scenario | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| sp-rowid row identity | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **Checks** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| c-suite suite double run | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ |
| c-probe coverage-probe | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ |
| c-trace trace | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ |
| c-derive derivatives | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ |
| c-proven proven/free-repair | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ |
| c-envfail env-fail | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| c-cr change request line by line | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| c-volatile volatile | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ |
| **Gates** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| gt-types types | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ |
| gt-lint lint | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| gt-review review-diff | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ |
| gt-cost cost | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ |
| gt-latency latency | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ |
| gt-dc double-compile | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| gt-rent gate rental | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| gt-budget attempt budget | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| gt-quar quarantine+gap | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **Seeds and materials** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| fx-seeds seeds+escapes | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ |
| **System fixes** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| s-c1 C1 byte capture | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ |
| s-c1h C1 amend assertion | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| s-c2 C2 fragment | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| s-c3 C3 zero files | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ |
| s-c4 C4 wish language detection | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ |
| s-c5 C5 defaults | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| s-c6 C6 registry+search | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ |
| s-c6r C6 human spec render | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ |
| s-c7 C7 reconciliation | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **Renders** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| r-agents AGENTS.md | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ |
| r-spec SPEC.md (.punchtape) | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ |
| r-feat spec catalog (file per feature + index) | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ |
| r-sum SUMMARY.md (.punchtape) | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ |
| r-chg CHANGELOG.md (.punchtape) | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ |
| **Human attention** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| a-q question batches | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| a-def silence→defaults | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| a-esc escalation | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| a-batch batch in one reply | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| **Accounting and knowledge** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| k-ledger ledger | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ |
| k-journal journal | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ |
| k-att attention | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ |
| k-diary pain diary | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| k-retro retro | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| k-kbadv kb advice | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| k-oos out-of-scope | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| k-hyp hypotheses | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| k-heal healer | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| k-envdiag environment diagnoses | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| k-tokens TOKENS/PT_SUBMITTER | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **Execution and environment** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| e-run run isolation | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ |
| e-lock run-lock/snapshot | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| e-build surface materialization | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ |
| e-conv {out}/{name} conventions | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ |
| e-router intent router | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |

### conduit (cd)

| Feature | py-S | py-M | py-L | go-S | go-M | go-L | rs-S | rs-M | rs-L | cc-S | cc-M | cc-L | js-S | js-M | js-L |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| **Verbs** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| v-next next: slot, --full, self-doc | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| v-submit submit: file/stdin/--check/idempotent | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| v-why why | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| v-bg bg-suite | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| v-driver driver fanout/recompile | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **Why topics** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| w-price price | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-stage stage | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-derived derived | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-red red | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-latency latency | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-passport passport | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| w-diary diary | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-impact impact | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-knowledge knowledge | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-verif verifications | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-amend amend | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-drift drift | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-accept acceptance | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-surface surface | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| w-coverage coverage | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-spec spec: registry+search | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-attention attention | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-retro retro | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-handoff handoff | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-hyp hypotheses | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-compat compatibility | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-rerun rerun | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-regen regenerate | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-repair repair | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-freeze freeze | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-frag fragment | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-card entity cards | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| **Stages and verdict** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| st-intake intake | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| st-spec spec | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| st-ac ac-compile | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| st-impl implement | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| st-rev review | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| st-del deliver | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| st-freeze composition freeze | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| st-verdict verdict+digest | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| st-accept acceptance moment | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| st-refreeze refreeze | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **Deltas** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| d-intent intent | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| d-lint lint | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-clar clarify | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-checks checks | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| d-spec spec | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| d-assert assert | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-split split | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-code code | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| d-fix fix | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-feature feature: feature prose + ratification | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-probe probe | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-conv conventions | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-amend amend | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-waive waive | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-dec decision | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-kb kb | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **Spec and edits** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| sp-req requirements CRUD | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| sp-scn scenarios CRUD/rebind | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| sp-ua update-assertions | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| sp-detail detail block | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| sp-dep dependencies | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| sp-prose prose scenarios | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| sp-invariant requirement without a scenario | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| sp-rowid row identity | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **Checks** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| c-suite suite double run | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| c-probe coverage-probe | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| c-trace trace | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| c-derive derivatives | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| c-proven proven/free-repair | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| c-envfail env-fail | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| c-cr change request line by line | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| c-volatile volatile | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **Gates** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| gt-types types | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| gt-lint lint | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| gt-review review-diff | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| gt-cost cost | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| gt-latency latency | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| gt-dc double-compile | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| gt-rent gate rental | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| gt-budget attempt budget | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| gt-quar quarantine+gap | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **Seeds and materials** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| fx-seeds seeds+escapes | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| **System fixes** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| s-c1 C1 byte capture | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| s-c1h C1 amend assertion | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| s-c2 C2 fragment | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| s-c3 C3 zero files | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| s-c4 C4 wish language detection | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| s-c5 C5 defaults | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| s-c6 C6 registry+search | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| s-c6r C6 human spec render | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| s-c7 C7 reconciliation | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **Renders** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| r-agents AGENTS.md | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| r-spec SPEC.md (.punchtape) | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| r-feat spec catalog (file per feature + index) | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| r-sum SUMMARY.md (.punchtape) | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| r-chg CHANGELOG.md (.punchtape) | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| **Human attention** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| a-q question batches | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| a-def silence→defaults | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| a-esc escalation | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| a-batch batch in one reply | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **Accounting and knowledge** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| k-ledger ledger | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| k-journal journal | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| k-att attention | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| k-diary pain diary | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| k-retro retro | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| k-kbadv kb advice | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| k-oos out-of-scope | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| k-hyp hypotheses | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| k-heal healer | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| k-envdiag environment diagnoses | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| k-tokens TOKENS/PT_SUBMITTER | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **Execution and environment** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| e-run run isolation | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| e-lock run-lock/snapshot | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| e-build surface materialization | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| e-conv {out}/{name} conventions | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| e-router intent router | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |

### errand (er)

| Feature | py-S | py-M | py-L | go-S | go-M | go-L | rs-S | rs-M | rs-L | cc-S | cc-M | cc-L | js-S | js-M | js-L |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| **Verbs** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| v-next next: slot, --full, self-doc | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| v-submit submit: file/stdin/--check/idempotent | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| v-why why | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| v-bg bg-suite | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| v-driver driver fanout/recompile | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **Why topics** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| w-price price | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-stage stage | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-derived derived | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-red red | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-latency latency | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-passport passport | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-diary diary | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-impact impact | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-knowledge knowledge | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-verif verifications | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-amend amend | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-drift drift | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-accept acceptance | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-surface surface | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| w-coverage coverage | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-spec spec: registry+search | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-attention attention | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-retro retro | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-handoff handoff | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-hyp hypotheses | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-compat compatibility | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-rerun rerun | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-regen regenerate | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-repair repair | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-freeze freeze | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-frag fragment | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-card entity cards | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| **Stages and verdict** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| st-intake intake | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| st-spec spec | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| st-ac ac-compile | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| st-impl implement | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| st-rev review | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| st-del deliver | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| st-freeze composition freeze | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| st-verdict verdict+digest | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| st-accept acceptance moment | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| st-refreeze refreeze | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **Deltas** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| d-intent intent | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| d-lint lint | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| d-clar clarify | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| d-checks checks | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| d-spec spec | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| d-assert assert | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-split split | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-code code | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| d-fix fix | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-feature feature: feature prose + ratification | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-probe probe | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-conv conventions | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| d-amend amend | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-waive waive | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-dec decision | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-kb kb | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **Spec and edits** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| sp-req requirements CRUD | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| sp-scn scenarios CRUD/rebind | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| sp-ua update-assertions | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| sp-detail detail block | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| sp-dep dependencies | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| sp-prose prose scenarios | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| sp-invariant requirement without a scenario | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| sp-rowid row identity | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| **Checks** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| c-suite suite double run | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| c-probe coverage-probe | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| c-trace trace | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| c-derive derivatives | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| c-proven proven/free-repair | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| c-envfail env-fail | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| c-cr change request line by line | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| c-volatile volatile | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **Gates** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| gt-types types | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| gt-lint lint | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| gt-review review-diff | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| gt-cost cost | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| gt-latency latency | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| gt-dc double-compile | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| gt-rent gate rental | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| gt-budget attempt budget | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| gt-quar quarantine+gap | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **Seeds and materials** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| fx-seeds seeds+escapes | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| **System fixes** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| s-c1 C1 byte capture | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| s-c1h C1 amend assertion | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| s-c2 C2 fragment | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| s-c3 C3 zero files | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| s-c4 C4 wish language detection | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| s-c5 C5 defaults | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| s-c6 C6 registry+search | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| s-c6r C6 human spec render | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| s-c7 C7 reconciliation | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **Renders** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| r-agents AGENTS.md | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| r-spec SPEC.md (.punchtape) | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| r-feat spec catalog (file per feature + index) | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| r-sum SUMMARY.md (.punchtape) | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| r-chg CHANGELOG.md (.punchtape) | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| **Human attention** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| a-q question batches | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| a-def silence→defaults | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| a-esc escalation | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| a-batch batch in one reply | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **Accounting and knowledge** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| k-ledger ledger | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| k-journal journal | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| k-att attention | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| k-diary pain diary | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| k-retro retro | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| k-kbadv kb advice | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| k-oos out-of-scope | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| k-hyp hypotheses | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| k-heal healer | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| k-envdiag environment diagnoses | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| k-tokens TOKENS/PT_SUBMITTER | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **Execution and environment** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| e-run run isolation | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| e-lock run-lock/snapshot | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| e-build surface materialization | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | 🟢 | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| e-conv {out}/{name} conventions | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| e-router intent router | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |

### pressmark (pm)

| Feature | py-S | py-M | py-L | go-S | go-M | go-L | rs-S | rs-M | rs-L | cc-S | cc-M | cc-L | js-S | js-M | js-L |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| **Verbs** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| v-next next: slot, --full, self-doc | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| v-submit submit: file/stdin/--check/idempotent | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| v-why why | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| v-bg bg-suite | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| v-driver driver fanout/recompile | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **Why topics** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| w-price price | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-stage stage | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-derived derived | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-red red | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-latency latency | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-passport passport | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-diary diary | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-impact impact | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-knowledge knowledge | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-verif verifications | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-amend amend | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-drift drift | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-accept acceptance | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-surface surface | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-coverage coverage | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-spec spec: registry+search | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-attention attention | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-retro retro | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-handoff handoff | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-hyp hypotheses | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-compat compatibility | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-rerun rerun | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-regen regenerate | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-repair repair | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-freeze freeze | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-frag fragment | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-card entity cards | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **Stages and verdict** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| st-intake intake | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| st-spec spec | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| st-ac ac-compile | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| st-impl implement | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| st-rev review | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| st-del deliver | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| st-freeze composition freeze | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| st-verdict verdict+digest | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| st-accept acceptance moment | 🟡 | ⚪ | ⚪ | 🟡 | ⚪ | ⚪ | 🟡 | ⚪ | ⚪ | 🟡 | ⚪ | ⚪ | 🟡 | ⚪ | ⚪ |
| st-refreeze refreeze | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **Deltas** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| d-intent intent | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| d-lint lint | 🟡 | ⚪ | ⚪ | 🟡 | ⚪ | ⚪ | 🟡 | ⚪ | ⚪ | 🟡 | ⚪ | ⚪ | 🟡 | ⚪ | ⚪ |
| d-clar clarify | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-checks checks | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| d-spec spec | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| d-assert assert | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-split split | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-code code | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| d-fix fix | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-feature feature: feature prose + ratification | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-probe probe | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-conv conventions | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| d-amend amend | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-waive waive | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-dec decision | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-kb kb | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **Spec and edits** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| sp-req requirements CRUD | 🟡 | ⚪ | ⚪ | 🟡 | ⚪ | ⚪ | 🟡 | ⚪ | ⚪ | 🟡 | ⚪ | ⚪ | 🟡 | ⚪ | ⚪ |
| sp-scn scenarios CRUD/rebind | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| sp-ua update-assertions | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| sp-detail detail block | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| sp-dep dependencies | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| sp-prose prose scenarios | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| sp-invariant requirement without a scenario | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| sp-rowid row identity | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **Checks** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| c-suite suite double run | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| c-probe coverage-probe | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| c-trace trace | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| c-derive derivatives | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| c-proven proven/free-repair | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| c-envfail env-fail | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| c-cr change request line by line | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| c-volatile volatile | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **Gates** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| gt-types types | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| gt-lint lint | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| gt-review review-diff | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| gt-cost cost | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| gt-latency latency | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| gt-dc double-compile | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| gt-rent gate rental | 🟡 | ⚪ | ⚪ | 🟡 | ⚪ | ⚪ | 🟡 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| gt-budget attempt budget | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟡 | ⚪ | ⚪ |
| gt-quar quarantine+gap | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **Seeds and materials** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| fx-seeds seeds+escapes | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| **System fixes** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| s-c1 C1 byte capture | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| s-c1h C1 amend assertion | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| s-c2 C2 fragment | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| s-c3 C3 zero files | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| s-c4 C4 wish language detection | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| s-c5 C5 defaults | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| s-c6 C6 registry+search | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| s-c6r C6 human spec render | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| s-c7 C7 reconciliation | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **Renders** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| r-agents AGENTS.md | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| r-spec SPEC.md (.punchtape) | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| r-feat spec catalog (file per feature + index) | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| r-sum SUMMARY.md (.punchtape) | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| r-chg CHANGELOG.md (.punchtape) | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| **Human attention** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| a-q question batches | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| a-def silence→defaults | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| a-esc escalation | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| a-batch batch in one reply | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| **Accounting and knowledge** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| k-ledger ledger | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| k-journal journal | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| k-att attention | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| k-diary pain diary | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| k-retro retro | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| k-kbadv kb advice | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| k-oos out-of-scope | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| k-hyp hypotheses | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| k-heal healer | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| k-envdiag environment diagnoses | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| k-tokens TOKENS/PT_SUBMITTER | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **Execution and environment** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| e-run run isolation | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| e-lock run-lock/snapshot | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| e-build surface materialization | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| e-conv {out}/{name} conventions | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| e-router intent router | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |

### quire (qr)

| Feature | py-S | py-M | py-L | go-S | go-M | go-L | rs-S | rs-M | rs-L | cc-S | cc-M | cc-L | js-S | js-M | js-L |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| **Verbs** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| v-next next: slot, --full, self-doc | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| v-submit submit: file/stdin/--check/idempotent | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| v-why why | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| v-bg bg-suite | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| v-driver driver fanout/recompile | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **Why topics** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| w-price price | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-stage stage | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-derived derived | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-red red | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-latency latency | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-passport passport | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-diary diary | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-impact impact | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-knowledge knowledge | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-verif verifications | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-amend amend | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-drift drift | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-accept acceptance | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-surface surface | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-coverage coverage | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| w-spec spec: registry+search | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| w-attention attention | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-retro retro | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-handoff handoff | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-hyp hypotheses | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-compat compatibility | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-rerun rerun | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-regen regenerate | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-repair repair | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-freeze freeze | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-frag fragment | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| w-card entity cards | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **Stages and verdict** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| st-intake intake | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| st-spec spec | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| st-ac ac-compile | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| st-impl implement | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| st-rev review | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| st-del deliver | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| st-freeze composition freeze | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| st-verdict verdict+digest | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| st-accept acceptance moment | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| st-refreeze refreeze | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **Deltas** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| d-intent intent | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| d-lint lint | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-clar clarify | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| d-checks checks | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| d-spec spec | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| d-assert assert | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-split split | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-code code | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| d-fix fix | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-feature feature: feature prose + ratification | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-probe probe | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-conv conventions | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| d-amend amend | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-waive waive | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-dec decision | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| d-kb kb | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **Spec and edits** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| sp-req requirements CRUD | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| sp-scn scenarios CRUD/rebind | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| sp-ua update-assertions | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| sp-detail detail block | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| sp-dep dependencies | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| sp-prose prose scenarios | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| sp-invariant requirement without a scenario | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| sp-rowid row identity | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **Checks** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| c-suite suite double run | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| c-probe coverage-probe | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| c-trace trace | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| c-derive derivatives | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| c-proven proven/free-repair | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| c-envfail env-fail | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| c-cr change request line by line | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| c-volatile volatile | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **Gates** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| gt-types types | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| gt-lint lint | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| gt-review review-diff | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| gt-cost cost | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| gt-latency latency | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| gt-dc double-compile | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| gt-rent gate rental | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| gt-budget attempt budget | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| gt-quar quarantine+gap | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **Seeds and materials** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| fx-seeds seeds+escapes | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| **System fixes** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| s-c1 C1 byte capture | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| s-c1h C1 amend assertion | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| s-c2 C2 fragment | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| s-c3 C3 zero files | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| s-c4 C4 wish language detection | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| s-c5 C5 defaults | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| s-c6 C6 registry+search | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| s-c6r C6 human spec render | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| s-c7 C7 reconciliation | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **Renders** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| r-agents AGENTS.md | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| r-spec SPEC.md (.punchtape) | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| r-feat spec catalog (file per feature + index) | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| r-sum SUMMARY.md (.punchtape) | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| r-chg CHANGELOG.md (.punchtape) | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| **Human attention** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| a-q question batches | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| a-def silence→defaults | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| a-esc escalation | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| a-batch batch in one reply | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| **Accounting and knowledge** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| k-ledger ledger | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| k-journal journal | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| k-att attention | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| k-diary pain diary | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| k-retro retro | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| k-kbadv kb advice | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| k-oos out-of-scope | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| k-hyp hypotheses | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| k-heal healer | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| k-envdiag environment diagnoses | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| k-tokens TOKENS/PT_SUBMITTER | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| **Execution and environment** |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
| e-run run isolation | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| e-lock run-lock/snapshot | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |
| e-build surface materialization | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| e-conv {out}/{name} conventions | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ | 🟢 | ⚪ | ⚪ |
| e-router intent router | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ | ⚪ |

## 6. Governance rules

0. ONE SCENARIO = ONE SEPARATE BRIEF IN HUMAN LANGUAGE. Every
   scenario from section 2, once it is reworked into a task and
   run, lives in its own separate brief — in human language,
   under the campaign brief rules (no codes, byte-level notes or
   technical plumbing in the brief text).
1. Running a task in a campaign colors its row in sections 4–5
   from actual evidence (run metrics and journal).
2. Reworking a task = add scenarios to its brief (section 4,
   empty cells of the row) and features to its checks (section 5,
   empty cells): the task grows until it covers the needed share.
3. New runs create no new sections — only circles.
4. Size boundaries are derived only from run measurements
   (calibration in docs/SAMPLES.md): task axes were recut from
   the baseline run; every new run refines the bands and axes.

## 7. Honest caveats

- The auto-run coloring (2026-08-30, automated run): the 25 S-rows
  of sections 1/4/5 were colored on 2026-08-30 by actual
  instance evidence (canon/ledgers/journals). Automated run:
  machine acceptances PASS; there were no live probes by
  the operator.
- Background rehearsals (v-bg): colored by the auto-run for all five
  errand tasks (bg-suite entries scope: background in ledgers);
  other meanings recorded no rehearsal verb — not colored.
- S14 (requirement clarification): did not occur in the baseline run; the
  auto-run closed
  it for errand/pressmark/quire/brim with explicit answers; S27
  (operator silence) — fired only in cd-rs/cd-js (the recommended
  default applied; not colored in section 4); S28 — empty in fact.
- S39 (documentation): the .punchtape/specs directory — the
  renders r-agents/r-spec/r-feat/r-sum/r-chg colored by the auto-run for all 25.
- The baseline pressmark run's tokens were lost by the platform ("no data");
  auto-run tokens were not reported by instances ("no data") — k-tokens not colored.
- Calibration judgments (marked by the analysts as judgments):
  errand st-accept/gt-cost/gt-latency — by the baseline errand row calibration
  (equivalent meters); quire — by the baseline quire row anchor; waivers —
  one per stack.
- The br-cc-M row (baseline) — a baseline-run artifact: its green
  v-why/k-att/c-volatile/sp-ua/
  s-c3/s-c6 are neither confirmed nor refuted by the auto-run evidence
  (a different artifact).