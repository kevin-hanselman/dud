package testutil

import (
	"github.com/kevlar1818/duc/artifact"
	"io/ioutil"
	"os"
	"path"
)

// TempDirs holds a pair of cache and workspace directory paths for integration testing.
type TempDirs struct {
	CacheDir string
	WorkDir  string
}

// CreateTempDirs creates a DUC cache and workspace in the OS temp FS.
func CreateTempDirs() (dirs TempDirs, err error) {
	dirs.CacheDir, err = ioutil.TempDir("", "duc_cache")
	if err != nil {
		return
	}
	dirs.WorkDir, err = ioutil.TempDir("", "duc_wspace")
	if err != nil {
		os.RemoveAll(dirs.CacheDir)
		return
	}
	return
}

// AllTestCases returns a slice of all possible combinations TestCaseArgs values.
func AllTestCases() (allArgs []artifact.Status) {
	for _, inCache := range []bool{true, false} {
		for _, wspaceStatus := range []artifact.WorkspaceStatus{artifact.Absent, artifact.RegularFile, artifact.Link} {
			allArgs = append(
				allArgs,
				artifact.Status{
					ChecksumInCache: inCache, WorkspaceStatus: wspaceStatus,
				},
			)
		}
	}
	return allArgs
}

// CreateArtifactTestCase sets up an integration test environment with a single
// artifact according the arguments provided. The bool argument specifies whether the
// artifact is present in the cache.
func CreateArtifactTestCase(status artifact.Status) (dirs TempDirs, art artifact.Artifact, err error) {
	dirs, err = CreateTempDirs()
	if err != nil {
		return
	}

	fileContents := []byte("Hello, World!")

	art = artifact.Artifact{
		Checksum: "0a0a9f2a6772942557ab5355d76af442f8f65e01",
		Path:     "hello.txt",
	}

	fileCacheDir := path.Join(dirs.CacheDir, art.Checksum[:2])
	fileCachePath := path.Join(fileCacheDir, art.Checksum[2:])
	fileWorkspacePath := path.Join(dirs.WorkDir, art.Path)

	if status.ChecksumInCache {
		if err = os.Mkdir(fileCacheDir, 0755); err != nil {
			return
		}
		if err = ioutil.WriteFile(fileCachePath, fileContents, 0444); err != nil {
			return
		}
	}
	// TODO: if !status.ContentsInCache && !WorkspaceStatus != Absent, change workspace file contents

	switch status.WorkspaceStatus {
	case artifact.RegularFile:
		if err = ioutil.WriteFile(fileWorkspacePath, fileContents, 0644); err != nil {
			return
		}
	case artifact.Link:
		if err = os.Symlink(fileCachePath, fileWorkspacePath); err != nil {
			return
		}
	}
	return
}
