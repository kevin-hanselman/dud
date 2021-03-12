package cmd

import (
	"os"

	"github.com/kevin-hanselman/dud/src/cache"
	"github.com/kevin-hanselman/dud/src/index"
	"github.com/kevin-hanselman/dud/src/strategy"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(commitCmd)
	commitCmd.Flags().BoolVarP(
		&useCopyStrategy, // defined in cmd/checkout.go
		"copy",
		"c",
		false,
		"On checkout, copy the file instead of linking.",
	)
}

var commitCmd = &cobra.Command{
	Use:   "commit [flags] [stage_file]...",
	Short: "Save artifacts to the cache and record their checksums",
	Long: `Commit saves artifacts to the cache and record their checksums.

For each stage file passed in, commit saves all output artifacts in the cache
and records their checksums in the stage file. If no stage files are passed
in, commit will act on all stages in the index. By default, commit will act
recursively on all upstream stages (i.e. dependencies).`,
	PreRun: cdToProjectRootAndReadConfig,
	Run: func(cmd *cobra.Command, args []string) {
		strat := strategy.LinkStrategy
		if useCopyStrategy {
			strat = strategy.CopyStrategy
		}

		ch, err := cache.NewLocalCache(viper.GetString("cache"))
		if err != nil {
			logger.Fatal(err)
		}

		idx, err := index.FromFile(indexPath)
		if err != nil {
			logger.Fatal(err)
		}

		if len(args) == 0 { // By default, commit all Stages.
			for path := range idx {
				args = append(args, path)
			}
		}

		committed := make(map[string]bool)
		for _, path := range args {
			inProgress := make(map[string]bool)
			err := idx.Commit(path, ch, rootDir, strat, committed, inProgress, logger)
			if err != nil {
				logger.Fatal(err)
			}
			stageFile, err := os.Create(path)
			if err != nil {
				logger.Fatal(err)
			}
			defer stageFile.Close()
			if err := idx[path].Serialize(stageFile); err != nil {
				logger.Fatal(err)
			}
		}

		if err := idx.ToFile(indexPath); err != nil {
			logger.Fatal(err)
		}
	},
}
