// The remaining branches' battery: the empty-PATH launch failure, the
// row deadline, the double birth's single winner, the close-only
// marker, the unfindable row bytes, the small helpers' whole
// behavior — every line the case batteries never crossed.
package application

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neurophant/punchtape/internal/domain/doc"
	"github.com/neurophant/punchtape/internal/domain/fprint"
	"github.com/neurophant/punchtape/internal/domain/journal"
	"github.com/neurophant/punchtape/internal/domain/zone"
)

func TestEmptyPathFailsTheLaunchByName(t *testing.T) {
	mountPorts()
	saved := os.Getenv("PATH")
	t.Setenv("PATH", "")
	// The build's launch failure carries its own words.
	root := newBornRoot(t)
	writeZone(t, root, zone.InboxSlot, `
work: a build that cannot launch
build: "./no-such-tool x"
behaviors:
    - id: tool/x
      name: X
      prose: P.
      covers: [R1]
      rows:
        - run: cat src/in.txt > out.txt
          expect:
            files:
                out.txt: ok
`)
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "build: RED — launch:") {
		t.Fatalf("the launch failure must be named, report:\n%s", report)
	}
	_ = saved
}

func TestRowDeadlineCutsTheRun(t *testing.T) {
	mountPorts()
	root := newRoot(t)
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.EnvSlot, "PUNCHTAPE_ROW_TIMEOUT=1\n")
	writeZone(t, root, zone.InboxSlot, `
work: the slow row under a one-second pin
behaviors:
    - id: tool/slow
      name: Slow
      prose: P.
      claims:
        - src/in.txt:1-1
      rows:
        - run: sleep 30
          expect:
            rc: 0
            files:
                out.txt: ok
`)
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "timed out after 1s — the deadline defaults to 300 s; PUNCHTAPE_ROW_TIMEOUT names whole seconds") {
		t.Fatalf("the pinned deadline must cut, name itself and teach the default and the pin's whole law, report:\n%s", report)
	}
}

// TestInstructionsWithoutBehaviors — the corpus-empty instruction
// belongs to the empty corpus alone: when registry parts stand and
// no behavior does, the terminal names the missing proof and never
// calls a standing corpus empty.
func TestInstructionsWithoutBehaviors(t *testing.T) {
	mountPorts()
	// The truly empty corpus — birth keeps its first-document teaching.
	fresh := t.TempDir()
	writeZone(t, fresh, zone.DemandsSlot, "a wish")
	if term := runIn(t, fresh); !strings.Contains(term, "corpus empty — submit the first document") {
		t.Fatalf("the empty corpus must keep its first-document instruction, term %q", term)
	}
	// A standing registry without behaviors — the truth, not "empty".
	root := newBornRoot(t)
	writeZone(t, root, zone.InboxSlot, "work: retiring the only behavior\nmandate: [{line: D1, target: tool/copy}]\nremove: [tool/copy]\n")
	term := runIn(t, root)
	if !strings.Contains(term, "no behaviors stand — submit to .punchtape/inbox.yaml a behavior with a pinned row") {
		t.Fatalf("the behaviorless corpus with standing parts must name the missing proof, term %q", term)
	}
	if strings.Contains(term, "corpus empty") {
		t.Fatalf("a corpus with standing requirements is not empty, term %q", term)
	}
}

func TestSecondBirthLosesGracefully(t *testing.T) {
	mountPorts()
	root := t.TempDir()
	writeZone(t, root, zone.DemandsSlot, "one goal")
	if err := birth(filepath.Join(root, zone.Dirname), []byte("one goal")); err != nil {
		t.Fatalf("the first birth wins: %v", err)
	}
	// The journal already exists: the second birth sees a living
	// instance and answers nothing wrong.
	if err := birth(filepath.Join(root, zone.Dirname), []byte("one goal")); err != nil {
		t.Fatalf("the lost race must answer nil: %v", err)
	}
}

func TestCloseOnlyMarkerRefusesLoudly(t *testing.T) {
	root := newBornRoot(t)
	path := filepath.Join(root, agentsFile)
	// The closing marker alone: no guess, no touch, one loud line.
	os.WriteFile(path, []byte("words\n"+blockClose+"\nmore\n"), 0o644)
	var term bytes.Buffer
	if code := Run(root, &term); code != 1 || !strings.Contains(term.String(), "marker without its pair") {
		t.Fatalf("a lone closing marker must refuse loudly: code %d term %q", code, term.String())
	}
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), "more") {
		t.Fatal("a lone closing marker must leave the file alone")
	}
}

func TestLessonWithUnfindableRowBytes(t *testing.T) {
	root := newBornRoot(t)
	// The replacement row under its mandate pair — the pin grown by
	// its exit, the genesis intent already declared — its objects
	// land with its first acceptance.
	heal := strings.Replace(bornDocFor(t, root),
		"intents:\n    - {lines: [D1], outcome: work, prose: \"the tool and its check, told again\"}\n", "", 1)
	heal = strings.Replace(heal,
		"work: the first document of the copy\n",
		"work: the first document of the copy\nmandate: [{line: D1, target: tool/copy}]\n", 1)
	heal = strings.Replace(heal, "          expect:\n", "          expect:\n            rc: 0\n", 1)
	writeZone(t, root, zone.InboxSlot, heal)
	runIn(t, root)
	// A deferred order moves the standing past the replacement's own
	// shape: the resubmission below is a real replacement, not a
	// stale copy of the corpus.
	writeZone(t, root, zone.InboxSlot, `
work: a question deferred
open:
    - {demand: D1, question: which unit?, state: deferred, reason: the human said later}
`)
	runIn(t, root)
	// Red by the tree's own move: the row's expectation fails on the
	// changed bytes.
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("NOPE"), 0o644); err != nil {
		t.Fatal(err)
	}
	runIn(t, root)
	// The heal restores the bytes and resubmits with the objects
	// unreadable: the lesson's red spelling falls back to the
	// digest's name, the green still lands. The heal document's own
	// bytes were journaled by its first submission — every object it
	// names already stands.
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, heal)
	code, _ := runCode(t, root, map[string]bool{"read:objects!after-write": true})
	if code != 0 {
		t.Fatalf("the heal must still answer: %d", code)
	}
	knowledge := zoneFile(t, root, zone.KnowledgeFile)
	// The fold's own row bytes index answers the lesson without one
	// read of the behavior objects: the red and the green spell
	// their true rows.
	if !strings.Contains(knowledge, "run: ") || strings.Contains(knowledge, "<row ") {
		t.Fatalf("the lesson must spell the row's own bytes, knowledge:\n%s", knowledge)
	}
}

func TestSmallHelpersWhole(t *testing.T) {
	// The refusal list's own lines.
	list := refusalList{
		{Where: "a", Law: "law:document", Fact: "one"},
		{Where: "b", Law: "law:paths", Fact: "two"},
	}
	if msg := list.Error(); !strings.Contains(msg, "one") || !strings.Contains(msg, "two") || strings.Count(msg, "\n") != 1 {
		t.Fatalf("one line per refusal: %q", msg)
	}
	// The short digest of a short string.
	if shortDigest("abc") != "abc" {
		t.Fatal("a short digest stands whole")
	}
	// A noteless instance counts zero.
	m := &machine{standing: &journal.Standing{}}
	if len(m.standing.Notes) != 0 {
		t.Fatal("no submissions carry no notes")
	}
	// The argv command's spelling.
	if commandOf(doc.Cmd{Argv: []string{"./tool", "x"}}) != "./tool x" {
		t.Fatal("an argv spells its words")
	}
	// The long row's bounded line.
	long := strings.Join(strings.Fields(strings.Repeat("a ", 100)), " ")
	if got := oneLine(long); got != long[:100]+"…" {
		t.Fatalf("the bounded line: %q", got)
	}
	// A universe the run never numbered answers no line.
	m2 := &machine{standing: &journal.Standing{}}
	if text, ok := m2.demandLine("D1"); ok || text != "" {
		t.Fatal("an unloaded universe names no line")
	}
	// A rowless standing behavior says so — the shape laws keep such
	// heads out of the corpus; the forged journal's own view stays
	// honest.
	m3 := &machine{standing: &journal.Standing{Rows: map[string]journal.RowFact{}}}
	if got := m3.behaviorStatus(doc.Behavior{ID: "x", Name: "n"}); got != "no rows" {
		t.Fatalf("the rowless standing says so, got %q", got)
	}
}

// TestDefensiveSpellsAreTotal — the small spells over gated data
// answer totals, never guesses: an unparsable order line and an
// unparseable exclusion range.
func TestDefensiveSpellsAreTotal(t *testing.T) {
	if n := orderLine(doc.OpenOrder{Demand: "junk"}); n != 0 {
		t.Fatalf("an unparsable order line answers zero, got %d", n)
	}
	m := &machine{standing: &journal.Standing{
		Registry: doc.Registry{Exclusions: []doc.Exclusion{{Range: "junk"}}},
	}}
	if got := m.excludedShare(fprint.Coverage{Total: 10}); got != 0 {
		t.Fatalf("an unparsable exclusion holds no lines, got %d%%", got)
	}
	if _, state := (&machine{standing: &journal.Standing{}}).orderOn(1); state != "" {
		t.Fatal("a line with no standing order names no order")
	}
}

// TestRemoveRetiresRegistryParts — the remove list's marker and
// exclusion spellings retire their standing parts through the whole
// machine: the D<n> key lifts the marker on the line, the range key
// lifts the exclusion by its own range, and the resolution honors
// the marker-first context.
func TestRemoveRetiresRegistryParts(t *testing.T) {
	mountPorts()
	root := newRoot(t)
	if err := os.WriteFile(filepath.Join(root, "gen.txt"), []byte("g\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, `
work: a marker and an exclusion stand
not-a-requirement:
    - lines: [D1]
      reason: the title frames, it demands nothing
uncovered-lines:
    - {range: "gen.txt:1-1", reason: generated code, refs: [D1]}
`)
	runIn(t, root)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "markers: 1") || !strings.Contains(state, "excluded lines: 1 ranges") {
		t.Fatalf("the marker and the exclusion must stand first, state:\n%s", state)
	}
	// The marker retires by its line, the exclusion by its range.
	writeZone(t, root, zone.InboxSlot, `
work: both retired by their keys
remove: [D1, "gen.txt:1-1"]
`)
	runIn(t, root)
	state = zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "markers: 0") || !strings.Contains(state, "excluded lines: 0 ranges") {
		t.Fatalf("the remove keys must retire their standing parts, state:\n%s", state)
	}
	// A remove key that stands nowhere is refused by its own name.
	writeZone(t, root, zone.InboxSlot, `
work: the same keys again
remove: [D1, "gen.txt:1-1"]
`)
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, `remove[0]: the corpus carries no marker or order on D1 (law:document)`) ||
		!strings.Contains(report, `remove[1]: the corpus carries no exclusion "gen.txt:1-1" (law:document)`) {
		t.Fatalf("the retired keys must refuse to retire twice:\n%s", report)
	}
}

// TestRemoveResolvesMarkerBeforeOrder — the D<n> key names the
// marker standing on the line first: with both a marker and an order
// on their own lines, each key retires exactly its own part.
func TestRemoveResolvesMarkerBeforeOrder(t *testing.T) {
	mountPorts()
	root := newRoot(t)
	writeZone(t, root, zone.DemandsSlot, "a checked tool\na heading\n")
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, `
work: a marker and an order stand on their own lines
intents:
    - {lines: [D2], outcome: mark, prose: the heading orders nothing}
not-a-requirement:
    - lines: [D2]
      reason: framing
open:
    - {demand: D1, question: which color, state: deferred, reason: the human said later}
`)
	runIn(t, root)
	// The marker's line retires the marker, the deferred order stands.
	writeZone(t, root, zone.InboxSlot, `
work: the marker retires
remove: [D2]
`)
	runIn(t, root)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "markers: 0") || !strings.Contains(state, "deferred orders:") {
		t.Fatalf("the marker's key retires the marker alone, state:\n%s", state)
	}
	// The order's line retires the order.
	writeZone(t, root, zone.InboxSlot, `
work: the deferred order retires
remove: [D1]
`)
	runIn(t, root)
	state = zoneFile(t, root, zone.StateFile)
	if strings.Contains(state, "deferred orders:") {
		t.Fatalf("the order's key retires the order, state:\n%s", state)
	}
}

// TestRemoveTailsWithoutFullResolvers — the D<n> remove spelling
// with a resolver that carries neither the marker nor the order
// function stands down: the gate skips, nothing refuses.
func TestRemoveTailsWithoutFullResolvers(t *testing.T) {
	if r := doc.Validate(doc.Document{Work: "w", Remove: []string{"D4"}}, &doc.Resolvers{}); len(r) != 0 {
		t.Fatalf("a resolver without the marker and order functions skips the D<n> gate: %v", r)
	}
}

// TestSpellsOverForgedDataStayTotal — the defensive spells over data
// the gates would never let stand: an unparsable order demand, a
// marker whose lines do not parse, a marker key of junk lines.
func TestSpellsOverForgedDataStayTotal(t *testing.T) {
	if n := orderLineNumber(doc.OpenOrder{Demand: "junk"}); n != 0 {
		t.Fatalf("an unparsable order demand answers zero, got %d", n)
	}
	m := &machine{standing: &journal.Standing{
		Registry: doc.Registry{Markers: []doc.Marker{{Lines: []string{"junk"}, Reason: "forged"}}},
	}}
	if got := m.markerReasonOf(1); got != "" {
		t.Fatalf("a forged marker names no reason, got %q", got)
	}
	if key := doc.MarkerKey(doc.Marker{Lines: []string{"junk"}}); key != "junk" {
		t.Fatalf("a junk line set folds to its own spelling, got %q", key)
	}
}
