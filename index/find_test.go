package index

import (
	"github.com/google/go-cmp/cmp"
	"github.com/kevin-hanselman/duc/artifact"
	"github.com/kevin-hanselman/duc/stage"

	"testing"
)

func TestFindOwner(t *testing.T) {

	t.Run("single stage", func(t *testing.T) {
		idx := make(Index)
		expectedEntry := &entry{
			Stage: stage.Stage{
				Outputs: map[string]*artifact.Artifact{
					"foo.bin": {Path: "foo.bin"},
					"bar.bin": {Path: "bar.bin"},
				},
			},
		}
		idx["foo.yaml"] = expectedEntry

		owner, ok, err := idx.findOwner("bar.bin")

		if err != nil {
			t.Fatal(err)
		}

		if !ok {
			t.Fatal("findOwner returned ok = false")
		}

		if diff := cmp.Diff(expectedEntry, owner); diff != "" {
			t.Fatalf("Stage -want +got:\n%s", diff)
		}
	})

	t.Run("no owner", func(t *testing.T) {
		idx := make(Index)
		en := &entry{
			Stage: stage.Stage{
				Outputs: map[string]*artifact.Artifact{
					"foo.bin": {Path: "foo.bin"},
					"bar.bin": {Path: "bar.bin"},
				},
			},
		}
		idx["foo.yaml"] = en

		owner, ok, err := idx.findOwner("other.bin")

		if err != nil {
			t.Fatal(err)
		}

		if ok {
			t.Fatal("findOwner returned ok = true")
		}

		if owner != nil {
			t.Fatal("expected owner to be nil")
		}
	})

	t.Run("file in dir artifact", func(t *testing.T) {
		idx := make(Index)
		expectedEntry := &entry{
			Stage: stage.Stage{
				Outputs: map[string]*artifact.Artifact{
					"foo": {Path: "foo", IsDir: true, IsRecursive: false},
				},
			},
		}
		idx["foo.yaml"] = expectedEntry

		owner, ok, err := idx.findOwner("foo/bar.bin")

		if err != nil {
			t.Fatal(err)
		}

		if !ok {
			t.Fatal("findOwner returned ok = false")
		}

		if diff := cmp.Diff(expectedEntry, owner); diff != "" {
			t.Fatalf("Stage -want +got:\n%s", diff)
		}
	})

	t.Run("non-zero WorkingDir", func(t *testing.T) {
		idx := make(Index)
		expectedEntry := &entry{
			Stage: stage.Stage{
				WorkingDir: "foo",
				Outputs: map[string]*artifact.Artifact{
					"bar.bin": {Path: "bar.bin"},
				},
			},
		}
		idx["foo.yaml"] = expectedEntry

		owner, ok, err := idx.findOwner("foo/bar.bin")

		if err != nil {
			t.Fatal(err)
		}

		if !ok {
			t.Fatal("findOwner returned ok = false")
		}

		if diff := cmp.Diff(expectedEntry, owner); diff != "" {
			t.Fatalf("Stage -want +got:\n%s", diff)
		}
	})

	t.Run("file in sub-dir of non-recursive dir artifact", func(t *testing.T) {
		idx := make(Index)
		expectedEntry := &entry{
			Stage: stage.Stage{
				Outputs: map[string]*artifact.Artifact{
					"foo": {Path: "foo", IsDir: true, IsRecursive: false},
				},
			},
		}
		idx["foo.yaml"] = expectedEntry

		owner, ok, err := idx.findOwner("foo/bar/test.bin")

		if err != nil {
			t.Fatal(err)
		}

		if ok {
			t.Fatal("findOwner returned ok = true")
		}

		if owner != nil {
			t.Fatal("expected owner to be nil")
		}
	})

	t.Run("file in sub-dir of recursive dir artifact", func(t *testing.T) {
		idx := make(Index)
		expectedEntry := &entry{
			Stage: stage.Stage{
				Outputs: map[string]*artifact.Artifact{
					"foo": {Path: "foo", IsDir: true, IsRecursive: true},
				},
			},
		}
		idx["foo.yaml"] = expectedEntry

		owner, ok, err := idx.findOwner("foo/bar/test.bin")

		if err != nil {
			t.Fatal(err)
		}

		if !ok {
			t.Fatal("findOwner returned ok = false")
		}

		if diff := cmp.Diff(expectedEntry, owner); diff != "" {
			t.Fatalf("Stage -want +got:\n%s", diff)
		}
	})
}
