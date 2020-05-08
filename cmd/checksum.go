package cmd

import (
	"fmt"
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
			args = append(args, "/dev/stdin")
		}

		for _, path := range args {
			//info, err := os.Stat(path)
			//if err != nil {
			//	log.Fatal(err)
			//}
			f, err := os.Open(path)
			if err != nil {
				log.Fatal(err)
			}
			checksum, err := checksum.Checksum(f, bufSize)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%s  %s\n", checksum, path)
		}
	},
}
