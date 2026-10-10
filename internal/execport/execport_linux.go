//go:build linux

package execport

import "syscall"

// netnsSysProcAttr — the standing disciplines (the group) plus the
// denial: a fresh user+network namespace, the caller's uid/gid mapped
// 1:1 so the tree sees its own owner.
func netnsSysProcAttr() (*syscall.SysProcAttr, error) {
	return &syscall.SysProcAttr{
		Setpgid:     true,
		Cloneflags:  syscall.CLONE_NEWUSER | syscall.CLONE_NEWNET,
		UidMappings: []syscall.SysProcIDMap{{ContainerID: 0, HostID: syscall.Getuid(), Size: 1}},
		GidMappings: []syscall.SysProcIDMap{{ContainerID: 0, HostID: syscall.Getgid(), Size: 1}},
	}, nil
}
