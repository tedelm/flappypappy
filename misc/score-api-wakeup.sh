#!/usr/bin/env bash
# Wake / health-check the self-hosted score API.
set -euo pipefail

URL="${1:-${FLAPPY_SCORE_API_URL:-}}"
if [[ -z "$URL" ]]; then
  echo "usage: $0 https://scores.example.com" >&2
  echo "or set FLAPPY_SCORE_API_URL" >&2
  exit 1
fi
URL="${URL%/}"
curl -fsS "${URL}/health"
echo
