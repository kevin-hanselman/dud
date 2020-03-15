package cache

import (
	"github.com/kevlar1818/duc/artifact"
	"github.com/kevlar1818/duc/strategy"
	"github.com/kevlar1818/duc/testutil"
	"os"
	"path"
	"testing"
)

func TestCheckoutIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Run("Copy", func(t *testing.T) { testCheckoutIntegration(strategy.CopyStrategy, t) })
	t.Run("Link", func(t *testing.T) { testCheckoutIntegration(strategy.LinkStrategy, t) })
}

func testCheckoutIntegration(strat strategy.CheckoutStrategy, t *testing.T) {
	dirs, art, err := testutil.CreateArtifactTestCase(
		artifact.Status{ChecksumInCache: true, WorkspaceStatus: artifact.Absent},
	)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dirs.CacheDir)
	defer os.RemoveAll(dirs.WorkDir)
	cache := LocalCache{Dir: dirs.CacheDir}

	fileWorkspacePath := path.Join(dirs.WorkDir, art.Path)
	fileCachePath, err := cache.CachePathForArtifact(art)
	if err != nil {
		t.Fatal(err)
	}

	if err := cache.Checkout(dirs.WorkDir, &art, strat); err != nil {
		t.Fatal(err)
	}

	assertCheckoutExpectations(strat, fileWorkspacePath, fileCachePath, t)
}
