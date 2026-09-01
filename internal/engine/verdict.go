// Canonical verdict: fixed format, digest signature.
// The machine renders it; an agent's retelling is not the verdict.
package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/compiler"
	"github.com/neurophant/punchtape/internal/ledger"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// Verdict statuses.
const (
	VerdictReady      = "READY"
	VerdictQuarantine = "QUARANTINE"
	VerdictNotReady   = "NOT-READY"
)

// Verdict — the canonical verdict.
type Verdict struct {
	Status string
	Lines  []string
}

// RenderVerdict assembles the verdict from canon and ledger. The stop
// condition is computed, not judged: every scenario is linked to a green
// check, there are no red ones, the attempts budget is not exhausted,
// prose sits in the taste queue. Cost and latency are release gates:
// their red blocks the release.
//
// At deliver the verdict freezes: rendered once on entering the
// stage, then the .punchtape/verdict.txt file is returned — the digest
// is stable, the file is rendered to the human channel bypassing the
// agent, an agent's retelling is not the verdict. A new cycle (change
// request) changes the journal head — the verdict is re-frozen.
func (e *Engine) RenderVerdict() (*Verdict, error) {
	state, err := e.Store.State()
	if err != nil {
		return nil, err
	}
	if state.Stage != canon.StageDeliver {
		return e.renderVerdictFresh()
	}
	if v, ok := e.frozenVerdict(); ok {
		e.lastVerdictDrift = false
		return v, nil
	}
	v, err := e.renderVerdictFresh()
	if err != nil {
		return nil, err
	}
	// A drifted instance is not frozen: the verdict is recomputed until
	// the instance returns to the acceptance digest (re-verification by
	// a fix).
	if !e.lastVerdictDrift {
		e.freezeVerdict(v)
	}
	return v, nil
}

// frozenVerdict — the frozen delivery verdict: a file keyed by the
// journal head. If the head has not changed, the verdict is the same,
// byte for byte. The world is part of the key: a surface that has
// moved away from the acceptance digest (external corruption) makes
// the freeze stale — the verdict is rendered anew and honestly names
// the drift.
func (e *Engine) frozenVerdict() (*Verdict, bool) {
	keyPath := filepath.Join(e.Store.Root(), "cache", "verdict-key")
	textPath := filepath.Join(e.Store.Root(), "cache", "verdict.txt")
	key, err := os.ReadFile(keyPath)
	if err != nil || strings.TrimSpace(string(key)) != e.journalHead() {
		return nil, false
	}
	if drifted, _, _ := e.acceptanceDrift(); drifted {
		return nil, false
	}
	data, err := os.ReadFile(textPath)
	if err != nil {
		return nil, false
	}
	return verdictFromText(string(data)), true
}

// freezeVerdict fixes the delivery verdict with a file: a cache keyed
// by the journal head, written atomically. The same freeze regenerates
// the system hand's artifacts in the instance's data: SPEC.md (a canon
// render), the specs/ directory (the human spec: a file per feature +
// an index.md registry), SUMMARY.md (a human-readable verdict summary)
// and CHANGELOG.md (a diff of the instance's life from the journal);
// the working directory's root stays clean — only the AGENTS.md
// self-doc.
func (e *Engine) freezeVerdict(v *Verdict) {
	root := e.Store.Root()
	_ = yamlio.WriteAtomic(filepath.Join(root, "cache", "verdict-key"), []byte(e.journalHead()))
	_ = yamlio.WriteAtomic(filepath.Join(root, "cache", "verdict.txt"), []byte(v.Text()))
	writeSpecRender(e)
	writeFeatureRender(e)
	writeSummaryRender(e)
	writeChangelogRender(e)
	// Legacy root renders (from before the artifacts moved into the
	// instance's data): the root holds only the AGENTS.md self-doc.
	removeLegacyRender(e.Workdir, "SPEC.md")
	removeLegacyRender(e.Workdir, "SUMMARY.md")
	removeLegacyRender(e.Workdir, "CHANGELOG.md")
}

// verdictFromText parses the frozen verdict text: the status is the
// first line, the rest as is.
func verdictFromText(text string) *Verdict {
	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
	status := ""
	if len(lines) > 0 {
		status = strings.TrimPrefix(lines[0], "VERDICT: ")
	}
	return &Verdict{Status: status, Lines: lines}
}

func (e *Engine) renderVerdictFresh() (*Verdict, error) {
	scns, err := e.Store.Scenarios()
	if err != nil {
		return nil, err
	}
	chk, err := e.Store.Checks()
	if err != nil {
		return nil, err
	}
	cards, err := e.Store.Cards()
	if err != nil {
		return nil, err
	}
	state, err := e.Store.State()
	if err != nil {
		return nil, err
	}
	greenByScn := map[string]bool{}
	for _, c := range chk {
		if c.Outcome == canon.CheckGreen {
			greenByScn[c.Scenario] = true
		}
	}

	var executable, prose, greenScn, redChk, greenChk int
	var tasteQueue []string
	for _, sc := range scns {
		d := compiler.DegreeOf(sc)
		switch d.Level {
		case compiler.DegreeProse:
			prose++
			tasteQueue = append(tasteQueue, canondata.T("verdict.taste-item", canondata.M{
				"id": sc.ID, "summary": sc.Summary, "reason": d.Reason,
			}))
		case compiler.DegreeExecutable:
			executable++
		}
		if greenByScn[sc.ID] {
			greenScn++
		}
	}
	// Completeness lint rule waives: active waives with reasons go to
	// the verdict's taste queue — the deviation is visible, no silence.
	for _, w := range state.Waives {
		tasteQueue = append(tasteQueue, canondata.T("verdict.waive-item", canondata.M{
			"rule": w.Rule, "target": w.Target, "reason": w.Reason,
		}))
	}
	for _, c := range chk {
		switch c.Outcome {
		case canon.CheckGreen:
			greenChk++
		case canon.CheckRed:
			redChk++
		}
	}

	// The human spec is part of "done": every delivered feature carries
	// an author's description. The machine checks only non-emptiness —
	// what to write is the executor's knowledge; an honest gap is
	// accepted by the human as a waive with a reason.
	reqs, err := e.Store.Requirements()
	if err != nil {
		return nil, err
	}
	waivedDoc := map[string]bool{}
	for _, w := range state.Waives {
		if w.Rule == "feature.doc" {
			waivedDoc[w.Target] = true
		}
	}
	var noDoc []string
	described := 0
	for _, r := range reqs {
		if r.Status != canon.ReqAccepted && r.Status != canon.ReqRealized {
			continue
		}
		if waivedDoc[r.ID] {
			continue
		}
		if r.SpecDoc == nil || strings.TrimSpace(r.SpecDoc.Text) == "" {
			noDoc = append(noDoc, r.ID)
		} else {
			described++
		}
	}

	var quarantine []string
	for _, c := range cards {
		if c.Status != canon.CardQuarantine {
			continue
		}
		// The quarantine reason is visible from the verdict itself:
		// quarantine on static blockers with green checks and no reason
		// would force the reader into the ledger.
		line := canondata.T("verdict.quarantine-item", canondata.M{
			"id": c.ID, "attempts": strconv.Itoa(c.Attempts),
		})
		if c.Gap != "" {
			line += canondata.T("verdict.quarantine-gap", canondata.M{"gap": c.Gap})
		}
		if len(c.Blockers) > 0 {
			line += canondata.T("verdict.quarantine-blockers", canondata.M{
				"blockers": strings.Join(c.Blockers, "; "),
			})
		}
		// Quarantine is a live human decision point (escalation rc 4):
		// the verdict names it itself and does not leave the reader to
		// guess what comes next.
		line += canondata.T("verdict.quarantine-ask")
		quarantine = append(quarantine, line)
	}

	costLine, costFail := e.costGate()
	rentLine := e.gateRent(state)
	latencyLine := e.latencyGate()

	dcFail := false
	for _, en := range e.Ledger.All() {
		if en.Event == "gate" && en.Details["name"] == "double-compile" &&
			en.Details["outcome"] != "green" {
			dcFail = true
		}
	}
	// Acceptance broken by env doctor: the recipes' environment fails
	// its probe — the release is not proven, the verdict is not green.
	envFail := acceptanceEnvFail(e)

	status := VerdictReady
	switch {
	case len(quarantine) > 0:
		status = VerdictQuarantine
	case redChk > 0 || greenScn < executable || executable == 0:
		status = VerdictNotReady
	case len(noDoc) > 0:
		status = VerdictNotReady
	case costFail || dcFail || envFail:
		status = VerdictNotReady
	}

	// The verdict gate describes the current instance: a surface changed
	// after the acceptance moment makes the verdict's numbers alien —
	// READY on an unreconciled instance would be a lie (QUARANTINE
	// stays: the budget is a journal fact, not the instance's).
	drifted, oldD, newD := e.acceptanceDrift()
	e.lastVerdictDrift = drifted
	if drifted && status == VerdictReady {
		status = VerdictNotReady
	}

	// Stage counts: submissions and runs from the journal.
	submitted, runs := 0, 0
	for _, en := range e.Journal.All() {
		if en.DeltaKind == deltaKindRun {
			runs++
		} else {
			submitted++
		}
	}

	var wallMs, calls, tokens int64
	for _, en := range e.Ledger.All() {
		wallMs += en.WallMs
		if en.Event == "verb" {
			calls++
		}
		if en.Tokens != nil {
			tokens += *en.Tokens
		}
	}

	v := &Verdict{Status: status}
	// The acceptance line: on drift it names the drift itself — the
	// verdict is honest about which instance its numbers describe.
	accLine := acceptanceLine(e)
	if drifted {
		accLine = canondata.T("verdict.acceptance-drift", canondata.M{
			"old": oldD, "new": newD,
		})
	}
	headLine := canondata.T("verdict.head", canondata.M{"status": status})
	// A held amend changes the picture of "done": the verdict describes
	// the state BEFORE it, and the line must say so — otherwise "READY"
	// reads as "everything done" and the hold is lost.
	if state.PendingAmend != nil {
		headLine += canondata.T("verdict.head-held", canondata.M{
			"count": fmt.Sprintf("%d", len(state.PendingAmend.Changes)),
		})
	}
	v.Lines = []string{
		headLine,
		"", // the digest is inserted after the fields are assembled
		canondata.T("verdict.scenarios", canondata.M{
			"total": strconv.Itoa(len(scns)), "green": strconv.Itoa(greenScn),
			"red": strconv.Itoa(redChk), "prose": strconv.Itoa(described),
		}),
		canondata.T("verdict.traceability", canondata.M{
			"green": strconv.Itoa(greenScn), "executable": strconv.Itoa(executable),
		}),
		canondata.T("verdict.checks", canondata.M{
			"green": strconv.Itoa(greenChk), "red": strconv.Itoa(redChk),
			"total": strconv.Itoa(len(chk)),
		}),
		e.derivedLine(state),
		canondata.T("verdict.stages", canondata.M{
			"submitted": strconv.Itoa(submitted), "runs": strconv.Itoa(runs),
		}),
		originLine(e),
		canondata.T("verdict.changelog", canondata.M{
			"count": strconv.Itoa(len(e.Journal.All())),
		}),
		canondata.T("verdict.bill", canondata.M{
			"wall": formatWall(wallMs), "calls": strconv.FormatInt(calls, 10),
			"tokens": tokensOrNone(tokens),
		}),
		e.attentionLine(),
		e.convergeLine(cards),
		accLine,
		doubleCompileLine(e),
		costLine,
		rentLine,
		latencyLine,
		canondata.T("verdict.quarantine", canondata.M{
			"list": listOr(canondata.T("verdict.list-none"), quarantine),
		}),
		canondata.T("verdict.taste-queue", canondata.M{
			"list": listOr(canondata.T("verdict.list-empty"), tasteQueue),
		}),
		"",
		// The price menu: an option must reflect a state fact —
		// reviewing the taste queue is offered only when it is
		// non-empty; for an empty queue "see the list" is noise, not a
		// choice.
		canondata.T("verdict.next"),
		canondata.T("verdict.next-task"),
	}
	if len(tasteQueue) > 0 {
		v.Lines = append(v.Lines, canondata.T("verdict.next-review"))
	}
	if len(noDoc) > 0 {
		v.Lines = append(v.Lines, "", canondata.T("verdict.features-doc", canondata.M{
			"count": strconv.Itoa(len(noDoc)),
			"ids":   strings.Join(noDoc, ", "),
		}))
	}
	// Digest signature: over the normalized text of the fields without
	// the digest line itself.
	body := normalizeVerdict(v.Lines)
	v.Lines[1] = canondata.T("verdict.digest", canondata.M{
		"digest": yamlio.DigestShort([]byte(body)),
	})
	return v, nil
}

// The verdict's latency line is INFORMATIONAL, the verdict does not
// block on it: the duration of honest check work (slow behavior,
// timers, heavy builds) is a property of the task, not a machine
// defect. The line's budget counts machine overhead: the submission
// wall minus the sum of the check/probe run times themselves (the
// ledger's runs-ms detail). Failure is a sustained overshoot (see
// latencyGate): at least two calls over budget that are not adjacent;
// a single sample and an adjacent pair of one burst are environment
// jitter, not a regression.

// readBudgetMs — fixed budgets for the instance read verbs: reading
// the whole instance per call is forbidden.
func readBudgetMs(verb string) int64 {
	if verb == "why" {
		return int64(canondata.Limit("latency.read.why-ms"))
	}
	return int64(canondata.Limit("latency.read.next-ms"))
}

// submitBudgetMs — the OVERHEAD budget of a submit call: check run
// work is subtracted from the wall (runs-ms), the slot term is the
// nested slot rendered in the reply.
func submitBudgetMs(slot bool) int64 {
	budget := int64(canondata.Limit("latency.base-ms"))
	if slot {
		budget += int64(canondata.Limit("latency.slot-ms"))
	}
	return budget
}

// latencyGate — the informational latency line: machine overhead per
// verb (p95), the submit budget counts overhead only, reads are
// fixed. Failure is a sustained overshoot: at least two calls over
// budget that are NOT adjacent (adjacent — consecutive calls of one
// verb with no in-budget call between them: samples of one burst at
// the start of a pair). A single sample is indistinguishable from
// environment jitter, while a machine regression repeats on every
// call — no calm windows remain between overshoots.
func (e *Engine) latencyGate() string {
	type verbCall struct {
		wall      int64
		budget    int64
		hasBudget bool
	}
	byVerb := map[string][]verbCall{}
	for _, en := range e.Ledger.All() {
		if en.Event != "verb" {
			continue
		}
		name := en.Details["name"]
		if name == "" {
			continue
		}
		var budget int64
		hasBudget := true
		switch name {
		case "next", "why":
			budget = readBudgetMs(name)
		case "submit":
			budget = submitBudgetMs(en.Details["slot"] == "yes")
		default:
			hasBudget = false
		}
		// The submit call's wall time is overhead: the honest time of
		// check/probe runs is subtracted, it never goes negative.
		wall := en.WallMs
		if name == "submit" {
			wall -= atoiOrZero64(en.Details["runs-ms"])
			if wall < 0 {
				wall = 0
			}
		}
		byVerb[name] = append(byVerb[name], verbCall{wall: wall, budget: budget, hasBudget: hasBudget})
	}
	if len(byVerb) == 0 {
		return canondata.T("verdict.latency-none")
	}
	names := make([]string, 0, len(byVerb))
	for name := range byVerb {
		names = append(names, name)
	}
	sort.Strings(names)
	parts := make([]string, 0, len(names))
	for _, name := range names {
		calls := byVerb[name]
		walls := make([]int64, 0, len(calls))
		for _, c := range calls {
			walls = append(walls, c.wall)
		}
		sort.Slice(walls, func(i, j int) bool { return walls[i] < walls[j] })
		rank := (95*len(walls) + 99) / 100
		p95 := walls[rank-1]
		worst := walls[len(walls)-1]
		overs, bursts := 0, 0
		prevOver := false
		for _, c := range calls {
			if !c.hasBudget || c.wall <= c.budget {
				prevOver = false
				continue
			}
			overs++
			if !prevOver {
				bursts++
			}
			prevOver = true
		}
		mark := "ok"
		if bursts >= 2 {
			mark = "over"
		} else if overs > 0 {
			mark = canondata.T("verdict.latency-burst", canondata.M{"bursts": strconv.Itoa(bursts)})
		}
		switch {
		case !calls[0].hasBudget:
			parts = append(parts, canondata.T("verdict.latency-part-no-budget", canondata.M{
				"name": name, "p95": strconv.FormatInt(p95, 10),
			}))
		case name == "submit":
			parts = append(parts, canondata.T("verdict.latency-part-submit", canondata.M{
				"name": name, "p95": strconv.FormatInt(p95, 10),
				"worst": strconv.FormatInt(worst, 10),
				"base":  strconv.Itoa(canondata.Limit("latency.base-ms")),
				"mark":  mark,
			}))
		default:
			parts = append(parts, canondata.T("verdict.latency-part-read", canondata.M{
				"name": name, "p95": strconv.FormatInt(p95, 10),
				"budget": strconv.FormatInt(calls[0].budget, 10), "mark": mark,
			}))
		}
	}
	return canondata.T("verdict.latency", canondata.M{
		"mark": "INFO", "parts": strings.Join(parts, ", "),
	})
}

// costGate — the cost gate: ceremony overhead ≤10% of the task. An
// honest count over what the machine measured: tokens of the context
// the machine invoked (slots and replies) against the task's tokens
// by the caller's counter. No counter — no number: session pacing is
// not a cost, the machine has no right to go red on an undefined
// base; a regression against a twin is caught by a live run, not by
// self-attestation.
func (e *Engine) costGate() (string, bool) {
	var ceremonyTokens, taskTokens int64
	for _, en := range e.Ledger.All() {
		switch en.Event {
		case "verb":
			switch en.Details["name"] {
			case "next":
				ceremonyTokens += int64(atoiOrZero(en.Details["context-bytes"]) / 4)
			case "submit":
				ceremonyTokens += int64(atoiOrZero(en.Details["reply-bytes"]) / 4)
				// The verb/submit counter is the agent's self-report,
				// not a machine meter: the bill line and why stage show
				// it, but the verdict gate goes red only on a trusted
				// base (apply/executor are the driver's caller-side
				// counters).
			}
		case "apply", "executor":
			if en.Tokens != nil {
				taskTokens += *en.Tokens
			}
		}
	}
	if taskTokens <= 0 {
		return canondata.T("verdict.cost-none", canondata.M{
			"ceremony": strconv.FormatInt(ceremonyTokens, 10),
		}), false
	}
	pct := ceremonyTokens * 100 / taskTokens
	fail := pct > int64(canondata.Limit("cost.overhead-pct"))
	mark := "PASS"
	if fail {
		mark = "FAIL"
	}
	return canondata.T("verdict.cost", canondata.M{
		"mark": mark, "pct": strconv.FormatInt(pct, 10),
		"ceremony": strconv.FormatInt(ceremonyTokens, 10),
		"task":     strconv.FormatInt(taskTokens, 10),
	}), fail
}

// gateRent — ceremony rent: retired gates are named honestly, live
// ones pay rent by the fact of influencing outcomes.
func (e *Engine) gateRent(state canon.State) string {
	runsByName := map[string]string{}
	for _, en := range e.Ledger.All() {
		if en.Event == "gate-retire" {
			runsByName[en.Details["name"]] = en.Details["runs"]
		}
	}
	if len(state.RetiredGates) == 0 {
		return canondata.T("verdict.rent-none")
	}
	names := append([]string{}, state.RetiredGates...)
	sort.Strings(names)
	parts := make([]string, 0, len(names))
	for _, name := range names {
		parts = append(parts, canondata.T("verdict.rent-part", canondata.M{
			"name": name, "runs": runsByName[name],
		}))
	}
	return canondata.T("verdict.rent", canondata.M{"parts": strings.Join(parts, "; ")})
}

// attentionLine — human attention, computed from the ledger, not
// postulated: the wall from the question batch to the answer (how
// long the human was answering) and the number of input points after
// the choice. Zero between choice and verdict is the goal; here it is
// a ledger fact or honestly non-zero. Attention points are two
// independent streams: the question batch (clarification questions
// and their closing) and the held amend assertion (amend, kind:
// amend-assertion on both events). The line describes the question
// batch; the verdict names the hold with its own held marker. A
// choice without kind in an open-hold window belongs to the hold
// stream: while a hold is open, substantive submissions are
// forbidden, there is no other choice in that window (records
// without kind are legacy).
func (e *Engine) attentionLine() string {
	all := e.Ledger.All()
	var lastBatchQuestion, lastBatchChoice *ledger.Entry
	assertOpen := false
	assertPoints := 0
	for i := range all {
		en := &all[i]
		switch en.Event {
		case "question":
			if en.Details["kind"] == "amend-assertion" {
				assertOpen = true
				if lastBatchChoice != nil {
					assertPoints++
				}
			} else {
				lastBatchQuestion = en
				lastBatchChoice = nil
			}
		case "choice":
			if en.Details["kind"] == "amend-assertion" || assertOpen {
				assertOpen = false
				continue
			}
			lastBatchChoice = en
			assertPoints = 0
		}
	}
	if lastBatchQuestion == nil {
		return canondata.T("verdict.attention-none")
	}
	if lastBatchChoice == nil {
		return canondata.T("verdict.attention-open", canondata.M{
			"questions": lastBatchQuestion.Details["questions"],
		})
	}
	batchMs := lastBatchChoice.At.Sub(lastBatchQuestion.At).Milliseconds()
	if assertPoints > 0 {
		return canondata.T("verdict.attention-closed-points", canondata.M{
			"wall": formatWall(batchMs), "questions": lastBatchChoice.Details["questions"],
			"answered": lastBatchChoice.Details["answered"], "defaults": lastBatchChoice.Details["defaults"],
			"points": strconv.Itoa(assertPoints),
		})
	}
	return canondata.T("verdict.attention-closed", canondata.M{
		"wall": formatWall(batchMs), "questions": lastBatchChoice.Details["questions"],
		"answered": lastBatchChoice.Details["answered"], "defaults": lastBatchChoice.Details["defaults"],
	})
}

// doubleCompileLine — the double-compile gate: the last run from the
// ledger. Never run — an honest line about missing data; red — the
// spec did not reproduce under an independent implementation, the
// release is not proven.
func doubleCompileLine(e *Engine) string {
	var last *ledger.Entry
	for i := range e.Ledger.All() {
		en := &e.Ledger.All()[i]
		if en.Event == "gate" && en.Details["name"] == "double-compile" {
			last = en
		}
	}
	if last == nil {
		return canondata.T("verdict.double-compile-none")
	}
	mark := "PASS"
	if last.Details["outcome"] != "green" {
		mark = "FAIL"
	}
	return canondata.T("verdict.double-compile", canondata.M{
		"mark": mark, "reason": last.Details["reason"],
	})
}

// convergeLine — the auto-convergence summary: how many repair cycles
// the machine spun without a human and how close the cards came to
// the attempts budget. Numbers from the journal and the cards, not
// from self-perception.
func (e *Engine) convergeLine(cards []canon.Card) string {
	fixCycles := 0
	for _, en := range e.Journal.All() {
		if en.DeltaKind == deltaKindFix {
			fixCycles++
		}
	}
	attempts, quarantines := 0, 0
	for _, c := range cards {
		if c.Attempts > attempts {
			attempts = c.Attempts
		}
		if c.Status == canon.CardQuarantine {
			quarantines++
		}
	}
	return canondata.T("verdict.converge", canondata.M{
		"cycles": strconv.Itoa(fixCycles), "attempts": strconv.Itoa(attempts),
		"budget":      strconv.Itoa(canondata.Limit("card.attempts-budget")),
		"quarantines": strconv.Itoa(quarantines),
	})
}

// atoiOrZero — reading a machine number from ledger details.
func atoiOrZero(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}

// atoiOrZero64 — the same for 64-bit milliseconds.
func atoiOrZero64(s string) int64 {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return n
}

// lastAcceptanceEntry — the last acceptance moment in the ledger.
func lastAcceptanceEntry(e *Engine) *ledger.Entry {
	var last *ledger.Entry
	for i := range e.Ledger.All() {
		en := &e.Ledger.All()[i]
		if en.Event == "gate" && en.Details["name"] == "suite" &&
			en.Details["scope"] == "acceptance" {
			last = en
		}
	}
	return last
}

// acceptanceLine — the acceptance line: the last acceptance moment
// (outcomes + the instance state digest). None yet — an honest line.
// Acceptance broken by env doctor is named point by point: which
// tools were not found and what to install.
func acceptanceLine(e *Engine) string {
	last := lastAcceptanceEntry(e)
	if last == nil {
		return canondata.T("verdict.acceptance-none")
	}
	if last.Details["env"] != "" {
		return canondata.T("verdict.acceptance-env", canondata.M{
			"diag": canondata.EnvDiag("exec-not-found", canondata.M{
				"command": last.Details["env"],
			}),
		})
	}
	if last.Details["outcome"] == "green" {
		if last.Details["digest"] == "" {
			return canondata.T("verdict.acceptance-nosurface", canondata.M{
				"green": last.Details["green"],
				"total": last.Details["checks"],
			})
		}
		return canondata.T("verdict.acceptance-pass", canondata.M{
			"digest": last.Details["digest"],
			"green":  last.Details["green"],
			"total":  last.Details["checks"],
		})
	}
	return canondata.T("verdict.acceptance-fail", canondata.M{
		"red": last.Details["red"], "total": last.Details["checks"],
	})
}

// acceptanceEnvFail — whether the last acceptance is red from an
// environment probe.
func acceptanceEnvFail(e *Engine) bool {
	last := lastAcceptanceEntry(e)
	return last != nil && last.Details["env"] != ""
}

// acceptanceDrift — reconciliation of the instance against the
// acceptance moment: the current digest of the surface and of all
// check definitions against the acceptance digest, both sides
// computed through the same normalization (the legacy proven
// backfill): a signature must not depend on which side of a load it
// was computed on. A drift claim needs TWO signatures: an acceptance
// that could not see the surface (an empty recorded digest) or a
// moment where the surface is not measurable says nothing about
// change — a false drift after prose is not verdict material; a
// world that never had a verifiable surface does not count as drift
// either, there is nothing to check there.
func (e *Engine) acceptanceDrift() (drifted bool, oldD, newD string) {
	last := lastAcceptanceEntry(e)
	if last == nil {
		return false, "", ""
	}
	oldD = last.Details["digest"]
	if checks, err := e.Store.Checks(); err == nil {
		if cur, ok := surfaceDigest(e, backfillCheckProven(checks)); ok {
			newD = shortHex(cur)
		}
	}
	if oldD == "" || newD == "" {
		return false, oldD, newD
	}
	return newD != oldD, oldD, newD
}

// originLine — code origin: synthesized — the machine assembled it
// from a fully formal spec.
func originLine(e *Engine) string {
	for _, en := range e.Journal.All() {
		if en.Origin == "synthesized" {
			return canondata.T("verdict.origin")
		}
	}
	return ""
}

// Text — the canonical verdict text.
func (v *Verdict) Text() string {
	return strings.Join(v.Lines, "\n")
}

// normalizeVerdict brings the verdict fields to a normal form: lines
// without empties, UTF-8, stable order — the digest is reproducible.
// Live meters (call cost, attention, latency) do not enter the
// signature: the digest is the identity of the verdict's state, not
// of environment counters; the same verdict is signed identically
// before and after unrelated verb calls.
func normalizeVerdict(lines []string) string {
	var sb strings.Builder
	for _, l := range lines {
		if strings.TrimSpace(l) == "" || strings.HasPrefix(l, "digest:") ||
			strings.HasPrefix(l, "bill:") || strings.HasPrefix(l, "attention:") ||
			strings.HasPrefix(l, "cost gate:") || strings.HasPrefix(l, "latency") {
			continue
		}
		sb.WriteString(l)
		sb.WriteString("\n")
	}
	return sb.String()
}

func formatWall(ms int64) string {
	d := time.Duration(ms) * time.Millisecond
	return d.Truncate(time.Millisecond).String()
}

func tokensOrNone(tokens int64) string {
	if tokens > 0 {
		return strconv.FormatInt(tokens, 10)
	}
	return canondata.T("verdict.no-data")
}

func listOr(empty string, xs []string) string {
	if len(xs) == 0 {
		return empty
	}
	sort.Strings(xs)
	return strings.Join(xs, "; ")
}
