package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/kevin-hanselman/duc/cache"
	"github.com/kevin-hanselman/duc/index"
	"github.com/kevin-hanselman/duc/stage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

func printStageStatus(stagePath string, status stage.Status) error {
	// TODO: include entry.IsLocked and entry.ToCommit in output
	fmt.Println(stagePath)
	for path, artStatus := range status {
		// TODO: use test/tabwriter?
		if _, err := fmt.Printf("  %s  %s\n", path, artStatus); err != nil {
			return err
		}
	}
	return nil
}

var statusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"stat", "st"},
	Short:   "Print the status of one or more DUC stages.",
	Long:    "Print the status of one or more DUC stages.",
	Run: func(cmd *cobra.Command, args []string) {

		ch, err := cache.NewLocalCache(viper.GetString("cache"))
		if err != nil {
			log.Fatal(err)
		}

		indexPath, err := getIndexPath()
		if err != nil {
			log.Fatal(err)
		}

		rootDir, err := getProjectRootDir()
		if err != nil {
			log.Fatal(err)
		}

		idx, err := index.FromFile(indexPath)
		if os.IsNotExist(err) { // TODO: print error instead?
			idx = make(index.Index)
		} else if err != nil {
			log.Fatal(err)
		}

		if len(args) == 0 { // By default, check status of everything in the Index.
			for path := range idx {
				args = append(args, path)
			}
		}

		for _, path := range args {
			entry, ok := idx[path]
			if !ok {
				log.Fatal(fmt.Errorf("path %s not present in Index", path))
			}
			status, err := entry.Stage.Status(ch, false, rootDir)
			if err != nil {
				log.Fatal(err)
			}
			if err := printStageStatus(path, status); err != nil {
				log.Fatal(err)
			}
		}
	},
}
