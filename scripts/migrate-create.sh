#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

NAME="${1:?usage: ./scripts/migrate-create.sh migration_name}"

mkdir -p migrations

docker run --rm \
  -v "$ROOT_DIR/migrations:/migrations" \
  migrate/migrate \
  create -ext sql -dir /migrations -seq "$NAME"
