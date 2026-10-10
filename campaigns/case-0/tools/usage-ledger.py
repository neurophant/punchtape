#!/usr/bin/env python3
# """usage-ledger.py — the reproducible usage ledger of the wave on
# the PLATFORM'S OWN db: per agent -
# the model_usage table of ~/.zcode/cli/db/db.sqlite grouped by the
# agent's session (Σ input, Σ output, Σ cache tokens, Σ duration_ms
# = ACTIVE seconds, the request count), validated against the
# transcript itself (the step-finish parts of the same session: a
# divergence over 1% is a red accounting line - the ledger carries
# the flag, the driver writes the wave's red line). Idle and waiting
# are counted NOWHERE: the wave's time axis is active seconds only.
#
# The agent<->session mapping (the case-0 protocol roles: hand |
# operator | judge | measurer): each HAND works inside its cell, and
# the platform records every session's working directory - a cell's
# hand sessions are the sessions whose directory lies under the cell
# path (the manifests carry cell_path); explicit --session <id8>=<sess>
# overrides beat the mapping. The micro-agents (operators, judges,
# measurers) are spawned by the driver session from its sterile cwd -
# the driver PINS their session ids at registration.
#
# Usage:
#   usage-ledger.py --root <run-root> [--db <db.sqlite>]
#              [--driver-session <sess-id>] [--session <id8>=<sess> ...]
#              ALL registry agents + per-cell time parts ->
#              <root>/usage/ledger.json (+ per-agent slices)
#   usage-ledger.py <session-id>... [--db <db.sqlite>] [--out F|-]
#              ad-hoc: the bare sessions table + validation
# No number is hand-entered; re-running reproduces the committed
# ledger byte-for-byte (deterministic extraction, no clock, no RNG).
# """
import argparse
import glob
import json
import os
import re
import sqlite3
import sys

# the accepted divergence of the request count against the
# transcript's own step count (>1% is a red accounting
# line)
DIVERGENCE_PCT = 1.0


def id8_of(uuid):
    m = re.search(r"([0-9a-fA-F]{8})", uuid or "")
    return m.group(1).lower() if m else None


def db_facts(db, sessions):
    """The per-session facts from the platform's own tables:
    model_usage (tokens, duration, request count - summed here, the
    SQL stays flat) and the part table (the transcript's step-finish
    count - the validation axis)."""
    if not sessions:
        return {}
    con = sqlite3.connect("file:%s?mode=ro" % db, uri=True)
    qmarks = ",".join("?" for _ in sessions)
    usage = {s: {"input_tokens": 0, "output_tokens": 0, "cache_tokens": 0,
                 "active_s": 0.0, "requests": 0, "tool_calls": 0,
                 "model": None} for s in sessions}
    for sid, inp, outp, cache, dur, tools, model in con.execute(
            "SELECT session_id, input_tokens, output_tokens,"
            " COALESCE(cache_read_input_tokens,0)"
            " + COALESCE(cache_creation_input_tokens,0),"
            " COALESCE(duration_ms,0), COALESCE(tool_call_count,0), model_id"
            f" FROM model_usage WHERE status='completed' AND session_id IN ({qmarks})",
            tuple(sessions)):
        u = usage.setdefault(sid, {"input_tokens": 0, "output_tokens": 0,
                                   "cache_tokens": 0, "active_s": 0.0,
                                   "requests": 0, "tool_calls": 0,
                                   "model": None})
        u["input_tokens"] += inp or 0
        u["output_tokens"] += outp or 0
        u["cache_tokens"] += cache or 0
        u["active_s"] += (dur or 0) / 1000.0
        u["requests"] += 1
        u["tool_calls"] += tools or 0
        u["model"] = model or u["model"]
    steps = {s: 0 for s in usage}
    for (sid,) in con.execute(
            "SELECT session_id"
            f" FROM part WHERE session_id IN ({qmarks}) AND data LIKE '%\"type\":\"step-finish\"%'",
            tuple(sessions)):
        steps[sid] = steps.get(sid, 0) + 1
    con.close()
    facts = {}
    for sid, u in usage.items():
        u["active_s"] = round(u["active_s"], 1)
        u["steps"] = steps.get(sid, 0)
        base = max(u["steps"], 1)
        div = 100.0 * abs(u["requests"] - u["steps"]) / base
        u["divergence_pct"] = round(div, 2)
        u["validation"] = "ok" if div <= DIVERGENCE_PCT else "red"
        u["total_tokens"] = u["input_tokens"] + u["output_tokens"] + u["cache_tokens"]
        facts[sid] = u
    return facts


def sessions_by_directory(db, prefixes):
    """session id -> its working directory, from the platform's own
    session table; a cell owns every session whose directory lies
    under its path."""
    con = sqlite3.connect("file:%s?mode=ro" % db, uri=True)
    out = {p: [] for p in prefixes}
    for sid, d in con.execute("SELECT id, directory FROM session"):
        if not d:
            continue
        for p in prefixes:
            if d == p or d.startswith(p.rstrip("/") + "/"):
                out[p].append(sid)
    con.close()
    return out


def merge_entry(entry, facts, sid):
    """Fold one session's facts into an agent's entry."""
    for k in ("input_tokens", "output_tokens", "cache_tokens", "requests",
              "steps", "tool_calls"):
        entry[k] = entry.get(k, 0) + facts[k]
    entry["active_s"] = round(entry.get("active_s", 0.0) + facts["active_s"], 1)
    entry["model"] = entry.get("model") or facts.get("model")
    if facts["validation"] != "ok":
        entry["validation"] = "red"
    entry["sessions"].append(sid)
    entry["session_facts"] = {sid: facts}


def agent_entry(sids_facts, cells=None, roles=None):
    entry = {"sessions": [], "session_facts": {},
             "validation": "ok", "input_tokens": 0, "output_tokens": 0,
             "cache_tokens": 0, "requests": 0, "steps": 0, "active_s": 0.0,
             "tool_calls": 0, "model": None,
             "cells": sorted(cells) if cells else None,
             "roles": sorted(roles) if roles else None}
    for sid, facts in sids_facts:
        merge_entry(entry, facts, sid)
    entry["total_tokens"] = (entry["input_tokens"] + entry["output_tokens"]
                             + entry["cache_tokens"])
    base = max(entry["steps"], 1)
    entry["divergence_pct"] = round(
        100.0 * abs(entry["requests"] - entry["steps"]) / base, 2)
    return entry


def registry_of(root):
    """driver-agents.json: cell -> [{uuid, role}] (the register's own
    shape); agents keyed by id8 accumulate cells and roles."""
    reg_path = os.path.join(root, "driver-agents.json")
    reg = json.load(open(reg_path, encoding="utf-8")) \
        if os.path.isfile(reg_path) else {}
    agents = {}
    for cell, entries in sorted(reg.items()):
        for e in entries or []:
            aid = id8_of(e.get("uuid") if isinstance(e, dict) else None)
            if not aid:
                continue
            a = agents.setdefault(aid, {"cells": set(), "roles": set()})
            a["cells"].add(cell)
            a["roles"].add(e.get("role") or "hand")
    return agents


def cell_paths_of(root):
    """manifests/*.json: cell -> cell_path (the session-directory
    mapping anchor)."""
    out = {}
    for m in sorted(glob.glob(os.path.join(root, "manifests", "*.json"))):
        try:
            d = json.load(open(m, encoding="utf-8"))
        except (json.JSONDecodeError, OSError):
            continue
        if isinstance(d, dict) and d.get("cell_path"):
            out[d.get("cell")] = d["cell_path"]
    return out


def cell_time_parts(root, ledger):
    """Per cell: work_s = the hand agents' active seconds; operator_s
    = the operator agents' (customer/owner/judge/measurer) active
    seconds; wait_s = the decision waits of the finished receipts.
    Idle and pauses are counted nowhere."""
    parts = {}
    for fin_path in sorted(glob.glob(os.path.join(root, "finished", "*.json"))):
        try:
            fin = json.load(open(fin_path, encoding="utf-8"))
        except (json.JSONDecodeError, OSError):
            continue
        cell = fin.get("cell")
        if not cell:
            continue
        reg = json.load(open(os.path.join(root, "driver-agents.json"),
                             encoding="utf-8")) \
            if os.path.isfile(os.path.join(root, "driver-agents.json")) else {}
        hand_ids = [id8_of(e.get("uuid")) for e in (reg.get(cell) or [])
                    if isinstance(e, dict) and e.get("role") == "hand"]
        oper_ids = [id8_of(e.get("uuid")) for e in (reg.get(cell) or [])
                    if isinstance(e, dict)
                    and e.get("role") in ("operator", "judge", "measurer")]
        work = round(sum((ledger.get(i) or {}).get("active_s") or 0
                         for i in hand_ids), 1) or None
        oper = round(sum((ledger.get(i) or {}).get("active_s") or 0
                         for i in oper_ids), 1) or None
        waits = fin.get("decision_wait_s") or []
        parts[cell] = {
            "work_s": work,
            "operator_s": oper,
            "wait_s": round(sum(waits), 1) if waits else None,
        }
    return parts


def write_ledger(root, doc):
    out = os.path.join(root, "usage", "ledger.json")
    os.makedirs(os.path.dirname(out), exist_ok=True)
    with open(out, "w", encoding="utf-8") as f:
        f.write(json.dumps(doc, indent=1, sort_keys=True) + "\n")
    for aid, entry in doc["agents"].items():
        with open(os.path.join(root, "usage", f"{aid}.json"), "w",
                  encoding="utf-8") as f:
            f.write(json.dumps(entry, indent=1, sort_keys=True) + "\n")
    return out


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("sessions", nargs="*",
                    help="bare mode: platform session ids")
    ap.add_argument("--db", default=os.path.expanduser(
        "~/.zcode/cli/db/db.sqlite"))
    ap.add_argument("--out", default="-")
    ap.add_argument("--root", help="the run root: ALL registry agents")
    ap.add_argument("--session", action="append", default=[],
                    metavar="ID8=SESS",
                    help="explicit agent->session pin (beats the mapping)")
    ap.add_argument("--driver-session",
                    help="the driver session's own id (role driver)")
    args = ap.parse_args()

    if args.root:
        root = args.root
        agents = registry_of(root)
        paths = cell_paths_of(root)
        prefixes = sorted(set(p for p in paths.values()))
        by_dir = sessions_by_directory(args.db, prefixes) if prefixes else {}
        pinned = {}
        for spec in args.session:
            if "=" in spec:
                aid, sid = spec.split("=", 1)
                pinned.setdefault(aid, []).append(sid)
        ledger = {}
        for aid in sorted(agents):
            # a pin REPLACES the directory mapping for that agent: the
            # driver pins when it knows the session better than the
            # directory rule (operator sessions, relocated sessions)
            if aid in pinned:
                sids = list(pinned[aid])
            else:
                sids = []
                # the directory rule hands a cell's sessions to its
                # HAND (the hand works in the cell; an operator session
                # of the same cell would be indistinguishable by
                # directory - the driver pins operator sessions)
                hand_cells = sorted(c for c in agents[aid]["cells"]
                                    if "hand" in agents[aid]["roles"])
                for cell in hand_cells:
                    sids.extend(by_dir.get(paths.get(cell), []))
            facts = db_facts(args.db, sorted(set(sids)))
            ledger[aid] = agent_entry(
                [(s, f) for s, f in sorted(facts.items())],
                cells=agents[aid]["cells"], roles=agents[aid]["roles"])
        if args.driver_session:
            # E2C (the owner's clean-metrics word): the driver camp may
            # span SEVERAL sessions (a driver that handed off by context
            # limit, its successor, orchestration sub-drivers) — pass them
            # comma-separated; every one of them is driver-side overhead
            ids = [x.strip() for x in args.driver_session.split(",") if x.strip()]
            aid = "driver"
            facts = db_facts(args.db, ids)
            entry = agent_entry([(s, f) for s, f in sorted(facts.items())],
                                cells={"(driver)"}, roles={"driver"})
            entry["driver_note"] = (
                "the driver is the parent session: its tokens are the wave's "
                "overhead, not any camp's (gate 1 excludes it)")
            ledger[aid] = entry
        parts = cell_time_parts(root, ledger)
        doc = {"agents": ledger, "cell_time_parts": parts}
        out = write_ledger(root, doc)
        print(f"written {out} ({len(ledger)} agents, {len(parts)} cells with time parts)")
        return 0

    if not args.sessions:
        ap.error("either --root or explicit session ids")
    facts = db_facts(args.db, args.sessions)
    txt = json.dumps(facts, indent=1, sort_keys=True)
    if args.out == "-":
        print(txt)
    else:
        open(args.out, "w").write(txt)
        print("written", args.out)
    return 0


if __name__ == "__main__":
    sys.exit(main())
