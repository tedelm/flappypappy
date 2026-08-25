#!/usr/bin/env bash
# Install Go score API + SQLite + Caddy (game + API) on Ubuntu.
# Run as root from a clone of this repo:
#   sudo bash scripts/install-score-api.sh
# Optional: SCORE_API_HOST=other.example.com WEB_ROOT=/path/to/docs
set -euo pipefail

if [[ "${EUID}" -ne 0 ]]; then
  echo "run as root: sudo bash scripts/install-score-api.sh" >&2
  exit 1
fi

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SRC="$ROOT/src"
GO_VERSION="${GO_VERSION:-1.25.6}"
LISTEN="${LISTEN:-127.0.0.1:8088}"
DB_PATH="${DB_PATH:-/var/lib/flappy/flappypappy.sqlite}"
SCORE_API_HOST="${SCORE_API_HOST:-flappy-pappy.se}"
WEB_DEST="${WEB_DEST:-/var/www/flappy}"
WEB_ROOT="${WEB_ROOT:-}"
SKIP_CADDY="${SKIP_CADDY:-0}"
SKIP_UFW="${SKIP_UFW:-0}"

export DEBIAN_FRONTEND=noninteractive
apt-get update -y
apt-get install -y --no-install-recommends \
  ca-certificates curl sqlite3 tar gzip ufw python3 rsync \
  build-essential libsqlite3-dev \
  debian-keyring debian-archive-keyring apt-transport-https gnupg

install_go() {
  local have=""
  if command -v go >/dev/null 2>&1; then
    have="$(go env GOVERSION 2>/dev/null || go version | awk '{print $3}')"
  fi
  if [[ "$have" == "go${GO_VERSION}" ]]; then
    echo "Go ${GO_VERSION} already installed"
    return
  fi
  local arch
  arch="$(uname -m)"
  case "$arch" in
    x86_64) arch=amd64 ;;
    aarch64|arm64) arch=arm64 ;;
    *)
      echo "unsupported architecture: $arch" >&2
      exit 1
      ;;
  esac
  local tarball="go${GO_VERSION}.linux-${arch}.tar.gz"
  curl -fsSL "https://go.dev/dl/${tarball}" -o "/tmp/${tarball}"
  rm -rf /usr/local/go
  tar -C /usr/local -xzf "/tmp/${tarball}"
  rm -f "/tmp/${tarball}"
  ln -sfn /usr/local/go/bin/go /usr/local/bin/go
  ln -sfn /usr/local/go/bin/gofmt /usr/local/bin/gofmt
  echo "installed Go ${GO_VERSION}"
}

install_go
export PATH="/usr/local/go/bin:/usr/local/bin:${PATH}"

if [[ ! -f "$SRC/go.mod" ]]; then
  echo "expected Go module at $SRC/go.mod (copy the repo onto this host first)" >&2
  exit 1
fi

id -u flappy >/dev/null 2>&1 || useradd --system --home /var/lib/flappy --shell /usr/sbin/nologin flappy
install -d -o flappy -g flappy -m 0750 /var/lib/flappy
install -d -o root -g flappy -m 0750 /etc/flappy
install -d -o root -g root -m 0755 "$WEB_DEST"

ENV_FILE=/etc/flappy/scoreapi.env
if [[ -f "$ENV_FILE" ]]; then
  # shellcheck disable=SC1090
  set -a
  # shellcheck source=/dev/null
  source "$ENV_FILE"
  set +a
fi
if [[ -z "${API_KEY:-}" ]]; then
  API_KEY="$(openssl rand -hex 24 2>/dev/null || python3 -c 'import secrets; print(secrets.token_hex(24))')"
fi
if [[ -z "${RUN_HMAC_SECRET:-}" ]]; then
  RUN_HMAC_SECRET="$(openssl rand -hex 32 2>/dev/null || python3 -c 'import secrets; print(secrets.token_hex(32))')"
fi

cat > "$ENV_FILE" <<EOF
LISTEN=${LISTEN}
DB_PATH=${DB_PATH}
API_KEY=${API_KEY}
RUN_HMAC_SECRET=${RUN_HMAC_SECRET}
EOF
chown root:flappy "$ENV_FILE"
chmod 0640 "$ENV_FILE"

if [[ -n "${SCOREAPI_BIN:-}" ]]; then
  echo "installing prebuilt scoreapi from ${SCOREAPI_BIN}..."
  install -m 0755 "$SCOREAPI_BIN" /usr/local/bin/scoreapi
else
  echo "building scoreapi..."
  cd "$SRC"
  # Low-RAM hosts (common on Strato): single compiler job avoids OOM kills.
  # Uses CGO + libsqlite3 (mattn) instead of modernc.org/sqlite which OOMs easily.
  export CGO_ENABLED=1
  export GOMAXPROCS="${GOMAXPROCS:-1}"
  if ! go build -p 1 -o /usr/local/bin/scoreapi ./cmd/scoreapi; then
    echo "build failed (often OOM). Options:" >&2
    echo "  1) add swap, then re-run" >&2
    echo "  2) cross-build on a larger machine and copy:" >&2
    echo "       GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o scoreapi ./cmd/scoreapi" >&2
    echo "       SCOREAPI_BIN=./scoreapi sudo -E bash scripts/install-score-api.sh" >&2
    exit 1
  fi
fi
chown root:root /usr/local/bin/scoreapi
chmod 0755 /usr/local/bin/scoreapi

install -m 0644 "$ROOT/scripts/score-api.service" /etc/systemd/system/score-api.service
systemctl daemon-reload
systemctl enable --now score-api.service

write_web_config() {
  local dest="$1"
  local public_url="https://${SCORE_API_HOST}"
  export PUBLIC_URL="$public_url"
  export API_KEY
  export DEST="$dest"
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
}

# Optional: copy a prebuilt docs/ tree (build WASM off-box — OOM risk on small VPS).
if [[ -n "$WEB_ROOT" ]]; then
  if [[ ! -f "$WEB_ROOT/index.html" ]]; then
    echo "WEB_ROOT=$WEB_ROOT has no index.html" >&2
    exit 1
  fi
  echo "deploying web from $WEB_ROOT -> $WEB_DEST"
  rsync -a --delete --exclude '.gitkeep' "$WEB_ROOT"/ "$WEB_DEST"/
elif [[ -f "$ROOT/docs/index.html" ]]; then
  echo "deploying web from $ROOT/docs -> $WEB_DEST"
  rsync -a --delete --exclude '.gitkeep' "$ROOT/docs"/ "$WEB_DEST"/
else
  echo "no web build found (set WEB_ROOT or place docs/ after build-web.sh). Serving placeholder."
  if [[ ! -f "$WEB_DEST/index.html" ]]; then
    cat > "$WEB_DEST/index.html" <<EOF
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>Flappy Pappy</title>
</head>
<body>
  <p>Game not deployed yet. Build with <code>scripts/build-web.sh</code> then
  <code>sudo bash scripts/deploy-web.sh</code>.</p>
</body>
</html>
EOF
  fi
fi
write_web_config "$WEB_DEST"
chown -R root:root "$WEB_DEST"
find "$WEB_DEST" -type d -exec chmod 0755 {} \;
find "$WEB_DEST" -type f -exec chmod 0644 {} \;

install_caddy() {
  if ! command -v caddy >/dev/null 2>&1; then
    curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' \
      | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
    curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' \
      | tee /etc/apt/sources.list.d/caddy-stable.list >/dev/null
    apt-get update -y
    apt-get install -y caddy
  fi
  local tmp
  tmp="$(mktemp)"
  sed "s/{\$SCORE_API_HOST}/${SCORE_API_HOST}/g" "$ROOT/scripts/score-api.Caddyfile" > "$tmp"
  install -m 0644 "$tmp" /etc/caddy/Caddyfile
  rm -f "$tmp"
  systemctl enable --now caddy
  systemctl reload caddy || systemctl restart caddy
}

if [[ "$SKIP_CADDY" != "1" ]]; then
  if [[ -z "$SCORE_API_HOST" ]]; then
    echo "SCORE_API_HOST is empty: skipping Caddy." >&2
  else
    install_caddy
  fi
fi

if [[ "$SKIP_UFW" != "1" ]] && command -v ufw >/dev/null 2>&1; then
  ufw allow OpenSSH || true
  if [[ "$SKIP_CADDY" != "1" && -n "$SCORE_API_HOST" ]]; then
    ufw allow 80/tcp || true
    ufw allow 443/tcp || true
  fi
  ufw --force enable || true
fi

systemctl --no-pager --full status score-api.service || true

echo
echo "score API + site installed."
echo "  health (local): curl -sS http://${LISTEN}/health"
echo "  API_KEY is in ${ENV_FILE} (client GET /scores)"
echo "  RUN_HMAC_SECRET is in ${ENV_FILE} (server-only; never put in config.js)"
if [[ -n "$SCORE_API_HOST" && "$SKIP_CADDY" != "1" ]]; then
  echo "  public game URL: https://${SCORE_API_HOST}/"
  echo "  FLAPPY_SCORE_API_URL=https://${SCORE_API_HOST}"
fi
echo "  set FLAPPY_SCORE_API_KEY from ${ENV_FILE}"
echo "  to refresh the game: build off-box, then sudo bash scripts/deploy-web.sh"
