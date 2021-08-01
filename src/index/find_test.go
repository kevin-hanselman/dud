package index

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kevin-hanselman/dud/src/artifact"
	"github.com/kevin-hanselman/dud/src/stage"
)

func TestFindOwner(t *testing.T) {
	t.Run("single stage", func(t *testing.T) {
		targetArt := artifact.Artifact{Path: "bar.bin"}
		idx := Index{
			"foo.yaml": &stage.Stage{
				Outputs: map[string]*artifact.Artifact{
					"foo.bin": {Path: "foo.bin"},
					"bar.bin": &targetArt,
				},
			},
		}

		owner, foundArt := idx.findOwner("bar.bin")

		if owner != "foo.yaml" {
			t.Fatalf("got owner = %#v, want foo.yaml", owner)
		}

		if diff := cmp.Diff(&targetArt, foundArt); diff != "" {
			t.Fatalf("artifact -want +got:\n%s", diff)
		}
	})

	t.Run("no owner", func(t *testing.T) {
		idx := Index{
			"foo.yaml": &stage.Stage{
				Outputs: map[string]*artifact.Artifact{
					"foo.bin": {Path: "foo.bin"},
					"bar.bin": {Path: "bar.bin"},
				},
			},
		}

		owner, _ := idx.findOwner("other.bin")

		if owner != "" {
			t.Fatalf("got owner = %#v, want empty string", owner)
		}
	})

	t.Run("file in dir artifact", func(t *testing.T) {
		targetArt := artifact.Artifact{Path: "foo", IsDir: true, DisableRecursion: true}
		idx := Index{
			"foo.yaml": &stage.Stage{
				Outputs: map[string]*artifact.Artifact{
					"foo": &targetArt,
				},
			},
		}

		owner, foundArt := idx.findOwner("foo/bar.bin")

		if owner != "foo.yaml" {
			t.Fatalf("got owner = %#v, want foo.yaml", owner)
		}

		if diff := cmp.Diff(&targetArt, foundArt); diff != "" {
			t.Fatalf("artifact -want +got:\n%s", diff)
		}
	})

	t.Run("working dir doesn't affect artifact paths", func(t *testing.T) {
		targetArt := artifact.Artifact{Path: "bar.bin"}
		idx := Index{
			"foo.yaml": &stage.Stage{
				WorkingDir: "foo",
				Outputs: map[string]*artifact.Artifact{
					"foo/bar.bin": &targetArt,
				},
			},
		}

		owner, foundArt := idx.findOwner("foo/bar.bin")

		if owner != "foo.yaml" {
			t.Fatalf("got owner = %#v, want foo.yaml", owner)
		}

		if diff := cmp.Diff(&targetArt, foundArt); diff != "" {
			t.Fatalf("artifact -want +got:\n%s", diff)
		}
	})

	t.Run("file in sub-dir of non-recursive dir artifact", func(t *testing.T) {
		idx := Index{
			"foo.yaml": &stage.Stage{
				Outputs: map[string]*artifact.Artifact{
					"foo": {Path: "foo", IsDir: true, DisableRecursion: true},
				},
			},
		}

		owner, _ := idx.findOwner("foo/bar/test.bin")

		if owner != "" {
			t.Fatalf("got owner = %#v, want empty string", owner)
		}
	})

	t.Run("file in sub-dir of recursive dir artifact", func(t *testing.T) {
		targetArt := artifact.Artifact{Path: "foo", IsDir: true}
		idx := Index{
			"foo.yaml": &stage.Stage{
				Outputs: map[string]*artifact.Artifact{
					"foo": &targetArt,
				},
			},
		}

		owner, foundArt := idx.findOwner("foo/bar/test.bin")

		if owner != "foo.yaml" {
			t.Fatalf("got owner = %#v, want foo.yaml", owner)
		}

		if diff := cmp.Diff(&targetArt, foundArt); diff != "" {
			t.Fatalf("artifact -want +got:\n%s", diff)
		}
	})
}
