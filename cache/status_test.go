package cache

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/kevlar1818/duc/artifact"
	"github.com/kevlar1818/duc/testutil"
	"os"
	"path"
	"testing"
)

func TestDirectoryStatus(t *testing.T) {

	t.Run("happy path: all files up to date", func(t *testing.T) {
		dirQuickStatus := artifact.Status{
			WorkspaceFileStatus: artifact.Directory,
			HasChecksum:         true,
			ChecksumInCache:     true,
		}

		manifestFileStatuses := map[string]artifact.Status{
			"my_file1": {WorkspaceFileStatus: artifact.RegularFile, HasChecksum: true, ChecksumInCache: true, ContentsMatch: true},
			"my_link":  {WorkspaceFileStatus: artifact.Link, HasChecksum: true, ChecksumInCache: true, ContentsMatch: true},
			"my_file2": {WorkspaceFileStatus: artifact.RegularFile, HasChecksum: true, ChecksumInCache: true, ContentsMatch: true},
		}

		workspaceFiles := []os.FileInfo{
			testutil.MockFileInfo{MockName: "my_file1"},
			testutil.MockFileInfo{MockName: "my_dir", MockMode: os.ModeDir},
			// TODO: cover handling of symlinks (and other irregular files?)
			testutil.MockFileInfo{MockName: "my_link", MockMode: os.ModeSymlink},
			testutil.MockFileInfo{MockName: "my_file2"},
		}

		expectedDirStatus := artifact.Status{
			WorkspaceFileStatus: artifact.Directory,
			HasChecksum:         true,
			ChecksumInCache:     true,
			ContentsMatch:       true,
		}

		testDirectoryStatus(dirQuickStatus, manifestFileStatuses, workspaceFiles, expectedDirStatus, t)
	})

	t.Run("file missing from workspace", func(t *testing.T) {
		dirQuickStatus := artifact.Status{
			WorkspaceFileStatus: artifact.Directory,
			HasChecksum:         true,
			ChecksumInCache:     true,
		}

		manifestFileStatuses := map[string]artifact.Status{
			"my_file1": {WorkspaceFileStatus: artifact.Absent},
			"my_link":  {WorkspaceFileStatus: artifact.Link, HasChecksum: true, ChecksumInCache: true, ContentsMatch: true},
			"my_file2": {WorkspaceFileStatus: artifact.RegularFile, HasChecksum: true, ChecksumInCache: true, ContentsMatch: true},
		}

		workspaceFiles := []os.FileInfo{
			testutil.MockFileInfo{MockName: "my_dir", MockMode: os.ModeDir},
			testutil.MockFileInfo{MockName: "my_link", MockMode: os.ModeSymlink},
			testutil.MockFileInfo{MockName: "my_file2"},
		}

		expectedDirStatus := artifact.Status{
			WorkspaceFileStatus: artifact.Directory,
			HasChecksum:         true,
			ChecksumInCache:     true,
			ContentsMatch:       false,
		}

		testDirectoryStatus(dirQuickStatus, manifestFileStatuses, workspaceFiles, expectedDirStatus, t)
	})

	t.Run("untracked file in dir", func(t *testing.T) {
		dirQuickStatus := artifact.Status{
			WorkspaceFileStatus: artifact.Directory,
			HasChecksum:         true,
			ChecksumInCache:     true,
		}

		manifestFileStatuses := map[string]artifact.Status{
			"my_file1": {WorkspaceFileStatus: artifact.RegularFile, HasChecksum: true, ChecksumInCache: true, ContentsMatch: true},
			"my_link":  {WorkspaceFileStatus: artifact.Link, HasChecksum: true, ChecksumInCache: true, ContentsMatch: true},
			"my_file2": {WorkspaceFileStatus: artifact.RegularFile, HasChecksum: true, ChecksumInCache: true, ContentsMatch: true},
		}

		workspaceFiles := []os.FileInfo{
			testutil.MockFileInfo{MockName: "my_file1"},
			testutil.MockFileInfo{MockName: "my_other_file"},
			testutil.MockFileInfo{MockName: "my_dir", MockMode: os.ModeDir},
			testutil.MockFileInfo{MockName: "my_link", MockMode: os.ModeSymlink},
			testutil.MockFileInfo{MockName: "my_file2"},
		}

		expectedDirStatus := artifact.Status{
			WorkspaceFileStatus: artifact.Directory,
			HasChecksum:         true,
			ChecksumInCache:     true,
			ContentsMatch:       false,
		}

		testDirectoryStatus(dirQuickStatus, manifestFileStatuses, workspaceFiles, expectedDirStatus, t)
	})

	t.Run("file out of date", func(t *testing.T) {
		dirQuickStatus := artifact.Status{
			WorkspaceFileStatus: artifact.Directory,
			HasChecksum:         true,
			ChecksumInCache:     true,
		}

		manifestFileStatuses := map[string]artifact.Status{
			"my_file1": {WorkspaceFileStatus: artifact.RegularFile, HasChecksum: true, ChecksumInCache: true, ContentsMatch: true},
			"my_link":  {WorkspaceFileStatus: artifact.Link, HasChecksum: true, ChecksumInCache: true, ContentsMatch: true},
			"my_file2": {WorkspaceFileStatus: artifact.RegularFile, HasChecksum: false},
		}

		workspaceFiles := []os.FileInfo{
			testutil.MockFileInfo{MockName: "my_file1"},
			testutil.MockFileInfo{MockName: "my_dir", MockMode: os.ModeDir},
			testutil.MockFileInfo{MockName: "my_link", MockMode: os.ModeSymlink},
			testutil.MockFileInfo{MockName: "my_file2"},
		}

		expectedDirStatus := artifact.Status{
			WorkspaceFileStatus: artifact.Directory,
			HasChecksum:         true,
			ChecksumInCache:     true,
			ContentsMatch:       false,
		}

		testDirectoryStatus(dirQuickStatus, manifestFileStatuses, workspaceFiles, expectedDirStatus, t)
	})

	t.Run("shortcircuit on quickStatus results", func(t *testing.T) {
		dirQuickStatus := artifact.Status{
			WorkspaceFileStatus: artifact.Directory,
			HasChecksum:         true,
			ChecksumInCache:     false,
		}

		manifestFileStatuses := map[string]artifact.Status{}

		workspaceFiles := []os.FileInfo{}

		expectedDirStatus := artifact.Status{
			WorkspaceFileStatus: artifact.Directory,
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
			t.Errorf("fileArtifactStatus called with unexpected artifact: %+v", art)
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
	quickStatus = func(ch *LocalCache, workingDir string, art artifact.Artifact) (status artifact.Status, cachePath, workPath string, err error) {
		status = dirQuickStatus
		workPath = path.Join(workingDir, art.Path)
		cachePath = "foobar"
		return
	}
	defer func() { quickStatus = quickStatusOrig }()

	cache, err := NewLocalCache("/cache_root")
	if err != nil {
		t.Fatal(err)
	}
	dirArt := artifact.Artifact{IsDir: true, Checksum: "dummy_checksum", Path: "art_dir"}

	status, commitErr := cache.Status("work_dir", dirArt)
	if commitErr != nil {
		t.Fatal(commitErr)
	}

	if diff := cmp.Diff(expectedDirStatus, status); diff != "" {
		t.Fatalf("directory artifact.Status -want +got:\n%s", diff)
	}
}

func TestStatusIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	for _, testCase := range testutil.AllTestCases() {
		t.Run(fmt.Sprintf("%+v", testCase), func(t *testing.T) {
			testStatusIntegration(testCase, t)
		})
	}
}

func testStatusIntegration(statusWant artifact.Status, t *testing.T) {
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
	if diff := cmp.Diff(statusWant, statusGot); diff != "" {
		t.Fatalf("Status() -want +got:\n%s", diff)
	}
}
