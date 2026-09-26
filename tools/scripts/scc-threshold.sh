#!/usr/bin/env bash
set -euo pipefail

# Print per-directory Go code line counts under a target path (default: repo root)
# and enforce thresholds: warn at >= WARN_LINES, fail (exit 1) at >= ERROR_LINES.

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

WARN_LINES=700
ERROR_LINES=1000

cd "$ROOT_DIR"

counts=$(go tool -modfile=misc/scc-go.mod scc --no-gen --include-ext go --no-size -M '_test\.go$' --by-file --format json "${1:-.}" \
    | jq -r '
        [ .[].Files[]
          | { dir: (.Location | rindex("/") as $i
                    | if $i == null then "." else .[0:$i] | sub("^\\./"; ".") end),
              code: .Code } ]
        | group_by(.dir)
        | map({ key: .[0].dir, value: (map(.code) | add) })
        | .[] | "\(.value) \(.key)"
      ')

failed=0
while read -r count dir; do
    [[ -z "$dir" ]] && continue
    echo "$count $dir"
    if (( count >= ERROR_LINES )); then
        echo "ERROR: $dir: $count lines of Go code (>= $ERROR_LINES)"
        failed=1
    elif (( count >= WARN_LINES )); then
        echo "WARN: $dir: $count lines of Go code (>= $WARN_LINES)"
    fi
done <<< "$counts"

exit "$failed"
