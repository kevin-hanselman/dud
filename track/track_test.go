package track

import (
	"github.com/kevlar1818/duc/stage"
	"reflect"
	"testing"
)

func TestTrackOnePath(t *testing.T) {
	fileExists = func(path string) error {
		return nil
	}
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
	fileExists = func(path string) error {
		return nil
	}
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
