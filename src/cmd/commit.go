package cmd

import (
	"github.com/kevin-hanselman/dud/src/cache"
	"github.com/kevin-hanselman/dud/src/index"
	"github.com/kevin-hanselman/dud/src/stage"
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
	Use:   "commit",
	Short: "Compute the checksum value and move to cache",
	Long:  "Compute the checksum value and move to cache",
	Run: func(cmd *cobra.Command, args []string) {

		strat := strategy.LinkStrategy
		if useCopyStrategy {
			strat = strategy.CopyStrategy
		}

		ch, err := cache.NewLocalCache(viper.GetString("cache"))
		if err != nil {
			logger.Fatal(err)
		}

		indexPath := ".dud/index"

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
			err := idx.Commit(path, ch, strat, committed, inProgress, logger)
			if err != nil {
				logger.Fatal(err)
			}
			lockPath := stage.FilePathForLock(path)
			if idx[path].Stage.ToFile(lockPath); err != nil {
				logger.Fatal(err)
			}
		}

		if err := idx.ToFile(indexPath); err != nil {
			logger.Fatal(err)
		}
	},
}
