// punchtape — the single-call machine core.
// Entry point: one call, no arguments — the agent's whole surface. The
// composition root mounts the production ports and hands the call to
// the machine; the exit code is the machine's own. Two flags serve the
// human alone: --version answers before any state is touched, and
// --verify re-proves the zone — the build and every row run again,
// nothing of the zone is written — by its exit code. Any other
// arguments are refused loudly with the call's own law — one line,
// exit 1 — before any machine state is touched.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/neurophant/punchtape/internal/application"
	"github.com/neurophant/punchtape/internal/infrastructure/fsys"
	"github.com/neurophant/punchtape/internal/infrastructure/process"
	"github.com/neurophant/punchtape/internal/ports"
)

func main() {
	os.Exit(launch(os.Args[1:], os.Stdout))
}

// launch — the one action of a binary launch over its arguments: the
// call itself (no arguments), one of the human's two asks — --version
// or --verify, each alone on the line — or the loud refusal of a call
// that carries arguments the contract bans. The returned code is the
// process's own.
func launch(args []string, stdout io.Writer) int {
	switch {
	case len(args) == 0:
		return run()
	case len(args) == 1 && args[0] == "--version":
		fmt.Fprintln(stdout, "punchtape "+application.Version)
		return 0
	case len(args) == 1 && args[0] == "--verify":
		return verify(stdout)
	default:
		fmt.Fprintln(stdout, "punchtape: the call takes no arguments — run `punchtape` alone; this refusal is the whole action of this launch")
		return 1
	}
}

// mounted — a launch's one setup: the production ports and the
// working directory, then the call over them. A vanished working
// directory refuses before the call.
func mounted(call func(cwd string, terminal io.Writer) int, stdout io.Writer) int {
	ports.FSBind = fsys.OS{}
	ports.ExecBind = process.OS{}
	cwd, err := os.Getwd()
	if err != nil {
		return 1
	}
	return call(cwd, stdout)
}

// run — the mounted call: the one run, speaking to the process's
// own stdout.
func run() int {
	return mounted(application.Run, os.Stdout)
}

// verify — the mounted check, speaking to the writer the launch
// received.
func verify(stdout io.Writer) int {
	return mounted(application.Verify, stdout)
}
