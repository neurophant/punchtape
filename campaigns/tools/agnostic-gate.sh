#!/bin/bash
# agnostic-gate.sh (extended) — prove the machine is task-agnostic:
# machine code, code comments AND canon texts must carry no battery
# vocabulary, no campaign-grading references, no non-synthetic
# example nouns, and no locale drift.
#
# Layers (each prints [name] clean/FAIL; exit 1 on any hit):
#   sense       — battery task-sense words (as before)
#   stack       — stack vocabulary (as before)
#   stack-nearmiss — the near-miss dictionary: `node <script>`
#              invocations and .mjs/.cjs extensions. Bare `node` is
#              NOT matched: `node` is a legitimate Go identifier for
#              yaml.Node values (delta.go et al.) — the fossil class
#              is the INVOCATION, not the word. Bare `go` stays
#              excluded (the everyday verb, e.g. intents.yaml
#              "go on").
#   batref      — campaign-grading references in machine code:
#              M-<n>, F-<n>, wave names (the generic hyphenated
#              wave-name form wave-<name>, added by the phase-C
#              audit: naming by pattern, not by generation, so the
#              layer does not go stale), grading
#              counts n/m with any side >= 2 digits ("15/15",
#              "37/45"; single-digit forms like rc "3/4" or exit
#              "0/1" are alternative notation, not grading).
#              Machine code cites fixes by commit, never by id.
#   nouns       — canondata example nouns must belong to the
#              synthetic set; scanned on the VALUE
#              side of canondata lines only (key names like
#              features.out.stdout.json-equals are machine
#              vocabulary, not examples).
#   locale      — every key with .ru has .zh and vice versa,
#              plus a drift blacklist of known-wrong locale
#              forms (the known drift rows; entries are added in the
#              fix commits that define the corrected wording).
#
# Self-test: agnostic-gate-selftest.sh plants each violation class
# in a scratch tree and asserts this gate FAILS on it — a gate that
# cannot fail proves nothing.
#
# Usage: agnostic-gate.sh [ROOT]   (default: the repository root;
#       ROOT lets the self-test scan a planted scratch tree with
#       cmd/ and internal/canondata/data/ subdirs.)

set -euo pipefail

ROOT="${1:-}"
if [ -z "$ROOT" ]; then
  ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
fi
cd "$ROOT"

# Task-sense vocabulary (the battery's five senses).
SENSE_RE='\b(conduit|pressmark|errand|brim|quire)\b'
# Stack/toolchain vocabulary (language runtimes, build tools,
# package managers — anything a task world pins).
STACK_RE='\b(python3?|golang|cargo|rustc|rust-lang|nodejs|node_modules|npm|GCC|g\+\+|clang|GraphicsMagick|PIL|Pillow|numpy|redis|jimp|jinja|pydantic|slugify|markdown|md4c|tera|inja|ejs)\b|GOPROXY|GOMEMLIMIT|GOFLAGS'
# The near-miss dictionary: node INVOKED on a script + the bare
# extensions (the `nodejs|node_modules` gap is how `node app.mjs`
# passed the old gate).
STACK_NEARMISS_RE='\bnode[[:space:]]+[^\s"'"'"']+\.(mjs|cjs|js)\b|\.(mjs|cjs)\b'
# Campaign-grading references: fix/milestone ids, wave names (the
# generic hyphenated form — naming by pattern, not by generation,
# so the layer does not go stale again), grading counts (n/m with
# any side two-plus digits).
BATREF_RE='\b[MF]-[0-9]+\b|\bwave-[a-z0-9]+\b|\b[0-9]{2,}/[0-9]+\b|\b[0-9]+/[0-9]{2,}\b'

# The synthetic example set + the machine's own file
# names (index.md is the spec registry — machine vocabulary, not a
# task example). Extend ONLY by an explicit decision, never ad hoc.
NOUN_ALLOW='^(bin/tool|materials/sample\.bin|notes|tool|sample\.bin|store\.json|notes\.json|index\.md|cli|ide|api)$'

# Known-wrong locale forms (the known drift rows). Each entry: a
# fixed string that must not appear in any canondata locale value.
LOCALE_DRIFT_BLACKLIST=(
  'грани разделены'      # split: object change, not card split
  'приколоты'            # pinned: slang (закреплено is the term)
  'зазор принят'         # waive: deviation became endorsement
  'ждёт ратификации оператора'  # collapses review/ratify (en: "awaits operator review")
  '检查通过'              # zh "checks green": a state became a judgment
  '需要确认'              # zh amend block: confirmation word for assertion
)

status=0
fail() { echo "[$1] FAIL:"; echo "$2"; status=1; }
clean() { echo "[$1] clean"; }

echo "== agnostic gate (extended): machine code, comments, canon =="

# --- sense + stack + batref over machine code (all text incl. comments)
for re in "sense:$SENSE_RE" "stack:$STACK_RE" "stack-nearmiss:$STACK_NEARMISS_RE" "batref:$BATREF_RE"; do
  name="${re%%:*}"; pat="${re#*:}"
  if [ -d cmd ] || [ -d internal ]; then
    hits="$(grep -rniE "$pat" cmd/ internal/ go.mod go.sum 2>/dev/null || true)"
  else
    hits=""
  fi
  if [ -n "$hits" ]; then fail "$name" "$hits"; else clean "$name"; fi
done

# --- nouns + locale completeness + drift blacklist (python helpers:
#     value-side extraction and per-file key sets are awk-fragile)
readarray -t canon_reports < <(python3 - "$ROOT" "$NOUN_ALLOW" "${LOCALE_DRIFT_BLACKLIST[@]}" <<'PY'
import re, glob, os, sys
root = sys.argv[1]
allow = re.compile(sys.argv[2])
blacklist = sys.argv[3:]
noun_re = re.compile(r'(?:\b(?:bin|src|materials)/[a-z][a-z0-9._-]*|\b[a-z][a-z0-9_-]*\.(?:md|mjs|cjs|js|py|bin)\b)')
for f in sorted(glob.glob(os.path.join(root, 'internal/canondata/data/*.yaml'))):
    name = os.path.basename(f)
    keys = set()
    for i, line in enumerate(open(f, encoding='utf-8'), 1):
        m = re.match(r'^([A-Za-z0-9_.-]+(\.ru|\.zh)):', line)
        if m:
            keys.add(m.group(1))
        if not line.lstrip().startswith('#'):
            # value side of a key line; block-scalar continuation
            # lines (no key) are scanned whole — the agentsref
            # `bin/serve` sites live in block scalars
            m = re.match(r'^([A-Za-z0-9_.-]+):', line)
            value = line.split(':', 1)[1] if m else line
            for noun in noun_re.findall(value):
                if not allow.match(noun):
                    print(f"NOUN|{name}:{i}| example noun '{noun}' outside the synthetic set")
    ru = {k[:-3] for k in keys if k.endswith('.ru')}
    zh = {k[:-3] for k in keys if k.endswith('.zh')}
    for k in sorted(ru - zh):
        print(f"LOCALE|{name}| {k}: has .ru, missing .zh")
    for k in sorted(zh - ru):
        print(f"LOCALE|{name}| {k}: has .zh, missing .ru")
    text = open(f, encoding='utf-8').read()
    for bad in blacklist:
        if bad in text:
            for i, line in enumerate(text.splitlines(), 1):
                if bad in line:
                    print(f"DRIFT|{name}:{i}| drift form '{bad}'")
PY
)

noun_hits=""; locale_hits=""; drift_hits=""
for rep in "${canon_reports[@]}"; do
  [ -z "$rep" ] && continue
  kind="${rep%%|*}"; rest="${rep#*|}"
  case "$kind" in
    NOUN)   noun_hits+="$rest"$'\n' ;;
    LOCALE) locale_hits+="$rest"$'\n' ;;
    DRIFT)  drift_hits+="$rest"$'\n' ;;
  esac
done
if [ -n "$noun_hits" ]; then fail "nouns" "$noun_hits"; else clean "nouns"; fi
if [ -n "$locale_hits" ]; then fail "locale" "$locale_hits"; else clean "locale"; fi
if [ -n "$drift_hits" ]; then fail "locale-drift" "$drift_hits"; else clean "locale-drift"; fi

if [ "$status" = 0 ]; then
  echo "GATE: PASS — machine code, comments and canon are task-agnostic"
fi
exit "$status"
