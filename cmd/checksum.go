package cmd

import (
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"github.com/kevlar1818/duc/checksum"
	"github.com/spf13/cobra"
	"log"
	"os"
)

func init() {
	rootCmd.AddCommand(checksumCmd)
	checksumCmd.Flags().Int64VarP(&bufSize, "bufsize", "b", 8000000, "buffer size")
}

var bufSize int64

var checksumCmd = &cobra.Command{
	Use:   "checksum",
	Short: "Checksum a file or bytes from STDIN",
	Long:  "Checksum a file or bytes from STDIN",
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 0 {
			checksum, err := checksum.Checksum(os.Stdin, bufSize)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%s  -\n", checksum)
			return
		}

		for _, path := range args {
			fileInfo, err := os.Stat(path)
			if err != nil {
				log.Fatal(err)
			}
			file, err := os.Open(path)
			if err != nil {
				log.Fatal(err)
			}
			bar := pb.Full.Start64(fileInfo.Size())
			barReader := bar.NewProxyReader(file)
			checksum, err := checksum.Checksum(barReader, bufSize)
			if err != nil {
				log.Fatal(err)
			}
			bar.Finish()
			fmt.Printf("%s  %s\n", checksum, path)
		}
	},
}
