package application

// The AGENTS block: the machine's discovery hook — the one record it
// writes outside its zone. Between the markers stands the block's
// current content, ensured on every run: identical content touches
// nothing, changed content replaces itself, a foreign file gains the
// block at its end. Outside the markers not one byte belongs to the
// machine and not one byte of it is ever touched.

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
// zone's reading addresses.
func agentsBlock() []byte {
	var b bytes.Buffer
	fmt.Fprintf(&b, "%s\n", blockOpen)
	fmt.Fprintf(&b, "punchtape owns this repo's work: run `punchtape` (no\n")
	fmt.Fprintf(&b, "arguments) and read `.punchtape/state.md` for the standing and\n")
	fmt.Fprintf(&b, "the next step. The laws live in `.punchtape/contract.md`; the\n")
	fmt.Fprintf(&b, "document's shape in `.punchtape/template.yaml`; submissions go to\n")
	fmt.Fprintf(&b, "`.punchtape/inbox.yaml`; the last run's facts are in\n")
	fmt.Fprintf(&b, "`.punchtape/report.md`.\n")
	fmt.Fprintf(&b, "%s\n", blockClose)
	return b.Bytes()
}

// ensureAgentsBlock — the block stands, current and whole. A file
// that cannot be read, a write that cannot land, or damaged markers
// (one without its pair, or the pair out of order — damage by hand)
// are loud failures: the machine does not guess and never keeps a
// broken hook in silence.
func (m *machine) ensureAgentsBlock() error {
	path := filepath.Join(m.root, agentsFile)
	data, err := ports.FSBind.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	want := agentsBlock()
	open, closeAt := bytes.Index(data, []byte(blockOpen)), bytes.Index(data, []byte(blockClose))
	switch {
	case open < 0 && closeAt < 0:
		// A foreign file (or none): the block lands at the end.
		var out []byte
		out = append(out, data...)
		if len(out) > 0 && !bytes.HasSuffix(out, []byte("\n")) {
			out = append(out, '\n')
		}
		out = append(out, '\n')
		out = append(out, want...)
		return ports.FSBind.WriteAtomic(path, out)
	case open >= 0 && closeAt >= 0 && closeAt > open:
		start := open
		end := closeAt + len(blockClose) + 1 // the closing marker's newline
		if end > len(data) {
			end = len(data)
		}
		if bytes.Equal(data[start:end], want) {
			return nil
		}
		var out []byte
		out = append(out, data[:start]...)
		out = append(out, want...)
		out = append(out, data[end:]...)
		return ports.FSBind.WriteAtomic(path, out)
	default:
		// A lone marker, or the pair out of order, is damage by hand:
		// the machine does not guess — it refuses loudly, never
		// silently.
		return fmt.Errorf("AGENTS.md carries damaged punchtape markers — a marker without its pair, or the pair out of order — the machine does not guess")
	}
}
