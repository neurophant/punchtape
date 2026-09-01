#!/usr/bin/env bash
# get-punchtape.sh — install punchtape with one command:
#
#   curl -fsSL <base-URL>/get-punchtape.sh | bash
#
# The script detects the OS and architecture, downloads the binary and
# its checksum, verifies the sum, and puts the `punchtape` command into
# the SINGLE path ~/.local/bin (one binary in one path), adding it to
# the PATH of the current session and of the shell profile if needed.
#
# Environment variables:
#   PUNCHTAPE_DIST_BASE   — where the binaries live (default: the
#                           project's stable distribution address)
#   PUNCHTAPE_NO_VERIFY=1 — skip the checksum (not recommended)
set -euo pipefail

PUNCHTAPE_DIST_BASE="${PUNCHTAPE_DIST_BASE:-https://github.com/neurophant/punchtape/releases/latest/download}"

fail() { echo "FAIL: $*" >&2; exit 1; }

need() { command -v "$1" >/dev/null 2>&1 || fail "$2"; }
need curl "curl not found — install curl and retry"
need uname "uname not found — unknown system"

os="$(uname -s)"; arch="$(uname -m)"
case "$os" in
  Linux)  os=linux ;;
  Darwin) os=darwin ;;
  *) fail "unsupported OS: $os (linux and macOS today)" ;;
esac
case "$arch" in
  x86_64|amd64) arch=amd64 ;;
  aarch64|arm64) arch=arm64 ;;
  *) fail "unsupported architecture: $arch" ;;
esac

name="punchtape-$os-$arch"
bin_url="$PUNCHTAPE_DIST_BASE/$name"
sum_url="$PUNCHTAPE_DIST_BASE/$name.sha256"

echo "== punchtape: $os/$arch, source $PUNCHTAPE_DIST_BASE"

tmp="$(mktemp -d)"; trap 'rm -rf "$tmp"' EXIT
curl -fsSL "$bin_url" -o "$tmp/$name" || fail "could not download $bin_url"

if [ "${PUNCHTAPE_NO_VERIFY:-0}" != "1" ]; then
  curl -fsSL "$sum_url" -o "$tmp/$name.sha256" || fail "could not download the checksum $sum_url"
  if command -v sha256sum >/dev/null 2>&1; then
    (cd "$tmp" && sha256sum -c "$name.sha256") || fail "checksum mismatch"
  elif command -v shasum >/dev/null 2>&1; then
    (cd "$tmp" && shasum -a 256 -c "$name.sha256") || fail "checksum mismatch"
  else
    fail "neither sha256sum nor shasum is available — verification impossible"
  fi
  echo "== checksum verified"
fi
chmod +x "$tmp/$name"

# The single install path — the user's ~/.local/bin: no sudo, a
# standard directory, the same one for every installer of the project.
dst="$HOME/.local/bin"
mkdir -p "$dst" 2>/dev/null || fail "cannot create $dst"
cp "$tmp/$name" "$dst/punchtape" || fail "cannot write $dst/punchtape"

case ":$PATH:" in
  *":$dst:"*) : ;;
  *)
    export PATH="$dst:$PATH"
    line="export PATH=\"\$PATH:$dst\""
    added=0
    for rc in "$HOME/.bashrc" "$HOME/.profile" "$HOME/.zshrc"; do
      [ -f "$rc" ] || continue
      grep -qF "$line" "$rc" 2>/dev/null && { added=1; break; }
    done
    if [ "$added" = 0 ]; then
      for rc in "$HOME/.bashrc" "$HOME/.profile"; do
        [ -f "$rc" ] || continue
        printf '\n# punchtape installer\n%s\n' "$line" >> "$rc"
        added=1; break
      done
    fi
    # no rc file exists — create a minimal ~/.profile
    if [ "$added" = 0 ]; then
      printf '# punchtape installer\n%s\n' "$line" >> "$HOME/.profile"
    fi
    echo "== $dst added to PATH (restart the terminal or open a new one)"
    ;;
esac

echo "== installed: $dst/punchtape"
"$dst/punchtape" || true
echo "== done: the punchtape command is available. Check: punchtape"
