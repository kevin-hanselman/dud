package index

import (
	"os"
	"testing"

	"github.com/kevlar1818/duc/stage"
)

func TestAdd(t *testing.T) {

	stageFromFileOrig := stage.FromFile
	stage.FromFile = func(path string) (stage.Stage, bool, error) {
		return stage.Stage{}, false, nil
	}
	defer func() { stage.FromFile = stageFromFileOrig }()

	t.Run("add new stage", func(t *testing.T) {
		idx := make(Index)
		path := "foo/bar.duc"

		if err := idx.AddStagesFromPaths(path); err != nil {
			t.Fatal(err)
		}

		entry, added := idx[path]
		if !added {
			t.Fatal("path wasn't added to the index")
		}
		if !entry.ToCommit {
			t.Fatal("path wasn't added to commit list")
		}
	})

	t.Run("add already tracked stage", func(t *testing.T) {
		idx := make(Index)
		path := "foo/bar.duc"

		var stg stage.Stage
		idx[path] = &entry{ToCommit: false, Stage: stg}

		if err := idx.AddStagesFromPaths(path); err != nil {
			t.Fatal(err)
		}

		entry, added := idx[path]
		if !added {
			t.Fatal("path wasn't added to the index")
		}
		if !entry.ToCommit {
			t.Fatal("path wasn't added to commit list")
		}
	})

	t.Run("error if invalid stage", func(t *testing.T) {
		idx := make(Index)
		path := "foo/bar.duc"

		stage.FromFile = func(path string) (stage.Stage, bool, error) {
			return stage.Stage{}, false, os.ErrNotExist
		}

		err := idx.AddStagesFromPaths(path)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
