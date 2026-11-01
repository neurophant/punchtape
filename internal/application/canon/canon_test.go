// The canon's own battery: the source parses strictly, and the three
// contracts stand in structural parity — every law of the base and
// every zone file named by the base appears in each mirror. The
// check is structural on purpose: it proves the mirrors carry the
// whole law set, never that the words match — the meaning fidelity
// lives in the one source, reviewed as a whole; word-matching is not
// verification and is not attempted. The map battery pins the
// parity between the enforced rules and the laws the contract
// prints, and the token chase and the freshness cycle stay out of
// every teaching face.
package canon

import (
	"sort"
	"strings"
	"testing"

	"github.com/neurophant/punchtape/internal/domain/doc"
)

var zoneFiles = []string{
	"contract.md", "template.yaml",
	"demands.md", "inbox.yaml", "universe", "corpus.yaml", "state.md",
	"report.md", "knowledge.md", "spec/", "journal/", "holes.md",
	"trace.md", "exclusions-proposal.yaml",
}

func TestSourceParsesStrictly(t *testing.T) {
	s := Load()
	if s.Contract == "" || s.ContractRU == "" || s.ContractZH == "" ||
		s.Template == "" || s.Example == "" || len(s.LawMap) == 0 || len(s.Env) == 0 {
		t.Fatalf("the canon must carry every part: %+v", s)
	}
	if s.Reference == "" || s.ReferenceRU == "" || s.ReferenceZH == "" ||
		!strings.Contains(string(source), "reference:") {
		t.Fatal("the canon must carry the reference with its mirrors")
	}
}

func TestContractStructuralParity(t *testing.T) {
	s := Load()
	for name, text := range map[string]string{
		"base": s.Contract, "ru": s.ContractRU, "zh": s.ContractZH,
	} {
		for _, law := range []string{
			"law:document", "law:paths", "law:demand", "law:claims", "law:proof", "law:candor", "law:done", "law:your-tree",
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
	for _, law := range LawIDs() {
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

// TestLawMapParity — the map stands 1:1 with the validation's own
// rule set: the eleven rules of the document plus the shape laws
// that ride beside them, one row each — nothing dead survives, no
// enforced rule goes unmapped. The pin is the parity: the map and
// the gates drift together or the build fails.
func TestLawMapParity(t *testing.T) {
	want := []LawRef{
		{Rule: "the document decodes against the template's shape — unknown fields are refused", Law: "law:document"},
		{Rule: "the submission merges by id — requirements and behaviors by id, markers by their lines, exclusions by range, orders by demand and question; rows merge by run", Law: "law:document"},
		{Rule: "behavior ids are [a-z0-9-] segments joined by /, requirement ids are R then letters, digits and dots — each unique in a submission", Law: "law:document"},
		{Rule: "D ranges parse in the D grammar and name the stream's own numbered lines", Law: "law:document"},
		{Rule: "claims are line ranges the tree carries — path:first-last, root-relative", Law: "law:claims"},
		{Rule: "one range names one behavior — a claim already held is refused at its second holder", Law: "law:claims"},
		{Rule: "every new intent carries one of the six outcomes with non-empty prose and its carrier in the same document — answer its text, execute its run, work and describe a requirement's span, mark a marker, ask an awaiting order", Law: "law:demand"},
		{Rule: "every new stream line is declared by exactly one intent of the submission that meets it — a declared line is not redeclared", Law: "law:demand"},
		{Rule: "requirement prose is non-empty", Law: "law:demand"},
		{Rule: "refines names a requirement that stands", Law: "law:candor"},
		{Rule: "a marker never stands on a requirement's line — a line is either a demand or not", Law: "law:demand"},
		{Rule: "exclusions name real holes, a reason that cites a demand line or a law, references that resolve", Law: "law:candor"},
		{Rule: "behaviors carry a name, covers that stand after the merge, at least one row whose expect names rc, out or files, and runs unique in their rows", Law: "law:document"},
		{Rule: "every behavior carries at least one pinned row", Law: "law:proof"},
		{Rule: "one run proves one behavior — a run held by another behavior is refused as its proof", Law: "law:proof"},
		{Rule: "a mandate names the line that ordered the change and the standing thing it permits to replace", Law: "law:candor"},
		{Rule: "notes carry one of the five kinds, text, and references that resolve", Law: "law:candor"},
		{Rule: "a done document is exactly done: true", Law: "law:document"},
		{Rule: "the work narrative and every behavior's prose are non-empty", Law: "law:candor"},
		{Rule: "the build command is non-empty when present", Law: "law:document"},
		{Rule: "open orders ride standing demand lines with a real state", Law: "law:candor"},
		{Rule: "an awaiting order closes only over a later human line, cited by a decision note", Law: "law:candor"},
		{Rule: "remove keys of the five spellings name standing things", Law: "law:document"},
		{Rule: "a marker carries its reason", Law: "law:document"},
		{Rule: "expect file paths are root-relative", Law: "law:paths"},
		{Rule: "the tree does not move under answer, execute or describe intents — the digests hold the verdict", Law: "law:demand"},
		{Rule: "the universe names the counted roots; a line no claim and no loud exclusion names holds the verdict", Law: "law:claims"},
		{Rule: "the verdict needs the stream whole, claims mutual, pinned proofs green, no awaiting order — and waits out unresolved lines, unnamed drift, a moved tree", Law: "law:done"},
	}
	got := Load().LawMap
	if len(got) != len(want) {
		t.Fatalf("the map carries %d rules, the enforced set carries %d", len(got), len(want))
	}
	rows := map[string]string{}
	for _, row := range got {
		if _, dup := rows[row.Rule]; dup {
			t.Fatalf("the map carries the rule twice: %q", row.Rule)
		}
		rows[row.Rule] = row.Law
	}
	for _, w := range want {
		law, ok := rows[w.Rule]
		if !ok {
			t.Fatalf("the enforced rule is missing from the map: %q", w.Rule)
		}
		if law != w.Law {
			t.Fatalf("the rule %q cites %q, the gates cite %q", w.Rule, law, w.Law)
		}
	}
}

// TestContractMentionsNoTokens — the token chase and the freshness
// cycle never enter the canon's teaching: the contract, its mirrors
// and the template teach none of it.
func TestContractMentionsNoTokens(t *testing.T) {
	s := Load()
	for name, text := range map[string]string{
		"contract": s.Contract, "contract-ru": s.ContractRU,
		"contract-zh": s.ContractZH, "template": s.Template,
	} {
		for _, word := range []string{"token", "freshness"} {
			if strings.Contains(text, word) {
				t.Fatalf("the %s must not teach %q", name, word)
			}
		}
	}
}

// TestLawIDsFromCanon — the law ids parse from the contract's own
// headers: every id the machine resolves is a header the contract
// prints, and every header is an id — one source, no hardcode
// beside it.
func TestLawIDsFromCanon(t *testing.T) {
	contract := Load().Contract
	printed := LawIDs()
	want := []string{
		"law:demand", "law:document", "law:paths", "law:claims",
		"law:proof", "law:candor", "law:done", "law:your-tree",
	}
	if strings.Join(printed, ",") != strings.Join(want, ",") {
		t.Fatalf("the ids must be the contract's own headers in print order:\nwant %v\ngot  %v", want, printed)
	}
	headers := 0
	for _, line := range strings.Split(contract, "\n") {
		if strings.HasPrefix(line, "law:") {
			headers++
		}
	}
	if headers != len(printed) {
		t.Fatalf("every law header is an id: %d headers, %d ids", headers, len(printed))
	}
	for _, law := range printed {
		if strings.Count(contract, law) != 1 {
			t.Fatalf("the id %q must be printed exactly once", law)
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
// D grammar are format, not prose: the template and the example —
// the one shape the zone carries — teach them verbatim, so a reader
// meets one shape.
func TestMirrorsCarryTheRegistry(t *testing.T) {
	s := Load()
	for _, base := range []string{s.Template, s.Example} {
		for _, mark := range []string{"requirements:", "not-a-requirement", "covers:", "claims:", "intents:"} {
			if !strings.Contains(base, mark) {
				t.Fatalf("the canon's own %q is missing from the template or the example", mark)
			}
		}
	}
}

// TestTeachingLimits — the teaching faces stay lean: the contract is
// one page, the template and the example are short — bloat is
// measured, not argued, and a canon that grows must say so here.
func TestTeachingLimits(t *testing.T) {
	s := Load()
	for _, face := range []struct {
		name  string
		text  string
		limit int
	}{
		{"contract", s.Contract, 95},
		{"template", s.Template, 45},
		{"reference", s.Reference, 80},
		{"example", s.Example, 30},
	} {
		if lines := strings.Count(strings.TrimSuffix(face.text, "\n"), "\n") + 1; lines > face.limit {
			t.Fatalf("the %s carries %d lines, the limit is %d", face.name, lines, face.limit)
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
