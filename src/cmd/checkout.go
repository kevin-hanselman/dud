package cmd

import (
	"errors"

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

For each stage file passed in, checkout makes the stage's output artifacts
available in the workspace. By default, checkout creates symlinks to the cache,
but copies of the cached artifacts can be checked out using --copy. If no
stage files are passed in, checkout will act on all stages in the index. By
default, checkout will act recursively on all upstream stages (i.e.
dependencies).`,
	Run: func(cmd *cobra.Command, args []string) {
		strat := strategy.LinkStrategy
		if useCopyStrategy {
			strat = strategy.CopyStrategy
		}

		rootDir, paths, err := cdToProjectRootAndReadConfig(args)
		if err != nil {
			fatal(err)
		}

		ch, err := cache.NewLocalCache(viper.GetString("cache"))
		if err != nil {
			fatal(err)
		}

		idx, err := index.FromFile(indexPath)
		if err != nil {
			fatal(err)
		}

		if len(idx) == 0 {
			fatal(errors.New(emptyIndexMessage))
		}

		if len(paths) == 0 {
			// Ignore disableRecursion flag when no args passed.
			disableRecursion = false
			for path := range idx {
				paths = append(paths, path)
			}
		}

		checkedOut := make(map[string]bool)
		for _, path := range paths {
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
