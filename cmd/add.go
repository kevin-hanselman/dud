package cmd

import (
	"errors"
	"github.com/kevlar1818/duc/fsutil"
	"github.com/kevlar1818/duc/index"
	"github.com/kevlar1818/duc/stage"
	"github.com/kevlar1818/duc/track"
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
	"path/filepath"
)

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().StringVarP(&ducfileFlag, "output", "o", "", "output path for Ducfile")
}

var ducfileFlag string
var toFile = fsutil.ToYamlFile
var fromFile = fsutil.FromYamlFile

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add files or stages to the index and commit list",
	Long:  "Add files or stages to the index and commit list",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		root, err := getRoot()
		if err != nil {
			log.Fatal(err)
		}
		indexPath := filepath.Join(root, ".duc", "index")

		// Create an index file if it doesn't exist
		if _, err := os.Stat(indexPath); os.IsNotExist(err) {
			_, err := os.Create(indexPath)
			if err != nil {
				log.Fatal(err)
			}
		}

		ducfilePath, err := cmd.Flags().GetString("output")
		if err != nil {
			log.Fatal(err)
		}
		if ducfilePath == "" {
			ducfilePath = "Ducfile"
		}

		idx := index.NewIndex()
		if err := fsutil.FromYamlFile(indexPath, idx); err != nil && err != io.EOF {
			log.Fatal(err)
		}

		if err := add(args, idx, ducfilePath); err != nil {
			log.Fatal(err)
		}

		if err := fsutil.ToYamlFile(indexPath, idx); err != nil {
			log.Fatal(err)
		}
	},
}

type addType int

const (
	stageType addType = iota
	fileType
)

func (addtype addType) String() string {
	return [...]string{"stageType", "fileType"}[addtype]
}

// add will add new files to a stage and the new stage to the index and the commit list,
// or will add existing stages to the commit list
func add(paths []string, idx *index.Index, ducfilePath string) error {

	pathTypes, err := checkAddTypes(paths)

	if err != nil {
		return nil
	}

	switch pathTypes {
	case stageType:
		if err := addStages(paths, idx); err != nil {
			return err
		}
	case fileType:
		if err := addFiles(paths, idx, ducfilePath); err != nil {
			return err
		}
	}

	return nil
}

func checkAddTypes(paths []string) (addType, error) {
	var firstPathType addType
	var err error

	for idx, path := range paths {
		if idx == 0 {
			firstPathType = checkAddType(path)
		} else if pathType := checkAddType(path); pathType != firstPathType {
			err = errors.New("cannot mix files and stages")
			return 0, err
		}
	}
	return firstPathType, nil
}

// for now, any error opening decoding yaml means it is not a stage
func checkAddType(path string) addType {
	stg := stage.Stage{}
	err := fromFile(path, &stg)

	if err != nil {
		return fileType
	}

	return stageType
}

func addFiles(files []string, idx *index.Index, stagePath string) error {

	_, err := os.Stat(stagePath)

	if !os.IsNotExist(err) {
		return errors.New("stage file already exists")
	}

	stg, err := track.Track(files...)
	if err != nil {
		return err
	}
	toFile(stagePath, stg)

	if err := idx.Add(stagePath); err != nil {
		return err
	}
	return nil
}

func addStages(stages []string, idx *index.Index) error {
	for _, stage := range stages {
		if err := idx.Add(stage); err != nil {
			return err
		}
	}

	return nil
}
