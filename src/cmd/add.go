package cmd

import (
	"log"
	"os"

	"github.com/kevin-hanselman/dud/src/index"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(addCmd)
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add stages to the index",
	Long:  "Add stages to the index",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		indexPath := ".dud/index"

		idx, err := index.FromFile(indexPath)
		if os.IsNotExist(err) {
			idx = make(index.Index)
		} else if err != nil {
			log.Fatal(err)
		}

		for _, path := range args {
			if err := idx.AddStageFromPath(path); err != nil {
				log.Fatal(errors.Wrap(err, "add"))
			}
		}

		if err := idx.ToFile(indexPath); err != nil {
			log.Fatal(errors.Wrap(err, "add"))
		}
	},
}
