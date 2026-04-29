#!/usr/bin/env bash
set -euo pipefail

if [ -d apps/web ]; then
  cd apps/web
else
  cd ../../apps/web
fi

npm install
