// The environment battery: the zone's pin file .punchtape/env rides
// every command (strict KEY=VALUE — comments and blanks pass,
// everything else is a loud refusal naming the file, the line and
// the law), the environment hash moves with the pins — a changed
// pin file is a changed environment, proofs follow, never falsely
// green — and the working root's own ENV-named file is nothing to
// the machine: no variable it names rides a row, no hash moves.
package application

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neurophant/punchtape/internal/domain/zone"
)

// writePins — the zone's pin file.
func writePins(t *testing.T, root, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, zone.Dirname, zone.EnvSlot), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestZonePinsRideEveryRow(t *testing.T) {
	mountPorts()
	root := newRoot(t)
	if err := os.WriteFile(filepath.Join(root, "src.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, fmt.Sprintf(`
work: the pinned row
behaviors:
    - id: tool/env
      name: The zone environment
      prose: Rows see the zone pins.
      claims:
        - src.txt:1-1
      rows:
        - run: ["sh", "-c", "printf '%%s\\n' %s; printf %%s \"$PTA\" > out.txt"]
          expect:
            files:
                out.txt: from-the-zone
`, claimsTokenOf(t, root, "src.txt:1-1")))
	writePins(t, root, "# a comment is not a pin\nPTA=from-the-zone\n\n")
	runIn(t, root)
	if report := zoneFile(t, root, zone.ReportFile); !strings.Contains(report, "green: 1") {
		t.Fatalf("the row must see the zone pin:\n%s", report)
	}
}

func TestEnvHashMovesWithPins(t *testing.T) {
	mountPorts()
	root := newRoot(t)
	without := EnvHash(nil)
	writePins(t, root, "PTA=one\n")
	one := pinsHash(t, root)
	writePins(t, root, "PTA=two\n")
	two := pinsHash(t, root)
	if one == without || two == without || one == two {
		t.Fatal("a changed pin file is a changed environment")
	}
	// The same pins hash the same on every parse.
	if pinsHash(t, root) != two {
		t.Fatal("the same pins must hash identically")
	}
}

func pinsHash(t *testing.T, root string) string {
	t.Helper()
	pins, err := ZoneEnv(filepath.Join(root, zone.Dirname))
	if err != nil {
		t.Fatal(err)
	}
	return EnvHash(pins)
}

func TestZonePinRefusalsAreLoud(t *testing.T) {
	mountPorts()
	for name, content := range map[string]string{
		"export":        "export PTA=1\n",
		"quoted value":  "PTA=\"once\"\n",
		"single quotes": "PTA='once'\n",
		"bad key":       "PT-A=1\n",
		"leading digit": "2PA=1\n",
		"absolute path": "PTA=/absolutely/one/disk\n",
		"no equals":     "no-equals-line\n",
		"duplicate pin": "PTA=one\nPTA=two\n",
	} {
		t.Run(name, func(t *testing.T) {
			root := newBornRoot(t)
			writePins(t, root, content)
			code, term := runCode(t, root, nil)
			if code != 1 {
				t.Fatalf("a bad pin file must refuse the run, exit %d: %s", code, term)
			}
			// The refused line is the file's own: the last of its
			// pin lines (a duplicate names its second spelling).
			lineNo := strings.Count(content, "\n")
			for _, want := range []string{"env pin refused", fmt.Sprintf("line %d", lineNo), ".punchtape/env"} {
				if !strings.Contains(term, want) {
					t.Fatalf("the refusal must quote %q:\n%s", want, term)
				}
			}
			if !strings.Contains(term, "KEY=VALUE") && !strings.Contains(term, "absolute") && !strings.Contains(term, "quotes") && !strings.Contains(term, "pins once") {
				t.Fatalf("the refusal must state the law it enforces:\n%s", term)
			}
		})
	}
}

func TestRootEnvFileIsDeadToTheMachine(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	// The probe: a KEY=VALUE file named ENV in the working root.
	// It is nothing to the machine — no variable it names rides
	// a row, no hash moves.
	if err := os.WriteFile(filepath.Join(root, "ENV"), []byte("PTA=leaked\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, fmt.Sprintf(`
work: the dead door check
behaviors:
    - id: tool/silent
      name: The dead door
      prose: The root ENV leaks nothing.
      claims:
        - ENV:1-1
      rows:
        - run: ["sh", "-c", "printf '%%s\\n' %s; printf %%s \"${PTA-unset}\" > seen.txt"]
          expect:
            files:
                seen.txt: unset
`, claimsTokenOf(t, root, "ENV:1-1")))
	runIn(t, root)
	if report := zoneFile(t, root, zone.ReportFile); !strings.Contains(report, "green: 1") {
		t.Fatalf("the root ENV file must leak nothing into rows:\n%s", report)
	}
	if pinsHash(t, root) != EnvHash(nil) {
		t.Fatal("the root ENV file must not move the environment hash")
	}
}

func TestRowNotFoundCarriesTheLaw(t *testing.T) {
	mountPorts()
	root := newRoot(t)
	writeZone(t, root, zone.InboxSlot, fmt.Sprintf(`
work: the absent command
behaviors:
    - id: tool/absent
      name: The absent command
      prose: A row calling nothing that exists.
      rows:
        - run: sh -c 'printf "ok %%s\n" %s'
          expect:
            out: "ok %s\n"
        - run: this-command-does-not-exist-anywhere
          expect:
            rc: 127
            out: ""
`, claimsTokenOf(t, root), claimsTokenOf(t, root)))
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "hint: the command was not found") {
		t.Fatalf("a not-found row must carry the machine's own law:\n%s", report)
	}
	for _, want := range []string{"PATH and HOME of the shell", ".punchtape/env pins", "relative"} {
		if !strings.Contains(report, want) {
			t.Fatalf("the hint must be self-quoting (missing %q):\n%s", want, report)
		}
	}
}

func TestUnreadablePinFileRefusesTheRun(t *testing.T) {
	mountPorts()
	root := newBornRoot(t)
	writePins(t, root, "PTA=one\n")
	code, term := runCode(t, root, map[string]bool{"read:/.punchtape/env": true})
	if code != 1 {
		t.Fatalf("an unreadable pin file must refuse the run, exit %d: %s", code, term)
	}
	if !strings.Contains(term, "punchtape: ") {
		t.Fatalf("the refusal must reach the terminal:\n%s", term)
	}
}
