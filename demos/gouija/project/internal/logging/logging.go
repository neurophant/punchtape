// Package logging provides conditional logging for Gouija.
// When quiet mode is enabled, informational logs are suppressed
// while errors are still logged.
package logging

import (
	"io"
	"log"
	"os"
)

var (
	// quiet indicates whether quiet mode is enabled
	quiet bool

	// infoLogger logs informational messages (suppressed in quiet mode)
	infoLogger = log.New(os.Stdout, "", log.LstdFlags)

	// errLogger always logs errors
	errLogger = log.New(os.Stderr, "", log.LstdFlags)
)

// SetQuiet enables or disables quiet mode.
// When quiet mode is enabled, Info() calls are suppressed.
func SetQuiet(q bool) {
	quiet = q
	if q {
		infoLogger.SetOutput(io.Discard)
	} else {
		infoLogger.SetOutput(os.Stdout)
	}
}

// IsQuiet returns whether quiet mode is enabled.
func IsQuiet() bool {
	return quiet
}

// Info logs an informational message.
// Suppressed when quiet mode is enabled.
func Info(format string, v ...interface{}) {
	infoLogger.Printf(format, v...)
}

// Error logs an error message.
// Never suppressed, always logs.
func Error(format string, v ...interface{}) {
	errLogger.Printf(format, v...)
}

// Debug logs a debug message.
// Suppressed when quiet mode is enabled (same as Info).
func Debug(format string, v ...interface{}) {
	infoLogger.Printf(format, v...)
}
