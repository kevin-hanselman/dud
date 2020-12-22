package artifact

import (
	"testing"

	"github.com/kevin-hanselman/dud/src/fsutil"
)

func TestArtifactStatusString(t *testing.T) {
	t.Run("regular file cached up-to-date", func(t *testing.T) {
		testStatus(
			Status{WorkspaceFileStatus: fsutil.StatusRegularFile, HasChecksum: true, ChecksumInCache: true, ContentsMatch: true},
			false,
			false,
			"up-to-date",
			t,
		)
	})

	t.Run("regular file not cached up-to-date", func(t *testing.T) {
		testStatus(
			Status{WorkspaceFileStatus: fsutil.StatusRegularFile, HasChecksum: true, ChecksumInCache: false, ContentsMatch: true},
			true,
			false,
			"up-to-date (not cached)",
			t,
		)
	})

	t.Run("regular file cached modified", func(t *testing.T) {
		testStatus(
			Status{WorkspaceFileStatus: fsutil.StatusRegularFile, HasChecksum: true, ChecksumInCache: true, ContentsMatch: false},
			false,
			false,
			"modified",
			t,
		)
	})

	t.Run("regular file not cached modified", func(t *testing.T) {
		testStatus(
			Status{WorkspaceFileStatus: fsutil.StatusRegularFile, HasChecksum: true, ChecksumInCache: false, ContentsMatch: false},
			true,
			false,
			"modified (not cached)",
			t,
		)
	})

	t.Run("regular file but IsDir true", func(t *testing.T) {
		testStatus(
			Status{WorkspaceFileStatus: fsutil.StatusRegularFile, HasChecksum: true, ChecksumInCache: true, ContentsMatch: true},
			true,
			true,
			"incorrect file type: RegularFile",
			t,
		)
	})

	t.Run("directory but IsDir false", func(t *testing.T) {
		testStatus(
			Status{WorkspaceFileStatus: fsutil.StatusDirectory, HasChecksum: true, ChecksumInCache: true, ContentsMatch: true},
			false,
			false,
			"incorrect file type: Directory",
			t,
		)
	})

	t.Run("directory but SkipCache true", func(t *testing.T) {
		testStatus(
			Status{WorkspaceFileStatus: fsutil.StatusDirectory, HasChecksum: true, ChecksumInCache: true, ContentsMatch: true},
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
