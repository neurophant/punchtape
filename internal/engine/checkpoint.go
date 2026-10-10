// Checkpoints — digest snapshots of the work surface: the machine's
// only view of the project tree. The surface is the working root
// minus the .punchtape/ instance directory minus the spec's
// build-ignore patterns: file set + per-file sha256, one digest over
// the whole. Check rows run in the LIVE working tree (the spec's
// build once before them, their seeds self-removed after each row); a
// digest taken before the run and one after must name the same tree
// outside the ignores — a tree that moved during its own measurement
// is a machine-visible red line, and the verdict signs the pre-run
// digest: it names the exact tree the outcomes stand on. The machine
// reads the tree for digests and runs commands — it never interprets
// the project's structure.
package engine

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/neurophant/punchtape/internal/fsport"
	"github.com/neurophant/punchtape/internal/journal"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// checkpointFile — one accounted file of the surface: a slash-form
// path relative to the working root and its sha256 (a symlink — the
// recorded "link:<target>" fact).
type checkpointFile struct {
	Path   string `yaml:"path"`
	SHA256 string `yaml:"sha256"`
}

// Checkpoint — one digest snapshot of the repo surface. Written only
// by the machine inside gate/verdict/acceptance transactions; the seq
// is the instance's monotonic snapshot counter.
type Checkpoint struct {
	Seq    int              `yaml:"seq"`
	Digest string           `yaml:"digest"`
	Files  []checkpointFile `yaml:"files"`
	Tx     string           `yaml:"tx,omitempty"`
	At     time.Time        `yaml:"at"`
}

// checkpointFileName — a checkpoint record's effect path (from the
// project root): .punchtape/checkpoints/ckp-000001.yaml.
func checkpointFileName(seq int) string {
	return fmt.Sprintf(".punchtape/checkpoints/ckp-%06d.yaml", seq)
}

// ignoreMatcher — the passport's snapshot excludes, compiled per
// walk: a pattern ending in "/" excludes a directory prefix; a plain
// pattern matches the relative path or any file name on it (the
// gitignore habit for "*.log"-style patterns). The .punchtape
// instance directory is excluded unconditionally — the machine's zone
// is not the product's surface.
type ignoreMatcher struct {
	dirPrefix []string
	globs     []string
}

func newIgnoreMatcher(patterns []string) *ignoreMatcher {
	m := &ignoreMatcher{}
	for _, p := range patterns {
		p = strings.TrimSpace(filepath.ToSlash(p))
		if p == "" || p == "." || p == ".." || strings.Contains(p, "..") {
			continue
		}
		if strings.HasSuffix(p, "/") {
			m.dirPrefix = append(m.dirPrefix, strings.TrimSuffix(p, "/")+"/")
			continue
		}
		m.globs = append(m.globs, p)
	}
	return m
}

// excluded — whether a slash-form relative path is outside the
// accounted surface.
func (m *ignoreMatcher) excluded(rel string) bool {
	if rel == ".punchtape" || strings.HasPrefix(rel, ".punchtape/") {
		return true
	}
	for _, p := range m.dirPrefix {
		if strings.HasPrefix(rel, p) {
			return true
		}
	}
	base := path.Base(rel)
	for _, g := range m.globs {
		if ok, err := path.Match(g, rel); err == nil && ok {
			return true
		}
		if ok, err := path.Match(g, base); err == nil && ok {
			return true
		}
	}
	return false
}

// snapshotTree — the checkpoint of this moment: a deterministic walk
// of the surface (regular files hashed byte-exactly; symlinks are
// recorded by their target — a swapped link changes the digest), the
// digest over sorted "path\0hash" lines. Raw bytes, no newline
// normalization: the digest is the tree's exact identity.
func (e *Engine) snapshotTree() (*Checkpoint, error) {
	state, err := e.Store.State()
	if err != nil {
		return nil, err
	}
	return e.snapshotTreeIgnores(state.BuildIgnore)
}

// snapshotTreeIgnores — the snapshot under explicit ignore patterns: a
// run transaction snapshots with ITS OWN submission's declaration
// (the builder's state, still uncommitted), so a build-ignore that
// arrives with the repair covers that repair's own tree control.
func (e *Engine) snapshotTreeIgnores(buildIgnore []string) (*Checkpoint, error) {
	m := newIgnoreMatcher(buildIgnore)
	state, err := e.Store.State()
	if err != nil {
		return nil, err
	}
	var lines []string
	ckp := &Checkpoint{Seq: state.Counters["CKP"] + 1}
	err = fsport.Bind.Walk(e.Workdir, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// the walk is rooted at the working directory: the relative
		// form always exists by construction
		rel, _ := filepath.Rel(e.Workdir, p)
		rel = filepath.ToSlash(rel)
		if rel == "." {
			return nil
		}
		if m.excluded(rel) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if info.Mode()&os.ModeSymlink != 0 {
			tgt, lerr := fsport.Bind.Readlink(p)
			if lerr != nil {
				return lerr
			}
			ckp.Files = append(ckp.Files, checkpointFile{Path: rel, SHA256: "link:" + tgt})
			lines = append(lines, rel+"\x00link\x00"+tgt)
			return nil
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		data, rerr := fsport.Bind.ReadFile(p)
		if rerr != nil {
			return rerr
		}
		d := yamlio.Digest(data)
		ckp.Files = append(ckp.Files, checkpointFile{Path: rel, SHA256: d})
		lines = append(lines, rel+"\x00"+d)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(ckp.Files, func(i, j int) bool { return ckp.Files[i].Path < ckp.Files[j].Path })
	sort.Strings(lines)
	ckp.Digest = yamlio.Digest([]byte(strings.Join(lines, "\n")))
	ckp.At = timeNow().UTC()
	return ckp, nil
}

// checkpointEffect — the snapshot as a writer effect: the seq counter
// bumps in the same transaction. The snapshot is taken by the caller
// (the run substrate) — the record and the outcomes share one tree.
func (b *builder) checkpointEffect(ckp *Checkpoint, tx string) {
	ckp.Tx = tx
	// Checkpoint is plain data: marshal cannot fail
	data, _ := yamlio.Marshal(ckp)
	b.effects = append(b.effects, journal.Effect{
		Path: checkpointFileName(ckp.Seq), Content: machineFile(data),
	})
	if b.state.Counters == nil {
		b.state.Counters = map[string]int{}
	}
	b.state.Counters["CKP"] = ckp.Seq
}

// lastCheckpoint — the newest checkpoint record on disk (by seq);
// nil — the instance has no checkpoints yet.
func (e *Engine) lastCheckpoint() *Checkpoint {
	state, err := e.Store.State()
	if err != nil {
		return nil
	}
	for seq := state.Counters["CKP"]; seq > 0; seq-- {
		data, err := fsport.Bind.ReadFile(filepath.Join(e.Workdir,
			filepath.FromSlash(checkpointFileName(seq))))
		if err != nil {
			continue
		}
		var ckp Checkpoint
		if yamlio.DecodeDoc(data, &ckp) != nil {
			continue
		}
		return &ckp
	}
	return nil
}
