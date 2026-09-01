// Cache of instance file digests: mtime+size → digest. Instance files
// are re-read and re-hashed on every next (the freshness contour, the
// slot cache key) and every surface reconciliation — unchanged files
// reduce to a stat. The cache is not a source of truth: a corrupted or
// deleted cache file silently leads to recomputation; the truth is the
// data. Entry paths are relative to the project root (portability),
// and the digest is normalized across OSes.
package engine

import (
	"os"
	"path/filepath"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// digestEntry — a file digest, valid while the markers match.
type digestEntry struct {
	MTimeNs int64  `yaml:"mtime-ns"`
	Size    int64  `yaml:"size"`
	Digest  string `yaml:"digest"`
}

// digestCacheFile — cache files: slash paths from the project root.
// yaml.v3 Marshal sorts map keys — the file is deterministic.
type digestCacheFile struct {
	Files map[string]digestEntry `yaml:"files"`
}

// digestCache — the in-memory copy of the cache: entries and the dirty flag.
type digestCache struct {
	entries map[string]digestEntry
	dirty   bool
}

// digestOf — digest of an instance file (slash path from the project
// root): matching mtime and size reuses the entry, otherwise the file
// is re-read and the entry updated. ok=false — the file does not exist
// or is unreadable: disappearance is the caller's fact (missing), not
// the cache's.
func (e *Engine) digestOf(rel string) (string, bool) {
	c := e.theDigestCache()
	abs := filepath.Join(e.Workdir, filepath.FromSlash(rel))
	info, err := os.Stat(abs)
	if err != nil {
		delete(c.entries, rel)
		c.dirty = true
		return "", false
	}
	if en, ok := c.entries[rel]; ok && en.MTimeNs == info.ModTime().UnixNano() && en.Size == info.Size() {
		return en.Digest, true
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return "", false
	}
	d := yamlio.DigestNorm(data)
	c.entries[rel] = digestEntry{
		MTimeNs: info.ModTime().UnixNano(),
		Size:    info.Size(),
		Digest:  d,
	}
	c.dirty = true
	return d, true
}

// flushDigestCache — flush to disk once per verb (the logVerb point):
// entries about missing files are purged, growth is bounded by a data
// threshold — above the threshold the cache is rebuilt (recomputation,
// not an error).
func (e *Engine) flushDigestCache() {
	c := e.digestCache
	if c == nil || !c.dirty || len(c.entries) == 0 {
		return
	}
	if len(c.entries) > canondata.Limit("digest-cache-max") {
		c.entries = map[string]digestEntry{}
	}
	cf := digestCacheFile{Files: c.entries}
	data, err := yamlio.Marshal(cf)
	if err != nil {
		return
	}
	data = append([]byte(canon.CacheHeader()+"\n"), data...)
	if err := yamlio.WriteAtomic(filepath.Join(e.Store.Root(), "cache", "file-digests.yamll"), data); err != nil {
		return
	}
	c.dirty = false
}

// theDigestCache — lazy loading: the cache file is read strictly;
// corruption or absence — an empty cache (recomputation, not a
// failure: the cache is not truth).
func (e *Engine) theDigestCache() *digestCache {
	if e.digestCache != nil {
		return e.digestCache
	}
	c := &digestCache{entries: map[string]digestEntry{}}
	if data, err := os.ReadFile(filepath.Join(e.Store.Root(), "cache", "file-digests.yamll")); err == nil {
		var cf digestCacheFile
		if err := yamlio.DecodeStrict(data, &cf); err == nil && cf.Files != nil {
			c.entries = cf.Files
		}
	}
	e.digestCache = c
	return c
}
