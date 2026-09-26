#!/bin/bash

# Block piping download tools (curl, wget, aria2c) to shell interpreters

# Read hook input from stdin
INPUT=$(cat)

# Only process PreToolUse events
HOOK_EVENT=$(echo "$INPUT" | jq -r '.hook_event_name // "unknown"')
if [[ "$HOOK_EVENT" != "PreToolUse" ]]; then
    exit 0
fi

# Only check Bash tool calls
TOOL_NAME=$(echo "$INPUT" | jq -r '.tool_name // "none"')
if [[ "$TOOL_NAME" != "Bash" ]]; then
    exit 0
fi

# Extract the full command string
COMMAND=$(echo "$INPUT" | jq -r '.tool_input.command // ""')

# Block: curl/wget/aria2c piped to shell interpreters
if echo "$COMMAND" | grep -qP '(^|\s)(curl|wget|aria2c)\s+.*\|\s*(bash|sh|zsh|fish|python|python3|node|perl|ruby|ruby)\b'; then
    jq -n '{"decision": "block", "reason": "Running scripts from the web is forbidden by the user. You MUST stop. DO NOT try alternative approaches to bypass this block."}'
    exit 0
fi

# Allow the command
exit 0
