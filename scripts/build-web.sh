#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SRC="$ROOT/src"
WEB="$ROOT/web"
DOCS="$ROOT/docs"

mkdir -p "$DOCS"
find "$DOCS" -mindepth 1 ! -name '.gitkeep' -delete 2>/dev/null || true

cd "$SRC"
GOOS=js GOARCH=wasm go build -o "$DOCS/flappy.wasm" ./cmd

GOROOT="$(go env GOROOT)"
if [ -f "$GOROOT/lib/wasm/wasm_exec.js" ]; then
  cp "$GOROOT/lib/wasm/wasm_exec.js" "$DOCS/"
else
  cp "$GOROOT/misc/wasm/wasm_exec.js" "$DOCS/"
fi

cp "$WEB/index.html" "$DOCS/"
cp "$WEB/manifest.webmanifest" "$DOCS/"
cp "$WEB/sw.js" "$DOCS/"
cp -r "$WEB/icons" "$DOCS/"

echo "Web build complete: $DOCS"
