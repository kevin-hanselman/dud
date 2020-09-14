package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kevin-hanselman/duc/artifact"
	"github.com/kevin-hanselman/duc/fsutil"
	"github.com/kevin-hanselman/duc/testutil"
)

func TestDirectoryStatus(t *testing.T) {

	t.Run("happy path: all files up to date", func(t *testing.T) {
		dirQuickStatus := artifact.Status{
			WorkspaceFileStatus: fsutil.Directory,
			HasChecksum:         true,
			ChecksumInCache:     true,
		}

		manifestFileStatuses := map[string]artifact.Status{
			"my_file1": {WorkspaceFileStatus: fsutil.RegularFile, HasChecksum: true, ChecksumInCache: true, ContentsMatch: true},
			"my_link":  {WorkspaceFileStatus: fsutil.Link, HasChecksum: true, ChecksumInCache: true, ContentsMatch: true},
			"my_file2": {WorkspaceFileStatus: fsutil.RegularFile, HasChecksum: true, ChecksumInCache: true, ContentsMatch: true},
		}

		workspaceFiles := []os.FileInfo{
			testutil.MockFileInfo{MockName: "my_file1"},
			testutil.MockFileInfo{MockName: "my_dir", MockMode: os.ModeDir},
			// TODO: cover handling of symlinks (and other irregular files?)
			testutil.MockFileInfo{MockName: "my_link", MockMode: os.ModeSymlink},
			testutil.MockFileInfo{MockName: "my_file2"},
		}

		expectedDirStatus := artifact.Status{
			WorkspaceFileStatus: fsutil.Directory,
			HasChecksum:         true,
			ChecksumInCache:     true,
			ContentsMatch:       true,
		}

		testDirectoryStatus(dirQuickStatus, manifestFileStatuses, workspaceFiles, expectedDirStatus, t)
	})

	t.Run("file missing from workspace", func(t *testing.T) {
		dirQuickStatus := artifact.Status{
			WorkspaceFileStatus: fsutil.Directory,
			HasChecksum:         true,
			ChecksumInCache:     true,
		}

		manifestFileStatuses := map[string]artifact.Status{
			"my_file1": {WorkspaceFileStatus: fsutil.Absent},
			"my_link":  {WorkspaceFileStatus: fsutil.Link, HasChecksum: true, ChecksumInCache: true, ContentsMatch: true},
			"my_file2": {WorkspaceFileStatus: fsutil.RegularFile, HasChecksum: true, ChecksumInCache: true, ContentsMatch: true},
		}

		workspaceFiles := []os.FileInfo{
			testutil.MockFileInfo{MockName: "my_dir", MockMode: os.ModeDir},
			testutil.MockFileInfo{MockName: "my_link", MockMode: os.ModeSymlink},
			testutil.MockFileInfo{MockName: "my_file2"},
		}

		expectedDirStatus := artifact.Status{
			WorkspaceFileStatus: fsutil.Directory,
			HasChecksum:         true,
			ChecksumInCache:     true,
			ContentsMatch:       false,
		}

		testDirectoryStatus(dirQuickStatus, manifestFileStatuses, workspaceFiles, expectedDirStatus, t)
	})

	t.Run("untracked file in dir", func(t *testing.T) {
		dirQuickStatus := artifact.Status{
			WorkspaceFileStatus: fsutil.Directory,
			HasChecksum:         true,
			ChecksumInCache:     true,
		}

		manifestFileStatuses := map[string]artifact.Status{
			"my_file1": {WorkspaceFileStatus: fsutil.RegularFile, HasChecksum: true, ChecksumInCache: true, ContentsMatch: true},
			"my_link":  {WorkspaceFileStatus: fsutil.Link, HasChecksum: true, ChecksumInCache: true, ContentsMatch: true},
			"my_file2": {WorkspaceFileStatus: fsutil.RegularFile, HasChecksum: true, ChecksumInCache: true, ContentsMatch: true},
		}

		workspaceFiles := []os.FileInfo{
			testutil.MockFileInfo{MockName: "my_file1"},
			testutil.MockFileInfo{MockName: "my_other_file"},
			testutil.MockFileInfo{MockName: "my_dir", MockMode: os.ModeDir},
			testutil.MockFileInfo{MockName: "my_link", MockMode: os.ModeSymlink},
			testutil.MockFileInfo{MockName: "my_file2"},
		}

		expectedDirStatus := artifact.Status{
			WorkspaceFileStatus: fsutil.Directory,
			HasChecksum:         true,
			ChecksumInCache:     true,
			ContentsMatch:       false,
		}

		testDirectoryStatus(dirQuickStatus, manifestFileStatuses, workspaceFiles, expectedDirStatus, t)
	})

	t.Run("file out of date", func(t *testing.T) {
		dirQuickStatus := artifact.Status{
			WorkspaceFileStatus: fsutil.Directory,
			HasChecksum:         true,
			ChecksumInCache:     true,
		}

		manifestFileStatuses := map[string]artifact.Status{
			"my_file1": {WorkspaceFileStatus: fsutil.RegularFile, HasChecksum: true, ChecksumInCache: true, ContentsMatch: true},
			"my_link":  {WorkspaceFileStatus: fsutil.Link, HasChecksum: true, ChecksumInCache: true, ContentsMatch: true},
			"my_file2": {WorkspaceFileStatus: fsutil.RegularFile, HasChecksum: false},
		}

		workspaceFiles := []os.FileInfo{
			testutil.MockFileInfo{MockName: "my_file1"},
			testutil.MockFileInfo{MockName: "my_dir", MockMode: os.ModeDir},
			testutil.MockFileInfo{MockName: "my_link", MockMode: os.ModeSymlink},
			testutil.MockFileInfo{MockName: "my_file2"},
		}

		expectedDirStatus := artifact.Status{
			WorkspaceFileStatus: fsutil.Directory,
			HasChecksum:         true,
			ChecksumInCache:     true,
			ContentsMatch:       false,
		}

		testDirectoryStatus(dirQuickStatus, manifestFileStatuses, workspaceFiles, expectedDirStatus, t)
	})

	t.Run("short circuit on quickStatus results", func(t *testing.T) {
		dirQuickStatus := artifact.Status{
			WorkspaceFileStatus: fsutil.Directory,
			HasChecksum:         true,
			ChecksumInCache:     false,
		}

		// The code under test should return before reading the manifest.
		manifestFileStatuses := map[string]artifact.Status{}

		workspaceFiles := []os.FileInfo{}

		expectedDirStatus := artifact.Status{
			WorkspaceFileStatus: fsutil.Directory,
			HasChecksum:         true,
			ChecksumInCache:     false,
			ContentsMatch:       false,
		}

		testDirectoryStatus(dirQuickStatus, manifestFileStatuses, workspaceFiles, expectedDirStatus, t)
	})
}

func testDirectoryStatus(
	dirQuickStatus artifact.Status,
	manifestFileStatuses map[string]artifact.Status,
	workspaceFiles []os.FileInfo,
	expectedDirStatus artifact.Status,
	t *testing.T,
) {
	fileArtifactStatusOrig := fileArtifactStatus
	fileArtifactStatus = func(ch *LocalCache, workingDir string, art artifact.Artifact) (status artifact.Status, err error) {
		status, ok := manifestFileStatuses[art.Path]
		if !ok {
			t.Fatalf("fileArtifactStatus called with unexpected artifact: %+v", art)
		}
		return
	}
	defer func() { fileArtifactStatus = fileArtifactStatusOrig }()

	readDirOrig := readDir
	readDir = func(dir string) ([]os.FileInfo, error) {
		return workspaceFiles, nil
	}
	defer func() { readDir = readDirOrig }()

	expectedArtifacts := []artifact.Artifact{}
	for path := range manifestFileStatuses {
		expectedArtifacts = append(expectedArtifacts, artifact.Artifact{Path: path})
	}

	readDirManifestOrig := readDirManifest
	readDirManifest = func(path string) (directoryManifest, error) {
		man := directoryManifest{}
		for i := range expectedArtifacts {
			man.Contents = append(man.Contents, &expectedArtifacts[i])
		}
		return man, nil
	}
	defer func() { readDirManifest = readDirManifestOrig }()

	quickStatusOrig := quickStatus
	quickStatus = func(
		ch *LocalCache,
		workingDir string,
		art artifact.Artifact,
	) (status artifact.Status, cachePath, workPath string, err error) {
		status = dirQuickStatus
		workPath = filepath.Join(workingDir, art.Path)
		cachePath = "foobar"
		return
	}
	defer func() { quickStatus = quickStatusOrig }()

	cache, err := NewLocalCache("/cache_root")
	if err != nil {
		t.Fatal(err)
	}
	dirArt := artifact.Artifact{IsDir: true, Checksum: "dummy_checksum", Path: "art_dir"}

	status, err := cache.Status("work_dir", dirArt)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(expectedDirStatus, status); diff != "" {
		t.Fatalf("directory artifact.Status -want +got:\n%s", diff)
	}
}

func TestDirectoryStatusRecursive(t *testing.T) {
	t.Run("up-to-date", func(t *testing.T) { testDirectoryStatusRecursive(true, t) })
	t.Run("out-of-date", func(t *testing.T) { testDirectoryStatusRecursive(false, t) })
}

func testDirectoryStatusRecursive(expectedContentsMatch bool, t *testing.T) {
	dirQuickStatus := artifact.Status{
		WorkspaceFileStatus: fsutil.Directory,
		HasChecksum:         true,
		ChecksumInCache:     true,
	}

	parentDirManifest := directoryManifest{
		Path: "parent_dir",
		Contents: []*artifact.Artifact{
			{Path: "my_file1"},
			{Path: "child_dir", IsDir: true, IsRecursive: true},
			{Path: "my_file2"},
			{Path: "my_link"},
		},
	}
	childDirManifest := directoryManifest{
		Path: "child_dir",
		Contents: []*artifact.Artifact{
			{Path: "nested_file1"},
			{Path: "nested_file2"},
		},
	}

	parentDirFileStatuses := map[string]artifact.Status{
		"my_file1": {WorkspaceFileStatus: fsutil.RegularFile, HasChecksum: true, ChecksumInCache: true, ContentsMatch: true},
		"my_link":  {WorkspaceFileStatus: fsutil.Link, HasChecksum: true, ChecksumInCache: true, ContentsMatch: true},
		"my_file2": {WorkspaceFileStatus: fsutil.RegularFile, HasChecksum: true, ChecksumInCache: true, ContentsMatch: true},
	}

	childDirFileStatuses := map[string]artifact.Status{
		"nested_file1": {WorkspaceFileStatus: fsutil.RegularFile, HasChecksum: true, ChecksumInCache: true, ContentsMatch: true},
		"nested_file2": {WorkspaceFileStatus: fsutil.RegularFile, HasChecksum: true, ChecksumInCache: true, ContentsMatch: true},
	}

	parentDirListing := []os.FileInfo{
		testutil.MockFileInfo{MockName: "my_file1"},
		testutil.MockFileInfo{MockName: "child_dir", MockMode: os.ModeDir},
		testutil.MockFileInfo{MockName: "my_link", MockMode: os.ModeSymlink},
		testutil.MockFileInfo{MockName: "my_file2"},
	}

	childDirListing := []os.FileInfo{
		testutil.MockFileInfo{MockName: "nested_file1"},
		testutil.MockFileInfo{MockName: "nested_file2"},
	}
	if !expectedContentsMatch {
		childDirListing = append(childDirListing, testutil.MockFileInfo{MockName: "another_dir", MockMode: os.ModeDir})
	}

	expectedDirStatus := artifact.Status{
		WorkspaceFileStatus: fsutil.Directory,
		HasChecksum:         true,
		ChecksumInCache:     true,
		ContentsMatch:       expectedContentsMatch,
	}

	fileArtifactStatusOrig := fileArtifactStatus
	fileArtifactStatus = func(ch *LocalCache, workingDir string, art artifact.Artifact) (status artifact.Status, err error) {
		var ok bool
		switch workingDir {
		case filepath.Join("work_dir", "parent_dir"):
			status, ok = parentDirFileStatuses[art.Path]
		case filepath.Join("work_dir", "parent_dir", "child_dir"):
			status, ok = childDirFileStatuses[art.Path]
		default:
			t.Fatalf("unexpected workingDir: %#v", workingDir)
		}
		if !ok {
			t.Fatalf("unxpected fileArtifactStatus call: workingDir= %#v, artifact= %+v", workingDir, art)
		}
		return
	}
	defer func() { fileArtifactStatus = fileArtifactStatusOrig }()

	readDirOrig := readDir
	readDir = func(dir string) ([]os.FileInfo, error) {
		switch dir {
		case filepath.Join("work_dir", "parent_dir"):
			return parentDirListing, nil
		case filepath.Join("work_dir", "parent_dir", "child_dir"):
			return childDirListing, nil
		}
		t.Fatalf("unexpected dir: %#v", dir)
		return parentDirListing, nil // unreachable, but required
	}
	defer func() { readDir = readDirOrig }()

	readDirManifestCallCount := 0
	readDirManifestOrig := readDirManifest
	readDirManifest = func(path string) (directoryManifest, error) {
		readDirManifestCallCount++
		if readDirManifestCallCount == 1 {
			return parentDirManifest, nil
		} else if readDirManifestCallCount == 2 {
			return childDirManifest, nil
		}
		t.Fatal("unexpected call to readDirManifest")
		return parentDirManifest, nil // unreachable, but required
	}
	defer func() { readDirManifest = readDirManifestOrig }()

	quickStatusOrig := quickStatus
	quickStatus = func(
		ch *LocalCache,
		workingDir string,
		art artifact.Artifact,
	) (status artifact.Status, cachePath, workPath string, err error) {
		status = dirQuickStatus
		workPath = filepath.Join(workingDir, art.Path)
		cachePath = "foobar"
		return
	}
	defer func() { quickStatus = quickStatusOrig }()

	cache, err := NewLocalCache("/cache_root")
	if err != nil {
		t.Fatal(err)
	}
	dirArt := artifact.Artifact{IsDir: true, IsRecursive: true, Checksum: "dummy_checksum", Path: "parent_dir"}

	status, err := cache.Status("work_dir", dirArt)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(expectedDirStatus, status); diff != "" {
		t.Fatalf("directory artifact.Status -want +got:\n%s", diff)
	}
}

func TestStatusIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	for _, testCase := range testutil.AllFileTestCases() {
		t.Run(fmt.Sprintf("%+v", testCase), func(t *testing.T) {
			testStatusIntegration(testCase, t)
		})
	}
}

func testStatusIntegration(statusWant artifact.ArtifactWithStatus, t *testing.T) {
	dirs, art, err := testutil.CreateArtifactTestCase(statusWant)
	defer os.RemoveAll(dirs.CacheDir)
	defer os.RemoveAll(dirs.WorkDir)
	if err != nil {
		t.Fatal(err)
	}
	cache, err := NewLocalCache(dirs.CacheDir)
	if err != nil {
		t.Fatal(err)
	}

	statusGot, err := cache.Status(dirs.WorkDir, art)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(statusWant.Status, statusGot); diff != "" {
		t.Fatalf("Status() -want +got:\n%s", diff)
	}
}
