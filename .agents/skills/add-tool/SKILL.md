---
name: add-tool
description: Use when adding a new Go tool dependency to the project
---
# Add Tool Skill

Follow these steps to add a new tool to the project.

## 1. Create Tool Modfile

Create a dedicated `*-go.mod` file in `tools/`:

```bash
cd tools
go mod init -modfile=<tool>-go.mod github.com/WinPooh32/psychic-waddle/tools
go get -tool -modfile=<tool>-go.mod <module>@latest
```

Example for `gofumpt`:
```bash
go mod init -modfile=gofumpt-go.mod github.com/WinPooh32/psychic-waddle/tools
go get -tool -modfile=gofumpt-go.mod mvdan.cc/gofumpt@latest
```

## 2. Integrate Into Makefile

Add a target in `Makefile` using the `go tool` command:

```makefile
target/name:
	@echo "Description"
	@go tool -modfile=tools/<tool>-go.mod <binary> [args]
```

**Important:** Test the command manually first. Some tools don't support `./...` pattern
(e.g., `gofumpt` uses `.` instead).

## 3. Verify

- Run `make install/tools` to confirm dependencies download.
- Run the new target to confirm it works.
