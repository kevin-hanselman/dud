package cmd

import (
	"fmt"
	"os"

	"github.com/kevin-hanselman/dud/src/cache"
	"github.com/kevin-hanselman/dud/src/index"
	"github.com/kevin-hanselman/dud/src/stage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

func printStageStatus(stagePath string, status stage.Status) error {
	// TODO: use text/tabwriter?
	fmt.Println(stagePath)
	for path, artStatus := range status {
		if _, err := fmt.Printf("  %s  %s\n", path, artStatus); err != nil {
			return err
		}
	}
	return nil
}

var statusCmd = &cobra.Command{
	Use:     "status [flags] [stage_file]...",
	Aliases: []string{"stat", "st"},
	Short:   "Print the state of one or more stages",
	Long: `Status prints the state of one or more stages.

For each stage file passed in, status will print the current state of the
stage.  If no stage files are passed in, status will act on all stages in the
index. By default, status will act recursively on all upstream stages (i.e.
dependencies).`,
	PreRun: requireInitializedProject,
	Run: func(cmd *cobra.Command, args []string) {

		ch, err := cache.NewLocalCache(viper.GetString("cache"))
		if err != nil {
			logger.Fatal(err)
		}

		idx, err := index.FromFile(".dud/index")
		if os.IsNotExist(err) {
			idx = make(index.Index)
		} else if err != nil {
			logger.Fatal(err)
		}

		if len(args) == 0 { // By default, check status of everything in the Index.
			for path := range idx {
				args = append(args, path)
			}
		}

		status := make(index.Status)
		for _, path := range args {
			inProgress := make(map[string]bool)
			err := idx.Status(path, ch, rootDir, status, inProgress)
			if err != nil {
				logger.Fatal(err)
			}
		}

		for path, stageStatus := range status {
			if err := printStageStatus(path, stageStatus); err != nil {
				logger.Fatal(err)
			}
		}
	},
}
