// The demand stream's battery: the coverage map over the numbered
// universe, the chain that holds the past, the machine's own taking
// of the stream's tail at the submission that meets it, the
// journaled none-fact, the foreign-version refusal, and the state's
// demands block that carries the whole account.
package application

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neurophant/punchtape/internal/domain/journal"
	"github.com/neurophant/punchtape/internal/domain/zone"
)

func TestDemandCoverageMap(t *testing.T) {
	universe := []string{"one", "two", "three", "four", "five"}
	// An empty registry: nothing is claimed, every line is a hole.
	claimed, holes := demandCoverage(universe, nil)
	if claimed != 0 || len(holes) != 5 || holes[0] != 1 || holes[4] != 5 {
		t.Fatalf("no claims must leave the whole universe open: %d %v", claimed, holes)
	}
	// One range over everything: the whole stream is accounted.
	if claimed, holes = demandCoverage(universe, []string{"D1..D5"}); claimed != 5 || len(holes) != 0 {
		t.Fatalf("one full range must claim the universe: %d %v", claimed, holes)
	}
	// Two requirements overlapping: the union claims, no double count.
	if claimed, holes = demandCoverage(universe, []string{"D1..D3", "D3..D5"}); claimed != 5 || len(holes) != 0 {
		t.Fatalf("overlaps claim once: %d %v", claimed, holes)
	}
	// Numbers outside the universe claim nothing; one broken spelling
	// voids the whole claim list — the map only counts what parses.
	if claimed, holes = demandCoverage(universe, []string{"D1..D2", "D9"}); claimed != 2 || len(holes) != 3 || holes[0] != 3 {
		t.Fatalf("outside numbers claim nothing: %d %v", claimed, holes)
	}
	if claimed, holes = demandCoverage(universe, []string{"D1..D2", "junk"}); claimed != 0 || len(holes) != 5 {
		t.Fatalf("a broken spelling voids the claims: %d %v", claimed, holes)
	}
}

func TestStateCarriesTheDemandsMap(t *testing.T) {
	mountPorts()
	root := t.TempDir()
	// A stream of three lines with blanks between them: the state
	// counts the whole stream open, the stream itself alive in the
	// zone.
	writeZone(t, root, zone.DemandsSlot, "first\n\nsecond\n\nthird\n")
	runIn(t, root)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "demands: 0/3 lines; requirements 0 standing, 0 covered; markers: 0") {
		t.Fatalf("the demands block must carry the counts:\n%s", state)
	}
	if _, err := os.Stat(filepath.Join(root, zone.Dirname, zone.DemandsSlot)); err != nil {
		t.Fatalf("the stream survives the birth, open forever: %v", err)
	}
}

// TestPastImmutableStands — the journal's chain holds the stream's
// journaled past: an appended tail is the stream's open future, a
// touched past line is zone damage, a gone stream file the same.
func TestPastImmutableStands(t *testing.T) {
	mountPorts()
	root := newRoot(t)
	// The stream is open forever: a new line lands as an unjournaled
	// tail, the run stands, the state numbers it unclaimed.
	writeZone(t, root, zone.DemandsSlot, "  a checked tool  \na second wish\n")
	runIn(t, root)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "demands: 0/2 lines") {
		t.Fatalf("an appended tail must number without claiming, state:\n%s", state)
	}
	// The honest submission declares the tail line: the machine takes
	// the words itself, the chain grows over them.
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, fmt.Sprintf(`work: the second wish accounted
intents:
    - {lines: [D2], outcome: work, prose: the second wish wants its own tool}
requirements:
    - id: R2
      lines: [D2]
      prose: the second line
behaviors:
    - id: tool/second
      name: Second
      prose: Covers the second line.
      claims:
        - src/in.txt:1-1
      covers: [R2]
      rows:
        - run: sh -c 'printf "%%s %%s\n" %s %s; cat src/in.txt > out2.txt'
          expect:
            files:
                out2.txt: ok
`, claimsTokenOf(t, root, "src/in.txt:1-1"), demandTokenOf("a second wish")))
	runIn(t, root)
	if state = zoneFile(t, root, zone.StateFile); !strings.Contains(state, "demands: 1/2 lines; requirements 1 standing, 1 covered") {
		t.Fatalf("the grown chain must make the tail claimable past, state:\n%s", state)
	}
	// The genesis line edited: its link breaks, the zone is damaged,
	// the run never starts.
	writeZone(t, root, zone.DemandsSlot, "A CHECKED TOOL\na second wish\n")
	var term strings.Builder
	if code := Run(root, &term); code != 1 || !strings.Contains(term.String(),
		"punchtape: zone damaged: the demand stream's past changed") {
		t.Fatalf("a touched past line must refuse the run: code %d term %q", code, term.String())
	}
	// The appended line edited: its own link breaks the same way.
	writeZone(t, root, zone.DemandsSlot, "a checked tool\nA SECOND WISH\n")
	term.Reset()
	if code := Run(root, &term); code != 1 || !strings.Contains(term.String(), "the demand stream's past changed") {
		t.Fatalf("a touched appended line must refuse the run: code %d term %q", code, term.String())
	}
	// The last journaled line deleted: the earlier links recompose,
	// the past outruns the file — the same damage.
	writeZone(t, root, zone.DemandsSlot, "a checked tool\n")
	term.Reset()
	if code := Run(root, &term); code != 1 || !strings.Contains(term.String(), "the demand stream's past changed") {
		t.Fatalf("a shrunken past must refuse the run: code %d term %q", code, term.String())
	}
	// The stream file gone: the same damage.
	fresh := newBornRoot(t)
	if err := os.Remove(filepath.Join(fresh, zone.Dirname, zone.DemandsSlot)); err != nil {
		t.Fatal(err)
	}
	term.Reset()
	if code := Run(fresh, &term); code != 1 || !strings.Contains(term.String(), "the demand stream's past changed") {
		t.Fatalf("a missing stream must refuse the run: code %d term %q", code, term.String())
	}
}

// TestMachineTakesTailItself — the words' entrance is the file alone:
// a submission that declares the tail's lines carries none of their
// bytes, and the machine journals the tail as the chain's new link
// with the chain value over the stream's own lines.
func TestMachineTakesTailItself(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	if err := os.WriteFile(filepath.Join(root, "src", "late.txt"), []byte("late"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.DemandsSlot, "a checked tool\na late wish\n")
	writeZone(t, root, zone.InboxSlot, fmt.Sprintf(`work: the late wish enters the stream
intents:
    - {lines: [D2], outcome: work, prose: the late wish wants its own tool}
requirements:
    - id: R2
      lines: [D2]
      prose: the late line
behaviors:
    - id: tool/late
      name: Late
      prose: Covers the late line.
      claims:
        - src/late.txt:1-1
      covers: [R2]
      rows:
        - run: sh -c 'printf "%%s %%s\n" %s %s; cat src/late.txt > out3.txt'
          expect:
            files:
                out3.txt: late
`, claimsTokenOf(t, root, "src/late.txt:1-1"), demandTokenOf("a late wish")))
	runIn(t, root)
	if _, err := os.Stat(filepath.Join(root, zone.Dirname, zone.InboxSlot)); !os.IsNotExist(err) {
		t.Fatalf("the declared tail must be accepted without the words' bytes")
	}
	records, err := journalRecords(t, root)
	if err != nil {
		t.Fatal(err)
	}
	var link *journal.DemandAppend
	for _, rec := range records {
		if rec.Kind == journal.KindDemandAppend && rec.DemandAppend != nil {
			fact := *rec.DemandAppend
			link = &fact
		}
	}
	if link == nil || link.Lines != 1 || link.Chain != chainOver(bornChainHead(t, root), []string{"a late wish"}) {
		t.Fatalf("the machine must journal the tail itself with the chain over its lines: %+v", link)
	}
	if state := zoneFile(t, root, zone.StateFile); !strings.Contains(state, "demands: 2/2 lines; requirements 2 standing, 2 covered") {
		t.Fatalf("the grown stream must be claimed whole, state:\n%s", state)
	}
}

// bornChainHead — the genesis link's chain value of a born root.
func bornChainHead(t *testing.T, root string) string {
	t.Helper()
	records, err := journalRecords(t, root)
	if err != nil {
		t.Fatal(err)
	}
	for _, rec := range records {
		if rec.Kind == journal.KindDemandAppend && rec.DemandAppend != nil {
			return rec.DemandAppend.Chain
		}
	}
	t.Fatal("no genesis link stands")
	return ""
}

// TestNothingNewJournaledOnce — a submission over an empty tail
// lands the machine's own none-fact once: the stream's count at the
// signing and the signing run's number, standing in the state's face
// until the next demand-append clears it.
func TestNothingNewJournaledOnce(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	// The birth document declared the whole stream: the none-fact
	// stands in the journal and in the state.
	records, err := journalRecords(t, root)
	if err != nil {
		t.Fatal(err)
	}
	var signed *journal.NoNewDemands
	for _, rec := range records {
		if rec.Kind == journal.KindNoNewDemands && rec.NoNewDemands != nil {
			fact := *rec.NoNewDemands
			signed = &fact
		}
	}
	if signed == nil || signed.Lines != 1 || signed.Run != 1 {
		t.Fatalf("the none-fact must journal the stream's count and the run: %+v", signed)
	}
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "last no-new-demands: run 1") {
		t.Fatalf("the standing none-fact must show in the state:\n%s", state)
	}
	// A silent second submission over the same empty tail: the
	// none-fact lands again for its own run, once per submission.
	writeZone(t, root, zone.InboxSlot, "work: a silent pass\n")
	runIn(t, root)
	if state = zoneFile(t, root, zone.StateFile); !strings.Contains(state, "last no-new-demands: run 2") {
		t.Fatalf("each accepted submission over an empty tail signs its own none:\n%s", state)
	}
	records, err = journalRecords(t, root)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, rec := range records {
		if rec.Kind == journal.KindNoNewDemands {
			count++
		}
	}
	if count != 2 {
		t.Fatalf("two accepted submissions over an empty tail journal two none-facts, got %d", count)
	}
	// The next demand-append clears the standing signature.
	writeZone(t, root, zone.DemandsSlot, "a checked tool\na late wish\n")
	writeZone(t, root, zone.InboxSlot, `work: the late wish enters
intents:
    - {lines: [D2], outcome: mark, prose: the late wish orders nothing}
not-a-requirement:
    - lines: [D2]
      reason: talk, not an order
`)
	runIn(t, root)
	if state = zoneFile(t, root, zone.StateFile); strings.Contains(state, "last no-new-demands") {
		t.Fatalf("a demand-append must clear the standing signature:\n%s", state)
	}
}

// TestUncoveredTailLineRefused — a tail line the submission's intents
// do not declare is refused by its own number; the refusal names the
// line, the inbox stays, nothing journals.
func TestUncoveredTailLineRefused(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	writeZone(t, root, zone.DemandsSlot, "a checked tool\na spoken word\nanother spoken word\n")
	writeZone(t, root, zone.InboxSlot, `work: only one of the two spoken lines declared
intents:
    - {lines: [D2], outcome: mark, prose: the first spoken word orders nothing}
not-a-requirement:
    - lines: [D2]
      reason: talk, not an order
`)
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "intents: the new demand lines D3 carry no outcome — declare answer, execute, describe, work, mark or ask (law:demand)") {
		t.Fatalf("the undeclared tail line must be refused by its number:\n%s", report)
	}
	if _, err := os.Stat(filepath.Join(root, zone.Dirname, zone.InboxSlot)); err != nil {
		t.Fatalf("the refused submission keeps the inbox")
	}
}

// TestForeignVersionZoneIsRefused — a zone born by another machine
// version names its creator and is refused before any record folds:
// the owner archives it and starts fresh. A birth that names no
// machine at all is a foreign shape the same way.
func TestForeignVersionZoneIsRefused(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	records, err := journalRecords(t, root)
	if err != nil {
		t.Fatal(err)
	}
	reforgeBirth := func(version string) {
		t.Helper()
		first := epochPaths(t, root)[0]
		birth := journal.Record{Kind: journal.KindBirth, Birth: &journal.Birth{Version: version}}
		var b strings.Builder
		b.WriteString("---\n")
		b.Write(journal.MarshalRecord(birth))
		b.WriteString("---\n")
		b.Write(journal.MarshalRecord(records[1])) // the genesis demand-append
		if err := os.WriteFile(first, []byte(b.String()), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	reforgeBirth("0.3.0")
	var term strings.Builder
	if code := Run(root, &term); code != 1 || !strings.Contains(term.String(), fmt.Sprintf(
		"punchtape: zone born by machine 0.3.0 — this machine is %s; archive it (cp -r .punchtape .punchtape-archive) and start fresh", Version)) {
		t.Fatalf("the foreign zone must be refused by name: code %d term %q", code, term.String())
	}
	// A birth with no version: the same refusal — the zone does
	// not name this machine.
	reforgeBirth("")
	term.Reset()
	if code := Run(root, &term); code != 1 || !strings.Contains(term.String(), "archive it (cp -r .punchtape .punchtape-archive) and start fresh") {
		t.Fatalf("a version-less birth must be refused the same way: code %d term %q", code, term.String())
	}
}

// TestMarkerStandsAndCounts — a marker on an open line stands by its
// reason and shows in the demands map's count; a marker on a line a
// requirement claims is the one contradiction the law refuses.
func TestMarkerStandsAndCounts(t *testing.T) {
	mountPorts()
	root := t.TempDir()
	writeZone(t, root, zone.DemandsSlot, "a heading\n\ndo the work\n")
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, `work: marking the heading
not-a-requirement:
    - lines: [D1]
      reason: the title frames, it demands nothing
`)
	runIn(t, root)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "demands: 1/2 lines; requirements 0 standing, 0 covered; markers: 1") {
		t.Fatalf("the standing mark must show in the demands map's count, state:\n%s", state)
	}
	// A marker on a requirement's line: the quiet contradiction,
	// refused by name.
	writeZone(t, root, zone.InboxSlot, `work: the contradiction
requirements:
    - id: R1
      lines: [D1]
      prose: the heading made specific
not-a-requirement:
    - lines: [D1]
      reason: the same line as a marker
`)
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "not-a-requirement[0].lines") || !strings.Contains(report, "D1 is claimed by R1 — a line is either a demand or not, not both") {
		t.Fatalf("the marker-requirement conflict must be refused by name:\n%s", report)
	}
}

// The awaiting order's own closure law: an order closes only over a
// later human line, occupied and cited by a decision note — a removal
// of its line and a switch to deferred are both a close.
func TestAwaitingCloseRequiresLaterHumanLine(t *testing.T) {
	root := newBornRoot(t)
	writeZone(t, root, zone.InboxSlot, `
work: a question stands before the human
open:
    - {demand: D1, question: which color, state: awaiting}
`)
	runIn(t, root)
	// No line newer than the order exists: the removal closes it
	// childless, refused by name.
	writeZone(t, root, zone.InboxSlot, `
work: the order closes
remove: [D1]
`)
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "open[0]: the order was closed without a later human line — bring the human's answer into demands.md, given its outcome, cited by a decision note (law:candor)") {
		t.Fatalf("the childless close must be refused by name:\n%s", report)
	}
}

func TestAwaitingCloseWithoutNoteRefused(t *testing.T) {
	root := newBornRoot(t)
	writeZone(t, root, zone.InboxSlot, `
work: a question stands before the human
open:
    - {demand: D1, question: which color, state: awaiting}
`)
	runIn(t, root)
	// The answer's line stands and is occupied by its marker — but no
	// decision note cites the line itself: the same refusal.
	writeZone(t, root, zone.DemandsSlot, "a checked tool\nthe green one\n")
	writeZone(t, root, zone.InboxSlot, `work: the answer stands uncited
intents:
    - {lines: [D2], outcome: mark, prose: the answer of the human carries no order}
not-a-requirement:
    - lines: [D2]
      reason: the human's answer, not new work
notes:
    - kind: decision
      text: cites the requirement, not the line
      refs: [R1]
remove: [D1]
`)
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "open[0]: the order was closed without a later human line — bring the human's answer into demands.md, given its outcome, cited by a decision note (law:candor)") {
		t.Fatalf("the uncited close must be refused by name:\n%s", report)
	}
}

func TestAwaitingCloseWithFreshLineAndNotePasses(t *testing.T) {
	root := newBornRoot(t)
	writeZone(t, root, zone.InboxSlot, `
work: a question stands before the human
open:
    - {demand: D1, question: which color, state: awaiting}
`)
	runIn(t, root)
	// The answer's line, its marker and the decision note that cites
	// it: the close is legal, the verdict may land.
	writeZone(t, root, zone.DemandsSlot, "a checked tool\nthe green one\n")
	writeZone(t, root, zone.InboxSlot, `work: the cited answer closes the order
intents:
    - {lines: [D2], outcome: mark, prose: the answer of the human carries no order}
not-a-requirement:
    - lines: [D2]
      reason: the human's answer, not new work
notes:
    - kind: decision
      text: the color is green by the human's answer
      refs: [D2]
remove: [D1]
`)
	runIn(t, root)
	if _, err := os.Stat(filepath.Join(root, zone.Dirname, zone.InboxSlot)); !os.IsNotExist(err) {
		t.Fatalf("the cited close must be accepted")
	}
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if state := zoneFile(t, root, zone.StateFile); !strings.Contains(state, "verdict: #1") {
		t.Fatalf("the closed order must let the verdict land, state:\n%s", state)
	}
}

func TestDeferredSwitchIsAlsoClosure(t *testing.T) {
	root := newBornRoot(t)
	writeZone(t, root, zone.InboxSlot, `
work: a question stands before the human
open:
    - {demand: D1, question: which color, state: awaiting}
`)
	runIn(t, root)
	// The switch to deferred without a later human line is a close
	// all the same: the same refusal by the same name.
	writeZone(t, root, zone.InboxSlot, `
work: the order defers
open:
    - {demand: D1, question: which color, state: deferred, reason: the human said later}
`)
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "open[0]: the order was closed without a later human line — bring the human's answer into demands.md, given its outcome, cited by a decision note (law:candor)") {
		t.Fatalf("the deferred switch must be refused as a close:\n%s", report)
	}
}

// TestAwaitingCloseOverRecords — the later human line may carry any
// carrier: an execute intent's line and an answer intent's line both
// occupy their line for the closure's law.
func TestAwaitingCloseOverRecords(t *testing.T) {
	root := newBornRoot(t)
	writeZone(t, root, zone.InboxSlot, `
work: a question stands before the human
open:
    - {demand: D1, question: which color, state: awaiting}
`)
	runIn(t, root)
	writeZone(t, root, zone.DemandsSlot, "a checked tool\nrun the echo\nwhat does the check read?\n")
	writeZone(t, root, zone.InboxSlot, `work: the execution and the answer close the order
intents:
    - {lines: [D2], outcome: execute, prose: the command's own output answers, run: "printf done"}
    - {lines: [D3], outcome: answer, prose: the question of the input answered, text: the check reads the input's own bytes}
notes:
    - kind: work
      text: the narrative beside the closing decision
      refs: [D2, D3]
    - kind: decision
      text: both carriers answer the standing question
      refs: [D2, D3]
remove: [D1]
`)
	runIn(t, root)
	if _, err := os.Stat(filepath.Join(root, zone.Dirname, zone.InboxSlot)); !os.IsNotExist(err) {
		t.Fatalf("the close over the execution's and the answer's lines must be accepted")
	}
	// The mirrored order — the answer's line first — closes the same
	// way: the occupation is the carrier's own, whichever kind it is.
	root = newBornRoot(t)
	writeZone(t, root, zone.InboxSlot, `
work: a question stands before the human
open:
    - {demand: D1, question: which color, state: awaiting}
`)
	runIn(t, root)
	writeZone(t, root, zone.DemandsSlot, "a checked tool\nwhat does the check read?\nrun the echo\n")
	writeZone(t, root, zone.InboxSlot, `work: the answer and the execution close the order
intents:
    - {lines: [D2], outcome: answer, prose: the question of the input answered, text: the check reads the input's own bytes}
    - {lines: [D3], outcome: execute, prose: the command's own output answers, run: "printf done"}
notes:
    - kind: decision
      text: both carriers answer the standing question
      refs: [D2, D3]
remove: [D1]
`)
	runIn(t, root)
	if _, err := os.Stat(filepath.Join(root, zone.Dirname, zone.InboxSlot)); !os.IsNotExist(err) {
		t.Fatalf("the close over the answer's and the execution's lines must be accepted")
	}
}
