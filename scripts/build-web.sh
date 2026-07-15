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
VENDOR_DIR="$WEB/vendor"
VENDOR_FILE="$VENDOR_DIR/sqlitecloud-drivers.mjs"
if [ ! -f "$VENDOR_FILE" ]; then
  mkdir -p "$VENDOR_DIR"
  curl -fsSL "https://cdn.jsdelivr.net/npm/@sqlitecloud/drivers/+esm" -o "$VENDOR_FILE"
fi
if [ -d "$WEB/vendor" ]; then
  cp -r "$WEB/vendor" "$DOCS/"
fi
if [ -f "$WEB/config.js" ]; then
  cp "$WEB/config.js" "$DOCS/"
else
  echo 'window.FLAPPY_SQLITECLOUD_URL = "";' > "$DOCS/config.js"
fi

BUILD_ID="$(date -u +%Y%m%d%H%M%S)"
if git -C "$ROOT" rev-parse --short HEAD >/dev/null 2>&1; then
  BUILD_ID="$(git -C "$ROOT" rev-parse --short HEAD)"
fi
sed "s/__BUILD_ID__/$BUILD_ID/g" "$WEB/sw.js" > "$DOCS/sw.js"

cp -r "$WEB/icons" "$DOCS/"

echo "Web build complete: $DOCS (cache: flappy-beer-$BUILD_ID)"
