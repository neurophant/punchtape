// Package fsport — the machine's one filesystem port. Go has no
// external function patching: a direct os call cannot be intercepted
// from a test, so every package that touches the filesystem goes
// through this port. Production binds OS below; unit tests bind an
// in-memory fake defined in their own test files (mock code lives in
// tests only, never here). The port carries no policy — it is the os
// surface the machine uses, nothing more.
package fsport

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// Port — the filesystem operations the machine performs. WriteAtomic
// writes data to a file in one action: first a temp file next to the
// target, then a rename — a reader at any moment sees either the
// whole old content or the new one. AppendFile appends bytes to a
// file in one action, creating it when absent, with an fsync before
// returning.
type Port interface {
	Stat(path string) (os.FileInfo, error)
	MkdirAll(dir string) error
	ReadFile(path string) ([]byte, error)
	ReadDir(dir string) ([]os.DirEntry, error)
	WriteAtomic(path string, data []byte) error
	AppendFile(path string, data []byte) error
	Remove(path string) error
	Lstat(path string) (os.FileInfo, error)
	MkdirTemp(dir, pattern string) (string, error)
	WriteFile(path string, data []byte, perm os.FileMode) error
	Chmod(path string, perm os.FileMode) error
	RemoveAll(path string) error
	Walk(root string, fn filepath.WalkFunc) error
	// LockExclusive takes an exclusive advisory lock on a file and
	// returns its release. The holder dies with the process (the
	// kernel releases the lock); used to serialize one instance's
	// executors.
	LockExclusive(path string) (release func(), err error)
	// TryLockExclusive is the non-blocking form: held=false when the
	// lock is busy.
	TryLockExclusive(path string) (release func(), held bool, err error)
	// Readlink resolves a symbolic link (checkpoints record link
	// targets as observable facts).
	Readlink(path string) (string, error)
	// Symlink creates a symbolic link (a checkpoint copy reproduces
	// recorded links).
	Symlink(target, link string) error
}

// OS — the production binding over the os package.
type OS struct{}

func (OS) Stat(path string) (os.FileInfo, error)     { return os.Stat(path) }
func (OS) MkdirAll(dir string) error                 { return os.MkdirAll(dir, 0o755) }
func (OS) ReadFile(path string) ([]byte, error)      { return os.ReadFile(path) }
func (OS) ReadDir(dir string) ([]os.DirEntry, error) { return os.ReadDir(dir) }

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

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("temp file write: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("temp file sync: %w", err)
	}
	// close after a successful fsync cannot report new failures
	_ = tmp.Close()
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("rename to %s: %w", path, err)
	}
	return nil
}

func (OS) AppendFile(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	// close after a successful fsync cannot report new failures
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	_ = f.Close()
	return nil
}

func (OS) Remove(path string) error               { return os.Remove(path) }
func (OS) Lstat(path string) (os.FileInfo, error) { return os.Lstat(path) }

func (OS) MkdirTemp(dir, pattern string) (string, error) {
	return os.MkdirTemp(dir, pattern)
}

func (OS) WriteFile(path string, data []byte, perm os.FileMode) error {
	return os.WriteFile(path, data, perm)
}

func (OS) Chmod(path string, perm os.FileMode) error {
	return os.Chmod(path, perm)
}

func (OS) RemoveAll(path string) error { return os.RemoveAll(path) }

func (OS) Walk(root string, fn filepath.WalkFunc) error {
	return filepath.Walk(root, fn)
}

func (OS) LockExclusive(path string) (func(), error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		f.Close()
		return nil, err
	}
	return func() { _ = f.Close() }, nil
}

func (OS) Readlink(path string) (string, error) { return os.Readlink(path) }
func (OS) Symlink(target, link string) error    { return os.Symlink(target, link) }

func (OS) TryLockExclusive(path string) (func(), bool, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, false, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		return nil, false, nil
	}
	return func() { _ = f.Close() }, true, nil
}

// Bind — the active port binding (the http.DefaultClient pattern):
// production never touches it, tests swap it per-case and restore it
// on cleanup.
var Bind Port = OS{}
