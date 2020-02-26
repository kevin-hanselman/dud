package main

import (
    "os"
    "github.com/kevlar1818/duc/add"
    "log"
    "fmt"
)

func main() {
    for _, path := range os.Args[1:] {
        file, err := os.Open(path)
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
