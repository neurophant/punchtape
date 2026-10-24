//go:build linux || darwin

package fsys

import (
	"os/signal"
	"path/filepath"
	"syscall"
	"testing"
)

// TestOSWriteFaultBySizeLimit drives the atomic write's temp-file
// write fault with a real OS fault: RLIMIT_FSIZE zeroed for the
// duration of the one call (SIGXFSZ ignored, both restored before
// the test says anything — under the limit no output is possible).
func TestOSWriteFaultBySizeLimit(t *testing.T) {
	var osfs OS
	var before syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_FSIZE, &before); err != nil {
		t.Skipf("the size limit is not readable: %v", err)
	}
	signal.Ignore(syscall.SIGXFSZ)
	t.Cleanup(func() { signal.Reset(syscall.SIGXFSZ) })
	zero := syscall.Rlimit{Cur: 0, Max: before.Max}
	if err := syscall.Setrlimit(syscall.RLIMIT_FSIZE, &zero); err != nil {
		t.Skipf("the size limit is not settable: %v", err)
	}
	err := osfs.WriteAtomic(filepath.Join(t.TempDir(), "f.yaml"), []byte("d"))
	if rerr := syscall.Setrlimit(syscall.RLIMIT_FSIZE, &before); rerr != nil {
		t.Fatalf("the size limit must be restored: %v", rerr)
	}
	if err == nil {
		t.Fatal("an atomic write under a zero size limit must fail at the temp write step")
	}
}
