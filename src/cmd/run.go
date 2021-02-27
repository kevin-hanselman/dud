package cmd

import (
	"os"

	"github.com/kevin-hanselman/dud/src/cache"
	"github.com/kevin-hanselman/dud/src/index"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolVarP(
		&runSingleStage,
		"single-stage",
		"s",
		false,
		"don't recursively operate on dependencies",
	)
}

var runSingleStage bool

var runCmd = &cobra.Command{
	Use:   "run [flags] [stage_file]...",
	Short: "Run stages or pipelines",
	Long: `Run runs stages or pipelines.

For each stage passed in, run executes a stage's command if it is out-of-date.
If no stage files are passed in, run will act on all stages in the index. By
default, run will act recursively on all upstream stages (i.e. dependencies),
and thus run will execute a stage's command if any upstream stages are
out-of-date.`,
	PreRun: cdToProjectRootAndReadConfig,
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

		if len(args) == 0 {
			for path := range idx {
				args = append(args, path)
			}
		}

		ran := make(map[string]bool)
		for _, path := range args {
			inProgress := make(map[string]bool)
			err := idx.Run(path, ch, rootDir, !runSingleStage, ran, inProgress, logger)
			if err != nil {
				logger.Fatal(err)
			}
		}
	},
}
