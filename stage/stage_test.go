package stage

import (
	"github.com/google/go-cmp/cmp"
	"github.com/kevlar1818/duc/artifact"
	"github.com/kevlar1818/duc/cache"
	"github.com/kevlar1818/duc/fsutil"
	"github.com/kevlar1818/duc/testutil"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestSetChecksum(t *testing.T) {
	s := Stage{
		Checksum:   "",
		WorkingDir: "foo",
		Outputs: []artifact.Artifact{
			{
				Checksum: "abc",
				Path:     "bar.txt",
			},
		},
	}

	s.SetChecksum()

	if s.Checksum == "" {
		t.Fatal("stage.SetChecksum() didn't change (empty) checksum")
	}

	expected := s

	s.Checksum = "this should not affect the checksum"

	s.SetChecksum()

	if diff := cmp.Diff(expected, s); diff != "" {
		t.Fatalf("stage.SetChecksum() -want +got:\n%s", diff)
	}

	origChecksum := s.Checksum
	s.WorkingDir = "this should affect the checksum"

	s.SetChecksum()

	if s.Checksum == origChecksum {
		t.Fatal("changing stage.WorkingDir should have affected checksum")
	}
}

func TestCommitIntegration(t *testing.T) {
	t.Run("Copy", func(t *testing.T) { testCommitIntegration(cache.CopyStrategy, t) })
	t.Run("Copy", func(t *testing.T) { testCommitIntegration(cache.LinkStrategy, t) })
}

func testCommitIntegration(strategy cache.CheckoutStrategy, t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	cacheDir, workDir, err := testutil.CreateTempDirs()
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(cacheDir)
	defer os.RemoveAll(workDir)

	fileWorkspacePath := path.Join(workDir, "foo.txt")
	if err = ioutil.WriteFile(fileWorkspacePath, []byte("Hello, World!"), 0644); err != nil {
		t.Error(err)
	}
	fileChecksum := "0a0a9f2a6772942557ab5355d76af442f8f65e01"
	fileCachePath := path.Join(cacheDir, fileChecksum[:2], fileChecksum[2:])

	s := Stage{
		WorkingDir: workDir,
		Outputs: []artifact.Artifact{
			{
				Checksum: "",
				Path:     "foo.txt",
			},
		},
	}

	expected := Stage{
		WorkingDir: workDir,
		Outputs: []artifact.Artifact{
			{
				Checksum: fileChecksum,
				Path:     "foo.txt",
			},
		},
	}

	if err := s.Commit(cacheDir, strategy); err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(expected, s); diff != "" {
		t.Errorf("Commit(stage) -want +got:\n%s", diff)
	}

	exists, err := fsutil.Exists(fileWorkspacePath)
	if err != nil {
		t.Error(err)
	}
	if !exists {
		t.Errorf("file %#v should exist", fileWorkspacePath)
	}
	cachedFileInfo, err := os.Stat(fileCachePath)
	if err != nil {
		t.Error(err)
	}
	if cachedFileInfo.Mode() != 0400 {
		t.Errorf("%#v has perms %#o, want %#o", fileCachePath, cachedFileInfo.Mode(), 0400)
	}
	switch strategy {
	case cache.CopyStrategy:
		// check that files are distinct, but have the same contents
		sameFile, err := fsutil.SameFile(fileWorkspacePath, fileCachePath)
		if err != nil {
			t.Error(err)
		}
		if sameFile {
			t.Errorf(
				"files %#v and %#v should not be the same",
				fileWorkspacePath,
				fileCachePath,
			)
		}
		sameContents, err := fsutil.SameContents(fileWorkspacePath, fileCachePath)
		if err != nil {
			t.Error(err)
		}
		if !sameContents {
			t.Errorf(
				"files %#v and %#v should have the same contents",
				fileWorkspacePath,
				fileCachePath,
			)
		}
	case cache.LinkStrategy:
		// check that workspace file is a link to cache file
		sameFile, err := fsutil.SameFile(fileWorkspacePath, fileCachePath)
		if err != nil {
			t.Fatal(err)
		}
		if !sameFile {
			t.Fatalf(
				"files %#v and %#v should be the same file",
				fileWorkspacePath,
				fileCachePath,
			)
		}
		linkDst, err := os.Readlink(fileWorkspacePath)
		if err != nil {
			t.Fatal(err)
		}
		if linkDst != fileCachePath {
			t.Fatalf(
				"file %#v links to %#v, want %#v",
				fileWorkspacePath,
				linkDst,
				fileCachePath,
			)
		}
	}
}
