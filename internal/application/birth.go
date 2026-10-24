package application

// The birth: the instance begins as one atomic mkdir of its journal
// — exactly one winner among parallel runs, the losers see a living
// instance and run as ordinary passes. The winner fixes the goal (the
// wish's bytes as a content-addressed object, first record of the
// journal), consumes the wish slot, and writes the zone's standing
// files: the contract, the template, the empty corpus and the views.

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/neurophant/punchtape/internal/application/canon"
	"github.com/neurophant/punchtape/internal/domain"
	"github.com/neurophant/punchtape/internal/domain/doc"
	"github.com/neurophant/punchtape/internal/domain/journal"
	"github.com/neurophant/punchtape/internal/domain/zone"
	"github.com/neurophant/punchtape/internal/infrastructure/store"
	"github.com/neurophant/punchtape/internal/infrastructure/store/cas"
	"github.com/neurophant/punchtape/internal/ports"
)

// birth — the atomic birth of the instance in the zone.
func birth(zoneDir string) error {
	journalDir := filepath.Join(zoneDir, zone.JournalDir)
	// The mkdir IS the birth: EEXIST means a parallel run won it —
	// this run continues as an ordinary pass over a living instance.
	if err := ports.FSBind.Mkdir(journalDir); err != nil {
		if os.IsExist(err) {
			return nil
		}
		return err
	}
	wish, err := ports.FSBind.ReadFile(filepath.Join(zoneDir, zone.WishSlot))
	if err != nil {
		return err
	}
	objects := cas.New(filepath.Join(zoneDir, zone.ObjectsDir))
	addr, err := objects.Put(wish)
	if err != nil {
		return err
	}
	stream, err := store.OpenStream(journalDir)
	if err != nil {
		return err
	}
	if err := stream.Append(journal.Record{
		Kind: journal.KindBirth,
		Birth: &journal.Birth{
			Addr: addr, Digest: domain.Digest(wish),
			Version: Version,
		},
	}); err != nil {
		return err
	}
	// The wish slot dies at the birth: the goal is fixed forever.
	if err := ports.FSBind.Remove(filepath.Join(zoneDir, zone.WishSlot)); err != nil {
		return err
	}
	src := canon.Load()
	writes := []struct {
		name string
		data []byte
	}{
		{zone.ContractFile, []byte(src.Contract)},
		{zone.ContractRUFile, []byte(src.ContractRU)},
		{zone.ContractZHFile, []byte(src.ContractZH)},
		{zone.TemplateFile, []byte(src.Template)},
		{zone.CorpusFile, doc.Marshal(doc.Document{})},
		{zone.StateFile, initialViews.state},
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
