package main

import (
	"fmt"
	"os"

	"github.com/kevin-hanselman/dud/src/cmd"
)

// Version string set by goreleaser.
var version string = "NONE"

func main() {
	cmd.Version = version
	if os.Geteuid() == 0 {
		fmt.Printf(`WARNING: Running as root.
The root user does not respect read-only files. You can (and eventually will)
accidentally corrupt your Dud cache by overwriting an artifact linked to the
cache. Please consider running as a non-root user.

`)
	}
	cmd.Main()
}
