package doc

// The row-merge battery: the submission economy's own laws — the
// same run replaces the whole row, a new run is appended, unlisted
// rows stand, the head replaces, a re-spelled run is a new run, and
// a duplicate run in one behavior is refused by its named place.

import (
	"strings"
	"testing"
)

func lineRow(line, out string) Row {
	return Row{Run: Cmd{Line: line}, Expect: Expect{Out: out}}
}

func runsOf(rows []Row) []string {
	out := make([]string, len(rows))
	for i, r := range rows {
		if r.Run.Argv != nil {
			out[i] = strings.Join(r.Run.Argv, " ")
		} else {
			out[i] = r.Run.Line
		}
	}
	return out
}

func TestMergeRowsSameRunReplacesWholeRowInPlace(t *testing.T) {
	standing := []Row{lineRow("make one", "1"), lineRow("make two", "2"), lineRow("make three", "3")}
	submitted := []Row{lineRow("make two", "TWO")}
	merged := MergeRows(standing, submitted)
	if got := runsOf(merged); strings.Join(got, "|") != "make one|make two|make three" {
		t.Fatalf("the standing order must never move: %v", got)
	}
	if merged[1].Expect.Out != "TWO" || merged[0].Expect.Out != "1" || merged[2].Expect.Out != "3" {
		t.Fatalf("only the named run's row may change: %#v", merged)
	}
}

func TestMergeRowsExpectReplacesWhole(t *testing.T) {
	rc := 1
	standing := []Row{{Run: Cmd{Line: "x"}, Expect: Expect{RC: &rc, Out: "old", Files: map[string]string{"a": "b"}}}}
	submitted := []Row{lineRow("x", "new")}
	merged := MergeRows(standing, submitted)
	if merged[0].Expect.Out != "new" || merged[0].Expect.RC != nil || merged[0].Expect.Files != nil {
		t.Fatalf("a partial patch does not exist — the whole row replaces: %#v", merged[0])
	}
}

func TestMergeRowsNewRunAppendsAfterStanding(t *testing.T) {
	standing := []Row{lineRow("a", "1"), lineRow("b", "2")}
	submitted := []Row{lineRow("c", "3"), lineRow("b", "TWO")}
	merged := MergeRows(standing, submitted)
	if got := runsOf(merged); strings.Join(got, "|") != "a|b|c" {
		t.Fatalf("a new run appends after the standing rows: %v", got)
	}
	if merged[1].Expect.Out != "TWO" || merged[2].Expect.Out != "3" {
		t.Fatalf("the replace and the append must both land: %#v", merged)
	}
}

func TestMergeRowsEmptySides(t *testing.T) {
	standing := []Row{lineRow("a", "1")}
	if got := MergeRows(standing, nil); len(got) != 1 || got[0].Expect.Out != "1" {
		t.Fatalf("a head-only submission leaves the rows standing: %#v", got)
	}
	if got := MergeRows(nil, standing); len(got) != 1 || got[0].Expect.Out != "1" {
		t.Fatalf("a new behavior carries its rows as listed: %#v", got)
	}
}

func TestMergeBehaviorHeadReplaces(t *testing.T) {
	standing := Behavior{ID: "a", Name: "Old", Prose: "was", Scope: []string{"x/**"}, Rows: []Row{lineRow("a", "1")}}
	submitted := Behavior{ID: "a", Name: "New", Prose: "now", Scope: []string{"y/**"}, Rows: []Row{lineRow("b", "2")}}
	merged := MergeBehavior(standing, submitted)
	if merged.Name != "New" || merged.Prose != "now" || strings.Join(merged.Scope, ",") != "y/**" {
		t.Fatalf("the submitted head replaces whole: %#v", merged)
	}
	if got := runsOf(merged.Rows); strings.Join(got, "|") != "a|b" {
		t.Fatalf("the rows merge onto the standing ones: %v", got)
	}
}

func TestRespelledRunIsANewRun(t *testing.T) {
	standing := []Row{lineRow("make check", "1")}
	submitted := []Row{{Run: Cmd{Argv: []string{"make", "check"}}}}
	merged := MergeRows(standing, submitted)
	if len(merged) != 2 {
		t.Fatalf("a re-spelled run is a new run — the standing one stands: %#v", merged)
	}
}

func TestMergeRegistryLaws(t *testing.T) {
	standing := Registry{
		Requirements: map[string]Requirement{
			"R1": {ID: "R1", Brief: []string{"W1"}},
			"R2": {ID: "R2", Brief: []string{"W2"}},
		},
		Order:   []string{"R1", "R2"},
		Markers: []Marker{{Lines: []string{"W9"}, Quote: "the old tail", Reason: "old"}},
	}
	// A repeated id replaces its head whole and keeps its place; a
	// remove retires first and a re-add starts fresh at the end.
	merged := MergeRegistry(standing, Document{
		Remove:          []string{"R2"},
		Requirements:    []Requirement{{ID: "R1", Brief: []string{"W1..W2"}}, {ID: "R2", Brief: []string{"W3"}}},
		NotRequirements: []Marker{{Lines: []string{"W8"}, Quote: "the new tail", Reason: "new"}},
	})
	if len(merged.Requirements) != 2 || strings.Join(merged.Requirements["R1"].Brief, ",") != "W1..W2" {
		t.Fatalf("the repeated id replaces its head: %#v", merged)
	}
	if strings.Join(merged.Order, ",") != "R1,R2" {
		t.Fatalf("the re-added id lands after the survivors: %v", merged.Order)
	}
	if len(merged.Markers) != 1 || merged.Markers[0].Reason != "new" || merged.Markers[0].Quote != "the new tail" {
		t.Fatalf("the submitted markers replace the standing list entire: %#v", merged.Markers)
	}
	// An absent marker list and an absent artifact list change
	// nothing; a carried artifact list replaces.
	keep := MergeRegistry(merged, Document{Requirements: []Requirement{}})
	if len(keep.Markers) != 1 || keep.Markers[0].Reason != "new" || keep.Markers[0].Quote != "the new tail" {
		t.Fatalf("a submission without markers keeps the standing ones: %#v", keep.Markers)
	}
	withArtifacts := MergeRegistry(keep, Document{Artifacts: []string{"out/**"}})
	if strings.Join(withArtifacts.Artifacts, ",") != "out/**" {
		t.Fatalf("a carried artifact list replaces: %v", withArtifacts.Artifacts)
	}
	again := MergeRegistry(withArtifacts, Document{})
	if strings.Join(again.Artifacts, ",") != "out/**" {
		t.Fatalf("a submission without artifacts keeps the standing ones: %v", again.Artifacts)
	}
	// The removed id retires; the remove-and-re-add keeps it once.
	gone := MergeRegistry(standing, Document{Remove: []string{"R1"}})
	if _, ok := gone.Requirements["R1"]; ok || strings.Join(gone.Order, ",") != "R2" {
		t.Fatalf("the removed requirement leaves: %#v", gone)
	}
}

func TestDuplicateRunInOneBehaviorRefused(t *testing.T) {
	d := Document{Behaviors: []Behavior{{
		ID: "a/b", Name: "n",
		Rows: []Row{lineRow("same cmd", "1"), lineRow("other", "2"), lineRow("same cmd", "3")},
	}}}
	r := Validate(d, nil)
	if len(r) != 1 {
		t.Fatalf("the duplicate run must be the only refusal: %v", r)
	}
	got := r[0].Error()
	for _, want := range []string{
		"behaviors[0].rows[2].run",
		`"same cmd"`,
		`a/b`,
		"law:document",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("the refusal must carry %q: %s", want, got)
		}
	}
}

func TestDuplicateRunArgvSpellingRefused(t *testing.T) {
	d := Document{Behaviors: []Behavior{{
		ID: "a", Name: "n",
		Rows: []Row{
			{Run: Cmd{Argv: []string{"make", "check"}}, Expect: Expect{Out: "1"}},
			{Run: Cmd{Argv: []string{"make", "check"}}, Expect: Expect{Out: "2"}},
		},
	}}}
	r := Validate(d, nil)
	if len(r) != 1 || !strings.Contains(r[0].Error(), "behaviors[0].rows[1].run") {
		t.Fatalf("the argv-spelled duplicate must be refused at its place: %v", r)
	}
}
