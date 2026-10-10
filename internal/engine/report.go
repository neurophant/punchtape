// The report verb: the operator's table of the
// observable behavior — every row's exact command, its demand, the
// measured answer (the exit code and output digest of the last run),
// the run's time and the freshness mark. A READ-ONLY render of the
// stored canon runs: no row, no build, no ledger entry, no baseline
// write —
// the machine lifts the uncertainty table, the operator's own eyes
// keep their duty. --since vN aims the same facts at the delta after
// a frozen verdict: the changed, the new, the re-measured and the
// retired rows by name (the corpus at vN is folded from the journal
// — the two stored versions are the diff's ground, never a second
// language).
package engine

import (
	"sort"
	"strconv"
	"strings"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/journal"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// Report renders the report. The DEFAULT surface is the DOOR view
// (the door law): per-atom counters and the red rows in full — the
// green rows are a counter, the full table of every row is the
// canon's. full=true prints the whole table (the driver's and the
// judge's surface). since is "" (the whole table) or a verdict
// spelling ("v2"/"2"/"v0" — v0 names the case's open, the empty
// corpus); the since-walk is delta-shaped in both surfaces. A render
// is never a submission: rc 0 with the text, Rejected for a verdict
// the case does not carry.
func (e *Engine) Report(since string, full bool) (string, int) {
	state, err := e.Store.State()
	if err != nil {
		return canondata.T("cli.fail", canondata.M{"err": firstLine(err)}), Failure
	}
	if since == "" {
		var text string
		if full {
			text, err = e.reportFullTable(state)
		} else {
			text, err = e.reportDoor(state)
		}
		if err != nil {
			return canondata.T("cli.fail", canondata.M{"err": firstLine(err)}), Failure
		}
		return text, Accepted
	}
	n, bad := parseVerdictSpelling(since)
	if bad {
		return canondata.T("report.error.no-verdict", canondata.M{
			"verdict": since, "latest": "v" + strconv.Itoa(e.verdictVersion()),
		}), Rejected
	}
	if latest := e.verdictVersion(); n > latest {
		return canondata.T("report.error.no-verdict", canondata.M{
			"verdict": "v" + strconv.Itoa(n), "latest": "v" + strconv.Itoa(latest),
		}), Rejected
	}
	text, err := e.reportSince(state, n)
	if err != nil {
		return canondata.T("cli.fail", canondata.M{"err": firstLine(err)}), Failure
	}
	return text, Accepted
}

// parseVerdictSpelling — "vN" or "N", N >= 0; anything else is not a
// verdict spelling.
func parseVerdictSpelling(s string) (int, bool) {
	s = strings.TrimPrefix(s, "v")
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 || strconv.Itoa(n) != s {
		return 0, true
	}
	return n, false
}

// reportLine — one row's table line: the label with its standing,
// the exact command, the demand, the measured answer, the run's time.
func reportLine(label string, row canon.Row, fresh bool) string {
	got := canondata.T("report.got.never")
	switch {
	case row.Outcome == canon.RowRed && row.Reason != "":
		got = canondata.T("report.got.red", canondata.M{"reason": firstLineOf(row.Reason)})
	case row.Memoized():
		rc := "?"
		if row.RanRc != nil {
			rc = strconv.Itoa(*row.RanRc)
		}
		got = canondata.T("report.got.fresh", canondata.M{"rc": rc, "out": row.RanOut})
	}
	at := row.RanAt
	if at == "" {
		at = "-"
	}
	return canondata.T("report.row", canondata.M{
		"label": label, "standing": reportStanding(row, fresh),
		"run": oneLineValue(row.CmdForm()), "want": oneLineValue(expectationLine(row)),
		"got": got, "at": at,
	})
}

// reportStanding — the row's freshness word: the standings walk and
// the report speak one vocabulary.
func reportStanding(row canon.Row, fresh bool) string {
	switch {
	case row.Outcome == canon.RowRed:
		return "red"
	case row.Outcome != canon.RowGreen:
		return "never"
	case fresh:
		return "green, fresh"
	default:
		return "green, stale"
	}
}

// reportDoor — the door view: the atoms with their counters, every
// red row in full (its measured reason is the work), the greens one
// counter. The hand pays for what is red and what moved; the full
// table is one command away.
func (e *Engine) reportDoor(state canon.State) (string, error) {
	var sb strings.Builder
	writeLine(&sb, canondata.T("report.head.door"))
	writeLine(&sb, e.reportStateLine(state))
	b := newBuilder(e, state)
	if err := b.load(); err != nil {
		return "", err
	}
	if len(b.atoms) == 0 {
		writeLine(&sb, canondata.T("report.head.no-rows"))
		return sb.String(), nil
	}
	greenTotal := 0
	type redLine struct {
		label string
		row   canon.Row
	}
	var reds []redLine
	for _, id := range sortedIDs(b.atoms) {
		a := b.atoms[id]
		green, red, never := 0, 0, 0
		for i, row := range a.Rows {
			switch row.Outcome {
			case canon.RowGreen:
				green++
			case canon.RowRed:
				red++
				reds = append(reds, redLine{canon.RowLabel(id, i+1), row})
			default:
				never++
			}
		}
		greenTotal += green
		writeLine(&sb, canondata.T("report.door.atom", canondata.M{
			"id": a.ID, "name": a.Name, "rows": strconv.Itoa(len(a.Rows)),
			"green": strconv.Itoa(green), "red": strconv.Itoa(red), "never": strconv.Itoa(never),
		}))
	}
	if len(reds) > 0 {
		writeLine(&sb, canondata.T("report.door.red.header", canondata.M{
			"count": strconv.Itoa(len(reds)),
		}))
		budget := canondata.Limit("slot.list-budget")
		for i, r := range reds {
			if i >= budget {
				writeLine(&sb, canondata.T("slot.list.more", canondata.M{
					"count": strconv.Itoa(len(reds) - budget),
				}))
				break
			}
			sb.WriteString("  ")
			writeLine(&sb, reportLine(r.label, r.row, b.rowFresh(r.row)))
		}
	}
	if greenTotal > 0 {
		writeLine(&sb, canondata.T("report.door.green.counter", canondata.M{
			"count": strconv.Itoa(greenTotal),
		}))
	}
	return sb.String(), nil
}

// reportFullTable — the whole table, EVERY row (the --full surface:
// the driver's and the judge's table — the budget law belongs to the
// door, not to the full canon render).
func (e *Engine) reportFullTable(state canon.State) (string, error) {
	var sb strings.Builder
	writeLine(&sb, canondata.T("report.head"))
	writeLine(&sb, e.reportStateLine(state))
	b := newBuilder(e, state)
	if err := b.load(); err != nil {
		return "", err
	}
	labels, rows := corpusRows(b)
	if len(labels) == 0 {
		writeLine(&sb, canondata.T("report.head.no-rows"))
		return sb.String(), nil
	}
	writeLine(&sb, canondata.T("report.rows.header", canondata.M{
		"count": strconv.Itoa(len(labels)),
	}))
	for _, label := range labels {
		writeLine(&sb, reportLine(label, rows[label], b.rowFresh(rows[label])))
	}
	return sb.String(), nil
}

// reportStateLine — the one orientation line under the head: the
// stage, the frozen verdict, the acceptance moment the green stands
// on.
func (e *Engine) reportStateLine(state canon.State) string {
	last := lastAcceptanceEntry(e)
	if e.verdictVersion() == 0 || last == nil {
		return canondata.T("report.head.state.no-verdict", canondata.M{
			"stage": state.Stage,
		})
	}
	return canondata.T("report.head.state", canondata.M{
		"stage": state.Stage, "verdict": "v" + strconv.Itoa(e.verdictVersion()),
		"digest": last.Details["digest"],
		"green":  last.Details["green"], "total": last.Details["checks"],
	})
}

// corpusRows — the live corpus's rows in TABLE order: atoms by id,
// rows inside an atom in their stored (insertion-stable) order — a
// replacement edits its row in place, an added row lands at the end,
// so the render of the next version is a prefix extension of the
// previous one (the platform cache reads the
// same bytes again). Lexicographic label order is NOT table order
// (B1/10 would ride before B1/2) and never renders again.
func corpusRows(b *builder) ([]string, map[string]canon.Row) {
	var labels []string
	rows := map[string]canon.Row{}
	for _, id := range sortedIDs(b.atoms) {
		for i, row := range b.atoms[id].Rows {
			label := canon.RowLabel(id, i+1)
			labels = append(labels, label)
			rows[label] = row
		}
	}
	return labels, rows
}

// --- the --since delta ---

// rowKey — one row's identity for the since-walk: the definition
// digest (what the demand asked) and the memo digest (what the
// last run measured).
type rowKey struct {
	def  string
	memo string
}

// memoDigest — the memo's own digest (the ran facts, not the
// definition): a moved memo means the row re-measured.
func memoDigest(row canon.Row) string {
	data, _ := yamlio.Marshal(canon.Row{
		Outcome: row.Outcome, Reason: row.Reason, EnvFail: row.EnvFail,
		RanAt: row.RanAt, RanTree: row.RanTree, Inputs: row.Inputs,
		RanRc: row.RanRc, RanOut: row.RanOut,
	})
	return yamlio.DigestShort(data)
}

// corpusAtVerdict — the corpus the frozen verdict vN signed, folded
// from the journal's atom-file effects up to and including vN's own
// ratifying transaction (v0 — the case's open, the empty corpus).
// The caller checked vN exists; a boundary no entry names folds the
// empty corpus.
func (e *Engine) corpusAtVerdict(n int) map[string]rowKey {
	at := map[string]rowKey{}
	boundary := -1
	if n > 0 {
		for i, en := range e.Journal.All() {
			if en.VerdictVersion == n {
				boundary = i
				break
			}
		}
	}

	files := map[string]string{}
	var order []string
	for _, en := range e.Journal.All()[:boundary+1] {
		for _, ef := range en.Effects {
			p := strings.TrimPrefix(ef.Path, ".punchtape/canon/atoms/")
			if p == ef.Path {
				continue
			}
			if ef.Delete {
				delete(files, p)
				continue
			}
			if _, seen := files[p]; !seen {
				order = append(order, p)
			}
			files[p] = ef.Content
		}
	}
	sort.Strings(order)
	for _, name := range order {
		var a canon.Atom
		if yamlio.DecodeDoc([]byte(files[name]), &a) != nil {
			continue
		}
		for i, row := range a.Rows {
			at[canon.RowLabel(a.ID, i+1)] = rowKey{
				def: row.RowDefDigest(), memo: memoDigest(row),
			}
		}
	}
	return at
}

// retiredAfter — the retired-row frames recorded after the verdict
// vN's own transaction, one entry per label (the LAST frame wins).
func (e *Engine) retiredAfter(n int) []journal.RetiredRow {
	byLabel := map[string]journal.RetiredRow{}
	var order []string
	start := 0
	for i, en := range e.Journal.All() {
		if n > 0 && en.VerdictVersion == n {
			start = i + 1
			break
		}
	}
	for _, en := range e.Journal.All()[start:] {
		for _, r := range en.Retired {
			if _, seen := byLabel[r.Label]; !seen {
				order = append(order, r.Label)
			}
			byLabel[r.Label] = r
		}
	}
	out := make([]journal.RetiredRow, 0, len(order))
	for _, label := range order {
		out = append(out, byLabel[label])
	}
	return out
}

// reportSince — the delta after the frozen verdict vN: the changed,
// new, re-measured and retired rows, each named with its current
// standing. The untouched rows stay silent — the operator verifies
// the delta, the acceptance already signed the rest.
func (e *Engine) reportSince(state canon.State, n int) (string, error) {
	at := e.corpusAtVerdict(n)
	b := newBuilder(e, state)
	if err := b.load(); err != nil {
		return "", err
	}
	labels, rows := corpusRows(b)
	v := "v" + strconv.Itoa(n)
	var changed, newRows, measured []string
	now := map[string]bool{}
	for _, label := range labels {
		row := rows[label]
		now[label] = true
		key, had := at[label]
		switch {
		case !had:
			newRows = append(newRows, label)
		case key.def != row.RowDefDigest():
			changed = append(changed, label)
		case key.memo != memoDigest(row):
			measured = append(measured, label)
		}
	}
	var retiredLabels []string
	retired := map[string]journal.RetiredRow{}
	for _, r := range e.retiredAfter(n) {
		if now[r.Label] {
			continue // re-added later: the current corpus names it above
		}
		if _, seen := retired[r.Label]; !seen {
			retiredLabels = append(retiredLabels, r.Label)
		}
		retired[r.Label] = r
	}
	if len(changed)+len(newRows)+len(measured)+len(retiredLabels) == 0 {
		return canondata.T("report.since.none", canondata.M{"verdict": v}) + "\n", nil
	}
	var sb strings.Builder
	writeLine(&sb, canondata.T("report.since.head", canondata.M{"verdict": v}))
	section := func(head string, names []string, line func(string) string) {
		if len(names) == 0 {
			return
		}
		writeLine(&sb, canondata.T(head, canondata.M{
			"verdict": v, "count": strconv.Itoa(len(names)),
		}))
		budget := canondata.Limit("slot.list-budget")
		for i, label := range names {
			if i >= budget {
				writeLine(&sb, canondata.T("slot.list.more", canondata.M{
					"count": strconv.Itoa(len(names) - budget),
				}))
				break
			}
			sb.WriteString("  ")
			writeLine(&sb, line(label))
		}
	}
	section("report.since.changed", changed, func(l string) string {
		return reportLine(l, rows[l], b.rowFresh(rows[l]))
	})
	section("report.since.new", newRows, func(l string) string {
		return reportLine(l, rows[l], b.rowFresh(rows[l]))
	})
	section("report.since.measured", measured, func(l string) string {
		return reportLine(l, rows[l], b.rowFresh(rows[l]))
	})
	section("report.since.retired", retiredLabels, func(l string) string {
		outcome := retired[l].Outcome
		if outcome == "" {
			outcome = canon.RowUnknown
		}
		return canondata.T("report.since.retired.line", canondata.M{
			"label": l, "outcome": outcome,
		})
	})
	return sb.String(), nil
}
