package stage

import (
	"github.com/kevlar1818/duc/artifact"
	cachePkg "github.com/kevlar1818/duc/cache"
	"github.com/kevlar1818/duc/checksum"
	"github.com/kevlar1818/duc/strategy"
	"github.com/pkg/errors"
	"strings"
)

// A Stage holds all information required to reproduce data. It is the primary
// building block of DUC pipelines.
type Stage struct {
	Checksum   string `yaml:",omitempty"`
	WorkingDir string
	Outputs    []artifact.Artifact
}

// Status holds a map of artifact names to statuses
type Status map[string]string

// UpdateChecksum updates the Checksum field of the Stage.
func (s *Stage) UpdateChecksum() (err error) {
	cleanStage := Stage{
		WorkingDir: s.WorkingDir,
		Outputs:    make([]artifact.Artifact, len(s.Outputs)),
	}
	for i, art := range s.Outputs {
		art.Checksum = ""
		cleanStage.Outputs[i] = art
	}
	s.Checksum, err = checksum.ChecksumObject(cleanStage)
	return
}

// Commit commits all Outputs of the Stage.
func (s *Stage) Commit(cache cachePkg.Cache, strat strategy.CheckoutStrategy) error {
	for i := range s.Outputs {
		if err := cache.Commit(s.WorkingDir, &s.Outputs[i], strat); err != nil {
			// TODO: unwind anything?
			return errors.Wrap(err, "stage commit failed")
		}
	}
	return s.UpdateChecksum()
}

// Checkout checks out all Outputs of the Stage.
// TODO: will eventually checkout all inputs as well
func (s *Stage) Checkout(cache cachePkg.Cache, strat strategy.CheckoutStrategy) error {
	for i := range s.Outputs {
		if err := cache.Checkout(s.WorkingDir, &s.Outputs[i], strat); err != nil {
			// TODO: unwind anything?
			return errors.Wrap(err, "stage checkout failed")
		}
	}
	return nil
}

// Status checks the status of all Outputs of the Stage.
// TODO: eventually report status of inputs as well
func (s *Stage) Status(cache cachePkg.Cache) (Status, error) {
	stat := make(Status)
	for _, art := range s.Outputs {
		artStatus, err := cache.Status(s.WorkingDir, art)
		if err != nil {
			return stat, errors.Wrap(err, "stage status failed")
		}
		stat[art.Path] = artStatus.String()
	}
	return stat, nil
}

// FromPaths creates a Stage from one or more file paths.
func FromPaths(isRecursive bool, paths ...string) (stg Stage, err error) {
	stg.Outputs = make([]artifact.Artifact, len(paths))

	for i, path := range paths {
		stg.Outputs[i], err = artifact.FromPath(path, isRecursive)
		if err != nil {
			return
		}
	}
	return
}

// FilePathForLock returns the lock-file path given a stage file path.
func FilePathForLock(stagePath string) string {
	var str strings.Builder
	// TODO: check for valid YAML, or at least .y(a)ml extension?
	str.WriteString(stagePath)
	str.WriteString(".lock")
	return str.String()
}
