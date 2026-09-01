// Package engine is the protocol engine: it owns the instance state,
// applies deltas transactionally, drives the stages and answers the
// next / submit / why verbs. The sequence of steps always belongs to
// the machine.
package engine

import (
	"os"
	"path/filepath"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/journal"
	"github.com/neurophant/punchtape/internal/ledger"
)

// Engine is the engine of one instance. All paths are relative to
// the instance working directory (workdir).
type Engine struct {
	Workdir string
	Store   *canon.Store
	Journal *journal.Journal
	Ledger  *ledger.Ledger
	// lastSubmitChecks — how many checks the last submit actually ran.
	// lastSubmitProbes — how many probe commands submit ran.
	// lastSubmitRunsMs — the sum of the run times of the last submit's
	// checks/probes: the verdict latency line counts the machine's
	// overhead (submission wall minus this sum); the honest working
	// time of checks is not a gate.
	lastSubmitChecks int
	lastSubmitProbes int
	lastSubmitRunsMs int64
	// lastDeltaKind — the kind of the last submission: the background
	// suite rehearsal needs only code/fix, when a surface exists.
	lastDeltaKind string
	// draftSubmit — the next submission is marked with the origin
	// "draft": a spec draft written by a cheap executor routed by the
	// machine; review and editing belong to the main hand.
	draftSubmit bool
	// lastIdempotent — the last submission was an idempotent repeat of
	// a key: a fact for the ledger and the pain diary (no work was
	// applied).
	lastIdempotent bool
	// lastContentDigest — the content digest of the last submission:
	// written to the journal for idempotency by content.
	lastContentDigest string
	// lastVerdictDrift — the last verdict render found the surface
	// drifted against the acceptance moment: freezing is forbidden,
	// the delivery slot asks for re-verification of the changed
	// instance.
	lastVerdictDrift bool
	// lastRenderDrift — the last slot render found stale boundaries
	// (the "freshness before work" contour): only a fresh next writes
	// the fact to the ledger, the repeat cache stays silent.
	lastRenderDrift []string
	// digestCache — a cache of instance file digests mtime+size→digest
	//: it speeds up re-reads, it is not a source of truth.
	digestCache *digestCache
	// writerLock/writerDepth — the writer lock with depth counting:
	// a submission holds the write monopoly for the whole
	// read-modify-write transaction, inner commits reuse the same lock.
	writerLock  *os.File
	writerDepth int
}

// Init creates a new instance in the working directory.
func Init(workdir string) (*Engine, error) {
	store, err := canon.Init(workdir)
	if err != nil {
		return nil, err
	}
	ensureAgentsRef(workdir)
	return openParts(workdir, store)
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
	return e, nil
}

// openEngine — engine construction without the writer discipline:
// journals are open, the lock is not taken, recovery is not executed.
// For processes that write no journal states (background rehearsal):
// the writer lock does not belong to them, there is no race with the
// hand's verbs.
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

// HeadDigest — the digest of the journal head (empty journal — a
// dash): a key for machine transactions of service modes.
func (e *Engine) HeadDigest() string {
	return e.journalHead()
}

// CommitRunTransaction commits a minimal machine transaction without
// effects: a service-run fact in the journal (the head changes — a
// frozen verdict will be re-frozen at the next render).
func (e *Engine) CommitRunTransaction(submissionKey string) error {
	entry := journalEntryRun(submissionKey)
	return e.commit(entry)
}
