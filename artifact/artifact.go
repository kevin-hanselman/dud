package artifact

// An Artifact is a file tracked by DUC
type Artifact struct {
	Checksum string
	Path     string
	IsDir    bool
}

// Status captures an Artifact's status as it pertains to a Cache and a workspace.
type Status struct {
	// WorkspaceFileStatus represents the status of Artifact's file in the workspace.
	// TODO: We need some way to identify a "bad" workspace file status.
	// Replace and/or augment this with a boolean?
	WorkspaceFileStatus FileStatus
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
	case Absent:
		if stat.HasChecksum {
			if stat.ChecksumInCache {
				return "missing from workspace"
			}
			return "missing from cache and workspace"
		}
		return "unknown artifact"

	case RegularFile, Directory:
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

	case Link:
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

	case Other:
		return "invalid file type"
	}
	panic("exited switch unexpectedly")
}

// FileStatus enumerates the states of a file on the filesystem.
// TODO: move to fsutil?
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

func (ws FileStatus) String() string {
	return [...]string{"Absent", "RegularFile", "Link", "Directory", "Other"}[ws]
}
