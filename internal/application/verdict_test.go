// The verdict's, the knowledge's and the discovery block's battery:
// the completion act closes a fully green corpus, the empty corpus
// is refused, the drift is named; each note stands once with its tags
// inline and the index gathers them, the lessons carry the row's
// bytes before and after; the AGENTS block stands idempotently; and
// the canon's example must pass validation and run green against the
// real machine, or the build fails.
package application

import (
	"os"
	"path/filepath"
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
	writeZone(t, root, zone.WishSlot, "a checked tool")
	runIn(t, root)
	return root
}

func TestVerdictClosesGreenCorpusAndRefusesEmpty(t *testing.T) {
	root := newRoot(t)

	// The empty corpus: the completion act is refused by name.
	writeZone(t, root, zone.InboxSlot, "{}\n")
	runIn(t, root)
	if report := zoneFile(t, root, zone.ReportFile); !strings.Contains(report, "verdict: refused — the corpus is empty") {
		t.Fatalf("the empty corpus must be refused by name, got:\n%s", report)
	}

	// A green corpus with its wish accounted: the completion act
	// records the verdict.
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, `requirements:
    - id: R1
      brief: [W1]
artifacts: [out.txt]
behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies.
      scope:
        - src/**
      covers: [R1]
      rows:
        - run: cat src/in.txt > out.txt
          expect:
            files:
                out.txt: "contains: ok"
`)
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, zoneFile(t, root, zone.CorpusFile))
	runIn(t, root)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "verdict: #1 on v1") || !strings.Contains(state, "standing") {
		t.Fatalf("the verdict must stand, state:\n%s", state)
	}

	// A scope file changes and stays changed: the mark names the
	// behavior whose certified scope the tree outgrew.
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("changed"), 0o644); err != nil {
		t.Fatal(err)
	}
	runIn(t, root)
	if state = zoneFile(t, root, zone.StateFile); !strings.Contains(state, "scopes moved: tool/copy") {
		t.Fatalf("the moved scope must name the behavior, state:\n%s", state)
	}
}

func TestVerdictNotYetWhenAFrozenRowIsStale(t *testing.T) {
	root := newBornRoot(t)
	// A broken build freezes the rows; a moved scope file leaves the
	// green fact on an old proof: the completion act says not yet.
	writeZone(t, root, zone.InboxSlot, "build: false\n"+`behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies.
      scope:
        - src/**
      covers: [R1]
      rows:
        - run: cat src/in.txt > out.txt
          expect:
            files:
                out.txt: "contains: ok"
`)
	runIn(t, root)
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("moved on"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, zoneFile(t, root, zone.CorpusFile))
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "verdict: not yet — a row is red or stale") {
		t.Fatalf("a stale frozen row must hold the verdict:\n%s", report)
	}
}

// wishFixture — a born instance of a three-line wish with one scoped
// green row writing out.txt; the wish's account is the caller's.
func wishFixture(t *testing.T) string {
	t.Helper()
	mountPorts()
	root := t.TempDir()
	writeZone(t, root, zone.WishSlot, "one\n\ntwo\n\nthree\n")
	runIn(t, root)
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, `artifacts: [out.txt]
behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies.
      scope:
        - src/**
      rows:
        - run: cat src/in.txt > out.txt
          expect:
            files:
                out.txt: "contains: ok"
`)
	runIn(t, root)
	return root
}

func TestVerdictNeedsTheWishAccountedWhole(t *testing.T) {
	root := wishFixture(t)
	// A green behavior-only corpus: the wish stands wholly uncovered.
	writeZone(t, root, zone.InboxSlot, zoneFile(t, root, zone.CorpusFile))
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "verdict: not yet — the wish lines W1..W3 stand uncovered by any requirement or marker") {
		t.Fatalf("the uncovered wish must hold the verdict by name:\n%s", report)
	}
	// The registry arrives: two lines as one requirement, the third
	// marked — but no behavior carries the requirement yet.
	writeZone(t, root, zone.InboxSlot, `requirements:
    - id: R1
      brief: [W1..W2]
not-a-requirement:
    - lines: [W3]
      quote: three
      reason: the closing line names no work
`)
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, zoneFile(t, root, zone.CorpusFile))
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
	writeZone(t, root, zone.InboxSlot, `behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies.
      scope:
        - src/**
      covers: [R1]
      rows:
        - run: cat src/in.txt > out.txt
          expect:
            files:
                out.txt: "contains: ok"
`)
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, zoneFile(t, root, zone.CorpusFile))
	runIn(t, root)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "verdict: #1") {
		t.Fatalf("the accounted corpus must close the verdict, state:\n%s", state)
	}
	if !strings.Contains(state, "brief: 3/3 lines; requirements 1 standing, 1 covered; not-a-requirement: 1") {
		t.Fatalf("the brief map must show the whole account, state:\n%s", state)
	}
}

func TestVerdictNeedsNoUncoveredSource(t *testing.T) {
	root := wishFixture(t)
	// A stray source no scope covers and no artifact declares.
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("words"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, `requirements:
    - id: R1
      brief: [W1..W3]
behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies.
      scope:
        - src/**
      covers: [R1]
      rows:
        - run: cat src/in.txt > out.txt
          expect:
            files:
                out.txt: "contains: ok"
`)
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, zoneFile(t, root, zone.CorpusFile))
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "verdict: not yet — the sources README.md stand uncovered") {
		t.Fatalf("the uncovered source must hold the verdict by name:\n%s", report)
	}
	// A behavior's scope closes the source: the verdict stands.
	writeZone(t, root, zone.InboxSlot, `behaviors:
    - id: tool/read
      name: The readme
      prose: The readme stands read.
      scope:
        - README.md
      rows:
        - run: "true"
          expect:
            rc: 0
`)
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, zoneFile(t, root, zone.CorpusFile))
	runIn(t, root)
	if !strings.Contains(zoneFile(t, root, zone.StateFile), "verdict: #1") {
		t.Fatalf("the closed tree must let the verdict stand, state:\n%s", zoneFile(t, root, zone.StateFile))
	}
}

// TestVerdictRecomputesFromTheJournal — the journal-only proof: an
// outside reader with the journal's records, the wish's object and
// the working tree re-spells the verdict's four conditions and the
// registry digest with the domain's own primitives, and reaches the
// machine's recorded verdict — zero unaccounted information.
func TestVerdictRecomputesFromTheJournal(t *testing.T) {
	root := wishFixture(t)
	writeZone(t, root, zone.InboxSlot, `requirements:
    - id: R1
      brief: [W1..W2]
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
      covers: [R1]
      rows:
        - run: cat src/in.txt > out.txt
          expect:
            files:
                out.txt: "contains: ok"
`)
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, zoneFile(t, root, zone.CorpusFile))
	runIn(t, root)
	if !strings.Contains(zoneFile(t, root, zone.StateFile), "verdict: #1") {
		t.Fatalf("the fixture must close its verdict")
	}

	// The reader's material: the journal's records, the object
	// store, the tree, the environment — nothing of the machine's
	// own standing.
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
	var birth *journal.Birth
	for _, rec := range records {
		if rec.Kind == journal.KindBirth && rec.Birth != nil {
			birth = rec.Birth
		}
	}
	wish, err := objects.Get(birth.Addr)
	if err != nil {
		t.Fatal(err)
	}

	// Condition 1: every row green on the verdict's own scope
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

	// Condition 2: the wish's every non-empty line inside some
	// requirement's span or some marker's span.
	var universe []int
	for _, line := range strings.Split(string(wish), "\n") {
		if strings.TrimSpace(line) != "" {
			universe = append(universe, len(universe)+1)
		}
	}
	claimed := map[int]bool{}
	for _, id := range st.Registry.Order {
		nums, err := doc.ParseRanges(st.Registry.Requirements[id].Brief)
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

	// Condition 4: no uncovered source in the tree.
	tree, err := TreeFiles(root)
	if err != nil {
		t.Fatal(err)
	}
	var globs []string
	for _, id := range st.Order {
		globs = append(globs, st.Behaviors[id].Scope...)
	}
	globs = append(globs, st.Registry.Artifacts...)
	for _, rel := range tree {
		if rel == agentsFile {
			continue
		}
		covered := false
		for _, g := range globs {
			if fprint.MatchGlob(g, rel) {
				covered = true
				break
			}
		}
		if !covered {
			t.Fatalf("condition 4 fails the recompute: %s", rel)
		}
	}

	// The digest: the registry's canonical sections, re-marshaled
	// here from the fold, must equal the record's own seal.
	var sections doc.Document
	for _, id := range st.Registry.Order {
		sections.Requirements = append(sections.Requirements, st.Registry.Requirements[id])
	}
	sections.NotRequirements = st.Registry.Markers
	if got := domain.Digest(doc.Marshal(sections)); got != st.Verdict.RegistryDigest {
		t.Fatalf("the registry digest must recompute: %s != %s", got, st.Verdict.RegistryDigest)
	}
}

func TestKnowledgeNotesTagsAndLessons(t *testing.T) {
	root := newRoot(t)
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("ok"), 0o644); err == nil {
		t.Fatal("src must not pre-exist here")
	}
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Red: the expectation does not hold.
	writeZone(t, root, zone.InboxSlot, `notes:
    - text: the copy must carry the input bytes
behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies.
      scope:
        - src/**
      rows:
        - run: cat src/in.txt > out.txt
          expect:
            files:
                out.txt: "contains: NOPE"
`)
	runIn(t, root)
	if report := zoneFile(t, root, zone.ReportFile); !strings.Contains(report, `want "contains: NOPE"`) {
		t.Fatalf("the row must be red first, report:\n%s", report)
	}

	// Green: the same behavior, the honest expectation — the lesson
	// lands with the row's bytes before and after.
	writeZone(t, root, zone.InboxSlot, `behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies.
      scope:
        - src/**
      rows:
        - run: cat src/in.txt > out.txt
          expect:
            files:
                out.txt: "contains: ok"
`)
	runIn(t, root)
	knowledge := zoneFile(t, root, zone.KnowledgeFile)
	if !strings.Contains(knowledge, "lesson:") ||
		!strings.Contains(knowledge, "before:") || !strings.Contains(knowledge, "after:") ||
		!strings.Contains(knowledge, "NOPE") {
		t.Fatalf("the lesson must carry the fix's bytes, knowledge:\n%s", knowledge)
	}
	if !strings.Contains(knowledge, "- [n1] (tags: copy, the) the copy must carry the input bytes") {
		t.Fatalf("the note must stand once with its tags inline, knowledge:\n%s", knowledge)
	}
	if !strings.Contains(knowledge, "## index") || !strings.Contains(knowledge, "copy: n1") || !strings.Contains(knowledge, "the: n1") {
		t.Fatalf("the index must gather the note's tags, knowledge:\n%s", knowledge)
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

// TestCanonSelfValidation — the canon's own honesty: its example
// passes the machine's validation, runs green against the real
// machine in a fresh fixture and closes the verdict whole — the
// registry, the marker, the artifact and the covers all exercising
// the laws the contract prints. A canon that lies fails the build.
func TestCanonSelfValidation(t *testing.T) {
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
	writeZone(t, root, zone.WishSlot, "a counter\ncount the lines of count.txt\nprint the count\n")
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, src.Example)
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "green: 1") {
		t.Fatalf("the canon example must run green, report:\n%s", report)
	}
	// The example's own corpus is the completion act: the whole
	// chain — rows green, the wish accounted, covers standing, no
	// uncovered source — must close on the canon's own words.
	writeZone(t, root, zone.InboxSlot, zoneFile(t, root, zone.CorpusFile))
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
	writeZone(t, root, zone.InboxSlot, zoneFile(t, root, zone.CorpusFile))
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
// scope file inside the run leaves the tree exactly as the verdict
// certified: the render's mark, measured after the execution, says
// standing, the history holds no red — the two truths agree.
func TestSelfHealingRowLeavesNoMovedMark(t *testing.T) {
	root := newRoot(t)
	writeZone(t, root, zone.InboxSlot, `requirements:
    - id: R1
      brief: [W1]
artifacts: [ex.txt]
behaviors:
    - id: tool/write
      name: The writer
      prose: Writes its own proof file.
      scope:
        - ex.txt
      covers: [R1]
      rows:
        - run: printf 'ok' > ex.txt
          expect:
            files:
                ex.txt: "contains: ok"
`)
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, zoneFile(t, root, zone.CorpusFile))
	runIn(t, root)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "verdict: #1") {
		t.Fatalf("the verdict must be recorded, state:\n%s", state)
	}
	// The break and the heal inside one run: the scope file breaks,
	// the row rewrites it and stays green.
	if err := os.WriteFile(filepath.Join(root, "ex.txt"), []byte("no"), 0o644); err != nil {
		t.Fatal(err)
	}
	runIn(t, root)
	state = zoneFile(t, root, zone.StateFile)
	if strings.Contains(state, "scopes moved") {
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
	writeZone(t, root, zone.InboxSlot, zoneFile(t, root, zone.CorpusFile))
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
	writeZone(t, root, zone.InboxSlot, zoneFile(t, root, zone.CorpusFile))
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
	writeZone(t, root, zone.InboxSlot, zoneFile(t, root, zone.CorpusFile))
	runIn(t, root)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "verdict history:\n- run 1: verdict #1 on v1 recorded") {
		t.Fatalf("the history must open with the first verdict:\n%s", state)
	}
	// The drift: a scope file breaks the row, the run catches it.
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
	writeZone(t, root, zone.InboxSlot, zoneFile(t, root, zone.CorpusFile))
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
	writeZone(t, root, zone.InboxSlot, `behaviors:
    - id: tool/second
      name: The second
      prose: A plain later work.
      scope:
        - second.txt
      rows:
        - run: ["sh", "-c", "printf second > second.txt"]
          expect:
            files:
                second.txt: second
`)
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, zoneFile(t, root, zone.CorpusFile))
	runIn(t, root)
	if report := zoneFile(t, root, zone.ReportFile); strings.Contains(report, "drift healed") {
		t.Fatalf("a close with no drift of its own must not retell the old heal:\n%s", report)
	}
	if state := zoneFile(t, root, zone.StateFile); !strings.Contains(state, "- run 1: verdict #1 on v1 recorded") {
		t.Fatalf("the history keeps its first fact:\n%s", state)
	}
}

func TestHistoryNamesAnUnreadableJournal(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	writeZone(t, root, zone.InboxSlot, zoneFile(t, root, zone.CorpusFile))
	code, term := runCode(t, root, map[string]bool{"read:epoch-!after-write": true})
	if code != 0 {
		t.Fatalf("an unreadable journal at render time must not kill the run: %s", term)
	}
	if !strings.Contains(zoneFile(t, root, zone.StateFile), "verdict history: (journal unreadable:") {
		t.Fatalf("the history must say so by name:\n%s", zoneFile(t, root, zone.StateFile))
	}
}
