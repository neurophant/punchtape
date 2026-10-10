// The engine battery of the surface modes over the in-memory stack
// — the batch door call, the quiet and delta slots, the standings
// baseline, and every branch the CLI cycle cannot reach from above.
package engine

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// quietDoc — a two-behavior spec over the baseline tools.
func quietDoc(listOut string) string {
	return "kind: spec\nbehaviors:\n" +
		"  - id: LIST\n    name: the run listing\n    prose: |\n      The listing names what the run wrote.\n" +
		"    rows:\n      - run: \"echo " + listOut + "\"\n        out: \"" + listOut + "\"\n        rc: 0\n        form: \"corpus: the listing surface\"\n" +
		"  - id: HOLD\n    name: held behavior\n    prose: |\n      Stable rows stay green untouched.\n" +
		"    rows:\n      - run: \"echo held\"\n        out: \"held\"\n        rc: 0\n        form: \"corpus: the held surface\"\n"
}

// quietCycle — birth, spec v1, green run (the baseline exists).
func quietCycle(t *testing.T, stack *fakeStack, eng *Engine) {
	t.Helper()
	reply, code := submit(t, stack, eng, quietDoc("s/index.html"))
	mustAccept(t, reply, code)
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "suite=green")
}

// submitWith — the door call with options.
func submitWith(t *testing.T, stack *fakeStack, eng *Engine, doc string, opts SubmitOpts) (string, int) {
	t.Helper()
	if err := stack.fs.WriteFile(filepath.Join(eng.Workdir, door()), []byte(doc), 0o644); err != nil {
		t.Fatal(err)
	}
	return eng.SubmitSetOpts(door(), opts)
}

// The batch over the stack: spec + run in one call, two transactions,
// one reply carrying both blocks.
func TestBatchDoorCall(t *testing.T) {
	stack, eng := newInst(t)
	buildProduct(t, stack, eng)
	quietCycle(t, stack, eng)
	reply, code := submitWith(t, stack, eng, fr(quietDoc("index.html")), SubmitOpts{RunAfter: true, Mode: SlotModeDelta})
	mustAccept(t, reply, code, "SPEC v2", "suite=green")
	// the slot rides next (the thin-ack law); the batch's own reply
	// carries the two transactions' measured lines only
	next, err := eng.NextWith(SlotModeDelta)
	if err != nil || !strings.Contains(next, "STAGE:") || !strings.Contains(next, "ACTION:") {
		t.Fatalf("the post-batch delta slot: %v\n%s", err, next)
	}
	if strings.Count(reply, "ACCEPTED:") != 2 {
		t.Fatalf("the batch reply does not carry both transactions:\n%s", reply)
	}
	// The journal carries the same two transactions the manual steps
	// would have written (the synthetic run is the bare run itself).
	kinds := []string{}
	for _, en := range eng.Journal.All() {
		kinds = append(kinds, en.DeltaKind)
	}
	want := []string{"brief", "spec", "run", "spec", "run"}
	if strings.Join(kinds, ",") != strings.Join(want, ",") {
		t.Fatalf("journal kinds: %v, want %v", kinds, want)
	}
	// A run-kind document with --run-after submits as itself (the
	// flag names a thing the document already is).
	reply, code = submitWith(t, stack, eng, "kind: run\nratify: true\n", SubmitOpts{RunAfter: true, Mode: SlotModeQuiet})
	mustAccept(t, reply, code, "VERDICT: READY")
	if strings.Count(reply, "ACCEPTED:") != 1 {
		t.Fatalf("a run-kind --run-after chained a second transaction:\n%s", reply)
	}
	// A rejected spec does not run.
	reply, code = submitWith(t, stack, eng, "kind: spec\nbehaviors: []\n", SubmitOpts{RunAfter: true})
	mustReject(t, reply, code, "no behaviors")
	if strings.Count(reply, "ACCEPTED:") != 0 {
		t.Fatalf("a rejected batch accepted something:\n%s", reply)
	}
}

// The delta and quiet slots over the stack: the changed row names
// itself, the untouched rows do not ride; the stage move carries its
// was-marker; a vanished baseline says so.
func TestDeltaSlotBranches(t *testing.T) {
	stack, eng := newInst(t)
	buildProduct(t, stack, eng)
	quietCycle(t, stack, eng)
	// the hand's follow-up next establishes the baseline (submissions
	// do not render; the surfaces ride next)
	if _, err := eng.Next(); err != nil {
		t.Fatal(err)
	}
	// v2 changes ONE row's demand to a value the plan does not
	// answer: the batch's run redden it; the follow-up delta names
	// the standing move, the untouched row stays home
	v2 := strings.Replace(quietDoc("s/index.html"), "out: \"s/index.html\"", "out: \"depth.html\"", 1)
	reply, code := submitWith(t, stack, eng, fr(v2), SubmitOpts{RunAfter: true, Mode: SlotModeDelta})
	mustAccept(t, reply, code, "SPEC v2", "suite=red", "LIST/1")
	// the ack is thin; the changed row names itself in
	// the follow-up delta surface
	delta, derr := eng.NextWith(SlotModeDelta)
	if derr != nil || !strings.Contains(delta, "LIST/1: red") {
		t.Fatalf("the delta does not name the changed row: %v\n%s", derr, delta)
	}
	if strings.Contains(delta, "HOLD/1") {
		t.Fatalf("the delta re-prints the untouched row:\n%s", delta)
	}
	// The repair: the correct demand returns as a run-kind upsert
	// (a fresh document key), the suite greens, then the ratify signs.
	repair := strings.Replace(quietDoc("s/index.html"), "kind: spec", "kind: run\nfull-replace: true", 1)
	reply, code = submitWith(t, stack, eng, repair, SubmitOpts{})
	mustAccept(t, reply, code, "suite=green")
	// The stage move: the ratify's ack carries the verdict, the
	// follow-up delta slot carries the was-marker.
	reply, code = submitWith(t, stack, eng, "kind: run\nratify: true\n", SubmitOpts{Mode: SlotModeDelta})
	mustAccept(t, reply, code, "VERDICT: READY")
	delta, derr = eng.NextWith(SlotModeDelta)
	if derr != nil || !strings.Contains(delta, "(was implement)") {
		t.Fatalf("the stage move's was-marker: %v\n%s", derr, delta)
	}
	// A retired row names itself in the delta: v3 drops HOLD.
	spec3 := "kind: spec\nbehaviors:\n  - id: LIST\n    name: the run listing\n    prose: |\n      p\n" +
		"    rows:\n      - run: \"echo index.html\"\n        out: \"index.html\"\n        rc: 0\n        form: \"corpus: the pinned surface\"\n"
	reply, code = submitWith(t, stack, eng, fr(spec3), SubmitOpts{Mode: SlotModeDelta})
	mustAccept(t, reply, code, "-1 retired")
	delta, derr = eng.NextWith(SlotModeDelta)
	if derr != nil || !strings.Contains(delta, "HOLD/1: retired") {
		t.Fatalf("the retired row does not name itself in the delta: %v\n%s", derr, delta)
	}
	// The quiet slot: the facts stay, the card does not.
	text, err := eng.NextWith(SlotModeQuiet)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "STAGE:") || !strings.Contains(text, "ACTION:") {
		t.Fatalf("the quiet slot loses the orientation:\n%s", text)
	}
	if strings.Contains(text, "executable truth") {
		t.Fatalf("the quiet slot carries the card:\n%s", text)
	}
	// The full next is unchanged by the modes (the default tables).
	full, err := eng.Next()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(full, "CONTEXT:") {
		t.Fatalf("the full slot lost its context:\n%s", full)
	}
	// A delta next with no changes says so.
	text, err = eng.NextWith(SlotModeDelta)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "nothing changed since the last slot") {
		t.Fatalf("an unchanged delta next:\n%s", text)
	}
	// A NEW row names itself in a delta-mode spec reply: the row is
	// absent in the baseline, never-measured now.
	spec4 := spec3 + "  - id: LATE\n    name: late\n    prose: |\n      p\n" +
		"    rows:\n      - run: \"echo late\"\n        out: \"late\"\n        rc: 0\n        form: \"corpus: the pinned surface\"\n"
	reply, code = submitWith(t, stack, eng, fr(spec4), SubmitOpts{Mode: SlotModeDelta})
	mustAccept(t, reply, code, "+1 new")
	delta, derr = eng.NextWith(SlotModeDelta)
	if derr != nil || !strings.Contains(delta, "LATE/1: never (was new)") {
		t.Fatalf("the new row does not name itself in the delta: %v\n%s", derr, delta)
	}
	// Fault shapes: a state fault and a corpus fault inside the mode
	// renders are honest errors, never a half slot.
	knobRead(stack, filepath.Join(eng.Store.Root(), "state.yaml"), func() {
		if _, err := eng.NextWith(SlotModeDelta); err == nil {
			t.Fatal("a delta next under a state fault succeeded")
		}
	})
	knobRead(stack, filepath.Join(eng.Store.Root(), "canon", "atoms", "list.yaml"), func() {
		if _, err := eng.NextWith(SlotModeQuiet); err == nil {
			t.Fatal("a quiet next under a corpus fault succeeded")
		}
		if _, err := eng.NextWith(SlotModeDelta); err == nil {
			t.Fatal("a delta next under a corpus fault succeeded")
		}
	})
}

// The standings baseline: a vanished or corrupt baseline is the
// honest no-baseline delta; the standings walk every standing word.
func TestStandingsBaselineEdges(t *testing.T) {
	stack, eng := newInst(t)
	buildProduct(t, stack, eng)
	quietCycle(t, stack, eng)
	// every standing word: LIST green, a red row, a never-ran row, a
	// stale row (a named input moved).
	spec := fr(quietDoc("s/index.html")) +
		"  - id: RED\n    name: r\n    prose: |\n      p\n    rows:\n      - run: \"echo other\"\n        out: \"want\"\n        rc: 0\n        form: \"corpus: the pinned surface\"\n" +
		"  - id: NEW\n    name: n\n    prose: |\n      p\n    rows:\n      - run: \"cat named.txt\"\n        out: \"one\"\n        rc: 0\n        form: \"corpus: the pinned surface\"\n"
	_ = stack.fs.WriteFile(filepath.Join(eng.Workdir, "named.txt"), []byte("one\n"), 0o644)
	reply, code := submit(t, stack, eng, spec)
	mustAccept(t, reply, code)
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "suite=red")
	st, err := eng.Store.State()
	if err != nil {
		t.Fatal(err)
	}
	standings := eng.currentStandings(st)
	got := map[string]string{}
	for _, l := range standings.Rows {
		got[l.Label] = l.Standing
	}
	for label, want := range map[string]string{
		"LIST/1": "green", "HOLD/1": "green", "RED/1": "red", "NEW/1": "green",
	} {
		if got[label] != want {
			t.Fatalf("%s standing %q, want %q", label, got[label], want)
		}
	}
	// The named input moves: NEW/1 goes stale in the standings.
	_ = stack.fs.WriteFile(filepath.Join(eng.Workdir, "named.txt"), []byte("two\n"), 0o644)
	standings = eng.currentStandings(st)
	for _, l := range standings.Rows {
		if l.Label == "NEW/1" && l.Standing != "stale" {
			t.Fatalf("NEW/1 standing after the input moved: %q", l.Standing)
		}
	}
	// A delta render names the stale row.
	text, err := eng.NextWith(SlotModeDelta)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "NEW/1: stale") {
		t.Fatalf("the delta does not name the stale row:\n%s", text)
	}
	// The baseline vanished: the honest no-baseline delta.
	delete(stack.fs.files, filepath.Join(eng.Store.Root(), "cache", "slot-standings.yaml"))
	if base := eng.readSlotBaseline(); base != nil {
		t.Fatalf("a vanished baseline read back: %+v", base)
	}
	text, err = eng.NextWith(SlotModeDelta)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "no earlier slot") {
		t.Fatalf("the no-baseline delta:\n%s", text)
	}
	// A corrupt baseline is the same honest shape.
	_ = stack.fs.WriteFile(filepath.Join(eng.Store.Root(), "cache", "slot-standings.yaml"), []byte("%%bogus"), 0o644)
	if base := eng.readSlotBaseline(); base != nil {
		t.Fatalf("a corrupt baseline read back: %+v", base)
	}
	// The budget caps a wide delta: 12 rows flip in one run.
	stack2, eng2 := newInst(t)
	var sb strings.Builder
	sb.WriteString("kind: spec\nbehaviors:\n")
	for i := 0; i < 12; i++ {
		id := "R" + string(rune('A'+i))
		sb.WriteString("  - id: ")
		sb.WriteString(id)
		sb.WriteString("\n    name: n\n    prose: |\n      p\n    rows:\n      - run: \"true\"\n        rc: 0\n")
	}
	reply, code = submit(t, stack2, eng2, sb.String())
	mustAccept(t, reply, code)
	// the run's delta names the twelve measured rows — budget-capped
	reply, code = submitWith(t, stack2, eng2, "kind: run\n", SubmitOpts{Mode: SlotModeDelta})
	mustAccept(t, reply, code, "suite=green")
	delta, derr := eng2.NextWith(SlotModeDelta)
	if derr != nil || !strings.Contains(delta, "(+2 more)") {
		t.Fatalf("the wide delta is not budget-capped: %v\n%s", derr, delta)
	}
}

// The anchor law: the delta's ground is the
// last SLOT — any next surface anchors it; submit acks carry no slot
// and never re-anchor. A spec+run pair whose net standings return to
// the anchor's leaves the follow-up delta empty (the spec did not
// re-anchor at its own never); a spec that moves one row lands its
// name ALONE in the immediate follow-up delta, no run between.
func TestDeltaAnchorIsTheLastSlot(t *testing.T) {
	stack, eng := newInst(t)
	buildProduct(t, stack, eng)
	quietCycle(t, stack, eng)
	// The anchor: a next after the green run — the standings all green.
	if _, err := eng.Next(); err != nil {
		t.Fatal(err)
	}
	// v2 redefines LIST (run and out move together, the suite stays
	// green), then the run remeasures it: two submits, the net
	// standings IDENTICAL to the anchor's.
	v2 := strings.Replace(quietDoc("s/index.html"), "s/index.html", "depth.html", -1)
	reply, code := submit(t, stack, eng, fr(v2))
	mustAccept(t, reply, code, "SPEC v2")
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "suite=green")
	delta, err := eng.NextWith(SlotModeDelta)
	if err != nil || !strings.Contains(delta, "nothing changed since the last slot") {
		t.Fatalf("a submit pair that nets to the anchor must leave the delta empty: %v\n%s", err, delta)
	}
	// v3 moves ONE row again: the immediate delta (no run between)
	// names it alone, under the header that names the true anchor.
	v3 := strings.Replace(quietDoc("s/index.html"), "s/index.html", "breadth.html", -1)
	reply, code = submit(t, stack, eng, fr(v3))
	mustAccept(t, reply, code, "SPEC v3")
	delta, err = eng.NextWith(SlotModeDelta)
	if err != nil || !strings.Contains(delta, "CHANGED SINCE THE LAST SLOT (1):") {
		t.Fatalf("the delta header lost its anchor-naming form: %v\n%s", err, delta)
	}
	if !strings.Contains(delta, "LIST/1: never (was green)") || strings.Contains(delta, "HOLD/1") {
		t.Fatalf("the delta does not isolate the moved row:\n%s", delta)
	}
}

// The surface options' honest refusals: a foreign path is the door
// refusal whatever the flags; a door that cannot be read fails; the
// mode parse maps only the three spellings.
func TestDoorAndModeEdges(t *testing.T) {
	stack, eng := newInst(t)
	buildProduct(t, stack, eng)
	quietCycle(t, stack, eng)
	if _, code := eng.SubmitSetOpts("notes.txt", SubmitOpts{RunAfter: true}); code != Rejected {
		t.Fatalf("a foreign path with flags: %d", code)
	}
	knobRead(stack, filepath.Join(eng.Workdir, door()), func() {
		_ = stack.fs.WriteFile(filepath.Join(eng.Workdir, door()), []byte("kind: run\n"), 0o644)
		if _, code := eng.SubmitSetOpts(door(), SubmitOpts{}); code != Failure {
			t.Fatalf("an unreadable door: %d", code)
		}
	})
	for _, c := range []struct {
		in   string
		want slotMode
	}{{SlotModeQuiet, slotQuiet}, {SlotModeDelta, slotDelta}, {SlotModeFull, slotFull}, {"bogus", slotFull}} {
		if got := parseSlotMode(c.in); got != c.want {
			t.Fatalf("parseSlotMode(%q) = %d, want %d", c.in, got, c.want)
		}
	}
	// The mode names ride the ledger facts.
	if got := modeName(slotQuiet); got != "quiet" {
		t.Fatalf("modeName quiet: %q", got)
	}
	if got := modeName(slotFull); got != "full" {
		t.Fatalf("modeName full: %q", got)
	}
	// A standings walk under a corpus fault is the honest empty set.
	knobRead(stack, filepath.Join(eng.Store.Root(), "canon", "atoms", "list.yaml"), func() {
		got := eng.currentStandings(canon.State{Stage: canon.StageImplement})
		if len(got.Rows) != 0 {
			t.Fatalf("standings under corpus fault: %+v", got)
		}
	})
	// The baseline round-trips.
	st, _ := eng.Store.State()
	eng.writeSlotBaseline(eng.currentStandings(st))
	if base := eng.readSlotBaseline(); base == nil || base.Stage != st.Stage {
		t.Fatalf("the baseline did not round-trip: %+v", base)
	}
	_ = yamlio.Digest // keep the import honest if unused paths change
}
