package fsport

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

// memFS — the in-memory fake of the port (mock code lives in test
// files only): plain path→content storage with directories implied by
// the paths and per-operation fault injection.
type memFS struct {
	files        map[string][]byte
	nextTmp      int
	mkdirErr     error
	statErr      error
	readErr      error
	readdirErr   error
	writeErr     error
	appendErr    error
	tmpErr       error
	writeFileErr error
	chmodErr     error
	removeErr    error
	walkErr      error
	lockErr      error
	locks        []string
	writes       []string
}

type memInfo struct {
	name string
	dir  bool
}

func (m memInfo) Name() string       { return m.name }
func (m memInfo) Size() int64        { return 0 }
func (m memInfo) Mode() fs.FileMode  { return 0 }
func (m memInfo) ModTime() time.Time { return time.Time{} }
func (m memInfo) IsDir() bool        { return m.dir }
func (m memInfo) Sys() any           { return nil }

type memEntry struct {
	name string
	dir  bool
}

func (e memEntry) Name() string               { return e.name }
func (e memEntry) IsDir() bool                { return e.dir }
func (e memEntry) Type() fs.FileMode          { return 0 }
func (e memEntry) Info() (fs.FileInfo, error) { return memInfo(e), nil }

func notExist(path string) error { return &fs.PathError{Op: "open", Path: path, Err: fs.ErrNotExist} }

func (m *memFS) Stat(path string) (os.FileInfo, error) {
	if m.statErr != nil {
		return nil, m.statErr
	}
	for name := range m.files {
		if name == path {
			return memInfo{m.base(path), false}, nil
		}
	}
	if strings.HasSuffix(path, "/") || m.hasDir(path) {
		return memInfo{m.base(path), true}, nil
	}
	return nil, notExist(path)
}

func (m *memFS) hasDir(dir string) bool {
	for name := range m.files {
		if strings.HasPrefix(name, dir+"/") {
			return true
		}
	}
	return false
}

func (m *memFS) base(path string) string {
	if i := strings.LastIndexByte(path, '/'); i >= 0 {
		return path[i+1:]
	}
	return path
}

func (m *memFS) MkdirAll(string) error { return m.mkdirErr }

func (m *memFS) ReadFile(path string) ([]byte, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	if data, ok := m.files[path]; ok {
		return data, nil
	}
	return nil, notExist(path)
}

func (m *memFS) ReadDir(dir string) ([]os.DirEntry, error) {
	if m.readdirErr != nil {
		return nil, m.readdirErr
	}
	seen := map[string]bool{}
	var out []os.DirEntry
	for name := range m.files {
		if !strings.HasPrefix(name, dir+"/") {
			continue
		}
		rest := name[len(dir)+1:]
		if rest == "" {
			continue
		}
		child, isDir := rest, false
		if i := strings.IndexByte(rest, '/'); i >= 0 {
			child, isDir = rest[:i], true
		}
		if !seen[child] {
			seen[child] = true
			out = append(out, memEntry{child, isDir})
		}
	}
	if len(out) == 0 && !m.hasDir(dir) {
		return nil, notExist(dir)
	}
	return out, nil
}

func (m *memFS) WriteAtomic(path string, data []byte) error {
	if m.writeErr != nil {
		return m.writeErr
	}
	m.writes = append(m.writes, path)
	m.files[path] = data
	return nil
}

func (m *memFS) AppendFile(path string, data []byte) error {
	if m.appendErr != nil {
		return m.appendErr
	}
	m.files[path] = append(m.files[path], data...)
	return nil
}

func (m *memFS) MkdirTemp(dir, pattern string) (string, error) {
	if m.tmpErr != nil {
		return "", m.tmpErr
	}
	m.nextTmp++
	base := dir
	if base == "" {
		base = "/tmp"
	}
	name := strings.ReplaceAll(pattern, "*", "") + itoa(m.nextTmp)
	path := base + "/" + name
	m.files["__dir__"+path] = nil // a marker so the dir exists for Stat
	return path, nil
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

func (m *memFS) WriteFile(path string, data []byte, _ fs.FileMode) error {
	if m.writeFileErr != nil {
		return m.writeFileErr
	}
	m.files[path] = data
	return nil
}

func (m *memFS) Chmod(string, fs.FileMode) error { return m.chmodErr }

func (m *memFS) RemoveAll(path string) error {
	if m.removeErr != nil {
		return m.removeErr
	}
	for name := range m.files {
		if name == path || strings.HasPrefix(name, path+"/") {
			delete(m.files, name)
		}
	}
	return nil
}

func (m *memFS) Walk(root string, fn filepath.WalkFunc) error {
	if m.walkErr != nil {
		return m.walkErr
	}
	if !m.hasDir(root) {
		if _, ok := m.files[root]; !ok {
			return fn(root, memInfo{m.base(root), false}, notExist(root))
		}
	}
	paths := []string{root}
	for name := range m.files {
		if strings.HasPrefix(name, root+"/") && !strings.HasPrefix(name, "__dir__") {
			paths = append(paths, name)
		}
	}
	sort.Strings(paths)
	for _, p := range paths {
		if p != root && strings.Contains(p[len(root)+1:], "/") {
			// intermediate directories materialize on demand below
		}
		isDir := p == root || m.hasDir(p)
		if err := fn(p, memInfo{m.base(p), isDir}, nil); err != nil {
			return err
		}
	}
	return nil
}

func (m *memFS) Remove(path string) error {
	delete(m.files, path)
	return nil
}

func (m *memFS) Lstat(path string) (os.FileInfo, error) { return m.Stat(path) }

func (m *memFS) Readlink(string) (string, error) { return "", notExist("link") }

func (m *memFS) Symlink(target, link string) error { return notExist(link) }

func (m *memFS) TryLockExclusive(string) (func(), bool, error) {
	return func() {}, true, nil
}

func (m *memFS) LockExclusive(path string) (func(), error) {
	if m.lockErr != nil {
		return nil, m.lockErr
	}
	m.locks = append(m.locks, path)
	return func() {}, nil
}

func bindMem(t *testing.T, m *memFS) {
	t.Helper()
	old := Bind
	Bind = m
	t.Cleanup(func() { Bind = old })
}

func TestPortContractOverFake(t *testing.T) {
	m := &memFS{files: map[string][]byte{
		"/w/.punchtape/state.yaml":        []byte("x"),
		"/w/.punchtape/canon/req001.yaml": []byte("x"),
		"/w/.punchtape/canon/scn001.yaml": []byte("x"),
	}}
	bindMem(t, m)

	if _, err := Bind.Stat("/w/.punchtape/state.yaml"); err != nil {
		t.Fatalf("stat of a file: %v", err)
	}
	if fi, err := Bind.Stat("/w/.punchtape/canon"); err != nil || !fi.IsDir() {
		t.Fatalf("stat of an implied dir: %v %+v", err, fi)
	}
	if _, err := Bind.Stat("/w/absent"); !os.IsNotExist(err) {
		t.Fatalf("stat of an absent path must be not-exist: %v", err)
	}

	if data, err := Bind.ReadFile("/w/.punchtape/state.yaml"); err != nil || string(data) != "x" {
		t.Fatalf("read: %q %v", data, err)
	}
	if _, err := Bind.ReadFile("/w/absent"); !os.IsNotExist(err) {
		t.Fatalf("read of an absent path must be not-exist: %v", err)
	}

	entries, err := Bind.ReadDir("/w/.punchtape/canon")
	if err != nil || len(entries) != 2 {
		t.Fatalf("readdir: %v %v", entries, err)
	}
	if _, err := Bind.ReadDir("/w/absent"); !os.IsNotExist(err) {
		t.Fatalf("readdir of an absent dir must be not-exist: %v", err)
	}

	if err := Bind.MkdirAll("/w/anything"); err != nil {
		t.Fatalf("mkdirall over the fake: %v", err)
	}
	if err := Bind.WriteAtomic("/w/.punchtape/new.yaml", []byte("x")); err != nil {
		t.Fatalf("write over the fake: %v", err)
	}
	if string(m.files["/w/.punchtape/new.yaml"]) != "x" {
		t.Fatal("the fake must store what was written")
	}
}

func TestPortFaultInjection(t *testing.T) {
	cases := []struct {
		name string
		prep func(*memFS)
	}{
		{"stat fault", func(m *memFS) { m.statErr = errors.New("io") }},
		{"read fault", func(m *memFS) { m.readErr = errors.New("io") }},
		{"readdir fault", func(m *memFS) { m.readdirErr = errors.New("io") }},
		{"mkdir fault", func(m *memFS) { m.mkdirErr = errors.New("io") }},
		{"write fault", func(m *memFS) { m.writeErr = errors.New("io") }},
		{"append fault", func(m *memFS) { m.appendErr = errors.New("io") }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := &memFS{files: map[string][]byte{
				"/x":   []byte("data"),
				"/d/f": []byte("data"),
			}}
			bindMem(t, m)
			tc.prep(m)
			ops := map[string]func() error{
				"stat":    func() error { _, err := Bind.Stat("/x"); return err },
				"read":    func() error { _, err := Bind.ReadFile("/x"); return err },
				"readdir": func() error { _, err := Bind.ReadDir("/d"); return err },
				"mkdir":   func() error { return Bind.MkdirAll("/x") },
				"write":   func() error { return Bind.WriteAtomic("/x", nil) },
				"append":  func() error { return Bind.AppendFile("/x", nil) },
			}
			for name, op := range ops {
				err := op()
				if tc.name == name+" fault" && err == nil {
					t.Fatalf("%s must surface the injected fault", name)
				}
				if tc.name != name+" fault" && err != nil {
					t.Fatalf("%s must stay healthy under a foreign fault: %v", name, err)
				}
			}
		})
	}
}

// TestOSWiring exercises the production os binding on error paths
// only — empty and type-broken paths fail without creating, writing
// or removing anything real. The success paths of the os binding run
// for real inside the make test gate's instrumented binary (the gate
// performs a real instance round-trip); a unit test must not touch
// real disk.
func TestOSWiring(t *testing.T) {
	var osfs OS
	if _, err := osfs.Stat(""); err == nil {
		t.Fatal("stat of an empty path must fail")
	}
	if _, err := osfs.ReadFile(""); err == nil {
		t.Fatal("read of an empty path must fail")
	}
	// ENOTDIR: a path under a file, not a directory
	if _, err := osfs.ReadDir("/dev/null"); err == nil {
		t.Fatal("readdir of a file must fail")
	}
	if err := osfs.MkdirAll(""); err == nil {
		t.Fatal("mkdirall of an empty path must fail")
	}
	if err := osfs.MkdirAll("/dev/null/x"); err == nil {
		t.Fatal("mkdirall under a file must fail")
	}
	if err := osfs.WriteAtomic("/dev/null/x/f.yaml", []byte("d")); err == nil {
		t.Fatal("a write under a file path must fail at the directory step")
	}
	if err := osfs.WriteAtomic("/proc/fsport-no-write.yaml", []byte("d")); err == nil {
		t.Fatal("a write into the read-only proc must fail at the temp file step")
	}
}
