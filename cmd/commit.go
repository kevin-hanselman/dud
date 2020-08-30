package cmd

import (
	"log"

	"github.com/kevlar1818/duc/cache"
	"github.com/kevlar1818/duc/fsutil"
	"github.com/kevlar1818/duc/index"
	"github.com/kevlar1818/duc/stage"
	"github.com/kevlar1818/duc/strategy"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

		idx, err := index.FromFile(indexPath)
		if err != nil {
			log.Fatal(err)
		}

		for stagePath, entry := range idx {
			if !entry.ToCommit {
				continue
			}
			if err := entry.Stage.Commit(ch, strat); err != nil {
				log.Fatal(err)
			}
			lockPath := stage.FilePathForLock(stagePath)
			if err := fsutil.ToYamlFile(lockPath, entry.Stage); err != nil {
				log.Fatal(err)
			}
			entry.ToCommit = false
		}

		if err := idx.ToFile(indexPath); err != nil {
			log.Fatal(err)
		}
	},
}
