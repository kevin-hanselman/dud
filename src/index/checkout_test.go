package index

import (
	"io/ioutil"
	"log"

	"github.com/google/go-cmp/cmp"
	"github.com/kevin-hanselman/dud/src/artifact"
	"github.com/kevin-hanselman/dud/src/mocks"
	"github.com/kevin-hanselman/dud/src/stage"
	"github.com/kevin-hanselman/dud/src/strategy"

	"testing"
)

func expectOutputsCheckedOut(
	stg *stage.Stage,
	mockCache *mocks.Cache,
	strat strategy.CheckoutStrategy,
) {
	for _, art := range stg.Outputs {
		mockCache.On("Checkout", stg.WorkingDir, art, strat).Return(nil).Once()
	}
}

func TestCheckout(t *testing.T) {

	strat := strategy.LinkStrategy
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
		idx := make(Index)
		idx["foo.yaml"] = &entry{Stage: stgA}
		idx["bar.yaml"] = &entry{Stage: stgB}

		mockCache := mocks.Cache{}

		expectOutputsCheckedOut(&stgA, &mockCache, strat)

		checkedOut := make(map[string]bool)
		inProgress := make(map[string]bool)
		if err := idx.Checkout("foo.yaml", &mockCache, strat, checkedOut, inProgress, logger); err != nil {
			t.Fatal(err)
		}

		mockCache.AssertExpectations(t)

		expectedCheckedOutSet := map[string]bool{
			"foo.yaml": true,
		}
		if diff := cmp.Diff(expectedCheckedOutSet, checkedOut); diff != "" {
			t.Fatalf("checkedOut -want +got:\n%s", diff)
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
		idx := make(Index)
		idx["foo.yaml"] = &entry{Stage: stgA}
		idx["bar.yaml"] = &entry{Stage: stgB}

		mockCache := mocks.Cache{}

		expectOutputsCheckedOut(&stgA, &mockCache, strat)
		expectOutputsCheckedOut(&stgB, &mockCache, strat)

		checkedOut := make(map[string]bool)
		inProgress := make(map[string]bool)
		if err := idx.Checkout("bar.yaml", &mockCache, strat, checkedOut, inProgress, logger); err != nil {
			t.Fatal(err)
		}

		// The linked Artifact should not be checkedOut as a dependency.
		linkedArtifactOrig.SkipCache = true
		mockCache.AssertNotCalled(t, "Checkout", stgB.WorkingDir, &linkedArtifactOrig, strat)

		mockCache.AssertExpectations(t)

		expectedCheckedOutSet := map[string]bool{
			"foo.yaml": true,
			"bar.yaml": true,
		}
		if diff := cmp.Diff(expectedCheckedOutSet, checkedOut); diff != "" {
			t.Fatalf("checkedOut -want +got:\n%s", diff)
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
		idx := make(Index)
		idx["bish.yaml"] = &entry{Stage: stgA}
		idx["bash.yaml"] = &entry{Stage: stgB}
		idx["bosh.yaml"] = &entry{Stage: stgC}

		mockCache := mocks.Cache{}

		expectOutputsCheckedOut(&stgA, &mockCache, strat)
		expectOutputsCheckedOut(&stgB, &mockCache, strat)
		expectOutputsCheckedOut(&stgC, &mockCache, strat)

		checkedOut := make(map[string]bool)
		inProgress := make(map[string]bool)
		if err := idx.Checkout("bosh.yaml", &mockCache, strat, checkedOut, inProgress, logger); err != nil {
			t.Fatal(err)
		}

		mockCache.AssertExpectations(t)

		expectedCheckedOutSet := map[string]bool{
			"bish.yaml": true,
			"bash.yaml": true,
			"bosh.yaml": true,
		}
		if diff := cmp.Diff(expectedCheckedOutSet, checkedOut); diff != "" {
			t.Fatalf("checkedOut -want +got:\n%s", diff)
		}
	})

	t.Run("handle relative paths to other work dirs", func(t *testing.T) {
		// It's important to create identical copies rather than use the same
		// pointer in both Stages.
		linkedArtifactOrig := artifact.Artifact{Path: "foo.bin"}
		linkedArtifactA := linkedArtifactOrig
		stgA := stage.Stage{
			Outputs: map[string]*artifact.Artifact{
				"foo.bin": &linkedArtifactA,
			},
		}
		linkedArtifactB := linkedArtifactOrig
		linkedArtifactB.Path = "../foo.bin"
		stgB := stage.Stage{
			WorkingDir: "workDir",
			Dependencies: map[string]*artifact.Artifact{
				"../foo.bin": &linkedArtifactB,
			},
			Outputs: map[string]*artifact.Artifact{
				"bar.bin": {Path: "bar.bin"},
			},
		}
		idx := make(Index)
		idx["foo.yaml"] = &entry{Stage: stgA}
		idx["bar.yaml"] = &entry{Stage: stgB}

		mockCache := mocks.Cache{}

		expectOutputsCheckedOut(&stgA, &mockCache, strat)
		expectOutputsCheckedOut(&stgB, &mockCache, strat)

		checkedOut := make(map[string]bool)
		inProgress := make(map[string]bool)
		if err := idx.Checkout("bar.yaml", &mockCache, strat, checkedOut, inProgress, logger); err != nil {
			t.Fatal(err)
		}

		// The linked Artifact should not be checkedOut as a dependency.
		linkedArtifactOrig.SkipCache = true
		mockCache.AssertNotCalled(t, "Checkout", stgB.WorkingDir, &linkedArtifactOrig, strat)

		mockCache.AssertExpectations(t)

		expectedCheckedOutSet := map[string]bool{
			"foo.yaml": true,
			"bar.yaml": true,
		}
		if diff := cmp.Diff(expectedCheckedOutSet, checkedOut); diff != "" {
			t.Fatalf("checkedOut -want +got:\n%s", diff)
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
		idx["a.yaml"] = &entry{Stage: stgA}
		idx["b.yaml"] = &entry{Stage: stgB}
		idx["c.yaml"] = &entry{Stage: stgC}
		idx["d.yaml"] = &entry{Stage: stgD}

		mockCache := mocks.Cache{}
		// Stage D is the only Stage that could possibly be checked out
		// successfully. We mock it to prevent a panic, but we don't enforce
		// that it must be called (due to random order).
		expectOutputsCheckedOut(&stgD, &mockCache, strat)

		checkedOut := make(map[string]bool)
		inProgress := make(map[string]bool)
		err := idx.Checkout("c.yaml", &mockCache, strat, checkedOut, inProgress, logger)
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
}
