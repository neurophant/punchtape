package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/engine"
	"github.com/neurophant/punchtape/internal/fsport"
)

// The cli unit runs the real machine over the real filesystem: the
// verbs own an instance in a fresh t.TempDir(), the working directory
// arrives over the package seam (seams.go), the retry pause is a no-op.
// Mock code lives in tests only — here, only the two seam variables.

// bindWD points the working-directory seam at dir; the retries pause
// for nothing.
func bindWD(t *testing.T, dir string) {
	t.Helper()
	oldWD, oldSleep := getwd, sleep
	getwd = func() (string, error) { return dir, nil }
	sleep = func(time.Duration) {}
	t.Cleanup(func() { getwd, sleep = oldWD, oldSleep })
}

// door — the one legal submission path inside the instance.
const door = ".punchtape/inbox/current.txt"

func writeDoor(t *testing.T, workdir, doc string) {
	t.Helper()
	path := filepath.Join(workdir, filepath.FromSlash(door))
	if err := os.WriteFile(path, []byte(doc), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestBareLaunch(t *testing.T) {
	var buf bytes.Buffer
	if code := Run(nil, &buf); code != 0 {
		t.Fatalf("a bare launch succeeds: %d", code)
	}
	if got := buf.String(); got != "punchtape "+Version+"\n" {
		t.Fatalf("the version line: %q", got)
	}
}

func TestUnknownVerb(t *testing.T) {
	bindWD(t, t.TempDir())
	var buf bytes.Buffer
	if code := Run([]string{"frobnicate"}, &buf); code != 2 {
		t.Fatalf("an unknown verb fails: %d", code)
	}
	out := buf.String()
	if !strings.Contains(out, `"frobnicate"`) || !strings.Contains(out, "next, submit") {
		t.Fatalf("the refusal names the verb and the two real ones: %q", out)
	}
}

func TestWorkdirFailure(t *testing.T) {
	old := getwd
	getwd = func() (string, error) { return "", errors.New("gone") }
	t.Cleanup(func() { getwd = old })
	var buf bytes.Buffer
	if code := Run([]string{"next"}, &buf); code != 2 {
		t.Fatalf("a broken working directory fails: %d", code)
	}
	if !strings.Contains(buf.String(), "working directory") {
		t.Fatalf("the failure names the working directory: %q", buf.String())
	}
}

func TestSubmitWithoutFile(t *testing.T) {
	bindWD(t, t.TempDir())
	var buf bytes.Buffer
	if code := Run([]string{"submit"}, &buf); code != 1 {
		t.Fatalf("submit without a file is a rejection: %d", code)
	}
	if !strings.Contains(buf.String(), door) {
		t.Fatalf("the rejection names the door: %q", buf.String())
	}
}

func TestSubmitWithoutInstance(t *testing.T) {
	bindWD(t, t.TempDir()) // a virgin directory
	var buf bytes.Buffer
	if code := Run([]string{"submit", door}, &buf); code != 1 {
		t.Fatalf("submit without an instance is a rejection: %d", code)
	}
	if !strings.Contains(buf.String(), "no instance") {
		t.Fatalf("the rejection says what is missing: %q", buf.String())
	}
}

func TestNextCreatesInstance(t *testing.T) {
	dir := t.TempDir()
	bindWD(t, dir)
	var buf bytes.Buffer
	if code := Run([]string{"next", "a tool that greps lines in files"}, &buf); code != 0 {
		t.Fatalf("next with a phrase in an empty dir creates the instance: %d %s", code, buf.String())
	}
	out := buf.String()
	// the contract injection comes first, before any slot content
	inject := canondata.T("cli.next.contract-inject")
	if !strings.HasPrefix(out, inject) {
		t.Fatalf("the first next injects the contract before anything: %q", out[:min(120, len(out))])
	}
	if !strings.Contains(out, engine.AgentsRefText()) {
		t.Fatal("the first next prints AGENTS.md in full")
	}
	// the instance skeleton exists
	for _, rel := range []string{
		filepath.Join(".punchtape", "state.yaml"),
		filepath.FromSlash(door),
		"AGENTS.md",
		filepath.Join(".punchtape", "reference", "grammar.md"),
	} {
		if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
			t.Fatalf("skeleton %s: %v", rel, err)
		}
	}
	s, exists, err := canon.Open(dir)
	if err != nil || !exists {
		t.Fatalf("canon open: %v %v", exists, err)
	}
	if st, err := s.State(); err != nil || st.Stage != canon.StageIntake {
		t.Fatalf("fresh state: %v %+v", err, st)
	}

	// a repeat next does not re-inject the contract
	buf.Reset()
	if code := Run([]string{"next"}, &buf); code != 0 {
		t.Fatalf("a repeat next succeeds: %d %s", code, buf.String())
	}
	if strings.Contains(buf.String(), "THE WORKING CONTRACT") {
		t.Fatal("only the first next injects the contract")
	}
	if !strings.Contains(buf.String(), "CONTRACT: AGENTS.md in your tree (unchanged; sha256 ") {
		t.Fatalf("the repeat next carries the one-line contract reference:\n%s", buf.String())
	}
	if buf.Len() == 0 {
		t.Fatal("a plain next answers the current slot")
	}
}

func TestNextWithBrief(t *testing.T) {
	dir := t.TempDir()
	bindWD(t, dir)
	var buf bytes.Buffer
	if code := Run([]string{"next", "make", "the", "echo", "tool", "sing"}, &buf); code != 0 {
		t.Fatalf("the birth next accepts the brief: %d %s", code, buf.String())
	}
	if !strings.Contains(buf.String(), "make the echo tool sing") {
		t.Fatalf("the slot shows the brief: %q", buf.String())
	}
	s, exists, err := canon.Open(dir)
	if err != nil || !exists {
		t.Fatalf("canon open: %v %v", exists, err)
	}
	st, err := s.State()
	if err != nil {
		t.Fatal(err)
	}
	if st.Wish == nil || *st.Wish != "make the echo tool sing" {
		t.Fatalf("the brief is fixed once: %+v", st.Wish)
	}

	// the brief is not edited through rebirth
	buf.Reset()
	if code := Run([]string{"next", "a different wish"}, &buf); code != 1 {
		t.Fatalf("a repeated brief is a rejection: %d", code)
	}
	if !strings.Contains(buf.String(), "already fixed") {
		t.Fatalf("the rejection explains: %q", buf.String())
	}
}

func TestSubmitNotTheDoorFile(t *testing.T) {
	dir := t.TempDir()
	bindWD(t, dir)
	if code := Run([]string{"next", "a tool that greps lines in files"}, &bytes.Buffer{}); code != 0 {
		t.Fatalf("instance: %d", code)
	}
	var buf bytes.Buffer
	if code := Run([]string{"submit", "notes.txt"}, &buf); code != 1 {
		t.Fatalf("a foreign path is a rejection: %d", code)
	}
	if !strings.Contains(buf.String(), door) {
		t.Fatalf("the rejection names the one legal door: %q", buf.String())
	}
}

func TestWriterBusyRefusal(t *testing.T) {
	dir := t.TempDir()
	bindWD(t, dir)
	if code := Run([]string{"next", "a tool that greps lines in files"}, &bytes.Buffer{}); code != 0 {
		t.Fatalf("instance: %d", code)
	}
	// hold the instance's write monopoly: opening then honestly refuses
	// after the short retries (the pause is a no-op seam)
	lockPath := filepath.Join(dir, ".punchtape", "cache", "writer.lock")
	release, err := fsport.Bind.LockExclusive(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	defer release()

	var buf bytes.Buffer
	if code := Run([]string{"next"}, &buf); code != 2 {
		t.Fatalf("a busy writer fails the verb: %d %s", code, buf.String())
	}
	if !strings.Contains(buf.String(), "writer") {
		t.Fatalf("the failure names the writer lock: %q", buf.String())
	}
	buf.Reset()
	if code := Run([]string{"submit", door}, &buf); code != 2 {
		t.Fatalf("a busy writer fails submit too: %d %s", code, buf.String())
	}
}

func TestInitFailureRefusal(t *testing.T) {
	dir := t.TempDir()
	bindWD(t, dir)
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })
	var buf bytes.Buffer
	if code := Run([]string{"next", "a tool that greps lines in files"}, &buf); code != 2 {
		t.Fatalf("a read-only workdir fails the instance birth: %d %s", code, buf.String())
	}
	if !strings.Contains(buf.String(), "FAIL") {
		t.Fatalf("the failure is honest: %q", buf.String())
	}
}

// The submission loop over the real engine: birth, spec, run — the
// two verbs of the contract driving one instance end to end.
func TestSubmissionLoop(t *testing.T) {
	dir := t.TempDir()
	bindWD(t, dir)
	var buf bytes.Buffer
	if code := Run([]string{"next", "a tiny echo tool"}, &buf); code != 0 {
		t.Fatalf("birth: %d %s", code, buf.String())
	}

	spec := `kind: spec
behaviors:
  - id: B1
    name: hello-row
    prose: |
      The row says hi.
    rows:
      - run: ["sh", "-c", "echo hi"]
        out: hi
        form: "the brief's greeting line"
`
	writeDoor(t, dir, spec)
	buf.Reset()
	if code := Run([]string{"submit", door}, &buf); code != 0 {
		t.Fatalf("spec submit: %d %s", code, buf.String())
	}
	if !strings.Contains(buf.String(), "ACCEPTED") {
		t.Fatalf("the spec submission is accepted: %q", buf.String())
	}
	// the spec moved the instance to implement
	s, _, err := canon.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if st, err := s.State(); err != nil || st.Stage != canon.StageImplement {
		t.Fatalf("stage after spec: %v %+v", err, st)
	}

	// the door is reset to the blank template after acceptance
	blank, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(door)))
	if err != nil {
		t.Fatal(err)
	}
	if string(blank) == spec {
		t.Fatal("an accepted draft must not stay in the door")
	}

	writeDoor(t, dir, "kind: run\n")
	buf.Reset()
	if code := Run([]string{"submit", door}, &buf); code != 0 {
		t.Fatalf("run submit: %d %s", code, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "ACCEPTED") {
		t.Fatalf("the run submission is accepted: %q", out)
	}
	if !strings.Contains(out, "suite=green") || !strings.Contains(out, "1 rows") {
		t.Fatalf("the run table reports the green row: %q", out)
	}

	// a red run row: the transaction applies, the table reports the red
	// (a fresh behavior id: the engine's current collision rule refuses
	// a repeated id — not this package's contract to test)
	writeDoor(t, dir, "kind: run\nbehaviors:\n  - id: B2\n    name: red-row\n    prose: |\n      A row that misses.\n    rows:\n      - run: [\"sh\", \"-c\", \"echo bye\"]\n        out: hi\n        form: \"the brief's greeting line\"\n")
	buf.Reset()
	if code := Run([]string{"submit", door}, &buf); code != 0 {
		t.Fatalf("a red run still applies (rejections are for forms): %d %s", code, buf.String())
	}
	out = buf.String()
	if !strings.Contains(out, "suite=red") || !strings.Contains(out, "B2/1:") {
		t.Fatalf("the run table reports the red row: %q", out)
	}

	// a broken form is a rejection and the draft stays in the door
	broken := "kind: spek\n"
	writeDoor(t, dir, broken)
	buf.Reset()
	if code := Run([]string{"submit", door}, &buf); code != 1 {
		t.Fatalf("a broken form is a rejection: %d %s", code, buf.String())
	}
	out = buf.String()
	if !strings.Contains(out, "kind: spec") || !strings.Contains(out, "draft stays") {
		t.Fatalf("the rejection hints the kind and keeps the draft: %q", out)
	}
	stayed, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(door)))
	if err != nil || string(stayed) != broken {
		t.Fatalf("a rejected draft stays for the rework: %v %q", err, stayed)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestPlainNextRenderFailure(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".punchtape"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".punchtape", "state.yaml"), []byte("schema-version: 8\nstage: [broken\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	old := getwd
	getwd = func() (string, error) { return dir, nil }
	t.Cleanup(func() { getwd = old })
	var buf bytes.Buffer
	if code := Run([]string{"next"}, &buf); code != 2 || !strings.Contains(buf.String(), "FAIL:") {
		t.Fatalf("broken state next: %d %q", code, buf.String())
	}
}

// Born-without-a-phrase: an instance born by a
// bare next in an empty directory carried no task phrase — the
// verdict belonged to no task, and one executor built the whole
// wrong library surface because the brief never reached the instance.
// Birth now REQUIRES the phrase: a bare next in a virgin directory
// is a named refusal and nothing is created; an EXISTING phraseless
// instance (born under the old machine) still reads — retro law.
func TestNextRequiresPhraseAtBirth(t *testing.T) {
	dir := t.TempDir()
	bindWD(t, dir)
	var buf bytes.Buffer
	if code := Run([]string{"next"}, &buf); code != 1 {
		t.Fatalf("a bare next in a virgin dir must refuse: %d %s", code, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "phrase") {
		t.Fatalf("the refusal names the missing phrase: %q", out)
	}
	for _, rel := range []string{
		filepath.Join(".punchtape", "state.yaml"),
		"AGENTS.md",
	} {
		if _, err := os.Stat(filepath.Join(dir, rel)); err == nil {
			t.Fatalf("the refused birth created %s — nothing may exist", rel)
		}
	}
	// the birth with the phrase creates the bound instance
	buf.Reset()
	if code := Run([]string{"next", "make", "the", "echo", "tool", "sing"}, &buf); code != 0 {
		t.Fatalf("next with a phrase births: %d %s", code, buf.String())
	}
	if !strings.HasPrefix(buf.String(), canondata.T("cli.next.contract-inject")) {
		t.Fatalf("the birth injects the contract first: %q", buf.String()[:min(120, buf.Len())])
	}
	s, exists, err := canon.Open(dir)
	if err != nil || !exists {
		t.Fatalf("canon open: %v %v", exists, err)
	}
	st, err := s.State()
	if err != nil || st.Wish == nil || *st.Wish != "make the echo tool sing" {
		t.Fatalf("the phrase is fixed in the state: %v %+v", err, st)
	}

	// retro: a phraseless instance born by the old machine reads on
	legacy := t.TempDir()
	bindWD(t, legacy)
	if _, err := engine.Init(legacy); err != nil {
		t.Fatalf("legacy init: %v", err)
	}
	buf.Reset()
	if code := Run([]string{"next"}, &buf); code != 0 {
		t.Fatalf("a bare next on a legacy phraseless instance reads: %d %s", code, buf.String())
	}
	if !strings.Contains(buf.String(), "STAGE:") {
		t.Fatalf("the legacy slot renders: %q", buf.String())
	}
}
