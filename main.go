package main

import (
	"fmt"
	"os"

	"github.com/kevin-hanselman/dud/src/cmd"
)

// Version string set by goreleaser.
var version string = "NONE"

func main() {
	if os.Geteuid() == 0 {
		fmt.Println("refusing to run as root")
		os.Exit(1)
	}
	cmd.Main(version)
}
