# AGENTS.md

Behavioral guidelines to reduce common LLM coding mistakes. Merge with project-specific instructions as needed.

**Tradeoff:** These guidelines bias toward caution over speed. For trivial tasks, use judgment.

## 1. Think Before Coding

### Don't assume. Don't hide confusion. Surface tradeoffs

Before implementing:

- State your assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them - don't pick silently.
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, stop. Name what's confusing. Ask.

## 2. Simplicity First

### Minimum code that solves the problem. Nothing speculative

- No features beyond what was asked.
- No abstractions for single-use code.
- No "flexibility" or "configurability" that wasn't requested.
- No error handling for impossible scenarios.
- If you write 200 lines and it could be 50, rewrite it.

Ask yourself: "Would a senior engineer say this is overcomplicated?" If yes, simplify.

## 3. Surgical Changes

### Touch only what you must. Clean up only your own mess

When editing existing code:

- Don't "improve" adjacent code, comments, or formatting.
- Don't refactor things that aren't broken.
- Match existing style, even if you'd do it differently.
- If you notice unrelated dead code, mention it - don't delete it.

When your changes create orphans:

- Remove imports/variables/functions that YOUR changes made unused.
- Don't remove pre-existing dead code unless asked.

The test: Every changed line should trace directly to the user's request.

## 4. Goal-Driven Execution

### Define success criteria. Loop until verified

Transform tasks into verifiable goals:

- "Add validation" → "Write tests for invalid inputs, then make them pass"
- "Fix the bug" → "Write a test that reproduces it, then make it pass"
- "Refactor X" → "Ensure tests pass before and after"

For multi-step tasks, state a brief plan:

```text
1. [Step] → verify: [check]
2. [Step] → verify: [check]
3. [Step] → verify: [check]
```

Strong success criteria let you loop independently. Weak criteria ("make it work") require constant clarification.

**These guidelines are working if:** fewer unnecessary changes in diffs, fewer rewrites,
and clarifying questions come before implementation.

## 5. Tasks tracking

This project uses bd (beads) for issue tracking.
- Do not create MEMORY.md files.
- Do not use markdown TODO lists for task tracking.

## Core Rules
- **Default**: Use beads for ALL task tracking (`make issue/create`, `make issue/ready`, `make issue/close`)
- **Prohibited**: Do NOT use TodoWrite, TaskCreate, or markdown files for task tracking
- **Workflow**: Create beads issue BEFORE writing code, claim it when starting
- Persistence you don't need beats lost context
- Git authority: no git operations in this context
- Git workflow: stealth mode (no git ops)
- Session management: check `make issue/ready` for available work

### Creating & Updating

#### New issue
```bash
make issue/create TITLE="<issue name>" DESCRIPTION="<issue description>" TYPE="<task|bug|feature>"
```

Newlines in DESCRIPTION: use literal \n (the recipe converts it); keep the value on one line.

#### Hierarchical child (task under epic, subtask under task; inherits parent labels)
```bash
make issue/create-child TITLE="<issue name>" DESCRIPTION="<issue description>" TYPE="<task|bug|feature>" PARENT="<id>"
```

### Common Workflows

#### Starting work
```bash
make issue/ready                 # Find available work
make issue/show ISSUE_ID="<id>"   # Review issue details
make issue/claim ISSUE_ID="<id>"  # Claim it
```

#### Completing work
```bash
make issue/close ISSUES="<id_1> <id_2> ..." REASON="reason to close"    # Close all completed issues at once
```

#### Set blocked by another issue
```bash
make issue/set-status ISSUE_ID="<id>" STATUS="blocked" REASON="<reason why its blocked by another issue>"
```

### Issue Templates
When creating issues, include the required sections for the issue type:

**bug** — requires `## Steps to Reproduce` and `## Acceptance Criteria`
```markdown
## Steps to Reproduce
1. Do this
2. Do that

## Acceptance Criteria
- Bug is fixed
```

**task** — requires `## Acceptance Criteria`
```markdown
Description of the work

## Acceptance Criteria
- [ ] Task complete
```

**feature** — requires `## Acceptance Criteria`
```markdown
Description of the feature

## Acceptance Criteria
- [ ] Feature works as expected
```

**epic** — requires `## Success Criteria` (or `## Acceptance Criteria`)
```markdown
Big project description

## Success Criteria
- Project ships
- Users happy
```

**decision** — requires `## Decision`, `## Rationale`, `## Alternatives Considered`
```markdown
## Decision
What was decided

## Rationale
Why this option was chosen

## Alternatives Considered
List of alternatives and why they were rejected
```

**spike** — requires `## Goal`, `## Findings`
```markdown
## Goal
What question does this spike answer?

## Findings
What was learned?
```

**chore** / **message** / **molecule** — no required sections

## 6. Validation

### Run tests
```bash
make test
```

### Run linter
```bash
make lint
```

### Testing coverage
Minimal value is 60%

```bash
make test/coverage
```

### Mutation testing
Minimal mutation testing score is 0.60

#### Run selected package mutation testing
```bash
make test/mutest-pkg PACKAGE="github.com/WinPooh32/psychic-waddle/any/package"
```

7. DO NOT EDIT

You are not allowed to edit these files without human's permission:
- .env
- .go-arch-lint.yml
- .golangci.yml
- .mdsmith.yml
- .mutesting.yml
- tools/scripts/scc-threshold.sh
- tools/scripts/find-untested-pkgs.sh
- tools/scripts/mutest-tested-pkgs.sh
