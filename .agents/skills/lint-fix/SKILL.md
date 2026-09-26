---
name: lint-fix
description: Use when fixing linter errors in the project
---
# Lint Fix Skill

Instructions for correctly fixing linter errors in this project.

## Golden Rule

Fix the root cause. This project forbids inline linter suppression comments.

If a linter fires, fix the code or the config — never silence it.

## Linters Overview

| Linter        | Target       | Config                  | Purpose                                  |
| ------------- | ------------ | ----------------------- | ---------------------------------------- |
| golangci-lint | `lint/go`    | `.golangci.yml`         | Go static analysis (90+ enabled linters) |
| typos         | `lint/typos` | `.typos.toml`           | Spell-checking                           |
| mdsmith       | `lint/md`    | `.mdsmith.yml`          | Markdown linting                         |
| go-arch-lint  | `lint/arch`  | `misc/go-arch-lint.yml` | Architecture constraints                 |

Run `make lint` to check all linters. Run `make fmt` to auto-fix formatting.

## Workflow

1. Run `make lint` to see the errors.
2. Read the error — it tells you file, line, linter name, and rule.
3. Fix the code to satisfy the linter.
4. If the rule is too strict for a legitimate pattern, add a config-level exclusion.
5. Run `make lint` again to verify.

## Common Linters and Fixes

### `forbidigo` — Forbidden patterns

Pattern `^(fmt\.Print(|f|ln)|print|println)$` blocks `fmt.Print*`, `print`, and `println`.

```go
// Before (forbidden)
fmt.Println("hello")

// After — use slog
slog.Println("hello")
```

### `revive` — Style rules

Includes `package-comments`, `var-naming`, `exported`, etc.

```go
// Before (missing package comment)
package main

// After
// Package main provides the entry point for the application.
package main
```

### `wrapcheck` — Unwrapped errors

External errors must be wrapped with `fmt.Errorf("...: %w", err)`.

```go
// Before
return err

// After
return fmt.Errorf("failed to connect: %w", err)
```

Note: the `context` package is excluded in `.golangci.yml`.

### `mnd` — Magic numbers

Extract numeric literals to named constants.

```go
// Before
time.Sleep(5 * time.Second)

// After
const timeout = 5 * time.Second
time.Sleep(timeout)
```

### `exhaustruct` — Missing struct fields

Set all fields explicitly.

```go
// Before
config := Config{Host: "localhost"}

// After
config := Config{Host: "localhost", Port: 0, Timeout: 0}

// Or better alternative
config := NewConfigShort("localhost")
```

### `gochecknoglobals` — Global variables

Avoid package-level variables. When unavoidable (singleton, driver registration), document why.

```go
// Before
var db *sql.DB

// After — move into main() or a constructor
func newApp() (*App, error) {
    db, err := sql.Open("postgres", dsn)
    // ...
}
```

### `gochecknoinits` — Init functions

Avoid `init()`. When required by an external API (driver registration), use a regular function called explicitly.

### `govet` — Suspicious constructs

Fix the actual issue: unreachable code, incorrect format verbs, nil pointer dereference, etc.

### `errcheck` — Unchecked errors

Handle or explicitly ignore errors.

```go
// Before
w.Write(data)

// After
_, err := w.Write(data)
if err != nil {
    return fmt.Errorf("write data: %w", err)
}
```

### `staticcheck` — Deep analysis

These are usually real bugs. Read the message carefully and fix the logic.

### `gosec` — Security issues

Fix the vulnerability. Common patterns:

```go
// Before (gosec: potential SQL injection)
query := fmt.Sprintf("SELECT * FROM users WHERE id = %s", input)

// After
row := db.QueryRow("SELECT * FROM users WHERE id = $1", input)
```

### `contextcheck` — Missing context

Pass context to functions that require it.

```go
// Before
db.Query("SELECT 1")

// After
db.QueryCtx(ctx, "SELECT 1")
```

### `containedctx` — Context struct field order

`context.Context` must be the first field in a struct.

```go
// Before
type Server struct {
    cfg Config
    ctx context.Context
}

// After
type Server struct {
    ctx context.Context
    cfg Config
}
```

### `bodyclose` — HTTP response body

Close response bodies.

```go
// Before
resp, err := http.Get(url)
...

// After
resp, err := http.Get(url)
...
defer resp.Body.Close()
```

## Config-Level Exclusions

When the code is correct, add an exclusion in `.golangci.yml`.
This is better than scattering inline comments:

```yaml
linters:
  exclusions:
    rules:
      - linters: [forbidigo]
        path: _test\.go  # Allow fmt.Println in tests
```

Note: `.golangci.yml` is write-protected in permissions. To change it, ask the user to
edit the file manually.

## typos Suppression

For legitimate words flagged as typos, add to `.typos.toml`:

```toml
[default.extend-words]
"petstore" = "petstore"
```

## Anti-Patterns

| Anti-Pattern                                       | Why It's Bad               | Correct Approach                      |
| -------------------------------------------------- | -------------------------- | ------------------------------------- |
| Removing features to silence linter                | Breaks the program         | Replace with allowed alternative      |
| Using inline suppression comments                  | Forbidden by project hook  | Fix the code or config                |
| Disabling a linter globally for one false positive | Loses the check everywhere | Fix locally or add targeted exclusion |
| Ignoring linter output                             | Technical debt accumulates | Address every finding                 |
| Changing adjacent code "while you're at it"        | Violates surgical changes  | Only touch what the error points to   |

## Remember

- Never remove functionality to satisfy a linter. Replace it.
- Never add inline suppression comments — the hook will block the edit.
- If you're unsure how to fix an error, ask the user before proceeding.
