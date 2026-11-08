package doc

// The document's shape laws. Every refusal names the field it
// found, the law it broke and the fact itself — no advice, no
// interpretation: the author fixes the named place or the named
// intent. The laws are cited by the ids the contract prints, so a
// refusal and the contract always speak of the same rule.

import (
	"fmt"
	"strings"

	"github.com/neurophant/punchtape/internal/domain/fprint"
	"github.com/neurophant/punchtape/internal/domain/zone"
)

// The law ids of the contract the refusals cite.
const (
	LawDocument = "law:document"
	LawPaths    = "law:paths"
	LawDemand   = "law:demand"
	LawClaims   = "law:claims"
	LawProof    = "law:proof"
	LawCandor   = "law:candor"
)

// The open order's states.
const (
	StateAwaiting = "awaiting"
	StateDeferred = "deferred"
)

// noteKinds — the five kinds a note carries, no other spelling.
var noteKinds = map[string]bool{
	"directive": true, "decision": true, "exclusion": true, "lesson": true, "work": true,
}

// Refusal — one broken rule: the field, the law, the fact.
type Refusal struct {
	Where string
	Law   string
	Fact  string
}

// Error — the refusal as one line. The pointer carries the intent: a
// naked Refusal is data, *Refusal is the error.
func (r *Refusal) Error() string {
	return fmt.Sprintf("%s: %s (%s)", r.Where, r.Fact, r.Law)
}

// Resolvers — the standing facts a submission resolves against: the
// ids that stand in the corpus (behaviors and requirements — remove
// and mandate name them), the registry parts that stand (the remove
// list's marker, order and exclusion spellings name them), the
// requirements that stand after this submission's merge (covers and
// refines name them), the demand stream's line universe (ranges name
// it; the resolver answers the line's own text with its existence),
// the stream's unjournaled tail lines (the intents must declare
// them), the tree's line counts (claims name real lines), the law
// ids the contract prints (notes cite them) and the holes the tree
// carries (loud exclusions must mark one). A nil field skips its
// check; a nil Resolvers skips them all.
type Resolvers struct {
	Stands      func(id string) bool
	Requirement func(id string) bool
	DemandLine  func(id string) (string, bool)
	FileLines   func(rel string) (int, bool)
	Law         func(id string) bool
	Hole        func(rel string, first, last int) bool
	// MarkerStands — whether a standing marker claims the line;
	// OrderStands — whether a standing open order rides it;
	// ExclusionStands — whether a standing exclusion carries the
	// range. The remove list names standing things.
	MarkerStands    func(line int) bool
	OrderStands     func(line int) bool
	ExclusionStands func(rel string) bool
	// RequirementLines — one requirement's line span as it stands
	// after this submission's merge; RequirementIDs — the ids of every
	// requirement that stands after it, in registry order. Together
	// they give the post-merge span map the outcome and marker gates
	// read.
	RequirementLines func(id string) ([]int, bool)
	RequirementIDs   func() []string
	// DemandsTailLines — the line numbers of the stream's unjournaled
	// tail: the lines the intents must declare.
	DemandsTailLines func() []int
	// UniverseSize — the stream's own line count: the bound every
	// submitted D range must live within — a range past the last line
	// is refused before its numbers are generated.
	UniverseSize func() int
	// LineOutcome — the outcome a journaled line declared at the
	// submission that brought it, when one stands.
	LineOutcome func(id string) (string, bool)
	// RunsInUse — the standing holder of a run key after this
	// submission's merge: one command proves one behavior.
	RunsInUse func(runKey string) (holder string, inUse bool)
}

// Validate — the whole rule set against the standing corpus. A done
// document is exactly done: true — the completion act's own shape.
// An ordinary submission declares its intents (every new demand line
// one outcome of the six, retold in prose, its carrier in the same
// breath), carries a narrative, ids that are
// stable path-shaped and unique in the submission, requirement ids
// that are R-shaped and unique, markers with a reason, claims that
// are line ranges the tree carries, loud exclusions that mark real
// holes with a reason, open orders that ride standing demand lines,
// behaviors that each carry a pinned row and a non-empty prose, a
// mandate for replaced standing work, notes whose references
// resolve, and a remove list whose every key stands. Replacing a
// standing id is legal — the last submission wins, every version
// lives in the journal.
func Validate(d Document, r *Resolvers) []Refusal {
	if d.Done {
		return doneShapeGate(d)
	}
	var out []Refusal
	spans, spanOrder := postMergeSpans(d, r)
	out = append(out, intentsGate(d, r, spans)...)
	if strings.TrimSpace(d.Work) == "" {
		out = append(out, Refusal{Where: "work", Law: LawCandor,
			Fact: "the narrative is empty — tell what broke, what was fixed, what was decided"})
	}
	if d.Build != nil && strings.TrimSpace(*d.Build) == "" {
		out = append(out, refusalAt("build")("the build command is empty — delete the line to keep the standing one"))
	}
	out = append(out, listCeilings(d)...)
	out = append(out, notesGates(d.Notes, r)...)
	seenReqs := map[string]bool{}
	for i, req := range d.Requirements {
		ref := refusalAt(fmt.Sprintf("requirements[%d].id", i))
		if req.ID == "" {
			out = append(out, ref("the id is empty"))
		} else if !validRequirementID(req.ID) {
			out = append(out, ref("the id must be R followed by letters, digits and dots"))
		}
		if seenReqs[req.ID] {
			out = append(out, ref(fmt.Sprintf("the id %q appears twice in one submission", req.ID)))
		}
		seenReqs[req.ID] = true
		if len(req.Lines) == 0 {
			out = append(out, Refusal{Where: fmt.Sprintf("requirements[%d].lines", i), Law: LawDemand,
				Fact: "the list is empty — a requirement claims at least one line"})
		}
		out = append(out, resolveRanges(fmt.Sprintf("requirements[%d].lines", i), req.Lines, r)...)
		if strings.TrimSpace(req.Prose) == "" {
			out = append(out, Refusal{Where: fmt.Sprintf("requirements[%d].prose", i), Law: LawDemand,
				Fact: "the prose is empty — unwind the demand lines in human words"})
		}
		if req.Refines != "" && r != nil && r.Requirement != nil && !r.Requirement(req.Refines) {
			out = append(out, Refusal{Where: fmt.Sprintf("requirements[%d].refines", i), Law: LawCandor,
				Fact: fmt.Sprintf("no requirement %q stands", req.Refines)})
		}
	}
	for i, marker := range d.Markers {
		if strings.TrimSpace(marker.Reason) == "" {
			out = append(out, refusalAt(fmt.Sprintf("not-a-requirement[%d].reason", i))("the reason is empty — a mark without a reason marks nothing"))
		}
		if len(marker.Lines) == 0 {
			out = append(out, Refusal{Where: fmt.Sprintf("not-a-requirement[%d].lines", i), Law: LawDemand,
				Fact: "the list is empty — a marker claims at least one line"})
		}
		markerRefusals := resolveRanges(fmt.Sprintf("not-a-requirement[%d].lines", i), marker.Lines, r)
		out = append(out, markerRefusals...)
		// A line is either a demand or not, never both: a marker on a
		// line a post-merge requirement claims is a quiet contradiction.
		if len(markerRefusals) == 0 {
			if nums, err := ParseRangesWithin(marker.Lines, withinOf(r)); err == nil {
				for _, n := range nums {
					if holder, claimed := spanHolder(spans, spanOrder, n); claimed {
						out = append(out, Refusal{Where: fmt.Sprintf("not-a-requirement[%d].lines", i), Law: LawDemand,
							Fact: fmt.Sprintf("%s is claimed by %s — a line is either a demand or not, not both", DemandID(n), holder)})
						break
					}
				}
			}
		}
	}
	for i, ex := range d.Exclusions {
		if strings.TrimSpace(ex.Reason) == "" {
			out = append(out, Refusal{Where: fmt.Sprintf("uncovered-lines[%d].reason", i), Law: LawCandor,
				Fact: "the reason is empty — a mark without a reason marks nothing"})
		}
		rel, first, last, err := fprint.ParseLineRange(ex.Range)
		if err != nil {
			out = append(out, Refusal{Where: fmt.Sprintf("uncovered-lines[%d]", i), Law: LawClaims,
				Fact: fmt.Sprintf("%q is not a line range path:first-last", ex.Range)})
			continue
		}
		if r != nil && r.Hole != nil && !r.Hole(rel, first, last) {
			out = append(out, Refusal{Where: fmt.Sprintf("uncovered-lines[%d]", i), Law: LawCandor,
				Fact: "the range names no uncovered line — a mark that marks nothing is refused"})
		}
		if len(ex.Refs) == 0 {
			out = append(out, Refusal{Where: fmt.Sprintf("uncovered-lines[%d].refs", i), Law: LawCandor,
				Fact: "the reason must cite a demand line or a law"})
		}
		for j, ref := range ex.Refs {
			if r != nil && !demandStands(ref, r) && !(r.Law != nil && r.Law(ref)) {
				out = append(out, Refusal{Where: fmt.Sprintf("uncovered-lines[%d].refs[%d]", i, j), Law: LawCandor,
					Fact: fmt.Sprintf("%q is neither a demand line nor a law", ref)})
			}
		}
	}
	for i, order := range d.Open {
		if !demandStands(order.Demand, r) {
			out = append(out, Refusal{Where: fmt.Sprintf("open[%d].demand", i), Law: LawCandor,
				Fact: fmt.Sprintf("the demand stream carries no line %s", order.Demand)})
		}
		if strings.TrimSpace(order.Question) == "" {
			out = append(out, Refusal{Where: fmt.Sprintf("open[%d].question", i), Law: LawCandor,
				Fact: "the question is empty — an order awaiting the human names its question"})
		}
		if order.State != StateAwaiting && order.State != StateDeferred {
			out = append(out, Refusal{Where: fmt.Sprintf("open[%d].state", i), Law: LawCandor,
				Fact: "the state must be awaiting or deferred"})
		}
		if order.State == StateDeferred && strings.TrimSpace(order.Reason) == "" {
			out = append(out, Refusal{Where: fmt.Sprintf("open[%d].reason", i), Law: LawCandor,
				Fact: "a deferred order carries its reason"})
		}
	}
	out = append(out, mandateGates(d, r)...)
	runHolders := map[string]string{}
	seen := map[string]bool{}
	for i, b := range d.Behaviors {
		ref := refusalAt(fmt.Sprintf("behaviors[%d].id", i))
		if b.ID == "" {
			out = append(out, ref("the id is empty"))
			continue
		}
		if !validID(b.ID) {
			out = append(out, ref("the id must be [a-z0-9-] segments joined by '/'"))
		}
		if seen[b.ID] {
			out = append(out, ref(fmt.Sprintf("the id %q appears twice in one submission", b.ID)))
		}
		seen[b.ID] = true
		if b.Name == "" {
			out = append(out, refusalAt(fmt.Sprintf("behaviors[%d].name", i))("the name is empty"))
		}
		if strings.TrimSpace(b.Prose) == "" {
			out = append(out, Refusal{Where: fmt.Sprintf("behaviors[%d].prose", i), Law: LawCandor,
				Fact: "the prose is empty — tell the behavior's story"})
		}
		if len(b.Claims) > ClaimsPerBehavior {
			out = append(out, Refusal{Where: fmt.Sprintf("behaviors[%d].claims", i), Law: LawDocument,
				Fact: fmt.Sprintf("the behavior carries %d claims — the ceiling is %d", len(b.Claims), ClaimsPerBehavior)})
		}
		if len(b.Rows) > RowsPerBehavior {
			out = append(out, Refusal{Where: fmt.Sprintf("behaviors[%d].rows", i), Law: LawDocument,
				Fact: fmt.Sprintf("the behavior carries %d rows — the ceiling is %d", len(b.Rows), RowsPerBehavior)})
		}
		for j, claim := range b.Claims {
			if !ClaimCarries(claim, fileLinesOf(r)) {
				out = append(out, Refusal{Where: fmt.Sprintf("behaviors[%d].claims[%d]", i, j), Law: LawClaims,
					Fact: fmt.Sprintf("%q is not a line range the tree carries", claim)})
			}
		}
		for ci, cover := range b.Covers {
			if r != nil && r.Requirement != nil && !r.Requirement(cover) {
				out = append(out, refusalAt(fmt.Sprintf("behaviors[%d].covers[%d]", i, ci))(
					fmt.Sprintf("no requirement %q stands after this submission", cover)))
			}
		}
		if len(b.Rows) == 0 {
			out = append(out, refusalAt(fmt.Sprintf("behaviors[%d].rows", i))("a behavior carries at least one row — rows are all it proves"))
		}
		if !hasPinnedRow(b.Rows) {
			out = append(out, Refusal{Where: fmt.Sprintf("behaviors[%d].rows", i), Law: LawProof,
				Fact: "no pinned row — rows are commentary without exact bytes; pin the output of one row: out or files with exact bytes (\"contains: \" prefixes a substring); rc alone pins nothing"})
		}
		runs := map[string]bool{}
		for j, row := range b.Rows {
			if row.Run.Line == "" && len(row.Run.Argv) == 0 {
				out = append(out, refusalAt(fmt.Sprintf("behaviors[%d].rows[%d].run", i, j))("the command is empty"))
			}
			if holder, seen := runHolders[row.Run.Key()]; seen && holder != b.ID {
				out = append(out, Refusal{Where: fmt.Sprintf("behaviors[%d].rows[%d].run", i, j), Law: LawProof,
					Fact: fmt.Sprintf("the run is already the proof of %s — one command proves one behavior", holder)})
			} else if !seen {
				runHolders[row.Run.Key()] = b.ID
			}
			if r != nil && r.RunsInUse != nil {
				if holder, inUse := r.RunsInUse(row.Run.Key()); inUse && holder != b.ID {
					out = append(out, Refusal{Where: fmt.Sprintf("behaviors[%d].rows[%d].run", i, j), Law: LawProof,
						Fact: fmt.Sprintf("the run is already the proof of %s — one command proves one behavior", holder)})
				}
			}
			if row.Expect.RC == nil && row.Expect.Out == "" && len(row.Expect.Files) == 0 {
				out = append(out, refusalAt(fmt.Sprintf("behaviors[%d].rows[%d].expect", i, j))("the expect is empty — name rc, out or files"))
			}
			if len(row.Expect.Files) > FilesPerExpect {
				out = append(out, Refusal{Where: fmt.Sprintf("behaviors[%d].rows[%d].expect.files", i, j), Law: LawDocument,
					Fact: fmt.Sprintf("the row expects %d files — the ceiling is %d", len(row.Expect.Files), FilesPerExpect)})
			}
			if key := row.Run.Key(); runs[key] {
				out = append(out, refusalAt(fmt.Sprintf("behaviors[%d].rows[%d].run", i, j))(
					fmt.Sprintf("the run %s appears twice in the rows of %q — the run is the row's key", spellRun(row.Run), b.ID)))
			} else {
				runs[key] = true
			}
			for path := range row.Expect.Files {
				if err := zone.ValidateRel(path); err != nil {
					out = append(out, Refusal{
						Where: fmt.Sprintf("behaviors[%d].rows[%d].expect.files[%s]", i, j, path),
						Law:   LawPaths,
						Fact:  err.Error(),
					})
				}
			}
		}
	}
	out = append(out, removeGates(d, r)...)
	return out
}

// doneShapeGate — the completion act's own law: done: true and
// nothing else. Any other field beside the flag is refused.
func doneShapeGate(d Document) []Refusal {
	empty := d.Work == "" && d.Build == nil && len(d.Intents) == 0 &&
		len(d.Remove) == 0 && len(d.Requirements) == 0 && len(d.Markers) == 0 &&
		len(d.Exclusions) == 0 && len(d.Open) == 0 && len(d.Behaviors) == 0 &&
		len(d.Notes) == 0 && len(d.Mandate) == 0
	if empty {
		return nil
	}
	return []Refusal{{Where: "done", Law: LawDocument,
		Fact: "a done document carries only done: true"}}
}

// removeGates — the remove list's own law: every key of the five
// spellings names a standing thing. The D<n> spelling resolves by
// context — the marker standing on the line first, then the order
// riding it.
func removeGates(d Document, r *Resolvers) []Refusal {
	var out []Refusal
	for i, key := range d.Remove {
		ref := refusalAt(fmt.Sprintf("remove[%d]", i))
		switch ParseRemoveKey(key) {
		case RemoveRequirement:
			if r != nil && r.Stands != nil && !r.Stands(key) {
				out = append(out, ref(fmt.Sprintf("the corpus carries no requirement %q", key)))
			}
		case RemoveBehavior:
			if r != nil && r.Stands != nil && !r.Stands(key) {
				out = append(out, ref(fmt.Sprintf("the corpus carries no behavior %q", key)))
			}
		case RemoveMarkerOrOrder:
			if r == nil || r.MarkerStands == nil || r.OrderStands == nil {
				continue
			}
			// The key's own grammar check has spoken: the parse is
			// total and names exactly one line.
			nums, _ := ParseRangesWithin([]string{key}, withinOf(r))
			if !r.MarkerStands(nums[0]) && !r.OrderStands(nums[0]) {
				out = append(out, ref(fmt.Sprintf("the corpus carries no marker or order on %s", key)))
			}
		case RemoveExclusion:
			if r != nil && r.ExclusionStands != nil && !r.ExclusionStands(key) {
				out = append(out, ref(fmt.Sprintf("the corpus carries no exclusion %q", key)))
			}
		default:
			out = append(out, ref("the key must be a requirement id, a behavior id, a demand line D<n> or a line range path:first-last"))
		}
	}
	return out
}

// listCeilings — the submission's registry lists against their one
// ceiling: a list past it is refused by its own name with its count
// and the bound.
func listCeilings(d Document) []Refusal {
	var out []Refusal
	lists := []struct {
		where string
		n     int
	}{
		{"intents", len(d.Intents)},
		{"notes", len(d.Notes)},
		{"requirements", len(d.Requirements)},
		{"not-a-requirement", len(d.Markers)},
		{"open", len(d.Open)},
		{"uncovered-lines", len(d.Exclusions)},
	}
	for _, l := range lists {
		if l.n > ListPerSubmission {
			out = append(out, Refusal{Where: l.where, Law: LawDocument,
				Fact: fmt.Sprintf("the submission carries %d %s entries — the ceiling is %d", l.n, l.where, ListPerSubmission)})
		}
	}
	return out
}

// notesGates — the note's own shape: a kind of the five, text, and
// references that resolve against the standing world.
func notesGates(notes []Note, r *Resolvers) []Refusal {
	var out []Refusal
	for i, note := range notes {
		if !noteKinds[note.Kind] {
			out = append(out, Refusal{Where: fmt.Sprintf("notes[%d].kind", i), Law: LawCandor,
				Fact: "the kind must be one of directive, decision, exclusion, lesson, work"})
		}
		if strings.TrimSpace(note.Text) == "" {
			out = append(out, Refusal{Where: fmt.Sprintf("notes[%d].text", i), Law: LawCandor,
				Fact: "the text is empty"})
		}
		for j, ref := range note.Refs {
			if !refResolves(ref, r) {
				out = append(out, Refusal{Where: fmt.Sprintf("notes[%d].refs[%d]", i, j), Law: LawCandor,
					Fact: fmt.Sprintf("%q resolves to nothing — a reference names a demand, a requirement, a behavior, a line range or a law", ref)})
			}
		}
	}
	return out
}

// refResolves — one reference against the standing world: a demand
// line of the stream, a standing requirement id, a standing
// behavior id, a parseable line range the tree carries, or a law id
// of the contract.
func refResolves(ref string, r *Resolvers) bool {
	if r == nil {
		return true
	}
	if r.DemandLine != nil {
		if _, ok := r.DemandLine(ref); ok {
			return true
		}
	}
	if r.Requirement != nil && r.Requirement(ref) {
		return true
	}
	if r.Stands != nil && r.Stands(ref) {
		return true
	}
	if rel, _, _, err := fprint.ParseLineRange(ref); err == nil {
		if r.FileLines == nil {
			return true
		}
		_, ok := r.FileLines(rel)
		return ok
	}
	return r.Law != nil && r.Law(ref)
}

// mandateGates — the replacement's own law: a submission that
// replaces or retires standing work names, for each permission, the
// demand line that ordered the change and the standing thing it
// permits to replace. The line stands in the stream; the target
// stands in the corpus and is replaced by this very submission.
func mandateGates(d Document, r *Resolvers) []Refusal {
	var out []Refusal
	if replacesStanding(d, r) && len(d.Mandate) == 0 {
		out = append(out, Refusal{Where: "mandate", Law: LawCandor,
			Fact: "the submission replaces standing behaviors without a mandate — name the demand lines that ordered it"})
	}
	replaced := replacedByThis(d, r)
	for i, ref := range d.Mandate {
		if !demandStands(ref.Line, r) {
			out = append(out, Refusal{Where: fmt.Sprintf("mandate[%d].line", i), Law: LawCandor,
				Fact: fmt.Sprintf("the demand stream carries no line %s", ref.Line)})
		}
		if r == nil || r.Stands == nil {
			continue
		}
		if !r.Stands(ref.Target) {
			out = append(out, Refusal{Where: fmt.Sprintf("mandate[%d].target", i), Law: LawCandor,
				Fact: fmt.Sprintf("%q stands nowhere — the mandate names what it permits to replace", ref.Target)})
			continue
		}
		if !replaced[ref.Target] {
			out = append(out, Refusal{Where: fmt.Sprintf("mandate[%d].target", i), Law: LawCandor,
				Fact: fmt.Sprintf("%q is not replaced by this submission — the mandate names what it permits to replace", ref.Target)})
		}
	}
	return out
}

// replacedByThis — the standing ids this submission replaces: listed
// behaviors whose id stands, and removes of standing ids.
func replacedByThis(d Document, r *Resolvers) map[string]bool {
	out := map[string]bool{}
	if r == nil || r.Stands == nil {
		return out
	}
	for _, b := range d.Behaviors {
		if r.Stands(b.ID) {
			out[b.ID] = true
		}
	}
	for _, id := range d.Remove {
		if validID(id) && r.Stands(id) {
			out[id] = true
		}
		if validRequirementID(id) && r.Stands(id) {
			out[id] = true
		}
	}
	return out
}

// withinOf — the stream bound the resolvers carry, zero when absent:
// no stream to bound by, the plain grammar rules.
func withinOf(r *Resolvers) int {
	if r == nil || r.UniverseSize == nil {
		return 0
	}
	return r.UniverseSize()
}

// postMergeSpans — every requirement's line span as it stands after
// this submission's merge: the submission's own requirements plus
// the standing ones the resolvers name, in first-seen order.
func postMergeSpans(d Document, r *Resolvers) (map[string]map[int]bool, []string) {
	spans := map[string]map[int]bool{}
	var order []string
	put := func(id string, nums []int) {
		set := map[int]bool{}
		for _, n := range nums {
			set[n] = true
		}
		if _, seen := spans[id]; !seen {
			order = append(order, id)
		}
		spans[id] = set
	}
	for _, req := range d.Requirements {
		if nums, err := ParseRangesWithin(req.Lines, withinOf(r)); err == nil {
			put(req.ID, nums)
		}
	}
	if r != nil && r.RequirementIDs != nil && r.RequirementLines != nil {
		for _, id := range r.RequirementIDs() {
			if _, inDoc := spans[id]; inDoc {
				continue
			}
			if nums, ok := r.RequirementLines(id); ok {
				put(id, nums)
			}
		}
	}
	return spans, order
}

// spanHolder — the first requirement in span order that claims the
// line, for the marker conflict's own naming.
func spanHolder(spans map[string]map[int]bool, order []string, n int) (string, bool) {
	for _, id := range order {
		if spans[id][n] {
			return id, true
		}
	}
	return "", false
}

// intentsGate — the outcome law of the stream's new lines: every
// line the stream's unjournaled tail carries is declared by exactly
// one intent of the submission that meets it — one of the six
// outcomes, retold in prose — and the declaration's carrier stands
// in the same breath: work and describe over a requirement's span,
// mark over a marker, ask over an awaiting order; answer and execute
// carry their own carrier inside the intent — the text and the
// command.
func intentsGate(d Document, r *Resolvers, spans map[string]map[int]bool) []Refusal {
	var out []Refusal
	maxLine := withinOf(r)
	declared := map[int]bool{}
	declaredHere := map[int]bool{}
	for i, in := range d.Intents {
		if len(in.Lines) == 0 {
			out = append(out, Refusal{Where: fmt.Sprintf("intents[%d].lines", i), Law: LawDemand,
				Fact: "the list is empty — an intent declares at least one line"})
		}
		lineRefusals := resolveRanges(fmt.Sprintf("intents[%d].lines", i), in.Lines, r)
		out = append(out, lineRefusals...)
		nums, err := ParseRangesWithin(in.Lines, withinOf(r))
		if err != nil {
			continue
		}
		for _, n := range nums {
			declared[n] = true
		}
		if !validOutcome(in.Outcome) {
			out = append(out, Refusal{Where: fmt.Sprintf("intents[%d].outcome", i), Law: LawDemand,
				Fact: fmt.Sprintf("%q is not one of answer, execute, describe, work, mark, ask", in.Outcome)})
			continue
		}
		proseRefusal := Refusal{Where: fmt.Sprintf("intents[%d].prose", i), Law: LawDemand,
			Fact: "the prose is empty — retell the line in human words"}
		if strings.TrimSpace(in.Prose) == "" {
			out = append(out, proseRefusal)
		}
		out = append(out, carrierShapeRefusals(i, in)...)
		for _, n := range nums {
			// A declared line is not redeclared: the refusal fires on
			// a line that already carries its outcome — standing in
			// the stream or declared earlier in this same submission.
			// A line the birth journaled carries nothing yet; its
			// first declaration may come with any later submission.
			already := declaredHere[n]
			if !already && r != nil && r.LineOutcome != nil {
				if _, ok := r.LineOutcome(DemandID(n)); ok {
					already = true
				}
			}
			if already {
				out = append(out, Refusal{Where: fmt.Sprintf("intents[%d].lines", i), Law: LawDemand,
					Fact: fmt.Sprintf("%s already carries its outcome — a declared line is not redeclared", DemandID(n))})
				break
			}
		}
		for _, n := range nums {
			declaredHere[n] = true
		}
		out = append(out, carrierRefusals(i, in, nums, d, spans, maxLine)...)
	}
	if r != nil && r.DemandsTailLines != nil {
		var missing []int
		for _, n := range r.DemandsTailLines() {
			if !declared[n] {
				missing = append(missing, n)
			}
		}
		if len(missing) > 0 {
			out = append(out, Refusal{Where: "intents", Law: LawDemand,
				Fact: fmt.Sprintf("the new demand lines %s carry no outcome — declare answer, execute, describe, work, mark or ask", SpellRanges(missing))})
		}
	}
	return out
}

// ValidOutcome — the six spellings of the protocol's own
// vocabulary.
func ValidOutcome(s string) bool {
	switch s {
	case OutcomeAnswer, OutcomeExecute, OutcomeDescribe, OutcomeWork, OutcomeMark, OutcomeAsk:
		return true
	}
	return false
}

// validOutcome — the six spellings, one spelling everywhere.
func validOutcome(s string) bool { return ValidOutcome(s) }

// carrierShapeRefusals — the answer and execute intents' own
// carriers: answer carries its text, execute its command — and no
// other outcome carries either.
func carrierShapeRefusals(i int, in Intent) []Refusal {
	var out []Refusal
	switch in.Outcome {
	case OutcomeAnswer:
		if strings.TrimSpace(in.Text) == "" {
			out = append(out, Refusal{Where: fmt.Sprintf("intents[%d].text", i), Law: LawDemand,
				Fact: "an answer intent carries its text — the answering words ride the intent"})
		}
	case OutcomeExecute:
		if in.Run.Line == "" && in.Run.Argv == nil {
			out = append(out, Refusal{Where: fmt.Sprintf("intents[%d].run", i), Law: LawDemand,
				Fact: "an execute intent carries its command"})
		}
	}
	if in.Outcome != OutcomeAnswer && strings.TrimSpace(in.Text) != "" {
		out = append(out, Refusal{Where: fmt.Sprintf("intents[%d].text", i), Law: LawDemand,
			Fact: "text rides an answer intent only"})
	}
	if in.Outcome != OutcomeExecute && !(in.Run.Line == "" && in.Run.Argv == nil) {
		out = append(out, Refusal{Where: fmt.Sprintf("intents[%d].run", i), Law: LawDemand,
			Fact: "run rides an execute intent only"})
	}
	return out
}

// carrierRefusals — the declared outcome's carrier must stand in the
// same submission: work and describe over a requirement's span of
// the post-merge map, mark over a marker of the same document, ask
// over an awaiting order of the same document.
func carrierRefusals(i int, in Intent, nums []int, d Document, spans map[string]map[int]bool, maxLine int) []Refusal {
	requirementOn := func(n int) bool {
		for _, set := range spans {
			if set[n] {
				return true
			}
		}
		return false
	}
	markers := inDocRanges(d.Markers, maxLine)
	awaiting := map[int]bool{}
	for _, order := range d.Open {
		if order.State != StateAwaiting {
			continue
		}
		if ns, err := ParseRangesWithin([]string{order.Demand}, maxLine); err == nil {
			for _, n := range ns {
				awaiting[n] = true
			}
		}
	}
	missing := func(carried func(n int) bool) *int {
		for _, n := range nums {
			if !carried(n) {
				return &n
			}
		}
		return nil
	}
	var fact string
	switch in.Outcome {
	case OutcomeWork, OutcomeDescribe:
		if n := missing(requirementOn); n != nil {
			fact = fmt.Sprintf("outcome %s names no requirement on %s — a %s line is claimed by a requirement: one whose lines carry it, this same document", in.Outcome, DemandID(*n), in.Outcome)
		}
	case OutcomeMark:
		if n := missing(func(n int) bool { return markers[n] }); n != nil {
			fact = fmt.Sprintf("outcome mark names no marker on %s — a mark line is claimed by a marker: a not-a-requirement entry with its lines and reason, this same document", DemandID(*n))
		}
	case OutcomeAsk:
		if n := missing(func(n int) bool { return awaiting[n] }); n != nil {
			fact = fmt.Sprintf("outcome ask names no awaiting order on %s: an open entry with state awaiting on its demand, this same document", DemandID(*n))
		}
	}
	if fact == "" {
		return nil
	}
	return []Refusal{{Where: fmt.Sprintf("intents[%d]", i), Law: LawDemand, Fact: fact}}
}

// inDocRanges — the line sets of the markers the submission itself
// carries, keyed by their own range lists, bounded by the stream.
func inDocRanges(records []Marker, maxLine int) map[int]bool {
	out := map[int]bool{}
	for _, rec := range records {
		if ns, err := ParseRangesWithin(rec.Lines, maxLine); err == nil {
			for _, n := range ns {
				out[n] = true
			}
		}
	}
	return out
}

// replacesStanding — whether the submission touches standing work:
// a listed behavior whose id stands, or a remove of a standing id.
func replacesStanding(d Document, r *Resolvers) bool {
	if r == nil || r.Stands == nil {
		return false
	}
	for _, b := range d.Behaviors {
		if r.Stands(b.ID) {
			return true
		}
	}
	for _, id := range d.Remove {
		if r.Stands(id) {
			return true
		}
	}
	return false
}

// Pinned — whether one expectation pins exact bytes: an out that is
// neither empty nor a contains-prefix, or files where every value
// pins exact bytes. A pinned row is the behavior's proof; everything
// else is commentary.
func Pinned(e Expect) bool {
	if e.Out != "" && !strings.HasPrefix(e.Out, ContainsPrefix) {
		return true
	}
	if len(e.Files) == 0 {
		return false
	}
	for _, want := range e.Files {
		if strings.HasPrefix(want, ContainsPrefix) {
			return false
		}
	}
	return true
}

// hasPinnedRow — whether any row of the behavior pins exact bytes.
func hasPinnedRow(rows []Row) bool {
	for _, row := range rows {
		if Pinned(row.Expect) {
			return true
		}
	}
	return false
}

// demandStands — whether a demand line exists in the stream through
// the resolver; an absent resolver answers nothing.
func demandStands(id string, r *Resolvers) bool {
	if r == nil || r.DemandLine == nil {
		return false
	}
	_, ok := r.DemandLine(id)
	return ok
}

// fileLinesOf — the resolver's line-count function, nil when absent.
func fileLinesOf(r *Resolvers) func(rel string) (int, bool) {
	if r == nil {
		return nil
	}
	return r.FileLines
}

// resolveRanges — one range list's law: every spelling parses in the
// D grammar, stays within the stream's own count, and every line it
// names exists in the stream's universe.
func resolveRanges(where string, ranges []string, r *Resolvers) []Refusal {
	var out []Refusal
	for j, list := range ranges {
		nums, err := ParseRangesWithin([]string{list}, withinOf(r))
		if err != nil {
			out = append(out, refusalAt(fmt.Sprintf("%s[%d]", where, j))(err.Error()))
			continue
		}
		if r == nil || r.DemandLine == nil {
			continue
		}
		var missing []int
		for _, n := range nums {
			if _, ok := r.DemandLine(DemandID(n)); !ok {
				missing = append(missing, n)
			}
		}
		if len(missing) > 0 {
			out = append(out, refusalAt(fmt.Sprintf("%s[%d]", where, j))(
				fmt.Sprintf("the demand stream carries no line %s", SpellRanges(missing))))
		}
	}
	return out
}

// validID — the id alphabet: lowercase segments of letters, digits
// and dashes joined by single slashes.
func validID(id string) bool {
	if strings.HasPrefix(id, "/") || strings.HasSuffix(id, "/") || strings.Contains(id, "//") {
		return false
	}
	for _, seg := range strings.Split(id, "/") {
		if seg == "" {
			return false
		}
		for _, c := range seg {
			switch {
			case c >= 'a' && c <= 'z', c >= '0' && c <= '9', c == '-':
			default:
				return false
			}
		}
	}
	return true
}

// validRequirementID — the requirement id alphabet: an upper R
// followed by letters, digits and dots. The two id families share no
// spelling — a behavior id is lowercase and slash-segmented.
func validRequirementID(id string) bool {
	if !strings.HasPrefix(id, "R") {
		return false
	}
	for _, c := range id[1:] {
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9', c == '.':
		default:
			return false
		}
	}
	return true
}

// refusalAt — the refusal builder for one field.
func refusalAt(where string) func(fact string) Refusal {
	return func(fact string) Refusal {
		return Refusal{Where: where, Law: LawDocument, Fact: fact}
	}
}

// spellRun — the command as one readable word for a refusal's fact.
func spellRun(c Cmd) string {
	if c.Argv != nil {
		return fmt.Sprintf("%q", strings.Join(c.Argv, " "))
	}
	return fmt.Sprintf("%q", c.Line)
}
