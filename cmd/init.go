package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the .duc folder",
	Long:  "Initialize the .duc folder",
	Run: func(cmd *cobra.Command, args []string) {
		if err := os.Mkdir(".duc", 0755); err != nil {
			log.Fatal(err)
		}
		file, err := os.Create(".duc/index")
		if err != nil {
			log.Fatal(err)
		}
		file.Close()

		fmt.Println("Initialized .duc folder")
	},
}
