// The stage machine: transitions only by the machine and only by a machine check.
package engine

import "github.com/neurophant/punchtape/internal/delta"

// advance moves the stage as far as the machine checks allow:
// the runner executes the process data (stages.yaml), every loop
// here is part of the same transaction.
func (b *builder) advance() {
	b.runProcess()
	b.putState()
}

// specAccepted — the first accepted spec opened the work: the corpus
// has behaviors.
func (b *builder) specAccepted() bool {
	return len(b.atoms) > 0
}

// specResubmitted — the deliver reopen gate: a ratification froze
// the case's verdict, but the case LIVES ON — the next spec version
// through the same door reopens the work (a run alone does not: it
// has nothing new to verify against).
func (b *builder) specResubmitted() bool {
	return b.kind == delta.KindSpec
}

// ratifyGreen — the ratification gate: THIS submission carries
// ratify: true and every atom's rows stand green on it. The
// signature law owns the freshness proof before the stage machine
// runs (the predicate signs the recorded green or refuses with the
// names); this gate is the second lock — a corpus with any open row
// does not deliver, whatever declared it ready.
func (b *builder) ratifyGreen() bool {
	if !b.ratifying || len(b.atoms) == 0 {
		return false
	}
	for _, a := range b.atoms {
		if !a.Green() {
			return false
		}
	}
	return true
}
