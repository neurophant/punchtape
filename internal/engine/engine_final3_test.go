// The closing coverage battery of the engine: the branches the
// cycle and fault batteries did not reach — the exact rejection and
// failure lines of the submission pipeline, the writer discipline's
// recovery windows (the pending marker, the journal-ahead-of-files
// replay, the legacy manifest adoption), the renders' fault paths,
// the digest cache lifecycle, the snapshot walk's special inodes.
// Everything runs over the in-memory stack; the one exception is the
// linkFS wrapper (below), which presents symlink and non-regular
// inodes to the snapshot walk — modes the plain memory filesystem
// cannot produce.
package engine

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/delta"
	"github.com/neurophant/punchtape/internal/execport"
	"github.com/neurophant/punchtape/internal/fsport"
	"github.com/neurophant/punchtape/internal/journal"
	"github.com/neurophant/punchtape/internal/ledger"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// --- small helpers over the knobs ---

// knobRead — one path's reads fail for the duration of f.
func knobRead(stack *fakeStack, path string, f func()) {
	stack.fs.readErr = errIO
	stack.fs.readErrPath = path
	defer func() { stack.fs.readErr = nil; stack.fs.readErrPath = "" }()
	f()
}

func markerPathOf(eng *Engine) string {
	return filepath.Join(eng.Store.Root(), "cache", "pending-transaction")
}

func lastgoodPathOf(eng *Engine) string {
	return filepath.Join(eng.Store.Root(), "cache", "lastgood.yaml")
}

// --- the rejection error and its line helpers ---

func TestRejectErrErrorLine(t *testing.T) {
	err := rejection("delta", "the reason")
	if !isRejection(err) {
		t.Fatal("rejection not recognized")
	}
	if got := err.(*rejectErr).Error(); got != "delta: the reason" {
		t.Fatalf("error line: %q", got)
	}
	if isRejection(errors.New("plain")) {
		t.Fatal("plain error passed as rejection")
	}
}

func TestFirstErrorLineFirstOfMany(t *testing.T) {
	if got := firstErrorLine("one\ntwo\nthree"); got != "one" {
		t.Fatalf("multi-line reply: %q", got)
	}
}

func TestValidRelPathEmpty(t *testing.T) {
	err := validRelPath("")
	if err == nil || !strings.Contains(err.Error(), "path is empty") {
		t.Fatalf("empty path: %v", err)
	}
}

// --- the submission pipeline's failure branches ---

func TestSubmitStateReadFault(t *testing.T) {
	stack, eng := newInst(t)
	knobRead(stack, filepath.Join(eng.Store.Root(), "state.yaml"), func() {
		reply, code := submit(t, stack, eng, specDoc("B1", "hello"))
		if code != Failure || !strings.Contains(reply, "FAIL:") {
			t.Fatalf("state fault: %d %s", code, reply)
		}
	})
}

func TestDeliverFaultAndDriftBranches(t *testing.T) {
	stack, eng := newInst(t)
	submit(t, stack, eng, specDoc("B1", "hello"))
	submit(t, stack, eng, "kind: run\n")
	submit(t, stack, eng, "kind: run\nratify: true\n")
	// A corpus fault under a deliver submission: the reply is an
	// honest failure (the snapshot cannot be consulted for the reply
	// slot either).
	knobRead(stack, filepath.Join(eng.Store.Root(), "canon", "atoms", "b1.yaml"), func() {
		reply, code := submit(t, stack, eng, fr(specDoc("B2", "hello")))
		if code != Failure {
			t.Fatalf("deliver submission under corpus fault: %d %s", code, reply)
		}
	})
	// A drifted surface at deliver: the rendered verdict names the
	// drift honestly (the snapshot itself stays untouched).
	if err := stack.fs.WriteFile(filepath.Join(eng.Workdir, "post.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	eng.flushDigestCache()
	v, err := eng.RenderVerdict()
	if err != nil || v.Status != VerdictNotReady {
		t.Fatalf("drifted deliver verdict: %v %s", err, v.Status)
	}
	if !strings.Contains(v.Text(), "the tree moved") {
		t.Fatalf("drift line missing:\n%s", v.Text())
	}
}

func TestLedgerAppendFaults(t *testing.T) {
	// The apply entry's append fails after the journal rode.
	stack, eng := newInst(t)
	stack.fs.appendCalls = 0
	stack.fs.appendFailAt = 2
	reply, code := submit(t, stack, eng, specDoc("B1", "hello"))
	if code != Failure || !strings.Contains(reply, "ledger") {
		t.Fatalf("apply ledger fault: %d %s", code, reply)
	}
}

func TestBuilderLoadPaths(t *testing.T) {
	stack, eng := newInst(t)
	submit(t, stack, eng, specDoc("B1", "hello"))
	b := newBuilder(eng, mustState(eng))
	if err := b.load(); err != nil {
		t.Fatal(err)
	}
	if err := b.load(); err != nil {
		t.Fatalf("second load: %v", err)
	}
	// A failing corpus read is a system failure, both through load
	// and through applyDoc (which loads first).
	knobRead(stack, filepath.Join(eng.Store.Root(), "canon", "atoms", "b1.yaml"), func() {
		b2 := newBuilder(eng, mustState(eng))
		if err := b2.load(); err == nil {
			t.Fatal("load under fault succeeded")
		}
		b3 := newBuilder(eng, mustState(eng))
		err := b3.applyDoc(&delta.Doc{Kind: delta.KindRun})
		if err == nil || isRejection(err) {
			t.Fatalf("applyDoc under fault: %v", err)
		}
	})
}

func TestTouchDeduplicates(t *testing.T) {
	_, eng := newInst(t)
	b := newBuilder(eng, mustState(eng))
	b.putAtom(canon.Atom{ID: "B1", Name: "n"})
	b.putAtom(canon.Atom{ID: "B1", Name: "n"})
	if len(b.touched) != 1 {
		t.Fatalf("touched: %v", b.touched)
	}
}

func TestApplyDocFormRules(t *testing.T) {
	stack := bindUnits(t)
	eng, err := Init(stack.dir)
	if err != nil {
		t.Fatal(err)
	}
	// A ratifying run over an empty corpus is the empty-behaviors
	// refusal (the rule mirrors the spec's).
	b := newBuilder(eng, mustState(eng))
	err = b.applyDoc(&delta.Doc{Kind: delta.KindRun, Ratify: true})
	if !isRejection(err) || !strings.Contains(err.Error(), "behaviors") {
		t.Fatalf("ratify-empty refusal: %v", err)
	}
	// A run argv that cannot produce fields is a row-label rejection.
	b2 := newBuilder(eng, mustState(eng))
	err = b2.applyDoc(&delta.Doc{Kind: delta.KindSpec, Behaviors: []delta.Behavior{{
		ID: "B1", Name: "n", Prose: "p",
		Rows: []canon.Row{{Run: canon.Cmd{IsArgv: true}}},
	}}})
	if !isRejection(err) || !strings.Contains(err.Error(), "argv list is empty") {
		t.Fatalf("argv-fields refusal: %v", err)
	}
}

func TestSeedAndBytesPathRefusals(t *testing.T) {
	stack, eng := newInst(t)
	base := "kind: spec\nbehaviors:\n  - id: B1\n    name: n\n    prose: p\n    rows:\n"
	reply, code := submit(t, stack, eng, base+"      - run: \"true\"\n        seeds:\n          - path: s.txt\n            from-input: /abs/x.bin\n")
	mustReject(t, reply, code, "from-input", "must be relative")
	reply, code = submit(t, stack, eng, base+"      - run: \"true\"\n        bytes: file://../esc.bin\n")
	mustReject(t, reply, code, "bytes sample path")
}

func TestStatSeesDirectory(t *testing.T) {
	stack, eng := newInst(t)
	in := filepath.Join(stack.dir, "..", "input")
	stack.fs.MkdirAll(filepath.Join(in, "data"))
	stack.fs.MkdirAll(filepath.Join(eng.Store.Root(), "inbox", "sub"))
	doc := "kind: spec\nbehaviors:\n  - id: B1\n    name: n\n    prose: p\n    rows:\n      - run: \"true\"\n        seeds:\n          - path: s.txt\n            from-input: data\n  - id: B2\n    name: n\n    prose: p\n    rows:\n      - run: \"true\"\n        bytes: file://sub\n"
	reply, code := submit(t, stack, eng, doc)
	mustReject(t, reply, code, "no such file")
}

// --- agentsref edge shapes ---

func TestAgentsRefStaleRewrites(t *testing.T) {
	stack, eng := newInst(t)
	path := filepath.Join(eng.Workdir, "AGENTS.md")
	// A machine file without a tail and stale: rewritten whole.
	if err := stack.fs.WriteFile(path, []byte(refMarker+"\nstale body\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ensureAgentsRef(eng.Workdir)
	data, _ := stack.fs.ReadFile(path)
	if string(data) != agentsRefText() {
		t.Fatalf("stale headless ref: %q", string(data)[:40])
	}
	// A machine file with a tail and a stale head: the head is
	// regenerated, the tail preserved verbatim.
	if err := stack.fs.WriteFile(path, []byte(refMarker+"\nold head\n"+tailBoundary+"\nrepo rules\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ensureAgentsRef(eng.Workdir)
	data, _ = stack.fs.ReadFile(path)
	if string(data) != machineHead()+"repo rules\n" {
		t.Fatalf("stale combined ref: %q", string(data)[:60])
	}
}

// --- the snapshot walk's shapes ---

func TestSnapshotTreeStateFault(t *testing.T) {
	stack, eng := newInst(t)
	knobRead(stack, filepath.Join(eng.Store.Root(), "state.yaml"), func() {
		if _, err := eng.snapshotTree(); err == nil {
			t.Fatal("snapshot under state fault succeeded")
		}
	})
}

func TestSnapshotExcludedAndPlainDirs(t *testing.T) {
	stack, eng := newInst(t)
	submit(t, stack, eng, "kind: spec\nbuild-ignore:\n  - skipdir/\n  - \"*.log\"\nbehaviors:\n  - id: B1\n    name: n\n    prose: p\n    rows:\n      - run: \"true\"\n")
	stack.fs.MkdirAll(filepath.Join(eng.Workdir, "skipdir"))
	if err := stack.fs.WriteFile(filepath.Join(eng.Workdir, "skipdir", "x.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := stack.fs.WriteFile(filepath.Join(eng.Workdir, "plain.log"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	stack.fs.MkdirAll(filepath.Join(eng.Workdir, "sub"))
	if err := stack.fs.WriteFile(filepath.Join(eng.Workdir, "sub", "keep.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	ckp, err := eng.snapshotTree()
	if err != nil {
		t.Fatal(err)
	}
	saw, sawSkip, sawLog := false, false, false
	for _, f := range ckp.Files {
		if f.Path == "sub/keep.txt" {
			saw = true
		}
		if f.Path == "skipdir/x.txt" {
			sawSkip = true
		}
		if f.Path == "plain.log" {
			sawLog = true
		}
	}
	if !saw || sawSkip || sawLog {
		t.Fatalf("surface: keep=%v skip=%v log=%v", saw, sawSkip, sawLog)
	}
}

// linkFS — the memory filesystem with a overlay of special inodes:
// symlinks (by target) and one non-regular file. The plain memFS
// cannot produce these modes; the checkpoint walk must still account
// for them exactly as a real host presents them.
type linkFS struct {
	inner    *memFS
	links    map[string]string
	badLink  string
	specials map[string]os.FileMode
}

type modeInfo struct {
	name string
	mode os.FileMode
	mod  time.Time
}

func (m modeInfo) Name() string       { return m.name }
func (m modeInfo) Size() int64        { return 0 }
func (m modeInfo) Mode() os.FileMode  { return m.mode }
func (m modeInfo) ModTime() time.Time { return m.mod }
func (m modeInfo) IsDir() bool        { return m.mode.IsDir() }
func (m modeInfo) Sys() any           { return nil }

func (l *linkFS) info(p string, orig os.FileInfo) os.FileInfo {
	if _, ok := l.links[p]; ok || l.badLink == p {
		return modeInfo{name: filepath.Base(p), mode: os.ModeSymlink | 0o777, mod: orig.ModTime()}
	}
	if mode, ok := l.specials[p]; ok {
		return modeInfo{name: filepath.Base(p), mode: mode, mod: orig.ModTime()}
	}
	return orig
}

func (l *linkFS) Stat(path string) (os.FileInfo, error) {
	i, err := l.inner.Stat(path)
	if err != nil {
		return nil, err
	}
	return l.info(path, i), nil
}

func (l *linkFS) Lstat(path string) (os.FileInfo, error) { return l.Stat(path) }

func (l *linkFS) Readlink(path string) (string, error) {
	if l.badLink == path {
		return "", notExist(path)
	}
	if tgt, ok := l.links[path]; ok {
		return tgt, nil
	}
	return "", notExist(path)
}

func (l *linkFS) Walk(root string, fn filepath.WalkFunc) error {
	return l.inner.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err == nil {
			info = l.info(p, info)
		}
		return fn(p, info, err)
	})
}

func (l *linkFS) MkdirAll(dir string) error            { return l.inner.MkdirAll(dir) }
func (l *linkFS) ReadFile(path string) ([]byte, error) { return l.inner.ReadFile(path) }
func (l *linkFS) ReadDir(dir string) ([]os.DirEntry, error) {
	return l.inner.ReadDir(dir)
}
func (l *linkFS) WriteAtomic(path string, data []byte) error {
	return l.inner.WriteAtomic(path, data)
}
func (l *linkFS) AppendFile(path string, data []byte) error {
	return l.inner.AppendFile(path, data)
}
func (l *linkFS) Remove(path string) error { return l.inner.Remove(path) }
func (l *linkFS) MkdirTemp(dir, pattern string) (string, error) {
	return l.inner.MkdirTemp(dir, pattern)
}
func (l *linkFS) WriteFile(path string, data []byte, perm os.FileMode) error {
	return l.inner.WriteFile(path, data, perm)
}
func (l *linkFS) Chmod(path string, perm os.FileMode) error { return l.inner.Chmod(path, perm) }
func (l *linkFS) RemoveAll(path string) error               { return l.inner.RemoveAll(path) }
func (l *linkFS) LockExclusive(path string) (func(), error) {
	return l.inner.LockExclusive(path)
}
func (l *linkFS) TryLockExclusive(path string) (func(), bool, error) {
	return l.inner.TryLockExclusive(path)
}
func (l *linkFS) Symlink(target, link string) error { return l.inner.Symlink(target, link) }

func TestSnapshotTreeSymlinksAndSpecials(t *testing.T) {
	stack, eng := newInst(t)
	if err := stack.fs.WriteFile(filepath.Join(eng.Workdir, "real.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := stack.fs.WriteFile(filepath.Join(eng.Workdir, "good.lnk"), []byte("target"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := stack.fs.WriteFile(filepath.Join(eng.Workdir, "bad.lnk"), []byte("target"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := stack.fs.WriteFile(filepath.Join(eng.Workdir, "fifo"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	l := &linkFS{
		inner: stack.fs,
		links: map[string]string{filepath.Join(eng.Workdir, "good.lnk"): "real.txt"},
		specials: map[string]os.FileMode{
			filepath.Join(eng.Workdir, "fifo"): os.ModeNamedPipe | 0o600,
		},
	}
	old := fsport.Bind
	fsport.Bind = l
	t.Cleanup(func() { fsport.Bind = old })
	ckp, err := eng.snapshotTree()
	if err != nil {
		t.Fatal(err)
	}
	var linkRec, fifoRec bool
	for _, f := range ckp.Files {
		if f.Path == "good.lnk" {
			linkRec = true
			if f.SHA256 != "link:real.txt" {
				t.Fatalf("link record: %q", f.SHA256)
			}
		}
		if f.Path == "fifo" {
			fifoRec = true
		}
	}
	if !linkRec || fifoRec {
		t.Fatalf("special inodes: link=%v fifo=%v", linkRec, fifoRec)
	}
	// A link whose target cannot be read is a walk failure.
	l.badLink = filepath.Join(eng.Workdir, "bad.lnk")
	if _, err := eng.snapshotTree(); err == nil {
		t.Fatal("unreadable link swallowed")
	}
	l.badLink = ""
}

// --- checkpoint record faults ---

func TestCheckpointEffectNilCounters(t *testing.T) {
	_, eng := newInst(t)
	b := newBuilder(eng, canon.State{Stage: canon.StageIntake})
	b.checkpointEffect(&Checkpoint{Seq: 3}, "tx9")
	if b.state.Counters["CKP"] != 3 {
		t.Fatalf("counter: %+v", b.state.Counters)
	}
	if len(b.effects) != 1 || b.effects[0].Path != checkpointFileName(3) {
		t.Fatalf("effect: %+v", b.effects)
	}
}

func TestLastCheckpointFaults(t *testing.T) {
	stack, eng := newInst(t)
	submit(t, stack, eng, specDoc("B1", "hello"))
	submit(t, stack, eng, "kind: run\n")
	submit(t, stack, eng, "kind: run\n")
	newest := filepath.Join(eng.Workdir, filepath.FromSlash(checkpointFileName(2)))
	data, err := stack.fs.ReadFile(newest)
	if err != nil {
		t.Fatal(err)
	}
	knobRead(stack, filepath.Join(eng.Store.Root(), "state.yaml"), func() {
		if ckp := eng.lastCheckpoint(); ckp != nil {
			t.Fatalf("checkpoint under state fault: %+v", ckp)
		}
	})
	if err := stack.fs.Remove(newest); err != nil {
		t.Fatal(err)
	}
	if ckp := eng.lastCheckpoint(); ckp == nil || ckp.Seq != 1 {
		t.Fatalf("fallback to older record: %+v", ckp)
	}
	if err := stack.fs.WriteFile(newest, []byte("not: a checkpoint"), 0o644); err != nil {
		t.Fatal(err)
	}
	if ckp := eng.lastCheckpoint(); ckp == nil || ckp.Seq != 1 {
		t.Fatalf("fallback past corrupt record: %+v", ckp)
	}
	_ = data
}

// --- the digest cache lifecycle ---

func TestDigestOfFaults(t *testing.T) {
	stack, eng := newInst(t)
	if d, ok := eng.digestOf(".punchtape/gone.yaml"); ok || d != "" {
		t.Fatalf("missing file digest: %q %v", d, ok)
	}
	// A file that stats but does not read.
	if err := stack.fs.WriteFile(filepath.Join(eng.Store.Root(), "state.yaml"), []byte("schema-version: 8\nstage: implement\ncounters:\n  CKP: 1\nwish: longer now\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	knobRead(stack, filepath.Join(eng.Store.Root(), "state.yaml"), func() {
		if d, ok := eng.digestOf(".punchtape/state.yaml"); ok || d != "" {
			t.Fatalf("unreadable file digest: %q %v", d, ok)
		}
	})
}

func TestDigestCacheEvictionAndWriteFault(t *testing.T) {
	_, eng := newInst(t)
	c := eng.theDigestCache()
	for i := 0; i < canondata.Limit("digest-cache-max")+1; i++ {
		c.entries["f"+strconv.Itoa(i)] = digestEntry{MTimeNs: 1, Size: 1, Digest: "d"}
	}
	c.dirty = true
	eng.flushDigestCache()
	if len(c.entries) != 0 {
		t.Fatalf("over-budget cache not rebuilt: %d", len(c.entries))
	}
	c.entries["x"] = digestEntry{MTimeNs: 1, Size: 1, Digest: "d"}
	c.dirty = true
	stack := lastStack
	stack.fs.writeErrPath = filepath.Join(eng.Store.Root(), "cache", "file-digests.yaml")
	eng.flushDigestCache()
	stack.fs.writeErrPath = ""
	if !c.dirty {
		t.Fatal("failed flush cleared the dirty flag")
	}
}

func TestTheDigestCacheLoadAndCorruption(t *testing.T) {
	stack, eng := newInst(t)
	if _, ok := eng.digestOf(".punchtape/state.yaml"); !ok {
		t.Fatal("digest miss")
	}
	eng.flushDigestCache()
	eng.digestCache = nil
	if c := eng.theDigestCache(); len(c.entries) == 0 {
		t.Fatal("flushed cache did not reload")
	}
	cacheFile := filepath.Join(eng.Store.Root(), "cache", "file-digests.yaml")
	if err := stack.fs.WriteFile(cacheFile, []byte("files: [unclosed"), 0o644); err != nil {
		t.Fatal(err)
	}
	eng.digestCache = nil
	if c := eng.theDigestCache(); len(c.entries) != 0 {
		t.Fatalf("corrupt cache loaded: %d entries", len(c.entries))
	}
}

// --- Init / Begin / Open faults ---

func TestInitWriterFault(t *testing.T) {
	stack := bindUnits(t)
	stack.fs.lockErr = errIO
	if _, err := Init(stack.dir); err == nil {
		t.Fatal("init under writer lock fault succeeded")
	}
}

func TestBeginFaultPathsExtended(t *testing.T) {
	stack := bindUnits(t)
	eng, err := Init(stack.dir)
	if err != nil {
		t.Fatal(err)
	}
	knobRead(stack, filepath.Join(eng.Store.Root(), "state.yaml"), func() {
		if _, code := eng.Begin("a brief"); code != Failure {
			t.Fatal("state fault begin accepted")
		}
	})
	stack.fs.appendCalls = 0
	stack.fs.appendFailAt = 2
	if _, code := eng.Begin("a brief"); code != Failure {
		t.Fatal("ledger append fault begin accepted")
	}
	stack.fs.appendFailAt = 0
	// A render failure after a committed birth: the state is brought
	// to implement by hand (a stage with a corpus read in its slot).
	if err := stack.fs.WriteFile(filepath.Join(eng.Store.Root(), "state.yaml"), []byte("schema-version: 9\nstage: implement\ncounters:\n  CKP: 0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := stack.fs.WriteFile(filepath.Join(eng.Store.Root(), "canon", "atoms", "b1.yaml"), []byte("id: B1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	knobRead(stack, filepath.Join(eng.Store.Root(), "canon", "atoms", "b1.yaml"), func() {
		if _, code := eng.Begin("third brief"); code != Failure {
			t.Fatal("render fault begin accepted")
		}
	})
}

func TestOpenJournalLedgerFaults(t *testing.T) {
	stack, eng := newInst(t)
	submit(t, stack, eng, specDoc("B1", "hello"))
	knobRead(stack, filepath.Join(eng.Store.Root(), "journal.yamll"), func() {
		if _, _, err := Open(stack.dir); err == nil {
			t.Fatal("journal fault open succeeded")
		}
	})
	knobRead(stack, filepath.Join(eng.Store.Root(), "ledger.yamll"), func() {
		if _, _, err := Open(stack.dir); err == nil {
			t.Fatal("ledger fault open succeeded")
		}
	})
}

// --- the writer discipline ---

func TestAcquireWriterFaults(t *testing.T) {
	stack, eng := newInst(t)
	stack.fs.mkdirErr = errIO
	if _, err := eng.acquireWriter(); err == nil || !strings.Contains(err.Error(), "writer lock") {
		t.Fatalf("mkdir fault: %v", err)
	}
	stack.fs.mkdirErr = nil
	lockPath := filepath.Join(eng.Store.Root(), "cache", "writer.lock")
	release, held, err := stack.fs.TryLockExclusive(lockPath)
	if err != nil || !held {
		t.Fatalf("lock probe: %v %v", held, err)
	}
	defer release()
	if _, err := eng.acquireWriter(); !errors.Is(err, ErrWriterBusy) {
		t.Fatalf("held lock: %v", err)
	}
	// Persistent contention stays an honest refusal after the
	// retries.
	reply, code := submit(t, stack, eng, specDoc("B1", "hello"))
	if code != Failure || !strings.Contains(reply, "another writer holds") {
		t.Fatalf("busy reply: %d %s", code, reply)
	}
}

func TestCommitStepFaults(t *testing.T) {
	// The marker write fails: the birth fails.
	stack := bindUnits(t)
	eng, err := Init(stack.dir)
	if err != nil {
		t.Fatal(err)
	}
	stack.fs.writeErrPath = markerPathOf(eng)
	if _, code := eng.Begin("brief one"); code != Failure {
		t.Fatal("marker write fault begin accepted")
	}
	stack.fs.writeErrPath = ""

	// The effects write fails.
	stack2 := bindUnits(t)
	eng2, err := Init(stack2.dir)
	if err != nil {
		t.Fatal(err)
	}
	stack2.fs.writeErrPath = filepath.Join(stack2.dir, ".punchtape", "state.yaml")
	if _, code := eng2.Begin("brief two"); code != Failure {
		t.Fatal("effect write fault begin accepted")
	}
	stack2.fs.writeErrPath = ""

	// The lastgood write fails: the marker stays, the next open
	// completes the transaction by replay.
	stack3 := bindUnits(t)
	eng3, err := Init(stack3.dir)
	if err != nil {
		t.Fatal(err)
	}
	stack3.fs.writeErrPath = lastgoodPathOf(eng3)
	if _, code := eng3.Begin("brief three"); code != Failure {
		t.Fatal("lastgood write fault begin accepted")
	}
	stack3.fs.writeErrPath = ""
	if _, err := stack3.fs.ReadFile(markerPathOf(eng3)); err != nil {
		t.Fatalf("marker missing after torn commit: %v", err)
	}
	eng4, exists, err := Open(stack3.dir)
	if err != nil || !exists {
		t.Fatalf("recovery open: %v %v", exists, err)
	}
	if _, err := stack3.fs.ReadFile(markerPathOf(eng4)); err == nil {
		t.Fatal("marker survived the recovery")
	}
}

func TestApplyEffectsDirectShapes(t *testing.T) {
	stack := bindUnits(t)
	if err := applyEffects(stack.dir, []journal.Effect{{Path: "never.txt", Delete: true}}); err != nil {
		t.Fatalf("missing delete: %v", err)
	}
	if err := stack.fs.WriteFile(filepath.Join(stack.dir, "kill.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	stack.fs.removeErr = errIO
	if err := applyEffects(stack.dir, []journal.Effect{{Path: "kill.txt", Delete: true}}); err == nil {
		t.Fatal("failing delete swallowed")
	}
	stack.fs.removeErr = nil
	if err := applyEffects(stack.dir, []journal.Effect{{Path: "kill.txt", Delete: true}}); err != nil {
		t.Fatal(err)
	}
	if _, err := stack.fs.Stat(filepath.Join(stack.dir, "kill.txt")); err == nil {
		t.Fatal("deleted file remains")
	}
	if err := stack.fs.WriteFile(filepath.Join(stack.dir, "probe.txt"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	knobRead(stack, filepath.Join(stack.dir, "probe.txt"), func() {
		err := applyEffects(stack.dir, []journal.Effect{{Path: "probe.txt", Content: "new"}})
		if err == nil || !strings.Contains(err.Error(), "read-back") {
			t.Fatalf("read-back: %v", err)
		}
	})
}

func TestFoldInstanceFilesShapes(t *testing.T) {
	files := foldInstanceFiles([]journal.Entry{{Effects: []journal.Effect{
		{Path: "plain.txt", Content: "x"},
		{Path: ".punchtape/canon/atoms/b1.yaml", Content: "a"},
		{Path: ".punchtape/canon/atoms/b1.yaml", Delete: true},
	}}})
	if len(files) != 0 {
		t.Fatalf("fold: %+v", files)
	}
	files = foldInstanceFiles([]journal.Entry{{Effects: []journal.Effect{
		{Path: ".punchtape/state.yaml", Content: "live"},
	}}})
	if files[stateFilePath] != yamlio.Digest([]byte("live")) {
		t.Fatalf("state fold: %+v", files)
	}
}

func TestVerifyInstanceReadFaults(t *testing.T) {
	stack, eng := newInst(t)
	submit(t, stack, eng, specDoc("B1", "hello"))
	knobRead(stack, markerPathOf(eng), func() {
		if err := eng.verifyInstance(); err == nil || !strings.Contains(err.Error(), "transaction marker") {
			t.Fatalf("marker read fault: %v", err)
		}
	})
	knobRead(stack, lastgoodPathOf(eng), func() {
		if err := eng.verifyInstance(); err == nil || !strings.Contains(err.Error(), "lastgood") {
			t.Fatalf("lastgood read fault: %v", err)
		}
	})
}

func TestOpenCleanAndVirginInstances(t *testing.T) {
	// A virgin instance: an empty journal and no manifest — nothing
	// to reconcile.
	stack := bindUnits(t)
	if _, err := Init(stack.dir); err != nil {
		t.Fatal(err)
	}
	if _, exists, err := Open(stack.dir); err != nil || !exists {
		t.Fatalf("virgin open: %v %v", exists, err)
	}
	// A healthy committed instance: the manifest head matches the
	// journal head.
	stack2, eng2 := newInst(t)
	submit(t, stack2, eng2, specDoc("B1", "hello"))
	if _, exists, err := Open(stack2.dir); err != nil || !exists {
		t.Fatalf("clean open: %v %v", exists, err)
	}
}

func TestOpenLegacyAdoption(t *testing.T) {
	stack, eng := newInst(t)
	submit(t, stack, eng, specDoc("B1", "hello"))
	b1 := filepath.Join(eng.Store.Root(), "canon", "atoms", "b1.yaml")
	orig, err := stack.fs.ReadFile(b1)
	if err != nil {
		t.Fatal(err)
	}
	// No manifest: the expectation is derived from the journal and
	// the manifest is fixed.
	drop := func() { stack.fs.Remove(lastgoodPathOf(eng)) }
	drop()
	if _, _, err := Open(stack.dir); err != nil {
		t.Fatalf("legacy adoption: %v", err)
	}
	if _, err := stack.fs.ReadFile(lastgoodPathOf(eng)); err != nil {
		t.Fatalf("manifest not fixed: %v", err)
	}
	// A journaled file missing: an honest refusal.
	drop()
	stack.fs.Remove(b1)
	if _, _, err := Open(stack.dir); err == nil || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("legacy missing: %v", err)
	}
	// A journaled file differing.
	stack.fs.WriteFile(b1, orig, 0o644)
	drop()
	stack.fs.WriteFile(b1, append(append([]byte{}, orig...), ' '), 0o644)
	if _, _, err := Open(stack.dir); err == nil || !strings.Contains(err.Error(), "differs") {
		t.Fatalf("legacy differs: %v", err)
	}
	// A file without a machine record.
	stack.fs.WriteFile(b1, orig, 0o644)
	drop()
	stack.fs.WriteFile(filepath.Join(eng.Store.Root(), "canon", "atoms", "zz9.yaml"), []byte("id: Z\n"), 0o644)
	if _, _, err := Open(stack.dir); err == nil || !strings.Contains(err.Error(), "without a machine-written record") {
		t.Fatalf("legacy unjournaled: %v", err)
	}
	stack.fs.Remove(filepath.Join(eng.Store.Root(), "canon", "atoms", "zz9.yaml"))
}

func TestOpenLastgoodCorruptAndBogusHead(t *testing.T) {
	stack, eng := newInst(t)
	submit(t, stack, eng, specDoc("B1", "hello"))
	if err := stack.fs.WriteFile(lastgoodPathOf(eng), []byte("journal-head: [unclosed"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Open(stack.dir); err == nil || !strings.Contains(err.Error(), "lastgood") {
		t.Fatalf("corrupt manifest: %v", err)
	}
	// A manifest head the journal does not have.
	eng.writeLastgood(eng.Journal.All()[len(eng.Journal.All())-1].Transaction)
	data, _ := stack.fs.ReadFile(lastgoodPathOf(eng))
	lines := strings.SplitN(string(data), "\n", 2)
	lines[0] = "journal-head: deadbeefcafe"
	if err := stack.fs.WriteFile(lastgoodPathOf(eng), []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Open(stack.dir); err == nil {
		t.Fatal("bogus manifest head accepted")
	}
}

func TestMarkerRecoveryPaths(t *testing.T) {
	stack, eng := newInst(t)
	submit(t, stack, eng, specDoc("B1", "hello"))
	specTx := eng.Journal.All()[len(eng.Journal.All())-1].Transaction
	marker := markerPathOf(eng)

	// The marker names the head: the effects are idempotent, the
	// earlier entries are passed over.
	if err := stack.fs.WriteFile(marker, []byte(specTx), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Open(stack.dir); err != nil {
		t.Fatalf("marker recovery: %v", err)
	}
	if _, err := stack.fs.ReadFile(marker); err == nil {
		t.Fatal("marker survived recovery")
	}

	// The marker names a transaction whose effects are forbidden: the
	// recovery is an honest refusal.
	evil := journal.Entry{
		Transaction: "eviltx000001", DeltaKind: "run",
		RecordedAt: time.Now().UTC(),
		Effects:    []journal.Effect{{Path: ".punchtape/cache/evil", Content: "x"}},
	}
	if err := eng.Journal.Append(evil); err != nil {
		t.Fatal(err)
	}
	if err := stack.fs.WriteFile(marker, []byte("eviltx000001"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Open(stack.dir); err == nil || !strings.Contains(err.Error(), "recovery of transaction") {
		t.Fatalf("evil recovery: %v", err)
	}

	// The manifest write fails during the recovery.
	if err := stack.fs.WriteFile(marker, []byte(specTx), 0o644); err != nil {
		t.Fatal(err)
	}
	stack.fs.writeErrPath = lastgoodPathOf(eng)
	if _, _, err := Open(stack.dir); err == nil {
		t.Fatal("lastgood fault recovery succeeded")
	}
	stack.fs.writeErrPath = ""

	// The marker's removal fails.
	stack.fs.removeErrPath = marker
	if _, _, err := Open(stack.dir); err == nil {
		t.Fatal("marker removal fault recovery succeeded")
	}
	stack.fs.removeErrPath = ""

	// The marker names a transaction the journal does not have.
	if err := stack.fs.WriteFile(marker, []byte("ghost-tx"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Open(stack.dir); err == nil || !strings.Contains(err.Error(), "integrity") {
		t.Fatalf("ghost marker: %v", err)
	}
	stack.fs.Remove(marker)
}

func TestJournalReplayAfterManifest(t *testing.T) {
	stack, eng := newInst(t)
	submit(t, stack, eng, specDoc("B1", "hello"))
	specTx := eng.Journal.All()[len(eng.Journal.All())-1].Transaction

	// The journal moved ahead of an intact manifest: the tail is
	// replayed (a no-effect tail) and the manifest is rewritten.
	if err := eng.Journal.Append(journal.Entry{
		Transaction: "tailtx0000001", DeltaKind: "run",
		RecordedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Open(stack.dir); err != nil {
		t.Fatalf("replay: %v", err)
	}
	data, _ := stack.fs.ReadFile(lastgoodPathOf(eng))
	if !strings.HasPrefix(string(data), "journal-head: tailtx0000001") {
		t.Fatalf("manifest head after replay: %q", strings.SplitN(string(data), "\n", 2)[0])
	}

	// A tail with a forbidden effect: the replay is an honest refusal.
	if err := eng.Journal.Append(journal.Entry{
		Transaction: "evil-tx-000002", DeltaKind: "run",
		RecordedAt: time.Now().UTC(),
		Effects:    []journal.Effect{{Path: ".punchtape/cache/evil", Content: "x"}},
	}); err != nil {
		t.Fatal(err)
	}
	if err := eng.writeLastgood(specTx); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Open(stack.dir); err == nil || !strings.Contains(err.Error(), "replay of transaction") {
		t.Fatalf("evil replay: %v", err)
	}
}

func TestWriteLastgoodStateFallback(t *testing.T) {
	stack := bindUnits(t)
	eng, err := Init(stack.dir)
	if err != nil {
		t.Fatal(err)
	}
	// A journal without state effects: the observed state becomes
	// the base of the manifest.
	if err := eng.writeLastgood("headtx00000001"); err != nil {
		t.Fatal(err)
	}
	data, err := stack.fs.ReadFile(lastgoodPathOf(eng))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), stateFilePath) {
		t.Fatalf("state not folded from observation: %q", string(data))
	}
}

func TestScanInstanceFilesReadDirFault(t *testing.T) {
	stack, eng := newInst(t)
	submit(t, stack, eng, specDoc("B1", "hello"))
	stack.fs.readdirErr = errIO
	defer func() { stack.fs.readdirErr = nil }()
	out := eng.scanInstanceFiles()
	if len(out) != 1 || out[0] != stateFilePath {
		t.Fatalf("scan under readdir fault: %v", out)
	}
}

func TestDiffManifestSortedTwo(t *testing.T) {
	stack, eng := newInst(t)
	submit(t, stack, eng, "kind: spec\nbehaviors:\n  - id: B1\n    name: n\n    prose: p\n    rows:\n      - run: \"true\"\n  - id: B2\n    name: n\n    prose: p\n    rows:\n      - run: \"true\"\n")
	atomsDir := filepath.Join(eng.Store.Root(), "canon", "atoms")
	if err := stack.fs.Remove(filepath.Join(atomsDir, "b1.yaml")); err != nil {
		t.Fatal(err)
	}
	if err := stack.fs.WriteFile(filepath.Join(atomsDir, "b2.yaml"), []byte("id: B2\nhacked\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, _, err := Open(stack.dir)
	if err == nil || !strings.Contains(err.Error(), "missing") || !strings.Contains(err.Error(), "differs") {
		t.Fatalf("two-diff refusal: %v", err)
	}
}

// --- the door and the reference library ---

func TestEnsureCurrentSetMkdirFault(t *testing.T) {
	stack, eng := newInst(t)
	doorFile := filepath.Join(eng.Workdir, filepath.FromSlash(inboxCurrent))
	if err := stack.fs.Remove(doorFile); err != nil {
		t.Fatal(err)
	}
	stack.fs.mkdirErr = errIO
	defer func() { stack.fs.mkdirErr = nil }()
	eng.ensureCurrentSet()
	if _, err := stack.fs.ReadFile(doorFile); err == nil {
		t.Fatal("door created under mkdir fault")
	}
}

func TestEnsureReferenceMkdirFault(t *testing.T) {
	stack, eng := newInst(t)
	doc := filepath.Join(eng.Store.Root(), "reference", "model.md")
	if err := stack.fs.Remove(doc); err != nil {
		t.Fatal(err)
	}
	stack.fs.mkdirErr = errIO
	defer func() { stack.fs.mkdirErr = nil }()
	ensureReference(eng.Workdir)
	if _, err := stack.fs.ReadFile(doc); err == nil {
		t.Fatal("reference written under mkdir fault")
	}
}

// --- the renders' fault paths ---

func TestEnsureSpecRenderAtomsFault(t *testing.T) {
	stack, eng := newInst(t)
	submit(t, stack, eng, specDoc("B1", "hello"))
	specFile := filepath.Join(eng.Store.Root(), "spec.md")
	before, _ := stack.fs.ReadFile(specFile)
	knobRead(stack, filepath.Join(eng.Store.Root(), "canon", "atoms", "b1.yaml"), func() {
		eng.ensureSpecRender()
	})
	after, _ := stack.fs.ReadFile(specFile)
	if string(before) != string(after) {
		t.Fatal("spec render changed under fault")
	}
}

func TestSummaryRenderFaultsAndShapes(t *testing.T) {
	stack, eng := newInst(t)
	submit(t, stack, eng, specDoc("B1", "hello"))
	root := eng.Store.Root()
	knobRead(stack, filepath.Join(root, "state.yaml"), func() {
		if _, err := eng.renderSummaryText(); err == nil {
			t.Fatal("summary under state fault succeeded")
		}
		writeSummaryRender(eng)
		if _, err := stack.fs.ReadFile(filepath.Join(root, "SUMMARY.md")); err == nil {
			t.Fatal("summary written under fault")
		}
	})
	knobRead(stack, filepath.Join(root, "canon", "atoms", "b1.yaml"), func() {
		if _, err := eng.renderSummaryText(); err == nil {
			t.Fatal("summary under corpus fault succeeded")
		}
	})
	// Red rows are counted in the checked block.
	submit(t, stack, eng, fr(specDoc("B1", "wrong")))
	submit(t, stack, eng, "kind: run\n")
	text, err := eng.renderSummaryText()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "1 red") {
		t.Fatalf("red count: %q", text)
	}
	// An instance born without a phrase: the honest none line.
	stack2 := bindUnits(t)
	eng2, err := Init(stack2.dir)
	if err != nil {
		t.Fatal(err)
	}
	text, err = eng2.renderSummaryText()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "(none") {
		t.Fatalf("wish-none line: %q", text)
	}
}

// --- the next verb's fault paths ---

func TestNextVerbStateFault(t *testing.T) {
	stack, eng := newInst(t)
	if _, err := eng.Next(); err != nil {
		t.Fatal(err)
	}
	stateFile := filepath.Join(eng.Store.Root(), "state.yaml")
	if err := stack.fs.WriteFile(stateFile, append([]byte("schema-version: 8\nstage: intake\ncounters:\n  CKP: 0\n"), '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	knobRead(stack, stateFile, func() {
		if _, err := eng.Next(); err == nil {
			t.Fatal("next under state fault succeeded")
		}
	})
}

func TestNextVerbRenderFault(t *testing.T) {
	stack, eng := newInst(t)
	submit(t, stack, eng, specDoc("B1", "hello"))
	b1 := filepath.Join(eng.Store.Root(), "canon", "atoms", "b1.yaml")
	if _, err := eng.Next(); err != nil {
		t.Fatal(err)
	}
	if err := stack.fs.WriteFile(b1, []byte("id: B1\nname: n\nprose: changed bytes\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	knobRead(stack, b1, func() {
		if _, err := eng.Next(); err == nil {
			t.Fatal("next under corpus fault succeeded")
		}
	})
}

func TestRenderNextDeliverRenderFault(t *testing.T) {
	stack, eng := newInst(t)
	submit(t, stack, eng, specDoc("B1", "hello"))
	submit(t, stack, eng, "kind: run\n")
	submit(t, stack, eng, "kind: run\nratify: true\n")
	// The deliver slot reads the frozen snapshot; a MISSING one (the
	// legacy shape) falls back to the fresh render, which reads the
	// corpus — the knob bites there.
	delete(stack.fs.files, filepath.Join(eng.Store.Root(), "verdicts", "v1.txt"))
	knobRead(stack, filepath.Join(eng.Store.Root(), "canon", "atoms", "b1.yaml"), func() {
		if _, _, err := eng.renderNext(canon.State{Stage: canon.StageDeliver}); err == nil {
			t.Fatal("deliver render under corpus fault succeeded")
		}
	})
}

func TestPaintAtomCardRowBudget(t *testing.T) {
	stack, eng := newInst(t)
	doc := "kind: spec\nbehaviors:\n  - id: B1\n    name: n\n    prose: p\n    rows:\n"
	for i := 0; i < canondata.Limit("slot.list-budget")+3; i++ {
		doc += "      - run: \"false\"\n"
	}
	reply, code := submit(t, stack, eng, doc)
	mustAccept(t, reply, code)
	// every row reddens; the card's red list is budget-capped (the
	// door law: the card carries the reds in full, the cap holds);
	// the card rides next (the ack is thin)
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "suite=red")
	next, err := eng.Next()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(next, "(+3 more)") {
		t.Fatalf("row budget cap:\n%s", next)
	}
}

func TestExpectationLineForms(t *testing.T) {
	stack, eng := newInst(t)
	if err := stack.fs.WriteFile(filepath.Join(eng.Store.Root(), "inbox", "sample.bin"), []byte("hello\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	doc := "kind: spec\nbehaviors:\n  - id: B1\n    name: n\n    prose: p\n    rows:\n      - run: [\"false\"]\n        err: boom\n        form: \"corpus: the refusal surface\"\n        bytes: file://sample.bin\n        state:\n          - path: out.txt\n            text: hi\n      - run: [\"cat\", \"x\"]\n        state:\n          - path: y.txt\n            from-run: out.txt\n"
	reply, code := submit(t, stack, eng, doc)
	mustAccept(t, reply, code)
	// both rows redden on their own behavior; the next slot's focus
	// card carries each red row's want line in full (the door law)
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "suite=red")
	next, err := eng.Next()
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"err boom", "state out.txt text hi", "state y.txt from-run out.txt", "bytes sample.bin"} {
		if !strings.Contains(next, want) {
			t.Fatalf("expectation line misses %q:\n%s", want, next)
		}
	}
}

func TestCachedNextTextFault(t *testing.T) {
	stack, eng := newInst(t)
	if _, err := eng.Next(); err != nil {
		t.Fatal(err)
	}
	knobRead(stack, filepath.Join(eng.Store.Root(), "cache", "next.txt"), func() {
		next, err := eng.Next()
		if err != nil || !strings.Contains(next, "STAGE:") {
			t.Fatalf("next under cache-text fault: %v %q", err, next)
		}
	})
}

func TestCacheKeySurfaceGone(t *testing.T) {
	stack, eng := newInst(t)
	submit(t, stack, eng, specDoc("B1", "hello"))
	submit(t, stack, eng, "kind: run\n")
	submit(t, stack, eng, "kind: run\nratify: true\n")
	good := eng.cacheKey()
	stack.fs.walkErr = errIO
	gone := eng.cacheKey()
	stack.fs.walkErr = nil
	if good == gone || good == "" || gone == "" {
		t.Fatalf("surface key: %q vs %q", good, gone)
	}
}

func TestJournalHeadEmptyAndFirstLineOf(t *testing.T) {
	stack := bindUnits(t)
	eng, err := Init(stack.dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := eng.journalHead(); got != "-" {
		t.Fatalf("empty journal head: %q", got)
	}
	long := firstLineOf(strings.Repeat("ж", 100))
	if !strings.HasSuffix(long, "…") || len([]rune(long)) != 81 {
		t.Fatalf("rune-cut line: %q (%d runes)", long, len([]rune(long)))
	}
}

func TestLogVerbStateFault(t *testing.T) {
	stack, eng := newInst(t)
	before := len(eng.Ledger.All())
	knobRead(stack, filepath.Join(eng.Store.Root(), "state.yaml"), func() {
		eng.logVerb("next", "ok", nil, 1, "x")
	})
	if after := len(eng.Ledger.All()); after != before {
		t.Fatalf("ledger grew under fault: %d -> %d", before, after)
	}
}

// --- the verdict's fault paths ---

func TestVerdictFaultPaths(t *testing.T) {
	stack, eng := newInst(t)
	submit(t, stack, eng, specDoc("B1", "hello"))
	root := eng.Store.Root()
	b1 := filepath.Join(root, "canon", "atoms", "b1.yaml")
	knobRead(stack, filepath.Join(root, "state.yaml"), func() {
		if _, err := eng.RenderVerdict(); err == nil {
			t.Fatal("verdict under state fault succeeded")
		}
		if _, err := eng.renderVerdictFresh(); err == nil {
			t.Fatal("fresh verdict under state fault succeeded")
		}
	})
	knobRead(stack, b1, func() {
		if _, err := eng.renderVerdictFresh(); err == nil {
			t.Fatal("fresh verdict under corpus fault succeeded")
		}
	})
	// At deliver: a missing snapshot (the legacy shape) falls back to
	// a fresh render, which reads the corpus — the knob bites there.
	submit(t, stack, eng, "kind: run\n")
	submit(t, stack, eng, "kind: run\nratify: true\n")
	delete(stack.fs.files, filepath.Join(root, "verdicts", "v1.txt"))
	knobRead(stack, b1, func() {
		if _, err := eng.RenderVerdict(); err == nil {
			t.Fatal("deliver verdict under corpus fault succeeded")
		}
	})
}

func TestAcceptanceDriftShapes(t *testing.T) {
	stack, eng := newInst(t)
	submit(t, stack, eng, specDoc("B1", "hello"))
	submit(t, stack, eng, "kind: run\n")
	stack.fs.walkErr = errIO
	drifted, oldD, newD := eng.acceptanceDrift()
	stack.fs.walkErr = nil
	if drifted || oldD == "" || newD != "" {
		t.Fatalf("snapshot fault drift: %v %q %q", drifted, oldD, newD)
	}
	// An acceptance moment without a digest says nothing about drift.
	le := ledger.NewEntry("gate", canon.StageImplement)
	le.Details = map[string]string{"name": "suite", "scope": "acceptance", "outcome": "green"}
	if err := eng.Ledger.Append(le); err != nil {
		t.Fatal(err)
	}
	drifted, oldD, newD = eng.acceptanceDrift()
	if drifted || oldD != "" || newD == "" {
		t.Fatalf("digestless acceptance drift: %v %q %q", drifted, oldD, newD)
	}
}

// --- the run transaction's substrate ---

func TestRunRowsSnapshotFault(t *testing.T) {
	stack, eng := newInst(t)
	submit(t, stack, eng, specDoc("B1", "hello"))
	stack.fs.walkErr = errIO
	reply, code := submit(t, stack, eng, "kind: run\n")
	stack.fs.walkErr = nil
	mustAccept(t, reply, code, "rows not run", "io failure")
}

func TestRunMovedTreeFlipsGreenRows(t *testing.T) {
	stack, eng := newInst(t)
	doc := "kind: spec\nbehaviors:\n  - id: A1\n    name: n\n    prose: p\n    rows:\n      - run: [sh, -c, \"echo x > note.txt\"]\n  - id: A2\n    name: n\n    prose: p\n    rows:\n      - run: \"echo hello\"\n        out: wrong\n        form: \"corpus: the wrong surface\"\n"
	reply, code := submit(t, stack, eng, doc)
	mustAccept(t, reply, code)
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "tree moved during the run", "note.txt")
	atoms, _ := eng.Store.Atoms()
	if !atoms[0].Rows[0].EnvFail || atoms[0].Rows[0].Outcome != canon.RowRed {
		t.Fatalf("green row not flipped: %+v", atoms[0].Rows[0])
	}
	if atoms[1].Rows[0].EnvFail {
		t.Fatalf("behavior red marked env: %+v", atoms[1].Rows[0])
	}
}

func TestRunBuildFaults(t *testing.T) {
	stack, eng := newInst(t)
	knobRead(stack, filepath.Join(eng.Store.Root(), "state.yaml"), func() {
		if err := eng.runBuild(); err == nil {
			t.Fatal("build under state fault succeeded")
		}
	})
	// A build script that hangs: the build ceiling answers.
	if err := stack.fs.WriteFile(filepath.Join(eng.Workdir, "tool.sh"), []byte("#!/bin/sh\ntrue\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	stack.proc.plan["tool.sh"] = execport.Outcome{TimedOut: true}
	reply, code := submit(t, stack, eng, "kind: spec\nbuild:\n  - ./tool.sh\nbehaviors:\n  - id: B1\n    name: n\n    prose: p\n    rows:\n      - run: \"true\"\n")
	mustAccept(t, reply, code)
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "the build failed", "timed out")
	// A failing build with output: the output is the reason.
	reply, code = submit(t, stack, eng, "kind: spec\nfull-replace: true\nbuild: [sh, -c, \"echo boom 1>&2; exit 3\"]\nbehaviors:\n  - id: B1\n    name: n\n    prose: p\n    rows:\n      - run: \"true\"\n")
	mustAccept(t, reply, code)
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "the build failed", "boom")
	// A failing build with no output: the exit code is the reason.
	reply, code = submit(t, stack, eng, "kind: spec\nfull-replace: true\nbuild: [false]\nbehaviors:\n  - id: B1\n    name: n\n    prose: p\n    rows:\n      - run: \"true\"\n")
	mustAccept(t, reply, code)
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "the build failed", "exit 1")
	// A succeeding build: the rows run on top of it.
	reply, code = submit(t, stack, eng, "kind: spec\nfull-replace: true\nbuild: [true]\nbehaviors:\n  - id: B1\n    name: n\n    prose: p\n    rows:\n      - run: \"true\"\n")
	mustAccept(t, reply, code)
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "suite=green")
}

func TestTreeMovedDuringShapes(t *testing.T) {
	stack, eng := newInst(t)
	submit(t, stack, eng, specDoc("B1", "hello"))
	ckp, err := eng.snapshotTree()
	if err != nil {
		t.Fatal(err)
	}
	stack.fs.walkErr = errIO
	if moved := eng.treeMovedDuring(ckp); !strings.Contains(moved, "tree moved") {
		t.Fatalf("walk fault moved line: %q", moved)
	}
	stack.fs.walkErr = nil
	// The state read fails: the control runs without the spec's
	// ignores (the honest degraded form), and the underlying snapshot
	// read fails on its own just the same.
	knobRead(stack, filepath.Join(eng.Store.Root(), "state.yaml"), func() {
		if _, err := eng.snapshotTreeIgnores(nil); err == nil {
			t.Fatal("snapshot under state fault succeeded")
		}
		if moved := eng.treeMovedDuring(ckp); !strings.Contains(moved, "tree moved") {
			t.Fatalf("moved line under state fault: %q", moved)
		}
	})
	// Files that left the surface between the snapshots: the line is
	// capped with the +more suffix (AGENTS.md is on the surface too,
	// so budget+1 crafted paths make budget+2 moved in total).
	budget := canondata.Limit("slot.list-budget")
	var files []checkpointFile
	for i := 0; i < budget+1; i++ {
		files = append(files, checkpointFile{Path: "gone" + strconv.Itoa(i), SHA256: "d"})
	}
	if moved := eng.treeMovedDuring(&Checkpoint{Files: files}); !strings.Contains(moved, "(+2 more)") {
		t.Fatalf("capped moved line: %q", moved)
	}
}

func TestAcceptanceNoteNilAndMustStateFault(t *testing.T) {
	stack, eng := newInst(t)
	submit(t, stack, eng, specDoc("B1", "hello"))
	b := newBuilder(eng, mustState(eng))
	before := len(eng.Ledger.All())
	b.acceptanceNote(nil)
	if after := len(eng.Ledger.All()); after != before {
		t.Fatalf("nil checkpoint acceptance: %d -> %d", before, after)
	}
	knobRead(stack, filepath.Join(eng.Store.Root(), "state.yaml"), func() {
		if st := mustState(eng); st.Stage != "" || st.SchemaVersion != 0 {
			t.Fatalf("state under fault: %+v", st)
		}
	})
}

var _ = fsport.Bind

// The production bindings the unit stack shadows for every test:
// exercised once directly (no bindUnits in this test — the seam
// restores the originals after each bound test anyway).
func TestProductionClockAndAlwaysGate(t *testing.T) {
	if d := timeSince(time.Now()); d < 0 {
		t.Fatalf("negative since: %v", d)
	}
	if !stageGateFn("always")(nil) {
		t.Fatal("the always gate refused")
	}
}
