package cmd

import (
	"errors"
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
recursively on all stages upstream of the given stage(s).`,
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

		if len(paths) == 0 { // By default, commit all Stages.
			for path := range idx {
				paths = append(paths, path)
			}
		}

		if len(paths) == 0 {
			fatal(errors.New(emptyIndexMessage))
		}

		committed := make(map[string]bool)
		for _, path := range paths {
			inProgress := make(map[string]bool)
			err := idx.Commit(path, ch, rootDir, strat, committed, inProgress, logger)
			if err != nil {
				fatal(err)
			}
			stageFile, err := os.Create(path)
			if err != nil {
				fatal(err)
			}
			defer stageFile.Close()
			if err := idx[path].Serialize(stageFile); err != nil {
				fatal(err)
			}
		}

		if err := idx.ToFile(indexPath); err != nil {
			fatal(err)
		}
	},
}
