package yamlio

import (
	"bytes"
	"fmt"
)

// The frame layer of the machine's append-only journals: every
// append writes a "---" separator line at column zero before its
// YAML document (the writer's invariant), so the stream splits into
// document frames without parsing. Marshaled documents never carry a
// column-zero "---" line of their own (block scalars are indented,
// flow scalars are single-line) — the split is exact.

// separator — one document frame's opening line.
var separator = []byte("---\n")

// SplitFrames splits a journal stream into its preamble (the lines
// before the first separator: the machine header comment) and the
// document frames, frame bytes without the separator line. A stream
// with no separator line is all preamble (zero records).
func SplitFrames(data []byte) (preamble []byte, frames [][]byte) {
	var pre [][]byte
	var cur []byte
	inFrame := false
	for len(data) > 0 {
		n := bytes.IndexByte(data, '\n')
		var line []byte
		if n < 0 {
			line, data = data, nil
		} else {
			line, data = data[:n+1], data[n+1:]
		}
		if isSeparator(line) {
			if inFrame {
				frames = append(frames, cur)
			} else {
				preamble = bytes.Join(pre, nil)
			}
			inFrame, cur = true, nil
			continue
		}
		if inFrame {
			cur = append(cur, line...)
		} else {
			pre = append(pre, line)
		}
	}
	if inFrame {
		frames = append(frames, cur)
	} else {
		preamble = bytes.Join(pre, nil)
	}
	return preamble, frames
}

func isSeparator(line []byte) bool {
	return bytes.Equal(line, separator) || bytes.Equal(line, []byte("---"))
}

// TornTail — a journal tail that does not decode as a complete
// record. The intact history always loads (the machine never goes
// blind over one torn append); the torn region is named here and is
// quarantined to a sidecar by the next write, never silently
// dropped.
type TornTail struct {
	Good   int // records before the torn region
	Offset int // byte offset where the torn region starts
	Length int // torn region length in bytes
}

func (t *TornTail) Error() string {
	return fmt.Sprintf("torn tail after %d records (%d bytes at offset %d)",
		t.Good, t.Length, t.Offset)
}

// FrameOffset — the byte offset of frame i in a stream with the
// given preamble: the separator line rides before every frame.
func FrameOffset(preamble []byte, frames [][]byte, i int) int {
	off := len(preamble)
	for j := 0; j < i; j++ {
		off += len(separator) + len(frames[j])
	}
	return off
}

// Rebuild — the byte image of a healthy journal: the preamble and
// the given frames verbatim, each behind its separator line.
func Rebuild(preamble []byte, frames [][]byte) []byte {
	var out bytes.Buffer
	out.Write(preamble)
	for _, fr := range frames {
		out.Write(separator)
		out.Write(fr)
	}
	return out.Bytes()
}
