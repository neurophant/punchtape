package ledger

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/neurophant/punchtape/internal/fsport"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// memFS — the in-memory fake of the filesystem port (mock code lives
// in this test file only).
type memFS struct {
	files      map[string][]byte
	readErr    error
	mkdirErr   error
	writeErr   error
	appendErr  error
	quarantine func(path string) error
}

var errIO = errors.New("io failure")

func notExist(path string) error {
	return &fs.PathError{Op: "open", Path: path, Err: fs.ErrNotExist}
}

func (m *memFS) Stat(string) (os.FileInfo, error)      { return nil, notExist("") }
func (m *memFS) MkdirAll(string) error                 { return m.mkdirErr }
func (m *memFS) ReadDir(string) ([]os.DirEntry, error) { return nil, notExist("") }

func (m *memFS) ReadFile(path string) ([]byte, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	if data, ok := m.files[path]; ok {
		return data, nil
	}
	return nil, notExist(path)
}

func (m *memFS) WriteAtomic(path string, data []byte) error {
	if m.writeErr != nil {
		return m.writeErr
	}
	m.files[path] = data
	return nil
}

func (m *memFS) AppendFile(path string, data []byte) error {
	if m.appendErr != nil {
		return m.appendErr
	}
	if m.quarantine != nil {
		if err := m.quarantine(path); err != nil {
			return err
		}
	}
	m.files[path] = append(m.files[path], data...)
	return nil
}

func (m *memFS) MkdirTemp(dir, pattern string) (string, error) { return "/tmp/mem", nil }
func (m *memFS) WriteFile(path string, data []byte, _ fs.FileMode) error {
	m.files[path] = data
	return nil
}
func (m *memFS) Chmod(string, fs.FileMode) error { return nil }
func (m *memFS) RemoveAll(path string) error {
	for name := range m.files {
		if name == path || strings.HasPrefix(name, path+"/") {
			delete(m.files, name)
		}
	}
	return nil
}
func (m *memFS) Walk(string, filepath.WalkFunc) error { return nil }
func (m *memFS) LockExclusive(string) (func(), error) { return func() {}, nil }

func (m *memFS) Remove(path string) error {
	delete(m.files, path)
	return nil
}
func (m *memFS) Lstat(path string) (os.FileInfo, error) { return m.Stat(path) }
func (m *memFS) Readlink(string) (string, error)        { return "", notExist("link") }
func (m *memFS) Symlink(target, link string) error      { return notExist(link) }

func (m *memFS) TryLockExclusive(string) (func(), bool, error) { return func() {}, true, nil }

func bind(t *testing.T, m *memFS) {
	t.Helper()
	old := fsport.Bind
	fsport.Bind = m
	t.Cleanup(func() { fsport.Bind = old })
}

const lpath = "/w/.punchtape/ledger.yamll"

func at() time.Time {
	return time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
}

// record builds one canonically marshaled record frame.
func record(event string, wall int64, tokens *int64) (Entry, []byte) {
	e := Entry{
		Event: event, Stage: "implement", WallMs: wall, At: at(),
		Details: map[string]string{},
	}
	data, _ := yamlio.Marshal(e)
	return e, data
}

func TestNewEntryStopwatch(t *testing.T) {
	e := NewEntry("next", "spec")
	if e.Event != "next" || e.At.IsZero() || e.Start != e.At {
		t.Fatalf("the record must fix its time and start: %+v", e)
	}
	if e.Duration() < 0 {
		t.Fatal("duration cannot be negative")
	}
}

func TestOpenClasses(t *testing.T) {
	t.Run("missing file is an empty ledger", func(t *testing.T) {
		bind(t, &memFS{files: map[string][]byte{}})
		l, err := Open(lpath)
		if err != nil || len(l.All()) != 0 || l.TornTail() != nil {
			t.Fatalf("empty: %v %v", err, l.All())
		}
	})
	t.Run("reader fault surfaces", func(t *testing.T) {
		bind(t, &memFS{files: map[string][]byte{}, readErr: errIO})
		if _, err := Open(lpath); err == nil {
			t.Fatal("a read failure must surface")
		}
	})
	t.Run("healthy file loads in full", func(t *testing.T) {
		_, d1 := record("next", 5, nil)
		_, d2 := record("submit", 7, nil)
		stream := yamlio.Rebuild(nil, [][]byte{d1, d2})
		bind(t, &memFS{files: map[string][]byte{lpath: stream}})
		l, err := Open(lpath)
		if err != nil || len(l.All()) != 2 || l.All()[0].Event != "next" {
			t.Fatalf("load: %v %v", err, l.All())
		}
	})
	t.Run("torn classes carry a diagnosis", func(t *testing.T) {
		cases := []struct {
			name   string
			broken []byte
		}{
			{"cut mid-line", []byte(`{"event": "ne`)},
			{"not strict yaml", []byte("{bad\n")},
			{"missing anchors", []byte(`{"stage":"x","wall-ms":1,` +
				`"details":{},"recorded-at":"2000-01-01T00:00:00Z"}` + "\n")},
			{"not byte-exact on remarshal", []byte(`{"stage":"implement","event":"next",` +
				`"wall-ms":5,"calls":0,"details":{},` +
				`"recorded-at":"2000-01-01T00:00:00Z"}` + "\n")},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				_, d1 := record("next", 5, nil)
				stream := yamlio.Rebuild(nil, [][]byte{d1, tc.broken})
				bind(t, &memFS{files: map[string][]byte{lpath: stream}})
				l, err := Open(lpath)
				if err != nil {
					t.Fatalf("open: %v", err)
				}
				if len(l.All()) != 1 || l.TornTail() == nil || l.TornTail().Good != 1 {
					t.Fatalf("diagnosis: %d %+v", len(l.All()), l.TornTail())
				}
			})
		}
	})
}

func TestAppendAndRepair(t *testing.T) {
	t.Run("append fills the times and writes the frame", func(t *testing.T) {
		m := &memFS{files: map[string][]byte{}}
		bind(t, m)
		l, _ := Open(lpath)
		e := NewEntry("submit", "implement")
		if err := l.Append(e); err != nil {
			t.Fatalf("append: %v", err)
		}
		got := l.All()[0]
		if got.At.IsZero() {
			t.Fatal("a zero record time must be filled")
		}
		want := string(append([]byte("---\n"), mustMarshal(got)...))
		if string(m.files[lpath]) != want {
			t.Fatalf("file bytes:\n got %q\nwant %q", m.files[lpath], want)
		}
		// a bare record (no stopwatch) fills its time and keeps a
		// zero wall time
		if err := l.Append(Entry{Event: "bare", Stage: "spec", Details: map[string]string{}}); err != nil {
			t.Fatalf("bare append: %v", err)
		}
		if back := l.All()[1]; back.At.IsZero() || back.WallMs != 0 {
			t.Fatalf("bare record: %+v", back)
		}
	})
	t.Run("append fault classes", func(t *testing.T) {
		e := NewEntry("submit", "implement")
		m := &memFS{files: map[string][]byte{}, mkdirErr: errIO}
		bind(t, m)
		l, _ := Open(lpath)
		if err := l.Append(e); err == nil || !strings.Contains(err.Error(), "dir") {
			t.Fatalf("a mkdir failure must surface: %v", err)
		}
		m.mkdirErr = nil
		m.appendErr = errIO
		if err := l.Append(e); err == nil {
			t.Fatal("an append failure must surface")
		}
	})
	t.Run("a torn tail heals on the next append", func(t *testing.T) {
		_, intact := record("next", 5, nil)
		broken := []byte(`{"event": "cut`)
		stream := yamlio.Rebuild(nil, [][]byte{intact, broken})
		m := &memFS{files: map[string][]byte{lpath: stream}}
		bind(t, m)
		l, _ := Open(lpath)
		if l.TornTail() == nil {
			t.Fatal("the file must load with a diagnosis")
		}
		e, fresh := record("submit", 9, nil)
		if err := l.Append(e); err != nil {
			t.Fatalf("append over a torn tail: %v", err)
		}
		if l.TornTail() != nil {
			t.Fatal("the write path must clear the diagnosis")
		}
		sidecar := string(m.files[lpath+".torn"])
		if !strings.Contains(sidecar, "# quarantined") || !strings.Contains(sidecar, `{"event": "cut`) {
			t.Fatalf("sidecar: %q", sidecar)
		}
		want := string(yamlio.Rebuild(nil, [][]byte{intact, fresh}))
		if string(m.files[lpath]) != want {
			t.Fatalf("rebuilt bytes:\n got %q\nwant %q", m.files[lpath], want)
		}
	})
	t.Run("repair fault classes", func(t *testing.T) {
		e := NewEntry("submit", "implement")
		faults := []struct {
			name string
			prep func(*memFS)
		}{
			{"repair read", func(m *memFS) { m.readErr = errIO }},
			{"quarantine write", func(m *memFS) {
				m.quarantine = func(p string) error {
					if strings.HasSuffix(p, ".torn") {
						return errIO
					}
					return nil
				}
			}},
			{"repair write", func(m *memFS) { m.writeErr = errIO }},
		}
		for _, tc := range faults {
			t.Run(tc.name, func(t *testing.T) {
				_, intact := record("next", 5, nil)
				stream := yamlio.Rebuild(nil, [][]byte{intact, []byte("cut")})
				m := &memFS{files: map[string][]byte{lpath: stream}}
				bind(t, m)
				l, _ := Open(lpath)
				tc.prep(m)
				if err := l.Append(e); err == nil {
					t.Fatalf("%s must surface", tc.name)
				}
			})
		}
	})
	t.Run("repair with the good count at the frame count", func(t *testing.T) {
		_, intact := record("next", 5, nil)
		stream := yamlio.Rebuild(nil, [][]byte{intact})
		bind(t, &memFS{files: map[string][]byte{lpath: stream}})
		l, _ := Open(lpath)
		l.torn = &yamlio.TornTail{Good: 1}
		if err := l.repair(); err != nil {
			t.Fatalf("repair: %v", err)
		}
		data, _ := fsport.Bind.ReadFile(lpath)
		if string(data) != string(stream) {
			t.Fatal("every intact frame must survive the rebuild")
		}
	})
}

func TestMetrics(t *testing.T) {
	bind(t, &memFS{files: map[string][]byte{}})
	l, _ := Open(lpath)
	mk := func(event, stage string, wall int64) {
		e := Entry{Event: event, Stage: stage, WallMs: wall,
			At: at(), Details: map[string]string{}}
		if err := l.Append(e); err != nil {
			t.Fatal(err)
		}
	}
	mk("next", "spec", 10)
	mk("next", "implement", 30)
	mk("submit", "implement", 50)

	if got := l.ByEvent("next"); len(got) != 2 {
		t.Fatalf("by event: %d", len(got))
	}
	if got := l.CountBy("next"); got != 2 {
		t.Fatalf("count: %d", got)
	}
	if got := l.CountBy("zz"); got != 0 {
		t.Fatalf("count of an unknown event: %d", got)
	}
	walls := l.WallMsByStage()
	if walls["implement"] != 80 || walls["spec"] != 10 {
		t.Fatalf("walls by stage: %v", walls)
	}
	// nearest-rank p95: for n=2 the rank is 2 (the max)
	if got := l.P95WallMs("next"); got != 30 {
		t.Fatalf("p95 of two: %d", got)
	}
	if got := l.P95WallMs("submit"); got != 50 {
		t.Fatalf("p95 of one: %d", got)
	}
	if got := l.P95WallMs("zz"); got != 0 {
		t.Fatalf("p95 of none: %d", got)
	}
	// a fuller rank check: 20 records, p95 = the 19th by size
	for i := 0; i < 20; i++ {
		mk("bulk", "bulk", int64(i+1))
	}
	if got := l.P95WallMs("bulk"); got != 19 {
		t.Fatalf("p95 nearest-rank of twenty: %d", got)
	}
}

func mustMarshal(e Entry) []byte {
	data, err := yamlio.Marshal(e)
	if err != nil {
		panic(err)
	}
	return data
}
