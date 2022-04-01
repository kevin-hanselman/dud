package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(fetchCmd)
	fetchCmd.Flags().BoolVarP(
		&disableRecursion,
		"single-stage",
		"s",
		false,
		"don't operate recursively over Stage inputs",
	)
}

type noRemoteError struct{}

func (e noRemoteError) Error() string {
	return "no remote specified in the config"
}

var fetchCmd = &cobra.Command{
	Use:   "fetch [flags] [stage_file]...",
	Short: "Fetch committed artifacts from the remote cache",
	Long: `Fetch downloads previously committed artifacts from a remote cache.

For each stage passed in, fetch downloads the stage's committed outputs from the
remote cache specified in the Dud config file. If no stage files are passed
in, fetch will act on all stages in the index. By default, fetch will act
recursively on all stages upstream of the given stage(s).

This command requires rclone to be installed on your machine. Visit
https://rclone.org/ for more information and installation instructions.`,
	Run: func(cmd *cobra.Command, paths []string) {
		rootDir, ch, idx, err := prepare(paths)
		if err != nil {
			fatal(err)
		}

		remote := viper.GetString("remote")
		if remote == "" {
			fatal(noRemoteError{})
		}

		if len(paths) == 0 {
			// Ignore disableRecursion flag when no args passed.
			disableRecursion = false
			for path := range idx {
				paths = append(paths, path)
			}
		}

		fetched := make(map[string]bool)
		for _, path := range paths {
			inProgress := make(map[string]bool)
			if err := idx.Fetch(
				path,
				ch,
				rootDir,
				!disableRecursion,
				remote,
				fetched,
				inProgress,
				logger,
			); err != nil {
				fatal(err)
			}
		}
	},
}
