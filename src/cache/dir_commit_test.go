package cache

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kevin-hanselman/dud/src/agglog"
	"github.com/kevin-hanselman/dud/src/artifact"
	"github.com/kevin-hanselman/dud/src/fsutil"
	"github.com/kevin-hanselman/dud/src/strategy"
	"github.com/kevin-hanselman/dud/src/testutil"
)

func TestDirectoryCommitIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	makeExpectedStatus := func(art artifact.Artifact) artifact.Status {
		upToDate := func(art artifact.Artifact) *artifact.Status {
			return &artifact.Status{
				WorkspaceFileStatus: fsutil.StatusLink,
				Artifact:            art,
				HasChecksum:         true,
				ChecksumInCache:     true,
				ContentsMatch:       true,
			}
		}

		art.Checksum = ""

		return artifact.Status{
			Artifact:            art,
			WorkspaceFileStatus: fsutil.StatusDirectory,
			HasChecksum:         true,
			ChecksumInCache:     true,
			ContentsMatch:       true,
			ChildrenStatus: map[string]*artifact.Status{
				"1.txt": upToDate(artifact.Artifact{Path: "1.txt"}),
				"2.txt": upToDate(artifact.Artifact{Path: "2.txt"}),
				"3.txt": upToDate(artifact.Artifact{Path: "3.txt"}),
				"4.txt": upToDate(artifact.Artifact{Path: "4.txt"}),
				"5.txt": upToDate(artifact.Artifact{Path: "5.txt"}),
				"bar": {
					Artifact:            artifact.Artifact{Path: "bar", IsDir: true},
					WorkspaceFileStatus: fsutil.StatusDirectory,
					HasChecksum:         true,
					ChecksumInCache:     true,
					ContentsMatch:       true,
					ChildrenStatus: map[string]*artifact.Status{
						"4.txt": upToDate(artifact.Artifact{Path: "4.txt"}),
						"5.txt": upToDate(artifact.Artifact{Path: "5.txt"}),
						"6.txt": upToDate(artifact.Artifact{Path: "6.txt"}),
						"7.txt": upToDate(artifact.Artifact{Path: "7.txt"}),
						"8.txt": upToDate(artifact.Artifact{Path: "8.txt"}),
					},
				},
			},
		}
	}

	logger := agglog.NewNullLogger()

	maxSharedWorkers = 1
	maxDedicatedWorkers = 1

	t.Run("happy path", func(t *testing.T) {
		dirs, art, cache := setupDirTest(t)
		defer os.RemoveAll(dirs.CacheDir)
		defer os.RemoveAll(dirs.WorkDir)

		if err := cache.Commit(dirs.WorkDir, &art, strategy.LinkStrategy, logger); err != nil {
			t.Fatal(err)
		}

		actualStatus, err := cache.Status(dirs.WorkDir, art, false)
		if err != nil {
			t.Fatal(err)
		}

		expectedStatus := makeExpectedStatus(art)

		assertThenRemoveChecksums(t, &actualStatus)

		if diff := cmp.Diff(expectedStatus, actualStatus); diff != "" {
			t.Fatalf("Status -want +got:\n%s", diff)
		}
	})

	t.Run("partially up-to-date, rm subdir", func(t *testing.T) {
		dirs, art, cache := setupDirTest(t)
		defer os.RemoveAll(dirs.CacheDir)
		defer os.RemoveAll(dirs.WorkDir)

		if err := cache.Commit(dirs.WorkDir, &art, strategy.LinkStrategy, logger); err != nil {
			t.Fatal(err)
		}

		if err := os.RemoveAll(filepath.Join(dirs.WorkDir, "foo", "bar")); err != nil {
			t.Fatal(err)
		}

		if err := cache.Commit(dirs.WorkDir, &art, strategy.LinkStrategy, logger); err != nil {
			t.Fatal(err)
		}

		actualStatus, err := cache.Status(dirs.WorkDir, art, false)
		if err != nil {
			t.Fatal(err)
		}

		expectedStatus := makeExpectedStatus(art)
		delete(expectedStatus.ChildrenStatus, "bar")

		assertThenRemoveChecksums(t, &actualStatus)

		if diff := cmp.Diff(expectedStatus, actualStatus); diff != "" {
			t.Fatalf("Status -want +got:\n%s", diff)
		}
	})

	t.Run("partially up-to-date, rm file", func(t *testing.T) {
		dirs, art, cache := setupDirTest(t)
		defer os.RemoveAll(dirs.CacheDir)
		defer os.RemoveAll(dirs.WorkDir)

		if err := cache.Commit(dirs.WorkDir, &art, strategy.LinkStrategy, logger); err != nil {
			t.Fatal(err)
		}

		if err := os.Remove(filepath.Join(dirs.WorkDir, "foo", "1.txt")); err != nil {
			t.Fatal(err)
		}

		if err := cache.Commit(dirs.WorkDir, &art, strategy.LinkStrategy, logger); err != nil {
			t.Fatal(err)
		}

		actualStatus, err := cache.Status(dirs.WorkDir, art, false)
		if err != nil {
			t.Fatal(err)
		}

		expectedStatus := makeExpectedStatus(art)
		delete(expectedStatus.ChildrenStatus, "1.txt")

		assertThenRemoveChecksums(t, &actualStatus)

		if diff := cmp.Diff(expectedStatus, actualStatus); diff != "" {
			t.Fatalf("Status -want +got:\n%s", diff)
		}
	})
}

func assertThenRemoveChecksums(t *testing.T, statusGot *artifact.Status) {
	if statusGot.Checksum == "" {
		t.Fatalf("expected checksum for artifact %s", statusGot.Path)
	}
	statusGot.Checksum = ""
	if statusGot.ChildrenStatus == nil {
		return
	}
	for _, childStatus := range statusGot.ChildrenStatus {
		assertThenRemoveChecksums(t, childStatus)
	}
}

func setupDirTest(t *testing.T) (testutil.TempDirs, artifact.Artifact, LocalCache) {
	dirs, err := testutil.CreateTempDirs()
	if err != nil {
		t.Fatal(err)
	}

	cache, err := NewLocalCache(dirs.CacheDir)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(filepath.Join(dirs.WorkDir, "foo", "bar"), 0o755); err != nil {
		t.Fatal(err)
	}

	for i := 1; i <= 5; i++ {
		s := fmt.Sprint(i)
		path := filepath.Join(dirs.WorkDir, "foo", fmt.Sprintf("%s.txt", s))
		if err := ioutil.WriteFile(path, []byte(s), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// Intentionally overlap file names/contents with parent dir.
	for i := 4; i <= 8; i++ {
		s := fmt.Sprint(i)
		path := filepath.Join(dirs.WorkDir, "foo", "bar", fmt.Sprintf("%s.txt", s))
		if err := ioutil.WriteFile(path, []byte(s), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	art := artifact.Artifact{Path: "foo", IsDir: true}

	return dirs, art, cache
}
