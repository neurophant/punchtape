// The verdict's, the knowledge's and the discovery block's battery:
// the completion act closes a fully green corpus, the empty corpus
// is refused, the drift is named; each note stands once with its tags
// inline and the index gathers them, the lessons carry the row's
// bytes before and after; the AGENTS block stands idempotently; and
// the canon's example must pass validation and run green against the
// real machine, or the build fails.
package application

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/neurophant/punchtape/internal/application/canon"
	"github.com/neurophant/punchtape/internal/domain"
	"github.com/neurophant/punchtape/internal/domain/doc"
	"github.com/neurophant/punchtape/internal/domain/fprint"
	"github.com/neurophant/punchtape/internal/domain/journal"
	"github.com/neurophant/punchtape/internal/domain/zone"
	"github.com/neurophant/punchtape/internal/infrastructure/fsys"
	"github.com/neurophant/punchtape/internal/infrastructure/process"
	"github.com/neurophant/punchtape/internal/infrastructure/store"
	"github.com/neurophant/punchtape/internal/infrastructure/store/cas"
	"github.com/neurophant/punchtape/internal/ports"
)

func newRoot(t *testing.T) string {
	t.Helper()
	ports.FSBind = fsys.OS{}
	ports.ExecBind = process.OS{}
	root := t.TempDir()
	writeZone(t, root, zone.DemandsSlot, "a checked tool")
	runIn(t, root)
	return root
}

func TestVerdictClosesGreenCorpusAndRefusesEmpty(t *testing.T) {
	root := newRoot(t)

	// The empty corpus: the completion act is refused by name.
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if report := zoneFile(t, root, zone.ReportFile); !strings.Contains(report, "verdict: refused — the corpus is empty") {
		t.Fatalf("the empty corpus must be refused by name, got:\n%s", report)
	}

	// A green corpus with its stream accounted: the completion act
	// records the verdict.
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, fmt.Sprintf(`
work: the copy's first document
intents:
    - {lines: [D1], outcome: work, prose: "the tool and its check, told again"}
requirements:
    - id: R1
      lines: [D1]
      prose: the tool and its check
uncovered-lines:
    - {range: "AGENTS.md:1-9", reason: the machine's own hook, refs: [D1]}
    - {range: "out.txt:1-1", reason: the row's own output, refs: [D1]}
behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies.
      claims:
        - src/in.txt:1-1
      covers: [R1]
      rows:
        - run: sh -c 'printf "%%s %%s\n" %s %s; cat src/in.txt > out.txt'
          expect:
            files:
                out.txt: ok
`, claimsTokenOf(t, root, "src/in.txt:1-1"), demandTokenOf("a checked tool")))
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "verdict: #1 on v1") || !strings.Contains(state, "standing") {
		t.Fatalf("the verdict must stand, state:\n%s", state)
	}

	// A claimed line changes and stays changed: the verdict opens by
	// the run that first saw the moved tree.
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("changed"), 0o644); err != nil {
		t.Fatal(err)
	}
	runIn(t, root)
	if state = zoneFile(t, root, zone.StateFile); !strings.Contains(state, "open (drift since run 2)") {
		t.Fatalf("the moved claims must open the verdict by its run, state:\n%s", state)
	}
}

// TestDoneActCloses — the completion act is one line, done: true:
// the inbox goes, the journal carries the completion fact, the
// verdict records over the standing corpus.
func TestDoneActCloses(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if _, err := os.Stat(filepath.Join(root, zone.Dirname, zone.InboxSlot)); !os.IsNotExist(err) {
		t.Fatalf("the done act consumes the inbox")
	}
	records, err := journalRecords(t, root)
	if err != nil {
		t.Fatal(err)
	}
	var completions int
	for _, rec := range records {
		if rec.Kind == journal.KindCompletion {
			completions++
		}
	}
	if completions != 1 {
		t.Fatalf("the journal must carry the completion fact once, got %d", completions)
	}
	if state := zoneFile(t, root, zone.StateFile); !strings.Contains(state, "verdict: #1") {
		t.Fatalf("the one-line act must record the verdict, state:\n%s", state)
	}
}

// TestDoneWithExtraFieldsRefused — done beside any other field is
// the shape law's own refusal; the inbox stays.
func TestDoneWithExtraFieldsRefused(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	writeZone(t, root, zone.InboxSlot, "done: true\nwork: a story beside the act\n")
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "done: a done document carries only done: true (law:document)") {
		t.Fatalf("the crowded done act must be refused by name:\n%s", report)
	}
	if _, err := os.Stat(filepath.Join(root, zone.Dirname, zone.InboxSlot)); err != nil {
		t.Fatalf("the refused act keeps the inbox")
	}
}

// TestDoneOverRedCorpus — the act over a corpus that is not green
// records nothing: the report names the hole by its own law.
func TestDoneOverRedCorpus(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("NO"), 0o644); err != nil {
		t.Fatal(err)
	}
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if state := zoneFile(t, root, zone.StateFile); strings.Contains(state, "verdict: #") {
		t.Fatalf("a red corpus must record no verdict, state:\n%s", state)
	}
	if report := zoneFile(t, root, zone.ReportFile); !strings.Contains(report, "verdict: not yet — a row is red or stale") {
		t.Fatalf("the report must name the red hole:\n%s", report)
	}
}

func TestVerdictNotYetWhenAFrozenRowIsStale(t *testing.T) {
	root := newBornRoot(t)
	// A broken build freezes the rows; a moved claimed line leaves the
	// green fact on an old proof: the completion act says not yet.
	writeZone(t, root, zone.InboxSlot, `
work: the frozen document
mandate: [{line: D1, target: tool/copy}]
build: false
behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies.
      claims:
        - src/in.txt:1-1
      covers: [R1]
      rows:
        - run: cat src/in.txt > out.txt
          expect:
            files:
                out.txt: ok
`)
	runIn(t, root)
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("moved on"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "verdict: not yet — a row is red or stale") {
		t.Fatalf("a stale frozen row must hold the verdict:\n%s", report)
	}
}

// streamFixture — a born instance of a three-line stream with one
// claiming green row writing out.txt; the stream's account is the
// caller's.
func streamFixture(t *testing.T) string {
	t.Helper()
	mountPorts()
	root := t.TempDir()
	writeZone(t, root, zone.DemandsSlot, "one\n\ntwo\n\nthree\n")
	runIn(t, root)
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, fmt.Sprintf(`
work: the fixture's first document
behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies.
      claims:
        - src/in.txt:1-1
      rows:
        - run: sh -c 'printf "%%s\n" %s; cat src/in.txt > out.txt'
          expect:
            files:
                out.txt: ok
`, claimsTokenOf(t, root, "src/in.txt:1-1")))
	runIn(t, root)
	return root
}

func TestVerdictNeedsTheStreamAccountedWhole(t *testing.T) {
	root := streamFixture(t)
	// A green behavior-only corpus: the stream stands wholly unresolved.
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "verdict: not yet — unresolved lines: D1 (no outcome), D2 (no outcome), D3 (no outcome)") {
		t.Fatalf("the unresolved stream must hold the verdict by name:\n%s", report)
	}
	// The registry arrives: two lines as one requirement, the third
	// marked — but no behavior carries the requirement yet.
	writeZone(t, root, zone.InboxSlot, `
work: the registry arrives
intents:
    - {lines: [D1..D2], outcome: work, prose: the working pair of the stream}
    - {lines: [D3], outcome: mark, prose: the closing line orders nothing}
requirements:
    - id: R1
      lines: [D1..D2]
      prose: the working lines
not-a-requirement:
    - lines: [D3]
      reason: the closing line names no work
`)
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	report = zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "verdict: not yet — the requirements R1 stand uncovered by any behavior's covers") {
		t.Fatalf("the orphan requirement must hold the verdict by name:\n%s", report)
	}
	if strings.Contains(zoneFile(t, root, zone.StateFile), "verdict: #") {
		t.Fatalf("no verdict may stand over a hole")
	}
	// The cover arrives: the verdict stands, its record carrying the
	// registry's own digest.
	writeZone(t, root, zone.InboxSlot, fmt.Sprintf(`
work: the cover arrives
mandate: [{line: D1, target: tool/copy}]
uncovered-lines:
    - {range: "AGENTS.md:1-9", reason: the machine's own hook, refs: [D1]}
    - {range: "out.txt:1-1", reason: the row's own output, refs: [D1]}
behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies.
      claims:
        - src/in.txt:1-1
      covers: [R1]
      rows:
        - run: sh -c 'printf "%%s %%s\n" %s %s; cat src/in.txt > out.txt'
          expect:
            files:
                out.txt: ok
`, claimsTokenOf(t, root, "src/in.txt:1-1"), demandTokenOf("one\ntwo")))
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "verdict: #1") {
		t.Fatalf("the accounted corpus must close the verdict, state:\n%s", state)
	}
	if !strings.Contains(state, "demands: 3/3 lines; requirements 1 standing, 1 covered; markers: 1") {
		t.Fatalf("the demands map must show the whole account, state:\n%s", state)
	}
}

// TestVerdictRecomputesFromTheJournal — the journal-only proof: an
// outside reader with the journal's records, the stream's object and
// the working tree re-spells the verdict's conditions and the
// registry digest with the domain's own primitives, and reaches the
// machine's recorded verdict — zero unaccounted information.
func TestVerdictRecomputesFromTheJournal(t *testing.T) {
	root := streamFixture(t)
	writeZone(t, root, zone.InboxSlot, fmt.Sprintf(`
work: the whole account in one document
mandate: [{line: D1, target: tool/copy}]
intents:
    - {lines: [D1..D2], outcome: work, prose: the working pair of the stream}
    - {lines: [D3], outcome: mark, prose: the closing line orders nothing}
requirements:
    - id: R1
      lines: [D1..D2]
      prose: the working lines
not-a-requirement:
    - lines: [D3]
      reason: the closing line names no work
uncovered-lines:
    - {range: "AGENTS.md:1-9", reason: the machine's own hook, refs: [D1]}
    - {range: "out.txt:1-1", reason: the row's own output, refs: [D1]}
behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies.
      claims:
        - src/in.txt:1-1
      covers: [R1]
      rows:
        - run: sh -c 'printf "%%s %%s\n" %s %s; cat src/in.txt > out.txt'
          expect:
            files:
                out.txt: ok
`, claimsTokenOf(t, root, "src/in.txt:1-1"), demandTokenOf("one\ntwo")))
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if !strings.Contains(zoneFile(t, root, zone.StateFile), "verdict: #1") {
		t.Fatalf("the fixture must close its verdict")
	}

	// The reader's material: the journal's records, the object
	// store, the demand stream file, the tree, the environment —
	// nothing of the machine's own standing.
	stream, err := store.OpenStream(filepath.Join(root, zone.Dirname, zone.JournalDir))
	if err != nil {
		t.Fatal(err)
	}
	records, err := stream.Records()
	if err != nil {
		t.Fatal(err)
	}
	objects := cas.New(filepath.Join(root, zone.Dirname, zone.ObjectsDir))
	st, err := journal.Fold(records, objects.Get)
	if err != nil {
		t.Fatal(err)
	}
	if st.Verdict == nil {
		t.Fatal("the journal must carry the verdict")
	}
	wish, err := os.ReadFile(filepath.Join(root, zone.Dirname, zone.DemandsSlot))
	if err != nil {
		t.Fatal(err)
	}

	// Condition 1: every row green on the verdict's own claims
	// pointers and the environment's hash.
	pointers := map[string]string{}
	for _, sc := range st.Verdict.Scopes {
		pointers[sc.ID] = sc.Pointer
	}
	pins, err := ZoneEnv(filepath.Join(root, zone.Dirname))
	if err != nil {
		t.Fatal(err)
	}
	env := EnvHash(pins)
	for _, id := range st.Order {
		b := st.Behaviors[id]
		for i, row := range b.Rows {
			fact, ok := st.Rows[id+"/"+strconv.Itoa(i)]
			if !ok || fact.Status != journal.StatusGreen ||
				fact.Proof != fprint.Proof(doc.MarshalRow(row), pointers[id], env) {
				t.Fatal("condition 1 fails the recompute")
			}
		}
	}

	// Condition 2: the stream's every non-empty line inside some
	// requirement's span or some marker's span.
	var universe []int
	for _, line := range strings.Split(string(wish), "\n") {
		if strings.TrimSpace(line) != "" {
			universe = append(universe, len(universe)+1)
		}
	}
	claimed := map[int]bool{}
	for _, id := range st.Registry.Order {
		nums, err := doc.ParseRanges(st.Registry.Requirements[id].Lines)
		if err != nil {
			t.Fatal(err)
		}
		for _, n := range nums {
			claimed[n] = true
		}
	}
	for _, marker := range st.Registry.Markers {
		nums, err := doc.ParseRanges(marker.Lines)
		if err != nil {
			t.Fatal(err)
		}
		for _, n := range nums {
			claimed[n] = true
		}
	}
	for _, n := range universe {
		if !claimed[n] {
			t.Fatal("condition 2 fails the recompute")
		}
	}

	// Condition 3: every standing requirement named in some
	// standing behavior's covers.
	covers := map[string]bool{}
	for _, id := range st.Order {
		for _, c := range st.Behaviors[id].Covers {
			covers[c] = true
		}
	}
	for _, id := range st.Registry.Order {
		if !covers[id] {
			t.Fatal("condition 3 fails the recompute")
		}
	}

	// The digest: the registry's canonical sections, re-marshaled
	// here from the fold, must equal the record's own seal.
	var sections doc.Document
	for _, id := range st.Registry.Order {
		sections.Requirements = append(sections.Requirements, st.Registry.Requirements[id])
	}
	sections.Markers = st.Registry.Markers
	sections.Exclusions = st.Registry.Exclusions
	sections.Open = st.Registry.Open
	if got := domain.Digest(doc.Marshal(sections)); got != st.Verdict.RegistryDigest {
		t.Fatalf("the registry digest must recompute: %s != %s", got, st.Verdict.RegistryDigest)
	}
}

func TestKnowledgeNotesTagsAndLessons(t *testing.T) {
	root := newRoot(t)
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Red: the expectation does not hold.
	writeZone(t, root, zone.InboxSlot, fmt.Sprintf(`
work: the red copy first
notes:
    - kind: work
      text: the copy must carry the input bytes
      refs: [D1]
behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies.
      claims:
        - src/in.txt:1-1
      rows:
        - run: sh -c 'printf "%%s\n" %s; cat src/in.txt > out.txt'
          expect:
            files:
                out.txt: NOPE
`, claimsTokenOf(t, root, "src/in.txt:1-1")))
	runIn(t, root)
	if report := zoneFile(t, root, zone.ReportFile); !strings.Contains(report, `want "NOPE"`) {
		t.Fatalf("the row must be red first, report:\n%s", report)
	}

	// Green: the same behavior, the honest expectation — the lesson
	// lands with the row's bytes before and after.
	writeZone(t, root, zone.InboxSlot, fmt.Sprintf(`
work: the green heal
mandate: [{line: D1, target: tool/copy}]
behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies.
      claims:
        - src/in.txt:1-1
      rows:
        - run: sh -c 'printf "%%s\n" %s; cat src/in.txt > out.txt'
          expect:
            files:
                out.txt: ok
`, claimsTokenOf(t, root, "src/in.txt:1-1")))
	runIn(t, root)
	knowledge := zoneFile(t, root, zone.KnowledgeFile)
	if !strings.Contains(knowledge, "lesson:") ||
		!strings.Contains(knowledge, "before:") || !strings.Contains(knowledge, "after:") ||
		!strings.Contains(knowledge, "NOPE") {
		t.Fatalf("the lesson must carry the fix's bytes, knowledge:\n%s", knowledge)
	}
	if !strings.Contains(knowledge, "- [n1] (work; refs: D1) the copy must carry the input bytes") {
		t.Fatalf("the note must stand once with its kind and references inline, knowledge:\n%s", knowledge)
	}
	if strings.Contains(knowledge, "## index") {
		t.Fatalf("the reverse index is gone, knowledge:\n%s", knowledge)
	}
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "knowledge: 1 notes, 1 lessons") {
		t.Fatalf("the counter must count, state:\n%s", state)
	}
}

func TestAgentsBlockStandsIdempotently(t *testing.T) {
	root := newRoot(t)
	path := filepath.Join(root, agentsFile)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("the block must stand after a run: %v", err)
	}
	first, _ := os.ReadFile(path)
	if !strings.Contains(string(first), blockOpen) || !strings.Contains(string(first), "state.md") {
		t.Fatalf("the block must carry the essence and the addresses:\n%s", first)
	}
	// A foreign line above the block survives; the block's bytes do
	// not change on a re-run.
	foreign := "# the project's own words\n" + string(first)
	if err := os.WriteFile(path, []byte(foreign), 0o644); err != nil {
		t.Fatal(err)
	}
	runIn(t, root)
	second, _ := os.ReadFile(path)
	if !strings.HasPrefix(string(second), "# the project's own words\n") {
		t.Fatalf("the foreign text must survive:\n%s", second)
	}
	runIn(t, root)
	third, _ := os.ReadFile(path)
	if string(second) != string(third) {
		t.Fatalf("the ensured block must be byte-stable")
	}
}

func TestAgentsBlockShapes(t *testing.T) {
	mountPorts()
	root := newRoot(t)
	path := filepath.Join(root, agentsFile)
	// A foreign file without a trailing newline gains the block
	// cleanly padded.
	if err := os.WriteFile(path, []byte("# words without an end"), 0o644); err != nil {
		t.Fatal(err)
	}
	runIn(t, root)
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), "# words without an end\n\n"+blockOpen) {
		t.Fatalf("the foreign text must be padded:\n%s", data)
	}
	// A changed block with the closing marker at the file's very end:
	// the replacement clamps to the bytes that exist.
	stale := blockOpen + "\nold words\n" + blockClose
	if err := os.WriteFile(path, []byte("# mine\n"+stale), 0o644); err != nil {
		t.Fatal(err)
	}
	runIn(t, root)
	data, _ = os.ReadFile(path)
	if strings.Contains(string(data), "old words") ||
		!strings.HasPrefix(string(data), "# mine\n"+blockOpen) ||
		!strings.HasSuffix(string(data), blockClose+"\n") {
		t.Fatalf("the changed block must replace itself whole:\n%s", data)
	}
}

// TestExampleRunsGreen — the canon's own honesty: its example
// passes the machine's validation, runs green against the real
// machine in a fresh fixture (./count — a script printing the
// count) and closes the verdict whole — the registry, the marker,
// the claims and the covers all exercising the laws the contract
// prints. A canon that lies fails the build.
func TestExampleRunsGreen(t *testing.T) {
	src := canon.Load()
	d, _, err := doc.Normalize([]byte(src.Example))
	if err != nil {
		t.Fatalf("the canon example must decode: %v", err)
	}
	if r := doc.Validate(d, nil); len(r) != 0 {
		t.Fatalf("the canon example must validate: %v", r)
	}
	mountPorts()
	root := t.TempDir()
	writeZone(t, root, zone.DemandsSlot, "a counter\ncount the lines of count.txt\n")
	// The project's own files stand before the submission: the
	// counted file and the tool that counts it — claims name real
	// lines, the example's exclusion names the tool's own lines.
	if err := os.WriteFile(filepath.Join(root, "count.txt"), []byte("one\ntwo\nthree\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "count"), []byte("#!/bin/sh\nprintf '3\\n'\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, src.Example)
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "green: 1") {
		t.Fatalf("the canon example must run green, report:\n%s", report)
	}
	// The example's own corpus is the completion act: the whole
	// chain — rows green, the stream accounted, covers standing —
	// must close on the canon's own words.
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "verdict: #1") {
		t.Fatalf("the canon example must close its verdict, state:\n%s", state)
	}
}

// TestCompletionActIsCounted — the completion act is a counted,
// journaled fact of its own: closing the verdict shows itself in
// the render's arithmetic and in the journal — an act invisible to
// every counter is a defect this test holds closed.
func TestCompletionActIsCounted(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	// The completion act: the standing corpus into the inbox slot.
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "submissions: 1 accepted, 0 refused, 1 completed") {
		t.Fatalf("the completion act must show in the render's arithmetic:\n%s", state)
	}
	records, err := journalRecords(t, root)
	if err != nil {
		t.Fatal(err)
	}
	var completions int
	for _, rec := range records {
		if rec.Kind == journal.KindCompletion {
			completions++
		}
	}
	if completions != 1 {
		t.Fatalf("the journal must carry the completion fact once, got %d", completions)
	}
}

// TestSelfHealingRowLeavesNoMovedMark — a row that repairs its own
// claimed file inside the run leaves the tree exactly as the verdict
// certified: the render's mark, measured after the execution, says
// standing, the history holds no red — the two truths agree. The
// proof reads its claimed line as it stands, then rewrites it — the
// token it prints is always the token of the moment.
func TestSelfHealingRowLeavesNoMovedMark(t *testing.T) {
	root := newRoot(t)
	if err := os.WriteFile(filepath.Join(root, "ex.txt"), []byte("junk"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, fmt.Sprintf(`
work: the self-healing writer
intents:
    - {lines: [D1], outcome: work, prose: "the tool and its check, told again"}
requirements:
    - id: R1
      lines: [D1]
      prose: the tool and its check
uncovered-lines:
    - {range: "AGENTS.md:1-9", reason: the machine's own hook, refs: [D1]}
behaviors:
    - id: tool/write
      name: The writer
      prose: Writes its own proof file.
      claims:
        - ex.txt:1-1
      covers: [R1]
      rows:
        - run: |-
            sh -c 'd=$(sha256sum ex.txt | cut -d" " -f1); c=$(printf "ex.txt:1-1=%%s" "$d" | sha256sum | cut -c1-12); printf "%%s %%s\n" "$c" %s; printf ok > ex.txt'
          expect:
            files:
                ex.txt: ok
`, demandTokenOf("a checked tool")))
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "verdict: #1") {
		t.Fatalf("the verdict must be recorded, state:\n%s", state)
	}
	// The break and the heal inside one run: the claimed file breaks,
	// the row rewrites it and stays green.
	if err := os.WriteFile(filepath.Join(root, "ex.txt"), []byte("no"), 0o644); err != nil {
		t.Fatal(err)
	}
	runIn(t, root)
	state = zoneFile(t, root, zone.StateFile)
	if strings.Contains(state, "claims moved") {
		t.Fatalf("a self-healed run must carry no moved mark — the tree matches the certificate:\n%s", state)
	}
	if !strings.Contains(state, "— standing") {
		t.Fatalf("the verdict must stand after the self-heal:\n%s", state)
	}
	if strings.Contains(state, "rows RED") {
		t.Fatalf("no red row ever stood — the history must hold no drift:\n%s", state)
	}
}

// journalRecords — the zone's journal records, read back for the
// test's own eyes.
func journalRecords(t *testing.T, root string) ([]journal.Record, error) {
	t.Helper()
	stream, err := store.OpenStream(filepath.Join(root, zone.Dirname, zone.JournalDir))
	if err != nil {
		return nil, err
	}
	return stream.Records()
}

func TestCompletionAppendFailureIsHonest(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	// The verdict's own append succeeds (WriteAtomic flips the
	// arm), the completion's append fails: the run says so instead
	// of counting a completion it could not journal.
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	code, term := runCode(t, root, map[string]bool{"append-content:kind: completion": true})
	if code != 1 {
		t.Fatalf("the unjournaled completion must fail the run, exit %d: %s", code, term)
	}
	if !strings.Contains(term, "journal write:") {
		t.Fatalf("the failure must name its cause:\n%s", term)
	}
}

func TestVerdictAppendFailureIsLoud(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	code, term := runCode(t, root, map[string]bool{"append-content:kind: verdict": true})
	if code != 1 {
		t.Fatalf("the unjournaled verdict must fail the run, exit %d: %s", code, term)
	}
	if !strings.Contains(term, "journal write:") {
		t.Fatalf("the failure must name its cause:\n%s", term)
	}
}

// TestVerdictHistorySurvivesTheByteExactHeal — the drift episode is
// a fact of the render, not only of the journal: a byte-exact
// repair must leave the catch and the heal standing in state.md —
// a render that says "never broke" while the journal holds a red
// run is a lie this test holds closed.
func TestVerdictHistorySurvivesTheByteExactHeal(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "verdict history:\n- run 1: verdict #1 on v1 recorded") {
		t.Fatalf("the history must open with the first verdict:\n%s", state)
	}
	// The drift: a claimed line breaks the row, the run catches it.
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("drifted"), 0o644); err != nil {
		t.Fatal(err)
	}
	runIn(t, root)
	state = zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "run 2: 1 rows RED — drift: tool/copy/0") {
		t.Fatalf("the catch must stand in the history:\n%s", state)
	}
	// The byte-exact heal: the file back to its own bytes.
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	state = zoneFile(t, root, zone.StateFile)
	for _, want := range []string{
		"- run 2: 1 rows RED — drift: tool/copy/0",
		"- run 3: healed",
		"- run 3: verdict #2 on v1 recorded",
	} {
		if !strings.Contains(state, want) {
			t.Fatalf("the healed history must carry %q:\n%s", want, state)
		}
	}
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "drift healed: run 2 caught, run 3 green again") {
		t.Fatalf("the closing report must carry the heal:\n%s", report)
	}
	// An idle run rewrites nothing: the history is stable bytes.
	before := state
	runIn(t, root)
	if zoneFile(t, root, zone.StateFile) != before {
		t.Fatalf("an idle run must not rewrite the standing history")
	}
	// A later close with no drift of its own does not retell the
	// old heal: the report of that close carries no heal line.
	if err := os.WriteFile(filepath.Join(root, "second.txt"), []byte("second"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, fmt.Sprintf(`
work: a plain later work
behaviors:
    - id: tool/second
      name: The second
      prose: A plain later work.
      claims:
        - second.txt:1-1
      rows:
        - run: ["sh", "-c", "printf '%%s\\n' %s; printf second > second.txt"]
          expect:
            files:
                second.txt: second
`, claimsTokenOf(t, root, "second.txt:1-1")))
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if report := zoneFile(t, root, zone.ReportFile); strings.Contains(report, "drift healed") {
		t.Fatalf("a close with no drift of its own must not retell the old heal:\n%s", report)
	}
	if state := zoneFile(t, root, zone.StateFile); !strings.Contains(state, "- run 1: verdict #1 on v1 recorded") {
		t.Fatalf("the history keeps its first fact:\n%s", state)
	}
}

func TestHistoryRidesTheOnceReadRecords(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	// The journal's records are read once a run: a disk that turns
	// the epochs unreadable past the open cannot blind the render —
	// the history stands whole from the once-read slice.
	code, term := runCode(t, root, map[string]bool{"read:epoch-!after-write": true})
	if code != 0 {
		t.Fatalf("a render over the once-read records must not die on the disk: %s", term)
	}
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "verdict history:") || strings.Contains(state, "journal unreadable") {
		t.Fatalf("the history must stand whole from the once-read slice:\n%s", state)
	}
}

// TestUncoveredLinesHoldVerdict — the totality gate: a green corpus
// over a tree it does not claim holds the verdict by name, per file
// with counts and ranges; the loud exclusion of the open lines
// closes it. A behavior with zero claims claims nothing — the whole
// tree stands open.
func TestUncoveredLinesHoldVerdict(t *testing.T) {
	mountPorts()
	root := newRoot(t)
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, fmt.Sprintf(`
work: the copy that owns its input only
intents:
    - {lines: [D1], outcome: work, prose: "the tool and its check, told again"}
requirements:
    - id: R1
      lines: [D1]
      prose: the tool and its check
behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies the input, pinned by its bytes.
      claims:
        - src/in.txt:1-1
      covers: [R1]
      rows:
        - run: sh -c 'printf "%%s %%s\n" %s %s; cat src/in.txt > out.txt'
          expect:
            files:
                out.txt: ok
`, claimsTokenOf(t, root, "src/in.txt:1-1"), demandTokenOf("a checked tool")))
	runIn(t, root)
	// The completion act: the corpus is green and whole over the
	// stream — the tree is not, and the verdict waits on the named
	// hole.
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "verdict: not yet — uncovered lines: AGENTS.md 9 (1-9), out.txt 1 (1)") {
		t.Fatalf("the unclaimed output must hold the verdict by name:\n%s", report)
	}
	// The loud exclusions of the open lines close the holes — the
	// verdict lands.
	writeZone(t, root, zone.InboxSlot, `
work: the open lines stand excluded, loudly
uncovered-lines:
    - {range: "AGENTS.md:1-9", reason: the machine's own hook, refs: [D1]}
    - {range: "out.txt:1-1", reason: the row's own output, refs: [D1]}
`)
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if state := zoneFile(t, root, zone.StateFile); !strings.Contains(state, "verdict: #1") {
		t.Fatalf("the excluded hole must let the verdict land, state:\n%s", state)
	}

	// Zero claims: the behavior stands on nothing it names — the
	// whole tree is open, every file named.
	fresh := newRoot(t)
	if err := os.MkdirAll(filepath.Join(fresh, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fresh, "src", "in.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, fresh, zone.InboxSlot, `
work: a claims-less behavior
intents:
    - {lines: [D1], outcome: work, prose: "the tool and its check, told again"}
requirements:
    - id: R1
      lines: [D1]
      prose: the tool and its check
behaviors:
    - id: tool/loose
      name: Loose
      prose: Stands on nothing it names.
      covers: [R1]
      rows:
        - run: sh -c 'printf ok'
          expect:
            out: ok
`)
	runIn(t, fresh)
	writeZone(t, fresh, zone.InboxSlot, "done: true\n")
	runIn(t, fresh)
	if report = zoneFile(t, fresh, zone.ReportFile); !strings.Contains(report, "verdict: not yet — uncovered lines: AGENTS.md 9 (1-9), src/in.txt 1 (1)") {
		t.Fatalf("a claims-less corpus must leave the tree open by name:\n%s", report)
	}
}

// TestCoverageUniverseNamesItsOutside — a file the walk cannot read
// stands outside the coverage universe, named by its reason: never a
// silent hole, never a verdict holder. The report line caps at the
// first ten files and counts the rest.
func TestCoverageUniverseNamesItsOutside(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	// An unreadable file among the tree's own: the verdict still
	// closes — the outside file holds nothing.
	if err := os.WriteFile(filepath.Join(root, "locked.txt"), []byte("secret\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Join(root, "locked.txt"), 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(filepath.Join(root, "locked.txt"), 0o644) })
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if state := zoneFile(t, root, zone.StateFile); !strings.Contains(state, "verdict: #1") {
		t.Fatalf("an outside file must not hold the verdict, state:\n%s", state)
	}

	// Twelve open files: the first ten named, the rest counted.
	fresh := newRoot(t)
	if err := os.MkdirAll(filepath.Join(fresh, "junk"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(fresh, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 11; i++ {
		if err := os.WriteFile(filepath.Join(fresh, "junk", fmt.Sprintf("j%02d.txt", i)), []byte("x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(fresh, "src", "in.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, fresh, zone.InboxSlot, `
work: a claims-less behavior
intents:
    - {lines: [D1], outcome: work, prose: "the tool and its check, told again"}
requirements:
    - id: R1
      lines: [D1]
      prose: the tool and its check
behaviors:
    - id: tool/loose
      name: Loose
      prose: Stands on nothing it names.
      covers: [R1]
      rows:
        - run: sh -c 'printf ok'
          expect:
            out: ok
`)
	runIn(t, fresh)
	writeZone(t, fresh, zone.InboxSlot, "done: true\n")
	runIn(t, fresh)
	report := zoneFile(t, fresh, zone.ReportFile)
	if !strings.Contains(report, "verdict: not yet — uncovered lines: AGENTS.md 9 (1-9), junk/j00.txt 1 (1)") {
		t.Fatalf("the open files must be named in path order:\n%s", report)
	}
	if !strings.Contains(report, ", and 3 more files") {
		t.Fatalf("the twelfth open file must be counted, not printed:\n%s", report)
	}
}

// TestAwaitingOrderHoldsVerdict — an order awaiting the human holds
// the verdict by its own demand and question; the answer enters the
// stream as the human's words and the order closes by the remove of
// its line — the verdict lands.
func TestAwaitingOrderHoldsVerdict(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	writeZone(t, root, zone.InboxSlot, `
work: a question stands before the human
open:
    - {demand: D1, question: which color, state: awaiting}
`)
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "verdict: not yet — awaiting order D1: which color") {
		t.Fatalf("the awaiting order must hold the verdict by name:\n%s", report)
	}
	// The answer arrives as the human's words; the order closes over
	// the answer's own line, cited by the decision note — the verdict
	// lands.
	writeZone(t, root, zone.DemandsSlot, "a checked tool\nthe green one\n")
	writeZone(t, root, zone.InboxSlot, `work: the human answered, the order closes
intents:
    - {lines: [D2], outcome: mark, prose: the answer of the human carries no order}
remove: [D1]
not-a-requirement:
    - lines: [D2]
      reason: the human's answer, not new work
notes:
    - kind: decision
      text: the color is green by the human's answer
      refs: [D2]
`)
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if state := zoneFile(t, root, zone.StateFile); !strings.Contains(state, "verdict: #1") {
		t.Fatalf("the closed order must let the verdict land, state:\n%s", state)
	}
}

// TestDeferredOrderVisible — a deferred order never holds the
// verdict and stays visible forever: the state carries its demand,
// question and reason.
func TestDeferredOrderVisible(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	writeZone(t, root, zone.InboxSlot, `
work: the human said later
open:
    - {demand: D1, question: which color, state: deferred, reason: the human said later}
`)
	runIn(t, root)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "deferred orders:\n- D1 — which color — the human said later") {
		t.Fatalf("the deferred order must stand visible with its reason, state:\n%s", state)
	}
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if state = zoneFile(t, root, zone.StateFile); !strings.Contains(state, "verdict: #1") {
		t.Fatalf("a deferred order must not hold the verdict, state:\n%s", state)
	}
	if !strings.Contains(state, "deferred orders:") {
		t.Fatalf("the deferred order stays visible under the verdict, state:\n%s", state)
	}
}

// TestQueueLineBothStates — the queue is a fact of the files: the
// inbox stands or not, one line, both spellings.
func TestQueueLineBothStates(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	runIn(t, root)
	if state := zoneFile(t, root, zone.StateFile); !strings.Contains(state, "queue: empty\n") {
		t.Fatalf("the empty queue must say so, state:\n%s", state)
	}
	writeZone(t, root, zone.InboxSlot, "\tbroken: [")
	runIn(t, root)
	if state := zoneFile(t, root, zone.StateFile); !strings.Contains(state, "queue: submission waiting\n") {
		t.Fatalf("the refused submission must stand in the queue, state:\n%s", state)
	}
}

// TestVerdictOpenOnDrift — a claimed line edited between runs opens
// the verdict by the run that first saw the move; the terminal
// instructs the re-earn. A journal that cannot read at render time
// still names the drift — by the current run.
func TestVerdictOpenOnDrift(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if state := zoneFile(t, root, zone.StateFile); !strings.Contains(state, "verdict: #1") {
		t.Fatalf("the fixture must close its verdict, state:\n%s", state)
	}
	// The tree moves between runs: the row re-executes over the
	// moved line and turns red — the verdict opens by this run.
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("moved"), 0o644); err != nil {
		t.Fatal(err)
	}
	term := runIn(t, root)
	if !strings.Contains(term, "verdict drifted since run 2 — re-earn it: submit after that run, then write done: true") {
		t.Fatalf("the terminal must name the witness run and instruct the re-earn, got: %q", term)
	}
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "open (drift since run 2)") {
		t.Fatalf("the verdict must open by the run that saw the move, state:\n%s", state)
	}
	if next := lastNext(state); !strings.Contains(next, "verdict drifted since run 2") {
		t.Fatalf("the state's next must name the same witness run: %q", next)
	}
	// The journal's records, read once, keep naming the drift's own
	// first witness even over a disk that turns the epochs
	// unreadable past the open: the render never re-reads.
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("again"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, _ := runCode(t, root, map[string]bool{"read:epoch-!after-write": true})
	if code != 0 {
		t.Fatalf("a render over the once-read records must not die on the disk")
	}
	if state = zoneFile(t, root, zone.StateFile); !strings.Contains(state, "open (drift since run 2)") {
		t.Fatalf("the once-read records keep the drift's first witness, state:\n%s", state)
	}
}

// TestDriftSinceNamesTheFirstWitness — the drift line names the
// first run past the verdict that saw a certified behavior's claims
// move: an earlier re-execution over the certified lines is not a
// witness, and the tree's own move is.
func TestDriftSinceNamesTheFirstWitness(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	// A second behavior whose commentary row rides before its pinned
	// proof, over its own file.
	if err := os.WriteFile(filepath.Join(root, "src", "b.txt"), []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, fmt.Sprintf(`
work: the second behavior
uncovered-lines:
    - {range: "AGENTS.md:1-9", reason: the machine's own hook, refs: [D1]}
    - {range: "out.txt:1-1", reason: the row's own output, refs: [D1]}
    - {range: "b_out.txt:1-1", reason: the row's own output, refs: [D1]}
behaviors:
    - id: tool/bee
      name: Bee
      prose: Comments, then pins.
      claims:
        - src/b.txt:1-1
      covers: [R1]
      rows:
        - run: cat src/b.txt
          expect:
            out: "contains: b"
        - run: sh -c 'printf "%%s %%s\n" %s %s; cat src/b.txt > b_out.txt'
          expect:
            files:
                b_out.txt: b
`, claimsTokenOf(t, root, "src/b.txt:1-1"), demandTokenOf("a checked tool")))
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if state := zoneFile(t, root, zone.StateFile); !strings.Contains(state, "verdict: #1") {
		t.Fatalf("the fixture must close its verdict, state:\n%s", state)
	}
	// A row edit over the certified lines: the fact re-records on the
	// verdict's own pointer — not a witness of any drift.
	writeZone(t, root, zone.InboxSlot, fmt.Sprintf(`
work: the row edit over certified lines
mandate: [{line: D1, target: tool/copy}]
behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies.
      claims:
        - src/in.txt:1-1
      covers: [R1]
      rows:
        - run: sh -c 'printf "%%s %%s\n" %s %s; cat src/in.txt > out.txt'
          expect:
            rc: 0
            files:
                out.txt: ok
`, claimsTokenOf(t, root, "src/in.txt:1-1"), demandTokenOf("a checked tool")))
	runIn(t, root)
	if state := zoneFile(t, root, zone.StateFile); !strings.Contains(state, "— standing") {
		t.Fatalf("a row edit over certified lines is no drift, state:\n%s", state)
	}
	// A row edit of the second behavior: its fact re-records too, on
	// its own certified lines — still no witness.
	writeZone(t, root, zone.InboxSlot, fmt.Sprintf(`
work: the bee row edit
mandate: [{line: D1, target: tool/bee}]
behaviors:
    - id: tool/bee
      name: Bee
      prose: Comments, then pins.
      claims:
        - src/b.txt:1-1
      covers: [R1]
      rows:
        - run: cat src/b.txt
          expect:
            out: "contains: b"
        - run: sh -c 'printf "%%s %%s\n" %s %s; cat src/b.txt > b_out.txt'
          expect:
            rc: 0
            files:
                b_out.txt: b
`, claimsTokenOf(t, root, "src/b.txt:1-1"), demandTokenOf("a checked tool")))
	runIn(t, root)
	if state := zoneFile(t, root, zone.StateFile); !strings.Contains(state, "— standing") {
		t.Fatalf("a row edit of the second behavior is no drift either, state:\n%s", state)
	}
	// The tree moves: the run that first sees it is the witness.
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("moved"), 0o644); err != nil {
		t.Fatal(err)
	}
	runIn(t, root)
	if state := zoneFile(t, root, zone.StateFile); !strings.Contains(state, "open (drift since run 5)") {
		t.Fatalf("the drift must name its first witness run, state:\n%s", state)
	}
}

// TestRetiredScopeNeverDrifts — the verdict's record may outlive a
// behavior it certified: the retired id carries no current proof and
// never drifts; the living corpus answers for itself.
func TestRetiredScopeNeverDrifts(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if state := zoneFile(t, root, zone.StateFile); !strings.Contains(state, "verdict: #1") {
		t.Fatalf("the fixture must close its verdict, state:\n%s", state)
	}
	writeZone(t, root, zone.InboxSlot, `
work: the copy retires
mandate: [{line: D1, target: tool/copy}]
uncovered-lines:
    - {range: "AGENTS.md:1-9", reason: the machine's own hook, refs: [D1]}
    - {range: "out.txt:1-1", reason: the row's own output, refs: [D1]}
remove: [tool/copy]
`)
	runIn(t, root)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "— standing") {
		t.Fatalf("a retired scope never drifts, state:\n%s", state)
	}
}

// TestReplayEighteenHoles — the acceptance replay: one synthetic
// zone, twenty-six steps in the problems' own order, every step
// ending in a refusal, a verdict block or a loud state line. The
// chain closes in step ten and again over the outcome economy
// (steps twenty and twenty-two through twenty-five); everything
// else tries to cheat and is named.
func TestReplayEighteenHoles(t *testing.T) {
	mountPorts()
	root := t.TempDir()
	writeZone(t, root, zone.DemandsSlot, "count the lines of count.txt\n")
	if err := os.WriteFile(filepath.Join(root, "count.txt"), []byte("one\ntwo\nthree\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runIn(t, root)

	// 1 — the words of the tail belong to the machine's own taking:
	// a submission that meets an undeclared tail line is refused by
	// the line's own number.
	writeZone(t, root, zone.DemandsSlot, "count the lines of count.txt\nan undeclared word\n")
	writeZone(t, root, zone.InboxSlot, `work: the count begins
behaviors:
    - id: tool/count
      name: Count
      prose: Counts.
      claims:
        - count.txt:1-3
      rows:
        - run: sh -c "wc -l < count.txt | tr -d ' '"
          expect:
            out: "3\n"
`)
	runIn(t, root)
	if got := zoneFile(t, root, zone.ReportFile); !strings.Contains(got,
		"intents: the new demand lines D2 carry no outcome — declare answer, execute, describe, work, mark or ask (law:demand)") {
		t.Fatalf("step 1 must refuse the undeclared tail line:\n%s", got)
	}
	// The unjournaled tail returns to the journaled past alone: the
	// stray word leaves the file as quietly as it came.
	writeZone(t, root, zone.DemandsSlot, "count the lines of count.txt\n")

	// 2 — the requirement carries no content: its prose is empty.
	writeZone(t, root, zone.InboxSlot, `
work: the first account
requirements:
    - id: R1
      lines: [D1]
      prose: " "
behaviors:
    - id: tool/count
      name: Count
      prose: Counts.
      claims:
        - count.txt:1-3
      rows:
        - run: sh -c "wc -l < count.txt | tr -d ' '"
          expect:
            out: "3\n"
`)
	runIn(t, root)
	if got := zoneFile(t, root, zone.ReportFile); !strings.Contains(got,
		"requirements[0].prose: the prose is empty — unwind the demand lines in human words (law:demand)") {
		t.Fatalf("step 2 must refuse the empty requirement by its prose:\n%s", got)
	}

	// 3 — the marker without its reason marks nothing.
	writeZone(t, root, zone.InboxSlot, `
work: marking the demand away
not-a-requirement:
    - lines: [D1]
      reason: " "
`)
	runIn(t, root)
	if got := zoneFile(t, root, zone.ReportFile); !strings.Contains(got,
		"not-a-requirement[0].reason: the reason is empty — a mark without a reason marks nothing (law:document)") {
		t.Fatalf("step 3 must refuse the reasonless marker:\n%s", got)
	}

	// 4 — consents live in the agent's head: the honest submission
	// stands first, then a replacement without a mandate is refused.
	writeZone(t, root, zone.InboxSlot, `
work: the first account
intents:
    - {lines: [D1], outcome: work, prose: the counter and its check of the file}
requirements:
    - id: R1
      lines: [D1]
      prose: the counter and its check
uncovered-lines:
    - {range: "AGENTS.md:1-9", reason: the machine's own hook, refs: [D1]}
behaviors:
    - id: tool/count
      name: Count
      prose: Counts the file's lines, pinned by exact bytes.
      claims:
        - count.txt:1-3
      covers: [R1]
      rows:
        - run: sh -c "wc -l < count.txt | tr -d ' '"
          expect:
            out: "3\n"
`)
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, `
work: the silent replacement
behaviors:
    - id: tool/count
      name: Count anew
      prose: Counts differently.
      claims:
        - count.txt:1-3
      covers: [R1]
      rows:
        - run: sh -c "wc -l < count.txt | tr -d ' '"
          expect:
            out: "3\n"
`)
	runIn(t, root)
	if got := zoneFile(t, root, zone.ReportFile); !strings.Contains(got,
		"mandate: the submission replaces standing behaviors without a mandate — name the demand lines that ordered it (law:candor)") {
		t.Fatalf("step 4 must refuse the headless consent:\n%s", got)
	}

	// 5 — unfinished orders invisible: the awaiting order holds the
	// verdict by name; it closes only over a later human line, so it
	// stands re-listed until the answer's own line arrives (step 10).
	writeZone(t, root, zone.InboxSlot, `
work: a question stands
open:
    - {demand: D1, question: which unit?, state: awaiting}
`)
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if got := zoneFile(t, root, zone.ReportFile); !strings.Contains(got,
		"verdict: not yet — awaiting order D1: which unit?") {
		t.Fatalf("step 5 must hold the verdict on the open order:\n%s", got)
	}
	writeZone(t, root, zone.InboxSlot, `
work: the question stands before the human
open:
    - {demand: D1, question: which unit?, state: awaiting}
`)
	runIn(t, root)

	// 6 — globs fake coverage: a claim that is not a line range is
	// refused.
	writeZone(t, root, zone.InboxSlot, `
work: the glob claim
behaviors:
    - id: tool/glob
      name: Glob
      prose: Claims everything.
      claims:
        - "count.txt/**"
      covers: [R1]
      rows:
        - run: "true"
          expect:
            rc: 0
            out: ""
`)
	runIn(t, root)
	if got := zoneFile(t, root, zone.ReportFile); !strings.Contains(got,
		`behaviors[0].claims[0]: "count.txt/**" is not a line range the tree carries (law:claims)`) {
		t.Fatalf("step 6 must refuse the glob:\n%s", got)
	}

	// 7 — proof is an event: contains-only rows are commentary.
	writeZone(t, root, zone.InboxSlot, `
work: the commentary row
behaviors:
    - id: tool/event
      name: Event
      prose: Contains only.
      claims:
        - count.txt:1-3
      covers: [R1]
      rows:
        - run: sh -c "wc -l < count.txt | tr -d ' '"
          expect:
            out: "contains: 3"
`)
	runIn(t, root)
	if got := zoneFile(t, root, zone.ReportFile); !strings.Contains(got,
		"behaviors[0].rows: no pinned row — rows are commentary without exact bytes; pin the output of one row: out or files with exact bytes (\"contains: \" prefixes a substring); rc alone pins nothing (law:proof)") {
		t.Fatalf("step 7 must refuse the event proof:\n%s", got)
	}

	// 8 — no mutuality: a line no behavior claims and no exclusion
	// names holds the verdict, named per file.
	if err := os.WriteFile(filepath.Join(root, "extra.txt"), []byte("loose\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if got := zoneFile(t, root, zone.ReportFile); !strings.Contains(got,
		"verdict: not yet — uncovered lines: extra.txt 1 (1)") {
		t.Fatalf("step 8 must hold the verdict on the unclaimed line:\n%s", got)
	}
	writeZone(t, root, zone.InboxSlot, `
work: the loose line stands excluded, loudly
uncovered-lines:
    - {range: "AGENTS.md:1-9", reason: the machine's own hook, refs: [D1]}
    - {range: "extra.txt:1-1", reason: a stray note of the project, refs: [D1]}
`)
	runIn(t, root)

	// 9 — specs bend silently with code: a mandate line that does not
	// stand is refused.
	writeZone(t, root, zone.InboxSlot, `
work: the invented consent
mandate: [{line: D9, target: tool/count}]
behaviors:
    - id: tool/count
      name: Count
      prose: Counts the file's lines, pinned by exact bytes.
      claims:
        - count.txt:1-3
      covers: [R1]
      rows:
        - run: sh -c "wc -l < count.txt | tr -d ' '"
          expect:
            out: "3\n"
`)
	runIn(t, root)
	if got := zoneFile(t, root, zone.ReportFile); !strings.Contains(got,
		"mandate[0].line: the demand stream carries no line D9 (law:candor)") {
		t.Fatalf("step 9 must refuse the invented consent:\n%s", got)
	}

	// 10 — completeness on the agent's conscience: the whole chain —
	// words to the file, quote, cover, claim, pin, work — closes the
	// verdict. The awaiting order closes over the answer's own line,
	// cited by the decision note.
	writeZone(t, root, zone.DemandsSlot, "count the lines of count.txt\nprint the count\n")
	writeZone(t, root, zone.InboxSlot, `work: the second wish accounted whole
intents:
    - {lines: [D2], outcome: work, prose: the count printed for the human}
notes:
    - kind: decision
      text: the unit question closes over the printed answer
      refs: [D2]
remove: [D1]
requirements:
    - id: R2
      lines: [D2]
      prose: the count reaches the human as output
behaviors:
    - id: tool/print
      name: Print
      prose: Prints the count, pinned by exact bytes.
      claims:
        - count.txt:2-2
      covers: [R2]
      rows:
        - run: sh -c "echo 3"
          expect:
            out: "3\n"
`)
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if got := zoneFile(t, root, zone.StateFile); !strings.Contains(got, "verdict: #1 on v5, run 2") {
		t.Fatalf("step 10 must record the verdict over the closed chain, state:\n%s", got)
	}

	// 11 — notes carry obligations: a reference to nothing is refused.
	writeZone(t, root, zone.InboxSlot, `
work: a note cites nothing
notes:
    - kind: decision
      text: cites a ghost
      refs: [R7]
`)
	runIn(t, root)
	if got := zoneFile(t, root, zone.ReportFile); !strings.Contains(got,
		`notes[0].refs[0]: "R7" resolves to nothing — a reference names a demand, a requirement, a behavior, a line range or a law (law:candor)`) {
		t.Fatalf("step 11 must refuse the ghost reference:\n%s", got)
	}

	// 12 — lessons blind to real fixes: the narrative is the gate.
	writeZone(t, root, zone.InboxSlot, `
work: " "
`)
	runIn(t, root)
	if got := zoneFile(t, root, zone.ReportFile); !strings.Contains(got,
		"work: the narrative is empty — tell what broke, what was fixed, what was decided (law:candor)") {
		t.Fatalf("step 12 must refuse the empty narrative:\n%s", got)
	}

	// 13 — self-rules improvised: only the printed laws resolve.
	writeZone(t, root, zone.InboxSlot, `
work: the invented rule
notes:
    - kind: directive
      text: follows a law the contract does not print
      refs: [law:invented]
`)
	runIn(t, root)
	if got := zoneFile(t, root, zone.ReportFile); !strings.Contains(got,
		`notes[0].refs[0]: "law:invented" resolves to nothing`) {
		t.Fatalf("step 13 must refuse the invented law:\n%s", got)
	}

	// 14 — drift never unstands a verdict: the tree moves between
	// runs, the verdict opens by its run and the terminal re-earns.
	if err := os.Remove(filepath.Join(root, zone.Dirname, zone.InboxSlot)); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "count.txt"), []byte("one\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	term := runIn(t, root)
	if !strings.Contains(term, "verdict drifted since run 3 — re-earn it: submit after that run, then write done: true") {
		t.Fatalf("step 14 must name the witness run and instruct the re-earn, got: %q", term)
	}
	if got := zoneFile(t, root, zone.StateFile); !strings.Contains(got, "open (drift since run 3)") {
		t.Fatalf("step 14 must open the verdict by its run, state:\n%s", got)
	}
	if err := os.WriteFile(filepath.Join(root, "count.txt"), []byte("one\ntwo\nthree\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runIn(t, root)

	// 15 — inbox lifecycle undocumented: the refused submission waits
	// in the queue, the accepted one empties it.
	writeZone(t, root, zone.InboxSlot, "\tbroken: [")
	runIn(t, root)
	if got := zoneFile(t, root, zone.StateFile); !strings.Contains(got, "queue: submission waiting\n") {
		t.Fatalf("step 15 must show the waiting submission, state:\n%s", got)
	}
	// The queue empties over the re-earned pins: the drift's red rows
	// heal by a submission whose rows hold their pins again.
	writeZone(t, root, zone.InboxSlot, `
work: the queue empties over the re-earned pins
mandate: [{line: D1, target: tool/count}, {line: D2, target: tool/print}]
behaviors:
    - id: tool/count
      name: Count
      prose: Counts the file's lines, pinned by exact bytes.
      claims:
        - count.txt:1-3
      covers: [R1]
      rows:
        - run: sh -c "wc -l < count.txt | tr -d ' '"
          expect:
            out: "3\n"
    - id: tool/print
      name: Print
      prose: Prints the count, pinned by exact bytes.
      claims:
        - count.txt:2-2
      covers: [R2]
      rows:
        - run: sh -c "echo 3"
          expect:
            out: "3\n"
`)
	runIn(t, root)
	if got := zoneFile(t, root, zone.StateFile); !strings.Contains(got, "queue: empty\n") {
		t.Fatalf("step 15 must show the empty queue, state:\n%s", got)
	}

	// 16 — the machine parallel to VCS: the journal's chain holds the
	// past; a touched past line is zone damage, the restored bytes
	// recompose.
	writeZone(t, root, zone.DemandsSlot, "COUNT THE LINES OF count.txt\nprint the count\n")
	var damaged strings.Builder
	if code := Run(root, &damaged); code != 1 || !strings.Contains(damaged.String(),
		"punchtape: zone damaged: the demand stream's past changed") {
		t.Fatalf("step 16 must refuse the touched past: code %d term %q", code, damaged.String())
	}
	writeZone(t, root, zone.DemandsSlot, "count the lines of count.txt\nprint the count\n")
	runIn(t, root)

	// 17 — the knowledge page carries the note with its references
	// inline; the reverse index is gone, and no word matching
	// survives anywhere.
	writeZone(t, root, zone.InboxSlot, `
work: a noted decision
notes:
    - kind: decision
      text: the count is pinned by its exact bytes
      refs: [D1, R1, tool/count]
`)
	runIn(t, root)
	knowledge := zoneFile(t, root, zone.KnowledgeFile)
	if !strings.Contains(knowledge, "- [n2] (decision; refs: D1, R1, tool/count) the count is pinned by its exact bytes") {
		t.Fatalf("step 17 must carry the note with its references inline, knowledge:\n%s", knowledge)
	}
	if strings.Contains(knowledge, "## index") || strings.Contains(knowledge, "(tags:") || strings.Contains(knowledge, "untagged") {
		t.Fatalf("step 17 must carry no index and no word tags, knowledge:\n%s", knowledge)
	}

	// 18 — the state is the human-auditable face: every map on one
	// page, the standing named whole.
	state := zoneFile(t, root, zone.StateFile)
	for _, want := range []string{
		"demands: 2/2 lines; requirements 2 standing, 2 covered; markers: 0",
		"queue: empty\n",
		"corpus: v7, 2 behaviors",
		"submissions: 7 accepted, 11 refused, 1 completed",
		"machine at birth: " + Version,
		"verdict: #1 on v5, run 2, signature ",
		"- tool/count: 1 rows — 1 green (claims 1 files / 3 lines; proof ",
		"- tool/print: 1 rows — 1 green (claims 1 files / 1 lines; proof ",
		"excluded lines: 2 ranges",
		"knowledge: 2 notes, ",
		"\nnext: ",
	} {
		if !strings.Contains(state, want) {
			t.Fatalf("step 18 must carry %q on the auditable page, state:\n%s", want, state)
		}
	}

	// 19 — a new line without a declared outcome: the submission
	// brings the words and no intent for them.
	writeZone(t, root, zone.DemandsSlot, "count the lines of count.txt\nprint the count\nCheck everything once more.\n")
	writeZone(t, root, zone.InboxSlot, `work: the third wish arrives undeclared
`)
	runIn(t, root)
	if got := zoneFile(t, root, zone.ReportFile); !strings.Contains(got,
		"intents: the new demand lines D3 carry no outcome — declare answer, execute, describe, work, mark or ask (law:demand)") {
		t.Fatalf("step 19 must refuse the undeclared line:\n%s", got)
	}

	// 20 — the tree lock and the machine-run execution: the audit's
	// describe and the version's execute ride one submission; the
	// execution leaves its junk behind and the lock names it.
	writeZone(t, root, zone.DemandsSlot, "count the lines of count.txt\nprint the count\nCheck everything once more.\nShow the tool's version.\n")
	writeZone(t, root, zone.InboxSlot, `work: the third wish is a check, the fourth a run
intents:
    - {lines: [D3], outcome: describe, prose: "the owner asks for a check, no code change"}
    - {lines: [D4], outcome: execute, prose: the owner asks for the command output, run: printf junk > junk.tmp}
requirements:
    - id: R3
      lines: [D3]
      prose: the count check the owner demanded
behaviors:
    - id: tool/audit
      name: Audit
      prose: Audits the count, pinned by exact bytes.
      claims:
        - count.txt:3-3
      covers: [R3]
      rows:
        - run: sh -c "printf 'count '; wc -l < count.txt | tr -d ' '"
          expect:
            out: "count 3\n"
`)
	runIn(t, root)
	records, err := journalRecords(t, root)
	if err != nil {
		t.Fatal(err)
	}
	var execFact, lockFact bool
	for _, rec := range records {
		if rec.Kind == journal.KindExecution && rec.Execution != nil &&
			rec.Execution.Exit == 0 && strings.Join(rec.Execution.Lines, ",") == "D4" {
			execFact = true
		}
		if rec.Kind == journal.KindTreeLock && rec.TreeLock != nil {
			lockFact = true
		}
	}
	if !execFact {
		t.Fatal("step 20 must journal the execution fact with exit 0")
	}
	if !lockFact {
		t.Fatal("step 20 must journal the tree lock fact over the junk the execution left")
	}
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if got := zoneFile(t, root, zone.ReportFile); !strings.Contains(got,
		"verdict: not yet — the tree changed under answer/execute/describe intents (law:demand)") {
		t.Fatalf("step 20 must hold the verdict on the moved tree:\n%s", got)
	}
	// The junk goes away by hand: the tree returns to the lock's own
	// before-digest, the hold lifts, the fact stands in the journal.
	if err := os.Remove(filepath.Join(root, "junk.tmp")); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if state = zoneFile(t, root, zone.StateFile); !strings.Contains(state, "verdict: #2") {
		t.Fatalf("step 20 must record the verdict once the tree returned, state:\n%s", state)
	}
	if records, err = journalRecords(t, root); err != nil {
		t.Fatal(err)
	}
	lockFact = false
	for _, rec := range records {
		if rec.Kind == journal.KindTreeLock {
			lockFact = true
		}
	}
	if !lockFact {
		t.Fatal("step 20's tree lock fact must stand in the journal forever")
	}

	// 21 — the wc-l row earns its green in one run: clean bytes in
	// the pin, no cycle constant, no learn loop; the honest row of
	// step 20 stands green beside it — the control.
	writeZone(t, root, zone.InboxSlot, `
work: the count by another road
behaviors:
    - id: tool/lazy
      name: The other road
      prose: Counts by the sequence, pinned by exact bytes.
      claims:
        - count.txt:1-1
      covers: [R1]
      rows:
        - run: sh -c "seq 3 | wc -l | tr -d ' '"
          expect:
            out: "3\n"
`)
	runIn(t, root)
	if _, err := os.Stat(filepath.Join(root, zone.Dirname, zone.InboxSlot)); !os.IsNotExist(err) {
		t.Fatalf("step 21's clean row must be accepted in one run")
	}
	if state = zoneFile(t, root, zone.StateFile); !strings.Contains(state, "- tool/lazy: 1 rows — 1 green") {
		t.Fatalf("step 21's clean row must stand green on the first run, state:\n%s", state)
	}
	if !strings.Contains(state, "- tool/audit: 1 rows — 1 green") {
		t.Fatalf("step 21's honest row must stand green — the control, state:\n%s", state)
	}

	// 22 — the awaiting order: it opens on a demand line, it refuses
	// to close without a later human line, and it closes over the
	// answer's own line cited by a decision note; the lazy tool
	// retires by the same submission's mandate.
	writeZone(t, root, zone.InboxSlot, `
work: a question stands on the count's unit
open:
    - {demand: D1, question: what unit is the count in?, state: awaiting}
`)
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, `
work: the question tries to close itself
remove: [D1]
`)
	runIn(t, root)
	if got := zoneFile(t, root, zone.ReportFile); !strings.Contains(got,
		"open[0]: the order was closed without a later human line — bring the human's answer into demands.md, given its outcome, cited by a decision note (law:candor)") {
		t.Fatalf("step 22 must refuse the self-closed order:\n%s", got)
	}
	writeZone(t, root, zone.DemandsSlot, "count the lines of count.txt\nprint the count\nCheck everything once more.\nShow the tool's version.\nThe line is the unit.\n")
	writeZone(t, root, zone.InboxSlot, `work: the answer's line closes the question, the lazy tool retires
intents:
    - {lines: [D5], outcome: mark, prose: "the human's answer, not an order"}
not-a-requirement:
    - lines: [D5]
      reason: the answer to the unit question
remove: [D1, tool/lazy]
notes:
    - kind: decision
      text: the unit question closed over the human's line
      refs: [D1, D5]
mandate: [{line: D1, target: tool/lazy}]
`)
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if state = zoneFile(t, root, zone.StateFile); !strings.Contains(state, "verdict: #3") {
		t.Fatalf("step 22's act must record the verdict over the closed order, state:\n%s", state)
	}

	// 23 — words that came unmet: the submission over the undeclared
	// tail line is refused, and the act over the standing corpus
	// names the line.
	writeZone(t, root, zone.DemandsSlot, "count the lines of count.txt\nprint the count\nCheck everything once more.\nShow the tool's version.\nThe line is the unit.\nThanks.\n")
	writeZone(t, root, zone.InboxSlot, `
work: meeting words that came with no intent
`)
	runIn(t, root)
	if got := zoneFile(t, root, zone.ReportFile); !strings.Contains(got,
		"intents: the new demand lines D6 carry no outcome") {
		t.Fatalf("step 23 must refuse the unmet words:\n%s", got)
	}
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if got := zoneFile(t, root, zone.ReportFile); !strings.Contains(got,
		"verdict: not yet — unresolved lines: D6 (no outcome)") {
		t.Fatalf("step 23's act must name the undeclared tail line:\n%s", got)
	}

	// 24 — the drift holds the act until a submission names the
	// change: the tree moves by hand, the bare run witnesses, two
	// acts stay held; the bytes return ahead of the naming
	// submission, whose re-earned pins re-run green on them.
	if err = os.WriteFile(filepath.Join(root, "count.txt"), []byte("one\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runIn(t, root)
	if state = zoneFile(t, root, zone.StateFile); !strings.Contains(state, "open (drift since run ") {
		t.Fatalf("step 24 must open the verdict by its witness run, state:\n%s", state)
	}
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if got := zoneFile(t, root, zone.ReportFile); !strings.Contains(got,
		"verdict: not yet — drift unresolved: a submission must come after run 7") {
		t.Fatalf("step 24's act must stay held by the witness run:\n%s", got)
	}
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if got := zoneFile(t, root, zone.ReportFile); !strings.Contains(got, "drift unresolved: a submission must come after run 7") {
		t.Fatalf("step 24's repeated act without a submission must stay held by the same run:\n%s", got)
	}
	if err = os.WriteFile(filepath.Join(root, "count.txt"), []byte("one\ntwo\nthree\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, `work: the marker names the change, the pins re-earn on the returned bytes
intents:
    - {lines: [D6], outcome: mark, prose: "gratitude, not an order"}
not-a-requirement:
    - lines: [D5]
      reason: the answer to the unit question
    - lines: [D6]
      reason: "gratitude, not work"
mandate: [{line: D1, target: tool/count}, {line: D2, target: tool/print}, {line: D3, target: tool/audit}]
behaviors:
    - id: tool/count
      name: Count
      prose: Counts the file's lines, pinned by exact bytes.
      claims:
        - count.txt:1-3
      covers: [R1]
      rows:
        - run: sh -c "wc -l < count.txt | tr -d ' '"
          expect:
            out: "3\n"
    - id: tool/print
      name: Print
      prose: Prints the count, pinned by exact bytes.
      claims:
        - count.txt:2-2
      covers: [R2]
      rows:
        - run: sh -c "echo 3"
          expect:
            out: "3\n"
    - id: tool/audit
      name: Audit
      prose: Audits the count, pinned by exact bytes.
      claims:
        - count.txt:3-3
      covers: [R3]
      rows:
        - run: sh -c "printf 'count '; wc -l < count.txt | tr -d ' '"
          expect:
            out: "count 3\n"
`)
	runIn(t, root)
	if got := zoneFile(t, root, zone.ReportFile); !strings.Contains(got, "green: 3") {
		t.Fatalf("step 24's re-earned rows must re-run green on the returned bytes:\n%s", got)
	}
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if state = zoneFile(t, root, zone.StateFile); !strings.Contains(state, "verdict: #4") {
		t.Fatalf("step 24's act must record the verdict over the named drift, state:\n%s", state)
	}

	// 25 — the mandate pair in battle: the row replacement demands
	// its permission; a line the stream does not carry is refused by
	// name; the honest pair accepts and the verdict records.
	countRowDoc := func(mandate string) string {
		head := ""
		if mandate != "" {
			head = mandate + "\n"
		}
		return head + `
work: the count's row replaced by the honest row
behaviors:
    - id: tool/count
      name: Count
      prose: Counts the file's lines, pinned by exact bytes.
      claims:
        - count.txt:1-3
      covers: [R1]
      rows:
        - run: sh -c "grep -c . count.txt"
          expect:
            out: "3\n"
`
	}
	writeZone(t, root, zone.InboxSlot, countRowDoc(""))
	runIn(t, root)
	if got := zoneFile(t, root, zone.ReportFile); !strings.Contains(got,
		"mandate: the submission replaces standing behaviors without a mandate — name the demand lines that ordered it (law:candor)") {
		t.Fatalf("step 25 must refuse the headless row replacement:\n%s", got)
	}
	writeZone(t, root, zone.InboxSlot, countRowDoc("mandate: [{line: D99, target: tool/count}]"))
	runIn(t, root)
	if got := zoneFile(t, root, zone.ReportFile); !strings.Contains(got,
		"mandate[0].line: the demand stream carries no line D99 (law:candor)") {
		t.Fatalf("step 25 must refuse the invented mandate line:\n%s", got)
	}
	writeZone(t, root, zone.InboxSlot, countRowDoc("mandate: [{line: D1, target: tool/count}]"))
	runIn(t, root)
	if got := zoneFile(t, root, zone.ReportFile); !strings.Contains(got, "green: 1") {
		t.Fatalf("step 25's honest row must run green under its pair:\n%s", got)
	}
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if state = zoneFile(t, root, zone.StateFile); !strings.Contains(state, "verdict: #5") {
		t.Fatalf("step 25's honest pair must let the verdict record, state:\n%s", state)
	}

	// 26 — the declared universe: the vendored line stands outside the
	// counted roots, named aloud; the verdict does not stand over it.
	if err = os.MkdirAll(filepath.Join(root, "vendor"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(root, "vendor", "deps.txt"), []byte("vendored\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.UniverseSlot, "count.txt\n")
	runIn(t, root)
	state = zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "outside the universe: 3 files") {
		t.Fatalf("step 26 must name every file beyond the declared roots — the hook, the stray note, the vendored line:\n%s", state)
	}
	verdictLine := ""
	for _, line := range strings.Split(state, "\n") {
		if strings.HasPrefix(line, "verdict: #") {
			verdictLine = line
		}
	}
	if !strings.HasSuffix(verdictLine, " — standing") {
		t.Fatalf("step 26's verdict must not stand over the file outside the universe:\n%s", verdictLine)
	}
}

// TestVerdictWaitsUnresolvedTail — the verdict waits for the stream's
// every line to carry its outcome: a line without a declaration holds
// the act by name; the declaration closes it.
func TestVerdictWaitsUnresolvedTail(t *testing.T) {
	root := newBornRoot(t)
	writeZone(t, root, zone.DemandsSlot, "a checked tool\na loose word\n")
	writeZone(t, root, zone.InboxSlot, `
work: the signature over the loose word
`)
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "verdict: not yet — unresolved lines: D2 (no outcome)") {
		t.Fatalf("the undeclared line must hold the verdict by name:\n%s", report)
	}
	// The loose word declared mark over its marker: the verdict lands.
	writeZone(t, root, zone.InboxSlot, `work: the loose word marked away
intents:
    - {lines: [D2], outcome: mark, prose: the loose word orders nothing}
not-a-requirement:
    - lines: [D2]
      reason: talk, not work
`)
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if state := zoneFile(t, root, zone.StateFile); !strings.Contains(state, "verdict: #1") {
		t.Fatalf("the declared line must let the verdict land, state:\n%s", state)
	}
}

// TestCompletionBlockedUnderDrift — the act over an open drift is
// held until a submission comes after the drift's first witness; a
// bare repeat of the act changes nothing.
func TestCompletionBlockedUnderDrift(t *testing.T) {
	root := newBornRoot(t)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if state := zoneFile(t, root, zone.StateFile); !strings.Contains(state, "verdict: #1") {
		t.Fatalf("the fixture must close its verdict, state:\n%s", state)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("moved"), 0o644); err != nil {
		t.Fatal(err)
	}
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "verdict: not yet — drift unresolved: a submission must come after run 2") {
		t.Fatalf("the open drift must hold the act by the witness run:\n%s", report)
	}
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if report = zoneFile(t, root, zone.ReportFile); !strings.Contains(report, "drift unresolved: a submission must come after run 2") {
		t.Fatalf("a repeated act with no submission must stay held by the same run:\n%s", report)
	}
}

// TestDriftResolvedBySubmission — a submission that lands after the
// drift's witness names the change: the next act records the verdict
// anew over the returned bytes.
func TestDriftResolvedBySubmission(t *testing.T) {
	root := newBornRoot(t)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("moved"), 0o644); err != nil {
		t.Fatal(err)
	}
	runIn(t, root)
	// The bytes return, and a submission names the change with the
	// re-earned pin of the cycle.
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, fmt.Sprintf(`
work: the drift named by this submission
mandate: [{line: D1, target: tool/copy}]
behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies the input, pinned by its bytes.
      claims:
        - src/in.txt:1-1
      covers: [R1]
      rows:
        - run: sh -c 'printf "%%s %%s\n" %s %s; cat src/in.txt > out.txt'
          expect:
            files:
                out.txt: ok
`, claimsTokenOf(t, root, "src/in.txt:1-1"), demandTokenOf("a checked tool")))
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if state := zoneFile(t, root, zone.StateFile); !strings.Contains(state, "verdict: #2") {
		t.Fatalf("the named drift must let the verdict record anew, state:\n%s", state)
	}
}

// TestNextInstructionNamesHoles — the tree's open lines reach the
// instruction as the first verdict blocker: claim or exclude.
func TestNextInstructionNamesHoles(t *testing.T) {
	root := newBornRoot(t)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if err := os.WriteFile(filepath.Join(root, "stray.txt"), []byte("loose\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runIn(t, root)
	next := lastNext(zoneFile(t, root, zone.StateFile))
	if !strings.Contains(next, "1 uncovered lines — submit their claims or loud exclusions") {
		t.Fatalf("the tree's hole must be the spoken blocker, got: %q", next)
	}
}

// TestStateVerdictTailNamesUncovered — the verdict's own line carries
// the tree's uncovered count while it stands over holes.
func TestStateVerdictTailNamesUncovered(t *testing.T) {
	root := newBornRoot(t)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if err := os.WriteFile(filepath.Join(root, "stray.txt"), []byte("loose\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runIn(t, root)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, " — standing over 1 uncovered lines") {
		t.Fatalf("the verdict line must carry its uncovered tail, state:\n%s", state)
	}
}

// TestDriftHoldNamesTheWitness — the drift hold names its own fact:
// the report's hold line and the terminal's instruction both carry
// the witness run, the very run the state's verdict line already
// names — three faces, one number, nothing left to guess. The edit
// stays (the pin is independent of the moved line, exactly the live
// case): only a submission after the witness closes the drift.
func TestDriftHoldNamesTheWitness(t *testing.T) {
	mountPorts()
	root := t.TempDir()
	writeZone(t, root, zone.DemandsSlot, "a checked tool\n")
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, `
work: the pinned account
intents:
    - {lines: [D1], outcome: work, prose: the tool and its check}
requirements:
    - id: R1
      lines: [D1]
      prose: the tool and its check
uncovered-lines:
    - {range: "AGENTS.md:1-9", reason: the machine's own hook, refs: [D1]}
behaviors:
    - id: tool/pin
      name: The pin
      prose: Prints three, pinned by exact bytes.
      claims:
        - src/in.txt:1-1
      covers: [R1]
      rows:
        - run: sh -c "printf '3\n'"
          expect:
            out: "3\n"
`)
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if state := zoneFile(t, root, zone.StateFile); !strings.Contains(state, "verdict: #1") {
		t.Fatalf("the fixture must close its verdict, state:\n%s", state)
	}
	// The edit stays inside the claimed line; the pin is independent
	// of it — the row re-runs green, the verdict still opens by the
	// run that first saw the move.
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("moved"), 0o644); err != nil {
		t.Fatal(err)
	}
	term := runIn(t, root)
	if !strings.Contains(term, "verdict drifted since run 2 — re-earn it: submit after that run") {
		t.Fatalf("the terminal must name the witness run, got: %q", term)
	}
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "open (drift since run 2)") {
		t.Fatalf("the state's verdict line names the same run:\n%s", state)
	}
	if !strings.Contains(state, "tool/pin: 1 rows — 1 green") {
		t.Fatalf("the re-run row stands green — the pin is independent of the move:\n%s", state)
	}
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "verdict: not yet — drift unresolved: a submission must come after run 2") {
		t.Fatalf("the hold must name the witness run:\n%s", report)
	}
	// A submission after the witness names the drift: the next act
	// records the verdict anew over the moved bytes.
	writeZone(t, root, zone.InboxSlot, "work: the move named by this late submission\n")
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if state = zoneFile(t, root, zone.StateFile); !strings.Contains(state, "verdict: #2") {
		t.Fatalf("the submission after the witness must close the drift, state:\n%s", state)
	}
}

// TestFacesSpeakOneTruth — the terminal's pointer line and state.md's
// next: line are one computed string through every standing state,
// and the spoken blocker names the same fact the report's hold
// carries: the drift's witness run, the awaiting order, the
// unresolved line. A reason cannot stand voiceless or speak past its
// own expiry — the drift instruction dies with the naming
// submission, never after it.
func TestFacesSpeakOneTruth(t *testing.T) {
	mountPorts()
	root := t.TempDir()
	writeZone(t, root, zone.DemandsSlot, "a checked tool\n")
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, `
work: the pinned account
intents:
    - {lines: [D1], outcome: work, prose: the tool and its check}
requirements:
    - id: R1
      lines: [D1]
      prose: the tool and its check
uncovered-lines:
    - {range: "AGENTS.md:1-9", reason: the machine's own hook, refs: [D1]}
behaviors:
    - id: tool/pin
      name: The pin
      prose: Prints three, pinned by exact bytes.
      claims:
        - src/in.txt:1-1
      covers: [R1]
      rows:
        - run: sh -c "printf '3\n'"
          expect:
            out: "3\n"
`)
	runIn(t, root)
	agree := func(term string) string {
		t.Helper()
		next := lastNext(zoneFile(t, root, zone.StateFile))
		if !strings.HasPrefix(term, "punchtape: run ") || !strings.HasSuffix(strings.TrimRight(term, "\n"), next) {
			t.Fatalf("the terminal and the state's next line must be one text:\nterminal: %q\nnext: %q", term, next)
		}
		return next
	}
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if state := zoneFile(t, root, zone.StateFile); !strings.Contains(state, "verdict: #1") {
		t.Fatalf("the fixture must record its first verdict, state:\n%s", state)
	}

	// The drift: the witness run is spoken and held by one number.
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("moved"), 0o644); err != nil {
		t.Fatal(err)
	}
	next := agree(runIn(t, root))
	if !strings.Contains(next, "verdict drifted since run ") {
		t.Fatalf("the drift must be the spoken blocker, got: %q", next)
	}
	witness := regexp.MustCompile(`since run (\d+)`).FindStringSubmatch(next)
	if witness == nil {
		t.Fatalf("the spoken drift must name its witness run: %q", next)
	}
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "a submission must come after run "+witness[1]) {
		t.Fatalf("the hold must name the same witness the faces speak:\n%s", report)
	}

	// The naming submission: the drift instruction dies here — the
	// faces move to the next standing truth, never past it.
	writeZone(t, root, zone.InboxSlot, "work: the move named by this late submission\n")
	next = agree(runIn(t, root))
	if strings.Contains(next, "drifted") {
		t.Fatalf("a named drift must not be spoken again, got: %q", next)
	}
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if state := zoneFile(t, root, zone.StateFile); !strings.Contains(state, "verdict: #2") {
		t.Fatalf("the named drift must let the verdict record, state:\n%s", state)
	}

	// The awaiting order: spoken, held and closed by one name.
	writeZone(t, root, zone.DemandsSlot, "a checked tool\nwhich of the tools needs sharpening?\n")
	writeZone(t, root, zone.InboxSlot, `
work: the question stands open
intents:
    - {lines: [D2], outcome: ask, prose: the owner asks which tool to sharpen}
open:
    - {demand: D2, question: which of the tools needs sharpening?, state: awaiting}
`)
	next = agree(runIn(t, root))
	if !strings.Contains(next, "awaiting order D2") {
		t.Fatalf("the awaiting order must be the spoken blocker, got: %q", next)
	}
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if report = zoneFile(t, root, zone.ReportFile); !strings.Contains(report, "awaiting order D2") {
		t.Fatalf("the hold must name the same order the faces speak:\n%s", report)
	}

	// The order closes over a later human line; the verdict records.
	writeZone(t, root, zone.DemandsSlot, "a checked tool\nwhich of the tools needs sharpening?\nLater.\n")
	writeZone(t, root, zone.InboxSlot, `
work: the human answered the question
intents:
    - {lines: [D3], outcome: mark, prose: "the human answered: later"}
not-a-requirement:
    - {lines: [D3], reason: "an answer of deferral, not an order"}
open:
    - {demand: D2, question: which of the tools needs sharpening?, state: deferred, reason: "the human said Later. in D3"}
notes:
    - {kind: decision, text: "the question is deferred by the human's answer", refs: [D2, D3]}
`)
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if state := zoneFile(t, root, zone.StateFile); !strings.Contains(state, "verdict: #3") {
		t.Fatalf("the answered order must let the verdict record, state:\n%s", state)
	}

	// The unresolved line: a removed carrier speaks until it returns.
	writeZone(t, root, zone.DemandsSlot, "a checked tool\nwhich of the tools needs sharpening?\nLater.\njust a note.\n")
	writeZone(t, root, zone.InboxSlot, `
work: the marker line stands declared
intents:
    - {lines: [D4], outcome: mark, prose: "a note, not an order"}
not-a-requirement:
    - {lines: [D4], reason: "just a note"}
`)
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if state := zoneFile(t, root, zone.StateFile); !strings.Contains(state, "verdict: #4") {
		t.Fatalf("the declared marker must let the verdict record, state:\n%s", state)
	}
	writeZone(t, root, zone.InboxSlot, `
work: the marker leaves its line
remove: [D4]
`)
	next = agree(runIn(t, root))
	if !strings.Contains(next, "unresolved lines: D4 (mark pending)") {
		t.Fatalf("the orphaned outcome must be the spoken blocker, got: %q\nreport:\n%s", next, zoneFile(t, root, zone.ReportFile))
	}
	writeZone(t, root, zone.InboxSlot, `
work: the marker returns
not-a-requirement:
    - {lines: [D4], reason: "just a note"}
`)
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if state := zoneFile(t, root, zone.StateFile); !strings.Contains(state, "verdict: #5") {
		t.Fatalf("the returned carrier must let the verdict record, state:\n%s", state)
	}
}

// TestRedBuildHoldsVerdict — the printed first blocker exists: a red
// declared build holds the verdict, the report names it, the
// instruction says fix the build.
func TestRedBuildHoldsVerdict(t *testing.T) {
	root := newRoot(t)
	if err := os.WriteFile(filepath.Join(root, "f.txt"), []byte("one line of words\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, `
work: the green corpus
intents:
    - {lines: [D1], outcome: work, prose: the line is declared as work}
requirements:
    - {id: R1, lines: [D1], prose: the one line}
uncovered-lines:
    - {range: "AGENTS.md:1-9", reason: the machine's own hook, refs: [D1]}
behaviors:
    - id: b/one
      name: one line
      prose: the line stands verified
      covers: [R1]
      claims: ["f.txt:1-1"]
      rows:
        - {run: "cat f.txt", expect: {out: "one line of words\n"}}
`)
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "verdict: #1") {
		t.Fatalf("the green corpus must record its verdict:\n%s", state)
	}
	// The declared build that fails: the corpus moves, the build is red.
	writeZone(t, root, zone.InboxSlot, "work: declaring a failing build\nbuild: \"false\"\nnotes:\n  - {kind: work, text: the build is declared and it fails, refs: [D1]}\n")
	runIn(t, root)
	// The completion act over the red build: no second verdict.
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "build: RED — exit 1") {
		t.Fatalf("the report must name the red build:\n%s", report[:400])
	}
	if !strings.Contains(report, "verdict: not yet — the build is red") {
		t.Fatalf("the red build must hold the verdict by name:\n%s", report[:400])
	}
	state = zoneFile(t, root, zone.StateFile)
	if strings.Contains(state, "verdict: #2") {
		t.Fatalf("no verdict may stand over a red build:\n%s", state)
	}
	if !strings.Contains(state, "next: build red — read .punchtape/report.md, fix the build, run again") {
		t.Fatalf("the instruction must say fix the build:\n%s", state)
	}
}

// TestGreenBuildVerdictUnchanged — a green declared build records
// the verdict as before.
func TestGreenBuildVerdictUnchanged(t *testing.T) {
	root := newRoot(t)
	if err := os.WriteFile(filepath.Join(root, "f.txt"), []byte("one line of words\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, `
work: the green corpus with a green build
build: "true"
intents:
    - {lines: [D1], outcome: work, prose: the line is declared as work}
requirements:
    - {id: R1, lines: [D1], prose: the one line}
uncovered-lines:
    - {range: "AGENTS.md:1-9", reason: the machine's own hook, refs: [D1]}
behaviors:
    - id: b/one
      name: one line
      prose: the line stands verified
      covers: [R1]
      claims: ["f.txt:1-1"]
      rows:
        - {run: "cat f.txt", expect: {out: "one line of words\n"}}
`)
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "verdict: #1 on v1, run 2") {
		t.Fatalf("the green build must record its verdict:\n%s", state)
	}
}
