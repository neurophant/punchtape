// The loop process is a data-driven state machine: stages, human
// choice points and exits live in canon data (scenarios.yaml), the
// advance runner executes them. Gates, holds and entries are named
// machine checks: the name→mechanic dictionaries are below, name
// literals are verified by mirrors. Ambiguities are resolved by a
// human from outside — an open point holds the stage, the runner
// does not decide. A liveness safeguard: in one pass the runner
// visits each stage at most once — more means a cycle in the process
// data (legitimate rollbacks return as a new submission, not another
// turn of this cycle); an unknown mechanic name is an honest
// refusal, not a panic. A process error accumulates in the builder
// and brings the submission down with an honest failure.
package engine

import (
	"fmt"

	"github.com/neurophant/punchtape/internal/canondata"
)

// stageGates — machine checks of transitions, by name from the data.
var stageGates = map[string]func(*builder) bool{
	"intent-accepted":      func(b *builder) bool { return b.state.IntentAccepted },
	"spec-complete":        (*builder).specComplete,
	"scenarios-executable": (*builder).scenariosReady,
	"cards-closed":         (*builder).cardGreen,
	"always":               func(*builder) bool { return true },
	"cards-reopened":       (*builder).cardsReopened,
}

// stageChoices — human choice points by data name: true means "the
// choice is open, the stage is held".
var stageChoices = map[string]func(*builder) bool{
	// An open question batch holds the stage: the submission's reply
	// must return the questions to the human (the clarification loop
	// lives inside the reply to the table).
	"open-questions": func(b *builder) bool {
		return b.state.Lint != nil && len(b.state.Lint.Findings) > 0 && !b.state.Clarified
	},
	// A spec draft (cheap executor) does not become the spec by
	// itself: until the hand's review the stage is held; the hand
	// submitting the table is the review.
	"draft-review": func(b *builder) bool {
		return b.e.draftPending() && !(b.appliedChecks && !b.e.draftSubmit)
	},
	// An open pending amend holds the stage: the fate of already-run
	// expectations is decided by the human (or by silence — with the
	// next submission).
	"pending-amend": func(b *builder) bool { return b.state.PendingAmend != nil },
}

// stageEntries — machine actions on stage entry, by data name.
var stageEntries = map[string]func(*builder){
	"whole-product-card": (*builder).enterImplement,
}

// stageGateFn — the mechanic by data name; an unknown name is
// canon/code drift, an honest refusal (mirrors catch it before
// release, here it is defense in depth).
func stageGateFn(name string) (func(*builder) bool, error) {
	fn, ok := stageGates[name]
	if !ok {
		return nil, fmt.Errorf("process data: stage gate %q is unimplemented", name)
	}
	return fn, nil
}

func stageChoiceFn(name string) (func(*builder) bool, error) {
	fn, ok := stageChoices[name]
	if !ok {
		return nil, fmt.Errorf("process data: stage choice %q is unimplemented", name)
	}
	return fn, nil
}

func stageEntryFn(name string) (func(*builder), error) {
	fn, ok := stageEntries[name]
	if !ok {
		return nil, fmt.Errorf("process data: stage entry %q is unimplemented", name)
	}
	return fn, nil
}

// runProcess — the process runner: advances the stage as far as the
// data's machine checks allow. An open choice point holds the stage;
// a gate exit moves it and runs the entry action; a transition is
// always a consequence of a check over projections, never an agent's
// decision.
func (b *builder) runProcess() {
	stages := canondata.StagesByID()
	for visits := 0; ; visits++ {
		if visits > len(stages) {
			b.processErr = fmt.Errorf("process data: stage cycle at %q", b.state.Stage)
			return
		}
		stage, ok := stages[b.state.Stage]
		if !ok {
			break
		}
		held := false
		for _, name := range stage.Choices {
			fn, err := stageChoiceFn(name)
			if err != nil {
				b.processErr = err
				return
			}
			if fn(b) {
				held = true
				break
			}
		}
		if held {
			break
		}
		gate, err := stageGateFn(stage.Exit.Gate)
		if err != nil {
			b.processErr = err
			return
		}
		if !gate(b) {
			break
		}
		b.state.Stage = stage.Exit.Goto
		if stage.Exit.Entry != "" {
			entry, err := stageEntryFn(stage.Exit.Entry)
			if err != nil {
				b.processErr = err
				return
			}
			entry(b)
		}
	}
}
