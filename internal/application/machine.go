// Package application — the machine: the one call's whole algorithm.
// The behavior is a pure function of the zone's content — find the
// zone, act on its state (first contact, birth, the demand stream, a
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
	"time"

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
	// One zone, one run at a time: the lock serializes every pass —
	// the birth included — and the kernel frees it when a holder
	// dies mid-run. A parallel caller waits its turn; local
	// filesystems only.
	release, err := ports.FSBind.Flock(filepath.Join(zoneDir, zone.RunLockFile))
	if err != nil {
		fmt.Fprintf(terminal, "punchtape: zone lock: %v\n", err)
		return 1
	}
	defer func() {
		if rerr := release(); rerr != nil {
			fmt.Fprintf(terminal, "punchtape: zone lock release: %v\n", rerr)
		}
	}()
	if _, err := ports.FSBind.Stat(filepath.Join(zoneDir, zone.JournalDir)); err != nil {
		if !os.IsNotExist(err) {
			fmt.Fprintf(terminal, "punchtape: zone damaged: %v\n", err)
			return 1
		}
		// The contract prints whenever no journal stands: the first
		// record not existing is the first contact, whatever else
		// the zone directory carries.
		fmt.Fprint(terminal, canon.Load().Contract)
		// The birth waits for the stream's first non-empty line: the
		// file may stand empty — no words, no instance, nothing
		// created.
		words, rerr := ports.FSBind.ReadFile(filepath.Join(zoneDir, zone.DemandsSlot))
		if rerr != nil && !os.IsNotExist(rerr) {
			fmt.Fprintf(terminal, "punchtape: birth failed: %v\n", rerr)
			return 1
		}
		if rerr != nil || len(doc.NonEmptyLines(string(words))) == 0 {
			fmt.Fprintf(terminal, "punchtape: no instance yet — put the human's words verbatim into %s and run again\n", slotFrom(cwd, root, zone.DemandsSlot))
			return 0
		}
		if err := birth(zoneDir, words); err != nil {
			fmt.Fprintf(terminal, "punchtape: birth failed: %v\n", err)
			return 1
		}
	}

	stream, err := store.OpenStream(filepath.Join(zoneDir, zone.JournalDir))
	if err != nil {
		fmt.Fprintf(terminal, "punchtape: %v\n", err)
		return 1
	}
	// The zone's own creator, named before any record folds: an
	// instance born by another machine version is not this machine's
	// to read — the owner archives it and starts fresh.
	first, err := stream.FirstRecord()
	if err != nil {
		fmt.Fprintf(terminal, "punchtape: %v\n", err)
		return 1
	}
	if first.Kind == journal.KindBirth && first.Birth != nil && first.Birth.Version != Version {
		fmt.Fprintf(terminal, "punchtape: zone born by machine %s — this machine is %s; archive it (cp -r .punchtape .punchtape-archive) and start fresh\n",
			first.Birth.Version, Version)
		return 1
	}
	m, err := open(cwd, root, zoneDir, stream)
	if err != nil {
		fmt.Fprintf(terminal, "punchtape: %v\n", err)
		return 1
	}
	return m.run(terminal)
}

// slotFrom — a zone file's name as the caller sees it from the
// working directory: the relative path from the caller — up the tree
// and into the zone — whenever the spelling is reachable, the
// absolute path otherwise. The instructions speak this name — an
// obedient reader lands in the zone from any depth of the tree.
func slotFrom(cwd, root, name string) string {
	p := filepath.Join(root, zone.Dirname, name)
	rel, err := filepath.Rel(cwd, p)
	if err == nil {
		return filepath.ToSlash(rel)
	}
	return p
}

// slot — the instruction's name for a zone file, from the caller's
// own working directory.
func (m *machine) slot(name string) string {
	return slotFrom(m.cwd, m.root, name)
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

// machine — one open instance: the journal stream, the object store,
// the standing rebuilt from the journal, the demand stream's
// numbered universe (loaded once by the run's chain check) and the
// run's one coverage fact over the tree. The journal's records,
// read once at the open, ride the whole run.
type machine struct {
	root        string
	cwd         string // the caller's working directory: the instructions' paths name the zone from here
	zoneDir     string
	stream      *store.Stream
	objects     *cas.Store
	standing    *journal.Standing
	universe    []string // the stream's non-empty lines, numbered D1..Dn
	covered     *fprint.Coverage
	pins        []string // the zone's env pins, parsed once, loud on a bad line
	rowTimeout  time.Duration
	runBudget   time.Duration // the run's own budget: what it cannot fit waits the next run
	spent       time.Duration // the budget already spent on this run's own commands
	completed   bool          // the completion act rode this run
	buildFailed bool          // the declared build stood red this run
	records     []journal.Record
	submitted   *doc.Document // the accepted submission, when one rode
	execFacts   []journal.ExecutionFact
	treeBefore  string // the tree's digest ahead of a locked submission
	warnings    []string
	tree        *treeScope // the run's walk scope: the declared universe and the verified cache
}

// open — the instance over its zone and its already-open journal.
func open(cwd, root, zoneDir string, stream *store.Stream) (*machine, error) {
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
	rowTimeout, err := RowTimeoutOf(pins)
	if err != nil {
		return nil, err
	}
	runBudget, err := RunBudgetOf(pins)
	if err != nil {
		return nil, err
	}
	tree, err := newTreeScope(root, zoneDir)
	if err != nil {
		return nil, err
	}
	m := &machine{
		root:       root,
		cwd:        cwd,
		zoneDir:    zoneDir,
		stream:     stream,
		objects:    objects,
		standing:   standing,
		pins:       pins,
		rowTimeout: rowTimeout,
		runBudget:  runBudget,
		records:    records,
		tree:       tree,
	}
	return m, nil
}

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
	d.Markers = m.standing.Registry.Markers
	d.Exclusions = m.standing.Registry.Exclusions
	d.Open = m.standing.Registry.Open
	for _, id := range m.standing.Order {
		d.Behaviors = append(d.Behaviors, m.standing.Behaviors[id])
	}
	return d
}

// run — the algorithm's steps past the birth.
func (m *machine) run(terminal io.Writer) int {
	// The chain: every run recomputes it over the stream's current
	// bytes — a moved past line is zone damage, the run never starts.
	if err := m.checkChain(); err != nil {
		fmt.Fprintf(terminal, "punchtape: zone damaged: %v\n", err)
		return 1
	}

	// The submission slot. Ahead of a locked submission the tree's
	// own digest stands: what the run found, against what it leaves.
	// The inbox is read and parsed once — the lock's question and
	// the submission's own handling see the same document.
	inbox := filepath.Join(m.zoneDir, zone.InboxSlot)
	if info, err := ports.FSBind.Stat(inbox); err == nil {
		// The document is bounded ahead of its read: a giant inbox
		// is refused before a byte of it enters the machine, and
		// the read's own length answers the same refusal when the
		// file grew between the look and the read.
		oversized := func(n int) error {
			return fmt.Errorf("the document is bounded — %d bytes, this one carries %d", doc.InboxBytes, n)
		}
		var normErr error
		if info.Size() > doc.InboxBytes {
			normErr = oversized(int(info.Size()))
		}
		var raw []byte
		if normErr == nil {
			var rerr error
			raw, rerr = ports.FSBind.ReadFile(inbox)
			if rerr != nil {
				fmt.Fprintf(terminal, "punchtape: %v\n", rerr)
				return 1
			}
			if len(raw) > doc.InboxBytes {
				normErr = oversized(len(raw))
			}
		}
		d, canonical, parseErr := doc.Normalize(raw)
		if parseErr != nil {
			d = doc.Document{}
		}
		if normErr == nil {
			normErr = parseErr
		}
		if m.treeLockArmed(d) {
			facts, derr := m.tree.walk()
			if derr != nil {
				fmt.Fprintf(terminal, "punchtape: %v\n", derr)
				return 1
			}
			m.treeBefore = facts.digest
		}
		if err := m.handleInbox(inbox, d, canonical, normErr); err != nil {
			if refusals, ok := err.(refusalList); ok {
				// The refusal is the machine's own journaled fact —
				// the journal first, the screens after it: a screen
				// that cannot land never eats the fact.
				if err := m.append(journal.Record{Kind: journal.KindRefusal, Refusals: refusals}); err != nil {
					fmt.Fprintf(terminal, "punchtape: %v\n", err)
					return 1
				}
				m.standing.Refused++
				if err := m.reportRefusals(refusals); err != nil {
					fmt.Fprintf(terminal, "punchtape: report view: %v\n", err)
					return 1
				}
				// The queue is a fact of the files: the refused
				// submission stands in the inbox — the state says so,
				// in the same words the terminal speaks.
				corpus := m.corpusDoc()
				claims := m.claimsMap(corpus)
				instruction := "submission refused — read " + m.slot(zone.ReportFile) + ", fix inbox.yaml in place, run again"
				if err := m.writeState(corpus, claims, instruction); err != nil {
					fmt.Fprintf(terminal, "punchtape: state view: %v\n", err)
					return 1
				}
				return m.finish(terminal, instruction)
			}
			fmt.Fprintf(terminal, "punchtape: %v\n", err)
			return 1
		}
	}

	// The declared executions, run by the machine itself between the
	// submission and the rows: their facts ride the journal, the
	// tree must come back unchanged.
	if m.submitted != nil {
		if err := m.runExecutions(); err != nil {
			fmt.Fprintf(terminal, "punchtape: %v\n", err)
			return 1
		}
	}

	// The build: the run's first step when declared.
	corpus := m.corpusDoc()
	var buildFact *journal.Build
	if corpus.BuildCmd() != "" {
		res := RunCommand(Argv(doc.Cmd{Line: corpus.BuildCmd()}), m.root, RunEnv(m.pins), m.rowTimeout)
		if res.Err != nil {
			buildFact = &journal.Build{Exit: -1, Reason: "launch: " + res.Err.Error()}
		} else {
			buildFact = &journal.Build{Exit: res.Exit, TimedOut: res.TimedOut}
		}
	}
	buildFailed := buildFact != nil && (buildFact.Exit != 0 || buildFact.TimedOut)
	m.buildFailed = buildFailed

	// The fingerprints: the claims once ahead of the execution —
	// every behavior's digest once, the proofs' own measurement.
	envHash := EnvHash(m.pins)
	claimsBefore := m.claimsMap(corpus)

	// Staleness by fingerprints; execution of the stale rows.
	var executed []journal.RowFact
	var rowReports []rowReport
	if !buildFailed {
		executed, rowReports = m.executeStale(corpus, claimsBefore, envHash)
	}

	// The tree lock: a locked submission leaves the tree as it found
	// it — a move under answer, execute or describe intents is a
	// journaled fact that holds the verdict until the tree returns
	// or the change becomes someone's declared work.
	if m.treeBefore != "" {
		facts, derr := m.tree.walk()
		if derr != nil {
			fmt.Fprintf(terminal, "punchtape: %v\n", derr)
			return 1
		}
		if facts.digest != m.treeBefore {
			fact := &journal.TreeLock{Before: m.treeBefore, After: facts.digest}
			if err := m.append(journal.Record{Kind: journal.KindTreeLock, TreeLock: fact}); err != nil {
				fmt.Fprintf(terminal, "punchtape: %v\n", err)
				return 1
			}
			m.standing.TreeLockViolation = fact
		}
		m.treeBefore = ""
	}

	// The journal's run record, with the lessons the run fixed.
	changed := buildFact != nil || len(executed) > 0 || len(m.execFacts) > 0
	if changed {
		run := journal.Run{N: m.standing.RunNumber + 1, Build: buildFact, Rows: executed}
		run.Lessons = m.lessons(executed, run.N)
		if err := m.append(journal.Record{Kind: journal.KindRun, Run: &run}); err != nil {
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
	// whole picture — the facts, the warnings and the verdict's
	// outcome in one rewrite, so no old word outlives its run. The
	// render's own fingerprints are measured once, after the
	// execution — the tree as the run left it.
	claimsAfter := m.claimsMap(corpus)
	var verdictNotes []string
	if m.completed {
		var err error
		verdictNotes, err = m.closeVerdict(corpus, claimsBefore, claimsAfter)
		if err != nil {
			fmt.Fprintf(terminal, "punchtape: %v\n", err)
			return 1
		}
		// The closing report names a healed drift by its own runs:
		// the catch the journal holds, the green that closed it now.
		events := historyEvents(m.records)
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
		// The journal carries the act's outcome — only now does the
		// document leave the slot. A crash before this point leaves
		// the done document standing, and the next run takes the
		// same act again; a re-read of an act whose identical verdict
		// already stands consumes quietly instead of recording a
		// second point of the same truth.
		if err := ports.FSBind.Remove(inbox); err != nil {
			fmt.Fprintf(terminal, "punchtape: %v\n", err)
			return 1
		}
	}
	if changed || len(verdictNotes) > 0 || len(m.warnings) > 0 {
		notes := append(append([]string(nil), m.warnings...), verdictNotes...)
		if err := m.writeReport(m.standing.RunNumber, buildFact, rowReports, notes); err != nil {
			fmt.Fprintf(terminal, "punchtape: report view: %v\n", err)
			return 1
		}
	}
	// The render's own fingerprints: the claims measured again after
	// the execution — the state page's digests describe the tree as
	// the run left it, not as the run found it. The verdict's own
	// record keeps the pointers the proofs were measured against.
	// The instruction is computed once, here — the state's next line
	// and the terminal's pointer line are this one string.
	instruction := m.nextInstruction(corpus, claimsBefore, claimsAfter)
	if err := m.writeState(corpus, claimsAfter, instruction); err != nil {
		fmt.Fprintf(terminal, "punchtape: state view: %v\n", err)
		return 1
	}
	if err := m.writeTrace(corpus, claimsAfter); err != nil {
		fmt.Fprintf(terminal, "punchtape: trace view: %v\n", err)
		return 1
	}
	if err := m.writeExclusionsProposal(corpus); err != nil {
		fmt.Fprintf(terminal, "punchtape: proposal view: %v\n", err)
		return 1
	}
	if err := m.writeKnowledge(); err != nil {
		fmt.Fprintf(terminal, "punchtape: knowledge view: %v\n", err)
		return 1
	}
	if err := m.writeSpec(corpus, claimsAfter); err != nil {
		fmt.Fprintf(terminal, "punchtape: spec view: %v\n", err)
		return 1
	}
	return m.finish(terminal, instruction)
}

// treeLockArmed — whether the inbox's submission, when it lands,
// locks the tree: declared intents exist, none of them is work, and
// no build stands or is declared — a build is a work attribute, its
// artifacts are legal, the first-declared build as much as the
// standing one. The document arrives already parsed — the run's
// one read.
func (m *machine) treeLockArmed(d doc.Document) bool {
	if len(d.Intents) == 0 {
		return false
	}
	for _, in := range d.Intents {
		if in.Outcome == doc.OutcomeWork {
			return false
		}
	}
	return m.standing.Build == "" && d.Build == nil
}

// handleInbox — the submission slot: the completion act, an accepted
// document, or a refusal that changes nothing. The document arrives
// already read and parsed — the run's one read; a broken parse is
// the refusal's own fact.
func (m *machine) handleInbox(inbox string, d doc.Document, canonical []byte, normErr error) error {
	if normErr != nil {
		return refusalList{{Where: m.slot(zone.InboxSlot), Law: doc.LawDocument, Fact: normErr.Error()}}
	}
	// The completion act goes first: done: true and nothing else.
	// Any other field beside the flag is the shape law's own refusal;
	// the bare act closes the run over the standing corpus.
	if d.Done {
		if refusals := doc.Validate(d, nil); len(refusals) > 0 {
			return refusalList(refusals)
		}
		// The act rides this run; the inbox leaves only after the
		// journal carries the act's outcome — a run that dies before
		// that leaves the document in the slot, and the next run
		// takes the same act again. Never a vanished word.
		m.completed = true
		return nil
	}
	// The document's D ranges meet the stream's own size before any
	// merge keys or stamps parse them: a spelling past the last line
	// is refused here, the same law the validation's own gates carry.
	if overflows := m.streamRangeRefusals(d); len(overflows) > 0 {
		return refusalList(overflows)
	}
	// The registry as it stands after this submission — the fact the
	// covers resolve against.
	preview := doc.MergeRegistry(m.standing.Registry, d)
	if len(d.Open) > 0 {
		stamped := journal.StampOpen(d.Open, m.standing.Registry.Open, m.standing.PastLines, m.standing.AskedLines)
		preview.Open = stamped
		d.Open = stamped
	}
	var previewClaims []string
	for _, b := range d.Behaviors {
		previewClaims = append(previewClaims, b.Claims...)
	}
	for _, b := range m.standing.Behaviors {
		if !containsID(d, b.ID) {
			previewClaims = append(previewClaims, b.Claims...)
		}
	}
	// One range, one behavior: a claim a standing behavior holds or
	// an earlier behavior of this same submission holds is refused
	// at its second holder, the holder named.
	var claimRefusals []doc.Refusal
	holders := map[string]string{}
	for _, id := range m.standing.Order {
		if containsID(d, id) {
			continue
		}
		for _, claim := range m.standing.Behaviors[id].Claims {
			holders[claim] = id
		}
	}
	for i, b := range d.Behaviors {
		for j, claim := range b.Claims {
			if holder, held := holders[claim]; held && holder != b.ID {
				claimRefusals = append(claimRefusals, doc.Refusal{
					Where: fmt.Sprintf("behaviors[%d].claims[%d]", i, j),
					Law:   doc.LawClaims,
					Fact:  fmt.Sprintf("the claim %s is already held by %s — one range, one behavior", claim, holder),
				})
				continue
			}
			holders[claim] = b.ID
		}
	}
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
		DemandLine: m.demandLine,
		FileLines:  m.FileLines,
		Law:        lawPrinted,
		Hole: func(rel string, first, last int) bool {
			return m.Hole(previewClaims, rel, first, last)
		},
		MarkerStands: func(line int) bool {
			for _, marker := range m.standing.Registry.Markers {
				if nums, err := doc.ParseRanges(marker.Lines); err == nil {
					for _, n := range nums {
						if n == line {
							return true
						}
					}
				}
			}
			return false
		},
		OrderStands: func(line int) bool {
			for _, order := range m.standing.Registry.Open {
				if n := orderLineNumber(order); n == line {
					return true
				}
			}
			return false
		},
		ExclusionStands: func(rel string) bool {
			for _, ex := range m.standing.Registry.Exclusions {
				if ex.Range == rel {
					return true
				}
			}
			return false
		},
		RequirementLines: func(id string) ([]int, bool) {
			// The ids come from the preview's own order — the preview
			// carries submitted requirements too; the parse is bounded
			// by the stream the machine carries.
			req := preview.Requirements[id]
			nums, _ := doc.ParseRangesWithin(req.Lines, len(m.universe))
			return nums, true
		},
		RequirementIDs: func() []string {
			return preview.Order
		},
		DemandsTailLines: func() []int {
			var out []int
			for n := m.standing.PastLines + 1; n <= len(m.universe); n++ {
				out = append(out, n)
			}
			return out
		},
		UniverseSize: func() int { return len(m.universe) },
		LineOutcome: func(id string) (string, bool) {
			// The callers spell single demand ids; the parse is total
			// by construction.
			nums, _ := doc.ParseRanges([]string{id})
			lo, ok := m.standing.LineOutcomes[nums[0]]
			return lo.Outcome, ok
		},
		RunsInUse: func(runKey string) (string, bool) {
			removed := map[string]bool{}
			for _, id := range d.Remove {
				removed[id] = true
			}
			for _, id := range m.standing.Order {
				// A listed behavior's rows are its own to merge; a
				// removed behavior releases its runs.
				if containsID(d, id) || removed[id] {
					continue
				}
				for _, row := range m.standing.Behaviors[id].Rows {
					if row.Run.Key() == runKey {
						return id, true
					}
				}
			}
			return "", false
		},
	})
	refusals = append(refusals, claimRefusals...)
	refusals = append(refusals, m.danglingCovers(d, preview)...)
	refusals = append(refusals, m.askCloseRefusals(d, preview)...)
	if len(refusals) > 0 {
		return refusalList(refusals)
	}
	if err := m.accept(d, canonical, preview); err != nil {
		return err
	}
	m.submitted = &d
	return ports.FSBind.Remove(inbox)
}

// streamRangeRefusals — the submitted D spellings that meet a parser
// ahead of the validation's own gates (the preview's merge keys and
// the orders' stamps): each is bounded by the stream the machine
// carries; an overflow answers its own refusal, a broken spelling
// flows on to the gates that own its voice.
func (m *machine) streamRangeRefusals(d doc.Document) []doc.Refusal {
	var out []doc.Refusal
	overflow := func(where string, err error) {
		if _, ok := err.(*doc.RangeError); ok {
			out = append(out, doc.Refusal{Where: where, Law: doc.LawDocument, Fact: err.Error()})
		}
	}
	for i, marker := range d.Markers {
		for j, list := range marker.Lines {
			_, err := doc.ParseRangesWithin([]string{list}, len(m.universe))
			overflow(fmt.Sprintf("not-a-requirement[%d].lines[%d]", i, j), err)
		}
	}
	for i, key := range d.Remove {
		_, err := doc.ParseRangesWithin([]string{key}, len(m.universe))
		overflow(fmt.Sprintf("remove[%d]", i), err)
	}
	for i, order := range d.Open {
		_, err := doc.ParseRangesWithin([]string{order.Demand}, len(m.universe))
		overflow(fmt.Sprintf("open[%d].demand", i), err)
	}
	return out
}

// orderLineNumber — the order's demand line number, zero when
// unparsable.
func orderLineNumber(order doc.OpenOrder) int {
	nums, err := doc.ParseRanges([]string{order.Demand})
	if err != nil || len(nums) == 0 {
		return 0
	}
	return nums[0]
}

// askCloseRefusals — the awaiting order's own law: an order closes
// only over a later human line, cited by a decision note. Under the
// merge-by-key law an unlisted order stands — a close is a removal
// of its line or a replacement that leaves the awaiting state.
func (m *machine) askCloseRefusals(d doc.Document, preview doc.Registry) []doc.Refusal {
	retired := doc.ParseRemovals(d.Remove)
	stream := len(m.universe)
	occupied := func(n int) bool {
		for _, id := range preview.Order {
			if nums, err := doc.ParseRangesWithin(preview.Requirements[id].Lines, stream); err == nil {
				for _, k := range nums {
					if k == n {
						return true
					}
				}
			}
		}
		for _, marker := range preview.Markers {
			if nums, err := doc.ParseRangesWithin(marker.Lines, stream); err == nil {
				for _, k := range nums {
					if k == n {
						return true
					}
				}
			}
		}
		for _, in := range d.Intents {
			if in.Outcome != doc.OutcomeAnswer && in.Outcome != doc.OutcomeExecute {
				continue
			}
			if nums, err := doc.ParseRangesWithin(in.Lines, stream); err == nil {
				for _, k := range nums {
					if k == n {
						return true
					}
				}
			}
		}
		return false
	}
	cited := map[int]bool{}
	for _, note := range d.Notes {
		if note.Kind != "decision" {
			continue
		}
		for _, ref := range note.Refs {
			if nums, err := doc.ParseRangesWithin([]string{ref}, stream); err == nil && len(nums) == 1 {
				cited[nums[0]] = true
			}
		}
	}
	var out []doc.Refusal
	for i, order := range m.standing.Registry.Open {
		if order.State != doc.StateAwaiting {
			continue
		}
		stillAwaiting := false
		for _, listed := range d.Open {
			if listed.Demand == order.Demand && listed.Question == order.Question && listed.State == doc.StateAwaiting {
				stillAwaiting = true
				break
			}
		}
		if stillAwaiting {
			continue
		}
		closed := retired.Lines[orderLineNumber(order)]
		if !closed {
			// An unlisted order stands; only a replacement that leaves
			// the awaiting state is a close.
			relisted := false
			for _, listed := range d.Open {
				if listed.Demand == order.Demand && listed.Question == order.Question {
					relisted = true
					break
				}
			}
			if !relisted {
				continue
			}
		}
		ok := false
		for n := order.OpenedAtLines + 1; n <= len(m.universe); n++ {
			if occupied(n) && cited[n] {
				ok = true
				break
			}
		}
		if !ok {
			out = append(out, doc.Refusal{
				Where: fmt.Sprintf("open[%d]", i),
				Law:   doc.LawCandor,
				Fact:  "the order was closed without a later human line — bring the human's answer into demands.md, given its outcome, cited by a decision note",
			})
		}
	}
	return out
}

// containsID — whether the submission lists the behavior id.
func containsID(d doc.Document, id string) bool {
	for _, b := range d.Behaviors {
		if b.ID == id {
			return true
		}
	}
	for _, removed := range d.Remove {
		if removed == id {
			return true
		}
	}
	return false
}

// lawPrinted — whether the id is a law the contract prints: the one
// source is the canon's own list.
func lawPrinted(id string) bool {
	for _, law := range canon.LawIDs() {
		if law == id {
			return true
		}
	}
	return false
}

// demandLine — the D id's line of the run's numbered universe: its
// own text and whether it exists. The parse is bounded by the
// universe's own count — a wide spelling answers false, it never
// generates.
func (m *machine) demandLine(id string) (string, bool) {
	nums, err := doc.ParseRangesWithin([]string{id}, len(m.universe))
	if err != nil || len(nums) != 1 || nums[0] < 1 || nums[0] > len(m.universe) {
		return "", false
	}
	return m.universe[nums[0]-1], true
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
					Fact:  fmt.Sprintf("the behavior covers %q — no requirement %q stands after this submission — keep it standing or drop the cover, this same submission", cover, cover),
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
		CorpusVersion: m.standing.CorpusVersion + 1,
		Build:         d.Build,
		Intents:       d.Intents,
		Notes:         d.Notes,
		Requirements:  d.Requirements,
		Markers:       d.Markers,
		Exclusions:    d.Exclusions,
		Open:          d.Open,
		Removed:       d.Remove,
		DocAddr:       docAddr,
	}
	// The document's own order, computed pure: removes retire first,
	// then the listed behaviors merge onto what remains standing — a
	// standing id merges (the head replaces, the rows merge by run,
	// the journaled object is the merged result), a remove-and-re-add
	// starts fresh. The fold replays the same order; the standing
	// itself moves only after the journal carries the submission.
	removedBehaviors := map[string]bool{}
	for _, key := range d.Remove {
		if doc.ParseRemoveKey(key) == doc.RemoveBehavior {
			removedBehaviors[key] = true
		}
	}
	merged := make([]doc.Behavior, len(d.Behaviors))
	for i, b := range d.Behaviors {
		if standing, seen := m.standing.Behaviors[b.ID]; seen && !removedBehaviors[b.ID] {
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
	if err := m.append(journal.Record{Kind: journal.KindSubmission, Submission: &sub}); err != nil {
		return err
	}
	// The same law the fold replays: the record index of the last
	// accepted submission — the live standing and its replay can
	// never diverge.
	m.standing.LastSubmissionAt = len(m.records) - 1
	// The words' own fact, journaled beside the submission by the
	// machine itself: the stream's unjournaled tail becomes the new
	// link of the chain, or the machine's own record answers that no
	// words came — the submitting agent never transcribes the words.
	if tail := m.demandTail(); len(tail) > 0 {
		link := journal.DemandAppend{Lines: len(tail), Chain: chainOver(m.standing.ChainHead, tail)}
		if err := m.append(journal.Record{Kind: journal.KindDemandAppend, DemandAppend: &link}); err != nil {
			return err
		}
		m.standing.ChainAppends = append(m.standing.ChainAppends, link)
		m.standing.ChainHead = link.Chain
		m.standing.PastLines += link.Lines
		m.standing.LastNoNewDemands = 0
	} else {
		signed := journal.NoNewDemands{Lines: len(m.universe), Run: m.standing.RunNumber + 1}
		if err := m.append(journal.Record{Kind: journal.KindNoNewDemands, NoNewDemands: &signed}); err != nil {
			return err
		}
		m.standing.LastNoNewDemands = signed.Run
	}
	m.standing.CorpusVersion = sub.CorpusVersion
	journal.ApplyCarriers(m.standing, &sub)
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
		// The same index the fold builds: each journaled row's
		// digest answers its canonical bytes — the live standing
		// and its replay carry one truth.
		for _, row := range b.Rows {
			rowBytes := doc.MarshalRow(row)
			m.standing.RowBytes[domain.Digest(rowBytes)] = string(rowBytes)
		}
	}
	journal.NormalizeOrder(m.standing)
	m.standing.Accepted++
	m.standing.Notes = append(m.standing.Notes, sub.Notes...)
	return nil
}

// append — one record to the live epoch and to the run's once-read
// slice: the single read of the journal stays the run's whole
// truth.
func (m *machine) append(rec journal.Record) error {
	if err := m.stream.Append(rec); err != nil {
		return err
	}
	m.records = append(m.records, rec)
	return nil
}

// runExecutions — the declared executions, run by the machine
// itself under the row law: each execute intent's command, its exit
// and output riding the journal — stdout whole up to four kilobytes,
// the object store's address beyond. A run that executes lands its
// run record, so the fact's run always names a recorded run.
func (m *machine) runExecutions() error {
	for _, in := range m.submitted.Intents {
		if in.Outcome != doc.OutcomeExecute {
			continue
		}
		start := time.Now()
		res := RunCommand(Argv(in.Run), m.root, RunEnv(m.pins), m.rowTimeout)
		m.spent += time.Since(start)
		fact := journal.ExecutionFact{
			At:           m.standing.RunNumber + 1,
			Lines:        in.Lines,
			Run:          in.Run,
			StdoutDigest: domain.Digest([]byte(res.Stdout)),
		}
		switch {
		case res.Err != nil:
			fact.Exit = -1
		default:
			fact.Exit = res.Exit
			if len(res.Stdout) <= 4096 {
				fact.Stdout = res.Stdout
			} else {
				addr, err := m.objects.Put([]byte(res.Stdout))
				if err != nil {
					return err
				}
				fact.StdoutAddr = addr
			}
		}
		if err := m.append(journal.Record{Kind: journal.KindExecution, Execution: &fact}); err != nil {
			return err
		}
		m.execFacts = append(m.execFacts, fact)
		m.standing.LastExecution = &fact
	}
	return nil
}

// claimRange — one claimed range with its own digest's short token.
type claimRange struct {
	Claim    string
	Digest12 string
}

// claimsFacts — the behavior's dependency fingerprint whole: the
// claims digest and each claim's own range digest, each claimed
// file read once per computation.
type claimsFacts struct {
	digest string
	ranges []claimRange
}

// claimsFactsOf — the behavior's dependency fingerprint: the sha256
// of the sorted "range=digest" list of its claims, with each
// claim's own range digest beside it. A claimed file that no longer
// reads spells absent — the hash sees the present — and every
// claimed file is read once, however many ranges claim it. The
// proofs' own economy: the walk cache never answers here.
func (m *machine) claimsFactsOf(b doc.Behavior) claimsFacts {
	reads := map[string][]byte{}
	entries := make([]string, 0, len(b.Claims))
	ranges := make([]claimRange, 0, len(b.Claims))
	for _, claim := range b.Claims {
		rel, first, last, err := fprint.ParseLineRange(claim)
		digest := fprint.AbsentDigest
		if err == nil {
			data, seen := reads[rel]
			if !seen {
				var rerr error
				data, rerr = ports.FSBind.ReadFile(filepath.Join(m.root, filepath.FromSlash(rel)))
				if rerr != nil {
					data = nil
				}
				reads[rel] = data
			}
			if data != nil {
				digest = fprint.LineRangeDigest(data, first, last)
			}
		}
		entries = append(entries, claim+"="+digest)
		ranges = append(ranges, claimRange{Claim: claim, Digest12: fprint.Token12(digest)})
	}
	sort.Strings(entries)
	return claimsFacts{digest: domain.Digest([]byte(strings.Join(entries, "\n"))), ranges: ranges}
}

// claimsMap — every behavior's claims facts over the corpus, one
// computation.
func (m *machine) claimsMap(corpus doc.Document) map[string]claimsFacts {
	out := make(map[string]claimsFacts, len(corpus.Behaviors))
	for _, b := range corpus.Behaviors {
		out[b.ID] = m.claimsFactsOf(b)
	}
	return out
}

// executeStale — the fingerprints' verdict per row, the execution of
// the stale ones — one at a time inside the run's own budget: what
// the budget cannot fit waits the next run, named by a warning.
func (m *machine) executeStale(corpus doc.Document, claims map[string]claimsFacts, envHash string) ([]journal.RowFact, []rowReport) {
	type pending struct {
		key      string
		row      doc.Row
		rowBytes []byte
		proof    string
	}
	var queue []pending
	for _, b := range corpus.Behaviors {
		behFacts := claims[b.ID]
		for i, row := range b.Rows {
			key := b.ID + "/" + strconv.Itoa(i)
			rowBytes := doc.MarshalRow(row)
			proof := fprint.Proof(rowBytes, behFacts.digest, envHash)
			// A behavior with rows and no claims stands on nothing it
			// names: its rows are honestly always stale.
			fact, ok := m.standing.Rows[key]
			stale := !ok || fact.Proof != proof || len(b.Claims) == 0
			if stale {
				queue = append(queue, pending{key: key, row: row, rowBytes: rowBytes, proof: proof})
			}
		}
	}
	var facts []journal.RowFact
	var reports []rowReport
	for _, p := range queue {
		if m.spent+m.rowTimeout > m.runBudget {
			m.warnings = append(m.warnings, fmt.Sprintf("run budget spent — %d rows wait the next run", len(queue)-len(facts)))
			return facts, reports
		}
		row := p.row
		key := p.key
		start := time.Now()
		res := RunCommand(Argv(row.Run), m.root, RunEnv(m.pins), m.rowTimeout)
		m.spent += time.Since(start)
		rp := rowReport{Key: key, Run: row.Run}
		var broken []string
		switch {
		case res.Err != nil:
			broken = []string{"launch: " + res.Err.Error()}
			rp.Exit = -1
		case res.TimedOut:
			broken = []string{fmt.Sprintf("timed out after %s — the deadline defaults to 300 s; PUNCHTAPE_ROW_TIMEOUT names whole seconds", m.rowTimeout)}
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
			Key: key, Row: domain.Digest(p.rowBytes), Status: status, Exit: rp.Exit,
			Facts: broken, Proof: p.proof,
		})
		rp.Facts = broken
		reports = append(reports, rp)
	}
	return facts, reports
}

// unpinnedBehaviors — the standing behaviors no green pinned row
// proves on the current fingerprints, in corpus order: green over an
// unpinned behavior is not coverage.
func (m *machine) unpinnedBehaviors(corpus doc.Document, claims map[string]claimsFacts, envHash string) []string {
	var out []string
	for _, b := range corpus.Behaviors {
		if m.greenPin(b, claims[b.ID].digest, envHash) {
			continue
		}
		out = append(out, b.ID)
	}
	return out
}

// greenPin — whether one pinned row of the behavior stands green on
// the current fingerprints.
func (m *machine) greenPin(b doc.Behavior, claimsDigest, envHash string) bool {
	for i, row := range b.Rows {
		if !doc.Pinned(row.Expect) {
			continue
		}
		fact, ok := m.standing.Rows[b.ID+"/"+strconv.Itoa(i)]
		if ok && fact.Status == journal.StatusGreen &&
			fact.Proof == fprint.Proof(doc.MarshalRow(row), claimsDigest, envHash) {
			return true
		}
	}
	return false
}

// driftNamedBySubmission — whether a submission landed after the
// drift's own witness run: the naming rides the record order, the
// witness run arrives computed — one truth a run.
func (m *machine) driftNamedBySubmission(since int) bool {
	witness, submission := -1, -1
	for i, rec := range m.records {
		switch {
		case rec.Kind == journal.KindSubmission && rec.Submission != nil:
			submission = i
		case rec.Kind == journal.KindRun && rec.Run != nil && rec.Run.N == since:
			witness = i
		}
	}
	return submission > witness
}

// proven — whether every row of the corpus stands green on the
// current fingerprints. The caller has already refused the empty
// corpus.
func (m *machine) proven(corpus doc.Document, claims map[string]claimsFacts, envHash string) bool {
	for _, b := range corpus.Behaviors {
		for i, row := range b.Rows {
			key := b.ID + "/" + strconv.Itoa(i)
			fact, ok := m.standing.Rows[key]
			if !ok || fact.Status != journal.StatusGreen {
				return false
			}
			if fact.Proof != fprint.Proof(doc.MarshalRow(row), claims[b.ID].digest, envHash) {
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

// rowBytesOf — the row bytes a digest names, answered by the one
// index the live accept and the fold both build: every digest the
// journal ever named rides it.
func (m *machine) rowBytesOf(rowDigest string) []byte {
	return []byte(m.standing.RowBytes[rowDigest])
}

// closeVerdict — the act's own close over the standing corpus: the
// verdict stands only over a corpus every behavior of which proves
// through a green pinned row on the current fingerprints, a stream
// accounted whole, every requirement covered, the tree equal to its
// claims plus loud exclusions, and no order awaiting the human —
// anything short names its holes in the report, one line per class.
// The proof arithmetic rides the before-map (the digests the proofs
// were measured against); the drift question rides the after-map
// (the tree as the run left it) — two truths, each named, never
// recomputed.
// verdictBlocker — one reason the completion act cannot record the
// verdict: the note the report holds it by and the imperative the
// terminal speaks, born together at the single point the reason is
// measured — a reason cannot stand voiceless or speak a stale fact.
type verdictBlocker struct {
	note        string
	instruction string
}

// verdictBlockers — every reason the verdict cannot record now, in
// the one priority order every face consumes: the report lists the
// notes whole, the terminal and state's next line speak the first
// instruction. One computation, one truth. The two claims maps are
// two truths on purpose: proofs match the fingerprints they were
// measured against (the map from before the execution), while the
// drift must see the tree as the run left it (the map from after) —
// a row may lawfully edit its own claimed file, and the maps then
// differ.
func (m *machine) verdictBlockers(corpus doc.Document, claims, claimsAfter map[string]claimsFacts) ([]verdictBlocker, error) {
	envHash := EnvHash(m.pins)
	var out []verdictBlocker
	// The build speaks first: a verdict over a red build is a seal on
	// a broken foundation — the reference names it the first blocker.
	if m.buildFailed {
		out = append(out, verdictBlocker{
			note:        "verdict: not yet — the build is red",
			instruction: m.buildRedInstruction(),
		})
	}
	if v := m.standing.TreeLockViolation; v != nil {
		if facts, werr := m.tree.walk(); werr == nil && facts.digest != v.Before {
			out = append(out, verdictBlocker{
				note:        "verdict: not yet — the tree changed under answer/execute/describe intents (law:demand)",
				instruction: "the tree moved under answer/execute/describe intents (law:demand) — fix the bytes back or submit the change as work to " + m.slot(zone.InboxSlot),
			})
		}
	}
	if v := m.standing.Verdict; v != nil {
		since := driftSince(v, claimsAfter, envHash, m.standing.RunNumber, m.records, m.rowBytesOf)
		if since > 0 && !m.driftNamedBySubmission(since) {
			out = append(out, verdictBlocker{
				note: fmt.Sprintf("verdict: not yet — drift unresolved: a submission must come after run %d", since),
				instruction: fmt.Sprintf("verdict drifted since run %d — re-earn it: submit after that run, then write done: true to %s and run",
					since, m.slot(zone.InboxSlot)),
			})
		}
	}
	if unpinned := m.unpinnedBehaviors(corpus, claims, envHash); len(unpinned) > 0 {
		names := strings.Join(unpinned, ", ")
		out = append(out, verdictBlocker{
			note:        fmt.Sprintf("verdict: not yet — behaviors without a green pinned row: %s", names),
			instruction: fmt.Sprintf("behaviors without a green pinned row: %s — submit to %s a row whose expect pins its exact output", names, m.slot(zone.InboxSlot)),
		})
	}
	if !m.proven(corpus, claims, envHash) {
		// A failed proof is a red fact or a missing one — the counts
		// carry it; a green fact with a stale proof was already
		// re-run by this run's execution.
		red, stale := m.countRows(corpus)
		instruction := fmt.Sprintf("%d row(s) stale — run again to execute them", stale)
		if red > 0 {
			instruction = fmt.Sprintf("%d row(s) red — read %s, fix, run again", red, m.slot(zone.ReportFile))
		}
		out = append(out, verdictBlocker{
			note:        "verdict: not yet — a row is red or stale",
			instruction: instruction,
		})
	}
	if holes := m.unresolvedLines(); len(holes) > 0 {
		spelled := strings.Join(holes, ", ")
		out = append(out, verdictBlocker{
			note:        fmt.Sprintf("verdict: not yet — unresolved lines: %s", spelled),
			instruction: fmt.Sprintf("unresolved lines: %s — submit to %s the carrier each declared outcome asks (a requirement on the line, a not-a-requirement marker, or an answer — removing a carrier re-opens its line)", spelled, m.slot(zone.InboxSlot)),
		})
	}
	if orphans := m.orphanRequirements(corpus); len(orphans) > 0 {
		names := strings.Join(orphans, ", ")
		out = append(out, verdictBlocker{
			note:        fmt.Sprintf("verdict: not yet — the requirements %s stand uncovered by any behavior's covers", names),
			instruction: fmt.Sprintf("the requirements %s stand uncovered — submit to %s a behavior whose covers name them", names, m.slot(zone.InboxSlot)),
		})
	}
	uncovered, err := m.uncoveredLines(corpus)
	if err != nil {
		return nil, err
	}
	if uncovered != "" {
		out = append(out, verdictBlocker{
			note:        uncovered,
			instruction: fmt.Sprintf("%d uncovered lines — submit their claims or loud exclusions to %s", m.treeHoles(corpus), m.slot(zone.InboxSlot)),
		})
	}
	for _, order := range m.standing.Registry.Open {
		if order.State == doc.StateAwaiting {
			out = append(out, verdictBlocker{
				note:        fmt.Sprintf("verdict: not yet — awaiting order %s: %s", order.Demand, order.Question),
				instruction: fmt.Sprintf("awaiting order %s: %s — put the human's answer into %s to close it", order.Demand, order.Question, m.slot(zone.DemandsSlot)),
			})
		}
	}
	return out, nil
}

func (m *machine) closeVerdict(corpus doc.Document, claims, claimsAfter map[string]claimsFacts) ([]string, error) {
	if len(corpus.Behaviors) == 0 {
		return []string{"verdict: refused — the corpus is empty"}, nil
	}
	blockers, err := m.verdictBlockers(corpus, claims, claimsAfter)
	if err != nil {
		return nil, err
	}
	if len(blockers) > 0 {
		notes := make([]string, 0, len(blockers))
		for _, b := range blockers {
			notes = append(notes, b.note)
		}
		return notes, nil
	}
	v := journal.Verdict{
		N:              verdictNumber(m.standing) + 1,
		CorpusVersion:  m.standing.CorpusVersion,
		Signature:      m.signature(corpus, claims),
		RegistryDigest: m.registryDigest(),
		Run:            m.standing.RunNumber,
	}
	ids := make([]string, 0, len(corpus.Behaviors))
	for _, b := range corpus.Behaviors {
		ids = append(ids, b.ID)
	}
	sort.Strings(ids)
	for _, id := range ids {
		v.Scopes = append(v.Scopes, journal.VerdictScope{ID: id, Pointer: claims[id].digest})
	}
	if err := m.append(journal.Record{Kind: journal.KindVerdict, Verdict: &v}); err != nil {
		return nil, err
	}
	// The act that closed it is a fact of its own, journaled beside
	// the verdict: the render's arithmetic counts submissions,
	// refusals and completions each by its own name.
	if err := m.append(journal.Record{
		Kind:      journal.KindCompletion,
		Completed: &journal.Completion{Run: m.standing.RunNumber},
	}); err != nil {
		return nil, err
	}
	m.standing.Completions++
	m.standing.Verdict = &v
	return []string{fmt.Sprintf("verdict: #%d on v%d, run %d recorded — signature %s", v.N, v.CorpusVersion, v.Run, shortDigest(v.Signature))}, nil
}

// orphanRequirements — the standing requirements no standing
// behavior that carries claims names in its covers, in registry
// order. A cover of a claims-less behavior is a cover of nothing:
// coverage lives in claimed lines.
func (m *machine) orphanRequirements(corpus doc.Document) []string {
	covers := map[string]bool{}
	for _, b := range corpus.Behaviors {
		if len(b.Claims) == 0 {
			continue
		}
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

// coverage — the run's one fact of the tree against the standing
// claims and the loud exclusions: computed once from one walk,
// shared by the verdict's totality gate and the state's faces.
func (m *machine) coverage(corpus doc.Document) (fprint.Coverage, error) {
	if m.covered != nil {
		return *m.covered, nil
	}
	facts, err := m.tree.walk()
	if err != nil {
		return fprint.Coverage{}, err
	}
	files := facts.files
	var claims, excl []string
	for _, b := range corpus.Behaviors {
		claims = append(claims, b.Claims...)
	}
	for _, ex := range m.standing.Registry.Exclusions {
		excl = append(excl, ex.Range)
	}
	cov := fprint.LineCoverage(claims, excl, files)
	m.covered = &cov
	return cov, nil
}

// uncoveredLines — the totality gate's own answer over the run's
// coverage fact: holes hold the verdict, named per file with counts
// and maximal ranges, the first ten files shown and the rest
// counted. Files outside the coverage universe never hold the
// verdict.
func (m *machine) uncoveredLines(corpus doc.Document) (string, error) {
	cov, err := m.coverage(corpus)
	if err != nil {
		return "", err
	}
	paths := make([]string, 0, len(cov.Holes))
	for path := range cov.Holes {
		paths = append(paths, path)
	}
	if len(paths) == 0 {
		return "", nil
	}
	sort.Strings(paths)
	parts := make([]string, 0, len(paths))
	for _, path := range paths {
		parts = append(parts, spellHoles(path, cov.Holes[path], false))
	}
	shown := parts
	if len(parts) > 10 {
		shown = parts[:10]
	}
	line := "verdict: not yet — uncovered lines: " + strings.Join(shown, ", ")
	if len(parts) > 10 {
		line += fmt.Sprintf(", and %d more files", len(parts)-10)
	}
	return line, nil
}

// registryDigest — the demands' account as one digest: the canonical
// bytes of the standing requirements and markers together. Any
// change to either moves the digest the verdict's record carries.
func (m *machine) registryDigest() string {
	d := doc.Document{}
	for _, id := range m.standing.Registry.Order {
		d.Requirements = append(d.Requirements, m.standing.Registry.Requirements[id])
	}
	d.Markers = m.standing.Registry.Markers
	d.Exclusions = m.standing.Registry.Exclusions
	d.Open = m.standing.Registry.Open
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
// and every behavior's claims fingerprint.
func (m *machine) signature(corpus doc.Document, claims map[string]claimsFacts) string {
	var b bytes.Buffer
	b.Write(doc.Marshal(corpus))
	for _, beh := range corpus.Behaviors {
		b.WriteByte(0)
		b.WriteString(beh.ID)
		b.WriteByte(0)
		b.WriteString(claims[beh.ID].digest)
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
	// The walk cache's own record: one write a run, only when the
	// walk moved it.
	if m.tree.walked {
		if err := m.writeIfChanged(filepath.Join(m.zoneDir, zone.CacheTreeDir), m.tree.cacheBytes()); err != nil {
			fmt.Fprintf(terminal, "punchtape: tree cache: %v\n", err)
			return 1
		}
	}
	if err := m.ensureAgentsBlock(); err != nil {
		fmt.Fprintf(terminal, "punchtape: %v\n", err)
		return 1
	}
	// The teaching faces stay current with the canon: identical bytes
	// touch nothing, a changed canon replaces its own page — a zone
	// never teaches a law its machine has left.
	if err := m.writeIfChanged(filepath.Join(m.zoneDir, zone.ContractFile), []byte(canon.Load().Contract)); err != nil {
		fmt.Fprintf(terminal, "punchtape: contract view: %v\n", err)
		return 1
	}
	if err := m.writeIfChanged(filepath.Join(m.zoneDir, zone.TemplateFile), []byte(canon.Load().Template)); err != nil {
		fmt.Fprintf(terminal, "punchtape: template view: %v\n", err)
		return 1
	}
	if err := m.writeIfChanged(filepath.Join(m.zoneDir, zone.ReferenceFile), []byte(canon.Load().Reference)); err != nil {
		fmt.Fprintf(terminal, "punchtape: reference view: %v\n", err)
		return 1
	}
	if err := ports.FSBind.RemoveAll(filepath.Join(m.zoneDir, zone.TmpDir)); err != nil {
		fmt.Fprintf(terminal, "punchtape: tmp clear: %v\n", err)
		return 1
	}
	fmt.Fprintf(terminal, "punchtape: run %d — %s\n", m.standing.RunNumber, instruction)
	return 0
}

// buildRedInstruction — the red build's own imperative, the one
// voice of the terminal line and the verdict's first blocker.
func (m *machine) buildRedInstruction() string {
	return "build red — read " + m.slot(zone.ReportFile) + ", fix the build, run again"
}

// nextInstruction — the one imperative that answers "what do I do
// now" from the standing alone: every state the machine can stand in
// has exactly one instruction, and it names the file to touch. The
// first standing verdict blocker speaks — the very table the report
// holds the verdict by — and with none, the corpus's own state does.
// The terminal's pointer line and state.md's next: line are this one
// computed text — never two truths.
func (m *machine) nextInstruction(corpus doc.Document, claims, claimsAfter map[string]claimsFacts) string {
	if m.buildFailed {
		return m.buildRedInstruction()
	}
	if len(corpus.Behaviors) == 0 {
		if len(corpus.Requirements) == 0 && len(corpus.Markers) == 0 && len(corpus.Exclusions) == 0 &&
			len(corpus.Open) == 0 && corpus.BuildCmd() == "" {
			return "corpus empty — submit the first document as " + m.slot(zone.InboxSlot) + " (copy " + m.slot(zone.TemplateFile) + ")"
		}
		return "no behaviors stand — submit to " + m.slot(zone.InboxSlot) + " a behavior with a pinned row"
	}
	if blockers, err := m.verdictBlockers(corpus, claims, claimsAfter); err == nil && len(blockers) > 0 {
		return blockers[0].instruction
	}
	if m.standing.Verdict != nil && m.standing.Verdict.CorpusVersion == m.standing.CorpusVersion {
		return fmt.Sprintf("verdict #%d standing — show %s to the human; new work goes to %s as a document",
			m.standing.Verdict.N, m.slot(zone.StateFile), m.slot(zone.InboxSlot))
	}
	return "all green — write done: true to " + m.slot(zone.InboxSlot) + " and run to record the verdict"
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
