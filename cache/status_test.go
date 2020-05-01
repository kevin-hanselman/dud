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
	dirQuickStatus := artifact.Status{
		WorkspaceFileStatus: artifact.Directory,
		HasChecksum:         true,
		ChecksumInCache:     true,
	}

	expectedFileStatuses := map[string]artifact.Status{
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

	testDirectoryStatus(dirQuickStatus, expectedFileStatuses, workspaceFiles, expectedDirStatus, t)
}

func testDirectoryStatus(
	dirQuickStatus artifact.Status,
	expectedFileStatuses map[string]artifact.Status,
	workspaceFiles []os.FileInfo,
	expectedDirStatus artifact.Status,
	t *testing.T,
) {
	fileArtifactStatusCalls := []statusArgs{}
	fileArtifactStatusOrig := fileArtifactStatus
	fileArtifactStatus = func(args statusArgs) (status artifact.Status, err error) {
		fileArtifactStatusCalls = append(fileArtifactStatusCalls, args)
		status = expectedFileStatuses[args.Artifact.Path]
		return
	}
	defer func() { fileArtifactStatus = fileArtifactStatusOrig }()

	readDirOrig := readDir
	readDir = func(dir string) ([]os.FileInfo, error) {
		return workspaceFiles, nil
	}
	defer func() { readDir = readDirOrig }()

	expectedArtifacts := []artifact.Artifact{}
	for path := range expectedFileStatuses {
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
	quickStatus = func(args statusArgs) (status artifact.Status, cachePath, workPath string, err error) {
		status = dirQuickStatus
		workPath = path.Join(args.WorkingDir, args.Artifact.Path)
		cachePath = "foobar"
		return
	}
	defer func() { quickStatus = quickStatusOrig }()

	cache := LocalCache{Dir: "cache_root"}
	dirArt := artifact.Artifact{IsDir: true, Checksum: "", Path: "art_dir"}

	status, commitErr := cache.Status("work_dir", dirArt)
	if commitErr != nil {
		t.Fatal(commitErr)
	}

	expectedFileArtifactStatusCalls := []statusArgs{}
	baseDir := path.Join("work_dir", "art_dir")
	for i := range expectedArtifacts {
		expectedFileArtifactStatusCalls = append(
			expectedFileArtifactStatusCalls,
			statusArgs{
				Cache:      &cache,
				WorkingDir: baseDir,
				Artifact:   expectedArtifacts[i],
			},
		)
	}

	if diff := cmp.Diff(expectedFileArtifactStatusCalls, fileArtifactStatusCalls); diff != "" {
		t.Fatalf("fileArtifactStatusCalls -want +got:\n%s", diff)
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
	cache := LocalCache{Dir: dirs.CacheDir}

	statusGot, err := cache.Status(dirs.WorkDir, art)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(statusWant, statusGot); diff != "" {
		t.Fatalf("Status() -want +got:\n%s", diff)
	}
}
