package canon

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/fsport"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// Store — the canon store of an instance. All writes are atomic, all
// reads strict, all lists sorted: the same canon reads the same.
type Store struct {
	root string // .punchtape in the instance's working directory
}

// Dir returns the instance directory path for a working directory.
func Dir(workdir string) string { return filepath.Join(workdir, ".punchtape") }

// Init creates the instance skeleton in the working directory. A
// repeated call on an existing instance is an error: no silent
// overwrites.
func Init(workdir string) (*Store, error) {
	root := Dir(workdir)
	if _, err := fsport.Bind.Stat(filepath.Join(root, "state.yaml")); err == nil {
		return nil, fmt.Errorf("instance already exists: %s", root)
	}
	s := &Store{root: root}
	// This list is the manually maintained mirror of the zone model:
	// the canon collection directories plus cache and inbox. The model
	// is the source; this mirror must follow it (see the coupling note there; TestZoneModel pins
	// the derived lists).
	dirs := []string{
		filepath.Join(root, "canon", "atoms"),
		filepath.Join(root, "checkpoints"),
		filepath.Join(root, "verdicts"),
		filepath.Join(root, "cache"),
		filepath.Join(root, "inbox"),
	}
	for _, d := range dirs {
		if err := fsport.Bind.MkdirAll(d); err != nil {
			return nil, fmt.Errorf("dir %s: %w", d, err)
		}
	}
	st := State{
		SchemaVersion: SchemaVersion,
		Stage:         StageIntake,
		Counters:      map[string]int{"CKP": 0},
	}
	// PutState writes the virgin state into the freshly made skeleton:
	// the write can only fail when the filesystem is already broken
	// (the guard and the mkdirs have passed) — the error rides out
	// unwrapped, no branch to hide it.
	return s, s.PutState(st)
}

// Open opens an existing instance. A missing instance is not an
// error but a fact: exists=false.
func Open(workdir string) (*Store, bool, error) {
	root := Dir(workdir)
	if _, err := fsport.Bind.Stat(filepath.Join(root, "state.yaml")); err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("instance dir: %w", err)
	}
	return &Store{root: root}, true, nil
}

// Root — the instance directory.
func (s *Store) Root() string { return s.root }

// --- atoms ---

// Atoms — all behavior atoms, sorted by identifier.
func (s *Store) Atoms() ([]Atom, error) {
	var out []Atom
	err := eachFile(s.path("canon", "atoms"), "", func(data []byte) error {
		var a Atom
		if err := yamlio.DecodeDoc(data, &a); err != nil {
			return err
		}
		out = append(out, a)
		return nil
	})
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, err
}

// --- state ---

func (s *Store) State() (State, error) {
	data, err := fsport.Bind.ReadFile(s.path("state.yaml"))
	if err != nil {
		return State{}, fmt.Errorf("state.yaml: %w", err)
	}
	var st State
	if err := yamlio.DecodeDoc(data, &st); err != nil {
		return State{}, err
	}
	// Schema version: a foreign version is an honest refusal — the
	// schema break leaves no legacy instances, silently working with
	// a foreign version is not allowed.
	if st.SchemaVersion != SchemaVersion {
		return State{}, fmt.Errorf("%s", canondata.T("store.schema-mismatch", canondata.M{
			"have": fmt.Sprintf("%d", st.SchemaVersion), "want": fmt.Sprintf("%d", SchemaVersion),
		}))
	}
	return st, nil
}

// PutState writes the state atomically — the only direct path for
// writing state: initializing a virgin instance; all dynamics beyond
// that are written by writer transactions.
func (s *Store) PutState(st State) error {
	// State is plain scalar/slice data: marshal cannot fail
	data, _ := yamlio.Marshal(st)
	return yamlio.WriteAtomic(s.path("state.yaml"), data)
}

// --- service ---

func (s *Store) path(parts ...string) string {
	return filepath.Join(append([]string{s.root}, parts...)...)
}

// eachFile reads a directory's *.yaml in name order and calls fn for
// each file's content. A prefix narrows the name set ("" — every
// *.yaml of the directory).
func eachFile(dir, prefix string, fn func([]byte) error) error {
	entries, err := fsport.Bind.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("dir %s: %w", dir, err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if (prefix == "" || strings.HasPrefix(name, prefix)) && strings.HasSuffix(name, ".yaml") {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	for _, name := range names {
		data, err := fsport.Bind.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return fmt.Errorf("file %s: %w", name, err)
		}
		if err := fn(data); err != nil {
			return fmt.Errorf("file %s: %w", name, err)
		}
	}
	return nil
}
