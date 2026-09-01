// Run lock: serialization of one instance's check executors. A
// background suite rehearsal and live submission gates run the same
// suite — it is the instance's single resource, and two simultaneous
// executors collide on one run. The lock is held for the duration of
// one run (Run/Capture); the queue lives in the core. The holder
// dies with the process (the kernel releases the flock); there is no
// eternal waiting; check contents are not read — this is the
// executor's resource discipline, not knowledge about the project.
package checks

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/neurophant/punchtape/internal/canon"
)

// acquireRunLock takes the instance's exclusive run lock and returns
// a release function. The wait is bounded above by the duration of
// one run: a live caller (submission gates) marks the background
// rehearsal for stop before taking the lock; the rehearsal exits
// after the current check.
func acquireRunLock(workdir string) (func(), error) {
	lockPath := filepath.Join(canon.Dir(workdir), "cache", "run.lock")
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		return nil, fmt.Errorf("run lock dir: %w", err)
	}
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, fmt.Errorf("run lock: %w", err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		f.Close()
		return nil, fmt.Errorf("run lock: %w", err)
	}
	return func() { _ = f.Close() }, nil
}
