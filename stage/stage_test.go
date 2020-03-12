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

func (c *mockCache) Commit(workingDir string, art *artifact.Artifact, strat strategy.CheckoutStrategy) error {
	args := c.Called(workingDir, art, strat)
	return args.Error(0)
}

func (c *mockCache) Checkout(workingDir string, art *artifact.Artifact, strat strategy.CheckoutStrategy) error {
	args := c.Called(workingDir, art, strat)
	return args.Error(0)
}

func (c *mockCache) CachePathForArtifact(art artifact.Artifact) (string, error) {
	args := c.Called(art)
	return args.String(0), args.Error(1)
}

func (c *mockCache) Status(workingDir string, art artifact.Artifact, strat strategy.CheckoutStrategy) (string, error) {
	return "up to date", nil
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

	s.SetChecksum()

	if s.Checksum == "" {
		t.Fatal("stage.SetChecksum() didn't change (empty) checksum")
	}

	expected := s

	s.Checksum = "this should not affect the checksum"

	s.SetChecksum()

	if diff := cmp.Diff(expected, s); diff != "" {
		t.Fatalf("stage.SetChecksum() -want +got:\n%s", diff)
	}

	origChecksum := s.Checksum
	s.WorkingDir = "this should affect the checksum"

	s.SetChecksum()

	if s.Checksum == origChecksum {
		t.Fatal("changing stage.WorkingDir should have affected checksum")
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

	stg.Commit(&cache, strat)

	cache.AssertExpectations(t)
}

func TestCheckout(t *testing.T) {
	t.Run("Copy", func(t *testing.T) { testCheckout(strategy.CopyStrategy, t) })
	t.Run("Link", func(t *testing.T) { testCheckout(strategy.LinkStrategy, t) })
}

func testCheckout(strat strategy.CheckoutStrategy, t *testing.T) {
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

	stg.Checkout(&cache, strat)

	cache.AssertExpectations(t)
}
