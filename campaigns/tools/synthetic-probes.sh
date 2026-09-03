#!/bin/bash
# synthetic-probes.sh — the five synthetic probes of the
# agnosticism audit (rpn / translit / dice / tick / bin) plus one
# zh-wish variant, as ONE fixed script with recorded expected
# outcomes. A script plays the naive
# executor: deterministic next/submit/why sequences against a
# fresh build, zero LLM tokens, ~10 min CPU.
#
# Each probe replays fixed deltas; every step asserts (a) the exit
# code and (b) fixed-string patterns of the reply; the full
# normalized transcript is diffed against the committed baseline
# (probes-baseline/<probe>.txt) — "no new problems":
# any UNEXPECTED divergence fails; a fix that intends a change
# updates its baseline lines IN THAT FIX'S COMMIT (documented
# there), never silently.
#
# Normalization masks only run-varying bytes: timestamps, tx/digest
# hexes, wall times, /tmp paths, long digit runs (the dice probe's
# nanosecond clock values). Everything semantic — verdicts, red
# reasons, byte windows, labels — stays byte-checked.
#
# Usage:
#   synthetic-probes.sh                 # run all, PASS/FAIL per probe
#   synthetic-probes.sh rpn tick        # run selected probes
#   synthetic-probes.sh --record [...]  # (re)write the baselines
#   PUNCHTAPE_BIN=<path> ...            # use a given binary

set -uo pipefail

HERE="$(cd "$(dirname "$0")" && pwd)"
REPO="$(cd "$HERE/../.." && pwd)"
BASELINE_DIR="$HERE/probes-baseline"
WORKROOT="/tmp/punchtape-probes"
RECORD=0
PROBES=()

for a in "$@"; do
  case "$a" in
    --record) RECORD=1 ;;
    *) PROBES+=("$a") ;;
  esac
done
[ ${#PROBES[@]} -eq 0 ] && PROBES=(rpn translit dice tick bin zh)

BIN="${PUNCHTAPE_BIN:-$WORKROOT/punchtape}"
if [ -z "${PUNCHTAPE_BIN:-}" ]; then
  echo "== building the machine (make build)"
  (cd "$REPO" && make build) >/dev/null || { echo "BUILD FAIL"; exit 2; }
  mkdir -p "$WORKROOT"
  cp "$REPO/punchtape" "$BIN"
fi

mkdir -p "$WORKROOT"
overall=0
cur_probe=""; cur_step=0; transcript=""

# step <expected-rc> <pattern>... -- <args...>: run one verb call,
# append rc+output to the transcript, assert rc and grep -F patterns.
step() {
  local want_rc="$1"; shift
  local pats=()
  while [ "$1" != "--" ]; do pats+=("$1"); shift; done
  shift
  cur_step=$((cur_step+1))
  local out rc=0
  out="$("$BIN" "$@" 2>&1)" || rc=$?
  {
    echo "### step $cur_step: $*"
    echo "rc=$rc"
    echo "$out"
    echo
  } >> "$transcript"
  if [ "$rc" != "$want_rc" ]; then
    echo "  STEP $cur_step FAIL: expected rc=$want_rc, got $rc ($*)"
    echo "$out" | head -3 | sed 's/^/    /'
    return 1
  fi
  local p
  for p in "${pats[@]}"; do
    if ! grep -qF -- "$p" <<<"$out"; then
      echo "  STEP $cur_step FAIL: pattern not found: $p ($*)"
      echo "$out" | head -5 | sed 's/^/    /'
      return 1
    fi
  done
  return 0
}

# submit_delta <expected-rc> <pattern>... -- <delta-file>: submit a
# delta file that lives in the probe dir.
submit_delta() {
  local want_rc="$1"; shift
  local pats=()
  while [ "$1" != "--" ]; do pats+=("$1"); shift; done
  shift
  local f="$1"; shift
  step "$want_rc" "${pats[@]}" -- submit "$f"
}

normalize() {  # mask run-varying bytes only
  sed -E \
    -e 's/[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9:.]+Z/<ts>/g' \
    -e 's/transaction [0-9a-f]+/transaction <tx>/g' \
    -e 's/\b[0-9a-f]{64}\b/<sha256>/g' \
    -e 's/\b[0-9a-f]{12}\b/<digest12>/g' \
    -e 's/\b[0-9]{6,}\b/<num>/g' \
    -e 's/[0-9]+(\.[0-9]+)?[[:space:]]?(ms|µs)\b/<time>/g' \
    -e 's/([：:][[:space:]]*)[0-9]+(\.[0-9]+)?[[:space:]]?(ms|µs|s)\b/\1<time>/g' \
    -e 's/~[0-9]+ tokens\b/~<num> tokens/g' \
    -e 's/ceremony tokens ~[0-9]+/ceremony tokens ~<num>/g' \
    -e 's/tokens ~[0-9]+/tokens ~<num>/g' \
    -e 's/[0-9]+\.[0-9]+[[:space:]]?s\b/<time>/g' \
    -e 's#/tmp/punchtape-probes[^ ,"`]*#<tmp>#g'
}

finish_probe() {  # compare the normalized transcript with the baseline
  local base="$BASELINE_DIR/$cur_probe.txt"
  normalize < "$transcript" > "$transcript.norm"
  if [ "$RECORD" = 1 ]; then
    mkdir -p "$BASELINE_DIR"
    cp "$transcript.norm" "$base"
    echo "PROBE $cur_probe: RECORDED ($(wc -l < "$base") lines)"
    return 0
  fi
  if [ ! -f "$base" ]; then
    echo "PROBE $cur_probe: FAIL — no baseline (run with --record first)"
    overall=1
    return 0
  fi
  if diff -u "$base" "$transcript.norm" > "$transcript.diff"; then
    echo "PROBE $cur_probe: PASS"
  else
    echo "PROBE $cur_probe: FAIL — transcript diverges from the baseline:"
    head -30 "$transcript.diff" | sed 's/^/  /'
    overall=1
  fi
}

begin_probe() {
  cur_probe="$1"
  cur_step=0
  probe_dir="$WORKROOT/$cur_probe"
  rm -rf "$probe_dir"; mkdir -p "$probe_dir"
  cd "$probe_dir" || exit 2   # delta heredocs land beside the instance
  transcript="$probe_dir/transcript.txt"
  : > "$transcript"
}

run_rpn() {
  begin_probe rpn
  cat > d1.yaml <<'YAML'
kind: checks
submission-key: checks-001
rows:
  - run: ["rpn 3 4 +"]
    out: "7"
    rc: 0
  - run: ["rpn 10 2 /"]
    out: "5"
    rc: 0
  - run: ["rpn 10 4 /"]
    out: "2.5"
    rc: 0
  - run: ["rpn 10 0 /"]
    rc: 2
    err: "rpn: division by zero"
  - run: ["rpn 3 x +"]
    rc: 2
    err: "rpn: bad token: x"
questions:
  - id: Q1
    dimension: ambiguous
    finding: |
      Целый результат: печатать без точки (7) или всегда с точкой (7.0)?
    question: |
      Как печатать целый результат деления?
    options:
      - id: a
        label: без точки, как целое (10 2 / = 5)
        scope: ~10 токенов
        default: true
      - id: b
        label: всегда с точкой (10 2 / = 5.0)
        scope: ~10 токенов
YAML
  cat > d2.yaml <<'YAML'
kind: checks
submission-key: checks-002
answers:
  - question: Q1
    option: a
YAML
  cat > d3.json <<'JSON'
{"kind":"code","submission-key":"impl-001","card":"CRD-001","files":[{"path":"rpn","content":"#!/usr/bin/env python3\nimport sys\ndef die(msg):\n    print(f\"rpn: {msg}\", file=sys.stderr)\n    sys.exit(2)\nstack = []\nfor tok in sys.argv[1:]:\n    if tok in (\"+\", \"-\", \"*\", \"/\"):\n        if len(stack) < 2:\n            die(\"stack underflow\")\n        b = stack.pop()\n        a = stack.pop()\n        if tok == \"+\":\n            stack.append(a + b)\n        elif tok == \"-\":\n            stack.append(a - b)\n        elif tok == \"*\":\n            stack.append(a * b)\n        else:\n            if b == 0:\n                die(\"division by zero\")\n            q = a / b\n            stack.append(int(q) if q == int(q) else q)\n    else:\n        try:\n            stack.append(int(tok))\n        except ValueError:\n            die(f\"bad token: {tok}\")\nif len(stack) != 1:\n    die(\"stack has leftovers\")\nr = stack[0]\nprint(int(r) if isinstance(r, int) else r)\n"}]}
JSON
  cat > bad.yaml <<'YAML'
kind: checks
submission-key: preflight-x
rows:
  - run: ["rpn 1 2 +"]
    bogus-field: 1
YAML
  # the repair IS the machine's terminator proposal, verbatim
  cat > d4.yaml <<'YAML'
kind: spec
submission-key: checks-003
operations:
  - update-assertions:
      scenario: SCN-001
      ops:
        - match: {observation: stdout, condition: equals, value: '7'}
          set: {observation: stdout, condition: equals, value: '7\x0A'}
  - update-assertions:
      scenario: SCN-002
      ops:
        - match: {observation: stdout, condition: equals, value: '5'}
          set: {observation: stdout, condition: equals, value: '5\x0A'}
  - update-assertions:
      scenario: SCN-003
      ops:
        - match: {observation: stdout, condition: equals, value: '2.5'}
          set: {observation: stdout, condition: equals, value: '2.5\x0A'}
  - update-assertions:
      scenario: SCN-004
      ops:
        - match: {observation: stderr, condition: equals, value: 'rpn: division by zero'}
          set: {observation: stderr, condition: equals, value: 'rpn: division by zero\x0A'}
  - update-assertions:
      scenario: SCN-005
      ops:
        - match: {observation: stderr, condition: equals, value: 'rpn: bad token: x'}
          set: {observation: stderr, condition: equals, value: 'rpn: bad token: x\x0A'}
YAML
  cat > d5.yaml <<'YAML'
kind: feature
submission-key: feature-001
entries:
  - requirement: REQ-001
    text: |
      Калькулятор обратной польской записи. Вызов: rpn <токен>... —
      выражение приходит аргументами командной строки, каждый токен
      отдельным аргументом. Поддерживаются целые числа и операции
      + - * /; деление даёт точное частное (10 4 / = 2.5, целое
      печатается без точки). Результат печатается одной строкой в
      stdout. Деление на ноль и мусорный токен — отказ: сообщение
      rpn: ... в stderr и код выхода 2.
YAML
  local ok=0
  step 0 "AGENTS.md in full" -- next "Калькулятор обратной польской записи: rpn <токен>... — выражение аргументами командной строки; целые и + - * /; деление даёт точное частное; результат одной строкой в stdout; деление на ноль и мусорный токен — отказ rpn: ... в stderr и код 2" || ok=1
  submit_delta 3 "ВОПРОСЫ ЧЕЛОВЕКУ" "suite=red" -- d1.yaml || ok=1
  submit_delta 0 "STAGE: implement" "COMPOSITION FREEZE" -- d2.yaml || ok=1
  submit_delta 0 'TST-001 SCN-001: stdout is not equal to "7"' "first difference at byte 1" -- d3.json || ok=1
  step 0 "RED — 5 red check(s)" -- why red || ok=1
  # --check must not change the instance: the journal (submission
  # transactions; the bg-suite never writes it) is the race-free witness
  local journal_before
  journal_before=$(grep -c '^submission-key:' "$probe_dir/.punchtape/journal.yamll")
  submit_delta 1 "unknown field bogus-field" -- bad.yaml || ok=1
  local journal_after
  journal_after=$(grep -c '^submission-key:' "$probe_dir/.punchtape/journal.yamll")
  if [ "$journal_before" != "$journal_after" ]; then
    echo "  STEP FAIL: submit --check changed the journal ($journal_before -> $journal_after)"
    ok=1
  fi
  submit_delta 0 "suite=green" -- d4.yaml || ok=1
  submit_delta 0 "VERDICT: READY" -- d5.yaml || ok=1
  [ "$ok" = 0 ] && finish_probe || { echo "PROBE rpn: FAIL (step assertions)"; overall=1; }
}

run_translit() {
  begin_probe translit
  cat > d1.yaml <<'YAML'
kind: checks
submission-key: checks-001
rows:
  - seeds:
      - in.txt первая\x0Aвторая
    run: ["translit in.txt out.txt"]
    out: "pervaya\nvtoraya"
    state:
      - out.txt text pervaya\nvtoraya
YAML
  cat > d3.json <<'JSON'
{"kind":"code","submission-key":"impl-001","card":"CRD-001","files":[{"path":"translit","content":"#!/usr/bin/env python3\nimport sys\nM = {'п':'p','е':'e','р':'r','в':'v','а':'a','я':'ya','т':'t','о':'o','в':'v'}\ndata = open(sys.argv[1], encoding='utf-8').read()\nout = ''.join(M.get(ch, ch) for ch in data)\nopen(sys.argv[2], 'w', encoding='utf-8', newline='').write(out.replace('\\n', '\\r\\n'))\nsys.stdout.write(out.replace('\\n', '\\r\\n'))\n"}]}
JSON
  cat > d5.yaml <<'YAML'
kind: spec
submission-key: spec-001
operations:
  - add-scenario:
      id: SCN-002
      requirement: REQ-001
      summary: translit full-form file equals (EOL-normalized channel)
      seed:
        - path: in.txt
          content: "первая\nвторая"
      when:
        surface: cli
        command: [translit, in.txt, out.txt]
        timeout-sec: 10
      then:
        - observation: file
          condition: equals
          path: out.txt
          value: "pervaya\nvtoraya"
YAML
  local ok=0
  step 0 "AGENTS.md in full" -- next "Транслитератор: translit <вход> <выход> — читает текстовый файл, заменяет кириллицу на латиницу, пишет результат в файл и печатает его в stdout" || ok=1
  submit_delta 0 "suite=red" -- d1.yaml || ok=1
  submit_delta 0 "FILES WRITTEN" -- d3.json || ok=1
  step 0 'got "pervaya\r\nvtoraya"' -- why red || ok=1
  submit_delta 0 "suite=green" -- d5.yaml || ok=1
  step 0 "RED LIST (1)" -- next --full || ok=1
  [ "$ok" = 0 ] && finish_probe || { echo "PROBE translit: FAIL (step assertions)"; overall=1; }
}

run_dice() {
  begin_probe dice
  cat > d1.yaml <<'YAML'
kind: checks
submission-key: checks-001
rows:
  - run: ["dice 6"]
    out: "1\n"
    rc: 0
  - run: ["dice 12"]
    rc: 0
YAML
  cat > d2.json <<'JSON'
{"kind":"code","submission-key":"impl-001","card":"CRD-001","files":[{"path":"dice","content":"#!/bin/sh\necho $((RANDOM % $1 + 1))\n"}]}
JSON
  cat > d3.json <<'JSON'
{"kind":"fix","submission-key":"fix-001","card":"CRD-001","reply-to":"suite","files":[{"path":"dice","content":"#!/bin/sh\ndate +%s%N\n"}]}
JSON
  cat > d4.yaml <<'YAML'
kind: checks
submission-key: checks-002
rows:
  - run: ["dice 6"]
    out: "1\n"
    rc: 0
  - run: ["dice 12"]
    rc: 0
    volatile: [stdout]
YAML
  cat > d5.yaml <<'YAML'
kind: amend
submission-key: amend-001
decision: approve
YAML
  cat > d7.json <<'JSON'
{"kind":"fix","submission-key":"fix-002","card":"CRD-001","reply-to":"suite","files":[{"path":"dice","content":"#!/bin/sh\nn=${1:-6}\necho $(( 1 ))\n"}]}
JSON
  cat > d6.yaml <<'YAML'
kind: feature
submission-key: feature-001
entries:
  - requirement: REQ-001
    text: |
      Игральная кость dice <N>: печатает случайное целое от 1 до N
      одной строкой в stdout. Непредсказуемость значения заявлена в
      таблице проверок (volatile stdout); проверяемая часть — код
      выхода и наличие значения.
YAML
  local ok=0
  step 0 "AGENTS.md in full" -- next "Игральная кость: dice <N> — печатает случайное целое от 1 до N в stdout" || ok=1
  submit_delta 0 "suite=red" -- d1.yaml || ok=1
  submit_delta 0 "suite=green" -- d2.json || ok=1
  submit_delta 0 "not transient" -- d3.json || ok=1
  submit_delta 3 "ТРЕБУЕТСЯ УТВЕРЖДЕНИЕ" -- d4.yaml || ok=1
  submit_delta 0 "suite=green" -- d5.yaml || ok=1
  submit_delta 0 "suite=green" -- d7.json || ok=1
  submit_delta 0 "VERDICT: READY" -- d6.yaml || ok=1
  [ "$ok" = 0 ] && finish_probe || { echo "PROBE dice: FAIL (step assertions)"; overall=1; }
}

run_tick() {
  begin_probe tick
  cat > d1.yaml <<'YAML'
kind: checks
submission-key: checks-001
rows:
  - run: ["tick 60"]
    rc: 0
YAML
  cat > d2.json <<'JSON'
{"kind":"code","submission-key":"impl-001","card":"CRD-001","files":[{"path":"tick","content":"#!/bin/sh\ni=1\nwhile :\ndo\n  echo \"tick $i\"\n  i=$((i+1))\n  sleep 1\ndone\n"}]}
JSON
  cat > d3.yaml <<'YAML'
kind: checks
submission-key: checks-002
rows:
  - run: ["tick 60"]
    rc: 0
  - run: ["sh -c './tick & p=$!; sleep 2; kill $p 2>/dev/null'"]
    rc: 0
  - run: ["sh -c 'sleep 12'"]
    rc: 0
    timeout-sec: 15
YAML
  cat > rm1.yaml <<'YAML'
kind: spec
submission-key: spec-rm1
operations:
  - remove-scenario: {id: SCN-001}
YAML
  cat > rm2.yaml <<'YAML'
kind: spec
submission-key: spec-rm2
operations:
  - remove-requirement: {id: REQ-001}
YAML
  cat > rm3.yaml <<'YAML'
kind: spec
submission-key: spec-rm3
operations:
  - remove-scenario: {id: SCN-001}
  - remove-requirement: {id: REQ-001}
YAML
  cat > rm4.yaml <<'YAML'
kind: spec
submission-key: spec-rm4
operations:
  - remove-requirement: {id: REQ-002}
YAML
  local ok=0
  step 0 "AGENTS.md in full" -- next "Демон tick: tick <сек> — печатает строку tick каждую секунду в stdout, работает до сигнала завершения" || ok=1
  submit_delta 0 "suite=red" -- d1.yaml || ok=1
  submit_delta 0 "timed out after 10s" -- d2.json || ok=1
  submit_delta 0 "RED LIST (1)" -- d3.yaml || ok=1
  submit_delta 1 "no scenarios" -- rm1.yaml || ok=1
  # the placeholder requirement (every row never green) removal
  # STAGES as a deferred edit instead of the deadlock refusal
  submit_delta 3 "ТРЕБУЕТСЯ УТВЕРЖДЕНИЕ" "placeholder" -- rm2.yaml || ok=1
  cat > rm5.yaml <<'YAML'
kind: amend
submission-key: amend-001
decision: approve
YAML
  submit_delta 0 "ACCEPTED" -- rm5.yaml || ok=1
  # the phantom (proven row) survives the cleanup: its removal keeps
  # the change-request refusal — the ladder check
  step 0 "команда" -- why REQ-002 || ok=1
  submit_delta 1 "frozen" -- rm4.yaml || ok=1
  [ "$ok" = 0 ] && finish_probe || { echo "PROBE tick: FAIL (step assertions)"; overall=1; }
}

run_bin() {
  begin_probe bin
  cat > d1.yaml <<'YAML'
kind: checks
submission-key: checks-001
rows:
  - run: ["crc8 31"]
    out: "\xE4"
    rc: 0
YAML
  cat > d1sq.yaml <<'YAML'
kind: checks
submission-key: checks-002
rows:
  - run: ["crc8 31"]
    out: '\xE4'
    rc: 0
YAML
  cat > d2.json <<'JSON'
{"kind":"code","submission-key":"impl-001","card":"CRD-001","files":[{"path":"crc8","content":"#!/usr/bin/env python3\nimport sys\ndata = sys.argv[1].encode()\ncrc = 0\nfor b in data:\n    crc ^= b\n    for _ in range(8):\n        crc = ((crc << 1) ^ 0x07) & 0xFF if crc & 0x80 else (crc << 1) & 0xFF\nsys.stdout.buffer.write(bytes([crc]))\n"}]}
JSON
  cat > d3.yaml <<'YAML'
kind: checks
submission-key: checks-003
rows:
  - run: ["crc8 31"]
    out: '\x51'
    rc: 0
YAML
  cat > d4.yaml <<'YAML'
kind: feature
submission-key: feature-001
entries:
  - requirement: REQ-001
    text: |
      Контрольная сумма crc8: crc8 <аргументы> — печатает один
      сырой байт контрольной суммы CRC-8 (полином 0x07) в stdout,
      без переводов строк.
YAML
  local ok=0
  step 0 "AGENTS.md in full" -- next "Контрольная сумма: crc8 <аргументы> — печатает один сырой байт crc8 суммы в stdout" || ok=1
  # the double-quoted \xE4 pin is REFUSED at intake with the recipe
  submit_delta 1 "expectation value" "single quotes" -- d1.yaml || ok=1
  # the single-quote form is legal at intake (write→read byte-exact)
  submit_delta 0 "suite=red" -- d1sq.yaml || ok=1
  submit_delta 0 "not equal" -- d2.json || ok=1
  submit_delta 0 "suite=green" -- d3.yaml || ok=1
  submit_delta 0 "VERDICT: READY" -- d4.yaml || ok=1
  [ "$ok" = 0 ] && finish_probe || { echo "PROBE bin: FAIL (step assertions)"; overall=1; }
}

run_zh() {
  begin_probe zh
  cat > d1.yaml <<'YAML'
kind: checks
submission-key: checks-001
rows:
  - run: ["calc 2 3 +"]
    out: "5\n"
    rc: 0
  - run: ["calc"]
    rc: 2
    err: "usage: calc <a> <b> + [--help]\n"
YAML
  cat > d2.json <<'JSON'
{"kind":"code","submission-key":"impl-001","card":"CRD-001","files":[{"path":"calc","content":"#!/usr/bin/env python3\nimport sys\nif len(sys.argv) != 4:\n    print(\"usage: calc <a> <b> + [--help]\", file=sys.stderr)\n    sys.exit(2)\na, b = int(sys.argv[1]), int(sys.argv[2])\nprint(a + b)\n"}]}
JSON
  cat > d3.yaml <<'YAML'
kind: feature
submission-key: feature-001
entries:
  - requirement: REQ-001
    text: |
      简单计算器 calc：calc <数> <数> + 打印两数之和；无参数时
      在 stderr 打印用法（含 --help）并以退出码 2 拒绝。
YAML
  local ok=0
  step 0 "AGENTS.md in full" -- next "简单计算器 calc：calc <数> <数> + 打印两数之和；calc 无参数时打印用法（含 --help）到 stderr 并以退出码 2 拒绝" || ok=1
  submit_delta 0 "suite=red" -- d1.yaml || ok=1
  submit_delta 0 "suite=green" -- d2.json || ok=1
  submit_delta 0 "VERDICT: READY" -- d3.yaml || ok=1
  local summ="$probe_dir/.punchtape/SUMMARY.md"
  if [ -f "$summ" ]; then
    {
      echo "### summary"
      cat "$summ"
    } >> "$transcript"
  else
    echo "### summary MISSING" >> "$transcript"
    echo "  STEP FAIL: SUMMARY.md not rendered"
    ok=1
  fi
  [ "$ok" = 0 ] && finish_probe || { echo "PROBE zh: FAIL (step assertions)"; overall=1; }
}

echo "== synthetic probes: ${PROBES[*]} (binary: $BIN)"
for p in "${PROBES[@]}"; do
  case "$p" in
    rpn) run_rpn ;;
    translit) run_translit ;;
    dice) run_dice ;;
    tick) run_tick ;;
    bin) run_bin ;;
    zh) run_zh ;;
    *) echo "unknown probe: $p"; overall=1 ;;
  esac
done

if [ "$RECORD" = 1 ]; then
  echo "== baselines recorded to $BASELINE_DIR (review the diff before committing)"
fi
[ "$overall" = 0 ] && echo "== ALL PROBES PASS" || echo "== PROBES FAILED"
exit "$overall"
