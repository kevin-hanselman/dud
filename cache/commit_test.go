package cache

import (
	"github.com/google/go-cmp/cmp"
	"github.com/kevlar1818/duc/artifact"
	"github.com/kevlar1818/duc/strategy"
	"github.com/kevlar1818/duc/testutil"
	"os"
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
	dirs, art, err := testutil.CreateArtifactTestCase(
		artifact.Status{HasChecksum: false, ChecksumInCache: false, WorkspaceStatus: artifact.RegularFile},
	)
	defer os.RemoveAll(dirs.CacheDir)
	defer os.RemoveAll(dirs.WorkDir)
	if err != nil {
		t.Fatal(err)
	}
	cache := LocalCache{Dir: dirs.CacheDir}

	if err := cache.Commit(dirs.WorkDir, &art, strat); err != nil {
		t.Fatal(err)
	}

	fileCachePath, err := cache.CachePathForArtifact(art)
	if err != nil {
		t.Fatal(err)
	}

	cachedFileInfo, err := os.Stat(fileCachePath)
	if err != nil {
		t.Fatal(err)
	}
	// TODO: check this in cache.Status?
	if cachedFileInfo.Mode() != 0444 {
		t.Fatalf("%#v has perms %#o, want %#o", fileCachePath, cachedFileInfo.Mode(), 0444)
	}

	statusGot, err := cache.Status(dirs.WorkDir, art)
	if err != nil {
		t.Fatal(err)
	}
	statusWant := artifact.Status{
		HasChecksum:     true,
		ChecksumInCache: true,
		ContentsMatch:   true,
	}
	switch strat {
	case strategy.CopyStrategy:
		statusWant.WorkspaceStatus = artifact.RegularFile
	case strategy.LinkStrategy:
		statusWant.WorkspaceStatus = artifact.Link
	}

	if diff := cmp.Diff(statusWant, statusGot); diff != "" {
		t.Fatalf("Status() -want +got:\n%s", diff)
	}
}
