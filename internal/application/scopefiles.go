package application

// The working tree as the fingerprints see it: one walk collects the
// files, the scope globs of the corpus's behaviors divide them, the
// digests come through a (size, mtime) cache that is an accelerator
// and never a source of truth — delete it and the next run recomputes
// the same numbers. Files outside every scope are named: the map's
// gaps are visible facts.

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/neurophant/punchtape/internal/domain"
	"github.com/neurophant/punchtape/internal/domain/fprint"
	"github.com/neurophant/punchtape/internal/domain/zone"
	"github.com/neurophant/punchtape/internal/ports"
)

// vcsMetadataDirs — the version-control bookkeeping families the
// contract's SCOPE section names: the tree's machinery, never part
// of the verified surface.
var vcsMetadataDirs = map[string]bool{".git": true, ".hg": true, ".svn": true}

// TreeFiles — the working root's files, sorted, root-relative: the
// zone and the VCS metadata stay out.
func TreeFiles(root string) ([]string, error) {
	var out []string
	err := ports.FSBind.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if p == root {
			return nil
		}
		// The walk yields paths under the root: the relative spelling
		// is the prefix cut — there is no error branch to armor.
		rel := strings.TrimPrefix(filepath.ToSlash(p), filepath.ToSlash(root)+"/")
		if info.IsDir() {
			if rel == zone.Dirname || vcsMetadataDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		out = append(out, rel)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(out)
	return out, nil
}

// ScopeDigests — one behavior's scope files with their measured
// digests: the files of the tree its globs cover, a missing file
// spelled absent.
func ScopeDigests(root string, tree []string, globs []string, cache *HashCache) []fprint.FileDigest {
	want := map[string]bool{}
	for _, rel := range tree {
		for _, g := range globs {
			if fprint.MatchGlob(g, rel) {
				want[rel] = true
				break
			}
		}
	}
	out := make([]fprint.FileDigest, 0, len(want))
	for rel := range want {
		out = append(out, fprint.FileDigest{Path: rel, Digest: cache.Digest(root, rel)})
	}
	return out
}

// Uncovered — the tree's files no scope glob of the corpus covers.
func Uncovered(tree []string, allGlobs []string) []string {
	var out []string
	for _, rel := range tree {
		if !matchesAny(allGlobs, rel) {
			out = append(out, rel)
		}
	}
	return out
}

// SplitUncovered — the uncovered files divided by the standing
// document's declared artifacts: the files an artifact glob names,
// and the rest — the sources no scope covers and no artifact
// declares. The same matcher the scopes use decides both; the two
// lists are a partition of the uncovered tree, and the machine's own
// record file is none of the agent's business in either.
func SplitUncovered(tree []string, allGlobs, artifacts []string) (declared, sources []string) {
	for _, rel := range Uncovered(tree, allGlobs) {
		if rel == agentsFile {
			continue
		}
		if matchesAny(artifacts, rel) {
			declared = append(declared, rel)
			continue
		}
		sources = append(sources, rel)
	}
	return declared, sources
}

// matchesAny — whether any glob names the path.
func matchesAny(globs []string, rel string) bool {
	for _, g := range globs {
		if fprint.MatchGlob(g, rel) {
			return true
		}
	}
	return false
}

// HashCache — the (path, size, mtime) → digest accelerator over the
// zone's cache file. A touched file recomputes honestly; a corrupt or
// missing cache answers empty and the run recomputes everything. The
// digest is the sha256 of the file's own bytes, no transformation
// ever — entries of a foreign digest law (no algo tag) drop out and
// recompute rather than lie.
type HashCache struct {
	zoneDir string
	entries map[string]cacheEntry
	loaded  bool
	dirty   bool
}

// digestAlgo — the digest law the cache speaks; raised when the
// formula changes so stale entries of the old law cannot survive.
const digestAlgo = 2

type cacheEntry struct {
	Algo   int    `yaml:"algo"`
	Size   int64  `yaml:"size"`
	Mtime  int64  `yaml:"mtime"`
	Digest string `yaml:"digest"`
}

// OpenHashCache — the cache over a zone directory.
func OpenHashCache(zoneDir string) *HashCache {
	return &HashCache{zoneDir: zoneDir, entries: map[string]cacheEntry{}}
}

// Load — read the cache file once; unreadable is empty, not an
// error; entries of another digest law are not the machine's truth
// and drop out for an honest recompute.
func (c *HashCache) Load() {
	if c.loaded {
		return
	}
	c.loaded = true
	data, err := ports.FSBind.ReadFile(filepath.Join(c.zoneDir, zone.CacheDir, "hashes.yaml"))
	if err != nil {
		return
	}
	var file struct {
		Files map[string]cacheEntry `yaml:"files"`
	}
	if yaml.UnmarshalWithOptions(data, &file, yaml.Strict()) != nil {
		return
	}
	for rel, e := range file.Files {
		if e.Algo == digestAlgo {
			c.entries[rel] = e
		}
	}
}

// Digest — one file's digest: cached when the law, size and mtime
// stand; the sha256 of the file's own bytes otherwise.
func (c *HashCache) Digest(root, rel string) string {
	c.Load()
	info, err := ports.FSBind.Stat(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		return fprint.AbsentDigest
	}
	if e, ok := c.entries[rel]; ok && e.Size == info.Size() && e.Mtime == info.ModTime().UnixNano() {
		return e.Digest
	}
	data, err := ports.FSBind.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		return fprint.AbsentDigest
	}
	d := domain.Digest(data)
	c.entries[rel] = cacheEntry{Algo: digestAlgo, Size: info.Size(), Mtime: info.ModTime().UnixNano(), Digest: d}
	c.dirty = true
	return d
}

// Flush — write the cache once per run; nothing to write is no file.
func (c *HashCache) Flush() {
	if !c.dirty {
		return
	}
	var file struct {
		Files map[string]cacheEntry `yaml:"files"`
	}
	file.Files = c.entries
	// The shape is plain scalar/map data: marshal cannot fail.
	data, _ := yaml.MarshalWithOptions(file, yaml.Indent(2))
	_ = ports.FSBind.WriteAtomic(filepath.Join(c.zoneDir, zone.CacheDir, "hashes.yaml"), data)
}
