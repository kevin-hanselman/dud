package cache

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kevin-hanselman/dud/src/agglog"
	"github.com/kevin-hanselman/dud/src/artifact"
	"github.com/kevin-hanselman/dud/src/strategy"
	"github.com/kevin-hanselman/dud/src/testutil"
	"github.com/pkg/errors"
)

// See fetch_test.go for various helper functions.

func TestPushIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := agglog.NewNullLogger()

	remoteCopyOrig := remoteCopy
	remoteCopyPanic := func(src, dst string, fileSet map[string]struct{}) error {
		panic("unexpected call to remoteCopy")
	}
	remoteCopy = remoteCopyPanic
	defer func() { remoteCopy = remoteCopyOrig }()

	resetMocks := func() {
		remoteCopy = remoteCopyPanic
	}

	t.Run("push file artifact happy path", func(t *testing.T) {
		defer resetMocks()

		artStatus := artifact.Status{HasChecksum: true, ChecksumInCache: true}

		dirs, art, err := testutil.CreateArtifactTestCase(artStatus)
		defer os.RemoveAll(dirs.CacheDir)
		defer os.RemoveAll(dirs.WorkDir)
		if err != nil {
			t.Fatal(err)
		}

		fakeRemote := filepath.Join(dirs.WorkDir, "fake_remote")

		ch, err := NewLocalCache(dirs.CacheDir)
		if err != nil {
			t.Fatal(err)
		}

		remoteCopy = mockRemoteCopy

		if err := ch.Push(dirs.WorkDir, fakeRemote, art); err != nil {
			t.Fatal(err)
		}

		assertCacheDirsEqual(dirs.CacheDir, fakeRemote, t)
	})

	t.Run("push file artifact returns error if no checksum", func(t *testing.T) {
		defer resetMocks()

		artStatus := artifact.Status{HasChecksum: false}

		dirs, art, err := testutil.CreateArtifactTestCase(artStatus)
		defer os.RemoveAll(dirs.CacheDir)
		defer os.RemoveAll(dirs.WorkDir)
		if err != nil {
			t.Fatal(err)
		}

		ch, err := NewLocalCache(dirs.CacheDir)
		if err != nil {
			t.Fatal(err)
		}

		pushErr := ch.Push(dirs.WorkDir, "/dev/null", art)
		if pushErr == nil {
			t.Fatal("expected Push to return error")
		}

		if !errors.Is(pushErr, InvalidChecksumError{}) {
			t.Fatalf("expected InvalidChecksumError, got %#v", pushErr)
		}
	})

	t.Run("push file artifact returns error if checksum not in cache", func(t *testing.T) {
		defer resetMocks()

		artStatus := artifact.Status{HasChecksum: true, ChecksumInCache: false}

		dirs, art, err := testutil.CreateArtifactTestCase(artStatus)
		defer os.RemoveAll(dirs.CacheDir)
		defer os.RemoveAll(dirs.WorkDir)
		if err != nil {
			t.Fatal(err)
		}

		ch, err := NewLocalCache(dirs.CacheDir)
		if err != nil {
			t.Fatal(err)
		}

		pushErr := ch.Push(dirs.WorkDir, "/dev/null", art)
		if pushErr == nil {
			t.Fatal("expected Push to return error")
		}

		if !errors.Is(pushErr, MissingFromCacheError{art.Checksum}) {
			t.Fatalf("expected MissingFromCacheError, got %#v", pushErr)
		}
	})

	t.Run("push dir artifact happy path", func(t *testing.T) {
		defer resetMocks()

		dirs, art, cache := setupDirTest(t)
		defer os.RemoveAll(dirs.CacheDir)
		defer os.RemoveAll(dirs.WorkDir)

		if err := cache.Commit(dirs.WorkDir, &art, strategy.LinkStrategy, logger); err != nil {
			t.Fatal(err)
		}

		fakeRemote := filepath.Join(dirs.WorkDir, "fake_remote")

		ch, err := NewLocalCache(dirs.CacheDir)
		if err != nil {
			t.Fatal(err)
		}

		remoteCopy = mockRemoteCopy

		if err := ch.Push(dirs.WorkDir, fakeRemote, art); err != nil {
			t.Fatal(err)
		}

		assertCacheDirsEqual(dirs.CacheDir, fakeRemote, t)
	})
}
