package index

import (
	"os"
	"testing"

	"github.com/kevlar1818/duc/stage"
)

func TestAdd(t *testing.T) {

	fromYamlFileOrig := fromYamlFile
	fromYamlFile = func(path string, v interface{}) error {
		stg, ok := v.(*stage.Stage)
		if ok {
			stg.WorkingDir = "."
		}
		return nil
	}
	defer func() { fromYamlFile = fromYamlFileOrig }()

	t.Run("add new stage", func(t *testing.T) {
		idx := make(Index)
		path := "foo/bar.duc"

		if err := idx.Add(path); err != nil {
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

		if err := idx.Add(path); err != nil {
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

		fromYamlFile = func(path string, v interface{}) error {
			return os.ErrNotExist
		}

		err := idx.Add(path)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
