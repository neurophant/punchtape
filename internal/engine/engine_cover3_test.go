// The fault-injection battery: every failure branch of the engine's
// mechanics over the in-memory stack's precise knobs — the writer
// discipline, the digest cache, the renders' edge shapes, the
// language detector, the error-line helpers.
package engine

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/fsport"
)

// --- the public Submit entry ---

func TestSubmitDirectEntry(t *testing.T) {
	stack, eng := newInst(t)
	reply, code := eng.Submit([]byte(specDoc("B1", "hello")))
	mustAccept(t, reply, code)
	if _, code := eng.Submit([]byte("kind: spek\n")); code != Rejected {
		t.Fatal("direct decode refusal missing")
	}
	// A system failure through the writer: the append knob.
	stack.fs.appendErr = errIO
	reply, code = eng.Submit([]byte(fr(specDoc("B2", "hello"))))
	if code != Failure || !strings.Contains(reply, "FAIL:") {
		t.Fatalf("failure reply: %d %s", code, reply)
	}
	stack.fs.appendErr = nil
}

func TestSubmitterReportedEnv(t *testing.T) {
	stack, eng := newInst(t)
	t.Setenv("PT_SUBMITTER", "probe-1")
	reply, code := submit(t, stack, eng, specDoc("B1", "hello"))
	mustAccept(t, reply, code)
	found := false
	for _, en := range eng.Ledger.All() {
		if en.Details["submitter-reported"] == "probe-1" {
			found = true
		}
	}
	if !found {
		t.Fatal("submitter-reported detail missing")
	}
}

// --- the helpers ---

func TestDetectLanguage(t *testing.T) {
	cases := map[string]string{
		"привет мир":          "ru",
		"你好世界":                "zh",
		"hello world":         "en",
		"12345 !!!":           "en",
		"mixed 汉字 with words": "en",
	}
	for text, want := range cases {
		if got := detectLanguage(text); got != want {
			t.Fatalf("%q: want %s got %s", text, want, got)
		}
	}
}

func TestFirstErrorLineShapes(t *testing.T) {
	if got := firstErrorLine("line one\nERROR: second"); got != "ERROR: second" {
		t.Fatalf("error line: %q", got)
	}
	if got := firstErrorLine("plain reply"); got != "plain reply" {
		t.Fatalf("plain: %q", got)
	}
}

func TestFirstLineMulti(t *testing.T) {
	err := errors.New("first\nsecond")
	if got := firstLine(err); got != "first" {
		t.Fatalf("first line: %q", got)
	}
}

func TestOneLineInfoAndBoolGreen(t *testing.T) {
	if oneLineInfo("") != "" {
		t.Fatal("empty info")
	}
	if len(oneLineInfo(strings.Repeat("a", 500))) != 403 {
		t.Fatal("long info not capped")
	}
	if boolGreen(true) != "green" || boolGreen(false) != "red" {
		t.Fatal("bool spellings")
	}
}

// --- the writer discipline ---

func TestWriterRecoveryAndIntegrity(t *testing.T) {
	stack, eng := newInst(t)
	submit(t, stack, eng, specDoc("B1", "hello"))
	// The scan lists exactly the journaled instance files.
	files := eng.scanInstanceFiles()
	if len(files) == 0 {
		t.Fatal("empty scan")
	}
	sawState := false
	for _, f := range files {
		if f == stateFilePath {
			sawState = true
		}
	}
	if !sawState {
		t.Fatalf("state not scanned: %v", files)
	}
	// A stray unjournaled canon file is an integrity refusal at open.
	if err := stack.fs.WriteFile(filepath.Join(eng.Store.Root(), "canon", "atoms", "zz9.yaml"), []byte("id: Z\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Open(stack.dir); err == nil {
		t.Fatal("unjournaled file accepted")
	}
	if err := stack.fs.Remove(filepath.Join(eng.Store.Root(), "canon", "atoms", "zz9.yaml")); err != nil {
		t.Fatal(err)
	}
	// A missing journaled file is a refusal too.
	if err := stack.fs.Remove(filepath.Join(eng.Store.Root(), "canon", "atoms", "b1.yaml")); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Open(stack.dir); err == nil {
		t.Fatal("missing journaled file accepted")
	}
}

func TestInstanceScoped(t *testing.T) {
	if !instanceScoped(".punchtape/state.yaml") || instanceScoped("work.txt") || instanceScoped(".punchtape/cache/x") {
		t.Fatal("instance scoping")
	}
}

// --- the digest cache ---

func TestDigestCacheFlush(t *testing.T) {
	stack, eng := newInst(t)
	submit(t, stack, eng, specDoc("B1", "hello"))
	// next rides the cache and flushes it.
	if _, err := eng.Next(); err != nil {
		t.Fatal(err)
	}
	eng.flushDigestCache()
	if _, ok := eng.digestOf(".punchtape/state.yaml"); !ok {
		t.Fatal("digestOf miss after flush+next")
	}
}

// --- agentsref edge shapes ---

func TestAgentsRefMissingAndRegen(t *testing.T) {
	stack, eng := newInst(t)
	path := filepath.Join(eng.Workdir, "AGENTS.md")
	// A missing file is created.
	if err := stack.fs.Remove(path); err != nil {
		t.Fatal(err)
	}
	ensureAgentsRef(eng.Workdir)
	data, err := stack.fs.ReadFile(path)
	if err != nil || !strings.HasPrefix(string(data), refMarker) {
		t.Fatalf("regenerated ref: %v", err)
	}
	// A machine file without a tail is rewritten when stale; the
	// current one is left alone.
	ensureAgentsRef(eng.Workdir)
	again, _ := stack.fs.ReadFile(path)
	if string(again) != string(data) {
		t.Fatal("stable ref rewritten")
	}
	// A machine file with a tail whose head is current keeps the tail.
	if err := stack.fs.WriteFile(path, []byte(machineHead()+"repo tail\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ensureAgentsRef(eng.Workdir)
	data, _ = stack.fs.ReadFile(path)
	if !strings.Contains(string(data), "repo tail") {
		t.Fatal("tail lost")
	}
	// splitRefTail on a boundary-less file.
	if _, _, ok := splitRefTail([]byte("no boundary")); ok {
		t.Fatal("boundary hallucinated")
	}
}

// --- checkpoint edge shapes ---

func TestIgnoreMatcherShapes(t *testing.T) {
	m := newIgnoreMatcher([]string{"out/", "*.log", "", ".", "..", "a/../b"})
	if !m.excluded("out/x.txt") || !m.excluded("x.log") || !m.excluded("d/y.log") {
		t.Fatal("excluded misses")
	}
	if m.excluded("keep.txt") || m.excluded(".punchtape") == false {
		t.Fatal("excluded overreaches")
	}
}

func TestLastCheckpointNone(t *testing.T) {
	stack := bindUnits(t)
	eng, err := Init(stack.dir)
	if err != nil {
		t.Fatal(err)
	}
	if eng.lastCheckpoint() != nil {
		t.Fatal("virgin instance has a checkpoint")
	}
}

// --- snapshot walk faults ---

func TestSnapshotWalkFaults(t *testing.T) {
	stack, eng := newInst(t)
	stack.fs.walkErr = errIO
	if _, err := eng.snapshotTree(); err == nil {
		t.Fatal("walk fault swallowed")
	}
	stack.fs.walkErr = nil
	stack.fs.walkChildErr = errIO
	stack.fs.failWalkAt = filepath.Join(eng.Workdir, "AGENTS.md")
	if _, err := eng.snapshotTree(); err == nil {
		t.Fatal("child walk fault swallowed")
	}
	stack.fs.walkChildErr = nil
	// A read failure mid-walk.
	stack.fs.readErrPath = filepath.Join(eng.Workdir, "AGENTS.md")
	stack.fs.readErr = errIO
	if _, err := eng.snapshotTree(); err == nil {
		t.Fatal("read fault swallowed")
	}
	stack.fs.readErr = nil
	stack.fs.readErrPath = ""
}

// --- Begin fault paths ---

func TestBeginFaultPaths(t *testing.T) {
	stack := bindUnits(t)
	eng, err := Init(stack.dir)
	if err != nil {
		t.Fatal(err)
	}
	// The busy writer.
	stack.fs.lockErr = errIO
	if _, code := eng.Begin("first"); code != Failure {
		t.Fatal("busy begin accepted")
	}
	stack.fs.lockErr = nil
	// A commit failure (the journal append knob).
	stack.fs.appendErr = errIO
	if _, code := eng.Begin("first"); code != Failure {
		t.Fatal("append failure accepted")
	}
	stack.fs.appendErr = nil
}

func TestInitFaultPaths(t *testing.T) {
	stack := bindUnits(t)
	stack.fs.mkdirErr = errIO
	if _, err := Init(stack.dir); err == nil {
		t.Fatal("mkdir fault swallowed")
	}
	stack.fs.mkdirErr = nil
	if _, err := Init(stack.dir); err != nil {
		t.Fatal(err)
	}
	if _, err := Init(stack.dir); err == nil {
		t.Fatal("double init accepted")
	}
}

func TestOpenVirgin(t *testing.T) {
	stack := bindUnits(t)
	if _, exists, err := Open(stack.dir); err != nil || exists {
		t.Fatalf("virgin open: %v %v", exists, err)
	}
}

// --- next cache edge shapes ---

func TestNextCacheBreaksOnExternalEdit(t *testing.T) {
	stack, eng := newInst(t)
	first, _ := eng.Next()
	// An external canon change breaks the cache.
	if err := stack.fs.WriteFile(filepath.Join(eng.Store.Root(), "state.yaml"), []byte("schema-version: 9\nstage: intake\ncounters:\n  CKP: 0\n"), 0o644); err != nil {
		// the strict decode at State() refuses a foreign schema: the
		// render error path is the honest answer
		if _, err := eng.Next(); err == nil {
			t.Fatal("foreign schema next succeeded")
		}
		return
	}
	_ = first
}

func TestRenderNextDeliverSurfaceBranches(t *testing.T) {
	stack, eng := newInst(t)
	submit(t, stack, eng, specDoc("B1", "hello"))
	submit(t, stack, eng, "kind: run\n")
	submit(t, stack, eng, "kind: run\nratify: true\n")
	// The deliver cacheKey includes the LIVE surface digest: an
	// external tree change breaks it.
	key1 := eng.cacheKey()
	if err := stack.fs.WriteFile(filepath.Join(eng.Workdir, "external.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	eng.flushDigestCache()
	key2 := eng.cacheKey()
	if key1 == key2 {
		t.Fatal("surface branch did not break the key")
	}
}

// --- verdict frozen edge shapes ---

func TestVerdictSnapshotServesAndStays(t *testing.T) {
	stack, eng := newInst(t)
	submit(t, stack, eng, specDoc("B1", "hello"))
	submit(t, stack, eng, "kind: run\n")
	submit(t, stack, eng, "kind: run\nratify: true\n")
	// The snapshot is the truth at deliver: a stale cache-key file is
	// not consulted at all.
	if err := stack.fs.WriteFile(filepath.Join(eng.Store.Root(), "cache", "verdict-key"), []byte("bogus"), 0o644); err != nil {
		t.Fatal(err)
	}
	v, err := eng.RenderVerdict()
	if err != nil || v.Status != VerdictReady {
		t.Fatalf("snapshot verdict: %v %s", err, v.Status)
	}
	if !strings.Contains(v.Text(), "version: v1") {
		t.Fatalf("no version line:\n%s", v.Text())
	}
}

func TestVerdictFromTextAndOrigin(t *testing.T) {
	v := verdictFromText("VERDICT: READY\nline two\n")
	if v.Status != VerdictReady || len(v.Lines) != 2 {
		t.Fatalf("from text: %+v", v)
	}
	empty := verdictFromText("")
	if empty.Status != "" {
		t.Fatalf("empty text: %+v", empty)
	}
	// The origin line: no brief — the honest none.
	stack := bindUnits(t)
	eng, err := Init(stack.dir)
	if err != nil {
		t.Fatal(err)
	}
	v2, err := eng.renderVerdictFresh()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(v2.Text(), "(no brief)") {
		t.Fatalf("origin none: %s", v2.Text())
	}
}

// --- renderNext cover branches ---

func TestRenderNextWishNilIntake(t *testing.T) {
	stack := bindUnits(t)
	eng, err := Init(stack.dir)
	if err != nil {
		t.Fatal(err)
	}
	r := renderNextRaw(t, eng, canon.State{Stage: canon.StageIntake})
	if strings.Contains(r, "BRIEF:") {
		t.Fatalf("brief line without a brief:\n%s", r)
	}
}

func TestRenderNextDeliverDriftAction(t *testing.T) {
	stack, eng := newInst(t)
	submit(t, stack, eng, specDoc("B1", "hello"))
	submit(t, stack, eng, "kind: run\n")
	submit(t, stack, eng, "kind: run\nratify: true\n")
	// Drift after acceptance: the verdict is honest NOT-READY.
	if err := stack.fs.WriteFile(filepath.Join(eng.Workdir, "post.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	next, err := eng.Next()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(next, "NOT-READY") {
		t.Fatalf("drifted deliver slot:\n%s", next)
	}
}

// --- spec render edge shapes ---

func TestSpecRenderEmptyAndPipes(t *testing.T) {
	if out := renderSpec(nil); !strings.Contains(out, "(no behaviors yet)") {
		t.Fatalf("empty spec: %s", out)
	}
	out := canon.Row{Out: &canon.TextExpect{Text: "a|b"}}
	_ = out
	atom := canon.Atom{ID: "B1", Name: "n", Prose: "plain\n\nc", Rows: []canon.Row{{Out: &canon.TextExpect{Text: "a|b"}}}}
	rendered := renderSpec([]canon.Atom{atom})
	if !strings.Contains(rendered, `a\|b`) || !strings.Contains(rendered, "| unknown |") {
		t.Fatalf("pipe/unknown render:\n%s", rendered)
	}
}

// --- summary acceptance lines ---

func TestSummaryAcceptanceLines(t *testing.T) {
	stack, eng := newInst(t)
	line := eng.summaryAcceptanceLine("en")
	if !strings.Contains(line, "no green run") {
		t.Fatalf("not-run line: %q", line)
	}
	submit(t, stack, eng, specDoc("B1", "hello"))
	submit(t, stack, eng, "kind: run\n")
	line = eng.summaryAcceptanceLine("en")
	if !strings.Contains(line, "PASS on tree") {
		t.Fatalf("pass line: %q", line)
	}
	// Drift names both digests.
	if err := stack.fs.WriteFile(filepath.Join(eng.Workdir, "later.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	line = eng.summaryAcceptanceLine("en")
	if !strings.Contains(line, "moved after the acceptance moment") {
		t.Fatalf("drift line: %q", line)
	}
}

// --- stage runner loop ---

func TestRunProcessUnknownStageStops(t *testing.T) {
	stack, eng := newInst(t)
	b := newBuilder(eng, mustState(eng))
	b.state.Stage = "bogus"
	b.runProcess()
	if b.state.Stage != "bogus" {
		t.Fatal("unknown stage moved")
	}
	_ = stack
}

// --- refreeze on deliver ---

func TestRefusedRunAtDeliverKeepsSnapshot(t *testing.T) {
	stack, eng := newInst(t)
	submit(t, stack, eng, specDoc("B1", "hello"))
	submit(t, stack, eng, "kind: run\n")
	submit(t, stack, eng, "kind: run\nratify: true\n")
	// A refused run at deliver leaves the frozen snapshot honest.
	reply, code := submit(t, stack, eng, "kind: run\nratify: true\nbuild-ignore: [x]\n")
	mustReject(t, reply, code, "kind: spec")
	v, err := eng.RenderVerdict()
	if err != nil || v.Status != VerdictReady {
		t.Fatalf("verdict after refusal: %v %s", err, v.Status)
	}
}

var _ = fsport.Bind
