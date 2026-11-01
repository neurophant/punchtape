// Package doc — the document: the machine's one input format. A
// document is an intent document: it carries what the human wants and
// the agent's own words about it — never a machine-computed fact. A
// submission lists the behaviors it adds or replaces by id, the ids
// named in remove are retired, the registry it lists merges onto the
// standing one by each part's own key. The model is pure data; the
// laws of its shape live in validate.go, its bytes in codec.go.
package doc

// Note — one author's note: free text the machine never interprets,
// carried with its kind and references; every reference must name a
// real thing or the note is refused.
type Note struct {
	Kind string   `yaml:"kind"`
	Text string   `yaml:"text"`
	Refs []string `yaml:"refs,omitempty"`
}

// Requirement — one demand made specific: the demand lines it
// occupies and the unwinding in human words. The id is the
// requirement's stable name; refines names the parent when the
// requirement is a decomposition of one. The lines' own text the
// machine reads from the stream itself — the author never transcribes
// it.
type Requirement struct {
	ID      string   `yaml:"id"`
	Lines   []string `yaml:"lines,omitempty"`
	Prose   string   `yaml:"prose"`
	Refines string   `yaml:"refines,omitempty"`
}

// Marker — one demand line declared to be no requirement, beside the
// reason a reader can weigh. The marked lines' own text the machine
// reads from the stream itself.
type Marker struct {
	Lines  []string `yaml:"lines,omitempty"`
	Reason string   `yaml:"reason"`
}

// Exclusion — one loud exclusion of a line range from coverage, with
// the reason a reader can weigh and the references that tie the
// reason to a demand line or a law. A range that names no uncovered
// line is refused — a mark that marks nothing is noise.
type Exclusion struct {
	Range  string   `yaml:"range"`
	Reason string   `yaml:"reason"`
	Refs   []string `yaml:"refs"`
}

// OpenOrder — one question standing before the human: the demand
// line it rides on, the question itself and its state. An awaiting
// order holds the verdict; a deferred order stays visible forever
// with its reason. OpenedAtLines — the stream's size at the order's
// birth, written by the machine, never by the author.
type OpenOrder struct {
	Demand        string `yaml:"demand"`
	Question      string `yaml:"question"`
	State         string `yaml:"state"`
	Reason        string `yaml:"reason,omitempty"`
	OpenedAtLines int    `yaml:"opened-at-lines,omitempty"`
}

// Intent — the declared outcome of demand lines: what the line asks
// for, in the agent's own words, with the outcome's own carrier
// inside — answer carries the answering text, execute the command the
// machine runs. The other four outcomes carry nothing: their carriers
// are the registry the same document lists.
type Intent struct {
	Lines   []string `yaml:"lines"`
	Outcome string   `yaml:"outcome"`
	Prose   string   `yaml:"prose"`
	Text    string   `yaml:"text,omitempty"`
	Run     Cmd      `yaml:"run,omitempty"`
}

// The six outcomes — the protocol's own vocabulary.
const (
	OutcomeAnswer   = "answer"
	OutcomeExecute  = "execute"
	OutcomeDescribe = "describe"
	OutcomeWork     = "work"
	OutcomeMark     = "mark"
	OutcomeAsk      = "ask"
)

// MandateRef — one permission: the demand line that ordered the
// change and the standing thing it permits to replace (a behavior id
// or a requirement id).
type MandateRef struct {
	Line   string `yaml:"line"`
	Target string `yaml:"target"`
}

// Registry — the corpus's account of the demands: the requirements
// with their line spans, the not-a-requirement markers, the loud
// line exclusions and the open orders.
type Registry struct {
	Requirements map[string]Requirement
	Order        []string // the requirements' first-seen order
	Markers      []Marker
	Exclusions   []Exclusion
	Open         []OpenOrder
}

// Document — one submission's whole content: the agent's intent, not
// the machine's facts. The human's words enter the stream file
// demands.md; the machine takes them itself at the submission that
// meets them. Build absent (nil) leaves the standing build untouched
// — the merge law: what a submission does not list, it does not
// change; every registry list merges onto the standing one by its own
// key. A done document is exactly done: true — the completion act.
type Document struct {
	Done         bool          `yaml:"done,omitempty"`
	Intents      []Intent      `yaml:"intents,omitempty"`
	Work         string        `yaml:"work,omitempty"`
	Build        *string       `yaml:"build,omitempty"`
	Remove       []string      `yaml:"remove,omitempty"`
	Requirements []Requirement `yaml:"requirements,omitempty"`
	Markers      []Marker      `yaml:"not-a-requirement,omitempty"`
	Exclusions   []Exclusion   `yaml:"uncovered-lines,omitempty"`
	Open         []OpenOrder   `yaml:"open,omitempty"`
	Behaviors    []Behavior    `yaml:"behaviors,omitempty"`
	Notes        []Note        `yaml:"notes,omitempty"`
	Mandate      []MandateRef  `yaml:"mandate,omitempty"`
}

// BuildCmd — the declared build command, "" when none stands.
func (d *Document) BuildCmd() string {
	if d.Build == nil {
		return ""
	}
	return *d.Build
}

// Behavior — one named unit of visible behavior: the author's prose,
// the exact line ranges of code it stands on (the claims — its only
// coverage and its proof's dependency), the requirements it covers
// and the rows that pin it. The id is a stable path; a submission
// that repeats an id replaces the behavior's head and merges its
// rows by run.
type Behavior struct {
	ID     string   `yaml:"id"`
	Name   string   `yaml:"name"`
	Prose  string   `yaml:"prose"`
	Claims []string `yaml:"claims,omitempty"`
	Covers []string `yaml:"covers,omitempty"`
	Rows   []Row    `yaml:"rows"`
}

// Row — one executable check of the behavior: a single command and
// what it must leave behind. The run is the row's merge key: the
// same run replaces the whole row, a new run is appended, unlisted
// rows stand.
type Row struct {
	Run    Cmd    `yaml:"run"`
	Expect Expect `yaml:"expect,omitempty"`
}

// Cmd — one command: either a shell line (a string; the machine
// runs it through the shell) or an argv list (runs as is). The two
// spellings are one command; normalization keeps the author's
// spelling and canonical bytes carry it unchanged.
type Cmd struct {
	Line string
	Argv []string
}

// IsZero — the empty command, so an intent's run field omits itself
// when it carries nothing.
func (c Cmd) IsZero() bool { return c.Line == "" && c.Argv == nil }

// Expect — what a finished row must leave behind. Every field is
// optional; an absent field is not checked. Out matches the exact
// stdout or, prefixed "contains: ", a substring of it. Files maps a
// root-relative path to the exact file bytes or a "contains: "
// substring of them. RC absent means the exit must be zero; present
// — exactly the named code. The pinned bytes are the row's own
// output — the machine's constants never ride them.
type Expect struct {
	RC    *int              `yaml:"rc,omitempty"`
	Out   string            `yaml:"out,omitempty"`
	Files map[string]string `yaml:"files,omitempty"`
}
