// The ceremony meter: the machine's reply bytes to the
// hand, summed by render family. The families ride the verb entries of
// the cell's ledger (details.render: contract | slot | ack) — the
// reading below is deterministic arithmetic over recorded facts, never
// an estimate. The four published categories:
//
//	contract  — replies carrying the FULL contract bytes: the birth
//	            injection and the anti-tamper re-render (the contract
//	            one full payment, now visible as a number);
//	spec      — slot replies at intake, the slot that asks the spec;
//	report    — slot replies at implement and deliver, the slots that
//	            carry the work table and the verdict;
//	door-echo — the submit acks (the thin-echo surface).
package engine

import (
	"strings"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/ledger"
)

// CeremonyBytes — the ceremony counters over the whole ledger: reply
// bytes by category, zeros included (an absent family is a fact too).
func (e *Engine) CeremonyBytes() map[string]int64 {
	out := map[string]int64{
		"contract": 0, "spec": 0, "report": 0, "door-echo": 0,
	}
	for _, en := range e.Ledger.All() {
		if en.Event != "verb" || en.ReplyBytes == nil {
			continue
		}
		out[ceremonyCategory(en)] += *en.ReplyBytes
	}
	return out
}

// ceremonyCategory — one reply-bearing verb entry's category. The
// render detail is the ground (written by logVerb at the reply); the
// slot family splits by the entry's stage. Any other render value —
// the door's echo: the acks of submissions.
func ceremonyCategory(en ledger.Entry) string {
	switch en.Details["render"] {
	case "contract":
		return "contract"
	case "slot":
		if en.Stage == canon.StageIntake {
			return "spec"
		}
		return "report"
	default:
		return "door-echo"
	}
}

// renderCategory — the reply's render family for the ledger detail:
// a submit reply is the door's echo; a next reply carrying the full
// contract bytes (the reference marker rides only them) is the
// contract family; every other next reply is a slot.
func renderCategory(name, reply string) string {
	if name == "submit" {
		return "ack"
	}
	if strings.Contains(reply, refMarker) {
		return "contract"
	}
	return "slot"
}
