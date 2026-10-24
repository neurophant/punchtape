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
	"time"

	"github.com/neurophant/punchtape/internal/domain/doc"
	"github.com/neurophant/punchtape/internal/domain/journal"
	"github.com/neurophant/punchtape/internal/domain/zone"
	"github.com/neurophant/punchtape/internal/infrastructure/store/cas"
)

func TestEmptyPathFailsTheLaunchByName(t *testing.T) {
	mountPorts()
	saved := os.Getenv("PATH")
	t.Setenv("PATH", "")
	// The build's launch failure carries its own words.
	root := newBornRoot(t)
	writeZone(t, root, zone.InboxSlot, `build: "./no-such-tool x"
behaviors:
    - id: tool/x
      name: X
      prose: P.
      scope:
        - src/**
      rows:
        - run: "true"
          expect:
            rc: 0
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
	saved := RowTimeout
	RowTimeout = 300 * time.Millisecond
	defer func() { RowTimeout = saved }()
	root := newRoot(t)
	writeZone(t, root, zone.InboxSlot, `behaviors:
    - id: tool/slow
      name: Slow
      prose: P.
      scope:
        - src/**
      rows:
        - run: sleep 30
          expect:
            rc: 0
`)
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "timed out after") {
		t.Fatalf("the deadline must cut and name itself, report:\n%s", report)
	}
}

func TestSecondBirthLosesGracefully(t *testing.T) {
	mountPorts()
	root := t.TempDir()
	writeZone(t, root, zone.WishSlot, "one goal")
	if err := birth(filepath.Join(root, zone.Dirname)); err != nil {
		t.Fatalf("the first birth wins: %v", err)
	}
	// The journal already exists: the second birth sees a living
	// instance and answers nothing wrong.
	if err := birth(filepath.Join(root, zone.Dirname)); err != nil {
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
	// Red first.
	writeZone(t, root, zone.InboxSlot, `behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies.
      scope:
        - src/**
      rows:
        - run: cat src/in.txt > out.txt
          expect:
            files:
                out.txt: "contains: NOPE"
`)
	runIn(t, root)
	// The heal runs with the objects unreadable: the lesson's red
	// spelling falls back to the digest's name, the green still
	// lands. The heal document is the original submission's own
	// bytes — every object it names already stands.
	writeZone(t, root, zone.InboxSlot, `requirements:
    - id: R1
      brief: [W1]
artifacts: [out.txt]
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
	code, _ := runCode(t, root, map[string]bool{"read:objects!after-write": true})
	if code != 0 {
		t.Fatalf("the heal must still answer: %d", code)
	}
	knowledge := zoneFile(t, root, zone.KnowledgeFile)
	if !strings.Contains(knowledge, "<row ") {
		t.Fatalf("the unfindable bytes spell the digest, knowledge:\n%s", knowledge)
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
	// Membership both ways.
	if !contains([]string{"x"}, "x") || contains([]string{"x"}, "y") {
		t.Fatal("membership")
	}
	// The lesson lookup's miss.
	m := &machine{standing: &journal.Standing{}}
	if m.hasLessonFor("never") {
		t.Fatal("no lessons hold an unknown reason")
	}
	// A noteless instance counts zero.
	if m.noteCount() != 0 {
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
	// A wish object that does not read answers no line.
	m2 := &machine{
		objects:  cas.New(filepath.Join(t.TempDir(), "objects")),
		standing: &journal.Standing{GoalAddr: strings.Repeat("a", 64)},
	}
	if text, ok := m2.wishLine("W1"); ok || text != "" {
		t.Fatal("a missing wish object names no line")
	}
	// A rowless standing behavior says so — the shape laws keep such
	// heads out of the corpus; the forged journal's own view stays
	// honest.
	m3 := &machine{standing: &journal.Standing{Rows: map[string]journal.RowFact{}}}
	if got := m3.behaviorStatus(doc.Behavior{ID: "x", Name: "n"}); got != "no rows" {
		t.Fatalf("the rowless standing says so, got %q", got)
	}
}
