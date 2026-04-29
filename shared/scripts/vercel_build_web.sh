#!/usr/bin/env bash
set -euo pipefail

root="$(pwd)"
if [ -d apps/web ]; then
  web="$(cd apps/web && pwd)"
else
  web="$(cd ../../apps/web && pwd)"
fi

cd "$web"
npm run build

target="$root/apps/web/dist"
if [ "$web/dist" != "$target" ]; then
  mkdir -p "$root/apps/web"
  rm -rf "$target"
  cp -R "$web/dist" "$target"
fi
