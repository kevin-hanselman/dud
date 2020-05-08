package track

import (
	"github.com/google/go-cmp/cmp"
	"github.com/kevlar1818/duc/artifact"
	"github.com/kevlar1818/duc/stage"
	"io/ioutil"
	"os"
	"testing"
)

func TestTrackOnePath(t *testing.T) {
	fileStatusFromPathOrig := fileStatusFromPath
	fileStatusFromPath = func(path string) (artifact.FileStatus, error) {
		return artifact.RegularFile, nil
	}
	defer func() { fileStatusFromPath = fileStatusFromPathOrig }()
	path := "foobar.txt"
	expectedStage := stage.Stage{
		Outputs: []artifact.Artifact{
			{
				Checksum: "",
				Path:     path,
			},
		},
	}

	actualStage, err := Track(path)

	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(expectedStage, actualStage); diff != "" {
		t.Fatalf("Track() -want +got:\n%s", diff)
	}
}

func TestTrackMultiplePaths(t *testing.T) {
	fileStatusFromPathOrig := fileStatusFromPath
	fileStatusFromPath = func(path string) (artifact.FileStatus, error) {
		return artifact.RegularFile, nil
	}
	defer func() { fileStatusFromPath = fileStatusFromPathOrig }()
	paths := []string{"foo.txt", "bar.bin"}
	expectedStage := stage.Stage{
		Outputs: []artifact.Artifact{
			{
				Checksum: "",
				Path:     "foo.txt",
			},
			{
				Checksum: "",
				Path:     "bar.bin",
			},
		},
	}

	actualStage, err := Track(paths...)

	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(expectedStage, actualStage); diff != "" {
		t.Fatalf("Track() -want +got:\n%s", diff)
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

	expectedStage := stage.Stage{
		Outputs: []artifact.Artifact{
			{
				Checksum: "",
				Path:     paths[0],
			},
			{
				Checksum: "",
				Path:     paths[1],
			},
		},
	}

	actualStage, err := Track(paths...)

	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(expectedStage, actualStage); diff != "" {
		t.Fatalf("Track() -want +got:\n%s", diff)
	}
}
