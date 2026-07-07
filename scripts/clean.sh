#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
echo "🧹 Cleaning generated dirs in workspace..."

apps_and_pkgs="$ROOT_DIR/apps $ROOT_DIR/packages"

clean_targets=(
  node_modules
  .turbo
  .next
  dist
  build
  coverage
  out
  .velite
  tsconfig.tsbuildinfo
)

for dir_type in $apps_and_pkgs; do
  [ -d "$dir_type" ] || continue
  for d in "$dir_type"/*/; do
    [ -d "$d" ] || continue
    for target in "${clean_targets[@]}"; do
      if [ -e "$d$target" ]; then
        echo "  rm -rf ${d#$ROOT_DIR/}$target"
        rm -rf "$d$target"
      fi
    done
  done
done

if [ -d "$ROOT_DIR/node_modules" ]; then
  echo "  rm -rf node_modules/"
  rm -rf "$ROOT_DIR/node_modules"
fi

if [ -d "$ROOT_DIR/.turbo" ]; then
  echo "  rm -rf .turbo/"
  rm -rf "$ROOT_DIR/.turbo"
fi

echo "✅ Clean up complete! Run 'pnpm install' to recreate node_modules."
