// The machine-economy law tests: the door catches its
// own blank template (the door accidents are
// mechanically impossible now), later submissions are SPARSE PATCHES
// by row address (point 1a — the number of submissions stops being
// an economic variable), the deliberate full rewrite rides its
// explicit hatch, and patches replay idempotently by content.
// The 20-patch equivalence proof drives two instances with the SAME
// changes — one through patches, one through full re-submissions
// built by the test's own model — and demands equal canon table
// definitions at every step (and byte-equal canon while nothing
// executes: the machine's own run memos are its measurements, not
// the submission's content).
package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/execport"
)

// --- point 1: the template catch ---

// The conduit accident, mechanically dead: after an accepted spec the
// door RESETS to the blank template; submitting the template bytes
// (or an empty file) is refused unread, the instance is untouched,
// and a strip of repeats answers identically. The honest submission
// after overwriting the door is accepted, and the ack names the
// reset. (Without the template catch this test is RED: the template decoded
// as a grammatically valid spec v2 and drove the instance off its
// rails.)
func TestDoorCatchesBlankTemplate(t *testing.T) {
	stack, eng := newInst(t)
	reply, code := submit(t, stack, eng, specDoc("B1", "hello"))
	mustAccept(t, reply, code)
	canonBefore := canonBytes(t, eng)
	journalBefore := len(eng.Journal.All())
	specsBefore := mustState(eng).Counters["SPEC"]

	// the door as the machine left it: the blank template
	tpl := canondata.TemplateBytes(inboxTemplate)
	if err := stack.fs.WriteFile(filepath.Join(eng.Workdir, door()), tpl, 0o644); err != nil {
		t.Fatal(err)
	}
	reply, code = eng.SubmitSet(door())
	mustReject(t, reply, code, "the door file is the blank template")
	if line := strings.SplitN(reply, "\n", 2)[0]; line != reject("door", canondata.T("submit.error.template-resubmitted")) {
		t.Fatalf("the refusal's diagnosis must be exactly the one canon line:\n%s", reply)
	}
	// the instance is untouched: canon, journal, counters
	if after := canonBytes(t, eng); after != canonBefore {
		t.Fatal("the template refusal moved the canon")
	}
	if len(eng.Journal.All()) != journalBefore {
		t.Fatal("the template refusal wrote a journal transaction")
	}
	if mustState(eng).Counters["SPEC"] != specsBefore {
		t.Fatal("the template refusal bumped the spec counter")
	}
	// the empty door file: the same refusal
	if err := stack.fs.WriteFile(filepath.Join(eng.Workdir, door()), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	reply, code = eng.SubmitSet(door())
	mustReject(t, reply, code, "blank template")
	// the gate strip: five more template submissions answer
	// identically, the state stays put
	for i := 0; i < 5; i++ {
		if err := stack.fs.WriteFile(filepath.Join(eng.Workdir, door()), tpl, 0o644); err != nil {
			t.Fatal(err)
		}
		reply, code = eng.SubmitSet(door())
		mustReject(t, reply, code, "blank template")
	}
	if after := canonBytes(t, eng); after != canonBefore || len(eng.Journal.All()) != journalBefore {
		t.Fatal("the repeated template refusals are not idempotent")
	}
	// the honest route: overwrite the door, the submission is taken
	stack.proc.plan["echo hello"] = execport.Outcome{Exit: 0, Stdout: "hello\n"}
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "suite=green")
	if !strings.Contains(reply, "DOOR: reset to the blank template") {
		t.Fatalf("the accepted ack must name the door reset:\n%s", reply)
	}
}

// --- point 1a: the sparse patches ---

// pEcoRow — one model row: a scripted echo command with its exact
// stdout, the form source every exact pin owes, an optional launch
// mode.
type pEcoRow struct {
	cmd, out, launch string
}

// body — the row's YAML fields at the given indent.
func (r pEcoRow) body(indent string) string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "%srun: \"echo %s\"\n", indent, r.cmd)
	if r.launch != "" {
		fmt.Fprintf(sb, "%slaunch: %s\n", indent, r.launch)
	}
	fmt.Fprintf(sb, "%sout: \"%s\"\n%src: 0\n%sform: \"corpus: the pinned surface\"\n",
		indent, r.out, indent, indent)
	return sb.String()
}

// pEcoAtom — one model behavior.
type pEcoAtom struct {
	id, name, prose string
	rows            []pEcoRow
}

// pEcoModel — the test's own table model: the single source both
// documents (the patch and the full re-submission) render from.
type pEcoModel struct {
	atoms  []*pEcoAtom
	nextID int
}

func (m *pEcoModel) atom(id string) *pEcoAtom {
	for _, a := range m.atoms {
		if a.id == id {
			return a
		}
	}
	return nil
}

func (m *pEcoModel) newRow() pEcoRow {
	m.nextID++
	r := pEcoRow{cmd: fmt.Sprintf("s%d", m.nextID), out: ""}
	r.out = r.cmd
	return r
}

// fullDoc — the model as the FULL document a hand would re-send (the
// test's own construction, independent of the engine's expansion).
func (m *pEcoModel) fullDoc() string {
	var sb strings.Builder
	sb.WriteString("kind: spec\nfull-replace: true\nbehaviors:\n")
	for _, a := range m.atoms {
		fmt.Fprintf(&sb, "  - id: %s\n    name: %s\n    prose: |\n      %s\n    rows:\n", a.id, a.name, a.prose)
		for _, r := range a.rows {
			// the row item: "- " carries the first line, the fields
			// keep their own deeper indent
			sb.WriteString("      - " + strings.TrimPrefix(r.body("        "), "        "))
		}
	}
	return sb.String()
}

// defDigests — the canon's TABLE DEFINITION digests per atom (the
// row definitions, names and prose; the machine's own run records
// are its measurements and never ride a submission).
func defDigests(t *testing.T, eng *Engine) map[string]string {
	t.Helper()
	atoms, err := eng.Store.Atoms()
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]string{}
	for _, a := range atoms {
		var sb strings.Builder
		sb.WriteString(a.ID)
		sb.WriteString("\x00")
		sb.WriteString(a.Name)
		sb.WriteString("\x00")
		sb.WriteString(a.Prose)
		for _, r := range a.Rows {
			sb.WriteString("\x00")
			sb.WriteString(r.RowDefDigest())
		}
		out[a.ID] = sb.String()
	}
	return out
}

// The 20-patch equivalence proof: two instances, the
// same changes — A through patches by row address, B through full
// re-submissions of the test's own model — the table definitions are
// EQUAL after every step (and the full canon bytes are equal at the
// checkpoint before the first execution: identical documents through
// the one apply law), and each patch moves ONLY the atoms it touched
// (the canon delta IS the patch content; no hidden re-renders).
func TestTwentyPatchesEqualFullResubmissions(t *testing.T) {
	// two instances in ONE bound stack (the harness's single global
	// fs), two work directories: A walks the patches, B the full
	// re-submissions
	stack := bindUnits(t)
	engA := birthAt(t, stack, "/w/a")
	engB := birthAt(t, stack, "/w/b")
	model := &pEcoModel{atoms: []*pEcoAtom{{
		id: "T1", name: "core", prose: "The core behavior.",
		rows: []pEcoRow{{cmd: "v1", out: "v1"}, {cmd: "v2", out: "v2"}, {cmd: "v3", out: "v3"}},
	}}}
	planRow := func(rs ...pEcoRow) {
		for _, r := range rs {
			stack.proc.plan["echo "+r.cmd] = execport.Outcome{Exit: 0, Stdout: r.out + "\n"}
		}
	}
	planRow(model.atoms[0].rows...)
	// both instances take the same first full spec
	first := strings.Replace(model.fullDoc(), "kind: spec\nfull-replace: true\n", "kind: spec\n", 1)
	reply, code := submit(t, stack, engA, first)
	mustAccept(t, reply, code, "SPEC v1")
	reply, code = submit(t, stack, engB, first)
	mustAccept(t, reply, code, "SPEC v1")
	if canonBytes(t, engA) != canonBytes(t, engB) {
		t.Fatal("the two instances diverge at birth")
	}

	// one step: the patch to A, the model-rendered full to B — the
	// definitions must agree; untouched atoms stay byte-identical
	step := func(name, patchDoc string, touch map[string]bool) {
		t.Helper()
		before := atomDigests(t, engA)
		reply, code := submit(t, stack, engA, patchDoc)
		mustAccept(t, reply, code)
		reply, code = submit(t, stack, engB, model.fullDoc())
		mustAccept(t, reply, code)
		after := atomDigests(t, engA)
		for id, d := range before {
			if touch[id] {
				continue
			}
			if after[id] != d {
				t.Fatalf("step %s: the patch moved untouched atom %s", name, id)
			}
		}
		da, db := defDigests(t, engA), defDigests(t, engB)
		if len(da) != len(db) {
			t.Fatalf("step %s: atom sets diverge: %v vs %v", name, da, db)
		}
		for id, d := range da {
			if db[id] != d {
				t.Fatalf("step %s: atom %s definition diverges between the patch and the full path", name, id)
			}
		}
	}
	T := map[string]bool{"T1": true}
	N := map[string]bool{"N1": true}
	both := map[string]bool{"T1": true, "N1": true}
	addPatch := func(atom string, r pEcoRow) string {
		return "kind: spec\npatch:\n  add:\n    - atom: " + atom + "\n      row:\n        " +
			strings.TrimPrefix(r.body("        "), "        ")
	}
	replacePatch := func(label string, r pEcoRow) string {
		return "kind: spec\npatch:\n  replace:\n    " + label + ":\n      " +
			strings.TrimPrefix(r.body("      "), "      ")
	}

	// 1. add to the END of T1
	r := model.newRow()
	planRow(r)
	model.atom("T1").rows = append(model.atom("T1").rows, r)
	step("add-1", addPatch("T1", r), T)
	// 2. replace T1/1 in place
	fix := pEcoRow{cmd: "fix1", out: "fix1"}
	planRow(fix)
	model.atom("T1").rows[0] = fix
	step("replace-1", replacePatch("T1/1", fix), T)
	// 3. add another
	r = model.newRow()
	planRow(r)
	model.atom("T1").rows = append(model.atom("T1").rows, r)
	step("add-2", addPatch("T1", r), T)
	// 4. remove T1/2
	model.atom("T1").rows = append(model.atom("T1").rows[:1], model.atom("T1").rows[2:]...)
	step("remove-1", "kind: spec\npatch:\n  remove: [T1/2]", T)
	// CHECKPOINT: nothing has executed since birth — the full canon
	// bytes are equal (the one apply law, byte for byte)
	if canonBytes(t, engA) != canonBytes(t, engB) {
		t.Fatal("before the first execution the canon bytes diverge")
	}
	// 5. run T1/1 (a forced execution — a run-kind patch)
	step("run-1", "kind: run\npatch:\n  run:\n    T1/1: {}", T)
	// 6. replace T1/3 (the model's third row is index 3)
	fix3 := pEcoRow{cmd: "fix3", out: "fix3"}
	planRow(fix3)
	model.atom("T1").rows[2] = fix3
	step("replace-2", replacePatch("T1/3", fix3), T)
	// 7. add
	r = model.newRow()
	planRow(r)
	model.atom("T1").rows = append(model.atom("T1").rows, r)
	step("add-3", addPatch("T1", r), T)
	// 8. remove T1/3 — every patch's BYTES must be unique in the A
	// chain (the machine's own law: the same patch again is an honest
	// replay, and a replay applies nothing)
	model.atom("T1").rows = append(model.atom("T1").rows[:2], model.atom("T1").rows[3:]...)
	step("remove-2", "kind: spec\npatch:\n  remove: [T1/3]", T)
	// 9. open a NEW behavior with its first row
	r = model.newRow()
	planRow(r)
	model.atoms = append(model.atoms, &pEcoAtom{
		id: "N1", name: "second", prose: "The second behavior.",
		rows: []pEcoRow{r},
	})
	step("new-atom", "kind: spec\npatch:\n  add:\n    - atom: N1\n      name: second\n      prose: |\n        The second behavior.\n      row:\n        "+
		strings.TrimPrefix(r.body("        "), "        "), N)
	// 10. replace N1/1
	nfix := pEcoRow{cmd: "nfix", out: "nfix"}
	planRow(nfix)
	model.atom("N1").rows[0] = nfix
	step("replace-3", replacePatch("N1/1", nfix), N)
	// 11. add a row to N1
	r = model.newRow()
	planRow(r)
	model.atom("N1").rows = append(model.atom("N1").rows, r)
	step("add-4", addPatch("N1", r), N)
	// 12. remove T1/1
	model.atom("T1").rows = model.atom("T1").rows[1:]
	step("remove-3", "kind: spec\npatch:\n  remove: [T1/1]", T)
	// 13. prose replacement on T1 (the patch's literal block carries
	// the same trailing newline the model's full doc carries — a
	// block without it is different prose BYTES, not the same change)
	model.atom("T1").prose = "The core behavior, reworded."
	step("prose-1", "kind: spec\npatch:\n  prose:\n    T1: |\n      The core behavior, reworded.\n", T)
	// 14. add
	r = model.newRow()
	planRow(r)
	model.atom("T1").rows = append(model.atom("T1").rows, r)
	step("add-5", addPatch("T1", r), T)
	// 15. replace the (new) T1/1
	fix1b := pEcoRow{cmd: "fix1b", out: "fix1b"}
	planRow(fix1b)
	model.atom("T1").rows[0] = fix1b
	step("replace-4", replacePatch("T1/1", fix1b), T)
	// 16. remove N1/2
	model.atom("N1").rows = model.atom("N1").rows[:1]
	step("remove-4", "kind: spec\npatch:\n  remove: [N1/2]", N)
	// 17. add to both atoms in ONE patch
	r = model.newRow()
	r2 := model.newRow()
	planRow(r, r2)
	model.atom("T1").rows = append(model.atom("T1").rows, r)
	model.atom("N1").rows = append(model.atom("N1").rows, r2)
	step("add-6-both", addPatch("T1", r)+"    - atom: N1\n      row:\n        "+
		strings.TrimPrefix(r2.body("        "), "        "), both)
	// 18. run T1/2 forced (a different address than run-1: the same
	// patch bytes again would be an honest no-op replay). A run is a
	// CORPUS-WIDE measurement: every due row executes, and N1's rows
	// (opened after run-1, never run) are due — their outcomes are the
	// machine's own records, the definitions still match B exactly
	step("run-2", "kind: run\npatch:\n  run:\n    T1/2: {}", both)
	// 19. replace T1/1 with a launch mode (a definition change)
	fix1c := pEcoRow{cmd: "fix1c", out: "fix1c", launch: "nonet"}
	planRow(fix1c)
	model.atom("T1").rows[0] = fix1c
	step("replace-5", replacePatch("T1/1", fix1c), T)
	// 20. add the final row
	r = model.newRow()
	planRow(r)
	model.atom("T1").rows = append(model.atom("T1").rows, r)
	step("add-7", addPatch("T1", r), T)

	// the spec counter grew with every SPEC step: eighteen cheap spec
	// patches after v1 (the two run-kind steps measure, they do not
	// version the table), each the size of its own change
	if got := mustState(engA).Counters["SPEC"]; got != 19 {
		t.Fatalf("the spec counter after v1 + 18 spec steps: %d", got)
	}
	// the run sections really executed (A measured twice; B never
	// runs — its full re-submissions are kind: spec)
	if got := len(stack.proc.starts); got < 2 {
		t.Fatalf("the forced runs did not execute: %v", stack.proc.starts)
	}
}

// birthAt — an instance born at the given directory of the bound
// stack (the harness's single global filesystem carries both
// instances of the equivalence proof).
func birthAt(t *testing.T, stack *fakeStack, dir string) *Engine {
	t.Helper()
	eng, err := Init(dir)
	if err != nil {
		t.Fatalf("init %s: %v", dir, err)
	}
	if reply, code := eng.Begin("a tool that greps lines in files"); code != Accepted {
		t.Fatalf("begin %s: %d %s", dir, code, reply)
	}
	return eng
}

// The full-resubmit refusal and the deliberate hatch:
// a second FULL document without the mark is refused in one line; a
// run-kind document covering every corpus atom is refused the same
// way (a partial upsert stays the cheap legal repair); with
// full-replace: true the rewrite is taken, the ack counts the rows
// and the cell ledger carries the event.
func TestFullReplaceHatchAndRefusal(t *testing.T) {
	stack, eng := newInst(t)
	reply, code := submit(t, stack, eng, specDoc("B1", "hello"))
	mustAccept(t, reply, code)
	// the implicit repeat: refused, one line
	reply, code = submit(t, stack, eng, specDoc("B1", "reworded"))
	mustReject(t, reply, code, "repeated full document forbidden")
	if strings.Count(reply, "ERROR:") != 1 {
		t.Fatalf("the refusal must be one line:\n%s", reply)
	}
	// the run-kind full table: refused the same way
	stack.proc.plan["echo hello"] = execport.Outcome{Exit: 0, Stdout: "hello\n"}
	fullRun := "kind: run\nbehaviors:\n  - id: B1\n    name: sample\n    prose: |\n      One paragraph.\n    rows:\n      - run: \"echo hello\"\n        out: \"hello\"\n        rc: 0\n        form: \"test corpus: the pinned surface\"\n"
	reply, code = submit(t, stack, eng, fullRun)
	mustReject(t, reply, code, "repeated full document forbidden")
	// a partial upsert stays legal (one atom of a two-atom corpus)
	two := "kind: spec\nfull-replace: true\nbehaviors:\n" +
		"  - id: B1\n    name: sample\n    prose: |\n      One paragraph.\n    rows:\n      - run: \"echo hello\"\n        out: \"hello\"\n        rc: 0\n        form: \"test corpus: the pinned surface\"\n" +
		"  - id: B2\n    name: second\n    prose: |\n      One paragraph.\n    rows:\n      - run: \"echo hello\"\n        out: \"hello\"\n        rc: 0\n        form: \"test corpus: the pinned surface\"\n"
	reply, code = submit(t, stack, eng, two)
	mustAccept(t, reply, code, "FULL-REPLACE: 1 rows -> 2 rows")
	found := false
	for _, en := range eng.Ledger.All() {
		if en.Details["full-replace"] == "1->2" {
			found = true
		}
	}
	if !found {
		t.Fatal("the full-replace event is not in the cell ledger")
	}
	partial := "kind: run\nbehaviors:\n  - id: B1\n    name: sample\n    prose: |\n      One paragraph.\n    rows:\n      - run: \"echo hello\"\n        out: \"hello\"\n        rc: 0\n        form: \"test corpus: the pinned surface\"\n"
	reply, code = submit(t, stack, eng, partial)
	mustAccept(t, reply, code, "suite=green")
	// patch + full-replace never compose
	reply, code = submit(t, stack, eng, "kind: spec\nfull-replace: true\npatch:\n  remove: [B1/1]\n")
	mustReject(t, reply, code, "full-replace and patch do not compose")
	// a patch before any spec is refused (the first spec is the full one)
	stack2, eng2 := newInst(t)
	reply, code = submit(t, stack2, eng2, "kind: spec\npatch:\n  prose:\n    B1: |\n      p\n")
	mustReject(t, reply, code, "a patch needs the first spec")
}

// The patch grammar's honest refusals: an unknown address, a label
// riding two sections, an illegal run-mode value, an empty patch, a
// prose for an unknown behavior, a forged machine record — each one
// line, the corpus untouched.
func TestPatchGrammarRefusals(t *testing.T) {
	stack, eng := newInst(t)
	reply, code := submit(t, stack, eng, specDoc("B1", "hello"))
	mustAccept(t, reply, code)
	before := canonBytes(t, eng)
	refused := func(doc, want string) {
		t.Helper()
		reply, code = submit(t, stack, eng, doc)
		mustReject(t, reply, code, want)
		if after := canonBytes(t, eng); after != before {
			t.Fatalf("the refusal moved the canon: %s", want)
		}
	}
	refused("kind: spec\npatch:\n  replace:\n    B9/1:\n      run: \"echo x\"\n      out: x\n      rc: 0\n", "does not exist in the corpus")
	refused("kind: spec\npatch:\n  replace:\n    B1/99:\n      run: \"echo x\"\n      out: x\n      rc: 0\n", "does not exist in the corpus")
	refused("kind: spec\npatch:\n  replace:\n    B1/one:\n      run: \"echo x\"\n      out: x\n      rc: 0\n", "is not a row address")
	refused("kind: spec\npatch:\n  replace:\n    B1/1:\n      run: \"echo x\"\n      out: x\n      rc: 0\n  remove: [B1/1]\n", "addressed by more than one section")
	refused("kind: spec\npatch:\n  remove: [B1/1]\n  run:\n    B1/1: {}\n", "addressed by more than one section")
	refused("kind: spec\npatch:\n  run:\n    B1/1:\n      launch: everywhere\n", "illegal mode value")
	refused("kind: spec\npatch: {}\n", "carries no operations")
	refused("kind: spec\npatch:\n  prose:\n    NOPE: |\n      p\n", "which the corpus does not carry")
	refused("kind: spec\npatch:\n  replace:\n    B1/1:\n      run: \"echo x\"\n      out: x\n      rc: 0\n      outcome: green\n", "machine-recorded fields")
}

// The patch idempotency law: the same patch bytes again are an
// honest ALREADY-APPLIED replay — no transaction, no counter growth.
func TestPatchIdempotentReplay(t *testing.T) {
	stack, eng := newInst(t)
	reply, code := submit(t, stack, eng, specDoc("B1", "hello"))
	mustAccept(t, reply, code)
	patch := "kind: spec\npatch:\n  prose:\n    B1: |\n      The reworded paragraph.\n"
	reply, code = submit(t, stack, eng, patch)
	mustAccept(t, reply, code, "SPEC v2")
	journalBefore := len(eng.Journal.All())
	canonBefore := canonBytes(t, eng)
	reply, code = submit(t, stack, eng, patch)
	mustAccept(t, reply, code, "IDEMPOTENT REPLAY")
	if len(eng.Journal.All()) != journalBefore {
		t.Fatal("the patch replay wrote a transaction")
	}
	if after := canonBytes(t, eng); after != canonBefore {
		t.Fatal("the patch replay moved the canon")
	}
}

// The run section is an explicit measurement: a row whose green
// stands fresh re-runs when the patch names it (and only it).
func TestRunSectionForcesExecution(t *testing.T) {
	stack, eng := newInst(t)
	doc := "kind: spec\nbehaviors:\n  - id: B1\n    name: sample\n    prose: |\n      One paragraph.\n    rows:\n      - run: \"echo one\"\n        out: \"one\"\n        rc: 0\n        form: \"test corpus: the pinned surface\"\n      - run: \"echo two\"\n        out: \"two\"\n        rc: 0\n        form: \"test corpus: the pinned surface\"\n"
	stack.proc.plan["echo one"] = execport.Outcome{Exit: 0, Stdout: "one\n"}
	stack.proc.plan["echo two"] = execport.Outcome{Exit: 0, Stdout: "two\n"}
	reply, code := submit(t, stack, eng, doc)
	mustAccept(t, reply, code)
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "suite=green")
	stack.proc.starts = nil
	// both rows are fresh; the run patch names B1/1 only: it
	// re-executes, B1/2 keeps its memo
	reply, code = submit(t, stack, eng, "kind: run\npatch:\n  run:\n    B1/1: {}\n")
	mustAccept(t, reply, code)
	if got := len(stack.proc.starts); got != 1 || !strings.HasSuffix(strings.Join(stack.proc.starts[0], " "), "echo one") {
		t.Fatalf("the forced run executed the wrong rows: %v", stack.proc.starts)
	}
	if !strings.Contains(reply, "kept") {
		t.Fatalf("the suite line must name the kept rows:\n%s", reply)
	}
}

// The dry validator speaks the patch law too:
// a patch draft checks against the read-only corpus with the door's
// own vocabulary — the unknown address is a finding, a good patch
// passes clean.
func TestSpecCheckUnderstandsPatches(t *testing.T) {
	stack, eng := newInst(t)
	reply, code := submit(t, stack, eng, specDoc("B1", "hello"))
	mustAccept(t, reply, code)
	draft := filepath.Join(eng.Workdir, "draft.yaml")
	bad := "kind: spec\npatch:\n  replace:\n    B1/7:\n      run: \"echo x\"\n      out: x\n      rc: 0\n"
	if err := stack.fs.WriteFile(draft, []byte(bad), 0o644); err != nil {
		t.Fatal(err)
	}
	lines, err := CheckDocument(eng.Workdir, filepath.Base(draft))
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) == 0 || !strings.Contains(lines[0], "B1/7") {
		t.Fatalf("the draft check missed the unknown address: %v", lines)
	}
	good := "kind: spec\npatch:\n  prose:\n    B1: |\n      The checked paragraph.\n"
	if err := stack.fs.WriteFile(draft, []byte(good), 0o644); err != nil {
		t.Fatal(err)
	}
	lines, err = CheckDocument(eng.Workdir, filepath.Base(draft))
	if err != nil || len(lines) != 0 {
		t.Fatalf("a good patch draft must check clean: %v %v", lines, err)
	}
}

// The patch grammar's refusal branch sweep: every parse, address,
// mode and identity refusal the door can name, each one line, the
// corpus untouched — and the two structural branch pairs (the spec
// patch that retires an emptied atom; the run patch that cannot
// empty one).
func TestPatchRefusalBranches(t *testing.T) {
	stack, eng := newInst(t)
	reply, code := submit(t, stack, eng, specDoc("B1", "hello"))
	mustAccept(t, reply, code)
	before := canonBytes(t, eng)
	refused := func(doc, want string) {
		t.Helper()
		reply, code = submit(t, stack, eng, doc)
		mustReject(t, reply, code, want)
		if after := canonBytes(t, eng); after != before {
			t.Fatalf("the refusal moved the canon: %s", want)
		}
	}
	// parseRowAddr: no leading id, no trailing number, an illegal id
	refused("kind: spec\npatch:\n  replace:\n    /1:\n      run: \"echo x\"\n      out: x\n      rc: 0\n", "is not a row address")
	refused("kind: spec\npatch:\n  replace:\n    B1/:\n      run: \"echo x\"\n      out: x\n      rc: 0\n", "is not a row address")
	refused("kind: spec\npatch:\n  replace:\n    b!/1:\n      run: \"echo x\"\n      out: x\n      rc: 0\n", "is not a row address")
	// the remove and run sections resolve their own addresses
	refused("kind: spec\npatch:\n  remove: [B1/99]\n", "does not exist in the corpus")
	refused("kind: spec\npatch:\n  run:\n    NOPE/1: {}\n", "does not exist in the corpus")
	// the run mode's tree value is as closed as its launch value
	refused("kind: spec\npatch:\n  run:\n    B1/1:\n      tree: wat\n", "illegal mode value")
	// the add items carry the same laws as any row
	refused("kind: spec\npatch:\n  add:\n    - atom: b!\n      row:\n        run: \"echo x\"\n", "is not a row address")
	refused("kind: spec\npatch:\n  add:\n    - atom: B1\n      row:\n        run: \"echo x\"\n        out: x\n        rc: 0\n        outcome: green\n", "machine-recorded fields")
	refused("kind: spec\npatch:\n  add:\n    - atom: N1\n      row:\n        run: \"echo x\"\n    - atom: N1\n      row:\n        run: \"echo y\"\n", "twice")

	// the structural pairs on a two-atom corpus: a SPEC patch that
	// removes every row of one atom retires it (the other stands); a
	// RUN patch cannot empty a behavior — the expansion carries the
	// emptied atom and the door's own form rule refuses it
	two := "kind: spec\nfull-replace: true\nbehaviors:\n" +
		"  - id: B1\n    name: sample\n    prose: |\n      One paragraph.\n    rows:\n      - run: \"echo hello\"\n        out: \"hello\"\n        rc: 0\n        form: \"test corpus: the pinned surface\"\n" +
		"  - id: B2\n    name: second\n    prose: |\n      One paragraph.\n    rows:\n      - run: \"echo two\"\n        out: \"two\"\n        rc: 0\n        form: \"test corpus: the pinned surface\"\n"
	stack2, eng2 := newInst(t)
	reply, code = submit(t, stack2, eng2, two)
	mustAccept(t, reply, code, "FULL-REPLACE: 0 rows -> 2 rows")
	reply, code = submit(t, stack2, eng2, "kind: spec\npatch:\n  remove: [B2/1]\n")
	mustAccept(t, reply, code)
	if atoms, err := eng2.Store.Atoms(); err != nil || len(atoms) != 1 || atoms[0].ID != "B1" {
		t.Fatalf("the emptied atom did not retire: %v %v", atoms, err)
	}
	reply, code = submit(t, stack2, eng2, "kind: run\npatch:\n  remove: [B1/1]\n")
	mustReject(t, reply, code, "rows")
}

// The dry validator's own error shapes: the patch draft that also
// spells the hatch, the unreadable corpus zone, the corrupt atom
// file, and the plain-read branch points of the corpus walk.
func TestSpecCheckErrorShapes(t *testing.T) {
	stack, eng := newInst(t)
	reply, code := submit(t, stack, eng, specDoc("B1", "hello"))
	mustAccept(t, reply, code)
	draft := filepath.Join(eng.Workdir, "draft.yaml")
	check := func(wantLines int, wantText ...string) {
		t.Helper()
		lines, err := CheckDocument(eng.Workdir, filepath.Base(draft))
		if err != nil {
			t.Fatal(err)
		}
		if len(lines) != wantLines {
			t.Fatalf("want %d finding(s), got %v", wantLines, lines)
		}
		for _, w := range wantText {
			if !strings.Contains(lines[0], w) {
				t.Fatalf("the finding misses %q: %v", w, lines)
			}
		}
	}
	// a patch draft that also spells full-replace: one finding, the
	// composition law
	if err := stack.fs.WriteFile(draft, []byte("kind: spec\nfull-replace: true\npatch:\n  remove: [B1/1]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	check(1, "do not compose")
	// the unreadable corpus zone is a failure, not a finding (a plain
	// patch draft — the composition finding above returns earlier)
	if err := stack.fs.WriteFile(draft, []byte("kind: spec\npatch:\n  prose:\n    B1: |\n      fine\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stack.fs.readdirErr = errIO
	_, err := CheckDocument(eng.Workdir, filepath.Base(draft))
	if err == nil {
		t.Fatal("the unreadable corpus zone checked clean")
	}
	stack.fs.readdirErr = nil
	// a corpus zone that does not exist yet is the EMPTY corpus: the
	// patch draft's addresses find nothing — a finding, not a failure
	stack.fs.readdirErr = os.ErrNotExist
	lines, err := CheckDocument(eng.Workdir, filepath.Base(draft))
	if err != nil || len(lines) == 0 || !strings.Contains(lines[0], "B1") {
		t.Fatalf("the missing corpus zone must be the empty corpus: %v %v", err, lines)
	}
	stack.fs.readdirErr = nil
	// a non-yaml entry in the atoms zone is not an atom; a corrupt
	// atom file is a failure of the read-only walk
	notes := filepath.Join(eng.Store.Root(), "canon", "atoms", "notes.txt")
	if err := stack.fs.WriteFile(notes, []byte("not an atom\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := stack.fs.WriteFile(draft, []byte("kind: spec\npatch:\n  prose:\n    B1: |\n      fine\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	check(0)
	bad := filepath.Join(eng.Store.Root(), "canon", "atoms", "zz-bad.yaml")
	if err := stack.fs.WriteFile(bad, []byte(":: not yaml [\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := CheckDocument(eng.Workdir, filepath.Base(draft)); err == nil {
		t.Fatal("the corrupt atom file checked clean")
	}
	// the same corrupt corpus fails the door's patch path (the
	// writer monopoly never opens over a broken zone)
	if err := stack.fs.WriteFile(filepath.Join(eng.Workdir, door()), []byte("kind: spec\npatch:\n  prose:\n    B1: |\n      x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if reply, code := eng.SubmitSet(door()); code != Failure {
		t.Fatalf("the corrupt corpus did not fail the patch path: %d %s", code, reply)
	}
	// an atom file that cannot be READ at all is the same failure
	stack.fs.Remove(bad)
	stack.fs.readErr = errIO
	stack.fs.readErrPath = filepath.Join(eng.Store.Root(), "canon", "atoms", "b1.yaml")
	if _, err := CheckDocument(eng.Workdir, filepath.Base(draft)); err == nil {
		t.Fatal("the unreadable atom file checked clean")
	}
	stack.fs.readErr = nil
	stack.fs.readErrPath = ""
}

// atomDigests — the per-atom canon digests (which atoms moved).
func atomDigests(t *testing.T, eng *Engine) map[string]string {
	t.Helper()
	atoms, err := eng.Store.Atoms()
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]string{}
	for _, a := range atoms {
		out[a.ID] = a.Digest
	}
	return out
}
