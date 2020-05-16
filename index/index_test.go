package index

import (
	"os"
	"testing"
)

func TestAdd(t *testing.T) {

	fromFileOrig := fromFile
	fromFile = func(path string, v interface{}) error {
		return nil
	}
	defer func() { fromFile = fromFileOrig }()

	t.Run("add new stage", func(t *testing.T) {
		idx := NewIndex()
		path := "foo/bar.duc"

		if err := idx.Add(path); err != nil {
			t.Fatal(err)
		}

		onCommitList, added := idx.StageFiles[path]
		if !added {
			t.Fatal("path wasn't added to the index")
		}
		if !onCommitList {
			t.Fatal("path wasn't added to commit list")
		}
	})

	t.Run("add already tracked stage", func(t *testing.T) {
		idx := NewIndex()
		path := "foo/bar.duc"

		idx.StageFiles[path] = false

		if err := idx.Add(path); err != nil {
			t.Fatal(err)
		}

		onCommitList, added := idx.StageFiles[path]
		if !added {
			t.Fatal("path wasn't added to the index")
		}
		if !onCommitList {
			t.Fatal("path wasn't added to commit list")
		}
	})

	t.Run("error if invalid stage", func(t *testing.T) {
		idx := NewIndex()
		path := "foo/bar.duc"

		fromFile = func(path string, v interface{}) error {
			return os.ErrNotExist
		}

		err := idx.Add(path)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
