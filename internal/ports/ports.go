// Package ports — the machine's ports: the contracts the
// infrastructure layer implements and the application layer consumes.
// A leaf: every layer below and above rides on it, nothing below it.
package ports

import (
	"os"
	"path/filepath"
	"time"
)

// FS — the filesystem port: every file operation of the machine.
// WriteAtomic writes data to a file in one action: a temp file next
// to the target, then a rename — a reader at any moment sees either
// the whole old content or the new one. AppendFile appends bytes in
// one action, creating the file when absent, with an fsync before
// returning.
type FS interface {
	Stat(path string) (os.FileInfo, error)
	MkdirAll(dir string) error
	// Mkdir creates one directory: EEXIST answers the caller, so a
	// birth's atomicity is the mkdir itself — exactly one winner.
	Mkdir(dir string) error
	ReadFile(path string) ([]byte, error)
	ReadDir(dir string) ([]os.DirEntry, error)
	WriteAtomic(path string, data []byte) error
	AppendFile(path string, data []byte) error
	Remove(path string) error
	RemoveAll(path string) error
	Walk(root string, fn filepath.WalkFunc) error
}

// Exec — the process port: the build's and the rows' launches. The
// argv runs AS IS — the environment is a property of the launch,
// never a wrapper constructed around the command.
type Exec interface {
	StartOpts(argv []string, dir string, opts LaunchOpts) (Process, error)
}

// LaunchOpts — the launch's one discipline: the command sees exactly
// this declared environment and nothing else (the clean environment
// law); nil inherits the caller's environment.
type LaunchOpts struct {
	Env []string
}

// Outcome — what a finished process left behind. Exit is the exit
// code (0 on success); WaitErr carries a non-exit wait failure (the
// process could not be reaped normally); TimedOut marks a run cut by
// its deadline.
type Outcome struct {
	Exit     int
	Stdout   string
	Stderr   string
	WaitErr  string
	TimedOut bool
}

// Process — one started process: it runs in its own process group so
// its descendants (servers, workers, what a shell raised) are taken
// down together with it.
type Process interface {
	// Wait blocks until the process exits or the timeout passes.
	Wait(timeout time.Duration) Outcome
	// KillGroup removes the whole process group and waits for its
	// actual death: a kill signal is not death, a dying member still
	// holds what it held (a port, a lock). The wait is bounded.
	KillGroup()
}

// FSBind — the active filesystem binding (the http.DefaultClient
// pattern): production never touches it, tests swap it per-case and
// restore it on cleanup. The infrastructure layer wires it at boot.
var FSBind FS = unwiredFS{}

// ExecBind — the active process binding, the same discipline as
// FSBind.
var ExecBind Exec = unwiredExec{}
