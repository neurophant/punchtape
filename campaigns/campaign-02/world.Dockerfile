# world.Dockerfile — the campaign-02 world: punchtape-world plus the
# dependencies of all 25 S-layer cells (5 senses × 5 stacks). Built
# with network; task worlds need none — every dependency is baked into
# the image (the campaign rule).
# Build: docker build -f campaigns/campaign-02/world.Dockerfile \
#                    -t punchtape-world:c02 /tmp/c02ctx
FROM punchtape-world:latest

# 1. System packages: the C/C++ build chain (the base has only the
#    cppcheck linter, no compiler), redis (task daemons live in the
#    world on their own ports), the C++ task libraries (GraphicsMagick++,
#    hiredis, libyaml, md4c).
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      g++ make cmake pkg-config \
      redis-server \
      libgraphicsmagick++1-dev libhiredis-dev libyaml-dev libmd4c-dev \
 && rm -rf /var/lib/apt/lists/*

# 2. Node.js 22 (the distro node 18 is too old for marked 17) + global
#    tsc/eslint under it. The js task dependencies live in
#    /run/node_modules: any module under /run/task/** resolves them by
#    walking up the tree.
RUN curl -fsSL https://nodejs.org/dist/v22.14.0/node-v22.14.0-linux-x64.tar.gz \
      | tar -xz -C /opt \
 && mv /opt/node-v22.14.0-linux-x64 /opt/node \
 && /opt/node/bin/npm install -g --prefix /opt/node --no-audit --no-fund typescript @types/node eslint \
 && cd /run \
 && /opt/node/bin/npm install --no-audit --no-fund \
      jimp ioredis js-yaml marked@17.0.5 nunjucks yeahjs@0.3.1 @mourner/yeahml@1.0.0

# 3. Rust on a fresh toolchain (the distro rustc 1.63 is too old for
#    modern tokio/comrak/tera): rustup + clippy. The crate registry is
#    cached right here.
ENV CARGO_HOME=/opt/rust/cargo RUSTUP_HOME=/opt/rust/rustup
RUN curl -fsSL https://static.rust-lang.org/rustup/dist/x86_64-unknown-linux-gnu/rustup-init \
      -o /tmp/rustup-init && chmod +x /tmp/rustup-init \
 && /tmp/rustup-init -y --no-modify-path --profile minimal \
      --default-toolchain stable -c clippy \
 && rm /tmp/rustup-init

# The paths take effect from here: rustup-cargo and node22 come before
# the distro toolchains (otherwise the cargo fetch below runs on the
# distro cargo 1.65).
ENV PATH=/opt/node/bin:/opt/rust/cargo/bin:/usr/local/go/bin:/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

# 4. The python task dependencies; the brief-pinned versions exact,
#    the rest current (the staging pins them — see README.md).
RUN pip3 install --break-system-packages --no-cache-dir \
      Pillow numpy redis \
      PyYAML==6.0.1 Markdown==3.6 Jinja2==3.1.4 pydantic==2.8.2 \
      python-slugify==8.0.4

# 5. C++ headers without a distro package: inja 3.5.0 (single header,
#    needs nlohmann/json) — into /usr/include.
RUN mkdir -p /usr/include/inja /usr/include/nlohmann \
 && curl -fsSL -o /usr/include/inja/inja.hpp \
      https://github.com/pantor/inja/releases/latest/download/inja.hpp \
 && curl -fsSL -o /usr/include/nlohmann/json.hpp \
      https://github.com/nlohmann/json/releases/download/v3.11.3/json.hpp

# 6. The Go module cache of every go task (goldmark, yaml.v3,
#    go-redis, uuid).
RUN mkdir -p /tmp/gofetch && cd /tmp/gofetch \
 && go mod init c02fetch \
 && go get github.com/yuin/goldmark@latest gopkg.in/yaml.v3@latest \
      github.com/redis/go-redis/v9@latest github.com/google/uuid@latest \
 && go mod download all \
 && rm -rf /tmp/gofetch

# 7. The crate registry cache of every rust task: tokio (conduit),
#    image (brim), the pinned errand set (redis 0.8 / uuid 0.5 /
#    serde 1.0), the pressmark/quire site set (serde_yaml/comrak/tera).
RUN for d in tokio image errand site; do mkdir -p /tmp/cargo-$d/src; \
      printf '' > /tmp/cargo-$d/src/main.rs; done \
 && printf '[package]\nname="t"\nversion="0.0.0"\nedition="2021"\n[dependencies]\ntokio = { version = "1", features = ["full"] }\n' \
      > /tmp/cargo-tokio/Cargo.toml \
 && printf '[package]\nname="t"\nversion="0.0.0"\nedition="2021"\n[dependencies]\nimage = "*"\n' \
      > /tmp/cargo-image/Cargo.toml \
 && printf '[package]\nname="t"\nversion="0.0.0"\nedition="2021"\n[dependencies]\nredis = "0.8"\nuuid = { version = "0.5", features = ["v4"] }\nserde = "1.0"\nserde_json = "1.0"\nserde_derive = "1.0"\n' \
      > /tmp/cargo-errand/Cargo.toml \
 && printf '[package]\nname="t"\nversion="0.0.0"\nedition="2021"\n[dependencies]\nserde_yaml = "*"\ncomrak = "*"\ntera = "*"\nserde = "1.0"\nserde_json = "1.0"\n' \
      > /tmp/cargo-site/Cargo.toml \
 && cd /tmp/cargo-tokio && cargo fetch \
 && cd /tmp/cargo-image && cargo fetch \
 && cd /tmp/cargo-errand && cargo fetch \
 && cd /tmp/cargo-site && cargo fetch \
 && rm -rf /tmp/cargo-*

# 8. Offline modes (after the prefetches): the env applies both to
#    docker exec and to login shells (profile.d). Builds are forbidden
#    the network: the caches are complete.
ENV GOPROXY=file:///go/pkg/mod/cache/download GOSUMDB=off GOTOOLCHAIN=local GOFLAGS=-mod=mod \
    CARGO_NET_OFFLINE=true
RUN printf 'export PATH=/opt/node/bin:/opt/rust/cargo/bin:/usr/local/go/bin:/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\nexport GOPROXY=file:///go/pkg/mod/cache/download GOSUMDB=off GOTOOLCHAIN=local GOFLAGS=-mod=mod\nexport CARGO_NET_OFFLINE=true\n' \
      > /etc/profile.d/c02.sh \
 && chmod +x /etc/profile.d/c02.sh

WORKDIR /run/task
