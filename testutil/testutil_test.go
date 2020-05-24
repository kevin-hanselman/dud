package testutil

import (
	"fmt"
	"github.com/kevlar1818/duc/artifact"
	"github.com/kevlar1818/duc/cache"
	"github.com/kevlar1818/duc/fsutil"
	"os"
	"path/filepath"
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
	for _, status := range AllTestCases() {
		t.Run(fmt.Sprintf("%+v", status), func(t *testing.T) {
			testCreateArtifactTestCaseIntegration(status, t)
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

	// verify art.Checksum matches status.HasChecksum
	artHasChecksum := art.Checksum != ""
	if artHasChecksum != status.HasChecksum {
		t.Errorf("artifact checksum %v", art.Checksum)
	}

	// verify art.IsDir matches status.WorkspaceFileStatus
	if art.IsDir != (status.WorkspaceFileStatus == fsutil.Directory) {
		t.Fatalf(
			"artifact.IsDir = %v, but status.WorkspaceFileStatus = %v",
			art.IsDir,
			status.WorkspaceFileStatus,
		)
	}

	// verify workPath existences matches status.WorkspaceFileStatus
	workPath := filepath.Join(dirs.WorkDir, art.Path)
	workPathExists, err := fsutil.Exists(workPath, false)
	if err != nil {
		t.Fatal(err)
	}
	shouldExist := status.WorkspaceFileStatus != fsutil.Absent
	if workPathExists != shouldExist {
		t.Fatalf("Exists(%#v) = %#v", workPath, workPathExists)
	}

	// verify cachePath matches status.HasChecksum
	ch, err := cache.NewLocalCache(dirs.CacheDir)
	if err != nil {
		t.Fatal(err)
	}
	cachePath, err := ch.PathForChecksum(art.Checksum)
	foundValidChecksum := err == nil
	if status.HasChecksum != foundValidChecksum {
		t.Fatalf(
			"art.Checksum = %#v, but status.HasChecksum = %#v",
			art.Checksum,
			status.HasChecksum,
		)
	}

	// verify cachePath matches status.ChecksumInCache
	cachePathExists, err := fsutil.Exists(cachePath, false)
	if err != nil {
		t.Fatal(err)
	}
	if cachePathExists != status.ChecksumInCache {
		t.Fatalf("Exists(%#v) = %#v", cachePath, cachePathExists)
	}

	switch status.WorkspaceFileStatus {
	// verify workPath is a link and matches status.ContentsMatch
	case fsutil.Link:
		linkDst, err := os.Readlink(workPath)
		if err != nil {
			t.Fatal(err)
		}
		correctLink := linkDst == cachePath
		if correctLink != status.ContentsMatch {
			t.Fatalf("%#v links to %#v", workPath, linkDst)
		}
	// verify workPath is a regular file and matches status.ContentsMatch
	case fsutil.RegularFile:
		isRegFile, err := fsutil.IsRegularFile(workPath)
		if err != nil {
			t.Fatal(err)
		}
		if !isRegFile {
			t.Fatalf("expected %#v to be a regular file", workPath)
		}
		if status.ChecksumInCache {
			same, err := fsutil.SameContents(workPath, cachePath)
			if err != nil {
				t.Fatal(err)
			}
			if same != status.ContentsMatch {
				t.Fatalf("SameContents(%v, %v) = %v", workPath, cachePath, same)
			}
		}
	}
}
