package cas

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neurophant/punchtape/internal/infrastructure/fsys"
	"github.com/neurophant/punchtape/internal/ports"
)

// The tests ride the real OS filesystem through the port; the fault
// cases swap a fake that delegates everything but its one knob.
func TestMain(m *testing.M) {
	old := ports.FSBind
	ports.FSBind = fsys.OS{}
	code := m.Run()
	ports.FSBind = old
	os.Exit(code)
}

type faultFS struct {
	fsys.OS
	statErr  error  // every stat fails
	writeErr error  // every atomic write fails
	readErr  error  // every read fails
	readLie  []byte // every read answers these bytes instead
}

func (f faultFS) Stat(p string) (os.FileInfo, error) {
	if f.statErr != nil {
		return nil, f.statErr
	}
	return f.OS.Stat(p)
}

func (f faultFS) WriteAtomic(string, []byte) error { return f.writeErr }

func (f faultFS) ReadFile(p string) ([]byte, error) {
	if f.readErr != nil {
		return nil, f.readErr
	}
	if f.readLie != nil {
		return f.readLie, nil
	}
	return f.OS.ReadFile(p)
}

func bindFault(t *testing.T, f faultFS) {
	t.Helper()
	old := ports.FSBind
	ports.FSBind = f
	t.Cleanup(func() { ports.FSBind = old })
}

// Put then Get returns the very bytes; the address is their sha256;
// Exists agrees.
func TestPutGetRound(t *testing.T) {
	s := New(t.TempDir())
	sizes := []int{0, 1, 1024, 65536}
	for _, n := range sizes {
		data := make([]byte, n)
		for i := range data {
			data[i] = byte(i)
		}
		addr, err := s.Put(data)
		if err != nil {
			t.Fatalf("size %d: %v", n, err)
		}
		back, err := s.Get(addr)
		if err != nil {
			t.Fatalf("size %d: %v", n, err)
		}
		if string(back) != string(data) {
			t.Fatalf("size %d: bytes changed", n)
		}
	}
}

// The same bytes put twice answer the same address and rewrite
// nothing: the object's mtime holds.
func TestPutIdempotent(t *testing.T) {
	s := New(t.TempDir())
	a1, err := s.Put([]byte("stable"))
	if err != nil {
		t.Fatal(err)
	}
	fi1, err := os.Stat(s.path(a1))
	if err != nil {
		t.Fatal(err)
	}
	a2, err := s.Put([]byte("stable"))
	if err != nil {
		t.Fatal(err)
	}
	if a1 != a2 {
		t.Fatalf("addresses differ: %s vs %s", a1, a2)
	}
	fi2, err := os.Stat(s.path(a1))
	if err != nil {
		t.Fatal(err)
	}
	if !fi1.ModTime().Equal(fi2.ModTime()) {
		t.Fatal("a repeat put must not rewrite the object")
	}
}

// Bytes that differ under the same prefix still land at distinct
// addresses: the address is the whole content.
func TestDistinctBytesDistinctAddresses(t *testing.T) {
	s := New(t.TempDir())
	seen := map[string]bool{}
	for i := 0; i < 50; i++ {
		addr, err := s.Put([]byte{byte(i), 0, 0, 0})
		if err != nil {
			t.Fatal(err)
		}
		if seen[addr] {
			t.Fatalf("address collision at %d", i)
		}
		seen[addr] = true
	}
}

// A corrupted object is refused by name: the read bytes do not hash
// to the address.
func TestCorruptObjectRefused(t *testing.T) {
	s := New(t.TempDir())
	addr, err := s.Put([]byte("honest bytes"))
	if err != nil {
		t.Fatal(err)
	}
	p := s.path(addr)
	if err := os.WriteFile(p, []byte("tampered!"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Get(addr); !errors.Is(err, ErrCorruptObject) {
		t.Fatalf("corruption must be named: %v", err)
	}
}

// A lying write (the read-back does not match what was written) is
// caught at Put time — storage honesty.
func TestPutCatchesLyingStorage(t *testing.T) {
	s := New(t.TempDir())
	bindFault(t, faultFS{readLie: []byte("not what was written")})
	if _, err := s.Put([]byte("real bytes")); !errors.Is(err, ErrCorruptObject) {
		t.Fatalf("the lie must surface at write time: %v", err)
	}
}

// A failing disk is an honest error out of Put; a failing read at
// Put's verification and at Get is an honest error too.
func TestFaultsRideOut(t *testing.T) {
	s := New(t.TempDir())
	bindFault(t, faultFS{writeErr: errors.New("disk full")})
	if _, err := s.Put([]byte("x")); err == nil {
		t.Fatal("a full disk must refuse")
	}
	bindFault(t, faultFS{readErr: errors.New("unreadable")})
	if _, err := s.Put([]byte("y")); err == nil {
		t.Fatal("an unreadable verification must refuse")
	}
	if _, err := s.Get("00" + strings.Repeat("0", 62)); err == nil {
		t.Fatal("an unreadable object must refuse")
	}
}

// A valid address with no object under it answers the read's own
// honest error; a read that cannot happen is an honest error too.
func TestGetMissingAndFault(t *testing.T) {
	s := New(t.TempDir())
	if _, err := s.Get(strings.Repeat("a", 64)); !os.IsNotExist(err) {
		t.Fatalf("a missing object must be the read's not-exist: %v", err)
	}
	bindFault(t, faultFS{readErr: errors.New("read fault")})
	if _, err := s.Get(strings.Repeat("a", 64)); err == nil {
		t.Fatal("a failing read must refuse")
	}
}

// An address that is not the store's own form is refused before the
// disk is touched.
func TestBadAddresses(t *testing.T) {
	s := New(t.TempDir())
	for _, addr := range []string{
		"", "00", strings.Repeat("0", 63), strings.Repeat("0", 65),
		strings.Repeat("0", 63) + "A", // uppercase
		strings.Repeat("z", 64),       // not hex
	} {
		if _, err := s.Get(addr); err == nil {
			t.Errorf("get %q: must refuse", addr)
		}
	}
}

// A stray temp file in a shard directory is invisible: readers
// address objects exactly, and a new Put works beside the junk.
func TestStrayTempInvisible(t *testing.T) {
	s := New(t.TempDir())
	addr, err := s.Put([]byte("the object"))
	if err != nil {
		t.Fatal(err)
	}
	junk := filepath.Join(s.root, addr[:2], ".tmp-junk")
	if err := os.WriteFile(junk, []byte("half a write"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Get(addr); err != nil {
		t.Fatalf("a stray temp must not disturb reads: %v", err)
	}
	if _, err := s.Put([]byte("another object")); err != nil {
		t.Fatalf("a stray temp must not disturb writes: %v", err)
	}
}
