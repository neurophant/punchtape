#!/usr/bin/env bash
# Install punchtape with one script: Linux and macOS.
# Requires Go 1.26+. The single install path is ~/.local/bin
# (one binary in one path; the build is reproducible, -buildvcs=false).
set -euo pipefail

if ! command -v go >/dev/null 2>&1; then
  echo "FAIL: go is not installed (need Go 1.26+): https://go.dev/dl/" >&2
  exit 1
fi

REPO="$(cd "$(dirname "$0")" && pwd)"
echo "building punchtape from $REPO ..."
cd "$REPO"
VER="$(git describe --tags --abbrev=0 2>/dev/null || echo 0.1.0)"
go build -buildvcs=false -ldflags "-X github.com/neurophant/punchtape/internal/cli.Version=$VER" -o punchtape ./cmd/punchtape

DEST="${HOME}/.local/bin"
mkdir -p "$DEST"
mv punchtape "$DEST/punchtape"
case ":$PATH:" in
  *":$DEST:"*) ;;
  *) echo "note: add $DEST to PATH: export PATH=\"\$PATH:$DEST\"" ;;
esac
echo "installed: $DEST/punchtape"
# a smoke test in a temporary directory: next opens an instance in
# cwd — never run it from the repository root (a ghost instance)
smoke="$(mktemp -d)"
( cd "$smoke" && "$DEST/punchtape" next >/dev/null 2>&1 ) || true
rm -rf "$smoke"
echo "done. Try: mkdir demo && cd demo && punchtape next"
