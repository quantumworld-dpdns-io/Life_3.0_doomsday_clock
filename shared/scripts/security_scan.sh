#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
failures=0

export DOCKER_CONFIG="${DOCKER_CONFIG:-/tmp/life3-empty-docker-config}"
mkdir -p "$DOCKER_CONFIG"

export TRIVY_DB_REPOSITORY="${TRIVY_DB_REPOSITORY:-ghcr.io/aquasecurity/trivy-db}"
export TRIVY_JAVA_DB_REPOSITORY="${TRIVY_JAVA_DB_REPOSITORY:-ghcr.io/aquasecurity/trivy-java-db}"
export SEMGREP_LOG_FILE="${SEMGREP_LOG_FILE:-/tmp/life3-semgrep.log}"

if [ -z "${SSL_CERT_FILE:-}" ] && command -v python3 >/dev/null 2>&1; then
  cert_file="$(python3 - <<'PY' 2>/dev/null || true
try:
    import certifi
    print(certifi.where())
except Exception:
    pass
PY
)"
  if [ -n "$cert_file" ] && [ -f "$cert_file" ]; then
    export SSL_CERT_FILE="$cert_file"
  fi
fi

have() {
  command -v "$1" >/dev/null 2>&1
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

if have trivy; then
  run_step "Trivy filesystem scan" bash -lc "cd '$ROOT' && trivy fs --config trivy.yaml ."
else
  printf '==> SKIP: Trivy is not installed\n'
fi

if have semgrep; then
  if [ -f "${HOME:-}/.semgrep/settings.yml" ] || [ -n "${SEMGREP_APP_TOKEN:-}" ]; then
    run_step "Semgrep CI scan" bash -lc "cd '$ROOT' && SEMGREP_SEND_METRICS=off semgrep ci"
  else
    printf '==> SKIP: Semgrep is installed but not logged in; run semgrep login\n'
  fi
else
  printf '==> SKIP: Semgrep is not installed\n'
fi

exit "$failures"
