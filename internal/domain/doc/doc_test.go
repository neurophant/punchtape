// The document codec's and validator's own battery: the canonical
// round trip, the armor's refusals and the shape laws with their law
// citations.
package doc

import (
	"os"
	"strings"
	"testing"

	"github.com/neurophant/punchtape/internal/domain/fprint"
)

var example = `work: the first document
notes:
    - kind: work
      text: a note
build: make all
remove: []
behaviors:
    - id: tool/build
      name: Builds
      prose: The build.
      claims:
        - src/build.go:1-9
      rows:
        - run: ./check all
          expect:
            rc: 0
            out: "contains: ok"
            files:
                out/built.txt: "contains: done"
        - run: ./check exact
          expect:
            out: ok
`

func TestDecodeMarshalRoundTrip(t *testing.T) {
	d, err := Decode([]byte(example))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(d.Behaviors) != 1 || d.Behaviors[0].ID != "tool/build" {
		t.Fatalf("decode shape: %+v", d)
	}
	if d.Behaviors[0].Rows[0].Run.Line != "./check all" {
		t.Fatalf("cmd line: %+v", d.Behaviors[0].Rows[0].Run)
	}
	again, err := Decode(Marshal(d))
	if err != nil {
		t.Fatalf("re-decode: %v", err)
	}
	if string(Marshal(again)) != string(Marshal(d)) {
		t.Fatalf("canonical bytes not stable:\n%s\n%s", Marshal(d), Marshal(again))
	}
	// Normalization: reformatting the same document changes nothing.
	reformatted := strings.ReplaceAll(example, "make all", `"make all"`)
	d2, _, err := Normalize([]byte(reformatted))
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if string(Marshal(d2)) != string(Marshal(d)) {
		t.Fatalf("normalization must fold reformatting")
	}
}

func TestDecodeRefusals(t *testing.T) {
	cases := map[string]string{
		"two documents":         "build: x\n---\nbuild: x\n",
		"anchor":                "build: &a make\n",
		"alias":                 "x: &a 1\nbuild: *a\n",
		"unknown field":         "build: x\nbogus: 1\n",
		"duplicate key":         "build: x\nbuild: y\n",
		"tab indent":            "build:\n\t\tx\n",
		"bad run shape":         "behaviors:\n    - id: a/b\n      name: n\n      rows:\n        - run: 42\n",
		"the dead words field":  "demands: [a word]\n",
		"the dead answer list":  "answers:\n    - {lines: [D1], text: t}\n",
		"the dead capture sign": "no-new-demands: true\n",
	}
	for name, in := range cases {
		if _, err := Decode([]byte(in)); err == nil {
			t.Fatalf("%s: decode must refuse", name)
		}
	}
	// The dead quote field names itself: the strict decode's refusal
	// carries the unknown field.
	_, err := Decode([]byte("requirements:\n    - id: R1\n      lines: [D1]\n      quote: x\n      prose: p\n"))
	if err == nil || !strings.Contains(err.Error(), `unknown field "quote"`) {
		t.Fatalf("the quote field must be refused by name: %v", err)
	}
}

func TestValidate(t *testing.T) {
	standing := map[string]bool{"old/thing": true}
	ok := Document{
		Work: "w",
		Behaviors: []Behavior{
			{ID: "a/b", Name: "n", Prose: "p", Rows: []Row{{Run: Cmd{Line: "x"}, Expect: Expect{Out: "1"}}}},
		},
	}
	if r := Validate(ok, &Resolvers{Stands: func(id string) bool { return standing[id] }}); len(r) != 0 {
		t.Fatalf("valid document refused: %v", r)
	}
	bad := Document{
		Work: "w",
		Behaviors: []Behavior{
			{ID: "A/B", Name: "n", Prose: "p"},
			{ID: "a/b", Name: "n", Prose: "p"},
			{ID: "a/b", Name: "n", Prose: "p"},
			{ID: "c", Name: "n", Prose: "p", Rows: []Row{{Run: Cmd{Line: "x"}, Expect: Expect{Files: map[string]string{"/abs": "contains: x"}}}}},
		},
		Remove: []string{"no/such"},
	}
	r := Validate(bad, &Resolvers{Stands: func(id string) bool { return standing[id] }})
	var sawUpper, sawDup, sawAbs, sawMissing bool
	for _, ref := range r {
		switch {
		case strings.Contains(ref.Where, "behaviors[0].id"):
			sawUpper = ref.Law == LawDocument
		case strings.Contains(ref.Where, "behaviors[2].id"):
			sawDup = ref.Law == LawDocument
		case strings.Contains(ref.Where, "expect.files[/abs]"):
			sawAbs = ref.Law == LawPaths
		case strings.HasPrefix(ref.Where, "remove[0]"):
			sawMissing = ref.Law == LawDocument
		}
	}
	if !sawUpper || !sawDup || !sawAbs || !sawMissing {
		t.Fatalf("refusals incomplete: %v", r)
	}
}

// TestIntentCarriesAnswer — the answer outcome's own carrier: text
// rides an answer intent and an answer intent alone.
func TestIntentCarriesAnswer(t *testing.T) {
	good := Document{Work: "w", Intents: []Intent{
		{Lines: []string{"D1"}, Outcome: OutcomeAnswer, Prose: "the retold ask", Text: "the answer itself"},
	}}
	if r := Validate(good, nil); len(r) != 0 {
		t.Fatalf("an answer intent with its text must pass: %v", r)
	}
	noText := Document{Work: "w", Intents: []Intent{
		{Lines: []string{"D1"}, Outcome: OutcomeAnswer, Prose: "the retold ask"},
	}}
	if r := Validate(noText, nil); len(r) != 1 || r[0].Where != "intents[0].text" || r[0].Law != LawDemand {
		t.Fatalf("an answer intent without its text must be refused at its field: %v", r)
	}
	strayText := Document{Work: "w", Markers: []Marker{{Lines: []string{"D1"}, Reason: "a heading"}},
		Intents: []Intent{
			{Lines: []string{"D1"}, Outcome: OutcomeMark, Prose: "the retold ask", Text: "not an answer"},
		}}
	if r := Validate(strayText, nil); len(r) != 1 || r[0].Where != "intents[0].text" || r[0].Law != LawDemand {
		t.Fatalf("text on a non-answer intent must be refused at its field: %v", r)
	}
}

// TestIntentCarriesRun — the execute outcome's own carrier: the run
// rides an execute intent and an execute intent alone.
func TestIntentCarriesRun(t *testing.T) {
	good := Document{Work: "w", Intents: []Intent{
		{Lines: []string{"D1"}, Outcome: OutcomeExecute, Prose: "the retold ask", Run: Cmd{Line: "./tool"}},
	}}
	if r := Validate(good, nil); len(r) != 0 {
		t.Fatalf("an execute intent with its run must pass: %v", r)
	}
	noRun := Document{Work: "w", Intents: []Intent{
		{Lines: []string{"D1"}, Outcome: OutcomeExecute, Prose: "the retold ask"},
	}}
	if r := Validate(noRun, nil); len(r) != 1 || r[0].Where != "intents[0].run" || r[0].Law != LawDemand {
		t.Fatalf("an execute intent without its run must be refused at its field: %v", r)
	}
	strayRun := Document{Work: "w", Markers: []Marker{{Lines: []string{"D1"}, Reason: "a heading"}},
		Intents: []Intent{
			{Lines: []string{"D1"}, Outcome: OutcomeMark, Prose: "the retold ask", Run: Cmd{Line: "./tool"}},
		}}
	if r := Validate(strayRun, nil); len(r) != 1 || r[0].Where != "intents[0].run" || r[0].Law != LawDemand {
		t.Fatalf("a run on a non-execute intent must be refused at its field: %v", r)
	}
}

// TestRemoveKeyGrammar — the five spellings told apart mechanically,
// and the refusal of a key that stands nowhere.
func TestRemoveKeyGrammar(t *testing.T) {
	for key, kind := range map[string]RemoveKeyKind{
		"R1":           RemoveRequirement,
		"R2.10":        RemoveRequirement,
		"tool/old":     RemoveBehavior,
		"tool":         RemoveBehavior,
		"D4":           RemoveMarkerOrOrder,
		"src/a.go:1-9": RemoveExclusion,
	} {
		if got := ParseRemoveKey(key); got != kind {
			t.Fatalf("the key %q must parse as kind %d, got %d", key, kind, got)
		}
	}
	for _, key := range []string{"", "X1", "no..good", "src/a.go:1", "D0", "tool/OLD"} {
		if got := ParseRemoveKey(key); got != RemoveUnknown {
			t.Fatalf("the key %q must not parse, got kind %d", key, got)
		}
	}
	// A key that stands nowhere is refused by its named place.
	d := Document{Work: "w", Remove: []string{"D4"}}
	r := Validate(d, &Resolvers{
		MarkerStands: func(int) bool { return false },
		OrderStands:  func(int) bool { return false },
	})
	if len(r) != 1 || !strings.Contains(r[0].Fact, "the corpus carries no marker or order on D4") {
		t.Fatalf("an unresolved remove key must be refused: %v", r)
	}
}

// TestMergeByKey — the registry lists merge onto the standing ones
// by each part's own key: a repeated key replaces, a stranger is
// appended, unlisted parts stand.
func TestMergeByKey(t *testing.T) {
	standing := Registry{
		Requirements: map[string]Requirement{"R1": {ID: "R1", Lines: []string{"D1"}, Prose: "one"}},
		Order:        []string{"R1"},
		Markers:      []Marker{{Lines: []string{"D2"}, Reason: "first"}},
		Exclusions:   []Exclusion{{Range: "src/a.go:1-4", Reason: "generated"}},
		Open:         []OpenOrder{{Demand: "D5", Question: "stage?", State: StateAwaiting, OpenedAtLines: 4}},
	}
	d := Document{
		Markers:    []Marker{{Lines: []string{"D2"}, Reason: "retold"}},
		Exclusions: []Exclusion{{Range: "src/a.go:1-4", Reason: "regenerated"}, {Range: "src/b.go:1-2", Reason: "new"}},
		Open:       []OpenOrder{{Demand: "D6", Question: "later?", State: StateDeferred, Reason: "the human said later"}},
	}
	out := MergeRegistry(standing, d)
	if len(out.Markers) != 1 || out.Markers[0].Reason != "retold" {
		t.Fatalf("a repeated marker key replaces: %+v", out.Markers)
	}
	if len(out.Exclusions) != 2 || out.Exclusions[0].Reason != "regenerated" || out.Exclusions[1].Range != "src/b.go:1-2" {
		t.Fatalf("a repeated exclusion key replaces, a stranger appends: %+v", out.Exclusions)
	}
	if len(out.Open) != 2 || out.Open[0].Question != "stage?" || out.Open[1].Demand != "D6" {
		t.Fatalf("unlisted orders stand, strangers append: %+v", out.Open)
	}
	// The same line set spelled two ways is one marker key.
	if MarkerKey(Marker{Lines: []string{"D4", "D2..D3"}}) != MarkerKey(Marker{Lines: []string{"D2..D4"}}) {
		t.Fatal("two spellings of one line set must fold to one key")
	}
	// A remove line retires the marker standing on it, whole.
	out = MergeRegistry(standing, Document{Remove: []string{"D2"}})
	if len(out.Markers) != 0 {
		t.Fatalf("the marker on a removed line retires: %+v", out.Markers)
	}
}

// TestNilEmptyIndistinguishable — an absent list and an empty one
// merge to the same standing: nil and empty are one spelling.
func TestNilEmptyIndistinguishable(t *testing.T) {
	standing := Registry{
		Requirements: map[string]Requirement{"R1": {ID: "R1", Lines: []string{"D1"}, Prose: "one"}},
		Order:        []string{"R1"},
		Markers:      []Marker{{Lines: []string{"D2"}, Reason: "stays"}},
	}
	absent := MergeRegistry(standing, Document{})
	empty := MergeRegistry(standing, Document{Markers: []Marker{}})
	if string(Marshal(Document{Markers: absent.Markers})) != string(Marshal(Document{Markers: empty.Markers})) {
		t.Fatalf("nil and empty must merge to the same standing:\n%v\n%v", absent.Markers, empty.Markers)
	}
}

// TestDoneFieldShape — the completion act's own shape: done: true
// alone; any other field beside the flag is refused.
func TestDoneFieldShape(t *testing.T) {
	if r := Validate(Document{Done: true}, nil); len(r) != 0 {
		t.Fatalf("a bare done document must pass: %v", r)
	}
	for name, d := range map[string]Document{
		"work":     {Done: true, Work: "w"},
		"behavior": {Done: true, Behaviors: []Behavior{{ID: "a/b", Name: "n", Prose: "p", Rows: []Row{{Run: Cmd{Line: "x"}, Expect: Expect{Out: "1"}}}}}},
		"remove":   {Done: true, Remove: []string{"a/b"}},
	} {
		r := Validate(d, nil)
		if len(r) != 1 || r[0].Where != "done" || r[0].Law != LawDocument ||
			!strings.Contains(r[0].Fact, "a done document carries only done: true") {
			t.Fatalf("%s beside done must be refused by the shape law: %v", name, r)
		}
	}
}

func TestCmdSpellingsAndRefusals(t *testing.T) {
	line := Cmd{Line: "sh work"}
	argv := Cmd{Argv: []string{"./tool", "in", "out"}}
	d, err := Decode([]byte("behaviors:\n    - id: a/b\n      name: n\n      prose: p\n      rows:\n        - run: sh work\n        - run: [./tool, in, out]\n"))
	if err != nil {
		t.Fatal(err)
	}
	rows := d.Behaviors[0].Rows
	if rows[0].Run.Line != "sh work" || len(rows[1].Run.Argv) != 3 {
		t.Fatalf("spellings: %+v", rows)
	}
	_ = line
	_ = argv
	for name, in := range map[string]string{
		"empty argv":          "behaviors:\n    - id: a/b\n      name: n\n      prose: p\n      rows:\n        - run: []\n",
		"argv of non-strings": "behaviors:\n    - id: a/b\n      name: n\n      prose: p\n      rows:\n        - run: [1, 2]\n",
		"map run":             "behaviors:\n    - id: a/b\n      name: n\n      prose: p\n      rows:\n        - run: {a: b}\n",
	} {
		if _, err := Decode([]byte(in)); err == nil {
			t.Fatalf("%s must be refused", name)
		}
	}
}

func TestRegistryFieldsRoundTrip(t *testing.T) {
	rc := 0
	d := Document{
		Work:         "w",
		Requirements: []Requirement{{ID: "R3.2", Lines: []string{"D1..D4, D9"}, Prose: "the unwinding", Refines: "R1"}},
		Markers:      []Marker{{Lines: []string{"D5"}, Reason: "not a demand"}},
		Exclusions:   []Exclusion{{Range: "gen/out.go:1-40", Reason: "generated"}},
		Open:         []OpenOrder{{Demand: "D14", Question: "stage two?", State: "awaiting"}},
		Notes:        []Note{{Kind: "decision", Text: "a note", Refs: []string{"D14", "R3.2"}}},
		Behaviors: []Behavior{{ID: "tool/x", Name: "n", Prose: "p", Claims: []string{"src/x.go:1-8"}, Covers: []string{"R3.2"},
			Rows: []Row{{Run: Cmd{Line: "x"}, Expect: Expect{RC: &rc, Out: "1"}}}}},
	}
	back, err := Decode(Marshal(d))
	if err != nil {
		t.Fatalf("the registry fields must decode: %v", err)
	}
	if string(Marshal(back)) != string(Marshal(d)) {
		t.Fatalf("canonical bytes must carry the registry without loss:\n%s\n%s", Marshal(d), Marshal(back))
	}
	if back.Requirements[0].ID != "R3.2" || back.Requirements[0].Refines != "R1" ||
		back.Exclusions[0].Range != "gen/out.go:1-40" || back.Open[0].Demand != "D14" ||
		back.Notes[0].Kind != "decision" || back.Behaviors[0].Claims[0] != "src/x.go:1-8" {
		t.Fatalf("the round trip lost fields: %#v", back)
	}
}

// TestIntentCarriersRoundTrip — the intent's own carrier fields ride
// the canonical bytes: an answer's text and an execute's command
// come back whole, and an intent without a carrier omits it.
func TestIntentCarriersRoundTrip(t *testing.T) {
	d := Document{
		Work: "w",
		Intents: []Intent{
			{Lines: []string{"D1"}, Outcome: OutcomeAnswer, Prose: "the ask retold", Text: "the answer itself"},
			{Lines: []string{"D2"}, Outcome: OutcomeExecute, Prose: "the ask retold", Run: Cmd{Line: "./tool --version"}},
			{Lines: []string{"D3"}, Outcome: OutcomeWork, Prose: "the ask retold"},
		},
	}
	back, err := Decode(Marshal(d))
	if err != nil {
		t.Fatalf("the intent carriers must decode: %v", err)
	}
	if back.Intents[0].Text != "the answer itself" || back.Intents[1].Run.Line != "./tool --version" {
		t.Fatalf("the carriers round-tripped: %+v", back.Intents)
	}
	if back.Intents[2].Text != "" || !back.Intents[2].Run.IsZero() {
		t.Fatalf("a carrier-less intent stays carrier-less: %+v", back.Intents[2])
	}
	if bytes := Marshal(d); !strings.Contains(string(bytes), "text: the answer itself") || strings.Count(string(bytes), "run:") != 1 {
		t.Fatalf("the canonical bytes carry only the standing carriers:\n%s", bytes)
	}
}

func TestNormalizeErrorAndEmpty(t *testing.T) {
	if _, _, err := Normalize([]byte("bogus: [")); err == nil {
		t.Fatal("broken bytes must be refused")
	}
	d := Document{}
	if len(d.Behaviors) != 0 {
		t.Fatal("the naked document carries no behaviors")
	}
	if d.BuildCmd() != "" {
		t.Fatal("no build stands")
	}
	build := "./build"
	d.Build = &build
	if d.BuildCmd() != "./build" {
		t.Fatal("a build declared is read back")
	}
}

func TestBehaviorBytesRoundTrip(t *testing.T) {
	b := Behavior{ID: "x/y", Name: "Y", Prose: "p", Claims: []string{"s/a.go:1-2"}, Rows: []Row{{Run: Cmd{Line: "r"}, Expect: Expect{Out: "1"}}}}
	back, err := DecodeBehavior(MarshalBehavior(b))
	if err != nil {
		t.Fatal(err)
	}
	if back.ID != b.ID || len(back.Rows) != 1 || back.Claims[0] != "s/a.go:1-2" {
		t.Fatalf("behavior round trip: %+v", back)
	}
	if _, err := DecodeBehavior([]byte("id: [\n")); err == nil {
		t.Fatal("broken behavior bytes must be refused")
	}
	if len(MarshalRow(b.Rows[0])) == 0 {
		t.Fatal("row bytes must exist")
	}
}

func TestExpectCompare(t *testing.T) {
	rc0, rc3 := 0, 3
	files := map[string]string{"a.txt": "contains: xx", "gone.txt": "contains: y"}
	e := Expect{RC: &rc3, Out: "exact", Files: files}
	read := func(p string) ([]byte, error) {
		if p == "gone.txt" {
			return nil, os.ErrNotExist
		}
		return []byte("xx\n"), nil
	}
	facts := Compare(e, 3, "exact", read)
	if len(facts) != 1 || !strings.Contains(facts[0], "gone.txt") {
		t.Fatalf("the absent file is a broken fact: %v", facts)
	}
	facts = Compare(e, 0, "no", read)
	if len(facts) != 3 {
		t.Fatalf("exit, stdout and files must all break: %v", facts)
	}
	exact := Expect{RC: &rc0, Out: "contains: sub", Files: map[string]string{"a.txt": "xx\n"}}
	if facts := Compare(exact, 0, "say sub now", read); len(facts) != 0 {
		t.Fatalf("the honest row must be green: %v", facts)
	}
	_ = rc3
}

func TestRefusalErrorLine(t *testing.T) {
	r := Refusal{Where: "build", Law: LawDocument, Fact: "the build command is empty"}
	if !strings.Contains(r.Error(), "law:document") {
		t.Fatalf("the refusal names its law: %s", r.Error())
	}
}

func TestWalkReachesDeepShapes(t *testing.T) {
	// Mappings under sequences under mappings, literal blocks: the
	// armor's walk must cross them without refusing the innocent.
	in := `build: |
    ./build all
behaviors:
    - id: a/b
      name: n
      prose: |
        long
        prose
      claims:
        - src/a.go:1-3
      rows:
        - run: [./tool, in, out]
          expect:
            files:
                out/x: "contains: y"
`
	d, err := Decode([]byte(in))
	if err != nil {
		t.Fatalf("deep shapes must decode: %v", err)
	}
	if d.Behaviors[0].Rows[0].Run.Argv[0] != "./tool" {
		t.Fatalf("argv row: %+v", d.Behaviors[0].Rows[0].Run)
	}
	// The argv spelling marshals back through its own branch.
	again, err := Decode(Marshal(d))
	if err != nil || len(again.Behaviors[0].Rows) != 1 {
		t.Fatalf("argv round trip: %v", err)
	}
}

func TestValidIDEdges(t *testing.T) {
	for _, id := range []string{"a", "a-b/c-d", "x9/y8"} {
		if !validID(id) {
			t.Fatalf("%q must be a legal id", id)
		}
	}
	for _, id := range []string{"/a", "a/", "a//b", "A/b", "a/B", "a b", "a_b", "a.b", "", "a/"} {
		if validID(id) {
			t.Fatalf("%q must be refused", id)
		}
	}
	if _, _, _, err := fprint.ParseLineRange("../x:1-2"); err == nil {
		t.Fatal("a parent hop is not a line range")
	}
	if _, _, _, err := fprint.ParseLineRange("x.go:1"); err == nil {
		t.Fatal("a range without both ends parses to nothing")
	}
}

func TestMismatchSpellingsAndKeys(t *testing.T) {
	long := strings.Repeat("x", 100)
	rc := 2
	e := Expect{RC: &rc, Out: "want", Files: map[string]string{"c": "w", "a": "w", "b": "w"}}
	facts := Compare(e, 0, long, func(string) ([]byte, error) { return []byte("z"), nil })
	if len(facts) != 5 {
		t.Fatalf("exit, stdout and three files: %v", facts)
	}
	for _, f := range facts {
		if strings.Contains(f, "…") != strings.HasPrefix(f, "stdout") {
			t.Fatalf("the long stdout is bounded: %v", facts)
		}
	}
}

func TestValidateEveryBranch(t *testing.T) {
	empty := ""
	d := Document{
		Work:  "w",
		Build: &empty,
		Behaviors: []Behavior{
			{ID: "a/b", Name: "n", Prose: "p", Rows: []Row{
				{Run: Cmd{Line: "x"}, Expect: Expect{Out: "1"}},
				{Expect: Expect{Files: map[string]string{"ok.txt": "x"}}},
			}},
		},
	}
	r := Validate(d, nil)
	joined := ""
	for _, ref := range r {
		joined += ref.Error() + "\n"
	}
	for _, want := range []string{
		"build", "the build command is empty",
		"rows[1].run", "the command is empty",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing %q in:\n%s", want, joined)
		}
	}
	// The document's behaviors keep their order and ids.
	docs := Document{Behaviors: []Behavior{{ID: "x/y"}}}
	if len(docs.Behaviors) != 1 || docs.Behaviors[0].ID != "x/y" {
		t.Fatalf("behaviors: %+v", docs.Behaviors)
	}
}

func TestDecodeBehaviorArmorRefusal(t *testing.T) {
	if _, err := DecodeBehavior([]byte("id: &x a\n")); err == nil {
		t.Fatal("the armor guards single behaviors too")
	}
}
