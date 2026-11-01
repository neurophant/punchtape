package application

// The birth: the instance begins as one atomic mkdir of its journal
// — exactly one winner among parallel runs, the losers see a living
// instance and run as ordinary passes. The winner journals the birth
// (the creating machine's version) and the chain's genesis link over
// the stream's initial lines, then writes the zone's standing files.
// The stream file itself survives the birth, open forever: the
// human's words keep entering it at any moment of the instance's
// life.

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/neurophant/punchtape/internal/application/canon"
	"github.com/neurophant/punchtape/internal/domain/doc"
	"github.com/neurophant/punchtape/internal/domain/journal"
	"github.com/neurophant/punchtape/internal/domain/zone"
	"github.com/neurophant/punchtape/internal/infrastructure/store"
	"github.com/neurophant/punchtape/internal/ports"
)

// birth — the atomic birth of the instance in the zone. The caller
// has read the stream and seen its first non-empty line; the bytes
// ride along, read once.
func birth(zoneDir string, words []byte) error {
	journalDir := filepath.Join(zoneDir, zone.JournalDir)
	// The mkdir IS the birth: EEXIST means a parallel run won it —
	// this run continues as an ordinary pass over a living instance.
	if err := ports.FSBind.Mkdir(journalDir); err != nil {
		if os.IsExist(err) {
			return nil
		}
		return err
	}
	// The chain's genesis link covers all the stream's initial lines.
	lines := doc.NonEmptyLines(string(words))
	records, err := store.OpenStream(journalDir)
	if err != nil {
		return err
	}
	if err := records.Append(journal.Record{
		Kind:  journal.KindBirth,
		Birth: &journal.Birth{Version: Version},
	}); err != nil {
		return err
	}
	genesis := journal.Record{
		Kind:         journal.KindDemandAppend,
		DemandAppend: &journal.DemandAppend{Lines: len(lines), Chain: chainOver("", lines)},
	}
	if err := records.Append(genesis); err != nil {
		return err
	}
	src := canon.Load()
	state := fmt.Appendf(nil, "# punchtape state\n\ncorpus: empty — submit the first document\nnext: .punchtape/inbox.yaml\n")
	writes := []struct {
		name string
		data []byte
	}{
		{zone.ContractFile, []byte(src.Contract)},
		{zone.TemplateFile, []byte(src.Template)},
		{zone.ReferenceFile, []byte(src.Reference)},
		{zone.CorpusFile, doc.Marshal(doc.Document{})},
		{zone.StateFile, state},
		{zone.ReportFile, initialViews.report},
		{zone.KnowledgeFile, initialViews.knowledge},
		{filepath.Join(zone.SpecDir, "index.md"), initialViews.index},
	}
	for _, w := range writes {
		if err := ports.FSBind.WriteAtomic(filepath.Join(zoneDir, w.name), w.data); err != nil {
			return fmt.Errorf("birth write %s: %w", w.name, err)
		}
	}
	return nil
}
