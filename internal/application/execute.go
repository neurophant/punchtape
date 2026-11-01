package application

// The execution of the build and the rows (the law the contract's
// law:your-tree names): every command runs in the working root, in
// the declared clean environment, as one process group; a row's
// deadline is 300 seconds, and whatever the command left alive is
// taken down with it. The command's own bytes are never touched.

import (
	"fmt"
	"time"

	"github.com/neurophant/punchtape/internal/ports"

	"github.com/neurophant/punchtape/internal/domain/doc"
)

// CommandResult — what one executed command left behind.
type CommandResult struct {
	Exit     int
	Stdout   string
	Stderr   string
	TimedOut bool
	Err      error
}

// Argv — the command's argv: a shell line rides through the shell, an
// argv list runs as is.
func Argv(c doc.Cmd) []string {
	if c.Argv != nil {
		return c.Argv
	}
	return []string{"sh", "-c", c.Line}
}

// RunCommand — one command under the whole law, in the given
// environment (the clean base plus the zone's pins), under the
// given deadline — the run's own ceiling, the zone's pin or the
// standing default.
func RunCommand(argv []string, root string, env []string, deadline time.Duration) CommandResult {
	proc, err := ports.ExecBind.StartOpts(argv, root, ports.LaunchOpts{Env: env})
	if err != nil {
		return CommandResult{Err: err}
	}
	outcome := proc.Wait(deadline)
	// Whatever the command left alive dies with its group: a dying
	// member still holds what it held.
	proc.KillGroup()
	// A wait that could not be answered is a failure the machine has
	// seen: it must reach the facts, never pass for a clean exit.
	if outcome.WaitErr != "" {
		return CommandResult{Err: fmt.Errorf("wait: %s", outcome.WaitErr)}
	}
	return CommandResult{
		Exit:     outcome.Exit,
		Stdout:   outcome.Stdout,
		Stderr:   outcome.Stderr,
		TimedOut: outcome.TimedOut,
	}
}
