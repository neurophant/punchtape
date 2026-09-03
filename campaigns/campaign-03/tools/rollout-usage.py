#!/usr/bin/env python3
# rollout-usage.py — the full-coverage token accounting for EVERY role of a
# run (hands, operators, judges, blind measurers — the owner's rule: collect
# everything there is on everything, tokens input AND output separately).
#
# Sources, in order of trust:
#   1. rollout JSONL records (per-call response.usage: inputTokens,
#      outputTokens, cacheReadTokens) — the ONLY source of the in/out split;
#      the watcher's rotation snapshots under a cell dir are the primary
#      copies (the platform trims its store faster than any finish-time
#      copy);
#   2. the platform rollout store itself (~/.zcode/cli/rollout) matched by
#      agent id — whatever still survives there;
#   3. nothing else: a pruned record is an honest gap, printed as such —
#      notification totals (no split) live in platform.json/operator.json
#      and stay the marked fallback, never silently mixed in here.
#
# Usage:
#   rollout-usage.py <glob>...            # one JSON line per file set:
#                                         # {"source": path, "calls": N,
#                                         #  "tokens_in": X, "tokens_out": Y,
#                                         #  "tokens_cache_read": Z,
#                                         #  "total_inout": T}
#   rollout-usage.py --by-id <id>...      # sum over the live store files
#                                         # matching *<id>*.jsonl
import argparse
import glob
import json
import os
import sys

STORE = os.path.expanduser("~/.zcode/cli/rollout")


def scan(paths):
    calls = tin = tout = tcache = 0
    for p in paths:
        try:
            with open(p, errors="replace") as fh:
                for line in fh:
                    line = line.strip()
                    if not line:
                        continue
                    try:
                        d = json.loads(line)
                    except json.JSONDecodeError:
                        continue
                    u = (d.get("response") or {}).get("usage") or {}
                    if not u:
                        continue
                    calls += 1
                    tin += u.get("inputTokens") or 0
                    tout += u.get("outputTokens") or 0
                    tcache += u.get("cacheReadTokens") or 0
        except OSError:
            continue
    return {"files": len(paths), "calls": calls, "tokens_in": tin,
            "tokens_out": tout, "tokens_cache_read": tcache,
            "total_inout": tin + tout}


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("globs", nargs="*")
    ap.add_argument("--by-id", nargs="*", default=[])
    args = ap.parse_args()
    for g in args.globs:
        paths = sorted(glob.glob(g))
        if not paths:
            print(json.dumps({"source": g, "missing": True}))
            continue
        r = scan(paths)
        r["source"] = g
        print(json.dumps(r))
    for i in args.by_id:
        paths = sorted(glob.glob(os.path.join(
            STORE, "model-io-*%s*.jsonl" % i)))
        r = scan(paths)
        r["source"] = "store:*%s*" % i
        r["store_alive"] = bool(paths)
        print(json.dumps(r))


if __name__ == "__main__":
    sys.exit(main())
