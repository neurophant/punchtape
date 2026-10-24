package application

// The declared clean execution environment (the law the canon
// prints): the PATH and HOME of the shell that runs the machine,
// plus the zone's own pin file .punchtape/env when it carries one —
// the same set in every session for every caller. The environment's
// identity — the machine's version, the host it stands on, the set
// actually passed — becomes the environment hash inside every
// proof: a session that changes what the commands see moves the
// proofs, rows go stale, never falsely green. The pins are the
// tree's shared declaration: strict KEY=VALUE, never an absolute
// path — the file travels with the zone, so nothing in it may
// point at one machine's disk; a line of any other shape is a loud
// refusal naming the file, the line and the law, never a silent
// loss.

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"github.com/neurophant/punchtape/internal/application/canon"
	"github.com/neurophant/punchtape/internal/domain/fprint"
	"github.com/neurophant/punchtape/internal/domain/zone"
	"github.com/neurophant/punchtape/internal/ports"
)

// pinKey — a pin's name: a plain environment identifier, nothing
// else is a key.
var pinKey = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// CleanEnv — the base environment every command of the machine runs
// in: exactly the names the canon declares, taken from the shell
// that runs the machine, nothing else.
func CleanEnv() []string {
	names := canon.Load().Env
	out := make([]string, 0, len(names))
	for _, name := range names {
		out = append(out, name+"="+os.Getenv(name))
	}
	return out
}

// ZoneEnv — the zone's own pin lines (the .punchtape/env file):
// one KEY=VALUE a line, the key [A-Za-z_][A-Za-z0-9_]*, no export,
// no quotes, no absolute paths. Comments and blank lines pass. A
// line of any other shape is an error quoting the file, the line
// number, the line itself and this law — the machine never drops
// half an environment in silence.
func ZoneEnv(zoneDir string) ([]string, error) {
	data, err := ports.FSBind.ReadFile(filepath.Join(zoneDir, zone.EnvSlot))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []string
	for n, line := range strings.Split(string(data), "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok || !pinKey.MatchString(key) {
			return nil, fmt.Errorf(
				"env pin refused — %s, line %d: %q — the law: one KEY=VALUE a line, "+
					"the key [A-Za-z_][A-Za-z0-9_]*, no export, no quotes, no absolute paths; "+
					"rows run from the tree root with the PATH and HOME of the shell that "+
					"runs the machine plus these pins",
				zone.Path(zone.EnvSlot), n+1, line)
		}
		if strings.HasPrefix(value, "/") {
			return nil, fmt.Errorf(
				"env pin refused — %s, line %d: %q — the law: no absolute paths; "+
					"a pin travels with the tree, so its paths are relative to the tree root",
				zone.Path(zone.EnvSlot), n+1, line)
		}
		if len(value) >= 2 && (value[0] == '"' && value[len(value)-1] == '"' ||
			value[0] == '\'' && value[len(value)-1] == '\'') {
			return nil, fmt.Errorf(
				"env pin refused — %s, line %d: %q — the law: no quotes; "+
					"the value is the line after the first =, taken as written",
				zone.Path(zone.EnvSlot), n+1, line)
		}
		out = append(out, key+"="+value)
	}
	return out, nil
}

// RunEnv — what a command of the machine actually sees: the clean
// base plus the zone's pins, a pin winning on a name.
func RunEnv(pins []string) []string {
	return append(CleanEnv(), pins...)
}

// EnvHash — the fingerprint of the machine's version, the host and
// the environment actually passed, the zone's pins included.
func EnvHash(pins []string) string {
	env := RunEnv(pins)
	sort.Strings(env)
	return fprint.EnvHash(fprint.EnvIdentity{
		Version: Version,
		Arch:    runtime.GOOS + "/" + runtime.GOARCH,
		Kernel:  kernelRelease(),
		Env:     env,
	})
}
