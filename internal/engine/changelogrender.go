// CHANGELOG.md — the machine render of "what changed and why":
// one line per applied change of the instance; the source is the
// journal (the single truth), the "why" is the delta kind. The
// language is the wish's language; written at verdict freeze next
// to SPEC.md. A file with the machine marker is regenerated, a
// human-written one is not touched.
package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/journal"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// changelogMarker — the first line of the machine changelog render.
const changelogMarker = "<!-- punchtape: changelog render, machine-written -->"

// writeChangelogRender — CHANGELOG.md next to the instance at verdict
// time: the journal is read in full, the render is a derivative.
func writeChangelogRender(e *Engine) {
	path := filepath.Join(e.Store.Root(), "CHANGELOG.md")
	if data, err := os.ReadFile(path); err == nil {
		if len(data) < len(changelogMarker) || string(data[:len(changelogMarker)]) != changelogMarker {
			return // a human wrote the file — the machine does not touch it
		}
	}
	text, err := e.renderChangelogText()
	if err != nil {
		return
	}
	_ = yamlio.WriteAtomic(path, []byte(text))
}

// renderChangelogText — a line-by-line diff of the instance's life:
// time, change kind (human wording), submission key, effect summary.
func (e *Engine) renderChangelogText() (string, error) {
	state, err := e.Store.State()
	if err != nil {
		return "", err
	}
	lang := state.Language
	var sb strings.Builder
	sb.WriteString(changelogMarker)
	sb.WriteString("\n\n")
	sb.WriteString(canondata.TFor(lang, "changelog.title"))
	sb.WriteString("\n\n")
	sb.WriteString(canondata.TFor(lang, "changelog.head"))
	sb.WriteString("\n\n")
	for _, en := range e.Journal.All() {
		what := changelogKindText(lang, en.DeltaKind)
		if en.Held {
			what += canondata.TFor(lang, "changelog.held")
		}
		fmt.Fprintf(&sb, "%s\n", canondata.TFor(lang, "changelog.entry", canondata.M{
			"time":    en.RecordedAt.UTC().Format("2006-01-02 15:04"),
			"what":    what,
			"key":     en.SubmissionKey,
			"effects": changelogEffects(lang, en),
		}))
	}
	return sb.String(), nil
}

// changelogKindText — the human wording of a delta kind; the keys are
// listed literally (the mirror reconciler sees every norm); kinds
// beyond the registry keep their own names —
// the journal of facts loses nothing.
func changelogKindText(lang, kind string) string {
	switch kind {
	case "intent":
		return canondata.TFor(lang, "changelog.kind.intent")
	case "spec":
		return canondata.TFor(lang, "changelog.kind.spec")
	case "checks":
		return canondata.TFor(lang, "changelog.kind.checks")
	case "code":
		return canondata.TFor(lang, "changelog.kind.code")
	case "fix":
		return canondata.TFor(lang, "changelog.kind.fix")
	case "conventions":
		return canondata.TFor(lang, "changelog.kind.conventions")
	case "amend":
		return canondata.TFor(lang, "changelog.kind.amend")
	case "decision":
		return canondata.TFor(lang, "changelog.kind.decision")
	case "split":
		return canondata.TFor(lang, "changelog.kind.split")
	case "assert":
		return canondata.TFor(lang, "changelog.kind.assert")
	case "waive":
		return canondata.TFor(lang, "changelog.kind.waive")
	case "kb":
		return canondata.TFor(lang, "changelog.kind.kb")
	case "run":
		return canondata.TFor(lang, "changelog.kind.run")
	}
	return kind
}

// changelogEffects — the effect summary of an entry: the canon and
// zone files the change touched; an entry without effects changed
// only state — that is an honest line too.
func changelogEffects(lang string, en journal.Entry) string {
	touched, removed := 0, 0
	for _, ef := range en.Effects {
		if ef.Delete {
			removed++
			continue
		}
		touched++
	}
	if touched == 0 && removed == 0 {
		return canondata.TFor(lang, "changelog.effects.none")
	}
	return canondata.TFor(lang, "changelog.effects", canondata.M{
		"touched": strconv.Itoa(touched), "removed": strconv.Itoa(removed),
	})
}
