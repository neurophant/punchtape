//go:build darwin

package application

import "syscall"

// kernelRelease — the host kernel's release line, as the kernel
// reports it: empty when the kernel answered nothing.
func kernelRelease() string {
	s, err := syscall.Sysctl("kern.osrelease")
	if err != nil {
		return ""
	}
	return s
}
