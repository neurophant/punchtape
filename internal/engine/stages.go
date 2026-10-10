// The loop process is a data-driven state machine: stages live in
// canon data (stages.yaml), the advance runner executes them. Gates
// are named machine checks: the name→mechanic dictionary is below,
// name literals are verified by mirrors. An unknown mechanic name is
// an honest refusal, not a panic. The loop walks exits while gates
// pass; a failing gate stops it.
package engine

import (
	"fmt"

	"github.com/neurophant/punchtape/internal/canondata"
)

// stageGates — machine checks of transitions, by name from the data.
var stageGates = map[string]func(*builder) bool{
	"spec-accepted":    (*builder).specAccepted,
	"ratify-green":     (*builder).ratifyGreen,
	"spec-resubmitted": (*builder).specResubmitted,
	"always":           func(b *builder) bool { return true },
}

// stageGateFn — the mechanic by data name; an unknown name is
// canon/code drift, an honest refusal (mirrors catch it before
// release, here it is defense in depth).
func stageGateFn(name string) func(*builder) bool {
	fn, ok := stageGates[name]
	if !ok {
		panic(fmt.Sprintf("process data: stage gate %q is unimplemented", name))
	}
	return fn
}

// runProcess — the process runner: advances the stage as far as the
// data's machine checks allow. A gate exit moves the stage; a
// transition is always a consequence of a check over projections,
// never an agent's decision. A stage without an exit is terminal
// (deliver).
func (b *builder) runProcess() {
	stages := canondata.StagesByID()
	for {
		stage, ok := stages[b.state.Stage]
		if !ok || stage.Exit.Gate == "" {
			break
		}
		if !stageGateFn(stage.Exit.Gate)(b) {
			break
		}
		b.state.Stage = stage.Exit.Goto
	}
}
