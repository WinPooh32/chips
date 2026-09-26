---
name: update-tools
description: Use when updating a tools Go tool dependency to a new version
---
# Update Tools Skill

Follow these steps to update an existing tool in `tools/`.

## Update a Specific Tool

Bump the version in the modfile, then download to update `go.sum`:

```bash
# Edit tools/<tool>-go.mod: change the version line for the main module
# Example: github.com/WinPooh32/searxng-mcp v0.0.2 -> v0.0.3

# Download to update go.sum
go mod download -modfile=tools/<tool>-go.mod
```

Or use `go get` to bump directly:

```bash
go get -tool -modfile=tools/<tool>-go.mod <module>@<version>
```

For latest:

```bash
go get -tool -modfile=tools/<tool>-go.mod <module>@latest
```

## Update All Tools

```bash
for mod in tools/*-go.mod; do
    go get -tool -modfile="$mod" -u
done
```

## Verify

1. Run `make install/tools` to confirm all dependencies download.
2. Run the relevant `make` target to confirm the tool works.
