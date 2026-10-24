// punchtape — the single-call machine core.
// Entry point: one call, no arguments. The composition root mounts
// the production ports and hands the call to the machine; the exit
// code is the machine's own. The one auxiliary flag —version is a
// human affordance: it answers the version and exits before any
// machine state is touched. Any other arguments are refused loudly
// with the call's own law — one line, exit 1 — before any machine
// state is touched. The agent-facing surface stays argument-free
// and the contract never mentions the flag.
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

// launch — the one action of a binary launch over its arguments:
// the call itself (no arguments), the human's version ask (exactly
// --version, alone on the line), or the loud refusal of a call that
// carries arguments the contract bans. The returned code is the
// process's own.
func launch(args []string, stdout io.Writer) int {
	switch {
	case len(args) == 0:
		return run()
	case len(args) == 1 && args[0] == "--version":
		fmt.Fprintln(stdout, "punchtape "+application.Version)
		return 0
	default:
		fmt.Fprintln(stdout, "punchtape: the call takes no arguments — run `punchtape` alone; this refusal is the whole action of this launch")
		return 1
	}
}

// run — the mounted call: the production ports, the working
// directory, the one run.
func run() int {
	ports.FSBind = fsys.OS{}
	ports.ExecBind = process.OS{}
	cwd, err := os.Getwd()
	if err != nil {
		return 1
	}
	return application.Run(cwd, os.Stdout)
}
