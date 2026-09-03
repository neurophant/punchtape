#!/bin/bash
# canon-mirror.sh — the mechanical mirror reconciliation the canon
# loader promises (canondata.go: an unknown text key "is deterministic
# and caught by mirror reconciliation before a release, so an honest
# fail-fast"): every canon identifier the machine code references by
# LITERAL must exist in the delivery registries. The panic classes of
# the loader are checked — T/TFor text keys, LintMessage rule ids,
# EnvDiag entry ids, Limit thresholds. Heal and Pain answer empty on
# an absent id (no panic by design) and are not checked. A literal
# followed by "+" is a composed-key prefix: it must be a prefix of at
# least one registry key (the composition site owns the closed
# vocabulary, validated on input; the mirror owns the literal side).
#
# Self-test: canon-mirror-selftest.sh plants a missing key in a
# scratch tree and asserts this check FAILS — a check that cannot
# fail proves nothing.
#
# Usage: canon-mirror.sh [ROOT]   (default: the repository root;
#       ROOT lets the self-test scan a planted scratch copy.)

set -euo pipefail

ROOT="${1:-}"
if [ -z "$ROOT" ]; then
  ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
fi
cd "$ROOT"

python3 - "$ROOT" <<'PY'
import re, glob, os, sys

root = sys.argv[1]

# The loader's text registries: every data/*.yaml EXCEPT the
# struct-decoded ones (canondata.go init switch).
special = {'limits.yaml','pains.yaml','scenarios.yaml','trust.yaml',
           'lint.yaml','envdiag.yaml','intents.yaml','heal.yaml',
           'hypotheses.yaml'}
yaml_keys = set()
for f in glob.glob(os.path.join(root, 'internal/canondata/data/*.yaml')):
    if os.path.basename(f) in special:
        continue
    for m in re.finditer(r'^([A-Za-z0-9_.-]+):', open(f, encoding='utf-8').read(), re.M):
        yaml_keys.add(m.group(1))

def registry_ids(path):
    out = set()
    for m in re.finditer(r'^\s*-?\s*id:\s*([A-Za-z0-9_.-]+)\s*$',
                         open(path, encoding='utf-8').read(), re.M):
        out.add(m.group(1))
    return out

lint_ids = registry_ids(os.path.join(root, 'internal/canondata/data/lint.yaml'))
envdiag_ids = registry_ids(os.path.join(root, 'internal/canondata/data/envdiag.yaml'))
limit_ids = set()
limpath = os.path.join(root, 'internal/canondata/data/limits.yaml')
for m in re.finditer(r'^([A-Za-z0-9_.-]+):', open(limpath, encoding='utf-8').read(), re.M):
    limit_ids.add(m.group(1))

# Call-site scan: qualified canondata accesses only (verified: the
# codebase has no bare T/Limit/LintMessage/EnvDiag calls outside the
# canondata package, and the package itself holds no literal calls).
call_re = re.compile(
    r'canondata\.(T|TFor|LintMessage|EnvDiagAt|EnvDiag|Limit)\s*\(\s*'
    r'(?:[A-Za-z_][A-Za-z0-9_.]*\s*,\s*)?'      # TFor's leading lang arg
    r'"([A-Za-z0-9_.-]+)"(\s*\+)?')

missing = []
gofiles = []
for pat in ('internal/**/*.go', 'cmd/**/*.go'):
    gofiles += glob.glob(os.path.join(root, pat), recursive=True)
for f in sorted(gofiles):
    src = open(f, encoding='utf-8').read()
    rel = os.path.relpath(f, root)
    for m in call_re.finditer(src):
        fn, lit, concat = m.group(1), m.group(2), m.group(3)
        line = src[:m.start()].count('\n') + 1
        if fn in ('T', 'TFor'):
            if concat:
                # composed-key prefix: must prefix at least one key
                if not any(k.startswith(lit) for k in yaml_keys):
                    missing.append(f"{rel}:{line}: composed prefix {lit!r} matches no registry key")
            elif lit not in yaml_keys:
                missing.append(f"{rel}:{line}: text key {lit!r} absent from the registries")
        elif fn == 'LintMessage' and lit not in lint_ids:
            missing.append(f"{rel}:{line}: lint rule id {lit!r} absent from lint.yaml")
        elif fn in ('EnvDiag', 'EnvDiagAt') and lit not in envdiag_ids:
            missing.append(f"{rel}:{line}: env diagnosis id {lit!r} absent from envdiag.yaml")
        elif fn == 'Limit' and lit not in limit_ids:
            missing.append(f"{rel}:{line}: limit key {lit!r} absent from limits.yaml")

print(f"canon mirror: {len(gofiles)} go files; text keys {len(yaml_keys)}, "
      f"lint rules {len(lint_ids)}, env diagnoses {len(envdiag_ids)}, "
      f"limits {len(limit_ids)}")
if missing:
    print("MIRROR: FAIL — code references canon identifiers that do not exist:")
    for m in missing:
        print("  " + m)
    sys.exit(1)
print("MIRROR: PASS — every literal canon identifier exists in the registries")
PY
