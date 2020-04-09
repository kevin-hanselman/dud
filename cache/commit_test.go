package cache

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/kevlar1818/duc/artifact"
	"github.com/kevlar1818/duc/strategy"
	"github.com/kevlar1818/duc/testutil"
	"os"
	"testing"
)

func TestCommitDirectory(t *testing.T) {
	// TODO
	// mock function to list files in a directory, probably using package var instead of testify
	// one of:
	//  filepath.Walk
	//  ioutil.ReadDir
	//  os.File.Readdir

	// mock function to save directory struct to disk

	// call cache.Commit on a directory artifact
	// assert that mocked cache.Commit is called on each file in dir
	// assert that mocked "save" function is called as expected

	// assert result is correct
	//   dir artifact has correct checksum
	//   checksum should be invariant to file order
	//   an error when committing a file should cause an error for the dir
	//   don't commit any files if one them is invalid prior to starting commit?
	//   etc.
}

func TestCommitIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	for _, testCase := range testutil.AllTestCases() {
		for _, strat := range []strategy.CheckoutStrategy{strategy.CopyStrategy, strategy.LinkStrategy} {
			t.Run(fmt.Sprintf("%s %+v", strat, testCase), func(t *testing.T) {
				testCommitIntegration(strat, testCase, t)
			})
		}
	}
}

func testCommitIntegration(strat strategy.CheckoutStrategy, statusStart artifact.Status, t *testing.T) {
	dirs, art, err := testutil.CreateArtifactTestCase(statusStart)
	defer os.RemoveAll(dirs.CacheDir)
	defer os.RemoveAll(dirs.WorkDir)
	if err != nil {
		t.Fatal(err)
	}
	cache := LocalCache{Dir: dirs.CacheDir}

	commitErr := cache.Commit(dirs.WorkDir, &art, strat)

	if statusStart.WorkspaceStatus == artifact.Absent {
		if os.IsNotExist(commitErr) {
			return // TODO: assert expected status
		}
		t.Fatalf("expected Commit to raise NotExist error, got %#v", commitErr)
	} else if statusStart.WorkspaceStatus == artifact.Link {
		if commitErr != nil {
			return // TODO: assert expected status
		}
		t.Fatal("expected Commit to raise error")
	} else if commitErr != nil {
		t.Fatalf("unexpected error: %v", commitErr)
	}

	testCachePermissions(cache, art, t)

	statusGot, err := cache.Status(dirs.WorkDir, art)
	if err != nil {
		t.Fatal(err)
	}
	statusWant := artifact.Status{
		HasChecksum:     true,
		ChecksumInCache: true,
		ContentsMatch:   true,
	}
	switch strat {
	case strategy.CopyStrategy:
		statusWant.WorkspaceStatus = artifact.RegularFile
	case strategy.LinkStrategy:
		statusWant.WorkspaceStatus = artifact.Link
	}

	if diff := cmp.Diff(statusWant, statusGot); diff != "" {
		t.Fatalf("Status() -want +got:\n%s", diff)
	}
}

func testCachePermissions(cache LocalCache, art artifact.Artifact, t *testing.T) {
	fileCachePath, err := cache.CachePathForArtifact(art)
	if err != nil {
		t.Fatal(err)
	}
	cachedFileInfo, err := os.Stat(fileCachePath)
	if err != nil {
		t.Fatal(err)
	}
	// TODO: check this in cache.Status?
	if cachedFileInfo.Mode() != 0444 {
		t.Fatalf("%#v has perms %#o, want %#o", fileCachePath, cachedFileInfo.Mode(), 0444)
	}
}
