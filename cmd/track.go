package cmd

import (
	"github.com/kevlar1818/duc/fsutil"
	"github.com/kevlar1818/duc/track"
	"github.com/spf13/cobra"
	"log"
)

func init() {
	rootCmd.AddCommand(trackCmd)
	trackCmd.Flags().StringVarP(&ducfileFlag, "output", "o", "", "output path for Ducfile")
}

var ducfileFlag string

var trackCmd = &cobra.Command{
	Use:   "track",
	Short: "Create a Ducfile showing targets to be added",
	Long:  "Create a Ducfile that shows which targets need to be added without computing checks",
	Run: func(cmd *cobra.Command, args []string) {
		stg, _ := track.Track(args...)

		ducfileFlag, err := cmd.Flags().GetString("output")
		if err != nil {
			log.Fatal(err)
		}

		if ducfileFlag == "" {
			ducfileFlag = "Ducfile"
		}

		fsutil.ToYamlFile(ducfileFlag, stg)
	},
}
