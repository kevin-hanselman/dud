package index

import (
	"path/filepath"
	"strings"

	"github.com/kevin-hanselman/duc/fsutil"
	"github.com/kevin-hanselman/duc/stage"
)

type entry struct {
	// ToCommit is true if the stage is marked for commit
	ToCommit bool
	// IsLocked is true if the Stage in this entry is a locked version; i.e. it
	// has checksummed dependencies and outputs.
	IsLocked bool
	Stage    stage.Stage
}

// fileFormat stores Stage paths and their ToCommit status
// TODO: ToCommit should be separated from the Index file format at some
// point. The index should be tracked by Git, but the commit status should not
// be; it only makes sense on a per-clone basis.
type fileFormat map[string]bool

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
			ToCommit: true,
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
func (idx *Index) ToFile(path string) error {
	indexFile := make(fileFormat)
	for path, ent := range *idx {
		indexFile[path] = ent.ToCommit
	}
	return fsutil.ToYamlFile(path, indexFile)
}

// FromFile reads and returns an Index from the specified file path.
// See ToFile docs for more context.
// TODO no tests
// TODO Add a "full/bare" flag to enable only loading the fileFormat struct?
//      This would be a nice optimization not to load the whole Index when we
//      just want to check if a path is in the Index.
func FromFile(path string) (Index, error) {
	var idx Index
	indexFile := make(fileFormat)
	if err := fromYamlFile(path, indexFile); err != nil {
		return idx, err
	}
	idx = make(Index)
	for path, toCommit := range indexFile {
		stg, isLock, err := stage.FromFile(path)
		if err != nil {
			return idx, err
		}
		idx[path] = &entry{
			ToCommit: toCommit,
			IsLocked: isLock,
			Stage:    stg,
		}
	}
	return idx, nil
}

func (idx *Index) findOwner(path string) (*entry, bool, error) {
	for _, en := range *idx {
		relPath, err := filepath.Rel(en.Stage.WorkingDir, path)
		if err != nil {
			return nil, false, err
		}
		if _, ok := en.Stage.Outputs[relPath]; ok {
			return en, true, nil
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
				return en, true, nil
			}
		}
	}
	return nil, false, nil
}
