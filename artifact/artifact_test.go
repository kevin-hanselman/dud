package artifact

import (
	"github.com/google/go-cmp/cmp"
	"github.com/kevlar1818/duc/fsutil"
	"testing"
)

func TestFromPath(t *testing.T) {

	t.Run("regular file", func(t *testing.T) {
		testFromPath(fsutil.RegularFile, false, false, t)
	})

	t.Run("regular file ignores recursive flag", func(t *testing.T) {
		testFromPath(fsutil.RegularFile, true, false, t)
	})

	t.Run("recursive dir", func(t *testing.T) {
		testFromPath(fsutil.Directory, true, false, t)
	})

	t.Run("non-recursive dir", func(t *testing.T) {
		testFromPath(fsutil.Directory, false, false, t)
	})

	t.Run("error if absent", func(t *testing.T) {
		testFromPath(fsutil.Absent, false, true, t)
	})

	t.Run("error if link", func(t *testing.T) {
		testFromPath(fsutil.Link, false, true, t)
	})

	t.Run("error if other", func(t *testing.T) {
		testFromPath(fsutil.Other, false, true, t)
	})
}

func testFromPath(fileStatus fsutil.FileStatus, isRecursive bool, expectError bool, t *testing.T) {
	fileStatusFromPathOrig := fileStatusFromPath
	fileStatusFromPath = func(path string) (fsutil.FileStatus, error) {
		return fileStatus, nil
	}
	defer func() { fileStatusFromPath = fileStatusFromPathOrig }()

	path := "foobar"
	expectedArtifact := Artifact{
		Path: path,
	}
	if fileStatus == fsutil.Directory {
		expectedArtifact.IsDir = true
		expectedArtifact.IsRecursive = isRecursive
	}

	actualArtifact, err := FromPath(path, isRecursive)
	gotError := err != nil
	if gotError != expectError {
		t.Fatalf("expectError = %v, got %#v", expectError, err)
	}
	if gotError {
		return
	}

	if diff := cmp.Diff(expectedArtifact, actualArtifact); diff != "" {
		t.Fatalf("FromPath() -want +got:\n%s", diff)
	}
}
