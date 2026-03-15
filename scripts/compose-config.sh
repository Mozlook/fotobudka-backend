#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

MODE="${1:-dev}"

if [ "$MODE" = "prod" ]; then
  docker compose -f compose.yaml config
else
  docker compose -f compose.yaml -f compose.dev.yaml config
fi
