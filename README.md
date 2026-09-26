# chips

chips is a local task tracker. It tracks tasks as plain markdown files in a
`.chips/` directory inside your project — no server, no database, no git
integration. One small Go binary.

## Base principles

- State is the directory: `.chips/{open,in_progress,blocked,done}/`.
- One markdown file per issue, small flat frontmatter, free body.
- `blocked/` is automatic: derived from `blocked-by` dependencies.
- Never touches git — commit or ignore `.chips/` as you like.

## How to use

```console
$ chips create "Fix login timeout" --type bug --desc "repro in docs"
created a1b2 Fix login timeout

$ chips create "Add auth" --type feature --desc "jwt"
created c3d4 Add auth

$ chips dep add a1b2 c3d4 # a1b2 moves to blocked/
a1b2 is blocked by c3d4

$ chips ready # only unblocked open issues
c3d4  feature  Add auth

$ chips claim c3d4 # open/ -> in_progress/
claimed c3d4

$ chips close c3d4 --reason "done" # -> done/, a1b2 back to open/
closed c3d4

$ chips show a1b2
id: a1b2
title: Fix login timeout
type: bug
created-at: 2026-09-26T10:00:00Z
blocked-by: c3d4
state: ready
repro in docs
```

All commands:

```console
chips ready
chips create "title" --type <bug|task|feature|epic> --desc "..." [--parent <id>]
chips show <id>
chips claim <id>
chips status <id> open|in_progress|done [--note "..."]
chips dep add <id> <dep-id>
chips dep rm <id> <dep-id>
chips close <id>... --reason "..."
```

Set `CHIPS_ROOT` to point at a store directory other than `./.chips`.

## Installation

```console
go install github.com/WinPooh32/chips/cmd/chips@latest
```

The binary lands in `$(go env GOPATH)/bin`.
