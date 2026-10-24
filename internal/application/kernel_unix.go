//go:build linux

package application

import "syscall"

// kernelRelease — the host kernel's release line, as the kernel
// reports it: the parse of what came back, empty when the kernel
// answered nothing.
func kernelRelease() string {
	var u syscall.Utsname
	_ = syscall.Uname(&u)
	b := make([]byte, 0, len(u.Release))
	for _, c := range u.Release {
		if c == 0 {
			break
		}
		b = append(b, byte(c))
	}
	return string(b)
}
