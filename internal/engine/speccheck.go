// The dry validator behind `punchtape spec --check`: the door's
// document-level form rules as a deterministic
// function BEFORE the door. Every rejection class the door can teach
// used to
// cost a paid submit-rejected-reread cycle; the check names them all
// in one pass, each with its reason line, and a document that passes
// is taken by the door's form gate on the first try — the predicates
// are the door's own (docFindings, one law), never a second opinion.
package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/delta"
	"github.com/neurophant/punchtape/internal/fsport"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// CheckDocument — validate a draft document against the door's form
// rules without touching the door: the file is read from the work
// directory, the instance's surroundings are read as plain files (the
// stored build-ignore when the document declares none, the corpus's
// atom ids, the input and inbox directories), and every finding is
// returned as one "entity: reason" line. Nothing opens, nothing
// locks, nothing writes — the door's live facts (stage, idempotency)
// are the door's own.
func CheckDocument(workdir, source string) ([]string, error) {
	raw, err := fsport.Bind.ReadFile(filepath.Join(workdir, filepath.FromSlash(source)))
	if err != nil {
		return nil, fmt.Errorf("%s", canondata.T("cli.fail.reading", canondata.M{
			"source": source, "err": firstLine(err),
		}))
	}
	d, derr := delta.Decode(raw)
	if derr != nil {
		// a decode finding is a finding, not a system failure: the
		// draft's own bytes are the thing under check
		return []string{"delta: " + firstLine(derr)}, nil
	}
	if d.Patch != nil && d.FullReplace {
		return []string{"delta: " + canondata.T("submit.error.patch-with-full-replace")}, nil
	}
	lower := corpusLowerIDs(workdir)
	dc := docContext{
		workdir:     workdir,
		buildIgnore: d.BuildIgnore,
		corpusLower: lower,
		corpusCount: len(lower),
		statInput: func(rel string) (os.FileInfo, error) {
			return fsStat(filepath.Join(workdir, "..", "input", filepath.FromSlash(rel)))
		},
		statInbox: func(rel string) (os.FileInfo, error) {
			return fsStat(filepath.Join(workdir, ".punchtape", "inbox", filepath.FromSlash(rel)))
		},
	}
	if d.Patch != nil {
		// the patch draft checks by the SAME law the door applies: the
		// expansion over the read-only corpus, then the door's own
		// findings over the expanded document (one vocabulary, no
		// second law). The expansion is arithmetic over plain
		// reads: its only errors are the door's own rejections, and a
		// rejection Error() already prints "entity: reason"
		atoms, err := corpusAtomsRO(workdir)
		if err != nil {
			return nil, fmt.Errorf("%s", canondata.T("cli.fail", canondata.M{"err": firstLine(err)}))
		}
		expanded, _, err := expandPatchDoc(atoms, d)
		if err != nil {
			return []string{firstLine(err)}, nil
		}
		d = expanded
		dc.corpusLower = lower
		dc.corpusCount = len(atoms)
	}
	if d.BuildIgnore == nil {
		if st, err := readStoredState(workdir); err == nil {
			dc.buildIgnore = st.BuildIgnore
		}
	}
	findings := docFindings(d, dc)
	lines := make([]string, 0, len(findings))
	for _, f := range findings {
		lines = append(lines, f.entity+": "+f.reason)
	}
	return lines, nil
}

// readStoredState — the instance's state file as a plain read (the
// door's effective build-ignore when the document declares none);
// a missing or unreadable state reads as none.
func readStoredState(workdir string) (canon.State, error) {
	var st canon.State
	data, err := fsport.Bind.ReadFile(filepath.Join(workdir, ".punchtape", "state.yaml"))
	if err != nil {
		return st, err
	}
	if err := yamlio.DecodeDoc(data, &st); err != nil {
		return st, err
	}
	return st, nil
}

// corpusLowerIDs — the live corpus's ids by their lowercase spelling
// (the atoms zone decoded for the id lines only — the collision law
// needs the spellings, not the rows).
func corpusLowerIDs(workdir string) map[string]string {
	out := map[string]string{}
	entries, err := fsport.Bind.ReadDir(filepath.Join(workdir, ".punchtape", "canon", "atoms"))
	if err != nil {
		return out
	}
	for _, en := range entries {
		if en.IsDir() || !strings.HasSuffix(en.Name(), ".yaml") {
			continue
		}
		data, err := fsport.Bind.ReadFile(filepath.Join(workdir, ".punchtape", "canon", "atoms", en.Name()))
		if err != nil {
			continue
		}
		var a canon.Atom
		if yamlio.DecodeDoc(data, &a) != nil || a.ID == "" {
			continue
		}
		out[strings.ToLower(a.ID)] = a.ID
	}
	return out
}

// corpusAtomsRO — the live corpus's atoms as plain reads (the patch
// expansion of the dry validator: nothing opens, nothing locks — the
// door's live facts stay the door's own).
func corpusAtomsRO(workdir string) (map[string]canon.Atom, error) {
	out := map[string]canon.Atom{}
	entries, err := fsport.Bind.ReadDir(filepath.Join(workdir, ".punchtape", "canon", "atoms"))
	if err != nil {
		if os.IsNotExist(err) {
			return out, nil
		}
		return nil, err
	}
	for _, en := range entries {
		if en.IsDir() || !strings.HasSuffix(en.Name(), ".yaml") {
			continue
		}
		data, err := fsport.Bind.ReadFile(filepath.Join(workdir, ".punchtape", "canon", "atoms", en.Name()))
		if err != nil {
			return nil, err
		}
		var a canon.Atom
		if err := yamlio.DecodeDoc(data, &a); err != nil {
			return nil, err
		}
		if a.ID != "" {
			out[a.ID] = a
		}
	}
	return out, nil
}
