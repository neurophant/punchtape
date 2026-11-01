// The tree lock's and the executions' battery: a locked submission
// leaves the tree as it found it — a move under answer, execute or
// describe intents is a journaled fact that holds the verdict until
// the tree returns or the change becomes someone's declared work;
// the machine's own executions run under the row law, their facts
// journaled, their output whole on the report, their artifacts the
// lock's own to catch.
package application

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neurophant/punchtape/internal/domain/journal"
	"github.com/neurophant/punchtape/internal/domain/zone"
	"github.com/neurophant/punchtape/internal/infrastructure/store/cas"
)

// treeLockRecords — the zone's journaled tree-lock facts.
func treeLockRecords(t *testing.T, root string) []journal.TreeLock {
	t.Helper()
	records, err := journalRecords(t, root)
	if err != nil {
		t.Fatal(err)
	}
	var out []journal.TreeLock
	for _, rec := range records {
		if rec.Kind == journal.KindTreeLock && rec.TreeLock != nil {
			out = append(out, *rec.TreeLock)
		}
	}
	return out
}

// objectsOf — the zone's object store.
func objectsOf(t *testing.T, root string) *cas.Store {
	t.Helper()
	return cas.New(filepath.Join(root, zone.Dirname, zone.ObjectsDir))
}

func TestTreeLockRejectsChangesUnderDescribe(t *testing.T) {
	mountPorts()
	root := newRoot(t)
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	// The describe submission's proof writes its probe file inside the
	// run — the tree moves under a locked intent.
	writeZone(t, root, zone.InboxSlot, fmt.Sprintf(`
work: the describe account of the check
intents:
    - {lines: [D1], outcome: describe, prose: the tool told back without touching it}
requirements:
    - id: R1
      lines: [D1]
      prose: the tool and its check
uncovered-lines:
    - {range: "AGENTS.md:1-9", reason: the machine's own hook, refs: [D1]}
behaviors:
    - id: tool/probe
      name: The probe
      prose: Describes by writing its probe.
      claims:
        - src/in.txt:1-1
      covers: [R1]
      rows:
        - run: sh -c 'printf "%%s %%s\n" %s %s; cat src/in.txt > probe.txt'
          expect:
            files:
                probe.txt: ok
`, claimsTokenOf(t, root, "src/in.txt:1-1"), demandTokenOf("a checked tool")))
	runIn(t, root)
	if locks := treeLockRecords(t, root); len(locks) != 1 {
		t.Fatalf("the moved tree under a locked submission must journal its fact, got %d", len(locks))
	}
	// The act holds on the standing violation.
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "verdict: not yet — the tree changed under answer/execute/describe intents (law:demand)") {
		t.Fatalf("the tree lock must hold the verdict by its own words:\n%s", report)
	}
}

func TestTreeLockHealsOnReturn(t *testing.T) {
	mountPorts()
	root := newRoot(t)
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, fmt.Sprintf(`
work: the describe account of the check
intents:
    - {lines: [D1], outcome: describe, prose: the tool told back without touching it}
requirements:
    - id: R1
      lines: [D1]
      prose: the tool and its check
uncovered-lines:
    - {range: "AGENTS.md:1-9", reason: the machine's own hook, refs: [D1]}
behaviors:
    - id: tool/probe
      name: The probe
      prose: Describes by writing its probe.
      claims:
        - src/in.txt:1-1
      covers: [R1]
      rows:
        - run: sh -c 'printf "%%s %%s\n" %s %s; cat src/in.txt > probe.txt'
          expect:
            files:
                probe.txt: ok
`, claimsTokenOf(t, root, "src/in.txt:1-1"), demandTokenOf("a checked tool")))
	runIn(t, root)
	// The probe leaves: the tree returns to what the lock recorded —
	// the hold lifts, the fact stands in the journal forever.
	if err := os.Remove(filepath.Join(root, "probe.txt")); err != nil {
		t.Fatal(err)
	}
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if state := zoneFile(t, root, zone.StateFile); !strings.Contains(state, "verdict: #1") {
		t.Fatalf("the returned tree must let the verdict land, state:\n%s", state)
	}
	if locks := treeLockRecords(t, root); len(locks) != 1 {
		t.Fatalf("the healed lock keeps its journaled fact, got %d", len(locks))
	}
}

// TestFirstBuildArtifactsAreLegalWork — the first-declared build is
// a work attribute as much as the standing one: a submission of
// answer and describe intents whose build writes into the tree never
// arms the lock on its own build's artifacts — the verdict closes
// over them.
func TestFirstBuildArtifactsAreLegalWork(t *testing.T) {
	mountPorts()
	root := t.TempDir()
	writeZone(t, root, zone.DemandsSlot, "answer me\ndescribe the tool")
	runIn(t, root)
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	// The artifact stands as a one-line hole ahead of the submission;
	// the build overwrites it inside the run.
	if err := os.WriteFile(filepath.Join(root, "src", "built.tmp"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, `
work: the answer and the describe under a first build
build: "printf built > src/built.tmp"
intents:
    - {lines: [D1], outcome: answer, prose: the answer without touching the tree, text: the answer itself}
    - {lines: [D2], outcome: describe, prose: the tool told back without touching it}
requirements:
    - id: R1
      lines: [D2]
      prose: the tool and its check
uncovered-lines:
    - {range: "src/built.tmp:1-1", reason: the build's own artifact, refs: [D2]}
    - {range: "AGENTS.md:1-9", reason: the machine's own hook, refs: [D1]}
behaviors:
    - id: tool/probe
      name: The probe
      prose: Reads the input, pinned.
      claims:
        - src/in.txt:1-1
      covers: [R1]
      rows:
        - run: cat src/in.txt
          expect:
            out: "ok"
`)
	runIn(t, root)
	if locks := treeLockRecords(t, root); len(locks) != 0 {
		t.Fatalf("the first build's own artifacts must never arm the lock, got %d facts", len(locks))
	}
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if state := zoneFile(t, root, zone.StateFile); !strings.Contains(state, "verdict: #1") {
		t.Fatalf("the first build's artifacts are legal work, state:\n%s", state)
	}
}

func TestTreeLockCleanAnswerPasses(t *testing.T) {
	root := newBornRoot(t)
	// An answered line: the intent carries the answer, the tree stands
	// still — no lock fact ever lands, the verdict closes whole.
	writeZone(t, root, zone.DemandsSlot, "a checked tool\nwhat does the check read?\n")
	writeZone(t, root, zone.InboxSlot, `work: the question answered by its intent
intents:
    - {lines: [D2], outcome: answer, prose: the question of the check's input answered, text: the check reads the input's own bytes}
`)
	runIn(t, root)
	if locks := treeLockRecords(t, root); len(locks) != 0 {
		t.Fatalf("a clean answer must not journal a lock fact, got %d", len(locks))
	}
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if state := zoneFile(t, root, zone.StateFile); !strings.Contains(state, "verdict: #1") {
		t.Fatalf("the clean answer must let the verdict land, state:\n%s", state)
	}
}

func TestWorkIntentLiftsLock(t *testing.T) {
	mountPorts()
	root := newRoot(t)
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, fmt.Sprintf(`
work: the describe account of the check
intents:
    - {lines: [D1], outcome: describe, prose: the tool told back without touching it}
requirements:
    - id: R1
      lines: [D1]
      prose: the tool and its check
uncovered-lines:
    - {range: "AGENTS.md:1-9", reason: the machine's own hook, refs: [D1]}
behaviors:
    - id: tool/probe
      name: The probe
      prose: Describes by writing its probe.
      claims:
        - src/in.txt:1-1
      covers: [R1]
      rows:
        - run: sh -c 'printf "%%s %%s\n" %s %s; cat src/in.txt > probe.txt'
          expect:
            files:
                probe.txt: ok
`, claimsTokenOf(t, root, "src/in.txt:1-1"), demandTokenOf("a checked tool")))
	runIn(t, root)
	// The probe becomes someone's declared work: the work intent lifts
	// the hold, the moved tree stops being a violation.
	writeZone(t, root, zone.DemandsSlot, "a checked tool\nkeep the probe\n")
	writeZone(t, root, zone.InboxSlot, fmt.Sprintf(`work: the probe kept as declared work
intents:
    - {lines: [D2], outcome: work, prose: the probe kept as the work it is}
requirements:
    - id: R2
      lines: [D2]
      prose: the probe stays as work
uncovered-lines:
    - {range: "AGENTS.md:1-9", reason: the machine's own hook, refs: [D1]}
    - {range: "kept.txt:1-1", reason: the row's own output, refs: [D2]}
behaviors:
    - id: tool/keep
      name: The keeper
      prose: Keeps the probe.
      claims:
        - probe.txt:1-1
      covers: [R2]
      rows:
        - run: sh -c 'printf "%%s %%s\n" %s %s; cat probe.txt > kept.txt'
          expect:
            files:
                kept.txt: ok
`, claimsTokenOf(t, root, "probe.txt:1-1"), demandTokenOf("keep the probe")))
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if strings.Contains(report, "the tree changed under") {
		t.Fatalf("a work intent lifts the lock's hold:\n%s", report)
	}
	if state := zoneFile(t, root, zone.StateFile); !strings.Contains(state, "verdict: #1") {
		t.Fatalf("the lifted lock must let the verdict land, state:\n%s", state)
	}
}

func TestBuildCommandLiftsLock(t *testing.T) {
	mountPorts()
	root := newRoot(t)
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	// A standing build is a work attribute: its artifacts are legal,
	// the lock never arms over a zone that builds.
	writeZone(t, root, zone.InboxSlot, fmt.Sprintf(`
work: the built account of the check
build: "true"
intents:
    - {lines: [D1], outcome: work, prose: the tool and its check under a build}
requirements:
    - id: R1
      lines: [D1]
      prose: the tool and its check
uncovered-lines:
    - {range: "AGENTS.md:1-9", reason: the machine's own hook, refs: [D1]}
behaviors:
    - id: tool/probe
      name: The probe
      prose: Describes by writing its probe.
      claims:
        - src/in.txt:1-1
      covers: [R1]
      rows:
        - run: sh -c 'printf "%%s %%s\n" %s %s; cat src/in.txt > probe.txt'
          expect:
            files:
                probe.txt: ok
`, claimsTokenOf(t, root, "src/in.txt:1-1"), demandTokenOf("a checked tool")))
	runIn(t, root)
	if locks := treeLockRecords(t, root); len(locks) != 0 {
		t.Fatalf("a standing build must keep the lock disarmed, got %d facts", len(locks))
	}
}

func TestExecutionRunsAndJournals(t *testing.T) {
	root := newBornRoot(t)
	writeZone(t, root, zone.DemandsSlot, "a checked tool\nwhich word does printf carry?\n")
	writeZone(t, root, zone.InboxSlot, `work: the machine runs the command itself
intents:
    - {lines: [D2], outcome: execute, prose: the command's own output answers, run: "printf hi"}
`)
	runIn(t, root)
	records, err := journalRecords(t, root)
	if err != nil {
		t.Fatal(err)
	}
	var fact *journal.ExecutionFact
	for _, rec := range records {
		if rec.Kind == journal.KindExecution && rec.Execution != nil {
			f := *rec.Execution
			fact = &f
		}
	}
	if fact == nil || fact.Exit != 0 || fact.Stdout != "hi" || fact.StdoutDigest == "" {
		t.Fatalf("the execution must journal its exit, output and digest: %+v", fact)
	}
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "executions: last run 2 — D2 exit 0") {
		t.Fatalf("the state must carry the execution's own line, state:\n%s", state)
	}
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "- D2 — printf hi — exit 0") || !strings.Contains(report, "stdout: hi") {
		t.Fatalf("the report must carry the execution's whole output, report:\n%s", report)
	}
}

func TestExecutionArtifactsBreakTreeLock(t *testing.T) {
	root := newBornRoot(t)
	// The execution leaves its junk behind: the tree moves under an
	// execute intent — the lock's fact, the verdict's hold.
	writeZone(t, root, zone.DemandsSlot, "a checked tool\nrun the version probe\n")
	writeZone(t, root, zone.InboxSlot, `work: the machine runs the probe
intents:
    - {lines: [D2], outcome: execute, prose: the probe's own output answers, run: "printf junk > junk.tmp"}
`)
	runIn(t, root)
	if locks := treeLockRecords(t, root); len(locks) != 1 {
		t.Fatalf("the execution's artifact must break the lock, got %d facts", len(locks))
	}
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "verdict: not yet — the tree changed under answer/execute/describe intents (law:demand)") {
		t.Fatalf("the leftover artifact must hold the verdict:\n%s", report)
	}
	// The junk leaves: the hold lifts, the verdict lands.
	if err := os.Remove(filepath.Join(root, "junk.tmp")); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	if state := zoneFile(t, root, zone.StateFile); !strings.Contains(state, "verdict: #1") {
		t.Fatalf("the cleaned tree must let the verdict land, state:\n%s", state)
	}
}

func TestExecutionLongOutputGoesToObject(t *testing.T) {
	root := newBornRoot(t)
	writeZone(t, root, zone.DemandsSlot, "a checked tool\nprint a long line\n")
	writeZone(t, root, zone.InboxSlot, `work: the machine runs the long command
intents:
    - {lines: [D2], outcome: execute, prose: the long output answers, run: "awk 'BEGIN{for(i=0;i<5000;i++)printf \"x\"}'"}
`)
	runIn(t, root)
	records, err := journalRecords(t, root)
	if err != nil {
		t.Fatal(err)
	}
	var fact *journal.ExecutionFact
	for _, rec := range records {
		if rec.Kind == journal.KindExecution && rec.Execution != nil {
			f := *rec.Execution
			fact = &f
		}
	}
	if fact == nil || fact.Stdout != "" || fact.StdoutAddr == "" {
		t.Fatalf("the long output must go to the object store, fact: %+v", fact)
	}
	objects := objectsOf(t, root)
	data, err := objects.Get(fact.StdoutAddr)
	if err != nil || len(data) != 5000 || strings.Trim(string(data), "x") != "" {
		t.Fatalf("the object must carry the whole output: %v %d", err, len(data))
	}
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "stdout: "+fact.StdoutAddr) {
		t.Fatalf("the report must name the object for the long output, report:\n%s", report)
	}
}

// TestTreeLockWalkFaultsAreLoud — the lock's two digests fail loudly:
// the walk ahead of the armed lock, and the walk after the
// acceptance moved the journal.
func TestTreeLockWalkFaultsAreLoud(t *testing.T) {
	mountPorts()
	root := describeRoot(t)
	writeZone(t, root, zone.InboxSlot, describeDoc())
	code, term := runCode(t, root, map[string]bool{"walk:" + root: true})
	if code != 1 || !strings.Contains(term, "punchtape: ") {
		t.Fatalf("the before-digest's walk fault must fail the run: %d %q", code, term)
	}
	writeZone(t, root, zone.InboxSlot, describeDoc())
	code, term = runCode(t, root, map[string]bool{"walk:" + root + "!after-write": true})
	if code != 1 || !strings.Contains(term, "punchtape: ") {
		t.Fatalf("the after-digest's walk fault must fail the run: %d %q", code, term)
	}
}

// TestTreeLockFactAppendFailureIsLoud — a moved tree under a locked
// submission whose fact cannot journal fails the run: an unjournaled
// lock is not a lock it had.
func TestTreeLockFactAppendFailureIsLoud(t *testing.T) {
	mountPorts()
	root := describeRoot(t)
	// The writing row moves the tree; only the lock fact's own append
	// fails — the earlier appends land.
	writeZone(t, root, zone.InboxSlot, fmt.Sprintf(`
work: the describe account with a writing proof
intents:
    - {lines: [D1], outcome: describe, prose: the tool told back without touching it}
requirements:
    - id: R1
      lines: [D1]
      prose: the tool and its check
uncovered-lines:
    - {range: "AGENTS.md:1-9", reason: the machine's own hook, refs: [D1]}
behaviors:
    - id: tool/probe
      name: The probe
      prose: Describes by writing its probe.
      claims:
        - src/in.txt:1-1
      covers: [R1]
      rows:
        - run: sh -c 'printf "%%s %%s\n" %s %s; cat src/in.txt > probe.txt'
          expect:
            files:
                probe.txt: ok
`, claimsTokenOf(t, root, "src/in.txt:1-1"), demandTokenOf("a checked tool")))
	code, term := runCode(t, root, map[string]bool{"append-content:kind: tree-lock": true})
	if code != 1 || !strings.Contains(term, "journal write:") {
		t.Fatalf("the unjournaled lock fact must fail the run: %d %q", code, term)
	}
}

// TestTreeLockSeesUnreadableAndVCS — the lock's walk reads the same
// selection the coverage sees: an unreadable file spells unreadable
// on both sides of a clean locked run, the vcs bookkeeping stays out,
// and a directory the walk cannot enter fails the run loudly.
func TestTreeLockSeesUnreadableAndVCS(t *testing.T) {
	mountPorts()
	sealed := describeRoot(t)
	if err := os.WriteFile(filepath.Join(sealed, "locked.txt"), []byte("secret\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Join(sealed, "locked.txt"), 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(filepath.Join(sealed, "locked.txt"), 0o644) })
	if err := os.MkdirAll(filepath.Join(sealed, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sealed, ".git", "metadata"), []byte("noise"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, sealed, zone.InboxSlot, describeDoc())
	runIn(t, sealed)
	if locks := treeLockRecords(t, sealed); len(locks) != 0 {
		t.Fatalf("an unreadable file spelled on both sides is no move, got %d facts", len(locks))
	}
	// A directory the walk cannot enter: the walk's callback carries
	// the fault and the run answers it loudly — an unreadable tree is
	// never digested in silence.
	vaulted := describeRoot(t)
	if err := os.MkdirAll(filepath.Join(vaulted, "vault"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(vaulted, "vault", "key.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Join(vaulted, "vault"), 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(filepath.Join(vaulted, "vault"), 0o755) })
	writeZone(t, vaulted, zone.InboxSlot, describeDoc())
	code, term := runCode(t, vaulted, nil)
	if code != 1 || !strings.Contains(term, "punchtape: ") {
		t.Fatalf("the unreadable directory must fail the run loudly: %d %q", code, term)
	}
}

// describeRoot — a born zone ready for a locked describe
// submission over its own input.
func describeRoot(t *testing.T) string {
	t.Helper()
	root := newRoot(t)
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

// describeDoc — the locked submission of the describe class: the
// requirement account over the line, no behavior, no tree move.
func describeDoc() string {
	return `
work: the describe account without a proof
intents:
    - {lines: [D1], outcome: describe, prose: the tool told back without touching it}
requirements:
    - id: R1
      lines: [D1]
      prose: the tool and its check
uncovered-lines:
    - {range: "AGENTS.md:1-9", reason: the machine's own hook, refs: [D1]}
    - {range: "src/in.txt:1-1", reason: the described input itself, refs: [D1]}
`
}
