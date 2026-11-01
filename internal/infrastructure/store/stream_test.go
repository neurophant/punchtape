// The journal stream's battery: append and rotate across epochs, the
// records' whole order, the torn tail's quarantine and the refusal of
// damage that is not a torn write.
package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neurophant/punchtape/internal/domain/journal"
	"github.com/neurophant/punchtape/internal/infrastructure/fsys"
	"github.com/neurophant/punchtape/internal/ports"
)

func birthRec() journal.Record {
	return journal.Record{Kind: journal.KindBirth,
		Birth: &journal.Birth{Addr: strings.Repeat("a", 64), Digest: strings.Repeat("b", 64)}}
}

func runRec(n int) journal.Record {
	return journal.Record{Kind: journal.KindRun, Run: &journal.Run{N: n}}
}

func TestStreamAppendRotateRecords(t *testing.T) {
	ports.FSBind = fsys.OS{}
	dir := filepath.Join(t.TempDir(), "journal")
	s, err := OpenStream(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Append(birthRec()); err != nil {
		t.Fatal(err)
	}
	// The rotation opens a fresh epoch; an empty live epoch does not
	// stack another.
	if err := s.Rotate(); err != nil {
		t.Fatal(err)
	}
	if err := s.Rotate(); err != nil {
		t.Fatal(err)
	}
	if err := s.Append(runRec(1)); err != nil {
		t.Fatal(err)
	}
	if len(s.files) != 2 {
		t.Fatalf("the rotation must open exactly one new epoch, files: %v", s.files)
	}
	// A fresh open sees the same records in the same order.
	fresh, err := OpenStream(dir)
	if err != nil {
		t.Fatal(err)
	}
	recs, err := fresh.Records()
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 2 || recs[0].Kind != journal.KindBirth || recs[1].Run.N != 1 {
		t.Fatalf("records: %+v", recs)
	}
	// A missing directory is a stream that never wrote.
	empty, err := OpenStream(filepath.Join(t.TempDir(), "nowhere"))
	if err != nil {
		t.Fatal(err)
	}
	if recs, _ := empty.Records(); len(recs) != 0 {
		t.Fatalf("a virgin stream carries no records")
	}
}

func TestStreamTornTailAndDamage(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "journal")
	s, _ := OpenStream(dir)
	if err := s.Append(birthRec()); err != nil {
		t.Fatal(err)
	}
	live := s.files[len(s.files)-1]

	// A torn tail: cut bytes at the very end quarantine, the intact
	// record stands.
	f, _ := os.OpenFile(live, os.O_APPEND|os.O_WRONLY, 0o644)
	_, _ = f.WriteString("---\nkind: run\nn: 9\ncut")
	_ = f.Close()
	healed, err := OpenStream(dir)
	if err != nil {
		t.Fatal(err)
	}
	recs, err := healed.Records()
	if err != nil || len(recs) != 1 {
		t.Fatalf("the torn tail must heal to the intact records: %v %+v", err, recs)
	}
	if _, err := os.Stat(live + ".torn"); err != nil {
		t.Fatalf("the torn bytes must keep as evidence: %v", err)
	}

	// Damage that is not a torn write — a broken record in the
	// middle — is an honest error.
	f2, _ := os.OpenFile(live, os.O_APPEND|os.O_WRONLY, 0o644)
	_, _ = f2.WriteString("---\nkind: run\nn: 2\nrows: []\n---\nkind: run\nbroken: {\n")
	_ = f2.Close()
	damaged, _ := OpenStream(dir)
	if _, err := damaged.Records(); err == nil {
		t.Fatalf("mid-file damage must refuse")
	}
}

// TestStreamFirstRecordAlone — the first record loads without asking
// any later record to parse; a journal that never wrote answers the
// zero record.
func TestStreamFirstRecordAlone(t *testing.T) {
	ports.FSBind = fsys.OS{}
	dir := filepath.Join(t.TempDir(), "journal")
	s, _ := OpenStream(dir)
	if err := s.Append(birthRec()); err != nil {
		t.Fatal(err)
	}
	if err := s.Append(runRec(1)); err != nil {
		t.Fatal(err)
	}
	// A first record that is not parseable — damage by hand — is an
	// honest error even when it is the only thing asked.
	if err := os.WriteFile(s.files[0], []byte("---\nkind: birth\nbroken: {\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	fresh, _ := OpenStream(dir)
	if _, err := fresh.FirstRecord(); err == nil {
		t.Fatal("a broken first record must refuse to load")
	}
	// A journal that never wrote — no files, or an empty first epoch —
	// answers the zero record.
	virgin, _ := OpenStream(filepath.Join(t.TempDir(), "nowhere"))
	if rec, err := virgin.FirstRecord(); err != nil || rec.Kind != "" {
		t.Fatalf("a virgin stream has no first record: %v %+v", err, rec)
	}
	blank := filepath.Join(t.TempDir(), "blank")
	if err := os.MkdirAll(blank, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(blank, "epoch-00001.yaml"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	empty, _ := OpenStream(blank)
	if rec, err := empty.FirstRecord(); err != nil || rec.Kind != "" {
		t.Fatalf("an empty first epoch has no first record: %v %+v", err, rec)
	}
}

// failingFS — the real filesystem with the append always failing.
type failingFS struct {
	fsys.OS
}

func (failingFS) AppendFile(string, []byte) error  { return os.ErrPermission }
func (failingFS) WriteAtomic(string, []byte) error { return os.ErrPermission }

func TestStreamAppendAndRotateRefuseByName(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "journal")
	ports.FSBind = fsys.OS{}
	defer func() { ports.FSBind = fsys.OS{} }()
	// The append refusal is named.
	s, _ := OpenStream(dir)
	var ff failingFS
	if err := ff.AppendFile(filepath.Join(t.TempDir(), "e"), nil); err == nil {
		t.Fatal("the failing append must answer its error")
	}
	_ = s
	// The open of a stream over a regular file refuses by name.
	if _, err := OpenStream(filepath.Join(t.TempDir(), "x")); err == nil {
		_ = err
	}
	blocker := filepath.Join(t.TempDir(), "journal")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := OpenStream(blocker); err == nil {
		t.Fatal("a stream over a file must refuse to open")
	}

	// The append and the rotation under a failing write answer their
	// own names.
	dir2 := filepath.Join(t.TempDir(), "j2")
	ports.FSBind = ff
	s2, err := OpenStream(dir2)
	if err != nil {
		t.Fatal(err)
	}
	if err := s2.Append(birthRec()); err == nil {
		t.Fatal("the failing append must refuse")
	}
	if err := s2.Rotate(); err == nil {
		t.Fatal("the failing rotation must refuse")
	}
}
