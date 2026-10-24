package journal

// The record's bytes: the journal's own strict codec. Records are
// machine-written only — decode accepts exactly what marshal
// produces: strict fields, and a byte-exact decode→marshal round
// trip. A record that re-marshals differently is not a fact the
// machine wrote — a torn write or a hand edit, refused by name.

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/goccy/go-yaml"
)

// MarshalRecord — one record's canonical bytes.
func MarshalRecord(rec Record) []byte {
	// Records are plain scalar/slice/map data: marshal cannot fail.
	out, _ := yaml.MarshalWithOptions(rec, yaml.Indent(2))
	return out
}

// DecodeRecord — one record from its bytes, strict and round-trip
// exact.
func DecodeRecord(data []byte, rec *Record) error {
	if len(data) == 0 || data[len(data)-1] != '\n' {
		return fmt.Errorf("record cut mid-line")
	}
	if err := yaml.UnmarshalWithOptions(data, rec, yaml.Strict()); err != nil {
		return oneLine(err)
	}
	if !shaped(rec) {
		return fmt.Errorf("record kind %q carries no payload of its kind", rec.Kind)
	}
	if !bytes.Equal(MarshalRecord(*rec), data) {
		return fmt.Errorf("record does not round-trip byte-exact")
	}
	return nil
}

// shaped — exactly one payload matches the kind.
func shaped(rec *Record) bool {
	switch rec.Kind {
	case KindBirth:
		return rec.Birth != nil && rec.Submission == nil && rec.Run == nil && rec.Verdict == nil && rec.Refusals == nil && rec.Completed == nil
	case KindSubmission:
		return rec.Submission != nil && rec.Birth == nil && rec.Run == nil && rec.Verdict == nil && rec.Refusals == nil && rec.Completed == nil
	case KindRun:
		return rec.Run != nil && rec.Birth == nil && rec.Submission == nil && rec.Verdict == nil && rec.Refusals == nil && rec.Completed == nil
	case KindVerdict:
		return rec.Verdict != nil && rec.Birth == nil && rec.Submission == nil && rec.Run == nil && rec.Refusals == nil && rec.Completed == nil
	case KindRefusal:
		return rec.Refusals != nil && rec.Birth == nil && rec.Submission == nil && rec.Run == nil && rec.Verdict == nil && rec.Completed == nil
	case KindCompletion:
		return rec.Completed != nil && rec.Birth == nil && rec.Submission == nil && rec.Run == nil && rec.Verdict == nil && rec.Refusals == nil
	default:
		return false
	}
}

// oneLine — the machine's face wants one line per refusal. The
// caller passes a non-nil error.
func oneLine(err error) error {
	msg := strings.TrimSpace(err.Error())
	if i := strings.IndexByte(msg, '\n'); i >= 0 {
		msg = msg[:i]
	}
	return fmt.Errorf("%s", msg)
}
