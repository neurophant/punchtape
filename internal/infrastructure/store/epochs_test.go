// The epoch naming law's battery: one spelling — the formatter, the
// pattern, the number order, and the refusal of any other suffix.
package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/neurophant/punchtape/internal/infrastructure/fsys"
	"github.com/neurophant/punchtape/internal/ports"
)

func TestEpochNameOneSpelling(t *testing.T) {
	ports.FSBind = fsys.OS{}
	if got := EpochFileName(1); got != "epoch-00001.yaml" {
		t.Fatalf("the formatter spells one suffix: %q", got)
	}
	if !EpochNamePattern.MatchString("epoch-00042.yaml") {
		t.Fatal("the pattern must match the one spelling")
	}
	for _, name := range []string{"epoch-00042.yamll", "epoch-00042.yml", "epoch-00042"} {
		if EpochNamePattern.MatchString(name) {
			t.Fatalf("the pattern must refuse every other spelling, matched %q", name)
		}
	}
}

func TestEpochsOrderByNumber(t *testing.T) {
	ports.FSBind = fsys.OS{}
	dir := t.TempDir()
	// Written out of number order and padded with foreign names: the
	// listing answers by the number alone.
	for _, name := range []string{"epoch-00003.yaml", "epoch-00001.yaml", "notes.txt", "epoch-2.yaml"} {
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	paths, err := ListEpochs(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"epoch-00001.yaml", "epoch-00003.yaml"}
	if len(paths) != len(want) {
		t.Fatalf("the foreign names stay out, got %v", paths)
	}
	for i, p := range paths {
		if filepath.Base(p) != want[i] {
			t.Fatalf("the order is the number's: %v", paths)
		}
	}
	if got := NextEpochPath(paths, dir); filepath.Base(got) != "epoch-00004.yaml" {
		t.Fatalf("the next epoch follows the last number: %q", got)
	}
}
