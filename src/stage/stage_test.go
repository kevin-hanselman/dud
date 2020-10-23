package stage

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kevin-hanselman/dud/src/artifact"
)

func TestFromFile(t *testing.T) {

	fromYamlFileOrig := fromYamlFile
	fromYamlFile = func(path string, output *stageFileFormat) error {
		panic("Mock not implemented")
	}
	var resetFromYamlFileMock = func() { fromYamlFile = fromYamlFileOrig }

	t.Run("loads stage file if no lock file found", func(t *testing.T) {
		defer resetFromYamlFileMock()
		expectedStage := Stage{WorkingDir: "foo", Command: "echo 'bar'"}
		fromYamlFile = func(path string, output *stageFileFormat) error {
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
		fromYamlFile = func(path string, output *stageFileFormat) error {
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
		fromYamlFile = func(path string, output *stageFileFormat) error {
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
		fromYamlFile = func(path string, output *stageFileFormat) error {
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
		fromYamlFile = func(path string, output *stageFileFormat) error {
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

	t.Run("fail if artifact in both deps and outputs", func(t *testing.T) {
		defer resetFromYamlFileMock()
		stgFile := stageFileFormat{
			Dependencies: []*artifact.Artifact{
				{Path: "foo.txt"},
			},
			Outputs: []*artifact.Artifact{
				{Path: "foo.txt"},
			},
		}
		fromYamlFile = func(path string, output *stageFileFormat) error {
			if path == "stage.yaml" {
				*output = stgFile
				return nil
			}
			return os.ErrNotExist
		}

		_, _, err := FromFile("stage.yaml")
		if err == nil {
			t.Fatal("expected FromFile to return error")
		}
	})

	t.Run("fail if output dir artifact would contain a dep", func(t *testing.T) {
		defer resetFromYamlFileMock()
		stgFile := stageFileFormat{
			Dependencies: []*artifact.Artifact{
				{Path: "foo/bar.txt"},
			},
			Outputs: []*artifact.Artifact{
				{Path: "foo", IsDir: true},
			},
		}
		fromYamlFile = func(path string, output *stageFileFormat) error {
			if path == "stage.yaml" {
				*output = stgFile
				return nil
			}
			return os.ErrNotExist
		}

		_, _, err := FromFile("stage.yaml")
		if err == nil {
			t.Fatal("expected FromFile to return error")
		}
	})

	t.Run("fail if dep dir artifact would contain a dir output", func(t *testing.T) {
		defer resetFromYamlFileMock()
		stgFile := stageFileFormat{
			Dependencies: []*artifact.Artifact{
				{Path: "foo", IsDir: true},
			},
			Outputs: []*artifact.Artifact{
				{Path: "foo/bar", IsDir: true},
			},
		}
		fromYamlFile = func(path string, output *stageFileFormat) error {
			if path == "stage.yaml" {
				*output = stgFile
				return nil
			}
			return os.ErrNotExist
		}

		_, _, err := FromFile("stage.yaml")
		if err == nil {
			t.Fatal("expected FromFile to return error")
		}
	})

	t.Run("working dir should have no effect on artifact paths", func(t *testing.T) {
		defer resetFromYamlFileMock()
		stgFile := stageFileFormat{
			WorkingDir: "workDir",
			Dependencies: []*artifact.Artifact{
				{Path: "foo", IsDir: true},
			},
			Outputs: []*artifact.Artifact{
				{Path: "foo/bar", IsDir: true},
			},
		}
		fromYamlFile = func(path string, output *stageFileFormat) error {
			if path == "stage.yaml" {
				*output = stgFile
				return nil
			}
			return os.ErrNotExist
		}

		_, _, err := FromFile("stage.yaml")
		if err == nil {
			t.Fatal("expected FromFile to return error")
		}
	})

	t.Run("cleans paths", func(t *testing.T) {
		defer resetFromYamlFileMock()
		stgFile := stageFileFormat{
			WorkingDir: "foo/",
			Dependencies: []*artifact.Artifact{
				{Path: "foo/../bish.txt"},
			},
			Outputs: []*artifact.Artifact{
				{Path: "./bar/", IsDir: true},
			},
		}

		fromYamlFile = func(path string, output *stageFileFormat) error {
			if path == "stage.yaml" {
				*output = stgFile
				return nil
			}
			return os.ErrNotExist
		}

		expectedStage := stageFileFormat{
			WorkingDir: "foo",
			Dependencies: []*artifact.Artifact{
				{Path: "bish.txt", SkipCache: true},
			},
			Outputs: []*artifact.Artifact{
				{Path: "bar", IsDir: true},
			},
		}.toStage()

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

	t.Run("disallow working dirs outside project root", func(t *testing.T) {
		defer resetFromYamlFileMock()
		stgFile := stageFileFormat{
			WorkingDir: "foo/../../bar",
			Outputs: []*artifact.Artifact{
				{Path: "bar", IsDir: true},
			},
		}

		fromYamlFile = func(path string, output *stageFileFormat) error {
			if path == "stage.yaml" {
				*output = stgFile
				return nil
			}
			return os.ErrNotExist
		}

		_, _, err := FromFile("stage.yaml")
		if err == nil {
			t.Fatal("expcted FromFile to return error")
		}

		expectedError := "working directory ../bar is outside of the project root"
		if diff := cmp.Diff(expectedError, err.Error()); diff != "" {
			t.Fatalf("error -want +got:\n%s", diff)
		}
	})

	t.Run("disallow artifact paths outside project root", func(t *testing.T) {
		defer resetFromYamlFileMock()
		stgFile := stageFileFormat{
			Outputs: []*artifact.Artifact{
				{Path: "..", IsDir: true},
			},
		}

		fromYamlFile = func(path string, output *stageFileFormat) error {
			if path == "stage.yaml" {
				*output = stgFile
				return nil
			}
			return os.ErrNotExist
		}

		_, _, err := FromFile("stage.yaml")
		if err == nil {
			t.Fatal("expcted FromFile to return error")
		}

		expectedError := "artifact .. is outside of the project root"
		if diff := cmp.Diff(expectedError, err.Error()); diff != "" {
			t.Fatalf("error -want +got:\n%s", diff)
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
