#!/usr/bin/env bash

set -euo pipefail

project_dir="$(pwd)"
output_file="${project_dir}/code_context.txt"

# Coverage utama source + docs + database + agent files
# Go backend: internal, cmd, pkg
# Configuration: db, documentation, tests
# Monorepo infrastructure: turbo config, docker-compose
include_dirs=("internal" "cmd" "pkg" "db" "documentation" "tests" "packages" "deploy")

# File root penting untuk konteks project
include_root_files=(
  "README.md"
  "AGENTS.md"
  "go.mod"
  "go.sum"
  "Makefile"
  "turbo.json"
  "pnpm-workspace.yaml"
  "docker-compose.dev.yml"
  "docker-compose.prod.yml"
  ".env.example"
  "lefthook.yml"
  "casbin.json"
)

# Extension/file biner yang tidak perlu dikumpulkan
should_ignore_file() {
  local path="$1"
  case "${path,,}" in
    *.png|*.jpg|*.jpeg|*.gif|*.ico|*.svg|*.pdf|*.exe|*.drawio|*.pack|*.idx|*.rev)
      return 0
      ;;
    *.lock|*.lockb)
      return 0
      ;;
  esac
  return 1
}

# Skip direktori yang tidak relevan untuk code context
should_ignore_dir() {
  local dir_name="$1"
  case "$dir_name" in
    node_modules|.git|.next|dist|build|__pycache__|.venv|coverage)
      return 0
      ;;
  esac
  return 1
}

append_file() {
  local file_path="$1"
  if [[ "$file_path" == "$output_file" ]]; then
    return
  fi
  if should_ignore_file "$file_path"; then
    return
  fi
  if [[ ! -f "$file_path" ]]; then
    return
  fi

  local relative_path="${file_path#"$project_dir/"}"
  printf "// File: %s\n" "$relative_path" >> "$output_file"
  cat "$file_path" >> "$output_file"
  printf "\n\n" >> "$output_file"
}

rm -f "$output_file"

echo "🔍 Mengumpulkan code context dari project Casbin..."

for root_file in "${include_root_files[@]}"; do
  if [[ -f "${project_dir}/${root_file}" ]]; then
    append_file "${project_dir}/${root_file}"
  fi
done

for dir in "${include_dirs[@]}"; do
  abs_dir="${project_dir}/${dir}"
  if [[ ! -d "$abs_dir" ]]; then
    continue
  fi

  echo "  📁 Processing: $dir/"
  
  # Use find with exclusions for node_modules, .git, etc
  while IFS= read -r file_path; do
    # Skip if any component of the path should be ignored
    if [[ "$file_path" =~ /node_modules/|/.git/|/.next/|/dist/|/build/|/__pycache__/|/.venv/ ]]; then
      continue
    fi
    append_file "$file_path"
  done < <(find "$abs_dir" -type f ! -path '*/node_modules/*' ! -path '*/.git/*' ! -path '*/.next/*' ! -path '*/dist/*' ! -path '*/build/*' ! -path '*/__pycache__/*' ! -path '*/.venv/*' 2>/dev/null | sort)
done

printf "\n✅ Generated: %s\n" "$output_file"
wc -l "$output_file"
