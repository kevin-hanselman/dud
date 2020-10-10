package cmd

import (
	"log"
	"os"
	"path/filepath"

	"github.com/kevin-hanselman/duc/src/index"
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
		rootDir, err := getProjectRootDir()
		if err != nil {
			log.Fatal(err)
		}
		if err := os.Chdir(rootDir); err != nil {
			log.Fatal(err)
		}
		indexPath := filepath.Join(rootDir, ".duc", "index")

		idx, err := index.FromFile(indexPath)
		if os.IsNotExist(err) {
			idx = make(index.Index)
		} else if err != nil {
			log.Fatal(err)
		}

		if err := idx.AddStagesFromPaths(args...); err != nil {
			log.Fatal(err)
		}

		if err := idx.ToFile(indexPath); err != nil {
			log.Fatal(err)
		}
	},
}
