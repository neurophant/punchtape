// Package journal — the journal of applied transactions. It provides
// two properties: submission idempotency (a submission key applies
// once) and state reproduction by replaying the journal.
package journal

import (
	"bytes"
	"fmt"
	"io"
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
}

// Open reads the journal in full; a missing file is an empty
// journal.
func Open(path string) (*Journal, error) {
	j := &Journal{path: path}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return j, nil
		}
		return nil, fmt.Errorf("journal: %w", err)
	}
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	for {
		var e Entry
		err := dec.Decode(&e)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("journal %s: %w", path, err)
		}
		j.entries = append(j.entries, e)
	}
	return j, nil
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
