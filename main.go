package main

import (
	"log"
	"os"

	"github.com/kevin-hanselman/duc/src/cmd"
)

func main() {
	if os.Geteuid() == 0 {
		log.Fatal("refusing to run as root")
	}
	cmd.Execute()
}
