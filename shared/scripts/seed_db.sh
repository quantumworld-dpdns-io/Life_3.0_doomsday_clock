#!/usr/bin/env bash
set -euo pipefail
psql "${DATABASE_URL:-postgres://life3:life3@localhost:5432/life3}" \
  -f "$(dirname "$0")/../../services/intelligence/app/migrations/001_init.sql"
echo "Database seeded."
