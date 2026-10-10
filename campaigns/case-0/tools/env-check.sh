#!/bin/bash
# env-check.sh — the task environment READINESS CHECK: every
# toolchain of the task
# answers a version, every declared dependency imports or links, and
# the offline caches prove themselves. Host-only facts when no cell
# is given; with a cell (prep-cell calls this after filling env/)
# the cell's caches are proven too.
#
# Usage: env-check.sh <task> [<cell>]
#   task = <sense>-<stack> of the case-0 wave
#   cell = the staged cell root (its env/ must carry the caches)
#
# Exits 0 green (a one-line summary per check), non-zero red with
# the failing check named. No writes outside the cell's env/ (the
# offline builds ride scratch dirs inside env/).
set -euo pipefail

TASK="${1:?task: <sense>-<stack> of the s band}"
CELL="${2:-}"

fail() { echo "ENV-CHECK RED: $*" >&2; exit 1; }
say()  { printf 'env-check: %s\n' "$*"; }

SENSE="${TASK%-*}"; STACK="${TASK#*-}"

# --- the shared host facts: toolchains answer versions
GXX_V="$(g++ --version 2>/dev/null | head -1 || true)"
GO_V="$(go version 2>/dev/null || true)"
NODE_V="$(node --version 2>/dev/null || true)"
PY_V="$(python3 --version 2>/dev/null || true)"
CARGO_V="$(cargo --version 2>/dev/null || true)"
RUSTC_V="$(rustc --version 2>/dev/null || true)"
REDIS_V="$(LD_LIBRARY_PATH=$HOME/.local/lib:${LD_LIBRARY_PATH:-} ~/.local/bin/redis-server --version 2>/dev/null || true)"
REPO="$(cd "$(dirname "$0")/../.." && pwd)/.."
# the compile probes ride the disk-backed scratch root (never the
# tmpfs /tmp — the owner's word)
# the compile-probe scratch rides the SANCTIONED tmp/ zone of the
# experiments area (the structure law: runs/, tmp/, cache/ only
# at its root — nothing else may ever appear there)
SCR="${C0_SCRATCH_ROOT:-$HOME/.experiments}"
mkdir -p "$SCR/tmp/envcheck"

need_gxx()   { [ -n "$GXX_V" ]   || fail "g++ does not answer a version on PATH"; }
need_go()    { [ -n "$GO_V" ]    || fail "go does not answer a version on PATH"; }
need_node()  { [ -n "$NODE_V" ]  || fail "node does not answer a version on PATH"; }
need_python(){ [ -n "$PY_V" ]    || fail "python3 does not answer a version on PATH"; }
need_rust()  { [ -n "$CARGO_V" ] || fail "cargo does not answer a version on PATH"
                [ -n "$RUSTC_V" ] || fail "rustc does not answer a version on PATH"; }
need_redis() { [ -n "$REDIS_V" ] || fail "redis-server does not answer a version (expected at ~/.local/bin - the host install)"; }

# the user-local system libraries (this host's install of the series:
# GraphicsMagick++ for brim-cpp, hiredis for errand-cpp, libyaml +
# md4c for pressmark/quire-cpp; headers at ~/.local/include, libs at
# ~/.local/lib - reached through the session's include/library paths)
cpp_syslibs() {
  local inc="$HOME/.local/include" lib="$HOME/.local/lib"
  [ -f "$inc/yaml.h" ] || fail "libyaml headers missing at $inc (the host install)"
  [ -f "$inc/md4c-html.h" ] || fail "md4c headers missing at $inc"
  [ -e "$lib/libyaml.so" ] || fail "libyaml missing at $lib"
  [ -e "$lib/libmd4c-html.so" ] || fail "libmd4c-html missing at $lib"
  printf '#include <yaml.h>\n#include <md4c-html.h>\nint main(){ yaml_parser_t p; yaml_parser_initialize(&p); yaml_parser_delete(&p); md_html(0,0,0,0,0,0); return 0; }\n' > "$SCR/tmp/envcheck"/envcheck-cpp-$$.cpp
  if ! ( CPLUS_INCLUDE_PATH=$inc LIBRARY_PATH=$lib LD_LIBRARY_PATH=$lib g++ -std=c++17 "$SCR/tmp/envcheck"/envcheck-cpp-$$.cpp -lyaml -lmd4c-html -o "$SCR/tmp/envcheck"/envcheck-cpp-$$ \
         && LD_LIBRARY_PATH=$lib "$SCR/tmp/envcheck"/envcheck-cpp-$$ ); then
    rm -f "$SCR/tmp/envcheck"/envcheck-cpp-$$.cpp "$SCR/tmp/envcheck"/envcheck-cpp-$$
    fail "libyaml/md4c do not compile+link through $inc/$lib"
  fi
  rm -f "$SCR/tmp/envcheck"/envcheck-cpp-$$.cpp "$SCR/tmp/envcheck"/envcheck-cpp-$$
}
cpp_gm() {
  local inc="$HOME/.local/include" gmi="$HOME/.local/include/GraphicsMagick" lib="$HOME/.local/lib"
  [ -f "$gmi/Magick++.h" ] || fail "GraphicsMagick++ headers missing at $gmi"
  [ -e "$lib/libGraphicsMagick++.so" ] || fail "GraphicsMagick++ missing at $lib"
  echo '#include <Magick++.h>
int main(int argc, char** argv){ Magick::InitializeMagick(argv[0]); Magick::Image i; i.size("2x2"); i.read("xc:white"); return argc>1?1:0; }' > "$SCR/tmp/envcheck"/envcheck-gm-$$.cpp
  if ! ( CPLUS_INCLUDE_PATH=$inc:$gmi LIBRARY_PATH=$lib LD_LIBRARY_PATH=$lib \
         g++ -std=c++17 "$SCR/tmp/envcheck"/envcheck-gm-$$.cpp -lGraphicsMagick++ -lGraphicsMagick -o "$SCR/tmp/envcheck"/envcheck-gm-$$ \
         && LD_LIBRARY_PATH=$lib "$SCR/tmp/envcheck"/envcheck-gm-$$ ); then
    rm -f "$SCR/tmp/envcheck"/envcheck-gm-$$.cpp "$SCR/tmp/envcheck"/envcheck-gm-$$
    fail "GraphicsMagick++ does not compile+link through $gmi/$lib"
  fi
  rm -f "$SCR/tmp/envcheck"/envcheck-gm-$$.cpp "$SCR/tmp/envcheck"/envcheck-gm-$$
}
cpp_hiredis() {
  local inc="$HOME/.local/include" lib="$HOME/.local/lib"
  [ -f "$inc/hiredis/hiredis.h" ] || fail "hiredis headers missing at $inc/hiredis"
  [ -e "$lib/libhiredis.so" ] || fail "hiredis missing at $lib"
  echo '#include <hiredis/hiredis.h>
int main(){ redisContext* c = redisConnect("127.0.0.1", 1); if (c) redisFree(c); return 0; }' > "$SCR/tmp/envcheck"/envcheck-hr-$$.cpp
  if ! ( CPLUS_INCLUDE_PATH=$inc LIBRARY_PATH=$lib LD_LIBRARY_PATH=$lib g++ -std=c++17 "$SCR/tmp/envcheck"/envcheck-hr-$$.cpp -lhiredis -o "$SCR/tmp/envcheck"/envcheck-hr-$$ \
         && LD_LIBRARY_PATH=$lib "$SCR/tmp/envcheck"/envcheck-hr-$$ ); then
    rm -f "$SCR/tmp/envcheck"/envcheck-hr-$$.cpp "$SCR/tmp/envcheck"/envcheck-hr-$$
    fail "hiredis does not compile+link through $inc/$lib"
  fi
  rm -f "$SCR/tmp/envcheck"/envcheck-hr-$$.cpp "$SCR/tmp/envcheck"/envcheck-hr-$$
}
cpp_vendored() { # inja + nlohmann under the cell's env/include
  local inc="$CELL/env/include"
  [ -n "$CELL" ] || inc="$HOME/.experiments/cache/env-prep/include"
  [ -f "$inc/inja/inja.hpp" ] || fail "vendored inja missing under $inc (prep-cell stages env/include)"
  [ -f "$inc/nlohmann/json.hpp" ] || fail "vendored nlohmann missing under $inc"
  echo '#include <inja/inja.hpp>
#include <nlohmann/json.hpp>
int main(){ nlohmann::json d; d["x"]=1; return inja::render("{{ x }}", d)=="1"?0:1; }' > "$SCR/tmp/envcheck"/envcheck-inja-$$.cpp
  if ! ( CPLUS_INCLUDE_PATH=$inc g++ -std=c++17 "$SCR/tmp/envcheck"/envcheck-inja-$$.cpp -o "$SCR/tmp/envcheck"/envcheck-inja-$$ && "$SCR/tmp/envcheck"/envcheck-inja-$$ ); then
    rm -f "$SCR/tmp/envcheck"/envcheck-inja-$$.cpp "$SCR/tmp/envcheck"/envcheck-inja-$$
    fail "inja/nlohmann do not compile through $inc"
  fi
  rm -f "$SCR/tmp/envcheck"/envcheck-inja-$$.cpp "$SCR/tmp/envcheck"/envcheck-inja-$$
}
py_mods() { # module... — import through the system python3
  for m in "$@"; do
    python3 -c "import importlib; importlib.import_module('$m')" \
      || fail "python module '$m' does not import with the system python3"
  done
}

go_cache() { # cache-dir module... — the extracted modules present
  local cache="$1"; shift
  [ -d "$cache" ] || fail "no go module cache at $cache (warm it once)"
  for m in "$@"; do
    local hit
    hit="$(find "$cache" -type d -name "$(basename "$m")@*" -print -quit 2>/dev/null)"
    [ -n "$hit" ] || fail "go module $m not present in the cache at $cache"
  done
}

cargo_cache() { # cache-dir crate-src — offline scratch build
  local cache="$1" dep="$2"
  [ -d "$cache" ] || fail "no cargo cache at $cache (warm it once)"
  if [ -n "$CELL" ]; then
    local scratch="$CELL/env/.env-check-build"
    rm -rf "$scratch"; mkdir -p "$scratch/src"
    printf 'fn main() {}\n' > "$scratch/src/main.rs"
    printf '[package]\nname = "envcheck"\nversion = "0.1.0"\nedition = "2021"\n\n[dependencies]\n%s\n' "$dep" > "$scratch/Cargo.toml"
    if ( cd "$scratch" && CARGO_HOME="$cache" cargo build --release --offline >/dev/null 2>&1 ); then
      say "offline build: cargo build --release --offline against the cell cache — GREEN"
    else
      rm -rf "$scratch"; fail "cargo build --release --offline failed against $cache"
    fi
    rm -rf "$scratch"
  fi
}

case "$SENSE:$STACK" in
# ---- conduit: rust(tokio)/go(std)/js(std)/python(std)/cpp(posix)
conduit:rust) need_rust
  [ -n "$CELL" ] && cargo_cache "$CELL/env/deps/cargo-registry" 'tokio = { version = "1", features = ["full"] }'
  say "toolchain: $CARGO_V; $RUSTC_V" ;;
conduit:go) need_go; say "toolchain: $GO_V (stdlib only)" ;;
conduit:js) need_node; say "toolchain: node $NODE_V (built-ins only)" ;;
conduit:python) need_python; say "toolchain: $PY_V (stdlib only)" ;;
conduit:cpp) need_gxx; say "toolchain: $GXX_V (POSIX headers only)" ;;
# ---- pressmark: js(js-yaml,nunjucks,marked)/go(goldmark,yaml.v3)/
#      python(PyYAML,Markdown,Jinja2)/rust(serde_yaml,comrak,tera)/
#      cpp(libyaml,md4c,inja,nlohmann)
pressmark:js) need_node
  if [ -n "$CELL" ]; then
    MODS="$CELL/env/deps/node_modules"
    [ -d "$MODS" ] || fail "no module cache at $MODS"
    ( cd "$CELL/env/deps" && node -e "require('js-yaml'); require('nunjucks')" ) || fail "CJS require of js-yaml/nunjucks fails"
    ( cd "$CELL/env/deps" && node --input-type=module -e "import { marked } from 'marked'; if (typeof marked !== 'function') process.exit(1)" ) || fail "ESM named import { marked } fails"
    [ -L "$CELL/work/node_modules" ] || fail "work/node_modules is not a link (the dependency mechanism)"
  fi
  say "toolchain: node $NODE_V; modules: js-yaml + nunjucks (CJS) and { marked } (ESM named)" ;;
pressmark:go) need_go
  GO_CACHE_DIR="$HOME/.experiments/cache/env-prep/_gomod-pressmark"; [ -n "$CELL" ] && GO_CACHE_DIR="$CELL/env/deps/gomod-cache"
  go_cache "$GO_CACHE_DIR" github.com/yuin/goldmark gopkg.in/yaml.v3
  say "toolchain: $GO_V; modules: goldmark + yaml.v3 offline" ;;
pressmark:python) need_python; py_mods yaml markdown jinja2
  say "toolchain: $PY_V; modules: PyYAML + Markdown + Jinja2 import" ;;
pressmark:rust) need_rust
  [ -n "$CELL" ] && cargo_cache "$CELL/env/deps/cargo-registry" 'serde_yaml = "0.9"
comrak = "0.24"
tera = "1"'
  say "toolchain: $CARGO_V; $RUSTC_V" ;;
pressmark:cpp) need_gxx; cpp_syslibs; cpp_vendored
  say "toolchain: $GXX_V; libs: libyaml + md4c link; vendored inja + nlohmann compile" ;;
# ---- brim: rust(image)/go(std)/js(jimp)/python(Pillow,numpy)/cpp(GraphicsMagick++)
brim:rust) need_rust
  [ -n "$CELL" ] && cargo_cache "$CELL/env/deps/cargo-registry" 'image = "0.25"'
  say "toolchain: $CARGO_V; $RUSTC_V" ;;
brim:go) need_go; say "toolchain: $GO_V (stdlib JPEG decode)" ;;
brim:js) need_node
  if [ -n "$CELL" ]; then
    MODS="$CELL/env/deps/node_modules"
    [ -d "$MODS" ] || fail "no module cache at $MODS"
    ( cd "$CELL/env/deps" && node -e "const {Jimp} = require('jimp'); if (!Jimp) process.exit(1)" ) || fail "require of jimp fails"
    [ -L "$CELL/work/node_modules" ] || fail "work/node_modules is not a link"
  fi
  say "toolchain: node $NODE_V; module: jimp" ;;
brim:python) need_python; py_mods PIL numpy
  say "toolchain: $PY_V; modules: Pillow + numpy import" ;;
brim:cpp) need_gxx; cpp_gm
  say "toolchain: $GXX_V; GraphicsMagick++ compiles, links and initializes" ;;
# ---- errand: rust(redis,uuid,serde)/go(go-redis,uuid)/js(ioredis)/
#      python(redis)/cpp(hiredis); redis-server for all five
errand:rust) need_rust; need_redis
  [ -n "$CELL" ] && cargo_cache "$CELL/env/deps/cargo-registry" 'redis = { version = "0.27", default-features = false, features = ["script"] }
uuid = { version = "1", features = ["v4"] }'
  say "toolchain: $CARGO_V; $RUSTC_V; $REDIS_V" ;;
errand:go) need_go; need_redis
  GO_CACHE_DIR="$HOME/.experiments/cache/env-prep/_gomod-errand"; [ -n "$CELL" ] && GO_CACHE_DIR="$CELL/env/deps/gomod-cache"
  go_cache "$GO_CACHE_DIR" github.com/redis/go-redis/v9 github.com/google/uuid
  say "toolchain: $GO_V; modules: go-redis + uuid offline; $REDIS_V" ;;
errand:js) need_node; need_redis
  if [ -n "$CELL" ]; then
    MODS="$CELL/env/deps/node_modules"
    [ -d "$MODS" ] || fail "no module cache at $MODS"
    ( cd "$CELL/env/deps" && node -e "const r = require('ioredis'); if (!r) process.exit(1)" ) || fail "require of ioredis fails"
    [ -L "$CELL/work/node_modules" ] || fail "work/node_modules is not a link"
  fi
  say "toolchain: node $NODE_V; module: ioredis; $REDIS_V" ;;
errand:python) need_python; need_redis; py_mods redis
  say "toolchain: $PY_V; module: redis-py imports; $REDIS_V" ;;
errand:cpp) need_gxx; need_redis; cpp_hiredis
  say "toolchain: $GXX_V; hiredis compiles and links; $REDIS_V" ;;
# ---- quire: go(goldmark,yaml.v3)/python(markdown,Jinja2,PyYAML)/
#      js(marked,yeahjs,yeahml)/rust(comrak,tera,serde_yaml)/cpp(libyaml,md4c,inja)
quire:go) need_go
  GO_CACHE_DIR="$HOME/.experiments/cache/env-prep/_gomod-quire"; [ -n "$CELL" ] && GO_CACHE_DIR="$CELL/env/deps/gomod-cache"
  go_cache "$GO_CACHE_DIR" github.com/yuin/goldmark gopkg.in/yaml.v3
  say "toolchain: $GO_V; modules: goldmark + yaml.v3 offline" ;;
quire:python) need_python; py_mods markdown jinja2 yaml
  say "toolchain: $PY_V; modules: Markdown + Jinja2 + PyYAML import" ;;
quire:js) need_node
  if [ -n "$CELL" ]; then
    MODS="$CELL/env/deps/node_modules"
    [ -d "$MODS" ] || fail "no module cache at $MODS"
    ( cd "$CELL/env/deps" && node -e "require('yeahjs'); require('@mourner/yeahml'); const {marked} = require('marked'); if (typeof marked !== 'function') process.exit(1)" ) || fail "require of yeahjs/@mourner/yeahml/marked fails"
    [ -L "$CELL/work/node_modules" ] || fail "work/node_modules is not a link"
  fi
  say "toolchain: node $NODE_V; modules: marked + yeahjs + @mourner/yeahml" ;;
quire:rust) need_rust
  [ -n "$CELL" ] && cargo_cache "$CELL/env/deps/cargo-registry" 'comrak = "0.24"
tera = "1"
serde_yaml = "0.9"'
  say "toolchain: $CARGO_V; $RUSTC_V" ;;
quire:cpp) need_gxx; cpp_syslibs; cpp_vendored
  say "toolchain: $GXX_V; libs: libyaml + md4c link; vendored inja + nlohmann compile" ;;
*)
  fail "unknown task: $TASK (a <sense>-<stack> of the s band)" ;;
esac
say "ENV-CHECK GREEN ($TASK${CELL:+, cell caches proven})"
