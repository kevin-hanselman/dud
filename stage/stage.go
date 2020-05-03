package stage

import (
	"github.com/go-yaml/yaml"
	"github.com/kevlar1818/duc/artifact"
	cachePkg "github.com/kevlar1818/duc/cache"
	"github.com/kevlar1818/duc/strategy"
	"github.com/pkg/errors"
	"os"
)

// A Stage holds all information required to reproduce data. It is the primary
// artifact of DUC.
type Stage struct {
	Checksum   string
	WorkingDir string
	Outputs    []artifact.Artifact
}

// GetChecksum TODO
func (s *Stage) GetChecksum() string {
	return s.Checksum
}

// SetChecksum TODO
func (s *Stage) SetChecksum(c string) {
	s.Checksum = c
}

// Commit commits all Outputs of the Stage.
func (s *Stage) Commit(cache cachePkg.Cache, strategy strategy.CheckoutStrategy) error {
	for i := range s.Outputs {
		if err := cache.Commit(s.WorkingDir, &s.Outputs[i], strategy); err != nil {
			// TODO: unwind anything?
			return errors.Wrap(err, "stage commit failed")
		}
	}
	return nil
}

// Checkout checks out all Outputs of the Stage.
// TODO: will eventually checkout all inputs as well
func (s *Stage) Checkout(cache cachePkg.Cache, strategy strategy.CheckoutStrategy) error {
	for i := range s.Outputs {
		if err := cache.Checkout(s.WorkingDir, &s.Outputs[i], strategy); err != nil {
			// TODO: unwind anything?
			return errors.Wrap(err, "stage checkout failed")
		}
	}
	return nil
}

// ToFile saves the stage struct to yaml
func (s *Stage) ToFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	return yaml.NewEncoder(file).Encode(s)
}

// FromFile loads the stage struct from yaml
func FromFile(path string) (s Stage, err error) {
	stageFile, err := os.Open(path)
	if err != nil {
		return
	}
	err = yaml.NewDecoder(stageFile).Decode(&s)
	return
}
