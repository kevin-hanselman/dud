package index

import (
	"github.com/google/go-cmp/cmp"

	"github.com/kevin-hanselman/duc/artifact"
	"github.com/kevin-hanselman/duc/fsutil"
	"github.com/kevin-hanselman/duc/mocks"
	"github.com/kevin-hanselman/duc/stage"

	"testing"
)

func expectStageStatusCalled(
	stg *stage.Stage,
	mockCache *mocks.Cache,
	artStatus artifact.Status,
) stage.Status {
	stageStatus := make(stage.Status)
	for artPath, art := range stg.Outputs {
		stageStatus[artPath] = artifact.ArtifactWithStatus{
			Artifact: *art,
			Status:   artStatus,
		}
		mockCache.On("Status", stg.WorkingDir, *art).Return(artStatus, nil)
	}
	return stageStatus
}

func TestStatus(t *testing.T) {

	t.Run("single stage", func(t *testing.T) {
		stgA := stage.Stage{
			Outputs: map[string]*artifact.Artifact{
				"foo.bin": {Path: "foo.bin"},
			},
		}
		stgB := stage.Stage{
			Outputs: map[string]*artifact.Artifact{
				"bar.bin": {Path: "bar.bin"},
			},
		}
		idx := make(Index)
		idx["foo.yaml"] = &entry{Stage: stgA}
		idx["bar.yaml"] = &entry{Stage: stgB}

		mockCache := mocks.Cache{}

		artStatus := artifact.Status{
			WorkspaceFileStatus: fsutil.Link,
			HasChecksum:         true,
			ChecksumInCache:     true,
			ContentsMatch:       true,
		}

		expectedStatus := make(Status)
		expectedStatus["foo.yaml"] = expectStageStatusCalled(&stgA, &mockCache, artStatus)

		outputStatus := make(Status)
		err := idx.Status("foo.yaml", &mockCache, outputStatus)
		if err != nil {
			t.Fatal(err)
		}

		mockCache.AssertExpectations(t)
		if diff := cmp.Diff(expectedStatus, outputStatus); diff != "" {
			t.Fatalf("Stage -want +got:\n%s", diff)
		}
	})

	t.Run("two stages", func(t *testing.T) {
		stgA := stage.Stage{
			Outputs: map[string]*artifact.Artifact{
				"foo.bin": {Path: "foo.bin"},
			},
		}
		stgB := stage.Stage{
			Dependencies: map[string]*artifact.Artifact{
				"foo.bin": {Path: "foo.bin"},
			},
			Outputs: map[string]*artifact.Artifact{
				"bar.bin": {Path: "bar.bin"},
			},
		}

		mockCache := mocks.Cache{}

		artStatus := artifact.Status{
			WorkspaceFileStatus: fsutil.Link,
			HasChecksum:         true,
			ChecksumInCache:     true,
			ContentsMatch:       true,
		}

		expectedStatus := make(Status)
		expectedStatus["foo.yaml"] = expectStageStatusCalled(&stgA, &mockCache, artStatus)
		expectedStatus["bar.yaml"] = expectStageStatusCalled(&stgB, &mockCache, artStatus)

		idx := make(Index)
		idx["foo.yaml"] = &entry{Stage: stgA}
		idx["bar.yaml"] = &entry{Stage: stgB}

		outputStatus := make(Status)
		err := idx.Status("bar.yaml", &mockCache, outputStatus)
		if err != nil {
			t.Fatal(err)
		}

		mockCache.AssertExpectations(t)
		if diff := cmp.Diff(expectedStatus, outputStatus); diff != "" {
			t.Fatalf("Stage -want +got:\n%s", diff)
		}
	})

	// TODO list:
	// * make output ordered by recursive call ordering to aid interpretability?
	// * stop at first out-of-date dep? might be unintuitive/unhelpful
}
