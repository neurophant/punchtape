// The proof economy's battery: the pin is the proof — exact bytes
// and nothing else. The first submission can go green in one run:
// no token to chase, no cycle constant to copy; a twelve-hex word
// in a pin is plain text compared as bytes; a moved claimed line
// makes the proof stale and the re-run must hold the pin again.
// The helpers still compute the machine's internal digests for the
// fixtures that carry them.
package application

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/neurophant/punchtape/internal/domain"
	"github.com/neurophant/punchtape/internal/domain/doc"
	"github.com/neurophant/punchtape/internal/domain/fprint"
	"github.com/neurophant/punchtape/internal/domain/zone"
	"github.com/neurophant/punchtape/internal/infrastructure/fsys"
	"github.com/neurophant/punchtape/internal/ports"
)

// claimsTokenOf — the claims token of a behavior claiming the given
// ranges, by the machine's own fingerprint law over the tree as it
// stands.
func claimsTokenOf(t *testing.T, root string, claims ...string) string {
	t.Helper()
	m := &machine{root: root}
	return fprint.Token12(m.claimsFactsOf(doc.Behavior{Claims: claims}).digest)
}

// demandTokenOf — the requirement's quote token.
func demandTokenOf(quote string) string {
	return fprint.Token12(domain.Digest([]byte(quote)))
}

// proofRoot — a born zone over one checked input line.
func proofRoot(t *testing.T) string {
	t.Helper()
	mountPorts()
	root := newRoot(t)
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

// TestFirstSubmissionCanGoGreen — the proof economy's own acceptance:
// a birth over one line, a submission whose row prints plain bytes
// and pins them — one run, all green, no learn cycle ever.
func TestFirstSubmissionCanGoGreen(t *testing.T) {
	root := proofRoot(t)
	writeZone(t, root, zone.InboxSlot, `
work: the first account
intents:
    - {lines: [D1], outcome: work, prose: the tool and its check of the file}
requirements:
    - id: R1
      lines: [D1]
      prose: the tool and its check
uncovered-lines:
    - {range: "AGENTS.md:1-9", reason: the machine's own hook, refs: [D1]}
    - {range: "out.txt:1-1", reason: the row's own output, refs: [D1]}
behaviors:
    - id: tool/first
      name: The first
      prose: Prints three, pinned by exact bytes.
      claims:
        - src/in.txt:1-1
      covers: [R1]
      rows:
        - run: sh -c 'printf "3\n"; cat src/in.txt > out.txt'
          expect:
            out: "3\n"
            files:
                out.txt: ok
`)
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "green: 1") {
		t.Fatalf("the first submission must go green in one run:\n%s", report)
	}
	if strings.Contains(report, "missing") || strings.Contains(report, "RED") {
		t.Fatalf("no learn cycle may stand in the first run:\n%s", report)
	}
	if _, err := os.Stat(filepath.Join(root, zone.Dirname, zone.InboxSlot)); !os.IsNotExist(err) {
		t.Fatalf("the green first submission must be accepted")
	}
}

// TestNoFreshnessAnywhere — no face of the zone carries freshness:
// state, report, template and the contract page teach none of it.
func TestNoFreshnessAnywhere(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	for _, name := range []string{
		zone.StateFile, zone.ReportFile, zone.TemplateFile,
		zone.ContractFile,
	} {
		if text := zoneFile(t, root, name); strings.Contains(text, "freshness") {
			t.Fatalf("%s must carry no freshness:\n%s", name, text)
		}
	}
}

// TestNoTokensInExpect — a twelve-hex word in a pin is plain text:
// accepted, compared as bytes — the wrong word breaks the pin like
// any other byte would.
func TestNoTokensInExpect(t *testing.T) {
	root := proofRoot(t)
	writeZone(t, root, zone.InboxSlot, `
work: a pin of twelve hex, plain bytes
behaviors:
    - id: tool/hex
      name: The hex
      prose: Prints a word, pinned by it.
      claims:
        - src/in.txt:1-1
      rows:
        - run: sh -c 'printf "abc123def456\n"'
          expect:
            out: "abc123def456\n"
`)
	runIn(t, root)
	if state := zoneFile(t, root, zone.StateFile); !strings.Contains(state, "tool/hex: 1 rows — 1 green") {
		t.Fatalf("a twelve-hex pin is plain text and stands green:\n%s", state)
	}
	writeZone(t, root, zone.InboxSlot, `
work: the same pin off by a word
mandate: [{line: D1, target: tool/hex}]
behaviors:
    - id: tool/hex
      name: The hex
      prose: Prints a word, pinned by it.
      claims:
        - src/in.txt:1-1
      rows:
        - run: sh -c 'printf "abc123def456\n"'
          expect:
            out: "fff123def456\n"
`)
	runIn(t, root)
	if report := zoneFile(t, root, zone.ReportFile); !strings.Contains(report, `want "fff123def456`) {
		t.Fatalf("the wrong twelve-hex word must break the pin as plain text:\n%s", report)
	}
}

// TestStalenessStillWorks — the digests' own economy stands: a
// claimed line moved by hand makes the proof stale, the row re-runs,
// the broken pin lands red; the restored line re-earns the green.
func TestStalenessStillWorks(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	if state := zoneFile(t, root, zone.StateFile); !strings.Contains(state, "tool/copy: 1 rows — 1 green") {
		t.Fatalf("the standing row must be green before the move:\n%s", state)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("NO"), 0o644); err != nil {
		t.Fatal(err)
	}
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "tool/copy") || !strings.Contains(report, `want "ok"`) {
		t.Fatalf("the moved claimed line must stale the proof and break the pin:\n%s", report)
	}
	if state := zoneFile(t, root, zone.StateFile); !strings.Contains(state, "1 RED") {
		t.Fatalf("the re-run over the moved line must land red:\n%s", state)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "in.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	runIn(t, root)
	if state := zoneFile(t, root, zone.StateFile); !strings.Contains(state, "tool/copy: 1 rows — 1 green") {
		t.Fatalf("the restored line must re-earn the green pin:\n%s", state)
	}
}

// TestRunReuseAcrossBehaviorsRefused — one command proves one
// behavior: two behaviors in one submission sharing a run are
// refused by the later holder; a stranger meets the standing
// holder's refusal; the holder's retirement frees the run.
func TestRunReuseAcrossBehaviorsRefused(t *testing.T) {
	root := proofRoot(t)
	writeZone(t, root, zone.InboxSlot, `
work: two heads over one command
behaviors:
    - id: tool/one
      name: One
      prose: Holds the command first.
      claims:
        - src/in.txt:1-1
      rows:
        - run: cat src/in.txt > one.txt
          expect:
            files:
                one.txt: ok
    - id: tool/two
      name: Two
      prose: Wants the same command.
      claims:
        - src/in.txt:1-1
      rows:
        - run: cat src/in.txt > one.txt
          expect:
            files:
                one.txt: ok
`)
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "behaviors[1].rows[0].run: the run is already the proof of tool/one — one command proves one behavior (law:proof)") {
		t.Fatalf("the shared command inside one submission must be refused by name:\n%s", report)
	}
	writeZone(t, root, zone.InboxSlot, `
work: the honest first account
intents:
    - {lines: [D1], outcome: work, prose: the tool and its check of the file}
requirements:
    - id: R1
      lines: [D1]
      prose: the tool and its check
uncovered-lines:
    - {range: "AGENTS.md:1-9", reason: the machine's own hook, refs: [D1]}
    - {range: "one.txt:1-1", reason: the row's own output, refs: [D1]}
behaviors:
    - id: tool/one
      name: One
      prose: Holds the command.
      claims:
        - src/in.txt:1-1
      covers: [R1]
      rows:
        - run: cat src/in.txt > one.txt
          expect:
            files:
                one.txt: ok
`)
	runIn(t, root)
	writeZone(t, root, zone.InboxSlot, `
work: a stranger wants the standing command
behaviors:
    - id: tool/two
      name: Two
      prose: Wants the standing command.
      claims:
        - src/in.txt:1-1
      rows:
        - run: cat src/in.txt > one.txt
          expect:
            files:
                one.txt: ok
`)
	runIn(t, root)
	report = zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "behaviors[0].rows[0].run: the run is already the proof of tool/one — one command proves one behavior (law:proof)") {
		t.Fatalf("the standing command must refuse a stranger:\n%s", report)
	}
	writeZone(t, root, zone.InboxSlot, `
work: the retirement frees the command
mandate: [{line: D1, target: tool/one}]
remove: [tool/one]
uncovered-lines:
    - {range: "AGENTS.md:1-9", reason: the machine's own hook, refs: [D1]}
    - {range: "one.txt:1-1", reason: the row's own output, refs: [D1]}
`)
	runIn(t, root)
	if state := zoneFile(t, root, zone.StateFile); strings.Contains(state, "tool/one") {
		t.Fatalf("the retirement must free the run, state:\n%s", state)
	}
}

// blindFS — the submitting side's own filesystem: the wish slot and
// the template answer, every other machine face refuses — a
// submission that would feed on the machine's own output cannot
// even compose.
type blindFS struct {
	ports.FS
}

func (f blindFS) ReadFile(p string) ([]byte, error) {
	if strings.Contains(p, zone.Dirname+"/") &&
		!strings.HasSuffix(p, "/"+zone.DemandsSlot) &&
		!strings.HasSuffix(p, "/"+zone.TemplateFile) {
		return nil, fmt.Errorf("the submitting side reads no machine face: %s", p)
	}
	return f.FS.ReadFile(p)
}

// TestSubmissionNeedsNoMachineOutputs — the rebuild's main
// invariant: a whole green cycle — the birth by a word, the specs,
// the green row, the act, the verdict — closes over submissions
// whose author read nothing but the wish and the shape. The
// composing side runs behind a filesystem that refuses every other
// machine face, and the composed bytes carry no machine constant: a
// future law that feeds a submission on the machine's own output
// breaks this build, never the agent.
func TestSubmissionNeedsNoMachineOutputs(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, zone.Dirname), 0o755); err != nil {
		t.Fatal(err)
	}
	const wish = "the counter prints three, pinned by exact bytes\n"
	if err := os.WriteFile(filepath.Join(root, zone.Dirname, zone.DemandsSlot), []byte(wish), 0o644); err != nil {
		t.Fatal(err)
	}
	// The agent's own work: the counter script, two lines of code the
	// behavior claims whole.
	if err := os.WriteFile(filepath.Join(root, "three"), []byte("#!/bin/sh\nprintf '3\\n'\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	// The birth by the word: the machine answers with its one line,
	// the template stands, and the submitting side has its two
	// readable surfaces.
	mountPorts()
	runIn(t, root)

	// The composing phase — every read behind the blind filesystem.
	ports.FSBind = blindFS{FS: fsys.OS{}}
	words, err := ports.FSBind.ReadFile(filepath.Join(root, zone.Dirname, zone.DemandsSlot))
	if err != nil || string(words) != wish {
		t.Fatalf("the wish must answer the composing side byte for byte: %v %q", err, words)
	}
	if _, err := ports.FSBind.ReadFile(filepath.Join(root, zone.Dirname, zone.TemplateFile)); err != nil {
		t.Fatalf("the shape must answer the composing side: %v", err)
	}
	retold := strings.TrimSpace(string(words))
	doc := "work: the counter's first account — " + retold + `
intents:
    - {lines: [D1], outcome: work, prose: the counter and its exact-bytes check}
requirements:
    - id: R1
      lines: [D1]
      prose: the counter prints three and the pin holds its bytes
uncovered-lines:
    - {range: "AGENTS.md:1-9", reason: the machine's own hook, refs: [D1]}
behaviors:
    - id: tool/three
      name: The counter
      prose: Prints three, pinned by exact bytes.
      claims:
        - three:1-2
      covers: [R1]
      rows:
        - run: ./three
          expect:
            rc: 0
            out: "3\n"
`
	if regexp.MustCompile(`\b[0-9a-f]{12}\b`).MatchString(doc) {
		t.Fatal("the composed document must carry no machine constant")
	}
	writeZone(t, root, zone.InboxSlot, doc)
	mountPorts()
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "green: 1") || strings.Contains(report, "RED") || strings.Contains(report, "missing") {
		t.Fatalf("the blind-composed submission must run green in one breath:\n%s", report)
	}
	// The act composes blind as a bare literal — no read at all.
	ports.FSBind = blindFS{FS: fsys.OS{}}
	writeZone(t, root, zone.InboxSlot, "done: true\n")
	mountPorts()
	runIn(t, root)
	state := zoneFile(t, root, zone.StateFile)
	if !strings.Contains(state, "verdict: #1") {
		t.Fatalf("the whole cycle must close without one machine read, state:\n%s", state)
	}
}
