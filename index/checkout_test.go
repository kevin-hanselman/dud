package index

import (
	"github.com/google/go-cmp/cmp"
	"github.com/kevin-hanselman/duc/artifact"
	"github.com/kevin-hanselman/duc/mocks"
	"github.com/kevin-hanselman/duc/stage"
	"github.com/kevin-hanselman/duc/strategy"

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
		if err := idx.Checkout("foo.yaml", &mockCache, strat, checkedOut); err != nil {
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
		if err := idx.Checkout("bar.yaml", &mockCache, strat, checkedOut); err != nil {
			t.Fatal(err)
		}

		// The linked Artifact should not be checkedOut as a dependency.
		linkedArtifactOrig.SkipCache = true
		mockCache.AssertNotCalled(t, "Checkout", stgB.WorkingDir, &linkedArtifactOrig, strat)

		mockCache.AssertExpectations(t)

		if diff := cmp.Diff(linkedArtifactA, linkedArtifactB); diff != "" {
			t.Fatalf("Checkout -artifactA +artifactB:\n%s", diff)
		}

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
		if err := idx.Checkout("bosh.yaml", &mockCache, strat, checkedOut); err != nil {
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
}
