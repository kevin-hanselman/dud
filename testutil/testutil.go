package testutil

import (
	"io/ioutil"
)

// CreateTempDirs creates a DUC cache and workspace in the OS temp FS.
func CreateTempDirs() (cacheDir, workDir string, err error) {
	cacheDir, err = ioutil.TempDir("", "duc_cache")
	if err != nil {
		return "", "", err
	}
	workDir, err = ioutil.TempDir("", "duc_wspace")
	if err != nil {
		return "", "", err
	}
	return
}
