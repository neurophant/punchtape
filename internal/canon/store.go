package canon

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/neurophant/punchtape/internal/canondata"
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
	if _, err := os.Stat(filepath.Join(root, "state.yaml")); err == nil {
		return nil, fmt.Errorf("instance already exists: %s", root)
	}
	s := &Store{root: root}
	dirs := []string{
		filepath.Join(root, "canon", "requirements"),
		filepath.Join(root, "canon", "scenarios"),
		filepath.Join(root, "canon", "contracts"),
		filepath.Join(root, "canon", "cards"),
		filepath.Join(root, "canon", "checks"),
		filepath.Join(root, "passport", "boundaries"),
		filepath.Join(root, "cache"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return nil, fmt.Errorf("dir %s: %w", d, err)
		}
	}
	st := State{
		SchemaVersion: SchemaVersion,
		Stage:         StageIntake,
		Counters:      map[string]int{"REQ": 1, "SCN": 1, "IFC": 1, "CRD": 1, "TST": 1},
	}
	if err := s.PutState(st); err != nil {
		return nil, err
	}
	return s, nil
}

// Open opens an existing instance. A missing instance is not an
// error but a fact: exists=false.
func Open(workdir string) (*Store, bool, error) {
	root := Dir(workdir)
	if _, err := os.Stat(filepath.Join(root, "state.yaml")); err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("instance dir: %w", err)
	}
	return &Store{root: root}, true, nil
}

// Root — the instance directory.
func (s *Store) Root() string { return s.root }

// --- requirements ---

// Requirements — all requirements, sorted by identifier.
func (s *Store) Requirements() ([]Requirement, error) {
	var out []Requirement
	err := eachFile(s.path("canon", "requirements"), "req", func(data []byte) error {
		var r Requirement
		if err := yamlio.DecodeStrict(data, &r); err != nil {
			return err
		}
		out = append(out, r)
		return nil
	})
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, err
}

// --- scenarios ---

// Scenarios — all scenarios, sorted by identifier.
func (s *Store) Scenarios() ([]Scenario, error) {
	var out []Scenario
	err := eachFile(s.path("canon", "scenarios"), "scn", func(data []byte) error {
		var sc Scenario
		if err := yamlio.DecodeStrict(data, &sc); err != nil {
			return err
		}
		out = append(out, sc)
		return nil
	})
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, err
}

// --- interface contracts ---

// Contracts — all contracts, sorted by identifier.
func (s *Store) Contracts() ([]Contract, error) {
	var out []Contract
	err := eachFile(s.path("canon", "contracts"), "ifc", func(data []byte) error {
		var c Contract
		if err := yamlio.DecodeStrict(data, &c); err != nil {
			return err
		}
		out = append(out, c)
		return nil
	})
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, err
}

// --- cards ---

// Cards — all cards, sorted by identifier.
func (s *Store) Cards() ([]Card, error) {
	var out []Card
	err := eachFile(s.path("canon", "cards"), "crd", func(data []byte) error {
		var c Card
		if err := yamlio.DecodeStrict(data, &c); err != nil {
			return err
		}
		out = append(out, c)
		return nil
	})
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, err
}

// --- generated checks ---

// Checks — all checks, sorted by identifier.
func (s *Store) Checks() ([]Check, error) {
	var out []Check
	err := eachFile(s.path("canon", "checks"), "tst", func(data []byte) error {
		var c Check
		if err := yamlio.DecodeStrict(data, &c); err != nil {
			return err
		}
		out = append(out, c)
		return nil
	})
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, err
}

func (s *Store) State() (State, error) {
	data, err := os.ReadFile(s.path("state.yaml"))
	if err != nil {
		return State{}, fmt.Errorf("state.yaml: %w", err)
	}
	var st State
	if err := yamlio.DecodeStrict(data, &st); err != nil {
		return State{}, err
	}
	// Schema version: newer than the binary — an honest refusal;
	// older — migrations are not needed yet (there is one version);
	// silently working with a foreign version is not allowed.
	if st.SchemaVersion > SchemaVersion {
		return State{}, fmt.Errorf("%s", canondata.T("store.schema-newer", canondata.M{
			"have": fmt.Sprintf("%d", st.SchemaVersion), "want": fmt.Sprintf("%d", SchemaVersion),
		}))
	}
	return st, nil
}

// PutState writes the state atomically — the only direct path for
// writing state: initializing a virgin instance; all dynamics beyond
// that are written by writer transactions. The manifest header is
// the first line: the machine writes the file; it is not edited by
// hand.
func (s *Store) PutState(st State) error {
	data, err := yamlio.Marshal(st)
	if err != nil {
		return err
	}
	return yamlio.WriteAtomic(s.path("state.yaml"), append([]byte(MachineHeader("next, submit")+"\n"), data...))
}

// --- service ---

func (s *Store) path(parts ...string) string {
	return filepath.Join(append([]string{s.root}, parts...)...)
}

// eachFile reads a directory's *.yaml in name order and calls fn for
// each file's content.
func eachFile(dir, prefix string, fn func([]byte) error) error {
	entries, err := os.ReadDir(dir)
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
		if strings.HasPrefix(name, prefix) && strings.HasSuffix(name, ".yaml") {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	for _, name := range names {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return fmt.Errorf("file %s: %w", name, err)
		}
		if err := fn(data); err != nil {
			return fmt.Errorf("file %s: %w", name, err)
		}
	}
	return nil
}
