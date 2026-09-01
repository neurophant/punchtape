// The why verb: where a number comes from, its breakdown into
// components, the ceiling and how to raise it.
package engine

import (
	"fmt"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/delta"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// Why — "why": the topics price, stage, latency, or an entity ID.
func (e *Engine) Why(topic string) (string, error) {
	started := time.Now()
	text, err := e.whyTopic(topic)
	e.logVerb("why", "ok", map[string]string{"topic": topic},
		time.Since(started).Milliseconds(), 0)
	return text, err
}

// whyAmend — the held amend verbatim: exactly what the human must see
// before asserting it (a blind assertion is not an assertion).
func (e *Engine) whyAmend() string {
	state, err := e.Store.State()
	if err != nil || state.PendingAmend == nil {
		return canondata.T("why.amend.none")
	}
	return canondata.T("why.amend.head", canondata.M{
		"key":     state.PendingAmend.Key,
		"raw":     indentLines(strings.TrimRight(state.PendingAmend.Raw, "\n")),
		"changes": amendChangesBlock(state.PendingAmend.Changes, state.PendingAmend.Plan),
	})
}

// whyDrift — instance drift against the acceptance moment: the
// verdict is unreconciled, the cause and the action are named from
// the machine's own data.
func (e *Engine) whyDrift() string {
	drifted, oldD, newD := e.acceptanceDrift()
	if !drifted {
		if oldD == "" && newD == "" {
			return canondata.T("why.drift.none")
		}
		return canondata.T("why.drift.fresh", canondata.M{"digest": newD})
	}
	return canondata.T("why.drift.drifted", canondata.M{
		"old": oldD, "new": newD,
	})
}

// whyAcceptance — the acceptance moment: outcomes and the digest of
// the state on which the check set was finalized.
func (e *Engine) whyAcceptance() string {
	last := lastAcceptanceEntry(e)
	if last == nil {
		return canondata.T("why.acceptance.none")
	}
	return canondata.T("why.acceptance.head", canondata.M{
		"outcome": last.Details["outcome"],
		"digest":  last.Details["digest"],
		"green":   last.Details["green"],
		"red":     last.Details["red"],
		"total":   last.Details["checks"],
		"env":     last.Details["env"],
	})
}

// whySurface — the current verification surface: the declared
// artifact (or direct submission) and its digest on the current
// instance.
func (e *Engine) whySurface() string {
	artifact := e.surfaceArtifact()
	mode := canondata.T("why.surface.direct")
	if artifact != "" {
		mode = canondata.T("why.surface.artifact", canondata.M{"artifact": artifact})
	}
	chk, err := e.Store.Checks()
	if err != nil {
		return mode
	}
	digest, ok := surfaceDigest(e, chk)
	if !ok {
		return mode + "\n" + canondata.T("why.surface.no-digest")
	}
	return mode + "\n" + canondata.T("why.surface.digest", canondata.M{
		"digest": shortHex(digest),
	})
}

// whyCoverage — the completeness gaps the machine names: command
// families of the wish without a single executable scenario (exactly
// what the coverage gate holds on) and active waives with reasons. A
// gap can only be waived (kind: waive) once it is named — the surface
// must name them itself, from data, not wait for the question.
func (e *Engine) whyCoverage() (string, error) {
	state, err := e.Store.State()
	if err != nil {
		return "", err
	}
	scns, err := e.Store.Scenarios()
	if err != nil {
		return "", err
	}
	gaps := uncoveredFamilies(state, scns)
	var sb strings.Builder
	if len(gaps) == 0 {
		sb.WriteString(canondata.T("why.coverage.none"))
	} else {
		sb.WriteString(canondata.T("why.coverage.head", canondata.M{
			"n": strconv.Itoa(len(gaps)),
		}))
		for _, g := range gaps {
			sb.WriteString("\n")
			sb.WriteString(canondata.T("why.coverage.gap", canondata.M{"family": g}))
		}
	}
	if len(state.Waives) > 0 {
		sb.WriteString("\n")
		sb.WriteString(canondata.T("why.coverage.waives-head", canondata.M{
			"n": strconv.Itoa(len(state.Waives)),
		}))
		for _, w := range state.Waives {
			sb.WriteString("\n")
			sb.WriteString(canondata.T("why.coverage.waive", canondata.M{
				"rule": w.Rule, "target": w.Target, "reason": w.Reason,
			}))
		}
	}
	return sb.String(), nil
}

// WhyNotFound — the why verb's topic does not exist: rejected input
// (rc 1), not a machine failure (rc 2).
type WhyNotFound struct{ ID string }

func (e *WhyNotFound) Error() string {
	return canondata.T("why.not-found", canondata.M{"id": e.ID})
}

// whyTopic — the verb's own work, without meter accounting.
func (e *Engine) whyTopic(topic string) (string, error) {
	switch {
	case topic == "" || topic == "topics":
		return whyTopics(), nil
	case topic == "price":
		return e.whyPrice()
	case topic == "stage":
		return e.whyStage()
	case topic == "derived":
		return e.whyDerived(), nil
	case topic == "red":
		return e.whyRed()
	case topic == "latency":
		return e.whyLatency(), nil
	case topic == "passport":
		return e.whyPassport()
	case topic == "diary":
		return e.WhyDiary(), nil
	case topic == "impact":
		return e.whyImpact()
	case topic == "knowledge":
		return e.whyKnowledge()
	case topic == "verifications":
		return e.whyVerifications()
	case topic == "amend":
		return e.whyAmend(), nil
	case topic == "drift":
		return e.whyDrift(), nil
	case topic == "acceptance":
		return e.whyAcceptance(), nil
	case topic == "surface":
		return e.whySurface(), nil
	case topic == "coverage":
		return e.whyCoverage()
	case topic == "spec" || strings.HasPrefix(topic, "spec "):
		return e.whySpec(strings.TrimSpace(strings.TrimPrefix(topic, "spec")))
	case topic == "attention":
		return e.whyAttention(), nil
	case topic == "retro":
		return e.WhyRetro(), nil
	case topic == "handoff":
		return e.WhyHandoff()
	case topic == "hypotheses":
		return whyHypotheses(), nil
	case topic == "compatibility":
		return whyCompatibility(), nil
	case topic == "freeze":
		// The same key the slot prints: the concept's truth must
		// not diverge from the slot's truth.
		return canondata.T("slot.implement.composition-freeze"), nil
	case topic == "repair":
		return canondata.T("why.concept.repair"), nil
	case topic == "regenerate":
		return canondata.T("why.concept.regenerate"), nil
	case topic == "rerun":
		return canondata.T("why.concept.rerun"), nil
	case topic == "fragment" || strings.HasPrefix(topic, "fragment "):
		return e.whyFragment(strings.TrimSpace(strings.TrimPrefix(topic, "fragment")))
	default:
		return e.whyEntity(normalizeEntityTopic(topic))
	}
}

// entityKindWords — colloquial names of entity kinds: the contract
// names identifiers with the kind word ("card REQ-1"), the surface
// accepts both forms — the kind word is stripped, the identifier
// remains.
var entityKindWords = map[string]bool{
	"requirement": true, "scenario": true, "card": true, "check": true,
	"contract": true, "decision": true, "kb": true,
}

// normalizeEntityTopic — "card REQ-001" → "REQ-001": an entity kind
// word before the identifier does not change addressing.
func normalizeEntityTopic(topic string) string {
	word, rest, found := strings.Cut(topic, " ")
	if found && entityKindWords[strings.ToLower(word)] {
		return strings.TrimSpace(rest)
	}
	return topic
}

func whyTopics() string {
	return canondata.T("why.topics")
}

// whyPrice breaks the current slot's price down into components: the
// origin of each number, the rate, the reply floor and the ceiling.
// The machine has no complete counters — every figure is named as an
// estimate with its origin; the machine has no right to lie with
// figures.
func (e *Engine) whyPrice() (string, error) {
	state, err := e.Store.State()
	if err != nil {
		return "", err
	}
	text, _, err := e.renderNext(state, "")
	if err != nil {
		return "", err
	}
	lines := strings.Split(text, "\n")
	var action, example, context strings.Builder
	section := ""
	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "ACTION:"):
			section = "a"
			action.WriteString(line)
		case strings.HasPrefix(line, "EXAMPLE:"):
			section = "e"
		case strings.HasPrefix(line, "CONTEXT:"):
			section = "c"
		case strings.HasPrefix(line, "GATES"):
			section = ""
		default:
			switch section {
			case "e":
				example.WriteString(line)
				example.WriteString("\n")
			case "c":
				context.WriteString(line)
				context.WriteString("\n")
			}
		}
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s\n", canondata.T("why.price.head", canondata.M{
		"action": action.String(),
	}))
	fmt.Fprintf(&sb, "%s\n", canondata.T("why.price.origin", canondata.M{
		"example-bytes":  strconv.Itoa(len([]byte(example.String()))),
		"context-bytes":  strconv.Itoa(len([]byte(context.String()))),
		"example-tokens": strconv.Itoa(tokenEstimate(example.String())),
		"context-tokens": strconv.Itoa(tokenEstimate(context.String())),
	}))
	if floor := tokenEstimate(delta.Skeleton(kindForSlot(state.Stage, action.String()))); floor > 0 {
		fmt.Fprintf(&sb, "%s\n", canondata.T("why.price.floor", canondata.M{
			"floor": strconv.Itoa(floor),
		}))
	}
	fmt.Fprintf(&sb, "%s\n", canondata.T("why.price.overhead"))
	fmt.Fprintf(&sb, "%s\n", canondata.T("why.price.ceiling"))
	fmt.Fprintf(&sb, "%s", canondata.T("why.price.raise"))
	return sb.String(), nil
}

// kindForSlot — the delta kind expected from the current action: used
// to estimate the reply floor with the format skeleton.
func kindForSlot(stage, action string) string {
	switch stage {
	case canon.StageIntake:
		return delta.KindIntent
	case canon.StageSpec:
		return delta.KindSpec
	case canon.StageACCompile:
		return delta.KindAssert
	case canon.StageImplement:
		if strings.Contains(action, "fix delta") {
			return delta.KindFix
		}
		return delta.KindCode
	}
	return ""
}

// whyStage — meters of the current stage from the ledger: machine
// facts only.
func (e *Engine) whyStage() (string, error) {
	state, err := e.Store.State()
	if err != nil {
		return "", err
	}
	var applied, wall, calls, tokens int64
	for _, en := range e.Ledger.All() {
		if en.Stage != state.Stage {
			continue
		}
		switch en.Event {
		case "apply":
			applied++
			wall += en.WallMs
		case "verb":
			calls++
		}
		// The caller-side counter lives in different entries: apply/executor
		// is a driver run, verb submit is a manual submission with a
		// "# TOKENS" tail; each delta is counted once.
		if en.Tokens != nil {
			tokens += *en.Tokens
		}
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s\n", canondata.T("why.stage.head", canondata.M{"stage": state.Stage}))
	fmt.Fprintf(&sb, "%s\n", canondata.T("why.stage.deltas", canondata.M{
		"deltas": strconv.FormatInt(applied, 10),
	}))
	fmt.Fprintf(&sb, "%s\n", canondata.T("why.stage.wall", canondata.M{
		"wall": strconv.FormatInt(wall, 10),
	}))
	fmt.Fprintf(&sb, "%s\n", canondata.T("why.stage.calls", canondata.M{
		"calls": strconv.FormatInt(calls, 10),
	}))
	if tokens > 0 {
		fmt.Fprintf(&sb, "%s", canondata.T("why.stage.tokens", canondata.M{
			"tokens": strconv.FormatInt(tokens, 10),
		}))
	} else {
		fmt.Fprintf(&sb, "%s", canondata.T("why.stage.tokens-no-data"))
	}
	return sb.String(), nil
}

// whyDerived — the verdict's derived cases unfolded: a summary by
// kind and a full line for each red one (the verdict carries only the
// count).
func (e *Engine) whyDerived() string {
	state, err := e.Store.State()
	if err != nil || state.Stage != canon.StageDeliver {
		return canondata.T("why.derived.not-run")
	}
	scns, err := e.Store.Scenarios()
	if err != nil {
		return canondata.T("why.derived.unavailable")
	}
	chk, err := e.Store.Checks()
	if err != nil {
		return canondata.T("why.derived.unavailable")
	}
	d, ok := surfaceDigest(e, chk)
	if !ok {
		return canondata.T("why.derived.none")
	}
	rep := loadOrRunDerived(e, scns, d)
	if rep == nil {
		return canondata.T("why.derived.unavailable")
	}
	byKind := map[string][2]int{} // kind → [green, red]
	for _, o := range rep.Outs {
		c := byKind[o.Kind]
		if o.Green {
			c[0]++
		} else {
			c[1]++
		}
		byKind[o.Kind] = c
	}
	kinds := make([]string, 0, len(byKind))
	for k := range byKind {
		kinds = append(kinds, k)
	}
	sort.Strings(kinds)
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s\n", canondata.T("why.derived.head", canondata.M{
		"count":  strconv.Itoa(len(rep.Outs)),
		"budget": strconv.Itoa(canondata.Limit("derive.budget")),
	}))
	for _, k := range kinds {
		c := byKind[k]
		fmt.Fprintf(&sb, "%s\n", canondata.T("why.derived.kind", canondata.M{
			"kind": k, "green": strconv.Itoa(c[0]), "red": strconv.Itoa(c[1]),
		}))
	}
	for _, o := range rep.Outs {
		if !o.Green {
			fmt.Fprintf(&sb, "%s\n", canondata.T("why.derived.red", canondata.M{
				"parent": o.Parent, "kind": o.Kind, "reason": o.Reason,
			}))
		}
	}
	return strings.TrimRight(sb.String(), "\n")
}

// whyPassport — the computed view of the passport: the boundary
// registry with honest freshness (reconciliation is computed on the
// fly, nothing is stored), conventions and an inventory with
// anti-blindness — the unassigned is named, there is no silence.
func (e *Engine) whyPassport() (string, error) {
	boundaries, err := e.Store.Boundaries()
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
	cards, err := e.Store.Cards()
	if err != nil {
		return "", err
	}
	conv, err := e.Store.Conventions()
	if err != nil {
		return "", err
	}
	convDigest := e.Store.ConventionsDigest()

	var sb strings.Builder
	conventionsLine := canondata.T("why.passport.conventions-empty")
	if conv != nil {
		var parts []string
		for _, section := range []struct {
			name string
			rec  *canon.Recipe
		}{{"build", conv.Build}, {"types", conv.Types}, {"lint", conv.Lint}} {
			if section.rec == nil {
				continue
			}
			parts = append(parts, section.name+" ("+strconv.Itoa(len(section.rec.Commands))+" command(s))")
		}
		if len(parts) > 0 {
			conventionsLine = canondata.T("why.passport.conventions-declared", canondata.M{
				"parts": strings.Join(parts, ", "),
			})
		}
	}
	fmt.Fprintf(&sb, "%s\n", canondata.T("why.passport.head", canondata.M{
		"count":       strconv.Itoa(len(boundaries)),
		"conventions": conventionsLine,
	}))

	scope, _ := e.boundaryScope(cards, scns)
	trustTop := strconv.Itoa(canondata.TrustTop())
	// Drift direction: the cure depends on which is
	// newer. Code moved away from the verdict (scope-hash/conventions
	// diverged, a file disappeared) — re-proof by the machine; the
	// passport ahead of the code (red checks) — cards are reopened,
	// the code catches up. Both cures are already machine-side — the
	// line names them explicitly.
	chk, err := e.Store.Checks()
	if err != nil {
		return "", err
	}
	outcomeByScn := map[string]string{}
	for _, c := range chk {
		outcomeByScn[c.Scenario] = c.Outcome
	}
	redByBoundary := map[string]int{}
	for _, sc := range scns {
		if outcomeByScn[sc.ID] == canon.CheckRed {
			redByBoundary[sc.Requirement]++
		}
	}
	for _, bd := range boundaries {
		hash, missing := e.scopeHash(scope[bd.ID])
		fresh := canondata.T("why.passport.fresh-unchecked")
		switch {
		case bd.Verdict == nil:
			// not verified
		case len(missing) > 0:
			fresh = canondata.T("why.passport.fresh-missing")
		case bd.Verdict.ScopeHash != hash || bd.Verdict.ConventionsDigest != convDigest:
			fresh = canondata.T("why.passport.fresh-drifted")
		default:
			fresh = canondata.T("why.passport.fresh-fresh")
		}
		weight, onLadder := canondata.TrustWeight(bd.Status)
		if !onLadder {
			weight = 0
		}
		fmt.Fprintf(&sb, "%s\n", canondata.T("why.passport.boundary", canondata.M{
			"id": bd.ID, "status": bd.Status,
			"weight": strconv.Itoa(weight), "top": trustTop,
			"domain": strings.Join(bd.Domain, ", "),
			"files":  strconv.Itoa(len(bd.ScopeFiles)),
			"fresh":  fresh,
		}))
		if fresh == canondata.T("why.passport.fresh-drifted") || fresh == canondata.T("why.passport.fresh-missing") {
			fmt.Fprintf(&sb, "%s\n", canondata.T("why.passport.drift-code"))
		} else if redByBoundary[bd.ID] > 0 {
			fmt.Fprintf(&sb, "%s\n", canondata.T("why.passport.drift-spec"))
		}
		if bd.Verdict != nil {
			fmt.Fprintf(&sb, "%s\n", canondata.T("why.passport.verdict", canondata.M{
				"scope": shortHex(bd.Verdict.ScopeHash), "conv": shortHex(bd.Verdict.ConventionsDigest),
				"tx": shortHex(bd.Verdict.JournalTx),
			}))
		}
		if bd.Attribution != "" {
			fmt.Fprintf(&sb, "%s\n", canondata.T("why.passport.attribution", canondata.M{
				"note": bd.Attribution,
			}))
		}
	}

	// Inventory: features bound to boundaries, the unassigned — with a
	// reason.
	prose := 0
	for _, sc := range scns {
		if sc.Prose != nil {
			prose++
		}
	}
	// Unassigned code: a journal file outside all scopes is impossible
	// by construction (card ownership and the shared boundary cover
	// everything) — we count honestly instead of believing.
	unassigned := e.codeFilesOfJournal()
	covered := map[string]bool{}
	for _, files := range scope {
		for _, f := range files {
			covered[f] = true
		}
	}
	n := 0
	for _, f := range unassigned {
		if !covered[f] {
			n++
		}
	}
	fmt.Fprintf(&sb, "%s\n", canondata.T("why.passport.inventory", canondata.M{
		"features":   strconv.Itoa(len(reqs)),
		"prose":      strconv.Itoa(prose),
		"unassigned": strconv.Itoa(n),
	}))

	// Piecemeal population and backfill: known behavior without a
	// boundary — probe observations without a scenario (completeness:
	// code→passport); detail-minimum — boundaries without a behavior
	// classes block; formalization priority — unverified boundaries by
	// the change frequency of their scope files per the journal
	// (frequency × risk), descending.
	scenarioIDs := map[string]bool{}
	for _, sc := range scns {
		scenarioIDs[sc.ID] = true
	}
	uncovered, _ := e.uncoveredProbes(scns)
	redChecks := 0
	for _, c := range chk {
		if c.Outcome == canon.CheckRed {
			redChecks++
		}
	}
	fmt.Fprintf(&sb, "%s\n", canondata.T("why.passport.inventory-probes", canondata.M{
		"probes": strconv.Itoa(len(uncovered)),
		"red":    strconv.Itoa(redChecks),
	}))
	noDetail := 0
	for _, r := range reqs {
		if r.Detail == nil {
			noDetail++
		}
	}
	fmt.Fprintf(&sb, "%s\n", canondata.T("why.passport.inventory-detail", canondata.M{
		"boundaries": strconv.Itoa(len(boundaries)),
		"nodetail":   strconv.Itoa(noDetail),
	}))
	if prio := e.backfillPriority(boundaries, scope); len(prio) > 0 {
		for _, line := range prio {
			fmt.Fprintf(&sb, "%s\n", line)
		}
	}
	return sb.String(), nil
}

// backfillPriority — boundary formalization priority:
// unverified (status below verified — risk) boundaries by the change
// frequency of their scope files per the journal (frequency),
// descending by frequency.
func (e *Engine) backfillPriority(boundaries []canon.Boundary, scope map[string][]string) []string {
	type row struct {
		id      string
		status  string
		weight  int
		changes int
	}
	// Frequency: how many code/fix transactions touched the boundary's
	// files.
	fileChanges := map[string]int{}
	for _, en := range e.Journal.All() {
		if en.DeltaKind != deltaKindCode && en.DeltaKind != deltaKindFix {
			continue
		}
		touched := map[string]bool{}
		for _, ef := range en.Effects {
			if !ef.Delete {
				touched[ef.Path] = true
			}
		}
		for f := range touched {
			fileChanges[f]++
		}
	}
	// The risk threshold is the "verified" weight from the ladder
	// data: only unverified boundaries carry machine formalization
	// priority.
	verifiedW, _ := canondata.TrustWeight("verified")
	var rows []row
	for _, bd := range boundaries {
		w, ok := canondata.TrustWeight(bd.Status)
		if !ok || w >= verifiedW {
			continue // verified boundaries need no formalization
		}
		changes := 0
		for _, f := range scope[bd.ID] {
			changes += fileChanges[f]
		}
		rows = append(rows, row{bd.ID, bd.Status, w, changes})
	}
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].changes > rows[j].changes })
	var out []string
	for _, r := range rows {
		out = append(out, canondata.T("why.passport.backfill-priority", canondata.M{
			"id": r.id, "status": r.status, "changes": strconv.Itoa(r.changes),
		}))
	}
	return out
}

// whyLatency — p95 of the core verbs from the ledger; the submit wall
// is reduced by the honest time of run work (machine overhead, the
// same computation as the verdict's latency line).
func (e *Engine) whyLatency() string {
	byVerb := map[string][]int64{}
	for _, en := range e.Ledger.All() {
		if en.Event != "verb" {
			continue
		}
		name := en.Details["name"]
		if name == "" {
			continue
		}
		wall := en.WallMs
		if name == "submit" {
			wall -= atoiOrZero64(en.Details["runs-ms"])
			if wall < 0 {
				wall = 0
			}
		}
		byVerb[name] = append(byVerb[name], wall)
	}
	if len(byVerb) == 0 {
		return canondata.T("why.latency-none")
	}
	names := make([]string, 0, len(byVerb))
	for name := range byVerb {
		names = append(names, name)
	}
	sort.Strings(names)
	var sb strings.Builder
	sb.WriteString(canondata.T("why.latency-head"))
	sb.WriteString("\n")
	for _, name := range names {
		xs := byVerb[name]
		sort.Slice(xs, func(i, j int) bool { return xs[i] < xs[j] })
		rank := (95*len(xs) + 99) / 100
		fmt.Fprintf(&sb, "%s\n", canondata.T("why.latency-verb", canondata.M{
			"name": name, "p95": strconv.FormatInt(xs[rank-1], 10),
			"calls": strconv.Itoa(len(xs)),
		}))
	}
	return strings.TrimRight(sb.String(), "\n")
}

// whyFragment — a scenario's observations as data: the canon itself
// emits its bytes as ready blocks for match (copy verbatim, do not
// retype the card). Header + YAML: the command and the observation
// list.
func (e *Engine) whyFragment(id string) (string, error) {
	if !strings.HasPrefix(id, "SCN-") {
		return "", fmt.Errorf("%s", canondata.T("why.fragment.usage"))
	}
	scns, err := e.Store.Scenarios()
	if err != nil {
		return "", err
	}
	for _, sc := range scns {
		if sc.ID != id {
			continue
		}
		doc := fragmentDoc{Scenario: sc.ID, Then: sc.Then}
		if sc.When != nil {
			doc.When = sc.When.Command
		}
		data, err := yamlio.Marshal(doc)
		if err != nil {
			return "", err
		}
		return canondata.T("why.fragment.head", canondata.M{
			"id": sc.ID,
		}) + "\n" + string(data), nil
	}
	return "", &WhyNotFound{ID: id}
}

// fragmentDoc — the output form of a scenario's observations: a
// minimal projection of the canon, Marshall of the observations
// carries the full form (match compares by decoded form — compact and
// full are equal).
type fragmentDoc struct {
	Scenario string            `yaml:"scenario"`
	When     []string          `yaml:"when,omitempty"`
	Then     []canon.Assertion `yaml:"then"`
}

// whyEntity — the full card of an entity by ID.
func (e *Engine) whyEntity(id string) (string, error) {
	switch {
	case strings.HasPrefix(id, "REQ-"):
		reqs, err := e.Store.Requirements()
		if err != nil {
			return "", err
		}
		for _, r := range reqs {
			if r.ID == id {
				data, err := yamlio.Marshal(r)
				if err != nil {
					return "", err
				}
				return string(data), nil
			}
		}
	case strings.HasPrefix(id, "SCN-"):
		scns, err := e.Store.Scenarios()
		if err != nil {
			return "", err
		}
		for _, sc := range scns {
			if sc.ID == id {
				data, err := yamlio.Marshal(sc)
				if err != nil {
					return "", err
				}
				return string(data), nil
			}
		}
	case strings.HasPrefix(id, "CRD-"):
		cards, err := e.Store.Cards()
		if err != nil {
			return "", err
		}
		for _, c := range cards {
			if c.ID == id {
				data, err := yamlio.Marshal(c)
				if err != nil {
					return "", err
				}
				return string(data), nil
			}
		}
	case strings.HasPrefix(id, "IFC-"):
		contracts, err := e.Store.Contracts()
		if err != nil {
			return "", err
		}
		for _, c := range contracts {
			if c.ID == id {
				data, err := yamlio.Marshal(c)
				if err != nil {
					return "", err
				}
				return string(data), nil
			}
		}
	case strings.HasPrefix(id, "TST-"):
		chks, err := e.Store.Checks()
		if err != nil {
			return "", err
		}
		for _, c := range chks {
			if c.ID == id {
				data, err := yamlio.Marshal(c)
				if err != nil {
					return "", err
				}
				return string(data), nil
			}
		}
	case strings.HasPrefix(id, "DEC-"):
		decisions, err := e.Store.Decisions()
		if err != nil {
			return "", err
		}
		for _, d := range decisions {
			if d.ID == id {
				data, err := yamlio.Marshal(d)
				if err != nil {
					return "", err
				}
				return string(data), nil
			}
		}
	case strings.HasPrefix(id, "KB-"):
		entries, err := e.Store.KbEntries()
		if err != nil {
			return "", err
		}
		for _, kb := range entries {
			if kb.ID == id {
				data, err := yamlio.Marshal(kb)
				if err != nil {
					return "", err
				}
				return string(data), nil
			}
		}
	}
	return "", &WhyNotFound{ID: id}
}

// whyImpact — the reverse index file→boundaries: what an edit will
// touch, before the edit itself. A computed projection of boundary
// scopes from the journal; no code files — an honest empty line.
func (e *Engine) whyImpact() (string, error) {
	cards, err := e.Store.Cards()
	if err != nil {
		return "", err
	}
	scns, err := e.Store.Scenarios()
	if err != nil {
		return "", err
	}
	scope, _ := e.boundaryScope(cards, scns)
	owners := map[string][]string{}
	for _, bd := range sortedBoundaryIDs(scope) {
		for _, f := range scope[bd] {
			owners[f] = mergeUnique(owners[f], []string{bd})
		}
	}
	files := e.codeFilesOfJournal()
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s\n", canondata.T("why.impact.head", canondata.M{
		"files": strconv.Itoa(len(files)),
	}))
	for _, f := range files {
		fmt.Fprintf(&sb, "%s\n", canondata.T("why.impact.file", canondata.M{
			"file": f, "boundaries": strings.Join(owners[f], ", "),
		}))
	}
	return strings.TrimRight(sb.String(), "\n"), nil
}

// sortedBoundaryIDs — a stable order of scope keys.
func sortedBoundaryIDs(scope map[string][]string) []string {
	out := make([]string, 0, len(scope))
	for id := range scope {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

// whyKnowledge — the instance's knowledge about the product:
// the Decision Log with dates, the forbidden zone and the active kb
// with sources; superseded kb entries are shown marked.
func (e *Engine) whyKnowledge() (string, error) {
	decisions, err := e.Store.Decisions()
	if err != nil {
		return "", err
	}
	outofscope, err := e.Store.OutOfScope()
	if err != nil {
		return "", err
	}
	kb, err := e.Store.KbEntries()
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s\n", canondata.T("why.knowledge.head", canondata.M{
		"decisions":  strconv.Itoa(len(decisions)),
		"outofscope": strconv.Itoa(len(outofscope)),
		"kb":         strconv.Itoa(len(kb)),
	}))
	for _, d := range decisions {
		fmt.Fprintf(&sb, "%s\n", canondata.T("why.knowledge.decision", canondata.M{
			"id": d.ID, "decision": d.Decision, "rationale": d.Rationale,
			"at": d.At.Format("2006-01-02"),
		}))
	}
	for _, ex := range outofscope {
		fmt.Fprintf(&sb, "%s\n", canondata.T("why.knowledge.outofscope", canondata.M{
			"what": ex.What, "why": ex.Why,
		}))
	}
	dead := kbSupersededBy(kb)
	for _, en := range kb {
		line := canondata.T("why.knowledge.kb", canondata.M{
			"id": en.ID, "when": strings.Join(en.When, " "),
			"do": en.Do, "source": en.Source,
		})
		if dead[en.ID] {
			line = canondata.T("why.knowledge.kb-superseded", canondata.M{"line": line})
		}
		fmt.Fprintf(&sb, "%s\n", line)
	}
	return strings.TrimRight(sb.String(), "\n"), nil
}

// whyRed — the full list of red checks with reasons: the slot trims
// the red list by the line budget ("...and N more"), the tail must be
// available through the verb — machine data, cheap, without re-runs.
func (e *Engine) whyRed() (string, error) {
	chk, err := e.Store.Checks()
	if err != nil {
		return "", err
	}
	red := []canon.Check{}
	for _, c := range chk {
		if c.Outcome == canon.CheckRed {
			red = append(red, c)
		}
	}
	if len(red) == 0 {
		return canondata.T("why.red.none"), nil
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s\n", canondata.T("why.red.head", canondata.M{
		"count": strconv.Itoa(len(red)),
	}))
	for _, c := range red {
		fmt.Fprintf(&sb, "%s\n", canondata.T("why.red.row", canondata.M{
			"id": c.ID, "scn": c.Scenario, "reason": c.Reason,
		}))
	}
	return sb.String(), nil
}

// whyVerifications — the verification log: by what, by whom and when
// each boundary was reconciled — a projection of verdicts and
// journal/ledger anchors, nothing is stored. By whom — always the
// machine (the submit gate transaction); when — the anchor
// transaction's time.
func (e *Engine) whyVerifications() (string, error) {
	boundaries, err := e.Store.Boundaries()
	if err != nil {
		return "", err
	}
	txTime := map[string]string{}
	for _, en := range e.Journal.All() {
		txTime[en.Transaction] = en.RecordedAt.UTC().Format("2006-01-02 15:04")
	}
	var acceptanceOut, acceptanceChecks, acceptanceGreen string
	for _, en := range e.Ledger.All() {
		if en.Event == "gate" && en.Details["name"] == "suite" && en.Details["scope"] == "acceptance" {
			acceptanceOut = en.Details["outcome"]
			acceptanceChecks = en.Details["checks"]
			acceptanceGreen = en.Details["green"]
		}
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s\n", canondata.T("why.verifications.head", canondata.M{
		"count": strconv.Itoa(len(boundaries)),
	}))
	for _, bd := range boundaries {
		if bd.Verdict == nil {
			fmt.Fprintf(&sb, "%s\n", canondata.T("why.verifications.unverified", canondata.M{
				"id": bd.ID,
			}))
			continue
		}
		fmt.Fprintf(&sb, "%s\n", canondata.T("why.verifications.row", canondata.M{
			"id": bd.ID, "status": bd.Status,
			"scope": shortHex(bd.Verdict.ScopeHash),
			"conv":  shortHex(bd.Verdict.ConventionsDigest),
			"tx":    shortHex(bd.Verdict.JournalTx),
			"at":    txTime[bd.Verdict.JournalTx],
		}))
	}
	if acceptanceOut != "" {
		fmt.Fprintf(&sb, "%s", canondata.T("why.verifications.acceptance", canondata.M{
			"outcome": acceptanceOut, "green": acceptanceGreen, "checks": acceptanceChecks,
		}))
	}
	return strings.TrimRight(sb.String(), "\n"), nil
}

// whyAttention — the human attention summary: how many choice points
// the instance consumed, how many were closed by silence defaults,
// how much manual prose was recorded; all is a projection of the
// ledger and the curator zone. The human only chooses: every manual
// record is a candidate for the norm (decisions/kb are written by a
// delta immediately).
func (e *Engine) whyAttention() string {
	var questions, choices, answered, defaults int64
	var attentionMs int64
	for _, en := range e.Ledger.All() {
		switch en.Event {
		case "question":
			questions++
		case "choice":
			choices++
			attentionMs += parseLedgerInt(en.Details["attention-ms"])
			answered += parseLedgerInt(en.Details["answered"])
			defaults += parseLedgerInt(en.Details["defaults"])
		case "answer-upgrade":
			// an explicit answer overrode a silence default: the
			// counters show the choice's final truth, not the history
			// of wavering
			answered += parseLedgerInt(en.Details["answers"])
			defaults -= parseLedgerInt(en.Details["answers"])
		}
	}
	decisions, _ := e.Store.Decisions()
	kb, _ := e.Store.KbEntries()
	return canondata.T("why.attention.head", canondata.M{
		"questions": strconv.FormatInt(questions, 10),
		"choices":   strconv.FormatInt(choices, 10),
		"answered":  strconv.FormatInt(answered, 10),
		"defaults":  strconv.FormatInt(defaults, 10),
		"ms":        strconv.FormatInt(attentionMs, 10),
		"decisions": strconv.Itoa(len(decisions)),
		"kb":        strconv.Itoa(len(kb)),
	})
}

// parseLedgerInt — a number from entry details; empty means zero.
func parseLedgerInt(s string) int64 {
	if s == "" {
		return 0
	}
	var v int64
	fmt.Sscanf(s, "%d", &v)
	return v
}

// whyHypotheses — the registry of canon-change hypotheses: what the
// delivery carries under a testable hypothesis; an empty registry is
// honestly empty — there are no canon changes waiting to be tested.
func whyHypotheses() string {
	list := canondata.Hypotheses()
	if len(list) == 0 {
		return canondata.T("why.hypotheses.none")
	}
	var sb strings.Builder
	sb.WriteString(canondata.T("why.hypotheses.head"))
	sb.WriteString("\n")
	for _, h := range list {
		fmt.Fprintf(&sb, "%s\n", canondata.T("why.hypotheses.item", canondata.M{
			"id": h.ID, "change": h.Change, "hypothesis": h.Hypothesis,
			"metric": h.Metric, "window": h.Window,
			"baseline": h.Baseline, "outcome": h.Outcome,
		}))
	}
	return strings.TrimRight(sb.String(), "\n")
}

// whyCompatibility — the compatibility matrix: "verified" only by
// live run facts. The binary's current environment is a measured
// fact of the delivery (it is executing); everything not run is
// honestly unmeasured: the unknown stays unknown, silence is
// forbidden. Compatibility axes are not invented — only a live run
// confirms them.
func whyCompatibility() string {
	return canondata.T("why.compatibility.head") + "\n" +
		canondata.T("why.compatibility.runtime", canondata.M{
			"os": runtime.GOOS, "arch": runtime.GOARCH,
		}) + "\n" +
		canondata.T("why.compatibility.notmeasured")
}
