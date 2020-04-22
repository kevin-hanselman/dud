package artifact

// An Artifact is a file tracked by DUC
type Artifact struct {
	Checksum string
	Path     string
	IsDir    bool
}

// Status captures an Artifact's status as it pertains to a Cache and a workspace.
type Status struct {
	WorkspaceStatus WorkspaceStatus
	// HasChecksum is true if the Artifact has a valid Checksum member, false otherwise.
	HasChecksum bool
	// ChecksumInCache is true if a cache entry exists for the given checksum, false otherwise.
	ChecksumInCache bool
	// ContentsMatch is true if the workspace and cache files match, false otherwise.
	// For regular files, true means that the file contents are identical.
	// For links, true means that the workspace link points to the correct cache file.
	ContentsMatch bool
}

// WorkspaceStatus enumerates the states of an Artifact as it pertains to the workspace
type WorkspaceStatus int

const (
	// Absent means that the artifact does not exist in the workspace.
	Absent WorkspaceStatus = iota
	// RegularFile means that the artifact exists as a regular file in the workspace.
	RegularFile
	// Link means that the artifact exists as a link in the workspace.
	Link
	// Directory means that the artifact exists as a directory in the workspace.
	Directory
)

func (ws WorkspaceStatus) String() string {
	return [...]string{"Absent", "RegularFile", "Link"}[ws]
}
