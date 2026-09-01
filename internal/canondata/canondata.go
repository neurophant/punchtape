// Package canondata — the delivery's canon data: the texts of all
// the machine's surfaces and the numeric thresholds live as data
// inside the binary; the code is only keys and render mechanics. The
// canon is immutable by construction: editing a text means releasing
// a new version of the application.
package canondata

import (
	"embed"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed data/*.yaml
var dataFS embed.FS

// M — canon text render parameters: name → value.
type M map[string]string

var (
	texts         = map[string]string{}
	limits        = map[string]int{}
	painCatalog   []PainEntry
	stageList     []ScenarioStage
	trustSteps    []TrustStatus
	trustMoves    []TrustTransition
	lintRules     []LintRule
	envDiagList   []EnvDiagEntry
	intentCatalog []IntentRow
	healCatalog   []HealEntry
	hypList       []HypothesisEntry
)

// PainEntry — a pain catalog entry: an identifier (it appears as a
// literal in the detection code — the mirror reconciler catches dead
// entries), a diagnosis and a mandatory workaround (an empty
// workaround is rendered by the finding itself).
type PainEntry struct {
	ID         string `yaml:"id"`
	Diagnosis  string `yaml:"diagnosis"`
	Workaround string `yaml:"workaround"`
}

// StageExit — a process stage's exit: a machine check by name, the
// transition stage and a machine action on entry (may be empty).
type StageExit struct {
	Gate  string `yaml:"gate"`
	Goto  string `yaml:"goto"`
	Entry string `yaml:"entry,omitempty"`
}

// ScenarioStage — a stage of the process state machine: human
// choice points hold the stage; the exit is executed by the runner
// by a named check. Ambiguities are resolved by the human from
// outside — the runner does not resolve them.
type ScenarioStage struct {
	ID      string    `yaml:"id"`
	Choices []string  `yaml:"choices,omitempty"`
	Exit    StageExit `yaml:"exit"`
}

// TrustStatus — one step of the boundary trust ladder.
type TrustStatus struct {
	ID     string `yaml:"id"`
	Weight int    `yaml:"weight"`
}

// TrustTransition — a machine transition of the ladder: a proof by
// name and the target status. No proofs above "verified" exist.
type TrustTransition struct {
	Proof string `yaml:"proof"`
	To    string `yaml:"to"`
}

// LintRule — an artifact lint rule: the message is the single
// source (the registry); waivable — a completeness rule that can be
// waived with a reason; enforceability and stop-conditions are not
// waivable.
type LintRule struct {
	ID       string `yaml:"id"`
	Section  string `yaml:"section"`
	Message  string `yaml:"message"`
	Waivable bool   `yaml:"waivable,omitempty"`
}

// EnvDiagEntry — an environment diagnosis: a deterministically
// classified failure to execute an external command; the line =
// diagnosis + FIX hint.
type EnvDiagEntry struct {
	ID        string `yaml:"id"`
	Diagnosis string `yaml:"diagnosis"`
	Fix       string `yaml:"fix"`
}

// HealEntry — a healing phrase: a predictable agent failure,
// detectable by the machine, and the standard data-delivered
// response.
type HealEntry struct {
	ID     string `yaml:"id"`
	Advice string `yaml:"advice"`
}

// HypothesisEntry — a canon change hypothesis: a canon edit ships
// only as a new delivery version, each with a testable hypothesis;
// not confirmed — a rollback, the outcome journaled in the outcome
// field. A content registry of the delivery process (like intents):
// lives as data, structurally validated by the loader, needs no
// by-name code reference — there are no dead entries; the registry
// is read by the human.
type HypothesisEntry struct {
	ID         string `yaml:"id"`
	Change     string `yaml:"change"`
	Hypothesis string `yaml:"hypothesis"`
	Metric     string `yaml:"metric"`
	Window     string `yaml:"window"`
	Baseline   string `yaml:"baseline"`
	Outcome    string `yaml:"outcome"`
}

// Heal — advice by detection identifier; absence is an honest empty
// reply (no advice — no invention).
func Heal(id string) string {
	for _, h := range healCatalog {
		if h.ID == id {
			return h.Advice
		}
	}
	return ""
}

// IntentRow — an intent table row: an everyday phrase (all keys
// among the phrase's tokens) routed to a machine verb; topic — a
// why topic.
type IntentRow struct {
	ID    string   `yaml:"id"`
	Keys  []string `yaml:"keys"`
	Verb  string   `yaml:"verb"`
	Topic string   `yaml:"topic,omitempty"`
}

// Pain — a catalog entry by detection identifier.
func Pain(id string) (PainEntry, bool) {
	for _, p := range painCatalog {
		if p.ID == id {
			return p, true
		}
	}
	return PainEntry{}, false
}

// StagesByID — the process stages keyed by identifier.
func StagesByID() map[string]ScenarioStage {
	out := make(map[string]ScenarioStage, len(stageList))
	for _, s := range stageList {
		out[s.ID] = s
	}
	return out
}

// StageIDs — stage identifiers in data order: the stage order is
// derived from the process; the code keeps no separate list.
func StageIDs() []string {
	out := make([]string, 0, len(stageList))
	for _, s := range stageList {
		out = append(out, s.ID)
	}
	return out
}

// TrustWeight — a ladder status's weight; absence — a status
// outside the ladder.
func TrustWeight(id string) (int, bool) {
	for _, s := range trustSteps {
		if s.ID == id {
			return s.Weight, true
		}
	}
	return 0, false
}

// TrustTop — the weight of the ladder's top step.
func TrustTop() int {
	top := 0
	for _, s := range trustSteps {
		if s.Weight > top {
			top = s.Weight
		}
	}
	return top
}

// TrustTarget — the target of a machine transition by proof name;
// absence — the proof does not exist, the machine does not guess.
func TrustTarget(proof string) (string, bool) {
	for _, t := range trustMoves {
		if t.Proof == proof {
			return t.To, true
		}
	}
	return "", false
}

// LintRuleByID — a registry rule by identifier.
func LintRuleByID(id string) (LintRule, bool) {
	for _, r := range lintRules {
		if r.ID == id {
			return r, true
		}
	}
	return LintRule{}, false
}

// LintMessage renders a rule message with parameter substitution:
// the registry is the single source of rule text.
func LintMessage(id string, params ...M) string {
	r, ok := LintRuleByID(id)
	if !ok {
		panic(fmt.Sprintf("canondata: unknown lint rule %q", id))
	}
	return substitute(r.Message, params...)
}

// EnvDiag renders the diagnosed environment line: the diagnosis and
// the FIX hint in one piece — so an environment failure carries
// both the cause and the action.
func EnvDiag(id string, params ...M) string {
	var e EnvDiagEntry
	for _, c := range envDiagList {
		if c.ID == id {
			e = c
			break
		}
	}
	if e.ID == "" {
		panic(fmt.Sprintf("canondata: unknown env diagnosis %q", id))
	}
	diag := substitute(e.Diagnosis, params...)
	fix := substitute(e.Fix, params...)
	return diag + "; FIX: " + fix
}

// Intents — the intent table rows in data order.
func Intents() []IntentRow { return intentCatalog }

// substitute — {name} substitution into a registry template: the
// same mechanics as for canon texts.
func substitute(tmpl string, params ...M) string {
	if len(params) == 0 {
		return tmpl
	}
	pairs := make([]string, 0, 2*len(params[0]))
	for name, value := range params[0] {
		pairs = append(pairs, "{"+name+"}", value)
	}
	return strings.NewReplacer(pairs...).Replace(tmpl)
}

func init() {
	names, err := dataFS.ReadDir("data")
	if err != nil {
		panic(fmt.Sprintf("canondata: %v", err))
	}
	for _, name := range names {
		raw, err := dataFS.ReadFile("data/" + name.Name())
		if err != nil {
			panic(fmt.Sprintf("canondata: %v", err))
		}
		switch name.Name() {
		case "limits.yaml":
			if err := yaml.Unmarshal(raw, &limits); err != nil {
				panic(fmt.Sprintf("canondata: limits: %v", err))
			}
		case "pains.yaml":
			if err := yaml.Unmarshal(raw, &painCatalog); err != nil {
				panic(fmt.Sprintf("canondata: pains: %v", err))
			}
		case "scenarios.yaml":
			var doc struct {
				Process string          `yaml:"process"`
				Stages  []ScenarioStage `yaml:"stages"`
			}
			if err := yaml.Unmarshal(raw, &doc); err != nil {
				panic(fmt.Sprintf("canondata: scenarios: %v", err))
			}
			stageList = doc.Stages
		case "trust.yaml":
			var doc struct {
				Ladder      []TrustStatus     `yaml:"ladder"`
				Transitions []TrustTransition `yaml:"transitions"`
			}
			if err := yaml.Unmarshal(raw, &doc); err != nil {
				panic(fmt.Sprintf("canondata: trust: %v", err))
			}
			trustSteps, trustMoves = doc.Ladder, doc.Transitions
		case "lint.yaml":
			if err := yaml.Unmarshal(raw, &lintRules); err != nil {
				panic(fmt.Sprintf("canondata: lint: %v", err))
			}
		case "envdiag.yaml":
			if err := yaml.Unmarshal(raw, &envDiagList); err != nil {
				panic(fmt.Sprintf("canondata: envdiag: %v", err))
			}
		case "intents.yaml":
			if err := loadIntents(raw); err != nil {
				panic(fmt.Sprintf("canondata: intents: %v", err))
			}
		case "heal.yaml":
			if err := yaml.Unmarshal(raw, &healCatalog); err != nil {
				panic(fmt.Sprintf("canondata: heal: %v", err))
			}
		case "hypotheses.yaml":
			var doc struct {
				Hypotheses []HypothesisEntry `yaml:"hypotheses"`
			}
			if err := yaml.Unmarshal(raw, &doc); err != nil {
				panic(fmt.Sprintf("canondata: hypotheses: %v", err))
			}
			if err := validateHypotheses(doc.Hypotheses); err != nil {
				panic(fmt.Sprintf("canondata: hypotheses: %v", err))
			}
			hypList = doc.Hypotheses
		default:
			var part map[string]string
			if err := yaml.Unmarshal(raw, &part); err != nil {
				panic(fmt.Sprintf("canondata: %s: %v", name.Name(), err))
			}
			for k, v := range part {
				if _, dup := texts[k]; dup {
					panic(fmt.Sprintf("canondata: duplicate key %s", k))
				}
				texts[k] = v
			}
		}
	}
	validateStructural()
}

// loadIntents reads the router's content rows and validates their
// structure: the verb is one of the known ones, a topic only with
// why, the keys non-empty and unique by composition.
func loadIntents(raw []byte) error {
	if err := yaml.Unmarshal(raw, &intentCatalog); err != nil {
		return err
	}
	seen := map[string]bool{}
	for _, row := range intentCatalog {
		switch row.Verb {
		case "next", "why":
		default:
			return fmt.Errorf("row %s: verb %q invalid", row.ID, row.Verb)
		}
		if row.Topic != "" && row.Verb != "why" {
			return fmt.Errorf("row %s: topic invalid", row.ID)
		}
		if len(row.Keys) == 0 {
			return fmt.Errorf("row %s: keys empty", row.ID)
		}
		key := strings.Join(row.Keys, "\x00")
		if seen[key] {
			return fmt.Errorf("duplicate keys: %s", row.ID)
		}
		seen[key] = true
	}
	return nil
}

// validateHypotheses — structural validation of the hypothesis
// registry: complete fields, unique identifiers, an outcome from
// the closed set. Data drift is caught here, before any work.
func validateHypotheses(list []HypothesisEntry) error {
	seen := map[string]bool{}
	for _, h := range list {
		if h.ID == "" || h.Change == "" || h.Hypothesis == "" ||
			h.Metric == "" || h.Window == "" || h.Baseline == "" {
			return fmt.Errorf("hypothesis %s: incomplete entry", h.ID)
		}
		switch h.Outcome {
		case "open", "confirmed", "refuted":
		default:
			return fmt.Errorf("hypothesis %s: outcome %q invalid", h.ID, h.Outcome)
		}
		if seen[h.ID] {
			return fmt.Errorf("hypothesis duplicate: %s", h.ID)
		}
		seen[h.ID] = true
	}
	return nil
}

// Hypotheses — the delivery's hypothesis registry.
func Hypotheses() []HypothesisEntry { return hypList }

// validateStructural — cross-registry load invariants: ladder
// statuses are unique and grow by weight, transitions lead to ladder
// statuses, process stages lead to known stages, lint rules are
// unique. Data drift is caught here, before any work.
func validateStructural() {
	weights := map[string]bool{}
	prev := 0
	for _, s := range trustSteps {
		if weights[s.ID] {
			panic(fmt.Sprintf("trust status duplicate: %s", s.ID))
		}
		weights[s.ID] = true
		if s.Weight <= prev {
			panic(fmt.Sprintf("trust weights: %s", s.ID))
		}
		prev = s.Weight
	}
	for _, t := range trustMoves {
		if !weights[t.To] {
			panic(fmt.Sprintf("trust transition %s: invalid", t.Proof))
		}
	}
	stageIDs := map[string]bool{}
	for _, s := range stageList {
		if stageIDs[s.ID] {
			panic(fmt.Sprintf("stage duplicate: %s", s.ID))
		}
		stageIDs[s.ID] = true
	}
	for _, s := range stageList {
		if !stageIDs[s.Exit.Goto] {
			panic(fmt.Sprintf("stage %s: exit %s unknown", s.ID, s.Exit.Goto))
		}
	}
	ruleIDs := map[string]bool{}
	for _, r := range lintRules {
		if ruleIDs[r.ID] {
			panic(fmt.Sprintf("canondata: duplicate lint rule %s", r.ID))
		}
		ruleIDs[r.ID] = true
	}
}

// T renders canon text by key with {name} parameter substitution. An
// unknown key is a programming error (code and canon diverge): it is
// deterministic and caught by mirror reconciliation before a
// release, so an honest fail-fast, not a silent stub.
func T(key string, params ...M) string {
	tmpl, ok := texts[key]
	if !ok {
		panic(fmt.Sprintf("canondata: unknown key %q", key))
	}
	if len(params) == 0 {
		return tmpl
	}
	pairs := make([]string, 0, 2*len(params[0]))
	for name, value := range params[0] {
		pairs = append(pairs, "{"+name+"}", value)
	}
	return strings.NewReplacer(pairs...).Replace(tmpl)
}

// TFor renders canon text in the dialogue language: the key with a
// ".<lang>" suffix when present, otherwise the base key. The machine
// protocol calls T (without a language); human blocks — TFor.
func TFor(lang, key string, params ...M) string {
	if lang != "" && lang != "en" {
		if _, ok := texts[key+"."+lang]; ok {
			return T(key+"."+lang, params...)
		}
	}
	return T(key, params...)
}

// Limit — a canon numeric threshold by key; thresholds live as
// delivery data and are not duplicated in code.
func Limit(key string) int {
	v, ok := limits[key]
	if !ok {
		panic(fmt.Sprintf("canondata: unknown limit %q", key))
	}
	return v
}
