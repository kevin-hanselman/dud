package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

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

		rootDir, err := getProjectRootDir()
		if err != nil {
			log.Fatal(err)
		}
		if err := os.Chdir(rootDir); err != nil {
			log.Fatal(err)
		}
		indexPath := filepath.Join(".duc", "index")

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

		checkedOut := make(map[string]bool)
		for _, path := range args {
			inProgress := make(map[string]bool)
			err := idx.Checkout(path, ch, strat, checkedOut, inProgress)
			if err != nil {
				log.Fatal(err)
			}
		}

		fmt.Println("checked out:")
		for stagePath := range checkedOut {
			fmt.Printf("  %s\n", stagePath)
		}
	},
}
