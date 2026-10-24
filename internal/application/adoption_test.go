package application

// The adoption battery, opt-in: an external agent — a command named
// by the PUNCHTAPE_AGENT variable — walks the machine's states by
// its outputs alone. Each scenario fixes a zone state, hands the
// agent one turn, and grades the zone mechanically: the correct next
// action is whatever the machine's own instruction demanded. The
// runner contract: PUNCHTAPE_AGENT <casedir> <promptfile>; the prompt
// names the machine's binary; the agent may read and write anything
// inside the case directory and nothing outside it. Skipped unless
// the variable names a command — the default battery stays fast.

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/neurophant/punchtape/internal/domain/doc"
	"github.com/neurophant/punchtape/internal/domain/journal"
	"github.com/neurophant/punchtape/internal/domain/zone"
	"github.com/neurophant/punchtape/internal/infrastructure/store"
	"github.com/neurophant/punchtape/internal/infrastructure/store/cas"
)

// adoptionTurn — one external agent's turn over a case directory.
// The prompt file sits beside the case directory, never inside it:
// the tree the machine walks must stay the project's own.
func adoptionTurn(t *testing.T, runner, dir, prompt string) {
	t.Helper()
	promptFile := filepath.Join(filepath.Dir(dir), filepath.Base(dir)+".prompt")
	if err := os.WriteFile(promptFile, []byte(prompt), 0o644); err != nil {
		t.Fatal(err)
	}
	fields := strings.Fields(runner)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, fields[0], append(fields[1:], dir, promptFile)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("the agent's turn failed: %v\n%s", err, out)
	}
}

// adoptionPrompt — the one turn's instruction: the machine's binary
// is the only authority named.
func adoptionPrompt(bin, task string) string {
	return fmt.Sprintf("Work ONLY inside the case directory given as your first argument. "+
		"The machine punchtape is at %s — run it with no arguments from the case directory "+
		"and follow what its terminal line and its zone files say; they are the only authority "+
		"(do not read punchtape's source code or any documentation outside the case directory). "+
		"When your turn is done, write your one-line report to the file named by your "+
		"prompt file's path with .answer appended, then stop. Task: %s", bin, task)
}

// adoptionGrades — the mechanical graders, one per scenario.
func journalExists(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, zone.Dirname, zone.JournalDir))
	return err == nil
}

func stateOf(t *testing.T, dir string) string {
	t.Helper()
	return zoneFile(t, dir, zone.StateFile)
}

// TestExternalAgentAdoptsTheMachine — every scenario of the machine's
// life, one graded turn each.
func TestExternalAgentAdoptsTheMachine(t *testing.T) {
	runner := os.Getenv("PUNCHTAPE_AGENT")
	if runner == "" {
		t.Skip("opt-in: PUNCHTAPE_AGENT names the external agent command (<command> <casedir> <promptfile>)")
	}
	mountPorts()
	bin := filepath.Join(t.TempDir(), "punchtape")
	repo := ".."
	for depth := 0; depth < 8; depth++ {
		if _, err := os.Stat(filepath.Join(repo, "go.mod")); err == nil {
			break
		}
		repo = filepath.Join(repo, "..")
	}
	build := exec.Command("go", "build", "-o", bin, "./cmd/punchtape")
	build.Dir = repo
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("the machine's binary must build: %v\n%s", err, out)
	}

	// The wish every scenario shares.
	const wish = "a tiny counter tool: count.sh prints how many lines count.txt has"

	scenarios := []struct {
		name  string
		setup func(t *testing.T, dir string)
		task  string
		grade func(t *testing.T, dir string)
	}{
		{
			name:  "first contact",
			setup: func(t *testing.T, dir string) {},
			task:  "the human's wish, verbatim: \"" + wish + "\" — establish the project through the machine",
			grade: func(t *testing.T, dir string) {
				wishBytes, err := os.ReadFile(filepath.Join(dir, zone.Dirname, zone.WishSlot))
				if err == nil && string(wishBytes) == wish && !journalExists(dir) {
					return // the correct next action, not yet run again
				}
				if !journalExists(dir) {
					t.Fatalf("first contact: neither the wish slot nor the birth happened")
				}
				if stateOf(t, dir) == "" || !strings.Contains(stateOf(t, dir), "goal:") {
					t.Fatalf("first contact: the birth must fix the goal")
				}
			},
		},
		{
			name: "birth",
			setup: func(t *testing.T, dir string) {
				writeZone(t, dir, zone.WishSlot, wish)
			},
			task: "run the machine and do what it says until the instance lives",
			grade: func(t *testing.T, dir string) {
				if !journalExists(dir) {
					t.Fatal("birth: the journal must exist")
				}
				if _, err := os.Stat(filepath.Join(dir, zone.Dirname, zone.WishSlot)); !os.IsNotExist(err) {
					t.Fatal("birth: the wish slot must be consumed")
				}
				for _, name := range []string{zone.ContractFile, zone.TemplateFile, zone.CorpusFile, zone.StateFile} {
					if _, err := os.Stat(filepath.Join(dir, zone.Dirname, name)); err != nil {
						t.Fatalf("birth: %s must stand: %v", name, err)
					}
				}
			},
		},
		{
			name: "submission",
			setup: func(t *testing.T, dir string) {
				writeZone(t, dir, zone.WishSlot, wish)
				runIn(t, dir)
			},
			task: "write the project's first document (one behavior is enough — a trivially true row) and get the machine to accept it",
			grade: func(t *testing.T, dir string) {
				state := stateOf(t, dir)
				if !strings.Contains(state, "corpus: v1") || !strings.Contains(state, "1 behaviors") {
					t.Fatalf("submission: the corpus must carry the first behavior, state:\n%s", state)
				}
				if _, err := os.Stat(filepath.Join(dir, zone.Dirname, zone.InboxSlot)); !os.IsNotExist(err) {
					t.Fatal("submission: the accepted inbox must be consumed")
				}
			},
		},
		{
			name: "refusal",
			setup: func(t *testing.T, dir string) {
				writeZone(t, dir, zone.WishSlot, wish)
				runIn(t, dir)
			},
			task: "submit EXACTLY this document first — behaviors:\n  - id: tool/x\n    name: X\n    rows:\n      - run: \"true\"\n        expect:\n          files:\n            /abs/path: \"contains: x\"\n— observe the machine's answer, then continue correctly until a submission is accepted",
			grade: func(t *testing.T, dir string) {
				state := stateOf(t, dir)
				if !strings.Contains(state, "corpus: v1") {
					t.Fatalf("refusal: the agent must recover to an accepted submission, state:\n%s", state)
				}
			},
		},
		{
			name: "iteration",
			setup: func(t *testing.T, dir string) {
				writeZone(t, dir, zone.WishSlot, wish)
				runIn(t, dir)
				writeZone(t, dir, zone.InboxSlot, `behaviors:
    - id: tool/broken
      name: Broken
      prose: Expects the impossible.
      rows:
        - run: "true"
          expect:
            rc: 1
`)
				runIn(t, dir)
			},
			task: "the machine shows a red row; make the corpus honest and green (fix the expectation or the code — your judgment)",
			grade: func(t *testing.T, dir string) {
				state := stateOf(t, dir)
				if strings.Contains(state, "RED") {
					t.Fatalf("iteration: the red must be gone, state:\n%s", state)
				}
				if !strings.Contains(state, "1 green") {
					t.Fatalf("iteration: the row must stand green, state:\n%s", state)
				}
			},
		},
		{
			name: "row edit",
			setup: func(t *testing.T, dir string) {
				writeZone(t, dir, zone.WishSlot, wish)
				runIn(t, dir)
				if err := os.WriteFile(filepath.Join(dir, "count.txt"), []byte("one\ntwo\n"), 0o644); err != nil {
					t.Fatal(err)
				}
				writeZone(t, dir, zone.InboxSlot, `behaviors:
    - id: tool/three
      name: Three rows
      prose: Two stand green, one is red.
      scope:
        - count.txt
      rows:
        - run: "true"
          expect:
            rc: 0
        - run: sh -c "echo one"
          expect:
            out: "contains: one"
        - run: "false"
          expect:
            rc: 0
`)
				runIn(t, dir)
			},
			task: "one row of tool/three is red; heal it by submitting ONLY that one row — rows merge by run, the untouched rows stand",
			grade: func(t *testing.T, dir string) {
				state := stateOf(t, dir)
				if strings.Contains(state, "RED") || !strings.Contains(state, "3 green") {
					t.Fatalf("row edit: every row must stand green, state:\n%s", state)
				}
				// The healing submission listed exactly the broken row —
				// the journal's own doc objects are the witness.
				stream, err := store.OpenStream(filepath.Join(dir, zone.Dirname, zone.JournalDir))
				if err != nil {
					t.Fatal(err)
				}
				records, err := stream.Records()
				if err != nil {
					t.Fatal(err)
				}
				objects := cas.New(filepath.Join(dir, zone.Dirname, zone.ObjectsDir))
				healedByOneRow := false
				for _, rec := range records {
					if rec.Kind != journal.KindSubmission || rec.Submission == nil {
						continue
					}
					data, err := objects.Get(rec.Submission.DocAddr)
					if err != nil {
						t.Fatal(err)
					}
					d, err := doc.Decode(data)
					if err != nil {
						t.Fatal(err)
					}
					for _, b := range d.Behaviors {
						if b.ID == "tool/three" && len(b.Rows) == 1 {
							healedByOneRow = true
						}
					}
				}
				if !healedByOneRow {
					t.Fatal("row edit: a submission listing exactly the one broken row must be journaled")
				}
			},
		},
		{
			name: "completion act",
			setup: func(t *testing.T, dir string) {
				writeZone(t, dir, zone.WishSlot, wish)
				runIn(t, dir)
				writeZone(t, dir, zone.InboxSlot, `behaviors:
    - id: tool/ping
      name: Ping
      prose: A trivially true row.
      rows:
        - run: "true"
          expect:
            rc: 0
`)
				runIn(t, dir)
			},
			task: "the corpus is green; obtain the machine's final word on it",
			grade: func(t *testing.T, dir string) {
				if !strings.Contains(stateOf(t, dir), "verdict: #1") {
					t.Fatalf("completion: the verdict must be recorded, state:\n%s", stateOf(t, dir))
				}
			},
		},
		{
			name: "standing verdict",
			setup: func(t *testing.T, dir string) {
				writeZone(t, dir, zone.WishSlot, wish)
				runIn(t, dir)
				writeZone(t, dir, zone.InboxSlot, `requirements:
    - id: R1
      brief: [W1]
behaviors:
    - id: tool/ping
      name: Ping
      prose: A trivially true row.
      covers: [R1]
      rows:
        - run: "true"
          expect:
            rc: 0
`)
				runIn(t, dir)
				writeZone(t, dir, zone.InboxSlot, zoneFile(t, dir, zone.CorpusFile))
				runIn(t, dir)
			},
			task: "a fresh session: run the machine once and report the standing to the human; do not start new work",
			grade: func(t *testing.T, dir string) {
				state := stateOf(t, dir)
				if !strings.Contains(state, "verdict: #1") || !strings.Contains(state, "standing") {
					t.Fatalf("standing verdict: the verdict must stand, state:\n%s", state)
				}
				if strings.Contains(state, "verdict: #2") {
					t.Fatal("standing verdict: no new verdict was asked for")
				}
				// An idle run leaves no zone trace by design: the turn's
				// own trace is the answer line.
				answer, err := os.ReadFile(filepath.Join(filepath.Dir(dir), filepath.Base(dir)+".prompt.answer"))
				if err != nil || !strings.Contains(string(answer), "verdict") {
					t.Fatalf("standing verdict: the agent must report the standing: %v %q", err, answer)
				}
			},
		},
		{
			name: "drift",
			setup: func(t *testing.T, dir string) {
				writeZone(t, dir, zone.WishSlot, wish)
				runIn(t, dir)
				if err := os.WriteFile(filepath.Join(dir, "count.txt"), []byte("one\ntwo\n"), 0o644); err != nil {
					t.Fatal(err)
				}
				writeZone(t, dir, zone.InboxSlot, `requirements:
    - id: R1
      brief: [W1]
behaviors:
    - id: tool/count
      name: Count
      prose: Counts the lines.
      scope:
        - count.txt
      covers: [R1]
      rows:
        - run: sh -c "wc -l < count.txt"
          expect:
            out: "2"
`)
				runIn(t, dir)
				writeZone(t, dir, zone.InboxSlot, zoneFile(t, dir, zone.CorpusFile))
				runIn(t, dir)
				os.WriteFile(filepath.Join(dir, "count.txt"), []byte("one\n"), 0o644)
			},
			task: "someone edited the tree after the verdict; re-establish the machine's truth over the changed tree",
			grade: func(t *testing.T, dir string) {
				state := stateOf(t, dir)
				red := strings.Contains(state, "RED")
				healed := strings.Contains(state, "1 green") && !red
				if !healed {
					t.Fatalf("drift: the changed tree must be re-verified green, state:\n%s", state)
				}
			},
		},
		{
			name: "new agent on a live instance",
			setup: func(t *testing.T, dir string) {
				writeZone(t, dir, zone.WishSlot, wish)
				runIn(t, dir)
				writeZone(t, dir, zone.InboxSlot, `behaviors:
    - id: tool/ok
      name: Ok
      prose: Stands green.
      rows:
        - run: "true"
          expect:
            rc: 0
    - id: tool/bad
      name: Bad
      prose: Stands red.
      rows:
        - run: "true"
          expect:
            rc: 1
`)
				runIn(t, dir)
			},
			task: "you inherit this project mid-work with no context but the directory; continue it correctly",
			grade: func(t *testing.T, dir string) {
				state := stateOf(t, dir)
				if strings.Contains(state, "RED") {
					t.Fatalf("new agent: the inherited red must be resolved, state:\n%s", state)
				}
				if !strings.Contains(state, "tool/ok") {
					t.Fatalf("new agent: the standing work must survive, state:\n%s", state)
				}
			},
		},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			dir := t.TempDir()
			sc.setup(t, dir)
			adoptionTurn(t, runner, dir, adoptionPrompt(bin, sc.task))
			sc.grade(t, dir)
		})
	}
}
