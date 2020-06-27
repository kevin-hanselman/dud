package cmd

import (
	"fmt"
	"github.com/kevlar1818/duc/cache"
	"github.com/kevlar1818/duc/fsutil"
	"github.com/kevlar1818/duc/index"
	"github.com/kevlar1818/duc/stage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

func printStageStatus(stagePath string, status stage.Status) error {
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

		idx := make(index.Index)
		if err := fsutil.FromYamlFile(indexPath, &idx); err != nil {
			log.Fatal(err)
		}

		if len(args) == 0 { // By default, check status of everything in the Index.
			for path := range idx {
				args = append(args, path)
			}
		}

		for _, path := range args {
			stg := new(stage.Stage)
			// TODO: We shouldn't error-out if a lockfile isn't found.
			// A missing lockfile is essentially another state a Stage can be
			// in. Should this be handled in Stage.Status()?
			if err := fsutil.FromYamlFile(stage.FilePathForLock(path), stg); err != nil {
				log.Fatal(err)
			}
			status, err := stg.Status(ch)
			if err != nil {
				log.Fatal(err)
			}
			if err := printStageStatus(path, status); err != nil {
				log.Fatal(err)
			}
		}
	},
}
