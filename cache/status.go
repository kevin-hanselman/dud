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

var fileArtifactStatus = func(args statusArgs) (artifact.Status, error) {
	var status artifact.Status
	workPath := path.Join(args.WorkingDir, args.Artifact.Path)
	cachePath, err := args.Cache.PathForChecksum(args.Artifact.Checksum)
	if err != nil { // An error means the checksum is invalid
		status.HasChecksum = false
	} else {
		status.HasChecksum = true
		status.ChecksumInCache, err = fsutil.Exists(cachePath, false) // TODO: check for regular file?
		if err != nil {
			return status, err
		}
	}

	exists, err := fsutil.Exists(workPath, false)
	if err != nil {
		return status, err
	}

	if !exists {
		status.WorkspaceStatus = artifact.Absent
		return status, nil
	}

	isReg, err := fsutil.IsRegularFile(workPath)
	if err != nil {
		return status, err
	}

	isLink, err := fsutil.IsLink(workPath)
	if err != nil {
		return status, err
	}

	if isReg {
		status.WorkspaceStatus = artifact.RegularFile
		if status.ChecksumInCache {
			status.ContentsMatch, err = fsutil.SameContents(workPath, cachePath)
			if err != nil {
				return status, err
			}
		}
	} else if isLink {
		status.WorkspaceStatus = artifact.Link
		if status.ChecksumInCache {
			linkDst, err := os.Readlink(workPath)
			if err != nil {
				return status, err
			}
			status.ContentsMatch = linkDst == cachePath
		}
	}

	return status, nil
}

func dirArtifactStatus(args statusArgs) (artifact.Status, error) {
	var dirStatus artifact.Status
	baseDir := path.Join(args.WorkingDir, args.Artifact.Path)
	entries, err := readDir(baseDir)
	if err != nil {
		return dirStatus, err
	}
	// manifest := directoryManifest{Path: baseDir}
	for _, entry := range entries {
		if !entry.Mode().IsDir() { // TODO: only proceed for reg files and links
			fileArt := artifact.Artifact{Path: entry.Name()}
			_, err := fileArtifactStatus(statusArgs{
				Cache:      args.Cache,
				WorkingDir: baseDir,
				Artifact:   fileArt,
			})
			if err != nil {
				return dirStatus, err
			}
		}
	}
	return dirStatus, nil
}
