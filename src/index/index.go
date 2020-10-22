package index

import (
	"bufio"
	"fmt"
	"os"

	"github.com/kevin-hanselman/dud/src/artifact"
	"github.com/kevin-hanselman/dud/src/stage"
)

type entry struct {
	// IsLocked is true if the Stage in this entry is a locked version; i.e. it
	// has checksummed dependencies and outputs.
	IsLocked bool
	Stage    stage.Stage
}

// An Index holds an exhaustive set of Stages for a repository.
// TODO: Not threadsafe
type Index map[string]*entry

// AddStageFromPath adds a Stage at the given file path to the Index.
func (idx *Index) AddStageFromPath(path string) error {
	if _, ok := (*idx)[path]; ok {
		return fmt.Errorf("stage %s already in index", path)
	}
	stg, isLock, err := stage.FromFile(path)
	if err != nil {
		return err
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
	(*idx)[path] = &entry{
		IsLocked: isLock,
		Stage:    stg,
	}
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
// TODO Add a "full/bare" flag to enable only loading the fileFormat struct?
//      This would be a nice optimization not to load the whole Index when we
//      just want to check if a path is in the Index.
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
		if err := idx.AddStageFromPath(scanner.Text()); err != nil {
			return idx, err
		}
	}
	if err := scanner.Err(); err != nil {
		return idx, err
	}
	return idx, nil
}

func (idx Index) findOwner(artPath string) (string, *artifact.Artifact, error) {
	for stagePath, en := range idx {
		if art, ok := en.Stage.Outputs[artPath]; ok {
			return stagePath, art, nil
		}
		art, ok, err := stage.FindDirArtifactOwnerForPath(artPath, en.Stage.Outputs)
		if err != nil {
			return "", art, err
		}
		if ok {
			return stagePath, art, nil
		}
	}
	return "", &artifact.Artifact{}, nil
}