package cache

import (
	"path/filepath"

	"github.com/kevin-hanselman/dud/src/artifact"
)

// Fetch downloads an Artifact from a remote location to the local cache.
func (ch LocalCache) Fetch(workspaceDir, remoteSrc string, art artifact.Artifact) error {
	fetchFiles := make(map[string]struct{})
	if err := gatherFilesToFetch(ch, workspaceDir, art, remoteSrc, fetchFiles); err != nil {
		return err
	}
	if len(fetchFiles) > 0 {
		return remoteCopy(remoteSrc, ch.dir, fetchFiles)
	}
	return nil
}

func gatherFilesToFetch(
	ch LocalCache,
	workspaceDir string,
	art artifact.Artifact,
	remoteSrc string,
	filesToFetch map[string]struct{},
) error {
	if art.SkipCache {
		return nil
	}
	status, cachePath, _, err := quickStatus(ch, workspaceDir, art)
	if err != nil {
		return err
	}
	if !status.HasChecksum {
		return InvalidChecksumError{art.Checksum}
	}
	if status.ChecksumInCache {
		return nil
	}
	if art.IsDir {
		if err := remoteCopy(remoteSrc, ch.dir, map[string]struct{}{cachePath: {}}); err != nil {
			return err
		}
		man, err := readDirManifest(filepath.Join(ch.dir, cachePath))
		if err != nil {
			return err
		}
		childWorkspaceDir := filepath.Join(workspaceDir, art.Path)
		for _, childArt := range man.Contents {
			if err := gatherFilesToFetch(
				ch,
				childWorkspaceDir,
				*childArt,
				remoteSrc,
				filesToFetch,
			); err != nil {
				return err
			}
		}
	} else {
		filesToFetch[cachePath] = struct{}{}
	}
	return nil
}
