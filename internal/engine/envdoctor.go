// Env doctor: an environment probe before acceptance. Every argv[0]
// of the conventions recipes is checked with exec.LookPath — cheap,
// deterministic, non-interactive; a missing tool honestly reddens the
// acceptance moment with a diagnosis line and a FIX hint from the
// environment catalog.
// The probe does not replace a run: it catches an environment that
// broke between runs and acceptance (a tool removed from PATH by a new
// session), so acceptance cannot pass silently on an unverifiable
// environment.
package engine

import (
	"os/exec"
	"sort"
	"strings"

	"github.com/neurophant/punchtape/internal/canon"
)

// envDoctor — argv[0] of the conventions recipes not found: command
// names without substitutions and without a path (a path is a project
// file, the recipe run checks it); a deterministic LookPath without
// execution. No recipes — an empty list, nothing to probe.
func (e *Engine) envDoctor() []string {
	conv, err := e.Store.Conventions()
	if err != nil || conv == nil {
		return nil
	}
	names := map[string]bool{}
	for _, rec := range []*canon.Recipe{conv.Build, conv.Types, conv.Lint} {
		if rec == nil {
			continue
		}
		for _, cmd := range rec.Commands {
			if len(cmd) == 0 {
				continue
			}
			tok := cmd[0]
			// {out}/{name} substitutions and slash paths are not an
			// environment tool name: the recipe run checks those itself.
			if tok == "" || strings.ContainsAny(tok, "/{") {
				continue
			}
			names[tok] = true
		}
	}
	var missing []string
	for name := range names {
		if _, err := exec.LookPath(name); err != nil {
			missing = append(missing, name)
		}
	}
	sort.Strings(missing)
	return missing
}
