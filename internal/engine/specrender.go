// SPEC.md — the machine render of the spec from the canon: the "spec"
// artifact of the system hand is written entirely by the machine (the
// canon is the source, the render is a derivative), manual
// duplication is not needed. Written at the freeze of the delivery
// verdict; a file with the machine marker is regenerated, a
// human-written one is not touched — the same policy as the agent
// reference.
package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// specMarker — the first line of the machine spec render.
const specMarker = "<!-- punchtape: spec render, machine-written -->"

// writeSpecRender — SPEC.md next to the instance, from the current canon.
func writeSpecRender(e *Engine) {
	path := filepath.Join(e.Store.Root(), "SPEC.md")
	if data, err := os.ReadFile(path); err == nil {
		if len(data) < len(specMarker) || string(data[:len(specMarker)]) != specMarker {
			return // a human wrote the file — the machine does not touch it
		}
	}
	text, err := e.renderSpecText()
	if err != nil {
		return
	}
	_ = yamlio.WriteAtomic(path, []byte(text))
}

// renderSpecText — the human-readable render of the spec: the wish
// verbatim, requirements with their check rows, the unmeasurable
// remainder as a separate section. Scenarios are printed in table
// form — the same one the executor interface accepts them in.
func (e *Engine) renderSpecText() (string, error) {
	state, err := e.Store.State()
	if err != nil {
		return "", err
	}
	reqs, err := e.Store.Requirements()
	if err != nil {
		return "", err
	}
	scns, err := e.Store.Scenarios()
	if err != nil {
		return "", err
	}
	chk, err := e.Store.Checks()
	if err != nil {
		return "", err
	}
	byReq := map[string][]canon.Scenario{}
	var prose []canon.Scenario
	for _, sc := range scns {
		if sc.Prose != nil {
			prose = append(prose, sc)
			continue
		}
		byReq[sc.Requirement] = append(byReq[sc.Requirement], sc)
	}
	outcomeOf := map[string]string{}
	for _, c := range chk {
		outcomeOf[c.Scenario] = c.Outcome
	}

	var sb strings.Builder
	sb.WriteString(specMarker)
	sb.WriteString("\n\n")
	sb.WriteString(canondata.T("specrender.title"))
	sb.WriteString("\n\n")
	if state.Intent != nil {
		sb.WriteString(canondata.T("specrender.wish"))
		sb.WriteString("\n\n")
		sb.WriteString(strings.TrimSpace(*state.Intent))
		sb.WriteString("\n\n")
	}
	sb.WriteString(canondata.T("specrender.requirements"))
	sb.WriteString("\n\n")
	for _, r := range reqs {
		fmt.Fprintf(&sb, "### %s — %s\n", r.ID, r.Formulation)
		rows := byReq[r.ID]
		sort.Slice(rows, func(i, j int) bool { return rows[i].ID < rows[j].ID })
		if len(rows) == 0 {
			sb.WriteString(canondata.T("specrender.no-scenarios"))
			sb.WriteString("\n")
			continue
		}
		for _, sc := range rows {
			fmt.Fprintf(&sb, "\n%s · %s\n", sc.ID, sc.Summary)
			if len(sc.Seed) > 0 || len(sc.Materials) > 0 {
				files := make([]string, 0, len(sc.Seed)+len(sc.Materials))
				for _, sd := range sc.Seed {
					files = append(files, sd.Path)
				}
				for _, m := range sc.Materials {
					files = append(files, m.From)
				}
				fmt.Fprintf(&sb, "  seeds: %s\n", strings.Join(files, ", "))
			}
			for _, pre := range sc.Pre {
				fmt.Fprintf(&sb, "  setup: %s\n", strings.Join(pre, " "))
			}
			if sc.When != nil {
				fmt.Fprintf(&sb, "  run:   %s\n", strings.Join(sc.When.Command, " "))
				if len(sc.When.Volatile) > 0 {
					fmt.Fprintf(&sb, "  volatile: %s\n", strings.Join(sc.When.Volatile, ", "))
				}
			}
			for _, line := range renderThenLines(sc.Then) {
				fmt.Fprintf(&sb, "  %s\n", line)
			}
		}
		sb.WriteString("\n")
	}
	if len(prose) > 0 {
		sort.Slice(prose, func(i, j int) bool { return prose[i].ID < prose[j].ID })
		sb.WriteString(canondata.T("specrender.unmeasurable"))
		sb.WriteString("\n\n")
		for _, sc := range prose {
			fmt.Fprintf(&sb, "- %s: %s\n", sc.ID, strings.ReplaceAll(*sc.Prose, "\n", " "))
		}
	}
	return sb.String(), nil
}

// escapeControls — control bytes become escapes, everything else
// (including the backslash) is printed as is: the spec render stays a
// text file over binary expectations without losing the verbatim
// quality of ordinary values.
func escapeControls(s string) string {
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		b := s[i]
		if b >= 0x20 && b != 0x7F {
			sb.WriteByte(b)
			continue
		}
		switch b {
		case '\n':
			sb.WriteString("\\n")
		case '\t':
			sb.WriteString("\\t")
		case '\r':
			sb.WriteString("\\r")
		default:
			fmt.Fprintf(&sb, "\\x%02X", b)
		}
	}
	return sb.String()
}

// renderThenLines — expectations line by line: compact form, multiline
// values as per-line continuation with the "| " marker: the marker
// separates the render's layout from the value's own indentation (the
// product's real output with leading spaces is seen as is, without
// guessing the depth).
func renderThenLines(then []canon.Assertion) []string {
	var out []string
	for _, a := range then {
		var head string
		switch a.Observation {
		case "stdout":
			head = "out:"
		case "stderr":
			head = "err:"
		case "exit-code":
			head = "rc:"
		case "file":
			head = "state " + a.Path + ":"
		default:
			head = a.Observation + ":"
		}
		switch a.Condition {
		case "equals":
			// the head already names the condition
		case "fails":
			head += " (non-zero)"
		default:
			head += " (" + a.Condition + ")"
		}
		value := escapeControls(a.Value)
		if !strings.Contains(value, "\n") {
			out = append(out, fmt.Sprintf("%s %s", head, value))
			continue
		}
		out = append(out, head)
		for _, line := range strings.Split(strings.TrimRight(value, "\n"), "\n") {
			out = append(out, "  | "+line)
		}
	}
	return out
}
