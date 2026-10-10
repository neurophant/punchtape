// The spec-as-state core: the living case's version
// machinery. A spec document IS the full corpus state — the machine
// computes the diff against the current corpus (new/changed/retired
// atoms and rows), records retired rows as visible journal frames,
// carries row results across versions by content (the memoization
// key: a row keeps its green while its definition and the digests of
// the files its command names — plus its from-input seeds, its
// sample and the declared build — are unchanged), and freezes each
// ratification as an immutable versioned verdict snapshot. The diff
// is a consequence of two stored versions, never a second language.
package engine

import (
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/checks"
	"github.com/neurophant/punchtape/internal/delta"
	"github.com/neurophant/punchtape/internal/fsport"
	"github.com/neurophant/punchtape/internal/journal"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// specDiff — the machine-computed diff of one spec submission against
// the corpus it arrives at: counts plus the named lines (labels) the
// reply prints and the retired rows the journal frame records.
type specDiff struct {
	atomsTotal   int
	newAtoms     int
	changedAtoms int
	retiredAtoms int
	rowsTotal    int
	newRows      int
	changedRows  []string
	retiredRows  []journal.RetiredRow
}

// summaryRender — the reply's one counts line for the thin ack (the
// named diff lines ride the next slot, not the ack).
func (d *specDiff) summaryRender(version int) string {
	return canondata.T("submit.spec-diff", canondata.M{
		"version": strconv.Itoa(version),
		"atoms":   strconv.Itoa(d.atomsTotal),
		"na":      strconv.Itoa(d.newAtoms),
		"ca":      strconv.Itoa(d.changedAtoms),
		"ra":      strconv.Itoa(d.retiredAtoms),
		"rows":    strconv.Itoa(d.rowsTotal),
		"nr":      strconv.Itoa(d.newRows),
		"cr":      strconv.Itoa(len(d.changedRows)),
		"rr":      strconv.Itoa(len(d.retiredRows)),
	})
}

// summaryLine — the compact one-line form for the ledger.
func (d *specDiff) summaryLine() string {
	return fmt.Sprintf("+%d/~%d/-%d atoms; +%d/~%d/-%d rows",
		d.newAtoms, d.changedAtoms, d.retiredAtoms,
		d.newRows, len(d.changedRows), len(d.retiredRows))
}

// corpusIDsEqual — whether the corpus's atom ids equal the given
// sorted ids: the spec replay gate (the arrival corpus is exactly
// what the document describes — applying it again would change
// nothing).
func (e *Engine) corpusIDsEqual(sortedIDs []string) bool {
	atoms, err := e.Store.Atoms()
	if err != nil {
		return false
	}
	if len(atoms) != len(sortedIDs) {
		return false
	}
	for i, a := range atoms {
		if a.ID != sortedIDs[i] {
			return false
		}
	}
	return true
}

// carryOutcome — the content-addressed carry-over: the incoming row
// keeps the recorded outcome and memo of a same-definition row of
// the same atom (a re-declared unchanged row is NOT new work; its
// green stands at its recorded digests). Returns the row with the
// memo carried, and whether a match was found.
func carryOutcome(incoming canon.Row, old []canon.Row, used map[int]bool) (canon.Row, bool) {
	want := incoming.RowDefDigest()
	for i, prev := range old {
		if used[i] || prev.RowDefDigest() != want {
			continue
		}
		used[i] = true
		incoming.Outcome, incoming.Reason = prev.Outcome, prev.Reason
		incoming.EnvFail = prev.EnvFail
		incoming.RanAt, incoming.RanTree, incoming.Inputs = prev.RanAt, prev.RanTree, prev.Inputs
		incoming.RanRc, incoming.RanOut = prev.RanRc, prev.RanOut
		return incoming, true
	}
	return incoming, false
}

// applySpecState — the spec kind's state semantics: the document's
// behaviors REPLACE the corpus. Absent ids retire their atoms (a
// delete effect plus diff rows); kept atoms replace their rows with
// the document's, carrying outcomes by definition; the diff and the
// retired frame are computed here. Run-kind callers pass upsert=true
// (repairs never retire).
func (b *builder) applySpecState(d *delta.Doc) error {
	fullState := d.Kind == delta.KindSpec
	// The retire frames keep their own moment: whether the case had
	// runs when THIS retirement happened (a later run does not
	// retroactively criminalize a pre-run edit of the first table).
	noRunsYet := !b.e.runsHappened()
	corpus := map[string]canon.Atom{}
	for id, a := range b.atoms {
		corpus[id] = a
	}
	diff := &specDiff{}
	seen := map[string]bool{}
	for _, beh := range d.Behaviors {
		atom := canon.Atom{ID: beh.ID, Name: beh.Name, Prose: beh.Prose}
		old, kept := corpus[beh.ID]
		used := map[int]bool{}
		for ri, row := range beh.Rows {
			label := canon.RowLabel(beh.ID, ri+1)
			if kept {
				carried, matched := carryOutcome(row, old.Rows, used)
				row = carried
				switch {
				case matched:
					// the definition is unchanged: its outcome stands
				case ri < len(old.Rows):
					diff.changedRows = append(diff.changedRows, label)
				default:
					diff.newRows++
				}
			} else {
				diff.newRows++
			}
			atom.Rows = append(atom.Rows, row)
		}
		if kept {
			// rows of the old atom no incoming definition matches: retired
			for i, prev := range old.Rows {
				if used[i] {
					continue
				}
				diff.retiredRows = append(diff.retiredRows, journal.RetiredRow{
					Label:          canon.RowLabel(beh.ID, i+1),
					Outcome:        prev.Outcome,
					Reason:         firstLineOf(prev.Reason),
					BeforeFirstRun: noRunsYet,
				})
			}
			if old.Name != beh.Name || old.Prose != beh.Prose ||
				rowDefsDiffer(old.Rows, atom.Rows) {
				diff.changedAtoms++
			}
		} else {
			diff.newAtoms++
		}
		seen[beh.ID] = true
		if err := b.putAtom(atom); err != nil {
			return err
		}
	}
	if fullState {
		// the document is the whole corpus: ids it does not name retire
		for _, id := range sortedIDs(corpus) {
			if seen[id] {
				continue
			}
			diff.retiredAtoms++
			for i, prev := range corpus[id].Rows {
				diff.retiredRows = append(diff.retiredRows, journal.RetiredRow{
					Label:          canon.RowLabel(id, i+1),
					Outcome:        prev.Outcome,
					Reason:         firstLineOf(prev.Reason),
					BeforeFirstRun: noRunsYet,
				})
			}
			b.retireAtom(id)
		}
	}
	diff.atomsTotal = len(seen)
	for _, beh := range d.Behaviors {
		diff.rowsTotal += len(beh.Rows)
	}
	b.specDiffOut = diff
	return nil
}

// rowDefsDiffer — whether two row lists differ by definition (order
// included; the carry already matched content).
func rowDefsDiffer(a, b []canon.Row) bool {
	if len(a) != len(b) {
		return true
	}
	for i := range a {
		if a[i].RowDefDigest() != b[i].RowDefDigest() {
			return true
		}
	}
	return false
}

// retireAtom — the delete effect for a retired atom's canon file.
func (b *builder) retireAtom(id string) {
	delete(b.atoms, id)
	b.effects = append(b.effects, journal.Effect{
		Path:   canonFile("atoms", id),
		Delete: true,
	})
}

// --- the content-addressed memo ---

// rowInputTokens — the row's freshness key at this moment: the
// digests of the files its command names in the work tree (the
// launch surface — a sh -c wrapper names nothing it does not spell),
// its from-input seeds' sources, its byte sample, and the declared
// build. Sorted, space-joined tokens; empty inputs still carry the
// build token, so a build declaration change staleness every row.
func (b *builder) rowInputTokens(row canon.Row) string {
	var tokens []string
	if fields, err := row.CommandFields(); err == nil {
		for _, f := range fields {
			rel := strings.TrimPrefix(f, "./")
			if rel == "" {
				continue
			}
			data, err := fsport.Bind.ReadFile(filepath.Join(b.e.Workdir, filepath.FromSlash(rel)))
			if err != nil {
				continue // not a tree file: a PATH tool or a deliberate missing path
			}
			tokens = append(tokens, rel+":"+shortHex(yamlio.Digest(data)))
		}
	}
	for _, s := range row.Seeds {
		if s.FromInput == "" {
			continue
		}
		if data, err := fsport.Bind.ReadFile(filepath.Join(b.e.Workdir, "..", "input",
			filepath.FromSlash(s.FromInput))); err == nil {
			tokens = append(tokens, "input/"+s.FromInput+":"+shortHex(yamlio.Digest(data)))
		} else {
			tokens = append(tokens, "input/"+s.FromInput+":gone")
		}
	}
	if ref := row.BytesRef(); ref != "" {
		if data, err := fsport.Bind.ReadFile(filepath.Join(b.e.Workdir, ".punchtape", "inbox",
			filepath.FromSlash(ref))); err == nil {
			tokens = append(tokens, "inbox/"+ref+":"+shortHex(yamlio.Digest(data)))
		} else {
			tokens = append(tokens, "inbox/"+ref+":gone")
		}
	}
	if ref := row.ProbeRef(); ref != "" {
		if data, err := fsport.Bind.ReadFile(filepath.Join(b.e.Workdir, ".punchtape", "inbox",
			filepath.FromSlash(ref))); err == nil {
			tokens = append(tokens, "probe/"+ref+":"+shortHex(yamlio.Digest(data)))
		} else {
			tokens = append(tokens, "probe/"+ref+":gone")
		}
	}
	tokens = append(tokens, "build:"+shortHex(yamlio.Digest([]byte(strings.Join(b.state.Build, "\x00")))))
	sort.Strings(tokens)
	return strings.Join(tokens, " ")
}

// rowFresh — the memoization predicate: the row's recorded result
// stands (nothing to re-run) exactly when it is green, carries a
// memo, and its input tokens are unchanged. A red or memoless row
// always re-runs; a touched input staleness its row.
func (b *builder) rowFresh(row canon.Row) bool {
	return row.Outcome == canon.RowGreen && row.Memoized() &&
		row.Inputs == b.rowInputTokens(row)
}

// recordMemo — the executed row's freshness key: the outcome now
// stands at exactly these inputs, on this run's tree, with the
// observed exit code and output digest as the measured answer.
func (b *builder) recordMemo(row *canon.Row, ckp *Checkpoint, res checks.Result) {
	row.RanAt = timeNow().UTC().Format("2006-01-02T15:04:05Z")
	row.Inputs = b.rowInputTokens(*row)
	if ckp != nil {
		row.RanTree = shortHex(ckp.Digest)
	}
	row.RanRc, row.RanOut = res.Exit, res.OutDigest
}

// --- the verdict version machinery ---

// verdictSnapshotName — the snapshot file of verdict vN.
func verdictSnapshotName(n int) string {
	return fmt.Sprintf(".punchtape/verdicts/v%d.txt", n)
}

// verdictVersion — the current verdict version (0 — no verdict yet).
func (e *Engine) verdictVersion() int {
	state, err := e.Store.State()
	if err != nil {
		return 0
	}
	return state.Counters["VERDICT"]
}

// latestVerdictSnapshot — the newest frozen verdict snapshot (the
// living case's current accepted state); false when none exists.
func (e *Engine) latestVerdictSnapshot() (*Verdict, int, bool) {
	n := e.verdictVersion()
	if n == 0 {
		return nil, 0, false
	}
	data, err := fsport.Bind.ReadFile(filepath.Join(e.Workdir,
		filepath.FromSlash(verdictSnapshotName(n))))
	if err != nil {
		return nil, n, false
	}
	return verdictFromText(string(data)), n, true
}

// retiredSinceVerdict — the retired-row frames recorded after the
// last frozen verdict (for v1: after the case opened), one entry per
// label (the LAST frame wins: a row re-added and re-retired explains
// its final retirement). The next verdict explains every one of them.
func (e *Engine) retiredSinceVerdict() []journal.RetiredRow {
	last := e.verdictVersion()
	byLabel := map[string]journal.RetiredRow{}
	var order []string
	reset := false
	for _, en := range e.Journal.All() {
		if last > 0 && en.VerdictVersion == last {
			reset = true // frames before (and including) the last verdict's own transaction
			continue
		}
		if reset {
			byLabel, order, reset = map[string]journal.RetiredRow{}, nil, false
		}
		for _, r := range en.Retired {
			if _, seen := byLabel[r.Label]; !seen {
				order = append(order, r.Label)
			}
			byLabel[r.Label] = r
		}
	}
	if reset { // the boundary was the last entry: nothing after it
		return nil
	}
	out := make([]journal.RetiredRow, 0, len(order))
	for _, label := range order {
		out = append(out, byLabel[label])
	}
	return out
}

// retiredLines — the verdict's explanation section for the retired
// rows: every disappearance named with its last recorded standing.
func retiredLines(rows []journal.RetiredRow, prev int) []string {
	if len(rows) == 0 {
		return nil
	}
	head := canondata.T("verdict.retired.header", canondata.M{
		"prev": strconv.Itoa(prev), "count": strconv.Itoa(len(rows)),
	})
	if prev == 0 {
		head = canondata.T("verdict.retired.header.first", canondata.M{
			"count": strconv.Itoa(len(rows)),
		})
	}
	out := []string{"", head}
	for _, r := range rows {
		if r.Outcome == "" {
			if r.BeforeFirstRun {
				out = append(out, canondata.T("verdict.retired.before-first-run", canondata.M{
					"label": r.Label,
				}))
				continue
			}
			out = append(out, canondata.T("verdict.retired.never-ran", canondata.M{
				"label": r.Label,
			}))
			continue
		}
		if r.Reason != "" {
			out = append(out, canondata.T("verdict.retired.line", canondata.M{
				"label": r.Label, "outcome": r.Outcome, "reason": r.Reason,
			}))
			continue
		}
		out = append(out, canondata.T("verdict.retired.line.bare", canondata.M{
			"label": r.Label, "outcome": r.Outcome,
		}))
	}
	return out
}

// explainRetired — the ratify accounting guard, two honest refusals:
// (1) the verdict about to be frozen must account for every retired
// row by name — a verdict that cannot explain a departure never
// ships; (2) a row retired while the case had runs, without ever
// running, cannot be explained at all — deleting a check before it
// measured anything is hiding it, and the ratify names the rows.
// The amnesty is judged by the retire frame's own moment
// (BeforeFirstRun): a row dropped before the first run never had a
// chance to measure — that is an edit of the hand's own first
// table, not a hiding, and a later run does not retroactively
// criminalize it.
func explainRetired(e *Engine, text string, rows []journal.RetiredRow) error {
	var missing, neverRan []string
	for _, r := range rows {
		if !strings.Contains(text, r.Label) {
			missing = append(missing, r.Label)
		}
		if !r.BeforeFirstRun && r.Outcome == "" {
			neverRan = append(neverRan, r.Label)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return rejection("verdict", canondata.T("verdict.retired.guard", canondata.M{
			"labels": strings.Join(missing, ", "),
		}))
	}
	if len(neverRan) > 0 {
		sort.Strings(neverRan)
		return rejection("verdict", canondata.T("verdict.retired.never-ran-guard", canondata.M{
			"labels": strings.Join(neverRan, ", "),
		}))
	}
	return nil
}
