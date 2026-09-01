// Package driver — the automatic mode: a deterministic scheduler that
// spins the loop without a human. Slots go to an external executor (a
// caller command), deltas come back into the machine. No LLM logic
// lives here — only scheduling and passing.
package driver

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/neurophant/punchtape/internal/brief"
	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/engine"
)

// openWithRetry opens the instance, surviving brief writer-lock
// occupancy (a background suite rehearsal opens the same instance):
// an honest refusal stays a refusal; three attempts with a pause is
// transience, not silence.
func openWithRetry(workdir string) (*engine.Engine, bool, error) {
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

// Run spins the loop sequentially until the verdict, quarantine or the
// iteration limit. executor is a caller command in the given shape:
// brief in, delta YAML out; the adapter owns the envelope, never the
// slot.
func Run(workdir, shape, executor string, limit int, stdout io.Writer) int {
	lastReply := ""
	for i := 1; i <= limit; i++ {
		eng, exists, err := openWithRetry(workdir)
		if err != nil {
			fmt.Fprintf(stdout, "%s\n", canondata.T("driver.fail", canondata.M{"err": err.Error()}))
			return 2
		}
		if !exists {
			// the first turn in an empty directory creates the
			// instance — like the first next in manual mode
			eng, err = engine.Init(workdir)
			if err != nil {
				fmt.Fprintf(stdout, "%s\n", canondata.T("driver.fail", canondata.M{"err": err.Error()}))
				return 2
			}
		}
		slot, info, err := eng.NextInfo()
		if err != nil {
			fmt.Fprintf(stdout, "%s\n", canondata.T("driver.fail", canondata.M{"err": err.Error()}))
			return 2
		}

		if info.Stage == canon.StageDeliver {
			// Deliver with a machine spec proposal (backfill) is not
			// the finish: unclosed observations await submissions; the
			// cycle keeps spinning.
			if info.Kind != "spec" {
				verdict, err := eng.RenderVerdict()
				if err != nil {
					fmt.Fprintf(stdout, "%s\n", canondata.T("driver.fail", canondata.M{"err": err.Error()}))
					return 2
				}
				fmt.Fprintln(stdout, verdict.Text())
				return 0
			}
		}

		deltaText, tokens, err := CallExecutor(shape, workdir, executor, slot, lastReply, "")
		if err != nil {
			var timeout *ExecutorTimeoutError
			if errors.As(err, &timeout) {
				// An executor timeout is a lost iteration, not a mode
				// failure: the loop continues with the next one.
				fmt.Fprintf(stdout, "%s\n", timeout.Error())
				lastReply = ""
				continue
			}
			fmt.Fprintf(stdout, "%s\n", canondata.T("driver.fail.executor", canondata.M{
				"executor": fmt.Sprintf("%q", executor), "err": err.Error(),
			}))
			return 2
		}
		// The caller's counter: an executor that knows its tokens
		// reports them — a machine fact into the ledger.
		if err := eng.NoteExecutorCall(info.Kind, info.Facet, executor, shape, tokens); err != nil {
			fmt.Fprintf(stdout, "%s\n", canondata.T("driver.fail.ledger", canondata.M{"err": err.Error()}))
			return 2
		}
		reply, _ := eng.SubmitAuto([]byte(deltaText))
		// The submit reply carries the next slot; the caller's
		// envelope needs only the outcome line — the slot arrives by
		// itself with the next call; there is nothing to pay for
		// duplicating it in the envelope.
		lastReply = oneLine(reply)
		fmt.Fprintf(stdout, "[%d] %s\n", i, lastReply)
	}
	fmt.Fprintf(stdout, "%s\n", canondata.T("driver.not-ready",
		canondata.M{"limit": fmt.Sprintf("%d", limit)}))
	return 1
}

// Fanout spins the loop with parallel cards: on implement every active
// card goes to its own executor at the same time, deltas apply
// sequentially in card order — the single writer stays single, only
// the executors are parallel. Routing picks the executor command per
// slot kind and card facet.
func Fanout(workdir, shape string, routing *Routing, limit int, stdout io.Writer) int {
	lastReply := ""
	for i := 1; i <= limit; i++ {
		eng, exists, err := openWithRetry(workdir)
		if err != nil {
			fmt.Fprintf(stdout, "%s\n", canondata.T("driver.fail", canondata.M{"err": err.Error()}))
			return 2
		}
		if !exists {
			eng, err = engine.Init(workdir)
			if err != nil {
				fmt.Fprintf(stdout, "%s\n", canondata.T("driver.fail", canondata.M{"err": err.Error()}))
				return 2
			}
		}
		slot, info, err := eng.NextInfo()
		if err != nil {
			fmt.Fprintf(stdout, "%s\n", canondata.T("driver.fail", canondata.M{"err": err.Error()}))
			return 2
		}
		if info.Stage == canon.StageDeliver {
			// Deliver with a machine spec proposal (backfill) is not
			// the finish: unclosed observations await submissions; the
			// cycle keeps spinning.
			if info.Kind != "checks" {
				verdict, err := eng.RenderVerdict()
				if err != nil {
					fmt.Fprintf(stdout, "%s\n", canondata.T("driver.fail", canondata.M{"err": err.Error()}))
					return 2
				}
				fmt.Fprintln(stdout, verdict.Text())
				return 0
			}
		}

		if info.Stage == canon.StageImplement {
			cards, err := eng.Store.Cards()
			if err != nil {
				fmt.Fprintf(stdout, "%s\n", canondata.T("driver.fail", canondata.M{"err": err.Error()}))
				return 2
			}
			// The wave works active and red cards: the red list is
			// work too; green and quarantined ones wait for
			// convergence.
			var working []canon.Card
			for _, c := range cards {
				if c.Status == canon.CardActive || c.Status == canon.CardRed {
					working = append(working, c)
				}
			}
			if len(working) > 0 {
				if reply, code := fanoutWave(eng, shape, routing, working, i, stdout); !reply {
					return code
				}
				lastReply = ""
				continue
			}
		}

		// Sequential stages — one slot, one executor.
		route := routing.Route(info)
		deltaText, tokens, err := CallExecutor(shape, workdir, route, slot, lastReply, "")
		if err != nil {
			var timeout *ExecutorTimeoutError
			if errors.As(err, &timeout) {
				fmt.Fprintf(stdout, "%s\n", timeout.Error())
				lastReply = ""
				continue
			}
			fmt.Fprintf(stdout, "%s\n", canondata.T("driver.fail.executor", canondata.M{
				"executor": fmt.Sprintf("%q", route), "err": err.Error(),
			}))
			return 2
		}
		if err := eng.NoteExecutorCall(info.Kind, info.Facet, route, shape, tokens); err != nil {
			fmt.Fprintf(stdout, "%s\n", canondata.T("driver.fail.ledger", canondata.M{"err": err.Error()}))
			return 2
		}
		// The spec draft slot (empty spec) — only when routing
		// explicitly assigned a "draft" executor: a cheap executor
		// writes the first draft, the submission is marked by origin
		// and held for the main executor's review. Without an
		// explicit route the slot is written by the main executor
		// with a normal submission — there is no draft cycle.
		var reply string
		if info.Kind == "draft" && routing.HasKind("draft") {
			reply, _ = eng.SubmitDraft([]byte(deltaText))
		} else {
			reply, _ = eng.SubmitAuto([]byte(deltaText))
		}
		// The caller's envelope gets the outcome line; the slot nested
		// in the reply arrives by itself with the next call.
		lastReply = oneLine(reply)
		fmt.Fprintf(stdout, "[%d] %s\n", i, lastReply)
	}
	fmt.Fprintf(stdout, "%s\n", canondata.T("driver.not-ready",
		canondata.M{"limit": fmt.Sprintf("%d", limit)}))
	return 1
}

// fanoutWave — one wave of parallel work: the working cards' slots
// (active and red) go to executors simultaneously, deltas apply in
// card order. false — a failure, the return code is attached.
func fanoutWave(eng *engine.Engine, shape string, routing *Routing, active []canon.Card, wave int, stdout io.Writer) (bool, int) {
	type call struct {
		card  canon.Card
		delta string
		kind  string
		facet string
		route string
	}
	calls := make([]call, len(active))
	var wg sync.WaitGroup
	for idx, card := range active {
		wg.Add(1)
		go func(idx int, card canon.Card) {
			defer wg.Done()
			slot, info, err := eng.SlotForCard(card.ID)
			if err != nil {
				return
			}
			route := routing.Route(info)
			// The exchange key is the card: parallel ide-shape calls
			// do not share one file pair.
			deltaText, tokens, err := CallExecutor(shape, eng.Workdir, route, slot, "", card.ID)
			if err != nil {
				fmt.Fprintf(stdout, "%s\n", canondata.T("driver.fail.executor-card", canondata.M{
					"executor": fmt.Sprintf("%q", route), "card": card.ID, "err": err.Error(),
				}))
				return
			}
			// The ledger — from the main thread: records are
			// serialized.
			if err := eng.NoteExecutorCall(info.Kind, info.Facet, route, shape, tokens); err != nil {
				return
			}
			calls[idx] = call{card: card, delta: deltaText, kind: info.Kind, facet: info.Facet, route: route}
		}(idx, card)
	}
	wg.Wait()
	for _, c := range calls {
		if c.delta == "" {
			// a wave executor did not answer — the card stays active;
			// the next wave retries
			continue
		}
		reply, _ := eng.SubmitAuto([]byte(c.delta))
		fmt.Fprintf(stdout, "[%d] %s: %s\n", wave, c.card.ID, oneLine(reply))
	}
	return true, 0
}

// CallExecutor passes the slot to the executor in the caller's shape
// and collects the reply: a ready delta and the token counter. The
// envelope belongs to the shape: cli — a terminal turn
// (stdin/stdout), ide — a file pair in the working directory, api —
// strict JSON. The slot is never translated. The exchange key routes
// parallel ide-shape calls to their own files.
func CallExecutor(shape, workdir, executor, slot, lastReply, exchange string) (string, int64, error) {
	switch shape {
	case brief.ShapeIDE:
		dir := filepath.Join(workdir, ".punchtape", "caller")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", 0, fmt.Errorf("caller dir: %w", err)
		}
		challenge := filepath.Join(dir, brief.IDEChallengeFile)
		replyPath := filepath.Join(dir, brief.IDEReplyFile)
		if exchange != "" {
			challenge = filepath.Join(dir, "slot-"+exchange+".md")
			replyPath = filepath.Join(dir, "delta-"+exchange+".yaml")
		}
		// A clean turn: a past delta cannot pass for a reply.
		_ = os.Remove(replyPath)
		if err := os.WriteFile(challenge, []byte(brief.MakeFor(shape, slot, lastReply)), 0o644); err != nil {
			return "", 0, fmt.Errorf("caller slot: %w", err)
		}
		if out, err := runExecutor(executor, workdir, ""); err != nil {
			return "", 0, fmt.Errorf("%s: %w", oneLine(out), err)
		}
		data, err := os.ReadFile(replyPath)
		if err != nil {
			return "", 0, fmt.Errorf("no reply file %s: %w", filepath.Base(replyPath), err)
		}
		text, tokens, _, _ := brief.SplitTokenReport(string(data))
		return brief.ExtractDelta(text), tokens, nil
	case brief.ShapeAPI:
		out, err := runExecutor(executor, workdir, brief.MakeFor(shape, slot, lastReply))
		if err != nil {
			return "", 0, fmt.Errorf("%s: %w", oneLine(out), err)
		}
		delta, tokens, err := brief.DecodeAPIReply(out)
		if err != nil {
			return "", 0, err
		}
		return delta, tokens, nil
	default: // brief.ShapeCLI
		out, err := runExecutor(executor, workdir, brief.MakeFor(shape, slot, lastReply))
		if err != nil {
			return "", 0, fmt.Errorf("%s: %w", oneLine(out), err)
		}
		deltaText, tokens, _, _ := brief.SplitTokenReport(out)
		return brief.ExtractDelta(deltaText), tokens, nil
	}
}

// runExecutor launches the caller command: input on stdin, combined
// output outward. The working directory is the instance directory:
// the caller works in the project workspace. The call is bounded by
// the data timeout (executor.timeout-s): a hung executor is stopped
// together with its process group and the cycle continues with the
// next iteration — the loop does not hang on one hung call.
func runExecutor(executor, workdir, input string) (string, error) {
	cmd := exec.Command("sh", "-c", executor)
	cmd.Dir = workdir
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if input != "" {
		cmd.Stdin = strings.NewReader(input)
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Start(); err != nil {
		return "", err
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		return out.String(), err
	case <-time.After(time.Duration(canondata.Limit("executor.timeout-s")) * time.Second):
		if cmd.Process != nil {
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
			_ = cmd.Process.Kill()
		}
		return out.String(), &ExecutorTimeoutError{Seconds: canondata.Limit("executor.timeout-s")}
	}
}

// ExecutorTimeoutError — an executor call exceeded the time budget:
// the iteration is lost, the cycle continues (not a permanent mode
// failure).
type ExecutorTimeoutError struct{ Seconds int }

func (e *ExecutorTimeoutError) Error() string {
	return canondata.T("driver.executor-timeout", canondata.M{
		"sec": fmt.Sprintf("%d", e.Seconds),
	})
}

func oneLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}
