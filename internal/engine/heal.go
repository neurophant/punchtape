// The self-damage cure, second half: damage the machine itself wrote.
// The legacy serializer could emit an atom file its own reader refused —
// every verb of the instance died on it, and the honest repair (the
// integrity check) prescribed a fresh instance, burning the whole
// whole work. The law split by authorship: bytes that digest exactly to the
// journal's own record but do not decode are TOOL damage — quarantine
// to a sidecar, journal the deletion, let the executor resubmit the
// behavior; bytes off the journal's record are someone else's edit —
// the honest integrity refusal stands (healing never launders
// manual edits).
package engine

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/fsport"
	"github.com/neurophant/punchtape/internal/journal"
	"github.com/neurophant/punchtape/internal/ledger"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// atomGate — the write gate seam: putAtom consults it before the
// bytes may enter the journal; the battery overrides it to prove the
// submission's refusal path (the guard law itself is proved by
// atomWriteGate's own cases).
var atomGate = atomWriteGate

// atomWriteGate — the machine never writes what it cannot read back
// (the round-trip law): an atom record's bytes must decode and re-marshal
// byte-exactly before they may enter the journal. The gate is the
// class guard — the serializer's exactness pass (yamlio) carries every
// grammar-valid text today; a form the pass misses dies here as a
// named refusal of the submission, never as a stuck instance.
func atomWriteGate(data []byte) error {
	var probe canon.Atom
	if err := yamlio.DecodeDoc(data, &probe); err != nil {
		return errors.New(canondata.T("store.atom.undecodable"))
	}
	// probe is plain data: marshal cannot fail
	again, _ := yamlio.Marshal(probe)
	if !bytes.Equal(again, data) {
		return errors.New(canondata.T("store.atom.noncanonical"))
	}
	return nil
}

// healToolBrokenAtoms — quarantine the machine-written undecodable
// atom files (see the file comment). Runs at open after the integrity
// reconciliation, under the writer lock (the same monopoly every open
// already takes): at that point every canon file digests to the
// journal's record, so an undecodable one is tool damage by
// definition — the digest is still verified, belt to the walk order.
// A healthy instance pays one decode walk. The heal is a transaction
// of its own: the quarantine sidecars keep the damaged bytes as
// evidence, the deletions ride the journal, the manifest follows.
func (e *Engine) healToolBrokenAtoms() error {
	lock, err := e.acquireWriterRetry()
	if err != nil {
		return err
	}
	defer e.releaseWriter(lock)
	toolBroken, err := e.toolBrokenAtoms()
	if err != nil || len(toolBroken) == 0 {
		return err
	}
	broken := sortedIDs(toolBroken)
	var effects []journal.Effect
	var names []string
	var quarantine []byte
	for _, rel := range broken {
		data := toolBroken[rel]
		sidecar := rel + ".torn"
		stamp := time.Now().UTC().Format(time.RFC3339)
		chunk := fmt.Appendf(nil, "# quarantined %s: the machine wrote these bytes and cannot read them back (heal)\n", stamp)
		chunk = append(chunk, data...)
		if err := yamlio.WriteAtomic(filepath.Join(e.Workdir, sidecar), chunk); err != nil {
			return fmt.Errorf("heal quarantine %s: %w", sidecar, err)
		}
		effects = append(effects, journal.Effect{Path: rel, Delete: true})
		names = append(names, filepath.Base(rel))
		quarantine = append(quarantine, chunk...)
	}
	entry := journal.Entry{
		Transaction:   yamlio.Digest(quarantine),
		SubmissionKey: "heal",
		DeltaKind:     "heal",
		WallMs:        timeSince(timeNow()).Milliseconds(),
		RecordedAt:    timeNow().UTC(),
		Effects:       effects,
	}
	if err := e.commit(entry); err != nil {
		return err
	}
	le := ledger.NewEntry("heal", mustState(e).Stage)
	le.Details = map[string]string{"atoms": strings.Join(names, ","), "outcome": "quarantined"}
	return e.Ledger.Append(le)
}

// toolBrokenAtoms — the canon atom files that fail the machine's own
// decode while digesting exactly to the journal's record of them:
// tool damage, healable, with the damaged bytes attached. A file
// without a journal record never reaches here (the integrity walk
// refused it first).
func (e *Engine) toolBrokenAtoms() (map[string][]byte, error) {
	expected := foldInstanceFiles(e.Journal.All())
	broken := map[string][]byte{}
	for _, c := range canon.Collections() {
		dir := filepath.Join(e.Store.Root(), filepath.FromSlash(c))
		entries, err := fsport.Bind.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("heal dir %s: %w", c, err)
		}
		for _, en := range entries {
			if en.IsDir() || !strings.HasSuffix(en.Name(), ".yaml") {
				continue
			}
			rel := ".punchtape/" + c + "/" + en.Name()
			data, err := fsport.Bind.ReadFile(filepath.Join(dir, en.Name()))
			if err != nil {
				return nil, fmt.Errorf("heal read %s: %w", rel, err)
			}
			var probe canon.Atom
			if yamlio.DecodeDoc(data, &probe) == nil {
				continue
			}
			want, ok := expected[rel]
			if ok && want == yamlio.Digest(data) {
				broken[rel] = data
			}
		}
	}
	return broken, nil
}
