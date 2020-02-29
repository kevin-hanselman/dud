package stage

import (
	"encoding/hex"
	"hash"
)

// A Stage holds all information required to reproduce data. It is the primary
// artifact of DUC.
type Stage struct {
	Checksum *Checksum
	Outputs  []Artifact
}

// An Artifact is a file that is a result of reproducing a stage.
type Artifact struct {
	Checksum *Checksum
	Path     string
}

// A Checksum represents the checksum of an artifact's bytes as a hex string.
type Checksum string

// ChecksumFromHash returns a Checksum from a Hash function
func ChecksumFromHash(h hash.Hash) Checksum {
	return Checksum(hex.EncodeToString(h.Sum(nil)))
}

func (c Checksum) String() string {
	return string(c)
}
