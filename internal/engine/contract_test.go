// The contract economy law: the hand pays for the
// full contract exactly ONCE — at the instance's birth. Every later
// render carries the one-line reference standing on the birth sha;
// a tree whose AGENTS.md is gone or does not digest to the record
// gets the full bytes again (anti-tamper, not an option).
package engine

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/neurophant/punchtape/internal/canon"
)

func TestContractOnceAtBirth(t *testing.T) {
	stack, eng := newInst(t)
	// the birth record lives in the state
	st, err := eng.Store.State()
	if err != nil || st.ContractSha == "" {
		t.Fatalf("the birth contract record: %v %+v", err, st)
	}
	// a later render: the one-line reference, never the full body
	text, err := eng.Next()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "CONTRACT: AGENTS.md in your tree (unchanged; sha256 ") {
		t.Fatalf("the later render carries no contract reference:\n%s", text)
	}
	if strings.Contains(text, "## punchtape - what this is") {
		t.Fatalf("the later render pays for the full contract body:\n%s", text)
	}
	render := func() string {
		t.Helper()
		state, err := eng.Store.State()
		if err != nil {
			t.Fatal(err)
		}
		r, _, err := eng.renderNext(state)
		if err != nil {
			t.Fatal(err)
		}
		return r.Text
	}
	// the tree's copy is tampered: the full contract renders again
	path := filepath.Join(eng.Workdir, "AGENTS.md")
	orig, err := stack.fs.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := stack.fs.WriteFile(path, append(append([]byte{}, orig...), []byte("tampered\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	if text := render(); !strings.Contains(text, "## punchtape - what this is") {
		t.Fatalf("the tampered tree must get the full contract again:\n%s", text)
	}
	// the copy is gone: the full contract renders again
	if err := stack.fs.Remove(path); err != nil {
		t.Fatal(err)
	}
	if text := render(); !strings.Contains(text, "## punchtape - what this is") {
		t.Fatalf("the tree without the contract must get the full bytes:\n%s", text)
	}
	// the birth bytes return: the one-line reference returns
	if err := stack.fs.WriteFile(path, orig, 0o644); err != nil {
		t.Fatal(err)
	}
	if text := render(); !strings.Contains(text, "CONTRACT: AGENTS.md in your tree (unchanged; sha256 ") {
		t.Fatalf("the restored tree must get the reference back:\n%s", text)
	}
}

// The delta surface rides the same law: the standing contract line,
// never the full body.
func TestContractLineInDelta(t *testing.T) {
	stack, eng := newInst(t)
	_ = stack
	state, err := eng.Store.State()
	if err != nil {
		t.Fatal(err)
	}
	text, err := eng.deltaSlot(state, eng.currentStandings(state), e2eBaseline(t, eng))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "CONTRACT: AGENTS.md in your tree (unchanged; sha256 ") {
		t.Fatalf("the delta render carries no contract reference:\n%s", text)
	}
	if strings.Contains(text, "## punchtape - what this is") {
		t.Fatalf("the delta render pays for the full contract body:\n%s", text)
	}
}

// e2eBaseline — a standing baseline read through the cache zone.
func e2eBaseline(t *testing.T, eng *Engine) *slotStandings {
	t.Helper()
	return eng.readSlotBaseline()
}

// The quiet surface rides the line too (the orientation lines stay).
func TestContractLineInQuiet(t *testing.T) {
	stack, eng := newInst(t)
	_ = stack
	text, err := eng.NextWith(SlotModeQuiet)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "CONTRACT: AGENTS.md in your tree") {
		t.Fatalf("the quiet render carries no contract reference:\n%s", text)
	}
}

// An instance older than the record (an empty birth sha) gets the
// full contract: the reference may stand only on a recorded birth.
func TestContractUnrecordedBirth(t *testing.T) {
	stack, eng := newInst(t)
	st, err := eng.Store.State()
	if err != nil {
		t.Fatal(err)
	}
	st.ContractSha = ""
	if err := eng.Store.PutState(st); err != nil {
		t.Fatal(err)
	}
	r, _, err := eng.renderNext(canon.State{Stage: st.Stage, ContractSha: ""})
	if err != nil {
		t.Fatal(err)
	}
	text := r.Text
	_ = stack
	if !strings.Contains(text, "## punchtape - what this is") {
		t.Fatalf("the unrecorded birth must render the full contract:\n%s", text)
	}
}
