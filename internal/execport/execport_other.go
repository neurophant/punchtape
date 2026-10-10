//go:build !linux

package execport

import (
	"errors"
	"syscall"
)

// netnsSysProcAttr — this OS has no unprivileged user/network
// namespaces to raise: the honest start failure LaunchOpts.Netns
// promises for a host that forbids the namespace.
func netnsSysProcAttr() (*syscall.SysProcAttr, error) {
	return nil, errors.New("launch discipline netns: this OS has no user/network namespaces")
}
