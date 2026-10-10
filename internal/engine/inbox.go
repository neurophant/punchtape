// The inbox: the machine's single submission door — ONE file,
// .punchtape/inbox/current.txt, prefilled from the embedded template
// when missing. The executor overwrites the file and submits its path;
// any other path is refused with the wanted one named. After an
// accepted submission the machine rewrites the file back to the
// blank template: the next submission starts from a clean form.
// Byte-sample files of rows (bytes: file://name) and witness probes
// (witness: probe: file://name) lie next to the door file; they are
// the submitter's own staged bytes and stay as they are.
package engine

import (
	"path/filepath"

	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/fsport"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// inboxCurrent — the one legal submission path, relative to the work
// directory.
const inboxCurrent = ".punchtape/inbox/current.txt"

// inboxTemplate — the embedded skeleton the door is prefilled from.
const inboxTemplate = "spec.txt"

// ensureCurrentSet — the door invariant: the inbox directory exists
// and carries the template form when the file is missing. A file
// that exists is never touched — it is the executor's draft.
func (e *Engine) ensureCurrentSet() {
	dir := filepath.Join(e.Workdir, ".punchtape", "inbox")
	if err := fsport.Bind.MkdirAll(dir); err != nil {
		return
	}
	path := filepath.Join(dir, "current.txt")
	if _, err := fsport.Bind.Stat(path); err == nil {
		return
	}
	_ = yamlio.WriteAtomic(path, canondata.TemplateBytes(inboxTemplate))
}

// resetCurrentSet — the accepted submission consumed the draft: the
// machine rewrites the door to the blank template (the content lives
// on in the machine's state — the canon, the journal with digests).
func (e *Engine) resetCurrentSet() {
	_ = yamlio.WriteAtomic(filepath.Join(e.Workdir, filepath.FromSlash(inboxCurrent)),
		canondata.TemplateBytes(inboxTemplate))
}

// inboxStatusLine — the slot's pointer to the door: the machine
// names where to write.
func (e *Engine) inboxStatusLine() string {
	return canondata.T("slot.inbox.path", canondata.M{"path": inboxCurrent})
}

// SubmitSet — the door: the path must be the one submission file.
// The file is read and the document enters the submission pipeline;
// an accepted submission resets the door to the template, a
// rejection leaves the draft for the rework.
func (e *Engine) SubmitSet(source string) (string, int) {
	rel := filepath.ToSlash(filepath.Clean(source))
	if rel != inboxCurrent {
		reply := reject("door", canondata.T("submit.error.not-current", canondata.M{
			"given": rel, "want": inboxCurrent,
		}))
		e.logVerb("submit", "rejected", map[string]string{"reason": "not-the-door-file"}, 0, reply)
		return reply, Rejected
	}
	raw, err := fsport.Bind.ReadFile(filepath.Join(e.Workdir, rel))
	if err != nil {
		return canondata.T("cli.fail.reading", canondata.M{
			"source": rel, "err": firstLine(err),
		}), Failure
	}
	reply, code := e.submitCore(raw, true)
	if code == Rejected {
		// A validation refusal points back at the file it came from:
		// the diagnosis is the reply's own line, the pointer names
		// where the rework edits.
		reply += "\n" + canondata.T("submit.error.stays", canondata.M{"manifest": rel})
	}
	if code == Accepted && !e.lastIdempotent {
		e.resetCurrentSet()
	}
	return reply, code
}
