package fsutil

import (
	"fmt"
	"os"
	"syscall"
)

// FileStatus enumerates the states of a file on the filesystem.
type FileStatus int

const (
	// Absent means that the file does not exist.
	Absent FileStatus = iota
	// RegularFile means that the file exists as a regular file.
	RegularFile
	// Link means that the artifact exists as a link.
	Link
	// Directory means that the file exists as a directory.
	Directory
	// Other means none of the above.
	Other
)

func (fs FileStatus) String() string {
	return [...]string{"Absent", "RegularFile", "Link", "Directory", "Other"}[fs]
}

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

// FileStatusFromPath converts a path into a FileStatus enum value.
func FileStatusFromPath(path string) (FileStatus, error) {
	fileInfo, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Absent, nil
		}
		return 0, err
	}
	mode := fileInfo.Mode()

	if mode.IsRegular() {
		return RegularFile, nil
	}

	if mode.IsDir() {
		return Directory, nil
	}

	if (mode & os.ModeSymlink) != 0 {
		return Link, nil
	}

	return Other, nil
}

// SameFilesystem returns true if two files live on the same filesystem; it
// returns false otherwise. Follows links.
func SameFilesystem(pathA, pathB string) (bool, error) {
	devA, err := getFileDevice(pathA)
	if err != nil {
		return false, err
	}
	devB, err := getFileDevice(pathB)
	if err != nil {
		return false, err
	}
	return devA == devB, nil
}

func getFileDevice(path string) (uint64, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	sys := fileInfo.Sys()
	if sys == nil {
		return 0, fmt.Errorf("os.FileInfo.Sys() on %#v returned nil", path)
	}
	return sys.(*syscall.Stat_t).Dev, nil
}
