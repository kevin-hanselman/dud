package stage

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kevin-hanselman/duc/src/artifact"
)

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

	t.Run("falls back to stage when differences in artifacts", func(t *testing.T) {
		defer resetFromYamlFileMock()
		stg := Stage{
			WorkingDir: "foo/bar",
			Command:    "echo 'bar'",
			Outputs: map[string]*artifact.Artifact{
				"bish.txt": {Path: "bish.txt"},
				"bash.txt": {Path: "bash.txt"},
				"bosh":     {Path: "bosh", IsDir: true},
			},
		}
		lockedStage := Stage{
			WorkingDir: "foo",
			Command:    "echo 'bar'",
			Outputs: map[string]*artifact.Artifact{
				"bish.txt": {
					Checksum: "abcdef",
					Path:     "bish.txt",
				},
				"bosh": {
					Checksum:    "ghijkl",
					Path:        "bosh",
					IsDir:       true,
					IsRecursive: true,
				},
				"deleted_from_stage.txt": {
					Checksum: "ghijkl",
					Path:     "deleted_from_stage.txt",
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
		if isLock {
			t.Fatal("FromFile returned isLock = true")
		}

		expectedStage := Stage{
			WorkingDir: "foo/bar",
			Command:    "echo 'bar'",
			Outputs: map[string]*artifact.Artifact{
				"bish.txt": {
					Path:     "bish.txt",
					Checksum: "abcdef",
				},
				"bash.txt": {Path: "bash.txt"},
				"bosh":     {Path: "bosh", IsDir: true},
			},
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

	t.Run("same/isLocked is false when only Stage metadata changes", func(t *testing.T) {
		defer resetFromYamlFileMock()
		stg := Stage{
			WorkingDir: "foo",
			Command:    "echo 'new command'",
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
		expectedStage := Stage{
			WorkingDir: "foo",
			Command:    "echo 'new command'",
			Outputs: map[string]*artifact.Artifact{
				"bar.txt": {
					Checksum: "abcdef",
					Path:     "bar.txt",
				},
			},
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

	t.Run("skipCache is always true for dependencies", func(t *testing.T) {
		defer resetFromYamlFileMock()
		stg := Stage{
			WorkingDir: "foo",
			Command:    "echo 'new command'",
			Dependencies: map[string]*artifact.Artifact{
				"foo.txt":  {Path: "foo.txt"},
				"bish.txt": {Path: "bish.txt"},
			},
			Outputs: map[string]*artifact.Artifact{
				"bar.txt": {Path: "bar.txt", SkipCache: true},
			},
		}
		lockedStage := Stage{
			WorkingDir: "foo",
			Command:    "echo 'bar'",
			Dependencies: map[string]*artifact.Artifact{
				"foo.txt": {
					Path:      "foo.txt",
					SkipCache: true,
					Checksum:  "foo_checksum",
				},
			},
			Outputs: map[string]*artifact.Artifact{
				"bar.txt": {
					Path:     "bar.txt",
					Checksum: "bar_checksum",
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
		expectedStage := Stage{
			WorkingDir: "foo",
			Command:    "echo 'new command'",
			Dependencies: map[string]*artifact.Artifact{
				"foo.txt": {
					Path:      "foo.txt",
					SkipCache: true,
					Checksum:  "foo_checksum",
				},
				"bish.txt": {Path: "bish.txt", SkipCache: true},
			},
			Outputs: map[string]*artifact.Artifact{
				"bar.txt": {
					Path:      "bar.txt",
					SkipCache: true,
				},
			},
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
}

func TestFilePathForLock(t *testing.T) {
	input := "foo/bar.yaml"
	want := "foo/bar.yaml.lock"
	got := FilePathForLock(input)
	if got != want {
		t.Fatalf("FilePathForLock(%#v) = %#v, want %#v", input, got, want)
	}
}
