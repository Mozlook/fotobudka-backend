#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

set -a
source .env
set +a

PROJECT_NAME="fotobudka-dev"
NETWORK_NAME="${PROJECT_NAME}_app"
TEST_DB_URL="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_TEST_DB}?sslmode=disable"

run_version() {
  local database_url="$1"

  docker run --rm \
    --network "$NETWORK_NAME" \
    -v "$ROOT_DIR/migrations:/migrations" \
    migrate/migrate \
    -path=/migrations \
    -database "$database_url" \
    version
}

echo "==> dev"
run_version "$DB_URL" || true

echo
echo "==> test"
run_version "$TEST_DB_URL" || true
