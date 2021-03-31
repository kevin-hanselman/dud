package index

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/kevin-hanselman/dud/src/artifact"
	"github.com/kevin-hanselman/dud/src/stage"
)

// An Index holds an exhaustive set of Stages for a repository.
// Not threadsafe.
type Index map[string]*stage.Stage

// AddStage adds the given Stage to the Index, with the given path as the key.
func (idx *Index) AddStage(stg stage.Stage, path string) error {
	if _, ok := (*idx)[path]; ok {
		return fmt.Errorf("stage %s already in index", path)
	}
	for artPath := range stg.Outputs {
		ownerPath, _, err := idx.findOwner(artPath)
		if err != nil {
			return err
		} else if ownerPath != "" {
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

// ToFile writes the Index to the specified file path.
// To prevent the Index from going stale, Stages themselves aren't written to
// the Index file; the Index only tracks their paths.
// TODO no tests
func (idx Index) ToFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	for path := range idx {
		if _, err := fmt.Fprintln(file, path); err != nil {
			return err
		}
	}
	return nil
}

// FromFile reads and returns an Index from the specified file path.
// See ToFile docs for more context.
// TODO no tests
func FromFile(path string) (Index, error) {
	var idx Index
	file, err := os.Open(path)
	if err != nil {
		return idx, err
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
			return idx, err
		}
		if err := idx.AddStage(stg, line); err != nil {
			return idx, err
		}
	}
	if err := scanner.Err(); err != nil {
		return idx, err
	}
	return idx, nil
}

func (idx Index) findOwner(artPath string) (string, *artifact.Artifact, error) {
	for stagePath, stg := range idx {
		if art, ok := stg.Outputs[artPath]; ok {
			return stagePath, art, nil
		}
		art, ok, err := stage.FindDirArtifactOwnerForPath(artPath, stg.Outputs)
		if err != nil {
			return "", art, err
		}
		if ok {
			return stagePath, art, nil
		}
	}
	return "", nil, nil
}
