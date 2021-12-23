package cache

import (
	"bytes"
	"os"
	"testing"

	"github.com/kevin-hanselman/dud/src/agglog"
	"github.com/kevin-hanselman/dud/src/artifact"
	"github.com/kevin-hanselman/dud/src/checksum"
	"github.com/kevin-hanselman/dud/src/fsutil"
	"github.com/kevin-hanselman/dud/src/strategy"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/goleak"
)

func TestDirectoryStatusIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	defer goleak.VerifyNone(t)

	notCommitted := func(art artifact.Artifact) *artifact.Status {
		return &artifact.Status{
			WorkspaceFileStatus: fsutil.StatusRegularFile,
			Artifact:            art,
			HasChecksum:         false,
			ChecksumInCache:     false,
			ContentsMatch:       false,
		}
	}

	upToDate := func(art artifact.Artifact, contents string, t *testing.T) *artifact.Status {
		var err error
		buf := bytes.NewBuffer([]byte(contents))
		art.Checksum, err = checksum.Checksum(buf)
		if err != nil {
			t.Fatal(err)
		}
		return &artifact.Status{
			WorkspaceFileStatus: fsutil.StatusLink,
			Artifact:            art,
			HasChecksum:         true,
			ChecksumInCache:     true,
			ContentsMatch:       true,
		}
	}

	logger := agglog.NewNullLogger()

	maxSharedWorkers = 1
	maxDedicatedWorkers = 1

	t.Run("untracked directory", func(t *testing.T) {
		dirs, art, cache := setupDirTest(t)
		defer os.RemoveAll(dirs.CacheDir)
		defer os.RemoveAll(dirs.WorkDir)

		actualStatus, err := cache.Status(dirs.WorkDir, art, false)
		if err != nil {
			t.Fatal(err)
		}

		expectedStatus := artifact.Status{
			Artifact:            art,
			WorkspaceFileStatus: fsutil.StatusDirectory,
			HasChecksum:         false,
			ChecksumInCache:     false,
			ContentsMatch:       false,
			ChildrenStatus: map[string]*artifact.Status{
				"1.txt": notCommitted(artifact.Artifact{Path: "1.txt"}),
				"2.txt": notCommitted(artifact.Artifact{Path: "2.txt"}),
				"3.txt": notCommitted(artifact.Artifact{Path: "3.txt"}),
				"4.txt": notCommitted(artifact.Artifact{Path: "4.txt"}),
				"5.txt": notCommitted(artifact.Artifact{Path: "5.txt"}),
				"bar": {
					Artifact:            artifact.Artifact{Path: "bar", IsDir: true},
					WorkspaceFileStatus: fsutil.StatusDirectory,
					HasChecksum:         false,
					ChecksumInCache:     false,
					ContentsMatch:       false,
					ChildrenStatus: map[string]*artifact.Status{
						"4.txt": notCommitted(artifact.Artifact{Path: "4.txt"}),
						"5.txt": notCommitted(artifact.Artifact{Path: "5.txt"}),
						"6.txt": notCommitted(artifact.Artifact{Path: "6.txt"}),
						"7.txt": notCommitted(artifact.Artifact{Path: "7.txt"}),
						"8.txt": notCommitted(artifact.Artifact{Path: "8.txt"}),
					},
				},
			},
		}

		if diff := cmp.Diff(expectedStatus, actualStatus); diff != "" {
			t.Fatalf("Status -want +got:\n%s", diff)
		}
	})

	t.Run("untracked directory, no recursion", func(t *testing.T) {
		dirs, art, cache := setupDirTest(t)
		defer os.RemoveAll(dirs.CacheDir)
		defer os.RemoveAll(dirs.WorkDir)

		art.DisableRecursion = true
		actualStatus, err := cache.Status(dirs.WorkDir, art, false)
		if err != nil {
			t.Fatal(err)
		}

		expectedStatus := artifact.Status{
			Artifact:            art,
			WorkspaceFileStatus: fsutil.StatusDirectory,
			HasChecksum:         false,
			ChecksumInCache:     false,
			ContentsMatch:       false,
			ChildrenStatus: map[string]*artifact.Status{
				"1.txt": notCommitted(artifact.Artifact{Path: "1.txt"}),
				"2.txt": notCommitted(artifact.Artifact{Path: "2.txt"}),
				"3.txt": notCommitted(artifact.Artifact{Path: "3.txt"}),
				"4.txt": notCommitted(artifact.Artifact{Path: "4.txt"}),
				"5.txt": notCommitted(artifact.Artifact{Path: "5.txt"}),
			},
		}

		if diff := cmp.Diff(expectedStatus, actualStatus); diff != "" {
			t.Fatalf("Status -want +got:\n%s", diff)
		}
	})

	t.Run("untracked directory, short-circuit", func(t *testing.T) {
		dirs, art, cache := setupDirTest(t)
		defer os.RemoveAll(dirs.CacheDir)
		defer os.RemoveAll(dirs.WorkDir)

		actualStatus, err := cache.Status(dirs.WorkDir, art, true)
		if err != nil {
			t.Fatal(err)
		}

		expectedStatus := artifact.Status{
			Artifact:            art,
			WorkspaceFileStatus: fsutil.StatusDirectory,
			HasChecksum:         false,
			ChecksumInCache:     false,
			ContentsMatch:       false,
		}

		if diff := cmp.Diff(expectedStatus, actualStatus); diff != "" {
			t.Fatalf("Status -want +got:\n%s", diff)
		}
	})

	t.Run("untracked sub-directory, short-circuit", func(t *testing.T) {
		dirs, art, cache := setupDirTest(t)
		defer os.RemoveAll(dirs.CacheDir)
		defer os.RemoveAll(dirs.WorkDir)

		// Disable recursion so the sub-dir doesn't get committed. Then enable
		// recursion so status reports the untracked sub-dir.
		art.DisableRecursion = true
		err := cache.Commit(dirs.WorkDir, &art, strategy.LinkStrategy, logger)
		if err != nil {
			t.Fatal(err)
		}

		art.DisableRecursion = false
		actualStatus, err := cache.Status(dirs.WorkDir, art, false)
		if err != nil {
			t.Fatal(err)
		}

		expectedStatus := artifact.Status{
			Artifact:            art,
			WorkspaceFileStatus: fsutil.StatusDirectory,
			HasChecksum:         true,
			ChecksumInCache:     true,
			ContentsMatch:       false,
			ChildrenStatus: map[string]*artifact.Status{
				"1.txt": upToDate(artifact.Artifact{Path: "1.txt"}, "1", t),
				"2.txt": upToDate(artifact.Artifact{Path: "2.txt"}, "2", t),
				"3.txt": upToDate(artifact.Artifact{Path: "3.txt"}, "3", t),
				"4.txt": upToDate(artifact.Artifact{Path: "4.txt"}, "4", t),
				"5.txt": upToDate(artifact.Artifact{Path: "5.txt"}, "5", t),
				"bar": {
					Artifact:            artifact.Artifact{Path: "bar", IsDir: true},
					WorkspaceFileStatus: fsutil.StatusDirectory,
					HasChecksum:         false,
					ChecksumInCache:     false,
					ContentsMatch:       false,
					ChildrenStatus: map[string]*artifact.Status{
						"4.txt": notCommitted(artifact.Artifact{Path: "4.txt"}),
						"5.txt": notCommitted(artifact.Artifact{Path: "5.txt"}),
						"6.txt": notCommitted(artifact.Artifact{Path: "6.txt"}),
						"7.txt": notCommitted(artifact.Artifact{Path: "7.txt"}),
						"8.txt": notCommitted(artifact.Artifact{Path: "8.txt"}),
					},
				},
			},
		}

		if diff := cmp.Diff(expectedStatus, actualStatus); diff != "" {
			t.Fatalf("Status -want +got:\n%s", diff)
		}

		// Enable short-circuiting. Expect the same status without the sub-dir's status included.
		actualStatus, err = cache.Status(dirs.WorkDir, art, true)
		if err != nil {
			t.Fatal(err)
		}
		delete(expectedStatus.ChildrenStatus, "bar")

		if diff := cmp.Diff(expectedStatus, actualStatus); diff != "" {
			t.Fatalf("Status -want +got:\n%s", diff)
		}
	})
}
