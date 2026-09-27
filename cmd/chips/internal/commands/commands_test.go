package commands_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/WinPooh32/chips/cmd/chips/internal/commands"
)

// runChips runs the chips command tree (store from CHIPS_ROOT) and returns
// stdout and stderr. It fails the test when the command exits with an error.
func runChips(t *testing.T, args ...string) (string, string) {
	t.Helper()

	cmd := commands.NewRoot()
	cmd.SetArgs(args)

	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)

	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("run chips %v: %v (stderr: %s)", args, err, stderr.String())
	}

	return stdout.String(), stderr.String()
}

// createViaCLI creates an issue and returns its id.
func createViaCLI(t *testing.T, title, typ string) string {
	t.Helper()

	out, _ := runChips(t, "create", title, "--type", typ)
	fields := strings.Fields(out)
	require.NotEmpty(t, fields)
	require.Equal(t, "created", fields[0])

	return fields[1]
}

func TestCreateCommand(t *testing.T) {
	work := t.TempDir()
	t.Setenv("CHIPS_ROOT", filepath.Join(work, ".chips"))

	id := createViaCLI(t, "Fix login timeout", "bug")

	files, err := os.ReadDir(filepath.Join(work, ".chips", "open"))
	require.NoError(t, err)
	require.Len(t, files, 1)
	require.Equal(t, id+"-fix-login-timeout.md", files[0].Name())
}

func TestCreateCommandMultilineDesc(t *testing.T) {
	work := t.TempDir()
	t.Setenv("CHIPS_ROOT", filepath.Join(work, ".chips"))

	out, _ := runChips(t, "create", "Multiline desc", "--type", "task", "--desc", "line one\\nline two")
	fields := strings.Fields(out)
	require.NotEmpty(t, fields)
	require.Equal(t, "created", fields[0])

	files, err := os.ReadDir(filepath.Join(work, ".chips", "open"))
	require.NoError(t, err)
	require.Len(t, files, 1)

	data, err := os.ReadFile(filepath.Join(work, ".chips", "open", files[0].Name()))
	require.NoError(t, err)
	require.Contains(t, string(data), "line one\nline two")
	require.NotContains(t, string(data), "line one\\nline two")
}

func TestReadyCommand(t *testing.T) {
	work := t.TempDir()
	t.Setenv("CHIPS_ROOT", filepath.Join(work, ".chips"))

	a := createViaCLI(t, "First task", "task")
	b := createViaCLI(t, "Second task", "task")

	out, _ := runChips(t, "ready")
	require.Contains(t, out, a)
	require.Contains(t, out, b)
}

func TestShowCommand(t *testing.T) {
	work := t.TempDir()
	t.Setenv("CHIPS_ROOT", filepath.Join(work, ".chips"))

	id := createViaCLI(t, "Show me", "feature")

	out, _ := runChips(t, "show", id)
	require.Contains(t, out, "id: "+id+"\n")
	require.Contains(t, out, "title: Show me\n")
	require.Contains(t, out, "type: feature\n")
	require.Contains(t, out, "state: ready\n")
}

func TestClaimCommand(t *testing.T) {
	work := t.TempDir()
	t.Setenv("CHIPS_ROOT", filepath.Join(work, ".chips"))

	id := createViaCLI(t, "Claim me", "task")

	out, _ := runChips(t, "claim", id)
	require.Equal(t, "claimed "+id+"\n", out)

	files, err := os.ReadDir(filepath.Join(work, ".chips", "in_progress"))
	require.NoError(t, err)
	require.Len(t, files, 1)

	out, _ = runChips(t, "show", id)
	require.Contains(t, out, "state: in progress\n")
	require.Contains(t, out, "claimed-at: ")
}

func TestStatusCommand(t *testing.T) {
	work := t.TempDir()
	t.Setenv("CHIPS_ROOT", filepath.Join(work, ".chips"))

	id := createViaCLI(t, "Ship it", "feature")

	out, _ := runChips(t, "status", id, "done", "--note", "shipped")
	require.Equal(t, "moved "+id+" to done\n", out)

	files, err := os.ReadDir(filepath.Join(work, ".chips", "done"))
	require.NoError(t, err)
	require.Len(t, files, 1)

	data, err := os.ReadFile(filepath.Join(work, ".chips", "done", files[0].Name()))
	require.NoError(t, err)
	require.Contains(t, string(data), "shipped")
}

func TestDepAddCommand(t *testing.T) {
	work := t.TempDir()
	t.Setenv("CHIPS_ROOT", filepath.Join(work, ".chips"))

	dep := createViaCLI(t, "Base lib", "task")
	id := createViaCLI(t, "On top", "task")

	out, _ := runChips(t, "dep", "add", id, dep)
	require.Equal(t, id+" is blocked by "+dep+"\n", out)

	files, err := os.ReadDir(filepath.Join(work, ".chips", "blocked"))
	require.NoError(t, err)
	require.Len(t, files, 1)

	out, _ = runChips(t, "show", id)
	require.Contains(t, out, "blocked on: "+dep+"\n")
}

func TestDepRemoveCommand(t *testing.T) {
	work := t.TempDir()
	t.Setenv("CHIPS_ROOT", filepath.Join(work, ".chips"))

	a := createViaCLI(t, "First dep", "task")
	b := createViaCLI(t, "Second dep", "task")
	id := createViaCLI(t, "Waiting", "task")
	runChips(t, "dep", "add", id, a)
	runChips(t, "dep", "add", id, b)
	runChips(t, "close", a, "--reason", "done")

	out, _ := runChips(t, "dep", "rm", id, b)
	require.Equal(t, id+" is no longer blocked by "+b+"\n", out)

	files, err := os.ReadDir(filepath.Join(work, ".chips", "open"))
	require.NoError(t, err)
	require.Len(t, files, 2)
}

func TestCloseCommand(t *testing.T) {
	work := t.TempDir()
	t.Setenv("CHIPS_ROOT", filepath.Join(work, ".chips"))

	a := createViaCLI(t, "Close one", "task")
	b := createViaCLI(t, "Close two", "task")

	out, _ := runChips(t, "close", a, b, "--reason", "wrapped up")
	require.Contains(t, out, "closed "+a+"\n")
	require.Contains(t, out, "closed "+b+"\n")

	files, err := os.ReadDir(filepath.Join(work, ".chips", "done"))
	require.NoError(t, err)
	require.Len(t, files, 2)
}

func TestCloseCommandMixedBatch(t *testing.T) {
	work := t.TempDir()
	t.Setenv("CHIPS_ROOT", filepath.Join(work, ".chips"))

	a := createViaCLI(t, "Close valid", "task")
	b := createViaCLI(t, "Close other", "task")

	cmd := commands.NewRoot()
	cmd.SetArgs([]string{"close", a, "zz99", b, "yy88", "--reason", "wrapped up"})

	err := cmd.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "zz99: issue not found")
	require.Contains(t, err.Error(), "yy88: issue not found")

	files, err := os.ReadDir(filepath.Join(work, ".chips", "done"))
	require.NoError(t, err)
	require.Len(t, files, 2)
}

func TestReadyDanglingWarning(t *testing.T) {
	work := t.TempDir()
	t.Setenv("CHIPS_ROOT", filepath.Join(work, ".chips"))

	id := createViaCLI(t, "Real issue", "task")

	dangling := "---\n" +
		"id: zz99\n" +
		"title: Dangling dep\n" +
		"type: task\n" +
		"created-at: 2026-01-01T00:00:00Z\n" +
		"blocked-by: [aa11]\n" +
		"---\nbody\n"
	name := filepath.Join(work, ".chips", "open", "zz99-dangling-dep.md")
	require.NoError(t, os.WriteFile(name, []byte(dangling), 0o600))

	out, errOut := runChips(t, "ready")
	require.Contains(t, out, id)
	require.NotContains(t, out, "zz99")
	require.Contains(t, errOut, "zz99")
	require.Contains(t, errOut, "aa11")
}

func TestVersionFlag(t *testing.T) {
	t.Parallel()

	out, _ := runChips(t, "--version")
	require.Equal(t, "chips version dev\n", out)
}

func TestErrorExit(t *testing.T) {
	work := t.TempDir()
	t.Setenv("CHIPS_ROOT", filepath.Join(work, ".chips"))

	cmd := commands.NewRoot()
	cmd.SetArgs([]string{"claim", "zz99"})

	err := cmd.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "issue not found")
}
