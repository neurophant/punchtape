// The execution battery: the row's argv spelling runs for real, and
// a wait that could not be answered is a red row carrying the
// failure's own words — never a silent green.
package application

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/neurophant/punchtape/internal/domain/journal"
	"github.com/neurophant/punchtape/internal/domain/zone"
	"github.com/neurophant/punchtape/internal/infrastructure/fsys"
	"github.com/neurophant/punchtape/internal/ports"
)

func TestArgvListRowRuns(t *testing.T) {
	mountPorts()
	root := newRoot(t)
	if err := os.WriteFile(filepath.Join(root, "src.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, fmt.Sprintf(`
work: the argv spelling
behaviors:
    - id: tool/argv
      name: The argv
      prose: Runs a list.
      claims:
        - src.txt:1-1
      rows:
        - run: ["sh", "-c", "printf '%%s\\n' %s; printf hi > out.txt"]
          expect:
            files:
                out.txt: hi
`, claimsTokenOf(t, root, "src.txt:1-1")))
	runIn(t, root)
	if report := zoneFile(t, root, zone.ReportFile); !strings.Contains(report, "green: 1") {
		t.Fatalf("the argv spelling must run green:\n%s", report)
	}
}

// fakeWaitExec — the process port whose wait cannot be answered.
type fakeWaitExec struct{}

func (fakeWaitExec) StartOpts(argv []string, dir string, opts ports.LaunchOpts) (ports.Process, error) {
	return fakeWaitProc{}, nil
}

type fakeWaitProc struct{}

func (fakeWaitProc) Wait(time.Duration) ports.Outcome {
	return ports.Outcome{WaitErr: "the process could not be reaped"}
}

func (fakeWaitProc) KillGroup() {}

func TestWaitFailureIsARedRow(t *testing.T) {
	root := newRoot(t)
	ports.FSBind = fsys.OS{}
	ports.ExecBind = fakeWaitExec{}
	defer mountPorts()
	writeZone(t, root, zone.InboxSlot, `
work: the wait that cannot be answered
behaviors:
    - id: tool/wait
      name: The waiter
      prose: Waits.
      rows:
        - run: "true"
          expect:
            rc: 0
            out: "x"
`)
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "launch: wait: the process could not be reaped") {
		t.Fatalf("a wait failure must be a named red row:\n%s", report)
	}
}

// TestRowsRunSequentially — the rows run one at a time, in corpus
// order, and the journal's run fact carries that order: no row starts
// before the one before it answers, and no agent ever needs a lock of
// its own — the machine is the serialization.
func TestRowsRunSequentially(t *testing.T) {
	mountPorts()
	root := newRoot(t)
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "a.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "b.txt"), []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeZone(t, root, zone.InboxSlot, `
work: two behaviors, two rows each, one order
intents:
    - {lines: [D1], outcome: work, prose: the tools and their checks}
requirements:
    - id: R1
      lines: [D1]
      prose: both tools answer in order
behaviors:
    - id: tool/aa
      name: The first
      prose: Prints twice.
      claims:
        - src/a.txt:1-1
      covers: [R1]
      rows:
        - run: sh -c "printf a0"
          expect:
            out: a0
        - run: sh -c "printf a1"
          expect:
            out: a1
    - id: tool/bb
      name: The second
      prose: Prints twice more.
      claims:
        - src/b.txt:1-1
      covers: [R1]
      rows:
        - run: sh -c "printf b0"
          expect:
            out: b0
        - run: sh -c "printf b1"
          expect:
            out: b1
`)
	runIn(t, root)
	records, err := journalRecords(t, root)
	if err != nil {
		t.Fatal(err)
	}
	var keys []string
	for _, rec := range records {
		if rec.Kind != journal.KindRun || rec.Run == nil {
			continue
		}
		for _, fact := range rec.Run.Rows {
			keys = append(keys, fact.Key)
		}
	}
	want := []string{"tool/aa/0", "tool/aa/1", "tool/bb/0", "tool/bb/1"}
	if strings.Join(keys, ",") != strings.Join(want, ",") {
		t.Fatalf("the run facts must carry the corpus order:\nwant %v\ngot  %v", want, keys)
	}
}
