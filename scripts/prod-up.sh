#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

mkdir -p logs

LOCAL_UID="$(id -u)"
LOCAL_GID="$(id -g)"

LOCAL_UID="$LOCAL_UID" LOCAL_GID="$LOCAL_GID" docker compose \
  -p fotobudka \
  -f compose.yaml \
  up -d --build
