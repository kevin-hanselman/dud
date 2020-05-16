package cmd

import (
	"fmt"
	"github.com/kevlar1818/duc/cache"
	"github.com/kevlar1818/duc/fsutil"
	"github.com/kevlar1818/duc/index"
	"github.com/kevlar1818/duc/stage"
	"github.com/kevlar1818/duc/strategy"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		strategyFlag, err := cmd.Flags().GetString("strategy")
		if err != nil {
			log.Fatal(err)
		}

		var strat strategy.CheckoutStrategy

		if strategyFlag == "" || strategyFlag == "link" {
			strat = strategy.LinkStrategy
		} else if strategyFlag == "copy" {
			strat = strategy.CopyStrategy
		} else {
			log.Fatal(fmt.Errorf("invalid strategy specified: %s", strategyFlag))
		}

		ch, err := cache.NewLocalCache(viper.GetString("cache"))
		if err != nil {
			log.Fatal(err)
		}

		indexPath, err := getIndexPath()
		if err != nil {
			log.Fatal(err)
		}

		idx := new(index.Index)
		if err := fsutil.FromYamlFile(indexPath, idx); err != nil {
			log.Fatal(err)
		}

		for commitStage := range idx.CommitSet() {
			stg := new(stage.Stage)
			if err := fsutil.FromYamlFile(commitStage, stg); err != nil {
				log.Fatal(err)
			}
			// TODO: add recursive flag
			if err := stg.Commit(ch, strat); err != nil {
				log.Fatal(err)
			}
			if err := fsutil.ToYamlFile(commitStage, stg); err != nil {
				log.Fatal(err)
			}
		}

		idx.ClearCommitSet()

		if err := fsutil.ToYamlFile(indexPath, idx); err != nil {
			log.Fatal(err)
		}
	},
}
