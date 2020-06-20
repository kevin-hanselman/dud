package index

import (
	"github.com/google/go-cmp/cmp"
	"github.com/kevlar1818/duc/stage"
	"os"
	"testing"
)

func TestAdd(t *testing.T) {

	fromFileOrig := FromFile
	FromFile = func(path string, v interface{}) error {
		stg, ok := v.(*stage.Stage)
		if ok {
			stg.WorkingDir = "."
		}
		return nil
	}
	defer func() { FromFile = fromFileOrig }()

	t.Run("add new stage", func(t *testing.T) {
		idx := make(Index)
		path := "foo/bar.duc"

		if err := idx.Add(path); err != nil {
			t.Fatal(err)
		}

		inCommitSet, added := idx[path]
		if !added {
			t.Fatal("path wasn't added to the index")
		}
		if !inCommitSet {
			t.Fatal("path wasn't added to commit list")
		}
	})

	t.Run("add already tracked stage", func(t *testing.T) {
		idx := make(Index)
		path := "foo/bar.duc"

		idx[path] = false

		if err := idx.Add(path); err != nil {
			t.Fatal(err)
		}

		inCommitSet, added := idx[path]
		if !added {
			t.Fatal("path wasn't added to the index")
		}
		if !inCommitSet {
			t.Fatal("path wasn't added to commit list")
		}
	})

	t.Run("error if invalid stage", func(t *testing.T) {
		idx := make(Index)
		path := "foo/bar.duc"

		FromFile = func(path string, v interface{}) error {
			return os.ErrNotExist
		}

		err := idx.Add(path)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestCommitSet(t *testing.T) {

	t.Run("get commit set", func(t *testing.T) {
		idx := Index{
			"a.duc": false,
			"b.duc": true,
			"c.duc": true,
			"d.duc": false,
		}

		actualCommitSet := idx.CommitSet()

		expectedCommitSet := map[string]bool{
			"b.duc": true,
			"c.duc": true,
		}

		if diff := cmp.Diff(expectedCommitSet, actualCommitSet); diff != "" {
			t.Fatalf("CommitSet() -want +got:\n%s", diff)
		}
	})

	t.Run("clear commit set", func(t *testing.T) {
		idx := Index{
			"a.duc": false,
			"b.duc": true,
			"c.duc": true,
			"d.duc": false,
		}

		idx.ClearCommitSet()

		expectedIndex := Index{
			"a.duc": false,
			"b.duc": false,
			"c.duc": false,
			"d.duc": false,
		}

		if diff := cmp.Diff(expectedIndex, idx); diff != "" {
			t.Fatalf("ClearCommitSet() -want +got:\n%s", diff)
		}
	})
}
