package index

import (
	"github.com/kevlar1818/duc/fsutil"
	"github.com/kevlar1818/duc/stage"
)

type entry struct {
	ToCommit bool
	Stage    stage.Stage
}

// fileFormat stores Stage paths and their ToCommit status
type fileFormat map[string]bool

// An Index holds an exhaustive set of Stages for a repository.
// TODO: Not threadsafe
type Index map[string]*entry

var fromYamlFile = fsutil.FromYamlFile

// Add adds a Stage at the given path to the Index. Add returns an error if the
// path is invalid.
func (idx *Index) Add(paths ...string) error {
	for _, path := range paths {
		stg := new(stage.Stage)
		// TODO: create and use function to pick between stage and lock file
		if err := fromYamlFile(path, stg); err != nil {
			return err
		}
		(*idx)[path] = &entry{ToCommit: true, Stage: *stg}
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
func FromFile(path string) (idx Index, err error) {
	indexFile := make(fileFormat)
	if err = fromYamlFile(path, indexFile); err != nil {
		return
	}
	idx = make(Index)
	for path, toCommit := range indexFile {
		stg := new(stage.Stage)

		// Try to load the lock file first.
		lockPath := stage.FilePathForLock(path)

		var lockPathExists bool
		lockPathExists, err = fsutil.Exists(lockPath, false)
		if err != nil {
			return
		}
		if lockPathExists {
			path = lockPath
		}

		if err = fromYamlFile(path, stg); err != nil {
			return
		}
		idx[path] = &entry{ToCommit: toCommit, Stage: *stg}
	}
	return
}
