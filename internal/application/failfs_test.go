// The failure battery: every honest error path of the machine — a
// zone that cannot be read, a journal that cannot be written, a
// launch that never started — exercised through a filesystem that
// fails exactly where the case needs. The machine's answer to a
// failure is a named refusal or a non-zero exit, never a guess.
package application

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neurophant/punchtape/internal/domain/journal"
	"github.com/neurophant/punchtape/internal/domain/zone"
	"github.com/neurophant/punchtape/internal/infrastructure/fsys"
	"github.com/neurophant/punchtape/internal/infrastructure/process"
	"github.com/neurophant/punchtape/internal/ports"
)

// failFS — the real OS filesystem with named failure points: an
// operation on a path containing the key answers the named error.
// A key ending in "!after-write" arms only after the first
// successful write — acceptance lands, what follows fails.
type failFS struct {
	ports.FS
	fail        map[string]bool // "op:substring" → fail
	contentFail map[string]bool // "append-content:marker" → fail that append
	wroteOnce   bool
}

func newFailFS(fail map[string]bool) *failFS {
	f := &failFS{FS: fsys.OS{}, fail: map[string]bool{}}
	for key, on := range fail {
		if marker, ok := strings.CutPrefix(key, "append-content:"); ok {
			if f.contentFail == nil {
				f.contentFail = map[string]bool{}
			}
			f.contentFail[marker] = on
			continue
		}
		f.fail[key] = on
	}
	return f
}

func (f *failFS) fails(op, path string) bool {
	for key := range f.fail {
		parts := strings.SplitN(key, ":", 2)
		armed := parts[0] == op
		if strings.HasSuffix(parts[1], "!after-write") {
			if !f.wroteOnce {
				continue
			}
			parts[1] = strings.TrimSuffix(parts[1], "!after-write")
		}
		if armed && strings.Contains(path, parts[1]) {
			return true
		}
	}
	return false
}

func (f *failFS) Stat(p string) (os.FileInfo, error) {
	if f.fails("stat", p) {
		return nil, os.ErrPermission
	}
	return f.FS.Stat(p)
}

func (f *failFS) ReadFile(p string) ([]byte, error) {
	if f.fails("read", p) {
		return nil, os.ErrPermission
	}
	return f.FS.ReadFile(p)
}

func (f *failFS) ReadDir(p string) ([]os.DirEntry, error) {
	if f.fails("readdir", p) {
		return nil, os.ErrPermission
	}
	return f.FS.ReadDir(p)
}

func (f *failFS) WriteAtomic(p string, data []byte) error {
	if f.fails("write", p) {
		return os.ErrPermission
	}
	f.wroteOnce = true
	return f.FS.WriteAtomic(p, data)
}

func (f *failFS) AppendFile(p string, data []byte) error {
	if f.fails("append", p) {
		return os.ErrPermission
	}
	for marker, on := range f.contentFail {
		if on && bytes.Contains(data, []byte(marker)) {
			return os.ErrPermission
		}
	}
	return f.FS.AppendFile(p, data)
}

func (f *failFS) Walk(root string, fn filepath.WalkFunc) error {
	if f.fails("walk", root) {
		return os.ErrPermission
	}
	return f.FS.Walk(root, fn)
}

func (f *failFS) Remove(p string) error {
	if f.fails("remove", p) {
		return os.ErrPermission
	}
	return f.FS.Remove(p)
}

func (f *failFS) Flock(p string) (func() error, error) {
	if f.fails("flock", p) {
		return nil, os.ErrPermission
	}
	if f.fails("flock-release", p) {
		return func() error { return os.ErrClosed }, nil
	}
	return f.FS.Flock(p)
}

func (f *failFS) RemoveAll(p string) error {
	if f.fails("removeall", p) {
		return os.ErrPermission
	}
	return f.FS.RemoveAll(p)
}

func (f *failFS) Mkdir(p string) error {
	if f.fails("mkdir-exist", p) {
		return os.ErrExist
	}
	if f.fails("mkdir", p) {
		return os.ErrPermission
	}
	return f.FS.Mkdir(p)
}

func (f *failFS) MkdirAll(p string) error {
	if f.fails("mkdirall", p) {
		return os.ErrPermission
	}
	return f.FS.MkdirAll(p)
}

// runCode — one run under a failing filesystem; the production
// process port stays.
func runCode(t *testing.T, root string, fail map[string]bool) (int, string) {
	t.Helper()
	ports.FSBind = newFailFS(fail)
	ports.ExecBind = process.OS{}
	defer mountPorts()
	var term strings.Builder
	code := Run(root, &term)
	return code, term.String()
}

func TestZoneDamageRefusesTheRun(t *testing.T) {
	root := newBornRoot(t)
	// The journal directory cannot be statted: the zone answers
	// damage, the run refuses.
	code, term := runCode(t, root, map[string]bool{"stat:journal": true})
	if code != 1 || !strings.Contains(term, "zone damaged") {
		t.Fatalf("code %d term %q", code, term)
	}
	// The journal cannot be opened: the named refusal.
	code, term = runCode(t, root, map[string]bool{"readdir:journal": true})
	if code != 1 || !strings.Contains(term, "punchtape:") {
		t.Fatalf("code %d term %q", code, term)
	}
	// The stream cannot be read at the birth.
	fresh := t.TempDir()
	writeZone(t, fresh, zone.DemandsSlot, "goal")
	code, term = runCode(t, fresh, map[string]bool{"read:demands.md": true})
	if code != 1 || !strings.Contains(term, "birth failed") {
		t.Fatalf("code %d term %q", code, term)
	}
	// The birth's standing files cannot be written.
	fresh = t.TempDir()
	writeZone(t, fresh, zone.DemandsSlot, "goal")
	code, term = runCode(t, fresh, map[string]bool{"write:contract.md": true})
	if code != 1 || !strings.Contains(term, "birth failed") {
		t.Fatalf("code %d term %q", code, term)
	}
}

func TestSubmissionAndRunFailuresAreNamed(t *testing.T) {
	root := newBornRoot(t)
	green := `
work: the copy, harder
mandate: [{line: D1, target: tool/copy}]
behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies harder.
      claims:
        - src/in.txt:1-1
      covers: [R1]
      rows:
        - run: cat src/in.txt > out.txt
          expect:
            files:
                out.txt: ok
`
	// The inbox cannot be read.
	writeZone(t, root, zone.InboxSlot, green)
	code, term := runCode(t, root, map[string]bool{"read:inbox.yaml": true})
	if code != 1 || !strings.Contains(term, "punchtape:") {
		t.Fatalf("inbox read: code %d term %q", code, term)
	}
	// The objects cannot take the submission.
	writeZone(t, root, zone.InboxSlot, green)
	code, term = runCode(t, root, map[string]bool{"write:objects": true})
	if code != 1 {
		t.Fatalf("objects write: code %d term %q", code, term)
	}
	// The epoch rotation cannot happen.
	writeZone(t, root, zone.InboxSlot, green)
	code, term = runCode(t, root, map[string]bool{"write:epoch-": true})
	if code != 1 {
		t.Fatalf("rotation: code %d term %q", code, term)
	}
	// The journal append cannot happen.
	root2 := newBornRoot(t)
	writeZone(t, root2, zone.InboxSlot, green)
	code, term = runCode(t, root2, map[string]bool{"append:epoch-": true})
	if code != 1 {
		t.Fatalf("append: code %d term %q", code, term)
	}
	// The tree cannot be walked.
	root3 := newBornRoot(t)
	os.WriteFile(filepath.Join(root3, "src", "in.txt"), []byte("changed"), 0o644)
	code, term = runCode(t, root3, map[string]bool{"walk:" + root3: true})
	if code != 1 {
		t.Fatalf("walk: code %d term %q", code, term)
	}
}

// TestAgentsBlockDamageIsLoud — the machine's hook outside the zone:
// damage by hand and an unreadable file are loud refusals of the
// run, never a silence that leaves a broken or stale hook standing.
func TestAgentsBlockDamageIsLoud(t *testing.T) {
	root := newBornRoot(t)
	path := filepath.Join(root, agentsFile)
	// One marker without its pair: the machine does not guess, does
	// not touch the file, and says so.
	os.WriteFile(path, []byte("words\n"+blockOpen+"\nmore words\n"), 0o644)
	code, term := runCode(t, root, nil)
	if code != 1 || !strings.Contains(term, "marker without its pair") {
		t.Fatalf("the damaged hook must refuse loudly: code %d term %q", code, term)
	}
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), "more words") {
		t.Fatalf("an unpaired marker must leave the file alone")
	}
	// The file cannot be read: the run names it instead of skipping
	// the hook in silence.
	code, term = runCode(t, root, map[string]bool{"read:AGENTS.md": true})
	if code != 1 || !strings.Contains(term, "punchtape:") {
		t.Fatalf("the unreadable hook must refuse loudly: code %d term %q", code, term)
	}
	// A missing block that cannot land its write: the same loudness.
	root = newBornRoot(t)
	if err := os.Remove(filepath.Join(root, agentsFile)); err != nil {
		t.Fatal(err)
	}
	code, term = runCode(t, root, map[string]bool{"write:AGENTS.md": true})
	if code != 1 || !strings.Contains(term, "punchtape:") {
		t.Fatalf("the unwritable hook must refuse loudly: code %d term %q", code, term)
	}
}

// TestViewWriteFailuresAreLoud — every screen of the machine's face
// fails loudly: a view that cannot land is a named refusal of the
// run, never a silently stale page pointing the reader at a lie.
func TestViewWriteFailuresAreLoud(t *testing.T) {
	// One changed submission: a new note moves knowledge, a new
	// expectation moves the row (it re-runs, the report moves), the
	// prose moves the corpus, and every screen moves with them.
	moving := `
work: a document that moves every screen
mandate: [{line: D1, target: tool/copy}]
notes:
    - kind: work
      text: a note that moves the knowledge page
behaviors:
    - id: tool/copy
      name: The Copy
      prose: Copies anew.
      claims:
        - src/in.txt:1-1
      covers: [R1]
      rows:
        - run: cat src/in.txt > out.txt
          expect:
            files:
                out.txt: o
`
	for _, tc := range []struct {
		name string
		fail map[string]bool
		want string
		// stale names a zone file stood stale ahead of the run, so a
		// face's refresh write actually fires under its failure point.
		stale string
	}{
		{"the report view", map[string]bool{"write:.punchtape/report.md": true}, "report view:", ""},
		{"the state view", map[string]bool{"write:.punchtape/state.md": true}, "state view:", ""},
		{"the knowledge view", map[string]bool{"write:.punchtape/knowledge.md": true}, "knowledge view:", ""},
		{"the spec page", map[string]bool{"write:.punchtape/spec": true}, "spec view:", ""},
		{"the spec index", map[string]bool{"write:.punchtape/spec/index.md": true}, "spec view:", ""},
		{"the spec directory", map[string]bool{"mkdirall:.punchtape/spec": true}, "spec view:", ""},
		{"the spec walk", map[string]bool{"walk:.punchtape/spec": true}, "spec view:", ""},
		{"the corpus image", map[string]bool{"write:.punchtape/corpus.yaml": true}, "corpus view:", ""},
		{"the contract face", map[string]bool{"write:.punchtape/contract.md": true}, "contract view:", "contract.md"},
		{"the template face", map[string]bool{"write:.punchtape/template.yaml": true}, "template view:", "template.yaml"},
		{"the reference face", map[string]bool{"write:.punchtape/reference.md": true}, "reference view:", "reference.md"},
		{"the scratch clear", map[string]bool{"removeall:.punchtape/tmp": true}, "tmp clear:", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := newBornRoot(t)
			if tc.stale != "" {
				if err := os.WriteFile(filepath.Join(root, zone.Dirname, tc.stale), []byte("stale\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			writeZone(t, root, zone.InboxSlot, moving)
			code, term := runCode(t, root, tc.fail)
			if code != 1 || !strings.Contains(term, tc.want) {
				t.Fatalf("code %d term %q", code, term)
			}
		})
	}
	// The refusal path: the refusal is journaled first — a report
	// screen that cannot land never eats the journaled fact.
	root := newBornRoot(t)
	writeZone(t, root, zone.InboxSlot, "\tbroken: [")
	code, term := runCode(t, root, map[string]bool{"write:report.md": true})
	if code != 1 || !strings.Contains(term, "report view:") {
		t.Fatalf("the refusal's report screen: code %d term %q", code, term)
	}
	records, err := journalRecords(t, root)
	if err != nil {
		t.Fatal(err)
	}
	var refusals int
	for _, rec := range records {
		if rec.Kind == journal.KindRefusal {
			refusals++
		}
	}
	if refusals != 1 {
		t.Fatalf("the refusal fact must stand in the journal despite the screen, got %d", refusals)
	}
	// A retired behavior's page that cannot leave: the spec tree is
	// an image of the standing corpus, a page that stays is a lie.
	root = newBornRoot(t)
	writeZone(t, root, zone.InboxSlot, "work: retiring the copy\nmandate: [{line: D1, target: tool/copy}]\nremove: [tool/copy]\n")
	code, term = runCode(t, root, map[string]bool{"remove:.punchtape/spec/": true})
	if code != 1 || !strings.Contains(term, "spec view:") {
		t.Fatalf("the retired page: code %d term %q", code, term)
	}
}

// TestSpecWalkFaultsAreLoud — the spec tree's own walk over an
// unreadable page directory fails loudly: a spec tree built on half
// a directory is a lie.
func TestSpecWalkFaultsAreLoud(t *testing.T) {
	// The spec tree's own walk over an unreadable directory of a
	// retired behavior (no page writes stand in the way): the retire
	// sweep says so instead of passing in silence.
	root2 := newBornRoot(t)
	writeZone(t, root2, zone.InboxSlot, "work: retiring the copy\nmandate: [{line: D1, target: tool/copy}]\nremove: [tool/copy]\n")
	runIn(t, root2)
	specTool := filepath.Join(root2, zone.Dirname, zone.SpecDir, "tool")
	if err := os.Chmod(specTool, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(specTool, 0o755) })
	var term strings.Builder
	if code := Run(root2, &term); code != 1 || !strings.Contains(term.String(), "spec view:") {
		t.Fatalf("the spec walk must fail loudly: code %d term %q", code, term.String())
	}
}

func TestVerdictAppendFailureAndSecondVerdict(t *testing.T) {
	root := newBornRoot(t)
	// A red row: the verdict says not yet.
	writeZone(t, root, zone.InboxSlot, fmt.Sprintf(`
work: the red row first
mandate: [{line: D1, target: tool/copy}]
behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies.
      claims:
        - src/in.txt:1-1
      covers: [R1]
      rows:
        - run: sh -c 'printf "%%s %%s\n" %s %s; cat src/in.txt > out.txt'
          expect:
            files:
                out.txt: NOPE
`, claimsTokenOf(t, root, "src/in.txt:1-1"), demandTokenOf("a checked tool")))
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if !strings.Contains(zoneFile(t, root, zone.ReportFile), "verdict: not yet") {
		t.Fatalf("a red corpus must say not yet")
	}
	// The row heals; the verdict's journal append fails: the verdict
	// simply does not land, the run still answers.
	writeZone(t, root, zone.InboxSlot, fmt.Sprintf(`
work: the heal of the copy
mandate: [{line: D1, target: tool/copy}]
behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies.
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
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runCode(t, root, map[string]bool{"append:epoch-": true})
	if strings.Contains(zoneFile(t, root, zone.StateFile), "verdict: #") {
		t.Fatalf("a failed append must not fake a verdict")
	}
	// Without the failure the verdict lands; a second full cycle —
	// new work, green, the completion act — records the second
	// verdict.
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if !strings.Contains(zoneFile(t, root, zone.StateFile), "verdict: #1") {
		t.Fatalf("the first point of truth must land")
	}
	writeZone(t, root, zone.InboxSlot, fmt.Sprintf(`
work: the second cycle's new row
mandate: [{line: D1, target: tool/copy}]
uncovered-lines:
    - {range: "AGENTS.md:1-9", reason: the machine's own hook, refs: [D1]}
    - {range: "out.txt:1-1", reason: the row's own output, refs: [D1]}
    - {range: "out2.txt:1-1", reason: the row's own output, refs: [D1]}
behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies.
      claims:
        - src/in.txt:1-1
      covers: [R1]
      rows:
        - run: sh -c 'printf "%%s %%s\n" %s %s; cat src/in.txt > out2.txt'
          expect:
            files:
                out2.txt: ok
`, claimsTokenOf(t, root, "src/in.txt:1-1"), demandTokenOf("a checked tool")))
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "verdict: #2") {
		t.Fatalf("the second point of truth must be recorded, state:\n%s", state)
	}
}

// TestDoneActJournalsBeforeTheInboxLeaves — the journal carries the
// act's outcome before the document leaves the slot: a run that
// dies at the removal leaves the verdict and the completion
// standing in the journal and the done document standing in the
// slot. The next run takes the act again — every act is a point of
// truth, the duplicate names the two takings honestly, and the
// words never vanish.
func TestDoneActJournalsBeforeTheInboxLeaves(t *testing.T) {
	root := newBornRoot(t)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	code, _ := runCode(t, root, map[string]bool{"remove:.punchtape/inbox.yaml": true})
	if code != 1 {
		t.Fatalf("the failed removal must fail the run, code %d", code)
	}
	records, err := journalRecords(t, root)
	if err != nil {
		t.Fatal(err)
	}
	var verdicts, completions int
	for _, rec := range records {
		if rec.Kind == journal.KindVerdict {
			verdicts++
		}
		if rec.Kind == journal.KindCompletion {
			completions++
		}
	}
	if verdicts != 1 || completions != 1 {
		t.Fatalf("the dead run must leave its verdict and completion journaled, got %d/%d", verdicts, completions)
	}
	if _, err := os.Stat(filepath.Join(root, zone.Dirname, zone.InboxSlot)); err != nil {
		t.Fatalf("the refused removal must leave the inbox in place: %v", err)
	}
	// The next run takes the standing document again: one more point
	// of truth, the slot empties, no word is lost.
	runIn(t, root)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "verdict: #2") {
		t.Fatalf("the re-taken act must record its own point:\n%s", state)
	}
	if _, err := os.Stat(filepath.Join(root, zone.Dirname, zone.InboxSlot)); !os.IsNotExist(err) {
		t.Fatal("the re-taken act must take the document out of the slot")
	}
}

func TestBirthFaultsAreNamed(t *testing.T) {
	fresh := func() string {
		dir := t.TempDir()
		writeZone(t, dir, zone.DemandsSlot, "goal")
		return dir
	}
	// The journal's mkdir cannot happen at all.
	code, term := runCode(t, fresh(), map[string]bool{"mkdir:journal": true})
	if code != 1 || !strings.Contains(term, "birth failed") {
		t.Fatalf("mkdir: code %d term %q", code, term)
	}
	// The journal cannot open.
	code, term = runCode(t, fresh(), map[string]bool{"readdir:journal": true})
	if code != 1 || !strings.Contains(term, "birth failed") {
		t.Fatalf("open stream: code %d term %q", code, term)
	}
	// The first record cannot land.
	code, term = runCode(t, fresh(), map[string]bool{"append:epoch-": true})
	if code != 1 || !strings.Contains(term, "birth failed") {
		t.Fatalf("append: code %d term %q", code, term)
	}
	// The genesis link cannot land: a birth without its chain is a
	// birth the machine refuses to own.
	code, term = runCode(t, fresh(), map[string]bool{"append-content:kind: demand-append": true})
	if code != 1 || !strings.Contains(term, "birth failed") {
		t.Fatalf("genesis link: code %d term %q", code, term)
	}
	// The reference page's own birth write cannot land: the birth
	// names the file it could not write.
	code, term = runCode(t, fresh(), map[string]bool{"write:reference.md": true})
	if code != 1 || !strings.Contains(term, "birth failed") || !strings.Contains(term, "reference.md") {
		t.Fatalf("reference write: code %d term %q", code, term)
	}
	// A parallel winner of the mkdir: this run continues as an
	// ordinary pass over a living instance.
	code, term = runCode(t, fresh(), map[string]bool{"mkdir-exist:journal": true})
	if code != 0 || !strings.Contains(term, "run 0") {
		t.Fatalf("a lost mkdir race must continue: code %d term %q", code, term)
	}
}

func TestJournalStreamFaultsAreNamed(t *testing.T) {
	root := newBornRoot(t)
	// The epoch file cannot be read.
	code, term := runCode(t, root, map[string]bool{"read:epoch-": true})
	if code != 1 || !strings.Contains(term, "punchtape:") {
		t.Fatalf("read epoch: code %d term %q", code, term)
	}
	// A run record cannot land: the append's directory step fails.
	os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("moved"), 0o644)
	code, term = runCode(t, root, map[string]bool{"mkdirall:journal": true})
	if code != 1 {
		t.Fatalf("append mkdirall: code %d term %q", code, term)
	}
	// A submission's rotation cannot open its directory step.
	writeZone(t, root, zone.InboxSlot, `
work: the moved copy
mandate: [{line: D1, target: tool/copy}]
behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies.
      claims:
        - src/in.txt:1-1
      covers: [R1]
      rows:
        - run: cat src/in.txt > out.txt
          expect:
            files:
                out.txt: moved
`)
	os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("moved"), 0o644)
	code, term = runCode(t, root, map[string]bool{"mkdirall:journal": true})
	if code != 1 {
		t.Fatalf("rotate mkdirall: code %d term %q", code, term)
	}
}

func TestTreeFaultsDoNotBlindTheFingerprints(t *testing.T) {
	root := newBornRoot(t)
	// A claimed file that cannot be read spells absent, honestly: the
	// claims digest moves, the row re-executes, the run answers — a
	// read fault is data, never blindness.
	code, fault := runCode(t, root, map[string]bool{"stat:src/in.txt": true})
	if code != 0 {
		t.Fatalf("stat: code %d term %q", code, fault)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("ok anew"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, fault = runCode(t, root, map[string]bool{"read:src/in.txt": true})
	if code != 0 {
		t.Fatalf("read: code %d term %q", code, fault)
	}
}

func TestAcceptFaultsAreNamed(t *testing.T) {
	root := newBornRoot(t)
	// The document object lands, the behavior object cannot: the
	// acceptance refuses, the standing keeps its truth.
	writeZone(t, root, zone.InboxSlot, `
work: the copy, more
mandate: [{line: D1, target: tool/copy}]
behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies more.
      claims:
        - src/in.txt:1-1
      covers: [R1]
      rows:
        - run: cat src/in.txt > out.txt
          expect:
            files:
                out.txt: ok
`)
	code, term := runCode(t, root, map[string]bool{"write:objects!after-write": true})
	if code != 1 {
		t.Fatalf("the behavior object's store fault must refuse: code %d term %q", code, term)
	}
}

func TestRefusalJournalAppendFailureIsNamed(t *testing.T) {
	root := newBornRoot(t)
	// The refusal's own journaling fails: the run answers non-zero —
	// a refusal the machine cannot record is not a refusal it had.
	writeZone(t, root, zone.InboxSlot, "\tbroken: [")
	code, term := runCode(t, root, map[string]bool{"append:epoch-": true})
	if code != 1 || !strings.Contains(term, "punchtape:") {
		t.Fatalf("the failed refusal journaling must refuse the run, code %d term %q", code, term)
	}
}

// TestAcceptanceStreamFactFailuresAreNamed — the words' own fact
// rides beside the accepted submission or the run refuses: a stream
// fact the machine cannot journal is not a fact it had, and the
// inbox stands for the next run.
func TestAcceptanceStreamFactFailuresAreNamed(t *testing.T) {
	// The demand-append beside an accepted submission.
	root := newBornRoot(t)
	writeZone(t, root, zone.DemandsSlot, "a checked tool\na late wish\n")
	writeZone(t, root, zone.InboxSlot, `work: the late wish
intents:
    - {lines: [D2], outcome: mark, prose: the late wish orders nothing yet}
not-a-requirement:
    - lines: [D2]
      reason: not new work
`)
	code, _ := runCode(t, root, map[string]bool{"append-content:kind: demand-append": true})
	if code != 1 {
		t.Fatalf("the unjournaled link must fail the run: %d", code)
	}
	if _, err := os.Stat(filepath.Join(root, zone.Dirname, zone.InboxSlot)); err != nil {
		t.Fatalf("the unjournaled link keeps the submission in place: %v", err)
	}
	// The signature beside an accepted submission.
	root = newBornRoot(t)
	writeZone(t, root, zone.InboxSlot, `
work: a noted pass
notes:
    - kind: work
      text: a note
`)
	code, _ = runCode(t, root, map[string]bool{"append-content:kind: no-new-demands": true})
	if code != 1 {
		t.Fatalf("the unjournaled signature must fail the run: %d", code)
	}
	if _, err := os.Stat(filepath.Join(root, zone.Dirname, zone.InboxSlot)); err != nil {
		t.Fatalf("the unjournaled signature keeps the submission in place: %v", err)
	}
}

// TestCompletionWalkFaultIsLoud — the totality gate's own walk over
// the tree fails loudly: a verdict claimed over a tree the machine
// could not read is a lie.
func TestCompletionWalkFaultIsLoud(t *testing.T) {
	root := newBornRoot(t)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	code, _ := runCode(t, root, map[string]bool{"walk:" + root: true})
	if code != 1 {
		t.Fatalf("the unreadable tree must refuse the completion: %d", code)
	}
}

// TestRefusalStateViewFaultIsLoud — the refusal path's own state
// write fails loudly: a queue the state cannot name is a stale face.
func TestRefusalStateViewFaultIsLoud(t *testing.T) {
	root := newBornRoot(t)
	writeZone(t, root, zone.InboxSlot, "\tbroken: [")
	code, term := runCode(t, root, map[string]bool{"write:.punchtape/state.md": true})
	if code != 1 || !strings.Contains(term, "state view:") {
		t.Fatalf("the refusal's state screen must fail loudly: code %d term %q", code, term)
	}
}

// TestExecutionFaultsAreNamed — the machine's own executions under
// failure: a launch that cannot be reaped journals its fact, a long
// output that cannot reach the object store fails the run, a fact
// that cannot journal fails the run.
func TestExecutionFaultsAreNamed(t *testing.T) {
	root := newBornRoot(t)
	ports.FSBind = fsys.OS{}
	ports.ExecBind = fakeWaitExec{}
	defer mountPorts()
	writeZone(t, root, zone.DemandsSlot, "a checked tool\nrun the failing probe\n")
	writeZone(t, root, zone.InboxSlot, `work: the execution that cannot be reaped
intents:
    - {lines: [D2], outcome: execute, prose: the probe's own output answers, run: "printf hi"}
`)
	runIn(t, root)
	records, err := journalRecords(t, root)
	if err != nil {
		t.Fatal(err)
	}
	var reaped bool
	for _, rec := range records {
		if rec.Kind == journal.KindExecution && rec.Execution != nil && rec.Execution.Exit == -1 {
			reaped = true
		}
	}
	if !reaped {
		t.Fatalf("the unreapable execution must journal its minus-one exit")
	}
	// The long output's object cannot land: the run says so. A new
	// line of the stream carries the new command — an outcome is
	// declared once, by the intent of its own line.
	mountPorts()
	writeZone(t, root, zone.DemandsSlot, "a checked tool\nrun the failing probe\nrun the long probe\n")
	writeZone(t, root, zone.InboxSlot, `
work: the long execution
intents:
    - {lines: [D3], outcome: execute, prose: the long output answers, run: "awk 'BEGIN{for(i=0;i<5000;i++)printf \"x\"}'"}
`)
	code, term := runCode(t, root, map[string]bool{"write:objects!after-write": true})
	if code != 1 || !strings.Contains(term, "cas put") {
		t.Fatalf("the unlandable long output must fail the run: %d %q", code, term)
	}
	// The fact itself cannot journal: the run says so.
	writeZone(t, root, zone.DemandsSlot, "a checked tool\nrun the failing probe\nrun the long probe\nrun the quiet probe\n")
	writeZone(t, root, zone.InboxSlot, `
work: the execution whose fact cannot journal
intents:
    - {lines: [D4], outcome: execute, prose: the command answers, run: "printf hi"}
`)
	code, term = runCode(t, root, map[string]bool{"append-content:kind: execution-fact": true})
	if code != 1 || !strings.Contains(term, "journal write:") {
		t.Fatalf("the unjournaled execution fact must fail the run: %d %q", code, term)
	}
}

// TestTraceViewWriteFaultIsLoud — a trace page that cannot land
// fails the run: a stale trace is a stale face.
func TestTraceViewWriteFaultIsLoud(t *testing.T) {
	root := newBornRoot(t)
	// A new stream line moves the trace's own content: the page must
	// land again — and cannot.
	writeZone(t, root, zone.DemandsSlot, "a checked tool\na later word\n")
	code, term := runCode(t, root, map[string]bool{"write:trace.md": true})
	if code != 1 || !strings.Contains(term, "trace view:") {
		t.Fatalf("the unlandable trace must fail the run: %d %q", code, term)
	}
}

// TestHolesViewWriteFaultIsLoud — the ledger that cannot land over
// changed holes fails the run inside the state's own write.
func TestHolesViewWriteFaultIsLoud(t *testing.T) {
	root := newBornRoot(t)
	if err := os.WriteFile(filepath.Join(root, "junk.txt"), []byte("loose\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, term := runCode(t, root, map[string]bool{"write:holes.md": true})
	if code != 1 || !strings.Contains(term, "state view:") {
		t.Fatalf("the unlandable ledger must fail the state's write: %d %q", code, term)
	}
}

// TestZoneLockFaultsAreNamed — the zone lock's own faults are loud:
// a take that cannot happen refuses the run, a release that cannot
// land says so on the terminal of a finished run.
func TestZoneLockFaultsAreNamed(t *testing.T) {
	root := newRoot(t)
	code, term := runCode(t, root, map[string]bool{"flock:run.lock": true})
	if code != 1 || !strings.Contains(term, "punchtape: zone lock:") {
		t.Fatalf("the untakeable lock must refuse the run, code %d:\n%s", code, term)
	}
	code, term = runCode(t, root, map[string]bool{"flock-release:run.lock": true})
	if code != 0 || !strings.Contains(term, "punchtape: zone lock release:") {
		t.Fatalf("the failed release must say so on a finished run, code %d:\n%s", code, term)
	}
}
