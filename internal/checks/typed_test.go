// The unit battery of the typed rows' execution:
// the repeat witness (agreement green, divergence a measured red
// naming both answers, the exit-code demand, the shared expectation),
// the staged-bytes probe witness (the probe rides the door, the
// machine stages and runs it), and the launch mode (the network-less
// namespace rides the launch itself — a command that needs the
// loopback fails under the fresh namespace and works without the
// mode).
package checks

import (
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/execport"
	"github.com/neurophant/punchtape/internal/fsport"
)

// repeatRow — a repeat-run row over a script.
func repeatRow(script string, out string) canon.Row {
	row := canon.Row{Repeat: &canon.RepeatRow{Cmd: canon.Cmd{Argv: []string{"sh", "-c", script}, IsArgv: true}}}
	if out != "" {
		row.Out = &canon.TextExpect{Text: out}
	}
	return row
}

// The repeat witness: the command runs twice over what it leaves.
func TestRepeatRunMeasures(t *testing.T) {
	dir := t.TempDir()
	res := Run(dir, "R/1", repeatRow("echo steady", "steady"), RunEnv{})
	if !res.Green {
		t.Fatalf("the agreeing repeat must be green: %+v", res)
	}
	if res.Exit == nil || *res.Exit != 0 || res.OutDigest == "" {
		t.Fatalf("the repeat's measured facts ride the result: %+v", res)
	}
	// both runs exit the demanded zero: rc disagreement is a red
	res = Run(dir, "R/2", repeatRow("exit 5", ""), RunEnv{})
	if res.Green || !strings.Contains(res.Reason, "exit code") {
		t.Fatalf("the nonzero repeat must red on the exit code: %+v", res)
	}
	// the two runs agree but not with the shared expectation
	res = Run(dir, "R/3", repeatRow("echo other", "wanted"), RunEnv{})
	if res.Green || !strings.Contains(res.Reason, "stdout") {
		t.Fatalf("the disagreeing expectation must red: %+v", res)
	}
	// the second run answers differently: the divergence is the
	// witness, quoting both answers (the leftover marker file is the
	// row's own business; the tree identity is the caller's gate)
	res = Run(dir, "R/4", repeatRow("if [ -f mark ]; then echo second; else echo first; touch mark; fi", ""), RunEnv{})
	if res.Green || !strings.Contains(res.Reason, "the two runs disagree") ||
		!strings.Contains(res.Reason, "first") || !strings.Contains(res.Reason, "second") {
		t.Fatalf("the diverging repeat must name both answers: %+v", res)
	}
	// a launch that cannot happen is an environment red, not behavior
	res = Run(dir, "R/5", canon.Row{Repeat: &canon.RepeatRow{
		Cmd: canon.Cmd{Argv: []string{"definitely-not-a-tool"}, IsArgv: true},
	}}, RunEnv{})
	if res.Green || !res.Env {
		t.Fatalf("the unlaunchable repeat must be an env red: %+v", res)
	}
}

// The launch mode: the network-less namespace is a launch
// discipline — the command's own bytes are untouched. A command that
// needs the loopback (a live listener on it) works without the mode
// and fails under the fresh network-less namespace by its own
// observable behavior.
func TestLaunchNonetIsALaunchFact(t *testing.T) {
	dir := t.TempDir()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("no loopback listener: %v", err)
	}
	defer ln.Close()
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			c.Close()
		}
	}()
	port := ln.Addr().(*net.TCPAddr).Port
	script := "exec 3<>/dev/tcp/127.0.0.1/" + itoaM10(port)
	row := canon.Row{
		Run:    canon.Cmd{Argv: []string{"bash", "-c", script}, IsArgv: true},
		Launch: canon.LaunchModeNonet,
	}
	res := Run(dir, "P/1", row, RunEnv{})
	if res.Green {
		t.Fatalf("the loopback connection must fail under the mode: %+v", res)
	}
	if res.Exit == nil || *res.Exit == 0 {
		t.Fatalf("the mode's launch must exit nonzero by its own behavior: %+v", res)
	}
	// without the mode the same command reaches the listener
	row.Launch = ""
	res = Run(dir, "P/2", row, RunEnv{})
	if !res.Green {
		t.Fatalf("the same command must work without the mode: %+v", res)
	}
}

// The staged-bytes probe: the probe's bytes ride the door (the inbox
// directory); the machine stages them into the tree for the row's
// duration, runs them and compares rc/stdout with the want. The
// staged file leaves the tree when the row ends; missing staged
// bytes are an honest env red.
func TestWitnessStagedBytes(t *testing.T) {
	dir := t.TempDir()
	inbox := filepath.Join(dir, "inbox")
	if err := os.MkdirAll(inbox, 0o755); err != nil {
		t.Fatal(err)
	}
	probe := "#!/bin/sh\necho probe-ok\n"
	if err := os.WriteFile(filepath.Join(inbox, "net-probe"), []byte(probe), 0o644); err != nil {
		t.Fatal(err)
	}
	row := canon.Row{
		Witness: &canon.WitnessRow{Probe: "file://net-probe"},
		Out:     &canon.TextExpect{Text: "probe-ok"},
	}
	res := Run(dir, "W/1", row, RunEnv{InboxDir: inbox})
	if !res.Green {
		t.Fatalf("the staged probe must run green: %+v", res)
	}
	// the staged bytes leave the tree when the row ends
	if _, err := os.Stat(filepath.Join(dir, "net-probe")); !os.IsNotExist(err) {
		t.Fatalf("the staged probe outlived its row: %v", err)
	}
	// rc composes: the probe's own refusal shape is a want, not a
	// surprise
	rc := 3
	if err := os.WriteFile(filepath.Join(inbox, "refuse-probe"), []byte("#!/bin/sh\necho no\nexit 3\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	row = canon.Row{
		Witness: &canon.WitnessRow{Probe: "file://refuse-probe"},
		Rc:      &rc,
		Out:     &canon.TextExpect{Text: "no"},
	}
	res = Run(dir, "W/2", row, RunEnv{InboxDir: inbox})
	if !res.Green {
		t.Fatalf("the refusing probe must measure green against its want: %+v", res)
	}
	// missing staged bytes: an env red naming the probe
	row = canon.Row{Witness: &canon.WitnessRow{Probe: "file://absent-probe"}}
	res = Run(dir, "W/3", row, RunEnv{InboxDir: inbox})
	if res.Green || !res.Env || !strings.Contains(res.Reason, "absent-probe") {
		t.Fatalf("the missing staged bytes must be an env red naming the probe: %+v", res)
	}
	// a tree file at the probe's name is restored byte for byte
	if err := os.WriteFile(filepath.Join(dir, "net-probe"), []byte("owner bytes"), 0o644); err != nil {
		t.Fatal(err)
	}
	row = canon.Row{Witness: &canon.WitnessRow{Probe: "file://net-probe"}}
	_ = Run(dir, "W/4", row, RunEnv{InboxDir: inbox})
	data, err := os.ReadFile(filepath.Join(dir, "net-probe"))
	if err != nil || string(data) != "owner bytes" {
		t.Fatalf("the tree file under the staged probe was not restored: %q %v", data, err)
	}
	// the staged file is read through the fs port like every byte
	// the machine touches
	_ = fsport.Bind
}

// The staged probe's honest fault shapes: a witness reference without
// the file:// form, a stage that cannot write, a probe that will not
// leave, and the two measured reds (rc, out).
func TestWitnessFaultShapes(t *testing.T) {
	dir := t.TempDir()
	inbox := filepath.Join(dir, "inbox")
	if err := os.MkdirAll(inbox, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inbox, "net-probe"), []byte("#!/bin/sh\ntrue\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// a witness reference without the door's form
	res := Run(dir, "W/E1", canon.Row{Witness: &canon.WitnessRow{Probe: "net-probe"}}, RunEnv{InboxDir: inbox})
	if res.Green || !res.Env || !strings.Contains(res.Reason, "file://name form") {
		t.Fatalf("the formless witness reference must be an env red: %+v", res)
	}
	// the stage cannot even make the directory
	bindFaults(t, faultFS{mkdirErr: errIO2}, faultExec{procFn: func([]string) (execport.Process, error) {
		return faultProc{out: execport.Outcome{Exit: 0}}, nil
	}})
	res = Run(dir, "W/E2", canon.Row{Witness: &canon.WitnessRow{Probe: "file://net-probe"}}, RunEnv{InboxDir: inbox})
	if res.Green || !res.Env || !strings.Contains(res.Reason, "did not stage") {
		t.Fatalf("the unstaged probe must be an env red: %+v", res)
	}
	// the stage cannot write the bytes
	bindFaults(t, faultFS{writeErr: errIO2}, faultExec{procFn: func([]string) (execport.Process, error) {
		return faultProc{out: execport.Outcome{Exit: 0}}, nil
	}})
	res = Run(dir, "W/E3", canon.Row{Witness: &canon.WitnessRow{Probe: "file://net-probe"}}, RunEnv{InboxDir: inbox})
	if res.Green || !res.Env || !strings.Contains(res.Reason, "did not stage") {
		t.Fatalf("the unwritable probe stage must be an env red: %+v", res)
	}
	// the staged probe will not leave the tree when the row ends
	bindFaults(t, faultFS{removeErr: errIO2}, faultExec{procFn: func([]string) (execport.Process, error) {
		return faultProc{out: execport.Outcome{Exit: 0}}, nil
	}})
	res = Run(dir, "W/E4", canon.Row{Witness: &canon.WitnessRow{Probe: "file://net-probe"}}, RunEnv{InboxDir: inbox})
	if res.Green || !res.Env || !strings.Contains(res.Reason, "did not leave the tree") {
		t.Fatalf("the leftover probe must be an env red: %+v", res)
	}
}

// probeOverwriteFS — the write port that lets the FIRST write (the
// stage) through and fails the SECOND (the restore of a tree file the
// probe's name collided with): the leftover branch of the restore.
type probeOverwriteFS struct {
	faultFS
	writes int
}

func (f *probeOverwriteFS) WriteFile(p string, data []byte, m os.FileMode) error {
	f.writes++
	if f.writes >= 2 {
		return errIO2
	}
	return f.faultFS.WriteFile(p, data, m)
}

// A tree file under the probe's name is restored byte for byte; a
// restore that cannot write is an honest env red.
func TestWitnessRestoreFault(t *testing.T) {
	dir := t.TempDir()
	inbox := filepath.Join(dir, "inbox")
	if err := os.MkdirAll(inbox, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "net-probe"), []byte("owner bytes"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inbox, "net-probe"), []byte("#!/bin/sh\ntrue\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	fs := &probeOverwriteFS{}
	bindFaults(t, fs, faultExec{procFn: func([]string) (execport.Process, error) {
		return faultProc{out: execport.Outcome{Exit: 0}}, nil
	}})
	res := Run(dir, "W/E5", canon.Row{Witness: &canon.WitnessRow{Probe: "file://net-probe"}}, RunEnv{InboxDir: inbox})
	if res.Green || !res.Env || !strings.Contains(res.Reason, "did not leave the tree") {
		t.Fatalf("the unrestorable tree file must be an env red: %+v", res)
	}
}

// The staged probe's measured reds: the exit code and the stdout are
// wants like any row's.
func TestWitnessMeasuredReds(t *testing.T) {
	dir := t.TempDir()
	inbox := filepath.Join(dir, "inbox")
	if err := os.MkdirAll(inbox, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inbox, "rc-probe"), []byte("#!/bin/sh\nexit 7\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	res := Run(dir, "W/R1", canon.Row{Witness: &canon.WitnessRow{Probe: "file://rc-probe"}}, RunEnv{InboxDir: inbox})
	if res.Green || !strings.Contains(res.Reason, "exit code 7") {
		t.Fatalf("the probe's own code must red against the want: %+v", res)
	}
	if err := os.WriteFile(filepath.Join(inbox, "out-probe"), []byte("#!/bin/sh\necho other\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	res = Run(dir, "W/R2", canon.Row{
		Witness: &canon.WitnessRow{Probe: "file://out-probe"},
		Out:     &canon.TextExpect{Text: "wanted"},
	}, RunEnv{InboxDir: inbox})
	if res.Green || !strings.Contains(res.Reason, "stdout") {
		t.Fatalf("the probe's own answer must red against the want: %+v", res)
	}
}

// itoaM10 — the decimal spelling of a port.
func itoaM10(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

// The repeat's second launch dies at the port: an environment red,
// not a behavior divergence (the fault port scripts the second call).
func TestRepeatSecondLaunchFault(t *testing.T) {
	dir := t.TempDir()
	calls := 0
	bindFaults(t, faultFS{}, faultExec{procFn: func([]string) (execport.Process, error) {
		calls++
		if calls >= 2 {
			return nil, errIO2
		}
		return faultProc{out: execport.Outcome{Exit: 0, Stdout: "x\n"}}, nil
	}})
	res := Run(dir, "R/6", repeatRow("script", ""), RunEnv{})
	if res.Green || !res.Env {
		t.Fatalf("the faulted second launch must be an env red: %+v", res)
	}
}
