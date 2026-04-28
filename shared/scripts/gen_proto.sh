#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
PROTO_DIR="$ROOT/shared/proto"

echo "==> Generating Go stubs..."
mkdir -p "$ROOT/services/api-gateway/internal/proto"
protoc -I "$PROTO_DIR" \
  --go_out="$ROOT/services/api-gateway/internal/proto" \
  --go-grpc_out="$ROOT/services/api-gateway/internal/proto" \
  "$PROTO_DIR"/*.proto

echo "==> Generating Python stubs..."
mkdir -p "$ROOT/services/intelligence/app/proto"
mkdir -p "$ROOT/services/quantum-sim/app/proto"
python -m grpc_tools.protoc -I "$PROTO_DIR" \
  --python_out="$ROOT/services/intelligence/app/proto" \
  --grpc_python_out="$ROOT/services/intelligence/app/proto" \
  "$PROTO_DIR"/*.proto
python -m grpc_tools.protoc -I "$PROTO_DIR" \
  --python_out="$ROOT/services/quantum-sim/app/proto" \
  --grpc_python_out="$ROOT/services/quantum-sim/app/proto" \
  "$PROTO_DIR"/*.proto

echo "==> Proto generation complete."
