package stage

import (
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kevin-hanselman/dud/src/artifact"
	"github.com/pkg/errors"
)

func TestFromFile(t *testing.T) {
	fromYamlFileOrig := fromYamlFile
	fromYamlFile = func(path string, output *Stage) error {
		panic("Mock not implemented")
	}
	resetFromYamlFileMock := func() { fromYamlFile = fromYamlFileOrig }

	fromFileErr := func(path string) error {
		_, err := FromFile(path)
		return errors.Cause(err)
	}

	t.Run("skipCache is always true for inputs", func(t *testing.T) {
		defer resetFromYamlFileMock()
		stageFile := Stage{
			WorkingDir: "foo",
			Command:    "echo 'new command'",
			Inputs: map[string]*artifact.Artifact{
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
			Inputs: map[string]*artifact.Artifact{
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

	t.Run("fail if artifact in both inputs and outputs", func(t *testing.T) {
		defer resetFromYamlFileMock()
		stageFile := Stage{
			Inputs: map[string]*artifact.Artifact{
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

		err := fromFileErr("stage.yaml")
		if err == nil {
			t.Fatal("expected FromFile to return error")
		}
	})

	t.Run("fail if output dir artifact would contain a input", func(t *testing.T) {
		defer resetFromYamlFileMock()
		stageFile := Stage{
			Inputs: map[string]*artifact.Artifact{
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

		err := fromFileErr("stage.yaml")
		if err == nil {
			t.Fatal("expected FromFile to return error")
		}
	})

	t.Run("fail if input dir artifact would contain a dir output", func(t *testing.T) {
		defer resetFromYamlFileMock()
		stageFile := Stage{
			Inputs: map[string]*artifact.Artifact{
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

		err := fromFileErr("stage.yaml")
		if err == nil {
			t.Fatal("expected FromFile to return error")
		}
	})

	t.Run("working dir should have no effect on artifact paths", func(t *testing.T) {
		defer resetFromYamlFileMock()
		stageFile := Stage{
			WorkingDir: "workDir",
			Inputs: map[string]*artifact.Artifact{
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

		err := fromFileErr("stage.yaml")
		if err == nil {
			t.Fatal("expected FromFile to return error")
		}
	})

	t.Run("no inputs and no outputs causes error", func(t *testing.T) {
		defer resetFromYamlFileMock()
		stageFile := Stage{
			WorkingDir: "workDir",
			Inputs:     map[string]*artifact.Artifact{},
			Outputs:    map[string]*artifact.Artifact{},
		}
		fromYamlFile = func(path string, output *Stage) error {
			if path == "stage.yaml" {
				*output = stageFile
				return nil
			}
			return os.ErrNotExist
		}

		err := fromFileErr("stage.yaml")
		if err == nil {
			t.Fatal("expected FromFile to return error")
		}

		expectedError := "declared no inputs and no outputs"
		if diff := cmp.Diff(expectedError, err.Error()); diff != "" {
			t.Fatalf("error -want +got:\n%s", diff)
		}
	})

	t.Run("cleans paths", func(t *testing.T) {
		defer resetFromYamlFileMock()
		stageFile := Stage{
			WorkingDir: "foo/",
			Inputs: map[string]*artifact.Artifact{
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
			Inputs: map[string]*artifact.Artifact{
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

	t.Run("disallow dirs outside project root", func(t *testing.T) {
		defer resetFromYamlFileMock()

		assert := func(t *testing.T, stageFile Stage, want string) {
			fromYamlFile = func(path string, output *Stage) error {
				if path == "stage.yaml" {
					*output = stageFile
					return nil
				}
				return os.ErrNotExist
			}
			err := fromFileErr("stage.yaml")
			if err == nil {
				t.Fatal("expected FromFile to return error")
			}

			got := err.Error()
			if !strings.Contains(got, want) {
				t.Fatalf("error want: %v got: %v", want, got)
			}
		}

		t.Run("double dot", func(t *testing.T) {
			badPath := "foo/../../bar"
			want := "outside of the project root"

			t.Run("working dir", func(t *testing.T) {
				stageFile := Stage{
					WorkingDir: badPath,
					Outputs: map[string]*artifact.Artifact{
						"out": {},
					},
				}
				assert(t, stageFile, want)
			})

			t.Run("artifact", func(t *testing.T) {
				stageFile := Stage{
					Outputs: map[string]*artifact.Artifact{
						badPath: {},
					},
				}
				assert(t, stageFile, want)
			})
		})

		t.Run("abs path", func(t *testing.T) {
			badPath := "/foo/bar"
			want := "absolute path"

			t.Run("working dir", func(t *testing.T) {
				stageFile := Stage{
					WorkingDir: badPath,
					Outputs: map[string]*artifact.Artifact{
						"out": {},
					},
				}
				assert(t, stageFile, want)
			})

			t.Run("artifact", func(t *testing.T) {
				stageFile := Stage{
					Outputs: map[string]*artifact.Artifact{
						badPath: {},
					},
				}
				assert(t, stageFile, want)
			})
		})
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

		err := fromFileErr("stage.yaml")
		if err == nil {
			t.Fatal("expected FromFile to return error")
		}

		expectedError := "artifact .. is outside of the project root"
		if diff := cmp.Diff(expectedError, err.Error()); diff != "" {
			t.Fatalf("error -want +got:\n%s", diff)
		}
	})

	t.Run("disallow no outputs and no command", func(t *testing.T) {
		defer resetFromYamlFileMock()
		stageFile := Stage{
			Inputs: map[string]*artifact.Artifact{
				"foo": {},
			},
		}

		fromYamlFile = func(path string, output *Stage) error {
			if path == "stage.yaml" {
				*output = stageFile
				return nil
			}
			return os.ErrNotExist
		}

		err := fromFileErr("stage.yaml")
		if err == nil {
			t.Fatal("expected FromFile to return error")
		}

		expectedError := "declared no outputs and no command"
		if diff := cmp.Diff(expectedError, err.Error()); diff != "" {
			t.Fatalf("error -want +got:\n%s", diff)
		}

		// Assert that stage.Command has spaces trimmed.
		stageFile = Stage{
			Command: "  ",
			Inputs: map[string]*artifact.Artifact{
				"foo": {},
			},
		}

		err = fromFileErr("stage.yaml")
		if err == nil {
			t.Fatal("expected FromFile to return error")
		}

		if diff := cmp.Diff(expectedError, err.Error()); diff != "" {
			t.Fatalf("error -want +got:\n%s", diff)
		}
	})

	t.Run("stage files cannot reference themselves", func(t *testing.T) {
		defer resetFromYamlFileMock()
		stageFile := Stage{
			Outputs: map[string]*artifact.Artifact{
				"stage.yaml": {},
			},
		}

		fromYamlFile = func(path string, output *Stage) error {
			if path == "stage.yaml" {
				*output = stageFile
				return nil
			}
			return os.ErrNotExist
		}

		expectedError := "stage references itself in outputs"

		err := fromFileErr("stage.yaml")
		if err == nil {
			t.Fatal("expected FromFile to return error")
		}

		if diff := cmp.Diff(expectedError, err.Error()); diff != "" {
			t.Fatalf("error -want +got:\n%s", diff)
		}

		stageFile = Stage{
			Command: "echo hello world",
			Inputs: map[string]*artifact.Artifact{
				"stage.yaml": {},
			},
		}

		expectedError = "stage references itself in inputs"

		err = fromFileErr("stage.yaml")
		if err == nil {
			t.Fatal("expected FromFile to return error")
		}

		if diff := cmp.Diff(expectedError, err.Error()); diff != "" {
			t.Fatalf("error -want +got:\n%s", diff)
		}
	})
}
