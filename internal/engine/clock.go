package engine

import "time"

// The clock seam: Go cannot intercept time calls from a test, so the
// engine reads its clock through these bindings. Production never
// touches them; unit tests substitute a deterministic clock.
var (
	timeNow   = time.Now
	timeSince = func(t time.Time) time.Duration { return time.Since(t) }
	sleep     = time.Sleep
)
