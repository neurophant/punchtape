// Agent reference of the instance: at initialization the machine places
// an AGENTS.md file next to .punchtape/ — a shared convention for
// executor environments that agents read themselves. A new session in
// the same directory picks up the contract without verbal instructions.
// Interface only: verbs, the loop, machine zones; there are and can be
// no behavior rules here. The text is delivery canon data; the code
// holds only the identity marker.
package engine

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// refMarker — the first line of the machine reference: the machine
// regenerates a file carrying this marker; a foreign file without the
// marker is left untouched.
const refMarker = "<!-- punchtape: agent reference, machine-written -->"

// AgentsRefText — the wholly machine-written reference: the identity
// marker plus text from canon data. Public for injecting the contract
// into the first next reply and every --full (fresh session): the
// executor gets the contract into its context automatically; choosing
// to "read the file" is not required.
func AgentsRefText() string {
	return refMarker + "\n\n" + canondata.T("ref.agents.text") + "\n"
}

func agentsRefText() string { return AgentsRefText() }

// ensureAgentsRef writes/maintains the AGENTS.md reference next to the
// instance. Idempotent: a missing file is created, a machine one (with
// the marker) is regenerated, a user one is left untouched.
func ensureAgentsRef(workdir string) {
	path := filepath.Join(workdir, "AGENTS.md")
	data, err := os.ReadFile(path)
	if err == nil {
		if len(data) >= len(refMarker) && string(data[:len(refMarker)]) == refMarker {
			if strings.TrimRight(string(data), "\n") == strings.TrimRight(agentsRefText(), "\n") {
				return
			}
		} else {
			return // a human wrote the file — the machine does not touch it
		}
	}
	_ = yamlio.WriteAtomic(path, []byte(agentsRefText()))
}
