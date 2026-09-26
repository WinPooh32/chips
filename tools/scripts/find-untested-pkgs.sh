#!/usr/bin/env bash
set -euo pipefail

# Find Go packages that have .go files but no *_test.go files.
# Useful for identifying coverage gaps.

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
MODULE="github.com/WinPooh32/insight"

# Directories to exclude from the report.
EXCLUDE_DIRS=("db" "migrations" "config" "cmd/insight")

cd "$ROOT_DIR"

# Check if any component of the path matches the exclude list.
is_excluded() {
    local dir="$1"
    for excl in "${EXCLUDE_DIRS[@]}"; do
        if [[ "$dir" == *"/${excl}"* ]]; then
            return 0
        fi
    done
    return 1
}

# Find directories containing .go files (non-test), excluding common non-code dirs.
# Then filter out directories that have *_test.go or match the exclude list.
untested=()
while IFS= read -r dir; do
    # Skip empty lines.
    [[ -z "$dir" ]] && continue

    # Skip excluded directories.
    is_excluded "$dir" && continue

    # Skip directories that have test files.
    test_files=$(find "$dir" -maxdepth 1 -name '*_test.go' -print -quit) || true
    [[ -n "$test_files" ]] && continue

    untested+=("$dir")
done < <(find . -name '*.go' ! -name '*_test.go' -not -path './.git/*' -not -path './vendor/*' -not -path './.claude/*' -not -path './node_modules/*' -exec dirname {} \; | sort -u)

if [[ ${#untested[@]} -eq 0 ]]; then
    echo "All packages have tests."
    exit 0
fi

echo "Packages without tests (${#untested[@]}):"
echo ""
for dir in "${untested[@]}"; do
    # Convert filepath to module path.
    pkg="${MODULE}/${dir#./}"
    file_count=$(find "$dir" -maxdepth 1 -name '*.go' ! -name '*_test.go' | wc -l)
    echo "  $pkg  ($file_count .go file(s))"
done

exit 1
