// Machine handoff: "done/next/ready prompt" is computed from state,
// the ledger, the canon and the verdict — bypassing the agent's
// memory (an agent's retelling is not a handoff). The human block is
// in the language of the wish: it is copied into a new session whole.
package engine

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/canondata"
)

// WhyHandoff — the computed handoff: instance facts and a ready
// continuation prompt. Nothing is stored — the view is assembled on
// the fly.
func (e *Engine) WhyHandoff() (string, error) {
	state, err := e.Store.State()
	if err != nil {
		return "", err
	}
	lang := state.Language
	chk, err := e.Store.Checks()
	if err != nil {
		return "", err
	}
	green := 0
	for _, c := range chk {
		if c.Outcome == canon.CheckGreen {
			green++
		}
	}
	boundaries, err := e.Store.Boundaries()
	if err != nil {
		return "", err
	}
	verified := 0
	for _, bd := range boundaries {
		if bd.Verdict != nil {
			verified++
		}
	}
	verdict := canondata.TFor(lang, "why.handoff.verdict-none")
	if state.Stage == canon.StageDeliver {
		if v, err := e.RenderVerdict(); err == nil {
			verdict = v.Status
		}
	}

	// Open points of the next session: everything that holds a choice
	// or work — from machine facts, not from memories.
	open := openFindings(state)
	questions := canondata.TFor(lang, "why.handoff.none")
	if n := len(open); n > 0 {
		questions = strconv.Itoa(n)
	}
	pending := canondata.TFor(lang, "why.handoff.none")
	if state.PendingAmend != nil {
		pending = strconv.Itoa(len(state.PendingAmend.Changes))
	}
	drifted := canondata.TFor(lang, "why.handoff.none")
	if ids := e.driftedBoundaries(); len(ids) > 0 {
		drifted = strings.Join(ids, ", ")
	}
	// The wish in the prompt is VERBATIM and whole: the pasted prompt
	// must be self-contained; truncation would make the new session
	// dependent on the old one's memory. A multi-line wish honestly
	// flows on as lines below — layout is not more important than
	// completeness.
	wish := canondata.TFor(lang, "why.handoff.none")
	if state.Intent != nil {
		wish = *state.Intent
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "%s\n", canondata.TFor(lang, "why.handoff.head"))
	fmt.Fprintf(&sb, "%s\n", canondata.TFor(lang, "why.handoff.done", canondata.M{
		"stage":      state.Stage,
		"deltas":     strconv.Itoa(len(e.Journal.All())),
		"green":      strconv.Itoa(green),
		"total":      strconv.Itoa(len(chk)),
		"verified":   strconv.Itoa(verified),
		"boundaries": strconv.Itoa(len(boundaries)),
		"verdict":    verdict,
	}))
	fmt.Fprintf(&sb, "%s\n", canondata.TFor(lang, "why.handoff.next", canondata.M{
		"stage-action": stageHandoffAction(lang, state.Stage),
		"questions":    questions, "pending": pending, "drift": drifted,
	}))
	fmt.Fprintf(&sb, "%s", canondata.TFor(lang, "why.handoff.prompt", canondata.M{
		"wish": wish,
	}))
	return sb.String(), nil
}

// stageHandoffAction — what is next by stage: explicit references to
// canon keys (the mirror reconciler sees each as the norm).
func stageHandoffAction(lang, stage string) string {
	switch stage {
	case canon.StageIntake:
		return canondata.TFor(lang, "why.handoff.next.intake")
	case canon.StageSpec:
		return canondata.TFor(lang, "why.handoff.next.spec")
	case canon.StageACCompile:
		return canondata.TFor(lang, "why.handoff.next.ac-compile")
	case canon.StageImplement:
		return canondata.TFor(lang, "why.handoff.next.implement")
	case canon.StageDeliver:
		return canondata.TFor(lang, "why.handoff.next.deliver")
	}
	return stage
}
