// The views' tails: the state's lesson counter, the build's timeout
// line, the refless note line, the references-only knowledge index,
// the completion act's whole-rewrite of the report.
package application

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/neurophant/punchtape/internal/domain/fprint"
	"github.com/neurophant/punchtape/internal/domain/zone"
)

func TestStatePointsAtKnowledgeForAReasonedRed(t *testing.T) {
	mountPorts()
	root := newRoot(t)
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Two behaviors fail the same way over disjoint claims; one
	// heals — its lesson stands; the other stays red with the same
	// reason: the state points at the knowledge.
	if err := os.WriteFile(filepath.Join(root, "src", "in2.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, fmt.Sprintf(`
work: two rows failing the same way
behaviors:
    - id: tool/one
      name: One
      prose: Fails then heals.
      claims:
        - src/in.txt:1-1
      rows:
        - run: sh -c 'printf "%%s\n" %s; printf x > one.txt'
          expect:
            files:
                one.txt: y
    - id: tool/two
      name: Two
      prose: Stays failing.
      claims:
        - src/in2.txt:1-1
      rows:
        - run: sh -c 'printf "%%s\n" %s; printf x > two.txt'
          expect:
            files:
                two.txt: y
`, claimsTokenOf(t, root, "src/in.txt:1-1"), claimsTokenOf(t, root, "src/in2.txt:1-1")))
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, fmt.Sprintf(`
work: the heal of one
mandate: [{line: D1, target: tool/one}]
behaviors:
    - id: tool/one
      name: One
      prose: Fails then heals.
      claims:
        - src/in.txt:1-1
      rows:
        - run: sh -c 'printf "%%s\n" %s; printf x > one.txt'
          expect:
            files:
                one.txt: x
`, claimsTokenOf(t, root, "src/in.txt:1-1")))
	runIn(t, root)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "knowledge: 0 notes, 1 lessons") {
		t.Fatalf("the state must count the standing lesson:\n%s", state)
	}
}

func TestBuildTimeoutIsARedLine(t *testing.T) {
	mountPorts()
	root := newRoot(t)
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.EnvSlot, "PUNCHTAPE_ROW_TIMEOUT=1\n")
	writeZone(t, root, zone.InboxSlot, `
work: the slow build under a one-second pin
build: sleep 5
behaviors:
    - id: tool/slow
      name: Slow
      prose: Slow.
      claims:
        - src/in.txt:1-1
      rows:
        - run: cat src/in.txt > out.txt
          expect:
            files:
                out.txt: ok
`)
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "build: RED — timed out after 1s") {
		t.Fatalf("the timed-out build must say so:\n%s", report)
	}
}

func TestReflessNoteStandsOnce(t *testing.T) {
	mountPorts()
	root := newRoot(t)
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, `
work: a note that references nothing
notes:
    - kind: work
      text: zzquux zzfrob
behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies.
      claims:
        - src/in.txt:1-1
      rows:
        - run: cat src/in.txt > out.txt
          expect:
            files:
                out.txt: ok
`)
	runIn(t, root)
	knowledge := zoneFile(t, root, zone.KnowledgeFile)
	if !strings.Contains(knowledge, "- [n1] (work) zzquux zzfrob") {
		t.Fatalf("a note without references stands once, kind only:\n%s", knowledge)
	}
	if strings.Contains(knowledge, "## index") {
		t.Fatalf("a note without references carries no index:\n%s", knowledge)
	}
}

// TestKnowledgeHasNoIndex — the knowledge page carries the notes
// and the lessons only: a note stands once with its kind and its
// references inline, and no reverse index of references is rebuilt —
// a showcase of a showcase is not a face.
func TestKnowledgeHasNoIndex(t *testing.T) {
	mountPorts()
	root := t.TempDir()
	writeZone(t, root, zone.DemandsSlot, "first words\nsecond words\n")
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, `
work: two noted references
notes:
    - kind: decision
      text: the first decision cites two references
      refs: [D1, D2]
    - kind: work
      text: the second note cites one
      refs: [D2]
`)
	runIn(t, root)
	knowledge := zoneFile(t, root, zone.KnowledgeFile)
	for _, want := range []string{
		"- [n1] (decision; refs: D1, D2) the first decision cites two references",
		"- [n2] (work; refs: D2) the second note cites one",
	} {
		if !strings.Contains(knowledge, want) {
			t.Fatalf("missing %q in:\n%s", want, knowledge)
		}
	}
	if strings.Contains(knowledge, "## index") || strings.Contains(knowledge, "D1: n1") || strings.Contains(knowledge, "D2: n1, n2") {
		t.Fatalf("the reverse index is gone, knowledge:\n%s", knowledge)
	}
	if got := strings.Count(knowledge, "the first decision cites two references"); got != 1 {
		t.Fatalf("the note's text must stand exactly once, found %d", got)
	}
}

// TestDuplicateClaimRefused — one range, one behavior: a second
// holder of a claim inside one submission is refused with the
// holder named; a stranger meeting a standing holder's claim is
// refused the same way; two claims of one behavior stand.
func TestDuplicateClaimRefused(t *testing.T) {
	mountPorts()
	root := newRoot(t)
	if err := os.WriteFile(filepath.Join(root, "shared.txt"), []byte("both\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, `
work: two holders of one range
behaviors:
    - id: tool/first
      name: First
      prose: Holds the range first.
      claims:
        - shared.txt:1-1
      rows:
        - run: cat shared.txt > first.txt
          expect:
            files:
                first.txt: both
    - id: tool/second
      name: Second
      prose: Wants the same range.
      claims:
        - shared.txt:1-1
      rows:
        - run: head shared.txt > second.txt
          expect:
            files:
                second.txt: both
`)
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "behaviors[1].claims[0]: the claim shared.txt:1-1 is already held by tool/first — one range, one behavior (law:claims)") {
		t.Fatalf("the second holder must be refused with the holder named:\n%s", report)
	}
	// The standing holder meets a stranger with the same claim.
	writeZone(t, root, zone.InboxSlot, `
work: the honest first account
intents:
    - {lines: [D1], outcome: work, prose: the tool and its check of the shared file}
requirements:
    - id: R1
      lines: [D1]
      prose: the tool and its check
uncovered-lines:
    - {range: "AGENTS.md:1-9", reason: the machine's own hook, refs: [D1]}
    - {range: "first.txt:1-1", reason: the row's own output, refs: [D1]}
behaviors:
    - id: tool/first
      name: First
      prose: Holds the range.
      claims:
        - shared.txt:1-1
      covers: [R1]
      rows:
        - run: cat shared.txt > first.txt
          expect:
            files:
                first.txt: both
`)
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, `
work: a stranger wants the standing range
behaviors:
    - id: tool/second
      name: Second
      prose: Wants the standing range.
      claims:
        - shared.txt:1-1
      rows:
        - run: head shared.txt > second.txt
          expect:
            files:
                second.txt: both
`)
	runIn(t, root)
	report = zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "behaviors[0].claims[0]: the claim shared.txt:1-1 is already held by tool/first — one range, one behavior (law:claims)") {
		t.Fatalf("the standing holder must refuse the stranger by name:\n%s", report)
	}
}

// TestNoAnomalyLineInState — the meaning-check theater is gone: the
// state page carries no anomaly count, whatever stands.
func TestNoAnomalyLineInState(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	runIn(t, root)
	if state := zoneFile(t, root, zone.StateFile); strings.Contains(state, "anomal") || strings.Contains(state, "meaning check") {
		t.Fatalf("the state must carry no anomaly theater:\n%s", state)
	}
}

// TestTraceIsShowcase — the trace carries the stream's own texts,
// the outcomes with their retellings and the rows' standings; it
// judges nothing — no anomaly line, no verdict word.
func TestTraceIsShowcase(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	runIn(t, root)
	trace := zoneFile(t, root, zone.TraceFile)
	for _, want := range []string{
		"D1 «a checked tool»",
		"├─ intent: work — «the tool and its check, told again»",
		"│  └─ R1 (span D1)",
		"│     └─ tool/copy (covers R1)",
		"│        row 0 PROOF: ",
	} {
		if !strings.Contains(trace, want) {
			t.Fatalf("the trace must carry %q:\n%s", want, trace)
		}
	}
	for _, gone := range []string{"anomal", "meaning check", "blind claim", "duplicate claim"} {
		if strings.Contains(trace, gone) {
			t.Fatalf("the trace judges nothing — %q must be gone:\n%s", gone, trace)
		}
	}
}

// TestHistoryCappedAtTen — the state's verdict history shows the
// recent ten events and names the elided rest by the journal's own
// address.
func TestHistoryCappedAtTen(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	// Twelve verdicts: each act over the standing green corpus
	// records one history event.
	for i := 0; i < 12; i++ {
		writeZone(t, root, zone.InboxSlot, "done: true\n")
		runIn(t, root)
	}
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "verdict history:\n- … and 2 more — see .punchtape/journal") {
		t.Fatalf("the elided events must be named by the journal's address:\n%s", state)
	}
	if strings.Contains(state, "verdict #1 on v1 recorded") || strings.Contains(state, "verdict #2 on v1 recorded") {
		t.Fatalf("only the recent ten stand on the page:\n%s", state)
	}
	if !strings.Contains(state, "verdict #3 on v1 recorded") || !strings.Contains(state, "verdict #12 on v1 recorded") {
		t.Fatalf("the recent ten must stand whole:\n%s", state)
	}
}

func TestCompletionRewritesTheReportWhole(t *testing.T) {
	mountPorts()
	root := newRoot(t)
	// A stale truncated report and an old refusal note must not
	// outlive the completion act: the report is the run's whole
	// picture, rewritten.
	if err := os.WriteFile(filepath.Join(root, zone.Dirname, zone.ReportFile), []byte("# punchtape report\n\nrun 0 — truncated\nverdict: not yet — old noise\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if strings.Contains(report, "truncated") || strings.Contains(report, "old noise") {
		t.Fatalf("the completion act must rewrite the report whole:\n%s", report)
	}
	if !strings.Contains(report, "verdict: refused — the corpus is empty") {
		t.Fatalf("the refusal stands as the run's own fact:\n%s", report)
	}
}

func TestAStandingVerdictLeavesNoStaleRefusalBehind(t *testing.T) {
	root := streamFixture(t)
	// The thin corpus: the completion act refuses and names the
	// holes — then the registry closes and the verdict stands; the
	// report of that run must carry the recording, not the refusal.
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if !strings.Contains(zoneFile(t, root, zone.ReportFile), "verdict: not yet") {
		t.Fatalf("the refusal must stand in its own run's report")
	}
	writeZone(t, root, zone.InboxSlot, fmt.Sprintf(`
work: the whole account closes
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
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "verdict: #1") || strings.Contains(report, "not yet") {
		t.Fatalf("the recorded verdict must be the run's whole report:\n%s", report)
	}
}

// TestStateGolden050 — the state page's whole 0.5.0 shape, byte for
// byte, every section in its one order. The proof and signature
// digests ride the machine's own environment; the golden masks them
// — their laws are pinned by their own tests.
func TestStateGolden050(t *testing.T) {
	mountPorts()
	root := t.TempDir()
	writeZone(t, root, zone.DemandsSlot, "a checked tool\nwhich color\n")
	runIn(t, root)
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, fmt.Sprintf(`
work: the whole standing in one document
intents:
    - {lines: [D1], outcome: work, prose: "the tool and its check, told again"}
    - {lines: [D2], outcome: mark, prose: the human's answer orders nothing}
requirements:
    - id: R1
      lines: [D1]
      prose: the tool and its check
not-a-requirement:
    - lines: [D2]
      reason: the human's answer, not new work
uncovered-lines:
    - {range: "AGENTS.md:1-9", reason: the machine's own hook, refs: [D1]}
    - {range: "out.txt:1-1", reason: the row's own output, refs: [D1]}
open:
    - {demand: D2, question: which color, state: deferred, reason: the human said later}
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
	golden := `# punchtape state — run 1

demands: 2/2 lines; requirements 1 standing, 1 covered; markers: 1
queue: empty
last no-new-demands: run 1
corpus: v1, 1 behaviors
submissions: 1 accepted, 0 refused, 1 completed
machine at birth: ` + Version + `
verdict: #1 on v1, run 1, signature DIGEST — standing
verdict history:
- run 1: verdict #1 on v1 recorded — DIGEST
deferred orders:
- D2 — which color — the human said later
- tool/copy: 1 rows — 1 green (claims 1 files / 1 lines; proof DIGEST)
excluded lines: 2 ranges (90 % of tree)

knowledge: 0 notes, 0 lessons

next: verdict #1 standing — show .punchtape/state.md to the human; new work goes to .punchtape/inbox.yaml as a document
`
	state := zoneFile(t, root, zone.StateFile)
	if masked := maskDigests(state); masked != golden {
		t.Fatalf("the state page must stand in its golden shape:\ngot:\n%s\nwant:\n%s", masked, golden)
	}
}

// maskDigests — the environment-borne 12-hex digests as one word.
func maskDigests(s string) string {
	return regexp.MustCompile(`[0-9a-f]{12}`).ReplaceAllString(s, "DIGEST")
}

// TestSpecPageSections — the spec page carries its sections in the
// one order: the head, the run, the story, the demands it answers,
// the claims it stands on, the rows that prove it, the commands that
// re-run them.
func TestSpecPageSections(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	page := zoneFile(t, root, filepath.Join(zone.SpecDir, "tool", "copy.md"))
	at := -1
	for _, want := range []string{
		"# tool/copy — The copy\n",
		"run: ",
		"## Story\n",
		"## Demands\n",
		"- R1 (lines D1): the tool and its check\n",
		"## Claims\n",
		"- src/in.txt:1-1 — ",
		"\ntoken: ",
		"## Rows\n",
		"## row 0 — PROOF\n",
		"## Verify\n",
		"- sh -c 'printf",
	} {
		i := strings.Index(page, want)
		if i < 0 || i < at {
			t.Fatalf("the page must carry %q in order:\n%s", want, page)
		}
		at = i
	}
}

// TestExcludedShareShown — the loud exclusions' share of the tree,
// whole percents over the already-counted coverage; the island label
// is a banned class and never prints; the empty share spells zero.
func TestExcludedShareShown(t *testing.T) {
	mountPorts()
	root := newRoot(t)
	if err := os.WriteFile(filepath.Join(root, "big.txt"), []byte(strings.Repeat("line\n", 40)), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, fmt.Sprintf(`
work: the loud share of the tree
intents:
    - {lines: [D1], outcome: work, prose: the tool and its check over a big file}
requirements:
    - id: R1
      lines: [D1]
      prose: the tool and its check
uncovered-lines:
    - {range: "AGENTS.md:1-9", reason: the machine's own hook, refs: [D1]}
    - {range: "big.txt:11-40", reason: generated tail, refs: [D1]}
    - {range: "head.txt:1-10", reason: the row's own output, refs: [D1]}
behaviors:
    - id: tool/big
      name: The big holder
      prose: Holds the head of the big file.
      claims:
        - big.txt:1-10
      covers: [R1]
      rows:
        - run: sh -c 'printf "%%s %%s\n" %s %s; head -n 10 big.txt > head.txt'
          expect:
            files:
                head.txt: %s
`, claimsTokenOf(t, root, "big.txt:1-10"), demandTokenOf("a checked tool"), strconv.Quote(strings.Repeat("line\n", 10))))
	runIn(t, root)
	// The tree: AGENTS.md 9 + big.txt 40 + head.txt 10 = 59 lines; the
	// exclusions hold all three's open parts: 49 of 59 = 83 %.
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "excluded lines: 3 ranges (83 % of tree)") {
		t.Fatalf("the exclusions' share must be shown whole, state:\n%s", state)
	}
	for _, banned := range []string{"island", "verdict on an island"} {
		if strings.Contains(state, banned) {
			t.Fatalf("the island label is a banned class, state:\n%s", state)
		}
	}
	// A zone with no exclusions spells the zero share.
	fresh := newRoot(t)
	writeZone(t, fresh, zone.InboxSlot, `
work: the untouched zone
`)
	runIn(t, fresh)
	if state = zoneFile(t, fresh, zone.StateFile); !strings.Contains(state, "excluded lines: 0 ranges (0 % of tree)") {
		t.Fatalf("the empty share must spell zero, state:\n%s", state)
	}
}

// TestStateHolesCapped — the state's hole list shows the first ten
// files and counts the rest, pointing at the ledger.
func TestStateHolesCapped(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	if err := os.MkdirAll(filepath.Join(root, "junk"), 0o755); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 11; i++ {
		if err := os.WriteFile(filepath.Join(root, "junk", fmt.Sprintf("j%02d.txt", i)), []byte("x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	runIn(t, root)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "- junk/j09.txt: 1 (1)") {
		t.Fatalf("the tenth file must still show, state:\n%s", state)
	}
	if strings.Contains(state, "- junk/j10.txt") {
		t.Fatalf("the eleventh file must not show, state:\n%s", state)
	}
	if !strings.Contains(state, "… and 1 more files — see .punchtape/holes.md") {
		t.Fatalf("the cap must count the rest, state:\n%s", state)
	}
}

// TestHolesFileComplete — the ledger carries every hole file with
// every maximal range, uncapped.
func TestHolesFileComplete(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	if err := os.MkdirAll(filepath.Join(root, "junk"), 0o755); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 11; i++ {
		if err := os.WriteFile(filepath.Join(root, "junk", fmt.Sprintf("j%02d.txt", i)), []byte("x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	runIn(t, root)
	holes := zoneFile(t, root, zone.HolesFile)
	for i := 0; i < 11; i++ {
		if !strings.Contains(holes, fmt.Sprintf("- junk/j%02d.txt: 1 (1)", i)) {
			t.Fatalf("the ledger must carry every hole file, holes:\n%s", holes)
		}
	}
	if strings.Contains(holes, "more files") {
		t.Fatalf("the ledger is never capped, holes:\n%s", holes)
	}
}

// TestTraceRendersIntentChain — the trace places the stream's every
// line with its quote, its intent's retelling beside it, the carrier,
// the behaviors, the claims tokens and the proof row's standing.
func TestTraceRendersIntentChain(t *testing.T) {
	root := newBornRoot(t)
	writeZone(t, root, zone.DemandsSlot, "a checked tool\nwhat does the check read?\nrun the echo\nthanks\nwhich unit?\nrun it again\nand the tool itself?\n")
	writeZone(t, root, zone.InboxSlot, `work: the stream's kinds in one account
intents:
    - {lines: [D2], outcome: answer, prose: the question of the input answered, text: the check reads the input's own bytes}
    - {lines: [D3], outcome: execute, prose: the command's own output answers, run: "printf done"}
    - {lines: [D4], outcome: mark, prose: the thanks order nothing}
    - {lines: [D5], outcome: ask, prose: the unit waits for the human}
    - {lines: [D6], outcome: execute, prose: the second command runs too, run: "printf again"}
    - {lines: [D7], outcome: answer, prose: the tool's own name answered, text: the tool is the copy itself}
not-a-requirement:
    - lines: [D4]
      reason: talk, not work
open:
    - {demand: D5, question: lines or bytes?, state: awaiting}
`)
	runIn(t, root)
	trace := zoneFile(t, root, zone.TraceFile)
	claims := claimsTokenOf(t, root, "src/in.txt:1-1")
	for _, want := range []string{
		"D1 «a checked tool»\n├─ intent: work — «the tool and its check, told again»",
		"│  └─ R1 (span D1)",
		"│     └─ tool/copy (covers R1)",
		fmt.Sprintf("│        claims: src/in.txt:1-1 — token %s", claims),
		"│        row 0 PROOF: ",
		"D2 «what does the check read?»\n├─ intent: answer — «the question of the input answered»",
		"│  └─ answer: the check reads the input's own bytes",
		"D3 «run the echo»\n├─ intent: execute — «the command's own output answers»",
		"│  └─ (execute pending)",
		"D4 «thanks»\n├─ intent: mark — «the thanks order nothing»",
		"│  └─ marker (reason: talk, not work)",
		"D5 «which unit?»\n├─ intent: ask — «the unit waits for the human»",
		"│  └─ order (awaiting): lines or bytes?",
		"D6 «run it again»\n├─ intent: execute — «the second command runs too»",
		"│  └─ execution (last): exit 0 — ",
		"D7 «and the tool itself?»\n├─ intent: answer — «the tool's own name answered»",
		"│  └─ answer: the tool is the copy itself",
	} {
		if !strings.Contains(trace, want) {
			t.Fatalf("the trace must carry %q:\n%s", want, trace)
		}
	}
	// The question closes over a later human line: the ask node spells
	// its closed shape.
	writeZone(t, root, zone.DemandsSlot, "a checked tool\nwhat does the check read?\nrun the echo\nthanks\nwhich unit?\nrun it again\nand the tool itself?\nthe line it is\n")
	writeZone(t, root, zone.InboxSlot, `work: the unit question closes over the answer
intents:
    - {lines: [D8], outcome: mark, prose: the human's own answer to the question}
not-a-requirement:
    - lines: [D8]
      reason: the human's answer to the standing question
notes:
    - kind: decision
      text: the unit is the line by the human's answer
      refs: [D8]
remove: [D5]
`)
	runIn(t, root)
	trace = zoneFile(t, root, zone.TraceFile)
	if !strings.Contains(trace, "D5 «which unit?»\n├─ intent: ask — «the unit waits for the human»\n│  └─ (ask pending)") {
		t.Fatalf("the closed ask must spell its shape:\n%s", trace)
	}
}

// TestExclusionsProposalMatchesHoles — the proposal drafts exactly
// the holes the ledger carries: the same ranges, byte for byte, both
// directions.
func TestExclusionsProposalMatchesHoles(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	if err := os.MkdirAll(filepath.Join(root, "junk"), 0o755); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		if err := os.WriteFile(filepath.Join(root, "junk", fmt.Sprintf("j%d.txt", i)), []byte("x\ny\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	runIn(t, root)
	proposal := zoneFile(t, root, zone.ExclusionsProposalFile)
	var proposed []string
	for _, line := range strings.Split(proposal, "\n") {
		if !strings.HasPrefix(line, "  - {range: ") {
			continue
		}
		spelling := strings.Trim(strings.TrimSuffix(strings.TrimPrefix(line, "  - {range: "), ", reason: '', refs: []}"), `"`)
		proposed = append(proposed, spelling)
	}
	if len(proposed) != 3 {
		t.Fatalf("the proposal must draft every hole range, got %d:\n%s", len(proposed), proposal)
	}
	for _, spelling := range proposed {
		rel, first, last, err := fprint.ParseLineRange(spelling)
		if err != nil {
			t.Fatalf("the proposal's spelling must parse: %q", spelling)
		}
		count := strconv.Itoa(last - first + 1)
		ranges := fmt.Sprintf("%d", first)
		if first != last {
			ranges = fmt.Sprintf("%d-%d", first, last)
		}
		want := fmt.Sprintf("- %s: %s (%s)", rel, count, ranges)
		if !strings.Contains(zoneFile(t, root, zone.HolesFile), want) {
			t.Fatalf("every proposed range must match a ledger line, want %q", want)
		}
	}
	// The other direction: once every hole is held — claimed or
	// loudly excluded — the draft leaves with them; a proposal that
	// outlives its arithmetic is a stale face.
	writeZone(t, root, zone.InboxSlot, `work: the junk holes go loud
uncovered-lines:
  - {range: "junk/j0.txt:1-2", reason: the test's own junk, refs: ["law:claims"]}
  - {range: "junk/j1.txt:1-2", reason: the test's own junk, refs: ["law:claims"]}
  - {range: "junk/j2.txt:1-2", reason: the test's own junk, refs: ["law:claims"]}
`)
	runIn(t, root)
	if _, err := os.Stat(filepath.Join(root, zone.Dirname, zone.ExclusionsProposalFile)); !os.IsNotExist(err) {
		t.Fatal("the proposal must leave when no hole stands")
	}
}

// TestStandingVerdictUnderUnreadableTree — a verdict standing over a
// tree the render cannot read: the state names the unreadable
// arithmetic, the verdict still spells its standing, the proposal's
// own walk fails the run loudly.
func TestStandingVerdictUnderUnreadableTree(t *testing.T) {
	root := newBornRoot(t)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if state := zoneFile(t, root, zone.StateFile); !strings.Contains(state, "verdict: #1") {
		t.Fatalf("the fixture must close its verdict, state:\n%s", state)
	}
	code, _ := runCode(t, root, map[string]bool{"walk:" + root: true})
	if code != 1 {
		t.Fatalf("the unreadable tree must fail the run at the proposal's walk, got %d", code)
	}
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "uncovered lines: (tree unreadable:") {
		t.Fatalf("the state must name the unreadable arithmetic, state:\n%s", state)
	}
	if !strings.Contains(state, " — standing\n") {
		t.Fatalf("the verdict's own standing does not hang on the walk, state:\n%s", state)
	}
}
