// Package canondata — the delivery's canon data: the texts of all
// the machine's surfaces and the numeric thresholds live as data
// inside the binary; the code is only keys and render mechanics. The
// canon is immutable by construction: editing a text means releasing
// a new version of the application.
package canondata

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"github.com/neurophant/punchtape/internal/yamlio"
)

//go:embed data
var dataFS embed.FS

// M — canon text render parameters: name → value.
type M map[string]string

var (
	texts       = map[string]string{}
	templates   = map[string][]byte{}
	limits      = map[string]int{}
	stageList   []Stage
	lintRules   []LintRule
	envDiagList []EnvDiagEntry
	healCatalog []HealEntry
)

// StageExit — a process stage's exit: a machine check by name, the
// transition stage and a machine action on entry (may be empty).
type StageExit struct {
	Gate  string `yaml:"gate"`
	Goto  string `yaml:"goto"`
	Entry string `yaml:"entry,omitempty"`
}

// Stage — a stage of the process state machine: human
// choice points hold the stage; the exit is executed by the runner
// by a named check. Ambiguities are resolved by the human from
// outside — the runner does not resolve them.
type Stage struct {
	ID      string    `yaml:"id"`
	Choices []string  `yaml:"choices,omitempty"`
	Exit    StageExit `yaml:"exit"`
}

// LintRule — an artifact lint rule: the message is the single
// source (the registry).
type LintRule struct {
	ID      string `yaml:"id"`
	Section string `yaml:"section"`
	Message string `yaml:"message"`
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

// StagesByID — the process stages keyed by identifier.
func StagesByID() map[string]Stage {
	out := make(map[string]Stage, len(stageList))
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
	return substitute(e.Diagnosis, params...) + "; FIX: " + substitute(e.Fix, params...)
}

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

func init() { mustLoad(dataFS) }

// mustLoad — the delivered-tree load of the package init; the panic
// branch is reachable over a broken tree in tests.
func mustLoad(fsys fs.FS) {
	if err := loadFS(fsys); err != nil {
		panic("canondata: " + err.Error())
	}
}

// loadFS — the data load over any fs (the embed tree in production;
// synthetic trees in tests): every branch of a broken tree is a
// named load error, caught before any work. The top level carries
// the structural registries and the limits (YAML, one-line values);
// texts/ is the tree of canon texts (a namespace dictionary per
// file at the top, one text document per file under ref/, the
// path is the key with dots as separators); templates/ is the raw
// template library (path → bytes, the prefill source of truth).
func loadFS(fsys fs.FS) error {
	names, err := fs.ReadDir(fsys, "data")
	if err != nil {
		return err
	}
	for _, name := range names {
		if name.IsDir() {
			continue
		}
		raw, err := fs.ReadFile(fsys, "data/"+name.Name())
		if err != nil {
			return err
		}
		if err := loadPart(name.Name(), raw); err != nil {
			return fmt.Errorf("%s: %w", name.Name(), err)
		}
	}
	if err := loadTexts(fsys); err != nil {
		return err
	}
	if err := loadTemplates(fsys); err != nil {
		return err
	}
	return validateStructuralErr()
}

// loadTexts — the canon texts in two forms, split by role (the role
// is the layout, not a size judgment): data/texts/<ns>.yaml are
// namespace DICTIONARIES — flat maps of console lines, one literal
// block per key, the block's lines are the text verbatim with no
// edge newlines (the composition owns spacing); data/texts/ref/**
// are DOCUMENTS — whole text documents (the agents contract, the
// reference docs) whose bytes are the text verbatim. Any other
// subdirectory is a load error: no silent hybrid zones.
func loadTexts(fsys fs.FS) error {
	return fs.WalkDir(fsys, "data/texts", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if p == "data/texts" && errors.Is(err, fs.ErrNotExist) {
				return nil // an empty text tree is a legal state (tests)
			}
			return err
		}
		if !d.IsDir() {
			rel := strings.TrimPrefix(strings.TrimPrefix(p, "data/texts"), "/")
			if !strings.Contains(rel, "/") {
				if !strings.HasSuffix(rel, ".yaml") {
					return fmt.Errorf("texts/%s: not a dictionary file", rel)
				}
				raw, rerr := fs.ReadFile(fsys, p)
				if rerr != nil {
					return rerr
				}
				segs, derr := yamlio.DecodeDict(raw)
				if derr != nil {
					return fmt.Errorf("texts/%s: %w", rel, derr)
				}
				ns := strings.TrimSuffix(rel, ".yaml")
				if ns == "" {
					return fmt.Errorf("texts/%s: not a dictionary file", rel)
				}
				for key, val := range segs {
					texts[ns+"."+key] = val
				}
				return nil
			}
			if !strings.HasSuffix(p, ".md") {
				return fmt.Errorf("texts/%s: not a text file", rel)
			}
			raw, rerr := fs.ReadFile(fsys, p)
			if rerr != nil {
				return rerr
			}
			key := strings.ReplaceAll(strings.TrimSuffix(rel, ".md"), "/", ".")
			if strings.Contains(key, "..") {
				return fmt.Errorf("texts/%s: not a key path", rel)
			}
			texts[key] = string(raw)
			return nil
		}
		if p != "data/texts" && !strings.HasPrefix(p, "data/texts/ref") {
			return fmt.Errorf("texts/%s: unknown zone (dictionaries at the top, documents under ref/)", p)
		}
		return nil
	})
}

// loadTemplates — the template tree: every file's bytes under its
// relative path. The skeleton files are pretty YAML, the content
// stubs are text — the loader does not care: a template is bytes the
// machine hands out verbatim.
func loadTemplates(fsys fs.FS) error {
	return fs.WalkDir(fsys, "data/templates", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if p == "data/templates" && errors.Is(err, fs.ErrNotExist) {
				return nil // an empty template tree is a legal state (tests)
			}
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel := strings.TrimPrefix(strings.TrimPrefix(p, "data/templates"), "/")
		raw, err := fs.ReadFile(fsys, p)
		if err != nil {
			return err
		}
		templates[rel] = raw
		return nil
	})
}

// loadPart — one structural registry (or the limits) from the top
// level of the data tree; a named error per broken part. Everything
// at the top level is structure with one-line values: texts live in
// texts/, templates in templates/. The field tags keep the grammar's
// field names; the yamlio layer reads them through the armor.
func loadPart(name string, raw []byte) error {
	switch name {
	case "limits.yaml":
		return yamlio.DecodeStrict(raw, &limits)
	case "stages.yaml":
		var doc struct {
			Process string  `yaml:"process"`
			Stages  []Stage `yaml:"stages"`
		}
		if err := yamlio.DecodeStrict(raw, &doc); err != nil {
			return err
		}
		stageList = doc.Stages
	case "lint.yaml":
		return yamlio.DecodeStrict(raw, &lintRules)
	case "envdiag.yaml":
		return yamlio.DecodeStrict(raw, &envDiagList)
	case "heal.yaml":
		return yamlio.DecodeStrict(raw, &healCatalog)
	default:
		return fmt.Errorf("unknown data part")
	}
	return nil
}

// validateStructural — cross-registry load invariants: process
// stages lead to known stages, lint rules are unique. Data drift is
// caught here, before any work.
func validateStructural() {
	if err := validateStructuralErr(); err != nil {
		panic(fmt.Sprintf("canondata: %v", err))
	}
}

// validateStructuralErr — the cross-registry invariants as named
// errors (the synthetic-tree load path of tests).
func validateStructuralErr() error {
	ids := map[string]bool{}
	for _, s := range stageList {
		if ids[s.ID] {
			return fmt.Errorf("stage duplicate: %s", s.ID)
		}
		ids[s.ID] = true
	}
	for _, s := range stageList {
		if s.Exit.Gate == "" {
			continue
		}
		if !ids[s.Exit.Goto] {
			return fmt.Errorf("stage exit %s: unknown target %s", s.ID, s.Exit.Goto)
		}
	}
	rules := map[string]bool{}
	for _, r := range lintRules {
		if rules[r.ID] {
			return fmt.Errorf("lint rule duplicate: %s", r.ID)
		}
		rules[r.ID] = true
	}
	return nil
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

// TemplateBytes — the raw bytes of a template file by its path under
// data/templates/ (pretty skeletons, content stubs). The bytes are
// the machine's single source of truth for what it hands the agent:
// the file in the source tree, in the binary, in the instance — one
// point of truth, byte for byte. An unknown name is a programming
// error (code and canon diverge): an honest fail-fast, the same
// contract as T.
func TemplateBytes(name string) []byte {
	data, ok := templates[name]
	if !ok {
		panic(fmt.Sprintf("canondata: unknown template %q", name))
	}
	return data
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

// TemplateSet — every template file under a path prefix (an inbox
// set's template is its manifest plus its content stubs): the name
// is the path under data/templates/, the value the embedded bytes.
func TemplateSet(prefix string) map[string][]byte {
	out := map[string][]byte{}
	for name, data := range templates {
		if strings.HasPrefix(name, prefix) {
			out[name] = data
		}
	}
	return out
}
