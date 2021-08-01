package artifact

import (
	"testing"

	"github.com/kevin-hanselman/dud/src/fsutil"
)

func TestArtifactStatusString(t *testing.T) {
	t.Run("regular file cached up-to-date", func(t *testing.T) {
		status := Status{
			Artifact:            Artifact{SkipCache: false, IsDir: false},
			WorkspaceFileStatus: fsutil.StatusRegularFile,
			HasChecksum:         true,
			ChecksumInCache:     true,
			ContentsMatch:       true,
		}

		want := "up-to-date"

		got := status.String()
		if got != want {
			t.Fatalf("Status.String() got %#v, want %#v", got, want)
		}
	})

	t.Run("regular file not cached up-to-date", func(t *testing.T) {
		status := Status{
			Artifact:            Artifact{SkipCache: true, IsDir: false},
			WorkspaceFileStatus: fsutil.StatusRegularFile,
			HasChecksum:         true,
			ChecksumInCache:     false,
			ContentsMatch:       true,
		}

		want := "up-to-date (not cached)"

		got := status.String()
		if got != want {
			t.Fatalf("Status.String() got %#v, want %#v", got, want)
		}
	})

	t.Run("regular file cached modified", func(t *testing.T) {
		status := Status{
			Artifact:            Artifact{SkipCache: false, IsDir: false},
			WorkspaceFileStatus: fsutil.StatusRegularFile,
			HasChecksum:         true,
			ChecksumInCache:     true,
			ContentsMatch:       false,
		}

		want := "modified"

		got := status.String()
		if got != want {
			t.Fatalf("Status.String() got %#v, want %#v", got, want)
		}
	})

	t.Run("regular file not cached modified", func(t *testing.T) {
		status := Status{
			Artifact:            Artifact{SkipCache: true, IsDir: false},
			WorkspaceFileStatus: fsutil.StatusRegularFile,
			HasChecksum:         true,
			ChecksumInCache:     true,
			ContentsMatch:       false,
		}

		want := "modified (not cached)"

		got := status.String()
		if got != want {
			t.Fatalf("Status.String() got %#v, want %#v", got, want)
		}
	})

	t.Run("regular file but IsDir true", func(t *testing.T) {
		status := Status{
			Artifact:            Artifact{SkipCache: false, IsDir: true},
			WorkspaceFileStatus: fsutil.StatusRegularFile,
			HasChecksum:         true,
			ChecksumInCache:     true,
			ContentsMatch:       true,
		}

		want := "incorrect file type: regular file"

		got := status.String()
		if got != want {
			t.Fatalf("Status.String() got %#v, want %#v", got, want)
		}
	})

	t.Run("up-to-date recursive directory", func(t *testing.T) {
		upToDate := Status{
			WorkspaceFileStatus: fsutil.StatusRegularFile,
			HasChecksum:         true,
			ChecksumInCache:     true,
			ContentsMatch:       true,
		}
		status := Status{
			Artifact:            Artifact{IsDir: true},
			WorkspaceFileStatus: fsutil.StatusDirectory,
			HasChecksum:         true,
			ChecksumInCache:     true,
			ContentsMatch:       true,
			ChildrenStatus: map[string]*Status{
				"a": &upToDate,
				"b": &upToDate,
				"c": {
					Artifact:            Artifact{IsDir: true},
					WorkspaceFileStatus: fsutil.StatusDirectory,
					HasChecksum:         true,
					ChecksumInCache:     true,
					ContentsMatch:       false,
					ChildrenStatus: map[string]*Status{
						"d": &upToDate,
						"e": {
							WorkspaceFileStatus: fsutil.StatusRegularFile,
							HasChecksum:         false,
							ChecksumInCache:     false,
							ContentsMatch:       false,
						},
					},
				},
			},
		}

		want := "x3 up-to-date, x1 not committed"

		got := status.String()
		if got != want {
			t.Fatalf("Status.String() got %#v, want %#v", got, want)
		}
	})

	t.Run("empty directory", func(t *testing.T) {
		status := Status{
			Artifact:            Artifact{IsDir: true},
			WorkspaceFileStatus: fsutil.StatusDirectory,
			HasChecksum:         true,
			ChecksumInCache:     true,
			ContentsMatch:       false,
		}

		want := "empty directory"

		got := status.String()
		if got != want {
			t.Fatalf("Status.String() got %#v, want %#v", got, want)
		}
	})

	t.Run("directory but IsDir false", func(t *testing.T) {
		status := Status{
			Artifact:            Artifact{IsDir: false},
			WorkspaceFileStatus: fsutil.StatusDirectory,
			HasChecksum:         true,
			ChecksumInCache:     true,
			ContentsMatch:       true,
		}

		want := "incorrect file type: directory"

		got := status.String()
		if got != want {
			t.Fatalf("Status.String() got %#v, want %#v", got, want)
		}
	})

	t.Run("directory missing from workspace", func(t *testing.T) {
		status := Status{
			Artifact:            Artifact{SkipCache: false, IsDir: true},
			WorkspaceFileStatus: fsutil.StatusAbsent,
			HasChecksum:         false,
			ChecksumInCache:     false,
			ContentsMatch:       false,
		}

		want := "missing and not committed"

		got := status.String()
		if got != want {
			t.Fatalf("Status.String() got %#v, want %#v", got, want)
		}
	})

	t.Run("directory but SkipCache true", func(t *testing.T) {
		status := Status{
			Artifact:            Artifact{SkipCache: true, IsDir: true},
			WorkspaceFileStatus: fsutil.StatusDirectory,
			HasChecksum:         true,
			ChecksumInCache:     true,
			ContentsMatch:       true,
		}

		want := "incorrect file type: directory (not cached)"

		got := status.String()
		if got != want {
			t.Fatalf("Status.String() got %#v, want %#v", got, want)
		}
	})
}
