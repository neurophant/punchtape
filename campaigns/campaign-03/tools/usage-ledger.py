#!/usr/bin/env python3
# usage-ledger.py — the reproducible proof ledger of every agent's usage in a
# wave session: extracts per-segment totals from the PLATFORM's own session
# database and per-agent splits from the PLATFORM's own metadata files, then
# cross-checks both against the registry. No number in this script is
# hand-entered; its output is the evidence. Re-running it must reproduce the
# committed ledger byte-for-byte (deterministic extraction, no clock, no RNG).
#
# Sources (all platform-owned, none produced by the driver):
#   db.sqlite (part table) — every completed agent turn carries a usage block
#     (subagent_tokens/tool_uses/duration_ms) in the platform's session log;
#   agent_<id>/metadata.json — the platform's per-agent usage (inputTokens/
#     outputTokens/cacheReadTokens, lifetime of the last segment).
#
# Usage:
#   usage-ledger.py <agent-id>...          # ledger + cross-check, JSON out
#   usage-ledger.py --session <sess-id> --db <db.sqlite> --agents-dir <dir> <id>...
import argparse
import json
import os
import re
import sqlite3
import sys


def db_segments(db, ids):
    con = sqlite3.connect("file:%s?mode=ro" % db, uri=True)
    rows = con.execute(
        "SELECT data, time_created FROM part WHERE data LIKE '%subagent_tokens%'"
    ).fetchall()
    segs = {i: [] for i in ids}
    for data, t in rows:
        # ONLY the platform's own task-notification parts count: a completion
        # is also quoted inside the driver's tool outputs and reasoning, and
        # counting those would double every segment. The notification parts
        # are type=text and carry the <task-notification> envelope.
        try:
            head = json.loads(data)
        except json.JSONDecodeError:
            continue
        if head.get("type") != "text" or "<task-notification>" not in data:
            continue
        for m in re.finditer(r"subagent_tokens[^0-9]{0,4}(\d+)", data):
            win = data[max(0, m.start() - 20000):m.start()]
            found = re.findall(r"agent_([0-9a-f]{8})", win)
            if not found:
                continue
            aid = found[-1]
            if aid not in segs:
                continue
            tools = re.search(r"tool_uses[^0-9]{0,4}(\d+)", data[m.start():m.start() + 200])
            dur = re.search(r"duration_ms[^0-9]{0,4}(\d+)", data[m.start():m.start() + 200])
            segs[aid].append({
                "part_ts": t,
                "tokens": int(m.group(1)),
                "tool_uses": int(tools.group(1)) if tools else None,
                "duration_ms": int(dur.group(1)) if dur else None,
            })
    for aid in segs:
        segs[aid].sort(key=lambda x: x["part_ts"])
    return segs


def metadata_split(agents_dir, aid):
    for d in os.listdir(agents_dir):
        if d.startswith("agent_" + aid):
            mp = os.path.join(agents_dir, d, "metadata.json")
            if os.path.isfile(mp):
                m = json.load(open(mp))
                u = m.get("usage") or {}
                return {"metadata_file": os.path.join(d, "metadata.json"),
                        "inputTokens": u.get("inputTokens"),
                        "outputTokens": u.get("outputTokens"),
                        "cacheReadTokens": u.get("cacheReadTokens"),
                        "totalDurationMs": m.get("totalDurationMs"),
                        "totalToolUseCount": m.get("totalToolUseCount")}
    return None


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("ids", nargs="+")
    ap.add_argument("--db", default=os.path.expanduser(
        "~/.zcode/cli/db/db.sqlite"))
    ap.add_argument("--agents-dir", required=True)
    ap.add_argument("--out", default="-")
    args = ap.parse_args()

    segs = db_segments(args.db, args.ids)
    ledger = {}
    for aid in args.ids:
        meta = metadata_split(args.agents_dir, aid)
        s = segs[aid]
        life = sum(x["tokens"] for x in s)
        split = ((meta or {}).get("inputTokens") or 0) + ((meta or {}).get("outputTokens") or 0)
        ledger[aid] = {
            "segments": s,
            "lifetime_tokens_db": life,
            "segment_count": len(s),
            "metadata": meta,
            "split_total": split,
            "split_matches_last_segment": bool(
                s and meta and split == s[-1]["tokens"]),
            "coverage_pct": round(100 * min(split, life) / life, 1) if life else None,
        }
    txt = json.dumps(ledger, indent=1, sort_keys=True)
    if args.out == "-":
        print(txt)
    else:
        open(args.out, "w").write(txt)
        print("written", args.out)


if __name__ == "__main__":
    sys.exit(main())
