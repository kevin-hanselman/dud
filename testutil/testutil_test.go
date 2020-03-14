package testutil

import (
	"fmt"
	"github.com/kevlar1818/duc/artifact"
	"github.com/kevlar1818/duc/cache"
	"github.com/kevlar1818/duc/fsutil"
	"os"
	"path"
	"testing"
)

func TestCreateTempDirsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dirs, err := CreateTempDirs()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dirs.WorkDir)
	defer os.RemoveAll(dirs.CacheDir)
	if exists, _ := fsutil.Exists(dirs.WorkDir, false); !exists {
		t.Errorf("directory %#v doesn't exist", dirs.WorkDir)
	}
	if exists, _ := fsutil.Exists(dirs.CacheDir, false); !exists {
		t.Errorf("directory %#v doesn't exist", dirs.CacheDir)
	}
}

func TestCreateArtifactTestCaseIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	for _, args := range AllTestCases() {
		t.Run(fmt.Sprintf("%#v", args), func(t *testing.T) {
			testCreateArtifactTestCaseIntegration(args, t)
		})
	}
}

func testCreateArtifactTestCaseIntegration(status artifact.Status, t *testing.T) {
	dirs, art, err := CreateArtifactTestCase(status)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dirs.WorkDir)
	defer os.RemoveAll(dirs.CacheDir)

	workPath := path.Join(dirs.WorkDir, art.Path)
	ch := cache.LocalCache{Dir: dirs.CacheDir}
	cachePath, err := ch.CachePathForArtifact(art)

	exists, err := fsutil.Exists(workPath, false)
	if err != nil {
		t.Fatal(err)
	}
	shouldExist := status.FileStatus != artifact.IsAbsent
	if exists != shouldExist {
		t.Fatalf("Exists(%#v) = %#v", workPath, exists)
	}

	exists, err = fsutil.Exists(cachePath, false)
	if err != nil {
		t.Fatal(err)
	}
	if exists != status.InCache {
		t.Fatalf("Exists(%#v) = %#v", cachePath, exists)
	}

	switch status.FileStatus {
	case artifact.IsLink:
		linkDst, err := os.Readlink(workPath)
		if err != nil {
			t.Fatal(err)
		}
		if linkDst != cachePath {
			t.Errorf("%#v links to %#v, want %#v", workPath, linkDst, cachePath)
		}
	case artifact.IsRegularFile:
		if status.InCache {
			same, err := fsutil.SameContents(workPath, cachePath)
			if err != nil {
				t.Fatal(err)
			}
			if !same {
				t.Errorf("SameContents(%#v, %#v) = false", workPath, cachePath)
			}
		}
	}
}
