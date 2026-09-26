package issues

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Store reads and writes issues under a .chips root directory.
type Store struct {
	root string
}

const (
	permDir  = 0o700
	permFile = 0o600
)

// ErrNotFound is returned when an issue id matches no file.
var ErrNotFound = errors.New("issue not found")

// NewStore creates a Store rooted at dir.
func NewStore(dir string) *Store {
	return &Store{root: dir}
}

// Create writes a new issue into open/ and returns it. An empty parent is
// omitted from the frontmatter.
func (s *Store) Create(title, typ, desc, parent string) (*Issue, error) {
	if !ValidType(typ) {
		return nil, fmt.Errorf("invalid type %q", typ)
	}

	id, err := s.uniqueID()
	if err != nil {
		return nil, err
	}

	iss := new(Issue)
	iss.ID = id
	iss.Title = title
	iss.Type = typ
	iss.CreatedAt = time.Now().UTC()
	iss.Parent = parent

	name := iss.ID + "-" + Slug(iss.Title) + ".md"
	if err := s.writeFile(Open, name, render(iss, desc)); err != nil {
		return nil, err
	}

	return iss, nil
}

// Read loads an issue with its status and markdown body.
func (s *Store) Read(id string) (*Issue, Status, string, error) {
	st, path, err := s.locate(id)
	if err != nil {
		return nil, "", "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", "", fmt.Errorf("read %s: %w", path, err)
	}

	iss, body, err := parse(data)
	if err != nil {
		return nil, "", "", fmt.Errorf("parse %s: %w", path, err)
	}

	return iss, st, body, nil
}

// List returns every issue in the given status directory.
func (s *Store) List(st Status) ([]*Issue, error) {
	entries, err := os.ReadDir(s.dir(st))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("read %s dir: %w", st, err)
	}

	out := make([]*Issue, 0, len(entries))
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".md") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(s.dir(st), e.Name()))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", e.Name(), err)
		}

		iss, _, err := parse(data)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", e.Name(), err)
		}

		out = append(out, iss)
	}

	return out, nil
}

// Ready returns open issues without missing dependencies, plus a warning
// line for every open issue whose dependency file is missing.
func (s *Store) Ready() ([]*Issue, []string, error) {
	open, err := s.List(Open)
	if err != nil {
		return nil, nil, err
	}

	var (
		ready []*Issue
		warns []string
	)

	for _, iss := range open {
		missing, err := s.missingDep(iss)
		if err != nil {
			return nil, nil, err
		}

		if missing != "" {
			msg := fmt.Sprintf("issue %s: dependency %s not found", iss.ID, missing)
			warns = append(warns, msg)

			continue
		}

		ready = append(ready, iss)
	}

	return ready, warns, nil
}

// Claim moves an open issue to in_progress/ and stamps claimed-at.
func (s *Store) Claim(id string) error {
	iss, st, body, err := s.Read(id)
	if err != nil {
		return err
	}

	if st != Open {
		return fmt.Errorf("cannot claim issue %s in %s state", id, st)
	}

	now := time.Now().UTC()

	iss.ClaimedAt = &now
	if err := s.Save(iss, body); err != nil {
		return err
	}

	return s.Move(id, Open, InProgress)
}

// Status moves an issue to a new status, appending a note when given.
func (s *Store) Status(id string, to Status, note string) error {
	iss, st, body, err := s.Read(id)
	if err != nil {
		return err
	}

	if st == to {
		return fmt.Errorf("issue %s is already %s", id, to)
	}

	if note != "" {
		ts := time.Now().UTC().Format(time.RFC3339)
		body = appendNote(body, fmt.Sprintf("## %s: %s", ts, note))
	}

	if err := s.Save(iss, body); err != nil {
		return err
	}

	return s.Move(id, st, to)
}

// Save rewrites the issue file in place, preserving its filename.
func (s *Store) Save(iss *Issue, body string) error {
	_, path, err := s.locate(iss.ID)
	if err != nil {
		return err
	}

	data := render(iss, body)
	if err := os.WriteFile(path, data, permFile); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	return nil
}

// Move relocates an issue file between status directories.
func (s *Store) Move(id string, from, to Status) error {
	st, src, err := s.locate(id)
	if err != nil {
		return err
	}

	if st != from {
		return fmt.Errorf("issue %s is in %s, not %s", id, st, from)
	}

	dst := filepath.Join(s.dir(to), filepath.Base(src))
	if err := os.MkdirAll(s.dir(to), permDir); err != nil {
		return fmt.Errorf("mkdir %s: %w", to, err)
	}

	if err := os.Rename(src, dst); err != nil {
		return fmt.Errorf("move %s: %w", id, err)
	}

	return nil
}

// DepAdd records dep as a dependency of id, moving id to blocked/ when it
// was open or in progress.
func (s *Store) DepAdd(id, dep string) error {
	if id == dep {
		return fmt.Errorf("issue %s cannot depend on itself", id)
	}

	// ponytail: no cycle check — a 2-issue cycle is invisible from a single
	// DepAdd; add detection if cycles actually appear.
	if _, _, err := s.locate(dep); err != nil {
		return fmt.Errorf("dependency %s: %w", dep, err)
	}

	iss, st, body, err := s.Read(id)
	if err != nil {
		return err
	}

	for _, d := range iss.BlockedBy {
		if d == dep {
			return fmt.Errorf("issue %s is already blocked by %s", id, dep)
		}
	}

	iss.BlockedBy = append(iss.BlockedBy, dep)
	if err := s.Save(iss, body); err != nil {
		return err
	}

	if st == Open || st == InProgress {
		return s.Move(id, st, Blocked)
	}

	return nil
}

// DepRemove removes dep from id's dependencies and moves id back to open/
// when it was blocked and every remaining dependency is done.
func (s *Store) DepRemove(id, dep string) error {
	iss, st, body, err := s.Read(id)
	if err != nil {
		return err
	}

	kept, found := removeDep(iss.BlockedBy, dep)
	if !found {
		return fmt.Errorf("issue %s is not blocked by %s", id, dep)
	}

	iss.BlockedBy = kept
	if err := s.Save(iss, body); err != nil {
		return err
	}

	if st != Blocked {
		return nil
	}

	if !s.allDepsDone(iss) {
		return nil
	}

	return s.Move(iss.ID, Blocked, Open)
}

// Close moves each issue to done/, appending the reason, then unblocks any
// blocked issue whose dependencies are all done.
func (s *Store) Close(ids []string, reason string) error {
	ts := time.Now().UTC().Format(time.RFC3339)

	for _, id := range ids {
		iss, st, body, err := s.Read(id)
		if err != nil {
			return err
		}

		body = appendNote(body, fmt.Sprintf("## Closed: %s (%s)", reason, ts))
		if err := s.Save(iss, body); err != nil {
			return err
		}

		if err := s.Move(id, st, Done); err != nil {
			return err
		}
	}

	blocked, err := s.List(Blocked)
	if err != nil {
		return err
	}

	for _, iss := range blocked {
		if !dependsOnAny(iss, ids) {
			continue
		}

		if !s.allDepsDone(iss) {
			continue
		}

		if err := s.Move(iss.ID, Blocked, Open); err != nil {
			return err
		}
	}

	return nil
}

// dependsOnAny reports whether any of iss's dependencies is in ids.
func dependsOnAny(iss *Issue, ids []string) bool {
	for _, dep := range iss.BlockedBy {
		for _, id := range ids {
			if dep == id {
				return true
			}
		}
	}

	return false
}

// allDepsDone reports whether every dependency of iss is in done/.
func (s *Store) allDepsDone(iss *Issue) bool {
	for _, dep := range iss.BlockedBy {
		st, _, err := s.locate(dep)
		if err != nil {
			return false
		}

		if st != Done {
			return false
		}
	}

	return true
}

// missingDep returns the first dependency id of iss whose file is missing.
func (s *Store) missingDep(iss *Issue) (string, error) {
	for _, dep := range iss.BlockedBy {
		_, _, err := s.locate(dep)
		if errors.Is(err, ErrNotFound) {
			return dep, nil
		}

		if err != nil {
			return "", err
		}
	}

	return "", nil
}

// uniqueID generates a collision-free issue id.
func (s *Store) uniqueID() (string, error) {
	for range idAttempts {
		id, err := randomID()
		if err != nil {
			return "", err
		}

		_, _, locErr := s.locate(id)
		if errors.Is(locErr, ErrNotFound) {
			return id, nil
		}

		if locErr != nil {
			return "", locErr
		}
	}

	return "", errors.New("no free issue id")
}

// locate finds the status directory and file path of an issue id.
func (s *Store) locate(id string) (Status, string, error) {
	for _, st := range Statuses() {
		entries, err := os.ReadDir(s.dir(st))
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}

		if err != nil {
			return "", "", fmt.Errorf("read %s dir: %w", st, err)
		}

		for _, e := range entries {
			name := e.Name()
			if strings.HasPrefix(name, id+"-") && strings.HasSuffix(name, ".md") {
				return st, filepath.Join(s.dir(st), name), nil
			}
		}
	}

	return "", "", ErrNotFound
}

func (s *Store) dir(st Status) string {
	return filepath.Join(s.root, string(st))
}

// writeFile creates the status directory if needed and writes name into it.
func (s *Store) writeFile(st Status, name string, data []byte) error {
	dir := s.dir(st)
	if err := os.MkdirAll(dir, permDir); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}

	if err := os.WriteFile(filepath.Join(dir, name), data, permFile); err != nil {
		return fmt.Errorf("write %s: %w", name, err)
	}

	return nil
}

// appendNote appends a markdown note line to a body.
func appendNote(body, note string) string {
	if body != "" && !strings.HasSuffix(body, "\n") {
		body += "\n"
	}

	return body + note + "\n"
}

// removeDep removes dep from deps and reports whether it was present.
func removeDep(deps []string, dep string) ([]string, bool) {
	kept := make([]string, 0, len(deps))
	found := false

	for _, d := range deps {
		if d == dep {
			found = true
			continue
		}

		kept = append(kept, d)
	}

	return kept, found
}
