package application

// The working tree as the machine sees it: one walk collects every
// fact the machine counts — each file under the declared universe
// with its line count and its bytes' digest, the tree's own
// fingerprint over them, and the count of everything standing
// outside the universe, named aloud. The zone and the
// version-control bookkeeping stay out — the tree's own machinery
// is never part of the verified surface. A verified walk cache
// keeps the unchanged files' bytes off the hot path; the cache
// never touches the proofs — claimed lines are read directly,
// always.

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/neurophant/punchtape/internal/domain"
	"github.com/neurophant/punchtape/internal/domain/fprint"
	"github.com/neurophant/punchtape/internal/domain/zone"
	"github.com/neurophant/punchtape/internal/ports"
)

// vcsMetadataDirs — the version-control bookkeeping families the
// contract names: the tree's machinery, never part of the verified
// surface.
var vcsMetadataDirs = map[string]bool{".git": true, ".hg": true, ".svn": true}

// treeEntry — the walk cache's one record: a counted file's own
// facts as the last walk that read its bytes verified them.
type treeEntry struct {
	size   int64
	mtime  int64
	lines  int
	digest string
}

// treeFacts — one walk's whole answer: the counted files with their
// line counts (the coverage universe's material), the tree's own
// fingerprint over every counted file's bytes, and the declared
// universe's inside and outside file counts.
type treeFacts struct {
	files   []fprint.FileLineCount
	digest  string
	inside  int
	outside int
}

// treeScope — one run's walk scope: the working root, the declared
// universe's roots (none — the whole tree counts) and the verified
// walk cache as it last stood.
type treeScope struct {
	root      string
	roots     []string
	cache     map[string]treeEntry
	lastFacts treeFacts
	walked    bool
}

// declared — whether the zone declares its universe: the outside
// line and the counts ride the declaration alone.
func (s *treeScope) declared() bool { return len(s.roots) > 0 }

// readUniverse — the zone's declared universe: one root-relative
// path a line, comments and blank lines passing — the same hygiene
// as the env pins. A line of any other shape is a loud refusal
// naming the file, the line and the law, never a silent loss. The
// line numbers ride along: a root that names no file of the tree
// is refused by its own place in the declaration.
func readUniverse(zoneDir string) ([]string, map[string]int, error) {
	data, err := ports.FSBind.ReadFile(filepath.Join(zoneDir, zone.UniverseSlot))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("%s: %w", zone.Path(zone.UniverseSlot), err)
	}
	var out []string
	lines := map[string]int{}
	for n, line := range strings.Split(string(data), "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if zone.ValidateRel(line) != nil {
			return nil, nil, fmt.Errorf(
				"universe refused — %s, line %d: %q — the law: one root-relative path a line",
				zone.Path(zone.UniverseSlot), n+1, line)
		}
		out = append(out, line)
		lines[line] = n + 1
	}
	return out, lines, nil
}

// carriesFiles — whether a declared root names at least one file the
// tree counts: the root itself when it is a file, else the first
// file the descent meets under it — the zone and the version-control
// bookkeeping never carry a root. A root that matches nothing is a
// declaration of an empty universe, refused loudly, never a silent
// loss of coverage.
func carriesFiles(root, rel string) (bool, error) {
	abs := filepath.Join(root, filepath.FromSlash(rel))
	info, err := ports.FSBind.Stat(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if !info.IsDir() {
		return true, nil
	}
	entries, err := ports.FSBind.ReadDir(abs)
	if err != nil {
		return false, err
	}
	for _, ent := range entries {
		name := ent.Name()
		if ent.IsDir() {
			if name == zone.Dirname || vcsMetadataDirs[name] {
				continue
			}
			carried, err := carriesFiles(abs, name)
			if err != nil || carried {
				return carried, err
			}
			continue
		}
		return true, nil
	}
	return false, nil
}

// under — whether a walked path stands inside the declared
// universe: a root spells the path itself or a directory above it.
func under(path string, roots []string) bool {
	for _, root := range roots {
		if path == root || strings.HasPrefix(path, root+"/") {
			return true
		}
	}
	return false
}

// newTreeScope — the run's scope over the zone: the declared
// universe — every root proven to carry at least one file of the
// tree — and the walk cache as it last stood on the disk.
func newTreeScope(root, zoneDir string) (*treeScope, error) {
	roots, lines, err := readUniverse(zoneDir)
	if err != nil {
		return nil, err
	}
	for _, rel := range roots {
		carried, err := carriesFiles(root, rel)
		if err != nil {
			return nil, err
		}
		if !carried {
			return nil, fmt.Errorf(
				"universe refused — %s, line %d: %q — the law: a root carries at least one file of the tree",
				zone.Path(zone.UniverseSlot), lines[rel], rel)
		}
	}
	s := &treeScope{root: root, roots: roots, cache: map[string]treeEntry{}}
	if err := s.loadCache(filepath.Join(zoneDir, zone.CacheTreeDir)); err != nil {
		return nil, err
	}
	return s, nil
}

// loadCache — the cache as it last stood: one record a line,
// "path\tsize\tmtime\tlines\tdigest". A line that does not parse is
// a miss for its path — the cache is the walk's own accelerator,
// rebuilt by the walk itself.
func (s *treeScope) loadCache(path string) error {
	data, err := ports.FSBind.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, line := range strings.Split(strings.TrimRight(string(data), "\n"), "\n") {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) != 5 {
			continue
		}
		size, errSize := strconv.ParseInt(parts[1], 10, 64)
		mtime, errMtime := strconv.ParseInt(parts[2], 10, 64)
		lines, errLines := strconv.Atoi(parts[3])
		if errSize != nil || errMtime != nil || errLines != nil || lines < 1 || parts[4] == "" {
			continue
		}
		s.cache[parts[0]] = treeEntry{size: size, mtime: mtime, lines: lines, digest: parts[4]}
	}
	return nil
}

// cacheBytes — the cache's whole content, one record a line, sorted
// by path: the deterministic bytes the zone carries between runs.
func (s *treeScope) cacheBytes() []byte {
	paths := make([]string, 0, len(s.cache))
	for p := range s.cache {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	var b bytes.Buffer
	for _, p := range paths {
		e := s.cache[p]
		fmt.Fprintf(&b, "%s\t%d\t%d\t%d\t%s\n", p, e.size, e.mtime, e.lines, e.digest)
	}
	return b.Bytes()
}

// walk — one pass over the working root: every file under the
// declared universe counts — its line count for the coverage
// universe, its bytes' digest for the tree's own fingerprint,
// binaries by their bytes — and everything outside counts by name
// alone, loudly, never read. A file whose size and mtime stand as
// the cache last saw them reuses the cached facts instead of its
// bytes; the proofs never ask this walk — claimed lines are read
// directly, always.
func (s *treeScope) walk() (treeFacts, error) {
	var facts treeFacts
	var digests []fprint.FileDigest
	entries := map[string]treeEntry{}
	err := ports.FSBind.Walk(s.root, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if p == s.root {
			return nil
		}
		// The walk yields paths under the root: the relative spelling
		// is the prefix cut — there is no error branch to armor.
		rel := strings.TrimPrefix(filepath.ToSlash(p), filepath.ToSlash(s.root)+"/")
		if info.IsDir() {
			if rel == zone.Dirname || vcsMetadataDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if len(s.roots) > 0 && !under(rel, s.roots) {
			facts.outside++
			return nil
		}
		facts.inside++
		if e, ok := s.cache[rel]; ok && e.size == info.Size() && e.mtime == info.ModTime().UnixNano() {
			entries[rel] = e
			facts.files = append(facts.files, fprint.FileLineCount{Path: rel, Lines: e.lines})
			digests = append(digests, fprint.FileDigest{Path: rel, Digest: e.digest})
			return nil
		}
		data, rerr := ports.FSBind.ReadFile(p)
		switch {
		case rerr != nil:
			facts.files = append(facts.files, fprint.FileLineCount{Path: rel, Outside: "unreadable bytes"})
			digests = append(digests, fprint.FileDigest{Path: rel, Digest: fprint.UnreadableDigest})
		case bytes.IndexByte(data, 0) >= 0:
			facts.files = append(facts.files, fprint.FileLineCount{Path: rel, Outside: "a NUL byte"})
			digests = append(digests, fprint.FileDigest{Path: rel, Digest: domain.Digest(data)})
		case fprint.CountLines(data) == 0:
			facts.files = append(facts.files, fprint.FileLineCount{Path: rel, Outside: "no lines"})
			digests = append(digests, fprint.FileDigest{Path: rel, Digest: domain.Digest(data)})
		default:
			e := treeEntry{
				size:   info.Size(),
				mtime:  info.ModTime().UnixNano(),
				lines:  fprint.CountLines(data),
				digest: domain.Digest(data),
			}
			entries[rel] = e
			facts.files = append(facts.files, fprint.FileLineCount{Path: rel, Lines: e.lines})
			digests = append(digests, fprint.FileDigest{Path: rel, Digest: e.digest})
		}
		return nil
	})
	if err != nil {
		return treeFacts{}, err
	}
	sort.Slice(facts.files, func(i, j int) bool { return facts.files[i].Path < facts.files[j].Path })
	facts.digest = fprint.TreeDigest(digests)
	// The walk's own view is the cache's new truth: vanished files
	// leave, fresh reads land, hits stand.
	s.cache = entries
	s.lastFacts = facts
	s.walked = true
	return facts, nil
}

// FileLines — one tree file's line count and existence, read from
// the working root: the claims' universe, always direct.
func (m *machine) FileLines(rel string) (int, bool) {
	data, err := ports.FSBind.ReadFile(filepath.Join(m.root, filepath.FromSlash(rel)))
	if err != nil {
		return 0, false
	}
	return fprint.CountLines(data), true
}

// Hole — whether a range names at least one line no post-merge
// behavior claims: the loud exclusion's own gate. A mark that marks
// nothing is refused.
func (m *machine) Hole(claims []string, rel string, first, last int) bool {
	claimed := map[int]bool{}
	for _, claim := range claims {
		cRel, cFirst, cLast, err := fprint.ParseLineRange(claim)
		if err != nil || cRel != rel {
			continue
		}
		for n := cFirst; n <= cLast; n++ {
			claimed[n] = true
		}
	}
	for n := first; n <= last; n++ {
		if !claimed[n] {
			return true
		}
	}
	return false
}
