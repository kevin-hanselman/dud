package cache

import (
	"github.com/kevlar1818/duc/artifact"
	"github.com/kevlar1818/duc/fsutil"
	"github.com/kevlar1818/duc/strategy"
	"os"
	"path"
	"testing"
)

func TestCachePathForArtifact(t *testing.T) {
	ch := LocalCache{Dir: "foo"}
	art := artifact.Artifact{Checksum: "123456789"}

	cachePath, err := ch.CachePathForArtifact(art)
	if err != nil {
		t.Fatal(err)
	}

	want := path.Join("foo", "12", "3456789")
	if cachePath != want {
		t.Fatalf("cache.CachePathForArtifact(%#v) = %#v, want %#v", art, cachePath, want)
	}

	art.Checksum = ""

	cachePath, err = ch.CachePathForArtifact(art)
	if err == nil {
		t.Fatalf("expected error for cache.CachePathForArtifact(%#v)", art)
	}
}

func assertCheckoutExpectations(strat strategy.CheckoutStrategy, fileWorkspacePath, fileCachePath string, t *testing.T) {
	switch strat {
	case strategy.CopyStrategy:
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
	case strategy.LinkStrategy:
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
