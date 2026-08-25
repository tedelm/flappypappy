#!/usr/bin/env bash
# One-shot import from SQLite Cloud into a local SQLite file.
# pip install sqlitecloud
set -euo pipefail
exec python3 "$(dirname "$0")/sqlitecloud_dump.py" "$@"
