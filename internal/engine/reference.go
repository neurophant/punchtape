// The instance reference library: at open the machine places the
// topical deep-dive docs into .punchtape/reference/ - AGENTS.md
// carries the loop, the core laws and the map, these docs carry the
// detail of one topic each. Interface only, canon data, regenerated
// on canon change (hand edits are lost, like the template library).
package engine

import (
	"path/filepath"

	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/fsport"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// referenceDocs — the library layout: file name under
// .punchtape/reference/ and its canon key. The reference-selftest
// proves presence, headers, ASCII and that the AGENTS.md map names
// every doc.
var referenceDocs = []struct{ File, Key string }{
	{"environment.md", "ref.doc.environment"},
	{"grammar.md", "ref.doc.grammar"},
	{"model.md", "ref.doc.model"},
}

// referenceDocText — the file body: a one-line usage header over
// the canon body; the header is the file's identity anchor.
func referenceDocText(key string) string {
	return "# punchtape reference - machine-written from canon,\n" +
		"# regenerated on canon change; hand edits are lost.\n" +
		canondata.T(key) + "\n"
}

// ensureReference writes/maintains the reference library inside the
// instance. Idempotent: missing files are created, machine files are
// regenerated when canon changed, foreign files are left alone.
func ensureReference(workdir string) {
	dir := filepath.Join(workdir, ".punchtape", "reference")
	if err := fsport.Bind.MkdirAll(dir); err != nil {
		return
	}
	for _, d := range referenceDocs {
		path := filepath.Join(dir, d.File)
		want := referenceDocText(d.Key)
		data, err := fsport.Bind.ReadFile(path)
		if err != nil || string(data) != want {
			_ = yamlio.WriteAtomic(path, []byte(want))
		}
	}
}
