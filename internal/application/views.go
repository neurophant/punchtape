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
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/neurophant/punchtape/internal/domain/doc"
	"github.com/neurophant/punchtape/internal/domain/fprint"
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
func historyEvents(records []journal.Record) []historyEvent {
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
	return out
}

// driftSince — the first run past the verdict that saw a certified
// behavior's claims move, zero when no drift stands. The journal's
// row facts are the witnesses: a fact whose proof no longer
// recomposes over the verdict's own pointer was measured against a
// moved tree. With no witness — the move is younger than any
// execution, or the journal will not read — the current run is the
// one that sees it.
func driftSince(v *journal.Verdict, claims map[string]claimsFacts, envHash string, current int, records []journal.Record, rowBytes func(string) []byte) int {
	drifted := map[string]bool{}
	pointer := map[string]string{}
	for _, sc := range v.Scopes {
		pointer[sc.ID] = sc.Pointer
		if facts, living := claims[sc.ID]; living && facts.digest != sc.Pointer {
			drifted[sc.ID] = true
		}
	}
	if len(drifted) == 0 {
		return 0
	}
	first := 0
	for _, rec := range records {
		if rec.Kind != journal.KindRun || rec.Run == nil || rec.Run.N <= v.Run {
			continue
		}
		for _, rf := range rec.Run.Rows {
			id := rf.Key[:strings.LastIndexByte(rf.Key, '/')]
			if !drifted[id] {
				continue
			}
			if rf.Proof == fprint.Proof(rowBytes(rf.Row), pointer[id], envHash) {
				continue
			}
			if first == 0 || rec.Run.N < first {
				first = rec.Run.N
			}
		}
	}
	if first == 0 {
		return current
	}
	return first
}

// writeState — state.md: the standing, the queue, the verdict's own
// standing or drift, the drift history, the deferred orders, the
// behaviors' proofs, the tree's open and excluded lines, the
// knowledge counters and the one next instruction, one page. The
// verdict line and the history are two different truths and never
// share a word: the line says the tree grew from what the verdict
// certified, the history says a red row ever stood under it.
func (m *machine) writeState(corpus doc.Document, claims map[string]claimsFacts, instruction string) error {
	var b bytes.Buffer
	fmt.Fprintf(&b, "# punchtape state — run %d\n\n", m.standing.RunNumber)
	brief := m.briefAccounting(corpus)
	fmt.Fprintf(&b, "demands: %d/%d lines; requirements %d standing, %d covered; markers: %d\n",
		brief.Claimed, brief.Lines, brief.Requirements, brief.Covered, brief.Markers)
	if holes := m.unresolvedLines(); len(holes) > 0 {
		fmt.Fprintf(&b, "unresolved lines: %s\n", strings.Join(holes, ", "))
	}
	// The queue is a fact of the files: the inbox stands or not.
	if _, err := ports.FSBind.Stat(filepath.Join(m.zoneDir, zone.InboxSlot)); err == nil {
		fmt.Fprintf(&b, "queue: submission waiting\n")
	} else {
		fmt.Fprintf(&b, "queue: empty\n")
	}
	// The standing signature, only while it stands: the next
	// demand-append clears it.
	if m.standing.LastNoNewDemands > 0 {
		fmt.Fprintf(&b, "last no-new-demands: run %d\n", m.standing.LastNoNewDemands)
	}
	fmt.Fprintf(&b, "corpus: v%d, %d behaviors\n",
		m.standing.CorpusVersion, len(corpus.Behaviors))
	if e := m.standing.LastExecution; e != nil {
		fmt.Fprintf(&b, "executions: last run %d — %s exit %d\n", e.At, strings.Join(e.Lines, ", "), e.Exit)
	}
	fmt.Fprintf(&b, "submissions: %d accepted, %d refused, %d completed\n",
		m.standing.Accepted, m.standing.Refused, m.standing.Completions)
	// The instance's own origin fact: the machine version that
	// created it — the run's gate has already refused any other.
	fmt.Fprintf(&b, "machine at birth: %s\n", m.standing.BirthVersion)
	if v := m.standing.Verdict; v != nil {
		fmt.Fprintf(&b, "verdict: #%d on v%d, run %d, signature %s", v.N, v.CorpusVersion, v.Run, shortDigest(v.Signature))
		since := driftSince(v, claims, EnvHash(m.pins), m.standing.RunNumber, m.records, m.rowBytesOf)
		switch {
		case since > 0:
			fmt.Fprintf(&b, " — open (drift since run %d)\n", since)
		case m.treeHoles(corpus) > 0:
			fmt.Fprintf(&b, " — standing over %d uncovered lines\n", m.treeHoles(corpus))
		default:
			fmt.Fprintf(&b, " — standing\n")
		}
	}
	if events := historyEvents(m.records); len(events) > 0 {
		fmt.Fprintf(&b, "verdict history:\n")
		// The page shows the recent ten; the whole story lives in the
		// journal, and the elision says so.
		if hidden := len(events) - 10; hidden > 0 {
			fmt.Fprintf(&b, "- … and %d more — see %s\n", hidden, zone.Path(zone.JournalDir))
		}
		shown := events
		if len(shown) > 10 {
			shown = shown[len(shown)-10:]
		}
		for _, ev := range shown {
			fmt.Fprintf(&b, "- %s\n", ev.text)
		}
	}
	var deferred []doc.OpenOrder
	for _, order := range m.standing.Registry.Open {
		if order.State == doc.StateDeferred {
			deferred = append(deferred, order)
		}
	}
	if len(deferred) > 0 {
		fmt.Fprintf(&b, "deferred orders:\n")
		for _, order := range deferred {
			fmt.Fprintf(&b, "- %s — %s — %s\n", order.Demand, order.Question, order.Reason)
		}
	}
	for _, beh := range corpus.Behaviors {
		green, stale, red := m.behaviorRows(beh)
		files, lines := claimsShape(beh)
		shape := fmt.Sprintf("(claims %d files / %d lines; proof %s)", files, lines, shortDigest(claims[beh.ID].digest))
		switch {
		case red > 0:
			fmt.Fprintf(&b, "- %s: %d rows %s — %d green, %d stale, %d RED\n",
				beh.ID, len(beh.Rows), shape, green, stale, red)
		case stale > 0:
			fmt.Fprintf(&b, "- %s: %d rows %s — %d green, %d stale\n",
				beh.ID, len(beh.Rows), shape, green, stale)
		default:
			fmt.Fprintf(&b, "- %s: %d rows — %d green %s\n",
				beh.ID, len(beh.Rows), green, shape)
		}
	}
	// The tree's own arithmetic: an unreadable walk is named aloud and
	// the page still lands — the ledger and the share wait for a walk
	// that reads.
	cov, covErr := m.coverage(corpus)
	if covErr != nil {
		fmt.Fprintf(&b, "uncovered lines: (tree unreadable: %v)\n", covErr)
	} else {
		paths := make([]string, 0, len(cov.Holes))
		for path := range cov.Holes {
			paths = append(paths, path)
		}
		sort.Strings(paths)
		if len(paths) > 0 {
			fmt.Fprintf(&b, "uncovered lines:\n")
			shown := paths
			if len(paths) > 10 {
				shown = paths[:10]
			}
			for _, path := range shown {
				fmt.Fprintf(&b, "- %s\n", spellHoles(path, cov.Holes[path], true))
			}
			if len(paths) > 10 {
				fmt.Fprintf(&b, "… and %d more files — see %s\n", len(paths)-10, zone.Path(zone.HolesFile))
			}
		}
		if len(cov.Excluded) > 0 {
			fmt.Fprintf(&b, "excluded files:\n")
			for _, ex := range cov.Excluded {
				fmt.Fprintf(&b, "- %s — %s\n", ex.Path, ex.Reason)
			}
		}
		share := 0
		if cov.Total > 0 {
			share = 100 * cov.ExcludedLines / cov.Total
		}
		fmt.Fprintf(&b, "excluded lines: %d ranges (%d %% of tree)\n",
			len(m.standing.Registry.Exclusions), share)
		if m.tree.declared() {
			fmt.Fprintf(&b, "outside the universe: %d files\n", m.tree.lastFacts.outside)
			fmt.Fprintf(&b, "universe: %d of %d files\n",
				m.tree.lastFacts.inside, m.tree.lastFacts.inside+m.tree.lastFacts.outside)
		}
	}
	fmt.Fprintf(&b, "\nknowledge: %d notes, %d lessons\n", len(m.standing.Notes), len(m.standing.Lessons))
	fmt.Fprintf(&b, "\nnext: %s\n", instruction)
	if covErr == nil {
		if err := m.writeHoles(cov); err != nil {
			return err
		}
	}
	return m.writeIfChanged(filepath.Join(m.zoneDir, zone.StateFile), b.Bytes())
}

// treeHoles — the tree's open line count of the run's coverage fact.
func (m *machine) treeHoles(corpus doc.Document) int {
	cov, err := m.coverage(corpus)
	if err != nil {
		return 0
	}
	n := 0
	for _, ranges := range cov.Holes {
		for _, r := range ranges {
			n += r.Last - r.First + 1
		}
	}
	return n
}

// writeHoles — the tree's whole hole ledger, uncapped: the state's
// list points here when it runs out of room. The caller hands the
// walk's own successful fact.
func (m *machine) writeHoles(cov fprint.Coverage) error {
	var b bytes.Buffer
	fmt.Fprintf(&b, "# punchtape holes — run %d\n\n", m.standing.RunNumber)
	paths := make([]string, 0, len(cov.Holes))
	for path := range cov.Holes {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	if len(paths) == 0 {
		fmt.Fprintf(&b, "no holes\n")
	}
	for _, path := range paths {
		fmt.Fprintf(&b, "- %s\n", spellHoles(path, cov.Holes[path], true))
	}
	return m.writeIfChanged(filepath.Join(m.zoneDir, zone.HolesFile), b.Bytes())
}

// writeExclusionsProposal — the holes as a ready draft: every open
// range in the submission's own form, the reasons and references
// the author's own to fill. With no hole standing, a stale draft
// leaves — the proposal never outlives its own arithmetic.
func (m *machine) writeExclusionsProposal(corpus doc.Document) error {
	cov, err := m.coverage(corpus)
	if err != nil {
		return err
	}
	paths := make([]string, 0, len(cov.Holes))
	for path := range cov.Holes {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	file := filepath.Join(m.zoneDir, zone.ExclusionsProposalFile)
	if len(paths) == 0 {
		if _, err := ports.FSBind.Stat(file); err == nil {
			return ports.FSBind.Remove(file)
		}
		return nil
	}
	var b bytes.Buffer
	fmt.Fprintf(&b, "# generated from the run's own hole arithmetic — add reasons and refs, drop what you claim\n")
	fmt.Fprintf(&b, "uncovered-lines:\n")
	for _, path := range paths {
		for _, r := range cov.Holes[path] {
			fmt.Fprintf(&b, "  - {range: %s, reason: '', refs: []}\n", strconv.Quote(fmt.Sprintf("%s:%d-%d", path, r.First, r.Last)))
		}
	}
	return m.writeIfChanged(file, b.Bytes())
}

// writeTrace — the meaning's own showcase: the stream's every line
// with its text, its declared outcome and the retelling beside each
// other, then the carrier, the requirements, the behaviors with
// their claims and the proof rows with their standings. Nothing is
// judged here; everything is placed for the eye.
func (m *machine) writeTrace(corpus doc.Document, claims map[string]claimsFacts) error {
	var b bytes.Buffer
	fmt.Fprintf(&b, "# punchtape trace — run %d\n\n", m.standing.RunNumber)
	reqSpan := map[int][]string{}
	for _, id := range m.standing.Registry.Order {
		if nums, err := doc.ParseRanges(m.standing.Registry.Requirements[id].Lines); err == nil {
			for _, n := range nums {
				reqSpan[n] = append(reqSpan[n], id)
			}
		}
	}
	behaviorsOf := func(reqID string) []doc.Behavior {
		var out []doc.Behavior
		for _, beh := range corpus.Behaviors {
			for _, cover := range beh.Covers {
				if cover == reqID {
					out = append(out, beh)
					break
				}
			}
		}
		return out
	}
	for n := 1; n <= len(m.universe); n++ {
		fmt.Fprintf(&b, "D%d «%s»\n", n, m.universe[n-1])
		lo, ok := m.standing.LineOutcomes[n]
		if !ok {
			fmt.Fprintf(&b, "├─ no intent — unresolved (no outcome)\n")
			continue
		}
		fmt.Fprintf(&b, "├─ intent: %s — «%s»\n", lo.Outcome, lo.Prose)
		switch lo.Outcome {
		case doc.OutcomeWork, doc.OutcomeDescribe:
			if len(reqSpan[n]) == 0 {
				fmt.Fprintf(&b, "│  └─ (%s pending)\n", lo.Outcome)
				continue
			}
			for _, id := range reqSpan[n] {
				fmt.Fprintf(&b, "│  └─ %s (span %s)\n", id, strings.Join(m.standing.Registry.Requirements[id].Lines, ", "))
				for _, beh := range behaviorsOf(id) {
					fmt.Fprintf(&b, "│     └─ %s (covers %s)\n", beh.ID, id)
					fmt.Fprintf(&b, "│        claims: %s — token %s\n",
						strings.Join(beh.Claims, ", "), fprint.Token12(claims[beh.ID].digest))
					for i, row := range beh.Rows {
						if !doc.Pinned(row.Expect) {
							continue
						}
						fmt.Fprintf(&b, "│        row %d PROOF: %s → %s\n", i, commandOf(row.Run), m.rowStatus(beh.ID, i, beh))
					}
				}
			}
		case doc.OutcomeMark:
			fmt.Fprintf(&b, "│  └─ marker (reason: %s)\n", m.markerReasonOf(n))
		case doc.OutcomeAnswer:
			// The answer's text stood at the declaration and rides the
			// line forever — it cannot be absent.
			fmt.Fprintf(&b, "│  └─ answer: %s\n", m.standing.Answers[n])
		case doc.OutcomeExecute:
			if e := m.standing.LastExecution; e != nil && slices.Contains(e.Lines, doc.DemandID(n)) {
				fmt.Fprintf(&b, "│  └─ execution (last): exit %d — %s\n", e.Exit, fprint.Token12(e.StdoutDigest))
				continue
			}
			fmt.Fprintf(&b, "│  └─ (execute pending)\n")
		case doc.OutcomeAsk:
			if order, state := m.orderOn(n); state != "" {
				fmt.Fprintf(&b, "│  └─ order (%s): %s\n", state, order.Question)
				continue
			}
			fmt.Fprintf(&b, "│  └─ (ask pending)\n")
		}
	}
	return m.writeIfChanged(filepath.Join(m.zoneDir, zone.TraceFile), b.Bytes())
}

// markerReasonOf — the standing marker that claims the line, its
// reason; empty when none stands.
func (m *machine) markerReasonOf(n int) string {
	for _, marker := range m.standing.Registry.Markers {
		if nums, err := doc.ParseRanges(marker.Lines); err == nil {
			for _, k := range nums {
				if k == n {
					return marker.Reason
				}
			}
		}
	}
	return ""
}

// orderOn — the standing open order of the line and its state.
func (m *machine) orderOn(n int) (doc.OpenOrder, string) {
	for _, order := range m.standing.Registry.Open {
		if orderLine(order) == n {
			return order, order.State
		}
	}
	return doc.OpenOrder{}, ""
}

// spellHoles — one file's open lines as one line: "path: count
// (ranges)" for the state's list, "path count (ranges)" for the
// verdict's report line.
func spellHoles(path string, ranges []fprint.Range, colon bool) string {
	count := 0
	spelled := make([]string, 0, len(ranges))
	for _, r := range ranges {
		count += r.Last - r.First + 1
		if r.First == r.Last {
			spelled = append(spelled, strconv.Itoa(r.First))
			continue
		}
		spelled = append(spelled, fmt.Sprintf("%d-%d", r.First, r.Last))
	}
	sep := " "
	if colon {
		sep = ": "
	}
	return fmt.Sprintf("%s%s%d (%s)", path, sep, count, strings.Join(spelled, ", "))
}

// claimsShape — the behavior's claims as the state's two numbers:
// the distinct files named and the union of their claimed lines.
func claimsShape(b doc.Behavior) (files, lines int) {
	perFile := map[string]map[int]bool{}
	for _, claim := range b.Claims {
		rel, first, last, err := fprint.ParseLineRange(claim)
		if err != nil {
			continue
		}
		if perFile[rel] == nil {
			perFile[rel] = map[int]bool{}
		}
		for n := first; n <= last; n++ {
			perFile[rel][n] = true
		}
	}
	files = len(perFile)
	for _, held := range perFile {
		lines += len(held)
	}
	return files, lines
}

// behaviorRows — one behavior's standing counts: green, stale (never
// run, or a row of a claims-less behavior), red. A red fact surfaces
// whatever the claims: staleness is secondary to a measured failure.
func (m *machine) behaviorRows(b doc.Behavior) (green, stale, red int) {
	for i := range b.Rows {
		key := b.ID + "/" + strconv.Itoa(i)
		fact, ok := m.standing.Rows[key]
		switch {
		case !ok:
			stale++
		case fact.Status == journal.StatusRed:
			red++
		case len(b.Claims) == 0:
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
			fmt.Fprintf(&b, "build: RED — timed out after %s\n", m.rowTimeout)
		case build.Exit != 0:
			fmt.Fprintf(&b, "build: RED — exit %d\n", build.Exit)
		default:
			fmt.Fprintf(&b, "build: exit %d\n", build.Exit)
		}
		fmt.Fprintf(&b, "\n")
	}
	if len(m.execFacts) > 0 {
		fmt.Fprintf(&b, "\nexecutions:\n")
		for _, e := range m.execFacts {
			fmt.Fprintf(&b, "- %s — %s — exit %d\n", strings.Join(e.Lines, ", "), commandOf(e.Run), e.Exit)
			if e.StdoutAddr != "" {
				fmt.Fprintf(&b, "  stdout: %s\n", e.StdoutAddr)
				continue
			}
			fmt.Fprintf(&b, "  stdout: %s\n", e.Stdout)
		}
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
// the fact — the header naming the run that refused, like every
// view's header names its fact's run.
func (m *machine) reportRefusals(refusals []doc.Refusal) error {
	var b bytes.Buffer
	fmt.Fprintf(&b, "# punchtape report — run %d\n\n", m.standing.RunNumber)
	for _, r := range refusals {
		fmt.Fprintf(&b, "%s: %s (%s)\n", r.Where, r.Fact, r.Law)
	}
	fmt.Fprintf(&b, "\nthe submission stays in %s — fix it there and run again\n", m.slot(zone.InboxSlot))
	return m.writeIfChanged(filepath.Join(m.zoneDir, zone.ReportFile), b.Bytes())
}

// writeKnowledge — knowledge.md: each note exactly once — its kind,
// its references and its whole text on one line — then the index
// that maps every reference to its notes by number, then the lessons
// grouped by reason. References are the note's own gated data — the
// index is arithmetic over them, never a matching of words.
func (m *machine) writeKnowledge() error {
	var b bytes.Buffer
	fmt.Fprintf(&b, "# punchtape knowledge — run %d\n\n", m.standing.RunNumber)
	notes := m.standing.Notes
	if len(notes) == 0 && len(m.standing.Lessons) == 0 {
		fmt.Fprintf(&b, "no notes yet\n")
		return m.writeIfChanged(filepath.Join(m.zoneDir, zone.KnowledgeFile), b.Bytes())
	}
	for i, note := range notes {
		num := fmt.Sprintf("n%d", i+1)
		if len(note.Refs) == 0 {
			fmt.Fprintf(&b, "- [%s] (%s) %s\n", num, note.Kind, note.Text)
			continue
		}
		refs := append([]string(nil), note.Refs...)
		sort.Strings(refs)
		fmt.Fprintf(&b, "- [%s] (%s; refs: %s) %s\n", num, note.Kind, strings.Join(refs, ", "), note.Text)
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
// an image of the standing corpus, nothing else. A page carries its
// sections in one order: the head, the story, the demands it
// answers, the claims it stands on, the rows that prove it and the
// commands that re-run them. A page that cannot land or leave is a
// loud failure: a stale spec is a stale face.
func (m *machine) writeSpec(corpus doc.Document, claims map[string]claimsFacts) error {
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
		fmt.Fprintf(&page, "run: %d\n\n", m.standing.RunNumber)
		fmt.Fprintf(&page, "## Story\n\n%s\n\n", b.Prose)
		fmt.Fprintf(&page, "## Demands\n\n")
		for _, cover := range b.Covers {
			req, ok := m.standing.Registry.Requirements[cover]
			if !ok {
				continue
			}
			fmt.Fprintf(&page, "- %s (lines %s): %s\n", cover, strings.Join(req.Lines, ", "), req.Prose)
		}
		fmt.Fprintf(&page, "\n## Claims\n\n")
		for _, r := range claims[b.ID].ranges {
			fmt.Fprintf(&page, "- %s — %s\n", r.Claim, r.Digest12)
		}
		fmt.Fprintf(&page, "token: %s\n", fprint.Token12(claims[b.ID].digest))
		fmt.Fprintf(&page, "\n## Rows\n\n")
		for i, row := range b.Rows {
			if doc.Pinned(row.Expect) {
				fmt.Fprintf(&page, "## row %d — PROOF\n\n", i)
			} else {
				fmt.Fprintf(&page, "## row %d\n\n", i)
			}
			fmt.Fprintf(&page, "command: %s\n", commandOf(row.Run))
			if expect := expectOf(row.Expect); expect != "" {
				fmt.Fprintf(&page, "%s\n", expect)
			}
			fmt.Fprintf(&page, "proof: %s (%s)\n\n", m.rowStatus(b.ID, i, b), proofShort(m, b.ID, i))
		}
		fmt.Fprintf(&page, "## Verify\n\n")
		for _, row := range b.Rows {
			fmt.Fprintf(&page, "- %s\n", commandOf(row.Run))
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
// fact says what it says; an empty claims list keeps its rows
// honestly stale.
func (m *machine) rowStatus(id string, i int, b doc.Behavior) string {
	key := id + "/" + strconv.Itoa(i)
	fact, ok := m.standing.Rows[key]
	if !ok {
		return "stale"
	}
	if fact.Status == journal.StatusRed {
		return "red"
	}
	if len(b.Claims) == 0 {
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

// writeIfChanged — the view's whole content, only when it differs.
// A write that cannot land is a loud failure: the view is the
// machine's face, a stale face is a lie.
func (m *machine) writeIfChanged(path string, data []byte) error {
	if bytes.Equal(data, m.readFileOrEmpty(path)) {
		return nil
	}
	return ports.FSBind.WriteAtomic(path, data)
}

// shortDigest — the human-facing 12 characters of a digest: the
// protocol's own token spelling.
func shortDigest(d string) string {
	return fprint.Token12(d)
}
