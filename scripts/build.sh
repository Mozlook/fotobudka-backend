#!/usr/bin/env bash
set -euo pipefail

mkdir -p bin
go build -o bin/api ./cmd/api
go build -o bin/worker ./cmd/worker
