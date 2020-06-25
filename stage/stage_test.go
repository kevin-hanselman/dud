package stage

import (
	"github.com/google/go-cmp/cmp"
	"github.com/kevlar1818/duc/artifact"
	"github.com/kevlar1818/duc/fsutil"
	"github.com/kevlar1818/duc/strategy"
	"github.com/stretchr/testify/mock"
	"testing"
)

type mockCache struct {
	mock.Mock
}

func (c *mockCache) Commit(
	workingDir string,
	art *artifact.Artifact,
	strat strategy.CheckoutStrategy,
) error {
	args := c.Called(workingDir, art, strat)
	return args.Error(0)
}

func (c *mockCache) Checkout(
	workingDir string,
	art *artifact.Artifact,
	strat strategy.CheckoutStrategy,
) error {
	args := c.Called(workingDir, art, strat)
	return args.Error(0)
}

func (c *mockCache) PathForChecksum(checksum string) (string, error) {
	args := c.Called(checksum)
	return args.String(0), args.Error(1)
}

func (c *mockCache) Status(workingDir string, art artifact.Artifact) (artifact.Status, error) {
	args := c.Called(workingDir, art)
	return args.Get(0).(artifact.Status), args.Error(1)
}

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

	cache := mockCache{}
	for i := range stg.Outputs {
		cache.On("Commit", "workDir", &stg.Outputs[i], strat).Return(nil)
	}

	if err := stg.Commit(&cache, strat); err != nil {
		t.Fatal(err)
	}

	cache.AssertExpectations(t)

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

	cache := mockCache{}
	for i := range stg.Outputs {
		cache.On("Checkout", "workDir", &stg.Outputs[i], strat).Return(nil)
	}

	if err := stg.Checkout(&cache, strat); err != nil {
		t.Fatal(err)
	}

	cache.AssertExpectations(t)
}

func TestStatus(t *testing.T) {
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

	artStatus := artifact.Status{
		WorkspaceFileStatus: fsutil.RegularFile,
		HasChecksum:         true,
		ChecksumInCache:     true,
		ContentsMatch:       true,
	}

	cache := mockCache{}
	for _, art := range stg.Outputs {
		cache.On("Status", "workDir", art).Return(artStatus, nil)
	}

	// TODO: check output (i.e. test stage.Status.String())
	_, err := stg.Status(&cache)
	if err != nil {
		t.Fatal(err)
	}

	cache.AssertExpectations(t)
}

func TestFilePathForLock(t *testing.T) {
	input := "foo/bar.yaml"
	want := "foo/bar.yaml.lock"
	got := FilePathForLock(input)
	if got != want {
		t.Fatalf("FilePathForLock(%#v) = %#v, want %#v", input, got, want)
	}
}
