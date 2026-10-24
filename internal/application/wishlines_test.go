// The wish's line universe and its coverage map: the numbering of
// the immutable bytes, the map over the claims, and the state's
// brief block that carries both.
package application

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neurophant/punchtape/internal/domain/zone"
)

func TestWishLinesNumberTheNonEmptyInFileOrder(t *testing.T) {
	wish := []byte("\nfirst\n\n  \nsecond\r\n\nthird\n\n")
	lines := wishLines(wish)
	if strings.Join(lines, "|") != "first|second|third" {
		t.Fatalf("the universe is the non-empty lines in file order: %q", lines)
	}
	if all := wishLines([]byte("only one line")); len(all) != 1 || all[0] != "only one line" {
		t.Fatalf("a wish without blanks numbers whole: %q", all)
	}
	if empty := wishLines(nil); len(empty) != 0 {
		t.Fatalf("an empty wish numbers nothing: %q", empty)
	}
}

func TestWishCoverageMap(t *testing.T) {
	universe := []string{"one", "two", "three", "four", "five"}
	// An empty registry: nothing is claimed, every line is a hole.
	claimed, holes := wishCoverage(universe, nil)
	if claimed != 0 || len(holes) != 5 || holes[0] != 1 || holes[4] != 5 {
		t.Fatalf("no claims must leave the whole universe open: %d %v", claimed, holes)
	}
	// One range over everything: the whole brief is accounted.
	if claimed, holes = wishCoverage(universe, []string{"W1..W5"}); claimed != 5 || len(holes) != 0 {
		t.Fatalf("one full range must claim the universe: %d %v", claimed, holes)
	}
	// Two requirements overlapping: the union claims, no double count.
	if claimed, holes = wishCoverage(universe, []string{"W1..W3", "W3..W5"}); claimed != 5 || len(holes) != 0 {
		t.Fatalf("overlaps claim once: %d %v", claimed, holes)
	}
	// Numbers outside the universe claim nothing; one broken spelling
	// voids the whole claim list — the map only counts what parses.
	if claimed, holes = wishCoverage(universe, []string{"W1..W2", "W9"}); claimed != 2 || len(holes) != 3 || holes[0] != 3 {
		t.Fatalf("outside numbers claim nothing: %d %v", claimed, holes)
	}
	if claimed, holes = wishCoverage(universe, []string{"W1..W2", "junk"}); claimed != 0 || len(holes) != 5 {
		t.Fatalf("a broken spelling voids the claims: %d %v", claimed, holes)
	}
}

func TestStateCarriesTheBriefMap(t *testing.T) {
	mountPorts()
	root := t.TempDir()
	// A wish of three lines with blanks between them: the state names
	// the wish's object and maps the whole brief open, the holes
	// spelled in the document's own grammar.
	writeZone(t, root, zone.WishSlot, "first\n\nsecond\n\nthird\n")
	runIn(t, root)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "brief: 0/3 lines; requirements 0 standing, 0 covered; not-a-requirement: 0") {
		t.Fatalf("the brief block must carry the counts:\n%s", state)
	}
	if !strings.Contains(state, "uncovered brief lines: W1..W3") {
		t.Fatalf("the open brief must spell its holes in the document's grammar:\n%s", state)
	}
	if !strings.Contains(state, "wish at .punchtape/journal/objects/") {
		t.Fatalf("the goal must name the wish's object in the zone:\n%s", state)
	}
}

func TestAMarkerMustQuoteTheWishItsOwnWords(t *testing.T) {
	mountPorts()
	root := t.TempDir()
	writeZone(t, root, zone.WishSlot, "a heading\n\ndo the work\n")
	runIn(t, root)
	// A marker whose quote is not the line's own text: refused by
	// the named field, the inbox stays.
	writeZone(t, root, zone.InboxSlot, `not-a-requirement:
    - lines: [W2]
      quote: not the line
      reason: framing
`)
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "not-a-requirement[0].quote") || !strings.Contains(report, `want "do the work"`) {
		t.Fatalf("a lying quote must be refused with the wish's own words:\n%s", report)
	}
	if _, err := os.Stat(filepath.Join(root, zone.Dirname, zone.InboxSlot)); err != nil {
		t.Fatalf("the refused marker keeps the inbox")
	}
	// The honest quote: the mark stands and the state shows the
	// dismissed line's own words beside the reason.
	writeZone(t, root, zone.InboxSlot, `not-a-requirement:
    - lines: [W1]
      quote: a heading
      reason: the title frames, it demands nothing
`)
	runIn(t, root)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, `not-a-requirement W1: "a heading" — the title frames, it demands nothing`) {
		t.Fatalf("the standing mark must show its line's words and reason, state:\n%s", state)
	}
}
