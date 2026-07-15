#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DOCS="$ROOT/docs"
PORT=8080

bash "$ROOT/scripts/build-web.sh"

echo "Serving $DOCS at http://localhost:$PORT"
echo "Press Ctrl+C to stop."

cd "$DOCS"
python3 -m http.server "$PORT"
