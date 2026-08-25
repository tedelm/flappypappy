#!/usr/bin/env bash
# Copy sites/pappapuben into /var/www/pappapuben on this host.
#   sudo bash scripts/deploy-pappapuben.sh
#   sudo PAPPAPUBEN_SRC=/path/to/sites/pappapuben bash scripts/deploy-pappapuben.sh
set -euo pipefail

if [[ "${EUID}" -ne 0 ]]; then
  echo "run as root: sudo bash scripts/deploy-pappapuben.sh" >&2
  exit 1
fi

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PAPPAPUBEN_SRC="${PAPPAPUBEN_SRC:-$ROOT/sites/pappapuben}"
DEST="${DEST:-/var/www/pappapuben}"

if [[ ! -f "$PAPPAPUBEN_SRC/index.html" ]]; then
  echo "no landing at $PAPPAPUBEN_SRC (expected index.html)" >&2
  exit 1
fi

install -d -o root -g root -m 0755 "$DEST"
rsync -a --delete \
  --exclude '.gitkeep' \
  "$PAPPAPUBEN_SRC"/ "$DEST"/

chown -R root:root "$DEST"
find "$DEST" -type d -exec chmod 0755 {} \;
find "$DEST" -type f -exec chmod 0644 {} \;

echo "deployed pappapuben landing to $DEST"
echo "  URL: https://pappapuben.se/"
