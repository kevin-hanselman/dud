package main

import (
	"fmt"
	"os"

	"github.com/kevin-hanselman/dud/src/cmd"
)

var (
	// Version string set by goreleaser.
	version string = "unset"
)

func main() {
	if os.Geteuid() == 0 {
		fmt.Println("refusing to run as root")
		os.Exit(1)
	}
	cmd.Main(version)
}
