// Package journal — the zone's one source of truth: the append-only
// records and the fold that rebuilds the present from them. Every
// other zone file is a generated image.
// Records carry no wall clock: order is the record's place, identity
// is the record's bytes — the journal of one history is
// byte-identical on every machine.
package journal

import (
	"github.com/neurophant/punchtape/internal/domain"
	"github.com/neurophant/punchtape/internal/domain/doc"
)

// The record kinds.
const (
	KindBirth        = "birth"
	KindSubmission   = "submission"
	KindDemandAppend = "demand-append"
	KindNoNewDemands = "no-new-demands"
	KindRun          = "run"
	KindVerdict      = "verdict"
	KindRefusal      = "refusal"
	KindCompletion   = "completion"
	KindExecution    = "execution-fact"
	KindTreeLock     = "tree-lock"
)

// Record — one journaled fact. Exactly one payload field matches
// the kind.
type Record struct {
	Kind         string         `yaml:"kind"`
	Birth        *Birth         `yaml:"birth,omitempty"`
	Submission   *Submission    `yaml:"submission,omitempty"`
	DemandAppend *DemandAppend  `yaml:"demand-append,omitempty"`
	NoNewDemands *NoNewDemands  `yaml:"no-new-demands,omitempty"`
	Run          *Run           `yaml:"run,omitempty"`
	Verdict      *Verdict       `yaml:"verdict,omitempty"`
	Refusals     []doc.Refusal  `yaml:"refusals,omitempty"`
	Completed    *Completion    `yaml:"completed,omitempty"`
	Execution    *ExecutionFact `yaml:"execution,omitempty"`
	TreeLock     *TreeLock      `yaml:"tree-lock,omitempty"`
}

// ExecutionFact — one execution the machine ran on the agent's
// behalf: the run that carried it, the lines it answered, the
// command, the exit, the stdout's digest and the stdout itself —
// whole up to four kilobytes, the object store's address beyond.
// A run that executes an execution lands its run record, so the
// fact's run always names a recorded run.
type ExecutionFact struct {
	At           int      `yaml:"at"`
	Lines        []string `yaml:"lines"`
	Run          doc.Cmd  `yaml:"run"`
	Exit         int      `yaml:"exit"`
	StdoutDigest string   `yaml:"stdout-digest"`
	Stdout       string   `yaml:"stdout,omitempty"`
	StdoutAddr   string   `yaml:"stdout-addr,omitempty"`
}

// TreeLock — the tree's own two digests of one locked run: the tree
// moved under answer, execute or describe intents. The fact stands
// in the journal forever; the hold it arms lifts when the tree
// returns or the change becomes someone's declared work.
type TreeLock struct {
	Before string `yaml:"before"`
	After  string `yaml:"after"`
}

// DemandAppend — one link of the demand stream's chain: how many new
// demand lines the act consumed and the chain value over them — the
// sha256 of the previous link's own bytes and the new lines joined by
// a newline; the genesis previous is the empty string. Every run
// recomputes the chain over the stream's current bytes: a past line
// that moved breaks its link.
type DemandAppend struct {
	Lines int    `yaml:"lines"`
	Chain string `yaml:"chain"`
}

// NoNewDemands — the signed statement that no new human words came:
// the stream's line count at the signing and the run that signed it.
// A signature over words that did come stands journaled forever —
// the one point where the guard is visibility, not a gate.
type NoNewDemands struct {
	Lines int `yaml:"lines"`
	Run   int `yaml:"run"`
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
// canonical bytes (the requirements and the markers — the demands'
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

// Birth — the instance's origin fact: the machine version that
// created it, the record's one payload. A zone whose birth names
// another machine version is refused before any of its records
// folds. Addr and Digest decode the birth spellings that name an
// object address and digest instead of a version — a birth without
// a version names a foreign machine the same way.
type Birth struct {
	Addr    string `yaml:"addr,omitempty"`
	Digest  string `yaml:"digest,omitempty"`
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
// listed (requirements, markers, loud exclusions, open orders —
// merged by their own keys), the behaviors it added or merged by id
// (each op's object is the behavior as it stands after the merge),
// the ids it retired, its own notes as records of the journal, and
// the address of its own canonical bytes (the submission's byte
// truth for the audit). Every list rides flat: an absent list and
// an empty one both carry nothing — the answer and execute carriers
// live inside the intents themselves.
type Submission struct {
	CorpusVersion int               `yaml:"corpus-version"`
	Build         *string           `yaml:"build,omitempty"`
	Intents       []doc.Intent      `yaml:"intents,omitempty"`
	Notes         []doc.Note        `yaml:"notes,omitempty"`
	Requirements  []doc.Requirement `yaml:"requirements,omitempty"`
	Markers       []doc.Marker      `yaml:"not-a-requirement,omitempty"`
	Exclusions    []doc.Exclusion   `yaml:"uncovered-lines,omitempty"`
	Open          []doc.OpenOrder   `yaml:"open,omitempty"`
	Added         []BehaviorOp      `yaml:"added,omitempty"`
	Removed       []string          `yaml:"removed,omitempty"`
	DocAddr       string            `yaml:"doc-addr"`
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

// LineOutcome — one declared outcome as it stands: the outcome, the
// agent's restatement, the corpus version that declared it.
type LineOutcome struct {
	Outcome string `yaml:"outcome"`
	Prose   string `yaml:"prose"`
	At      int    `yaml:"at"`
}

// Standing — the present rebuilt from the records: the fold's whole
// answer. The behaviors carry their content resolved from the
// object store; the registry carries the demands' account; the rows
// carry their standing fact; the chain carries the demand stream's
// journaled past; the outcomes, answers and executions carry the
// stream's account through its declared carriers — the answer's
// text and the execution's command live inside the intents that
// declared them. The notes arrive with their submissions; the row
// bytes index answers each standing row's digest with its canonical
// bytes.
type Standing struct {
	Born              bool                    `yaml:"born"`
	BirthVersion      string                  `yaml:"birth-version"`
	ChainAppends      []DemandAppend          `yaml:"chain-appends"`       // the chain's links, in order
	ChainHead         string                  `yaml:"chain-head"`          // the last link's value, "" before any
	PastLines         int                     `yaml:"past-lines"`          // the stream lines the chain has journaled
	LastNoNewDemands  int                     `yaml:"last-no-new-demands"` // the run of the last signed none; 0 — none stands
	CorpusVersion     int                     `yaml:"corpus-version"`
	Build             string                  `yaml:"build"`
	Registry          doc.Registry            `yaml:"registry"`
	Behaviors         map[string]doc.Behavior `yaml:"behaviors"`
	Order             []string                `yaml:"order"` // the behaviors' first-seen order
	Rows              map[string]RowFact      `yaml:"rows"`
	RunNumber         int                     `yaml:"run-number"`
	Verdict           *Verdict                `yaml:"verdict"`
	Accepted          int                     `yaml:"accepted"` // the accepted submissions' count; their bytes live in the journal's own records
	Refused           int                     `yaml:"refused"`
	Completions       int                     `yaml:"completions"`
	Lessons           []Lesson                `yaml:"lessons"`
	LineOutcomes      map[int]LineOutcome     `yaml:"line-outcomes"`       // the declared outcome of each journaled line
	AskedLines        map[int]bool            `yaml:"asked-lines"`         // lines that ever carried an open order
	Answers           map[int]string          `yaml:"answers"`             // the standing answer of each line
	Executions        map[int]doc.Cmd         `yaml:"executions"`          // the standing execution command of each line
	LastExecution     *ExecutionFact          `yaml:"last-execution"`      // the last execution the machine ran
	TreeLockViolation *TreeLock               `yaml:"tree-lock-violation"` // the standing tree-lock hold, nil when none
	LastSubmissionAt  int                     `yaml:"last-submission-at"`  // the record index of the last accepted submission
	Notes             []doc.Note              `yaml:"notes"`               // every submission's notes, in arrival order
	RowBytes          map[string]string       `yaml:"row-bytes"`           // each standing row's digest → its canonical bytes
}

// freshStanding — the empty standing the fold opens from.
func freshStanding() *Standing {
	return &Standing{
		Behaviors:    map[string]doc.Behavior{},
		Rows:         map[string]RowFact{},
		LineOutcomes: map[int]LineOutcome{},
		AskedLines:   map[int]bool{},
		Answers:      map[int]string{},
		Executions:   map[int]doc.Cmd{},
		RowBytes:     map[string]string{},
	}
}

// Fold — the records into the standing, one whole replay from the
// genesis. readObj answers an object's bytes by address; a missing
// object is an honest error — the journal names what must exist.
func Fold(records []Record, readObj func(addr string) ([]byte, error)) (*Standing, error) {
	st := freshStanding()
	for at := 0; at < len(records); at++ {
		rec := records[at]
		switch {
		case rec.Kind == KindBirth && rec.Birth != nil:
			st.Born = true
			st.BirthVersion = rec.Birth.Version
		case rec.Kind == KindDemandAppend && rec.DemandAppend != nil:
			st.ChainAppends = append(st.ChainAppends, *rec.DemandAppend)
			st.ChainHead = rec.DemandAppend.Chain
			st.PastLines += rec.DemandAppend.Lines
			st.LastNoNewDemands = 0
		case rec.Kind == KindNoNewDemands && rec.NoNewDemands != nil:
			st.LastNoNewDemands = rec.NoNewDemands.Run
		case rec.Kind == KindSubmission && rec.Submission != nil:
			s := rec.Submission
			st.CorpusVersion = s.CorpusVersion
			st.LastSubmissionAt = at
			if s.Build != nil {
				st.Build = *s.Build
			}
			st.Accepted++
			st.Notes = append(st.Notes, s.Notes...)
			merge := doc.Document{Remove: s.Removed, Requirements: s.Requirements,
				Markers: s.Markers, Exclusions: s.Exclusions}
			if len(s.Open) > 0 {
				merge.Open = StampOpen(s.Open, st.Registry.Open, st.PastLines, st.AskedLines)
			}
			st.Registry = doc.MergeRegistry(st.Registry, merge)
			ApplyCarriers(st, s)
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
				for _, row := range b.Rows {
					rowBytes := doc.MarshalRow(row)
					st.RowBytes[domain.Digest(rowBytes)] = string(rowBytes)
				}
			}
			// A removal may retire a behavior added earlier in the
			// same fold, and a remove-and-re-add of one id may name
			// it twice in one order: NormalizeOrder keeps only the
			// living, each once, first-seen.
			NormalizeOrder(st)
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
		case rec.Kind == KindExecution && rec.Execution != nil:
			st.LastExecution = rec.Execution
		case rec.Kind == KindTreeLock && rec.TreeLock != nil:
			st.TreeLockViolation = rec.TreeLock
		}
	}
	return st, nil
}

// NormalizeOrder — the behaviors' order law: only the living, each
// once, first-seen. The live accept and the fold share this one
// text — the standing's order can never diverge from its replay.
func NormalizeOrder(st *Standing) {
	ord := st.Order[:0]
	written := map[string]bool{}
	for _, id := range st.Order {
		if _, alive := st.Behaviors[id]; alive && !written[id] {
			written[id] = true
			ord = append(ord, id)
		}
	}
	st.Order = ord
}

// StampOpen — the submitted open list beside the standing one:
// repeated (demand, question) pairs keep their standing birth
// facts, new pairs are stamped with the stream's journaled size at
// their birth and their lines join the asked lines.
func StampOpen(submitted, standing []doc.OpenOrder, pastLines int, asked map[int]bool) []doc.OpenOrder {
	return stampNewOrders(submitted, standing, pastLines, asked)
}

// ApplyCarriers — the submission's declared carriers onto the
// standing: the line outcomes, the answers' texts and the
// executions' commands from the intents' own carriers, and the
// tree-lock lift a work intent carries. The fold's own law, called
// once more by the live accept.
func ApplyCarriers(st *Standing, s *Submission) {
	for _, in := range s.Intents {
		if !doc.ValidOutcome(in.Outcome) {
			continue
		}
		nums, err := doc.ParseRanges(in.Lines)
		if err != nil {
			continue
		}
		for _, n := range nums {
			st.LineOutcomes[n] = LineOutcome{Outcome: in.Outcome, Prose: in.Prose, At: s.CorpusVersion}
			switch in.Outcome {
			case doc.OutcomeAnswer:
				st.Answers[n] = in.Text
			case doc.OutcomeExecute:
				st.Executions[n] = in.Run
			}
		}
	}
	for _, in := range s.Intents {
		if in.Outcome == doc.OutcomeWork {
			st.TreeLockViolation = nil
			break
		}
	}
}

// stampNewOrders — the submitted open list beside the standing one:
// a repeated (demand, question) pair keeps its standing birth facts,
// a new pair is stamped with the stream's journaled size at its
// birth and its line joins the asked lines.
func stampNewOrders(submitted, standing []doc.OpenOrder, pastLines int, asked map[int]bool) []doc.OpenOrder {
	out := make([]doc.OpenOrder, len(submitted))
	for i, order := range submitted {
		stamped := order
		for _, prev := range standing {
			if prev.Demand == order.Demand && prev.Question == order.Question {
				stamped.OpenedAtLines = prev.OpenedAtLines
				break
			}
		}
		if stamped.OpenedAtLines == 0 {
			stamped.OpenedAtLines = pastLines
		}
		if nums, err := doc.ParseRanges([]string{order.Demand}); err == nil {
			for _, n := range nums {
				asked[n] = true
			}
		}
		out[i] = stamped
	}
	return out
}
