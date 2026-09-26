#!/usr/bin/env bash
set -euo pipefail

# Run go-mutesting only on packages that contain *_test.go files.
# Each package runs in its own git worktree for isolation (mutations are in-place).
# Worktrees are cleaned up automatically on exit.

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
MODULE="github.com/WinPooh32/insight"

WORKTREE_BASE=$(mktemp -d)
RESULT_DIR=$(mktemp -d)

cleanup() {
    # Remove worktrees and temp dirs (git worktree prune + rmdir)
    git -C "$ROOT_DIR" worktree prune -- "$WORKTREE_BASE" 2>/dev/null || true
    rm -rf "$WORKTREE_BASE" "$RESULT_DIR"
}
trap cleanup EXIT

cd "$ROOT_DIR"

# Find directories with _test.go files under cmd/insight/internal,
# deduplicate, and convert to full module paths.
PACKAGES=$(find cmd/insight/internal -mindepth 2 -name '*_test.go' \
    -exec dirname {} \; \
    | sort -u \
    | sed "s|^|${MODULE}/|")

if [[ -z "$PACKAGES" ]]; then
    echo "No test packages found"
    exit 0
fi

# Limit parallelism to available CPUs (min 2), override with MUTEST_PARALLEL.
CPUS=$(nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 2)
PARALLEL=${MUTEST_PARALLEL:-$((CPUS > 2 ? CPUS : 2))}

# Run mutation testing in a git worktree.
mutest_pkg() {
    local pkg="$1"
    local idx="$2"

    local wt="$WORKTREE_BASE/wt_${idx}"
    git worktree add --detach "$wt" HEAD

    local log="$RESULT_DIR/${idx}.log"
    local score_file="$RESULT_DIR/${idx}.score"

    (
        cd "$wt"
        go tool -modfile=misc/go-mutesting-go.mod go-mutesting \
            --exec=misc/scripts/mutate-test.sh \
            --config "$ROOT_DIR/.mutesting.yml" "$pkg" 2>&1 | tee /dev/stderr
    ) >"$log" || true

    # Extract score and total from log.
    local score total
    score=$(grep -oP 'The mutation score is \K[0-9]+\.[0-9]+' "$log" | tail -1) || true
    total=$(grep -oP 'total is \K[0-9]+' "$log" | tail -1) || true

    if [[ -n "$score" && -n "$total" ]]; then
        echo "$pkg $score $total" > "$score_file"
    fi
}

export -f mutest_pkg 2>/dev/null || true

# Launch jobs, respecting parallelism limit.
IDX=0
PIDS=()
while IFS= read -r pkg; do
    echo ">>> Starting: $pkg"
    mutest_pkg "$pkg" "$IDX" &
    PIDS+=($!)
    IDX=$((IDX + 1))

    # Throttle: wait for an existing job before launching more.
    if [[ ${#PIDS[@]} -ge $PARALLEL ]]; then
        wait "${PIDS[0]}" || true
        PIDS=("${PIDS[@]:1}")
    fi
done <<< "$PACKAGES"

# Wait for remaining jobs.
for pid in "${PIDS[@]}"; do
    wait "$pid" || true
done

# Collect results.
MIN_SCORE=""
TOTAL_MUTANTS=0
PACKAGE_COUNT=0
declare -a PKG_SCORES=()

for score_file in "$RESULT_DIR"/*.score; do
    [[ -f "$score_file" ]] || continue
    read -r pkg score total < "$score_file"
    TOTAL_MUTANTS=$((TOTAL_MUTANTS + total))
    PACKAGE_COUNT=$((PACKAGE_COUNT + 1))
    PKG_SCORES+=("  $pkg: $score ($total mutants)")

    # Only consider packages with actual mutants for minimum score.
    if [[ "$total" -gt 0 ]]; then
        if [[ -z "$MIN_SCORE" ]]; then
            MIN_SCORE="$score"
        else
            BELOW=$(echo "$score < $MIN_SCORE" | bc)
            if [[ "$BELOW" -eq 1 ]]; then
                MIN_SCORE="$score"
            fi
        fi
    fi
done

if [[ $TOTAL_MUTANTS -gt 0 && -n "$MIN_SCORE" ]]; then
    echo ""
    echo "========================================"
    echo "Package mutation scores:"
    for line in "${PKG_SCORES[@]}"; do
        echo "$line"
    done
    echo "----------------------------------------"
    echo "Min mutation score: $MIN_SCORE ($PACKAGE_COUNT packages, $TOTAL_MUTANTS total mutants)"
    echo "========================================"

    # Check against threshold if set.
    if [[ -n "${MUTEST_THRESHOLD:-}" ]]; then
        BELOW=$(echo "$MIN_SCORE < $MUTEST_THRESHOLD" | bc)
        if [[ "$BELOW" -eq 1 ]]; then
            echo "FAIL: mutation score $MIN_SCORE is below threshold $MUTEST_THRESHOLD"
            exit 1
        fi
    fi
fi
