package stage

import (
	"github.com/google/go-cmp/cmp"
	"github.com/kevlar1818/duc/artifact"
	"github.com/kevlar1818/duc/checksum"
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

func (c *mockCache) Status(workingDir string, art artifact.Artifact) (artifact.Status, error) {
	return artifact.Status{}, nil
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

	if err := checksum.Update(&s); err != nil {
		t.Fatal(err)
	}

	if s.Checksum == "" {
		t.Fatal("stage.SetChecksum() didn't change (empty) checksum")
	}

	expected := s

	s.Checksum = "this should not affect the checksum"

	if err := checksum.Update(&s); err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(expected, s); diff != "" {
		t.Fatalf("checksum.Update(*Stage) -want +got:\n%s", diff)
	}

	origChecksum := s.Checksum
	s.WorkingDir = "this should affect the checksum"

	if err := checksum.Update(&s); err != nil {
		t.Fatal(err)
	}

	if s.Checksum == origChecksum {
		t.Fatal("changing stage.WorkingDir should have affected checksum")
	}

	origChecksum = s.Checksum
	s.Outputs[0].Path = "cat.png"

	if err := checksum.Update(&s); err != nil {
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

	if err := stg.Checkout(&cache, strat); err != nil {
		t.Fatal(err)
	}

	cache.AssertExpectations(t)
}
