// The failure battery: every honest error path of the machine — a
// zone that cannot be read, a journal that cannot be written, a
// launch that never started — exercised through a filesystem that
// fails exactly where the case needs. The machine's answer to a
// failure is a named refusal or a non-zero exit, never a guess.
package application

import (
	"bytes"
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
	// The wish cannot be read at the birth.
	fresh := t.TempDir()
	writeZone(t, fresh, zone.WishSlot, "goal")
	code, term = runCode(t, fresh, map[string]bool{"read:wish.md": true})
	if code != 1 || !strings.Contains(term, "birth failed") {
		t.Fatalf("code %d term %q", code, term)
	}
	// The birth's standing files cannot be written.
	fresh = t.TempDir()
	writeZone(t, fresh, zone.WishSlot, "goal")
	code, term = runCode(t, fresh, map[string]bool{"write:contract.md": true})
	if code != 1 || !strings.Contains(term, "birth failed") {
		t.Fatalf("code %d term %q", code, term)
	}
	// The wish slot cannot die at the birth.
	fresh = t.TempDir()
	writeZone(t, fresh, zone.WishSlot, "goal")
	code, term = runCode(t, fresh, map[string]bool{"remove:wish.md": true})
	if code != 1 || !strings.Contains(term, "birth failed") {
		t.Fatalf("code %d term %q", code, term)
	}
}

func TestSubmissionAndRunFailuresAreNamed(t *testing.T) {
	root := newBornRoot(t)
	green := `behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies harder.
      scope:
        - src/**
      rows:
        - run: cat src/in.txt > out.txt
          expect:
            files:
                out.txt: "contains: ok"
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
	moving := `notes:
    - text: a note that moves the knowledge page
behaviors:
    - id: tool/copy
      name: The Copy
      prose: Copies anew.
      scope:
        - src/**
      rows:
        - run: cat src/in.txt > out.txt
          expect:
            files:
                out.txt: "contains: o"
`
	for _, tc := range []struct {
		name string
		fail map[string]bool
		want string
	}{
		{"the report view", map[string]bool{"write:.punchtape/report.md": true}, "report view:"},
		{"the state view", map[string]bool{"write:.punchtape/state.md": true}, "state view:"},
		{"the knowledge view", map[string]bool{"write:.punchtape/knowledge.md": true}, "knowledge view:"},
		{"the spec page", map[string]bool{"write:.punchtape/spec": true}, "spec view:"},
		{"the spec index", map[string]bool{"write:.punchtape/spec/index.md": true}, "spec view:"},
		{"the spec directory", map[string]bool{"mkdirall:.punchtape/spec": true}, "spec view:"},
		{"the spec walk", map[string]bool{"walk:.punchtape/spec": true}, "spec view:"},
		{"the corpus image", map[string]bool{"write:.punchtape/corpus.yaml": true}, "corpus view:"},
		{"the scratch clear", map[string]bool{"removeall:.punchtape/tmp": true}, "tmp clear:"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := newBornRoot(t)
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
	writeZone(t, root, zone.InboxSlot, "remove: [tool/copy]")
	code, term = runCode(t, root, map[string]bool{"remove:.punchtape/spec/": true})
	if code != 1 || !strings.Contains(term, "spec view:") {
		t.Fatalf("the retired page: code %d term %q", code, term)
	}
}

// TestRenderAndSpecWalkFaultsAreLoud — the render's second walk of
// the tree (measured after the execution) and the spec tree's own
// walk over an unreadable page directory fail loudly: a state page
// built on half a tree is a lie.
func TestRenderAndSpecWalkFaultsAreLoud(t *testing.T) {
	// The row itself locks a directory between the two walks: the
	// launch walk saw the whole tree, the render walk cannot.
	root := newBornRoot(t)
	writeZone(t, root, zone.InboxSlot, `behaviors:
    - id: tool/lock
      name: The lock
      prose: Locks a directory against the render walk.
      scope:
        - lock.txt
      rows:
        - run: chmod 000 src && echo locked > lock.txt
          expect:
            rc: 0
            files:
                lock.txt: "contains: locked"
`)
	t.Cleanup(func() { _ = os.Chmod(filepath.Join(root, "src"), 0o755) })
	mountPorts()
	var term strings.Builder
	if code := Run(root, &term); code != 1 || !strings.Contains(term.String(), "punchtape:") {
		t.Fatalf("the render walk must fail loudly: code %d term %q", code, term.String())
	}
	if err := os.Chmod(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	// The spec tree's own walk over an unreadable directory of a
	// retired behavior (no page writes stand in the way): the retire
	// sweep says so instead of passing in silence.
	root2 := newBornRoot(t)
	writeZone(t, root2, zone.InboxSlot, "remove: [tool/copy]")
	runIn(t, root2)
	specTool := filepath.Join(root2, zone.Dirname, zone.SpecDir, "tool")
	if err := os.Chmod(specTool, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(specTool, 0o755) })
	term.Reset()
	if code := Run(root2, &term); code != 1 || !strings.Contains(term.String(), "spec view:") {
		t.Fatalf("the spec walk must fail loudly: code %d term %q", code, term.String())
	}
}

func TestVerdictAppendFailureAndSecondVerdict(t *testing.T) {
	root := newBornRoot(t)
	// A red row: the verdict says not yet.
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
	writeZone(t, root, zone.InboxSlot, zoneFile(t, root, zone.CorpusFile))
	runIn(t, root)
	if !strings.Contains(zoneFile(t, root, zone.ReportFile), "verdict: not yet") {
		t.Fatalf("a red corpus must say not yet")
	}
	// The row heals; the verdict's journal append fails: the verdict
	// simply does not land, the run still answers.
	writeZone(t, root, zone.InboxSlot, `behaviors:
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
	runCode(t, root, map[string]bool{"append:epoch-": true})
	if strings.Contains(zoneFile(t, root, zone.StateFile), "verdict: #") {
		t.Fatalf("a failed append must not fake a verdict")
	}
	// Without the failure the verdict lands; a second full cycle
	// records the second verdict.
	writeZone(t, root, zone.InboxSlot, zoneFile(t, root, zone.CorpusFile))
	runIn(t, root)
	os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("ok ok"), 0o644)
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, zoneFile(t, root, zone.CorpusFile))
	runIn(t, root)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "verdict: #2") {
		t.Fatalf("the second point of truth must be recorded, state:\n%s", state)
	}
}

func TestBirthFaultsAreNamed(t *testing.T) {
	fresh := func() string {
		dir := t.TempDir()
		writeZone(t, dir, zone.WishSlot, "goal")
		return dir
	}
	// The journal's mkdir cannot happen at all.
	code, term := runCode(t, fresh(), map[string]bool{"mkdir:journal": true})
	if code != 1 || !strings.Contains(term, "birth failed") {
		t.Fatalf("mkdir: code %d term %q", code, term)
	}
	// The wish's object cannot be stored.
	code, term = runCode(t, fresh(), map[string]bool{"write:objects": true})
	if code != 1 || !strings.Contains(term, "birth failed") {
		t.Fatalf("put: code %d term %q", code, term)
	}
	// The stream cannot open.
	code, term = runCode(t, fresh(), map[string]bool{"readdir:journal": true})
	if code != 1 || !strings.Contains(term, "birth failed") {
		t.Fatalf("open stream: code %d term %q", code, term)
	}
	// The first record cannot land.
	code, term = runCode(t, fresh(), map[string]bool{"append:epoch-": true})
	if code != 1 || !strings.Contains(term, "birth failed") {
		t.Fatalf("append: code %d term %q", code, term)
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
                out.txt: "contains: moved"
`)
	os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("moved"), 0o644)
	code, term = runCode(t, root, map[string]bool{"mkdirall:journal": true})
	if code != 1 {
		t.Fatalf("rotate mkdirall: code %d term %q", code, term)
	}
}

func TestTreeFaultsDoNotBlindTheFingerprints(t *testing.T) {
	root := newBornRoot(t)
	// A directory of the tree cannot be entered during the walk: the
	// walk's own error reaches the caller, the run refuses.
	mountPorts()
	if err := os.Chmod(filepath.Join(root, "src"), 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(filepath.Join(root, "src"), 0o755) })
	var term strings.Builder
	if code := Run(root, &term); code != 1 {
		t.Fatalf("the unreadable tree must refuse the run: code %d term %q", code, term.String())
	}
	if err := os.Chmod(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	// The scope file cannot be statted: it spells absent, honestly.
	code, fault := runCode(t, root, map[string]bool{"stat:src/in.txt": true})
	if code != 0 {
		t.Fatalf("stat: code %d term %q", code, fault)
	}
	// The scope file cannot be read after a good stat: absent too.
	// A fresh mtime first — a cache hit would never read.
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
	writeZone(t, root, zone.InboxSlot, `behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies more.
      scope:
        - src/**
      rows:
        - run: cat src/in.txt > out.txt
          expect:
            files:
                out.txt: "contains: ok"
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
