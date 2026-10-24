// Package fsys — the machine's filesystem technique: the OS
// implementation of the application layer's FS port. Go has no
// external function patching: a direct os call cannot be intercepted
// from a test, so every package that touches the filesystem goes
// through the port. The port carries no policy — it is the os surface
// the machine uses, nothing more.
package fsys

import (
	"fmt"
	"os"
	"path/filepath"
)

// OS — the production binding over the os package.
type OS struct{}

// Stat — the os stat.
func (OS) Stat(path string) (os.FileInfo, error) { return os.Stat(path) }

// MkdirAll — the os mkdir -p.
func (OS) MkdirAll(dir string) error { return os.MkdirAll(dir, 0o755) }

// Mkdir — the os mkdir; EEXIST answers the caller.
func (OS) Mkdir(dir string) error { return os.Mkdir(dir, 0o755) }

// ReadFile — the os read.
func (OS) ReadFile(path string) ([]byte, error) { return os.ReadFile(path) }

// ReadDir — the os listing.
func (OS) ReadDir(dir string) ([]os.DirEntry, error) { return os.ReadDir(dir) }

// WriteAtomic — one whole-file write: temp file, sync, rename.
func (OS) WriteAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("dir %s: %w", dir, err)
	}
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // cleanup when the rename did not happen

	if err := writeSyncClose(tmp, data); err != nil {
		return fmt.Errorf("temp file: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("rename to %s: %w", path, err)
	}
	return nil
}

// AppendFile — one append action with an fsync.
func (OS) AppendFile(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	return writeSyncClose(f, data)
}

// writeSyncClose — the one file discipline both writers share: the
// bytes land, the sync follows, the close after a successful sync
// cannot report new failures.
func writeSyncClose(f *os.File, data []byte) error {
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	_ = f.Close()
	return nil
}

// Remove — the os remove.
func (OS) Remove(path string) error { return os.Remove(path) }

// RemoveAll — the os recursive remove.
func (OS) RemoveAll(path string) error { return os.RemoveAll(path) }

// Walk — the os walk of the tree under the root.
func (OS) Walk(root string, fn filepath.WalkFunc) error {
	return filepath.Walk(root, fn)
}
