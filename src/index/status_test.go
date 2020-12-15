package index

import (
	"github.com/google/go-cmp/cmp"

	"github.com/kevin-hanselman/dud/src/artifact"
	"github.com/kevin-hanselman/dud/src/fsutil"
	"github.com/kevin-hanselman/dud/src/mocks"
	"github.com/kevin-hanselman/dud/src/stage"

	"testing"
)

func expectStageStatusCalled(
	stg *stage.Stage,
	mockCache *mocks.Cache,
	rootDir string,
	artStatus artifact.Status,
) stage.Status {
	stageStatus := make(stage.Status)
	for artPath, art := range stg.Outputs {
		stageStatus[artPath] = artifact.ArtifactWithStatus{
			Artifact: *art,
			Status:   artStatus,
		}
		mockCache.On("Status", rootDir, *art).Return(stageStatus[artPath], nil).Once()
	}
	return stageStatus
}

func TestStatus(t *testing.T) {

	upToDate := artifact.Status{
		WorkspaceFileStatus: fsutil.StatusLink,
		HasChecksum:         true,
		ChecksumInCache:     true,
		ContentsMatch:       true,
	}

	rootDir := "project/root"

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
		idx["foo.yaml"] = &stgA
		idx["bar.yaml"] = &stgB

		mockCache := mocks.Cache{}

		expectedStatus := make(Status)
		expectedStatus["foo.yaml"] = expectStageStatusCalled(&stgA, &mockCache, rootDir, upToDate)

		outputStatus := make(Status)
		inProgress := make(map[string]bool)
		err := idx.Status("foo.yaml", &mockCache, rootDir, outputStatus, inProgress)
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
		expectedStatus["foo.yaml"] = expectStageStatusCalled(&stgA, &mockCache, rootDir, upToDate)
		expectedStatus["bar.yaml"] = expectStageStatusCalled(&stgB, &mockCache, rootDir, upToDate)

		idx := make(Index)
		idx["foo.yaml"] = &stgA
		idx["bar.yaml"] = &stgB

		outputStatus := make(Status)
		inProgress := make(map[string]bool)
		err := idx.Status("bar.yaml", &mockCache, rootDir, outputStatus, inProgress)
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
				"a.bin": {Path: "a.bin"},
			},
		}
		stgB := stage.Stage{
			Dependencies: map[string]*artifact.Artifact{
				"a.bin": {Path: "a.bin"},
			},
			Outputs: map[string]*artifact.Artifact{
				"b.bin": {Path: "b.bin"},
			},
		}
		stgC := stage.Stage{
			Dependencies: map[string]*artifact.Artifact{
				"b.bin": {Path: "b.bin"},
				"a.bin": {Path: "a.bin"},
			},
			Outputs: map[string]*artifact.Artifact{
				"c.bin": {Path: "c.bin"},
			},
		}
		idx := make(Index)
		idx["a.yaml"] = &stgA
		idx["b.yaml"] = &stgB
		idx["c.yaml"] = &stgC

		mockCache := mocks.Cache{}

		expectedStatus := make(Status)
		expectedStatus["a.yaml"] = expectStageStatusCalled(&stgA, &mockCache, rootDir, upToDate)
		expectedStatus["b.yaml"] = expectStageStatusCalled(&stgB, &mockCache, rootDir, upToDate)
		expectedStatus["c.yaml"] = expectStageStatusCalled(&stgC, &mockCache, rootDir, upToDate)

		outputStatus := make(Status)
		inProgress := make(map[string]bool)
		err := idx.Status("c.yaml", &mockCache, rootDir, outputStatus, inProgress)
		if err != nil {
			t.Fatal(err)
		}

		mockCache.AssertExpectations(t)
		if diff := cmp.Diff(expectedStatus, outputStatus); diff != "" {
			t.Fatalf("Stage -want +got:\n%s", diff)
		}
	})

	t.Run("cycles are prevented", func(t *testing.T) {
		// stgA <-- stgB <-- stgC --> stgD
		//    |---------------^
		stgA := stage.Stage{
			Dependencies: map[string]*artifact.Artifact{
				"c.bin": {Path: "c.bin"},
			},
			Outputs: map[string]*artifact.Artifact{
				"a.bin": {Path: "a.bin"},
			},
		}
		stgB := stage.Stage{
			Dependencies: map[string]*artifact.Artifact{
				"a.bin": {Path: "a.bin"},
			},
			Outputs: map[string]*artifact.Artifact{
				"b.bin": {Path: "b.bin"},
			},
		}
		stgC := stage.Stage{
			Dependencies: map[string]*artifact.Artifact{
				"b.bin": {Path: "b.bin"},
				"d.bin": {Path: "d.bin"},
			},
			Outputs: map[string]*artifact.Artifact{
				"c.bin": {Path: "c.bin"},
			},
		}
		stgD := stage.Stage{
			Outputs: map[string]*artifact.Artifact{
				"d.bin": {Path: "d.bin"},
			},
		}
		idx := make(Index)
		idx["a.yaml"] = &stgA
		idx["b.yaml"] = &stgB
		idx["c.yaml"] = &stgC
		idx["d.yaml"] = &stgD

		mockCache := mocks.Cache{}
		// Stage D is the only Stage that could possibly be committed
		// successfully. We mock it to prevent a panic, but we don't
		// enforce that it must be called (due to random order).
		expectStageStatusCalled(&stgD, &mockCache, rootDir, upToDate)

		outputStatus := make(Status)
		inProgress := make(map[string]bool)
		err := idx.Status("c.yaml", &mockCache, rootDir, outputStatus, inProgress)
		if err == nil {
			t.Fatal("expected error")
		}

		expectedError := "cycle detected"
		if diff := cmp.Diff(expectedError, err.Error()); diff != "" {
			t.Fatalf("error -want +got:\n%s", diff)
		}

		expectedInProgress := map[string]bool{
			"c.yaml": true,
			"b.yaml": true,
			"a.yaml": true,
		}
		if diff := cmp.Diff(expectedInProgress, inProgress); diff != "" {
			t.Fatalf("inProgress -want +got:\n%s", diff)
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
		idx["foo.yaml"] = &stg

		mockCache := mocks.Cache{}

		expectedStatus := make(Status)
		expectedStageStatus := expectStageStatusCalled(&stg, &mockCache, rootDir, upToDate)

		orphanArtStatus := artifact.ArtifactWithStatus{
			Artifact: orphanArt,
			Status:   upToDate,
		}
		expectedStageStatus["bish.bin"] = orphanArtStatus
		expectedStatus["foo.yaml"] = expectedStageStatus

		mockCache.On("Status", rootDir, orphanArt).Return(orphanArtStatus, nil).Once()

		outputStatus := make(Status)
		inProgress := make(map[string]bool)
		err := idx.Status("foo.yaml", &mockCache, rootDir, outputStatus, inProgress)
		if err != nil {
			t.Fatal(err)
		}

		mockCache.AssertExpectations(t)
		if diff := cmp.Diff(expectedStatus, outputStatus); diff != "" {
			t.Fatalf("Stage -want +got:\n%s", diff)
		}
	})
}
