// Command tapsum is a small, finished utility: tape-style checksums
// for files. The CLI prints its version line and its usage page
// byte-exact; the watch loop logs a wall clock on every pass, so
// its bytes are pinned by no one yet.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/neurophant/tapsum/internal/sum"
)

// versionLine is the whole version contract: one line, one newline.
const versionLine = "tapsum version 1.1.0\n"

// main dispatches the printed contract and the watch loop.
func main() {
	if len(os.Args) < 2 {
		fmt.Print(usagePage)
		return
	}
	switch os.Args[1] {
	case "version":
		fmt.Print(versionLine)
	case "help":
		fmt.Print(usagePage)
	case "watch":
		watch(os.Args[2:])
	default:
		fmt.Fprint(os.Stderr, usagePage)
		os.Exit(2)
	}
}

// usagePage is the whole usage contract: the page, byte for byte.
const usagePage = `tapsum - tape-style checksums for files

Usage:
  tapsum <command> [arguments]

Commands:
  version        Print the version line and exit
  help           Print this usage page and exit
  watch <file>   Re-hash a file forever, one line per pass

Examples:
  tapsum version
  tapsum watch server.log
`

// watch re-hashes a file once a second, logging every pass with a
// wall clock, until Ctrl-C stops the tape.
func watch(args []string) {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "tapsum watch: exactly one file")
		os.Exit(2)
	}
	path := args[0]
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	tick := time.NewTicker(time.Second)
	defer tick.Stop()
	fmt.Fprintf(os.Stderr, "tapsum watch: %s - Ctrl-C stops the tape\n", path)
	for {
		select {
		case <-stop:
			fmt.Fprintln(os.Stderr, "tapsum watch: stopped")
			return
		case <-tick.C:
			h, err := sum.File(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "tapsum watch: %v\n", err)
				continue
			}
			fmt.Printf("tapsum watch %s %s %s\n", time.Now().UTC().Format(time.RFC3339), h, path)
		}
	}
}
