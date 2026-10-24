// The uncovered tree's split battery: the artifact globs divide the
// uncovered files into the declared outputs and the uncovered
// sources — a partition, never an overlap — and the state renders
// both headers.
package application

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neurophant/punchtape/internal/domain/zone"
)

func TestSplitUncoveredPartitionsTheTree(t *testing.T) {
	tree := []string{"AGENTS.md", "README.md", "dist/app.tgz", "out/built.txt", "src/a.go"}
	scopes := []string{"src/**"}
	// No artifacts: every uncovered file lands in the sources, the
	// machine's own record file in neither list.
	declared, sources := SplitUncovered(tree, scopes, nil)
	if len(declared) != 0 || strings.Join(sources, ",") != "README.md,dist/app.tgz,out/built.txt" {
		t.Fatalf("empty artifacts put the uncovered tree whole into sources: %v | %v", declared, sources)
	}
	// Declared artifacts take their files out of the sources; the
	// same matcher the scopes use decides.
	declared, sources = SplitUncovered(tree, scopes, []string{"out/**", "dist/*.tgz"})
	if strings.Join(declared, ",") != "dist/app.tgz,out/built.txt" || strings.Join(sources, ",") != "README.md" {
		t.Fatalf("the artifact globs divide the uncovered tree: %v | %v", declared, sources)
	}
	for _, d := range declared {
		for _, s := range sources {
			if d == s {
				t.Fatalf("a file stands under both lists: %s", d)
			}
		}
	}
}

func TestStateRendersBothUncoveredHeaders(t *testing.T) {
	root := newBornRoot(t)
	// A stray source and the row's output: declaring the output
	// splits the uncovered tree into the two named lists.
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("words"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, "artifacts: [out.txt]\n")
	runIn(t, root)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "artifacts (declared):\n- out.txt") {
		t.Fatalf("the declared output stands under its header:\n%s", state)
	}
	if !strings.Contains(state, "uncovered sources:\n- README.md") {
		t.Fatalf("the uncovered source stands under its header:\n%s", state)
	}
	if strings.Contains(state, "- AGENTS.md") {
		t.Fatalf("the machine's own record file is no fact of the project:\n%s", state)
	}
}
