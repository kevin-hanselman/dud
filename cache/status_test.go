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
	fileArtifactStatusCalls := []statusArgs{}
	fileArtifactStatusOrig := fileArtifactStatus
	fileArtifactStatus = func(args statusArgs) (status artifact.Status, err error) {
		fileArtifactStatusCalls = append(fileArtifactStatusCalls, args)
		status = artifact.Status{
			WorkspaceFileStatus: artifact.RegularFile,
			HasChecksum:         true,
			ChecksumInCache:     true,
			ContentsMatch:       true,
		}
		return
	}
	defer func() { fileArtifactStatus = fileArtifactStatusOrig }()

	mockFiles := []os.FileInfo{
		testutil.MockFileInfo{MockName: "my_file1"},
		testutil.MockFileInfo{MockName: "my_dir", MockMode: os.ModeDir},
		// TODO: cover handling of symlinks (and other irregular files?)
		testutil.MockFileInfo{MockName: "my_link", MockMode: os.ModeSymlink},
		testutil.MockFileInfo{MockName: "my_file2"},
	}

	expectedArtifacts := []*artifact.Artifact{
		{Path: "my_file1"},
		{Path: "my_link"},
		{Path: "my_file2"},
	}
	t.Logf("%#v", expectedArtifacts)

	readDirOrig := readDir
	readDir = func(dir string) ([]os.FileInfo, error) {
		return mockFiles, nil
	}
	defer func() { readDir = readDirOrig }()

	readDirManifestOrig := readDirManifest
	readDirManifest = func(path string) (directoryManifest, error) {
		man := directoryManifest{}
		for _, art := range expectedArtifacts {
			man.Contents = append(man.Contents, art)
		}
		return man, nil
	}
	defer func() { readDirManifest = readDirManifestOrig }()

	quickStatusOrig := quickStatus
	quickStatus = func(args statusArgs) (status artifact.Status, cachePath, workPath string, err error) {
		status = artifact.Status{
			WorkspaceFileStatus: artifact.Directory,
			HasChecksum:         true,
			ChecksumInCache:     true,
		}
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

	expectedfileArtifactStatusCalls := []statusArgs{}
	baseDir := path.Join("work_dir", "art_dir")
	for i := range expectedArtifacts {
		expectedfileArtifactStatusCalls = append(
			expectedfileArtifactStatusCalls,
			statusArgs{Cache: &cache, WorkingDir: baseDir, Artifact: *expectedArtifacts[i]},
		)
	}

	if diff := cmp.Diff(expectedfileArtifactStatusCalls, fileArtifactStatusCalls); diff != "" {
		t.Fatalf("fileArtifactStatusCalls -want +got:\n%s", diff)
	}

	expectedStatus := artifact.Status{
		WorkspaceFileStatus: artifact.Directory,
		HasChecksum:         true,
		ChecksumInCache:     true,
		ContentsMatch:       true,
	}

	if diff := cmp.Diff(expectedStatus, status); diff != "" {
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
