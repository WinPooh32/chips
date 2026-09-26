---
name: research
description: Use when you need to conduct analysis or study something
---
# Research Skill

You MUST follow these instructions:

<research_instructions>

## Research

1. Review the list of previous researches in .agents/skills/research/index.md

  - Study researches relevant to the task.

2. If necessary, add an entry to the research index file
   .agents/skills/research/index.md

  - Format:

    ```md
    - [research name](.agents/skills/research/researches/<name>.md) — <date (year, month, day, e.g. 2026-09-26)> — <brief description of no more than 20 words>
    ```

  - Add new entries to the top of the list (above existing entries).
  - Update index entries order using command: `bash .agents/skills/research/scripts/sort-index.sh`.

3. Create a new research file in .agents/skills/research/researches/<name>.md

  - Briefly describe the original question.
  - Make updates to the research file during the study process, not just at
    the final summary.
  - On the final change to the research file, update the date in its index
    entry to the date (year, month, day) of that final change.

NOTICE: When delegating to an agent, require the research to be updated
incrementally, not all at once! Also, use code symbols instead of copying
structures or queries from the code.
</research_instructions>
