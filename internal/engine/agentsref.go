// Agent reference of the instance: at initialization the machine places
// an AGENTS.md file next to .punchtape/ — a shared convention for
// executor environments that agents read themselves. A new session in
// the same directory picks up the contract without verbal instructions.
// Interface only: verbs, the loop, machine zones; there are and can be
// no behavior rules here. The text is delivery canon data; the code
// holds only the identity marker. A work root that already carries the
// repository's own AGENTS.md is COMBINED: the machine contract first
// with declared total precedence, the repository's rules preserved
// verbatim below the tail boundary — full initialization, zero loss.
package engine

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// refMarker — the first line of the machine reference: the machine
// regenerates a file carrying this marker; a foreign file without the
// marker is combined under the machine contract.
const refMarker = "<!-- punchtape: agent reference, machine-written -->"

// tailBoundary — the line that closes the machine head of a combined
// reference: the repository's own rules live below it, preserved
// verbatim. The FIRST occurrence in a machine file is the boundary
// itself (the machine head owns the file's opening), so a repository
// text quoting the same line cannot displace it.
const tailBoundary = "<!-- punchtape: repo rules below, preserved verbatim -->"

// AgentsRefText — the wholly machine-written reference: the identity
// marker plus text from canon data. Public for injecting the contract
// into the first next reply and every --full (fresh session): the
// executor gets the contract into its context automatically; choosing
// to "read the file" is not required.
func AgentsRefText() string {
	return refMarker + "\n\n" + canondata.T("ref.agents.text") + "\n"
}

// combinedRefText — the machine head (the contract, always first,
// with the precedence line) over the repository's own reference
// preserved byte for byte below the tail boundary: a full instance
// initialization without losing the repository's rules. Pure
// concatenation — the machine learns nothing about the content.
func combinedRefText(foreign []byte) string {
	return AgentsRefText() + "\n" +
		canondata.T("ref.agents.precedence") + "\n" +
		tailBoundary + "\n" + string(foreign)
}

// machineHead — the current head of a combined reference (without
// the tail): what the machine regenerates when its canon changes.
func machineHead() string {
	return AgentsRefText() + "\n" + canondata.T("ref.agents.precedence") + "\n" + tailBoundary + "\n"
}

func agentsRefText() string { return AgentsRefText() }

// ensureAgentsRef writes/maintains the AGENTS.md reference next to the
// instance. Idempotent: a missing file is created; a machine one (with
// the marker) is regenerated when the canon head changed, its foreign
// tail (repository rules below the boundary) preserved verbatim; a
// foreign file is combined under the machine contract — the head
// always first, always prevailing, the repository text kept as is.
func ensureAgentsRef(workdir string) {
	path := filepath.Join(workdir, "AGENTS.md")
	data, err := os.ReadFile(path)
	if err != nil {
		_ = yamlio.WriteAtomic(path, []byte(agentsRefText()))
		return
	}
	if len(data) >= len(refMarker) && string(data[:len(refMarker)]) == refMarker {
		head, tail, hasTail := splitRefTail(data)
		if !hasTail {
			if strings.TrimRight(string(data), "\n") == strings.TrimRight(agentsRefText(), "\n") {
				return
			}
			_ = yamlio.WriteAtomic(path, []byte(agentsRefText()))
			return
		}
		if head == machineHead() {
			return // the canon head is current, the tail is the repository's own
		}
		_ = yamlio.WriteAtomic(path, []byte(machineHead()+tail))
		return
	}
	_ = yamlio.WriteAtomic(path, []byte(combinedRefText(data)))
}

// splitRefTail — a machine-written reference into its head (up to the
// tail boundary) and the foreign tail (everything after it, verbatim).
func splitRefTail(data []byte) (head, tail string, hasTail bool) {
	i := bytes.Index(data, []byte(tailBoundary))
	if i < 0 {
		return "", "", false
	}
	end := i + len(tailBoundary)
	for end < len(data) && data[end] == '\n' {
		end++
	}
	return string(data[:end]), string(data[end:]), true
}
