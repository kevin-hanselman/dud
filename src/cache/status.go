package cache

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/kevin-hanselman/dud/src/artifact"
	"github.com/kevin-hanselman/dud/src/checksum"
	"github.com/kevin-hanselman/dud/src/fsutil"
	"github.com/pkg/errors"
)

// Status reports the status of an Artifact in the Cache. If shortCircuit is
// true, Status will exit as soon as the overall state of the Artifact is known
// -- saving time, but potentially leaving artifact.Status.ChildrenStatus
// incomplete. If shortCircuit is false, Status will fully populate
// artifact.Status.ChildrenStatus.
func (ch LocalCache) Status(workspaceDir string, art artifact.Artifact, shortCircuit bool) (
	status artifact.Status,
	err error,
) {
	if art.IsDir {
		status, _, err = dirArtifactStatus(ch, workspaceDir, art, shortCircuit)
	} else {
		status, err = fileArtifactStatus(ch, workspaceDir, art)
	}
	status.Artifact = art
	err = errors.Wrapf(err, "status %s", art.Path)
	return
}

// checksumStatus populates the HasChecksum and ChecksumInCache fields of
// artifact.Status and returns any relevant cache file information.
func checksumStatus(ch LocalCache, art artifact.Artifact) (
	status artifact.Status,
	cachePath string,
	cacheFileInfo fs.FileInfo,
	err error,
) {
	cachePath, err = ch.PathForChecksum(art.Checksum)
	absCachePath := filepath.Join(ch.dir, cachePath)

	if _, ok := err.(InvalidChecksumError); ok {
		err = nil
		status.HasChecksum = false
	} else if err != nil {
		return
	} else {
		status.HasChecksum = true
		cacheFileInfo, err = os.Stat(absCachePath)
		if err == nil {
			status.ChecksumInCache = true
		} else if os.IsNotExist(err) {
			err = nil
			status.ChecksumInCache = false
		}
	}
	return
}

// quickStatus populates all artifact.Status fields except for ContentsMatch
// and ChildrenStatus. However, this function will set ContentsMatch if the
// Artifact is a file, the workspace file is a link, and the other status
// booleans are true. Checking to see if a link points to the cache is, as this
// function suggests, quick.
var quickStatus = func(
	ch LocalCache,
	workspaceDir string,
	art artifact.Artifact,
) (status artifact.Status, cachePath, workPath string, err error) {
	// These FileInfos are used to verify a committed file is correctly linked
	// to the cache.
	var cacheFileInfo, workFileInfo fs.FileInfo

	status, cachePath, cacheFileInfo, err = checksumStatus(ch, art)
	if err != nil {
		return
	}

	workPath = filepath.Join(workspaceDir, art.Path)
	status.WorkspaceFileStatus, err = fsutil.FileStatusFromPath(workPath)
	if err != nil {
		return
	}
	if status.HasChecksum &&
		status.ChecksumInCache &&
		status.WorkspaceFileStatus == fsutil.StatusLink {
		workFileInfo, err = os.Stat(workPath)
		// A NotExist error here means the link is dead. Leave ContentsMatch as
		// false and let the caller handle the invalid link.
		if os.IsNotExist(err) {
			err = nil
			return
		} else if err != nil {
			return
		}
		status.ContentsMatch = os.SameFile(cacheFileInfo, workFileInfo)
	}
	return
}

func fileArtifactStatus(
	ch LocalCache,
	workspaceDir string,
	art artifact.Artifact,
) (artifact.Status, error) {
	status, cachePath, workPath, err := quickStatus(ch, workspaceDir, art)
	if err != nil {
		return status, err
	}
	cachePath = filepath.Join(ch.dir, cachePath)

	if status.WorkspaceFileStatus != fsutil.StatusRegularFile {
		return status, nil
	}

	if art.SkipCache {
		if !status.HasChecksum {
			return status, nil
		}
		fileReader, err := os.Open(workPath)
		if err != nil {
			return status, err
		}
		defer fileReader.Close()
		workspaceFileChecksum, err := checksum.Checksum(fileReader)
		if err != nil {
			return status, err
		}
		status.ContentsMatch = workspaceFileChecksum == art.Checksum
	} else {
		if !status.ChecksumInCache {
			return status, nil
		}
		status.ContentsMatch, err = fsutil.SameContents(workPath, cachePath)
		if err != nil {
			return status, err
		}
	}
	return status, nil
}

func dirArtifactStatus(
	ch LocalCache,
	workspaceDir string,
	art artifact.Artifact,
	shortCircuit bool,
) (artifact.Status, directoryManifest, error) {
	var manifest directoryManifest
	status, cachePath, workPath, err := quickStatus(ch, workspaceDir, art)
	if err != nil {
		return status, manifest, err
	}
	cachePath = filepath.Join(ch.dir, cachePath)

	if !(status.HasChecksum && status.ChecksumInCache) && shortCircuit {
		return status, manifest, nil
	}

	if status.WorkspaceFileStatus != fsutil.StatusDirectory {
		return status, manifest, nil
	}

	// Any child Artifact that's out-of-date or not committed will flip this
	// value.
	status.ContentsMatch = true
	status.ChildrenStatus = make(map[string]*artifact.Status)

	// First, ensure all artifacts in the directoryManifest are up-to-date.
	if status.ChecksumInCache {
		manifest, err = readDirManifest(cachePath)
		if err != nil {
			return status, manifest, err
		}

		for path, art := range manifest.Contents {
			childStatus, err := ch.Status(workPath, *art, shortCircuit)
			if err != nil {
				return status, manifest, err
			}
			status.ChildrenStatus[path] = &childStatus
			status.ContentsMatch = status.ContentsMatch && childStatus.ContentsMatch
			if shortCircuit && !status.ContentsMatch {
				return status, manifest, nil
			}
		}
	}

	// Second, get a directory listing and check for untracked files.
	entries, err := os.ReadDir(workPath)
	if err != nil {
		return status, manifest, err
	}
	for _, entry := range entries {
		newArt := artifact.Artifact{Path: entry.Name(), IsDir: entry.IsDir()}
		// We've already checked all entries in the manifest.
		// (While assigning to a nil map panics, accessing a nil map is safe.)
		if _, ok := manifest.Contents[newArt.Path]; ok {
			continue
		}
		// Ignore sub-directories if recursion is disabled.
		if newArt.IsDir && art.DisableRecursion {
			continue
		}
		// After the two exceptions above, the presence of any untracked
		// files/directories means this Artifact is out-of-date.
		// Therefore we can exit early if needed.
		status.ContentsMatch = false
		if shortCircuit {
			return status, manifest, nil
		}
		childStatus, err := ch.Status(workPath, newArt, shortCircuit)
		if err != nil {
			return status, manifest, err
		}
		status.ChildrenStatus[newArt.Path] = &childStatus
	}
	return status, manifest, nil
}

func readDirManifest(path string) (man directoryManifest, err error) {
	var f *os.File
	f, err = os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&man)
	return
}
