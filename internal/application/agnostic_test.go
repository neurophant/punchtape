// The agnosticity gate: the machine carries no knowledge of any
// world outside its own contract — no tasks, stacks, domains, roles,
// cases, people, campaigns or experiments may live in its code, its
// canon, its configuration or its error lines; and no plan reference
// may live in its comments, docstrings or file names. The scan is
// mechanical and word-boundary aware (a witness word glued into an
// identifier is caught by its joint pattern); it reads the machine's
// own sources, so it fails the build the moment a violation lands.
package application

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// witnessWords — the vocabulary of the outside world: measurement
// campaigns and their parts, the roles they hire, the artifacts they
// exchange. None of these may appear in the machine.
var witnessWords = []string{
	// The measurement world and its artifacts.
	"campaign", "experiment", "canary", "ballot", "cage",
	// Case names of the outer world.
	"quire", "brim", "conduit", "errand", "pressmark",
	"brimdrift", "brimfeat", "brimfield", "brimstack",
	// Stacks the machine must never name.
	"python", "rust", "cargo", "npm", "webpack", "cmake",
	"clang", "django", "flask", "react", "pytest", "junit",
}

// planReferences — the plan's own vocabulary: stages, dashboards and
// section marks never belong in a comment, a docstring or a name.
var planReferences = []string{
	`§`, `щиток`, `стадия S`, `S0`, `S1`, `S2`, `S3`, `S4`, `S5`,
}

// machineSources — every source file of the machine's own trees.
func machineSources(t *testing.T) map[string]string {
	t.Helper()
	out := map[string]string{}
	// The test runs in its package directory: the repository root is
	// the nearest ancestor carrying go.mod.
	root := "."
	for depth := 0; depth < 8; depth++ {
		if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
			break
		}
		root = filepath.Join(root, "..")
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatal("go.mod not found upward")
	}
	roots := []string{filepath.Join(root, "cmd"), filepath.Join(root, "internal")}
	for _, root := range roots {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}
			// The machine's face is code AND the embedded data it
			// prints (the canon rides into the binary and out to
			// every agent) — both carry the words or neither does.
			if !strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, ".yaml") {
				return nil
			}
			// The scanner itself names the vocabulary it forbids —
			// the one exception AGENTS.md names for it.
			if strings.HasSuffix(path, "agnostic_test.go") {
				return nil
			}
			data, rerr := os.ReadFile(path)
			if rerr != nil {
				return rerr
			}
			out[path] = string(data)
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", root, err)
		}
	}
	return out
}

func TestNoWorldKnowledgeInSources(t *testing.T) {
	for path, src := range machineSources(t) {
		for _, word := range witnessWords {
			// Word boundaries: prose, comments and strings are where
			// world knowledge would live; the gate stays free of
			// false positives or it gets ignored.
			re := regexp.MustCompile(`(?i)\b` + word + `\b`)
			if loc := re.FindString(src); loc != "" {
				t.Fatalf("%s carries the outside word %q (matched %q) — the machine is agnostic",
					path, word, loc)
			}
		}
	}
}

func TestNoPlanReferencesInCommentsAndNames(t *testing.T) {
	for path, src := range machineSources(t) {
		for _, mark := range planReferences {
			if strings.Contains(src, mark) {
				t.Fatalf("%s carries the plan reference %q — comments, docstrings and names state intent only",
					path, mark)
			}
		}
		for _, seg := range strings.Split(filepath.ToSlash(path), "/") {
			for _, mark := range []string{"s0", "s1", "s2", "s3", "s4", "s5"} {
				if strings.HasPrefix(seg, mark+".") || strings.HasPrefix(seg, mark+"_") {
					t.Fatalf("the file name %s carries the stage reference %q", path, mark)
				}
			}
		}
	}
}
