//go:build !linux && !darwin

package application

// kernelRelease — no uname on this host: the honest constant.
func kernelRelease() string { return "unknown" }
