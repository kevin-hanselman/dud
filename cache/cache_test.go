package cache

import (
	"path/filepath"
	"testing"
)

func TestPathForChecksum(t *testing.T) {

	t.Run("happy path", func(t *testing.T) {
		ch, err := NewLocalCache("/foo")
		if err != nil {
			t.Fatal(err)
		}

		checksum := "123456789"
		cachePath, err := ch.PathForChecksum(checksum)
		if err != nil {
			t.Fatal(err)
		}

		want := filepath.Join("/foo", "12", "3456789")
		if cachePath != want {
			t.Fatalf("cache.PathForChecksum(%#v) = %#v, want %#v", checksum, cachePath, want)
		}

		checksum = ""

		_, err = ch.PathForChecksum(checksum)
		if err == nil {
			t.Fatalf("expected error for cache.PathForChecksum(%#v)", checksum)
		}
	})

	t.Run("reject relative paths ", func(t *testing.T) {
		path := "foo/bar"
		_, err := NewLocalCache(path)
		if err == nil {
			t.Fatalf("expected NewLocalCache(%#v) to raise an error (not absolute path)", path)
		}
	})
}
