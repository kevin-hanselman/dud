package cmd

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/kevin-hanselman/dud/src/fsutil"
)

func getProjectRootDir() (string, error) {

	dirname, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		dudFolderExists, err := hasDudFolder(dirname)
		if err != nil {
			return "", err
		}

		if dudFolderExists {
			return dirname, nil
		}

		dirname = filepath.Dir(dirname)

		if dirname == "/" {
			return "", errors.New("no project root directory found")
		}
	}
}

func hasDudFolder(dir string) (bool, error) {
	exists, err := fsutil.Exists(filepath.Join(dir, ".dud"), false)

	if err != nil {
		return false, err
	}

	if exists {
		return true, nil
	}
	return false, nil
}
