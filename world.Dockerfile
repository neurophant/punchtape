# world.Dockerfile — the punchtape-world image: the spartan
# environment for live field runs. The toolchains of the task bank
# stacks: Go, Python, JavaScript/TypeScript, Rust, C/C++. Tasks need
# no network: the builds are hermetic — every task dependency is
# pre-cached in the world image (cargo --offline, the module and
# node_modules caches, vendored headers), a task that needs the
# network is refused loudly.
#
# Campaign rule: every task world carries its stack's linters on PATH —
# the lint gate is claimed whenever the stack has one; without a
# linter in the environment, static strictness is lost silently.
#
# Build: docker build -f world.Dockerfile -t punchtape-world:latest .
FROM golang:1.26.8-bookworm

RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      python3 python3-pip nodejs npm rustc cargo \
 && pip3 install --break-system-packages --no-cache-dir pytest pytest-cov \
 && npm install -g --no-audit --no-fund typescript @types/node \
 && rm -rf /var/lib/apt/lists/*

# Stack linters on PATH: pyflakes (python), rust-clippy (rust),
# cppcheck (c/c++), eslint (js/ts), staticcheck (go). A separate
# layer, so the heavy base layer above stays cached across edits.
# rust-clippy installs as a pair to the apt rustc of the same version;
# a world with a rustup toolchain brings its own clippy component —
# the rule is the same.
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      rust-clippy cppcheck \
 && pip3 install --break-system-packages --no-cache-dir pyflakes \
 && npm install -g --no-audit --no-fund eslint \
 && rm -rf /var/lib/apt/lists/* \
 && go install honnef.co/go/tools/cmd/staticcheck@latest

WORKDIR /run

# Library resolution from throwaway run directories of a run
# (/tmp/check-*): node's upward lookup reaches /tmp/node_modules — a
# permanent symlink to the world's libraries (campaign rule: the task
# environment carries its own stack; the machine neither knows nor
# decides this).
#
# The per-campaign layers are built ON TOP of this base (see
# campaigns/campaign-02/world.Dockerfile, then world-js.Dockerfile for
# the JS resolution symlink); a fresh build of this file gives latest
# (the baseline without campaign libraries).
