package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/kevin-hanselman/dud/src/artifact"
	"github.com/kevin-hanselman/dud/src/fsutil"
	"github.com/kevin-hanselman/dud/src/index"
	"github.com/kevin-hanselman/dud/src/stage"

	"github.com/spf13/cobra"
)

var genStageCmd = &cobra.Command{
	Use:   "gen [flags] [--] [stage_command]...",
	Short: "Generate stage YAML using the CLI",
	Long: `Gen generates stage YAML and prints it to standard output.

The output of this command can be redirected to a file and modified further as
needed.`,
	Example: `dud stage gen -o data/ python download_data.py > download.yaml`,
	Run: func(cmd *cobra.Command, args []string) {
		rootDir, err := getProjectRootDir()
		if err != nil {
			fatal(err)
		}
		stageWorkingDir, err := pathAbsThenRel(rootDir, stageWorkingDir)
		if err != nil {
			fatal(err)
		}
		stage := stage.Stage{
			WorkingDir: stageWorkingDir,
			Command:    strings.Join(args, " "),
		}
		stage.Outputs = make(map[string]*artifact.Artifact, len(stageOutputs))
		for _, path := range stageOutputs {
			art, err := createArtifactFromPath(rootDir, path)
			if err != nil {
				fatal(err)
			}
			stage.Outputs[art.Path] = art
		}
		stage.Dependencies = make(map[string]*artifact.Artifact, len(stageDependencies))
		for _, path := range stageDependencies {
			art, err := createArtifactFromPath(rootDir, path)
			if err != nil {
				fatal(err)
			}
			stage.Dependencies[art.Path] = art
		}
		if err := stage.Serialize(os.Stdout); err != nil {
			fatal(err)
		}
	},
}

var addStageCmd = &cobra.Command{
	Use:   "add stage_file...",
	Short: "Add one or more stage files to the index",
	Long: `Add adds one or more stage files to the index.

Add loads each stage file passed on the command line, validates its contents,
checks if it conflicts with any stages already in the index, then adds the
stage to the index file.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		rootDir, err := getProjectRootDir()
		if err != nil {
			fatal(err)
		}

		fullIndexPath := filepath.Join(rootDir, indexPath)
		idx, err := index.FromFile(fullIndexPath)
		if err != nil {
			fatal(err)
		}

		for _, path := range args {
			pathRelToRoot, err := pathAbsThenRel(rootDir, path)
			if err != nil {
				fatal(err)
			}
			stg, err := stage.FromFile(path)
			if err != nil {
				fatal(err)
			}
			if err := idx.AddStage(stg, pathRelToRoot); err != nil {
				fatal(err)
			}
			logger.Info.Printf("Added %s to the index.", path)
		}

		if err := idx.ToFile(fullIndexPath); err != nil {
			fatal(err)
		}
	},
}

var (
	stageOutputs, stageDependencies []string
	stageWorkingDir                 string
)

func init() {
	genStageCmd.Flags().StringSliceVarP(
		&stageOutputs,
		"out",
		"o",
		[]string{},
		"one or more output files or directories",
	)

	genStageCmd.Flags().StringSliceVarP(
		&stageDependencies,
		"dep",
		"d",
		[]string{},
		"one or more dependent files or directories",
	)

	genStageCmd.Flags().StringVarP(
		&stageWorkingDir,
		"work-dir",
		"w",
		"",
		"working directory for the stage's command",
	)

	stageCmd := &cobra.Command{
		Use:   "stage",
		Short: "Commands for interacting with stages and the index",
		Long:  "Stage is a group of sub-commands for interacting with stages and the index.",
	}
	stageCmd.AddCommand(genStageCmd)
	stageCmd.AddCommand(addStageCmd)
	rootCmd.AddCommand(stageCmd)
}

func createArtifactFromPath(rootDir, path string) (art *artifact.Artifact, err error) {
	cleanPath, err := pathAbsThenRel(rootDir, path)
	if err != nil {
		return
	}
	fileStatus, err := fsutil.FileStatusFromPath(cleanPath)
	if err != nil {
		return
	}
	art = &artifact.Artifact{
		Path:  cleanPath,
		IsDir: fileStatus == fsutil.StatusDirectory,
	}
	return
}
