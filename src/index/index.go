package index

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/kevin-hanselman/dud/src/artifact"
	"github.com/kevin-hanselman/dud/src/stage"
	"github.com/pkg/errors"
)

// An Index holds an exhaustive set of Stages for a repository.
// Not threadsafe.
type Index map[string]*stage.Stage

type unknownStageError struct {
	stagePath string
}

func (e unknownStageError) Error() string {
	return fmt.Sprintf("unknown stage %#v", e.stagePath)
}

// AddStage adds the given Stage to the Index, with the given path as the key.
func (idx *Index) AddStage(stg stage.Stage, path string) error {
	if _, ok := (*idx)[path]; ok {
		return fmt.Errorf("stage %s already in index", path)
	}
	for artPath := range stg.Outputs {
		ownerPath, _ := idx.findOwner(artPath)
		if ownerPath != "" {
			return fmt.Errorf(
				"%s: artifact %s already owned by %s",
				path,
				artPath,
				ownerPath,
			)
		}
	}
	(*idx)[path] = &stg
	return nil
}

func (idx *Index) RemoveStage(path string) error {
	if _, ok := (*idx)[path]; !ok {
		return unknownStageError{path}
	}
	delete(*idx, path)
	return nil
}

// ToFile writes the Index to the specified file path.
// To prevent the Index from going stale, Stages themselves aren't written to
// the Index file; the Index only tracks their paths.
// TODO no tests
func (idx Index) ToFile(indexPath string) error {
	errPrefix := fmt.Sprintf("writing index to %s", indexPath)
	// TODO: If we stop relying on the project-wide lock file, this should be
	// flocked.
	file, err := os.Create(indexPath)
	if err != nil {
		return errors.Wrap(err, errPrefix)
	}
	defer file.Close()

	// Sort the stage paths so the index file is written deterministically.
	for _, stagePath := range idx.SortStagePaths() {
		if _, err := fmt.Fprintln(file, stagePath); err != nil {
			return errors.Wrapf(err, "%s: write %s", errPrefix, stagePath)
		}
	}
	return nil
}

// SortStagePaths returns a sorted slice of Stage paths stored in the Index.
func (idx Index) SortStagePaths() []string {
	paths := []string{}
	for stagePath := range idx {
		paths = append(paths, stagePath)
	}
	sort.Strings(paths)
	return paths
}

// FromFile reads and returns an Index from the specified file path.
// See ToFile docs for more context.
// TODO no tests
func FromFile(path string) (Index, error) {
	errPrefix := fmt.Sprintf("load index from %s", path)
	var idx Index
	file, err := os.Open(path)
	if err != nil {
		return idx, errors.Wrap(err, errPrefix)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	idx = make(Index)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		stg, err := stage.FromFile(line)
		if err != nil {
			return idx, errors.Wrap(err, errPrefix)
		}
		if err := idx.AddStage(stg, line); err != nil {
			return idx, errors.Wrap(err, errPrefix)
		}
	}
	if err := scanner.Err(); err != nil {
		return idx, errors.Wrap(err, errPrefix)
	}
	return idx, nil
}

func (idx Index) findOwner(artPath string) (string, *artifact.Artifact) {
	for stagePath, stg := range idx {
		if art, ok := stg.Outputs[artPath]; ok {
			return stagePath, art
		}
		art, ok := stage.FindDirArtifactOwnerForPath(artPath, stg.Outputs)
		if ok {
			return stagePath, art
		}
	}
	return "", nil
}
