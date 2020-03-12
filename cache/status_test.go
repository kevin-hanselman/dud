package cache

import (
	"github.com/kevlar1818/duc/strategy"
	"github.com/kevlar1818/duc/testutil"
	"os"
	"testing"
)

func TestStatusIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Run("TODO", func(t *testing.T) { testCommitIntegration(strategy.LinkStrategy, t) })
}

func testStatusIntegration(strat strategy.CheckoutStrategy, t *testing.T) {
	dirs, art, err := testutil.CreateArtifactTestCase(
		testutil.TestCaseArgs{InCache: true, WorkspaceFile: testutil.IsRegularFile},
	)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dirs.CacheDir)
	defer os.RemoveAll(dirs.WorkDir)
	cache := LocalCache{Dir: dirs.CacheDir}

	stat, err := cache.Status(dirs.WorkDir, art, strat)
	want := "up to date"
	if err != nil {
		t.Fatal(err)
	}
	if stat != want {
		t.Fatalf("Status() = %#v, want %#v", stat, want)
	}
}
