// spec.md — the machine's render of the current spec:
// after every accepted submission the machine updates
// work/.punchtape/spec.md — one section per behavior (the name, the
// prose verbatim in the executor's language, the rows table with the
// last outcomes). The template lives in this code; the file is a
// regenerated render, not a journal effect: it is always derivable
// from the canon.
package engine

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// specMarker — the first line of the machine render of the spec.
const specMarker = "<!-- punchtape: spec, machine-written -->"

// ensureSpecRender — write/refresh .punchtape/spec.md from the canon.
// Idempotent: the same canon renders the same bytes; a foreign file
// (no marker) is overwritten — the spec is the machine's own render.
func (e *Engine) ensureSpecRender() {
	atoms, err := e.Store.Atoms()
	if err != nil {
		return
	}
	path := filepath.Join(e.Store.Root(), "spec.md")
	_ = yamlio.WriteAtomic(path, []byte(renderSpec(atoms)))
}

// renderSpec — the spec document: the marker, then one section per
// behavior in id order.
func renderSpec(atoms []canon.Atom) string {
	var sb strings.Builder
	sb.WriteString(specMarker)
	sb.WriteString("\n\n")
	if len(atoms) == 0 {
		sb.WriteString("(no behaviors yet)\n")
		return sb.String()
	}
	for _, a := range atoms {
		fmt.Fprintf(&sb, "## %s — %s\n\n", a.ID, a.Name)
		for _, ln := range strings.Split(strings.TrimRight(a.Prose, "\n"), "\n") {
			if strings.TrimSpace(ln) == "" {
				sb.WriteString("\n")
				continue
			}
			fmt.Fprintf(&sb, "%s\n", ln)
		}
		sb.WriteString("\n")
		fmt.Fprintf(&sb, "| # | run | expects | last |\n")
		fmt.Fprintf(&sb, "|---|-----|---------|------|\n")
		for i, row := range a.Rows {
			outcome := row.Outcome
			if outcome == "" {
				outcome = canon.RowUnknown
			}
			fmt.Fprintf(&sb, "| %d | `%s` | %s | %s |\n",
				i+1, oneLineSpec(row.CmdForm()), oneLineSpec(expectationLine(row)), outcome)
		}
		reason := firstRedReason(a)
		if reason != "" {
			fmt.Fprintf(&sb, "\nlast red: %s\n", oneLineSpec(reason))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// firstRedReason — the atom's first standing red's reason (the table
// carries the outcome word; the reason rides below the table).
func firstRedReason(a canon.Atom) string {
	for _, row := range a.Rows {
		if row.Outcome == canon.RowRed && row.Reason != "" {
			return row.Reason
		}
	}
	return ""
}

// oneLineSpec — one table value's text: pipes and newlines collapse.
func oneLineSpec(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", "; ")
	if len(s) > 120 {
		return s[:120] + "…"
	}
	return s
}
