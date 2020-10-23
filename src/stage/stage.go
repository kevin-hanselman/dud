package stage

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kevin-hanselman/dud/src/artifact"

	"gopkg.in/yaml.v3"
)

// A Stage holds all information required to reproduce data. It is the primary
// building block of Dud pipelines.
type Stage struct {
	// The string to be evaluated and executed by a shell.
	Command string
	// WorkingDir is the directory in which the Stage's command is executed. It
	// is a directory path relative to the Dud root directory. An
	// empty value means the Stage's working directory _is_ the Dud root
	// directory. WorkingDir only affects the Stage's command; all outputs and
	// dependencies of the Stage should have paths relative to the project root.
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

var fromYamlFile = func(path string, sff *stageFileFormat) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	decoder := yaml.NewDecoder(file)
	decoder.KnownFields(true)
	return decoder.Decode(sff)
}

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
	// Clean all user-editable paths. Assume that the lock file has not been
	// tampered with. (We may consider relaxing that assumption at some point.)
	for i, art := range sff.Dependencies {
		art.Path = filepath.Clean(art.Path)
		sff.Dependencies[i] = art
	}
	for i, art := range sff.Outputs {
		art.Path = filepath.Clean(art.Path)
		sff.Outputs[i] = art
	}
	sff.WorkingDir = filepath.Clean(sff.WorkingDir)
	stg = sff.toStage()

	// Wipe any output from the previous deserialization.
	// This is important and hard to unit test.
	sff = stageFileFormat{}

	lockPath := FilePathForLock(stagePath)
	err := fromYamlFile(lockPath, &sff)
	if err != nil && !os.IsNotExist(err) {
		return locked, false, err
	}
	locked = sff.toStage()

	same := sameish(stg, locked)

	for artPath, art := range stg.Dependencies {
		// Dependencies are only committed-to/checked-out-of the Cache if they are an
		// output of (i.e. owned by) another Stage, in which case said owner
		// Stage is responsible for interacting with the Cache.
		art.SkipCache = true
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

	return stg, same, stg.validate()
}

func (stg Stage) validate() error {
	if strings.Contains(stg.WorkingDir, "..") {
		return fmt.Errorf("working directory %s is outside of the project root", stg.WorkingDir)
	}
	// First, check for direct overlap between Outputs and Dependencies.
	// Consolidate all Artifacts into a single map to facilitate the next step.
	allArtifacts := make(map[string]*artifact.Artifact, len(stg.Dependencies)+len(stg.Outputs))
	for artPath, art := range stg.Outputs {
		if _, ok := stg.Dependencies[artPath]; ok {
			return fmt.Errorf(
				"artifact %s is both a dependency and an output",
				artPath,
			)
		}
		allArtifacts[artPath] = art
	}
	for artPath, art := range stg.Dependencies {
		allArtifacts[artPath] = art
	}

	// Second, check if an Artifact is owned by any other (directory) Artifact
	// in the Stage.
	for artPath := range allArtifacts {
		if strings.Contains(artPath, "..") {
			return fmt.Errorf("artifact %s is outside of the project root", artPath)
		}
		parentArt, ok, err := FindDirArtifactOwnerForPath(artPath, allArtifacts)
		if err != nil {
			return err
		}
		if ok {
			return fmt.Errorf(
				"artifact %s conflicts with artifact %s",
				artPath,
				parentArt.Path,
			)
		}
	}
	return nil
}

// Serialize writes a Stage to the given writer.
func (stg *Stage) Serialize(writer io.Writer) error {
	return yaml.NewEncoder(writer).Encode(stg.toFileFormat())
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
	// TODO: Consider always running with "sh" for consistency.
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

// FindDirArtifactOwnerForPath searches the given map for a directory Artifact
// that should own relPath. relPath should share a base with the Artifacts in
// the map (hence the name).
func FindDirArtifactOwnerForPath(
	relPath string,
	artifacts map[string]*artifact.Artifact,
) (
	*artifact.Artifact,
	bool,
	error,
) {
	var owner *artifact.Artifact
	// Search for an Artifact whose Path is any directory in the input's lineage.
	// For example: given "bish/bash/bosh/file.txt", look for "bish", then
	// "bish/bash", then "bish/bash/bosh".
	fullDir := filepath.Dir(relPath)
	parts := strings.Split(fullDir, string(filepath.Separator))
	dir := ""
	for _, part := range parts {
		dir := filepath.Join(dir, part)
		owner, ok := artifacts[dir]
		// If we find a matching Artifact for any ancestor directory, the Stage in
		// question is only the owner if the Artifact is recursive, or
		// we've reached the immediate parent directory of the input.
		if ok && (owner.IsRecursive || dir == fullDir) {
			return owner, true, nil
		}
	}
	return owner, false, nil
}
