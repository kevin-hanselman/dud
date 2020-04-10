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
	commitFileArtifactCalls := []commitArgs{}
	commitFileArtifactOrig := commitFileArtifact
	commitFileArtifact = func(args commitArgs) error {
		commitFileArtifactCalls = append(commitFileArtifactCalls, args)
		return nil
	}
	defer func() { commitFileArtifact = commitFileArtifactOrig }()

	mockFiles := []os.FileInfo{
		testutil.MockFileInfo{MockName: "my_file1"},
		testutil.MockFileInfo{MockName: "my_dir", MockMode: os.ModeDir},
		testutil.MockFileInfo{MockName: "my_link", MockMode: os.ModeSymlink},
		testutil.MockFileInfo{MockName: "my_file2"},
	}

	readDirOrig := readDir
	readDir = func(dir string) ([]os.FileInfo, error) {
		return mockFiles, nil
	}
	defer func() { readDir = readDirOrig }()

	cache := LocalCache{Dir: "cache_root"}

	dirArt := artifact.Artifact{IsDir: true, Checksum: "", Path: "art_dir"}

	commitErr := cache.Commit("work_dir", &dirArt, strategy.LinkStrategy)
	if commitErr != nil {
		t.Fatal(commitErr)
	}

	expectedCommitFileArtifactCalls := []commitArgs{
		{Cache: &cache, WorkingDir: "work_dir/art_dir", Artifact: &artifact.Artifact{Path: "my_file1"}},
		{Cache: &cache, WorkingDir: "work_dir/art_dir", Artifact: &artifact.Artifact{Path: "my_link"}},
		{Cache: &cache, WorkingDir: "work_dir/art_dir", Artifact: &artifact.Artifact{Path: "my_file2"}},
	}

	if diff := cmp.Diff(commitFileArtifactCalls, expectedCommitFileArtifactCalls); diff != "" {
		t.Fatalf("commitFileArtifactCalls -want +got:\n%s", diff)
	}

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
