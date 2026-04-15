#!/usr/bin/env bash
# migrate.sh — wrapper untuk golang-migrate
# Usage: bash scripts/migrate.sh [up|down] "$DATABASE_URL"
set -euo pipefail

DIRECTION=${1:-up}
DATABASE_URL=${2:-}

if [[ -z "$DATABASE_URL" ]]; then
  echo "ERROR: DATABASE_URL tidak diset" >&2
  exit 1
fi

if ! command -v migrate &>/dev/null; then
  echo "ERROR: golang-migrate tidak ditemukan. Install via: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest" >&2
  exit 1
fi

MIGRATIONS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/migrations"

echo ">>> Running migrate $DIRECTION ..."
migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" "$DIRECTION"
echo "✓ Done"
