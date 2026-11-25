package application

// The AGENTS block: the machine's discovery hook — the one record it
// writes outside its zone. Between the markers stands the block's
// current content, ensured on every run as the file's first lines.
// Outside the markers not one byte belongs to the machine: the text
// around the block keeps its order and its words below the block —
// only its empty edge lines and its missing final newline go. The
// lines between the markers — and the one empty line the machine sets
// after them — are the machine's own in the coverage arithmetic as
// well: counted nowhere, never the agent's to exclude.

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/neurophant/punchtape/internal/ports"
)

const (
	agentsFile = "AGENTS.md"
	blockOpen  = "<!-- punchtape -->"
	blockClose = "<!-- /punchtape -->"
)

// agentsBlock — the block's content: the laws' short essence and the
// zone's reading addresses, one instruction on one line.
func agentsBlock() []byte {
	return []byte(blockOpen + "\n" +
		"punchtape owns this repo's work: run `punchtape` (no arguments) and read `.punchtape/state.md` for the standing and " +
		"the next step. The laws live in `.punchtape/contract.md`; the document's shape in `.punchtape/template.yaml`; " +
		"submissions go to `.punchtape/inbox.yaml`; the last run's facts are in `.punchtape/report.md`.\n" +
		blockClose + "\n")
}

// markers — the punchtape markers' standing in a file's bytes: how
// many times each is spelled and where its first spelling starts, -1
// when absent.
type markers struct {
	opens, closes int
	open, closeAt int
}

// findMarkers — the markers' standing in the file's bytes.
func findMarkers(data []byte) markers {
	return markers{
		opens:   bytes.Count(data, []byte(blockOpen)),
		closes:  bytes.Count(data, []byte(blockClose)),
		open:    bytes.Index(data, []byte(blockOpen)),
		closeAt: bytes.Index(data, []byte(blockClose)),
	}
}

// foreign — no marker at all: the file is the tree's own.
func (mk markers) foreign() bool { return mk.opens == 0 && mk.closes == 0 }

// whole — exactly one pair, the opening marker ahead of the closing
// one.
func (mk markers) whole() bool { return mk.opens == 1 && mk.closes == 1 && mk.open < mk.closeAt }

// ensureAgentsBlock — the block stands, current, whole and first. A
// file that cannot be read and a write that cannot land are loud
// failures; so is any marker count but none or one pair in order — a
// marker without its pair, the pair out of order, more than one pair
// is damage by hand: the machine does not guess and never keeps a
// broken hook in silence.
func (m *machine) ensureAgentsBlock() error {
	path := filepath.Join(m.root, agentsFile)
	data, err := ports.FSBind.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	mk := findMarkers(data)
	var rest []byte
	switch {
	case mk.foreign():
		// A foreign file (or none): its text stands below the block as
		// it was written.
		rest = closedText(data, false)
	case mk.whole():
		// The text around the block keeps its order — what stood above
		// it, then what stood below — behind one empty line under the
		// block; the empty lines at its edges go.
		below := bytes.TrimPrefix(data[mk.closeAt+len(blockClose):], []byte("\n"))
		rest = closedText(append(append([]byte(nil), data[:mk.open]...), below...), true)
	default:
		return fmt.Errorf("AGENTS.md carries damaged punchtape markers — a marker without its pair, the pair out of order, or more than one pair — the machine does not guess")
	}
	out := agentsBlock()
	if len(rest) > 0 {
		out = append(append(out, '\n'), rest...)
	}
	if bytes.Equal(out, data) {
		return nil
	}
	return ports.FSBind.WriteAtomic(path, out)
}

// machineHeld — the tree's lines the machine itself holds, as line
// ranges: the span of its AGENTS block in the working root's
// AGENTS.md — the opening marker's line through the closing marker's
// line, and the one empty line the machine sets after the block when
// it stands there — when exactly one pair stands in order. Whatever
// stands between the markers today, the run rewrites it at its end:
// the lines are the machine's, never a hole, never the agent's to
// exclude. No file and no whole pair hold nothing; a damaged pair
// refuses the run where it always did — where the block is ensured.
func machineHeld(root string) ([]string, error) {
	data, err := ports.FSBind.ReadFile(filepath.Join(root, agentsFile))
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	mk := findMarkers(data)
	if !mk.whole() {
		return nil, nil
	}
	first := 1 + bytes.Count(data[:mk.open], []byte{'\n'})
	last := 1 + bytes.Count(data[:mk.closeAt], []byte{'\n'})
	if bytes.HasPrefix(data[mk.closeAt+len(blockClose):], []byte("\n\n")) {
		last++
	}
	return []string{fmt.Sprintf("%s:%d-%d", agentsFile, first, last)}, nil
}

// closedText — a text as the file's tail wants it: ending in exactly
// one newline, its empty lines at the end dropped — and, when asked,
// the ones at its start too. No line left, no text.
func closedText(text []byte, startToo bool) []byte {
	lines := bytes.Split(text, []byte("\n"))
	for len(lines) > 0 && emptyLine(lines[len(lines)-1]) {
		lines = lines[:len(lines)-1]
	}
	if startToo {
		for len(lines) > 0 && emptyLine(lines[0]) {
			lines = lines[1:]
		}
	}
	if len(lines) == 0 {
		return nil
	}
	return append(bytes.Join(lines, []byte("\n")), '\n')
}

// emptyLine — a line carrying nothing but spaces and tabs.
func emptyLine(line []byte) bool {
	return len(bytes.Trim(line, " \t")) == 0
}
