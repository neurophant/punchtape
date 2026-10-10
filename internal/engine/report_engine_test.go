// The report engine battery: the report render's branches
// the CLI cycle cannot reach from above — the verdict spellings, the
// measured-answer forms, the since-walk's boundary folds, the budget
// cap and the honest fault shapes.
package engine

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/execport"
	"github.com/neurophant/punchtape/internal/journal"
)

// The verdict spellings: vN and N name a verdict; anything else does
// not (a leading zero is a different spelling, not a number).
func TestVerdictSpellings(t *testing.T) {
	for _, c := range []struct {
		in   string
		want int
		bad  bool
	}{
		{"2", 2, false}, {"v2", 2, false}, {"v0", 0, false}, {"0", 0, false},
		{"vx", 0, true}, {"v-1", 0, true}, {"v01", 0, true}, {"", 0, true},
		{"2x", 0, true},
	} {
		got, bad := parseVerdictSpelling(c.in)
		if bad != c.bad || got != c.want {
			t.Fatalf("parseVerdictSpelling(%q) = %d,%v want %d,%v", c.in, got, bad, c.want, c.bad)
		}
	}
}

// The measured-answer forms of one row line: a red row answers with
// its measured reason; a memo without the observed rc (a record of an
// older binary) answers with the honest question mark; a never-run
// row says so.
func TestRowLineAnswerForms(t *testing.T) {
	rc := 3
	line := reportLine("R/1", canon.Row{Outcome: canon.RowRed, Reason: "rc: want 0, got 3"}, false)
	if !strings.Contains(line, "rc: want 0, got 3") || !strings.Contains(line, "[red]") {
		t.Fatalf("the red row's answer: %s", line)
	}
	line = reportLine("R/2", canon.Row{Outcome: canon.RowGreen, RanAt: "T", Inputs: "k"}, true)
	if !strings.Contains(line, "rc ?") {
		t.Fatalf("the rc-less memo must say so: %s", line)
	}
	line = reportLine("R/3", canon.Row{Outcome: canon.RowGreen, RanAt: "T", RanRc: &rc, RanOut: "abcd1234"}, true)
	if !strings.Contains(line, "rc 3, out abcd1234") {
		t.Fatalf("the measured answer: %s", line)
	}
	if !strings.Contains(line, "[green, fresh]") {
		t.Fatalf("the fresh standing: %s", line)
	}
	line = reportLine("R/4", canon.Row{}, false)
	if !strings.Contains(line, "never ran") || !strings.Contains(line, "[never]") {
		t.Fatalf("the never-run row: %s", line)
	}
	line = reportLine("R/5", canon.Row{Outcome: canon.RowGreen, RanAt: "T", RanRc: &rc, RanOut: "abcd1234"}, false)
	if !strings.Contains(line, "[green, stale]") {
		t.Fatalf("the stale standing: %s", line)
	}
}

// The full report's honest shapes: an empty case says so; a case
// without a verdict carries the no-verdict orientation line; the
// budget caps a wide corpus; a bad since-spelling is a rejection.
func TestReportFullShapes(t *testing.T) {
	stack, eng := newInst(t)
	text, code := eng.Report("", true)
	if code != Accepted || !strings.Contains(text, "no behaviors yet") {
		t.Fatalf("the empty case report: %d %s", code, text)
	}
	if !strings.Contains(text, "no verdict frozen yet") {
		t.Fatalf("the no-verdict orientation line: %s", text)
	}
	// a wide corpus: 12 rows, the budget caps the table
	greenCycle(t, stack, eng)
	var sb strings.Builder
	sb.WriteString("kind: spec\nfull-replace: true\nbehaviors:\n")
	for i := 0; i < 12; i++ {
		sb.WriteString("  - id: W")
		sb.WriteString(string(rune('A' + i)))
		sb.WriteString("\n    name: wide\n    prose: |\n      p\n" +
			"    rows:\n      - run: \"true\"\n        rc: 0\n")
	}
	reply, rc := submit(t, stack, eng, sb.String())
	mustAccept(t, reply, rc)
	text, code = eng.Report("", true)
	if code != Accepted || !strings.Contains(text, "rows (12):") || !strings.Contains(text, "WL/1 [never]") {
		t.Fatalf("the wide full report: %d %s", code, text)
	}
	// the door view of the same corpus: counters only, no row bodies
	text, code = eng.Report("", false)
	if code != Accepted || strings.Contains(text, "W L") && strings.Contains(text, "run:") {
		t.Fatalf("the door report pays for row bodies: %d %s", code, text)
	}
	// the wide spec retired the green atoms: no greens stand, the
	// counter line honestly does not ride
	if strings.Contains(text, "green rows:") {
		t.Fatalf("the door report counts greens that do not stand: %s", text)
	}
	if reply2, rc2 := eng.Report("vx", false); rc2 != Rejected || !strings.Contains(reply2, "ERROR") {
		t.Fatalf("a bad since-spelling must reject: %d %s", rc2, reply2)
	}
}

// The since-walk: nothing moved after the ratification says so; v0
// names every live row as new; a round that changes one definition,
// retires one row and re-measures nothing else names exactly those;
// the budget caps a wide delta; a corrupted atom frame inside the
// fold is skipped honestly; a retired-then-re-added row is named by
// its CURRENT definition, not its retirement.
func TestSinceWalkEdges(t *testing.T) {
	stack, eng := newInst(t)
	greenCycle(t, stack, eng)
	text, code := eng.Report("v1", true)
	if code != Accepted || !strings.Contains(text, "nothing moved since v1") {
		t.Fatalf("the moved-nothing since-report: %d %s", code, text)
	}
	text, code = eng.Report("v0", true)
	if code != Accepted || !strings.Contains(text, "new since v0") {
		t.Fatalf("the v0 since-report: %d %s", code, text)
	}
	// the round: LIST's definition moves (a new expectation the same
	// plan answers), HOLD retires, REF stands untouched
	stack.proc.plan["sh -c ./tool a"] = execport.Outcome{Exit: 0, Stdout: "x/depth.html\n"}
	round := "kind: spec\nfull-replace: true\nbehaviors:\n" +
		"  - id: LIST\n    name: the run listing\n    prose: |\n      The listing names what the run wrote.\n" +
		"    rows:\n      - run: \"sh -c './tool a'\"\n        out: \"x/depth.html\"\n        rc: 0\n        form: \"corpus: the pinned surface\"\n" +
		"  - id: REF\n    name: refusal rows\n    prose: |\n      Refusals answer and exit non-zero.\n" +
		"    rows:\n      - run: \"sh -c './tool --bad'\"\n        out: \"\"\n        rc: 2\n        form: \"corpus: the refusal surface\"\n"
	reply, rc := submit(t, stack, eng, round)
	mustAccept(t, reply, rc)
	reply, rc = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, rc, "suite=green")
	reply, rc = submit(t, stack, eng, "kind: run\nratify: true\n")
	mustAccept(t, reply, rc, "VERDICT: READY")
	text, code = eng.Report("v1", true)
	if code != Accepted {
		t.Fatalf("the round's since-report: %d %s", code, text)
	}
	for _, want := range []string{"changed since v1", "LIST/1", "retired since v1", "HOLD/1 (was green)"} {
		if !strings.Contains(text, want) {
			t.Fatalf("the round's since-report misses %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "REF/1") {
		t.Fatalf("the round's since-report re-prints the untouched REF/1:\n%s", text)
	}
	// the v2 boundary folds the retire: HOLD is gone from BOTH
	// corpora, the report says nothing moved
	text, code = eng.Report("v2", true)
	if code != Accepted || !strings.Contains(text, "nothing moved since v2") {
		t.Fatalf("the post-round v2 since-report: %d %s", code, text)
	}
	// a wide delta: 12 new rows after v2, the section caps its list
	var sb strings.Builder
	sb.WriteString("kind: spec\nfull-replace: true\nbehaviors:\n")
	for i := 0; i < 12; i++ {
		sb.WriteString("  - id: N")
		sb.WriteString(string(rune('A' + i)))
		sb.WriteString("\n    name: late wide\n    prose: |\n      p\n" +
			"    rows:\n      - run: \"true\"\n        rc: 0\n")
	}
	reply, rc = submit(t, stack, eng, sb.String())
	mustAccept(t, reply, rc)
	text, code = eng.Report("v2", true)
	if code != Accepted || !strings.Contains(text, "(+2 more)") {
		t.Fatalf("the wide since-report is not budget-capped: %d %s", code, text)
	}
	// a journal frame with an undecodable atom file rides the fold:
	// the walk skips it, the render still lands
	if err := eng.Journal.Append(journal.Entry{
		Transaction: "bogus-frame",
		Effects: []journal.Effect{{
			Path:    ".punchtape/canon/atoms/bogus.yaml",
			Content: "%%not-an-atom",
		}},
	}); err != nil {
		t.Fatal(err)
	}
	text, code = eng.Report("v2", true)
	if code != Accepted {
		t.Fatalf("the since-report under a bogus frame: %d %s", code, text)
	}
}

// A retired-then-re-added row is named by its CURRENT definition,
// not its retirement: the re-add supersedes the frame. The same walk
// folds a corrupted atom frame inside its window (a journal frame
// between two verdicts: an honest skip) and refuses a verdict newer
// than the case carries.
func TestSinceReaddedRowNotRetired(t *testing.T) {
	stack, eng := newInst(t)
	greenCycle(t, stack, eng)
	spec2 := strings.Replace(stateSpec("s/index.html", "held", ""),
		"  - id: HOLD\n    name: held behavior\n    prose: |\n      Stable rows stay green untouched.\n    rows:\n      - run: \"sh -c './tool b'\"\n        out: \"held\"\n        rc: 0\n        form: \"corpus: the held surface\"\n", "", 1)
	reply, rc := submit(t, stack, eng, fr(spec2))
	mustAccept(t, reply, rc)
	reply, rc = submit(t, stack, eng, fr(stateSpec("s/index.html", "held", "")))
	mustAccept(t, reply, rc)
	text, code := eng.Report("v1", true)
	if code != Accepted {
		t.Fatalf("the re-added since-report: %d %s", code, text)
	}
	for _, line := range strings.Split(text, "\n") {
		if strings.Contains(line, "HOLD/1 (was") {
			t.Fatalf("the re-added row is named as retired:\n%s", text)
		}
	}
	// a corrupted atom frame rides the NEXT fold window (a journal
	// frame between two verdicts): the walk skips it honestly
	if err := eng.Journal.Append(journal.Entry{
		Transaction: "bogus-frame",
		Effects: []journal.Effect{{
			Path:    ".punchtape/canon/atoms/bogus.yaml",
			Content: "%%not-an-atom",
		}},
	}); err != nil {
		t.Fatal(err)
	}
	reply, rc = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, rc, "suite=green")
	reply, rc = submit(t, stack, eng, "kind: run\nratify: true\n")
	mustAccept(t, reply, rc, "VERDICT: READY")
	text, code = eng.Report("v2", true)
	if code != Accepted {
		t.Fatalf("the since-report under a bogus frame in the window: %d %s", code, text)
	}
	// a verdict newer than the case carries is a rejection naming
	// the latest
	text, code = eng.Report("v9", true)
	if code != Rejected || !strings.Contains(text, "verdict v2") {
		t.Fatalf("the newer-than-latest verdict must reject naming the latest: %d %s", code, text)
	}
}

// A never-run retirement names its UNKNOWN standing: a row that
// leaves before any run measured it carries no outcome, and the
// retired line says so (the report renders before any ratify could
// refuse the unexplained departure).
func TestSinceRetiredUnknownStanding(t *testing.T) {
	stack, eng := newInst(t)
	greenCycle(t, stack, eng)
	spec2 := strings.Replace(stateSpec("s/index.html", "held", ""),
		"  - id: HOLD\n    name: held behavior\n    prose: |\n      Stable rows stay green untouched.\n    rows:\n      - run: \"sh -c './tool b'\"\n        out: \"held\"\n        rc: 0\n        form: \"corpus: the held surface\"\n", "", 1)
	spec2 += "  - id: GHOST\n    name: never measured\n    prose: |\n      Leaves before any run.\n" +
		"    rows:\n      - run: \"sh -c './tool g'\"\n        out: \"\"\n        rc: 0\n        form: \"corpus: the pinned surface\"\n"
	reply, rc := submit(t, stack, eng, fr(spec2))
	mustAccept(t, reply, rc)
	reply, rc = submit(t, stack, eng, fr(stateSpec("s/index.html", "held", "")))
	mustAccept(t, reply, rc)
	text, code := eng.Report("v1", true)
	if code != Accepted {
		t.Fatalf("the unknown-standing since-report: %d %s", code, text)
	}
	if !strings.Contains(text, "GHOST/1 (was unknown)") {
		t.Fatalf("the never-run retirement is not named by its unknown standing:\n%s", text)
	}
	if strings.Contains(text, "HOLD/1 (was") {
		t.Fatalf("the re-added HOLD is named as retired:\n%s", text)
	}
}

// The report is a read-only render, but its inputs can fail: a state or
// corpus fault is the honest failure, never a half table.
func TestReportFaultShapes(t *testing.T) {
	stack, eng := newInst(t)
	greenCycle(t, stack, eng)
	knobRead(stack, filepath.Join(eng.Store.Root(), "state.yaml"), func() {
		if text, code := eng.Report("", true); code != Failure {
			t.Fatalf("a full report under a state fault: %d %s", code, text)
		}
		if text, code := eng.Report("v1", true); code != Failure {
			t.Fatalf("a since-report under a state fault: %d %s", code, text)
		}
	})
	knobRead(stack, filepath.Join(eng.Store.Root(), "canon", "atoms", "list.yaml"), func() {
		if text, code := eng.Report("", true); code != Failure {
			t.Fatalf("a full report under a corpus fault: %d %s", code, text)
		}
		if text, code := eng.Report("v1", true); code != Failure {
			t.Fatalf("a since-report under a corpus fault: %d %s", code, text)
		}
	})
}

// The memo carries the observed facts: a green row's record names the
// exit code and the output digest of its run.
func TestMemoCarriesObservedFacts(t *testing.T) {
	stack, eng := newInst(t)
	greenCycle(t, stack, eng)
	atoms, err := eng.Store.Atoms()
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range atoms {
		for _, r := range a.Rows {
			if r.Outcome != canon.RowGreen {
				continue
			}
			if r.RanRc == nil || r.RanOut == "" {
				t.Fatalf("%s memo misses the observed facts: %+v", a.ID, r)
			}
		}
	}
}
