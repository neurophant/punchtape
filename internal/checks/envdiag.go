// Environment failure classification: deterministic recognition of
// the "the machine could not execute an external command" class —
// the command is not installed, no permission. The diagnosis and
// FIX-hint texts are canon data (envdiag.yaml); the classification
// is mechanics. A non-zero exit with the tool's output does not
// count as the environment: the tool's output is itself the
// diagnosis; the catalog does not duplicate it.
package checks

import (
	"errors"
	"io/fs"
	"os/exec"
)

// EnvClass — the environment failure class by execution error: a
// catalog identifier or an empty string (not the environment).
func EnvClass(err error) string {
	if err == nil {
		return ""
	}
	var execErr *exec.Error
	if errors.As(err, &execErr) {
		if errors.Is(execErr.Err, exec.ErrNotFound) || errors.Is(execErr.Err, fs.ErrNotExist) {
			return "exec-not-found"
		}
		if errors.Is(execErr.Err, fs.ErrPermission) {
			return "exec-permission"
		}
	}
	if errors.Is(err, exec.ErrNotFound) || errors.Is(err, fs.ErrNotExist) {
		return "exec-not-found"
	}
	if errors.Is(err, fs.ErrPermission) {
		return "exec-permission"
	}
	return ""
}
