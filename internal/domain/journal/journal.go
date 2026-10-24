// Package journal — the zone's one source of truth: the append-only
// records and the fold that rebuilds the present from them. Every
// other zone file is a generated image.
// Records carry no wall clock: order is the record's place, identity
// is the record's bytes — the journal of one history is
// byte-identical on every machine.
package journal

import (
	"github.com/neurophant/punchtape/internal/domain/doc"
)

// The record kinds.
const (
	KindBirth      = "birth"
	KindSubmission = "submission"
	KindRun        = "run"
	KindVerdict    = "verdict"
	KindRefusal    = "refusal"
	KindCompletion = "completion"
)

// Record — one journaled fact. Exactly one payload field matches
// the kind.
type Record struct {
	Kind       string        `yaml:"kind"`
	Birth      *Birth        `yaml:"birth,omitempty"`
	Submission *Submission   `yaml:"submission,omitempty"`
	Run        *Run          `yaml:"run,omitempty"`
	Verdict    *Verdict      `yaml:"verdict,omitempty"`
	Refusals   []doc.Refusal `yaml:"refusals,omitempty"`
	Completed  *Completion   `yaml:"completed,omitempty"`
}

// Completion — the completion act as a journaled fact of its own:
// the run whose standing corpus the door received back, closing the
// verdict. Submissions and completions are different economies —
// one counts documents the door took, the other counts the acts
// that finished the work — and neither may hide the other.
type Completion struct {
	Run int
}

// Verdict — one recorded point of truth: the corpus version it
// certifies, the signature over the corpus bytes and every
// behavior's scope fingerprint, the digest over the registry's
// canonical bytes (the requirements and the markers — the wish's
// whole account attested with the same seal), and the run that
// closed it. Verdicts are append-only: a new one never rewrites an
// old one.
type Verdict struct {
	N              int            `yaml:"n"`
	CorpusVersion  int            `yaml:"corpus-version"`
	Signature      string         `yaml:"signature"`
	RegistryDigest string         `yaml:"registry-digest"`
	Scopes         []VerdictScope `yaml:"scopes"`
	Run            int            `yaml:"run"`
}

// VerdictScope — one certified behavior's scope fingerprint.
type VerdictScope struct {
	ID      string `yaml:"id"`
	Pointer string `yaml:"pointer"`
}

// Birth — the instance's goal, fixed once: the wish's bytes as a
// content-addressed object and their digest, and the machine version
// that created the instance. The goal is immutable forever; the
// recorded version is the instance's own origin fact.
type Birth struct {
	Addr    string `yaml:"addr"`
	Digest  string `yaml:"digest"`
	Version string `yaml:"version,omitempty"`
}

// BehaviorOp — one behavior of a submission: the id and the address
// of the behavior's canonical bytes in the object store.
type BehaviorOp struct {
	ID   string `yaml:"id"`
	Addr string `yaml:"addr"`
}

// Submission — one accepted document: the corpus version it
// produced, the build command it declares, the registry parts it
// listed (requirements, markers, artifacts — merged by their own
// laws), the behaviors it added or merged by id (each op's object is
// the behavior as it stands after the merge), the ids it retired,
// and the address of its own canonical bytes (the notes ride the
// journal through it).
type Submission struct {
	CorpusVersion   int               `yaml:"corpus-version"`
	Build           *string           `yaml:"build,omitempty"`
	Requirements    []doc.Requirement `yaml:"requirements,omitempty"`
	NotRequirements []doc.Marker      `yaml:"not-a-requirement,omitempty"`
	Artifacts       []string          `yaml:"artifacts,omitempty"`
	Added           []BehaviorOp      `yaml:"added,omitempty"`
	Removed         []string          `yaml:"removed,omitempty"`
	DocAddr         string            `yaml:"doc-addr"`
}

// RowFact — one executed row's measured fact.
type RowFact struct {
	Key    string   `yaml:"key"`    // "<behavior id>/<row index>"
	Row    string   `yaml:"row"`    // the row's canonical bytes' digest
	Status string   `yaml:"status"` // green | red
	Exit   int      `yaml:"exit"`
	Facts  []string `yaml:"facts,omitempty"` // the broken facts, red only
	Proof  string   `yaml:"proof"`           // the proof fingerprint the fact stands on
}

// Run — one machine run that changed facts: the number, the build's
// fact (a build ran), the executed rows' facts and the lessons the
// run's red→green transitions fixed.
type Run struct {
	N       int       `yaml:"n"`
	Build   *Build    `yaml:"build,omitempty"`
	Rows    []RowFact `yaml:"rows,omitempty"`
	Lessons []Lesson  `yaml:"lessons,omitempty"`
}

// Lesson — one fixed red, fixed as data: the reason class, the row,
// the broken facts, the row's bytes before and after the fix, the two
// run numbers. Zero tokens of any model — the machine measured it
// itself.
type Lesson struct {
	Reason   string   `yaml:"reason"`
	Key      string   `yaml:"key"`
	Facts    []string `yaml:"facts"`
	RedRow   string   `yaml:"red-row"`
	GreenRow string   `yaml:"green-row"`
	RedRun   int      `yaml:"red-run"`
	GreenRun int      `yaml:"green-run"`
}

// Build — the build's fact of one run: the exit code, whether the
// deadline cut it, and the launch failure's own words when the
// command never started. A build that failed is a red fact of the
// run; the rows do not execute behind it.
type Build struct {
	Exit     int    `yaml:"exit"`
	TimedOut bool   `yaml:"timed-out,omitempty"`
	Reason   string `yaml:"reason,omitempty"`
}

// Status spellings of a row fact.
const (
	StatusGreen = "green"
	StatusRed   = "red"
)

// Standing — the present rebuilt from the records: the fold's whole
// answer. The behaviors carry their content resolved from the
// object store; the registry carries the wish's account; the rows
// carry their standing fact.
type Standing struct {
	Born          bool
	GoalAddr      string
	GoalDigest    string
	BirthVersion  string
	CorpusVersion int
	Build         string
	Registry      doc.Registry
	Behaviors     map[string]doc.Behavior
	Order         []string // the behaviors' first-seen order
	Rows          map[string]RowFact
	RunNumber     int
	Verdict       *Verdict
	Submissions   []Submission
	Refused       int
	Completions   int
	Lessons       []Lesson
}

// Fold — the records into the standing. readObj answers a behavior
// object's bytes by address; a missing object is an honest error —
// the journal names what must exist.
func Fold(records []Record, readObj func(addr string) ([]byte, error)) (*Standing, error) {
	st := &Standing{
		Behaviors: map[string]doc.Behavior{},
		Rows:      map[string]RowFact{},
	}
	for _, rec := range records {
		switch {
		case rec.Kind == KindBirth && rec.Birth != nil:
			st.Born = true
			st.GoalAddr = rec.Birth.Addr
			st.GoalDigest = rec.Birth.Digest
			st.BirthVersion = rec.Birth.Version
		case rec.Kind == KindSubmission && rec.Submission != nil:
			s := rec.Submission
			st.CorpusVersion = s.CorpusVersion
			if s.Build != nil {
				st.Build = *s.Build
			}
			st.Submissions = append(st.Submissions, *s)
			st.Registry = doc.MergeRegistry(st.Registry, doc.Document{
				Remove:          s.Removed,
				Requirements:    s.Requirements,
				NotRequirements: s.NotRequirements,
				Artifacts:       s.Artifacts,
			})
			for _, rem := range s.Removed {
				delete(st.Behaviors, rem)
			}
			for _, op := range s.Added {
				if _, seen := st.Behaviors[op.ID]; !seen {
					st.Order = append(st.Order, op.ID)
				}
				data, err := readObj(op.Addr)
				if err != nil {
					return nil, err
				}
				b, err := doc.DecodeBehavior(data)
				if err != nil {
					return nil, err
				}
				st.Behaviors[op.ID] = b
			}
			// A removal may retire a behavior added earlier in the
			// same fold, and a remove-and-re-add of one id may name
			// it twice in one order: the order keeps only the
			// living, each once, first-seen.
			ord := st.Order[:0]
			written := map[string]bool{}
			for _, id := range st.Order {
				if _, alive := st.Behaviors[id]; alive && !written[id] {
					written[id] = true
					ord = append(ord, id)
				}
			}
			st.Order = ord
		case rec.Kind == KindRun && rec.Run != nil:
			if rec.Run.N > st.RunNumber {
				st.RunNumber = rec.Run.N
			}
			for _, rf := range rec.Run.Rows {
				st.Rows[rf.Key] = rf
			}
			st.Lessons = append(st.Lessons, rec.Run.Lessons...)
		case rec.Kind == KindVerdict && rec.Verdict != nil:
			st.Verdict = rec.Verdict
		case rec.Kind == KindRefusal && rec.Refusals != nil:
			st.Refused++
		case rec.Kind == KindCompletion && rec.Completed != nil:
			st.Completions++
		}
	}
	return st, nil
}
