package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolVarP(
		&runSingleStage,
		"single-stage",
		"s",
		false,
		"disable recursive operation on upstream stages",
	)
}

var runSingleStage bool

var runCmd = &cobra.Command{
	Use:   "run [flags] [stage_file]...",
	Short: "Run stages or pipelines",
	Long: `Run runs stages or pipelines.

For each stage passed in, run executes a stage's command if it is out-of-date.
If no stage files are passed in, run will act on all stages in the index. By
default, run will act recursively on all stages upstream of the given stage,
and thus run will execute a stage's command if any upstream stages are
out-of-date.`,
	Run: func(cmd *cobra.Command, paths []string) {
		rootDir, ch, idx, err := prepare(paths)
		if err != nil {
			fatal(err)
		}

		if len(idx) == 0 {
			fatal(emptyIndexError{})
		}

		if len(paths) == 0 {
			for path := range idx {
				paths = append(paths, path)
			}
		}

		ran := make(map[string]bool)
		for _, path := range paths {
			inProgress := make(map[string]bool)
			err := idx.Run(path, ch, rootDir, !runSingleStage, ran, inProgress, logger)
			if err != nil {
				fatal(err)
			}
		}
	},
}
