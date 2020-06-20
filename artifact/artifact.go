package artifact

import (
	"fmt"
	"github.com/kevlar1818/duc/fsutil"
)

// An Artifact is a file tracked by DUC
type Artifact struct {
	Checksum    string `yaml:",omitempty"`
	Path        string
	IsDir       bool `yaml:",omitempty"`
	IsRecursive bool `yaml:",omitempty"`
}

// Status captures an Artifact's status as it pertains to a Cache and a workspace.
type Status struct {
	// WorkspaceFileStatus represents the status of Artifact's file in the workspace.
	// TODO: We need some way to identify a "bad" workspace file status.
	// Replace and/or augment this with a boolean?
	WorkspaceFileStatus fsutil.FileStatus
	// HasChecksum is true if the Artifact has a valid Checksum member, false otherwise.
	HasChecksum bool
	// ChecksumInCache is true if a cache entry exists for the given checksum, false otherwise.
	ChecksumInCache bool
	// ContentsMatch is true if the workspace and cache files are identical; it
	// is false otherwise. For regular files, true means that the file contents
	// are identical. For links, true means that the workspace link points to
	// the correct cache file.
	ContentsMatch bool
}

func (stat Status) String() string {
	switch stat.WorkspaceFileStatus {
	case fsutil.Absent:
		if stat.HasChecksum {
			if stat.ChecksumInCache {
				return "missing from workspace"
			}
			return "missing from cache and workspace"
		}
		return "unknown artifact"

	case fsutil.RegularFile, fsutil.Directory:
		if stat.HasChecksum {
			if stat.ChecksumInCache {
				if stat.ContentsMatch {
					return "up-to-date"
				}
				return "modified"
			}
			return "missing from cache"
		}
		return "uncommitted"

	case fsutil.Link:
		if stat.HasChecksum {
			if stat.ChecksumInCache {
				if stat.ContentsMatch {
					return "linked to cache"
				}
				return "incorrect link"
			}
			return "broken link"
		}
		return "link with no checksum"

	case fsutil.Other:
		return "invalid file type"
	}
	panic("exited switch unexpectedly")
}

var fileStatusFromPath = fsutil.FileStatusFromPath

// FromPath returns a new Artifact tracking the given path.
// TODO: When adding new files, the Index needs to be consulted to ensure
// exactly one Artifact owns a given file, and that exactly one Stage owns
// a given Artifact.
func FromPath(path string, isRecursive bool) (art Artifact, err error) {
	status, err := fileStatusFromPath(path)
	if err != nil {
		return
	}
	switch status {
	case fsutil.Absent:
		return art, fmt.Errorf("path %v does not exist", path)
	case fsutil.Other, fsutil.Link:
		return art, fmt.Errorf("unsupported file type for path %v", path)
	}

	isDir := status == fsutil.Directory
	return Artifact{
		Checksum:    "",
		Path:        path,
		IsDir:       isDir,
		IsRecursive: isRecursive && isDir,
	}, nil
}
