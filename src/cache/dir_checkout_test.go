package cache

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kevin-hanselman/dud/src/agglog"
	"github.com/kevin-hanselman/dud/src/artifact"
	"github.com/kevin-hanselman/dud/src/fsutil"
	"github.com/kevin-hanselman/dud/src/strategy"
)

func TestDirectoryCheckoutIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := agglog.NewNullLogger()

	maxSharedWorkers = 1
	maxDedicatedWorkers = 1

	// TODO: add more tests
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

		progress := newProgress("test")

		if err := cache.Checkout(dirs.WorkDir, art, strategy.LinkStrategy, progress); err != nil {
			t.Fatal(err)
		}

		// Expect the progress bar to report 10 out of 10 files checked out.
		// Directories aren't counted. See setupDirTest for the 10 files.
		if progress.Total() != 10 {
			t.Fatalf("progress.Total() = %v, want 10", progress.Total())
		}
		if progress.Current() != 10 {
			t.Fatalf("progress.Current() = %v, want 10", progress.Current())
		}

		actualStatus, err := cache.Status(dirs.WorkDir, art)
		if err != nil {
			t.Fatal(err)
		}

		expectedStatus := artifact.Status{
			Artifact:            art,
			WorkspaceFileStatus: fsutil.StatusDirectory,
			HasChecksum:         true,
			ChecksumInCache:     true,
			ContentsMatch:       true,
		}

		if diff := cmp.Diff(expectedStatus, actualStatus); diff != "" {
			t.Fatalf("Status -want +got:\n%s", diff)
		}
	})
}
