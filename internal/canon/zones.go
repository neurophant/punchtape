// Instance zones: .punchtape has two zones with different write
// rights — the curator zone (the product passport: conventions and
// boundaries; future kb and diary kinds — the same zone) and the
// hidden service zone (state, the mutation journal, the ledger,
// locks, cache and snapshots), plus the canon zone — the spec IR,
// the machine protocol of the current work. The model is the single
// structural source: the writer (the effect whitelist, the
// reconciliation discipline, the integrity walk) asks it; separate
// path lists do not live in the code (zero mirrors inside the
// mechanics). Everything is written only by core verbs; the zone
// rules distinguish write mechanics, not the executor's right — the
// executor has no write right anywhere.
package canon

import (
	"strings"

	"github.com/neurophant/punchtape/internal/canondata"
)

// Instance zone identifiers.
const (
	// ZoneCanon — the spec IR: requirements, scenarios, checks,
	// cards, contracts, fixtures. Journaled+manifest: writer
	// transactions, reconciliation at open.
	ZoneCanon = "canon"
	// ZonePassport — the curator zone: conventions/environment and the
	// product's boundary registry; rendered to the human as why
	// topics. Journaled+manifest.
	ZonePassport = "passport"
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

// zones — the zone registry. The passport is declared before the
// service directories: parsing goes by first match, so the curator
// zone is not covered by a foreign rule.
var zones = []instanceZone{
	{ID: ZoneCanon, Write: WriteJournaled, Prefix: []string{
		"canon/requirements/", "canon/scenarios/", "canon/checks/",
		"canon/cards/", "canon/contracts/",
	}},
	{ID: ZonePassport, Write: WriteJournaled, Prefix: []string{
		"passport/",
	}},
	{ID: ZoneService, Write: WriteJournaled, Prefix: []string{
		"state.yaml",
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

// Collections — the canon collections from the zone model: the
// single list (the read walk and reconciliation keep no mirror of
// their own).
func Collections() []string {
	for _, z := range zones {
		if z.ID != ZoneCanon {
			continue
		}
		out := make([]string, 0, len(z.Prefix))
		for _, p := range z.Prefix {
			out = append(out, strings.TrimSuffix(p, "/"))
		}
		return out
	}
	return nil
}

// JournaledPath — whether a path is written by writer transactions
// under the manifest: the canon, passport and state. The journal and
// ledger are written by the writer itself, bypassing effects
// (append-only); the cache is rebuilt.
func JournaledPath(rel string) bool {
	rule, ok := ZoneWriteRule(rel)
	return ok && rule == WriteJournaled
}

// MachineHeader — the manifest header of an instance data file: it
// names the core verbs that write the file and the ban on manual
// edits. The first line of the file and part of a transaction
// effect's content — the journal and the manifest see the same text;
// the "file = journal" invariant does not diverge.
func MachineHeader(writers string) string {
	return canondata.T("store.file-header", canondata.M{"writers": writers})
}

// CacheHeader — the header of a rebuilt cache: a service value, not
// data; corruption and deletion are safe via recomputation.
func CacheHeader() string {
	return canondata.T("store.cache-header")
}
