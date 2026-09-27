package issues_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/WinPooh32/chips/cmd/chips/internal/lib/issues"
)

var idPattern = regexp.MustCompile(`^[a-hjkmnp-z2-9]{4}$`)

// newTestStore returns a store rooted in a fresh temp dir.
func newTestStore(t *testing.T) *issues.Store {
	t.Helper()
	return issues.NewStore(t.TempDir())
}

// createIssue creates an issue of the given type and returns it.
func createIssue(t *testing.T, s *issues.Store, title, typ string) *issues.Issue {
	t.Helper()

	iss, err := s.Create(title, typ, "body of "+title, "")
	require.NoError(t, err)

	return iss
}

func TestSlug(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		title string
		want  string
	}{
		{name: "plain words", title: "Fix login timeout", want: "fix-login-timeout"},
		{name: "punctuation", title: "Fix: login (urgent)!!", want: "fix-login-urgent"},
		{name: "digits", title: "A1 B2", want: "a1-b2"},
		{name: "no letters", title: "!!! ///", want: "issue"},
		{name: "trailing separator", title: "Done -", want: "done"},
		{name: "control chars stripped", title: "abc\x00def", want: "abcdef"},
		{name: "newline and tab stripped", title: "hello\n\tworld", want: "helloworld"},
		{name: "long title capped", title: strings.Repeat("a", 100), want: strings.Repeat("a", 60)},
		{name: "cap trims trailing dash",
			title: strings.Repeat("a", 59) + " " + strings.Repeat("b", 50),
			want:  strings.Repeat("a", 59)},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, issues.Slug(tt.title))
		})
	}
}

func TestCreate(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	s := issues.NewStore(root)

	iss, err := s.Create("Fix login timeout", issues.TypeBug, "steps to reproduce", "")
	require.NoError(t, err)
	require.Regexp(t, idPattern, iss.ID)

	path := filepath.Join(root, "open", iss.ID+"-fix-login-timeout.md")
	data, err := os.ReadFile(path)
	require.NoError(t, err)

	text := string(data)
	require.True(t, strings.HasPrefix(text, "---\nid: "+iss.ID+"\n"))
	require.Contains(t, text, "title: Fix login timeout\n")
	require.Contains(t, text, "type: bug\n")
	require.Contains(t, text, "steps to reproduce")
	require.NotContains(t, text, "parent:")

	got, st, body, err := s.Read(iss.ID)
	require.NoError(t, err)
	require.Equal(t, issues.Open, st)
	require.Equal(t, "steps to reproduce", body)
	require.Equal(t, iss.ID, got.ID)
	require.False(t, got.CreatedAt.IsZero())
}

func TestCreateWithParent(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)
	parent := createIssue(t, s, "Ship the thing", issues.TypeEpic)

	iss, err := s.Create("Child task", issues.TypeTask, "desc", parent.ID)
	require.NoError(t, err)

	got, st, _, err := s.Read(iss.ID)
	require.NoError(t, err)
	require.Equal(t, parent.ID, got.Parent)
	require.Equal(t, issues.Open, st)

	// The parent is blocked by the new child.
	p, pst, _, err := s.Read(parent.ID)
	require.NoError(t, err)
	require.Equal(t, issues.Blocked, pst)
	require.Equal(t, []string{iss.ID}, p.BlockedBy)
}

func TestCreateWithParentUnblocksParent(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)
	parent := createIssue(t, s, "Ship the thing", issues.TypeEpic)
	c1, err := s.Create("Child task", issues.TypeTask, "desc", parent.ID)
	require.NoError(t, err)
	c2, err := s.Create("Second child", issues.TypeTask, "desc", parent.ID)
	require.NoError(t, err)

	require.NoError(t, s.Close([]string{c1.ID}, "done"))

	// One child still open: parent stays blocked.
	_, st, _, err := s.Read(parent.ID)
	require.NoError(t, err)
	require.Equal(t, issues.Blocked, st)

	require.NoError(t, s.Close([]string{c2.ID}, "done"))

	// All children done: parent unblocks.
	_, st, _, err = s.Read(parent.ID)
	require.NoError(t, err)
	require.Equal(t, issues.Open, st)
}

func TestCreateWithDoneParent(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)
	parent := createIssue(t, s, "Ship the thing", issues.TypeEpic)
	require.NoError(t, s.Close([]string{parent.ID}, "done"))

	_, err := s.Create("Child task", issues.TypeTask, "desc", parent.ID)
	require.Error(t, err)
}

func TestCreateInvalidType(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)

	_, err := s.Create("Bad type", "chore", "", "")
	require.Error(t, err)
}

func TestFrontmatterRoundTrip(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	s := issues.NewStore(root)

	iss, err := s.Create("Fix: login (urgent)!!", issues.TypeBug, "some body", "")
	require.NoError(t, err)

	path := filepath.Join(root, "open", iss.ID+"-fix-login-urgent.md")
	data, err := os.ReadFile(path)
	require.NoError(t, err)

	require.Contains(t, string(data), "title: 'Fix: login (urgent)!!'\n")

	got, st, body, err := s.Read(iss.ID)
	require.NoError(t, err)
	require.Equal(t, issues.Open, st)
	require.Equal(t, "Fix: login (urgent)!!", got.Title)
	require.Equal(t, issues.TypeBug, got.Type)
	require.Equal(t, "some body", body)
}

func TestReady(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	s := issues.NewStore(root)
	iss := createIssue(t, s, "Ready one", issues.TypeTask)

	dangling := "---\n" +
		"id: zz99\n" +
		"title: Dangling dep\n" +
		"type: task\n" +
		"created-at: 2026-01-01T00:00:00Z\n" +
		"blocked-by: [aa11]\n" +
		"---\nbody\n"
	name := filepath.Join(root, "open", "zz99-dangling-dep.md")
	require.NoError(t, os.WriteFile(name, []byte(dangling), 0o600))

	got, warns, err := s.Ready()
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, iss.ID, got[0].ID)
	require.Len(t, warns, 1)
	require.Contains(t, warns[0], "zz99")
	require.Contains(t, warns[0], "aa11")
}

func TestClaim(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)
	iss := createIssue(t, s, "Claim me", issues.TypeTask)

	require.NoError(t, s.Claim(iss.ID))

	got, st, _, err := s.Read(iss.ID)
	require.NoError(t, err)
	require.Equal(t, issues.InProgress, st)
	require.NotNil(t, got.ClaimedAt)

	require.Error(t, s.Claim(iss.ID))
}

func TestStatus(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)
	iss := createIssue(t, s, "Ship feature", issues.TypeFeature)

	require.NoError(t, s.Status(iss.ID, issues.Done, "shipped"))

	_, st, body, err := s.Read(iss.ID)
	require.NoError(t, err)
	require.Equal(t, issues.Done, st)
	require.Contains(t, body, "shipped")
	require.Contains(t, body, "## ")

	require.Error(t, s.Status(iss.ID, issues.Done, "again"))
}

func TestDepAdd(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)
	dep := createIssue(t, s, "Base lib", issues.TypeTask)
	iss := createIssue(t, s, "On top", issues.TypeTask)

	require.NoError(t, s.DepAdd(iss.ID, dep.ID))

	_, st, _, err := s.Read(iss.ID)
	require.NoError(t, err)
	require.Equal(t, issues.Blocked, st)

	got, _, _, err := s.Read(iss.ID)
	require.NoError(t, err)
	require.Equal(t, []string{dep.ID}, got.BlockedBy)

	require.Error(t, s.DepAdd(iss.ID, dep.ID))
	require.Error(t, s.DepAdd(iss.ID, iss.ID))
	require.Error(t, s.DepAdd(iss.ID, "zz99"))
}

func TestCloseUnblocksDependents(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)
	dep := createIssue(t, s, "Blocking work", issues.TypeTask)
	iss := createIssue(t, s, "Waiting work", issues.TypeTask)
	require.NoError(t, s.DepAdd(iss.ID, dep.ID))

	require.NoError(t, s.Close([]string{dep.ID}, "done"))

	_, st, _, err := s.Read(iss.ID)
	require.NoError(t, err)
	require.Equal(t, issues.Open, st)
}

func TestCloseKeepsBlockedWhenPartial(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)
	dep := createIssue(t, s, "First dep", issues.TypeTask)
	other := createIssue(t, s, "Second dep", issues.TypeTask)
	iss := createIssue(t, s, "Waiting work", issues.TypeTask)
	require.NoError(t, s.DepAdd(iss.ID, dep.ID))
	require.NoError(t, s.DepAdd(iss.ID, other.ID))

	require.NoError(t, s.Close([]string{dep.ID}, "done"))

	_, st, _, err := s.Read(iss.ID)
	require.NoError(t, err)
	require.Equal(t, issues.Blocked, st)
}

func TestDepRemoveUnblocks(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)
	dep := createIssue(t, s, "First dep", issues.TypeTask)
	other := createIssue(t, s, "Second dep", issues.TypeTask)
	iss := createIssue(t, s, "Waiting work", issues.TypeTask)
	require.NoError(t, s.DepAdd(iss.ID, dep.ID))
	require.NoError(t, s.DepAdd(iss.ID, other.ID))

	require.NoError(t, s.Close([]string{dep.ID}, "done"))

	require.NoError(t, s.DepRemove(iss.ID, other.ID))

	_, st, _, err := s.Read(iss.ID)
	require.NoError(t, err)
	require.Equal(t, issues.Open, st)

	require.Error(t, s.DepRemove(iss.ID, other.ID))
}

func TestCloseMultiple(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)
	a := createIssue(t, s, "First close", issues.TypeTask)
	b := createIssue(t, s, "Second close", issues.TypeTask)

	require.NoError(t, s.Close([]string{a.ID, b.ID}, "reason text"))

	_, stA, bodyA, err := s.Read(a.ID)
	require.NoError(t, err)
	require.Equal(t, issues.Done, stA)
	require.Contains(t, bodyA, "## Closed: reason text (")

	_, stB, _, err := s.Read(b.ID)
	require.NoError(t, err)
	require.Equal(t, issues.Done, stB)
}

func TestListEmptyStore(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)

	got, err := s.List(issues.Open)
	require.NoError(t, err)
	require.Empty(t, got)
}

func TestReadMissing(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)

	require.ErrorIs(t, s.Claim("zz99"), issues.ErrNotFound)
}

func TestParseMissingRequiredFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		omit string
	}{
		{name: "missing id", omit: "id"},
		{name: "missing title", omit: "title"},
		{name: "missing type", omit: "type"},
		{name: "missing created-at", omit: "created-at"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			s := issues.NewStore(root)

			var lines []string
			if tt.omit != "id" {
				lines = append(lines, "id: zz11")
			}

			if tt.omit != "title" {
				lines = append(lines, "title: Some title")
			}

			if tt.omit != "type" {
				lines = append(lines, "type: task")
			}

			if tt.omit != "created-at" {
				lines = append(lines, "created-at: 2026-01-01T00:00:00Z")
			}

			content := "---\n" + strings.Join(lines, "\n") + "\n---\nbody\n"
			name := filepath.Join(root, "open", "zz11-some-title.md")
			require.NoError(t, os.MkdirAll(filepath.Join(root, "open"), 0o700))
			require.NoError(t, os.WriteFile(name, []byte(content), 0o600))

			_, _, _, err := s.Read("zz11")
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.omit)
		})
	}
}

func TestSaveAtomicReplace(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	s := issues.NewStore(root)
	iss := createIssue(t, s, "Atomic save", issues.TypeTask)

	require.NoError(t, s.Save(iss, "updated body"))

	path := filepath.Join(root, "open", iss.ID+"-atomic-save.md")
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Contains(t, string(data), "updated body")

	entries, err := os.ReadDir(filepath.Join(root, "open"))
	require.NoError(t, err)
	require.Len(t, entries, 1)
}

func TestCreateNoDuplicateIDs(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)

	seen := make(map[string]bool)

	for range 40 {
		iss := createIssue(t, s, "Collision check", issues.TypeTask)
		require.NotContains(t, seen, iss.ID)
		seen[iss.ID] = true
	}
}
