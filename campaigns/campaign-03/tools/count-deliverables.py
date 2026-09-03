#!/usr/bin/env python3
"""count-deliverables.py — mechanical deliverable counting over an
exported th3 task dir. Same script for every hand; the counting
METHOD is recorded beside every count.
Usage: count-deliverables.py <task-dir> <stack> <hand>
Prints JSON. Nothing is invented: what cannot be detected mechanically
is reported as not-detected with the method string saying so."""

import json, os, re, sys

def machine_root(root):
    for name in (".punchtape", "punchtape"):
        p = os.path.join(root, name)
        if os.path.isdir(p):
            return p
    return None

def ledger_facts(mroot):
    out = {"submissions_accepted": None, "submissions_rejected": None,
           "verdict": None, "verdict_digest": None, "checks_green": None}
    led = None
    for cand in ("ledger.yamll", os.path.join("punchtape", "ledger.yamll")):
        p = os.path.join(mroot, cand) if cand.startswith("punchtape") else os.path.join(mroot, cand)
        if os.path.isfile(p):
            led = p
            break
    if led:
        text = open(led, encoding="utf-8").read()
        acc = len(re.findall(r"name: submit\n\s+outcome: accepted", text))
        rej = len(re.findall(r"name: submit\n\s+outcome: rejected", text))
        out["submissions_accepted"], out["submissions_rejected"] = acc, rej
    vt = os.path.join(mroot, "cache", "verdict.txt")
    if os.path.isfile(vt):
        m = re.search(r"VERDICT: (\S+).*?digest: (\S+)", open(vt, encoding="utf-8").read(), re.S)
        if m:
            out["verdict"], out["verdict_digest"] = m.group(1), m.group(2)
        else:
            out["verdict"] = open(vt, encoding="utf-8").read().strip().splitlines()[0][:80]
    s = os.path.join(mroot, "SUMMARY.md")
    if os.path.isfile(s) and not out["checks_green"]:
        m = re.search(r"checks: green (\d+), red (\d+), total (\d+)", open(s, encoding="utf-8").read())
        if m:
            out["checks_green"] = f"{m.group(1)}/{m.group(3)}"
    return out

TEST_PATTERNS = {
    "python": re.compile(r"^\s*def (test_\w+)", re.M),
    "go": re.compile(r"^func (Test\w+)", re.M),
    "rust": re.compile(r"#\[(?:tokio::)?test\]", re.M),
    "js": re.compile(r"^\s*(?:test|it)\(\s*[\"'`]", re.M),
    "cpp": re.compile(r"^\s*TEST(?:_F)?\(", re.M),
}

def solo_spec(root):
    best, best_n, method = None, -1, "not-detected"
    for name in sorted(os.listdir(root)):
        if not name.lower().endswith(".md") or name.lower() in ("brief.md", "self-report.md"):
            continue
        p = os.path.join(root, name)
        if not os.path.isfile(p):
            continue
        n = len(re.findall(r"^## ", open(p, encoding="utf-8").read(), re.M))
        if n > best_n:
            best, best_n, method = n, n, f"md-sections({name})"
    return (best_n if best_n >= 0 else None), method

def main():
    root, stack, hand = sys.argv[1], sys.argv[2], sys.argv[3]
    d = {"specs_sections": None, "specs_method": None,
         "scenarios": None, "scenarios_method": None,
         "tests": None, "tests_method": None,
         "double_run_evidence": "not-detected-mechanically",
         "machine": None}
    mroot = machine_root(root)
    if hand in ("top", "weak") and mroot:
        specs = [f for f in os.listdir(os.path.join(mroot, "specs")) if re.match(r"REQ-.*\.md$", f)] \
            if os.path.isdir(os.path.join(mroot, "specs")) else []
        d["specs_sections"], d["specs_method"] = len(specs), "machine-spec-files"
        scn_dir = os.path.join(mroot, "canon", "scenarios")
        d["scenarios"] = len([f for f in os.listdir(scn_dir) if f.endswith(".yaml")]) if os.path.isdir(scn_dir) else 0
        d["scenarios_method"] = "canon-scenario-files"
        chk_dir = os.path.join(mroot, "canon", "checks")
        d["tests"] = len([f for f in os.listdir(chk_dir) if f.endswith(".yaml")]) if os.path.isdir(chk_dir) else 0
        d["tests_method"] = "canon-check-files (the machine's checks are the suite)"
        d["double_run_evidence"] = "machine-native (double-run is the machine's core invariant; ledger present)"
        d["machine"] = ledger_facts(mroot)
    else:
        n, method = solo_spec(root)
        d["specs_sections"], d["specs_method"] = n, method
        pat = TEST_PATTERNS.get(stack)
        total, files = 0, []
        if pat:
            for dirpath, _dirs, files_ in os.walk(root):
                if ".punchtape" in dirpath or "materials" in dirpath:
                    continue
                for f in files_:
                    if f.endswith((".py", ".go", ".rs", ".js", ".mjs", ".cpp", ".cc", ".h", ".hpp")):
                        p = os.path.join(dirpath, f)
                        total += len(pat.findall(open(p, encoding="utf-8", errors="replace").read()))
                        files.append(f)
        d["scenarios"] = total if total else None
        d["scenarios_method"] = f"test-pattern({stack})" if total else "not-detected"
        d["tests"], d["tests_method"] = d["scenarios"], d["scenarios_method"]
        # two recorded full-suite runs: run logs (any file whose name or
        # content header says run/twice/suite), reported as candidates
        hits = []
        for dirpath, _dirs, files_ in os.walk(root):
            if ".punchtape" in dirpath or "materials" in dirpath:
                continue
            for f in files_:
                if re.search(r"run|twice|suite", f, re.I) and not f.startswith("brief"):
                    hits.append(f)
        d["double_run_evidence"] = f"candidate-run-logs({', '.join(sorted(set(hits))[:4])})" if hits else "not-detected-mechanically"
    print(json.dumps(d, indent=1, ensure_ascii=False))

if __name__ == "__main__":
    main()
