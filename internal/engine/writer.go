// The single writer of the instance: all .punchtape
// dynamics are written by core verbs only. A transaction runs under the
// writer lock (flock) in the order "journal → marker → effects →
// lastgood → marker removal"; effect paths inside the instance are
// limited by a white list, canon names by a charset validator, and
// after a write the file is read back. Opening an instance reconciles
// the files against the lastgood manifest: a manual edit is detected
// by an honest refusal with a list, a crash mid-transaction is
// completed by journal replay.
package engine

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/journal"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// canonNamePattern — a canon file name: lowercase letters, digits, hyphens.
var canonNamePattern = regexp.MustCompile(`^[a-z0-9-]+\.yaml$`)

// Product passport paths (conventions and the boundary registry) under
// the writer discipline: written only by machine transactions.
const (
	conventionsFilePath = ".punchtape/passport/conventions.yaml"
	boundariesDirPath   = ".punchtape/passport/boundaries/"
)

// Instance paths under the writer discipline.
const (
	stateFilePath   = ".punchtape/state.yaml"
	journalFilePath = ".punchtape/journal.yamll"
)

// markerPath — the marker of an unfinished transaction.
func (e *Engine) markerPath() string {
	return filepath.Join(e.Store.Root(), "cache", "pending-transaction")
}

// lastgoodPath — the manifest of the last machine state of the files.
func (e *Engine) lastgoodPath() string {
	return filepath.Join(e.Store.Root(), "cache", "lastgood.yamll")
}

// ErrWriterBusy — a busy writer lock: an instance never has a second
// writer, the caller is entitled to tell an honest refusal from corruption.
var ErrWriterBusy = errors.New(canondata.T("store.writer.busy"))

// acquireWriter takes the instance's write monopoly: a non-blocking
// flock on cache/writer.lock. A busy lock is an honest refusal: an
// instance has no second writer. The cache is rebuilt (the service
// zone), a missing directory is created in place — a wiped cache does
// not break the machine. Re-acquisition by the same process is a
// depth counter: a submission holds the writer for its whole
// read-modify-write transaction, and its inner commits reuse the same
// lock instead of waiting on themselves.
func (e *Engine) acquireWriter() (*os.File, error) {
	if e.writerDepth > 0 && e.writerLock != nil {
		e.writerDepth++
		return e.writerLock, nil
	}
	lockPath := filepath.Join(e.Store.Root(), "cache", "writer.lock")
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		return nil, fmt.Errorf("writer lock: %w", err)
	}
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, fmt.Errorf("writer lock: %w", err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		return nil, ErrWriterBusy
	}
	e.writerLock = f
	e.writerDepth = 1
	return f, nil
}

// releaseWriter releases the writer lock: the outer one (depth 0) by
// simply closing; a nested one by the counter — the file lives until
// the last.
func (e *Engine) releaseWriter(f *os.File) {
	if e.writerDepth == 0 || f != e.writerLock {
		_ = f.Close()
		return
	}
	e.writerDepth--
	if e.writerDepth == 0 {
		e.writerLock = nil
		_ = f.Close()
	}
}

// acquireWriterRetry — acquireWriter with short retries: transient
// contention (a parallel read-recovery of opening) need not wreck a
// submission; persistent contention stays an honest refusal.
func (e *Engine) acquireWriterRetry() (*os.File, error) {
	var lock *os.File
	var err error
	for try := 0; try < 3; try++ {
		lock, err = e.acquireWriter()
		if err == nil || !errors.Is(err, ErrWriterBusy) {
			return lock, err
		}
		time.Sleep(250 * time.Millisecond)
	}
	return lock, err
}

// commit runs a transaction under the writer lock: journal → marker →
// effects → lastgood → marker removal. A failure at any step leaves
// the instance completable: the marker names the transaction, the
// effects are idempotent.
func (e *Engine) commit(entry journal.Entry) error {
	lock, err := e.acquireWriter()
	if err != nil {
		return err
	}
	defer e.releaseWriter(lock)
	if err := e.Journal.Append(entry); err != nil {
		return err
	}
	if err := yamlio.WriteAtomic(e.markerPath(), []byte(entry.Transaction)); err != nil {
		return err
	}
	if err := applyEffects(e.Workdir, entry.Effects); err != nil {
		return err
	}
	if err := e.writeLastgood(entry.Transaction); err != nil {
		return err
	}
	return os.Remove(e.markerPath())
}

// checkEffectPath — the white list of instance writes: a transaction
// writes only files of zones with the journaled rule — the zone model:
// state, passport, and canon collections with charset names; everything
// else inside .punchtape is forbidden. Project files are already
// validated by the delta (card ownership).
func checkEffectPath(p string) error {
	p = filepath.ToSlash(p)
	if !strings.HasPrefix(p, ".punchtape/") {
		return nil
	}
	rel := strings.TrimPrefix(p, ".punchtape/")
	if !canon.JournaledPath(rel) {
		return errors.New(canondata.T("store.effect.whitelist", canondata.M{"path": p}))
	}
	if strings.HasPrefix(rel, "canon/") || strings.HasPrefix(rel, "passport/boundaries/") {
		name := p[strings.LastIndexByte(p, '/')+1:]
		if !canonNamePattern.MatchString(name) {
			return errors.New(canondata.T("store.effect.charset", canondata.M{"name": name}))
		}
	}
	return nil
}

// applyEffects applies a transaction's file effects to the working
// directory: a write is atomic and read back (the digest of the written
// bytes must match), deleting a missing file is not an error.
func applyEffects(workdir string, effects []journal.Effect) error {
	for _, ef := range effects {
		if err := checkEffectPath(ef.Path); err != nil {
			return err
		}
		path := filepath.Join(workdir, ef.Path)
		if ef.Delete {
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("delete %s: %w", ef.Path, err)
			}
			continue
		}
		if err := yamlio.WriteAtomic(path, []byte(ef.Content)); err != nil {
			return fmt.Errorf("write %s: %w", ef.Path, err)
		}
		back, err := os.ReadFile(path)
		if err != nil || yamlio.Digest(back) != yamlio.Digest([]byte(ef.Content)) {
			return errors.New(canondata.T("store.effect.readback", canondata.M{"path": ef.Path}))
		}
	}
	return nil
}

// verifyInstance — the discipline of opening under the writer lock: a
// marker of an unfinished transaction is completed by replay; without
// a marker the instance's files are reconciled against the lastgood
// manifest (a mismatch — a manual edit, an honest refusal with a
// list); the journal head ahead of the manifest with files intact — a
// crash in the "journal → marker" window, the tail is completed by
// replay; a legacy instance without a manifest derives the expectation
// from the journal, reconciles, and fixes the manifest; a virgin
// instance without transactions is skipped.
func (e *Engine) verifyInstance() error {
	lock, err := e.acquireWriter()
	if err != nil {
		return err
	}
	defer e.releaseWriter(lock)
	data, err := os.ReadFile(e.markerPath())
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("transaction marker: %w", err)
		}
	} else {
		return e.recoverMarked(strings.TrimSpace(string(data)))
	}
	head := journalHeadTx(e.Journal.All())
	m, err := e.loadLastgood()
	if err != nil {
		return err
	}
	if m == nil {
		if head == "" {
			return nil
		}
		return e.adoptLegacyInstance(head)
	}
	if diff := e.diffManifest(m); len(diff) > 0 {
		return e.integrityRefusal(diff)
	}
	if m.Head == head {
		return nil
	}
	// The files are intact per the manifest, but the journal head moved
	// ahead: the tail's effects were not applied — the crash window
	// between journal and marker.
	if err := e.replayAfter(m.Head); err != nil {
		return err
	}
	return e.writeLastgood(head)
}

// recoverMarked completes the transaction named by the marker: the
// effects are idempotent, re-applying is safe; the manifest is
// rewritten to the completed head, the marker is removed.
func (e *Engine) recoverMarked(tx string) error {
	for _, entry := range e.Journal.All() {
		if entry.Transaction != tx {
			continue
		}
		if err := applyEffects(e.Workdir, entry.Effects); err != nil {
			return fmt.Errorf("recovery of transaction %s: %w", shortHex(tx), err)
		}
		if err := e.writeLastgood(entry.Transaction); err != nil {
			return err
		}
		return os.Remove(e.markerPath())
	}
	// The marker names a transaction the journal does not have: the
	// journal lacks what the machine wrote — the journal was corrupted
	// by hand.
	return e.integrityRefusal([]integrityDiff{{kind: integrityDiffers, path: journalFilePath}})
}

// replayAfter applies the effects of all transactions after the named
// one: completing a crash in the "journal → marker" window, when the
// files stayed at the previous manifest.
func (e *Engine) replayAfter(tx string) error {
	entries := e.Journal.All()
	for i, entry := range entries {
		if entry.Transaction != tx {
			continue
		}
		for _, en := range entries[i+1:] {
			if err := applyEffects(e.Workdir, en.Effects); err != nil {
				return fmt.Errorf("replay of transaction %s: %w", shortHex(en.Transaction), err)
			}
		}
		return nil
	}
	return e.integrityRefusal([]integrityDiff{{kind: integrityDiffers, path: journalFilePath}})
}

// --- lastgood manifest ---

// lastgoodFile — a manifest line: the path from the project root and the digest.
type lastgoodFile struct {
	Path   string `yaml:"path"`
	Digest string `yaml:"digest"`
}

// lastgoodManifest — a snapshot of the machine state of the instance's
// files at the journal head. Invariant: no marker ⇒ the snapshot
// matches the journal head.
type lastgoodManifest struct {
	Head  string         `yaml:"journal-head"`
	Files []lastgoodFile `yaml:"files"`
}

// writeLastgood fixes the manifest at the given head: the expectation
// is derived by folding the journal's effects. state.yaml without
// state effects in the journal (the instance was only initialized) is
// taken from the observed content: the first transaction writing
// state will make the record authoritative.
func (e *Engine) writeLastgood(head string) error {
	files := foldInstanceFiles(e.Journal.All())
	if _, ok := files[stateFilePath]; !ok {
		if data, err := os.ReadFile(filepath.Join(e.Workdir, stateFilePath)); err == nil {
			files[stateFilePath] = yamlio.Digest(data)
		}
	}
	m := lastgoodManifest{Head: head}
	for _, p := range sortedIDs(files) {
		m.Files = append(m.Files, lastgoodFile{Path: p, Digest: files[p]})
	}
	data, err := yamlio.Marshal(m)
	if err != nil {
		return err
	}
	data = append([]byte(canon.CacheHeader()+"\n"), data...)
	return yamlio.WriteAtomic(e.lastgoodPath(), data)
}

// loadLastgood reads the manifest; absence is not an error (legacy or
// a virgin instance). An unreadable or wrong-shaped manifest is
// corruption by hand: an honest refusal.
func (e *Engine) loadLastgood() (*lastgoodManifest, error) {
	data, err := os.ReadFile(e.lastgoodPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("lastgood: %w", err)
	}
	var m lastgoodManifest
	if err := yamlio.DecodeStrict(data, &m); err != nil {
		return nil, e.integrityRefusal([]integrityDiff{
			{kind: integrityDiffers, path: filepath.ToSlash(filepath.Join(".punchtape", "cache", "lastgood.yamll"))},
		})
	}
	return &m, nil
}

// foldInstanceFiles derives the expectation of the instance's files
// from the journal: the last effect on each path under the writer
// discipline.
func foldInstanceFiles(entries []journal.Entry) map[string]string {
	files := map[string]string{}
	for _, en := range entries {
		for _, ef := range en.Effects {
			p := filepath.ToSlash(ef.Path)
			if !instanceScoped(p) {
				continue
			}
			if ef.Delete {
				delete(files, p)
				continue
			}
			files[p] = yamlio.Digest([]byte(ef.Content))
		}
	}
	return files
}

// instanceScoped — a path under the writer's reconciliation discipline:
// canon, passport, and state zones — the zone model.
func instanceScoped(p string) bool {
	if !strings.HasPrefix(p, ".punchtape/") {
		return false
	}
	return canon.JournaledPath(strings.TrimPrefix(p, ".punchtape/"))
}

// journalHeadTx — the full transaction of the journal head; an empty
// journal is an empty string.
func journalHeadTx(entries []journal.Entry) string {
	if len(entries) == 0 {
		return ""
	}
	return entries[len(entries)-1].Transaction
}

// adoptLegacyInstance adopts an instance without a manifest: the
// expectation is derived from the journal, the files are reconciled,
// the manifest is fixed. state.yaml without a state effect in the
// journal was written by initialization — there is nothing to
// reconcile it against, the observed content becomes the base.
func (e *Engine) adoptLegacyInstance(head string) error {
	expected := foldInstanceFiles(e.Journal.All())
	var diffs []integrityDiff
	for _, p := range sortedIDs(expected) {
		data, err := os.ReadFile(filepath.Join(e.Workdir, p))
		switch {
		case err != nil:
			diffs = append(diffs, integrityDiff{kind: integrityMissing, path: p})
		case yamlio.Digest(data) != expected[p]:
			diffs = append(diffs, integrityDiff{kind: integrityDiffers, path: p})
		}
	}
	listed := map[string]bool{}
	for p := range expected {
		listed[p] = true
	}
	for _, p := range e.scanInstanceFiles() {
		// state.yaml outside the expectation is the initialization
		// base, the manifest will record it; other files without a
		// machine record are manual.
		if !listed[p] && p != stateFilePath {
			diffs = append(diffs, integrityDiff{kind: integrityUnjournaled, path: p})
		}
	}
	if len(diffs) > 0 {
		return e.integrityRefusal(diffs)
	}
	return e.writeLastgood(head)
}

// --- reconciling files against the manifest ---

// Kinds of mismatches between the instance's files and the machine state.
const (
	integrityDiffers = iota
	integrityUnjournaled
	integrityMissing
)

// integrityDiff — one mismatch: the path and the kind.
type integrityDiff struct {
	kind int
	path string
}

// diffManifest reconciles the instance's files on disk against the manifest.
func (e *Engine) diffManifest(m *lastgoodManifest) []integrityDiff {
	listed := map[string]bool{}
	var diffs []integrityDiff
	for _, f := range m.Files {
		listed[f.Path] = true
		data, err := os.ReadFile(filepath.Join(e.Workdir, f.Path))
		switch {
		case err != nil:
			diffs = append(diffs, integrityDiff{kind: integrityMissing, path: f.Path})
		case yamlio.Digest(data) != f.Digest:
			diffs = append(diffs, integrityDiff{kind: integrityDiffers, path: f.Path})
		}
	}
	for _, p := range e.scanInstanceFiles() {
		if !listed[p] {
			diffs = append(diffs, integrityDiff{kind: integrityUnjournaled, path: p})
		}
	}
	sort.Slice(diffs, func(i, j int) bool { return diffs[i].path < diffs[j].path })
	return diffs
}

// scanInstanceFiles — the instance's actual files under the
// reconciliation discipline: canon collections, passport, and state
// (the journaled zones).
func (e *Engine) scanInstanceFiles() []string {
	var out []string
	for _, c := range canon.Collections() {
		dir := filepath.Join(e.Store.Root(), filepath.FromSlash(c))
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, en := range entries {
			if !en.IsDir() && strings.HasSuffix(en.Name(), ".yaml") {
				out = append(out, ".punchtape/"+c+"/"+en.Name())
			}
		}
	}
	if _, err := os.Stat(filepath.Join(e.Store.Root(), "passport", "conventions.yaml")); err == nil {
		out = append(out, conventionsFilePath)
	}
	bdir := filepath.Join(e.Store.Root(), "passport", "boundaries")
	if entries, err := os.ReadDir(bdir); err == nil {
		for _, en := range entries {
			if !en.IsDir() && strings.HasSuffix(en.Name(), ".yaml") {
				out = append(out, boundariesDirPath+en.Name())
			}
		}
	}
	if _, err := os.Stat(filepath.Join(e.Store.Root(), "state.yaml")); err == nil {
		out = append(out, stateFilePath)
	}
	sort.Strings(out)
	return out
}

// integrityRefusal — an honest refusal with a list of mismatches: the
// machine is the only writer, the journal is intact, the files are
// recoverable.
func (e *Engine) integrityRefusal(diffs []integrityDiff) error {
	var sb strings.Builder
	sb.WriteString(canondata.T("store.integrity.header"))
	for _, d := range diffs {
		var line string
		switch d.kind {
		case integrityDiffers:
			line = canondata.T("store.integrity.differs", canondata.M{"path": d.path})
		case integrityUnjournaled:
			line = canondata.T("store.integrity.unjournaled", canondata.M{"path": d.path})
		case integrityMissing:
			line = canondata.T("store.integrity.missing", canondata.M{"path": d.path})
		}
		sb.WriteString("\n")
		sb.WriteString(line)
	}
	return errors.New(sb.String())
}
