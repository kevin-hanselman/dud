package cmd

import (
	"fmt"
	cachePkg "github.com/kevlar1818/duc/cache"
	"github.com/kevlar1818/duc/stage"
	strategyPkg "github.com/kevlar1818/duc/strategy"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
)

func init() {
	rootCmd.AddCommand(checkoutCmd)
	checkoutCmd.Flags().StringVarP(&checkoutStrategy, "strategy", "s", "", "Strategy to use for checkout. One of {link, copy}. Defaults to link.")
}

var checkoutStrategy string

var checkoutCmd = &cobra.Command{
	Use:   "checkout",
	Short: "checkout all artifacts from the cache",
	Long:  "checkout all artifacts from the cache",
	Run: func(cmd *cobra.Command, args []string) {
		strategyFlag, err := cmd.Flags().GetString("strategy")
		if err != nil {
			log.Fatal(err)
		}

		// TODO DRY
		var strategy strategyPkg.CheckoutStrategy

		if strategyFlag == "" || strategyFlag == "link" {
			strategy = strategyPkg.LinkStrategy
		} else if strategyFlag == "copy" {
			strategy = strategyPkg.CopyStrategy
		} else {
			log.Fatal(fmt.Errorf("invalid strategy specified: %s", strategyFlag))
		}

		cache := cachePkg.LocalCache{Dir: viper.GetString("cache")}

		if len(args) == 0 {
			args = append(args, "Ducfile")
		}

		for _, path := range args {
			stg := new(stage.Stage)
			if err := stg.FromFile(path); err != nil {
				log.Fatal(err)
			}
			if err := stg.Checkout(&cache, strategy); err != nil {
				log.Fatal(err)
			}
		}
	},
}
