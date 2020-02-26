package main

import (
	"fmt"
	"github.com/kevlar1818/duc/add"
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
		hash, err := add.Add(file)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s  %s\n", hash, path)
	}
}
