package store

// The journal's file stream over the epoch law: records append to
// the live epoch file, one YAML document per record behind a ---
// separator; a submission rotates the epoch, so every submission's
// records live in an epoch of their own. A torn tail (a record cut
// mid-write) does not blind the machine: the complete records load,
// the torn bytes quarantine to a sidecar, and the epoch file is
// rebuilt from the intact records — the write discipline restored,
// the evidence kept.

import (
	"bytes"
	"fmt"

	"github.com/neurophant/punchtape/internal/domain/journal"
	"github.com/neurophant/punchtape/internal/ports"
)

// Stream — the journal's records over the epoch files of a
// directory.
type Stream struct {
	dir   string
	files []string
}

// OpenStream — the stream over a journal directory; a missing
// directory is a stream that never wrote.
func OpenStream(dir string) (*Stream, error) {
	files, err := ListEpochs(dir)
	if err != nil {
		return nil, fmt.Errorf("journal: %w", err)
	}
	return &Stream{dir: dir, files: files}, nil
}

// Records — every record of every epoch in order.
func (s *Stream) Records() ([]journal.Record, error) {
	var out []journal.Record
	for _, path := range s.files {
		recs, err := s.loadFile(path)
		if err != nil {
			return nil, err
		}
		out = append(out, recs...)
	}
	return out, nil
}

// FirstRecord — the journal's first record alone: the birth fact
// without asking any later record to parse. A foreign machine's zone
// is named by its creator before the fold meets the zone's shapes. A
// journal that never wrote answers the zero record.
func (s *Stream) FirstRecord() (journal.Record, error) {
	if len(s.files) == 0 {
		return journal.Record{}, nil
	}
	data, err := ports.FSBind.ReadFile(s.files[0])
	if err != nil {
		return journal.Record{}, fmt.Errorf("journal: %w", err)
	}
	chunks := splitRecords(data)
	if len(chunks) == 0 {
		return journal.Record{}, nil
	}
	var rec journal.Record
	if err := journal.DecodeRecord(chunks[0], &rec); err != nil {
		return journal.Record{}, fmt.Errorf("journal: record 0 of %s does not parse: %w", s.files[0], err)
	}
	return rec, nil
}

// loadFile — one epoch file's records; a torn tail loads the intact
// records and quarantines the torn bytes. A chunk that does not
// parse outside the final position is damage by hand — an honest
// error, never a silent drop.
func (s *Stream) loadFile(path string) ([]journal.Record, error) {
	data, err := ports.FSBind.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("journal: %w", err)
	}
	chunks := splitRecords(data)
	var out []journal.Record
	for i, chunk := range chunks {
		var rec journal.Record
		if err := journal.DecodeRecord(chunk, &rec); err != nil {
			if i == len(chunks)-1 {
				if err := quarantine(path, data, chunks[:i]); err != nil {
					return nil, err
				}
				break
			}
			return nil, fmt.Errorf("journal: record %d of %s does not parse: %w", i, path, err)
		}
		out = append(out, rec)
	}
	return out, nil
}

// Append — one record to the live epoch, in one action.
func (s *Stream) Append(rec journal.Record) error {
	data := journal.MarshalRecord(rec)
	if err := ports.FSBind.MkdirAll(s.dir); err != nil {
		return fmt.Errorf("journal dir: %w", err)
	}
	if len(s.files) == 0 {
		first := NextEpochPath(nil, s.dir)
		if err := ports.FSBind.WriteAtomic(first, nil); err != nil {
			return fmt.Errorf("journal rotate: %w", err)
		}
		s.files = append(s.files, first)
	}
	path := s.files[len(s.files)-1]
	if err := ports.FSBind.AppendFile(path, append([]byte("---\n"), data...)); err != nil {
		return fmt.Errorf("journal write: %w", err)
	}
	return nil
}

// Rotate — close the live epoch: the next append opens a fresh
// file. A submission rotates before its records land, so its epoch
// is its own; an already-fresh epoch stays — rotation must not stack
// empty files.
func (s *Stream) Rotate() error {
	if len(s.files) > 0 && s.lastIsEmpty() {
		return nil
	}
	if err := ports.FSBind.MkdirAll(s.dir); err != nil {
		return fmt.Errorf("journal dir: %w", err)
	}
	next := NextEpochPath(s.files, s.dir)
	if err := ports.FSBind.WriteAtomic(next, nil); err != nil {
		return fmt.Errorf("journal rotate: %w", err)
	}
	s.files = append(s.files, next)
	return nil
}

// lastIsEmpty — whether the live epoch carries no records yet.
func (s *Stream) lastIsEmpty() bool {
	data, err := ports.FSBind.ReadFile(s.files[len(s.files)-1])
	return err == nil && len(bytes.TrimSpace(data)) == 0
}

// quarantine — the torn tail to a sidecar, the epoch file rebuilt
// byte-for-byte from its intact chunks.
func quarantine(path string, data []byte, intact [][]byte) error {
	good := joinChunks(intact)
	cut := len(data)
	if bytes.HasPrefix(data, good) {
		cut = len(good)
	}
	torn := data[cut:]
	if len(bytes.TrimSpace(torn)) == 0 {
		return nil
	}
	sidecar := append([]byte("# quarantined torn record\n"), torn...)
	if err := ports.FSBind.AppendFile(path+".torn", sidecar); err != nil {
		return fmt.Errorf("journal quarantine: %w", err)
	}
	if err := ports.FSBind.WriteAtomic(path, good); err != nil {
		return fmt.Errorf("journal repair write: %w", err)
	}
	return nil
}

// joinChunks — the epoch file's bytes from its record chunks: every
// record behind a line-start --- separator.
func joinChunks(chunks [][]byte) []byte {
	var out []byte
	for _, ch := range chunks {
		out = append(out, "---\n"...)
		out = append(out, ch...)
	}
	return out
}

// splitRecords — the document chunks of an epoch file: the bytes
// between the --- separators, in order; the leading chunk before the
// first separator is empty by construction, trailing emptiness is
// not a chunk. A separator matches only at a line's start, so
// indented content carrying dashes never splits.
func splitRecords(data []byte) [][]byte {
	var out [][]byte
	rest := data
	if bytes.HasPrefix(rest, []byte("---\n")) {
		rest = rest[4:]
	}
	for {
		i := indexSeparator(rest)
		if i < 0 {
			if len(bytes.TrimSpace(rest)) > 0 {
				out = append(out, rest)
			}
			return out
		}
		out = append(out, rest[:i+1]) // the record keeps its newline
		rest = rest[i+5:]
	}
}

// indexSeparator — the next line-start --- separator.
func indexSeparator(b []byte) int {
	for i := 0; i+4 < len(b); i++ {
		if b[i] == '\n' && b[i+1] == '-' && b[i+2] == '-' && b[i+3] == '-' && b[i+4] == '\n' {
			return i
		}
	}
	return -1
}
