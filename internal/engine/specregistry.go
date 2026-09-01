// The spec registry — the knowledge layer: requirements,
// scenarios and their links as a computable view of the canon (there
// is no second memory — every view is rebuilt from the canon anew,
// the registry is stored nowhere). Registry search is the surface of
// the why verb: `why spec` — the whole registry, `why spec <words>` —
// the features and scenarios where all the words meet. A feature is
// a requirement: the formulation carries the feature name in live
// words, the behavior — its scenarios.
package engine

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/canondata"
)

// featureView — one registry feature: the requirement, its
// executable scenarios, unmeasurable prose and the last check
// outcome per scenario.
type featureView struct {
	Req      canon.Requirement
	Scns     []canon.Scenario
	Prose    []canon.Scenario
	Outcomes map[string]string
}

// featureViews — the whole registry: requirements in identifier
// order, scenarios grouped by their requirement. The spec invariant
// guarantees every requirement at least one scenario (executable
// or prose) and a live scenario link to the requirement.
func (e *Engine) featureViews() ([]featureView, error) {
	reqs, err := e.Store.Requirements()
	if err != nil {
		return nil, err
	}
	scns, err := e.Store.Scenarios()
	if err != nil {
		return nil, err
	}
	chk, err := e.Store.Checks()
	if err != nil {
		return nil, err
	}
	outcomeOf := map[string]string{}
	for _, c := range chk {
		outcomeOf[c.Scenario] = c.Outcome
	}
	views := make([]featureView, 0, len(reqs))
	byReq := map[string]int{}
	for _, r := range reqs {
		byReq[r.ID] = len(views)
		views = append(views, featureView{Req: r, Outcomes: map[string]string{}})
	}
	for _, sc := range scns {
		i, ok := byReq[sc.Requirement]
		if !ok {
			continue
		}
		if sc.Prose != nil {
			views[i].Prose = append(views[i].Prose, sc)
			continue
		}
		views[i].Scns = append(views[i].Scns, sc)
		if out := outcomeOf[sc.ID]; out != "" {
			views[i].Outcomes[sc.ID] = out
		}
	}
	return views, nil
}

// reqStatusWord — the requirement status in live words: the machine
// enum code does not ride into the human block.
func reqStatusWord(lang, status string) string {
	switch status {
	case canon.ReqDraft:
		return canondata.TFor(lang, "features.status.draft")
	case canon.ReqAccepted:
		return canondata.TFor(lang, "features.status.accepted")
	case canon.ReqRealized:
		return canondata.TFor(lang, "features.status.realized")
	case canon.ReqQuarantine:
		return canondata.TFor(lang, "features.status.quarantine")
	case canon.ReqCancelled:
		return canondata.TFor(lang, "features.status.cancelled")
	}
	return status
}

// featureBlob — the feature's search field: everything a human
// searches knowledge by — the formulation, scenario summaries,
// invocations, expected texts, prose. Case is folded: words are
// found regardless of how they are typed.
func featureBlob(v featureView) string {
	var sb strings.Builder
	sb.WriteString(strings.ToLower(v.Req.Formulation))
	if v.Req.SpecDoc != nil {
		sb.WriteByte(' ')
		sb.WriteString(strings.ToLower(v.Req.SpecDoc.Text))
	}
	for _, sc := range v.Scns {
		sb.WriteByte(' ')
		sb.WriteString(scenarioBlob(sc))
	}
	for _, sc := range v.Prose {
		sb.WriteByte(' ')
		sb.WriteString(scenarioBlob(sc))
	}
	return sb.String()
}

// scenarioBlob — the scenario's search field.
func scenarioBlob(sc canon.Scenario) string {
	var sb strings.Builder
	sb.WriteString(strings.ToLower(sc.Summary))
	if sc.When != nil {
		sb.WriteByte(' ')
		sb.WriteString(strings.ToLower(strings.Join(sc.When.Command, " ")))
	}
	for _, a := range sc.Then {
		sb.WriteByte(' ')
		sb.WriteString(strings.ToLower(a.Value))
	}
	if sc.Prose != nil {
		sb.WriteByte(' ')
		sb.WriteString(strings.ToLower(*sc.Prose))
	}
	return sb.String()
}

// matchesQuery — a deterministic match of all query words against
// the field: every word occurs as a substring.
func matchesQuery(blob string, tokens []string) bool {
	for _, tok := range tokens {
		if !strings.Contains(blob, tok) {
			return false
		}
	}
	return true
}

// specRegistryLine — the registry feature line: the identifier as
// the address, status and name in live words, the count of
// scenarios and green checks.
func specRegistryLine(lang string, v featureView) string {
	green := 0
	for _, sc := range v.Scns {
		if v.Outcomes[sc.ID] == canon.CheckGreen {
			green++
		}
	}
	return canondata.TFor(lang, "why.spec.feature", canondata.M{
		"id": v.Req.ID, "status": reqStatusWord(lang, v.Req.Status),
		"name":      firstLineOf(v.Req.Formulation),
		"scenarios": strconv.Itoa(len(v.Scns) + len(v.Prose)),
		"green":     strconv.Itoa(green),
	})
}

// whySpec — the registry search surface: without words the whole
// registry, with words the features and scenarios where all the
// words meet. The language is the wish's language: a human reads
// the block.
func (e *Engine) whySpec(query string) (string, error) {
	state, err := e.Store.State()
	if err != nil {
		return "", err
	}
	lang := state.Language
	views, err := e.featureViews()
	if err != nil {
		return "", err
	}
	totalScns := 0
	totalProse := 0
	for _, v := range views {
		totalScns += len(v.Scns)
		totalProse += len(v.Prose)
	}

	query = strings.TrimSpace(query)
	var sb strings.Builder
	if query == "" {
		fmt.Fprintf(&sb, "%s\n", canondata.TFor(lang, "why.spec.head", canondata.M{
			"features":  strconv.Itoa(len(views)),
			"scenarios": strconv.Itoa(totalScns),
			"prose":     strconv.Itoa(totalProse),
			"render":    filepath.ToSlash(filepath.Join(canon.Dir(""), specsDirName)),
		}))
		for _, v := range views {
			fmt.Fprintf(&sb, "%s\n", specRegistryLine(lang, v))
		}
		return strings.TrimRight(sb.String(), "\n"), nil
	}

	tokens := strings.Fields(strings.ToLower(query))
	featHits, scnHits := 0, 0
	var body strings.Builder
	for _, v := range views {
		if !matchesQuery(featureBlob(v), tokens) {
			continue
		}
		featHits++
		fmt.Fprintf(&body, "%s\n", specRegistryLine(lang, v))
		for _, sc := range v.Scns {
			if !matchesQuery(scenarioBlob(sc), tokens) {
				continue
			}
			scnHits++
			fmt.Fprintf(&body, "%s\n", canondata.TFor(lang, "why.spec.scenario", canondata.M{
				"id": sc.ID, "line": humanCommand(sc.When.Command),
			}))
		}
	}
	if featHits == 0 {
		return canondata.TFor(lang, "why.spec.search-none", canondata.M{
			"query": query, "features": strconv.Itoa(len(views)),
			"scenarios": strconv.Itoa(totalScns),
		}), nil
	}
	fmt.Fprintf(&sb, "%s\n", canondata.TFor(lang, "why.spec.search-head", canondata.M{
		"query": query, "features": strconv.Itoa(featHits),
		"scenarios": strconv.Itoa(scnHits),
	}))
	sb.WriteString(body.String())
	return strings.TrimRight(sb.String(), "\n"), nil
}
