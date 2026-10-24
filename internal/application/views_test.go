// The views' tails: the state's knowledge pointer for a reasoned
// red, the build's timeout line, the untagged note line, the
// completion act's whole-rewrite of the report.
package application

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/neurophant/punchtape/internal/domain/zone"
)

func TestStatePointsAtKnowledgeForAReasonedRed(t *testing.T) {
	mountPorts()
	root := newRoot(t)
	// Two behaviors fail the same way; one heals — its lesson stands;
	// the other stays red with the same reason: the state points at
	// the knowledge.
	writeZone(t, root, zone.InboxSlot, `behaviors:
    - id: tool/one
      name: One
      prose: Fails then heals.
      rows:
        - run: "true"
          expect:
            rc: 1
    - id: tool/two
      name: Two
      prose: Stays failing.
      rows:
        - run: "true"
          expect:
            rc: 1
`)
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, `behaviors:
    - id: tool/one
      name: One
      prose: Fails then heals.
      rows:
        - run: "true"
          expect:
            rc: 0
`)
	runIn(t, root)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "knowledge for the reason") {
		t.Fatalf("the state must point at the lesson:\n%s", state)
	}
}

func TestBuildTimeoutIsARedLine(t *testing.T) {
	mountPorts()
	root := newRoot(t)
	old := RowTimeout
	RowTimeout = 100 * time.Millisecond
	t.Cleanup(func() { RowTimeout = old })
	writeZone(t, root, zone.InboxSlot, "build: sleep 5\n"+`behaviors:
    - id: tool/slow
      name: Slow
      prose: Slow.
      rows:
        - run: "true"
          expect:
            rc: 0
`)
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "build: RED — timed out") {
		t.Fatalf("the timed-out build must say so:\n%s", report)
	}
}

func TestUntaggedNoteStandsOnce(t *testing.T) {
	mountPorts()
	root := newRoot(t)
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, `notes:
    - text: zzquux zzfrob
behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies.
      scope:
        - src/**
      rows:
        - run: "true"
          expect:
            rc: 0
`)
	runIn(t, root)
	knowledge := zoneFile(t, root, zone.KnowledgeFile)
	if !strings.Contains(knowledge, "- [n1] (untagged) zzquux zzfrob") {
		t.Fatalf("a note the vocabulary does not know stands once, untagged:\n%s", knowledge)
	}
	if strings.Contains(knowledge, "## index") {
		t.Fatalf("a note without tags carries no index:\n%s", knowledge)
	}
}

func TestNoteStandsOnceWithItsWholeIndex(t *testing.T) {
	mountPorts()
	root := t.TempDir()
	// One note whose every word the vocabulary knows: a hundred tags,
	// one text. The render's size is a function of the note's size —
	// never the note's size times its tags.
	var words []string
	for i := 1; i <= 100; i++ {
		words = append(words, fmt.Sprintf("w%d", i))
	}
	text := strings.Join(words, " ")
	writeZone(t, root, zone.WishSlot, text)
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, "notes:\n    - text: "+text+"\n")
	runIn(t, root)
	knowledge := zoneFile(t, root, zone.KnowledgeFile)
	if got := strings.Count(knowledge, text); got != 1 {
		t.Fatalf("the note's text must stand exactly once, found %d:\n%s", got, knowledge)
	}
	if got := strings.Count(knowledge, ": n1\n"); got != 100 {
		t.Fatalf("the index must carry one line per tag, found %d:\n%s", got, knowledge)
	}
	if len(knowledge) > 10*len(text) {
		t.Fatalf("the render must grow with the tags, not copy the text per tag: %d bytes for a %d-byte note",
			len(knowledge), len(text))
	}
}

func TestCompletionRewritesTheReportWhole(t *testing.T) {
	mountPorts()
	root := newRoot(t)
	// A stale truncated report and an old refusal note must not
	// outlive the completion act: the report is the run's whole
	// picture, rewritten.
	if err := os.WriteFile(filepath.Join(root, zone.Dirname, zone.ReportFile), []byte("# punchtape report\n\nrun 0 — truncated\nverdict: not yet — old noise\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, "{}\n")
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if strings.Contains(report, "truncated") || strings.Contains(report, "old noise") {
		t.Fatalf("the completion act must rewrite the report whole:\n%s", report)
	}
	if !strings.Contains(report, "verdict: refused — the corpus is empty") {
		t.Fatalf("the refusal stands as the run's own fact:\n%s", report)
	}
}

func TestAStandingVerdictLeavesNoStaleRefusalBehind(t *testing.T) {
	root := wishFixture(t)
	// The thin corpus: the completion act refuses and names the
	// holes — then the registry closes and the verdict stands; the
	// report of that run must carry the recording, not the refusal.
	writeZone(t, root, zone.InboxSlot, zoneFile(t, root, zone.CorpusFile))
	runIn(t, root)
	if !strings.Contains(zoneFile(t, root, zone.ReportFile), "verdict: not yet") {
		t.Fatalf("the refusal must stand in its own run's report")
	}
	writeZone(t, root, zone.InboxSlot, `requirements:
    - id: R1
      brief: [W1..W2]
not-a-requirement:
    - lines: [W3]
      quote: three
      reason: the closing line names no work
behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies.
      scope:
        - src/**
      covers: [R1]
      rows:
        - run: cat src/in.txt > out.txt
          expect:
            files:
                out.txt: "contains: ok"
`)
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, zoneFile(t, root, zone.CorpusFile))
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "verdict: #1") || strings.Contains(report, "not yet") {
		t.Fatalf("the recorded verdict must be the run's whole report:\n%s", report)
	}
}
