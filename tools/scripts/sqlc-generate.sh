#!/usr/bin/env bash
set -euo pipefail

# Run sqlc generate for every sqlc.yaml found under the search root (default: repo root).

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

cd "$ROOT_DIR"

SEARCH_ROOT="$(cd "${1:-.}" && pwd)"

configs="$(find "$SEARCH_ROOT" -name sqlc.yaml -not -path '*/.git/*' | sort)"

if [[ -z "$configs" ]]; then
    echo "No sqlc.yaml files found"
    exit 0
fi

while IFS= read -r config; do
    [[ -z "$config" ]] && continue
    echo "Generating: $config"
    go tool -modfile=tools/sqlc-go.mod sqlc generate -f "$config"
done <<< "$configs"
