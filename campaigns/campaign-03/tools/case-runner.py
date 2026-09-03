#!/usr/bin/env python3
# case-runner.py — the unified suite runner over an exported punchtape
# case table: the SAME procedure for every product (the blind
# measurement of the wave: the suite runs TWICE on the own product).
# One throwaway container per run (--network none, the wave world
# image); a fresh directory per case; fixtures byte-for-byte; the
# product surface resolved exactly as the row spells it; exact
# comparison of exit/stdout/stderr/disk; a difference that is only
# trailing newlines/spaces counts as pass marked tolerant.
#
# Usage:
#   case-runner.py --product <exported-task-dir> [--runs 2]
#                  [--scenarios SCN-001,SCN-002|auto]
# Output: one JSON line per run on stdout:
#   {"run":n,"pass":N,"fail":M,"tolerant":K,"cases":[...]}
import argparse
import json
import os
import subprocess
import sys
import tempfile

HERE = os.path.dirname(os.path.abspath(__file__))

# The in-container executor: reads /tmp/cases.json (the full table),
# runs every case in a fresh /tmp/case directory against a writable
# copy of /product, prints one JSON verdict line per case.
CASE_PY = r'''
import json, os, shutil, signal, subprocess

cases = json.load(open("/tmp/cases.json"))


def untag(v):
    import base64
    if isinstance(v, dict) and set(v) == {"__b64__"}:
        return base64.b64decode(v["__b64__"]).decode("utf-8", "surrogateescape")
    return v

prod = "/tmp/prod"

# The product's declared surface recipe (kind: conventions): the build
# commands run once in the product copy — {out} is the declared
# artifact path, {name} the alphabetically first command[0] of the
# rows — before any case runs.
build = cases.get("build") or {}
for cmd in build.get("commands") or []:
    argv = [a.replace("{out}", str(build.get("out") or "")).replace("{name}", cases.get("name") or "") for a in cmd]
    r = subprocess.run(argv, cwd=prod, stdin=subprocess.DEVNULL,
                       stdout=subprocess.PIPE, stderr=subprocess.STDOUT, timeout=600)
    if r.returncode != 0:
        print(json.dumps({"id": "BUILD", "pass": False, "tolerant": False,
                          "assertions": 0,
                          "fails": [{"phase": "build", "cmd": argv,
                                     "detail": r.stdout.decode("utf-8", "replace")[-400:]}]}), flush=True)
        sys.exit(1)

def resolve(cmd0, rundir):
    p = os.path.join(rundir, cmd0)
    if os.path.exists(p):
        return p
    src = os.path.join(prod, cmd0)
    if os.path.isfile(src):
        dst = os.path.join(rundir, cmd0)
        os.makedirs(os.path.dirname(dst) or rundir, exist_ok=True)
        shutil.copy2(src, dst)
        os.chmod(dst, 0o755)
        return dst
    return cmd0

def trail(s):
    return s.rstrip(" \n\t\r")

def tolerant_eq(want, got):
    if want == got:
        return True, False
    if trail(want) == trail(got):
        return True, True
    return False, False

for case in cases["cases"]:
    rundir = "/tmp/case"
    shutil.rmtree(rundir, ignore_errors=True)
    os.makedirs(rundir)
    fails = []
    tolerant = 0
    oks = []
    try:
        art = case.get("artifact")
        if not art and not (build_cmd_ran if False else False):
            # no declared recipe: the submitted surface files ride into the
            # run dir directly (the machine's rule for recipe-less surfaces)
            import os as _os
            for item in _os.listdir(prod):
                if item in (".punchtape", "materials", "AGENTS.md", "brief.md", "operator-log.md", "scratch", "solo-product") or item.startswith("delta"):
                    continue
                s0 = _os.path.join(prod, item)
                if _os.path.isfile(s0):
                    d0 = _os.path.join(rundir, item)
                    shutil.copy2(s0, d0)
                    _os.chmod(d0, 0o755)
        if art in (".", ""):
            art = None
            # out "." = the artifact files at the run-dir root
            for item in os.listdir(prod):
                if item == ".punchtape":
                    continue
                s0 = os.path.join(prod, item)
                d0 = os.path.join(rundir, item)
                try:
                    if os.path.isdir(s0):
                        shutil.copytree(s0, d0)
                    else:
                        shutil.copy2(s0, d0)
                        os.chmod(d0, 0o755)
                except OSError:
                    pass
        if art:
            src = os.path.join(prod, art)
            dst = os.path.join(rundir, art)
            if os.path.isdir(src):
                shutil.copytree(src, dst,
                                ignore=shutil.ignore_patterns(".punchtape"))
                os.chmod(dst, 0o755)
                for root, dirs, files in os.walk(dst):
                    for f in files:
                        try:
                            os.chmod(os.path.join(root, f), 0o755)
                        except OSError:
                            pass
            elif os.path.isfile(src):
                os.makedirs(os.path.dirname(dst) or rundir, exist_ok=True)
                shutil.copy2(src, dst)
                os.chmod(dst, 0o755)
        for s in case.get("seed", []):
            dst = os.path.join(rundir, s["path"])
            d = os.path.dirname(dst)
            if d:
                os.makedirs(d, exist_ok=True)
            with open(dst, "wb") as f:
                c = s.get("content")
                if s.get("content_b64"):
                    import base64
                    f.write(base64.b64decode(s["content_b64"]))
                elif isinstance(c, dict) and set(c) == {"__b64__"}:
                    import base64
                    f.write(base64.b64decode(c["__b64__"]))
                elif isinstance(c, bytes):
                    f.write(c)
                else:
                    f.write(c.encode("utf-8", "surrogateescape"))
        for m in case.get("materials", []):
            src = os.path.join(prod, m["from"])
            dst = os.path.join(rundir, m["from"])
            d = os.path.dirname(dst)
            if d:
                os.makedirs(d, exist_ok=True)
            shutil.copy2(src, dst)
        env = dict(os.environ)
        env["PWD"] = rundir
        for pre in case.get("pre", []):
            argv = list(pre)
            argv[0] = resolve(argv[0], rundir)
            r = subprocess.run(argv, cwd=rundir, env=env,
                               stdin=subprocess.DEVNULL,
                               stdout=subprocess.PIPE, stderr=subprocess.PIPE,
                               timeout=case.get("timeout", 15))
            if r.returncode != 0:
                fails.append({"phase": "pre", "cmd": argv, "rc": r.returncode,
                              "stderr": r.stderr.decode("utf-8", "replace")[:400]})
                break
        if not fails:
            argv = list(case["command"])
            argv[0] = resolve(argv[0], rundir)
            try:
                p = subprocess.Popen(argv, cwd=rundir, env=env,
                                     stdout=subprocess.PIPE,
                                     stderr=subprocess.PIPE,
                                     stdin=subprocess.PIPE,
                                     start_new_session=True)
                stdin_v = untag(case.get("stdin"))
                out, err = p.communicate(
                    input=(stdin_v or "").encode("utf-8", "surrogateescape"),
                    timeout=case.get("timeout", 15))
                rc = p.returncode
            except subprocess.TimeoutExpired:
                try:
                    os.killpg(os.getpgid(p.pid), signal.SIGKILL)
                except Exception:
                    pass
                out, err = p.communicate()
                rc = "timeout"
            got = {"stdout": out.decode("utf-8", "surrogateescape"),
                   "stderr": err.decode("utf-8", "surrogateescape"),
                   "exit-code": rc}
            for a in case.get("then", []):
                obs, cond = a["observation"], a["condition"]
                want = untag(a.get("value"))
                tol = False
                ok = True
                detail = ""
                if obs in ("stdout", "stderr"):
                    g = got[obs]
                    if cond == "equals":
                        ok, tol = tolerant_eq(want, g)
                    elif cond == "contains":
                        ok = want in g
                    elif cond == "absent":
                        ok = g == ""
                    elif cond == "json-equals":
                        try:
                            ok = json.loads(g) == json.loads(want)
                        except Exception:
                            ok = False
                    elif cond == "json-contains":
                        try:
                            ok = json.loads(want) in json.loads(g)
                        except Exception:
                            ok = False
                    else:
                        ok = False
                    if not ok:
                        detail = "want %r got %r" % (want, g[:400])
                elif obs == "exit-code":
                    g = got["exit-code"]
                    if cond == "equals":
                        ok = str(g) == str(want)
                    elif cond == "fails":
                        ok = isinstance(g, int) and g != 0
                    if not ok:
                        detail = "want %s got %s" % (want, g)
                elif obs == "file":
                    path = os.path.join(rundir, a["path"])
                    g = None
                    if os.path.isfile(path):
                        with open(path, "rb") as f:
                            g = f.read().decode("utf-8", "surrogateescape")
                    if cond == "exists":
                        ok = os.path.exists(path)
                    elif cond == "not-exists":
                        ok = not os.path.exists(path)
                    elif cond == "absent":
                        ok = g is None
                    elif g is None:
                        ok = False
                    elif cond == "equals":
                        ok, tol = tolerant_eq(want, g)
                    elif cond == "contains":
                        ok = want in g
                    elif cond == "bytes-equals":
                        ok, tol = tolerant_eq(want, g)
                    elif cond == "json-equals":
                        try:
                            ok = json.loads(g) == json.loads(want)
                        except Exception:
                            ok = False
                    elif cond == "json-contains":
                        try:
                            ok = json.loads(want) in json.loads(g)
                        except Exception:
                            ok = False
                    else:
                        ok = False
                    if not ok:
                        detail = "file %s %s: want %r got %r" % (
                            a.get("path"), cond, want, (g or "")[:400])
                else:
                    ok = False
                    detail = "unknown observation %s" % obs
                if tol:
                    tolerant += 1
                oks.append(ok)
                if not ok:
                    fails.append({"assertion": a, "detail": detail})
    except Exception as e:
        fails.append({"phase": "runner", "detail": repr(e)[:400]})
    print(json.dumps({"id": case["id"],
                      "pass": bool(oks) and all(oks) and not fails,
                      "tolerant": tolerant > 0,
                      "assertions": len(oks),
                      "fails": fails}), flush=True)
'''

INNER = ("rm -rf /tmp/prod && cp -a /product /tmp/prod && chmod -R u+w /tmp/prod && "
         "python3 /tmp/runner.py")


def load_yaml(path):
    import yaml
    with open(path) as f:
        return norm_bytes(yaml.safe_load(f))


def norm_bytes(o):
    # !!binary values arrive from yaml as bytes; JSON cannot carry them.
    # Tag as {"__b64__": ...} — the in-container executor untags back to
    # bytes and compares exactly (the same discipline as seed content_b64).
    import base64
    if isinstance(o, bytes):
        return {"__b64__": base64.b64encode(o).decode()}
    if isinstance(o, dict):
        return {k: norm_bytes(v) for k, v in o.items()}
    if isinstance(o, list):
        return [norm_bytes(v) for v in o]
    return o


def load_conv(product):
    p = os.path.join(product, ".punchtape", "passport", "conventions.yaml")
    return load_yaml(p) if os.path.isfile(p) else None


def load_table(product):
    d = os.path.join(product, ".punchtape", "canon", "scenarios")
    return [load_yaml(os.path.join(d, n)) for n in sorted(os.listdir(d))]


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--product", required=True)
    ap.add_argument("--runs", type=int, default=2)
    ap.add_argument("--scenarios", default="auto")
    ap.add_argument("--image", default="punchtape-world:th3")
    args = ap.parse_args()

    scns = load_table(args.product)
    if args.scenarios != "auto":
        want = set(args.scenarios.split(","))
        scns = [s for s in scns if s["id"] in want]

    cases = []
    for sc in scns:
        when = sc.get("when") or {}
        import base64
        seed = []
        for s in (sc.get("seed") or []):
            c = s.get("content")
            if isinstance(c, bytes):
                seed.append({"path": s["path"],
                             "content_b64": base64.b64encode(c).decode()})
            else:
                seed.append({"path": s["path"], "content": c})
        cases.append({
            "id": sc["id"],
            "artifact": (load_conv(args.product) or {}).get("build", {}).get("out"),
            "seed": seed,
            "materials": sc.get("materials") or [],
            "pre": sc.get("pre") or [],
            "command": when.get("command"),
            "stdin": when.get("stdin"),
            "timeout": when.get("timeout-sec") or 15,
            "then": sc.get("then") or [],
        })

    for run in range(1, args.runs + 1):
        with tempfile.TemporaryDirectory(prefix="case-run-") as td:
            conv = load_conv(args.product) or {}
            payload = {"build": (conv.get("build") or {}),
                       "name": min([c["command"][0] for c in cases if c.get("command")] or [""]),
                       "cases": cases}
            with open(os.path.join(td, "cases.json"), "w") as f:
                json.dump(payload, f)
            with open(os.path.join(td, "runner.py"), "w") as f:
                f.write(CASE_PY)
            cmd = ["docker", "run", "--rm", "--network", "none",
                   "--cpus", "1", "--memory", "1g",
                   "-v", td + ":/tmp-host:ro",
                   "-v", os.path.abspath(args.product) + ":/product:ro",
                   args.image, "sh", "-c",
                   "cp /tmp-host/cases.json /tmp/cases.json && "
                   "cp /tmp-host/runner.py /tmp/runner.py && " + INNER]
            r = subprocess.run(cmd, stdout=subprocess.PIPE,
                               stderr=subprocess.PIPE,
                               timeout=max(900, 120 * len(cases)))
            results = []
            for line in r.stdout.decode().splitlines():
                line = line.strip()
                if line.startswith("{"):
                    try:
                        results.append(json.loads(line))
                    except json.JSONDecodeError:
                        pass
            if not results:
                results = [{"id": "ALL", "pass": False, "tolerant": False,
                            "assertions": 0,
                            "fails": [{"phase": "container",
                                       "detail": r.stderr.decode()[-500:]}]}]
        n_pass = sum(1 for x in results if x["pass"] and not x["tolerant"])
        n_tol = sum(1 for x in results if x["pass"] and x["tolerant"])
        n_fail = sum(1 for x in results if not x["pass"])
        print(json.dumps({"run": run, "pass": n_pass, "fail": n_fail,
                          "tolerant": n_tol, "cases": results}))
        sys.stdout.flush()


if __name__ == "__main__":
    main()
