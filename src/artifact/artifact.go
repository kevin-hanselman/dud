package artifact

import (
	"fmt"
	"strings"

	"github.com/kevin-hanselman/dud/src/fsutil"
)

// TODO: Consider adding JSON directives to Artifact as well for serializing
// directoryManifests.

// An Artifact is a file or directory that is tracked by Dud.
type Artifact struct {
	// Checksum is the hex digest Artifact's hashed contents. It is used to
	// locate the Artifact in a Cache.
	Checksum string `yaml:",omitempty"`
	// Path is the file path to the Artifact in the workspace. It is always
	// relative to the project root directory.
	Path string `yaml:",omitempty"`
	// If IsDir is true then the Artifact is a directory.
	IsDir bool `yaml:"is-dir,omitempty"`
	// If DisableRecursion is true then the Artifact does not recurse sub-directories
	DisableRecursion bool `yaml:"disable-recursion,omitempty"`
	// If SkipCache is true then the Artifact is not stored in the Cache. When
	// the Artifact is committed, its checksum is updated, but the Artifact is
	// not moved to the Cache. The checkout operation is a no-op.
	SkipCache bool `yaml:"skip-cache,omitempty"`
}

// Status captures an Artifact's status as it pertains to a Cache and a workspace.
type Status struct {
	// WorkspaceFileStatus represents the status of Artifact's file in the workspace.
	// TODO: We need some way to identify a "bad" workspace file status.
	// Replace and/or augment this with a boolean?
	WorkspaceFileStatus fsutil.FileStatus
	// HasChecksum is true if the Artifact has a valid Checksum field, false otherwise.
	HasChecksum bool
	// ChecksumInCache is true if a cache entry exists for the given checksum, false otherwise.
	ChecksumInCache bool
	// ContentsMatch is true if the workspace and cache files are identical; it
	// is false otherwise. For regular files, true means that the file contents
	// are identical. For links, true means that the workspace link points to
	// the correct cache file.
	ContentsMatch bool
}

// ArtifactWithStatus is an Artifact with a matched Status.
type ArtifactWithStatus struct {
	Artifact
	Status
}

func (stat ArtifactWithStatus) String() string {
	isDir := stat.WorkspaceFileStatus == fsutil.StatusDirectory
	isAbsent := stat.WorkspaceFileStatus == fsutil.StatusAbsent
	if (stat.IsDir != isDir) && !isAbsent {
		return fmt.Sprintf("incorrect file type: %s", stat.WorkspaceFileStatus)
	}
	isRegularFile := stat.WorkspaceFileStatus == fsutil.StatusRegularFile
	if stat.SkipCache && !isRegularFile {
		return fmt.Sprintf("incorrect file type: %s (not cached)", stat.WorkspaceFileStatus)
	}
	switch stat.WorkspaceFileStatus {
	case fsutil.StatusAbsent:
		if stat.HasChecksum {
			if stat.ChecksumInCache {
				return "missing from workspace"
			}
			return "missing from cache and workspace"
		}
		return "unknown artifact"

	case fsutil.StatusRegularFile, fsutil.StatusDirectory:
		var out strings.Builder
		if stat.HasChecksum {
			if stat.ChecksumInCache || stat.SkipCache {
				if stat.ContentsMatch {
					out.WriteString("up-to-date")
				} else {
					out.WriteString("modified")
				}
			} else {
				out.WriteString("missing from cache")
			}
		} else {
			out.WriteString("uncommitted")
		}
		if stat.SkipCache {
			out.WriteString(" (not cached)")
		}
		return out.String()

	case fsutil.StatusLink:
		if stat.HasChecksum {
			if stat.ChecksumInCache {
				if stat.ContentsMatch {
					return "up-to-date (link)"
				}
				return "incorrect link"
			}
			return "broken link"
		}
		return "link with no checksum"

	case fsutil.StatusOther:
		return "invalid file type"
	}
	panic("exited switch unexpectedly")
}
