# The verification harness

Committed FIRST, before fix #1: the infrastructure that proves the
fixes. Three pieces, all campaign tooling (no machine code, no
canon — the machine itself is untouched by this commit).

## 1. synthetic-probes.sh — the five probes + zh, one fixed script

`./synthetic-probes.sh [probe...] [--record]` — replays the
agnosticism audit's five synthetic task senses (rpn / translit /
dice / tick / bin — senses ABSENT from the battery) plus one
zh-wish variant, as fixed next/submit/why sequences against a fresh
build. Deterministic, zero LLM tokens, ~2-4 min wall. Every step
asserts its exit code and fixed-string patterns; the normalized
full transcript is diffed against the committed baseline
(probes-baseline/*.txt). Exit non-zero on any FAIL.

Per-probe coverage:

| Probe | Exercises | Baseline carries (pre-fix) |
|---|---|---|
| rpn | CLI/streams/terminator, question batch (ru), answers-start-freeze, env FIX hint at zero code, `why red`, `submit --check` | (answers → STAGE implement + COMPOSITION FREEZE); terminator reds with measured bytes and NO copy-ready op (move 4 target); env red advises "install the tool"; ACTION line without a --check clause (move 3 target) |
| translit | file observations, CRLF, three-channel EOL matrix | state-text red with the byte window; full-form file `equals` GREEN on the same CRLF divergence; submit reply `suite=green` headline while `next --full` shows RED LIST (1) |
| dice | determinism model, degenerate-green, transience red, volatile hold + amend | degenerate constant passes green free; "not transient" red WITHOUT a volatile proposal line; volatile on a proven row → ТРЕБУЕТСЯ УТВЕРЖДЕНИЕ (exit 3) → amend approve → green |
| tick | long-running shape, timeout, orchestration, phantom family, freeze | naive row "timed out after 10s" red; phantom REQ-002 «команда `sh`» drafted from the taught `sh -c` pattern; the deadlock pair refuses (rm1 "no scenarios", rm2 "has scenarios"), the both-ops delta is ACCEPTED (the untaught escape), removal of a PROVEN row's requirement refused "frozen" |
| bin | escape channels, one-byte output | the double-quoted `"\xE4"` pin ACCEPTED at intake, red with the UTF-8 ä evidence; single-quote `'\x51'` passes write→read byte-exact |
| zh | localized renders end-to-end | SUMMARY.md rendered in Chinese; the unpinned-CLI-surface WARNING falls back to ENGLISH — summary.checked.unpinned.zh missing |

## 2. agnostic-gate.sh (extended) + agnostic-gate-selftest.sh

The extended gate over machine code (incl. comments) and canon:
sense/stack (as before) + stack-nearmiss (`node <script>`,
.mjs/.cjs — bare `node` deliberately NOT matched: it is a
legitimate yaml.Node identifier; bare `go` excluded as the everyday
verb) + batref (M-<n>/F-<n>/wave names/two-digit grading counts) +
nouns (canondata example nouns must belong to the synthetic set) +
locale (ru↔zh completeness + a drift blacklist of known-wrong
forms). Takes an optional ROOT for the self-test's scratch trees.

`./agnostic-gate-selftest.sh` plants each violation class (7
fixtures: clean-pass, sense, node-invocation, battery-ref,
example-noun, locale-gap, locale-drift) and asserts the gate FAILS
on each with the right layer — a gate that cannot fail proves
nothing.

**The pre-fix gate state on this tree is RED, by design** — the
extended gate catches exactly the pending fix classes, and the
per-fix protocol flips them green:
- [stack-nearmiss] bgsuite.go:93 `node app.mjs`
- [batref] 10 lines across bgsuite.go, summaryrender.go, apply.go,
  verify.go (the class extends beyond the plan's two
  named files: ALL battery references in comments go, same fix
  semantics — cite fixes by commit, not by finding id)
- [nouns] agentsref.yaml:132/:167, slots.yaml:27/:55 (`bin/serve`,
  `materials/about.md`)
- [locale] summary.checked.unpinned.zh missing; [locale-drift]
  changelog/features/slots/why drift forms

## 3. double-build.sh — the freeze ritual

`./double-build.sh` — make build twice, cmp, print the sha256. The
reproducibility witness reused per batch and at the th3 binary
freeze. Current tree witness (this commit): PASS, sha256
6902cc112a5fbbc72b27563137fea372867e342af363e18ce37c0c7f1b087424.

## The per-fix protocol for baselines

The baselines are committed artifacts. After each fix: run the full
probe suite; a divergence is a FAILURE unless it is exactly the
fix's intended change — in that case update ONLY
the affected baseline lines in the FIX'S OWN COMMIT, with the
change named in the commit message. Never re-record wholesale; a
wholesale re-record hides regressions. The dice probe's random
clock values are the only masked content bytes (documented in the
script's normalizer); everything semantic is byte-checked.

## The batch validation gate record (2026-09-03)

Steps 1-4 on the final tree (HEAD d9a4191 + this record):
1. build / go vet / staticcheck — CLEAN.
2. The extended gate — GATE: PASS; the planted-violation self-test — PASS (7/7).
3. The six probes — ALL PASS; probes-baseline/* at this commit IS the
   pre-th3 baseline.
4. Double build — PASS, one sha256 across two clean builds:
   adab583bf143138df10f607274fe67a0f2dee98f46a3a955f7fffd87139155bb
   (7228611 bytes) — the th3 freeze candidate.
Step 5 (the ~6-cell verification control wave) follows per the plan.
