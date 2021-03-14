package cmd

import (
	"github.com/kevin-hanselman/dud/src/cache"
	"github.com/kevin-hanselman/dud/src/index"
	"github.com/kevin-hanselman/dud/src/strategy"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(checkoutCmd)
	checkoutCmd.Flags().BoolVarP(
		&useCopyStrategy,
		"copy",
		"c",
		false,
		"copy artifacts instead of linking",
	)
	checkoutCmd.Flags().BoolVarP(
		&disableRecursion,
		"single-stage",
		"s",
		false,
		"don't recursively operate on dependencies",
	)
}

var useCopyStrategy, disableRecursion bool

var checkoutCmd = &cobra.Command{
	Use:   "checkout [flags] [stage_file]...",
	Short: "Load committed artifacts from the cache",
	Long: `Checkout loads previously committed artifacts from the cache.

For each stage file passed in, checkout attempts to load all output artifacts
from the cache for the given. If no stage files are passed in, checkout will act on all
stages in the index. By default, checkout will act recursively on all upstream stages
(i.e. dependencies).`,
	PreRun: cdToProjectRootAndReadConfig,
	Run: func(cmd *cobra.Command, args []string) {
		strat := strategy.LinkStrategy
		if useCopyStrategy {
			strat = strategy.CopyStrategy
		}

		ch, err := cache.NewLocalCache(viper.GetString("cache"))
		if err != nil {
			fatal(err)
		}

		idx, err := index.FromFile(indexPath)
		if err != nil {
			fatal(err)
		}

		if len(args) == 0 {
			// Ignore disableRecursion flag when no args passed.
			disableRecursion = false
			for path := range idx {
				args = append(args, path)
			}
		}

		checkedOut := make(map[string]bool)
		for _, path := range args {
			inProgress := make(map[string]bool)
			if err := idx.Checkout(
				path,
				ch,
				rootDir,
				strat,
				!disableRecursion,
				checkedOut,
				inProgress,
				logger,
			); err != nil {
				fatal(err)
			}
		}
	},
}
