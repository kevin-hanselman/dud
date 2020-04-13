package cache

import (
	"github.com/kevlar1818/duc/artifact"
	"path"
	"testing"
)

func TestPathForChecksum(t *testing.T) {
	ch := LocalCache{Dir: "foo"}
	art := artifact.Artifact{Checksum: "123456789"}

	cachePath, err := ch.PathForChecksum(art.Checksum)
	if err != nil {
		t.Fatal(err)
	}

	want := path.Join("foo", "12", "3456789")
	if cachePath != want {
		t.Fatalf("cache.PathForChecksum(%#v) = %#v, want %#v", art.Checksum, cachePath, want)
	}

	art.Checksum = ""

	_, err = ch.PathForChecksum(art.Checksum)
	if err == nil {
		t.Fatalf("expected error for cache.PathForChecksum(%#v)", art.Checksum)
	}
}
