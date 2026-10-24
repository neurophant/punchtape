package application

// The wish's line universe and its coverage map. The wish's bytes
// are fixed at the birth; every non-empty line numbers W1..Wm in
// file order, forever — the numbering is always derived anew from
// the immutable bytes, so it cannot drift. The map — which lines the
// standing corpus's ranges claim — is one function for every reader:
// the state's block and the verdict's gate answer it, never two
// truths.

import (
	"strings"

	"github.com/neurophant/punchtape/internal/domain/doc"
)

// wishLines — the wish bytes as the numbered universe: the non-empty
// lines, trimmed, in file order.
func wishLines(wish []byte) []string {
	var out []string
	for _, line := range strings.Split(string(wish), "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

// wishCoverage — the universe against the claimed ranges: how many
// lines stand claimed, and which stand unclaimed, ascending. A
// spelling that does not parse, or a number outside the universe,
// claims nothing — the lines it names stay honest holes.
func wishCoverage(universe []string, ranges []string) (claimed int, holes []int) {
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

// wishClaims — the registry's ranges over the wish's lines: every
// standing requirement's span and every marker's span.
func wishClaims(r doc.Registry) []string {
	var out []string
	for _, id := range r.Order {
		out = append(out, r.Requirements[id].Brief...)
	}
	for _, marker := range r.Markers {
		out = append(out, marker.Lines...)
	}
	return out
}

// briefMap — the brief's whole accounting in one value: the line
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

// briefAccounting — the brief map of the standing corpus: the wish's
// universe against the registry's claims, and the requirements named
// in the covers of the corpus's behaviors. The second answer is
// false when the wish's object does not read — no map over a missing
// brief.
func (m *machine) briefAccounting(corpus doc.Document) (briefMap, bool) {
	wish, err := m.objects.Get(m.standing.GoalAddr)
	if err != nil {
		return briefMap{}, false
	}
	universe := wishLines(wish)
	claimed, holes := wishCoverage(universe, wishClaims(m.standing.Registry))
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
	return out, true
}
