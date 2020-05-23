package stage

import (
	"github.com/google/go-cmp/cmp"
	"github.com/kevlar1818/duc/artifact"
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
	s := Stage{
		Checksum:   "",
		WorkingDir: "foo",
		Outputs: []artifact.Artifact{
			{
				Checksum: "abc",
				Path:     "bar.txt",
			},
		},
	}

	if err := s.UpdateChecksum(); err != nil {
		t.Fatal(err)
	}

	if s.Checksum == "" {
		t.Fatal("stage.SetChecksum() didn't change (empty) checksum")
	}

	expected := s

	s.Checksum = "this should not affect the checksum"

	if err := s.UpdateChecksum(); err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(expected, s); diff != "" {
		t.Fatalf("checksum.Update(*Stage) -want +got:\n%s", diff)
	}

	origChecksum := s.Checksum
	s.WorkingDir = "this should affect the checksum"

	if err := s.UpdateChecksum(); err != nil {
		t.Fatal(err)
	}

	if s.Checksum == origChecksum {
		t.Fatal("changing stage.WorkingDir should have affected checksum")
	}

	origChecksum = s.Checksum
	s.Outputs[0].Path = "cat.png"

	if err := s.UpdateChecksum(); err != nil {
		t.Fatal(err)
	}
	if s.Checksum == origChecksum {
		t.Fatal("changing stage.Outputs should have affected checksum")
	}
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
		WorkspaceFileStatus: artifact.RegularFile,
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
