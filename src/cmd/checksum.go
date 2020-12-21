package cmd

import (
	"fmt"
	"os"

	"github.com/kevin-hanselman/dud/src/checksum"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(checksumCmd)
	checksumCmd.Flags().IntVarP(&bufSize, "bufsize", "b", 0, "internal buffer size in bytes")
}

var bufSize int

var checksumCmd = &cobra.Command{
	Use:   "checksum [flags] [file]...",
	Short: "Checksum files or bytes from STDIN",
	Long: `Checksum reads files (or bytes from STDIN) and prints their checksums.

The CLI is intended to be compatible with the *sum family of command-line tools
(although this version is currently incomplete).`,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			buffer []byte = nil
			cksum  string
			err    error
		)
		if bufSize > 0 {
			buffer = make([]byte, bufSize)
		}
		if len(args) == 0 {
			if buffer == nil {
				cksum, err = checksum.Checksum(os.Stdin)
			} else {
				cksum, err = checksum.ChecksumBuffer(os.Stdin, buffer)
			}
			if err != nil {
				logger.Fatal(err)
			}
			fmt.Printf("%s  -\n", cksum)
			return
		}

		for _, path := range args {
			file, err := os.Open(path)
			if err != nil {
				logger.Fatal(err)
			}
			if buffer == nil {
				cksum, err = checksum.Checksum(file)
			} else {
				cksum, err = checksum.ChecksumBuffer(file, buffer)
			}
			if err != nil {
				logger.Fatal(err)
			}
			fmt.Printf("%s  %s\n", cksum, path)
		}
	},
}
