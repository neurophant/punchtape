//go:build !linux && !darwin

package fsys

import "errors"

// Flock — the lock port's honest refusal where the platform carries
// no flock: the run cannot take the zone's one-pass line here.
func (OS) Flock(path string) (func() error, error) {
	return nil, errors.New("zone lock: no implementation on this host")
}
