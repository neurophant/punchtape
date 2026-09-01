// SUMMARY.md — the machine render of a human-readable verdict summary:
// one page, four mandatory blocks — done / checked / remaining /
// price — in the language of the wish. Computed from state, ledger,
// and canon without the hand's involvement; written by the same verdict
// freeze as SPEC.md. A file with the machine marker is regenerated;
// a human-written one is untouched — the agent-report policy.
package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// summaryMarker — the first line of the machine render of the summary.
const summaryMarker = "<!-- punchtape: verdict summary, machine-written -->"

// writeSummaryRender — SUMMARY.md next to the instance at verdict time.
func writeSummaryRender(e *Engine) {
	path := filepath.Join(e.Store.Root(), "SUMMARY.md")
	if data, err := os.ReadFile(path); err == nil {
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

// summaryMore — the cap on list items in the summary: one page —
// the summary names the count, the verdict enumerates everything.
const summaryMore = 3

// renderSummaryText — four blocks in the language of the wish from machine
// facts; nothing is stored, the view is assembled on the fly.
func (e *Engine) renderSummaryText() (string, error) {
	state, err := e.Store.State()
	if err != nil {
		return "", err
	}
	lang := state.Language
	reqs, err := e.Store.Requirements()
	if err != nil {
		return "", err
	}
	scns, err := e.Store.Scenarios()
	if err != nil {
		return "", err
	}
	cards, err := e.Store.Cards()
	if err != nil {
		return "", err
	}
	chk, err := e.Store.Checks()
	if err != nil {
		return "", err
	}

	// Counters for the "checked" block: green checks, traceability.
	var greenChk, redChk, greenScn, executable, prose int
	greenByScn := map[string]bool{}
	for _, c := range chk {
		switch c.Outcome {
		case canon.CheckGreen:
			greenChk++
		case canon.CheckRed:
			redChk++
		}
		if c.Outcome == canon.CheckGreen {
			greenByScn[c.Scenario] = true
		}
	}
	for _, sc := range scns {
		if sc.Prose != nil {
			prose++
			continue
		}
		executable++
		if greenByScn[sc.ID] {
			greenScn++
		}
	}

	// Counters for the "done" block: the flow of submissions and runs from the journal.
	submitted, runs := 0, 0
	for _, en := range e.Journal.All() {
		if en.DeltaKind == deltaKindRun {
			runs++
		} else {
			submitted++
		}
	}
	wish := canondata.TFor(lang, "why.handoff.none")
	if state.Intent != nil {
		wish = strings.TrimSpace(*state.Intent)
	}
	verdict := state.Stage
	if state.Stage == canon.StageDeliver {
		if v, err := e.RenderVerdict(); err == nil {
			verdict = v.Status
		}
	}

	// The "remaining" block: the taste queue, quarantine, pending amends —
	// from the same sources as the verdict.
	var taste []string
	for _, sc := range scns {
		if sc.Prose != nil {
			taste = append(taste, sc.ID)
		}
	}
	for _, w := range state.Waives {
		taste = append(taste, w.Rule)
	}
	var quarantined []string
	for _, c := range cards {
		if c.Status == canon.CardQuarantine {
			quarantined = append(quarantined, c.ID)
		}
	}
	pending := 0
	if state.PendingAmend != nil {
		pending = len(state.PendingAmend.Changes)
	}

	// The "price" block: per stage from the ledger. Platform tokens unreported —
	// the lines about them stay silent: listing "no data" for every stage
	// tells a human nothing; wall time and calls are machine facts.
	wallByStage := map[string]int64{}
	tokensByStage := map[string]int64{}
	tokensKnown := map[string]bool{}
	var totalWall, totalTokens, totalCalls int64
	for _, en := range e.Ledger.All() {
		wallByStage[en.Stage] += en.WallMs
		totalWall += en.WallMs
		if en.Event == "verb" {
			totalCalls++
		}
		if en.Tokens != nil {
			tokensByStage[en.Stage] += *en.Tokens
			tokensKnown[en.Stage] = true
			totalTokens += *en.Tokens
		}
	}
	anyTokens := len(tokensKnown) > 0

	var sb strings.Builder
	sb.WriteString(summaryMarker)
	sb.WriteString("\n\n")
	sb.WriteString(canondata.TFor(lang, "summary.title"))
	sb.WriteString("\n\n")

	sb.WriteString(canondata.TFor(lang, "summary.made.head"))
	sb.WriteString("\n")
	// The wish — in full: as a single origin line it was cut
	// mid-thought, and the summary's reader is a human with the brief in mind.
	if wish != canondata.TFor(lang, "why.handoff.none") {
		fmt.Fprintf(&sb, "%s\n", canondata.TFor(lang, "summary.made.wish-head"))
		for _, ln := range strings.Split(wish, "\n") {
			fmt.Fprintf(&sb, "> %s\n", ln)
		}
	} else {
		fmt.Fprintf(&sb, "%s\n", canondata.TFor(lang, "summary.made.wish-none"))
	}
	fmt.Fprintf(&sb, "%s\n", canondata.TFor(lang, "summary.made.composition", canondata.M{
		"reqs": strconv.Itoa(len(reqs)), "scenarios": strconv.Itoa(len(scns)),
		"prose": strconv.Itoa(prose), "cards": strconv.Itoa(len(cards)),
	}))
	fmt.Fprintf(&sb, "%s\n", canondata.TFor(lang, "summary.made.flow", canondata.M{
		"submitted": strconv.Itoa(submitted), "runs": strconv.Itoa(runs),
	}))
	fmt.Fprintf(&sb, "%s\n\n", canondata.TFor(lang, "summary.made.verdict", canondata.M{
		"status": verdict,
	}))

	sb.WriteString(canondata.TFor(lang, "summary.checked.head"))
	sb.WriteString("\n")
	fmt.Fprintf(&sb, "%s\n", canondata.TFor(lang, "summary.checked.checks", canondata.M{
		"green": strconv.Itoa(greenChk), "red": strconv.Itoa(redChk),
		"total": strconv.Itoa(len(chk)),
	}))
	fmt.Fprintf(&sb, "%s\n", canondata.TFor(lang, "summary.checked.trace", canondata.M{
		"green": strconv.Itoa(greenScn), "executable": strconv.Itoa(executable),
	}))
	fmt.Fprintf(&sb, "%s\n", canondata.TFor(lang, "summary.checked.derived", canondata.M{
		"line": e.summaryDerivedLine(state, lang),
	}))
	fmt.Fprintf(&sb, "%s\n\n", canondata.TFor(lang, "summary.checked.acceptance", canondata.M{
		"line": e.summaryAcceptanceLine(lang),
	}))

	sb.WriteString(canondata.TFor(lang, "summary.remaining.head"))
	sb.WriteString("\n")
	switch {
	case len(taste) == 0 && len(quarantined) == 0 && pending == 0:
		fmt.Fprintf(&sb, "%s\n", canondata.TFor(lang, "summary.remaining.none"))
	default:
		if n := len(taste); n > 0 {
			fmt.Fprintf(&sb, "%s\n", canondata.TFor(lang, "summary.remaining.taste", canondata.M{
				"items": summaryList(taste),
			}))
			if n > summaryMore {
				fmt.Fprintf(&sb, "%s\n", canondata.TFor(lang, "summary.remaining.more", canondata.M{
					"count": strconv.Itoa(n),
				}))
			}
		}
		if n := len(quarantined); n > 0 {
			fmt.Fprintf(&sb, "%s\n", canondata.TFor(lang, "summary.remaining.quarantine", canondata.M{
				"items": summaryList(quarantined),
			}))
			if n > summaryMore {
				fmt.Fprintf(&sb, "%s\n", canondata.TFor(lang, "summary.remaining.more", canondata.M{
					"count": strconv.Itoa(n),
				}))
			}
		}
		if pending > 0 {
			fmt.Fprintf(&sb, "%s\n", canondata.TFor(lang, "summary.remaining.pending", canondata.M{
				"count": strconv.Itoa(pending),
			}))
		}
	}
	sb.WriteString("\n")

	sb.WriteString(canondata.TFor(lang, "summary.price.head"))
	sb.WriteString("\n")
	for _, stage := range canon.StageOrder {
		wall, ok := wallByStage[stage]
		if !ok {
			continue
		}
		if tokensKnown[stage] {
			fmt.Fprintf(&sb, "%s\n", canondata.TFor(lang, "summary.price.stage", canondata.M{
				"stage": summaryStageName(lang, stage), "wall": formatWall(wall),
				"tokens": formatTokens(tokensByStage[stage]),
			}))
			continue
		}
		fmt.Fprintf(&sb, "%s\n", canondata.TFor(lang, "summary.price.stage-notokens", canondata.M{
			"stage": summaryStageName(lang, stage), "wall": formatWall(wall),
		}))
	}
	if anyTokens {
		fmt.Fprintf(&sb, "%s\n", canondata.TFor(lang, "summary.price.total", canondata.M{
			"wall": formatWall(totalWall), "tokens": formatTokens(totalTokens),
			"calls": strconv.FormatInt(totalCalls, 10),
		}))
	} else {
		fmt.Fprintf(&sb, "%s\n", canondata.TFor(lang, "summary.price.total-notokens", canondata.M{
			"wall": formatWall(totalWall), "calls": strconv.FormatInt(totalCalls, 10),
		}))
	}
	return sb.String(), nil
}

// summaryList — the first items of a summary list, comma-separated.
func summaryList(xs []string) string {
	if len(xs) > summaryMore {
		xs = xs[:summaryMore]
	}
	return strings.Join(xs, ", ")
}

// summaryDerivedLine — derived variants in the language of the wish: the same
// numbers as the verdict line (the report cache is shared), without the
// protocol tail.
func (e *Engine) summaryDerivedLine(state canon.State, lang string) string {
	if state.Stage != canon.StageDeliver {
		return canondata.TFor(lang, "summary.derived.not-run")
	}
	scns, err := e.Store.Scenarios()
	if err != nil {
		return canondata.TFor(lang, "summary.derived.none")
	}
	chk, err := e.Store.Checks()
	if err != nil {
		return canondata.TFor(lang, "summary.derived.none")
	}
	d, ok := surfaceDigest(e, chk)
	if !ok {
		return canondata.TFor(lang, "summary.derived.none")
	}
	green, red, total := loadOrRunDerived(e, scns, d).counts()
	return canondata.TFor(lang, "summary.derived.line", canondata.M{
		"total": strconv.Itoa(total), "green": strconv.Itoa(green), "red": strconv.Itoa(red),
		"budget": strconv.Itoa(canondata.Limit("derive.budget")),
	})
}

// summaryAcceptanceLine — the acceptance moment in the language of the wish:
// PASS and the digest stay identifiers, the tail is localized.
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
	if last.Details["env"] != "" {
		return canondata.TFor(lang, "summary.acceptance.env", canondata.M{
			"diag": last.Details["env"],
		})
	}
	if last.Details["outcome"] == "green" {
		return canondata.TFor(lang, "summary.acceptance.pass", canondata.M{
			"digest": last.Details["digest"], "green": last.Details["green"],
			"total": last.Details["checks"],
		})
	}
	return canondata.TFor(lang, "summary.acceptance.fail", canondata.M{
		"digest": last.Details["digest"], "green": last.Details["green"],
		"total": last.Details["checks"],
	})
}

// summaryStageName — a stage name in the language of the wish: the machine
// stage code does not ride into the human block; the keys are listed literally
// (the mirror checker sees every norm).
func summaryStageName(lang, stage string) string {
	switch stage {
	case canon.StageIntake:
		return canondata.TFor(lang, "summary.stage.intake")
	case canon.StageSpec:
		return canondata.TFor(lang, "summary.stage.spec")
	case canon.StageACCompile:
		return canondata.TFor(lang, "summary.stage.ac-compile")
	case canon.StageImplement:
		return canondata.TFor(lang, "summary.stage.implement")
	case canon.StageReview:
		return canondata.TFor(lang, "summary.stage.review")
	case canon.StageDeliver:
		return canondata.TFor(lang, "summary.stage.deliver")
	}
	return stage
}

// formatTokens — tokens with a thousands separator, for human reading.
func formatTokens(n int64) string {
	s := strconv.FormatInt(n, 10)
	if len(s) <= 3 {
		return s
	}
	var parts []string
	for len(s) > 3 {
		parts = append([]string{s[len(s)-3:]}, parts...)
		s = s[:len(s)-3]
	}
	parts = append([]string{s}, parts...)
	return strings.Join(parts, " ")
}
