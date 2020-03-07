package artifact

import (
	"github.com/kevlar1818/duc/cache"
	"github.com/kevlar1818/duc/fsutil"
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
	t.Run("Copy", func(t *testing.T) { testCheckoutIntegration(cache.CopyStrategy, t) })
	t.Run("Link", func(t *testing.T) { testCheckoutIntegration(cache.LinkStrategy, t) })
}

func testCheckoutIntegration(strategy cache.CheckoutStrategy, t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	cacheDir, workDir, err := testutil.CreateTempDirs()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(cacheDir)
	defer os.RemoveAll(workDir)

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

	art := Artifact{Checksum: fileChecksum, Path: "foo.txt"}

	if err := art.Checkout(workDir, cacheDir, strategy); err != nil {
		t.Fatal(err)
	}

	assertCheckoutExpectations(strategy, fileWorkspacePath, fileCachePath, t)
}

func TestCommitIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Run("Copy", func(t *testing.T) { testCommitIntegration(cache.CopyStrategy, t) })
	t.Run("Link", func(t *testing.T) { testCommitIntegration(cache.LinkStrategy, t) })
}

func testCommitIntegration(strategy cache.CheckoutStrategy, t *testing.T) {
	cacheDir, workDir, err := testutil.CreateTempDirs()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(cacheDir)
	defer os.RemoveAll(workDir)

	fileWorkspacePath := path.Join(workDir, "foo.txt")
	if err = ioutil.WriteFile(fileWorkspacePath, []byte("Hello, World!"), 0644); err != nil {
		t.Fatal(err)
	}
	fileChecksum := "0a0a9f2a6772942557ab5355d76af442f8f65e01"
	fileCachePath := path.Join(cacheDir, fileChecksum[:2], fileChecksum[2:])

	art := Artifact{
		Checksum: "",
		Path:     "foo.txt",
	}

	if err := art.Commit(workDir, cacheDir, strategy); err != nil {
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

	assertCheckoutExpectations(strategy, fileWorkspacePath, fileCachePath, t)
}

func assertCheckoutExpectations(strategy cache.CheckoutStrategy, fileWorkspacePath, fileCachePath string, t *testing.T) {
	switch strategy {
	case cache.CopyStrategy:
		// check that files are distinct, but have the same contents
		sameFile, err := fsutil.SameFile(fileWorkspacePath, fileCachePath)
		if err != nil {
			t.Fatal(err)
		}
		if sameFile {
			t.Fatalf(
				"files %#v and %#v should not be the same",
				fileWorkspacePath,
				fileCachePath,
			)
		}
		sameContents, err := fsutil.SameContents(fileWorkspacePath, fileCachePath)
		if err != nil {
			t.Fatal(err)
		}
		if !sameContents {
			t.Fatalf(
				"files %#v and %#v should have the same contents",
				fileWorkspacePath,
				fileCachePath,
			)
		}
	case cache.LinkStrategy:
		// check that workspace file is a link to cache file
		sameFile, err := fsutil.SameFile(fileWorkspacePath, fileCachePath)
		if err != nil {
			t.Fatal(err)
		}
		if !sameFile {
			t.Fatalf(
				"files %#v and %#v should be the same file",
				fileWorkspacePath,
				fileCachePath,
			)
		}
		linkDst, err := os.Readlink(fileWorkspacePath)
		if err != nil {
			t.Fatal(err)
		}
		if linkDst != fileCachePath {
			t.Fatalf(
				"file %#v links to %#v, want %#v",
				fileWorkspacePath,
				linkDst,
				fileCachePath,
			)
		}
	}
}
