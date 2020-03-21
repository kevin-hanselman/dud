package cache

import (
	"github.com/google/go-cmp/cmp"
	"github.com/kevlar1818/duc/artifact"
	"github.com/kevlar1818/duc/strategy"
	"github.com/kevlar1818/duc/testutil"
	"os"
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
		artifact.Status{HasChecksum: true, ChecksumInCache: true, WorkspaceStatus: artifact.Absent},
	)
	defer os.RemoveAll(dirs.CacheDir)
	defer os.RemoveAll(dirs.WorkDir)
	if err != nil {
		t.Fatal(err)
	}
	cache := LocalCache{Dir: dirs.CacheDir}

	if err := cache.Checkout(dirs.WorkDir, &art, strat); err != nil {
		t.Fatal(err)
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
