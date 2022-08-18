package fsutil

import (
	"encoding/json"
	"os"
)

// FileStatus enumerates the states of a file on the filesystem.
type FileStatus int

const (
	// StatusAbsent means that the file does not exist.
	StatusAbsent FileStatus = iota
	// StatusRegularFile means that the file exists as a regular file.
	StatusRegularFile
	// StatusLink means that the artifact exists as a link.
	StatusLink
	// StatusDirectory means that the file exists as a directory.
	StatusDirectory
	// StatusOther means none of the above.
	StatusOther
)

func (fs FileStatus) String() string {
	return [...]string{"absent", "regular file", "link", "directory", "other"}[fs]
}

// MarshalJSON marshals the FileStatus enum as a quoted JSON string.
// Note there is no UnmarshalJSON function, as marshalling to JSON is currently
// only used for printing debug information.
func (fs FileStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(fs.String())
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
			return StatusAbsent, nil
		}
		return 0, err
	}
	mode := fileInfo.Mode()

	if mode.IsRegular() {
		return StatusRegularFile, nil
	}

	if mode.IsDir() {
		return StatusDirectory, nil
	}

	if (mode & os.ModeSymlink) != 0 {
		return StatusLink, nil
	}

	return StatusOther, nil
}
