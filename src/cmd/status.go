package cmd

import (
	"errors"

	"github.com/kevin-hanselman/dud/src/cache"
	"github.com/kevin-hanselman/dud/src/index"
	"github.com/kevin-hanselman/dud/src/stage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

func printStageStatus(stagePath string, status stage.Status) error {
	// TODO: use text/tabwriter?
	var stageFileStatus string
	if status.ChecksumMatches {
		stageFileStatus = "up-to-date"
	} else if status.HasChecksum {
		stageFileStatus = "modified"
	} else {
		stageFileStatus = "not checksummed"
	}
	logger.Info.Printf("%-22s  stage definition %s\n", stagePath, stageFileStatus)
	for path, artStatus := range status.ArtifactStatus {
		logger.Info.Printf("  %-20s  %s\n", path, artStatus)
	}
	return nil
}

var statusCmd = &cobra.Command{
	Use:     "status [flags] [stage_file]...",
	Aliases: []string{"stat", "st"},
	Short:   "Print the state of one or more stages",
	Long: `Status prints the state of one or more stages.

For each stage file passed in, status will print the current state of the
stage.  If no stage files are passed in, status will act on all stages in the
index. By default, status will act recursively on all upstream stages (i.e.
dependencies).`,
	Run: func(cmd *cobra.Command, args []string) {
		rootDir, paths, err := cdToProjectRootAndReadConfig(args)
		if err != nil {
			fatal(err)
		}

		ch, err := cache.NewLocalCache(viper.GetString("cache"))
		if err != nil {
			fatal(err)
		}

		idx, err := index.FromFile(indexPath)
		if err != nil {
			fatal(err)
		}

		if len(idx) == 0 {
			fatal(errors.New(emptyIndexMessage))
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

		for path, stageStatus := range status {
			if err := printStageStatus(path, stageStatus); err != nil {
				fatal(err)
			}
		}
	},
}
