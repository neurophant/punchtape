// The battery of the quiet output and the batch call.
// Every machine call today re-prints every table — the measured
// executors re-read the full slot on each reply (the cache multiplier
// ~= x2 to the visible tokens on both listing-tool machine instances).
// The modes: --quiet (no slot body), --delta (only what changed since
// the last slot — the anchor is the last next reply, any mode; a
// submit ack never re-anchors it), --run-after (spec + run + status
// in ONE door call). The default full render for the human does not
// change.
// Every test names its episode; all fail on the current code (the
// flags do not exist and ride the argv as paths).
package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// frSpec — the deliberate full-rewrite spelling: a
// FULL document after the first accepted spec rides full-replace).
func frSpec(doc string) string {
	return strings.Replace(doc, "kind: spec\n", "kind: spec\nfull-replace: true\n", 1)
}

// quietSpec — a three-behavior spec over the baseline echo tool.
func quietSpec(listArg, listOut string) string {
	return "kind: spec\nbehaviors:\n" +
		"  - id: LIST\n    name: the run listing\n    prose: |\n      The listing names what the run wrote.\n" +
		"    rows:\n      - run: [echo, " + listArg + "]\n        out: \"" + listOut + "\"\n        rc: 0\n        form: \"corpus: the listing surface\"\n" +
		"  - id: HOLD\n    name: held behavior\n    prose: |\n      Stable rows stay green untouched.\n" +
		"    rows:\n      - run: [echo, held]\n        out: \"held\"\n        rc: 0\n        form: \"corpus: the pinned surface\"\n" +
		"  - id: REF\n    name: refusal rows\n    prose: |\n      Refusals answer and exit non-zero.\n" +
		"    rows:\n      - run: [sh, -c, \"exit 1\"]\n        out: \"\"\n        rc: 1\n        form: \"corpus: the pinned surface\"\n"
}

// quietBirth — an instance with a green v1 cycle (spec, run).
func quietBirth(t *testing.T, dir string) {
	t.Helper()
	bindWD(t, dir)
	if code := Run([]string{"next", "a tool that lists what it wrote"}, &bytes.Buffer{}); code != 0 {
		t.Fatalf("birth: %d", code)
	}
	writeDoor(t, dir, quietSpec("\"s/index.html\"", "s/index.html"))
	var buf bytes.Buffer
	if code := Run([]string{"submit", filepath.FromSlash(door)}, &buf); code != 0 {
		t.Fatalf("spec v1: %d %s", code, buf.String())
	}
	writeDoor(t, dir, "kind: run\n")
	if code := Run([]string{"submit", filepath.FromSlash(door)}, &buf); code != 0 {
		t.Fatalf("run: %d %s", code, buf.String())
	}
}

// (1) The cache-multiplier episode: the repeat run with ONE changed
// row in --delta prints only it — the untouched rows do not ride the
// reply again, the stage and the action still orient.
func TestDeltaRunPrintsOnlyTheChangedRow(t *testing.T) {
	dir := t.TempDir()
	quietBirth(t, dir)
	writeDoor(t, dir, frSpec(quietSpec("\"index.html\"", "index.html")))
	var buf bytes.Buffer
	if code := Run([]string{"submit", filepath.FromSlash(door)}, &buf); code != 0 {
		t.Fatalf("spec v2: %d %s", code, buf.String())
	}
	writeDoor(t, dir, "kind: run\n")
	buf.Reset()
	if code := Run([]string{"submit", filepath.FromSlash(door), "--delta"}, &buf); code != 0 {
		t.Fatalf("the delta run: %d %s", code, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "suite=green") {
		t.Fatalf("the delta reply loses the suite line:\n%s", out)
	}
	if strings.Contains(out, "HOLD/1") || strings.Contains(out, "REF/1") {
		t.Fatalf("the ack re-prints untouched rows:\n%s", out)
	}
	// the thin ack: the changed row and the orientation
	// ride the follow-up delta next
	buf.Reset()
	if code := Run([]string{"next", "--delta"}, &buf); code != 0 {
		t.Fatalf("the delta next: %d %s", code, buf.String())
	}
	if !strings.Contains(buf.String(), "LIST/1") || !strings.Contains(buf.String(), "STAGE:") || !strings.Contains(buf.String(), "ACTION:") {
		t.Fatalf("the delta next loses the changed row or the orientation:\n%s", buf.String())
	}
	// A delta next after nothing changed names nothing.
	buf.Reset()
	if code := Run([]string{"next", "--delta"}, &buf); code != 0 {
		t.Fatalf("the delta next: %d %s", code, buf.String())
	}
	if strings.Contains(buf.String(), "HOLD/1") || strings.Contains(buf.String(), "REF/1") || strings.Contains(buf.String(), "LIST/1") {
		t.Fatalf("the delta next re-prints unchanged rows:\n%s", buf.String())
	}
}

// (2) The quiet reply: the transaction's own block stays (the table,
// the verdict), the slot body does not ride.
func TestQuietDropsTheSlotBody(t *testing.T) {
	dir := t.TempDir()
	quietBirth(t, dir)
	writeDoor(t, dir, frSpec(quietSpec("\"index.html\"", "index.html")))
	var buf bytes.Buffer
	if code := Run([]string{"submit", filepath.FromSlash(door)}, &buf); code != 0 {
		t.Fatalf("spec v2: %d %s", code, buf.String())
	}
	writeDoor(t, dir, "kind: run\n")
	buf.Reset()
	if code := Run([]string{"submit", filepath.FromSlash(door), "--quiet"}, &buf); code != 0 {
		t.Fatalf("the quiet run: %d %s", code, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "suite=green") {
		t.Fatalf("the quiet reply loses the suite line:\n%s", out)
	}
	if strings.Contains(out, "CONTEXT:") || strings.Contains(out, "ATOM ") {
		t.Fatalf("the quiet ack carries the slot body:\n%s", out)
	}
	// the orientation lines ride the follow-up quiet next
	buf.Reset()
	if code := Run([]string{"next", "--quiet"}, &buf); code != 0 {
		t.Fatalf("the quiet next: %d %s", code, buf.String())
	}
	if !strings.Contains(buf.String(), "STAGE:") || !strings.Contains(buf.String(), "ACTION:") {
		t.Fatalf("the quiet next loses the orientation:\n%s", buf.String())
	}
	// The quiet next keeps the facts, drops the card.
	buf.Reset()
	if code := Run([]string{"next", "--quiet"}, &buf); code != 0 {
		t.Fatalf("the quiet next: %d %s", code, buf.String())
	}
	quiet := buf.String()
	if !strings.Contains(quiet, "STAGE:") || !strings.Contains(quiet, "ACTION:") {
		t.Fatalf("the quiet next loses the orientation:\n%s", quiet)
	}
	if strings.Contains(quiet, "executable truth") {
		t.Fatalf("the quiet next carries the atom card:\n%s", quiet)
	}
}

// (3) The round episode: spec + run + status in ONE door call — the
// reply carries the version diff AND the suite table AND the slot.
func TestBatchSpecRunStatusOneCall(t *testing.T) {
	dir := t.TempDir()
	quietBirth(t, dir)
	writeDoor(t, dir, frSpec(quietSpec("\"index.html\"", "index.html")))
	var buf bytes.Buffer
	if code := Run([]string{"submit", filepath.FromSlash(door), "--run-after", "--delta"}, &buf); code != 0 {
		t.Fatalf("the batch: %d %s", code, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "SPEC v2") {
		t.Fatalf("the batch reply loses the version diff:\n%s", out)
	}
	if !strings.Contains(out, "suite=green") {
		t.Fatalf("the batch reply loses the run table:\n%s", out)
	}
	if strings.Contains(out, "CONTEXT:") {
		t.Fatalf("the batch reply carries a slot body:\n%s", out)
	}
	if strings.Count(out, "ACCEPTED:") != 2 {
		t.Fatalf("the batch reply does not carry both transactions:\n%s", out)
	}
	// The door consumed the draft (the reset template may legitimately
	// carry sample ids — the real corpus ids must not ride it).
	doorBytes, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(door)))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(doorBytes), "id: LIST") {
		t.Fatalf("the door kept the consumed draft:\n%s", doorBytes)
	}
}

// quietRun — bind the working-directory seam to dir and run the argv.
func quietRun(t *testing.T, dir string, args ...string) (string, int) {
	t.Helper()
	bindWD(t, dir)
	var buf bytes.Buffer
	code := Run(args, &buf)
	return buf.String(), code
}

// (4) The representation law: the quiet and delta modes change
// NOTHING in the journal or the canon — the same cycle through the
// default and the quiet surface lands byte-identical machine state.
func TestQuietCanonJournalByteIdentical(t *testing.T) {
	plain := t.TempDir()
	quiet := t.TempDir()
	quietBirth(t, plain)
	quietBirth(t, quiet)
	// the same round: v2 + run + ratify, one default, one quiet+delta
	writeDoor(t, quiet, frSpec(quietSpec("\"index.html\"", "index.html")))
	if out, code := quietRun(t, quiet, "submit", filepath.FromSlash(door), "--run-after", "--quiet"); code != 0 {
		t.Fatalf("the quiet batch: %d %s", code, out)
	}
	writeDoor(t, plain, frSpec(quietSpec("\"index.html\"", "index.html")))
	if out, code := quietRun(t, plain, "submit", filepath.FromSlash(door)); code != 0 {
		t.Fatalf("the plain spec v2: %d %s", code, out)
	}
	writeDoor(t, plain, "kind: run\n")
	if out, code := quietRun(t, plain, "submit", filepath.FromSlash(door)); code != 0 {
		t.Fatalf("the plain run: %d %s", code, out)
	}
	// both close: run green, then the signature
	for _, dir := range []string{plain, quiet} {
		writeDoor(t, dir, "kind: run\nratify: true\n")
		if out, code := quietRun(t, dir, "submit", filepath.FromSlash(door), "--quiet"); code != 0 {
			t.Fatalf("the closing in %s: %d %s", dir, code, out)
		}
	}
	for _, rel := range []string{
		filepath.Join(".punchtape", "journal.yamll"),
		filepath.Join(".punchtape", "canon", "atoms", "list.yaml"),
		filepath.Join(".punchtape", "canon", "atoms", "hold.yaml"),
		filepath.Join(".punchtape", "canon", "atoms", "ref.yaml"),
		filepath.Join(".punchtape", "verdicts", "v1.txt"),
	} {
		a, err := os.ReadFile(filepath.Join(plain, rel))
		if err != nil {
			t.Fatalf("plain %s: %v", rel, err)
		}
		b, err := os.ReadFile(filepath.Join(quiet, rel))
		if err != nil {
			t.Fatalf("quiet %s: %v", rel, err)
		}
		// The clock is a fact of the run, not of the mode: the
		// comparison strips the time lines (wall-ms, recorded-at,
		// ran-at, at, bill) — every CONTENT byte must be identical.
		na, nb := stripClock(string(a)), stripClock(string(b))
		if na != nb {
			t.Fatalf("%s diverged between the default and the quiet surface:\n--- plain:\n%s\n--- quiet:\n%s", rel, na, nb)
		}
	}
}

// stripClock — blank the machine's time facts wherever they sit
// (top-level lines or inside quoted effect content): the recorded
// FACTS — definitions, outcomes, standings — are the byte-compat
// surface of the modes; the clock is a fact of the run, not of the
// mode. Inside a quoted effect scalar the timestamp arrives as
// `ran-at: \"TIMESTAMP\"` (escaped quotes) — the field regexps take
// the optional quotes; a digest computed OVER time-carrying row
// content (the atom digest covers ran-at) can only be blanked, it
// cannot be made time-free — the content the digest covers is
// still compared byte for byte.
func stripClock(s string) string {
	for _, re := range []*regexp.Regexp{
		regexp.MustCompile(`wall-ms: \d+`),
		regexp.MustCompile(`recorded-at: (\\"|")?20[0-9][^\n]*?Z(\\"|")?`),
		regexp.MustCompile(`ran-at: (\\"|")?20[0-9][^\n]*?Z(\\"|")?`),
		regexp.MustCompile(`at: (\\"|")?\d{4}-[^\n]*?Z(\\"|")?`),
		regexp.MustCompile(`bill: machine wall [^\n"]*`),
		regexp.MustCompile(`digest: [0-9a-f]{64}`),
	} {
		s = re.ReplaceAllString(s, "T")
	}
	return s
}

// (5) The flag surface is closed: an unknown flag is an honest
// refusal, not a file name.
func TestUnknownFlagRefused(t *testing.T) {
	dir := t.TempDir()
	quietBirth(t, dir)
	var buf bytes.Buffer
	if code := Run([]string{"submit", filepath.FromSlash(door), "--bogus"}, &buf); code != 2 {
		t.Fatalf("an unknown flag must refuse with rc 2, got %d: %s", code, buf.String())
	}
	if !strings.Contains(buf.String(), "unknown") {
		t.Fatalf("the refusal does not explain: %q", buf.String())
	}
}
