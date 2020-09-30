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
		mockCache.On("Status", stg.WorkingDir, *art).Return(stageStatus[artPath], nil).Once()
	}
	return stageStatus
}

func TestStatus(t *testing.T) {

	upToDate := artifact.Status{
		WorkspaceFileStatus: fsutil.Link,
		HasChecksum:         true,
		ChecksumInCache:     true,
		ContentsMatch:       true,
	}

	t.Run("disjoint stages", func(t *testing.T) {
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

		expectedStatus := make(Status)
		expectedStatus["foo.yaml"] = expectStageStatusCalled(&stgA, &mockCache, upToDate)

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

		expectedStatus := make(Status)
		expectedStatus["foo.yaml"] = expectStageStatusCalled(&stgA, &mockCache, upToDate)
		expectedStatus["bar.yaml"] = expectStageStatusCalled(&stgB, &mockCache, upToDate)

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

	t.Run("stages aren't repeated", func(t *testing.T) {
		// stgA <-- stgB <-- stgC
		//    ^---------------|
		stgA := stage.Stage{
			Outputs: map[string]*artifact.Artifact{
				"bish.bin": {Path: "bish.bin"},
			},
		}
		stgB := stage.Stage{
			Dependencies: map[string]*artifact.Artifact{
				"bish.bin": {Path: "bish.bin"},
			},
			Outputs: map[string]*artifact.Artifact{
				"bash.bin": {Path: "bash.bin"},
			},
		}
		stgC := stage.Stage{
			Dependencies: map[string]*artifact.Artifact{
				"bash.bin": {Path: "bash.bin"},
				"bish.bin": {Path: "bish.bin"},
			},
			Outputs: map[string]*artifact.Artifact{
				"bosh.bin": {Path: "bosh.bin"},
			},
		}
		mockCache := mocks.Cache{}

		expectedStatus := make(Status)
		expectedStatus["bish.yaml"] = expectStageStatusCalled(&stgA, &mockCache, upToDate)
		expectedStatus["bash.yaml"] = expectStageStatusCalled(&stgB, &mockCache, upToDate)
		expectedStatus["bosh.yaml"] = expectStageStatusCalled(&stgC, &mockCache, upToDate)

		idx := make(Index)
		idx["bish.yaml"] = &entry{Stage: stgA}
		idx["bash.yaml"] = &entry{Stage: stgB}
		idx["bosh.yaml"] = &entry{Stage: stgC}

		outputStatus := make(Status)
		err := idx.Status("bosh.yaml", &mockCache, outputStatus)
		if err != nil {
			t.Fatal(err)
		}

		mockCache.AssertExpectations(t)
		if diff := cmp.Diff(expectedStatus, outputStatus); diff != "" {
			t.Fatalf("Stage -want +got:\n%s", diff)
		}
	})

	t.Run("handle dependencies with no owner", func(t *testing.T) {
		orphanArt := artifact.Artifact{Path: "bish.bin"}
		stg := stage.Stage{
			Dependencies: map[string]*artifact.Artifact{
				"bish.bin": &orphanArt,
			},
			Outputs: map[string]*artifact.Artifact{
				"bash.bin": {Path: "bash.bin"},
			},
		}
		idx := make(Index)
		idx["foo.yaml"] = &entry{Stage: stg}

		mockCache := mocks.Cache{}

		expectedStatus := make(Status)
		expectedStageStatus := expectStageStatusCalled(&stg, &mockCache, upToDate)

		orphanArtStatus := artifact.ArtifactWithStatus{
			Artifact: orphanArt,
			Status:   upToDate,
		}
		expectedStageStatus["bish.bin"] = orphanArtStatus
		expectedStatus["foo.yaml"] = expectedStageStatus

		mockCache.On("Status", stg.WorkingDir, orphanArt).Return(orphanArtStatus, nil).Once()

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

	// TODO list:
	// * make output ordered by recursive call ordering to aid interpretability?
	// * stop at first out-of-date dep? might be unintuitive/unhelpful
	// * prevent multiple calls to cache.Status for unowned Artifacts?
	//   may require serious refactoring, or at least a sub-optimal search of the Status object
}
