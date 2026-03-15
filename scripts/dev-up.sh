#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

mkdir -p logs

docker compose \
  -p fotobudka-dev \
  -f compose.yaml \
  -f compose.dev.yaml \
  up -d --build
