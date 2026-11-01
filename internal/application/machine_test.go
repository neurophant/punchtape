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
	"testing"

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
	if !strings.Contains(report, "requirements[0].lines[0]") || !strings.Contains(report, "the demand stream carries no line D9") {
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
	m, err := open(root, zoneDir, stream)
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
