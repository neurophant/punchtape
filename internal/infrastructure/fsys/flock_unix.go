//go:build linux || darwin

package fsys

import (
	"fmt"
	"os"
	"syscall"
)

// Flock — an exclusive lock on the file at the path, created when
// absent: the call blocks until the lock is held, the release
// closure drops it by closing the descriptor — the kernel drops it
// just the same when the holder dies. Local filesystems only: the
// network ones carry no such promise. The open's own error is the
// take's only refusal: a blocking flock over a successfully opened
// regular file carries no failure on the supported platforms —
// measured, the kill of a group rides the same discipline.
func (OS) Flock(path string) (func() error, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	_ = syscall.Flock(int(f.Fd()), syscall.LOCK_EX)
	return func() error { return f.Close() }, nil
}
