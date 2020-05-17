package cmd

import (
	"fmt"
	"github.com/kevlar1818/duc/cache"
	"github.com/kevlar1818/duc/fsutil"
	"github.com/kevlar1818/duc/stage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

func printStageStatus(status stage.Status) error {
	for path, artStatusStr := range status {
		if _, err := fmt.Printf("%s  %s\n", path, artStatusStr); err != nil {
			return err
		}
	}
	return nil
}

var statusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"stat", "st"},
	Short:   "Print the status of one or more DUC stages.",
	Long:    "Print the status of one or more DUC stages.",
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 0 {
			args = append(args, "Ducfile")
		}

		ch, err := cache.NewLocalCache(viper.GetString("cache"))
		if err != nil {
			log.Fatal(err)
		}

		for _, path := range args {
			stg := new(stage.Stage)
			if err := fsutil.FromYamlFile(path, stg); err != nil {
				log.Fatal(err)
			}
			status, err := stg.Status(ch)
			if err != nil {
				log.Fatal(err)
			}
			if err := printStageStatus(status); err != nil {
				log.Fatal(err)
			}
		}
	},
}
