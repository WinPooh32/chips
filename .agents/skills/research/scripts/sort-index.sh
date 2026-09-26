#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
file=${1:-$SCRIPT_DIR/../index.md}
tmp=$(mktemp "${file}.XXXXXX")
trap 'rm -f "$tmp"' EXIT
awk -F' — ' '{d=(NF>=2 ? $2 : ""); print d "\t" $0}' "$file" \
  | sort -s -t"$(printf '\t')" -k1,1r \
  | cut -f2- > "$tmp"
mv "$tmp" "$file"
trap - EXIT
