// Package issues implements the chips storage layer: markdown files with
// flat frontmatter, organized into status directories under a .chips root.
package issues

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Issue is a single tracked task.
type Issue struct {
	ID        string
	Title     string
	Type      string
	CreatedAt time.Time
	Parent    string
	BlockedBy []string
	ClaimedAt *time.Time
}

// Status is the state directory an issue lives in.
type Status string

const (
	// TypeBug is the bug issue type.
	TypeBug = "bug"
	// TypeTask is the task issue type.
	TypeTask = "task"
	// TypeFeature is the feature issue type.
	TypeFeature = "feature"
	// TypeEpic is the epic issue type.
	TypeEpic = "epic"
)

const (
	// Open means the issue is ready to start.
	Open Status = "open"
	// InProgress means the issue is claimed and actively worked on.
	InProgress Status = "in_progress"
	// Blocked means the issue has at least one unfinished dependency.
	Blocked Status = "blocked"
	// Done means the issue is closed.
	Done Status = "done"
)

const (
	idLen        = 4
	idAlphabet   = "abcdefghjkmnpqrstuvwxyz23456789"
	idAttempts   = 10
	minQuotedLen = 2
)

// Statuses lists every status directory in fixed order.
func Statuses() []Status {
	return []Status{Open, InProgress, Blocked, Done}
}

// ValidType reports whether t is one of the supported issue types.
func ValidType(t string) bool {
	switch t {
	case TypeBug, TypeTask, TypeFeature, TypeEpic:
		return true
	default:
		return false
	}
}

// Slug renders the kebab-case filename slug of a title.
func Slug(title string) string {
	var b strings.Builder

	prevDash := false

	for _, r := range strings.ToLower(title) {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)

			prevDash = false
		default:
			if !prevDash && b.Len() > 0 {
				b.WriteByte('-')

				prevDash = true
			}
		}
	}

	s := strings.TrimRight(b.String(), "-")
	if s == "" {
		return "issue"
	}

	return s
}

// render serializes the issue as a markdown file with flat frontmatter.
func render(iss *Issue, body string) []byte {
	var b bytes.Buffer
	b.WriteString("---\n")
	b.WriteString("id: " + iss.ID + "\n")
	b.WriteString("title: " + formatTitle(iss.Title) + "\n")
	b.WriteString("type: " + iss.Type + "\n")
	b.WriteString("created-at: " + iss.CreatedAt.Format(time.RFC3339) + "\n")

	if iss.Parent != "" {
		b.WriteString("parent: " + iss.Parent + "\n")
	}

	if len(iss.BlockedBy) > 0 {
		b.WriteString("blocked-by: [" + strings.Join(iss.BlockedBy, ", ") + "]\n")
	}

	if iss.ClaimedAt != nil {
		b.WriteString("claimed-at: " + iss.ClaimedAt.Format(time.RFC3339) + "\n")
	}

	b.WriteString("---\n")
	b.WriteString(body)

	return b.Bytes()
}

// parse splits a file into its frontmatter issue and markdown body.
func parse(data []byte) (*Issue, string, error) {
	const marker = "---\n"
	if !bytes.HasPrefix(data, []byte(marker)) {
		return nil, "", errors.New("missing frontmatter")
	}

	rest := data[len(marker):]

	end := bytes.Index(rest, []byte("\n---\n"))
	if end < 0 {
		return nil, "", errors.New("unterminated frontmatter")
	}

	iss := new(Issue)
	if err := parseFields(rest[:end], iss); err != nil {
		return nil, "", err
	}

	body := string(rest[end+len("\n---\n"):])

	return iss, body, nil
}

// parseFields fills iss from flat frontmatter lines.
func parseFields(data []byte, iss *Issue) error {
	for _, line := range strings.Split(string(data), "\n") {
		k, v, found := strings.Cut(line, ":")
		if !found {
			return fmt.Errorf("bad frontmatter line %q", line)
		}

		if err := setField(iss, strings.TrimSpace(k), strings.TrimSpace(v)); err != nil {
			return err
		}
	}

	return nil
}

// setField assigns one frontmatter key to the issue.
func setField(iss *Issue, k, v string) error {
	switch k {
	case "id":
		iss.ID = v
	case "title":
		iss.Title = unquote(v)
	case "type":
		iss.Type = v
	case "parent":
		iss.Parent = v
	case "created-at", "claimed-at":
		return setTime(iss, k, v)
	case "blocked-by":
		return setDeps(iss, v)
	default:
		// Unknown keys are ignored for forward compatibility.
	}

	return nil
}

// setTime parses an RFC3339 timestamp field.
func setTime(iss *Issue, k, v string) error {
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return fmt.Errorf("parse %s: %w", k, err)
	}

	if k == "created-at" {
		iss.CreatedAt = t
		return nil
	}

	iss.ClaimedAt = &t

	return nil
}

// setDeps parses the blocked-by list field.
func setDeps(iss *Issue, v string) error {
	deps, err := parseList(v)
	if err != nil {
		return err
	}

	iss.BlockedBy = deps

	return nil
}

// unquote removes YAML single quotes from a title value.
func unquote(v string) string {
	if len(v) < minQuotedLen || v[0] != '\'' || v[len(v)-1] != '\'' {
		return v
	}

	return strings.ReplaceAll(v[1:len(v)-1], "''", "'")
}

// formatTitle writes a title as a plain or single-quoted YAML scalar.
func formatTitle(t string) string {
	if isPlainTitle(t) {
		return t
	}

	return "'" + strings.ReplaceAll(t, "'", "''") + "'"
}

// isAlphaNum reports whether r is an ASCII letter or digit.
func isAlphaNum(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}

// isPlainRune reports whether r is allowed in a plain YAML scalar title.
func isPlainRune(r rune, first bool) bool {
	if first {
		return isAlphaNum(r)
	}

	return isAlphaNum(r) || r == ' ' || r == '-' || r == '_' || r == '.'
}

// isPlainTitle reports whether t is safe as an unquoted YAML scalar.
func isPlainTitle(t string) bool {
	if t == "" {
		return false
	}

	for i, r := range t {
		if !isPlainRune(r, i == 0) {
			return false
		}
	}

	return true
}

// parseList parses a YAML flow list such as [a1b2, c3d4].
func parseList(v string) ([]string, error) {
	if v == "" {
		return nil, nil
	}

	if v[0] != '[' || v[len(v)-1] != ']' {
		return nil, fmt.Errorf("bad list %q", v)
	}

	inner := strings.TrimSpace(v[1 : len(v)-1])
	if inner == "" {
		return nil, nil
	}

	parts := strings.Split(inner, ",")

	out := make([]string, 0, len(parts))
	for _, p := range parts {
		out = append(out, strings.TrimSpace(p))
	}

	return out, nil
}

// randomID generates a random 4-char issue id.
func randomID() (string, error) {
	var buf [idLen]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", fmt.Errorf("read random id: %w", err)
	}

	out := make([]byte, idLen)
	for i := range buf {
		out[i] = idAlphabet[buf[i]%uint8(len(idAlphabet))]
	}

	return string(out), nil
}
