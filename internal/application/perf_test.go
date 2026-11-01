// The performance stage's battery: the declared universe and its
// verified walk cache, the journal's single records read and the
// fold's own indexes, and the zone's row deadline pin — every law
// named by its own counters and bytes, never by clocks.
package application

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/neurophant/punchtape/internal/domain"
	"github.com/neurophant/punchtape/internal/domain/doc"
	"github.com/neurophant/punchtape/internal/domain/fprint"
	"github.com/neurophant/punchtape/internal/domain/journal"
	"github.com/neurophant/punchtape/internal/domain/zone"
	"github.com/neurophant/punchtape/internal/infrastructure/fsys"
	"github.com/neurophant/punchtape/internal/infrastructure/process"
	"github.com/neurophant/punchtape/internal/infrastructure/store"
	"github.com/neurophant/punchtape/internal/ports"
)

// countFS — the real filesystem with a read counter over paths
// containing the key: the battery's own probe for the one-read and
// one-computation laws.
type countFS struct {
	ports.FS
	key   string
	reads int
}

func (f *countFS) ReadFile(p string) ([]byte, error) {
	if strings.Contains(p, f.key) {
		f.reads++
	}
	return f.FS.ReadFile(p)
}

// writeN — n one-line files under a directory of the root.
func writeN(t *testing.T, root, dir string, n int) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= n; i++ {
		path := filepath.Join(root, dir, fmt.Sprintf("f%03d.txt", i))
		if err := os.WriteFile(path, fmt.Appendf(nil, "line %d\n", i), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// TestUniverseDeclarationScopes — the declared universe bounds the
// count: with the zone's roots file only they walk; without it the
// whole tree; a line of any other shape is a loud refusal naming
// the file, the line and the law.
func TestUniverseDeclarationScopes(t *testing.T) {
	mountPorts()
	root := t.TempDir()
	writeZone(t, root, zone.DemandsSlot, "a checked tool")
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("ok\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "vendor"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "vendor", "x.txt"), []byte("vendored\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runIn(t, root)
	// Without the declaration the whole tree counts: the vendored
	// file stands among the holes by name.
	if state := zoneFile(t, root, zone.StateFile); !strings.Contains(state, "vendor/x.txt") {
		t.Fatalf("an undeclared universe must count the whole tree:\n%s", state)
	}
	// The declaration scopes the walk to its roots.
	writeZone(t, root, zone.UniverseSlot, "src\n")
	runIn(t, root)
	state := zoneFile(t, root, zone.StateFile)
	if strings.Contains(state, "vendor/x.txt") {
		t.Fatalf("a file outside the declared roots never holds the count:\n%s", state)
	}
	if !strings.Contains(state, "outside the universe: 2 files") {
		t.Fatalf("the outside must be loud with its count (vendor/x.txt and the hook):\n%s", state)
	}
	// A line of any other shape refuses the run by name.
	writeZone(t, root, zone.UniverseSlot, "src\n/abs\n")
	var term strings.Builder
	if code := Run(root, &term); code != 1 {
		t.Fatalf("a malformed universe must refuse the run, exit %d: %s", code, term.String())
	}
	want := `universe refused — .punchtape/universe, line 2: "/abs" — the law: one root-relative path a line`
	if !strings.Contains(term.String(), want) {
		t.Fatalf("the refusal must quote the file, the line and the law: %q", term.String())
	}
}

// TestOutsideUniverseLoud — every file outside the declared universe
// is named by the count, and the count follows the tree.
func TestOutsideUniverseLoud(t *testing.T) {
	mountPorts()
	root := newRoot(t)
	writeN(t, root, "src", 1)
	writeN(t, root, "vendor", 2)
	writeZone(t, root, zone.UniverseSlot, "src\n")
	runIn(t, root)
	// Three stand outside: two vendored files and the machine's own
	// hook at the root.
	if state := zoneFile(t, root, zone.StateFile); !strings.Contains(state, "outside the universe: 3 files") {
		t.Fatalf("the outside count must name every file beyond the roots:\n%s", state)
	}
	writeN(t, root, "vendor", 4)
	runIn(t, root)
	if state := zoneFile(t, root, zone.StateFile); !strings.Contains(state, "outside the universe: 5 files") {
		t.Fatalf("the outside count must follow the tree:\n%s", state)
	}
}

// TestUniverseCountsAlwaysSpeak — the declared universe's counts
// print at every share, with no threshold clause beside them: the
// narrowing is loud by its own numbers, never by a magic half.
func TestUniverseCountsAlwaysSpeak(t *testing.T) {
	mountPorts()
	root := newRoot(t)
	writeN(t, root, "src", 33)
	writeN(t, root, "vendor", 66)
	writeZone(t, root, zone.UniverseSlot, "src\n")
	runIn(t, root)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "universe: 33 of 100 files\n") {
		t.Fatalf("the counts must speak at a small share:\n%s", state)
	}
	if strings.Contains(state, "half") {
		t.Fatalf("no threshold clause survives beside the counts:\n%s", state)
	}
	// A covering majority: the same counts, nothing more.
	root = newRoot(t)
	writeN(t, root, "src", 51)
	writeN(t, root, "vendor", 49)
	writeZone(t, root, zone.UniverseSlot, "src\n")
	runIn(t, root)
	if state = zoneFile(t, root, zone.StateFile); !strings.Contains(state, "universe: 51 of 101 files\n") {
		t.Fatalf("the counts speak at any share:\n%s", state)
	}
}

// TestTreeCacheVerifiedAgainstClaims — the walk cache never touches
// the proofs: a claimed file swapped behind the cache's back (the
// size and mtime standing still) is caught by the claims' own
// direct read — the pin goes red by fact.
func TestTreeCacheVerifiedAgainstClaims(t *testing.T) {
	mountPorts()
	root := newRoot(t)
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, fmt.Sprintf(`
work: the pinned copy
intents:
    - {lines: [D1], outcome: work, prose: "the tool and its check, told again"}
requirements:
    - id: R1
      lines: [D1]
      prose: the tool and its check
uncovered-lines:
    - {range: "AGENTS.md:1-9", reason: the machine's own hook, refs: [D1]}
    - {range: "out.txt:1-1", reason: the row's own output, refs: [D1]}
behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies the input, pinned by its bytes.
      claims:
        - src/in.txt:1-1
      covers: [R1]
      rows:
        - run: sh -c 'printf "%%s %%s\n" %s %s; cat src/in.txt > out.txt'
          expect:
            files:
                out.txt: ok
`, claimsTokenOf(t, root, "src/in.txt:1-1"), demandTokenOf("a checked tool")))
	runIn(t, root)
	if report := zoneFile(t, root, zone.ReportFile); !strings.Contains(report, "green: 1") {
		t.Fatalf("the pinned row must stand green first:\n%s", report)
	}
	// The cache stands with the claimed file's own record.
	cache := zoneFile(t, root, zone.CacheTreeDir)
	if !strings.Contains(cache, "src/in.txt") {
		t.Fatalf("the walk cache must carry the claimed file's record:\n%s", cache)
	}
	// The swap: same size, same mtime — the cache cannot see it, the
	// claims' direct read must.
	path := filepath.Join(root, "src", "in.txt")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("NO"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, info.ModTime(), info.ModTime()); err != nil {
		t.Fatal(err)
	}
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, `want "ok"`) {
		t.Fatalf("the swapped claimed line must break the pin by fact:\n%s", report)
	}
}

// TestRowTimeoutPinShortens — the zone's deadline pin cuts the row
// by its own seconds, and the pin itself never reaches the command's
// environment.
func TestRowTimeoutPinShortens(t *testing.T) {
	mountPorts()
	root := newRoot(t)
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.EnvSlot, "PUNCHTAPE_ROW_TIMEOUT=2\n")
	writeZone(t, root, zone.InboxSlot, `
work: the slow rows under a short deadline
behaviors:
    - id: tool/slow
      name: The slow
      prose: Sleeps past the pin.
      claims:
        - src/in.txt:1-1
      rows:
        - run: sh -c 'sleep 3; printf "done\n"'
          expect:
            out: "done\n"
        - run: printenv PUNCHTAPE_ROW_TIMEOUT
          expect:
            rc: 1
`)
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "timed out after 2s") {
		t.Fatalf("the pinned deadline must cut the sleeping row:\n%s", report)
	}
	if !strings.Contains(report, "green: 1") {
		t.Fatalf("the probe row must pass — the pin never reaches the commands:\n%s", report)
	}
}

// TestRowTimeoutPinChangesEnvHash — the deadline pin rides the
// proofs' environment hash and never the commands' own environment.
func TestRowTimeoutPinChangesEnvHash(t *testing.T) {
	mountPorts()
	pin := "PUNCHTAPE_ROW_TIMEOUT=6"
	for _, kv := range RunEnv([]string{pin}) {
		if strings.HasPrefix(kv, "PUNCHTAPE_ROW_TIMEOUT=") {
			t.Fatalf("the deadline pin must never reach the commands: %q", kv)
		}
	}
	if EnvHash(nil) == EnvHash([]string{pin}) {
		t.Fatal("the deadline pin must move the proofs' environment hash")
	}
	// The moved hash stalemates the standing proof: the row re-earns
	// itself under the pin.
	root := newRoot(t)
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, fmt.Sprintf(`
work: the pinned copy
intents:
    - {lines: [D1], outcome: work, prose: "the tool and its check, told again"}
requirements:
    - id: R1
      lines: [D1]
      prose: the tool and its check
uncovered-lines:
    - {range: "AGENTS.md:1-9", reason: the machine's own hook, refs: [D1]}
    - {range: "out.txt:1-1", reason: the row's own output, refs: [D1]}
behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies the input, pinned by its bytes.
      claims:
        - src/in.txt:1-1
      covers: [R1]
      rows:
        - run: sh -c 'printf "%%s %%s\n" %s %s; cat src/in.txt > out.txt'
          expect:
            files:
                out.txt: ok
`, claimsTokenOf(t, root, "src/in.txt:1-1"), demandTokenOf("a checked tool")))
	runIn(t, root)
	writeZone(t, root, zone.EnvSlot, pin+"\n")
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "run 2") || !strings.Contains(report, "green: 1") {
		t.Fatalf("the moved environment hash must stale and re-earn the proof:\n%s", report)
	}
}

// TestSingleRecordsReadPerRun — one run reads the journal's records
// once: the epoch files answer one full pass plus the birth record's
// own single look.
func TestSingleRecordsReadPerRun(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	writeZone(t, root, zone.InboxSlot, "work: a second epoch\n")
	runIn(t, root)
	epochs := epochPaths(t, root)
	if len(epochs) < 2 {
		t.Fatalf("the probe needs moved epochs, got %d", len(epochs))
	}
	probe := &countFS{FS: fsys.OS{}, key: zone.JournalDir + "/epoch-"}
	ports.FSBind = probe
	ports.ExecBind = process.OS{}
	defer mountPorts()
	runIn(t, root)
	if probe.reads != len(epochs)+1 {
		t.Fatalf("the journal must read once a run (%d epochs): %d epoch reads", len(epochs), probe.reads)
	}
}

// TestNotesIndex — the notes ride the submission records and the
// fold's own accumulation: the render of the knowledge page never
// decodes a submission's document object.
func TestNotesIndex(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	writeZone(t, root, zone.InboxSlot, "work: the noted pass\nnotes:\n    - kind: work\n      text: the first note\n")
	runIn(t, root)
	records, err := journalRecords(t, root)
	if err != nil {
		t.Fatal(err)
	}
	var docAddr string
	for _, rec := range records {
		if rec.Kind == journal.KindSubmission && rec.Submission != nil {
			docAddr = rec.Submission.DocAddr
		}
	}
	if docAddr == "" {
		t.Fatal("the fixture needs a journaled submission")
	}
	// The store spells an object's path as objects/<first two>/<rest>:
	// the contiguous tail is the probe's own key.
	probe := &countFS{FS: fsys.OS{}, key: docAddr[2:]}
	ports.FSBind = probe
	ports.ExecBind = process.OS{}
	defer mountPorts()
	runIn(t, root)
	if probe.reads != 0 {
		t.Fatalf("the knowledge render must not decode submission documents: %d reads", probe.reads)
	}
	if state := zoneFile(t, root, zone.StateFile); !strings.Contains(state, "knowledge: 1 notes") {
		t.Fatalf("the notes must stand counted from the fold's own index:\n%s", state)
	}
}

// TestRowBytesIndex — the fold's row-bytes index answers the lessons
// and the drift witnesses without one read of the behavior objects.
func TestRowBytesIndex(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	stream, err := store.OpenStream(filepath.Join(root, zone.Dirname, zone.JournalDir))
	if err != nil {
		t.Fatal(err)
	}
	m, err := open(root, filepath.Join(root, zone.Dirname), stream)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.standing.Rows) == 0 {
		t.Fatal("the fixture needs a standing row fact")
	}
	probe := &countFS{FS: fsys.OS{}, key: zone.ObjectsDir}
	ports.FSBind = probe
	ports.ExecBind = process.OS{}
	defer mountPorts()
	for _, fact := range m.standing.Rows {
		if bytes := m.rowBytesOf(fact.Row); strings.HasPrefix(string(bytes), "<row ") {
			t.Fatalf("the index must answer the row's own bytes for %s", fact.Row)
		}
	}
	// The lesson path: a red fact healed green reads its two rows
	// through the index alone.
	var green journal.RowFact
	for digest := range m.standing.Rows {
		green = m.standing.Rows[digest]
	}
	red := green
	red.Status = journal.StatusRed
	red.Facts = []string{"the named reason"}
	m.standing.Rows[red.Key] = red
	if lessons := m.lessons([]journal.RowFact{green}, 9); len(lessons) != 1 {
		t.Fatalf("the healed row must fix its lesson, got %d", len(lessons))
	}
	if probe.reads != 0 {
		t.Fatalf("the index must answer without the behavior objects: %d reads", probe.reads)
	}
}

// TestClaimsReadTwicePerRun — the claims' one truth per run: over a
// whole run each claimed file is read at most twice — once ahead of
// the execution, once after it for the render — never again for the
// drift question, the trace or the next instruction. The walk's own
// economy answers from its verified cache.
func TestClaimsReadTwicePerRun(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	// The first run warms the walk's own cache; the probe then counts
	// the claims' own reads of the second run.
	runIn(t, root)
	probe := &countFS{FS: fsys.OS{}, key: "/src/in.txt"}
	ports.FSBind = probe
	defer mountPorts()
	runIn(t, root)
	if probe.reads > 2 {
		t.Fatalf("a claimed file is read at most twice per run — before and after the execution — got %d", probe.reads)
	}
}

// TestClaimsDigestComputedOnce — one computation of a behavior's
// claims facts reads each claimed file exactly once, however many
// ranges claim it.
func TestClaimsDigestComputedOnce(t *testing.T) {
	mountPorts()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	data := []byte("one\ntwo\nthree\n")
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	probe := &countFS{FS: fsys.OS{}, key: "/src/in.txt"}
	ports.FSBind = probe
	ports.ExecBind = process.OS{}
	defer mountPorts()
	m := &machine{root: root}
	facts := m.claimsFactsOf(doc.Behavior{Claims: []string{"src/in.txt:1-1", "src/in.txt:2-3"}})
	if probe.reads != 1 {
		t.Fatalf("two claims of one file read it once per computation: %d reads", probe.reads)
	}
	spell := "src/in.txt:1-1=" + fprint.LineRangeDigest(data, 1, 1) + "\n" +
		"src/in.txt:2-3=" + fprint.LineRangeDigest(data, 2, 3)
	if facts.digest != domain.Digest([]byte(spell)) {
		t.Fatalf("the digest must spell the sorted range law: %s", facts.digest)
	}
	if len(facts.ranges) != 2 ||
		facts.ranges[0].Claim != "src/in.txt:1-1" || facts.ranges[0].Digest12 != fprint.Token12(fprint.LineRangeDigest(data, 1, 1)) ||
		facts.ranges[1].Claim != "src/in.txt:2-3" || facts.ranges[1].Digest12 != fprint.Token12(fprint.LineRangeDigest(data, 2, 3)) {
		t.Fatalf("the ranges carry their own digests: %+v", facts.ranges)
	}
	// The next computation reads once again — one read per
	// computation, never one per claim.
	m.claimsFactsOf(doc.Behavior{Claims: []string{"src/in.txt:1-1", "src/in.txt:2-3"}})
	if probe.reads != 2 {
		t.Fatalf("one read per computation: %d reads", probe.reads)
	}
}

// TestRowTimeoutDefaultStands — without the pin the standing
// deadline of three hundred seconds stands; the pin names whole
// positive seconds or the run refuses by name.
func TestRowTimeoutDefaultStands(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	stream, err := store.OpenStream(filepath.Join(root, zone.Dirname, zone.JournalDir))
	if err != nil {
		t.Fatal(err)
	}
	m, err := open(root, filepath.Join(root, zone.Dirname), stream)
	if err != nil {
		t.Fatal(err)
	}
	if m.rowTimeout != defaultRowTimeout {
		t.Fatalf("a zone without the pin keeps the standing deadline, got %s", m.rowTimeout)
	}
	if d, err := RowTimeoutOf(nil); err != nil || d != defaultRowTimeout {
		t.Fatalf("no pin must answer the default: %s %v", d, err)
	}
	if d, err := RowTimeoutOf([]string{"PUNCHTAPE_ROW_TIMEOUT=7"}); err != nil || d != 7*time.Second {
		t.Fatalf("the pin names whole seconds: %s %v", d, err)
	}
	if _, err := RowTimeoutOf([]string{"PUNCHTAPE_ROW_TIMEOUT=soon"}); err == nil ||
		!strings.Contains(err.Error(), "the row deadline pin names whole seconds, a positive number") {
		t.Fatalf("a deadline that is not whole seconds must refuse by name: %v", err)
	}
}

// TestUniverseAndCacheFaultsAreLoud — the universe's own read, the
// walk cache's read and write, and an epoch past the first each
// fail loudly: an accelerator that cannot move never lies quietly.
func TestUniverseAndCacheFaultsAreLoud(t *testing.T) {
	// The universe file cannot be read past its existence.
	root := newRoot(t)
	writeZone(t, root, zone.UniverseSlot, "src\n")
	writeN(t, root, "src", 1)
	code, term := runCode(t, root, map[string]bool{"read:universe": true})
	if code != 1 || !strings.Contains(term, "universe") {
		t.Fatalf("an unreadable universe must refuse: %d %q", code, term)
	}
	// The walk cache cannot be read.
	root = newRoot(t)
	runIn(t, root)
	code, term = runCode(t, root, map[string]bool{"read:cache/tree": true})
	if code != 1 || !strings.Contains(term, "punchtape:") {
		t.Fatalf("an unreadable cache must refuse: %d %q", code, term)
	}
	// The walk cache cannot land its refreshed record.
	root = newRoot(t)
	runIn(t, root)
	writeN(t, root, "src", 1)
	code, term = runCode(t, root, map[string]bool{"write:cache/tree": true})
	if code != 1 || !strings.Contains(term, "tree cache:") {
		t.Fatalf("an unwritable cache must refuse: %d %q", code, term)
	}
	// An epoch past the first cannot be read at the open: the once
	// read is refused by name, never passed in silence.
	root = newBornRoot(t)
	writeZone(t, root, zone.InboxSlot, "work: a second epoch\n")
	runIn(t, root)
	code, term = runCode(t, root, map[string]bool{"read:epoch-00002": true})
	if code != 1 || !strings.Contains(term, "journal:") {
		t.Fatalf("an unreadable epoch must refuse: %d %q", code, term)
	}
}

// TestTreeCacheJunkLinesAreMisses — a cache line of any other shape
// is a miss for its path: the walk re-reads and rewrites the cache
// clean, never trusts a record it cannot parse.
func TestTreeCacheJunkLinesAreMisses(t *testing.T) {
	mountPorts()
	root := newRoot(t)
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("ok\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runIn(t, root)
	junk := "ghost.txt\t12\t34\n" + // four fields
		"\n" + // a blank line passes as a miss
		"src/in.txt\tnotanumber\t34\t1\t" + strings.Repeat("a", 64) + "\n" +
		"src/in.txt\t12\t34\t0\t" + strings.Repeat("b", 64) + "\n" + // no lines
		"src/in.txt\t12\t34\t1\t\n" // empty digest
	writeZone(t, root, zone.CacheTreeDir, junk)
	runIn(t, root)
	cache := zoneFile(t, root, zone.CacheTreeDir)
	well := regexp.MustCompile(`^[^\t]+\t\d+\t-?\d+\t[1-9]\d*\t[0-9a-f]{64}$`)
	for _, line := range strings.Split(strings.TrimRight(cache, "\n"), "\n") {
		if !well.MatchString(line) {
			t.Fatalf("the rewritten cache carries only its own records: %q\n%s", line, cache)
		}
	}
	if strings.Contains(cache, "ghost.txt") {
		t.Fatalf("a junk path never survives the walk:\n%s", cache)
	}
}

// TestBadDeadlinePinRefusesTheRun — a deadline pin that names no
// whole seconds refuses the run by name, ahead of any work.
func TestBadDeadlinePinRefusesTheRun(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	writeZone(t, root, zone.EnvSlot, "PUNCHTAPE_ROW_TIMEOUT=soon\n")
	var term strings.Builder
	code := Run(root, &term)
	if code != 1 || !strings.Contains(term.String(), "the row deadline pin names whole seconds, a positive number") {
		t.Fatalf("a malformed deadline pin must refuse by name: %d %q", code, term.String())
	}
}

// TestDriftWithoutAWitnessNamesTheCurrentRun — a move the rows never
// measured (a broken build freezes them) still names a run: the
// current one, the run that sees the moved tree now.
func TestDriftWithoutAWitnessNamesTheCurrentRun(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if state := zoneFile(t, root, zone.StateFile); !strings.Contains(state, "verdict: #1") {
		t.Fatalf("the fixture must close its verdict:\n%s", state)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("moved"), 0o644); err != nil {
		t.Fatal(err)
	}
	// The corpus's own shape under a broken build: the rows freeze,
	// the run record carries no row fact, the drift has no witness.
	doc := "work: the frozen copy\nbuild: false\nmandate: [{line: D1, target: tool/copy}]\n" +
		zoneFile(t, root, zone.CorpusFile)
	writeZone(t, root, zone.InboxSlot, doc)
	runIn(t, root)
	if state := zoneFile(t, root, zone.StateFile); !strings.Contains(state, "open (drift since run 2)") {
		t.Fatalf("a witness-less drift must name the current run:\n%s", state)
	}
}

// TestUniversePhantomRootRefused — a declared root must name at
// least one file the tree carries: a spelling that matches nothing
// — the yaml-list dash, a path that does not stand, a directory of
// empty directories, a directory of bookkeeping alone — is refused
// loudly with its own line in the declaration, never a silent loss
// of coverage.
func TestUniversePhantomRootRefused(t *testing.T) {
	mountPorts()
	root := newRoot(t)
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("ok\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run := func(universe string) (int, string) {
		t.Helper()
		writeZone(t, root, zone.UniverseSlot, universe)
		var term strings.Builder
		code := Run(root, &term)
		return code, term.String()
	}

	// The yaml-list spelling: a dash-prefixed line passes the path
	// shape and matches no file — refused by its line.
	code, term := run("# the counted roots\n- src\n")
	if code != 1 || !strings.Contains(term,
		`universe refused — .punchtape/universe, line 2: "- src" — the law: a root carries at least one file of the tree`) {
		t.Fatalf("the yaml-list root must refuse by its line: %d %q", code, term)
	}
	// A path that does not stand.
	code, term = run("ghost/dir\n")
	if code != 1 || !strings.Contains(term,
		`universe refused — .punchtape/universe, line 1: "ghost/dir" — the law: a root carries at least one file of the tree`) {
		t.Fatalf("a standing-nowhere root must refuse: %d %q", code, term)
	}
	// A directory of empty directories.
	if err := os.MkdirAll(filepath.Join(root, "empty", "deeper"), 0o755); err != nil {
		t.Fatal(err)
	}
	code, term = run("empty\n")
	if code != 1 || !strings.Contains(term,
		`universe refused — .punchtape/universe, line 1: "empty" — the law: a root carries at least one file of the tree`) {
		t.Fatalf("an empty tree of directories must refuse: %d %q", code, term)
	}
	// A directory whose only child is version-control bookkeeping.
	if err := os.MkdirAll(filepath.Join(root, "book", ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	code, term = run("book\n")
	if code != 1 || !strings.Contains(term, `line 1: "book"`) {
		t.Fatalf("a directory of bookkeeping alone must refuse: %d %q", code, term)
	}

	// The lawful spellings stand: a nested carrier and a file root.
	writeZone(t, root, zone.UniverseSlot, "src\n")
	if code, term = run("src\n"); code != 0 {
		t.Fatalf("a carrying root must stand: %d %q", code, term)
	}
	writeZone(t, root, zone.UniverseSlot, "src/in.txt\n")
	if code, term = run("src/in.txt\n"); code != 0 {
		t.Fatalf("a file root must stand: %d %q", code, term)
	}
	// A carrier found through an empty sibling: the descent walks
	// past directories that carry nothing and answers by the nested
	// one that does.
	if err := os.MkdirAll(filepath.Join(root, "pkg", "none"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "pkg", "inner"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "pkg", "inner", "real.txt"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.UniverseSlot, "pkg\n")
	if code, term = run("pkg\n"); code != 0 {
		t.Fatalf("a nested carrier must stand past its empty sibling: %d %q", code, term)
	}

	// The probe's own failures are loud, never a silent empty answer.
	writeZone(t, root, zone.UniverseSlot, "ghost2\n")
	code, term = runCode(t, root, map[string]bool{"stat:ghost2": true})
	if code != 1 || !strings.Contains(term, "permission") {
		t.Fatalf("an unstattable root must fail loudly: %d %q", code, term)
	}
	writeZone(t, root, zone.UniverseSlot, "src\n")
	code, term = runCode(t, root, map[string]bool{"readdir:src": true})
	if code != 1 || !strings.Contains(term, "permission") {
		t.Fatalf("an unreadable root directory must fail loudly: %d %q", code, term)
	}
	mountPorts()
}
