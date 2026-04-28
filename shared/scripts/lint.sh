#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
failures=0

export UV_CACHE_DIR="${UV_CACHE_DIR:-/tmp/life3-uv-cache}"
export GOCACHE="${GOCACHE:-/tmp/life3-go-build-cache}"
export GOMODCACHE="${GOMODCACHE:-/tmp/life3-go-mod-cache}"
export CARGO_HOME="${CARGO_HOME:-/tmp/life3-cargo-home}"
export CARGO_TARGET_DIR="${CARGO_TARGET_DIR:-/tmp/life3-risk-engine-target}"

have() {
  command -v "$1" >/dev/null 2>&1
}

skip() {
  printf '==> SKIP: %s\n' "$1"
}

run_step() {
  local name="$1"
  shift
  printf '==> %s\n' "$name"
  if ! "$@"; then
    printf '!! FAILED: %s\n' "$name" >&2
    failures=$((failures + 1))
  fi
}

lint_python_service() {
  local service="$1"
  local path="$ROOT/services/$service"

  [ -f "$path/pyproject.toml" ] || return 0
  if have uv; then
    run_step "Python lint: $service" bash -lc "cd '$path' && uv run ruff check ."
  elif have ruff; then
    run_step "Python lint: $service" bash -lc "cd '$path' && ruff check ."
  else
    skip "Python lint for $service requires uv or ruff"
  fi
}

lint_go_service() {
  local path="$ROOT/services/api-gateway"

  [ -f "$path/go.mod" ] || {
    skip "Go lint: services/api-gateway is scaffold-only"
    return 0
  }
  have go || {
    skip "Go lint requires go"
    return 0
  }

  run_step "Go format check: api-gateway" bash -lc "cd '$path' && test -z \"\$(gofmt -l .)\""
  run_step "Go vet: api-gateway" bash -lc "cd '$path' && go vet ./..."
}

lint_rust_service() {
  local path="$ROOT/services/risk-engine"

  [ -f "$path/Cargo.toml" ] || return 0
  have cargo || {
    skip "Rust lint requires cargo"
    return 0
  }

  run_step "Rust lint: risk-engine" bash -lc "cd '$path' && cargo clippy -- -D warnings"
}

lint_web_app() {
  local path="$ROOT/apps/web"

  [ -f "$path/package.json" ] || {
    skip "Frontend lint: apps/web has no package.json"
    return 0
  }
  have npm || {
    skip "Frontend lint requires npm"
    return 0
  }
  find "$path/src" -type f \( -name '*.ts' -o -name '*.tsx' -o -name '*.js' -o -name '*.jsx' \) \
    ! -name 'vite-env.d.ts' | grep -q . || {
    skip "Frontend lint: apps/web has no implementation files yet"
    return 0
  }
  [ -d "$path/node_modules" ] || {
    skip "Frontend lint requires node_modules; run npm install in apps/web"
    return 0
  }

  run_step "Frontend lint: web" bash -lc "cd '$path' && npm run lint"
}

lint_go_service
lint_python_service intelligence
lint_python_service quantum-sim
lint_rust_service
lint_web_app

exit "$failures"
