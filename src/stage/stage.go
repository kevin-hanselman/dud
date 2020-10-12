package stage

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kevin-hanselman/duc/src/artifact"
	"github.com/kevin-hanselman/duc/src/fsutil"
)

// A Stage holds all information required to reproduce data. It is the primary
// building block of Duc pipelines.
type Stage struct {
	// The string to be evaluated and executed by a shell.
	Command string
	// WorkingDir is a directory path relative to the Duc root directory. An
	// empty value means the Stage's working directory _is_ the Duc root
	// directory. All outputs and dependencies of the Stage are themselves
	// relative to WorkingDir. The Stage's Command is executed in WorkingDir.
	WorkingDir string
	// Dependencies is a set of Artifacts which the Stage's Command needs to
	// operate. The Artifacts are keyed by their Path for faster lookup.
	Dependencies map[string]*artifact.Artifact
	// Outputs is a set of Artifacts which are owned by the Stage. The
	// Artifacts are keyed by their Path for faster lookup.
	Outputs map[string]*artifact.Artifact
}

type stageFileFormat struct {
	Command      string               `yaml:",omitempty"`
	WorkingDir   string               `yaml:"working-dir,omitempty"`
	Dependencies []*artifact.Artifact `yaml:",omitempty"`
	Outputs      []*artifact.Artifact
}

// Status is a map of Artifact Paths to statuses
type Status map[string]artifact.ArtifactWithStatus

func (stg Stage) toFileFormat() (out stageFileFormat) {
	out.Command = stg.Command
	out.WorkingDir = stg.WorkingDir

	if len(stg.Dependencies) > 0 {
		out.Dependencies = make([]*artifact.Artifact, len(stg.Dependencies))
		var i int = 0
		for _, art := range stg.Dependencies {
			out.Dependencies[i] = art
			i++
		}
	}

	if len(stg.Outputs) > 0 {
		out.Outputs = make([]*artifact.Artifact, len(stg.Outputs))
		var i int = 0
		for _, art := range stg.Outputs {
			out.Outputs[i] = art
			i++
		}
	}
	return
}

func (sff stageFileFormat) toStage() (stg Stage) {
	stg.Command = sff.Command
	stg.WorkingDir = sff.WorkingDir

	if len(sff.Dependencies) > 0 {
		stg.Dependencies = make(
			map[string]*artifact.Artifact,
			len(sff.Dependencies),
		)
		for _, art := range sff.Dependencies {
			stg.Dependencies[art.Path] = art
		}
	}

	if len(sff.Outputs) > 0 {
		stg.Outputs = make(
			map[string]*artifact.Artifact,
			len(sff.Outputs),
		)
		for _, art := range sff.Outputs {
			stg.Outputs[art.Path] = art
		}
	}
	return
}

func sameish(stgA, stgB Stage) bool {
	if stgA.Command != stgB.Command {
		return false
	}
	if stgA.WorkingDir != stgB.WorkingDir {
		return false
	}
	if len(stgA.Outputs) != len(stgB.Outputs) {
		return false
	}
	if len(stgA.Dependencies) != len(stgB.Dependencies) {
		return false
	}
	return true
}

// for mocking
var fromYamlFile = fsutil.FromYamlFile

// FromFile loads a Stage from a file. If a lock file for the Stage exists,
// this function uses any Artifact.Checksums it can from the lock file.
var FromFile = func(stagePath string) (Stage, bool, error) {
	var (
		stg, locked Stage
		sff         stageFileFormat
	)
	if err := fromYamlFile(stagePath, &sff); err != nil {
		return stg, false, err
	}
	stg = sff.toStage()

	// Wipe any output from the previous deserialization.
	// This is important and hard to unit test.
	sff = stageFileFormat{}

	lockPath := FilePathForLock(stagePath)
	err := fromYamlFile(lockPath, &sff)
	if os.IsNotExist(err) {
		return stg, false, nil
	} else if err != nil {
		return locked, false, err
	}
	locked = sff.toStage()

	same := sameish(stg, locked)

	for artPath, art := range stg.Dependencies {
		lockedArt, ok := locked.Dependencies[artPath]
		if ok && art.IsEquivalent(*lockedArt) {
			stg.Dependencies[artPath] = lockedArt
		} else {
			same = false
		}
	}

	for artPath, art := range stg.Outputs {
		lockedArt, ok := locked.Outputs[artPath]
		if ok && art.IsEquivalent(*lockedArt) {
			stg.Outputs[artPath] = lockedArt
		} else {
			same = false
		}
	}

	return stg, same, nil
}

// ToFile writes a Stage to the given file path. It is important to use this
// method instead of bare fsutil.ToYamlFile because a Stage file is converted
// to a simplified format when stored on disk.
func (stg *Stage) ToFile(path string) error {
	return fsutil.ToYamlFile(path, stg.toFileFormat())
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

// CreateCommand return an exec.Cmd for the Stage.
func (stg Stage) CreateCommand() *exec.Cmd {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "sh"
	}
	cmd := exec.Command(shell, "-c", stg.Command)
	cmd.Dir = filepath.Clean(stg.WorkingDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}
