package cache

import (
	"encoding/json"
	"github.com/kevlar1818/duc/artifact"
	"github.com/kevlar1818/duc/fsutil"
	"github.com/pkg/errors"
	"os"
	"path"
)

// Status reports the status of an Artifact in the Cache.
func (ch *LocalCache) Status(workingDir string, art artifact.Artifact) (artifact.Status, error) {
	if art.IsDir {
		return dirArtifactStatus(ch, workingDir, art)
	}
	return fileArtifactStatus(ch, workingDir, art)
}

// quickStatus populates all artifact.Status fields except for ContentsMatch.
// TODO: It may be worth exposing this version of status (bypassing the full
// status check) using a CLI flag
var quickStatus = func(ch *LocalCache, workingDir string, art artifact.Artifact) (status artifact.Status, cachePath, workPath string, err error) {
	workPath = path.Join(workingDir, art.Path)
	cachePath, err = ch.PathForChecksum(art.Checksum)
	if err != nil { // An error means the checksum is invalid
		status.HasChecksum = false
	} else {
		status.HasChecksum = true
		status.ChecksumInCache, err = fsutil.Exists(cachePath, false) // TODO: check for regular file?
		if err != nil {
			return
		}
	}
	status.WorkspaceFileStatus, err = fsutil.FileStatusFromPath(workPath)
	return
}

var fileArtifactStatus = func(ch *LocalCache, workingDir string, art artifact.Artifact) (artifact.Status, error) {
	status, cachePath, workPath, err := quickStatus(ch, workingDir, art)
	if err != nil {
		return status, errors.Wrap(err, "fileStatus")
	}

	if !status.ChecksumInCache {
		return status, nil
	}

	switch status.WorkspaceFileStatus {
	case artifact.RegularFile:
		status.ContentsMatch, err = fsutil.SameContents(workPath, cachePath)
		if err != nil {
			return status, errors.Wrap(err, "fileStatus")
		}
	case artifact.Link:
		// TODO: make this a helper function? (to remove os dep)
		linkDst, err := os.Readlink(workPath)
		if err != nil {
			return status, errors.Wrap(err, "fileStatus")
		}
		status.ContentsMatch = linkDst == cachePath
	}
	return status, nil
}

func dirArtifactStatus(ch *LocalCache, workingDir string, art artifact.Artifact) (artifact.Status, error) {
	status, cachePath, workPath, err := quickStatus(ch, workingDir, art)
	if err != nil {
		return status, errors.Wrap(err, "dirStatus")
	}

	if !(status.HasChecksum && status.ChecksumInCache) {
		return status, nil
	}

	if status.WorkspaceFileStatus != artifact.Directory {
		return status, nil
	}

	dirManifest, err := readDirManifest(cachePath)
	if err != nil {
		return status, errors.Wrap(err, "dirStatus")
	}

	// first, ensure all artifacts in the directoryManifest are up-to-date;
	// quit early if any are not.
	manFilePaths := make(map[string]bool)
	for _, fileArt := range dirManifest.Contents {
		manFilePaths[fileArt.Path] = true
		artStatus, err := fileArtifactStatus(ch, workPath, *fileArt)
		if err != nil {
			return status, errors.Wrap(err, "dirStatus")
		}
		if !artStatus.ContentsMatch {
			return status, nil
		}
	}

	// second, get a directory listing and check for untracked files;
	// quit early if any exist.
	entries, err := readDir(workPath)
	if err != nil {
		return status, errors.Wrap(err, "dirStatus")
	}
	for _, entry := range entries {
		// TODO: only proceed for reg files and links
		if !entry.IsDir() && !manFilePaths[entry.Name()] {
			return status, nil
		}
	}
	status.ContentsMatch = true
	return status, nil
}

var readDirManifest = func(path string) (man directoryManifest, err error) {
	manifestFile, err := os.Open(path)
	if err != nil {
		return
	}
	err = json.NewDecoder(manifestFile).Decode(&man)
	return
}
