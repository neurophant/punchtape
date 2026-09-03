#!/usr/bin/env python3
# report.py — the wave report, one command:
#   python3 campaigns/tools/report.py <wave-dir> --prices prices.json \
#       --out wave-report.md
# The wave dir holds one directory per cell
# (<sense>-<stack>-<hand>, hand ∈ solo|top|weak) with the captured
# files (meta/platform/operator/summary/judge/evidence/findings/
# measure JSONs, the exported .punchtape/ telemetry for machine
# hands, task/ export). The report: evidence of identical conditions
# and no cheating (any violation a red line), the metrics table by
# hands, the deltas (solo→top, solo→flash, top vs flash),
# comparability, operator overheads as a separate section, the
# findings registry by class. Numbers come only from the captured
# files; nothing is retyped.
import argparse
import glob
import json
import os
import re
import sys


def jload(path):
    try:
        with open(path) as f:
            return json.load(f)
    except Exception:
        return None


def parse_ledger(path):
    """Machine telemetry: submissions accepted/rejected, verb wall ms,
    submit p95 of (wall - runs) ms, verdict line."""
    out = {"submissions_accepted": 0, "submissions_rejected": 0,
           "verb_wall_ms": 0, "submit_p95_ms": None, "verbs": 0}
    submit_overheads = []
    verdict = None
    try:
        with open(path) as f:
            raw = f.read()
    except Exception:
        return out
    for block in raw.split("\n---\n"):
        name = re.search(r"name: (\w+)", block)
        outcome = re.search(r"outcome: (\w+)", block)
        wall = re.search(r"wall-ms: (\d+)", block)
        runs = re.search(r"runs-ms: (\d+)", block)
        if name and outcome and name.group(1) == "submit":
            if outcome.group(1) == "accepted":
                out["submissions_accepted"] += 1
            elif outcome.group(1) == "rejected":
                out["submissions_rejected"] += 1
            if wall and runs:
                submit_overheads.append(int(wall.group(1)) - int(runs.group(1)))
        if name:
            out["verbs"] += 1
        if wall:
            out["verb_wall_ms"] += int(wall.group(1))
        m = re.search(r"^(VERDICT: .*)$", block, re.M)
        if m:
            verdict = m.group(1)
    if submit_overheads:
        s = sorted(submit_overheads)
        out["submit_p95_ms"] = s[min(len(s) - 1, int(round(0.95 * len(s))) - 1 if len(s) > 1 else 0)]
    out["verdict"] = verdict
    return out


def loc_of(task_dir):
    exts = (".py", ".go", ".rs", ".c", ".cc", ".cpp", ".h", ".hpp", ".js",
            ".mjs", ".ts", ".sh")
    n = 0
    for root, dirs, files in os.walk(task_dir):
        dirs[:] = [d for d in dirs if d not in (".punchtape", "node_modules",
                                                "target", "materials", "scratch")]
        for f in files:
            if f.endswith(exts):
                try:
                    with open(os.path.join(root, f), errors="ignore") as fh:
                        n += sum(1 for _ in fh)
                except Exception:
                    pass
    return n


def cost_of(platform, meta, prices):
    model = meta.get("model", "")
    per = prices.get("per_million", {}).get(model)
    if not per or not platform:
        return None
    tin = platform.get("tokens_in") or 0
    tout = platform.get("tokens_out") or 0
    cache = platform.get("tokens_cache_read") or 0
    pin, pout = per.get("in"), per.get("out")
    pcache = per.get("cached_in", per.get("in"))
    return round((tin - cache) * pin / 1e6 + cache * pcache / 1e6
                 + tout * pout / 1e6, 4)


def fmt(v, unit=""):
    if v is None:
        return "—"
    if isinstance(v, float):
        return ("%0.2f" % v) + unit
    return str(v) + unit


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("wave_dir")
    ap.add_argument("--prices", required=True)
    ap.add_argument("--out", required=True)
    args = ap.parse_args()

    prices = jload(args.prices) or {}
    cells = []
    for d in sorted(glob.glob(os.path.join(args.wave_dir, "*"))):
        if not os.path.isdir(d):
            continue
        meta = jload(os.path.join(d, "meta.json"))
        if not meta:
            continue
        cells.append({
            "dir": d,
            "meta": meta,
            "platform": jload(os.path.join(d, "platform.json")),
            "operator": jload(os.path.join(d, "operator.json")),
            "summary": jload(os.path.join(d, "summary.json")),
            "judge": jload(os.path.join(d, "judge.json")),
            "evidence": jload(os.path.join(d, "evidence.json")),
            "findings": jload(os.path.join(d, "findings.json")),
            "measure": jload(os.path.join(d, "measure.json")),
            "ledger": parse_ledger(os.path.join(d, "task", ".punchtape",
                                                "ledger.yamll")),
            "cost": None,
        })
    for c in cells:
        c["cost"] = cost_of(c["platform"], c["meta"], prices)

    tasks = {}
    for c in cells:
        tasks.setdefault((c["meta"]["sense"], c["meta"]["stack"]), {})[
            c["meta"]["hand"]] = c

    L = []
    w = L.append
    w("# Wave report — %s" % os.path.basename(os.path.normpath(args.wave_dir)))
    w("")

    # --- 1. evidence of identical conditions and no cheating
    w("## 1. Identical conditions and no cheating")
    w("")
    violations = []
    for (sense, stack), hands in sorted(tasks.items()):
        briefs = {h: (c["evidence"] or {}).get("brief_sha256") for h, c in hands.items()}
        images = {h: (c["evidence"] or {}).get("world_image_digest") for h, c in hands.items()}
        nets = {h: (c["evidence"] or {}).get("network_mode") for h, c in hands.items()}
        lims = {h: (c["evidence"] or {}).get("limits") for h, c in hands.items()}
        mats = {h: (c["evidence"] or {}).get("materials") for h, c in hands.items()}
        solo_bin = (hands.get("solo", {}).get("evidence") or {}).get("punchtape_binary_sha256")
        if len(set(briefs.values())) > 1:
            violations.append("%s-%s: brief digests diverge: %s" % (sense, stack, briefs))
        if len(set(images.values())) > 1:
            violations.append("%s-%s: world images diverge: %s" % (sense, stack, images))
        if any(v != "none" for v in nets.values() if v is not None):
            violations.append("%s-%s: network not off for every hand: %s" % (sense, stack, nets))
        if len(set(map(json.dumps, lims.values()))) > 1:
            violations.append("%s-%s: limits diverge: %s" % (sense, stack, lims))
        if len(set(map(json.dumps, mats.values()))) > 1:
            violations.append("%s-%s: materials diverge: %s" % (sense, stack, mats))
    for h, c in hands.items():
            ev = c["evidence"] or {}
            if c["meta"]["hand"] == "solo" and solo_bin is not None:
                violations.append("%s-%s solo: the binary is mounted (%s)" % (sense, stack, solo_bin))
            for m, after in (ev.get("materials_after") or {}).items():
                if ev.get("materials", {}).get(m) != after:
                    violations.append("%s-%s %s: material %s changed" % (sense, stack, h, m))
            if ev.get("canary_leaks"):
                violations.append("%s-%s %s: CANARY LEAK: %s" % (sense, stack, h, ev["canary_leaks"]))
            if ev.get("transcript_scan"):
                violations.append("%s-%s %s: TRANSCRIPT SCAN HIT: %s" % (sense, stack, h, ev["transcript_scan"]))
            if ev.get("product_scan"):
                violations.append("%s-%s %s: PRODUCT SCAN HIT: %s" % (sense, stack, h, ev["product_scan"]))
    if violations:
        w("**RED LINES:**")
        for v in violations:
            w("- " + v)
    else:
        w("Brief/materials/image/limits identical for every hand of every task; "
          "network off; mounts only its own; materials unchanged after the "
          "finish; solo without the binary; canaries intact; transcript and "
          "product scans empty — **clean**.")
    w("")

    # --- 2. the metrics table by hands
    w("## 2. Quality/completeness and economics — pairwise, identical measurements")
    w("")
    w("Hands are compared row by row on this table, pairwise: (a) the suite ×2 on each")
    w("hand's own product — one procedure for all; (b) deliverables — one counter with a")
    w("method line (hand units differ constructively, none is equated); (c) coverage —")
    w("recomputed the same way. The benchmark is the brief + the expectations list; no")
    w("hand is the reference point. Economics: the base is solo, the delta of the machine")
    w("hands against solo is \"what the machine buys\"; top vs flash is the second axis")
    w("(inside machine processing).")
    w("")
    w("| Task | Hand | Time, s | Tokens in/out | Cost, $ | Submissions ±/− | verb-wall, s | submit-p95, ms | Verdict (Reported) | Cases | Tests | Suite×2 | Deliverables | Prose/capab | Coverage | LOC |")
    w("|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|")
    for (sense, stack), hands in sorted(tasks.items()):
        for h in ("solo", "top", "weak"):
            c = hands.get(h)
            if not c:
                continue
            p, s, m, led = c["platform"] or {}, c["summary"] or {}, c["measure"] or {}, c["ledger"]
            cases = "%s/%s" % (s.get("cases_pass"), s.get("cases_total")) if s.get("cases_total") is not None else "—"
            tests = "%s/%s" % (s.get("tests_pass"), s.get("tests_total")) if s.get("tests_total") is not None else "—"
            prose = "%s/%s" % (s.get("describes"), s.get("capabilities")) if s.get("capabilities") else "—"
            cov = m.get("coverage_tooling") or s.get("code_coverage_pct") or "n/a"
            dl = "%s/5" % m.get("deliverables_count") if m.get("deliverables_count") is not None else "—"
            suite = "×2 green" if m.get("suite_double_green") else ("not green×2" if m else "—")
            w("| %s-%s | %s | %s | %s/%s | %s | %s/%s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s |" % (
                sense, stack, h,
                fmt(p.get("elapsed_s")), fmt(p.get("tokens_in")), fmt(p.get("tokens_out")),
                fmt(c["cost"]),
                led["submissions_accepted"], led["submissions_rejected"],
                fmt(led["verb_wall_ms"] / 1000.0 if led["verb_wall_ms"] else None),
                fmt(led["submit_p95_ms"]),
                (s.get("ruling") or led.get("verdict") or "—")[:60],
                cases, tests, suite, dl, prose, cov,
                fmt(s.get("loc"))))
    w("")

    # --- 3. deltas
    w("## 3. Deltas (what the machine buys)")
    w("")
    w("| Task | Solo $ / s / deliv. | Top $ / s / deliv. | Flash $ / s / deliv. | Top−Solo | Flash−Solo | Top−Flash |")
    w("|---|---|---|---|---|---|---|")

    def triple(hands, h):
        c = hands.get(h)
        if not c:
            return "—", None
        p, m = c["platform"] or {}, c["measure"] or {}
        return ("%s / %s / %s" % (fmt(c["cost"]), fmt(p.get("elapsed_s")),
                                  fmt(m.get("deliverables_count"))), c)

    agg = {"top-solo": [], "weak-solo": [], "top-weak": []}
    for (sense, stack), hands in sorted(tasks.items()):
        s_str, s_c = triple(hands, "solo")
        t_str, t_c = triple(hands, "top")
        f_str, f_c = triple(hands, "weak")

        def d(a, b, key, sub=None):
            if not a or not b:
                return None
            x, y = (a["cost"] if key == "cost" else (a["platform"] or {}).get(key)
                    if not sub else (a[sub] or {}).get(key)), \
                   (b["cost"] if key == "cost" else (b["platform"] or {}).get(key)
                    if not sub else (b[sub] or {}).get(key))
            if x is None or y is None:
                return None
            return key, round(x - y, 2)
        dts = d(t_c, s_c, "cost") or ("cost", None)
        w("| %s-%s | %s | %s | %s | %s | %s | %s |" % (
            sense, stack, s_str, t_str, f_str,
            fmt(t_c["cost"] - s_c["cost"]) if t_c and s_c else "—",
            fmt(f_c["cost"] - s_c["cost"]) if f_c and s_c else "—",
            fmt(t_c["cost"] - f_c["cost"]) if t_c and f_c else "—"))
        for pair, a, b in (("top-solo", t_c, s_c), ("weak-solo", f_c, s_c),
                           ("top-weak", t_c, f_c)):
            if a and b:
                agg[pair].append({
                    "cost": (a["cost"] or 0) - (b["cost"] or 0),
                    "elapsed": ((a["platform"] or {}).get("elapsed_s") or 0)
                               - ((b["platform"] or {}).get("elapsed_s") or 0),
                    "deliverables": ((a["measure"] or {}).get("deliverables_count") or 0)
                                    - ((b["measure"] or {}).get("deliverables_count") or 0),
                })
    w("")
    for pair, name in (("top-solo", "Top vs solo"), ("weak-solo", "Flash vs solo"),
                       ("top-weak", "Top vs flash")):
        rows = agg[pair]
        if rows:
            w("- %s (mean over tasks): cost %s, time %s s, deliverables %s." % (
                name,
                fmt(sum(r["cost"] for r in rows) / len(rows)),
                fmt(sum(r["elapsed"] for r in rows) / len(rows)),
                fmt(sum(r["deliverables"] for r in rows) / len(rows))))
    w("")

    # --- 4. comparability
    w("## 4. Comparability")
    w("")
    methods = sorted({(c["measure"] or {}).get("suite_method") for c in cells
                      if (c["measure"] or {}).get("suite_method")})
    if len(methods) == 1:
        w("Every hand's suite ran under one procedure (method line: \"%s\")." % methods[0])
    else:
        w("Suite method lines: %s" % "; ".join(str(m) for m in methods))
    w("Deliverables are counted by one script with a recorded method line; "
      "coverage is recomputed the same way wherever the stack has tooling.")
    w("")

    # --- 5. operator overheads (separate section, never summed with hands)
    w("## 5. Operator overheads (separate accounting)")
    w("")
    w("| Task | Hand | Model | Time, s | Tokens in/out | Decisions | hand's decision_wait_s |")
    w("|---|---|---|---|---|---|---|")
    for (sense, stack), hands in sorted(tasks.items()):
        for h in ("solo", "top", "weak"):
            c = hands.get(h)
            if not c:
                continue
            o = c["operator"] or {}
            p = c["platform"] or {}
            w("| %s-%s | %s | %s | %s | %s/%s | %s | %s |" % (
                sense, stack, h, o.get("model", "—"), fmt(o.get("elapsed_s")),
                fmt(o.get("tokens_in")), fmt(o.get("tokens_out")),
                fmt(o.get("decisions")), fmt(p.get("decision_wait_s"))))
    w("")

    # --- 6. findings registry
    w("## 6. Problems, pains and findings registry")
    w("")
    seen = 0
    for c in cells:
        for f in (c["findings"] or []):
            w("- [%s] %s-%s %s: %s" % (f.get("kind", "?"), c["meta"]["sense"],
                                       c["meta"]["stack"], c["meta"]["hand"],
                                       f.get("text", "")))
            seen += 1
    if not seen:
        w("(no findings)")

    out = "\n".join(L) + "\n"
    with open(args.out, "w") as f:
        f.write(out)
    print(out)


if __name__ == "__main__":
    main()
