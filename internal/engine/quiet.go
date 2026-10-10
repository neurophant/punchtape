// The quiet output and the batch call: the machine
// re-printed every table on every call — the executors re-read the
// full slot each reply. The modes are REPRESENTATION only: --quiet
// drops the slot body (the orientation lines stay), --delta prints
// only what changed since the last slot (the standings baseline
// lives in the cache zone — a render fact of the next verb, never
// canon; a submit ack is not a slot and never moves the anchor),
// --run-after chains the spec's run into the same door call. The
// journal and the canon are byte-identical across the modes (the
// clock is the only divergence); the default full render does not
// change.
package engine

import (
	"path/filepath"
	"strconv"
	"strings"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/delta"
	"github.com/neurophant/punchtape/internal/fsport"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// Slot mode spellings (the CLI's flags map here).
const (
	SlotModeFull  = ""
	SlotModeQuiet = "quiet"
	SlotModeDelta = "delta"
)

// slotMode — the reply's slot surface.
type slotMode int

const (
	slotFull slotMode = iota
	slotQuiet
	slotDelta
)

func parseSlotMode(s string) slotMode {
	switch s {
	case SlotModeQuiet:
		return slotQuiet
	case SlotModeDelta:
		return slotDelta
	default:
		return slotFull
	}
}

// SubmitOpts — the door call's surface options: the reply mode and
// the batch (chain the run of the just-accepted spec into the same
// call — one reply, the machine's own transactions).
type SubmitOpts struct {
	Mode     string
	RunAfter bool
}

// SubmitSetOpts — the door with surface options. The batch is an
// ORCHESTRATION, not a new transaction shape: the spec and the run
// land as the same two transactions a manual second call would have
// written (the same door, the same documents, the same bytes).
func (e *Engine) SubmitSetOpts(source string, opts SubmitOpts) (string, int) {
	rel := filepath.ToSlash(filepath.Clean(source))
	if rel != inboxCurrent {
		reply := reject("door", canondata.T("submit.error.not-current", canondata.M{
			"given": rel, "want": inboxCurrent,
		}))
		e.logVerb("submit", "rejected", map[string]string{"reason": "not-the-door-file"}, 0, reply)
		return reply, Rejected
	}
	raw, err := fsport.Bind.ReadFile(filepath.Join(e.Workdir, rel))
	if err != nil {
		return canondata.T("cli.fail.reading", canondata.M{
			"source": rel, "err": firstLine(err),
		}), Failure
	}
	mode := parseSlotMode(opts.Mode)
	// The batch chains the run of a just-accepted SPEC; a run-kind
	// document is already the run itself. The kind is known before
	// the submit (one decode, the same document): the spec's own slot
	// tail stays suppressed only when the chained run's reply — the
	// fresher slot — follows it.
	specChained := false
	if d, derr := delta.Decode(raw); derr == nil {
		specChained = opts.RunAfter && d.Kind == delta.KindSpec
	}
	reply, code := e.submitCoreMode(raw, mode, !specChained)
	if code == Rejected {
		reply += "\n" + canondata.T("submit.error.stays", canondata.M{"manifest": rel})
	}
	if code == Accepted && specChained {
		// The batch: the suite of the just-accepted version runs in
		// the same call — the synthetic document is exactly the bare
		// run the executor would have written next ("kind: run\n"),
		// so the journal carries the very same transaction.
		runReply, runCode := e.submitCoreMode([]byte("kind: run\n"), mode, true)
		reply = reply + "\n" + runReply
		code = runCode
	}
	if code == Accepted && !e.lastIdempotent {
		e.resetCurrentSet()
	}
	return reply, code
}

// NextWith — the slot in a surface mode ("" — the default full
// render, the same bytes as Next).
func (e *Engine) NextWith(mode string) (string, error) {
	switch parseSlotMode(mode) {
	case slotQuiet:
		return e.nextMode(slotQuiet)
	case slotDelta:
		return e.nextMode(slotDelta)
	default:
		return e.Next()
	}
}

// nextMode — a quiet or delta slot render (never the cache: the
// fresh compose is the point; the baseline updates with it).
func (e *Engine) nextMode(mode slotMode) (string, error) {
	started := timeNow()
	state, err := e.Store.State()
	if err != nil {
		return "", err
	}
	r, _, err := e.renderNextMode(state, mode)
	if err != nil {
		return "", err
	}
	e.ensureCurrentSet()
	e.logVerb("next", "ok", map[string]string{
		"slot": modeName(mode), "context-bytes": strconv.Itoa(len(r.Text)),
	}, timeSince(started).Milliseconds(), r.Text)
	return r.Text, nil
}

func modeName(mode slotMode) string {
	switch mode {
	case slotQuiet:
		return "quiet"
	case slotDelta:
		return "delta"
	default:
		return "full"
	}
}

// --- the standings baseline (the delta's ground) ---

// standingLine — one row's standing at a render: green (fresh at
// unchanged inputs), stale (a green that will re-run — its inputs
// moved), red, never (not measured yet).
type standingLine struct {
	Label    string `yaml:"label"`
	Standing string `yaml:"standing"`
}

// slotStandings — the baseline the next delta renders against: the
// stage and every live row's standing, in table order (atoms by id,
// rows insertion-stable — the same order every table renders).
type slotStandings struct {
	Stage string         `yaml:"stage"`
	Rows  []standingLine `yaml:"rows"`
}

// currentStandings — the corpus' standings now (table order).
func (e *Engine) currentStandings(state canon.State) slotStandings {
	out := slotStandings{Stage: state.Stage}
	b := newBuilder(e, state)
	if err := b.load(); err != nil {
		return out
	}
	for _, id := range sortedIDs(b.atoms) {
		for i, row := range b.atoms[id].Rows {
			out.Rows = append(out.Rows, standingLine{
				Label:    canon.RowLabel(id, i+1),
				Standing: rowStandingOf(b, row),
			})
		}
	}
	return out
}

// rowStandingOf — one row's standing word.
func rowStandingOf(b *builder, row canon.Row) string {
	switch {
	case row.Outcome == canon.RowRed:
		return "red"
	case row.Outcome != canon.RowGreen:
		return "never"
	case b.rowFresh(row):
		return "green"
	default:
		return "stale"
	}
}

// writeSlotBaseline — the render's ground for the next delta: the
// cache zone, a render fact (never canon, never the journal).
func (e *Engine) writeSlotBaseline(s slotStandings) {
	data, _ := yamlio.Marshal(s)
	_ = yamlio.WriteAtomic(filepath.Join(e.Store.Root(), "cache", "slot-standings.yaml"), data)
}

// readSlotBaseline — the last render's standings; nil when no render
// ever happened (the first delta says so honestly).
func (e *Engine) readSlotBaseline() *slotStandings {
	data, err := fsport.Bind.ReadFile(filepath.Join(e.Store.Root(), "cache", "slot-standings.yaml"))
	if err != nil {
		return nil
	}
	var s slotStandings
	if yamlio.DecodeDoc(data, &s) != nil {
		return nil
	}
	return &s
}

// standingChange — one row's move between two slots: the label,
// the standing now and the standing the last slot recorded ("new" —
// the row did not exist; "retired" — it is gone).
type standingChange struct {
	label, standing, prev string
}

// standingsChanges — the rows whose standing moved between the last
// slot's baseline and now (retirements included). The ONE changed-
// list law: the door carries what moved, never the
// whole corpus; the full table is the canon's (report --full).
func standingsChanges(now, base *slotStandings) []standingChange {
	var out []standingChange
	if base == nil {
		for _, r := range now.Rows {
			out = append(out, standingChange{r.Label, r.Standing, "new"})
		}
		return out
	}
	had := map[string]string{}
	for _, r := range base.Rows {
		had[r.Label] = r.Standing
	}
	seen := map[string]bool{}
	for _, r := range now.Rows {
		seen[r.Label] = true
		prev, ok := had[r.Label]
		switch {
		case !ok:
			out = append(out, standingChange{r.Label, r.Standing, "new"})
		case prev != r.Standing:
			out = append(out, standingChange{r.Label, r.Standing, prev})
		}
	}
	for _, r := range base.Rows {
		if !seen[r.Label] {
			out = append(out, standingChange{r.Label, "retired", r.Standing})
		}
	}
	return out
}

// writeLine — one text line onto a builder without the concatenation
// allocation: the text, then the newline, two writes.
func writeLine(sb *strings.Builder, s string) {
	sb.WriteString(s)
	sb.WriteString("\n")
}

// deltaSlot — the delta slot: the stage (with its move when it
// moved), the rows whose standing changed since the last slot
// (retirements included), the action — and the verdict at deliver
// (the closing's point; it is not boilerplate to cut).
func (e *Engine) deltaSlot(state canon.State, standings slotStandings, base *slotStandings) (string, error) {
	var sb strings.Builder
	if base != nil && base.Stage != state.Stage {
		writeLine(&sb, canondata.T("slot.delta.stage-was", canondata.M{
			"stage": state.Stage, "prev": base.Stage,
		}))
	} else {
		sb.WriteString("STAGE: ")
		sb.WriteString(state.Stage)
		sb.WriteString("\n")
	}
	// The standing contract line rides every mode; the
	// delta body below stays what it was — the changes only.
	writeLine(&sb, e.contractRender(state))
	if base == nil {
		writeLine(&sb, canondata.T("slot.delta.no-baseline"))
	}
	changes := standingsChanges(&standings, base)
	if len(changes) > 0 {
		budget := canondata.Limit("slot.list-budget")
		writeLine(&sb, canondata.T("slot.delta.changed.header", canondata.M{
			"count": strconv.Itoa(len(changes)),
		}))
		for i, c := range changes {
			if i >= budget {
				writeLine(&sb, canondata.T("slot.list.more", canondata.M{
					"count": strconv.Itoa(len(changes) - budget),
				}))
				break
			}
			writeLine(&sb, canondata.T("slot.delta.changed.line", canondata.M{
				"label": c.label, "standing": c.standing, "prev": c.prev,
			}))
		}
	} else if base != nil {
		writeLine(&sb, canondata.T("slot.delta.changed.none"))
	}
	if state.Stage == canon.StageDeliver {
		if v, err := e.RenderVerdict(); err == nil {
			writeLine(&sb, v.Text())
		}
	}
	// The action line orients every mode; the throwaway canvas holds
	// the stage's list paints (the delta names the changes itself).
	cc := &slotCanvas{}
	var si SlotInfo
	action, _, _, err := e.slotAction(state, cc, &si)
	if err != nil {
		return "", err
	}
	sb.WriteString("ACTION: ")
	sb.WriteString(action)
	sb.WriteString("\n")
	return sb.String(), nil
}
