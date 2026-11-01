// The tree walk's battery: the zone and the VCS bookkeeping stay
// out, the walk is sorted; the line counts answer the claims'
// universe; the hole gate answers the exclusions' law.
package application

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/neurophant/punchtape/internal/domain/doc"
	"github.com/neurophant/punchtape/internal/domain/fprint"
	"github.com/neurophant/punchtape/internal/domain/zone"
	"github.com/neurophant/punchtape/internal/infrastructure/fsys"
	"github.com/neurophant/punchtape/internal/infrastructure/process"
	"github.com/neurophant/punchtape/internal/ports"
)

// scopeOf — the walk scope over a root without a zone: the whole
// tree counts, no cache stands.
func scopeOf(t *testing.T, root string) *treeScope {
	t.Helper()
	s, err := newTreeScope(root, filepath.Join(root, zone.Dirname))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestTreeLineCountsWalkTheCoverageUniverse(t *testing.T) {
	ports.FSBind = fsys.OS{}
	ports.ExecBind = process.OS{}
	root := t.TempDir()
	files := map[string]string{
		"src/a.go": "x\ny\n", "src/deep/b.go": "x\ny\nz\n", "doc/x.md": "one",
		"bin/data.bin": "a\x00b\n", "blank.txt": "",
		".punchtape/state.md": "zone\n", ".git/config": "vcs\n", "AGENTS.md": "hook\n",
	}
	for p, content := range files {
		path := filepath.Join(root, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	facts, err := scopeOf(t, root).walk()
	if err != nil {
		t.Fatal(err)
	}
	want := []fprint.FileLineCount{
		{Path: "AGENTS.md", Lines: 1},
		{Path: "bin/data.bin", Outside: "a NUL byte"},
		{Path: "blank.txt", Outside: "no lines"},
		{Path: "doc/x.md", Lines: 1},
		{Path: "src/a.go", Lines: 2},
		{Path: "src/deep/b.go", Lines: 3},
	}
	if len(facts.files) != len(want) {
		t.Fatalf("the walk owns the tree alone, sorted: %+v", facts.files)
	}
	for i, f := range want {
		if facts.files[i] != f {
			t.Fatalf("file %d must be %+v, got %+v", i, f, facts.files[i])
		}
	}
	// A second walk over the unchanged tree answers from the cache
	// alone — the same facts, byte for byte.
	again, err := scopeOf(t, root).walk()
	if err != nil {
		t.Fatal(err)
	}
	if again.digest != facts.digest || len(again.files) != len(facts.files) {
		t.Fatal("the cached walk must answer the same facts")
	}
}

func TestFileLinesCountsTheBytes(t *testing.T) {
	ports.FSBind = fsys.OS{}
	ports.ExecBind = process.OS{}
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "a.txt"), []byte("one\ntwo\nthree"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "empty.txt"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	m := &machine{root: root}
	if n, ok := m.FileLines("a.txt"); !ok || n != 3 {
		t.Fatalf("three lines without a trailing newline: %d %v", n, ok)
	}
	if n, ok := m.FileLines("empty.txt"); !ok || n != 0 {
		t.Fatalf("an empty file carries no lines: %d %v", n, ok)
	}
	if _, ok := m.FileLines("gone.txt"); ok {
		t.Fatal("an absent file answers no existence")
	}
}

func TestHoleNamesTheUnclaimedLines(t *testing.T) {
	ports.FSBind = fsys.OS{}
	ports.ExecBind = process.OS{}
	root := t.TempDir()
	m := &machine{root: root}
	claims := []string{"src/a.go:1-2"}
	if m.Hole(claims, "src/a.go", 1, 2) {
		t.Fatal("a fully claimed range names no hole")
	}
	if !m.Hole(claims, "src/a.go", 2, 3) {
		t.Fatal("line three stands unclaimed")
	}
	if !m.Hole(claims, "other.go", 1, 1) {
		t.Fatal("an unclaimed file is whole a hole")
	}
	if !m.Hole(nil, "src/a.go", 1, 2) {
		t.Fatal("with no claims every line is a hole")
	}
}

func TestTreeWalkFaultRefusesAndClaimsShapeSurvives(t *testing.T) {
	ports.FSBind = fsys.OS{}
	ports.ExecBind = process.OS{}
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "a.txt"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Join(root, "src"), 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(filepath.Join(root, "src"), 0o755) })
	if _, err := scopeOf(t, root).walk(); err == nil {
		t.Fatal("an unreadable directory must fail the walk")
	}
	// A malformed claim in a standing head answers zero shape, not a
	// crash — the shape laws keep such spellings out of the corpus.
	if files, lines := claimsShape(doc.Behavior{Claims: []string{"junk"}}); files != 0 || lines != 0 {
		t.Fatalf("a malformed claim shapes to nothing: %d %d", files, lines)
	}
}
