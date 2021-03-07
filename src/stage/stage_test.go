package stage

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kevin-hanselman/dud/src/artifact"
)

func TestFromFile(t *testing.T) {
	fromYamlFileOrig := fromYamlFile
	fromYamlFile = func(path string, output *Stage) error {
		panic("Mock not implemented")
	}
	resetFromYamlFileMock := func() { fromYamlFile = fromYamlFileOrig }

	t.Run("skipCache is always true for dependencies", func(t *testing.T) {
		defer resetFromYamlFileMock()
		stageFile := Stage{
			WorkingDir: "foo",
			Command:    "echo 'new command'",
			Dependencies: map[string]*artifact.Artifact{
				"foo.txt":  {},
				"bish.txt": {},
			},
			Outputs: map[string]*artifact.Artifact{
				"bar.txt": {SkipCache: true},
			},
		}
		fromYamlFile = func(path string, output *Stage) error {
			*output = stageFile
			return nil
		}
		expectedStage := Stage{
			WorkingDir: "foo",
			Command:    "echo 'new command'",
			Dependencies: map[string]*artifact.Artifact{
				"foo.txt": {
					Path:      "foo.txt",
					SkipCache: true,
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

		outputStage, err := FromFile("stage.yaml")
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(expectedStage, outputStage); diff != "" {
			t.Fatalf("Stage -want +got:\n%s", diff)
		}
	})

	t.Run("fail if artifact in both deps and outputs", func(t *testing.T) {
		defer resetFromYamlFileMock()
		stageFile := Stage{
			Dependencies: map[string]*artifact.Artifact{
				"foo.txt": {},
			},
			Outputs: map[string]*artifact.Artifact{
				"foo.txt": {},
			},
		}
		fromYamlFile = func(path string, output *Stage) error {
			if path == "stage.yaml" {
				*output = stageFile
				return nil
			}
			return os.ErrNotExist
		}

		_, err := FromFile("stage.yaml")
		if err == nil {
			t.Fatal("expected FromFile to return error")
		}
	})

	t.Run("fail if output dir artifact would contain a dep", func(t *testing.T) {
		defer resetFromYamlFileMock()
		stageFile := Stage{
			Dependencies: map[string]*artifact.Artifact{
				"foo/bar.txt": {},
			},
			Outputs: map[string]*artifact.Artifact{
				"foo": {IsDir: true},
			},
		}
		fromYamlFile = func(path string, output *Stage) error {
			if path == "stage.yaml" {
				*output = stageFile
				return nil
			}
			return os.ErrNotExist
		}

		_, err := FromFile("stage.yaml")
		if err == nil {
			t.Fatal("expected FromFile to return error")
		}
	})

	t.Run("fail if dep dir artifact would contain a dir output", func(t *testing.T) {
		defer resetFromYamlFileMock()
		stageFile := Stage{
			Dependencies: map[string]*artifact.Artifact{
				"foo": {IsDir: true},
			},
			Outputs: map[string]*artifact.Artifact{
				"foo/bar": {IsDir: true},
			},
		}
		fromYamlFile = func(path string, output *Stage) error {
			if path == "stage.yaml" {
				*output = stageFile
				return nil
			}
			return os.ErrNotExist
		}

		_, err := FromFile("stage.yaml")
		if err == nil {
			t.Fatal("expected FromFile to return error")
		}
	})

	t.Run("working dir should have no effect on artifact paths", func(t *testing.T) {
		defer resetFromYamlFileMock()
		stageFile := Stage{
			WorkingDir: "workDir",
			Dependencies: map[string]*artifact.Artifact{
				"foo": {IsDir: true},
			},
			Outputs: map[string]*artifact.Artifact{
				"foo/bar": {IsDir: true},
			},
		}
		fromYamlFile = func(path string, output *Stage) error {
			if path == "stage.yaml" {
				*output = stageFile
				return nil
			}
			return os.ErrNotExist
		}

		_, err := FromFile("stage.yaml")
		if err == nil {
			t.Fatal("expected FromFile to return error")
		}
	})

	t.Run("cleans paths", func(t *testing.T) {
		defer resetFromYamlFileMock()
		stageFile := Stage{
			WorkingDir: "foo/",
			Dependencies: map[string]*artifact.Artifact{
				"foo/../bish.txt": {},
			},
			Outputs: map[string]*artifact.Artifact{
				"./bar/": {IsDir: true},
			},
		}

		fromYamlFile = func(path string, output *Stage) error {
			if path == "stage.yaml" {
				*output = stageFile
				return nil
			}
			return os.ErrNotExist
		}

		expectedStage := Stage{
			WorkingDir: "foo",
			Dependencies: map[string]*artifact.Artifact{
				"bish.txt": {Path: "bish.txt", SkipCache: true},
			},
			Outputs: map[string]*artifact.Artifact{
				"bar": {Path: "bar", IsDir: true},
			},
		}

		outputStage, err := FromFile("stage.yaml")
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(expectedStage, outputStage); diff != "" {
			t.Fatalf("Stage -want +got:\n%s", diff)
		}
	})

	t.Run("disallow working dirs outside project root", func(t *testing.T) {
		defer resetFromYamlFileMock()
		stageFile := Stage{
			WorkingDir: "foo/../../bar",
			Outputs: map[string]*artifact.Artifact{
				"bar": {IsDir: true},
			},
		}

		fromYamlFile = func(path string, output *Stage) error {
			if path == "stage.yaml" {
				*output = stageFile
				return nil
			}
			return os.ErrNotExist
		}

		_, err := FromFile("stage.yaml")
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
		stageFile := Stage{
			Outputs: map[string]*artifact.Artifact{
				"..": {IsDir: true},
			},
		}

		fromYamlFile = func(path string, output *Stage) error {
			if path == "stage.yaml" {
				*output = stageFile
				return nil
			}
			return os.ErrNotExist
		}

		_, err := FromFile("stage.yaml")
		if err == nil {
			t.Fatal("expcted FromFile to return error")
		}

		expectedError := "artifact .. is outside of the project root"
		if diff := cmp.Diff(expectedError, err.Error()); diff != "" {
			t.Fatalf("error -want +got:\n%s", diff)
		}
	})
}
