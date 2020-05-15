package index

import (
	//"github.com/stretchr/testify/mock"
	"github.com/kevlar1818/duc/mocks"
	"testing"
)

func TestAdd(t *testing.T) {

	t.Run("basic add", func(t *testing.T) {
		idx := NewIndex()
		stg := new(mocks.Stager)
		path := "foo/bar.duc"

		stg.On("FromFile", path).Return(nil)

		idx.Add(path, stg)

		stg.AssertExpectations(t)

		onCommitList, added := idx.StageFiles[path]
		if !added {
			t.Fatal("path wasn't added")
		}
		if !onCommitList {
			t.Fatal("path wasn't added to commit list")
		}
	})
}
