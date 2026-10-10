// Package ledger — the journal of machine facts: pipeline meters
// (wall time, calls, tokens) by event and stage. The file is a
// stream of YAML documents; records are only appended, one per
// write.
package ledger

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/neurophant/punchtape/internal/fsport"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// Entry — one journal record. Artifact and reply bytes are
// pointers not for beauty: null means "not measured", which is also a
// fact and must not disappear from the file.
type Entry struct {
	Event    string  `yaml:"event"`
	Stage    string  `yaml:"stage"`
	Artifact *string `yaml:"artifact"`
	WallMs   int64   `yaml:"wall-ms"`
	Calls    int64   `yaml:"calls"`
	// ReplyBytes — the length of the machine's reply in bytes (the
	// next/submit verbs); null — the event measures no reply.
	ReplyBytes *int64            `yaml:"reply-bytes"`
	Details    map[string]string `yaml:"details"`
	At         time.Time         `yaml:"recorded-at"`
	Start      time.Time         `yaml:"-"` // measurement start; Append computes the wall time from it
}

// NewEntry starts a record: fixes the time and starts the
// stopwatch.
func NewEntry(event, stage string) Entry {
	now := time.Now().UTC().Truncate(time.Millisecond)
	return Entry{Event: event, Stage: stage, At: now, Start: now}
}

// Duration — how much time has passed since the record started.
func (e Entry) Duration() time.Duration { return time.Since(e.Start) }

// Ledger — a journal fully loaded into memory.
type Ledger struct {
	path    string
	entries []Entry
	torn    *yamlio.TornTail
}

// TornTail — the diagnosis of a torn tail this journal was loaded
// with: nil when the file is healthy. The intact records are always
// loaded; the torn region is quarantined by the next Append.
func (l *Ledger) TornTail() *yamlio.TornTail { return l.torn }

// Open reads the journal in full. No file — an empty journal, not
// an error. Foreign fields in records are rejected: silently
// swallowing corrupted facts is not allowed. A torn tail (a record
// cut mid-write) does not blind the machine: every complete record
// loads, the torn region is carried as a named diagnosis and is
// quarantined by the next write.
func Open(path string) (*Ledger, error) {
	l := &Ledger{path: path}
	data, err := fsport.Bind.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return l, nil
		}
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	preamble, frames := yamlio.SplitFrames(data)
	for i, fr := range frames {
		e, err := decodeFrame(fr)
		if err != nil {
			off := yamlio.FrameOffset(preamble, frames, i)
			l.torn = &yamlio.TornTail{Good: i, Offset: off, Length: len(data) - off}
			return l, nil
		}
		l.entries = append(l.entries, e)
	}
	return l, nil
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
	if err := yamlio.DecodeDoc(fr, &e); err != nil {
		return e, err
	}
	if e.Event == "" || e.At.IsZero() {
		return e, errors.New("record missing its anchors")
	}
	// Entry is plain scalar/slice/time data: marshal cannot fail
	remarshaled, _ := yamlio.Marshal(&e)
	if !bytes.Equal(remarshaled, fr) {
		return e, errors.New("record does not round-trip byte-exact")
	}
	return e, nil
}

// Append appends a record to the end of the file in one write with
// fsync and refills memory with the same record. An empty wall time
// is computed from Start, an empty record time — now; NewEntry plus
// Append are enough for the caller.
func (l *Ledger) Append(e Entry) error {
	if e.At.IsZero() {
		e.At = time.Now().UTC().Truncate(time.Millisecond)
	}
	if e.WallMs == 0 && !e.Start.IsZero() {
		e.WallMs = time.Since(e.Start).Milliseconds()
	}
	// Entry is plain scalar/slice/time data: marshal cannot fail
	doc, _ := yamlio.Marshal(e)

	if err := fsport.Bind.MkdirAll(filepath.Dir(l.path)); err != nil {
		return fmt.Errorf("dir %s: %w", filepath.Dir(l.path), err)
	}
	// A torn tail heals on the write path (the writer holds the
	// lock; read-only processes never touch the file): the intact
	// records are rebuilt verbatim, the torn region is quarantined
	// to a sidecar — never silently dropped.
	if l.torn != nil {
		if err := l.repair(); err != nil {
			return err
		}
	}
	// A new file starts straight with the first frame separator:
	// the ledger is written by core verbs and never edited by hand;
	// YAML carries no comment header — machine authorship is a fact
	// of the format contract, not a line in the file.
	if err := fsport.Bind.AppendFile(l.path, append([]byte("---\n"), doc...)); err != nil {
		return fmt.Errorf("write to %s: %w", l.path, err)
	}

	l.entries = append(l.entries, e)
	return nil
}

// repair — quarantine the torn region and rebuild the journal from
// the intact records: the machine's own write discipline (append,
// never edit) is restored; the sidecar keeps the torn bytes as
// evidence with the diagnosis.
func (l *Ledger) repair() error {
	data, err := fsport.Bind.ReadFile(l.path)
	if err != nil {
		return fmt.Errorf("repair read %s: %w", l.path, err)
	}
	preamble, frames := yamlio.SplitFrames(data)
	good := frames
	if l.torn.Good < len(frames) {
		good = frames[:l.torn.Good]
	}
	off := yamlio.FrameOffset(preamble, frames, l.torn.Good)
	torn := data[min(off, len(data)):]
	q := l.path + ".torn"
	var sidecar []byte
	sidecar = fmt.Appendf(sidecar, "# quarantined %s: %s\n", time.Now().UTC().Format(time.RFC3339), l.torn.Error())
	sidecar = append(sidecar, torn...)
	if err := fsport.Bind.AppendFile(q, sidecar); err != nil {
		return fmt.Errorf("quarantine %s: %w", q, err)
	}
	if err := yamlio.WriteAtomic(l.path, yamlio.Rebuild(preamble, good)); err != nil {
		return fmt.Errorf("repair write %s: %w", l.path, err)
	}
	l.torn = nil
	return nil
}

// All — all records in file order.
func (l *Ledger) All() []Entry { return l.entries }

// ByEvent — one event's records in file order.
func (l *Ledger) ByEvent(event string) []Entry {
	var out []Entry
	for _, e := range l.entries {
		if e.Event == event {
			out = append(out, e)
		}
	}
	return out
}

// CountBy — how many records an event has.
func (l *Ledger) CountBy(event string) int64 {
	var n int64
	for _, e := range l.entries {
		if e.Event == event {
			n++
		}
	}
	return n
}

// WallMsByStage — total wall time by stage.
func (l *Ledger) WallMsByStage() map[string]int64 {
	out := make(map[string]int64)
	for _, e := range l.entries {
		out[e.Stage] += e.WallMs
	}
	return out
}

// P95WallMs — p95 of an event's record wall times by the
// nearest-rank rule. No records — zero.
func (l *Ledger) P95WallMs(event string) int64 {
	var xs []int64
	for _, e := range l.entries {
		if e.Event == event {
			xs = append(xs, e.WallMs)
		}
	}
	if len(xs) == 0 {
		return 0
	}
	sort.Slice(xs, func(i, j int) bool { return xs[i] < xs[j] })
	rank := (95*len(xs) + 99) / 100 // ceil(0.95·n), always ≥1 for n≥1
	return xs[rank-1]
}
