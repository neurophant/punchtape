#!/bin/bash
# collect-metrics.sh — the per-cell evidence capture of the
# case-0 campaign (the 0.2.0 METRICS discipline): GUARANTEE
# MECHANISM, not a promise — every
# field comes from a named mechanical source (the platform db ledger
# of the registered agents, the machine's own instance ledger, the
# cell tree, the broker's wait journal); a missing source FAILS the
# cell. The hand's and the operator's metrics are computed and
# reported SEPARATELY (the owner's hard rule). Writes the 0.2.0
# per-cell file set into tasks/<cell>/evidence/ (the report reads
# exactly these): meta.json, platform.json, operator.json,
# summary.json, evidence.json, findings.json (+ measure.json and
# judge.json copied in when the measurer/judge of the task wrote
# them), and metrics.json (the cell's own assembly).
#
# Usage: collect-metrics.sh <run-root> <cell> [--accept "<note>"]
#                           [--verdict-from-machine]
# Env: C0_DB (default ~/.zcode/cli/db/db.sqlite)
set -euo pipefail
RR="${1:?run root}"; CELL="${2:?cell}"; shift 2
ACCEPT=""; VFM=""
while [ $# -gt 0 ]; do
  case "$1" in
    --accept) ACCEPT="$2"; shift 2 ;;
    --verdict-from-machine) VFM=1; shift ;;
    *) echo "unknown: $1" >&2; exit 2 ;;
  esac
done
HERE="$(cd "$(dirname "$0")" && pwd)"
RR="$(readlink -f "$RR")"
DB="${C0_DB:-$HOME/.zcode/cli/db/db.sqlite}"

python3 - "$RR" "$CELL" "$HERE" "$DB" "$ACCEPT" "$VFM" <<'PY'
import datetime, glob, hashlib, json, os, re, subprocess, sys

root, cell, here, db, accept, vfm = sys.argv[1:7]
man = json.load(open(f"{root}/manifests/{cell}.json", encoding="utf-8"))
work = os.path.join(man["cell_path"], "work")
inp = os.path.join(man["cell_path"], "input")
hand = man["hand"]
sense, stack = man["sense"], man["stack"]

def fail(msg):
    print(f"COLLECT {cell} FAIL: {msg}", file=sys.stderr)
    sys.exit(3)

extra_problems = []

def jload(p):
    try:
        return json.load(open(p, encoding="utf-8"))
    except Exception:
        return None

def sha(p):
    return hashlib.sha256(open(p, "rb").read()).hexdigest()

# --- the registry: this cell's agents (keyed by the same id8 the
# --- ledger uses: the first 8 hex chars of the uuid)
reg = jload(f"{root}/driver-agents.json") or {}
cell_reg = reg.get(cell) or []
def id8(u):
    m = re.search(r"([0-9a-fA-F]{8})", u or "")
    return m.group(1).lower() if m else None
hand_ids = [id8(e["uuid"]) for e in cell_reg if e.get("role") == "hand"]
op_ids = [id8(e["uuid"]) for e in cell_reg if e.get("role") == "operator"]
hand_ids = [i for i in hand_ids if i]
op_ids = [i for i in op_ids if i]
if not hand_ids:
    fail("no hand agent registered for the cell")

# --- the platform ledger slices (usage-ledger.py --root output)
ledger = (jload(f"{root}/usage/ledger.json") or {}).get("agents") or {}
def agg(ids):
    out = {"input_tokens": 0, "output_tokens": 0, "cache_tokens": 0,
           "active_s": 0.0, "tool_calls": 0, "requests": 0, "model": None,
           "validation": "ok"}
    found = False
    for i in ids:
        e = ledger.get(i)
        if not e:
            continue
        found = True
        for k in ("input_tokens", "output_tokens", "cache_tokens",
                  "tool_calls", "requests"):
            out[k] += e.get(k) or 0
        out["active_s"] = round(out["active_s"] + (e.get("active_s") or 0), 1)
        out["model"] = out["model"] or e.get("model")
        if e.get("validation") != "ok":
            out["validation"] = e.get("validation")
    return out if found else None

hagg = agg(hand_ids)
oagg = agg(op_ids)
if hagg is None:
    fail("the hand's platform ledger slice is missing (run usage-ledger.py --root first)")

# --- the machine telemetry (machine hands): parse-ledger.py owns the
# --- verb numbers; the verdict is the machine's own immutable file
led_path = os.path.join(work, ".punchtape", "ledger.yamll")
machine = {"present": os.path.isfile(led_path), "verdict": None,
           "submissions_accepted": 0, "submissions_rejected": 0,
           "verb_wall_ms": 0, "submit_p95_ms": None,
           "source": "no machine by design (solo hand)"}
if machine["present"]:
    r = subprocess.run([sys.executable, os.path.join(here, "parse-ledger.py"),
                        "--sum", led_path], capture_output=True, text=True)
    # parse-ledger's aggregate line: files=N verbs=N submissions=A/R
    #   verb-wall=W.Ws submit-p95=Pms reply=K.KKB
    mm = re.search(r"submissions=(\d+)/(\d+) verb-wall=([\d.]+)s submit-p95=(\d+)ms", r.stdout or "")
    if mm:
        machine.update({
            "submissions_accepted": int(mm.group(1)),
            "submissions_rejected": int(mm.group(2)),
            "verb_wall_ms": int(float(mm.group(3)) * 1000),
            "submit_p95_ms": int(mm.group(4)),
            "source": "the instance ledger (the machine measures itself)",
        })
    else:
        machine["source"] = "the instance ledger (unparsable aggregate line — see metrics problems)"
        extra_problems.append(f"parse-ledger output not understood: {(r.stdout or '')[:120]!r}")
    vdir = os.path.join(work, ".punchtape", "verdicts")
    vfiles = sorted(glob.glob(os.path.join(vdir, "v*.txt")))
    if vfiles:
        vt = open(vfiles[-1], encoding="utf-8").read()
        vm = re.search(r"^VERDICT: .*$", vt, re.M)
        if vm:
            machine["verdict"] = vm.group(0)

# --- the corpus facts (machine): atoms and rows
atoms_dir = os.path.join(work, ".punchtape", "canon", "atoms")
cases_total = cases_green = 0
describes = capabilities = 0
if os.path.isdir(atoms_dir):
    import yaml
    for n in sorted(os.listdir(atoms_dir)):
        if not n.endswith(".yaml"):
            continue
        a = yaml.safe_load(open(os.path.join(atoms_dir, n), encoding="utf-8")) or {}
        capabilities += 1
        if (a.get("prose") or "").strip():
            describes += 1
        rows = a.get("rows") or []
        for row in rows:
            cases_total += 1
            if row.get("outcome") == "green":
                cases_green += 1

# --- LOC (code lines by extension, machine dirs excluded; an
# --- extensionless EXECUTABLE that is not binary is code too — an
# --- ESM module printed LOC=0 once because the counter only
# --- globbed extensions)
EXTS = (".py", ".go", ".rs", ".c", ".cc", ".cpp", ".h", ".hpp", ".js",
        ".mjs", ".ts", ".sh")
loc = 0
prod_root = os.path.join(work, "solo-product") if hand == "solo" else work
for base, dirs, files in os.walk(prod_root):
    dirs[:] = [d for d in dirs if d not in (".punchtape", "node_modules",
                                            "target", "materials", "scratch")]
    for f in files:
        p = os.path.join(base, f)
        is_code = f.endswith(EXTS)
        if not is_code and "." not in f and os.access(p, os.X_OK):
            try:
                with open(p, "rb") as fh:
                    is_code = b"\x00" not in fh.read(1024)
            except OSError:
                is_code = False
        if not is_code:
            continue
        try:
            with open(p, errors="ignore") as fh:
                loc += sum(1 for _ in fh)
        except OSError:
            pass

# --- decision waits (the broker journal)
waits = jload(f"{root}/operators/decision-waits-{cell}.json") or []
decision_wait_s = [w.get("wait_s") for w in waits if w.get("wait_s") is not None]

# --- the materials after the run (unchanged is the proof)
materials = dict(man["inputs"]["materials"])
materials_after = {}
matdir = os.path.join(inp, "materials")
for base, _, files in os.walk(matdir):
    for f in sorted(files):
        p = os.path.join(base, f)
        materials_after[f"materials/{os.path.relpath(p, matdir)}"] = sha(p)

# --- the census (registry roles) as findings source
findings = []

now = datetime.datetime.now(datetime.timezone.utc).isoformat(timespec="seconds")
ev = f"{root}/tasks/{cell}/evidence"
os.makedirs(ev, exist_ok=True)
spawn_ts = None
for e in cell_reg:
    if e.get("role") == "hand":
        spawn_ts = e.get("registered")

meta = {"cell": cell, "sense": sense, "stack": stack, "hand": hand,
        "model": hagg["model"], "container": None,
        "world": "host-direct", "protocol": "0.2.0",
        "captured": now}
platform = {
    "elapsed_s": hagg["active_s"],
    "tokens_in": hagg["input_tokens"], "tokens_out": hagg["output_tokens"],
    "tokens_cache_read": hagg["cache_tokens"],
    "tokens_inout": hagg["input_tokens"] + hagg["output_tokens"],
    "tokens_total": hagg["input_tokens"] + hagg["output_tokens"] + hagg["cache_tokens"],
    "tool_calls": hagg["tool_calls"], "decision_wait_s": decision_wait_s,
    "validation": hagg["validation"],
    "source": "the platform db ledger of the hand's registered session(s)",
    "spawn_ts": spawn_ts, "finish_ts": now,
}
operator = {
    "model": oagg["model"] if oagg else None,
    "elapsed_s": oagg["active_s"] if oagg else 0,
    "tokens_in": oagg["input_tokens"] if oagg else 0,
    "tokens_out": oagg["output_tokens"] if oagg else 0,
    "tokens_cache_read": oagg["cache_tokens"] if oagg else 0,
    "decisions": len(decision_wait_s),
    "source": "the platform db ledger of the cell's operator agents (never mixed with the hand)",
} if oagg else {
    "model": None, "elapsed_s": None, "tokens_in": None, "tokens_out": None,
    "tokens_cache_read": None, "decisions": 0,
    "source": "no operator invocations recorded",
}
ruling = None
# the acceptance verdict FREEZES at the first finish (a settled-
# ledger re-collect's service note --accept once landed in the
# verdict column and erased the operator's acceptance) — the
# note is stored beside the evidence and a later --accept can only
# append a recollect record, never replace the ruling
acc_path = f"{ev}/acceptance.json"
acc_doc = jload(acc_path)
if hand == "solo":
    if acc_doc and acc_doc.get("ruling"):
        ruling = acc_doc["ruling"]
        recollects = list(acc_doc.get("recollects") or [])
        if accept and accept != acc_doc.get("note"):
            recollects.append({"note": accept, "at": now})
        json.dump({"ruling": ruling, "note": acc_doc.get("note"),
                   "recorded": acc_doc.get("recorded"),
                   "recollects": recollects},
                  open(acc_path, "w", encoding="utf-8"),
                  ensure_ascii=False, indent=1)
    else:
        ruling = accept or None
        if not ruling:
            fail("the solo hand's acceptance note is required (--accept)")
        json.dump({"ruling": ruling, "note": accept, "recorded": now,
                   "recollects": []},
                  open(acc_path, "w", encoding="utf-8"),
                  ensure_ascii=False, indent=1)
else:
    ruling = machine.get("verdict")
summary = {
    "ruling": ruling,
    "cases_pass": cases_green if hand != "solo" else None,
    "cases_total": cases_total if hand != "solo" else None,
    "tests_pass": None, "tests_total": None,
    "describes": describes if hand != "solo" else None,
    "capabilities": capabilities if hand != "solo" else None,
    "loc": loc,
    "source": "the exported corpus (atoms) of the work tree" if hand != "solo"
              else "the solo product tree (its own suite counts live in measure.json)",
}
evidence = {
    "brief_sha256": man["inputs"]["brief"]["sha256"],
    "materials": materials, "materials_after": materials_after,
    "world_image_digest": None,
    "world_versions": "world-lock.txt (the 0.2.0 source) + env/VERSIONS.txt (the live host truth)",
    "limits": "none-net-directive / host cell",
    "network_mode": "textual-directive (the post-factum audit: audit-net.sh)",
    "mounts": "none (a plain host directory; no binary copy rides the cell)",
    "punchtape_binary_sha256": man["inputs"]["binary"]["sha256"] if man["inputs"]["binary"] else None,
    "product_scan": [], "canary_leaks": [], "transcript_scan": [],
}

for name, doc in (("meta.json", meta), ("platform.json", platform),
                  ("operator.json", operator), ("summary.json", summary),
                  ("evidence.json", evidence), ("findings.json", findings)):
    json.dump(doc, open(f"{ev}/{name}", "w", encoding="utf-8"),
              ensure_ascii=False, indent=1)
# the measurer's and the judge's results ride the same directory
for src_name in ("measure.json", "judge.json"):
    src = f"{root}/measurement/{src_name.replace('.json', '')}-{cell}.json"
    if os.path.isfile(src):
        import shutil
        shutil.copy2(src, f"{ev}/{src_name}")

metrics = {"cell": cell, "hand": hand, "machine": machine,
           "platform": platform, "operator": operator,
           "summary": summary, "problems": extra_problems}
if hand != "solo" and not machine.get("verdict"):
    metrics["problems"].append("machine verdict missing on a machine hand")
if hand == "solo" and machine["present"]:
    metrics["problems"].append("machine present on the solo hand")
json.dump(metrics, open(f"{ev}/../metrics.json", "w", encoding="utf-8"),
          ensure_ascii=False, indent=1)
print(f"collected: {ev} (problems={metrics['problems']})")
sys.exit(0 if not metrics["problems"] else 4)
PY
