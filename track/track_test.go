package track

import (
	"github.com/kevlar1818/duc/stage"
	"io/ioutil"
	"reflect"
	"testing"
)

func TestTrackOnePath(t *testing.T) {
	fileExistsOrig := fileExists
	fileExists = func(path string) error {
		return nil
	}
	defer func() { fileExists = fileExistsOrig }()
	path := "foobar.txt"
	expected := stage.Stage{
		Checksum: nil,
		Outputs: []stage.Artifact{
			stage.Artifact{
				Checksum: nil,
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
	fileExists = func(path string) error {
		return nil
	}
	defer func() { fileExists = fileExistsOrig }()
	paths := []string{"foo.txt", "bar.bin"}
	expected := stage.Stage{
		Checksum: nil,
		Outputs: []stage.Artifact{
			stage.Artifact{
				Checksum: nil,
				Path:     "foo.txt",
			},
			stage.Artifact{
				Checksum: nil,
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
	for _, path := range paths {
		ioutil.TempFile(dir, path)
	}

	expected := stage.Stage{
		Checksum: nil,
		Outputs: []stage.Artifact{
			stage.Artifact{
				Checksum: nil,
				Path:     "duc/foo.txt",
			},
			stage.Artifact{
				Checksum: nil,
				Path:     "duc/bar.bin",
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
