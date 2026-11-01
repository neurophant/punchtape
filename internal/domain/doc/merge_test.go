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
	standing := Behavior{ID: "a", Name: "Old", Prose: "was", Claims: []string{"x.go:1-2"}, Rows: []Row{lineRow("a", "1")}}
	submitted := Behavior{ID: "a", Name: "New", Prose: "now", Claims: []string{"y.go:1-2"}, Rows: []Row{lineRow("b", "2")}}
	merged := MergeBehavior(standing, submitted)
	if merged.Name != "New" || merged.Prose != "now" || strings.Join(merged.Claims, ",") != "y.go:1-2" {
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
			"R1": {ID: "R1", Lines: []string{"D1"}, Prose: "p1"},
			"R2": {ID: "R2", Lines: []string{"D2"}, Prose: "p2"},
		},
		Order:   []string{"R1", "R2"},
		Markers: []Marker{{Lines: []string{"D9"}, Reason: "old"}},
	}
	// A repeated id replaces its head whole and keeps its place; a
	// remove retires first and a re-add starts fresh at the end.
	merged := MergeRegistry(standing, Document{
		Remove:       []string{"R2"},
		Requirements: []Requirement{{ID: "R1", Lines: []string{"D1..D2"}, Prose: "p"}, {ID: "R2", Lines: []string{"D3"}, Prose: "p"}},
		Markers:      []Marker{{Lines: []string{"D8"}, Reason: "new"}},
		Exclusions:   []Exclusion{{Range: "gen/out.go:1-9", Reason: "generated"}},
		Open:         []OpenOrder{{Demand: "D8", Question: "stage two?", State: "awaiting"}},
	})
	if len(merged.Requirements) != 2 || strings.Join(merged.Requirements["R1"].Lines, ",") != "D1..D2" {
		t.Fatalf("the repeated id replaces its head: %#v", merged)
	}
	if strings.Join(merged.Order, ",") != "R1,R2" {
		t.Fatalf("the re-added id lands after the survivors: %v", merged.Order)
	}
	if len(merged.Markers) != 2 || merged.Markers[0].Reason != "old" || merged.Markers[1].Reason != "new" {
		t.Fatalf("the markers merge by their key — both stand: %#v", merged.Markers)
	}
	if len(merged.Exclusions) != 1 || merged.Exclusions[0].Range != "gen/out.go:1-9" {
		t.Fatalf("a carried exclusion appends by its key: %#v", merged.Exclusions)
	}
	if len(merged.Open) != 1 || merged.Open[0].Question != "stage two?" {
		t.Fatalf("a carried open order appends by its key: %#v", merged.Open)
	}
	// An absent list changes nothing; a carried one merges by key.
	keep := MergeRegistry(merged, Document{Requirements: []Requirement{}})
	if len(keep.Markers) != 2 || len(keep.Exclusions) != 1 || len(keep.Open) != 1 {
		t.Fatalf("a submission without the lists keeps the standing ones: %#v", keep)
	}
	replaced := MergeRegistry(keep, Document{
		Exclusions: []Exclusion{{Range: "gen/other.go:1-2", Reason: "generated"}},
		Open:       []OpenOrder{{Demand: "D1", Question: "later?", State: "deferred", Reason: "the human said later"}},
	})
	if len(replaced.Exclusions) != 2 || replaced.Exclusions[1].Range != "gen/other.go:1-2" ||
		len(replaced.Open) != 2 || replaced.Open[1].State != "deferred" {
		t.Fatalf("a carried list merges by key, strangers append: %#v", replaced)
	}
	// The removed id retires; the remove-and-re-add keeps it once.
	gone := MergeRegistry(standing, Document{Remove: []string{"R1"}})
	if _, ok := gone.Requirements["R1"]; ok || strings.Join(gone.Order, ",") != "R2" {
		t.Fatalf("the removed requirement leaves: %#v", gone)
	}
}

func TestDuplicateRunInOneBehaviorRefused(t *testing.T) {
	d := Document{Work: "w", Behaviors: []Behavior{{
		ID: "a/b", Name: "n", Prose: "p",
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
	d := Document{Work: "w", Behaviors: []Behavior{{
		ID: "a", Name: "n", Prose: "p",
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

// TestRemoveRetiresEveryKind — each of the five spellings retires
// its standing part: the requirement by id, the exclusion by range,
// the marker and the order by their line.
func TestRemoveRetiresEveryKind(t *testing.T) {
	standing := Registry{
		Requirements: map[string]Requirement{"R1": {ID: "R1", Lines: []string{"D1"}, Prose: "p"}},
		Order:        []string{"R1"},
		Markers:      []Marker{{Lines: []string{"D2..D3"}, Reason: "a heading"}},
		Exclusions:   []Exclusion{{Range: "gen/out.go:1-9", Reason: "generated"}},
		Open:         []OpenOrder{{Demand: "D5", Question: "stage?", State: StateAwaiting}},
	}
	out := MergeRegistry(standing, Document{Remove: []string{"R1", "D2", "gen/out.go:1-9", "D5"}})
	if len(out.Requirements) != 0 || len(out.Order) != 0 {
		t.Fatalf("the requirement retires: %#v", out)
	}
	if len(out.Markers) != 0 {
		t.Fatalf("the marker on the removed line retires whole: %#v", out.Markers)
	}
	if len(out.Exclusions) != 0 {
		t.Fatalf("the exclusion retires by its range: %#v", out.Exclusions)
	}
	if len(out.Open) != 0 {
		t.Fatalf("the order retires by its line: %#v", out.Open)
	}
}
