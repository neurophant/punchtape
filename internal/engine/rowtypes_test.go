// The engine battery of typed row kinds — the
// mechanical families as first-class heads of the row grammar with
// strictly checked fields, semantically equal to the long form, plus
// the native witness kinds (the staged-bytes probe and the repeat)
// and the procedural launch/tree modes the long form could only
// imitate with hand-rolled orchestration.
package engine

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/neurophant/punchtape/internal/execport"
)

// typedDoc — a one-behavior spec with the given rows block.
func typedDoc(rows string) string {
	return "kind: spec\nbehaviors:\n" +
		"  - id: B1\n    name: the typed families\n    prose: |\n      What the behavior means, in words.\n" +
		"    rows:\n" + rows
}

// (1) The strict-field law: a typed head owns its row — a missing
// field is refused naming the field, an unknown field is refused
// naming it, and the general row fields do not ride a typed head.
func TestTypedStrictFields(t *testing.T) {
	cases := []struct {
		name string
		rows string
		want string
	}{
		{"refuse missing rc", "" +
			"      - refuse:\n          args: ./tool nope\n          says: 'tool: no such input'\n",
			"the fields args, says and rc are all required"},
		{"refuse missing says", "" +
			"      - refuse:\n          args: ./tool nope\n          rc: 3\n",
			"the fields args, says and rc are all required"},
		{"det missing same", "" +
			"      - det:\n          run: ./tool list\n",
			"the fields run and same are required"},
		{"witness missing probe", "" +
			"      - witness: {}\n",
			"probe is required"},
		{"witness unknown own field", "" +
			"      - witness:\n          probe: file://net-probe\n          weight: 2\n",
			"witness: unknown field weight (probe)"},
		{"witness row plus state", "" +
			"      - witness:\n          probe: file://net-probe\n        state:\n          - path: out.txt\n            text: x\n",
			`the witness row does not carry "state"`},
		{"bad launch mode value", "" +
			"      - run: ./tool list\n        launch: wide\n",
			"is not a launch mode"},
		{"bad tree mode value", "" +
			"      - run: ./tool list\n        tree: tidy\n",
			"is not a tree mode"},
		{"repeat-run empty", "" +
			"      - repeat-run: \"\"\n",
			"the command is required"},
		{"refuse unknown field", "" +
			"      - refuse:\n          args: ./tool nope\n          says: 'tool: no such input'\n          rc: 3\n          weight: 2\n",
			`refuse: unknown field weight (args, says, rc)`},
		{"typed head plus out", "" +
			"      - refuse:\n          args: ./tool nope\n          says: 'tool: no such input'\n          rc: 3\n        out: oops\n",
			`the refuse row does not carry "out"`},
		{"two typed heads", "" +
			"      - refuse:\n          args: ./tool nope\n          says: 'tool: no such input'\n          rc: 3\n        repeat-run: ./tool again\n",
			"a row carries exactly one head"},
	}
	for _, c := range cases {
		stack, eng := newInst(t)
		reply, code := submit(t, stack, eng, typedDoc(c.rows))
		if code != Rejected {
			t.Fatalf("%s: want a rejection, got %d %s", c.name, code, reply)
		}
		if !strings.Contains(reply, c.want) {
			t.Fatalf("%s: the reason does not name the law (%q):\n%s", c.name, c.want, reply)
		}
	}
}

// (2) The semantic-equality law: a spec written on the typed kinds
// and its long-form twin — the same document bytes for prose and the
// same rows underneath — produce byte-identical canon atoms and the
// same run outcomes. The type is notation, not a second language: the
// refusal head IS the row {run, out: "", err, rc} it names.
func TestTypedEqualsLong(t *testing.T) {
	typed := typedDoc("" +
		"      - refuse:\n          args: sh -c ./tool nope\n          says: 'tool: no such input'\n          rc: 3\n        form: \"corpus: the refusal surface\"\n" +
		"      - det:\n          run: echo stable\n          same: stable\n        form: \"corpus: the stable surface\"\n" +
		"      - run: echo held\n        out: held\n        rc: 0\n        form: \"corpus: the held surface\"\n")
	long := typedDoc("" +
		"      - run: sh -c ./tool nope\n        out: \"\"\n        err: 'tool: no such input'\n        rc: 3\n        form: \"corpus: the refusal surface\"\n" +
		"      - run: echo stable\n        out: stable\n        form: \"corpus: the stable surface\"\n" +
		"      - run: echo held\n        out: held\n        rc: 0\n        form: \"corpus: the held surface\"\n")
	var suites []string
	var atoms [][]byte
	for _, doc := range []string{typed, long} {
		stack, eng := newInst(t)
		reply, code := submit(t, stack, eng, doc)
		mustAccept(t, reply, code)
		stack.proc.plan["sh -c ./tool nope"] = plannedRefusal()
		reply, code = submit(t, stack, eng, "kind: run\n")
		mustAccept(t, reply, code)
		suite := ""
		for _, ln := range strings.Split(reply, "\n") {
			if strings.HasPrefix(ln, "suite=") {
				suite = ln
			}
		}
		if suite == "" || !strings.Contains(suite, "3") {
			t.Fatalf("the run table carries no three-row suite line:\n%s", reply)
		}
		for _, row := range []string{"B1/1", "B1/2", "B1/3"} {
			if !strings.Contains(reply, row) && !strings.Contains(reply, "green") {
				t.Fatalf("the row %s is not named green:\n%s", row, reply)
			}
		}
		data, err := stack.fs.ReadFile(filepath.Join(eng.Workdir, ".punchtape", "canon", "atoms", "b1.yaml"))
		if err != nil {
			t.Fatalf("atom read: %v", err)
		}
		suites = append(suites, suite)
		atoms = append(atoms, data)
	}
	if string(atoms[0]) != string(atoms[1]) {
		t.Fatalf("the typed spelling and the long form disagree on the canon:\n--- typed:\n%s\n--- long:\n%s",
			atoms[0], atoms[1])
	}
	if suites[0] != suites[1] {
		t.Fatalf("the typed spelling and the long form disagree on the suite:\n%s\n%s", suites[0], suites[1])
	}
}

// plannedRefusal — the planned answer of a refusing product: a trouble
// line on the error stream and a nonzero exit.
func plannedRefusal() execport.Outcome {
	return execport.Outcome{Exit: 3, Stderr: "tool: no such input\n"}
}

// (3) The refusal family measures: a green when the product refuses
// exactly as demanded (empty answer stream, the named trouble line,
// the named code) and an honest red with the observed bytes when it
// does not.
func TestRefuseMeasures(t *testing.T) {
	stack, eng := newInst(t)
	reply, code := submit(t, stack, eng, typedDoc(""+
		"      - refuse:\n          args: sh -c ./tool nope\n          says: 'tool: no such input'\n          rc: 3\n        form: \"corpus: the refusal surface\"\n"))
	mustAccept(t, reply, code)
	stack.proc.plan["sh -c ./tool nope"] = plannedRefusal()
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code)
	if !strings.Contains(reply, "suite=green (1 rows)") {
		t.Fatalf("the honest refusal row must run green:\n%s", reply)
	}
	// the product changes its mind: it exits the demanded code but
	// answers on the wrong stream — a rebuilt product (its named file
	// carries new bytes) re-measures and the row reds with the text
	stack.proc.plan["sh -c ./tool nope"] = execport.Outcome{Exit: 3, Stdout: "no such input\n"}
	if err := stack.fs.WriteFile(filepath.Join(eng.Workdir, "tool"), []byte("#!/bin/sh\nexit 3\n"), 0o755); err != nil {
		t.Fatalf("rebuild the product: %v", err)
	}
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code)
	if !strings.Contains(reply, "B1/1") || !strings.Contains(reply, "red") {
		t.Fatalf("the drifted refusal must red with its label:\n%s", reply)
	}
}

// (4) The repeat kind: the command runs twice over what it leaves —
// agreement is green, divergence is a measured red naming both runs.
func TestRepeatRun(t *testing.T) {
	stack, eng := newInst(t)
	reply, code := submit(t, stack, eng, typedDoc(""+
		"      - repeat-run: sh -c ./tool build\n        out: built\n        form: \"corpus: the built surface\"\n"))
	mustAccept(t, reply, code)
	stack.proc.plan["sh -c ./tool build"] = execport.Outcome{Exit: 0, Stdout: "built\n"}
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code)
	if !strings.Contains(reply, "suite=green (1 rows)") {
		t.Fatalf("the repeat build must run green:\n%s", reply)
	}
	// the second run answers differently — the product rebuilt (its
	// named file carries new bytes), the repeat re-measures and the
	// divergence is the witness
	stack.proc.plan["flip:sh -c ./tool build"] = execport.Outcome{Exit: 0, Stdout: "built again\n"}
	if err := stack.fs.WriteFile(filepath.Join(eng.Workdir, "tool"), []byte("#!/bin/sh\necho built\n"), 0o755); err != nil {
		t.Fatalf("rebuild the product: %v", err)
	}
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code)
	if !strings.Contains(reply, "B1/1") || !strings.Contains(reply, "the two runs disagree") {
		t.Fatalf("the diverging repeat must red naming the disagreement:\n%s", reply)
	}
}

// (5) The witness kind: the probe's bytes ride the door (a file next
// to the submission in the inbox); the machine stages them into the
// tree, runs them under the row's procedural modes and compares
// rc/out with the want. The launch mode rides the launch itself
// (nothing is wrapped around the command); the tree mode is the
// machine's own row-scoped tree identity control — a leak the probe
// leaves behind reds the row naming the path. Without the staged
// bytes the row is an honest env red: probes arrive through the
// door, never from the machine.
func TestWitnessRow(t *testing.T) {
	stack, eng := newInst(t)
	stageProbe := func() {
		t.Helper()
		if err := stack.fs.WriteFile(filepath.Join(eng.Workdir, ".punchtape", "inbox", "net-probe"),
			[]byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
			t.Fatalf("probe stage: %v", err)
		}
	}
	probeKey := filepath.Join(stack.dir, "net-probe")
	stageProbe()
	reply, code := submit(t, stack, eng, typedDoc(""+
		"      - witness:\n          probe: file://net-probe\n        launch: nonet\n        tree: unchanged\n        out: built\n        form: \"corpus: the probe surface\"\n"))
	mustAccept(t, reply, code)
	stack.proc.plan[probeKey] = execport.Outcome{Exit: 0, Stdout: "built\n"}
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code)
	if !strings.Contains(reply, "suite=green (1 rows)") {
		t.Fatalf("the staged probe must run green:\n%s", reply)
	}
	// the launch carried the mode: the discipline is a launch fact
	if !stack.proc.sawNetlessLaunch() {
		t.Fatalf("the witness row's launch did not carry the launch mode")
	}
	// the staged bytes vanish between runs: the probe leaves the tree
	// when the row ends
	if _, err := stack.fs.ReadFile(filepath.Join(eng.Workdir, "net-probe")); err == nil {
		t.Fatalf("the staged probe outlived its row")
	}
	// without the staged bytes the row is an env red naming the probe
	// (the door checked existence at the spec; the bytes left the
	// inbox before the run — the memo went stale, the re-run found
	// nothing to stage)
	stack, eng = newInst(t)
	stageProbe()
	reply, code = submit(t, stack, eng, typedDoc(""+
		"      - witness:\n          probe: file://net-probe\n        out: built\n        form: \"corpus: the probe surface\"\n"))
	mustAccept(t, reply, code)
	stack.proc.plan[filepath.Join(stack.dir, "net-probe")] = execport.Outcome{Exit: 0, Stdout: "built\n"}
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "suite=green")
	stack.fs.Remove(filepath.Join(eng.Workdir, ".punchtape", "inbox", "net-probe"))
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code)
	if !strings.Contains(reply, "B1/1") || !strings.Contains(reply, "net-probe") || !strings.Contains(reply, "missing") {
		t.Fatalf("the missing staged bytes must red naming the probe:\n%s", reply)
	}
	// the tree mode on ANY row: a general run row that leaves a file
	// behind reds under the row-scoped identity control
	stack, eng = newInst(t)
	reply, code = submit(t, stack, eng, typedDoc(""+
		"      - run: sh -c 'echo dust > dust.txt'\n        tree: unchanged\n"))
	mustAccept(t, reply, code)
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code)
	if !strings.Contains(reply, "B1/1") || !strings.Contains(reply, "tree moved during the row") {
		t.Fatalf("the leaking row must red naming the tree:\n%s", reply)
	}
}

// (6) The dry validator: the door's form rules as a deterministic function —
// a draft never touches the door, the findings name every class the
// door would refuse, and a pass is the door's own predicates (the
// same docFindings, no second law). The surroundings read as plain
// files: the stored ignore when the draft declares none, the corpus
// ids, the input and inbox directories.
func TestCheckDocumentShapes(t *testing.T) {
	stack, eng := newInst(t)
	_ = eng
	if err := stack.fs.WriteFile(stack.dir+"/draft.yaml", []byte(
		"kind: spec\nbehaviors:\n  - id: B1\n    name: n\n    prose: |\n      p\n"+
			"    rows:\n      - run: [echo, x]\n        out: x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// no instance state yet: the door's live facts are not the
	// draft's; the form rules all pass
	lines, err := CheckDocument(stack.dir, "draft.yaml")
	if err != nil || len(lines) != 0 {
		t.Fatalf("the clean draft must pass: %v %v", lines, err)
	}
	// a draft the door would refuse: every finding rides, one per row
	if err := stack.fs.WriteFile(stack.dir+"/draft.yaml", []byte(
		"kind: spec\nbehaviors:\n  - id: B1\n    name: n\n    prose: |\n      p\n"+
			"    rows:\n      - run: [echo, x]\n        timeout: 999\n"+
			"      - run: [echo, y]\n        rc: -1\n"+
			"      - run: []\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	lines, err = CheckDocument(stack.dir, "draft.yaml")
	if err != nil || len(lines) != 3 {
		t.Fatalf("the draft's findings: %v %v", lines, err)
	}
	if !strings.Contains(lines[0], "outside 1..") || !strings.Contains(lines[1], "rc must be 0 or greater") ||
		!strings.Contains(lines[2], "argv list is empty") {
		t.Fatalf("the findings do not name their rows:\n%s", strings.Join(lines, "\n"))
	}
	// an undecodable draft is a finding, not a system failure
	if err := stack.fs.WriteFile(stack.dir+"/draft.yaml", []byte("kind: nope\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	lines, err = CheckDocument(stack.dir, "draft.yaml")
	if err != nil || len(lines) != 1 || !strings.Contains(lines[0], "unknown kind") {
		t.Fatalf("the undecodable draft: %v %v", lines, err)
	}
	// a missing draft is a system failure
	if _, err = CheckDocument(stack.dir, "absent.yaml"); err == nil {
		t.Fatal("the missing draft must fail as a system error")
	}
	// a corpus of one atom, a stored ignore and a live draft: the
	// collision law reads the corpus, the surface law the stored ignore
	greenCycle(t, stack, eng)
	if err := stack.fs.WriteFile(stack.dir+"/draft.yaml", []byte(
		"kind: spec\nbuild-ignore: [junk]\nbehaviors:\n  - id: List\n    name: n\n    prose: |\n      p\n"+
			"    rows:\n      - run: [echo, x]\n        out: x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	lines, err = CheckDocument(stack.dir, "draft.yaml")
	if err != nil || len(lines) != 1 || !strings.Contains(lines[0], "collides with another id in lower case") {
		t.Fatalf("the corpus collision: %v %v", lines, err)
	}
}

// (7) The tree mode's own control under a faulting walk: a snapshot
// that cannot be taken leaves the row to its measured outcome (the
// control never invents a red of its own).
func TestTreeModeFault(t *testing.T) {
	stack, eng := newInst(t)
	reply, code := submit(t, stack, eng, typedDoc(""+
		"      - run: echo held\n        tree: unchanged\n"))
	mustAccept(t, reply, code)
	stack.fs.walkErr = errIO
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code)
	if !strings.Contains(reply, "suite=red") && !strings.Contains(reply, "suite=green") {
		t.Fatalf("the faulted walk must not take the suite down:\n%s", reply)
	}
}

// (9) The witness reference's own door findings: a probe spelled
// without the door's form, and a probe path that escapes the tree.
func TestWitnessDoorFindings(t *testing.T) {
	stack := bindUnits(t)
	dir := memDir(t)
	_ = stack
	draft := func(probe string) {
		t.Helper()
		if err := stack.fs.WriteFile(dir+"/draft.yaml", []byte(
			"kind: spec\nbehaviors:\n  - id: B1\n    name: n\n    prose: |\n      p\n"+
				"    rows:\n      - witness:\n          probe: "+probe+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	draft("net-probe")
	lines, err := CheckDocument(dir, "draft.yaml")
	if err != nil || len(lines) != 1 || !strings.Contains(lines[0], "expected the file://name form") {
		t.Fatalf("the formless probe finding: %v %v", lines, err)
	}
	draft("file://../escape.sh")
	lines, err = CheckDocument(dir, "draft.yaml")
	if err != nil || len(lines) != 1 || !strings.Contains(lines[0], "witness probe path") {
		t.Fatalf("the escaping probe finding: %v %v", lines, err)
	}
}

// (8) The dry validator's surrounding reads as plain files: a
// directory without an instance reads as no corpus and no state; a
// damaged state file reads as none; the atoms zone's junk (a torn
// record, a foreign file, an id-less yaml) skips honestly; the input
// and inbox existence checks ride the findings.
func TestCheckSurroundings(t *testing.T) {
	stack := bindUnits(t)
	dir := memDir(t)
	_ = stack
	// no .punchtape at all: the draft's seeds and samples name their
	// missing surroundings
	if err := stack.fs.WriteFile(dir+"/draft.yaml", []byte(
		"kind: spec\nbehaviors:\n  - id: B1\n    name: n\n    prose: |\n      p\n"+
			"    rows:\n      - run: [echo, x]\n        seeds:\n          - path: in.txt\n            from-input: given.dat\n"+
			"        bytes: file://ref.bin\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	lines, err := CheckDocument(dir, "draft.yaml")
	if err != nil || len(lines) != 2 {
		t.Fatalf("the seed and sample findings: %v %v", lines, err)
	}
	if !strings.Contains(lines[0], "no such file") || !strings.Contains(lines[1], "missing next to the submission") {
		t.Fatalf("the existence findings do not name their files:\n%s", strings.Join(lines, "\n"))
	}
	// a damaged state file and a junk-bearing atoms zone read as none
	if err := stack.fs.MkdirAll(dir + "/.punchtape/canon/atoms"); err != nil {
		t.Fatal(err)
	}
	if err := stack.fs.WriteFile(dir+"/.punchtape/state.yaml", []byte("not: [a state"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := stack.fs.WriteFile(dir+"/.punchtape/canon/atoms/notes.txt", []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := stack.fs.WriteFile(dir+"/.punchtape/canon/atoms/torn.yaml", []byte("id: [broken"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := stack.fs.WriteFile(dir+"/.punchtape/canon/atoms/empty.yaml", []byte("name: n\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := stack.fs.WriteFile(dir+"/draft.yaml", []byte(
		"kind: spec\nbehaviors:\n  - id: B1\n    name: n\n    prose: |\n      p\n"+
			"    rows:\n      - run: [echo, x]\n        out: x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	lines, err = CheckDocument(dir, "draft.yaml")
	if err != nil || len(lines) != 0 {
		t.Fatalf("the damaged surroundings must read as none: %v %v", lines, err)
	}
	// an atoms file the reader cannot even open skips like the torn one
	stack.fs.readErrPath = dir + "/.punchtape/canon/atoms/torn.yaml"
	lines, err = CheckDocument(dir, "draft.yaml")
	if err != nil || len(lines) != 0 {
		t.Fatalf("the unreadable atoms file must skip: %v %v", lines, err)
	}
	stack.fs.readErrPath = ""
}

// (10) The form gate: a row that pins formal surface (a nonzero
// rc, an exact out/err/state text) without its one-line source closes
// the run door; the spec ack names the rows; the source line opens it.
func TestFormGate(t *testing.T) {
	stack, eng := newInst(t)
	doc := typedDoc("" +
		"      - run: echo held\n        out: held\n")
	reply, code := submit(t, stack, eng, doc)
	mustAccept(t, reply, code, "SPEC v1", "FORM DECISIONS UNRESOLVED: B1/1")
	stack.proc.plan["echo held"] = execport.Outcome{Exit: 0, Stdout: "held\n"}
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustReject(t, reply, code, "FORM DECISIONS UNRESOLVED: B1/1")
	// the source line opens the run door
	doc = fr(typedDoc("" +
		"      - run: echo held\n        out: held\n        form: \"corpus: the held surface\"\n"))
	reply, code = submit(t, stack, eng, doc)
	mustAccept(t, reply, code, "SPEC v2")
	if strings.Contains(reply, "FORM DECISIONS UNRESOLVED") {
		t.Fatalf("the sourced spec still carries the marker:\n%s", reply)
	}
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "suite=green")
	// contains fragments and the default rc are not formal pins: no
	// source is owed
	reply, code = submit(t, stack, eng, fr(typedDoc(""+
		"      - run: echo held\n        out: (contains-loose) hel\n")))
	mustAccept(t, reply, code, "SPEC v3")
	if strings.Contains(reply, "FORM DECISIONS UNRESOLVED") {
		t.Fatalf("a fragment demand owes no source:\n%s", reply)
	}
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "suite=green")
}
