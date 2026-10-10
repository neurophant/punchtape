// The next verb: the slot — stage, the door, the standing reds, one
// action and a ready context. The output is deterministic: the same
// canon — bit-for-bit the same text. The implement slot carries the
// work list and the red list with measured reasons; the deliver slot
// renders the verdict (the machine's canonical form, an agent's
// retelling is not a verdict).
package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/fsport"
	"github.com/neurophant/punchtape/internal/ledger"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// SlotInfo — the machine coordinates of a slot for the scheduler:
// the stage and the expected document kind. The coordinates are not
// visible to the executor agent: it reads the slot text.
type SlotInfo struct {
	Stage string
	Kind  string // spec | run
	Atom  string
}

// slotCanvas — a thin deterministic text builder of the slot.
type slotCanvas struct {
	sb strings.Builder
}

func (c *slotCanvas) Write(p []byte) (int, error) { return c.sb.Write(p) }
func (c *slotCanvas) String() string              { return c.sb.String() }

// slotRender — the rendered slot.
type slotRender struct {
	Text string
}

// slotPaint — a context body painted onto the canvas.
type slotPaint func(*slotCanvas)

// paintPlain — a plain text body: the text's lines, closed at a line
// boundary (the composition owns spacing — texts carry no edge
// newlines).
func paintPlain(text string) slotPaint {
	return func(c *slotCanvas) {
		if text == "" {
			return
		}
		fmt.Fprint(c, text+"\n")
	}
}

// Next — "what is next".
func (e *Engine) Next() (string, error) {
	return e.nextVerb()
}

func (e *Engine) nextVerb() (string, error) {
	started := timeNow()
	// A cache keyed by the input: while the journal and canon files
	// have not changed, a repeated next is free.
	if cached, ok := e.cachedNext(); ok {
		details := map[string]string{
			"slot": "cached", "context-bytes": fmt.Sprintf("%d", len(cached)),
		}
		e.ensureCurrentSet()
		e.logVerb("next", "ok", details,
			timeSince(started).Milliseconds(), cached)
		return cached, nil
	}

	state, err := e.Store.State()
	if err != nil {
		return "", err
	}
	r, _, err := e.renderNext(state)
	if err != nil {
		return "", err
	}
	e.storeCachedNext(r)
	e.ensureCurrentSet()
	e.ensureSpecRender()
	e.logVerb("next", "ok", map[string]string{"slot": "full", "context-bytes": fmt.Sprintf("%d", len(r.Text))},
		timeSince(started).Milliseconds(), r.Text)
	return r.Text, nil
}

// runsHappened — whether any run submission was applied: the
// implement slot tells the first run from the repairs.
func (e *Engine) runsHappened() bool {
	for _, en := range e.Journal.All() {
		if en.DeltaKind == "run" {
			return true
		}
	}
	return false
}

func (e *Engine) renderNext(state canon.State) (slotRender, SlotInfo, error) {
	return e.renderNextMode(state, slotFull)
}

// renderNextMode — the slot in a surface mode: full (the default
// tables), quiet (the same lines without the CONTEXT body), delta
// (the standings delta since the last slot + the action). Every
// mode writes the standings baseline — the delta's ground is the
// LAST SLOT: any next surface anchors it, while a submit ack is not
// a slot and never re-anchors.
func (e *Engine) renderNextMode(state canon.State, mode slotMode) (slotRender, SlotInfo, error) {
	if mode == slotDelta {
		standings := e.currentStandings(state)
		text, err := e.deltaSlot(state, standings, e.readSlotBaseline())
		if err != nil {
			return slotRender{}, SlotInfo{Stage: state.Stage}, err
		}
		r := slotRender{Text: text}
		e.writeSlotBaseline(standings)
		return r, SlotInfo{Stage: state.Stage}, nil
	}
	var info SlotInfo
	var r slotRender
	info.Stage = state.Stage
	c := &slotCanvas{}
	fmt.Fprintf(c, "STAGE: %s\n", state.Stage)
	// The submission door: the machine's one file, named by its path
	// — the console carries commands, statuses and paths, never
	// submission content.
	fmt.Fprintf(c, "%s\n", e.inboxStatusLine())
	// The contract line: the full contract rode the
	// birth reply once; every later render carries the one-line
	// reference — and the full bytes again, the moment the tree's
	// copy is gone or does not digest to the birth record.
	fmt.Fprintf(c, "%s\n", e.contractRender(state))

	action, gates, context, err := e.slotAction(state, c, &info)
	if err != nil {
		return r, info, err
	}

	// The compaction trigger: the journal threshold is data-driven —
	// a recommendation line to hand off the session; not a blocker:
	// silence continues, rc does not change.
	if n := len(e.Journal.All()); n >= canondata.Limit("compact.journal-entries") {
		fmt.Fprintf(c, "%s\n", canondata.T("slot.session.handoff", canondata.M{
			"n":         strconv.Itoa(n),
			"threshold": strconv.Itoa(canondata.Limit("compact.journal-entries")),
		}))
	}

	fmt.Fprintf(c, "ACTION: %s\n", action)
	if mode != slotQuiet || state.Stage == canon.StageDeliver {
		// The quiet slot drops the CONTEXT body — except the verdict
		// at deliver: the closing's answer is not boilerplate.
		fmt.Fprint(c, "CONTEXT:\n")
		context(c)
	}
	fmt.Fprintf(c, "GATES ON SUBMIT: %s", gates)
	r = slotRender{Text: c.String()}
	e.writeSlotBaseline(e.currentStandings(state))
	return r, info, nil
}

// contractRender — the slot's contract line: the one-line reference
// while the tree's AGENTS.md digests to the birth record; the FULL
// contract bytes when the copy is gone, changed, or the instance
// predates the record (the contract law: the hand pays for the whole
// contract once, at birth — the anti-tamper re-render is not an
// option).
func (e *Engine) contractRender(state canon.State) string {
	data, err := fsport.Bind.ReadFile(filepath.Join(e.Workdir, "AGENTS.md"))
	if err != nil || state.ContractSha == "" {
		return agentsRefText()
	}
	sha := yamlio.Digest(data)
	if sha != state.ContractSha {
		return agentsRefText()
	}
	return canondata.T("cli.next.contract-ref", canondata.M{"sha": sha[:16]})
}

// slotAction — the per-stage body: the stage's state lines and lists
// painted to the canvas, and the action, gates and CONTEXT body for
// the caller to compose (the byte order of the full render is the
// machine's contract: ACTION, then CONTEXT, then GATES).
func (e *Engine) slotAction(state canon.State, c *slotCanvas, info *SlotInfo) (string, string, slotPaint, error) {
	var action, gates string
	var context slotPaint
	switch state.Stage {
	case canon.StageIntake:
		// The single intake slot: the brief is fixed, the spec is the
		// executor's first document — one submission, the whole spec.
		if state.Wish != nil {
			fmt.Fprintf(c, "%s\n", canondata.T("slot.intake.state", canondata.M{
				"text": firstLineOf(*state.Wish),
			}))
		}
		action = canondata.T("slot.intake.action")
		context = paintPlain(canondata.T("slot.intake.context"))
		gates = canondata.T("slot.gates.none")
		info.Kind = "spec"
	case canon.StageImplement:
		return e.implementSlot(c, info)
	case canon.StageDeliver:
		// The verdict is the case's latest frozen snapshot — the
		// machine's canonical format; an agent's retelling is not a
		// verdict. The case lives on: the action names the same door.
		verdict, err := e.RenderVerdict()
		if err != nil {
			return action, gates, paintPlain(""), err
		}
		action = canondata.T("slot.deliver.action", canondata.M{
			"version": strconv.Itoa(e.verdictVersion()),
		})
		context = e.paintVerdict(verdict)
		gates = "—"
	default:
		// An honest render is cheaper than a panic on a stage the
		// render does not know: the empty body.
		context = paintPlain("")
	}
	return action, gates, context, nil
}

// implementSlot — the implement stage's body: the state line and the
// CHANGED list (the door law: only what moved since the last
// reply — never the whole corpus; the full table is the canon's,
// report --full), and the action/gates/context for the caller.
func (e *Engine) implementSlot(c *slotCanvas, info *SlotInfo) (string, string, slotPaint, error) {
	atoms, err := e.Store.Atoms()
	if err != nil {
		return "", "", paintPlain(""), err
	}
	open := openAtoms(atoms)
	fmt.Fprintf(c, "%s\n", canondata.T("slot.implement.state", canondata.M{
		"atoms": strconv.Itoa(len(atoms)),
		"rows":  strconv.Itoa(rowCount(atoms)),
		"open":  strconv.Itoa(len(open)),
	}))
	// The changed list: the door carries the rows whose standing moved
	// since the LAST SLOT (any next surface anchors the one baseline;
	// submit acks carry no slot and do not move it ),
	// retirements included. A first-ever render names the whole
	// standing (everything is new); a corpus that stood says so in one
	// line.
	state, err := e.Store.State()
	if err != nil {
		return "", "", paintPlain(""), err
	}
	base := e.readSlotBaseline()
	changes := standingsChanges(ptr(e.currentStandings(state)), base)
	if base == nil {
		fmt.Fprintln(c, canondata.T("slot.delta.no-baseline"))
	}
	if len(changes) > 0 {
		fmt.Fprintf(c, "%s\n", canondata.T("slot.implement.changed.header", canondata.M{
			"count": strconv.Itoa(len(changes)),
		}))
		for i, ch := range changes {
			if i >= canondata.Limit("slot.list-budget") {
				fmt.Fprintf(c, "%s\n", canondata.T("slot.list.more", canondata.M{
					"count": strconv.Itoa(len(changes) - canondata.Limit("slot.list-budget")),
				}))
				break
			}
			fmt.Fprintf(c, "%s\n", canondata.T("slot.implement.changed.line", canondata.M{
				"label": ch.label, "standing": ch.standing, "prev": ch.prev,
			}))
		}
	} else {
		fmt.Fprintln(c, canondata.T("slot.implement.changed.none"))
	}
	var action string
	var context slotPaint
	if len(open) > 0 {
		focus := open[0]
		info.Atom = focus.ID
		action = canondata.T("slot.implement.action", canondata.M{"atom": focus.ID})
		context = e.paintAtomCard(focus)
	} else {
		action = canondata.T("slot.implement.action.ratify")
		context = paintPlain(canondata.T("slot.implement.context.green"))
	}
	if !e.runsHappened() {
		// The spec was accepted, nothing has run yet: the rows wait
		// for the code.
		fmt.Fprintln(c, canondata.T("slot.implement.await-code"))
	}
	info.Kind = "run"
	return action, "suite, tree", context, nil
}

// ptr — the standings address of a value.
func ptr(s slotStandings) *slotStandings { return &s }

// openAtoms — the not-green atoms in id order (no rows yet, or red
// rows): the work list of the stage.
func openAtoms(atoms []canon.Atom) []canon.Atom {
	var open []canon.Atom
	for _, a := range atoms {
		if !a.Green() {
			open = append(open, a)
		}
	}
	return open
}

func rowCount(atoms []canon.Atom) int {
	n := 0
	for _, a := range atoms {
		n += len(a.Rows)
	}
	return n
}

// paintAtomCard — the focus atom's DOOR card (the door law): what to
// realize (the prose), the one status line, and the RED rows in full
// (their measured reasons are the work list); the green rows are a
// counter — the full table of every row is the canon's
// (punchtape report --full). Nothing about structure, layout or
// files — the code is the executor's free territory.
func (e *Engine) paintAtomCard(a canon.Atom) slotPaint {
	return func(c *slotCanvas) {
		fmt.Fprintf(c, "%s\n", canondata.T("slot.instruction.atom", canondata.M{
			"id": a.ID, "name": a.Name,
		}))
		fmt.Fprintf(c, "%s\n", canondata.T("slot.instruction.prose", canondata.M{
			"text": indentLines(a.Prose),
		}))
		green, red, never := 0, 0, 0
		for _, row := range a.Rows {
			switch row.Outcome {
			case canon.RowGreen:
				green++
			case canon.RowRed:
				red++
			default:
				never++
			}
		}
		fmt.Fprintf(c, "%s\n", canondata.T("slot.instruction.table.status", canondata.M{
			"count": strconv.Itoa(len(a.Rows)),
			"green": strconv.Itoa(green), "red": strconv.Itoa(red), "never": strconv.Itoa(never),
		}))
		shown := 0
		for i, row := range a.Rows {
			if row.Outcome != canon.RowRed {
				continue
			}
			if shown >= canondata.Limit("slot.list-budget") {
				fmt.Fprintf(c, "%s\n", canondata.T("slot.list.more", canondata.M{
					"count": strconv.Itoa(red - canondata.Limit("slot.list-budget")),
				}))
				break
			}
			reason := oneLineValue(row.Reason)
			fmt.Fprintf(c, "%s\n", canondata.T("slot.instruction.red.row", canondata.M{
				"id":     canon.RowLabel(a.ID, i+1),
				"run":    oneLineValue(row.CmdForm()),
				"want":   oneLineValue(expectationLine(row)),
				"reason": reason,
			}))
			shown++
		}
		if green > 0 {
			fmt.Fprintf(c, "%s\n", canondata.T("slot.instruction.green.counter", canondata.M{
				"count": strconv.Itoa(green),
			}))
		}
	}
}

// expectationLine — the row's expectations as one line.
func expectationLine(row canon.Row) string {
	parts := []string{}
	if row.Out != nil {
		parts = append(parts, "out "+row.Out.Form())
	}
	if row.Err != nil {
		parts = append(parts, "err "+row.Err.Form())
	}
	wantRc := 0
	if row.Rc != nil {
		wantRc = *row.Rc
	}
	parts = append(parts, "rc "+strconv.Itoa(wantRc))
	for _, st := range row.State {
		if st.Text != nil {
			parts = append(parts, fmt.Sprintf("state %s text %s", st.Path, st.Text.Form()))
			continue
		}
		parts = append(parts, fmt.Sprintf("state %s from-run %s", st.Path, st.FromRun))
	}
	if ref := row.BytesRef(); ref != "" {
		parts = append(parts, "bytes "+ref)
	}
	if ref := row.ProbeRef(); ref != "" {
		parts = append(parts, "probe "+ref)
	}
	if row.Launch != "" {
		parts = append(parts, "launch "+row.Launch)
	}
	if row.Tree != "" {
		parts = append(parts, "tree "+row.Tree)
	}
	return strings.Join(parts, ", ")
}

// oneLineValue — a value's one-line face for console renders: control
// bytes render as escapes, the line never breaks.
func oneLineValue(s string) string {
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}

// paintVerdict — the canonical verdict text as a context body: one
// block of its own (the VERDICT line opens it).
func (e *Engine) paintVerdict(verdict *Verdict) slotPaint {
	return func(c *slotCanvas) {
		fmt.Fprint(c, verdict.Text())
		fmt.Fprint(c, "\n")
	}
}

// indentLines — an indent on every line: the prose block reads as a
// separate body, it does not merge with the machine's header.
func indentLines(text string) string {
	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
	for i, l := range lines {
		lines[i] = "  " + l
	}
	return strings.Join(lines, "\n")
}

// --- the next cache by input digests ---

// nextRenderVersion — the slot render version: the canon itself does
// not change, but the next text does; the cache key must break when
// the render changes, otherwise an old binary would serve a stale
// slot.
const nextRenderVersion = 40

// cachedNext returns the cached reply if neither the journal nor the
// canon files changed since the last render: a repeated next is
// free, and an external canon edit (forbidden, but observable)
// breaks the cache.
func (e *Engine) cachedNext() (string, bool) {
	key := e.cacheKey()
	keyPath := filepath.Join(e.Store.Root(), "cache", "next-key")
	textPath := filepath.Join(e.Store.Root(), "cache", "next.txt")
	keyData, err := fsport.Bind.ReadFile(keyPath)
	if err != nil || strings.TrimSpace(string(keyData)) != key {
		return "", false
	}
	text, err := fsport.Bind.ReadFile(textPath)
	if err != nil {
		return "", false
	}
	return string(text), true
}

// storeCachedNext writes the cached render: the key and the text.
func (e *Engine) storeCachedNext(r slotRender) {
	key := e.cacheKey()
	root := e.Store.Root()
	_ = yamlio.WriteAtomic(filepath.Join(root, "cache", "next-key"), []byte(key))
	_ = yamlio.WriteAtomic(filepath.Join(root, "cache", "next.txt"), []byte(r.Text))
}

// cacheKey — the digest of the render input: the journal head plus
// the digests of all canon files and the state, in a stable order.
func (e *Engine) cacheKey() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "render:%d\n", nextRenderVersion)
	sb.WriteString(e.journalHead())
	sb.WriteString("\n")
	root := e.Store.Root()
	// On deliver the slot carries the verdict, and its truth depends
	// on the instance state: the key breaks on the tree digest too,
	// otherwise external drift would serve a stale verdict from the
	// cache.
	if st, err := e.Store.State(); err == nil && st.Stage == canon.StageDeliver {
		// The live snapshot, not the last recorded checkpoint: the
		// slot carries the verdict, and an externally drifted tree
		// must break the cache — a stale READY from the cache would
		// lie about the current instance.
		if ckp, err := e.snapshotTree(); err == nil {
			fmt.Fprintf(&sb, "surface:%s\n", shortHex(ckp.Digest))
		} else {
			sb.WriteString("surface:gone\n")
		}
	}
	var paths []string
	_ = fsport.Bind.Walk(filepath.Join(root, "canon"), func(p string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(p, ".yaml") {
			paths = append(paths, p)
		}
		return nil
	})
	paths = append(paths, filepath.Join(root, "state.yaml"))
	sort.Strings(paths)
	for _, p := range paths {
		// p is a child of the walked root: Rel cannot fail
		rel, _ := filepath.Rel(root, p)
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

// journalHead — the digest of the last journal transaction; an empty
// journal — a dash.
func (e *Engine) journalHead() string {
	all := e.Journal.All()
	if len(all) == 0 {
		return "-"
	}
	return shortHex(all[len(all)-1].Transaction)
}

// --- service ---

func firstLineOf(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	// Cut by runes, not bytes: Cyrillic is two bytes per rune and Han
	// three, a byte limit cut live words in half and twice as early;
	// a cut in the middle of a rune produces a broken glyph.
	r := []rune(s)
	if len(r) <= 80 {
		return s
	}
	return string(r[:80]) + "…"
}

// logVerb records a verb call in the ledger: a machine fact, not an
// opinion. The wall is the time of the call itself, not of the
// ledger write. reply is the verb's reply text (next/submit); the
// entry carries its length and its render family (the ceremony
// meter: contract | slot | ack); an empty reply — the event measures
// no reply.
func (e *Engine) logVerb(name, outcome string, details map[string]string, wallMs int64, reply string) {
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
	if replyBytes := int64(len(reply)); replyBytes > 0 {
		entry.ReplyBytes = &replyBytes
		entry.Details["render"] = renderCategory(name, reply)
	}
	_ = e.Ledger.Append(entry)
	// The verb has worked — flush the digest cache: one point per
	// call, there is no race between cache reads and the write.
	e.flushDigestCache()
}
