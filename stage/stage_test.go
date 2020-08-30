package stage

import (
	"errors"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kevlar1818/duc/artifact"
	"github.com/kevlar1818/duc/mocks"
	"github.com/kevlar1818/duc/strategy"
)

func TestEquivalency(t *testing.T) {

	newStage := func() Stage {
		return Stage{
			WorkingDir: "dir",
			Outputs: []artifact.Artifact{
				{Path: "foo.txt"},
				{Path: "bar.txt"},
			},
			Dependencies: []artifact.Artifact{
				{Path: "b", IsDir: true},
			},
		}
	}

	t.Run("identical stages are equivalent", func(t *testing.T) {
		a, b := newStage(), newStage()
		isEq, err := a.IsEquivalent(b)
		if err != nil {
			t.Fatal(err)
		}
		if !isEq {
			t.Fatal("stage.IsEquivalent() returned false")
		}
	})

	t.Run("differing WorkingDir not equivalent", func(t *testing.T) {
		a, b := newStage(), newStage()
		b.WorkingDir = "different/dir"

		isEq, err := a.IsEquivalent(b)
		if err != nil {
			t.Fatal(err)
		}
		if isEq {
			t.Fatal("stage.IsEquivalent() returned true")
		}
	})

	t.Run("differing Command not equivalent", func(t *testing.T) {
		a, b := newStage(), newStage()
		b.Command = "echo 'foo'"

		isEq, err := a.IsEquivalent(b)
		if err != nil {
			t.Fatal(err)
		}
		if isEq {
			t.Fatal("stage.IsEquivalent() returned true")
		}
	})

	t.Run("differing paths in Outputs not equivalent", func(t *testing.T) {
		a, b := newStage(), newStage()
		b.Outputs[0].Path = "fizz.buzz"

		isEq, err := a.IsEquivalent(b)
		if err != nil {
			t.Fatal(err)
		}
		if isEq {
			t.Fatal("stage.IsEquivalent() returned true")
		}
	})

	t.Run("differing checksums in Outputs are equivalent", func(t *testing.T) {
		a, b := newStage(), newStage()
		b.Outputs[0].Checksum = "doesn't matter"

		isEq, err := a.IsEquivalent(b)
		if err != nil {
			t.Fatal(err)
		}
		if !isEq {
			t.Fatal("stage.IsEquivalent() returned false")
		}
	})

	t.Run("differing IsDir in Dependencies not equivalent", func(t *testing.T) {
		a, b := newStage(), newStage()
		b.Dependencies[0].IsDir = false

		isEq, err := a.IsEquivalent(b)
		if err != nil {
			t.Fatal(err)
		}
		if isEq {
			t.Fatal("stage.IsEquivalent() returned true")
		}
	})

	t.Run("differing checksums in Dependencies are equivalent", func(t *testing.T) {
		a, b := newStage(), newStage()
		b.Dependencies[0].Checksum = "doesn't matter"

		isEq, err := a.IsEquivalent(b)
		if err != nil {
			t.Fatal(err)
		}
		if !isEq {
			t.Fatal("stage.IsEquivalent() returned false")
		}
	})

	// TODO: test both Stages are not modified by call to IsEquivalent
}

func TestFromFile(t *testing.T) {

	fromYamlFileOrig := fromYamlFile
	fromYamlFile = func(path string, v interface{}) error {
		return errors.New("Mock not implemented")
	}
	var resetFromYamlFileMock = func() { fromYamlFile = fromYamlFileOrig }

	t.Run("loads stage file if no lock file found", func(t *testing.T) {
		defer resetFromYamlFileMock()
		expectedStage := Stage{WorkingDir: "foo", Command: "echo 'bar'"}
		fromYamlFile = func(path string, v interface{}) error {
			output := v.(*Stage)
			if path == "stage.yaml" {
				*output = expectedStage
				return nil
			}
			return os.ErrNotExist
		}

		outputStage, isLock, err := FromFile("stage.yaml")
		if err != nil {
			t.Fatal(err)
		}
		if isLock {
			t.Fatal("FromFile returned isLock = true")
		}
		if diff := cmp.Diff(expectedStage, outputStage); diff != "" {
			t.Fatalf("Stage -want +got:\n%s", diff)
		}
	})

	t.Run("loads stage file if lock file found but not equivalent", func(t *testing.T) {
		defer resetFromYamlFileMock()
		expectedStage := Stage{
			WorkingDir: "foo",
			Command:    "echo 'bar'",
			Outputs: []artifact.Artifact{
				{Path: "bar.txt"},
			},
		}
		lockedStage := Stage{
			WorkingDir: "foo",
			Command:    "echo 'bar'",
			Outputs: []artifact.Artifact{
				{
					Checksum: "abcdef",
					Path:     "bar.txt",
				},
				{
					Checksum: "ghijkl",
					Path:     "inequivalent.txt",
				},
			},
		}
		fromYamlFile = func(path string, v interface{}) error {
			output := v.(*Stage)
			if path == "stage.yaml" {
				*output = expectedStage
			} else {
				*output = lockedStage
			}
			return nil
		}

		outputStage, isLock, err := FromFile("stage.yaml")
		if err != nil {
			t.Fatal(err)
		}
		if isLock {
			t.Fatal("FromFile returned isLock = true")
		}
		if diff := cmp.Diff(expectedStage, outputStage); diff != "" {
			t.Fatalf("Stage -want +got:\n%s", diff)
		}
	})

	t.Run("loads lock file if lock file found and equivalent", func(t *testing.T) {
		defer resetFromYamlFileMock()
		stg := Stage{
			WorkingDir: "foo",
			Command:    "echo 'bar'",
			Outputs: []artifact.Artifact{
				{Path: "bar.txt"},
			},
		}
		lockedStage := Stage{
			WorkingDir: "foo",
			Command:    "echo 'bar'",
			Outputs: []artifact.Artifact{
				{
					Checksum: "abcdef",
					Path:     "bar.txt",
				},
			},
		}
		fromYamlFile = func(path string, v interface{}) error {
			output := v.(*Stage)
			if path == "stage.yaml" {
				*output = stg
			} else {
				*output = lockedStage
			}
			return nil
		}

		outputStage, isLock, err := FromFile("stage.yaml")
		if err != nil {
			t.Fatal(err)
		}
		if !isLock {
			t.Fatal("FromFile returned isLock = false")
		}
		if diff := cmp.Diff(lockedStage, outputStage); diff != "" {
			t.Fatalf("Stage -want +got:\n%s", diff)
		}
	})
}

func TestCommit(t *testing.T) {
	t.Run("Copy", func(t *testing.T) { testCommit(strategy.CopyStrategy, t) })
	t.Run("Link", func(t *testing.T) { testCommit(strategy.LinkStrategy, t) })
}

func testCommit(strat strategy.CheckoutStrategy, t *testing.T) {
	stg := Stage{
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
	// TODO: test artifact checksums set
}

func TestCheckout(t *testing.T) {
	t.Run("Copy", func(t *testing.T) { testCheckout(strategy.CopyStrategy, t) })
	t.Run("Link", func(t *testing.T) { testCheckout(strategy.LinkStrategy, t) })
}

func testCheckout(strat strategy.CheckoutStrategy, t *testing.T) {
	stg := Stage{
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

func TestFilePathForLock(t *testing.T) {
	input := "foo/bar.yaml"
	want := "foo/bar.yaml.lock"
	got := FilePathForLock(input)
	if got != want {
		t.Fatalf("FilePathForLock(%#v) = %#v, want %#v", input, got, want)
	}
}
