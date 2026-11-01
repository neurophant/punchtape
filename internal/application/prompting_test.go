// The instruction completeness battery: in every state the machine
// can stand in, the terminal's one line and state.md's next: line
// are one imperative that names the file to touch — an agent that
// only obeys the last line it read can never be lost, whatever
// happened. The check is mechanical: every pointer line asserted by
// the case batteries carries a verb and a zone path; here the law
// itself is pinned for the states the other batteries reach.
package application

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neurophant/punchtape/internal/domain/zone"
)

// imperative — the one line carries an action verb and names a file
// of the zone.
func imperative(t *testing.T, line string) {
	t.Helper()
	verbs := []string{"put", "submit", "read", "copy", "run", "fix", "record", "show"}
	sawVerb := false
	for _, v := range verbs {
		if strings.Contains(line, v+" ") {
			sawVerb = true
		}
	}
	if !sawVerb {
		t.Fatalf("the instruction carries no action verb: %q", line)
	}
	if !strings.Contains(line, zone.Dirname+"/") && !strings.Contains(line, "punchtape") {
		t.Fatalf("the instruction names no file: %q", line)
	}
}

func TestEveryStateInstructs(t *testing.T) {
	mountPorts()

	// No instance: the terminal line says wish.
	root := t.TempDir()
	writeZone(t, root, ".keep", "")
	if err := removeFile(root, zone.Dirname+"/.keep"); err != nil {
		t.Fatal(err)
	}
	_, term := runInFail(t, root)
	imperative(t, term)

	// The birth's answer instructs.
	writeZone(t, root, zone.DemandsSlot, "goal")
	_, term = runInFail(t, root)
	imperative(t, term)

	// The empty corpus instructs; state.md's next: is the same law.
	_, term = runInFail(t, root)
	imperative(t, term)
	next := lastNext(zoneFile(t, root, zone.StateFile))
	imperative(t, next)

	// A red row instructs; a green corpus instructs the completion.
	if err := writeTree(t, root, "src/in.txt", "ok"); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, redDocFor(t, root))
	runInFail(t, root)
	imperative(t, lastNext(zoneFile(t, root, zone.StateFile)))
	writeZone(t, root, zone.InboxSlot, greenDocFor(t, root))
	runInFail(t, root)
	imperative(t, lastNext(zoneFile(t, root, zone.StateFile)))

	// The stream's account joins the corpus; the completion act
	// records the verdict. With the stream resolved, the green
	// standing instructs the act itself.
	writeZone(t, root, zone.InboxSlot, regDocFor(t, root))
	runInFail(t, root)
	next = lastNext(zoneFile(t, root, zone.StateFile))
	imperative(t, next)
	if !strings.Contains(next, "done: true") {
		t.Fatalf("the green standing must instruct the completion act: %q", next)
	}
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	runInFail(t, root)
	if !strings.Contains(zoneFile(t, root, zone.StateFile), "verdict: #1") {
		t.Fatalf("the completion act must record the verdict")
	}

	// The standing verdict instructs the delivery — not a repeat of
	// the act that already happened.
	next = lastNext(zoneFile(t, root, zone.StateFile))
	imperative(t, next)
	if !strings.Contains(next, "standing") || strings.Contains(next, "record the verdict") {
		t.Fatalf("the standing verdict must instruct the delivery: %q", next)
	}

	// The machine's own record file is never the agent's gap.
	if strings.Contains(zoneFile(t, root, zone.StateFile), "- AGENTS.md") {
		t.Fatalf("the machine's own marker is not an uncovered fact of the project")
	}
}

// removeFile — drop one file of the root.
func removeFile(root, rel string) error {
	return os.Remove(filepath.Join(root, rel))
}

// writeTree — one tree file, directories made on the way.
func writeTree(t *testing.T, root, rel, body string) error {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(body), 0o644)
}

// runInFail — a run that must answer zero; the terminal line is
// returned.
func runInFail(t *testing.T, root string) (int, string) {
	t.Helper()
	code, term := runCode(t, root, nil)
	if code != 0 {
		t.Fatalf("run exited %d: %s", code, term)
	}
	return code, term
}

// lastNext — the next: line of a state page.
func lastNext(state string) string {
	for _, line := range strings.Split(state, "\n") {
		if strings.HasPrefix(line, "next: ") {
			return strings.TrimPrefix(line, "next: ")
		}
	}
	return ""
}

// redDocFor — the first, honestly red copy: the pin wants bytes the
// tree does not carry, the proof row prints its claims token.
func redDocFor(t *testing.T, root string) string {
	t.Helper()
	return fmt.Sprintf(`
work: the red copy first
behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies.
      claims:
        - src/in.txt:1-1
      rows:
        - run: sh -c 'printf "%%s\n" %s; cat src/in.txt > out.txt'
          expect:
            files:
                out.txt: NOPE
`, claimsTokenOf(t, root, "src/in.txt:1-1"))
}

// greenDocFor — the healed pin: the same row with the tree's bytes.
func greenDocFor(t *testing.T, root string) string {
	t.Helper()
	return fmt.Sprintf(`
work: the green heal
mandate: [{line: D1, target: tool/copy}]
behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies.
      claims:
        - src/in.txt:1-1
      rows:
        - run: sh -c 'printf "%%s\n" %s; cat src/in.txt > out.txt'
          expect:
            files:
                out.txt: ok
`, claimsTokenOf(t, root, "src/in.txt:1-1"))
}

// regDocFor — the stream's own account: the one line declared work
// over its requirement, the behavior covering it with a row that
// answers the demand's own token, the tree's unowned files excluded
// loudly.
func regDocFor(t *testing.T, root string) string {
	t.Helper()
	return fmt.Sprintf(`
work: the registry closes the account
mandate: [{line: D1, target: tool/copy}]
intents:
    - {lines: [D1], outcome: work, prose: "the instance's one wish, told again"}
uncovered-lines:
    - {range: "AGENTS.md:1-9", reason: the machine's own hook, refs: [D1]}
    - {range: "out.txt:1-1", reason: the row's own output, refs: [D1]}
requirements:
    - id: R1
      lines: [D1]
      prose: the one demand of the instance
behaviors:
    - id: tool/copy
      name: The copy
      prose: Copies.
      claims:
        - src/in.txt:1-1
      covers: [R1]
      rows:
        - run: sh -c 'printf "%%s %%s\n" %s %s; cat src/in.txt > out.txt'
          expect:
            files:
                out.txt: ok
`, claimsTokenOf(t, root, "src/in.txt:1-1"), demandTokenOf("goal"))
}
