package cmd

import (
	"fmt"
	"os"

	"github.com/kevin-hanselman/dud/src/checksum"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(checksumCmd)
	checksumCmd.Flags().Int64VarP(&bufSize, "bufsize", "b", 0, "internal buffer size in bytes")
}

var bufSize int64

var checksumCmd = &cobra.Command{
	Use:   "checksum [flags] [file]...",
	Short: "Checksum files or bytes from STDIN",
	Long: `Checksum reads files (or bytes from STDIN) and prints their checksums.

The CLI is intended to be compatible with the *sum family of command-line tools
(although this version is currently incomplete).`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			checksum, err := checksum.Checksum(os.Stdin, bufSize)
			if err != nil {
				logger.Fatal(err)
			}
			fmt.Printf("%s  -\n", checksum)
			return
		}

		for _, path := range args {
			file, err := os.Open(path)
			if err != nil {
				logger.Fatal(err)
			}
			checksum, err := checksum.Checksum(file, bufSize)
			if err != nil {
				logger.Fatal(err)
			}
			fmt.Printf("%s  %s\n", checksum, path)
		}
	},
	// Override rootCmd's PersistentPreRun which changes dir to the project
	// root.
	PersistentPreRun: func(cmd *cobra.Command, args []string) {},
}
