package cmd

import (
	"fmt"
	"log"

	"github.com/kevin-hanselman/duc/cache"
	"github.com/kevin-hanselman/duc/index"
	"github.com/kevin-hanselman/duc/strategy"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(checkoutCmd)
	checkoutCmd.Flags().BoolVarP(&useCopyStrategy, "copy", "c", false, "Copy the file instead of linking.")
}

var useCopyStrategy bool

var checkoutCmd = &cobra.Command{
	Use:   "checkout",
	Short: "checkout all artifacts from the cache",
	Long:  "checkout all artifacts from the cache",
	Run: func(cmd *cobra.Command, args []string) {

		strat := strategy.LinkStrategy
		if useCopyStrategy {
			strat = strategy.CopyStrategy
		}

		ch, err := cache.NewLocalCache(viper.GetString("cache"))
		if err != nil {
			log.Fatal(err)
		}

		indexPath, err := getIndexPath()
		if err != nil {
			log.Fatal(err)
		}

		rootDir, err := getProjectRootDir()
		if err != nil {
			log.Fatal(err)
		}

		// TODO: forcing a checkout will require a "force load lock"
		// flag in index.FromFile
		idx, err := index.FromFile(indexPath)
		if err != nil {
			log.Fatal(err)
		}

		// By default, checkout everything in the Index.
		if len(args) == 0 {
			for path := range idx {
				args = append(args, path)
			}
		}

		// TODO: first, check all entry.IsLocked and error out if any are false
		for _, path := range args {
			entry, ok := idx[path]
			if !ok {
				log.Fatal(fmt.Errorf("path %s not present in Index", path))
			}
			if err := entry.Stage.Checkout(ch, strat, rootDir); err != nil {
				log.Fatal(err)
			}
		}
	},
}
