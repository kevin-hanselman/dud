package cache

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/kevlar1818/duc/artifact"
	"github.com/kevlar1818/duc/testutil"
	"os"
	"testing"
)

func TestDirectoryStatus(t *testing.T) {
	fileArtifactStatusCalls := []statusArgs{}
	fileArtifactStatusOrig := fileArtifactStatus
	fileArtifactStatus = func(args statusArgs) (artifact.Status, error) {
		fileArtifactStatusCalls = append(fileArtifactStatusCalls, args)
		return artifact.Status{}, nil
	}
	defer func() { fileArtifactStatus = fileArtifactStatusOrig }()

	mockFiles := []os.FileInfo{
		testutil.MockFileInfo{MockName: "my_file1"},
		testutil.MockFileInfo{MockName: "my_dir", MockMode: os.ModeDir},
		// TODO: cover handling of symlinks (and other irregular files?)
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

	status, commitErr := cache.Status("work_dir", dirArt)
	if commitErr != nil {
		t.Fatal(commitErr)
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
