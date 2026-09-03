// Verification factory: gates are executed by the machine's run
// transaction after the agent's delta is applied. The agent knows
// about gates only as a list in the slot; executing a gate by the
// agent is forbidden.
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
	"github.com/neurophant/punchtape/internal/checks"
	"github.com/neurophant/punchtape/internal/compiler"
	"github.com/neurophant/punchtape/internal/journal"
	"github.com/neurophant/punchtape/internal/ledger"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// The repair attempts budget per card: exhausted — quarantine, there
// are no eternal loops.

// Ceremony rent: an advisory gate that has not changed outcomes for
// at least three runs (all green) is retired — a ceremony pays rent
// by its influence, not by habit. Stop-condition gates (suite, trace,
// review-diff) are not retireable: they compute readiness.

// rentEligible — gates participating in rent: advisory static checks
// and the probe; their red outcome is caught further down the
// pipeline (surface build, the red list), so retiring does not blind
// the stop condition.
var rentEligible = map[string]bool{
	"types": true, "lint": true, "coverage-probe": true,
}

// retireGates retires gates that have not influenced outcomes: at
// least gateRentRuns runs in the ledger, all green. Retiring is a
// transaction: recorded in state and in the ledger.
func (e *Engine) retireGates(b *builder) {
	// Recipe rent window: types/lint evidence is valid only in the
	// era of the current recipe — a conventions change (apply
	// kind=conventions) opens a new window; green runs of the previous
	// recipe prove no rent for the new one.
	convEra := -1
	entries := e.Ledger.All()
	for i, en := range entries {
		if en.Event == "apply" && en.Details["kind"] == "conventions" {
			convEra = i
		}
	}
	runs, reds := map[string]int{}, map[string]int{}
	for i, en := range entries {
		if en.Event != "gate" {
			continue
		}
		name := en.Details["name"]
		if name == "" {
			continue
		}
		if convEra >= 0 && i < convEra && (name == "types" || name == "lint") {
			continue
		}
		runs[name]++
		if en.Details["outcome"] != "green" {
			reds[name]++
		}
	}
	retired := map[string]bool{}
	for _, name := range b.state.RetiredGates {
		retired[name] = true
	}
	for name := range rentEligible {
		if retired[name] || runs[name] < canondata.Limit("gate.rent-runs") || reds[name] > 0 {
			continue
		}
		retired[name] = true
		b.state.RetiredGates = append(b.state.RetiredGates, name)
		le := ledger.NewEntry("gate-retire", b.state.Stage)
		le.Details = map[string]string{"name": name, "runs": fmt.Sprintf("%d", runs[name])}
		_ = e.Ledger.Append(le)
	}
}

// gateRetired — the gate is retired by rent in the instance state.
func gateRetired(state canon.State, name string) bool {
	for _, g := range state.RetiredGates {
		if g == name {
			return true
		}
	}
	return false
}

// Delta kind names on the machine side.
const (
	deltaKindRun         = "run" // machine transaction that runs the gates
	deltaKindCode        = "code"
	deltaKindFix         = "fix"
	deltaKindSpec        = "spec"
	deltaKindChecks      = "checks"
	deltaKindAssert      = "assert"
	deltaKindProbe       = "probe"
	deltaKindConventions = "conventions"
	deltaKindAmend       = "amend"
)

// gateOutcome — one gate's outcome for the ledger.
type gateOutcome struct {
	name, reason string
	green        bool
}

// gateTx — the machine run transaction.
type gateTx struct {
	parent string          // parent transaction (the submission that caused the runs)
	gates  []gateOutcome   // gate outcomes
	ran    map[string]bool // check ids this transaction actually ran
}

// ranMark — record a check id as run in this transaction (the scope
// marker's witness: a red check NOT in the set is standing, not
// this submission's verdict).
func (g *gateTx) ranMark(id string) {
	if g.ran == nil {
		g.ran = map[string]bool{}
	}
	g.ran[id] = true
}

func (g *gateTx) outcome(name string, green bool, reason string) {
	g.gates = append(g.gates, gateOutcome{name: name, green: green, reason: reason})
}

// compileChecks rebuilds the check set after a spec change: the
// touched is regenerated, the untouched is not rewritten, the
// orphaned is deleted. Fresh and changed checks are remembered for
// the run — only the touched is re-checked, not the whole instance.
func (b *builder) compileChecks() error {
	scns := make([]canon.Scenario, 0, len(b.scns))
	for _, id := range sortedIDs(b.scns) {
		scns = append(scns, b.scns[id])
	}
	existing := make([]canon.Check, 0, len(b.checks))
	for _, id := range sortedIDs(b.checks) {
		existing = append(existing, b.checks[id])
	}
	plan, err := compiler.Sync(scns, existing)
	if err != nil {
		return rejection("spec", err.Error())
	}
	// Spec invariant: one run — one set of expectations. Two scenarios
	// with the same seed and command sequence but different
	// expectations are unexecutable by construction — caught before
	// code.
	if cols := runCollisions(b.scns); len(cols) > 0 {
		return rejection("spec", canondata.T("submit.reject.run-collision", canondata.M{
			"rows": strings.Join(cols, "; "),
		}))
	}
	// Change estimate: touched scenarios and regenerated checks are
	// counted here — before the recompile, not after.
	touchedScns := 0
	for _, id := range b.touched {
		if _, ok := b.scns[id]; ok {
			touchedScns++
		}
	}
	b.state.Impact = &canon.Impact{
		ScenariosTouched:  touchedScns,
		ChecksRegenerated: len(plan.Put),
		ChecksKept:        len(plan.Keep),
		ChecksRemoved:     len(plan.Remove),
	}
	for _, c := range plan.Put {
		if err := b.putCheck(c); err != nil {
			return err
		}
		b.freshChecks = append(b.freshChecks, c.ID)
	}
	for _, id := range plan.Remove {
		delete(b.checks, id)
		b.effects = append(b.effects, journal.Effect{
			Path: canonFile("checks", id), Delete: true,
		})
	}
	return nil
}

func (b *builder) putCheck(c canon.Check) error {
	data, err := yamlio.Marshal(c)
	if err != nil {
		return fmt.Errorf("system failure: %w", err)
	}
	b.checks[c.ID] = c
	b.effects = append(b.effects, journal.Effect{
		Path: canonFile("checks", c.ID), Content: machineFile(data),
	})
	return nil
}

// runAfterSubmit decides which gates to run after a submission and
// returns a one-line summary for the reply to the agent: the cycle's
// outcome is visible from the reply itself, no separate next is
// needed for the outcome. chargeReds is the set of behaviorally red
// checks before a code/fix submission: only the code/fix gates may
// charge budget attempts, the other paths get nil and do not touch
// the counter.
func (e *Engine) runAfterSubmit(kind, cardID, parentTx string, freshChecks []string, chargeReds map[string]bool) (string, error) {
	switch kind {
	case deltaKindCode, deltaKindFix:
		return e.gatesForCode(cardID, parentTx, chargeReds)
	case deltaKindConventions:
		// The recipe arrived for finished code (requested over a red
		// surface) — the gates of all cards restart at once:
		// conventions without a run prove nothing, a card without its
		// own run would keep a stale red "does not build", and at
		// deliver all cards are green — re-proving them is the
		// restoration of trust after invalidation.
		if !e.codeStarted() {
			return "", nil
		}
		cards, err := e.Store.Cards()
		if err != nil {
			return "", err
		}
		summary := ""
		for _, c := range cards {
			if c.Status == canon.CardQuarantine || !e.cardHasRunChecks(c) {
				// a card without a single run — there has been no code yet:
				// running its gates would paint "does not build" on work
				// that has not started
				continue
			}
			one, err := e.gatesForCode(c.ID, parentTx, nil)
			if err != nil {
				return "", err
			}
			if summary == "" || one == summary {
				summary = one
				continue
			}
			summary = summary + "; " + one
		}
		if summary != "" {
			summary = canondata.T("gates.conventions-reproof") + "\n" + summary
		}
		return summary, nil
	case deltaKindSpec, deltaKindChecks, deltaKindAssert, deltaKindAmend:
		if len(freshChecks) == 0 {
			// Sticky red: a red outcome lives in the canon until a
			// live run; background rehearsal does not update the canon.
			// A spec submission with no fresh lines while checks are
			// red triggers a card replay — a fresh outcome, not a
			// wasted code submission.
			if cardID := e.cardWithRedChecks(); cardID != "" {
				return e.gatesForCode(cardID, parentTx, nil)
			}
			return "", nil
		}
		return e.gatesForFreshChecks(freshChecks, parentTx)
	case deltaKindProbe:
		// a probe is an observation, not gates: the machine has nothing
		// to check, it has something to offer with the next slot
		st, err := e.Store.State()
		if err != nil {
			return "", nil
		}
		det, failed := 0, 0
		for _, p := range st.Probes {
			switch {
			case p.Failed:
				failed++
			case p.Deterministic:
				det++
			}
		}
		return canondata.T("gates.probed", canondata.M{
			"kept":          strconv.Itoa(len(st.Probes)),
			"deterministic": strconv.Itoa(det),
			"not":           strconv.Itoa(len(st.Probes) - det - failed),
			"failed":        strconv.Itoa(failed),
		}), nil
	default:
		return "", nil
	}
}

// runCheckOut — a check run in normal mode.
func runCheckNormal(e *Engine, check canon.Check) checks.Result {
	return e.runCheck(check, checks.Normal)
}

// gatesForCode — the full gate set after code/fix: types/lint by
// conventions recipes, suite for the card through the surface,
// coverage-probe, trace, review in both directions; quarantine on an
// exhausted attempts budget. chargeReds — behaviorally red lines
// before the submission (nil on all paths except code/fix): an
// attempt is charged after the run and only to a surviving red.
func (e *Engine) gatesForCode(cardID, parentTx string, chargeReds map[string]bool) (string, error) {
	g := &gateTx{parent: parentTx}
	b := newBuilder(e, mustState(e))
	if err := b.load(); err != nil {
		return "", err
	}
	card, ok := b.cards[cardID]
	if !ok {
		return "", nil
	}

	// Ceremony rent: retirement is computed before this transaction's
	// run, retired gates are not run.
	e.retireGates(b)

	// types and lint are static gates by conventions recipes; building
	// is their concern, suite answers only for runs through the
	// surface. Red static reasons become card blockers: they are not
	// visible in the red check list, the slot must carry them itself.
	// The honest working time of recipes (build + static gates) is the
	// same run work as for checks: counted in runs-ms, the latency
	// line subtracts it from the submission wall (the budget counts
	// overhead only).
	gateStarted := time.Now()
	static := e.staticGates(b, card)
	e.lastSubmitRunsMs += time.Since(gateStarted).Milliseconds()
	redLines := []string{}
	blockers := []string{}
	typesRed := false
	for _, oc := range static {
		if gateRetired(b.state, oc.Name) {
			continue
		}
		g.outcome(oc.Name, oc.Green || !oc.Applied, oc.Reason)
		if oc.Applied && !oc.Green {
			line := fmt.Sprintf("%s: %s", oc.Name, oc.Reason)
			redLines = append(redLines, line)
			blockers = append(blockers, line)
			if oc.Name == "types" {
				typesRed = true
			}
		}
	}

	// Outcome freshness: running a surface that failed statics or the
	// build is pointless — its outcomes are unreliable. Red types or
	// a recipe failure leaves the card's checks "not run" (a run
	// refusal, not behavior): old reds are not inherited, such a run
	// eats no behavioral budget.
	suiteReady := !typesRed
	if suiteReady {
		if err := e.buildSurfaces(b, card); err != nil {
			line := canondata.T("gates.surface-build", canondata.M{"err": err.Error()})
			redLines = append(redLines, line)
			blockers = append(blockers, line)
			suiteReady = false
		}
	}

	// suite: run the card's checks through the surface. A finished
	// background rehearsal report of the same digest replaces the run —
	// the agent has already seen these outcomes in a previous
	// submission's reply, there is no point repeating the race; the
	// background work's wall time goes into the ledger by the machine.
	allChecks := make([]canon.Check, 0, len(b.checks))
	for _, id := range sortedIDs(b.checks) {
		allChecks = append(allChecks, b.checks[id])
	}
	var replay *BgReport
	if suiteReady {
		if d, ok := surfaceDigest(e, allChecks); ok {
			replay, _ = LoadBgReport(e.Workdir, d)
		}
	}
	// A live run takes precedence over rehearsal: the stop marker —
	// before the first acquisition of the run lock (the rehearsal
	// finishes the current check and exits, waiting for its end is not
	// required).
	stopRehearsal(e.Workdir)
	greenChecks := []canon.Check{}
	suiteRed := []string{}
	for _, scnID := range card.Scenarios {
		check, ok := b.checks["TST-"+idNumSuffix(scnID)]
		if !ok {
			continue
		}
		if !suiteReady {
			// The run did not happen: the "not run" mark is a run-level
			// refusal (env), not behavior. Old reds are not inherited:
			// this line's outcome is about this submission.
			cause := canondata.T("gates.not-run-cause-types")
			if !typesRed {
				cause = canondata.T("gates.not-run-cause-build")
			}
			reason := canondata.T("gates.suite-not-run", canondata.M{"cause": cause})
			check.Outcome, check.Reason = canon.CheckRed, reason
			check.EnvFail = true
			redLines = append(redLines, fmt.Sprintf("%s %s: %s", check.ID, check.Scenario, reason))
			if err := b.putCheck(check); err != nil {
				return "", err
			}
			continue
		}
		var res checks.Result
		started := time.Now()
		// replayed or live — either way THIS transaction ran the check
		g.ranMark(check.ID)
		if replay != nil {
			green, reason, env, hit := replay.bgReplayOutcome(check)
			if hit {
				res = checks.Result{CheckID: check.ID, Green: green, Reason: reason, Env: env}
			} else {
				res = runCheckNormal(e, check)
			}
		} else {
			res = runCheckNormal(e, check)
		}
		e.lastSubmitRunsMs += time.Since(started).Milliseconds()
		if res.Green {
			check.Outcome, check.Reason = canon.CheckGreen, ""
			check.Proven = true
			check.EnvFail = false
			greenChecks = append(greenChecks, check)
		} else {
			check.Outcome, check.Reason = canon.CheckRed, res.Reason
			check.EnvFail = res.Env
			redLines = append(redLines, fmt.Sprintf("%s %s: %s", check.ID, check.Scenario, res.Reason))
			if !res.Env {
				suiteRed = append(suiteRed, fmt.Sprintf("%s %s: %s", check.ID, check.Scenario, res.Reason))
			}
		}
		if err := b.putCheck(check); err != nil {
			return "", err
		}
	}
	// Counting after the run: an attempt is a barren repair cycle. The
	// submission changed files (chargeReds is set only for code/fix)
	// and the suite actually ran — the attempt grows only if a red
	// line that lived before the submission stayed red and behavioral.
	// It went green or was replaced by another line — progress: the
	// counter resets; "three in a row" barren repairs is a truth about
	// the code, not about the environment.
	if chargeReds != nil && suiteReady {
		survived := false
		for _, scnID := range card.Scenarios {
			check, ok := b.checks["TST-"+idNumSuffix(scnID)]
			if ok && chargeReds[check.ID] && check.Outcome == canon.CheckRed && !check.EnvFail {
				survived = true
				break
			}
		}
		if survived {
			card.Attempts++
		} else {
			card.Attempts = 0
		}
	}
	suiteGreen := len(redLines) == 0 && len(greenChecks) > 0
	g.outcome("suite", suiteGreen, strings.Join(redLines, "; "))

	// coverage-probe: a green check must go red on a broken surface,
	// otherwise it does not run the declared path. A probe retired by
	// rent is not run. The probe outcome is written into every check:
	// a pin-red line proves nothing — repairing it (replacement,
	// edit, deletion) is free, without human assertion.
	probeRed := []string{}
	if !gateRetired(b.state, "coverage-probe") {
		for _, check := range greenChecks {
			started := time.Now()
			res := e.runCheck(check, checks.ProbeBrokenSurface)
			e.lastSubmitRunsMs += time.Since(started).Milliseconds()
			if res.Green {
				check.Pin = canon.PinRed
				probeRed = append(probeRed, canondata.T("gates.probe-red", canondata.M{
					"id": check.ID, "scn": check.Scenario,
				}))
			} else {
				check.Pin = canon.PinGreen
			}
			if err := b.putCheck(check); err != nil {
				return "", err
			}
		}
	}
	g.outcome("coverage-probe", len(probeRed) == 0, strings.Join(probeRed, "; "))
	redLines = append(redLines, probeRed...)
	e.lastSubmitChecks = len(card.Scenarios) + len(greenChecks)

	// trace: every executable scenario linked to a check.
	scns := make([]canon.Scenario, 0, len(b.scns))
	for _, id := range sortedIDs(b.scns) {
		scns = append(scns, b.scns[id])
	}
	missing := missingLinksLocal(scns, b.checks)
	g.outcome("trace", len(missing) == 0, canondata.T("gates.missing-links", canondata.M{
		"links": strings.Join(missing, ", "),
	}))

	// review-diff: declared-but-not-done (reds) and
	// done-but-not-declared (fictitious greenness — from the probe).
	g.outcome("review-diff", len(redLines) == 0, strings.Join(redLines, "; "))

	typesLintGreen := !typesRed
	for _, oc := range static {
		if oc.Applied && !oc.Green {
			typesLintGreen = false
		}
	}
	allGreen := suiteGreen && typesLintGreen && len(probeRed) == 0
	card.Blockers = blockers
	switch {
	case allGreen:
		card.Status = canon.CardGreen
	case card.Attempts >= canondata.Limit("card.attempts-budget") && len(suiteRed) > 0:
		// Quarantine — only on a red behavioral finale: the attempts
		// budget guards behavioral repair cycles. Weak pins
		// (coverage-probe), statics and build with an exhausted budget
		// keep the card in progress — they are fixed by the spec
		// (update-scenario, free) and clean code, not by new behavior.
		card.Status = canon.CardQuarantine
		card.Gap = e.gapClass(b, card)
		ql := ledger.NewEntry("quarantine", b.state.Stage)
		ql.Details = map[string]string{
			"card": card.ID, "gap": card.Gap,
			"attempts": fmt.Sprintf("%d", card.Attempts),
		}
		_ = e.Ledger.Append(ql)
	default:
		card.Status = canon.CardRed
	}
	if err := b.putCard(card); err != nil {
		return "", err
	}
	// Convergence of parallel facets: at the moment the last card goes
	// green, the machine runs the full suite of all cards — a
	// neighbor's edit that broke someone else's behavior reddens the
	// card owning the scenario. One card — nothing to converge, no
	// ceremony.
	summary := summarizeGates(g, b)
	if replay != nil {
		summary += "; " + replay.bgSummaryLine()
		le := ledger.NewEntry("gate", mustState(e).Stage)
		le.WallMs = replay.WallMs
		green := true
		for _, c := range replay.Checks {
			if c.Outcome != canon.CheckGreen {
				green = false
				break
			}
		}
		le.Details = map[string]string{
			"name": "suite", "scope": "background", "outcome": boolGreen(green),
			"checks": fmt.Sprintf("%d", len(replay.Checks)),
		}
		_ = e.Ledger.Append(le)
	}
	if allGreen && len(b.cards) > 1 {
		if reds := e.convergenceRun(b, cardID); len(reds) > 0 {
			summary += "; " + canondata.T("gates.convergence-red", canondata.M{
				"reds": strings.Join(reds, "; "),
			})
		}
	}
	// Acceptance moment: all cards closed — the full check set is
	// finalized on this instance state, the digest signs it.
	e.acceptanceNote(b)
	// Passport: machine trust transitions in the same transaction —
	// the anchor is the run transaction's hash, the same one that
	// goes into the journal.
	gateTx := yamlio.Digest([]byte("gate:" + parentTx))
	if err := b.passportProof(card, gateTx, allGreen); err != nil {
		return "", err
	}
	if err := e.commitGates(b, g, parentTx); err != nil {
		return "", err
	}
	return summary, nil
}

// convergenceRun — the full acceptance suite over all cards except the
// freshly run one (its checks are just green on the current code). A
// red check of another card returns it to work: the status honestly
// goes red, the red list lands in its slot. The run is a machine
// ledger fact with scope=convergence.
func (e *Engine) convergenceRun(b *builder, skipCard string) []string {
	var reds []string
	ran := 0
	for _, cardID := range sortedIDs(b.cards) {
		if cardID == skipCard {
			continue
		}
		card := b.cards[cardID]
		for _, scnID := range card.Scenarios {
			check, ok := b.checks["TST-"+idNumSuffix(scnID)]
			if !ok || check.Outcome != canon.CheckGreen {
				continue
			}
			ran++
			started := time.Now()
			res := e.runCheck(check, checks.Normal)
			e.lastSubmitRunsMs += time.Since(started).Milliseconds()
			if res.Green {
				continue
			}
			check.Outcome, check.Reason = canon.CheckRed, res.Reason
			check.EnvFail = res.Env
			if err := b.putCheck(check); err != nil {
				continue
			}
			reds = append(reds, fmt.Sprintf("%s %s: %s", check.ID, check.Scenario, res.Reason))
		}
	}
	for _, cardID := range sortedIDs(b.cards) {
		if cardID == skipCard {
			continue
		}
		card := b.cards[cardID]
		if card.Status != canon.CardGreen {
			continue
		}
		for _, scnID := range card.Scenarios {
			if c, ok := b.checks["TST-"+idNumSuffix(scnID)]; ok && c.Outcome == canon.CheckRed {
				card.Status = canon.CardRed
				_ = b.putCard(card)
				break
			}
		}
	}
	e.lastSubmitChecks += ran
	le := ledger.NewEntry("gate", mustState(e).Stage)
	le.Details = map[string]string{
		"name": "suite", "scope": "convergence", "outcome": boolGreen(len(reds) == 0),
	}
	if len(reds) > 0 {
		le.Details["reason"] = firstLineStr(strings.Join(reds, "; "))
	}
	_ = e.Ledger.Append(le)
	return reds
}

// summarizeGates — a one-line summary of the run transaction:
// name=outcome in fixed order. This line is the gates' entire trace
// in the agent's context; the details live in the ledger.
func summarizeGates(g *gateTx, b *builder) string {
	byName := map[string]gateOutcome{}
	for _, o := range g.gates {
		byName[o.name] = o
	}
	// The scope marker: the suite verdict names THIS transaction's
	// runs; red checks the transaction never touched are standing
	// reds, and the headline must not read as "all green" while they
	// stand (the reply said suite=green next to an untouched
	// red).
	standing := 0
	if b != nil {
		for id, c := range b.checks {
			if c.Outcome == canon.CheckRed && !g.ran[id] {
				standing++
			}
		}
	}
	order := []string{"types", "lint", "suite", "coverage-probe", "trace", "review-diff"}
	parts := make([]string, 0, len(order))
	for _, name := range order {
		o, ok := byName[name]
		if !ok {
			continue
		}
		verdict := "red"
		if o.green {
			verdict = "green"
		}
		if name == "suite" && standing > 0 {
			verdict += canondata.T("submit.reply.gates-scope", canondata.M{
				"count": fmt.Sprintf("%d", standing),
			})
		}
		parts = append(parts, name+"="+verdict)
	}
	return strings.Join(parts, " ")
}

// gatesForFreshChecks — run only the fresh/changed checks after a
// spec edit: re-checking is proportional to the edit, not to the
// instance.
func (e *Engine) gatesForFreshChecks(ids []string, parentTx string) (string, error) {
	g := &gateTx{parent: parentTx}
	b := newBuilder(e, mustState(e))
	if err := b.load(); err != nil {
		return "", err
	}
	// A live run takes precedence over rehearsal (see gatesForCode).
	stopRehearsal(e.Workdir)
	var reds []string
	probeRed := []string{}
	for _, id := range ids {
		check, ok := b.checks[id]
		if !ok {
			continue
		}
		g.ranMark(id)
		started := time.Now()
		res := e.runCheck(check, checks.Normal)
		e.lastSubmitRunsMs += time.Since(started).Milliseconds()
		if res.Green {
			check.Outcome, check.Reason = canon.CheckGreen, ""
			check.Proven = true
			check.EnvFail = false
		} else {
			check.Outcome, check.Reason = canon.CheckRed, res.Reason
			check.EnvFail = res.Env
			reds = append(reds, fmt.Sprintf("%s %s: %s", check.ID, check.Scenario, res.Reason))
		}
		if err := b.putCheck(check); err != nil {
			return "", err
		}
	}
	g.outcome("suite", len(reds) == 0, strings.Join(reds, "; "))

	// The probe of fresh and replaced checks: a replaced line must
	// prove the path no worse than the original — the pin is
	// re-checked on every fresh run line; without this the verdict
	// would see a non-proving line replace another without a
	// re-check.
	if !gateRetired(b.state, "coverage-probe") {
		for _, id := range ids {
			check, ok := b.checks[id]
			if !ok || check.Outcome != canon.CheckGreen {
				continue
			}
			started := time.Now()
			res := e.runCheck(check, checks.ProbeBrokenSurface)
			e.lastSubmitRunsMs += time.Since(started).Milliseconds()
			if res.Green {
				check.Pin = canon.PinRed
				probeRed = append(probeRed, canondata.T("gates.probe-red", canondata.M{
					"id": check.ID, "scn": check.Scenario,
				}))
			} else {
				check.Pin = canon.PinGreen
			}
			if err := b.putCheck(check); err != nil {
				return "", err
			}
		}
	}
	g.outcome("coverage-probe", len(probeRed) == 0, strings.Join(probeRed, "; "))

	// A red fresh check reopens the owning card: the card's status is
	// a projection of its checks' outcomes; a green card with a red
	// check is a desync.
	ownerOf := map[string]string{}
	for _, cardID := range sortedIDs(b.cards) {
		for _, scnID := range b.cards[cardID].Scenarios {
			ownerOf[scnID] = cardID
		}
	}
	for _, id := range ids {
		check, ok := b.checks[id]
		if !ok || check.Outcome != canon.CheckRed {
			continue
		}
		cardID, ok := ownerOf[check.Scenario]
		if !ok {
			continue
		}
		card := b.cards[cardID]
		if card.Status == canon.CardGreen {
			card.Status = canon.CardRed
			if err := b.putCard(card); err != nil {
				return "", err
			}
		}
	}
	// Mirror image: a red card without red checks and static blockers
	// is green — status is a projection of outcomes, and a spec repair
	// with correct behavior closes the cycle without a wasted code
	// submission. A pin-red check does not hold a card green: the
	// line proves nothing, the probe must be green. An active card
	// goes green on fully checked green coverage: unchecked work does
	// not close itself (pure growth at deliver goes this way — the
	// additions are run, the outcome projection closes the card).
	for _, cardID := range sortedIDs(b.cards) {
		card := b.cards[cardID]
		if card.Status != canon.CardRed && card.Status != canon.CardActive {
			continue
		}
		if card.Status == canon.CardActive && len(card.Scenarios) == 0 {
			continue
		}
		red := false
		seen := 0
		for _, scnID := range card.Scenarios {
			if c, ok := b.checks["TST-"+idNumSuffix(scnID)]; ok {
				seen++
				if c.Outcome == canon.CheckRed || c.Pin == canon.PinRed {
					red = true
					break
				}
			}
		}
		if card.Status == canon.CardActive && seen != len(card.Scenarios) {
			continue
		}
		if !red && len(card.Blockers) == 0 {
			card.Status = canon.CardGreen
			if err := b.putCard(card); err != nil {
				return "", err
			}
		}
	}

	// trace after every delta: an executable scenario linked to a check.
	scns := make([]canon.Scenario, 0, len(b.scns))
	for _, id := range sortedIDs(b.scns) {
		scns = append(scns, b.scns[id])
	}
	missing := missingLinksLocal(scns, b.checks)
	g.outcome("trace", len(missing) == 0, canondata.T("gates.missing-links", canondata.M{
		"links": strings.Join(missing, ", "),
	}))
	e.lastSubmitChecks = len(ids)
	// The acceptance moment happens here too: a spec edit that closed
	// all cards finalizes the set on the current instance.
	e.acceptanceNote(b)
	if err := e.commitGates(b, g, parentTx); err != nil {
		return "", err
	}
	return summarizeGates(g, b), nil
}

// cardWithRedChecks — the first card with a red check: a red outcome
// in the canon requires a fresh run (status is a projection of
// outcomes). Empty string — there are no reds.
func (e *Engine) cardWithRedChecks() string {
	cards, err := e.Store.Cards()
	if err != nil {
		return ""
	}
	chk, err := e.Store.Checks()
	if err != nil {
		return ""
	}
	redByScn := map[string]bool{}
	for _, c := range chk {
		if c.Outcome == canon.CheckRed {
			redByScn[c.Scenario] = true
		}
	}
	for _, card := range cards {
		for _, scnID := range card.Scenarios {
			if redByScn[scnID] {
				return card.ID
			}
		}
	}
	return ""
}

// reopenAutorun — auto re-verification of a rolled-back cycle: a
// change that touched neither the surface nor the check set since the
// acceptance moment (conflict closure, wish withdrawal) requires no
// code resubmission — the machine runs the gates with its own run
// transaction. Drift (the set changed) returns an empty string: a
// new cycle is the agent's work.
func (e *Engine) reopenAutorun(parentTx string) (string, error) {
	if drifted, _, _ := e.acceptanceDrift(); drifted {
		return "", nil
	}
	cards, err := e.Store.Cards()
	if err != nil {
		return "", nil
	}
	for _, c := range cards {
		if c.Status == canon.CardActive || c.Status == canon.CardRed {
			return e.gatesForCode(c.ID, parentTx, nil)
		}
	}
	return "", nil
}

// missingLinksLocal — connectivity over the builder's projections.
func missingLinksLocal(scns []canon.Scenario, checks map[string]canon.Check) []string {
	linked := map[string]bool{}
	for _, c := range checks {
		linked[c.Scenario] = true
	}
	var out []string
	for _, sc := range scns {
		if compiler.DegreeOf(sc).Level == compiler.DegreeExecutable && !linked[sc.ID] {
			out = append(out, sc.ID)
		}
	}
	return out
}

// buildSurfaces brings the surface to readiness: the operator-declared
// artifact (file or tree — the operator decides the surface kind, the
// machine executes the recipe once for the whole surface), or, without
// a recipe, direct submission — a file named after the command,
// submitted by a delta; the right to execute is a property of the
// record.
func (e *Engine) buildSurfaces(b *builder, card canon.Card) error {
	names := map[string]bool{}
	for _, scnID := range card.Scenarios {
		sc, ok := b.scns[scnID]
		if ok && sc.When != nil && len(sc.When.Command) > 0 {
			names[sc.When.Command[0]] = true
		}
	}
	sorted := make([]string, 0, len(names))
	for name := range names {
		sorted = append(sorted, name)
	}
	sort.Strings(sorted)
	if len(sorted) == 0 {
		return nil
	}
	conv, err := e.Store.Conventions()
	if err != nil {
		return err
	}
	if conv != nil && conv.Build != nil {
		// One recipe for the whole surface: the artifact is declared,
		// command names resolve inside it at run time. The recipe's
		// working time is honest verification work: counted in
		// runs-ms.
		shown := conv.Build.Out
		if shown == "" {
			shown = sorted[0]
		}
		buildStarted := time.Now()
		buildErr := e.runBuildRecipe(conv.Build, sorted[0])
		e.lastSubmitRunsMs += time.Since(buildStarted).Milliseconds()
		if buildErr != nil {
			return fmt.Errorf("%s", canondata.T("gates.build", canondata.M{
				"name": shown, "err": oneLineInfo(buildErr.Error()),
			}))
		}
		return nil
	}
	for _, name := range sorted {
		out := filepath.Join(e.Workdir, filepath.FromSlash(name))
		// The right to execute is a property of the record, not the
		// code: the agent-submitted surface runs without a manual
		// chmod after every submission (the input copy in the run
		// directory is executable anyway).
		if info, err := os.Stat(out); err == nil && !info.IsDir() && info.Mode()&0o111 == 0 {
			if err := os.Chmod(out, info.Mode()|0o111); err != nil {
				return fmt.Errorf("chmod %s: %s", name, err)
			}
		}
	}
	return nil
}

// surfaceArtifact — the operator-declared surface artifact from
// conventions (file or directory, a path in the project); an empty
// string — no artifact declared: the surface is submitted directly, a
// file named after the command.
// runCheck — a gate/probe run with the env advice routed by the
// delivered-code state: at zero code "install the tool" misleads
// (nothing is delivered to install against); the advice says
// deliver the code.
func (e *Engine) runCheck(check canon.Check, mode checks.Mode) checks.Result {
	if e.codeStarted() {
		return checks.Run(e.Workdir, e.surfaceArtifact(), check, mode)
	}
	return checks.RunNoCode(e.Workdir, e.surfaceArtifact(), check, mode)
}

func (e *Engine) surfaceArtifact() string {
	conv, err := e.Store.Conventions()
	if err != nil || conv == nil || conv.Build == nil {
		return ""
	}
	return conv.Build.Out
}

// staticGates — types and lint by conventions recipes: an undeclared
// recipe is a "not applied" gate (a soft empty-start status). A gate
// retired by rent is not executed at all and does not enter the
// outcomes — its red does not exist. Command placeholders are the
// same as in the build: {name} is the card's surface name, {out} is
// the artifact path (gate recipe out → build out → surface name).
func (e *Engine) staticGates(b *builder, card canon.Card) []toolOutcome {
	conv, err := e.Store.Conventions()
	if err != nil || conv == nil {
		return []toolOutcome{{Name: "types", Green: true}, {Name: "lint", Green: true}}
	}
	timeout := canondata.Limit("conventions.gate-timeout-s")
	name := e.surfaceName(b, card)
	gateOut := func(rec *canon.Recipe) string {
		outRel := ""
		switch {
		case rec != nil && rec.Out != "":
			outRel = rec.Out
		case conv.Build != nil && conv.Build.Out != "":
			outRel = conv.Build.Out
		default:
			outRel = name
		}
		return filepath.Join(e.Workdir, filepath.FromSlash(outRel))
	}
	res := []toolOutcome{}
	for _, g := range []struct {
		n   string
		rec *canon.Recipe
	}{{"types", conv.Types}, {"lint", conv.Lint}} {
		if gateRetired(b.state, g.n) {
			continue
		}
		res = append(res, gateFromRecipe(e.Workdir, g.n, name, g.rec, gateOut(g.rec), timeout))
	}
	return res
}

// surfaceName — the card's surface name: command[0] of its scenarios
// (with several, the first alphabetically, as in surface building);
// without commands — the declared artifact; otherwise the neutral
// "out".
func (e *Engine) surfaceName(b *builder, card canon.Card) string {
	names := map[string]bool{}
	for _, scnID := range card.Scenarios {
		if sc, ok := b.scns[scnID]; ok && sc.When != nil && len(sc.When.Command) > 0 {
			names[sc.When.Command[0]] = true
		}
	}
	sorted := make([]string, 0, len(names))
	for n := range names {
		sorted = append(sorted, n)
	}
	sort.Strings(sorted)
	if len(sorted) > 0 {
		return sorted[0]
	}
	if a := e.surfaceArtifact(); a != "" {
		return a
	}
	return "out"
}

// journalEntryRun — the skeleton of a service run's machine
// transaction.
func journalEntryRun(submissionKey string) journal.Entry {
	return journal.Entry{
		Transaction:   yamlio.Digest([]byte("run:" + submissionKey)),
		SubmissionKey: submissionKey,
		DeltaKind:     deltaKindRun,
		RecordedAt:    time.Now().UTC(),
	}
}

// commitGates commits the run transaction: effects, journal, ledger,
// stage advancement.
func (e *Engine) commitGates(b *builder, g *gateTx, parentTx string) error {
	b.advance()
	if b.processErr != nil {
		return b.processErr
	}
	entry := journal.Entry{
		Transaction:   yamlio.Digest([]byte("gate:" + parentTx)),
		SubmissionKey: "gate-" + shortHex(parentTx),
		DeltaKind:     deltaKindRun,
		WallMs:        0,
		RecordedAt:    time.Now().UTC(),
		Effects:       b.effects,
	}
	if err := e.commit(entry); err != nil {
		return fmt.Errorf("gate transaction: %w", err)
	}
	stage := mustState(e).Stage
	for _, o := range g.gates {
		le := ledger.NewEntry("gate", stage)
		le.Details = map[string]string{"name": o.name, "outcome": boolGreen(o.green)}
		if !o.green && o.reason != "" {
			le.Details["reason"] = firstLineStr(o.reason)
		}
		if err := e.Ledger.Append(le); err != nil {
			return err
		}
	}
	return nil
}

func boolGreen(b bool) string {
	if b {
		return "green"
	}
	return "red"
}

func firstLineStr(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

func oneLineInfo(s string) string {
	s = strings.ReplaceAll(s, "\n", "; ")
	if len(s) > 400 {
		s = s[:400] + "…"
	}
	return s
}

func mustState(e *Engine) canon.State {
	st, err := e.Store.State()
	if err != nil {
		return canon.State{}
	}
	return st
}

// idNumSuffix returns the numeric tail of an identifier: SCN-001 → 001.
func idNumSuffix(id string) string {
	for i := 0; i < len(id); i++ {
		if id[i] == '-' {
			return id[i+1:]
		}
	}
	return id
}

// runCollisions — a provable contradiction of lines: two scenarios
// with the same run (seed + commands) that have incompatible
// expectations, or a self-contradiction of a single set. Different
// observation angles of one run (stdout and a state file) are not a
// contradiction — both can hold together. Line identity
// (seed+materials+commands) already distinguishes different seeds: an
// honest "one command, different seeds" pair does not become a
// collision; the verdict names the COMMAND and the OBSERVATION of the
// contradiction — canon identifiers are not visible to the agent, a
// blind "SCN-009" is not fixable.
func runCollisions(scns map[string]canon.Scenario) []string {
	byRun := map[string][]string{}
	for _, id := range sortedIDs(scns) {
		key := scenarioKey(scns[id])
		if key == "" {
			continue
		}
		byRun[key] = append(byRun[key], id)
	}
	var out []string
	for _, key := range sortedIDs(byRun) {
		ids := byRun[key]
		for _, id := range ids {
			if target := selfContradictionOn(scns[id].Then); target != "" {
				out = append(out, collisionLine(scns[id],
					canondata.T("checktable.collision.self", canondata.M{"target": target})))
			}
		}
		for i, id := range ids {
			for _, other := range ids[i+1:] {
				if target := setsContradictionOn(scns[id].Then, scns[other].Then); target != "" {
					out = append(out, collisionLine(scns[id],
						canondata.T("checktable.collision.cross", canondata.M{"target": target})))
				}
			}
		}
	}
	return out
}

// collisionLine — the verdict line for the agent: the run command and
// the essence of the contradiction; the line is recognized by the
// command, not by an internal ID.
func collisionLine(sc canon.Scenario, reason string) string {
	return canondata.T("checktable.collision.run", canondata.M{
		"when": strings.Join(sc.When.Command, " "), "reason": reason,
	})
}

// contradictionTarget — the observation and path of a pair of
// incompatible exact expectations (values are not printed: the
// observation and the command are enough to find the line and see
// both values in the canon).
func contradictionTarget(x, y canon.Assertion) string {
	target := x.Observation
	if x.Path != "" {
		target += " " + x.Path
	}
	return target
}

// exactConditions — exact equality conditions: different values with
// matching observation/condition/path are incompatible. Containment
// and existence are compatible with anything — there is no provable
// contradiction, the machine does not guess.
var exactConditions = map[string]bool{
	"equals": true, "bytes-equals": true, "json-equals": true,
}

// setsContradictionOn — the observation of an incompatible pair of
// expectations from two sets of one run; an empty string — no
// contradiction.
func setsContradictionOn(a, b []canon.Assertion) string {
	for _, x := range a {
		if !exactConditions[x.Condition] {
			continue
		}
		for _, y := range b {
			if x.Observation == y.Observation && x.Condition == y.Condition &&
				x.Path == y.Path && x.Value != y.Value {
				return contradictionTarget(x, y)
			}
		}
	}
	return ""
}

// selfContradictionOn — the observation of a set's self-contradiction:
// two incompatible exact expectations inside one scenario.
func selfContradictionOn(then []canon.Assertion) string {
	for i, x := range then {
		if !exactConditions[x.Condition] {
			continue
		}
		for _, y := range then[i+1:] {
			if x.Observation == y.Observation && x.Condition == y.Condition &&
				x.Path == y.Path && x.Value != y.Value {
				return contradictionTarget(x, y)
			}
		}
	}
	return ""
}

// gapClass — deterministic classification of the quarantine gap from
// machine facts: canon — a detectable contradiction of spec lines,
// environment — all of the card's red checks are environment/launch
// refusals (not behavior: the machine classified them by refusal
// source), code — everything else. The honesty boundary: canon
// unexecutability without a contradiction is machine-indistinguishable
// from a code error — "code" means "everything the machine can check
// has converged; the rest is the executor". A missing surface/recipe
// is the environment class: it does not end in quarantine (such reds
// eat no behavior budget), an honest red with a diagnosis and FIX
// stays red.
func (e *Engine) gapClass(b *builder, card canon.Card) string {
	if len(runCollisions(b.scns)) > 0 {
		return canon.GapCanon
	}
	red, envRed := 0, 0
	for _, scnID := range card.Scenarios {
		check, ok := b.checks["TST-"+idNumSuffix(scnID)]
		if !ok || check.Outcome != canon.CheckRed {
			continue
		}
		red++
		if check.EnvFail {
			envRed++
		}
	}
	if red > 0 && red == envRed {
		return canon.GapEnvironment
	}
	return canon.GapCode
}

// acceptanceNote — the acceptance moment: a transaction closed all
// cards, the full check set is finalized on a specific instance
// state. The digest of the surface and of all check definitions is
// the moment's signature: verdict freezing reconciles the current
// instance against it, on divergence the verdict is not frozen.
func (e *Engine) acceptanceNote(b *builder) {
	if len(b.cards) == 0 {
		return
	}
	for _, c := range b.cards {
		if c.Status != canon.CardGreen && c.Status != canon.CardQuarantine {
			return
		}
	}
	all := make([]canon.Check, 0, len(b.checks))
	for _, id := range sortedIDs(b.checks) {
		all = append(all, b.checks[id])
	}
	green, red := 0, 0
	for _, c := range all {
		switch c.Outcome {
		case canon.CheckGreen:
			green++
		case canon.CheckRed:
			red++
		}
	}
	digest, files, skipped, surfaceOK := surfaceSignature(e, all)
	le := ledger.NewEntry("gate", b.state.Stage)
	le.Details = map[string]string{
		"name": "suite", "scope": "acceptance",
		"outcome": boolGreen(red == 0 && len(all) > 0 && green == len(all)),
		"digest":  shortHex(digest),
		"checks":  fmt.Sprintf("%d", len(all)),
		"green":   fmt.Sprintf("%d", green),
		"red":     fmt.Sprintf("%d", red),
	}
	// An acceptance that could not see the surface (no verifiable
	// signature at the moment) is named in the ledger: an empty digest
	// is a measurement gap, never a side to compare against.
	if !surfaceOK {
		le.Details["surface"] = "absent"
	}
	// The signature's exact terms ride along: the files it stands on
	// and the command names that signed nothing
	// (interpreter/system-tool rows whose argv named no project
	// file) — the acceptance fact carries its own measurement gap.
	if len(skipped) > 0 {
		sort.Strings(skipped)
		le.Details["surface-skipped"] = strings.Join(skipped, ",")
	}
	if len(files) > 0 {
		sort.Strings(files)
		le.Details["surface-files"] = strings.Join(files, ",")
	}
	// Env doctor before acceptance: a probe of the recipes'
	// environment. A missing tool reddens acceptance with a FIX hint —
	// acceptance does not pass silently on an environment where the
	// recipes cannot run; a green probe is recorded as a fact (env-ok),
	// no silence about it.
	if missing := e.envDoctor(); len(missing) > 0 {
		le.Details["outcome"] = "red"
		le.Details["env"] = strings.Join(missing, ",")
	} else {
		le.Details["env-ok"] = "yes"
	}
	_ = e.Ledger.Append(le)
}
