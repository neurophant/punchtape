package journal

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
	quarantine func(path string) error // optional per-path append fault
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

// entry builds one canonically marshaled record frame.
func entry(key string) (Entry, []byte) {
	e := Entry{
		Transaction:   "tx-" + key,
		SubmissionKey: key,
		DeltaKind:     "intent",
		RecordedAt:    time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		Effects:       []Effect{{Path: "a"}},
	}
	data, _ := yamlio.Marshal(e)
	return e, data
}

const jpath = "/w/.punchtape/journal.yamll"

func TestOpenClasses(t *testing.T) {
	t.Run("missing file is an empty journal", func(t *testing.T) {
		bind(t, &memFS{files: map[string][]byte{}})
		j, err := Open(jpath)
		if err != nil || len(j.All()) != 0 || j.TornTail() != nil {
			t.Fatalf("empty: %v %v %v", err, j.All(), j.TornTail())
		}
	})
	t.Run("reader fault surfaces", func(t *testing.T) {
		m := &memFS{files: map[string][]byte{}, readErr: errIO}
		bind(t, m)
		if _, err := Open(jpath); err == nil {
			t.Fatal("a read failure must surface")
		}
	})
	t.Run("healthy file loads in full", func(t *testing.T) {
		e1, d1 := entry("k1")
		e2, d2 := entry("k2")
		stream := yamlio.Rebuild(nil, [][]byte{d1, d2})
		bind(t, &memFS{files: map[string][]byte{jpath: stream}})
		j, err := Open(jpath)
		if err != nil || len(j.All()) != 2 || j.All()[0].SubmissionKey != e1.SubmissionKey ||
			j.All()[1].SubmissionKey != e2.SubmissionKey {
			t.Fatalf("load: %v %v", err, j.All())
		}
		if j.TornTail() != nil {
			t.Fatal("a healthy file carries no diagnosis")
		}
	})
}

// tornFile builds a stream whose last frame is broken in a named way.
func tornStream(broken []byte) []byte {
	_, d1 := entry("k1")
	return yamlio.Rebuild(nil, [][]byte{d1, broken})
}

func TestTornTailClasses(t *testing.T) {
	_, d1 := entry("k1")

	cases := []struct {
		name   string
		broken []byte
	}{
		{"cut mid-line", []byte(`{"transaction": "tx"`)}, // no trailing newline

		{"not strict yaml", append([]byte("{bad\n"), '\n')},
		{"missing anchors", func() []byte {
			// a strict-decodable record without the transaction anchor
			raw := `{"submission-key":"k","delta-kind":"intent","wall-ms":0,` +
				`"recorded-at":"2000-01-01T00:00:00Z","effects":[]}` + "\n"
			return []byte(raw)
		}()},
		{"not byte-exact on remarshal", func() []byte {
			// the same fields as the canonical record in a different
			// order: decodes strictly, re-marshals differently
			raw := `{"submission-key":"k1","transaction":"tx-k1",` +
				`"delta-kind":"intent","wall-ms":0,` +
				`"recorded-at":"2000-01-01T00:00:00Z","effects":[{"path":"a"}]}` + "\n"
			return []byte(raw)
		}()},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bind(t, &memFS{files: map[string][]byte{jpath: tornStream(tc.broken)}})
			j, err := Open(jpath)
			if err != nil {
				t.Fatalf("open: %v", err)
			}
			if len(j.All()) != 1 {
				t.Fatalf("intact records must load: %d", len(j.All()))
			}
			tt := j.TornTail()
			if tt == nil || tt.Good != 1 || tt.Length <= 0 {
				t.Fatalf("torn diagnosis: %+v", tt)
			}
			_ = d1
		})
	}
}

func TestAppendAndHeal(t *testing.T) {
	t.Run("append writes the frame and keeps the copy", func(t *testing.T) {
		m := &memFS{files: map[string][]byte{}}
		bind(t, m)
		j, _ := Open(jpath)
		e, _ := entry("k1")
		if err := j.Append(e); err != nil {
			t.Fatalf("append: %v", err)
		}
		if len(j.All()) != 1 {
			t.Fatal("the in-memory copy must grow")
		}
		want := string(append([]byte("---\n"), mustMarshal(e)...))
		if string(m.files[jpath]) != want {
			t.Fatalf("file bytes:\n got %q\nwant %q", m.files[jpath], want)
		}
		// reopen: the appended record reads back healthy
		j2, err := Open(jpath)
		if err != nil || len(j2.All()) != 1 || j2.TornTail() != nil {
			t.Fatalf("reopen: %v %v", err, j2.All())
		}
	})
	t.Run("append faults", func(t *testing.T) {
		m := &memFS{files: map[string][]byte{}, mkdirErr: errIO}
		bind(t, m)
		j, _ := Open(jpath)
		e, _ := entry("k1")
		if err := j.Append(e); err == nil || !strings.Contains(err.Error(), "dir") {
			t.Fatalf("a mkdir failure must surface: %v", err)
		}
		m.mkdirErr = nil
		m.appendErr = errIO
		if err := j.Append(e); err == nil {
			t.Fatal("an append failure must surface")
		}
	})
	t.Run("a torn tail heals on the next append", func(t *testing.T) {
		broken := []byte(`{"transaction": "cut`)
		_, intact := entry("k1")
		m := &memFS{files: map[string][]byte{jpath: tornStream(broken)}}
		bind(t, m)
		j, _ := Open(jpath)
		if j.TornTail() == nil {
			t.Fatal("the file must load with a diagnosis")
		}
		e, fresh := entry("k2")
		if err := j.Append(e); err != nil {
			t.Fatalf("append over a torn tail: %v", err)
		}
		if j.TornTail() != nil {
			t.Fatal("the write path must clear the diagnosis")
		}
		sidecar := string(m.files[jpath+".torn"])
		if !strings.Contains(sidecar, "# quarantined") || !strings.Contains(sidecar, `{"transaction": "cut`) {
			t.Fatalf("the sidecar keeps the torn bytes with the diagnosis: %q", sidecar)
		}
		// the rebuilt journal holds the intact record plus the new one
		want := string(yamlio.Rebuild(nil, [][]byte{intact, fresh}))
		if string(m.files[jpath]) != want {
			t.Fatalf("rebuilt bytes:\n got %q\nwant %q", m.files[jpath], want)
		}
	})
	t.Run("heal fault classes", func(t *testing.T) {
		e, _ := entry("k1")
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
				m := &memFS{files: map[string][]byte{jpath: tornStream([]byte("cut"))}}
				bind(t, m)
				j, _ := Open(jpath)
				tc.prep(m)
				if err := j.Append(e); err == nil {
					t.Fatalf("%s must surface", tc.name)
				}
			})
		}
	})
	t.Run("heal of a healthy journal is a no-op", func(t *testing.T) {
		bind(t, &memFS{files: map[string][]byte{}})
		j, _ := Open(jpath)
		if err := j.HealTail(); err != nil {
			t.Fatalf("a healthy journal needs no repair: %v", err)
		}
	})
	t.Run("heal with the diagnosis past the last frame", func(t *testing.T) {
		_, d1 := entry("k1")
		stream := yamlio.Rebuild(nil, [][]byte{d1})
		bind(t, &memFS{files: map[string][]byte{jpath: stream}})
		j, _ := Open(jpath)
		// a diagnosis whose region starts exactly after the last
		// intact frame (the good count equals the frame count)
		j.torn = &yamlio.TornTail{Good: 1}
		if err := j.HealTail(); err != nil {
			t.Fatalf("heal: %v", err)
		}
		if string(mustGetFile(t, jpath)) != string(stream) {
			t.Fatal("every intact frame must survive the rebuild")
		}
	})
}

func TestFinders(t *testing.T) {
	bind(t, &memFS{files: map[string][]byte{}})
	j, _ := Open(jpath)
	e1, _ := entry("k1")
	e1.ContentDigest = "dig1"
	_ = j.Append(e1)
	held := Entry{
		Transaction: "tx-h", SubmissionKey: "k2", DeltaKind: "spec",
		RecordedAt: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		Held:       true, ContentDigest: "dig2",
	}
	_ = j.Append(held)

	if _, ok := j.FindByKey("k1"); !ok {
		t.Fatal("an applied key must be found")
	}
	if _, ok := j.FindByKey("k2"); ok {
		t.Fatal("a held record is not an applied transaction")
	}
	if _, ok := j.FindByKey("zz"); ok {
		t.Fatal("an unknown key must not be found")
	}
	if _, ok := j.FindByContentDigest("dig1"); !ok {
		t.Fatal("an applied digest must be found")
	}
	if _, ok := j.FindByContentDigest("dig2"); ok {
		t.Fatal("a held record's digest is not an applied fact")
	}
	if _, ok := j.FindByContentDigest(""); ok {
		t.Fatal("an empty digest never matches")
	}
}

func mustMarshal(e Entry) []byte {
	data, err := yamlio.Marshal(e)
	if err != nil {
		panic(err)
	}
	return data
}

func mustGetFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := fsport.Bind.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
