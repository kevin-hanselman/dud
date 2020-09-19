package cmd

import (
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/kevin-hanselman/duc/fsutil"
	"github.com/kevin-hanselman/duc/index"
	"github.com/kevin-hanselman/duc/stage"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().StringVarP(&outputStagePath, "output", "o", "", "output path for Ducfile")
	addCmd.Flags().BoolVarP(&isRecursive, "recursive", "r", false, "Recursively add directories. Defaults to false.")
}

var outputStagePath string
var isRecursive bool
var toFile = fsutil.ToYamlFile
var fromFile = fsutil.FromYamlFile

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add artifacts or stages to the index and commit list",
	Long:  "Add artifacts or stages to the index and commit list",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		rootDir, err := getProjectRootDir()
		if err != nil {
			log.Fatal(err)
		}
		if err := os.Chdir(rootDir); err != nil {
			log.Fatal(err)
		}
		indexPath := filepath.Join(rootDir, ".duc", "index")

		if outputStagePath == "" {
			outputStagePath = "Ducfile"
		}

		idx, err := index.FromFile(indexPath)
		if os.IsNotExist(err) {
			idx = make(index.Index)
		} else if err != nil {
			log.Fatal(err)
		}

		if err := add(args, &idx, outputStagePath, isRecursive); err != nil {
			log.Fatal(err)
		}

		if err := idx.ToFile(indexPath); err != nil {
			log.Fatal(err)
		}
	},
}

type addType int

const (
	stageType addType = iota
	artifactType
)

func (addtype addType) String() string {
	return [...]string{"stageType", "artifactType"}[addtype]
}

// add will add new artifacts to a stage and the new stage to the Index (and
// the commit list), or will add existing stages to the Index.
func add(paths []string, idx *index.Index, outputStagePath string, isRecursive bool) error {

	pathTypes, err := checkAddTypes(paths)

	if err != nil {
		return err
	}

	switch pathTypes {
	case stageType:
		if err := idx.AddStagesFromPaths(paths...); err != nil {
			return err
		}
	case artifactType:
		stg, err := stage.FromPaths(isRecursive, paths...)
		if err != nil {
			return err
		}
		if err := fsutil.ToYamlFile(outputStagePath, stg); err != nil {
			return err
		}
		if err := idx.AddStagesFromPaths(outputStagePath); err != nil {
			return err
		}
	}

	return nil
}

func checkAddTypes(paths []string) (addType, error) {
	var firstPathType addType

	for idx, path := range paths {
		if idx == 0 {
			firstPathType = checkAddType(path)
		} else if pathType := checkAddType(path); pathType != firstPathType {
			return 0, errors.New("cannot mix artifacts and stages")
		}
	}
	return firstPathType, nil
}

// for now, any error opening decoding yaml means it is not a stage
func checkAddType(path string) addType {
	stg := stage.Stage{}
	err := fromFile(path, &stg)

	if err != nil {
		return artifactType
	}

	return stageType
}
