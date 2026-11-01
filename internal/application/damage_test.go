// The damaged-zone battery: a journal forged or cut by hand — the
// machine answers with an honest refusal or an honest load, never a
// guess.
package application

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neurophant/punchtape/internal/domain"
	"github.com/neurophant/punchtape/internal/domain/doc"
	"github.com/neurophant/punchtape/internal/domain/journal"
	"github.com/neurophant/punchtape/internal/domain/zone"
)

// epochPaths — the zone's epoch files, in listing order.
func epochPaths(t *testing.T, root string) []string {
	t.Helper()
	dir := filepath.Join(root, zone.Dirname, zone.JournalDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, e := range entries {
		if !e.IsDir() {
			out = append(out, filepath.Join(dir, e.Name()))
		}
	}
	return out
}

// objectAddr — an object address of the standing's first submission:
// its document object (not a behavior) or its first behavior object
// (not a document) — the forger's honest material.
func objectAddr(t *testing.T, root string, pick func(*journal.Submission) string) string {
	t.Helper()
	records, err := journalRecords(t, root)
	if err != nil {
		t.Fatal(err)
	}
	for _, rec := range records {
		if rec.Kind == journal.KindSubmission && rec.Submission != nil {
			if addr := pick(rec.Submission); addr != "" {
				return addr
			}
		}
	}
	t.Fatal("the journal carries no submission object")
	return ""
}

// docAddr — the standing's first document object's address.
func docAddr(t *testing.T, root string) string {
	t.Helper()
	return objectAddr(t, root, func(s *journal.Submission) string { return s.DocAddr })
}

// behaviorAddr — the standing's first behavior object's address.
func behaviorAddr(t *testing.T, root string) string {
	t.Helper()
	return objectAddr(t, root, func(s *journal.Submission) string {
		if len(s.Added) > 0 {
			return s.Added[0].Addr
		}
		return ""
	})
}

// appendRecord — a hand-written record into the live epoch.
func appendRecord(t *testing.T, root string, rec journal.Record) {
	t.Helper()
	data := journal.MarshalRecord(rec)
	last := epochPaths(t, root)[len(epochPaths(t, root))-1]
	f, err := os.OpenFile(last, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.WriteString("---\n" + string(data)); err != nil {
		t.Fatal(err)
	}
}

func TestForgedAddedAddressRefusesTheOpen(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	// A hand-written record names the document object as a behavior:
	// the fold refuses the lie by name.
	appendRecord(t, root, journal.Record{Kind: journal.KindSubmission, Submission: &journal.Submission{
		CorpusVersion: 2, DocAddr: docAddr(t, root),
		Added: []journal.BehaviorOp{{ID: "tool/fake", Addr: docAddr(t, root)}},
	}})
	var term strings.Builder
	if code := Run(root, &term); code != 1 || !strings.Contains(term.String(), "journal fold") {
		t.Fatalf("the forged address must refuse the open: code %d term %q", code, term.String())
	}
}

func TestLyingDocAddressSkipsTheNote(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	// The doc address points at a behavior object: not a document —
	// the note extraction skips it, the run stands.
	appendRecord(t, root, journal.Record{Kind: journal.KindSubmission, Submission: &journal.Submission{
		CorpusVersion: 2, DocAddr: behaviorAddr(t, root),
	}})
	runIn(t, root)
	if knowledge := zoneFile(t, root, zone.KnowledgeFile); !strings.Contains(knowledge, "no notes yet") {
		t.Fatalf("a lying doc address must leave no notes:\n%s", knowledge)
	}
}

func TestTornTailQuarantineFaultsAreNamed(t *testing.T) {
	root := newBornRoot(t)
	last := epochPaths(t, root)[len(epochPaths(t, root))-1]
	data, _ := os.ReadFile(last)
	// A record cut mid-write at the tail.
	torn := append(append([]byte(nil), data...), []byte("---\nkind: run\nrun:\n    n: ")...)
	if err := os.WriteFile(last, torn, 0o644); err != nil {
		t.Fatal(err)
	}
	// The sidecar cannot be written.
	code, _ := runCode(t, root, map[string]bool{"append:.torn": true})
	if code != 1 {
		t.Fatalf("the quarantine's sidecar fault must refuse: code %d", code)
	}
	// The epoch's repair write cannot happen.
	code, _ = runCode(t, root, map[string]bool{"write:epoch-": true})
	if code != 1 {
		t.Fatalf("the quarantine's repair fault must refuse: code %d", code)
	}
}

func TestEpochWithoutLeadingSeparatorLoadsItsRecords(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	last := epochPaths(t, root)[len(epochPaths(t, root))-1]
	data, _ := os.ReadFile(last)
	stripped := bytes.TrimPrefix(data, []byte("---\n"))
	// No leading separator and a junk tail: the intact record loads,
	// the junk is no record, and nothing identifies a torn cut.
	broken := append(append([]byte(nil), stripped...), []byte("---\njunk: [\n")...)
	if err := os.WriteFile(last, broken, 0o644); err != nil {
		t.Fatal(err)
	}
	runIn(t, root)
}

func TestGarbageInboxIsARefusal(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	writeZone(t, root, zone.InboxSlot, "\tbroken: [")
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "law:document") {
		t.Fatalf("the decode refusal must cite the document law:\n%s", report)
	}
}

func TestBirthRecordsTheVersion(t *testing.T) {
	mountPorts()
	root := newRoot(t)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "machine at birth: "+Version) {
		t.Fatalf("the state must name the creating machine, state:\n%s", state)
	}
	// The journal's first records carry the same fact, and the
	// stream's genesis link rides beside the birth — the chain value
	// over all the initial lines, from the empty previous.
	records, err := journalRecords(t, root)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) < 2 || records[0].Kind != journal.KindBirth || records[0].Birth.Version != Version {
		t.Fatalf("the birth record must carry the machine version: %+v", records)
	}
	link := records[1]
	if link.Kind != journal.KindDemandAppend || link.DemandAppend.Lines != 1 ||
		link.DemandAppend.Chain != chainOver("", []string{"a checked tool"}) {
		t.Fatalf("the genesis link must cover the initial lines: %+v", link.DemandAppend)
	}
}

// TestSpecSkipsADanglingCover — a hand-forged record may add a
// behavior whose cover names a requirement no journal carries: the
// fold trusts the object's shape; the spec page's demands section
// skips the dangling cover and keeps its one order.
func TestSpecSkipsADanglingCover(t *testing.T) {
	root := newBornRoot(t)
	rc := 0
	b := doc.Behavior{
		ID: "tool/fake", Name: "Fake", Prose: "A forged head.",
		Claims: []string{"src/in.txt:1-1"}, Covers: []string{"R9"},
		Rows: []doc.Row{{Run: doc.Cmd{Line: "true"}, Expect: doc.Expect{RC: &rc}}},
	}
	data := doc.MarshalBehavior(b)
	addr := domain.Digest(data)
	objPath := filepath.Join(root, zone.Dirname, zone.ObjectsDir, addr[:2], addr[2:])
	if err := os.MkdirAll(filepath.Dir(objPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(objPath, data, 0o644); err != nil {
		t.Fatal(err)
	}
	appendRecord(t, root, journal.Record{Kind: journal.KindSubmission, Submission: &journal.Submission{
		CorpusVersion: 2, DocAddr: docAddr(t, root),
		Added: []journal.BehaviorOp{{ID: "tool/fake", Addr: addr}},
	}})
	runIn(t, root)
	page := zoneFile(t, root, filepath.Join(zone.SpecDir, "tool", "fake.md"))
	if strings.Contains(page, "- R9:") {
		t.Fatalf("a dangling cover must not print, page:\n%s", page)
	}
	if !strings.Contains(page, "## Demands\n\n\n## Claims") {
		t.Fatalf("the demands section must stand empty in its order, page:\n%s", page)
	}
}
