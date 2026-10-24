//go:build !linux && !darwin

package process

import (
	"errors"

	"github.com/neurophant/punchtape/internal/ports"
)

// OS — the process port's honest refusal where the platform carries
// no implementation: launches answer one named error, nothing
// pretends to run.
type OS struct{}

// StartOpts — the unsupported-platform refusal.
func (OS) StartOpts(argv []string, dir string, opts ports.LaunchOpts) (ports.Process, error) {
	return nil, errors.New("process port: no implementation on this host")
}
