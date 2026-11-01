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

// shaped — exactly one payload stands and it is the kind's own.
func shaped(rec *Record) bool {
	slots := []struct {
		kind    string
		present bool
	}{
		{KindBirth, rec.Birth != nil},
		{KindSubmission, rec.Submission != nil},
		{KindDemandAppend, rec.DemandAppend != nil},
		{KindNoNewDemands, rec.NoNewDemands != nil},
		{KindRun, rec.Run != nil},
		{KindVerdict, rec.Verdict != nil},
		{KindRefusal, rec.Refusals != nil},
		{KindCompletion, rec.Completed != nil},
		{KindExecution, rec.Execution != nil},
		{KindTreeLock, rec.TreeLock != nil},
	}
	own, filled := false, 0
	for _, s := range slots {
		if !s.present {
			continue
		}
		filled++
		if s.kind == rec.Kind {
			own = true
		}
	}
	return own && filled == 1
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
