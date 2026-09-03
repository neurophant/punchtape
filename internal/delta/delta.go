// Package delta — delta formats: the single form of writing into the
// canon. Decoding is strict: an unknown field or a wrong type is an
// error, the delta is rejected whole; partial applies do not exist.
// Field names and kinds are English throughout.
package delta

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// Delta kinds.
const (
	KindIntent      = "intent"
	KindLint        = "lint"
	KindClarify     = "clarify"
	KindSpec        = "spec"
	KindChecks      = "checks"
	KindAssert      = "assert"
	KindSplit       = "split"
	KindCode        = "code"
	KindFix         = "fix"
	KindProbe       = "probe"
	KindConventions = "conventions"
	KindAmend       = "amend"
	KindWaive       = "waive"
	KindDecision    = "decision"
	KindKb          = "kb"
	KindFeature     = "feature"
)

// KindAllowedOnStage — whether a delta kind is allowed at a stage. A
// fixture extends the catalog at any stage.
func KindAllowedOnStage(kind, stage string) bool {
	switch kind {
	case KindConventions:
		// conventions/environment — the instance's knowledge about the
		// product, not a flow step: learned and edited at any stage; an
		// edit invalidates the freshness of all boundaries (the digest
		// takes part in reconciliation)
		return true
	case KindIntent:
		return stage == canon.StageIntake
	case KindLint:
		// goal-lint goes right after the wish is fixed, before the
		// spec; conflict escalation — at any moment of the instance's
		// life: a contradictory wish can arrive after the verdict too;
		// the stop-with-options contract promises the human a question,
		// not a formal rejection (the conflict dimension by content
		// lives in the lint apply)
		return true
	case KindClarify:
		// clarifications — one batch at intake, after the lint; further
		// on the cycle runs without the human until the verdict
		return stage == canon.StageIntake
	case KindSpec:
		// the spec changes at acceptance compilation and at deliver: a
		// change of the existing goes through the same stages anew; at
		// implement — only repairs of the pinned scenario form
		// (update-scenario): the spec composition is frozen by started
		// code
		return stage == canon.StageSpec || stage == canon.StageACCompile ||
			stage == canon.StageImplement || stage == canon.StageDeliver
	case KindChecks:
		// the checks table is the spec's executor interface: the same
		// stages as spec; at implement it grows and is repaired line by
		// line, the composition is not rebuilt
		return stage == canon.StageSpec || stage == canon.StageACCompile ||
			stage == canon.StageImplement || stage == canon.StageDeliver
	case KindAssert:
		return stage == canon.StageACCompile
	case KindSplit:
		// a split is accepted only at implement and only before the
		// first code: started work freezes the split boundaries
		return stage == canon.StageImplement
	case KindCode:
		return stage == canon.StageImplement
	case KindFix:
		// fix repairs behavior at implement; at deliver —
		// re-verification of an instance changed outside the machine
		// (external corruption): the card reopens, the cycle returns to
		// repair
		return stage == canon.StageImplement || stage == canon.StageDeliver
	case KindAmend:
		// approving a deferred expectation change — the human's answer
		// at any moment while the change is open; without an open
		// change the submission is honestly rejected by the apply
		return true
	case KindWaive:
		// waiving a completeness lint rule with a reason — while the
		// rule holds the stage (spec and acceptance compilation) or the
		// verdict (the feature description is part of "done");
		// enforceability and stop-conditions are not waivable
		return stage == canon.StageSpec || stage == canon.StageACCompile ||
			stage == canon.StageDeliver
	case KindDecision, KindKb:
		// Decision Log, the out-of-scope zone and kb — the instance's
		// knowledge about the product (the passport's curator zone):
		// accumulated and edited at any stage, like conventions
		return true
	case KindFeature:
		// the author's feature description — a knowledge layer of the
		// human spec, not a flow step: written and ratified at any stage
		// after the wish; changes no behavior, reopens no cycle
		return stage != canon.StageIntake
	case KindProbe:
		// a behavior probe — at deliver: the spec catches up with
		// stabilized code, observation goes through the surface
		return stage == canon.StageDeliver
	default:
		return false
	}
}

// Delta — an agent's submission. Exactly one kind; its payload is filled.
type Delta struct {
	Kind          string
	SubmissionKey string

	Intent      *Intent
	Lint        *Lint
	Clarify     *Clarify
	Spec        *Spec
	Checks      *Checks
	Assert      *Assert
	Split       *Split
	Code        *Code
	Fix         *Fix
	Probe       *Probe
	Conventions *Conventions
	Amend       *Amend
	Waive       *Waive
	Decision    *DecisionDelta
	Kb          *KbDelta
	Feature     *Feature
}

// Amend — approving a deferred expectation change: the human's answer
// to a machine escalation (a submission after code start changed the
// expectations of an already-run check). approve applies the deferred
// delta in a transaction with the origin "asserted", reject dismisses
// it; while the change is open, substantive non-amend submissions are
// rejected with a decision route — there is no silent displacement of
// work.
type Amend struct {
	Kind          string `yaml:"kind"`
	SubmissionKey string `yaml:"submission-key"`
	Decision      string `yaml:"decision"`
}

// Waive — waiving a completeness lint rule with a reason: the rule
// stops blocking the transition but stays visible — the verdict's
// taste queue carries it with its reason. The reason is mandatory
// (empty — rejection); revoke removes a previously given waive.
// Channel trust is as with amend: the operator answers with their own
// channel (the caller's contract).
type Waive struct {
	Kind          string `yaml:"kind"`
	SubmissionKey string `yaml:"submission-key"`
	Rule          string `yaml:"rule"`
	Target        string `yaml:"target"`
	Reason        string `yaml:"reason,omitempty"`
	Revoke        bool   `yaml:"revoke,omitempty"`
}

// Intent — fixing the human's wish.
type Intent struct {
	Kind          string `yaml:"kind"`
	SubmissionKey string `yaml:"submission-key"`
	Text          string `yaml:"text"`
}

// FeatureEntry — one requirement's prose inside a batched feature
// delta. The fields mean exactly what they mean on the single-entry
// form of the delta itself.
type FeatureEntry struct {
	Requirement string `yaml:"requirement"`
	Text        string `yaml:"text,omitempty"`
	Ratify      bool   `yaml:"ratify,omitempty"`
}

// Feature — the author's description of a feature for the human spec:
// prose in the living words of the wish's language (what it does, how
// it is invoked, what the user sees, boundaries and errors,
// implementation details; machine facts — by reference only). The
// requirement must exist; new text removes the previous ratification —
// changed prose awaits the operator anew. Ratify — operator
// ratification of the current text. One submission carries the texts
// of ALL features (entries:), the whole human spec in one round trip;
// the single requirement+text form remains valid for one feature.
// Exactly one of the two forms per submission.
type Feature struct {
	Kind          string         `yaml:"kind"`
	SubmissionKey string         `yaml:"submission-key"`
	Requirement   string         `yaml:"requirement"`
	Text          string         `yaml:"text,omitempty"`
	Ratify        bool           `yaml:"ratify,omitempty"`
	Entries       []FeatureEntry `yaml:"entries,omitempty"`
}

// Lint — goal-lint findings: achievable, unambiguous, non-conflicting.
// At most three questions to the human, each with 2–3 options and a
// default; an empty list — the goal is clear.
type Lint struct {
	Kind          string              `yaml:"kind"`
	SubmissionKey string              `yaml:"submission-key"`
	Findings      []canon.LintFinding `yaml:"findings"`
}

// ClarifyAnswer — one human answer: the question and the chosen option.
type ClarifyAnswer struct {
	Question string `yaml:"question"`
	Option   string `yaml:"option"`
}

// Clarify — the human's answers to a goal-lint batch. Empty or missing
// answers — silence: the machine applies defaults.
type Clarify struct {
	Kind          string          `yaml:"kind"`
	SubmissionKey string          `yaml:"submission-key"`
	Answers       []ClarifyAnswer `yaml:"answers"`
}

// Spec — a change of requirements and scenarios; operations are applied
// in listed order within one transaction.
type Spec struct {
	Kind          string          `yaml:"kind"`
	SubmissionKey string          `yaml:"submission-key"`
	Operations    []SpecOperation `yaml:"operations"`
}

// SpecOperation — one spec operation: exactly one variant is filled.
type SpecOperation struct {
	AddRequirement    *AddRequirement
	UpdateRequirement *UpdateRequirement
	RemoveRequirement *RemoveRequirement
	AddScenario       *AddScenario
	UpdateScenario    *UpdateScenario
	UpdateAssertions  *UpdateAssertions
	RemoveScenario    *RemoveScenario
}

// AddRequirement — a new requirement.
type AddRequirement struct {
	ID           string                   `yaml:"id"`
	Formulation  string                   `yaml:"formulation"`
	Dependencies []string                 `yaml:"dependencies"`
	Detail       *canon.RequirementDetail `yaml:"detail,omitempty"`
}

// UpdateRequirement — changing formulation/dependencies/detail; what
// is not filled stays unchanged.
type UpdateRequirement struct {
	ID           string                   `yaml:"id"`
	Formulation  *string                  `yaml:"formulation"`
	Dependencies []string                 `yaml:"dependencies"`
	Detail       *canon.RequirementDetail `yaml:"detail,omitempty"`
}

// DecisionDelta — an entry in the Decision Log and the out-of-scope
// zone: an ambiguity, the decision and its rationale; the human's
// prose submitted as a delta through the validator. Exactly one
// operation per submission.
type DecisionDelta struct {
	Kind             string            `yaml:"kind"`
	SubmissionKey    string            `yaml:"submission-key"`
	AddDecision      *AddDecision      `yaml:"add-decision,omitempty"`
	UpdateDecision   *UpdateDecision   `yaml:"update-decision,omitempty"`
	AddOutOfScope    *AddOutOfScope    `yaml:"add-out-of-scope,omitempty"`
	RemoveOutOfScope *RemoveOutOfScope `yaml:"remove-out-of-scope,omitempty"`
}

// AddDecision — a decision with a mandatory rationale.
type AddDecision struct {
	Decision  string `yaml:"decision"`
	Rationale string `yaml:"rationale"`
}

// UpdateDecision — editing a decision entry: typos and clarifications
// without duplicates (a repeated add would create a new entry). What is
// not filled stays unchanged; at least one field must be non-empty.
type UpdateDecision struct {
	ID        string `yaml:"id"`
	Decision  string `yaml:"decision,omitempty"`
	Rationale string `yaml:"rationale,omitempty"`
}

// AddOutOfScope — extending the out-of-scope zone: what is not done
// and why.
type AddOutOfScope struct {
	What string `yaml:"what"`
	Why  string `yaml:"why"`
}

// RemoveOutOfScope — lifting a prohibition (only the human moves the
// out-of-scope boundaries with their own delta).
type RemoveOutOfScope struct {
	What string `yaml:"what"`
}

// KbDelta — the "when X do Y" knowledge base: an addition with a
// mandatory source (supersedes deactivates a previous entry), an entry
// edit or a removal. Exactly one operation per submission.
type KbDelta struct {
	Kind          string    `yaml:"kind"`
	SubmissionKey string    `yaml:"submission-key"`
	Add           *KbAdd    `yaml:"add,omitempty"`
	Update        *KbUpdate `yaml:"update,omitempty"`
	Remove        *KbRemove `yaml:"remove,omitempty"`
}

// KbAdd — a new piece of knowledge: trigger tokens (all must be found
// in the question's context), the action, the source.
type KbAdd struct {
	When       []string `yaml:"when"`
	Do         string   `yaml:"do"`
	Source     string   `yaml:"source"`
	Supersedes string   `yaml:"supersedes,omitempty"`
}

// KbUpdate — editing a knowledge entry by identifier: the identifier
// is stable (assigned at creation), the content is updated by fields;
// what is not filled stays unchanged.
type KbUpdate struct {
	ID     string   `yaml:"id"`
	When   []string `yaml:"when,omitempty"`
	Do     string   `yaml:"do,omitempty"`
	Source string   `yaml:"source,omitempty"`
}

// KbRemove — removing knowledge by identifier.
type KbRemove struct {
	ID string `yaml:"id"`
}

// RemoveRequirement — remove a requirement (it must have no
// scenarios).
type RemoveRequirement struct {
	ID string `yaml:"id"`
}

// AddScenario — a new acceptance scenario. A prose scenario is
// submitted with a reason and without steps.
type AddScenario struct {
	ID          string            `yaml:"id"`
	Requirement string            `yaml:"requirement"`
	Summary     string            `yaml:"summary"`
	Seed        []canon.Seed      `yaml:"seed,omitempty"`
	Materials   []canon.Material  `yaml:"materials,omitempty"`
	Pre         [][]string        `yaml:"pre,omitempty"`
	When        *canon.When       `yaml:"when"`
	Then        []canon.Assertion `yaml:"then"`
	Prose       *string           `yaml:"prose"`
}

// UpdateScenario — changing scenario fields; what is not filled stays
// unchanged.
type UpdateScenario struct {
	ID        string            `yaml:"id"`
	Summary   *string           `yaml:"summary"`
	Seed      []canon.Seed      `yaml:"seed,omitempty"`
	Materials []canon.Material  `yaml:"materials,omitempty"`
	Pre       [][]string        `yaml:"pre,omitempty"`
	When      *canon.When       `yaml:"when"`
	Then      []canon.Assertion `yaml:"then"`
	Prose     *string           `yaml:"prose"`
	// Requirement — rebinding to another requirement: a regrouping of
	// the boundary without changing expectations (does not touch row
	// proofs).
	Requirement *string `yaml:"requirement,omitempty"`
}

// UpdateAssertions — surgical editing of a scenario's expectations:
// operations on individual then observations without retranscribing the
// whole block (submission size ∝ the number of changes, not the block
// size). Match equals exactly one scenario observation (accounting for
// the previous operations of the same submission — the order is
// literal), set replaces it, drop removes it, add appends at the end.
type UpdateAssertions struct {
	Scenario string        `yaml:"scenario"`
	Ops      []AssertionOp `yaml:"ops"`
}

// AssertionOp — one surgical operation: exactly one of the forms
// match+set, match+drop, match+capture or add. Capture — the machine
// measures the new expectation value by running the delivered product;
// bytes are not typed by hand.
type AssertionOp struct {
	Match   *canon.Assertion `yaml:"match,omitempty"`
	Set     *canon.Assertion `yaml:"set,omitempty"`
	Drop    bool             `yaml:"drop,omitempty"`
	Add     *canon.Assertion `yaml:"add,omitempty"`
	Capture bool             `yaml:"capture,omitempty"`
}

// RemoveScenario — remove a scenario.
type RemoveScenario struct {
	ID string `yaml:"id"`
}

// Checks — the checks table: the spec's executor interface. One row —
// one check: a state seed, a command sequence (the last one is under
// test), out/err/rc/state expectations. The canon, checks, family
// coverage, derivatives and traceability are compiled by the machine
// from the table itself; restart/transience — an automatic second run
// from a clean seed, not written by the executor. A row with the same
// seed and the same command sequence replaces the previous
// expectations (line-by-line repair, without identifiers on the
// executor's side).
type Checks struct {
	Kind          string     `yaml:"kind"`
	SubmissionKey string     `yaml:"submission-key"`
	Rows          []CheckRow `yaml:"rows"`
	// Questions to the human and answers ride with the table — the
	// clarification cycle lives inside the reply to the checks table,
	// there is no separate transfer: questions — a new batch (up to
	// three, the same validation as goal-lint), answers — replies to
	// the open batch; silence applies defaults.
	Questions []canon.LintFinding `yaml:"questions,omitempty"`
	Answers   []ClarifyAnswer     `yaml:"answers,omitempty"`
	// AcceptDraft — accept the spec draft (written by a cheap executor
	// per the machine's routing) as your own spec without edits: the
	// review is a hands-on confirmation, no rows required.
	AcceptDraft bool `yaml:"accept-draft,omitempty"`
}

// CheckRow — a table row. Seeds — one-line "<path> <content>" seeds
// (one file per line); Materials — paths of project material files
// (one per line): the bytes are copied into the run directory under
// the same path byte-for-byte — binary inputs (data of any format)
// are not transferred by hand; Run — arbitrary commands (a surface
// entry point, a utility, a shell) in execution order, the
// expectations belong to the last one (each
// previous one must exit with code 0 — state setup by a series of
// calls); Out/Err/Rc — the exact stdout/stderr and the exit code of
// the last command (rc defaults to 0); State — assertions about state
// files "<path> <json>" (structural comparison, json-equals);
// Volatile — byte unpredictability declared by the brief: stream
// names (stdout/stderr) and relative state file paths (random
// identifiers) — transience does not reconcile them; adding them
// after the code goes through the human's assertion (amend).
// Key — an optional stable repair identity of the row: while the row
// has never gone green, a resubmission with the same key replaces it
// even with different seeds and run (a broken row's repair must not
// orphan the broken original); once the row is proven green, the
// key's changes go through the change-request hold like any other.
type CheckRow struct {
	Key       string   `yaml:"key,omitempty"`
	Seeds     []string `yaml:"seeds,omitempty"`
	Materials []string `yaml:"materials,omitempty"`
	Run       []RunCommand
	Out       *string  `yaml:"out,omitempty"`
	Err       *string  `yaml:"err,omitempty"`
	Rc        *int     `yaml:"rc,omitempty"`
	State     []string `yaml:"state,omitempty"`
	Volatile  []string `yaml:"volatile,omitempty"`
	// TimeoutSec — the row's own timeout ceiling (seconds); unset
	// (nil or 0) means the delivery default (limits.yaml
	// check.timeout-default-s). Long-running shapes declare it per
	// row instead of inheriting a ceiling their task form cannot
	// meet.
	TimeoutSec *int `yaml:"timeout-sec,omitempty"`
}

// RunCommand — a run element in two equal forms: a "command args"
// string (whitespace splitting; a single-quote wrapper preserves
// inner double quotes) or a ready argv list — with no splitting at
// all (an argument with commas/quotes travels as is). Line identity
// is computed by the argv fields: both forms of the same commands are
// one identity; the form is only notation.
type RunCommand struct {
	Line   string
	Argv   []string
	IsArgv bool
}

// UnmarshalYAML accepts a scalar (a command line) or a sequence of
// scalars (argv).
func (rc *RunCommand) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		rc.Line = node.Value
		return nil
	case yaml.SequenceNode:
		argv := make([]string, 0, len(node.Content))
		for _, c := range node.Content {
			if c.Kind != yaml.ScalarNode {
				return fmt.Errorf("%s", canondata.T("delta.error.run-argv-scalar"))
			}
			argv = append(argv, c.Value)
		}
		rc.Argv = argv
		rc.IsArgv = true
		return nil
	default:
		return fmt.Errorf("%s", canondata.T("delta.error.run-form"))
	}
}

// MarshalYAML — the canonical form: a string as a scalar, argv as a
// list; a two-way rewrite of the same form.
func (rc RunCommand) MarshalYAML() (any, error) {
	if rc.IsArgv {
		return rc.Argv, nil
	}
	return rc.Line, nil
}

// Fields — the command's argv fields: a ready list or a string split.
func (rc RunCommand) Fields() ([]string, error) {
	if rc.IsArgv {
		if len(rc.Argv) == 0 {
			return nil, fmt.Errorf("%s", canondata.T("delta.error.run-argv-empty"))
		}
		return rc.Argv, nil
	}
	return canon.SplitCommand(rc.Line)
}

// YAML — the element's form for human renders: the string as is, argv
// as a flow list (a resubmission reads in the same form).
func (rc RunCommand) YAML() string {
	if !rc.IsArgv {
		return rc.Line
	}
	parts := make([]string, len(rc.Argv))
	for i, a := range rc.Argv {
		if a == "" || strings.ContainsAny(a, " \t,[]{}:\"'#") {
			parts[i] = doubleQuote(a)
		} else {
			parts[i] = a
		}
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// Assert — pin assertions onto an unfinished scenario (empty then);
// steps and fixtures are untouchable.
type Assert struct {
	Kind          string            `yaml:"kind"`
	SubmissionKey string            `yaml:"submission-key"`
	Scenario      string            `yaml:"scenario"`
	Assertions    []canon.Assertion `yaml:"assertions"`
}

// SplitCard — a split card: the facet, the scenarios it closes and the
// files the card owns. Ownership is exclusive: two parallel executors
// do not write one file.
type SplitCard struct {
	ID        string   `yaml:"id"`
	Facet     string   `yaml:"facet"`
	Scenarios []string `yaml:"scenarios"`
	Files     []string `yaml:"files"`
}

// SplitContract — a split contract: the shared observable asset of two
// facets in surface terms. The machine knows the status itself: an
// accepted split freezes the contracts; the agent does not choose the
// status.
type SplitContract struct {
	ID      string   `yaml:"id"`
	Version string   `yaml:"version"`
	Sides   []string `yaml:"sides"`
	Surface []string `yaml:"surface"`
}

// Split — splitting the product into facets of parallel work: cards
// replace the whole-product card, contracts fix the boundaries.
type Split struct {
	Kind          string          `yaml:"kind"`
	SubmissionKey string          `yaml:"submission-key"`
	Cards         []SplitCard     `yaml:"cards"`
	Contracts     []SplitContract `yaml:"contracts,omitempty"`
}

// FileChange — one submission file. Savings on repeated submissions:
// instead of the content, the digest field — the machine substitutes
// the latest content with this digest from the canon; exactly one of
// the two fields is filled.
type FileChange struct {
	Path    string `yaml:"path"`
	Content string `yaml:"content,omitempty"`
	Digest  string `yaml:"digest,omitempty"`
}

// Code — implementation files for a card.
type Code struct {
	Kind          string          `yaml:"kind"`
	SubmissionKey string          `yaml:"submission-key"`
	Card          string          `yaml:"card"`
	Files         []FileChange    `yaml:"files"`
	Delete        []string        `yaml:"delete"`
	Answers       []ClarifyAnswer `yaml:"answers,omitempty"`
}

// Fix — a repair by the gate's red list; the same as code plus a gate
// reference; the card's attempt counter grows only when the card's
// behavior is red (aligning to a repaired spec is free).
type Fix struct {
	Kind          string          `yaml:"kind"`
	SubmissionKey string          `yaml:"submission-key"`
	Card          string          `yaml:"card"`
	ReplyTo       string          `yaml:"reply-to"`
	Files         []FileChange    `yaml:"files"`
	Delete        []string        `yaml:"delete"`
	Answers       []ClarifyAnswer `yaml:"answers,omitempty"`
}

// RecipeDecl — an accumulated recipe in a submission: commands without
// a shell, {out}/{name} placeholders; out — the declared surface
// artifact (a project path, a file or a directory), empty — the
// surface is submitted directly as a file named after the command.
type RecipeDecl struct {
	Out      string     `yaml:"out,omitempty"`
	Commands [][]string `yaml:"commands"`
}

// Conventions — a declaration of the product's conventions/environment:
// a full replacement of the section (the latest submission is the
// truth). Build must build the surface (carry {out}); types/lint —
// static gate commands, optional. The operator decides the stack and
// tools — the machine executes.
type Conventions struct {
	Kind          string      `yaml:"kind"`
	SubmissionKey string      `yaml:"submission-key"`
	Build         *RecipeDecl `yaml:"build,omitempty"`
	Types         *RecipeDecl `yaml:"types,omitempty"`
	Lint          *RecipeDecl `yaml:"lint,omitempty"`
}

// Probe — surface commands to formalize: the machine will run each one
// twice through the surface in a clean directory with a seed and
// propose a spec from the observed. Submitted at deliver — the spec
// catches up with stabilized code.
type Probe struct {
	Kind          string       `yaml:"kind"`
	SubmissionKey string       `yaml:"submission-key"`
	Seed          []canon.Seed `yaml:"seed,omitempty"`
	Commands      [][]string   `yaml:"commands"`
}

// WithoutKeys — a copy of the delta without submission keys: a flat
// document is decoded both into the top level and into the payload —
// the key lives in both layers as a duplicate; numbering is not the
// delta's content, content idempotency compares the rest.
func (d Delta) WithoutKeys() Delta {
	d.SubmissionKey = ""
	if d.Intent != nil {
		p := *d.Intent
		p.SubmissionKey = ""
		d.Intent = &p
	}
	if d.Lint != nil {
		p := *d.Lint
		p.SubmissionKey = ""
		d.Lint = &p
	}
	if d.Clarify != nil {
		p := *d.Clarify
		p.SubmissionKey = ""
		d.Clarify = &p
	}
	if d.Spec != nil {
		p := *d.Spec
		p.SubmissionKey = ""
		d.Spec = &p
	}
	if d.Checks != nil {
		p := *d.Checks
		p.SubmissionKey = ""
		d.Checks = &p
	}
	if d.Assert != nil {
		p := *d.Assert
		p.SubmissionKey = ""
		d.Assert = &p
	}
	if d.Split != nil {
		p := *d.Split
		p.SubmissionKey = ""
		d.Split = &p
	}
	if d.Code != nil {
		p := *d.Code
		p.SubmissionKey = ""
		d.Code = &p
	}
	if d.Fix != nil {
		p := *d.Fix
		p.SubmissionKey = ""
		d.Fix = &p
	}
	if d.Probe != nil {
		p := *d.Probe
		p.SubmissionKey = ""
		d.Probe = &p
	}
	if d.Conventions != nil {
		p := *d.Conventions
		p.SubmissionKey = ""
		d.Conventions = &p
	}
	if d.Amend != nil {
		p := *d.Amend
		p.SubmissionKey = ""
		d.Amend = &p
	}
	if d.Waive != nil {
		p := *d.Waive
		p.SubmissionKey = ""
		d.Waive = &p
	}
	if d.Decision != nil {
		p := *d.Decision
		p.SubmissionKey = ""
		d.Decision = &p
	}
	if d.Kb != nil {
		p := *d.Kb
		p.SubmissionKey = ""
		d.Kb = &p
	}
	return d
}

// UnmarshalYAML first reads the kind, then strictly decodes the
// document into that kind's struct: foreign fields do not pass.
func (d *Delta) UnmarshalYAML(node *yaml.Node) error {
	var head struct {
		Kind          string `yaml:"kind"`
		SubmissionKey string `yaml:"submission-key"`
	}
	if err := node.Decode(&head); err != nil {
		return err
	}
	d.Kind = head.Kind
	d.SubmissionKey = head.SubmissionKey

	raw, err := yaml.Marshal(node)
	if err != nil {
		return err
	}
	switch d.Kind {
	case KindIntent:
		d.Intent = &Intent{}
		return yamlio.DecodeStrict(raw, d.Intent)
	case KindLint:
		d.Lint = &Lint{}
		return yamlio.DecodeStrict(raw, d.Lint)
	case KindClarify:
		d.Clarify = &Clarify{}
		return yamlio.DecodeStrict(raw, d.Clarify)
	case KindSpec:
		d.Spec = &Spec{}
		return yamlio.DecodeStrict(raw, d.Spec)
	case KindChecks:
		d.Checks = &Checks{}
		return yamlio.DecodeStrict(raw, d.Checks)
	case KindAssert:
		d.Assert = &Assert{}
		return yamlio.DecodeStrict(raw, d.Assert)
	case KindSplit:
		d.Split = &Split{}
		return yamlio.DecodeStrict(raw, d.Split)
	case KindCode:
		d.Code = &Code{}
		return yamlio.DecodeStrict(raw, d.Code)
	case KindFix:
		d.Fix = &Fix{}
		return yamlio.DecodeStrict(raw, d.Fix)
	case KindProbe:
		d.Probe = &Probe{}
		return yamlio.DecodeStrict(raw, d.Probe)
	case KindConventions:
		d.Conventions = &Conventions{}
		return yamlio.DecodeStrict(raw, d.Conventions)
	case KindAmend:
		d.Amend = &Amend{}
		return yamlio.DecodeStrict(raw, d.Amend)
	case KindWaive:
		d.Waive = &Waive{}
		return yamlio.DecodeStrict(raw, d.Waive)
	case KindDecision:
		d.Decision = &DecisionDelta{}
		return yamlio.DecodeStrict(raw, d.Decision)
	case KindKb:
		d.Kb = &KbDelta{}
		return yamlio.DecodeStrict(raw, d.Kb)
	case KindFeature:
		d.Feature = &Feature{}
		return yamlio.DecodeStrict(raw, d.Feature)
	default:
		return fmt.Errorf("%s", canondata.T("format.error.unknown-kind",
			canondata.M{"kind": fmt.Sprintf("%q", d.Kind)}))
	}
}

// specOpKeys — spec operation names in YAML.
var specOpKeys = map[string]func(*SpecOperation) any{
	"add-requirement":    func(o *SpecOperation) any { return &o.AddRequirement },
	"update-requirement": func(o *SpecOperation) any { return &o.UpdateRequirement },
	"remove-requirement": func(o *SpecOperation) any { return &o.RemoveRequirement },
	"add-scenario":       func(o *SpecOperation) any { return &o.AddScenario },
	"update-scenario":    func(o *SpecOperation) any { return &o.UpdateScenario },
	"update-assertions":  func(o *SpecOperation) any { return &o.UpdateAssertions },
	"remove-scenario":    func(o *SpecOperation) any { return &o.RemoveScenario },
}

// UnmarshalYAML parses a spec operation: a list element is a mapping
// with exactly one key — the operation name.
func (o *SpecOperation) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("%s", canondata.T("format.error.op-mapping"))
	}
	content := node.Content
	if len(content) != 2 {
		return fmt.Errorf("%s", canondata.T("format.error.op-one-key",
			canondata.M{"count": fmt.Sprintf("%d", len(content)/2)}))
	}
	key := content[0].Value
	mk, ok := specOpKeys[key]
	if !ok {
		return fmt.Errorf("%s", canondata.T("format.error.op-unknown",
			canondata.M{"op": fmt.Sprintf("%q", key)}))
	}
	target := mk(o)
	raw, err := yaml.Marshal(content[1])
	if err != nil {
		return err
	}
	return yamlio.DecodeStrict(raw, target)
}

// MarshalYAML prints the spec operation with the same single key:
// machine spec proposals (backfill) are assembled with the same
// structs that are parsed.
func (o SpecOperation) MarshalYAML() (any, error) {
	switch {
	case o.AddRequirement != nil:
		return map[string]any{"add-requirement": o.AddRequirement}, nil
	case o.UpdateRequirement != nil:
		return map[string]any{"update-requirement": o.UpdateRequirement}, nil
	case o.RemoveRequirement != nil:
		return map[string]any{"remove-requirement": o.RemoveRequirement}, nil
	case o.AddScenario != nil:
		return map[string]any{"add-scenario": o.AddScenario}, nil
	case o.UpdateScenario != nil:
		return map[string]any{"update-scenario": o.UpdateScenario}, nil
	case o.UpdateAssertions != nil:
		return map[string]any{"update-assertions": o.UpdateAssertions}, nil
	case o.RemoveScenario != nil:
		return map[string]any{"remove-scenario": o.RemoveScenario}, nil
	}
	return nil, fmt.Errorf("%s", canondata.T("format.error.op-empty"))
}
