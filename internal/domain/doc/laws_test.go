package doc

// The validation map: the eleven rules, one table — each row a rule,
// a submission and the refusal the rule answers. The pass tails below
// carry the boundary: a copied prose, an overlapping span and a
// cycling refinement all pass the form — the machine judges
// existence and set arithmetic, never the sense of strings.

import (
	"strings"
	"testing"
)

// lawsWitness — resolvers with full control for the gates' battery:
// the corpus sets, the demand stream's own texts, the tree's line
// counts, the printed law ids and the holes the tree carries.
func lawsWitness(stands, reqs, laws map[string]bool, demand map[string]string, files map[string]int, hole func(rel string, first, last int) bool) *Resolvers {
	if hole == nil {
		hole = func(rel string, first, last int) bool { return true }
	}
	return &Resolvers{
		Stands:      func(id string) bool { return stands[id] },
		Requirement: func(id string) bool { return reqs[id] },
		DemandLine: func(id string) (string, bool) {
			text, ok := demand[id]
			return text, ok
		},
		FileLines: func(rel string) (int, bool) {
			n, ok := files[rel]
			return n, ok
		},
		Law:  func(id string) bool { return laws[id] },
		Hole: hole,
	}
}

// lawfulDoc — the smallest submission every gate accepts: the
// narrative, one requirement on one line, one behavior claiming one
// range with a pinned row.
func lawfulDoc() Document {
	return Document{
		Work:         "the narrative",
		Requirements: []Requirement{{ID: "R1", Lines: []string{"D2"}, Prose: "the unwinding"}},
		Behaviors: []Behavior{{
			ID: "tool/x", Name: "X", Prose: "The story.",
			Claims: []string{"src/tool.go:1-3"}, Covers: []string{"R1"},
			Rows: []Row{{Run: Cmd{Line: "./tool"}, Expect: Expect{Out: "ok"}}},
		}},
	}
}

// theWitness — the lawful document's own standing facts: an empty
// corpus (the submission builds fresh), one standing requirement,
// the two-line stream, the three-line file.
func theWitness() *Resolvers {
	return lawsWitness(
		nil,
		map[string]bool{"R1": true},
		map[string]bool{"law:done": true},
		map[string]string{"D1": "a heading", "D2": "do the work"},
		map[string]int{"src/tool.go": 3},
		nil,
	)
}

func refusalText(r []Refusal) string {
	var b strings.Builder
	for _, ref := range r {
		b.WriteString(ref.Error())
		b.WriteString("\n")
	}
	return b.String()
}

// TestValidationRules — the eleven rules of the map plus the shape
// laws that ride beside them, one row per refusal: ids, ranges and
// claims, the six outcomes and their carriers, the tail declared
// once, requirements, the marker conflict, exclusions, behaviors,
// the mandate, notes, the done act — and the narrative, the build,
// the open orders and the remove keys.
func TestValidationRules(t *testing.T) {
	// standing — the witness of a living corpus: two behaviors and a
	// requirement stand, the stream and the tree resolve.
	standing := lawsWitness(
		map[string]bool{"tool/x": true, "tool/old": true},
		map[string]bool{"R1": true},
		map[string]bool{"law:done": true},
		map[string]string{"D1": "a heading", "D2": "do the work"},
		map[string]int{"src/tool.go": 3}, nil)
	// holey — the witness of a tree whose only hole is line 2 of the
	// one file.
	holey := lawsWitness(nil, map[string]bool{"R1": true}, map[string]bool{"law:done": true},
		map[string]string{"D1": "a heading", "D2": "do the work"},
		map[string]int{"src/tool.go": 3}, func(rel string, first, last int) bool {
			return rel == "src/tool.go" && last >= 2 && first <= 2
		})
	behavior := func(mut func(b *Behavior)) Document {
		d := lawfulDoc()
		mut(&d.Behaviors[0])
		return d
	}
	excluded := func(ex Exclusion) Document {
		d := lawfulDoc()
		d.Exclusions = []Exclusion{ex}
		return d
	}
	rowed := []Row{{Run: Cmd{Line: "x"}, Expect: Expect{Out: "1"}}}
	cases := []struct {
		rule string
		d    Document
		r    *Resolvers
		want []string
	}{
		// Rule 1 — ids: path-shaped behavior ids, R-shaped requirement
		// ids, each unique in the submission.
		{"1 ids: an empty behavior id", behavior(func(b *Behavior) { b.ID = "" }), theWitness(),
			[]string{"behaviors[0].id", "the id is empty"}},
		{"1 ids: a behavior id of the wrong alphabet", behavior(func(b *Behavior) { b.ID = "A/B" }), theWitness(),
			[]string{"behaviors[0].id", "the id must be [a-z0-9-] segments joined by '/'"}},
		{"1 ids: a behavior id twice in one submission", Document{Work: "w", Behaviors: []Behavior{
			{ID: "a/b", Name: "n", Prose: "p", Rows: rowed},
			{ID: "a/b", Name: "m", Prose: "p", Rows: []Row{{Run: Cmd{Line: "y"}, Expect: Expect{Out: "2"}}}},
		}}, nil, []string{"behaviors[1].id", "appears twice in one submission"}},
		{"1 ids: a requirement id of the wrong alphabet", Document{Work: "w",
			Requirements: []Requirement{{ID: "3.2", Lines: []string{"D1"}}}}, nil,
			[]string{"requirements[0].id", "R followed by letters, digits and dots"}},
		{"1 ids: an empty requirement id", Document{Work: "w",
			Requirements: []Requirement{{ID: "", Lines: []string{"D1"}}}}, nil,
			[]string{"requirements[0].id", "the id is empty"}},
		{"1 ids: a requirement id twice in one submission", Document{Work: "w",
			Requirements: []Requirement{{ID: "R1"}, {ID: "R1"}}}, nil,
			[]string{"requirements[1].id", "appears twice"}},

		// Rule 2 — ranges of the D grammar and claims of the
		// path:first-last grammar, every line real.
		{"2 ranges: a broken range", Document{Work: "w",
			Requirements: []Requirement{{ID: "R1", Lines: []string{"D5..D2"}}}}, nil,
			[]string{"requirements[0].lines[0]", "must ascend"}},
		{"2 ranges: a range outside the stream", Document{Work: "w",
			Requirements: []Requirement{{ID: "R1", Lines: []string{"D9..D11"}}}},
			lawsWitness(nil, nil, nil, nil, nil, nil),
			[]string{"requirements[0].lines[0]", "the demand stream carries no line D9..D11"}},
		{"2 claims: a claim beyond the file", behavior(func(b *Behavior) { b.Claims = []string{"src/tool.go:1-9"} }), theWitness(),
			[]string{"behaviors[0].claims[0]", "is not a line range the tree carries", "law:claims"}},
		{"2 claims: a claim of a gone file", behavior(func(b *Behavior) { b.Claims = []string{"gone.go:1-2"} }), theWitness(),
			[]string{"behaviors[0].claims[0]", "is not a line range the tree carries"}},
		{"2 claims: a claim of an absolute path", behavior(func(b *Behavior) { b.Claims = []string{"/abs.go:1-2"} }), theWitness(),
			[]string{"behaviors[0].claims[0]", "is not a line range the tree carries"}},
		{"2 claims: a claim that starts at zero", behavior(func(b *Behavior) { b.Claims = []string{"src/tool.go:0-2"} }), theWitness(),
			[]string{"behaviors[0].claims[0]", "is not a line range the tree carries"}},
		{"2 ranges: a marker beside a broken range", Document{Work: "w",
			Markers: []Marker{{Lines: []string{"junk"}, Reason: "r"}}},
			lawsWitness(nil, nil, nil, map[string]string{"D1": "one"}, nil, nil),
			[]string{"not-a-requirement[0].lines[0]"}},
		{"2 ranges: a marker beside a missing line", Document{Work: "w",
			Markers: []Marker{{Lines: []string{"D9"}, Reason: "r"}}},
			lawsWitness(nil, nil, nil, map[string]string{"D1": "one"}, nil, nil),
			[]string{"not-a-requirement[0].lines[0]", "the demand stream carries no line D9"}},

		// Rule 3 — every new intent: one of the six outcomes, prose
		// non-empty, and the outcome's own carrier standing in the
		// same document.
		{"3 intents: an outcome outside the six", Document{Work: "w",
			Intents: []Intent{{Lines: []string{"D3"}, Outcome: "WORK", Prose: "p"}}},
			intentsWitness([]int{3}, nil, nil),
			[]string{"intents[0].outcome", "is not one of answer, execute, describe, work, mark, ask"}},
		{"3 intents: an empty prose", func() Document {
			d := workLineDoc()
			d.Intents[0].Prose = " "
			return d
		}(), intentsWitness([]int{3}, nil, nil),
			[]string{"intents[0].prose", "the prose is empty — retell the line in human words"}},
		{"3 intents: an answer without its text", Document{Work: "w",
			Intents: []Intent{{Lines: []string{"D3"}, Outcome: OutcomeAnswer, Prose: "p"}}},
			intentsWitness([]int{3}, nil, nil),
			[]string{"intents[0].text", "an answer intent carries its text"}},
		{"3 intents: text on a non-answer", Document{Work: "w",
			Markers: []Marker{{Lines: []string{"D3"}, Reason: "a heading"}},
			Intents: []Intent{{Lines: []string{"D3"}, Outcome: OutcomeMark, Prose: "p",
				Text: "not an answer"}}},
			intentsWitness([]int{3}, nil, nil),
			[]string{"intents[0].text", "text rides an answer intent only"}},
		{"3 intents: an execute without its run", Document{Work: "w",
			Intents: []Intent{{Lines: []string{"D3"}, Outcome: OutcomeExecute, Prose: "p"}}},
			intentsWitness([]int{3}, nil, nil),
			[]string{"intents[0].run", "an execute intent carries its command"}},
		{"3 intents: run on a non-execute", Document{Work: "w",
			Markers: []Marker{{Lines: []string{"D3"}, Reason: "a heading"}},
			Intents: []Intent{{Lines: []string{"D3"}, Outcome: OutcomeMark, Prose: "p",
				Run: Cmd{Line: "./tool"}}}},
			intentsWitness([]int{3}, nil, nil),
			[]string{"intents[0].run", "run rides an execute intent only"}},
		{"3 intents: work over no requirement", Document{Work: "w",
			Intents: []Intent{{Lines: []string{"D3"}, Outcome: OutcomeWork, Prose: "a telling"}}},
			intentsWitness([]int{3}, nil, nil),
			[]string{"intents[0]", "outcome work names no requirement on D3 — a work line is claimed by a requirement: one whose lines carry it, this same document"}},
		{"3 intents: describe over no requirement", Document{Work: "w",
			Intents: []Intent{{Lines: []string{"D3"}, Outcome: OutcomeDescribe, Prose: "a telling"}}},
			intentsWitness([]int{3}, nil, nil),
			[]string{"intents[0]", "outcome describe names no requirement on D3 — a describe line is claimed by a requirement: one whose lines carry it, this same document"}},
		{"3 intents: mark over no marker", Document{Work: "w",
			Intents: []Intent{{Lines: []string{"D3"}, Outcome: OutcomeMark, Prose: "talk, not an order"}}},
			intentsWitness([]int{3}, nil, nil),
			[]string{"intents[0]", "outcome mark names no marker on D3 — a mark line is claimed by a marker: a not-a-requirement entry with its lines and reason, this same document"}},
		{"3 intents: ask over no awaiting order", Document{Work: "w",
			Intents: []Intent{{Lines: []string{"D4"}, Outcome: OutcomeAsk, Prose: "the wish asks back"}}},
			intentsWitness([]int{4}, nil, nil),
			[]string{"intents[0]", "outcome ask names no awaiting order on D4: an open entry with state awaiting on its demand, this same document"}},
		{"3 intents: ask over a deferred order", Document{Work: "w",
			Intents: []Intent{{Lines: []string{"D4"}, Outcome: OutcomeAsk, Prose: "the wish asks back"}},
			Open:    []OpenOrder{{Demand: "D4", Question: "which unit?", State: StateDeferred, Reason: "later"}}},
			intentsWitness([]int{4}, nil, nil),
			[]string{"intents[0]", "outcome ask names no awaiting order on D4: an open entry with state awaiting on its demand, this same document"}},

		// Rule 4 — every new stream line declared by exactly one
		// intent; a declared line is not redeclared.
		{"4 declarations: tail lines without intents", Document{Work: "w"},
			intentsWitness([]int{3, 4}, nil, nil),
			[]string{"intents:", "the new demand lines D3, D4 carry no outcome — declare answer, execute, describe, work, mark or ask", "law:demand"}},
		{"4 declarations: a line declared in one submission twice", Document{Work: "w",
			Intents: []Intent{
				{Lines: []string{"D3"}, Outcome: OutcomeWork, Prose: "first telling"},
				{Lines: []string{"D3"}, Outcome: OutcomeWork, Prose: "second telling"},
			},
			Requirements: []Requirement{{ID: "R1", Lines: []string{"D3"}, Prose: "the unwinding"}}},
			intentsWitness([]int{3}, nil, nil),
			[]string{"intents[1].lines", "D3 already carries its outcome — a declared line is not redeclared"}},

		// Rule 5 — requirements: prose non-empty, refines one lookup
		// into the standing world.
		{"5 requirements: an empty prose", Document{Work: "w",
			Requirements: []Requirement{{ID: "R1", Lines: []string{"D1"}, Prose: " "}}},
			lawsWitness(nil, nil, nil, map[string]string{"D1": "one"}, nil, nil),
			[]string{"requirements[0].prose", "the prose is empty — unwind the demand lines in human words"}},
		{"5 requirements: a refinement of nothing standing", Document{Work: "w",
			Requirements: []Requirement{{ID: "R1", Lines: []string{"D1"}, Prose: "p", Refines: "R9"}}},
			lawsWitness(nil, map[string]bool{"R1": true}, nil, map[string]string{"D1": "one"}, nil, nil),
			[]string{"requirements[0].refines", `no requirement "R9" stands`}},

		// Rule 6 — a marker never stands on a requirement's line: a
		// line is either a demand or not, not both.
		{"6 markers: a marker on a requirement's line", Document{Work: "w",
			Requirements: []Requirement{{ID: "R1", Lines: []string{"D4"}, Prose: "the unwinding"}},
			Markers:      []Marker{{Lines: []string{"D4"}, Reason: "talk"}}},
			lawsWitness(nil, map[string]bool{"R1": true}, nil,
				map[string]string{"D4": "four"}, nil, nil),
			[]string{"not-a-requirement[0].lines", "D4 is claimed by R1 — a line is either a demand or not, not both"}},

		// Rule 7 — exclusions: a real hole, a reason, references that
		// resolve to a demand line or a law.
		{"7 exclusions: a malformed range", excluded(Exclusion{Range: "junk", Reason: "generated"}), holey,
			[]string{`uncovered-lines[0]: "junk" is not a line range path:first-last`}},
		{"7 exclusions: a range over no hole", excluded(Exclusion{Range: "src/tool.go:1-3", Reason: "generated"}),
			lawsWitness(nil, map[string]bool{"R1": true}, nil,
				map[string]string{"D1": "a heading", "D2": "do the work"},
				map[string]int{"src/tool.go": 3}, func(rel string, first, last int) bool {
					return false
				}),
			[]string{"uncovered-lines[0]", "the range names no uncovered line — a mark that marks nothing is refused"}},
		{"7 exclusions: an empty reason", excluded(Exclusion{Range: "src/tool.go:2-2", Reason: " "}), holey,
			[]string{"uncovered-lines[0].reason", "the reason is empty — a mark without a reason marks nothing"}},
		{"7 exclusions: a reason without references", excluded(Exclusion{Range: "src/tool.go:2-2", Reason: "generated"}), holey,
			[]string{"uncovered-lines[0].refs", "the reason must cite a demand line or a law"}},
		{"7 exclusions: a dangling reference", excluded(Exclusion{Range: "src/tool.go:2-2", Reason: "generated", Refs: []string{"D99"}}), holey,
			[]string{"uncovered-lines[0].refs[0]", `"D99" is neither a demand line nor a law`}},
		{"7 exclusions: a requirement id as reference", excluded(Exclusion{Range: "src/tool.go:2-2", Reason: "generated", Refs: []string{"R1"}}), holey,
			[]string{`"R1" is neither a demand line nor a law`}},

		// Rule 8 — behaviors: name and prose, real claims, covers of
		// the standing-after-merge world, rows, a pin, one run one
		// behavior.
		{"8 behaviors: an empty name", behavior(func(b *Behavior) { b.Name = "" }), theWitness(),
			[]string{"behaviors[0].name", "the name is empty"}},
		{"8 behaviors: an empty prose", behavior(func(b *Behavior) { b.Prose = " " }), theWitness(),
			[]string{"behaviors[0].prose", "the prose is empty — tell the behavior's story"}},
		{"8 behaviors: a dangling cover", behavior(func(b *Behavior) { b.Covers = []string{"R9"} }),
			lawsWitness(nil, nil, nil, nil, nil, nil),
			[]string{"behaviors[0].covers[0]", `no requirement "R9" stands after this submission`}},
		{"8 behaviors: no rows", behavior(func(b *Behavior) { b.Rows = nil }), theWitness(),
			[]string{"behaviors[0].rows", "a behavior carries at least one row — rows are all it proves"}},
		{"8 behaviors: no pinned row", behavior(func(b *Behavior) {
			b.Rows = []Row{{Run: Cmd{Line: "./tool"}, Expect: Expect{Out: "contains: ok"}}}
		}), theWitness(),
			[]string{"behaviors[0].rows", "no pinned row — rows are commentary without exact bytes; pin the output of one row: out or files with exact bytes (\"contains: \" prefixes a substring); rc alone pins nothing", "law:proof"}},
		{"8 behaviors: an empty run", behavior(func(b *Behavior) {
			b.Rows = []Row{{Expect: Expect{Out: "1"}}}
		}), theWitness(),
			[]string{"behaviors[0].rows[0].run", "the command is empty"}},
		{"8 behaviors: a run twice in the rows", behavior(func(b *Behavior) {
			b.Rows = []Row{
				{Run: Cmd{Line: "x"}, Expect: Expect{Out: "1"}},
				{Run: Cmd{Line: "x"}, Expect: Expect{Out: "2"}},
			}
		}), theWitness(),
			[]string{"behaviors[0].rows[1].run", `appears twice in the rows of "tool/x" — the run is the row's key`}},
		{"8 behaviors: a run already proving another behavior", Document{Work: "w", Behaviors: []Behavior{
			{ID: "a/b", Name: "n", Prose: "p", Rows: rowed},
			{ID: "c/d", Name: "m", Prose: "p", Rows: rowed},
		}}, nil,
			[]string{"behaviors[1].rows[0].run", "the run is already the proof of a/b — one command proves one behavior"}},
		{"8 behaviors: a run held by the standing corpus", behavior(func(b *Behavior) {}), &Resolvers{
			RunsInUse: func(runKey string) (string, bool) { return "tool/old", true },
		},
			[]string{"behaviors[0].rows[0].run", "the run is already the proof of tool/old — one command proves one behavior"}},
		{"8 behaviors: an empty expect", behavior(func(b *Behavior) {
			b.Rows = []Row{{Run: Cmd{Line: "x"}}}
		}), theWitness(),
			[]string{"behaviors[0].rows[0].expect", "the expect is empty — name rc, out or files"}},
		{"8 behaviors: an absolute expect path", behavior(func(b *Behavior) {
			b.Rows = []Row{{Run: Cmd{Line: "x"}, Expect: Expect{Files: map[string]string{"/abs": "x"}}}}
		}), theWitness(),
			[]string{"law:paths"}},

		// Rule 9 — the mandate: a replacement of standing work names
		// the ordering line and the replaced target.
		{"9 mandate: a replacement without a mandate", behavior(func(b *Behavior) { b.ID = "tool/x" }), standing,
			[]string{"mandate:", "the submission replaces standing behaviors without a mandate — name the demand lines that ordered it", "law:candor"}},
		{"9 mandate: a line the stream does not carry", func() Document {
			d := lawfulDoc()
			d.Mandate = []MandateRef{{Line: "D9", Target: "tool/x"}}
			return d
		}(), standing,
			[]string{"mandate[0].line", "the demand stream carries no line D9"}},
		{"9 mandate: a target that stands nowhere", func() Document {
			d := lawfulDoc()
			d.Mandate = []MandateRef{{Line: "D2", Target: "seg/gone"}}
			return d
		}(), standing,
			[]string{"mandate[0].target", `"seg/gone" stands nowhere — the mandate names what it permits to replace`}},
		{"9 mandate: a target this submission does not replace", func() Document {
			d := lawfulDoc()
			d.Mandate = []MandateRef{{Line: "D2", Target: "tool/old"}}
			return d
		}(), standing,
			[]string{"mandate[0].target", `"tool/old" is not replaced by this submission — the mandate names what it permits to replace`}},

		// Rule 10 — notes: a kind of the five, text, references that
		// resolve.
		{"10 notes: a kind outside the five", Document{Work: "w",
			Notes: []Note{{Kind: "musing", Text: "x"}}}, standing,
			[]string{"notes[0].kind", "the kind must be one of directive, decision, exclusion, lesson, work"}},
		{"10 notes: an empty text", Document{Work: "w",
			Notes: []Note{{Kind: "work", Text: " "}}}, standing,
			[]string{"notes[0].text", "the text is empty"}},
		{"10 notes: a dangling reference", Document{Work: "w",
			Notes: []Note{{Kind: "work", Text: "a note", Refs: []string{"junk"}}}}, standing,
			[]string{"notes[0].refs[0]", "resolves to nothing — a reference names a demand, a requirement, a behavior, a line range or a law"}},

		// Rule 11 — the done act: done: true and nothing else.
		{"11 done: work beside the flag", Document{Done: true, Work: "w"}, nil,
			[]string{"done", "a done document carries only done: true", "law:document"}},
		{"11 done: a behavior beside the flag", Document{Done: true,
			Behaviors: []Behavior{{ID: "a/b", Name: "n", Prose: "p", Rows: rowed}}}, nil,
			[]string{"done", "a done document carries only done: true"}},
		{"11 done: a remove beside the flag", Document{Done: true, Remove: []string{"a/b"}}, nil,
			[]string{"done", "a done document carries only done: true"}},

		// The shape laws beside the eleven — the narrative, the build,
		// the open orders, the remove keys.
		{"shape: an empty narrative", func() Document {
			d := lawfulDoc()
			d.Work = "  "
			return d
		}(), theWitness(),
			[]string{"work:", "the narrative is empty — tell what broke, what was fixed, what was decided", "law:candor"}},
		{"shape: an empty build", func() Document {
			d := lawfulDoc()
			empty := ""
			d.Build = &empty
			return d
		}(), theWitness(),
			[]string{"build", "the build command is empty — delete the line to keep the standing one"}},
		{"shape: an order on a line the stream does not carry", func() Document {
			d := lawfulDoc()
			d.Open = []OpenOrder{{Demand: "D9", Question: "stage two?", State: StateAwaiting}}
			return d
		}(), theWitness(),
			[]string{"open[0].demand", "the demand stream carries no line D9"}},
		{"shape: an order without its question", func() Document {
			d := lawfulDoc()
			d.Open = []OpenOrder{{Demand: "D1", Question: "  ", State: StateAwaiting}}
			return d
		}(), theWitness(),
			[]string{"open[0].question", "the question is empty — an order awaiting the human names its question"}},
		{"shape: an order of an unknown state", func() Document {
			d := lawfulDoc()
			d.Open = []OpenOrder{{Demand: "D1", Question: "later?", State: "forgotten"}}
			return d
		}(), theWitness(),
			[]string{"open[0].state", "the state must be awaiting or deferred"}},
		{"shape: a deferred order without its reason", func() Document {
			d := lawfulDoc()
			d.Open = []OpenOrder{{Demand: "D1", Question: "later?", State: StateDeferred}}
			return d
		}(), theWitness(),
			[]string{"open[0].reason", "a deferred order carries its reason"}},
		{"remove: a requirement that does not stand", Document{Work: "w", Remove: []string{"R1"}},
			lawsWitness(nil, nil, nil, nil, nil, nil),
			[]string{"remove[0]", `the corpus carries no requirement "R1"`}},
		{"remove: a behavior that does not stand", Document{Work: "w", Remove: []string{"tool/gone"}}, standing,
			[]string{"remove[0]", `the corpus carries no behavior "tool/gone"`}},
		{"remove: a line with no marker and no order", Document{Work: "w", Remove: []string{"D4"}},
			&Resolvers{MarkerStands: func(int) bool { return false }, OrderStands: func(int) bool { return false }},
			[]string{"remove[0]", "the corpus carries no marker or order on D4"}},
		{"remove: an exclusion that does not stand", Document{Work: "w", Remove: []string{"gen/out.go:1-9"}},
			&Resolvers{ExclusionStands: func(string) bool { return false }},
			[]string{"remove[0]", `the corpus carries no exclusion "gen/out.go:1-9"`}},
		{"remove: a key of neither family", Document{Work: "w", Remove: []string{"R1/b"}}, nil,
			[]string{"remove[0]", "the key must be a requirement id, a behavior id, a demand line D<n> or a line range path:first-last"}},
	}
	// The rows speak their refusals one by one; each names the field
	// and the fact the rule broke.
	for _, tc := range cases {
		got := refusalText(Validate(tc.d, tc.r))
		for _, want := range tc.want {
			if !strings.Contains(got, want) {
				t.Errorf("%s: missing %q in:\n%s", tc.rule, want, got)
			}
		}
	}
}

// intentsWitness — the resolvers of a stream with a journaled past
// (D1, D2) and an unjournaled tail: the intents' own facts. The tail
// names its line numbers; the standing requirements give their spans;
// the standing outcomes answer line by line.
func intentsWitness(tail []int, standing map[string][]int, outcomes map[string]string) *Resolvers {
	demand := map[string]string{"D1": "a heading", "D2": "do the work", "D3": "run the tests", "D4": "thanks", "D5": "five"}
	ids := make([]string, 0, len(standing))
	reqs := map[string]bool{}
	for id := range standing {
		ids = append(ids, id)
		reqs[id] = true
	}
	return &Resolvers{
		Stands:      func(string) bool { return false },
		Requirement: func(id string) bool { return reqs[id] },
		RequirementIDs: func() []string {
			return ids
		},
		RequirementLines: func(id string) ([]int, bool) {
			nums, ok := standing[id]
			return nums, ok
		},
		DemandLine: func(id string) (string, bool) {
			text, ok := demand[id]
			return text, ok
		},
		DemandsTailLines: func() []int {
			return tail
		},
		LineOutcome: func(id string) (string, bool) {
			o, ok := outcomes[id]
			return o, ok
		},
		FileLines: func(rel string) (int, bool) { return 3, true },
		Law:       func(id string) bool { return id == "law:done" },
		Hole:      func(rel string, first, last int) bool { return true },
	}
}

// workLineDoc — the smallest submission that declares the tail line
// D3 as work over one requirement.
func workLineDoc() Document {
	return Document{
		Work:    "the narrative",
		Intents: []Intent{{Lines: []string{"D3"}, Outcome: OutcomeWork, Prose: "a telling of the wish"}},
		Requirements: []Requirement{
			{ID: "R1", Lines: []string{"D3"}, Prose: "the unwinding"},
		},
	}
}

// TestLawfulFormsPass — every rule's lawful spelling in one breath:
// the smallest honest submission of each family passes whole.
func TestLawfulFormsPass(t *testing.T) {
	passes := []struct {
		name string
		d    Document
		r    *Resolvers
	}{
		{"the lawful document", lawfulDoc(), theWitness()},
		{"an answer and an execute with their carriers", Document{Work: "w",
			Intents: []Intent{
				{Lines: []string{"D3"}, Outcome: OutcomeAnswer, Prose: "a telling of the answer", Text: "the telling"},
				{Lines: []string{"D4"}, Outcome: OutcomeExecute, Prose: "a telling of the run", Run: Cmd{Line: "tool --version"}},
			}}, intentsWitness([]int{3, 4}, nil, nil)},
		{"work over a standing requirement's span", Document{Work: "w",
			Intents: []Intent{{Lines: []string{"D3"}, Outcome: OutcomeWork, Prose: "a telling"}}},
			intentsWitness([]int{3}, map[string][]int{"R9": {3}}, nil)},
		{"a mark over its marker", Document{Work: "w",
			Markers: []Marker{{Lines: []string{"D3"}, Reason: "talk, not an order"}},
			Intents: []Intent{{Lines: []string{"D3"}, Outcome: OutcomeMark, Prose: "talk, not an order"}}},
			intentsWitness([]int{3}, nil, nil)},
		{"an ask over its awaiting order", Document{Work: "w",
			Intents: []Intent{{Lines: []string{"D4"}, Outcome: OutcomeAsk, Prose: "the wish asks back"}},
			Open:    []OpenOrder{{Demand: "D4", Question: "which unit?", State: StateAwaiting}}},
			intentsWitness([]int{4}, nil, nil)},
		{"a tail declared whole", func() Document {
			d := workLineDoc()
			d.Markers = []Marker{{Lines: []string{"D4"}, Reason: "gratitude, no order"}}
			d.Intents = append(d.Intents, Intent{Lines: []string{"D4"}, Outcome: OutcomeMark, Prose: "gratitude, no order"})
			return d
		}(), intentsWitness([]int{3, 4}, nil, nil)},
		{"the lawful open orders", func() Document {
			d := lawfulDoc()
			d.Open = []OpenOrder{
				{Demand: "D1", Question: "stage two?", State: StateAwaiting},
				{Demand: "D2", Question: "later?", State: StateDeferred, Reason: "the human said later"},
			}
			return d
		}(), theWitness()},
		{"notes of every lawful reference kind", func() Document {
			d := lawfulDoc()
			d.Behaviors[0].ID = "tool/notes"
			for _, ref := range []string{"D1", "R1", "tool/x", "src/tool.go:1-3", "law:done"} {
				d.Notes = append(d.Notes, Note{Kind: "work", Text: "a note", Refs: []string{ref}})
			}
			return d
		}(), lawsWitness(map[string]bool{"tool/x": true}, map[string]bool{"R1": true},
			map[string]bool{"law:done": true},
			map[string]string{"D1": "a heading", "D2": "do the work"},
			map[string]int{"src/tool.go": 3}, nil)},
		{"an exclusion citing a demand line", func() Document {
			d := lawfulDoc()
			d.Exclusions = []Exclusion{{Range: "src/tool.go:2-2", Reason: "generated", Refs: []string{"D1"}}}
			return d
		}(), lawsWitness(nil, map[string]bool{"R1": true}, map[string]bool{"law:done": true},
			map[string]string{"D1": "a heading", "D2": "do the work"},
			map[string]int{"src/tool.go": 3}, func(rel string, first, last int) bool {
				return rel == "src/tool.go" && last >= 2 && first <= 2
			})},
		{"the mandated replacement", func() Document {
			d := lawfulDoc()
			d.Mandate = []MandateRef{{Line: "D2", Target: "tool/x"}}
			return d
		}(), lawsWitness(map[string]bool{"tool/x": true}, map[string]bool{"R1": true}, nil,
			map[string]string{"D1": "a heading", "D2": "do the work"},
			map[string]int{"src/tool.go": 3}, nil)},
		{"the lawful registry", Document{
			Work: "w", Mandate: []MandateRef{{Line: "D7", Target: "old/thing"}},
			Requirements: []Requirement{{ID: "R1", Lines: []string{"D1..D3, D7"}, Prose: "the unwinding"}},
			Markers:      []Marker{{Lines: []string{"D4..D6"}, Reason: "the framing lines are not demands"}},
			Behaviors: []Behavior{{ID: "a", Name: "n", Prose: "p", Covers: []string{"R1"},
				Rows: []Row{{Run: Cmd{Line: "x"}, Expect: Expect{Out: "1"}}}}},
			Remove: []string{"old/thing", "R2"},
		}, lawsWitness(map[string]bool{"old/thing": true, "R2": true}, map[string]bool{"R1": true},
			map[string]bool{"law:done": true},
			map[string]string{"D1": "one", "D2": "two", "D3": "three", "D4": "four", "D5": "five", "D6": "six", "D7": "seven"},
			map[string]int{"src/tool.go": 3}, nil)},
		{"the whole file's own span as claim", func() Document {
			d := lawfulDoc()
			d.Behaviors[0].Claims = []string{"src/tool.go:1-3"}
			return d
		}(), theWitness()},
		{"honest short prose", func() Document {
			d := lawfulDoc()
			d.Requirements[0].Prose = "ok"
			return d
		}(), theWitness()},
		{"the bare done act", Document{Done: true}, nil},
		// A nil resolver leaves a note's references unchecked: the
		// shape law stands down with its witness.
		{"a note over a nil resolver", Document{Work: "w",
			Notes: []Note{{Kind: "work", Text: "a note", Refs: []string{"anything"}}}}, nil},
	}
	for _, tc := range passes {
		if r := Validate(tc.d, tc.r); len(r) != 0 {
			t.Errorf("%s must pass clean: %v", tc.name, r)
		}
	}
}

// TestTheDeadWordLawsPass — the deleted word-matching laws turned
// inside out: the copy, the overlap and the cycle all pass the form
// now; the machine judges bytes and sets, never the sense of strings.
func TestTheDeadWordLawsPass(t *testing.T) {
	// A prose that copies the line's own words passes.
	d := workLineDoc()
	d.Intents[0].Prose = "run the tests"
	if r := Validate(d, intentsWitness([]int{3}, nil, nil)); len(r) != 0 {
		t.Fatalf("a copied prose passes the form: %v", r)
	}
	// Two requirements claiming one line without any refinement pass —
	// a line may serve several requirements.
	pair := Document{Work: "w", Requirements: []Requirement{
		{ID: "R1", Lines: []string{"D4"}, Prose: "the first"},
		{ID: "R2", Lines: []string{"D4"}, Prose: "the second"},
	}}
	witness := lawsWitness(nil, map[string]bool{"R1": true, "R2": true}, nil,
		map[string]string{"D1": "one", "D4": "four"}, nil, nil)
	if r := Validate(pair, witness); len(r) != 0 {
		t.Fatalf("a plain overlap passes: %v", r)
	}
	// A standing holder's overlap with a newcomer passes.
	holding := intentsWitness(nil, map[string][]int{"R9": {5}}, nil)
	fresh := Document{Work: "w",
		Requirements: []Requirement{{ID: "R10", Lines: []string{"D5"}, Prose: "the new one"}}}
	if r := Validate(fresh, holding); len(r) != 0 {
		t.Fatalf("a standing overlap passes: %v", r)
	}
	// A refinement cycle passes — the machine carries no chain
	// arithmetic.
	cycle := Document{Work: "w", Requirements: []Requirement{
		{ID: "R1", Lines: []string{"D1"}, Prose: "p", Refines: "R2"},
		{ID: "R2", Lines: []string{"D1"}, Prose: "p", Refines: "R1"},
	}}
	cyclic := lawsWitness(nil, map[string]bool{"R1": true, "R2": true}, nil,
		map[string]string{"D1": "one"}, nil, nil)
	if r := Validate(cycle, cyclic); len(r) != 0 {
		t.Fatalf("a cycle passes: %v", r)
	}
	// The nested chain with its overlaps passes.
	nested := Document{Work: "w", Requirements: []Requirement{
		{ID: "R1", Lines: []string{"D1..D3"}, Prose: "the whole"},
		{ID: "R2", Lines: []string{"D2..D3"}, Prose: "the middle", Refines: "R1"},
		{ID: "R3", Lines: []string{"D3"}, Prose: "the last", Refines: "R2"},
	}}
	chained := lawsWitness(nil, map[string]bool{"R1": true, "R2": true, "R3": true}, nil,
		map[string]string{"D1": "one", "D2": "two", "D3": "three"}, nil, nil)
	if r := Validate(nested, chained); len(r) != 0 {
		t.Fatalf("the nested chain passes: %v", r)
	}
}

// TestIntentNotRedeclared — the redeclaration is the only refusal it
// causes, named at the intent's own lines.
func TestIntentNotRedeclared(t *testing.T) {
	w := intentsWitness([]int{3}, nil, map[string]string{"D2": OutcomeWork})
	d := Document{Work: "the narrative",
		Intents: []Intent{
			{Lines: []string{"D2"}, Outcome: OutcomeWork, Prose: "a telling of the standing wish"},
			{Lines: []string{"D3"}, Outcome: OutcomeWork, Prose: "a telling of the wish"},
		},
		Requirements: []Requirement{
			{ID: "R0", Lines: []string{"D2"}, Prose: "the standing unwinding"},
			{ID: "R1", Lines: []string{"D3"}, Prose: "the unwinding"},
		}}
	r := Validate(d, w)
	if len(r) != 1 {
		t.Fatalf("the redeclaration must be the only refusal: %v", r)
	}
	for _, want := range []string{
		"intents[0].lines",
		"D2 already carries its outcome — a declared line is not redeclared",
		"law:demand",
	} {
		if !strings.Contains(r[0].Error(), want) {
			t.Fatalf("missing %q in:\n%s", want, r[0].Error())
		}
	}
}

// TestMarkerRequirementConflictRefused — the one overlap that stays:
// a marker on a requirement's line is a real contradiction, the only
// refusal of its submission.
func TestMarkerRequirementConflictRefused(t *testing.T) {
	witness := lawsWitness(nil, map[string]bool{"R1": true}, nil,
		map[string]string{"D4": "four"}, nil, nil)
	d := Document{Work: "w",
		Requirements: []Requirement{
			{ID: "R1", Lines: []string{"D4"}, Prose: "the unwinding"},
		},
		Markers: []Marker{{Lines: []string{"D4"}, Reason: "talk"}}}
	r := Validate(d, witness)
	if len(r) != 1 {
		t.Fatalf("the conflict must be the only refusal: %v", r)
	}
	if !strings.Contains(r[0].Error(), "not-a-requirement[0].lines") ||
		!strings.Contains(r[0].Error(), "D4 is claimed by R1 — a line is either a demand or not, not both") {
		t.Fatalf("the conflict names both sides:\n%s", r[0].Error())
	}
	// The marker alone on its line passes.
	d.Requirements = nil
	if r := Validate(d, witness); len(r) != 0 {
		t.Fatalf("the lone marker passes: %v", r)
	}
}

// TestPinnedRowRequired — the pin's own family: contains-only rows
// are commentary, exact file bytes pin as well, one contains-only
// file voids the pin.
func TestPinnedRowRequired(t *testing.T) {
	d := lawfulDoc()
	d.Behaviors[0].Rows = []Row{{Run: Cmd{Line: "./tool"}, Expect: Expect{Out: "contains: ok"}}}
	if r := Validate(d, theWitness()); len(r) != 1 {
		t.Fatalf("a contains-only row voids the pin: %v", r)
	}
	d.Behaviors[0].Rows = []Row{{Run: Cmd{Line: "./tool"}, Expect: Expect{Files: map[string]string{"out/x": "exact"}}}}
	if r := Validate(d, theWitness()); len(r) != 0 {
		t.Fatalf("exact file bytes are a pin: %v", r)
	}
	d.Behaviors[0].Rows[0].Expect.Files["out/y"] = "contains: partial"
	if r := Validate(d, theWitness()); len(r) != 1 {
		t.Fatalf("a contains-only file voids the pin: %v", r)
	}
}

// TestMandateRequiredOnReplacement — the mandate's own flow: a
// replacement or a removal of standing work without a mandate is
// refused, a mandated one passes, fresh work carries no debt.
func TestMandateRequiredOnReplacement(t *testing.T) {
	standingWitness := lawsWitness(
		map[string]bool{"tool/x": true},
		map[string]bool{"R1": true}, nil,
		map[string]string{"D1": "a heading", "D2": "do the work"},
		map[string]int{"src/tool.go": 3}, nil)
	d := lawfulDoc()
	if r := Validate(d, standingWitness); len(r) != 1 {
		t.Fatalf("a replacement without a mandate is the one refusal: %v", r)
	}
	// A removal of standing work is a replacement the same way.
	d = lawfulDoc()
	d.Behaviors = nil
	d.Remove = []string{"tool/x"}
	if got := refusalText(Validate(d, standingWitness)); !strings.Contains(got, "without a mandate") {
		t.Fatalf("a removal is a replacement of standing work:\n%s", got)
	}
	d = lawfulDoc()
	d.Mandate = []MandateRef{{Line: "D2", Target: "tool/x"}}
	if r := Validate(d, standingWitness); len(r) != 0 {
		t.Fatalf("the mandated replacement must pass: %v", r)
	}
	d = lawfulDoc()
	d.Behaviors[0].ID = "tool/fresh"
	if r := Validate(d, standingWitness); len(r) != 0 {
		t.Fatalf("new work carries no mandate debt: %v", r)
	}
}

// TestGateTailsOfTheOutcomeLaws — the gates' own tails: a mandate
// pair without a stands resolver keeps its line gate; a broken intent
// range is refused at its place; a nil resolver leaves the span map
// unchecked; the post-merge map reads a repeated id from the document
// and skips a spanless one.
func TestGateTailsOfTheOutcomeLaws(t *testing.T) {
	d := Document{Work: "w", Mandate: []MandateRef{{Line: "D9", Target: "x/y"}}}
	got := refusalText(Validate(d, nil))
	if !strings.Contains(got, "mandate[0].line: the demand stream carries no line D9") {
		t.Fatalf("the line gate speaks without resolvers:\n%s", got)
	}
	w := intentsWitness([]int{3}, nil, nil)
	d = Document{Work: "w",
		Intents: []Intent{{Lines: []string{"junk"}, Outcome: OutcomeWork, Prose: "p"}}}
	got = refusalText(Validate(d, w))
	if !strings.Contains(got, "intents[0].lines[0]") {
		t.Fatalf("the broken range is refused at its place:\n%s", got)
	}
	if r := Validate(Document{Work: "w",
		Intents:      []Intent{{Lines: []string{"D1"}, Outcome: OutcomeWork, Prose: "x"}},
		Requirements: []Requirement{{ID: "R1", Lines: []string{"D1"}, Prose: "p"}}}, nil); len(r) != 0 {
		t.Fatalf("a nil resolver leaves the intent's span unchecked: %v", r)
	}
	// The post-merge span map: a submitted id the preview repeats is
	// read from the document, an id without a span is skipped.
	d = Document{Work: "w",
		Requirements: []Requirement{{ID: "R1", Lines: []string{"D2"}, Prose: "p"}}}
	withRepeat := intentsWitness(nil, map[string][]int{"R1": {2}, "R8": nil}, nil)
	if r := Validate(d, withRepeat); len(r) != 0 {
		t.Fatalf("the repeated id reads its document span, the spanless id skips: %v", r)
	}
}

// TestRefResolverTails — the reference resolver's own tails: a
// parseable range resolves without line counts, and the marker key
// of a junk line set folds to its own spelling.
func TestRefResolverTails(t *testing.T) {
	if !refResolves("a.go:1-2", &Resolvers{}) {
		t.Fatal("a parseable range without line counts resolves")
	}
	if !refResolves("a.go:1-2", &Resolvers{Stands: func(string) bool { return false },
		Requirement: func(string) bool { return false }}) {
		t.Fatal("a range resolves over refusing id resolvers when no counts stand")
	}
	if key := MarkerKey(Marker{Lines: []string{"junk"}}); key != "junk" {
		t.Fatalf("a junk line set folds to its own spelling, got %q", key)
	}
}
