package artifact

import "github.com/kevlar1818/duc/fsutil"

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
