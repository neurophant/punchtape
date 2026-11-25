// The verify call's battery: the human's re-proof of a zone — nothing
// is written, a forged journal turns red, a broken row is named, a
// partial run and every refusal are not proven, the inbox stays
// unread, and every row really runs.
package application

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neurophant/punchtape/internal/domain/doc"
	"github.com/neurophant/punchtape/internal/domain/fprint"
	"github.com/neurophant/punchtape/internal/domain/journal"
	"github.com/neurophant/punchtape/internal/domain/zone"
	"github.com/neurophant/punchtape/internal/infrastructure/process"
	"github.com/neurophant/punchtape/internal/infrastructure/store"
	"github.com/neurophant/punchtape/internal/ports"
)

// The measured scene: a hello script, and the behavior whose one row
// runs it and pins its exact output. The row writes nothing to the
// tree, so a check that writes nothing leaves every file as it stood.
const (
	helloScript   = "#!/bin/sh\nprintf \"hello\\n\"\n"
	goodbyeScript = "#!/bin/sh\nprintf \"goodbye\\n\"\n"

	// noJudgments — the judgments phrase of a zone that carries neither
	// a mark nor an exclusion over a one-line stream.
	noJudgments = "0 of 1 words marked not-a-requirement, 0 % of the tree excluded"
)

// provenRoot — the honest zone of the scene: the script written, its
// behavior submitted, verdict #1 standing.
func provenRoot(t *testing.T) string {
	t.Helper()
	mountPorts()
	root := newRoot(t)
	if err := writeTree(t, root, "hello", helloScript); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Join(root, "hello"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, `
work: the hello script, pinned to the demand's exact words
intents:
    - {lines: [D1], outcome: work, prose: the hello script and its check}
requirements:
    - {id: R1, lines: [D1], prose: the hello script}
behaviors:
    - id: hello/say
      name: say hello
      prose: prints hello and nothing else
      claims: ["hello:1-2"]
      covers: [R1]
      rows:
        - run: "./hello"
          expect:
            out: "hello\n"
`)
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	return root
}

// breakHello — the script says goodbye: its claimed lines move.
func breakHello(t *testing.T, root string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, "hello"), []byte(goodbyeScript), 0o755); err != nil {
		t.Fatal(err)
	}
}

// treeStamp — every file under the root with its mtime and its
// bytes' digest: what a check that writes nothing must leave untouched.
func treeStamp(t *testing.T, root string) map[string]string {
	t.Helper()
	stamp := map[string]string{}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		stamp[rel] = fmt.Sprintf("%d %x", info.ModTime().UnixNano(), sha256.Sum256(data))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return stamp
}

// verifyIn — the check's code and terminal.
func verifyIn(root string) (int, string) {
	var term strings.Builder
	code := Verify(root, &term)
	return code, term.String()
}

// verifyCode — the check under a failing filesystem; the production
// process port stays.
func verifyCode(t *testing.T, root string, fail map[string]bool) (int, string) {
	t.Helper()
	ports.FSBind = newFailFS(fail)
	ports.ExecBind = process.OS{}
	defer mountPorts()
	var term strings.Builder
	code := Verify(root, &term)
	return code, term.String()
}

// forgeGreen — the forger's two values for a behavior whose claimed
// file moved: the claims digest over the tree as it now stands, and
// the proof of its first row over that digest — what a green fact and
// its verdict scope would carry had the row really run green here.
func forgeGreen(t *testing.T, m *machine, id string) (proof, pointer string) {
	t.Helper()
	corpus := m.corpusDoc()
	pointer = m.claimsMap(corpus)[id].digest
	for _, b := range corpus.Behaviors {
		if b.ID == id {
			return fprint.Proof(doc.MarshalRow(b.Rows[0]), pointer, EnvHash(m.pins)), pointer
		}
	}
	t.Fatalf("the behavior %s does not stand", id)
	return "", ""
}

// TestVerifyWritesNothing — an honest zone is proven, and the check
// leaves every file's bytes and mtime as they stood; the zone's lock
// file alone may differ.
func TestVerifyWritesNothing(t *testing.T) {
	root := provenRoot(t)
	before := treeStamp(t, root)
	code, term := verifyIn(root)
	if code != 0 || !strings.HasSuffix(term, "punchtape: verify — proven — 1 rows run green — "+noJudgments+"\n") {
		t.Fatalf("an honest zone must be proven, code %d:\n%s", code, term)
	}
	after := treeStamp(t, root)
	lock := filepath.Join(zone.Dirname, zone.RunLockFile)
	delete(before, lock)
	delete(after, lock)
	var moved []string
	for path, stamp := range before {
		if after[path] != stamp {
			moved = append(moved, path)
		}
	}
	for path := range after {
		if _, stood := before[path]; !stood {
			moved = append(moved, path)
		}
	}
	if len(moved) > 0 {
		t.Fatalf("the check must write nothing, moved: %v", moved)
	}
}

// TestVerifyCatchesAForgedJournal — the measured forgery: the script
// says goodbye, and the journal's two values are recomputed to agree
// with the moved tree. The ordinary call believes it — the verdict
// stands; the check runs the row and names it red.
func TestVerifyCatchesAForgedJournal(t *testing.T) {
	root := provenRoot(t)
	breakHello(t, root)
	stream, err := store.OpenStream(filepath.Join(root, zone.Dirname, zone.JournalDir))
	if err != nil {
		t.Fatal(err)
	}
	m, err := open(root, root, filepath.Join(root, zone.Dirname), stream)
	if err != nil {
		t.Fatal(err)
	}
	proof, pointer := forgeGreen(t, m, "hello/say")
	oldProof := m.standing.Rows["hello/say/0"].Proof
	oldPointer := m.standing.Verdict.Scopes[0].Pointer
	for _, epoch := range epochPaths(t, root) {
		data, err := os.ReadFile(epoch)
		if err != nil {
			t.Fatal(err)
		}
		data = bytes.ReplaceAll(data, []byte("proof: "+oldProof), []byte("proof: "+proof))
		data = bytes.ReplaceAll(data, []byte("pointer: "+oldPointer), []byte("pointer: "+pointer))
		if err := os.WriteFile(epoch, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if code, term := runCode(t, root, nil); code != 0 || !strings.Contains(term, "verdict #1 standing") {
		t.Fatalf("the forged journal must fool the ordinary call, code %d:\n%s", code, term)
	}
	code, term := verifyIn(root)
	if code != 1 || !strings.Contains(term, "red hello/say/0 — ") {
		t.Fatalf("the check must name the forged row red, code %d:\n%s", code, term)
	}
}

// TestVerifyNamesABrokenRow — the script says goodbye, the journal
// untouched: the check runs the row and names it red.
func TestVerifyNamesABrokenRow(t *testing.T) {
	root := provenRoot(t)
	breakHello(t, root)
	code, term := verifyIn(root)
	if code != 1 || !strings.Contains(term, "red hello/say/0 — ") {
		t.Fatalf("the broken row must be named red, code %d:\n%s", code, term)
	}
}

// TestVerifyFailsOnAPartialRun — a budget that cannot fit every row
// is not proven: the check names what waited, whatever the journal's
// facts say about it.
func TestVerifyFailsOnAPartialRun(t *testing.T) {
	root := newRoot(t)
	writeZone(t, root, zone.EnvSlot, "PUNCHTAPE_ROW_TIMEOUT=1\nPUNCHTAPE_RUN_BUDGET=1\n")
	writeZone(t, root, zone.InboxSlot, `
work: two fast rows under a one-second budget
intents:
    - {lines: [D1], outcome: work, prose: the line is declared as work}
requirements:
    - {id: R1, lines: [D1], prose: the fast rows}
behaviors:
    - id: fast/one
      name: fast
      prose: two printing rows
      covers: [R1]
      rows:
        - {run: "printf 'one\\n'", expect: {out: "one\n"}}
        - {run: "printf 'two\\n'", expect: {out: "two\n"}}
`)
	runIn(t, root)
	code, term := verifyIn(root)
	if code != 1 || !strings.Contains(term, "run budget spent") ||
		!strings.HasSuffix(term, "punchtape: verify — not proven — "+noJudgments+"\n") {
		t.Fatalf("a partial run is not proven, code %d:\n%s", code, term)
	}
}

// TestVerifyLeavesATornTailAlone — a torn tail is damage to the
// check, never a repair: the epoch's bytes stand as they were and no
// sidecar is written.
func TestVerifyLeavesATornTailAlone(t *testing.T) {
	root := provenRoot(t)
	epochs := epochPaths(t, root)
	live := epochs[len(epochs)-1]
	f, err := os.OpenFile(live, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString("---\nkind: run\nn: 99\ncut mid-wri"); err != nil {
		t.Fatal(err)
	}
	f.Close()
	torn, err := os.ReadFile(live)
	if err != nil {
		t.Fatal(err)
	}
	code, term := verifyIn(root)
	if code != 1 || !strings.Contains(term, "does not parse") {
		t.Fatalf("a torn tail must refuse the check by name, code %d:\n%s", code, term)
	}
	after, err := os.ReadFile(live)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(after, torn) {
		t.Fatal("the check must leave the torn epoch's bytes alone")
	}
	if _, err := os.Stat(live + ".torn"); !os.IsNotExist(err) {
		t.Fatalf("the check must write no quarantine: %v", err)
	}
}

// TestVerifyRefusesWhatItCannotProve — every answer but "proven": no
// zone, no instance, a foreign machine, an unreadable journal, a
// moved past line, a tree that will not walk, an empty corpus, a lock
// that never came, a red build. Each answers one.
func TestVerifyRefusesWhatItCannotProve(t *testing.T) {
	// No zone: nothing is created.
	bare := t.TempDir()
	code, term := verifyIn(bare)
	if code != 1 || !strings.Contains(term, "no zone here") {
		t.Fatalf("no zone: code %d, term %q", code, term)
	}
	if _, err := os.Stat(filepath.Join(bare, zone.Dirname)); !os.IsNotExist(err) {
		t.Fatalf("the check without a zone creates nothing: %v", err)
	}

	// A zone without an instance: no journal, no lock file.
	fresh := t.TempDir()
	writeZone(t, fresh, zone.DemandsSlot, "a checked tool")
	code, term = verifyIn(fresh)
	if code != 1 || !strings.Contains(term, "no instance here yet") {
		t.Fatalf("no instance: code %d, term %q", code, term)
	}
	for _, name := range []string{zone.JournalDir, zone.RunLockFile} {
		if _, err := os.Stat(filepath.Join(fresh, zone.Dirname, name)); !os.IsNotExist(err) {
			t.Fatalf("the check without an instance must not leave %s: %v", name, err)
		}
	}

	// A zone born by another machine version: the ordinary call's own
	// line, word for word.
	foreign := provenRoot(t)
	records, err := journalRecords(t, foreign)
	if err != nil {
		t.Fatal(err)
	}
	birth := journal.Record{Kind: journal.KindBirth, Birth: &journal.Birth{Version: "9.9.9"}}
	var epoch bytes.Buffer
	epoch.WriteString("---\n")
	epoch.Write(journal.MarshalRecord(birth))
	epoch.WriteString("---\n")
	epoch.Write(journal.MarshalRecord(records[1]))
	if err := os.WriteFile(epochPaths(t, foreign)[0], epoch.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	_, runTerm := runCode(t, foreign, nil)
	code, term = verifyIn(foreign)
	if code != 1 || term != runTerm {
		t.Fatalf("a foreign zone: the check must answer the ordinary call's line, code %d:\nrun    %q\nverify %q", code, runTerm, term)
	}

	// An unreadable journal directory: zone damage, named.
	root := provenRoot(t)
	code, term = verifyCode(t, root, map[string]bool{"stat:journal": true})
	if code != 1 || !strings.Contains(term, "zone damaged") {
		t.Fatalf("an unreadable journal: code %d, term %q", code, term)
	}

	// A moved past line: the chain's own damage.
	writeZone(t, root, zone.DemandsSlot, "a broken tool")
	code, term = verifyIn(root)
	if code != 1 || !strings.Contains(term, "zone damaged: the demand stream's past changed") {
		t.Fatalf("a moved past line: code %d, term %q", code, term)
	}

	// A tree that will not walk: the refusal is loud, and no verdict
	// is spoken.
	unwalkable := provenRoot(t)
	code, term = verifyCode(t, unwalkable, map[string]bool{"walk:" + unwalkable: true})
	if code != 1 || !strings.HasPrefix(term, "punchtape: ") || strings.Contains(term, "verify —") {
		t.Fatalf("an unwalkable tree: code %d, term %q", code, term)
	}

	// An empty corpus: nothing stands to prove, and the check says so.
	empty := newRoot(t)
	code, term = verifyIn(empty)
	if code != 1 || !strings.Contains(term, "verdict: not yet — the corpus is empty\n") ||
		!strings.HasSuffix(term, "punchtape: verify — not proven — "+noJudgments+"\n") {
		t.Fatalf("an empty corpus: code %d, term %q", code, term)
	}

	// A lock that never came: the run's own refusal.
	locked := provenRoot(t)
	code, term = verifyCode(t, locked, map[string]bool{"flock:run.lock": true})
	if code != 1 || !strings.Contains(term, "punchtape: zone lock:") {
		t.Fatalf("an untakeable lock: code %d, term %q", code, term)
	}

	// A red build: the rows never run, and the first gate speaks.
	red := newRoot(t)
	writeZone(t, red, zone.InboxSlot, `
work: a red build over a row that never runs
build: "false"
behaviors:
    - id: wait/row
      name: wait
      prose: leaves its witness in the tree
      rows:
        - {run: "sh -c 'echo ran > witness.txt && echo done'", expect: {out: "done\n"}}
`)
	runIn(t, red)
	code, term = verifyIn(red)
	if code != 1 || !strings.Contains(term, "verdict: not yet — the build is red\n") {
		t.Fatalf("a red build: code %d, term %q", code, term)
	}
	if _, err := os.Stat(filepath.Join(red, "witness.txt")); !os.IsNotExist(err) {
		t.Fatalf("a red build must not let the row run: %v", err)
	}
}

// TestVerifyDoesNotReadTheInbox — the completion act stands in the
// slot and the check never touches it: proven over the standing
// corpus, the slot byte for byte as it was left.
func TestVerifyDoesNotReadTheInbox(t *testing.T) {
	root := provenRoot(t)
	act := "done: true\n"
	writeZone(t, root, zone.InboxSlot, act)
	code, term := verifyIn(root)
	if code != 0 {
		t.Fatalf("the inbox must not hold the check, code %d:\n%s", code, term)
	}
	slot, err := os.ReadFile(filepath.Join(root, zone.Dirname, zone.InboxSlot))
	if err != nil {
		t.Fatal(err)
	}
	if string(slot) != act {
		t.Fatalf("the check must leave the inbox byte for byte, got %q", slot)
	}
}

// TestVerifyRunsEveryRow — a row whose fingerprint stands in the
// journal is not run by the ordinary call; the check runs it anyway,
// witness and all.
func TestVerifyRunsEveryRow(t *testing.T) {
	root := newRoot(t)
	writeZone(t, root, zone.UniverseSlot, "src\n")
	if err := writeTree(t, root, "src/in.txt", "one line\n"); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, `
work: a row that leaves its witness
intents:
    - {lines: [D1], outcome: work, prose: the row and its witness}
requirements:
    - {id: R1, lines: [D1], prose: the witness row}
behaviors:
    - id: src/witness
      name: witness
      prose: leaves its witness in the tree
      claims: ["src/in.txt:1-1"]
      covers: [R1]
      rows:
        - {run: "sh -c 'echo ran > witness.txt && echo done'", expect: {out: "done\n"}}
`)
	runIn(t, root)
	witness := filepath.Join(root, "witness.txt")
	if err := os.Remove(witness); err != nil {
		t.Fatalf("the first run must leave the witness: %v", err)
	}
	if code, term := verifyIn(root); code != 0 {
		t.Fatalf("the check must run the row, code %d:\n%s", code, term)
	}
	if err := os.Remove(witness); err != nil {
		t.Fatalf("the check must leave the witness: %v", err)
	}
	runIn(t, root)
	if _, err := os.Stat(witness); !os.IsNotExist(err) {
		t.Fatalf("the ordinary call must not run a row whose fingerprint stands: %v", err)
	}
}

// TestVerifyClearsTheScratch — the zone's scratch goes the way it goes
// after a run: the temporary files the row left behind the re-proof
// are gone when the check ends, and a scratch that would not clear
// answers one.
func TestVerifyClearsTheScratch(t *testing.T) {
	root := newRoot(t)
	writeZone(t, root, zone.UniverseSlot, "src\n")
	if err := writeTree(t, root, "src/in.txt", "one line\n"); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, `
work: a row that writes into the scratch
intents:
    - {lines: [D1], outcome: work, prose: the row and its temporary file}
requirements:
    - {id: R1, lines: [D1], prose: the scratch row}
behaviors:
    - id: src/scratch
      name: scratch
      prose: leaves a temporary file in the scratch
      claims: ["src/in.txt:1-1"]
      covers: [R1]
      rows:
        - {run: "sh -c 'mkdir -p .punchtape/tmp && echo temp > .punchtape/tmp/note.txt && echo done'", expect: {out: "done\n"}}
`)
	runIn(t, root)
	scratch := filepath.Join(root, zone.Dirname, zone.TmpDir)
	if code, term := verifyIn(root); code != 0 {
		t.Fatalf("the check must run the row, code %d:\n%s", code, term)
	}
	if entries, err := os.ReadDir(scratch); err == nil && len(entries) > 0 {
		t.Fatalf("the check must clear the scratch the row wrote, left: %v", entries)
	}
	// A scratch that would not clear: proven or not, the answer is one.
	code, term := verifyCode(t, root, map[string]bool{"removeall:.punchtape/tmp": true})
	if code != 1 || !strings.Contains(term, "verify — proven") ||
		!strings.HasSuffix(term, "punchtape: tmp clear: permission denied\n") {
		t.Fatalf("an unclearable scratch must answer one, code %d, term %q", code, term)
	}
}

// pointerLine — the run's last terminal line: the instruction's own
// text, as the caller reads it.
func pointerLine(term string) string {
	lines := strings.Split(strings.TrimSuffix(term, "\n"), "\n")
	return lines[len(lines)-1]
}

// judgedRoot — the honest zone of a scene with a judgment to count:
// the hello script claimed and proven, a second file of three lines
// excluded from the ordered work — three of the tree's five lines —
// verdict #1 standing.
func judgedRoot(t *testing.T) string {
	t.Helper()
	mountPorts()
	root := t.TempDir()
	writeZone(t, root, zone.DemandsSlot, "a checked tool\n")
	runIn(t, root)
	if err := writeTree(t, root, "hello", helloScript); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Join(root, "hello"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := writeTree(t, root, "other.txt", "one\ntwo\nthree\n"); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.UniverseSlot, "hello\nother.txt\n")
	writeZone(t, root, zone.InboxSlot, `
work: the hello script and its check; a second file stands outside the ordered work
intents:
    - {lines: [D1], outcome: work, prose: the hello script and its check}
requirements:
    - {id: R1, lines: [D1], prose: the hello script}
uncovered-lines:
    - {range: "other.txt:1-3", reason: outside the ordered work, refs: [D1]}
behaviors:
    - id: hello/say
      name: say hello
      prose: prints hello and nothing else
      claims: ["hello:1-2"]
      covers: [R1]
      rows:
        - run: "./hello"
          expect:
            out: "hello\n"
`)
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runIn(t, root)
	return root
}

// TestStandingLineNamesTheJudgments — the standing instruction counts
// the agent's two judgments where the human reads them: the words
// marked not-a-requirement over the stream's size, and the share of the
// tree excluded. Both are printed always, zeros included.
func TestStandingLineNamesTheJudgments(t *testing.T) {
	// Neither a mark nor an exclusion: both counts stand at zero.
	if line := pointerLine(runIn(t, provenRoot(t))); !strings.Contains(line, "verdict #1 standing — "+noJudgments+" — show ") {
		t.Fatalf("a zone without judgments, line %q", line)
	}

	// Two of the three words of the stream are marked.
	mountPorts()
	marked := t.TempDir()
	writeZone(t, marked, zone.DemandsSlot, "say hello and nothing else\nmake it shout\nadd a counter\n")
	runIn(t, marked)
	if err := writeTree(t, marked, "hello", helloScript); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Join(marked, "hello"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeZone(t, marked, zone.InboxSlot, `
work: the hello script and its check; two lines of the stream order nothing
intents:
    - {lines: [D1], outcome: work, prose: the hello script and its check}
    - {lines: [D2], outcome: mark, prose: "shouting is a wish, not an order"}
    - {lines: [D3], outcome: mark, prose: "a counter is a wish, not an order"}
requirements:
    - {id: R1, lines: [D1], prose: the hello script}
not-a-requirement:
    - {lines: [D2], reason: "a wish, not an order"}
    - {lines: [D3], reason: "a wish, not an order"}
behaviors:
    - id: hello/say
      name: say hello
      prose: prints hello and nothing else
      claims: ["hello:1-2"]
      covers: [R1]
      rows:
        - run: "./hello"
          expect:
            out: "hello\n"
`)
	runIn(t, marked)
	writeZone(t, marked, zone.InboxSlot, "done: true\n")
	if line := pointerLine(runIn(t, marked)); !strings.Contains(line, "verdict #1 standing — 2 of 3 words marked not-a-requirement, 0 % of the tree excluded — show ") {
		t.Fatalf("two marked words, line %q", line)
	}

	// Three of the tree's five lines stand excluded.
	if line := pointerLine(runIn(t, judgedRoot(t))); !strings.Contains(line, "verdict #1 standing — 0 of 1 words marked not-a-requirement, 60 % of the tree excluded — show ") {
		t.Fatalf("an excluded file, line %q", line)
	}
}

// TestVerifyLineNamesTheJudgments — the check's last line carries the
// same phrase as the standing instruction: the judgments are named
// wherever the human reads the verdict.
func TestVerifyLineNamesTheJudgments(t *testing.T) {
	_, term := verifyIn(judgedRoot(t))
	if want := "punchtape: verify — proven — 1 rows run green — 0 of 1 words marked not-a-requirement, 60 % of the tree excluded\n"; !strings.HasSuffix(term, want) {
		t.Fatalf("the check's last line must name the judgments, term %q", term)
	}
}
