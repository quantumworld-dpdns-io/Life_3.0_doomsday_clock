#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
PROTO_DIR="$ROOT/shared/proto"
STRICT="${PROTO_STRICT:-0}"

export UV_CACHE_DIR="${UV_CACHE_DIR:-/tmp/life3-uv-cache}"

have() {
  command -v "$1" >/dev/null 2>&1
}

skip_or_fail() {
  local message="$1"
  if [ "$STRICT" = "1" ]; then
    echo "ERROR: $message" >&2
    exit 1
  fi
  echo "==> SKIP: $message"
}

python_bin_for() {
  local service_dir="$1"
  if have uv; then
    printf 'uv'
  elif have python3; then
    printf 'python3'
  elif have python; then
    printf 'python'
  else
    return 1
  fi
}

[ -d "$PROTO_DIR" ] || {
  echo "ERROR: missing proto directory: $PROTO_DIR" >&2
  exit 1
}

shopt -s nullglob
proto_files=("$PROTO_DIR"/*.proto)
shopt -u nullglob
[ "${#proto_files[@]}" -gt 0 ] || {
  echo "ERROR: no .proto files found in $PROTO_DIR" >&2
  exit 1
}

generate_go() {
  local gateway_dir="$ROOT/services/api-gateway"

  [ -f "$gateway_dir/go.mod" ] || {
    echo "==> SKIP: Go proto stubs; services/api-gateway is scaffold-only"
    return 0
  }
  have protoc || {
    skip_or_fail "Go proto stubs require protoc"
    return 0
  }
  have protoc-gen-go || {
    skip_or_fail "Go proto stubs require protoc-gen-go"
    return 0
  }
  have protoc-gen-go-grpc || {
    skip_or_fail "Go proto stubs require protoc-gen-go-grpc"
    return 0
  }

  echo "==> Generating Go stubs..."
  mkdir -p "$gateway_dir/internal/proto"
  protoc -I "$PROTO_DIR" \
    --go_out="$gateway_dir/internal/proto" \
    --go-grpc_out="$gateway_dir/internal/proto" \
    "${proto_files[@]}"
}

generate_python() {
  local service="$1"
  local out_dir="$2"
  local service_dir="$ROOT/services/$service"

  [ -f "$service_dir/pyproject.toml" ] || {
    echo "==> SKIP: Python proto stubs for $service; missing pyproject.toml"
    return 0
  }

  local py_cmd
  if ! py_cmd="$(python_bin_for "$service_dir")"; then
    skip_or_fail "Python proto stubs for $service require uv, python3, or python"
    return 0
  fi

  local env_dir="/tmp/life3-uv-env-$service"
  if [ "$py_cmd" = "uv" ]; then
    if ! (cd "$service_dir" && UV_PROJECT_ENVIRONMENT="$env_dir" uv sync >/dev/null); then
      skip_or_fail "Python proto stubs for $service could not sync dependencies with uv"
      return 0
    fi
    py_cmd="$env_dir/bin/python"
  fi

  if ! (cd "$service_dir" && "$py_cmd" -m grpc_tools.protoc --version >/dev/null 2>&1); then
    skip_or_fail "Python proto stubs for $service require grpcio-tools"
    return 0
  fi

  echo "==> Generating Python stubs for $service..."
  mkdir -p "$out_dir"
  touch "$out_dir/__init__.py"
  (cd "$service_dir" && "$py_cmd" -m grpc_tools.protoc -I "$PROTO_DIR" \
    --python_out="$out_dir" \
    --grpc_python_out="$out_dir" \
    "${proto_files[@]}")
}

generate_go
generate_python "intelligence" "$ROOT/services/intelligence/app/proto"
generate_python "quantum-sim" "$ROOT/services/quantum-sim/app/proto"

echo "==> Proto generation complete."
