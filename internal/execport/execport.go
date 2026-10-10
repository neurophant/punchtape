// Package execport — the machine's one process port. The same
// discipline as fsport: Go cannot intercept a direct exec call from a
// test, so every package that starts processes goes through this
// port. Production binds OS below; unit tests bind in-memory fakes
// defined in their own test files (mock code lives in tests only,
// never here).
package execport

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

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

// Port — the process operations the machine performs.
type Port interface {
	LookPath(file string) (string, error)
	Start(argv []string, dir string, stdin *string) (Process, error)
	// StartOpts is Start with launch disciplines (LaunchOpts): the
	// argv runs AS IS — a discipline is a property of the launch,
	// never a wrapper constructed around the command.
	StartOpts(argv []string, dir string, stdin *string, opts LaunchOpts) (Process, error)
	// StartCtx is Start with a context: cancellation takes the
	// process down (builds and rows run under deadlines).
	StartCtx(ctx context.Context, argv []string, dir string, stdin *string) (Process, error)
	// SpawnDetached starts a process in its own session, decoupled
	// from this one: the parent does not wait, the child outlives it.
	SpawnDetached(argv []string, dir string) error
	// Combined runs argv under a deadline and returns the exit code
	// and the combined output (stdout and stderr in one stream): the
	// row surfaces. err is a start or reap failure
	// only; a non-zero exit is a fact, not an error.
	Combined(ctx context.Context, argv []string, dir string, timeout time.Duration) (exit int, output string, timedOut bool, err error)
}

// LaunchOpts — the launch disciplines a caller can order beyond the
// standing ones (the process group). Netns — the launch happens
// inside a fresh network-less namespace: the command's own bytes are
// untouched, the network simply is not there; a host that forbids
// the namespace answers an honest start failure. The discipline is
// the executor of the grammar's launch mode (launch: nonet) — a
// property of the launch, never a wrapper constructed around the
// command.
type LaunchOpts struct {
	Netns bool
}

// OS — the production binding over the os/exec package.
type OS struct{}

func (OS) LookPath(file string) (string, error) { return exec.LookPath(file) }

func (OS) Start(argv []string, dir string, stdin *string) (Process, error) {
	return OS{}.StartCtx(context.Background(), argv, dir, stdin)
}

func (OS) StartOpts(argv []string, dir string, stdin *string, opts LaunchOpts) (Process, error) {
	if !opts.Netns {
		return OS{}.Start(argv, dir, stdin)
	}
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = dir
	// The namespace discipline is the OS's to give (see the tagged
	// netnsSysProcAttr files): Linux raises the denial, a system
	// without it answers an honest start failure. Nothing is added
	// to argv.
	attrs, err := netnsSysProcAttr()
	if err != nil {
		return nil, err
	}
	cmd.SysProcAttr = attrs
	if stdin != nil {
		cmd.Stdin = strings.NewReader(*stdin)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return &osProcess{cmd: cmd, stdout: &stdout, stderr: &stderr}, nil
}

func (OS) StartCtx(ctx context.Context, argv []string, dir string, stdin *string) (Process, error) {
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Dir = dir
	// Its own process group: the command's descendants are taken down
	// by the group — they hold no resources after the row.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if stdin != nil {
		cmd.Stdin = strings.NewReader(*stdin)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return &osProcess{cmd: cmd, stdout: &stdout, stderr: &stderr}, nil
}

func (OS) Combined(ctx context.Context, argv []string, dir string, timeout time.Duration) (int, string, bool, error) {
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Dir = dir
	var combined bytes.Buffer
	cmd.Stdout = &combined
	cmd.Stderr = &combined
	if err := cmd.Start(); err != nil {
		return 0, "", false, err
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		if err == nil {
			return 0, combined.String(), false, nil
		}
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return ee.ExitCode(), combined.String(), false, nil
		}
		return 0, "", false, err
	case <-time.After(timeout):
		pgid := cmd.Process.Pid
		_ = syscall.Kill(-pgid, syscall.SIGKILL)
		_ = cmd.Process.Kill()
		return 0, combined.String(), true, nil
	}
}

func (OS) SpawnDetached(argv []string, dir string) error {
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = dir
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() { _ = cmd.Wait() }()
	return nil
}

type osProcess struct {
	cmd    *exec.Cmd
	stdout *bytes.Buffer
	stderr *bytes.Buffer
	once   sync.Once
}

func (p *osProcess) Wait(timeout time.Duration) Outcome {
	var outcome Outcome
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
		outcome.TimedOut = true
	}
	outcome.Stdout = p.stdout.String()
	outcome.Stderr = p.stderr.String()
	return outcome
}

func (p *osProcess) KillGroup() {
	p.once.Do(func() {
		if p.cmd.Process == nil {
			return
		}
		pgid := p.cmd.Process.Pid
		_ = syscall.Kill(-pgid, syscall.SIGKILL)
		_ = p.cmd.Process.Kill()
		deadline := time.Now().Add(1 * time.Second)
		for {
			if err := syscall.Kill(-pgid, 0); err != nil {
				return // the group is gone
			}
			if time.Now().After(deadline) {
				return
			}
			time.Sleep(25 * time.Millisecond)
		}
	})
}

// Bind — the active port binding (the http.DefaultClient pattern):
// production never touches it, tests swap it per-case and restore it
// on cleanup.
var Bind Port = OS{}
