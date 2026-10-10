// Package delta — the submission document: the single form of
// writing into the machine. Exactly two kinds — spec
// (validate and store, no execution) and run (the same grammar;
// executes the rows and may finalize with ratify). Decoding is
// strict: an unknown field or a wrong type is an error, the
// submission is rejected whole; partial applies do not exist. Field
// names and kinds are English throughout.
package delta

import (
	"fmt"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// Delta kinds.
const (
	KindSpec = "spec"
	KindRun  = "run"
)

// KnownKind — whether a kind is one of the submission kinds.
func KnownKind(kind string) bool {
	return kind == KindSpec || kind == KindRun
}

// KindAllowedOnStage — whether a kind is allowed at a stage. A spec
// is the state of the living case: accepted at intake and implement,
// and at deliver it is the NEXT VERSION — the case reopens (a
// ratification never closes the door). A run needs the work open
// (implement): at deliver only the next spec version enters.
func KindAllowedOnStage(kind, stage string) bool {
	switch kind {
	case KindSpec:
		return stage == canon.StageIntake || stage == canon.StageImplement ||
			stage == canon.StageDeliver
	case KindRun:
		return stage == canon.StageImplement
	default:
		return false
	}
}

// Doc — the submission document. One shape for both kinds: the kind
// decides what the machine does with it, not which fields it has.
type Doc struct {
	Kind string `yaml:"kind"`
	// Build — the build command (argv); runs in the work dir before
	// the rows of every run submission; rc must be 0. nil — never
	// declared, [] — declared empty (no build). A declaration
	// replaces the previous one.
	Build []string `yaml:"build,omitempty"`
	// BuildIgnore — ignore globs: paths of the work tree outside the
	// accounted surface (build outputs, caches). Replaces the
	// previous list when present.
	BuildIgnore []string `yaml:"build-ignore,omitempty"`
	// Behaviors — the behavior units. In kind spec the document IS
	// the full state of the case: every behavior of the corpus lives
	// in it, an id absent from it is RETIRED (a visible diff frame;
	// the next verdict explains every retired row), a shrunk row
	// list retires the tail. In kind run the behaviors upsert by id
	// (repairs): a repeated id replaces the atom's name/prose/rows,
	// a new id adds an atom, nothing is ever retired by a run.
	// Required in spec (1+); optional in run (absent — a re-run of
	// the stale rows as is).
	// After the first accepted spec a full behaviors table is legal
	// only as a deliberate full-replace (the owner's economy law):
	// every later submission rides Patch instead.
	Behaviors []Behavior `yaml:"behaviors,omitempty"`
	// Patch — the sparse change set by row address (the
	// owner's law: the full table is submitted exactly once — with
	// the first spec; every later submission is a patch, and the
	// number of submissions stops being an economic variable). A
	// patch document carries no behaviors; the kinds stay exactly
	// two (spec validates the change, run executes it).
	Patch *Patch `yaml:"patch,omitempty"`
	// Answers — answers to a machine question batch. No batch can be
	// open in this machine generation: the field is parsed and
	// honestly refused (an answer without a question is not
	// silence, it is noise).
	Answers []Answer `yaml:"answers,omitempty"`
	// Ratify — finalize the instance with this run: the machine checks
	// every row green (the prose is a form guarantee) and the tree
	// intact, then renders the verdict. Only legal in kind: run.
	Ratify bool `yaml:"ratify,omitempty"`
	// FullReplace — the deliberate full-rewrite hatch: the document
	// IS a full behaviors table and the hand says so on purpose (a
	// regrouping of capabilities). The machine accepts it, confirms
	// with the row count "old N -> new M" and writes the event into
	// the cell ledger. Never composes with a patch.
	FullReplace bool `yaml:"full-replace,omitempty"`
}

// Patch — the sparse change set of a later submission: every section
// addresses the corpus by ROW ADDRESS — the label the machine itself
// prints everywhere (BEHAVIOR-ID/N, one-based, the corpus as it
// stands at the patch's arrival). Replacements edit their row in
// place; added rows go to the END of their behavior's table; removals
// retire their rows (a visible frame; the next verdict explains
// them). The run section names rows to execute in this transaction
// whatever their freshness, with optional procedural-mode updates.
// The prose section replaces a behavior's prose — the verdict-gate
// prose rides here, once.
type Patch struct {
	// Replace — full new row bytes at the address, edited in place.
	Replace map[string]canon.Row `yaml:"replace,omitempty"`
	// Add — rows appended to the end of their behavior's table; a new
	// behavior id opens the behavior (name and prose mandatory on
	// the first add item of that id).
	Add []PatchAdd `yaml:"add,omitempty"`
	// Remove — row addresses retired by this patch.
	Remove []string `yaml:"remove,omitempty"`
	// Run — rows to execute in this transaction (fresh or not), with
	// optional procedural-mode replacements (launch, tree).
	Run map[string]PatchRun `yaml:"run,omitempty"`
	// Prose — the behavior prose replacement (the verdict-gate texts
	// ride the spec patch's body, once per behavior).
	Prose map[string]string `yaml:"prose,omitempty"`
}

// PatchAdd — one appended row and the behavior it joins.
type PatchAdd struct {
	Atom string `yaml:"atom"`
	// Name/Prose — required on the first add of a NEW behavior id;
	// ignored (may be empty) when the behavior already stands.
	Name  string    `yaml:"name,omitempty"`
	Prose string    `yaml:"prose,omitempty"`
	Row   canon.Row `yaml:"row"`
}

// PatchRun — the execution request of one addressed row: the optional
// body replaces the row's procedural modes, then the row runs in this
// transaction whatever its freshness (an explicit measurement).
type PatchRun struct {
	Launch string `yaml:"launch,omitempty"`
	Tree   string `yaml:"tree,omitempty"`
}

// Behavior — one unit of behavior: the executor's short id, a living
// name, the mandatory prose and the check rows. The behavior is the
// atom of the canon and the unit of the verdict.
type Behavior struct {
	ID    string      `yaml:"id"`
	Name  string      `yaml:"name"`
	Prose string      `yaml:"prose"`
	Rows  []canon.Row `yaml:"rows"`
}

// Answer — one answer of a question batch reply: the question id and
// the chosen option.
type Answer struct {
	Q      string `yaml:"q"`
	Choice string `yaml:"choice"`
}

// Decode — the strict decode of a submission document. The kind head
// names the only two legal kinds; everything else refuses with a
// suggestion.
func Decode(raw []byte) (*Doc, error) {
	var d Doc
	if err := yamlio.DecodeDoc(raw, &d); err != nil {
		return nil, err
	}
	if !KnownKind(d.Kind) {
		hint := ""
		if s, ok := SuggestKind(d.Kind); ok {
			hint = s
		}
		return nil, fmt.Errorf("%s", canondata.T("format.error.unknown-kind", canondata.M{
			"kind": fmt.Sprintf("%q", d.Kind), "hint": hint,
		}))
	}
	return &d, nil
}
