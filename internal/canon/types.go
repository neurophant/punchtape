// Package canon — the canon entities and their store. The canon is the
// single source of truth about an instance; only the core writes the
// files, decoding is strict: an unknown field is an error. Field names
// and enumeration values are English: this is the machine layer.
package canon

import (
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/neurophant/punchtape/internal/canondata"
)

// Requirement statuses.
const (
	ReqDraft      = "draft"
	ReqAccepted   = "accepted"
	ReqRealized   = "realized"
	ReqQuarantine = "quarantine"
	ReqCancelled  = "cancelled"
)

// Requirement — one atomic requirement. The requirement's scenarios
// are not stored here: the link lives on the scenario's side, one
// source of truth.
type Requirement struct {
	ID           string   `yaml:"id"`
	Status       string   `yaml:"status"`
	Dependencies []string `yaml:"dependencies,omitempty"`
	Formulation  string   `yaml:"formulation"`
	// Detail — the boundary's detail minimum: each of the four
	// behavior classes (invariants, edge cases, failure matrix,
	// concurrency) is covered either by the requirement's scenarios or
	// by an explicit "no" with a reason. The block is optional
	// (incremental filling): the passport inventory sees boundaries
	// without the block; a filled block must be complete.
	Detail *RequirementDetail `yaml:"detail,omitempty"`
	// SpecDoc — the author's feature prose for the human spec: living
	// words in the wish's language. The author is the executor via a
	// kind: feature delta; ratification is the operator's. nil — an
	// honest "no description".
	SpecDoc *FeatureDoc `yaml:"spec-doc,omitempty"`
}

// FeatureDoc — the author's feature description: text in living words
// and the fact of operator ratification. Changed text removes the
// ratification — new prose awaits approval anew.
type FeatureDoc struct {
	Text     string `yaml:"text"`
	Ratified bool   `yaml:"ratified,omitempty"`
}

// DetailClass — one behavior class of a requirement: either references
// to the requirement's scenarios or an explicit "no" with a non-empty
// reason.
type DetailClass struct {
	Scenarios []string `yaml:"scenarios,omitempty"`
	Absent    string   `yaml:"absent,omitempty"`
}

// RequirementDetail — the boundary's minimal completeness: all four
// behavior classes are named — by scenarios or by an honest refusal
// with a reason.
type RequirementDetail struct {
	Invariants    *DetailClass `yaml:"invariants"`
	EdgeCases     *DetailClass `yaml:"edge-cases"`
	FailureMatrix *DetailClass `yaml:"failure-matrix"`
	Concurrency   *DetailClass `yaml:"concurrency"`
}

// Instance stages. The stage order is a property of the data process
// (the delivery's scenarios.yaml): the code only lists the values.
const (
	StageIntake    = "intake"
	StageSpec      = "spec"
	StageACCompile = "ac-compile"
	StageImplement = "implement"
	StageReview    = "review"
	StageDeliver   = "deliver"
)

// StageOrder — stages in the data process's order: derived from the
// delivery; the code keeps no separate list (zero mirrors).
var StageOrder = canondata.StageIDs()

// splitQuoted — deterministic command-line splitting: fields by
// spaces; an argument in double or single quotes is one argument with
// spaces, the quotes are stripped; an empty quoted argument ("")
// stays an empty argument instead of disappearing; an unbalanced
// quote is an error.
func splitQuoted(s string) ([]string, error) {
	var out []string
	var cur strings.Builder
	inQ := byte(0)
	quoted := false
	flush := func() {
		if cur.Len() > 0 || quoted {
			out = append(out, cur.String())
		}
		cur.Reset()
		quoted = false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case inQ != 0:
			if c == inQ {
				inQ = 0
			} else {
				cur.WriteByte(c)
			}
		case c == '"' || c == '\'':
			inQ = c
			quoted = true
		case c == ' ' || c == '\t':
			flush()
		default:
			cur.WriteByte(c)
		}
	}
	if inQ != 0 {
		return nil, fmt.Errorf("%s", canondata.T("canonvalid.error.unbalanced-quote"))
	}
	flush()
	return out, nil
}

// SplitCommand — the public split of one command line into argv: the
// checks table and the scenario's cmd line share one determinism.
func SplitCommand(s string) ([]string, error) { return splitQuoted(s) }

// When — an invocation of the usage surface. A black box: only the
// arguments and the observable exchange, nothing about the internals
// of the implementation.
// cmd line: a "<name> <args...>" scalar is expanded at decode into a
// cli-surface call with the default timeout; the full form (a mapping
// with stdin and its own timeout) loses no strictness.
func (w *When) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		fields, err := splitQuoted(node.Value)
		if err != nil {
			return err
		}
		if len(fields) == 0 {
			return fmt.Errorf("%s", canondata.T("canonvalid.error.when-empty"))
		}
		*w = When{Surface: "cli", Command: fields, TimeoutSec: 10}
		return nil
	}
	known := map[string]bool{"surface": true, "command": true, "stdin": true, "timeout-sec": true, "volatile": true}
	if node.Kind == yaml.MappingNode {
		for i := 0; i < len(node.Content); i += 2 {
			if !known[node.Content[i].Value] {
				fields := make([]string, 0, len(known))
				for name := range known {
					fields = append(fields, name)
				}
				sort.Strings(fields)
				return fmt.Errorf("%s", canondata.T("canonvalid.error.unknown-field",
					canondata.M{"field": node.Content[i].Value, "fields": strings.Join(fields, ", ")}))
			}
		}
	}
	var raw struct {
		Surface    string   `yaml:"surface"`
		Command    []string `yaml:"command"`
		Stdin      *string  `yaml:"stdin"`
		TimeoutSec int      `yaml:"timeout-sec"`
		Volatile   []string `yaml:"volatile,omitempty"`
	}
	if err := node.Decode(&raw); err != nil {
		return err
	}
	*w = When(raw)
	return nil
}

type When struct {
	Surface    string   `yaml:"surface"`
	Command    []string `yaml:"command"`
	Stdin      *string  `yaml:"stdin"`
	TimeoutSec int      `yaml:"timeout-sec"`
	// Volatile — observations whose exact bytes the brief declares
	// unpredictable (random identifiers, times): stream names
	// (stdout/stderr) or relative state file paths. Transience does
	// not reconcile the declared ones; file presence and everything
	// else stays byte-strict. The declaration is part of when: a
	// post-code edit goes through the human's assertion (amend).
	Volatile []string `yaml:"volatile,omitempty"`
}

// Assertion — an assertion-observation through the surface. The value
// may be a number or a string (exit code 0, for example): the decode
// coerces any scalar to a string; the json-* semantic conditions
// additionally parse it as JSON during validation.
type Assertion struct {
	Observation string `yaml:"observation"`
	Condition   string `yaml:"condition"`
	Value       string `yaml:"value"`
	Path        string `yaml:"path,omitempty"`
}

// UnmarshalYAML accepts a scalar of any type in value and coerces it
// to a string: the agent may write value: 0 without quotes. Besides
// the full form (a set of fields), an assertion arrives as compact
// shorthands of frequent patterns — the one-key mappings `fails:` and
// `state-is:`; they expand into the same typed assertions, one
// semantics.
func (a *Assertion) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.MappingNode && len(node.Content) == 2 {
		key, val := node.Content[0].Value, node.Content[1]
		switch key {
		case "out":
			// out line: an assertion about the exact stdout.
			if val.Kind != yaml.ScalarNode {
				return fmt.Errorf("%s", canondata.T("canonvalid.error.out-scalar"))
			}
			a.Observation, a.Condition = "stdout", "equals"
			a.Value = val.Value
			return nil
		case "err":
			// err line: an assertion about the exact stderr.
			if val.Kind != yaml.ScalarNode {
				return fmt.Errorf("%s", canondata.T("canonvalid.error.err-scalar"))
			}
			a.Observation, a.Condition = "stderr", "equals"
			a.Value = val.Value
			return nil
		case "rc":
			// rc line: an assertion about the exit code.
			if val.Kind != yaml.ScalarNode {
				return fmt.Errorf("%s", canondata.T("canonvalid.error.rc-scalar"))
			}
			a.Observation, a.Condition = "exit-code", "equals"
			a.Value = val.Value
			return nil
		case "fails":
			if val.Tag != "!!null" {
				return fmt.Errorf("%s", canondata.T("canonvalid.error.fails-value"))
			}
			a.Observation = "exit-code"
			a.Condition = "fails"
			return nil
		case "state-is":
			if val.Kind != yaml.ScalarNode {
				return fmt.Errorf("%s", canondata.T("canonvalid.error.state-is-scalar"))
			}
			rest := strings.TrimSpace(val.Value)
			i := strings.IndexByte(rest, ' ')
			if i <= 0 {
				return fmt.Errorf("%s", canondata.T("canonvalid.error.state-is-format"))
			}
			a.Observation = "file"
			a.Condition = "json-equals"
			a.Path = rest[:i]
			a.Value = strings.TrimSpace(rest[i+1:])
			return nil
		}
	}
	var raw struct {
		Observation string `yaml:"observation"`
		Condition   string `yaml:"condition"`
		Value       any    `yaml:"value"`
		Path        string `yaml:"path"`
	}
	if err := node.Decode(&raw); err != nil {
		return err
	}
	a.Observation = raw.Observation
	a.Condition = raw.Condition
	a.Path = raw.Path
	switch v := raw.Value.(type) {
	case nil:
		a.Value = ""
	case string:
		a.Value = v
	case int:
		a.Value = fmt.Sprintf("%d", v)
	case bool:
		a.Value = fmt.Sprintf("%t", v)
	default:
		return fmt.Errorf("%s", canondata.T("canonvalid.error.value-scalar",
			canondata.M{"type": fmt.Sprintf("%T", raw.Value)}))
	}
	return nil
}

// Scenario — an acceptance scenario: the action is an arbitrary argv
// (a surface entry point, a utility, a shell), the outcome is typed
// assertions. Pre — setup commands before the action (each must exit
// with code 0): state is built by a series of calls; the byte seed
// lives in the checks row. Free text in then is not provided for by
// the schema: a scenario is either measurable, or honestly unfinished
// (empty then), or marked as prose — the unmeasurable remainder with
// a named reason.
type Scenario struct {
	ID          string `yaml:"id"`
	Requirement string `yaml:"requirement"`
	Summary     string `yaml:"summary,omitempty"`
	// Key — the authorial repair identity of the row the scenario was
	// last submitted with: a never-green row is replaced under its key
	// even with different seeds and run; a proven one holds the
	// change-request channel.
	Key       string      `yaml:"key,omitempty"`
	Seed      []Seed      `yaml:"seed,omitempty"`
	Materials []Material  `yaml:"materials,omitempty"`
	Pre       [][]string  `yaml:"pre,omitempty"`
	When      *When       `yaml:"when"`
	Then      []Assertion `yaml:"then,omitempty"`
	Prose     *string     `yaml:"prose,omitempty"`
	// SurfaceNames — the entry point names the table names as
	// command[0] (the whole table): the runner brings them into each
	// row's run directory so that references to the product from
	// inside commands (a shell) find the files. Names without a file
	// in the project silently remain a PATH launch — the machine does
	// not know and does not decide.
	SurfaceNames []string `yaml:"surface-names,omitempty"`
}

// Interface contract statuses.
const (
	IfcDraft  = "draft"
	IfcFrozen = "frozen"
)

// Contract — an interface contract between facets of parallel work:
// the shared observable asset (a state file, a surface convention)
// fixed in terms of the usage surface. Only the machine sets the
// status: frozen — the boundary is frozen for the duration of the
// parallel work; a change only by a new version and only outside the
// work.
type Contract struct {
	ID      string   `yaml:"id"`
	Version string   `yaml:"version"`
	Sides   []string `yaml:"sides"`
	Surface []string `yaml:"surface"`
	Status  string   `yaml:"status"`
}

// Implementation card statuses.
const (
	CardActive     = "active"
	CardGreen      = "green"
	CardRed        = "red"
	CardQuarantine = "quarantine"
)

// Card — an implementation task: a product facet and the scenarios it
// closes. Files — file ownership: a split card may write only its own
// files; an empty list (the whole-product card) carries no
// restrictions — the work is sequential and there are no write
// conflicts.
type Card struct {
	ID        string   `yaml:"id"`
	Facet     string   `yaml:"facet"`
	Scenarios []string `yaml:"scenarios"`
	Files     []string `yaml:"files,omitempty"`
	Attempts  int      `yaml:"attempts"`
	Status    string   `yaml:"status"`
	// Blockers — machine reasons for red outside scenario checks: red
	// static gates (types/lint) and surface build errors, a "<gate>:
	// <reason>" line. Written only by a gate run and rebuilt on every
	// run: an empty list — no static obstacles. The executor's slot
	// must show them verbatim — otherwise the executor repairs blind.
	Blockers []string `yaml:"blockers,omitempty"`
	// Gap — the quarantine gap classification (written at the moment
	// of the transition into quarantine): canon — a contradiction of
	// spec rows, environment — all of the card's red checks are
	// environment/launch failures, code — the rest (recipes applied,
	// no contradictions). The honesty boundary: canon unenforceability
	// without a contradiction is machine-indistinguishable from a code
	// error — "code" means "what the machine can verify converged; the
	// rest is the executor's".
	Gap string `yaml:"gap,omitempty"`
}

// Quarantine gap classes.
const (
	GapCanon = "canon"
	GapCode  = "code"
	// GapEnvironment — all of the card's red checks are
	// environment/launch failures classified by the machine by failure
	// source: the behavior is not a divergence; the attempt budget is
	// not spent.
	GapEnvironment = "environment"
)

// Check run outcomes.
const (
	CheckUnknown = "unknown"
	CheckGreen   = "green"
	CheckRed     = "red"
)

// The outcome of a check's coverage probe: a green check must go red
// on a broken surface — otherwise it does not run the declared path
// and proves nothing. An empty value — the probe was not run.
const (
	PinRed   = "red"
	PinGreen = "green"
)

// Seed — a byte seed: a file with content in the run directory. The
// machine knows nothing about the meaning of the bytes.
type Seed struct {
	Path string `yaml:"path"`
	// Content — the seed bytes; Digest — the by-reference submission
	// form (the FILES WRITTEN mechanism): resolved into Content at
	// acceptance; only Content is stored in the canon.
	Content string `yaml:"content,omitempty"`
	Digest  string `yaml:"digest,omitempty"`
}

// Material — a material: a project file carried into the run
// directory byte-for-byte under the same relative path. The task's
// materials live in the project; the run row does not know the
// project root — and should not.
type Material struct {
	From string `yaml:"from"`
}

// Check — a generated check. Written only by the compiler from a
// scenario; the traceability anchor is the scenario identifier in the
// check itself, not in the code.
type Check struct {
	ID           string     `yaml:"id"`
	Scenario     string     `yaml:"scenario"`
	SourceDigest string     `yaml:"source-digest"`
	Seed         []Seed     `yaml:"seed,omitempty"`
	Materials    []Material `yaml:"materials,omitempty"`
	Pre          [][]string `yaml:"pre,omitempty"`
	When         *When      `yaml:"when"`
	// SurfaceNames — the entry point names the table names as
	// command[0] (the whole table): the runner brings them into each
	// row's run directory; names without a file in the project remain
	// a PATH launch.
	SurfaceNames []string    `yaml:"surface-names,omitempty"`
	Then         []Assertion `yaml:"then"`
	Outcome      string      `yaml:"outcome"`
	Reason       string      `yaml:"reason,omitempty"` // reason of the last outcome
	Pin          string      `yaml:"pin,omitempty"`    // coverage probe outcome of the last run
	// Proven — the check's expectations have been proven by a green
	// run at least once. Editing the expectations of a proven row is
	// the human's assertion; a row that has never gone green proves
	// nothing — its repair is free; a replacement is re-accepted by
	// the run and the probe.
	Proven bool `yaml:"proven,omitempty"`
	// EnvFail — the last red outcome is an environment or launch
	// failure (classified by the run executor by source), not an
	// expectation divergence: such reds do not eat the attempt budget.
	EnvFail bool `yaml:"env-fail,omitempty"`
}

// goal-lint dimensions: achievable / unambiguous / non-conflicting.
const (
	LintAchievable = "achievable"
	LintAmbiguous  = "ambiguous"
	LintConflict   = "conflict"
)

// LintOption — one answer option for a question to the human. Scope —
// the scenario reading this choice fixes: material for the machine's
// price of the option (bytes/4), not a promise of the result.
type LintOption struct {
	ID      string `yaml:"id"`
	Label   string `yaml:"label"`
	Scope   string `yaml:"scope"`
	Default bool   `yaml:"default"`
}

// LintFinding — one goal-lint finding: a dimension, a fact, a
// question to the human and options with a default. The machine
// assigns question identifiers in order: Q1, Q2, ...
type LintFinding struct {
	ID        string       `yaml:"id"`
	Dimension string       `yaml:"dimension"`
	Finding   string       `yaml:"finding"`
	Question  string       `yaml:"question"`
	Options   []LintOption `yaml:"options"`
}

// LintResult — the goal-lint result: findings turned into questions
// for the human. An empty list — the goal is clear, no questions.
type LintResult struct {
	Findings []LintFinding `yaml:"findings"`
}

// Clarification — a resolved goal-lint question: the chosen option
// (or the default) with its scenario. The spec author sees it as a
// fact of choice.
type Clarification struct {
	Question  string `yaml:"question"`
	Dimension string `yaml:"dimension"`
	Option    string `yaml:"option"`
	Label     string `yaml:"label"`
	Scope     string `yaml:"scope"`
	ByDefault bool   `yaml:"by-default"`
}

// SchemaVersion — the canon schema version. Changes when the formats
// change; migrations go from version to version, none may be skipped.
const SchemaVersion = 1

// Impact — the estimate of the last spec change: what was regenerated
// and what is untouched. Computed by the check compiler; the machine
// shows it before the translation starts — the cost of a change is
// visible before the work, not after.
type Impact struct {
	ScenariosTouched  int `yaml:"scenarios-touched"`
	ChecksRegenerated int `yaml:"checks-regenerated"`
	ChecksKept        int `yaml:"checks-kept"`
	ChecksRemoved     int `yaml:"checks-removed"`
}

// ProbeFileFact — an observed file after a behavior probe: the path
// and the content verbatim.
type ProbeFileFact struct {
	Path    string `yaml:"path"`
	Content string `yaml:"content"`
}

// ProbeObservation — a machine observation of surface behavior: the
// command, its observable outcomes (code, streams, changed files) and
// the determinism verdict over two runs. Material for the machine's
// spec proposal; not a judgment — facts.
type ProbeObservation struct {
	Command       []string        `yaml:"command"`
	Seed          []Seed          `yaml:"seed,omitempty"`
	ExitCode      int             `yaml:"exit-code"`
	Stdout        string          `yaml:"stdout"`
	Stderr        string          `yaml:"stderr"`
	Files         []ProbeFileFact `yaml:"files,omitempty"`
	Deterministic bool            `yaml:"deterministic"`
	// Failed — the run did not happen (surface not built, timeout):
	// no observations, an honest reason instead.
	Failed bool   `yaml:"failed,omitempty"`
	Reason string `yaml:"reason,omitempty"`
}

// State — the instance's state. Always derivable from the journal;
// the file is a cache — on a divergence the journal wins.
type State struct {
	SchemaVersion  int     `yaml:"schema-version"`
	Stage          string  `yaml:"stage"`
	Intent         *string `yaml:"intent,omitempty"`
	IntentAccepted bool    `yaml:"intent-accepted"`
	// Language — the wish's language from the deterministic script
	// detector (en/ru/zh); it steers the human blocks, while the
	// machine protocol stays English always.
	Language string `yaml:"language,omitempty"`
	// Lint — the goal-lint result: nil — the lint has not been
	// submitted yet; empty findings — the goal is clear, no questions
	// for the human.
	Lint *LintResult `yaml:"lint,omitempty"`
	// Clarified — the questions (if any) are resolved: by explicit
	// answers or by defaults.
	Clarified      bool            `yaml:"clarified"`
	Clarifications []Clarification `yaml:"clarifications,omitempty"`
	Touched        []string        `yaml:"touched,omitempty"`
	Counters       map[string]int  `yaml:"counters"`
	Impact         *Impact         `yaml:"impact,omitempty"`
	// RetiredGates — gates retired by lease: per the ledger they did
	// not affect outcomes (at least three runs, all green) and are not
	// run anymore.
	RetiredGates []string `yaml:"retired-gates,omitempty"`
	// Probes — machine observations of behavior from the deliver
	// stage: the spec catches up with stabilized code.
	Probes []ProbeObservation `yaml:"probes,omitempty"`
	// PendingAmend — a deferred change to the expectations of
	// already-run checks: a submission after code start that changes
	// the expectations of a scenario with a run is applied only by
	// the human's approval (kind: amend). Silence (the next non-amend
	// submission) dismisses the change — already-run expectations are
	// not weakened by silence.
	PendingAmend *PendingAmend `yaml:"pending-amend,omitempty"`
	// Waives — active waives of completeness lint rules with a reason:
	// the rule stops blocking the transition but stays
	// visible — the verdict's taste queue carries it with its reason.
	Waives []WaiveRecord `yaml:"waives,omitempty"`
}

// WaiveRecord — one lint rule waive: the rule, the waive target (a
// requirement identifier or a command family) and the reason in the
// human's words. Only the machine writes it (a kind: waive delta).
type WaiveRecord struct {
	Rule   string `yaml:"rule"`
	Target string `yaml:"target"`
	Reason string `yaml:"reason"`
}

// PendingAmend — the deferred change itself: the raw delta in full
// and the list of changes for the slot. Only the machine writes and
// dismisses it.
type PendingAmend struct {
	Kind string `yaml:"kind"` // the deferred delta's kind (spec/checks)
	Key  string `yaml:"key"`  // the submission key that produced the change
	Raw  string `yaml:"raw"`  // the raw delta verbatim
	// Staging — the staging position in the journal (the journal
	// length at staging time): the identity of a hold. Two holds of
	// the same content are different stagings (different positions):
	// an anonymous amend reply names the hold it answers — identical
	// replies to different holds do not collide by key.
	Staging string `yaml:"staging"`
	// Changes — what changes, one line per change: the machine states
	// the fact (the scenario, what exactly), not a judgment.
	Changes []string `yaml:"changes"`
	// Plan — the machine's apply plan (change request effects:
	// closing cards, unfreezing contracts, a stage rollback): facts of
	// the apply, not the changes themselves — the "N pending changes"
	// counter counts Changes.
	Plan []string `yaml:"plan,omitempty"`
}
