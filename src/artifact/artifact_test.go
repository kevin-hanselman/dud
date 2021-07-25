package artifact

import (
	"testing"

	"github.com/kevin-hanselman/dud/src/fsutil"
)

func TestArtifactStatusString(t *testing.T) {
	t.Run("regular file cached up-to-date", func(t *testing.T) {
		aws := ArtifactWithStatus{
			Artifact: Artifact{SkipCache: false, IsDir: false},
			Status: Status{
				WorkspaceFileStatus: fsutil.StatusRegularFile,
				HasChecksum:         true,
				ChecksumInCache:     true,
				ContentsMatch:       true,
			},
		}

		want := "up-to-date"

		got := aws.String()
		if got != want {
			t.Fatalf("ArtifactWithStatus.String() got %#v, want %#v", got, want)
		}
	})

	t.Run("regular file not cached up-to-date", func(t *testing.T) {
		aws := ArtifactWithStatus{
			Artifact: Artifact{SkipCache: true, IsDir: false},
			Status: Status{
				WorkspaceFileStatus: fsutil.StatusRegularFile,
				HasChecksum:         true,
				ChecksumInCache:     false,
				ContentsMatch:       true,
			},
		}

		want := "up-to-date (not cached)"

		got := aws.String()
		if got != want {
			t.Fatalf("ArtifactWithStatus.String() got %#v, want %#v", got, want)
		}
	})

	t.Run("regular file cached modified", func(t *testing.T) {
		aws := ArtifactWithStatus{
			Artifact: Artifact{SkipCache: false, IsDir: false},
			Status: Status{
				WorkspaceFileStatus: fsutil.StatusRegularFile,
				HasChecksum:         true,
				ChecksumInCache:     true,
				ContentsMatch:       false,
			},
		}

		want := "modified"

		got := aws.String()
		if got != want {
			t.Fatalf("ArtifactWithStatus.String() got %#v, want %#v", got, want)
		}
	})

	t.Run("regular file not cached modified", func(t *testing.T) {
		aws := ArtifactWithStatus{
			Artifact: Artifact{SkipCache: true, IsDir: false},
			Status: Status{
				WorkspaceFileStatus: fsutil.StatusRegularFile,
				HasChecksum:         true,
				ChecksumInCache:     true,
				ContentsMatch:       false,
			},
		}

		want := "modified (not cached)"

		got := aws.String()
		if got != want {
			t.Fatalf("ArtifactWithStatus.String() got %#v, want %#v", got, want)
		}
	})

	t.Run("regular file but IsDir true", func(t *testing.T) {
		aws := ArtifactWithStatus{
			Artifact: Artifact{SkipCache: false, IsDir: true},
			Status: Status{
				WorkspaceFileStatus: fsutil.StatusRegularFile,
				HasChecksum:         true,
				ChecksumInCache:     true,
				ContentsMatch:       true,
			},
		}

		want := "incorrect file type: regular file"

		got := aws.String()
		if got != want {
			t.Fatalf("ArtifactWithStatus.String() got %#v, want %#v", got, want)
		}
	})

	t.Run("directory but IsDir false", func(t *testing.T) {
		aws := ArtifactWithStatus{
			Artifact: Artifact{SkipCache: false, IsDir: false},
			Status: Status{
				WorkspaceFileStatus: fsutil.StatusDirectory,
				HasChecksum:         true,
				ChecksumInCache:     true,
				ContentsMatch:       true,
			},
		}

		want := "incorrect file type: directory"

		got := aws.String()
		if got != want {
			t.Fatalf("ArtifactWithStatus.String() got %#v, want %#v", got, want)
		}
	})

	t.Run("directory missing from workspace", func(t *testing.T) {
		aws := ArtifactWithStatus{
			Artifact: Artifact{SkipCache: false, IsDir: true},
			Status: Status{
				WorkspaceFileStatus: fsutil.StatusAbsent,
				HasChecksum:         false,
				ChecksumInCache:     false,
				ContentsMatch:       false,
			},
		}

		want := "missing and not committed"

		got := aws.String()
		if got != want {
			t.Fatalf("ArtifactWithStatus.String() got %#v, want %#v", got, want)
		}
	})

	t.Run("directory but SkipCache true", func(t *testing.T) {
		aws := ArtifactWithStatus{
			Artifact: Artifact{SkipCache: true, IsDir: true},
			Status: Status{
				WorkspaceFileStatus: fsutil.StatusDirectory,
				HasChecksum:         true,
				ChecksumInCache:     true,
				ContentsMatch:       true,
			},
		}

		want := "incorrect file type: directory (not cached)"

		got := aws.String()
		if got != want {
			t.Fatalf("ArtifactWithStatus.String() got %#v, want %#v", got, want)
		}
	})
}
