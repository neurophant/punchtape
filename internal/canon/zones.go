// Instance zones: .punchtape has two zones with different write
// rights — the canon zone (the corpus: one collection, the atoms)
// and the hidden service zone (state, the mutation journal, the
// ledger, locks, cache and snapshots). The model is the single
// structural source: the writer (the effect whitelist, the
// reconciliation discipline, the integrity walk) asks it; separate
// path lists do not live in the code (zero mirrors inside the
// mechanics). Everything is written only by core verbs; the zone
// rules distinguish write mechanics, not the executor's right —
// the executor has no write right anywhere.
package canon

import (
	"strings"
)

// Instance zone identifiers.
const (
	// ZoneCanon — the corpus: the behavior atoms. Journaled+manifest:
	// writer transactions, reconciliation at open.
	ZoneCanon = "canon"
	// ZoneService — the hidden service zone: state, the mutation
	// journal, the ledger, locks, cache and snapshots. state —
	// journaled+manifest; the journal and ledger are append-only; the
	// cache is rebuilt, not a source of truth, and does not pass the
	// integrity reconciliation.
	ZoneService = "service"
)

// Write rules for an instance data zone.
const (
	// WriteJournaled — written by writer transactions and guarded by
	// the manifest (manual edits are detected at open).
	WriteJournaled = "journaled"
	// WriteAppend — a stream of records, appended to only at the end.
	WriteAppend = "append-only"
	// WriteRebuilt — a rebuildable service value: not a source of
	// truth, safe to delete.
	WriteRebuilt = "rebuilt"
)

// instanceZone — one zone: an identifier, path prefixes from the
// .punchtape root and a write rule. The prefixes within a zone go
// from longest to shortest (an exact file before a directory).
type instanceZone struct {
	ID     string
	Prefix []string
	Write  string
}

// zones — the zone registry.
var zones = []instanceZone{
	{ID: ZoneCanon, Write: WriteJournaled, Prefix: []string{
		"canon/atoms/",
	}},
	{ID: ZoneService, Write: WriteJournaled, Prefix: []string{
		"state.yaml", "checkpoints/", "verdicts/",
	}},
	{ID: ZoneService, Write: WriteAppend, Prefix: []string{
		"journal.yamll", "ledger.yamll",
	}},
	{ID: ZoneService, Write: WriteRebuilt, Prefix: []string{
		"cache/",
	}},
}

// ZoneOf — the zone of a slash-form path inside .punchtape (with a
// trailing directory prefix); false — the path lies outside the
// marked zones.
func ZoneOf(rel string) (string, bool) {
	for _, z := range zones {
		for _, p := range z.Prefix {
			if rel == p || len(rel) > len(p) && rel[:len(p)] == p {
				return z.ID, true
			}
		}
	}
	return "", false
}

// ZoneWriteRule — the write rule of a path's zone; false — outside
// the zones.
func ZoneWriteRule(rel string) (string, bool) {
	for _, z := range zones {
		for _, p := range z.Prefix {
			if rel == p || len(rel) > len(p) && rel[:len(p)] == p {
				return z.Write, true
			}
		}
	}
	return "", false
}

// zoneDirs — a zone's directory prefixes (the trailing-slash form)
// as paths; nil — no such zone in the model.
func zoneDirs(id string) []string {
	for _, z := range zones {
		if z.ID == id {
			out := make([]string, 0, len(z.Prefix))
			for _, p := range z.Prefix {
				out = append(out, strings.TrimSuffix(p, "/"))
			}
			return out
		}
	}
	return nil
}

// Collections — the canon collections from the zone model: the
// single list (the read walk and reconciliation keep no mirror of
// their own). Coupled consumers of this derivation: the writer's
// reconciliation scan walks exactly these directories, and Init's
// skeleton mkdir list (store.go) mirrors them by hand — a zone-model
// edit that renames or adds a canon prefix must carry the skeleton
// with it; TestZoneModel pins the derived list so the drift is
// caught.
func Collections() []string { return zoneDirs(ZoneCanon) }

// JournaledPath — whether a path is written by writer transactions
// under the manifest: the canon and state. The journal and ledger
// are written by the writer itself, bypassing effects (append-only);
// the cache is rebuilt.
func JournaledPath(rel string) bool {
	rule, ok := ZoneWriteRule(rel)
	return ok && rule == WriteJournaled
}
