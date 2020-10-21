package index

import (
	"github.com/google/go-cmp/cmp"
	"github.com/kevin-hanselman/dud/src/artifact"
	"github.com/kevin-hanselman/dud/src/stage"

	"testing"
)

func TestFindOwner(t *testing.T) {

	t.Run("single stage", func(t *testing.T) {
		targetArt := artifact.Artifact{Path: "bar.bin"}
		idx := make(Index)
		idx["foo.yaml"] = &entry{
			Stage: stage.Stage{
				Outputs: map[string]*artifact.Artifact{
					"foo.bin": {Path: "foo.bin"},
					"bar.bin": &targetArt,
				},
			},
		}

		owner, foundArt, err := idx.findOwner("bar.bin")
		if err != nil {
			t.Fatal(err)
		}

		if owner != "foo.yaml" {
			t.Fatalf("got owner = %#v, want foo.yaml", owner)
		}

		if diff := cmp.Diff(&targetArt, foundArt); diff != "" {
			t.Fatalf("artifact -want +got:\n%s", diff)
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

		owner, _, err := idx.findOwner("other.bin")

		if err != nil {
			t.Fatal(err)
		}

		if owner != "" {
			t.Fatalf("got owner = %#v, want empty string", owner)
		}
	})

	t.Run("file in dir artifact", func(t *testing.T) {
		targetArt := artifact.Artifact{Path: "foo", IsDir: true, IsRecursive: false}
		idx := make(Index)
		idx["foo.yaml"] = &entry{
			Stage: stage.Stage{
				Outputs: map[string]*artifact.Artifact{
					"foo": &targetArt,
				},
			},
		}

		owner, foundArt, err := idx.findOwner("foo/bar.bin")
		if err != nil {
			t.Fatal(err)
		}

		if owner != "foo.yaml" {
			t.Fatalf("got owner = %#v, want foo.yaml", owner)
		}

		if diff := cmp.Diff(&targetArt, foundArt); diff != "" {
			t.Fatalf("artifact -want +got:\n%s", diff)
		}
	})

	t.Run("working dir doesn't affect artifact paths", func(t *testing.T) {
		targetArt := artifact.Artifact{Path: "bar.bin"}
		idx := make(Index)
		idx["foo.yaml"] = &entry{
			Stage: stage.Stage{
				WorkingDir: "foo",
				Outputs: map[string]*artifact.Artifact{
					"foo/bar.bin": &targetArt,
				},
			},
		}

		owner, foundArt, err := idx.findOwner("foo/bar.bin")
		if err != nil {
			t.Fatal(err)
		}

		if owner != "foo.yaml" {
			t.Fatalf("got owner = %#v, want foo.yaml", owner)
		}

		if diff := cmp.Diff(&targetArt, foundArt); diff != "" {
			t.Fatalf("artifact -want +got:\n%s", diff)
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

		owner, _, err := idx.findOwner("foo/bar/test.bin")

		if err != nil {
			t.Fatal(err)
		}

		if owner != "" {
			t.Fatalf("got owner = %#v, want empty string", owner)
		}
	})

	t.Run("file in sub-dir of recursive dir artifact", func(t *testing.T) {
		targetArt := artifact.Artifact{Path: "foo", IsDir: true, IsRecursive: true}
		idx := make(Index)
		idx["foo.yaml"] = &entry{
			Stage: stage.Stage{
				Outputs: map[string]*artifact.Artifact{
					"foo": &targetArt,
				},
			},
		}

		owner, foundArt, err := idx.findOwner("foo/bar/test.bin")
		if err != nil {
			t.Fatal(err)
		}

		if owner != "foo.yaml" {
			t.Fatalf("got owner = %#v, want foo.yaml", owner)
		}

		if diff := cmp.Diff(&targetArt, foundArt); diff != "" {
			t.Fatalf("artifact -want +got:\n%s", diff)
		}
	})
}
