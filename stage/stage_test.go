package stage

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kevin-hanselman/duc/artifact"
)

func TestEquivalency(t *testing.T) {

	newStage := func() Stage {
		return Stage{
			WorkingDir: "dir",
			Outputs: map[string]*artifact.Artifact{
				"foo.txt": {Path: "foo.txt"},
				"bar.txt": {Path: "bar.txt"},
			},
			Dependencies: map[string]*artifact.Artifact{
				"dep": {Path: "dep", IsDir: true},
			},
		}
	}

	t.Run("identical stages are equivalent", func(t *testing.T) {
		a, b := newStage(), newStage()
		if !a.IsEquivalent(b) {
			t.Fatal("Stages are not equivalent")
		}
	})

	t.Run("differing WorkingDir not equivalent", func(t *testing.T) {
		a, b := newStage(), newStage()
		b.WorkingDir = "different/dir"

		if a.IsEquivalent(b) {
			t.Fatal("Stages are equivalent")
		}
	})

	t.Run("differing Command not equivalent", func(t *testing.T) {
		a, b := newStage(), newStage()
		b.Command = "echo 'foo'"

		if a.IsEquivalent(b) {
			t.Fatal("Stages are equivalent")
		}
	})

	t.Run("differing paths in Outputs not equivalent", func(t *testing.T) {
		a, b := newStage(), newStage()
		b.Outputs["foo.txt"].Path = "fizz.buzz"

		if a.IsEquivalent(b) {
			t.Fatal("Stages are equivalent")
		}
	})

	t.Run("differing checksums in Outputs are equivalent", func(t *testing.T) {
		a, b := newStage(), newStage()
		b.Outputs["foo.txt"].Checksum = "doesn't matter"

		if !a.IsEquivalent(b) {
			t.Fatal("Stages are not equivalent")
		}
	})

	t.Run("differing IsDir in Dependencies not equivalent", func(t *testing.T) {
		a, b := newStage(), newStage()
		b.Dependencies["dep"].IsDir = false

		if a.IsEquivalent(b) {
			t.Fatal("Stages are equivalent")
		}
	})

	t.Run("differing checksums in Dependencies are equivalent", func(t *testing.T) {
		a, b := newStage(), newStage()
		b.Dependencies["dep"].Checksum = "doesn't matter"

		if !a.IsEquivalent(b) {
			t.Fatal("Stages are not equivalent")
		}
	})

	// TODO: test both Stages are not modified by call to IsEquivalent
}

func TestFromFile(t *testing.T) {

	fromYamlFileOrig := fromYamlFile
	fromYamlFile = func(path string, v interface{}) error {
		panic("Mock not implemented")
	}
	var resetFromYamlFileMock = func() { fromYamlFile = fromYamlFileOrig }

	t.Run("loads stage file if no lock file found", func(t *testing.T) {
		defer resetFromYamlFileMock()
		expectedStage := Stage{WorkingDir: "foo", Command: "echo 'bar'"}
		fromYamlFile = func(path string, v interface{}) error {
			output := v.(*stageFileFormat)
			if path == "stage.yaml" {
				*output = expectedStage.toFileFormat()
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
			Outputs: map[string]*artifact.Artifact{
				"bar.txt": {Path: "bar.txt"},
			},
		}
		lockedStage := Stage{
			WorkingDir: "foo",
			Command:    "echo 'bar'",
			Outputs: map[string]*artifact.Artifact{
				"bar.txt": {
					Checksum: "abcdef",
					Path:     "bar.txt",
				},
				"inequivalent.txt": {
					Checksum: "ghijkl",
					Path:     "inequivalent.txt",
				},
			},
		}
		fromYamlFile = func(path string, v interface{}) error {
			output := v.(*stageFileFormat)
			if path == "stage.yaml" {
				*output = expectedStage.toFileFormat()
			} else {
				*output = lockedStage.toFileFormat()
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
			Outputs: map[string]*artifact.Artifact{
				"bar.txt": {Path: "bar.txt"},
			},
		}
		lockedStage := Stage{
			WorkingDir: "foo",
			Command:    "echo 'bar'",
			Outputs: map[string]*artifact.Artifact{
				"bar.txt": {
					Checksum: "abcdef",
					Path:     "bar.txt",
				},
			},
		}
		fromYamlFile = func(path string, v interface{}) error {
			output := v.(*stageFileFormat)
			if path == "stage.yaml" {
				*output = stg.toFileFormat()
			} else {
				*output = lockedStage.toFileFormat()
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

func TestFilePathForLock(t *testing.T) {
	input := "foo/bar.yaml"
	want := "foo/bar.yaml.lock"
	got := FilePathForLock(input)
	if got != want {
		t.Fatalf("FilePathForLock(%#v) = %#v, want %#v", input, got, want)
	}
}
