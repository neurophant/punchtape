// Package cli — argument parsing and output. The whole surface visible
// to the agent is three verbs: next, submit, why. The human writes no
// commands: they talk to their own agent, the agent calls the verbs.
package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/neurophant/punchtape/internal/brief"
	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/delta"
	"github.com/neurophant/punchtape/internal/driver"
	"github.com/neurophant/punchtape/internal/engine"
	"github.com/neurophant/punchtape/internal/intent"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// usageKey — the agent contract text key in canon data: a bare
// punchtape launch teaches the loop without external instructions.
// The machine teaches how to work with it by construction — the
// slot carries the action and a ready example; this header closes
// the cycle.
const usageKey = "cli.usage"

// Version — the machine's delivery version (semver). Stamped by the
// build (-ldflags -X); without stamping — a dev build from the
// working tree.
var Version = "dev"

// Run executes the verb and returns the process code: 0 — success,
// 1 — submission rejected, 2 — failure, 3 — a choice is required (a
// batch of questions to the human is open), 4 — escalation (an open
// conflict requires an explicit live choice; silence does not
// resolve it).
func Run(args []string, stdout io.Writer) int {
	workdir, err := os.Getwd()
	if err != nil {
		fmt.Fprint(stdout, canondata.T("cli.fail.workdir", canondata.M{"err": err.Error()}))
		return 2
	}
	if len(args) == 0 {
		// A bare launch — a pure function without branching on the
		// directory state: bit-for-bit identical output anywhere,
		// zero side effects. There is one entry into work: next.
		fmt.Fprintln(stdout, canondata.T(usageKey))
		return 2
	}

	switch args[0] {
	case "version":
		// A delivery service verb: which binary version.
		fmt.Fprintln(stdout, "punchtape "+Version)
		return 0
	case "bg-suite":
		// The background suite rehearsal service mode: launched by
		// the machine as a separate detached process; not an executor
		// verb.
		if engine.RunBackgroundSuite(workdir) != nil {
			return 1
		}
		return 0
	case "next":
		forceFull := false
		wishParts := []string{}
		for _, a := range args[1:] {
			if a == "--full" {
				forceFull = true
				continue
			}
			wishParts = append(wishParts, a)
		}
		return runNext(workdir, stdout, forceFull, strings.Join(wishParts, " "))
	case "submit":
		// submit --check <file|-> — a preflight check: the same
		// validation, without applying. submit --try <file|-> — a
		// dry probe of ONE table row: the same runner in a clean
		// directory, without entering the canon or changing the
		// stage.
		checkOnly := false
		tryRow := false
		rest := args[1:]
		if len(rest) > 0 && rest[0] == "--check" {
			checkOnly = true
			rest = rest[1:]
		}
		if len(rest) > 0 && rest[0] == "--try" {
			tryRow = true
			rest = rest[1:]
		}
		if len(rest) < 1 {
			fmt.Fprint(stdout, canondata.T("cli.error.submit-no-file"))
			return 1
		}
		return runSubmit(workdir, rest[0], stdout, checkOnly, tryRow)
	case "why":
		// The topic is assembled from all arguments: a conversational
		// entity-kind form ("why card REQ-1") — two words, not one.
		return runWhy(workdir, strings.Join(args[1:], " "), stdout)
	case "driver":
		// The driver is not an agent verb: an automatic mode, not
		// used by the human, launched by the run infrastructure. The
		// caller shape (--shape cli|ide|api) is the slot envelope;
		// the slot structure is one.
		args, shape, err := takeShape(args[1:])
		if err != nil {
			fmt.Fprint(stdout, canondata.T("cli.fail", canondata.M{"err": err.Error()}))
			return 2
		}
		if len(args) >= 2 && args[0] == "fanout" {
			routing, err := driver.LoadRouting(args[1])
			if err != nil {
				fmt.Fprint(stdout, canondata.T("cli.fail", canondata.M{"err": err.Error()}))
				return 2
			}
			limit := 30
			if len(args) > 2 {
				fmt.Sscanf(args[2], "%d", &limit)
			}
			return driver.Fanout(workdir, shape, routing, limit, stdout)
		}
		if len(args) >= 2 && args[0] == "recompile" {
			limit := 30
			if len(args) > 2 {
				fmt.Sscanf(args[2], "%d", &limit)
			}
			return driver.Recompile(workdir, shape, args[1], limit, stdout)
		}
		if len(args) < 2 || args[0] != "run" {
			fmt.Fprintln(stdout, canondata.T("cli.error.driver-usage"))
			return 2
		}
		limit := 30
		if len(args) > 2 {
			fmt.Sscanf(args[2], "%d", &limit)
		}
		return driver.Run(workdir, shape, args[1], limit, stdout)
	default:
		fmt.Fprint(stdout, canondata.T("cli.error.no-verb", canondata.M{"verb": fmt.Sprintf("%q", args[0])})+"\n\n"+canondata.T(usageKey)+"\n")
		return 2
	}
}

// takeShape extracts the caller shape flag from the service mode's
// arguments: `--shape <cli|ide|api>`. Shapes outside the catalog are
// an error.
func takeShape(args []string) ([]string, string, error) {
	shape := "cli"
	for i := 0; i < len(args); i++ {
		if args[i] != "--shape" {
			continue
		}
		if i+1 >= len(args) {
			return args, "", fmt.Errorf("%s", canondata.T("cli.error.shape-missing"))
		}
		if !brief.KnownShape(args[i+1]) {
			return args, "", fmt.Errorf("%s", canondata.T("cli.error.shape-unknown", canondata.M{"shape": fmt.Sprintf("%q", args[i+1])}))
		}
		shape = args[i+1]
		rest := append([]string{}, args[:i]...)
		return append(rest, args[i+2:]...), shape, nil
	}
	return args, shape, nil
}

// openEngineRetry opens the instance with short retries: a
// submission holds the write monopoly for the whole transaction
// (including gates); a few attempts are enough for reading in this
// window; persistent occupancy — an honest refusal.
func openEngineRetry(workdir string) (*engine.Engine, bool, error) {
	var eng *engine.Engine
	var exists bool
	var err error
	for try := 0; try < 3; try++ {
		eng, exists, err = engine.Open(workdir)
		if err == nil || !errors.Is(err, engine.ErrWriterBusy) {
			return eng, exists, err
		}
		time.Sleep(250 * time.Millisecond)
	}
	return eng, exists, err
}

// choiceCode — the stop-with-options rc contract on top of a
// successful call: an open batch of questions — a human choice
// point (3), an unresolved conflict — escalation (4). The code
// reports the choice point; the cycle does not block: submitting
// the next delta is legal, silence applies the recommended defaults
// (except a conflict — that one waits for an explicit answer).
func choiceCode(eng *engine.Engine, code int) int {
	if code != engine.Accepted {
		return code
	}
	choice, escalate := eng.ChoicePoint()
	switch {
	case escalate:
		return 4
	case choice:
		return 3
	}
	return code
}

func runNext(workdir string, stdout io.Writer, forceFull bool, wish string) int {
	// next in an empty directory creates the instance: the loop's
	// first point.
	eng, exists, err := openEngineRetry(workdir)
	if err != nil {
		fmt.Fprint(stdout, canondata.T("cli.fail", canondata.M{"err": err.Error()}))
		return 2
	}
	if !exists {
		eng, err = engine.Init(workdir)
		if err != nil {
			fmt.Fprint(stdout, canondata.T("cli.fail", canondata.M{"err": err.Error()}))
			return 2
		}
	}
	// Contract auto-injection (generating AGENTS.md is not enough —
	// the executor must receive it into its context automatically).
	// The first next (instance creation) and every --full (a fresh
	// session's entry) print the full reference before the slot: no
	// "read the file" step is required of the executor.
	if !exists || forceFull {
		fmt.Fprint(stdout, canondata.T("cli.next.contract-inject"))
		fmt.Fprint(stdout, engine.AgentsRefText())
		fmt.Fprintln(stdout)
	}
	// The everyday-phrase intent router: a data table decides
	// whether it is a loop control phrase or a product wish;
	// ambiguity — a menu with a recommended default (the machine
	// does not resolve the fork, nothing is applied, rc 3 — a
	// choice point).
	phrase := strings.TrimSpace(wish)
	if phrase == "" {
		return plainNext(eng, stdout, forceFull)
	}
	target, ambiguous, matched := intent.Route(phrase)
	if ambiguous {
		opts := make([]string, 0, len(matched))
		for _, t := range matched {
			opts = append(opts, canondata.T("cli.intent.option", canondata.M{"label": t.Label()}))
		}
		fmt.Fprintln(stdout, canondata.T("cli.intent.menu", canondata.M{
			"options": strings.Join(opts, "\n"),
		}))
		return 3
	}
	switch target.Verb {
	case "next":
		// a continuation everyday phrase — not a wish: the same
		// plain next.
		return plainNext(eng, stdout, forceFull)
	case "why":
		return runWhy(workdir, target.Topic, stdout)
	}
	// A wish with initialization: the text arrives with the first
	// next and is applied in the same transaction as the intent
	// delta — the brief hand-off round trip disappears; the reply
	// is the first slot (spec).
	raw, err := yamlio.Marshal(delta.Intent{
		Kind:          delta.KindIntent,
		SubmissionKey: fmt.Sprintf("intent-%03d", len(eng.Journal.All())+1),
		Text:          phrase,
	})
	if err != nil {
		fmt.Fprint(stdout, canondata.T("cli.fail", canondata.M{"err": err.Error()}))
		return 2
	}
	reply, code := eng.SubmitInit(raw)
	fmt.Fprintln(stdout, reply)
	return choiceCode(eng, code)
}

// plainNext — a next call without a phrase: the current work's
// slot.
func plainNext(eng *engine.Engine, stdout io.Writer, forceFull bool) int {
	text, err := eng.Next()
	if forceFull {
		text, err = eng.NextFull()
	}
	if err != nil {
		fmt.Fprint(stdout, canondata.T("cli.fail", canondata.M{"err": err.Error()}))
		return 2
	}
	fmt.Fprintln(stdout, text)
	return choiceCode(eng, 0)
}

func runSubmit(workdir, source string, stdout io.Writer, checkOnly, tryRow bool) int {
	var data []byte
	var err error
	if source == "-" {
		data, err = io.ReadAll(os.Stdin)
	} else {
		data, err = os.ReadFile(source)
	}
	if err != nil {
		fmt.Fprint(stdout, canondata.T("cli.fail.reading", canondata.M{"source": source, "err": err.Error()}))
		return 2
	}
	// Empty stdin — not an empty delta but one that never arrived:
	// an honest one-line reason (the pipe is not connected to the
	// process), not "EOF".
	if source == "-" && len(data) == 0 {
		fmt.Fprint(stdout, canondata.T("cli.error.stdin-empty"))
		return 1
	}
	eng, exists, err := openEngineRetry(workdir)
	if err != nil {
		fmt.Fprint(stdout, canondata.T("cli.fail", canondata.M{"err": err.Error()}))
		return 2
	}
	if !exists {
		fmt.Fprint(stdout, canondata.T("cli.error.no-instance"))
		return 1
	}
	var reply string
	var code int
	switch {
	case checkOnly:
		reply, code = eng.Check(data)
	case tryRow:
		reply, code = eng.Try(data)
	default:
		reply, code = eng.Submit(data)
		// A preflight check changes no state — the choice contract
		// speaks about the loop, not about the delta's validity.
		code = choiceCode(eng, code)
	}
	fmt.Fprintln(stdout, reply)
	return code
}

func runWhy(workdir, topic string, stdout io.Writer) int {
	eng, exists, err := openEngineRetry(workdir)
	if err != nil {
		fmt.Fprint(stdout, canondata.T("cli.fail", canondata.M{"err": err.Error()}))
		return 2
	}
	if !exists {
		fmt.Fprint(stdout, canondata.T("cli.error.no-instance"))
		return 1
	}
	text, err := eng.Why(topic)
	if err != nil {
		fmt.Fprintf(stdout, "%s\n", err)
		var nf *engine.WhyNotFound
		if errors.As(err, &nf) {
			return 1
		}
		return 2
	}
	fmt.Fprintln(stdout, text)
	return 0
}
