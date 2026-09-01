// The next verb: state, exactly one action with a price, a ready
// example, a layer-by-layer slice and delivery gates. The output is
// deterministic: the same canon — bit-for-bit the same text.
package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/compiler"
	"github.com/neurophant/punchtape/internal/delta"
	"github.com/neurophant/punchtape/internal/ledger"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// SlotInfo — the machine coordinates of a slot for the scheduler:
// stage, expected delta kind, and the card with facet on implement.
// The coordinates are not visible to the hand agent: it reads the
// slot text.
type SlotInfo struct {
	Stage string
	Kind  string // intent | spec | assert | code | fix
	Card  string
	Facet string
}

// Next — "what is next".
func (e *Engine) Next() (string, error) {
	return e.nextVerb(false)
}

// NextFull — an explicit request for the full printout: for a new hand.
func (e *Engine) NextFull() (string, error) {
	return e.nextVerb(true)
}

func (e *Engine) nextVerb(forceFull bool) (string, error) {
	started := time.Now()
	// A cache keyed by the input: while the journal and canon files
	// have not changed, a repeated next is free. A repeated read is a
	// contextual diff: changed blocks in full, unchanged ones as
	// anchors with a digest; a fresh render serves the same diff once
	// anything was shown — the full printout goes to the first read
	// of an instance or an explicit request.
	if cached, ok := e.cachedNext(); ok {
		if forceFull {
			e.logVerb("next", "ok", map[string]string{
				"slot": "full-on-request", "context-bytes": fmt.Sprintf("%d", len(cached)),
			}, time.Since(started).Milliseconds(), 0)
			return cached, nil
		}
		text, changed, anchored, diff := compactSlot(cached, loadShownBlocks(e))
		e.logVerb("next", "ok", map[string]string{
			"slot": "compact", "context-bytes": fmt.Sprintf("%d", len(text)),
			"slot-blocks-changed":  fmt.Sprintf("%d", changed),
			"slot-blocks-anchored": fmt.Sprintf("%d", anchored),
			"slot-diff":            diff,
		}, time.Since(started).Milliseconds(), 0)
		return text, nil
	}

	state, err := e.Store.State()
	if err != nil {
		return "", err
	}
	text, _, err := e.renderNext(state, "")
	if err != nil {
		return "", err
	}
	prev := loadShownBlocks(e)
	e.storeCachedNext(text)
	storeShownBlocks(e, text)
	e.noteContourDrift(e.lastRenderDrift)
	if !forceFull && len(prev) > 0 {
		var changed, anchored int
		var diff string
		text, changed, anchored, diff = compactSlot(text, prev)
		e.logVerb("next", "ok", map[string]string{
			"slot": "compact-fresh", "context-bytes": fmt.Sprintf("%d", len(text)),
			"slot-blocks-changed":  fmt.Sprintf("%d", changed),
			"slot-blocks-anchored": fmt.Sprintf("%d", anchored),
			"slot-diff":            diff,
		}, time.Since(started).Milliseconds(), 0)
		return text, nil
	}
	e.logVerb("next", "ok", map[string]string{"slot": "full", "context-bytes": fmt.Sprintf("%d", len(text))},
		time.Since(started).Milliseconds(), 0)
	return text, nil
}

// NextInfo — next with the machine coordinates of the slot: the
// driver needs the expected delta kind for routing executors.
func (e *Engine) NextInfo() (string, SlotInfo, error) {
	state, err := e.Store.State()
	if err != nil {
		return "", SlotInfo{}, err
	}
	return e.renderNext(state, "")
}

// SlotForCard — the slot of a specific card on implement: a
// parallel scheduler works cards simultaneously, each with its own
// slice and its own red list.
func (e *Engine) SlotForCard(cardID string) (string, SlotInfo, error) {
	state, err := e.Store.State()
	if err != nil {
		return "", SlotInfo{}, err
	}
	if state.Stage != canon.StageImplement {
		return "", SlotInfo{}, fmt.Errorf("%s", canondata.T("slot.error.no-per-card-slots", canondata.M{"stage": state.Stage}))
	}
	return e.renderNext(state, cardID)
}

func (e *Engine) renderNext(state canon.State, cardID string) (string, SlotInfo, error) {
	var info SlotInfo
	info.Stage = state.Stage
	var sb strings.Builder
	fmt.Fprintf(&sb, "STAGE: %s\n", state.Stage)

	// An open pending amend: the judge does not rewrite its own
	// verdict — confirming changes to the expectations of already-run
	// checks is held by the human. The block lives on every stage
	// while the amend is open; silence resolves it with the
	// conservative default.
	if state.PendingAmend != nil {
		fmt.Fprintf(&sb, "%s\n", canondata.TFor(state.Language, "slot.amend.block", canondata.M{
			"count":   fmt.Sprintf("%d", len(state.PendingAmend.Changes)),
			"changes": amendChangesBlock(state.PendingAmend.Changes, state.PendingAmend.Plan),
			"key":     state.PendingAmend.Key,
		}))
	}

	// Structured-intake: with an open batch the slot carries the intake (the block lives on every stage — the machine's compositionality question may open later than the spec)
	// in full — "what was understood" (the wish verbatim, without
	// embellishment), clarifications in a batch with defaults, "what I
	// can do", an escalation (conflicts are a live choice, silence
	// does not resolve them by default). Only the unresolved is
	// rendered: answered questions live in the CLARIFIED list below
	// and are not asked again.
	if open := openFindings(state); len(open) > 0 {
		var plainQ, conflicts []canon.LintFinding
		for _, f := range open {
			if f.Dimension == canon.LintConflict {
				conflicts = append(conflicts, f)
			} else {
				plainQ = append(plainQ, f)
			}
		}
		// Human blocks are in the wish's language;
		// there are no internal codes in the human dialog: the
		// finding's dimension is a machine classification, the human
		// gets the meaning.
		lang := state.Language
		if state.Intent != nil {
			fmt.Fprintf(&sb, "%s\n", canondata.TFor(lang, "slot.intake.understood", canondata.M{
				"text": indentLines(*state.Intent),
			}))
		}
		if len(plainQ) > 0 {
			fmt.Fprintf(&sb, "%s\n", canondata.TFor(lang, "slot.spec.questions", canondata.M{
				"batch": renderQuestionBatch(lang, plainQ),
				// the reply example carries the actual ID of the
				// batch's first open question: a sample letter without
				// an ID is a lost submission
				"qid": plainQ[0].ID,
			}))
		}
		fmt.Fprintln(&sb, canondata.TFor(lang, "slot.intake.can-do"))
		if len(conflicts) > 0 {
			fmt.Fprintf(&sb, "%s\n", canondata.TFor(lang, "slot.spec.escalation", canondata.M{
				"batch": renderQuestionBatch(lang, conflicts),
				"qid":   conflicts[0].ID,
			}))
		}
	}

	// The contour point "freshness before work": drift of the
	// passport boundaries against the current instance is named
	// before work starts — building new on stale ground silently is
	// not allowed. A fresh contour stays silent.
	if line, drifted := e.contourBeforeWork(); line != "" {
		e.lastRenderDrift = drifted
		fmt.Fprintf(&sb, "%s\n", line)
	} else {
		e.lastRenderDrift = nil
	}

	var action, example, context, gates string
	// slotKey — the slot's identity (stage + work kind): by it the
	// shown-slots registry decides whether to reprint the example
	// block. An empty key means the example is not a format template
	// (a machine proposal) — it is never slimmed.
	var slotKey string
	switch state.Stage {
	case canon.StageIntake:
		// The single intake slot — an unpinned wish; the
		// clarification loop lives in the checks table (questions and
		// answers travel as row submissions), there is no separate
		// relay stage.
		if !state.IntentAccepted {
			action = canondata.T("slot.intake.action.pin")
			example = exampleIntent(e.nextKeyNum())
			context = canondata.T("slot.intake.context.empty")
			gates = "trace"
			info.Kind = "intent"
			slotKey = "intake/intent"
			fmt.Fprintln(&sb, canondata.T("slot.intake.state"))
		} else {
			// Unreachable after any submission (advance moves the
			// stage), but an honest render is cheaper than a panic on
			// a forgotten state.
			action = canondata.T("slot.intake.action.pinned")
			context = canondata.T("slot.intake.context.pinned")
			gates = "trace"
			info.Kind = "intent"
		}
	case canon.StageSpec:
		reqs, scns, err := e.loadCanon()
		if err != nil {
			return "", info, err
		}
		fmt.Fprint(&sb, canondata.T("slot.spec.state", canondata.M{
			"reqs": fmt.Sprintf("%d", len(reqs)), "scns": fmt.Sprintf("%d", len(scns)),
		}))
		if state.Intent != nil {
			fmt.Fprint(&sb, canondata.T("slot.spec.state-intent", canondata.M{"intent": firstLineOf(*state.Intent)}))
		}
		fmt.Fprintln(&sb)
		if len(state.Clarifications) > 0 {
			fmt.Fprintf(&sb, "%s\n", canondata.T("slot.spec.clarified", canondata.M{
				"count": fmt.Sprintf("%d", len(state.Clarifications)),
			}))
			for _, c := range state.Clarifications {
				line := fmt.Sprintf("  %s: \"%s\" %s", c.Question, c.Option, c.Label)
				if c.ByDefault {
					line += " — default (silence)"
				}
				fmt.Fprintf(&sb, "%s\n", line)
			}
		}
		fmt.Fprintln(&sb, canondata.T("slot.spec.surface-contract"))
		fmt.Fprintln(&sb, canondata.T("slot.spec.table"))
		if e.draftPending() {
			// The draft was written by a cheap executor — the spec is
			// held for the hand's review: row edits (the same
			// submission) or accepting it whole.
			action = canondata.T("slot.spec.review-draft.action")
			example = canondata.T("slot.spec.review-draft.example", canondata.M{
				"key": fmt.Sprintf("%03d", e.nextKeyNum()),
			})
			context = e.sliceTouched(state, reqs, scns)
			gates = "trace"
			info.Kind = "spec"
			slotKey = "spec/review-draft"
		} else {
			action = canondata.T("slot.spec.action")
			example = exampleChecks(e.nextKeyNum())
			context = e.sliceTouched(state, reqs, scns)
			gates = "trace"
			info.Kind = "spec"
			slotKey = "spec/spec"
			// An empty spec without a draft is the draft slot: routing
			// may hand it to a cheap executor, SubmitDraft will mark
			// the origin; without routing the hand itself writes it.
			if len(scns) == 0 {
				info.Kind = "draft"
			}
		}
	case canon.StageACCompile:
		scns, chk, err := e.loadForCompile()
		if err != nil {
			return "", info, err
		}
		var skeletons []canon.Scenario
		executable, prose := 0, 0
		for _, sc := range scns {
			d := compiler.DegreeOf(sc)
			switch d.Level {
			case compiler.DegreeExecutable:
				executable++
			case compiler.DegreeProse:
				prose++
			default:
				skeletons = append(skeletons, sc)
			}
		}
		fmt.Fprintf(&sb, "%s\n", canondata.T("slot.ac-compile.state", canondata.M{
			"total": fmt.Sprintf("%d", len(scns)),
			"exec":  fmt.Sprintf("%d", executable),
			"pct":   fmt.Sprintf("%d", percent(executable, len(scns))),
			"skel":  fmt.Sprintf("%d", len(skeletons)),
			"await": "0",
			"prose": fmt.Sprintf("%d", prose),
		}))
		fmt.Fprintf(&sb, "%s\n", canondata.T("slot.ac-compile.traceability", canondata.M{
			"total":   fmt.Sprintf("%d", len(chk)),
			"green":   fmt.Sprintf("%d", countOutcome(chk, canon.CheckGreen)),
			"red":     fmt.Sprintf("%d", countOutcome(chk, canon.CheckRed)),
			"unknown": fmt.Sprintf("%d", countOutcome(chk, canon.CheckUnknown)),
		}))
		if line := e.impactLine(state, scns); line != "" {
			fmt.Fprintln(&sb, line)
		}
		if missing := missingLinks(scns, chk); len(missing) > 0 {
			fmt.Fprintf(&sb, "%s\n", canondata.T("slot.missing-links", canondata.M{
				"links": strings.Join(missing, ", "),
			}))
		}
		// The family coverage gate: the refusal is by name — the
		// hand sees which command sequences of the brief still lack
		// an executable scenario. A gap accepted by the human with a
		// reason does not hold code — but stays in the list and in
		// the verdict.
		unc := uncoveredFamilies(state, scns)
		waivedFam := familiesWaived(state)
		uncOpen := make([]string, 0, len(unc))
		for _, f := range unc {
			if !waivedFam[f] {
				uncOpen = append(uncOpen, f)
			}
		}
		if len(unc) > 0 {
			fmt.Fprintf(&sb, "%s\n", canondata.T("slot.ac-compile.uncovered.header", canondata.M{
				"count": fmt.Sprintf("%d", len(unc)),
			}))
			for i, f := range unc {
				if i >= canondata.Limit("slot.list-budget") {
					fmt.Fprintf(&sb, "%s\n", canondata.T("slot.list.more", canondata.M{
						"count": fmt.Sprintf("%d", len(unc)-canondata.Limit("slot.list-budget")),
					}))
					break
				}
				fmt.Fprintf(&sb, "  `%s`\n", f)
			}
			// The waiver path with a reason: the rule is named,
			// the reason is mandatory — an opinion queue, not a
			// silent stop.
			if len(uncOpen) > 0 {
				fmt.Fprintf(&sb, "%s\n", canondata.T("slot.ac-compile.waive-hint", canondata.M{
					"rule": "coverage.family",
				}))
			}
		}
		switch {
		case len(skeletons) > 0:
			action = canondata.T("slot.ac-compile.action.pin-assert", canondata.M{"scenario": skeletons[0].ID})
			example = exampleAssert(skeletons[0], e.nextKeyNum())
			// The slice is only the touched facet of one action; the
			// other unfinished ones as a counter: the output volume
			// does not grow with the instance.
			context = e.sliceScenarios(skeletons[:1])
			if len(skeletons) > 1 {
				context += "\n" + canondata.T("slot.ac-compile.context.skeletons-remaining", canondata.M{
					"count": fmt.Sprintf("%d", len(skeletons)-1),
				})
			}
			info.Kind = "assert"
			slotKey = "ac-compile/assert/" + skeletons[0].ID
		case len(uncOpen) > 0:
			// The coverage gate: no unfinished scenarios remain, but
			// brief families without an executable scenario do — the
			// entry into code is closed, the work returns to the spec
			// composition (a gap can be accepted with a reason — the
			// path is named above).
			action = canondata.T("slot.ac-compile.action.uncovered")
			example = exampleChecks(e.nextKeyNum())
			context = canondata.T("slot.ac-compile.context.coverage-gate", canondata.M{
				"count": fmt.Sprintf("%d", len(uncOpen)),
			})
			info.Kind = "spec"
			slotKey = "ac-compile/coverage"
		default:
			action = canondata.T("slot.ac-compile.action.none")
			example = ""
			context = canondata.T("slot.ac-compile.context.ready")
		}
		gates = "trace"
	case canon.StageImplement:
		cards, err := e.Store.Cards()
		if err != nil {
			return "", info, err
		}
		scns, err := e.Store.Scenarios()
		if err != nil {
			return "", info, err
		}
		chk, err := e.Store.Checks()
		if err != nil {
			return "", info, err
		}
		scnsByID := map[string]canon.Scenario{}
		for _, sc := range scns {
			scnsByID[sc.ID] = sc
		}
		// The action card: the first active one, else the first red —
		// the red list is the work; the parallel scheduler requests
		// the slot of a specific card itself. Green and quarantine
		// cards carry no work.
		var card canon.Card
		var haveCard bool
		for _, c := range cards {
			if c.Status == canon.CardActive {
				card, haveCard = c, true
				break
			}
		}
		if !haveCard {
			for _, c := range cards {
				if c.Status == canon.CardRed {
					card, haveCard = c, true
					break
				}
			}
		}
		if !haveCard && len(cards) > 0 {
			card = cards[0]
		}
		if cardID != "" {
			card = canon.Card{}
			for _, c := range cards {
				if c.ID == cardID {
					card = c
					break
				}
			}
		}
		info.Card, info.Facet = card.ID, card.Facet
		active, green := 0, 0
		for _, c := range cards {
			switch c.Status {
			case canon.CardActive:
				active++
			case canon.CardGreen:
				green++
			}
		}
		if len(cards) > 1 {
			fmt.Fprintf(&sb, "%s\n", canondata.T("slot.implement.state.cards", canondata.M{
				"total":     fmt.Sprintf("%d", len(cards)),
				"active":    fmt.Sprintf("%d", active),
				"green":     fmt.Sprintf("%d", green),
				"card":      card.ID,
				"facet":     card.Facet,
				"status":    card.Status,
				"scenarios": fmt.Sprintf("%d", len(card.Scenarios)),
				"attempts":  fmt.Sprintf("%d", card.Attempts),
			}))
			fmt.Fprintln(&sb, "CARDS:")
			for i, c := range cards {
				if i >= canondata.Limit("slot.list-budget") {
					fmt.Fprintf(&sb, "%s\n", canondata.T("slot.list.more", canondata.M{
						"count": fmt.Sprintf("%d", len(cards)-canondata.Limit("slot.list-budget")),
					}))
					break
				}
				fmt.Fprintf(&sb, "  %s %s: %d scenarios, %d files, attempts %d [%s]\n",
					c.ID, c.Facet, len(c.Scenarios), len(c.Files), c.Attempts, c.Status)
			}
		} else {
			fmt.Fprintf(&sb, "%s\n", canondata.T("slot.implement.state.card", canondata.M{
				"card":      card.ID,
				"facet":     card.Facet,
				"status":    card.Status,
				"scenarios": fmt.Sprintf("%d", len(card.Scenarios)),
				"attempts":  fmt.Sprintf("%d", card.Attempts),
			}))
		}
		if line := e.impactLine(state, scns); line != "" {
			fmt.Fprintln(&sb, line)
		}
		fmt.Fprintln(&sb, canondata.T("slot.implement.surface-contract"))
		// Empty conventions start: the card's surfaces do not resolve
		// from the project and there is no recipe — the menu of
		// options with price and recommendation is shown from THE
		// FIRST implement slot (before code): rows without a surface
		// fail "not installed", the order "conventions or submission
		// files" must be known before the first code submission, not
		// after it. The price of the "teach the recipe" option is
		// bytes/4 of the machine example: exactly what the hand would
		// have to type.
		if missing := e.missingSurfaces(card); len(missing) > 0 {
			if conv, err := e.Store.Conventions(); err == nil || conv == nil || conv.Build == nil {
				example := canondata.T("slot.example.conventions", canondata.M{
					"key": fmt.Sprintf("%03d", e.nextKeyNum()),
				})
				fmt.Fprintln(&sb, canondata.TFor(state.Language, "slot.implement.conventions-needed", canondata.M{
					"names": strings.Join(missing, ", "),
					"price": fmt.Sprintf("%d", tokenEstimate(example)),
				}))
				// A kb consultation before the question: active
				// knowledge for the question's context is shown with
				// its source — the action stays with the operator.
				if hint := e.conventionsKbHint(state, missing); hint != "" {
					fmt.Fprintln(&sb, hint)
				}
				// The example is a block scalar without a trailing
				// newline: the newline is appended here, the next
				// block does not stick to it.
				fmt.Fprintln(&sb, example)
			}
		}
		// A link deleted by hand is visible at once, not only at the verdict.
		if missing := missingLinks(scns, chk); len(missing) > 0 {
			fmt.Fprintf(&sb, "%s\n", canondata.T("slot.missing-links", canondata.M{
				"links": strings.Join(missing, ", "),
			}))
		}
		// The red list is one line of feedback per check. Static
		// blockers (types/lint, build) do not live in checks — a
		// separate block, verbatim: the hand must see the reason for
		// a red in the slot.
		red := redChecks(card, chk)
		blockers := card.Blockers
		var cardScenarios []canon.Scenario
		for _, id := range card.Scenarios {
			cardScenarios = append(cardScenarios, scnsByID[id])
		}
		if len(blockers) > 0 {
			fmt.Fprintf(&sb, "%s\n", canondata.T("slot.implement.gate-reds.header", canondata.M{
				"count": fmt.Sprintf("%d", len(blockers)),
			}))
			for i, line := range blockers {
				if i >= canondata.Limit("slot.list-budget") {
					fmt.Fprintf(&sb, "%s\n", canondata.T("slot.list.more", canondata.M{
						"count": fmt.Sprintf("%d", len(blockers)-canondata.Limit("slot.list-budget")),
					}))
					break
				}
				fmt.Fprintf(&sb, "  %s\n", line)
			}
		}
		if len(red) > 0 {
			// Grouping identical reasons: check runs with one reason
			// ("not built", the same failure) are printed as one
			// line — the list of identifiers folds, the reason stays
			// verbatim.
			type run struct {
				first, last int
				reason      string
			}
			var runs []run
			for i, line := range red {
				if len(runs) > 0 && runs[len(runs)-1].reason == line.Reason {
					runs[len(runs)-1].last = i
					continue
				}
				runs = append(runs, run{first: i, last: i, reason: line.Reason})
			}
			fmt.Fprintf(&sb, "%s\n", canondata.T("slot.implement.red-list.header", canondata.M{
				"count": fmt.Sprintf("%d", len(red)),
			}))
			for i, r := range runs {
				if i >= canondata.Limit("slot.list-budget") {
					fmt.Fprintf(&sb, "%s\n", canondata.T("slot.list.more", canondata.M{
						"count": fmt.Sprintf("%d", len(red)-runs[i-1].last-1),
					}))
					break
				}
				if r.first == r.last {
					fmt.Fprintf(&sb, "  %s %s: %s\n", red[r.first].ID, red[r.first].Scenario, r.reason)
					continue
				}
				fmt.Fprintf(&sb, "  %s..%s %s..%s (%d checks): %s\n",
					red[r.first].ID, red[r.last].ID, red[r.first].Scenario, red[r.last].Scenario,
					r.last-r.first+1, r.reason)
			}
			// Repair-vs-regenerate by measured drift: a red
			// majority of the facet or two failed repairs —
			// regeneration is cheaper; otherwise a point fix.
			regenerate := len(red) > len(card.Scenarios)/2 || card.Attempts >= 2
			withBlockers := ""
			if len(blockers) > 0 {
				withBlockers = canondata.T("slot.implement.with-blockers")
			}
			if regenerate {
				action = canondata.T("slot.implement.action.regenerate", canondata.M{
					"card":     card.ID,
					"blockers": withBlockers,
					"red":      fmt.Sprintf("%d", len(red)),
					"total":    fmt.Sprintf("%d", len(card.Scenarios)),
					"attempts": fmt.Sprintf("%d", card.Attempts),
					"budget":   fmt.Sprintf("%d", canondata.Limit("card.attempts-budget")),
				})
				example = exampleCode(card, e.nextKeyNum())
				context = e.sliceScenarios(cardScenarios)
				info.Kind = "code"
				slotKey = "implement/code/" + card.ID
			} else {
				action = canondata.T("slot.implement.action.fix", canondata.M{
					"card":     card.ID,
					"blockers": withBlockers,
					"reply":    "suite",
				})
				example = exampleFix(card, e.nextKeyNum(), "suite")
				// The repair slice carries only the red
				// list's scenarios — the context is proportional to
				// the edit, not the facet's size; the full depth
				// stays on demand via why.
				redIDs := map[string]bool{}
				for _, line := range red {
					redIDs[line.Scenario] = true
				}
				var redScenarios []canon.Scenario
				for _, sc := range cardScenarios {
					if redIDs[sc.ID] {
						redScenarios = append(redScenarios, sc)
					}
				}
				context = e.sliceScenarios(redScenarios) + "\n" +
					canondata.T("slot.implement.context.fix-cycle", canondata.M{
						"red":   fmt.Sprintf("%d", len(redScenarios)),
						"total": fmt.Sprintf("%d", len(cardScenarios)),
					})
				info.Kind = "fix"
				slotKey = "implement/fix/" + card.ID
			}
		} else if len(blockers) > 0 {
			// Behavior is green, only the statics are red: the reason
			// is in the GATE REDS block, that is what must be fixed,
			// not the behavior rewritten.
			reply := blockerReplyTo(blockers)
			action = canondata.T("slot.implement.action.fix-static", canondata.M{
				"card":  card.ID,
				"reply": reply,
			})
			example = exampleFix(card, e.nextKeyNum(), reply)
			context = canondata.T("slot.implement.context.static-red", canondata.M{
				"count": fmt.Sprintf("%d", len(cardScenarios)),
			})
			info.Kind = "fix"
			slotKey = "implement/fix/" + card.ID
		} else if pinRed := pinRedLines(card, chk); len(pinRed) > 0 {
			// Behavior is green, the probe is red: the rows prove
			// nothing, they are fixed by replacing them in the checks
			// table — freely, without human confirmation; there is no
			// code to fix.
			fmt.Fprintf(&sb, "%s\n", canondata.T("slot.implement.pin-reds.header", canondata.M{
				"count": fmt.Sprintf("%d", len(pinRed)),
			}))
			for _, line := range pinRed {
				fmt.Fprintf(&sb, "  %s\n", line)
			}
			action = canondata.T("slot.implement.action.pin-repair", canondata.M{
				"count": fmt.Sprintf("%d", len(pinRed)),
			})
			example = exampleChecks(e.nextKeyNum())
			context = e.sliceScenarios(cardScenarios)
			info.Kind = "checks"
			slotKey = "implement/checks/" + card.ID
		} else {
			action = canondata.T("slot.implement.action.code", canondata.M{"card": card.ID})
			example = exampleCode(card, e.nextKeyNum())
			context = e.sliceScenarios(cardScenarios)
			info.Kind = "code"
			slotKey = "implement/code/" + card.ID
		}
		// The first code submission freezes the spec composition: the
		// warning stands exactly before that submission — started
		// code lives by the rule without reminders.
		if info.Kind == "code" && !e.codeStarted() {
			fmt.Fprintln(&sb, canondata.T("slot.implement.composition-freeze"))
		}
		// The cut proposal is part of the context: the slot's price
		// honestly accounts for it too. It appears only on an
		// untouched whole-product card, when the spec divides into
		// independent facets.
		if proposal := proposalCut(state, cards, scns); proposal != "" {
			context = proposal + "\n" + context
		}
		gates = liveGateList(state)
	case canon.StageDeliver:
		// The verdict is rendered by the machine in canonical format;
		// an agent's retelling is not a verdict.
		verdict, err := e.RenderVerdict()
		if err != nil {
			return "", info, err
		}
		action = canondata.T("slot.deliver.action")
		example = ""
		context = verdict.Text()
		gates = "—"
		// Backfill: the spec catches up with stabilized code. The
		// machine proposes a spec delta from the observable — the
		// observations are already measured, the proposal assembles
		// them into ready text; the submission stays with the agent,
		// "realized" is decided by checks.
		scns, err := e.Store.Scenarios()
		if err != nil {
			return "", info, err
		}
		switch uncovered, errored := e.uncoveredProbes(scns); {
		case len(uncovered) > 0:
			info.Kind = "checks"
			action = canondata.T("slot.deliver.action.backfill", canondata.M{
				"count": fmt.Sprintf("%d", len(uncovered)),
			})
			example = exampleBackfillChecks(uncovered, e.nextKeyNum())
			context = probeSummary(uncovered, errored) + "\n\n" + verdict.Text()
		case len(errored) > 0:
			// Nothing to formalize: failed probes stay an honest line
			// next to the verdict, they do not become a proposal.
			context = probeSummary(nil, errored) + "\n\n" + verdict.Text()
		}
		// Instance drift against the acceptance moment: the verdict
		// is honestly unreconciled, the action is re-verification of
		// the changed instance with the card's fix (fix on deliver is
		// allowed exactly for this).
		if e.lastVerdictDrift {
			action = canondata.T("slot.deliver.action.drift", canondata.M{
				"card": e.workingCard(),
			})
			example = ""
			info.Kind = "fix"
		}
	}

	// The compaction trigger: the journal threshold is data-driven —
	// a recommendation line to hand off the session (why handoff
	// carries a ready prompt); not a blocker: silence continues, rc
	// does not change.
	if n := len(e.Journal.All()); n >= canondata.Limit("compact.journal-entries") {
		fmt.Fprintf(&sb, "%s\n", canondata.T("slot.session.handoff", canondata.M{
			"n":         fmt.Sprintf("%d", n),
			"threshold": fmt.Sprintf("%d", canondata.Limit("compact.journal-entries")),
		}))
	}

	fmt.Fprintf(&sb, "ACTION: %s\n", action)
	// The price was removed from the slot: it recurred in every slot
	// at 0.9% of the bytes — noise for the hand.
	// The origin of numbers goes to the human, via why price; the
	// hand's slot is only the action and its content.
	if example != "" {
		if slotKey != "" && e.slotShown(slotKey) {
			// A repeat visit to the same slot: the delta format is
			// already shown — one reference line with a fresh
			// submission key instead of reprinting the example block
			// (a ceremony tax; the spec slot is up to one third
			// example).
			ref := canondata.T("slot.example.not-repeated", canondata.M{"key": exampleKeyOf(example)})
			fmt.Fprintf(&sb, "EXAMPLE: %s\n", ref)
		} else {
			fmt.Fprintf(&sb, "EXAMPLE:\n%s\n", example)
			if slotKey != "" {
				e.markSlotShown(slotKey)
			}
		}
	}
	fmt.Fprintf(&sb, "CONTEXT:\n%s\n", context)
	fmt.Fprintf(&sb, "GATES ON SUBMIT: %s", gates)
	return sb.String(), info, nil
}

// --- context slices ---

// sliceScenarios gives full scenario cards of the touched facet:
// that is what the agent needs for the task; the rest of the
// instance is not its business.
func (e *Engine) sliceScenarios(scns []canon.Scenario) string {
	if len(scns) == 0 {
		return canondata.T("slot.context.no-scenarios")
	}
	var sb strings.Builder
	for _, sc := range scns {
		data, err := yamlio.Marshal(sc)
		if err != nil {
			continue
		}
		fmt.Fprintf(&sb, "---\n%s", data)
	}
	return strings.TrimRight(sb.String(), "\n")
}

// sliceTouched — the spec-stage slice: full cards of the entities
// touched by the last submission.
func (e *Engine) sliceTouched(state canon.State, reqs []canon.Requirement, scns []canon.Scenario) string {
	if len(state.Touched) == 0 {
		return canondata.T("slot.context.first-spec", canondata.M{
			"reqs": fmt.Sprintf("%d", len(reqs)), "scns": fmt.Sprintf("%d", len(scns)),
		})
	}
	touched := map[string]bool{}
	for _, id := range state.Touched {
		touched[id] = true
	}
	var sb strings.Builder
	for _, r := range reqs {
		if touched[r.ID] {
			data, _ := yamlio.Marshal(r)
			fmt.Fprintf(&sb, "---\n%s", data)
		}
	}
	for _, sc := range scns {
		if touched[sc.ID] {
			data, _ := yamlio.Marshal(sc)
			fmt.Fprintf(&sb, "---\n%s", data)
		}
	}
	if sb.Len() == 0 {
		return canondata.T("slot.context.bookkeeping-only", canondata.M{
			"reqs": fmt.Sprintf("%d", len(reqs)), "scns": fmt.Sprintf("%d", len(scns)),
		})
	}
	return strings.TrimRight(sb.String(), "\n")
}

// liveGateList — the list of delivery gates of the implement stage,
// honest about the ones retired by rent: what is not run is not in
// the list.
func liveGateList(state canon.State) string {
	all := []string{"suite", "types", "lint", "trace"}
	live := make([]string, 0, len(all))
	for _, name := range all {
		if !gateRetired(state, name) {
			live = append(live, name)
		}
	}
	return strings.Join(live, ", ")
}

// impactLine — the estimate line of the last spec change: the check
// compiler's facts and the slot estimate of the touched facet. An
// empty line means the spec did not change, there is no estimate.
func (e *Engine) impactLine(state canon.State, scns []canon.Scenario) string {
	if state.Impact == nil {
		return ""
	}
	byID := map[string]canon.Scenario{}
	for _, sc := range scns {
		byID[sc.ID] = sc
	}
	var touched []canon.Scenario
	for _, id := range state.Touched {
		if sc, ok := byID[id]; ok {
			touched = append(touched, sc)
		}
	}
	return canondata.T("slot.impact", canondata.M{
		"touched":     fmt.Sprintf("%d", state.Impact.ScenariosTouched),
		"regenerated": fmt.Sprintf("%d", state.Impact.ChecksRegenerated),
		"kept":        fmt.Sprintf("%d", state.Impact.ChecksKept),
		"removed":     fmt.Sprintf("%d", state.Impact.ChecksRemoved),
		"tokens":      fmt.Sprintf("%d", tokenEstimate(e.sliceScenarios(touched))),
	})
}

// --- the cut proposal ---

// conventionsKbHint — kb advice for the build recipe question: the
// context is the wish's tokens and the names of missing surfaces;
// the active knowledge found is rendered with its source.
func (e *Engine) conventionsKbHint(state canon.State, missing []string) string {
	kb, err := e.Store.KbEntries()
	if err != nil || len(kb) == 0 {
		return ""
	}
	tokens := map[string]bool{}
	if state.Intent != nil {
		for _, tok := range strings.Fields(strings.ToLower(*state.Intent)) {
			tokens[tok] = true
		}
	}
	for _, name := range missing {
		tokens[strings.ToLower(name)] = true
	}
	if hit := kbAdvice(kb, tokens); hit != nil {
		return canondata.T("slot.conventions.kb-hint", canondata.M{
			"id": hit.ID, "do": hit.Do, "source": hit.Source,
		})
	}
	return ""
}

// proposalCut — the machine's proposal to cut the product:
// connectivity components of executable scenarios by shared state
// files. It appears only when the work has not started yet (an
// untouched whole-product card) and there is more than one
// component: parallelism without shared assets needs no contracts,
// the boundaries are already separated by state.
func proposalCut(state canon.State, cards []canon.Card, scns []canon.Scenario) string {
	if len(cards) != 1 || cards[0].Facet != "whole-product" ||
		cards[0].Status != canon.CardActive || cards[0].Attempts != 0 {
		return ""
	}
	comps := stateComponents(scns)
	if len(comps) < 2 {
		// A big scope is split: above the data threshold —
		// an honest line about the large composition, even if there
		// are no shared files yet and no cut is possible for now.
		if max := canondata.Limit("detail.split-threshold"); len(scns) >= max {
			return canondata.T("slot.cut.big-scope", canondata.M{
				"count": fmt.Sprintf("%d", len(scns)), "max": fmt.Sprintf("%d", max),
			})
		}
		return ""
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s\n", canondata.T("slot.cut.proposal", canondata.M{
		"count": fmt.Sprintf("%d", len(comps)),
	}))
	for i, comp := range comps {
		files := groupStateFiles(comp)
		line := fmt.Sprintf("  facet-%d: %s", i+1, strings.Join(scenarioIDs(comp), ", "))
		if len(files) > 0 {
			line += fmt.Sprintf(" (shares %s)", strings.Join(files, ", "))
		} else {
			line += " (no state files)"
		}
		fmt.Fprintln(&sb, line)
	}
	fmt.Fprintf(&sb, "%s\n", canondata.T("slot.cut.split-hint", canondata.M{
		"seed": fmt.Sprintf("%03d", state.Counters["CRD"]),
	}))
	sb.WriteString(exampleSplit(comps, state.Counters["CRD"]))
	fmt.Fprintln(&sb, canondata.T("slot.cut.different-allowed"))
	return sb.String()
}

// stateComponents — connectivity components of executable scenarios
// by shared state files: scenarios of one component share an
// observable asset, different components are independent. Fileless
// scenarios form one component: no shared assets, no conflicts
// either.
func stateComponents(scns []canon.Scenario) [][]canon.Scenario {
	parent := map[string]string{}
	find := func(x string) string {
		for parent[x] != x {
			parent[x] = parent[parent[x]]
			x = parent[x]
		}
		return x
	}
	union := func(a, b string) {
		ra, rb := find(a), find(b)
		if ra != rb {
			parent[ra] = rb
		}
	}
	filesOf := map[string][]string{}
	for _, sc := range scns {
		if sc.Prose != nil {
			continue
		}
		parent[sc.ID] = sc.ID
		filesOf[sc.ID] = stateFilesOf(sc)
	}
	byFile := map[string]string{}
	for _, sc := range scns {
		if sc.Prose != nil {
			continue
		}
		for _, f := range filesOf[sc.ID] {
			if first, ok := byFile[f]; ok {
				union(first, sc.ID)
			} else {
				byFile[f] = sc.ID
			}
		}
	}
	firstStateless := ""
	for _, sc := range scns {
		if sc.Prose != nil || len(filesOf[sc.ID]) > 0 {
			continue
		}
		if firstStateless == "" {
			firstStateless = sc.ID
			continue
		}
		union(firstStateless, sc.ID)
	}
	groups := map[string][]canon.Scenario{}
	for _, sc := range scns {
		if sc.Prose != nil {
			continue
		}
		root := find(sc.ID)
		groups[root] = append(groups[root], sc)
	}
	roots := make([]string, 0, len(groups))
	for root := range groups {
		roots = append(roots, root)
	}
	sort.Strings(roots)
	out := make([][]canon.Scenario, 0, len(roots))
	for _, root := range roots {
		comp := groups[root]
		sort.Slice(comp, func(i, j int) bool { return comp[i].ID < comp[j].ID })
		out = append(out, comp)
	}
	return out
}

func scenarioIDs(comp []canon.Scenario) []string {
	ids := make([]string, 0, len(comp))
	for _, sc := range comp {
		ids = append(ids, sc.ID)
	}
	return ids
}

// groupStateFiles — the shared state files of a component: the
// union of its scenarios' files, in a stable order.
func groupStateFiles(comp []canon.Scenario) []string {
	var out []string
	for _, sc := range comp {
		out = mergeUnique(out, stateFilesOf(sc))
	}
	sort.Strings(out)
	return out
}

func exampleSplit(comps [][]canon.Scenario, next int) string {
	var sb strings.Builder
	sb.WriteString(canondata.T("slot.example.split.head"))
	sb.WriteString("\n")
	for i, comp := range comps {
		sb.WriteString(canondata.T("slot.example.split.card", canondata.M{
			"id":  fmt.Sprintf("%03d", next+i),
			"n":   fmt.Sprintf("%d", i+1),
			"ids": strings.Join(scenarioIDs(comp), ", "),
		}))
		sb.WriteString("\n")
	}
	sb.WriteString(canondata.T("slot.example.split.tail"))
	sb.WriteString("\n")
	return sb.String()
}

// --- delta examples for slots ---

// renderQuestionBatch — the machine render of a clarification batch:
// what the human reads verbatim. An option's price is a machine-
// computable estimate (bytes/4 of the scope the choice fixes), not
// an invented one; the marked default is the machine's
// recommendation (silence applies exactly it). The calling slot
// prints the batch header — here only the body.
func renderQuestionBatch(lang string, findings []canon.LintFinding) string {
	var sb strings.Builder
	for _, f := range findings {
		fmt.Fprintf(&sb, "%s\n", canondata.TFor(lang, "slot.example.question.item", canondata.M{
			"id": f.ID, "question": f.Question,
		}))
		fmt.Fprintf(&sb, "%s\n", canondata.TFor(lang, "slot.example.question.why", canondata.M{"finding": f.Finding}))
		for _, o := range f.Options {
			price := tokenEstimate(o.Label + " " + o.Scope)
			line := canondata.TFor(lang, "slot.example.question.option", canondata.M{
				"id": o.ID, "label": o.Label, "price": fmt.Sprintf("%d", price),
			})
			if o.Default {
				line += canondata.TFor(lang, "slot.example.question.option-default")
			}
			fmt.Fprintln(&sb, line)
		}
	}
	return strings.TrimRight(sb.String(), "\n")
}

// openFindings — the unresolved questions of an open batch: there
// are findings, the batch is not closed, no explicit answer for a
// finding is recorded yet. A conflict without an answer keeps the
// batch open (escalation) and stays here; answered questions are
// not asked again.
func openFindings(state canon.State) []canon.LintFinding {
	if state.Lint == nil || len(state.Lint.Findings) == 0 || state.Clarified {
		return nil
	}
	answered := map[string]bool{}
	for _, c := range state.Clarifications {
		answered[c.Question] = true
	}
	var open []canon.LintFinding
	for _, f := range state.Lint.Findings {
		if !answered[f.ID] {
			open = append(open, f)
		}
	}
	return open
}

// indentLines — an indent on every line: the wish's verbatim text
// in the relay block reads as a separate body, it does not merge
// with the machine's header.
func indentLines(text string) string {
	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
	for i, l := range lines {
		lines[i] = "  " + l
	}
	return strings.Join(lines, "\n")
}

// ChoicePoint — the stop-with-options contract (CLI rc 3/4): a
// choice point is open when the machine has unresolved questions
// for the human; escalation is when among them there is a conflict
// (an explicit live choice, silence does not resolve it by default)
// or the delivery verdict is quarantine (a live decision: lift it
// by changing requirements, grant a budget, refuse). An open
// pending amend is a choice point (silence applies the conservative
// default reject). A choice point is visible from the process code,
// no parsing of surface text is needed.
func (e *Engine) ChoicePoint() (choice, escalate bool) {
	state, err := e.Store.State()
	if err != nil {
		return false, false
	}
	for _, f := range openFindings(state) {
		choice = true
		if f.Dimension == canon.LintConflict {
			escalate = true
		}
	}
	if state.PendingAmend != nil {
		choice = true
	}
	if state.Stage == canon.StageDeliver {
		if v, err := e.RenderVerdict(); err == nil && v.Status == VerdictQuarantine {
			escalate = true
		}
	}
	return choice, escalate
}

// amendChangesBlock — the list of changes of a pending amend: one
// line per change, a fact without judgment; the application plan
// (the effects of the change request) as separate lines after the
// changes: the changes ask the human, the plan explains the price
// of confirmation.
func amendChangesBlock(changes, plan []string) string {
	var sb strings.Builder
	for _, c := range changes {
		fmt.Fprintf(&sb, "  - %s\n", c)
	}
	for _, p := range plan {
		fmt.Fprintf(&sb, "  · %s\n", p)
	}
	return strings.TrimRight(sb.String(), "\n")
}

func exampleIntent(keyNum int) string {
	return canondata.T("slot.example.intent", canondata.M{
		"key": fmt.Sprintf("%03d", keyNum),
	})
}

// exampleChecks — a sample checks table: the full format, minimal fields.
func exampleChecks(keyNum int) string {
	return canondata.T("slot.example.checks", canondata.M{
		"key": fmt.Sprintf("%03d", keyNum),
	})
}

// exampleAssert — a sample assert (below).
// uncoveredProbes — observations without a scenario: a deterministic
// one is covered by an executable scenario with the same command, a
// non-deterministic one by prose containing the command. Failed
// runs do not become a proposal — only an honest line.
func (e *Engine) uncoveredProbes(scns []canon.Scenario) (open []canon.ProbeObservation, failed []canon.ProbeObservation) {
	state, err := e.Store.State()
	if err != nil || len(state.Probes) == 0 {
		return nil, nil
	}
	for _, p := range state.Probes {
		if p.Failed {
			failed = append(failed, p)
			continue
		}
		cmd := strings.Join(p.Command, " ")
		covered := false
		for _, sc := range scns {
			if sc.When != nil && strings.Join(sc.When.Command, " ") == cmd {
				covered = true
				break
			}
			if sc.Prose != nil && strings.Contains(*sc.Prose, cmd) {
				covered = true
				break
			}
		}
		if !covered {
			open = append(open, p)
		}
	}
	return open, failed
}

// probeSummary — the machine slice of observations for the slot context.
func probeSummary(open, failed []canon.ProbeObservation) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s\n", canondata.T("slot.deliver.observed.header", canondata.M{
		"count": fmt.Sprintf("%d", len(open)+len(failed)),
	}))
	for _, p := range open {
		cmd := strings.Join(p.Command, " ")
		if !p.Deterministic {
			fmt.Fprintf(&sb, "  %s — NOT deterministic (%s)\n", cmd, p.Reason)
			continue
		}
		line := fmt.Sprintf("  %s — exit %d", cmd, p.ExitCode)
		if strings.TrimSpace(p.Stdout) != "" {
			line += fmt.Sprintf(", stdout %q", firstLineOf(p.Stdout))
		}
		if len(p.Files) > 0 {
			line += fmt.Sprintf(", files: %s", joinProbePaths(p.Files))
		}
		fmt.Fprintln(&sb, line)
	}
	for _, p := range failed {
		fmt.Fprintf(&sb, "  %s — probe failed: %s\n", strings.Join(p.Command, " "), p.Reason)
	}
	return strings.TrimRight(sb.String(), "\n")
}

func joinProbePaths(files []canon.ProbeFileFact) string {
	paths := make([]string, 0, len(files))
	for _, f := range files {
		paths = append(paths, f.Path)
	}
	return strings.Join(paths, ", ")
}

// exampleBackfillChecks — the machine's proposal of table rows from
// observations: deterministic commands become rows with the observed
// outcomes. The submission stays with the agent: the machine
// proposes, it does not write; non-deterministic observations do not
// become rows and stay in the observation summary.
func exampleBackfillChecks(open []canon.ProbeObservation, keyNum int) string {
	var rows []delta.CheckRow
	for _, p := range open {
		if !p.Deterministic {
			continue
		}
		row := delta.CheckRow{Run: []delta.RunCommand{{Argv: p.Command, IsArgv: true}}}
		rc := p.ExitCode
		row.Rc = &rc
		if strings.TrimSpace(p.Stdout) != "" {
			out := p.Stdout
			row.Out = &out
		}
		if strings.TrimSpace(p.Stderr) != "" {
			errs := p.Stderr
			row.Err = &errs
		}
		for _, f := range p.Files {
			if isStructuredJSON(f.Content) {
				row.State = append(row.State, f.Path+" "+compactJSON(f.Content))
			}
		}
		rows = append(rows, row)
	}
	if len(rows) == 0 {
		return ""
	}
	return renderCheckTable(rows, keyNum)
}

// compactJSON — the canonical one-line notation of a JSON value: no
// spaces after separators, as in brief examples.
func compactJSON(s string) string {
	var v any
	if json.Unmarshal([]byte(s), &v) != nil {
		return s
	}
	data, err := json.Marshal(v)
	if err != nil {
		return s
	}
	return string(data)
}

func exampleAssert(sc canon.Scenario, keyNum int) string {
	return canondata.T("slot.example.assert", canondata.M{
		"key": fmt.Sprintf("%03d", keyNum),
		"scn": strings.ToLower(sc.ID),
		"id":  sc.ID,
	})
}

// blockerReplyTo — the gate for the reply-to of a static-blocker
// fix: the name from the blocker line ("lint: …" → lint, "types: …"
// → types); a surface build failure belongs to suite.
func blockerReplyTo(blockers []string) string {
	for _, b := range blockers {
		if name, _, ok := strings.Cut(b, ": "); ok {
			switch name {
			case "types", "lint":
				return name
			}
		}
	}
	return "suite"
}

func exampleFix(card canon.Card, keyNum int, replyTo string) string {
	return canondata.T("slot.example.fix", canondata.M{
		"key":   fmt.Sprintf("%03d", keyNum),
		"card":  card.ID,
		"reply": replyTo,
	})
}

func exampleCode(card canon.Card, keyNum int) string {
	return canondata.T("slot.example.code", canondata.M{
		"key":  fmt.Sprintf("%03d", keyNum),
		"card": card.ID,
	})
}

// codeStarted — whether code has started: the journal has code or fix submissions.
func (e *Engine) codeStarted() bool {
	for _, en := range e.Journal.All() {
		if en.DeltaKind == deltaKindCode || en.DeltaKind == deltaKindFix {
			return true
		}
	}
	return false
}

// --- the shown-slots registry ---

// servedSlotsPath — the registry of slots whose example block has
// already been shown to the instance: repeat visits to the same
// slot carry one reference line instead of reprinting the example.
// It lives in the instance cache; the next cache key breaks
// together with it.
func (e *Engine) servedSlotsPath() string {
	return filepath.Join(e.Store.Root(), "cache", "served-slots")
}

// slotShown — whether this slot's full example has already been shown.
func (e *Engine) slotShown(slotKey string) bool {
	data, err := os.ReadFile(e.servedSlotsPath())
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(data), "\n") {
		if line == slotKey {
			return true
		}
	}
	return false
}

// markSlotShown marks a slot as shown; the write is an append — an
// unfinished line after a failure matches no key.
func (e *Engine) markSlotShown(slotKey string) {
	if e.slotShown(slotKey) {
		return
	}
	f, err := os.OpenFile(e.servedSlotsPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "%s\n", slotKey)
}

// exampleKeyOf extracts the submission key line from an example: the
// repeat reference carries a fresh key (the number grows with the
// journal) without reprinting the format.
func exampleKeyOf(example string) string {
	for _, line := range strings.Split(example, "\n") {
		if strings.HasPrefix(line, "submission-key: ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "submission-key: "))
		}
	}
	return ""
}

// --- the next cache by input digests ---

// nextRenderVersion — the slot render version: the canon itself
// does not change, but the next text does; the cache key must break
// when the render changes, otherwise an old binary would serve a
// stale slot.
const nextRenderVersion = 21

// cachedNext returns the cached reply if neither the journal nor the
// canon files changed since the last render: a repeated next is
// free, and an external canon edit (forbidden, but observable)
// breaks the cache.
func (e *Engine) cachedNext() (string, bool) {
	key := e.cacheKey()
	keyPath := filepath.Join(e.Store.Root(), "cache", "next-key")
	textPath := filepath.Join(e.Store.Root(), "cache", "next.txt")
	keyData, err := os.ReadFile(keyPath)
	if err != nil || strings.TrimSpace(string(keyData)) != key {
		return "", false
	}
	text, err := os.ReadFile(textPath)
	if err != nil {
		return "", false
	}
	return string(text), true
}

func (e *Engine) storeCachedNext(text string) {
	key := e.cacheKey()
	root := e.Store.Root()
	_ = yamlio.WriteAtomic(filepath.Join(root, "cache", "next-key"), []byte(key))
	_ = yamlio.WriteAtomic(filepath.Join(root, "cache", "next.txt"), []byte(text))
}

// cacheKey — the digest of the render input: the journal head plus
// the digests of all canon files and the state, in a stable order.
func (e *Engine) cacheKey() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "render:%d\n", nextRenderVersion)
	sb.WriteString(e.journalHead())
	sb.WriteString("\n")
	root := e.Store.Root()
	// The shown-slots registry changes the render (a repeat without
	// EXAMPLE) — the cache key must break together with it.
	if data, err := os.ReadFile(filepath.Join(root, "cache", "served-slots")); err == nil {
		fmt.Fprintf(&sb, "served-slots:%s\n", yamlio.Digest(data))
	}
	// The passport conventions change the render (the CONVENTIONS
	// NEEDED block goes away once the recipe is declared) — the key
	// breaks on them too. The kb reference changes the question
	// block (advice with a source) — the key breaks on it too.
	if data, err := os.ReadFile(filepath.Join(root, "passport", "conventions.yaml")); err == nil {
		fmt.Fprintf(&sb, "conventions:%s\n", yamlio.Digest(data))
	}
	if data, err := os.ReadFile(filepath.Join(root, "passport", "kb.yaml")); err == nil {
		fmt.Fprintf(&sb, "kb:%s\n", yamlio.Digest(data))
	}
	// On deliver the slot carries the verdict, and its truth depends
	// on the instance state: the key breaks on the surface too,
	// otherwise external drift would serve a stale verdict from the
	// cache.
	if st, err := e.Store.State(); err == nil && st.Stage == canon.StageDeliver {
		if checks, err := e.Store.Checks(); err == nil {
			if d, ok := surfaceDigest(e, checks); ok {
				fmt.Fprintf(&sb, "surface:%s\n", shortHex(d))
			} else {
				sb.WriteString("surface:gone\n")
			}
		}
	}
	// The freshness contour: the slot carries a CONTOUR line by the
	// instance state — external drift of boundary scope files must
	// break the cache, otherwise stale freshness would live until
	// the next submission.
	if driftedKey, ok := e.boundaryScopeKey(); ok {
		fmt.Fprintf(&sb, "boundary-scope:%s\n", driftedKey)
	}
	var paths []string
	_ = filepath.Walk(filepath.Join(root, "canon"), func(p string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(p, ".yaml") {
			paths = append(paths, p)
		}
		return nil
	})
	paths = append(paths, filepath.Join(root, "state.yaml"))
	sort.Strings(paths)
	for _, p := range paths {
		rel, err := filepath.Rel(root, p)
		if err != nil {
			continue
		}
		// The digest goes through the cache: normalize to a hash,
		// unchanged files are not re-read.
		d, ok := e.digestOf(".punchtape/" + filepath.ToSlash(rel))
		if !ok {
			continue
		}
		fmt.Fprintf(&sb, "%s:%s\n", filepath.ToSlash(rel), d)
	}
	return yamlio.DigestShort([]byte(sb.String()))
}

// journalHead — the digest of the last journal transaction; an
// empty journal — a dash.
func (e *Engine) journalHead() string {
	all := e.Journal.All()
	if len(all) == 0 {
		return "-"
	}
	return shortHex(all[len(all)-1].Transaction)
}

// --- service ---

// nextKeyNum — the number for the submission key example: it grows
// with the journal so that a replaced executor does not copy its
// predecessor's key — a key repeat is an idempotent repeat for the
// machine, not a new submission.
func (e *Engine) nextKeyNum() int {
	return len(e.Journal.All()) + 1
}

// loadForCompile — scenarios and checks for the render.
func (e *Engine) loadForCompile() ([]canon.Scenario, []canon.Check, error) {
	scns, err := e.Store.Scenarios()
	if err != nil {
		return nil, nil, err
	}
	chk, err := e.Store.Checks()
	if err != nil {
		return nil, nil, err
	}
	return scns, chk, nil
}

// loadCanon — requirements and scenarios in one pass.
func (e *Engine) loadCanon() ([]canon.Requirement, []canon.Scenario, error) {
	reqs, err := e.Store.Requirements()
	if err != nil {
		return nil, nil, err
	}
	scns, err := e.Store.Scenarios()
	if err != nil {
		return nil, nil, err
	}
	return reqs, scns, nil
}

func firstLineOf(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	// Cut by runes, not bytes: Cyrillic is two bytes per rune and Han
	// three, a byte limit cut live words in half and twice
	// as early; a cut in the middle of a rune produces a broken
	// glyph.
	r := []rune(s)
	if len(r) <= 80 {
		return s
	}
	return string(r[:80]) + "…"
}

// tokenEstimate — an honest estimate of text volume in tokens: one
// byte per four. The machine has no exact counters and may not lie
// with numbers.
func tokenEstimate(text string) int {
	return len([]byte(text)) / 4
}

// logVerb records a verb call in the ledger: a machine fact, not an
// opinion. The wall is the time of the call itself, not of the
// ledger write. Tokens are the executor's counter when the verb
// brought one (submit with a "# TOKENS" tail); zero — it did not.
func (e *Engine) logVerb(name, outcome string, details map[string]string, wallMs, tokens int64) {
	state, err := e.Store.State()
	if err != nil {
		return
	}
	entry := ledger.NewEntry("verb", state.Stage)
	entry.Calls = 1
	entry.WallMs = wallMs
	entry.Details = map[string]string{"name": name, "outcome": outcome}
	for k, v := range details {
		entry.Details[k] = v
	}
	if tokens > 0 {
		entry.Tokens = &tokens
	}
	_ = e.Ledger.Append(entry)
	// The verb has worked — flush the digest cache: one point
	// per call, there is no race between cache reads and the write.
	e.flushDigestCache()
}

func percent(part, total int) int {
	if total == 0 {
		return 0
	}
	return part * 100 / total
}

func countOutcome(chk []canon.Check, outcome string) int {
	n := 0
	for _, c := range chk {
		if c.Outcome == outcome {
			n++
		}
	}
	return n
}

// missingLinks — executable scenarios without a check: a missing
// link is red by default, staying silent is not allowed.
func missingLinks(scns []canon.Scenario, chk []canon.Check) []string {
	linked := map[string]bool{}
	for _, c := range chk {
		linked[c.Scenario] = true
	}
	var out []string
	for _, sc := range scns {
		if compiler.DegreeOf(sc).Level == compiler.DegreeExecutable && !linked[sc.ID] {
			out = append(out, sc.ID)
		}
	}
	return out
}

// redChecks — the card's red list: one line per check.
func redChecks(card canon.Card, chk []canon.Check) []canon.Check {
	inCard := map[string]bool{}
	for _, id := range card.Scenarios {
		inCard[id] = true
	}
	var out []canon.Check
	for _, c := range chk {
		if inCard[c.Scenario] && c.Outcome == canon.CheckRed {
			out = append(out, c)
		}
	}
	return out
}

// pinRedLines — the card's pin-red lines: green ones pinned on a
// broken surface, they prove the path only by declaration. The
// repair is free.
func pinRedLines(card canon.Card, chk []canon.Check) []string {
	inCard := map[string]bool{}
	for _, id := range card.Scenarios {
		inCard[id] = true
	}
	var out []string
	for _, c := range chk {
		if inCard[c.Scenario] && c.Pin == canon.PinRed {
			out = append(out, canondata.T("gates.probe-red", canondata.M{
				"id": c.ID, "scn": c.Scenario,
			}))
		}
	}
	return out
}

// isStructuredJSON — the value parses as JSON and is an object or
// an array: a bare scalar ("1") does not count as structure, byte
// comparison of a scalar fixes no convention.
func isStructuredJSON(s string) bool {
	var v any
	if json.Unmarshal([]byte(s), &v) != nil {
		return false
	}
	switch v.(type) {
	case map[string]any, []any:
		return true
	}
	return false
}
