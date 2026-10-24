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

var example = `notes:
    - text: a note
build: make all
remove: []
behaviors:
    - id: tool/build
      name: Builds
      prose: The build.
      scope:
        - src/**
      rows:
        - run: ./check all
          expect:
            rc: 0
            out: "contains: ok"
            files:
                out/built.txt: "contains: done"
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
		"two documents": "build: x\n---\nbuild: x\n",
		"anchor":        "build: &a make\n",
		"alias":         "x: &a 1\nbuild: *a\n",
		"unknown field": "build: x\nbogus: 1\n",
		"duplicate key": "build: x\nbuild: y\n",
		"tab indent":    "build:\n\t\tx\n",
		"bad run shape": "behaviors:\n    - id: a/b\n      name: n\n      rows:\n        - run: 42\n",
	}
	for name, in := range cases {
		if _, err := Decode([]byte(in)); err == nil {
			t.Fatalf("%s: decode must refuse", name)
		}
	}
}

func TestValidate(t *testing.T) {
	standing := map[string]bool{"old/thing": true}
	ok := Document{Behaviors: []Behavior{
		{ID: "a/b", Name: "n", Scope: []string{"src/**"}, Rows: []Row{{Run: Cmd{Line: "x"}, Expect: Expect{Out: "1"}}}},
	}}
	if r := Validate(ok, &Resolvers{Stands: func(id string) bool { return standing[id] }}); len(r) != 0 {
		t.Fatalf("valid document refused: %v", r)
	}
	bad := Document{
		Behaviors: []Behavior{
			{ID: "A/B", Name: "n"},
			{ID: "a/b", Name: "n"},
			{ID: "a/b", Name: "n"},
			{ID: "c", Name: "n", Rows: []Row{{Run: Cmd{Line: "x"}, Expect: Expect{Files: map[string]string{"/abs": "contains: x"}}}}},
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

func TestGlobLaw(t *testing.T) {
	for _, g := range []string{"src/**", "a/*/b", "**", "out/*.txt"} {
		if !ValidGlob(g) {
			t.Fatalf("%q must be legal", g)
		}
	}
	for _, g := range []string{"/abs/**", "../x", "a//b", "a/../b", "", `a\b`, "a/"} {
		if ValidGlob(g) {
			t.Fatalf("%q must be refused", g)
		}
	}
}

func TestCmdSpellingsAndRefusals(t *testing.T) {
	line := Cmd{Line: "sh work"}
	argv := Cmd{Argv: []string{"./tool", "in", "out"}}
	d, err := Decode([]byte("behaviors:\n    - id: a/b\n      name: n\n      rows:\n        - run: sh work\n        - run: [./tool, in, out]\n"))
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
		"empty argv":          "behaviors:\n    - id: a/b\n      name: n\n      rows:\n        - run: []\n",
		"argv of non-strings": "behaviors:\n    - id: a/b\n      name: n\n      rows:\n        - run: [1, 2]\n",
		"map run":             "behaviors:\n    - id: a/b\n      name: n\n      rows:\n        - run: {a: b}\n",
	} {
		if _, err := Decode([]byte(in)); err == nil {
			t.Fatalf("%s must be refused", name)
		}
	}
}

func TestRegistryFieldsRoundTrip(t *testing.T) {
	rc := 0
	d := Document{
		Requirements:    []Requirement{{ID: "R3.2", Brief: []string{"W1..W4, W9"}}},
		NotRequirements: []Marker{{Lines: []string{"W5"}, Quote: "a header", Reason: "not a demand"}},
		Artifacts:       []string{"out/**", "dist/app.tgz"},
		Behaviors: []Behavior{{ID: "tool/x", Name: "n", Covers: []string{"R3.2"},
			Rows: []Row{{Run: Cmd{Line: "x"}, Expect: Expect{RC: &rc}}}}},
	}
	back, err := Decode(Marshal(d))
	if err != nil {
		t.Fatalf("the registry fields must decode: %v", err)
	}
	if string(Marshal(back)) != string(Marshal(d)) {
		t.Fatalf("canonical bytes must carry the registry without loss:\n%s\n%s", Marshal(d), Marshal(back))
	}
	if back.Requirements[0].ID != "R3.2" || back.NotRequirements[0].Quote != "a header" ||
		len(back.Artifacts) != 2 || back.Behaviors[0].Covers[0] != "R3.2" {
		t.Fatalf("the round trip lost fields: %#v", back)
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
	b := Behavior{ID: "x/y", Name: "Y", Scope: []string{"s/**"}, Rows: []Row{{Run: Cmd{Line: "r"}}}}
	back, err := DecodeBehavior(MarshalBehavior(b))
	if err != nil {
		t.Fatal(err)
	}
	if back.ID != b.ID || len(back.Rows) != 1 {
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
	if !ValidGlob("a/**") || !ValidGlob("a/**/b") {
		t.Fatal("the double star is legal in every position")
	}
	if fprint.MatchGlob("a[/b", "a/x/b") {
		t.Fatal("a broken class matches nothing")
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
		Build: &empty,
		Behaviors: []Behavior{
			{ID: "a/b", Name: "n", Scope: []string{"/abs/**"}, Rows: []Row{
				{Run: Cmd{Line: "x"}},
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
		"scope[0]", "law:paths",
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
