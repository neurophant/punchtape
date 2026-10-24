package application

// The zone's views: state.md (the standing, the scopes-moved mark,
// the drift history and the next step, one page), report.md (the
// last run's facts, or the
// submission's refusals), knowledge.md (the notes and the lessons)
// and the spec tree (one small file per behavior). Views are
// generated images of the journal — rewritten whole when their
// content changes, never touched when it does not; every view's
// header carries the number of the last fact-changing run.

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/neurophant/punchtape/internal/domain/doc"
	"github.com/neurophant/punchtape/internal/domain/journal"
	"github.com/neurophant/punchtape/internal/domain/zone"
	"github.com/neurophant/punchtape/internal/ports"
)

// rowReport — one executed row's measured story for the report.
type rowReport struct {
	Key    string
	Run    doc.Cmd
	Exit   int
	Stdout string
	Facts  []string
}

// initialViews — the birth's empty views.
var initialViews = struct{ state, report, knowledge, index []byte }{
	state:     []byte("# punchtape state\n\ncorpus: empty — submit the first document\nnext: .punchtape/inbox.yaml\n"),
	report:    []byte("# punchtape report\n\nno run yet\n"),
	knowledge: []byte("# punchtape knowledge\n\nno notes yet\n"),
	index:     []byte("# punchtape spec\n\nempty\n"),
}

// historyEvent — one line of the verdict history, exactly as the
// journal's own records tell it.
type historyEvent struct {
	run  int
	kind string // verdict | drift | healed
	text string
}

// historyEvents — the verdict history from the journal's records:
// every recorded verdict, every run that caught drift (red rows
// under a standing verdict) and the run that healed it. The render
// never invents: what the journal does not carry, the history does
// not say — and a byte-exact heal cannot erase what the journal
// already holds.
func (m *machine) historyEvents() ([]historyEvent, error) {
	records, err := m.stream.Records()
	if err != nil {
		return nil, err
	}
	var out []historyEvent
	verdictSeen := false
	openDrift := 0
	for _, rec := range records {
		switch {
		case rec.Kind == journal.KindVerdict && rec.Verdict != nil:
			verdictSeen = true
			out = append(out, historyEvent{run: rec.Verdict.Run, kind: "verdict",
				text: fmt.Sprintf("run %d: verdict #%d on v%d recorded — %s",
					rec.Verdict.Run, rec.Verdict.N, rec.Verdict.CorpusVersion, shortDigest(rec.Verdict.Signature))})
		case rec.Kind == journal.KindRun && rec.Run != nil:
			reds := 0
			var keys []string
			for _, rf := range rec.Run.Rows {
				if rf.Status == journal.StatusRed {
					reds++
					keys = append(keys, rf.Key)
				}
			}
			switch {
			case reds > 0 && verdictSeen:
				openDrift = rec.Run.N
				out = append(out, historyEvent{run: rec.Run.N, kind: "drift",
					text: fmt.Sprintf("run %d: %d rows RED — drift: %s", rec.Run.N, reds, strings.Join(keys, ", "))})
			case reds == 0 && openDrift > 0 && rec.Run.N > openDrift:
				out = append(out, historyEvent{run: rec.Run.N, kind: "healed",
					text: fmt.Sprintf("run %d: healed", rec.Run.N)})
				openDrift = 0
			}
		}
	}
	return out, nil
}

// writeState — state.md: the standing, the scopes-moved mark, the
// drift history and the one next instruction, one page. The mark and
// the history are two different truths and never share a word: the
// mark says the tree grew from what the verdict certified, the
// history says a red row ever stood under it.
func (m *machine) writeState(corpus doc.Document, scopes map[string]string, tree []string) error {
	var b bytes.Buffer
	fmt.Fprintf(&b, "# punchtape state — run %d\n\n", m.standing.RunNumber)
	fmt.Fprintf(&b, "goal: %s", shortDigest(m.standing.GoalDigest))
	if brief, ok := m.briefAccounting(corpus); ok {
		fmt.Fprintf(&b, " — wish at %s", zone.Path(zone.ObjectsDir, m.standing.GoalAddr[:2], m.standing.GoalAddr[2:]))
		fmt.Fprintf(&b, "\nbrief: %d/%d lines; requirements %d standing, %d covered; not-a-requirement: %d\n",
			brief.Claimed, brief.Lines, brief.Requirements, brief.Covered, brief.Markers)
		// Every mark on the open page: the dismissed lines' own
		// words beside the reason — a mark a reader can weigh.
		for _, marker := range m.standing.Registry.Markers {
			nums, _ := doc.ParseRanges(marker.Lines)
			fmt.Fprintf(&b, "not-a-requirement %s: %q — %s\n", doc.SpellRanges(nums), marker.Quote, marker.Reason)
		}
		if len(brief.Holes) > 0 {
			fmt.Fprintf(&b, "uncovered brief lines: %s\n", doc.SpellRanges(brief.Holes))
		}
	} else {
		fmt.Fprintf(&b, "\n")
	}
	fmt.Fprintf(&b, "corpus: v%d, %d behaviors\n",
		m.standing.CorpusVersion, len(corpus.Behaviors))
	fmt.Fprintf(&b, "submissions: %d accepted, %d refused, %d completed\n",
		len(m.standing.Submissions), m.standing.Refused, m.standing.Completions)
	// The instance's own origin fact: the machine version that
	// created it. An instance born before the fact says so honestly.
	atBirth := m.standing.BirthVersion
	if atBirth == "" {
		atBirth = "(not recorded)"
	}
	fmt.Fprintf(&b, "machine at birth: %s\n", atBirth)
	if v := m.standing.Verdict; v != nil {
		fmt.Fprintf(&b, "verdict: #%d on v%d, run %d, signature %s", v.N, v.CorpusVersion, v.Run, shortDigest(v.Signature))
		moved := false
		for _, sc := range v.Scopes {
			if _, ok := m.standing.Behaviors[sc.ID]; ok && scopes[sc.ID] != sc.Pointer {
				if !moved {
					fmt.Fprintf(&b, " — scopes moved:")
					moved = true
				}
				fmt.Fprintf(&b, " %s", sc.ID)
			}
		}
		if !moved {
			fmt.Fprintf(&b, " — standing")
		}
		fmt.Fprintf(&b, "\n")
	}
	if events, err := m.historyEvents(); err != nil {
		fmt.Fprintf(&b, "verdict history: (journal unreadable: %v)\n", err)
	} else if len(events) > 0 {
		fmt.Fprintf(&b, "verdict history:\n")
		for _, ev := range events {
			fmt.Fprintf(&b, "- %s\n", ev.text)
		}
	}
	for _, beh := range corpus.Behaviors {
		green, stale, red := m.behaviorRows(beh)
		mark := ""
		switch {
		case red > 0:
			mark = fmt.Sprintf(" — %d green, %d stale, %d RED", green, stale, red)
		case stale > 0:
			mark = fmt.Sprintf(" — %d green, %d stale", green, stale)
		case len(beh.Rows) > 0:
			mark = fmt.Sprintf(" — %d green", green)
		}
		fmt.Fprintf(&b, "- %s: %d rows%s (scope %s)\n", beh.ID, len(beh.Rows), mark, shortDigest(scopes[beh.ID]))
	}
	declared, sources := SplitUncovered(tree, m.allScopes(corpus), corpus.Artifacts)
	if len(declared) > 0 {
		fmt.Fprintf(&b, "\nartifacts (declared):\n")
		for _, p := range declared {
			fmt.Fprintf(&b, "- %s\n", p)
		}
	}
	if len(sources) > 0 {
		fmt.Fprintf(&b, "\nuncovered sources:\n")
		for _, p := range sources {
			fmt.Fprintf(&b, "- %s\n", p)
		}
	}
	fmt.Fprintf(&b, "\nknowledge: %d notes, %d lessons\n", m.noteCount(), len(m.standing.Lessons))
	for _, reason := range m.redReasons(corpus) {
		if m.hasLessonFor(reason) {
			fmt.Fprintf(&b, "knowledge for the reason %q — knowledge.md\n", reason)
		}
	}
	fmt.Fprintf(&b, "\nnext: %s\n", m.nextInstruction(corpus, false))
	return m.writeIfChanged(filepath.Join(m.zoneDir, zone.StateFile), b.Bytes())
}

// behaviorRows — one behavior's standing counts: green, stale (never
// run, or a row of a scope-less behavior), red. A red fact surfaces
// whatever the scope: staleness is secondary to a measured failure.
func (m *machine) behaviorRows(b doc.Behavior) (green, stale, red int) {
	for i := range b.Rows {
		key := b.ID + "/" + strconv.Itoa(i)
		fact, ok := m.standing.Rows[key]
		switch {
		case !ok:
			stale++
		case fact.Status == journal.StatusRed:
			red++
		case len(b.Scope) == 0:
			stale++
		default:
			green++
		}
	}
	return green, stale, red
}

// countRows — the whole corpus's red and stale counts, the one source
// of the standing's next instruction.
func (m *machine) countRows(corpus doc.Document) (red, stale int) {
	for _, b := range corpus.Behaviors {
		_, s, r := m.behaviorRows(b)
		red += r
		stale += s
	}
	return red, stale
}

// redReasons — the distinct broken facts of the standing's red rows.
func (m *machine) redReasons(corpus doc.Document) []string {
	var out []string
	for _, b := range corpus.Behaviors {
		for i := range b.Rows {
			key := b.ID + "/" + strconv.Itoa(i)
			if fact, ok := m.standing.Rows[key]; ok && fact.Status == journal.StatusRed {
				for _, f := range fact.Facts {
					if !contains(out, f) {
						out = append(out, f)
					}
				}
			}
		}
	}
	return out
}

// allScopes — every behavior's scope globs in corpus order.
func (m *machine) allScopes(corpus doc.Document) []string {
	var globs []string
	for _, b := range corpus.Behaviors {
		globs = append(globs, b.Scope...)
	}
	return globs
}

// writeReport — report.md: the run's whole picture, rewritten whole
// every time it moves — the build and row facts, red grouped by
// reason, and the completion act's outcome among them, so no old
// word outlives its run.
func (m *machine) writeReport(n int, build *journal.Build, rows []rowReport, notes []string) error {
	var b bytes.Buffer
	fmt.Fprintf(&b, "# punchtape report — run %d\n\n", n)
	if build != nil {
		switch {
		case build.Reason != "":
			fmt.Fprintf(&b, "build: RED — %s\n", build.Reason)
		case build.TimedOut:
			fmt.Fprintf(&b, "build: RED — timed out after %s\n", RowTimeout)
		case build.Exit != 0:
			fmt.Fprintf(&b, "build: RED — exit %d\n", build.Exit)
		default:
			fmt.Fprintf(&b, "build: exit %d\n", build.Exit)
		}
		fmt.Fprintf(&b, "\n")
	}
	if len(rows) == 0 && len(notes) == 0 {
		fmt.Fprintf(&b, "no stale rows\n")
	}
	green := 0
	byReason := map[string][]string{}
	var order []string
	for _, r := range rows {
		if len(r.Facts) == 0 {
			green++
			continue
		}
		for _, f := range r.Facts {
			if _, seen := byReason[f]; !seen {
				order = append(order, f)
			}
			byReason[f] = append(byReason[f], r.Key)
		}
	}
	if green > 0 {
		fmt.Fprintf(&b, "green: %d\n", green)
	}
	for _, reason := range order {
		fmt.Fprintf(&b, "\n%s\n", reason)
		for _, key := range byReason[reason] {
			fmt.Fprintf(&b, "- %s\n", key)
		}
	}
	for _, note := range notes {
		fmt.Fprintf(&b, "\n%s\n", note)
	}
	return m.writeIfChanged(filepath.Join(m.zoneDir, zone.ReportFile), b.Bytes())
}

// reportRefusals — report.md as the refusal list: the field, the law,
// the fact.
func (m *machine) reportRefusals(refusals []doc.Refusal) error {
	var b bytes.Buffer
	fmt.Fprintf(&b, "# punchtape report\n\n")
	for _, r := range refusals {
		fmt.Fprintf(&b, "%s: %s (%s)\n", r.Where, r.Fact, r.Law)
	}
	fmt.Fprintf(&b, "\nthe submission stays in %s — fix it there and run again\n", zone.Path(zone.InboxSlot))
	return m.writeIfChanged(filepath.Join(m.zoneDir, zone.ReportFile), b.Bytes())
}

// writeKnowledge — knowledge.md: each note exactly once — its whole
// text with its tags spelled on the same line — then the index that
// maps every tag to its notes by number, then the lessons grouped by
// reason. Tags are a mechanical intersection of bytes — grouping
// only, never a check.
func (m *machine) writeKnowledge() error {
	var b bytes.Buffer
	fmt.Fprintf(&b, "# punchtape knowledge — run %d\n\n", m.standing.RunNumber)
	vocab := m.vocabulary()
	notes := m.notes()
	if len(notes) == 0 && len(m.standing.Lessons) == 0 {
		fmt.Fprintf(&b, "no notes yet\n")
		return m.writeIfChanged(filepath.Join(m.zoneDir, zone.KnowledgeFile), b.Bytes())
	}
	// One print per note: the text stands once, the tags ride its
	// line, the index is the only other place a tag appears.
	byTag := map[string][]string{}
	var tagOrder []string
	for i, note := range notes {
		num := fmt.Sprintf("n%d", i+1)
		tags := noteTags(note, vocab)
		if len(tags) == 0 {
			fmt.Fprintf(&b, "- [%s] (untagged) %s\n", num, note)
			continue
		}
		fmt.Fprintf(&b, "- [%s] (tags: %s) %s\n", num, strings.Join(tags, ", "), note)
		for _, tag := range tags {
			if _, seen := byTag[tag]; !seen {
				tagOrder = append(tagOrder, tag)
			}
			byTag[tag] = append(byTag[tag], num)
		}
	}
	if len(tagOrder) > 0 {
		sort.Strings(tagOrder)
		fmt.Fprintf(&b, "\n## index\n\n")
		for _, tag := range tagOrder {
			fmt.Fprintf(&b, "%s: %s\n", tag, strings.Join(byTag[tag], ", "))
		}
	}
	byReason := map[string][]journal.Lesson{}
	var reasons []string
	for _, l := range m.standing.Lessons {
		if _, seen := byReason[l.Reason]; !seen {
			reasons = append(reasons, l.Reason)
		}
		byReason[l.Reason] = append(byReason[l.Reason], l)
	}
	sort.Strings(reasons)
	for _, reason := range reasons {
		fmt.Fprintf(&b, "\n## lesson: %s\n", reason)
		for _, l := range byReason[reason] {
			fmt.Fprintf(&b, "- %s: runs %d→%d\n  before: %s\n  after:  %s\n",
				l.Key, l.RedRun, l.GreenRun, oneLine(l.RedRow), oneLine(l.GreenRow))
		}
	}
	return m.writeIfChanged(filepath.Join(m.zoneDir, zone.KnowledgeFile), b.Bytes())
}

// writeSpec — the spec tree: the index and one small file per
// behavior; the pages of retired behaviors are removed — the tree is
// an image of the standing corpus, nothing else. A page that cannot
// land or leave is a loud failure: a stale spec is a stale face.
func (m *machine) writeSpec(corpus doc.Document) error {
	specDir := filepath.Join(m.zoneDir, zone.SpecDir)
	var index bytes.Buffer
	fmt.Fprintf(&index, "# punchtape spec — run %d\n\n", m.standing.RunNumber)
	if len(corpus.Behaviors) == 0 {
		fmt.Fprintf(&index, "empty\n")
	}
	living := map[string]bool{}
	for _, b := range corpus.Behaviors {
		living[b.ID] = true
		status := m.behaviorStatus(b)
		fmt.Fprintf(&index, "- %s — %s — %s\n", b.ID, b.Name, status)
		var page bytes.Buffer
		fmt.Fprintf(&page, "# %s — %s\n\n", b.ID, b.Name)
		fmt.Fprintf(&page, "run: %d\n\n%s\n\n", m.standing.RunNumber, b.Prose)
		if len(b.Scope) > 0 {
			fmt.Fprintf(&page, "scope: %s\n\n", strings.Join(b.Scope, ", "))
		}
		for i, row := range b.Rows {
			fmt.Fprintf(&page, "## row %d\n\ncommand: %s\n", i, commandOf(row.Run))
			if expect := expectOf(row.Expect); expect != "" {
				fmt.Fprintf(&page, "%s\n", expect)
			}
			fmt.Fprintf(&page, "proof: %s (%s)\n\n", m.rowStatus(b.ID, i, b), proofShort(m, b.ID, i))
		}
		path := filepath.Join(specDir, filepath.FromSlash(b.ID)+".md")
		if err := ports.FSBind.MkdirAll(filepath.Dir(path)); err != nil {
			return err
		}
		if err := m.writeIfChanged(path, page.Bytes()); err != nil {
			return err
		}
	}
	if err := m.writeIfChanged(filepath.Join(specDir, "index.md"), index.Bytes()); err != nil {
		return err
	}
	// The retired behaviors' pages leave with them. The walk yields
	// paths under the spec directory: the relative spelling is the
	// prefix cut — there is no error branch to armor.
	return ports.FSBind.Walk(specDir, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(p, ".md") || p == filepath.Join(specDir, "index.md") {
			return nil
		}
		rel := strings.TrimPrefix(filepath.ToSlash(p), filepath.ToSlash(specDir)+"/")
		id := strings.TrimSuffix(rel, ".md")
		if !living[id] {
			return ports.FSBind.Remove(p)
		}
		return nil
	})
}

// behaviorStatus — one behavior's one-word standing for the index.
func (m *machine) behaviorStatus(b doc.Behavior) string {
	if len(b.Rows) == 0 {
		return "no rows"
	}
	for i := range b.Rows {
		if m.rowStatus(b.ID, i, b) != "green" {
			return "not green"
		}
	}
	return "green"
}

// rowStatus — one row's standing: green, stale or red. The standing
// fact says what it says; an empty scope keeps its rows honestly
// stale.
func (m *machine) rowStatus(id string, i int, b doc.Behavior) string {
	key := id + "/" + strconv.Itoa(i)
	fact, ok := m.standing.Rows[key]
	if !ok {
		return "stale"
	}
	if fact.Status == journal.StatusRed {
		return "red"
	}
	if len(b.Scope) == 0 {
		return "stale"
	}
	return "green"
}

// proofShort — the row's recorded proof fingerprint, short; a row
// that never ran carries none.
func proofShort(m *machine, id string, i int) string {
	fact, ok := m.standing.Rows[id+"/"+strconv.Itoa(i)]
	if !ok {
		return "none"
	}
	return shortDigest(fact.Proof)
}

// vocabulary — the instance's own bytes as a word set: the recorded
// goal and the standing corpus. No semantics anywhere — a mechanical
// intersection is all a tag ever is.
func (m *machine) vocabulary() map[string]bool {
	words := map[string]bool{}
	if goal, err := m.objects.Get(m.standing.GoalAddr); err == nil {
		collectWords(string(goal), words)
	}
	collectWords(string(doc.Marshal(m.corpusDoc())), words)
	return words
}

// notes — the authors' notes of every submission, in arrival order.
func (m *machine) notes() []string {
	var out []string
	for _, sub := range m.standing.Submissions {
		data, err := m.objects.Get(sub.DocAddr)
		if err != nil {
			continue
		}
		d, err := doc.Decode(data)
		if err != nil {
			continue
		}
		for _, n := range d.Notes {
			out = append(out, n.Text)
		}
	}
	return out
}

// noteCount — the notes' count for the state's counter.
func (m *machine) noteCount() int { return len(m.notes()) }

// hasLessonFor — whether the lessons hold a reason class.
func (m *machine) hasLessonFor(reason string) bool {
	for _, l := range m.standing.Lessons {
		if l.Reason == reason {
			return true
		}
	}
	return false
}

// collectWords — the byte-level word set of a text: lowercase runs
// of letters and digits, two characters or longer.
func collectWords(text string, into map[string]bool) {
	var sb strings.Builder
	flush := func() {
		if sb.Len() >= 2 {
			into[strings.ToLower(sb.String())] = true
		}
		sb.Reset()
	}
	for _, r := range text {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			sb.WriteRune(r)
		default:
			flush()
		}
	}
	flush()
}

// noteTags — the note's own words met with the vocabulary.
func noteTags(note string, vocab map[string]bool) []string {
	words := map[string]bool{}
	collectWords(note, words)
	var tags []string
	for w := range words {
		if vocab[w] {
			tags = append(tags, w)
		}
	}
	sort.Strings(tags)
	return tags
}

// commandOf — the row's command as a human reads it.
func commandOf(c doc.Cmd) string {
	if c.Argv != nil {
		return strings.Join(c.Argv, " ")
	}
	return c.Line
}

// expectOf — the row's expectation in human lines.
func expectOf(e doc.Expect) string {
	var lines []string
	if e.RC != nil {
		lines = append(lines, fmt.Sprintf("exit: %d", *e.RC))
	}
	if e.Out != "" {
		lines = append(lines, fmt.Sprintf("stdout: %s", e.Out))
	}
	paths := make([]string, 0, len(e.Files))
	for p := range e.Files {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	for _, p := range paths {
		lines = append(lines, fmt.Sprintf("file %s: %s", p, e.Files[p]))
	}
	return strings.Join(lines, "\n")
}

// oneLine — a row's bytes as one line for the lesson's story.
func oneLine(rowBytes string) string {
	s := strings.Join(strings.Fields(rowBytes), " ")
	if len(s) > 100 {
		s = s[:100] + "…"
	}
	return s
}

// contains — a slice's plain membership.
func contains(xs []string, v string) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

// writeIfChanged — the view's whole content, only when it differs.
// A write that cannot land is a loud failure: the view is the
// machine's face, a stale face is a lie.
func (m *machine) writeIfChanged(path string, data []byte) error {
	if bytes.Equal(data, m.readFileOrEmpty(path)) {
		return nil
	}
	return ports.FSBind.WriteAtomic(path, data)
}

// shortDigest — the human-facing 12 characters of a digest.
func shortDigest(d string) string {
	if len(d) > 12 {
		return d[:12]
	}
	return d
}
