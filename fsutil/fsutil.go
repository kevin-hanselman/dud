package fsutil

import (
	"github.com/kevlar1818/duc/artifact"
	"os"
)

// Exists returns true if path is an existing file or directory, otherwise it
// returns false. If followLinks is true, then Exists will attempt to follow
// links to their target and report said target's existence. If followLinks is
// false, Exist will operate on the link itself.
func Exists(path string, followLinks bool) (bool, error) {
	var err error
	if followLinks {
		_, err = os.Stat(path)
	} else {
		_, err = os.Lstat(path)
	}
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// IsLink returns true if path represents a symlink, otherwise it returns false.
func IsLink(path string) (bool, error) {
	fileInfo, err := os.Lstat(path)
	if err != nil {
		return false, err
	}
	return fileInfo.Mode()&os.ModeSymlink != 0, nil
}

// IsRegularFile returns true if path represents a regular file, otherwise it returns false.
func IsRegularFile(path string) (bool, error) {
	// If we use os.Stat, it'll follow links, which we don't want.
	fileInfo, err := os.Lstat(path)
	if err != nil {
		return false, err
	}
	return fileInfo.Mode().IsRegular(), nil
}

// FileStatusFromPath converts a path into a artifact.FileStatus enum value.
func FileStatusFromPath(path string) (artifact.FileStatus, error) {
	fileInfo, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return artifact.Absent, nil
		}
		return 0, err
	}
	mode := fileInfo.Mode()

	if mode.IsRegular() {
		return artifact.RegularFile, nil
	}

	if mode.IsDir() {
		return artifact.Directory, nil
	}

	if (mode & os.ModeSymlink) != 0 {
		return artifact.Link, nil
	}

	return artifact.Other, nil
}
