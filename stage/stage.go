package stage

import (
	"crypto/sha1"
	"encoding/gob"
	"github.com/kevlar1818/duc/artifact"
	"github.com/kevlar1818/duc/cache"
	"github.com/kevlar1818/duc/fsutil"
	"github.com/pkg/errors"
)

// A Stage holds all information required to reproduce data. It is the primary
// artifact of DUC.
type Stage struct {
	Checksum   string
	WorkingDir string
	Outputs    []artifact.Artifact
}

// SetChecksum calculates the checksum of a Stage (sans Checksum field)
// then sets the Checksum field accordingly.
// TODO: make private (and call inside Commit)?
func (s *Stage) SetChecksum() error {
	s.Checksum = ""

	h := sha1.New()
	enc := gob.NewEncoder(h)
	if err := enc.Encode(s); err != nil {
		return err
	}
	s.Checksum = fsutil.HashToHexString(h)
	return nil
}

// Commit commits all Outputs of the Stage.
func (s *Stage) Commit(cacheDir string, strategy cache.CheckoutStrategy) error {
	for i := range s.Outputs {
		if err := s.Outputs[i].Commit(s.WorkingDir, cacheDir, strategy); err != nil {
			// TODO: unwind anything?
			return errors.Wrap(err, "stage commit failed")
		}
	}
	return nil
}

// Checkout checks out all Outputs of the Stage.
// TODO: will eventually checkout all inputs as well
func (s *Stage) Checkout(cacheDir string, strategy cache.CheckoutStrategy) error {
	for i := range s.Outputs {
		if err := s.Outputs[i].Checkout(s.WorkingDir, cacheDir, strategy); err != nil {
			// TODO: unwind anything?
			return errors.Wrap(err, "stage checkout failed")
		}
	}
	return nil
}
