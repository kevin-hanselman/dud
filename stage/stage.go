package stage

import (
	"encoding/hex"
	"hash"
)

// A Stage holds all information required to reproduce data. It is the primary
// artifact of DUC.
type Stage struct {
	Checksum   string
	WorkingDir string
	Outputs    []Artifact
}

// An Artifact is a file that is a result of reproducing a stage.
type Artifact struct {
	Checksum string
	Path     string
}

// ChecksumFromHash returns a Checksum from a Hash function
func ChecksumFromHash(h hash.Hash) string {
	return hex.EncodeToString(h.Sum(nil))
}
