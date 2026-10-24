// Package application — the machine: the one call's whole algorithm.
// The behavior is a pure function of the zone's content — find the
// zone, act on its state (first contact, birth, the wish slot, a
// submission), run the build, decide staleness by fingerprints,
// execute what is stale, journal the facts, close a verdict when the
// completion act meets a fully green corpus, and regenerate the
// changed views. The terminal receives exactly one pointer line;
// every piece of content is a file of the zone.
package application

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/neurophant/punchtape/internal/application/canon"
	"github.com/neurophant/punchtape/internal/domain"
	"github.com/neurophant/punchtape/internal/domain/doc"
	"github.com/neurophant/punchtape/internal/domain/fprint"
	"github.com/neurophant/punchtape/internal/domain/journal"
	"github.com/neurophant/punchtape/internal/domain/zone"
	"github.com/neurophant/punchtape/internal/infrastructure/store"
	"github.com/neurophant/punchtape/internal/infrastructure/store/cas"
	"github.com/neurophant/punchtape/internal/ports"
)

// Run — the one call. The exit code: 0 the machine worked (red rows
// are data, not an error), 1 the call did not happen (zone damage).
func Run(cwd string, terminal io.Writer) int {
	root, found := findZone(cwd)
	if !found {
		// First contact: the contract is the terminal's one
		// exception — no file channel exists yet, and nothing is
		// created.
		fmt.Fprint(terminal, canon.Load().Contract)
		return 0
	}
	zoneDir := filepath.Join(root, zone.Dirname)
	if _, err := ports.FSBind.Stat(filepath.Join(zoneDir, zone.JournalDir)); err != nil {
		if !os.IsNotExist(err) {
			fmt.Fprintf(terminal, "punchtape: zone damaged: %v\n", err)
			return 1
		}
		if _, werr := ports.FSBind.Stat(filepath.Join(zoneDir, zone.WishSlot)); werr != nil {
			fmt.Fprintf(terminal, "punchtape: no instance yet — put the human's wish verbatim into %s and run again\n", zone.Path(zone.WishSlot))
			return 0
		}
		if err := birth(zoneDir); err != nil {
			fmt.Fprintf(terminal, "punchtape: birth failed: %v\n", err)
			return 1
		}
	}

	m, err := open(root, zoneDir)
	if err != nil {
		fmt.Fprintf(terminal, "punchtape: %v\n", err)
		return 1
	}
	code := m.run(terminal)
	m.close()
	return code
}

// findZone — the working root: the nearest zone directory at or above
// the starting point.
func findZone(dir string) (string, bool) {
	for {
		if _, err := ports.FSBind.Stat(filepath.Join(dir, zone.Dirname)); err == nil {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

// machine — one open instance: the journal stream, the object store
// and the standing rebuilt from the journal.
type machine struct {
	root      string
	zoneDir   string
	stream    *store.Stream
	objects   *cas.Store
	standing  *journal.Standing
	cache     *HashCache
	pins      []string // the zone's env pins, parsed once, loud on a bad line
	completed bool     // the completion act rode this run
}

// open — the instance over its zone.
func open(root, zoneDir string) (*machine, error) {
	stream, err := store.OpenStream(filepath.Join(zoneDir, zone.JournalDir))
	if err != nil {
		return nil, err
	}
	records, err := stream.Records()
	if err != nil {
		return nil, err
	}
	objects := cas.New(filepath.Join(zoneDir, zone.ObjectsDir))
	standing, err := journal.Fold(records, objects.Get)
	if err != nil {
		return nil, fmt.Errorf("journal fold: %w", err)
	}
	pins, err := ZoneEnv(zoneDir)
	if err != nil {
		return nil, err
	}
	return &machine{
		root:     root,
		zoneDir:  zoneDir,
		stream:   stream,
		objects:  objects,
		standing: standing,
		pins:     pins,
		cache:    OpenHashCache(zoneDir),
	}, nil
}

// close — the run's cache flush.
func (m *machine) close() { m.cache.Flush() }

// corpusDoc — the standing corpus as a document.
func (m *machine) corpusDoc() doc.Document {
	d := doc.Document{}
	if m.standing.Build != "" {
		build := m.standing.Build
		d.Build = &build
	}
	for _, id := range m.standing.Registry.Order {
		d.Requirements = append(d.Requirements, m.standing.Registry.Requirements[id])
	}
	d.NotRequirements = m.standing.Registry.Markers
	d.Artifacts = m.standing.Registry.Artifacts
	for _, id := range m.standing.Order {
		d.Behaviors = append(d.Behaviors, m.standing.Behaviors[id])
	}
	return d
}

// run — the algorithm's steps past the birth.
func (m *machine) run(terminal io.Writer) int {
	// The submission slot.
	inbox := filepath.Join(m.zoneDir, zone.InboxSlot)
	if _, err := ports.FSBind.Stat(inbox); err == nil {
		if err := m.handleInbox(inbox); err != nil {
			if refusals, ok := err.(refusalList); ok {
				// The refusal is the machine's own journaled fact —
				// the journal first, the screens after it: a screen
				// that cannot land never eats the fact.
				if err := m.stream.Append(journal.Record{Kind: journal.KindRefusal, Refusals: refusals}); err != nil {
					fmt.Fprintf(terminal, "punchtape: %v\n", err)
					return 1
				}
				m.standing.Refused++
				if err := m.reportRefusals(refusals); err != nil {
					fmt.Fprintf(terminal, "punchtape: report view: %v\n", err)
					return 1
				}
				return m.finish(terminal, "submission refused — read "+zone.Path(zone.ReportFile)+", fix inbox.yaml in place, run again")
			}
			fmt.Fprintf(terminal, "punchtape: %v\n", err)
			return 1
		}
	}

	// The build: the run's first step when declared.
	corpus := m.corpusDoc()
	var buildFact *journal.Build
	if corpus.BuildCmd() != "" {
		res := RunCommand(Argv(doc.Cmd{Line: corpus.BuildCmd()}), m.root, RunEnv(m.pins))
		if res.Err != nil {
			buildFact = &journal.Build{Exit: -1, Reason: "launch: " + res.Err.Error()}
		} else {
			buildFact = &journal.Build{Exit: res.Exit, TimedOut: res.TimedOut}
		}
	}
	buildFailed := buildFact != nil && (buildFact.Exit != 0 || buildFact.TimedOut)

	// The fingerprints: the tree once, every behavior's scope once.
	tree, err := TreeFiles(m.root)
	if err != nil {
		fmt.Fprintf(terminal, "punchtape: %v\n", err)
		return 1
	}
	envHash := EnvHash(m.pins)
	scopes := map[string]string{}
	for _, b := range corpus.Behaviors {
		scopes[b.ID] = fprint.ScopeHash(ScopeDigests(m.root, tree, b.Scope, m.cache))
	}

	// Staleness by fingerprints; execution of the stale rows.
	var executed []journal.RowFact
	var rowReports []rowReport
	if !buildFailed {
		executed, rowReports = m.executeStale(corpus, scopes, envHash)
	}

	// The journal's run record, with the lessons the run fixed.
	changed := buildFact != nil || len(executed) > 0
	if changed {
		run := journal.Run{N: m.standing.RunNumber + 1, Build: buildFact, Rows: executed}
		run.Lessons = m.lessons(executed, run.N)
		if err := m.stream.Append(journal.Record{Kind: journal.KindRun, Run: &run}); err != nil {
			fmt.Fprintf(terminal, "punchtape: %v\n", err)
			return 1
		}
		m.standing.RunNumber = run.N
		for _, rf := range executed {
			m.standing.Rows[rf.Key] = rf
		}
		m.standing.Lessons = append(m.standing.Lessons, run.Lessons...)
	}

	// The verdict: the completion act's close, on a fully green
	// corpus over the current fingerprints. The report is the run's
	// whole picture — the facts and the verdict's outcome in one
	// rewrite, so no old word outlives its run.
	var verdictNotes []string
	if m.completed {
		var err error
		verdictNotes, err = m.closeVerdict(corpus, scopes, tree)
		if err != nil {
			fmt.Fprintf(terminal, "punchtape: %v\n", err)
			return 1
		}
		// The closing report names a healed drift by its own runs:
		// the catch the journal holds, the green that closed it now.
		if events, herr := m.historyEvents(); herr == nil {
			for i := len(events) - 1; i >= 0; i-- {
				if events[i].kind == "drift" {
					verdictNotes = append(verdictNotes, fmt.Sprintf(
						"drift healed: run %d caught, run %d green again",
						events[i].run, m.standing.RunNumber))
					break
				}
				// A heal of an older run already told its story; the
				// current run's own heal is the one being named now.
				if events[i].kind == "healed" && events[i].run < m.standing.RunNumber {
					break
				}
			}
		}
	}
	if changed || len(verdictNotes) > 0 {
		if err := m.writeReport(m.standing.RunNumber, buildFact, rowReports, verdictNotes); err != nil {
			fmt.Fprintf(terminal, "punchtape: report view: %v\n", err)
			return 1
		}
	}
	// The render's own fingerprints: the tree and every scope,
	// measured again after the execution — the state page's digests
	// and its moved-mark describe the tree as the run left it, not
	// as the run found it. The verdict's own record keeps the
	// launch-state pointers the proofs were measured against.
	renderTree, err := TreeFiles(m.root)
	if err != nil {
		fmt.Fprintf(terminal, "punchtape: %v\n", err)
		return 1
	}
	renderScopes := map[string]string{}
	for _, b := range corpus.Behaviors {
		renderScopes[b.ID] = fprint.ScopeHash(ScopeDigests(m.root, renderTree, b.Scope, m.cache))
	}
	if err := m.writeState(corpus, renderScopes, renderTree); err != nil {
		fmt.Fprintf(terminal, "punchtape: state view: %v\n", err)
		return 1
	}
	if err := m.writeKnowledge(); err != nil {
		fmt.Fprintf(terminal, "punchtape: knowledge view: %v\n", err)
		return 1
	}
	if err := m.writeSpec(corpus); err != nil {
		fmt.Fprintf(terminal, "punchtape: spec view: %v\n", err)
		return 1
	}
	return m.finish(terminal, m.nextInstruction(corpus, buildFailed))
}

// handleInbox — the submission slot: the completion act, an accepted
// document, or a refusal that changes nothing.
func (m *machine) handleInbox(inbox string) error {
	raw, err := ports.FSBind.ReadFile(inbox)
	if err != nil {
		return err
	}
	d, canonical, err := doc.Normalize(raw)
	if err != nil {
		return refusalList{{Where: zone.Path(zone.InboxSlot), Law: doc.LawDocument, Fact: err.Error()}}
	}
	if bytes.Equal(canonical, doc.Marshal(m.corpusDoc())) {
		// The completion act: the same bytes as the standing corpus —
		// the run continues its ordinary steps and closes the verdict.
		m.completed = true
		return ports.FSBind.Remove(inbox)
	}
	// The registry as it stands after this submission — the fact the
	// covers resolve against.
	preview := doc.MergeRegistry(m.standing.Registry, d)
	refusals := doc.Validate(d, &doc.Resolvers{
		Stands: func(id string) bool {
			_, behavior := m.standing.Behaviors[id]
			_, requirement := m.standing.Registry.Requirements[id]
			return behavior || requirement
		},
		Requirement: func(id string) bool {
			_, ok := preview.Requirements[id]
			return ok
		},
		WishLine: m.wishLine,
	})
	refusals = append(refusals, m.danglingCovers(d, preview)...)
	if len(refusals) > 0 {
		return refusalList(refusals)
	}
	if err := m.accept(d, canonical, preview); err != nil {
		return err
	}
	return ports.FSBind.Remove(inbox)
}

// wishLine — the W id's line of the wish's universe: its own text
// and whether it exists.
func (m *machine) wishLine(id string) (string, bool) {
	wish, err := m.objects.Get(m.standing.GoalAddr)
	if err != nil {
		return "", false
	}
	lines := wishLines(wish)
	nums, err := doc.ParseRanges([]string{id})
	if err != nil || len(nums) != 1 || nums[0] < 1 || nums[0] > len(lines) {
		return "", false
	}
	return lines[nums[0]-1], true
}

// danglingCovers — the covers of the behaviors that survive the
// submission without being listed in it: each must name a
// requirement of the post-merge registry. The submission's own
// behaviors are doc.Validate's own charge; the ones it does not
// touch are checked here, so retiring a requirement demands the
// same submission clears every cover of it.
func (m *machine) danglingCovers(d doc.Document, preview doc.Registry) []doc.Refusal {
	listed := map[string]bool{}
	for _, b := range d.Behaviors {
		listed[b.ID] = true
	}
	removed := map[string]bool{}
	for _, id := range d.Remove {
		removed[id] = true
	}
	var ids []string
	for id := range m.standing.Behaviors {
		if !listed[id] && !removed[id] {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	var out []doc.Refusal
	for _, id := range ids {
		for j, cover := range m.standing.Behaviors[id].Covers {
			if _, ok := preview.Requirements[cover]; !ok {
				out = append(out, doc.Refusal{
					Where: fmt.Sprintf("%s.covers[%d]", id, j),
					Law:   doc.LawDocument,
					Fact:  fmt.Sprintf("the behavior covers %q — no requirement %q stands after this submission", cover, cover),
				})
			}
		}
	}
	return out
}

// accept — the validated document into the journal: the objects first,
// then the submission's epoch, then the standing moves.
func (m *machine) accept(d doc.Document, canonical []byte, registry doc.Registry) error {
	docAddr, err := m.objects.Put(canonical)
	if err != nil {
		return err
	}
	sub := journal.Submission{
		CorpusVersion:   m.standing.CorpusVersion + 1,
		Build:           d.Build,
		Requirements:    d.Requirements,
		NotRequirements: d.NotRequirements,
		Artifacts:       d.Artifacts,
		Removed:         d.Remove,
		DocAddr:         docAddr,
	}
	// The document's own order, computed pure: removes retire first,
	// then the listed behaviors merge onto what remains standing — a
	// standing id merges (the head replaces, the rows merge by run,
	// the journaled object is the merged result), a remove-and-re-add
	// starts fresh. The fold replays the same order; the standing
	// itself moves only after the journal carries the submission.
	removed := make(map[string]bool, len(d.Remove))
	for _, rem := range d.Remove {
		removed[rem] = true
	}
	merged := make([]doc.Behavior, len(d.Behaviors))
	for i, b := range d.Behaviors {
		if standing, seen := m.standing.Behaviors[b.ID]; seen && !removed[b.ID] {
			b = doc.MergeBehavior(standing, b)
		}
		merged[i] = b
		addr, err := m.objects.Put(doc.MarshalBehavior(b))
		if err != nil {
			return err
		}
		sub.Added = append(sub.Added, journal.BehaviorOp{ID: b.ID, Addr: addr})
	}
	if err := m.stream.Rotate(); err != nil {
		return err
	}
	if err := m.stream.Append(journal.Record{Kind: journal.KindSubmission, Submission: &sub}); err != nil {
		return err
	}
	m.standing.CorpusVersion = sub.CorpusVersion
	if d.Build != nil {
		m.standing.Build = *d.Build
	}
	m.standing.Registry = registry
	for _, rem := range d.Remove {
		delete(m.standing.Behaviors, rem)
	}
	for _, b := range merged {
		if _, seen := m.standing.Behaviors[b.ID]; !seen {
			m.standing.Order = append(m.standing.Order, b.ID)
		}
		m.standing.Behaviors[b.ID] = b
	}
	// The order keeps only the living, each once, first-seen — the
	// same law the fold replays.
	ord := m.standing.Order[:0]
	written := map[string]bool{}
	for _, id := range m.standing.Order {
		if _, alive := m.standing.Behaviors[id]; alive && !written[id] {
			written[id] = true
			ord = append(ord, id)
		}
	}
	m.standing.Order = ord
	m.standing.Submissions = append(m.standing.Submissions, sub)
	return nil
}

// executeStale — the fingerprints' verdict per row, the execution of
// the stale ones.
func (m *machine) executeStale(corpus doc.Document, scopes map[string]string, envHash string) ([]journal.RowFact, []rowReport) {
	var facts []journal.RowFact
	var reports []rowReport
	for _, b := range corpus.Behaviors {
		scopeHash := scopes[b.ID]
		for i, row := range b.Rows {
			key := b.ID + "/" + strconv.Itoa(i)
			rowBytes := doc.MarshalRow(row)
			proof := fprint.Proof(rowBytes, scopeHash, envHash)
			// A behavior with rows and no scope stands on nothing it
			// names: its rows are honestly always stale.
			fact, ok := m.standing.Rows[key]
			stale := !ok || fact.Proof != proof || len(b.Scope) == 0
			if !stale {
				continue
			}
			res := RunCommand(Argv(row.Run), m.root, RunEnv(m.pins))
			rp := rowReport{Key: key, Run: row.Run}
			var broken []string
			switch {
			case res.Err != nil:
				broken = []string{"launch: " + res.Err.Error()}
				rp.Exit = -1
			case res.TimedOut:
				broken = []string{fmt.Sprintf("timed out after %s", RowTimeout)}
				rp.Exit = res.Exit
			default:
				rp.Exit = res.Exit
				rp.Stdout = res.Stdout
				broken = doc.Compare(row.Expect, res.Exit, res.Stdout, func(p string) ([]byte, error) {
					return ports.FSBind.ReadFile(filepath.Join(m.root, filepath.FromSlash(p)))
				})
			}
			// A command the row could not find is the environment's
			// own story: the row carries the machine's law with it —
			// inheritance, pins, relative paths — so the fix needs
			// no knowledge of any world.
			if res.Err == nil && !res.TimedOut &&
				(res.Exit == 127 || strings.Contains(res.Stderr, "command not found")) {
				broken = append(broken,
					"hint: the command was not found in the row's PATH — rows run with the "+
						"PATH and HOME of the shell that runs the machine plus the "+
						".punchtape/env pins; make the command reachable from that shell, "+
						"pin it (KEY=VALUE, relative paths only) or call project files by "+
						"explicit relative path")
			}
			status := journal.StatusGreen
			if len(broken) > 0 || res.Err != nil || res.TimedOut {
				status = journal.StatusRed
			}
			facts = append(facts, journal.RowFact{
				Key: key, Row: domain.Digest(rowBytes), Status: status, Exit: rp.Exit,
				Facts: broken, Proof: proof,
			})
			rp.Facts = broken
			reports = append(reports, rp)
		}
	}
	return facts, reports
}

// proven — whether every row of the corpus stands green on the
// current fingerprints. The caller has already refused the empty
// corpus.
func (m *machine) proven(corpus doc.Document, scopes map[string]string, envHash string) bool {
	for _, b := range corpus.Behaviors {
		for i, row := range b.Rows {
			key := b.ID + "/" + strconv.Itoa(i)
			fact, ok := m.standing.Rows[key]
			if !ok || fact.Status != journal.StatusGreen {
				return false
			}
			if fact.Proof != fprint.Proof(doc.MarshalRow(row), scopes[b.ID], envHash) {
				return false
			}
		}
	}
	return true
}

// lessons — the red→green transitions of this run, fixed as data:
// the reason, the broken facts, the row's bytes before and after,
// the two run numbers. Zero tokens of any model — the machine
// measured the fix itself.
func (m *machine) lessons(executed []journal.RowFact, greenRun int) []journal.Lesson {
	var out []journal.Lesson
	for _, rf := range executed {
		if rf.Status != journal.StatusGreen {
			continue
		}
		prev, ok := m.standing.Rows[rf.Key]
		if !ok || prev.Status != journal.StatusRed || len(prev.Facts) == 0 {
			continue
		}
		out = append(out, journal.Lesson{
			Reason:   prev.Facts[0],
			Key:      rf.Key,
			Facts:    prev.Facts,
			RedRow:   string(m.rowBytesOf(prev.Row)),
			GreenRow: string(m.rowBytesOf(rf.Row)),
			RedRun:   m.standing.RunNumber, // the red stood at the previous run
			GreenRun: greenRun,
		})
	}
	return out
}

// rowBytesOf — the row bytes a digest names, recovered from the
// behavior objects the submissions carry.
func (m *machine) rowBytesOf(rowDigest string) []byte {
	for _, sub := range m.standing.Submissions {
		for _, op := range sub.Added {
			data, err := m.objects.Get(op.Addr)
			if err != nil {
				continue
			}
			// The fold already decoded every behavior object of the
			// standing at open, and the store verified the digest:
			// decoding here cannot fail — a zero behavior simply
			// carries no rows.
			b, _ := doc.DecodeBehavior(data)
			for _, row := range b.Rows {
				if domain.Digest(doc.MarshalRow(row)) == rowDigest {
					return doc.MarshalRow(row)
				}
			}
		}
	}
	return fmt.Appendf(nil, "<row %s>", rowDigest)
}

// closeVerdict — the completion act's close: the verdict stands only
// over a corpus that is green on the current fingerprints, a wish
// accounted whole — every line inside some requirement's span or
// marked not-a-requirement —, every standing requirement named in
// the covers of a standing behavior, and no uncovered source file.
// Anything short says so in the report, one line per class of hole —
// incompleteness is not a refusal of the submission, the door closes
// once, here.
func (m *machine) closeVerdict(corpus doc.Document, scopes map[string]string, tree []string) ([]string, error) {
	envHash := EnvHash(m.pins)
	if len(corpus.Behaviors) == 0 {
		return []string{"verdict: refused — the corpus is empty"}, nil
	}
	if !m.proven(corpus, scopes, envHash) {
		return []string{"verdict: not yet — a row is red or stale"}, nil
	}
	var notes []string
	if brief, ok := m.briefAccounting(corpus); ok && len(brief.Holes) > 0 {
		notes = append(notes, fmt.Sprintf("verdict: not yet — the wish lines %s stand uncovered by any requirement or marker", doc.SpellRanges(brief.Holes)))
	}
	if orphans := m.orphanRequirements(corpus); len(orphans) > 0 {
		notes = append(notes, fmt.Sprintf("verdict: not yet — the requirements %s stand uncovered by any behavior's covers", strings.Join(orphans, ", ")))
	}
	if _, sources := SplitUncovered(tree, m.allScopes(corpus), corpus.Artifacts); len(sources) > 0 {
		notes = append(notes, fmt.Sprintf("verdict: not yet — the sources %s stand uncovered", strings.Join(sources, ", ")))
	}
	if len(notes) > 0 {
		return notes, nil
	}
	v := journal.Verdict{
		N:              verdictNumber(m.standing) + 1,
		CorpusVersion:  m.standing.CorpusVersion,
		Signature:      m.signature(corpus, scopes),
		RegistryDigest: m.registryDigest(),
		Run:            m.standing.RunNumber,
	}
	ids := make([]string, 0, len(corpus.Behaviors))
	for _, b := range corpus.Behaviors {
		ids = append(ids, b.ID)
	}
	sort.Strings(ids)
	for _, id := range ids {
		v.Scopes = append(v.Scopes, journal.VerdictScope{ID: id, Pointer: scopes[id]})
	}
	if err := m.stream.Append(journal.Record{Kind: journal.KindVerdict, Verdict: &v}); err != nil {
		return nil, err
	}
	// The act that closed it is a fact of its own, journaled beside
	// the verdict: the render's arithmetic counts submissions,
	// refusals and completions each by its own name.
	if err := m.stream.Append(journal.Record{
		Kind:      journal.KindCompletion,
		Completed: &journal.Completion{Run: m.standing.RunNumber},
	}); err != nil {
		return nil, err
	}
	m.standing.Completions++
	m.standing.Verdict = &v
	return []string{fmt.Sprintf("verdict: #%d on v%d, run %d recorded — signature %s", v.N, v.CorpusVersion, v.Run, shortDigest(v.Signature))}, nil
}

// orphanRequirements — the standing requirements no standing behavior
// names in its covers, in registry order. Every standing behavior
// carries rows by the shape laws — a cover is a cover of proof.
func (m *machine) orphanRequirements(corpus doc.Document) []string {
	covers := map[string]bool{}
	for _, b := range corpus.Behaviors {
		for _, c := range b.Covers {
			covers[c] = true
		}
	}
	var out []string
	for _, id := range m.standing.Registry.Order {
		if !covers[id] {
			out = append(out, id)
		}
	}
	return out
}

// registryDigest — the wish's account as one digest: the canonical
// bytes of the standing requirements and markers together. Any
// change to either moves the digest the verdict's record carries.
func (m *machine) registryDigest() string {
	d := doc.Document{}
	for _, id := range m.standing.Registry.Order {
		d.Requirements = append(d.Requirements, m.standing.Registry.Requirements[id])
	}
	d.NotRequirements = m.standing.Registry.Markers
	return domain.Digest(doc.Marshal(d))
}

// verdictNumber — the standing verdict's number, zero before any.
func verdictNumber(st *journal.Standing) int {
	if st.Verdict == nil {
		return 0
	}
	return st.Verdict.N
}

// signature — the verdict's signature: the corpus's canonical bytes
// and every behavior's scope fingerprint.
func (m *machine) signature(corpus doc.Document, scopes map[string]string) string {
	var b bytes.Buffer
	b.Write(doc.Marshal(corpus))
	for _, beh := range corpus.Behaviors {
		b.WriteByte(0)
		b.WriteString(beh.ID)
		b.WriteByte(0)
		b.WriteString(scopes[beh.ID])
	}
	return domain.Digest(b.Bytes())
}

// finish — the run's last steps: the corpus image heals, the AGENTS
// block stands, the scratch clears, the terminal gets its one line.
// A machine-owned write that cannot land is a loud failure: the run
// says so and answers one.
func (m *machine) finish(terminal io.Writer, instruction string) int {
	if canonical := doc.Marshal(m.corpusDoc()); !bytes.Equal(canonical, m.readFileOrEmpty(filepath.Join(m.zoneDir, zone.CorpusFile))) {
		if err := ports.FSBind.WriteAtomic(filepath.Join(m.zoneDir, zone.CorpusFile), canonical); err != nil {
			fmt.Fprintf(terminal, "punchtape: corpus view: %v\n", err)
			return 1
		}
	}
	if err := m.ensureAgentsBlock(); err != nil {
		fmt.Fprintf(terminal, "punchtape: %v\n", err)
		return 1
	}
	if err := ports.FSBind.RemoveAll(filepath.Join(m.zoneDir, zone.TmpDir)); err != nil {
		fmt.Fprintf(terminal, "punchtape: tmp clear: %v\n", err)
		return 1
	}
	fmt.Fprintf(terminal, "punchtape: run %d — %s\n", m.standing.RunNumber, instruction)
	return 0
}

// nextInstruction — the one imperative that answers "what do I do
// now" from the standing alone: every state the machine can stand in
// has exactly one instruction, and it names the file to touch. The
// terminal's pointer line and state.md's next: line are this one
// text — never two truths.
func (m *machine) nextInstruction(corpus doc.Document, buildFailed bool) string {
	if buildFailed {
		return "build red — read " + zone.Path(zone.ReportFile) + ", fix the build, run again"
	}
	red, stale := m.countRows(corpus)
	switch {
	case len(corpus.Behaviors) == 0:
		return "corpus empty — submit the first document as " + zone.Path(zone.InboxSlot) + " (copy " + zone.Path(zone.TemplateFile) + ")"
	case red > 0:
		return fmt.Sprintf("%d row(s) red — read %s, fix, run again", red, zone.Path(zone.ReportFile))
	case stale > 0:
		return fmt.Sprintf("%d row(s) stale — run again to execute them", stale)
	case m.standing.Verdict != nil && m.standing.Verdict.CorpusVersion == m.standing.CorpusVersion:
		return fmt.Sprintf("verdict #%d standing — show %s to the human; new work goes to %s as a document",
			m.standing.Verdict.N, zone.Path(zone.StateFile), zone.Path(zone.InboxSlot))
	default:
		return "all green — copy " + zone.Path(zone.CorpusFile) + " to " + zone.Path(zone.InboxSlot) + " and run to record the verdict"
	}
}

// readFileOrEmpty — a missing file reads as empty.
func (m *machine) readFileOrEmpty(path string) []byte {
	data, err := ports.FSBind.ReadFile(path)
	if err != nil {
		return nil
	}
	return data
}

// refusalList — the submission's refusals as one error.
type refusalList []doc.Refusal

// Error — the refusals, one line each.
func (r refusalList) Error() string {
	var sb bytes.Buffer
	for i, ref := range r {
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(ref.Error())
	}
	return sb.String()
}
