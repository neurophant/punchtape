// The stage machine: transitions only by the machine and only by a machine check.
package engine

import (
	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/compiler"
	"github.com/neurophant/punchtape/internal/journal"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// advance moves the stage as far as the machine checks allow:
// the runner executes the process data (scenarios.yaml), every loop
// here is part of the same transaction.
func (b *builder) advance() {
	b.runProcess()
	b.refreshCounters()
	b.putState()
}

// specComplete — the spec is complete: there are requirements, there
// are scenarios, every requirement has a scenario (or an honest
// waiver with a reason — completeness with an open gap is lawful, the
// machine shows it). The "at least one scenario" floor is strict: a
// spec of waivers alone does not open code.
func (b *builder) specComplete() bool {
	if len(b.reqs) == 0 || len(b.scns) == 0 {
		return false
	}
	return len(b.reqsWithoutScenarios()) == 0
}

// scenariosReady — all scenarios have reached a final degree: prose
// (the unmeasurable remainder) is allowed, no skeletons remain, there
// is at least one executable, and every command family of the wish is
// covered by an executable scenario (the coverage gate: a leaky spec
// does not enter code; a gap accepted by the human with a reason is
// not a hole). The degree is computed by the compiler — the single
// point of truth.
func (b *builder) scenariosReady() bool {
	executable := 0
	for _, sc := range b.scns {
		d := compiler.DegreeOf(sc)
		if d.Level == compiler.DegreeSkeleton {
			return false
		}
		if d.Level == compiler.DegreeExecutable {
			executable++
		}
	}
	if executable == 0 {
		return false
	}
	return len(b.uncoveredFamiliesWaived()) == 0
}

// enterImplement creates the implementation card on first entry into
// the stage: one card for the whole product, sequential work.
func (b *builder) enterImplement() {
	if len(b.cards) > 0 {
		return
	}
	ids := make([]string, 0, len(b.scns))
	for _, sc := range b.scns {
		if sc.Prose == nil {
			ids = append(ids, sc.ID)
		}
	}
	sortStrings(ids)
	card := canon.Card{
		ID:        "CRD-001",
		Facet:     "whole-product",
		Scenarios: ids,
		Status:    canon.CardActive,
	}
	// An error is impossible here: the card is local, marshal is deterministic.
	_ = b.putCard(card)
}

// cardGreen — the machine check of the exit from implement: every
// card is closed — green or quarantine (quarantine honestly reaches
// the verdict); unclosed work does not pass the transition.
func (b *builder) cardGreen() bool {
	if len(b.cards) == 0 {
		return false
	}
	for _, c := range b.cards {
		if c.Status != canon.CardGreen && c.Status != canon.CardQuarantine {
			return false
		}
	}
	return true
}

// cardsReopened — a reopened card on delivery returns the cycle to
// repair: a verdict on unready cards is a deadlock without action.
func (b *builder) cardsReopened() bool {
	return len(b.cards) > 0 && !b.cardGreen()
}

// refreshCounters pulls up identifier counters from projections:
// the next free number is always greater than any taken one.
func (b *builder) refreshCounters() {
	if b.state.Counters == nil {
		b.state.Counters = map[string]int{}
	}
	nextOf := func(kind, prefix string, ids map[string]bool) {
		next := b.state.Counters[kind]
		for id := range ids {
			next = maxInt(next, digitsOf(id, prefix)+1)
		}
		if next < 1 {
			next = 1
		}
		b.state.Counters[kind] = next
	}
	reqIDs := map[string]bool{}
	for id := range b.reqs {
		reqIDs[id] = true
	}
	scnIDs := map[string]bool{}
	for id := range b.scns {
		scnIDs[id] = true
	}
	crdIDs := map[string]bool{}
	for id := range b.cards {
		crdIDs[id] = true
	}
	ifcIDs := map[string]bool{}
	for id := range b.contracts {
		ifcIDs[id] = true
	}
	nextOf("REQ", "REQ-", reqIDs)
	nextOf("SCN", "SCN-", scnIDs)
	nextOf("CRD", "CRD-", crdIDs)
	nextOf("IFC", "IFC-", ifcIDs)
	if _, ok := b.state.Counters["TST"]; !ok {
		b.state.Counters["TST"] = 1
	}
}

// putState records the final state of the transaction as an effect.
func (b *builder) putState() {
	// Touched — what the last submission touched: slices and estimates
	// count on it, it accumulates on the builder side. Empty means
	// absence
	//; the manifest header rides inside the effect.
	b.state.Touched = b.touched
	data, err := yamlio.Marshal(b.state)
	if err != nil {
		return // state marshal does not fail: the fields are simple
	}
	b.effects = append(b.effects, journal.Effect{
		Path:    stateFilePath,
		Content: canon.MachineHeader("next, submit") + "\n" + string(data),
	})
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// digitsOf reads the numeric tail of an identifier after the prefix;
// unreadable — zero.
func digitsOf(id, prefix string) int {
	n := 0
	rest := id[len(prefix):]
	for _, r := range rest {
		if r < '0' || r > '9' {
			return 0
		}
		n = n*10 + int(r-'0')
	}
	return n
}

func sortStrings(xs []string) {
	for i := 1; i < len(xs); i++ {
		for j := i; j > 0 && xs[j] < xs[j-1]; j-- {
			xs[j], xs[j-1] = xs[j-1], xs[j]
		}
	}
}
