// SUMMARY.md — the machine render of a human-readable verdict summary:
// one page — done / checked / price — in the language of the brief.
// Computed from state, ledger, and canon without the executor's
// involvement; written by the verdict freeze. A file with the machine
// marker is regenerated; a human-written one is untouched — the
// agent-report policy.
package engine

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/fsport"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// summaryMarker — the first line of the machine render of the summary.
const summaryMarker = "<!-- punchtape: verdict summary, machine-written -->"

// writeSummaryRender — SUMMARY.md inside the instance at verdict time.
func writeSummaryRender(e *Engine) {
	path := filepath.Join(e.Store.Root(), "SUMMARY.md")
	if data, err := fsport.Bind.ReadFile(path); err == nil {
		if len(data) < len(summaryMarker) || string(data[:len(summaryMarker)]) != summaryMarker {
			return // a human wrote the file — the machine does not touch it
		}
	}
	text, err := e.renderSummaryText()
	if err != nil {
		return
	}
	_ = yamlio.WriteAtomic(path, []byte(text))
}

// renderSummaryText — three blocks in the language of the brief from
// machine facts; nothing is stored, the view is assembled on the fly.
func (e *Engine) renderSummaryText() (string, error) {
	state, err := e.Store.State()
	if err != nil {
		return "", err
	}
	lang := state.Language
	atoms, err := e.Store.Atoms()
	if err != nil {
		return "", err
	}

	// Counters: green/red rows, green atoms.
	var greenRows, redRows, greenAtoms, totalRows int
	for _, a := range atoms {
		totalRows += len(a.Rows)
		if a.Green() {
			greenAtoms++
		}
		for _, r := range a.Rows {
			if r.Outcome == canon.RowGreen {
				greenRows++
			} else if r.Outcome == canon.RowRed {
				redRows++
			}
		}
	}

	// The flow of submissions and runs from the journal.
	submitted, runs := 0, 0
	for _, en := range e.Journal.All() {
		switch en.DeltaKind {
		case "run":
			runs++
		case "spec":
			submitted++
		}
	}
	verdict := state.Stage
	if state.Stage == canon.StageDeliver {
		if v, err := e.RenderVerdict(); err == nil {
			verdict = v.Status
		}
	}

	// The price block: wall time and calls are machine facts.
	var totalWall, totalCalls int64
	for _, en := range e.Ledger.All() {
		totalWall += en.WallMs
		if en.Event == "verb" {
			totalCalls++
		}
	}

	var sb strings.Builder
	sb.WriteString(summaryMarker)
	sb.WriteString("\n\n")
	sb.WriteString(canondata.TFor(lang, "summary.title"))
	sb.WriteString("\n\n")

	sb.WriteString(canondata.TFor(lang, "summary.made.head"))
	sb.WriteString("\n")
	// The brief — in full: the summary's reader is a human with the
	// brief in mind.
	if state.Wish != nil {
		fmt.Fprintf(&sb, "%s\n", canondata.TFor(lang, "summary.made.wish-head"))
		for _, ln := range strings.Split(strings.TrimSpace(*state.Wish), "\n") {
			fmt.Fprintf(&sb, "> %s\n", ln)
		}
	} else {
		fmt.Fprintf(&sb, "%s\n", canondata.TFor(lang, "summary.made.wish-none"))
	}
	fmt.Fprintf(&sb, "%s\n", canondata.TFor(lang, "summary.made.composition", canondata.M{
		"atoms": strconv.Itoa(len(atoms)), "rows": strconv.Itoa(totalRows),
	}))
	fmt.Fprintf(&sb, "%s\n", canondata.TFor(lang, "summary.made.flow", canondata.M{
		"submitted": strconv.Itoa(submitted), "runs": strconv.Itoa(runs),
	}))
	fmt.Fprintf(&sb, "%s\n\n", canondata.TFor(lang, "summary.made.verdict", canondata.M{
		"status": verdict,
	}))

	sb.WriteString(canondata.TFor(lang, "summary.checked.head"))
	sb.WriteString("\n")
	fmt.Fprintf(&sb, "%s\n", canondata.TFor(lang, "summary.checked.rows", canondata.M{
		"green": strconv.Itoa(greenRows), "red": strconv.Itoa(redRows),
		"atoms": strconv.Itoa(greenAtoms), "total": strconv.Itoa(len(atoms)),
	}))
	fmt.Fprintf(&sb, "%s\n\n", canondata.TFor(lang, "summary.checked.acceptance", canondata.M{
		"line": e.summaryAcceptanceLine(lang),
	}))

	sb.WriteString(canondata.TFor(lang, "summary.price.head"))
	sb.WriteString("\n")
	fmt.Fprintf(&sb, "%s\n", canondata.TFor(lang, "summary.price.total", canondata.M{
		"wall": formatWall(totalWall), "calls": strconv.FormatInt(totalCalls, 10),
	}))
	// The ceremony meter: the reply bytes the hand paid, by render
	// family — the ledger's own counters, deterministic arithmetic.
	cer := e.CeremonyBytes()
	fmt.Fprintf(&sb, "%s\n", canondata.TFor(lang, "summary.price.ceremony", canondata.M{
		"contract": strconv.FormatInt(cer["contract"], 10),
		"spec":     strconv.FormatInt(cer["spec"], 10),
		"report":   strconv.FormatInt(cer["report"], 10),
		"door":     strconv.FormatInt(cer["door-echo"], 10),
	}))
	return sb.String(), nil
}

// summaryAcceptanceLine — the acceptance moment in the language of
// the brief: the digest stays an identifier, the tail is localized.
func (e *Engine) summaryAcceptanceLine(lang string) string {
	if drifted, oldD, newD := e.acceptanceDrift(); drifted {
		return canondata.TFor(lang, "summary.acceptance.drift", canondata.M{
			"old": oldD, "new": newD,
		})
	}
	last := lastAcceptanceEntry(e)
	if last == nil {
		return canondata.TFor(lang, "summary.acceptance.not-run")
	}
	return canondata.TFor(lang, "summary.acceptance.pass", canondata.M{
		"digest": last.Details["digest"], "green": last.Details["green"],
		"total": last.Details["checks"],
	})
}
