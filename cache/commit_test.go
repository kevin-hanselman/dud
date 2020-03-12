package cache

import (
	"github.com/kevlar1818/duc/fsutil"
	"github.com/kevlar1818/duc/strategy"
	"github.com/kevlar1818/duc/testutil"
	"os"
	"path"
	"testing"
)

func TestCommitIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Run("Copy", func(t *testing.T) { testCommitIntegration(strategy.CopyStrategy, t) })
	t.Run("Link", func(t *testing.T) { testCommitIntegration(strategy.LinkStrategy, t) })
}

func testCommitIntegration(strat strategy.CheckoutStrategy, t *testing.T) {
	dirs, art, err := testutil.CreateArtifactTestCase(false, testutil.IsRegularFile)
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

	if err := cache.Commit(dirs.WorkDir, &art, strat); err != nil {
		t.Fatal(err)
	}

	if art.Checksum != art.Checksum {
		t.Fatalf("artifact.Commit checksum = %#v, expected %#v", art.Checksum, art.Checksum)
	}

	exists, err := fsutil.Exists(fileWorkspacePath, false)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatalf("file %#v should exist", fileWorkspacePath)
	}
	cachedFileInfo, err := os.Stat(fileCachePath)
	if err != nil {
		t.Fatal(err)
	}
	if cachedFileInfo.Mode() != 0444 {
		t.Fatalf("%#v has perms %#o, want %#o", fileCachePath, cachedFileInfo.Mode(), 0444)
	}

	assertCheckoutExpectations(strat, fileWorkspacePath, fileCachePath, t)
}