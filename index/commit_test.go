package index

import (
	"github.com/google/go-cmp/cmp"
	"github.com/kevin-hanselman/duc/artifact"
	"github.com/kevin-hanselman/duc/mocks"
	"github.com/kevin-hanselman/duc/stage"
	"github.com/kevin-hanselman/duc/strategy"

	"testing"
)

func mockCommit(workDir string, art *artifact.Artifact, strat strategy.CheckoutStrategy) error {
	art.Checksum = "committed"
	return nil
}

func expectOutputsCommitted(
	stg *stage.Stage,
	mockCache *mocks.Cache,
	strat strategy.CheckoutStrategy,
) {
	for _, art := range stg.Outputs {
		mockCache.On("Commit", stg.WorkingDir, art, strat).Return(mockCommit).Once()
	}
}

func TestCommit(t *testing.T) {

	strat := strategy.LinkStrategy

	t.Run("disjoint stages with oprhan dependency", func(t *testing.T) {
		orphanArt := artifact.Artifact{Path: "bish.bin"}
		stgA := stage.Stage{
			WorkingDir: "a",
			Dependencies: map[string]*artifact.Artifact{
				"bish.bin": &orphanArt,
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

		expectOutputsCommitted(&stgA, &mockCache, strat)

		orphanCopy := orphanArt
		orphanCopy.SkipCache = true
		mockCache.On("Commit", stgA.WorkingDir, &orphanCopy, strat).Return(mockCommit).Once()

		committed := make(map[string]bool)
		if err := idx.Commit("foo.yaml", &mockCache, strat, committed); err != nil {
			t.Fatal(err)
		}

		mockCache.AssertExpectations(t)

		if stgA.Outputs["foo.bin"].Checksum == "" {
			t.Fatal("expected output Artifact to have Checksum set")
		}
		if stgA.Dependencies["bish.bin"].Checksum == "" {
			t.Fatal("expected dependency Artifact to have Checksum set")
		}

		expectedCommitSet := map[string]bool{
			"foo.yaml": true,
		}
		if diff := cmp.Diff(expectedCommitSet, committed); diff != "" {
			t.Fatalf("committed -want +got:\n%s", diff)
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

		expectOutputsCommitted(&stgA, &mockCache, strat)
		expectOutputsCommitted(&stgB, &mockCache, strat)

		committed := make(map[string]bool)
		if err := idx.Commit("bar.yaml", &mockCache, strat, committed); err != nil {
			t.Fatal(err)
		}

		// The linked Artifact should not be committed as a dependency.
		linkedArtifactOrig.SkipCache = true
		mockCache.AssertNotCalled(t, "Commit", stgB.WorkingDir, &linkedArtifactOrig, strat)

		mockCache.AssertExpectations(t)

		// Both instances of the linked Artifact should be committed and the same.
		if linkedArtifactA.Checksum == "" {
			t.Fatalf("Expected artifact %v has empty checksum", linkedArtifactA.Path)
		}

		if diff := cmp.Diff(linkedArtifactA, linkedArtifactB); diff != "" {
			t.Fatalf("Commit -artifactA +artifactB:\n%s", diff)
		}

		expectedCommitSet := map[string]bool{
			"foo.yaml": true,
			"bar.yaml": true,
		}
		if diff := cmp.Diff(expectedCommitSet, committed); diff != "" {
			t.Fatalf("committed -want +got:\n%s", diff)
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

		expectOutputsCommitted(&stgA, &mockCache, strat)
		expectOutputsCommitted(&stgB, &mockCache, strat)
		expectOutputsCommitted(&stgC, &mockCache, strat)

		committed := make(map[string]bool)
		if err := idx.Commit("bosh.yaml", &mockCache, strat, committed); err != nil {
			t.Fatal(err)
		}

		mockCache.AssertExpectations(t)

		expectedCommitSet := map[string]bool{
			"bish.yaml": true,
			"bash.yaml": true,
			"bosh.yaml": true,
		}
		if diff := cmp.Diff(expectedCommitSet, committed); diff != "" {
			t.Fatalf("committed -want +got:\n%s", diff)
		}
	})
}
