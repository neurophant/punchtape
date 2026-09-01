// Product passport — the boundary registry and environment
// conventions: product knowledge lives in the instance; the machine
// accumulates and maintains it. A boundary = a requirement (REQ): a
// stable product unit that survives splits and change requests. Only
// the machine writes the records (gates and spec synchronization);
// conventions are an operator declaration accepted as a delta through
// the validator.
package canon

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// Boundary trust statuses. The full ladder (claimed→verified→valid→
// reliable, as data with weights) is a separate table feature; here
// are the machine transitions: creation — claimed, a green run —
// verified, drift — claimed again. The machine does not raise above
// "verified": there was no proof.
const (
	BoundaryClaimed  = "claimed"
	BoundaryVerified = "verified"
	BoundaryValid    = "valid"
	BoundaryReliable = "reliable"
	NoBoundaryStatus = ""
)

// Boundary origin: wish — from the wish through the spec; backfill —
// from exploring existing code (the backfill pipeline).
const (
	OriginWish     = "wish"
	OriginBackfill = "backfill"
)

// BoundaryVerdict — with what and when a boundary was verified:
// freshness is computed by comparison with the current state (the
// scope-hash of code files and the conventions digest); git is not
// read. JournalTx — the run transaction that produced the proof: an
// anchor checkable against the journal.
type BoundaryVerdict struct {
	ScopeHash         string `yaml:"scope-hash"`
	ConventionsDigest string `yaml:"conventions-digest"`
	JournalTx         string `yaml:"journal-tx"`
}

// Boundary — a passport boundary registry record.
type Boundary struct {
	ID     string `yaml:"id"`
	Status string `yaml:"status"`
	Origin string `yaml:"origin"`
	// Domain — the boundary's observable command families: the first
	// tokens of scenario commands (command[0]); the machine derives
	// them from the spec.
	Domain []string `yaml:"domain,omitempty"`
	// ScopeFiles — the boundary's code files; the machine derives them
	// from the journal by implementation facets. Attribution — the
	// machine's reason for coarse attribution (the whole product
	// before a split shares files across all boundaries).
	ScopeFiles  []string `yaml:"scope-files,omitempty"`
	Attribution string   `yaml:"attribution,omitempty"`
	// Digest — the boundary's spec digest (requirement + scenarios):
	// changes with a spec edit; a change is a re-claim (the status
	// drops).
	Digest string `yaml:"digest"`
	// Verdict — the latest proof; nil — the boundary is not verified
	// yet.
	Verdict *BoundaryVerdict `yaml:"verdict,omitempty"`
}

// Recipe — an accumulated recipe: a command sequence without a
// shell, {out} (the surface artifact path) and {name} (the surface
// name, command[0]) placeholders. The artifact is a file or
// directory: the operator decides the surface's form; the machine
// runs the commands and carries the declared into the run directory
// as is. All commands must exit with code 0.
type Recipe struct {
	// Out — the declared surface artifact: a project path, a file or
	// a directory; empty — no artifact declared, the surface is
	// submitted directly as a file named after the command.
	Out      string     `yaml:"out,omitempty"`
	Commands [][]string `yaml:"commands"`
}

// Conventions — the product's conventions/environment in one section
// (an edit invalidates the freshness of all boundaries — the digest
// takes part in reconciliation). An empty file/absence — an honest
// empty start: no recipes; the machine applies nothing and asks when
// a recipe is needed.
type Conventions struct {
	SchemaVersion int     `yaml:"schema-version"`
	Build         *Recipe `yaml:"build,omitempty"`
	Types         *Recipe `yaml:"types,omitempty"`
	Lint          *Recipe `yaml:"lint,omitempty"`
}

// Passport paths in the instance (from the project root).
const (
	ConventionsFile = ".punchtape/passport/conventions.yaml"
	BoundariesDir   = ".punchtape/passport/boundaries"
	DecisionsFile   = ".punchtape/passport/decisions.yaml"
	OutOfScopeFile  = ".punchtape/passport/outofscope.yaml"
	KbFile          = ".punchtape/passport/kb.yaml"
)

// Decision — a Decision Log entry: an ambiguity, the decision and
// its rationale. Written by a kind: decision delta through the
// validator; the machine assigns identifiers deterministically
// (DEC-<n>).
type Decision struct {
	ID        string    `yaml:"id"`
	Decision  string    `yaml:"decision"`
	Rationale string    `yaml:"rationale"`
	At        time.Time `yaml:"recorded-at"`
}

// OutOfScopeItem — an explicit out-of-scope zone item of the product:
// what is not done and why — against "the agent invented extra".
type OutOfScopeItem struct {
	What string `yaml:"what"`
	Why  string `yaml:"why"`
}

// KbEntry — a "when X do Y" piece of knowledge: trigger tokens (all
// must be found in the context — a deterministic match), the action,
// a mandatory source; supersedes deactivates a previous entry.
type KbEntry struct {
	ID         string   `yaml:"id"`
	When       []string `yaml:"when"`
	Do         string   `yaml:"do"`
	Source     string   `yaml:"source"`
	Supersedes string   `yaml:"supersedes,omitempty"`
}

// Decisions — the Decision Log; a missing file is an empty log.
func (s *Store) Decisions() ([]Decision, error) {
	return readPassportList[Decision](s, "decisions.yaml")
}

// OutOfScope — the out-of-scope zone; a missing file is an empty
// zone.
func (s *Store) OutOfScope() ([]OutOfScopeItem, error) {
	return readPassportList[OutOfScopeItem](s, "outofscope.yaml")
}

// KbEntries — active knowledge; a missing file is an empty
// reference.
func (s *Store) KbEntries() ([]KbEntry, error) {
	return readPassportList[KbEntry](s, "kb.yaml")
}

// readPassportList — the shared list reader of the curator zone: the
// file is strictly decoded, absence — empty, corruption — a system
// failure (the passport is written only by the machine).
func readPassportList[T any](s *Store, name string) ([]T, error) {
	data, err := os.ReadFile(filepath.Join(s.root, "passport", name))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []T
	if err := yamlio.DecodeStrict(data, &out); err != nil {
		return nil, fmt.Errorf("passport %s: %w", name, err)
	}
	return out, nil
}

// ConventionsDigest — the digest of the current conventions: a
// participant in the boundary freshness reconciliation. Empty
// conventions — an empty stub digest.
func (s *Store) ConventionsDigest() string {
	c, err := s.Conventions()
	if err != nil || c == nil {
		return "-"
	}
	data, err := yamlio.Marshal(c)
	if err != nil {
		return "-"
	}
	return yamlio.Digest(data)
}

// Conventions reads the product's conventions; a missing file — nil
// without an error (an empty start). The file exists but does not
// read strictly — a system failure: conventions are written only by
// the machine.
func (s *Store) Conventions() (*Conventions, error) {
	data, err := os.ReadFile(filepath.Join(s.root, "passport", "conventions.yaml"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var c Conventions
	if err := yamlio.DecodeStrict(data, &c); err != nil {
		return nil, fmt.Errorf("conventions: %w", err)
	}
	return &c, nil
}

// Boundaries reads the boundary registry; a missing directory is an
// empty registry. Every record's status must belong to the delivery's
// trust ladder (trust.yaml): only the machine sets statuses; there
// are no off-ladder ones.
func (s *Store) Boundaries() ([]Boundary, error) {
	dir := filepath.Join(s.root, "passport", "boundaries")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []Boundary
	for _, en := range entries {
		if en.IsDir() || !strings.HasSuffix(en.Name(), ".yaml") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, en.Name()))
		if err != nil {
			return nil, err
		}
		var b Boundary
		if err := yamlio.DecodeStrict(data, &b); err != nil {
			return nil, fmt.Errorf("boundary %s: %w", en.Name(), err)
		}
		if _, ok := canondata.TrustWeight(b.Status); !ok {
			return nil, fmt.Errorf("boundary %s: %s", en.Name(),
				canondata.T("canonvalid.error.trust-status", canondata.M{"status": b.Status}))
		}
		out = append(out, b)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}
