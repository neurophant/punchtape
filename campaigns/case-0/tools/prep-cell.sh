#!/bin/bash
# prep-cell.sh — PREPARE ONE CELL of the case-0 wave (the host
# world by the campaign's OWN mechanics and copy): stage
# the cell (stage-cell.sh), fill its env/ from the PLATFORM master
# caches (the shared-cache law: the cache area is METHODOLOGY §0.4
# substrate, warmed once at series prep; inside prep no network
# exists), write the driver-side environment exports and the exact
# tool versions (env/VERSIONS.txt; the 0.2.0
# world-lock.txt rides the campaign as the VERSIONS SOURCE, the live
# versions are written from --version at preparation), and prove
# readiness with env-check.sh.
#
# Usage: prep-cell.sh <run-root> <task> <hand> [--class pilot|smoke|main]
#   task = <sense>-<stack> (the same 25 tasks as the 0.2.0 bank)
#   hand = solo | top   (the owner's word: TWO hands — solo and
#          the machine hand on the top model; no third/flash hand
#          in this repo's reality)
# Env:  C0_ENV_PREP (default: the platform cache area);
#       C0_BINARY and C0_DRIVER_UUID pass through to stage-cell.sh.
set -euo pipefail
RR="${1:?run root}"; TASK="${2:?task: <sense>-<stack>}"; HAND="${3:?hand: solo|top}"; shift 3
CLASS=main
while [ $# -gt 0 ]; do case "$1" in --class) CLASS="$2"; shift 2;; *) echo "unknown: $1" >&2; exit 2;; esac; done
HERE="$(cd "$(dirname "$0")" && pwd)"
CAMP="$(cd "$HERE/.." && pwd)"
RR="$(readlink -f "$RR")"

SENSE="${TASK%-*}"; STACK="${TASK#*-}"
case "$SENSE" in brim|conduit|errand|pressmark|quire) ;; *) echo "prep-cell: unknown sense $SENSE" >&2; exit 2 ;; esac
case "$STACK" in cpp|go|js|python|rust) ;; *) echo "prep-cell: unknown stack $STACK" >&2; exit 2 ;; esac
case "$HAND" in solo|top) ;; *) echo "hand must be solo|top (two hands — no third/flash hand exists)" >&2; exit 2 ;; esac

PREP="${C0_ENV_PREP:-$HOME/.experiments/cache/env-prep}"
LOCALBIN="$HOME/.local/bin"; LOCALLIB="$HOME/.local/lib"

# --- 1. the cell itself (inputs, spawn archive, canary, manifest)
bash "$HERE/stage-cell.sh" "$SENSE" "$STACK" "$HAND" "$RR" --class "$CLASS"

read -r CELL CELLNAME < <(python3 - "$RR" "$SENSE" "$STACK" "$HAND" <<'PY'
import glob, json, sys
root, sense, stack, hand = sys.argv[1:5]
for m in sorted(glob.glob(f"{root}/manifests/*.json")):
    d = json.load(open(m))
    if d.get("sense") == sense and d.get("stack") == stack and d.get("hand") == hand:
        print(d["cell_path"], d["cell"])
        break
PY
)
[ -n "${CELL:-}" ] || { echo "prep-cell: no staged manifest for $TASK-$HAND" >&2; exit 2; }

# --- 2. the task environment: the platform master caches copied in
# --- (no network here — the substrate was warmed once at series prep)
mkdir -p "$CELL/env/deps"
EXPORTS=""
add_export() { EXPORTS="$EXPORTS
$1"; }
# every temp file of the hand's nested programs rides the cell's own
# disk-backed dir (nothing of a cell rides the tmpfs /tmp)
mkdir -p "$CELL/env/tmp"
add_export 'export TMPDIR="'"$CELL"'/env/tmp"'

case "$SENSE:$STACK" in
  conduit:rust)
  [ -d "$PREP/cargo-registry" ] || { echo "prep-cell: no master cargo cache at $PREP/cargo-registry" >&2; exit 2; }
  rm -rf "$CELL/env/deps/cargo-registry"; cp -r "$PREP/cargo-registry" "$CELL/env/deps/cargo-registry"
  add_export 'export CARGO_HOME="'"$CELL"'/env/deps/cargo-registry"'
  ;;
  pressmark:rust)
  [ -d "$PREP/cargo-pressmark" ] || { echo "prep-cell: no master cargo cache at $PREP/cargo-pressmark" >&2; exit 2; }
  rm -rf "$CELL/env/deps/cargo-registry"; cp -r "$PREP/cargo-pressmark" "$CELL/env/deps/cargo-registry"
  add_export 'export CARGO_HOME="'"$CELL"'/env/deps/cargo-registry"'
  ;;
  brim:rust)
  [ -d "$PREP/cargo-image" ] || { echo "prep-cell: no master cargo cache at $PREP/cargo-image" >&2; exit 2; }
  rm -rf "$CELL/env/deps/cargo-registry"; cp -r "$PREP/cargo-image" "$CELL/env/deps/cargo-registry"
  add_export 'export CARGO_HOME="'"$CELL"'/env/deps/cargo-registry"'
  ;;
  errand:rust)
  [ -d "$PREP/cargo-errand" ] || { echo "prep-cell: no master cargo cache at $PREP/cargo-errand" >&2; exit 2; }
  rm -rf "$CELL/env/deps/cargo-registry"; cp -r "$PREP/cargo-errand" "$CELL/env/deps/cargo-registry"
  add_export 'export CARGO_HOME="'"$CELL"'/env/deps/cargo-registry"'
  ;;
  quire:rust)
  [ -d "$PREP/cargo-quire" ] || { echo "prep-cell: no master cargo cache at $PREP/cargo-quire" >&2; exit 2; }
  rm -rf "$CELL/env/deps/cargo-registry"; cp -r "$PREP/cargo-quire" "$CELL/env/deps/cargo-registry"
  add_export 'export CARGO_HOME="'"$CELL"'/env/deps/cargo-registry"'
  ;;
  pressmark:js)
  [ -d "$PREP/node_modules" ] || { echo "prep-cell: no master module cache at $PREP/node_modules" >&2; exit 2; }
  rm -rf "$CELL/env/deps/node_modules"; cp -r "$PREP/node_modules" "$CELL/env/deps/node_modules"
  ln -sfn ../env/deps/node_modules "$CELL/work/node_modules"
  ;;
  brim:js)
  [ -d "$PREP/nm-brim/node_modules" ] || { echo "prep-cell: no master module cache at $PREP/nm-brim" >&2; exit 2; }
  rm -rf "$CELL/env/deps/node_modules"; cp -r "$PREP/nm-brim/node_modules" "$CELL/env/deps/node_modules"
  ln -sfn ../env/deps/node_modules "$CELL/work/node_modules"
  ;;
  errand:js)
  [ -d "$PREP/nm-errand/node_modules" ] || { echo "prep-cell: no master module cache at $PREP/nm-errand" >&2; exit 2; }
  rm -rf "$CELL/env/deps/node_modules"; cp -r "$PREP/nm-errand/node_modules" "$CELL/env/deps/node_modules"
  ln -sfn ../env/deps/node_modules "$CELL/work/node_modules"
  ;;
  quire:js)
  [ -d "$PREP/nm-quire/node_modules" ] || { echo "prep-cell: no master module cache at $PREP/nm-quire" >&2; exit 2; }
  rm -rf "$CELL/env/deps/node_modules"; cp -r "$PREP/nm-quire/node_modules" "$CELL/env/deps/node_modules"
  ln -sfn ../env/deps/node_modules "$CELL/work/node_modules"
  ;;
  pressmark:go|quire:go)
  MC="$PREP/_gomod-pressmark"; [ "$SENSE" = quire ] && MC="$PREP/_gomod-quire"
  [ -d "$MC" ] || { echo "prep-cell: no master go cache at $MC" >&2; exit 2; }
  rm -rf "$CELL/env/deps/gomod-cache"; cp -r "$MC" "$CELL/env/deps/gomod-cache"
  add_export 'export GOMODCACHE="'"$CELL"'/env/deps/gomod-cache"'
  add_export 'export GOPROXY=off'
  ;;
  errand:go)
  [ -d "$PREP/_gomod-errand" ] || { echo "prep-cell: no master go cache at $PREP/_gomod-errand" >&2; exit 2; }
  rm -rf "$CELL/env/deps/gomod-cache"; cp -r "$PREP/_gomod-errand" "$CELL/env/deps/gomod-cache"
  add_export 'export GOMODCACHE="'"$CELL"'/env/deps/gomod-cache"'
  add_export 'export GOPROXY=off'
  ;;
  pressmark:cpp|quire:cpp)
  [ -d "$PREP/include/inja" ] && [ -d "$PREP/include/nlohmann" ] \
    || { echo "prep-cell: no vendored include tree at $PREP/include" >&2; exit 2; }
  mkdir -p "$CELL/env/include"
  cp -r "$PREP/include/inja" "$PREP/include/nlohmann" "$CELL/env/include/"
  add_export 'export CPLUS_INCLUDE_PATH="'"$CELL"'/env/include:$HOME/.local/include${CPLUS_INCLUDE_PATH:+:$CPLUS_INCLUDE_PATH}"'
  add_export 'export LIBRARY_PATH="$HOME/.local/lib${LIBRARY_PATH:+:$LIBRARY_PATH}"'
  add_export 'export LD_LIBRARY_PATH="$HOME/.local/lib${LD_LIBRARY_PATH:+:$LD_LIBRARY_PATH}"'
  ;;
  brim:cpp)
  add_export 'export CPLUS_INCLUDE_PATH="$HOME/.local/include:$HOME/.local/include/GraphicsMagick${CPLUS_INCLUDE_PATH:+:$CPLUS_INCLUDE_PATH}"'
  add_export 'export C_INCLUDE_PATH="$HOME/.local/include:$HOME/.local/include/GraphicsMagick${C_INCLUDE_PATH:+:$C_INCLUDE_PATH}"'
  add_export 'export LIBRARY_PATH="$HOME/.local/lib${LIBRARY_PATH:+:$LIBRARY_PATH}"'
  add_export 'export LD_LIBRARY_PATH="$HOME/.local/lib${LD_LIBRARY_PATH:+:$LD_LIBRARY_PATH}"'
  ;;
  errand:cpp)
  add_export 'export CPLUS_INCLUDE_PATH="$HOME/.local/include${CPLUS_INCLUDE_PATH:+:$CPLUS_INCLUDE_PATH}"'
  add_export 'export LIBRARY_PATH="$HOME/.local/lib${LIBRARY_PATH:+:$LIBRARY_PATH}"'
  add_export 'export LD_LIBRARY_PATH="$HOME/.local/lib${LD_LIBRARY_PATH:+:$LD_LIBRARY_PATH}"'
  ;;
esac

# the go toolchain writes its build cache and GOPATH — both ride the
# CELL (the owner's cleanliness law: nothing outside
# ~/.experiments and the cell; a shared ~/.cache/go-build or ~/go
# would be a cross-hand leak and host litter)
case "$STACK" in
  go)
  mkdir -p "$CELL/env/deps/gocache" "$CELL/env/deps/gopath"
  add_export 'export GOCACHE="'"$CELL"'/env/deps/gocache"'
  add_export 'export GOPATH="'"$CELL"'/env/deps/gopath"'
  ;;
  rust)
  # llvm-cov/debuginfod of the host toolchain phones its servers and
  # drops scratch into the system /tmp unless switched off —
  # the cell's ENV silences it for every
  # llvm-tools invocation
  add_export 'export DEBUGINFOD_URLS=""'
  ;;
esac

# --- 2b. the cell's own PORT RANGE (the port-collision law:
# --- 8 parallel cells on one host collide on fixed ports; a cell
# --- needs its own disjoint window below the host's ephemeral range
# --- 32768–60999). The window is claimed mechanically in the run
# --- root: the first free 256-port step from the cell's hash — two
# --- cells of one wave can never share a window
mkdir -p "$RR/port-ranges"
PORT_BASE=$(( 20000 + 256 * ( $(printf '%s' "$CELL" | cksum | cut -d' ' -f1) % 50 ) ))
PORT_WALKS=0
while [ -e "$RR/port-ranges/$PORT_BASE" ]; do
  PORT_BASE=$(( PORT_BASE + 256 )); PORT_WALKS=$(( PORT_WALKS + 1 ))
  [ "$PORT_BASE" -ge 32768 ] && PORT_BASE=20000
  [ "$PORT_WALKS" -gt 50 ] && { echo "prep-cell: no free port window under $RR/port-ranges" >&2; exit 2; }
done
printf 'claimed by %s (%s-%s)\n' "$CELLNAME" "$PORT_BASE" "$(( PORT_BASE + 255 ))" > "$RR/port-ranges/$PORT_BASE"
PORT_RANGE="$PORT_BASE-$(( PORT_BASE + 255 ))"
add_export 'export C0_PORT_RANGE="'"$PORT_RANGE"'"'

# --- 3. the driver-side exports (spawn-time source file): env.sh
# --- archives them with the task dir; work/ENV is THE HAND'S OWN
# --- mechanical copy (the harness orders every
# --- command as: cd work && . ./ENV && sh -lc '<cmd>' — the session
# --- env of a subagent resets between its tool calls, so per-command
# --- sourcing is the only delivery that cannot be lost)
ENVS="$RR/tasks/$CELLNAME/env.sh"
{
  echo "# the cell's session environment (prep-cell, case-0):"
  echo "# source this in the hand's session before the first turn"
  echo "# the exact tool versions this cell was built with: env/VERSIONS.txt"
  echo "# (the 0.2.0 world lock rides the campaign as world-lock.txt —"
  echo "#  the VERSIONS SOURCE; the live versions below are the host truth)"
  if [ -n "$EXPORTS" ]; then printf '%s\n' "$EXPORTS"; else echo "# (this task needs no driver-side exports)"; fi
} > "$ENVS"
{
  echo "# ENV of this cell (prep-cell, case-0): source before every"
  echo "# command (the harness prefix: cd <work> && . ./ENV && sh -lc '...')."
  echo "# Everything a program of yours writes — scratch, caches, TMPDIR —"
  echo "# stays INSIDE the work directory; /tmp and every other host"
  echo "# directory are write-forbidden (the filesystem walls)."
  echo "# NETWORK PORTS: listen and connect ONLY on the ports of your"
  echo "# own \$C0_PORT_RANGE range (the window is allocated to this cell"
  echo "# and does not intersect the wave's other cells; do not take the"
  echo "# host's ephemeral range 32768–60999 — it is shared)."
  if [ -n "$EXPORTS" ]; then printf '%s\n' "$EXPORTS"; else echo "# (this task needs no exports beyond the walls)"; fi
} > "$CELL/work/ENV"

# --- 4. env/VERSIONS.txt (written from --version)
{
  echo "# the exact tool versions this cell was prepared with (one per line,"
  echo "# name + version string); identical for every hand of the task"
  case "$STACK" in
    cpp)  echo "g++ $(g++ --version | head -1)"
          echo "make $(make --version | head -1)" ;;
    go)   echo "go $(go version)" ;;
    js)   echo "node $(node --version)"
          echo "npm $(npm --version)" ;;
    python) echo "python3 $(python3 --version)" ;;
    rust) echo "cargo $(cargo --version)"
          echo "rustc $(rustc --version)"
          echo "clippy-driver $(clippy-driver --version 2>/dev/null || echo n/a)" ;;
  esac
  case "$SENSE:$STACK" in
    errand:*) echo "redis-server $(LD_LIBRARY_PATH=$LOCALLIB:$LD_LIBRARY_PATH "$LOCALBIN/redis-server" --version 2>/dev/null | head -1)" ;;
    brim:cpp) echo "graphicsmagick++ $(LD_LIBRARY_PATH=$LOCALLIB "$LOCALBIN/GraphicsMagick++-config" --version 2>/dev/null | head -1 || echo n/a)" ;;
    pressmark:cpp|quire:cpp) echo "libyaml $(pkg-config --modversion yaml-0-1 2>/dev/null || echo '0.2.5 (host install)')"
          echo "md4c 0.5.2 (host install)" ;;
    pressmark:python|quire:python) python3 -c "import yaml, markdown, jinja2; print('PyYAML', yaml.__version__); print('Markdown', markdown.__version__); print('Jinja2', jinja2.__version__)" 2>/dev/null || true ;;
    brim:python) python3 -c "import PIL, numpy; print('Pillow', PIL.__version__); print('numpy', numpy.__version__)" 2>/dev/null || true ;;
    errand:python) python3 -c "import redis; print('redis-py', redis.__version__)" 2>/dev/null || true ;;
  esac
} > "$CELL/env/VERSIONS.txt"

# --- 5. readiness: the toolchain answers, the cache proves itself
bash "$HERE/env-check.sh" "$TASK" "$CELL"

echo "PREPARED: $CELL ($TASK-$HAND, class $CLASS; exports: $ENVS)"
