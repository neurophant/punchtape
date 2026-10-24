// The launcher's battery over the real OS: the declared
// environment, the captured streams, the exit code, the group kill.
package process

import (
	"testing"
	"time"

	"github.com/neurophant/punchtape/internal/ports"
)

func TestLaunchEnvironmentAndStreams(t *testing.T) {
	var osfs OS
	proc, err := osfs.StartOpts(
		[]string{"sh", "-c", "printf '%s' \"$MARKER\"; printf err >&2"},
		t.TempDir(), ports.LaunchOpts{Env: []string{"MARKER=clean"}},
	)
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	outcome := proc.Wait(10 * time.Second)
	if outcome.Exit != 0 || outcome.Stdout != "clean" {
		t.Fatalf("outcome: %+v", outcome)
	}
	if outcome.Stderr != "err" {
		t.Fatalf("stderr: %q", outcome.Stderr)
	}
}

func TestNonZeroExitIsAFact(t *testing.T) {
	var osfs OS
	proc, err := osfs.StartOpts([]string{"sh", "-c", "exit 3"}, t.TempDir(), ports.LaunchOpts{})
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	if outcome := proc.Wait(10 * time.Second); outcome.Exit != 3 {
		t.Fatalf("exit must be measured, got %+v", outcome)
	}
}

func TestGroupKillTakesDescendants(t *testing.T) {
	var osfs OS
	proc, err := osfs.StartOpts(
		[]string{"sh", "-c", "sleep 30 & sleep 30"},
		t.TempDir(), ports.LaunchOpts{},
	)
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	proc.KillGroup()
	if outcome := proc.Wait(5 * time.Second); outcome.Exit == 0 && !outcome.TimedOut {
		t.Fatalf("a killed group must not answer success: %+v", outcome)
	}
}

func TestDoubleWaitNamesTheReapFailure(t *testing.T) {
	var osfs OS
	proc, err := osfs.StartOpts([]string{"true"}, t.TempDir(), ports.LaunchOpts{})
	if err != nil {
		t.Fatal(err)
	}
	first := proc.Wait(10 * time.Second)
	if first.Exit != 0 || first.WaitErr != "" {
		t.Fatalf("the first wait reaps cleanly: %+v", first)
	}
	second := proc.Wait(time.Second)
	if second.WaitErr == "" {
		t.Fatalf("the second wait names the reap failure: %+v", second)
	}
}
