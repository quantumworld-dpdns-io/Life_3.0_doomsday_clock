#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
MIGRATION="$ROOT/services/intelligence/app/migrations/001_init.sql"

if [ -f "$ROOT/.env" ]; then
  set -a
  # shellcheck disable=SC1091
  . "$ROOT/.env"
  set +a
fi

DATABASE_URL="${DATABASE_URL:-postgres://life3:life3@localhost:5432/life3}"

[ -f "$MIGRATION" ] || {
  echo "ERROR: migration file not found: $MIGRATION" >&2
  exit 1
}

if command -v psql >/dev/null 2>&1; then
  psql "$DATABASE_URL" -f "$MIGRATION"
elif command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
  docker compose exec -T postgres psql \
    -U "${POSTGRES_USER:-life3}" \
    -d "${POSTGRES_DB:-life3}" <"$MIGRATION"
else
  echo "ERROR: migrate requires either psql or docker compose" >&2
  exit 1
fi

echo "Database migration complete."
