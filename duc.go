package main

import (
	"fmt"
	"github.com/kevlar1818/duc/commit"
	"log"
	"os"
)

func main() {
	for _, path := range os.Args[1:] {
		file, err := os.Open(path)
		defer file.Close()
		if err != nil {
			log.Fatal(err)
		}
		hash, err := commit.Commit(file, nil)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s  %s\n", hash, path)
	}
}
