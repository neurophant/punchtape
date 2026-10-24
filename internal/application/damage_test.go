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

// birthAddr — the goal object's address from the journal's first
// record.
func birthAddr(t *testing.T, root string) string {
	t.Helper()
	data, err := os.ReadFile(epochPaths(t, root)[0])
	if err != nil {
		t.Fatal(err)
	}
	chunk := bytes.TrimPrefix(data, []byte("---\n"))
	if i := bytes.Index(chunk, []byte("\n---\n")); i >= 0 {
		chunk = chunk[:i+1]
	}
	var rec journal.Record
	if err := journal.DecodeRecord(chunk, &rec); err != nil {
		t.Fatal(err)
	}
	return rec.Birth.Addr
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
	// A hand-written record names the goal object as a behavior: the
	// fold refuses the lie by name.
	appendRecord(t, root, journal.Record{Kind: journal.KindSubmission, Submission: &journal.Submission{
		CorpusVersion: 2, DocAddr: birthAddr(t, root),
		Added: []journal.BehaviorOp{{ID: "tool/fake", Addr: birthAddr(t, root)}},
	}})
	var term strings.Builder
	if code := Run(root, &term); code != 1 || !strings.Contains(term.String(), "journal fold") {
		t.Fatalf("the forged address must refuse the open: code %d term %q", code, term.String())
	}
}

func TestLyingDocAddressSkipsTheNote(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	// The doc address points at the goal object: not a document — the
	// note extraction skips it, the run stands.
	appendRecord(t, root, journal.Record{Kind: journal.KindSubmission, Submission: &journal.Submission{
		CorpusVersion: 2, DocAddr: birthAddr(t, root),
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

func TestGarbageCacheRecomputesHonestly(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	writeZone(t, root, "cache/hashes.yaml", ":: not yaml [")
	runIn(t, root)
	data, err := os.ReadFile(filepath.Join(root, zone.Dirname, "cache", "hashes.yaml"))
	if err != nil || !strings.Contains(string(data), "files:") {
		t.Fatalf("the cache must heal: %v %s", err, data)
	}
}

func TestBirthRecordsTheVersion(t *testing.T) {
	mountPorts()
	root := newRoot(t)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "machine at birth: "+Version) {
		t.Fatalf("the state must name the creating machine, state:\n%s", state)
	}
	// The journal's first record carries the same fact.
	chunk := bytes.TrimPrefix(mustRead(t, epochPaths(t, root)[0]), []byte("---\n"))
	var rec journal.Record
	if err := journal.DecodeRecord(chunk, &rec); err != nil {
		t.Fatal(err)
	}
	if rec.Birth.Version != Version {
		t.Fatalf("the birth record must carry the machine version: %+v", rec.Birth)
	}
	// An instance born before the field existed says so honestly.
	first := epochPaths(t, root)[0]
	old := journal.MarshalRecord(journal.Record{
		Kind:  journal.KindBirth,
		Birth: &journal.Birth{Addr: rec.Birth.Addr, Digest: rec.Birth.Digest},
	})
	if err := os.WriteFile(first, append([]byte("---\n"), old...), 0o644); err != nil {
		t.Fatal(err)
	}
	runIn(t, root)
	if !strings.Contains(zoneFile(t, root, zone.StateFile), "machine at birth: (not recorded)") {
		t.Fatalf("an old birth must say it is not recorded, state:\n%s", zoneFile(t, root, zone.StateFile))
	}
}

// mustRead — a file's bytes or the test dies.
func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// TestForeignLawCacheEntriesDropOut — a cache written under another
// digest law (no algo tag) is not the machine's truth: every entry
// drops out and recomputes under the standing law, the honest bytes'
// sha256.
func TestForeignLawCacheEntriesDropOut(t *testing.T) {
	mountPorts()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "f.txt"), []byte("plain"), 0o644); err != nil {
		t.Fatal(err)
	}
	zoneDir := filepath.Join(root, ".z")
	if err := os.MkdirAll(zoneDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// A valid YAML of the old law: right shape, no algo tag, a lying
	// digest.
	old := "files:\n  f.txt:\n    size: 5\n    mtime: 1\n    digest: deadbeef\n"
	if err := os.MkdirAll(filepath.Join(zoneDir, zone.CacheDir), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(zoneDir, zone.CacheDir, "hashes.yaml"), []byte(old), 0o644); err != nil {
		t.Fatal(err)
	}
	fresh := OpenHashCache(zoneDir)
	if got := fresh.Digest(root, "f.txt"); got != domain.Digest([]byte("plain")) {
		t.Fatalf("the foreign entry must not answer: %s", got)
	}
	fresh.Flush()
	data, err := os.ReadFile(filepath.Join(zoneDir, zone.CacheDir, "hashes.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "algo: 2") || strings.Contains(string(data), "deadbeef") {
		t.Fatalf("the flushed cache speaks only the standing law:\n%s", data)
	}
	// The standing law's own entry is a hit: the same size and mtime
	// answer from the cache, the digest unchanged.
	if again := fresh.Digest(root, "f.txt"); again != domain.Digest([]byte("plain")) {
		t.Fatalf("the cache hit must answer the same digest: %s", again)
	}
}
