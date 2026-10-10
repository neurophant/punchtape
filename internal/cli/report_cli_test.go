// The battery of the operator's report — the machine
// lifts the observable-behavior table out of the stored canon runs,
// so the operator's verification aims at the delta instead of
// re-deriving the whole case by hand. Every test names its episode;
// all fail on the current code (the report verb does not exist).
package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neurophant/punchtape/internal/fsport"
)

// reportSpec — the listing/stable/refusal corpus of the report tests.
func reportSpec(listOut string) string {
	return "kind: spec\nbehaviors:\n" +
		"  - id: LIST\n    name: the run listing\n    prose: |\n      The listing names what the run wrote.\n" +
		"    rows:\n      - run: [echo, " + listOut + "]\n        out: \"" + listOut + "\"\n        rc: 0\n        form: \"corpus: the listing surface\"\n" +
		"  - id: HOLD\n    name: held behavior\n    prose: |\n      Stable rows stay green untouched.\n" +
		"    rows:\n      - run: [echo, held]\n        out: \"held\"\n        rc: 0\n        form: \"corpus: the pinned surface\"\n" +
		"  - id: REF\n    name: refusal rows\n    prose: |\n      Refusals answer and exit non-zero.\n" +
		"    rows:\n      - run: [sh, -c, \"exit 1\"]\n        out: \"\"\n        rc: 1\n        form: \"corpus: the pinned surface\"\n"
}

// reportBirth — an instance with a green v1 cycle (spec, run, ratify: v1).
func reportBirth(t *testing.T, dir string) {
	t.Helper()
	bindWD(t, dir)
	var buf bytes.Buffer
	if code := Run([]string{"next", "a tool that lists what it wrote"}, &buf); code != 0 {
		t.Fatalf("birth: %d %s", code, buf.String())
	}
	writeDoor(t, dir, reportSpec("s/index.html"))
	if code := Run([]string{"submit", filepath.FromSlash(door)}, &buf); code != 0 {
		t.Fatalf("spec v1: %d %s", code, buf.String())
	}
	writeDoor(t, dir, "kind: run\n")
	if code := Run([]string{"submit", filepath.FromSlash(door)}, &buf); code != 0 {
		t.Fatalf("run: %d %s", code, buf.String())
	}
	writeDoor(t, dir, "kind: run\nratify: true\n")
	if code := Run([]string{"submit", filepath.FromSlash(door)}, &buf); code != 0 {
		t.Fatalf("ratify v1: %d %s", code, buf.String())
	}
}

// reportRun — bind the working-directory seam and run the argv.
func reportRun(t *testing.T, dir string, args ...string) (string, int) {
	t.Helper()
	bindWD(t, dir)
	var buf bytes.Buffer
	code := Run(args, &buf)
	return buf.String(), code
}

// (1) The operator-price episode: the review side of the loop
// re-derived the whole observable behavior by hand every time — the
// most expensive share of the measured rounds. The report prints
// EVERY row from the stored runs — the exact command, the demand,
// the measured rc and output digest, the last run's time, the
// freshness mark — and re-runs nothing.
func TestReportPrintsEveryRowWithFreshness(t *testing.T) {
	dir := t.TempDir()
	reportBirth(t, dir)
	out, code := reportRun(t, dir, "report", "--full")
	if code != 0 {
		t.Fatalf("the report: %d %s", code, out)
	}
	for _, label := range []string{"LIST/1", "HOLD/1", "REF/1"} {
		if !strings.Contains(out, label) {
			t.Fatalf("the report does not list %s:\n%s", label, out)
		}
	}
	// the demand and the measured answer ride every row: the exact
	// command, the want, the rc of the last run, the output digest
	for _, want := range []string{
		"run: [echo, s/index.html]",
		"want:", "rc 0",
		"got: rc 1",
		"ran-at",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("the report loses %q:\n%s", want, out)
		}
	}
	// the freshness marks: every green row stands fresh at its inputs
	if !strings.Contains(out, "fresh") {
		t.Fatalf("the report carries no freshness mark:\n%s", out)
	}
	// the render is stored facts only: it never claims to run
	if !strings.Contains(out, "nothing re-ran") {
		t.Fatalf("the report does not name its law (stored runs, nothing re-ran):\n%s", out)
	}
	if strings.Contains(out, "suite=") {
		t.Fatalf("the report re-prints the run table (it is not a run):\n%s", out)
	}
}

// (2) The one-line-edit round episode: a measured round changed one
// listing call of the product; the report must show the stale row
// BEFORE any run re-measures it — the touched input marks its
// dependent rows stale in the table.
func TestTouchedFileMarksDependentRowStale(t *testing.T) {
	dir := t.TempDir()
	reportBirth(t, dir)
	// a row whose command names a tree file; green at "one"
	if err := os.WriteFile(filepath.Join(dir, "named.txt"), []byte("one\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeDoor(t, dir, frSpec(reportSpec("s/index.html"))+
		"  - id: FILE\n    name: file reader\n    prose: |\n      Reads the named file.\n"+
		"    rows:\n      - run: \"cat named.txt\"\n        out: \"one\"\n        rc: 0\n        form: \"corpus: the pinned surface\"\n")
	var buf bytes.Buffer
	if code := Run([]string{"submit", filepath.FromSlash(door)}, &buf); code != 0 {
		t.Fatalf("spec v2: %d %s", code, buf.String())
	}
	writeDoor(t, dir, "kind: run\n")
	if code := Run([]string{"submit", filepath.FromSlash(door)}, &buf); code != 0 {
		t.Fatalf("run: %d %s", code, buf.String())
	}
	// the named input moves: NO run happens — the report alone must
	// mark the dependent row stale
	if err := os.WriteFile(filepath.Join(dir, "named.txt"), []byte("two\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, code := reportRun(t, dir, "report", "--full")
	if code != 0 {
		t.Fatalf("the report: %d %s", code, out)
	}
	if !strings.Contains(out, "FILE/1") {
		t.Fatalf("the report does not list the dependent row:\n%s", out)
	}
	fileMark := false
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "FILE/1") && strings.Contains(line, "stale") {
			fileMark = true
		}
	}
	if !fileMark {
		t.Fatalf("the touched input does not mark FILE/1 stale:\n%s", out)
	}
	// the untouched rows stay fresh
	heldMark := false
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "HOLD/1") && strings.Contains(line, "fresh") {
			heldMark = true
		}
	}
	if !heldMark {
		t.Fatalf("the untouched HOLD/1 is not fresh:\n%s", out)
	}
}

// (3) The acceptance-gap episode: the operator accepted v1; the round
// changed one row's demand, added a row and retired another. The
// report --since v1 aims the verification at exactly that delta —
// the changed, the new and the retired rows by name, the untouched
// rows silent; an unknown verdict is an honest refusal.
func TestReportSinceVerdictNamesTheDelta(t *testing.T) {
	dir := t.TempDir()
	reportBirth(t, dir)
	// v2: LIST's demand changes, LATE arrives, HOLD retires
	writeDoor(t, dir, "kind: spec\nfull-replace: true\nbehaviors:\n"+
		"  - id: LIST\n    name: the run listing\n    prose: |\n      The listing names what the run wrote.\n"+
		"    rows:\n      - run: [echo, index.html]\n        out: \"index.html\"\n        rc: 0\n        form: \"corpus: the pinned surface\"\n"+
		"  - id: REF\n    name: refusal rows\n    prose: |\n      Refusals answer and exit non-zero.\n"+
		"    rows:\n      - run: [sh, -c, \"exit 1\"]\n        out: \"\"\n        rc: 1\n        form: \"corpus: the pinned surface\"\n"+
		"  - id: LATE\n    name: late row\n    prose: |\n      Arrived after the acceptance.\n"+
		"    rows:\n      - run: [echo, late]\n        out: \"late\"\n        rc: 0\n        form: \"corpus: the pinned surface\"\n")
	var buf bytes.Buffer
	if code := Run([]string{"submit", filepath.FromSlash(door)}, &buf); code != 0 {
		t.Fatalf("spec v2: %d %s", code, buf.String())
	}
	writeDoor(t, dir, "kind: run\n")
	if code := Run([]string{"submit", filepath.FromSlash(door)}, &buf); code != 0 {
		t.Fatalf("run v2: %d %s", code, buf.String())
	}
	out, code := reportRun(t, dir, "report", "--since", "v1")
	if code != 0 {
		t.Fatalf("the since-report: %d %s", code, out)
	}
	for _, want := range []string{"LIST/1", "LATE/1", "HOLD/1"} {
		if !strings.Contains(out, want) {
			t.Fatalf("the since-report does not name %s:\n%s", want, out)
		}
	}
	if strings.Contains(out, "REF/1") {
		t.Fatalf("the since-report re-prints the untouched REF/1:\n%s", out)
	}
	// the sections say WHICH movement each name is
	if !strings.Contains(out, "retired since v1") {
		t.Fatalf("the retired section does not name its verdict:\n%s", out)
	}
	// an unknown verdict is an honest refusal naming the latest
	unknown, code := reportRun(t, dir, "report", "--since", "v9")
	if code != 1 {
		t.Fatalf("an unknown verdict must refuse with rc 1, got %d: %s", code, unknown)
	}
	if !strings.Contains(unknown, "ERROR") {
		t.Fatalf("the unknown-verdict refusal is not an ERROR line:\n%s", unknown)
	}
}

// (4) The flag surface stays closed: --since rides report only; a
// valueless --since is a refusal; the unknown-flag text names the
// whole surface; report without an instance refuses honestly.
func TestReportFlagSurface(t *testing.T) {
	dir := t.TempDir()
	reportBirth(t, dir)
	if out, code := reportRun(t, dir, "submit", filepath.FromSlash(door), "--since", "v1"); code != 2 {
		t.Fatalf("--since on submit must refuse with rc 2, got %d: %s", code, out)
	}
	if out, code := reportRun(t, dir, "report", "--since"); code != 2 {
		t.Fatalf("a valueless --since must refuse with rc 2, got %d: %s", code, out)
	}
	var buf bytes.Buffer
	bindWD(t, dir)
	if code := Run([]string{"submit", filepath.FromSlash(door), "--bogus"}, &buf); code != 2 {
		t.Fatalf("an unknown flag must refuse with rc 2, got %d", code)
	}
	if !strings.Contains(buf.String(), "--since") {
		t.Fatalf("the unknown-flag text does not name the whole surface:\n%s", buf.String())
	}
	// no instance: the honest no-instance refusal
	empty := t.TempDir()
	noInst, code := reportRun(t, empty, "report")
	if code != 1 {
		t.Fatalf("report without an instance must refuse with rc 1, got %d: %s", code, noInst)
	}
	if !strings.Contains(noInst, "no instance") {
		t.Fatalf("the no-instance refusal does not explain:\n%s", noInst)
	}
	// the verb list names report
	bindWD(t, dir)
	buf.Reset()
	if code := Run([]string{"frobnicate"}, &buf); code != 2 {
		t.Fatalf("an unknown verb fails: %d", code)
	}
	if !strings.Contains(buf.String(), "report") {
		t.Fatalf("the verb list does not name report:\n%s", buf.String())
	}
	// --since rides report only: next and submit both refuse it
	if out, code := reportRun(t, dir, "next", "--since", "v1"); code != 2 || !strings.Contains(out, "--since") {
		t.Fatalf("--since on next must refuse with rc 2: %d %s", code, out)
	}
	// the report's own surface: the slot surface modes do not ride it
	if out, code := reportRun(t, dir, "report", "--quiet"); code != 2 || !strings.Contains(out, "--since") {
		t.Fatalf("report --quiet must refuse with rc 2: %d %s", code, out)
	}
	// a broken instance is the honest FAIL of the report verb
	broken := t.TempDir()
	if err := os.MkdirAll(filepath.Join(broken, ".punchtape"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(broken, ".punchtape", "state.yaml"),
		[]byte("schema-version: 8\nstage: [broken\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, code := reportRun(t, broken, "report"); code != 2 || !strings.Contains(out, "FAIL:") {
		t.Fatalf("the report over a broken instance: %d %s", code, out)
	}
	// a held write monopoly fails the report verb the same honest way
	lockPath := filepath.Join(dir, ".punchtape", "cache", "writer.lock")
	release, err := fsport.Bind.LockExclusive(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	if out, code := reportRun(t, dir, "report"); code != 2 || !strings.Contains(out, "writer") {
		t.Fatalf("the report under a busy writer: %d %s", code, out)
	}
}
