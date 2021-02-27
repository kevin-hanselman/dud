package cmd

import (
	"github.com/kevin-hanselman/dud/src/cache"
	"github.com/kevin-hanselman/dud/src/index"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	pushCmd.Flags().BoolVarP(
		&disableRecursion,
		"single-stage",
		"s",
		false,
		"don't recursively operate on dependencies",
	)
	rootCmd.AddCommand(pushCmd)
}

var pushCmd = &cobra.Command{
	Use:   "push [flags] [stage_file]...",
	Short: "Push committed artifacts to the remote cache",
	Long: `Push uploads previously committed artifacts to a remote cache.

For each stage passed in, push uploads the stage's committed outputs to the
remote cache specified in the Dud config file. If no stage files are passed
in, push will act on all stages in the index. By default, push will act
recursively on all upstream stages (i.e. dependencies).`,
	PreRun: cdToProjectRootAndReadConfig,
	Run: func(cmd *cobra.Command, args []string) {

		ch, err := cache.NewLocalCache(viper.GetString("cache"))
		if err != nil {
			logger.Fatal(err)
		}

		remote := viper.GetString("remote")
		if remote == "" {
			logger.Fatal("no remote specified in the config")
		}

		// TODO: Always read stage lock files?
		idx, err := index.FromFile(".dud/index")
		if err != nil {
			logger.Fatal(err)
		}

		if len(args) == 0 {
			// Ignore disableRecursion flag when no args passed.
			disableRecursion = false
			for path := range idx {
				args = append(args, path)
			}
		}

		pushed := make(map[string]bool)
		for _, path := range args {
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
				logger.Fatal(err)
			}
		}
	},
}
