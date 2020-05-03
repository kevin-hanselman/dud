package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "duc",
	Short: "DUC is a tool for storing, version, and reproducing big data files",
	Long: `Data Under Control (duc) is a tool to store, version, and reproduce big
	data files alongside the source code that creates it.
	Inspired by Git.`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

// Execute is the main entry point to the cobra cli
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
