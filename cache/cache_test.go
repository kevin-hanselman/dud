package cache

import (
	"github.com/kevlar1818/duc/artifact"
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
