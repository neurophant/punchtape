// The machine's own walk through its states: the first contact, the
// empty wish slot, the birth, a green submission, the identical-run
// idleness, the staleness on a scope file's change, the completion
// act and the refused submission — over the real filesystem and the
// real launcher, in one temporary working root.
package application

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neurophant/punchtape/internal/domain/zone"
	"github.com/neurophant/punchtape/internal/infrastructure/fsys"
	"github.com/neurophant/punchtape/internal/infrastructure/process"
	"github.com/neurophant/punchtape/internal/ports"
)

func runIn(t *testing.T, root string) string {
	t.Helper()
	var term bytes.Buffer
	code := Run(root, &term)
	if code != 0 {
		t.Fatalf("run exited %d: %s", code, term.String())
	}
	return term.String()
}

func zoneFile(t *testing.T, root, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, zone.Dirname, name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return string(data)
}

func writeZone(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, zone.Dirname, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestMachineLife(t *testing.T) {
	ports.FSBind = fsys.OS{}
	ports.ExecBind = process.OS{}
	root := t.TempDir()

	// First contact: the contract prints, nothing is created.
	term := runIn(t, root)
	if !strings.Contains(term, "punchtape — the contract") {
		t.Fatalf("the contract must print, got: %q", term)
	}
	if _, err := os.Stat(filepath.Join(root, zone.Dirname)); !os.IsNotExist(err) {
		t.Fatalf("the first contact must create nothing")
	}

	// The empty wish slot says so and creates nothing.
	writeZone(t, root, ".keep", "")
	if err := os.Remove(filepath.Join(root, zone.Dirname, ".keep")); err != nil {
		t.Fatal(err)
	}
	if term = runIn(t, root); !strings.Contains(term, "no instance yet") {
		t.Fatalf("the empty slot must say so, got: %q", term)
	}

	// The birth: the wish is consumed, the zone stands.
	writeZone(t, root, zone.WishSlot, "a small tool, honestly checked")
	term = runIn(t, root)
	if !strings.Contains(term, "run 0") {
		t.Fatalf("the birth run must answer, got: %q", term)
	}
	if _, err := os.Stat(filepath.Join(root, zone.Dirname, zone.WishSlot)); !os.IsNotExist(err) {
		t.Fatalf("the wish slot must die at the birth")
	}
	for _, name := range []string{zone.ContractFile, zone.TemplateFile, zone.CorpusFile, zone.StateFile, zone.ReportFile} {
		if _, err := os.Stat(filepath.Join(root, zone.Dirname, name)); err != nil {
			t.Fatalf("the birth must write %s: %v", name, err)
		}
	}
	corpus := zoneFile(t, root, zone.CorpusFile)

	// A submission with one real row: green on the first execution.
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "input.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, `behaviors:
    - id: tool/echo
      name: The echo
      prose: Copies the input.
      scope:
        - src/**
      rows:
        - run: cat src/input.txt > out.txt
          expect:
            files:
                out.txt: "contains: hello"
`)
	runIn(t, root)
	if _, err := os.Stat(filepath.Join(root, zone.Dirname, zone.InboxSlot)); !os.IsNotExist(err) {
		t.Fatalf("an accepted submission consumes the inbox")
	}
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "run 1") || !strings.Contains(report, "green: 1") {
		t.Fatalf("the first run must be green, report:\n%s", report)
	}
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "tool/echo: 1 rows — 1 green") {
		t.Fatalf("the state must show green, state:\n%s", state)
	}

	// The identical run: nothing stale, nothing written.
	runIn(t, root)
	if zoneFile(t, root, zone.ReportFile) != report {
		t.Fatalf("an idle run must not rewrite the report")
	}

	// A scope file changes: exactly that behavior goes stale, runs
	// again, stays green.
	if err := os.WriteFile(filepath.Join(root, "src", "input.txt"), []byte("hello world"), 0o644); err != nil {
		t.Fatal(err)
	}
	runIn(t, root)
	if report = zoneFile(t, root, zone.ReportFile); !strings.Contains(report, "run 2") {
		t.Fatalf("a scope change must re-execute, report:\n%s", report)
	}

	// The completion act: the corpus's own bytes as the submission.
	writeZone(t, root, zone.InboxSlot, zoneFile(t, root, zone.CorpusFile))
	runIn(t, root)
	if _, err := os.Stat(filepath.Join(root, zone.Dirname, zone.InboxSlot)); !os.IsNotExist(err) {
		t.Fatalf("the completion act consumes the inbox")
	}
	if zoneFile(t, root, zone.CorpusFile) == corpus {
		t.Fatalf("the corpus must have grown past the birth's empty image")
	}

	// A refused submission: the refusals land in the report, the
	// inbox stays, the journal gains nothing.
	writeZone(t, root, zone.InboxSlot, `behaviors:
    - id: tool/bad
      name: Bad
      rows:
        - run: "true"
          expect:
            files:
                /abs/path: "contains: x"
`)
	runIn(t, root)
	if _, err := os.Stat(filepath.Join(root, zone.Dirname, zone.InboxSlot)); err != nil {
		t.Fatalf("a refused submission keeps the inbox")
	}
	if report = zoneFile(t, root, zone.ReportFile); !strings.Contains(report, "law:paths") {
		t.Fatalf("the refusal must cite the law, report:\n%s", report)
	}

	// A red row: the reason is grouped and named.
	writeZone(t, root, zone.InboxSlot, `behaviors:
    - id: tool/echo
      name: The echo
      prose: Copies the input.
      scope:
        - src/**
      rows:
        - run: cat src/input.txt > out.txt
          expect:
            out: "contains: NOPE"
`)
	runIn(t, root)
	if report = zoneFile(t, root, zone.ReportFile); !strings.Contains(report, `want "contains: NOPE"`) {
		t.Fatalf("the red fact must be named, report:\n%s", report)
	}
}

func TestRegistryCoversAndMarkersRideTheMachine(t *testing.T) {
	ports.FSBind = fsys.OS{}
	ports.ExecBind = process.OS{}
	root := t.TempDir()
	writeZone(t, root, zone.WishSlot, "one\n\ntwo\n\nthree\n")
	runIn(t, root)

	// A document with the registry: two requirements over the three
	// wish lines, one marker over the third, a behavior covering
	// both requirements.
	writeZone(t, root, zone.InboxSlot, `requirements:
    - id: R1
      brief: [W1..W2]
    - id: R2
      brief: [W3]
not-a-requirement:
    - lines: [W3]
      quote: three
      reason: the closing line names no work
behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies.
      scope:
        - src/**
      covers: [R1, R2]
      rows:
        - run: "true"
          expect:
            rc: 0
`)
	runIn(t, root)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "brief: 3/3 lines; requirements 2 standing, 2 covered; not-a-requirement: 1") {
		t.Fatalf("the brief block must count the standing registry, state:\n%s", state)
	}
	corpus := zoneFile(t, root, zone.CorpusFile)
	if !strings.Contains(corpus, "requirements:") || !strings.Contains(corpus, "R1") || !strings.Contains(corpus, "not-a-requirement:") {
		t.Fatalf("the corpus must carry the registry sections:\n%s", corpus)
	}

	// The idle run replays the registry from the journal alone.
	runIn(t, root)
	if zoneFile(t, root, zone.CorpusFile) != corpus {
		t.Fatalf("the fold must replay the registry byte for byte:\n%s", zoneFile(t, root, zone.CorpusFile))
	}

	// A range outside the wish: refused by the line it names.
	writeZone(t, root, zone.InboxSlot, `requirements:
    - id: R9
      brief: [W9]
`)
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "requirements[0].brief[0]") || !strings.Contains(report, "the wish carries no line W9") {
		t.Fatalf("a range outside the wish must be refused by name:\n%s", report)
	}

	// Retiring a requirement while a standing behavior still covers
	// it: refused by the dangling reference, the inbox stays.
	writeZone(t, root, zone.InboxSlot, "remove: [R2]\n")
	runIn(t, root)
	report = zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "tool/copy.covers[1]") || !strings.Contains(report, `no requirement "R2" stands`) {
		t.Fatalf("the leftover cover must be refused at its place:\n%s", report)
	}
	if _, err := os.Stat(filepath.Join(root, zone.Dirname, zone.InboxSlot)); err != nil {
		t.Fatalf("the refused removal keeps the inbox")
	}

	// The same retirement with the cover cleared in the same
	// submission: the requirement leaves, the brief stays whole.
	writeZone(t, root, zone.InboxSlot, `remove: [R2]
behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies.
      scope:
        - src/**
      covers: [R1]
      rows:
        - run: "true"
          expect:
            rc: 0
`)
	runIn(t, root)
	state = zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "requirements 1 standing, 1 covered") || !strings.Contains(state, "brief: 3/3 lines") {
		t.Fatalf("the retired requirement leaves the counts honest, state:\n%s", state)
	}
}
