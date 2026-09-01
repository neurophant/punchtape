// Pending amends of expectations: once code has started, the fate of
// PROVEN checks is decided by a human alone — a row green at least
// once tested behavior, weakening it is an assertion. A submission
// changing such a row's expectations (a table row replacement, an update
// of behavior fields, a remove) is not applied — it entirely becomes
// a pending amend of state; the slot carries an assertion block, the
// answer arrives as a separate amend delta kind (approve/reject), silence
// dismisses the amend with a conservative default: proven expectations
// are not weakened by silence. Adding rows is free — an addition
// does not erase red.
//
// Repair is free for rows that proved nothing (a generalization of the
// pin-red principle): a pin-red row (green on a broken surface) and a
// never-green row tested no behavior — there is nothing to weaken, the
// replacement is re-accepted by a run and a probe on the next
// submission. Without free repair, a card with defective or
// unfulfillable rows has no route to the verdict: amendment is a
// human's, quarantine is only for red behavior.
package engine

import (
	"fmt"
	"slices"
	"strings"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/delta"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// applyAmend — a human's assertion of a pending amend: approve
// applies the stored delta in full (the "asserted" origin in
// the journal is the evidence of assertion), reject dismisses. A delta without
// an open amend is an honest one-line refusal.
func (b *builder) applyAmend(am *delta.Amend) error {
	if am == nil {
		return rejection("amend", canondata.T("submit.reject.empty-delta"))
	}
	pending := b.state.PendingAmend
	if pending == nil {
		return rejection("amend", canondata.T("submit.reject.amend-none"))
	}
	switch am.Decision {
	case "approve", "reject":
	default:
		return rejection("amend", canondata.T("submit.reject.amend-decision"))
	}
	b.state.PendingAmend = nil
	if am.Decision == "reject" {
		return nil
	}
	var d delta.Delta
	if err := yamlio.DecodeStrict([]byte(pending.Raw), &d); err != nil {
		return rejection("amend", firstLine(err))
	}
	if d.SubmissionKey == "" {
		d.SubmissionKey = pending.Key
	}
	b.assertedAmend = true
	return b.applyDelta(d)
}

// rowProtected — a row's expectations are protected by human assertion:
// the scenario's check exists, was proven by a green run at least once,
// and is not pin-red. A never-green row proves nothing — its repair
// is free (a generalization of the pin-red principle): a replacement
// is re-accepted by a run and a probe on the next submission; a weak
// replacement stays red or pin-red.
func (b *builder) rowProtected(sc canon.Scenario) bool {
	c, ok := b.checks["TST-"+idNumSuffix(sc.ID)]
	return ok && c.Proven && c.Pin != canon.PinRed
}

// amendTriggerRows — table rows changing the expectations of a protected
// check: a replacement of an existing row (the same seed and the same
// sequence of commands) with different expectations or a different
// declared volatility (when.volatile — a post-code weakening of
// transience is a human assertion too). A proven row is
// protected; one that proved nothing (pin-red, never green) is repaired
// freely. A byte-for-byte identical row does not count as an amend.
// Returns the indexes of the held rows and a line for each.
func (b *builder) amendTriggerRows(ck *delta.Checks) ([]int, []string) {
	existing := map[string]canon.Scenario{}
	for _, sc := range b.scns {
		if k := scenarioKey(sc); k != "" {
			existing[k] = sc
		}
	}
	// The authorial repair keys: a key-matched change to a protected
	// row holds the same channel as an identity-matched one — the key
	// must not become a bypass of the assertion. First holder wins,
	// deterministically by id.
	byKey := map[string]canon.Scenario{}
	for _, id := range sortedIDs(b.scns) {
		sc := b.scns[id]
		if sc.Key != "" {
			if _, taken := byKey[sc.Key]; !taken {
				byKey[sc.Key] = sc
			}
		}
	}
	var idx []int
	var changes []string
	for i, row := range ck.Rows {
		if len(row.Run) == 0 {
			continue
		}
		seed, materials, pre, when, err := b.rowIdentity(row)
		if err != nil {
			continue // broken shape — the normal path will refuse with a number
		}
		sc, ok := existing[rowKey(seed, materials, pre, when)]
		if !ok && row.Key != "" {
			sc, ok = byKey[row.Key]
		}
		if !ok || !b.rowProtected(sc) {
			continue
		}
		then, _, err := rowExpectations(row)
		volatileSame := sc.When != nil && slices.Equal(row.Volatile, sc.When.Volatile)
		if err != nil || (slices.Equal(then, sc.Then) && volatileSame) {
			continue
		}
		runs := make([]string, len(row.Run))
		for j, cmd := range row.Run {
			runs[j] = cmd.YAML()
		}
		idx = append(idx, i)
		changes = append(changes, canondata.T("amend.change.row", canondata.M{
			"scenario": sc.ID,
			"index":    fmt.Sprintf("%d", i+1),
			"run":      strings.Join(runs, " ; "),
		}))
	}
	return idx, changes
}

// opsHaveCapture — whether the operations include a measurement capture:
// the hold lines of such an amend carry the measured bytes themselves.
func opsHaveCapture(ops []delta.AssertionOp) bool {
	for _, op := range ops {
		if op.Capture {
			return true
		}
	}
	return false
}

// captureChangeLines — hold lines for a capture: the human asserts the
// meaning while looking at "was → measured by the machine"; values are
// collapsed into one line with a visible diff, the verbatim bytes go
// into why amend.
func captureChangeLines(sc *canon.Scenario, then []canon.Assertion) []string {
	var lines []string
	for i, fresh := range then {
		if i >= len(sc.Then) || fresh == sc.Then[i] {
			continue
		}
		lines = append(lines, canondata.T("amend.change.capture", canondata.M{
			"scenario":    sc.ID,
			"observation": fresh.Observation,
			"old":         oneLine(sc.Then[i].Value),
			"new":         oneLine(fresh.Value),
		}))
	}
	return lines
}

// oneLine — an observation value as one line: newlines show as
// \n, long values collapse with an explicit ellipsis.
func oneLine(s string) string {
	const limit = 60
	if len(s) > limit {
		s = s[:limit-1] + "…"
	}
	return strings.ReplaceAll(s, "\n", `\n`)
}

// specTriggerOps — indexes of spec operations changing the expectations
// of a protected scenario: updates of behavior fields
// (given/when/then/prose) and a remove of such a row. Protected is the
// scenario that proved behavior; one that proved nothing is repaired
// freely. Summary is cosmetics, not an expectation.
func (b *builder) specTriggerOps(sp *delta.Spec) ([]int, []string) {
	var idx []int
	var changes []string
	for i, op := range sp.Operations {
		switch {
		case op.UpdateScenario != nil:
			id := op.UpdateScenario.ID
			sc, ok := b.scns[id]
			if !ok || !b.rowProtected(sc) {
				continue
			}
			fields := changedBehaviorFields(op.UpdateScenario, sc)
			if fields == "" {
				continue
			}
			idx = append(idx, i)
			changes = append(changes, canondata.T("amend.change.update", canondata.M{
				"scenario": id, "fields": fields,
			}))
		case op.UpdateAssertions != nil:
			id := op.UpdateAssertions.Scenario
			sc, ok := b.scns[id]
			if !ok || !b.rowProtected(sc) {
				continue
			}
			then, err := b.applyAssertionOps(id, &sc, op.UpdateAssertions.Ops)
			if err != nil {
				return nil, nil // broken shape — the normal path will refuse with a number
			}
			if slices.Equal(then, sc.Then) {
				continue
			}
			idx = append(idx, i)
			if opsHaveCapture(op.UpdateAssertions.Ops) {
				changes = append(changes, captureChangeLines(&sc, then)...)
			} else {
				changes = append(changes, canondata.T("amend.change.update", canondata.M{
					"scenario": id, "fields": "then",
				}))
			}
		case op.RemoveScenario != nil:
			id := op.RemoveScenario.ID
			sc, ok := b.scns[id]
			if !ok || !b.rowProtected(sc) {
				continue
			}
			idx = append(idx, i)
			changes = append(changes, canondata.T("amend.change.remove", canondata.M{
				"scenario": id,
			}))
		}
	}
	return idx, changes
}

// partitionSpecOps — the fate of spec operations on deliver: change
// request status awaits replacements/removals of proven rows and
// edits/removals of requirements (the boundaries of the delivered
// product are changed by a human); free — additions and repair of
// rows that proved nothing (parity with the checks table, row by row).
func (b *builder) partitionSpecOps(sp *delta.Spec) (held, free []delta.SpecOperation, changes []string) {
	heldIdx, heldChanges := b.specTriggerOps(sp)
	heldSet := map[int]bool{}
	for _, i := range heldIdx {
		heldSet[i] = true
	}
	for i, op := range sp.Operations {
		triggered := heldSet[i]
		reqOp := op.UpdateRequirement != nil || op.RemoveRequirement != nil
		switch {
		case triggered || reqOp:
			held = append(held, op)
			if reqOp && !triggered {
				if op.UpdateRequirement != nil {
					changes = append(changes, canondata.T("amend.change.update-req", canondata.M{
						"requirement": op.UpdateRequirement.ID,
					}))
				} else {
					changes = append(changes, canondata.T("amend.change.remove-req", canondata.M{
						"requirement": op.RemoveRequirement.ID,
					}))
				}
			}
		default:
			free = append(free, op)
		}
	}
	return held, free, append(changes, heldChanges...)
}

// splitSpecOps — splitting a submission's operations by hold indexes.
func splitSpecOps(ops []delta.SpecOperation, heldIdx []int) (held, free []delta.SpecOperation) {
	heldSet := map[int]bool{}
	for _, i := range heldIdx {
		heldSet[i] = true
	}
	for i, op := range ops {
		if heldSet[i] {
			held = append(held, op)
		} else {
			free = append(free, op)
		}
	}
	return held, free
}

// validateSpecRefs — referential integrity of spec operations at intake:
// update/remove reference what exists, add takes a free ID and
// an existing requirement (from the canon or an addition of the same delta).
// The messages are the same as at application: the hand sees a familiar
// refusal.
func (b *builder) validateSpecRefs(sp *delta.Spec) error {
	reqs := map[string]bool{}
	for id := range b.reqs {
		reqs[id] = true
	}
	scns := map[string]bool{}
	for id := range b.scns {
		scns[id] = true
	}
	for _, op := range sp.Operations {
		switch {
		case op.AddRequirement != nil:
			if !validID("REQ", op.AddRequirement.ID) {
				return rejection(op.AddRequirement.ID, canondata.T("submit.reject.id-req"))
			}
			if reqs[op.AddRequirement.ID] {
				return rejection(op.AddRequirement.ID, canondata.T("submit.reject.id-taken"))
			}
			reqs[op.AddRequirement.ID] = true
		case op.UpdateRequirement != nil:
			if !reqs[op.UpdateRequirement.ID] {
				return rejection(op.UpdateRequirement.ID, canondata.T("submit.reject.not-found"))
			}
		case op.RemoveRequirement != nil:
			if !reqs[op.RemoveRequirement.ID] {
				return rejection(op.RemoveRequirement.ID, canondata.T("submit.reject.not-found"))
			}
		case op.AddScenario != nil:
			if !validID("SCN", op.AddScenario.ID) {
				return rejection(op.AddScenario.ID, canondata.T("submit.reject.id-scn"))
			}
			if scns[op.AddScenario.ID] {
				return rejection(op.AddScenario.ID, canondata.T("submit.reject.id-taken"))
			}
			if !reqs[op.AddScenario.Requirement] {
				return rejection(op.AddScenario.ID, canondata.T("submit.reject.requirement-not-found", canondata.M{
					"id": op.AddScenario.Requirement,
				}))
			}
			scns[op.AddScenario.ID] = true
		case op.UpdateScenario != nil:
			if !scns[op.UpdateScenario.ID] {
				return rejection(op.UpdateScenario.ID, canondata.T("submit.reject.not-found"))
			}
			if op.UpdateScenario.Requirement != nil && !reqs[*op.UpdateScenario.Requirement] {
				return rejection(op.UpdateScenario.ID, canondata.T("submit.reject.requirement-not-found", canondata.M{
					"id": *op.UpdateScenario.Requirement,
				}))
			}
		case op.UpdateAssertions != nil:
			if !scns[op.UpdateAssertions.Scenario] {
				return rejection(op.UpdateAssertions.Scenario, canondata.T("submit.reject.not-found"))
			}
		case op.RemoveScenario != nil:
			if !scns[op.RemoveScenario.ID] {
				return rejection(op.RemoveScenario.ID, canondata.T("submit.reject.not-found"))
			}
			delete(scns, op.RemoveScenario.ID)
		default:
			return rejection("operation", canondata.T("submit.reject.spec-op-empty"))
		}
	}
	return nil
}

// changedBehaviorFields — which behavior fields an update changes; an
// empty string means behavior untouched.
func changedBehaviorFields(up *delta.UpdateScenario, sc canon.Scenario) string {
	var fields []string
	if up.Seed != nil && !slices.Equal(up.Seed, sc.Seed) {
		fields = append(fields, "seed")
	}
	if up.Materials != nil && !slices.Equal(up.Materials, sc.Materials) {
		fields = append(fields, "materials")
	}
	if up.Pre != nil && !slices.EqualFunc(up.Pre, sc.Pre, func(a, b []string) bool { return slices.Equal(a, b) }) {
		fields = append(fields, "pre")
	}
	if up.When != nil && (sc.When == nil || !sameWhen(*up.When, *sc.When)) {
		fields = append(fields, "when")
	}
	if up.Then != nil && !slices.Equal(up.Then, sc.Then) {
		fields = append(fields, "then")
	}
	if up.Prose != nil && (sc.Prose == nil || *up.Prose != *sc.Prose) {
		fields = append(fields, "prose")
	}
	return strings.Join(fields, "/")
}

// sameWhen — equality of surface calls: nil ≡ [] in commands and
// sameness of the other fields (including declared volatility — its
// post-code weakening is a human assertion too).
func sameWhen(a, b canon.When) bool {
	return a.Surface == b.Surface && a.TimeoutSec == b.TimeoutSec &&
		slices.Equal(a.Command, b.Command) && eqStrPtr(a.Stdin, b.Stdin) &&
		slices.Equal(a.Volatile, b.Volatile)
}

func eqStrPtr(a, b *string) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

// stageChangeRequest — a submission on deliver becomes a pending
// transaction: a change request is a bulk destructive operation
// (closes cards, unfreezes contracts, rolls the cycle back to
// acceptance anew); the plan is shown by the slot, it is applied only
// by human assertion, silence dismisses it with a conservative default.
func (b *builder) stageChangeRequest(kind, key, raw string, changes []string) {
	b.stageAmend(kind, key, raw, changes, b.changeRequestPlan())
}

// changeRequestPlan — the machine's change request plan: application
// facts, not judgment.
func (b *builder) changeRequestPlan() []string {
	var plan []string
	if n := len(b.cards); n > 0 {
		plan = append(plan, canondata.T("amend.change.cr-cards", canondata.M{
			"n": fmt.Sprintf("%d", n),
		}))
	}
	frozen := 0
	for _, id := range sortedIDs(b.contracts) {
		if b.contracts[id].Status == canon.IfcFrozen {
			frozen++
		}
	}
	if frozen > 0 {
		plan = append(plan, canondata.T("amend.change.cr-contracts", canondata.M{
			"n": fmt.Sprintf("%d", frozen),
		}))
	}
	return append(plan, canondata.T("amend.change.cr-stage"))
}

// stageAmend — a submission becomes a pending amend: what is asked of
// the human (raw — the delta of the held part verbatim) is not applied,
// the question opens, the stage holds until an answer or silence.
// Staging is the staging position in the journal (hold identity for
// anonymous amend answers), key is the submission's effective key
// (explicit or a machine auto-key, never empty: the refusal of an open
// amend and the slot must both name the amend).
func (b *builder) stageAmend(kind, key, raw string, changes, plan []string) {
	if key == "" {
		key = b.deltaKey
	}
	if raw == "" {
		raw = b.rawInput
	}
	b.state.PendingAmend = &canon.PendingAmend{
		Kind: kind, Key: key, Raw: raw, Staging: b.stagingTx,
		Changes: changes, Plan: plan,
	}
	b.stagedAmend = true
}
