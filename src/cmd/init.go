package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the .dud folder",
	Long:  "Initialize the .dud folder",
	Run: func(cmd *cobra.Command, args []string) {
		if err := os.MkdirAll(filepath.Join(".dud", "cache"), 0755); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Initialized .dud folder")
	},
}
