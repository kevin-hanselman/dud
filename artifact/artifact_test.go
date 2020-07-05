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

func TestArtifactStatusString(t *testing.T) {
	t.Run("regular file cached up-to-date", func(t *testing.T) {
		testStatus(
			Status{WorkspaceFileStatus: fsutil.RegularFile, HasChecksum: true, ChecksumInCache: true, ContentsMatch: true},
			false,
			false,
			"up-to-date",
			t,
		)
	})

	t.Run("regular file not cached up-to-date", func(t *testing.T) {
		testStatus(
			Status{WorkspaceFileStatus: fsutil.RegularFile, HasChecksum: true, ChecksumInCache: false, ContentsMatch: true},
			true,
			false,
			"up-to-date (not cached)",
			t,
		)
	})

	t.Run("regular file cached modified", func(t *testing.T) {
		testStatus(
			Status{WorkspaceFileStatus: fsutil.RegularFile, HasChecksum: true, ChecksumInCache: true, ContentsMatch: false},
			false,
			false,
			"modified",
			t,
		)
	})

	t.Run("regular file not cached modified", func(t *testing.T) {
		testStatus(
			Status{WorkspaceFileStatus: fsutil.RegularFile, HasChecksum: true, ChecksumInCache: false, ContentsMatch: false},
			true,
			false,
			"modified (not cached)",
			t,
		)
	})

	t.Run("regular file but IsDir true", func(t *testing.T) {
		testStatus(
			Status{WorkspaceFileStatus: fsutil.RegularFile, HasChecksum: true, ChecksumInCache: true, ContentsMatch: true},
			true,
			true,
			"incorrect file type: RegularFile",
			t,
		)
	})

	t.Run("directory but IsDir false", func(t *testing.T) {
		testStatus(
			Status{WorkspaceFileStatus: fsutil.Directory, HasChecksum: true, ChecksumInCache: true, ContentsMatch: true},
			false,
			false,
			"incorrect file type: Directory",
			t,
		)
	})

	t.Run("directory but SkipCache true", func(t *testing.T) {
		testStatus(
			Status{WorkspaceFileStatus: fsutil.Directory, HasChecksum: true, ChecksumInCache: true, ContentsMatch: true},
			true,
			true,
			"incorrect file type: Directory (not cached)",
			t,
		)
	})
}

func testStatus(status Status, skipCache, isDir bool, want string, t *testing.T) {
	artWithStatus := ArtifactWithStatus{
		Artifact: Artifact{SkipCache: skipCache, IsDir: isDir},
		Status:   status,
	}
	got := artWithStatus.String()
	if got != want {
		t.Fatalf("ArtifactWithStatus.String() got %#v, want %#v", got, want)
	}
}
