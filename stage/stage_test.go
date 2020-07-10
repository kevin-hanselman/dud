package stage

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kevlar1818/duc/artifact"
	"github.com/kevlar1818/duc/fsutil"
	"github.com/kevlar1818/duc/mocks"
	"github.com/kevlar1818/duc/strategy"
)

func TestSetChecksum(t *testing.T) {

	newStage := func() Stage {
		return Stage{
			Checksum:   "",
			WorkingDir: "dir",
			Outputs: []artifact.Artifact{
				{Path: "foo.txt"},
				{Path: "bar.txt"},
			},
		}
	}

	t.Run("checksum gets populated", func(t *testing.T) {
		stg := newStage()
		if err := stg.UpdateChecksum(); err != nil {
			t.Fatal(err)
		}
		if stg.Checksum == "" {
			t.Fatal("stage.SetChecksum() didn't change (empty) checksum")
		}
	})

	t.Run("stage checksum should not affect checksum", func(t *testing.T) {
		stg := newStage()
		if err := stg.UpdateChecksum(); err != nil {
			t.Fatal(err)
		}
		expected := stg

		stg.Checksum = "this should not affect the checksum"

		if err := stg.UpdateChecksum(); err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(expected, stg); diff != "" {
			t.Fatalf("checksum.Update(*Stage) -want +got:\n%s", diff)
		}
	})

	t.Run("stage workdir should affect checksum", func(t *testing.T) {
		stg := newStage()
		if err := stg.UpdateChecksum(); err != nil {
			t.Fatal(err)
		}

		origChecksum := stg.Checksum
		stg.WorkingDir = "this/should/affect/the/checksum"

		if err := stg.UpdateChecksum(); err != nil {
			t.Fatal(err)
		}

		if stg.Checksum == origChecksum {
			t.Fatal("changing stage.WorkingDir should have affected checksum")
		}
	})

	t.Run("output artifact path should affect checksum", func(t *testing.T) {
		stg := newStage()
		if err := stg.UpdateChecksum(); err != nil {
			t.Fatal(err)
		}

		origChecksum := stg.Checksum
		stg.Outputs[0].Path = "cat.png"

		if err := stg.UpdateChecksum(); err != nil {
			t.Fatal(err)
		}
		if stg.Checksum == origChecksum {
			t.Fatal("changing stage.Outputs should have affected checksum")
		}
	})

	t.Run("output artifact checksums should not affect checksum", func(t *testing.T) {
		stg := newStage()
		if err := stg.UpdateChecksum(); err != nil {
			t.Fatal(err)
		}

		expected := stg
		stg.Outputs[0].Checksum = "123456789"

		if err := stg.UpdateChecksum(); err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(expected, stg); diff != "" {
			t.Fatalf("checksum.Update(*Stage) -want +got:\n%s", diff)
		}
	})
}

func TestCommit(t *testing.T) {
	t.Run("Copy", func(t *testing.T) { testCommit(strategy.CopyStrategy, t) })
	t.Run("Link", func(t *testing.T) { testCommit(strategy.LinkStrategy, t) })
}

func testCommit(strat strategy.CheckoutStrategy, t *testing.T) {
	stg := Stage{
		Checksum:   "",
		WorkingDir: "workDir",
		Outputs: []artifact.Artifact{
			{
				Checksum: "",
				Path:     "foo.txt",
			},
			{
				Checksum: "",
				Path:     "bar.txt",
			},
		},
	}

	mockCache := mocks.Cache{}
	for i := range stg.Outputs {
		mockCache.On("Commit", "workDir", &stg.Outputs[i], strat).Return(nil)
	}

	if err := stg.Commit(&mockCache, strat); err != nil {
		t.Fatal(err)
	}

	mockCache.AssertExpectations(t)

	if stg.Checksum == "" {
		t.Error("expected stage checksum to be set")
	}
}

func TestCheckout(t *testing.T) {
	t.Run("Copy", func(t *testing.T) { testCheckout(strategy.CopyStrategy, t) })
	t.Run("Link", func(t *testing.T) { testCheckout(strategy.LinkStrategy, t) })
}

func testCheckout(strat strategy.CheckoutStrategy, t *testing.T) {
	// TODO: Test stage checksum? What does it mean if the checksum is invalid/empty?
	stg := Stage{
		Checksum:   "",
		WorkingDir: "workDir",
		Outputs: []artifact.Artifact{
			{
				Checksum: "",
				Path:     "foo.txt",
			},
			{
				Checksum: "",
				Path:     "bar.txt",
			},
		},
	}

	mockCache := mocks.Cache{}
	for i := range stg.Outputs {
		mockCache.On("Checkout", "workDir", &stg.Outputs[i], strat).Return(nil)
	}

	if err := stg.Checkout(&mockCache, strat); err != nil {
		t.Fatal(err)
	}

	mockCache.AssertExpectations(t)
}

func TestStatus(t *testing.T) {
	stg := Stage{
		Checksum:   "",
		WorkingDir: "workDir",
		Outputs: []artifact.Artifact{
			{Path: "foo.txt"},
			{Path: "bar.txt"},
		},
	}

	artStatus := artifact.Status{
		WorkspaceFileStatus: fsutil.RegularFile,
		HasChecksum:         true,
		ChecksumInCache:     true,
		ContentsMatch:       true,
	}

	mockCache := mocks.Cache{}
	expectedStageStatus := Status{}
	for _, art := range stg.Outputs {
		mockCache.On("Status", "workDir", art).Return(artStatus, nil)
		expectedStageStatus[art.Path] = artifact.ArtifactWithStatus{
			Artifact: art,
			Status:   artStatus,
		}
	}

	stageStatus, err := stg.Status(&mockCache)
	if err != nil {
		t.Fatal(err)
	}

	mockCache.AssertExpectations(t)

	if diff := cmp.Diff(expectedStageStatus, stageStatus); diff != "" {
		t.Fatalf("stage.Status() -want +got:\n%s", diff)
	}

}

func TestFilePathForLock(t *testing.T) {
	input := "foo/bar.yaml"
	want := "foo/bar.yaml.lock"
	got := FilePathForLock(input)
	if got != want {
		t.Fatalf("FilePathForLock(%#v) = %#v, want %#v", input, got, want)
	}
}
