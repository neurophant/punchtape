#!/usr/bin/env python3
"""case-runner-host.py — the unified suite runner of the case-0
campaign (the 0.2.0 blind measurement on the host world): the SAME
procedure for every product —
the hand's exported corpus replayed TWICE, each run in a fresh
throwaway directory, byte-exact comparison, a doubt counts against
the product.

The corpus (canon, schema 9): <product>/.punchtape/canon/
atoms/*.yaml — behaviors with rows. A row replays exactly as the
machine's grammar spells it: seeds (text | from-input of the cell's
input/), the run command, the observations (out/err exact with ONE
trailing newline tolerated, the (contains)/(contains-loose) forms,
rc, state text/from-run), the procedural modes (launch: nonet — a
fresh network-less user+net namespace, tree: unchanged — the row-
scoped tree identity), and the native typed heads (witness — the
staged probe bytes of the cell's input/probes/, repeat-run — two
identical runs). A bytes: file:// sample names a file that rode the
hand's inbox and is gone at export — the row is honestly
not-replayable and counts against the product's proof (the blind
standard does not accept what it cannot see).

Solo products have no machine corpus: P1 (solo-product/check.sh)
runs by the measurer's own procedure — this tool serves the machine
hands' corpus replay.

Usage:
  case-runner-host.py --product <dir> [--input <cell-input>]
                      [--runs 2] [--scratch <dir>]
Output: one JSON line per run on stdout:
  {"run":n,"pass":N,"fail":M,"tolerant":K,"env_fail":E,"cases":[...]}
"""
import argparse
import hashlib
import json
import os
import shutil
import signal
import subprocess
import sys
import tempfile

import yaml

CONTAINS = "(contains) "
CONTAINS_LOOSE = "(contains-loose) "
DEFAULT_TIMEOUT = 60
# build-ignore carries no bytes of its own here: the replay works on
# the exported tree AS LEFT (artifacts present); a row whose artifact
# is missing reds with the reason named


def load_atoms(product):
    d = os.path.join(product, ".punchtape", "canon", "atoms")
    if not os.path.isdir(d):
        return None
    out = []
    for n in sorted(os.listdir(d)):
        if not n.endswith(".yaml"):
            continue
        with open(os.path.join(d, n), encoding="utf-8") as f:
            out.append(yaml.safe_load(f))
    return out


def expect_of(v):
    """A TextExpect value from YAML: None (absent), or the string
    forms — exact / (contains) / (contains-loose)."""
    if v is None:
        return None
    if isinstance(v, dict):  # never in the stored canon; be honest
        return ("exact", yaml.safe_dump(v, default_flow_style=True).strip())
    if v.startswith(CONTAINS_LOOSE):
        return ("loose", v[len(CONTAINS_LOOSE):])
    if v.startswith(CONTAINS):
        return ("contains", v[len(CONTAINS):])
    return ("exact", v)


def trail(s):
    return s.rstrip(" \n\t\r")


def tree_digest(root, seeds):
    """The tree identity outside .punchtape/ and the row's own seeds
    (the tree: unchanged mode)."""
    h = hashlib.sha256()
    for base, dirs, files in os.walk(root):
        dirs[:] = sorted(d for d in dirs if d not in (".punchtape", "node_modules",
                                                      "target", ".git"))
        for f in sorted(files):
            p = os.path.join(base, f)
            rel = os.path.relpath(p, root)
            if rel in seeds:
                continue
            h.update(rel.encode("utf-8", "surrogateescape"))
            with open(p, "rb") as fh:
                h.update(fh.read())
    return h.hexdigest()


def resolve(cmd0, rundir):
    p = os.path.join(rundir, cmd0)
    if os.path.exists(p):
        return p
    return cmd0


def wrap_launch(argv, launch):
    """launch: nonet — a fresh network-less user+net namespace, the
    same discipline the machine's runner uses (CLONE_NEWUSER |
    CLONE_NEWNET); unshare -rn is its command-line face."""
    if launch == "nonet":
        return ["unshare", "-rn", "--"] + argv
    return argv


def run_argv(argv, cwd, env, timeout, stdin_bytes=b""):
    try:
        p = subprocess.Popen(argv, cwd=cwd, env=env,
                             stdout=subprocess.PIPE,
                             stderr=subprocess.PIPE,
                             stdin=subprocess.PIPE,
                             start_new_session=True)
        out, err = p.communicate(input=stdin_bytes, timeout=timeout)
        rc = p.returncode
    except subprocess.TimeoutExpired:
        try:
            os.killpg(os.getpgid(p.pid), signal.SIGKILL)
        except Exception:
            pass
        out, err = p.communicate()
        rc = "timeout"
    return out, err, rc


def check_expect(kind, exp, got):
    """-> (ok, tolerant, detail). kind: out|err|state-text.

    The EOL rule is the replay METHOD's own normalization: ONE
    trailing newline on either side is the same text — a STRICT
    pass, not a tolerant marker (the
    canon stores expectations without the trailing \n, the real
    streams carry it); the tolerant marker is reserved for honest
    oddities the method still accepts, and nothing currently
    produces one."""
    mode, want = exp
    g = got
    if mode in ("contains", "loose"):
        ok = want in g
        if not ok:
            return False, False, f"{kind}: want CONTAINS {want!r}, got {g[:200]!r}"
        return True, False, ""
    if want == g:
        return True, False, ""
    if want + "\n" == g or g + "\n" == want:
        return True, False, ""
    return False, False, f"{kind}: want {want[:200]!r}, got {g[:200]!r}"


def replay_row(atom, idx, row, rundir, input_dir, probe_dir):
    """Replay ONE row. -> the case record dict."""
    rid = f"{atom.get('id', 'atom')}/{idx + 1}"
    fails = []
    tolerant = 0
    env_fail = False
    seed_paths = []

    def seed_of(s):
        return s.get("path")

    try:
        # seeds: written before, removed after (their original bytes
        # restored where they overwrote — the machine's rule)
        originals = {}
        for s in row.get("seeds") or []:
            dst = os.path.join(rundir, s["path"])
            os.makedirs(os.path.dirname(dst) or rundir, exist_ok=True)
            if os.path.exists(dst):
                with open(dst, "rb") as fh:
                    originals[dst] = fh.read()
            if s.get("text") is not None and s.get("text") != "":
                with open(dst, "w", encoding="utf-8") as fh:
                    fh.write(s["text"])
            elif s.get("from-input"):
                src = os.path.join(input_dir, s["from-input"])
                shutil.copy2(src, dst)
            seed_paths.append(s["path"])

        launch = row.get("launch") or ""
        tree_mode = row.get("tree") or ""
        timeout = int(row.get("timeout") or DEFAULT_TIMEOUT)
        pre_digest = tree_digest(rundir, set(seed_paths)) if tree_mode == "unchanged" else None

        witness = row.get("witness") or {}
        repeat = row.get("repeat-run")

        if witness:
            # the staged-bytes probe head: the probe's bytes are the
            # cell's staged input/probes (the campaign materialized
            # them at staging; the door copy is gone at export)
            name = (witness.get("probe") or "").removeprefix("file://")
            probe = os.path.join(probe_dir, name)
            if not os.path.isfile(probe):
                fails.append({"assertion": "witness", "detail":
                              f"probe {name} not staged in the cell input"})
            else:
                argv = wrap_launch(["sh", probe], launch)
                out, err, rc = run_argv(argv, rundir, dict(os.environ, PWD=rundir),
                                        timeout)
                if rc == "timeout" and not isinstance(rc, int):
                    fails.append({"assertion": "witness", "detail": "timeout"})
                exp = expect_of(row.get("out"))
                if exp:
                    ok, tol, d = check_expect("witness-out", exp,
                                              out.decode("utf-8", "surrogateescape"))
                    tolerant += 1 if (ok and tol) else 0
                    if not ok:
                        fails.append({"assertion": "witness-out", "detail": d})
                if row.get("rc") is not None and str(rc) != str(row["rc"]):
                    fails.append({"assertion": "witness-rc",
                                  "detail": f"want {row['rc']} got {rc}"})
        elif repeat:
            rep = repeat if isinstance(repeat, list) else None
            if rep is None:
                import shlex
                rep = shlex.split(str(repeat))
            argv = [str(a) for a in rep]
            argv[0] = resolve(argv[0], rundir)
            argv = wrap_launch(argv, launch)
            env = dict(os.environ, PWD=rundir)
            o1, e1, r1 = run_argv(argv, rundir, env, timeout)
            o2, e2, r2 = run_argv(argv, rundir, env, timeout)
            if (o1, e1, str(r1)) != (o2, e2, str(r2)):
                fails.append({"assertion": "repeat-run",
                              "detail": "two runs differ (out/err/rc)"})
        else:
            cmd = row.get("run")
            if cmd is None:
                fails.append({"assertion": "row", "detail": "no command head"})
            else:
                # the machine's two forms: an argv list, or one string
                # split by spaces with quotes grouping — the replay
                # NEVER mutates the atom's own list (an aliased row
                # would poison every later run)
                if isinstance(cmd, list):
                    argv = [str(a) for a in cmd]
                else:
                    import shlex
                    argv = shlex.split(str(cmd))
                argv[0] = resolve(argv[0], rundir)
                argv = wrap_launch(argv, launch)
                out, err, rc = run_argv(argv, rundir, dict(os.environ, PWD=rundir),
                                        timeout)
                if rc == "timeout":
                    fails.append({"assertion": "timeout", "detail":
                                  f"row exceeded {timeout}s"})
                else:
                    for field, got in (("out", out.decode("utf-8", "surrogateescape")),
                                       ("err", err.decode("utf-8", "surrogateescape"))):
                        exp = expect_of(row.get(field))
                        if exp is None:
                            continue
                        ok, tol, d = check_expect(field, exp, got)
                        tolerant += 1 if (ok and tol) else 0
                        if not ok:
                            fails.append({"assertion": field, "detail": d})
                    want_rc = row.get("rc")
                    if want_rc is None:
                        want_rc = 0
                    if str(rc) != str(want_rc):
                        fails.append({"assertion": "exit-code",
                                      "detail": f"want {want_rc} got {rc}"})
                # state checks
                for st in row.get("state") or []:
                    p = os.path.join(rundir, st["path"])
                    if st.get("from-run"):
                        src = os.path.join(rundir, st["from-run"])
                        with open(p, "rb") as a, open(src, "rb") as b:
                            if a.read() != b.read():
                                fails.append({"assertion": "state",
                                              "detail": f"{st['path']} not byte-equal with {st['from-run']}"})
                        continue
                    exp = expect_of(st.get("text"))
                    if exp is None:
                        continue
                    g = None
                    if os.path.isfile(p):
                        with open(p, "rb") as fh:
                            g = fh.read().decode("utf-8", "surrogateescape")
                    elif exp[1] == "":
                        g = None  # absent file satisfies the empty want
                    if g is None and exp[1] != "":
                        fails.append({"assertion": "state",
                                      "detail": f"{st['path']} missing"})
                        continue
                    ok, tol, d = check_expect(f"state {st['path']}", exp, g or "")
                    tolerant += 1 if (ok and tol) else 0
                    if not ok:
                        fails.append({"assertion": "state", "detail": d})
            # bytes: the sample rode the inbox and is gone at export —
            # honestly not replayable by the blind procedure
            if row.get("bytes"):
                fails.append({"assertion": "bytes", "detail":
                              "file:// sample is not replayable blind (the inbox copy is gone at export)"})

        if tree_mode == "unchanged":
            post = tree_digest(rundir, set(seed_paths))
            if pre_digest != post:
                dirty = "row left bytes behind (tree: unchanged)"
                fails.append({"assertion": "tree", "detail": dirty})
    except subprocess.SubprocessError as e:
        env_fail = True
        fails.append({"assertion": "env", "detail": repr(e)[:200]})
    except OSError as e:
        env_fail = True
        fails.append({"assertion": "env", "detail": f"{e}"})
    finally:
        for sp in seed_paths:
            p = os.path.join(rundir, sp)
            if os.path.exists(p):
                os.remove(p)

    return {"id": rid, "pass": not fails, "tolerant": tolerant > 0,
            "env_fail": env_fail, "fails": fails}


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--product", required=True)
    ap.add_argument("--input", default=None,
                    help="the cell's input/ (from-input seeds, probes)")
    ap.add_argument("--runs", type=int, default=2)
    ap.add_argument("--scratch", default=None)
    args = ap.parse_args()

    atoms = load_atoms(args.product)
    if atoms is None:
        print(json.dumps({"run": 0, "pass": 0, "fail": 0, "tolerant": 0,
                          "env_fail": 0, "cases": [], "error":
                          "no .punchtape/canon/atoms under the product"}))
        sys.exit(0)

    input_dir = args.input or os.path.join(args.product, "..", "input")
    probe_dir = os.path.join(input_dir, "probes")

    for run in range(1, args.runs + 1):
        results = []
        # the owner's law: every run and every scratch of
        # this campaign lives under ~/.experiments ONLY — the default
        # scratch is its sanctioned tmp/ zone, NEVER the system /tmp
        scratch_base = args.scratch or os.path.join(
            os.path.expanduser("~"), ".experiments", "tmp")
        os.makedirs(scratch_base, exist_ok=True)
        with tempfile.TemporaryDirectory(prefix="c0-case-run-",
                                         dir=scratch_base) as td:
            prod = os.path.join(td, "prod")
            shutil.copytree(args.product, prod,
                            ignore=shutil.ignore_patterns(".punchtape"))
            # make the tree the row world: fresh dir per row (the
            # 0.2.0 discipline — cross-case interference cannot hide)
            for atom in atoms:
                for i, row in enumerate(atom.get("rows") or []):
                    rundir = os.path.join(td, "case")
                    shutil.rmtree(rundir, ignore_errors=True)
                    shutil.copytree(prod, rundir)
                    results.append(replay_row(atom, i, row, rundir,
                                              input_dir, probe_dir))
        n_pass = sum(1 for x in results if x["pass"] and not x["tolerant"])
        n_tol = sum(1 for x in results if x["pass"] and x["tolerant"])
        n_fail = sum(1 for x in results if not x["pass"])
        n_env = sum(1 for x in results if x["env_fail"])
        print(json.dumps({"run": run, "pass": n_pass, "fail": n_fail,
                          "tolerant": n_tol, "env_fail": n_env,
                          "cases": results}))
        sys.stdout.flush()


if __name__ == "__main__":
    main()
