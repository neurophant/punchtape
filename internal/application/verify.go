package application

// The verify call: the human's re-proof of a zone the agent may have
// written. The machine checks that its journal agrees with itself,
// not who wrote it — a consistent rewrite passes the ordinary call.
// This call trusts none of it: the build and every row of the
// standing corpus run again, the journal's green facts do not count,
// nothing of the zone is written — the scratch clears, as after a
// run — and the answer is the exit code.

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/neurophant/punchtape/internal/domain/journal"
	"github.com/neurophant/punchtape/internal/domain/zone"
	"github.com/neurophant/punchtape/internal/ports"
)

// Verify — the human's check of a zone: under the same lock as the
// run, the build and every row of the standing corpus run again, the
// journal's green facts do not count, and nothing is written — not a
// record, not a view; the inbox is never read; the scratch clears as
// after a run. The exit code: 0 proven, 1 not proven, a scratch that
// would not clear, or the call did not happen.
func Verify(cwd string, terminal io.Writer) int {
	root, found := findZone(cwd)
	if !found {
		fmt.Fprintln(terminal, "punchtape: no zone here — nothing to verify")
		return 1
	}
	zoneDir := filepath.Join(root, zone.Dirname)
	if _, err := ports.FSBind.Stat(filepath.Join(zoneDir, zone.JournalDir)); err != nil {
		if !os.IsNotExist(err) {
			fmt.Fprintf(terminal, "punchtape: zone damaged: %v\n", err)
			return 1
		}
		// No instance, no birth and no lock: answered ahead of the
		// lock, whose file is the one thing the check may leave.
		fmt.Fprintln(terminal, "punchtape: no instance here yet — nothing to verify")
		return 1
	}
	unlock, ok := lockZone(zoneDir, terminal)
	if !ok {
		return 1
	}
	defer unlock()
	m, err := openBorn(cwd, root, zoneDir, true)
	if err != nil {
		fmt.Fprintf(terminal, "punchtape: %v\n", err)
		return 1
	}
	m.verify = true
	code := m.reprove(terminal)
	// The scratch clears under the same lock, as after a run: the rows
	// that just ran may have left theirs.
	if err := m.clearTmp(); err != nil {
		fmt.Fprintf(terminal, "punchtape: tmp clear: %v\n", err)
		return 1
	}
	return code
}

// reprove — the re-proof itself: the build, then every row of the
// standing corpus in corpus order under the run's budget and the row
// deadline. The executed facts replace the journal's in memory alone,
// and the verdict's own gates — the drift included — judge this run.
// The terminal carries the whole answer: a red row a line, a blocker
// a line, the budget's warning, and the last line's verdict.
func (m *machine) reprove(terminal io.Writer) int {
	// The chain, as every run: a moved past line is zone damage, and
	// the check never starts.
	if err := m.checkChain(); err != nil {
		fmt.Fprintf(terminal, "punchtape: zone damaged: %v\n", err)
		return 1
	}
	corpus := m.corpusDoc()
	m.runBuild(corpus)
	claims := m.claimsMap(corpus)
	var ran []journal.RowFact
	if !m.buildFailed {
		// Every row is stale under the re-proof. The facts of this run
		// take the journal's place in memory alone, so the gates
		// below judge what ran now, never what the journal claims.
		ran, _ = m.executeStale(corpus, claims, EnvHash(m.pins))
		for _, fact := range ran {
			m.standing.Rows[fact.Key] = fact
		}
	}
	blockers, err := m.verdictBlockers(corpus, claims, m.claimsMap(corpus))
	if err != nil {
		fmt.Fprintf(terminal, "punchtape: %v\n", err)
		return 1
	}
	holds := len(corpus.Behaviors) > 0 && len(blockers) == 0 && len(m.warnings) == 0
	if len(corpus.Behaviors) == 0 {
		fmt.Fprintln(terminal, "verdict: not yet — the corpus is empty")
	}
	green := 0
	for _, fact := range ran {
		if fact.Status == journal.StatusGreen {
			green++
			continue
		}
		holds = false
		fmt.Fprintf(terminal, "red %s — %s\n", fact.Key, strings.Join(fact.Facts, "; "))
	}
	for _, b := range blockers {
		fmt.Fprintln(terminal, b.note)
	}
	for _, warning := range m.warnings {
		fmt.Fprintln(terminal, warning)
	}
	if !holds {
		fmt.Fprintf(terminal, "punchtape: verify — not proven — %s\n", m.judgments(corpus))
		return 1
	}
	fmt.Fprintf(terminal, "punchtape: verify — proven — %d rows run green — %s\n", green, m.judgments(corpus))
	return 0
}
