// Package canon — the canon entities and their store. The canon is the
// single source of truth about an instance; only the core writes the
// files, decoding is strict: an unknown field is an error. Field names
// and enumeration values are English: this is the machine layer.
//
// The model: the corpus is behaviors — one atom record per
// behavior (id, name, prose, rows); the rows live inside the atom
// The atom is the unit of the verdict: an atom is green
// when all its rows are green.
package canon

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// Instance stages. The stage order is a property of the data process
// (the delivery's stages.yaml): the code only lists the values.
const (
	StageIntake    = "intake"
	StageImplement = "implement"
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
// row run form shares one determinism.
func SplitCommand(s string) ([]string, error) { return splitQuoted(s) }

// Cmd — a row's command in two equal forms: a "command args" string
// (whitespace splitting with quotes) or a ready argv list — with no
// splitting at all (an argument with commas/quotes travels as is).
// The form is only notation; Fields() is the truth.
type Cmd struct {
	Line   string
	Argv   []string
	IsArgv bool
}

// UnmarshalYAML accepts a scalar (a command line) or a sequence of
// scalars (argv).
func (rc *Cmd) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// an interface{} probe decode cannot fail under the armor
	var probe interface{}
	_ = unmarshal(&probe)
	switch v := probe.(type) {
	case string:
		rc.Line = v
		return nil
	case []interface{}:
		argv := make([]string, 0, len(v))
		for _, el := range v {
			s, ok := el.(string)
			if !ok {
				return fmt.Errorf("%s", canondata.T("delta.error.run-argv-scalar"))
			}
			argv = append(argv, s)
		}
		rc.Argv = argv
		rc.IsArgv = true
		return nil
	}
	return fmt.Errorf("%s", canondata.T("delta.error.run-form"))
}

// MarshalYAML — the canonical form: a string as a scalar, argv as a
// list; a two-way rewrite of the same form. The product rides the
// yamlio exactness pass: a control byte in a command line survives.
func (rc Cmd) MarshalYAML() (interface{}, error) {
	if rc.IsArgv {
		return yamlio.Wrap(rc.Argv)
	}
	return yamlio.Wrap(rc.Line)
}

// Fields — the command's argv fields: a ready list or a string split.
func (rc Cmd) Fields() ([]string, error) {
	if rc.IsArgv {
		if len(rc.Argv) == 0 {
			return nil, fmt.Errorf("%s", canondata.T("delta.error.run-argv-empty"))
		}
		return rc.Argv, nil
	}
	return splitQuoted(rc.Line)
}

// Form — the element's form for human renders: the string as is,
// argv as a flow list (a resubmission reads in the same form).
func (rc Cmd) Form() string {
	if !rc.IsArgv {
		return rc.Line
	}
	parts := make([]string, len(rc.Argv))
	for i, a := range rc.Argv {
		if a == "" || strings.ContainsAny(a, " \t,[]{}:\"'#") {
			parts[i] = "\"" + strings.ReplaceAll(strings.ReplaceAll(a, "\\", "\\\\"), "\"", "\\\"") + "\""
		} else {
			parts[i] = a
		}
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// containsPrefix — the fragment form marker of a text expectation:
// "(contains) X" — the fragment X must occur; any other spelling is
// the exact form. The loose variant "(contains-loose) X" is the SAME
// fragment form carrying the executor's explicit mark that the
// fragment is short and imprecise (a short fragment is
// legal only under its honest name).
const (
	containsPrefix      = "(contains) "
	containsLoosePrefix = "(contains-loose) "
)

// TextExpect — one text expectation: exact (EOL-tolerant: exactly one
// trailing newline is ignored on either side) or a fragment that must
// occur. An empty text is the explicit empty stream/file.
type TextExpect struct {
	Contains bool
	Text     string
	// Loose — the fragment form's explicit mark: the executor names
	// the fragment short on purpose; rides only in the loose prefix.
	Loose bool
}

// UnmarshalYAML — any scalar value; a string starting with the
// fragment marker becomes the contains form (the fragment after the
// marker must be non-empty; the loose marker additionally sets the
// loose mark).
func (t *TextExpect) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var probe interface{}
	_ = unmarshal(&probe)
	s, err := scalarText(probe)
	if err != nil {
		return fmt.Errorf("%s", canondata.T("canonvalid.error.expect-scalar"))
	}
	if strings.HasPrefix(s, containsLoosePrefix) {
		rest := strings.TrimPrefix(s, containsLoosePrefix)
		if rest == "" {
			return fmt.Errorf("%s", canondata.T("canonvalid.error.expect-contains-empty"))
		}
		t.Contains, t.Loose, t.Text = true, true, rest
		return nil
	}
	if strings.HasPrefix(s, containsPrefix) {
		rest := strings.TrimPrefix(s, containsPrefix)
		if rest == "" {
			return fmt.Errorf("%s", canondata.T("canonvalid.error.expect-contains-empty"))
		}
		t.Contains, t.Text = true, rest
		return nil
	}
	t.Text = s
	return nil
}

// MarshalYAML — the same one-line spelling the submission used. A
// nil expectation is the absent field (the encoder's null).
func (t *TextExpect) MarshalYAML() (interface{}, error) {
	if t == nil {
		return nil, nil
	}
	if t.Contains && t.Loose {
		return yamlio.Wrap(containsLoosePrefix + t.Text)
	}
	if t.Contains {
		return yamlio.Wrap(containsPrefix + t.Text)
	}
	return yamlio.Wrap(t.Text)
}

// Form — the expectation's one-line face for renders.
func (t TextExpect) Form() string {
	if t.Contains && t.Loose {
		return containsLoosePrefix + t.Text
	}
	if t.Contains {
		return containsPrefix + t.Text
	}
	return t.Text
}

// Seed — a row's seed: a file written into the work tree before the
// row and removed after it. Exactly one source form: Text — inline
// content (the document's own bytes), FromInput — a copy of a file
// of the input directory (the machine copies the file as is).
type Seed struct {
	Path string `yaml:"path"`
	Text string `yaml:"text,omitempty"`
	// FromInput — the path inside the instance's input directory; the
	// bytes are copied file-to-file at row time.
	FromInput string `yaml:"from-input,omitempty"`
}

// UnmarshalYAML — path plus exactly one source form.
func (s *Seed) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var probe map[string]interface{}
	if err := unmarshal(&probe); err != nil {
		return err
	}
	for name := range probe {
		switch name {
		case "path", "text", "from-input":
		default:
			return fmt.Errorf("%s", canondata.T("delta.error.seed-field",
				canondata.M{"field": name}))
		}
	}
	type plain Seed
	var p plain
	if err := unmarshal(&p); err != nil {
		return err
	}
	*s = Seed(p)
	forms := 0
	if s.Text != "" {
		forms++
	}
	if s.FromInput != "" {
		forms++
	}
	if forms != 1 || s.Path == "" {
		return fmt.Errorf("%s", canondata.T("delta.error.seed-form"))
	}
	return nil
}

// StateWant — one state expectation of a row: the observed path and
// exactly one of the two forms — Text (exact EOL-tolerant or a
// fragment, the same forms as out) or FromRun (byte equality with a
// file the run produced).
type StateWant struct {
	Path    string      `yaml:"path"`
	Text    *TextExpect `yaml:"text,omitempty"`
	FromRun string      `yaml:"from-run,omitempty"`
}

// UnmarshalYAML — path plus exactly one expectation form.
func (s *StateWant) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var probe map[string]interface{}
	if err := unmarshal(&probe); err != nil {
		return err
	}
	for name := range probe {
		switch name {
		case "path", "text", "from-run":
		default:
			return fmt.Errorf("%s", canondata.T("delta.error.state-field",
				canondata.M{"field": name}))
		}
	}
	type plain StateWant
	var p plain
	if err := unmarshal(&p); err != nil {
		return err
	}
	*s = StateWant(p)
	forms := 0
	if s.Text != nil {
		forms++
	}
	if s.FromRun != "" {
		forms++
	}
	if forms != 1 || s.Path == "" {
		return fmt.Errorf("%s", canondata.T("delta.error.state-form"))
	}
	return nil
}

// Row run outcomes.
const (
	RowUnknown = "unknown"
	RowGreen   = "green"
	RowRed     = "red"
)

// Row — one check row inside an atom. Run — the one command (argv or
// a quoted string; orchestration inside the command via sh -c is the
// row's own business); Seeds — files before the row (self-removed
// after); Out/Err — the stream expectations (nil — not checked);
// Rc — the expected exit code (nil — 0); State — file expectations
// after the row; Bytes — a sample file reference (file://name — the
// sample lies next to the submission; stdout must equal its bytes);
// Timeout — the row's own ceiling in seconds (0 — the mechanism
// default); Launch/Tree — the procedural modes of the runner (any
// row; see the constants below); Witness/Repeat — the two NATIVE
// typed heads (the staged-bytes probe and the repeat; see
// UnmarshalYAML). Outcome/Reason/EnvFail — the
// machine's record of the last run; RanAt/Tree/Inputs — the machine's
// memo of that run (the content-addressed freshness key: the row
// keeps its green while its definition and the digests of the files
// its command names — plus its from-input seeds, its sample and the
// declared build — are unchanged); RanRc/RanOut — the observed exit
// code and the digest of the observed stdout of that run (the report's
// measured answer). The memo fields are machine-recorded only: a
// submission carrying them is refused like any attempt to write the
// machine's own records.
type Row struct {
	Run     Cmd         `yaml:"run"`
	Seeds   []Seed      `yaml:"seeds,omitempty"`
	Out     *TextExpect `yaml:"out,omitempty"`
	Err     *TextExpect `yaml:"err,omitempty"`
	Rc      *int        `yaml:"rc,omitempty"`
	State   []StateWant `yaml:"state,omitempty"`
	Bytes   string      `yaml:"bytes,omitempty"`
	Timeout int         `yaml:"timeout,omitempty"`
	// Launch — the procedural launch mode of the runner: "" (none) or
	// "nonet" (the launch happens inside a fresh network-less
	// namespace — a launch discipline, nothing is wrapped around the
	// command). A mode of the mechanics: any meaning the mode carries
	// lives in the probe or the row's own expectations, never here.
	Launch string `yaml:"launch,omitempty"`
	// Tree — the procedural tree mode of the runner: "" (none) or
	// "unchanged" (the row-scoped tree identity control: outside
	// .punchtape/, build-ignore and the row's own seeds, nothing the
	// row leaves behind survives it).
	Tree string `yaml:"tree,omitempty"`
	// Form — the one-line source of the row's formal choices: the
	// brief/prereg phrase an exact pin quotes. A row
	// that pins formal surface — a nonzero rc, an exact (non-contains)
	// out/err/state text — without a citation is a FORM DECISION
	// UNRESOLVED: the door closes runs until the hand cites the
	// source or asks the operator. The machine never judges the
	// citation's content — presence is the whole mechanism.
	Form string `yaml:"form,omitempty"`
	// The typed heads. Refuse and Det are DECODE-ONLY notations: the
	// grammar accepts them (strictly checked), then maps each onto the
	// general fields it names and nils itself — a stored row never
	// carries them, the typed spelling and the long form are one row.
	// Witness and Repeat are the native heads; they live in the row.
	Refuse  *RefuseRow  `yaml:"refuse,omitempty"`
	Det     *DetRow     `yaml:"det,omitempty"`
	Witness *WitnessRow `yaml:"witness,omitempty"`
	Repeat  *RepeatRow  `yaml:"repeat-run,omitempty"`
	// Outcome/Reason/EnvFail — the machine's record of the last run;
	// RanAt/RanTree/Inputs — the machine's memo of that run (the
	// content-addressed freshness key: the row keeps its green while
	// its definition and the digests of the files its command names —
	// plus its from-input seeds, its sample and the declared build —
	// are unchanged); RanRc/RanOut — the observed exit code and the
	// digest of the observed stdout of that run (the report's measured
	// answer). The memo fields are machine-recorded only: a submission
	// carrying them is refused like any attempt to write the machine's
	// own records.
	Outcome string `yaml:"outcome,omitempty"`
	Reason  string `yaml:"reason,omitempty"`
	EnvFail bool   `yaml:"env-fail,omitempty"`
	RanAt   string `yaml:"ran-at,omitempty"`
	RanTree string `yaml:"ran-tree,omitempty"`
	Inputs  string `yaml:"inputs,omitempty"`
	RanRc   *int   `yaml:"ran-rc,omitempty"`
	RanOut  string `yaml:"ran-out,omitempty"`
}

// RefuseRow — the decode-only refusal head: refuse{args, says, rc}.
type RefuseRow struct {
	Args Cmd    `yaml:"args"`
	Says string `yaml:"says"`
	Rc   int    `yaml:"rc"`
}

// DetRow — the decode-only determinism head: det{run, same}.
type DetRow struct {
	Run  Cmd    `yaml:"run"`
	Same string `yaml:"same"`
}

// RepeatRow — the native repeat witness: the command the row runs
// twice. A wrapper (not Cmd itself) so the row's field is a plain
// struct pointer; the body rides the command's own two forms (a
// string or an argv list) through the same marshal contract.
type RepeatRow struct {
	Cmd Cmd `yaml:",inline"`
}

// UnmarshalYAML — the command's forms: a scalar or an argv list.
func (rr *RepeatRow) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return unmarshal(&rr.Cmd)
}

// MarshalYAML — the command's own canonical form (a nil head is the
// row's null spelling, like every other absent field).
func (rr *RepeatRow) MarshalYAML() (interface{}, error) {
	if rr == nil {
		return nil, nil
	}
	return rr.Cmd.MarshalYAML()
}

// The row's procedural modes (the runner's own disciplines, meaning-
// free): the launch mode and the tree mode. Any row carries them; the
// meaning a mode serves lives in the row's own expectations or its
// probe, never in the machine.
const (
	// LaunchModeNonet — the launch happens inside a fresh network-less
	// namespace: the command's own bytes are untouched, the network
	// simply is not there; a host that cannot raise the namespace
	// answers an honest environment red.
	LaunchModeNonet = "nonet"
	// TreeModeUnchanged — the row-scoped tree identity control:
	// outside .punchtape/, build-ignore and the row's own seeds,
	// nothing the row leaves behind survives it.
	TreeModeUnchanged = "unchanged"
)

// WitnessRow — the staged-bytes probe head: the probe's bytes (the
// checker program) ride THE DOOR — a file next to the submission in
// the machine's inbox, named by the file:// form. The machine stages
// the bytes into the tree for the row's duration, runs them under the
// row's procedural modes and compares rc/stdout with the row's want.
// What the probe checks is not the machine's knowledge: the machine
// reads only the universal observables.
type WitnessRow struct {
	Probe string `yaml:"probe"`
}

// The row heads: the general form and the typed kinds. The typed
// heads name the mechanical families the long form wrote by hand —
// each with a strictly checked field set.
const (
	headRun       = "run"
	headRefuse    = "refuse"
	headDet       = "det"
	headWitness   = "witness"
	headRepeatRun = "repeat-run"
)

// rowHeads — the head fields in their printed order.
var rowHeads = []string{headRun, headRefuse, headDet, headWitness, headRepeatRun}

// headList — the head names as one printed list.
func headList() string { return strings.Join(rowHeads, ", ") }

// rowGeneralFields — the fields of the general form, beyond the head.
var rowGeneralFields = []string{
	"seeds", "out", "err", "rc", "state", "bytes", "timeout", "launch", "tree", "form",
	"outcome", "reason", "env-fail", "ran-at", "ran-tree", "inputs",
	"ran-rc", "ran-out",
}

// typedFields — the row-level fields a typed head composes with:
// seeds and timeout on every head, the procedural modes (launch, tree)
// on every head, out where the head leaves the success stream open
// (the native heads), and the machine's
// recorded fields (a stored row carries them back; a document that
// forges them lands in the row for the door's forgery guard, never
// drops silently).
func typedFields(head string) map[string]bool {
	m := map[string]bool{"seeds": true, "timeout": true, "launch": true, "tree": true, "form": true}
	if head == headWitness || head == headRepeatRun {
		m["out"] = true
	}
	if head == headWitness {
		m["rc"] = true
	}
	for _, f := range rowGeneralFields {
		switch f {
		case "seeds", "out", "err", "rc", "state", "bytes", "timeout", "launch", "tree", "form":
		default:
			m[f] = true
		}
	}
	return m
}

// headOwnFields — the head's own field names, for the strict probe of
// the head's own body.
var headOwnFields = map[string][]string{
	headRefuse:    {"args", "says", "rc"},
	headDet:       {"run", "same"},
	headWitness:   {"probe"},
	headRepeatRun: {},
}

// headRequiredFields — the head's own fields whose ABSENCE is a form
// error (the rest may ride as their zero value only when legal).
var headRequiredFields = map[string][]string{
	headRefuse:  {"args", "says", "rc"},
	headDet:     {"run", "same"},
	headWitness: {"probe"},
}

// headFormKey — the canon text key of each head's form error.
var headFormKey = map[string]string{
	headRefuse:    "canonvalid.error.typed-refuse-form",
	headDet:       "canonvalid.error.typed-det-form",
	headWitness:   "canonvalid.error.typed-witness-form",
	headRepeatRun: "canonvalid.error.typed-repeat-form",
}

// UnmarshalYAML — the row grammar: the general form (a run command
// plus expectation fields) or ONE typed head. refuse{args, says, rc}
// and det{run, same} decode INTO the general fields they name — the
// typed spelling and the long form are one row, one canon (the same
// law as the command's string/argv forms); there is no expansion
// step because nothing is ever expanded — the row IS what the head
// spells. witness and repeat-run are native heads: the staged-bytes
// probe and the repeat run through the machine's own launch
// disciplines, so the head survives into the stored row. A row
// carrying no head, or more than one, is refused; a field that does
// not belong to the carried head — on the row or inside the head's
// own body — is refused naming it. The procedural modes (launch,
// tree) compose with every head and carry exactly their two legal
// values.
func (r *Row) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var probe map[string]interface{}
	if err := unmarshal(&probe); err != nil {
		return err
	}
	var heads []string
	for _, h := range rowHeads[1:] {
		v, ok := probe[h]
		// a typed head is present when it carries a value (the stored
		// records spell the absent heads null); an empty-string repeat
		// is a present-but-empty command — its own form error
		if ok && v != nil {
			heads = append(heads, h)
		}
	}
	// the general head is the row's run key whenever no typed head
	// claims the row — an empty or null run is the general form's own
	// honest error at the door, exactly as before the typed heads
	if len(heads) == 0 {
		if _, ok := probe[headRun]; ok {
			heads = []string{headRun}
		}
	}
	if len(heads) != 1 {
		found := strings.Join(heads, ", ")
		if found == "" {
			found = "none"
		}
		return fmt.Errorf("%s", canondata.T("canonvalid.error.row-head", canondata.M{
			"heads": headList(), "found": found,
		}))
	}
	head := heads[0]
	// a null- or empty-spelled key is an absent field (the stored
	// records spell every field; the general form's expectations spell
	// null where they carry none)
	live := map[string]interface{}{}
	for k, v := range probe {
		if v != nil && v != "" {
			live[k] = v
		}
	}
	if head == headRun {
		allowed := map[string]bool{headRun: true}
		for _, f := range rowGeneralFields {
			allowed[f] = true
		}
		for name := range live {
			if !allowed[name] {
				return fmt.Errorf("%s", canondata.T("canonvalid.error.row-field", canondata.M{
					"field": name, "fields": headRun,
				}))
			}
		}
	} else {
		// the head's own body is strictly its own fields
		if own, ok := probe[head].(map[string]interface{}); ok {
			legal := map[string]bool{}
			for _, f := range headOwnFields[head] {
				legal[f] = true
			}
			for name := range own {
				if !legal[name] {
					return fmt.Errorf("%s", canondata.T("canonvalid.error.typed-field", canondata.M{
						"kind": head, "field": name,
						"fields": strings.Join(headOwnFields[head], ", "),
					}))
				}
			}
		} else if len(headOwnFields[head]) > 0 {
			return fmt.Errorf("%s", canondata.T("canonvalid.error.typed-map-form", canondata.M{
				"kind": head, "fields": strings.Join(headOwnFields[head], ", "),
			}))
		}
		// the head's REQUIRED own fields are present in the body (a
		// missing field refuses by name; the decode alone cannot tell
		// absence from the zero value)
		if own, ok := probe[head].(map[string]interface{}); ok {
			for _, f := range headRequiredFields[head] {
				if _, present := own[f]; !present {
					return fmt.Errorf("%s", canondata.T(headFormKey[head]))
				}
			}
		}
		// the row-level fields the head composes with
		allowed := map[string]bool{head: true}
		for f := range typedFields(head) {
			allowed[f] = true
		}
		for name := range live {
			if !allowed[name] {
				return fmt.Errorf("%s", canondata.T("canonvalid.error.row-typed-field", canondata.M{
					"kind": head, "field": name,
				}))
			}
		}
	}
	// one plain decode fills every legal field at once (the strict
	// decoder knows them all: the heads ride as tagged fields)
	type plain Row
	var p plain
	if err := unmarshal(&p); err != nil {
		return err
	}
	switch head {
	case headRun:
		// the general form: the plain decode already filled the row
	case headRefuse:
		f := p.Refuse
		if f == nil || cmdEmpty(f.Args) || f.Says == "" || f.Rc < 0 {
			return fmt.Errorf("%s", canondata.T("canonvalid.error.typed-refuse-form"))
		}
		rc := f.Rc
		p.Run, p.Out, p.Err, p.Rc = f.Args, &TextExpect{Text: ""}, &TextExpect{Text: f.Says}, &rc
		p.Refuse = nil
	case headDet:
		f := p.Det
		if f == nil || cmdEmpty(f.Run) || f.Same == "" {
			return fmt.Errorf("%s", canondata.T("canonvalid.error.typed-det-form"))
		}
		p.Run, p.Out = f.Run, &TextExpect{Text: f.Same}
		p.Det = nil
	case headWitness:
		f := p.Witness
		if f == nil || f.Probe == "" {
			return fmt.Errorf("%s", canondata.T("canonvalid.error.typed-witness-form"))
		}
	default: // headRepeatRun
		if p.Repeat == nil || cmdEmpty(p.Repeat.Cmd) {
			return fmt.Errorf("%s", canondata.T("canonvalid.error.typed-repeat-form"))
		}
	}
	// The procedural modes' values are the grammar's own: one legal
	// spelling each, everything else is refused naming it.
	switch p.Launch {
	case "", LaunchModeNonet:
	default:
		return fmt.Errorf("%s", canondata.T("canonvalid.error.launch-mode", canondata.M{
			"value": p.Launch, "mode": LaunchModeNonet,
		}))
	}
	switch p.Tree {
	case "", TreeModeUnchanged:
	default:
		return fmt.Errorf("%s", canondata.T("canonvalid.error.tree-mode", canondata.M{
			"value": p.Tree, "mode": TreeModeUnchanged,
		}))
	}
	*r = Row(p)
	return nil
}

// cmdEmpty — whether a command form carries nothing.
func cmdEmpty(c Cmd) bool {
	if c.IsArgv {
		return len(c.Argv) == 0
	}
	return c.Line == ""
}

// CommandFields — the row's command argv whichever head carries it
// (the general run, the staged probe or the repeat). The freshness
// key, the dependency walk and the renders all ask through here: a
// row has exactly one command.
func (r Row) CommandFields() ([]string, error) {
	switch {
	case r.Witness != nil:
		return []string{"./" + r.ProbeRef()}, nil
	case r.Repeat != nil:
		return r.Repeat.Cmd.Fields()
	default:
		return r.Run.Fields()
	}
}

// CmdForm — the row's command one-line face for renders (whichever
// head carries it).
func (r Row) CmdForm() string {
	switch {
	case r.Witness != nil:
		return "./" + r.ProbeRef()
	case r.Repeat != nil:
		return r.Repeat.Cmd.Form()
	default:
		return r.Run.Form()
	}
}

// ProbeRef — the probe name the row's witness reference carries (the
// file:// form decoded); empty — the row stages no probe.
func (r Row) ProbeRef() string {
	const scheme = "file://"
	if r.Witness != nil && strings.HasPrefix(r.Witness.Probe, scheme) {
		return strings.TrimPrefix(r.Witness.Probe, scheme)
	}
	return ""
}

// FormPins — whether the row pins formal surface the brief may not
// have named: a nonzero expected exit code, or an exact (non-contains)
// stream/state text — a report-line format, a diagnostic's wording.
// The decision is the HAND's to present: a source line (Form) or an
// operator question; the machine only closes the door on silence.
func (r Row) FormPins() bool {
	if r.Rc != nil && *r.Rc != 0 {
		return true
	}
	if exactPin(r.Out) || exactPin(r.Err) {
		return true
	}
	for _, st := range r.State {
		if exactPin(st.Text) {
			return true
		}
	}
	return false
}

// exactPin — an exact (non-fragment) text demand: pinning the exact
// bytes of a stream or file is a format decision, whatever the text.
func exactPin(t *TextExpect) bool {
	return t != nil && !t.Contains
}

// FormUnresolved — the row pins formal surface and carries no source
// line: a FORM DECISION UNRESOLVED.
func (r Row) FormUnresolved() bool {
	return r.FormPins() && r.Form == ""
}

// RowDefDigest — the digest of the row's DEFINITION: the canonical
// bytes of the row without the machine-recorded fields (outcome,
// reason, env-fail and the memo). Two spellings of one definition
// are one definition; a changed expectation or command is a new one.
func (r Row) RowDefDigest() string {
	r.Outcome, r.Reason = "", ""
	r.EnvFail = false
	r.RanAt, r.RanTree, r.Inputs = "", "", ""
	r.RanRc, r.RanOut = nil, ""
	// Row is plain data: marshal cannot fail
	data, _ := yamlio.Marshal(r)
	return yamlio.Digest(data)
}

// Memoized — whether the row carries a recorded run memo.
func (r Row) Memoized() bool {
	return r.RanAt != "" || r.RanTree != "" || r.Inputs != "" ||
		r.RanRc != nil || r.RanOut != ""
}

// BytesRef — the sample name the row's bytes reference carries (the
// file:// form decoded); empty — the row has no sample.
func (r Row) BytesRef() string {
	const scheme = "file://"
	if strings.HasPrefix(r.Bytes, scheme) {
		return strings.TrimPrefix(r.Bytes, scheme)
	}
	return ""
}

// Atom — one behavior of the corpus: the executor's short id,
// a living name, the mandatory prose and the rows. The atom is green
// when all its rows are green; the verdict needs every atom green,
// the prose on every atom and the ratification.
type Atom struct {
	ID    string `yaml:"id"`
	Name  string `yaml:"name"`
	Prose string `yaml:"prose"`
	Rows  []Row  `yaml:"rows"`
	// Digest — the record's own digest: computed by the
	// machine at every write over the record without the digest
	// itself; the atom's identity as stored, for renders and audits.
	Digest string `yaml:"digest,omitempty"`
}

// DigestOf — the atom record's own digest: over the canonical bytes
// of the record without the digest field itself.
func (a Atom) DigestOf() string {
	a.Digest = ""
	// Atom is plain data: marshal cannot fail (a map key with control
	// bytes cannot occur — ids and paths are form-checked)
	data, _ := yamlio.Marshal(a)
	return yamlio.Digest(data)
}

// Green — all the atom's rows ran green (an atom without rows is not
// green).
func (a Atom) Green() bool {
	if len(a.Rows) == 0 {
		return false
	}
	for _, r := range a.Rows {
		if r.Outcome != RowGreen {
			return false
		}
	}
	return true
}

// RowLabel — the row's address in replies and renders: atom id plus
// the one-based index inside the atom.
func RowLabel(atomID string, index int) string {
	return fmt.Sprintf("%s/%d", atomID, index)
}

// scalarText — any YAML scalar as the machine's text: a string as
// is, a number or a boolean in their shortest canonical spelling,
// null the empty string; anything else is not a scalar.
func scalarText(v interface{}) (string, error) {
	switch t := v.(type) {
	case nil:
		return "", nil
	case string:
		return t, nil
	case bool:
		return strconv.FormatBool(t), nil
	// goccy decodes document integers as int64/uint64, never int
	case int64:
		return strconv.FormatInt(t, 10), nil
	case uint64:
		return strconv.FormatUint(t, 10), nil
	case float64:
		if t == math.Trunc(t) && math.Abs(t) < 1e15 {
			return strconv.FormatInt(int64(t), 10), nil
		}
		return strconv.FormatFloat(t, 'g', -1, 64), nil
	}
	return "", fmt.Errorf("not a scalar")
}

// SchemaVersion — the canon schema version. Changes when the formats
// change; migrations go from version to version, none may be skipped.
// Version 7: the legacy shape (atoms with triggers, contracts,
// catalogs, checks, the inbox set protocol).
// Version 8: the behavior shape — the corpus is behaviors
// (id/name/prose/rows, the rows inside the atom), the submission is
// one document of two kinds (spec|run), the passport, the traces,
// the amend holds, the contracts, the catalogs and the question
// batches died. A schema-7 instance honestly refuses to open (strict
// decode): the schema break leaves no legacy instances.
// Version 9: the current shape — the row grammar is wholly
// procedural: the witness head stages its probe bytes through the
// door (file://NAME in the inbox), the launch/tree modes (nonet,
// unchanged) compose with every head, the machine-recorded memo field
// "tree" is renamed "ran-tree" (the mode owns the bare name). A
// schema-8 instance honestly refuses to open: the closed series roots
// are experiment data, never re-read.
const SchemaVersion = 9

// State — the instance's state. Always derivable from the journal;
// the file is a cache — on a divergence the journal wins.
type State struct {
	SchemaVersion int    `yaml:"schema-version"`
	Stage         string `yaml:"stage"`
	// Wish — the brief text, fixed once at the instance's birth (the
	// phrase of the first next).
	Wish *string `yaml:"wish,omitempty"`
	// ContractSha — the sha256 of the AGENTS.md bytes as the machine
	// wrote them at the instance's birth. The render's one-line
	// contract reference stands on it; a tree whose AGENTS.md is gone
	// or changed gets the FULL contract rendered again (anti-tamper,
	// not an option).
	ContractSha string `yaml:"contract-sha,omitempty"`
	// Language — the brief's language from the deterministic script
	// detector (en/ru/zh); it steers the human blocks (the summary),
	// while the machine protocol stays English always.
	Language string `yaml:"language,omitempty"`
	// Build — the spec's build command (argv); nil — the document
	// never declared one, [] — declared and cleared. Runs before the
	// rows of every run submission.
	Build []string `yaml:"build,omitempty"`
	// BuildIgnore — the spec's ignore globs: paths of the work tree
	// outside the accounted surface (build outputs, caches). The
	// .punchtape zone is excluded unconditionally.
	BuildIgnore []string       `yaml:"build-ignore,omitempty"`
	Touched     []string       `yaml:"touched,omitempty"`
	Counters    map[string]int `yaml:"counters"`
}
