package cache

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kevin-hanselman/dud/src/agglog"
	"github.com/kevin-hanselman/dud/src/artifact"
	"github.com/kevin-hanselman/dud/src/fsutil"
	"github.com/kevin-hanselman/dud/src/progress"
	"github.com/kevin-hanselman/dud/src/strategy"
	"github.com/kevin-hanselman/dud/src/testutil"
	"go.uber.org/goleak"
)

func TestDirectoryCheckoutIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	defer goleak.VerifyNone(t)

	logger := agglog.NewNullLogger()

	maxSharedWorkers = 1
	maxDedicatedWorkers = 1

	t.Run("committed and absent from workspace", func(t *testing.T) {
		dirs, art, cache := setupDirTest(t)
		defer os.RemoveAll(dirs.CacheDir)
		defer os.RemoveAll(dirs.WorkDir)

		if err := cache.Commit(dirs.WorkDir, &art, strategy.LinkStrategy, logger); err != nil {
			t.Fatal(err)
		}

		if err := os.RemoveAll(filepath.Join(dirs.WorkDir, art.Path)); err != nil {
			t.Fatal(err)
		}

		prog := progress.NewHiddenProgress()

		if err := cache.Checkout(dirs.WorkDir, art, strategy.LinkStrategy, prog, logger); err != nil {
			t.Fatal(err)
		}

		// Expect the progress bar to report 10 out of 10 files checked out.
		// Directories aren't counted. See setupDirTest for the 10 files.
		if total := prog.TotalFiles(); total != 10 {
			t.Fatalf("progress.TotalFiles() = %v, want 10", total)
		}
		if current := prog.CurrentFiles(); current != 10 {
			t.Fatalf("progress.CurrentFiles() = %v, want 10", current)
		}

		actualStatus, err := cache.Status(dirs.WorkDir, art, false)
		if err != nil {
			t.Fatal(err)
		}

		art.Checksum = ""

		upToDate := func(art artifact.Artifact) *artifact.Status {
			return &artifact.Status{
				WorkspaceFileStatus: fsutil.StatusLink,
				Artifact:            art,
				HasChecksum:         true,
				ChecksumInCache:     true,
				ContentsMatch:       true,
			}
		}

		expectedStatus := artifact.Status{
			Artifact:            art,
			WorkspaceFileStatus: fsutil.StatusDirectory,
			HasChecksum:         true,
			ChecksumInCache:     true,
			ContentsMatch:       true,
			ChildrenStatus: map[string]*artifact.Status{
				"1.txt": upToDate(artifact.Artifact{Path: "1.txt"}),
				"2.txt": upToDate(artifact.Artifact{Path: "2.txt"}),
				"3.txt": upToDate(artifact.Artifact{Path: "3.txt"}),
				"4.txt": upToDate(artifact.Artifact{Path: "4.txt"}),
				"5.txt": upToDate(artifact.Artifact{Path: "5.txt"}),
				"bar": {
					Artifact:            artifact.Artifact{Path: "bar", IsDir: true},
					WorkspaceFileStatus: fsutil.StatusDirectory,
					HasChecksum:         true,
					ChecksumInCache:     true,
					ContentsMatch:       true,
					ChildrenStatus: map[string]*artifact.Status{
						"4.txt": upToDate(artifact.Artifact{Path: "4.txt"}),
						"5.txt": upToDate(artifact.Artifact{Path: "5.txt"}),
						"6.txt": upToDate(artifact.Artifact{Path: "6.txt"}),
						"7.txt": upToDate(artifact.Artifact{Path: "7.txt"}),
						"8.txt": upToDate(artifact.Artifact{Path: "8.txt"}),
					},
				},
			},
		}

		assertThenRemoveChecksums(t, &actualStatus)

		if diff := cmp.Diff(expectedStatus, actualStatus); diff != "" {
			t.Fatalf("Status -want +got:\n%s", diff)
		}
	})

	t.Run("empty directory", func(t *testing.T) {
		dirs, err := testutil.CreateTempDirs()
		if err != nil {
			t.Fatal(err)
		}

		cache, err := NewLocalCache(dirs.CacheDir)
		if err != nil {
			t.Fatal(err)
		}

		artFullPath := filepath.Join(dirs.WorkDir, "foo")

		if err := os.MkdirAll(artFullPath, 0o755); err != nil {
			t.Fatal(err)
		}

		art := artifact.Artifact{Path: "foo", IsDir: true}

		if err := cache.Commit(dirs.WorkDir, &art, strategy.LinkStrategy, logger); err != nil {
			t.Fatal(err)
		}

		if err := os.RemoveAll(artFullPath); err != nil {
			t.Fatal(err)
		}

		prog := progress.NewHiddenProgress()

		if err := cache.Checkout(dirs.WorkDir, art, strategy.LinkStrategy, prog, logger); err != nil {
			t.Fatal(err)
		}

		fileInfo, err := os.Lstat(artFullPath)
		if err != nil {
			t.Fatal(err)
		}
		if !fileInfo.IsDir() {
			t.Fatalf("expected %s to be a directory, got %s", artFullPath, fileInfo)
		}
	})
}
