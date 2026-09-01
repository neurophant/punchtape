// Human specs of the instance — the .punchtape/specs directory: a
// shared index.md registry (table of contents: features, addresses,
// statuses) and one file per feature — in living words in the language
// of the wish: the executor's authorial prose (or an honest "no
// description"), calls, behavior, boundaries. Machine terms
// (requirement and scenario identifiers, observation conditions)
// live only as address references, not as text. The working directory
// stays clean: all machine renders are inside the instance data,
// the root holds only the AGENTS.md selfdoc. Written by the verdict
// freeze; a file with the machine marker is regenerated, a
// human-written one is untouched.
package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// featuresMarker — the first line of the machine render of a human spec.
const featuresMarker = "<!-- punchtape: features render, machine-written -->"

// specsDirName — the specs directory inside the instance data.
const specsDirName = "specs"

// writeFeatureRender — the specs directory in the instance data at
// verdict time: the registry is the source, the render is a derivative.
// Feature files that left the registry are removed; the old root render
// (before the move into the instance) too.
func writeFeatureRender(e *Engine) {
	dir := filepath.Join(e.Store.Root(), specsDirName)
	views, err := e.featureViews()
	if err != nil {
		return
	}
	state, err := e.Store.State()
	if err != nil {
		return
	}
	live := map[string]bool{}
	for _, v := range views {
		text, err := renderFeatureFile(state.Language, v)
		if err != nil {
			return
		}
		path := filepath.Join(dir, v.Req.ID+".md")
		if machineWritable(path) {
			if err := yamlio.WriteAtomic(path, []byte(text)); err != nil {
				return
			}
		}
		live[v.Req.ID+".md"] = true
	}
	index, err := renderSpecIndex(state.Language, views, state.Intent)
	if err != nil {
		return
	}
	indexPath := filepath.Join(dir, "index.md")
	if machineWritable(indexPath) {
		if err := yamlio.WriteAtomic(indexPath, []byte(index)); err != nil {
			return
		}
	}
	live["index.md"] = true
	// Cleanup: machine spec files without a feature in the registry and the
	// old root render — a clean working directory matters more than forgotten files.
	entries, err := os.ReadDir(dir)
	if err == nil {
		for _, en := range entries {
			if en.IsDir() || !live[en.Name()] {
				p := filepath.Join(dir, en.Name())
				if machineWritable(p) {
					_ = os.Remove(p)
				}
			}
		}
	}
	removeLegacyRender(e.Workdir, "FEATURES.md")
}

// machineWritable — the path can be regenerated: no file, or its
// first line is the machine marker. Human writing is untouchable.
func machineWritable(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return true
	}
	first := strings.SplitN(string(data), "\n", 2)[0]
	return first == featuresMarker || first == specMarker ||
		first == summaryMarker || first == changelogMarker
}

// removeLegacyRender — remove the machine's old root render: all
// human-readable artifacts moved into the instance data, the root
// holds only the AGENTS.md selfdoc. A foreign file without the marker is not touched.
func removeLegacyRender(workdir, name string) {
	path := filepath.Join(workdir, name)
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	first := strings.SplitN(string(data), "\n", 2)[0]
	if first == featuresMarker || first == specMarker ||
		first == summaryMarker || first == changelogMarker {
		_ = os.Remove(path)
	}
}

// docStatusWord — the status of the authorial description in living words.
func docStatusWord(lang string, doc *canon.FeatureDoc) string {
	switch {
	case doc == nil || strings.TrimSpace(doc.Text) == "":
		return canondata.TFor(lang, "features.doc.none")
	case doc.Ratified:
		return canondata.TFor(lang, "features.doc.ratified")
	default:
		return canondata.TFor(lang, "features.doc.awaiting")
	}
}

// renderSpecIndex — the registry table of contents of the specs directory:
// features, statuses of descriptions and requirements, the number of
// scenarios (row addresses live in the feature files, the registry does
// not haul raw identifier lists); the wish — in full, as a single
// origin line it was cut mid-thought.
func renderSpecIndex(lang string, views []featureView, intent *string) (string, error) {
	var sb strings.Builder
	sb.WriteString(featuresMarker)
	sb.WriteString("\n\n")
	sb.WriteString(canondata.TFor(lang, "features.index.title"))
	sb.WriteString("\n\n")
	sb.WriteString(canondata.TFor(lang, "features.index.note"))
	sb.WriteString("\n")
	if intent != nil {
		wish := strings.TrimSpace(*intent)
		if wish != "" {
			fmt.Fprintf(&sb, "\n%s:\n\n> %s\n", canondata.TFor(lang, "features.index.origin"),
				strings.ReplaceAll(wish, "\n", "\n> "))
		}
	}
	fmt.Fprintf(&sb, "\n| %s | %s | %s | %s | %s |\n",
		canondata.TFor(lang, "features.index.address"),
		canondata.TFor(lang, "features.toc.feature"),
		canondata.TFor(lang, "features.index.doc"),
		canondata.TFor(lang, "features.toc.status"),
		canondata.TFor(lang, "features.toc.scenarios"))
	sb.WriteString("|---|---|---|---|---|\n")
	for _, v := range views {
		fmt.Fprintf(&sb, "| %s.md | %s | %s | %s | %d |\n", v.Req.ID,
			firstLineOf(v.Req.Formulation), docStatusWord(lang, v.Req.SpecDoc),
			reqStatusWord(lang, v.Req.Status), len(v.Scns)+len(v.Prose))
	}
	return sb.String(), nil
}

// renderFeatureFile — one feature's file: the formulation in living words,
// authorial prose (or an honest "no description"), calls, behavior,
// boundaries. Identifiers — as addresses only.
func renderFeatureFile(lang string, v featureView) (string, error) {
	var sb strings.Builder
	sb.WriteString(featuresMarker)
	sb.WriteString("\n\n")
	full := strings.TrimSpace(v.Req.Formulation)
	head := firstLineOf(full)
	fmt.Fprintf(&sb, "# %s — %s\n\n", v.Req.ID, head)
	if full != head {
		fmt.Fprintf(&sb, "%s\n\n", full)
	}
	fmt.Fprintf(&sb, "%s\n", canondata.TFor(lang, "features.section.status",
		canondata.M{"status": reqStatusWord(lang, v.Req.Status)}))
	fmt.Fprintf(&sb, "%s: %s\n", canondata.TFor(lang, "features.index.doc"),
		docStatusWord(lang, v.Req.SpecDoc))
	if v.Req.SpecDoc != nil && strings.TrimSpace(v.Req.SpecDoc.Text) != "" {
		fmt.Fprintf(&sb, "\n%s\n\n%s\n", canondata.TFor(lang, "features.doc.head"),
			strings.TrimSpace(v.Req.SpecDoc.Text))
	}
	if len(v.Req.Dependencies) > 0 {
		fmt.Fprintf(&sb, "\n%s\n", canondata.TFor(lang, "features.section.depends",
			canondata.M{"ids": strings.Join(v.Req.Dependencies, ", ")}))
	}

	// How it is called: distinct scenario calls in order of
	// first appearance — living examples of the product's surface.
	seen := map[string]bool{}
	var calls []string
	for _, sc := range v.Scns {
		if sc.When == nil {
			continue
		}
		cmd := humanCommand(sc.When.Command)
		if cmd == "" || seen[cmd] {
			continue
		}
		seen[cmd] = true
		calls = append(calls, "`"+cmd+"`")
	}
	if len(calls) > 0 {
		fmt.Fprintf(&sb, "\n%s\n\n", canondata.TFor(lang, "features.section.calls"))
		for _, c := range calls {
			fmt.Fprintf(&sb, "- %s\n", c)
		}
	}

	// Behavior: one scenario per line — in words, what the user sees;
	// the scenario's address at the end of the line. Boundary classes (when the
	// author marked them) collect their lines into the "Boundaries" section — the
	// main behavior stays the happy path, not a wall of all lines.
	classified := map[string]bool{}
	if v.Req.Detail != nil {
		for _, dc := range detailClasses(v.Req.Detail) {
			for _, id := range dc.class.Scenarios {
				classified[id] = true
			}
		}
	}
	byID := map[string]canon.Scenario{}
	for _, sc := range v.Scns {
		byID[sc.ID] = sc
	}
	var behavior []canon.Scenario
	for _, sc := range v.Scns {
		if !classified[sc.ID] {
			behavior = append(behavior, sc)
		}
	}
	fmt.Fprintf(&sb, "\n%s\n", canondata.TFor(lang, "features.section.behavior"))
	for _, sc := range behavior {
		fmt.Fprintf(&sb, "- %s\n", scenarioHumanLine(lang, sc))
	}
	for _, sc := range v.Prose {
		fmt.Fprintf(&sb, "- %s\n", canondata.TFor(lang, "features.prose", canondata.M{
			"text": humanEscape(strings.ReplaceAll(*sc.Prose, "\n", " ")),
			"id":   sc.ID,
		}))
	}

	// Boundaries: four behavior classes — the class's lines in living words
	// under the class name, or an honest refusal with a reason.
	if v.Req.Detail != nil {
		fmt.Fprintf(&sb, "\n%s\n", canondata.TFor(lang, "features.section.boundaries"))
		for _, dc := range detailClasses(v.Req.Detail) {
			sb.WriteString(renderDetailBlock(lang, dc.name, dc.class, byID))
		}
	}
	return sb.String(), nil
}

// detailClassRef — one boundary class of the canon: a living name and data.
type detailClassRef struct {
	name  string
	class *canon.DetailClass
}

// detailClasses — the four boundary classes in a stable order.
func detailClasses(d *canon.RequirementDetail) []detailClassRef {
	return []detailClassRef{
		{"invariants", d.Invariants},
		{"edge-cases", d.EdgeCases},
		{"failure-matrix", d.FailureMatrix},
		{"concurrency", d.Concurrency},
	}
}

// renderDetailBlock — a boundary class: its scenarios' lines in living
// words (address at the end of the line) or the refusal reason. A filled
// detail block is always complete (spec validation): a class either carries
// scenarios or a refusal.
func renderDetailBlock(lang, class string, dc *canon.DetailClass, byID map[string]canon.Scenario) string {
	name := canondata.TFor(lang, "features.detail."+class)
	if dc == nil || len(dc.Scenarios) == 0 {
		reason := ""
		if dc != nil {
			reason = dc.Absent
		}
		return "- " + name + ": " + canondata.TFor(lang, "features.detail.absent", canondata.M{
			"reason": reason,
		}) + "\n"
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "- %s:\n", name)
	for _, id := range dc.Scenarios {
		if sc, ok := byID[id]; ok {
			fmt.Fprintf(&sb, "  - %s\n", scenarioHumanLine(lang, sc))
		}
	}
	return sb.String()
}

// humanEscape — values in a human spec stay living words:
// printable unicode passes as is, only control characters are escaped
// (newline, tab, carriage return, other controls, backslash).
// Byte strictness is the machine render's business; this is reading.
func humanEscape(s string) string {
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		b := s[i]
		switch {
		case b >= 0x20 && b != 0x7F && b != '\\':
			sb.WriteByte(b)
		case b == '\\':
			sb.WriteString("\\\\")
		case b == '\n':
			sb.WriteString("\\n")
		case b == '\t':
			sb.WriteString("\\t")
		case b == '\r':
			sb.WriteString("\\r")
		default:
			fmt.Fprintf(&sb, "\\x%02X", b)
		}
	}
	return sb.String()
}

// humanCommand — a call for human reading: an argument that is empty or
// contains spaces is quoted — a boundary like "added an empty
// line" stays visible instead of vanishing into the spaces.
func humanCommand(argv []string) string {
	parts := make([]string, 0, len(argv))
	for _, arg := range argv {
		if arg == "" || strings.ContainsAny(arg, " \t") {
			parts = append(parts, "\""+arg+"\"")
			continue
		}
		parts = append(parts, arg)
	}
	return strings.Join(parts, " ")
}

// scenarioHumanLine — one scenario in living words: the given,
// preliminary calls, the call, and everything the user sees;
// observation conditions are not shown as text — behavior only.
func scenarioHumanLine(lang string, sc canon.Scenario) string {
	var parts []string
	if len(sc.Seed) > 0 || len(sc.Materials) > 0 {
		files := make([]string, 0, len(sc.Seed)+len(sc.Materials))
		for _, sd := range sc.Seed {
			files = append(files, sd.Path)
		}
		for _, m := range sc.Materials {
			files = append(files, m.From)
		}
		parts = append(parts, canondata.TFor(lang, "features.bullet.given", canondata.M{
			"ops": strings.Join(files, "; "),
		}))
	}
	if len(sc.Pre) > 0 {
		cmds := make([]string, 0, len(sc.Pre))
		for _, pre := range sc.Pre {
			cmds = append(cmds, "`"+humanCommand(pre)+"`")
		}
		parts = append(parts, canondata.TFor(lang, "features.bullet.pre", canondata.M{
			"commands": strings.Join(cmds, ", "),
		}))
	}
	if sc.When != nil {
		parts = append(parts, canondata.TFor(lang, "features.bullet.call", canondata.M{
			"command": humanCommand(sc.When.Command),
		}))
	}
	for _, a := range sc.Then {
		parts = append(parts, assertionHumanText(lang, a))
	}
	return strings.Join(parts, ": ") + " [" + sc.ID + "]"
}

// assertionHumanText — an observation in living words: what was printed,
// what was answered, what remained in files. An empty stream under equals
// reads in the same words as absences — a template with an empty value
// does not read at all. The observation vocabulary is closed (stdout,
// stderr, exit-code, file) and validated on input.
func assertionHumanText(lang string, a canon.Assertion) string {
	value := humanEscape(trimOneTrailingNL(a.Value))
	key := "features.out." + a.Observation + "." + a.Condition
	switch a.Observation {
	case "stdout", "stderr":
		if a.Condition == "equals" && a.Value == "" {
			return canondata.TFor(lang, "features.out."+a.Observation+".absent")
		}
		return canondata.TFor(lang, key, canondata.M{"value": value})
	case "exit-code":
		if a.Condition == "equals" && a.Value == "0" {
			return canondata.TFor(lang, "features.out.exit.ok")
		}
		return canondata.TFor(lang, "features.out.exit.fail")
	default: // file
		return canondata.TFor(lang, key, canondata.M{"path": a.Path, "value": value})
	}
}

// trimOneTrailingNL — exactly one trailing newline leaves the
// human-readable value: line output must end with
// a newline, and its escape at the end of the line is noise, not a fact;
// newlines inside the value stay visible.
func trimOneTrailingNL(s string) string {
	return strings.TrimSuffix(s, "\n")
}
