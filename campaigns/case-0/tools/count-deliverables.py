#!/usr/bin/env python3
"""count-deliverables.py — mechanical deliverable counting over ONE
PRODUCT TREE (the input is the product path,
not the task dir — the products live in the cells, and the old
task-dir input left the tool blind). Same script for every hand; the
counting METHOD is recorded beside every count.

Usage: count-deliverables.py <product-path> <stack> <hand>
  <product-path> — the product tree itself: the solo hand's product
                   dir, or the machine hand's work dir (the one with
                   .punchtape/)
  hand = solo | top
Prints JSON. Nothing is invented: what cannot be detected
mechanically is reported as not-detected with the method string
saying so.

The owner's word: no third hand, no
flash): this copy is a named ADAPTATION, not a byte transfer — the
machine branch answers "top" only; the source's dead weak branch is
not carried into this series (README, two-hands adaptation).

The machine branch reads the machine instance layout (<product>/
.punchtape/: canon/atoms/*.yaml behaviors-with-rows, ledger.yamll,
verdicts/vN.txt) — the scenario layout (specs/scenarios/checks) is gone
with the 0.2.0 world and is not waited for anymore."""

import glob
import json
import os
import re
import sys

import yaml


def ledger_facts(mroot):
    """the machine ledger: event: submit + stage: accepted|rejected."""
    out = {"submissions_accepted": None, "submissions_rejected": None,
           "verdict": None, "verdict_digest": None, "verdict_version": None}
    led = os.path.join(mroot, "ledger.yamll")
    if os.path.isfile(led):
        text = open(led, encoding="utf-8").read()
        acc = len(re.findall(r"event: submit\n\s+stage: accepted", text))
        rej = len(re.findall(r"event: submit\n\s+stage: rejected", text))
        out["submissions_accepted"], out["submissions_rejected"] = acc, rej
    # the latest frozen verdict snapshot is the accepted state's fact
    snaps = sorted(glob.glob(os.path.join(mroot, "verdicts", "v*.txt")),
                   key=lambda p: int(re.search(r"v(\d+)\.txt$", p).group(1)
                                    if re.search(r"v(\d+)\.txt$", p) else 0))
    if snaps:
        vt = open(snaps[-1], encoding="utf-8").read()
        m = re.search(r"VERDICT: (\S+)", vt)
        d = re.search(r"digest: (\S+)", vt)
        v = re.search(r"version: v(\d+)", vt)
        if m:
            out["verdict"] = m.group(1)
        if d:
            out["verdict_digest"] = d.group(1)
        if v:
            out["verdict_version"] = int(v.group(1))
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


def machine_corpus(mroot):
    """the machine corpus: behaviors and rows across canon/atoms/*.yaml."""
    d = os.path.join(mroot, "canon", "atoms")
    behaviors, rows = 0, 0
    if not os.path.isdir(d):
        return None, None
    for n in sorted(os.listdir(d)):
        if not n.endswith(".yaml"):
            continue
        doc = yaml.safe_load(open(os.path.join(d, n), encoding="utf-8"))
        if not isinstance(doc, dict):
            continue
        rb = doc.get("rows")
        behaviors += 1
        rows += len(rb) if isinstance(rb, list) else 0
    return behaviors, rows


def main():
    root, stack, hand = sys.argv[1], sys.argv[2], sys.argv[3]
    d = {"product": os.path.abspath(root),
         "specs_sections": None, "specs_method": None,
         "scenarios": None, "scenarios_method": None,
         "tests": None, "tests_method": None,
         "coverage": None, "coverage_method": None,
         "double_run_evidence": "not-detected-mechanically",
         "machine": None}
    mroot = os.path.join(root, ".punchtape")
    if hand == "top" and os.path.isdir(mroot):
        behaviors, rows = machine_corpus(mroot)
        d["specs_sections"] = behaviors
        d["specs_method"] = "canon-atom-files (behaviors of .punchtape/canon/atoms)"
        d["scenarios"] = rows
        d["scenarios_method"] = "canon-atom-rows (rows of .punchtape/canon/atoms)"
        d["tests"], d["tests_method"] = rows, d["scenarios_method"]
        # the machine-native coverage linkage:
        # a frozen READY verdict certifies every brief behavior as a
        # green corpus row — the replay of the measurer's step 1
        # exercises exactly those rows
        facts = ledger_facts(mroot)
        if facts["verdict"] == "READY":
            d["coverage"] = 1
            d["coverage_method"] = ("machine-verdict-coverage (frozen verdict "
                                    f"v{facts['verdict_version']} READY: every brief "
                                    "behavior a green corpus row)")
        else:
            d["coverage"] = 0
            d["coverage_method"] = (f"machine-verdict-coverage (latest verdict "
                                    f"{facts['verdict'] or 'none'} — not READY)")
        d["double_run_evidence"] = ("machine-native (double-run is the machine's "
                                    "core invariant; the blind replay x2 is the "
                                    "measurer's own step 1)")
        d["machine"] = facts
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
        # coverage stays the measurer's honest eye for solo products
        # (COVERAGE.md / specdocs linkage): not mechanically decidable
        d["coverage"], d["coverage_method"] = None, "measurer-eye (COVERAGE.md / specdocs linkage)"
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
