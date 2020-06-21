package cmd

import (
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
	commitCmd.Flags().BoolVarP(
		&useCopyStrategy, // defined in cmd/checkout.go
		"copy",
		"c",
		false,
		"On checkout, copy the file instead of linking.",
	)
}

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Compute the checksum value and move to cache",
	Long:  "Compute the checksum value and move to cache",
	Args:  cobra.NoArgs,
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

		idx := make(index.Index)
		if err := fsutil.FromYamlFile(indexPath, &idx); err != nil {
			log.Fatal(err)
		}

		for stagePath := range idx.CommitSet() {
			stg := new(stage.Stage)
			if err := fsutil.FromYamlFile(stagePath, stg); err != nil {
				log.Fatal(err)
			}
			if err := stg.Commit(ch, strat); err != nil {
				log.Fatal(err)
			}
			if err := fsutil.ToYamlFile(stage.FilePathForLock(stagePath), stg); err != nil {
				log.Fatal(err)
			}
		}

		idx.ClearCommitSet()

		if err := fsutil.ToYamlFile(indexPath, idx); err != nil {
			log.Fatal(err)
		}
	},
}
