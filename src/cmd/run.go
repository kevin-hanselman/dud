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
	Use:   "run",
	Short: "Run the given Stage and all upstream Stages.",
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

		rootDir, err := os.Getwd()
		if err != nil {
			logger.Fatal(err)
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
