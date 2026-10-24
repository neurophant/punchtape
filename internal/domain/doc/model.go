// Package doc — the document: the machine's one input format. A
// document is a whole table or a subset of one: the behaviors it
// lists are added or replaced by id, the ids named in remove are
// retired, the registry it lists merges onto the standing one. The
// model is pure data; the laws of its shape live in validate.go,
// its bytes in codec.go.
package doc

// Note — one author's note: free text the machine never interprets.
type Note struct {
	Text string `yaml:"text"`
}

// Requirement — one demand the wish makes: an id and the brief lines
// it occupies. The id is the requirement's stable name; the brief
// ranges name wish lines in the W grammar.
type Requirement struct {
	ID    string   `yaml:"id"`
	Brief []string `yaml:"brief,omitempty"`
}

// Marker — one wish line declared to be no requirement: the line's
// own words quoted verbatim beside the reason a reader can weigh.
// Dismissing a demand demands copying its exact text into the
// signed corpus — the mark cannot be quiet.
type Marker struct {
	Lines  []string `yaml:"lines,omitempty"`
	Quote  string   `yaml:"quote"`
	Reason string   `yaml:"reason"`
}

// Registry — the corpus's account of the wish: the requirements
// with their brief-line spans, the not-a-requirement markers and the
// declared artifacts.
type Registry struct {
	Requirements map[string]Requirement
	Order        []string // the requirements' first-seen order
	Markers      []Marker
	Artifacts    []string
}

// Document — one submission's whole content. Build absent (nil)
// leaves the standing build untouched — the merge law: what a
// submission does not list, it does not change. The registry lists
// follow their own merge laws (merge.go).
type Document struct {
	Notes           []Note        `yaml:"notes,omitempty"`
	Build           *string       `yaml:"build,omitempty"`
	Remove          []string      `yaml:"remove,omitempty"`
	Requirements    []Requirement `yaml:"requirements,omitempty"`
	NotRequirements []Marker      `yaml:"not-a-requirement,omitempty"`
	Artifacts       []string      `yaml:"artifacts,omitempty"`
	Behaviors       []Behavior    `yaml:"behaviors,omitempty"`
}

// BuildCmd — the declared build command, "" when none stands.
func (d *Document) BuildCmd() string {
	if d.Build == nil {
		return ""
	}
	return *d.Build
}

// Behavior — one named unit of visible behavior: the author's prose,
// the files it stands on (the scope map), the requirements it covers
// and the rows that pin it. The id is a stable path; a submission
// that repeats an id replaces the behavior's head and merges its
// rows by run.
type Behavior struct {
	ID     string   `yaml:"id"`
	Name   string   `yaml:"name"`
	Prose  string   `yaml:"prose"`
	Scope  []string `yaml:"scope,omitempty"`
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

// Expect — what a finished row must leave behind. Every field is
// optional; an absent field is not checked. Out matches the exact
// stdout or, prefixed "contains: ", a substring of it. Files maps a
// root-relative path to the exact file bytes or a "contains: "
// substring of them. RC absent means the exit must be zero; present
// — exactly the named code.
type Expect struct {
	RC    *int              `yaml:"rc,omitempty"`
	Out   string            `yaml:"out,omitempty"`
	Files map[string]string `yaml:"files,omitempty"`
}
