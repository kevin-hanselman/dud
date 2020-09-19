package cmd

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/kevin-hanselman/duc/fsutil"
)

func getProjectRootDir() (string, error) {

	dirname, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		ducFolderExists, err := hasDucFolder(dirname)
		if err != nil {
			return "", err
		}

		if ducFolderExists {
			return dirname, nil
		}

		dirname = filepath.Dir(dirname)

		if dirname == "/" {
			return "", errors.New("no project root directory found")
		}
	}
}

func hasDucFolder(dir string) (bool, error) {
	exists, err := fsutil.Exists(filepath.Join(dir, ".duc"), false)

	if err != nil {
		return false, err
	}

	if exists {
		return true, nil
	}
	return false, nil
}
