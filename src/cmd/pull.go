package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(pullCmd)
	pullCmd.Flags().BoolVarP(
		&useCopyStrategy,
		"copy",
		"c",
		false,
		"copy artifacts instead of linking",
	)
	pullCmd.Flags().BoolVarP(
		&disableRecursion,
		"single-stage",
		"s",
		false,
		"don't operate recursively over Stage inputs",
	)
}

var pullCmd = &cobra.Command{
	Use:   "pull [flags] [stage_file]...",
	Short: "Fetch artifacts from the remote and checkout",
	Long: `Pull runs fetch followed by checkout.

This command requires rclone to be installed on your machine. Visit
https://rclone.org/ for more information and installation instructions.`,
	Run: func(cmd *cobra.Command, args []string) {
		fetchCmd.Run(cmd, args)
		checkoutCmd.Run(cmd, args)
	},
}
