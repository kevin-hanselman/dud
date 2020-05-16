package index

import (
	"github.com/kevlar1818/duc/fsutil"
	"github.com/kevlar1818/duc/stage"
)

// An Index holds an exhaustive set of Stages for a repository.
type Index struct {
	StageFiles map[string]bool
}

var fromFile = fsutil.FromYamlFile

// NewIndex initializers a new Index
func NewIndex() *Index {
	idx := new(Index)
	idx.StageFiles = make(map[string]bool)
	return idx
}

// Add adds a Stage at the given path to the Index. Add returns an error if the
// path is invalid.
func (idx *Index) Add(path string) error {
	stg := new(stage.Stage)
	if err := fromFile(path, stg); err != nil {
		return err
	}
	idx.StageFiles[path] = true
	return nil
}
