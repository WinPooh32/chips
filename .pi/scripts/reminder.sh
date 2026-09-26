#!/usr/bin/env bash
MESSAGE=$(jq -Rs '.' << 'EOF'
<system-reminder>
You MUST follow these instructions:
1. Think about available skills. Specify which ones to connect for this task.
2. Analyze the user's task.
3. If there are contradictions, ask clarifying questions.
4. Load relevant skills.
5. You MUST delegate the task to a relevant subagent.
</system-reminder>
EOF
)

echo "{\"additionalContext\": $MESSAGE}"