// Package cas — the content-addressed object store: an object's
// name is the sha256 of its bytes, its file is
// <root>/<first two hex>/<remaining 62 hex>. Identical bytes are
// stored once — dedup and idempotence by construction; a repeat Put
// touches nothing and answers the same address. Every write is
// re-read and re-verified, so a lying disk surfaces at write time,
// not a month later at read time; every read verifies the digest
// and answers ErrCorruptObject on mismatch. No compression: the
// objects are kilobyte-sized, stable raw bytes are worth more than
// the bytes saved. The store rides the FS port only.
package cas

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/neurophant/punchtape/internal/ports"
)

// ErrCorruptObject — the bytes under an address do not hash to the
// address: the storage lies, the caller must not trust the object.
var ErrCorruptObject = errors.New("cas: object corrupt")

// Store — the objects tree of one instance under the given root.
type Store struct {
	root string
}

// New — the store over the objects root directory.
func New(root string) *Store { return &Store{root: root} }

func (s *Store) path(addr string) string {
	return filepath.Join(s.root, addr[:2], addr[2:])
}

// Put — the address of data, stored once: the sha256 of the bytes
// in hex. An object that already exists is left untouched; a new
// one lands through a temp file in its own shard and one rename,
// then the written bytes are read back and re-verified.
func (s *Store) Put(data []byte) (string, error) {
	addr := digest(data)
	p := s.path(addr)
	if _, err := ports.FSBind.Stat(p); err == nil {
		return addr, nil
	}
	if err := ports.FSBind.WriteAtomic(p, data); err != nil {
		return "", fmt.Errorf("cas put %s: %w", addr, err)
	}
	back, err := ports.FSBind.ReadFile(p)
	if err != nil {
		return "", fmt.Errorf("cas verify %s: %w", addr, err)
	}
	if digest(back) != addr {
		return "", fmt.Errorf("cas verify %s: %w", addr, ErrCorruptObject)
	}
	return addr, nil
}

// Get — the bytes under an address, verified: the digest must match
// the address. A missing object is the filesystem's own honest
// error, corruption is ErrCorruptObject.
func (s *Store) Get(addr string) ([]byte, error) {
	if err := validAddr(addr); err != nil {
		return nil, err
	}
	data, err := ports.FSBind.ReadFile(s.path(addr))
	if err != nil {
		return nil, err
	}
	if digest(data) != addr {
		return nil, fmt.Errorf("cas get %s: %w", addr, ErrCorruptObject)
	}
	return data, nil
}

// digest — the object address of data: sha256 in hex.
func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// validAddr — the store's own address form: exactly 64 lowercase
// hex digits.
func validAddr(addr string) error {
	if len(addr) != 64 {
		return fmt.Errorf("cas: bad address %q: want 64 hex digits", addr)
	}
	for i := 0; i < len(addr); i++ {
		c := addr[i]
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return fmt.Errorf("cas: bad address %q: want lowercase hex", addr)
		}
	}
	return nil
}
