// Contour gates: mandatory check points — the data (contour.yaml), the
// mechanics — the process by name. Freshness before work is the point
// whose mechanics are visible to the human: next before work reconciles
// the passport boundaries with the current instance; drift or a missing
// scope file — a CONTOUR line in the slot and a ledger event; a green
// contour stays silent (zero attention cost). The other points exist
// through process mechanics: boundary-before-code — the scenario gate;
// sync-after-changes — passport synchronization on every spec
// transaction; a person's decision — the rc contract of choice; facts
// and pains — ledger events with the diary.
package engine

import (
	"fmt"
	"sort"
	"strings"

	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/ledger"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// driftedBoundaries — IDs of stale boundaries in a stable order: the
// passport verdict against the current scope-hash and conventions
// digest; a missing scope file is also drift (fail-safe).
func (e *Engine) driftedBoundaries() []string {
	boundaries, err := e.Store.Boundaries()
	if err != nil || len(boundaries) == 0 {
		return nil
	}
	cards, _ := e.Store.Cards()
	scns, _ := e.Store.Scenarios()
	scope, _ := e.boundaryScope(cards, scns)
	convDigest := e.Store.ConventionsDigest()
	var out []string
	for _, bd := range boundaries {
		if bd.Verdict == nil {
			continue
		}
		hash, missing := e.scopeHash(scope[bd.ID])
		if bd.Verdict.ScopeHash != hash || bd.Verdict.ConventionsDigest != convDigest || len(missing) > 0 {
			out = append(out, bd.ID)
		}
	}
	sort.Strings(out)
	return out
}

// contourBeforeWork — the CONTOUR line of the slot when boundaries
// drift: work does not start silently on a stale passport; a green
// contour renders no line. Returns the line (empty — fresh) and the
// list of stale boundaries for the ledger.
func (e *Engine) contourBeforeWork() (string, []string) {
	drifted := e.driftedBoundaries()
	if len(drifted) == 0 {
		return "", nil
	}
	return canondata.T("slot.contour.drift", canondata.M{
		"ids": strings.Join(drifted, ", "),
	}), drifted
}

// boundaryScopeKey — digest of the scope files of all boundaries with
// their IDs: a participant of the next cache key. false — there are no
// boundaries, nothing to invalidate.
func (e *Engine) boundaryScopeKey() (string, bool) {
	boundaries, err := e.Store.Boundaries()
	if err != nil || len(boundaries) == 0 {
		return "", false
	}
	cards, _ := e.Store.Cards()
	scns, _ := e.Store.Scenarios()
	scope, _ := e.boundaryScope(cards, scns)
	var sb strings.Builder
	for _, bd := range boundaries {
		hash, _ := e.scopeHash(scope[bd.ID])
		fmt.Fprintf(&sb, "%s:%s\n", bd.ID, hash)
	}
	return shortHex(yamlio.Digest([]byte(sb.String()))), true
}

// noteContourDrift — a contour machine fact into the ledger:
// freshness-before-work found drift. Written only by a fresh next
// render (the replay cache does not write): the event is an
// observation, not a call counter.
func (e *Engine) noteContourDrift(drifted []string) {
	if len(drifted) == 0 {
		return
	}
	state, err := e.Store.State()
	if err != nil {
		return
	}
	le := ledger.NewEntry("contour", state.Stage)
	le.Details = map[string]string{
		"name": "freshness-before-work", "outcome": "red",
		"boundaries": strings.Join(drifted, ","),
	}
	_ = e.Ledger.Append(le)
}
