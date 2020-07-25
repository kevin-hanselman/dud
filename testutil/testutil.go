package testutil

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/kevlar1818/duc/artifact"
	"github.com/kevlar1818/duc/fsutil"
)

// MockFileInfo mocks os.FileInfo
type MockFileInfo struct {
	MockName    string
	MockSize    int64
	MockMode    os.FileMode
	MockModTime time.Time
}

// Name returns the basename of the file
func (m MockFileInfo) Name() string {
	return m.MockName
}

// Size return the size of the file
func (m MockFileInfo) Size() int64 {
	return m.MockSize
}

// Mode returns the os.FileMode of the file
func (m MockFileInfo) Mode() os.FileMode {
	return m.MockMode
}

// ModTime returns the modification time of the file
func (m MockFileInfo) ModTime() time.Time {
	return m.MockModTime
}

// IsDir returns true if the file is a directory, false otherwise
func (m MockFileInfo) IsDir() bool {
	return m.Mode().IsDir()
}

// Sys only exists to satisfy the FileInfo interface; it always returns nil
func (m MockFileInfo) Sys() interface{} {
	return nil
}

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

// AllFileTestCases returns a slice of all valid artifact.Status structs.
// (See status.txt in the project root)
func AllFileTestCases() (out []artifact.ArtifactWithStatus) {
	allArtifactStatuses := []artifact.Status{}
	allWorkspaceStatuses := []fsutil.FileStatus{
		fsutil.Absent,
		fsutil.RegularFile,
		fsutil.Link,
	}
	for _, workspaceStatus := range allWorkspaceStatuses {
		if workspaceStatus != fsutil.Absent {
			allArtifactStatuses = append(
				allArtifactStatuses,
				artifact.Status{
					WorkspaceFileStatus: workspaceStatus,
					HasChecksum:         true,
					ChecksumInCache:     true,
					ContentsMatch:       true,
				},
			)
		}
		allArtifactStatuses = append(
			allArtifactStatuses,
			artifact.Status{
				WorkspaceFileStatus: workspaceStatus,
				HasChecksum:         true,
				ChecksumInCache:     true,
				ContentsMatch:       false,
			},
			artifact.Status{
				WorkspaceFileStatus: workspaceStatus,
				HasChecksum:         true,
				ChecksumInCache:     false,
			},
			artifact.Status{
				WorkspaceFileStatus: workspaceStatus,
				HasChecksum:         false,
			},
		)
	}
	// Promote Status structs to ArtifactWithStatus.
	for _, artStatus := range allArtifactStatuses {
		artWithStatus := artifact.ArtifactWithStatus{
			Artifact: artifact.Artifact{SkipCache: false},
			Status:   artStatus,
		}
		out = append(out, artWithStatus)
	}

	artWithStatus := artifact.ArtifactWithStatus{
		Artifact: artifact.Artifact{SkipCache: true},
		Status:   artifact.Status{WorkspaceFileStatus: fsutil.RegularFile},
	}
	out = append(out, artWithStatus)
	artWithStatus.HasChecksum = true
	out = append(out, artWithStatus)
	artWithStatus.ContentsMatch = true
	out = append(out, artWithStatus)
	return
}

// CreateArtifactTestCase sets up an integration test environment with a single
// artifact that complies with the provided artifact.Status.
func CreateArtifactTestCase(status artifact.ArtifactWithStatus) (dirs TempDirs, art artifact.Artifact, err error) {
	dirs, err = CreateTempDirs()
	if err != nil {
		return
	}

	fileContents := []byte("Hello, World!")
	// Put the file in a sub-directory to test that full paths can be recreated with a checkout.
	parentDir := "greet"

	art = status.Artifact
	art.Checksum = "288a86a79f20a3d6dccdca7713beaed178798296bdfa7913fa2a62d9727bf8f8"
	art.Path = filepath.Join(parentDir, "hello.txt")

	fileCacheDir := filepath.Join(dirs.CacheDir, art.Checksum[:2])
	fileCachePath := filepath.Join(fileCacheDir, art.Checksum[2:])
	fileWorkspacePath := filepath.Join(dirs.WorkDir, art.Path)

	if !status.HasChecksum {
		art.Checksum = ""
	}

	if status.ChecksumInCache {
		if err = os.Mkdir(fileCacheDir, 0755); err != nil {
			return
		}
		if err = ioutil.WriteFile(fileCachePath, fileContents, 0444); err != nil {
			return
		}
	}

	if status.WorkspaceFileStatus != fsutil.Absent {
		if err = os.Mkdir(filepath.Join(dirs.WorkDir, parentDir), 0755); err != nil {
			return
		}
	}

	switch status.WorkspaceFileStatus {
	case fsutil.RegularFile:
		if !status.ContentsMatch {
			fileContents = []byte("Not the same as in the cache")
		}
		if err = ioutil.WriteFile(fileWorkspacePath, fileContents, 0644); err != nil {
			return
		}
	case fsutil.Link:
		targetPath := fileCachePath
		if !status.ContentsMatch {
			targetPath = "foobar"
		}
		if err = os.Symlink(targetPath, fileWorkspacePath); err != nil {
			return
		}
	}
	return
}
