// Applying submissions: strict decode, idempotency, form validation
// and the "journal ahead of files" transaction. A run submission
// executes its rows inside the same transaction (apply + run +
// checkpoint + state in one writer monopoly).
package engine

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/delta"
	"github.com/neurophant/punchtape/internal/fsport"
	"github.com/neurophant/punchtape/internal/journal"
	"github.com/neurophant/punchtape/internal/ledger"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// Submission outcome codes.
const (
	Accepted = 0
	Rejected = 1
	Failure  = 2
)

// rejectErr — a validation error: one line, entity and reason.
// System failures bypass this type — they have a different reply.
type rejectErr struct {
	entity string
	reason string
}

func (e *rejectErr) Error() string { return e.entity + ": " + e.reason }

func rejection(entity, reason string) error { return &rejectErr{entity: entity, reason: reason} }

func isRejection(err error) bool {
	var r *rejectErr
	return errors.As(err, &r)
}

// Submit handles a submission: strict decode, idempotency, form
// validation and the transactional "journal ahead of files" apply.
// Reply: ACCEPTED with the next slot (on deliver, the verdict — a
// merge of the read loop: the executor needs no separate next) / one
// error line / FAIL.
func (e *Engine) Submit(raw []byte) (string, int) {
	return e.submitCore(raw, true)
}

func (e *Engine) submitCore(raw []byte, withSlot bool) (string, int) {
	return e.submitCoreMode(raw, slotFull, withSlot)
}

func (e *Engine) submitCoreMode(raw []byte, mode slotMode, withSlot bool) (string, int) {
	started := timeNow()
	e.lastIdempotent = false
	// The write monopoly spans the whole submission: state read,
	// validation, apply (including the rows of a run), commit and
	// ledger form one read-modify-write transaction. Busy after short
	// retries is an honest refusal: there is no second writer.
	lock, err := e.acquireWriterRetry()
	if err != nil {
		reply := canondata.T("submit.writer-busy", canondata.M{"err": err.Error()})
		e.logVerb("submit", "failure", nil,
			timeSince(started).Milliseconds(), reply)
		return reply, Failure
	}
	defer e.releaseWriter(lock)
	reply, code := e.submitMode(raw, mode, withSlot)
	outcome := "accepted"
	switch code {
	case Rejected:
		outcome = "rejected"
	case Failure:
		outcome = "failure"
	}
	details := map[string]string{
		"rows": fmt.Sprintf("%d", e.lastSubmitChecks),
	}
	// The rejection reason is a machine fact for the ledger: without
	// a reason the entry is unidentifiable once the live instance is
	// lost (the ERROR line is the reason).
	if outcome == "rejected" {
		details["reason"] = firstErrorLine(reply)
	}
	// Self-described submission source (a voluntary declaration by
	// the caller, the PT_SUBMITTER environment): the machine cannot
	// tell channels apart — an honest "as reported" fact, without
	// trust.
	if v := os.Getenv("PT_SUBMITTER"); v != "" {
		details["submitter-reported"] = v
	}
	// An idempotent replay is a machine fact (no work was applied).
	if e.lastIdempotent {
		details["idempotent"] = "yes"
	}
	if withSlot {
		// The executor waits not only for the transaction but also for the
		// nested slot render — this submission's latency budget
		// includes the slot term.
		details["slot"] = "yes"
	}
	e.logVerb("submit", outcome, details, timeSince(started).Milliseconds(), reply)
	// The verdict frozen while building the reply (the slot tail
	// renders the verdict BEFORE this submission's counter booking is
	// written to the ledger) is re-rendered after logVerb: bill must
	// see the tail of the latest submission; the slot cache is
	// updated in the same pass — the nested verdict matches bit for
	// bit.
	e.refreshDeliverRenders()
	if code == Accepted && !e.lastIdempotent {
		// The spec render follows every accepted submission.
		e.ensureSpecRender()
	}
	return reply, code
}

// refreshDeliverRenders — the delivery renders follow the committed
// transaction: the verdict itself is the frozen snapshot (immutable
// by construction), the summary and the spec render are regenerated
// from the now-current store.
func (e *Engine) refreshDeliverRenders() {
	st, err := e.Store.State()
	if err != nil || st.Stage != canon.StageDeliver {
		return
	}
	writeSummaryRender(e)
	e.ensureSpecRender()
}

func (e *Engine) submitMode(raw []byte, mode slotMode, withSlot bool) (string, int) {
	started := timeNow()
	e.lastSubmitChecks = 0

	// The template catch: the door file that
	// still carries the machine's own blank template (or nothing at
	// all) is NOT a submission — the template is grammatically valid,
	// and reading it once drove a whole instance off its rails (the
	// conduit door accidents). Byte-exact comparison with the
	// machine's own template bytes: no semantics, no guessing.
	if len(raw) == 0 || bytes.Equal(raw, canondata.TemplateBytes(inboxTemplate)) {
		return reject("door", canondata.T("submit.error.template-resubmitted")), Rejected
	}

	d, err := delta.Decode(raw)
	if err != nil {
		// The rejection carries the decoder's own address ([line:col])
		// and, when the kind head names a near-miss, a hint.
		return reject("delta", firstLine(err)), Rejected
	}
	e.lastDeltaKind = d.Kind
	if d.Patch != nil && d.FullReplace {
		return reject("delta", canondata.T("submit.error.patch-with-full-replace")), Rejected
	}
	// The machine derives its own submission key from the content
	// digest — the key-numbering ritual is lifted from the executor, and
	// the machine itself admits no content collisions.
	state, err := e.Store.State()
	if err != nil {
		return failure(err), Failure
	}
	if d.Patch != nil && state.Counters["SPEC"] == 0 {
		// a patch needs the table it patches: the first spec is the
		// one FULL document (the owner's economy law)
		return reject("delta", canondata.T("submit.error.patch-no-spec")), Rejected
	}
	if !delta.KindAllowedOnStage(d.Kind, state.Stage) {
		return reject("delta", stageRejectReason(d.Kind, state.Stage)), Rejected
	}
	// The submission's machine key: the content digest; a spec key
	// also carries the corpus the document IMPLIES (its behavior ids)
	// — the replay gate then checks the arrival corpus against it
	// (spec-as-state: the same document against a corpus it already
	// describes is a repeat; against a different corpus it is a new
	// version — it retires what the document no longer names). A
	// PATCH is its own content: the same patch bytes are the same
	// fact whatever the corpus (the patch idempotency law).
	cd := e.contentDigest(d)
	specRepeat := false
	if d.Kind == delta.KindSpec && d.Patch == nil {
		postIDs := behaviorIDs(d.Behaviors)
		sort.Strings(postIDs)
		cd = shortHex(yamlio.Digest([]byte(cd + ":" + strings.Join(postIDs, "\x00"))))
		specRepeat = e.corpusIDsEqual(postIDs)
	}
	e.lastContentDigest = cd
	key := "auto-" + shortHex(cd)

	// Idempotency by content, except for the bare re-run driver: a
	// document of kind run that carries no behaviors, no patch and no
	// build declaration exists to execute (its whole content is "run
	// again") — the same bytes after a fix of the tree are a NEW
	// application, not a replay. Everything else: the same document is
	// the same fact; a replay does not inflate the submission
	// counters. The slot in the reply is the current one (after a
	// lost reply it is exactly what the executor needs), while the
	// transaction is not repeated.
	bareRerun := d.Kind == delta.KindRun && d.Patch == nil && len(d.Behaviors) == 0 &&
		d.Build == nil && d.BuildIgnore == nil
	if entry, ok := e.Journal.FindByKey(key); ok && !bareRerun &&
		(d.Patch != nil || d.Kind != delta.KindSpec || specRepeat) {
		e.lastIdempotent = true
		// Healing phrase: a predictable agent failure — a replay — the
		// machine answers normally from data, the executor does not spin.
		// The header is IDEMPOTENT REPLAY, not ACCEPTED: the replay
		// applied nothing, and an honest header does not claim an
		// apply.
		ack := canondata.T("submit.idempotent.ack", canondata.M{
			"tx": shortHex(entry.Transaction), "sha": e.lastContentDigest,
		})
		if advice := canondata.Heal("idempotent-resubmit"); advice != "" {
			ack = ack + "\n" + advice
		}
		return ack, Accepted
	}
	// The implicit full-document repeat refusal: the full table rode
	// the FIRST spec exactly once; a later
	// full document is legal only as the deliberate full-replace. A
	// kind-spec document is the full state by its own law; a kind-run
	// document is full exactly when its behaviors cover every corpus
	// atom (a partial upsert stays the cheap legal repair it always
	// was).
	if d.Patch == nil && !d.FullReplace && len(d.Behaviors) > 0 && state.Counters["SPEC"] > 0 {
		fullTable := d.Kind == delta.KindSpec
		if !fullTable {
			if corpusAtoms, aerr := e.Store.Atoms(); aerr == nil && len(corpusAtoms) > 0 {
				named := map[string]bool{}
				for _, beh := range d.Behaviors {
					named[beh.ID] = true
				}
				fullTable = true
				for _, a := range corpusAtoms {
					if !named[a.ID] {
						fullTable = false
						break
					}
				}
			}
		}
		if fullTable {
			return reject("delta", canondata.T("submit.error.full-resubmit")), Rejected
		}
	}
	e.lastSubmissionKey = key
	if bareRerun {
		// a fresh key per application: the position makes the identity
		e.lastSubmissionKey = key + "-" + shortHex(yamlio.Digest([]byte(e.journalHead())))
	}

	tx, replyTail, err2 := e.buildAndApply(d, raw, state, started)
	if err2 != nil {
		if isRejection(err2) {
			r := err2.(*rejectErr)
			return reject(r.entity, r.reason), Rejected
		}
		return failure(err2), Failure
	}
	// The thin ack: status, transaction, the accepted bytes' sha,
	// the gate names, the measured one-liners (the suite line and the
	// red rows of a run; the verdict of a ratification). No slot
	// body, no submission content — the slot rides punchtape next
	// (the cache makes the follow-up free).
	gates := "form"
	switch {
	case d.Ratify:
		gates = "freshness, digest, verdict"
	case d.Kind == delta.KindRun:
		gates = "suite, tree"
	}
	table := replyTail
	if table != "" {
		table = table + "\n"
	}
	return canondata.T("submit.accepted.ack", canondata.M{
		"tx":    shortHex(tx),
		"sha":   e.lastContentDigest,
		"gates": gates,
		"table": table,
	}), Accepted
}

// contentDigest — the content digest of a submission document: the
// canonical bytes of the decoded doc (the exactness law owns the
// bytes; two spellings of one content are one content).
func (e *Engine) contentDigest(d *delta.Doc) string {
	// A decoded doc always marshals: plain data
	data, _ := yamlio.Marshal(d)
	return yamlio.Digest(data)
}

// buildAndApply validates the document on canon projections, applies
// it (a run executes its stale rows in the same transaction), freezes
// the verdict snapshot of a successful ratification, advances the
// stages and records the transaction in the journal with full
// effects. replyTail — the diff block (spec) or the run table (run)
// for the reply to the agent.
func (e *Engine) buildAndApply(d *delta.Doc, raw []byte, state canon.State, started time.Time) (string, string, error) {
	// The stage is recorded before advancing: a submission is a fact
	// of the stage it arrived at, even if the machine immediately ran
	// the cycle to deliver.
	submitStage := state.Stage

	b := newBuilder(e, state)
	b.rawInput = string(raw)
	doc := d
	if d.Patch != nil {
		// The sparse path: expand the patch into the full
		// document a full re-submission of the same change would have
		// been — the apply law stays ONE, the canon bytes stay equal
		// to the full-document path by construction.
		if err := b.load(); err != nil {
			return "", "", err
		}
		expanded, force, err := expandPatchDoc(b.atoms, d)
		if err != nil {
			return "", "", err
		}
		doc = expanded
		b.forceRun = force
		b.patched = true
	}
	if d.FullReplace {
		// the deliberate full rewrite: the byte-count confirmation
		// stands on the corpus AS IT STOOD (load before the apply)
		if err := b.load(); err != nil {
			return "", "", err
		}
		for _, a := range b.atoms {
			b.fullReplaceOld += len(a.Rows)
		}
		for _, beh := range d.Behaviors {
			b.fullReplaceNew += len(beh.Rows)
		}
		b.fullReplaced = true
	}
	if err := b.applyDoc(doc); err != nil {
		return "", "", err
	}
	// A run executes after the apply, inside the same monopoly: the
	// upserted rows are the rows that run, each at its own freshness.
	// Its own substrate failures (build, snapshot) are executed into
	// env reds — runRows cannot fail, there is no error to branch on.
	// A ratification NEVER executes (the signature law): the
	// predicate signs the recorded green or refuses with the names.
	switch {
	case d.Kind == delta.KindRun && d.Ratify:
		if err := b.ratifyPredicate(); err != nil {
			return "", "", err
		}
	case d.Kind == delta.KindRun:
		// The form gate: a corpus that pins formal surface (a
		// nonzero rc, an exact out/err/state text) without its
		// one-line source closes the run door — the hand cites the
		// brief/prereg phrase or asks the operator; the machine never
		// judges the citation, silence is the only red.
		if unresolved := b.unresolvedFormLabels(); len(unresolved) > 0 {
			return "", "", rejection("form", canondata.T("format.error.form-unresolved", canondata.M{
				"labels": budgetJoin(unresolved),
			}))
		}
		b.runRows()
	}
	b.advance()
	// A successful ratification freezes the case's next verdict
	// snapshot inside the same transaction: vN is pinned to its
	// digest forever, the case lives on (the next spec version
	// reopens it through the same door).
	verdictVersion := 0
	if d.Ratify && b.state.Stage == canon.StageDeliver {
		n, err := b.freezeVerdictSnapshot()
		if err != nil {
			return "", "", err
		}
		verdictVersion = n
		// the verdict's own one line rides the ack (the digest lines
		// are the verdict's law, never cut); the full snapshot is the
		// frozen file and the deliver slot
		if v, verr := b.e.RenderVerdict(); verr == nil {
			b.runTable = append(b.runTable, firstLineOf(v.Text())+" (v"+strconv.Itoa(n)+")")
		}
	}
	// The acceptance moment: all atoms green finalizes the full set
	// on this tree, the digest signs it.
	b.acceptanceNote(b.lastCkp)

	tx := yamlio.Digest(raw)
	entry := journal.Entry{
		Transaction:    tx,
		SubmissionKey:  b.e.lastSubmissionKey,
		DeltaKind:      d.Kind,
		WallMs:         timeSince(started).Milliseconds(),
		RecordedAt:     timeNow().UTC(),
		Effects:        b.effects,
		ContentDigest:  e.lastContentDigest,
		SpecVersion:    b.specVersion,
		VerdictVersion: verdictVersion,
	}
	if b.specDiffOut != nil {
		entry.Retired = b.specDiffOut.retiredRows
	}
	if err := e.commit(entry); err != nil {
		return "", "", fmt.Errorf("transaction: %w", err)
	}

	// The ledger holds machine facts; by this point the canon is
	// consistent. The sparse path and the deliberate full rewrite are
	// machine facts the cell analytics reads.
	le := ledger.NewEntry("apply", submitStage)
	le.WallMs = entry.WallMs
	le.Details = map[string]string{"kind": d.Kind, "outcome": "accepted"}
	if b.patched {
		le.Details["patch"] = "yes"
	}
	if b.fullReplaced {
		le.Details["full-replace"] = strconv.Itoa(b.fullReplaceOld) +
			"->" + strconv.Itoa(b.fullReplaceNew)
	}
	if b.specVersion > 0 && b.specDiffOut != nil {
		le.Details["spec-version"] = strconv.Itoa(b.specVersion)
		le.Details["diff"] = b.specDiffOut.summaryLine()
	}
	if err := e.Ledger.Append(le); err != nil {
		return "", "", fmt.Errorf("ledger: %w", err)
	}
	return tx, b.replyTail(), nil
}

// --- transaction builder ---

// builder holds canon projections and accumulates file effects. A
// validation error aborts the build entirely: partial applies do not
// exist.
type builder struct {
	e       *Engine
	state   canon.State
	atoms   map[string]canon.Atom
	effects []journal.Effect
	touched []string
	loaded  bool
	// kind — the submission's kind (spec | run)
	kind string
	// specVersion — the spec version this submission created (spec
	// kind only; 0 — a run)
	specVersion int
	// specDiffOut — the machine-computed diff of this submission
	// against the corpus it arrived at (spec kind: the full state
	// diff; run kind: the repair's upsert diff)
	specDiffOut *specDiff
	// lastCkp — the run's pre-run tree snapshot (the outcomes and
	// the acceptance stand on it)
	lastCkp *Checkpoint
	// ranRows — the row labels this transaction executed (a
	// transaction that ran nothing finalizes no acceptance)
	ranRows int
	// keptRows — the rows whose recorded green stood at unchanged
	// inputs (the memoization savings of this run)
	keptRows int
	// ratifying — this submission carries ratify: true
	ratifying bool
	// runTable — the reply lines of a run: the suite line first,
	// one line per red row after
	runTable []string
	// buildFailed — the run's build refused to run: every row is a
	// not-run red
	buildFailed string
	// raw delta of this submission verbatim (renders)
	rawInput string
	// forceRun — the post-expansion labels of the rows the patch's
	// run section forces to execute (a patch fact; nil — none)
	forceRun map[string]bool
	// patched — this submission rode the sparse-patch path
	patched bool
	// fullReplaced/fullReplaceOld/fullReplaceNew — the deliberate
	// full-rewrite hatch: the event's row counts for the ack and the
	// cell ledger
	fullReplaced                   bool
	fullReplaceOld, fullReplaceNew int
}

func newBuilder(e *Engine, state canon.State) *builder {
	return &builder{e: e, state: state, atoms: map[string]canon.Atom{}}
}

// load reads the projections from disk once per transaction.
func (b *builder) load() error {
	if b.loaded {
		return nil
	}
	atoms, err := b.e.Store.Atoms()
	if err != nil {
		return fmt.Errorf("system failure: %w", err)
	}
	for _, a := range atoms {
		b.atoms[a.ID] = a
	}
	b.loaded = true
	return nil
}

func (b *builder) putAtom(a canon.Atom) error {
	// The record's own digest rides inside it
	a.Digest = a.DigestOf()
	// Atom is plain scalar/slice/struct data: marshal cannot fail
	data, _ := yamlio.Marshal(a)
	// The machine never writes what it cannot read back: the
	// bytes enter the journal only through the round-trip gate.
	if err := atomGate(data); err != nil {
		return rejection(a.ID, err.Error())
	}
	b.atoms[a.ID] = a
	b.effects = append(b.effects, journal.Effect{
		Path: canonFile("atoms", a.ID), Content: machineFile(data),
	})
	b.touch(a.ID)
	return nil
}

func (b *builder) putState() {
	// State is plain scalar/slice data: marshal cannot fail
	data, _ := yamlio.Marshal(b.state)
	b.effects = append(b.effects, journal.Effect{
		Path: ".punchtape/state.yaml", Content: machineFile(data),
	})
}

func (b *builder) touch(id string) {
	for _, t := range b.touched {
		if t == id {
			return
		}
	}
	b.touched = append(b.touched, id)
}

// replyTail — the submission's measured one-liners for the thin ack
// (the thin-ack law): the spec kind answers its version-counts line (the
// named diff lines ride the next slot's changed list, not the ack);
// the run kind answers the suite line and the red rows (the reds
// are the hand's work list, measured bytes — not submission
// content); a ratification adds its signature line and the verdict.
func (b *builder) replyTail() string {
	// the deliberate full rewrite's byte-count confirmation rides
	// every kind's tail (an event the ledger mirrors)
	tail := ""
	if b.fullReplaced {
		tail = canondata.T("submit.full-replace.ack", canondata.M{
			"old": strconv.Itoa(b.fullReplaceOld),
			"new": strconv.Itoa(b.fullReplaceNew),
		})
	}
	if b.kind == delta.KindSpec {
		if tail != "" {
			tail += "\n"
		}
		tail += b.specDiffOut.summaryRender(b.specVersion)
		// the ratification render's own marker (the form law): the spec
		// landed, its unresolved form decisions are named, the run
		// door stays closed until each carries its source line
		if unresolved := b.unresolvedFormLabels(); len(unresolved) > 0 {
			tail += "\n" + canondata.T("submit.form-unresolved", canondata.M{
				"labels": budgetJoin(unresolved),
			})
		}
		return tail
	}
	if tail != "" {
		b.runTable = append([]string{tail}, b.runTable...)
	}
	return strings.Join(b.runTable, "\n")
}

// unresolvedFormLabels — the corpus's rows that pin formal surface
// without a source line, in label order.
func (b *builder) unresolvedFormLabels() []string {
	// the caller's applyDoc already loaded the corpus
	var out []string
	for _, id := range sortedIDs(b.atoms) {
		for i, row := range b.atoms[id].Rows {
			if row.FormUnresolved() {
				out = append(out, canon.RowLabel(id, i+1))
			}
		}
	}
	return out
}

// --- applying the document ---

// applyDoc — the form validation and the corpus state change. The
// spec kind IS the state: the document's behaviors replace the
// corpus (retirements are visible frames); the run kind upserts
// (repairs never retire). Everything the machine can check without
// running the product is checked here: a run refuses on the same
// form errors as a spec. The predicates live in docFindings — ONE
// law shared with the dry validator (spec --check); the door rejects
// on the first finding, the validator names them all.
func (b *builder) applyDoc(d *delta.Doc) error {
	if err := b.load(); err != nil {
		return err
	}
	// The submission's own build-ignore is the effective surface for
	// its rows' dependency check inside docFindings (the surface law
	// reads it there).
	if d.BuildIgnore != nil {
		b.state.BuildIgnore = d.BuildIgnore
	}
	if findings := docFindings(d, b.docContext()); len(findings) > 0 {
		return rejection(findings[0].entity, findings[0].reason)
	}
	b.kind = d.Kind
	if err := b.applySpecState(d); err != nil {
		return err
	}
	if d.Build != nil {
		b.state.Build = d.Build
	}
	if d.Kind == delta.KindSpec {
		// the living case's version counter: this document is version N
		// (the map exists from the instance's birth)
		b.state.Counters["SPEC"]++
		b.specVersion = b.state.Counters["SPEC"]
	}
	b.state.Touched = behaviorIDs(d.Behaviors)
	b.ratifying = d.Ratify
	return nil
}

// docFinding — one document-level form finding: the entity address
// (delta, behaviors[2], B1/3) and the reason line.
type docFinding struct {
	entity string
	reason string
}

// docContext — the document's surroundings for the form rules: the
// work tree the dependency walk stats, the effective build-ignore
// (the document's own when it declares one, else the stored state's),
// and the live corpus's ids (the case-collision law reads them).
type docContext struct {
	workdir     string
	buildIgnore []string
	corpusLower map[string]string
	corpusCount int
	statInput   func(rel string) (os.FileInfo, error)
	statInbox   func(rel string) (os.FileInfo, error)
}

// docContext — the door's own context: the loaded corpus, the
// effective ignore, the input and inbox directories.
func (b *builder) docContext() docContext {
	lower := map[string]string{}
	for _, id := range sortedIDs(b.atoms) {
		lower[strings.ToLower(id)] = id
	}
	return docContext{
		workdir:     b.e.Workdir,
		buildIgnore: b.state.BuildIgnore,
		corpusLower: lower,
		corpusCount: len(b.atoms),
		statInput:   b.e.statInput,
		statInbox:   b.e.statInboxSample,
	}
}

// docFindings — EVERY document-level form rule, in the door's
// rejection order: the kind's own laws, then the behaviors' ids and
// prose, then every row's command, expectations (the directness
// law), dependencies (the surface law), seeds, state and samples —
// including the existence of every external file the rows reference.
// The door refuses on the first finding (one honest line); the dry
// validator (spec --check) collects them all — the same predicates,
// one vocabulary, no second law to drift.
func docFindings(d *delta.Doc, dc docContext) []docFinding {
	var out []docFinding
	add := func(entity, reason string) {
		out = append(out, docFinding{entity: entity, reason: reason})
	}
	if d.Ratify && d.Kind != delta.KindRun {
		add("delta", canondata.T("format.error.ratify-not-run"))
		return out
	}
	if len(d.Answers) > 0 {
		// No machine question batch can be open in this generation:
		// an answer without a question is noise, not silence.
		add("answers", canondata.T("format.error.answers-no-batch"))
		return out
	}
	if d.Kind == delta.KindSpec && len(d.Behaviors) == 0 {
		add("behaviors", canondata.T("format.error.behaviors-empty"))
		return out
	}
	if d.Kind == delta.KindRun && d.Ratify && dc.corpusCount == 0 && len(d.Behaviors) == 0 {
		add("behaviors", canondata.T("format.error.behaviors-empty"))
		return out
	}
	seen := map[string]bool{}
	// The canon file name is the id's lowercase spelling: two ids
	// differing only in case would share one file — the collision is
	// refused at the form, inside the document and against the corpus.
	lower := map[string]string{}
	for k, v := range dc.corpusLower {
		lower[k] = v
	}
	for i, beh := range d.Behaviors {
		where := fmt.Sprintf("behaviors[%d]", i)
		if !behaviorIDRe.MatchString(beh.ID) {
			add(where, canondata.T("format.error.behavior-id", canondata.M{
				"id": fmt.Sprintf("%q", beh.ID),
			}))
			continue
		}
		if seen[beh.ID] {
			add(where, canondata.T("format.error.behavior-dup", canondata.M{
				"id": beh.ID,
			}))
			continue
		}
		if seenLower, ok := lower[strings.ToLower(beh.ID)]; ok && seenLower != beh.ID {
			// a DIFFERENT spelling of the same lowercase name: the same
			// spelling is the legal replace-by-id (upsert)
			add(where, canondata.T("format.error.behavior-dup-case", canondata.M{
				"id": beh.ID,
			}))
			continue
		}
		lower[strings.ToLower(beh.ID)] = beh.ID
		seen[beh.ID] = true
		if strings.TrimSpace(beh.Name) == "" {
			add(beh.ID, canondata.T("format.error.behavior-name-empty"))
			continue
		}
		if strings.TrimSpace(beh.Prose) == "" {
			add(beh.ID, canondata.T("format.error.behavior-prose-empty"))
			continue
		}
		if len(beh.Rows) == 0 {
			add(beh.ID, canondata.T("format.error.behavior-rows-empty"))
			continue
		}
		checkRowFindings(beh, dc, add)
	}
	return out
}

// checkRowFindings — the per-row form rules of one behavior.
func checkRowFindings(beh delta.Behavior, dc docContext, add func(entity, reason string)) {
	for ri, row := range beh.Rows {
		label := canon.RowLabel(beh.ID, ri+1)
		fields, err := row.CommandFields()
		if err != nil {
			add(label, err.Error())
			continue
		}
		if len(fields) == 0 {
			add(label, canondata.T("delta.error.run-argv-empty"))
			continue
		}
		if row.Timeout < 0 || row.Timeout > canondata.Limit("check.timeout-max-s") {
			add(label, canondata.T("format.error.timeout-range", canondata.M{
				"got": fmt.Sprintf("%d", row.Timeout),
				"max": fmt.Sprintf("%d", canondata.Limit("check.timeout-max-s")),
			}))
			continue
		}
		if row.Rc != nil && *row.Rc < 0 {
			add(label, canondata.T("format.error.rc-negative"))
			continue
		}
		// The machine's own records are never submission content:
		// outcomes and memos are written by runs (a forged green
		// or a forged freshness key would stand where nothing ran).
		if row.Outcome != "" || row.Reason != "" || row.EnvFail || row.Memoized() {
			add(label, canondata.T("format.error.machine-fields"))
			continue
		}
		// The stream acceptance law: expectations are DIRECT —
		// no echo of the harness's own verdict lines, no unmarked
		// short fragments.
		if err := checkExpectationQuality(label, row.Out); err != nil {
			if r, ok := err.(*rejectErr); ok {
				add(r.entity, r.reason)
			}
			continue
		}
		if err := checkExpectationQuality(label, row.Err); err != nil {
			if r, ok := err.(*rejectErr); ok {
				add(r.entity, r.reason)
			}
			continue
		}
		// The surface acceptance law: no argument may depend on bytes
		// outside the verdict's signed surface.
		if err := rowDependencies(dc, label, fields, row); err != nil {
			if r, ok := err.(*rejectErr); ok {
				add(r.entity, r.reason)
			}
			continue
		}
		for si, s := range row.Seeds {
			if err := validRelPath(s.Path); err != nil {
				add(label, canondata.T("format.error.seed-path", canondata.M{
					"index": fmt.Sprintf("%d", si+1), "err": err.Error(),
				}))
				continue
			}
			if s.FromInput != "" {
				if err := validRelPath(s.FromInput); err != nil {
					add(label, canondata.T("format.error.from-input-path", canondata.M{
						"index": fmt.Sprintf("%d", si+1), "err": err.Error(),
					}))
					continue
				}
				if _, err := dc.statInput(s.FromInput); err != nil {
					add(label, canondata.T("format.error.from-input-missing", canondata.M{
						"path": s.FromInput,
					}))
					continue
				}
			}
		}
		for si, st := range row.State {
			if err := validRelPath(st.Path); err != nil {
				add(label, canondata.T("format.error.state-path", canondata.M{
					"index": fmt.Sprintf("%d", si+1), "err": err.Error(),
				}))
				continue
			}
			if st.FromRun != "" {
				if err := validRelPath(st.FromRun); err != nil {
					add(label, canondata.T("format.error.from-run-path", canondata.M{
						"index": fmt.Sprintf("%d", si+1), "err": err.Error(),
					}))
					continue
				}
			}
		}
		if row.Bytes != "" {
			ref := canon.Row(row).BytesRef()
			if ref == "" || ref != row.Bytes[len("file://"):] {
				add(label, canondata.T("format.error.bytes-form"))
				continue
			}
			if err := validRelPath(ref); err != nil {
				add(label, canondata.T("format.error.bytes-path", canondata.M{
					"err": err.Error(),
				}))
				continue
			}
			if _, err := dc.statInbox(ref); err != nil {
				add(label, canondata.T("format.error.bytes-missing", canondata.M{
					"path": ref,
				}))
			}
		}
		// The witness probe rides the door like a bytes sample: the
		// file:// form names the staged bytes next to the submission
		// in the inbox — the machine checks existence, never content.
		if row.Witness != nil {
			ref := row.ProbeRef()
			if ref == "" || ref != strings.TrimPrefix(row.Witness.Probe, "file://") {
				add(label, canondata.T("format.error.probe-form"))
				continue
			}
			if err := validRelPath(ref); err != nil {
				add(label, canondata.T("format.error.probe-path", canondata.M{
					"err": err.Error(),
				}))
				continue
			}
			if _, err := dc.statInbox(ref); err != nil {
				add(label, canondata.T("format.error.probe-missing", canondata.M{
					"path": ref,
				}))
			}
		}
	}
}

// behaviorIDRe — the executor's short identifier: one leading
// alphanumeric, then the identifier alphabet, at most 32 runes.
var behaviorIDRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,31}$`)

// validRelPath — a path inside the work tree: relative, no escape, no
// home shortcut.
func validRelPath(p string) error {
	if p == "" {
		return errors.New(canondata.T("canonvalid.error.path-empty"))
	}
	if strings.HasPrefix(p, "/") || strings.HasPrefix(p, "~") {
		return errors.New(canondata.T("canonvalid.error.path-absolute", canondata.M{"path": p}))
	}
	for _, seg := range strings.Split(filepath.ToSlash(p), "/") {
		if seg == ".." {
			return errors.New(canondata.T("canonvalid.error.path-escape", canondata.M{"path": p}))
		}
	}
	return nil
}

// statInput — the input directory entry of a named file
// (the input directory is the sibling directory input/ of the work
// root; absent in a bare dev instance — an honest form rejection
// names it).
func (e *Engine) statInput(rel string) (os.FileInfo, error) {
	return fsStat(filepath.Join(e.Workdir, "..", "input", filepath.FromSlash(rel)))
}

// statInboxSample — the submission's sample file (it lies next to
// the door file in the machine's inbox).
func (e *Engine) statInboxSample(rel string) (os.FileInfo, error) {
	return fsStat(filepath.Join(e.Workdir, ".punchtape", "inbox", filepath.FromSlash(rel)))
}

func fsStat(p string) (os.FileInfo, error) {
	info, err := fsport.Bind.Stat(p)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, fmt.Errorf("%s is a directory", p)
	}
	return info, nil
}

// behaviorIDs — the submitted behaviors' ids in document order.
func behaviorIDs(behs []delta.Behavior) []string {
	ids := make([]string, 0, len(behs))
	for _, b := range behs {
		ids = append(ids, b.ID)
	}
	return ids
}

// stageRejectReason — a kind rejection by stage also names the legal
// way forward: a formal rejection without a route leaves the executor
// without a next step. The two illegal combinations: a run at deliver
// (the case lives on — only the next spec version reopens the work)
// and a run before any accepted spec.
func stageRejectReason(kind, stage string) string {
	if stage == canon.StageDeliver {
		return canondata.T("submit.reject.stage-deliver", canondata.M{"kind": kind})
	}
	return canondata.T("submit.reject.run-no-spec")
}

// --- helpers ---

// canonFile — the canon file path in effects (from the project root).
func canonFile(dir, id string) string {
	return path.Join(".punchtape", "canon", dir, strings.ToLower(id)+".yaml")
}

// machineFile — instance data file content with the manifest header:
// the header rides inside the effect, the journal and the manifest
// reconcile the very same one — the "file = journal" invariant does
// not diverge.
func machineFile(data []byte) string {
	return string(data)
}

func sortedIDs[T any](m map[string]T) []string {
	ids := make([]string, 0, len(m))
	for id := range m {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// reject assembles the rejection reply: the diagnosis line alone.
// The form is taught by the template in the inbox and the reference
// docs, never by a skeleton in the console.
func reject(entity, reason string) string {
	return canondata.T("submit.reject.line", canondata.M{"entity": entity, "reason": reason})
}

func failure(err error) string {
	return canondata.T("submit.fail", canondata.M{"err": firstLine(err)})
}

func firstLine(err error) string {
	line := err.Error()
	if i := strings.IndexByte(line, '\n'); i >= 0 {
		line = line[:i]
	}
	return strings.TrimSpace(line)
}

// firstErrorLine — the first ERROR line of a multi-line reply.
func firstErrorLine(reply string) string {
	for _, line := range strings.Split(reply, "\n") {
		if strings.HasPrefix(line, "ERROR:") {
			return strings.TrimSpace(line)
		}
	}
	line := reply
	if i := strings.IndexByte(line, '\n'); i >= 0 {
		line = line[:i]
	}
	return strings.TrimSpace(line)
}

func shortHex(hex string) string {
	if len(hex) > 12 {
		return hex[:12]
	}
	return hex
}

// detectLanguage — script detection of the brief: Han characters ≥
// the delivery threshold → zh; otherwise Cyrillic letters ≥ the
// threshold → ru; otherwise en. Han is checked first, so a mixed
// script resolves deterministically. No guessing: only countable
// letters.
func detectLanguage(text string) string {
	letters, cyr, han := 0, 0, 0
	for _, r := range strings.ToLower(text) {
		if (r >= 'a' && r <= 'z') || (r >= '\u0430' && r <= '\u044f') {
			letters++
			if r >= '\u0430' && r <= '\u044f' {
				cyr++
			}
		} else if r >= '\u4e00' && r <= '\u9fff' {
			letters++
			han++
		}
	}
	if letters == 0 {
		return "en"
	}
	if 100*han/letters >= canondata.Limit("language.han-share") {
		return "zh"
	}
	if 100*cyr/letters >= canondata.Limit("language.cyrillic-share") {
		return "ru"
	}
	return "en"
}

// selfReportTokens — the bare verdict tokens of a checking harness:
// an expectation whose every line is one of these (plus a label)
// echoes the harness's report about its own checks, not the product's
// behavior (the echo episodes: a scenario's PASS self-prints standing
// for real expectations, and constant-true checks surfacing as the
// same PASS lines).
var selfReportTokens = map[string]bool{
	"pass": true, "passed": true, "fail": true, "failed": true,
	"ok": true, "okay": true, "success": true, "done": true,
	"skipped": true,
}

// echoExpectation — whether every non-empty line of the text begins
// with a self-report token as a whole word (the leading alphanumeric
// run IS the token, so longer words — passwords, ok3files — never
// match): the expectation is a verdict about the checks, not a
// behavior of the product.
func echoExpectation(text string) bool {
	any := false
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimLeft(line, " \t")
		if line == "" {
			continue
		}
		any = true
		i := 0
		for i < len(line) && isASCIIAlnum(line[i]) {
			i++
		}
		if i == 0 || !selfReportTokens[strings.ToLower(line[:i])] {
			return false
		}
	}
	return any
}

func isASCIIAlnum(b byte) bool {
	return b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b >= '0' && b <= '9'
}

// checkExpectationQuality — the acceptance law for one stream
// expectation of a row: expectations are DIRECT. An echo expectation
// is refused outright; a contains fragment shorter than the canon
// limit is refused unless it carries the executor's explicit loose
// mark — short and imprecise is legal only under its honest name.
func checkExpectationQuality(label string, e *canon.TextExpect) error {
	if e == nil {
		return nil
	}
	if echoExpectation(e.Text) {
		return rejection(label, canondata.T("format.error.expect-echo"))
	}
	if e.Contains && !e.Loose &&
		utf8.RuneCountInString(e.Text) < canondata.Limit("expect.contains-min-runes") {
		return rejection(label, canondata.T("format.error.expect-contains-loose", canondata.M{
			"text": e.Text,
			"min":  fmt.Sprintf("%d", canondata.Limit("expect.contains-min-runes")),
		}))
	}
	return nil
}

// rowDependencies — the surface law: a row's run may not name an existing
// work-tree FILE outside the verdict's signed surface (a build-ignore
// pattern covers the file or an ancestor directory) unless the row
// itself carries the bytes as its seed. A green measured on unsigned
// fixtures travels nowhere: on any other tree the digest does not
// resolve and the file is absent (the registry's episode: a two-file
// digest while four rows measured ignored fixtures). The FIRST field
// is exempt — the executed program is the product under test, rebuilt
// from the signed sources by the declared build; a path that does
// not exist is exempt — the product's refusal shapes name missing
// paths on purpose.
func rowDependencies(dc docContext, label string, fields []string, row canon.Row) error {
	m := newIgnoreMatcher(dc.buildIgnore)
	seeded := map[string]bool{}
	for _, s := range row.Seeds {
		seeded[filepath.ToSlash(s.Path)] = true
	}
	for _, arg := range fields[1:] {
		rel := strings.TrimPrefix(arg, "./")
		if rel == "" || seeded[rel] {
			continue
		}
		info, err := fsStat(filepath.Join(dc.workdir, filepath.FromSlash(rel)))
		if err != nil || info.IsDir() {
			continue
		}
		if !outsideSurface(m, rel) {
			continue
		}
		return rejection(label, canondata.T("format.error.ignore-dependency", canondata.M{"arg": arg}))
	}
	return nil
}

// outsideSurface — whether the walk's accounting excludes the path:
// the file itself or any ancestor directory matches an ignore
// pattern (the snapshot walk's SkipDir semantics, ancestor-aware).
func outsideSurface(m *ignoreMatcher, rel string) bool {
	parts := strings.Split(rel, "/")
	for i := 1; i <= len(parts); i++ {
		if m.excluded(strings.Join(parts[:i], "/")) {
			return true
		}
	}
	return false
}
