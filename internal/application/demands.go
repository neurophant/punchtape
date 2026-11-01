package application

// The demand stream: the zone file demands.md, open forever. The
// stream's bytes are the machine's one source of the human's words;
// every non-empty line numbers D1..Dn in file order, derived anew
// from the bytes on every run. The journal's chain holds the past:
// every run recomputes the chain over the current bytes, and a past
// line that moved is zone damage. The map — which lines the standing
// corpus's ranges claim — is one function for every reader: the
// state's block and the verdict's gate answer it, never two truths.

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/neurophant/punchtape/internal/domain"
	"github.com/neurophant/punchtape/internal/domain/doc"
	"github.com/neurophant/punchtape/internal/domain/zone"
	"github.com/neurophant/punchtape/internal/ports"
)

// errDamagedPast — the chain's own verdict: the stream's journaled
// past does not recompose from the current bytes.
var errDamagedPast = errors.New("the demand stream's past changed")

// chainOver — one link of the chain: the sha256 of the previous
// link's own bytes and the new lines joined by a newline. The
// genesis previous is the empty string.
func chainOver(prev string, lines []string) string {
	return domain.Digest(append([]byte(prev), []byte(strings.Join(lines, "\n"))...))
}

// checkChain — the chain law, every run's first step: the stream is
// read, its lines numbered as the run's universe, and the journal's
// whole chain recomputed over them, link by link from the genesis. A
// missing stream file on a born instance, a shrunken past or a link
// that does not recompose is the same damage.
func (m *machine) checkChain() error {
	data, err := ports.FSBind.ReadFile(filepath.Join(m.zoneDir, zone.DemandsSlot))
	if err != nil {
		return errDamagedPast
	}
	m.universe = doc.NonEmptyLines(string(data))
	prev, at := "", 0
	for _, link := range m.standing.ChainAppends {
		if at+link.Lines > len(m.universe) {
			return errDamagedPast
		}
		next := chainOver(prev, m.universe[at:at+link.Lines])
		if next != link.Chain {
			return errDamagedPast
		}
		prev, at = next, at+link.Lines
	}
	return nil
}

// demandTail — the stream's unjournaled lines: the run's universe
// past the chain's journaled past — the lines the next submission's
// demands must carry byte-for-byte.
func (m *machine) demandTail() []string {
	if m.standing.PastLines >= len(m.universe) {
		return nil
	}
	return m.universe[m.standing.PastLines:]
}

// demandCoverage — the universe against the claimed ranges: how many
// lines stand claimed, and which stand unclaimed, ascending. A
// spelling that does not parse, or a number outside the universe,
// claims nothing — the lines it names stay honest holes.
func demandCoverage(universe []string, ranges []string) (claimed int, holes []int) {
	held := map[int]bool{}
	if nums, err := doc.ParseRanges(ranges); err == nil {
		for _, n := range nums {
			if n >= 1 && n <= len(universe) {
				held[n] = true
			}
		}
	}
	for n := 1; n <= len(universe); n++ {
		if !held[n] {
			holes = append(holes, n)
		}
	}
	return len(universe) - len(holes), holes
}

// demandClaims — the registry's ranges over the stream's lines:
// every standing requirement's span and every marker's span.
func demandClaims(r doc.Registry) []string {
	var out []string
	for _, id := range r.Order {
		out = append(out, r.Requirements[id].Lines...)
	}
	for _, marker := range r.Markers {
		out = append(out, marker.Lines...)
	}
	return out
}

// unresolvedLines — the stream's lines without a declared outcome
// or without the standing carrier of their declared outcome, each
// with its reason: "no outcome" or "<outcome> pending". Answer,
// execute and ask lines never pend: their carriers are merge-by-line
// records that only ever replace in place, and a line once asked
// stays asked forever — the closure law answers it. The one account
// the verdict's gate, the state's line and the trace's nodes read.
func (m *machine) unresolvedLines() []string {
	reqSpan := map[int]bool{}
	for _, id := range m.standing.Registry.Order {
		if nums, err := doc.ParseRanges(m.standing.Registry.Requirements[id].Lines); err == nil {
			for _, n := range nums {
				reqSpan[n] = true
			}
		}
	}
	markerSpan := map[int]bool{}
	for _, marker := range m.standing.Registry.Markers {
		if nums, err := doc.ParseRanges(marker.Lines); err == nil {
			for _, n := range nums {
				markerSpan[n] = true
			}
		}
	}
	var out []string
	for n := 1; n <= len(m.universe); n++ {
		lo, ok := m.standing.LineOutcomes[n]
		if !ok {
			out = append(out, fmt.Sprintf("%s (no outcome)", doc.DemandID(n)))
			continue
		}
		pending := false
		switch lo.Outcome {
		case doc.OutcomeWork, doc.OutcomeDescribe:
			pending = !reqSpan[n]
		case doc.OutcomeMark:
			pending = !markerSpan[n]
		}
		if pending {
			out = append(out, fmt.Sprintf("%s (%s pending)", doc.DemandID(n), lo.Outcome))
		}
	}
	return out
}

// orderLine — the order's demand line number, zero when unparsable.
func orderLine(order doc.OpenOrder) int {
	nums, err := doc.ParseRanges([]string{order.Demand})
	if err != nil || len(nums) == 0 {
		return 0
	}
	return nums[0]
}

// briefMap — the stream's whole accounting in one value: the line
// universe's size, the claimed and the open lines, the standing and
// the covered requirements, the markers. The state's block prints
// it; the counts are plain arithmetic over the same map the
// verdict's gate reads.
type briefMap struct {
	Lines        int
	Claimed      int
	Holes        []int
	Requirements int
	Covered      int
	Markers      int
}

// briefAccounting — the brief map of the standing corpus: the run's
// numbered universe against the registry's claims, and the
// requirements named in the covers of the corpus's behaviors.
func (m *machine) briefAccounting(corpus doc.Document) briefMap {
	universe := m.universe
	claimed, holes := demandCoverage(universe, demandClaims(m.standing.Registry))
	covers := map[string]bool{}
	for _, b := range corpus.Behaviors {
		for _, c := range b.Covers {
			covers[c] = true
		}
	}
	out := briefMap{
		Lines:        len(universe),
		Claimed:      claimed,
		Holes:        holes,
		Requirements: len(m.standing.Registry.Order),
		Markers:      len(m.standing.Registry.Markers),
	}
	for _, id := range m.standing.Registry.Order {
		if covers[id] {
			out.Covered++
		}
	}
	return out
}
