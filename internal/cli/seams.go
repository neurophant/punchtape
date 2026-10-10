// The CLI's world seams: the working directory and the pause of the
// open retries live behind package variables — a unit binds its own
// and the verbs run without the real world (the same discipline as
// the engine's clock; the files ride fsport).
package cli

import (
	"os"
	"time"
)

var (
	getwd = os.Getwd
	sleep = time.Sleep
)
