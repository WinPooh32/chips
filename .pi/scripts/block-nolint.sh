#!/usr/bin/env bash
# Block edits and commands that introduce //nolint comments.
# nolibnt should be used to suppress linter warnings — the linter rules should be fixed instead.

set -euo pipefail

# Read hook input from stdin
INPUT="$(cat)"

# Extract content based on tool type
TOOL="$(echo "$INPUT" | jq -r '.tool_name // empty')"

case "$TOOL" in
  Write)
    CONTENT="$(echo "$INPUT" | jq -r '.tool_input.content // empty')"
    ;;
  Edit)
    CONTENT="$(echo "$INPUT" | jq -r '.tool_input.new_string // empty')"
    ;;
  Bash)
    CONTENT="$(echo "$INPUT" | jq -r '.tool_input.command // empty')"
    ;;
  *)
    # Not a tool we care about
    exit 0
    ;;
esac

# Check for nolint pattern
if echo "$CONTENT" | grep -q '//nolint'; then
  echo '{"decision": "block", "reason": "//nolint comments are forbidden. Fix the underlying linter issue instead of suppressing it with //nolint.", "systemMessage": "Blocked: //nolint found in changes. Please fix the actual linter violation rather than silencing it."}'
  exit 1
fi

exit 0
