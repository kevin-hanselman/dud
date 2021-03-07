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

		want := filepath.Join("12", "3456789")
		if cachePath != want {
			t.Fatalf("cache.PathForChecksum(%#v) = %#v, want %#v", checksum, cachePath, want)
		}

		checksum = ""

		_, err = ch.PathForChecksum(checksum)
		if err == nil {
			t.Fatalf("expected error for cache.PathForChecksum(%#v)", checksum)
		}
	})

	t.Run("reject empty paths", func(t *testing.T) {
		_, err := NewLocalCache("")
		if err == nil {
			t.Fatal("expected NewLocalCache to raise an error")
		}
	})
}
