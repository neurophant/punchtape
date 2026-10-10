#!/usr/bin/env python3
"""parse-ledger.py — the one parser of instance ledgers.

Extracts the machine-side metrics of a run from one or more
.punchtape/ledger.yamll files:

- verbs          — verb calls by name, and the total
- submissions    — accepted (+) and rejected (-) submit calls
- verb-wall      — the sum of wall-ms across all verb calls, seconds
- submit-p95     — p95 of (wall-ms - runs-ms) over submissions, ms
- reply-bytes    — the sum of reply-bytes, KB

Usage:
    parse-ledger.py LEDGER [LEDGER ...]        # per-file report
    parse-ledger.py --sum LEDGER [LEDGER ...]  # one aggregate row

A ledger record lives between --- separators; a verb record carries
details.name, details.outcome (ok / accepted / rejected / ...),
details.runs-ms and details.reply-bytes, plus the header wall-ms. One parser for every run,
baseline and wave — the numbers are only comparable when the
extraction is identical.
"""

import re
import sys
import math

REC_SEP = re.compile(r"^---\s*$", re.M)
NAME = re.compile(r"name:\s*(\S+)")
WALL = re.compile(r'wall-ms:\s*"?([0-9]+)"?')
RUNS = re.compile(r'runs-ms:\s*"?([0-9]+)"?')
REPLY = re.compile(r'reply-bytes:\s*"?([0-9]+)"?')
ACCEPTED = re.compile(r"outcome:\s*accepted")


def records(text):
    for chunk in REC_SEP.split(text):
        if "name:" not in chunk:
            continue
        yield chunk


def parse(path):
    """One ledger -> a metrics dict."""
    with open(path, encoding="utf-8") as fh:
        text = fh.read()
    verbs = {}
    wall_total = 0
    reply_total = 0
    submits_accepted = 0
    submits_rejected = 0
    net_walls = []
    for rec in records(text):
        m = NAME.search(rec)
        if not m:
            continue
        verb = m.group(1)
        verbs[verb] = verbs.get(verb, 0) + 1
        w = WALL.search(rec)
        if w:
            wall_total += int(w.group(1))
            if verb == "submit":
                r = RUNS.search(rec)
                net = int(w.group(1)) - (int(r.group(1)) if r else 0)
                net_walls.append(max(net, 0))
        rb = REPLY.search(rec)
        if rb:
            reply_total += int(rb.group(1))
        if verb == "submit":
            if ACCEPTED.search(rec):
                submits_accepted += 1
            else:
                submits_rejected += 1
    p95 = percentile(net_walls, 95)
    return {
        "verbs_total": sum(verbs.values()),
        "verbs": verbs,
        "submissions": (submits_accepted, submits_rejected),
        "verb_wall_s": wall_total / 1000.0,
        "submit_p95_ms": p95,
        "reply_kb": reply_total / 1024.0,
    }


def percentile(values, p):
    """Nearest-rank percentile; 0 for an empty sample."""
    if not values:
        return 0
    ordered = sorted(values)
    rank = max(1, math.ceil(p / 100.0 * len(ordered)))
    return ordered[rank - 1]


def main(argv):
    args = argv[1:]
    aggregate = "--sum" in args
    paths = [a for a in args if a != "--sum"]
    if not paths:
        sys.stderr.write(__doc__)
        return 2
    if aggregate:
        merged = {
            "verbs_total": 0,
            "verbs": {},
            "submissions": (0, 0),
            "verb_wall_s": 0.0,
            "net_walls": [],
            "reply_kb": 0.0,
        }
        for path in paths:
            with open(path, encoding="utf-8") as fh:
                text = fh.read()
            for rec in records(text):
                m = NAME.search(rec)
                if not m:
                    continue
                verb = m.group(1)
                merged["verbs"][verb] = merged["verbs"].get(verb, 0) + 1
                merged["verbs_total"] += 1
                w = WALL.search(rec)
                if w:
                    merged["verb_wall_s"] += int(w.group(1)) / 1000.0
                    if verb == "submit":
                        r = RUNS.search(rec)
                        merged["net_walls"].append(
                            max(int(w.group(1)) - (int(r.group(1)) if r else 0), 0)
                        )
                rb = REPLY.search(rec)
                if rb:
                    merged["reply_kb"] += int(rb.group(1)) / 1024.0
                if verb == "submit":
                    acc = bool(ACCEPTED.search(rec))
                    a, rj = merged["submissions"]
                    merged["submissions"] = (a + 1, rj) if acc else (a, rj + 1)
        a, rj = merged["submissions"]
        print(
            f"files={len(paths)} verbs={merged['verbs_total']} "
            f"submissions={a}/{rj} verb-wall={merged['verb_wall_s']:.1f}s "
            f"submit-p95={percentile(merged['net_walls'], 95)}ms "
            f"reply={merged['reply_kb']:.1f}KB"
        )
        return 0
    for path in paths:
        m = parse(path)
        a, rj = m["submissions"]
        verbs = " ".join(f"{k}={v}" for k, v in sorted(m["verbs"].items()))
        print(
            f"{path}: verbs={m['verbs_total']} ({verbs}) "
            f"submissions={a}/{rj} verb-wall={m['verb_wall_s']:.1f}s "
            f"submit-p95={m['submit_p95_ms']}ms "
            f"reply={m['reply_kb']:.1f}KB"
        )
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv))
