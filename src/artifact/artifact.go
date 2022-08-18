package artifact

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/kevin-hanselman/dud/src/fsutil"
)

// An Artifact is a file or directory that is tracked by Dud.
type Artifact struct {
	// Checksum is the hex digest Artifact's hashed contents. It is used to
	// locate the Artifact in a Cache.
	Checksum string `yaml:",omitempty" json:"checksum,omitempty"`
	// Path is the file path to the Artifact in the workspace. It is always
	// relative to the project root directory.
	Path string `yaml:",omitempty" json:"path,omitempty"`
	// If IsDir is true then the Artifact is a directory.
	IsDir bool `yaml:"is-dir,omitempty" json:"is-dir,omitempty"`
	// If DisableRecursion is true then the Artifact does not recurse
	// sub-directories.
	DisableRecursion bool `yaml:"disable-recursion,omitempty" json:"disable-recursion,omitempty"`
	// If SkipCache is true then the Artifact is not stored in the Cache. When
	// the Artifact is committed, its checksum is updated, but the Artifact is
	// not moved to the Cache. The checkout operation is a no-op.
	SkipCache bool `yaml:"skip-cache,omitempty" json:"skip-cache,omitempty"`
}

type oldArtifact struct {
	Checksum         string
	Path             string
	IsDir            bool
	DisableRecursion bool
	SkipCache        bool
}

// UnmarshalJSON enables backwards-compatibility with the original Artifact
// struct, which did not have struct tags for custom JSON serialization.
func (a *Artifact) UnmarshalJSON(b []byte) error {
	// First, we try to unmarshal an Artifact using YAML in "strict" mode, which
	// will error-out if any extra fields are found (e.g. the old "IsDir" field
	// instead of "is-dir"). (JSON is valid YAML, so unmarshalling from YAML
	// when the underlying encoding is JSON is perfectly safe.) If
	// unmarshalling from YAML succeeds, the resulting Artifact is already
	// using the latest schema and we can exit.
	if err := yaml.UnmarshalStrict(b, a); err == nil {
		return nil
	}
	// If unmarshalling from YAML failed, chances are the underlying data is
	// using the old schema, so try to unmarshal it. If we still get an error,
	// fail with that error; otherwise, copy the data from the old schema to
	// the Artifact and exit.
	// TODO: Technically, yaml.UnmarshalStrict and yaml.Unmarshal should both work
	// here, but I get spurious "field X not found" errors. oldArtifact not
	// being exported is not the issue. Need to investigate.
	var old oldArtifact
	if err := json.Unmarshal(b, &old); err != nil {
		return err
	}
	*a = Artifact(old)
	return nil
}

// Status captures an Artifact's status as it pertains to a Cache and a workspace.
type Status struct {
	Artifact
	// WorkspaceFileStatus represents the status of Artifact's file in the workspace.
	WorkspaceFileStatus fsutil.FileStatus
	// HasChecksum is true if the Artifact has a valid Checksum field, false otherwise.
	// TODO: Might be able to get rid of this if we have the Artifact in question.
	HasChecksum bool
	// ChecksumInCache is true if a cache entry exists for the given checksum, false otherwise.
	ChecksumInCache bool
	// ContentsMatch is true if the workspace and cache files are identical; it
	// is false otherwise. For regular files, true means that the file contents
	// are identical. For links, true means that the workspace link points to
	// the correct cache file.
	ContentsMatch bool
	// ChildrenStatus holds the status of any child artifacts, mapped to their
	// respective file paths.
	ChildrenStatus map[string]*Status
}

func (stat Status) dirStatusCounts(counts map[string]int) {
	// len(nil map) returns 0
	if len(stat.ChildrenStatus) == 0 {
		counts["empty directory"]++
	} else {
		counts["directory"]++
	}
	for _, childStatus := range stat.ChildrenStatus {
		if childStatus.IsDir {
			childStatus.dirStatusCounts(counts)
		} else {
			counts[childStatus.String()]++
		}
	}
}

func sortCounts(counts map[string]int) []string {
	keys := make([]string, len(counts))
	i := 0
	for key := range counts {
		keys[i] = key
		i++
	}
	sort.Slice(keys, func(a, b int) bool {
		aKey := keys[a]
		aVal := counts[aKey]
		bKey := keys[b]
		bVal := counts[keys[b]]
		// If the values are equal, fallback to lexicographic order of the
		// string keys. This is primarily for writing deterministic tests.
		if aVal == bVal {
			return aKey < bKey
		}
		// Sort by highest value first.
		return aVal > bVal
	})
	return keys
}

func (stat Status) String() string {
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
		return "missing and not committed"

	case fsutil.StatusRegularFile:
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
			out.WriteString("not committed")
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

	if stat.IsDir {
		counts := make(map[string]int)
		stat.dirStatusCounts(counts)
		countStrings := make([]string, len(counts))
		for i, status := range sortCounts(counts) {
			countStrings[i] = fmt.Sprintf("%dx %s", counts[status], status)
		}
		return strings.Join(countStrings, ", ")
	}

	panic(fmt.Sprintf("unhandled case in artifact.Status.String(): %#v", stat))
}
