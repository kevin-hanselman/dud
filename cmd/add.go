package cmd

import (
	"github.com/go-yaml/yaml"
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
)

func init() {
	rootCmd.AddCommand(addCmd)
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add Ducfiles to the index",
	Long:  "Add Ducfiles to the index",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// TODO get the index path regardless of sub directory location
		path := ".duc/index"

		// TODO make index unique
		index := []string{}

		indexFile, err := os.Open(path)
		if err != nil {
			log.Fatal(err)
		}

		if err = yaml.NewDecoder(indexFile).Decode(&index); err != nil && err != io.EOF {
			log.Fatal(err)
		}

		// TODO validate ducfile before adding to index
		for _, arg := range args {
			index = append(index, arg)
		}

		indexFile, err = os.Create(path)
		if err != nil {
			log.Fatal(err)
		}

		if err = yaml.NewEncoder(indexFile).Encode(&index); err != nil {
			log.Fatal(err)
		}
	},
}
