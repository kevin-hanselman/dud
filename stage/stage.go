package stage

import (
	"os"
	"os/exec"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/jinzhu/copier"
	"github.com/kevlar1818/duc/artifact"
	"github.com/kevlar1818/duc/cache"
	"github.com/kevlar1818/duc/fsutil"
	"github.com/kevlar1818/duc/strategy"
	"github.com/pkg/errors"
)

var fromYamlFile = fsutil.FromYamlFile

// A Stage holds all information required to reproduce data. It is the primary
// building block of Duc pipelines.
type Stage struct {
	Command      string `yaml:",omitempty"`
	WorkingDir   string
	Dependencies []artifact.Artifact `yaml:",omitempty"`
	Outputs      []artifact.Artifact
}

// Status holds a map of artifact names to statuses
type Status map[string]artifact.ArtifactWithStatus

// IsEquivalent return true if the two Stage objects are deeply equal in all
// fields besides Artifact Checksum fields.
func (s *Stage) IsEquivalent(other Stage) (bool, error) {
	var selfClean, otherClean Stage
	if err := copier.Copy(&selfClean, s); err != nil {
		return false, err
	}
	if err := copier.Copy(&otherClean, other); err != nil {
		return false, err
	}
	// Remove all artifact checksums
	for i := range selfClean.Outputs {
		selfClean.Outputs[i].Checksum = ""
	}
	for i := range selfClean.Dependencies {
		selfClean.Dependencies[i].Checksum = ""
	}
	for i := range otherClean.Outputs {
		otherClean.Outputs[i].Checksum = ""
	}
	for i := range otherClean.Dependencies {
		otherClean.Dependencies[i].Checksum = ""
	}
	return cmp.Equal(selfClean, otherClean), nil
}

// FromFile loads a Stage from a file. If a lock file exists and is equivalent
// (see stage.IsEquivalent), it loads the Stage's locked version.
var FromFile = func(stagePath string) (Stage, bool, error) {
	var stg, locked Stage
	if err := fromYamlFile(stagePath, &stg); err != nil {
		return stg, false, err
	}
	lockPath := FilePathForLock(stagePath)
	err := fromYamlFile(lockPath, &locked)
	if os.IsNotExist(err) {
		return stg, false, nil
	} else if err != nil {
		return locked, false, err
	}
	isEquiv, err := locked.IsEquivalent(stg)
	if err != nil {
		return stg, false, err
	}
	if isEquiv {
		return locked, true, nil
	}
	return stg, false, nil
}

// Commit commits all Outputs of the Stage.
func (s *Stage) Commit(ch cache.Cache, strat strategy.CheckoutStrategy) error {
	for i := range s.Outputs {
		if err := ch.Commit(s.WorkingDir, &s.Outputs[i], strat); err != nil {
			// TODO: unwind anything?
			return errors.Wrap(err, "stage commit failed")
		}
	}
	return nil
}

// Checkout checks out all Outputs of the Stage.
func (s *Stage) Checkout(ch cache.Cache, strat strategy.CheckoutStrategy) error {
	for i := range s.Outputs {
		if err := ch.Checkout(s.WorkingDir, &s.Outputs[i], strat); err != nil {
			// TODO: unwind anything?
			return errors.Wrap(err, "stage checkout failed")
		}
	}
	return nil
}

// Status checks the status of all Outputs of the Stage.
func (s *Stage) Status(ch cache.Cache) (Status, error) {
	stat := make(Status)
	for _, art := range s.Outputs {
		artStatus, err := ch.Status(s.WorkingDir, art)
		if err != nil {
			return stat, errors.Wrap(err, "stage status failed")
		}
		stat[art.Path] = artifact.ArtifactWithStatus{
			Artifact: art,
			Status:   artStatus,
		}
	}
	return stat, nil
}

// Run runs the Stage's command unless the Stage is up-to-date.
func (s *Stage) Run(ch cache.Cache) (upToDate bool, err error) {
	status, err := s.Status(ch)
	if err != nil {
		return false, err
	}
	if isUpToDate(status) {
		return true, nil
	}
	if s.Command == "" {
		return false, nil
	}
	return false, runCommand(s.Command)
}

// FromPaths creates a Stage from one or more file paths.
// TODO: rename or delete (to differentiate from FromFile)
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

// FilePathForLock returns the lock file path given a Stage path.
func FilePathForLock(stagePath string) string {
	var str strings.Builder
	// TODO: check for valid YAML, or at least .y(a)ml extension?
	// TODO: check for .lock suffix already in input?
	str.WriteString(stagePath)
	str.WriteString(".lock")
	return str.String()
}

func isUpToDate(status Status) bool {
	for _, artStatus := range status {
		if !artStatus.ContentsMatch {
			return false
		}
	}
	return true
}

var runCommand = func(command string) error {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "sh"
	}
	cmd := exec.Command(shell, "-c", command)
	// TODO: Set cmd.Dir appropriately. This could be tricky, as Stage working
	// dirs are relative and need to be resolved to full paths for cmd.Dir to
	// be set correctly. We'll probably have to look up a Stage in the Index to
	// get its path, then concatenate the WorkingDir and resolve the path.
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
