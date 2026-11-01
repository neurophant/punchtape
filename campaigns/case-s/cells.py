"""Prepare the cells of one wave.

A cell is one hand's isolated world, laid out as input/, env/ and
work/: the brief and its material as read-only inputs, the stack's
pinned libraries and scratch directories as real bytes, the hand's
own working tree in work/ with the cell's ENV file in it — every
command the hand runs goes through work/ and its exports, and the
machine passes the tree's own ENV to every row it executes, so the
cell's substrate rides both kinds identically. Every hand of the
same brief gets byte-identical inputs; the script proves it by
hashing what it copied and refusing to finish on a mismatch.

The series' law: a full run is every meaning in
ALL FIVE stacks — a wave config that names a meaning without its
complete stack set is refused, and so is a stack whose warmed master
is missing from the cache area (a wave never builds the substrate it
stands on; platform maintenance warms it, offline-tested per combo).
The cells carry nothing machine-side: the machine's command is the
system-installed one, and the script refuses to run unless the
system binary is exactly the pinned build (version and checksum
checked). Each hand's spawn bytes are the campaign's harness
preamble (environment mechanics, identical for both kinds) followed
by the hand's own text, the cell's paths rendered in.
"""

from __future__ import annotations

import argparse
import datetime
import hashlib
import json
import os
import shutil
import subprocess
import sys
from pathlib import Path
from typing import Any

import yaml

CAMPAIGN = Path(__file__).resolve().parent
SPAWN_SOURCE = CAMPAIGN / "spawn.yaml"
HARNESS_SOURCE = CAMPAIGN / "harness.md"
# The platform cache area: the stacks' master environments live here,
# outside the repository and outside every wave — platform
# maintenance warms them once per series; a wave only checks and
# copies (a wave never builds the substrate it stands on).
ENV_PREP = Path(
    os.environ.get("C0_ENV_PREP", Path.home() / ".experiments/cache/env-prep")
)

# The series' stack set: one meaning runs in all of them, always.
STACKS = ["go", "js", "python", "rust", "cpp"]

# The warmed master each combo stands on; a combo absent here needs
# no master (the platform prefix or the stdlib carries it). The cpp
# stacks read the platform prefix (~/.local: md4c, libyaml, hiredis,
# GraphicsMagick++ with its -config wrappers, inja and nlohmann
# headers), checked at staging time.
MASTERS: dict[tuple[str, str], str] = {
    ("quire", "go"): "_gomod-quire",
    ("pressmark", "go"): "_gomod-pressmark",
    ("errand", "go"): "_gomod-errand",
    ("quire", "js"): "_nodemods-quire",
    ("brim", "js"): "_nodemods-brim",
    ("errand", "js"): "_nodemods-errand",
    ("pressmark", "js"): "_nodemods-pressmark",
    ("quire", "python"): "_venv-quire",
    ("brim", "python"): "_venv-brim",
    ("errand", "python"): "_venv-errand",
    ("pressmark", "python"): "_venv-pressmark",
    ("quire", "rust"): "_cargo-quire",
    ("brim", "rust"): "_cargo-brim",
    ("conduit", "rust"): "_cargo-conduit",
    ("errand", "rust"): "_cargo-errand",
    ("pressmark", "rust"): "_cargo-pressmark",
}

# The cpp platform prefix's own pieces, named for the refusal.
CPP_PREFIX_PIECES = [
    "include/md4c.h",
    "include/md4c-html.h",
    "include/yaml.h",
    "include/hiredis/hiredis.h",
    "include/GraphicsMagick/Magick++.h",
    "include/inja/inja.hpp",
    "include/nlohmann/json.hpp",
    "lib/libmd4c.so",
    "lib/libyaml.so",
    "lib/libhiredis.so",
    "lib/libGraphicsMagick++.so",
    "bin/GraphicsMagick++-config",
]

# The exports of a go-stack cell: all scratch and caches ride the
# cell, module resolution goes offline through the cell's own copy
# of the pinned libraries (a file proxy over the download cache —
# GOPROXY=off would forbid even that resolution).
GO_ENV = """export TMPDIR="{cell}/env/tmp"
export GOMODCACHE="{cell}/env/deps/gomod-cache"
export GOCACHE="{cell}/env/deps/gocache"
export GOPATH="{cell}/env/deps/gopath"
export GOPROXY="file://{cell}/env/deps/gomod-cache/cache/download"
export GOSUMDB=off
export GOTOOLCHAIN=local
"""

JS_ENV = """export TMPDIR="{cell}/env/tmp"
"""

PY_ENV = """export TMPDIR="{cell}/env/tmp"
export PATH="{cell}/env/venv/bin:{path}"
"""

RUST_ENV = """export TMPDIR="{cell}/env/tmp"
export CARGO_HOME="{cell}/env/deps/cargo-home"
export CARGO_NET_OFFLINE=true
"""

# The rust master is a vendored source tree: resolution and build
# both run purely from the cell's own copy, no registry, no network.
RUST_CONFIG = """[source.crates-io]
replace-with = "vendored-sources"

[source.vendored-sources]
directory = "{cell}/env/deps/cargo-vendor"
"""

# The cpp stacks read the platform prefix through the plain
# compiler variables; the -config wrapper rides PATH.
CPP_ENV = """export TMPDIR="{cell}/env/tmp"
export CPLUS_INCLUDE_PATH="{prefix}/include"
export LIBRARY_PATH="{prefix}/lib"
export LD_LIBRARY_PATH="{prefix}/lib"
export PATH="{prefix}/bin:{path}"
"""


def parse_args(argv: list[str]) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--wave", required=True, help="wave config file")
    parser.add_argument("--root", required=True, help="directory for wave roots")
    parser.add_argument(
        "--machine",
        required=True,
        help="the pinned machine build; the system-installed punchtape must be it",
    )
    return parser.parse_args(argv)


def copy_tree(src: Path, dst: Path) -> None:
    shutil.copytree(src, dst)


def copy_bytes(src: Path, dst: Path) -> None:
    shutil.copyfile(src, dst)
    dst.chmod(0o644)


def tree_digest(root: Path) -> str:
    """The content identity of a copied tree: sorted paths and bytes."""
    digest = hashlib.sha256()
    for path in sorted(p for p in root.rglob("*") if p.is_file()):
        digest.update(str(path.relative_to(root)).encode())
        digest.update(path.read_bytes())
    return digest.hexdigest()


def check_system_machine(pin: Path) -> dict[str, str]:
    """The system punchtape must be the pinned build, exactly.

    The hand calls the machine by name; a stale or foreign system
    binary would quietly measure the wrong machine.
    """
    system = shutil.which("punchtape")
    if system is None:
        print(
            "punchtape is not installed on the system (command -v punchtape)",
            file=sys.stderr,
        )
        raise SystemExit(1)
    system_path = Path(system).resolve()
    pin_sha = hashlib.sha256(pin.read_bytes()).hexdigest()
    sys_sha = hashlib.sha256(system_path.read_bytes()).hexdigest()
    if pin_sha != sys_sha:
        print(
            f"system punchtape ({sys_sha}) != pinned ({pin_sha}): install the pinned binary",
            file=sys.stderr,
        )
        raise SystemExit(1)
    version = subprocess.run(
        [str(system_path), "--version"], capture_output=True, text=True, check=True
    ).stdout.strip()
    return {"path": str(system_path), "version": version, "sha256": sys_sha}


def wave_pairs(config: dict[str, Any]) -> list[tuple[str, str]]:
    """The wave's meaning×stack pairs, the full-set law checked.

    A brief entry may name one meaning (then the config's stacks must
    be exactly the series' five) or one meaning×stack pair outright.
    """
    stacks = [str(s) for s in config.get("stacks", STACKS)]
    out: list[tuple[str, str]] = []
    for brief in [str(b) for b in config["briefs"]]:
        if "/" in brief:
            out.append(tuple(brief.split("/", 1)))  # type: ignore[arg-type]
            continue
        if sorted(stacks) != sorted(STACKS):
            print(
                f"the series' law: meaning {brief!r} runs in ALL FIVE stacks; "
                f"the config names {sorted(stacks)}",
                file=sys.stderr,
            )
            raise SystemExit(1)
        out.extend((brief, stack) for stack in STACKS)
    return out


def check_masters(pairs: list[tuple[str, str]]) -> None:
    """Every warmed master the wave stands on must exist, and the cpp
    platform prefix must carry its pieces — a wave never builds the
    substrate it stands on."""
    for sense, stack in pairs:
        master = MASTERS.get((sense, stack))
        if master is not None and not (ENV_PREP / master).is_dir():
            print(
                f"no warmed master {master} at {ENV_PREP} for {sense}/{stack} — "
                "platform maintenance warms the cache area",
                file=sys.stderr,
            )
            raise SystemExit(1)
        if stack == "cpp":
            prefix = Path.home() / ".local"
            for piece in CPP_PREFIX_PIECES:
                if not (prefix / piece).exists():
                    print(
                        f"the cpp platform prefix misses {prefix / piece} — "
                        "platform maintenance installs it",
                        file=sys.stderr,
                    )
                    raise SystemExit(1)


def stage_stack(cell: Path, sense: str, stack: str, staged_path: str) -> None:
    """The stack's substrate into the cell and its ENV file."""
    work = cell / "work"
    if stack == "go":
        master = MASTERS.get((sense, "go"))
        if master is not None:
            copy_tree(ENV_PREP / master, cell / "env" / "deps" / "gomod-cache")
        else:
            (cell / "env" / "deps" / "gomod-cache").mkdir(parents=True)
        (cell / "env" / "deps" / "gocache").mkdir()
        (cell / "env" / "deps" / "gopath").mkdir()
        (work / "ENV").write_text(GO_ENV.format(cell=cell))
    elif stack == "js":
        master = MASTERS.get((sense, "js"))
        if master is not None:
            copy_tree(ENV_PREP / master, work / "node_modules")
        (work / "ENV").write_text(JS_ENV.format(cell=cell))
    elif stack == "python":
        master = MASTERS.get((sense, "python"))
        if master is not None:
            copy_tree(ENV_PREP / master, cell / "env" / "venv")
        (work / "ENV").write_text(PY_ENV.format(cell=cell, path=staged_path))
    elif stack == "rust":
        copy_tree(
            ENV_PREP / MASTERS[(sense, "rust")], cell / "env" / "deps" / "cargo-vendor"
        )
        (cell / "env" / "deps" / "cargo-home").mkdir(parents=True)
        (work / ".cargo").mkdir()
        (work / ".cargo" / "config.toml").write_text(RUST_CONFIG.format(cell=cell))
        (work / "ENV").write_text(RUST_ENV.format(cell=cell))
    elif stack == "cpp":
        (work / "ENV").write_text(
            CPP_ENV.format(cell=cell, prefix=Path.home() / ".local", path=staged_path)
        )
    else:
        print(f"unknown stack {stack!r}", file=sys.stderr)
        raise SystemExit(1)
    # The errand bench brings up its own redis from the platform
    # prefix; its loader path rides every errand cell.
    if sense == "errand":
        with (work / "ENV").open("a") as env:
            env.write(f'export LD_LIBRARY_PATH="{Path.home()}/.local/lib"\n')


def main(argv: list[str]) -> int:
    args = parse_args(argv)
    wave_file = Path(args.wave).resolve()
    root_dir = Path(args.root)
    pin = Path(args.machine).resolve()
    if not wave_file.is_file():
        print(f"wave config not found: {wave_file}", file=sys.stderr)
        return 1
    config = yaml.safe_load(wave_file.read_text())
    spawns = yaml.safe_load(SPAWN_SOURCE.read_text())
    harness = HARNESS_SOURCE.read_text()
    machine = check_system_machine(pin)
    pairs = wave_pairs(config)
    check_masters(pairs)

    kinds = [str(k) for k in config["kinds"]]
    hands_per_kind = int(config["hands_per_kind"])
    model = str(config["model"])

    stamp = datetime.datetime.now(datetime.UTC).strftime("%Y%m%dT%H%M%SZ")
    wave_dir = root_dir / f"{stamp}-{wave_file.stem}"
    if wave_dir.exists():
        print(f"wave root already exists: {wave_dir}", file=sys.stderr)
        return 1
    cells_dir = wave_dir / "cells"
    cells_dir.mkdir(parents=True)
    (wave_dir / "operators").mkdir()

    manifest: dict[str, Any] = {
        "wave_config": str(wave_file),
        "model": model,
        "machine": machine,
        "hands": [],
    }
    digests: dict[str, dict[str, str]] = {}

    for brief in sorted({f"{sense}/{stack}" for sense, stack in pairs}):
        sense, stack = brief.split("/", 1)
        brief_file = CAMPAIGN / "briefs" / f"{brief}.md"
        material_dir = CAMPAIGN / "materials" / sense
        for source, name in ((brief_file, "brief"), (material_dir, "material")):
            if not source.is_file() and not source.is_dir():
                print(f"missing {name} for {brief}: {source}", file=sys.stderr)
                return 1
        for kind in kinds:
            for n in range(1, hands_per_kind + 1):
                hand = f"{brief.replace('/', '-')}-{kind}-{n}"
                cell = cells_dir / hand
                input_dir = cell / "input"
                work_dir = cell / "work"
                env_dir = cell / "env"
                for d in (input_dir, work_dir, env_dir / "tmp"):
                    d.mkdir(parents=True)
                copy_bytes(brief_file, input_dir / "brief.md")
                copy_tree(material_dir, input_dir / "material")
                stage_stack(cell, sense, stack, os.environ.get("PATH", ""))
                preamble = harness.replace("__WORK__", str(work_dir)).replace(
                    "__INPUT__", str(input_dir)
                )
                text = str(spawns[kind]).replace("{{INPUT}}", str(input_dir))
                prompt_file = wave_dir / f"{hand}.prompt"
                prompt_file.write_text(preamble + "\n" + text)
                digests.setdefault(brief, {})[hand] = tree_digest(input_dir)
                manifest["hands"].append(
                    {
                        "hand": hand,
                        "kind": kind,
                        "brief": brief,
                        "stack": stack,
                        "model": model,
                        "cell": str(cell),
                        "input": str(input_dir),
                        "work": str(work_dir),
                        "prompt": str(prompt_file),
                    }
                )

    for brief, hands in digests.items():
        unique = set(hands.values())
        if len(unique) != 1:
            print(f"inputs of {brief} differ across hands: {hands}", file=sys.stderr)
            return 1

    (wave_dir / "manifest.json").write_text(json.dumps(manifest, indent=2) + "\n")
    print(f"wave: {wave_dir}")
    print(
        f"machine: {machine['version']} {machine['sha256'][:12]} at {machine['path']}"
    )
    for hand in manifest["hands"]:
        print(f"  {hand['hand']}: {hand['cell']}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))
