package track

import (
	"github.com/kevlar1818/duc/stage"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func TestTrackOnePath(t *testing.T) {
	fileExistsOrig := fileExists
	fileExists = func(path string) (bool, error) {
		return true, nil
	}
	defer func() { fileExists = fileExistsOrig }()
	path := "foobar.txt"
	expected := stage.Stage{
		Outputs: []stage.Artifact{
			stage.Artifact{
				Checksum: "",
				Path:     path,
			},
		},
	}

	stage, err := Track(path)

	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(stage, expected) {
		t.Errorf("Track(%s) = %#v, want %#v", path, stage, expected)
	}
}

func TestTrackMultiplePaths(t *testing.T) {
	fileExistsOrig := fileExists
	fileExists = func(path string) (bool, error) {
		return true, nil
	}
	defer func() { fileExists = fileExistsOrig }()
	paths := []string{"foo.txt", "bar.bin"}
	expected := stage.Stage{
		Outputs: []stage.Artifact{
			stage.Artifact{
				Checksum: "",
				Path:     "foo.txt",
			},
			stage.Artifact{
				Checksum: "",
				Path:     "bar.bin",
			},
		},
	}

	stage, err := Track(paths...)

	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(stage, expected) {
		t.Errorf("Track(%s) = %#v, want %#v", paths, stage, expected)
	}
}

func TestTrackIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	paths := []string{"foo.txt", "bar.bin"}

	dir, err := ioutil.TempDir("", "duc")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(dir)
	for i, path := range paths {
		f, err := ioutil.TempFile(dir, path)
		if err != nil {
			t.Error(err)
		}
		paths[i] = f.Name()
	}

	expected := stage.Stage{
		Outputs: []stage.Artifact{
			stage.Artifact{
				Checksum: "",
				Path:     paths[0],
			},
			stage.Artifact{
				Checksum: "",
				Path:     paths[1],
			},
		},
	}

	stage, err := Track(paths...)

	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(stage, expected) {
		t.Errorf("Track(%s) = %#v, want %#v", paths, stage, expected)
	}
}
