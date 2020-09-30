package cache

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kevin-hanselman/duc/artifact"
	"github.com/kevin-hanselman/duc/fsutil"
	"github.com/kevin-hanselman/duc/strategy"
)

func TestDirectoryCheckoutIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Run("committed and absent from workspace", func(t *testing.T) {
		dirs, art, cache := setupDirTest(t)
		defer os.RemoveAll(dirs.CacheDir)
		defer os.RemoveAll(dirs.WorkDir)

		if err := cache.Commit(dirs.WorkDir, &art, strategy.LinkStrategy); err != nil {
			t.Fatal(err)
		}

		if err := os.RemoveAll(filepath.Join(dirs.WorkDir, art.Path)); err != nil {
			t.Fatal(err)
		}

		if err := cache.Checkout(dirs.WorkDir, &art, strategy.LinkStrategy); err != nil {
			t.Fatal(err)
		}

		actualStatus, err := cache.Status(dirs.WorkDir, art)
		if err != nil {
			t.Fatal(err)
		}

		expectedStatus := artifact.Status{
			WorkspaceFileStatus: fsutil.Directory,
			HasChecksum:         true,
			ChecksumInCache:     true,
			ContentsMatch:       true,
		}

		if diff := cmp.Diff(expectedStatus, actualStatus.Status); diff != "" {
			t.Fatalf("Status -want +got:\n%s", diff)
		}
	})
}
