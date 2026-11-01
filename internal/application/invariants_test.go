// The invariants battery: the machine's own laws, each pinned by an
// assertion — not merely executed, but held. Every test here answers
// one stated law of the machine's surface, zone, birth, algorithm,
// fingerprints, knowledge or execution semantics.
package application

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neurophant/punchtape/internal/domain/zone"
)

// TestZoneFoundFromASubdirectory — the zone is sought upward from the
// working directory; a run inside the tree's subdirectory serves the
// same instance, and a nearer zone wins over the farther one.
func TestZoneFoundFromASubdirectory(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	sub := filepath.Join(root, "deep", "down")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Skipf("the working directory is not readable: %v", err)
	}
	if err := os.Chdir(sub); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(wd) })
	runIn(t, sub) // settle: the row's own output enters the walk once
	stateBefore := zoneFile(t, root, zone.StateFile)
	term := runIn(t, sub)
	if zoneFile(t, root, zone.StateFile) != stateBefore {
		t.Fatalf("a run from a subdirectory must serve the same instance")
	}
	if !strings.Contains(term, "run 1") {
		t.Fatalf("the subdirectory run must answer over the same zone: %q", term)
	}
	// A nearer zone is its own instance.
	near := filepath.Join(sub, zone.Dirname)
	if err := os.MkdirAll(near, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(near, zone.DemandsSlot), []byte("a nearer goal"), 0o644); err != nil {
		t.Fatal(err)
	}
	runIn(t, sub)
	if _, err := os.Stat(filepath.Join(near, zone.JournalDir)); err != nil {
		t.Fatalf("the nearer zone must win: %v", err)
	}
	if data, err := os.ReadFile(filepath.Join(near, zone.DemandsSlot)); err != nil || string(data) != "a nearer goal" {
		t.Fatalf("the nearer zone's birth keeps its own stream, open forever: %v %q", err, data)
	}
}

// TestTerminalIsExactlyOneLine — every run answers with one pointer
// line, nothing else.
func TestTerminalIsExactlyOneLine(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	var term strings.Builder
	code := Run(root, &term)
	if code != 0 {
		t.Fatalf("run exited %d", code)
	}
	if n := strings.Count(strings.TrimSpace(term.String()), "\n"); n != 0 {
		t.Fatalf("the terminal must carry exactly one line, got %d extra: %q", n, term.String())
	}
}

// TestTmpClearedAtTheRunEnd — the machine's scratch leaves with the
// run, even someone else's bytes inside it.
func TestTmpClearedAtTheRunEnd(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	tmp := filepath.Join(root, zone.Dirname, zone.TmpDir)
	if err := os.MkdirAll(tmp, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "leftover"), []byte("junk"), 0o644); err != nil {
		t.Fatal(err)
	}
	runIn(t, root)
	if _, err := os.Stat(tmp); !os.IsNotExist(err) {
		t.Fatalf("the scratch must leave with the run: %v", err)
	}
}

// TestStreamOpenAfterBirthAndLastWordingWins — the stream is open
// forever: a later line lands as the unjournaled tail and shows in
// the state's demands map. Before the birth, the bytes at the moment
// of the look are the universe: the last wording wins.
func TestStreamOpenAfterBirthAndLastWordingWins(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	runIn(t, root) // settle the row's output into the standing first
	writeZone(t, root, zone.DemandsSlot, "a checked tool\na late word\n")
	runIn(t, root)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "demands: 1/2 lines") {
		t.Fatalf("the late word must land as the open tail, state:\n%s", state)
	}

	fresh := t.TempDir()
	writeZone(t, fresh, zone.DemandsSlot, "the first wording\nand its second line\n")
	writeZone(t, fresh, zone.DemandsSlot, "the wording that counts\n")
	runIn(t, fresh)
	if state = zoneFile(t, fresh, zone.StateFile); !strings.Contains(state, "demands: 0/1 lines") {
		t.Fatalf("the universe is the bytes at the moment of the look, state:\n%s", state)
	}
}

// TestReformatIsNotAChange — a reformatted corpus whose meaning
// decodes to the same document is the completion act, not a change.
// TestARefusalJournalsOnlyItsOwnFact — a refused submission writes
// none of its bytes: no submission record, no corpus move. The
// machine's own decision stands as the one new record, and the
// state's counter answers it — through the fold, from the journal
// alone.
func TestARefusalJournalsOnlyItsOwnFact(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	corpusBefore := zoneFile(t, root, zone.CorpusFile)
	writeZone(t, root, zone.InboxSlot, "\tbroken: [")
	runIn(t, root)
	if zoneFile(t, root, zone.CorpusFile) != corpusBefore {
		t.Fatalf("a refusal must not move the corpus")
	}
	entries, err := os.ReadDir(filepath.Join(root, zone.Dirname, zone.JournalDir))
	if err != nil {
		t.Fatal(err)
	}
	var last string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".yaml") {
			last = e.Name()
		}
	}
	data, err := os.ReadFile(filepath.Join(root, zone.Dirname, zone.JournalDir, last))
	if err != nil {
		t.Fatal(err)
	}
	chunks := strings.Split(string(data), "---\n")
	lastRecord := strings.TrimSpace(chunks[len(chunks)-1])
	if !strings.Contains(lastRecord, "kind: refusal") || strings.Contains(lastRecord, "doc-addr") {
		t.Fatalf("the refusal journals its own fact and no document bytes:\n%s", lastRecord)
	}
	// The counter answers through the fold: with the broken inbox
	// gone, the next run reopens from the journal alone and the
	// state carries the count. (The refused runs themselves do not
	// move the state — a refusal changes no standing fact; the
	// terminal line sends the hand to report.md.)
	if err := os.Remove(filepath.Join(root, zone.Dirname, zone.InboxSlot)); err != nil {
		t.Fatal(err)
	}
	runIn(t, root)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "submissions: 1 accepted, 1 refused") {
		t.Fatalf("the state must count the refusal from the journal, state:\n%s", state)
	}
}

// TestVCSAndZoneExcludedFromTheTree — version-control metadata and
// the zone itself never enter the fingerprints or the gaps.
func TestVCSAndZoneExcludedFromTheTree(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	runIn(t, root) // settle the row's output into the standing first
	stateBefore := zoneFile(t, root, zone.StateFile)
	for _, dir := range []string{".git", ".hg", ".svn"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, dir, "metadata"), []byte("noise"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	runIn(t, root)
	state := zoneFile(t, root, zone.StateFile)
	if state != stateBefore {
		t.Fatalf("VCS metadata must not move the fingerprints, state:\n%s", state)
	}
	for _, gap := range []string{"- .git/", "- .hg/", "- .svn/", "- .punchtape/"} {
		if strings.Contains(state, gap) {
			t.Fatalf("VCS metadata and the zone are never the project's gaps (%s), state:\n%s", gap, state)
		}
	}
}

// TestMtimeOnlyChangeKeepsTheProof — a touched file with unchanged
// bytes is not a staleness: the claims digest the bytes, and
// unchanged bytes re-execute nothing.
func TestMtimeOnlyChangeKeepsTheProof(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	reportBefore := zoneFile(t, root, zone.ReportFile)
	same, err := os.ReadFile(filepath.Join(root, "src", "in.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), same, 0o644); err != nil {
		t.Fatal(err)
	}
	runIn(t, root)
	if zoneFile(t, root, zone.ReportFile) != reportBefore {
		t.Fatalf("a mtime-only change must not re-execute anything, report:\n%s", zoneFile(t, root, zone.ReportFile))
	}
}

// TestBuildRunsOnEveryLaunch — a declared build runs on every launch,
// even when no row is stale; the rows still stand from memory.
func TestBuildRunsOnEveryLaunch(t *testing.T) {
	mountPorts()
	root := newRoot(t)
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	// The first submission of a fresh root carries no mandate —
	// nothing stands to replace yet.
	firstDoc := strings.Replace(greenDocFor(t, root), "mandate: [{line: D1, target: tool/copy}]\n", "", 1)
	writeZone(t, root, zone.InboxSlot, "build: \"true\"\n"+firstDoc)
	runIn(t, root)
	if !strings.Contains(zoneFile(t, root, zone.ReportFile), "green: 1") {
		t.Fatalf("the first run must execute the row, report:\n%s", zoneFile(t, root, zone.ReportFile))
	}
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "build: exit 0") || !strings.Contains(report, "no stale rows") {
		t.Fatalf("the idle run must still run the declared build and no rows, report:\n%s", report)
	}
}

// TestViewsCarryTheFingerprints — the state names each behavior's
// claims shape and proof; the spec page names each row's proof.
func TestViewsCarryTheFingerprints(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "(claims 1 files / 1 lines; proof ") {
		t.Fatalf("the state must carry the claims shape and proof, state:\n%s", state)
	}
	page := zoneFile(t, root, filepath.Join(zone.SpecDir, "tool", "copy.md"))
	if !strings.Contains(page, "proof: green (") {
		t.Fatalf("the spec page must carry the proof fingerprint, page:\n%s", page)
	}
}
