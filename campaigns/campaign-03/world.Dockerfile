# world.Dockerfile — the th3 campaign layer (campaign-03).
#
# Built fresh for this campaign FROM THE BRIEFS' PINS ONLY: every
# stack and library below is named by a brief (briefs/<sense>/<stack>.md,
# "## Stack and tools — my requirement, not the executor's choice").
# Versions are pinned where a brief pins them; the rest resolve at
# build time and are recorded into /world-lock.txt (the lock file is
# part of the campaign's evidence — every cell of every hand runs the
# SAME image).
#
# Base: the project's world image (campaigns/world.Dockerfile — the
# five toolchains + linters + pytest/pytest-cov). This layer adds:
# the brief-pinned libraries, redis-server (the errand briefs require
# "integration tests against a running Redis"), a modern rust
# toolchain (the briefs' crates need it; the distro rustc is too
# old), and the offline caches.
#
# Build (from the repo root, network available at BUILD time only):
#   docker build -f campaigns/campaign-03/world.Dockerfile -t punchtape-world:th3 campaigns/campaign-03
# then copy the lock out:
#   docker run --rm punchtape-world:th3 cat /world-lock.txt > campaigns/campaign-03/world-lock.txt

FROM punchtape-world:latest

# --- system packages: C/C++ build + the brief-pinned C/C++ libraries,
# --- redis (errand), and the image toolchain GraphicsMagick++ (brim-cpp).
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      g++ make cmake pkg-config \
      redis-server \
      libgraphicsmagick++1-dev graphicsmagick \
      libhiredis-dev libyaml-dev libmd4c-dev \
 && rm -rf /var/lib/apt/lists/*

# --- vendored C/C++ headers (pressmark/quire/errand cpp briefs):
# --- inja + nlohmann/json are header-only, fetched at build time.
RUN mkdir -p /usr/local/include/inja /usr/local/include/nlohmann \
 && curl -fsSL https://raw.githubusercontent.com/pantor/inja/v3.5.0/single_include/inja/inja.hpp -o /usr/local/include/inja/inja.hpp \
 && curl -fsSL https://raw.githubusercontent.com/nlohmann/json/v3.11.3/single_include/nlohmann/json.hpp -o /usr/local/include/nlohmann/json.hpp

# --- python libraries (brim-python, errand-python, pressmark-python,
# --- quire-python). pressmark-python pins exact versions; the rest
# --- resolve at build and are locked.
RUN pip3 install --break-system-packages --no-cache-dir \
      Pillow numpy redis \
      PyYAML==6.0.1 Markdown==3.6 Jinja2==3.1.4 pydantic==2.8.2 python-slugify==8.0.4

# --- node libraries (brim-js, errand-js, pressmark-js, quire-js;
# --- quire-js pins exact versions). Installed into the world's
# --- library dir; /tmp/node_modules links to it so node's upward
# --- lookup from throwaway run directories resolves bare imports.
RUN mkdir -p /opt/th3-node_modules \
 && cd /opt/th3-node_modules \
 && npm install --no-audit --no-fund --no-save \
      jimp ioredis js-yaml nunjucks \
      marked@17.0.5 yeahjs@0.3.1 @mourner/yeahml@1.0.0 \
 && ln -sfn /opt/th3-node_modules /tmp/node_modules \
 && ln -sfn /opt/th3-node_modules /run/node_modules

# --- go module cache (errand-go, pressmark-go, quire-go briefs).
ENV GOMODCACHE=/opt/th3-gomodcache
RUN mkdir -p /opt/th3-gomods \
 && cd /opt/th3-gomods \
 && go mod init th3-cache \
 && go get github.com/redis/go-redis/v9@latest \
           github.com/google/uuid@latest \
           github.com/yuin/goldmark@latest \
           gopkg.in/yaml.v3@latest \
 && go mod download all \
 && rm -f go.mod go.sum

# --- rust: the briefs' crates (tokio, image, redis, uuid, serde*,
# --- comrak, tera, serde_yaml) need a current toolchain; rustup
# --- stable + clippy, then the crate cache for offline builds.
ENV RUSTUP_HOME=/opt/rustup CARGO_HOME=/opt/cargo
RUN curl -fsSL https://sh.rustup.rs | sh -s -- -y --profile minimal --component clippy \
 && . /opt/cargo/env \
 && mkdir -p /opt/th3-crates \
 && cd /opt/th3-crates \
 && cargo init --lib --name th3-cache \
 && cargo add tokio --features full \
 && cargo add image redis uuid --features uuid/v4 \
 && cargo add serde serde_json comrak tera serde_yaml \
 && cargo fetch

# --- offline modes: a task that needs the network fails loudly
# --- instead of quietly downloading.
ENV GOPROXY=off GOSUMDB=off GOFLAGS=-mod=mod \
    CARGO_NET_OFFLINE=true \
    PATH="/opt/cargo/bin:${PATH}"

# --- the lock: every resolved version, recorded at build time.
RUN { echo "== th3 world lock =="; date -u; \
     echo "-- pip"; pip3 list --format=freeze; \
     echo "-- npm (th3 libs)"; cd /opt/th3-node_modules && npm ls --depth=0 2>/dev/null; \
     echo "-- go modules"; cd /opt/th3-gomods && go list -m all 2>/dev/null; \
     echo "-- rust toolchain"; rustc --version; cargo --version; \
     echo "-- system"; g++ --version | head -1; node --version; python3 --version; redis-server --version; \
   } > /world-lock.txt

WORKDIR /run
