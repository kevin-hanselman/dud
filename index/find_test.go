package index

import (
	"github.com/kevin-hanselman/duc/artifact"
	"github.com/kevin-hanselman/duc/stage"

	"testing"
)

func TestFindOwner(t *testing.T) {

	t.Run("single stage", func(t *testing.T) {
		idx := make(Index)
		idx["foo.yaml"] = &entry{
			Stage: stage.Stage{
				Outputs: map[string]*artifact.Artifact{
					"foo.bin": {Path: "foo.bin"},
					"bar.bin": {Path: "bar.bin"},
				},
			},
		}

		owner, err := idx.findOwner("bar.bin")

		if err != nil {
			t.Fatal(err)
		}

		if owner != "foo.yaml" {
			t.Fatalf("got owner = %v, want foo.yaml", owner)
		}
	})

	t.Run("no owner", func(t *testing.T) {
		idx := make(Index)
		idx["foo.yaml"] = &entry{
			Stage: stage.Stage{
				Outputs: map[string]*artifact.Artifact{
					"foo.bin": {Path: "foo.bin"},
					"bar.bin": {Path: "bar.bin"},
				},
			},
		}

		owner, err := idx.findOwner("other.bin")

		if err != nil {
			t.Fatal(err)
		}

		if owner != "" {
			t.Fatalf("got owner = %v, want empty string", owner)
		}
	})

	t.Run("file in dir artifact", func(t *testing.T) {
		idx := make(Index)
		idx["foo.yaml"] = &entry{
			Stage: stage.Stage{
				Outputs: map[string]*artifact.Artifact{
					"foo": {Path: "foo", IsDir: true, IsRecursive: false},
				},
			},
		}

		owner, err := idx.findOwner("foo/bar.bin")

		if err != nil {
			t.Fatal(err)
		}

		if owner != "foo.yaml" {
			t.Fatalf("got owner = %v, want foo.yaml", owner)
		}
	})

	t.Run("non-zero WorkingDir", func(t *testing.T) {
		idx := make(Index)
		idx["foo.yaml"] = &entry{
			Stage: stage.Stage{
				WorkingDir: "foo",
				Outputs: map[string]*artifact.Artifact{
					"bar.bin": {Path: "bar.bin"},
				},
			},
		}

		owner, err := idx.findOwner("foo/bar.bin")

		if err != nil {
			t.Fatal(err)
		}

		if owner != "foo.yaml" {
			t.Fatalf("got owner = %v, want foo.yaml", owner)
		}
	})

	t.Run("file in sub-dir of non-recursive dir artifact", func(t *testing.T) {
		idx := make(Index)
		idx["foo.yaml"] = &entry{
			Stage: stage.Stage{
				Outputs: map[string]*artifact.Artifact{
					"foo": {Path: "foo", IsDir: true, IsRecursive: false},
				},
			},
		}

		owner, err := idx.findOwner("foo/bar/test.bin")

		if err != nil {
			t.Fatal(err)
		}

		if owner != "" {
			t.Fatalf("got owner = %v, want empty string", owner)
		}
	})

	t.Run("file in sub-dir of recursive dir artifact", func(t *testing.T) {
		idx := make(Index)
		idx["foo.yaml"] = &entry{
			Stage: stage.Stage{
				Outputs: map[string]*artifact.Artifact{
					"foo": {Path: "foo", IsDir: true, IsRecursive: true},
				},
			},
		}

		owner, err := idx.findOwner("foo/bar/test.bin")

		if err != nil {
			t.Fatal(err)
		}

		if owner != "foo.yaml" {
			t.Fatalf("got owner = %v, want foo.yaml", owner)
		}
	})
}
