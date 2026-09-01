// punchtape — the deterministic "spec → code with guarantees" core.
// Entry point: verb parsing and handoff to the engine.
package main

import (
	"os"

	"github.com/neurophant/punchtape/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdout))
}
