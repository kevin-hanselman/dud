package stage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kevin-hanselman/dud/src/artifact"
	"github.com/kevin-hanselman/dud/src/checksum"
	"github.com/pkg/errors"

	"gopkg.in/yaml.v2"
)

// A Stage holds all information required to reproduce data. It is the primary
// building block of Dud pipelines.
type Stage struct {
	// Checksum is the checksum of the Stage definition excluding Artifact
	// checksums. This checksum is used to determine when a Stage definition
	// has been modified by the user.
	Checksum string `yaml:",omitempty"`
	// Command is the string to be evaluated and executed by a shell.
	Command string `yaml:",omitempty"`
	// WorkingDir is the directory in which the Stage's command is executed. It
	// is a directory path relative to the Dud root directory. An
	// empty value means the Stage's working directory _is_ the Dud root
	// directory. WorkingDir only affects the Stage's command; all inputs and
	// outputs of the Stage should have paths relative to the project root.
	WorkingDir string `yaml:"working-dir,omitempty"`
	// Inputs is a set of Artifacts which the Stage's Command needs to
	// operate. The Artifacts are keyed by their Path for faster lookup.
	Inputs map[string]*artifact.Artifact `yaml:",omitempty"`
	// Outputs is a set of Artifacts which are owned by the Stage. The
	// Artifacts are keyed by their Path for faster lookup.
	Outputs map[string]*artifact.Artifact
}

// Status holds everything necessary to qualify the state of a Stage.
type Status struct {
	// HasChecksum is true if the Stage had a non-empty Checksum field.
	HasChecksum bool
	// ChecksumMatches is true if the checksum of the Stage's definition
	// matches its Checksum field.
	ChecksumMatches bool
	ArtifactStatus  map[string]artifact.ArtifactWithStatus
}

// NewStatus initializes a new Status object.
func NewStatus() Status {
	s := Status{}
	s.ArtifactStatus = make(map[string]artifact.ArtifactWithStatus)
	return s
}

func (stg Stage) toFileFormat() (out Stage) {
	out.Checksum = stg.Checksum
	out.Command = stg.Command
	out.WorkingDir = stg.WorkingDir

	if len(stg.Inputs) > 0 {
		out.Inputs = make(map[string]*artifact.Artifact, len(stg.Inputs))
		for _, art := range stg.Inputs {
			// SkipCache is implicitly true for all inputs. It's
			// redundant and noisy to write it to the Stage file, so we hide
			// it (making use of the 'omitempty' YAML directive) and set
			// SkipCache to true when loading the file (see FromFile).
			art.SkipCache = false
			path := art.Path
			art.Path = ""
			out.Inputs[path] = art
		}
	}

	if len(stg.Outputs) > 0 {
		out.Outputs = make(map[string]*artifact.Artifact, len(stg.Outputs))
		for _, art := range stg.Outputs {
			path := art.Path
			art.Path = ""
			out.Outputs[path] = art
		}
	}
	return
}

var fromYamlFile = func(path string, stg *Stage) error {
	file, err := os.Open(path)
	if err != nil {
		return errors.Wrap(err, path)
	}
	defer file.Close()
	decoder := yaml.NewDecoder(file)
	decoder.SetStrict(true)
	err = decoder.Decode(stg)
	if err != nil {
		return errors.Wrap(err, path)
	}
	return nil
}

// FromFile loads a Stage from a file.
func FromFile(stagePath string) (stg Stage, err error) {
	var tempStage Stage
	if err = fromYamlFile(stagePath, &tempStage); err != nil {
		return
	}
	stg.Checksum = tempStage.Checksum
	stg.Command = tempStage.Command
	stg.Inputs = make(map[string]*artifact.Artifact, len(stg.Inputs))
	stg.Outputs = make(map[string]*artifact.Artifact, len(stg.Outputs))

	// Clean all user-editable paths.
	stg.WorkingDir = filepath.Clean(tempStage.WorkingDir)

	for path, art := range tempStage.Inputs {
		// yaml.v2 (and currently v3 as well) deserializes "  path.txt:" as
		// a nil map value, while "  path.txt: {}" deserializes as a zero map
		// value. Because of this, it's important to check for null pointers
		// here. This issue may be related:
		// https://github.com/go-yaml/yaml/issues/681
		if art == nil {
			art = new(artifact.Artifact)
		}
		art.Path = filepath.Clean(path)
		// Inputs are only committed-to/checked-out-of the Cache if they
		// are an output of (i.e. owned by) another Stage, in which case said
		// owner Stage is responsible for interacting with the Cache.
		art.SkipCache = true
		stg.Inputs[art.Path] = art
	}
	for path, art := range tempStage.Outputs {
		if art == nil {
			art = new(artifact.Artifact)
		}
		art.Path = filepath.Clean(path)
		stg.Outputs[art.Path] = art
	}

	return stg, stg.Validate()
}

// Validate returns an error describing a problem with the given Stage.
// If there are no problems with the Stage definition this method returns nil.
func (stg Stage) Validate() error {
	if strings.Contains(stg.WorkingDir, "..") {
		return fmt.Errorf("working directory %s is outside of the project root", stg.WorkingDir)
	}
	if filepath.IsAbs(stg.WorkingDir) {
		return fmt.Errorf("working directory %s is an absolute path", stg.WorkingDir)
	}
	if len(stg.Inputs)+len(stg.Outputs) == 0 {
		return errors.New("declared no inputs and no outputs")
	}
	// First, check for direct overlap between Outputs and Inputs.
	// Consolidate all Artifacts into a single map to facilitate the next step.
	// TODO: Only consolidate Artifacts with IsDir = true?
	allArtifacts := make(map[string]*artifact.Artifact, len(stg.Inputs)+len(stg.Outputs))
	for artPath, art := range stg.Outputs {
		if _, ok := stg.Inputs[artPath]; ok {
			return fmt.Errorf(
				"artifact %s is both an input and an output",
				artPath,
			)
		}
		allArtifacts[artPath] = art
	}
	for artPath, art := range stg.Inputs {
		allArtifacts[artPath] = art
	}

	// Second, check if an Artifact is owned by any other (directory) Artifact
	// in the Stage.
	for artPath := range allArtifacts {
		if strings.Contains(artPath, "..") {
			return fmt.Errorf("artifact %s is outside of the project root", artPath)
		}
		if filepath.IsAbs(artPath) {
			return fmt.Errorf("artifact %s is an absolute path", artPath)
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

// CalculateChecksum returns the checksum of the Stage as it would be set in
// the Checksum field.
func (stg Stage) CalculateChecksum() (string, error) {
	cleanStage := Stage{
		Command:    stg.Command,
		WorkingDir: stg.WorkingDir,
	}
	cleanStage.Inputs = make(map[string]*artifact.Artifact, len(stg.Inputs))
	for _, art := range stg.Inputs {
		newArt := *art
		newArt.Checksum = ""
		cleanStage.Inputs[art.Path] = &newArt
	}
	cleanStage.Outputs = make(map[string]*artifact.Artifact, len(stg.Outputs))
	for _, art := range stg.Outputs {
		newArt := *art
		newArt.Checksum = ""
		cleanStage.Outputs[art.Path] = &newArt
	}
	// TODO: Using encoding/gob gives sporadic checksum differences with this
	// method. Using encoding/json seems to alleviate the issue. We need to
	// understand this better.
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(cleanStage); err != nil {
		return "", err
	}
	return checksum.Checksum(buf)
}

// CreateCommand return an exec.Cmd for the Stage.
func (stg Stage) CreateCommand() *exec.Cmd {
	cmd := exec.Command("sh", "-c", stg.Command)
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
	// Search for an Artifact whose Path is any directory in the input's
	// lineage. For example: given "bish/bash/bosh/file.txt", look for "bish",
	// then "bish/bash", then "bish/bash/bosh".
	fullDir := filepath.Dir(relPath)
	parts := strings.Split(fullDir, string(filepath.Separator))
	dir := ""
	for _, part := range parts {
		dir := filepath.Join(dir, part)
		owner, ok := artifacts[dir]
		// If we find a matching Artifact for any ancestor directory, the Artifact
		// in question is only the owner if it is recursive, or if we've
		// reached the immediate parent directory of the input.
		if ok && (!owner.DisableRecursion || dir == fullDir) {
			return owner, true, nil
		}
	}
	return owner, false, nil
}
