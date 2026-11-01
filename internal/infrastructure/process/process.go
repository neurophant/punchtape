//go:build linux || darwin

// Package process — the machine's process technique: the OS
// implementation of the ports.Exec interface. The same discipline as
// fsys: Go cannot intercept a direct exec call from a test, so the
// machine starts every command through the port. One launch law: the
// argv runs as is, in the declared environment, in its own process
// group — and whatever it left alive dies with the group.
package process

import (
	"bytes"
	"errors"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/neurophant/punchtape/internal/ports"
)

// OS — the production binding over the os/exec package.
type OS struct{}

// StartOpts — the one launch: argv as is, working directory, the
// declared environment (nil inherits the caller's).
func (OS) StartOpts(argv []string, dir string, opts ports.LaunchOpts) (ports.Process, error) {
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = dir
	cmd.Env = opts.Env
	// Its own process group: the command's descendants are taken down
	// by the group — they hold no resources after the row.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return &osProcess{cmd: cmd, stdout: &stdout, stderr: &stderr}, nil
}

type osProcess struct {
	cmd    *exec.Cmd
	stdout *bytes.Buffer
	stderr *bytes.Buffer
	once   sync.Once
}

// Wait — the process's own outcome: the exit code, the captured
// streams, the deadline's honest mark. On the deadline path the
// killed command's wait is joined — bounded, as the group's death
// is — before the buffers are read: os/exec's copier goroutines own
// them until cmd.Wait returns.
func (p *osProcess) Wait(timeout time.Duration) ports.Outcome {
	var outcome ports.Outcome
	done := make(chan error, 1)
	go func() { done <- p.cmd.Wait() }()
	select {
	case err := <-done:
		if err != nil {
			var ee *exec.ExitError
			if errors.As(err, &ee) {
				outcome.Exit = ee.ExitCode()
			} else {
				outcome.WaitErr = err.Error()
			}
		}
	case <-time.After(timeout):
		p.KillGroup()
		select {
		case <-done:
		case <-time.After(1 * time.Second):
		}
		outcome.TimedOut = true
	}
	outcome.Stdout = p.stdout.String()
	outcome.Stderr = p.stderr.String()
	return outcome
}

// KillGroup — the whole group down, then the bounded wait for its
// actual death: a kill signal is not death, a dying member still
// holds what it held (a port, a lock).
func (p *osProcess) KillGroup() {
	p.once.Do(func() {
		pgid := p.cmd.Process.Pid
		_ = syscall.Kill(-pgid, syscall.SIGKILL)
		_ = p.cmd.Process.Kill()
		deadline := time.Now().Add(1 * time.Second)
		for {
			if err := syscall.Kill(-pgid, 0); err != nil {
				return // the group is gone
			}
			if time.Now().After(deadline) {
				return // bounded: the reaper has the rest
			}
			time.Sleep(10 * time.Millisecond)
		}
	})
}
