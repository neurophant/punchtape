// The refusal tails of the format: every forbidden node kind of the
// deep armor and every empty field of the shape laws answers its own
// named refusal.
package doc

import (
	"strings"
	"testing"
)

func TestArmorRefusesForbiddenNodes(t *testing.T) {
	for _, tc := range []struct{ name, in string }{
		{"an alias", "a: *x\n"},
		{"a tag", "a: !!str 5\n"},
		{"a merge key", "<<: {b: 1}\n"},
		{"an anchored sequence item", "rows:\n    - &x 1\n"},
	} {
		if _, err := Decode([]byte(tc.in)); err == nil {
			t.Errorf("%s must be refused", tc.name)
		}
	}
}

func TestValidateEmptyFieldsAndRemoveCharset(t *testing.T) {
	d := Document{
		Behaviors: []Behavior{{Name: "n", Rows: []Row{{Run: Cmd{Line: "x"}, Expect: Expect{Out: "1"}}}}},
		Remove:    []string{"UP"},
	}
	r := Validate(d, nil)
	joined := ""
	for _, ref := range r {
		joined += ref.Error() + "\n"
	}
	for _, want := range []string{
		"behaviors[0].id", "the id is empty",
		"remove[0]", "the id must be [a-z0-9-] segments joined by '/'",
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing %q in:\n%s", want, joined)
		}
	}

	nameless := Document{Behaviors: []Behavior{{ID: "a/b", Rows: []Row{{Run: Cmd{Line: "x"}, Expect: Expect{Out: "1"}}}}}}
	r = Validate(nameless, nil)
	if len(r) != 1 || !strings.Contains(r[0].Error(), "behaviors[0].name") || !strings.Contains(r[0].Error(), "the name is empty") {
		t.Fatalf("the empty name must be refused by its own field: %v", r)
	}

	colon := Document{Behaviors: []Behavior{{ID: "a", Name: "n", Scope: []string{"a:b"}, Rows: []Row{{Run: Cmd{Line: "x"}, Expect: Expect{Out: "1"}}}}}}
	r = Validate(colon, nil)
	if len(r) != 1 || !strings.Contains(r[0].Error(), "law:paths") {
		t.Fatalf("a glob outside the alphabet must cite the paths law: %v", r)
	}
}

func TestDuplicateIdInOneSubmission(t *testing.T) {
	d := Document{Behaviors: []Behavior{
		{ID: "a/b", Name: "n", Rows: []Row{{Run: Cmd{Line: "x"}, Expect: Expect{Out: "1"}}}},
		{ID: "a/b", Name: "m", Rows: []Row{{Run: Cmd{Line: "y"}, Expect: Expect{Out: "2"}}}},
	}}
	r := Validate(d, nil)
	if len(r) != 1 || !strings.Contains(r[0].Error(), "appears twice") {
		t.Fatalf("a duplicate id must be refused once by its own field: %v", r)
	}
}

// The first-submission battery: a bad indent in free text is
// refused WITH its source position, and the template's multi-line
// notes form decodes — the guess that cost both field hands one
// refusal each has nothing left to guess.
func TestRefusalCarriesSourcePosition(t *testing.T) {
	_, err := Decode([]byte("behaviors:\n  - id: b1\n    name: n\n    notes: >\n   folded wrong indent\n"))
	if err == nil {
		t.Fatal("the broken document must be refused")
	}
	if !strings.Contains(err.Error(), "[5:") {
		t.Fatalf("the refusal must name the line: %v", err)
	}
}

func TestMultilineNotesDecode(t *testing.T) {
	d, err := Decode([]byte("notes:\n  - text: a free-text note; a longer note continues\n      on the next line, indented deeper than its key\n"))
	if err != nil {
		t.Fatalf("the template's notes form must decode: %v", err)
	}
	if len(d.Notes) != 1 || !strings.Contains(d.Notes[0].Text, "on the next line") {
		t.Fatalf("the folded note must keep its text: %#v", d.Notes)
	}
}

// rcZero — the plain silent-success expectation.
func rcZero() *int {
	rc := 0
	return &rc
}

// registryWitness — resolvers with full control: Stands and
// Requirement answer from sets, WishLine from a line-id → text map.
func registryWitness(stands, reqs map[string]bool, wish map[string]string) *Resolvers {
	return &Resolvers{
		Stands:      func(id string) bool { return stands[id] },
		Requirement: func(id string) bool { return reqs[id] },
		WishLine: func(id string) (string, bool) {
			text, ok := wish[id]
			return text, ok
		},
	}
}

func TestRegistryRefusalFamily(t *testing.T) {
	rowed := []Row{{Run: Cmd{Line: "x"}, Expect: Expect{RC: rcZero()}}}
	cases := []struct {
		name string
		d    Document
		r    *Resolvers
		want []string
	}{
		{
			"a bad requirement id",
			Document{Requirements: []Requirement{{ID: "3.2", Brief: []string{"W1"}}}},
			nil,
			[]string{"requirements[0].id", "R followed by letters, digits and dots"},
		},
		{
			"an empty requirement id",
			Document{Requirements: []Requirement{{ID: "", Brief: []string{"W1"}}}},
			nil,
			[]string{"requirements[0].id", "the id is empty"},
		},
		{
			"a duplicate requirement id",
			Document{Requirements: []Requirement{{ID: "R1"}, {ID: "R1"}}},
			nil,
			[]string{"requirements[1].id", "appears twice"},
		},
		{
			"a broken range",
			Document{Requirements: []Requirement{{ID: "R1", Brief: []string{"W5..W2"}}}},
			nil,
			[]string{"requirements[0].brief[0]", "must ascend"},
		},
		{
			"a range outside the wish",
			Document{Requirements: []Requirement{{ID: "R1", Brief: []string{"W9..W11"}}}},
			registryWitness(nil, nil, nil),
			[]string{"requirements[0].brief[0]", "the wish carries no line W9..W11"},
		},
		{
			"a dangling cover",
			Document{Behaviors: []Behavior{{ID: "a", Name: "n", Covers: []string{"R9"}, Rows: rowed}}},
			registryWitness(nil, nil, nil),
			[]string{"behaviors[0].covers[0]", `no requirement "R9" stands`},
		},
		{
			"an empty marker reason",
			Document{NotRequirements: []Marker{{Lines: []string{"W2"}, Reason: "  "}}},
			registryWitness(nil, nil, nil),
			[]string{"not-a-requirement[0].reason", "the reason is empty"},
		},
		{
			"an empty marker quote",
			Document{NotRequirements: []Marker{{Lines: []string{"W2"}, Reason: "a heading"}}},
			registryWitness(nil, nil, map[string]string{"W2": "Setup"}),
			[]string{"not-a-requirement[0].quote", "the quote is empty"},
		},
		{
			"a marker quote that is not the wish's own words",
			Document{NotRequirements: []Marker{{Lines: []string{"W2"}, Quote: "a heading", Reason: "a heading"}}},
			registryWitness(nil, nil, map[string]string{"W2": "Setup"}),
			[]string{"not-a-requirement[0].quote", `want "Setup"`},
		},
		{
			"a marker beside broken ranges and a resolver",
			Document{NotRequirements: []Marker{{Lines: []string{"junk"}, Quote: "x", Reason: "r"}}},
			registryWitness(nil, nil, map[string]string{"W1": "one"}),
			[]string{"not-a-requirement[0].lines[0]"},
		},
		{
			"a marker beside a missing line and a resolver",
			Document{NotRequirements: []Marker{{Lines: []string{"W9"}, Quote: "x", Reason: "r"}}},
			registryWitness(nil, nil, map[string]string{"W1": "one"}),
			[]string{"not-a-requirement[0].lines[0]", "the wish carries no line W9"},
		},
		{
			"a marker without a resolver passes the quote unchecked",
			Document{NotRequirements: []Marker{{Lines: []string{"W1"}, Quote: "whatever", Reason: "r"}}},
			nil,
			[]string{},
		},
		{
			"a marker quote of several lines in order",
			Document{NotRequirements: []Marker{{Lines: []string{"W2, W4"}, Quote: "Setup\nTeardown", Reason: "framing lines"}}},
			registryWitness(nil, nil, map[string]string{"W2": "Setup", "W4": "Teardown"}),
			[]string{},
		},
		{
			"an empty expect",
			Document{Behaviors: []Behavior{{ID: "a", Name: "n", Rows: []Row{{Run: Cmd{Line: "x"}}}}}},
			nil,
			[]string{"behaviors[0].rows[0].expect", "the expect is empty"},
		},
		{
			"a rowless behavior",
			Document{Behaviors: []Behavior{{ID: "a", Name: "n"}}},
			nil,
			[]string{"behaviors[0].rows", "at least one row"},
		},
		{
			"a bad artifact glob",
			Document{Artifacts: []string{"/abs/**"}},
			nil,
			[]string{"artifacts[0]", "law:paths"},
		},
		{
			"a remove of a requirement that does not stand",
			Document{Remove: []string{"R1"}},
			registryWitness(nil, nil, nil),
			[]string{"remove[0]", `the corpus carries no requirement "R1"`},
		},
		{
			"a remove of an id of neither family",
			Document{Remove: []string{"R1/b"}},
			nil,
			[]string{"remove[0]", "or a requirement id"},
		},
	}
	for _, tc := range cases {
		r := Validate(tc.d, tc.r)
		if tc.want != nil && len(tc.want) == 0 {
			if len(r) != 0 {
				t.Errorf("%s: must pass clean, got %v", tc.name, r)
			}
			continue
		}
		joined := ""
		for _, ref := range r {
			joined += ref.Error() + "\n"
		}
		for _, want := range tc.want {
			if !strings.Contains(joined, want) {
				t.Errorf("%s: missing %q in:\n%s", tc.name, want, joined)
			}
		}
	}
}

func TestRegistryResolvesAgainstTheWitnesses(t *testing.T) {
	// The whole lawful registry: ranges inside the wish, a marker
	// quoting the wish's own lines, covers of standing requirements,
	// a remove of what stands.
	rc := 0
	wish := map[string]string{"W1": "one", "W2": "two", "W3": "three", "W4": "four", "W5": "five", "W6": "six", "W7": "seven"}
	d := Document{
		Requirements: []Requirement{{ID: "R1", Brief: []string{"W1..W3, W7"}}},
		NotRequirements: []Marker{{Lines: []string{"W4..W6"}, Quote: "four\nfive\nsix",
			Reason: "the framing lines are not demands"}},
		Behaviors: []Behavior{{ID: "a", Name: "n", Covers: []string{"R1"},
			Rows: []Row{{Run: Cmd{Line: "x"}, Expect: Expect{RC: &rc}}}}},
		Remove: []string{"old/thing", "R2"},
	}
	r := Validate(d, registryWitness(map[string]bool{"old/thing": true, "R2": true}, map[string]bool{"R1": true}, wish))
	if len(r) != 0 {
		t.Fatalf("the lawful registry must pass: %v", r)
	}
}
