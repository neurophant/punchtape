// The door economy law: the renders are thin
// — the hand pays for what moved and what is red, never for the
// corpus it already knows; the canon (the .punchtape files) is
// byte-stable under every render. The numbers below are FIXED by
// this test's corpus (5 atoms x 6 rows with 100-byte wants): the
// door slot of a stood corpus fits in 2000 bytes where the full
// table of the same corpus is an order larger.
package engine

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/execport"
)

// wideCorpus — a wide green corpus: 5 atoms x 6 rows, long wants.
func wideCorpus(t *testing.T, stack *fakeStack, eng *Engine) {
	t.Helper()
	want := "held-" + strings.Repeat("x", 96)
	stack.proc.plan["sh -c wide"] = execport.Outcome{Exit: 0, Stdout: want + "\n"}
	var sb strings.Builder
	sb.WriteString("kind: spec\nfull-replace: true\nbehaviors:\n")
	for a := 0; a < 5; a++ {
		fmt.Fprintf(&sb, "  - id: A%d\n    name: wide-%d\n    prose: |\n      Prose of atom %d.\n    rows:\n", a, a, a)
		for r := 0; r < 6; r++ {
			fmt.Fprintf(&sb, "      - run: \"sh -c wide\"\n        out: \"%s\"\n        rc: 0\n        form: \"corpus: the wide surface\"\n", want)
		}
	}
	reply, code := submit(t, stack, eng, sb.String())
	mustAccept(t, reply, code)
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "suite=green")
}

// canonBytes — the atoms zone's bytes (canon stability witness).
func canonBytes(t *testing.T, eng *Engine) string {
	t.Helper()
	atoms, err := eng.Store.Atoms()
	if err != nil {
		t.Fatal(err)
	}
	var sb strings.Builder
	for _, a := range atoms {
		sb.WriteString(a.Digest)
	}
	return sb.String()
}

// The door slot of a stood corpus: the changed list is empty, the
// card is the status line plus the green counter — the whole render
// fits the fixed byte budget while the corpus carries 30 wide rows.
// The canon does not move under the render.
func TestDoorSlotByteBudget(t *testing.T) {
	stack, eng := newInst(t)
	wideCorpus(t, stack, eng)
	before := canonBytes(t, eng)
	// the hand's follow-up next drains the greens (the
	// ack is thin, the slot rides next)
	drained, err := eng.Next()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(drained, "CHANGED (30)") {
		t.Fatalf("the drained slot must name the greens once:\n%s", drained)
	}
	r, _, err := eng.renderNext(mustState(eng))
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Text) == 0 || len(r.Text) > 2000 {
		t.Fatalf("the door slot of a stood 30-row corpus is %d bytes (budget 2000):\n%s", len(r.Text), r.Text)
	}
	if !strings.Contains(r.Text, "nothing changed since the last slot") {
		t.Fatalf("the stood corpus must say so in one line:\n%s", r.Text)
	}
	if strings.Contains(r.Text, "held-xxxx") {
		t.Fatalf("the stood-corpus slot pays for row bodies:\n%s", r.Text)
	}
	if after := canonBytes(t, eng); after != before {
		t.Fatal("the render moved the canon")
	}
	// one row reddens: the door carries its change and its full red
	// line; the budget still holds
	stack.fs.files[filepath.Join(eng.Workdir, "moved")] = []byte("touched")
	reply, code := submit(t, stack, eng, "kind: run\nbehaviors:\n  - id: A0\n    name: wide-0\n    prose: |\n      Prose of atom 0.\n    rows:\n      - run: \"echo held\"\n        out: \"wrong\"\n        rc: 0\n        form: \"corpus: the wrong surface\"\n")
	mustAccept(t, reply, code, "A0/1: stdout: not equal")
	// the ack carries the reddened row's measured line; the corpus it
	// did not touch stays home
	if len(reply) > 6000 {
		t.Fatalf("the red door reply is %d bytes — the reds ride, the corpus does not:\n%s", len(reply), reply)
	}
	next, err := eng.Next()
	if err != nil || !strings.Contains(next, "A0/1: red") {
		t.Fatalf("the changed list must name the redden row: %v\n%s", err, next)
	}
}

// The door report: per-atom counters and the red rows in full; the
// full table lists every row (--full).
func TestDoorReportCounters(t *testing.T) {
	stack, eng := newInst(t)
	wideCorpus(t, stack, eng)
	door, code := eng.Report("", false)
	if code != Accepted {
		t.Fatalf("the door report: %d %s", code, door)
	}
	if !strings.Contains(door, "A0 wide-0 — rows 6: green 6, red 0, never 0") {
		t.Fatalf("the door report's atom counters:\n%s", door)
	}
	if !strings.Contains(door, "green rows: 30 (canon: report --full)") {
		t.Fatalf("the door report's green counter:\n%s", door)
	}
	if strings.Contains(door, "held-xxxx") {
		t.Fatalf("the door report pays for green row bodies:\n%s", door)
	}
	full, code := eng.Report("", true)
	if code != Accepted {
		t.Fatalf("the full report: %d %s", code, full)
	}
	for i := 0; i < 5; i++ {
		for j := 1; j <= 6; j++ {
			if !strings.Contains(full, fmt.Sprintf("A%d/%d", i, j)) {
				t.Fatalf("the full report misses A%d/%d:\n%s", i, j, full)
			}
		}
	}
	// a red row rides the door in full: label, command, want, reason
	stack.fs.files[filepath.Join(eng.Workdir, "moved")] = []byte("touched")
	reply, code := submit(t, stack, eng, "kind: spec\nfull-replace: true\nbehaviors:\n  - id: A0\n    name: wide-0\n    prose: |\n      Prose of atom 0.\n    rows:\n      - run: \"echo held\"\n        out: \"wrong\"\n        rc: 0\n        form: \"corpus: the wrong surface\"\n")
	mustAccept(t, reply, code)
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code)
	door, code = eng.Report("", false)
	mustAccept(t, door, code)
	if !strings.Contains(door, "A0/1 [red]") || !strings.Contains(door, "want: out wrong") {
		t.Fatalf("the red row does not ride the door in full:\n%s", door)
	}
	_ = canon.RowGreen
}

// The door surfaces' honest fault and edge shapes: an empty case says
// so, a state that cannot be read fails the render, a load fault
// fails the door report, and the red list keeps the reply budget.
func TestDoorEdgeShapes(t *testing.T) {
	stack, eng := newInst(t)
	door, code := eng.Report("", false)
	if code != Accepted || !strings.Contains(door, "no behaviors yet") {
		t.Fatalf("the empty-case door report: %d %s", code, door)
	}
	st := mustState(eng)
	st.Stage = canon.StageImplement
	knobRead(stack, filepath.Join(eng.Store.Root(), "state.yaml"), func() {
		if _, _, err := eng.renderNext(st); err == nil {
			t.Fatal("the render under a state fault succeeded")
		}
	})
	reply, rc := submit(t, stack, eng, fr(specDoc("B1", "hello")))
	mustAccept(t, reply, rc)
	knobRead(stack, filepath.Join(eng.Store.Root(), "canon", "atoms", "b1.yaml"), func() {
		if _, code := eng.Report("", false); code != Failure {
			t.Fatal("the door report under a load fault did not fail")
		}
	})
	// a wide red corpus: the door's red list is budget-capped
	wideCorpus(t, stack, eng)
	var sb strings.Builder
	sb.WriteString("kind: run\nbehaviors:\n")
	for a := 0; a < 3; a++ {
		fmt.Fprintf(&sb, "  - id: R%d\n    name: red-%d\n    prose: |\n      Red atom %d.\n    rows:\n", a, a, a)
		for r := 0; r < 5; r++ {
			sb.WriteString("      - run: \"false\"\n")
		}
	}
	reply, rc = submit(t, stack, eng, sb.String())
	mustAccept(t, reply, rc)
	door, code = eng.Report("", false)
	if code != Accepted {
		t.Fatalf("the red door report: %d %s", code, door)
	}
	if !strings.Contains(door, "(+5 more)") {
		t.Fatalf("the door red list is not budget-capped:\n%s", door)
	}
}

// The focus card of a mixed atom: the green counter rides beside the
// full red row (the door never pays for the green row bodies).
func TestMixedAtomCardCounter(t *testing.T) {
	stack, eng := newInst(t)
	stack.proc.plan["sh -c fine"] = execport.Outcome{Exit: 0, Stdout: "fine\n"}
	doc := "kind: spec\nbehaviors:\n  - id: MIX\n    name: mixed\n    prose: |\n      p\n    rows:\n      - run: \"sh -c fine\"\n        out: fine\n        form: \"corpus: the fine surface\"\n      - run: \"false\"\n"
	reply, code := submit(t, stack, eng, doc)
	mustAccept(t, reply, code)
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "suite=red")
	next, err := eng.Next()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(next, "rows (2): green 1, red 1, never 0") {
		t.Fatalf("the card's status line:\\n%s", next)
	}
	if !strings.Contains(next, "green rows: 1 (canon: report --full)") {
		t.Fatalf("the card's green counter:\\n%s", next)
	}
	if !strings.Contains(next, "MIX/2") || !strings.Contains(next, "exit code 1") {
		t.Fatalf("the red row rides in full:\\n%s", next)
	}
	if strings.Contains(next, "MIX/1: run") {
		t.Fatalf("the green row's body rides the card:\\n%s", next)
	}
}
