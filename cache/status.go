package cache

import (
	"github.com/kevlar1818/duc/artifact"
	"github.com/kevlar1818/duc/fsutil"
	"os"
	"path"
)

// Status reports the status of an Artifact in the Cache.
func (ch *LocalCache) Status(workingDir string, art artifact.Artifact) (artifact.Status, error) {
	args := statusArgs{
		Cache:      ch,
		WorkingDir: workingDir,
		Artifact:   art,
	}
	if art.IsDir {
		return dirArtifactStatus(args)
	}
	return fileArtifactStatus(args)
}

type statusArgs struct {
	Cache      *LocalCache
	WorkingDir string
	Artifact   artifact.Artifact
}

// quickStatus populates all artifact.Status fields except for ContentsMatch.
// TODO: It may be worth exposing this version of status (bypassing the full
// status check) using a CLI flag
var quickStatus = func(args statusArgs) (status artifact.Status, cachePath, workPath string, err error) {
	workPath = path.Join(args.WorkingDir, args.Artifact.Path)
	cachePath, err = args.Cache.PathForChecksum(args.Artifact.Checksum)
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

var fileArtifactStatus = func(args statusArgs) (artifact.Status, error) {
	status, cachePath, workPath, err := quickStatus(args)

	if !status.ChecksumInCache {
		return status, nil
	}

	switch status.WorkspaceFileStatus {
	case artifact.RegularFile:
		status.ContentsMatch, err = fsutil.SameContents(workPath, cachePath)
		if err != nil {
			return status, err
		}
	case artifact.Link:
		// TODO: make this a helper function? (to remove os dep)
		linkDst, err := os.Readlink(workPath)
		if err != nil {
			return status, err
		}
		status.ContentsMatch = linkDst == cachePath
	}
	return status, nil
}

func dirArtifactStatus(args statusArgs) (artifact.Status, error) {
	status, _, baseDir, err := quickStatus(args)
	if err != nil {
		return status, err
	}

	if !status.ChecksumInCache {
		return status, nil
	}

	if status.WorkspaceFileStatus != artifact.Directory {
		return status, nil
	}

	entries, err := readDir(baseDir)
	if err != nil {
		return status, err
	}

	//manifest := directoryManifest{Path: baseDir}
	for _, entry := range entries {
		if !entry.Mode().IsDir() { // TODO: only proceed for reg files and links
			fileArt := artifact.Artifact{Path: entry.Name()}
			_, err := fileArtifactStatus(statusArgs{
				Cache:      args.Cache,
				WorkingDir: baseDir,
				Artifact:   fileArt,
			})
			if err != nil {
				return status, err
			}
		}
	}
	return status, nil
}
