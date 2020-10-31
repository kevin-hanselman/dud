package cmd

import (
	"github.com/kevin-hanselman/dud/src/cache"
	"github.com/kevin-hanselman/dud/src/index"
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
		"don't recursively operate on dependencies",
	)
}

var fetchCmd = &cobra.Command{
	Use:   "fetch [flags] [stage_file]...",
	Short: "Fetch committed artifacts to the default remote cache",
	Long: `Fetch downloads previously committed artifacts from a remote cache.

For each stage passed in, fetch downloads the stage's committed outputs from the
remote cache specified in the Dud config file. If no stage files are passed
in, fetch will act on all stages in the index. By default, fetch will act
recursively on all upstream stages (i.e. dependencies).`,
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

		fetched := make(map[string]bool)
		for _, path := range args {
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
				logger.Fatal(err)
			}
		}
	},
}
