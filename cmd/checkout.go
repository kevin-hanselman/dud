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

		for _, path := range args {
			entry, ok := idx[path]
			// TODO: --force flag will need to always load the locked version
			// of the stage.
			if !ok {
				log.Fatal(fmt.Errorf("path %s not present in Index", path))
			}
			if err := entry.Stage.Checkout(ch, strat); err != nil {
				log.Fatal(err)
			}
		}
	},
}
