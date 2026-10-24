// The entry point's battery: the mounted call answers the machine's
// own codes over a real zone, and the launch's argument law — the
// call, the version ask, the loud refusal.
package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunFirstContactAnswersZero(t *testing.T) {
	dir := t.TempDir()
	wd, _ := os.Getwd()
	defer os.Chdir(wd)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	if code := run(); code != 0 {
		t.Fatalf("the first contact answers zero, got %d", code)
	}
	if _, err := os.Stat(filepath.Join(dir, ".punchtape")); !os.IsNotExist(err) {
		t.Fatalf("the first contact creates nothing")
	}
}

func TestRunWithoutAWorkingDirectoryRefuses(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Skipf("the working directory is not readable: %v", err)
	}
	defer os.Chdir(wd)
	dir := t.TempDir()
	sub := filepath.Join(dir, "gone")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(sub); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(sub); err != nil {
		t.Fatal(err)
	}
	if code := run(); code != 1 {
		t.Fatalf("a vanished working directory must refuse, got %d", code)
	}
}

func TestLaunchIsTheVersionAsk(t *testing.T) {
	var out bytes.Buffer
	if code := launch([]string{"--version"}, &out); code != 0 {
		t.Fatalf("the version ask answers zero, got %d", code)
	}
	if !strings.HasPrefix(out.String(), "punchtape ") {
		t.Fatalf("the version ask prints the version, got %q", out.String())
	}
}

// TestLaunchRefusesArguments — the contract's own law, enforced: a
// call that carries arguments is refused with one line quoting the
// law, before any machine state is touched.
func TestLaunchRefusesArguments(t *testing.T) {
	for _, args := range [][]string{
		{"run"}, {"submit", "file"}, {"--version", "extra"}, {"-version"}, {""},
	} {
		var out bytes.Buffer
		if code := launch(args, &out); code != 1 {
			t.Fatalf("%v: a refused call answers one, got %d", args, code)
		}
		if !strings.Contains(out.String(), "the call takes no arguments") {
			t.Fatalf("%v: the refusal must quote the law, got %q", args, out.String())
		}
	}
}
