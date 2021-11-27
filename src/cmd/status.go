package cmd

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/kevin-hanselman/dud/src/index"
	"github.com/kevin-hanselman/dud/src/stage"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

func writeStageStatus(writer io.Writer, stagePath string, status stage.Status) error {
	var stageFileStatus string
	if status.ChecksumMatches {
		stageFileStatus = "up-to-date"
	} else if status.HasChecksum {
		stageFileStatus = "modified"
	} else {
		stageFileStatus = "not checksummed"
	}
	fmt.Fprintf(writer, "%s\tstage definition %s\n", stagePath, stageFileStatus)
	for path, artStatus := range status.ArtifactStatus {
		fmt.Fprintf(writer, "  %s\t%s\n", path, artStatus)
	}
	return nil
}

var statusCmd = &cobra.Command{
	Use:     "status [flags] [stage_file]...",
	Aliases: []string{"stat", "st"},
	Short:   "Print the state of one or more stages",
	Long: `Status prints the state of one or more stages.

For each stage file passed in, status will print the current state of the
stage. If no stage files are passed in, status will act on all stages in the
index. By default, status will act recursively on all stages upstream of the
given stage(s).`,
	Run: func(cmd *cobra.Command, paths []string) {
		rootDir, ch, idx, err := prepare(paths...)
		if err != nil {
			fatal(err)
		}

		if len(idx) == 0 {
			fatal(emptyIndexError{})
		}

		if len(paths) == 0 { // By default, check status of everything in the Index.
			for path := range idx {
				paths = append(paths, path)
			}
		}

		status := make(index.Status)
		for _, path := range paths {
			inProgress := make(map[string]bool)
			err := idx.Status(path, ch, rootDir, status, inProgress)
			if err != nil {
				fatal(err)
			}
		}

		writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		for _, path := range idx.SortStagePaths() {
			if err := writeStageStatus(writer, path, status[path]); err != nil {
				fatal(err)
			}
			fmt.Fprintln(writer)
		}
		writer.Flush()
	},
}
