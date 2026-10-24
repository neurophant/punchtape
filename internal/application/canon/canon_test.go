// The canon's own battery: the source parses strictly, and the three
// contracts stand in structural parity — every law of the base and
// every zone file named by the base appears in each mirror. The
// check is structural on purpose: it proves the mirrors carry the
// whole law set, never that the words match — the meaning fidelity
// lives in the one source, reviewed as a whole; word-matching is not
// verification and is not attempted.
package canon

import (
	"sort"
	"strings"
	"testing"

	"github.com/neurophant/punchtape/internal/domain/doc"
)

var zoneFiles = []string{
	"contract.md", "contract-ru.md", "contract-zh.md", "template.yaml",
	"wish.md", "inbox.yaml", "corpus.yaml", "state.md", "report.md",
	"knowledge.md", "spec/", "journal/", "cache/", "tmp/",
}

func TestSourceParsesStrictly(t *testing.T) {
	s := Load()
	if s.Contract == "" || s.ContractRU == "" || s.ContractZH == "" ||
		s.Template == "" || s.Example == "" || len(s.LawMap) == 0 || len(s.Env) == 0 {
		t.Fatalf("the canon must carry every part: %+v", s)
	}
}

func TestContractStructuralParity(t *testing.T) {
	s := Load()
	for name, text := range map[string]string{
		"base": s.Contract, "ru": s.ContractRU, "zh": s.ContractZH,
	} {
		for _, law := range []string{
			"law:wish", "law:document", "law:paths", "law:done", "law:your-tree",
		} {
			if strings.Count(text, law) != 1 {
				t.Fatalf("%s mirror: the law %s must appear exactly once", name, law)
			}
		}
		for _, file := range zoneFiles {
			if !strings.Contains(text, file) {
				t.Fatalf("%s mirror: the zone file %s is missing", name, file)
			}
		}
		if !strings.Contains(text, "template.yaml") || !strings.Contains(text, "inbox.yaml") {
			t.Fatalf("%s mirror: the document's addresses are missing", name)
		}
	}
}

func TestLawMapCitesPrintedLawsOnly(t *testing.T) {
	s := Load()
	printed := map[string]bool{}
	for _, law := range []string{
		"law:wish", "law:document", "law:paths", "law:done", "law:your-tree",
	} {
		printed[law] = strings.Contains(s.Contract, law)
	}
	for _, ref := range s.LawMap {
		if !printed[ref.Law] {
			t.Fatalf("the map cites %q, which the contract does not print", ref.Law)
		}
		if ref.Rule == "" {
			t.Fatalf("the map carries an empty rule")
		}
	}
}

func TestRefusalLawsAreThePrintedOnes(t *testing.T) {
	contract := Load().Contract
	for _, law := range []string{doc.LawDocument, doc.LawPaths} {
		if !strings.Contains(contract, law) {
			t.Fatalf("the contract must print the refusal law %q", law)
		}
	}
}

// TestMirrorsCarryTheRegistry — the registry's field names and the
// W grammar are format, not prose: every mirror carries them
// verbatim, so a reader of any language meets one shape.
func TestMirrorsCarryTheRegistry(t *testing.T) {
	s := Load()
	marks := []string{
		"requirements:", "not-a-requirement:", "artifacts:", "covers:", "brief:",
		"quote:", "W1..Wm",
	}
	for name, mirror := range map[string]string{"base": s.Contract, "ru": s.ContractRU, "zh": s.ContractZH} {
		for _, mark := range marks {
			if !strings.Contains(mirror, mark) {
				t.Fatalf("the %s mirror lost the current page's %q", name, mark)
			}
		}
	}
	for _, base := range []string{s.Template, s.Example} {
		for _, mark := range marks[:5] {
			if !strings.Contains(base, mark) {
				t.Fatalf("the canon's own %q is missing from the template or the example", mark)
			}
		}
	}
}

func TestEnvIsSortedAndUnique(t *testing.T) {
	env := Load().Env
	sorted := append([]string(nil), env...)
	sort.Strings(sorted)
	for i, name := range env {
		if i > 0 && name == env[i-1] {
			t.Fatalf("the env rule repeats %q", name)
		}
	}
}

func TestBrokenSourcePanics(t *testing.T) {
	saved := source
	defer func() { source = saved }()
	source = []byte("contract: [")
	defer func() {
		if recover() == nil {
			t.Fatal("a broken source must panic the load")
		}
	}()
	Load()
}
