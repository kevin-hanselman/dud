package index

import (
	"os"
	"testing"

	"github.com/kevin-hanselman/dud/src/artifact"
	"github.com/kevin-hanselman/dud/src/stage"
)

func TestAdd(t *testing.T) {

	stageFromFileOrig := stage.FromFile
	stage.FromFile = func(path string) (stage.Stage, bool, error) {
		return stage.Stage{}, false, nil
	}
	defer func() { stage.FromFile = stageFromFileOrig }()

	t.Run("add new stage", func(t *testing.T) {
		idx := make(Index)
		path := "foo/bar.dud"

		if err := idx.AddStageFromPath(path); err != nil {
			t.Fatal(err)
		}

		_, added := idx[path]
		if !added {
			t.Fatal("path wasn't added to the index")
		}
	})

	t.Run("error if already tracked", func(t *testing.T) {
		idx := make(Index)
		path := "foo/bar.dud"

		var stg stage.Stage
		idx[path] = &entry{Stage: stg}

		if err := idx.AddStageFromPath(path); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("error if invalid stage", func(t *testing.T) {
		idx := make(Index)
		path := "foo/bar.dud"

		stage.FromFile = func(path string) (stage.Stage, bool, error) {
			return stage.Stage{}, false, os.ErrNotExist
		}

		err := idx.AddStageFromPath(path)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("error if new stage declares already owned outputs", func(t *testing.T) {
		stage.FromFile = func(path string) (stage.Stage, bool, error) {
			stg := stage.Stage{
				Outputs: map[string]*artifact.Artifact{
					"subDir/foo.bin": {Path: "subDir/foo.bin"},
				},
			}
			return stg, false, nil
		}
		idx := make(Index)
		idx["foo.yaml"] = &entry{
			Stage: stage.Stage{
				WorkingDir: "subDir",
				Outputs: map[string]*artifact.Artifact{
					"subDir/foo.bin": {Path: "subDir/foo.bin"},
				},
			},
		}
		err := idx.AddStageFromPath("bar.yaml")
		if err == nil {
			t.Fatal("expected error")
		}
		expectedError := "bar.yaml: artifact subDir/foo.bin already owned by foo.yaml"
		if err.Error() != expectedError {
			t.Fatalf("\nerror want: %s\nerror got: %s", expectedError, err.Error())
		}
	})

	t.Run("working dir should have no effect on artifact paths", func(t *testing.T) {
		stage.FromFile = func(path string) (stage.Stage, bool, error) {
			stg := stage.Stage{
				Outputs: map[string]*artifact.Artifact{
					"subDir/foo.bin": {Path: "subDir/foo.bin"},
				},
			}
			return stg, false, nil
		}
		idx := make(Index)
		idx["foo.yaml"] = &entry{
			Stage: stage.Stage{
				WorkingDir: "subDir",
				Outputs: map[string]*artifact.Artifact{
					"foo.bin": {Path: "foo.bin"},
				},
			},
		}
		err := idx.AddStageFromPath("bar.yaml")
		if err != nil {
			t.Fatal(err)
		}
	})
}
