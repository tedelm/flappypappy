#!/usr/bin/env bash
# Copy a prebuilt web bundle (docs/) into /var/www/flappy on this host.
# Build WASM on a machine with enough RAM first:
#   bash scripts/build-web.sh
# Then on the container (as root):
#   sudo bash scripts/deploy-web.sh
#   sudo WEB_ROOT=/path/to/docs bash scripts/deploy-web.sh
set -euo pipefail

if [[ "${EUID}" -ne 0 ]]; then
  echo "run as root: sudo bash scripts/deploy-web.sh" >&2
  exit 1
fi

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
WEB_ROOT="${WEB_ROOT:-$ROOT/docs}"
DEST="${DEST:-/var/www/flappy}"
ENV_FILE="${ENV_FILE:-/etc/flappy/scoreapi.env}"
SCORE_API_HOST="${SCORE_API_HOST:-flappy-pappy.se}"

if [[ ! -f "$WEB_ROOT/index.html" ]]; then
  echo "no web build at $WEB_ROOT (expected index.html). Build with scripts/build-web.sh first." >&2
  exit 1
fi

install -d -o root -g root -m 0755 "$DEST"
rsync -a --delete \
  --exclude '.gitkeep' \
  "$WEB_ROOT"/ "$DEST"/

API_KEY=""
if [[ -f "$ENV_FILE" ]]; then
  # shellcheck disable=SC1090
  set -a
  # shellcheck source=/dev/null
  source "$ENV_FILE"
  set +a
fi

PUBLIC_URL="https://${SCORE_API_HOST}"
export PUBLIC_URL
export API_KEY="${API_KEY:-}"
export DEST
python3 - <<'PY'
import json, os
url = json.dumps(os.environ.get("PUBLIC_URL", ""))
key = json.dumps(os.environ.get("API_KEY", ""))
path = os.path.join(os.environ["DEST"], "config.js")
with open(path, "w", encoding="utf-8") as f:
    f.write(f"window.FLAPPY_SCORE_API_URL = {url};\n")
    f.write(f"window.FLAPPY_SCORE_API_KEY = {key};\n")
print(f"wrote {path}")
PY

chown -R root:root "$DEST"
find "$DEST" -type d -exec chmod 0755 {} \;
find "$DEST" -type f -exec chmod 0644 {} \;

echo "deployed web to $DEST"
echo "  game URL: ${PUBLIC_URL}/"
echo "  API URL:  ${PUBLIC_URL} (same origin)"
