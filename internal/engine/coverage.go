// Command family coverage gate: entry into the code stage is
// forbidden while any command family of the brief has no executable
// scenario. A family is the literal tokens of a command sequence from
// the wish text (an occurrence in backticks whose first token names an
// executable file of the surface); backticks of output formats
// ("<id> <name>") do not count as a family. Parsing is deterministic,
// no LLM.
package engine

import (
	"sort"
	"strings"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/compiler"
)

// isLiteralToken tells a command word from a placeholder or an
// optional group of the brief.
func isLiteralToken(tok string) bool {
	if tok == "" || tok == "..." {
		return false
	}
	return !strings.ContainsAny(tok, "<>[]|")
}

// commandFamilies extracts the distinct command families from the wish
// text: the contents of backticks whose first token is a literal
// naming an executable file of the surface. The family key is its
// literal tokens joined by spaces; the result order is lexicographic.
func commandFamilies(intent string, surfaceNames map[string]bool) []string {
	seen := map[string]bool{}
	var fams []string
	parts := strings.Split(intent, "`")
	for i := 1; i < len(parts); i += 2 {
		toks := strings.Fields(parts[i])
		if len(toks) == 0 || !isLiteralToken(toks[0]) || !surfaceNames[toks[0]] {
			continue
		}
		var lits []string
		for _, t := range toks {
			if isLiteralToken(t) {
				lits = append(lits, t)
			}
		}
		if len(lits) == 0 {
			continue
		}
		key := strings.Join(lits, " ")
		if !seen[key] {
			seen[key] = true
			fams = append(fams, key)
		}
	}
	sort.Strings(fams)
	return fams
}

// uncoveredFamilies — families without a single executable scenario.
// A scenario covers a family when its surface command starts with the
// family's literal tokens.
func uncoveredFamilies(state canon.State, scns []canon.Scenario) []string {
	if state.Intent == nil {
		return nil
	}
	names := map[string]bool{}
	for _, sc := range scns {
		if sc.When != nil && len(sc.When.Command) > 0 {
			names[sc.When.Command[0]] = true
		}
	}
	var uncovered []string
	for _, fam := range commandFamilies(*state.Intent, names) {
		lits := strings.Fields(fam)
		covered := false
		for _, sc := range scns {
			if sc.When == nil || len(sc.When.Command) < len(lits) {
				continue
			}
			if compiler.DegreeOf(sc).Level != compiler.DegreeExecutable {
				continue
			}
			match := true
			for i, l := range lits {
				if sc.When.Command[i] != l {
					match = false
					break
				}
			}
			if match {
				covered = true
				break
			}
		}
		if !covered {
			uncovered = append(uncovered, fam)
		}
	}
	return uncovered
}

// uncoveredFamilies on the builder's scenario map — for stage
// transitions.
func (b *builder) uncoveredFamilies() []string {
	scns := make([]canon.Scenario, 0, len(b.scns))
	for _, sc := range b.scns {
		scns = append(scns, sc)
	}
	return uncoveredFamilies(b.state, scns)
}
