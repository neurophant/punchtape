// The CLI battery of the dry validator — the door's
// form rules as a deterministic function BEFORE the door, with one reason line
// per finding; a document that passes the check is taken by the door
// on the first try. Every rejection class the door can teach used to
// cost a paid submit-rejected-reread cycle; the check names them all
// in one pass.
package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// checkBirth — an instance in a fresh directory.
func checkBirth(t *testing.T, dir string) {
	t.Helper()
	bindWD(t, dir)
	var buf bytes.Buffer
	if code := Run([]string{"next", "a tool that lists what it wrote"}, &buf); code != 0 {
		t.Fatalf("birth: %d %s", code, buf.String())
	}
}

// checkRun — bind the working directory and run the argv.
func checkRun(t *testing.T, dir string, args ...string) (string, int) {
	t.Helper()
	bindWD(t, dir)
	var buf bytes.Buffer
	code := Run(args, &buf)
	return buf.String(), code
}

// checkWrite — a file in the work directory.
func checkWrite(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, name)), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

// (1) The five-rejection episode: a document carrying one error of
// each class the door refuses — a self-report echo, an unmarked short
// fragment, a machine-recorded field, a seed naming a file the input
// directory does not carry, a run naming a file outside the accounted
// surface. The door taught them one paid cycle at a time; the check
// names all five in one pass, each with its reason line.
func TestCheckNamesEveryFinding(t *testing.T) {
	dir := t.TempDir()
	checkBirth(t, dir)
	checkWrite(t, dir, "fixtures/x.dat", "fixture bytes\n")
	doc := "kind: spec\n" +
		"build-ignore: [fixtures]\n" +
		"behaviors:\n" +
		"  - id: B1\n    name: echo row\n    prose: |\n      One behavior.\n" +
		"    rows:\n" +
		"      - run: [./tool, a]\n        out: PASS\n" +
		"  - id: B2\n    name: loose row\n    prose: |\n      One behavior.\n" +
		"    rows:\n" +
		"      - run: [./tool, b]\n        out: (contains) x9\n" +
		"  - id: B3\n    name: forged row\n    prose: |\n      One behavior.\n" +
		"    rows:\n" +
		"      - run: [./tool, c]\n        outcome: green\n" +
		"  - id: B4\n    name: seed row\n    prose: |\n      One behavior.\n" +
		"    rows:\n" +
		"      - run: [./tool, d]\n        seeds:\n          - path: in.txt\n            from-input: missing.dat\n" +
		"  - id: B5\n    name: ignored row\n    prose: |\n      One behavior.\n" +
		"    rows:\n" +
		"      - run: [./tool, fixtures/x.dat]\n"
	checkWrite(t, dir, "spec-draft.yaml", doc)
	out, code := checkRun(t, dir, "spec", "--check", "spec-draft.yaml")
	if code != 1 {
		t.Fatalf("the five-error document must refuse with rc 1, got %d:\n%s", code, out)
	}
	for _, want := range []string{
		"echoes a self-report line",
		"shorter than",
		"machine-recorded fields",
		"has no such file",
		"outside the accounted surface",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("the check loses the finding %q:\n%s", want, out)
		}
	}
	if !strings.Contains(out, "B5/1") {
		t.Fatalf("the findings name their rows:\n%s", out)
	}
}

// (2) The pass guarantee: a document the check passes is taken by the
// door on the first try — the same predicates, one vocabulary.
func TestCheckPassFirstTry(t *testing.T) {
	dir := t.TempDir()
	checkBirth(t, dir)
	doc := "kind: spec\n" +
		"behaviors:\n" +
		"  - id: B1\n    name: refusal row\n    prose: |\n      One behavior.\n" +
		"    rows:\n" +
		"      - refuse:\n          args: [sh, -c, \"exit 3\"]\n          says: trouble\n          rc: 3\n"
	checkWrite(t, dir, "spec-draft.yaml", doc)
	out, code := checkRun(t, dir, "spec", "--check", "spec-draft.yaml")
	if code != 0 || !strings.Contains(out, "PASS") {
		t.Fatalf("the clean document must pass the check: %d %s", code, out)
	}
	// the same bytes through the door: accepted on the first try
	writeDoor(t, dir, doc)
	out, code = checkRun(t, dir, "submit", filepath.FromSlash(door))
	if code != 0 {
		t.Fatalf("the passed document must be accepted first try: %d %s", code, out)
	}
}

// (3) The closed surface: --check rides the spec verb alone; the spec
// verb carries --check alone.
func TestCheckSurface(t *testing.T) {
	dir := t.TempDir()
	checkBirth(t, dir)
	for _, argv := range [][]string{
		{"next", "--check", "x"},
		{"submit", "--check", "x"},
		{"report", "--check", "x"},
	} {
		out, code := checkRun(t, dir, argv...)
		if code != 2 || !strings.Contains(out, "--check") {
			t.Fatalf("%v: want rc 2 naming --check, got %d %s", argv, code, out)
		}
	}
	out, code := checkRun(t, dir, "spec")
	if code != 2 || !strings.Contains(out, "--check") {
		t.Fatalf("bare spec: want rc 2 naming --check, got %d %s", code, out)
	}
	out, code = checkRun(t, dir, "spec", "--check")
	if code != 2 {
		t.Fatalf("valueless --check: want rc 2, got %d %s", code, out)
	}
	out, code = checkRun(t, dir, "spec", "--quiet")
	if code != 2 {
		t.Fatalf("spec with a foreign flag: want rc 2, got %d %s", code, out)
	}
	// a missing draft is a system failure, not a finding
	out, code = checkRun(t, dir, "spec", "--check", "absent.yaml")
	if code != 2 || !strings.Contains(out, "reading") {
		t.Fatalf("the missing draft: want rc 2 naming the read, got %d %s", code, out)
	}
}
