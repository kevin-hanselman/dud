package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/kevin-hanselman/dud/src/index"
	"github.com/kevin-hanselman/dud/src/stage"
	"github.com/spf13/cobra"
)

func init() {
	statusCmd.Flags().BoolVar(&debugStatus, "debug", false, "print verbose JSON instead of regular output")
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

var (
	debugStatus bool

	statusCmd = &cobra.Command{
		Use:     "status [flags] [stage_file]...",
		Aliases: []string{"stat", "st"},
		Short:   "Print the state of one or more stages",
		Long: `Status prints the state of one or more stages.

For each stage file passed in, status will print the current state of the
stage. If no stage files are passed in, status will act on all stages in the
index. By default, status will act recursively on all stages upstream of the
given stage(s).`,
		Run: func(_ *cobra.Command, paths []string) {
			rootDir, ch, idx, err := prepare(paths)
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

			sort.Strings(paths)

			indexStatus := make(index.Status)
			for _, path := range paths {
				inProgress := make(map[string]bool)
				err := idx.Status(path, ch, rootDir, indexStatus, inProgress)
				if err != nil {
					fatal(err)
				}
			}

			if debugStatus {
				encoder := json.NewEncoder(os.Stdout)
				if err := encoder.Encode(indexStatus); err != nil {
					fatal(err)
				}
				return
			}

			writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			for path, stageStatus := range indexStatus {
				if err := writeStageStatus(writer, path, stageStatus); err != nil {
					fatal(err)
				}
				fmt.Fprintln(writer)
			}
			writer.Flush()
		},
	}
)
