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
	defer os.RemoveAll(dirs.WorkDir)
	defer os.RemoveAll(dirs.CacheDir)
	if err != nil {
		t.Fatal(err)
	}

	checksumEmpty := art.Checksum == ""
	if checksumEmpty == status.HasChecksum {
		t.Errorf("artifact checksum %v", art.Checksum)
	}

	workPath := path.Join(dirs.WorkDir, art.Path)
	ch := cache.LocalCache{Dir: dirs.CacheDir}
	cachePath, err := ch.CachePathForArtifact(art)

	exists, err := fsutil.Exists(workPath, false)
	if err != nil {
		t.Fatal(err)
	}
	shouldExist := status.WorkspaceStatus != artifact.Absent
	if exists != shouldExist {
		t.Fatalf("Exists(%#v) = %#v", workPath, exists)
	}

	exists, err = fsutil.Exists(cachePath, false)
	if err != nil {
		t.Fatal(err)
	}
	if exists != status.ChecksumInCache {
		t.Fatalf("Exists(%#v) = %#v", cachePath, exists)
	}

	switch status.WorkspaceStatus {
	case artifact.Link:
		linkDst, err := os.Readlink(workPath)
		if err != nil {
			t.Fatal(err)
		}
		correctLink := linkDst == cachePath
		if correctLink != status.ContentsMatch {
			t.Errorf("%v links to %v", workPath, linkDst)
		}
	case artifact.RegularFile:
		if status.ChecksumInCache {
			same, err := fsutil.SameContents(workPath, cachePath)
			if err != nil {
				t.Fatal(err)
			}
			if same != status.ContentsMatch {
				t.Errorf("SameContents(%v, %v) = %v", workPath, cachePath, same)
			}
		}
	}
}
