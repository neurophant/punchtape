// The ceremony meter battery: the cell's ledger
// carries every reply's render family (contract | slot | ack); the
// counters are deterministic sums over the recorded reply bytes —
// the numbers the campaign's economy columns read.
package engine

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/canondata"
)

// driveM5Cycle — a full gate-roundtrip-shaped cycle over the engine:
// birth, the intake slot, the spec, the implement slot, the run, the
// ratification, the deliver slot. Every reply is returned in order
// with the category the caller expects it to land in.
func TestCeremonyCounters(t *testing.T) {
	stack, eng := newInst(t)

	// The birth entry: the contract family, and its reply bytes are
	// the FULL ceremony the CLI printed — the injection line, the
	// contract, the slot.
	var birthBytes int64
	var seenBirth bool
	for _, en := range eng.Ledger.All() {
		if en.Event != "verb" || en.Details["slot"] != "birth" {
			continue
		}
		if seenBirth {
			t.Fatal("two birth entries")
		}
		seenBirth = true
		if en.Details["render"] != "contract" {
			t.Fatalf("birth render: %q", en.Details["render"])
		}
		if en.ReplyBytes == nil {
			t.Fatal("birth carries no reply bytes")
		}
		birthBytes = *en.ReplyBytes
	}
	if !seenBirth {
		t.Fatal("no birth entry")
	}
	wantBirth := len(canondata.T("cli.next.contract-inject")) + 1 + len(AgentsRefText()) + 1
	if birthBytes < int64(wantBirth) {
		t.Fatalf("birth bytes %d < injection+contract %d", birthBytes, wantBirth)
	}

	// The full cycle, every reply recorded with its family.
	type replyFact struct {
		stage  string
		render string
		bytes  int64
	}
	var replies []replyFact
	// the birth entry rides ahead of the driven replies (checked
	// separately above; here it holds its place in the order)
	replies = append(replies, replyFact{stage: canon.StageIntake, render: "contract", bytes: birthBytes})
	record := func(render string, reply string) {
		st, err := eng.Store.State()
		if err != nil {
			t.Fatal(err)
		}
		replies = append(replies, replyFact{stage: st.Stage, render: render, bytes: int64(len(reply))})
	}

	next, err := eng.Next() // the cached birth slot, intake
	if err != nil {
		t.Fatal(err)
	}
	record("slot", next)
	spec, code := submit(t, stack, eng, specDoc("B1", "hello"))
	mustAccept(t, spec, code, "SPEC v1")
	record("ack", spec)
	next, err = eng.Next() // implement
	if err != nil {
		t.Fatal(err)
	}
	record("slot", next)
	run, code := submit(t, stack, eng, "kind: run\n")
	mustAccept(t, run, code, "suite=green")
	record("ack", run)
	ratify, code := submit(t, stack, eng, "kind: run\nratify: true\n")
	mustAccept(t, ratify, code, "VERDICT: READY")
	record("ack", ratify)
	// the summary freezes at the ratification: its counters stand on
	// the ceremony paid by then, not on the replies that follow
	cerAtRatify := eng.CeremonyBytes()
	next, err = eng.Next() // deliver
	if err != nil {
		t.Fatal(err)
	}
	record("slot", next)

	// Every reply-bearing verb entry carries its render family, and
	// the entry's bytes are the reply's own length.
	byIndex := map[int]replyFact{}
	i := 0
	for _, en := range eng.Ledger.All() {
		if en.Event != "verb" || en.ReplyBytes == nil {
			continue
		}
		switch en.Details["render"] {
		case "contract", "slot", "ack":
		default:
			t.Fatalf("verb entry without a render family: %+v", en.Details)
		}
		if i < len(replies) {
			byIndex[i] = replies[i]
			if en.Details["render"] != byIndex[i].render {
				t.Fatalf("entry %d render %q, want %q", i, en.Details["render"], byIndex[i].render)
			}
			if en.Stage != byIndex[i].stage {
				t.Fatalf("entry %d stage %q, want %q", i, en.Stage, byIndex[i].stage)
			}
			if *en.ReplyBytes != byIndex[i].bytes {
				t.Fatalf("entry %d bytes %d, want %d", i, *en.ReplyBytes, byIndex[i].bytes)
			}
		}
		i++
	}

	// The counters are the sums of the recorded render bytes — the
	// four families, zeros included.
	want := map[string]int64{"contract": 0, "spec": 0, "report": 0, "door-echo": 0}
	want["contract"] += birthBytes
	for _, r := range replies {
		switch r.render {
		case "slot":
			if r.stage == canon.StageIntake {
				want["spec"] += r.bytes
			} else {
				want["report"] += r.bytes
			}
		case "ack":
			want["door-echo"] += r.bytes
		}
	}
	got := eng.CeremonyBytes()
	for _, cat := range []string{"contract", "spec", "report", "door-echo"} {
		if got[cat] != want[cat] {
			t.Fatalf("category %s: got %d, want %d (all: %v)", cat, got[cat], want[cat], got)
		}
	}

	// The summary carries the counters of its freeze moment (the
	// ratification) as its ceremony line.
	summary, err := stack.fs.ReadFile(filepath.Join(eng.Store.Root(), "SUMMARY.md"))
	if err != nil {
		t.Fatal(err)
	}
	line := canondata.T("summary.price.ceremony", canondata.M{
		"contract": fmt.Sprintf("%d", cerAtRatify["contract"]),
		"spec":     fmt.Sprintf("%d", cerAtRatify["spec"]),
		"report":   fmt.Sprintf("%d", cerAtRatify["report"]),
		"door":     fmt.Sprintf("%d", cerAtRatify["door-echo"]),
	})
	if !strings.Contains(string(summary), line) {
		t.Fatalf("the summary carries no ceremony line %q:\n%s", line, summary)
	}
}

// The anti-tamper re-render is the contract family too: the tree's
// AGENTS.md broken — the full bytes ride the next fresh reply again
// (the tamper test moves the journal first: a cached next would
// honestly serve the standing slot).
func TestTamperRerenderIsContract(t *testing.T) {
	stack, eng := newInst(t)
	reply, code := submit(t, stack, eng, specDoc("B1", "hello"))
	mustAccept(t, reply, code, "SPEC v1")
	before := eng.CeremonyBytes()["contract"]
	path := filepath.Join(eng.Workdir, "AGENTS.md")
	orig, err := stack.fs.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := stack.fs.WriteFile(path, append(append([]byte{}, orig...), []byte("tampered\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	text, err := eng.Next()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, refMarker) {
		t.Fatalf("the tampered tree must pay the full contract:\n%s", text)
	}
	after := eng.CeremonyBytes()["contract"]
	if after <= before {
		t.Fatalf("the contract family did not grow: %d -> %d", before, after)
	}
	// The reply's own bytes are the growth — one render, one entry.
	if after-before != int64(len(text)) {
		t.Fatalf("growth %d != reply bytes %d", after-before, len(text))
	}
}
