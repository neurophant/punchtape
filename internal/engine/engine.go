// Package engine is the protocol engine: it owns the instance state,
// applies deltas transactionally, drives the stages and answers the
// next / submit / why verbs. The sequence of steps always belongs to
// the machine.
package engine

import (
	"path/filepath"
	"strings"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/fsport"
	"github.com/neurophant/punchtape/internal/journal"
	"github.com/neurophant/punchtape/internal/ledger"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// Engine is the engine of one instance. All paths are relative to
// the instance working directory (workdir).
type Engine struct {
	Workdir string
	Store   *canon.Store
	Journal *journal.Journal
	Ledger  *ledger.Ledger
	// lastSubmitChecks — how many checks the last submit actually ran.
	lastSubmitChecks int
	// lastDeltaKind — the kind of the last submission.
	lastDeltaKind string
	// lastIdempotent — the last submission was an idempotent repeat of
	// a key: a fact for the ledger and the pain diary (no work was
	// applied).
	lastIdempotent bool
	// lastContentDigest — the content digest of the last submission:
	// written to the journal for forensics.
	lastContentDigest string
	// lastSubmissionKey — the machine key of the last submission.
	lastSubmissionKey string
	// lastVerdictDrift — the last verdict render found the surface
	// drifted against the acceptance moment: freezing is forbidden,
	// the delivery slot asks for re-verification of the changed
	// instance.
	lastVerdictDrift bool
	// digestCache — a cache of instance file digests mtime+size→digest
	//: it speeds up re-reads, it is not a source of truth.
	digestCache *digestCache
	// a submission holds the write monopoly for its whole
	// read-modify-write transaction: every commit of the transaction
	// rides the one lock the submit took.
}

// Init creates a new instance in the working directory.
func Init(workdir string) (*Engine, error) {
	store, err := canon.Init(workdir)
	if err != nil {
		return nil, err
	}
	ensureAgentsRef(workdir)
	ensureReference(workdir)
	// The birth records the contract's own bytes: the sha256 of the
	// AGENTS.md the machine just wrote. Every later render stands on
	// it — the one-line reference while it matches, the FULL contract
	// again the moment the tree's copy is gone or changed (the
	// contract law: the hand pays for the whole contract exactly once, at
	// birth — unless the tree itself breaks the record).
	if st, err := store.State(); err == nil {
		if data, rerr := fsport.Bind.ReadFile(filepath.Join(workdir, "AGENTS.md")); rerr == nil {
			st.ContractSha = yamlio.Digest(data)
			_ = store.PutState(st)
		}
	}
	eng, err := openParts(workdir, store)
	if err != nil {
		return nil, err
	}
	eng.ensureCurrentSet()
	eng.ensureSpecRender()
	return eng, nil
}

// Begin — the birth transaction: the brief text arrives with the
// first next and is fixed in the state (it is not a submission — the
// wish kind died with the one-document grammar). One journal
// transaction under the writer monopoly; the reply is the first
// slot.
func (e *Engine) Begin(brief string) (string, int) {
	started := timeNow()
	brief = strings.TrimSpace(brief)
	if brief == "" {
		reply := canondata.T("submit.reject.brief-empty")
		e.logVerb("next", "rejected", map[string]string{"reason": "empty-brief"}, 0, reply)
		return reply, Rejected
	}
	lock, err := e.acquireWriterRetry()
	if err != nil {
		reply := canondata.T("submit.writer-busy", canondata.M{"err": err.Error()})
		e.logVerb("next", "failure", nil, timeSince(started).Milliseconds(), reply)
		return reply, Failure
	}
	defer e.releaseWriter(lock)
	state, err := e.Store.State()
	if err != nil {
		return canondata.T("cli.fail", canondata.M{"err": firstLine(err)}), Failure
	}
	if state.Wish != nil {
		// The brief is already fixed: a repeated birth phrase is an
		// honest refusal — the wish is not edited through rebirth.
		reply := canondata.T("submit.reject.brief-fixed")
		e.logVerb("next", "rejected", map[string]string{"reason": "brief-already-fixed"}, 0, reply)
		return reply, Rejected
	}
	b := newBuilder(e, state)
	b.state.Wish = &brief
	b.state.Language = detectLanguage(brief)
	b.advance()
	tx := yamlio.Digest([]byte("brief:" + brief))
	entry := journal.Entry{
		Transaction:   tx,
		SubmissionKey: "brief",
		DeltaKind:     "brief",
		WallMs:        timeSince(started).Milliseconds(),
		RecordedAt:    timeNow().UTC(),
		Effects:       b.effects,
	}
	if err := e.commit(entry); err != nil {
		return canondata.T("cli.fail", canondata.M{"err": firstLine(err)}), Failure
	}
	le := ledger.NewEntry("apply", canon.StageIntake)
	le.WallMs = entry.WallMs
	le.Details = map[string]string{"kind": "brief", "outcome": "accepted"}
	if err := e.Ledger.Append(le); err != nil {
		return canondata.T("cli.fail", canondata.M{"err": firstLine(err)}), Failure
	}
	reply, _, err := e.renderNext(mustState(e))
	if err != nil {
		return canondata.T("cli.fail", canondata.M{"err": firstLine(err)}), Failure
	}
	e.storeCachedNext(reply)
	// The birth's ledger reply measures the FULL ceremony the hand
	// received: the CLI injects the contract above the slot (the
	// contract law's one full payment) — the meter counts those bytes
	// where they were paid, in the contract family.
	birthReply := canondata.T("cli.next.contract-inject") + "\n" + AgentsRefText() + "\n" + reply.Text
	e.logVerb("next", "ok", map[string]string{"slot": "birth"}, timeSince(started).Milliseconds(), birthReply)
	return reply.Text, Accepted
}

// Open opens an existing instance; exists=false if there is no
// instance. Opening goes through the writer discipline: an unfinished
// transaction is completed from the journal, files are reconciled with
// the manifest — the canon is always consistent with the journal, a
// manual edit is honestly refused.
func Open(workdir string) (*Engine, bool, error) {
	store, exists, err := canon.Open(workdir)
	if err != nil || !exists {
		return nil, exists, err
	}
	eng, err := openParts(workdir, store)
	if err != nil {
		return nil, true, err
	}
	ensureAgentsRef(workdir)
	ensureReference(workdir)
	return eng, true, nil
}

func openParts(workdir string, store *canon.Store) (*Engine, error) {
	e, err := openEngine(workdir, store)
	if err != nil {
		return nil, err
	}
	if err := e.verifyInstance(); err != nil {
		return nil, err
	}
	// The self-damage cure: machine-written undecodable atoms heal at open
	// (quarantine + journaled deletion), never a dead instance.
	if err := e.healToolBrokenAtoms(); err != nil {
		return nil, err
	}
	return e, nil
}

// openEngine — engine construction without the writer discipline:
// journals are open, the lock is not taken, recovery is not executed.
// For processes that write no journal states: the writer lock does
// not belong to them, there is no race with the executor's verbs.
func openEngine(workdir string, store *canon.Store) (*Engine, error) {
	root := canon.Dir(workdir)
	j, err := journal.Open(filepath.Join(root, "journal.yamll"))
	if err != nil {
		return nil, err
	}
	l, err := ledger.Open(filepath.Join(root, "ledger.yamll"))
	if err != nil {
		return nil, err
	}
	return &Engine{Workdir: workdir, Store: store, Journal: j, Ledger: l}, nil
}
