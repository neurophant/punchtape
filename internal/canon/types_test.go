// The unit battery of the row grammar: the typed
// heads' decode — the strict field sets, the head exclusivity, the
// decode-into-the-long-form law for refuse and det, the native
// heads for witness and repeat-run, and the procedural launch/tree
// modes' strict values.
package canon

import (
	"strings"
	"testing"

	"github.com/neurophant/punchtape/internal/yamlio"
)

// typedRow — decode one row body into a Row.
func typedRow(t *testing.T, body string) (Row, error) {
	t.Helper()
	var row Row
	err := yamlio.DecodeDoc([]byte(body+"\n"), &row)
	return row, err
}

// typedRowErr — the row decode must refuse with the wanted reason.
func typedRowErr(t *testing.T, body, want string) {
	t.Helper()
	_, err := typedRow(t, body)
	if err == nil {
		t.Fatalf("the row body %q decoded; want the refusal %q", body, want)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("the refusal of %q does not name the law %q: %v", body, want, err)
	}
}

// The strict field sets: every typed head refuses its missing and
// unknown fields by name, the row refuses a second head, and the
// general form refuses a typed field.
func TestTypedHeadsStrict(t *testing.T) {
	typedRowErr(t, "refuse:\n  args: ./t x\n  says: no\n", "the fields args, says and rc are all required")
	typedRowErr(t, "refuse:\n  args: []\n  says: no\n  rc: 2\n", "the fields args, says and rc are all required")
	typedRowErr(t, "refuse:\n  args: ./t x\n  says: no\n  rc: -1\n", "0 or greater")
	typedRowErr(t, "refuse:\n  args: ./t x\n  rc: 2\n", "the fields args, says and rc are all required")
	typedRowErr(t, "refuse: yes", "expected its field map")
	typedRowErr(t, "refuse:\n  args: ./t x\n  says: no\n  rc: 2\n  weight: 1\n", "unknown field weight (args, says, rc)")
	typedRowErr(t, "det:\n  run: ./t\n", "the fields run and same are required")
	typedRowErr(t, "det:\n  same: out\n", "the fields run and same are required")
	typedRowErr(t, "det:\n  run: []\n  same: out\n", "the fields run and same are required")
	typedRowErr(t, "det:\n  run: ./t\n  same: \"\"\n", "the fields run and same are required")
	typedRowErr(t, "witness:\n  run: ./t\n", "unknown field run (probe)")
	typedRowErr(t, "witness: yes", "expected its field map")
	typedRowErr(t, "witness:\n  probe: \"\"\n", "probe is required")
	typedRowErr(t, "run: ./t\nlaunch: wide\n", "is not a launch mode")
	typedRowErr(t, "run: ./t\ntree: tidy\n", "is not a tree mode")
	typedRowErr(t, "repeat-run: \"\"", "the command is required")
	typedRowErr(t, "repeat-run: []", "the command is required")
	typedRowErr(t, "refuse:\n  args: ./t x\n  says: no\n  rc: 2\ndet:\n  run: ./t\n  same: s\n",
		"a row carries exactly one head")
	typedRowErr(t, "out: x\n", "a row carries exactly one head")
	typedRowErr(t, "run: ./t\nweight: 1\n", "the row's fields are the run command")
	typedRowErr(t, "run: ./t\nrefuse:\n  args: ./t x\n  says: no\n  rc: 2\n", `the refuse row does not carry "run"`)
	// the row node itself must be a mapping, and the typed fields
	// carry their own types
	typedRowErr(t, "42", "where mapping is expected")
	typedRowErr(t, "refuse:\n  args: ./t x\n  says: no\n  rc: abc\n", "cannot unmarshal")
}

// The decode-into-the-long-form law: a refuse head IS the row it
// spells, a det head IS its recorded-expectation row; the command
// keeps both of its own forms.
func TestTypedDecodeIntoLongForm(t *testing.T) {
	row, err := typedRow(t, "refuse:\n  args: [./tool, nope]\n  says: trouble line\n  rc: 7\ntimeout: 90\n")
	if err != nil {
		t.Fatal(err)
	}
	if row.Refuse != nil || row.Run.Line != "" || !row.Run.IsArgv || len(row.Run.Argv) != 2 {
		t.Fatalf("the refusal head did not become the row's command: %+v", row)
	}
	if row.Out == nil || row.Out.Text != "" || row.Err == nil || row.Err.Text != "trouble line" {
		t.Fatalf("the refusal head's expectations are wrong: %+v", row)
	}
	if row.Rc == nil || *row.Rc != 7 || row.Timeout != 90 {
		t.Fatalf("the refusal head's code and ceiling are wrong: %+v", row)
	}
	row, err = typedRow(t, "refuse:\n  args: ./tool nope\n  says: trouble line\n  rc: 7\n")
	if err != nil {
		t.Fatal(err)
	}
	if row.Run.Line != "./tool nope" || row.Run.IsArgv {
		t.Fatalf("the string command form did not survive the head: %+v", row.Run)
	}
	row, err = typedRow(t, "det:\n  run: ./tool list\n  same: one line\n")
	if err != nil {
		t.Fatal(err)
	}
	if row.Det != nil || row.Run.Line != "./tool list" || row.Out == nil || row.Out.Text != "one line" || row.Rc != nil {
		t.Fatalf("the determinism head did not become its recorded-expectation row: %+v", row)
	}
}

// The native heads survive the decode and marshal back to their
// own spelling; the row's one command answers through CommandFields
// and CmdForm whichever head carries it; the procedural modes ride
// every head and round-trip with it.
func TestNativeWitnessHeads(t *testing.T) {
	row, err := typedRow(t, "witness:\n  probe: file://net-probe.sh\nlaunch: nonet\ntree: unchanged\nout: probe-ok\nrc: 0\n")
	if err != nil {
		t.Fatal(err)
	}
	if row.Witness == nil || row.Witness.Probe != "file://net-probe.sh" {
		t.Fatalf("the witness head did not survive: %+v", row)
	}
	if row.Launch != LaunchModeNonet || row.Tree != TreeModeUnchanged {
		t.Fatalf("the procedural modes did not survive: %+v", row)
	}
	if ref := row.ProbeRef(); ref != "net-probe.sh" {
		t.Fatalf("the probe reference decode: %q", ref)
	}
	if fields, ferr := row.CommandFields(); ferr != nil || len(fields) != 1 || fields[0] != "./net-probe.sh" {
		t.Fatalf("the witness row's command fields: %v %v", fields, ferr)
	}
	if got := row.CmdForm(); got != "./net-probe.sh" {
		t.Fatalf("the witness row's command face: %q", got)
	}
	again, err := yamlio.Marshal(row)
	if err != nil {
		t.Fatal(err)
	}
	var back Row
	if err := yamlio.DecodeDoc(again, &back); err != nil {
		t.Fatalf("the witness row does not round-trip: %v\n%s", err, again)
	}
	if back.Witness == nil || back.Launch != LaunchModeNonet || back.Tree != TreeModeUnchanged {
		t.Fatalf("the round-tripped witness row lost its knobs: %+v", back)
	}
	row, err = typedRow(t, "repeat-run: [./tool, again]\nout: again\n")
	if err != nil {
		t.Fatal(err)
	}
	if row.Repeat == nil || len(row.Repeat.Cmd.Argv) != 2 {
		t.Fatalf("the repeat head did not survive: %+v", row)
	}
	if fields, ferr := row.CommandFields(); ferr != nil || len(fields) != 2 {
		t.Fatalf("the repeat row's command fields: %v %v", fields, ferr)
	}
	if got := row.CmdForm(); !strings.Contains(got, "./tool") {
		t.Fatalf("the repeat row's command face: %q", got)
	}
	again, err = yamlio.Marshal(row)
	if err != nil {
		t.Fatal(err)
	}
	var back2 Row
	if err := yamlio.DecodeDoc(again, &back2); err != nil || back2.Repeat == nil {
		t.Fatalf("the repeat row does not round-trip: %v\n%s", err, again)
	}
}

// The pin predicate: a nonzero rc, an exact stream or state text
// are formal pins; fragments and the default rc are not; a pin
// without its source line is unresolved.
func TestFormPins(t *testing.T) {
	rc0, rc2 := 0, 2
	if (canonRow("run: ./t\n")).FormPins() {
		t.Fatal("a bare row pins nothing")
	}
	if (canonRow("run: ./t\nrc: 0\n")).FormPins() {
		t.Fatal("the default rc is not a pin")
	}
	if !(canonRow("run: ./t\nrc: 2\n")).FormPins() {
		t.Fatal("a nonzero rc is a pin")
	}
	if !(canonRow("run: ./t\nout: \"\"\n")).FormPins() {
		t.Fatal("the exact empty stream is a pin")
	}
	if (canonRow("run: ./t\nout: (contains) frag here\n")).FormPins() {
		t.Fatal("a fragment is not a pin")
	}
	if !(canonRow("run: ./t\nstate:\n  - path: f\n    text: v\n")).FormPins() {
		t.Fatal("an exact state text is a pin")
	}
	if (canonRow("run: ./t\nstate:\n  - path: f\n    from-run: g\n")).FormPins() {
		t.Fatal("a from-run state check pins no text")
	}
	pinned := canonRow("run: ./t\nrc: 2\n")
	if !pinned.FormUnresolved() {
		t.Fatal("a pin without a source is unresolved")
	}
	sourced := canonRow("run: ./t\nrc: 2\nform: \"the brief's code\"\n")
	if sourced.FormUnresolved() {
		t.Fatal("a sourced pin is resolved")
	}
	_ = rc0
	_ = rc2
}

// canonRow — decode one row body (the battery's own spelling).
func canonRow(body string) Row {
	var row Row
	if err := yamlio.DecodeDoc([]byte(body+"\n"), &row); err != nil {
		panic(err)
	}
	return row
}
