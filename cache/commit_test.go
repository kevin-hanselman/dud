package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kevlar1818/duc/artifact"
	"github.com/kevlar1818/duc/fsutil"
	"github.com/kevlar1818/duc/strategy"
	"github.com/kevlar1818/duc/testutil"
	"github.com/pkg/errors"
)

func TestCommitDirectory(t *testing.T) {
	t.Run("non-recursive", testCommitDirectoryNonRecursive)
	t.Run("recursive", testCommitDirectoryRecursive)
}

func testCommitDirectoryNonRecursive(t *testing.T) {
	commitFileArtifactsThreadSafe := sync.Map{}
	commitFileArtifactOrig := commitFileArtifact
	commitFileArtifact = func(
		ch *LocalCache,
		workingDir string,
		art *artifact.Artifact,
		strat strategy.CheckoutStrategy,
	) error {
		art.Checksum = "123456789"
		count, ok := commitFileArtifactsThreadSafe.Load(art.Path)
		countInt := 0
		if ok {
			countInt = count.(int)
		}
		countInt++
		commitFileArtifactsThreadSafe.Store(art.Path, countInt)
		return nil
	}
	defer func() { commitFileArtifact = commitFileArtifactOrig }()

	mockFiles := []os.FileInfo{
		testutil.MockFileInfo{MockName: "my_file1"},
		testutil.MockFileInfo{MockName: "child_dir", MockMode: os.ModeDir},
		// TODO: cover handling of symlinks (and other irregular files?)
		testutil.MockFileInfo{MockName: "my_link", MockMode: os.ModeSymlink},
		testutil.MockFileInfo{MockName: "my_file2"},
	}

	readDirOrig := readDir
	readDir = func(dir string) ([]os.FileInfo, error) {
		return mockFiles, nil
	}
	defer func() { readDir = readDirOrig }()

	var actualManifest directoryManifest
	expectedChecksum := "deadbeef"
	commitDirManifestOrig := commitDirManifest
	commitDirManifest = func(ch *LocalCache, man *directoryManifest) (string, error) {
		sort.Sort(byPath(man.Contents))
		actualManifest = *man
		return expectedChecksum, nil
	}
	defer func() { commitDirManifest = commitDirManifestOrig }()

	cache, err := NewLocalCache("/cache_root")
	if err != nil {
		t.Fatal(err)
	}
	dirArt := artifact.Artifact{IsDir: true, Checksum: "", Path: "art_dir"}

	commitErr := cache.Commit("work_dir", &dirArt, strategy.LinkStrategy)
	if commitErr != nil {
		t.Fatal(commitErr)
	}

	expectedFileArtifacts := []*artifact.Artifact{
		{Checksum: "123456789", Path: "my_file1"},
		{Checksum: "123456789", Path: "my_link"},
		{Checksum: "123456789", Path: "my_file2"},
	}

	expectedCommitFileArtifactCalls := make(map[string]int)

	baseDir := filepath.Join("work_dir", "art_dir")

	for i := range expectedFileArtifacts {
		path := expectedFileArtifacts[i].Path
		expectedCommitFileArtifactCalls[path] = 1
	}

	commitFileArtifactCalls := make(map[string]int)
	commitFileArtifactsThreadSafe.Range(func(key, val interface{}) bool {
		commitFileArtifactCalls[key.(string)] = val.(int)
		return true
	})

	if diff := cmp.Diff(expectedCommitFileArtifactCalls, commitFileArtifactCalls); diff != "" {
		t.Fatalf("commitFileArtifactCalls -want +got:\n%s", diff)
	}

	expectedManifest := directoryManifest{
		Path:     baseDir,
		Contents: expectedFileArtifacts,
	}
	sort.Sort(byPath(expectedManifest.Contents))

	if dirArt.Checksum != expectedChecksum {
		t.Fatalf("artifact checksum want %#v, got %#v", expectedChecksum, dirArt.Checksum)
	}

	// assert result is correct
	//   dir artifact has correct checksum
	//   checksum should be invariant to file order
	//   an error when committing a file should cause an error for the dir
	//   don't commit any files if one them is invalid prior to starting commit?
	//   etc.

	if diff := cmp.Diff(expectedManifest, actualManifest); diff != "" {
		t.Fatalf("directoryManifest -want +got:\n%s", diff)
	}
}

func testCommitDirectoryRecursive(t *testing.T) {
	commitFileArtifactsThreadSafe := sync.Map{} // switch to a simple mutex?
	commitFileArtifactOrig := commitFileArtifact
	commitFileArtifact = func(
		ch *LocalCache,
		workingDir string,
		art *artifact.Artifact,
		strat strategy.CheckoutStrategy,
	) error {
		art.Checksum = "123456789"
		count, ok := commitFileArtifactsThreadSafe.Load(art.Path)
		countInt := 0
		if ok {
			countInt = count.(int)
		}
		countInt++
		commitFileArtifactsThreadSafe.Store(art.Path, countInt)
		return nil
	}
	defer func() { commitFileArtifact = commitFileArtifactOrig }()

	mockFilesParent := []os.FileInfo{
		testutil.MockFileInfo{MockName: "my_file1"},
		testutil.MockFileInfo{MockName: "my_link", MockMode: os.ModeSymlink},
		testutil.MockFileInfo{MockName: "my_file2"},
		// NOTE: The order here will affect the expected order of commitFileArtifactcalls
		testutil.MockFileInfo{MockName: "child_dir", MockMode: os.ModeDir},
	}

	mockFilesChild := []os.FileInfo{
		testutil.MockFileInfo{MockName: "nested_file1"},
		testutil.MockFileInfo{MockName: "nested_file2"},
	}

	readDirCallCount := 0
	readDirOrig := readDir
	readDir = func(dir string) ([]os.FileInfo, error) {
		readDirCallCount++
		if readDirCallCount == 1 {
			return mockFilesParent, nil
		} else if readDirCallCount == 2 {
			return mockFilesChild, nil
		} else {
			t.Fatal("unexpected call to readDir")
		}
		return []os.FileInfo{}, nil
	}
	defer func() { readDir = readDirOrig }()

	actualManifests := []directoryManifest{}
	expectedChecksum := "deadbeef"
	commitDirManifestOrig := commitDirManifest
	commitDirManifest = func(ch *LocalCache, man *directoryManifest) (string, error) {
		sort.Sort(byPath(man.Contents))
		actualManifests = append(actualManifests, *man)
		return expectedChecksum, nil
	}
	defer func() { commitDirManifest = commitDirManifestOrig }()

	cache, err := NewLocalCache("/cache_root")
	if err != nil {
		t.Fatal(err)
	}
	dirArt := artifact.Artifact{IsDir: true, IsRecursive: true, Checksum: "", Path: "art_dir"}

	commitErr := cache.Commit("work_dir", &dirArt, strategy.LinkStrategy)
	if commitErr != nil {
		t.Fatal(commitErr)
	}

	expectedFileArtifactsParent := []*artifact.Artifact{
		{Checksum: "123456789", Path: "my_file1"},
		{Checksum: "123456789", Path: "my_link"},
		{Checksum: "123456789", Path: "my_file2"},
	}
	expectedFileArtifactsChild := []*artifact.Artifact{
		{Checksum: "123456789", Path: "nested_file1"},
		{Checksum: "123456789", Path: "nested_file2"},
	}

	expectedCommitFileArtifactCalls := make(map[string]int)

	baseDir := filepath.Join("work_dir", "art_dir")
	childDir := filepath.Join(baseDir, "child_dir")

	for i := range expectedFileArtifactsParent {
		path := expectedFileArtifactsParent[i].Path
		expectedCommitFileArtifactCalls[path] = 1
	}
	for i := range expectedFileArtifactsChild {
		path := expectedFileArtifactsChild[i].Path
		expectedCommitFileArtifactCalls[path] = 1
	}

	commitFileArtifactCalls := make(map[string]int)
	commitFileArtifactsThreadSafe.Range(func(key, _ interface{}) bool {
		commitFileArtifactCalls[key.(string)] = 1
		return true
	})

	if diff := cmp.Diff(expectedCommitFileArtifactCalls, commitFileArtifactCalls); diff != "" {
		t.Fatalf("commitFileArtifactCalls -want +got:\n%s", diff)
	}

	expectedManifestParent := directoryManifest{
		Path: baseDir,
		Contents: append(
			expectedFileArtifactsParent,
			&artifact.Artifact{
				Checksum:    expectedChecksum,
				Path:        "child_dir",
				IsDir:       true,
				IsRecursive: true,
			},
		),
	}
	sort.Sort(byPath(expectedManifestParent.Contents))
	expectedManifestChild := directoryManifest{
		Path:     childDir,
		Contents: expectedFileArtifactsChild,
	}
	sort.Sort(byPath(expectedManifestChild.Contents))

	if dirArt.Checksum != expectedChecksum {
		t.Fatalf("artifact checksum want %#v, got %#v", expectedChecksum, dirArt.Checksum)
	}

	if diff := cmp.Diff(expectedManifestChild, actualManifests[0]); diff != "" {
		t.Fatalf("child directoryManifest -want +got:\n%s", diff)
	}

	if diff := cmp.Diff(expectedManifestParent, actualManifests[1]); diff != "" {
		t.Fatalf("parent directoryManifest -want +got:\n%s", diff)
	}
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

func testCommitIntegration(strat strategy.CheckoutStrategy, statusStart artifact.ArtifactWithStatus, t *testing.T) {
	dirs, art, err := testutil.CreateArtifactTestCase(statusStart)
	defer os.RemoveAll(dirs.CacheDir)
	defer os.RemoveAll(dirs.WorkDir)
	if err != nil {
		t.Fatal(err)
	}
	cache, err := NewLocalCache(dirs.CacheDir)
	if err != nil {
		t.Fatal(err)
	}

	commitErr := cache.Commit(dirs.WorkDir, &art, strat)
	causeErr := errors.Cause(commitErr)
	statusWant := statusStart.Status // By default we expect nothing to change.

	// TODO: TDD for #14 (skip/don't error on up-to-date links)
	if statusStart.WorkspaceFileStatus == fsutil.Absent {
		if !os.IsNotExist(causeErr) {
			t.Fatalf("expected Commit to return a NotExist error, got %#v", causeErr)
		}
	} else if statusStart.WorkspaceFileStatus != fsutil.RegularFile {
		if causeErr == nil {
			t.Fatal("expected Commit to return an error")
		}
	} else if commitErr != nil {
		t.Fatalf("unexpected error: %v", commitErr)
	} else {
		statusWant.HasChecksum = true
		statusWant.ContentsMatch = true
		// Artifact.SkipCache will not modify the cache or the workspace file, so we retain
		// the starting values for ChecksumInCache and WorkspaceFileStatus.
		if !art.SkipCache {
			statusWant.ChecksumInCache = true
			switch strat {
			case strategy.CopyStrategy:
				statusWant.WorkspaceFileStatus = fsutil.RegularFile
			case strategy.LinkStrategy:
				statusWant.WorkspaceFileStatus = fsutil.Link
			}
		}
	}

	statusGot, err := cache.Status(dirs.WorkDir, art)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(statusWant, statusGot); diff != "" {
		t.Fatalf("Status() -want +got:\n%s", diff)
	}

	if statusWant.ChecksumInCache {
		testCachePermissions(cache, art, t)
	}
}

type byPath []*artifact.Artifact

func (a byPath) Len() int           { return len(a) }
func (a byPath) Less(i, j int) bool { return strings.Compare(a[i].Path, a[j].Path) < 0 }
func (a byPath) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func testCachePermissions(cache *LocalCache, art artifact.Artifact, t *testing.T) {
	fileCachePath, err := cache.PathForChecksum(art.Checksum)
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
