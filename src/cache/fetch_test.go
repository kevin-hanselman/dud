package cache

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kevin-hanselman/dud/src/agglog"
	"github.com/kevin-hanselman/dud/src/artifact"
	"github.com/kevin-hanselman/dud/src/fsutil"
	"github.com/kevin-hanselman/dud/src/strategy"
	"github.com/kevin-hanselman/dud/src/testutil"
	"github.com/pkg/errors"
)

func getCacheFiles(cacheDir string) (map[string]struct{}, error) {
	files, err := filepath.Glob(filepath.Join(cacheDir, "*", "*"))
	if err != nil {
		return nil, err
	}
	fileSet := make(map[string]struct{})
	for _, file := range files {
		relFile, err := filepath.Rel(cacheDir, file)
		if err != nil {
			return nil, err
		}
		fileSet[relFile] = struct{}{}
	}
	return fileSet, nil
}

func assertCacheDirsEqual(dirA, dirB string, t *testing.T) {
	filesA, err := getCacheFiles(dirA)
	if err != nil {
		t.Fatal(err)
	}
	filesB, err := getCacheFiles(dirB)
	if err != nil {
		t.Fatal(err)
	}
	if len(filesA) == 0 {
		t.Fatalf("dir %s has no files", dirA)
	}
	if len(filesB) == 0 {
		t.Fatalf("dir %s has no files", dirB)
	}
	for file := range filesA {
		if _, ok := filesB[file]; !ok {
			t.Fatalf("file %#v in %s but not %s", file, dirA, dirB)
		}
	}
	for file := range filesB {
		if _, ok := filesA[file]; !ok {
			t.Fatalf("file %#v in %s but not %s", file, dirB, dirA)
		}
	}
}

func mkdirsThen(src, dst string, f func(src, dst string) error) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return f(src, dst)
}

// Mock remoteCopy with a version that creates hard links between directories
// TODO: Consider asserting the number of calls to remoteCopy.
func mockRemoteCopy(src, dst string, fileSet map[string]struct{}) error {
	for file := range fileSet {
		if err := mkdirsThen(
			filepath.Join(src, file),
			filepath.Join(dst, file),
			os.Link,
		); err != nil {
			return err
		}
	}
	return nil
}

func TestFetchIntegration(t *testing.T) {
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

	t.Run("fetch file artifact happy path", func(t *testing.T) {
		defer resetMocks()
		artStatus := artifact.Status{
			HasChecksum:         true,
			WorkspaceFileStatus: fsutil.StatusRegularFile,
		}

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

		// Add the Artifact to the fake remote to prep for the fetch.
		artCachePath, err := ch.PathForChecksum(art.Checksum)
		if err != nil {
			t.Fatal(err)
		}
		if err := mkdirsThen(
			filepath.Join(dirs.WorkDir, art.Path),
			filepath.Join(fakeRemote, artCachePath),
			os.Rename,
		); err != nil {
			t.Fatal(err)
		}

		remoteCopy = mockRemoteCopy

		if err := ch.Fetch(fakeRemote, map[string]*artifact.Artifact{"art": &art}); err != nil {
			t.Fatal(err)
		}

		assertCacheDirsEqual(dirs.CacheDir, fakeRemote, t)
	})

	t.Run("fetch file artifact noop if already in cache", func(t *testing.T) {
		defer resetMocks()
		artStatus := artifact.Status{HasChecksum: true, ChecksumInCache: true}

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

		if err := ch.Fetch("/dev/null", map[string]*artifact.Artifact{"art": &art}); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("fetch file artifact returns error if no checksum", func(t *testing.T) {
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

		fetchErr := ch.Fetch("/dev/null", map[string]*artifact.Artifact{"art": &art})
		if fetchErr == nil {
			t.Fatal("expected Fetch to return error")
		}

		if !errors.Is(fetchErr, InvalidChecksumError{}) {
			t.Fatalf("expected InvalidChecksumError, got %#v", fetchErr)
		}
	})

	t.Run("fetch dir artifact happy path", func(t *testing.T) {
		defer resetMocks()
		dirs, art, cache := setupDirTest(t)
		defer os.RemoveAll(dirs.CacheDir)
		defer os.RemoveAll(dirs.WorkDir)

		fakeRemote := filepath.Join(dirs.WorkDir, "fake_remote")

		if err := cache.Commit(dirs.WorkDir, &art, strategy.LinkStrategy, logger); err != nil {
			t.Fatal(err)
		}

		// Find all files that were committed and move them to the fake remote
		// cache. Moving the files instead of copying (the latter would emulate
		// a push) checks that fetch is actually calling remoteCopy.
		cachedFiles, err := getCacheFiles(dirs.CacheDir)
		if err != nil {
			t.Fatal(err)
		}
		for cacheFile := range cachedFiles {
			if err := mkdirsThen(
				filepath.Join(dirs.CacheDir, cacheFile),
				filepath.Join(fakeRemote, cacheFile),
				os.Rename,
			); err != nil {
				t.Fatal(err)
			}
		}

		remoteCopy = mockRemoteCopy

		if err := cache.Fetch(fakeRemote, map[string]*artifact.Artifact{"art": &art}); err != nil {
			t.Fatal(err)
		}

		assertCacheDirsEqual(dirs.CacheDir, fakeRemote, t)
	})

	t.Run("fetch dir artifact with identically named grandchildren", func(t *testing.T) {
		defer resetMocks()
		dirs, err := testutil.CreateTempDirs()
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(dirs.CacheDir)
		defer os.RemoveAll(dirs.WorkDir)

		art := artifact.Artifact{Path: "bish", IsDir: true}

		// Create two sub-directories in the Artifact, each with its own
		// file with the same basename.
		subDirA := filepath.Join(dirs.WorkDir, art.Path, "bash")
		if err := os.MkdirAll(subDirA, 0o755); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(filepath.Join(subDirA, "same_name.txt"), []byte("bash"), 0o644); err != nil {
			t.Fatal(err)
		}

		subDirB := filepath.Join(dirs.WorkDir, art.Path, "bosh")
		if err := os.MkdirAll(subDirB, 0o755); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(filepath.Join(subDirB, "same_name.txt"), []byte("bosh"), 0o644); err != nil {
			t.Fatal(err)
		}

		fakeRemote := filepath.Join(dirs.WorkDir, "fake_remote")

		cache, err := NewLocalCache(dirs.CacheDir)
		if err != nil {
			t.Fatal(err)
		}

		if err := cache.Commit(dirs.WorkDir, &art, strategy.LinkStrategy, logger); err != nil {
			t.Fatal(err)
		}

		// Find all files that were committed and move them to the fake remote
		// cache. Moving the files instead of copying (the latter would emulate
		// a push) checks that fetch is actually calling remoteCopy.
		cachedFiles, err := getCacheFiles(dirs.CacheDir)
		if err != nil {
			t.Fatal(err)
		}
		for cacheFile := range cachedFiles {
			if err := mkdirsThen(
				filepath.Join(dirs.CacheDir, cacheFile),
				filepath.Join(fakeRemote, cacheFile),
				os.Rename,
			); err != nil {
				t.Fatal(err)
			}
		}

		remoteCopy = mockRemoteCopy

		if err := cache.Fetch(fakeRemote, map[string]*artifact.Artifact{"art": &art}); err != nil {
			t.Fatal(err)
		}

		assertCacheDirsEqual(dirs.CacheDir, fakeRemote, t)
	})
}
