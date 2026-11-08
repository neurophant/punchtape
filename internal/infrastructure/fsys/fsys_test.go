package fsys

import "github.com/neurophant/punchtape/internal/ports"

import (
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
	files      map[string][]byte
	mkdirErr   error
	statErr    error
	readErr    error
	readdirErr error
	writeErr   error
	appendErr  error
	removeErr  error
	walkErr    error
	writes     []string
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
func (m *memFS) Mkdir(string) error    { return m.mkdirErr }

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

func bindMem(t *testing.T, m *memFS) {
	t.Helper()
	old := ports.FSBind
	ports.FSBind = m
	t.Cleanup(func() { ports.FSBind = old })
}

func TestPortContractOverFake(t *testing.T) {
	m := &memFS{files: map[string][]byte{
		"/w/.punchtape/state.yaml":        []byte("x"),
		"/w/.punchtape/canon/req001.yaml": []byte("x"),
		"/w/.punchtape/canon/scn001.yaml": []byte("x"),
	}}
	bindMem(t, m)

	if _, err := ports.FSBind.Stat("/w/.punchtape/state.yaml"); err != nil {
		t.Fatalf("stat of a file: %v", err)
	}
	if fi, err := ports.FSBind.Stat("/w/.punchtape/canon"); err != nil || !fi.IsDir() {
		t.Fatalf("stat of an implied dir: %v %+v", err, fi)
	}
	if _, err := ports.FSBind.Stat("/w/absent"); !os.IsNotExist(err) {
		t.Fatalf("stat of an absent path must be not-exist: %v", err)
	}

	if data, err := ports.FSBind.ReadFile("/w/.punchtape/state.yaml"); err != nil || string(data) != "x" {
		t.Fatalf("read: %q %v", data, err)
	}
	if _, err := ports.FSBind.ReadFile("/w/absent"); !os.IsNotExist(err) {
		t.Fatalf("read of an absent path must be not-exist: %v", err)
	}

	entries, err := ports.FSBind.ReadDir("/w/.punchtape/canon")
	if err != nil || len(entries) != 2 {
		t.Fatalf("readdir: %v %v", entries, err)
	}
	if _, err := ports.FSBind.ReadDir("/w/absent"); !os.IsNotExist(err) {
		t.Fatalf("readdir of an absent dir must be not-exist: %v", err)
	}

	if err := ports.FSBind.MkdirAll("/w/anything"); err != nil {
		t.Fatalf("mkdirall over the fake: %v", err)
	}
	if err := ports.FSBind.WriteAtomic("/w/.punchtape/new.yaml", []byte("x")); err != nil {
		t.Fatalf("write over the fake: %v", err)
	}
	if string(m.files["/w/.punchtape/new.yaml"]) != "x" {
		t.Fatal("the fake must store what was written")
	}
}

// TestOSWiring exercises the production os binding on error paths
// only, without creating, writing, or removing anything real.
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
	ro := filepath.Join(t.TempDir(), "read-only")
	if err := os.Mkdir(ro, 0o500); err != nil {
		t.Fatal(err)
	}
	// EACCES: the temp file step under a read-only directory
	if err := osfs.WriteAtomic(filepath.Join(ro, "no-write.yaml"), []byte("d")); err == nil {
		t.Fatal("a write into a read-only directory must fail at the temp file step")
	}
	if err := osfs.AppendFile("/proc/fsys-no-append.log", []byte("d")); err == nil {
		t.Fatal("an append into the read-only proc must fail at the open step")
	}
}

// TestOSRealTree drives the production binding over a real scratch
// tree: the atomic write's rename refusal onto an occupied directory
// and the recursive remove of a grown tree.
func TestOSRealTree(t *testing.T) {
	var osfs OS
	root := t.TempDir()

	if err := osfs.WriteAtomic(filepath.Join(root, "note.txt"), []byte("body")); err != nil {
		t.Fatalf("atomic write: %v", err)
	}
	occupied := filepath.Join(root, "occupied")
	if err := osfs.MkdirAll(occupied); err != nil {
		t.Fatal(err)
	}
	if err := osfs.WriteAtomic(occupied, []byte("body")); err == nil {
		t.Fatal("an atomic write onto an existing directory must fail at the rename step")
	}

	deep := filepath.Join(root, "a", "b", "c")
	if err := osfs.MkdirAll(deep); err != nil {
		t.Fatal(err)
	}
	if err := osfs.Remove(filepath.Join(root, "note.txt")); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if err := osfs.RemoveAll(filepath.Join(root, "a")); err != nil {
		t.Fatalf("removeall: %v", err)
	}
	if _, err := osfs.Stat(deep); !os.IsNotExist(err) {
		t.Fatalf("the removed tree must be gone: %v", err)
	}
}

// TestOSDeviceFaults drives the production binding against real
// device semantics: /dev/full takes every byte and stays empty
// (the write fault), /dev/null accepts bytes and cannot fsync (the
// sync fault after a successful write).
func TestOSDeviceFaults(t *testing.T) {
	var osfs OS
	if err := osfs.AppendFile("/dev/full", []byte("d")); err == nil {
		t.Fatal("an append into the full device must fail at the write step")
	}
	if err := osfs.AppendFile("/dev/null", []byte("d")); err == nil {
		t.Fatal("an append into the null device must fail at the sync step: the device takes bytes but cannot fsync")
	}
}

// Flock — the memory fake's lock: one holder remembered, the
// release a no-op — the fake never runs two holders at once.
func (m *memFS) Flock(path string) (func() error, error) {
	if m.mkdirErr != nil {
		return nil, m.mkdirErr
	}
	m.writes = append(m.writes, path)
	return func() error { return nil }, nil
}

// TestFlockHoldsAndReleases — the production lock answers only one
// holder at a time and frees on release: the second take waits
// behind the first, stands first in line after it.
func TestFlockHoldsAndReleases(t *testing.T) {
	path := filepath.Join(t.TempDir(), "run.lock")
	var osfs OS
	release, err := osfs.Flock(path)
	if err != nil {
		t.Fatalf("the first take must hold: %v", err)
	}
	held := make(chan error, 1)
	go func() {
		second, serr := osfs.Flock(path)
		if serr != nil {
			held <- serr
			return
		}
		held <- second()
	}()
	select {
	case err := <-held:
		t.Fatalf("the second take must wait behind the first: %v", err)
	case <-time.After(200 * time.Millisecond):
	}
	if err := release(); err != nil {
		t.Fatalf("the release must land: %v", err)
	}
	select {
	case err := <-held:
		if err != nil {
			t.Fatalf("the waiting take must hold after the release: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("the waiting take must hold after the release")
	}
	// A lock over an unopenable path answers its own error.
	if _, err := osfs.Flock(filepath.Join(t.TempDir(), "gone", "x", "run.lock")); err == nil {
		t.Fatal("an unopenable lock path must answer its error")
	}
}
