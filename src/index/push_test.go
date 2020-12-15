package index

import (
	"io/ioutil"
	"log"

	"github.com/google/go-cmp/cmp"
	"github.com/kevin-hanselman/dud/src/artifact"
	"github.com/kevin-hanselman/dud/src/mocks"
	"github.com/kevin-hanselman/dud/src/stage"

	"testing"
)

func expectOutputsPushed(
	stg *stage.Stage,
	mockCache *mocks.Cache,
	rootDir,
	remote string,
) {
	for _, art := range stg.Outputs {
		mockCache.On("Push", rootDir, remote, *art).Return(nil).Once()
	}
}

func TestPush(t *testing.T) {

	rootDir := "project/root"
	remote := "my_remote:my_bucket"

	// TODO: Consider checking the logs instead of throwing them away.
	logger := log.New(ioutil.Discard, "", 0)

	t.Run("disjoint stages with oprhan dependency", func(t *testing.T) {
		stgA := stage.Stage{
			WorkingDir: "a",
			Dependencies: map[string]*artifact.Artifact{
				"bish.bin": {Path: "bish.bin"},
			},
			Outputs: map[string]*artifact.Artifact{
				"foo.bin": {Path: "foo.bin"},
			},
		}
		stgB := stage.Stage{
			WorkingDir: "b",
			Outputs: map[string]*artifact.Artifact{
				"bar.bin": {Path: "bar.bin"},
			},
		}
		idx := Index{
			"foo.yaml": &stgA,
			"bar.yaml": &stgB,
		}

		mockCache := mocks.Cache{}

		expectOutputsPushed(&stgA, &mockCache, rootDir, remote)

		pushed := make(map[string]bool)
		inProgress := make(map[string]bool)
		if err := idx.Push(
			"foo.yaml",
			&mockCache,
			rootDir,
			true,
			remote,
			pushed,
			inProgress,
			logger,
		); err != nil {
			t.Fatal(err)
		}

		mockCache.AssertExpectations(t)

		expectedCheckedOutSet := map[string]bool{
			"foo.yaml": true,
		}
		if diff := cmp.Diff(expectedCheckedOutSet, pushed); diff != "" {
			t.Fatalf("pushed -want +got:\n%s", diff)
		}
	})

	t.Run("two stages", func(t *testing.T) {
		// It's important to create identical copies rather than use the same
		// pointer in both Stages.
		linkedArtifactOrig := artifact.Artifact{Path: "foo.bin"}
		linkedArtifactA := linkedArtifactOrig
		linkedArtifactB := linkedArtifactOrig
		stgA := stage.Stage{
			Outputs: map[string]*artifact.Artifact{
				"foo.bin": &linkedArtifactA,
			},
		}
		stgB := stage.Stage{
			Dependencies: map[string]*artifact.Artifact{
				"foo.bin": &linkedArtifactB,
			},
			Outputs: map[string]*artifact.Artifact{
				"bar.bin": {Path: "bar.bin"},
			},
		}
		idx := Index{
			"foo.yaml": &stgA,
			"bar.yaml": &stgB,
		}

		mockCache := mocks.Cache{}

		expectOutputsPushed(&stgA, &mockCache, rootDir, remote)
		expectOutputsPushed(&stgB, &mockCache, rootDir, remote)

		pushed := make(map[string]bool)
		inProgress := make(map[string]bool)
		if err := idx.Push(
			"bar.yaml",
			&mockCache,
			rootDir,
			true,
			remote,
			pushed,
			inProgress,
			logger,
		); err != nil {
			t.Fatal(err)
		}

		// The linked Artifact should not be pushed as a dependency.
		linkedArtifactOrig.SkipCache = true
		mockCache.AssertNotCalled(t, "Push", rootDir, remote, &linkedArtifactOrig)

		mockCache.AssertExpectations(t)

		expectedCheckedOutSet := map[string]bool{
			"foo.yaml": true,
			"bar.yaml": true,
		}
		if diff := cmp.Diff(expectedCheckedOutSet, pushed); diff != "" {
			t.Fatalf("pushed -want +got:\n%s", diff)
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
		idx := Index{
			"bish.yaml": &stgA,
			"bash.yaml": &stgB,
			"bosh.yaml": &stgC,
		}

		mockCache := mocks.Cache{}

		expectOutputsPushed(&stgA, &mockCache, rootDir, remote)
		expectOutputsPushed(&stgB, &mockCache, rootDir, remote)
		expectOutputsPushed(&stgC, &mockCache, rootDir, remote)

		pushed := make(map[string]bool)
		inProgress := make(map[string]bool)
		if err := idx.Push(
			"bosh.yaml",
			&mockCache,
			rootDir,
			true,
			remote,
			pushed,
			inProgress,
			logger,
		); err != nil {
			t.Fatal(err)
		}

		mockCache.AssertExpectations(t)

		expectedCheckedOutSet := map[string]bool{
			"bish.yaml": true,
			"bash.yaml": true,
			"bosh.yaml": true,
		}
		if diff := cmp.Diff(expectedCheckedOutSet, pushed); diff != "" {
			t.Fatalf("pushed -want +got:\n%s", diff)
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
		idx := Index{
			"a.yaml": &stgA,
			"b.yaml": &stgB,
			"c.yaml": &stgC,
			"d.yaml": &stgD,
		}

		mockCache := mocks.Cache{}
		// Stage D is the only Stage that could possibly be checked out
		// successfully. We mock it to prevent a panic, but we don't enforce
		// that it must be called (due to random order).
		expectOutputsPushed(&stgD, &mockCache, rootDir, remote)

		pushed := make(map[string]bool)
		inProgress := make(map[string]bool)
		err := idx.Push(
			"c.yaml",
			&mockCache,
			rootDir,
			true,
			remote,
			pushed,
			inProgress,
			logger,
		)
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

	t.Run("can disable recursion", func(t *testing.T) {
		// It's important to create identical copies rather than use the same
		// pointer in both Stages.
		linkedArtifactOrig := artifact.Artifact{Path: "foo.bin"}
		linkedArtifactA := linkedArtifactOrig
		linkedArtifactB := linkedArtifactOrig
		stgA := stage.Stage{
			Outputs: map[string]*artifact.Artifact{
				"foo.bin": &linkedArtifactA,
			},
		}
		stgB := stage.Stage{
			Dependencies: map[string]*artifact.Artifact{
				"foo.bin": &linkedArtifactB,
			},
			Outputs: map[string]*artifact.Artifact{
				"bar.bin": {Path: "bar.bin"},
			},
		}
		idx := Index{
			"foo.yaml": &stgA,
			"bar.yaml": &stgB,
		}

		mockCache := mocks.Cache{}

		expectOutputsPushed(&stgB, &mockCache, rootDir, remote)

		pushed := make(map[string]bool)
		inProgress := make(map[string]bool)
		if err := idx.Push(
			"bar.yaml",
			&mockCache,
			rootDir,
			false,
			remote,
			pushed,
			inProgress,
			logger,
		); err != nil {
			t.Fatal(err)
		}

		// The linked Artifact should not be pushed as a dependency.
		linkedArtifactOrig.SkipCache = true
		mockCache.AssertNotCalled(t, "Push", stgB.WorkingDir, &linkedArtifactOrig, remote)

		mockCache.AssertExpectations(t)

		expectedCheckedOutSet := map[string]bool{
			"bar.yaml": true,
		}
		if diff := cmp.Diff(expectedCheckedOutSet, pushed); diff != "" {
			t.Fatalf("pushed -want +got:\n%s", diff)
		}
	})
}
