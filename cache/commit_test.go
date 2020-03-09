package cache

import (
	"github.com/kevlar1818/duc/artifact"
	"github.com/kevlar1818/duc/fsutil"
	"github.com/kevlar1818/duc/strategy"
	"github.com/kevlar1818/duc/testutil"
	"io/ioutil"
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
	dirs, err := testutil.CreateTempDirs()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dirs.CacheDir)
	defer os.RemoveAll(dirs.WorkDir)
	cache := LocalCache{Dir: dirs.CacheDir}

	fileWorkspacePath := path.Join(dirs.WorkDir, "foo.txt")
	if err = ioutil.WriteFile(fileWorkspacePath, []byte("Hello, World!"), 0644); err != nil {
		t.Fatal(err)
	}
	fileChecksum := "0a0a9f2a6772942557ab5355d76af442f8f65e01"
	fileCachePath := path.Join(dirs.CacheDir, fileChecksum[:2], fileChecksum[2:])

	art := artifact.Artifact{
		Checksum: "",
		Path:     "foo.txt",
	}

	if err := cache.Commit(dirs.WorkDir, &art, strat); err != nil {
		t.Fatal(err)
	}

	if art.Checksum != fileChecksum {
		t.Fatalf("artifact.Commit checksum = %#v, expected %#v", art.Checksum, fileChecksum)
	}

	exists, err := fsutil.Exists(fileWorkspacePath)
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
