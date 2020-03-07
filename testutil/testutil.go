package testutil

import (
	"io/ioutil"
)

type ArtifactChecksumStatus int
type ArtifactWorkspaceStatus int

const (
	ChecksumEmpty ArtifactChecksumStatus = iota
	ChecksumCurrent
	ChecksumStale

	WorkspaceAbsent ArtifactWorkspaceStatus = iota
	WorkspaceCopy
	WorkspaceLink
)

type ArtifactTestCaseArgs struct {
	InCache bool
	WorkspaceStatus ArtifactWorkspaceStatus
	ChecksumStatus ArtifactChecksumStatus
}

// CreateTempDirs creates a DUC cache and workspace in the OS temp FS.
func CreateTempDirs() (cacheDir, workDir string, err error) {
	cacheDir, err = ioutil.TempDir("", "duc_cache")
	if err != nil {
		return "", "", err
	}
	workDir, err = ioutil.TempDir("", "duc_wspace")
	if err != nil {
		return "", "", err
	}
	return
}
