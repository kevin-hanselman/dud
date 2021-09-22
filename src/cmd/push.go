package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	pushCmd.Flags().BoolVarP(
		&disableRecursion,
		"single-stage",
		"s",
		false,
		"disable recursive operation on upstream stages",
	)
	rootCmd.AddCommand(pushCmd)
}

var pushCmd = &cobra.Command{
	Use:   "push [flags] [stage_file]...",
	Short: "Push committed artifacts to the remote cache",
	Long: `Push uploads previously committed artifacts to a remote cache.

For each stage passed in, push uploads the stage's committed outputs to the
remote cache specified in the Dud config file. If no stage files are passed
in, push will act on all stages in the index. By default, push will
act recursively on all stages upstream of the given stage(s).

This command requires rclone to be installed on your machine. Visit
https://rclone.org/ for more information and installation instructions.`,
	Run: func(cmd *cobra.Command, paths []string) {
		rootDir, ch, idx, err := prepare(paths...)
		if err != nil {
			fatal(err)
		}

		remote := viper.GetString("remote")
		if remote == "" {
			fatal(noRemoteError{})
		}

		if len(idx) == 0 {
			fatal(emptyIndexError{})
		}

		if len(paths) == 0 {
			// Ignore disableRecursion flag when no args passed.
			disableRecursion = false
			for path := range idx {
				paths = append(paths, path)
			}
		}

		pushed := make(map[string]bool)
		for _, path := range paths {
			inProgress := make(map[string]bool)
			if err := idx.Push(
				path,
				ch,
				rootDir,
				!disableRecursion,
				remote,
				pushed,
				inProgress,
				logger,
			); err != nil {
				fatal(err)
			}
		}
	},
}
