---
name: adr
description: Use when creating a new Architecture Decision Record (ADR)
---
# ADR Skill

Follow these steps to add a new Architecture Decision Record.

## 1. Generate the Filename

Use a 4-digit zero-padded sequence number + short descriptive slug.
Take the highest existing number, increment it, zero-pad to 4 digits
(start at `0001`). Append a kebab-case slug that summarizes the
decision:

```text
docs/adr/0001-use-postgres-for-primary-db.md
docs/adr/0002-adopt-gofumpt-linter.md
```

## 2. Write the ADR

Use the **MADR** format. Every ADR must include:

```markdown
# [Title]

- **Status:** [accepted | proposed | deprecated | superseded]
- **Date:** [YYYY-MM-DD]

## Context
What problem or requirement motivates this decision? List constraints.

## Options Considered
- **Option A** — Brief description, pros and cons.
- **Option B** — Brief description, pros and cons.

## Decision
What was chosen? One clear sentence.

## Consequences
What becomes easier, harder, or uncertain because of this decision?
Include negative consequences — they're valuable, not failures.
```

## 3. Rules

- **Be brief.** ADRs should be a couple of pages max, not architecture treatises.
- **Record the decision, not the debate.** Capture what was decided and why, not every argument.
- **Honest consequences.** List trade-offs and downsides alongside upsides.
- **Never edit accepted ADRs.** If circumstances change, write a new ADR that
  **supersedes** the old one and set the old status to `superseded`.
- **Status lifecycle:** `proposed` → `accepted` → (optionally) `superseded` or
  `deprecated`. Start as `accepted` if the decision is already made.

## 4. Verify

- File lives in `docs/adr/` and follows the naming convention.
- Filename is lexicographically sortable (zero-padded number ensures this).
- The ADR is self-contained — a reader doesn't need to open other files to understand it.
