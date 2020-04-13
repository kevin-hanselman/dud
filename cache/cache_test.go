package cache

import (
	"github.com/kevlar1818/duc/artifact"
	"path"
	"testing"
)

func TestCachePathForChecksum(t *testing.T) {
	ch := LocalCache{Dir: "foo"}
	art := artifact.Artifact{Checksum: "123456789"}

	cachePath, err := ch.CachePathForChecksum(art.Checksum)
	if err != nil {
		t.Fatal(err)
	}

	want := path.Join("foo", "12", "3456789")
	if cachePath != want {
		t.Fatalf("cache.CachePathForChecksum(%#v) = %#v, want %#v", art.Checksum, cachePath, want)
	}

	art.Checksum = ""

	_, err = ch.CachePathForChecksum(art.Checksum)
	if err == nil {
		t.Fatalf("expected error for cache.CachePathForChecksum(%#v)", art.Checksum)
	}
}
