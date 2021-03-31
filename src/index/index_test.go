package index

import (
	"testing"

	"github.com/kevin-hanselman/dud/src/artifact"
	"github.com/kevin-hanselman/dud/src/stage"
)

func TestAdd(t *testing.T) {
	t.Run("add new stage", func(t *testing.T) {
		idx := make(Index)
		path := "foo/bar.dud"

		if err := idx.AddStage(stage.Stage{}, path); err != nil {
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
		idx[path] = &stg

		if err := idx.AddStage(stg, path); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("error if new stage declares already owned outputs", func(t *testing.T) {
		stg := stage.Stage{
			Outputs: map[string]*artifact.Artifact{
				"subDir/foo.bin": {Path: "subDir/foo.bin"},
			},
		}
		idx := Index{
			"foo.yaml": &stage.Stage{
				WorkingDir: "subDir",
				Outputs: map[string]*artifact.Artifact{
					"subDir/foo.bin": {Path: "subDir/foo.bin"},
				},
			},
		}
		err := idx.AddStage(stg, "bar.yaml")
		if err == nil {
			t.Fatal("expected error")
		}
		expectedError := "bar.yaml: artifact subDir/foo.bin already owned by foo.yaml"
		if err.Error() != expectedError {
			t.Fatalf("\nerror want: %s\nerror got: %s", expectedError, err.Error())
		}
	})

	t.Run("working dir should have no effect on artifact paths", func(t *testing.T) {
		stg := stage.Stage{
			Outputs: map[string]*artifact.Artifact{
				"subDir/foo.bin": {Path: "subDir/foo.bin"},
			},
		}
		idx := Index{
			"foo.yaml": &stage.Stage{
				WorkingDir: "subDir",
				Outputs: map[string]*artifact.Artifact{
					"foo.bin": {Path: "foo.bin"},
				},
			},
		}
		err := idx.AddStage(stg, "bar.yaml")
		if err != nil {
			t.Fatal(err)
		}
	})
}
