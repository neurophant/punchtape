#!/usr/bin/env python3
"""check-staged-freeze.py — the staging-vs-freeze check of the
case-0 campaign (the P6 law): every staged input (brief,
expectations, materials, probes) must match the freeze line of its
campaign source; a post-freeze edit of any input is a staging stop.
Called by stage-cell.sh after the copy; the selftest drives it
directly against doctored copies.

Usage: check-staged-freeze.py <freeze-file> <sense> <stack> <input-dir>
Exit 0 = matches; 1 = drift (each divergence printed).
"""
import hashlib
import os
import sys


def main():
    freeze, sense, stack, inputdir = sys.argv[1:5]
    # keyed (kind, repo-rel): two kinds may share one repo path — the
    # brief itself (kind brief) and its generated form-hints sidcar
    # (kind form-hints) both hang off
    # briefs/<sense>/<stack>.md
    frozen = {}
    for line in open(freeze, encoding="utf-8"):
        parts = line.split()
        if len(parts) >= 3 and parts[0] in ("brief", "material", "prompt",
                                            "probe", "prices", "world-lock",
                                            "models", "form-hints"):
            frozen[(parts[0], parts[2])] = parts[1]
    bad = []

    def check(local_rel, repo_rel, kind):
        f = os.path.join(inputdir, local_rel)
        if not os.path.isfile(f):
            return
        got = hashlib.sha256(open(f, "rb").read()).hexdigest()
        want = frozen.get((kind, repo_rel))
        if want is None:
            print(f"FREEZE MISS: {repo_rel} ({kind}) is not in the freeze")
            bad.append(repo_rel)
        elif got != want:
            print(f"FREEZE DRIFT: {repo_rel} ({kind}) staged sha {got[:12]} != frozen {want[:12]}")
            bad.append(repo_rel)

    # the per-stack brief of 0.2.0 and the common expectations list
    check("brief.md", f"briefs/{sense}/{stack}.md", "brief")
    check("EXPECTATIONS.md", "prompts/EXPECTATIONS.md", "prompt")

    # materials: the staged subtree is the sense's own (brim: the two
    # jpegs at the top; pressmark: the pressmark subtree; quire: the
    # stack's subtree) — staged files map back onto the campaign tree
    matdir = os.path.join(inputdir, "materials")
    # conduit and errand carry no materials (stage-cell stages none for
    # them): the walk below never runs — the entries keep the sense set
    # closed instead of a KeyError before the first cell of such a wave
    prefix = {
        "brim": "materials/",
        "pressmark": "materials/pressmark/",
        "quire": f"materials/quire/{stack}/",
        "conduit": "materials/",
        "errand": "materials/",
    }[sense]
    for root, _, files in os.walk(matdir):
        for fn in sorted(files):
            f = os.path.join(root, fn)
            sub = os.path.relpath(f, matdir)
            check(os.path.join("materials", sub), prefix + sub, "material")

    # the staged probe bytes (the A2 witness materials)
    probedir = os.path.join(inputdir, "probes")
    for fn in sorted(os.listdir(probedir) if os.path.isdir(probedir) else []):
        f = os.path.join(probedir, fn)
        if os.path.isfile(f):
            check(os.path.join("probes", fn), f"materials/probes/{fn}", "probe")

    # the form-hints sidcar: a
    # machine cell stages it (solo cells never do — the check below
    # only runs for the staged file); the frozen digest is the same
    # generator run over the same brief bytes at freeze time, so a
    # doctored sidcar is a drift and a regenerated-but-unfrozen one is
    # a miss
    check("form-hints.md", f"briefs/{sense}/{stack}.md", "form-hints")

    sys.exit(1 if bad else 0)


if __name__ == "__main__":
    main()
