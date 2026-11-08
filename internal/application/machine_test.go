// The machine's own walk through its states: the first contact, the
// empty wish slot, the birth, a green submission, the identical-run
// idleness, the staleness on a claimed line's change, the completion
// act and the refused submission — over the real filesystem and the
// real launcher, in one temporary working root.
package application

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/neurophant/punchtape/internal/application/canon"
	"github.com/neurophant/punchtape/internal/domain/doc"
	"github.com/neurophant/punchtape/internal/domain/journal"
	"github.com/neurophant/punchtape/internal/domain/zone"
	"github.com/neurophant/punchtape/internal/infrastructure/fsys"
	"github.com/neurophant/punchtape/internal/infrastructure/process"
	"github.com/neurophant/punchtape/internal/infrastructure/store"
	"github.com/neurophant/punchtape/internal/infrastructure/store/cas"
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

	// The birth: the stream survives it, the zone stands.
	writeZone(t, root, zone.DemandsSlot, "a small tool, honestly checked")
	term = runIn(t, root)
	if !strings.Contains(term, "run 0") {
		t.Fatalf("the birth run must answer, got: %q", term)
	}
	if data, err := os.ReadFile(filepath.Join(root, zone.Dirname, zone.DemandsSlot)); err != nil || string(data) != "a small tool, honestly checked" {
		t.Fatalf("the stream survives the birth, open forever: %v %q", err, data)
	}
	for _, name := range []string{zone.ContractFile, zone.TemplateFile, zone.ReferenceFile, zone.CorpusFile, zone.StateFile, zone.ReportFile} {
		if _, err := os.Stat(filepath.Join(root, zone.Dirname, name)); err != nil {
			t.Fatalf("the birth must write %s: %v", name, err)
		}
	}
	corpus := zoneFile(t, root, zone.CorpusFile)

	// A submission with one real pinned row: green on the first
	// execution.
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "input.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, fmt.Sprintf(`
work: the first document of the echo
behaviors:
    - id: tool/echo
      name: The echo
      prose: Copies the input, pinned by its bytes.
      claims:
        - src/input.txt:1-1
      rows:
        - run: sh -c 'printf "%%s\n" %s; cat src/input.txt > out.txt'
          expect:
            files:
                out.txt: hello
`, claimsTokenOf(t, root, "src/input.txt:1-1")))
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

	// A claimed line changes: exactly that behavior goes stale, runs
	// again — and the pin catches the drift as red.
	if err := os.WriteFile(filepath.Join(root, "src", "input.txt"), []byte("hello world"), 0o644); err != nil {
		t.Fatal(err)
	}
	runIn(t, root)
	if report = zoneFile(t, root, zone.ReportFile); !strings.Contains(report, "run 2") {
		t.Fatalf("a claimed line's change must re-execute, report:\n%s", report)
	}
	if !strings.Contains(report, `want "hello"`) {
		t.Fatalf("the pinned row must catch the drift, report:\n%s", report)
	}

	// The completion act: the corpus's own bytes as the submission.
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if _, err := os.Stat(filepath.Join(root, zone.Dirname, zone.InboxSlot)); !os.IsNotExist(err) {
		t.Fatalf("the completion act consumes the inbox")
	}
	if zoneFile(t, root, zone.CorpusFile) == corpus {
		t.Fatalf("the corpus must have grown past the birth's empty image")
	}

	// A refused submission: the refusals land in the report, the
	// inbox stays, the journal gains nothing.
	writeZone(t, root, zone.InboxSlot, `
work: a broken document
behaviors:
    - id: tool/bad
      name: Bad
      prose: P.
      rows:
        - run: "true"
          expect:
            rc: 0
            files:
                /abs/path: x
`)
	runIn(t, root)
	if _, err := os.Stat(filepath.Join(root, zone.Dirname, zone.InboxSlot)); err != nil {
		t.Fatalf("a refused submission keeps the inbox")
	}
	if report = zoneFile(t, root, zone.ReportFile); !strings.Contains(report, "law:paths") {
		t.Fatalf("the refusal must cite the law, report:\n%s", report)
	}

	// A red row: the reason is grouped and named.
	writeZone(t, root, zone.InboxSlot, fmt.Sprintf(`
work: the honest expectation heals the echo
mandate: [{line: D1, target: tool/echo}]
behaviors:
    - id: tool/echo
      name: The echo
      prose: Copies the input, pinned by its bytes.
      claims:
        - src/input.txt:1-1
      rows:
        - run: sh -c 'printf "%%s\n" %s; cat src/input.txt > out.txt'
          expect:
            files:
                out.txt: hello world
`, claimsTokenOf(t, root, "src/input.txt:1-1")))
	runIn(t, root)
	if report = zoneFile(t, root, zone.ReportFile); !strings.Contains(report, "green: 1") {
		t.Fatalf("the healed pin must stand green, report:\n%s", report)
	}
}

func TestRegistryCoversAndMarkersRideTheMachine(t *testing.T) {
	ports.FSBind = fsys.OS{}
	ports.ExecBind = process.OS{}
	root := t.TempDir()
	writeZone(t, root, zone.DemandsSlot, "one\n\ntwo\n\nthree\n\nfour\n")
	runIn(t, root)

	// A document with the registry: two requirements over the first
	// three demand lines, one marker over the fourth, a behavior
	// covering both requirements.
	claims := claimsTokenOf(t, root)
	writeZone(t, root, zone.InboxSlot, fmt.Sprintf(`
work: the registry's first account
intents:
    - {lines: [D1..D2], outcome: work, prose: the pair that names the work}
    - {lines: [D3], outcome: work, prose: the closing line asks the same work}
    - {lines: [D4], outcome: mark, prose: the last line names no order}
requirements:
    - id: R1
      lines: [D1..D2]
      prose: the two working lines
    - id: R2
      lines: [D3]
      prose: the closing line
not-a-requirement:
    - lines: [D4]
      reason: the closing line names no work
behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies.
      rows:
        - run: sh -c 'printf "one %%s %%s %%s\n" %s %s %s'
          expect:
            out: "one %s %s %s\n"
      covers: [R1, R2]
`, claims, demandTokenOf("one\ntwo"), demandTokenOf("three"), claims, demandTokenOf("one\ntwo"), demandTokenOf("three")))
	runIn(t, root)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "demands: 4/4 lines; requirements 2 standing, 2 covered; markers: 1") {
		t.Fatalf("the demands block must count the standing registry, state:\n%s", state)
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

	// A range outside the stream: refused by the line it names.
	writeZone(t, root, zone.InboxSlot, `
work: a range too far
requirements:
    - id: R9
      lines: [D9]
      prose: p
`)
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "requirements[0].lines[0]") || !strings.Contains(report, "the range names line D9 past the stream's 4 lines") {
		t.Fatalf("a range outside the stream must be refused by name:\n%s", report)
	}

	// Retiring a requirement while a standing behavior still covers
	// it: refused by the dangling reference, the inbox stays.
	writeZone(t, root, zone.InboxSlot, "work: retiring R2\nmandate: [{line: D3, target: R2}]\nremove: [R2]\n")
	runIn(t, root)
	report = zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "tool/copy.covers[1]") || !strings.Contains(report, `no requirement "R2" stands`) {
		t.Fatalf("the leftover cover must be refused at its place:\n%s", report)
	}
	if _, err := os.Stat(filepath.Join(root, zone.Dirname, zone.InboxSlot)); err != nil {
		t.Fatalf("the refused removal keeps the inbox")
	}

	// The same retirement with the cover cleared in the same
	// submission: the requirement leaves, the stream stays whole.
	writeZone(t, root, zone.InboxSlot, fmt.Sprintf(`
work: retiring R2 with its cover cleared
mandate: [{line: D3, target: R2}]
remove: [R2]
not-a-requirement:
    - lines: [D3]
      reason: retired with its requirement
    - lines: [D4]
      reason: the closing line names no work
behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies.
      rows:
        - run: sh -c 'printf "one %%s %%s %%s\n" %s %s %s'
          expect:
            out: "one %s %s %s\n"
      covers: [R1]
`, claims, demandTokenOf("one\ntwo"), demandTokenOf("three"), claims, demandTokenOf("one\ntwo"), demandTokenOf("three")))
	runIn(t, root)
	state = zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "requirements 1 standing, 1 covered") || !strings.Contains(state, "demands: 4/4 lines") {
		t.Fatalf("the retired requirement leaves the counts honest, state:\n%s", state)
	}
}

func TestNotesRefsResolveThroughTheMachine(t *testing.T) {
	ports.FSBind = fsys.OS{}
	ports.ExecBind = process.OS{}
	root := newRoot(t)
	// A note citing a printed law and a standing demand line: the
	// submission stands.
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "input.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	echo := fmt.Sprintf(`
work: a noted document
notes:
    - kind: decision
      text: the law citation resolves
      refs: [D1, law:done]
behaviors:
    - id: tool/echo
      name: The echo
      prose: Copies the input, pinned by its bytes.
      claims:
        - src/input.txt:1-1
      rows:
        - run: sh -c 'printf "%%s\n" %s; cat src/input.txt > out.txt'
          expect:
            files:
                out.txt: hello
`, claimsTokenOf(t, root, "src/input.txt:1-1"))
	writeZone(t, root, zone.InboxSlot, echo)
	runIn(t, root)
	// A note citing a law the contract does not print: refused by name.
	writeZone(t, root, zone.InboxSlot, `
work: a cited law that is not printed
mandate: [{line: D1, target: tool/echo}]
notes:
    - kind: decision
      text: the made-up rule
      refs: [law:invented]
behaviors:
    - id: tool/echo
      name: The echo
      prose: Copies the input, pinned by its bytes.
      claims:
        - src/input.txt:1-1
      rows:
        - run: cat src/input.txt > out.txt
          expect:
            files:
                out.txt: hello
`)
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, `notes[0].refs[0]: "law:invented" resolves to nothing`) {
		t.Fatalf("an unprinted law citation must be refused by name:\n%s", report)
	}
}

func TestUncoveredLinesRideTheMachine(t *testing.T) {
	ports.FSBind = fsys.OS{}
	ports.ExecBind = process.OS{}
	root := newRoot(t)
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "input.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "gen.txt"), []byte("g1\ng2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// An exclusion of a real hole — lines no behavior claims — stands.
	writeZone(t, root, zone.InboxSlot, fmt.Sprintf(`
work: the generated file stands excluded, loudly
uncovered-lines:
    - {range: "gen.txt:1-2", reason: generated code, refs: [D1]}
behaviors:
    - id: tool/echo
      name: The echo
      prose: Copies the input, pinned by its bytes.
      claims:
        - src/input.txt:1-1
      rows:
        - run: sh -c 'printf "%%s\n" %s; cat src/input.txt > out.txt'
          expect:
            files:
                out.txt: hello
`, claimsTokenOf(t, root, "src/input.txt:1-1")))
	runIn(t, root)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "tool/echo: 1 rows — 1 green") {
		t.Fatalf("the submission with a loud exclusion must stand, state:\n%s", state)
	}
	// An exclusion of claimed-only lines: a mark that marks nothing.
	writeZone(t, root, zone.InboxSlot, `
work: the empty mark
mandate: [{line: D1, target: tool/echo}]
uncovered-lines:
    - {range: "src/input.txt:1-1", reason: nothing to mark, refs: [D1]}
behaviors:
    - id: tool/echo
      name: The echo
      prose: Copies the input, pinned by its bytes.
      claims:
        - src/input.txt:1-1
      rows:
        - run: cat src/input.txt > out.txt
          expect:
            files:
                out.txt: hello
`)
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "uncovered-lines[0]: the range names no uncovered line — a mark that marks nothing is refused") {
		t.Fatalf("the empty mark must be refused by name:\n%s", report)
	}
}

// TestContractPrintsWithoutJournal — the contract prints whenever no
// journal stands: a zone directory with pins but no birth is still
// the first contact; with words the same run births the instance.
func TestContractPrintsWithoutJournal(t *testing.T) {
	ports.FSBind = fsys.OS{}
	ports.ExecBind = process.OS{}
	root := t.TempDir()
	writeZone(t, root, zone.EnvSlot, "PTA=one\n")
	var term strings.Builder
	code := Run(root, &term)
	if code != 0 || !strings.Contains(term.String(), "punchtape — the contract") {
		t.Fatalf("the pinned zone without a journal must see the contract: %d %q", code, term.String())
	}
	if !strings.Contains(term.String(), "no instance yet — put the human's words verbatim into "+zone.Path(zone.DemandsSlot)) {
		t.Fatalf("the empty stream must be named: %q", term.String())
	}
	// With words the same first contact births: the contract prints,
	// the instance begins.
	writeZone(t, root, zone.DemandsSlot, "a checked tool")
	term.Reset()
	code = Run(root, &term)
	if code != 0 || !strings.Contains(term.String(), "punchtape — the contract") {
		t.Fatalf("the birth run must open with the contract: %d %q", code, term.String())
	}
	if _, err := os.Stat(filepath.Join(root, zone.Dirname, zone.JournalDir)); err != nil {
		t.Fatalf("the words must birth the instance: %v", err)
	}
}

// TestForeignZoneSoftRefusal — the foreign version's refusal is one
// line, the archive recipe inside it.
func TestForeignZoneSoftRefusal(t *testing.T) {
	root := newBornRoot(t)
	records, err := journalRecords(t, root)
	if err != nil {
		t.Fatal(err)
	}
	first := epochPaths(t, root)[0]
	birth := journal.Record{Kind: journal.KindBirth, Birth: &journal.Birth{Version: "0.3.0"}}
	var b strings.Builder
	b.WriteString("---\n")
	b.Write(journal.MarshalRecord(birth))
	b.WriteString("---\n")
	b.Write(journal.MarshalRecord(records[1]))
	if err := os.WriteFile(first, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	var term strings.Builder
	code := Run(root, &term)
	if code != 1 {
		t.Fatalf("the foreign zone must refuse, exit %d: %s", code, term.String())
	}
	want := "punchtape: zone born by machine 0.3.0 — this machine is " + Version +
		"; archive it (cp -r .punchtape .punchtape-archive) and start fresh\n"
	if term.String() != want {
		t.Fatalf("the refusal is exactly one line with the recipe inside:\nwant: %q\ngot:  %q", want, term.String())
	}
}

// TestZoneCarriesOnlyEnglishContract — the zone carries the English
// contract alone: the mirrors are human data in the machine's own
// source, never files of the zone.
func TestZoneCarriesOnlyEnglishContract(t *testing.T) {
	ports.FSBind = fsys.OS{}
	ports.ExecBind = process.OS{}
	root := t.TempDir()
	writeZone(t, root, zone.DemandsSlot, "a checked tool")
	runIn(t, root)
	if _, err := os.Stat(filepath.Join(root, zone.Dirname, zone.ContractFile)); err != nil {
		t.Fatalf("the birth must write the English contract: %v", err)
	}
	for _, name := range []string{"contract-ru.md", "contract-zh.md"} {
		if _, err := os.Stat(filepath.Join(root, zone.Dirname, name)); !os.IsNotExist(err) {
			t.Fatalf("the zone carries no mirror, found %s", name)
		}
	}
}

// TestTerminalOneLineEverywhere — the terminal's one-line law: every
// branch of Run answers with exactly one line. The contract page of
// the first contact — no zone, the empty wish slot, the birth run —
// is the one declared exception, and it rides with exactly one
// further line. The foreign version's one-line refusal stands pinned
// by its own test.
func TestTerminalOneLineEverywhere(t *testing.T) {
	ports.FSBind = fsys.OS{}
	ports.ExecBind = process.OS{}
	root := t.TempDir()
	page := canon.Load().Contract

	// First contact: the page alone, nothing after it.
	var term strings.Builder
	if code := Run(root, &term); code != 0 || term.String() != page {
		t.Fatalf("the first contact is the page alone: %d %q", code, term.String())
	}

	// The empty wish slot: the page plus one line.
	writeZone(t, root, ".keep", "")
	if err := os.Remove(filepath.Join(root, zone.Dirname, ".keep")); err != nil {
		t.Fatal(err)
	}
	term.Reset()
	if code := Run(root, &term); code != 0 {
		t.Fatalf("the empty slot must answer, exit %d: %s", code, term.String())
	}
	pagePlusOneLine(t, term.String(), page)

	// The birth: the page plus the run's one line.
	writeZone(t, root, zone.DemandsSlot, "a checked tool")
	term.Reset()
	if code := Run(root, &term); code != 0 {
		t.Fatalf("the birth must answer, exit %d: %s", code, term.String())
	}
	pagePlusOneLine(t, term.String(), page)

	// An ordinary idle run: one line.
	term.Reset()
	if code := Run(root, &term); code != 0 {
		t.Fatalf("the idle run must answer, exit %d: %s", code, term.String())
	}
	assertOneLine(t, term.String())

	// A refused submission: one line — a refusal is data, not an error.
	writeZone(t, root, zone.InboxSlot, "\tbroken: [")
	term.Reset()
	if code := Run(root, &term); code != 0 {
		t.Fatalf("the refusal must answer as data, exit %d: %s", code, term.String())
	}
	assertOneLine(t, term.String())

	// The damaged past: one line, exit 1.
	writeZone(t, root, zone.DemandsSlot, "A CHECKED TOOL")
	term.Reset()
	if code := Run(root, &term); code != 1 {
		t.Fatalf("the touched past must fail the run, exit %d: %s", code, term.String())
	}
	assertOneLine(t, term.String())
}

// pagePlusOneLine — the contract page opens the answer and exactly
// one line rides after it.
func pagePlusOneLine(t *testing.T, term, page string) {
	t.Helper()
	if !strings.HasPrefix(term, page) {
		t.Fatalf("the contract page must open the answer:\n%s", term)
	}
	rest := strings.TrimSuffix(strings.TrimPrefix(term, page), "\n")
	if rest == "" || strings.Contains(rest, "\n") {
		t.Fatalf("the page rides with exactly one further line, got %q", rest)
	}
}

// assertOneLine — the terminal answer is exactly one line.
func assertOneLine(t *testing.T, term string) {
	t.Helper()
	if line := strings.TrimSuffix(term, "\n"); line == "" || strings.Contains(line, "\n") {
		t.Fatalf("the terminal answer is exactly one line, got %q", term)
	}
}

// TestTmpClearedEveryRun — the machine's own temporary directory is
// the machine's to clear: whatever a row leaves in .punchtape/tmp is
// gone when the run ends, every run, and the agent never cleans up
// inside its own commands.
func TestTmpClearedEveryRun(t *testing.T) {
	ports.FSBind = fsys.OS{}
	ports.ExecBind = process.OS{}
	root := newBornRoot(t)
	tmp := filepath.Join(root, zone.Dirname, zone.TmpDir)
	if err := os.MkdirAll(tmp, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "row-junk.txt"), []byte("junk"), 0o644); err != nil {
		t.Fatal(err)
	}
	runIn(t, root)
	if _, err := os.Stat(tmp); !os.IsNotExist(err) {
		t.Fatalf("the run must clear .punchtape/tmp whole: %v", err)
	}
}

// TestTmpOutsideUniverse — the machine's temporary directory stands
// outside the coverage universe: a file there is no line of work,
// never holds a verdict, never asks for a claim.
func TestTmpOutsideUniverse(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	tmp := filepath.Join(root, zone.Dirname, zone.TmpDir)
	if err := os.MkdirAll(tmp, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "junk.txt"), []byte("junk"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "verdict: #1") {
		t.Fatalf("a file in the machine's tmp must not hold the verdict, state:\n%s", state)
	}
	if strings.Contains(state, "junk.txt") {
		t.Fatalf("a file in the machine's tmp is invisible to the universe, state:\n%s", state)
	}
}

// TestOrderLawSingleSource — the behaviors' order lives by one text:
// a remove-and-re-add of one id in a single submission keeps the id
// at its first-seen place, once; and the journal's fold — the same
// law, the same bytes — rebuilds byte-for-byte the standing the
// live accept left.
func TestOrderLawSingleSource(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	stream, err := store.OpenStream(filepath.Join(root, zone.Dirname, zone.JournalDir))
	if err != nil {
		t.Fatal(err)
	}
	zoneDir := filepath.Join(root, zone.Dirname)
	m, err := open(root, root, zoneDir, stream)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.checkChain(); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, `
work: the re-add of one id
mandate: [{line: D1, target: tool/copy}]
remove: [tool/copy]
behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies again, pinned by exact bytes.
      claims:
        - src/in.txt:1-1
      covers: [R1]
      rows:
        - run: sh -c 'printf "one\n"; cat src/in.txt > out.txt'
          expect:
            files:
                out.txt: ok
`)
	inbox := filepath.Join(zoneDir, zone.InboxSlot)
	raw, err := os.ReadFile(inbox)
	if err != nil {
		t.Fatal(err)
	}
	d, canonical, normErr := doc.Normalize(raw)
	if err := m.handleInbox(inbox, d, canonical, normErr); err != nil {
		t.Fatalf("the re-add must accept: %v", err)
	}
	if len(m.standing.Order) != 1 || m.standing.Order[0] != "tool/copy" {
		t.Fatalf("the order keeps the id once at its first-seen place: %v", m.standing.Order)
	}
	// The fold replays the one law: the rebuilt standing equals the
	// live accept's, byte for byte.
	records, err := journalRecords(t, root)
	if err != nil {
		t.Fatal(err)
	}
	folded, err := journal.Fold(records, cas.New(filepath.Join(zoneDir, zone.ObjectsDir)).Get)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(folded, m.standing) {
		t.Fatalf("the fold must rebuild the live standing exactly:\nlive:   %+v\nfolded: %+v", m.standing, folded)
	}
}

// TestStaleActHeuristicGone — a submission carrying the standing
// corpus's own bytes with a fresh narrative, or the corpus
// reformatted, walks the ordinary validation path: no pattern
// heuristic exists, the mandate law answers by its own name, and no
// "stale corpus" line exists anywhere.
func TestStaleActHeuristicGone(t *testing.T) {
	root := newBornRoot(t)
	stale := "work: a whole new story of the same corpus\n" + zoneFile(t, root, zone.CorpusFile)
	writeZone(t, root, zone.InboxSlot, stale)
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if strings.Contains(report, "stale corpus") {
		t.Fatalf("the dead heuristic must not speak:\n%s", report)
	}
	if !strings.Contains(report, "mandate: the submission replaces standing behaviors without a mandate — name the demand lines that ordered it (law:candor)") {
		t.Fatalf("the old act's shape walks the ordinary path:\n%s", report)
	}
	reformatted := strings.Replace(zoneFile(t, root, zone.CorpusFile), `out.txt: ok`, `'out.txt': ok`, 1)
	writeZone(t, root, zone.InboxSlot, reformatted)
	runIn(t, root)
	report = zoneFile(t, root, zone.ReportFile)
	if strings.Contains(report, "stale corpus") {
		t.Fatalf("a reformatted corpus is an ordinary submission:\n%s", report)
	}
	if !strings.Contains(report, "without a mandate") {
		t.Fatalf("the ordinary path answers the reformatted corpus by its own law:\n%s", report)
	}
}

// TestSubmissionWithHugeDRangeRefusedFast — a D range spelled far
// past the stream is refused at once, the refusal journaled, the
// zone alive for the next run: the gates never generate the
// spelling's span. The spelling's own size stays battery-safe — a
// regressed guard costs this test milliseconds, not the host; the
// wide-scale proof runs fenced in a sandbox.
func TestSubmissionWithHugeDRangeRefusedFast(t *testing.T) {
	root := newRoot(t)
	writeZone(t, root, zone.InboxSlot, `
work: a requirement spanning a range far past the stream
requirements:
    - id: R1
      lines: ["D1..D100000"]
      prose: everything
`)
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "past the stream's 1 lines") {
		t.Fatalf("the refusal must name the stream's size, got:\n%s", report)
	}
	if !strings.Contains(report, "requirements[0].lines[0]") {
		t.Fatalf("the refusal must name the field, got:\n%s", report)
	}
	// The zone stands alive: the next run is an ordinary pass.
	writeZone(t, root, zone.InboxSlot, "")
	runIn(t, root)
}

// TestUniverseSizeWired — the stream's own count reaches the range
// gates through the wired resolver: a spelling past the end is
// refused with its line number and the stream's size.
func TestUniverseSizeWired(t *testing.T) {
	root := newRoot(t)
	writeZone(t, root, zone.InboxSlot, `
work: a marker claiming lines the stream does not carry
not-a-requirement:
    - lines: ["D1..D99"]
      reason: wide
`)
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "not-a-requirement[0].lines[0]: the range names line D99 past the stream's 1 lines") {
		t.Fatalf("the overflow must be refused by name, got:\n%s", report)
	}
}

// runFrom — one call of the machine from a working directory; the
// terminal's whole text.
func runFrom(t *testing.T, dir string) string {
	t.Helper()
	var term bytes.Buffer
	if code := Run(dir, &term); code != 0 {
		t.Fatalf("run from %s exited %d: %s", dir, code, term.String())
	}
	return term.String()
}

// TestInstructionFromRootByteIdentical — a call from the root names
// the zone's files exactly as the zone spells them.
func TestInstructionFromRootByteIdentical(t *testing.T) {
	root := newRoot(t)
	term := runFrom(t, root)
	if !strings.Contains(term, "submit the first document as .punchtape/inbox.yaml (copy .punchtape/template.yaml)") {
		t.Fatalf("the root's instruction must carry the zone-relative name:\n%s", term)
	}
}

// TestInstructionFromSubdirNamesZoneFile — a call from below names
// the zone's files from the caller's directory: an obedient reader
// lands in the real zone, no stray zone forks below.
func TestInstructionFromSubdirNamesZoneFile(t *testing.T) {
	root := newRoot(t)
	sub := filepath.Join(root, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	term := runFrom(t, sub)
	if !strings.Contains(term, "submit the first document as ../.punchtape/inbox.yaml (copy ../.punchtape/template.yaml)") {
		t.Fatalf("the subdirectory's instruction must name the zone from the caller:\n%s", term)
	}
	// The obedient reader writes exactly where the instruction points:
	// the submission lands in the real zone and the machine takes it.
	if err := os.WriteFile(filepath.Join(sub, "..", ".punchtape", "inbox.yaml"), []byte("work: called from below\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runFrom(t, sub)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "queue: empty") {
		t.Fatalf("the submission written by the printed path must be taken:\n%s", state)
	}
	// Nothing forks below: the caller's directory carries no zone.
	if _, err := os.Stat(filepath.Join(sub, ".punchtape")); !os.IsNotExist(err) {
		t.Fatalf("no stray zone may stand below: %v", err)
	}
}

// TestInstructionOutsideTreeAbsolute — a caller whose spelling the
// relative path cannot reach hears the absolute path.
func TestInstructionOutsideTreeAbsolute(t *testing.T) {
	if got := slotFrom("sub", "/root", zone.InboxSlot); got != "/root/.punchtape/inbox.yaml" {
		t.Fatalf("an unreachable spelling hears the absolute path, got %q", got)
	}
}

// TestRefusalNamesTheCaller — the refusal's own voices name the inbox
// from the caller's directory too: the where of the fact and the
// footer's fix-it line, both spoken to the reader the report addresses.
func TestRefusalNamesTheCaller(t *testing.T) {
	root := newRoot(t)
	writeZone(t, root, zone.InboxSlot, "work: a giant note\nnotes:\n  - kind: work\n    text: "+
		strings.Repeat("a", 600*1024)+"\n")
	sub := filepath.Join(root, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	runFrom(t, sub)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "../.punchtape/inbox.yaml: the document is bounded") {
		t.Fatalf("the refusal's where must name the inbox from the caller:\n%s", report[:300])
	}
	if !strings.Contains(report, "the submission stays in ../.punchtape/inbox.yaml — fix it there") {
		t.Fatalf("the refusal's footer must name the inbox from the caller:\n%s", report[:300])
	}
}

// TestConcurrentRunsSerialize — two simultaneous runs on one zone
// stand in one line behind the zone lock: sequential run numbers, a
// journal that parses whole, no torn tail, the standing equal to
// the fold.
func TestConcurrentRunsSerialize(t *testing.T) {
	mountPorts()
	root := newRoot(t)
	writeZone(t, root, zone.InboxSlot, `
work: a claims-less behavior whose row always re-runs
intents:
    - {lines: [D1], outcome: work, prose: the line is declared as work}
requirements:
    - id: R1
      lines: [D1]
      prose: the slow check
behaviors:
    - id: slow/one
      name: slow row
      prose: a row that sleeps
      covers: [R1]
      rows:
        - run: "sleep 0.3"
          expect:
            out: "pin"
`)
	runIn(t, root)
	var wg sync.WaitGroup
	codes := make([]int, 2)
	terms := make([]string, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			var term bytes.Buffer
			codes[i] = Run(root, &term)
			terms[i] = term.String()
		}(i)
	}
	wg.Wait()
	for i := 0; i < 2; i++ {
		if codes[i] != 0 {
			t.Fatalf("the parallel run %d exited %d:\n%s", i, codes[i], terms[i])
		}
	}
	stream, err := store.OpenStream(filepath.Join(root, zone.Dirname, zone.JournalDir))
	if err != nil {
		t.Fatal(err)
	}
	records, err := stream.Records()
	if err != nil {
		t.Fatalf("the journal must parse whole after the parallel runs: %v", err)
	}
	var ns []int
	for _, rec := range records {
		if rec.Kind == journal.KindRun && rec.Run != nil {
			ns = append(ns, rec.Run.N)
		}
	}
	if len(ns) != 3 {
		t.Fatalf("three run records expected, got %v", ns)
	}
	for i := 1; i < len(ns); i++ {
		if ns[i] != ns[i-1]+1 {
			t.Fatalf("the run numbers must be sequential, got %v", ns)
		}
	}
	entries, err := os.ReadDir(filepath.Join(root, zone.Dirname, zone.JournalDir))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".torn") {
			t.Fatalf("the torn tail must stay empty: %s", e.Name())
		}
	}
	// The standing equals the fold: the next pass is ordinary.
	runIn(t, root)
}

// TestLockReleasedOnExit — a finished run leaves the zone lock free:
// the next take answers at once.
func TestLockReleasedOnExit(t *testing.T) {
	mountPorts()
	root := newRoot(t)
	runIn(t, root)
	path := filepath.Join(root, zone.Dirname, zone.RunLockFile)
	held := make(chan error, 1)
	go func() {
		release, serr := fsys.OS{}.Flock(path)
		if serr != nil {
			held <- serr
			return
		}
		held <- release()
	}()
	select {
	case err := <-held:
		if err != nil {
			t.Fatalf("the free lock must answer: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("the lock must be free after the run")
	}
}

// TestOversizedInboxRefusedBeforeRead — a giant inbox is refused
// ahead of its read, the fact naming the bound and the size.
func TestOversizedInboxRefusedBeforeRead(t *testing.T) {
	root := newRoot(t)
	big := strings.Repeat("a", 600*1024)
	writeZone(t, root, zone.InboxSlot, "work: a giant note\nnotes:\n  - kind: work\n    text: "+big+"\n")
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "the document is bounded — 524288 bytes, this one carries 6144") {
		t.Fatalf("the oversized document must be refused by its size:\n%s", report[:400])
	}
}

// TestClaimsCeilingRefused — the thirty-third claim of one behavior
// is refused with its count and the ceiling.
func TestClaimsCeilingRefused(t *testing.T) {
	root := newRoot(t)
	if err := os.WriteFile(filepath.Join(root, "f.txt"), []byte(strings.Repeat("l\n", 40)), 0o644); err != nil {
		t.Fatal(err)
	}
	var claims []string
	for i := 1; i <= 33; i++ {
		claims = append(claims, fmt.Sprintf("f.txt:%d-%d", i, i))
	}
	writeZone(t, root, zone.InboxSlot, fmt.Sprintf(`
work: thirty-three claims
intents:
    - {lines: [D1], outcome: work, prose: the line is declared as work}
requirements:
    - {id: R1, lines: [D1], prose: the wide claims}
behaviors:
    - id: b/one
      name: one
      prose: too many claims
      covers: [R1]
      claims: [%s]
      rows:
        - {run: "true", expect: {out: x}}
`, strings.Join(claims, ", ")))
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "behaviors[0].claims: the behavior carries 33 claims — the ceiling is 32") {
		t.Fatalf("the claims ceiling must be refused by name:\n%s", report[:400])
	}
}

// TestRowsCeilingRefused — the seventeenth row of one behavior is
// refused with its count and the ceiling.
func TestRowsCeilingRefused(t *testing.T) {
	root := newRoot(t)
	var rows strings.Builder
	for i := 1; i <= 17; i++ {
		fmt.Fprintf(&rows, "        - {run: \"echo %d\", expect: {out: \"%d\\n\"}}\n", i, i)
	}
	writeZone(t, root, zone.InboxSlot, fmt.Sprintf(`
work: seventeen rows
intents:
    - {lines: [D1], outcome: work, prose: the line is declared as work}
requirements:
    - {id: R1, lines: [D1], prose: the many rows}
uncovered-lines:
    - {range: "AGENTS.md:1-9", reason: the machine's own hook, refs: [D1]}
behaviors:
    - id: b/one
      name: one
      prose: too many rows
      covers: [R1]
      claims: ["AGENTS.md:1-1"]
      rows:
%s`, rows.String()))
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "behaviors[0].rows: the behavior carries 17 rows — the ceiling is 16") {
		t.Fatalf("the rows ceiling must be refused by name:\n%s", report[:400])
	}
}

// TestEmptyLinesListsRefused — an empty lines list and a multi-line
// order demand are refused at their own fields.
func TestEmptyLinesListsRefused(t *testing.T) {
	root := newRoot(t)
	writeZone(t, root, zone.InboxSlot, `
work: empty lists
requirements:
    - id: R1
      lines: []
      prose: nothing claimed
not-a-requirement:
    - lines: []
      reason: nothing marked
intents:
    - lines: []
      outcome: mark
      prose: nothing declared
open:
    - {demand: "D1", question: q?, state: awaiting}
`)
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	for _, fact := range []string{
		"requirements[0].lines: the list is empty — a requirement claims at least one line",
		"not-a-requirement[0].lines: the list is empty — a marker claims at least one line",
		"intents[0].lines: the list is empty — an intent declares at least one line",
	} {
		if !strings.Contains(report, fact) {
			t.Fatalf("the empty list must be refused by name (%s):\n%s", fact, report[:500])
		}
	}
}

// TestRunBudgetCarriesRowsOver — a spent budget carries the waiting
// rows to the next run, named by a warning; the next run takes them.
func TestRunBudgetCarriesRowsOver(t *testing.T) {
	root := newRoot(t)
	writeZone(t, root, zone.EnvSlot, "PUNCHTAPE_ROW_TIMEOUT=1\nPUNCHTAPE_RUN_BUDGET=1\n")
	writeZone(t, root, zone.InboxSlot, `
work: two slow rows under a one-second budget
intents:
    - {lines: [D1], outcome: work, prose: the line is declared as work}
requirements:
    - {id: R1, lines: [D1], prose: the slow rows}
behaviors:
    - id: slow/one
      name: slow
      prose: two sleeping rows
      covers: [R1]
      rows:
        - {run: "sleep 1 #one", expect: {out: x}}
        - {run: "sleep 1 #two", expect: {out: x}}
`)
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "run budget spent — 1 rows wait the next run") {
		t.Fatalf("the spent budget must name its waiting row:\n%s", report[:400])
	}
	// The next run, under the standing default budget, takes both.
	writeZone(t, root, zone.EnvSlot, "")
	runIn(t, root)
	report = zoneFile(t, root, zone.ReportFile)
	if strings.Contains(report, "run budget spent") {
		t.Fatalf("the standing budget must take the waiting row:\n%s", report[:400])
	}
}

// smallStatFS — the read-race witness: the stat answers a small
// size, the read answers the file's whole bytes.
type smallStatFS struct {
	ports.FS
}

type smallInfo struct {
	os.FileInfo
}

func (smallInfo) Size() int64 { return 1 }

func (f smallStatFS) Stat(p string) (os.FileInfo, error) {
	info, err := f.FS.Stat(p)
	if err != nil {
		return nil, err
	}
	return smallInfo{info}, nil
}

// TestInboxGrownBetweenLookAndRead — a document that grew between
// the stat and the read meets the same bound at the read's own
// length.
func TestInboxGrownBetweenLookAndRead(t *testing.T) {
	root := newRoot(t)
	writeZone(t, root, zone.InboxSlot, "work: a giant note\nnotes:\n  - kind: work\n    text: "+strings.Repeat("a", 600*1024)+"\n")
	ports.FSBind = smallStatFS{FS: fsys.OS{}}
	defer mountPorts()
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "the document is bounded — 524288 bytes, this one carries 6144") {
		t.Fatalf("the grown document must meet the same bound:\n%s", report[:300])
	}
}

// TestRegistryListCeilingRefused — the sixty-fifth entry of one
// registry list is refused with the list's name and the ceiling.
func TestRegistryListCeilingRefused(t *testing.T) {
	root := newRoot(t)
	if err := os.WriteFile(filepath.Join(root, "f.txt"), []byte("l\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var excl strings.Builder
	for i := 0; i < 65; i++ {
		excl.WriteString("    - {range: \"f.txt:1-1\", reason: r, refs: [D1]}\n")
	}
	writeZone(t, root, zone.InboxSlot, "work: sixty-five exclusions\nuncovered-lines:\n"+excl.String())
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "uncovered-lines: the submission carries 65 uncovered-lines entries — the ceiling is 64") {
		t.Fatalf("the list ceiling must be refused by name:\n%s", report[:400])
	}
}

// TestRunBudgetPinRefusalsAreLoud — the budget pin obeys the same
// law as the deadline pin: whole positive seconds or a loud refusal
// naming the line, the run never starts.
func TestRunBudgetPinRefusalsAreLoud(t *testing.T) {
	root := newRoot(t)
	writeZone(t, root, zone.EnvSlot, "PUNCHTAPE_RUN_BUDGET=x\n")
	var term bytes.Buffer
	code := Run(root, &term)
	if code != 1 || !strings.Contains(term.String(), `env pin refused — .punchtape/env: "PUNCHTAPE_RUN_BUDGET=x"`) {
		t.Fatalf("the bad budget pin must refuse the run loudly, code %d:\n%s", code, term.String())
	}
}

// TestExpectFilesCeilingRefused — the ninth expected file of one row
// is refused with its count and the ceiling.
func TestExpectFilesCeilingRefused(t *testing.T) {
	root := newRoot(t)
	var files strings.Builder
	for i := 1; i <= 9; i++ {
		fmt.Fprintf(&files, "              f%d.txt: x\n", i)
	}
	writeZone(t, root, zone.InboxSlot, fmt.Sprintf(`
work: nine expected files
intents:
    - {lines: [D1], outcome: work, prose: the line is declared as work}
requirements:
    - {id: R1, lines: [D1], prose: the many files}
behaviors:
    - id: b/one
      name: one
      prose: too many expected files
      covers: [R1]
      rows:
        - run: "true"
          expect:
            files:
%s`, files.String()))
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "behaviors[0].rows[0].expect.files: the row expects 9 files — the ceiling is 8") {
		t.Fatalf("the expect-files ceiling must be refused by name:\n%s", report[:400])
	}
}
