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

test_python_service() {
  local service="$1"
  local path="$ROOT/services/$service"

  [ -f "$path/pyproject.toml" ] || return 0
  if have uv; then
    run_step "Python tests: $service" bash -lc "cd '$path' && uv run pytest"
  elif have python3 && python3 -m pytest --version >/dev/null 2>&1; then
    run_step "Python tests: $service" bash -lc "cd '$path' && python3 -m pytest"
  else
    skip "Python tests for $service require uv or python3 with pytest"
  fi
}

test_go_service() {
  local path="$ROOT/services/api-gateway"

  [ -f "$path/go.mod" ] || {
    skip "Go tests: services/api-gateway is scaffold-only"
    return 0
  }
  have go || {
    skip "Go tests require go"
    return 0
  }

  run_step "Go tests: api-gateway" bash -lc "cd '$path' && go test ./..."
}

test_rust_service() {
  local path="$ROOT/services/risk-engine"

  [ -f "$path/Cargo.toml" ] || return 0
  have cargo || {
    skip "Rust tests require cargo"
    return 0
  }

  run_step "Rust tests: risk-engine" bash -lc "cd '$path' && cargo test"
}

test_web_app() {
  local path="$ROOT/apps/web"

  [ -f "$path/package.json" ] || {
    skip "Frontend tests: apps/web has no package.json"
    return 0
  }
  have npm || {
    skip "Frontend tests require npm"
    return 0
  }
  find "$path/src" -type f \( -name '*.test.ts' -o -name '*.test.tsx' -o -name '*.spec.ts' -o -name '*.spec.tsx' \) \
    | grep -q . || {
    skip "Frontend tests: apps/web has no test files yet"
    return 0
  }
  [ -d "$path/node_modules" ] || {
    skip "Frontend tests require node_modules; run npm install in apps/web"
    return 0
  }

  run_step "Frontend tests: web" bash -lc "cd '$path' && npm test -- --run"
}

test_go_service
test_python_service intelligence
test_python_service quantum-sim
test_rust_service
test_web_app

exit "$failures"
