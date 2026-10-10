// The engine battery of the full cycle: birth → spec → run →
// ratify over the in-memory stack; every form refusal of the
// grammar; the execution transaction (build, rows, tree identity,
// checkpoints); the renders (slot, verdict, summary, spec.md,
// AGENTS.md, reference); the door; the writer discipline.
package engine

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/execport"
	"github.com/neurophant/punchtape/internal/fsport"
	"github.com/neurophant/punchtape/internal/journal"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// newInst — a bound stack with a born instance (Init + Begin).
func newInst(t *testing.T) (*fakeStack, *Engine) {
	t.Helper()
	stack := bindUnits(t)
	eng, err := Init(stack.dir)
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	if reply, code := eng.Begin("a tool that greps lines in files"); code != Accepted {
		t.Fatalf("begin: %d %s", code, reply)
	}
	return stack, eng
}

// door — the one legal submission path.
func door() string { return filepath.ToSlash(inboxCurrent) }

// submit — write the document into the door file and submit it.
func submit(t *testing.T, stack *fakeStack, eng *Engine, doc string) (string, int) {
	t.Helper()
	if err := stack.fs.WriteFile(filepath.Join(eng.Workdir, door()), []byte(doc), 0o644); err != nil {
		t.Fatalf("door write: %v", err)
	}
	return eng.SubmitSet(door())
}

// fr — the deliberate full-rewrite spelling of a spec document:
// a FULL document after the first accepted spec rides
// full-replace: true; the v2-semantics tests spell the hatch).
func fr(doc string) string {
	return strings.Replace(doc, "kind: spec\n", "kind: spec\nfull-replace: true\n", 1)
}

// specDoc — a minimal one-behavior spec.
func specDoc(id, out string) string {
	return "kind: spec\nbehaviors:\n  - id: " + id + "\n    name: sample\n    prose: |\n      One paragraph.\n    rows:\n      - run: \"echo hello\"\n        out: \"" + out + "\"\n        rc: 0\n        form: \"test corpus: the pinned surface\"\n"
}

func mustAccept(t *testing.T, reply string, code int, want ...string) {
	t.Helper()
	if code != Accepted {
		t.Fatalf("want accepted, got %d: %s", code, reply)
	}
	for _, w := range want {
		if !strings.Contains(reply, w) {
			t.Fatalf("reply misses %q:\n%s", w, reply)
		}
	}
}

func mustReject(t *testing.T, reply string, code int, want ...string) {
	t.Helper()
	if code != Rejected {
		t.Fatalf("want rejected, got %d: %s", code, reply)
	}
	for _, w := range want {
		if !strings.Contains(reply, w) {
			t.Fatalf("reply misses %q:\n%s", w, reply)
		}
	}
}

// --- the cycle ---

func TestCycleSpecRunRatifyReady(t *testing.T) {
	stack, eng := newInst(t)
	// The slot after the birth names the door and the spec action.
	next, err := eng.Next()
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"STAGE: intake", door(), "ACTION: write the spec"} {
		if !strings.Contains(next, want) {
			t.Fatalf("intake slot misses %q:\n%s", want, next)
		}
	}
	reply, code := submit(t, stack, eng, specDoc("B1", "hello"))
	mustAccept(t, reply, code, "SPEC v1", "ACCEPTED: transaction", "BYTES: sha256")
	// the slot rides next (the ack is thin)
	next2, _ := eng.Next()
	if !strings.Contains(next2, "STAGE: implement") || !strings.Contains(next2, "the rows have not run yet") {
		t.Fatalf("the post-spec slot:\n%s", next2)
	}
	// A spec never runs: the row outcome stays unknown.
	atoms, _ := eng.Store.Atoms()
	if atoms[0].Rows[0].Outcome != canon.RowUnknown && atoms[0].Rows[0].Outcome != "" {
		t.Fatalf("spec ran rows: %q", atoms[0].Rows[0].Outcome)
	}
	// The bare re-run: green, then ratify.
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "suite=green (1 rows)")
	// The idempotent replay: the same content-bearing document repeats
	// honestly (a lost reply does not double-apply).
	reply, code = submit(t, stack, eng, specDoc("B1", "hello"))
	mustAccept(t, reply, code, "IDEMPOTENT REPLAY", "the work is done")
	// A bare re-run driver never dedupes as a FACT (a fresh key per
	// application) — but with the content-addressed memo it executes
	// only what went stale: at unchanged inputs nothing re-runs (the
	// memo IS the re-verification; a touched input staleness its row).
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "all fresh")
	reply, code = submit(t, stack, eng, "kind: run\nratify: true\n")
	mustAccept(t, reply, code, "VERDICT: READY", "digest:")
	// Deliver: frozen verdict, closed door.
	next, err = eng.Next()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(next, "STAGE: deliver") || !strings.Contains(next, "VERDICT: READY") {
		t.Fatalf("deliver slot:\n%s", next)
	}
	// The case lives on: a spec at deliver is the NEXT VERSION — the
	// same door accepts it and the work reopens.
	reply, code = submit(t, stack, eng, fr(specDoc("B2", "hello")))
	mustAccept(t, reply, code, "SPEC v2")
	// The verdict snapshot and the renders exist; the snapshot is
	// immutable across the versions that follow.
	v1Path := filepath.Join(eng.Store.Root(), "verdicts", "v1.txt")
	v1, err := stack.fs.ReadFile(v1Path)
	if err != nil {
		t.Fatalf("verdicts/v1.txt: %v", err)
	}
	if !strings.Contains(string(v1), "VERDICT: READY") || !strings.Contains(string(v1), "version: v1") {
		t.Fatalf("snapshot v1:\n%s", v1)
	}
	// The signature law: the run measures the new row, the ratify
	// signs the recorded green (nothing re-runs at the closing).
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "suite=green")
	reply, code = submit(t, stack, eng, "kind: run\nratify: true\n")
	mustAccept(t, reply, code, "VERDICT: READY", "(v2)")
	again, _ := stack.fs.ReadFile(v1Path)
	if string(again) != string(v1) {
		t.Fatal("snapshot v1 mutated after v2")
	}
	// A RUN at deliver is still refused — only the next spec version
	// reopens the case.
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustReject(t, reply, code, "kind: spec")
	for _, name := range []string{"spec.md", "SUMMARY.md", "AGENTS.md"} {
		if name == "AGENTS.md" {
			continue
		}
		if _, err := stack.fs.ReadFile(filepath.Join(eng.Store.Root(), filepath.FromSlash(name))); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
	}
	if _, err := stack.fs.ReadFile(filepath.Join(eng.Workdir, "AGENTS.md")); err != nil {
		t.Fatalf("AGENTS.md: %v", err)
	}
}

func TestCycleRedRepairByBehaviorID(t *testing.T) {
	stack, eng := newInst(t)
	reply, code := submit(t, stack, eng, specDoc("B1", "wrong"))
	mustAccept(t, reply, code)
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "suite=red (1 rows, 1 red)", "  B1/1:", "not equal to \"wrong\"", "first difference at byte 0")
	// The standing red rides the run ack (the measured table) and the
	// next slot's changed list and focus card (the door law).
	mustAccept(t, reply, code, "suite=red (1 rows, 1 red)")
	next, _ := eng.Next()
	if !strings.Contains(next, "B1/1: red (was") {
		t.Fatalf("slot changed list:\n%s", next)
	}
	if !strings.Contains(next, "not equal to \"wrong\"") {
		t.Fatalf("slot red row in full:\n%s", next)
	}
	// The repair: the same behavior id with a corrected row (a
	// second FULL document spells the deliberate hatch).
	reply, code = submit(t, stack, eng, fr(specDoc("B1", "hello")))
	mustAccept(t, reply, code)
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "suite=green")
	reply, code = submit(t, stack, eng, "kind: run\nratify: true\n")
	mustAccept(t, reply, code, "VERDICT: READY")
}

// The signature law at the door: a ratify without a green run behind
// it is REFUSED with every never-ran row named — the closing is a
// signature on a recorded green, never the measurement itself.
func TestCycleRatifyNotReadyStaysImplement(t *testing.T) {
	stack, eng := newInst(t)
	submit(t, stack, eng, specDoc("B1", "wrong"))
	reply, code := submit(t, stack, eng, "kind: run\nratify: true\n")
	mustReject(t, reply, code, "B1/1", "kind: run")
	st, _ := eng.Store.State()
	if st.Stage != canon.StageImplement {
		t.Fatalf("stage moved on a refused ratify: %s", st.Stage)
	}
}

func TestRunBeforeSpecRefused(t *testing.T) {
	stack, eng := newInst(t)
	reply, code := submit(t, stack, eng, "kind: run\n")
	mustReject(t, reply, code, "needs an accepted spec first")
}

// --- decode and door refusals ---

func TestDecodeRefusals(t *testing.T) {
	stack, eng := newInst(t)
	cases := []struct {
		name, doc, want string
	}{
		{"unknown kind", "kind: spek\n", "unknown kind"},
		{"unknown kind hint", "kind: spek\n", "kind: spec"},
		{"unknown field", "kind: spec\nbogus: 1\n", "bogus"},
		{"not a mapping", "- a\n", "ERROR"},
		{"ratify in spec", "kind: spec\nratify: true\nbehaviors:\n  - id: B1\n    name: n\n    prose: p\n    rows:\n      - run: \"true\"\n", "ratify: true is only legal in kind: run"},
		{"answers", "kind: spec\nanswers:\n  - q: Q1\n    choice: a\nbehaviors:\n  - id: B1\n    name: n\n    prose: p\n    rows:\n      - run: \"true\"\n", "answers without an open question batch"},
		{"behaviors empty spec", "kind: spec\n", "no behaviors"},
	}
	for _, c := range cases {
		reply, code := submit(t, stack, eng, c.doc)
		if code != Rejected {
			t.Fatalf("%s: want rejected, got %d: %s", c.name, code, reply)
		}
		if !strings.Contains(reply, c.want) {
			t.Fatalf("%s: reply misses %q:\n%s", c.name, c.want, reply)
		}
		if !strings.Contains(reply, "the draft stays") {
			t.Fatalf("%s: no stays pointer:\n%s", c.name, reply)
		}
	}
}

func TestDoorImpermeability(t *testing.T) {
	stack, eng := newInst(t)
	reply, code := eng.SubmitSet("d-stray.yaml")
	mustReject(t, reply, code, "ERROR: door:", "not the door file", door())
	reply, code = eng.SubmitSet(".punchtape/inbox/0001-spec.yaml")
	mustReject(t, reply, code, "ERROR: door:", "not the door file")
	// A missing door file is an honest failure, not a rejection.
	if err := stack.fs.Remove(filepath.Join(eng.Workdir, door())); err != nil {
		t.Fatal(err)
	}
	reply, code = eng.SubmitSet(door())
	if code != Failure || !strings.Contains(reply, "FAIL: reading") {
		t.Fatalf("missing door: %d %s", code, reply)
	}
	// The door is re-created by next and reset to the template after
	// an accepted submission.
	if _, err := eng.Next(); err != nil {
		t.Fatal(err)
	}
	data, err := stack.fs.ReadFile(filepath.Join(eng.Workdir, door()))
	if err != nil {
		t.Fatalf("door not re-created: %v", err)
	}
	if !strings.Contains(string(data), "kind: spec") {
		t.Fatalf("door template missing: %q", string(data)[:40])
	}
	submit(t, stack, eng, specDoc("B1", "hello"))
	data, _ = stack.fs.ReadFile(filepath.Join(eng.Workdir, door()))
	if string(data) != string(canondata.TemplateBytes("spec.txt")) {
		t.Fatal("door not reset to the template after acceptance")
	}
}

// --- form validation of behaviors ---

func TestBehaviorFormRefusals(t *testing.T) {
	stack, eng := newInst(t)
	base := "kind: spec\nbuild: [true]\nbehaviors:\n"
	cases := []struct {
		name, doc, want string
	}{
		{"bad id", base + "  - id: \"\"\n    name: n\n    prose: p\n    rows:\n      - run: \"true\"\n", "behavior id"},
		{"dup id", base + "  - id: B1\n    name: n\n    prose: p\n    rows:\n      - run: \"true\"\n  - id: B1\n    name: n\n    prose: p\n    rows:\n      - run: \"true\"\n", "appears twice"},
		{"name empty", base + "  - id: B1\n    name: \"\"\n    prose: p\n    rows:\n      - run: \"true\"\n", "living name"},
		{"prose empty", base + "  - id: B1\n    name: n\n    prose: \"\"\n    rows:\n      - run: \"true\"\n", "prose is empty"},
		{"rows empty", base + "  - id: B1\n    name: n\n    prose: p\n    rows: []\n", "rows are empty"},
		{"run empty", base + "  - id: B1\n    name: n\n    prose: p\n    rows:\n      - run: \"\"\n", "empty"},
		{"timeout over max", base + "  - id: B1\n    name: n\n    prose: p\n    rows:\n      - run: \"true\"\n        timeout: 301\n", "outside 1..300"},
		{"rc negative", base + "  - id: B1\n    name: n\n    prose: p\n    rows:\n      - run: \"true\"\n        rc: -1\n", "0 or greater"},
		{"seed absolute", base + "  - id: B1\n    name: n\n    prose: p\n    rows:\n      - run: \"true\"\n        seeds:\n          - path: /etc/passwd\n            text: x\n", "relative"},
		{"seed escape", base + "  - id: B1\n    name: n\n    prose: p\n    rows:\n      - run: \"true\"\n        seeds:\n          - path: a/../../b\n            text: x\n", ".."},
		{"from-input missing", base + "  - id: B1\n    name: n\n    prose: p\n    rows:\n      - run: \"true\"\n        seeds:\n          - path: in.txt\n            from-input: data/x.bin\n", "no such file"},
		{"state escape", base + "  - id: B1\n    name: n\n    prose: p\n    rows:\n      - run: \"true\"\n        state:\n          - path: ../x\n            text: v\n", ".."},
		{"from-run escape", base + "  - id: B1\n    name: n\n    prose: p\n    rows:\n      - run: \"true\"\n        state:\n          - path: ok\n            from-run: ../x\n", ".."},
		{"bytes wrong form", base + "  - id: B1\n    name: n\n    prose: p\n    rows:\n      - run: \"true\"\n        bytes: sample.bin\n", "file://name"},
		{"bytes missing", base + "  - id: B1\n    name: n\n    prose: p\n    rows:\n      - run: \"true\"\n        bytes: file://gone.bin\n", "missing next to the submission"},
	}
	for _, c := range cases {
		reply, code := submit(t, stack, eng, c.doc)
		if code != Rejected {
			t.Fatalf("%s: want rejected, got %d: %s", c.name, code, reply)
		}
		if !strings.Contains(reply, c.want) {
			t.Fatalf("%s: reply misses %q:\n%s", c.name, c.want, reply)
		}
	}
}

// --- execution: build, rows, tree identity ---

func TestRunBuildFailureMarksRowsNotRun(t *testing.T) {
	stack, eng := newInst(t)
	stack.proc.plan["make build"] = execport.Outcome{Exit: 2, Stderr: "boom"}
	submit(t, stack, eng, "kind: spec\nbuild: [make, build]\nbehaviors:\n  - id: B1\n    name: n\n    prose: p\n    rows:\n      - run: \"echo hello\"\n        out: \"hello\"\n        form: \"corpus: the pinned surface\"\n")
	reply, code := submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "suite=red", "rows not run", "the build failed")
	atoms, _ := eng.Store.Atoms()
	if !atoms[0].Rows[0].EnvFail {
		t.Fatal("build failure must be an env red")
	}
}

func TestRunRowTimeout(t *testing.T) {
	stack, eng := newInst(t)
	// A real script file in the tree (the runner launches it through
	// its tree path); its outcome rides the plan: a hang.
	if err := stack.fs.WriteFile(filepath.Join(eng.Workdir, "slowtool"), []byte("#!/bin/sh\nsleep 5\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	stack.proc.plan["slowtool"] = execport.Outcome{TimedOut: true}
	submit(t, stack, eng, "kind: spec\nbehaviors:\n  - id: B1\n    name: n\n    prose: p\n    rows:\n      - run: \"slowtool x\"\n        timeout: 5\n")
	reply, code := submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "suite=red", "timed out after 5s")
}

func TestRunTreeMovedFlipsGreens(t *testing.T) {
	stack, eng := newInst(t)
	submit(t, stack, eng, "kind: spec\nbuild-ignore: [junk]\nbehaviors:\n  - id: B1\n    name: n\n    prose: p\n    rows:\n      - run: \"echo hello\"\n        out: \"hello\"\n        form: \"corpus: the pinned surface\"\n")
	ckp, err := eng.snapshotTree()
	if err != nil {
		t.Fatal(err)
	}
	if err := stack.fs.WriteFile(filepath.Join(eng.Workdir, "stray.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	moved := eng.treeMovedDuring(ckp)
	if !strings.Contains(moved, "stray.txt") {
		t.Fatalf("moved line: %q", moved)
	}
	// An ignored path does not count.
	ckp2, _ := eng.snapshotTree()
	stack.fs.MkdirAll(filepath.Join(eng.Workdir, "junk"))
	if err := stack.fs.WriteFile(filepath.Join(eng.Workdir, "junk", "deep.bin"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if moved := eng.treeMovedDuring(ckp2); moved != "" {
		t.Fatalf("ignored path moved the tree: %q", moved)
	}
}

func TestCheckpointEffectAndAcceptance(t *testing.T) {
	stack, eng := newInst(t)
	submit(t, stack, eng, specDoc("B1", "hello"))
	submit(t, stack, eng, "kind: run\n")
	st, _ := eng.Store.State()
	if st.Counters["CKP"] == 0 {
		t.Fatal("no checkpoint recorded")
	}
	if ckp := eng.lastCheckpoint(); ckp == nil || ckp.Digest == "" {
		t.Fatal("last checkpoint unreadable")
	}
	// The acceptance note: all green on a run that ran rows.
	found := false
	for _, en := range eng.Ledger.All() {
		if en.Event == "gate" && en.Details["scope"] == "acceptance" && en.Details["outcome"] == "green" {
			found = true
		}
	}
	if !found {
		t.Fatal("no acceptance entry after a green run")
	}
}

// --- from-input and bytes samples ---

func TestFromInputSeedAndBytesSample(t *testing.T) {
	stack, eng := newInst(t)
	// The instance's input directory: the sibling input/ of the work root.
	in := filepath.Join(stack.dir, "..", "input")
	stack.fs.MkdirAll(in)
	if err := stack.fs.WriteFile(filepath.Join(in, "data", "x.bin"), []byte("raw-bytes"), 0o644); err != nil {
		stack.fs.MkdirAll(filepath.Join(in, "data"))
		stack.fs.WriteFile(filepath.Join(in, "data", "x.bin"), []byte("raw-bytes"), 0o644)
	}
	if err := stack.fs.WriteFile(filepath.Join(eng.Store.Root(), "inbox", "sample.bin"), []byte("hello\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	doc := "kind: spec\nbehaviors:\n  - id: B1\n    name: n\n    prose: p\n    rows:\n      - run: \"cat seed.txt\"\n        out: \"raw-bytes\"\n        form: \"corpus: the pinned surface\"\n        seeds:\n          - path: seed.txt\n            from-input: data/x.bin\n  - id: B2\n    name: n\n    prose: p\n    rows:\n      - run: \"echo hello\"\n        bytes: file://sample.bin\n"
	reply, code := submit(t, stack, eng, doc)
	mustAccept(t, reply, code)
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "suite=green")
	reply, code = submit(t, stack, eng, "kind: run\nratify: true\n")
	mustAccept(t, reply, code, "VERDICT: READY")
}

// --- the slot render branches ---

func TestSlotRenderBranches(t *testing.T) {
	stack, eng := newInst(t)
	// The intake slot with the brief line.
	next, _ := eng.Next()
	if !strings.Contains(next, "BRIEF: a tool that greps lines in files") {
		t.Fatalf("intake brief line:\n%s", next)
	}
	// A cached repeat is byte-identical.
	again, _ := eng.Next()
	if next != again {
		t.Fatal("cached next differs")
	}
	submit(t, stack, eng, specDoc("B1", "hello"))
	submit(t, stack, eng, "kind: run\n")
	next, _ = eng.Next()
	for _, want := range []string{"STAGE: implement", "STATE: atoms 1, rows 1, open 0", "ACTION: every row is green"} {
		if !strings.Contains(next, want) {
			t.Fatalf("green slot misses %q:\n%s", want, next)
		}
	}
	// An unknown stage renders an honest empty body.
	info := renderNextRaw(t, eng, canon.State{Stage: "bogus"})
	if !strings.Contains(info, "STAGE: bogus") {
		t.Fatalf("bogus stage render:\n%s", info)
	}
}

func renderNextRaw(t *testing.T, eng *Engine, st canon.State) string {
	t.Helper()
	r, _, err := eng.renderNext(st)
	if err != nil {
		t.Fatal(err)
	}
	return r.Text
}

func TestSessionHandoffLine(t *testing.T) {
	_, eng := newInst(t)
	for i := 0; i < canondata.Limit("compact.journal-entries"); i++ {
		if err := eng.Journal.Append(journal.Entry{Transaction: "tx" + string(rune('a'+i%26)) + string(rune('a'+i/26)), DeltaKind: "spec"}); err != nil {
			t.Fatal(err)
		}
	}
	next, _ := eng.Next()
	if !strings.Contains(next, "SESSION HANDOFF") {
		t.Fatalf("handoff line missing:\n%s", next)
	}
}

// --- verdict ---

func TestVerdictFreshAndFrozen(t *testing.T) {
	stack, eng := newInst(t)
	submit(t, stack, eng, specDoc("B1", "hello"))
	// NOT-READY before any run: rows unknown.
	v, err := eng.RenderVerdict()
	if err != nil {
		t.Fatal(err)
	}
	if v.Status != VerdictNotReady {
		t.Fatalf("fresh verdict: %s", v.Status)
	}
	if !strings.Contains(v.Text(), "VERDICT: NOT-READY") {
		t.Fatal("verdict head line missing")
	}
	submit(t, stack, eng, "kind: run\n")
	submit(t, stack, eng, "kind: run\nratify: true\n")
	v1, _ := eng.RenderVerdict()
	v2, _ := eng.RenderVerdict()
	if v1.Text() != v2.Text() {
		t.Fatal("frozen verdict drifted")
	}
	// Drift after acceptance: an honest NOT-READY.
	if err := stack.fs.WriteFile(filepath.Join(eng.Workdir, "post.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	eng.lastVerdictDrift = false
	v3, err := eng.renderVerdictFresh()
	if err != nil {
		t.Fatal(err)
	}
	if v3.Status != VerdictNotReady || !strings.Contains(v3.Text(), "moved after the acceptance moment") {
		t.Fatalf("drift verdict: %s\n%s", v3.Status, v3.Text())
	}
	// The digest line is stable across bill changes (normalize).
	if !strings.HasPrefix(v1.Lines[1], "digest: ") {
		t.Fatalf("digest line: %q", v1.Lines[1])
	}
}

func TestVerdictNoAtomsAndRedListCap(t *testing.T) {
	stack, eng := newInst(t)
	// No corpus: the verdict names it.
	eng2 := eng
	v, err := eng2.renderVerdictFresh()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(v.Text(), "no behaviors in the corpus") {
		t.Fatalf("empty corpus verdict:\n%s", v.Text())
	}
	// A red list over the budget carries the +more line.
	var doc strings.Builder
	doc.WriteString("kind: spec\nbehaviors:\n")
	for i := 0; i < canondata.Limit("slot.list-budget")+3; i++ {
		doc.WriteString("  - id: B")
		doc.WriteString(string(rune('a' + i)))
		doc.WriteString("\n    name: n\n    prose: p\n    rows:\n      - run: \"echo hello\"\n        out: wrong\n        form: \"corpus: the wrong surface\"\n")
	}
	reply, code := submit(t, stack, eng, doc.String())
	mustAccept(t, reply, code)
	submit(t, stack, eng, "kind: run\n")
	v, err = eng.renderVerdictFresh()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(v.Text(), "(+3 more)") {
		t.Fatalf("red list cap:\n%s", v.Text())
	}
}

// --- renders: spec.md, AGENTS.md, reference, SUMMARY.md ---

func TestSpecRenderShapes(t *testing.T) {
	stack, eng := newInst(t)
	submit(t, stack, eng, specDoc("B1", "wrong"))
	eng.ensureSpecRender()
	data, err := stack.fs.ReadFile(filepath.Join(eng.Store.Root(), "spec.md"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	for _, want := range []string{"## B1 — sample", "One paragraph.", "| 1 | `echo hello` |", "| wrong |"} {
		if want == "| wrong |" {
			continue
		}
		if !strings.Contains(s, want) {
			t.Fatalf("spec.md misses %q:\n%s", want, s)
		}
	}
	// The red reason rides below the table after a red run.
	submit(t, stack, eng, "kind: run\n")
	eng.ensureSpecRender()
	data, _ = stack.fs.ReadFile(filepath.Join(eng.Store.Root(), "spec.md"))
	if !strings.Contains(string(data), "last red:") {
		t.Fatalf("spec.md red reason:\n%s", string(data))
	}
	// A long expectation truncates in the table.
	long := strings.Repeat("x", 300)
	submit(t, stack, eng, "kind: spec\nbehaviors:\n  - id: B9\n    name: n\n    prose: p\n    rows:\n      - run: \"echo "+long+"\"\n        out: \""+long+"\"\n")
	eng.ensureSpecRender()
	data, _ = stack.fs.ReadFile(filepath.Join(eng.Store.Root(), "spec.md"))
	if !strings.Contains(string(data), "…") {
		t.Fatal("long value not truncated")
	}
}

func TestAgentsRefCombining(t *testing.T) {
	stack, eng := newInst(t)
	path := filepath.Join(eng.Workdir, "AGENTS.md")
	// A foreign file is combined under the machine contract.
	if err := stack.fs.WriteFile(path, []byte("repo rules\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ensureAgentsRef(eng.Workdir)
	data, _ := stack.fs.ReadFile(path)
	s := string(data)
	if !strings.HasPrefix(s, refMarker) || !strings.Contains(s, "repo rules") || !strings.Contains(s, tailBoundary) {
		t.Fatalf("combined ref:\n%.200s", s)
	}
	// A machine file regenerates its head, the tail preserved.
	fresh := ensureAgentsRef
	_ = fresh
	ensureAgentsRef(eng.Workdir)
	again, _ := stack.fs.ReadFile(path)
	if string(again) != s {
		t.Fatal("idempotent regeneration changed bytes")
	}
}

func TestReferenceLibrary(t *testing.T) {
	stack, eng := newInst(t)
	dir := filepath.Join(eng.Store.Root(), "reference")
	for _, d := range []string{"environment.md", "grammar.md", "model.md"} {
		if _, err := stack.fs.ReadFile(filepath.Join(dir, d)); err != nil {
			t.Fatalf("%s: %v", d, err)
		}
	}
	// A stale machine file is rewritten (the canary word is chosen
	// outside the canon vocabulary).
	if err := stack.fs.WriteFile(filepath.Join(dir, "model.md"), []byte("canary-junk"), 0o644); err != nil {
		t.Fatal(err)
	}
	ensureReference(eng.Workdir)
	data, _ := stack.fs.ReadFile(filepath.Join(dir, "model.md"))
	if strings.Contains(string(data), "canary-junk") {
		t.Fatal("stale reference not rewritten")
	}
}

func TestSummaryHumanFileUntouched(t *testing.T) {
	stack, eng := newInst(t)
	path := filepath.Join(eng.Store.Root(), "SUMMARY.md")
	if err := stack.fs.WriteFile(path, []byte("human words"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeSummaryRender(eng)
	data, _ := stack.fs.ReadFile(path)
	if string(data) != "human words" {
		t.Fatal("human summary overwritten")
	}
	// The machine render regenerates its own.
	if err := stack.fs.Remove(path); err != nil {
		t.Fatal(err)
	}
	submit(t, stack, eng, specDoc("B1", "hello"))
	submit(t, stack, eng, "kind: run\n")
	submit(t, stack, eng, "kind: run\nratify: true\n")
	data, err := stack.fs.ReadFile(path)
	if err != nil {
		t.Fatalf("summary not written: %v", err)
	}
	if !strings.Contains(string(data), "a tool that greps lines in files") {
		t.Fatalf("summary brief:\n%s", string(data)[:200])
	}
}

// --- Begin ---

func TestBeginRefusals(t *testing.T) {
	stack := bindUnits(t)
	eng, err := Init(stack.dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, code := eng.Begin("   "); code != Rejected {
		t.Fatal("empty brief accepted")
	}
	if _, code := eng.Begin("first brief"); code != Accepted {
		t.Fatal("first brief refused")
	}
	if _, code := eng.Begin("second brief"); code != Rejected {
		t.Fatal("second brief accepted")
	}
	st, _ := eng.Store.State()
	if st.Wish == nil || *st.Wish != "first brief" || st.Stage != canon.StageIntake {
		t.Fatalf("state after begin: %+v", st)
	}
}

// --- writer discipline ---

func TestEffectWhitelist(t *testing.T) {
	if err := checkEffectPath(".punchtape/canon/atoms/b1.yaml"); err != nil {
		t.Fatalf("atoms effect: %v", err)
	}
	if err := checkEffectPath(".punchtape/state.yaml"); err != nil {
		t.Fatalf("state effect: %v", err)
	}
	if err := checkEffectPath("plain-project-file.txt"); err != nil {
		t.Fatalf("project file: %v", err)
	}
	if err := checkEffectPath(".punchtape/cache/x"); err == nil {
		t.Fatal("cache effect allowed")
	}
	if err := checkEffectPath(".punchtape/canon/atoms/Bad Name.yaml"); err == nil {
		t.Fatal("bad charset allowed")
	}
}

func TestIntegrityRefusalOnCorruption(t *testing.T) {
	stack, eng := newInst(t)
	submit(t, stack, eng, specDoc("B1", "hello"))
	// A manual canon edit is detected at the next open.
	if err := stack.fs.WriteFile(filepath.Join(eng.Store.Root(), "canon", "atoms", "b1.yaml"), []byte("id: B1\nname: hacked\nprose: p\nrows: []\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Open(stack.dir); err == nil {
		t.Fatal("corrupted instance opened")
	}
}

func TestWriterBusyReply(t *testing.T) {
	stack, eng := newInst(t)
	stack.fs.lockErr = errBusy()
	reply, code := submit(t, stack, eng, specDoc("B1", "hello"))
	if code != Failure || !strings.Contains(reply, "another writer holds") {
		t.Fatalf("busy reply: %d %s", code, reply)
	}
	stack.fs.lockErr = nil
}

// --- stages runner ---

func TestStageGateUnknownPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("unknown gate did not panic")
		}
	}()
	stageGateFn("bogus-gate")
}

func TestRatifyGreenGateSemantics(t *testing.T) {
	stack, eng := newInst(t)
	submit(t, stack, eng, specDoc("B1", "hello"))
	b := newBuilder(eng, mustState(eng))
	if b.ratifyGreen() {
		t.Fatal("ratify gate green without ratify")
	}
	// Green rows without ratifying: still not deliver.
	submit(t, stack, eng, "kind: run\n")
	b2 := newBuilder(eng, mustState(eng))
	b2.load()
	if b2.ratifyGreen() {
		t.Fatal("green without ratify passed the gate")
	}
	// The second lock: a red row under a ratifying flag fails the
	// gate too (the signature predicate refuses such a submission
	// long before the stage machine — this is the belt behind it).
	b3 := newBuilder(eng, mustState(eng))
	b3.load()
	b3.ratifying = true
	for id := range b3.atoms {
		atom := b3.atoms[id]
		atom.Rows[0].Outcome = canon.RowRed
		b3.atoms[id] = atom
	}
	if b3.ratifyGreen() {
		t.Fatal("a red row passed the ratify gate")
	}
	// The full cycle reaches deliver only through a ratifying run.
	submit(t, stack, eng, "kind: run\nratify: true\n")
	if st := mustState(eng); st.Stage != canon.StageDeliver {
		t.Fatalf("ratify did not deliver: %s", st.Stage)
	}
}

// --- helpers used above ---

func errBusy() error {
	return os.ErrDeadlineExceeded
}

var _ = fsport.Bind

func TestBehaviorIDCaseCollision(t *testing.T) {
	stack, eng := newInst(t)
	submit(t, stack, eng, specDoc("B1", "hello"))
	reply, code := submit(t, stack, eng, fr(specDoc("b1", "hello")))
	mustReject(t, reply, code, "collides with another id in lower case")
}

// The self-damage stuck-instance half: the
// The legacy serializer wrote an atom the machine itself could not decode,
// every verb of the instance died on it, and the integrity check
// prescribed a new instance. The machine must HEAL damage it wrote
// itself: the file digests to the journal's own record but does not
// decode — quarantine it to a sidecar, journal the deletion, and let
// the executor resubmit the behavior; the rest of the instance
// survives. A hand-edited file (digest off the journal record) stays
// the honest integrity refusal — healing never launders manual edits.
func TestHealToolBrokenAtom(t *testing.T) {
	stack, eng := newInst(t)
	submit(t, stack, eng, specDoc("B1", "hello"))
	// The exact bytes the legacy serializer emitted for a seed text of
	// one line feed: the literal block `text: |` with an empty body
	// decodes to an empty seed — the atom is unreadable at its form.
	broken := "id: B2\nname: one-line-feed\nprose: prose\nrows:\n- run: cat seed.txt\n" +
		"  seeds:\n  - path: seed.txt\n    text: |\n\n  out: |\n\n  err: null\n  rc: null\n" +
		"  bytes: \"\"\n  timeout: 0\n  outcome: \"\"\n  reason: \"\"\n  env-fail: false\ndigest: \"\"\n"
	entry := journal.Entry{
		Transaction:   yamlio.Digest([]byte("tool-broken-b2")),
		SubmissionKey: "0002-spec",
		DeltaKind:     "spec",
		WallMs:        3,
		RecordedAt:    timeNow().UTC(),
		Effects: []journal.Effect{{
			Path: canonFile("atoms", "B2"), Content: broken,
		}},
	}
	lock, err := eng.acquireWriter()
	if err != nil {
		t.Fatalf("writer: %v", err)
	}
	if err := eng.commit(entry); err != nil {
		eng.releaseWriter(lock)
		t.Fatalf("commit: %v", err)
	}
	eng.releaseWriter(lock)

	// The heal: opening quarantines the tool-written damage (the
	// digests match — the machine wrote both — and the record does not
	// decode back), the healthy atom survives, the broken id is free
	// for a resubmission. Before the cure this same open left every verb of
	// the instance dying on the one broken atom.
	eng2, exists, err := Open(stack.dir)
	if err != nil || !exists {
		t.Fatalf("open: %v exists=%v", err, exists)
	}
	if _, err := eng2.Next(); err != nil {
		t.Fatalf("next after heal: %v", err)
	}
	atoms, err := eng2.Store.Atoms()
	if err != nil || len(atoms) != 1 || atoms[0].ID != "B1" {
		t.Fatalf("corpus after heal: %v %v", atoms, err)
	}
	sidecar := filepath.Join(eng2.Workdir, ".punchtape", "canon", "atoms", "b2.yaml.torn")
	if data, err := fsport.Bind.ReadFile(sidecar); err != nil || !strings.Contains(string(data), "text: |") {
		t.Fatalf("quarantine sidecar: %v", err)
	}
	var healed bool
	for _, en := range eng2.Journal.All() {
		if en.DeltaKind == "heal" {
			healed = true
		}
	}
	if !healed {
		t.Fatal("the heal must be a journaled transaction of its own")
	}
	// The resubmission is a SPEC — the full state: it carries both
	// behaviors (B1 untouched keeps its green, B2 is new).
	both := fr(specDoc("B1", "hello")) + "  - id: B2\n    name: sample\n    prose: |\n      One paragraph.\n    rows:\n      - run: \"echo hello\"\n        out: \"hello\"\n        rc: 0\n        form: \"corpus: the pinned surface\"\n"
	reply, code := submit(t, stack, eng2, both)
	mustAccept(t, reply, code)
	atoms, err = eng2.Store.Atoms()
	if err != nil || len(atoms) != 2 {
		t.Fatalf("corpus after resubmit: %v %v", atoms, err)
	}

	// A hand-edited file never heals: the digest leaves the journal's
	// record and the honest integrity refusal stands.
	hand := filepath.Join(eng2.Workdir, ".punchtape", "canon", "atoms", "b2.yaml")
	if err := fsport.Bind.WriteFile(hand, []byte("id: B2\nname: x\nprose: y\nrows:\n- run: [x]\n"), 0o644); err != nil {
		t.Fatalf("hand edit: %v", err)
	}
	if _, _, err := Open(stack.dir); err == nil {
		t.Fatal("a manual edit must stay the integrity refusal")
	}
}

// The self-damage guard and heal: the write gate's own cases, the gate's
// refusal in the submission flow (through the seam), and the heal
// path's fault branches (the walk, the read, the quarantine write,
// the journal append, the writer lock).
func TestHealAndGateBranches(t *testing.T) {
	// The gate on real bytes: a marshaled atom passes, a broken
	// document refuses, a valid-but-noncanonical spelling refuses.
	ok, err := yamlio.Marshal(canon.Atom{ID: "B2", Name: "n", Prose: "p", Rows: []canon.Row{{Run: canon.Cmd{Argv: []string{"x"}, IsArgv: true}}}})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := atomWriteGate(ok); err != nil {
		t.Fatalf("canonical bytes must pass: %v", err)
	}
	if err := atomWriteGate([]byte("id: [unclosed\n")); err == nil {
		t.Fatal("undecodable bytes must refuse")
	}
	if err := atomWriteGate([]byte("id: B2\nname: n\nprose: p\nrows:\n- run: [x]\n")); err == nil {
		t.Fatal("a noncanonical spelling must refuse")
	}

	// The gate in the flow: a rigged gate refuses the submission, the
	// atom never enters the corpus.
	stack, eng := newInst(t)
	oldGate := atomGate
	atomGate = func([]byte) error { return errors.New("gate refuses") }
	t.Cleanup(func() { atomGate = oldGate })
	reply, code := submit(t, stack, eng, specDoc("B9", "hello"))
	mustReject(t, reply, code, "gate refuses")
	atomGate = oldGate
	if atoms, err := eng.Store.Atoms(); err != nil || len(atoms) != 0 {
		t.Fatalf("corpus after the gate refusal: %v %v", atoms, err)
	}

	// The heal fault branches over a fresh tool-broken instance.
	stack2, eng2 := newInst(t)
	submit(t, stack2, eng2, specDoc("B1", "hello"))
	brokenAtom := "id: B2\nname: one-line-feed\nprose: prose\nrows:\n- run: cat seed.txt\n" +
		"  seeds:\n  - path: seed.txt\n    text: |\n\n  out: |\n\n  err: null\n  rc: null\n" +
		"  bytes: \"\"\n  timeout: 0\n  outcome: \"\"\n  reason: \"\"\n  env-fail: false\ndigest: \"\"\n"
	entry := journal.Entry{
		Transaction:   yamlio.Digest([]byte("broken-again")),
		SubmissionKey: "0009-spec",
		DeltaKind:     "spec",
		WallMs:        3,
		RecordedAt:    timeNow().UTC(),
		Effects:       []journal.Effect{{Path: canonFile("atoms", "B2"), Content: brokenAtom}},
	}
	lock, err := eng2.acquireWriter()
	if err != nil {
		t.Fatalf("writer: %v", err)
	}
	if err := eng2.commit(entry); err != nil {
		eng2.releaseWriter(lock)
		t.Fatalf("commit: %v", err)
	}
	eng2.releaseWriter(lock)
	fresh := func() *Engine {
		store, exists, err := canon.Open(stack2.dir)
		if err != nil || !exists {
			t.Fatalf("canon open: %v %v", err, exists)
		}
		e, err := openEngine(stack2.dir, store)
		if err != nil {
			t.Fatalf("engine: %v", err)
		}
		return e
	}
	atomPath := filepath.Join(stack2.dir, canonFile("atoms", "B2"))
	sidecarPath := filepath.Join(stack2.dir, canonFile("atoms", "B2")+".torn")

	stack2.fs.lockErr = os.ErrDeadlineExceeded
	if err := fresh().healToolBrokenAtoms(); err == nil {
		t.Fatal("a failing lock must refuse the heal")
	}
	stack2.fs.lockErr = nil
	stack2.fs.readdirErr = errIO
	if err := fresh().healToolBrokenAtoms(); err == nil {
		t.Fatal("a failing walk must refuse the heal")
	}
	stack2.fs.readdirErr = nil
	stack2.fs.readErrPath = atomPath
	if err := fresh().healToolBrokenAtoms(); err == nil {
		t.Fatal("a failing read must refuse the heal")
	}
	stack2.fs.readErrPath = ""
	stack2.fs.writeErrPath = sidecarPath
	if err := fresh().healToolBrokenAtoms(); err == nil {
		t.Fatal("a failing quarantine write must refuse the heal")
	}
	stack2.fs.writeErrPath = ""
	stack2.fs.appendFailAt = stack2.fs.appendCalls + 1
	if err := fresh().healToolBrokenAtoms(); err == nil {
		t.Fatal("a failing journal append must refuse the heal")
	}
	stack2.fs.appendFailAt = 0
	// The walk fault through the FULL open: the integrity scan swallows
	// the ReadDir error, the heal surfaces it (openParts refusal).
	stack2.fs.readdirErr = errIO
	if _, _, err := Open(stack2.dir); err == nil {
		t.Fatal("a failing walk must refuse the open")
	}
	stack2.fs.readdirErr = nil
	// The healthy pass after the faults: the heal completes.
	if err := fresh().healToolBrokenAtoms(); err != nil {
		t.Fatalf("heal: %v", err)
	}
	if _, err := fsport.Bind.ReadFile(sidecarPath); err != nil {
		t.Fatalf("sidecar: %v", err)
	}

	// A work root without the canon collection directory: the walk's
	// missing-dir branch is a quiet pass, not a failure.
	bare := memDir(t)
	if err := fsport.Bind.MkdirAll(filepath.Join(bare, ".punchtape")); err != nil {
		t.Fatalf("skeleton: %v", err)
	}
	st := canon.State{SchemaVersion: canon.SchemaVersion, Stage: canon.StageIntake, Counters: map[string]int{"CKP": 0}}
	stData, _ := yamlio.Marshal(st)
	if err := fsport.Bind.WriteAtomic(filepath.Join(bare, ".punchtape", "state.yaml"), stData); err != nil {
		t.Fatalf("state: %v", err)
	}
	bstore, exists, err := canon.Open(bare)
	if err != nil || !exists {
		t.Fatalf("bare open: %v %v", err, exists)
	}
	beng, err := openEngine(bare, bstore)
	if err != nil {
		t.Fatalf("bare engine: %v", err)
	}
	if err := beng.healToolBrokenAtoms(); err != nil {
		t.Fatalf("a missing collection directory is a quiet pass: %v", err)
	}
}

// The directness law: stream expectations
// are DIRECT. The echo row (the scenario's own PASS line as the
// expectation), the vacuum block (PASS lines printed by
// constant-true checks), the bare PASS marks — all refused at the
// form; a short contains fragment is refused
// unless it carries the explicit loose mark.
func TestExpectationQualityRefusals(t *testing.T) {
	stack, eng := newInst(t)
	rejectDoc := func(out string) (string, int) {
		doc := "kind: spec\nfull-replace: true\nbehaviors:\n  - id: B1\n    name: sample\n    prose: |\n      One paragraph.\n    rows:\n      - run: \"echo hello\"\n        out: " + out + "\n        rc: 0\n"
		return submit(t, stack, eng, doc)
	}
	// the registry's cpp episode: the scenario's PASS self-report as the expectation
	reply, code := rejectDoc("|\n          PASS twenty puts return twenty distinct standard textual UUIDs")
	mustReject(t, reply, code, "self-report")
	// the registry's js episode: a block of PASS lines (constant-true
	// checks surface as the same shape)
	reply, code = rejectDoc("|\n          PASS client-break-closes-chain\n          PASS nodes-live-on-after-break")
	mustReject(t, reply, code, "self-report")
	// the registry's go episode: the bare PASS-mark line
	reply, code = rejectDoc("PASS-B1")
	mustReject(t, reply, code, "self-report")
	// the same law on the err stream
	doc := "kind: spec\nfull-replace: true\nbehaviors:\n  - id: B1\n    name: sample\n    prose: |\n      One paragraph.\n    rows:\n      - run: \"echo hello\"\n        err: PASS scenario checked itself\n        rc: 0\n"
	reply, code = submit(t, stack, eng, doc)
	mustReject(t, reply, code, "self-report")
	// the loose-contains law: a fragment shorter than the canon limit
	// without the explicit mark
	reply, code = rejectDoc("\"(contains) xy\"")
	mustReject(t, reply, code, "(contains-loose) xy")
	// the marked form is legal, and real fragments pass untouched
	reply, code = rejectDoc("\"(contains-loose) xy\"")
	mustAccept(t, reply, code)
	reply, code = rejectDoc("\"(contains) usage:\"")
	mustAccept(t, reply, code)
	reply, code = rejectDoc("\"password: keeper\"")
	mustAccept(t, reply, code)
}

// The signed-surface law: the run's arguments
// may not name an existing work-tree FILE that sits outside the
// verdict's signed surface (a build-ignore pattern covers it) unless
// the row itself carries the bytes as a seed — a green measured on
// unsigned fixtures is not portable. The executed program (the first
// field) is exempt: it is the product under test, rebuilt from the
// signed sources by the declared build. Seeds and state wants under
// an ignore are legal (the atom carries their bytes); a path that
// does not exist is legal (the product's refusal shapes).
func TestBuildIgnoreRowDependencies(t *testing.T) {
	stack, eng := newInst(t)
	fix := filepath.Join(stack.dir, "fixtures", "white.jpg")
	if err := stack.fs.MkdirAll(filepath.Dir(fix)); err != nil {
		t.Fatal(err)
	}
	if err := stack.fs.WriteFile(fix, []byte("jpeg bytes"), 0o644); err != nil {
		t.Fatal(err)
	}
	// the episode: the argument names an existing ignored file with no
	// seed carrying it
	reply, code := submit(t, stack, eng, "kind: spec\nbuild-ignore: [vendordir, fixtures]\nbehaviors:\n"+
		"  - id: B1\n    name: measure\n    prose: |\n      One paragraph.\n    rows:\n"+
		"      - run: [\"./tool\", \"detect\", \"fixtures/white.jpg\"]\n        out: \"\"\n        rc: 0\n        form: \"corpus: the pinned surface\"\n")
	mustReject(t, reply, code, "fixtures/white.jpg")
	// the bytes carried by the row's own seed: legal
	reply, code = submit(t, stack, eng, "kind: spec\nbuild-ignore: [vendordir, fixtures]\nbehaviors:\n"+
		"  - id: B2\n    name: carried\n    prose: |\n      One paragraph.\n    rows:\n"+
		"      - run: [\"./tool\", \"detect\", \"fixtures/white.jpg\"]\n        seeds:\n"+
		"          - path: fixtures/white.jpg\n            text: jpeg bytes\n        out: \"\"\n        rc: 0\n        form: \"corpus: the pinned surface\"\n")
	mustAccept(t, reply, code)
	// a missing path under the ignore: legal (the product refuses it)
	reply, code = submit(t, stack, eng, "kind: spec\nfull-replace: true\nbuild-ignore: [vendordir, fixtures]\nbehaviors:\n"+
		"  - id: B3\n    name: missing\n    prose: |\n      One paragraph.\n    rows:\n"+
		"      - run: [\"./tool\", \"detect\", \"fixtures/gone.jpg\"]\n        out: \"\"\n        rc: 0\n        form: \"corpus: the pinned surface\"\n")
	mustAccept(t, reply, code)
	// the program field is exempt: an ignored existing binary as the
	// first field is the build's product
	if err := stack.fs.WriteFile(filepath.Join(stack.dir, "tool"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	reply, code = submit(t, stack, eng, "kind: spec\nfull-replace: true\nbuild-ignore: [tool]\nbehaviors:\n"+
		"  - id: B4\n    name: program\n    prose: |\n      One paragraph.\n    rows:\n"+
		"      - run: [\"./tool\", \"detect\"]\n        out: \"\"\n        rc: 0\n        form: \"corpus: the pinned surface\"\n")
	mustAccept(t, reply, code)
	// an ignored DIRECTORY argument is not a file dependency (the
	// row's seeds and the run's outputs own its life)
	nmd := filepath.Join(stack.dir, "vendordir", "marked")
	if err := stack.fs.MkdirAll(nmd); err != nil {
		t.Fatal(err)
	}
	if err := stack.fs.WriteFile(filepath.Join(nmd, "index.js"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	reply, code = submit(t, stack, eng, "kind: spec\nfull-replace: true\nbuild-ignore: [vendordir, fixtures]\nbehaviors:\n"+
		"  - id: B5\n    name: dirarg\n    prose: |\n      One paragraph.\n    rows:\n"+
		"      - run: [\"./tool\", \"build\", \"vendordir\"]\n        out: \"\"\n        rc: 0\n        form: \"corpus: the pinned surface\"\n")
	mustAccept(t, reply, code)
}
