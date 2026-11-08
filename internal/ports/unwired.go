package ports

import (
	"os"
	"path/filepath"
)

// The unwired port bindings: a scenario reached a port before the
// infrastructure layer was mounted — a programming error of the
// assembly, never a silent default. Every method panics with the
// port's own name.
type unwiredFS struct{}

func (unwiredFS) Stat(string) (os.FileInfo, error) { panic("ports: filesystem port not wired") }
func (unwiredFS) MkdirAll(string) error            { panic("ports: filesystem port not wired") }
func (unwiredFS) Mkdir(string) error               { panic("ports: filesystem port not wired") }
func (unwiredFS) ReadFile(string) ([]byte, error)  { panic("ports: filesystem port not wired") }
func (unwiredFS) ReadDir(string) ([]os.DirEntry, error) {
	panic("ports: filesystem port not wired")
}
func (unwiredFS) WriteAtomic(string, []byte) error { panic("ports: filesystem port not wired") }
func (unwiredFS) AppendFile(string, []byte) error  { panic("ports: filesystem port not wired") }
func (unwiredFS) Remove(string) error              { panic("ports: filesystem port not wired") }
func (unwiredFS) RemoveAll(string) error           { panic("ports: filesystem port not wired") }
func (unwiredFS) Walk(string, filepath.WalkFunc) error {
	panic("ports: filesystem port not wired")
}
func (unwiredFS) Flock(string) (func() error, error) {
	panic("ports: filesystem port not wired")
}

type unwiredExec struct{}

func (unwiredExec) StartOpts([]string, string, LaunchOpts) (Process, error) {
	panic("ports: process port not wired")
}
