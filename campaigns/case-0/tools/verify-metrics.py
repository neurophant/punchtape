#!/usr/bin/env python3
"""verify-metrics.py — the PER-CELL printed verification of a closed
wave (the owner's hard-audit word): every
number of every report table carries its FULL PATH from the source —
the platform db's own model_usage rows by session -> the usage ledger
agent slice -> the cell's evidence file -> the table cell — and every
Σ is re-added, every camp independently recounted. Census: registry
<-> pins <-> db sessions of the wave's window. The tool is called by
close-wave automatically; the output goes to the screen and to
<root>/verification/ (the method heir of the manual run).

Usage: verify-metrics.py <run-root> [--db <db.sqlite>]
                          [--report <wave-report.md>]
Exit 0 = every check AGREEs; 1 = at least one RED.
"""
import argparse
import datetime
import glob
import importlib.util
import json
import os
import sqlite3
import sys

HERE = os.path.dirname(os.path.abspath(__file__))


def load_usage_ledger_mod():
    spec = importlib.util.spec_from_file_location(
        "usage_ledger", os.path.join(HERE, "usage-ledger.py"))
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod


def jload(path):
    try:
        with open(path, encoding="utf-8") as f:
            return json.load(f)
    except Exception:
        return None


def f1(v):
    return "%0.1f" % v


def f2(v):
    return "%0.2f" % v


def f0(v):
    return str(int(v))


def cellsplit(row):
    return [c.strip() for c in row.strip().strip("|").split("|")]


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("run_root")
    ap.add_argument("--db", default=os.path.expanduser(
        "~/.zcode/cli/db/db.sqlite"))
    ap.add_argument("--report", default=None)
    args = ap.parse_args()
    root = os.path.abspath(args.run_root)
    report_path = args.report or os.path.join(root, "wave-report.md")

    ul = load_usage_ledger_mod()
    out_lines = []
    red = [0]
    checks = [0]

    def say(line=""):
        out_lines.append(line)
        print(line)

    def ok(line):
        checks[0] += 1
        say("AGREE " + line)

    def bad(line):
        red[0] += 1
        say("RED   " + line)

    def cmp_num(label, table_val, recompute, fmtfun):
        # the report prints "—" for a None (fmt's own law): a cell
        # without a source agrees only with the dash
        if recompute is None:
            if table_val == "—":
                checks[0] += 1
                say(f"AGREE {label}: — (no source value, the report's dash)")
            else:
                red[0] += 1
                say(f"RED   {label}: table={table_val} recomputed=— (no source)")
            return
        want = fmtfun(recompute)
        if table_val == want:
            ok(f"{label}: {table_val} (path: db rows -> session -> ledger -> evidence -> table)")
        else:
            bad(f"{label}: table={table_val} recomputed={want}")

    def multv(a, b):
        # report.py's own mult: None when unknown or a zero divisor
        # (the cell then prints the dash)
        if a is None or b is None or b == 0:
            return None
        return a / b

    say(f"== verify-metrics {os.path.basename(root)} "
        f"({datetime.datetime.now(datetime.timezone.utc).isoformat(timespec='seconds')})")

    # --- the census basis: registry, pins, manifests
    reg = jload(os.path.join(root, "driver-agents.json")) or {}
    pins = {}
    pinf = os.path.join(root, "session-pins.txt")
    if os.path.isfile(pinf):
        for l in open(pinf, encoding="utf-8"):
            if "=" in l:
                k, v = l.split("=", 1)
                pins.setdefault(k.strip(), []).append(v.strip())
    manifests = {}
    for m in sorted(glob.glob(os.path.join(root, "manifests", "*.json"))):
        d = jload(m)
        if d:
            manifests[d["cell"]] = d
    agents_by_id8 = {}
    for cell, entries in sorted(reg.items()):
        for e in entries or []:
            id8 = ul.id8_of(e.get("uuid"))
            if id8:
                a = agents_by_id8.setdefault(id8, {"cells": set(), "roles": set()})
                a["cells"].add(cell)
                a["roles"].add(e.get("role") or "hand")

    # --- 1. census: every subagent session created inside the wave's
    # --- window must be a pinned census member (the census hole
    # --- cannot happen silently again)
    say("-- census: subagent sessions of the wave's window vs the pins")
    try:
        start_ms = int(os.stat(os.path.join(root, "driver-uuid.txt")).st_mtime * 1000)
        con = sqlite3.connect("file:%s?mode=ro" % args.db, uri=True)
        window_rows = con.execute(
            "SELECT id, time_created FROM session"
            " WHERE id LIKE 'sess_subagent_agent_%'").fetchall()
        con.close()
        pinned_sess = {s for v in pins.values() for s in v}
        in_window = [sid for sid, tc in window_rows if (tc or 0) >= start_ms]
        unpinned = sorted(sid for sid in in_window if sid not in pinned_sess)
        if unpinned:
            for sid in unpinned:
                bad(f"census: unpinned session in the wave window: {sid}")
        else:
            ok(f"census: every subagent session of the window is pinned "
               f"({len(in_window)} session(s), {len(pinned_sess)} pin(s))")
    except sqlite3.Error as e:
        bad(f"census: db read failed: {e}")

    # --- 2. agents: the db's own rows == the usage ledger slice
    say("-- agents: db model_usage (by session) == usage/ledger.json")
    ledger = (jload(os.path.join(root, "usage", "ledger.json")) or {}).get("agents") or {}
    cellpaths = {c: d["cell_path"] for c, d in manifests.items()}
    by_dir = ul.sessions_by_directory(args.db, sorted(set(cellpaths.values()))) \
        if cellpaths else {}
    for id8 in sorted(agents_by_id8):
        a = agents_by_id8[id8]
        if id8 in pins:
            sids = sorted(set(pins[id8]))
            how = "pin"
        else:
            sids = []
            for cell in sorted(a["cells"]):
                if "hand" in a["roles"]:
                    sids.extend(by_dir.get(cellpaths.get(cell), []))
            sids = sorted(set(sids))
            how = "directory"
        facts = ul.db_facts(args.db, sids) if sids else {}
        ent = ledger.get(id8)
        if not ent:
            bad(f"agent {id8} ({'/'.join(sorted(a['roles']))}): no ledger slice")
            continue
        d_in = sum(f["input_tokens"] for f in facts.values())
        d_out = sum(f["output_tokens"] for f in facts.values())
        d_cache = sum(f["cache_tokens"] for f in facts.values())
        d_act = round(sum(f["active_s"] for f in facts.values()), 1)
        if (d_in, d_out, d_cache, d_act) == (
                ent.get("input_tokens"), ent.get("output_tokens"),
                ent.get("cache_tokens"), ent.get("active_s")):
            ok(f"agent {id8} ({'/'.join(sorted(a['roles']))}, {how} over "
               f"{len(sids)} session(s)): db in/out/cache/active "
               f"{d_in}/{d_out}/{d_cache}/{d_act} == ledger slice")
        else:
            bad(f"agent {id8}: db {d_in}/{d_out}/{d_cache}/{d_act} != ledger "
                f"{ent.get('input_tokens')}/{ent.get('output_tokens')}/"
                f"{ent.get('cache_tokens')}/{ent.get('active_s')}")

    # --- 3. cells: the evidence files == the ledger slices
    say("-- cells: evidence/platform.json + operator.json == ledger slices")
    ev_of = {}
    for cell in sorted(manifests):
        evd = os.path.join(root, "tasks", cell, "evidence")
        ev_of[cell] = {
            "platform": jload(os.path.join(evd, "platform.json")),
            "operator": jload(os.path.join(evd, "operator.json")),
        }
        ids = {"hand": [], "operator": []}
        for e in reg.get(cell) or []:
            id8 = ul.id8_of(e.get("uuid"))
            if not id8:
                continue
            role = e.get("role") or "hand"
            if role == "hand":
                ids["hand"].append(id8)
            elif role == "operator":
                ids["operator"].append(id8)

        def slice_sum(idlist, keys):
            tot = {k: 0 for k in keys}
            found = False
            for i in idlist:
                e = ledger.get(i)
                if not e:
                    continue
                found = True
                for k in keys:
                    tot[k] += e.get(k) or 0
            return tot if found else None

        hsum = slice_sum(ids["hand"], ["input_tokens", "output_tokens",
                                       "cache_tokens", "active_s"])
        p = ev_of[cell]["platform"] or {}
        if hsum:
            want = {
                "elapsed_s": round(hsum["active_s"], 1),
                "tokens_in": hsum["input_tokens"],
                "tokens_out": hsum["output_tokens"],
                "tokens_cache_read": hsum["cache_tokens"],
                "tokens_inout": hsum["input_tokens"] + hsum["output_tokens"],
                "tokens_total": hsum["input_tokens"] + hsum["output_tokens"]
                                + hsum["cache_tokens"],
            }
            got = {k: p.get(k) for k in want}
            if want == got:
                ok(f"{cell} hand evidence == Σ ledger of {ids['hand']}: "
                   f"time {want['elapsed_s']}s, in+out {want['tokens_inout']}, "
                   f"total {want['tokens_total']}")
            else:
                bad(f"{cell} hand evidence {got} != ledger {want}")
        else:
            bad(f"{cell}: no hand ledger slice behind the evidence")
        o = ev_of[cell]["operator"] or {}
        osum = slice_sum(ids["operator"], ["input_tokens", "output_tokens",
                                           "cache_tokens", "active_s"])
        if ids["operator"]:
            if osum is None and o.get("tokens_in") is None:
                bad(f"{cell}: operator agents registered but no slice/evidence")
            elif osum is not None:
                want_o = (round(osum["active_s"], 1), osum["input_tokens"],
                          osum["output_tokens"], osum["cache_tokens"])
                got_o = (o.get("elapsed_s") or 0, o.get("tokens_in") or 0,
                         o.get("tokens_out") or 0, o.get("tokens_cache_read") or 0)
                if want_o == got_o:
                    ok(f"{cell} operator evidence == Σ ledger of {ids['operator']}")
                else:
                    bad(f"{cell} operator evidence {got_o} != ledger {want_o}")

    # --- 4. the report tables, cell by cell
    say("-- tables: wave-report.md cells vs the evidence (full paths)")
    if not os.path.isfile(report_path):
        bad(f"no report to verify at {report_path}")
    else:
        text = open(report_path, encoding="utf-8").read()
        rows = [cellsplit(l) for l in text.splitlines() if l.startswith("|")]

        def find_row(pred):
            return [r for r in rows if pred(r)]

        # §2 hands: | cell | hand | time | in+out | total | cache | ... (14 cols)
        for cell in sorted(manifests):
            hand = manifests[cell]["hand"]
            rr = [r for r in find_row(lambda r: len(r) > 6 and r[0] == cell
                                      and r[1] == hand and not r[0].startswith("**"))
                  if len(r) >= 13]
            p = ev_of[cell]["platform"] or {}
            if not rr:
                bad(f"§2 {cell}: row not found")
                continue
            r = rr[0]
            cmp_num(f"§2 {cell} time,s", r[2], p.get("elapsed_s") or 0.0, f1)
            cmp_num(f"§2 {cell} in+out", r[3], p.get("tokens_inout") or 0, f0)
            cmp_num(f"§2 {cell} total", r[4], p.get("tokens_total") or 0, f0)
            cmp_num(f"§2 {cell} cache", r[5], p.get("tokens_cache_read") or 0, f0)

        # §3 operators: | cell | hand | time | in+out | total | decisions | (6 cols)
        for cell in sorted(manifests):
            hand = manifests[cell]["hand"]
            rr = [r for r in find_row(lambda r: r and r[0] == cell and len(r) == 6
                                      and r[1] == hand)]
            o = ev_of[cell]["operator"] or {}
            if not rr:
                bad(f"§3 {cell}: row not found")
                continue
            r = rr[0]
            has_oper = o.get("elapsed_s") is not None
            cmp_num(f"§3 {cell} oper time,s", r[2],
                    o.get("elapsed_s") if has_oper else None, f1)
            io_ = (o.get("tokens_in") or 0) + (o.get("tokens_out") or 0) if has_oper else None
            cmp_num(f"§3 {cell} oper in+out", r[3], io_, f0)
            cmp_num(f"§3 {cell} oper total", r[4],
                    io_ + (o.get("tokens_cache_read") or 0) if has_oper else None, f0)

        # §4 layers: | layer | time | in+out | total |
        tasks = {}
        for cell in sorted(manifests):
            m = manifests[cell]
            tasks.setdefault((m["sense"], m["stack"]), {})[m["hand"]] = cell
        layer_ids = {}
        for cell, entries in reg.items():
            m = manifests.get(cell)
            if not m:
                continue
            for e in entries or []:
                role = e.get("role")
                if role in ("judge", "measurer"):
                    layer_ids.setdefault((role, m["sense"], m["stack"]), []).append(
                        ul.id8_of(e.get("uuid")))
        for (role, sense, stack), idlist in sorted(layer_ids.items()):
            t = 0.0
            s_io = 0
            s_tot = 0
            for i in filter(None, idlist):
                e = ledger.get(i) or {}
                t += e.get("active_s") or 0
                s_io += (e.get("input_tokens") or 0) + (e.get("output_tokens") or 0)
                s_tot += ((e.get("input_tokens") or 0) + (e.get("output_tokens") or 0)
                          + (e.get("cache_tokens") or 0))
            name = f"{role}/{sense}-{stack}"
            rr = find_row(lambda r: len(r) > 3 and r[0] == name)
            if not rr:
                bad(f"§4 {name}: row not found")
                continue
            r = rr[0]
            cmp_num(f"§4 {name} time,s", r[1], t, f1)
            cmp_num(f"§4 {name} in+out", r[2], s_io, f0)
            cmp_num(f"§4 {name} total", r[3], s_tot, f0)

        # §5 deltas: | task | metric | solo | top | solo/top |
        for (sense, stack), hands in sorted(tasks.items()):
            ev = {}
            for h, cell in hands.items():
                p = ev_of[cell]["platform"] or {}
                ev[h] = {
                    "time": p.get("elapsed_s") or 0.0,
                    "inout": p.get("tokens_inout") or 0,
                    "total": p.get("tokens_total") or 0,
                }
            if "solo" not in hands or "top" not in hands:
                continue
            for met, key in (("time, s", "time"), ("in+out", "inout"),
                             ("total", "total")):
                rr = find_row(lambda r: len(r) > 4 and r[0] == f"{sense}-{stack}"
                              and r[1] == met)
                if not rr:
                    continue
                r = rr[0]
                fmtfun = f1 if key == "time" else f0
                cmp_num(f"§5 {sense}-{stack} {met} solo", r[2], ev["solo"][key], fmtfun)
                cmp_num(f"§5 {sense}-{stack} {met} top", r[3], ev["top"][key], fmtfun)
                cmp_num(f"§5 {sense}-{stack} {met} solo/top", r[4],
                        multv(ev["solo"][key], ev["top"][key]), f2)

        # §6 camps: | task | hand | hs | os | cs | hio | oio | cio | ht | ot | ct |
        camp_tot = {"s": [0.0, 0.0], "io": [0, 0], "tot": [0, 0],
                    "hand_s": [0.0, 0.0], "hand_io": [0, 0], "hand_tot": [0, 0],
                    "oper_s": [0.0, 0.0], "oper_io": [0, 0], "oper_tot": [0, 0]}
        for (sense, stack), hands in sorted(tasks.items()):
            camp = {}
            for h in ("solo", "top"):
                if h not in hands:
                    continue
                cell = hands[h]
                p = ev_of[cell]["platform"] or {}
                o = ev_of[cell]["operator"] or {}
                ph, po = p.get("elapsed_s") or 0.0, o.get("elapsed_s") or 0.0
                ih = (p.get("tokens_in") or 0) + (p.get("tokens_out") or 0)
                io_ = (o.get("tokens_in") or 0) + (o.get("tokens_out") or 0)
                th = ih + (p.get("tokens_cache_read") or 0)
                to_ = io_ + (o.get("tokens_cache_read") or 0)
                camp[h] = (ph, po, ph + po, ih, io_, ih + io_, th, to_, th + to_)
                i = 0 if h == "solo" else 1
                camp_tot["s"][i] += ph + po
                camp_tot["io"][i] += ih + io_
                camp_tot["tot"][i] += th + to_
                camp_tot["hand_s"][i] += ph
                camp_tot["hand_io"][i] += ih
                camp_tot["hand_tot"][i] += th
                camp_tot["oper_s"][i] += po
                camp_tot["oper_io"][i] += io_
                camp_tot["oper_tot"][i] += to_
                rr = find_row(lambda r: len(r) > 10 and r[0] == f"{sense}-{stack}"
                              and r[1] == h)
                if not rr:
                    bad(f"§6 {sense}-{stack} {h}: row not found")
                    continue
                r = rr[0]
                v = camp[h]
                for col, val, ff in ((2, v[0], f1), (3, v[1], f1), (4, v[2], f1),
                                     (5, v[3], f0), (6, v[4], f0), (7, v[5], f0),
                                     (8, v[6], f0), (9, v[7], f0), (10, v[8], f0)):
                    cmp_num(f"§6 {sense}-{stack} {h} col{col}", r[col], val, ff)
            if "solo" in camp and "top" in camp:
                s, t = camp["solo"], camp["top"]
                rr = find_row(lambda r: len(r) > 10 and r[0] == f"{sense}-{stack}"
                              and "solo/top" in r[1])
                if rr:
                    r = rr[0]
                    cmp_num(f"§6 {sense}-{stack} × camp,s", r[4], multv(s[2], t[2]), f2)
                    cmp_num(f"§6 {sense}-{stack} × camp in+out", r[7], multv(s[5], t[5]), f2)
                    cmp_num(f"§6 {sense}-{stack} × camp total", r[10], multv(s[8], t[8]), f2)

        # Σ rows: §3 operators by side, §6 camps by side, Σ × row
        for side_i, side in ((0, "solo"), (1, "top")):
            ids = []
            for cell, entries in reg.items():
                m = manifests.get(cell)
                if not m or m["hand"] != side:
                    continue
                for e in entries or []:
                    if (e.get("role") or "hand") == "operator":
                        ids.append(ul.id8_of(e.get("uuid")))
            ids = [i for i in ids if i]
            t = sum((ledger.get(i) or {}).get("active_s") or 0 for i in ids)
            io_ = sum(((ledger.get(i) or {}).get("input_tokens") or 0)
                      + ((ledger.get(i) or {}).get("output_tokens") or 0) for i in ids)
            tot = sum(((ledger.get(i) or {}).get("input_tokens") or 0)
                      + ((ledger.get(i) or {}).get("output_tokens") or 0)
                      + ((ledger.get(i) or {}).get("cache_tokens") or 0) for i in ids)
            rr = find_row(lambda r: len(r) > 5 and r[0].startswith("**Σ operators")
                          and r[1] == side)
            if rr:
                r = rr[0]
                cmp_num(f"§3 Σ operators {side} time,s", r[2], t, f1)
                cmp_num(f"§3 Σ operators {side} in+out", r[3], io_, f0)
                cmp_num(f"§3 Σ operators {side} total", r[4], tot, f0)
            rr = find_row(lambda r: len(r) > 10 and r[0].startswith("**Σ camps")
                          and r[1] == side)
            if rr:
                r = rr[0]
                i = side_i
                for col, val, ff in (
                        (2, camp_tot["hand_s"][i], f1), (3, camp_tot["oper_s"][i], f1),
                        (4, camp_tot["s"][i], f1), (5, camp_tot["hand_io"][i], f0),
                        (6, camp_tot["oper_io"][i], f0), (7, camp_tot["io"][i], f0),
                        (8, camp_tot["hand_tot"][i], f0), (9, camp_tot["oper_tot"][i], f0),
                        (10, camp_tot["tot"][i], f0)):
                    cmp_num(f"§6 Σ camps {side} col{col}", r[col], val, ff)
        rr = find_row(lambda r: len(r) > 10 and r[0].startswith("**Σ solo/top"))
        if rr:
            r = rr[0]
            for col, a, b in ((2, camp_tot["hand_s"][0], camp_tot["hand_s"][1]),
                              (3, camp_tot["oper_s"][0], camp_tot["oper_s"][1]),
                              (4, camp_tot["s"][0], camp_tot["s"][1]),
                              (5, camp_tot["hand_io"][0], camp_tot["hand_io"][1]),
                              (6, camp_tot["oper_io"][0], camp_tot["oper_io"][1]),
                              (7, camp_tot["io"][0], camp_tot["io"][1]),
                              (8, camp_tot["hand_tot"][0], camp_tot["hand_tot"][1]),
                              (9, camp_tot["oper_tot"][0], camp_tot["oper_tot"][1]),
                              (10, camp_tot["tot"][0], camp_tot["tot"][1])):
                cmp_num(f"§6 Σ solo/top × col{col}", r[col], multv(a, b), f2)

        # the camps INDEPENDENT recount: straight from the ledger
        # (agents of each side), not through the evidence files
        say("-- camps independent recount (ledger only, no evidence hop)")
        for side in ("solo", "top"):
            ids = {"hand": [], "operator": []}
            for cell, entries in reg.items():
                m = manifests.get(cell)
                if not m or m["hand"] != side:
                    continue
                for e in entries or []:
                    role = e.get("role") or "hand"
                    if role in ids:
                        ids[role].append(ul.id8_of(e.get("uuid")))

            def agg(idlist, key):
                return sum((ledger.get(i) or {}).get(key) or 0
                           for i in filter(None, idlist))

            hand_s = agg(ids["hand"], "active_s")
            oper_s = agg(ids["operator"], "active_s")
            hand_io = agg(ids["hand"], "input_tokens") + agg(ids["hand"], "output_tokens")
            oper_io = agg(ids["operator"], "input_tokens") + agg(ids["operator"], "output_tokens")
            hand_tot = hand_io + agg(ids["hand"], "cache_tokens")
            oper_tot = oper_io + agg(ids["operator"], "cache_tokens")
            rr = find_row(lambda r: len(r) > 10 and r[0].startswith("**Σ camps")
                          and r[1] == side)
            if not rr:
                bad(f"§6 Σ camps {side}: row not found for the recount")
                continue
            r = rr[0]
            got = (r[2], r[3], r[4], r[5], r[6], r[7], r[8], r[9], r[10])
            want = (f1(hand_s), f1(oper_s), f1(hand_s + oper_s),
                    f0(hand_io), f0(oper_io), f0(hand_io + oper_io),
                    f0(hand_tot), f0(oper_tot), f0(hand_tot + oper_tot))
            if got == want:
                ok(f"§6 Σ camps {side} independent recount from the ledger: "
                   f"hand {got[0]}s/{got[3]}/{got[6]}, oper {got[1]}s/{got[4]}/{got[7]}")
            else:
                bad(f"§6 Σ camps {side} recount: table {got} != ledger {want}")

    say(f"== RESULT: {checks[0]} check(s) AGREE, {red[0]} RED")

    ts = datetime.datetime.now(datetime.timezone.utc).strftime("%Y%m%dT%H%M%SZ")
    vdir = os.path.join(root, "verification")
    os.makedirs(vdir, exist_ok=True)
    vpath = os.path.join(vdir, f"verify-metrics-{ts}.txt")
    with open(vpath, "w", encoding="utf-8") as f:
        f.write("\n".join(out_lines) + "\n")
    print(f"verification written: {vpath}", file=sys.stderr)
    return 1 if red[0] else 0


if __name__ == "__main__":
    sys.exit(main())
