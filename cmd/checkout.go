package cmd

import (
	"log"

	"github.com/kevlar1818/duc/cache"
	"github.com/kevlar1818/duc/fsutil"
	"github.com/kevlar1818/duc/stage"
	"github.com/kevlar1818/duc/strategy"
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

		if len(args) == 0 {
			args = append(args, "Ducfile")
		}

		for _, path := range args {
			stg := new(stage.Stage)
			if err := fsutil.FromYamlFile(path, stg); err != nil {
				log.Fatal(err)
			}
			if err := stg.Checkout(ch, strat); err != nil {
				log.Fatal(err)
			}
		}
	},
}
