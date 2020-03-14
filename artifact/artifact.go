package artifact

// An Artifact is a file tracked by DUC
type Artifact struct {
	Checksum string
	Path     string
}

// FileStatus enumerates the states of an Artifact as it pertains to the workspace
type FileStatus int

const (
	// IsAbsent means that the artifact is absent from the workspace
	IsAbsent FileStatus = iota
	// IsRegularFile means that the artifact is present as a regular file in the workspace
	// TODO expand this to ContentsMatch, ContentsDiffer
	IsRegularFile
	// IsLink means that the artifact is present as a link in the workspace
	IsLink
)

// Status captures an Artifact's status as it pertains to a Cache and a workspace.
type Status struct {
	InCache    bool
	FileStatus FileStatus
}
