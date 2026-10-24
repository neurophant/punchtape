package doc

// The document's shape laws. Every refusal names the field it
// found, the law it broke and the fact itself — no advice, no
// interpretation: the author fixes the named place or the named
// intent. The laws are cited by the ids the contract prints, so a
// refusal and the contract always speak of the same rule.

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/neurophant/punchtape/internal/domain/zone"
)

// The law ids of the contract the refusals cite.
const (
	LawDocument = "law:document"
	LawPaths    = "law:paths"
)

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
// names them), the requirements that stand after this submission's
// merge (covers names them), and the wish's line universe (ranges
// name it; the resolver answers the line's own text with its
// existence). A nil field skips its check; a nil Resolvers skips
// them all.
type Resolvers struct {
	Stands      func(id string) bool
	Requirement func(id string) bool
	WishLine    func(id string) (string, bool)
}

// Validate — the whole rule set against the standing corpus: ids are
// stable path-shaped and unique in the submission; every path-valued
// field is root-relative; requirement ids are R-shaped and unique;
// every brief range parses and names lines of the wish; covers name
// requirements that stand after the merge; every row carries an
// explicit expect and every behavior carries at least one row; a
// remove names only ids the corpus carries. Replacing a standing id
// is legal — the last submission wins, every version lives in the
// journal.
func Validate(d Document, r *Resolvers) []Refusal {
	var out []Refusal
	if d.Build != nil && strings.TrimSpace(*d.Build) == "" {
		out = append(out, refusalAt("build")("the build command is empty — delete the line to keep the standing one"))
	}
	for i, glob := range d.Artifacts {
		if !ValidGlob(glob) {
			out = append(out, Refusal{
				Where: fmt.Sprintf("artifacts[%d]", i),
				Law:   LawPaths,
				Fact:  "the artifact glob must be relative to the working root (segments of [a-z0-9-._*?[]], '**' for any depth)",
			})
		}
	}
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
		out = append(out, resolveRanges(fmt.Sprintf("requirements[%d].brief", i), req.Brief, r)...)
	}
	for i, marker := range d.NotRequirements {
		if strings.TrimSpace(marker.Reason) == "" {
			out = append(out, refusalAt(fmt.Sprintf("not-a-requirement[%d].reason", i))("the reason is empty — a mark without a reason marks nothing"))
		}
		markerRefusals := resolveRanges(fmt.Sprintf("not-a-requirement[%d].lines", i), marker.Lines, r)
		out = append(out, markerRefusals...)
		if len(markerRefusals) == 0 {
			out = append(out, markerQuote(i, marker, r)...)
		}
	}
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
		for gi, glob := range b.Scope {
			if !ValidGlob(glob) {
				out = append(out, Refusal{
					Where: fmt.Sprintf("behaviors[%d].scope[%d]", i, gi),
					Law:   LawPaths,
					Fact:  "the scope glob must be relative to the working root (segments of [a-z0-9-._*?[]], '**' for any depth)",
				})
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
		runs := map[string]bool{}
		for j, row := range b.Rows {
			if row.Run.Line == "" && len(row.Run.Argv) == 0 {
				out = append(out, refusalAt(fmt.Sprintf("behaviors[%d].rows[%d].run", i, j))("the command is empty"))
			}
			if row.Expect.RC == nil && row.Expect.Out == "" && len(row.Expect.Files) == 0 {
				out = append(out, refusalAt(fmt.Sprintf("behaviors[%d].rows[%d].expect", i, j))("the expect is empty — name rc, out or files"))
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
	for i, id := range d.Remove {
		if validRequirementID(id) {
			if r != nil && r.Stands != nil && !r.Stands(id) {
				out = append(out, refusalAt(fmt.Sprintf("remove[%d]", i))(fmt.Sprintf("the corpus carries no requirement %q", id)))
			}
			continue
		}
		if !validID(id) {
			out = append(out, refusalAt(fmt.Sprintf("remove[%d]", i))("the id must be [a-z0-9-] segments joined by '/' or a requirement id (R followed by letters, digits and dots)"))
			continue
		}
		if r != nil && r.Stands != nil && !r.Stands(id) {
			out = append(out, refusalAt(fmt.Sprintf("remove[%d]", i))(fmt.Sprintf("the corpus carries no behavior %q", id)))
		}
	}
	return out
}

// resolveRanges — one range list's law: every spelling parses in the
// W grammar, and every line it names exists in the wish's universe.
func resolveRanges(where string, ranges []string, r *Resolvers) []Refusal {
	var out []Refusal
	for j, list := range ranges {
		nums, err := ParseRanges([]string{list})
		if err != nil {
			out = append(out, refusalAt(fmt.Sprintf("%s[%d]", where, j))(err.Error()))
			continue
		}
		if r == nil || r.WishLine == nil {
			continue
		}
		var missing []int
		for _, n := range nums {
			if _, ok := r.WishLine("W" + strconv.Itoa(n)); !ok {
				missing = append(missing, n)
			}
		}
		if len(missing) > 0 {
			out = append(out, refusalAt(fmt.Sprintf("%s[%d]", where, j))(
				fmt.Sprintf("the wish carries no line %s", SpellRanges(missing))))
		}
	}
	return out
}

// markerQuote — the marker's own law: the quote must be the marked
// wish lines' own text, verbatim, in line order. The caller has
// already refused broken ranges and missing lines; dismissing a line
// demands copying its exact words into the signed corpus — the mark
// cannot be quiet.
func markerQuote(i int, marker Marker, r *Resolvers) []Refusal {
	if r == nil || r.WishLine == nil {
		return nil
	}
	nums, _ := ParseRanges(marker.Lines)
	var texts []string
	for _, n := range nums {
		text, _ := r.WishLine("W" + strconv.Itoa(n))
		texts = append(texts, text)
	}
	if marker.Quote == "" {
		return []Refusal{refusalAt(fmt.Sprintf("not-a-requirement[%d].quote", i))(
			"the quote is empty — copy the marked lines' own text, verbatim")}
	}
	if want := strings.Join(texts, "\n"); want != marker.Quote {
		return []Refusal{refusalAt(fmt.Sprintf("not-a-requirement[%d].quote", i))(
			fmt.Sprintf("the quote must be the wish's own lines verbatim — want %q", want))}
	}
	return nil
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

// ValidGlob — the scope glob's law: a root-relative slash pattern whose
// segments carry letters, digits, dash, dot, underscore and the glob
// wildcards; a '**' segment matches any depth. No absolute spelling, no
// parent hop, no backslash.
func ValidGlob(g string) bool {
	if g == "" || strings.HasPrefix(g, "/") || strings.Contains(g, "\\") || strings.HasSuffix(g, "/") {
		return false
	}
	for _, seg := range strings.Split(g, "/") {
		if seg == "" || seg == "." || seg == ".." {
			return false
		}
		if seg != "**" {
			for _, c := range seg {
				switch {
				case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9',
					c == '-', c == '.', c == '_', c == '*', c == '?', c == '[', c == ']':
				default:
					return false
				}
			}
		}
	}
	return true
}
