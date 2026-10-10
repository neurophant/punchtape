// The spec-as-state battery: the model — a living
// instance, repeat submission as the next spec version, versioned
// verdict snapshots, machine-computed version diffs with visible
// retired frames, and content-addressed memoization of row results.
// Every test names its field episode; all of them fail on the
// rebirth-era code (the instance closed by its one ratification).
package engine

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/execport"
	"github.com/neurophant/punchtape/internal/fsport"
	"github.com/neurophant/punchtape/internal/journal"
)

// buildProduct — a one-script product tree: the tool answers per the
// plan (the fake port), the rows drive it through an sh -c wrapper
// exactly like the measured executors (no literal product path in argv).
func buildProduct(t *testing.T, stack *fakeStack, eng *Engine) {
	t.Helper()
	_ = stack.fs.WriteFile(filepath.Join(eng.Workdir, "tool"), []byte("#!/bin/sh\necho ok\n"), 0o755)
}

// stateSpec — a three-atom spec: LIST (the listing row, changed in the
// round), HOLD (rows that must never re-run) and REF (refusal rows
// that must never re-run either — the refusal-cost episode).
func stateSpec(listOut, holdOut, refOut string) string {
	return "kind: spec\nbehaviors:\n" +
		"  - id: LIST\n    name: the run listing\n    prose: |\n      The listing names what the run wrote.\n" +
		"    rows:\n      - run: \"sh -c './tool a'\"\n        out: \"" + listOut + "\"\n        rc: 0\n        form: \"corpus: the listing surface\"\n" +
		"  - id: HOLD\n    name: held behavior\n    prose: |\n      Stable rows stay green untouched.\n" +
		"    rows:\n      - run: \"sh -c './tool b'\"\n        out: \"" + holdOut + "\"\n        rc: 0\n        form: \"corpus: the held surface\"\n" +
		"  - id: REF\n    name: refusal rows\n    prose: |\n      Refusals answer and exit non-zero.\n" +
		"    rows:\n      - run: \"sh -c './tool --bad'\"\n        out: \"\"\n        rc: 2\n        form: \"corpus: the refusal surface\"\n"
}

func setPlan(stack *fakeStack, listOut string) {
	stack.proc.plan["sh -c ./tool a"] = execport.Outcome{Exit: 0, Stdout: listOut + "\n"}
	stack.proc.plan["sh -c ./tool b"] = execport.Outcome{Exit: 0, Stdout: "held\n"}
	stack.proc.plan["sh -c ./tool --bad"] = execport.Outcome{Exit: 2}
}

// countStartsExact — how many times a command line started, exact join.
func countStartsExact(stack *fakeStack, line string) int {
	n := 0
	for _, argv := range stack.proc.starts {
		if strings.Join(argv, " ") == line {
			n++
		}
	}
	return n
}

// countStarts — how many times a command line started.
func countStarts(stack *fakeStack, marker string) int {
	n := 0
	for _, argv := range stack.proc.starts {
		if strings.Contains(strings.Join(argv, " "), marker) {
			n++
		}
	}
	return n
}

// greenCycle — spec v1, green run, ratification: verdict v1 READY.
func greenCycle(t *testing.T, stack *fakeStack, eng *Engine) {
	t.Helper()
	buildProduct(t, stack, eng)
	setPlan(stack, "s/index.html")
	reply, code := submit(t, stack, eng, stateSpec("s/index.html", "held", ""))
	mustAccept(t, reply, code)
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "suite=green")
	reply, code = submit(t, stack, eng, "kind: run\nratify: true\n")
	mustAccept(t, reply, code, "VERDICT: READY")
	// the hand's follow-up next (the ack is thin)
	if _, err := eng.Next(); err != nil {
		t.Fatal(err)
	}
}

// (a) The one-line-fix round through ONE living instance: the v1
// instance was closed by its ratification — "this blocks the door
// for submissions" — and the round was forced to rebirth (2.78M
// tokens for a one-line fix). The same door must accept the next
// version; the journal must carry BOTH verdicts of the accepted
// states.
func TestLivingInstanceRoundCarriesBothVerdicts(t *testing.T) {
	stack, eng := newInst(t)
	greenCycle(t, stack, eng)
	// The one-line fix: the product and the listing expectation move.
	stack.fs.files[filepath.Join(eng.Workdir, "tool")] = []byte("#!/bin/sh\necho index.html\n")
	setPlan(stack, "index.html")
	// The round: the SAME instance, the SAME door — spec v2.
	reply, code := submit(t, stack, eng, fr(stateSpec("index.html", "held", "")))
	mustAccept(t, reply, code, "SPEC v2")
	st, err := eng.Store.State()
	if err != nil || st.Stage != canon.StageImplement {
		t.Fatalf("the case did not reopen: stage=%q err=%v", st.Stage, err)
	}
	// The changed row is the only thing that must run; the run
	// measures it, the closing signs the recorded green (the
	// ratify itself starts nothing).
	stack.proc.starts = nil
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "suite=green")
	if got := countStarts(stack, "./tool a"); got != 1 {
		t.Fatalf("the round re-ran the listing row %d times, want 1", got)
	}
	if got := countStarts(stack, "./tool b"); got != 0 {
		t.Fatalf("the round re-ran a held row %d times, want 0", got)
	}
	stack.proc.starts = nil
	reply, code = submit(t, stack, eng, "kind: run\nratify: true\n")
	mustAccept(t, reply, code, "VERDICT: READY")
	if got := len(stack.proc.starts); got != 0 {
		t.Fatalf("the closing started %d commands, want 0", got)
	}
	// The journal carries BOTH verdicts of the accepted states.
	v1 := readCellFile(t, eng, "verdicts/v1.txt")
	v2 := readCellFile(t, eng, "verdicts/v2.txt")
	for _, v := range []string{v1, v2} {
		if !strings.Contains(v, "VERDICT: READY") {
			t.Fatalf("a verdict snapshot is not READY:\n%s", v)
		}
	}
	if !strings.Contains(v2, "version: v2") || !strings.Contains(v1, "version: v1") {
		t.Fatalf("the snapshots do not name their versions:\nv1:\n%s\nv2:\n%s", v1, v2)
	}
	// Two acceptance moments in the ledger.
	accepts := 0
	for _, en := range eng.Ledger.All() {
		if en.Event == "gate" && en.Details["name"] == "suite" && en.Details["scope"] == "acceptance" {
			accepts++
		}
	}
	if accepts != 2 {
		t.Fatalf("acceptance moments: %d, want 2", accepts)
	}
}

// (a) The door reply of the next version names the version and the
// row diff (new/changed/retired).
func TestSpecVersionDiffInReply(t *testing.T) {
	stack, eng := newInst(t)
	greenCycle(t, stack, eng)
	setPlan(stack, "index.html")
	reply, code := submit(t, stack, eng, fr(stateSpec("index.html", "held", "")))
	mustAccept(t, reply, code, "SPEC v2", "~1 changed")
	// the named diff lines ride the next slot's changed list (the
	// ack is the counts line only)
	next, err := eng.Next()
	if err != nil || !strings.Contains(next, "LIST/1: never (was green)") {
		t.Fatalf("the changed row is not named by the slot: %v\n%s", err, next)
	}
}

// (b) The expected-value episode: changing one expected value must not drive
// the refusal rows (or any untouched row) to re-run.
func TestExpectedValueChangeSkipsUntouchedRows(t *testing.T) {
	stack, eng := newInst(t)
	greenCycle(t, stack, eng)
	stack.proc.starts = nil
	setPlan(stack, "index.html")
	reply, code := submit(t, stack, eng, fr(stateSpec("index.html", "held", "")))
	mustAccept(t, reply, code)
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "suite=green")
	if got := countStarts(stack, "./tool --bad"); got != 0 {
		t.Fatalf("a refusal row re-ran %d times after an unrelated expectation change", got)
	}
	if got := countStarts(stack, "./tool b"); got != 0 {
		t.Fatalf("a held row re-ran %d times after an unrelated expectation change", got)
	}
	if got := countStarts(stack, "./tool a"); got != 1 {
		t.Fatalf("the changed row ran %d times, want 1", got)
	}
	// The kept rows are named as kept, not silently skipped.
	if !strings.Contains(reply, "kept") {
		t.Fatalf("the run table does not mark the kept rows:\n%s", reply)
	}
}

// (c) A disappeared row is visible: the door diff names it, the
// journal carries the retired frame, and the next verdict explains
// it — greening by deletion without a trace is impossible.
func TestRetiredRowVisibleAndExplained(t *testing.T) {
	stack, eng := newInst(t)
	greenCycle(t, stack, eng)
	setPlan(stack, "s/index.html")
	// v2 retires one row of REF (the refusal row disappears).
	spec2 := "kind: spec\nbehaviors:\n" +
		"  - id: LIST\n    name: the run listing\n    prose: |\n      The listing names what the run wrote.\n" +
		"    rows:\n      - run: \"sh -c './tool a'\"\n        out: \"s/index.html\"\n        rc: 0\n        form: \"corpus: the listing surface\"\n" +
		"  - id: HOLD\n    name: held behavior\n    prose: |\n      Stable rows stay green untouched.\n" +
		"    rows:\n      - run: \"sh -c './tool b'\"\n        out: \"held\"\n        rc: 0\n        form: \"corpus: the held surface\"\n" +
		"  - id: REF\n    name: refusal rows\n    prose: |\n      Refusals answer and exit non-zero.\n" +
		"    rows:\n      - run: \"sh -c './tool --bad2'\"\n        out: \"bad2\\n\"\n        rc: 2\n        form: \"corpus: the refusal surface\"\n"
	stack.proc.plan["sh -c ./tool --bad2"] = execport.Outcome{Exit: 2, Stdout: "bad2\n"}
	reply, code := submit(t, stack, eng, fr(spec2))
	mustAccept(t, reply, code, "-1 retired")
	next, nerr := eng.Next()
	if nerr != nil || !strings.Contains(next, "REF/1: never (was green)") {
		t.Fatalf("the replaced row is not named by the slot: %v\n%s", nerr, next)
	}
	// The journal carries the retired frame.
	found := retiredFrame(eng, "REF/1")
	if !found {
		t.Fatal("no retired frame in the journal")
	}
	// The new refusal row measures in a plain run; the ratify signs.
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "suite=green")
	reply, code = submit(t, stack, eng, "kind: run\nratify: true\n")
	mustAccept(t, reply, code, "VERDICT: READY")
	v2 := readCellFile(t, eng, "verdicts/v2.txt")
	if !strings.Contains(v2, "retired since v1") || !strings.Contains(v2, "REF/1") {
		t.Fatalf("the verdict does not explain the retired row:\n%s", v2)
	}
}

// (d) An untouched row keeps its green status at unchanged input
// digests — and a touched dependency staleness its row (the memo is
// keyed on content, not on time).
func TestUnchangedRowKeepsGreenAtUnchangedDigests(t *testing.T) {
	stack, eng := newInst(t)
	greenCycle(t, stack, eng)
	// A spec version that changes nothing (fresh content, same rows):
	// the run re-executes nothing, every green stands.
	spec2 := strings.Replace(stateSpec("s/index.html", "held", ""), "The listing names what the run wrote.",
		"The listing names what the run wrote (v2 wording).", 1)
	reply, code := submit(t, stack, eng, fr(spec2))
	mustAccept(t, reply, code)
	stack.proc.starts = nil
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "suite=green")
	if got := countStarts(stack, "./tool"); got != 0 {
		t.Fatalf("a nothing-changed version re-ran %d rows, want 0", got)
	}
	atoms, _ := eng.Store.Atoms()
	for _, a := range atoms {
		if !a.Green() {
			t.Fatalf("atom %s lost its green without a change", a.ID)
		}
	}
	// A row that NAMES its input file (a literal argv path) goes
	// stale when that file's bytes move — the memo is content-honest.
	_ = stack.fs.WriteFile(filepath.Join(eng.Workdir, "named.txt"), []byte("one\n"), 0o644)
	stack.proc.plan["cat named.txt"] = execport.Outcome{Exit: 0, Stdout: "one\n"}
	spec3 := spec2 + "  - id: NAMED\n    name: named input\n    prose: |\n      A row with a named file input.\n" +
		"    rows:\n      - run: \"cat named.txt\"\n        out: \"one\"\n        rc: 0\n        form: \"corpus: the pinned surface\"\n"
	reply, code = submit(t, stack, eng, fr(spec3))
	mustAccept(t, reply, code)
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "suite=green")
	stack.proc.starts = nil
	// The named input's bytes move: only NAMED/1 may re-run.
	stack.proc.plan["cat named.txt"] = execport.Outcome{Exit: 0, Stdout: "two\n"}
	_ = stack.fs.WriteFile(filepath.Join(eng.Workdir, "named.txt"), []byte("two\n"), 0o644)
	spec4 := strings.Replace(spec3, "out: \"one\"", "out: \"two\"", 1)
	reply, code = submit(t, stack, eng, fr(spec4))
	mustAccept(t, reply, code)
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "suite=green")
	if got := countStarts(stack, "cat named.txt"); got != 1 {
		t.Fatalf("the named-input row re-ran %d times after its input moved, want 1", got)
	}
	if got := countStarts(stack, "./tool"); got != 0 {
		t.Fatalf("untouched rows re-ran %d times after an unrelated input moved", got)
	}
}

// The door law at deliver: a RUN does not reopen the case — only the
// next spec version does; the refusal names the legal way forward.
func TestRunAtDeliverRefused(t *testing.T) {
	stack, eng := newInst(t)
	greenCycle(t, stack, eng)
	reply, code := submit(t, stack, eng, "kind: run\n")
	mustReject(t, reply, code, "kind: spec")
}

// retiredFrame — whether the journal carries a retired frame for
// the label. RED-RUN NOTE: on the rebirth-era code the frame API does
// not exist; the red run probes the journal bytes textually (no
// frame — false, as expected); the fix swaps the probe for the typed
// walk over Entry.Retired.
func retiredFrame(eng *Engine, label string) bool {
	data, err := fsport.Bind.ReadFile(filepath.Join(eng.Workdir, ".punchtape", "journal.yamll"))
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "retired") && strings.Contains(string(data), label)
}

func readCellFile(t *testing.T, eng *Engine, rel string) string {
	t.Helper()
	data, err := fsport.Bind.ReadFile(filepath.Join(eng.Store.Root(), filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(data)
}

// The machine's row records are never submission content: a forged
// green (or a forged freshness key) is refused at the door.
func TestMachineFieldsRefused(t *testing.T) {
	stack, eng := newInst(t)
	buildProduct(t, stack, eng)
	setPlan(stack, "s/index.html")
	doc := "kind: spec\nbehaviors:\n  - id: LIST\n    name: l\n    prose: |\n      p\n" +
		"    rows:\n      - run: \"echo hello\"\n        out: \"hello\"\n        rc: 0\n        form: \"corpus: the pinned surface\"\n        outcome: green\n"
	reply, code := submit(t, stack, eng, doc)
	mustReject(t, reply, code, "machine-recorded fields")
	doc = strings.Replace(doc, "outcome: green", "ran-at: \"2000-01-01T00:00:00Z\"\n        ran-tree: abc\n        inputs: x:y", 1)
	reply, code = submit(t, stack, eng, doc)
	mustReject(t, reply, code, "machine-recorded fields")
}

// A whole ATOM may retire too (the document is the full state): the
// diff names it, the frame carries its rows, the verdict explains.
func TestRetiredAtomVisible(t *testing.T) {
	stack, eng := newInst(t)
	greenCycle(t, stack, eng)
	setPlan(stack, "s/index.html")
	spec2 := strings.Replace(stateSpec("s/index.html", "held", ""),
		"  - id: REF\n    name: refusal rows\n    prose: |\n      Refusals answer and exit non-zero.\n"+
			"    rows:\n      - run: \"sh -c './tool --bad'\"\n        out: \"\"\n        rc: 2\n        form: \"corpus: the refusal surface\"\n", "", 1)
	reply, code := submit(t, stack, eng, fr(spec2))
	mustAccept(t, reply, code, "SPEC v2", "-1 retired")
	next, nerr := eng.Next()
	if nerr != nil || !strings.Contains(next, "REF/1: retired") {
		t.Fatalf("the retired row is not named by the slot: %v\n%s", nerr, next)
	}
	if !retiredFrame(eng, "REF/1") {
		t.Fatal("no retired frame for the atom's row")
	}
	// The canon file is gone (a delete effect).
	if _, err := eng.Store.Atoms(); err != nil {
		t.Fatal(err)
	}
	for _, a := range mustAtoms(t, eng) {
		if a.ID == "REF" {
			t.Fatal("the retired atom still lives in the corpus")
		}
	}
	reply, code = submit(t, stack, eng, "kind: run\nratify: true\n")
	mustAccept(t, reply, code, "VERDICT: READY")
	if v := readCellFile(t, eng, "verdicts/v2.txt"); !strings.Contains(v, "REF/1") {
		t.Fatalf("verdict v2 does not explain the retired atom's row:\n%s", v)
	}
}

// The memo's input tokens bind by content: from-input seeds and byte
// samples are part of the key (a moved input staleness its row; a
// vanished one records "gone" honestly).
func TestMemoBindsInputsAndSamples(t *testing.T) {
	stack, eng := newInst(t)
	buildProduct(t, stack, eng)
	// the instance's input directory with a seed source
	inDir := filepath.Join(eng.Workdir, "..", "input")
	_ = stack.fs.WriteFile(filepath.Join(inDir, "given.txt"), []byte("one\n"), 0o644)
	stack.proc.plan["cat seeded.txt"] = execport.Outcome{Exit: 0, Stdout: "one\n"}
	doc := "kind: spec\nbehaviors:\n  - id: IN\n    name: input-bound\n    prose: |\n      p\n" +
		"    rows:\n      - run: \"cat seeded.txt\"\n        seeds:\n          - path: seeded.txt\n            from-input: given.txt\n" +
		"        out: \"one\"\n        rc: 0\n        form: \"corpus: the pinned surface\"\n"
	reply, code := submit(t, stack, eng, doc)
	mustAccept(t, reply, code)
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "suite=green")
	// The input's bytes move: the row goes stale and re-runs.
	_ = stack.fs.WriteFile(filepath.Join(inDir, "given.txt"), []byte("two\n"), 0o644)
	stack.proc.plan["cat seeded.txt"] = execport.Outcome{Exit: 0, Stdout: "two\n"}
	spec2 := strings.Replace(doc, "out: \"one\"", "out: \"two\"", 1)
	reply, code = submit(t, stack, eng, fr(spec2))
	mustAccept(t, reply, code)
	stack.proc.starts = nil
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "suite=green")
	if got := countStarts(stack, "cat seeded.txt"); got != 1 {
		t.Fatalf("the moved input's row re-ran %d times, want 1", got)
	}
	// A TEXT seed rides inside the definition (no external token) and
	// a "./" argv field names nothing: both walk the token path
	// honestly without staleness.
	stack.proc.plan["cat text-seeded.txt"] = execport.Outcome{Exit: 0, Stdout: "seeded\n"}
	doc2 := "kind: spec\nfull-replace: true\nbehaviors:\n  - id: TX\n    name: text seed\n    prose: |\n      p\n" +
		"    rows:\n      - run: \"cat text-seeded.txt ./\"\n        seeds:\n          - path: text-seeded.txt\n            text: seeded\n" +
		"        err: \"(contains) text-seeded.txt\"\n        rc: 1\n        form: \"corpus: the refusal surface\"\n"
	reply, code = submit(t, stack, eng, doc2)
	mustAccept(t, reply, code)
	reply, code = submit(t, stack, eng, "kind: run\n")
	if code != Accepted {
		t.Fatalf("the text-seed row must answer (red ok): %d %s", code, reply)
	}
	// The input VANISHES between the spec and the run: the memo
	// records the honest "gone" and the row runs on the world as is.
	delete(stack.fs.files, filepath.Join(inDir, "given.txt"))
	spec3 := strings.Replace(spec2, "from-input: given.txt", "from-input: gone.txt", 1)
	spec3 = strings.Replace(spec3, "out: \"two\"", "out: \"one\"", 1)
	_ = stack.fs.WriteFile(filepath.Join(inDir, "gone.txt"), []byte("x\n"), 0o644)
	// plant the file again under the new name, then remove it post-spec
	reply, code = submit(t, stack, eng, fr(spec3))
	mustAccept(t, reply, code)
	delete(stack.fs.files, filepath.Join(inDir, "gone.txt"))
	stack.proc.plan["cat seeded.txt"] = execport.Outcome{Exit: 0, Stdout: "one\n"}
	reply, code = submit(t, stack, eng, "kind: run\n")
	if code != Accepted {
		t.Fatalf("the vanished-input run must answer honestly (env red), got %d: %s", code, reply)
	}
}

// A byte-sample row's memo carries the sample's digest: a replaced
// sample staleness its row.
func TestMemoBindsSample(t *testing.T) {
	stack, eng := newInst(t)
	buildProduct(t, stack, eng)
	inbox := filepath.Join(eng.Workdir, ".punchtape", "inbox")
	_ = stack.fs.WriteFile(filepath.Join(inbox, "ref.bin"), []byte("hello\n"), 0o644)
	doc := "kind: spec\nbehaviors:\n  - id: S\n    name: sample-bound\n    prose: |\n      p\n" +
		"    rows:\n      - run: \"echo hello\"\n        bytes: file://ref.bin\n"
	reply, code := submit(t, stack, eng, doc)
	mustAccept(t, reply, code)
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "suite=green")
	_ = stack.fs.WriteFile(filepath.Join(inbox, "ref.bin"), []byte("two\n"), 0o644)
	stack.proc.starts = nil
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code)
	if got := countStarts(stack, "echo hello"); got != 1 {
		t.Fatalf("the moved sample's row re-ran %d times, want 1", got)
	}
	// The sample VANISHES: the memo records "gone", the row runs and
	// answers its honest red (sample missing).
	delete(stack.fs.files, filepath.Join(inbox, "ref.bin"))
	stack.proc.starts = nil
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code)
	if got := countStarts(stack, "echo hello"); got != 1 {
		t.Fatalf("the vanished sample's row re-ran %d times, want 1", got)
	}
}

// A kept atom may GROW: the new row beyond the old tail is new work
// (and the atom's definition changed with it).
func TestKeptAtomGrows(t *testing.T) {
	stack, eng := newInst(t)
	buildProduct(t, stack, eng)
	setPlan(stack, "s/index.html")
	reply, code := submit(t, stack, eng, fr(stateSpec("s/index.html", "held", "")))
	mustAccept(t, reply, code)
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "suite=green")
	holdRow := "    rows:\n      - run: \"sh -c './tool b'\"\n        out: \"held\"\n        rc: 0\n        form: \"corpus: the held surface\"\n"
	holdRow2 := "    rows:\n      - run: \"sh -c './tool b'\"\n        out: \"held\"\n        rc: 0\n        form: \"corpus: the held surface\"\n" +
		"      - run: \"sh -c './tool b2'\"\n        out: \"held2\"\n        rc: 0\n        form: \"corpus: the pinned surface\"\n"
	spec2 := strings.Replace(stateSpec("s/index.html", "held", ""), holdRow, holdRow2, 1)
	stack.proc.plan["sh -c ./tool b2"] = execport.Outcome{Exit: 0, Stdout: "held2\n"}
	reply, code = submit(t, stack, eng, fr(spec2))
	mustAccept(t, reply, code, "~1 changed", "+1 new")
	stack.proc.starts = nil
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "suite=green")
	if got := countStarts(stack, "./tool b2"); got != 1 {
		t.Fatalf("the grown row ran %d times, want 1", got)
	}
	if got := countStartsExact(stack, "sh -c ./tool b"); got != 0 {
		t.Fatalf("the old row of the grown atom re-ran %d times, want 0", got)
	}
}

// The suite line names the kept rows when reds and kept greens mix.
func TestSuiteLineMixedKeptAndRed(t *testing.T) {
	stack, eng := newInst(t)
	buildProduct(t, stack, eng)
	setPlan(stack, "s/index.html")
	reply, code := submit(t, stack, eng, fr(stateSpec("s/index.html", "held", "")))
	mustAccept(t, reply, code)
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "suite=green")
	// v2 breaks ONE row's expectation: it re-runs red, the rest keep.
	spec2 := strings.Replace(stateSpec("s/index.html", "held", ""), "out: \"held\"", "out: \"held-changed\"", 1)
	reply, code = submit(t, stack, eng, fr(spec2))
	mustAccept(t, reply, code)
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "suite=red")
	if !strings.Contains(reply, "kept") {
		t.Fatalf("the red suite line does not name the kept rows:\n%s", reply)
	}
}

// The version machinery's honest edges: no snapshot yet, a state
// read fault, and the accounting guard itself.
func TestVersionMachineryEdges(t *testing.T) {
	stack, eng := newInst(t)
	if v, n, ok := eng.latestVerdictSnapshot(); ok || v != nil || n != 0 {
		t.Fatalf("no verdict yet: %+v %d %v", v, n, ok)
	}
	knobRead(stack, filepath.Join(eng.Store.Root(), "state.yaml"), func() {
		if got := eng.verdictVersion(); got != 0 {
			t.Fatalf("state fault version: %d", got)
		}
	})
	// The accounting guard: a text that omits a retired label refuses.
	if err := explainRetired(eng, "nothing here", []journal.RetiredRow{{Label: "B9/2"}}); err == nil {
		t.Fatal("the guard did not refuse an unexplained retirement")
	}
	if err := explainRetired(eng, "B9/2 green", []journal.RetiredRow{{Label: "B9/2", Outcome: "green"}}); err != nil {
		t.Fatalf("the guard refused an explained retirement: %v", err)
	}
	_ = stack
}

// A row retired without ever running (the case has runs) cannot be
// explained: the ratify refuses and names the rows — a check deleted
// before it measured anything hides it.
func TestNeverRanRetirementRefused(t *testing.T) {
	stack, eng := newInst(t)
	setPlan(stack, "s/index.html")
	// v1 verdict first: the case has a run and an accepted state.
	greenCycle(t, stack, eng)
	// v2 ADDS a TEMP atom and retires nothing; no run happens.
	specWithTemp := stateSpec("s/index.html", "held", "") +
		"  - id: TEMP\n    name: never run\n    prose: |\n      Added and dropped before any run.\n" +
		"    rows:\n      - run: \"sh -c './tool t'\"\n        out: \"t\"\n        rc: 0\n        form: \"corpus: the pinned surface\"\n"
	reply, code := submit(t, stack, eng, fr(specWithTemp))
	mustAccept(t, reply, code)
	// v3 retires TEMP entirely — its row NEVER RAN while the case has
	// runs: the ratify must refuse and name it.
	stack.proc.plan["sh -c ./tool t"] = execport.Outcome{Exit: 0, Stdout: "t\n"}
	spec2 := stateSpec("s/index.html", "held", "")
	reply, code = submit(t, stack, eng, fr(spec2))
	// the ack's counts line names the retirement; the never-ran row's
	// own naming is the ratify guard's job (an
	// intermediate version no next observed is not the changed
	// list's business)
	mustAccept(t, reply, code, "SPEC", "-1 retired")
	reply, code = submit(t, stack, eng, "kind: run\nratify: true\n")
	mustReject(t, reply, code, "TEMP/1", "without ever running")
	// The honest way out: restore the row, run it green, retire it
	// again — now its retirement is explainable (it measured).
	reply, code = submit(t, stack, eng, fr(specWithTemp))
	mustAccept(t, reply, code)
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "suite=green")
	reply, code = submit(t, stack, eng, fr(spec2))
	mustAccept(t, reply, code)
	reply, code = submit(t, stack, eng, "kind: run\nratify: true\n")
	mustAccept(t, reply, code, "VERDICT: READY")
	// the hand's follow-up next (the ack is thin)
	if _, err := eng.Next(); err != nil {
		t.Fatal(err)
	}
}

// A row retired BEFORE the case's first run is an edit of the
// hand's own first table, not a hiding: the retire frame records
// its own moment (before-first-run), the never-ran guard amnesties
// it, and the verdict explains it as "outcome —". Four refusals
// to close one pre-run edit cannot happen again. A retire AFTER
// runs opened keeps the guard (TestNeverRanRetirementRefused
// above).
func TestRetireBeforeFirstRunAmnestied(t *testing.T) {
	stack, eng := newInst(t)
	buildProduct(t, stack, eng)
	setPlan(stack, "s/index.html")
	// v1 arrives WITH a TEMP atom; the case has no run yet.
	specWithTemp := stateSpec("s/index.html", "held", "") +
		"  - id: TEMP\n    name: never run\n    prose: |\n      Added and dropped before any run.\n" +
		"    rows:\n      - run: \"sh -c './tool t'\"\n        out: \"t\"\n        rc: 0\n        form: \"corpus: the pinned surface\"\n"
	reply, code := submit(t, stack, eng, fr(specWithTemp))
	mustAccept(t, reply, code)
	// v2 drops TEMP — the row retires BEFORE the first run.
	reply, code = submit(t, stack, eng, fr(stateSpec("s/index.html", "held", "")))
	mustAccept(t, reply, code, "SPEC", "-1 retired")
	// The frame records its moment.
	recorded := false
	for _, en := range eng.Journal.All() {
		for _, r := range en.Retired {
			if r.Label == "TEMP/1" && r.BeforeFirstRun {
				recorded = true
			}
		}
	}
	if !recorded {
		t.Fatal("the retire frame does not record before-first-run")
	}
	// The green run and the ratify: the verdict must NOT refuse —
	// the row never had a chance to measure, its retirement is
	// explainable.
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "suite=green")
	reply, code = submit(t, stack, eng, "kind: run\nratify: true\n")
	mustAccept(t, reply, code, "VERDICT: READY")
	v1 := readCellFile(t, eng, "verdicts/v1.txt")
	if !strings.Contains(v1, "TEMP/1: retired before the first run (outcome —)") {
		t.Fatalf("the verdict does not explain the amnestied row:\n%s", v1)
	}
	if strings.Contains(v1, "never ran") {
		t.Fatalf("the amnestied row is presented as never-ran:\n%s", v1)
	}
}

func mustAtoms(t *testing.T, eng *Engine) []canon.Atom {
	t.Helper()
	atoms, err := eng.Store.Atoms()
	if err != nil {
		t.Fatal(err)
	}
	return atoms
}
