// Waivers of completeness lint rules: a rule holds the stage until
// the human accepts the gap with a reason. A waive is a delta of kind:
// waive {rule, target, reason}: the rule stops blocking the
// transition, the record lives in state (written only by the
// machine), the event — in the ledger, the verdict's opinion queue
// carries the waiver with its reason — it neither stays silent nor
// blocks. Executability and stop conditions are not waivable: the
// registry marks only completeness rules as waivable.
package engine

import (
	"strings"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/delta"
	"github.com/neurophant/punchtape/internal/ledger"
)

// applyWaive — validation and application of a waiver: the rule must
// be in the registry and waivable, the target non-empty, the reason
// mandatory (except for revocation). A repeated waiver of the same
// target replaces the reason.
func (b *builder) applyWaive(w *delta.Waive) error {
	if w == nil {
		return rejection("waive", canondata.T("submit.reject.waive-empty"))
	}
	rule, known := canondata.LintRuleByID(w.Rule)
	if !known || !rule.Waivable {
		return rejection("waive", canondata.T("submit.reject.waive-rule", canondata.M{
			"rule": w.Rule,
		}))
	}
	if strings.TrimSpace(w.Target) == "" {
		return rejection("waive", canondata.T("submit.reject.waive-target"))
	}
	reason := strings.TrimSpace(w.Reason)
	if w.Revoke {
		kept := b.state.Waives[:0]
		for _, rec := range b.state.Waives {
			if rec.Rule == w.Rule && rec.Target == w.Target {
				continue
			}
			kept = append(kept, rec)
		}
		b.state.Waives = kept
		b.e.noteWaive(w.Rule, w.Target, true)
		return nil
	}
	if reason == "" {
		return rejection("waive", canondata.T("submit.reject.waive-reason"))
	}
	kept := b.state.Waives[:0]
	for _, rec := range b.state.Waives {
		if rec.Rule == w.Rule && rec.Target == w.Target {
			continue
		}
		kept = append(kept, rec)
	}
	b.state.Waives = append(kept, canon.WaiveRecord{
		Rule: w.Rule, Target: w.Target, Reason: reason,
	})
	b.e.noteWaive(w.Rule, w.Target, false)
	return nil
}

// familiesWaived — families accepted by the human with a reason:
// coverage consults waivers, a gap closed with a reason does not hold
// code.
func familiesWaived(st canon.State) map[string]bool {
	out := map[string]bool{}
	for _, w := range st.Waives {
		if w.Rule == "coverage.family" {
			out[w.Target] = true
		}
	}
	return out
}

// noteWaive — a machine fact of a waiver into the ledger: rule,
// target and revoked/granted. The reason lives in state (humans read
// it in the verdict), the ledger carries the event.
func (e *Engine) noteWaive(rule, target string, revoked bool) {
	state, err := e.Store.State()
	if err != nil {
		return
	}
	le := ledger.NewEntry("waive", state.Stage)
	le.Details = map[string]string{
		"rule": rule, "target": target,
	}
	if revoked {
		le.Details["revoked"] = "yes"
	}
	_ = e.Ledger.Append(le)
}

// waiveActive — whether a rule's waiver is active for a target: only
// completeness rules are consulted, the machine's stop conditions
// have no waivers.
func (b *builder) waiveActive(rule, target string) bool {
	for _, rec := range b.state.Waives {
		if rec.Rule == rule && rec.Target == target {
			return true
		}
	}
	return false
}

// reqsWithoutScenarios — requirements without a single scenario: the
// material of the spec-complete gate. The rule is validity
// (unavoidable); on submission paths the applySpec invariant holds it;
// here it is defense in depth.
func (b *builder) reqsWithoutScenarios() []string {
	hasScn := map[string]bool{}
	for _, sc := range b.scns {
		hasScn[sc.Requirement] = true
	}
	var out []string
	for id := range b.reqs {
		if !hasScn[id] {
			out = append(out, id)
		}
	}
	sortStrings(out)
	return out
}

// uncoveredFamiliesWaived — uncovered families without a waiver:
// the coverage gate consults waives; a gap closed with a reason does
// not hold code.
func (b *builder) uncoveredFamiliesWaived() []string {
	var out []string
	for _, f := range b.uncoveredFamilies() {
		if b.waiveActive("coverage.family", f) {
			continue
		}
		out = append(out, f)
	}
	return out
}
