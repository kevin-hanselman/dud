package cmd

import (
	"log"
	"os"
	"path/filepath"

	"github.com/kevin-hanselman/duc/cache"
	"github.com/kevin-hanselman/duc/index"
	"github.com/kevin-hanselman/duc/stage"
	"github.com/kevin-hanselman/duc/strategy"
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

		rootDir, err := getProjectRootDir()
		if err != nil {
			log.Fatal(err)
		}
		if err := os.Chdir(rootDir); err != nil {
			log.Fatal(err)
		}
		indexPath := filepath.Join(".duc", "index")

		idx, err := index.FromFile(indexPath)
		if err != nil {
			log.Fatal(err)
		}

		if len(args) == 0 { // By default, commit all Stages.
			for path := range idx {
				args = append(args, path)
			}
		}

		committed := make(map[string]bool)
		for _, path := range args {
			inProgress := make(map[string]bool)
			err := idx.Commit(path, ch, strat, committed, inProgress)
			if err != nil {
				log.Fatal(err)
			}
			lockPath := stage.FilePathForLock(path)
			if idx[path].Stage.ToFile(lockPath); err != nil {
				log.Fatal(err)
			}
		}

		log.Println("committed:")
		for stagePath := range committed {
			log.Printf("  %s\n", stagePath)
		}

		if err := idx.ToFile(indexPath); err != nil {
			log.Fatal(err)
		}
	},
}
