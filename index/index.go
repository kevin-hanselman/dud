package index

import (
	"github.com/kevlar1818/duc/fsutil"
	"github.com/kevlar1818/duc/stage"
)

// An Index holds an exhaustive set of Stages for a repository.
type Index struct {
	StageFiles map[string]bool
}

// FromFile is the function used when reading a file
var FromFile = fsutil.FromYamlFile

// New initializes a new Index
func New() *Index {
	idx := new(Index)
	idx.StageFiles = make(map[string]bool)
	return idx
}

// Add adds a Stage at the given path to the Index. Add returns an error if the
// path is invalid.
func (idx *Index) Add(path string) error {
	stg := new(stage.Stage)
	if err := FromFile(path, stg); err != nil {
		return err
	}

	idx.StageFiles[path] = true
	return nil
}

// CommitSet returns a set of stages marked for commit
func (idx *Index) CommitSet() map[string]bool {
	commitSet := make(map[string]bool)
	for path, inCommitSet := range idx.StageFiles {
		if inCommitSet {
			commitSet[path] = true
		}
	}
	return commitSet
}

// ClearCommitSet unmarks all stages for commit
func (idx *Index) ClearCommitSet() {
	for path := range idx.StageFiles {
		idx.StageFiles[path] = false
	}
}
