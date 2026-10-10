// The sparse patch core (the owner's economy law): the rule that
// after the first accepted spec the full table
// never rides a submission again — every later document is a patch by
// row address, and the number of submissions stops being an economic
// variable. The patch is EXPANDED into the full document a hand would
// have re-sent: the expansion is plain arithmetic over the corpus (a
// replacement edits its row IN PLACE, adds go to the END of their
// table, removals retire), so the canon bytes of a patch path and of
// the equivalent full re-submission are IDENTICAL by construction —
// there is no second apply law. Row order is insertion-stable
// throughout (point 2): the platform cache's prefix survives every
// patch.
package engine

import (
	"sort"
	"strconv"
	"strings"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/delta"
)

// rowAddr — one row address parsed off its label.
type rowAddr struct {
	atom string
	n    int // one-based
}

// parseRowAddr — "BEHAVIOR-ID/N": the id is the behavior's own (a
// legal id carries no slash), N the one-based row number in its
// canonical spelling (no leading zeros).
func parseRowAddr(s string) (rowAddr, bool) {
	i := strings.LastIndexByte(s, '/')
	if i <= 0 || i == len(s)-1 {
		return rowAddr{}, false
	}
	id, num := s[:i], s[i+1:]
	if !behaviorIDRe.MatchString(id) {
		return rowAddr{}, false
	}
	n, err := strconv.Atoi(num)
	if err != nil || n < 1 || strconv.Itoa(n) != num {
		return rowAddr{}, false
	}
	return rowAddr{atom: id, n: n}, true
}

// atomPlan — one behavior's share of a patch.
type atomPlan struct {
	replace map[int]canon.Row
	remove  map[int]bool
	runMode map[int]delta.PatchRun
	adds    []canon.Row
	// prose/proseSet — the patch's prose replacement for the behavior.
	prose    string
	proseSet bool
	// newName/newProse — the identity of a behavior the patch OPENS
	// (its first add item); ignored for a standing behavior.
	newName, newProse string
}

func newAtomPlan() *atomPlan {
	return &atomPlan{
		replace: map[int]canon.Row{},
		remove:  map[int]bool{},
		runMode: map[int]delta.PatchRun{},
	}
}

// stripMachineRow — the hand-document form of a row: the machine's
// own records (the outcome, the memo) never ride a document; the
// carry-by-digest restores them on the apply side.
func stripMachineRow(r canon.Row) canon.Row {
	r.Outcome, r.Reason = "", ""
	r.EnvFail = false
	r.RanAt, r.RanTree, r.Inputs = "", "", ""
	r.RanRc, r.RanOut = nil, ""
	return r
}

// expandPatchDoc — the patch's full-document equivalent against the
// corpus it arrives at: the document a full re-submission of the same
// change would have been (one law with the door's apply), plus the
// post-expansion labels of the rows the run section forces to
// execute. Every address reads the corpus AS IT STANDS at the patch's
// arrival; a label may ride at most one section. The caller owns the
// corpus load (the door under the writer monopoly, the dry validator
// over plain reads).
func expandPatchDoc(atoms map[string]canon.Atom, d *delta.Doc) (*delta.Doc, map[string]bool, error) {
	p := d.Patch
	ops := len(p.Replace) + len(p.Add) + len(p.Remove) + len(p.Run) + len(p.Prose)
	if ops == 0 {
		return nil, nil, rejection("patch", canondata.T("delta.error.patch-empty"))
	}

	plans := map[string]*atomPlan{}
	plan := func(id string) *atomPlan {
		if pl, ok := plans[id]; ok {
			return pl
		}
		pl := newAtomPlan()
		plans[id] = pl
		return pl
	}

	// one address, one operation: every section's labels are claimed
	// in one fixed-order pre-scan BEFORE anything applies — the
	// conflict branch is one, and reachable for every section pair
	seen := map[string]string{}
	claim := func(label, section string) error {
		if prev, ok := seen[label]; ok && prev != section {
			return rejection("patch", canondata.T("delta.error.patch-conflict", canondata.M{
				"label": label,
			}))
		}
		seen[label] = section
		return nil
	}
	claimSection := func(labels []string, section string) error {
		for _, label := range labels {
			if err := claim(label, section); err != nil {
				return err
			}
		}
		return nil
	}
	// replace is the FIRST claimed section over unique map keys — its
	// claim cannot conflict; the error returns below (remove, run)
	// are the reachable conflict branches
	claimSection(sortedPatchKeys(p.Replace), "replace")
	if err := claimSection(p.Remove, "remove"); err != nil {
		return nil, nil, err
	}
	if err := claimSection(sortedPatchKeys(p.Run), "run"); err != nil {
		return nil, nil, err
	}
	resolve := func(label string) (rowAddr, error) {
		addr, ok := parseRowAddr(label)
		if !ok {
			return addr, rejection("patch", canondata.T("delta.error.patch-label", canondata.M{
				"label": label,
			}))
		}
		atom, ok := atoms[addr.atom]
		if !ok || addr.n > len(atom.Rows) {
			return addr, rejection("patch", canondata.T("delta.error.patch-label-unknown", canondata.M{
				"label": label,
			}))
		}
		return addr, nil
	}

	// replace: full new row bytes at the address. The submitted row
	// is validated AS SUBMITTED first: the machine's own records
	// never ride a document (the strip that feeds the expansion must
	// not launder a forged memo past the door).
	for _, label := range sortedPatchKeys(p.Replace) {
		addr, err := resolve(label)
		if err != nil {
			return nil, nil, err
		}
		if machineFields(p.Replace[label]) {
			return nil, nil, rejection("patch", canondata.T("format.error.machine-fields"))
		}
		plan(addr.atom).replace[addr.n] = p.Replace[label]
	}
	// remove: the retired addresses
	for _, label := range p.Remove {
		addr, err := resolve(label)
		if err != nil {
			return nil, nil, err
		}
		plan(addr.atom).remove[addr.n] = true
	}
	// run: the forced executions, with their optional mode updates
	for _, label := range sortedPatchKeys(p.Run) {
		addr, err := resolve(label)
		if err != nil {
			return nil, nil, err
		}
		mode := p.Run[label]
		switch mode.Launch {
		case "", canon.LaunchModeNonet:
		default:
			return nil, nil, rejection("patch", canondata.T("delta.error.patch-run-mode", canondata.M{
				"label": label,
			}))
		}
		switch mode.Tree {
		case "", canon.TreeModeUnchanged:
		default:
			return nil, nil, rejection("patch", canondata.T("delta.error.patch-run-mode", canondata.M{
				"label": label,
			}))
		}
		plan(addr.atom).runMode[addr.n] = mode
	}
	// add: rows to the END; a new id opens the behavior (one identity
	// per id across the items)
	opened := map[string]bool{}
	openedOrder := []string{}
	for _, item := range p.Add {
		if !behaviorIDRe.MatchString(item.Atom) {
			return nil, nil, rejection("patch", canondata.T("delta.error.patch-label", canondata.M{
				"label": item.Atom,
			}))
		}
		if machineFields(item.Row) {
			return nil, nil, rejection("patch", canondata.T("format.error.machine-fields"))
		}
		if _, stands := atoms[item.Atom]; !stands {
			if opened[item.Atom] {
				return nil, nil, rejection("patch", canondata.T("delta.error.patch-add-atom", canondata.M{
					"id": item.Atom,
				}))
			}
			opened[item.Atom] = true
			openedOrder = append(openedOrder, item.Atom)
			pl := plan(item.Atom)
			pl.newName, pl.newProse = item.Name, item.Prose
		}
		plan(item.Atom).adds = append(plan(item.Atom).adds, item.Row)
	}
	// prose: the verdict-gate texts, once per behavior; only a
	// standing behavior's prose is replaceable (a new behavior's prose
	// rides its opening add item)
	for _, id := range sortedPatchKeys(p.Prose) {
		if _, stands := atoms[id]; !stands {
			return nil, nil, rejection("patch", canondata.T("delta.error.patch-atom-unknown", canondata.M{
				"id": id,
			}))
		}
		pl := plan(id)
		pl.prose, pl.proseSet = p.Prose[id], true
	}

	// compose: every planned behavior's new row list — replacements
	// in place, removals out, adds at the end, mode updates on their
	// (possibly shifted) rows; the forced labels ride the POST table
	force := map[string]bool{}
	newAtoms := map[string]canon.Atom{}
	for id, pl := range plans {
		old := atoms[id]
		var rows []canon.Row
		postOf := map[int]int{}
		for i, row := range old.Rows {
			if pl.remove[i+1] {
				continue
			}
			nr := row
			if repl, ok := pl.replace[i+1]; ok {
				nr = repl
			}
			postOf[i+1] = len(rows) + 1
			rows = append(rows, stripMachineRow(nr))
		}
		rows = append(rows, pl.adds...)
		// a run label's pre-address survived removals by the conflict
		// pre-scan (one address, one operation) — its post-address
		// exists by construction
		for pre, mode := range pl.runMode {
			post := postOf[pre]
			rows[post-1].Launch = mode.Launch
			rows[post-1].Tree = mode.Tree
			force[canon.RowLabel(id, post)] = true
		}
		atom := canon.Atom{ID: id, Name: old.Name, Prose: old.Prose, Rows: rows}
		if _, stands := atoms[id]; !stands {
			atom.Name, atom.Prose = pl.newName, pl.newProse
		}
		if pl.proseSet {
			atom.Prose = pl.prose
		}
		newAtoms[id] = atom
	}

	// the expansion: kind spec IS the full state (every atom, machine
	// records stripped); kind run carries the TOUCHED atoms only (the
	// upsert law — a run never retires a behavior it does not name)
	out := &delta.Doc{
		Kind:        d.Kind,
		Build:       d.Build,
		BuildIgnore: d.BuildIgnore,
		Ratify:      d.Ratify,
	}
	if d.Kind == delta.KindSpec {
		for _, id := range sortedIDs(atoms) {
			if na, planned := newAtoms[id]; planned {
				if len(na.Rows) == 0 {
					continue // emptied by removals: the atom retires
				}
				out.Behaviors = append(out.Behaviors, delta.Behavior{
					ID: na.ID, Name: na.Name, Prose: na.Prose, Rows: na.Rows,
				})
				continue
			}
			a := atoms[id]
			rows := make([]canon.Row, 0, len(a.Rows))
			for _, row := range a.Rows {
				rows = append(rows, stripMachineRow(row))
			}
			out.Behaviors = append(out.Behaviors, delta.Behavior{
				ID: a.ID, Name: a.Name, Prose: a.Prose, Rows: rows,
			})
		}
		// the atoms the patch OPENS ride after the standing corpus —
		// the END of the table, the same insertion-stable law as new
		// rows (point 2); an opened atom always carries its opening
		// add's rows, and the spec expansion is the full state: an
		// atom missing here would VANISH from the corpus, not land
		for _, id := range openedOrder {
			na := newAtoms[id]
			out.Behaviors = append(out.Behaviors, delta.Behavior{
				ID: na.ID, Name: na.Name, Prose: na.Prose, Rows: na.Rows,
			})
		}
		return out, force, nil
	}
	for _, id := range sortedIDs(newAtoms) {
		na := newAtoms[id]
		if len(na.Rows) == 0 {
			// a run patch cannot empty a behavior: the expansion
			// carries it and the door's own form rule names it
			out.Behaviors = append(out.Behaviors, delta.Behavior{
				ID: na.ID, Name: na.Name, Prose: na.Prose,
			})
			continue
		}
		out.Behaviors = append(out.Behaviors, delta.Behavior{
			ID: na.ID, Name: na.Name, Prose: na.Prose, Rows: na.Rows,
		})
	}
	return out, force, nil
}

// machineFields — whether a SUBMITTED row tries to write the
// machine's own records (the outcome, the memo): the same law the
// door's form gate holds over full documents, held over the patch's
// own rows before anything strips or composes.
func machineFields(r canon.Row) bool {
	return r.Outcome != "" || r.Reason != "" || r.EnvFail || r.Memoized()
}

// sortedPatchKeys — a patch map's labels in a deterministic order
// (the findings and the reply never depend on map iteration).
func sortedPatchKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
