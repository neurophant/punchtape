package execport

import (
	"context"
	"errors"
	"testing"
	"time"
)

// fakeProcess — a scripted in-memory process (mock code lives in
// this test file only).
type fakeProcess struct {
	outcome Outcome
	hang    bool
	killed  bool
}

func (p *fakeProcess) Wait(timeout time.Duration) Outcome {
	if p.hang {
		// a hanging process only ends by the deadline's kill
		time.Sleep(2 * timeout)
		p.killed = true
		return Outcome{TimedOut: true}
	}
	return p.outcome
}

func (p *fakeProcess) KillGroup() { p.killed = true }

// fakePort — the scripted port: outcomes by argv[0], a start error,
// a lookPath answer.
type fakePort struct {
	outcomes  map[string]Outcome
	hangs     map[string]bool
	starts    [][]string
	startErr  error
	lookFound bool
	lookPath  string
}

func (f *fakePort) LookPath(string) (string, error) {
	if f.lookFound {
		return f.lookPath, nil
	}
	return "", errors.New("not found")
}

func (f *fakePort) Start(argv []string, dir string, stdin *string) (Process, error) {
	if f.startErr != nil {
		return nil, f.startErr
	}
	f.starts = append(f.starts, argv)
	out, ok := f.outcomes[argv[0]]
	if !ok {
		out = Outcome{}
	}
	return &fakeProcess{outcome: out, hang: f.hangs[argv[0]]}, nil
}

func (f *fakePort) StartCtx(_ context.Context, argv []string, dir string, stdin *string) (Process, error) {
	return f.Start(argv, dir, stdin)
}

func (f *fakePort) SpawnDetached(argv []string, dir string) error {
	f.starts = append(f.starts, argv)
	return nil
}

func (f *fakePort) Combined(_ context.Context, argv []string, _ string, _ time.Duration) (int, string, bool, error) {
	if f.startErr != nil {
		return 0, "", false, f.startErr
	}
	out, ok := f.outcomes[argv[0]]
	if !ok {
		return 0, "", false, nil
	}
	return out.Exit, out.Stdout, out.TimedOut, nil
}

// StartOpts — Start with the launch disciplines: the fake answers the
// same either way (the denial is a launch fact the scripted outcomes
// do not depend on).
func (f *fakePort) StartOpts(argv []string, dir string, stdin *string, _ LaunchOpts) (Process, error) {
	return f.Start(argv, dir, stdin)
}

func bindFake(t *testing.T, f *fakePort) {
	t.Helper()
	old := Bind
	Bind = f
	t.Cleanup(func() { Bind = old })
}

func TestPortOverFake(t *testing.T) {
	f := &fakePort{
		outcomes: map[string]Outcome{
			"ok":    {Exit: 0, Stdout: "out\n", Stderr: "err\n"},
			"fail":  {Exit: 3, Stdout: "partial"},
			"weird": {WaitErr: "reap failure"},
		},
		hangs:     map[string]bool{"hang": true},
		lookFound: true, lookPath: "/bin/sh",
	}
	bindFake(t, f)

	stdinText := "feed"
	p, err := f.Start([]string{"ok", "arg"}, "/run", &stdinText)
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	out := p.Wait(time.Second)
	if out.Exit != 0 || out.Stdout != "out\n" || out.Stderr != "err\n" || out.TimedOut || out.WaitErr != "" {
		t.Fatalf("ok outcome: %+v", out)
	}

	out = f.mustStart(t, "fail").Wait(time.Second)
	if out.Exit != 3 {
		t.Fatalf("exit code: %+v", out)
	}

	out = f.mustStart(t, "weird").Wait(time.Second)
	if out.WaitErr == "" {
		t.Fatalf("a wait failure must carry its reason: %+v", out)
	}

	hanging := f.mustStart(t, "hang")
	out = hanging.Wait(10 * time.Millisecond)
	if !out.TimedOut {
		t.Fatalf("a hanging process must time out: %+v", out)
	}

	f.startErr = errors.New("no launch")
	if _, err := Start0("ok"); err == nil {
		t.Fatal("a start failure must surface")
	}

	if got, err := LookPath0("sh"); err != nil || got != "/bin/sh" {
		t.Fatalf("lookpath: %q %v", got, err)
	}
	f.lookFound = false
	if _, err := LookPath0("sh"); err == nil {
		t.Fatal("a missing tool must not be found")
	}

	// the started argv is recorded verbatim with the workdir flag
	if len(f.starts) == 0 || f.starts[0][0] != "ok" || f.starts[0][1] != "arg" {
		t.Fatalf("recorded argv: %v", f.starts)
	}
}

func (f *fakePort) mustStart(t *testing.T, name string) Process {
	t.Helper()
	p, err := f.Start([]string{name}, "/run", nil)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func Start0(name string) (Process, error) { return Bind.Start([]string{name}, "/run", nil) }

func LookPath0(name string) (string, error) { return Bind.LookPath(name) }

// TestOSWiring exercises the production os binding on error paths
// only: launching a nonexistent binary fails and classifies, without
// starting anything real. The success paths of the os binding run
// inside the make test gate's instrumented binary (a real instance
// round-trip with real checks); a unit test must not start real
// processes.
func TestOSWiring(t *testing.T) {
	var osPort OS
	if _, err := osPort.LookPath("definitely-not-a-tool-xyz"); err == nil {
		t.Fatal("a missing tool must not be found")
	}
	if _, err := osPort.Start([]string{"definitely-not-a-tool-xyz"}, "/run", nil); err == nil {
		t.Fatal("a launch of a missing binary must fail")
	}
}
