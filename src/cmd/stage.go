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

var stageCmd = &cobra.Command{
	Use:   "stage",
	Short: "Commands for interacting with stages and the index",
	Long: `Stage is a group of commands for interacting with stages and the index.

A Stage is a group of Artifacts, or an operation that consumes and/or produces
a group of Artifacts. Stages are defined by the user in YAML files and should be
tracked with source control.

Below is a fully-annotated Stage YAML file for reference.

` + "``` yaml" + `
# The checksum of this Stage definition, written during 'dud commit'. This
# checksum is used to determine when a Stage definition has been modified by
# a user. The checksum does not include Artifact checksums.
checksum: abcdefghijklmnopqrstuvwxyz1234567890

# The shell command to run when 'dud run' is called. '/bin/sh' is used to
# run the command. (Stage commands are optional.)
command: python train.py

# The directory in which the Stage's command is executed. Like all paths in
# a Stage definition, it must be a directory path relative to the project's root
# directory. An empty or omitted value means the command is executed in the
# project root directory. The working directory only affects the Stage's command;
# all inputs and outputs of the Stage must also have paths relative to the
# project root.
working-dir: .

# The set of Artifacts which the Stage requires to run 'command' above.
inputs:
  # The Artifact path. All paths are relative to the project's root
  # directory.
  train.py:
    # The checksum of the artifact's contents, written during 'dud commit'.
    checksum: abcdefghijklmnopqrstuvwxyz1234567890

# The set of Artifacts which are owned by the Stage.
outputs:
  # This is how to define a file Artifact with default options. The colon (:)
  # at the end is still required. You may see 'dud stage gen' include empty curly
  # braces ({}) after the colon; this is equivalent to the below.
  model.pkl:
  tensorboard:
    # 'is-dir' tells Dud to expect this Artifact to be a directory, not a file.
    # Defaults to false when omitted.
    is-dir: true

    # 'disable-recursion' tells Dud not to recurse directories when operating on
    # this directory Artifact; all sub-directories will be ignored. Defaults to
    # false (applies recursion) when omitted. Not applicable for file
    # Artifacts.
    disable-recursion: true

  metrics.json:
    # 'skip-cache' tells Dud not to commit this Artifact to the cache. Dud will
    # still write a checksum for this Artifact during 'dud commit', and it will
    # use the checksum to inform other actions, such as 'dud run'. This is useful
    # for declaring Stage outputs which can be safely stored in source control
    # rather than Dud. This option is implicit for Artifacts in 'inputs'.
    skip-cache: true
` + "```",
}

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
		stg := stage.Stage{
			WorkingDir: stageWorkingDir,
			Command:    strings.Join(args, " "),
		}
		stg.Outputs = make(map[string]*artifact.Artifact, len(stageOutputs))
		for _, path := range stageOutputs {
			art, err := createArtifactFromPath(rootDir, path)
			if err != nil {
				fatal(err)
			}
			stg.Outputs[art.Path] = art
		}
		stg.Inputs = make(map[string]*artifact.Artifact, len(stageInputs))
		for _, path := range stageInputs {
			art, err := createArtifactFromPath(rootDir, path)
			if err != nil {
				fatal(err)
			}
			stg.Inputs[art.Path] = art
		}
		if err := stg.Validate(); err != nil {
			fatal(err)
		}
		if err := stg.Serialize(os.Stdout); err != nil {
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
	stageOutputs, stageInputs []string
	stageWorkingDir           string
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
		&stageInputs,
		"in",
		"i",
		[]string{},
		"one or more input files or directories",
	)

	genStageCmd.Flags().StringVarP(
		&stageWorkingDir,
		"work-dir",
		"w",
		"",
		"working directory for the stage's command",
	)

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
