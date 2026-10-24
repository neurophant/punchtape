// The unwired bindings' battery: a port reached before the
// infrastructure is mounted is a programming error — every method
// panics with its port's own name, never a silent default.
package ports

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func mustPanic(t *testing.T, name string, call func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatalf("%s must panic unwired", name)
		}
	}()
	call()
}

func TestUnwiredFSPanics(t *testing.T) {
	var fs unwiredFS
	mustPanic(t, "Stat", func() { _, _ = fs.Stat("x") })
	mustPanic(t, "MkdirAll", func() { _ = fs.MkdirAll("x") })
	mustPanic(t, "Mkdir", func() { _ = fs.Mkdir("x") })
	mustPanic(t, "ReadFile", func() { _, _ = fs.ReadFile("x") })
	mustPanic(t, "ReadDir", func() { _, _ = fs.ReadDir("x") })
	mustPanic(t, "WriteAtomic", func() { _ = fs.WriteAtomic("x", nil) })
	mustPanic(t, "AppendFile", func() { _ = fs.AppendFile("x", nil) })
	mustPanic(t, "Remove", func() { _ = fs.Remove("x") })
	mustPanic(t, "RemoveAll", func() { _ = fs.RemoveAll("x") })
	mustPanic(t, "Walk", func() { _ = fs.Walk("x", func(string, os.FileInfo, error) error { return nil }) })
}

func TestUnwiredExecPanics(t *testing.T) {
	var ex unwiredExec
	mustPanic(t, "StartOpts", func() { _, _ = ex.StartOpts(nil, ".", LaunchOpts{}) })
}

// The port types' contracts a compiler must keep: the OS binding
// satisfies both ports whole.
var (
	_ FS      = unwiredFS{}
	_ Exec    = unwiredExec{}
	_ Process = (*osProcessStub)(nil)
)

type osProcessStub struct{}

func (*osProcessStub) Wait(time.Duration) Outcome { return Outcome{} }
func (*osProcessStub) KillGroup()                 {}

// The default binds are the unwired ones until the composition root
// mounts the production values.
func TestDefaultBindsUnwired(t *testing.T) {
	if _, ok := FSBind.(unwiredFS); !ok {
		t.Fatalf("the default filesystem bind must be unwired")
	}
	if _, ok := ExecBind.(unwiredExec); !ok {
		t.Fatalf("the default process bind must be unwired")
	}
	_ = filepath.Separator
}
