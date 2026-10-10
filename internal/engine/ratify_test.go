// The ratify battery: the cheap ratify — a SIGNATURE, not a
// run. A ratification never executes a row or the build: it signs
// exactly when the recorded green still stands (every row fresh at
// unchanged inputs, the surface digest equal to the last green
// run's); anything stale closes the door with the names. Every test
// names its field episode.
package engine

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/execport"
)

// ratifyCycle — spec v1 with a DECLARED BUILD, a green run (the
// build starts once, the rows measure), the acceptance moment: the
// state a closing signs.
func ratifyCycle(t *testing.T, stack *fakeStack, eng *Engine) {
	t.Helper()
	buildProduct(t, stack, eng)
	setPlan(stack, "s/index.html")
	doc := "kind: spec\nbuild: [true]\nbehaviors:\n" +
		"  - id: LIST\n    name: the run listing\n    prose: |\n      The listing names what the run wrote.\n" +
		"    rows:\n      - run: \"sh -c './tool a'\"\n        out: \"s/index.html\"\n        rc: 0\n        form: \"corpus: the pinned surface\"\n" +
		"  - id: HOLD\n    name: held behavior\n    prose: |\n      Stable rows stay green untouched.\n" +
		"    rows:\n      - run: \"sh -c './tool b'\"\n        out: \"held\"\n        rc: 0\n        form: \"corpus: the pinned surface\"\n"
	reply, code := submit(t, stack, eng, doc)
	mustAccept(t, reply, code)
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "suite=green")
}

// (1) The closing-cost episode (registry: the threshold tool's
// measured machine instance): the final turn cost 451,265 tokens
// just to close the door — the machine re-ran the
// whole ceremony to sign an unchanged green. The clean closing must
// start NOTHING: not one row, not the build; the signature stands on
// the recorded green, and the bookkeeping (checkpoints, acceptance
// moments) does not grow.
func TestCleanClosingRunsNothing(t *testing.T) {
	stack, eng := newInst(t)
	ratifyCycle(t, stack, eng)
	ckpBefore := countCheckpoints(t, stack, eng)
	accBefore := countAcceptances(eng)
	stack.proc.starts = nil
	reply, code := submit(t, stack, eng, "kind: run\nratify: true\n")
	mustAccept(t, reply, code, "VERDICT: READY")
	if got := len(stack.proc.starts); got != 0 {
		t.Fatalf("the clean closing started %d commands, want 0 (not a row, not the build)", got)
	}
	if !strings.Contains(reply, "kept") || !strings.Contains(reply, "ratified") {
		t.Fatalf("the reply does not name the signature:\n%s", reply)
	}
	if got := countCheckpoints(t, stack, eng); got != ckpBefore {
		t.Fatalf("the closing wrote %d new checkpoints, want 0", got-ckpBefore)
	}
	if got := countAcceptances(eng); got != accBefore {
		t.Fatalf("the closing recorded %d new acceptance moments, want 0", got-accBefore)
	}
	if v := readCellFile(t, eng, "verdicts/v1.txt"); !strings.Contains(v, "VERDICT: READY") {
		t.Fatalf("verdict v1:\n%s", v)
	}
}

// (2) The one-line-fix round episode (registry: the listing
// tool's measured machine instances): the fix arrives as
// the next spec version; a ratify without a run must REFUSE with the
// stale row named (the old machine silently re-measured it inside
// the closing); the honest route is a plain run first — and after it
// the closing starts nothing.
func TestStaleRowRefusesWithNames(t *testing.T) {
	stack, eng := newInst(t)
	ratifyCycle(t, stack, eng)
	setPlan(stack, "index.html")
	reply, code := submit(t, stack, eng, fr(stateSpec("index.html", "held", "")))
	mustAccept(t, reply, code, "SPEC v2")
	stack.proc.starts = nil
	reply, code = submit(t, stack, eng, "kind: run\nratify: true\n")
	mustReject(t, reply, code, "LIST/1", "kind: run")
	if got := len(stack.proc.starts); got != 0 {
		t.Fatalf("the refused ratify started %d commands, want 0", got)
	}
	st, err := eng.Store.State()
	if err != nil || st.Stage != "implement" {
		t.Fatalf("the refused ratify moved the stage: %s %v", st.Stage, err)
	}
	// The honest route: the plain run re-measures only the stale row.
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "suite=green")
	if got := countStarts(stack, "./tool a"); got != 1 {
		t.Fatalf("the round re-ran the listing row %d times, want 1", got)
	}
	stack.proc.starts = nil
	reply, code = submit(t, stack, eng, "kind: run\nratify: true\n")
	mustAccept(t, reply, code, "VERDICT: READY")
	if got := len(stack.proc.starts); got != 0 {
		t.Fatalf("the round's closing started %d commands, want 0", got)
	}
	if v := readCellFile(t, eng, "verdicts/v1.txt"); !strings.Contains(v, "version: v1") {
		t.Fatalf("verdict v1:\n%s", v)
	}
}

// (3) A tree that moved where no row looks (a file no command names)
// still breaks the signature's ground: the digest of the surface no
// longer equals the digest of the last green run, and the ratify
// refuses naming the moved path — the old machine signed a drifted
// tree as READY.
func TestTreeDriftRefusesWithPaths(t *testing.T) {
	stack, eng := newInst(t)
	ratifyCycle(t, stack, eng)
	_ = stack.fs.WriteFile(filepath.Join(eng.Workdir, "notes.txt"), []byte("touched\n"), 0o644)
	stack.proc.starts = nil
	reply, code := submit(t, stack, eng, "kind: run\nratify: true\n")
	mustReject(t, reply, code, "notes.txt", "kind: run")
	if got := len(stack.proc.starts); got != 0 {
		t.Fatalf("the refused ratify started %d commands, want 0", got)
	}
}

// (4) The ratify-on-v1 episode (registry: both listing-tool
// machine instances): the executors ratified straight after the spec — the run and the closing were
// one paid ceremony. The door now teaches the law first: no green
// run — no signature; every never-ran row is named.
func TestNoGreenRunRefuses(t *testing.T) {
	stack, eng := newInst(t)
	buildProduct(t, stack, eng)
	setPlan(stack, "s/index.html")
	reply, code := submit(t, stack, eng, stateSpec("s/index.html", "held", ""))
	mustAccept(t, reply, code)
	stack.proc.starts = nil
	reply, code = submit(t, stack, eng, "kind: run\nratify: true\n")
	mustReject(t, reply, code, "LIST/1", "HOLD/1", "kind: run")
	if got := len(stack.proc.starts); got != 0 {
		t.Fatalf("the refused ratify started %d commands, want 0", got)
	}
	// The taught route: run green, then the closing signs for free.
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "suite=green")
	stack.proc.starts = nil
	reply, code = submit(t, stack, eng, "kind: run\nratify: true\n")
	mustAccept(t, reply, code, "VERDICT: READY")
	if got := len(stack.proc.starts); got != 0 {
		t.Fatalf("the closing started %d commands, want 0", got)
	}
}

// countCheckpoints — the stored snapshot records (the closing adds
// none: its observation is not a run's checkpoint).
func countCheckpoints(t *testing.T, stack *fakeStack, eng *Engine) int {
	t.Helper()
	n := 0
	for p := range stack.fs.files {
		if strings.HasPrefix(p, filepath.Join(eng.Store.Root(), "checkpoints")+"/") {
			n++
		}
	}
	return n
}

func countAcceptances(eng *Engine) int {
	n := 0
	for _, en := range eng.Ledger.All() {
		if en.Event == "gate" && en.Details["name"] == "suite" && en.Details["scope"] == "acceptance" {
			n++
		}
	}
	return n
}

// The stale causes are named per row: a red row re-measures, a row
// whose named input moved is stale by its inputs. The budget keeps
// the refusal readable at scale.
func TestStaleCausesAndBudget(t *testing.T) {
	// A red row at the ratify: named with its cause.
	stack, eng := newInst(t)
	buildProduct(t, stack, eng)
	setPlan(stack, "s/index.html")
	reply, code := submit(t, stack, eng, stateSpec("WRONG", "held", ""))
	mustAccept(t, reply, code)
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "suite=red")
	reply, code = submit(t, stack, eng, "kind: run\nratify: true\n")
	mustReject(t, reply, code, "LIST/1", "red")

	// A named-input row: the input's bytes move, the row is stale by
	// its inputs (a sh -c wrapper names nothing — no other row moves).
	stack2, eng2 := newInst(t)
	buildProduct(t, stack2, eng2)
	_ = stack2.fs.WriteFile(filepath.Join(eng2.Workdir, "named.txt"), []byte("one\n"), 0o644)
	stack2.proc.plan["cat named.txt"] = execport.Outcome{Exit: 0, Stdout: "one\n"}
	doc := "kind: spec\nbehaviors:\n  - id: NAMED\n    name: named input\n    prose: |\n      p\n" +
		"    rows:\n      - run: \"cat named.txt\"\n        out: \"one\"\n        rc: 0\n        form: \"corpus: the pinned surface\"\n"
	reply, code = submit(t, stack2, eng2, doc)
	mustAccept(t, reply, code)
	reply, code = submit(t, stack2, eng2, "kind: run\n")
	mustAccept(t, reply, code, "suite=green")
	stack2.proc.plan["cat named.txt"] = execport.Outcome{Exit: 0, Stdout: "two\n"}
	_ = stack2.fs.WriteFile(filepath.Join(eng2.Workdir, "named.txt"), []byte("two\n"), 0o644)
	reply, code = submit(t, stack2, eng2, "kind: run\nratify: true\n")
	mustReject(t, reply, code, "NAMED/1", "inputs")

	// Twelve never-ran rows: the refusal caps at the budget with the
	// honest "+N more" tail.
	stack3, eng3 := newInst(t)
	var sb strings.Builder
	sb.WriteString("kind: spec\nbehaviors:\n")
	for i := 0; i < 12; i++ {
		id := "R" + string(rune('A'+i))
		fmt.Fprintf(&sb, "  - id: %s\n    name: n\n    prose: |\n      p\n    rows:\n      - run: \"sh -c './tool %s'\"\n        out: \"x\"\n        rc: 0\n        form: \"corpus: the pinned surface\"\n", id, id)
	}
	reply, code = submit(t, stack3, eng3, sb.String())
	mustAccept(t, reply, code)
	reply, code = submit(t, stack3, eng3, "kind: run\nratify: true\n")
	mustReject(t, reply, code, "12 row", "(+2 more)")
}

// The predicate's honest edges: a system fault (the surface cannot
// be observed) is a failure, not a refusal; the digest fallback when
// the stored checkpoint of the acceptance digest is gone or corrupt;
// a removed file names its path too; the no-green branch of a
// hand-built corpus.
func TestPredicateEdges(t *testing.T) {
	stack, eng := newInst(t)
	ratifyCycle(t, stack, eng)
	// The surface cannot be observed: an honest failure.
	stack.fs.walkErr = errIO
	reply, code := submit(t, stack, eng, "kind: run\nratify: true\n")
	if code != Failure {
		t.Fatalf("the unobservable surface must fail, got %d: %s", code, reply)
	}
	stack.fs.walkErr = nil
	// A REMOVED file is a moved path like any other.
	delete(stack.fs.files, filepath.Join(eng.Workdir, "tool"))
	reply, code = submit(t, stack, eng, "kind: run\nratify: true\n")
	mustReject(t, reply, code, "tool")
	// The stored checkpoints vanish: the drift still refuses, naming
	// the digests alone (the fallback).
	_ = stack.fs.WriteFile(filepath.Join(eng.Workdir, "post.txt"), []byte("x"), 0o644)
	for p := range stack.fs.files {
		if strings.HasPrefix(p, filepath.Join(eng.Store.Root(), "checkpoints")+"/") {
			delete(stack.fs.files, p)
		}
	}
	reply, code = submit(t, stack, eng, "kind: run\nratify: true\n")
	mustReject(t, reply, code, "since the last green run")
	if strings.Contains(reply, "post.txt") {
		t.Fatalf("the fallback names paths it cannot know:\n%s", reply)
	}
	// A corrupt checkpoint record is skipped on the walk to the
	// acceptance digest (the fallback again).
	ckpPath := filepath.Join(eng.Store.Root(), "checkpoints", "ckp-000001.yaml")
	delete(stack.fs.files, filepath.Join(eng.Workdir, "post.txt"))
	_ = stack.fs.WriteFile(ckpPath, []byte("%%not yaml at all\n"), 0o644)
	reply, code = submit(t, stack, eng, "kind: run\nratify: true\n")
	mustReject(t, reply, code, "since the last green run")
	// The no-green branch: a fresh-row corpus with no acceptance
	// behind it (a hand-built builder — the door always needs a run
	// first, this is the belt behind the suspenders).
	stack4, eng4 := newInst(t)
	b := newBuilder(eng4, mustState(eng4))
	row := canon.Row{Run: canon.Cmd{Line: "true"}, Outcome: canon.RowGreen,
		RanAt: "2000-01-01T00:00:00Z"}
	row.Inputs = b.rowInputTokens(row)
	b.atoms["B1"] = canon.Atom{ID: "B1", Name: "n", Prose: "p", Rows: []canon.Row{row}}
	if err := b.ratifyPredicate(); !isRejection(err) || !strings.Contains(err.Error(), "no green run") {
		t.Fatalf("no-green branch: %v", err)
	}
	// checkpointOfDigest under a state read fault: nil, honestly.
	knobRead(stack4, filepath.Join(eng4.Store.Root(), "state.yaml"), func() {
		if got := eng4.checkpointOfDigest("x"); got != nil {
			t.Fatalf("state fault digest walk: %+v", got)
		}
	})
}
