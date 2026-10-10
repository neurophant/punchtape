// Package cli — argument parsing and output. The whole surface visible
// to the agent is the two verbs of the machine contract. The human
// writes no commands: they talk to their own agent, the agent calls
// the verbs. A bare launch is not part of that surface: it answers
// one version line — a terminal affordance for the person, never
// mentioned in the agent-facing texts.
package cli

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/engine"
)

// Version — the machine's delivery version (semver). Stamped by the
// build (-ldflags -X); without stamping — the current release.
var Version = "0.3.0"

// flags — the parsed surface options of a verb call: the reply mode
// ("" full | quiet | delta), the batch (chain the run of the
// just-accepted spec into the same door call), the report's
// --since verdict and the dry validator's --check document.
// Unknown --flags are an honest refusal, never a file name.
type flags struct {
	args     []string
	mode     string
	runAfter bool
	since    string
	check    string
	full     bool
}

func parseFlags(args []string) (flags, error) {
	f := flags{}
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch a {
		case "--quiet":
			f.mode = engine.SlotModeQuiet
		case "--delta":
			f.mode = engine.SlotModeDelta
		case "--run-after":
			f.runAfter = true
		case "--since":
			if i+1 >= len(args) {
				return f, errors.New(canondata.T("cli.error.since-no-value"))
			}
			i++
			f.since = args[i]
		case "--check":
			if i+1 >= len(args) {
				return f, errors.New(canondata.T("cli.error.check-no-value"))
			}
			i++
			f.check = args[i]
		case "--full":
			f.full = true
		default:
			if strings.HasPrefix(a, "--") {
				return f, errors.New(canondata.T("cli.error.unknown-flag", canondata.M{
					"flag": a,
				}))
			}
			f.args = append(f.args, a)
		}
	}
	return f, nil
}

// Run executes the verb and returns the process code: 0 — success,
// 1 — submission rejected, 2 — failure.
func Run(args []string, stdout io.Writer) int {
	if len(args) == 0 {
		// A bare launch prints the one version line and nothing else:
		// a deterministic function — no directory reads, no side
		// effects, the same bytes anywhere. Teaching the loop is the
		// first next's job (the contract injection), not the terminal's.
		fmt.Fprintln(stdout, "punchtape "+Version)
		return 0
	}
	workdir, err := getwd()
	if err != nil {
		fmt.Fprint(stdout, canondata.T("cli.fail.workdir", canondata.M{"err": err.Error()})+"\n")
		return 2
	}
	f, err := parseFlags(args[1:])
	if err != nil {
		fmt.Fprint(stdout, err.Error()+"\n")
		return 2
	}

	switch args[0] {
	case "next":
		if f.since != "" {
			fmt.Fprint(stdout, canondata.T("cli.error.since-report-only")+"\n")
			return 2
		}
		if f.check != "" {
			fmt.Fprint(stdout, canondata.T("cli.error.check-spec-only")+"\n")
			return 2
		}
		return runNext(workdir, stdout, strings.Join(f.args, " "), f)
	case "submit":
		if f.since != "" {
			fmt.Fprint(stdout, canondata.T("cli.error.since-report-only")+"\n")
			return 2
		}
		if f.check != "" {
			fmt.Fprint(stdout, canondata.T("cli.error.check-spec-only")+"\n")
			return 2
		}
		if len(f.args) < 1 {
			fmt.Fprint(stdout, canondata.T("cli.error.submit-no-file")+"\n")
			return 1
		}
		return runSubmit(workdir, f.args[0], stdout, f)
	case "report":
		if f.check != "" {
			fmt.Fprint(stdout, canondata.T("cli.error.check-spec-only")+"\n")
			return 2
		}
		if f.mode != "" || f.runAfter {
			fmt.Fprint(stdout, canondata.T("cli.error.report-surface")+"\n")
			return 2
		}
		return runReport(workdir, stdout, f.since, f.full)
	case "spec":
		if f.mode != "" || f.runAfter || f.since != "" || f.check == "" || len(f.args) > 0 {
			fmt.Fprint(stdout, canondata.T("cli.error.spec-surface")+"\n")
			return 2
		}
		return runSpecCheck(workdir, f.check, stdout)
	default:
		fmt.Fprint(stdout, canondata.T("cli.error.no-verb", canondata.M{"verb": fmt.Sprintf("%q", args[0])})+"\n")
		return 2
	}
}

// runSpecCheck — the spec verb's one surface: the dry validator
// (spec --check <file>). The door's form rules over the draft, with
// one reason line per finding; a pass is the door's own promise —
// the same predicates, the first try.
func runSpecCheck(workdir, source string, stdout io.Writer) int {
	findings, err := engine.CheckDocument(workdir, source)
	if err != nil {
		fmt.Fprint(stdout, err.Error()+"\n")
		return 2
	}
	if len(findings) == 0 {
		fmt.Fprint(stdout, canondata.T("cli.check.pass")+"\n")
		return 0
	}
	for _, line := range findings {
		entity, reason := line, ""
		if i := strings.Index(line, ": "); i >= 0 {
			entity, reason = line[:i], line[i+2:]
		}
		fmt.Fprint(stdout, canondata.T("submit.reject.line", canondata.M{
			"entity": entity, "reason": reason,
		})+"\n")
	}
	return 1
}

// runReport — the report verb: the stored-runs table of the observed
// behavior (nothing re-runs); --since vN aims it at the delta after
// a frozen verdict.
func runReport(workdir string, stdout io.Writer, since string, full bool) int {
	eng, exists, err := openEngineRetry(workdir)
	if err != nil {
		fmt.Fprint(stdout, canondata.T("cli.fail", canondata.M{"err": err.Error()})+"\n")
		return 2
	}
	if !exists {
		fmt.Fprint(stdout, canondata.T("cli.error.no-instance")+"\n")
		return 1
	}
	reply, code := eng.Report(since, full)
	fmt.Fprintln(stdout, reply)
	return code
}

// openEngineRetry opens the instance with short retries: a
// submission holds the write monopoly for the whole transaction
// (including the rows of a run); a few attempts are enough for
// reading in this window; persistent occupancy — an honest refusal.
func openEngineRetry(workdir string) (*engine.Engine, bool, error) {
	var eng *engine.Engine
	var exists bool
	var err error
	for try := 0; try < 3; try++ {
		eng, exists, err = engine.Open(workdir)
		if err == nil || !errors.Is(err, engine.ErrWriterBusy) {
			return eng, exists, err
		}
		sleep(250 * time.Millisecond)
	}
	return eng, exists, err
}

func runNext(workdir string, stdout io.Writer, brief string, f flags) int {
	// next in an empty directory creates the instance: the loop's
	// first point — but ONLY with the task phrase: an instance is
	// born bound to its task, never free-floating (a
	// phraseless verdict belongs to no task). A bare next on an
	// EXISTING instance stays the plain slot ask.
	eng, exists, err := openEngineRetry(workdir)
	if err != nil {
		fmt.Fprint(stdout, canondata.T("cli.fail", canondata.M{"err": err.Error()})+"\n")
		return 2
	}
	phrase := strings.TrimSpace(brief)
	if !exists && phrase == "" {
		fmt.Fprint(stdout, canondata.T("cli.error.birth-no-phrase")+"\n")
		return 1
	}
	if !exists {
		eng, err = engine.Init(workdir)
		if err != nil {
			fmt.Fprint(stdout, canondata.T("cli.fail", canondata.M{"err": err.Error()})+"\n")
			return 2
		}
	}
	// Contract auto-injection (generating AGENTS.md is not enough —
	// the executor must receive it into its context automatically).
	// The first next (instance creation) prints the full reference
	// before the slot: no "read the file" step is required of the
	// executor.
	if !exists {
		fmt.Fprint(stdout, canondata.T("cli.next.contract-inject")+"\n")
		fmt.Fprint(stdout, engine.AgentsRefText())
		fmt.Fprintln(stdout)
	}
	if phrase == "" {
		return plainNext(eng, stdout, f)
	}
	// The brief with initialization: the text arrives with the first
	// next and is fixed in the same transaction — the brief
	// hand-off round trip disappeared; the reply is the first slot.
	reply, code := eng.Begin(phrase)
	fmt.Fprintln(stdout, reply)
	return code
}

// plainNext — a next call without a phrase: the current work's
// slot (in the call's surface mode).
func plainNext(eng *engine.Engine, stdout io.Writer, f flags) int {
	text, err := eng.NextWith(f.mode)
	if err != nil {
		fmt.Fprint(stdout, canondata.T("cli.fail", canondata.M{"err": err.Error()})+"\n")
		return 2
	}
	fmt.Fprintln(stdout, text)
	return 0
}

func runSubmit(workdir, source string, stdout io.Writer, f flags) int {
	eng, exists, err := openEngineRetry(workdir)
	if err != nil {
		fmt.Fprint(stdout, canondata.T("cli.fail", canondata.M{"err": err.Error()})+"\n")
		return 2
	}
	if !exists {
		fmt.Fprint(stdout, canondata.T("cli.error.no-instance")+"\n")
		return 1
	}
	reply, code := eng.SubmitSetOpts(source, engine.SubmitOpts{Mode: f.mode, RunAfter: f.runAfter})
	fmt.Fprintln(stdout, reply)
	return code
}
