package stage

import (
	"crypto/sha1"
	"encoding/gob"
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

// SetChecksum calculates the checksum of a Stage (sans Checksum field)
// then sets the Checksum field accordingly.
func (s *Stage) SetChecksum() error {
	s.Checksum = ""

	h := sha1.New()
	enc := gob.NewEncoder(h)
	if err := enc.Encode(s); err != nil {
		return err
	}
	s.Checksum = ChecksumFromHash(h)
	return nil
}
