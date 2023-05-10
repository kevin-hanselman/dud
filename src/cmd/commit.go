package cmd

import (
	"github.com/kevin-hanselman/dud/src/strategy"
	"github.com/spf13/cobra"
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
	Run: func(cmd *cobra.Command, paths []string) {
		strat := strategy.LinkStrategy
		if useCopyStrategy {
			strat = strategy.CopyStrategy
		}

		rootDir, ch, idx, err := prepare(paths)
		if err != nil {
			fatal(err)
		}

		if len(paths) == 0 { // By default, commit all Stages.
			for path := range idx {
				paths = append(paths, path)
			}
		}

		if len(paths) == 0 {
			fatal(emptyIndexError{})
		}

		committed := make(map[string]bool)
		written := make(map[string]bool)
		for _, path := range paths {
			inProgress := make(map[string]bool)
			err := idx.Commit(path, ch, rootDir, strat, committed, inProgress, logger)
			if err != nil {
				fatal(err)
			}
			for path := range committed {
				if written[path] {
					continue
				}
				if err := idx[path].ToFile(path); err != nil {
					fatal(err)
				}
				written[path] = true
			}
			logger.Info.Println()
		}
	},
}
