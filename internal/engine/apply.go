// Applying deltas: strict decode, idempotency, validation on canon
// projections and the "journal ahead of files" transaction.
package engine

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/neurophant/punchtape/internal/brief"
	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/checks"
	"github.com/neurophant/punchtape/internal/delta"
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

// gateNames — catalog gate names (reply-to validation in a fix).
var gateNames = map[string]bool{
	"suite": true, "types": true, "lint": true, "trace": true,
	"review-diff": true, "coverage-probe": true, "latency": true, "cost": true,
}

// facetPattern — facet name: lowercase dash-joined words, machine layer.
var facetPattern = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)

// versionPattern — contract semantic version: three numbers.
var versionPattern = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

// rejectErr — a validation error: one line, entity and reason.
// System failures bypass this type — they have a different reply.
type rejectErr struct {
	entity string
	reason string
}

func (e *rejectErr) Error() string { return e.entity + ": " + e.reason }

func rejection(entity, reason string) error { return &rejectErr{entity, reason} }

func isRejection(err error) bool {
	var r *rejectErr
	return errors.As(err, &r)
}

// Submit handles a delta submission: strict decode, idempotency,
// validation on projections and the transactional "journal ahead of
// files" apply. The trailing "# TOKENS <n>" line is an OPTIONAL
// executor counter on top of the delta: it is cut off before decode
// and goes into the ledger (a fact, not an opinion); without a
// caller counter the verdict honestly says "no data" — no ceremony,
// no warning. Reply: ACCEPTED with the next slot (or the verdict —
// a merge of the read loop: the hand needs no separate next) / one
// error line plus a skeleton / FAIL.
func (e *Engine) Submit(raw []byte) (string, int) {
	return e.submitCore(raw, true)
}

// SubmitInit — the initialization submission (the wish arrives with
// the first next, the brief relay round disappears): the same
// transactional path and a slot in the reply.
func (e *Engine) SubmitInit(raw []byte) (string, int) {
	return e.submitCore(raw, true)
}

// SubmitDraft — a spec draft submission: a cheap executor on the
// machine's routing writes the first draft of the checks table; the
// same transaction is marked with origin "draft" in the journal, and
// the spec waits for the main hand's review (an edit or accept-draft)
// — a draft does not become a spec by itself.
func (e *Engine) SubmitDraft(raw []byte) (string, int) {
	e.draftSubmit = true
	defer func() { e.draftSubmit = false }()
	return e.submitCore(raw, false)
}

// SubmitAuto — the auto-mode submission (the driver): the same
// transactional path but no slot in the reply. The driver reads the
// slot itself via NextInfo and hands it to the executor in a separate
// envelope; a slot in the reply would mean bytes nobody reads — the
// ceremony counter must not see them.
func (e *Engine) SubmitAuto(raw []byte) (string, int) {
	return e.submitCore(raw, false)
}

func (e *Engine) submitCore(raw []byte, withSlot bool) (string, int) {
	started := time.Now()
	e.lastIdempotent = false
	// The write monopoly spans the whole submission: state read,
	// validation, apply, commit and ledger form one read-modify-write
	// transaction; a lock on the commit window alone left a window of
	// lost updates for two parallel submissions. Busy after short
	// retries is an honest refusal: there is no second writer.
	lock, err := e.acquireWriterRetry()
	if err != nil {
		reply := canondata.T("submit.writer-busy", canondata.M{"err": err.Error()})
		e.logVerb("submit", "failure", map[string]string{"reply-bytes": fmt.Sprintf("%d", len(reply))},
			time.Since(started).Milliseconds(), 0)
		return reply, Failure
	}
	defer e.releaseWriter(lock)
	// The live run outranks the rehearsal: the stop marker spans the
	// whole submission (this submission's gates and probes displace
	// the background suite); it is cleared before the next rehearsal
	// is spawned.
	stopRehearsal(e.Workdir)
	body, fixes := delta.RepairFormat(string(raw))
	body, tokens, _, _ := brief.SplitTokenReport(body)
	reply, code := e.submit([]byte(body), withSlot)
	// Repair notes append, never prepend: the reply's head is the
	// outcome line (a rejection's reason, an acceptance's verdict
	// path) — a note printed above it buries the cause under its own
	// repairs.
	if len(fixes) > 0 {
		reply = reply + "\n" + strings.Join(fixes, "\n")
	}
	outcome := "accepted"
	switch code {
	case Rejected:
		outcome = "rejected"
	case Failure:
		outcome = "failure"
	}
	details := map[string]string{
		"reply-bytes": fmt.Sprintf("%d", len(reply)),
		"checks":      fmt.Sprintf("%d", e.lastSubmitChecks),
		"probes":      fmt.Sprintf("%d", e.lastSubmitProbes),
		// Honest run working time: the latency line subtracts it from
		// the submission wall — the budget counts the machine's
		// overhead, not the duration of the behavior under check.
		"runs-ms": fmt.Sprintf("%d", e.lastSubmitRunsMs),
	}
	// The rejection reason is a machine fact for the ledger: the diary
	// counts rejected-delta; without a reason the entry is
	// unidentifiable once the live instance is lost (the ERROR line is
	// the reason; format auto-repair lines above the reply do not
	// overwrite it).
	if outcome == "rejected" {
		details["reason"] = firstErrorLine(reply)
	}
	// Self-described submission source (a voluntary declaration by the
	// caller, the PT_SUBMITTER environment): the machine cannot tell
	// channels apart — an honest "as reported" fact, without trust.
	if v := os.Getenv("PT_SUBMITTER"); v != "" {
		details["submitter-reported"] = v
	}
	// An idempotent key replay is a machine fact (the diary's pain:
	// new work under an old key was not applied).
	if e.lastIdempotent {
		details["idempotent"] = "yes"
	}
	if withSlot {
		// The hand waits not only for the transaction but also for the
		// nested slot render — this submission's latency budget
		// includes the slot term.
		details["slot"] = "yes"
	}
	e.logVerb("submit", outcome, details, time.Since(started).Milliseconds(), tokens)
	// The verdict frozen while building the reply (the slot tail
	// renders the verdict BEFORE this submission's counter booking) is
	// re-rendered after logVerb: bill must see the tail of the latest
	// submission; the slot cache is updated in the same pass — the
	// nested verdict matches bit for bit.
	e.refreezeDeliverVerdict()
	// Background rehearsal of the full suite: after accepted
	// code/fix the machine, in a separate process, runs all instance
	// checks against the surface digest while the hand prepares the
	// next submission; the report lands in the reply of a fitting
	// submission. Nothing before surface code exists. The start is
	// AFTER the ledger write and the re-freeze: the background process
	// holds the writer lock for the duration of its run; starting
	// before the write would keep this submission's wall waiting on
	// the lock (measured: 8 s on a 4-second suite).
	if code == Accepted && (e.lastDeltaKind == deltaKindCode || e.lastDeltaKind == deltaKindFix) {
		resumeRehearsal(e.Workdir)
		SpawnBackgroundSuite(e.Workdir)
	}
	return reply, code
}

// refreezeDeliverVerdict — re-render of the delivery verdict and the
// slot cache after this submission's counter is written to the ledger.
// The "reply first, ledger second" order ages the frozen bill by
// exactly one tail; the freeze key is the journal head and the ledger
// does not break it, so without a re-render the stale bill would live
// forever.
func (e *Engine) refreezeDeliverVerdict() {
	st, err := e.Store.State()
	if err != nil || st.Stage != canon.StageDeliver {
		return
	}
	v, err := e.renderVerdictFresh()
	if err != nil {
		return
	}
	if e.lastVerdictDrift {
		return
	}
	e.freezeVerdict(v)
	if text, _, err := e.renderNext(st, ""); err == nil {
		e.storeCachedNext(text)
	}
}

// nextSlotTail — the next slot (on deliver, the verdict) for the
// submit reply. The render is the same as next's, and so is the cache:
// a subsequent next matches the nested slot bit for bit and is free.
// The reply serves a contextual diff: the hand has just seen the
// previous slot, and blocks unchanged since then are anchored by one
// line with the digest (46% of CONTEXT bytes were resent
// unchanged); the full render stays in the slot cache and the
// shown-blocks map, the full print is next --full. A render failure
// does not bring down an accepted transaction: the reply honestly
// loses the slot, next keeps working.
func (e *Engine) nextSlotTail() string {
	state, err := e.Store.State()
	if err != nil {
		return ""
	}
	text, _, err := e.renderNext(state, "")
	if err != nil {
		return ""
	}
	e.storeCachedNext(text)
	compacted, _, _, _ := compactSlot(text, loadShownBlocks(e))
	storeShownBlocks(e, text)
	return canondata.T("submit.next-slot") + "\n" + compacted
}

// draftPending — a spec draft awaits the hand's review: the last
// table submission in the journal has origin "draft", and the hand has
// not yet submitted its own table on top. A DRAFT DOES NOT BECOME A
// SPEC BY ITSELF: while there is no review, the spec stage holds.
func (e *Engine) draftPending() bool {
	pending := false
	for _, en := range e.Journal.All() {
		if en.DeltaKind == deltaKindChecks {
			pending = en.Origin == "draft"
		}
	}
	if e.draftSubmit {
		pending = true
	}
	return pending
}

func (e *Engine) submit(raw []byte, withSlot bool) (string, int) {
	started := time.Now()
	e.lastSubmitChecks = 0
	e.lastSubmitProbes = 0
	e.lastSubmitRunsMs = 0

	var d delta.Delta
	if err := yamlio.DecodeStrict(raw, &d); err != nil {
		// The rejection carries a deterministic hint with the corrected
		// line when the correction is unique; semantics are not guessed.
		// Plus the dialect's root cause in one line when the pattern
		// is known (an anchor, a comma in a flow value, etc.).
		kind := d.Kind
		reason := firstLine(err)
		if diag, ok := delta.DiagnoseYAML(string(raw), err); ok {
			reason += canondata.T("submit.diag-eg", canondata.M{"hint": diag})
		}
		if hint, ok := delta.SuggestLine(string(raw), err, kind); ok {
			reason += canondata.T("submit.hint-eg", canondata.M{"hint": hint})
		}
		return reject("delta", reason, kind), Rejected
	}
	e.lastDeltaKind = d.Kind
	// The submission key is optional: the machine derives its own from
	// the content digest — the key-numbering ritual is lifted from the
	// hand, and the machine itself admits no content collisions. An
	// amend decision without a key is named by the open hold:
	// identical replies to different holds are different facts (a
	// false replay of the second reject used to lock the lifting), an
	// identical reply to the same hold is an honest replay.
	cd := contentDigestOf(d)
	if d.SubmissionKey == "" {
		d.SubmissionKey = e.autoSubmissionKey(cd, d.Kind)
	}
	e.lastContentDigest = cd

	// Idempotency: the same key — the same transaction, nothing is
	// applied. The slot in the reply is the current one (after a lost
	// reply it is exactly what the hand needs), while the transaction
	// is not repeated. There used to be a stage gate: a replay does
	// not depend on the current stage.
	if entry, ok := e.Journal.FindByKey(d.SubmissionKey); ok {
		e.lastIdempotent = true
		tail := ""
		if withSlot {
			tail = e.nextSlotTail()
		}
		// Healing phrase: a predictable agent failure — a key replay —
		// the machine answers normally from data, the hand does not
		// spin. The header is IDEMPOTENT REPLAY, not ACCEPTED: the
		// replay applied nothing, and an honest header does not claim
		// an apply.
		advice := canondata.Heal("idempotent-resubmit")
		if advice != "" {
			tail = "\n" + advice + "\n" + tail
		}
		return canondata.T("submit.idempotent", canondata.M{
			"tx":   shortHex(entry.Transaction),
			"note": e.idempotentConventionsNote(d.Kind),
			"tail": tail,
		}), Accepted
	}
	// Idempotency by content: the same delta under a fresh key is the
	// same fact; a replay does not inflate the submission and
	// fix-cycle counters. Amend is excluded: its content is only the
	// decision (approve/reject), the effect is determined by the
	// server-side hold — identical decisions of different holds are
	// not different facts.
	if d.Kind != delta.KindAmend {
		if entry, ok := e.Journal.FindByContentDigest(cd); ok {
			e.lastIdempotent = true
			tail := ""
			if withSlot {
				tail = e.nextSlotTail()
			}
			return canondata.T("submit.idempotent-content", canondata.M{
				"tx":   shortHex(entry.Transaction),
				"key":  entry.SubmissionKey,
				"note": e.idempotentConventionsNote(d.Kind),
				"tail": tail,
			}), Accepted
		}
	}

	state, err := e.Store.State()
	if err != nil {
		return failure(err), Failure
	}
	if !delta.KindAllowedOnStage(d.Kind, state.Stage) {
		return reject("delta", stageRejectReason(d.Kind, state.Stage), d.Kind), Rejected
	}

	tx, replyTail, filesNote, rerun, err := e.buildAndApply(d, raw, state, started)
	if err != nil {
		if isRejection(err) {
			r := err.(*rejectErr)
			return reject(r.entity, r.reason, d.Kind), Rejected
		}
		return failure(err), Failure
	}
	tail := ""
	if withSlot {
		tail = e.nextSlotTail()
	}
	files := ""
	if filesNote != "" {
		files = "\n" + filesNote
	}
	if rerun {
		return canondata.T("submit.rerun", canondata.M{
			"gates": replyTail, "tail": tail,
		}), Accepted
	}
	if replyTail != "" {
		return canondata.T("submit.accepted.with-gates", canondata.M{
			"tx":    shortHex(tx),
			"gates": replyTail,
			"files": files,
			"tail":  tail,
		}), Accepted
	}
	return canondata.T("submit.accepted", canondata.M{
		"tx":    shortHex(tx),
		"files": files,
		"tail":  tail,
	}), Accepted
}

// NoteExecutorCall records the scheduler invoking an executor: the
// route (command), the caller's shape, the slot kind and the card
// facet — routing facts; tokens come from the caller's counter if the
// executor reported them in its reply (zero — did not report). Tokens
// are added to the task cost in the cost gate.
func (e *Engine) NoteExecutorCall(kind, facet, route, shape string, tokens int64) error {
	state, err := e.Store.State()
	if err != nil {
		return err
	}
	le := ledger.NewEntry("executor", state.Stage)
	le.Details = map[string]string{"kind": kind, "facet": facet, "route": route, "shape": shape}
	if tokens > 0 {
		le.Tokens = &tokens
	}
	return e.Ledger.Append(le)
}

// contentDigestOf — the content digest of a delta: submission keys are
// stripped in both layers (numbering is not content), the token
// trailer is cut off before decode. One content under different keys
// is one fact.
func contentDigestOf(d delta.Delta) string {
	data, err := yamlio.Marshal(d.WithoutKeys())
	if err != nil {
		return ""
	}
	return yamlio.Digest(data)
}

// idempotentConventionsNote — a hint on a conventions recipe replay:
// the replay applies nothing; replacing the recipe is OTHER content
// (the section is replaced wholesale by the last submission). The
// current recipe is included as data, so the hand sees what exactly
// lives there and does not change cosmetics blindly.
func (e *Engine) idempotentConventionsNote(kind string) string {
	if kind != delta.KindConventions {
		return ""
	}
	conv, err := e.Store.Conventions()
	if err != nil || conv == nil {
		return ""
	}
	data, err := yamlio.Marshal(conv)
	if err != nil {
		return ""
	}
	return canondata.T("submit.idempotent-recipe", canondata.M{
		"recipe": strings.TrimRight(string(data), "\n"),
	})
}

// autoSubmissionKey — the machine key of an anonymous submission:
// content, and for amend also the hold the submission answers (key
// and staging position are the hold's identity). A submission is a
// function application (content × state): the same reply to the same
// hold — the same submission (an honest replay after a lost reply);
// the same reply to a different hold (including re-staging the same
// content) — a new submission. State is read under the writer
// monopoly; there is no race.
func (e *Engine) autoSubmissionKey(contentDigest, kind string) string {
	if kind == delta.KindAmend {
		if st, err := e.Store.State(); err == nil && st.PendingAmend != nil {
			hold := st.PendingAmend
			scoped := contentDigest + "\x00" + hold.Key + "\x00" + hold.Staging
			return "auto-" + shortHex(yamlio.Digest([]byte(scoped)))
		}
	}
	return "auto-" + shortHex(contentDigest)
}

// Try — a dry probe of ONE row of the checks table: the same runner,
// clean directory, double run and probe — WITHOUT writing to the
// canon, journal or ledger and without moving the stage or counters.
// Try the shape before it enters the canon: the reply is the outcome
// (green / red with a reason). The row runs against the current
// project state (files per the journal) and the declared surface.
func (e *Engine) Try(raw []byte) (string, int) {
	started := time.Now()
	var d delta.Delta
	if err := yamlio.DecodeStrict(raw, &d); err != nil {
		return reject("delta", firstLine(err), d.Kind), Rejected
	}
	if d.Kind != delta.KindChecks || d.Checks == nil {
		return reject("delta", canondata.T("submit.reject.try-not-checks"), d.Kind), Rejected
	}
	if line, trap := delta.SeedQuoteEscape(string(raw)); trap {
		return reject("seeds", canondata.T("checktable.error.seed-quote-escape",
			canondata.M{"line": fmt.Sprintf("%d", line)}), d.Kind), Rejected
	}
	if line, trap := delta.ExpectationQuoteEscape(string(raw)); trap {
		return reject("expectations", canondata.T("checktable.error.expectation-quote-escape",
			canondata.M{"line": fmt.Sprintf("%d", line)}), d.Kind), Rejected
	}
	ck := d.Checks
	if len(ck.Rows) != 1 {
		return reject("delta", canondata.T("submit.reject.try-one-row", canondata.M{
			"got": fmt.Sprintf("%d", len(ck.Rows)),
		}), d.Kind), Rejected
	}
	st, err := e.Store.State()
	if err != nil {
		return canondata.T("cli.fail", canondata.M{"err": err.Error()}), Failure
	}
	b := newBuilder(e, st)
	row := ck.Rows[0]
	seed, materials, pre, when, err := b.rowIdentity(row)
	if err != nil {
		return reject("delta", err.Error(), d.Kind), Rejected
	}
	then, _, err := rowExpectations(row)
	if err != nil {
		if r, ok := err.(*rejectErr); ok {
			return reject(r.entity, r.reason, d.Kind), Rejected
		}
		return reject("delta", firstLine(err), d.Kind), Rejected
	}
	names := map[string]bool{}
	for _, cmd := range row.Run {
		if fields, ferr := cmd.Fields(); ferr == nil && len(fields) > 0 {
			names[fields[0]] = true
		}
	}
	check := canon.Check{
		ID: "TRY", Seed: seed, Materials: materials, Pre: pre,
		When:         &canon.When{Surface: "cli", Command: when, TimeoutSec: rowTimeoutOr(row.TimeoutSec), Volatile: row.Volatile},
		SurfaceNames: sortedKeys(names),
		Then:         then,
	}
	res := e.runCheck(check, checks.Normal)
	outcome := "red"
	if res.Green {
		outcome = "green"
	}
	e.logVerb("try", outcome, map[string]string{}, time.Since(started).Milliseconds(), 0)
	if res.Green {
		return canondata.T("submit.try.green"), Accepted
	}
	return canondata.T("submit.try.red", canondata.M{"reason": res.Reason}), Rejected
}

// Check — the preflight validation of a submission: the same
// validation as submit (whitelist format repair, strict decode, stage
// gate, validation on canon projections) but without applying — no
// journal, no files, no gates, no ledger. The hand checks a delta
// before the real submission; an idempotent replay is recognized and
// reported.
func (e *Engine) Check(raw []byte) (string, int) {
	body, fixes := delta.RepairFormat(string(raw))
	body, _, _, _ = brief.SplitTokenReport(body)
	raw = []byte(body)

	var d delta.Delta
	if err := yamlio.DecodeStrict(raw, &d); err != nil {
		kind := d.Kind
		reason := firstLine(err)
		if diag, ok := delta.DiagnoseYAML(string(raw), err); ok {
			reason += canondata.T("submit.diag-eg", canondata.M{"hint": diag})
		}
		if hint, ok := delta.SuggestLine(string(raw), err, kind); ok {
			reason += canondata.T("submit.hint-eg", canondata.M{"hint": hint})
		}
		return reject("delta", reason, kind), Rejected
	}
	cd := contentDigestOf(d)
	if d.SubmissionKey == "" {
		d.SubmissionKey = e.autoSubmissionKey(cd, d.Kind)
	}
	if _, ok := e.Journal.FindByKey(d.SubmissionKey); ok {
		return canondata.T("submit.check.journaled", canondata.M{
			"kind": d.Kind, "key": d.SubmissionKey,
		}), Accepted
	}
	if d.Kind != delta.KindAmend {
		if _, ok := e.Journal.FindByContentDigest(cd); ok {
			return canondata.T("submit.check.journaled-content", canondata.M{
				"kind": d.Kind, "digest": shortHex(cd),
			}), Accepted
		}
	}
	state, err := e.Store.State()
	if err != nil {
		return failure(err), Failure
	}
	if !delta.KindAllowedOnStage(d.Kind, state.Stage) {
		return reject("delta", stageRejectReason(d.Kind, state.Stage), d.Kind), Rejected
	}
	b := newBuilder(e, state)
	b.rawInput = string(raw)
	if err := b.applyDelta(d); err != nil {
		if isRejection(err) {
			r := err.(*rejectErr)
			return reject(r.entity, r.reason, d.Kind), Rejected
		}
		return failure(err), Failure
	}
	head := canondata.T("submit.check.ok", canondata.M{"kind": d.Kind, "key": d.SubmissionKey})
	if len(fixes) > 0 {
		head = head + "\n" + strings.Join(fixes, "\n")
	}
	return head, Accepted
}

// buildAndApply validates the delta on canon projections, advances
// the stages and records the transaction in the journal with full
// effects. replyTail — the one-line gates summary for the reply to
// the agent, empty when no gates ran; filesNote — the block of
// digests of written files (code/fix); rerun — a submission without
// a single new byte: nothing applied, gates re-run synchronously.
func (e *Engine) buildAndApply(d delta.Delta, raw []byte, state canon.State, started time.Time) (string, string, string, bool, error) {
	// The stage is recorded before advancing: a code submission is a
	// fact of the implement stage, even if the machine immediately ran
	// the cycle to deliver.
	submitStage := state.Stage

	b := newBuilder(e, state)
	b.rawInput = string(raw)
	if err := b.applyDelta(d); err != nil {
		return "", "", "", false, err
	}
	// A pure re-run (not a single new byte; only code/fix get here —
	// only they write files): no submission or effects transaction,
	// the submission ledger stays silent, the attempt budget does not
	// move (chargeReds nil); gates run synchronously from the same
	// hand — disguising a re-run as a fix is no longer needed. The
	// submission does not enter the journal: a re-run is a repeatable
	// action, not a new fact.
	if b.pureRerun {
		tx := yamlio.Digest(raw)
		summary, err := e.runAfterSubmit(d.Kind, b.gateCard, tx, b.freshChecks, nil)
		if err != nil {
			return "", "", "", true, fmt.Errorf("gates: %w", err)
		}
		return tx, summary, "", true, nil
	}
	// The passport boundary registry is synchronized with the spec on
	// every transaction: a new requirement is a boundary claim, a
	// vanished one goes away, a changed spec is a repeated claim.
	if err := b.passportSync(); err != nil {
		return "", "", "", false, err
	}
	b.advance()
	if b.processErr != nil {
		return "", "", "", false, b.processErr
	}

	tx := yamlio.Digest(raw)
	entry := journal.Entry{
		Transaction:   tx,
		SubmissionKey: d.SubmissionKey,
		DeltaKind:     d.Kind,
		WallMs:        time.Since(started).Milliseconds(),
		RecordedAt:    time.Now().UTC(),
		Effects:       b.effects,
		ContentDigest: e.lastContentDigest,
	}
	if e.draftSubmit {
		entry.Origin = "draft"
	}
	if b.assertedAmend {
		entry.Origin = "asserted"
	}
	if b.stagedAmend {
		entry.Held = true
	}
	if err := e.commit(entry); err != nil {
		return "", "", "", false, fmt.Errorf("transaction: %w", err)
	}

	// The ledger holds machine facts; by this point the canon is
	// consistent.
	le := ledger.NewEntry("apply", submitStage)
	le.WallMs = entry.WallMs
	le.Details = map[string]string{"kind": d.Kind, "outcome": "accepted"}
	if err := e.Ledger.Append(le); err != nil {
		return "", "", "", false, fmt.Errorf("ledger: %w", err)
	}

	// A deferred edit of expectations is a human attention point: the
	// question is opened by a submission and closed by an explicit
	// amend or by silence (the default).
	if b.stagedAmend {
		qe := ledger.NewEntry("question", submitStage)
		qe.Details = map[string]string{"questions": "1", "kind": "amend-assertion"}
		if err := e.Ledger.Append(qe); err != nil {
			return "", "", "", false, fmt.Errorf("ledger: %w", err)
		}
	}
	if d.Kind == delta.KindAmend {
		ce := ledger.NewEntry("choice", submitStage)
		ce.Details = map[string]string{
			"questions": "1", "answered": "1", "defaults": "0",
			"kind": "amend-assertion",
		}
		if err := e.Ledger.Append(ce); err != nil {
			return "", "", "", false, fmt.Errorf("ledger: %w", err)
		}
	}

	// Human attention: the question batch and the answer to it are
	// the points where the cycle stands on a person. The wall between
	// them is a machine fact of attention. Questions and answers ride
	// with the checks table — the events are the same.
	if d.Kind == delta.KindChecks {
		if b.openedQuestions > 0 {
			qe := ledger.NewEntry("question", submitStage)
			qe.Details = map[string]string{"questions": fmt.Sprintf("%d", b.openedQuestions)}
			if err := e.Ledger.Append(qe); err != nil {
				return "", "", "", false, fmt.Errorf("ledger: %w", err)
			}
		}
		if b.resolvedBatch {
			ce := ledger.NewEntry("choice", submitStage)
			ce.Details = map[string]string{
				"questions": fmt.Sprintf("%d", b.batchTotal),
				"answered":  fmt.Sprintf("%d", b.batchAnswered),
				"defaults":  fmt.Sprintf("%d", b.batchDefaults),
			}
			if err := e.Ledger.Append(ce); err != nil {
				return "", "", "", false, fmt.Errorf("ledger: %w", err)
			}
		}
		// Explicit answers to a closed batch are an attention-accounting
		// correction: silence defaults overridden by a person's word
		// move from the defaults counter to the explicit-answer
		// counter.
		if b.answerUpgraded > 0 {
			ue := ledger.NewEntry("answer-upgrade", submitStage)
			ue.Details = map[string]string{"answers": fmt.Sprintf("%d", b.answerUpgraded)}
			if err := e.Ledger.Append(ue); err != nil {
				return "", "", "", false, fmt.Errorf("ledger: %w", err)
			}
		}
	}
	if d.Kind == delta.KindLint && d.Lint != nil && len(d.Lint.Findings) > 0 {
		qe := ledger.NewEntry("question", submitStage)
		qe.Details = map[string]string{"questions": fmt.Sprintf("%d", len(d.Lint.Findings))}
		if err := e.Ledger.Append(qe); err != nil {
			return "", "", "", false, fmt.Errorf("ledger: %w", err)
		}
	}
	if d.Kind == delta.KindClarify {
		if err := e.logChoice(d.Clarify); err != nil {
			return "", "", "", false, err
		}
	}
	// A behavior probe is a machine observation fact: how many
	// commands this delta submitted and what the state holds after it
	// (determinism — by the two runs of each command).
	if d.Kind == delta.KindProbe && d.Probe != nil {
		st := mustState(e)
		det, failed := 0, 0
		for _, p := range st.Probes {
			switch {
			case p.Failed:
				failed++
			case p.Deterministic:
				det++
			}
		}
		pe := ledger.NewEntry("probe", submitStage)
		pe.Details = map[string]string{
			"commands":      fmt.Sprintf("%d", len(d.Probe.Commands)),
			"observations":  fmt.Sprintf("%d", len(st.Probes)),
			"deterministic": fmt.Sprintf("%d", det),
			"failed":        fmt.Sprintf("%d", failed),
		}
		if err := e.Ledger.Append(pe); err != nil {
			return "", "", "", false, fmt.Errorf("ledger: %w", err)
		}
	}

	// Gates are the machine's post-apply run transaction: the agent
	// knows them only as a list in the slot; the summary is one line
	// in the reply.
	summary, err := e.runAfterSubmit(d.Kind, b.gateCard, tx, b.freshChecks, b.chargeReds)
	if err != nil {
		return "", "", "", false, fmt.Errorf("gates: %w", err)
	}
	// Auto re-verification of an unchanged rollback: the cycle
	// returned to acceptance, the surface and the check set are
	// untouched — the machine runs the gates, resending byte-for-byte
	// the same code is not needed (the gates already ran above — a
	// replay of the reds or fresh rows; an empty summary = did not
	// run).
	if b.reopenedCycle && summary == "" {
		extra, aerr := e.reopenAutorun(tx)
		if aerr != nil {
			return "", "", "", false, fmt.Errorf("gates: %w", aerr)
		}
		summary = extra
	}
	// Notes of answers to a closed batch go on top of the gates
	// summary: an explicit answer is named in the submission reply,
	// there is no silent loss of human input. The IDs of assigned
	// curator entries come as the first layer: addressability from the
	// moment of creation.
	notes := strings.Join(b.knowledgeNotes, "\n")
	if rows := strings.Join(b.rowNotes, "\n"); rows != "" {
		if notes != "" {
			notes += "\n"
		}
		notes += rows
	}
	if ans := strings.Join(b.answerNotes, "\n"); ans != "" {
		if notes != "" {
			notes += "\n"
		}
		notes += ans
	}
	// Digests of written files go to the hand: referencing a digest in
	// the next submission replaces resending an unchanged file.
	filesNote := ""
	if len(b.writtenFiles) > 0 {
		var lines []string
		lines = append(lines, canondata.T("submit.files.header", canondata.M{
			"count": fmt.Sprintf("%d", len(b.writtenFiles)),
		}))
		for _, f := range b.writtenFiles {
			lines = append(lines, canondata.T("submit.files.line", canondata.M{
				"path": f.path, "digest": f.digest,
			}))
		}
		filesNote = strings.Join(lines, "\n")
	}
	if summary != "" {
		if notes != "" {
			notes += "\n"
		}
		return tx, notes + canondata.T("submit.reply.gates-summary", canondata.M{"summary": summary}), filesNote, false, nil
	}
	return tx, notes, filesNote, false, nil
}

// logChoice records the answer to a clarification batch: how many
// answered, how many taken by default, and how much human attention
// the batch ate — the wall from the question event to the answer,
// measured by the machine.
func (e *Engine) logChoice(cl *delta.Clarify) error {
	questions := 0
	if st, err := e.Store.State(); err == nil && st.Lint != nil {
		questions = len(st.Lint.Findings)
	}
	answered := 0
	if cl != nil {
		answered = len(cl.Answers)
	}
	attentionMs := int64(0)
	var lastQuestion ledger.Entry
	found := false
	for _, en := range e.Ledger.All() {
		if en.Event == "question" {
			lastQuestion, found = en, true
		}
	}
	if found {
		attentionMs = time.Since(lastQuestion.At).Milliseconds()
	}
	ce := ledger.NewEntry("choice", canon.StageIntake)
	ce.WallMs = 0
	ce.Details = map[string]string{
		"questions":    fmt.Sprintf("%d", questions),
		"answered":     fmt.Sprintf("%d", answered),
		"defaults":     fmt.Sprintf("%d", questions-answered),
		"attention-ms": fmt.Sprintf("%d", attentionMs),
	}
	return e.Ledger.Append(ce)
}

// --- transaction builder ---

// builder holds canon projections and accumulates file effects. A
// validation error aborts the build entirely: partial applies do not
// exist.
type builder struct {
	e          *Engine
	state      canon.State
	reqs       map[string]canon.Requirement
	scns       map[string]canon.Scenario
	cards      map[string]canon.Card
	contracts  map[string]canon.Contract
	checks     map[string]canon.Check
	boundaries map[string]canon.Boundary
	decisions  []canon.Decision
	outofscope []canon.OutOfScopeItem
	kb         []canon.KbEntry
	effects    []journal.Effect
	touched    []string
	loaded     bool
	// fresh/changed checks of this submission: only the touched is rechecked
	freshChecks []string
	// the card whose gates to run after commit (code/fix)
	gateCard string
	// behaviorally red checks of the card BEFORE the code/fix
	// submission: the budget attempt is incremented by the gates after
	// the run and only if some of them stayed red — a barren repair
	// cycle
	chargeReds map[string]bool
	// files written by a code/fix submission with canonical digests:
	// the reply carries them to the hand — the next code/fix
	// references the digest instead of resending an unchanged file
	writtenFiles []fileWritten
	// a code/fix submission carries not a single new byte (all writes
	// match the project, all deletions are already gone) — a pure
	// re-run of the gates without applying effects
	pureRerun bool
	// the clarification cycle of this submission: how many questions
	// asked and whether the open batch is resolved (by explicit
	// answers or defaults) — facts for the attention ledger; the
	// closure counters count the batch being resolved, even if this
	// same submission opens the next one
	openedQuestions int
	resolvedBatch   bool
	batchTotal      int
	batchAnswered   int
	batchDefaults   int
	// answers to a closed batch: silence defaults promoted to the
	// person's explicit words (record rewrite + attention event), and
	// answers to already-explicit choices (a notice, no rewrite)
	answerNotes    []string
	answerUpgraded int
	// assigned IDs of this submission's curator entries: published in
	// the reply — a record is addressed by ID at once, without
	// guessing the numbering scheme (kb — the content digest)
	knowledgeNotes []string
	// notes about the shape of submitted rows (not rejections): an
	// add-row that looks like a replacement — the machine names it
	// explicitly, there are no silent duplicates
	rowNotes []string
	// whether the checks table delta was applied in THIS transaction:
	// a spec draft review releases the stage hold
	appliedChecks bool
	// the raw delta of this submission verbatim: material for the
	// deferred edit (pending-amend stores the submission whole — all
	// or nothing)
	rawInput string
	// deltaKey — the effective submission key (explicit or the
	// machine auto-key): a hold must carry a non-empty key; the open
	// edit's rejection and the slot name the edit by it
	deltaKey string
	// stagingTx — the submission's position in the journal: hold
	// identity for anonymous amend answers
	stagingTx string
	// assertedAmend — a person-asserted edit is being applied: the
	// deferral trigger does not fire again
	assertedAmend bool
	// stagedAmend — the submission became a deferred edit (a question
	// to the person)
	stagedAmend bool
	// processErr — a process data error (stage cycle, unknown
	// mechanics name): an honest failure that aborts the submission
	// instead of a panic or a spin
	processErr error
	// reopenedCycle — the submission rolled the cycle back from
	// deliver to acceptance: a candidate for auto re-verification
	// without resending code (if the surface and the check set are
	// unchanged since the acceptance moment)
	reopenedCycle bool
}

func newBuilder(e *Engine, state canon.State) *builder {
	return &builder{
		e: e, state: state,
		reqs:       map[string]canon.Requirement{},
		scns:       map[string]canon.Scenario{},
		cards:      map[string]canon.Card{},
		contracts:  map[string]canon.Contract{},
		checks:     map[string]canon.Check{},
		boundaries: map[string]canon.Boundary{},
	}
}

// load reads the projections from disk once per transaction.
func (b *builder) load() error {
	if b.loaded {
		return nil
	}
	reqs, err := b.e.Store.Requirements()
	if err != nil {
		return fmt.Errorf("system failure: %w", err)
	}
	for _, r := range reqs {
		b.reqs[r.ID] = r
	}
	scns, err := b.e.Store.Scenarios()
	if err != nil {
		return fmt.Errorf("system failure: %w", err)
	}
	for _, s := range scns {
		b.scns[s.ID] = s
	}
	cards, err := b.e.Store.Cards()
	if err != nil {
		return fmt.Errorf("system failure: %w", err)
	}
	for _, c := range cards {
		b.cards[c.ID] = c
	}
	contracts, err := b.e.Store.Contracts()
	if err != nil {
		return fmt.Errorf("system failure: %w", err)
	}
	for _, c := range contracts {
		b.contracts[c.ID] = c
	}
	checks, err := b.e.Store.Checks()
	if err != nil {
		return fmt.Errorf("system failure: %w", err)
	}
	for _, c := range checks {
		b.checks[c.ID] = backfillCheckProvenOne(c)
	}
	boundaries, err := b.e.Store.Boundaries()
	if err != nil {
		return fmt.Errorf("system failure: %w", err)
	}
	for _, bd := range boundaries {
		b.boundaries[bd.ID] = bd
	}
	if b.decisions, err = b.e.Store.Decisions(); err != nil {
		return fmt.Errorf("system failure: %w", err)
	}
	if b.outofscope, err = b.e.Store.OutOfScope(); err != nil {
		return fmt.Errorf("system failure: %w", err)
	}
	if b.kb, err = b.e.Store.KbEntries(); err != nil {
		return fmt.Errorf("system failure: %w", err)
	}
	b.loaded = true
	return nil
}

func (b *builder) putReq(r canon.Requirement) error {
	data, err := yamlio.Marshal(r)
	if err != nil {
		return fmt.Errorf("system failure: %w", err)
	}
	b.reqs[r.ID] = r
	b.effects = append(b.effects, journal.Effect{
		Path: canonFile("requirements", r.ID), Content: machineFile(data),
	})
	b.touch(r.ID)
	return nil
}

func (b *builder) putScn(sc canon.Scenario) error {
	data, err := yamlio.Marshal(sc)
	if err != nil {
		return fmt.Errorf("system failure: %w", err)
	}
	b.scns[sc.ID] = sc
	b.effects = append(b.effects, journal.Effect{
		Path: canonFile("scenarios", sc.ID), Content: machineFile(data),
	})
	b.touch(sc.ID)
	return nil
}

func (b *builder) putCard(c canon.Card) error {
	data, err := yamlio.Marshal(c)
	if err != nil {
		return fmt.Errorf("system failure: %w", err)
	}
	b.cards[c.ID] = c
	b.effects = append(b.effects, journal.Effect{
		Path: canonFile("cards", c.ID), Content: machineFile(data),
	})
	b.touch(c.ID)
	return nil
}

func (b *builder) putContract(c canon.Contract) error {
	data, err := yamlio.Marshal(c)
	if err != nil {
		return fmt.Errorf("system failure: %w", err)
	}
	b.contracts[c.ID] = c
	b.effects = append(b.effects, journal.Effect{
		Path: canonFile("contracts", c.ID), Content: machineFile(data),
	})
	b.touch(c.ID)
	return nil
}

func (b *builder) touch(id string) {
	for _, t := range b.touched {
		if t == id {
			return
		}
	}
	b.touched = append(b.touched, id)
}

// --- applying delta kinds ---

func (b *builder) applyDelta(d delta.Delta) error {
	if d.Kind == delta.KindChecks {
		b.appliedChecks = true
	}
	if err := b.load(); err != nil {
		return err
	}
	if d.SubmissionKey != "" {
		b.deltaKey = d.SubmissionKey
	}
	// Hold staging identity: submission content + position in the
	// journal (journal length at submission time) — identical deltas
	// staged by different transactions get different identities:
	// re-staging the same content is a new hold, not a replay of the
	// old one; the position is unique to every applied submission
	// (append-only journal, one writer).
	b.stagingTx = fmt.Sprintf("%d", len(b.e.Journal.All()))
	// An open deferred edit and a new submission: silent displacement
	// used to lose prepared work (the field is one slot). While the
	// edit is open, substantive submissions are rejected with the
	// edit's name and a legal route — assertion first (kind: amend
	// approve | reject), row merging goes as one delta after the
	// decision. The conflict escalation (lint) is an exception: a
	// contradiction must be accepted at any moment, its channel does
	// not cross the edit's.
	if b.state.PendingAmend != nil && d.Kind != delta.KindAmend &&
		d.Kind != delta.KindLint && !b.assertedAmend {
		return rejection("delta", canondata.T("submit.reject.amend-open", canondata.M{
			"key":   b.state.PendingAmend.Key,
			"count": fmt.Sprintf("%d", len(b.state.PendingAmend.Changes)),
		}))
	}
	switch d.Kind {
	case delta.KindIntent:
		return b.applyIntent(d.Intent)
	case delta.KindLint:
		return b.applyLint(d.Lint)
	case delta.KindClarify:
		return b.applyClarify(d.Clarify)
	case delta.KindSpec:
		if err := b.applySpec(d.Spec); err != nil {
			return err
		}
		return b.compileChecks()
	case delta.KindChecks:
		if err := b.applyChecks(d.Checks); err != nil {
			return err
		}
		return b.compileChecks()
	case delta.KindAssert:
		if err := b.applyAssert(d.Assert); err != nil {
			return err
		}
		return b.compileChecks()
	case delta.KindSplit:
		return b.applySplit(d.Split)
	case delta.KindCode:
		return b.applyCode(d.Code)
	case delta.KindFix:
		return b.applyFix(d.Fix)
	case delta.KindProbe:
		return b.applyProbe(d.Probe)
	case delta.KindConventions:
		return b.applyConventions(d.Conventions)
	case delta.KindAmend:
		return b.applyAmend(d.Amend)
	case delta.KindWaive:
		return b.applyWaive(d.Waive)
	case delta.KindFeature:
		return b.applyFeature(d.Feature)
	case delta.KindDecision:
		return b.applyDecision(d.Decision)
	case delta.KindKb:
		return b.applyKb(d.Kb)
	default:
		if hint, ok := delta.SuggestKind(d.Kind); ok {
			return rejection("delta", canondata.T("submit.reject.unknown-kind-hint", canondata.M{
				"kind": fmt.Sprintf("%q", d.Kind), "hint": hint,
			}))
		}
		return rejection("delta", canondata.T("submit.reject.unknown-kind", canondata.M{
			"kind": fmt.Sprintf("%q", d.Kind),
		}))
	}
}

// applyProbe runs the machine observation of behavior: each command
// is run twice through the surface in a clean seeded directory;
// determinism is a comparison of the two runs, not a property. The
// observations accumulate into state — the next slot assembles the
// machine's spec proposal from them (backfill: the spec catches up
// with stabilized code).
func (b *builder) applyProbe(pb *delta.Probe) error {
	if pb == nil || len(pb.Commands) == 0 {
		return rejection("probe", canondata.T("submit.reject.probe-commands-empty"))
	}
	for i, command := range pb.Commands {
		if len(command) == 0 {
			return rejection("probe", canondata.T("submit.reject.probe-command-empty", canondata.M{
				"index": fmt.Sprintf("%d", i+1),
			}))
		}
		b.e.lastSubmitProbes++
		probeStarted := time.Now()
		first, err := checks.Capture(b.e.Workdir, b.e.surfaceArtifact(), pb.Seed, command, canondata.Limit("probe.capture-timeout-s"))
		if err != nil {
			b.e.lastSubmitRunsMs += time.Since(probeStarted).Milliseconds()
			b.state.Probes = appendProbe(b.state.Probes, canon.ProbeObservation{
				Command: command, Failed: true, Reason: firstLine(err),
			})
			continue
		}
		// The second run comes after a pause of over a second:
		// second-granularity outputs (clocks, timestamps) must not
		// pass as deterministic because of a lucky coincidence of
		// seconds.
		time.Sleep(1100 * time.Millisecond)
		second, err := checks.Capture(b.e.Workdir, b.e.surfaceArtifact(), pb.Seed, command, canondata.Limit("probe.capture-timeout-s"))
		b.e.lastSubmitRunsMs += time.Since(probeStarted).Milliseconds()
		if err != nil {
			b.state.Probes = appendProbe(b.state.Probes, canon.ProbeObservation{
				Command: command, Failed: true, Reason: firstLine(err),
			})
			continue
		}
		obs := canon.ProbeObservation{
			Command:       command,
			Seed:          pb.Seed,
			ExitCode:      first.ExitCode,
			Stdout:        first.Stdout,
			Stderr:        first.Stderr,
			Deterministic: true,
		}
		if reason := capturesDiffer(first, second); reason != "" {
			obs.Deterministic = false
			obs.Reason = reason
		}
		for path, content := range first.Files {
			obs.Files = append(obs.Files, canon.ProbeFileFact{Path: path, Content: content})
		}
		sort.Slice(obs.Files, func(x, y int) bool { return obs.Files[x].Path < obs.Files[y].Path })
		b.state.Probes = appendProbe(b.state.Probes, obs)
	}
	return nil
}

// appendProbe — a command observation takes its place: a repeat probe
// of the same command replaces the previous one (the instance may
// have changed).
func appendProbe(list []canon.ProbeObservation, obs canon.ProbeObservation) []canon.ProbeObservation {
	for i, old := range list {
		if strings.Join(old.Command, " ") == strings.Join(obs.Command, " ") {
			list[i] = obs
			return list
		}
	}
	return append(list, obs)
}

// capturesDiffer — the first difference of two runs of one command:
// an empty string — the facts matched (deterministic).
func capturesDiffer(a, b checks.Captured) string {
	switch {
	case a.ExitCode != b.ExitCode:
		return canondata.T("submit.probe.diff-exit", canondata.M{
			"a": fmt.Sprintf("%d", a.ExitCode), "b": fmt.Sprintf("%d", b.ExitCode),
		})
	case a.Stdout != b.Stdout:
		return canondata.T("submit.probe.diff-stdout")
	case a.Stderr != b.Stderr:
		return canondata.T("submit.probe.diff-stderr")
	case len(a.Files) != len(b.Files):
		return canondata.T("submit.probe.diff-files-count", canondata.M{
			"a": fmt.Sprintf("%d", len(a.Files)), "b": fmt.Sprintf("%d", len(b.Files)),
		})
	}
	for path, content := range a.Files {
		other, ok := b.Files[path]
		if !ok {
			return canondata.T("submit.probe.diff-file-one-run", canondata.M{"path": path})
		}
		if string(yamlio.NormalizeEOL([]byte(content))) != string(yamlio.NormalizeEOL([]byte(other))) {
			return canondata.T("submit.probe.diff-file", canondata.M{"path": path})
		}
	}
	return ""
}

func (b *builder) applyIntent(in *delta.Intent) error {
	if in == nil || strings.TrimSpace(in.Text) == "" {
		return rejection("intent", canondata.T("submit.reject.intent-text-empty"))
	}
	text := strings.TrimSpace(in.Text)
	b.state.Intent = &text
	b.state.IntentAccepted = true
	// The wish's language: a deterministic script detector — the
	// share of Han characters, then of Cyrillic letters, against data
	// thresholds; the language lives in state and drives the machine's
	// human blocks.
	b.state.Language = detectLanguage(text)
	return nil
}

// detectLanguage — script detection: Han characters ≥ the delivery
// threshold → zh; otherwise Cyrillic letters ≥ the threshold → ru;
// otherwise en. Han is checked first, so a mixed script resolves
// deterministically. No guessing: only countable letters.
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

// validateLintFindings — the shape of a clarification batch: at most
// three questions, each with 2–3 options and exactly one default; the
// machine assigns the IDs by order. One validation for a standalone
// lint delta and for questions riding with the checks table.
func validateLintFindings(findings []canon.LintFinding) error {
	// questions.batch-max — the clarification batch ceiling: questions
	// to a person are counted one by one, the cycle must fit into one
	// short round.
	if max := canondata.Limit("questions.batch-max"); len(findings) > max {
		return rejection("questions", canondata.T("submit.reject.questions-exceed", canondata.M{
			"got": fmt.Sprintf("%d", len(findings)),
			"max": fmt.Sprintf("%d", max),
		}))
	}
	for i, f := range findings {
		wantID := fmt.Sprintf("Q%d", i+1)
		if f.ID != wantID {
			return rejection("questions", canondata.T("submit.reject.finding-id-order", canondata.M{
				"index": fmt.Sprintf("%d", i+1), "want": wantID,
			}))
		}
		switch f.Dimension {
		case canon.LintAchievable, canon.LintAmbiguous, canon.LintConflict:
		default:
			return rejection("questions", canondata.T("submit.reject.finding-dimension", canondata.M{"id": f.ID}))
		}
		if strings.TrimSpace(f.Finding) == "" {
			return rejection("questions", canondata.T("submit.reject.finding-empty", canondata.M{"id": f.ID}))
		}
		if strings.TrimSpace(f.Question) == "" {
			return rejection("questions", canondata.T("submit.reject.question-empty", canondata.M{"id": f.ID}))
		}
		if len(f.Options) < 2 || len(f.Options) > 3 {
			return rejection("questions", canondata.T("submit.reject.options-count", canondata.M{
				"id": f.ID, "got": fmt.Sprintf("%d", len(f.Options)),
			}))
		}
		defaults := 0
		for j, o := range f.Options {
			wantOpt := string(rune('a' + j))
			if o.ID != wantOpt {
				return rejection("questions", canondata.T("submit.reject.option-id-order", canondata.M{
					"id": f.ID, "index": fmt.Sprintf("%d", j+1), "want": wantOpt,
				}))
			}
			if strings.TrimSpace(o.Label) == "" {
				return rejection("questions", canondata.T("submit.reject.option-label-empty", canondata.M{
					"id": f.ID, "option": o.ID,
				}))
			}
			if strings.TrimSpace(o.Scope) == "" {
				return rejection("questions", canondata.T("submit.reject.option-scope-empty", canondata.M{
					"id": f.ID, "option": o.ID,
				}))
			}
			if o.Default {
				defaults++
			}
		}
		if defaults != 1 {
			return rejection("questions", canondata.T("submit.reject.defaults-count", canondata.M{
				"id": f.ID, "got": fmt.Sprintf("%d", defaults),
			}))
		}
	}
	return nil
}

// applyLint accepts goal-lint findings (a legacy of a separate intake
// cycle; the new clarification cycle rides as questions with the
// checks table).

// stageRejectReason — a kind rejection by stage also names the legal
// way forward: a formal rejection without a route leaves the executor
// without a next step (the wish after the verdict — rows/spec, a
// contradiction — a lint conflict finding).
func stageRejectReason(kind, stage string) string {
	reason := canondata.T("submit.reject.kind-stage", canondata.M{
		"kind":  fmt.Sprintf("%q", kind),
		"stage": stage,
	})
	if kind == delta.KindIntent && stage != canon.StageIntake {
		reason += canondata.T("submit.reject.kind-stage-path")
	}
	return reason
}

func (b *builder) applyLint(li *delta.Lint) error {
	if li == nil {
		return rejection("lint", canondata.T("submit.reject.empty-delta"))
	}
	if !b.state.IntentAccepted {
		return rejection("lint", canondata.T("submit.reject.lint-no-intent"))
	}
	if err := validateLintFindings(li.Findings); err != nil {
		return err
	}
	// Outside intake, lint is conflict escalation only: the goal is
	// already in progress, the stage's regular goal questions have
	// passed; a conflict can open at any time (a contradictory wish
	// after the verdict too).
	escalation := b.state.Stage != canon.StageIntake
	if escalation {
		for _, f := range li.Findings {
			if f.Dimension != canon.LintConflict {
				return rejection("lint", canondata.T("submit.reject.lint-stage-conflict-only", canondata.M{
					"id": f.ID, "dimension": f.Dimension, "stage": b.state.Stage,
				}))
			}
		}
	}
	if b.state.Lint != nil && !escalation {
		return rejection("lint", canondata.T("submit.reject.lint-duplicate"))
	}
	findings := make([]canon.LintFinding, len(li.Findings))
	copy(findings, li.Findings)
	if b.state.Lint != nil {
		// An escalation is reported into the same batch; the machine
		// continues the numbering: the delta carries its own Q1.., the
		// store holds running numbers, answers reference the stored
		// IDs.
		for i := range findings {
			findings[i].ID = fmt.Sprintf("Q%d", len(b.state.Lint.Findings)+i+1)
		}
		findings = append(append([]canon.LintFinding{}, b.state.Lint.Findings...), findings...)
	}
	b.state.Lint = &canon.LintResult{Findings: findings}
	if len(findings) == 0 {
		// the goal is clear — no questions, nothing to resolve
		b.state.Clarified = true
	} else {
		b.state.Clarified = coversFindings(b.state.Clarifications, findings)
	}
	return nil
}

// resolveLintBatch closes the open clarification batch: explicit
// answers are applied, the unanswered are silence, the default is
// recorded with the ByDefault label (the clarification journal). A
// conflict is not closed by silence: an explicit live choice — the
// escalation keeps the batch open, the stage waits (a silent stop is
// forbidden). Questions answered in earlier submissions are not
// re-asked and not rewritten. The caller must make sure the batch is
// open (Lint with questions, Clarified=false).
func (b *builder) resolveLintBatch(answers []delta.ClarifyAnswer) error {
	byQuestion := map[string]canon.LintFinding{}
	for _, f := range b.state.Lint.Findings {
		byQuestion[f.ID] = f
	}
	// Already recorded choices (earlier submissions): a resolved
	// question is not subject to re-asking or rewriting.
	resolved := map[string]bool{}
	for _, c := range b.state.Clarifications {
		resolved[c.Question] = true
	}
	answered := map[string]string{}
	for i, a := range answers {
		f, ok := byQuestion[a.Question]
		if !ok {
			return rejection("answers", canondata.T("submit.reject.answer-not-found", canondata.M{
				"index": fmt.Sprintf("%d", i+1), "question": a.Question,
			}))
		}
		if resolved[a.Question] {
			return rejection("answers", canondata.T("submit.reject.answer-already-resolved", canondata.M{
				"index": fmt.Sprintf("%d", i+1), "question": a.Question,
			}))
		}
		if _, dup := answered[a.Question]; dup {
			return rejection("answers", canondata.T("submit.reject.answer-twice", canondata.M{
				"index": fmt.Sprintf("%d", i+1), "question": a.Question,
			}))
		}
		found := false
		for _, o := range f.Options {
			if o.ID == a.Option {
				found = true
				break
			}
		}
		if !found {
			return rejection("answers", canondata.T("submit.reject.answer-not-option", canondata.M{
				"index": fmt.Sprintf("%d", i+1), "option": a.Option, "question": a.Question,
			}))
		}
		answered[a.Question] = a.Option
	}
	for _, f := range b.state.Lint.Findings {
		if resolved[f.ID] {
			continue
		}
		chosen, explicit := answered[f.ID]
		if !explicit && f.Dimension == canon.LintConflict {
			// Escalation: a conflict is closed only by an explicit
			// answer — a default applied by silence does not resolve
			// it.
			continue
		}
		if !explicit {
			for _, o := range f.Options {
				if o.Default {
					chosen = o.ID
					break
				}
			}
		}
		for _, o := range f.Options {
			if o.ID != chosen {
				continue
			}
			b.state.Clarifications = append(b.state.Clarifications, canon.Clarification{
				Question: f.ID, Dimension: f.Dimension, Option: o.ID,
				Label: o.Label, Scope: o.Scope, ByDefault: !explicit,
			})
		}
	}
	// The batch is closed when every finding has a recorded choice;
	// the closure counters count only the findings resolved by THIS
	// submission: earlier closures are booked by their own events, a
	// repeated recount would inflate defaults with answers of former
	// batches.
	b.state.Clarified = coversFindings(b.state.Clarifications, b.state.Lint.Findings)
	b.batchTotal = 0
	b.batchAnswered, b.batchDefaults = 0, 0
	for _, f := range b.state.Lint.Findings {
		if resolved[f.ID] {
			continue
		}
		b.batchTotal++
		for _, c := range b.state.Clarifications {
			if c.Question != f.ID {
				continue
			}
			if c.ByDefault {
				b.batchDefaults++
			} else {
				b.batchAnswered++
			}
			break
		}
	}
	return nil
}

// upgradeSilentAnswers applies explicit answers to a CLOSED batch: a
// person may override a default recorded by silence with a word — the
// record is promoted to explicit (ByDefault cleared, the ledger
// recounts attention with an answer-upgrade event). Already explicit
// choices are not rewritten ("answered is not rewritten") — the
// answer is named in a notice in the submission reply, not silently
// lost. The shape of the answers is validated by the same rules as
// for an open batch.
func (b *builder) upgradeSilentAnswers(answers []delta.ClarifyAnswer) error {
	byQuestion := map[string]canon.LintFinding{}
	for _, f := range b.state.Lint.Findings {
		byQuestion[f.ID] = f
	}
	byIndex := map[string]int{}
	for i, c := range b.state.Clarifications {
		byIndex[c.Question] = i
	}
	seen := map[string]bool{}
	upgraded, kept := []string{}, []string{}
	for i, a := range answers {
		f, ok := byQuestion[a.Question]
		if !ok {
			return rejection("answers", canondata.T("submit.reject.answer-not-found", canondata.M{
				"index": fmt.Sprintf("%d", i+1), "question": a.Question,
			}))
		}
		if seen[a.Question] {
			return rejection("answers", canondata.T("submit.reject.answer-twice", canondata.M{
				"index": fmt.Sprintf("%d", i+1), "question": a.Question,
			}))
		}
		seen[a.Question] = true
		var chosen *canon.LintOption
		for j := range f.Options {
			if f.Options[j].ID == a.Option {
				chosen = &f.Options[j]
				break
			}
		}
		if chosen == nil {
			return rejection("answers", canondata.T("submit.reject.answer-not-option", canondata.M{
				"index": fmt.Sprintf("%d", i+1), "option": a.Option, "question": a.Question,
			}))
		}
		idx, ok := byIndex[a.Question]
		if !ok {
			// a closed batch covers all findings with records; no
			// record — the answer cannot be silently dropped here
			// either: an honesty notice
			kept = append(kept, a.Question+" "+a.Option)
			continue
		}
		c := &b.state.Clarifications[idx]
		if !c.ByDefault {
			kept = append(kept, a.Question+" "+a.Option)
			continue
		}
		c.Option, c.Label, c.Scope, c.ByDefault = chosen.ID, chosen.Label, chosen.Scope, false
		upgraded = append(upgraded, a.Question+" "+a.Option)
	}
	if len(upgraded) > 0 {
		b.answerNotes = append(b.answerNotes, canondata.T("submit.answers.upgraded", canondata.M{
			"list": strings.Join(upgraded, ", "),
		}))
		b.answerUpgraded = len(upgraded)
	}
	if len(kept) > 0 {
		b.answerNotes = append(b.answerNotes, canondata.T("submit.answers.kept", canondata.M{
			"list": strings.Join(kept, ", "),
		}))
	}
	return nil
}

// coversFindings — every finding of the current batch has a recorded
// choice (records of former batches do not count).
func coversFindings(cls []canon.Clarification, findings []canon.LintFinding) bool {
	covered := map[string]bool{}
	for _, c := range cls {
		covered[c.Question] = true
	}
	for _, f := range findings {
		if !covered[f.ID] {
			return false
		}
	}
	return true
}

// applyClarify accepts a person's answers to the open clarification
// batch (a legacy of a separate cycle; the new path — answers ride
// with the checks table). A missing answer is silence: the default is
// applied.
func (b *builder) applyClarify(cl *delta.Clarify) error {
	if cl == nil {
		return rejection("clarify", canondata.T("submit.reject.empty-delta"))
	}
	if b.state.Lint == nil {
		return rejection("clarify", canondata.T("submit.reject.clarify-no-lint"))
	}
	if len(b.state.Lint.Findings) == 0 {
		return rejection("clarify", canondata.T("submit.reject.clarify-no-questions"))
	}
	if b.state.Clarified {
		return rejection("clarify", canondata.T("submit.reject.clarify-already"))
	}
	return b.resolveLintBatch(cl.Answers)
}

func (b *builder) applySpec(sp *delta.Spec) error {
	if sp == nil || len(sp.Operations) == 0 {
		return rejection("spec", canondata.T("submit.reject.spec-operations-empty"))
	}
	// The seed channel's quoting trap (see applyChecks): the scenario
	// map form of a seed carries the same double-quote hazard — the
	// refusal fires from the raw text on this path too.
	if b.rawInput != "" {
		if line, trap := delta.SeedQuoteEscape(b.rawInput); trap {
			return rejection("seeds", canondata.T("checktable.error.seed-quote-escape",
				canondata.M{"line": fmt.Sprintf("%d", line)}))
		}
		if line, trap := delta.ExpectationQuoteEscape(b.rawInput); trap {
			return rejection("expectations", canondata.T("checktable.error.expectation-quote-escape",
				canondata.M{"line": fmt.Sprintf("%d", line)}))
		}
	}
	// Referential integrity of operations is checked at intake, not
	// at apply: holding a nonexistent ID in a hold is pointless (the
	// "not found" error used to surface only at approve, the hand paid
	// with a reject + a resubmission). The validation mirrors the
	// apply: references are read against the canon and this delta's
	// additions in operation order.
	if err := b.validateSpecRefs(sp); err != nil {
		return err
	}
	// Channel parity: on deliver a change request is only what
	// touches the proven or changes boundaries (replacements and
	// deletions of proven rows, edits and deletions of requirements);
	// pure growth (additions) and repair of what proved nothing are
	// free — the same semantics as the checks table: edits identical
	// in meaning are not gated differently depending on the delta
	// kind. A mixed submission is partitioned per operation: free
	// operations are applied by this same transaction, the held part
	// waits for the person.
	if b.state.Stage == canon.StageDeliver && !b.assertedAmend {
		held, free, changes := b.partitionSpecOps(sp)
		if len(held) > 0 {
			if len(held) == len(sp.Operations) {
				// nothing to hold apart from boundaries — the
				// submission whole
				b.stageChangeRequest(delta.KindSpec, b.deltaKey, "", changes)
				return nil
			}
			if !b.freeOpsKeepSpecIR(free) {
				// The free operations applied alone would leave the
				// spec IR invalid — the parts are one amendment, not
				// two submissions: the whole delta stages, the
				// approve applies it atomically. Splitting here
				// wedged the composition repair: the free part was
				// refused by the very invariant the held part
				// resolves.
				changes = append(changes, canondata.T("amend.change.atomic"))
				b.stageChangeRequest(delta.KindSpec, b.deltaKey, "", changes)
				return nil
			}
			data, err := yamlio.Marshal(delta.Spec{
				Kind: delta.KindSpec, SubmissionKey: b.deltaKey, Operations: held,
			})
			if err != nil {
				return fmt.Errorf("held delta: %w", err)
			}
			// The plan is not attached: the free operations of
			// this same submission already produced the mass
			// effects — what is asked is the boundaries.
			b.stageAmend(delta.KindSpec, b.deltaKey, string(data), changes, nil)
			sp.Operations = free
		}
	}
	// After code has started, updating behavior fields and removing an
	// already-run scenario require a person's assertion; the
	// submission becomes a deferred edit in full.
	if b.e.codeStarted() && !b.assertedAmend {
		heldIdx, changes := b.specTriggerOps(sp)
		if len(heldIdx) > 0 {
			if len(heldIdx) == len(sp.Operations) {
				b.stageAmend(delta.KindSpec, b.deltaKey, "", changes, nil)
				return nil
			}
			held, free := splitSpecOps(sp.Operations, heldIdx)
			if !b.freeOpsKeepSpecIR(free) {
				// Same atomicity as on deliver: the free operations
				// alone would leave the spec IR invalid — one delta,
				// one amendment, one approve.
				changes = append(changes, canondata.T("amend.change.atomic"))
				b.stageAmend(delta.KindSpec, b.deltaKey, "", changes, nil)
				return nil
			}
			data, err := yamlio.Marshal(delta.Spec{
				Kind: delta.KindSpec, SubmissionKey: b.deltaKey, Operations: held,
			})
			if err != nil {
				return fmt.Errorf("held delta: %w", err)
			}
			b.stageAmend(delta.KindSpec, b.deltaKey, string(data), changes, nil)
			sp.Operations = free
		}
	}
	// On implement the spec's composition is serviced, not rebuilt:
	// update-scenario repairs a pinned form, add-scenario grows the
	// composition, remove-scenario removes a row that has NEVER gone
	// green (a row without proof is not a fact, a garbage row does not
	// stay in the composition) — billing covers only the fresh
	// checks. Requirements and deletions of proven rows wait for a
	// change request on deliver. A person's assertion (approve of a
	// held edit) bypasses the freeze: the freeze protects the proven
	// from the hand, and the operator's assertion is exactly the
	// assertion the freeze demands.
	// The freeze starts with the FIRST ACCEPTED CODE submission, not
	// with the stage: the stage jumps to implement when the checks
	// table turns executable — before any code exists (the rows run
	// red against nothing). Until code is delivered the composition
	// may be re-cut freely, exactly as AGENTS.md promises ("re-cut
	// and name them in living words BEFORE the first code
	// submission"); a stage-keyed guard contradicted that contract.
	if b.state.Stage == canon.StageImplement && b.e.codeStarted() && !b.assertedAmend {
		for _, op := range sp.Operations {
			if op.UpdateScenario == nil && op.AddScenario == nil && op.UpdateAssertions == nil &&
				!(op.RemoveScenario != nil && b.scnNeverGreen(op.RemoveScenario.ID)) &&
				!(op.RemoveRequirement != nil && b.reqNeverGreen(op.RemoveRequirement.ID)) {
				return rejection("spec", canondata.T("submit.reject.spec-implement-frozen"))
			}
		}
	}
	for _, op := range sp.Operations {
		var err error
		switch {
		case op.AddRequirement != nil:
			err = b.opAddReq(op.AddRequirement)
		case op.UpdateRequirement != nil:
			err = b.opUpdateReq(op.UpdateRequirement)
		case op.RemoveRequirement != nil:
			err = b.opRemoveReq(op.RemoveRequirement)
		case op.AddScenario != nil:
			err = b.opAddScn(op.AddScenario)
		case op.UpdateScenario != nil:
			err = b.opUpdateScn(op.UpdateScenario)
		case op.UpdateAssertions != nil:
			err = b.opUpdateAsserts(op.UpdateAssertions)
		case op.RemoveScenario != nil:
			err = b.opRemoveScn(op.RemoveScenario)
		default:
			err = rejection("operation", canondata.T("submit.reject.spec-op-empty"))
		}
		if err != nil {
			return err
		}
	}
	// Invariant: a requirement without a scenario is not accepted —
	// spec IR validity, not a matter of taste: not subject to
	// rejection.
	if err := b.checkSpecIR(); err != nil {
		return err
	}
	// Growth on implement: new scenarios land in a card by state-file
	// ownership — the split is frozen, growth can only go inside an
	// existing facet.
	if b.state.Stage == canon.StageImplement {
		for _, op := range sp.Operations {
			if op.AddScenario != nil {
				if err := b.attachGrownScenario(op.AddScenario.ID); err != nil {
					return err
				}
			}
		}
	}
	// Change request: a spec on deliver rolls the cycle back to
	// acceptance compilation — the change passes the same stages
	// anew.
	if b.state.Stage == canon.StageDeliver {
		b.state.Stage = canon.StageACCompile
		b.refreshCard()
		b.reopenedCycle = true
	}
	return nil
}

// refreshCard collapses the split after a change request: a spec
// change rolls the cycle back and reopens decomposition — one card
// for the whole product again, contracts unfrozen until the next
// split.
func (b *builder) refreshCard() {
	for _, id := range sortedIDs(b.cards) {
		delete(b.cards, id)
		b.effects = append(b.effects, journal.Effect{Path: canonFile("cards", id), Delete: true})
		b.touch(id)
	}
	ids := make([]string, 0, len(b.scns))
	for _, sc := range b.scns {
		if sc.Prose == nil {
			ids = append(ids, sc.ID)
		}
	}
	sortStrings(ids)
	next := b.state.Counters["CRD"]
	if next < 1 {
		next = 1
	}
	card := canon.Card{
		ID:        fmt.Sprintf("CRD-%03d", next),
		Facet:     "whole-product",
		Scenarios: ids,
		Status:    canon.CardActive,
	}
	_ = b.putCard(card)
	for _, id := range sortedIDs(b.contracts) {
		c := b.contracts[id]
		if c.Status == canon.IfcFrozen {
			c.Status = canon.IfcDraft
			_ = b.putContract(c)
		}
	}
}

// attachGrownScenario places a scenario added on implement into the
// card owning its state files: the split is frozen, a new scenario
// must live inside one facet. The whole-product card (without file
// ownership) accepts everything; prose is not implemented by cards
// and is not embedded. A green card reopens: a new scenario is new
// work of the same facet.
func (b *builder) attachGrownScenario(id string) error {
	sc, ok := b.scns[id]
	if !ok || sc.Prose != nil {
		return nil
	}
	files := stateFilesOf(sc)
	owners := map[string]bool{}
	for _, cid := range sortedIDs(b.cards) {
		for _, f := range b.cards[cid].Files {
			for _, sf := range files {
				if f == sf {
					owners[cid] = true
				}
			}
		}
	}
	var target string
	switch {
	case len(owners) == 1:
		for cid := range owners {
			target = cid
		}
	case len(owners) == 0 && len(b.cards) == 1:
		for cid := range b.cards {
			target = cid
		}
	case len(owners) > 1:
		return rejection(id, canondata.T("submit.reject.owners-cross-facets", canondata.M{
			"owners": strings.Join(sortedIDs2(owners), ", "),
		}))
	default:
		return rejection(id, canondata.T("submit.reject.no-owner-facet"))
	}
	card := b.cards[target]
	if card.Status == canon.CardQuarantine {
		return rejection(id, canondata.T("submit.reject.card-quarantined", canondata.M{"card": target}))
	}
	if !slices.Contains(card.Scenarios, id) {
		card.Scenarios = append(card.Scenarios, id)
		sortStrings(card.Scenarios)
	}
	if card.Status == canon.CardGreen {
		card.Status = canon.CardActive
	}
	return b.putCard(card)
}

func (b *builder) opAddReq(op *delta.AddRequirement) error {
	if !validID("REQ", op.ID) {
		return rejection(op.ID, canondata.T("submit.reject.id-req"))
	}
	if _, ok := b.reqs[op.ID]; ok {
		return rejection(op.ID, canondata.T("submit.reject.id-taken"))
	}
	if strings.TrimSpace(op.Formulation) == "" {
		return rejection(op.ID, canondata.T("submit.reject.formulation-empty"))
	}
	return b.putReq(canon.Requirement{
		ID: op.ID, Status: canon.ReqAccepted,
		Dependencies: op.Dependencies, Formulation: strings.TrimSpace(op.Formulation),
		Detail: op.Detail,
	})
}

func (b *builder) opUpdateReq(op *delta.UpdateRequirement) error {
	r, ok := b.reqs[op.ID]
	if !ok {
		return rejection(op.ID, canondata.T("submit.reject.not-found"))
	}
	if op.Formulation != nil {
		if strings.TrimSpace(*op.Formulation) == "" {
			return rejection(op.ID, canondata.T("submit.reject.formulation-empty"))
		}
		r.Formulation = strings.TrimSpace(*op.Formulation)
	}
	if op.Dependencies != nil {
		r.Dependencies = op.Dependencies
	}
	if op.Detail != nil {
		r.Detail = op.Detail
	}
	return b.putReq(r)
}

func (b *builder) opRemoveReq(op *delta.RemoveRequirement) error {
	if _, ok := b.reqs[op.ID]; !ok {
		return rejection(op.ID, canondata.T("submit.reject.not-found"))
	}
	// Before the first accepted code submission (the pre-code
	// window) a machine-drafted placeholder requirement goes in ONE
	// operation, together with its never-green scenarios — the
	// row-first-then-requirement ordering forced placeholder cleanup
	// through the waive/amend channel, paying a human round for
	// machine noise. A requirement with a green row is not a draft:
	// its removal keeps the change-request channel, exactly as after
	// code.
	if !b.e.codeStarted() {
		var proven []string
		for _, id := range sortedIDs(b.scns) {
			if b.scns[id].Requirement != op.ID {
				continue
			}
			if b.scnNeverGreen(id) {
				delete(b.scns, id)
				b.effects = append(b.effects, journal.Effect{
					Path: canonFile("scenarios", id), Delete: true,
				})
				b.touch(id)
				continue
			}
			proven = append(proven, id)
		}
		if len(proven) == 0 {
			delete(b.reqs, op.ID)
			b.effects = append(b.effects, journal.Effect{
				Path: canonFile("requirements", op.ID), Delete: true,
			})
			b.touch(op.ID)
			return nil
		}
		return rejection(op.ID, canondata.T("submit.reject.req-has-proven-scenarios",
			canondata.M{"scns": strings.Join(proven, ", ")}))
	}
	// Post-code removal of a placeholder (every scenario never
	// green) was STAGED as a deferred edit by the pre-apply hold;
	// this is its asserted replay — the requirement goes in one
	// operation with its never-green scenarios, exactly as pre-code.
	// A proven row keeps the refusal: nothing green is weakened
	// without the change-request channel.
	if b.assertedAmend {
		var proven []string
		for _, id := range sortedIDs(b.scns) {
			if b.scns[id].Requirement != op.ID {
				continue
			}
			if b.scnNeverGreen(id) {
				delete(b.scns, id)
				b.effects = append(b.effects, journal.Effect{
					Path: canonFile("scenarios", id), Delete: true,
				})
				b.touch(id)
				continue
			}
			proven = append(proven, id)
		}
		if len(proven) > 0 {
			return rejection(op.ID, canondata.T("submit.reject.req-has-proven-scenarios",
				canondata.M{"scns": strings.Join(proven, ", ")}))
		}
		delete(b.reqs, op.ID)
		b.effects = append(b.effects, journal.Effect{
			Path: canonFile("requirements", op.ID), Delete: true,
		})
		b.touch(op.ID)
		return nil
	}
	for _, sc := range b.scns {
		if sc.Requirement == op.ID {
			return rejection(op.ID, canondata.T("submit.reject.req-has-scenarios"))
		}
	}
	delete(b.reqs, op.ID)
	b.effects = append(b.effects, journal.Effect{
		Path: canonFile("requirements", op.ID), Delete: true,
	})
	b.touch(op.ID)
	return nil
}

// resolveSeedDigests — a scenario seed of the form {path: X, digest:
// D}: the bytes are taken from the journal by digest (the FILES
// WRITTEN mechanism). The ordinary {path, content} form passes
// through.
func (b *builder) resolveSeedDigests(seed []canon.Seed) ([]canon.Seed, error) {
	for i := range seed {
		if seed[i].Content != "" || seed[i].Digest == "" {
			continue
		}
		if !validCanonDigest(seed[i].Digest) {
			return nil, rejection("seed", canondata.T("submit.reject.seed-digest-format", canondata.M{
				"path": seed[i].Path, "digest": seed[i].Digest,
			}))
		}
		content, ok := b.e.canonContentByDigest(seed[i].Digest)
		if !ok {
			return nil, rejection("seed", canondata.T("submit.reject.seed-digest-unknown", canondata.M{
				"path": seed[i].Path, "digest": seed[i].Digest,
			}))
		}
		seed[i].Content = content
		seed[i].Digest = ""
	}
	return seed, nil
}

// decodeSeedLiterals — the single literal decoder for the seeds of
// all surfaces: scenario seed content is decoded by the same
// DecodeLiteral as the one-line table seeds (\n, \t, \r, \\, \xNN →
// bytes). Applied to the NON-canonicalized delta input; the canon
// already stores canonicalized bytes and is not decoded again.
func decodeSeedLiterals(seed []canon.Seed) []canon.Seed {
	for i := range seed {
		seed[i].Content = yamlio.DecodeLiteral(seed[i].Content)
	}
	return seed
}

func (b *builder) opAddScn(op *delta.AddScenario) error {
	if !validID("SCN", op.ID) {
		return rejection(op.ID, canondata.T("submit.reject.id-scn"))
	}
	if _, ok := b.scns[op.ID]; ok {
		return rejection(op.ID, canondata.T("submit.reject.id-taken"))
	}
	if _, ok := b.reqs[op.Requirement]; !ok {
		return rejection(op.ID, canondata.T("submit.reject.requirement-not-found", canondata.M{
			"id": op.Requirement,
		}))
	}
	opSeed, err := b.resolveSeedDigests(op.Seed)
	if err != nil {
		return err
	}
	sc := canon.Scenario{
		ID: op.ID, Requirement: op.Requirement,
		Summary: strings.TrimSpace(op.Summary),
		Seed:    decodeSeedLiterals(opSeed), Materials: op.Materials, Pre: op.Pre,
		When: op.When, Then: op.Then, Prose: op.Prose,
	}
	b.noteSeedQuotes(op.ID, sc.Seed)
	if err := normalizeAssertions(sc.Then); err != nil {
		return rejection(op.ID, err.Error())
	}
	if err := b.validateScenario(sc); err != nil {
		return err
	}
	return b.putScn(sc)
}

func (b *builder) opUpdateScn(op *delta.UpdateScenario) error {
	sc, ok := b.scns[op.ID]
	if !ok {
		return rejection(op.ID, canondata.T("submit.reject.not-found"))
	}
	if op.Summary != nil {
		sc.Summary = strings.TrimSpace(*op.Summary)
	}
	if op.Seed != nil {
		updSeed, err := b.resolveSeedDigests(op.Seed)
		if err != nil {
			return err
		}
		sc.Seed = decodeSeedLiterals(updSeed)
		b.noteSeedQuotes(op.ID, sc.Seed)
	}
	if op.Materials != nil {
		sc.Materials = op.Materials
	}
	if op.Pre != nil {
		sc.Pre = op.Pre
	}
	if op.When != nil {
		sc.When = op.When
	}
	if op.Then != nil {
		if err := normalizeAssertions(op.Then); err != nil {
			return rejection(op.ID, err.Error())
		}
		sc.Then = op.Then
	}
	if op.Prose != nil {
		sc.Prose = op.Prose
	}
	// Rebinding is a boundary regrouping: expectations do not change,
	// no assertion required; the new requirement must exist
	// (otherwise the row dangles — the "requirement without a
	// scenario" invariant is caught by the same validation below).
	if op.Requirement != nil && *op.Requirement != sc.Requirement {
		if _, ok := b.reqs[*op.Requirement]; !ok {
			return rejection(op.ID, canondata.T("submit.reject.requirement-not-found", canondata.M{
				"id": *op.Requirement,
			}))
		}
		sc.Requirement = *op.Requirement
	}
	if err := b.validateScenario(sc); err != nil {
		return err
	}
	return b.putScn(sc)
}

// opUpdateAsserts — surgical editing of expectations: individual
// operations over the then observations without retranscribing the
// block. The same shape control and the same result validation as a
// full then replacement.
func (b *builder) opUpdateAsserts(op *delta.UpdateAssertions) error {
	if op.Scenario == "" || len(op.Ops) == 0 {
		return rejection("spec", canondata.T("submit.reject.assertion-ops-empty"))
	}
	sc, ok := b.scns[op.Scenario]
	if !ok {
		return rejection(op.Scenario, canondata.T("submit.reject.not-found"))
	}
	then, err := b.applyAssertionOps(op.Scenario, &sc, op.Ops)
	if err != nil {
		return err
	}
	if err := normalizeAssertions(then); err != nil {
		return rejection(op.Scenario, err.Error())
	}
	sc.Then = then
	if err := b.validateScenario(sc); err != nil {
		return err
	}
	return b.putScn(sc)
}

// applyAssertionOps — applying surgical operations to the observation
// list: match must equal exactly one observation of the current list
// (the operation order is literal — each sees the result of the
// previous ones), set replaces, drop removes, capture replaces with
// the value measured by a machine run of the delivered product, add
// appends to the end.
func (b *builder) applyAssertionOps(scenario string, sc *canon.Scenario, ops []delta.AssertionOp) ([]canon.Assertion, error) {
	out := slices.Clone(sc.Then)
	for i, op := range ops {
		num := fmt.Sprintf("%d", i+1)
		switch {
		case op.Match != nil && op.Set != nil && !op.Drop && !op.Capture:
			idx, n := matchAssertion(out, *op.Match)
			switch n {
			case 0:
				return nil, rejection(scenario, canondata.T("submit.reject.assertion-not-found",
					canondata.M{"scenario": scenario, "num": num}))
			case 1:
				out[idx] = *op.Set
			default:
				return nil, rejection(scenario, canondata.T("submit.reject.assertion-ambiguous",
					canondata.M{"scenario": scenario, "num": num, "count": fmt.Sprintf("%d", n)}))
			}
		case op.Match != nil && op.Drop && op.Set == nil && !op.Capture:
			idx, n := matchAssertion(out, *op.Match)
			switch n {
			case 0:
				return nil, rejection(scenario, canondata.T("submit.reject.assertion-not-found",
					canondata.M{"scenario": scenario, "num": num}))
			case 1:
				out = slices.Delete(out, idx, idx+1)
			default:
				return nil, rejection(scenario, canondata.T("submit.reject.assertion-ambiguous",
					canondata.M{"scenario": scenario, "num": num, "count": fmt.Sprintf("%d", n)}))
			}
		case op.Match != nil && op.Capture && op.Set == nil && !op.Drop && op.Add == nil:
			idx, n := matchAssertion(out, *op.Match)
			switch n {
			case 0:
				return nil, rejection(scenario, canondata.T("submit.reject.assertion-not-found",
					canondata.M{"scenario": scenario, "num": num}))
			case 1:
				fresh, err := b.captureAssertion(sc, out[idx], num)
				if err != nil {
					return nil, err
				}
				out[idx] = fresh
			default:
				return nil, rejection(scenario, canondata.T("submit.reject.assertion-ambiguous",
					canondata.M{"scenario": scenario, "num": num, "count": fmt.Sprintf("%d", n)}))
			}
		case op.Add != nil && op.Match == nil && op.Set == nil && !op.Drop && !op.Capture:
			out = append(out, *op.Add)
		default:
			return nil, rejection(scenario, canondata.T("submit.reject.assertion-op-form",
				canondata.M{"num": num}))
		}
	}
	return out, nil
}

// captureAssertion — an expectation by measurement: the machine runs
// the delivered product along the scenario and puts the observed
// actual value as the new expectation. Nobody types the bytes by
// hand; the person asserts the meaning on the hold, seeing "was →
// measured". The capture is honest only for exact expectations
// (equals, and for files also bytes-equals) and only when the
// behavior has already changed: an unchanged measurement is a
// rejection, not a silent confirmation of the status quo.
func (b *builder) captureAssertion(sc *canon.Scenario, old canon.Assertion, num string) (canon.Assertion, error) {
	if old.Condition != "equals" && !(old.Observation == "file" && old.Condition == "bytes-equals") {
		return old, rejection(sc.ID, canondata.T("submit.reject.capture-condition",
			canondata.M{"scenario": sc.ID, "num": num, "observation": old.Observation, "condition": old.Condition}))
	}
	if sc.When == nil || len(sc.When.Command) == 0 {
		return old, rejection(sc.ID, canondata.T("submit.reject.capture-no-when",
			canondata.M{"scenario": sc.ID, "num": num}))
	}
	cap, err := checks.CaptureSeq(b.e.Workdir, b.e.surfaceArtifact(), sc.Seed, sc.Materials, sc.Pre, sc.When.Command, sc.When.TimeoutSec)
	if err != nil {
		return old, rejection(sc.ID, canondata.T("submit.reject.capture-run",
			canondata.M{"scenario": sc.ID, "num": num, "error": err.Error()}))
	}
	fresh := old
	switch old.Observation {
	case "stdout":
		fresh.Value = cap.Stdout
	case "stderr":
		fresh.Value = cap.Stderr
	case "exit-code":
		fresh.Value = strconv.Itoa(cap.ExitCode)
	case "file":
		v, ok := cap.Files[old.Path]
		if !ok {
			return old, rejection(sc.ID, canondata.T("submit.reject.capture-file",
				canondata.M{"scenario": sc.ID, "num": num, "path": old.Path}))
		}
		fresh.Value = v
	default:
		return old, rejection(sc.ID, canondata.T("submit.reject.capture-observation",
			canondata.M{"scenario": sc.ID, "num": num, "observation": old.Observation}))
	}
	if fresh == old {
		return old, rejection(sc.ID, canondata.T("submit.reject.capture-unchanged",
			canondata.M{"scenario": sc.ID, "num": num, "observation": old.Observation}))
	}
	return fresh, nil
}

// matchAssertion — how many observations of the list equal the
// pattern, and the index of the first. Equality by the decoded typed
// form: a compact shorthand and the full form of one observation are
// equal.
func matchAssertion(then []canon.Assertion, want canon.Assertion) (int, int) {
	idx, n := -1, 0
	for i, a := range then {
		if a == want {
			if n == 0 {
				idx = i
			}
			n++
		}
	}
	return idx, n
}

// reqNeverGreen — a requirement entirely without green proof: all of
// its rows have never gone green (or there are no rows) — deleting it
// together with them breaks nothing; a garbage composition is
// cleaned freely.
func (b *builder) reqNeverGreen(id string) bool {
	for scnID, sc := range b.scns {
		if sc.Requirement != id {
			continue
		}
		if !b.scnNeverGreen(scnID) {
			return false
		}
	}
	return true
}

// scnNeverGreen — a row has never gone green: deleting it proves and
// breaks nothing; the composition is cleaned freely.
func (b *builder) scnNeverGreen(id string) bool {
	check, ok := b.checks["TST-"+idNumSuffix(id)]
	return !ok || !check.Proven
}

func (b *builder) opRemoveScn(op *delta.RemoveScenario) error {
	if _, ok := b.scns[op.ID]; !ok {
		return rejection(op.ID, canondata.T("submit.reject.not-found"))
	}
	delete(b.scns, op.ID)
	b.effects = append(b.effects, journal.Effect{
		Path: canonFile("scenarios", op.ID), Delete: true,
	})
	b.touch(op.ID)
	return nil
}

// validateScenario — semantics: prose without steps and assertions,
// non-prose with a mandatory when and valid assertions.
func (b *builder) validateScenario(sc canon.Scenario) error {
	if sc.Prose != nil {
		if len(sc.Then) > 0 || sc.When != nil {
			return rejection(sc.ID, canondata.T("submit.reject.prose-steps"))
		}
		if strings.TrimSpace(*sc.Prose) == "" {
			return rejection(sc.ID, canondata.T("submit.reject.prose-empty"))
		}
		return nil
	}
	if sc.When == nil {
		return rejection(sc.ID, canondata.T("submit.reject.when-required"))
	}
	if err := canon.ValidateWhen(sc.When); err != nil {
		return rejection(sc.ID, err.Error())
	}
	if err := normalizeStdinEncoding(sc.When); err != nil {
		return rejection(sc.ID, err.Error())
	}
	for i, pre := range sc.Pre {
		if len(pre) == 0 {
			return rejection(sc.ID, canondata.T("submit.reject.pre-empty", canondata.M{
				"index": fmt.Sprintf("%d", i+1),
			}))
		}
	}
	for i, a := range sc.Then {
		if err := canon.ValidateAssertion(a); err != nil {
			return rejection(sc.ID, canondata.T("submit.reject.then-invalid", canondata.M{
				"index": fmt.Sprintf("%d", i+1), "err": err.Error(),
			}))
		}
	}
	return nil
}

// normalizeAssertions — expectation encoding of full spec forms: the
// same normalization as for table rows (rowExpectations), in one
// pass — a second pass over already-decoded values is impossible.
func normalizeAssertions(then []canon.Assertion) error {
	for i := range then {
		if err := normalizeAssertionEncoding(&then[i]); err != nil {
			return err
		}
	}
	return nil
}

// normalizeAssertionEncoding — validation of expectation encoding on
// input: a literal "\n"/"\t"/"\r" or a lone backslash in a value is
// an encoding error (a control character was intended, two bytes came
// out) — a rejection with a recipe, not an eternal red of the run;
// \xNN escapes expand into bytes (a literal backslash is \x5C).
// Values of semantic JSON conditions are not touched: JSON has its
// own escapes, validated by parsing.
func normalizeAssertionEncoding(a *canon.Assertion) error {
	switch a.Condition {
	case "json-equals", "json-contains", "fails":
		return nil
	}
	if a.Value == "" {
		return nil
	}
	if err := yamlio.ValidateExpectLiteral(a.Value); err != nil {
		return err
	}
	a.Value = yamlio.DecodeHexOnly(a.Value)
	return nil
}

// normalizeStdinEncoding — the same encoding language for stdin: YAML
// writes control characters itself, a literal backslash is \x5C,
// other lone escapes are an error with a recipe.
func normalizeStdinEncoding(w *canon.When) error {
	if w.Stdin == nil || *w.Stdin == "" {
		return nil
	}
	if err := yamlio.ValidateExpectLiteral(*w.Stdin); err != nil {
		return err
	}
	v := yamlio.DecodeHexOnly(*w.Stdin)
	w.Stdin = &v
	return nil
}

func (b *builder) applyAssert(as *delta.Assert) error {
	if as == nil {
		return rejection("assert", canondata.T("submit.reject.empty-delta"))
	}
	sc, ok := b.scns[as.Scenario]
	if !ok {
		return rejection(as.Scenario, canondata.T("submit.reject.not-found"))
	}
	if sc.Prose != nil {
		return rejection(as.Scenario, canondata.T("submit.reject.assert-prose"))
	}
	if len(sc.Then) > 0 {
		return rejection(as.Scenario, canondata.T("submit.reject.assert-pinned"))
	}
	if len(as.Assertions) == 0 {
		return rejection(as.Scenario, canondata.T("submit.reject.assertions-empty"))
	}
	for i := range as.Assertions {
		if err := normalizeAssertionEncoding(&as.Assertions[i]); err != nil {
			return rejection(as.Scenario, canondata.T("submit.reject.assertion-invalid", canondata.M{
				"index": fmt.Sprintf("%d", i+1), "err": err.Error(),
			}))
		}
		if err := canon.ValidateAssertion(as.Assertions[i]); err != nil {
			return rejection(as.Scenario, canondata.T("submit.reject.assertion-invalid", canondata.M{
				"index": fmt.Sprintf("%d", i+1), "err": err.Error(),
			}))
		}
	}
	sc.Then = as.Assertions
	return b.putScn(sc)
}

// applySplit carries out the split of the product into facets of
// parallel work. The machine validates: the split is possible only
// before the first code, scenarios are distributed exactly once, file
// ownership is exclusive, contracts are versioned and frozen by
// acceptance. A state file shared by two cards without a contract
// between them — a rejection: parallelism outside frozen contracts is
// forbidden.
func (b *builder) applySplit(sp *delta.Split) error {
	if sp == nil || len(sp.Cards) == 0 {
		return rejection("split", canondata.T("submit.reject.split-cards-empty"))
	}
	for _, c := range b.cards {
		if c.Status != canon.CardActive || c.Attempts > 0 {
			return rejection("split", canondata.T("submit.reject.split-started", canondata.M{
				"card": c.ID, "status": c.Status, "attempts": fmt.Sprintf("%d", c.Attempts),
			}))
		}
	}

	cardIDs := map[string]bool{}
	facetOf := map[string]string{} // facet → card
	filesOfCard := map[string][]string{}
	covered := map[string]bool{}
	scenarioFiles := map[string][]string{} // card → state files of its scenarios
	for _, c := range sp.Cards {
		if !validID("CRD", c.ID) {
			return rejection("split", canondata.T("submit.reject.id-crd", canondata.M{"id": c.ID}))
		}
		if cardIDs[c.ID] {
			return rejection("split", canondata.T("submit.reject.card-listed-twice", canondata.M{"id": c.ID}))
		}
		cardIDs[c.ID] = true
		if !facetPattern.MatchString(c.Facet) {
			return rejection("split", canondata.T("submit.reject.facet-format", canondata.M{
				"id": c.ID, "facet": fmt.Sprintf("%q", c.Facet),
			}))
		}
		if other, ok := facetOf[c.Facet]; ok {
			return rejection("split", canondata.T("submit.reject.facet-taken", canondata.M{
				"id": c.ID, "facet": fmt.Sprintf("%q", c.Facet), "other": other,
			}))
		}
		facetOf[c.Facet] = c.ID
		if len(c.Scenarios) == 0 {
			return rejection("split", canondata.T("submit.reject.card-scenarios-empty", canondata.M{"id": c.ID}))
		}
		for _, sid := range c.Scenarios {
			sc, ok := b.scns[sid]
			if !ok {
				return rejection("split", canondata.T("submit.reject.card-scenario-not-found", canondata.M{
					"id": c.ID, "scenario": sid,
				}))
			}
			if sc.Prose != nil {
				return rejection("split", canondata.T("submit.reject.card-scenario-prose", canondata.M{
					"id": c.ID, "scenario": sid,
				}))
			}
			if covered[sid] {
				return rejection("split", canondata.T("submit.reject.scenario-listed-twice", canondata.M{"id": sid}))
			}
			covered[sid] = true
			scenarioFiles[c.ID] = mergeUnique(scenarioFiles[c.ID], stateFilesOf(sc))
		}
		if len(c.Files) == 0 {
			return rejection("split", canondata.T("submit.reject.card-files-empty", canondata.M{"id": c.ID}))
		}
		for _, f := range c.Files {
			clean, err := cleanProjectPath(f)
			if err != nil {
				return rejection("split", canondata.T("submit.reject.card-file-path", canondata.M{
					"id": c.ID, "file": f, "err": err.Error(),
				}))
			}
			filesOfCard[c.ID] = append(filesOfCard[c.ID], clean)
		}
	}
	// file ownership is exclusive: two parallel executors do not
	// write one file
	owner := map[string]string{}
	for cardID, files := range filesOfCard {
		for _, f := range files {
			if other, ok := owner[f]; ok {
				return rejection("split", canondata.T("submit.reject.file-owned-by-both", canondata.M{
					"file": f, "other": other, "card": cardID,
				}))
			}
			owner[f] = cardID
		}
	}
	// every executable scenario is covered by exactly one card
	for _, id := range sortedIDs(b.scns) {
		if b.scns[id].Prose == nil && !covered[id] {
			return rejection("split", canondata.T("submit.reject.scenario-not-covered", canondata.M{"id": id}))
		}
	}

	// contracts: IDs, versions, sides from the declared facets
	declared := map[string]bool{}
	for _, k := range sp.Contracts {
		if !validID("IFC", k.ID) {
			return rejection("split", canondata.T("submit.reject.id-ifc", canondata.M{"id": k.ID}))
		}
		if declared[k.ID] {
			return rejection("split", canondata.T("submit.reject.contract-listed-twice", canondata.M{"id": k.ID}))
		}
		declared[k.ID] = true
		if !versionPattern.MatchString(k.Version) {
			return rejection("split", canondata.T("submit.reject.contract-version-format", canondata.M{
				"id": k.ID, "version": fmt.Sprintf("%q", k.Version),
			}))
		}
		if len(k.Surface) == 0 {
			return rejection("split", canondata.T("submit.reject.contract-surface-empty", canondata.M{"id": k.ID}))
		}
		for _, line := range k.Surface {
			if strings.TrimSpace(line) == "" {
				return rejection("split", canondata.T("submit.reject.contract-surface-line", canondata.M{"id": k.ID}))
			}
		}
		sides := map[string]bool{}
		for _, s := range k.Sides {
			if !facetOfKnown(s, facetOf) {
				return rejection("split", canondata.T("submit.reject.contract-side-unknown", canondata.M{
					"id": k.ID, "side": fmt.Sprintf("%q", s),
				}))
			}
			if sides[s] {
				return rejection("split", canondata.T("submit.reject.contract-side-twice", canondata.M{
					"id": k.ID, "side": fmt.Sprintf("%q", s),
				}))
			}
			sides[s] = true
		}
		if len(sides) < 2 {
			return rejection("split", canondata.T("submit.reject.contract-sides-few", canondata.M{"id": k.ID}))
		}
		if old, ok := b.contracts[k.ID]; ok && cmpVersion(k.Version, old.Version) <= 0 {
			return rejection("split", canondata.T("submit.reject.contract-version-grow", canondata.M{
				"id": k.ID, "version": k.Version, "old": old.Version,
			}))
		}
	}
	// parallelism outside frozen contracts is forbidden: a state file
	// shared by two cards must be bound by a contract between their
	// facets
	ids := sortedIDs2(filesOfCard)
	for i := 0; i < len(ids); i++ {
		for j := i + 1; j < len(ids); j++ {
			shared := intersectFiles(scenarioFiles[ids[i]], scenarioFiles[ids[j]])
			if len(shared) == 0 {
				continue
			}
			if !pairContracted(sp.Contracts, facetByCard(facetOf, ids[i]), facetByCard(facetOf, ids[j])) {
				return rejection("split", canondata.T("submit.reject.share-without-contract", canondata.M{
					"a": ids[i], "b": ids[j], "files": strings.Join(shared, ", "),
				}))
			}
		}
	}

	// effects: old cards out, new ones and contracts into the canon;
	// freezing is a fact of acceptance, the agent does not choose the
	// status
	for _, id := range sortedIDs(b.cards) {
		delete(b.cards, id)
		b.effects = append(b.effects, journal.Effect{Path: canonFile("cards", id), Delete: true})
	}
	for _, c := range sp.Cards {
		if err := b.putCard(canon.Card{
			ID: c.ID, Facet: c.Facet, Scenarios: c.Scenarios,
			Files: filesOfCard[c.ID], Status: canon.CardActive,
		}); err != nil {
			return err
		}
	}
	for _, k := range sp.Contracts {
		if err := b.putContract(canon.Contract{
			ID: k.ID, Version: k.Version, Sides: k.Sides,
			Surface: k.Surface, Status: canon.IfcFrozen,
		}); err != nil {
			return err
		}
	}
	return nil
}

// stateFilesOf — the state files of a scenario: seeded files and the
// file observations of assertions. These are the observable shared
// assets of facets.
func stateFilesOf(sc canon.Scenario) []string {
	var out []string
	for _, sd := range sc.Seed {
		out = mergeUnique(out, []string{sd.Path})
	}
	for _, m := range sc.Materials {
		out = mergeUnique(out, []string{m.From})
	}
	for _, a := range sc.Then {
		if a.Observation == "file" && a.Path != "" {
			out = mergeUnique(out, []string{a.Path})
		}
	}
	return out
}

func mergeUnique(base, add []string) []string {
	seen := map[string]bool{}
	for _, s := range base {
		seen[s] = true
	}
	for _, s := range add {
		if !seen[s] {
			seen[s] = true
			base = append(base, s)
		}
	}
	return base
}

func intersectFiles(a, b []string) []string {
	set := map[string]bool{}
	for _, s := range b {
		set[s] = true
	}
	var out []string
	for _, s := range a {
		if set[s] {
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}

func facetOfKnown(facet string, facetOf map[string]string) bool {
	_, ok := facetOf[facet]
	return ok
}

func facetByCard(facetOf map[string]string, cardID string) string {
	for facet, id := range facetOf {
		if id == cardID {
			return facet
		}
	}
	return cardID
}

// pairContracted — whether a declared contract names both facets (in
// any order).
func pairContracted(list []delta.SplitContract, facetA, facetB string) bool {
	for _, k := range list {
		hasA, hasB := false, false
		for _, s := range k.Sides {
			if s == facetA {
				hasA = true
			}
			if s == facetB {
				hasB = true
			}
		}
		if hasA && hasB {
			return true
		}
	}
	return false
}

// cmpVersion compares semantic triples: −1, 0, 1.
func cmpVersion(a, b string) int {
	pa, pb := parseVersion(a), parseVersion(b)
	for i := 0; i < 3; i++ {
		switch {
		case pa[i] < pb[i]:
			return -1
		case pa[i] > pb[i]:
			return 1
		}
	}
	return 0
}

func parseVersion(v string) [3]int {
	var out [3]int
	fmt.Sscanf(v, "%d.%d.%d", &out[0], &out[1], &out[2])
	return out
}

func sortedIDs2[T any](m map[string]T) []string {
	return sortedIDs(m)
}

func (b *builder) applyFix(fx *delta.Fix) error {
	if fx == nil {
		return rejection("fix", canondata.T("submit.reject.empty-delta"))
	}
	if err := b.acceptAnswers(fx.Answers); err != nil {
		return err
	}
	if !gateNames[fx.ReplyTo] {
		return rejection("fix", canondata.T("submit.reject.reply-to-gate", canondata.M{
			"gate": fmt.Sprintf("%q", fx.ReplyTo),
		}))
	}
	card, ok := b.cards[fx.Card]
	if !ok {
		return rejection("fix", canondata.T("submit.reject.card-not-found", canondata.M{"id": fx.Card}))
	}
	if err := b.applyCodeFiles(fx.Card, fx.Files, fx.Delete); err != nil {
		return err
	}
	// The attempt budget is counted AFTER the run: before it, the
	// card's behaviorally red rows are remembered, and the gates
	// themselves charge the attempt — only to a surviving red (see
	// gatesForCode). A repair over someone else's reds no longer
	// burns the budget at submission time.
	b.chargeReds = b.cardBehaviorReds(card)
	b.gateCard = fx.Card
	return nil
}

// acceptAnswers — the answer transport in code/fix submissions: the
// explicit-answer window is not lost after code starts — an open
// batch is resolved, in a closed one a default is promoted to the
// person's explicit choice (the same mechanism as the checks table;
// answers without a batch — an honest rejection).
func (b *builder) acceptAnswers(answers []delta.ClarifyAnswer) error {
	if len(answers) == 0 {
		return nil
	}
	if b.state.Lint != nil && len(b.state.Lint.Findings) > 0 && !b.state.Clarified && !b.assertedAmend {
		if err := b.resolveLintBatch(answers); err != nil {
			return err
		}
		b.resolvedBatch = b.state.Clarified
		return nil
	}
	return b.upgradeSilentAnswers(answers)
}

func (b *builder) applyCode(cd *delta.Code) error {
	if cd == nil {
		return rejection("code", canondata.T("submit.reject.empty-delta"))
	}
	if err := b.acceptAnswers(cd.Answers); err != nil {
		return err
	}
	card, ok := b.cards[cd.Card]
	if !ok {
		return rejection("code", canondata.T("submit.reject.card-not-found", canondata.M{"id": cd.Card}))
	}
	if err := b.applyCodeFiles(cd.Card, cd.Files, cd.Delete); err != nil {
		return err
	}
	// Regeneration lives in the same budget as repair (one hard
	// ceiling, there are no eternal loops), but the count comes after
	// the run and only for a surviving red.
	b.chargeReds = b.cardBehaviorReds(card)
	b.gateCard = cd.Card
	return nil
}

// cardBehaviorReds — the set of the card's behaviorally red checks (a
// red outcome, not an environment failure): only such rows may spend
// the attempt budget, and only after surviving the next submission's
// run.
func (b *builder) cardBehaviorReds(card canon.Card) map[string]bool {
	reds := map[string]bool{}
	for _, scnID := range card.Scenarios {
		check, ok := b.checks["TST-"+idNumSuffix(scnID)]
		if !ok {
			continue
		}
		if check.Outcome == canon.CheckRed && !check.EnvFail {
			reds[check.ID] = true
		}
	}
	return reds
}

// applyCodeFiles writes project files as transaction effects. Paths
// are relative and stay within the project; the canon is inviolable.
// A split card may write only its own files: foreign ownership — a
// rejection.
// canonContentByDigest searches the journal backwards for the last
// file content with the given digest: a repeat submission references
// the past file without resending it.
func (e *Engine) canonContentByDigest(digest string) (string, bool) {
	entries := e.Journal.All()
	for i := len(entries) - 1; i >= 0; i-- {
		effects := entries[i].Effects
		for j := len(effects) - 1; j >= 0; j-- {
			if yamlio.Digest([]byte(effects[j].Content)) == digest {
				return effects[j].Content, true
			}
		}
	}
	return "", false
}

// validCanonDigest — the canon digest format: exactly 64 lowercase
// hex characters; the machine itself prints these bytes (FILES
// WRITTEN).
func validCanonDigest(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, c := range s {
		if !('0' <= c && c <= '9' || 'a' <= c && c <= 'f') {
			return false
		}
	}
	return true
}

func (b *builder) applyCodeFiles(cardID string, files []delta.FileChange, del []string) error {
	if len(files) == 0 && len(del) == 0 {
		return rejection(cardID, canondata.T("submit.reject.files-empty"))
	}
	type fileOp struct {
		path, content string
		del           bool
	}
	var ops []fileOp
	for _, f := range files {
		clean, err := cleanProjectPath(f.Path)
		if err != nil {
			return rejection(cardID, canondata.T("submit.reject.path-invalid", canondata.M{
				"path": f.Path, "err": err.Error(),
			}))
		}
		content := f.Content
		if f.Digest != "" {
			if content != "" {
				return rejection(cardID, canondata.T("submit.reject.digest-and-content", canondata.M{"file": clean}))
			}
			// Own identifier format: the canon writes 64 lowercase
			// hex; anything else is not canon content but a foreign
			// transcription — a rejection before the lookup, with a
			// reason and a rule.
			if !validCanonDigest(f.Digest) {
				return rejection(cardID, canondata.T("submit.reject.digest-format", canondata.M{
					"file": clean, "digest": f.Digest,
				}))
			}
			got, ok := b.e.canonContentByDigest(f.Digest)
			if !ok {
				return rejection(cardID, canondata.T("submit.reject.digest-unknown", canondata.M{
					"file": clean, "digest": f.Digest,
				}))
			}
			content = got
		}
		ops = append(ops, fileOp{path: clean, content: content})
	}
	for _, p := range del {
		clean, err := cleanProjectPath(p)
		if err != nil {
			return rejection(cardID, canondata.T("submit.reject.delete-path-invalid", canondata.M{
				"path": p, "err": err.Error(),
			}))
		}
		ops = append(ops, fileOp{path: clean, del: true})
	}
	if card, ok := b.cards[cardID]; ok && len(card.Files) > 0 {
		owned := map[string]bool{}
		for _, f := range card.Files {
			owned[f] = true
		}
		owner := map[string]string{}
		for _, c := range b.cards {
			for _, f := range c.Files {
				owner[f] = c.ID
			}
		}
		for _, op := range ops {
			if owned[op.path] {
				continue
			}
			if who, taken := owner[op.path]; taken {
				return rejection(cardID, canondata.T("submit.reject.path-owned", canondata.M{
					"path": op.path, "owner": who,
				}))
			}
			return rejection(cardID, canondata.T("submit.reject.path-not-in-card", canondata.M{"path": op.path}))
		}
	}
	// Digests of what was written go to the hand in the reply: truth
	// comes from the machine, the hand does not compute digests
	// itself. Not a single new byte (all writes byte-for-byte with
	// the project, deletions already absent) — a pure re-run: effects
	// are not applied, counters do not move.
	changed := false
	for _, op := range ops {
		if op.del {
			b.effects = append(b.effects, journal.Effect{Path: op.path, Delete: true})
			if _, err := os.Stat(filepath.Join(b.e.Workdir, filepath.FromSlash(op.path))); err == nil {
				changed = true
			}
			continue
		}
		b.effects = append(b.effects, journal.Effect{Path: op.path, Content: op.content})
		b.writtenFiles = append(b.writtenFiles, fileWritten{
			path: op.path, digest: yamlio.Digest([]byte(op.content)),
		})
		cur, err := os.ReadFile(filepath.Join(b.e.Workdir, filepath.FromSlash(op.path)))
		if err != nil || string(cur) != op.content {
			changed = true
		}
	}
	b.pureRerun = !changed
	return nil
}

// fileWritten — a file written by a submission with the canonical
// digest of its content (short form — to the hand in the reply).
type fileWritten struct {
	path, digest string
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
	return canon.MachineHeader("submit") + "\n" + string(data)
}

func validID(kind, id string) bool {
	switch kind {
	case "REQ":
		return strings.HasPrefix(id, "REQ-") && len(id) > 4 && isDigits(id[4:])
	case "SCN":
		return strings.HasPrefix(id, "SCN-") && len(id) > 4 && isDigits(id[4:])
	case "IFC":
		return strings.HasPrefix(id, "IFC-") && len(id) > 4 && isDigits(id[4:])
	case "CRD":
		return strings.HasPrefix(id, "CRD-") && len(id) > 4 && isDigits(id[4:])
	}
	return false
}

func isDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return len(s) > 0
}

// cleanProjectPath — a relative path within the project, not in the
// instance directory.
func cleanProjectPath(p string) (string, error) {
	clean := path.Clean(p)
	if path.IsAbs(clean) || strings.HasPrefix(clean, "..") || clean == "." {
		return "", fmt.Errorf("%s", canondata.T("submit.reject.path-not-relative"))
	}
	if clean == ".punchtape" || strings.HasPrefix(clean, ".punchtape/") {
		return "", fmt.Errorf("%s", canondata.T("submit.reject.instance-untouchable"))
	}
	return clean, nil
}

func sortedIDs[T any](m map[string]T) []string {
	ids := make([]string, 0, len(m))
	for id := range m {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// checkDependencies — dependencies exist and there are no cycles.
func checkDependencies(reqs map[string]canon.Requirement) error {
	const (
		white = 0
		grey  = 1
		black = 2
	)
	color := map[string]int{}
	var visit func(id, from string) error
	visit = func(id, from string) error {
		switch color[id] {
		case grey:
			return rejection(id, canondata.T("submit.reject.deps-cycle", canondata.M{"id": from}))
		case black:
			return nil
		}
		color[id] = grey
		for _, dep := range reqs[id].Dependencies {
			if _, ok := reqs[dep]; !ok {
				return rejection(id, canondata.T("submit.reject.deps-not-found", canondata.M{"dep": dep}))
			}
			if err := visit(dep, id); err != nil {
				return err
			}
		}
		color[id] = black
		return nil
	}
	for _, id := range sortedIDs(reqs) {
		if err := visit(id, id); err != nil {
			return err
		}
	}
	return nil
}

// reject assembles the rejection reply: an error line plus a skeleton.
func reject(entity, reason, kind string) string {
	line := canondata.T("submit.reject.line", canondata.M{"entity": entity, "reason": reason})
	skel := delta.Skeleton(kind)
	if skel == "" {
		return line
	}
	return line + canondata.T("submit.reject.format") + "\n" + skel
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

// firstErrorLine — the first ERROR line of a multi-line reply: format
// auto-repair lines above the rejection do not overwrite the reason.
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
