// Package journal — the journal of applied transactions. It provides
// two properties: submission idempotency (a submission key applies
// once) and state reproduction by replaying the journal.
package journal

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// Effect — one file effect of a transaction: write content or
// delete a file. The path is relative to the instance's working
// directory.
type Effect struct {
	Path    string `yaml:"path"`
	Content string `yaml:"content,omitempty"`
	Delete  bool   `yaml:"delete,omitempty"`
}

// Entry — one applied transaction. Effects are stored in full: the
// journal is the source of truth; state is reproduced from it after
// a failure. Rejected submissions do not reach the journal:
// replaying them gives the same answer by itself, without a record.
type Entry struct {
	Transaction   string    `yaml:"transaction"` // digest of the applied delta
	SubmissionKey string    `yaml:"submission-key"`
	DeltaKind     string    `yaml:"delta-kind"`
	WallMs        int64     `yaml:"wall-ms"`
	RecordedAt    time.Time `yaml:"recorded-at"`
	Effects       []Effect  `yaml:"effects"`
	// Origin — the submission origin of historical records: the
	// value "synthesized" was written by a removed machine assembly
	// path from a formal spec; the field is kept for reading old
	// journals.
	Origin string `yaml:"origin,omitempty"`
	// Held — the submission is held as a deferred change: recorded
	// in the journal, but nothing is applied until the human approves
	// (kind: amend).
	Held bool `yaml:"held,omitempty"`
	// ContentDigest — the digest of the delta's content (without
	// the submission key and the token trailer): content
	// idempotency — a repeat of the same delta under a fresh key
	// does not count as a new cycle. Legacy records without the
	// field are not compared by content (conservative).
	ContentDigest string `yaml:"content-digest,omitempty"`
}

// Journal — the journal's in-memory copy; the file is a stream of
// YAML documents, records are only appended.
type Journal struct {
	path    string
	entries []Entry
	torn    *yamlio.TornTail
}

// TornTail — the diagnosis of a torn tail this journal was loaded
// with: nil when the file is healthy. The intact records always
// load; the torn region is quarantined by the next Append.
func (j *Journal) TornTail() *yamlio.TornTail { return j.torn }

// Open reads the journal in full; a missing file is an empty
// journal. A torn tail (a record cut mid-write) does not blind the
// machine: every complete record loads, the torn region is carried
// as a named diagnosis and is quarantined by the next write.
func Open(path string) (*Journal, error) {
	j := &Journal{path: path}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return j, nil
		}
		return nil, fmt.Errorf("journal: %w", err)
	}
	preamble, frames := yamlio.SplitFrames(data)
	for i, fr := range frames {
		e, err := decodeFrame(fr)
		if err != nil {
			off := yamlio.FrameOffset(preamble, frames, i)
			j.torn = &yamlio.TornTail{Good: i, Offset: off, Length: len(data) - off}
			return j, nil
		}
		j.entries = append(j.entries, e)
	}
	return j, nil
}

// decodeFrame — one frame as one record, healthy only when it is
// EXACTLY a machine-written record: complete line (the writer ends
// every document with a newline), strict fields, and a byte-exact
// decode→marshal round trip — the append writes the marshaled bytes
// verbatim, so a frame that re-marshals differently is a write cut
// at a line boundary (its tail fields are gone), not a fact.
func decodeFrame(fr []byte) (Entry, error) {
	if len(fr) == 0 || fr[len(fr)-1] != '\n' {
		return Entry{}, errors.New("record cut mid-line")
	}
	var e Entry
	dec := yaml.NewDecoder(bytes.NewReader(fr))
	dec.KnownFields(true)
	if err := dec.Decode(&e); err != nil {
		return e, err
	}
	if e.Transaction == "" || e.RecordedAt.IsZero() {
		return e, errors.New("record missing its anchors")
	}
	remarshaled, err := yamlio.Marshal(&e)
	if err != nil {
		return e, err
	}
	if !bytes.Equal(remarshaled, fr) {
		return e, errors.New("record does not round-trip byte-exact")
	}
	return e, nil
}

// Append appends a record in one action and keeps a copy in memory.
// The document separator goes before the record: the file must not
// end with an empty document, otherwise it reads back as a phantom
// record.
func (j *Journal) Append(e Entry) error {
	data, err := yamlio.Marshal(e)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(j.path), 0o755); err != nil {
		return fmt.Errorf("journal dir: %w", err)
	}
	// A torn tail heals on the write path (the writer holds the
	// lock; read-only processes never touch the file): the intact
	// records are rebuilt verbatim, the torn region is quarantined
	// to a sidecar — never silently dropped.
	if j.torn != nil {
		if err := j.HealTail(); err != nil {
			return err
		}
	}
	// The manifest header is the first line of a new file: only the
	// machine writes the journal; it is not edited by hand; comments
	// do not break the YAML document stream.
	head := []byte(nil)
	if _, err := os.Stat(j.path); os.IsNotExist(err) {
		head = []byte(canon.MachineHeader("submit") + "\n")
	}
	f, err := os.OpenFile(j.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("journal %s: %w", j.path, err)
	}
	defer f.Close()
	if _, err := f.Write(append(append(head, []byte("---\n")...), data...)); err != nil {
		return fmt.Errorf("journal write: %w", err)
	}
	if err := f.Sync(); err != nil {
		return fmt.Errorf("journal sync: %w", err)
	}
	j.entries = append(j.entries, e)
	return nil
}

// HealTail — quarantine the torn region and rebuild the journal
// from the intact records: the machine's own write discipline
// (append, never edit) is restored; the sidecar keeps the torn
// bytes as evidence with the diagnosis. Caller holds the writer
// lock (the engine's verification or the next append).
func (j *Journal) HealTail() error {
	data, err := os.ReadFile(j.path)
	if err != nil {
		return fmt.Errorf("journal repair read: %w", err)
	}
	preamble, frames := yamlio.SplitFrames(data)
	good := frames
	if j.torn.Good < len(frames) {
		good = frames[:j.torn.Good]
	}
	off := yamlio.FrameOffset(preamble, frames, j.torn.Good)
	torn := data[min(off, len(data)):]
	q := j.path + ".torn"
	qf, err := os.OpenFile(q, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("journal quarantine: %w", err)
	}
	if _, err := fmt.Fprintf(qf, "# quarantined %s: %s\n", time.Now().UTC().Format(time.RFC3339), j.torn.Error()); err != nil {
		qf.Close()
		return fmt.Errorf("journal quarantine: %w", err)
	}
	if len(torn) > 0 {
		if _, err := qf.Write(torn); err != nil {
			qf.Close()
			return fmt.Errorf("journal quarantine: %w", err)
		}
	}
	if err := qf.Close(); err != nil {
		return fmt.Errorf("journal quarantine: %w", err)
	}
	if err := yamlio.WriteAtomic(j.path, yamlio.Rebuild(preamble, good)); err != nil {
		return fmt.Errorf("journal repair write: %w", err)
	}
	j.torn = nil
	return nil
}

// All — all records in apply order.
func (j *Journal) All() []Entry { return j.entries }

// FindByKey looks up an applied submission by key. Held records do
// not match: a hold is not a transaction with effects, its "replay"
// promises nothing; resubmitting the same content either re-stages
// the same hold idempotently, or (after the human's decision) is
// processed anew against the current state.
func (j *Journal) FindByKey(key string) (Entry, bool) {
	for _, e := range j.entries {
		if e.SubmissionKey == key && !e.Held {
			return e, true
		}
	}
	return Entry{}, false
}

// FindByContentDigest looks up an applied submission by content
// digest: the same delta under any key is the same fact; a repeat
// is not applied and does not count as a cycle. Held records do
// not match: they had no effect; re-submitting the same content
// after the hold is lifted is a legal new request, not a repeat.
func (j *Journal) FindByContentDigest(digest string) (Entry, bool) {
	if digest == "" {
		return Entry{}, false
	}
	for _, e := range j.entries {
		if e.ContentDigest == digest && !e.Held {
			return e, true
		}
	}
	return Entry{}, false
}
