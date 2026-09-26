#!/bin/bash

# 1. Read the JSON input sent by Claude Code to stdin
INPUT=$(cat)

# 2. Extract Event and Session Context
HOOK_EVENT=$(echo "$INPUT" | jq -r '.hook_event_name // "unknown"')
SESSION_ID=$(echo "$INPUT" | jq -r '.session_id // "unknown"')

# State file specific to this session
STATE_FILE="/tmp/claude-loop-detector-${SESSION_ID}.state"

# 3. Clean up state file when the session ends
if [[ "$HOOK_EVENT" == "SessionEnd" ]]; then
    rm -f "$STATE_FILE"
    exit 0
fi

# We only care about intercepting tool usage before it executes
if [[ "$HOOK_EVENT" != "PreToolUse" ]]; then
    exit 0
fi

# 4. Extract Tool Details
TOOL_NAME=$(echo "$INPUT" | jq -r '.tool_name // "none"')
# Use jq -c to compact the JSON input into a consistent, single-line string
TOOL_INPUT=$(echo "$INPUT" | jq -c '.tool_input // {}')

# Configuration: Block if called more than N times in a row
MAX_REPEATS=3

# Create a unique signature for this specific tool call
CURRENT_SIG="${TOOL_NAME}::${TOOL_INPUT}"

# 5. Check previous state
COUNT=1
if [[ -f "$STATE_FILE" ]]; then
    LAST_SIG=$(jq -r '.last_sig // ""' "$STATE_FILE")
    OLD_COUNT=$(jq -r '.count // 0' "$STATE_FILE")

    # If the exact same tool and args are called again, increment the count
    if [[ "$LAST_SIG" == "$CURRENT_SIG" ]]; then
        COUNT=$((OLD_COUNT + 1))
    fi
fi

# 6. Save new state
jq -n --arg sig "$CURRENT_SIG" --argjson count "$COUNT" \
    '{last_sig: $sig, count: $count}' > "$STATE_FILE"

# 7. Evaluate Loop Threshold
if [[ "$COUNT" -gt "$MAX_REPEATS" ]]; then
    # Return a structured JSON block to stdout to stop the tool execution
    REASON="Infinite loop detected: '$TOOL_NAME' has been called with identical arguments $COUNT times in a row. You MUST stop and try a different approach."

    # Print the blocking JSON and exit 0 so Claude Code parses it
    jq -n --arg reason "$REASON" '{"decision": "block", "reason": $reason}'
    exit 0
fi
