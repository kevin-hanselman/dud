package cmd

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/kevlar1818/duc/fsutil"
)

func getRoot() (string, error) {

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
			return "", errors.New("no root")
		}
	}
}

func getIndexPath() (string, error) {
	root, err := getRoot()
	if err != nil {
		return "", err
	}

	return filepath.Join(root, ".duc", "index"), nil
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
