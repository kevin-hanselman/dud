package cache

import (
	"github.com/kevlar1818/duc/artifact"
	"github.com/kevlar1818/duc/strategy"
	"github.com/kevlar1818/duc/testutil"
	"io/ioutil"
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
	cacheDir, workDir, err := testutil.CreateTempDirs()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(cacheDir)
	defer os.RemoveAll(workDir)
	cache := LocalCache{dir: cacheDir}

	fileChecksum := "0a0a9f2a6772942557ab5355d76af442f8f65e01"
	fileCacheDir := path.Join(cacheDir, fileChecksum[:2])
	fileCachePath := path.Join(fileCacheDir, fileChecksum[2:])
	fileWorkspacePath := path.Join(workDir, "foo.txt")
	if err := os.Mkdir(fileCacheDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err = ioutil.WriteFile(fileCachePath, []byte("Hello, World!"), 0444); err != nil {
		t.Fatal(err)
	}

	art := artifact.Artifact{Checksum: fileChecksum, Path: "foo.txt"}

	if err := cache.Checkout(workDir, &art, strat); err != nil {
		t.Fatal(err)
	}

	assertCheckoutExpectations(strat, fileWorkspacePath, fileCachePath, t)
}
