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

	"github.com/neurophant/punchtape/internal/domain"
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
	if err := os.WriteFile(filepath.Join(near, zone.WishSlot), []byte("a nearer goal"), 0o644); err != nil {
		t.Fatal(err)
	}
	runIn(t, sub)
	if _, err := os.Stat(filepath.Join(near, zone.JournalDir)); err != nil {
		t.Fatalf("the nearer zone must win: %v", err)
	}
	if _, err := os.Stat(filepath.Join(near, zone.WishSlot)); !os.IsNotExist(err) {
		t.Fatalf("the nearer zone's birth must consume its own wish")
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

// TestWishAfterBirthIgnoredAndOverwriteBeforeConsumption — the goal
// is fixed at the birth forever: a later wish is ignored; before the
// consumption the bytes at the moment of the look are the goal.
func TestWishAfterBirthIgnoredAndOverwriteBeforeConsumption(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	runIn(t, root) // settle the row's output into the standing first
	stateBefore := zoneFile(t, root, zone.StateFile)
	writeZone(t, root, zone.WishSlot, "a different goal entirely")
	runIn(t, root)
	if zoneFile(t, root, zone.StateFile) != stateBefore {
		t.Fatalf("a wish after the birth must change nothing")
	}
	if _, err := os.Stat(filepath.Join(root, zone.Dirname, zone.WishSlot)); err != nil {
		t.Fatalf("the ignored wish stays where it is: %v", err)
	}

	fresh := t.TempDir()
	writeZone(t, fresh, zone.WishSlot, "the first wording")
	writeZone(t, fresh, zone.WishSlot, "the wording that counts")
	runIn(t, fresh)
	state := zoneFile(t, fresh, zone.StateFile)
	if !strings.Contains(state, "goal: "+domain.Digest([]byte("the wording that counts"))[:12]) {
		t.Fatalf("the goal is the bytes at the moment of the look, state:\n%s", state)
	}
}

// TestReformatIsNotAChange — a reformatted corpus whose meaning
// decodes to the same document is the completion act, not a change.
func TestReformatIsNotAChange(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	corpus := zoneFile(t, root, zone.CorpusFile)
	reformatted := strings.Replace(corpus, `"contains: ok"`, `'contains: ok'`, 1)
	if reformatted == corpus {
		t.Fatal("the test needs a real reformatting")
	}
	writeZone(t, root, zone.InboxSlot, reformatted)
	runIn(t, root)
	if !strings.Contains(zoneFile(t, root, zone.StateFile), "verdict: #1") {
		t.Fatalf("a reformatted corpus must close the verdict, state:\n%s", zoneFile(t, root, zone.StateFile))
	}
}

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

// TestDigestIsOfRawBytes — the fingerprint is the sha256 of the
// bytes as they lie, no transformation ever: a file holding CRLF
// digests its own bytes, and the CRLF and LF spellings of one text
// are two different trees (a digest of processed
// bytes is a digest of bytes that never existed).
func TestDigestIsOfRawBytes(t *testing.T) {
	mountPorts()
	lf := t.TempDir()
	crlf := t.TempDir()
	os.WriteFile(filepath.Join(lf, "f.bin"), []byte("a\nb\nc\n"), 0o644)
	os.WriteFile(filepath.Join(crlf, "f.bin"), []byte("a\r\nb\r\nc\r\n"), 0o644)
	if got, want := OpenHashCache(filepath.Join(lf, ".z")).Digest(lf, "f.bin"), domain.Digest([]byte("a\nb\nc\n")); got != want {
		t.Fatalf("the LF file digests its own bytes: %s != %s", got, want)
	}
	if got, want := OpenHashCache(filepath.Join(crlf, ".z")).Digest(crlf, "f.bin"), domain.Digest([]byte("a\r\nb\r\nc\r\n")); got != want {
		t.Fatalf("the CRLF file digests its own bytes: %s != %s", got, want)
	}
	if OpenHashCache(filepath.Join(lf, ".z")).Digest(lf, "f.bin") == OpenHashCache(filepath.Join(crlf, ".z")).Digest(crlf, "f.bin") {
		t.Fatal("different bytes are a different tree — the spellings must not collide")
	}
}

// TestMtimeOnlyChangeKeepsTheProof — a touched file with unchanged
// bytes is a cache miss, not a staleness: nothing re-executes.
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

// TestInjectionLineOncePerReason — the knowledge pointer for a
// reasoned red stands once, however many rows carry the reason.
func TestInjectionLineOncePerReason(t *testing.T) {
	mountPorts()
	root := newRoot(t)
	writeZone(t, root, zone.InboxSlot, `behaviors:
    - id: tool/one
      name: One
      prose: Stays failing.
      rows:
        - run: "true"
          expect:
            rc: 1
    - id: tool/two
      name: Two
      prose: Stays failing the same way.
      rows:
        - run: "true"
          expect:
            rc: 1
`)
	runIn(t, root)
	// One of the two heals: its lesson exists for the shared reason.
	writeZone(t, root, zone.InboxSlot, `behaviors:
    - id: tool/one
      name: One
      prose: Heals.
      rows:
        - run: "true"
          expect:
            rc: 0
`)
	runIn(t, root)
	state := zoneFile(t, root, zone.StateFile)
	if n := strings.Count(state, "knowledge for the reason"); n != 1 {
		t.Fatalf("the injection must stand once per reason, found %d, state:\n%s", n, state)
	}
}

// TestRowGarbageIsAVisibleGap — a row's leftover file outside every
// scope is a visible fact of the state, not a silent hole.
func TestRowGarbageIsAVisibleGap(t *testing.T) {
	mountPorts()
	root := newRoot(t)
	writeZone(t, root, zone.InboxSlot, `behaviors:
    - id: tool/litter
      name: The litterer
      prose: Leaves a file behind.
      scope:
        - out.txt
      rows:
        - run: sh -c "echo junk > stray.txt; echo ok > out.txt"
          expect:
            files:
                out.txt: "contains: ok"
`)
	runIn(t, root)
	runIn(t, root) // the walk of the next run sees the leftover
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "- stray.txt") {
		t.Fatalf("the stray file must be a visible gap, state:\n%s", state)
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
	writeZone(t, root, zone.InboxSlot, "build: \"true\"\n"+greenDoc)
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
// scope fingerprint; the spec page names each row's proof.
func TestViewsCarryTheFingerprints(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "(scope ") {
		t.Fatalf("the state must carry the scope fingerprints, state:\n%s", state)
	}
	page := zoneFile(t, root, filepath.Join(zone.SpecDir, "tool", "copy.md"))
	if !strings.Contains(page, "proof: green (") {
		t.Fatalf("the spec page must carry the proof fingerprint, page:\n%s", page)
	}
}
