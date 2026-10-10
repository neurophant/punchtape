// Row execution: a run submission executes ALL rows of ALL behaviors
// inside its own transaction — the build (when the spec declares
// one) first, then the rows in the live work tree, the tree identity
// control around them, the checkpoint and the outcomes in one
// journal entry. The agent knows the outcome as a table in the
// reply; executing a row by the agent itself is forbidden.
package engine

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/checks"
	"github.com/neurophant/punchtape/internal/execport"
	"github.com/neurophant/punchtape/internal/fsport"
	"github.com/neurophant/punchtape/internal/ledger"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// runRows — the run transaction's execution stage: the pre-run tree
// snapshot, the build, then every STALE row of every atom (in atom
// id order, the atom's row order inside) — a row whose green stands
// at unchanged inputs does not re-run (the content-addressed memo);
// the post-run tree identity control, the checkpoint. Outcomes land
// in the atom records of the same transaction. The reply table is
// accumulated in b.runTable. A failing substrate is executed into
// env reds — the run cannot fail.
func (b *builder) runRows() {
	e := b.e
	// The run's snapshots carry THIS submission's build-ignore
	// declaration (the builder's state, not the committed one): the
	// ignore that arrives with the repair covers the repair's own
	// tree identity control.
	ckp, subErr := e.snapshotTreeIgnores(b.state.BuildIgnore)
	b.lastCkp = ckp
	notRun := ""
	if subErr != nil {
		notRun = subErr.Error()
	} else if berr := e.runBuild(); berr != nil {
		notRun = canondata.T("gates.not-run-cause-build", canondata.M{
			"output": berr.Error(),
		})
		b.buildFailed = notRun
	}
	env := checks.RunEnv{
		InputDir: inputDirOf(e.Workdir),
		InboxDir: filepath.Join(e.Workdir, ".punchtape", "inbox"),
	}
	var redLines []string
	for _, id := range sortedIDs(b.atoms) {
		atom := b.atoms[id]
		for i := range atom.Rows {
			label := canon.RowLabel(id, i+1)
			// The memoization gate: a green result with an unchanged
			// definition and unchanged input digests stands — the row
			// does not re-run (a failed build cannot borrow it: only the
			// rows that would run pay for the build). A row the patch's
			// run section names executes WHATEVER its freshness: the
			// explicit request is a measurement, not a repair.
			if notRun == "" && !b.forceRun[label] && b.rowFresh(atom.Rows[i]) {
				b.keptRows++
				continue
			}
			if notRun != "" {
				// The run did not happen for this row: the "not run"
				// mark is a run-level refusal (env), not behavior.
				reason := canondata.T("gates.suite-not-run", canondata.M{
					"cause": notRun,
				})
				atom.Rows[i].Outcome, atom.Rows[i].Reason = canon.RowRed, reason
				atom.Rows[i].EnvFail = true
				redLines = append(redLines, tableLine(label, reason))
				b.ranRows++
				continue
			}
			b.ranRows++
			// The row's own tree mode (tree: unchanged): the snapshot
			// before its run, the comparison after it — a row-scoped
			// identity the whole-run control cannot see (a
			// write-then-delete pair inside one suite reconciles at the
			// end, but not per row).
			var modeCkp *Checkpoint
			if atom.Rows[i].Tree == canon.TreeModeUnchanged {
				if ckp2, serr := e.snapshotTreeIgnores(b.state.BuildIgnore); serr == nil {
					modeCkp = ckp2
				}
			}
			res := checks.Run(e.Workdir, label, atom.Rows[i], env)
			if res.Green {
				atom.Rows[i].Outcome, atom.Rows[i].Reason = canon.RowGreen, ""
				atom.Rows[i].EnvFail = false
			} else {
				atom.Rows[i].Outcome, atom.Rows[i].Reason = canon.RowRed, res.Reason
				atom.Rows[i].EnvFail = res.Env
				redLines = append(redLines, tableLine(label, res.Reason))
			}
			if modeCkp != nil {
				if paths, merr := e.treeMovedPathsIgnores(modeCkp, b.state.BuildIgnore); merr == nil && paths != "" {
					reason := canondata.T("gates.tree-moved-row", canondata.M{"paths": paths})
					atom.Rows[i].Outcome, atom.Rows[i].Reason = canon.RowRed, reason
					atom.Rows[i].EnvFail = true
					redLines = append(redLines, tableLine(label, reason))
				}
			}
			b.recordMemo(&atom.Rows[i], ckp, res)
		}
		b.putAtom(atom)
	}
	// The tree identity control: a digest after the run must equal
	// the one before it outside the ignores. A tree that moved under
	// its own measurement makes this transaction's greens non-proof:
	// they flip env-red with the moved paths, the suite reddens, the
	// acceptance cannot fire on a moving tree.
	if subErr == nil {
		if moved := e.treeMovedDuringIgnores(ckp, b.state.BuildIgnore); moved != "" {
			redLines = append(redLines, tableLine("tree", moved))
			for _, id := range sortedIDs(b.atoms) {
				atom := b.atoms[id]
				flipped := false
				for i := range atom.Rows {
					if atom.Rows[i].Outcome != canon.RowGreen {
						continue
					}
					atom.Rows[i].Outcome, atom.Rows[i].Reason = canon.RowRed, moved
					atom.Rows[i].EnvFail = true
					flipped = true
				}
				if flipped {
					b.putAtom(atom)
				}
			}
		}
	}
	e.lastSubmitChecks = b.ranRows
	// The suite line first, one line per red row after (the measured
	// values ride inside the reasons). The kept count names the
	// memoization: rows that did not re-run because nothing they
	// stand on moved.
	if b.ranRows > 0 || b.keptRows > 0 {
		total := b.ranRows + b.keptRows
		line := ""
		if len(redLines) == 0 {
			if b.keptRows > 0 && b.ranRows > 0 {
				line = canondata.T("gates.suite-green-kept", canondata.M{
					"rows": strconv.Itoa(total), "ran": strconv.Itoa(b.ranRows),
					"kept": strconv.Itoa(b.keptRows),
				})
			} else if b.keptRows > 0 {
				line = canondata.T("gates.suite-green-allkept", canondata.M{
					"rows": strconv.Itoa(total), "kept": strconv.Itoa(b.keptRows),
				})
			} else {
				line = canondata.T("gates.suite-green", canondata.M{
					"rows": strconv.Itoa(total),
				})
			}
		} else if b.keptRows > 0 {
			line = canondata.T("gates.suite-red-kept", canondata.M{
				"rows": strconv.Itoa(total), "red": strconv.Itoa(len(redLines)),
				"ran": strconv.Itoa(b.ranRows), "kept": strconv.Itoa(b.keptRows),
			})
		} else {
			line = canondata.T("gates.suite-red", canondata.M{
				"rows": strconv.Itoa(total), "red": strconv.Itoa(len(redLines)),
			})
		}
		b.runTable = append(b.runTable, line)
		budget := canondata.Limit("slot.list-budget")
		shown := redLines
		if len(shown) > budget {
			shown = shown[:budget]
		}
		b.runTable = append(b.runTable, shown...)
		if len(redLines) > len(shown) {
			b.runTable = append(b.runTable, canondata.T("slot.list.more", canondata.M{
				"count": strconv.Itoa(len(redLines) - len(shown)),
			}))
		}
	}
	// The checkpoint of the run: the tree the outcomes stand on is
	// fixed in the same transaction as the outcomes themselves.
	if subErr == nil {
		b.checkpointEffect(ckp, "")
	}
}

// ratifyPredicate — the ratification is a SIGNATURE, not a run: it
// never executes a row or the build. It signs exactly when the
// recorded green still stands — every row fresh at unchanged inputs
// and the surface digest equal to the digest of the last green run
// (the acceptance moment). Anything stale closes the door with the
// names; the honest route back is a plain run.
func (b *builder) ratifyPredicate() error {
	var stale []string
	for _, id := range sortedIDs(b.atoms) {
		atom := b.atoms[id]
		for i, row := range atom.Rows {
			label := canon.RowLabel(id, i+1)
			switch {
			case row.Outcome == canon.RowRed:
				stale = append(stale, label+" — "+canondata.T("gates.ratify-stale.red"))
			case !row.Memoized():
				stale = append(stale, label+" — "+canondata.T("gates.ratify-stale.never"))
			case row.Inputs != b.rowInputTokens(row):
				stale = append(stale, label+" — "+canondata.T("gates.ratify-stale.inputs"))
			}
		}
	}
	if len(stale) > 0 {
		return rejection("ratify", canondata.T("gates.ratify-stale", canondata.M{
			"count": strconv.Itoa(len(stale)),
			"rows":  budgetJoin(stale),
		}))
	}
	last := lastAcceptanceEntry(b.e)
	if last == nil {
		return rejection("ratify", canondata.T("gates.ratify-no-green"))
	}
	// The signature's ground: the surface as it stands now must be
	// the surface the last green run accepted. The snapshot is an
	// observation, not a run's checkpoint — nothing is persisted.
	ckp, err := b.e.snapshotTreeIgnores(b.state.BuildIgnore)
	if err != nil {
		return err
	}
	b.lastCkp = ckp
	digest := shortHex(ckp.Digest)
	if last.Details["digest"] != digest {
		reason := canondata.T("gates.ratify-drift.digests", canondata.M{
			"old": last.Details["digest"], "new": digest,
		})
		if before := b.e.checkpointOfDigest(last.Details["digest"]); before != nil {
			if paths := diffPaths(before, ckp); len(paths) > 0 {
				reason = canondata.T("gates.ratify-drift", canondata.M{
					"old": last.Details["digest"], "new": digest,
					"paths": budgetJoin(paths),
				})
			}
		}
		return rejection("ratify", reason)
	}
	rows := 0
	for _, id := range sortedIDs(b.atoms) {
		rows += len(b.atoms[id].Rows)
	}
	b.runTable = append(b.runTable, canondata.T("gates.ratify-signed", canondata.M{
		"rows": strconv.Itoa(rows), "digest": digest,
	}))
	return nil
}

// budgetJoin — the list items joined, capped by the reply budget
// with the honest "+N more" tail.
func budgetJoin(items []string) string {
	budget := canondata.Limit("slot.list-budget")
	shown := items
	if len(shown) > budget {
		shown = shown[:budget]
	}
	line := strings.Join(shown, "; ")
	if len(items) > len(shown) {
		line += fmt.Sprintf(" (+%d more)", len(items)-len(shown))
	}
	return line
}

// checkpointOfDigest — the stored checkpoint carrying the given
// short digest, walked from the newest: the file list the acceptance
// moment measured. nil — no stored snapshot matches (an externally
// pruned zone): the drift still refuses, without path names.
func (e *Engine) checkpointOfDigest(short string) *Checkpoint {
	state, err := e.Store.State()
	if err != nil {
		return nil
	}
	for seq := state.Counters["CKP"]; seq > 0; seq-- {
		data, err := fsport.Bind.ReadFile(filepath.Join(e.Workdir,
			filepath.FromSlash(checkpointFileName(seq))))
		if err != nil {
			continue
		}
		var ckp Checkpoint
		if yamlio.DecodeDoc(data, &ckp) != nil {
			continue
		}
		if shortHex(ckp.Digest) == short {
			return &ckp
		}
	}
	return nil
}

// diffPaths — the paths whose identity differs between two snapshots
// (added, replaced or removed alike; sorted).
func diffPaths(before, after *Checkpoint) []string {
	had := map[string]string{}
	for _, f := range before.Files {
		had[f.Path] = f.SHA256
	}
	var moved []string
	for _, f := range after.Files {
		old, ok := had[f.Path]
		delete(had, f.Path)
		if !ok || old != f.SHA256 {
			moved = append(moved, f.Path)
		}
	}
	for p := range had {
		moved = append(moved, p)
	}
	sort.Strings(moved)
	return moved
}

// acceptanceNote — the acceptance moment: all atoms green (a
// transaction that ran nothing finalizes nothing — a full-fresh
// ratification signs without re-running, that is its point) and no
// failed build finalize the full row set on this tree, the digest
// signs it.
func (b *builder) acceptanceNote(ckp *Checkpoint) {
	if ckp == nil || len(b.atoms) == 0 || b.buildFailed != "" {
		return
	}
	if b.ranRows == 0 && !b.ratifying {
		return
	}
	// One acceptance per tree digest: a full-fresh ratification signs
	// the tree the last green run already accepted — the moment is
	// recorded once, the snapshot carries it.
	digest := shortHex(ckp.Digest)
	if b.ranRows == 0 {
		if last := lastAcceptanceEntry(b.e); last != nil && last.Details["digest"] == digest {
			return
		}
	}
	green := 0
	total := 0
	for _, id := range sortedIDs(b.atoms) {
		if !b.atoms[id].Green() {
			return
		}
		green++
		total += len(b.atoms[id].Rows)
	}
	le := ledger.NewEntry("gate", mustState(b.e).Stage)
	le.Details = map[string]string{
		"name": "suite", "scope": "acceptance", "outcome": "green",
		"digest": digest,
		"green":  strconv.Itoa(total),
		"checks": strconv.Itoa(total),
	}
	if b.keptRows > 0 {
		le.Details["kept"] = strconv.Itoa(b.keptRows)
	}
	_ = b.e.Ledger.Append(le)
}

// tableLine — one red row's line of the reply table.
func tableLine(label, reason string) string {
	return "  " + label + ": " + reason
}

// inputDirOf — the instance's input directory (the sibling of the work
// root): seeds copy their from-input bytes from there.
func inputDirOf(workdir string) string {
	return filepath.Join(workdir, "..", "input")
}

// runBuild — the spec's build command once in the live tree before
// the rows: rc must be 0 within the mechanism's build ceiling. An
// empty build — nothing to build (the tree runs as is). What the
// build leaves outside build-ignore reddens the run as a tree that
// moved during the run — build artifacts are the ignores' business.
func (e *Engine) runBuild() error {
	state, err := e.Store.State()
	if err != nil {
		return err
	}
	if len(state.Build) == 0 {
		return nil
	}
	timeout := canondata.Limit("build.timeout-s")
	d := time.Duration(timeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), d)
	defer cancel()
	exit, output, timedOut, err := execport.Bind.Combined(ctx, state.Build, e.Workdir, d)
	switch {
	case err != nil:
		return err
	case timedOut:
		return errors.New(canondata.EnvDiag("command-timeout", canondata.M{
			"sec":     strconv.Itoa(timeout),
			"command": strings.Join(state.Build, " "),
		}))
	case exit != 0:
		text := oneLineInfo(string(output))
		if text == "" {
			text = fmt.Sprintf("exit %d", exit)
		}
		return errors.New(text)
	}
	return nil
}

// treeMovedDuring — the live-tree identity control by reading: a
// fresh snapshot after the run, compared with the pre-run one (both
// exclude .punchtape and build-ignore — the ignores are the only
// allowed garbage, the build outputs of the project). A tree that
// moved under its own measurement makes the run's outcomes
// non-proof; the line names the diverged paths.
func (e *Engine) treeMovedDuring(before *Checkpoint) string {
	state, err := e.Store.State()
	if err != nil {
		state = canon.State{}
	}
	return e.treeMovedDuringIgnores(before, state.BuildIgnore)
}

// treeMovedDuringIgnores — the post-run control under the same
// ignores the pre-run snapshot used.
func (e *Engine) treeMovedDuringIgnores(before *Checkpoint, buildIgnore []string) string {
	line, err := e.treeMovedPathsIgnores(before, buildIgnore)
	if err != nil {
		return canondata.T("gates.tree-moved-live", canondata.M{"paths": err.Error()})
	}
	if line == "" {
		return ""
	}
	return canondata.T("gates.tree-moved-live", canondata.M{"paths": line})
}

// treeMovedPathsIgnores — the diverged-paths line of the tree
// identity control ("" — the tree stood); the caller wraps it in its
// own vocabulary (the run control, the row's own tree mode). The
// error is the control's own failure (an unreadable tree).
func (e *Engine) treeMovedPathsIgnores(before *Checkpoint, buildIgnore []string) (string, error) {
	after, err := e.snapshotTreeIgnores(buildIgnore)
	if err != nil {
		return "", err
	}
	had := map[string]string{}
	for _, f := range before.Files {
		had[f.Path] = f.SHA256
	}
	var moved []string
	for _, f := range after.Files {
		old, ok := had[f.Path]
		delete(had, f.Path)
		if !ok || old != f.SHA256 {
			moved = append(moved, f.Path)
		}
	}
	for p := range had {
		moved = append(moved, p)
	}
	if len(moved) == 0 {
		return "", nil
	}
	sort.Strings(moved)
	shown := moved
	if len(shown) > canondata.Limit("slot.list-budget") {
		shown = shown[:canondata.Limit("slot.list-budget")]
	}
	line := strings.Join(shown, ", ")
	if len(moved) > len(shown) {
		line += fmt.Sprintf(" (+%d more)", len(moved)-len(shown))
	}
	return line, nil
}

func mustState(e *Engine) canon.State {
	st, err := e.Store.State()
	if err != nil {
		return canon.State{}
	}
	return st
}

func oneLineInfo(s string) string {
	s = strings.ReplaceAll(s, "\n", "; ")
	if len(s) > 400 {
		s = s[:400] + "…"
	}
	return s
}

func boolGreen(b bool) string {
	if b {
		return "green"
	}
	return "red"
}
