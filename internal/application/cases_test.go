// The user cases' battery: every way real work meets the machine —
// a virgin project, an existing tree, new work on a live instance,
// parallel workers on disjoint behaviors, a replaced behavior, a
// regression edited outside the machine, the audit by reading alone,
// the recovery from a cut journal write — each case asserts its own
// invariants over the zone's files and the machine's one-line
// answers.
package application

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neurophant/punchtape/internal/domain/zone"
	"github.com/neurophant/punchtape/internal/infrastructure/fsys"
	"github.com/neurophant/punchtape/internal/infrastructure/process"
	"github.com/neurophant/punchtape/internal/ports"
)

// newBornRoot — a root with a born instance and a working behavior
// (cat src/in.txt > out.txt, scope src/**), its wish accounted by
// one requirement the behavior covers and its output declared.
func newBornRoot(t *testing.T) string {
	t.Helper()
	root := newRoot(t)
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
	return root
}

func TestExistingTreeIsNotTouchedByBirth(t *testing.T) {
	mountPorts()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "main.txt"), []byte("the project's own"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.WishSlot, "wrap the standing tree")
	before, _ := os.ReadFile(filepath.Join(root, "main.txt"))
	runIn(t, root)
	after, _ := os.ReadFile(filepath.Join(root, "main.txt"))
	if string(before) != string(after) {
		t.Fatalf("the birth must not touch the tree's own files")
	}
	// The standing tree's file is honest material for a scope.
	writeZone(t, root, zone.InboxSlot, `behaviors:
    - id: tool/keep
      name: The keeper
      prose: Keeps.
      scope:
        - main.txt
      rows:
        - run: cat main.txt > kept.txt
          expect:
            files:
                kept.txt: "contains: the project's own"
`)
	runIn(t, root)
	if !strings.Contains(zoneFile(t, root, zone.ReportFile), "green: 1") {
		t.Fatalf("the standing tree must serve as scope material")
	}
}

func TestParallelWorkersOnDisjointBehaviors(t *testing.T) {
	root := newBornRoot(t)
	// Two submissions touching disjoint ids: both stand, order is
	// irrelevant, the corpus carries both.
	writeZone(t, root, zone.InboxSlot, `behaviors:
    - id: tool/alpha
      name: Alpha
      prose: A.
      scope:
        - src/**
      rows:
        - run: cat src/in.txt > alpha.txt
          expect:
            files:
                alpha.txt: "contains: ok"
`)
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, `behaviors:
    - id: tool/beta
      name: Beta
      prose: B.
      scope:
        - src/**
      rows:
        - run: cat src/in.txt > beta.txt
          expect:
            files:
                beta.txt: "contains: ok"
`)
	runIn(t, root)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "tool/alpha") || !strings.Contains(state, "tool/beta") {
		t.Fatalf("both workers' behaviors must stand, state:\n%s", state)
	}
	if !strings.Contains(state, "corpus: v3") {
		t.Fatalf("each submission bumps the corpus version, state:\n%s", state)
	}
}

func TestRowMergeById(t *testing.T) {
	root := newBornRoot(t)
	// A new run alone: the standing row's proof survives the merge —
	// only the appended row executes this run.
	writeZone(t, root, zone.InboxSlot, `behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies.
      scope:
        - src/**
      rows:
        - run: cat src/in.txt > out2.txt
          expect:
            files:
                out2.txt: "contains: ok"
`)
	runIn(t, root)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "tool/copy: 2 rows — 2 green") {
		t.Fatalf("the merged behavior must stand with both rows, state:\n%s", state)
	}
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "green: 1") {
		t.Fatalf("only the appended row may execute — the unlisted row stands from memory, report:\n%s", report)
	}
	// The same run alone: the whole row replaces in place, the other
	// row's proof survives again.
	writeZone(t, root, zone.InboxSlot, `behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies.
      scope:
        - src/**
      rows:
        - run: cat src/in.txt > out.txt
          expect:
            rc: 1
`)
	runIn(t, root)
	if report = zoneFile(t, root, zone.ReportFile); !strings.Contains(report, "exit: got 0, want 1") || !strings.Contains(report, "tool/copy/0") {
		t.Fatalf("the replaced row must re-execute with its whole new expectation, report:\n%s", report)
	}
	// A duplicate run in one behavior: refused by its named place,
	// the inbox stays, nothing merges.
	writeZone(t, root, zone.InboxSlot, `behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies.
      scope:
        - src/**
      rows:
        - run: cat src/in.txt > out.txt
        - run: cat src/in.txt > out.txt
`)
	runIn(t, root)
	report = zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "behaviors[0].rows[1].run") || !strings.Contains(report, "law:document") {
		t.Fatalf("the duplicate run must be refused with behavior, key and place, report:\n%s", report)
	}
	if _, err := os.Stat(filepath.Join(root, zone.Dirname, zone.InboxSlot)); err != nil {
		t.Fatalf("a refused submission keeps the inbox")
	}
	// The removal of a standing id retires the behavior and its spec
	// page leaves with it.
	writeZone(t, root, zone.InboxSlot, "remove: [tool/copy]\n")
	runIn(t, root)
	if strings.Contains(zoneFile(t, root, zone.StateFile), "tool/copy") {
		t.Fatalf("the retired behavior must leave the state")
	}
	if _, err := os.Stat(filepath.Join(root, zone.Dirname, zone.SpecDir, "tool", "copy.md")); !os.IsNotExist(err) {
		t.Fatalf("the retired behavior's page must leave the spec tree")
	}
}

func TestRegressionEditedOutsideTheMachine(t *testing.T) {
	root := newBornRoot(t)
	// Someone edits the tree directly: the next run re-executes what
	// the edit touched and shows the honest red.
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("NOPE"), 0o644); err != nil {
		t.Fatal(err)
	}
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, `want "contains: ok"`) {
		t.Fatalf("the regression must show its red fact, report:\n%s", report)
	}
	// The fix lands in the tree: the row goes green again, and the
	// lesson of the pass is recorded with both byte spellings.
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	runIn(t, root)
	knowledge := zoneFile(t, root, zone.KnowledgeFile)
	if !strings.Contains(knowledge, "lesson:") {
		t.Fatalf("the fixed red must leave its lesson, knowledge:\n%s", knowledge)
	}
}

func TestAuditByReadingAlone(t *testing.T) {
	root := newBornRoot(t)
	// The completion act closes the verdict, and the audit is pure
	// reading: the spec tree carries the behavior with its standing.
	writeZone(t, root, zone.InboxSlot, zoneFile(t, root, zone.CorpusFile))
	runIn(t, root)
	index := zoneFile(t, root, filepath.Join(zone.SpecDir, "index.md"))
	if !strings.Contains(index, "tool/copy — The copy — green") {
		t.Fatalf("the index must show the standing, index:\n%s", index)
	}
	page := zoneFile(t, root, filepath.Join(zone.SpecDir, "tool", "copy.md"))
	if !strings.Contains(page, "Copies.") || !strings.Contains(page, "command: cat src/in.txt > out.txt") {
		t.Fatalf("the page must carry the prose and the row, page:\n%s", page)
	}
	// A run that changes nothing executes nothing: the audit costs
	// zero and the standing files do not move.
	stateBefore := zoneFile(t, root, zone.StateFile)
	runIn(t, root)
	if zoneFile(t, root, zone.StateFile) != stateBefore {
		t.Fatalf("an idle run must not rewrite the standing")
	}
}

func TestBuildFailureStopsTheRows(t *testing.T) {
	mountPorts()
	root := t.TempDir()
	writeZone(t, root, zone.WishSlot, "a built tool")
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, `build: "exit 7"
behaviors:
    - id: tool/built
      name: The built
      prose: Needs the build.
      scope:
        - src/**
      rows:
        - run: "true"
          expect:
            rc: 0
`)
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "build: RED — exit 7") {
		t.Fatalf("the build's red must be named, report:\n%s", report)
	}
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "tool/built: 1 rows — 0 green, 1 stale") &&
		!strings.Contains(state, "1 stale") {
		t.Fatalf("the rows must not execute behind a failed build, state:\n%s", state)
	}
	// The build heals: the rows execute on the next run.
	writeZone(t, root, zone.InboxSlot, `build: "true"
behaviors:
    - id: tool/built
      name: The built
      prose: Needs the build.
      scope:
        - src/**
      rows:
        - run: "true"
          expect:
            rc: 0
`)
	runIn(t, root)
	if !strings.Contains(zoneFile(t, root, zone.ReportFile), "green: 1") {
		t.Fatalf("the healed build must let the rows run")
	}
}

func TestCutJournalWriteRecoversByQuarantine(t *testing.T) {
	root := newBornRoot(t)
	// A torn tail: bytes cut into the journal's live epoch after its
	// last complete record. The next run loads every complete record
	// and quarantines the torn bytes to the sidecar.
	journalDir := filepath.Join(root, zone.Dirname, zone.JournalDir)
	entries, err := os.ReadDir(journalDir)
	if err != nil {
		t.Fatalf("the journal must exist: %v", err)
	}
	live := ""
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "epoch-") && !e.IsDir() {
			live = filepath.Join(journalDir, e.Name())
		}
	}
	if live == "" {
		t.Fatal("no epoch file to cut")
	}
	f, err := os.OpenFile(live, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString("---\nkind: run\nn: 99\ncut mid-wri"); err != nil {
		t.Fatal(err)
	}
	f.Close()
	runIn(t, root) // must not fail: the torn tail quarantines
	if _, err := os.Stat(live + ".torn"); err != nil {
		t.Fatalf("the torn bytes must quarantine to the sidecar: %v", err)
	}
	// The standing facts survived whole.
	if !strings.Contains(zoneFile(t, root, zone.StateFile), "tool/copy: 1 rows — 1 green") {
		t.Fatalf("the complete records must survive the cut")
	}
	// A hand-edited middle record is damage, not a torn write: the
	// machine refuses by name instead of guessing.
	data, _ := os.ReadFile(live)
	if i := strings.Index(string(data), "n: 1"); i >= 0 {
		forged := append([]byte(nil), data...)
		forged[i] = 'x'
		_ = os.WriteFile(live, forged, 0o644)
	}
	var term strings.Builder
	code := Run(root, &term)
	if code != 1 || !strings.Contains(term.String(), "punchtape:") {
		t.Fatalf("journal damage must refuse the run, code %d, term %q", code, term.String())
	}
}

// mountPorts — the production ports over the real OS.
func mountPorts() {
	ports.FSBind = fsys.OS{}
	ports.ExecBind = process.OS{}
}

func TestStackMigrationKeepsTheBehaviors(t *testing.T) {
	root := newBornRoot(t)
	// The same behavior id over a different implementation: a new
	// command is a new run — it appends, the old row stands (both
	// carry the same observable contract, both stay green).
	writeZone(t, root, zone.InboxSlot, `build: "true"
behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies, by cp now.
      scope:
        - src/**
      covers: [R1]
      rows:
        - run: cp src/in.txt out.txt
          expect:
            files:
                out.txt: "contains: ok"
`)
	runIn(t, root)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "tool/copy: 2 rows — 2 green") {
		t.Fatalf("the new command appends its row, the old one stands, state:\n%s", state)
	}
	// A whole swap retires the id and re-adds it in one document:
	// the behavior lands on exactly the new implementation's rows.
	writeZone(t, root, zone.InboxSlot, `remove: [tool/copy]
build: "true"
behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies, by cp now.
      scope:
        - src/**
      covers: [R1]
      rows:
        - run: cp src/in.txt out.txt
          expect:
            files:
                out.txt: "contains: ok"
`)
	runIn(t, root)
	state = zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "tool/copy: 1 rows — 1 green") {
		t.Fatalf("the retired-and-re-added behavior must carry only the new rows, state:\n%s", state)
	}
	writeZone(t, root, zone.InboxSlot, zoneFile(t, root, zone.CorpusFile))
	runIn(t, root)
	if !strings.Contains(zoneFile(t, root, zone.StateFile), "verdict: #1") {
		t.Fatalf("the migrated corpus must reach its verdict")
	}
}

func TestARowlessBehaviorIsRefused(t *testing.T) {
	mountPorts()
	root := newRoot(t)
	writeZone(t, root, zone.WishSlot, "a small tool, honestly checked")
	runIn(t, root)
	// A behavior without rows proves nothing: the submission is
	// refused by its named place and stays in the inbox.
	writeZone(t, root, zone.InboxSlot, `behaviors:
    - id: tool/late
      name: Late rows
      prose: A head alone carries no proof.
      scope:
        - src/**
`)
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "behaviors[0].rows") || !strings.Contains(report, "at least one row") {
		t.Fatalf("a rowless behavior must be refused by its named place:\n%s", report)
	}
	if _, err := os.Stat(filepath.Join(root, zone.Dirname, zone.InboxSlot)); err != nil {
		t.Fatalf("the refused submission keeps the inbox")
	}
	// The head and its rows arrive together: the behavior stands.
	if err := os.WriteFile(filepath.Join(root, "src.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, `behaviors:
    - id: tool/late
      name: Late rows
      prose: A head alone carries no proof.
      scope:
        - src/**
      rows:
        - run: cat src.txt > out.txt
          expect:
            files:
                out.txt: "contains: ok"
`)
	runIn(t, root)
	if state := zoneFile(t, root, zone.StateFile); !strings.Contains(state, "tool/late: 1 rows — 1 green") {
		t.Fatalf("the behavior with its rows must stand, state:\n%s", state)
	}
}

func TestFoldReplaysTheLiveStanding(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	// The journal's fold must replay the live standing byte for byte
	// through the merge laws: a fresh id, and a remove-and-re-add of
	// one id in one document.
	if err := os.WriteFile(filepath.Join(root, "extra.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, `behaviors:
    - id: tool/extra
      name: Extra
      prose: A second behavior.
      scope:
        - extra.txt
      rows:
        - run: "true"
`)
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, `remove: [tool/copy]
behaviors:
    - id: tool/copy
      name: The copy
      prose: Starts fresh.
      scope:
        - src/**
      rows:
        - run: cat src/in.txt > out.txt
          expect:
            files:
                out.txt: "contains: ok"
`)
	runIn(t, root)
	corpus := zoneFile(t, root, zone.CorpusFile)
	state := zoneFile(t, root, zone.StateFile)
	if strings.Contains(state, "2 rows") || !strings.Contains(state, "tool/copy: 1 rows") {
		t.Fatalf("the remove-and-re-add must start fresh, state:\n%s", state)
	}
	// The idle run reopens the zone: the fold rebuilds the standing
	// from the journal alone — the views must not move a byte.
	runIn(t, root)
	if zoneFile(t, root, zone.CorpusFile) != corpus {
		t.Fatalf("the fold must replay the live corpus byte for byte:\n%s", zoneFile(t, root, zone.CorpusFile))
	}
	if zoneFile(t, root, zone.StateFile) != state {
		t.Fatalf("the fold must replay the live state byte for byte:\n%s", zoneFile(t, root, zone.StateFile))
	}
}

func TestBigEditRechecksOnlyTheTouched(t *testing.T) {
	root := newBornRoot(t)
	for _, dir := range []string{"src-a", "src-b"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	os.WriteFile(filepath.Join(root, "src-a", "in.txt"), []byte("alpha"), 0o644)
	os.WriteFile(filepath.Join(root, "src-b", "in.txt"), []byte("beta"), 0o644)
	writeZone(t, root, zone.InboxSlot, `behaviors:
    - id: tool/one
      name: One
      prose: Stands on src-a.
      scope:
        - src-a/**
      rows:
        - run: cat src-a/in.txt > one.txt
          expect:
            files:
                one.txt: "contains: alpha"
    - id: tool/two
      name: Two
      prose: Stands on src-b.
      scope:
        - src-b/**
      rows:
        - run: cat src-b/in.txt > two.txt
          expect:
            files:
                two.txt: "contains: beta"
`)
	runIn(t, root)
	// A big edit touches one scope: exactly that behavior's rows
	// re-execute, the other stands from memory.
	os.WriteFile(filepath.Join(root, "src-a", "in.txt"), []byte("alpha again"), 0o644)
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "green: 1") || strings.Contains(report, "tool/two") {
		t.Fatalf("exactly the touched behavior's row must re-execute, report:\n%s", report)
	}
	one, err := os.ReadFile(filepath.Join(root, "one.txt"))
	if err != nil || !strings.Contains(string(one), "alpha again") {
		t.Fatalf("the touched row must actually re-run: %v %q", err, one)
	}
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "tool/one: 1 rows — 1 green") || !strings.Contains(state, "tool/two: 1 rows — 1 green") {
		t.Fatalf("both behaviors must stand green, state:\n%s", state)
	}
}
