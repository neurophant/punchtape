// The execution battery: the row's argv spelling runs for real, and
// a wait that could not be answered is a red row carrying the
// failure's own words — never a silent green.
package application

import (
	"strings"
	"testing"
	"time"

	"github.com/neurophant/punchtape/internal/domain/zone"
	"github.com/neurophant/punchtape/internal/infrastructure/fsys"
	"github.com/neurophant/punchtape/internal/ports"
)

func TestArgvListRowRuns(t *testing.T) {
	mountPorts()
	root := newRoot(t)
	writeZone(t, root, zone.InboxSlot, `behaviors:
    - id: tool/argv
      name: The argv
      prose: Runs a list.
      rows:
        - run: ["sh", "-c", "printf hi > out.txt"]
          expect:
            files:
                out.txt: hi
`)
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
	writeZone(t, root, zone.InboxSlot, `behaviors:
    - id: tool/wait
      name: The waiter
      prose: Waits.
      rows:
        - run: "true"
          expect:
            rc: 0
`)
	runIn(t, root)
	report := zoneFile(t, root, zone.ReportFile)
	if !strings.Contains(report, "launch: wait: the process could not be reaped") {
		t.Fatalf("a wait failure must be a named red row:\n%s", report)
	}
}
