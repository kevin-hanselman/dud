package index

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kevin-hanselman/duc/artifact"
	"github.com/kevin-hanselman/duc/fsutil"
	"github.com/kevin-hanselman/duc/stage"
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

var fromYamlFile = fsutil.FromYamlFile

// AddStagesFromPaths adds the Stages at the given paths to the Index.
func (idx *Index) AddStagesFromPaths(paths ...string) error {
	for _, path := range paths {
		stg, isLock, err := stage.FromFile(path)
		if err != nil {
			return err
		}
		(*idx)[path] = &entry{
			IsLocked: isLock,
			Stage:    stg,
		}
	}
	return nil
}

// ToFile writes the Index to the specified file path.
// To prevent the Index from going stale, Stages themselves aren't written to
// the Index file; the Index only tracks their paths and other metadata (e.g.
// commit set membership).
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
		stagePath := scanner.Text()
		stg, isLock, err := stage.FromFile(stagePath)
		if err != nil {
			return idx, err
		}
		idx[stagePath] = &entry{
			IsLocked: isLock,
			Stage:    stg,
		}
	}
	return idx, nil
}

func (idx Index) findOwner(artPath string) (string, *artifact.Artifact, error) {
	for stagePath, en := range idx {
		relPath, err := filepath.Rel(en.Stage.WorkingDir, artPath)
		if err != nil {
			return "", &artifact.Artifact{}, err
		}
		if art, ok := en.Stage.Outputs[relPath]; ok {
			return stagePath, art, nil
		}
		// Search for an Artifact whose Path is any directory in the input's lineage.
		// For example: given "bish/bash/bosh/file.txt", look for "bish", then
		// "bish/bash", then "bish/bash/bosh".
		fullDir := filepath.Dir(relPath)
		parts := strings.Split(fullDir, string(filepath.Separator))
		dir := ""
		for _, part := range parts {
			dir := filepath.Join(dir, part)
			art, ok := en.Stage.Outputs[dir]
			// If we find a matching Artifact for any ancestor directory, the Stage in
			// question is only the owner if the Artifact is recursive, or
			// we've reached the immediate parent directory of the input.
			if ok && (art.IsRecursive || dir == fullDir) {
				return stagePath, art, nil
			}
		}
	}
	return "", &artifact.Artifact{}, nil
}
