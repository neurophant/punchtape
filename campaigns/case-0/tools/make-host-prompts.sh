#!/bin/bash
# make-host-prompts.sh — the NAMED world adaptation of the 0.2.0
# prompt bytes: prompts/ stays byte-frozen 0.2.0; prompts-host/
# carries the minimal host-world renders. Every substitution is an
# EXACT anchor replace — a drifted source byte fails the render (no
# silent adaptation drift). The table of substitutions lives here
# and in README.md; PROVENANCE.json records source/anchor/result
# digests for the freeze.
#
# Substitutions (each named):
#   world-host        HARNESS.md: the container/docker-exec bullets
#                     -> the host cell bullets (__WORK__/__INPUT__
#                     placeholders) + the filesystem walls (every
#                     command runs as cd __WORK__ && . ./ENV &&
#                     sh -lc '...' — the cell's own exports ride
#                     work/ENV, written by prep-cell; ALL scratch
#                     lives inside __WORK__ (TMPDIR is in ENV);
#                     /tmp, $HOME, the repo and every other host
#                     directory are write-forbidden)
#   contract-paths    CONTRACT-MACHINE.md + OPERATOR-MACHINE.md:
#                     /run/task... -> __WORK__/__INPUT__
#   measurer-host     MEASURER.md: containers -> fresh host dirs;
#                     the 0.2.0 canon path (canon/scenarios) -> the
#                     corpus (canon/atoms); the runner path -> the
#                     campaign's own copy (__CASERUNNER__)
#   two-hands         MEASURER.md: three products -> TWO products
#                     (P1 solo / P2 machine) — the owner's word: no
#                     third hand, no flash; this repo's reality is
#                     two hands, as in every other campaign
#   economy           CONTRACT-MACHINE.md: the mandatory dry
#                     validator before EVERY submission (spec
#                     --check); files instead of dialog (drafts and
#                     tables in the drafts/ directory, declared
#                     build-ignore in the FIRST spec — a draft in
#                     the accounted tree moves the digest and ratify
#                     refuses "tree moved"); the dead kind: feature
#                     spelling -> the patch prose section (the
#                     machine speaks exactly two kinds)
#   stable-rows       CONTRACT-MACHINE.md: after the first accepted
#                     spec every submission is a SPARSE PATCH by row
#                     address (the full table never rides twice;
#                     full-replace: true is the only deliberate
#                     rewrite); the accepted row order never changes
#                     (a repair edits its row in place, new rows go
#                     to the END: the platform cache re-reads the
#                     same bytes)
#   form-hints        CONTRACT-MACHINE.md + tools/form-hints.py +
#                     stage/verify/freeze/check-staged-freeze: the
#                     mechanical form-hints sidcar staged to machine
#                     cells only — verbatim brief lines a form pin
#                     can quote (a formal pin wants a source line;
#                     the machine itself never sees the brief)
#
# Usage: make-host-prompts.sh [--verify]
set -euo pipefail
HERE="$(cd "$(dirname "$0")" && pwd)"
CAMP="$(cd "$HERE/.." && pwd)"
MODE="run"; [ "${1:-}" = "--verify" ] && MODE="verify"

python3 - "$CAMP" "$MODE" <<'PY'
import hashlib, json, os, sys

camp, mode = sys.argv[1:3]

def render(name, subs):
    src = open(os.path.join(camp, "prompts", name), encoding="utf-8").read()
    text = src
    for anchor, repl in subs:
        if anchor not in text:
            print(f"MAKE-HOST-PROMPTS STOP: anchor of a named substitution "
                  f"not found in prompts/{name}", file=sys.stderr)
            print("  anchor head: " + anchor[:60].replace("\n", "\\n"), file=sys.stderr)
            sys.exit(2)
        text = text.replace(anchor, repl)
    return src, text

SUB_HARNESS = [(
"""- Your working container __CONTAINER__ is already running; every
  command goes through docker exec (example:
  docker exec -w /run/task __CONTAINER__ sh -lc '<command>').
- The network in the container is off; do not try to reach anything
  outside it.
- On the host the directory __HOSTDIR__ is the mounted /run/task:
  read and write only it and nothing else on the host.""",
"""- Your working directory __WORK__ is already prepared on this
  machine; run EVERY command exactly so:
  cd __WORK__ && . ./ENV && sh -lc '<command>'
  (the ENV file with this cell's environment exports already lies
  in __WORK__; do not run a single command without this prefix).
- The network is off by directive; do not try to reach anything
  outside the cell.
- The task's inputs lie in __INPUT__ and are read only (the brief
  brief.md, the materials materials/, the expectations list
  EXPECTATIONS.md).
- WRITE BOUNDS (absolute): write ONLY inside your own cell — the
  working directory __WORK__ and the temporary directory TMPDIR
  from ENV; every file of any program of yours (builds, tests,
  scratches, caches) lives there. Writing to /tmp, to the home
  directory, to the repository directory, to any other host
  directory — forbidden and tracked by the audit; the foot must
  not leave the cell by a single byte.""",
)]

SUB_CONTRACT_MACHINE = [
("""Work in /run/task through the framework. The brief (brief.md) is the
only input; read it in full and read AGENTS.md (the framework's
self-doc) before the first call.""",
"""Work in __WORK__ through the framework. The brief (in
__INPUT__/brief.md) is the only input; read it in full and read
AGENTS.md (the framework's self-doc) before the first call."""),
("""- The materials in /run/task/materials are read-only inputs:""",
"""- The materials in __INPUT__/materials are read-only inputs:"""),
]

SUB_OPERATOR_MACHINE = [(
"""You are the operator of the experiment's machine hand. The executor
works in /run/task through the punchtape framework""",
"""You are the operator of the experiment's machine hand. The executor
works in __WORK__ through the punchtape framework""")]

SUB_MEASURER = [
("""suite: the hand's exported case table
(.punchtape/canon/scenarios/*.yaml — records: seed [{path,content}],
materials [{from}], pre [argv], when.command [argv], then
[expectations]; the single runner campaigns/case-0/tools/case-runner-host.py).
Material: <path to the meaning materials> (read-only). Image:
punchtape-world.""",
"""suite: the hand's exported corpus
(.punchtape/canon/atoms/*.yaml — behaviors with rows: seeds
[{path,text}|{path,from-input}], run [argv], out/err/rc/state
[expectations]; the single runner __CASERUNNER__ — the campaign's
own copy). Material: <path to the meaning materials> (read-only).
World: host cells (toolchain versions — the cell's env/VERSIONS.txt)."""),
("""1. Suite×2 on its own product in temporary containers
   (--network none):""",
"""1. Suite×2 on its own product in temporary directories
   (the network is off):"""),
("""   - P1: cd solo-product && sh check.sh — twice in fresh containers,""",
"""   - P1: cd solo-product && sh check.sh — twice in fresh directories,"""),
("""     run when.command, timeout 15, check then: exit/stdout/
     stderr/disk by exact comparison;""",
"""     run the row's command (run), the row's timeout, check
     the expectations: exit/stdout/stderr/disk by exact comparison;"""),
# ONE trailing \n is the replay method's own normalization — a
# strict pass, not a tolerant marker; the runner
# (case-runner-host.py) implements the same law
("""a difference only in trailing
     \\n/spaces — pass with the tolerant mark.""",
"""one trailing \\n is normalized by the replay method itself
     (a strict pass, without the tolerant mark — it is a method of
     measurement, not a leniency)."""),
# the machine cells' coverage deliverable is by construction (every
# brief behavior a green corpus row, certified by the frozen
# verdict) — the method line separates the sources
("""   coverage (COVERAGE.md / specdocs binding [SCN-xxx]). A missing one —
   a reason line.""",
"""   coverage (P1: COVERAGE.md / specdocs binding [SCN-xxx]; P2:
   the machine binding by construction — every brief behavior is
   a corpus row, the frozen READY verdict in .punchtape/verdicts/
   and a green replay ×2 from item 1 close the deliverable; the
   method line separates the sources: machine-verdict-coverage).
   A missing one — a reason line."""),
]

# two hands (the owner's word: no third hand, no flash): the
# measurer judges TWO products of a task — P1 the solo product, P2
# the machine hand's product; P3 is gone with the dead weak hand
SUB_TWO_HANDS = [
("""Three products (host): <paths to the task's hand directories>. The
products: P1 — solo-product/ (suite solo-product/check.sh); P2/P3 —
the product at the root,""",
"""Two products (host): <paths to the task's hand directories>. The
products: P1 — solo-product/ (suite solo-product/check.sh); P2 —
the product at the root,"""),
("""   - P2/P3: the single runner over the case records: a fresh directory per case,""",
"""   - P2: the single runner over the case records: a fresh directory per case,"""),
("""End the turn with a summary over the three products: suite×2 (run
counts),""",
"""End the turn with a summary over the two products: suite×2 (run
counts),"""),
]

# economy: the dry validator before every submission; files instead
# of dialog (drafts/ + the first spec's build-ignore: a draft in the
# accounted tree moves the digest and ratify refuses "tree moved");
# the dead kind: feature prose spelling -> the patch prose section
SUB_ECONOMY = [
("""- Loop in next/submit until a reply carries the VERDICT line — that
  line is the end. Submit your proposals as delta.yaml files, each
  file ends with the tail: # TOKENS n/a""",
"""- Loop in next/submit until a reply carries the VERDICT line — that
  line is the end. Submit your proposals as delta.yaml files, each
  file ends with the tail: # TOKENS n/a
- NOT ONE submission past the dry check: before EVERY submission —
  punchtape spec --check delta.yaml (free, without the machine);
  put the file into the door only after a clean answer. A rejected
  submission is a whole turn spent.
- FILES INSTEAD OF DIALOG: keep drafts, tables and deltas in files
  in the drafts/ directory (create it first of all; every
  delta.yaml you are preparing for the door lives there too); in
  the dialog — only commands and short answers; do not reread your
  own large files in full without need (read by address, the lines
  you need): the dialog history rides with you every turn and is
  paid for on every turn.
- The FIRST spec must declare build-ignore: ["drafts/"]: the drafts
  directory lies OUTSIDE the accounted tree — files in it do not
  move the tree digest, and ratify will not refuse "tree moved"
  because of a draft that appeared between a green run and the
  signing."""),
("""- The verdict demands the author's prose: when the tests are green,
  submit kind: feature with a text for every capability (all texts
  in one submission), in living words — the human spec is part of
  done. Budget submissions for it.""",
"""- The verdict demands the author's prose: when the tests are green,
  submit texts for all capabilities in ONE submission — a prose:
  patch (one text per behavior), in living words — the human spec
  is part of done. Budget a submission for it."""),
]

# stable-rows: sparse patches by row address, the full table rides
# the first spec exactly once; the accepted row order is stable, new
# rows at the end, repairs in place
SUB_STABLE_ROWS = [(
"""- Submit tables whole: the case table is one submission (split it
  only if it truly does not fit, never into shards); repairs — by
  replacing rows, not with new partial tables.""",
"""- The FIRST spec is the full table, and that is the only time: all
  further submissions are SPARSE PATCHES by row address (replace/
  add/remove/run/prose — the grammar in AGENTS.md); the machine
  rejects a full repeat of the table, a deliberate full regrouping
  — only with full-replace: true.
- DO NOT CHANGE THE ORDER OF ACCEPTED ROWS: a repair replaces its
  row in place, new rows go only to the END of the table, do not
  rearrange accepted rows (the platform cache rereads the same
  bytes — that is many times cheaper).""")]

# form-hints: the machine cell's mechanical sidcar of verbatim brief
# lines for form citations; the anchor rides the stable-rows render —
# a drift of either fails the render loudly
SUB_FORM_HINTS = [(
"""  row in place, new rows go only to the END of the table, do not
  rearrange accepted rows (the platform cache rereads the same
  bytes — that is many times cheaper).""",
"""  row in place, new rows go only to the END of the table, do not
  rearrange accepted rows (the platform cache rereads the same
  bytes — that is many times cheaper).
- Form pins (exact flags, strings, formats, codes) require a form
  citation verbatim from the brief: ready citation lines with
  their line numbers lie in __INPUT__/form-hints.md (generated
  from your own brief); a pin without an ascribed citation closes
  the run door.""")]

RENDERS = [
    ("HARNESS.md", "HARNESS-HOST.md", "world-host", SUB_HARNESS),
    ("CONTRACT-MACHINE.md", "CONTRACT-MACHINE-HOST.md",
     "contract-paths+economy+stable-rows+form-hints",
     SUB_CONTRACT_MACHINE + SUB_ECONOMY + SUB_STABLE_ROWS
     + SUB_FORM_HINTS),
    ("OPERATOR-MACHINE.md", "OPERATOR-MACHINE-HOST.md", "contract-paths",
     SUB_OPERATOR_MACHINE),
    # the measurer carries BOTH adaptations: the host world AND the
    # two-hand reality (the adaptation field names them in order)
    ("MEASURER.md", "MEASURER-HOST.md", "measurer-host+two-hands",
     SUB_MEASURER + SUB_TWO_HANDS),
]

sha = lambda b: hashlib.sha256(b.encode("utf-8")).hexdigest()
prov = []
bad = False
for src_name, dst_name, adapt_name, subs in RENDERS:
    src_text, text = render(src_name, subs)
    dst = os.path.join(camp, "prompts-host", dst_name)
    entry = {
        "source": f"prompts/{src_name}", "render": f"prompts-host/{dst_name}",
        "adaptation": adapt_name,
        "source_sha256": sha(src_text),
        "anchors_sha256": [sha(a) for a, _ in subs],
        "render_sha256": sha(text),
    }
    if mode == "verify":
        if not os.path.isfile(dst) or open(dst, encoding="utf-8").read() != text:
            print(f"MAKE-HOST-PROMPTS VERIFY: DRIFT prompts-host/{dst_name}",
                  file=sys.stderr)
            bad = True
        else:
            print(f"  ok  prompts-host/{dst_name} == render of prompts/{src_name}")
    else:
        os.makedirs(os.path.dirname(dst), exist_ok=True)
        with open(dst, "w", encoding="utf-8") as f:
            f.write(text)
        print(f"  rendered prompts-host/{dst_name} <- prompts/{src_name} ({adapt_name})")
    prov.append(entry)

if mode == "verify":
    sys.exit(1 if bad else 0)

out = os.path.join(camp, "prompts-host", "PROVENANCE.json")
with open(out, "w", encoding="utf-8") as f:
    json.dump({"law": "named minimal world adaptations over the "
                      "byte-frozen 0.2.0 prompts",
               "renders": prov}, f, ensure_ascii=False, indent=1, sort_keys=True)
    f.write("\n")
print(f"provenance: prompts-host/PROVENANCE.json ({len(prov)} renders)")
PY
