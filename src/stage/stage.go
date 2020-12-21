package stage

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kevin-hanselman/dud/src/artifact"
	"github.com/kevin-hanselman/dud/src/checksum"

	"gopkg.in/yaml.v3"
)

// A Stage holds all information required to reproduce data. It is the primary
// building block of Dud pipelines.
type Stage struct {
	// Checksum is the checksum of the Stage definition excluding Artifact
	// checksums. This checksum is used to determine when a Stage definition
	// has been modified by the user.
	Checksum string
	// Command is the string to be evaluated and executed by a shell.
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
	Checksum     string               `yaml:",omitempty"`
	Command      string               `yaml:",omitempty"`
	WorkingDir   string               `yaml:"working-dir,omitempty"`
	Dependencies []*artifact.Artifact `yaml:",omitempty"`
	Outputs      []*artifact.Artifact
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

func (stg Stage) toFileFormat() (out stageFileFormat) {
	out.Checksum = stg.Checksum
	out.Command = stg.Command
	out.WorkingDir = stg.WorkingDir

	if len(stg.Dependencies) > 0 {
		out.Dependencies = make([]*artifact.Artifact, len(stg.Dependencies))
		var i int = 0
		for _, art := range stg.Dependencies {
			// SkipCache is implicitly true for all dependencies. It's
			// redundant and noisy to write it to the Stage file, so we hide
			// it (making use of the 'omitempty' YAML directive) and set
			// SkipCache to true when loading the file (see FromFile).
			art.SkipCache = false
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
	stg.Checksum = sff.Checksum
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
var FromFile = func(stagePath string) (Stage, error) {
	var (
		stg Stage
		sff stageFileFormat
	)
	if err := fromYamlFile(stagePath, &sff); err != nil {
		return stg, err
	}
	// Clean all user-editable paths.
	for i, art := range sff.Dependencies {
		art.Path = filepath.Clean(art.Path)
		// Dependencies are only committed-to/checked-out-of the Cache if they
		// are an output of (i.e. owned by) another Stage, in which case said
		// owner Stage is responsible for interacting with the Cache.
		art.SkipCache = true
		sff.Dependencies[i] = art
	}
	for i, art := range sff.Outputs {
		art.Path = filepath.Clean(art.Path)
		sff.Outputs[i] = art
	}
	sff.WorkingDir = filepath.Clean(sff.WorkingDir)
	stg = sff.toStage()

	return stg, stg.validate()
}

func (stg Stage) validate() error {
	if strings.Contains(stg.WorkingDir, "..") {
		return fmt.Errorf("working directory %s is outside of the project root", stg.WorkingDir)
	}
	// First, check for direct overlap between Outputs and Dependencies.
	// Consolidate all Artifacts into a single map to facilitate the next step.
	// TODO: Only consolidate Artifacts with IsDir = true?
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

// CalculateChecksum returns the checksum of the Stage as it would be set in
// the Checksum field.
func (stg Stage) CalculateChecksum() (string, error) {
	cleanStage := Stage{
		Command:    stg.Command,
		WorkingDir: stg.WorkingDir,
	}
	cleanStage.Dependencies = make(map[string]*artifact.Artifact, len(stg.Dependencies))
	for _, art := range stg.Dependencies {
		newArt := *art
		newArt.Checksum = ""
		cleanStage.Dependencies[art.Path] = &newArt
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
	reader, writer := io.Pipe()
	defer reader.Close()
	go func() {
		json.NewEncoder(writer).Encode(cleanStage)
		writer.Close()
	}()
	return checksum.Checksum(reader)
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
