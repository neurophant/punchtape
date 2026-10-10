// Canonical verdict: fixed format, digest signature. The machine
// renders it; an agent's retelling is not a verdict. READY needs
// every atom's rows green on the ratifying run and a tree that did
// not move outside the ignores; anything else is NOT-READY with the
// red list. Every successful ratification freezes an IMMUTABLE
// VERSIONED SNAPSHOT (.punchtape/verdicts/vN.txt, a journaled
// effect): v1 is pinned to its digest forever, v2 to its own — the
// model of git commits; the case lives on, the snapshots never move.
package engine

import (
	"strconv"
	"strings"
	"time"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/journal"
	"github.com/neurophant/punchtape/internal/ledger"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// Verdict statuses.
const (
	VerdictReady    = "READY"
	VerdictNotReady = "NOT-READY"
)

// Verdict — the canonical verdict.
type Verdict struct {
	Status string
	Lines  []string
}

// RenderVerdict assembles the verdict from canon and ledger. At
// deliver the answer is the case's LATEST FROZEN SNAPSHOT — the
// accepted state's signed fact (an agent's retelling is not a
// verdict, and a live re-render is not the snapshot); a tree that
// drifted after the acceptance appends the honest drift line without
// touching the snapshot. Mid-work it is the fresh computation.
func (e *Engine) RenderVerdict() (*Verdict, error) {
	state, err := e.Store.State()
	if err != nil {
		return nil, err
	}
	if state.Stage != canon.StageDeliver {
		return e.renderVerdictFresh()
	}
	v, n, ok := e.latestVerdictSnapshot()
	if !ok {
		// A deliver state without a snapshot: rendered fresh each
		// time (the honest fallback; this machine's own ratifications
		// always freeze one).
		return e.renderVerdictFresh()
	}
	e.lastVerdictDrift = false
	if drifted, oldD, newD := e.acceptanceDrift(); drifted {
		drift := verdictFromText(v.Text())
		drift.Status = VerdictNotReady
		drift.Lines[0] = canondata.T("verdict.head", canondata.M{"status": VerdictNotReady})
		line := canondata.T("verdict.acceptance-drift", canondata.M{"old": oldD, "new": newD})
		for i, l := range drift.Lines {
			if strings.HasPrefix(l, "acceptance:") {
				drift.Lines[i] = line
				break
			}
		}
		return drift, nil
	}
	_ = n
	return v, nil
}

// verdictFromText parses a frozen verdict text: the status is the
// first line, the rest as is.
func verdictFromText(text string) *Verdict {
	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
	status := ""
	if len(lines) > 0 {
		status = strings.TrimPrefix(lines[0], "VERDICT: ")
	}
	return &Verdict{Status: status, Lines: lines}
}

// verdictOpts — the inputs of a verdict render: either the live
// store's (a fresh render) or the ratifying transaction's own
// projections (the snapshot freeze).
type verdictOpts struct {
	version     int // 0 — no version line (a fresh render); N — snapshot vN
	prevVerdict int // the version the retired rows are explained against
	retired     []journal.RetiredRow
	submitted   int
	runs        int
	// acceptance — the acceptance line's details; nil — none yet
	acceptance map[string]string
}

// renderVerdictFresh — the live computation over the committed store.
func (e *Engine) renderVerdictFresh() (*Verdict, error) {
	atoms, err := e.Store.Atoms()
	if err != nil {
		return nil, err
	}
	state, err := e.Store.State()
	if err != nil {
		return nil, err
	}
	submitted, runs := 0, 0
	for _, en := range e.Journal.All() {
		switch en.DeltaKind {
		case "run":
			runs++
		case "spec":
			submitted++
		}
	}
	opts := verdictOpts{
		prevVerdict: e.verdictVersion(),
		retired:     e.retiredSinceVerdict(),
		submitted:   submitted,
		runs:        runs,
	}
	if last := lastAcceptanceEntry(e); last != nil {
		opts.acceptance = last.Details
	}
	v := e.buildVerdict(atoms, state, opts)

	// The verdict gate describes the current instance: a tree changed
	// after the acceptance moment makes the verdict's numbers alien —
	// READY on an unreconciled instance would be a lie.
	drifted, oldD, newD := e.acceptanceDrift()
	e.lastVerdictDrift = drifted
	if drifted && v.Status == VerdictReady {
		v.Status = VerdictNotReady
		for i, l := range v.Lines {
			if strings.HasPrefix(l, "acceptance:") {
				v.Lines[i] = canondata.T("verdict.acceptance-drift", canondata.M{
					"old": oldD, "new": newD,
				})
			}
		}
	}
	return v, nil
}

// renderVerdictFor — the ratifying transaction's own snapshot: the
// builder's projections (not yet committed), this run's acceptance
// numbers, and the version line of the snapshot being frozen.
func (e *Engine) renderVerdictFor(b *builder, opts verdictOpts) *Verdict {
	atoms := make([]canon.Atom, 0, len(b.atoms))
	for _, id := range sortedIDs(b.atoms) {
		atoms = append(atoms, b.atoms[id])
	}
	return e.buildVerdict(atoms, b.state, opts)
}

// buildVerdict — the one render: status, fields, digest signature.
func (e *Engine) buildVerdict(atoms []canon.Atom, state canon.State, opts verdictOpts) *Verdict {
	var greenAtoms, redRows, totalRows int
	for _, a := range atoms {
		totalRows += len(a.Rows)
		if a.Green() {
			greenAtoms++
		}
		for _, r := range a.Rows {
			if r.Outcome == canon.RowRed {
				redRows++
			}
		}
	}

	status := VerdictReady
	if len(atoms) == 0 || greenAtoms < len(atoms) || redRows > 0 {
		status = VerdictNotReady
	}

	var wallMs, calls int64
	for _, en := range e.Ledger.All() {
		wallMs += en.WallMs
		if en.Event == "verb" {
			calls++
		}
	}

	v := &Verdict{Status: status}
	accLine := canondata.T("verdict.acceptance-none")
	if d := opts.acceptance; d != nil {
		accLine = canondata.T("verdict.acceptance-pass", canondata.M{
			"digest": d["digest"], "green": d["green"], "total": d["checks"],
		})
	}
	v.Lines = []string{
		canondata.T("verdict.head", canondata.M{"status": status}),
		"", // the digest is inserted after the fields are assembled
	}
	if opts.version > 0 {
		v.Lines = append(v.Lines, canondata.T("verdict.version", canondata.M{
			"version": strconv.Itoa(opts.version),
		}))
	}
	v.Lines = append(v.Lines,
		canondata.T("verdict.atoms", canondata.M{
			"total": strconv.Itoa(len(atoms)), "green": strconv.Itoa(greenAtoms),
			"rows": strconv.Itoa(totalRows),
		}),
		canondata.T("verdict.stages", canondata.M{
			"submitted": strconv.Itoa(opts.submitted), "runs": strconv.Itoa(opts.runs),
		}),
		originLine(state),
		canondata.T("verdict.bill", canondata.M{
			"wall": formatWall(wallMs), "calls": strconv.FormatInt(calls, 10),
		}),
		accLine,
	)
	v.Lines = append(v.Lines, retiredLines(opts.retired, opts.prevVerdict)...)
	if status == VerdictNotReady {
		v.Lines = append(v.Lines, "", canondata.T("verdict.red-list.header", canondata.M{
			"count": strconv.Itoa(redRows),
		}))
		budget := canondata.Limit("slot.list-budget")
		shown := 0
		for _, a := range atoms {
			for i, r := range a.Rows {
				if r.Outcome != canon.RowRed {
					continue
				}
				if shown >= budget {
					v.Lines = append(v.Lines, canondata.T("slot.list.more", canondata.M{
						"count": strconv.Itoa(redRows - shown),
					}))
					shown = -1
					break
				}
				v.Lines = append(v.Lines, canondata.T("verdict.red-list.line", canondata.M{
					"label":  canon.RowLabel(a.ID, i+1),
					"reason": oneLineInfo(r.Reason),
				}))
				shown++
			}
			if shown < 0 {
				break
			}
		}
		if len(atoms) == 0 {
			v.Lines = append(v.Lines, canondata.T("verdict.red-list.no-atoms"))
		}
	}
	// Digest signature: over the normalized text of the fields without
	// the digest line itself.
	body := normalizeVerdict(v.Lines)
	v.Lines[1] = canondata.T("verdict.digest", canondata.M{
		"digest": yamlio.DigestShort([]byte(body)),
	})
	return v
}

// freezeVerdictSnapshot — the ratification's immutable snapshot:
// vN written ONCE inside the ratifying transaction (a journaled
// effect — the journal carries every verdict of the case), the
// version counter riding the same transaction's state. The
// accounting guard refuses a verdict that cannot explain a retired
// row by name.
func (b *builder) freezeVerdictSnapshot() (int, error) {
	prev := b.e.verdictVersion()
	version := prev + 1
	retired := b.e.retiredSinceVerdict()
	// A ratification is always kind run (the form refusal guards it):
	// this transaction is the run the snapshot's counts carry.
	submitted, runs := 0, 0
	for _, en := range b.e.Journal.All() {
		switch en.DeltaKind {
		case "run":
			runs++
		case "spec":
			submitted++
		}
	}
	runs++
	// this run's acceptance numbers (what acceptanceNote is about to
	// record — the snapshot and the ledger entry must agree)
	acc := map[string]string{}
	total := 0
	for _, id := range sortedIDs(b.atoms) {
		total += len(b.atoms[id].Rows)
	}
	if b.lastCkp != nil {
		acc["digest"] = shortHex(b.lastCkp.Digest)
		acc["green"] = strconv.Itoa(total)
		acc["checks"] = strconv.Itoa(total)
	}
	v := b.e.renderVerdictFor(b, verdictOpts{
		version:     version,
		prevVerdict: prev,
		retired:     retired,
		submitted:   submitted,
		runs:        runs,
		acceptance:  acc,
	})
	if err := explainRetired(b.e, v.Text(), retired); err != nil {
		return 0, err
	}
	b.effects = append(b.effects, journal.Effect{
		Path:    verdictSnapshotName(version),
		Content: machineFile([]byte(v.Text())),
	})
	b.state.Counters["VERDICT"] = version
	b.putState()
	return version, nil
}

// lastAcceptanceEntry — the last acceptance moment in the ledger.
func lastAcceptanceEntry(e *Engine) *ledger.Entry {
	var last *ledger.Entry
	for i := range e.Ledger.All() {
		en := &e.Ledger.All()[i]
		if en.Event == "gate" && en.Details["name"] == "suite" &&
			en.Details["scope"] == "acceptance" {
			last = en
		}
	}
	return last
}

// acceptanceDrift — reconciliation of the instance against the
// acceptance moment: the current digest of the surface against the
// acceptance digest. A drift claim needs TWO signatures: an
// acceptance that could not see the surface or a moment where the
// surface is not measurable says nothing about change — a false
// drift after a render is not verdict material.
func (e *Engine) acceptanceDrift() (drifted bool, oldD, newD string) {
	last := lastAcceptanceEntry(e)
	if last == nil {
		return false, "", ""
	}
	oldD = last.Details["digest"]
	// The drift check re-snapshots the CURRENT tree (an observation,
	// not a journal record): the acceptance stands on tree D, the
	// verdict looks at the tree as it is NOW — D' != D is the
	// machine-visible red line.
	ckp, err := e.snapshotTree()
	if err != nil {
		return false, oldD, ""
	}
	newD = shortHex(ckp.Digest)
	if oldD == "" || newD == "" {
		return false, oldD, newD
	}
	return newD != oldD, oldD, newD
}

// originLine — the brief's first line: the origin of the work.
func originLine(state canon.State) string {
	if state.Wish == nil {
		return canondata.T("verdict.origin-none")
	}
	return canondata.T("verdict.origin", canondata.M{
		"text": firstLineOf(*state.Wish),
	})
}

// Text — the canonical verdict text.
func (v *Verdict) Text() string {
	return strings.Join(v.Lines, "\n")
}

// normalizeVerdict brings the verdict fields to a normal form: lines
// without empties, UTF-8, stable order — the digest is reproducible.
// Live meters (the bill) do not enter the signature: the digest is
// the identity of the verdict's state, not of environment counters.
func normalizeVerdict(lines []string) string {
	var sb strings.Builder
	for _, l := range lines {
		if strings.TrimSpace(l) == "" || strings.HasPrefix(l, "digest:") ||
			strings.HasPrefix(l, "bill:") {
			continue
		}
		sb.WriteString(l)
		sb.WriteString("\n")
	}
	return sb.String()
}

func formatWall(ms int64) string {
	d := time.Duration(ms) * time.Millisecond
	return d.Truncate(time.Millisecond).String()
}
