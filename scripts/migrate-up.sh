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

compose() {
  docker compose -p "$PROJECT_NAME" -f compose.yaml -f compose.dev.yaml "$@"
}

run_migrate() {
  local database_url="$1"

  set +e
  OUTPUT=$(docker run --rm \
    --network "$NETWORK_NAME" \
    -v "$ROOT_DIR/migrations:/migrations" \
    migrate/migrate \
    -path=/migrations \
    -database "$database_url" \
    up 2>&1)
  STATUS=$?
  set -e

  if [ $STATUS -ne 0 ] && [[ "$OUTPUT" != *"no change"* ]]; then
    echo "$OUTPUT" >&2
    return $STATUS
  fi

  echo "$OUTPUT"
}

echo "==> ensuring test database exists"
compose exec -T postgres sh -c "
psql -U \"$POSTGRES_USER\" -d postgres -tc \"SELECT 1 FROM pg_database WHERE datname = '${POSTGRES_TEST_DB}'\" | grep -q 1 || \
psql -U \"$POSTGRES_USER\" -d postgres -c \"CREATE DATABASE ${POSTGRES_TEST_DB}\"
"

echo "==> migrate dev database"
run_migrate "$DB_URL"

echo "==> migrate test database"
run_migrate "$TEST_DB_URL"

echo "==> done"
