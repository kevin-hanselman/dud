package cmd

import (
	"fmt"
	cachePkg "github.com/kevlar1818/duc/cache"
	"github.com/kevlar1818/duc/stage"
	strategyPkg "github.com/kevlar1818/duc/strategy"
	"github.com/spf13/cobra"
	"log"
)

func init() {
	rootCmd.AddCommand(commitCmd)
	commitCmd.Flags().StringVarP(&commitStrategy, "strategy", "s", "", "Strategy to use for commit. One of {link, copy}. Defaults to link.")
}

var commitStrategy string

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Compute the checksum value and move to cache",
	Long:  "Compute the checksum value and move to cache",
	Run: func(cmd *cobra.Command, args []string) {
		strategyFlag, err := cmd.Flags().GetString("strategy")
		if err != nil {
			log.Fatal(err)
		}

		var strategy strategyPkg.CheckoutStrategy

		if strategyFlag == "" || strategyFlag == "link" {
			strategy = strategyPkg.LinkStrategy
		} else if strategyFlag == "copy" {
			strategy = strategyPkg.CopyStrategy
		} else {
			log.Fatal(fmt.Errorf("invalid strategy specified: %s", strategyFlag))
		}

		// TODO load cache from config
		cache := cachePkg.LocalCache{Dir: "/tmp/.duc"}

		if len(args) == 0 {
			args = append(args, "Ducfile")
		}

		for _, arg := range args {
			stage, err := stage.FromFile(arg)
			if err != nil {
				log.Fatal(err)
			}
			if err := stage.Commit(&cache, strategy); err != nil {
				log.Fatal(err)
			}
			if err := stage.ToFile(arg); err != nil {
				log.Fatal(err)
			}
		}
	},
}
